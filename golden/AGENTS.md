# AGENTS.md — cross-tool agent instructions

OpenCode (and other `AGENTS.md`-aware tools) load this file as the project's agent instructions.
Claude Code uses `CLAUDE.md`. To keep a single source of truth and avoid drift, **the authoritative
project rules live in [`CLAUDE.md`](./CLAUDE.md)** — read it in full; everything there applies here.

@CLAUDE.md

> If your tool does not expand the `@CLAUDE.md` import above, open and read `CLAUDE.md` directly before
> proceeding — it is the real ruleset (development philosophy, validation command, issue conventions,
> architecture rules, and project-specific configuration below its bootstrap marker).

## OpenCode notes

- **Slash commands** are in `.opencode/commands/` (generated from `.claude/commands/` — same commands,
  same bodies; see `golden/scripts/gen-opencode.sh`). Invoke them as `/name` in the OpenCode TUI.
- The **agent-pipeline commands** (`/wiggum`, `/setup-release`, `/review-pr`) were designed around
  Claude Code's subagent model. They work in OpenCode, but multi-agent fan-out and permission behavior
  differ under OpenCode's agent model — use with that awareness.
- **Memory** is harness-native (`MEMORY.md` + one file per fact); there is no memory CLI or database.

## Available commands

Invoke as `/name` in the OpenCode TUI. (Generated from `.claude/commands/` frontmatter by
`golden/scripts/gen-opencode.sh` — do not edit the block below by hand.)

<!-- BEGIN generated command index (gen-opencode.sh) -->
- `/blast-radius` — Lightweight impact analysis — trace imports, call chains, test coverage, and downstream consumers for any code target
- `/bootstrap` — Scan the current project and adapt the full Agentic Engineering Toolbox to its tech stack, conventions, and structure
- `/close-issue` — Validate acceptance criteria, close issue, unblock downstream, write to memory
- `/create-issues` — Create issues — single or batch. Auto-detects mode from input.
- `/improve-golden-set` — Extract generalized improvements from a bootstrapped project back into the golden set
- `/investigate` — Research a feature request — deep-dive the codebase, probe the local API, surface tradeoffs, and iterate toward a plan ready for /create-issues
- `/pomo` — Post-mortem reflection — capture lessons and write to memory
- `/release-notes` — Generate a scannable, inverted-pyramid release PR description from closed milestone issues
- `/review-pr` — Review a PR — 7-section standardized review with optional deep agent escalation
- `/setup-release` — Plan a release — blast-radius analysis, index the codebase, scope issues, create milestone, branch, and phased implementation plan
- `/slim` — Audit the golden set for bloat, redundancy, and budget compliance — compress, prune memory, and remove content to stay within limits
- `/triage` — Analyze the issue backlog — dependency graph, readiness, label validation, and prioritization
- `/update-claude` — Pull golden set updates into a bootstrapped project while preserving project-specific customizations
- `/wiggum` — The workhorse — single issue or autonomous release loop with full agent pipeline
<!-- END generated command index -->
