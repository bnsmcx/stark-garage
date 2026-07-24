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
  same bodies; see `scripts/gen-opencode.sh`). Invoke them as `/name` in the OpenCode TUI.
- The **agent-pipeline commands** (`/wiggum`, `/setup-release`, `/review-pr`) were designed around
  Claude Code's subagent model. They work in OpenCode, but multi-agent fan-out and permission behavior
  differ under OpenCode's agent model — use with that awareness.
- **Memory** is harness-native (`MEMORY.md` + one file per fact); there is no memory CLI or database.
