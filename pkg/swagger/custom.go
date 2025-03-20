package swagger

// CustomEndpoint defines a custom endpoint for Swagger documentation.
type CustomEndpoint struct {
	Method    string    // HTTP method in lowercase (e.g., "get", "post")
	Path      string    // The URL path, e.g., "/auth/login"
	Operation Operation // The operation object that defines the endpoint.
}

var customEndpoints []CustomEndpoint

// RegisterCustomEndpoint registers a custom endpoint to be included in the OpenAPI spec.
func RegisterCustomEndpoint(endpoint CustomEndpoint) {
	customEndpoints = append(customEndpoints, endpoint)
}

// GetCustomEndpoints returns all registered custom endpoints.
func GetCustomEndpoints() []CustomEndpoint {
	return customEndpoints
}

// ResetCustomEndpoints clears all registered custom endpoints (useful in testing scenarios).
func ResetCustomEndpoints() {
	customEndpoints = []CustomEndpoint{}
}
