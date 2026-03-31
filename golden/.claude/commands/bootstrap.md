---
name: bootstrap
description: Scan the current project and adapt the full Agentic Engineering Toolbox to its tech stack, conventions, and structure
user_invocable: true
---

# /bootstrap — Project Configuration Adapter

Scans the current project, builds a profile of its tech stack and conventions, confirms findings with the user, and adapts the full Agentic Engineering Toolbox (all commands, all agents, memory) to this specific project.

## When to Use

Run this after deploying the golden set into a new project:
```
./deploy.sh /path/to/project
cd /path/to/project
claude
> /bootstrap
```

Can also be re-run to update configuration after significant project changes.

## Phase 1: Discovery

Scan the project to build a comprehensive profile. Check for ALL of the following:

### Tech Stack Detection

| File/Pattern | Indicates |
|-------------|-----------|
| `package.json` | Node.js/JavaScript/TypeScript project |
| `tsconfig.json` | TypeScript |
| `go.mod` | Go |
| `Cargo.toml` | Rust |
| `requirements.txt` / `pyproject.toml` / `setup.py` | Python |
| `Gemfile` | Ruby |
| `pom.xml` / `build.gradle` | Java/Kotlin |
| `Makefile` | Make-based build |
| `Dockerfile` / `docker-compose.yml` | Docker |
| `.csproj` / `.sln` | .NET |

### Framework Detection

| File/Pattern | Indicates |
|-------------|-----------|
| `next.config.*` | Next.js |
| `vite.config.*` | Vite |
| `angular.json` | Angular |
| `svelte.config.*` | SvelteKit |
| `nuxt.config.*` | Nuxt |
| `remix.config.*` | Remix |
| `expo-*` in package.json | Expo / React Native |
| `tailwind.config.*` | Tailwind CSS |
| `flask` / `django` in requirements | Flask/Django |
| `gin` / `echo` / `fiber` in go.mod | Go web frameworks |
| `actix` / `axum` in Cargo.toml | Rust web frameworks |

### Build/Test Tooling

Detect test runners, linters, formatters, build commands by checking:
- `package.json` scripts (test, lint, build, validate, format, type-check)
- `Makefile` targets (test, lint, build, validate)
- `pyproject.toml` tool configs (pytest, ruff, black, mypy)
- `.eslintrc*`, `.prettierrc*`, `biome.json`
- `jest.config.*`, `vitest.config.*`, `pytest.ini`
- `golangci-lint` config, `.golangci.yml`

**Identify the validation command:** Look for a single command that runs all checks:
- `npm run validate` / `make validate` / `make check`
- Fallback: compose from individual commands (test + lint + type-check)

### CI/CD Detection

Check for: `.github/workflows/*.yml`, `Jenkinsfile`, `.gitlab-ci.yml`, `.circleci/config.yml`, `Dockerfile`

### Project Structure

Identify the primary architecture pattern:
- **Monorepo:** `packages/`, `apps/`, `libs/`, or workspaces in package.json
- **API + Frontend:** Separate api/ and web/ directories
- **Library:** Single package with src/ and tests/
- **CLI tool:** bin/ or cli/ with argument parsing
- **Desktop app:** Electron, Tauri indicators

### Issue Tracker Detection

| Signal | Indicates |
|--------|-----------|
| `.github/` + `gh` available | GitHub Issues (default) |
| `.jira.d/`, `JIRA_*` env vars | Jira |
| `.linear` config | Linear |
| `.gitlab-ci.yml`, `GITLAB_*` env vars | GitLab Issues |

### Git State

- Is this a git repo? Default branch? Existing branches?
- Does `.gitignore` exist?

### Existing Claude Config

- Does `CLAUDE.md` already have content below the bootstrap marker?
- Are there existing project-specific commands? (re-bootstrap warning)

## Phase 2: Confirm with User

Present the discovery results and ask questions.

### Question 1: Project Profile

Present the detected profile and let the user correct anything:

```
## Detected Project Profile
**Tech Stack:** [languages and frameworks detected]
**Build System:** [build tool and key commands]
**Test Runner:** [test framework and command]
**Validation Command:** [detected or composed validation command]
**Linter/Formatter:** [detected tools]
**CI/CD:** [detected system]
**Architecture:** [detected pattern]
```

### Question 2: Git Integration

Ask: **"Should Claude configuration files be checked into git or gitignored?"**
- **Check into git** — Team shares Claude config
- **Add to .gitignore** — Personal config, not shared

### Question 3: Issue Scopes

Ask: **"What scopes should be used for issue titles and commit messages?"**

Suggest scopes based on detected architecture (package names for monorepos, `api`/`web` for API+Frontend, etc.). Let the user adjust.

### Question 4: Task Tracking Mode

Ask: **"How should this project track tasks?"**
- **External issue tracker** (default) — Full workflow command support
- **In-repo task file** (`tasks/todo.md`) — Lightweight tracking for solo projects

## Phase 3: Adapt

Based on discovery + user answers, make the following changes:

### 3.1 Append to CLAUDE.md

Find the `<!-- bootstrap: project-specific below -->` marker. Append below it:

```markdown
## Project Overview
**Project:** [name]
**Architecture:** [detected pattern]
**Tech Stack:** [confirmed stack]

## How to Build / Test / Run
**Validation command:** `[command]` — hard gate for all workflow commands.
For all build/test/run commands, see `agent_docs/build-and-test.md`.

## Issue Scopes
- `[scope1]`: [what it covers]

## Architecture Rules
[Based on detected architecture — e.g., process boundary rules, SDK enforcement]

## Key Files
For project structure and key files, see `agent_docs/project-structure.md`.
```

### 3.2 Create project-specific agent_docs

**`agent_docs/build-and-test.md`:** Table of detected build/test/lint/run commands.

**`agent_docs/project-structure.md`:** Directory structure and key files reference.

### 3.3 Add Permissions to settings.local.json

Merge project-specific permissions into the existing `allow` array based on detected tech stack (e.g., `Bash(go build:*)` for Go, `Bash(cargo test:*)` for Rust, `Bash(docker:*)` for Docker).

### 3.4 Create Project-Specific Commands

Based on the detected primary abstraction:
- **API projects:** `.claude/commands/add-endpoint.md`
- **React/frontend:** `.claude/commands/add-component.md`
- **Pipeline/data:** `.claude/commands/add-pipeline-step.md`
- **All projects with docs/:** `.claude/commands/update-docs.md`

### 3.5 Augment Code Reviewer

Find the `<!-- bootstrap: project-specific checks below -->` marker in `.claude/agents/code-reviewer.md`. Append architecture-specific review criteria.

### 3.6 Configure .mcp.json

Add project-relevant MCP servers. Always include:
- **Playwright MCP** — for browser automation and E2E testing
- **Chrome DevTools MCP** — for runtime debugging and inspection

Only add additional servers that clearly match the project's needs.

### 3.7 Configure .gitignore

If the user chose to gitignore Claude config, append the relevant entries.

### 3.8 Settings.json Hooks

Create or update `.claude/settings.json` with PostToolUse hooks for detected linters/formatters (ESLint for TypeScript, ruff for Python, etc.) and PreToolUse hooks to block editing sensitive files (.env).

### 3.9 Configure Issue Tracker

For non-GitHub trackers (Jira, Linear, GitLab): update `agent_docs/issue-tracker-ops.md` with the appropriate CLI equivalents and add permissions.

### 3.10 Configure Task Tracking Mode

If in-repo task file selected: create `tasks/todo.md`, modify CLAUDE.md references.

### 3.11 Initialize Memory Database

Initialize the toolbox memory system:

```bash
toolbox-memory init
```

If the `toolbox-memory` CLI is not available, create an empty memory database placeholder. This enables `/slim` memory pruning and cross-session knowledge retention.

### 3.12 Post-bootstrap budget check

1. Count CLAUDE.md total lines (baseline + project-specific)
2. Check against combined budget from `BUDGETS.md`: baseline + project max
3. If over budget, identify sections to relocate to `agent_docs/`
4. Report: "CLAUDE.md: NN/140 lines (NN%). Budget: healthy / warning / exceeded."

## Phase 4: Summary

```
## Bootstrap Complete!

### Changes Made:
- CLAUDE.md: Appended project-specific configuration
- agent_docs/: Created project reference docs
- settings.local.json: Added [N] project-specific permissions
- Commands created: [list]
- Code reviewer: Augmented with [project-type] checks
- MCP servers: Playwright MCP, Chrome DevTools MCP [+ others]
- Memory database: Initialized
- [.gitignore updated / docs/ scaffold / hooks — if applicable]

### What's Configured:
- Validation command: `[command]`
- Test command: `[command]`
- Issue scopes: [list]
- CLAUDE.md: NN/140 lines (NN%). Budget: [healthy / warning / exceeded]

### Next Steps:
1. Review the changes — especially CLAUDE.md and the generated commands
2. Adjust anything that doesn't look right
3. Start working! Use `/triage` to analyze your backlog or `/create-issues` to plan new work.
```

## Rules

- ALWAYS scan before asking — present findings, don't ask the user to describe their project
- ALWAYS confirm with user before making changes
- NEVER overwrite baseline sections of CLAUDE.md — only append below the marker
- NEVER remove baseline permissions from settings.local.json — only add
- If re-bootstrapping (existing project-specific content detected), warn user and ask before overwriting
- Keep generated commands practical — don't create commands for patterns the project doesn't use
- Prefer detecting over guessing — if you can't determine something from the project files, ask
