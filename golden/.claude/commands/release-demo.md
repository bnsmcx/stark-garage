---
name: release-demo
description: Generate an interactive HTML release demo embedded in the dev API binary
user_invocable: true
---

# /release-demo — Interactive Release Demo HTML

Generate a self-contained interactive HTML page that walks every change in a release as a **live API round-trip**. The page is committed alongside the source, embedded into the binary (e.g. via `go:embed` in `internal/demo/`), and served at a dev route (e.g. `GET /demo/{filename}`) when the app runs in dev mode. The same file opens as a standalone document via `file://` for emailing.

This supersedes the old shell-script + recorded-GIF flow. We no longer generate E2E shell scripts, `.tape` files, or upload demo GIFs to releases — reviewers want to *click endpoints* and *see request/response*, not watch a terminal recording.

## Invocation

```
/release-demo                    # Auto-detect release branch + milestone
/release-demo X.Y.Z              # Target a specific version
```

## Steps

### 1. Detect Release Context

```bash
git branch --show-current        # expect release/*
```

If not on a release branch, ask the user for the version.

Fetch all closed issues in the milestone:
```bash
gh issue list --milestone "MILESTONE" --state closed --limit 200 --json number,title,body,labels
```

### 2. Copy Scaffolding From the Most Recent Demo

```bash
ls -t internal/demo/release-*.html | head -1   # or wherever demos live
```

Copy the inline CSS, the inline JS scaffolding (state object, `http()` / `httpOnce()` with 401-retry, `formatRequest()`, `bootstrap()`, `runAll()`, `resetAll()`, verdict pill helpers, the tab framework `activateTab()` + `renderMermaidIn()`, and mermaid wiring), and the overall page skeleton **verbatim**. Only the per-release content (frontmatter values, tab content, per-issue sections, `RUNNERS` map, `STEP_ORDER`) changes. **Do not re-derive the scaffolding** — it has been audited for:

- token refresh on 401, `file://` vs dev-API origin detection, and CORS skip behavior
- **Curl-formatted request previews with a copy button** — `formatRequest()` renders each request as a copy-paste-runnable `curl` command (Bearer shown as `$API_TOKEN`), with a "Copy curl" button. Don't revert to a raw fetch dump.
- **Tabbed layout** — `Overview / Demo / Frontend Changes / DevOps Info` panels driven by `activateTab()`. The sticky results bar is shown only on the Demo tab.
- **Lazy per-tab mermaid rendering** — `mermaid.initialize({ startOnLoad: false })` plus `renderMermaidIn(panel)` called from `activateTab()`. This is load-bearing: `startOnLoad: true` renders every diagram at page load, including those in `hidden` (zero-size) tab panels, and geometry-sensitive shapes (self-loops, cylinders) then render as a **"Syntax error" bomb even though the source parses cleanly**. Diagrams must render only when their tab is visible. (Learned the hard way — cost a full debugging cycle.)
- **Host-aware URL rewrite** — any signed/redirect URL the demo rewrites for browser-reachability derives its host from `new URL(state.apiBase).hostname` (falling back to `localhost`), never a hardcoded `localhost`. This lets the demo work both same-machine and from another box on the LAN. (Also learned the hard way.)

### 3. Verify Endpoint Shapes Against the OpenAPI Spec

Before writing any runner, fetch the spec (e.g. `/swagger/doc.json`):
```bash
curl -s "$BASE/swagger/doc.json" | jq '.paths["/api/ENDPOINT"]'
```

Use the spec — not memory — for the request shape, query parameters, and response field names. Pass the raw spec to subagents if you need them to do bulk jq queries.

### 4. Write `internal/demo/release-{version}.html`

The page serves **four audiences** — a manager/QC reader skimming what shipped, a frontend dev checking what they must change, an ops engineer wiring up config, and a reviewer who wants to click endpoints. Split it into **navigation tabs** so each audience lands on their content without scrolling past everyone else's. Copy the tab framework from the previous demo; only the panel content changes.

The file includes, in order:

**`<header class="frontmatter">`** (always visible, above the tabs)
- Kicker, `<h1>{version}</h1>`, theme line
- Vitals `<dl>`: Milestone, Issues count, Release PR (`gh pr view --json url`), Tag, language/toolchain pins that changed
- `.api-config` Connection panel: base URL input, Bootstrap button, status pill, short prose about what the bootstrap actually proves
- "What ships in {version}" headlines `<ul>` — one line per shipped item, scannable

**Tab strip** — `<button class="tab-btn" data-tab="...">` for each panel, then one `<section class="tab-panel" id="tab-{name}">` per tab:

- **Overview** — lede prose framing the release theme + marquee items. The place for a high-level `<div class="mermaid">` (e.g. a decision/architecture diagram) that orients a non-technical reader. No live steps here.
- **Demo** — the live API round-trips. Carries `Run all assertions` / `Reset` and the sticky results bar. One subsection per issue (or cluster):
  - `<h2><span class="num">N.</span>Title — #NN</h2>`, prose on *what changed and why* (link the issue)
  - Optional `<figure class="diagram">` flowchart for a non-trivial flow
  - One or more `<div class="step" data-step="STEP_ID">` widgets: `<h4>` + step-id span, prose, `<div class="controls">` (Run button, verdict span, optional `needs:` chain), hidden request (curl) and response panes
  - Optionally fold in `regression` (unchanged-endpoint sweep proving no collateral breakage), `negative` (bad IDs, 404s, 401s), and `deferred` (what was intentionally NOT shipped) when they sharpen the story; skip when they'd be filler.
- **Frontend Changes** — what a client dev must actually do: concrete tweaks with **file + line-number references and suggested implementation**, or an explicit "nothing changes for the frontend" if true. This mirrors the `/release-notes` "Frontend action" column.
- **DevOps Info** — the **exact env-var names the code reads** (grep them, don't recall) in a per-environment table, plus any operator walkthrough (e.g. a step-by-step config/provider switch). Mark local-only / never-in-prod vars loudly.

Thin tabs are fine — if a release genuinely has nothing for one audience, keep the tab and say so in one line rather than deleting it; readers learn where to look.

**Scripts** — mermaid CDN, then a self-invoking `<script>` block containing:
- `state` object (`apiBase`, `token`, `results`, plus any chained-step state)
- HTTP helpers + `formatRequest()` (curl) + `activateTab()`/`renderMermaidIn()` copied from the previous demo
- `RUNNERS` map: one async function per step ID, returning `{ pass, summary, request, body }` for runs or `{ skip: 'reason', body }` for skipped runs
- `STEP_ORDER` array driving Run All
- `wire()` that wires tab buttons, activates the default (Overview) tab, and calls `bootstrap()` on load so first-time visitors don't see a wall of "skipped · bootstrap session first" verdicts
- `mermaid.initialize({ startOnLoad: false })` — diagrams render lazily per tab via `renderMermaidIn()` (see §2; do not switch this back to `startOnLoad: true`)

### 5. Validate in a Browser

The demo HTML is embedded into the binary, so iterating on it has two modes — pick the one that matches what you're testing:

```
# Fast HTML iteration: open the file directly, no rebuild. The on-disk
# edits are reflected immediately; defaultApiBase() falls back to the
# running API. Use this while writing runners/markup.
#   file:///abs/path/internal/demo/release-{version}.html
#
# Embedded validation (the real gate): rebuild from branch source so the
# served copy matches what ships. A published-image run will NOT contain
# your edits — you must build from source.
```

Build from source and serve the embedded demo using the project's local-run/build command (see `agent_docs/build-and-test.md`).

Use the **browser-automation** skill to:
1. Open the demo — `file://` for fast iteration, or the served dev route after a from-source build for the embedded gate
2. Wait for the auto-bootstrap status pill to read "connected"
3. Click "Run all assertions"
4. Screenshot the final state and inspect every verdict
5. **Click through every tab** and confirm each `.mermaid` div rendered a real `<svg>` and not a "Syntax error" bomb. A diagram can parse cleanly yet still error at render if it was drawn while its tab was hidden — the lazy `renderMermaidIn()` path guards this, so a syntax-error SVG means the lazy-render wiring regressed, not that the diagram source is wrong.

If any step fails:
- **Script bug** (wrong assertion, bad selector, missing chained state, schema drift) — fix the HTML and re-run.
- **Code bug** — file an issue, note it in the section's prose, and either mark the runner as `skip` with a link OR leave it failing if the release is expected to fix it before merge.

Re-run until every step is `pass` or a knowing `skip`.

### 6. Commit

```bash
git add internal/demo/release-{version}.html
git commit -m "docs(demo): add interactive release demo for {version}"
git push
```

The embedded HTML ships with the binary on the next build. No release-asset upload, no separate hosting.

### 7. Reference in the Release PR

Add to the release PR description:

```markdown
## Live demo

After pulling this branch and running the project's local-run command,
open the served demo route, e.g. http://localhost/demo/release-{version}.html

The page auto-bootstraps an admin token and walks every API change as
a live round-trip. The frontmatter has a "Run all assertions" button
for the whole sweep, or click individual `Run` buttons per step.
Tabs across the top split the content by audience
(Overview / Demo / Frontend Changes / DevOps Info).

The file is also openable directly via `file://` (handy for sharing
by email) — it falls back to a localhost API base. Opening it from
another machine on the network works too: set the connection-panel
API base to the API host's address (not localhost) and the demo
rewrites any signed/redirect URLs to match.
```

Optionally drop a screenshot of the "Run all" pass state as a PR comment.

## Rules

- ALWAYS copy CSS + JS scaffolding from the most recent `release-*.html` — never re-derive
- ALWAYS use the tabbed layout (Overview / Demo / Frontend Changes / DevOps Info) — one audience per tab; keep a thin tab with a one-line "nothing here this release" rather than deleting it
- ALWAYS render mermaid lazily per tab (`startOnLoad: false` + `renderMermaidIn()` on tab activation) — `startOnLoad: true` renders diagrams in hidden zero-size panels and they fail at render despite parsing cleanly
- ALWAYS render request previews as copy-paste-runnable `curl` (via `formatRequest()`) with a copy button — never a raw fetch dump
- ALWAYS derive any rewritten signed/redirect URL host from `state.apiBase`, never a hardcoded `localhost` — so the demo works from a remote machine as well as same-machine
- ALWAYS put the exact env-var names the code reads (grepped, not recalled) in the DevOps tab, and file+line references with suggested edits in the Frontend tab
- ALWAYS verify endpoint shapes against the OpenAPI spec (e.g. `/swagger/doc.json`) before writing a runner
- ALWAYS use the project's dev-token endpoint — never hardcode credentials
- ALWAYS discover reference data (the IDs the page operates on) dynamically from the API at runtime
- ALWAYS auto-bootstrap on page load (`wire()` calls `bootstrap()`)
- ALWAYS retry once on 401 inside `http()` with a fresh dev token; long-lived demo pages are common
- ALWAYS keep the page openable via `file://` — `defaultApiBase()` falls back to a localhost API base when `window.location.protocol === 'file:'`
- ALWAYS validate by actually running the page in a browser (browser-automation skill) before committing — a green Run-All is the gate
- NEVER hardcode resource IDs — the page must work on any developer's local database
- NEVER generate shell scripts, recording tapes, or GIFs — those are deprecated
- NEVER upload anything to releases for the demo — the HTML embeds into the binary
- If a runner is structurally CORS-blocked (e.g. a direct probe to an internal service) or otherwise can't run from a browser, return `{ skip: 'reason' }` with an explanatory body — do not fail
- If the API is not running locally, report and stop — the demo cannot be validated without a live API
- Fix HTML/runner bugs immediately; file issues for code bugs surfaced by the demo
