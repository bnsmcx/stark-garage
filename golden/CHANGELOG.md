# Changelog

## [2026-08-12] — /improve-golden-set from armada-runtime

### Changed
- `scripts/gen-opencode.sh` — now preserves `name:` and `user_invocable` in generated frontmatter.
  Previously dropped both (comment: "name is implicit from filename"), but OpenCode requires `name:`
  to map `/triage` → `triage.md`. Without it, slash commands are not registered.
- `commands/opencode.map` — updated comment to reflect that `name`/`user_invocable` are now preserved.
- `AGENTS.md` — fixed script path references from `scripts/gen-opencode.sh` to
  `golden/scripts/gen-opencode.sh` (script lives under `golden/`).
- `.opencode/commands/*.md` (14 files) — regenerated with fixed script; now include `name:` and
  `user_invocable` in frontmatter.
- `.opencode/agents/*.md` (7 files) — regenerated with fixed script; now include `name:` in frontmatter.

### Added
- `.claude/settings.json` — PreToolUse hook that blocks editing any `.env*` files (except
  `.env.example`). Universal security practice: prevents accidental commits of secrets.
- `.mcp.json` — replaced `context7` server with `playwright` + `chrome-devtools`. The golden set
  already ships a `browser-automation` skill covering these tools but didn't enable them in `.mcp.json`.

### Why
- **gen-opencode.sh fix**: A downstream project (armada-runtime) discovered that none of its
  `.opencode/commands/*.md` files had `name:` in frontmatter, making all slash commands
  unrecognizable in OpenCode. The root cause was the generator stripping `name:` with the
  incorrect assumption that it was "implicit from filename."
- **.claude/settings.json**: The armada-runtime project added a PreToolUse hook to block `.env`
  edits — a universal security practice that prevents accidental credential commits. Generalizable
  to all projects.
- **.mcp.json swap**: The golden set ships a `browser-automation` skill that covers Playwright MCP
  and Chrome DevTools MCP, but `.mcp.json` only enabled `context7`. Swapping to the browser tools
  that the skill already documents makes the golden set self-consistent.

### Skipped (reviewed but not extracted)
- `.claude/agents/extensions/reviewer.md` (armada-runtime) — 9 project-specific review criteria
  tied to the armada repo B1W1 split; not generalizable.

### Budget impact
- CLAUDE.md baseline: unchanged
- Commands: unchanged (script change, no command bodies touched)
- agent_docs/: unchanged
- New files: `.claude/settings.json` (15 lines)
- Modified files: `.mcp.json` (8 → 12 lines); `scripts/gen-opencode.sh` (139 → 146 lines)

## 2026-07-25 — v1.3.0: generated command index in AGENTS.md (#57)

### Added
- `gen-opencode.sh` now generates a command index (`/name` + description from frontmatter) into a
  marked block in `AGENTS.md`, so OpenCode's agent — which loads `AGENTS.md` into context — is aware
  of the available commands and can suggest/use them proactively. `--check` covers `AGENTS.md` drift.

### Why
- Downstream OpenCode test: `.opencode/commands` are auto-discovered (invocation works), but OpenCode
  does not inject the command roster into the agent's context the way Claude Code does. A generated
  index in the loaded rules file closes that proactive-awareness gap without drift.

## 2026-07-24 — v1.3.0: port agents to .opencode/agents (#36)

### Added
- `golden/.opencode/agents/*.md` — 7 agents generated from `.claude/agents/` (frontmatter:
  `description` + `mode: subagent`; body verbatim). `gen-opencode.sh` now emits both commands and
  agents; `--check` covers both.
- `deploy.sh` deploys `.opencode/agents/`; `smoke-test.sh` asserts the count matches `.claude/agents/`.

### Portability decision (per-agent)
- **All 7 agents port** (builder, debugger, indexer, ops-reviewer, planner, reviewer, security-reviewer)
  as OpenCode **subagents** — their bodies are tool-agnostic process/verdict instructions.
- **`model` is dropped** in the OpenCode form: Claude Code aliases (indexer's `model: sonnet`) are not
  valid OpenCode `provider/model` identifiers, so ported agents use OpenCode's configured default.
  Per-agent model/permission tuning (OpenCode's richer permission model) is left as future refinement.
- **browser-automation skill** needs no port — OpenCode lists `.claude/skills/` as a discovery path, so
  the deployed skill works as-is.

### Notes
- Fourth step of the OpenCode portability epic (#43). `.opencode` remains fully derived from `.claude`.

## 2026-07-24 — v1.3.0: AGENTS.md for OpenCode rule discovery (#35)

### Added
- `golden/AGENTS.md` — thin, zero-duplication pointer that makes `CLAUDE.md` the single source of
  project rules for `AGENTS.md`-aware tools (OpenCode), with an `@CLAUDE.md` import + prose fallback
  and OpenCode-specific orientation notes. No content duplication, so nothing to drift.
- `deploy.sh` deploys `AGENTS.md`; `smoke-test.sh` asserts it lands in the target.

### Notes
- OpenCode reads `AGENTS.md` as primary (falls back to `CLAUDE.md` only when absent), so the pointer is
  load-bearing; kept as a reference rather than a copy to avoid drift with each project's CLAUDE.md.
  Third step of the OpenCode portability epic (#43).

## 2026-07-24 — v1.3.0: deploy .opencode/commands (#34)

### Changed
- `deploy.sh` now installs `.opencode/commands/*.md` into target projects (alongside
  `.claude/commands/`), so OpenCode discovers the same commands. Install summary reports both counts.
- `tests/smoke-test.sh` — asserts `.opencode/commands/` is deployed with the same count as
  `.claude/commands/` (1:1 generated).

### Notes
- Second step of the OpenCode portability epic (#43), on the #33 generator. Command bodies are
  tool-agnostic (`$ARGUMENTS`/`$1`), so the deployed OpenCode commands are functionally equivalent;
  interactive TUI confirmation is a manual/downstream check.

## 2026-07-24 — mandatory re-index (from Athena v2 services, ships in v1.3.0)

### Changed
- `/wiggum` Release Completion — new **hard-gate step 5: re-index the codebase state**, with an
  explicit instruction to verify the state file actually changed rather than trusting the Indexer's
  report. Releases are the largest source of index drift, and deferring the re-index to the next
  `/setup-release` is how a state file goes unverified across several releases while still reading as
  authoritative.
- `/wiggum` Step 4 (Planner Enrichment) — cheap two-command freshness check before invoking Planner.
  If the state file has drifted, either re-index the touched areas or declare the staleness in the
  Planner prompt. A stale state file becomes premises the Builder implements and the Reviewer
  validates against.
- `/setup-release` Step 4 — the Indexer invocation is now **staleness-gated and blocking** (was
  "recommended but not blocking"), with a decision table and a post-index verification step. If the
  Indexer is unavailable, stop and ask rather than running Planner on stale context.
- `/wiggum` Rules — two new rules covering the above.

### Notes
- Motivating incident: an Athena v2 services state file sat at `api_version: 0.27.2` after 0.28.0
  had shipped, with its own drift log admitting two detail files were "UNVERIFIED this pass." A stale
  pre-computed oracle is a *high-authority wrong answer*, which is worse than an absent one.
- Related discovery issues: #48 (planner→builder→reviewer never validates a spec's premises),
  #49 (subagent reports consumed as sources without verification). The "verify the file, not the
  report" wording here is a local instance of #49 and may be superseded by a general convention.

## 2026-07-24 — v1.3.0: OpenCode command generator (#33)

### Added
- `golden/scripts/gen-opencode.sh` — derives `.opencode/commands/*.md` from `.claude/commands/*.md`
  (the single source of truth): drops `name`/`user_invocable`, keeps `description`, applies overrides,
  copies the body verbatim. `--check` mode regenerates to a temp dir and diffs vs the committed output
  (fails on drift). Uses only bash/awk/sed — no yq/jq/python.
- `golden/commands/opencode.map` — sidecar for per-command OpenCode frontmatter overrides
  (agent/model/subtask). Currently empty; the mechanism awaits agent routing (#36).
- `golden/.opencode/commands/*.md` — 14 generated + committed OpenCode command files.

### Changed
- `golden/tests/smoke-test.sh` — new assertion: `.opencode/commands` stays in sync with
  `.claude/commands` (runs `gen-opencode.sh --check`).

### Notes
- Architecture: `.opencode` is **derived from `.claude`** (which stays canonical) rather than both being
  generated from a separate `definitions/` source — no command migration, no byte-equivalence proof.
  First step of the OpenCode cross-tool portability epic (#43); #34 wires `deploy.sh` to ship it.

## 2026-07-22 — /improve-golden-set from Athena v2 services

### Removed
- `Write(.claude/**)` from `golden/.claude/settings.local.json` `allow` list. Claude Code's
  permission checker does not recognize `Write()` rules for file-editing tools — only `Edit(path)`
  is honored — so the entry was inert and printed an on-exit warning
  (`Write(.claude/**) is not matched by file permission checks`) in every deployed project.
  `Edit(.claude/**)` (kept) already covers editing under `.claude/`.

### Why
- Surfaced downstream (Athena v2 services, 2026-07-22): the project removed the redundant rule on
  its own to silence the exit-time warning, then `/improve-golden-set` traced it back to the golden
  baseline so every future bootstrap/deploy is warning-free.

### Budget impact
- No budgeted file touched (settings.local.json is not budget-tracked).

## 2026-07-22 — v1.2.0: /update-claude reads the CHANGELOG to catch removals (#41)

### Changed
- `/update-claude` Step 8 now reads `golden/CHANGELOG.md` and, for each unsynced **Removed**/**Changed**
  entry, checks the project for both lingering files AND lingering *references* (grepping CLAUDE.md,
  `.claude/`, and `agent_docs/`). Previously it only detected whole-file removals, so removed content
  or references inside files (e.g. the toolbox-memory / lessons.md references dropped this release)
  could persist unnoticed in deployed projects. Step 10 surfaces both kinds for cleanup.

## 2026-07-22 — v1.2.0: remove /release-demo and bootstrap-generated project-specific commands (#39)

### Removed
- **`/release-demo`** command (`golden/.claude/commands/release-demo.md`). It assumed a web-API
  service with a dev binary, a servable demo route, and an OpenAPI/Bearer surface — overly
  project-specific for a general-purpose toolbox (inapplicable to CLI/library/pipeline/Markdown
  projects). Command count 15 → 14.
- **Bootstrap-generated project-specific commands.** Deleted `/bootstrap` Phase 3.4 entirely;
  bootstrap no longer generates `add-endpoint`, `add-component`, `add-pipeline-step`, or
  `update-docs`. Subsequent phases renumbered (3.4–3.11).

### Changed
- `wiggum.md` Release Completion: dropped the `/release-demo` step (remaining steps renumbered).
- `release-notes.md`: removed the live-demo header link, the demo verification line, and the
  demo-GIF prose/rule (the template no longer assumes a demo artifact).
- `update-claude.md` / `improve-golden-set.md`: skip-lists now reference "user-authored
  project-specific commands" instead of naming the (removed) generated commands.
- `README.md`: removed the `/release-demo` section; command count updated to 14 in three places.

## 2026-07-22 — v1.2.0: drop toolbox-memory + lessons.md for harness-native memory (#27)

### Decision
The golden set previously shipped two competing self-improvement memory layers — the `toolbox-memory`
SQLite+FTS5 CLI and the `.claude/lessons.md` flat file — while most command bodies had drifted to only
one of them. A downstream usage audit (Athena, 2026-07-22) found **both effectively unused**: the
`toolbox-memory` DB held a single never-read row; `lessons.md` was weeks stale. Meanwhile the
harness-native file-based memory (`MEMORY.md` index + one file per fact, auto-recalled each session)
was actively maintained. **Decision: drop both golden-set mechanisms and adopt harness-native memory
as the single source of truth.**

### Removed
- `cmd/toolbox-memory/` and `internal/memory/` (the entire Go CLI + library), `go.mod`, `go.sum`,
  `tests/cli-integration-test.sh`. The repo is now Markdown-only — no Go toolchain required.
- `.claude/lessons.md` / `.claude/lessons-archive.md` machinery: deploy.sh no longer creates them;
  `examples/lessons.md.example` deleted; the `Bash(toolbox-memory:*)` permission removed.
- deploy.sh: `go` prerequisite check, `toolbox-memory` build/install, memory-db init, `.claude/memory/` dir.

### Changed
- `CLAUDE.md` (+ `examples/CLAUDE.md.example`): "Session Start" and "Continuous Improvement" now point
  at harness-native memory.
- Agents (`builder`, `debugger`, `planner`, `reviewer`): memory read/write steps rewritten to consult
  and save native memory files (bug-pattern / spec-gap / calibration facts) instead of `toolbox-memory`.
- Commands (`pomo`, `slim`, `triage`, `wiggum`, `update-claude`, `improve-golden-set`, `bootstrap`):
  lesson capture/pruning and memory checks rewritten for native memory.
- `agent_docs/self-improvement.md`: rewritten for the native-memory format (frontmatter + `MEMORY.md` pointer).
- `BUDGETS.md`: replaced the `lessons.md` (40-entry) and `toolbox.db` (200-active) rows with a single
  harness-managed native-memory row.
- `tests/smoke-test.sh`: dropped the `.claude/memory/` and lessons-file assertions; command/agent
  counts now derived from the golden source instead of hardcoded magic numbers (also resolves #14).

### Supersedes
- #30 (gofmt drift in `main.go`) and #23 (`migrate()` error suppression) — closed as won't-fix; the
  files they targeted are deleted.

## 2026-07-22 — v1.2.0: design decision — OpenCode-compatible commands via canonical generator (#29)

### Decision
The golden set will drop its "Claude Code only" commitment and become **portable across Claude Code
and OpenCode**. Approach: **Option A — a canonical command generator.** Command/skill definitions live
in a single tool-agnostic source of truth; a generator emits both `.claude/commands/*.md` (Claude Code
format) and `.opencode/commands/*.md` (OpenCode format), so the two never drift.

Rationale: Option B (thin wrapper files) drifts over time; Option C (deploy-time mirroring) hides the
translation in `deploy.sh` and can't be reviewed as source. A canonical generator is the only approach
with a single reviewable source of truth. Simplified by the v1.2.0 memory change (#27): with
`toolbox-memory` gone, cross-tool memory reduces to native memory + a `CLAUDE.md`/`AGENTS.md` pair.

### Scope this release
Decision only — **no generator is built in v1.2.0.** Implementation is split into scoped follow-up
issues (see #29). This entry + the README "Design Decisions" update record the direction.

## 2026-05-27 — /improve-golden-set from Athena v2 services-0.17.0

### Changed
- `/release-demo`: full rewrite. Replaces the shell-E2E-script + recorded-GIF flow with a self-contained **interactive HTML page** committed to the repo, embedded into the binary, and served at a dev route — reviewers click endpoints and see live request/response instead of watching a terminal recording. Adds: a **four-tab layout** (Overview / Demo / Frontend Changes / DevOps Info), one audience per tab; **curl-formatted request previews with a copy button**; **lazy per-tab mermaid rendering** (`startOnLoad:false` + render-on-tab-activation) with the load-bearing gotcha that diagrams drawn in hidden/zero-size panels render as a "Syntax error" bomb even when the source parses; a **host-aware signed/redirect-URL rewrite** (derive host from the connection-panel API base, never hardcode `localhost`) so the demo works from a remote machine; and two documented validation modes (`file://` fast-iteration vs from-source embedded gate).
- `/release-notes`: restructured from a narrative "What's New" into an **inverted pyramid** — frontend-impact-first (breaking → new → changed → behaviour, with a mandatory `Frontend action` column), then backend / deps / risk / verification, with implementation progress + follow-ups below the fold in `<details>`. A frontend dev reads only the first section to know what (if anything) to change.
- `/improve-golden-set`: documented the changelog template to match real usage — `Added / Changed / Removed` headings, an optional `Why` section, and a per-budgeted-file `Budget impact` block keyed on `BUDGETS.md` limits (was `Added / Moved / Removed` only, with a CLAUDE.md-only budget block).

### Why
Both commands' previous golden form predated heavy real-world use. The HTML demo proved far more reviewable than GIFs across a multi-issue release; the inverted-pyramid notes let a frontend reader act from the first screen. The mermaid and host-rewrite gotchas each cost a real debugging cycle — capturing them here saves the next project from rediscovering them.

### Budget impact
- Commands: release-demo.md 163 → 178 lines; release-notes.md 193 → 263 lines; improve-golden-set.md 231 → 234 lines (all under the 300 budget).
- CLAUDE.md, agent_docs/, .mcp.json: unchanged.

### Follow-ups (not applied here)
- Reconsider re-adding `toolbox-memory` (the SQLite+FTS5 memory layer) and reconcile it against the `grep lessons.md` drift — tracked in #27. Not included in this changeset.

## 2026-05-04 — /improve-golden-set from muskrat-v2 (v2.11.0 release loop)

### Changed
- `/wiggum` Step 8 (Deep Review): replaced the binary skip-or-three-parallel gate with a three-tier table (Skip / Combined / Parallel three). New "Combined" tier handles 50–300-line non-high-stakes diffs with one sectioned reviewer agent (`[SPEC] [SECURITY] [OPS]`); roughly 30–40% cheaper in tokens than the parallel-three pipeline. The high-stakes-surface trigger keeps full rigor on auth, schema, migrations, and project-flagged privileged paths.
- `/review-pr` Deep Review Escalation: same three-tier table aligned with `/wiggum`. Adds an explicit escalation rule — if a combined-tier review surfaces concerning ambiguity in a domain it didn't fully cover, escalate to parallel-three rather than approving.

### Why
The v2.11.0 release loop ran 27 deep-review invocations (8 PRs × 3 reviewers + 1 release-level × 3). Many PRs were 100–300 lines / non-destructive; the parallel-three pipeline burned ~30% more tokens than necessary on those. The new Combined tier captures most of the structural value (sectioned coverage forces the agent to address each domain) at single-agent cost. Parallel-three remains the default for high-stakes surfaces and `--deep`.

### Budget impact
- Commands: wiggum.md 267 → 275 lines (well under 300 budget); review-pr.md 147 → 156 lines.
- CLAUDE.md, agent_docs/, .mcp.json: unchanged.

### Skipped (reviewed but not extracted)
- Lessons from `.claude/lessons.md` (artifact-size verification, ffmpeg banner suppression) — project-local lessons, not general enough to belong in the golden set.
- Memory entry `feedback_wiggum_no_mid_loop_pause.md` — already implied by `feedback_wiggum_autonomous_completion.md`; no new extraction needed.
- Two LOW follow-up items (TLS-skip log key parity; AR audit-log caller provenance) — muskrat-specific.

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
