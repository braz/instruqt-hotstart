package instruqt

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHotStartPoolInputMarshal_OmitsUnset(t *testing.T) {
	in := HotStartPoolInput{
		Type:     PoolTypeShared,
		TeamSlug: strPtr("demo-team"),
	}

	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Present fields.
	if got["type"] != "shared" {
		t.Errorf("type = %v, want shared", got["type"])
	}
	if got["team_slug"] != "demo-team" {
		t.Errorf("team_slug = %v, want demo-team", got["team_slug"])
	}

	// Unset optional pointers must be absent, not null/zero.
	for _, k := range []string{"size", "name", "auto_refill", "starts_at", "ends_at", "region", "invite_id", "tracks", "configs"} {
		if _, ok := got[k]; ok {
			t.Errorf("expected key %q to be omitted, got %v", k, got[k])
		}
	}

	// organization_* must never appear.
	for k := range got {
		if k == "organization_slug" || k == "organization" {
			t.Errorf("unexpected organization key %q in payload", k)
		}
	}
}

func TestHotStartPoolInputMarshal_AllFields(t *testing.T) {
	start := time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	in := HotStartPoolInput{
		Type:       PoolTypeDedicated,
		Tracks:     []string{"t1", "t2"},
		SandboxIDs: []string{"c1"},
		Size:       intPtr(50),
		Name:       strPtr("workshop"),
		AutoRefill: boolPtr(true),
		StartsAt:   &start,
		EndsAt:     &end,
		TeamSlug:   strPtr("demo-team"),
		Region:     strPtr("europe-west1"),
		InviteID:   strPtr("inv123"),
	}

	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got["size"].(float64) != 50 {
		t.Errorf("size = %v, want 50", got["size"])
	}
	if got["auto_refill"] != true {
		t.Errorf("auto_refill = %v, want true", got["auto_refill"])
	}
	if got["starts_at"] != "2026-07-09T10:00:00Z" {
		t.Errorf("starts_at = %v", got["starts_at"])
	}
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }
