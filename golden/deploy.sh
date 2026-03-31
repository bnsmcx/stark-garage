#!/usr/bin/env bash
set -euo pipefail

# Agentic Engineering Toolbox — Deploy golden set to a project
# Usage: ./deploy.sh /path/to/project

GOLDEN_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET="${1:?Usage: ./deploy.sh /path/to/project}"
ERRORS=()

if [ ! -d "$TARGET" ]; then
  echo "Error: $TARGET is not a directory"
  exit 1
fi

# --- Prerequisites check ---
MISSING=()

if ! command -v git &> /dev/null; then
  MISSING+=("git — required for all version control operations")
fi

if ! command -v gh &> /dev/null; then
  MISSING+=("gh (GitHub CLI) — required for issue tracking, PRs, milestones, and releases (https://cli.github.com)")
fi

if ! command -v node &> /dev/null || ! command -v npm &> /dev/null; then
  MISSING+=("node + npm — required for browser automation tools and MCP servers")
fi

if ! command -v go &> /dev/null; then
  MISSING+=("go — required to build toolbox-memory CLI (https://go.dev)")
fi

if ! command -v claude &> /dev/null; then
  MISSING+=("claude (Claude Code CLI) — required to run the toolbox (https://claude.ai/code)")
fi

if [ ${#MISSING[@]} -gt 0 ]; then
  echo "Missing prerequisites:"
  for dep in "${MISSING[@]}"; do
    echo "  [!] $dep"
  done
  echo ""
  read -p "Continue anyway? (y/N) " -n 1 -r
  echo ""
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

# Check gh auth status
if command -v gh &> /dev/null; then
  if ! gh auth status &> /dev/null; then
    echo "[!] gh is installed but not authenticated. Run: gh auth login"
    echo ""
  fi
fi

echo "Deploying Agentic Engineering Toolbox to $TARGET"
echo "================================================"

# --- CLAUDE.md — ask before overwriting ---
if [ ! -f "$TARGET/CLAUDE.md" ]; then
  cp "$GOLDEN_DIR/CLAUDE.md" "$TARGET/CLAUDE.md"
  echo "[+] CLAUDE.md"
else
  echo ""
  echo "CLAUDE.md already exists at $TARGET/CLAUDE.md"
  echo "  (o) Overwrite with golden set baseline"
  echo "  (m) Merge — append golden set baseline above the bootstrap marker, keep project-specific content"
  echo "  (s) Skip — keep existing file unchanged"
  read -p "Choice [o/m/s]: " -n 1 -r
  echo ""
  case "$REPLY" in
    o|O)
      cp "$GOLDEN_DIR/CLAUDE.md" "$TARGET/CLAUDE.md"
      echo "[+] CLAUDE.md (overwritten)"
      ;;
    m|M)
      # Extract project-specific content (below the bootstrap marker)
      MARKER="<!-- bootstrap: project-specific below -->"
      if grep -q "$MARKER" "$TARGET/CLAUDE.md"; then
        # Keep everything from the marker onward
        PROJECT_SPECIFIC=$(sed -n "/$MARKER/,\$p" "$TARGET/CLAUDE.md")
        cp "$GOLDEN_DIR/CLAUDE.md" "$TARGET/CLAUDE.md"
        # Append the project-specific content (golden CLAUDE.md already has the marker)
        # Replace golden's marker-and-below with the project's marker-and-below
        GOLDEN_ABOVE_MARKER=$(sed "/$MARKER/,\$d" "$GOLDEN_DIR/CLAUDE.md")
        printf '%s\n%s\n' "$GOLDEN_ABOVE_MARKER" "$PROJECT_SPECIFIC" > "$TARGET/CLAUDE.md"
        echo "[+] CLAUDE.md (merged — golden baseline + your project config)"
      else
        echo "[!] No bootstrap marker found in existing CLAUDE.md. Cannot merge."
        echo "    Keeping existing file. Use /update-claude for manual sync."
        ERRORS+=("CLAUDE.md merge failed — no bootstrap marker in existing file")
      fi
      ;;
    *)
      echo "[=] CLAUDE.md skipped"
      ;;
  esac
fi

# BUDGETS.md
cp "$GOLDEN_DIR/BUDGETS.md" "$TARGET/BUDGETS.md"
echo "[+] BUDGETS.md"

# .claude/ directory structure
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

# --- Build and install toolbox-memory ---
echo ""
echo "Building toolbox-memory..."

if command -v toolbox-memory &> /dev/null; then
  echo "[=] toolbox-memory already on PATH"
else
  if command -v go &> /dev/null; then
    INSTALL_DIR="${GOBIN:-$(go env GOPATH)/bin}"
    echo "[*] Building toolbox-memory and installing to $INSTALL_DIR ..."
    if (cd "$GOLDEN_DIR" && go build -o "$INSTALL_DIR/toolbox-memory" ./cmd/toolbox-memory/) 2>&1; then
      echo "[+] toolbox-memory installed to $INSTALL_DIR/toolbox-memory"
      # Verify it's on PATH now
      if ! command -v toolbox-memory &> /dev/null; then
        echo "[!] toolbox-memory built but $INSTALL_DIR is not on your PATH"
        echo "    Add this to your shell profile: export PATH=\"\$PATH:$INSTALL_DIR\""
        ERRORS+=("toolbox-memory built at $INSTALL_DIR/toolbox-memory but not on PATH")
      fi
    else
      echo "[!] Failed to build toolbox-memory"
      ERRORS+=("toolbox-memory build failed — run: cd $GOLDEN_DIR && go build -o /usr/local/bin/toolbox-memory ./cmd/toolbox-memory/")
    fi
  else
    echo "[!] go not found — cannot build toolbox-memory"
    ERRORS+=("toolbox-memory not built — go is not installed")
  fi
fi

# Initialize memory database
if [ ! -f "$TARGET/.claude/memory/toolbox.db" ]; then
  if command -v toolbox-memory &> /dev/null; then
    if toolbox-memory init --db "$TARGET/.claude/memory/toolbox.db" 2>&1; then
      echo "[+] .claude/memory/toolbox.db initialized"
    else
      echo "[!] Failed to init memory db"
      ERRORS+=("memory database initialization failed")
    fi
  else
    echo "[!] Skipping memory db init (toolbox-memory not available)"
    ERRORS+=("memory database not initialized — toolbox-memory not on PATH")
  fi
fi

# --- Install browser automation ---
echo ""
echo "Installing browser automation..."

if command -v agent-browser &> /dev/null; then
  echo "[=] agent-browser already installed"
else
  if command -v npm &> /dev/null; then
    echo "[*] Installing agent-browser..."
    # Try user-local install first, fall back to global with sudo
    if npm install -g agent-browser 2>/dev/null; then
      echo "[+] agent-browser installed"
    elif sudo -n npm install -g agent-browser 2>/dev/null; then
      echo "[+] agent-browser installed (via sudo)"
    else
      echo "[!] Could not install agent-browser globally (permission denied)"
      echo "    Fix: run 'sudo npm install -g agent-browser'"
      echo "    Or configure npm for user-local installs: npm config set prefix ~/.npm-global"
      ERRORS+=("agent-browser not installed — npm global install permission denied")
    fi
  else
    echo "[!] npm not found — cannot install agent-browser"
    ERRORS+=("agent-browser not installed — npm not found")
  fi
fi

# --- Summary ---
echo ""
echo "================================================"

if [ ${#ERRORS[@]} -gt 0 ]; then
  echo "Deployment completed with ${#ERRORS[@]} issue(s):"
  echo ""
  for err in "${ERRORS[@]}"; do
    echo "  [!] $err"
  done
  echo ""
  echo "Fix the issues above, then re-run deploy.sh."
  echo "(File deployment is complete — only tool installation needs attention.)"
else
  echo "Deployment complete. No issues."
fi

echo ""
CMDS=$(ls "$TARGET/.claude/commands/"*.md 2>/dev/null | wc -l)
AGENTS=$(ls "$TARGET/.claude/agents/"*.md 2>/dev/null | wc -l)
echo "Installed:"
echo "  $CMDS commands"
echo "  $AGENTS agents"
echo "  agent_docs/, skills/, memory/"
echo ""
echo "Next steps:"
echo "  1. cd $TARGET"
echo "  2. claude"
echo "  3. /bootstrap    # Scan project, adapt configuration"
echo ""
