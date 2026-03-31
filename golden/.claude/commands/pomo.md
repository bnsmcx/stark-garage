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

Check both sources:

1. Read `.claude/lessons.md` for existing coverage
2. Search memory:
   ```bash
   toolbox-memory search --ns lesson --query "<pattern keywords>"
   ```

If duplicate found:
- **Same pattern:** Update existing lesson, increment incident count
- **Related but distinct:** Create new lesson, cross-reference

### 3b. Lifecycle Management

If lessons.md exceeds 40 entries, prune:
1. Scan for promoted lessons (already in CLAUDE.md or commands) → archive
2. Scan for stale lessons (no recent matches) → archive
3. Move to `.claude/lessons-archive.md`
4. Run `toolbox-memory prune` for memory lifecycle transitions
5. Report: "Pruned N lessons. Active: NN/40."

### 4. Write Lesson

Add to `.claude/lessons.md` per `agent_docs/self-improvement.md` format:

```markdown
### [Pattern name]
- **Wrong:** [what was done incorrectly]
- **Right:** [the correct approach]
- **Why:** [root cause or reasoning]
```

Also write to memory:
```bash
toolbox-memory write --ns lesson --agent pomo --key "<pattern-name>" --value '{"wrong":"...","right":"...","why":"...","scope":"<scope>"}'
```

**Lifecycle:**
- New lessons → Active (default)
- 2nd+ matching incident → update existing, mark Validated
- hit_count >= 3 → flag for promotion (Step 5)

### 5. Promotion Check

Check memory for high-hit lessons:
```bash
toolbox-memory read --ns lesson --key "<pattern-name>"
```

If hit_count >= 3: propose promoting to CLAUDE.md instruction or command rule. Ask user before modifying CLAUDE.md.

After promotion:
```bash
toolbox-memory promote --ns lesson --key "<pattern-name>" --to "CLAUDE.md [section]"
```

Remove from lessons.md with note: "Promoted to CLAUDE.md [section]"

### 6. Summarize

Tell user:
- What lesson was captured (or why none was needed)
- Which files updated (lessons.md, memory)
- Whether CLAUDE.md was modified
- Promotion candidates (if any)

## Rules

- Lessons must be **generalizable** — not incident-specific
- Always check for duplicates before writing
- Always write to BOTH lessons.md AND memory
- Keep lessons concise
- Guidelines only — not code examples unless significantly clearer
