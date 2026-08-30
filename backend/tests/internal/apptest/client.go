//go:build integration

package apptest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// Client is a small HTTP client for driving the app like a real API
// consumer: it carries cookies (needed for the refresh_token flow) across
// requests the same way a browser would, and gives every step definition
// one place to send a request and inspect the result.
type Client struct {
	http    *http.Client
	baseURL string
}

// NewClient returns a client with its own empty cookie jar. Give every
// scenario its own Client (see auth's Before hook) — sharing one across
// scenarios would leak one scenario's session into the next.
func NewClient(baseURL string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		http:    &http.Client{Jar: jar},
		baseURL: baseURL,
	}
}

// BaseURL returns the URL this client sends requests relative to.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Response is the captured result of one request: enough for steps to
// assert on status, headers, and body without re-parsing raw net/http
// types.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// JSON decodes the response body into v.
func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.Body, v)
}

// Do sends a request to path (relative to baseURL) through this client's
// cookie jar, with an optional JSON body and optional extra headers (e.g.
// "Authorization"). Use this for the common case: let the jar replay
// whatever cookies a prior response set, exactly like a browser would.
func (c *Client) Do(method, path string, body any, headers map[string]string) (*Response, error) {
	req, err := newJSONRequest(c.baseURL, method, path, body, headers)
	if err != nil {
		return nil, err
	}
	return doRequest(c.http, req)
}

// DoWithCookie sends a request carrying exactly one cookie, ignoring
// whatever this client's jar currently holds. Use it to present a
// specific — possibly stale, rotated-away, or outright invalid — cookie
// value that the jar itself could never produce.
func (c *Client) DoWithCookie(method, path, cookieName, cookieValue string) (*Response, error) {
	req, err := newJSONRequest(c.baseURL, method, path, nil, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{Name: cookieName, Value: cookieValue})
	return doRequest(&http.Client{}, req)
}

// DoNoCookies sends a request with no cookies at all, even if this
// client's jar holds one — for scenarios asserting behavior when a
// required cookie is simply missing.
func (c *Client) DoNoCookies(method, path string) (*Response, error) {
	req, err := newJSONRequest(c.baseURL, method, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return doRequest(&http.Client{}, req)
}

// CookieValue returns the value of the named cookie this client's jar
// currently holds for rawURL, if any. Used to capture a refresh token
// before a rotating call replaces it, so a later step can replay the
// now-stale value.
func (c *Client) CookieValue(rawURL, name string) (string, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	for _, ck := range c.http.Jar.Cookies(u) {
		if ck.Name == name {
			return ck.Value, true
		}
	}
	return "", false
}

func newJSONRequest(baseURL, method, path string, body any, headers map[string]string) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func doRequest(client *http.Client, req *http.Request) (*Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{StatusCode: resp.StatusCode, Header: resp.Header, Body: data}, nil
}
