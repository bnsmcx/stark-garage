---
name: pomo
description: Post-mortem reflection — capture lessons and write to memory
user_invocable: true
---

# /pomo — Post-Mortem

Reflect on a recently resolved issue, capture generalizable lessons, persist to lessons.md and SQLite memory.

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

```bash
grep -ri "<pattern keywords>" .claude/lessons.md .claude/lessons-archive.md
```

If duplicate found:
- **Same pattern:** Update existing lesson, increment incident count in place
- **Related but distinct:** Create new lesson, cross-reference

### 3b. Lifecycle Management

`.claude/lessons.md` is organized into four sections: `## Active`, `## Validated`, `## Promoted`, and a pointer to `.claude/lessons-archive.md`. Before writing a new lesson, scan for pruning candidates:

1. **Stale `## Active` entries** (no matching incident in 60+ days) → move to `.claude/lessons-archive.md` under the archival date
2. **`## Active` entries with a second matching incident** → move to `## Validated`
3. **Lessons already encoded in CLAUDE.md or a command** → move content out; leave a one-line stub under `## Promoted` pointing to the CLAUDE.md section

Report: "Pruned N lessons. Active: NN/40."

### 4. Write Lesson

Add to the `## Active` section of `.claude/lessons.md` per `agent_docs/self-improvement.md` format:

```markdown
### [Pattern name]
- **Wrong:** [what was done incorrectly]
- **Right:** [the correct approach]
- **Why:** [root cause or reasoning]
```

**Lifecycle transitions (all markdown edits — no CLI calls):**
- New lessons → append under `## Active`
- 2nd+ matching incident → update existing, move to `## Validated`
- 3+ incidents → propose promotion (Step 5)

### 5. Promotion Check

Scan `## Validated` for entries with 3+ confirmed incidents. For each:

1. Propose promoting to a CLAUDE.md section or command rule. Ask user before modifying CLAUDE.md.
2. On approval: move the entry's content into the target CLAUDE.md section, and leave a one-line stub under `## Promoted` in `.claude/lessons.md` of the form `- [Pattern name] → CLAUDE.md [section]`.

### 6. Summarize

Tell user:
- What lesson was captured (or why none was needed)
- Which files updated (lessons.md, archive, CLAUDE.md)
- Whether CLAUDE.md was modified
- Promotion candidates (if any)

## Rules

- Lessons must be **generalizable** — not incident-specific
- Always check for duplicates before writing
- Keep lessons concise
- Guidelines only — not code examples unless significantly clearer
