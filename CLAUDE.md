# CLAUDE.md

Context for working in this repo. Read this first.

## What this is

`instruqt-hotstart` — a Go CLI that creates and inspects Instruqt **hot start
pools** (pools of sandboxes warmed up before tracks start) via the Instruqt
GraphQL API. Module path: `github.com/eoinbrazil/instruqt-hotstart`.

- **API:** `POST https://play.instruqt.com/graphql`, header
  `Authorization: Bearer <key>`. Wraps `createHotStartPool` / `hotStartPools` /
  `hotStartPool` / `sandboxConfigs`.
- **Everything is team-scoped** (team slug). `organization_*` fields/queries are
  intentionally excluded — do not add them.
- Commands: `create`, `list`, `get`, `sandboxes`.

## Layout

- `instruqt/` — domain client, no CLI knowledge. `client.go` (net/http GraphQL
  transport, functional options), `operations.go` (mutation/queries + types +
  cursor pagination + `sandboxConfigs`), `validate.go` (pure `Validate(now)`),
  `profiles.go` (event-type profile table).
- `cmd/` — Cobra/Viper CLI. `root.go` (flags, viper wiring, client construction),
  `create.go`, `list.go`, `get.go`, `sandboxes.go`, `render.go`.
- `main.go` — wiring only. `docs/superpowers/specs/2026-07-09-instruqt-hotstart-design.md`
  is the design spec. `README.md` = reference, `HOWTO.md` = new-user guide.

## Build / test

```sh
make build          # go build -o instruqt-hotstart .
make test           # go test ./...   (currently 45 tests, must stay green)
make vet            # go vet ./...
make fmt            # gofmt -w
```

TDD is the norm here: write the failing test before the implementation. Tests
use `httptest` + table-driven cases; no mocking framework.

## Non-obvious gotchas (do not relearn these the hard way)

1. **Naming vs wire contract (intentional mismatch).** The CLI/Go say
   "sandbox": flag `--sandbox-ids`, `sandboxes` command, `Sandbox` type,
   `Client.Sandboxes`, `HotStartPoolInput.SandboxIDs`. But the **wire is
   unchanged**: the JSON tag is still `json:"configs"` and the GraphQL operation
   is still `sandboxConfigs`. Do **not** "fix" this to match — it is deliberate
   (users found "configs" confusing; the API name stays correct).

2. **Pools are built from sandbox IDs, not tracks.** `createHotStartPool`
   rejects `tracks` at runtime (`tracks: must be blank; config_versions: cannot
   be blank`). Use `--sandbox-ids` (values like `0bgr0ddoarzk`); find them with
   `instruqt-hotstart sandboxes`. `--tracks` still exists but the CLI warns.

3. **Viper env keys need the hyphen→underscore replacer.** `AutomaticEnv` maps
   `api-key` to `INSTRUQT_API-KEY`, not `INSTRUQT_API_KEY`. `configureViper`
   (cmd/root.go) sets `SetEnvKeyReplacer(strings.NewReplacer("-", "_"))`. Any new
   hyphenated flag relies on this.

4. **`.env` must use `export`.** Sourcing plain `KEY=value` lines only sets shell
   variables the CLI (a child process) can't see, producing a misleading
   "no API key". `env.example` uses `export`; the error message says so.

5. **Test isolation:** `TestConfigPrecedence` clears ambient `INSTRUQT_*` via
   `t.Setenv(...,"")` (viper treats empty env as unset) because a sourced `.env`
   in the dev shell otherwise leaks in.

## Validation & profiles (create)

- Required: `--type` (dedicated|shared) and `--name`. `--name` is a required
  cobra flag (checked before `Validate`, not bypassable by `--force`) AND a
  `Validate()` hard error (for library callers).
- Hard errors (block unless `--force`): missing type/name, `size<=0`,
  `ends_at` before `starts_at`.
- Warnings (never block): no `ends_at`; `starts_at` under the size-tiered lead
  time (`<50`→20m, `50-100`→30m, `100<size<200`→1h, `200-400`→90m, `>400`→90m).
- Profiles fill only unset fields and never set `type` or `name`. Table in
  `instruqt/profiles.go`.

## Status

Implemented, 45 tests green, vet/fmt clean. Live create→list→get against real
Instruqt needs a real API key + team (not run in-repo). Use `--dry-run` to
preview payloads without spending sandboxes.
