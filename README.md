# Stark Garage

An agentic engineering toolbox for Claude Code. 13 slash commands and 7 specialized agents that handle everything from a single bug fix to a full autonomous release with parallel builds and deep security review.

Built from a synthesis of two production-tested systems keeping what worked and cutting what didn't.

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

## Scenarios

### 1. Quick Bug Fix (5 minutes)

A user reports that the `/api/users` endpoint returns 500 when the email field is empty. You have an issue for it already.

```
> /wiggum 87
```

That's it. Wiggum reads issue #87, creates a branch (`87-fix-empty-email-500`), writes a regression test that sends an empty email and asserts a 400 response, implements the validation, runs the project's validation suite, creates a PR, runs `/review-pr`, calls `/close-issue 87` to verify the acceptance criteria, and merges. One command, fully autonomous.

### 2. Investigating a Feature Request Before Building

A product manager asks: "Can we add role-based access control to the API?" You don't know the codebase well enough to estimate the work yet.

```
> /investigate add role-based access control to the API
```

Investigate launches parallel subagents to explore the codebase — one looks at the auth middleware, one maps the existing endpoints and their permission checks, one examines the database schema for user/role tables. It probes the running local API to see what `/api/users` actually returns today and what headers are required.

It comes back with a structured impact analysis:

```
| Layer     | Files/Areas              | Nature of Change        |
|-----------|--------------------------|-------------------------|
| Storage   | internal/storage/        | New roles table + migration |
| Auth      | internal/auth/           | Extend JWT claims, new middleware |
| Handlers  | internal/api/            | Add permission checks to 12 endpoints |
| Tests     | internal/api/*_test.go   | Update all handler tests with role fixtures |
```

You iterate — "What about the admin panel?" — and it explores further. Once you're aligned on scope:

```
> /create-issues me
```

Create-issues detects the plan from your conversation, drafts a tracking epic plus 6 implementation issues with explicit dependencies (`- Blocked by: #91 -- needs roles table before middleware`), validates the dependency graph for cycles, and presents everything for your approval before creating.

### 3. Triaging a Messy Backlog

You've got 30 open issues and aren't sure what to work on first.

```
> /triage
```

Triage fetches all open issues, parses every `- Blocked by: #NN` dependency, builds the full dependency graph, detects any cycles, and computes impact scores — how many other issues each one transitively unblocks. Output:

```
## Highest Impact (unblocks the most work)
| Issue | Impact | Blocks |
|-------|--------|--------|
| #91 — feat(storage): Add roles table | 5 | #92, #93, #94, #95, #96 |
| #88 — fix(auth): Token refresh race  | 2 | #89, #90 |

## Ready (no blockers)
- #91 — feat(storage): Add roles table
- #88 — fix(auth): Token refresh race
- #85 — docs: Update API reference

## Blocked
- #92 — feat(auth): Role middleware (blocked by #91)
- #93 — feat(api): Permission checks (blocked by #92)

## Label Issues
- #94 has `blocked` label but all blockers are closed — remove label?
```

It offers to fix the label inconsistencies automatically.

### 4. Planning and Shipping a Full Release

You've investigated the RBAC feature, created the issues, and triaged the backlog. Time to ship.

```
> /setup-release enhancement
```

Setup-release filters for enhancement issues, runs `/blast-radius` on each to map cross-package impact (the roles migration touches storage, auth, and all 12 API handlers), invokes the Indexer agent to build a fresh codebase state file, creates a milestone, a release branch, and a draft PR with a phased checklist:

```
## Scope Analysis
- Issues: 6
- Packages touched: storage, auth, api, tests
- Max dependency depth: 3
- Cross-package impact: HIGH (auth middleware change affects all handlers)

## Implementation Checklist
### Phase 1 — No dependencies
- [ ] #91 — feat(storage): Add roles table and migration
### Phase 2 — Depends on Phase 1
- [ ] #92 — feat(auth): Role-based middleware
### Phase 3 — Depends on Phase 2
- [ ] #93 — feat(api): Add permission checks to endpoints
- [ ] #94 — feat(api): Admin-only endpoints
### Phase 4 — Depends on Phase 3
- [ ] #95 — feat(api): Role assignment endpoints
- [ ] #96 — docs: RBAC documentation
```

Now run the release:

```
> /wiggum
```

Wiggum detects the release context (release branch + milestone). For each issue in dependency order, the full agent pipeline activates:

1. **Planner** reads the state file and enriches issue #91 with detailed specs — exact migration SQL, repository interface signatures, test fixtures needed, plus "Known Pitfalls" from memory (e.g., a past bug where GORM's zero-value handling broke boolean defaults).

2. **Builder** implements the migration + repository + tests. For issues touching multiple packages, it spawns parallel sub-agents.

3. Three **reviewers** run in parallel — the code Reviewer checks against the Planner's spec, the Security Reviewer audits the new auth middleware for IDOR and privilege escalation vectors, and the Ops Reviewer verifies logging and health check coverage.

4. If any reviewer flags issues, the **Debugger** reads all review reports and fixes them, prioritizing CRITICAL over HIGH over MEDIUM. Re-runs only the failed reviewers. Max 3 fix iterations.

5. `/close-issue` validates every acceptance criterion from the issue, posts a structured closing comment, and unblocks downstream issues.

6. Merge to release branch, check off the item in the draft PR, pick up the next issue.

When all 6 issues are closed, wiggum runs final validation and marks the draft PR ready for review. You merge to main when you're satisfied.

### 5. Deep-Reviewing a Security-Sensitive PR

A teammate opens a PR that modifies the authentication middleware. You want more than the standard review.

```
> /review-pr 142 --deep
```

The `--deep` flag (also auto-triggered because the PR touches auth) launches three agent reviewers in parallel:

- **Reviewer** checks the implementation against the spec in the linked issue, verifies cross-package consistency, flags if any consuming handler wasn't updated.
- **Security Reviewer** runs OWASP checks — is the new permission check vulnerable to IDOR? Are JWT claims validated correctly? Any injection vectors in the role name parameter? Dependency CVE scan on new packages.
- **Ops Reviewer** audits whether the auth middleware logs failed permission checks with request IDs, whether health endpoints still pass, whether the new middleware has a timeout.

Each produces a structured report in `.claude/reviews/`. The aggregate verdict determines if the PR is approved.

### 6. Checking Blast Radius Before a Refactor

You want to rename the `UserService` interface but aren't sure how far the impact reaches.

```
> /blast-radius UserService
```

Blast-radius traces all imports, call chains (3 levels deep), test coverage, and downstream consumers:

```
## Blast Radius: UserService

### Direct References (8 files)
- internal/services/user.go:15 — interface definition
- internal/api/users.go:32 — handler dependency injection
- internal/api/admin.go:18 — admin handler uses UserService
...

### Test Coverage
- 4 test files exercise UserService
- Tests at risk: user_test.go, admin_test.go, auth_test.go, integration_test.go

### Downstream Packages
- internal/api/ — depends via import "internal/services"
- internal/auth/ — depends via UserService.GetByID

### Risk Assessment: MODERATE
8 files across 3 packages. Recommend creating issues for the refactor.
```

### 7. Post-Mortem After a Tricky Debug Session

You just spent 30 minutes debugging a race condition where the token refresh goroutine was writing to a map concurrently with the request handler. The fix was adding a sync.RWMutex.

```
> /pomo
```

Pomo reconstructs the incident from your conversation, evaluates whether it's a generalizable pattern (yes — concurrent map access without synchronization), checks memory and lessons.md for duplicates, and writes:

**To `.claude/lessons.md`:**
```markdown
### Concurrent map access
- **Wrong:** Sharing a map between goroutines without synchronization
- **Right:** Use sync.RWMutex or sync.Map for concurrent map access
- **Why:** Go maps are not goroutine-safe; concurrent read+write causes fatal runtime panic
```

**To memory:**
```bash
toolbox-memory write --ns lesson --agent pomo --key "race-condition-map" \
  --value '{"wrong":"shared map without sync","right":"sync.RWMutex","why":"runtime panic"}'
```

Next time the Planner generates a spec for a feature involving shared state, it queries memory, finds this pattern, and adds it to "Known Pitfalls" in the issue spec — before any code is written.

### 8. Onboarding a New Project

You're starting work on a new codebase for the first time.

```bash
./golden/deploy.sh ~/projects/new-api
cd ~/projects/new-api
claude
```

```
> /bootstrap
```

Bootstrap scans the project — detects Go 1.22, chi router, PostgreSQL, Docker Compose, Makefile with `make test` and `make lint`. It asks you to confirm the profile, then adapts:

- Appends project-specific sections to CLAUDE.md (architecture rules, validation command, scopes)
- Generates `agent_docs/build-and-test.md` with actual Makefile targets
- Generates `agent_docs/project-structure.md` with the directory layout
- Configures `.mcp.json` with Playwright and Chrome DevTools MCP
- Initializes the memory database
- Creates a project-specific `/add-endpoint` command if it detects a REST API pattern

Now every command knows how to build, test, and lint this specific project.

### 9. Evolving the Toolbox Across Projects

After working on Project A for a month, you've accumulated lessons and improved some commands. You want Project B (and all future projects) to benefit.

```
> /improve-golden-set ~/golden-set
```

Improve-golden-set scans Project A's customizations, classifies them as golden-original / modified / novel, and proposes extractions. A new CLAUDE.md instruction about "always validate pagination params" gets generalized (strip project-specific API paths) and offered for promotion to the baseline.

Later, on Project B:

```
> /update-claude ~/golden-set
```

Update-claude diffs Project B's configuration against the updated golden set, shows what changed, and applies approved updates — without touching Project B's custom sections below the bootstrap marker.

If the golden set is getting bloated:

```
> /slim
```

Slim measures every file against its budget, scans for redundant instructions (duplicated in commands, or trained-in behavior Claude already knows), prunes stale lessons, runs `toolbox-memory prune` for memory lifecycle transitions, and reports before/after utilization.

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
