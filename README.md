# Stark Garage

An agentic engineering toolbox for Claude Code. 15 slash commands, 7 specialized agents, and a browser-automation skill that handle everything from a single bug fix to a full autonomous release with parallel builds and deep security review.

**GitHub Issues are the backbone.** Every piece of work is a GitHub issue. Specs live in issue bodies. Dependencies are tracked in issue bodies. Milestones scope releases. The `gh` CLI drives everything.

## The Core Dev Loop

Five commands do most of the work. The loop runs: **investigate → create-issues → setup-release → wiggum → review-pr**.

```
/investigate      →  deep-dive a feature, iterate on scope, produce a plan
     ↓
/create-issues    →  turn the plan into GitHub issues (epic + children with deps)
     ↓
/setup-release    →  scope issues into a milestone, index the codebase, draft a release PR
     ↓
/wiggum           →  autonomous loop: for each issue, plan → build → review → close → merge
     ↓
/review-pr        →  final aggregate review on the release PR (deep, parallel agents)
```

### Step 1 — `/investigate`

You don't know the codebase well enough to scope the work. Investigate launches parallel subagents to explore — auth middleware here, DB schema there, running API over there — and returns an impact-analysis table. You iterate on scope with it until you're aligned. It's read-only and never edits files.

```
> /investigate add role-based access control to the API
```

### Step 2 — `/create-issues`

The plan becomes GitHub issues. Create-issues auto-detects whether you're describing one thing or a multi-step plan. For a plan, it drafts a tracking epic plus ordered children with explicit dependencies (`- Blocked by: #91 — needs roles table before middleware`), validates the dependency graph for cycles, and asks for approval before creating.

```
> /create-issues me
```

### Step 3 — `/setup-release`

Scope issues into a milestone. Setup-release filters issues by label or range, runs `/blast-radius` on each to map cross-package impact, invokes the **Indexer** agent to build a fresh codebase state file, creates the milestone, the release branch, and a draft PR with a phased implementation checklist.

```
> /setup-release enhancement
```

### Step 4 — `/wiggum`

The workhorse. Two modes:

- **Ad-hoc** (`/wiggum 53`) — single issue, no agents, fast path: branch, TDD, implement, PR, close.
- **Release** (`/wiggum` on a release branch) — full autonomous pipeline. For each issue in dependency order: **Planner** enriches the issue with specs, **Builder** implements (parallel sub-agents for multi-package work), three reviewers run in parallel (**Reviewer**, **Security Reviewer**, **Ops Reviewer**), **Debugger** handles the fix loop, `/close-issue` validates acceptance criteria, merge to release branch, pick up the next issue.

```
> /wiggum 53       # one issue
> /wiggum          # every issue in the release
```

### Step 5 — `/review-pr`

Final aggregate review on the release PR. Standard mode posts a 7-section structured review. `--deep` escalates to three parallel agent reviewers and is auto-triggered for PRs targeting `main`/`release/*` or touching auth/security/migration files.

```
> /review-pr 142 --deep
```

That's the loop. Everything else is support.

---

## Usage Reference

### Commands

Command implementations live in [`golden/.claude/commands/`](golden/.claude/commands/). Each file is a self-contained prompt.

#### [`/investigate`](golden/.claude/commands/investigate.md) — Feature Research

Read-only deep-dive before building. Probes the codebase and running API, surfaces an impact analysis table, iterates with you on scope.

```
/investigate                           # Prompt for the feature description
/investigate add role-based access     # Inline description
/investigate #42                       # Pull context from an existing GitHub issue
```

Outputs a structured plan ready for `/create-issues`.

#### [`/create-issues`](golden/.claude/commands/create-issues.md) — Issue Creation

Creates GitHub issues from a description or plan. Auto-detects single vs. batch mode.

```
/create-issues              # Create from conversation context, leave unassigned
/create-issues me           # Assign to you (resolves via `gh api user`)
/create-issues ben          # Assign to collaborator (fuzzy-matches against repo collaborators)
```

Single description = one issue. Numbered plan = tracking epic + children with dependency graph.

#### [`/setup-release`](golden/.claude/commands/setup-release.md) — Release Preparation

Scopes issues into a release, runs blast-radius analysis, indexes the codebase, creates milestone + branch + draft PR.

```
/setup-release                         # Interactive — asks what to include
/setup-release bugs                    # Filter: issues labeled "bug"
/setup-release enhancement             # Filter: issues labeled "enhancement"
/setup-release enhancement 10-25       # Label + specific issue number range
```

#### [`/wiggum`](golden/.claude/commands/wiggum.md) — The Workhorse

Implements issues end-to-end. Ad-hoc or release mode.

```
/wiggum 53                  # Ad-hoc: implement issue #53 (no agents, no release context)
/wiggum                     # Release: auto-detect release branch + milestone
/wiggum release/v2.1        # Release: target a specific release branch
```

Bare issue number = ad-hoc fast path. No arguments on a release branch = full agent pipeline.

#### [`/review-pr`](golden/.claude/commands/review-pr.md) — PR Review

7-section standardized review. Posts structured verdict on the PR.

```
/review-pr 42               # Standard review
/review-pr 42 --diff-only   # Skip build gates, review code changes only
/review-pr 42 --deep        # Escalate to 3 parallel agent reviewers
```

`--deep` is auto-triggered for PRs targeting `main`/`release/*` or touching auth/security/migration files.

#### [`/close-issue`](golden/.claude/commands/close-issue.md) — Acceptance Validation & Closure

Validates every acceptance criterion before closing. Unblocks downstream issues automatically.

```
/close-issue 53             # Validate and close issue #53
/close-issue 53 54 55       # Close multiple issues
```

#### [`/triage`](golden/.claude/commands/triage.md) — Backlog Analysis

Builds the full dependency graph, computes impact scores, validates labels.

```
/triage                     # Analyze all open issues in the repo
```

Fetches all open issues, parses `- Blocked by: #NN` dependencies, detects cycles, scores each issue by how many others it transitively unblocks. Offers to auto-fix stale labels.

#### [`/blast-radius`](golden/.claude/commands/blast-radius.md) — Impact Analysis

Traces imports, call chains (3 levels deep), test coverage, and downstream consumers.

```
/blast-radius UserService              # Analyze a type or interface
/blast-radius HandleCreateUser         # Analyze a function
/blast-radius internal/auth/middleware # Analyze a file or directory
/blast-radius #42                      # Analyze files likely affected by an issue
```

Read-only. Also called internally by `/setup-release`.

#### [`/pomo`](golden/.claude/commands/pomo.md) — Post-Mortem

Captures lessons from debugging sessions. Writes to both [`lessons.md`](golden/examples/) and SQLite memory.

```
/pomo                                  # Reflect on what just happened in this session
/pomo #42                              # Reflect on a specific issue
/pomo https://github.com/.../pull/10   # Reflect on a specific PR
```

Auto-invoked by `/wiggum` after 2+ retry attempts and by `/review-pr` after REQUEST_CHANGES on Claude-authored PRs.

#### [`/bootstrap`](golden/.claude/commands/bootstrap.md) — Project Setup

Scans the project, detects tech stack, adapts the toolbox configuration.

```
/bootstrap                             # Interactive scan and setup
```

Run once after `deploy.sh`. Writes project-specific sections to CLAUDE.md, generates `agent_docs/build-and-test.md` and `agent_docs/project-structure.md`, configures MCP servers and permissions.

#### [`/update-claude`](golden/.claude/commands/update-claude.md) — Pull Golden Set Updates

Syncs improvements from the golden set into a bootstrapped project without overwriting customizations.

```
/update-claude                         # Fetch from default repo (bnsmcx/stark-garage)
/update-claude owner/repo              # Specify GitHub repo
/update-claude ~/path/to/golden-set    # Local path
```

Never touches project-specific content below the bootstrap marker.

#### [`/improve-golden-set`](golden/.claude/commands/improve-golden-set.md) — Extract Improvements

Reverse flow: pull generalizable improvements from a project back into the golden set.

```
/improve-golden-set ~/path/to/project  # Local mode
/improve-golden-set                    # GitHub mode (push PR to default repo)
/improve-golden-set --repo owner/repo  # GitHub mode with explicit repo
```

Classifies project changes as golden-original, modified, or novel. Generalizes before extracting.

#### [`/slim`](golden/.claude/commands/slim.md) — Audit & Compress

Prevents bloat in CLAUDE.md, agent_docs, lessons, and memory. Enforces [`BUDGETS.md`](golden/BUDGETS.md).

```
/slim                                  # Audit everything
```

Scans for redundant instructions, prunes stale lessons, runs `toolbox-memory prune`.

#### [`/release-notes`](golden/.claude/commands/release-notes.md) — Release PR Description

Generates a comprehensive, user-facing release PR description from closed milestone issues.

```
/release-notes                         # Auto-detect release branch + milestone
/release-notes v0.9.26                 # Target a specific version
```

Produces narrative "What's New" sections, API/database change tables, and implementation progress. Applies it via `gh pr edit`.

#### [`/release-demo`](golden/.claude/commands/release-demo.md) — Release Demo & E2E Validation

Generates an E2E test script from closed milestone issues, runs it, fixes failures, and records a looping VHS gif for the release PR.

```
/release-demo                          # Auto-detect release branch + milestone
/release-demo v0.9.26                  # Target a specific version
```

### Agents

Auto-invoked during release mode and deep reviews. You never call these directly. Definitions live in [`golden/.claude/agents/`](golden/.claude/agents/).

| Agent | Role | Verdict |
|-------|------|---------|
| [**Indexer**](golden/.claude/agents/indexer.md) | Crawls codebase, builds state file for Planner context | (data, no verdict) |
| [**Planner**](golden/.claude/agents/planner.md) | Enriches GitHub issues with specs from state file + memory | (enriches issues) |
| [**Builder**](golden/.claude/agents/builder.md) | Spawns parallel sub-agents for multi-package builds | BUILD_COMPLETE / FAILED / PARTIAL |
| [**Reviewer**](golden/.claude/agents/reviewer.md) | Deep code review + spec compliance | APPROVED / NEEDS_FIXES / BLOCKING |
| [**Security Reviewer**](golden/.claude/agents/security-reviewer.md) | OWASP, CVE, secrets, auth, input validation | SECURE / WARNINGS / VULNERABLE |
| [**Ops Reviewer**](golden/.claude/agents/ops-reviewer.md) | Logging, health checks, timeouts, metrics | PRODUCTION_READY / NEEDS_INSTRUMENTATION / NOT_READY |
| [**Debugger**](golden/.claude/agents/debugger.md) | Bug diagnosis + automatic pattern learning to memory | FIXED / CANNOT_REPRODUCE / ESCALATE |

**When they fire:**
- `/setup-release` → Indexer
- `/wiggum` (release mode), per issue → Planner → Builder → Reviewer + Security + Ops (parallel) → Debugger (on NEEDS_FIXES)
- `/review-pr --deep` → Reviewer + Security + Ops (parallel)

### Skills

#### [Browser Automation](golden/skills/browser-automation/SKILL.md)

Auto-triggered for any browser interaction — navigating pages, filling forms, taking screenshots, testing web apps, verifying UI changes. Three tools installed together:

| Tool | Best for |
|------|----------|
| **agent-browser** (CLI) | Scripted flows, before/after diffing, batch operations, auth persistence |
| **Playwright MCP** | Exploratory interaction, drag-and-drop, atomic form fills |
| **Chrome DevTools MCP** | Console debugging, network inspection, Web Vitals |

Core workflow: `navigate → snapshot → interact → re-snapshot`. Without this, agents generate frontend code blind.

---

## Quick Start

```bash
# Deploy to any project
./golden/deploy.sh /path/to/your/project

# Then open Claude Code in that project
cd /path/to/your/project
claude

# Scan the project and adapt configuration
> /bootstrap

# Pick up a GitHub issue and fix it:
> /wiggum 53

# Or run the full loop:
> /investigate add user authentication
> /create-issues me
> /setup-release enhancement
> /wiggum
```

---

## Technical Details

The sections below are reference material — read when you need them.

### State Management — GitHub Issues as Source of Truth

This toolbox is built around GitHub Issues as the single state machine for all work. Not a separate project tracker, not a local task file — GitHub Issues, managed via the `gh` CLI.

**What lives in issues:**
- **Task definitions** — summary, acceptance criteria, implementation notes
- **Specs** — the Planner agent enriches issue bodies directly with schemas, API shapes, and known pitfalls
- **Dependencies** — canonical format `- Blocked by: #NN — reason`, parsed by every command
- **Status** — open or closed, nothing more complex
- **Release scope** — issues are assigned to milestones by `/setup-release`

**What does NOT live in issues:**
- Codebase structure — that's the project state file (`.claude/project-state.md`), a cache built by the Indexer agent.

There's a `tasks/todo.md` fallback for projects without a GitHub remote, but the full pipeline (milestones, labels, dependency unblocking, release PRs) requires GitHub Issues.

### Memory System

Three complementary stores, each with a narrow scope:

- **Auto-memory flat files** (`~/.claude/projects/<slug>/memory/`) — user, feedback, project, and reference memories. System-prompt-native; Claude writes and reads these automatically.
- **`.claude/lessons.md`** — project-scoped learned patterns with an in-file markdown lifecycle (`## Active` → `## Validated` → `## Promoted`, with archived entries moving into `.claude/lessons-archive.md`). Managed by `/pomo`.
- **`toolbox-memory`** — SQLite + FTS5 store for agent-emitted signal only: `bug_pattern`, `spec_gap`, `calibration`, `routing`. Every bug fix and estimation miss is recorded automatically.

```bash
# Build the CLI
cd golden && go build -o toolbox-memory ./cmd/toolbox-memory/

# Query directly:
toolbox-memory search --ns bug_pattern --query "nil pointer"
toolbox-memory stats
```

| Trigger | What's recorded |
|---------|----------------|
| Debugger fixes a bug | Bug class, root cause, prevention strategy (`bug_pattern`) |
| Reviewer finds spec gap | What the spec should have included (`spec_gap`) |
| Builder completes a feature | Estimated vs actual hours (`calibration`) |

`toolbox-memory` entries follow a simpler lifecycle: **active** (on write) → **validated** (once `hit_count >= 2`, applied by `toolbox-memory prune`) → optionally **promoted** (explicit) → **stale** (60 days idle) → **archived** (30 days after stale). `/pomo` runs lessons.md; `/slim` runs `toolbox-memory prune`.

### Golden Set Lifecycle

The toolbox improves across projects:

```
./deploy.sh  →  /bootstrap  →  work  →  /improve-golden-set  →  /update-claude
   (install)    (adapt)                  (extract learnings)     (propagate)
                                                 ↓
                                         /slim (prevent bloat)
```

### Project Structure

```
golden/
  CLAUDE.md                          # Baseline instructions (deployed to projects)
  BUDGETS.md                         # Line/instruction limits
  deploy.sh                          # Install into any project
  .claude/
    commands/                        # 15 slash commands
    agents/                          # 7 agent definitions
    settings.local.json              # Baseline permissions
  agent_docs/                        # On-demand reference docs
  skills/browser-automation/         # Browser tool decision guide
  cmd/toolbox-memory/                # Go CLI source
  internal/memory/                   # SQLite+FTS5 implementation
  tests/
    smoke-test.sh                    # Deploy verification (29 checks)
    cli-integration-test.sh          # Memory CLI end-to-end (25 checks)
  examples/                          # CLAUDE.md example, lessons template
```

### Testing

```bash
cd golden

# Tier 1: Go unit tests (26 tests)
go test ./internal/memory/ -v

# Tier 2: Deploy smoke test (29 checks)
bash tests/smoke-test.sh

# Tier 3: CLI integration test (25 checks)
go build -o toolbox-memory ./cmd/toolbox-memory/
bash tests/cli-integration-test.sh
```

### Design Decisions

1. **15 commands + 7 agents**, not 29 agents. Composability over completeness.
2. **Claude Code only.** No Copilot/Cursor variants. Biggest maintenance win.
3. **Two clean modes.** Ad-hoc or release. No heuristics.
4. **Memory writes are automatic.** Triggered by events, never manual.
5. **Issues are the task truth.** State file is codebase truth. No overlap.
6. **Retry budget of 3 everywhere.** Prevents infinite loops.
7. **Golden set evolves.** Every project improves the template.

### Origins

This toolbox is a hybrid of two systems, each battle-tested on production projects:

- **The [Avengers](https://github.com/terrancedjones/Avengers)** — 29 specialized agents with pipeline orchestration, parallel review, SQLite+FTS5 memory, and 6 autopilot modes. Great depth, high maintenance cost.
- **The [Skills System](https://github.com/quadradad/claude-bootstrapping)** — 18 composable slash commands with issues as state, fail-forward loops, and a golden set portability pattern. Great simplicity, limited depth.

The deep-dive analysis that informed this design is in the companion document [`agentic-engineering-deep-dive.md`](https://github.com/bnsmcx/stark-garage).
