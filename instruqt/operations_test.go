package instruqt

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSandboxConfigs(t *testing.T) {
	var cap capturedRequest
	c := newTestClient(t, http.StatusOK,
		`{"data":{"sandboxConfigs":[{"id":"cfg1","slug":"my-track","name":"My Track","version":3}]}}`,
		&cap)

	configs, err := c.SandboxConfigs(context.Background(), "demo")
	if err != nil {
		t.Fatalf("configs: %v", err)
	}
	if len(configs) != 1 || configs[0].ID != "cfg1" || configs[0].Slug != "my-track" || configs[0].Version != 3 {
		t.Errorf("unexpected configs: %+v", configs)
	}
	if cap.variables["teamSlug"] != "demo" {
		t.Errorf("teamSlug = %v, want demo", cap.variables["teamSlug"])
	}
}

func TestCreateHotStartPool(t *testing.T) {
	var cap capturedRequest
	c := newTestClient(t, http.StatusOK,
		`{"data":{"createHotStartPool":{"id":"4","type":"shared","size":50,"name":"ws","status":"creating"}}}`,
		&cap)

	pool, err := c.CreateHotStartPool(context.Background(), HotStartPoolInput{
		Type:     PoolTypeShared,
		TeamSlug: strPtr("demo"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if pool.ID != "4" || pool.Size != 50 || pool.Status != "creating" {
		t.Errorf("unexpected pool: %+v", pool)
	}
	// Variables must nest under "pool".
	if _, ok := cap.variables["pool"]; !ok {
		t.Errorf("expected variables.pool, got %v", cap.variables)
	}
}

func TestHotStartPool_ByID(t *testing.T) {
	var cap capturedRequest
	c := newTestClient(t, http.StatusOK,
		`{"data":{"hotStartPool":{"id":"7","type":"dedicated","size":3,"name":"demo","status":"ready"}}}`,
		&cap)

	pool, err := c.HotStartPool(context.Background(), "7")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if pool.ID != "7" {
		t.Errorf("id = %q, want 7", pool.ID)
	}
	if cap.variables["id"] != "7" {
		t.Errorf("variables.id = %v, want 7", cap.variables["id"])
	}
}

// TestHotStartPools_Paginates verifies the client follows endCursor across
// pages and accumulates all nodes, sending After on the second request.
func TestHotStartPools_Paginates(t *testing.T) {
	page := 0
	var secondAfter any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var env struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.Unmarshal(raw, &env)
		w.WriteHeader(http.StatusOK)
		if page == 0 {
			page++
			io.WriteString(w, `{"data":{"hotStartPools":{"nodes":[{"id":"1"},{"id":"2"}],"pageInfo":{"endCursor":"CUR","hasNextPage":true}}}}`)
			return
		}
		if paging, ok := env.Variables["paging"].(map[string]any); ok {
			secondAfter = paging["After"]
		}
		io.WriteString(w, `{"data":{"hotStartPools":{"nodes":[{"id":"3"}],"pageInfo":{"endCursor":"","hasNextPage":false}}}}`)
	}))
	t.Cleanup(srv.Close)

	c := New("k", WithEndpoint(srv.URL), WithHTTPClient(srv.Client()))
	pools, err := c.HotStartPools(context.Background(), "demo")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(pools) != 3 {
		t.Fatalf("got %d pools, want 3", len(pools))
	}
	if secondAfter != "CUR" {
		t.Errorf("second page After = %v, want CUR", secondAfter)
	}
}
