---
name: triage
description: Analyze the issue backlog — dependency graph, readiness, label validation, and prioritization
user_invocable: true
---

# /triage — Backlog Analysis

Analyze all open issues, build a dependency graph, and produce an actionable summary.

## Startup

Query memory for relevant lessons before beginning:
```bash
toolbox-memory search --ns lesson --query "triage dependency blocked"
```
Apply any returned lessons to the analysis below.

## Task Tracking Mode

When CLAUDE.md defines a Task Tracker section using `tasks/todo.md`:
- **Step 1:** Read `tasks/todo.md` instead of fetching from an external tracker
- **Step 2:** Parse `- Blocked by: T-NN — reason` dependency format
- **Steps 3-4:** Same logic, using `T-NN` references
- **Step 5:** Skip label validation — labels are not available in todo.md mode
- **Step 8:** Skip label fix offers

## Steps

### 1. Fetch all open issues

Fetch all open issues using the **list open issues** operation (see `agent_docs/issue-tracker-ops.md`).

Store the full result for processing.

### 2. Parse dependency references

Scan each issue body for the **canonical** dependency format:

```
- Blocked by: #NN — reason
```

Rules:
- Only recognize lines matching the canonical format (with or without ` — reason` suffix)
- Do NOT parse other patterns like "depends on #NN", "waiting on #NN", etc.
- Build adjacency list: `blockGraph[NN] = [list of issues NN blocks]`
- Build reverse: `blockedBy[NN] = [list of issues blocking NN]`

### 3. Build dependency graph and detect cycles

Perform a topological sort on the dependency graph:
- If the sort succeeds, store the ordering
- If it fails, identify and report the cycle(s) — list the issue numbers involved

Algorithm: Kahn's algorithm or DFS-based. Track in-degrees for each node, repeatedly
remove zero-in-degree nodes. Remaining nodes after exhaustion form cycle(s).

### 4. Classify issues

For each open issue, determine:
- **Ready**: No open blockers (all `blockedBy` issues are closed)
- **Blocked**: Has at least one open blocker
- **Impact score**: Count how many issues (transitively) depend on this one.
  Walk the `blockGraph` recursively. Higher score = more impactful to complete.

### 5. Validate labels

Check each issue for consistency:
- Issues with `blocked` label but NO open blockers — flag as "stale blocked label"
- Issues WITHOUT `blocked` label but WITH open blockers — flag as "missing blocked label"
- Issues missing acceptance criteria (no `- [ ]` checkboxes in body)
- Issues using non-standard dependency formats (anything that looks like a dependency
  reference but doesn't match the canonical format)

### 6. Group by category

Group ready issues by project labels and title scopes. If the project defines layers
or scopes in CLAUDE.md, use those. Otherwise fall back to defaults:
- **Feature** (`enhancement` label or title contains `feat(`)
- **Bug** (`bug` label or title contains `fix(`)
- **Documentation** (`documentation` label or title contains `docs`)
- **Infrastructure** (`infra` label or title contains `infra(`)
- **Discovery/Design** (`discovery` or `design` label)
- **Tracking** (`tracking` label)
- **Cross-cutting** (none of the above)

### 7. Output summary

Present a structured report:

```
## Backlog Triage Summary

**Total open:** NN | **Ready:** NN | **Blocked:** NN | **Missing criteria:** NN

### Highest-Impact Issues (unblock the most work)
| # | Title | Impact | Labels |
|---|-------|--------|--------|
| ... |

### Ready Issues by Category
#### Feature (N ready)
- #NN — title
#### Bug (N ready)
- #NN — title
[...etc]

### Blocked Issues
| # | Title | Blocked by |
|---|-------|------------|
| ... |

### Label Issues
- #NN: has `blocked` label but all blockers are closed
- #NN: missing `blocked` label (blocked by #MM which is open)
- #NN: missing acceptance criteria

### Dependency Cycles (if any)
- Cycle: #A -> #B -> #C -> #A
```

### 8. Offer label fixes (optional)

If label inconsistencies were found, ask the user:
> "Found N label issues. Apply fixes? (This will add/remove `blocked` labels as needed)"

If confirmed, apply fixes using the **add label** and **remove label** operations
(see `agent_docs/issue-tracker-ops.md`).

## Notes

- This command is **read-only by default** — it only modifies labels if the user confirms
- Other commands (`/create-issues`, `/wiggum`) reuse the dependency graph logic
- The canonical dependency format (`- Blocked by: #NN — reason`) is the only recognized format
