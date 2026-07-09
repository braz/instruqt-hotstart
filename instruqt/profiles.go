package instruqt

import (
	"fmt"
	"math"
	"slices"
	"time"
)

// profile encodes best-practice defaults for an event type. A profile only
// fills fields the user left unset; explicit flags always win.
type profile struct {
	autoRefill  bool
	endOffset   time.Duration // 0 means "no scheduled end" — emit a manual note
	defaultSize int           // used when registrations is nil
	sizeRatio   float64       // 0 means fixed: ignore registrations, use defaultSize
}

// profiles is the central best-practice table. Values encode current guidance
// (see the Instruqt hot-starts best-practices doc) and are expected to need
// occasional upkeep; keeping them here makes that a one-file change.
var profiles = map[string]profile{
	"self-paced":         {autoRefill: true, endOffset: 0, defaultSize: 3},
	"live-workshop":      {autoRefill: false, endOffset: 30 * time.Minute, defaultSize: 20, sizeRatio: 0.70},
	"webinar":            {autoRefill: true, endOffset: 30 * time.Minute, defaultSize: 100, sizeRatio: 0.25},
	"multi-day":          {autoRefill: false, endOffset: 30 * time.Minute, defaultSize: 20, sizeRatio: 1.0},
	"conference-session": {autoRefill: false, endOffset: 45 * time.Minute, defaultSize: 80, sizeRatio: 0.75},
	"booth-demo":         {autoRefill: true, endOffset: 0, defaultSize: 4},
	"sales-demo":         {autoRefill: true, endOffset: 0, defaultSize: 2},
}

// ProfileNames returns the available profile names, sorted.
func ProfileNames() []string {
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// ApplyProfile fills unset fields of in from the named profile's defaults.
// registrations, when non-nil, drives the suggested size via the profile's
// ratio (fixed profiles ignore it). It returns advisory notes and an error for
// an unknown profile name.
func ApplyProfile(name string, in *HotStartPoolInput, registrations *int) (notes []string, err error) {
	p, ok := profiles[name]
	if !ok {
		return nil, fmt.Errorf("unknown profile %q (available: %v)", name, ProfileNames())
	}

	if in.AutoRefill == nil {
		v := p.autoRefill
		in.AutoRefill = &v
	}

	if in.Size == nil {
		size := p.defaultSize
		if p.sizeRatio > 0 && registrations != nil {
			size = int(math.Ceil(float64(*registrations) * p.sizeRatio))
		}
		in.Size = &size
	}

	if in.EndsAt == nil {
		if p.endOffset > 0 && in.StartsAt != nil {
			end := in.StartsAt.Add(p.endOffset)
			in.EndsAt = &end
		} else if p.endOffset == 0 {
			notes = append(notes, fmt.Sprintf(
				"profile %q has no scheduled end: set --ends-at manually or this pool will bill continuously", name))
		}
	}

	return notes, nil
}
