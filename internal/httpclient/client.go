package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	authHeader string
	authValue  string
	http       *http.Client
}

func New(baseURL, authHeader, authValue string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		authHeader: authHeader,
		authValue:  authValue,
		http:       &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Get(path string, params url.Values, out any) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", u, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s returned %d: %s", u, resp.StatusCode, b)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) Delete(path string) error {
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DELETE %s returned %d: %s", path, resp.StatusCode, b)
	}
	return nil
}

func (c *Client) Put(path string, body []byte) error {
	req, err := http.NewRequest(http.MethodPut, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("PUT %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PUT %s returned %d: %s", path, resp.StatusCode, b)
	}
	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set(c.authHeader, c.authValue)
	req.Header.Set("Accept", "application/json")
}
