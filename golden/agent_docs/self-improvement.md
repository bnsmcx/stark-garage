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

Update `.claude/lessons.md` AND write to memory when:
- After **any user correction** — capture what went wrong and the better approach
- After **/review-pr** finds issues Claude introduced — reflect on why
- After **/wiggum** retry failures (2+ attempts before validation passes)
- Ruthlessly deduplicate — update existing entries rather than adding redundant ones

**/pomo is the primary entry point for all self-improvement.** Commands invoke
/pomo rather than doing inline reflection. /pomo handles evaluation, deduplication,
format compliance, lifecycle management, and memory writes.

## Memory Integration

When writing a lesson to `.claude/lessons.md`, also persist to SQLite memory:
```bash
toolbox-memory write --ns lesson --agent pomo --key "<pattern-name>" --value '{"wrong":"...","right":"...","why":"..."}'
```

When checking for existing lessons, also search memory:
```bash
toolbox-memory search --ns lesson --query "<pattern keywords>"
```

## Deduplication Rules

Before writing a new lesson:
1. Read `.claude/lessons.md` and check for existing coverage
2. Search memory: `toolbox-memory search --ns lesson --query "<pattern>"`
3. **Same pattern, new example:** Update the existing lesson's incident count
4. **Related but distinct:** Create a new lesson and cross-reference the related one

## Lesson Lifecycle

Every lesson has an implicit lifecycle (tracked in both lessons.md and memory):

1. **Active** — Recently captured, not yet proven across multiple incidents.
2. **Validated** — Confirmed by recurrence (2+ incidents / hit_count >= 2).
3. **Promoted** — Encoded into CLAUDE.md or a command as a permanent rule.
   Remove from lessons.md with note: "Promoted to CLAUDE.md [section]".
   Mark in memory: `toolbox-memory promote --ns lesson --key "<name>" --to "CLAUDE.md [section]"`
4. **Stale** — No matching incidents in 60+ days. Candidate for archival.

## Pruning Rules

Enforced by /slim and /pomo:

- **Max entries:** 40 active lessons in `.claude/lessons.md`
- **Promotion:** When hit_count >= 3, flag for promotion to CLAUDE.md or command rule
- **Archival:** Lessons with no hits in 60+ days move to `.claude/lessons-archive.md`
- **Memory pruning:** Run `toolbox-memory prune` to transition lifecycle states
- **Deduplication:** Always merge rather than creating a second entry for the same pattern
