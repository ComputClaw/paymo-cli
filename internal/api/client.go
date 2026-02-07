package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	DefaultBaseURL = "https://app.paymoapp.com/api"
	DefaultTimeout = 30 * time.Second
)

// Client is the Paymo API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Auth       Authenticator
	
	// Rate limiting
	rateMu        sync.Mutex
	rateLimit     int
	rateRemaining int
	rateReset     time.Time
}

// Authenticator interface for different auth methods
type Authenticator interface {
	SetAuth(req *http.Request) error
	Type() string
}

// APIKeyAuth implements API key authentication
type APIKeyAuth struct {
	APIKey string
}

func (a *APIKeyAuth) SetAuth(req *http.Request) error {
	req.SetBasicAuth(a.APIKey, "x") // API key as username, any password
	return nil
}

func (a *APIKeyAuth) Type() string {
	return "api_key"
}

// BasicAuth implements email/password authentication
type BasicAuth struct {
	Email    string
	Password string
}

func (b *BasicAuth) SetAuth(req *http.Request) error {
	req.SetBasicAuth(b.Email, b.Password)
	return nil
}

func (b *BasicAuth) Type() string {
	return "basic"
}

// NewClient creates a new Paymo API client
func NewClient(auth Authenticator) *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		Auth: auth,
	}
}

// NewClientWithBaseURL creates a client with a custom base URL
func NewClientWithBaseURL(baseURL string, auth Authenticator) *Client {
	client := NewClient(auth)
	client.BaseURL = strings.TrimSuffix(baseURL, "/")
	return client
}

// APIError represents an error from the Paymo API
type APIError struct {
	StatusCode int
	Message    string
	Details    map[string]interface{}
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("Paymo API error (%d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("Paymo API error: HTTP %d", e.StatusCode)
}

// Request makes an authenticated request to the Paymo API
func (c *Client) Request(method, path string, body io.Reader, result interface{}) error {
	// Check rate limiting
	c.rateMu.Lock()
	if c.rateRemaining == 0 && time.Now().Before(c.rateReset) {
		waitTime := time.Until(c.rateReset)
		c.rateMu.Unlock()
		time.Sleep(waitTime)
		c.rateMu.Lock()
	}
	c.rateMu.Unlock()

	// Build URL
	reqURL := fmt.Sprintf("%s/%s", c.BaseURL, strings.TrimPrefix(path, "/"))
	
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set authentication
	if c.Auth != nil {
		if err := c.Auth.SetAuth(req); err != nil {
			return fmt.Errorf("setting auth: %w", err)
		}
	}

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limit info
	c.updateRateLimit(resp)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		
		// Try to parse error message
		var errResp map[string]interface{}
		if json.Unmarshal(respBody, &errResp) == nil {
			if msg, ok := errResp["message"].(string); ok {
				apiErr.Message = msg
			}
			apiErr.Details = errResp
		}
		
		return apiErr
	}

	// Parse successful response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
	}

	return nil
}

// updateRateLimit updates rate limit tracking from response headers
func (c *Client) updateRateLimit(resp *http.Response) {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()

	if limit := resp.Header.Get("X-Ratelimit-Limit"); limit != "" {
		fmt.Sscanf(limit, "%d", &c.rateLimit)
	}
	if remaining := resp.Header.Get("X-Ratelimit-Remaining"); remaining != "" {
		fmt.Sscanf(remaining, "%d", &c.rateRemaining)
	}
	if decay := resp.Header.Get("X-Ratelimit-Decay-Period"); decay != "" {
		var seconds int
		fmt.Sscanf(decay, "%d", &seconds)
		c.rateReset = time.Now().Add(time.Duration(seconds) * time.Second)
	}
}

// Get makes a GET request
func (c *Client) Get(path string, result interface{}) error {
	return c.Request(http.MethodGet, path, nil, result)
}

// GetWithParams makes a GET request with query parameters
func (c *Client) GetWithParams(path string, params url.Values, result interface{}) error {
	if len(params) > 0 {
		path = fmt.Sprintf("%s?%s", path, params.Encode())
	}
	return c.Get(path, result)
}

// Post makes a POST request
func (c *Client) Post(path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling body: %w", err)
		}
		bodyReader = strings.NewReader(string(jsonBody))
	}
	return c.Request(http.MethodPost, path, bodyReader, result)
}

// Put makes a PUT request
func (c *Client) Put(path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling body: %w", err)
		}
		bodyReader = strings.NewReader(string(jsonBody))
	}
	return c.Request(http.MethodPut, path, bodyReader, result)
}

// Delete makes a DELETE request
func (c *Client) Delete(path string) error {
	return c.Request(http.MethodDelete, path, nil, nil)
}