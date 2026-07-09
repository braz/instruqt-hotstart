# instruqt-hotstart

A small Go CLI to create and inspect [Instruqt](https://instruqt.com) **hot
start pools** — pools of sandboxes provisioned before tracks start — via the
Instruqt GraphQL API. It wraps `createHotStartPool`, `hotStartPools`, and
`hotStartPool` with cost-safety guardrails and best-practice profiles.

All operations scope by **team**.

## Install

```sh
go build -o instruqt-hotstart .
```

## Authentication & configuration

The CLI resolves settings with this precedence: **flag > env > config file > default**.

| Setting | Flag | Env | Default |
|---|---|---|---|
| API key | `--api-key` | `INSTRUQT_API_KEY` | — (required) |
| Team slug | `--team` | `INSTRUQT_TEAM` | — (required for create/list) |
| Endpoint | `--endpoint` | `INSTRUQT_ENDPOINT` | `https://play.instruqt.com/graphql` |
| Config file | `--config` | — | `./config.yaml` if present |
| Output | `--json` | — | table |

Generate an API key in the Instruqt web UI under **Settings → API keys**. Prefer
supplying it via `INSTRUQT_API_KEY` (see `env.example`). A `config.yaml`
(see `config.example.yaml`) can hold non-secret defaults like `team`.

> **Sourcing a `.env`?** The lines must use `export` (as `env.example` does).
> `source`-ing plain `KEY=value` lines only sets shell variables, which the CLI
> — a child process — cannot see, giving a misleading "no API key" error.
> Alternatively run `set -a; source .env; set +a`.

## Usage

```sh
# Find the sandbox IDs for your team (pools are keyed by sandbox, not track)
instruqt-hotstart sandboxes --team my-team

# Create a pool (always supply --type, --name, and --sandbox-ids)
instruqt-hotstart create --team my-team --type shared --name spring-workshop --size 50 \
  --sandbox-ids 0bgr0ddoarzk,1cfp2eepbsam --auto-refill --starts-at +2h --ends-at +150m

# Preview without sending
instruqt-hotstart create --team my-team --type shared --name demo --sandbox-ids 0bgr0ddoarzk \
  --size 250 --starts-at +45m --dry-run

# List and inspect pools
instruqt-hotstart list --team my-team
instruqt-hotstart get --id <pool-id> --json
```

`--starts-at` / `--ends-at` accept RFC3339 (`2026-07-09T14:00:00Z`) or a
relative offset (`+90m`, `+2h`).

### Pools use sandbox IDs, not tracks

The API creates hot start pools from **sandbox IDs** (e.g. `0bgr0ddoarzk`), not
track IDs — it rejects `--tracks` for pool creation (`tracks: must be blank;
config_versions: cannot be blank`). Run `instruqt-hotstart sandboxes --team
<team>` to list the sandboxes for your team (id, slug, name, version) and pass
the matching id(s) to `--sandbox-ids`. The `--tracks` flag remains for
completeness but the CLI warns when you use it on `create`.

### Cost guardrails

`create` validates the pool before sending. Hard errors (missing `type`,
missing `name`, `size <= 0`, `ends_at` before `starts_at`) block unless you pass
`--force`. `--name` is additionally a required flag enforced by the CLI itself,
so `--force` cannot bypass it. Warnings never block:

- **No `ends_at`** → indefinite pool that bills continuously.
- **Insufficient lead time** → `starts_at` is closer than the recommended
  provisioning lead time for the pool size:

  | Size | Recommended lead time |
  |---|---|
  | `< 50` | 20 minutes |
  | `50–100` | 30 minutes |
  | `100 < size < 200` | 1 hour |
  | `200–400` | 1 hour 30 minutes |
  | `> 400` | 1h30m minimum (test provisioning time) |

### Profiles

`--profile` pre-fills unset fields from event-type best practices; explicit
flags always win. With `--registrations N`, ratio-based profiles suggest a size;
fixed profiles ignore it. Note: profiles do **not** set `type` or `name` —
always pass `--type` and `--name` yourself.

| Profile | auto_refill | end offset | default size | size ratio |
|---|---|---|---|---|
| `self-paced` | on | none (always-hot) | 3 | fixed |
| `live-workshop` | off | +30m | 20 | 0.70 |
| `webinar` | on | +30m | 100 | 0.25 |
| `multi-day` | off | +30m | 20 | 1.0 |
| `conference-session` | off | +45m | 80 | 0.75 |
| `booth-demo` | on | none | 4 | fixed |
| `sales-demo` | on | none | 2 | fixed |

```sh
instruqt-hotstart create --team my-team --type shared --name spring-webinar \
  --profile webinar --registrations 500 --starts-at +2h --dry-run
# -> size 125, auto_refill on, ends_at = starts_at + 30m
```

### Pool configuration vs. `type`

The Instruqt API `type` enum is only `dedicated` | `shared`. The
best-practices doc's "scheduled", "invite-scoped", and "always-hot" pools are
*configurations* layered on top, mapped to existing fields:

- **scheduled** → `--starts-at` / `--ends-at`
- **invite-scoped** → `--invite-id`
- **always-hot** → no schedule + `--auto-refill`

See the Instruqt
[hot starts best practices](https://docs.instruqt.com/resources/hot-starts-best-practices)
for sizing and scheduling guidance; the profile values above encode it.

## Development

```sh
make test   # go test ./...
make vet    # go vet ./...
make fmt    # gofmt -w
make build
```

Design and rationale: `docs/superpowers/specs/2026-07-09-instruqt-hotstart-design.md`.
