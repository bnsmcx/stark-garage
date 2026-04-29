---
name: close-issue
description: Validate acceptance criteria, close issue, unblock downstream, write to memory
user_invocable: true
---

# /close-issue — Issue Validation & Closure

Quality gate at the end of implementation. Validates acceptance criteria before closing.

## Invocation

```
/close-issue 53         # Close issue #53
/close-issue 53 54 55   # Close multiple issues
```

## Task Tracking Mode

When CLAUDE.md defines `tasks/todo.md`:
- Use `T-NN` references
- Move row from Active to Done table with completion date
- Check downstream `Blocked by: T-NN` references

## Steps

### 0. Mark Start

```bash
mkdir -p .claude/usage && printf '{"event":"iter_start","command":"close-issue","issue":NUMBER,"ts":"%s"}\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> .claude/usage/iterations.jsonl
```

### 1. Fetch Issue

Fetch via `gh issue view NUMBER --json number,title,body,labels,milestone,assignees`.

Parse body for: summary, dependencies, acceptance criteria (`- [ ]` items), implementation notes.

### 2. Validate

Run project's validation command (CLAUDE.md) — **hard gate**. If fails, stop immediately.

Then validate each acceptance criterion:

**Automated:** File existence, project-specific checks.
**Code-verifiable:** Module exports, component structure, service methods — inspect implementation.
**Manual/judgment:** Interactive: ask user. Autonomous (called by /wiggum): verify by code inspection. If unverifiable without human, mark SKIP.

Track each as PASS, FAIL, or SKIP.

### 3. Gate

If validation failed or ANY criterion is FAIL:
```
CLOSE_FAILED:
- Validation: PASS/FAIL
- "Criterion text": FAIL — reason
- "Criterion text": PASS
```
When called by `/wiggum`, this signals fix-and-retry.

If all PASS (with optional SKIPs), proceed.

### 4. Check Off Criteria

Update issue body: replace `- [ ]` with `- [x]` for passing criteria.
```bash
gh issue view NUMBER --json body --jq '.body'
gh issue edit NUMBER --body "UPDATED_BODY"
```

### 5. Architecture Change Detection

If files matching architecture-sensitive patterns changed (routes, middleware, schema, migrations) but docs were not updated, add a WARN to the closing comment.

### 6. Closing Comment

```bash
gh issue comment NUMBER --body "$(cat <<'EOF'
## Closed

### Summary
[what was implemented]

### Changes
- [key files changed]

### Acceptance Criteria
- [x] Criterion 1 — PASS
- [x] Criterion 2 — PASS
- [ ] Criterion 3 — SKIP (requires human verification)

### Verification
- Validation: PASS
EOF
)"
```

### 7. Close Issue

```bash
gh issue close NUMBER
```

Interactive: ask confirmation first. Autonomous (called by /wiggum): proceed.

### 8. Downstream Impact

Find all open issues with `- Blocked by: #NUMBER` in their body:
1. For each, check if ALL their blockers are now closed
2. If yes: remove `blocked` label via `gh issue edit NUMBER --remove-label "blocked"`
3. Report which issues were unblocked

### 9. Milestone Progress

Report: "Milestone: X/Y closed (Z%)"

### 10. Mark End

```bash
printf '{"event":"iter_end","command":"close-issue","issue":NUMBER,"ts":"%s"}\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> .claude/usage/iterations.jsonl
```

## Rules

- NEVER close an issue with FAIL criteria
- ALWAYS run validation command first (hard gate)
- ALWAYS post structured closing comment
- ALWAYS check downstream for newly-unblocked issues
