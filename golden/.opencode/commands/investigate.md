---
name: investigate
description: Research a feature request — deep-dive the codebase, probe the local API, surface tradeoffs, and iterate toward a plan ready for /create-issues
user_invocable: true
---

# /investigate — Feature Request Research Pipeline

Deep-dive a proposed feature: understand the current state, identify affected layers, surface tradeoffs, and iterate with the user until there's an agreed plan ready for `/create-issues`.

## Invocation

```
/investigate                     # Prompt for the feature request description
/investigate <description>       # Inline feature request text
/investigate #NN                 # Pull context from an existing issue
```

The argument is a **feature request** — free-form text, an issue number, or nothing (will prompt).

## Ground Rules

- **Read-only.** Do not edit any project files. Research only.
- **Use subagents** for parallel codebase exploration (Explore agents) to keep the main context lean.
- **Probe the local API** freely — GET requests, dev tokens, schema inspection. If the app isn't running, **launch it** (see Step 1) rather than skipping live probing.
- **Iterate with the user.** Present findings, ask clarifying questions, refine understanding. Do not rush to a plan.
- **Output a plan, not code.** The deliverable is a shared understanding + a plan outline ready for `/create-issues`.

## Step 0. Capture the Feature Request

If an issue number was provided:
1. Fetch the issue using `gh issue view <number>`
2. Extract title, body, labels, and any linked issues

If text was provided, use it directly.

If nothing was provided, ask: "What feature would you like me to investigate?"

Restate the feature request in 1-2 sentences and confirm understanding with the user before proceeding.

## Step 1. Ensure the App Is Running, Then Obtain API Access

Live API probing is a hard requirement — do not skip to static analysis just because the server isn't up. Boot it yourself.

### 1a. Detect running instance

Probe the project's known health endpoint(s). The base URL and the run command live in `agent_docs/build-and-test.md` (or CLAUDE.md's project-specific section). Record the base URL if reachable.

### 1b. Launch the app if it isn't running

Start the server as a **background** Bash process (`run_in_background: true`) so the rest of the investigation can proceed. Use the project's primary run command (see `agent_docs/build-and-test.md`). Redirect logs to a tempfile so you can grep them later if an endpoint misbehaves.

Wait for the health endpoint to return `200` with a sensible timeout (startup is usually <10s, but cold caches or local DBs can push it further). If it never comes up, surface the error, tail the log, and ask whether to continue with static analysis or fix the boot issue.

**Important:** you started this process — **stop it when the investigation ends** (Step 7). Leaving it running wastes resources and can collide with the user's own dev server. Use the specific binary name or port from the project's run command; generic `pkill -f "go run"` or `pkill node` often misses the actual listening process.

### 1c. Dev token (if needed)

If the project exposes a dev-token endpoint (check the route list or existing docs), fetch one:

```bash
curl -s "$BASE/api/dev/token" | jq . 2>/dev/null || echo "no dev token endpoint"
```

Store the token for subsequent calls. Absence of a dev token is fine — many endpoints are unauthenticated on localhost.

## Step 2. Codebase Reconnaissance

Launch **parallel Explore subagents** to investigate the areas of the codebase most likely affected by the feature. Typical targets:

- **Routes & Handlers**: Which existing endpoints are adjacent? What patterns do similar features follow?
- **Models & Storage**: What database models exist? What queries would need to change or be added?
- **Services**: Is there existing business logic to extend or a new service needed?
- **Auth/ACL**: Does this feature need new permissions or change existing access patterns?
- **OpenAPI Schema**: What does the current API contract look like for related resources?

Tailor the subagent tasks to the specific feature — don't explore irrelevant areas.

Also inspect the Swagger docs (if the project exposes them):
```bash
curl -s "$BASE/swagger/doc.json" | jq '.paths | keys'       # list all endpoints
curl -s "$BASE/swagger/doc.json" | jq '.definitions | keys' # list all models
```

## Step 3. Probe Current Behavior

Make targeted API calls to understand the current state of related resources:

- List/get existing resources that the feature would touch
- Check what fields, filters, and relationships are already exposed
- Note any gaps between what the API exposes and what the feature needs

Use the dev token from Step 1. Document all findings.

## Step 4. Impact Analysis

Synthesize the research into a structured assessment:

### Affected Layers
| Layer | Files/Areas | Nature of Change |
|-------|------------|-----------------|
| Routes | ... | New route / modify existing |
| Handlers | ... | New handler / extend existing |
| Services | ... | New service / extend existing |
| Storage | ... | New queries / new model / migration |
| Auth/ACL | ... | New permission / existing sufficient |

### Schema/DDL Consistency Check
If the investigation identifies model or schema changes (new columns, nullable changes, type changes, new tables):
1. Search for SQL init/migration files that define the affected table(s) (e.g., `docker-entrypoint-initdb.d/`, `migrations/`, `schema.sql`)
2. Compare the proposed model change against the DDL definition
3. Flag any constraint mismatches (NOT NULL vs nullable, type differences, missing columns)
4. Include required DDL changes in the impact analysis

This catches cases where a model change passes CI but fails at runtime because the database schema wasn't updated.

### Key Findings
- What exists today that the feature can build on
- What's missing that needs to be created
- Any surprising constraints or complications discovered

### Open Questions
- Ambiguities in the feature request
- Design decisions that need user input
- Tradeoffs worth discussing (performance, complexity, scope)

Present this to the user and discuss. **Do not proceed until questions are resolved.**

## Step 5. Iterate

This is the collaborative phase. Based on user answers:

- Refine understanding of scope and requirements
- Do additional targeted research if new questions arise
- Narrow down implementation approach
- Identify what's in scope vs. out of scope

Repeat Steps 2-4 as needed for newly surfaced areas. Continue until the user signals alignment.

## Step 6. Draft Plan Outline

Once aligned, present a plan outline structured for `/create-issues`:

```
## Feature: {feature title}

### Summary
{2-3 sentences on what we're building and why}

### Implementation Sequence

1. **{type}({scope}): {description}**
   - What: {what this step does}
   - Why: {why it's needed / what it unblocks}
   - Key files: {files likely touched}
   - Blocked by: {dependencies}

2. **{type}({scope}): {description}**
   ...

### Scope Boundaries
- In scope: {what's included}
- Out of scope: {what's explicitly deferred}
- Future considerations: {things to keep in mind but not implement now}

### Risks & Mitigations
- {risk}: {mitigation}
```

Ask: "Does this plan look right? I can adjust scope, ordering, or details before we run `/create-issues`."

## Step 7. Handoff to /create-issues (and clean up)

Once the user approves the plan:

1. Confirm the plan is in conversation context (it should be from Step 6)
2. Ask if there's an assignee preference
3. Instruct the user to run `/create-issues` (or offer to invoke it)
4. **If you launched the app in Step 1b**, stop it now. Use the specific binary name or port from the project's run command. Confirm it's down with a final health-check (should return a connection-refused or non-200). Leave any server the user started themselves alone.

The plan outline from Step 6 is designed to be directly consumable by `/create-issues`.

## Rules

- **NEVER edit project files** — this is a research-only command
- **NEVER skip user confirmation** — always iterate, never assume alignment
- **ALWAYS use subagents** for broad codebase exploration to keep context clean
- **ALWAYS probe the live API** — if it isn't running, launch it per Step 1b rather than falling back to static-only analysis
- **ALWAYS present open questions** — surface ambiguity rather than guessing
- **ALWAYS structure findings** — use tables and headers, not walls of text
- One feature request per invocation — if the user has multiple, run `/investigate` for each
- If you launched the server, shut it down in Step 7 (don't orphan it)
- Keep API calls read-only (GET requests, no mutations)
