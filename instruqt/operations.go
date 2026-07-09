package instruqt

import (
	"context"
	"fmt"
)

// poolFields is the shared selection set for a HotStartPool.
const poolFields = `
	id
	type
	size
	created
	name
	auto_refill
	starts_at
	ends_at
	status
	region
`

const createHotStartPoolMutation = `mutation createHotStartPool($pool: HotStartPoolInput!) {
	createHotStartPool(pool: $pool) {` + poolFields + `}
}`

const hotStartPoolQuery = `query hotStartPool($id: String!) {
	hotStartPool(id: $id) {` + poolFields + `}
}`

const hotStartPoolsQuery = `query hotStartPools($teamSlug: String, $paging: Pagination) {
	hotStartPools(teamSlug: $teamSlug, paging: $paging) {
		nodes {` + poolFields + `}
		pageInfo { endCursor hasNextPage }
	}
}`

const sandboxConfigsQuery = `query sandboxConfigs($teamSlug: String) {
	sandboxConfigs(teamSlug: $teamSlug) {
		id
		slug
		name
		version
	}
}`

// defaultPageSize is the number of pools requested per page when listing.
const defaultPageSize = 100

// SandboxConfig identifies a sandbox config. Its ID is what a hot start pool's
// `configs` field expects (the API resolves it to a config version).
type SandboxConfig struct {
	ID      string `json:"id"`
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Version int    `json:"version"`
}

// SandboxConfigs lists the sandbox configs for a team. Use these IDs with the
// create command's --configs flag.
func (c *Client) SandboxConfigs(ctx context.Context, teamSlug string) ([]SandboxConfig, error) {
	var out struct {
		SandboxConfigs []SandboxConfig `json:"sandboxConfigs"`
	}
	if err := c.execute(ctx, sandboxConfigsQuery, map[string]any{"teamSlug": teamSlug}, &out); err != nil {
		return nil, fmt.Errorf("listing sandbox configs for team %s: %w", teamSlug, err)
	}
	return out.SandboxConfigs, nil
}

// CreateHotStartPool creates a new hot start pool and returns it.
func (c *Client) CreateHotStartPool(ctx context.Context, in HotStartPoolInput) (*HotStartPool, error) {
	var out struct {
		CreateHotStartPool HotStartPool `json:"createHotStartPool"`
	}
	if err := c.execute(ctx, createHotStartPoolMutation, map[string]any{"pool": in}, &out); err != nil {
		return nil, fmt.Errorf("creating hot start pool: %w", err)
	}
	return &out.CreateHotStartPool, nil
}

// HotStartPool fetches a single hot start pool by id.
func (c *Client) HotStartPool(ctx context.Context, id string) (*HotStartPool, error) {
	var out struct {
		HotStartPool HotStartPool `json:"hotStartPool"`
	}
	if err := c.execute(ctx, hotStartPoolQuery, map[string]any{"id": id}, &out); err != nil {
		return nil, fmt.Errorf("getting hot start pool %s: %w", id, err)
	}
	return &out.HotStartPool, nil
}

// HotStartPools returns every hot start pool for the given team, following
// forward cursor pagination until there are no more pages.
func (c *Client) HotStartPools(ctx context.Context, teamSlug string) ([]HotStartPool, error) {
	var all []HotStartPool
	var after string

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		paging := map[string]any{"First": defaultPageSize}
		if after != "" {
			paging["After"] = after
		}

		var out struct {
			HotStartPools struct {
				Nodes    []HotStartPool `json:"nodes"`
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"hotStartPools"`
		}

		vars := map[string]any{"teamSlug": teamSlug, "paging": paging}
		if err := c.execute(ctx, hotStartPoolsQuery, vars, &out); err != nil {
			return nil, fmt.Errorf("listing hot start pools for team %s: %w", teamSlug, err)
		}

		all = append(all, out.HotStartPools.Nodes...)
		if !out.HotStartPools.PageInfo.HasNextPage {
			break
		}
		after = out.HotStartPools.PageInfo.EndCursor
	}

	return all, nil
}
