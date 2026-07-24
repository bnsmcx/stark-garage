#!/usr/bin/env bash
set -euo pipefail

# Derive .opencode/ artifacts from .claude/ (the single source of truth):
#   .opencode/commands/*.md  <- .claude/commands/*.md  (frontmatter: description [+ overrides])
#   .opencode/agents/*.md    <- .claude/agents/*.md     (frontmatter: description + mode: subagent)
#
# In both cases: drop `name` (implicit from filename) and `user_invocable`; the body (everything
# after the closing frontmatter `---`) is copied VERBATIM. Uses only bash/awk/sed — no yq/jq/python.
#
# Per-command OpenCode overrides (agent/model/subtask) come from commands/opencode.map.
# Agent `model` is intentionally NOT carried over: Claude Code aliases (e.g. `sonnet`) are not valid
# OpenCode `provider/model` identifiers, so ported agents use OpenCode's configured default.
#
# Usage:
#   gen-opencode.sh            (re)generate golden/.opencode/{commands,agents}
#   gen-opencode.sh --check    regenerate to a temp dir and diff vs the committed output; exit 1 on drift

GOLDEN_DIR="$(cd "$(dirname "$0")/.." && pwd)"
MAP="$GOLDEN_DIR/commands/opencode.map"

CHECK=0
[ "${1:-}" = "--check" ] && CHECK=1
if [ "$CHECK" -eq 1 ]; then
  TMP="$(mktemp -d)"
  trap 'rm -rf "$TMP"' EXIT
fi

# override <command> <key> -> prints the override value, or nothing
override() {
  [ -f "$MAP" ] || return 0
  awk -v c="$1" -v k="$2" '
    /^[[:space:]]*#/ { next }
    NF < 3           { next }
    $1 == c && $2 == k { $1=""; $2=""; sub(/^[[:space:]]+/, ""); print; exit }
  ' "$MAP"
}

# emit <kind> <src> <name>  -> writes the translated .opencode file to stdout
# kind: commands | agents
emit() {
  local kind="$1" src="$2" name="$3" desc
  desc="$(awk '
    NR==1 && $0=="---" { f=1; next }
    f && $0=="---"     { exit }
    f && sub(/^description:[[:space:]]*/, "") { print; exit }
  ' "$src")"

  # reject malformed source rather than silently emitting empty output (also catches CRLF: `---\r`)
  if ! head -n1 "$src" | grep -q '^---$'; then
    echo "[!] $name: source does not open with a '---' frontmatter block (CRLF line endings?)" >&2
    exit 1
  fi
  if [ -z "$desc" ]; then
    echo "[!] $name: no 'description:' found in frontmatter" >&2
    exit 1
  fi

  echo "---"
  printf 'description: %s\n' "$desc"
  if [ "$kind" = "commands" ]; then
    for k in agent model subtask; do
      v="$(override "$name" "$k")"
      [ -n "$v" ] && printf '%s: %s\n' "$k" "$v"
    done
  else
    echo "mode: subagent"
  fi
  echo "---"
  awk '
    NR==1 && $0=="---" { f=1; next }
    f && $0=="---"     { f=0; b=1; next }
    b                  { print }
  ' "$src"
}

# gen_dir <kind> <src_dir> <out_dir>  (globs only *.md directly in src_dir; skips subdirs like agents/extensions)
gen_dir() {
  local kind="$1" src_dir="$2" out_dir="$3" src name
  mkdir -p "$out_dir"
  for src in "$src_dir"/*.md; do
    [ -e "$src" ] || continue
    name="$(basename "$src" .md)"
    emit "$kind" "$src" "$name" > "$out_dir/$name.md"
  done
}

# check_dir <committed> <regenerated> -> returns 1 on drift, prints diagnostics
check_dir() {
  if ! diff -r "$1" "$2" >/dev/null 2>&1; then
    echo "[!] ${1##*/opencode/} is stale — run: golden/scripts/gen-opencode.sh" >&2
    diff -r "$1" "$2" >&2 || true
    return 1
  fi
}

if [ "$CHECK" -eq 1 ]; then
  gen_dir commands "$GOLDEN_DIR/.claude/commands" "$TMP/commands"
  gen_dir agents   "$GOLDEN_DIR/.claude/agents"   "$TMP/agents"
  rc=0
  check_dir "$GOLDEN_DIR/.opencode/commands" "$TMP/commands" || rc=1
  check_dir "$GOLDEN_DIR/.opencode/agents"   "$TMP/agents"   || rc=1
  [ "$rc" -eq 0 ] && echo "[+] .opencode/ is up to date with .claude/" || exit 1
else
  gen_dir commands "$GOLDEN_DIR/.claude/commands" "$GOLDEN_DIR/.opencode/commands"
  gen_dir agents   "$GOLDEN_DIR/.claude/agents"   "$GOLDEN_DIR/.opencode/agents"
fi
