# mon — a Monday.com CLI

A small, fast CLI over the Monday.com API. Sign in once, register the boards you
care about under **names** (favorites), then list / fetch / create / edit items
with **server-side filtering** — including "just the items assigned to me".

Because every team's boards use different column IDs, each favorite stores its own
**column-role map** (auto-detected when you add it), so the same generic commands
work across your Tasks, Bugs, and Sprints boards without anything hardcoded.

## Build & install

```bash
make build      # -> ./bin/mon
make install    # -> /usr/bin/mon   (uses sudo)
# or:
go build -o bin/mon ./cmd/mon
```

## Getting started

```bash
# 1. Sign in (token from https://<team>.monday.com/apps/manage/tokens).
#    Reads $MONDAY_API_TOKEN if you don't pass one.
mon login <api-token>

# 2. Find the boards you want and register them under names.
mon board discover                     # list boards you can access
mon board add 2116466717 --name tasks --default
mon board add 2116466713 --name bugs
mon board add 2116466722 --name sprint
#   (run `mon board add` with no ID to pick interactively)

# 3. Use them.
mon items f -b tasks -u                # fetch MY tasks (server-side filtered)
mon items f -b bugs  -u                # fetch bugs I reported
mon items ls -b tasks                  # re-list from the local cache (offline)
```

Set a default board (`mon board default tasks`) and you can drop `-b` for the
common case: `mon items f -u`.

## Commands

### Auth
- `mon login [TOKEN]` — store the token (or `$MONDAY_API_TOKEN`) and record your identity
- `mon whoami [--refresh]` — show the signed-in user

### Boards (favorites)
- `mon board discover [--all] [--workspace ID]` — list accessible boards
- `mon board add [BOARD_ID] [--name N] [--default]` — register a board, auto-detecting its columns
- `mon board ls` — list your registered boards
- `mon board show <name>` — show a board's detected column-role map and labels
- `mon board set-col <name> <role> <columnId>` — override a detected column (roles: people, status, priority, type, sprint, active)
- `mon board default <name>` — set the default board
- `mon board rm <name>` — remove a favorite

### Items
- `mon items f [-b name] [-u] [--status S] [--priority P] [--type T] [--sprint X] [--limit N]` — fetch from the API (filtered server-side) and cache
- `mon items ls [-b name] [-u] [--status S] [--priority P] [--type T]` — list from the local cache (offline)
- `mon items show <localId> [-b name]` — show one cached item by its short id
- `mon items create <name…> [-b name] [--status/-s S] [--priority/-p P] [--type/-t T] [--assign me]`
- `mon items edit <localId> [-b name] [--status/-s S] [--priority/-p P] [--type/-t T]`

`-u/--me` filters to items assigned to you using Monday's `assigned_to_me` token on
the board's people column (server-side). `--status/--priority/--type` take the
label text as shown in Monday (e.g. `--status "In Progress"`); the CLI translates
it to the right index for the board. Boards that lack a role (e.g. the Bugs board
has no Type column) report a clear error rather than filtering silently.

### Sprints
- `mon sprint ls [-b name]` — list sprints (marks the active one)
- `mon sprint active [-b name]` — show the active sprint(s)

### Config
- `mon config show` — show the current config (token redacted)
- `mon config path` — print the config file path

## Configuration & data

- **Config**: `~/.config/monday-cli/config.json` (override with `$MONDAY_CONFIG`). Holds
  your token, identity, and named favorites with their column-role maps.
- **Cache**: `~/.cache/monday-cli/tasks.json` (override with `$MONDAY_CACHE`). The last
  fetched items per board, with the short LocalID index used by `show`/`edit`.
- **Env**: `MONDAY_API_TOKEN` overrides the stored token; `NO_COLOR` disables color.

A legacy config from the previous version (flat `api_key` / `board_id`) is migrated
automatically for the token and identity — re-register boards with `mon board add`.

## Development

Go + Cobra. `monday/` is the API client, domain models, config, cache, and column
detection; `cmd/mon/` is the Cobra command layer. `make test` / `make vet`.
