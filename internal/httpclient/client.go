package httpclient

import "net/http"

// HTTPClient defines the interface for HTTP operations
// This allows for easy mocking in tests
type HTTPClient interface {
	// Do executes an HTTP request and returns the response
	Do(req *http.Request) (*http.Response, error)
}
