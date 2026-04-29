#!/usr/bin/env python3
"""Hook target: append per-message token usage to .claude/usage/messages.jsonl.

Wired to Stop and SubagentStop. Reads the hook payload from stdin, walks the
referenced transcript, and records each assistant message's usage exactly once
per session via a small cursor file."""
import json
import os
import pathlib
import sys

USAGE_DIR = pathlib.Path(".claude/usage")
USAGE_DIR.mkdir(parents=True, exist_ok=True)

try:
    payload = json.load(sys.stdin)
except Exception:
    sys.exit(0)

transcript = payload.get("transcript_path")
session = payload.get("session_id") or "unknown"
event = payload.get("hook_event_name") or "unknown"
cwd = payload.get("cwd") or os.getcwd()

if not transcript or not os.path.isfile(transcript):
    sys.exit(0)

cursor = USAGE_DIR / f".cursor-{session}"
try:
    seen = int(cursor.read_text().strip()) if cursor.exists() else 0
except Exception:
    seen = 0

records = []
with open(transcript) as f:
    for line in f:
        try:
            row = json.loads(line)
        except Exception:
            continue
        if row.get("type") != "assistant":
            continue
        msg = row.get("message") or {}
        usage = msg.get("usage")
        if not usage:
            continue
        records.append({
            "ts": row.get("timestamp"),
            "uuid": row.get("uuid"),
            "session": session,
            "event": event,
            "cwd": cwd,
            "model": msg.get("model"),
            "input": usage.get("input_tokens", 0),
            "output": usage.get("output_tokens", 0),
            "cache_read": usage.get("cache_read_input_tokens", 0),
            "cache_creation": usage.get("cache_creation_input_tokens", 0),
        })

new = records[seen:]
if new:
    with open(USAGE_DIR / "messages.jsonl", "a") as f:
        for r in new:
            f.write(json.dumps(r) + "\n")
    cursor.write_text(str(len(records)))
elif not cursor.exists():
    cursor.write_text(str(len(records)))
