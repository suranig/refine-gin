package dto

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
}
