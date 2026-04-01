---
name: release-demo
description: Generate, run, and record a release E2E demo GIF for the release PR
user_invocable: true
---

# /release-demo — Release Demo & E2E Validation

Generate an E2E test script from closed milestone issues, run it, fix any failures, and record a looping VHS gif for the release PR.

## Invocation

```
/release-demo                    # Auto-detect release branch + milestone
/release-demo v0.9.26            # Target a specific version
```

## Steps

### 1. Detect Release Context

Determine the release version, milestone, and branch:

```bash
git branch --show-current    # expect release/*
```

If not on a release branch, ask the user for the version.

Fetch all closed issues in the milestone:
```bash
gh issue list --milestone "MILESTONE" --state closed --limit 200 --json number,title,body,labels
```

### 2. Generate the E2E Test Script

For each closed issue, read the body and extract:
- **Endpoint(s) changed** — from implementation notes or title scope
- **Acceptance criteria** — the testable checkboxes
- **Expected behavior** — what the API should return

Check the OpenAPI spec for endpoint signatures:
```bash
curl -s http://localhost/swagger/doc.json | jq '.paths["ENDPOINT"]'
```

Generate `scripts/test-release-{version}.sh` following this template:

```bash
#!/usr/bin/env bash
# Release {version} — End-to-End API Test
set -euo pipefail
BASE_URL="${BASE_URL:-http://localhost}"
```

**Template structure:**
1. **Pre-flight:** Connectivity check, admin token via `GET /api/dev/token/admin`, reference data discovery (programs, projects, contacts)
2. **Test groups:** One section per issue (or group of related issues), with:
   - Section header with issue number(s) and description
   - Narration explaining what changed and why
   - Numbered steps with curl calls and assertions
   - `pass`/`fail` helper calls for each assertion
3. **Cleanup:** Archive or revert any test data created
4. **Report:** Summary table of all pass/fail results with exit code

**Helpers to include** (copy from existing test scripts):
- `pass()` / `fail()` — colored output with result tracking
- `section()` / `narrate()` — formatted headers and descriptions
- `json_field()` / `json_excerpt()` / `json_block()` — JSON parsing via python
- Pass/fail counters and final report

### 3. Run the Test Script

```bash
bash scripts/test-release-{version}.sh
```

If any tests fail:
1. Diagnose the failure from the output
2. Determine if the failure is a **script bug** (wrong assertion, bad test data) or a **code bug**
3. Fix script bugs directly. For code bugs, create an issue and note it.
4. Re-run until all tests pass or only known code bugs remain

### 4. Generate VHS Tape File

Create `scripts/release-demo-{version}.tape`:

```
# Release Demo: {version}

Output scripts/release-demo-{version}.gif

Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Dracula"
Set Padding 20
Set PlaybackSpeed 0.1
Set LoopOffset 0%

Type@0 "bash scripts/test-release-{version}.sh"
Enter
Sleep {measured_duration + 5s buffer}
```

**PlaybackSpeed 0.1** for readable output. **LoopOffset 0%** for continuous looping.

To measure duration:
```bash
START=$(date +%s); bash scripts/test-release-{version}.sh; echo $(($(date +%s) - START))
```

### 5. Record the GIF

```bash
vhs scripts/release-demo-{version}.tape
```

If VHS fails with sandbox errors, retry with:
```bash
VHS_NO_SANDBOX=true vhs scripts/release-demo-{version}.tape
```

Verify the gif was created:
```bash
ls -lh scripts/release-demo-{version}.gif
```

### 6. Commit Assets

Add the test script and tape file to the release branch:

```bash
git add scripts/test-release-{version}.sh scripts/release-demo-{version}.tape
git commit -m "chore: add release demo script and VHS tape for {version}"
git push
```

Do NOT commit the gif to the repo — it will be uploaded as a release asset.

### 7. Upload GIF

Create or update a GitHub release to host the gif:

```bash
gh release create v{version} --target RELEASE_BRANCH --title "Release {version}" --notes "See release PR for details." scripts/release-demo-{version}.gif
```

Report the gif URL for embedding in the PR description:
```
![Release Demo](https://github.com/OWNER/REPO/releases/download/v{version}/release-demo-{version}.gif)
```

## Rules

- ALWAYS generate the test script before recording — never record a blank or failing run
- ALWAYS use PlaybackSpeed 0.1 and LoopOffset 0% for readable, looping output
- ALWAYS use the admin dev token (`GET /api/dev/token/admin`) — never hardcode credentials
- ALWAYS discover reference data (program IDs, project IDs, contact IDs) dynamically from the API
- NEVER hardcode UUIDs — the script must work on any developer's local database
- NEVER commit the gif to the repo — upload it as a release asset
- If the API is not running, report and stop — cannot generate demo without live API
- Fix script bugs immediately; create issues for code bugs
