# Changelog

## 2026-04-21 — /improve-golden-set from chartcruises v0.17.0 → v0.18.0

### Changed
- `/create-issues`: added **Quality Bar** section (anchored claims, verified line numbers, grouped ACs, explicit non-changes, named out-of-scope), mandatory **Research First** step (read source + adjacent tests + dependency reality + drift scan before drafting), and a **Canonical Format** template with grouped acceptance-criteria headings + structured Implementation Notes subsections
- `/release-notes`: new **Step 3 Release Title & Motivation** paragraph (what triggered the release, what it achieves, why now) and new **Step 4 Baseline → Target metrics table** (for performance/quality releases); remaining steps renumbered 5–11; PR-body template updated with new Motivation and Baseline → Target sections
- `/setup-release`: new **Step 10 Bump version strings** — first commit on the release branch bumps the version string so every subsequent commit has access to the new version for UI, logging, telemetry
- `/investigate` Step 1: rewritten as **Ensure the App Is Running, Then Obtain API Access** — detect a running instance, launch in background if not running (with health-check wait + log tempfile), optional dev-token fetch; new Step 7 cleanup rule to stop any server the command launched; promoted from best-effort probing to a hard requirement

### Added
- `.mcp.json` at the golden-set root with `context7` server pre-configured — matches the existing CLAUDE.md context7 guidance and eliminates the per-project manual config step after `deploy.sh`
- `Write(.claude/**)` and `Edit(.claude/**)` baseline permissions — needed by `/bootstrap`, `/slim`, `/improve-golden-set`, `/update-claude`

### Budget impact
- Commands: create-issues.md 171 → 306 lines (6 over 300 budget, justified by scope of the quality upgrade); release-notes.md +~25 lines; setup-release.md +~11 lines; investigate.md +~20 lines
- CLAUDE.md: unchanged
- agent_docs/: unchanged
- New files: `.mcp.json` (7 lines)

### Skipped (reviewed but not extracted)
- `code-reviewer` agent (deleted from chartcruises — redundant with golden's `reviewer` agent)
- `bootstrap-claude.md` (611 lines, legacy variant of existing `bootstrap.md` — deleted from chartcruises)
- `update-docs.md` (project-specific, generated per-project by `/bootstrap`)
- `add-endpoint.md` (project-specific, generated per-project by `/bootstrap`)
- Project-specific permissions: `Bash(go vet:*)`, `Bash(curl localhost:8080/...)`, hardcoded `/home/ben/` paths

## 2026-04-13 — /improve-golden-set from Athena v2 services-0.9.28

### Changed
- `/bootstrap` Step 3.3: always add Edit/Write permissions for `.claude/project-state.md` and `.claude/state/*` so the indexer agent can persist codebase index without permission denials

### Budget impact
- CLAUDE.md baseline: unchanged
- Commands: bootstrap.md +10 lines
- agent_docs/: unchanged

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
