#!/usr/bin/env bash
set -euo pipefail

# Agentic Engineering Toolbox — Deploy golden set to a project
# Usage: ./deploy.sh /path/to/project

GOLDEN_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET="${1:?Usage: ./deploy.sh /path/to/project}"

if [ ! -d "$TARGET" ]; then
  echo "Error: $TARGET is not a directory"
  exit 1
fi

echo "Deploying Agentic Engineering Toolbox to $TARGET"
echo "================================================"

# --- Copy golden set files ---

# CLAUDE.md — only if it doesn't exist (don't overwrite project config)
if [ ! -f "$TARGET/CLAUDE.md" ]; then
  cp "$GOLDEN_DIR/CLAUDE.md" "$TARGET/CLAUDE.md"
  echo "[+] CLAUDE.md"
else
  echo "[=] CLAUDE.md exists, skipping (use /update-claude to sync)"
fi

# BUDGETS.md
cp "$GOLDEN_DIR/BUDGETS.md" "$TARGET/BUDGETS.md"
echo "[+] BUDGETS.md"

# .claude/ directory (commands, agents, settings, skills)
mkdir -p "$TARGET/.claude/commands"
mkdir -p "$TARGET/.claude/agents/extensions"
mkdir -p "$TARGET/.claude/skills/browser-automation"
mkdir -p "$TARGET/.claude/memory"
mkdir -p "$TARGET/.claude/state"
mkdir -p "$TARGET/.claude/reviews"
mkdir -p "$TARGET/.claude/builder/checkpoints"
mkdir -p "$TARGET/.claude/autopilot"

# Commands
for f in "$GOLDEN_DIR/.claude/commands/"*.md; do
  cp "$f" "$TARGET/.claude/commands/"
  echo "[+] .claude/commands/$(basename "$f")"
done

# Agents
for f in "$GOLDEN_DIR/.claude/agents/"*.md; do
  cp "$f" "$TARGET/.claude/agents/"
  echo "[+] .claude/agents/$(basename "$f")"
done

# Skills
cp "$GOLDEN_DIR/skills/browser-automation/SKILL.md" "$TARGET/.claude/skills/browser-automation/SKILL.md"
echo "[+] .claude/skills/browser-automation/SKILL.md"

# Settings (merge, don't overwrite)
if [ ! -f "$TARGET/.claude/settings.local.json" ]; then
  cp "$GOLDEN_DIR/.claude/settings.local.json" "$TARGET/.claude/settings.local.json"
  echo "[+] .claude/settings.local.json"
else
  echo "[=] .claude/settings.local.json exists, skipping"
fi

# agent_docs/
mkdir -p "$TARGET/agent_docs"
for f in "$GOLDEN_DIR/agent_docs/"*; do
  # Skip templates — /bootstrap generates the real files
  if [[ "$(basename "$f")" == *.template ]]; then
    continue
  fi
  cp "$f" "$TARGET/agent_docs/"
  echo "[+] agent_docs/$(basename "$f")"
done

# Lessons files (only if they don't exist)
if [ ! -f "$TARGET/.claude/lessons.md" ]; then
  touch "$TARGET/.claude/lessons.md"
  echo "[+] .claude/lessons.md (empty)"
fi
if [ ! -f "$TARGET/.claude/lessons-archive.md" ]; then
  touch "$TARGET/.claude/lessons-archive.md"
  echo "[+] .claude/lessons-archive.md (empty)"
fi

# --- MCP Configuration ---
if [ ! -f "$TARGET/.mcp.json" ]; then
  cat > "$TARGET/.mcp.json" << 'MCPEOF'
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["@playwright/mcp@latest"]
    },
    "chrome-devtools": {
      "command": "npx",
      "args": ["chrome-devtools-mcp@latest"]
    }
  }
}
MCPEOF
  echo "[+] .mcp.json (Playwright MCP + Chrome DevTools MCP)"
else
  echo "[=] .mcp.json exists, skipping"
fi

# --- Install dependencies ---
echo ""
echo "Installing browser automation dependencies..."

# agent-browser CLI
if command -v agent-browser &> /dev/null; then
  echo "[=] agent-browser already installed"
else
  echo "[*] Installing agent-browser..."
  npm install -g agent-browser 2>/dev/null || echo "[!] Failed to install agent-browser (run: npm install -g agent-browser)"
fi

# toolbox-memory CLI
if command -v toolbox-memory &> /dev/null; then
  echo "[=] toolbox-memory already installed"
else
  echo "[!] toolbox-memory not found. Build from golden/cmd/toolbox-memory/ and add to PATH."
fi

# Initialize memory database
if [ ! -f "$TARGET/.claude/memory/toolbox.db" ]; then
  if command -v toolbox-memory &> /dev/null; then
    toolbox-memory init --db "$TARGET/.claude/memory/toolbox.db" 2>/dev/null || echo "[!] Failed to init memory db"
    echo "[+] .claude/memory/toolbox.db initialized"
  else
    echo "[!] Skipping memory db init (toolbox-memory not installed)"
  fi
fi

# --- Summary ---
echo ""
echo "================================================"
echo "Deployment complete."
echo ""
echo "Installed:"
CMDS=$(ls "$TARGET/.claude/commands/"*.md 2>/dev/null | wc -l)
AGENTS=$(ls "$TARGET/.claude/agents/"*.md 2>/dev/null | wc -l)
echo "  $CMDS commands"
echo "  $AGENTS agents"
echo "  agent_docs/, skills/, memory/"
echo ""
echo "Next steps:"
echo "  1. cd $TARGET"
echo "  2. claude"
echo "  3. /bootstrap    # Scan project, adapt configuration"
echo ""
