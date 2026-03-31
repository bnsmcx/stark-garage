---
name: browser-automation
description: Use when any task requires browser interaction — navigating pages, clicking, filling forms, taking screenshots, testing web apps, verifying UI changes, scraping data, or automating browser workflows. Triggers include "open a website", "check this page", "fill out a form", "take a screenshot", "test this in the browser", "verify the UI", or any task needing programmatic web interaction.
allowed-tools: Bash(agent-browser:*), Bash(npx agent-browser:*)
---

# Browser Automation

Three browser tools are available. Pick the right one per task.

## Quick Reference

| Scenario | Tool | Why |
|----------|------|-----|
| Quick page check | agent-browser | `open && wait && snapshot -i` in one Bash call |
| Multi-step known flow | agent-browser | Chain commands with `&&`, no round-trips |
| Before/after verification | agent-browser | `diff snapshot` built-in |
| Visual debugging with labels | agent-browser | `screenshot --annotate` |
| Auth state persistence | agent-browser | `state save/load` across sessions |
| iOS Simulator testing | agent-browser | `-p ios --device "iPhone 16 Pro"` |
| Parallel page testing | agent-browser | `--session name` isolation |
| Exploratory clicking around | Playwright MCP | Snapshot → decide → act loop fits MCP tools |
| Complex drag-and-drop | Playwright MCP | `browser_drag` handles coordinates cleanly |
| Atomic multi-field form fill | Playwright MCP | `browser_fill_form` fills many fields at once |
| Mid-workflow JS evaluation | Playwright MCP | `browser_evaluate` with element context |
| Console error debugging | Chrome DevTools MCP | Source-mapped stack traces, error context |
| Network request inspection | Chrome DevTools MCP | See request/response headers, bodies, timing |
| Performance analysis | Chrome DevTools MCP | Web Vitals (LCP/FID/CLS), performance tracing |
| Attach to existing session | Chrome DevTools MCP | Uses your running Chrome with auth/cookies |

## agent-browser Core Workflow

Every interaction follows: **navigate → snapshot → interact → re-snapshot**

```bash
# Open and inspect
agent-browser open https://localhost:3000 && agent-browser wait --load networkidle && agent-browser snapshot -i

# Interact using refs from snapshot
agent-browser fill @e1 "user@example.com"
agent-browser fill @e2 "password123"
agent-browser click @e3

# Verify result
agent-browser wait --load networkidle
agent-browser diff snapshot  # Shows what changed
```

### Key Commands

```bash
# Navigation
agent-browser open <url>
agent-browser close

# Snapshot (always use -i for interactive elements)
agent-browser snapshot -i              # Element refs: @e1, @e2, ...
agent-browser snapshot -i -C           # Include cursor-interactive elements

# Interaction (use @refs from snapshot)
agent-browser click @e1
agent-browser fill @e2 "text"          # Clear + type
agent-browser type @e2 "text"          # Type without clearing
agent-browser select @e1 "option"
agent-browser check @e1
agent-browser press Enter
agent-browser scroll down 500

# Information
agent-browser get text @e1
agent-browser get url
agent-browser get title

# Waiting
agent-browser wait @e1                 # Wait for element
agent-browser wait --load networkidle  # Wait for network idle
agent-browser wait --url "**/page"     # Wait for URL pattern

# Capture
agent-browser screenshot              # To temp dir
agent-browser screenshot --full       # Full page
agent-browser screenshot --annotate   # Numbered labels on elements
agent-browser pdf output.pdf

# Diffing
agent-browser diff snapshot                        # Current vs last snapshot
agent-browser diff screenshot --baseline before.png # Visual pixel diff

# State persistence
agent-browser state save auth.json
agent-browser state load auth.json

# iOS Simulator
agent-browser -p ios --device "iPhone 16 Pro" open <url>
agent-browser -p ios snapshot -i
agent-browser -p ios tap @e1

# Parallel sessions
agent-browser --session site1 open <url1>
agent-browser --session site2 open <url2>

# Debugging
agent-browser --headed open <url>      # Visible browser
agent-browser highlight @e1            # Highlight element
```

### Critical Rules

- **Refs invalidate on page change** — always re-snapshot after navigation or DOM changes
- **Chain when possible** — `open && wait && snapshot` saves round-trips
- **Always close** — run `agent-browser close` when done to avoid leaked daemons
- **Use `wait --load networkidle`** after `open` for SPAs and slow pages

## Playwright MCP Core Workflow

Use MCP tools directly: `browser_navigate` → `browser_snapshot` → `browser_click` → repeat.

Best for exploratory interaction where you inspect the snapshot and decide your next action based on what you see. The MCP tool interface handles element refs and descriptions natively.

Uses accessibility tree snapshots (2-5KB) instead of screenshots, making it 10-100x more token-efficient.

## Chrome DevTools MCP

Attaches to a running Chrome instance. Use when you need:
- Console logs with source-mapped stack traces
- Network request/response inspection (headers, bodies, timing)
- Web Vitals metrics (LCP, FID, CLS)
- Performance tracing
- Access to your existing browser session (auth cookies, logged-in state)

Does NOT control the browser — use agent-browser or Playwright MCP for navigation and interaction, Chrome DevTools MCP for inspection and debugging.

## Choosing in Practice

**Default to agent-browser** for most tasks. It's faster (one Bash call vs multiple MCP round-trips), has better diffing, and supports command chaining.

**Switch to Playwright MCP** when you need the exploratory inspect-then-decide loop, complex drag-and-drop, or atomic multi-field form fills.

**Add Chrome DevTools MCP** when debugging: console errors, network issues, performance problems. It complements the other two tools — use it alongside, not instead of.

## Frontend Dev Loop Pattern

```
make change → agent-browser open localhost:3000 → snapshot → compare against intent → iterate
```

This visual feedback loop is the single highest-leverage practice for frontend work. Without it, agents generate code blind.
