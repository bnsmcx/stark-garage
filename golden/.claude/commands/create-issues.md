---
name: create-issues
description: Create issues — single or batch. Auto-detects mode from input.
user_invocable: true
---

# /create-issues — Issue Creation

Creates issues in work-compatible format. Auto-detects mode:
- **Single mode:** One bug description or feature request = one issue
- **Batch mode:** A numbered plan or multi-step description = tracking epic + ordered children

## Invocation

```
/create-issues                    # Unassigned
/create-issues me                 # Assign to authenticated user
/create-issues ben                # Resolve to collaborator username
```

## Task Tracking Mode

When CLAUDE.md defines a Task Tracker section using `tasks/todo.md`:
- Skip assignee resolution
- Use `T-NN` IDs instead of `#NN`
- Add rows to Active table in `tasks/todo.md`

## Mode Detection

Look at the user's input and conversation context:
- **Single mode:** A single bug report, feature request, error description, or PR feedback
- **Batch mode:** A numbered list of steps, a multi-step plan, bullet-point breakdown with dependencies, or output from `/investigate`

If ambiguous, ask: "Is this one issue or a plan with multiple issues?"

---

## Single Mode

### 1. Gather Details

From the provided context, determine:
- **Title** — `{type}({scope}): {description}` (see `agent_docs/issue-conventions.md`)
- **Summary** — what and why
- **Dependencies** — which open issues must complete first
- **Acceptance Criteria** — testable checkboxes; always include "Validation passes"

### 2. Create the Issue

```bash
gh issue create --title "TITLE" --body "$(cat <<'EOF'
## Summary
[description]

## Dependencies
[- Blocked by: #NN — reason, or "None"]

## Acceptance Criteria
- [ ] [specific criterion]
- [ ] Validation passes

## Implementation Notes
[key files, approach, constraints]
EOF
)" --label "LABELS" [--assignee "$ASSIGNEE"]
```

### 3. Report

Output the issue URL and whether it is unblocked or waiting.

---

## Batch Mode

### 0. Resolve Assignee

If assignee argument provided:
1. `me` — resolve via `gh api user --jq '.login'`
2. Other name — match against `gh api repos/{owner}/{repo}/collaborators --jq '.[].login'`
3. No argument — leave unassigned

### 1. Extract Plan

Look back through conversation for the plan. Extract:
- Epic title (overall goal)
- Steps (each discrete unit of work)
- Implementation details
- Dependencies and ordering

If no plan found: "I don't see a plan in our conversation. Could you describe what issues you'd like to create?"

### 2. Survey Existing Issues

Fetch all open issues. Build set of titles to detect duplicates.

### 3. Draft Tracking Epic

**Title:** `tracking: {epic title}`

**Body:**
```markdown
## Summary
{1-3 sentences}

## Issues

| Order | Issue | Status |
|-------|-------|--------|
| 1 | {title} | :white_circle: |
| 2 | {title} | :white_circle: |

## Acceptance Criteria
- [ ] All child issues closed
- [ ] Validation passes
```

### 4. Draft Child Issues

For each step, one issue in canonical format (see `agent_docs/issue-conventions.md`):

```markdown
## Summary
[1-3 sentences from plan]

## Dependencies
- Blocked by: #NN — [reason]
- Part of: #EPIC — [epic title]

## Acceptance Criteria
- [ ] [specific, testable criterion]
- [ ] Validation passes

## Implementation Notes
[key files, approach]
```

Apply `blocked` label if issue has open dependencies.

### 5. Validate Dependency Graph

Build combined graph (existing + proposed issues):
- Run topological sort for cycle detection
- Verify all `#NN` references exist
- Verify sequence matches plan

### 6. Present for Review

Show epic + all issues in a summary table with dependency graph. Ask: "Create these N issues?"

### 7. Create Issues

Create in dependency order (blockers first):
1. Create epic first
2. Create children, updating `#NN` references with real numbers after each creation
3. Update epic with real issue numbers
4. Apply `blocked` label to issues with open dependencies

### 8. Post-Creation Validation

Rebuild dependency graph, confirm no broken references. Show summary tree.

## Rules

- NEVER create issues without user confirmation
- NEVER create duplicates — flag similar existing issues
- ALWAYS include "Validation passes" in acceptance criteria
- ALWAYS create tracking epic when creating 2+ related issues
- Dependencies MUST use canonical format: `- Blocked by: #NN — reason`
- One issue per discrete piece of work
