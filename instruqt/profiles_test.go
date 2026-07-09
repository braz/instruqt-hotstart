package instruqt

import (
	"testing"
	"time"
)

func TestApplyProfile_UnknownName(t *testing.T) {
	in := HotStartPoolInput{}
	_, err := ApplyProfile("nonexistent", &in, nil)
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestApplyProfile_FillsUnset(t *testing.T) {
	in := HotStartPoolInput{}
	_, err := ApplyProfile("webinar", &in, nil)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if in.AutoRefill == nil || *in.AutoRefill != true {
		t.Errorf("webinar should set auto_refill true, got %v", in.AutoRefill)
	}
	if in.Size == nil || *in.Size != 100 {
		t.Errorf("webinar default size should be 100, got %v", in.Size)
	}
}

func TestApplyProfile_ExplicitWins(t *testing.T) {
	in := HotStartPoolInput{
		AutoRefill: boolPtr(false),
		Size:       intPtr(7),
	}
	_, err := ApplyProfile("webinar", &in, intPtr(500))
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if *in.AutoRefill != false {
		t.Errorf("explicit auto_refill=false should win, got %v", *in.AutoRefill)
	}
	if *in.Size != 7 {
		t.Errorf("explicit size=7 should win, got %v", *in.Size)
	}
}

func TestApplyProfile_RegistrationsRatioCeil(t *testing.T) {
	in := HotStartPoolInput{}
	// webinar ratio 0.25 * 501 = 125.25 -> ceil 126
	_, err := ApplyProfile("webinar", &in, intPtr(501))
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if in.Size == nil || *in.Size != 126 {
		t.Errorf("size = %v, want 126", in.Size)
	}
}

func TestApplyProfile_FixedIgnoresRegistrations(t *testing.T) {
	in := HotStartPoolInput{}
	_, err := ApplyProfile("self-paced", &in, intPtr(1000))
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if in.Size == nil || *in.Size != 3 {
		t.Errorf("self-paced fixed size should be 3, got %v", in.Size)
	}
}

func TestApplyProfile_DerivesEndsAt(t *testing.T) {
	start := time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC)
	in := HotStartPoolInput{StartsAt: &start}
	_, err := ApplyProfile("webinar", &in, nil) // +30m offset
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if in.EndsAt == nil || !in.EndsAt.Equal(start.Add(30*time.Minute)) {
		t.Errorf("ends_at = %v, want start+30m", in.EndsAt)
	}
}

func TestApplyProfile_NoOffsetEmitsNote(t *testing.T) {
	in := HotStartPoolInput{}
	notes, err := ApplyProfile("self-paced", &in, nil)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if in.EndsAt != nil {
		t.Errorf("self-paced should not derive ends_at, got %v", in.EndsAt)
	}
	if len(notes) == 0 {
		t.Error("expected a manual-end-time note for self-paced")
	}
}
