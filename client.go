package safebooru

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultBaseURL   = "https://safebooru.org"
	DefaultUserAgent = "safebooru-go/0.1 (+https://github.com/Ouja/safebooru)"

	// MaxPostLimit is the API's hard limit of posts per request.
	MaxPostLimit = 1000

	// maxBodySize protects us from unexpectedly huge responses.
	maxBodySize = 10 << 20 // 10 MB
)

// Client talks to a Safebooru-compatible API. Create one with NewClient;
// it is safe for concurrent use by multiple goroutines.
type Client struct {
	baseURL    string
	userAgent  string
	httpClient *http.Client
}

// Option changes a Client setting. Usage:
//
//	safebooru.NewClient()                                      // defaults
//	safebooru.NewClient(safebooru.WithUserAgent("my-app/1.0")) // custom
type Option func(*Client)

// WithBaseURL points the client at another Gelbooru-style (dapi) site.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = baseURL }
}

// WithUserAgent sets the User-Agent header. Please identify your app.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) { c.userAgent = userAgent }
}

// WithHTTPClient replaces the default *http.Client (10s timeout).
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) { c.httpClient = httpClient }
}

// NewClient returns a Client with sensible defaults, changed by opts.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		userAgent:  DefaultUserAgent,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// get calls {baseURL}/index.php?page=dapi&<params> and returns the raw body.
// Every endpoint (posts, tags, ...) goes through here, so the HTTP
// plumbing is written only once.
func (c *Client) get(ctx context.Context, params url.Values) ([]byte, error) {
	params.Set("page", "dapi")
	endpoint := c.baseURL + "/index.php?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("safebooru: build request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("safebooru: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return nil, fmt.Errorf("safebooru: read body: %w", err)
	}

	body = bytes.TrimSpace(body)
	if err := checkSuccess(body); err != nil {
		return nil, err
	}
	return body, nil
}

// checkSuccess detects the API's "soft" errors: HTTP 200 with
// {"success": false, "message": "search down"} (or the XML equivalent).
func checkSuccess(body []byte) error {
	switch {
	case len(body) == 0:
		return nil
	case body[0] == '{':
		var res struct {
			Success *bool  `json:"success"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &res) == nil && res.Success != nil && !*res.Success {
			return fmt.Errorf("%w: %s", ErrSearchDown, res.Message)
		}
	case bytes.HasPrefix(body, []byte("<response")) && bytes.Contains(body, []byte(`success="false"`)):
		return ErrSearchDown
	}
	return nil
}
