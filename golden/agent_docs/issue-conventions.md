# Issue Conventions

> Reference document. Loaded by workflow commands when creating or editing issues.
> Not read every session — see CLAUDE.md for behavioral instructions.

## Issue Title Format

```
{type}({scope}): {description}

Types: feat | fix | refactor | docs | discovery | design | infra
Scopes: defined per-project (see CLAUDE.md Issue Scopes)
```

## Issue Body Structure

```markdown
## Summary
[1-3 sentences describing what and why]

## Dependencies
- Blocked by: #NN — [reason]
- Part of: #EPIC — [epic title]

## Acceptance Criteria
- [ ] [Specific, testable criterion]
- [ ] [Another criterion]
- [ ] Validation passes

## Implementation Notes
[Key files, approach, constraints]
```

## Dependency Format

**Canonical format — the ONLY recognized format:**

```
- Blocked by: #NN — reason
```

`#NN` uses the configured issue reference format (see `agent_docs/issue-tracker-ops.md`). Other patterns (`depends on`, `waiting on`, `after #NN`) are **NOT** recognized by automation.

## Planner-Enriched Issues

When the Planner agent enriches an issue, it appends these sections to the body:

```markdown
## Spec (Planner-generated)
### Schema Changes
[Tables, columns, migrations]
### API Changes
[Endpoints, request/response shapes]
### Implementation Hints
[Key files, function signatures, order of operations]
### Known Pitfalls
[Bug patterns from memory, spec gaps from past features]
### Estimated Effort
[Hours estimate informed by calibration memory]
```

These sections are additive — they do not replace the original issue body.
