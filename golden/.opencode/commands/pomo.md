---
name: pomo
description: Post-mortem reflection — capture lessons and write to memory
user_invocable: true
---

# /pomo — Post-Mortem

Reflect on a recently resolved issue, capture generalizable lessons, and persist them to the
harness-native memory system (`MEMORY.md` + one file per fact).

## Invocation

```
/pomo              # Reflect on what just happened
/pomo <context>    # Reflect on a specific issue or PR
```

Also auto-invoked by:
- `/wiggum` after 2+ retry attempts
- `/review-pr` after REQUEST_CHANGES on Claude-authored PRs

## Process

### 1. Reconstruct the Incident

From conversation context, identify:
1. **Symptom** — what was observed
2. **Root cause** — the actual defect
3. **Cause chain** — how root cause produced symptom
4. **Fix** — what changed and why

Write 3-5 sentence summary. Confirm with user (if interactive).

### 2. Evaluate Worth

Not every fix needs a lesson. Decision tree:
- Could a reasonable developer make this same mistake in new code? → YES
- Did the bug involve compounding failures? → YES
- Did existing docs/conventions fail to prevent this? → YES
- Simple typo, obvious from error message, or one-off config issue? → SKIP

If skipping, tell user why and stop.

### 3. Check for Duplicates

Scan `MEMORY.md` (and the auto-recalled memories) for an existing fact covering this pattern.

If a match is found:
- **Same pattern:** update the existing memory file — sharpen the fact and note the recurrence.
- **Related but distinct:** create a new file and cross-link with `[[name]]`.

### 3b. Prune

While reviewing memory, delete any facts that have turned out to be wrong or obsolete, and remove
their `MEMORY.md` pointers. (Native memory is otherwise harness-managed — no section machinery or
entry cap to enforce.) Report: "Pruned N stale/wrong memories."

### 4. Write Lesson

Create a native memory file per the `agent_docs/self-improvement.md` format (most lessons are
`type: feedback`), then add a one-line pointer to `MEMORY.md`:

```markdown
---
name: <short-kebab-case-slug>
description: <one-line summary for recall>
metadata:
  type: feedback
---

<the lesson.>
**Why:** <root cause or reasoning>
**How to apply:** <the corrected approach, generalized>
```

### 5. Promotion Check

If a pattern has recurred **3+ times** (its memory file notes multiple incidents), propose promoting
it into a permanent rule:

1. Propose promoting to a CLAUDE.md section or command rule. Ask the user before modifying CLAUDE.md.
2. On approval: add the rule to the target CLAUDE.md section and **delete** the now-redundant memory
   file (CLAUDE.md becomes the single source of truth for that rule).

### 6. Summarize

Tell user:
- What lesson was captured (or why none was needed)
- Which memory files were written/updated/pruned, and whether CLAUDE.md was modified
- Promotion candidates (if any)

## Rules

- Lessons must be **generalizable** — not incident-specific
- Always check for duplicates before writing
- Keep lessons concise
- Guidelines only — not code examples unless significantly clearer
