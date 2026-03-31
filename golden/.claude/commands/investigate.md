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
- **Probe the local API** (port 80) freely — GET requests, dev tokens, schema inspection.
- **Iterate with the user.** Present findings, ask clarifying questions, refine understanding. Do not rush to a plan.
- **Output a plan, not code.** The deliverable is a shared understanding + a plan outline ready for `/create-issues`.

## Step 0. Capture the Feature Request

If an issue number was provided:
1. Fetch the issue using `gh issue view <number>`
2. Extract title, body, labels, and any linked issues

If text was provided, use it directly.

If nothing was provided, ask: "What feature would you like me to investigate?"

Restate the feature request in 1-2 sentences and confirm understanding with the user before proceeding.

## Step 1. Obtain API Access

Get a dev token for local API exploration:

```bash
# Discover available dev token routes
curl -s http://localhost/api/dev/token | jq .
```

Store the token for use in subsequent API calls. If dev tokens are unavailable or the API is not running, note this and proceed with static analysis only.

## Step 2. Codebase Reconnaissance

Launch **parallel Explore subagents** to investigate the areas of the codebase most likely affected by the feature. Typical targets:

- **Routes & Handlers**: Which existing endpoints are adjacent? What patterns do similar features follow?
- **Models & Storage**: What database models exist? What queries would need to change or be added?
- **Services**: Is there existing business logic to extend or a new service needed?
- **Auth/ACL**: Does this feature need new permissions or change existing access patterns?
- **OpenAPI Schema**: What does the current API contract look like for related resources?

Tailor the subagent tasks to the specific feature — don't explore irrelevant areas.

Also inspect the Swagger docs:
```bash
curl -s http://localhost/swagger/doc.json | jq '.paths | keys' # list all endpoints
curl -s http://localhost/swagger/doc.json | jq '.definitions | keys' # list all models
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

## Step 7. Handoff to /create-issues

Once the user approves the plan:

1. Confirm the plan is in conversation context (it should be from Step 6)
2. Ask if there's an assignee preference
3. Instruct the user to run `/create-issues` (or offer to invoke it)

The plan outline from Step 6 is designed to be directly consumable by `/create-issues`.

## Rules

- **NEVER edit project files** — this is a research-only command
- **NEVER skip user confirmation** — always iterate, never assume alignment
- **ALWAYS use subagents** for broad codebase exploration to keep context clean
- **ALWAYS probe the live API** when it's available — don't rely solely on reading code
- **ALWAYS present open questions** — surface ambiguity rather than guessing
- **ALWAYS structure findings** — use tables and headers, not walls of text
- One feature request per invocation — if the user has multiple, run `/investigate` for each
- If the API is not running locally, note it and proceed with static analysis
- Keep API calls read-only (GET requests, no mutations)
