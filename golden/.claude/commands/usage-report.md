---
name: usage-report
description: Aggregate token usage from .claude/usage/ into a per-command, per-issue breakdown
user_invocable: true
---

# /usage-report — Token Usage Breakdown

Reports per-command and per-issue token usage from instrumentation written by
the `Stop`/`SubagentStop` hooks (`.claude/usage/messages.jsonl`) and the
iteration markers emitted by `/wiggum` and `/close-issue`
(`.claude/usage/iterations.jsonl`).

## Invocation

```
/usage-report                  # Last 7 days
/usage-report 1d               # Last 24 hours
/usage-report 30d              # Last 30 days
/usage-report all              # Everything recorded
```

## Steps

### 1. Resolve window

Parse the argument to a UTC ISO cutoff. Default `7d`. `all` → epoch.

### 2. Verify data exists

If `.claude/usage/messages.jsonl` is missing, report that hooks haven't fired
yet (no Stop event since instrumentation was added) and stop.

### 3. Aggregate

Run the aggregator below — it joins messages to iteration windows by
`session × timestamp` and emits markdown.

```bash
python3 .claude/usage/report.py "$WINDOW"
```

### 4. Print the report

Print the markdown produced by the aggregator. Sections:

- **Totals** — input, output, cache_read, cache_creation, billable estimate
- **By command** — `/wiggum`, `/close-issue`, `unattributed`, …
- **By issue** — sorted by total tokens descending; flag any issue >2x the median
- **Hot subagents** — count of `SubagentStop` events per session, top 5
- **Context-pressure flag** — pct of messages where `cache_read + input > 150k`

### 5. Suggest next levers

Based on what the report shows, surface the biggest single lever:

- If one command dominates (>30%): suggest scoping it down or routing to a
  cheaper model via skill/agent frontmatter
- If subagent count is high relative to messages: suggest conditional fan-out
  in `/wiggum` step 8 (only invoke `security-reviewer` / `ops-reviewer` for
  diffs that touch their domains)
- If context-pressure flag >50%: suggest `/clear` between issues in the
  release loop, or compaction of `agent_docs/` referenced from CLAUDE.md

## Rules

- Read-only — never modify or delete jsonl files (audit trail)
- If `iterations.jsonl` is missing but `messages.jsonl` exists, still produce
  the totals + hot-subagents view; mark per-command/per-issue as unattributed
- Token costs are estimates; treat as relative not absolute
