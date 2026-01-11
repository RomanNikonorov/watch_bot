package watch

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// MockHTTPClient implements HTTPClient interface for mocking
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return m.DoFunc(req)
}

// TestIsUrlOkHealthy verifies that IsUrlOk returns true for a valid URL
func TestIsUrlOkHealthy(t *testing.T) {
	// Create a fake HTTP client that responds with 200 OK
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("OK")),
			}, nil
		},
	}

	// Create an instance of RealURLChecker
	checker := RealURLChecker{}

	// Call the IsUrlOk method
	url := "http://example.com"
	unhealthyThreshold := 1
	unhealthyDelay := 1
	result := checker.IsUrlOk(url, unhealthyThreshold, unhealthyDelay, mockClient)

	// Verify the result
	if !result {
		t.Errorf("Expected IsUrlOk to return true, but got false")
	}
}

// TestIsUrlOkUnhealthy verifies that IsUrlOk returns false for an invalid URL
func TestIsUrlOkUnhealthy(t *testing.T) {
	// Create a fake HTTP client that responds with 404 Not Found
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("OK")),
			}, nil
		},
	}

	// Create an instance of RealURLChecker
	checker := RealURLChecker{}

	// Call the IsUrlOk method
	url := "http://example.com"
	unhealthyThreshold := 1
	unhealthyDelay := 1
	result := checker.IsUrlOk(url, unhealthyThreshold, unhealthyDelay, mockClient)

	// Verify the result
	if result {
		t.Errorf("Expected IsUrlOk to return false, but got true")
	}
}

// TestIsUrlOkUnreachable verifies that IsUrlOk returns false for an unreachable URL
func TestIsUrlOkUnreachable(t *testing.T) {
	// Create a fake HTTP client that always returns an error
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, http.ErrNoLocation
		},
	}

	// Create an instance of RealURLChecker
	checker := RealURLChecker{}

	// Call the IsUrlOk method
	url := "http://example.com"
	unhealthyThreshold := 1
	unhealthyDelay := 1
	result := checker.IsUrlOk(url, unhealthyThreshold, unhealthyDelay, mockClient)

	// Verify the result
	if result {
		t.Errorf("Expected IsUrlOk to return false, but got true")
	}
}
