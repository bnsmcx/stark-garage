#!/usr/bin/env bash
set -euo pipefail

# CLI integration test for toolbox-memory
# Usage: ./tests/cli-integration-test.sh

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BINARY="$SCRIPT_DIR/toolbox-memory"
DB="$(mktemp -d)/test.db"
PASS=0
FAIL=0

cleanup() {
  rm -rf "$(dirname "$DB")"
}
trap cleanup EXIT

check() {
  local desc="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    echo "  PASS  $desc"
    PASS=$((PASS + 1))
  else
    echo "  FAIL  $desc ($*)"
    FAIL=$((FAIL + 1))
  fi
}

check_output() {
  local desc="$1"
  local expected="$2"
  shift 2
  local output
  output=$("$@" 2>&1) || true
  if echo "$output" | grep -q "$expected"; then
    echo "  PASS  $desc"
    PASS=$((PASS + 1))
  else
    echo "  FAIL  $desc (expected '$expected' in output)"
    echo "        got: $output"
    FAIL=$((FAIL + 1))
  fi
}

# --- Build binary if needed ---
if [ ! -f "$BINARY" ]; then
  echo "Building toolbox-memory..."
  (cd "$SCRIPT_DIR" && go build -o toolbox-memory ./cmd/toolbox-memory/)
fi

echo "=== CLI Integration Test ==="
echo "Binary: $BINARY"
echo "DB:     $DB"
echo ""

# --- version ---
check_output "version" "0.1.0" "$BINARY" version

# --- init ---
check_output "init creates db" "initialized" "$BINARY" init --db "$DB"
check "db file exists after init" test -f "$DB"

# --- write ---
check_output "write returns ok" '"status": "ok"' \
  "$BINARY" write --db "$DB" --ns calibration --agent planner --key "test-pattern-1" --value '{"wrong":"bad approach","right":"good approach","why":"because"}'

check_output "write second entry" '"status": "ok"' \
  "$BINARY" write --db "$DB" --ns bug_pattern --agent debugger --key "nil-pointer-reset" --value '{"class":"state-corruption","prevention":"reconstruct after destroy"}'

check_output "write third entry" '"status": "ok"' \
  "$BINARY" write --db "$DB" --ns calibration --agent planner --key "test-pattern-2" --value '{"wrong":"wrong way","right":"right way","why":"reasons"}'

# --- read ---
check_output "read returns entry" '"key": "test-pattern-1"' \
  "$BINARY" read --db "$DB" --ns calibration --key "test-pattern-1"

check_output "read returns value" "bad approach" \
  "$BINARY" read --db "$DB" --ns calibration --key "test-pattern-1"

# --- read increments hit_count ---
# Read twice more, then check hitCount
"$BINARY" read --db "$DB" --ns calibration --key "test-pattern-1" >/dev/null 2>&1
OUTPUT=$("$BINARY" read --db "$DB" --ns calibration --key "test-pattern-1" 2>&1)
if echo "$OUTPUT" | grep -q '"hitCount": [3-9]'; then
  echo "  PASS  read increments hitCount (3+ after 3 reads)"
  PASS=$((PASS + 1))
else
  echo "  FAIL  read increments hitCount"
  echo "        got: $(echo "$OUTPUT" | grep hitCount)"
  FAIL=$((FAIL + 1))
fi

# --- FTS hyphen-safe search (#6) ---
"$BINARY" write --db "$DB" --ns bug_pattern --agent debugger --key "hyphen-test" --value "this value has alt-screen and logger tokens" >/dev/null
check_output "hyphen-safe search" "hyphen-test" \
  "$BINARY" search --db "$DB" --ns bug_pattern --query "alt-screen logger"

check_output "--raw passes through" "hyphen-test" \
  "$BINARY" search --db "$DB" --ns bug_pattern --query "logger" --raw

# Clean up so it doesn't skew downstream counts.
"$BINARY" delete --db "$DB" --ns bug_pattern --key "hyphen-test" >/dev/null

# --- peek (side-effect-free) ---
check_output "peek returns entry" '"key": "test-pattern-1"' \
  "$BINARY" peek --db "$DB" --ns calibration --key "test-pattern-1"

# Seed a fresh entry and verify peek doesn't mutate hit_count.
"$BINARY" write --db "$DB" --ns calibration --agent planner --key "peek-target" --value '{"a":"b"}' >/dev/null
"$BINARY" peek --db "$DB" --ns calibration --key "peek-target" >/dev/null
"$BINARY" peek --db "$DB" --ns calibration --key "peek-target" >/dev/null
"$BINARY" peek --db "$DB" --ns calibration --key "peek-target" >/dev/null
# Now a single read should bump hit_count to exactly 1, then another peek reads it back unchanged.
"$BINARY" read --db "$DB" --ns calibration --key "peek-target" >/dev/null
OUTPUT=$("$BINARY" peek --db "$DB" --ns calibration --key "peek-target" 2>&1)
if echo "$OUTPUT" | grep -q '"hitCount": 1'; then
  echo "  PASS  peek is side-effect-free (hitCount=1 after 3 peeks + 1 read)"
  PASS=$((PASS + 1))
else
  echo "  FAIL  peek is side-effect-free"
  echo "        got: $(echo "$OUTPUT" | grep hitCount)"
  FAIL=$((FAIL + 1))
fi

# peek on missing key exits non-zero
"$BINARY" peek --db "$DB" --ns calibration --key "does-not-exist" >/dev/null 2>&1 && {
  echo "  FAIL  peek on missing key should exit non-zero"
  FAIL=$((FAIL + 1))
} || {
  echo "  PASS  peek on missing key exits non-zero"
  PASS=$((PASS + 1))
}

# Clean up the peek-target so it doesn't disturb the list count later.
"$BINARY" delete --db "$DB" --ns calibration --key "peek-target" >/dev/null

# --- search ---
check_output "search finds entry" "test-pattern-1" \
  "$BINARY" search --db "$DB" --ns calibration --query "bad approach"

check_output "search respects namespace" "nil-pointer" \
  "$BINARY" search --db "$DB" --ns bug_pattern --query "state corruption"

# Search in wrong namespace should not find it
OUTPUT=$("$BINARY" search --db "$DB" --ns calibration --query "state corruption destroy" 2>&1)
if echo "$OUTPUT" | grep -q "nil-pointer-reset"; then
  echo "  FAIL  search namespace isolation (found bug_pattern entry in calibration ns)"
  FAIL=$((FAIL + 1))
else
  echo "  PASS  search namespace isolation"
  PASS=$((PASS + 1))
fi

# --- list ---
OUTPUT=$("$BINARY" list --db "$DB" --ns calibration 2>&1)
COUNT=$(echo "$OUTPUT" | grep -c '"key"' || true)
if [ "$COUNT" -eq 2 ]; then
  echo "  PASS  list returns 2 calibration entries"
  PASS=$((PASS + 1))
else
  echo "  FAIL  list returns $COUNT entries, want 2"
  FAIL=$((FAIL + 1))
fi

# --- stats ---
check_output "stats shows active" '"active": 3' \
  "$BINARY" stats --db "$DB"

check_output "stats shows total" '"total": 3' \
  "$BINARY" stats --db "$DB"

# --- promote ---
check_output "promote returns ok" '"status": "promoted"' \
  "$BINARY" promote --db "$DB" --ns calibration --key "test-pattern-1" --to "CLAUDE.md Development Philosophy"

# Verify promoted entry shows in stats
check_output "stats shows promoted" '"promoted": 1' \
  "$BINARY" stats --db "$DB"

# Verify promoted entry excluded from list
OUTPUT=$("$BINARY" list --db "$DB" --ns calibration 2>&1)
if echo "$OUTPUT" | grep -q "test-pattern-1"; then
  echo "  FAIL  list excludes promoted entries"
  FAIL=$((FAIL + 1))
else
  echo "  PASS  list excludes promoted entries"
  PASS=$((PASS + 1))
fi

# --- prune (with nothing to prune) ---
check_output "prune with nothing to do" '"transitions": 0' \
  "$BINARY" prune --db "$DB"

# --- write + upsert ---
check_output "upsert updates value" '"status": "ok"' \
  "$BINARY" write --db "$DB" --ns calibration --agent planner --key "test-pattern-2" --value '{"wrong":"updated wrong","right":"updated right","why":"updated"}'

check_output "upsert preserved key" "updated wrong" \
  "$BINARY" read --db "$DB" --ns calibration --key "test-pattern-2"

# --- delete ---
check_output "delete returns ok" '"status": "deleted"' \
  "$BINARY" delete --db "$DB" --ns calibration --key "test-pattern-2"

# Verify deleted
OUTPUT=$("$BINARY" read --db "$DB" --ns calibration --key "test-pattern-2" 2>&1) && {
  echo "  FAIL  delete removes entry (read succeeded)"
  FAIL=$((FAIL + 1))
} || {
  echo "  PASS  delete removes entry (read failed as expected)"
  PASS=$((PASS + 1))
}

# --- error cases ---
OUTPUT=$("$BINARY" write --db "$DB" 2>&1) && {
  echo "  FAIL  write without required flags should fail"
  FAIL=$((FAIL + 1))
} || {
  echo "  PASS  write without required flags exits non-zero"
  PASS=$((PASS + 1))
}

OUTPUT=$("$BINARY" search --db "$DB" 2>&1) && {
  echo "  FAIL  search without required flags should fail"
  FAIL=$((FAIL + 1))
} || {
  echo "  PASS  search without required flags exits non-zero"
  PASS=$((PASS + 1))
}

# --- write --value-file ---
VAL_FILE="$(dirname "$DB")/val.json"
printf '{"rule":"from-file"}' > "$VAL_FILE"
check_output "write --value-file" '"status": "ok"' \
  "$BINARY" write --db "$DB" --ns bug_pattern --agent t --key vf --value-file "$VAL_FILE"

check_output "value-file round-trip" 'from-file' \
  "$BINARY" read --db "$DB" --ns bug_pattern --key vf

# --- write --value - (stdin) ---
OUTPUT=$(printf 'stdin-value' | "$BINARY" write --db "$DB" --ns bug_pattern --agent t --key sv --value - 2>&1)
if echo "$OUTPUT" | grep -q '"status": "ok"'; then
  echo "  PASS  write --value - (stdin)"
  PASS=$((PASS + 1))
else
  echo "  FAIL  write --value - (stdin)"
  echo "        got: $OUTPUT"
  FAIL=$((FAIL + 1))
fi

check_output "stdin round-trip" 'stdin-value' \
  "$BINARY" read --db "$DB" --ns bug_pattern --key sv

# --- mutual exclusion ---
"$BINARY" write --db "$DB" --ns bug_pattern --agent t --key x --value a --value-file "$VAL_FILE" >/dev/null 2>&1 && {
  echo "  FAIL  mutual exclusion --value + --value-file should exit non-zero"
  FAIL=$((FAIL + 1))
} || {
  echo "  PASS  mutual exclusion --value + --value-file exits non-zero"
  PASS=$((PASS + 1))
}

# --- namespaces ---
check_output "namespaces lists bug_pattern" '"namespace": "bug_pattern"' \
  "$BINARY" namespaces --db "$DB"

check_output "namespaces lists calibration" '"namespace": "calibration"' \
  "$BINARY" namespaces --db "$DB"

# --- stats --by-ns ---
check_output "stats --by-ns has byNamespace" '"byNamespace"' \
  "$BINARY" stats --db "$DB" --by-ns

check_output "stats --by-ns has lifecycle" '"lifecycle"' \
  "$BINARY" stats --db "$DB" --by-ns

# --- stats default shape unchanged (no byNamespace key) ---
OUTPUT=$("$BINARY" stats --db "$DB" 2>&1)
if echo "$OUTPUT" | grep -q '"byNamespace"'; then
  echo "  FAIL  default stats should not include byNamespace"
  FAIL=$((FAIL + 1))
else
  echo "  PASS  default stats shape unchanged (no byNamespace)"
  PASS=$((PASS + 1))
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
