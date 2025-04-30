package auth

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stanxing/refine-gin/pkg/resource"
)

// AuthorizationProvider provides authorization functionality
type AuthorizationProvider interface {
	// CanAccess checks if the current user can access a resource with the given operation
	CanAccess(c *gin.Context, res resource.Resource, op resource.Operation) bool

	// CanAccessRecord checks if the current user can access a specific record
	CanAccessRecord(c *gin.Context, res resource.Resource, op resource.Operation, record interface{}) bool
}

// DefaultAuthorizationProvider is a default implementation of AuthorizationProvider
type DefaultAuthorizationProvider struct {
	// Rules for authorization
	Rules map[string]map[resource.Operation]AuthorizationRule
}

// AuthorizationRule is a function that checks if access is allowed
type AuthorizationRule func(c *gin.Context, record interface{}) bool

// CanAccess checks if the current user can access a resource with the given operation
func (p *DefaultAuthorizationProvider) CanAccess(c *gin.Context, res resource.Resource, op resource.Operation) bool {
	resourceRules, ok := p.Rules[res.GetName()]
	if !ok {
		return true // Allow by default
	}

	rule, ok := resourceRules[op]
	if !ok {
		return true // Allow by default
	}

	return rule(c, nil)
}

// CanAccessRecord checks if the current user can access a specific record
func (p *DefaultAuthorizationProvider) CanAccessRecord(c *gin.Context, res resource.Resource, op resource.Operation, record interface{}) bool {
	resourceRules, ok := p.Rules[res.GetName()]
	if !ok {
		return true // Allow by default
	}

	rule, ok := resourceRules[op]
	if !ok {
		return true // Allow by default
	}

	return rule(c, record)
}

// NewDefaultAuthorizationProvider creates a new default authorization provider
func NewDefaultAuthorizationProvider() *DefaultAuthorizationProvider {
	return &DefaultAuthorizationProvider{
		Rules: make(map[string]map[resource.Operation]AuthorizationRule),
	}
}

// AddRule adds a rule for a resource and operation
func (p *DefaultAuthorizationProvider) AddRule(resourceName string, op resource.Operation, rule AuthorizationRule) *DefaultAuthorizationProvider {
	if _, ok := p.Rules[resourceName]; !ok {
		p.Rules[resourceName] = make(map[resource.Operation]AuthorizationRule)
	}

	p.Rules[resourceName][op] = rule
	return p
}

// JWTAuthorizationProvider is an implementation of AuthorizationProvider that uses JWT claims
type JWTAuthorizationProvider struct {
	// Rules for authorization
	Rules map[string]map[resource.Operation]JWTAuthorizationRule
}

// JWTAuthorizationRule is a function that checks if access is allowed based on JWT claims
type JWTAuthorizationRule func(claims jwt.MapClaims, record interface{}) bool

// CanAccess checks if the current user can access a resource with the given operation
func (p *JWTAuthorizationProvider) CanAccess(c *gin.Context, res resource.Resource, op resource.Operation) bool {
	// Get claims from context
	claimsValue, exists := c.Get("claims")
	if !exists {
		return false
	}

	claims, ok := claimsValue.(jwt.MapClaims)
	if !ok {
		return false
	}

	resourceRules, ok := p.Rules[res.GetName()]
	if !ok {
		return true // Allow by default
	}

	rule, ok := resourceRules[op]
	if !ok {
		return true // Allow by default
	}

	return rule(claims, nil)
}

// CanAccessRecord checks if the current user can access a specific record
func (p *JWTAuthorizationProvider) CanAccessRecord(c *gin.Context, res resource.Resource, op resource.Operation, record interface{}) bool {
	// Get claims from context
	claimsValue, exists := c.Get("claims")
	if !exists {
		return false
	}

	claims, ok := claimsValue.(jwt.MapClaims)
	if !ok {
		return false
	}

	resourceRules, ok := p.Rules[res.GetName()]
	if !ok {
		return true // Allow by default
	}

	rule, ok := resourceRules[op]
	if !ok {
		return true // Allow by default
	}

	return rule(claims, record)
}

// NewJWTAuthorizationProvider creates a new JWT authorization provider
func NewJWTAuthorizationProvider() *JWTAuthorizationProvider {
	return &JWTAuthorizationProvider{
		Rules: make(map[string]map[resource.Operation]JWTAuthorizationRule),
	}
}

// AddRule adds a rule for a resource and operation
func (p *JWTAuthorizationProvider) AddRule(resourceName string, op resource.Operation, rule JWTAuthorizationRule) *JWTAuthorizationProvider {
	if _, ok := p.Rules[resourceName]; !ok {
		p.Rules[resourceName] = make(map[resource.Operation]JWTAuthorizationRule)
	}

	p.Rules[resourceName][op] = rule
	return p
}

// AuthorizationMiddleware creates a middleware for authorization
func AuthorizationMiddleware(provider AuthorizationProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get resource and operation from context
		res, ok := c.Get("resource")
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
			return
		}

		op, ok := c.Get("operation")
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
			return
		}

		// Check if access is allowed
		if !provider.CanAccess(c, res.(resource.Resource), op.(resource.Operation)) {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.Next()
	}
}

// Common JWT authorization rules

// HasRole checks if the user has a specific role
func HasRole(role string) JWTAuthorizationRule {
	return func(claims jwt.MapClaims, record interface{}) bool {
		// Check if roles claim exists
		rolesValue, ok := claims["roles"]
		if !ok {
			return false
		}

		// Check if roles is an array
		roles, ok := rolesValue.([]interface{})
		if !ok {
			return false
		}

		// Check if role is in roles
		for _, r := range roles {
			if r == role {
				return true
			}
		}

		return false
	}
}

// HasAnyRole checks if the user has any of the specified roles
func HasAnyRole(roles ...string) JWTAuthorizationRule {
	return func(claims jwt.MapClaims, record interface{}) bool {
		// Check if roles claim exists
		rolesValue, ok := claims["roles"]
		if !ok {
			return false
		}

		// Check if roles is an array
		userRoles, ok := rolesValue.([]interface{})
		if !ok {
			return false
		}

		// Check if any role is in user roles
		for _, role := range roles {
			for _, r := range userRoles {
				if r == role {
					return true
				}
			}
		}

		return false
	}
}

// HasAllRoles checks if the user has all of the specified roles
func HasAllRoles(roles ...string) JWTAuthorizationRule {
	return func(claims jwt.MapClaims, record interface{}) bool {
		// Check if roles claim exists
		rolesValue, ok := claims["roles"]
		if !ok {
			return false
		}

		// Check if roles is an array
		userRoles, ok := rolesValue.([]interface{})
		if !ok {
			return false
		}

		// Check if all roles are in user roles
		for _, role := range roles {
			found := false
			for _, r := range userRoles {
				if r == role {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}

		return true
	}
}

// IsOwner checks if the user is the owner of the record
func IsOwner(userIDField, ownerField string) JWTAuthorizationRule {
	return func(claims jwt.MapClaims, record interface{}) bool {
		if record == nil {
			return false
		}

		// Get user ID from claims
		userID, ok := claims[userIDField]
		if !ok {
			return false
		}

		// Get owner ID from record using reflection
		recordValue := reflect.ValueOf(record)
		if recordValue.Kind() == reflect.Ptr {
			recordValue = recordValue.Elem()
		}

		if recordValue.Kind() != reflect.Struct {
			return false
		}

		ownerFieldValue := recordValue.FieldByName(ownerField)
		if !ownerFieldValue.IsValid() {
			return false
		}

		// Compare user ID and owner ID
		return fmt.Sprintf("%v", userID) == fmt.Sprintf("%v", ownerFieldValue.Interface())
	}
}
