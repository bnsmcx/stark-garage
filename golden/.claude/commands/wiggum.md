---
name: wiggum
description: The workhorse — single issue or autonomous release loop with full agent pipeline
user_invocable: true
---

# /wiggum — Development Workhorse

Handles everything from a single ad-hoc issue to a full autonomous release with parallel agents.

## Invocation

```
/wiggum 53                     # Ad-hoc: implement issue #53 (no agents)
/wiggum                        # Release: auto-detect release branch, loop all issues (full agent pipeline)
/wiggum release/v1.0           # Release: target a specific release
```

## Mode Detection

1. If a **bare issue number** is provided → **Ad-hoc mode** (no agents, no release context)
2. If on a **release/* branch** or a release is specified → **Release mode** (full agent pipeline)
3. If neither → check for most recent open milestone. If found → Release mode. If not → ask user.

## Task Tracking Mode

When CLAUDE.md defines a Task Tracker section using `tasks/todo.md`:
- Read `tasks/todo.md` Active table instead of querying milestones
- Use `T-NN-slug` branch naming
- Use `Completes T-NN` in commits
- Target `main` (no release branch)
- Loop ends when Active table is empty

---

## Ad-hoc Mode (`/wiggum 53`)

Single issue, no agents, no release context. Fast path.

### 1. Memory Check

Query memory for relevant lessons:
```bash
toolbox-memory search --ns lesson --query "<issue scope keywords>"
```
If results found, note them as context warnings.

### 2. Fetch & Branch

```bash
gh issue view 53 --json number,title,body,labels
git checkout main && git pull
git checkout -b 53-short-slug
```

### 3. Understand

- Read issue description, acceptance criteria, implementation notes
- For bugs: attempt to reproduce first
- Review relevant code and architecture rules from CLAUDE.md

### 4. TDD Implement

1. Write failing tests covering acceptance criteria
2. Confirm tests fail (red)
3. Implement minimum code to pass
4. Confirm tests pass (green)
5. Refactor if needed

### 5. Validate

Run the project's validation command (from CLAUDE.md). **Hard gate.**

Retry logic (max 3 attempts):
- Attempt 1-2: analyze error, fix, re-run
- After 2 failures: re-read requirements, question approach
- Attempt 3: different strategy or final fix
- After 3 failures: revert, log failure comment on issue, stop

If 2+ retries needed → run `/pomo` with retry context.

### 6. Commit, PR, Review, Close

```bash
git add [files] && git commit -m "type(scope): description\n\nCloses #53"
git push -u origin 53-short-slug
gh pr create --base main --title "type(scope): description (#53)" --body "..."
```

Run `/review-pr` on the PR. Then run `/close-issue 53`.

### 7. Done

Report completion. Single issue, no loop.

---

## Release Mode (`/wiggum` or `/wiggum release/v1.0`)

Full autonomous loop with agent pipeline. Runs until release complete.

### 1. Context

```bash
git branch --show-current
```

Detect release branch, milestone, and release PR:
```bash
gh pr list --base main --head release/RELEASE_NAME --state open --json number,isDraft --jq '.[0].number'
```

### 2. Select Next Issue

- Fetch all open issues in milestone
- Run dependency analysis (triage logic): parse `- Blocked by: #NN`, build graph, compute impact scores
- Filter to unblocked issues only
- Sort by impact score (issues that unblock the most others first)
- Pick the top issue

If no unblocked issues remain and issues are still open → all blocked, report and stop.
If no open issues remain → release complete (jump to Release Completion).

### 3. Memory Check

```bash
toolbox-memory search --ns lesson --query "<issue scope>"
```

### 4. Planner Enrichment

Invoke the Planner agent to enrich the issue with specs:

> Use planner. Enrich issue #NN with implementation specs. Read the state file at .claude/project-state.md for codebase context.

The Planner reads the state file + memory (bug patterns, spec gaps, calibration) and appends spec sections to the issue body: schema changes, API changes, implementation hints, known pitfalls, estimated effort.

### 5. Branch

```bash
git checkout release/RELEASE_NAME && git pull
git checkout -b NN-short-slug
```

### 6. Build

For multi-package issues, invoke the Builder agent:

> Use builder. Implement issue #NN. Read specs from the issue body.

Builder spawns parallel sub-agents if the issue touches multiple packages. Manages checkpoints in `.claude/builder/`.

For single-package issues, implement directly using TDD (same as ad-hoc step 4).

### 7. Validate

Run validation command. Same retry logic as ad-hoc mode (max 3 attempts). If 2+ retries → `/pomo`.

### 8. Deep Review

Launch parallel review agents:

> Use reviewer. Review the PR for issue #NN against the spec in the issue body.
> Use security-reviewer. Security scan the PR for issue #NN.
> Use ops-reviewer. Observability audit the PR for issue #NN.

All three run in parallel. Aggregate verdicts.

### 9. Fix Loop

If any reviewer returns NEEDS_FIXES / VULNERABLE / NOT_READY:

1. Invoke Debugger agent:
   > Use debugger. Fix the issues found in .claude/reviews/. Prioritize CRITICAL > HIGH > MEDIUM.
2. Re-run only the failed reviewers
3. Max 3 fix iterations
4. After 3: escalate to human with full context

### 10. Commit & PR

```bash
git add [files]
git commit -m "type(scope): description\n\nCloses #NN"
git push -u origin NN-slug
gh pr create --base release/RELEASE_NAME --title "..." --body "..."
```

### 11. Close Issue

Run `/close-issue NN` — validates acceptance criteria, posts closing comment, unblocks downstream.

### 12. Merge & Update Release PR

```bash
gh pr merge PR_NUMBER --merge --delete-branch
git checkout release/RELEASE_NAME && git pull
```

Update release PR checklist: replace `- [ ] #NN` with `- [x] #NN`.

### 13. Loop

Log completion summary. Check milestone progress. Select next issue. Continue.

---

## Release Completion

When all milestone issues are closed:
1. Run final validation
2. Mark draft PR ready for review: `gh pr ready RELEASE_PR_NUMBER`
3. Do NOT auto-merge to main
4. Report: milestone complete, PR number, issue count

## Discovery Escape Hatch

If during implementation you discover missing functionality or blocking bugs:
1. Create new issues via `/create-issues` format
2. Add to milestone and dependency graph
3. If it blocks current work: commit progress, skip, pick up blocker next
4. If independent: continue current work, new issue picked up in future iteration

## Autonomy & Stopping

Fully autonomous. Stops when:
- All milestone issues closed (release complete)
- All remaining issues blocked or skipped
- Dependency cycle detected
- User intervenes

Individual failures are logged and skipped, not terminal.

## Rules

- ALWAYS follow architecture rules from CLAUDE.md
- ALWAYS use project's validation command as hard gate
- ALWAYS practice TDD — tests before implementation
- NEVER force-push or rewrite history
- NEVER skip validation
- One issue per feature branch
- Feature branches target release branch in release mode, main in ad-hoc mode
- Pre-existing test failures → create issue, don't silently ignore
