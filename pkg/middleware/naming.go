package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/naming"
)

// responseBodyWriter is a wrapper for gin.ResponseWriter
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write writes the response body
func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// NamingConventionMiddleware converts JSON field names to the specified convention
func NamingConventionMiddleware(convention naming.NamingConvention) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if not JSON request/response
		if c.ContentType() != "application/json" && c.GetHeader("Accept") != "application/json" {
			c.Next()
			return
		}

		// Handle request body conversion
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body.Close()

			// Try to parse as JSON
			var data map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &data); err == nil {
				// Convert keys to desired convention
				data = naming.ConvertKeys(data, convention)
				// Marshal back to JSON
				if newBody, err := json.Marshal(data); err == nil {
					c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
					c.Request.ContentLength = int64(len(newBody))
					c.Request.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
				} else {
					// Reset body if marshal fails
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			} else {
				// Reset body if unmarshal fails
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Create a response body writer to capture the response
		w := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		// Continue processing the request
		c.Next()

		// Convert response body JSON keys
		if w.Header().Get("Content-Type") == "application/json" {
			responseBody := w.body.Bytes()
			var data map[string]interface{}
			if err := json.Unmarshal(responseBody, &data); err == nil {
				// Convert keys to desired convention
				data = naming.ConvertKeys(data, convention)
				// Marshal back to JSON
				if newBody, err := json.Marshal(data); err == nil {
					// Replace original response
					w.ResponseWriter.Header().Set("Content-Length", strconv.Itoa(len(newBody)))
					w.ResponseWriter.WriteHeader(w.Status())
					w.ResponseWriter.Write(newBody)
					return
				}
			}
		}
	}
}
