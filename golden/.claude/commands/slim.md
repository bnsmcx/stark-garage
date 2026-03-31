---
name: slim
description: Audit the golden set for bloat, redundancy, and budget compliance — compress, prune memory, and remove content to stay within limits
user_invocable: true
---

# /slim — Golden Set Audit & Compression

Audit CLAUDE.md, agent_docs/, and the memory database for bloat, redundancy, and stale content. Compress or remove content to stay within budget limits.

## When to run

- After every 5th `/improve-golden-set` cycle
- When any file reaches 80% of its budget (flagged by `/improve-golden-set` Step 11)
- On user request
- When memory utilization seems high or entries appear stale

## Steps

### 1. Measure current state

Count lines and instructions for every file listed in `BUDGETS.md`. Report current utilization:

```
## Current Budget Utilization

| File | Lines | Budget | % | Instructions | Budget | % |
|------|-------|--------|---|-------------|--------|---|
| CLAUDE.md (baseline) | NN | 60 | NN% | NN | 25 | NN% |
| CLAUDE.md (project) | NN | 80 | NN% | NN | 30 | NN% |
| agent_docs/issue-conventions.md | NN | 120 | NN% | — | — | — |
| agent_docs/issue-tracker-ops.md | NN | 120 | NN% | — | — | — |
| agent_docs/self-improvement.md | NN | 120 | NN% | — | — | — |
| .claude/lessons.md | NN entries | 40 | NN% | — | — | — |
| settings.local.json (allow) | NN entries | 100 | NN% | — | — | — |
```

Use the instruction counting rules from `BUDGETS.md` "What counts as an instruction".

### 2. Redundancy scan

For each instruction in CLAUDE.md, check:

- **Duplicated in commands?** — If a command file already enforces this rule in its steps, the CLAUDE.md instruction may be redundant.
- **Trained-in behavior?** — Is this something the model does well by default without being told? (e.g., "follow existing conventions", "no hardcoded values")
- **Internal duplication?** — Is this instruction stated twice in CLAUDE.md under different wording?

Flag duplicates and trained-in behaviors for removal. Present evidence for each flag.

### 3. Lessons.md pruning

Read `.claude/lessons.md` and evaluate each entry:

- **Promoted:** Encoded into a CLAUDE.md instruction or command rule? Flag for removal (note: "Promoted to [location]")
- **Project-specific:** Specific to one project rather than universal? Keep in project but don't extract to golden
- **Mergeable:** 2+ lessons expressing the same principle? Flag for merge
- **Stale:** No matching incidents in recent project context? Flag for archival to `.claude/lessons-archive.md`

### 4. Memory pruning

Run memory lifecycle transitions to keep the database lean:

```bash
toolbox-memory prune
```

This transitions entries through lifecycle states:
- **Active entries with no hits in 60 days** transition to **stale**
- **Stale entries with no hits for 30 more days** transition to **archived**

If `toolbox-memory` is not available, skip this step and note it in the report.

After pruning, query memory utilization:

```bash
toolbox-memory stats
```

Report memory state:

```
## Memory Utilization

| Lifecycle State | Count |
|----------------|-------|
| Active | NN |
| Validated | NN |
| Promoted | NN |
| Stale | NN |
| Archived | NN |
| **Total** | **NN** |

Pruning actions taken:
- N entries transitioned: active -> stale (no hits in 60+ days)
- N entries transitioned: stale -> archived (no hits in 90+ days)
```

### 5. Reference data audit

For each file in `agent_docs/`:

- **Accuracy:** Is the reference data still correct? (Check CLI commands, format specs)
- **Referenced:** Is it referenced by at least one command? (Dead references: flag for removal)
- **Current:** Has the format or tooling changed since this was written?

### 6. Present findings

Group all findings into categories:

```
## Audit Findings

### Remove (N items)
- [item] — [reason: redundant with command X / trained-in / promoted]

### Merge (N items)
- [lesson A] + [lesson B] -> [merged version]

### Compress (N items)
- [item]: current NN lines -> proposed NN lines
  [show compressed version]

### Memory Pruned (N entries)
- N active -> stale
- N stale -> archived

### Keep (N items)
- [item] — [reason it's still valuable]
```

Wait for user approval before applying any changes.

### 7. Apply approved changes

For each approved change:
- Remove flagged content from CLAUDE.md, lessons.md, or agent_docs/
- Merge lessons as approved
- Apply compressed versions
- Move stale lessons to `.claude/lessons-archive.md`

Memory pruning (Step 4) is applied automatically since it only transitions lifecycle states — it does not delete data.

### 8. Post-audit measurement

Re-measure all budgeted files and report before/after:

```
## Audit Results

| File | Before | After | Change |
|------|--------|-------|--------|
| CLAUDE.md baseline | NN/60 (NN%) | NN/60 (NN%) | -N lines |
| lessons.md | NN/40 entries | NN/40 entries | -N entries |
| Memory (active) | NN | NN | -N entries |
| Memory (total) | NN | NN | -N entries (archived) |
| ... | ... | ... | ... |
```

### 9. Changelog entry

Append an entry to `golden/CHANGELOG.md`:

```markdown
## [Date] — /slim audit

### Removed
- [items removed with reasons]

### Merged
- [lessons merged]

### Compressed
- [items compressed]

### Memory Pruned
- N entries: active -> stale
- N entries: stale -> archived

### Budget impact
- CLAUDE.md baseline: NN/60 -> NN/60 lines (-N)
- lessons.md: NN/40 -> NN/40 entries (-N)
- Memory active: NN -> NN (-N)
```

## Rules

- NEVER remove content without user approval
- ALWAYS present evidence for redundancy claims (which command duplicates it, or why it's trained-in)
- ALWAYS measure before and after — quantify the impact
- ALWAYS update CHANGELOG.md with audit results
- ALWAYS report memory utilization when the memory database exists
- Memory pruning transitions lifecycle states — it does not permanently delete data
- When in doubt about whether something is trained-in, keep it
