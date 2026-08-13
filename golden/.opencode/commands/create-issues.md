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

## Quality Bar — What an Issue Body Must Deliver

An issue body is a handoff contract to the implementer. The implementer should be able to open the issue cold — no prior conversation, no chat history — and know what to do without a second research pass. That means every body, single or batch, clears this bar:

- **Every claim anchored.** `file:line`, `function name`, or `selector` — not "the code", not "that file".
- **Every line number verified against current source** before the issue ships. Numbers rot fast across branches.
- **Diffs shown as code, not described in prose.** Before/after blocks when the change is non-trivial; concrete test templates, not "add tests".
- **What stays the same is explicit.** Spelling out non-changes prevents scope creep.
- **Dependencies and sequencing are load-bearing.** "Blocks X because…" and "Blocked by Y so that…" — the *why* matters for edge cases.
- **Acceptance criteria are grouped by concern** (markup / behavior / tests / verification / shipping), not a flat list of checkboxes.
- **Risks have specific mitigations.** "Unlikely to regress X because Y" beats "risk: regression".
- **Out-of-scope is named.** Especially for features where the design space is open.

Trivial one-liners (e.g. typo fix, unused import) can scale down — but even then, all four sections are present (Summary, Dependencies, Acceptance Criteria, Implementation Notes) with anchored claims.

---

## Research First (Both Modes)

Before drafting any issue body, for each proposed issue:

1. **Read the referenced source files.** Open `file:line` targets and verify the current state matches what the plan assumes. Branches drift; line numbers are the first thing to rot.
2. **Read adjacent tests.** Existing test file structure dictates the form of any new tests you prescribe. If tests use `vitest` with `describe`/`it`, don't propose `test.each`. If Go tests use a project-specific test-db helper, reference it by name.
3. **Check dependency reality.** If the body says "blocked by #221", verify #221 exists and that the blocker relationship is real (not stale from an earlier conversation).
4. **Confirm the bug exists (for fixes).** Read the code path end-to-end — the reported symptom and the suspected cause should both check out. If the bug turns out to already be fixed, flag it before creating the issue.
5. **Scan for drift.** Plans from 2+ conversations ago can reference renamed files, removed functions, or merged branches. Spot-check a handful of claims.

This step is the difference between an issue body the implementer can act on and one they have to re-research. Skipping it means shipping wrong line numbers and vague file refs.

Budget: a few minutes of `Read` + `Grep` per issue. Always worth it.

---

## Canonical Format

Every issue body, single or child of a batch, uses this structure:

```markdown
## Summary

[One tight paragraph. What the problem/feature is, where it lives (file:line),
why it matters (user impact or engineering value), and the high-level approach.
Not a sentence fragment. Aim for 3-6 sentences.]

## Dependencies

- Blocked by: #NN — [why, e.g. "must extract helper first"]
- Blocks: #NN — [why, e.g. "downstream issue inherits new markup"]
- Part of: #EPIC — [epic title]

(Use "None" if genuinely unblocked.)

## Acceptance Criteria

**[Group heading — code change / markup / etc.]**
- [ ] [Specific, testable criterion with file:line anchor]
- [ ] [Another]

**[Group heading — behavior]**
- [ ] [What the user or downstream system sees]

**[Group heading — tests]**
- [ ] [Specific new test name + file + key assertions]
- [ ] [Existing tests that must continue to pass]

**[Group heading — manual verification]**
- [ ] [URL / flow / screen to check + expected behavior]

**Shipping checks**
- [ ] Validation passes

## Implementation Notes

### [Root cause / Design rationale]

[For fixes: why the bug exists and why the proposed fix is safe.
For features: the design decision and what alternatives were ruled out.]

### Exact changes

[Before/after code blocks for non-trivial diffs. Target markup or helper
signatures when adding new code. File:line anchors for every block.]

### Test strategy

[Concrete test template — actual code, not vague guidance. Describe the
helpers to extract if needed to make the code testable.]

### What stays the same

[Explicit list of adjacent code that must not change. Prevents scope creep
and gives the reviewer a checklist.]

### Risks & mitigations

- [Risk] — [Specific mitigation or "unlikely because…"]

### Out of scope

- [Related work that could be tempting but belongs in a separate issue]
- [Design options explicitly rejected for this pass]
```

### Section-by-section guidance

**Summary.** Don't just restate the title. Set up the problem, cite the evidence (file:line, corpus stat, error signature), name the approach. The reader should know by the end of the paragraph whether this is a one-line fix or a refactor.

**Acceptance Criteria groupings.** Pick the groupings that fit the issue. Typical groupings:
- **Code change** or **Markup** — the mechanical diff
- **Behavior** — what the user / downstream system observes
- **Tests** — regression coverage + new assertions
- **Manual verification** — URL matrix, click-path, visual check
- **Shipping checks** — the mandatory `Validation passes`

One-to-two-line fixes might only need **Code change** + **Tests** + **Shipping**.

**Implementation Notes subsections.** Use the ones that help, skip the ones that don't:
- **Root cause** (for bugs) — why the bug exists; evidence it's a bug vs. intended behavior
- **Design rationale** (for features) — key decisions + alternatives considered
- **Exact changes** — before/after code blocks with file:line
- **Test strategy** — concrete templates, extract-helper pattern if needed
- **What stays the same** — explicit non-changes
- **Risks & mitigations** — each risk gets a mitigation or a reason it's unlikely
- **Out of scope** — named boundaries
- **Sequencing** — when ordering with other issues matters, spell out why

Lean into tables when they help (verification matrices, retention checklists, cost breakdowns). A 3×4 table often beats three paragraphs.

**Anchoring.** Every non-trivial claim carries an anchor: `file:line`, `function name`, `selector`, `route`. "The handler" is not anchored; `drawLabels at routes.js:549` is.

---

## Single Mode

### 1. Research (see above)

Read the referenced source, verify line numbers, check adjacent tests, confirm the bug exists.

### 2. Draft the body

Use the canonical format. Pick the subsections that fit. For a trivial one-liner, you may have only 2-3 acceptance-criteria groupings and 2-3 Implementation Notes subsections — but all four top-level sections are always present.

### 3. Create the Issue

```bash
gh issue create --title "TITLE" --body-file /tmp/issue_body.md --label "LABELS" [--assignee "$ASSIGNEE"]
```

Prefer `--body-file` over inline heredocs for non-trivial bodies — the gh CLI handles multi-line content more reliably.

### 4. Report

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

Fetch all open issues. Build set of titles to detect duplicates. If any proposed issue overlaps meaningfully with an existing open issue, surface it and ask: "Create a new issue, update the existing #NN in place, or close one?"

### 3. Research — All Referenced Code Up Front

**Before drafting any bodies**, batch the research for every proposed issue in parallel:
- `Read` every file+line the plan references
- `Grep` for related identifiers (selectors, function names, constants)
- Read adjacent test files for each affected module

This is faster as one parallel research pass than interleaved with drafting, and surfaces drift across the whole plan before you commit prose.

### 4. Draft Tracking Epic

**Title:** `tracking: {epic title}`

**Body:**
```markdown
## Summary
{2-4 sentences on the overall goal + why these children belong together}

## Issues

| Order | Issue | Status |
|-------|-------|--------|
| 1 | {title} | :white_circle: |
| 2 | {title} | :white_circle: |

## Dependency Graph

```
#NN (unblocked)
  └─ #NN (blocked by the above)
#NN (unblocked, parallel)
```

## Acceptance Criteria

- [ ] All child issues closed
- [ ] Manual verification: {concrete, multi-point check across the child set}
- [ ] Validation passes

## Notes
{Cross-references to related closed epics, deferred backlog items, explicit
non-goals for this epic}
```

### 5. Draft Child Issues

For each step, one issue body in the canonical format (see Quality Bar + Canonical Format sections above). Every child body must clear the quality bar — anchored claims, grouped acceptance criteria, structured implementation notes, tests and risks spelled out.

Apply `blocked` label if the issue has open dependencies.

### 6. Validate Dependency Graph

Build combined graph (existing + proposed issues):
- Run topological sort for cycle detection
- Verify all `#NN` references exist
- Verify sequence matches plan

### 7. Present for Review

Show epic + all issues in a summary table with dependency graph. For non-trivial batches, include a word-count or section-count per issue so the user can eyeball depth. Ask: "Create these N issues?"

### 8. Create Issues

Create in dependency order (blockers first):
1. Create epic first
2. Create children, updating `#NN` references with real numbers after each creation
3. Update epic with real issue numbers
4. Apply `blocked` label to issues with open dependencies

Prefer `--body-file /tmp/issue_NN_body.md` over inline heredocs for reliability on long, structured bodies.

### 9. Post-Creation Validation

Rebuild dependency graph, confirm no broken references. Show summary tree.

---

## Rules

- NEVER create issues without user confirmation
- NEVER create duplicates — flag similar existing issues and ask
- NEVER skip the research step, even for "obvious" issues — line numbers rot
- ALWAYS include "Validation passes" in acceptance criteria
- ALWAYS create tracking epic when creating 2+ related issues
- ALWAYS anchor claims to `file:line`, `function`, or `selector`
- ALWAYS show diffs as code blocks when the change is non-trivial
- ALWAYS name what's out of scope — especially for features
- Dependencies MUST use canonical format: `- Blocked by: #NN — reason`
- One issue per discrete piece of work
- Prefer `--body-file` over inline heredocs for multi-section bodies
