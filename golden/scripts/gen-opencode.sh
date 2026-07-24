#!/usr/bin/env bash
set -euo pipefail

# Derive .opencode/commands/*.md from .claude/commands/*.md (the single source of truth).
#
# .claude/commands is canonical. For each command this generator:
#   - drops `name` (implicit from filename in both tools) and `user_invocable`
#   - keeps `description`
#   - appends any per-command overrides from commands/opencode.map (agent/model/subtask)
#   - copies the body (everything after the closing frontmatter `---`) VERBATIM
#
# Uses only bash/awk/sed — no yq/jq/python (none are guaranteed on target machines;
# deploy.sh's prerequisites are git/gh/node/npm/claude).
#
# Usage:
#   gen-opencode.sh            (re)generate golden/.opencode/commands/
#   gen-opencode.sh --check    regenerate into a temp dir and diff vs the committed
#                              output; exit 1 on drift (used by smoke-test.sh)

GOLDEN_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SRC_DIR="$GOLDEN_DIR/.claude/commands"
MAP="$GOLDEN_DIR/commands/opencode.map"
COMMITTED="$GOLDEN_DIR/.opencode/commands"

CHECK=0
[ "${1:-}" = "--check" ] && CHECK=1

if [ "$CHECK" -eq 1 ]; then
  OUT_DIR="$(mktemp -d)"
  trap 'rm -rf "$OUT_DIR"' EXIT
else
  OUT_DIR="$COMMITTED"
  mkdir -p "$OUT_DIR"
fi

# override <command> <key> -> prints the override value, or nothing
override() {
  [ -f "$MAP" ] || return 0
  awk -v c="$1" -v k="$2" '
    /^[[:space:]]*#/ { next }              # comments
    NF < 3           { next }              # blank / malformed
    $1 == c && $2 == k { $1=""; $2=""; sub(/^[[:space:]]+/, ""); print; exit }
  ' "$MAP"
}

for src in "$SRC_DIR"/*.md; do
  name="$(basename "$src" .md)"
  # description from the first `---` ... `---` frontmatter block
  desc="$(awk '
    NR==1 && $0=="---" { f=1; next }
    f && $0=="---"     { exit }
    f && sub(/^description:[[:space:]]*/, "") { print; exit }
  ' "$src")"

  # Reject malformed source rather than silently emitting an empty file. The first
  # line must be exactly `---` (this also catches CRLF, whose opener is `---\r`),
  # and a description must be present.
  if ! head -n1 "$src" | grep -q '^---$'; then
    echo "[!] $name: source does not open with a '---' frontmatter block (CRLF line endings?)" >&2
    exit 1
  fi
  if [ -z "$desc" ]; then
    echo "[!] $name: no 'description:' found in frontmatter" >&2
    exit 1
  fi

  {
    echo "---"
    printf 'description: %s\n' "$desc"
    for k in agent model subtask; do
      v="$(override "$name" "$k")"
      [ -n "$v" ] && printf '%s: %s\n' "$k" "$v"
    done
    echo "---"
    # body: everything after the SECOND `---` (verbatim; bodies contain their own
    # `---` and code fences, so only the first block is treated as frontmatter)
    awk '
      NR==1 && $0=="---" { f=1; next }
      f && $0=="---"     { f=0; b=1; next }
      b                  { print }
    ' "$src"
  } > "$OUT_DIR/$name.md"
done

if [ "$CHECK" -eq 1 ]; then
  if ! diff -r "$COMMITTED" "$OUT_DIR" >/dev/null 2>&1; then
    echo "[!] .opencode/commands is stale — run: golden/scripts/gen-opencode.sh" >&2
    diff -r "$COMMITTED" "$OUT_DIR" >&2 || true
    exit 1
  fi
  echo "[+] .opencode/commands is up to date with .claude/commands"
fi
