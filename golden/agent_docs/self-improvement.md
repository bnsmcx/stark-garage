# Self-Improvement System

> Reference document. Loaded by /pomo, /review-pr, and /wiggum when capturing lessons.
> Not read every session — see CLAUDE.md for behavioral instructions.

Self-improvement is persisted with the **harness-native file-based memory** system (the `# Memory`
behavior injected by Claude Code): one fact per file under the project's memory directory, indexed by
`MEMORY.md`, which is auto-loaded into context at session start. There is no separate lessons file or
database to maintain — the harness surfaces relevant memories automatically.

## Memory Format

Each lesson is one memory file with frontmatter:

```markdown
---
name: <short-kebab-case-slug>
description: <one-line summary — used to decide relevance during recall>
metadata:
  type: feedback   # feedback (corrections/approaches) | project | user | reference
---

<the fact.>
**Why:** <root cause or reasoning>
**How to apply:** <the corrected approach, generalized>

Link related memories with [[their-name]].
```

Most self-improvement lessons are `type: feedback`. After writing the file, add a one-line pointer to
`MEMORY.md`: `- [Title](file.md) — hook`.

Guidelines:
- Lessons must be **generalizable** — not specific to one incident.
- Include a code example only if it makes the rule significantly clearer.
- One fact per file; keep it concise. Link related facts with `[[name]]`.

## Triggers

Capture a memory when:
- After **any user correction** — record what went wrong and the better approach (with **Why**).
- After **/review-pr** finds issues Claude introduced — reflect on the root cause.
- After **/wiggum** retry failures (2+ attempts before validation passes).
- Before saving, check `MEMORY.md` for an existing file that covers it — **update that file** rather
  than creating a duplicate.

**/pomo is the primary entry point for all self-improvement.** Commands invoke /pomo rather than
doing inline reflection. /pomo handles evaluation, deduplication, and format compliance.

## Deduplication Rules

Before writing a new memory:
1. Scan `MEMORY.md` (and recalled memories) for the same pattern or keywords.
2. **Same pattern, new example:** update the existing file (sharpen the fact; note the recurrence).
3. **Related but distinct:** create a new file and cross-link with `[[name]]`.

## What NOT to save

Don't persist what the repo already records (code structure, past fixes, git history, CLAUDE.md) or
what only matters to the current conversation. Delete memories that turn out to be wrong. When a
lesson recurs often enough to be a standing rule, promote it into CLAUDE.md or a command and remove
the now-redundant memory file (leaving CLAUDE.md as the single source of truth for that rule).
