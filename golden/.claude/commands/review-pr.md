---
name: review-pr
description: Review a PR — 7-section standardized review with optional deep agent escalation
user_invocable: true
---

# /review-pr — Pull Request Review

Consistent, repeatable PR review. Same checks, same order, structured verdict.
Collaborative: the review is drafted with the user and posted only after the user approves the final draft.

## Invocation

```
/review-pr 42              # Standard 7-section review
/review-pr 42 --diff-only  # Skip build gates, code review only
/review-pr 42 --deep       # Escalate to parallel agent reviewers
/review-pr 42 --auto-post  # Skip sign-off (ad-hoc /wiggum per-issue review ONLY)
```

## Deep Review Escalation

Tier the review by diff size and surface. Count lines changed and files touched; check whether the diff hits any high-stakes surface (auth/middleware, schema/migration, anything labelled `security` / `migration` / `destructive` on the issue, code paths the project flags as privileged).

| Tier | Trigger | Action |
|---|---|---|
| **Standard 7-section** | <50 lines AND <3 files AND no high-stakes surface | Run the inline review below. No agents. |
| **Combined deep** | 50–300 lines, no high-stakes surface, no `--deep` flag | One reviewer agent with `[SPEC] [SECURITY] [OPS]` sections. Sequential. ~30–40% cheaper than parallel-three. |
| **Parallel three** | `--deep` passed, OR PR targets `main`/`release/*`, OR >300 lines, OR any high-stakes surface | Three parallel agents (reviewer + security-reviewer + ops-reviewer). Maximum rigor; parallel wall-clock. |

**Combined-tier prompt:**
> Use reviewer. Review PR #NN against the spec in the linked issue. Produce three sections in order — `[SPEC]` (acceptance criteria, architecture, holistic update), `[SECURITY]` (OWASP, secrets, injection, auth surface), `[OPS]` (logging, error wrapping, context plumbing, test coverage of failure modes). Each section gets a verdict (APPROVED/NEEDS_FIXES, SECURE/VULNERABLE, READY/NOT_READY). Do not skip sections under context pressure.

**Parallel-three prompts:**
> Use reviewer. Review PR #NN against the spec in the linked issue.
> Use security-reviewer. Full security scan of PR #NN.
> Use ops-reviewer. Observability audit of PR #NN.

For combined: any non-pass section blocks. For parallel-three: wait for all three; any non-pass agent blocks. If a combined-tier review surfaces concerning ambiguity (e.g. partial findings in a domain it didn't have time for), escalate to parallel-three rather than approving.

If none of the tier triggers fire, proceed with the standard 7-section review below.

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

Draft the structured review in the session (do NOT post yet — see Sign-off):

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

## Sign-off (collaborative — required before posting)

Unless `--auto-post`, the user owns the final review. Loop until the user approves or declines:

1. **Present the full draft** in the session — verdict, every section, action items.
2. **Ask what's next** (AskUserQuestion, or plain prompt if unavailable). Offer:
   - **Post as-is** — sign off on this exact draft
   - **Ask a question** — answer clarifying questions about findings, the diff, or reasoning
   - **Look into something** — investigate a file, path, edge case, or concern the user names
   - **Manual test** — run it (check out the branch, build, exercise the behavior) and report
     what was run and observed; if it needs the user (UI, credentials, hardware), give exact steps
     and wait for their result
   - **Edit** — change the verdict, add/drop/reword findings or action items
   - **Don't post** — end without posting anything
3. **Fold results into the draft.** New evidence can change PASS/WARN/FAIL and the verdict; apply
   Verdict Rules (never APPROVE with a FAIL — if the user insists, say so and offer COMMENT).
   Re-present the revised draft (or a clear diff of what changed) and return to step 2.
4. **Post only on explicit approval of the current draft** — a reply to a question is not approval:
   ```bash
   gh pr review NUMBER --approve|--request-changes|--comment --body-file REVIEW.md
   ```
   If GitHub rejects the verdict (e.g. approving your own PR), tell the user and offer `--comment`.
   Report the posted review URL.

`--auto-post` is reserved for `/wiggum` ad-hoc per-issue reviews: post the draft directly. Never use
it for a release-PR review or when the user invoked `/review-pr` themselves.

## Self-Improvement

If verdict is REQUEST_CHANGES or 2+ WARNs, AND PR was authored by Claude:
Run `/pomo` with findings as context. Skip for non-Claude PRs.

## Rules

- NEVER post a review without the user's explicit sign-off on the final draft (unless `--auto-post`)
- NEVER approve a PR with FAIL findings
- NEVER skip a section — every section gets evaluated
- ALWAYS show sections in the same order
- Be specific — reference file paths and line numbers
