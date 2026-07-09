package instruqt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultEndpoint is the Instruqt GraphQL API endpoint.
const DefaultEndpoint = "https://play.instruqt.com/graphql"

// Client talks to the Instruqt GraphQL API. Construct it with New.
type Client struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithEndpoint overrides the GraphQL endpoint URL.
func WithEndpoint(url string) Option {
	return func(c *Client) { c.endpoint = url }
}

// WithHTTPClient overrides the underlying *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// New returns a Client authenticated with the given API key.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		endpoint:   DefaultEndpoint,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// gqlError is one entry in a GraphQL response "errors" array.
type gqlError struct {
	Message string `json:"message"`
}

func (e gqlError) Error() string { return e.Message }

// execute sends a GraphQL request and decodes data into out.
func (c *Client) execute(ctx context.Context, query string, vars any, out any) error {
	reqBody, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("graphql http status %d: %s", resp.StatusCode, truncate(string(body), 256))
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []gqlError      `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	if len(envelope.Errors) > 0 {
		errs := make([]error, len(envelope.Errors))
		for i, e := range envelope.Errors {
			errs[i] = e
		}
		return fmt.Errorf("graphql errors: %w", errors.Join(errs...))
	}

	if out != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("decoding data: %w", err)
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
