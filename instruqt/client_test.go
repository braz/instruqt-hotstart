package instruqt

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// echoServer captures the last request and returns a canned body/status.
type capturedRequest struct {
	auth      string
	query     string
	variables map[string]any
}

func newTestClient(t *testing.T, status int, body string, captured *capturedRequest) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if captured != nil {
			captured.auth = r.Header.Get("Authorization")
			raw, _ := io.ReadAll(r.Body)
			var env struct {
				Query     string         `json:"query"`
				Variables map[string]any `json:"variables"`
			}
			_ = json.Unmarshal(raw, &env)
			captured.query = env.Query
			captured.variables = env.Variables
		}
		w.WriteHeader(status)
		io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return New("test-key", WithEndpoint(srv.URL), WithHTTPClient(srv.Client()))
}

func TestExecute_Success(t *testing.T) {
	var cap capturedRequest
	c := newTestClient(t, http.StatusOK, `{"data":{"ping":"pong"}}`, &cap)

	var out struct {
		Ping string `json:"ping"`
	}
	err := c.execute(context.Background(), "query { ping }", map[string]any{"x": 1}, &out)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if out.Ping != "pong" {
		t.Errorf("ping = %q, want pong", out.Ping)
	}
	if cap.auth != "Bearer test-key" {
		t.Errorf("auth = %q, want Bearer test-key", cap.auth)
	}
	if cap.query != "query { ping }" {
		t.Errorf("query = %q", cap.query)
	}
	if cap.variables["x"].(float64) != 1 {
		t.Errorf("variables = %v", cap.variables)
	}
}

func TestExecute_GraphQLErrors(t *testing.T) {
	c := newTestClient(t, http.StatusOK,
		`{"errors":[{"message":"boom"},{"message":"bang"}]}`, nil)

	var out map[string]any
	err := c.execute(context.Background(), "q", nil, &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "boom") || !strings.Contains(err.Error(), "bang") {
		t.Errorf("error should contain both messages: %v", err)
	}
}

func TestExecute_HTTPError(t *testing.T) {
	c := newTestClient(t, http.StatusUnauthorized, `unauthorized`, nil)

	var out map[string]any
	err := c.execute(context.Background(), "q", nil, &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention status 401: %v", err)
	}
}

func TestExecute_MalformedJSON(t *testing.T) {
	c := newTestClient(t, http.StatusOK, `{not json`, nil)

	var out map[string]any
	err := c.execute(context.Background(), "q", nil, &out)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}
