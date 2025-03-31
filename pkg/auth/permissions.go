package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/suranig/refine-gin/pkg/resource"
)

// PermissionMiddleware creates a middleware for checking resource-level permissions
func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get resource and operation from context
		res, ok := c.Get("resource")
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden: resource not found in context"})
			return
		}

		op, ok := c.Get("operation")
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden: operation not found in context"})
			return
		}

		// Get user role from JWT claims
		claimsValue, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden: no authentication claims found"})
			return
		}

		claims, ok := claimsValue.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden: invalid authentication claims"})
			return
		}

		// Extract user role(s)
		userRoles := extractUserRoles(claims)
		if len(userRoles) == 0 {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden: no roles found"})
			return
		}

		// Convert operation to string
		opStr := ""
		switch op.(resource.Operation) {
		case resource.OperationList:
			opStr = "list"
		case resource.OperationCreate:
			opStr = "create"
		case resource.OperationRead:
			opStr = "read"
		case resource.OperationUpdate:
			opStr = "update"
		case resource.OperationDelete:
			opStr = "delete"
		default:
			opStr = "unknown"
		}

		// Check resource-level permissions
		if !hasResourcePermission(res.(resource.Resource), opStr, userRoles) {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden: insufficient permissions for this resource"})
			return
		}

		// Store user roles in context for field-level permission filtering
		c.Set("userRoles", userRoles)

		c.Next()
	}
}

// extractUserRoles extracts user roles from JWT claims
func extractUserRoles(claims jwt.MapClaims) []string {
	// Check if roles claim exists
	rolesValue, ok := claims["roles"]
	if !ok {
		return nil
	}

	// Try to extract roles as array
	if roles, ok := rolesValue.([]interface{}); ok {
		result := make([]string, 0, len(roles))
		for _, r := range roles {
			if roleStr, ok := r.(string); ok {
				result = append(result, roleStr)
			}
		}
		return result
	}

	// Try to extract single role as string
	if roleStr, ok := rolesValue.(string); ok {
		return []string{roleStr}
	}

	return nil
}

// hasResourcePermission checks if any of the user roles has permission for the operation
func hasResourcePermission(res resource.Resource, operation string, userRoles []string) bool {
	// Get resource permissions
	permissions := res.GetPermissions()
	if permissions == nil {
		return true // If no permissions are defined, allow access by default
	}

	// Get allowed roles for the operation
	allowedRoles, exists := permissions[operation]
	if !exists || len(allowedRoles) == 0 {
		return true // If no roles are specified for the operation, allow access
	}

	// Check if any user role is in the allowed roles
	for _, userRole := range userRoles {
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				return true
			}
		}
	}

	return false
}

// FilterFieldsByPermission filters fields based on field-level permissions
func FilterFieldsByPermission(fields []resource.FieldMetadata, operation string, userRoles []string) []resource.FieldMetadata {
	if len(fields) == 0 {
		return fields
	}

	result := make([]resource.FieldMetadata, 0, len(fields))

	for _, field := range fields {
		// Check if the field has permission restrictions
		if field.Permissions == nil || len(field.Permissions) == 0 {
			// No permissions defined for this field, include it
			result = append(result, field)
			continue
		}

		// Check if there are permissions defined for this operation
		allowedRoles, exists := field.Permissions[operation]
		if !exists || len(allowedRoles) == 0 {
			// No roles specified for this operation, include the field
			result = append(result, field)
			continue
		}

		// Check if the user has any of the required roles
		hasPermission := false
		for _, userRole := range userRoles {
			for _, allowedRole := range allowedRoles {
				if userRole == allowedRole {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if hasPermission {
			result = append(result, field)
		}
	}

	return result
}
