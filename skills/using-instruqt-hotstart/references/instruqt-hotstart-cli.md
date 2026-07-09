# instruqt-hotstart CLI reference

Complete flag, validation, and profile reference for the `instruqt-hotstart`
CLI. For the workflow and common mistakes, see the parent
[SKILL.md](../SKILL.md). Every command supports `--help`.

Binary: `instruqt-hotstart` (build with `make build` → `go build -o
instruqt-hotstart .`). No flag has a shorthand.

## Global (persistent) flags

| Flag | Default | Meaning |
|------|---------|---------|
| `--api-key` | `""` | Instruqt API key (or env `INSTRUQT_API_KEY`). Required; the CLI errors before any network call if unset. |
| `--team` | `""` | Team slug scoping operations (or env `INSTRUQT_TEAM`). Required for `create`, `list`, `sandboxes`; not used by `get`. |
| `--endpoint` | `https://play.instruqt.com/graphql` | GraphQL endpoint (or env `INSTRUQT_ENDPOINT`). |
| `--config` | `""` | Config file (defaults to `./config.yaml` if present). |
| `--timeout` | `30s` | Per-request HTTP timeout (or env `INSTRUQT_TIMEOUT`). No whole-command deadline; SIGINT cancels. |
| `--json` | `false` | Output JSON instead of a table. |

## Configuration resolution

Precedence: **flag > env > config file > default** (Viper `BindPFlags` +
`AutomaticEnv`).

- Env prefix `INSTRUQT`, with a hyphen→underscore replacer, so hyphenated keys
  map to underscored env vars:
  - `api-key` → `INSTRUQT_API_KEY`
  - `team` → `INSTRUQT_TEAM`
  - `endpoint` → `INSTRUQT_ENDPOINT`
  - `timeout` → `INSTRUQT_TIMEOUT`
- Config file: `--config <path>`, else `./config.yaml` (missing file is not an
  error). See `config.example.yaml`.
- **`.env` must use `export`** on each line, or the CLI (a child process) can't
  see the values (misleading `no API key`). Alternatively `set -a; source .env;
  set +a`. See `env.example`.

## `create` — create a hot start pool

Requires a team (via `--team`/env). Flags:

| Flag | Type / default | Notes |
|------|----------------|-------|
| `--type` | string `""` | **Required.** `dedicated` or `shared`. Any other value is a hard error. |
| `--name` | string `""` | **Required** (cobra-required; not bypassable by `--force`). |
| `--size` | int `0` | Sandboxes per pool. Must be `> 0` (hard error otherwise). |
| `--sandbox-ids` | stringSlice `nil` | Sandbox IDs, e.g. `0bgr0ddoarzk`. Repeatable or comma-separated. Discover with `sandboxes`. |
| `--tracks` | stringSlice `nil` | **Do not use for creation** — the API rejects it (`tracks: must be blank; config_versions: cannot be blank`). CLI warns. |
| `--auto-refill` | bool `false` | Auto-refill the pool as sandboxes are consumed. |
| `--starts-at` | string `""` | RFC3339 (`2026-07-09T14:00:00Z`) or relative `+duration` (`+2h`, `+90m`). |
| `--ends-at` | string `""` | Same formats. Omitting it warns (indefinite, continuously-billing pool). |
| `--region` | string `""` | |
| `--invite-id` | string `""` | Invite ID to scope the pool. |
| `--profile` | string `""` | Best-practice event profile (see below). Names surfaced via `--help`. |
| `--registrations` | int `0` | Expected registrations; drives ratio-based profile sizing. |
| `--dry-run` | bool `false` | Resolve and print the payload (JSON to stdout) without sending. No spend. |
| `--force` | bool `false` | Proceed despite validation **errors** (cannot bypass required `--name`). |

Only flags you actually set are sent. On success prints the created pool
(`renderPool`, table or `--json`).

### Validation (`create`)

**Hard errors** (block unless `--force`, except `--name` which is always required):
- missing `--type`; missing `--name`; `--size <= 0`; `--ends-at` before `--starts-at`.

**Warnings** (print to stderr, never block):
- no `--ends-at` (indefinite pool).
- `--starts-at` closer than the recommended provisioning lead time for the size:

  | size | min lead time |
  |------|---------------|
  | `< 50` | 20m |
  | `50–100` | 30m |
  | `100 < size < 200` | 1h |
  | `200–400` | 90m |
  | `> 400` | 90m |

  The band is keyed off the **resolved `size`**, not raw `--registrations`. With
  a profile (e.g. `webinar` at 25%), 500 registrations resolves to size 125, so
  the `100 < size < 200` → 1h band applies — not the `>400` band.

With `--force` and a validation error, prints `warning: proceeding despite
validation errors: ...` and continues.

### Time flags

`--starts-at` / `--ends-at`: a value starting with `+` is parsed as a duration
added to now (`time.ParseDuration`: `+90m`, `+2h`); otherwise RFC3339. Errors:
`invalid relative duration` or `expected RFC3339 or +duration, got %q`.

### Profiles

`--profile <name>` applies best-practice defaults to **unset fields only**
(explicit flags always win). Profiles **never set `--type` or `--name`** — you
must still pass both. Profiles emit `note:` lines to stderr. `--registrations N`
drives ratio-based sizing (fixed-size profiles ignore it). The authoritative
list is printed in `create --help` via `ProfileNames()`; documented profiles
include `self-paced`, `live-workshop`, `webinar`, `multi-day`,
`conference-session`, `booth-demo`, `sales-demo` (e.g. `webinar` → auto-refill
on, `ends-at = starts-at + 30m`, ~25% registrations→size ratio).

Example:
```sh
instruqt-hotstart create --team acme --type shared --name spring-webinar \
  --profile webinar --registrations 500 --starts-at +2h --dry-run
# -> size 125, auto_refill on, ends_at = starts_at + 30m
```

## `list` — list a team's pools

No own flags. Requires `--team`. Output columns: `ID  NAME  TYPE  SIZE  STATUS
STARTS_AT  ENDS_AT` (or `--json` array).

## `get` — inspect one pool

| Flag | Default | Notes |
|------|---------|-------|
| `--id` | `""` | **Required.** Hot start pool ID. |

Does **not** require a team. There is no name-based lookup — get the ID from
`list` first (`list --json` + `jq` to script it). Output: single object/row.

## `sandboxes` — discover sandbox IDs

No own flags. Requires `--team`. Output columns: `ID  SLUG  NAME  VERSION` (or
`--json`). This is how you find the IDs to pass to `create --sandbox-ids`. There
is no server-side filter — match the SLUG/NAME yourself (or grep `--json`).

## Wire-contract note (for maintainers)

The CLI deliberately says "sandbox" (`--sandbox-ids`, `sandboxes`, `Sandbox`)
while the GraphQL wire still uses `configs` / `sandboxConfigs` (JSON tag
`json:"configs"`, operation `sandboxConfigs`). This mismatch is intentional —
users found "configs" confusing. Do not "fix" it.
