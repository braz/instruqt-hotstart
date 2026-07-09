package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eoinbrazil/instruqt-hotstart/instruqt"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a hot start pool",
		RunE:  runCreate,
	}
	f := c.Flags()
	f.String("type", "", "pool type: dedicated or shared")
	f.Int("size", 0, "sandboxes per track")
	f.StringSlice("tracks", nil, "track IDs (repeatable or comma-separated)")
	f.StringSlice("configs", nil, "config IDs (repeatable or comma-separated)")
	f.String("name", "", "pool name (required)")
	f.Bool("auto-refill", false, "auto-refill the pool")
	f.String("starts-at", "", "scheduled start: RFC3339 or relative (e.g. +1h)")
	f.String("ends-at", "", "scheduled end: RFC3339 or relative (e.g. +90m)")
	f.String("region", "", "region")
	f.String("invite-id", "", "invite ID to scope the pool")
	f.String("profile", "", fmt.Sprintf("best-practice profile %v", instruqt.ProfileNames()))
	f.Int("registrations", 0, "expected registrations (drives profile sizing)")
	f.Bool("dry-run", false, "print the resolved payload without sending")
	f.Bool("force", false, "proceed despite validation errors")

	_ = c.MarkFlagRequired("name")

	rootCmd.AddCommand(c)
}

func runCreate(cmd *cobra.Command, _ []string) error {
	f := cmd.Flags()
	now := time.Now()

	in, err := buildInput(cmd, now)
	if err != nil {
		return err
	}

	// Team slug: explicit flag falls back to the persistent --team/env.
	if in.TeamSlug == nil {
		if team := v.GetString("team"); team != "" {
			in.TeamSlug = &team
		}
	}

	// Apply profile to unset fields.
	if profile, _ := f.GetString("profile"); profile != "" {
		regs := intFlagPtr(cmd, "registrations")
		notes, err := instruqt.ApplyProfile(profile, &in, regs)
		if err != nil {
			return err
		}
		for _, n := range notes {
			fmt.Fprintln(os.Stderr, "note:", n)
		}
	}

	// Validate.
	warnings, verr := in.Validate(now)
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	force, _ := f.GetBool("force")
	if verr != nil && !force {
		return fmt.Errorf("validation failed (use --force to override): %w", verr)
	}
	if verr != nil {
		fmt.Fprintln(os.Stderr, "warning: proceeding despite validation errors:", verr)
	}

	asJSON, _ := f.GetBool("json")
	if dry, _ := f.GetBool("dry-run"); dry {
		fmt.Fprintln(os.Stderr, "dry-run: not sending. Resolved payload:")
		return writeJSON(cmd.OutOrStdout(), in)
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	ctx, cancel := signalContext()
	defer cancel()

	pool, err := client.CreateHotStartPool(ctx, in)
	if err != nil {
		return err
	}
	return renderPool(cmd.OutOrStdout(), pool, asJSON)
}

// buildInput maps only the flags the user actually set into a HotStartPoolInput.
func buildInput(cmd *cobra.Command, now time.Time) (instruqt.HotStartPoolInput, error) {
	f := cmd.Flags()
	var in instruqt.HotStartPoolInput

	if f.Changed("type") {
		t, _ := f.GetString("type")
		switch instruqt.PoolType(t) {
		case instruqt.PoolTypeDedicated, instruqt.PoolTypeShared:
			in.Type = instruqt.PoolType(t)
		default:
			return in, fmt.Errorf("invalid --type %q: must be dedicated or shared", t)
		}
	}
	if f.Changed("size") {
		in.Size = intFlagPtr(cmd, "size")
	}
	if f.Changed("tracks") {
		in.Tracks, _ = f.GetStringSlice("tracks")
	}
	if f.Changed("configs") {
		in.Configs, _ = f.GetStringSlice("configs")
	}
	if f.Changed("name") {
		s, _ := f.GetString("name")
		in.Name = &s
	}
	if f.Changed("auto-refill") {
		b, _ := f.GetBool("auto-refill")
		in.AutoRefill = &b
	}
	if f.Changed("region") {
		s, _ := f.GetString("region")
		in.Region = &s
	}
	if f.Changed("invite-id") {
		s, _ := f.GetString("invite-id")
		in.InviteID = &s
	}
	if f.Changed("starts-at") {
		s, _ := f.GetString("starts-at")
		t, err := parseTimeFlag(s, now)
		if err != nil {
			return in, fmt.Errorf("--starts-at: %w", err)
		}
		in.StartsAt = &t
	}
	if f.Changed("ends-at") {
		s, _ := f.GetString("ends-at")
		t, err := parseTimeFlag(s, now)
		if err != nil {
			return in, fmt.Errorf("--ends-at: %w", err)
		}
		in.EndsAt = &t
	}
	return in, nil
}

func intFlagPtr(cmd *cobra.Command, name string) *int {
	val, _ := cmd.Flags().GetInt(name)
	return &val
}

// parseTimeFlag accepts an RFC3339 timestamp or a relative offset like "+90m".
func parseTimeFlag(s string, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "+") {
		d, err := time.ParseDuration(strings.TrimPrefix(s, "+"))
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid relative duration %q: %w", s, err)
		}
		return now.Add(d), nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected RFC3339 or +duration, got %q", s)
	}
	return t, nil
}
