# Changelog

## 2026-04-01 — /improve-golden-set from Athena v2 services-0.9.26

### Added
- `/release-demo` command (Level 0) — generate E2E test script, run it, record looping VHS gif
- `/release-notes` command (Level 0) — generate full narrative release PR description from milestone issues
- Schema/DDL consistency check in `/investigate` Step 4 and `/wiggum` Step 6b

### Changed
- `/wiggum` Step 8: deep review is now complexity-gated (<50 lines AND <3 files = skip per-issue review)
- `/wiggum` Release Completion: aggregate `/review-pr` on release PR is now a hard gate; calls `/release-notes` and `/release-demo`
- `/update-claude`: now fetches from GitHub via `gh` by default (bnsmcx/stark-garage); local path still supported as fallback; diverged items offer bidirectional sync (push local to golden)
- `/improve-golden-set`: GitHub mode added — can run from any project and push a PR to the golden set repo via `gh`

### Budget impact
- Commands: 13 -> 15 (+2: release-demo, release-notes)
- agent_docs/: unchanged
- CLAUDE.md baseline: unchanged

## v1.0.0 — 2026-03-31

Initial release of the Agentic Engineering Toolbox.

### Added
- 13 slash commands: /triage, /create-issues, /wiggum, /review-pr, /close-issue, /pomo, /investigate, /setup-release, /blast-radius, /bootstrap, /update-claude, /improve-golden-set, /slim
- 7 agents: Indexer, Planner, Builder, Reviewer, Security Reviewer, Debugger, Ops Reviewer
- SQLite+FTS5 memory system with lifecycle management (toolbox-memory CLI)
- Browser automation skill (agent-browser + Playwright MCP + Chrome DevTools MCP)
- Golden set lifecycle: deploy.sh, /bootstrap, /update-claude, /improve-golden-set, /slim
- agent_docs/: issue-conventions, issue-tracker-ops, self-improvement, build-and-test, project-structure
- Two execution modes: ad-hoc (/wiggum 53) and release (/setup-release + /wiggum)
