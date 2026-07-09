# HOWTO: Using the Instruqt Hot Start CLI

This guide walks a brand-new user from zero to creating and checking Instruqt
hot start pools. If you just want the reference details, see `README.md`.

## What this tool does

Instruqt **hot start pools** keep a set of sandboxes warmed up and ready before
your tracks begin, so participants get an almost instant start instead of
waiting for a sandbox to provision.

`instruqt-hotstart` is a command-line tool that talks to the Instruqt GraphQL
API for you. With it you can:

- **create** a new hot start pool,
- **list** the pools for a team,
- **get** the details of one pool by its ID.

Everything is scoped to a single **team**.

## What you need before you start

1. **Go 1.22 or newer**, to build the tool. Check with `go version`.
2. **An Instruqt API key.** In the Instruqt web app, go to
   **Settings → API keys → Generate API Key** and copy the key.
3. **Your team slug** — the short name of your team in Instruqt.

## Step 1: Build the tool

From the project directory:

```sh
go build -o instruqt-hotstart .
```

This produces an `instruqt-hotstart` binary in the current directory. Run it
with `./instruqt-hotstart`. You can move it onto your `PATH` if you want to call
it from anywhere.

Check it works:

```sh
./instruqt-hotstart --help
```

## Step 2: Give it your credentials

The tool needs your API key and team. The cleanest way is environment
variables, so your secret never ends up in a command history or a file:

```sh
export INSTRUQT_API_KEY="paste-your-key-here"
export INSTRUQT_TEAM="your-team-slug"
```

You can also copy `env.example` to `.env`, fill it in, and source it:

```sh
cp env.example .env
# edit .env
source .env
```

**Important:** the lines in `.env` must start with `export` (as `env.example`
does). Sourcing a file with plain `KEY=value` lines only sets shell variables,
which the CLI — a separate child process — cannot see, and you will get
"no API key" even though the value looks set. If your file has no `export`
keywords, either add them or source it like this:

```sh
set -a; source .env; set +a
```

If you prefer, non-secret defaults like the team can live in a `config.yaml`
file (copy `config.example.yaml`). The tool reads settings in this order, so
anything later overrides anything earlier:

**command-line flag  →  environment variable  →  config file  →  built-in default**

That means a flag always wins, which is handy for one-off overrides.

## Step 3: Find your sandbox IDs

A hot start pool is built from **sandboxes**, not tracks. The API rejects track
IDs for pool creation, so you first need the sandbox ID(s) for the content you
want to warm up. List them for your team:

```sh
./instruqt-hotstart sandboxes
```

This prints a table of `ID`, `SLUG`, `NAME`, and `VERSION`. Note the `ID` (e.g.
`0bgr0ddoarzk`) of the sandbox whose `SLUG` matches the content you want — you
pass it to `--sandbox-ids` in the next step. (`--json` works here too if you
want to script it.)

## Step 4: Preview a pool before creating it (dry run)

Always start with `--dry-run`. It shows you exactly what would be sent and warns
you about anything risky, but it does **not** create anything or cost you money.

```sh
./instruqt-hotstart create \
  --type shared \
  --name spring-workshop \
  --size 50 \
  --sandbox-ids 0bgr0ddoarzk \
  --auto-refill \
  --starts-at +2h \
  --ends-at +150m \
  --dry-run
```

A few things to know about the flags:

- `--type` is **required** and must be `dedicated` or `shared`.
- `--name` is **required** — give the pool a short, recognisable name.
- `--sandbox-ids` holds the sandbox ID(s) from Step 3 (comma-separated, or
  repeat the flag). This is what the pool is built from.
- `--tracks` exists but the API **rejects it** for pool creation; the tool warns
  if you use it. Use `--sandbox-ids` instead.
- `--auto-refill` (optional) keeps the pool topped up as sandboxes are claimed.
- `--starts-at` and `--ends-at` accept either a full timestamp
  (`2026-07-09T14:00:00Z`) or a relative offset from now (`+2h`, `+90m`,
  `+30m`).

The dry run prints the resolved request and any warnings to your screen.

## Step 5: Understand the warnings

The tool tries to stop you from wasting money. Two kinds of messages can appear:

- **Warnings** (start with `warning:`) do not stop the command. The most common
  ones are "no ends_at set" (an open-ended pool keeps billing until you delete
  it) and "under the recommended provisioning lead time" (you scheduled the
  start too soon for a pool that size to warm up in time).
- **Errors** stop the command. These are real mistakes: no `--type`, no
  `--name`, a size of zero or less, or an end time that comes before the start
  time. For most of these you can add `--force` to proceed anyway, but `--name`
  is genuinely required and cannot be forced past.

The recommended lead time grows with pool size:

| Pool size | Give it at least |
|---|---|
| under 50 | 20 minutes |
| 50–100 | 30 minutes |
| 101–199 | 1 hour |
| 200–400 | 1 hour 30 minutes |
| over 400 | 1 hour 30 minutes (and test it) |

## Step 6: Create the pool for real

When the dry run looks right, remove `--dry-run`:

```sh
./instruqt-hotstart create \
  --type shared \
  --name spring-workshop \
  --size 50 \
  --sandbox-ids 0bgr0ddoarzk \
  --auto-refill \
  --starts-at +2h \
  --ends-at +150m
```

The tool prints the created pool, including its **ID**. Keep that ID — you use
it to look the pool up later.

## Step 7: Check on your pools

List every pool for your team:

```sh
./instruqt-hotstart list
```

Look at one pool in detail:

```sh
./instruqt-hotstart get --id <pool-id>
```

By default both commands print a readable table. Add `--json` to any command
when you want machine-readable output for a script:

```sh
./instruqt-hotstart list --json
```

## Using a profile (optional shortcut)

If you are running a common kind of event, a **profile** fills in sensible
defaults for you (auto-refill behaviour, an end time, and a suggested size).
Anything you set explicitly still wins, and profiles never choose `--type` or
`--name` for you — you always pass those yourself.

Tell the profile how many people you expect with `--registrations`, and it
suggests a size:

```sh
./instruqt-hotstart create \
  --type shared \
  --name spring-webinar \
  --profile webinar \
  --registrations 500 \
  --starts-at +2h \
  --dry-run
```

For a webinar this produces a size of 125 (25% of 500), turns auto-refill on,
and sets the end time to 30 minutes after the start.

Available profiles and their defaults:

| Profile | Auto-refill | End time | Default size | Size from registrations |
|---|---|---|---|---|
| `self-paced` | on | none | 3 | fixed |
| `live-workshop` | off | +30m | 20 | 70% |
| `webinar` | on | +30m | 100 | 25% |
| `multi-day` | off | +30m | 20 | 100% |
| `conference-session` | off | +45m | 80 | 75% |
| `booth-demo` | on | none | 4 | fixed |
| `sales-demo` | on | none | 2 | fixed |

Profiles with "none" for the end time will remind you to set `--ends-at`
yourself so the pool does not run forever.

## Common problems

- **"no API key: set --api-key or INSTRUQT_API_KEY"** — the key is not in the
  CLI's environment. The most common cause is sourcing a `.env` whose lines lack
  `export`: `source .env` then sets shell variables the CLI cannot see. Add
  `export` to each line (see `env.example`), or run `set -a; source .env; set +a`.
  You can also pass `--api-key` directly.
- **"no team: set --team or INSTRUQT_TEAM"** — set `INSTRUQT_TEAM` or pass
  `--team` for `create` and `list`.
- **"type is required"** — add `--type dedicated` or `--type shared`. Profiles
  do not set this.
- **'required flag(s) "name" not set'** — add `--name <pool-name>`. Every pool
  needs a name; profiles do not set it.
- **"validation failed"** — read the message; fix the input, or add `--force`
  if you really mean it.
- **"tracks: must be blank; config_versions: cannot be blank"** — you passed
  `--tracks`, but pools are created from sandboxes. Run
  `./instruqt-hotstart sandboxes` to find the sandbox ID and pass it with
  `--sandbox-ids` instead.

## Where to learn more

- `README.md` — full flag and configuration reference.
- Instruqt hot start best practices:
  https://docs.instruqt.com/resources/hot-starts-best-practices
