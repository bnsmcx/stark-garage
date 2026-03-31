---
name: review-pr
description: Review a PR — 7-section standardized review with optional deep agent escalation
user_invocable: true
---

# /review-pr — Pull Request Review

Consistent, repeatable PR review. Same checks, same order, structured verdict.

## Invocation

```
/review-pr 42              # Standard 7-section review
/review-pr 42 --diff-only  # Skip build gates, code review only
/review-pr 42 --deep       # Escalate to parallel agent reviewers
```

## Deep Review Escalation

When `--deep` is passed, OR the PR targets `main`/`release/*`, OR the diff touches auth/security middleware or database migrations:

Launch 3 parallel review agents:
> Use reviewer. Review PR #NN against the spec in the linked issue.
> Use security-reviewer. Full security scan of PR #NN.
> Use ops-reviewer. Observability audit of PR #NN.

Wait for all 3 to complete. Aggregate verdicts — all must pass for APPROVE.

Otherwise, proceed with the standard 7-section review below.

---

## Review Sections

### 1. PR Metadata

```bash
gh pr view NUMBER --json number,title,body,baseRefName,headRefName,files,additions,deletions
gh pr diff NUMBER
```

- [ ] Title follows conventional format: `type(scope): description`
- [ ] Description is filled in (not boilerplate)
- [ ] Linked issue exists (smart close syntax) — WARN if missing
- [ ] Base branch correct (release branch if exists, else `main`)

### 2. Architecture Compliance

Scan diff for CLAUDE.md architecture rule violations (**FAIL** findings):
- Layer boundary violations
- Import restriction violations
- Data access pattern bypasses
- New abstractions not wired/registered

### 3. Holistic Update Check

If shared types/interfaces/contracts changed, verify all consuming layers updated:
- Type definitions, implementations, wiring, consumers, tests
- WARN if a layer is missing

### 4. Code Quality

- Error handling at boundaries
- No hardcoded values that should be config
- Proper types (no unjustified `any`)
- No `.env` files in diff
- No `console.log` in production code (WARN)

### 5. Test Coverage

- New functionality has tests (WARN if missing)
- Bug fixes include regression tests (WARN if missing)
- Test files in correct location
- Pure refactors: tests optional

### 6. Security

- No committed secrets/credentials
- No injection vulnerabilities
- External inputs validated
- Sensitive data handled appropriately

### 7. Build Gates

Unless `--diff-only`, run the project's validation command (from CLAUDE.md):
- Passes → PASS
- Fails → FAIL (show error output)

---

## Output Format

Post structured review on the PR:

```markdown
## PR Review: #NN — title

### Verdict: {APPROVE | REQUEST_CHANGES | COMMENT}

### Summary
[2-3 sentence assessment]

### Findings

#### {PASS|WARN|FAIL} 1. PR Metadata
[findings]

#### {PASS|WARN|FAIL} 2. Architecture Compliance
[findings or "No violations"]

#### {PASS|WARN|FAIL} 3. Holistic Update Check
[findings or "N/A"]

#### {PASS|WARN|FAIL} 4. Code Quality
[findings]

#### {PASS|WARN|FAIL} 5. Test Coverage
[findings]

#### {PASS|WARN|FAIL} 6. Security
[findings or "No issues"]

#### {PASS|WARN|FAIL} 7. Build Gates
[PASS|FAIL or "Skipped (--diff-only)"]

### Action Items
[numbered list of required fixes, if any]
```

## Verdict Rules

- **APPROVE**: Zero FAILs, at most minor WARNs
- **REQUEST_CHANGES**: Any FAIL, or multiple significant WARNs
- **COMMENT**: No FAILs but notable WARNs

## Self-Improvement

If verdict is REQUEST_CHANGES or 2+ WARNs, AND PR was authored by Claude:
Run `/pomo` with findings as context. Skip for non-Claude PRs.

## Rules

- NEVER approve a PR with FAIL findings
- NEVER skip a section — every section gets evaluated
- ALWAYS show sections in the same order
- Be specific — reference file paths and line numbers
