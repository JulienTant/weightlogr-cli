---
name: weightlogr-cli
description: Use this skill when the user wants to log body weight, check weigh-in history, or analyze weight trends. Triggers on "log my weight", "weigh-in", "weight check", "how much do I weigh", or any mention of tracking body weight over time.
---

# weightlogr-cli

Weight tracking CLI. All output is machine-parseable.

## Installation

```bash
# Download and install latest release (linux amd64)
curl -sL "https://api.github.com/repos/JulienTant/weightlogr-cli/releases/latest" \
  | grep -o '"browser_download_url": *"[^"]*linux_amd64[^"]*"' \
  | sed 's/"browser_download_url": *"//;s/"$//' \
  | xargs curl -sL | tar -xz -C /usr/local/bin weightlogr-cli
```

## Binary location

```
/usr/local/bin/weightlogr-cli
```

## Global flags

Every command accepts these. Also settable via `WEIGHTLOGR_*` env vars or `.weightlogr.yaml`.

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--db` | `WEIGHTLOGR_DB` | `/opt/data/weights.db` | SQLite database path |
| `--format` | `WEIGHTLOGR_FORMAT` | `json` | Output: `json`, `csv` |
| `--log-file` | `WEIGHTLOGR_LOG_FILE` | `/opt/data/weightlogr.log` | Log file path (`stderr` for stderr) |
| `--log-level` | `WEIGHTLOGR_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

Default format is `json`. Use `--format csv` for CSV output.

## Commands

### insert — Log a weigh-in

```bash
weightlogr-cli insert <weight> [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--timestamp` | now | RFC3339 timestamp (e.g. `2026-04-05T08:00:00-07:00`, `2026-04-05T15:00:00Z`) |
| `--source` | `daily-check` | Source label for categorization |
| `--notes` | | Free-text notes (commas and special chars are safe) |

**Examples:**

```bash
# Quick log (now, UTC)
weightlogr-cli insert 185.2 --format json

# With timestamp and notes
weightlogr-cli insert 184.0 --timestamp 2026-04-03T08:00:00-07:00 --notes "after gym" --source gym-check --format json

# Output (json):
# {"created_at":"2026-04-05T15:06:55Z","id":1,"notes":"","source":"daily-check","weight":185.2}
```

### list — Query weigh-ins

```bash
weightlogr-cli list [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--since` | | Start date/time inclusive (ISO 8601) |
| `--until` | | End date/time exclusive (ISO 8601) |
| `--source` | | Filter by source label |
| `--order` | `desc` | Sort: `asc` or `desc` |
| `--limit` | `0` | Max rows (0 = unlimited) |
| `--timezone` | | Convert output timestamps to this timezone (e.g. America/Phoenix). Defaults to UTC |

**Examples:**

```bash
# Last 5 entries as JSON
weightlogr-cli list --limit 5 --format json

# Date range
weightlogr-cli list --since 2026-03-31 --until 2026-04-07 --format json

# Filter by source, oldest first
weightlogr-cli list --source gym-check --order asc --format json

# Output (json):
# [{"id":1,"weight":185.2,"created_at":"2026-04-05T15:06:55Z","source":"daily-check","notes":"test"}]
```

## Database schema

```sql
CREATE TABLE weigh_ins (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    weight     REAL    NOT NULL,            -- pounds
    created_at TEXT    NOT NULL UNIQUE,      -- RFC3339 UTC (e.g. "2026-04-05T15:00:00Z")
    source     TEXT,                         -- e.g. "daily-check", "gym-check"
    notes      TEXT
);
```

### version — Print build info

```bash
weightlogr-cli version
# {"version":"dev","commit":"none","date":"unknown"}
```

## AI integration notes

- Timestamps are always RFC3339 — all commands produce and consume RFC3339 formatted timestamps
- Default output format is `json` — all commands produce structured output by default
- `created_at` is UNIQUE — two entries at the same second will conflict
- The `--source` flag is useful for distinguishing manual vs automated entries
- Empty `notes` returns `""` in JSON (not null)
- The database auto-migrates on first use — no setup needed
- Logs go to file by default, keeping stdout clean for output parsing
