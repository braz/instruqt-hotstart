package instruqt

import (
	"strings"
	"testing"
	"time"
)

var refNow = time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)

func TestValidate_HardErrors(t *testing.T) {
	start := refNow.Add(2 * time.Hour)
	tests := []struct {
		name string
		in   HotStartPoolInput
	}{
		{"missing type", HotStartPoolInput{Size: intPtr(10)}},
		{"size zero", HotStartPoolInput{Type: PoolTypeShared, Size: intPtr(0)}},
		{"size negative", HotStartPoolInput{Type: PoolTypeShared, Size: intPtr(-1)}},
		{"ends before starts", HotStartPoolInput{
			Type:     PoolTypeShared,
			StartsAt: &start,
			EndsAt:   ptrTime(start.Add(-time.Hour)),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.in.Validate(refNow)
			if err == nil {
				t.Fatalf("expected hard error, got nil")
			}
		})
	}
}

func TestValidate_Clean(t *testing.T) {
	start := refNow.Add(2 * time.Hour)
	in := HotStartPoolInput{
		Type:     PoolTypeShared,
		Size:     intPtr(10),
		StartsAt: &start,
		EndsAt:   ptrTime(start.Add(30 * time.Minute)),
	}
	warnings, err := in.Validate(refNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func TestValidate_NoEndsAtWarns(t *testing.T) {
	start := refNow.Add(2 * time.Hour)
	in := HotStartPoolInput{Type: PoolTypeShared, Size: intPtr(10), StartsAt: &start}
	warnings, err := in.Validate(refNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasWarning(warnings, "indefinite") {
		t.Errorf("expected indefinite-pool warning, got %v", warnings)
	}
}

func TestValidate_LeadTimeTiers(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		lead     time.Duration
		wantWarn bool
	}{
		{"size49 ok at 20m", 49, 20 * time.Minute, false},
		{"size49 warn under 20m", 49, 19 * time.Minute, true},
		{"size50 ok at 30m", 50, 30 * time.Minute, false},
		{"size50 warn at 25m", 50, 25 * time.Minute, true},
		{"size100 ok at 30m", 100, 30 * time.Minute, false},
		{"size101 warn at 30m", 101, 30 * time.Minute, true},
		{"size101 ok at 60m", 101, 60 * time.Minute, false},
		{"size199 warn at 59m", 199, 59 * time.Minute, true},
		{"size200 warn at 60m", 200, 60 * time.Minute, true},
		{"size200 ok at 90m", 200, 90 * time.Minute, false},
		{"size400 ok at 90m", 400, 90 * time.Minute, false},
		{"size401 warn at 89m", 401, 89 * time.Minute, true},
		{"size401 ok at 90m", 401, 90 * time.Minute, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := refNow.Add(tt.lead)
			in := HotStartPoolInput{
				Type:     PoolTypeShared,
				Size:     intPtr(tt.size),
				StartsAt: &start,
				EndsAt:   ptrTime(start.Add(30 * time.Minute)),
			}
			warnings, err := in.Validate(refNow)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := hasWarning(warnings, "lead time")
			if got != tt.wantWarn {
				t.Errorf("lead-time warning = %v, want %v (warnings=%v)", got, tt.wantWarn, warnings)
			}
		})
	}
}

func hasWarning(ws []string, substr string) bool {
	for _, w := range ws {
		if strings.Contains(strings.ToLower(w), strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

func ptrTime(t time.Time) *time.Time { return &t }
