---
name: blast-radius
description: Lightweight impact analysis — trace imports, call chains, test coverage, and downstream consumers for any code target
user_invocable: true
---

# /blast-radius — Impact Analysis

Map the impact radius of a code change. Given a type, function, file, or issue number, trace all references, callers, tests, and downstream consumers to assess the scope of a modification.

## Invocation

```
/blast-radius MyStructName            # Analyze a type
/blast-radius HandleCreateUser        # Analyze a function
/blast-radius src/lib/auth.ts         # Analyze a file
/blast-radius #42                     # Extract targets from an issue
```

## Ground Rules

- **Read-only.** Never modify any files. Analysis only.
- **Depth-limited.** Call chain tracing stops at 3 levels deep to keep output actionable.
- **Deterministic.** Uses static analysis (grep, imports, references) — does not execute code.

## Step 1. Identify the Target

**If an issue number was provided:**
1. Fetch the issue using `gh issue view <number>`
2. Read the body for implementation notes, file references, type/function mentions
3. Extract all identifiable targets (types, functions, file paths)
4. Run Steps 2-5 for each target and combine into a single report

**If a file path was provided:**
1. Confirm the file exists
2. Extract the primary exports/public symbols from the file
3. Use the file itself as the target, plus its key exports

**If a type or function name was provided:**
1. Search the codebase to locate the definition
2. If multiple matches, list them and ask the user which one to analyze
3. Record the definition file and line number

Search patterns by language:
- Go: `type <name>`, `func <name>`, `func (.*) <name>`
- TypeScript/JS: `export.*<name>`, `class <name>`, `interface <name>`, `function <name>`
- Python: `class <name>`, `def <name>`
- Rust: `struct <name>`, `fn <name>`, `trait <name>`

## Step 2. Trace Imports

Find all files that import or reference the target:

- **Direct imports**: files that import the target's module/package
- **Re-exports**: files that re-export the target
- **Type references**: files that use the target as a type annotation

Search strategies by language:
- **Go**: search for `<package>.<Name>` across `.go` files
- **TypeScript/JS**: search for `import.*from.*<module>` and direct name usage
- **Python**: search for `from <module> import` and `import <module>`
- **Rust**: search for `use <crate>::<module>::<Name>`

Record each reference: `file:line — usage description`

## Step 3. Trace Call Chains (up to 3 levels)

Starting from the target, trace callers outward:

**Level 1 — Direct callers:**
- Functions/methods that call the target directly
- Event handlers or hooks that trigger the target

**Level 2 — Callers of callers:**
- For each Level 1 caller, find its callers
- Note the propagation path: `grandcaller -> caller -> target`

**Level 3 — Third-degree callers:**
- For each Level 2 caller, find its callers
- Stop here regardless of further depth

Use subagents for parallel tracing when there are many Level 1 callers.

Record each chain: `caller_chain -> target` with file locations.

## Step 4. Check Test Coverage

Identify tests that exercise the target:

1. Search test files for direct references to the target name
2. Search test files for references to the target's callers (Level 1)
3. Check for integration tests that exercise the target's API endpoints or commands

Classify coverage:
- **Direct tests**: tests that explicitly call/reference the target
- **Indirect tests**: tests that exercise the target through a caller
- **Missing coverage**: callers or code paths with no test coverage

## Step 5. Identify Downstream Consumers

Map package/module-level dependencies:

1. Identify which package/module the target belongs to
2. Find all other packages that depend on that package
3. For monorepos: check workspace dependency graphs
4. For libraries: check if the target is part of the public API

Record each consumer: `package — depends via [import path]`

## Output Format

Produce a structured report:

```
## Blast Radius: [target name]
**Definition:** [file:line]
**Package:** [package/module name]

### Direct References (N files)
| File | Line | Usage |
|------|------|-------|
| path/to/file.go | 42 | Calls target in request handler |
| path/to/other.ts | 18 | Uses target as type annotation |

### Call Chain (max depth: N)
caller_level3 (file:line)
  -> caller_level2 (file:line)
    -> caller_level1 (file:line)
      -> TARGET (file:line)

### Test Coverage
- **Direct tests:** N tests reference the target
  - test_file.go:TestFunctionName
- **Indirect tests:** N tests exercise the target via callers
  - integration_test.go:TestAPIEndpoint
- **At risk:** [callers or paths with no test coverage]

### Downstream Packages (N packages)
| Package | Dependency Path |
|---------|----------------|
| pkg/api | imports pkg/auth directly |
| pkg/cli | imports pkg/api which imports pkg/auth |

### Risk Assessment
**Rating: CONTAINED / MODERATE / WIDE**

Criteria:
- CONTAINED: <= 5 direct references, 1 package, good test coverage
- MODERATE: 6-20 direct references OR 2-3 packages OR partial test gaps
- WIDE: > 20 direct references OR 4+ packages OR significant test gaps

**Summary:** [1-2 sentence explanation of the rating]
```

## Rules

- **NEVER modify any files** — this is a read-only analysis command
- **NEVER execute project code** — use static analysis only (grep, file reading)
- **ALWAYS stop call chain tracing at depth 3** — deeper chains are noise
- **ALWAYS report the risk assessment** — CONTAINED, MODERATE, or WIDE
- **ALWAYS use subagents** for parallel tracing when analyzing multiple targets
- If the target cannot be found, report "Target not found" with search details and stop
- If an issue has no identifiable code targets, report this and suggest the user provide a specific type/function/file
