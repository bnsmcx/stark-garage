#!/usr/bin/env python3
"""Aggregate .claude/usage/{messages,iterations}.jsonl into a markdown report.

Usage: python3 .claude/usage/report.py [WINDOW]
  WINDOW = 1d | 7d | 30d | all  (default 7d)

Joins each token-usage record to the most recent open iteration in the same
session (started before the message timestamp, no end yet, or end after).
Records outside any iteration window are bucketed as "unattributed"."""
import json
import pathlib
import statistics
import sys
from collections import defaultdict
from datetime import datetime, timedelta, timezone

USAGE_DIR = pathlib.Path(".claude/usage")
MESSAGES = USAGE_DIR / "messages.jsonl"
ITERATIONS = USAGE_DIR / "iterations.jsonl"


def parse_window(arg: str) -> datetime:
    if arg == "all":
        return datetime.fromtimestamp(0, tz=timezone.utc)
    suffix = arg[-1]
    n = int(arg[:-1])
    if suffix == "d":
        return datetime.now(timezone.utc) - timedelta(days=n)
    if suffix == "h":
        return datetime.now(timezone.utc) - timedelta(hours=n)
    raise ValueError(f"unknown window: {arg}")


def load_jsonl(path: pathlib.Path) -> list[dict]:
    if not path.exists():
        return []
    out = []
    with open(path) as f:
        for line in f:
            try:
                out.append(json.loads(line))
            except Exception:
                continue
    return out


def parse_ts(ts: str | None) -> datetime | None:
    if not ts:
        return None
    try:
        return datetime.fromisoformat(ts.replace("Z", "+00:00"))
    except Exception:
        return None


def attribute(messages: list[dict], iterations: list[dict]) -> list[dict]:
    """Tag each message with (command, issue) by joining on session × time window."""
    starts: dict[tuple[str, int], datetime] = {}
    ends: dict[tuple[str, int], datetime] = {}
    issue_command: dict[tuple[str, int], str] = {}

    for it in iterations:
        cmd = it.get("command", "unknown")
        issue = it.get("issue")
        ts = parse_ts(it.get("ts"))
        if issue is None or ts is None:
            continue
        key = (cmd, issue)
        if it.get("event") == "iter_start":
            if key not in starts or ts < starts[key]:
                starts[key] = ts
            issue_command[key] = cmd
        elif it.get("event") == "iter_end":
            if key not in ends or ts > ends[key]:
                ends[key] = ts

    windows = sorted(
        (
            (start, ends.get(key, start + timedelta(hours=2)), key[0], key[1])
            for key, start in starts.items()
        ),
        key=lambda w: w[0],
    )

    out = []
    for m in messages:
        ts = parse_ts(m.get("ts"))
        cmd, issue = "unattributed", None
        if ts:
            for start, end, c, i in windows:
                if start <= ts <= end:
                    cmd, issue = c, i
        m2 = dict(m)
        m2["_command"] = cmd
        m2["_issue"] = issue
        out.append(m2)
    return out


def billable(rec: dict) -> int:
    return (
        rec.get("input", 0)
        + rec.get("output", 0)
        + rec.get("cache_creation", 0)
        + rec.get("cache_read", 0) // 10
    )


def fmt(n: int) -> str:
    if n >= 1_000_000:
        return f"{n / 1_000_000:.2f}M"
    if n >= 1_000:
        return f"{n / 1_000:.1f}k"
    return str(n)


def main() -> int:
    window_arg = sys.argv[1] if len(sys.argv) > 1 else "7d"
    cutoff = parse_window(window_arg)

    messages = [
        m for m in load_jsonl(MESSAGES)
        if (parse_ts(m.get("ts")) or datetime.fromtimestamp(0, tz=timezone.utc)) >= cutoff
    ]
    iterations = load_jsonl(ITERATIONS)

    if not messages:
        print(f"# Usage report ({window_arg})\n\nNo messages recorded in window.")
        return 0

    tagged = attribute(messages, iterations)

    totals = defaultdict(int)
    by_command = defaultdict(lambda: defaultdict(int))
    by_issue = defaultdict(lambda: defaultdict(int))
    by_session_subagents = defaultdict(int)
    high_context = 0

    for m in tagged:
        for k in ("input", "output", "cache_read", "cache_creation"):
            totals[k] += m.get(k, 0)
        totals["billable"] += billable(m)
        cmd = m["_command"]
        issue = m["_issue"]
        by_command[cmd]["billable"] += billable(m)
        by_command[cmd]["count"] += 1
        if issue is not None:
            by_issue[(cmd, issue)]["billable"] += billable(m)
            by_issue[(cmd, issue)]["count"] += 1
        if m.get("event") == "SubagentStop":
            by_session_subagents[m.get("session", "?")] += 1
        if m.get("input", 0) + m.get("cache_read", 0) > 150_000:
            high_context += 1

    lines: list[str] = []
    lines.append(f"# Usage report ({window_arg})")
    lines.append("")
    lines.append(f"- Messages: {len(tagged)}")
    lines.append(f"- Input: {fmt(totals['input'])}  Output: {fmt(totals['output'])}")
    lines.append(
        f"- Cache read: {fmt(totals['cache_read'])}  Cache creation: {fmt(totals['cache_creation'])}"
    )
    lines.append(f"- Billable estimate: {fmt(totals['billable'])}")
    lines.append("")

    lines.append("## By command")
    lines.append("| Command | Billable | Messages | % |")
    lines.append("|---|---:|---:|---:|")
    grand = totals["billable"] or 1
    for cmd, d in sorted(by_command.items(), key=lambda x: -x[1]["billable"]):
        pct = 100 * d["billable"] / grand
        lines.append(f"| {cmd} | {fmt(d['billable'])} | {d['count']} | {pct:.0f}% |")
    lines.append("")

    lines.append("## By issue (top 10)")
    issue_rows = sorted(by_issue.items(), key=lambda x: -x[1]["billable"])[:10]
    if issue_rows:
        median = statistics.median(d["billable"] for _, d in issue_rows) or 1
        lines.append("| Command | Issue | Billable | Flag |")
        lines.append("|---|---:|---:|---|")
        for (cmd, issue), d in issue_rows:
            flag = "outlier" if d["billable"] > 2 * median else ""
            lines.append(f"| {cmd} | #{issue} | {fmt(d['billable'])} | {flag} |")
    else:
        lines.append("_No iteration markers recorded — run `/wiggum` or `/close-issue` first._")
    lines.append("")

    lines.append("## Hot subagent sessions")
    hot = sorted(by_session_subagents.items(), key=lambda x: -x[1])[:5]
    if hot:
        lines.append("| Session | SubagentStop count |")
        lines.append("|---|---:|")
        for sid, n in hot:
            lines.append(f"| {sid[:8]}… | {n} |")
    else:
        lines.append("_No subagent stops recorded in window._")
    lines.append("")

    pct_high = 100 * high_context / len(tagged)
    lines.append(f"## Context pressure: {pct_high:.0f}% of messages above 150k context")
    if pct_high > 50:
        lines.append(
            "**Lever:** `/clear` between issues in `/wiggum`, or move heavy reference docs out of CLAUDE.md."
        )

    print("\n".join(lines))
    return 0


if __name__ == "__main__":
    sys.exit(main())
