# Changelog

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
