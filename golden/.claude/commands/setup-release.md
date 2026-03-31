---
name: setup-release
description: Plan a release — blast-radius analysis, index the codebase, scope issues, create milestone, branch, and phased implementation plan
user_invocable: true
---

# /setup-release — Release Planning

Scope a release by filtering issues, analyzing blast radius, refreshing the codebase index, creating a milestone, setting up a release branch, and producing a phased implementation plan with scope analysis.

## Task Tracking Mode

This command is **not applicable** in todo.md mode. It requires milestone support, which is only available with external issue trackers.

## Invocation

```
/setup-release                                  # Interactive — asks what to include
/setup-release bugs                             # Filter: bug issues
/setup-release enhancement                      # Filter: feature/enhancement issues
/setup-release enhancement 10-25                # Filter + specific issue range
```

## Steps

### 1. Parse input

Extract filters from user input:
- **Label filters**: map keywords to issue labels. Common mappings:
  - `bugs` / `bug` — `bug`
  - `features` / `feat` — `enhancement`
  - `docs` — `documentation`
  - For project-specific labels, check CLAUDE.md
- **Issue range**: if numbers are provided (e.g., `10-25`), include those specific issues regardless of labels
- If no input, ask the user what they want to release

### 2. Query matching issues

Fetch matching issues using the **list open issues** operation (see `agent_docs/issue-tracker-ops.md`), filtering by the relevant labels.

If an issue range was specified, also fetch those individually using the **view issue** operation.

Combine results, deduplicate by issue number.

### 3. Blast Radius Analysis

For each issue in the release set, run `/blast-radius` to map cross-package impact:

1. Read each issue body and extract target files, types, or functions mentioned in implementation notes
2. For each target, invoke `/blast-radius` (or perform the equivalent analysis inline)
3. Collect all blast radius reports

Aggregate results into a scope summary:
- **Total files touched**: deduplicated set of all files referenced across all blast-radius reports
- **Packages touched**: unique packages/modules affected
- **Max dependency depth**: deepest call chain found
- **Risk distribution**: count of CONTAINED / MODERATE / WIDE assessments
- **Hotspots**: files that appear in multiple blast-radius reports (high-change areas)

### 4. Indexer Invocation

Invoke the Indexer agent to build or refresh the codebase state file so Planner has fresh context:

```
Use indexer. Index this project.
```

Wait for the index to complete before proceeding to dependency analysis. If the Indexer is not available, note this and proceed — the index step is recommended but not blocking.

### 5. Dependency analysis

Run the triage dependency analysis on the filtered set:
- Parse `- Blocked by: #NN` from each issue body
- Build the dependency graph for this subset
- Identify **external blockers**: open issues NOT in this release set that block issues IN the set

### 6. Report external blockers

If external blockers exist, present them:
```
## External Blockers
These open issues are NOT in this release but block issues that are:

| Blocker | Blocks (in release) | Action needed |
|---------|---------------------|---------------|
| #5 — feat(scope): Feature X | #12, #13 | Include in release OR complete first |
```

Ask the user:
- "Include these blockers in the release?" — adds them to the set
- "They'll be done before we start" — proceed without them
- "Remove the blocked issues from the release" — exclude them

### 7. Create milestone

Generate the milestone name: `v{version}` or `release/YYYY-MM-{scope}`
- scope = primary label filter (e.g., `bugs`, `features`, `full`)
- Example: `v1.0` or `release/2026-03-full`

Create using the **create milestone** operation (see `agent_docs/issue-tracker-ops.md`).

Assign all release issues to the milestone using the **assign to milestone** operation.

### 8. Generate implementation order

Topological sort of the release issues, grouped into phases:

```
## Implementation Plan

### Phase 1: Foundation (no dependencies)
1. #10 — feat(scope): Define shared types and constants
2. #11 — feat(scope): Implement data layer

### Phase 2: Core services (depend on Phase 1)
3. #12 — feat(scope): Service A (blocked by #10)
4. #13 — feat(scope): Service B (blocked by #11)

### Phase 3: Integration (depend on Phase 2)
5. #18 — feat(scope): Wire up services (blocked by #12)

### Phase 4: Features (depend on Phase 3)
6. #22 — feat(scope): Feature X (blocked by #18)
```

### 9. Create release branch

```bash
git checkout main
git pull origin main
git checkout -b release/v1.0
git push -u origin release/v1.0
```

Tell the user: "Release branch `release/v1.0` created from `main`."

### 10. Create draft release PR

Create a draft PR from the release branch to `main`:

```bash
gh pr create \
  --base main \
  --head release/RELEASE_NAME \
  --draft \
  --title "Release: RELEASE_NAME" \
  --body "FILLED_IN_BODY"
```

Fill in the body with the scope analysis and phased implementation plan:

```markdown
## Description
Release: `release/v1.0`

## Scope Analysis
- Issues: N
- Estimated effort: ~X hours
- Packages touched: [list from blast radius aggregation]
- Max dependency depth: N
- Cross-package impact: [CONTAINED / MODERATE / WIDE — based on risk distribution]
- Hotspots: [files appearing in multiple blast-radius reports]
- Execution mode: Release (full agent pipeline)

## Implementation Progress
### Phase 1: Foundation
- [ ] #10 — feat(scope): Define shared types and constants
- [ ] #11 — feat(scope): Implement data layer

### Phase 2: Core services
- [ ] #12 — feat(scope): Service A (blocked by #10)
[...all phases from Step 8]

*Updated automatically by /wiggum as issues are completed.*
```

Note the PR number returned — report it in the summary.

### 11. Summary

```
## Release Setup Complete

Milestone: v1.0
Branch: release/v1.0
Issues: 12 total (8 ready, 4 blocked)
Release PR: #30 (draft)

Scope Analysis:
  Packages touched: [list]
  Max dependency depth: N
  Cross-package impact: CONTAINED / MODERATE / WIDE
  Hotspots: [files appearing in 3+ blast-radius reports]

Implementation order:
  Phase 1: #10, #11 (ready now)
  Phase 2: #12, #13 (after Phase 1)
  Phase 3: #18 (after Phase 2)
  Phase 4: #22, #23, #24 (after Phase 3)

Next step: Run /wiggum to start implementing. Track progress on PR #30.
```

## Rules

- NEVER create a milestone that already exists — check first using the **list milestones** operation
- NEVER force-push the release branch
- ALWAYS run blast-radius analysis before creating the implementation plan
- ALWAYS invoke the Indexer before dependency analysis when available
- ALWAYS ask user confirmation before creating the milestone and branch
- External blockers must be resolved (included, pre-completed, or dependents excluded) before proceeding
- The release branch is created from the latest `main` — always pull first
