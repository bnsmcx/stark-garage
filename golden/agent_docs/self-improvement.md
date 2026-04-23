# Self-Improvement System

> Reference document. Loaded by /pomo, /review-pr, and /wiggum when updating lessons.
> Not read every session — see CLAUDE.md for behavioral instructions.

## Lessons Format

Each entry in `.claude/lessons.md` follows this format:

```markdown
### [Pattern name]
- **Wrong:** [what was done incorrectly]
- **Right:** [the correct approach]
- **Why:** [root cause or reasoning]
```

Guidelines:
- Rules should be **generalizable** — not specific to one incident
- Include a code example only if it makes the rule significantly clearer
- Keep each lesson concise

## Triggers

Update `.claude/lessons.md` when:
- After **any user correction** — capture what went wrong and the better approach
- After **/review-pr** finds issues Claude introduced — reflect on why
- After **/wiggum** retry failures (2+ attempts before validation passes)
- Ruthlessly deduplicate — update existing entries rather than adding redundant ones

**/pomo is the primary entry point for all self-improvement.** Commands invoke
/pomo rather than doing inline reflection. /pomo handles evaluation, deduplication,
format compliance, and lifecycle management.

## Deduplication Rules

Before writing a new lesson:
1. Read `.claude/lessons.md` and grep for the pattern name or keywords
2. **Same pattern, new example:** Update the existing lesson's incident count
3. **Related but distinct:** Create a new lesson and cross-reference the related one

## Lesson Lifecycle

Every lesson has an explicit lifecycle, tracked by markdown section in `.claude/lessons.md`:

1. **Active** — Recently captured, not yet proven across multiple incidents. Sits under `## Active`.
2. **Validated** — Confirmed by recurrence (2+ incidents). Moved to `## Validated`.
3. **Promoted** — Encoded into CLAUDE.md or a command as a permanent rule. Move the entry into the target CLAUDE.md section; leave a one-line stub under `## Promoted` referencing the new location.
4. **Archived** — No matching incidents in 60+ days. Moved out of `.claude/lessons.md` into `.claude/lessons-archive.md` under the archival date.

## Pruning Rules

Enforced by /slim and /pomo:

- **Max entries:** 40 active lessons in `.claude/lessons.md`
- **Promotion:** When a pattern has 3+ confirmed incidents, flag for promotion to CLAUDE.md or a command rule
- **Archival:** On `/pomo` invocation, scan `## Active` entries; any with no matching incident in 60+ days moves to `.claude/lessons-archive.md` under the date
- **Deduplication:** Always merge rather than creating a second entry for the same pattern
