package joplin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the local Web Clipper service address that Joplin Desktop
// binds by default.
const DefaultBaseURL = "http://localhost:41184"

// DefaultTimeout is the default per-request timeout. Joplin's local API is
// fast; if it does not respond inside this window something is wrong.
const DefaultTimeout = 10 * time.Second

// Client talks to Joplin Desktop's local Web Clipper REST API.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// Options configures a new Client.
type Options struct {
	// Token is the Joplin Web Clipper API token. Required.
	Token string
	// BaseURL is the Joplin Web Clipper base URL. Defaults to DefaultBaseURL.
	BaseURL string
	// HTTPClient lets the caller supply a customised http.Client. If nil, a
	// new one is created with Timeout = DefaultTimeout.
	HTTPClient *http.Client
}

// New returns a Client configured with the supplied options.
func New(opts Options) (*Client, error) {
	if opts.Token == "" {
		return nil, errors.New("joplin: token is required")
	}
	base := opts.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	base = strings.TrimRight(base, "/")
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: DefaultTimeout}
	}
	return &Client{baseURL: base, token: opts.Token, http: hc}, nil
}

// APIError is returned for any non-2xx HTTP response from Joplin.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("joplin: http %d", e.StatusCode)
	}
	return fmt.Sprintf("joplin: http %d: %s", e.StatusCode, e.Message)
}

// IsNotFound reports whether err is a 404 from the Joplin API.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// do performs the HTTP request, decoding the JSON response into out (if non-nil)
// and returning an *APIError for any non-2xx status.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	if query == nil {
		query = url.Values{}
	}
	query.Set("token", c.token)

	u := c.baseURL + path
	if encoded := query.Encode(); encoded != "" {
		u = u + "?" + encoded
	}

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("joplin: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("joplin: build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(msg))}
	}

	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("joplin: decode response: %w", err)
	}
	return nil
}

// rawGET issues a GET and returns the raw response body bytes plus content type.
// Used for binary endpoints like resource downloads.
func (c *Client) rawGET(ctx context.Context, path string, query url.Values) ([]byte, string, error) {
	if query == nil {
		query = url.Values{}
	}
	query.Set("token", c.token)
	u := c.baseURL + path + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, "", fmt.Errorf("joplin: build request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, "", &APIError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(msg))}
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("joplin: read body: %w", err)
	}
	return data, resp.Header.Get("Content-Type"), nil
}

// listQuery builds a url.Values for list endpoints from ListOptions.
func listQuery(opts ListOptions) url.Values {
	q := url.Values{}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.OrderBy != "" {
		q.Set("order_by", opts.OrderBy)
	}
	if opts.OrderDir != "" {
		q.Set("order_dir", opts.OrderDir)
	}
	if len(opts.Fields) > 0 {
		q.Set("fields", strings.Join(opts.Fields, ","))
	}
	return q
}
