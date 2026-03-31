# Stark Garage

An agentic engineering toolbox for Claude Code. 13 slash commands and 7 specialized agents that handle everything from a single bug fix to a full autonomous release with parallel builds and deep security review.

Built from a synthesis of two production-tested systems (the [Avengers](https://github.com/bnsmcx/Avengers) pipeline and a composable [Skills](https://github.com/bnsmcx/claude-skills) system), keeping what worked and cutting what didn't.

## Quick Start

```bash
# Deploy to any project
./golden/deploy.sh /path/to/your/project

# Then open Claude Code in that project
cd /path/to/your/project
claude

# Scan the project and adapt configuration
> /bootstrap

# You're ready. Fix a bug:
> /wiggum 53

# Or plan and ship a full release:
> /investigate add user authentication
> /create-issues me
> /setup-release
> /wiggum
```

## How It Works

Two modes. No configuration needed — the system figures it out.

**Ad-hoc mode** (`/wiggum 53`) — Fix a single issue. No agents, no overhead. TDD implement, review, close. Done.

**Release mode** (`/setup-release` then `/wiggum`) — Full autonomous pipeline:

```
/setup-release
  |-- runs /blast-radius on every issue (cross-package impact)
  |-- invokes Indexer agent (builds codebase state file)
  |-- creates milestone, branch, draft PR with scope analysis
  
/wiggum (detects release context, activates full agent pipeline)
  |-- for each issue in dependency order:
  |     |-- Planner enriches issue with specs (schemas, APIs, known pitfalls from memory)
  |     |-- Builder implements (parallel sub-agents for multi-package work)
  |     |-- 3 reviewers run in parallel:
  |     |     |-- Reviewer (code quality + spec compliance)
  |     |     |-- Security Reviewer (OWASP, CVE, secrets, auth)
  |     |     |-- Ops Reviewer (logging, health checks, timeouts)
  |     |-- Debugger handles fix loop (max 3 iterations)
  |     |-- /close-issue validates acceptance criteria
  |     |-- merge to release branch
  |-- marks release PR ready for review when all issues closed
```

## Commands

| Command | What it does |
|---------|-------------|
| `/wiggum 53` | Implement a single issue end-to-end (TDD, review, close) |
| `/wiggum` | Autonomous release loop with full agent pipeline |
| `/create-issues` | Create one issue or batch from a plan (auto-detects) |
| `/review-pr 42` | 7-section standardized PR review |
| `/review-pr 42 --deep` | Escalate to 3 parallel agent reviewers |
| `/close-issue 53` | Validate acceptance criteria, close, unblock downstream |
| `/triage` | Dependency graph, impact scores, label validation |
| `/investigate` | Deep-dive a feature request before building |
| `/setup-release` | Blast radius + index + milestone + branch + phased plan |
| `/blast-radius` | Trace imports, call chains, test coverage for a target |
| `/pomo` | Post-mortem reflection, captures lessons to memory |
| `/bootstrap` | Scan project, detect stack, adapt toolbox configuration |
| `/update-claude` | Pull golden set updates into a project |
| `/improve-golden-set` | Extract improvements back to the golden set |
| `/slim` | Audit for bloat, prune lessons and memory |

## Agents

Auto-invoked during release mode. You never call these directly.

| Agent | Role | Verdict |
|-------|------|---------|
| **Indexer** | Crawls codebase, builds state file for Planner context | (data, no verdict) |
| **Planner** | Enriches GitHub issues with specs from state file + memory | (enriches issues) |
| **Builder** | Spawns parallel sub-agents for multi-package builds | BUILD_COMPLETE / FAILED / PARTIAL |
| **Reviewer** | Deep code review + spec compliance | APPROVED / NEEDS_FIXES / BLOCKING |
| **Security Reviewer** | OWASP, CVE, secrets, auth, input validation | SECURE / WARNINGS / VULNERABLE |
| **Debugger** | Bug diagnosis + automatic pattern learning to memory | FIXED / CANNOT_REPRODUCE / ESCALATE |
| **Ops Reviewer** | Logging, health checks, timeouts, metrics | PRODUCTION_READY / NEEDS_INSTRUMENTATION / NOT_READY |

## Memory System

SQLite + FTS5 database that gets smarter over time. Every bug fix, every review finding, every estimation miss is recorded automatically — not manually.

```bash
# Build the CLI
cd golden && go build -o toolbox-memory ./cmd/toolbox-memory/

# The system uses it automatically, but you can query directly:
toolbox-memory search --ns bug_pattern --query "nil pointer"
toolbox-memory stats
```

| Trigger | What's recorded |
|---------|----------------|
| Debugger fixes a bug | Bug class, root cause, prevention strategy |
| Reviewer finds spec gap | What the spec should have included |
| Builder completes a feature | Estimated vs actual hours |
| `/pomo` runs after retries | Lesson (wrong approach, right approach, why) |

Lessons follow a lifecycle: **Active** (new) -> **Validated** (2+ hits) -> **Promoted** (encoded into CLAUDE.md) -> **Stale** (60 days no hits) -> **Archived**.

## Browser Automation

Three tools installed together for frontend dev loops:

| Tool | Best for |
|------|----------|
| **agent-browser** (CLI) | Scripted flows, before/after diffing, batch operations |
| **Playwright MCP** | Exploratory interaction, accessibility tree snapshots |
| **Chrome DevTools MCP** | Console logs, network inspection, Web Vitals |

The pattern: `make change -> agent-browser open localhost:3000 -> snapshot -> compare -> iterate`. Without this, agents generate frontend code blind.

## Golden Set Lifecycle

The toolbox improves across projects:

```
./deploy.sh  -->  /bootstrap  -->  work  -->  /improve-golden-set  -->  /update-claude
   (install)      (adapt)                     (extract learnings)       (propagate)
                                                      |
                                              /slim (prevent bloat)
```

Every project benefits from improvements discovered in previous projects.

## State Management

**GitHub Issues** are the single source of truth for all work. Specs live in issue bodies (enriched by Planner). Dependencies use the canonical format: `- Blocked by: #NN -- reason`.

**Project state file** (`.claude/project-state.md`) is a codebase cache — packages, endpoints, schemas, coverage. Built by the Indexer agent. Read by Planner for context. Never used for task tracking.

One methodology, no parallel systems.

## Project Structure

```
golden/
  CLAUDE.md                          # Baseline instructions (deployed to projects)
  BUDGETS.md                         # Line/instruction limits
  deploy.sh                          # Install into any project
  .claude/
    commands/                        # 13 slash commands
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

## Testing

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

## Design Decisions

1. **13 commands + 7 agents**, not 29 agents. Composability over completeness.
2. **Claude Code only.** No Copilot/Cursor variants. Biggest maintenance win.
3. **Two clean modes.** Ad-hoc or release. No heuristics.
4. **Memory writes are automatic.** Triggered by events, never manual.
5. **Issues are the task truth.** State file is codebase truth. No overlap.
6. **Retry budget of 3 everywhere.** Prevents infinite loops.
7. **Golden set evolves.** Every project improves the template.

## Origins

This toolbox is a hybrid of two systems, each battle-tested on production projects:

- **The Avengers** — 29 specialized agents with pipeline orchestration, parallel review, SQLite+FTS5 memory, and 6 autopilot modes. Great depth, high maintenance cost.
- **The Skills System** — 18 composable slash commands with issues as state, fail-forward loops, and a golden set portability pattern. Great simplicity, limited depth.

The deep-dive analysis that informed this design is in the companion document [`agentic-engineering-deep-dive.md`](https://github.com/bnsmcx/stark-garage).
