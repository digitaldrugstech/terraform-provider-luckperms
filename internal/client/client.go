package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for the LuckPerms REST API.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// New creates a new LuckPerms API client.
func New(baseURL, apiKey string, timeout time.Duration, insecure bool) *Client {
	var transport *http.Transport
	if dt, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = dt.Clone()
	} else {
		transport = &http.Transport{}
	}
	if insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // user-configured
		}
	}

	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

// APIError represents an error response from the LuckPerms REST API.
type APIError struct {
	StatusCode int
	Body       string
	Method     string
	Path       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("luckperms api: %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

// IsNotFound returns true if the error (or any wrapped error) is a 404 response.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsConflict returns true if the error (or any wrapped error) is a 409 response.
func IsConflict(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 409
	}
	return false
}

var retryDelays = []time.Duration{0, 1 * time.Second, 2 * time.Second, 4 * time.Second}

// isIdempotent returns true for HTTP methods safe to retry.
func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead:
		return true
	default:
		return false
	}
}

// doRequest executes an HTTP request with retry logic for 5xx errors on idempotent methods.
// POST requests are never retried to avoid duplicate resource creation.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var jsonBytes []byte
	if body != nil {
		var err error
		jsonBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
	}

	url := c.BaseURL + path
	canRetry := isIdempotent(method)

	maxAttempts := 1
	if canRetry {
		maxAttempts = len(retryDelays)
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if canRetry && attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelays[attempt]):
			}
		}

		var bodyReader io.Reader
		if jsonBytes != nil {
			bodyReader = bytes.NewReader(jsonBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		if c.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.APIKey)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("executing request: %w", err)
			if canRetry && attempt < maxAttempts-1 {
				continue
			}
			return nil, lastErr
		}

		const maxResponseSize = 10 << 20 // 10 MB
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}

		// Retry on 5xx (only for idempotent methods)
		if resp.StatusCode >= 500 && canRetry && attempt < maxAttempts-1 {
			lastErr = &APIError{
				StatusCode: resp.StatusCode,
				Body:       string(respBody),
				Method:     method,
				Path:       path,
			}
			continue
		}

		// Return error for non-2xx
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Body:       string(respBody),
				Method:     method,
				Path:       path,
			}
		}

		return respBody, nil
	}

	return nil, lastErr
}

// doRequestNoBody executes a request that expects no response body (204, etc).
func (c *Client) doRequestNoBody(ctx context.Context, method, path string, body interface{}) error {
	_, err := c.doRequest(ctx, method, path, body)
	return err
}
