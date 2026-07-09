package instruqt

import "time"

// PoolType is the Instruqt HotStartPoolType enum.
type PoolType string

const (
	PoolTypeDedicated PoolType = "dedicated"
	PoolTypeShared    PoolType = "shared"
)

// HotStartPoolInput is the argument to the createHotStartPool mutation.
//
// Optional fields are pointers so that omitempty distinguishes "unset" from a
// meaningful zero value, matching the nullable fields in the GraphQL schema.
// organization_slug exists in the schema but is intentionally excluded — this
// tool scopes exclusively by team.
type HotStartPoolInput struct {
	Type       PoolType   `json:"type,omitempty"`
	Tracks     []string   `json:"tracks,omitempty"`
	Configs    []string   `json:"configs,omitempty"`
	Size       *int       `json:"size,omitempty"`
	Name       *string    `json:"name,omitempty"`
	AutoRefill *bool      `json:"auto_refill,omitempty"`
	StartsAt   *time.Time `json:"starts_at,omitempty"`
	EndsAt     *time.Time `json:"ends_at,omitempty"`
	TeamSlug   *string    `json:"team_slug,omitempty"`
	Region     *string    `json:"region,omitempty"`
	InviteID   *string    `json:"invite_id,omitempty"`
}

// HotStartPool is the subset of the HotStartPool return type this tool renders.
type HotStartPool struct {
	ID         string     `json:"id"`
	Type       PoolType   `json:"type"`
	Size       int        `json:"size"`
	Created    *time.Time `json:"created,omitempty"`
	Name       string     `json:"name"`
	AutoRefill bool       `json:"auto_refill"`
	StartsAt   *time.Time `json:"starts_at,omitempty"`
	EndsAt     *time.Time `json:"ends_at,omitempty"`
	Status     string     `json:"status"`
	Region     string     `json:"region,omitempty"`
}
