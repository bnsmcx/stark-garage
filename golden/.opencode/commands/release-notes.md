---
name: release-notes
description: Generate a scannable, inverted-pyramid release PR description from closed milestone issues
user_invocable: true
---

# /release-notes — Release PR Description Generator

Generate a release PR description that puts **what changes for the frontend** at the top, then progressively less-actionable detail below it. A frontend dev should be able to read **only the first section** and know exactly what (if anything) they have to change in their client. Backend reviewers, ops, and QA find what they need by scrolling further. Implementation progress and follow-ups live below the fold.

Favor **bullets and tables**. The only prose allowed above the fold is one sentence summarizing the release theme. Everything else is structured.

## Invocation

```
/release-notes                    # Auto-detect release branch + milestone
/release-notes X.Y.Z              # Target a specific version
```

## Steps

### 1. Detect Release Context

```bash
git branch --show-current
gh pr list --base main --head RELEASE_BRANCH --state open --json number --jq '.[0].number'
gh issue list --milestone "MILESTONE" --state closed --limit 200 --json number,title,body,labels
```

### 2. Classify Every Closed Issue

For each issue, decide which bucket it falls into. **The bucket determines where it appears in the PR body.**

| Bucket | What goes here |
|---|---|
| **Frontend-visible API change** | Endpoint added, removed, response shape changed, new status code, new query param, new error code, auth requirement changed |
| **Frontend-visible behaviour change** | Same endpoints but different results (recall fix, ordering change, filter change) |
| **Breaking change** | Anything that requires a client code change to keep working |
| **Backend-only** | Observability, logging, lifecycle, internal refactors, RBAC slow path, etc. |
| **Dependencies / toolchain** | Language/runtime bump, dependency-manifest changes, container base, security patches |
| **Docs / infra** | CI tweaks, internal docs |

If an issue belongs in two buckets (e.g. a frontend-visible fix that also adds a new error code), put it in the higher-priority bucket and link to it from the lower one.

### 3. Write the Header (2 lines, max)

```markdown
# Release: X.Y.Z

**N issues · M PRs · K epics** — [one-sentence theme]
```

The one-sentence theme is the only prose above the fold. Don't repeat the bullet content in this sentence — say what the release *is about*, not what's in it. "Security hardening + recall-visibility fix" not "This release adds X, Y, Z, and also W."

### 4. Section 1 — Frontend Impact (always first, even if empty)

This is the section a frontend dev came here to read. If there is literally nothing for the frontend to do, say so explicitly — don't omit the section.

**Shape:**

```markdown
## What changes for the frontend

[If nothing changes:]
_No client-visible API changes this release._ Skip to the [backend changes](#backend-changes) if you're here for ops/infra context.

[Otherwise, in this order:]

### Breaking changes
[Bullet list, screaming loud. Skip section if none.]
- `POST /api/foo` — `bar` field renamed to `baz` in request body (#NN)

### New endpoints
| Method | Path | Auth | Purpose | Issue |
|---|---|---|---|---|
| `GET` | `/api/widgets/{id}/history` | Bearer + scope | Returns 30-day audit trail | #NN |

### Changed endpoints
| Endpoint | What changed | Frontend action | Issue |
|---|---|---|---|
| `GET /api/widgets` | Previously hidden item types now appear in results | None — recall expansion only | #NN |
| `POST /api/imports` | New `errorCode` field on error responses; new `408` and `504` status codes | Switch error UI to read `errorCode`; treat 504 as retryable | #NN |

### Behaviour changes (same endpoints, different results)
- `GET /api/projects` — list now sorted by `lastModified DESC` (was insertion order) — #NN
```

Each cell is one line — if you need more, link to the issue. The **Frontend action** column is mandatory on the Changed-endpoints table and is the most important content in the whole document.

### 5. Section 2 — Backend Changes

```markdown
## Backend changes
_(No frontend code action required.)_

**Observability**
- `rbac.approver_fallback{object_type, outcome}` counter on the RBAC slow path (#NN)
- Engine log lines now carry the middleware `requestId` for trace correlation (#NN)

**Lifecycle**
- `Handler.Close()` releases the connection pool on graceful shutdown (#NN)

**Internal / refactor**
- One bullet per item — skip the section if empty
```

Group by sub-bucket only if there are 3+ bullets total; otherwise a flat list is fine.

### 6. Section 3 — Dependencies + Toolchain

Always a table. Skip the section only if literally nothing moved.

```markdown
## Dependencies + toolchain

| Component | Before | After | Notes |
|---|---|---|---|
| Language toolchain | 1.25.9 | **1.25.10** | Clears 5 stdlib CVEs (#NN) |
| `golang.org/x/net` | v0.52.0 | **v0.53.0** | Auto-promoted by dependency tidy (#NN) |
| Framework dep | 2.7.18 | **3.5.14** | Namespace migration (#NN) |
```

### 7. Section 4 — Metrics Delta (only when measurable)

Only include when there's a real before/after number — vuln counts, p95 latency, error rate, test coverage. Skip for feature releases without measured outcomes.

```markdown
## Vulnerability delta (`image:latest`)

| Metric | Before | After | Δ |
|---|---|---|---|
| Critical | 3 | **0** | −3 (100%) |
| High | 22 | **0** | −22 (100%) |
| Total | 77 | 13 | −64 (−83%) |
```

### 8. Section 5 — Risk / Retest

Two-to-five bullets. Things a reviewer or QA would want to flag. Skip if the answer is genuinely "nothing surprising."

```markdown
## Risk / what to retest

- **Recall ≠ authorization:** #NN widens which rows the helper *finds*, not who can *see* them — auth gates unchanged but worth a sanity check against a restricted account
- **Namespace migration:** codebase had zero legacy-namespace imports pre-upgrade, so the major framework bump is a no-op for our code — a version test guards against resolver drift
- **No data migration:** all index/column changes are idempotent (`IF NOT EXISTS`)
```

### 9. Section 6 — Verification (one block, no prose)

```markdown
## Verification

- Project validation command ✓
- Per-PR reviews: **APPROVED**
- Cross-issue release review: **APPROVED** (reviewer + security-reviewer + ops-reviewer)
```

### 10. Below the Fold — Implementation Progress

Always wrap in `<details>`. One line per issue, grouped by phase or bucket if helpful.

```markdown
<details>
<summary><strong>Implementation progress (N/N complete)</strong></summary>

**Security / dependencies**
- [x] #NN — language toolchain bump (PR #MM)
- [x] #NN — framework major-version bump (PR #MM)

**Regression fix**
- [x] #NN — recall-visibility fix (PR #MM)

**Observability / lifecycle**
- [x] #NN — propagate requestId into engine logs (PR #MM)
- [x] #NN — Handler.Close on graceful shutdown (PR #MM)

</details>
```

### 11. Below the Fold — Follow-ups

```markdown
<details>
<summary><strong>Follow-ups (non-blocking)</strong></summary>

- #NN — migrate base image to hardened variant (blocked on entitlement)
- #NN — batch the recall lookups for list endpoints

</details>
```

Skip entirely if there are none — don't leave an empty "Follow-ups" section.

### 12. Assemble Final PR Body

Target shape — everything from the header through Verification should fit on **one laptop screen** above the `<details>` folds:

```markdown
# Release: X.Y.Z

**N issues · M PRs · K epics** — [one-sentence theme]

## What changes for the frontend
[Breaking → New → Changed → Behaviour, each subsection omitted if empty.
 If the entire section is empty, say so in one line and link to "Backend changes" below.]

## Backend changes
[Bulleted, grouped if 3+ items.]

## Dependencies + toolchain
[Table.]

## Vulnerability delta
[Only for security-themed releases. Skip otherwise.]

## Risk / what to retest
[2–5 bullets.]

## Verification
- bullets

---

<details>
<summary><strong>Implementation progress (N/N complete)</strong></summary>
...
</details>

<details>
<summary><strong>Follow-ups (non-blocking)</strong></summary>
...
</details>
```

### 13. Apply to PR

```bash
gh pr edit PR_NUMBER --body-file /tmp/release-notes.md
```

## Rules

- **Inverted pyramid.** Frontend impact comes first, always. Backend, deps, risk, verification follow in that order. Implementation progress and follow-ups live in `<details>` below the fold.
- **Frontend section is never silently omitted.** If there are no client-visible changes, say so in one line. Don't make the reader infer it.
- **One line per bullet, one line per table cell.** If a cell needs more, link to the issue and let the issue carry the detail.
- **Tables beat bullet lists for ≥3 items with consistent shape.** Endpoints, deps, metrics — always tables. Behaviour changes can be bullets if they don't share a shape.
- **The Changed-endpoints table always carries a `Frontend action` column.** That column is the most-read content in the entire PR body — fill it with imperatives ("switch error UI to read `errorCode`"), not descriptions.
- **Prose budget.** One sentence in the theme line. Everything else is structured. No narrative "Why now" paragraph — if motivation matters, link to the source ticket from the relevant bullet.
- **Use horizontal rules to separate above-fold from below-fold.** Visual whitespace is part of the readability budget.
- **Skip empty sections entirely.** No "Database Changes: N/A" placeholders. Drop the heading if it doesn't apply (except the Frontend section, which always appears).
- **No marketing prose.** "Previously hidden item types now appear on `GET /api/widgets`" is fine. "Empowers users with comprehensive visibility…" is not.
- ALWAYS count PRs, issues, and epics accurately from the milestone.
- NEVER use bare issue numbers without context — each `#NN` reference appears alongside a one-line description of what shipped.
- NEVER describe internal implementation in the Frontend section — that section is strictly about the API surface the client sees.
