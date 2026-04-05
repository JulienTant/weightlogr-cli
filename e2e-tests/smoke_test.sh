#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
EXPECTED_DIR="$SCRIPT_DIR/expected"
BIN="go run $PROJECT_DIR"
DB=$(mktemp -t weightlogr-smoke-XXXXXX.db)
LOG=$(mktemp -t weightlogr-smoke-XXXXXX.log)
PASS=0
FAIL=0

cleanup() {
  rm -f "$DB" "$LOG"
}
trap cleanup EXIT

FLAGS="--db $DB --log-file $LOG --log-level debug"

pass() {
  PASS=$((PASS + 1))
  echo "  ✓ $1"
}

fail() {
  FAIL=$((FAIL + 1))
  echo "  ✗ $1"
  echo "    $2"
}

# Compare command output against an expected file.
# Usage: assert_output "label" "expected_file" command args...
assert_output() {
  local label="$1" expected_file="$EXPECTED_DIR/$2"
  shift 2

  local actual
  actual=$("$@" 2>/dev/null) || true

  local expected
  expected=$(cat "$expected_file")

  if [ "$actual" = "$expected" ]; then
    pass "$label"
  else
    fail "$label" "output differs from $2"
    diff --color=auto <(echo "$expected") <(echo "$actual") || true
  fi
}

assert_exit_code() {
  local code="$1" expected="$2" label="$3"
  if [ "$code" -eq "$expected" ]; then
    pass "$label"
  else
    fail "$label" "expected exit code $expected, got $code"
  fi
}

# --- Seed data ---
echo "Seeding data..."
$BIN insert 185.2 $FLAGS --timestamp "2026-04-05T08:00:00" --notes "morning" >/dev/null 2>&1
$BIN insert 183.8 $FLAGS --timestamp "2026-04-04T07:30:00" --source "gym-check" --notes "after gym" >/dev/null 2>&1
$BIN insert 184.0 $FLAGS --timestamp "2026-04-06T09:00:00" --notes "before lunch, felt light" >/dev/null 2>&1
echo ""

# --- Snapshot tests ---
echo "List output:"
assert_output "list table" "list_table.txt" $BIN list $FLAGS
assert_output "list json" "list_json.txt" $BIN list $FLAGS --format json
assert_output "list csv" "list_csv.txt" $BIN list $FLAGS --format csv

echo ""
echo "Filter output:"
assert_output "filter --since" "filter_since.txt" $BIN list $FLAGS --format json --since "2026-04-05T00:00:00"
assert_output "filter --until" "filter_until.txt" $BIN list $FLAGS --format json --until "2026-04-05T00:00:00"
assert_output "filter --source" "filter_source.txt" $BIN list $FLAGS --format json --source "gym-check"

echo ""
echo "Order and limit:"
assert_output "order asc" "order_asc.txt" $BIN list $FLAGS --format json --order asc
assert_output "limit 1" "limit_1.txt" $BIN list $FLAGS --format json --limit 1

echo ""
echo "Timezone conversion:"
assert_output "UTC to Phoenix" "insert_utc.txt" $BIN insert 186.0 --db "$DB" --log-file "$LOG" --timestamp "2026-04-07T15:00:00Z"

echo ""
echo "Empty db:"
EMPTY_DB=$(mktemp -t weightlogr-empty-XXXXXX.db)
assert_output "empty table" "list_empty_table.txt" $BIN list --db "$EMPTY_DB" --log-file "$LOG"
assert_output "empty json" "list_empty_json.txt" $BIN list --db "$EMPTY_DB" --log-file "$LOG" --format json
rm -f "$EMPTY_DB"

echo ""
echo "Error handling:"

rc=0
$BIN insert not-a-number $FLAGS >/dev/null 2>&1 || rc=$?
assert_exit_code "$rc" "1" "invalid weight returns exit code 1"

rc=0
$BIN insert 180.0 $FLAGS --timestamp "2026-04-05T08:00:00" >/dev/null 2>&1 || rc=$?
assert_exit_code "$rc" "1" "duplicate timestamp returns exit code 1"

rc=0
$BIN insert $FLAGS >/dev/null 2>&1 || rc=$?
assert_exit_code "$rc" "1" "missing weight arg returns exit code 1"

echo ""
echo "================================"
echo "Results: $PASS passed, $FAIL failed"
echo "================================"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
