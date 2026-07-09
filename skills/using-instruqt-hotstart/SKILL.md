---
name: using-instruqt-hotstart
description: Use when creating, listing, or inspecting Instruqt hot start pools (pools of pre-warmed sandboxes) for a team with the instruqt-hotstart CLI — including "warm a pool before a workshop/webinar", finding sandbox IDs, previewing a pool, or diagnosing a "no API key" error.
---

# Using instruqt-hotstart

## Overview

`instruqt-hotstart` is a team-scoped CLI over the Instruqt GraphQL API for
managing **hot start pools** — pools of sandboxes warmed up before tracks start.
Four commands: `create`, `list`, `get`, `sandboxes`. Everything is scoped to a
**team slug**.

Full flag/validation/profile reference:
[references/instruqt-hotstart-cli.md](references/instruqt-hotstart-cli.md). Every
command also supports `--help`.

## When to use

- Creating a warm pool ahead of a workshop, webinar, conference session, or demo.
- Listing or inspecting existing pools for a team.
- Finding the sandbox IDs a pool must be built from.
- Diagnosing setup errors ("no API key", "no team").

Not for: creating Instruqt *tracks* or content, org-level operations (this tool
is team-scoped only).

## Setup (do this first)

The CLI needs an **API key** and, for most commands, a **team slug**. Resolution
precedence: flag > env > `./config.yaml` > default.

```sh
# any of these work; flag wins
export INSTRUQT_API_KEY=team-xxxxxxxx   # note: MUST be `export`ed (see below)
export INSTRUQT_TEAM=acme
instruqt-hotstart list --team acme --api-key team-xxxxxxxx
```

**`.env` gotcha:** a plain `KEY=value` line only sets a shell variable the CLI (a
child process) can't see, producing a misleading `no API key`. Each line must use
`export`, or source with `set -a`:

```sh
set -a; source .env; set +a         # or put `export ` on each line of .env
```

## Workflow: Find → Preview → Operate

Pools are built from **sandbox IDs, not tracks**. Always:

1. **Find** the sandbox IDs for your content:
   ```sh
   instruqt-hotstart sandboxes --team acme        # lists ID / SLUG / NAME / VERSION
   ```
   Match the SLUG/NAME to your track; copy the ID (e.g. `0bgr0ddoarzk`).
2. **Preview** with `--dry-run` — resolves and prints the payload + warnings,
   sends nothing, spends no sandboxes:
   ```sh
   instruqt-hotstart create --team acme --type shared --name demo --size 50 \
     --sandbox-ids 0bgr0ddoarzk --starts-at +2h --dry-run
   ```
3. **Operate** — drop `--dry-run` to create, then verify:
   ```sh
   instruqt-hotstart list --team acme
   instruqt-hotstart get --id <pool-id>            # get needs --id, not --team
   ```

Add `--json` to any command for machine-readable output (pipe `list --json` to
`jq` to grab a pool ID).

## Quick reference

| Intent | Command |
|--------|---------|
| Find sandbox IDs for a team | `sandboxes --team <slug>` |
| Preview a pool (no spend) | `create ... --dry-run` |
| Create a pool | `create --team <slug> --type <dedicated\|shared> --name <n> --size <n> --sandbox-ids <id,...>` |
| List a team's pools | `list --team <slug>` |
| Inspect one pool | `get --id <pool-id>` |

`--type` and `--name` are **always required** for create and are never set by a
profile. Use `--profile <name> --registrations <N>` to auto-size for an event
type (e.g. `webinar`); see the reference for the profile table.

## Common mistakes

| Mistake | Fix |
|---------|-----|
| Using `--tracks` to build a pool | Pools use `--sandbox-ids`; the API rejects tracks. Run `sandboxes` to get IDs. |
| Inventing/guessing sandbox IDs | Discover them with `sandboxes --team <slug>` — never fabricate. |
| Creating without `--dry-run` first | Always preview; real creation warms (and bills for) sandboxes. |
| `.env` without `export` → "no API key" | `export` each line, or `set -a; source .env; set +a`. |
| Passing `--team` to `get` | `get` is keyed by `--id` only; team is ignored there. Forgetting `--team` on `create`/`list`/`sandboxes` errors. |
| Expecting `--profile` to set type/name | Profiles fill only unset fields; you must still pass `--type` and `--name`. |
