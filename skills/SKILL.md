---
name: weightlogr-cli
description: Use this skill when the user wants to log body weight, check weigh-in history, or analyze weight trends. Triggers on "log my weight", "weigh-in", "weight check", "how much do I weigh", or any mention of tracking body weight over time.
---

# weightlogr-cli

Weight tracking CLI. All output is machine-parseable.

## Binary location

```
/Users/julientant/projects/weightlogr-cli/dist/weightlogr
```

## Global flags

Every command accepts these. Also settable via `WEIGHTLOGR_*` env vars or `.weightlogr.yaml`.

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--db` | `WEIGHTLOGR_DB` | `/opt/data/weights.db` | SQLite database path |
| `--timezone` | `WEIGHTLOGR_TIMEZONE` | `America/Phoenix` | Timezone for all timestamps |
| `--format` | `WEIGHTLOGR_FORMAT` | `json` | Output: `json`, `csv` |
| `--log-file` | `WEIGHTLOGR_LOG_FILE` | `/opt/data/weightlogr.log` | Log file path (`stderr` for stderr) |
| `--log-level` | `WEIGHTLOGR_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

Default format is `json`. Use `--format csv` for CSV output.

## Commands

### insert — Log a weigh-in

```bash
weightlogr insert <weight> [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--timestamp` | now | ISO 8601 timestamp (e.g. `2026-04-05T08:00`, `2026-04-05T08:00:00-07:00`) |
| `--source` | `daily-check` | Source label for categorization |
| `--notes` | | Free-text notes (commas and special chars are safe) |

**Examples:**

```bash
# Quick log (now, Phoenix time)
weightlogr insert 185.2 --format json

# With timestamp and notes
weightlogr insert 184.0 --timestamp 2026-04-03T08:00 --notes "after gym" --source gym-check --format json

# Output (json):
# {"created_at":"2026-04-05 11:06:55","id":1,"notes":"","source":"daily-check","weight":185.2}
```

### list — Query weigh-ins

```bash
weightlogr list [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--since` | | Start date/time inclusive (ISO 8601) |
| `--until` | | End date/time exclusive (ISO 8601) |
| `--source` | | Filter by source label |
| `--order` | `desc` | Sort: `asc` or `desc` |
| `--limit` | `0` | Max rows (0 = unlimited) |

**Examples:**

```bash
# Last 5 entries as JSON
weightlogr list --limit 5 --format json

# Date range
weightlogr list --since 2026-03-31 --until 2026-04-07 --format json

# Filter by source, oldest first
weightlogr list --source gym-check --order asc --format json

# Output (json):
# [{"id":1,"weight":185.2,"created_at":"2026-04-05 11:06:55","source":"daily-check","notes":"test"}]
```

## Database schema

```sql
CREATE TABLE weigh_ins (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    weight     REAL    NOT NULL,            -- pounds
    created_at TEXT    NOT NULL UNIQUE,      -- "YYYY-MM-DD HH:MM:SS" in configured timezone
    source     TEXT,                         -- e.g. "daily-check", "gym-check"
    notes      TEXT
);
```

### version — Print build info

```bash
weightlogr version
# {"version":"dev","commit":"none","date":"unknown"}

weightlogr version --format csv
# version,commit,date
# dev,none,unknown
```

## AI integration notes

- Default output format is `json` — all commands produce structured output by default
- Timestamps with timezone offsets are auto-converted to the configured timezone
- `created_at` is UNIQUE — two entries at the same second will conflict
- The `--source` flag is useful for distinguishing manual vs automated entries
- Empty `notes` returns `""` in JSON (not null)
- The database auto-migrates on first use — no setup needed
- Logs go to file by default, keeping stdout clean for output parsing
