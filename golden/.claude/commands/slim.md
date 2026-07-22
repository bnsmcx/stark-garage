---
name: slim
description: Audit the golden set for bloat, redundancy, and budget compliance — compress, prune memory, and remove content to stay within limits
user_invocable: true
---

# /slim — Golden Set Audit & Compression

Audit CLAUDE.md, agent_docs/, and native memory for bloat, redundancy, and stale content. Compress or remove content to stay within budget limits.

## When to run

- After every 5th `/improve-golden-set` cycle
- When any file reaches 80% of its budget (flagged by `/improve-golden-set` Step 11)
- On user request
- When native memory accumulates stale or superseded facts

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
| settings.local.json (allow) | NN entries | 100 | NN% | — | — | — |
```

Use the instruction counting rules from `BUDGETS.md` "What counts as an instruction".

### 2. Redundancy scan

For each instruction in CLAUDE.md, check:

- **Duplicated in commands?** — If a command file already enforces this rule in its steps, the CLAUDE.md instruction may be redundant.
- **Trained-in behavior?** — Is this something the model does well by default without being told? (e.g., "follow existing conventions", "no hardcoded values")
- **Internal duplication?** — Is this instruction stated twice in CLAUDE.md under different wording?

Flag duplicates and trained-in behaviors for removal. Present evidence for each flag.

### 3. Native memory pruning

Scan `MEMORY.md` (and the auto-recalled memories) and evaluate each fact:

- **Promoted:** already encoded into a CLAUDE.md instruction or command rule? Flag the memory file for deletion (note: "Promoted to [location]").
- **Stale/superseded:** no longer accurate, or overtaken by a code/convention change? Flag for deletion.
- **Mergeable:** 2+ facts expressing the same principle? Flag to merge into one file.
- **Wrong:** contradicted by how the code actually behaves? Flag for deletion.

Native memory is harness-managed (no entry cap or lifecycle states to enforce) — pruning here just
means deleting obsolete facts and their `MEMORY.md` pointers. Report a count of memories flagged.

### 4. Reference data audit

For each file in `agent_docs/`:

- **Accuracy:** Is the reference data still correct? (Check CLI commands, format specs)
- **Referenced:** Is it referenced by at least one command? (Dead references: flag for removal)
- **Current:** Has the format or tooling changed since this was written?

### 5. Present findings

Group all findings into categories:

```
## Audit Findings

### Remove (N items)
- [item] — [reason: redundant with command X / trained-in / promoted]

### Merge (N items)
- [item A] + [item B] -> [merged version]

### Compress (N items)
- [item]: current NN lines -> proposed NN lines
  [show compressed version]

### Memory Pruned (N facts)
- [memory file] — [reason: promoted / stale / wrong / merged]

### Keep (N items)
- [item] — [reason it's still valuable]
```

Wait for user approval before applying any changes.

### 6. Apply approved changes

For each approved change:
- Remove flagged content from CLAUDE.md or agent_docs/
- Merge as approved
- Apply compressed versions
- Delete flagged native-memory facts and their `MEMORY.md` pointers

### 7. Post-audit measurement

Re-measure all budgeted files and report before/after:

```
## Audit Results

| File | Before | After | Change |
|------|--------|-------|--------|
| CLAUDE.md baseline | NN/60 (NN%) | NN/60 (NN%) | -N lines |
| agent_docs/self-improvement.md | NN | NN | -N lines |
| Native memory (facts) | NN | NN | -N facts |
| ... | ... | ... | ... |
```

### 8. Changelog entry

Append an entry to `golden/CHANGELOG.md`:

```markdown
## [Date] — /slim audit

### Removed
- [items removed with reasons]

### Merged
- [items merged]

### Compressed
- [items compressed]

### Memory Pruned
- N native-memory facts deleted (promoted / stale / wrong)

### Budget impact
- CLAUDE.md baseline: NN/60 -> NN/60 lines (-N)
- Native memory: NN -> NN facts (-N)
```

## Rules

- NEVER remove content without user approval
- ALWAYS present evidence for redundancy claims (which command duplicates it, or why it's trained-in)
- ALWAYS measure before and after — quantify the impact
- ALWAYS update CHANGELOG.md with audit results
- ALWAYS report memory utilization when the memory database exists
- Memory pruning transitions lifecycle states — it does not permanently delete data
- When in doubt about whether something is trained-in, keep it
