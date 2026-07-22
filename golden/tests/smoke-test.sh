#!/usr/bin/env bash
set -euo pipefail

# Smoke test for deploy.sh — verifies the golden set deploys correctly
# Usage: ./tests/smoke-test.sh

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
TARGET="$(mktemp -d)"
PASS=0
FAIL=0

cleanup() {
  rm -rf "$TARGET"
}
trap cleanup EXIT

check() {
  local desc="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    echo "  PASS  $desc"
    PASS=$((PASS + 1))
  else
    echo "  FAIL  $desc"
    FAIL=$((FAIL + 1))
  fi
}

check_file() {
  local desc="$1"
  local path="$2"
  check "$desc" test -f "$path"
}

check_dir() {
  local desc="$1"
  local path="$2"
  check "$desc" test -d "$path"
}

echo "=== Deploy Smoke Test ==="
echo "Golden set: $SCRIPT_DIR"
echo "Target:     $TARGET"
echo ""

# --- Run deploy.sh ---
echo "--- Deploying ---"
"$SCRIPT_DIR/deploy.sh" "$TARGET" 2>&1 | sed 's/^/  /'
echo ""

# --- Verify structure ---
echo "--- Verifying Structure ---"

check_file "CLAUDE.md created" "$TARGET/CLAUDE.md"
check_file "BUDGETS.md created" "$TARGET/BUDGETS.md"
check_file ".mcp.json created" "$TARGET/.mcp.json"
check_dir  ".claude/commands/ exists" "$TARGET/.claude/commands"
check_dir  ".claude/agents/ exists" "$TARGET/.claude/agents"
check_dir  ".claude/agents/extensions/ exists" "$TARGET/.claude/agents/extensions"
check_dir  ".claude/skills/browser-automation/ exists" "$TARGET/.claude/skills/browser-automation"
check_dir  "agent_docs/ exists" "$TARGET/agent_docs"

# --- Verify command count matches the golden source (no magic number; catches drift + deletions) ---
EXPECTED_CMDS=$(ls "$SCRIPT_DIR/.claude/commands/"*.md 2>/dev/null | wc -l)
CMD_COUNT=$(ls "$TARGET/.claude/commands/"*.md 2>/dev/null | wc -l)
check "all $EXPECTED_CMDS commands deployed" test "$CMD_COUNT" -eq "$EXPECTED_CMDS"

# --- Verify agent count matches the golden source ---
EXPECTED_AGENTS=$(ls "$SCRIPT_DIR/.claude/agents/"*.md 2>/dev/null | wc -l)
AGENT_COUNT=$(ls "$TARGET/.claude/agents/"*.md 2>/dev/null | wc -l)
check "all $EXPECTED_AGENTS agents deployed" test "$AGENT_COUNT" -eq "$EXPECTED_AGENTS"

# --- Verify specific files ---
check_file "wiggum.md exists" "$TARGET/.claude/commands/wiggum.md"
check_file "review-pr.md exists" "$TARGET/.claude/commands/review-pr.md"
check_file "setup-release.md exists" "$TARGET/.claude/commands/setup-release.md"
check_file "indexer.md exists" "$TARGET/.claude/agents/indexer.md"
check_file "planner.md exists" "$TARGET/.claude/agents/planner.md"
check_file "builder.md exists" "$TARGET/.claude/agents/builder.md"
check_file "browser-automation SKILL.md exists" "$TARGET/.claude/skills/browser-automation/SKILL.md"
check_file "settings.local.json exists" "$TARGET/.claude/settings.local.json"
check_file "issue-conventions.md exists" "$TARGET/agent_docs/issue-conventions.md"
check_file "self-improvement.md exists" "$TARGET/agent_docs/self-improvement.md"

# --- Verify .mcp.json content ---
check ".mcp.json has playwright" grep -q "playwright" "$TARGET/.mcp.json"
check ".mcp.json has chrome-devtools" grep -q "chrome-devtools" "$TARGET/.mcp.json"

# --- Verify CLAUDE.md has bootstrap marker ---
check "CLAUDE.md has bootstrap marker" grep -q "bootstrap: project-specific below" "$TARGET/CLAUDE.md"

# --- Verify idempotency (re-deploy with "skip" preserves CLAUDE.md) ---
echo "test-marker" >> "$TARGET/CLAUDE.md"
echo "s" | "$SCRIPT_DIR/deploy.sh" "$TARGET" >/dev/null 2>&1
check "Re-deploy skip preserves CLAUDE.md" grep -q "test-marker" "$TARGET/CLAUDE.md"

# --- Verify overwrite option works ---
echo "o" | "$SCRIPT_DIR/deploy.sh" "$TARGET" >/dev/null 2>&1
if grep -q "test-marker" "$TARGET/CLAUDE.md" 2>/dev/null; then
  echo "  FAIL  Re-deploy overwrite replaces CLAUDE.md"
  FAIL=$((FAIL + 1))
else
  echo "  PASS  Re-deploy overwrite replaces CLAUDE.md"
  PASS=$((PASS + 1))
fi

# --- Verify templates were NOT deployed ---
check "build-and-test.md.template NOT deployed" test ! -f "$TARGET/agent_docs/build-and-test.md.template"
check "project-structure.md.template NOT deployed" test ! -f "$TARGET/agent_docs/project-structure.md.template"

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
