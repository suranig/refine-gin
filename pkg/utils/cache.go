package utils

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"time"
)

// GenerateETag creates a hash of the given string to use as an ETag.
// The ETag is a quoted string as per HTTP spec.
func GenerateETag(content string) string {
	h := fnv.New32a()
	h.Write([]byte(content))
	return fmt.Sprintf("\"%d\"", h.Sum32())
}

// GenerateResourceETag creates a hash specifically for a resource and ID combination.
// This is useful for individual resource endpoints.
func GenerateResourceETag(resourceName string, id string) string {
	return GenerateETag(resourceName + ":" + id)
}

// GenerateQueryETag creates a hash for a query string.
// This is useful for list endpoints that depend on query parameters.
func GenerateQueryETag(queryString string) string {
	return GenerateETag(queryString)
}

// GenerateRequestETag creates an ETag based on request URL path and query parameters.
// This provides a unique identifier for the specific request.
func GenerateRequestETag(r *http.Request) string {
	return GenerateETag(r.URL.Path + "?" + r.URL.RawQuery)
}

// FormatLastModified formats a time.Time value as an HTTP-compatible Last-Modified header value.
func FormatLastModified(t time.Time) string {
	return t.UTC().Format(http.TimeFormat)
}

// IsETagMatch compares the provided ETag with the If-None-Match header value.
// Returns true if they match, indicating the client's cache is still valid.
func IsETagMatch(etag string, ifNoneMatch string) bool {
	return ifNoneMatch != "" && (ifNoneMatch == etag || ifNoneMatch == "*")
}

// IsModifiedSince checks if the resource has been modified since the time specified
// in the If-Modified-Since header. Returns true if resource is newer than the header value.
func IsModifiedSince(lastModified time.Time, ifModifiedSince string) bool {
	// If no If-Modified-Since header is provided, consider it modified
	if ifModifiedSince == "" {
		return true
	}

	// Parse the If-Modified-Since header
	modifiedSinceTime, err := http.ParseTime(ifModifiedSince)
	if err != nil {
		// If parsing fails, consider it modified
		return true
	}

	// Truncate to seconds for accurate comparison (HTTP dates don't have sub-second precision)
	lastModified = lastModified.Truncate(time.Second)
	modifiedSinceTime = modifiedSinceTime.Truncate(time.Second)

	// Resource is modified if it's newer than the If-Modified-Since time
	return lastModified.After(modifiedSinceTime)
}

// SetCacheHeaders sets standard HTTP cache headers on an http.ResponseWriter.
// This is a helper function to consistently apply cache settings.
func SetCacheHeaders(w http.ResponseWriter, maxAge int, etag string, lastModified *time.Time, varyHeaders []string) {
	// Set Cache-Control header
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))

	// Set ETag header if provided
	if etag != "" {
		w.Header().Set("ETag", etag)
	}

	// Set Last-Modified header if provided
	if lastModified != nil && !lastModified.IsZero() {
		w.Header().Set("Last-Modified", FormatLastModified(*lastModified))
	}

	// Set Vary header if headers are provided
	if len(varyHeaders) > 0 {
		w.Header().Set("Vary", strings.Join(varyHeaders, ", "))
	}
}

// DisableCaching sets headers that prevent caching on an http.ResponseWriter.
// This is useful for dynamic content or endpoints that should never be cached.
func DisableCaching(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// GenerateETagFromSlice creates a hash of the given slice of strings to use as an ETag.
// The ETag is a quoted string as per HTTP spec.
func GenerateETagFromSlice(items []string) string {
	return GenerateETag(strings.Join(items, ","))
}
