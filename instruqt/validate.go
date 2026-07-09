package instruqt

import (
	"errors"
	"fmt"
	"time"
)

// leadTimeTier maps an upper size bound to the recommended provisioning lead
// time. Larger pools provision in batches and need more lead time. Tiers are
// evaluated in order; the first matching maxSize (inclusive) wins.
type leadTimeTier struct {
	maxSize int // inclusive upper bound; 0 means "no upper bound"
	lead    time.Duration
}

var leadTimeTiers = []leadTimeTier{
	{maxSize: 49, lead: 20 * time.Minute},
	{maxSize: 100, lead: 30 * time.Minute},
	{maxSize: 199, lead: 1 * time.Hour},
	{maxSize: 400, lead: 90 * time.Minute},
	{maxSize: 0, lead: 90 * time.Minute}, // >400: 90m minimum
}

func requiredLeadTime(size int) time.Duration {
	for _, tier := range leadTimeTiers {
		if tier.maxSize == 0 || size <= tier.maxSize {
			return tier.lead
		}
	}
	return 90 * time.Minute
}

// Validate checks a pool input against hard rules and cost/timing best
// practices. It returns non-blocking warnings and, separately, a hard error
// that callers should treat as blocking (unless the user forces past it).
//
// now is passed explicitly so the function stays pure and testable.
func (in HotStartPoolInput) Validate(now time.Time) (warnings []string, err error) {
	var errs []error

	if in.Type == "" {
		errs = append(errs, errors.New("type is required (dedicated or shared)"))
	}
	if in.Name == nil || *in.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if in.Size != nil && *in.Size <= 0 {
		errs = append(errs, fmt.Errorf("size must be positive, got %d", *in.Size))
	}
	if in.StartsAt != nil && in.EndsAt != nil && in.EndsAt.Before(*in.StartsAt) {
		errs = append(errs, fmt.Errorf("ends_at (%s) is before starts_at (%s)",
			in.EndsAt.Format(time.RFC3339), in.StartsAt.Format(time.RFC3339)))
	}

	if in.EndsAt == nil {
		warnings = append(warnings, "no ends_at set: this is an indefinite pool that bills continuously until deleted")
	}

	if in.StartsAt != nil && in.Size != nil && *in.Size > 0 {
		lead := requiredLeadTime(*in.Size)
		if in.StartsAt.Sub(now) < lead {
			warnings = append(warnings, fmt.Sprintf(
				"starts_at is under the recommended %s provisioning lead time for a pool of size %d",
				lead, *in.Size))
		}
	}

	return warnings, errors.Join(errs...)
}
