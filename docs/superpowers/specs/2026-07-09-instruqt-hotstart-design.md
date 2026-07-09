# Instruqt Hot Start Pool CLI — Design

**Date:** 2026-07-09
**Status:** Approved (design), pending spec review

## 1. Purpose

A Go command-line tool for creating and inspecting Instruqt **hot start pools** —
pools of sandboxes provisioned before tracks start — via the Instruqt GraphQL
API. The tool targets both ad-hoc operator use and CI/scripted automation.

Scope for v1:

- **Create** a hot start pool (`createHotStartPool`).
- **List** pools for a team (`hotStartPools`, read `nodes`).
- **Get** a single pool by id (`hotStartPool`).

Delete and other lifecycle operations are explicitly out of scope for v1.

All operations scope by **team** (team slug). Organization-level fields and
arguments are deliberately excluded.

## 2. API Reference (verified against api-docs.instruqt.com)

- **Endpoint:** `POST https://play.instruqt.com/graphql`
- **Auth:** header `Authorization: Bearer <API_KEY>` (key from Settings → API keys)
- **Mutation:** `createHotStartPool(pool: HotStartPoolInput!): HotStartPool`
- **Query:** `hotStartPools(organizationSlug: String, teamSlug: String): HotStartPoolConnection`
  — we use `teamSlug` only and read `nodes: [HotStartPool!]!`.
- **Query:** `hotStartPool(id: String!): HotStartPool!`

### `HotStartPoolInput` fields

| Field | GraphQL type | Meaning |
|---|---|---|
| `type` | `HotStartPoolType` | `dedicated` \| `shared` |
| `tracks` | `[String!]` | track IDs |
| `configs` | `[String!]` | config IDs |
| `size` | `Int` | sandboxes per track |
| `name` | `String` | pool name |
| `auto_refill` | `Boolean` | auto-refill the pool |
| `starts_at` | `Time` | scheduled start (begin creating sandboxes) |
| `ends_at` | `Time` | scheduled stop (removes only unclaimed sandboxes) |
| `team_slug` | `String` | team creating the pool |
| `region` | `String` | region |
| `invite_id` | `String` | invite to scope the pool |

`organization_slug` exists in the schema but is **excluded** from this tool.

### `HotStartPool` (return) fields we consume

`id: ID!`, `type: HotStartPoolType!`, `size: Int!`, `created`, `name`,
`auto_refill`, `starts_at`, `ends_at`, `status`, `region`. Nested `team` /
`created_by` are reduced to the identifying fields we render (slug/name); we do
not over-fetch.

### Configuration vs. type (documentation note)

The best-practices doc describes "scheduled", "invite-scoped", and "always-hot"
pools. These are **configurations**, not the API `type` enum. They map onto
existing fields:

- scheduled → `starts_at` / `ends_at`
- invite-scoped → `invite_id`
- always-hot → no schedule + `auto_refill`

No schema change is needed; this is documented in the README so users are not
confused between configuration intent and the `dedicated`/`shared` enum.

## 3. Architecture

Cobra/Viper CLI wrapping a small hand-written GraphQL client. `main` is the
wiring point; the `instruqt` package has no knowledge of the CLI.

```
instruqt-hotstart/
├── main.go                 # tiny: calls cmd.Execute()
├── cmd/                    # CLI surface only, no business logic
│   ├── root.go             # root cmd, persistent flags, viper binding, client construction
│   ├── create.go           # create pool (flags, profiles, --dry-run, --force)
│   ├── list.go             # list pools for a team
│   ├── get.go              # get one pool by id
│   └── render.go           # table vs --json output helpers
├── instruqt/               # domain package: GraphQL client + types
│   ├── client.go           # Client struct, New(...), execute() over net/http
│   ├── hotstart.go         # CreateHotStartPool / HotStartPools / HotStartPool + types
│   ├── profiles.go         # event-type profile table + resolution
│   └── *_test.go           # httptest-based tests
├── config.example.yaml
├── .env.example
├── Makefile
├── go.mod / go.sum
```

**Boundary:** `cmd/` knows Cobra/Viper/output formatting; `instruqt/` knows
GraphQL and domain rules. No sideways imports; no `organization*` anywhere.

## 4. The `instruqt` package

### Client

```go
type Client struct {
    endpoint   string        // default https://play.instruqt.com/graphql
    apiKey     string
    httpClient *http.Client  // default 30s timeout
}

func New(apiKey string, opts ...Option) *Client
// Options: WithEndpoint(url), WithHTTPClient(c)
```

Transport — one unexported method:

```go
func (c *Client) execute(ctx context.Context, query string, vars any, out any) error
```

POSTs `{"query":..., "variables":...}` with the bearer header; decodes the
GraphQL envelope `{ "data":..., "errors":[...] }`. Non-empty `errors` (even on
HTTP 200) are joined with `errors.Join` and returned, wrapped with context.
Non-2xx HTTP status becomes an error including status + truncated body snippet.

### Input type

Pointers for every optional field so `omitempty` distinguishes "unset" from
"zero", matching nullable GraphQL fields.

```go
type PoolType string // "dedicated" | "shared"

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
```

### Output type

`HotStartPool` mirrors the consumed return fields (§2). Nested objects reduced
to identifying fields.

### Methods

```go
func (c *Client) CreateHotStartPool(ctx context.Context, in HotStartPoolInput) (*HotStartPool, error)
func (c *Client) HotStartPools(ctx context.Context, teamSlug string) ([]HotStartPool, error) // reads .nodes
func (c *Client) HotStartPool(ctx context.Context, id string) (*HotStartPool, error)
```

GraphQL queries are `const` strings colocated in `hotstart.go`.

### Validation

Pure, no I/O, table-testable:

```go
func (in HotStartPoolInput) Validate() (warnings []string, err error)
```

- **Hard errors** (`err != nil`; `create` aborts unless `--force`):
  - `ends_at` before `starts_at`
  - `size <= 0` (when set)
  - missing required `type`
- **Warnings** (printed, never block):
  - no `ends_at` set → "indefinite pool, bills continuously"
  - `starts_at` less than 1h out → "below recommended provisioning lead time"
  - `size >= 100` with `starts_at` under ~2h → "large pools provision in
    batches, allow more lead time"

Thresholds (1h lead, 100 size, 2h large-pool lead) are named constants so they
are easy to adjust as guidance evolves.

### Profiles

`profiles.go` holds a small central table mapping profile name → defaults. A
profile fills only fields the user left unset (explicit flag always wins) and
suggests `size` from `--registrations` via a ratio.

| Profile | auto_refill | end offset | default size (no `--registrations`) | size ratio (with `--registrations`) |
|---|---|---|---|---|
| `self-paced` | on | none (always-hot) | 3 | n/a (fixed) |
| `live-workshop` | off | +30m | 20 | 0.70 |
| `webinar` | on | +30m | 100 | 0.25 |
| `multi-day` | off | +30m | 20 | 1.0 |
| `conference-session` | off | +45m | 80 | 0.75 |
| `booth-demo` | on | none | 4 | n/a (fixed) |
| `sales-demo` | on | none | 2 | n/a (fixed) |

Semantics of the columns:

- **end offset** — when set and `starts_at` is known, the profile derives
  `ends_at = starts_at + offset` (only if the user did not set `--ends-at`).
  `none` means the profile leaves `ends_at` unset and, for `booth-demo` /
  `sales-demo` / `self-paced`, prints a note that the operator should set an end
  time manually (these are event-shaped-but-open cases the doc treats as
  judgement calls).
- **default size** — used when `--registrations` is not supplied and the user
  did not set `--size`.
- **size ratio** — when `--registrations=N` is supplied and `--size` is unset,
  suggested `size = ceil(N × ratio)`. Profiles marked `n/a (fixed)` ignore
  `--registrations` and use the default size.

Resolution order per field: explicit flag → profile-derived → config file →
unset. Unknown profile name is an error. These values encode current
best-practice guidance and are expected to need occasional upkeep; keeping them
in one table makes that cheap.

## 5. CLI (`cmd`)

### Persistent flags & Viper precedence

Precedence: explicit flag > env > config file > default.

| Setting | Flag | Env | Notes |
|---|---|---|---|
| API key | `--api-key` | `INSTRUQT_API_KEY` | secret; env preferred |
| Team slug | `--team` | `INSTRUQT_TEAM` | scopes all commands |
| Endpoint | `--endpoint` | `INSTRUQT_ENDPOINT` | default play.instruqt.com/graphql |
| Config file | `--config` | — | default `./config.yaml` if present |
| Output | `--json` | — | table default |

### `create` flags

Map to `HotStartPoolInput`, all overridable: `--type`, `--size`, `--tracks`
(repeatable/CSV), `--configs`, `--name`, `--auto-refill`, `--starts-at`
(RFC3339 or relative e.g. `+1h`), `--ends-at`, `--region`, `--invite-id`, plus
`--profile`, `--registrations`, `--dry-run`, `--force`.

Flow: build input → apply profile to unset fields → `Validate()` → print
warnings to stderr → abort on hard error unless `--force` → send (or, with
`--dry-run`, print resolved payload + warnings and exit without sending).

### Output (`render.go`)

`list`/`get` render a `text/tabwriter` table (id, name, type, size, status,
starts_at, ends_at) by default; `--json` emits indented JSON. `create` prints
the created pool the same way. Results go to stdout; warnings/diagnostics to
stderr so `--json` stays pipeable.

## 6. Error handling

- Client wraps failures with context via `fmt.Errorf("...: %w", err)`.
- GraphQL `errors[]` joined with `errors.Join`, even on HTTP 200.
- Non-2xx HTTP → error with status + truncated body.
- `context.Context` threaded through all methods; root sets a signal-cancelled
  (Ctrl-C) context with a default timeout.
- Commands return errors to Cobra (prints + exit code); no `log.Fatal` in logic.
- Missing API key or team → clear actionable error before any network call.

## 7. Testing

- `instruqt/client_test.go` — `httptest.Server` with canned GraphQL envelopes;
  table-driven: success, GraphQL `errors[]`, non-2xx, malformed JSON. Assert
  request body query/variables and `Authorization` header.
- `instruqt/hotstart_test.go` — marshal round-trip: unset pointers absent,
  `team_slug` present, no `organization_*` keys ever; `go-cmp` for diffs.
- `Validate()` — pure table tests: each hard error, each warning, clean case,
  boundaries (ends before starts, size 0, starts_at just under/over 1h).
- Profiles — table tests: explicit flag beats profile, profile fills unset,
  registrations × ratio math, unknown profile errors.
- `t.Helper()` in shared assertion helpers. No mocking framework, no BDD.
- Config precedence exercised via one integration-style test using a temp file.

## 8. Dependencies

- `github.com/spf13/cobra`, `github.com/spf13/viper` — CLI + config.
- `github.com/google/go-cmp` — test diffs.
- Standard library for HTTP/JSON/GraphQL transport. No GraphQL client library.

## 9. Out of scope (v1)

- Delete / update pool operations.
- Organization-scoped queries or fields.
- Pagination beyond `nodes` (assume result set fits one response for v1;
  revisit if `pageInfo` proves necessary).
- Filesystem abstraction (afero) — not needed yet.
