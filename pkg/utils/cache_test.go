package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGenerateETag(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"Empty string", "", "\"2166136261\""},
		{"Simple string", "test", "\"2949673445\""},
		{"Resource string", "products:123", "\"3771203499\""},
		{"Query string", "page=1&limit=10", "\"1852160973\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateETag(tt.content); got != tt.want {
				t.Errorf("GenerateETag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateResourceETag(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
		id           string
		want         string
	}{
		{"Product resource", "products", "123", "\"3771203499\""},
		{"User resource", "users", "abc", "\"2619466549\""},
		{"Empty resource", "", "", "\"1057798253\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateResourceETag(tt.resourceName, tt.id); got != tt.want {
				t.Errorf("GenerateResourceETag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateQueryETag(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		want        string
	}{
		{"Empty query", "", "\"2166136261\""},
		{"Pagination query", "page=1&limit=10", "\"1852160973\""},
		{"Filter query", "filter[name]=test", "\"1756232217\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateQueryETag(tt.queryString); got != tt.want {
				t.Errorf("GenerateQueryETag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateRequestETag(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		query    string
		expected string
	}{
		{"Root path", "/", "", "\"2463255987\""},
		{"API path", "/api/products", "", "\"2909290312\""},
		{"Path with query", "/api/products", "page=1&limit=10", "\"2090847066\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path+"?"+tt.query, nil)
			got := GenerateRequestETag(req)
			if got != tt.expected {
				t.Errorf("GenerateRequestETag() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormatLastModified(t *testing.T) {
	// Create a fixed time for testing
	testTime := time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC)
	expected := "Mon, 15 May 2023 10:30:00 GMT"

	result := FormatLastModified(testTime)
	if result != expected {
		t.Errorf("FormatLastModified() = %v, want %v", result, expected)
	}
}

func TestIsETagMatch(t *testing.T) {
	tests := []struct {
		name        string
		etag        string
		ifNoneMatch string
		want        bool
	}{
		{"Exact match", "\"123\"", "\"123\"", true},
		{"No match", "\"123\"", "\"456\"", false},
		{"Wildcard match", "\"123\"", "*", true},
		{"Empty If-None-Match", "\"123\"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsETagMatch(tt.etag, tt.ifNoneMatch); got != tt.want {
				t.Errorf("IsETagMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsModifiedSince(t *testing.T) {
	now := time.Now().UTC()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	pastTimeStr := pastTime.Format(http.TimeFormat)
	futureTimeStr := futureTime.Format(http.TimeFormat)

	tests := []struct {
		name            string
		lastModified    time.Time
		ifModifiedSince string
		want            bool
	}{
		{"Modified after header", now, pastTimeStr, true},
		{"Not modified since header", pastTime, futureTimeStr, false},
		{"Empty header", now, "", true},
		{"Invalid header", now, "invalid-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsModifiedSince(tt.lastModified, tt.ifModifiedSince); got != tt.want {
				t.Errorf("IsModifiedSince() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetCacheHeaders(t *testing.T) {
	tests := []struct {
		name         string
		maxAge       int
		etag         string
		lastModified *time.Time
		varyHeaders  []string
		checkHeaders map[string]string
	}{
		{
			name:         "All headers",
			maxAge:       60,
			etag:         "\"123\"",
			lastModified: func() *time.Time { tm := time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC); return &tm }(),
			varyHeaders:  []string{"Accept", "Authorization"},
			checkHeaders: map[string]string{
				"Cache-Control": "public, max-age=60",
				"ETag":          "\"123\"",
				"Last-Modified": "Mon, 15 May 2023 10:30:00 GMT",
				"Vary":          "Accept, Authorization",
			},
		},
		{
			name:         "No ETag",
			maxAge:       120,
			etag:         "",
			lastModified: func() *time.Time { tm := time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC); return &tm }(),
			varyHeaders:  []string{"Accept"},
			checkHeaders: map[string]string{
				"Cache-Control": "public, max-age=120",
				"Last-Modified": "Mon, 15 May 2023 10:30:00 GMT",
				"Vary":          "Accept",
			},
		},
		{
			name:         "No Last-Modified",
			maxAge:       30,
			etag:         "\"456\"",
			lastModified: nil,
			varyHeaders:  nil,
			checkHeaders: map[string]string{
				"Cache-Control": "public, max-age=30",
				"ETag":          "\"456\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SetCacheHeaders(w, tt.maxAge, tt.etag, tt.lastModified, tt.varyHeaders)

			for header, expected := range tt.checkHeaders {
				if got := w.Header().Get(header); got != expected {
					t.Errorf("Header %s = %v, want %v", header, got, expected)
				}
			}

			// Check that headers not specified in checkHeaders are not set
			if tt.etag == "" && w.Header().Get("ETag") != "" {
				t.Errorf("ETag header should not be set")
			}
			if tt.lastModified == nil && w.Header().Get("Last-Modified") != "" {
				t.Errorf("Last-Modified header should not be set")
			}
			if len(tt.varyHeaders) == 0 && w.Header().Get("Vary") != "" {
				t.Errorf("Vary header should not be set")
			}
		})
	}
}

func TestDisableCaching(t *testing.T) {
	w := httptest.NewRecorder()
	DisableCaching(w)

	expectedHeaders := map[string]string{
		"Cache-Control": "no-store, no-cache, must-revalidate, max-age=0",
		"Pragma":        "no-cache",
		"Expires":       "0",
	}

	for header, expected := range expectedHeaders {
		if got := w.Header().Get(header); got != expected {
			t.Errorf("Header %s = %v, want %v", header, got, expected)
		}
	}
}
