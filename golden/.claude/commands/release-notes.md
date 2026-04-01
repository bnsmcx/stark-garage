---
name: release-notes
description: Generate full narrative release PR description from closed milestone issues
user_invocable: true
---

# /release-notes — Release PR Description Generator

Generate a comprehensive, user-facing release PR description from the closed issues in a milestone. Produces a narrative "What's New" section, endpoint/API change tables, database change tables, and implementation progress.

## Invocation

```
/release-notes                    # Auto-detect release branch + milestone
/release-notes v0.9.26            # Target a specific version
```

## Steps

### 1. Detect Release Context

Determine the release version, milestone, release branch, and release PR:

```bash
git branch --show-current
gh pr list --base main --head RELEASE_BRANCH --state open --json number --jq '.[0].number'
```

Fetch all closed issues in the milestone:
```bash
gh issue list --milestone "MILESTONE" --state closed --limit 200 --json number,title,body,labels
```

### 2. Categorize Issues

Group closed issues by type (from title prefix):
- **Features** (`feat(`)
- **Fixes** (`fix(`)
- **Documentation** (`docs(`)
- **Refactors** (`refactor(`)
- **Tracking epics** (`tracking:`)
- **Other**

Separate implementation issues from tracking epics.

### 3. Generate "What's New" Narrative

For each feature group (not individual issues), write a paragraph explaining:
- What changed from the user's perspective
- Which endpoints are affected
- Why it matters (the user problem it solves)

Use the issue bodies for technical detail but write for a technical audience who hasn't read the issues. Bold the feature name at the start of each paragraph.

Include a separate paragraph for:
- Bug fixes (group related fixes)
- Security fixes (dependency upgrades, vulnerability patches)
- Breaking changes (if any — call these out prominently)

### 4. Build API Changes Table

If the release touches HTTP endpoints, generate:

**Modified Endpoints:**

| Endpoint | Change |
|----------|--------|
| `METHOD /path` | Description of change |

**New Endpoints (if any):**

| Method | Path | Auth | Description |
|--------|------|------|-------------|

Extract endpoint info from issue bodies (implementation notes), swagger diff, or handler file changes.

### 5. Build Database Changes Table

If the release touches schema or models:

| Change | Table | Details |
|--------|-------|---------|
| Type of change | Table name | What changed |

Include dependency upgrades here too.

### 6. Build Implementation Progress

Generate a collapsed `<details>` section with phased implementation progress:

```markdown
<details>
<summary><strong>Implementation Progress (N/N issues complete)</strong></summary>

#### Phase 1: ...
- [x] #NN -- description (PR #MM)

#### Phase 2: ...
...

</details>
```

Include tracking epics at the bottom.

### 7. Assemble PR Description

Combine all sections into the final PR body:

```markdown
## Release: {milestone-name}

> **Note:** [Any prerequisite instructions, e.g., DB rebuild required]

**N issues, M epics, K PRs** -- [one-line summary of themes]

### What's New

[Generated narrative paragraphs]

### Modified Endpoints

[Endpoint table]

### Database Changes

[Database table]

### Release Demo & Testing

**E2E demo script:** [`scripts/test-release-{version}.sh`](scripts/test-release-{version}.sh) -- N assertions across M test groups.

![Release Demo](https://github.com/OWNER/REPO/releases/download/v{version}/release-demo-{version}.gif)

**N/N E2E tests passed** -- [summary of what was tested]

**Additional testing:**
- **Unit tests:** [summary]
- **CI:** `make ci-checks` passes all gates

---

[Implementation progress details section]

### Tracking Epics
- #NN -- title
```

### 8. Apply to PR

Update the release PR description:
```bash
gh pr edit PR_NUMBER --body-file /tmp/release-notes.md
```

## Rules

- ALWAYS write for a technical audience who hasn't read the individual issues
- ALWAYS include the demo gif embed (generate with `/release-demo` first if needed)
- ALWAYS use the `<details>` collapsed section for implementation progress
- ALWAYS count PRs, issues, and epics accurately
- NEVER include internal implementation details in the "What's New" section — focus on user-visible changes
- NEVER use bare issue numbers without context — always include the title or description
- If the demo gif URL doesn't exist yet, leave a placeholder and note it
