# Issue Tracker Operations

> Reference document. Loaded by workflow commands when executing issue tracker operations.
> Not read every session — see CLAUDE.md for behavioral instructions.

## Tracker Configuration

**Tool:** GitHub CLI (`gh`)
**Issue reference format:** `#NN` (e.g., `#53`)
**Smart close syntax:** `Closes #NN`
**Dependency format:** `- Blocked by: #NN — reason`

> Configured by `/bootstrap`. Default: GitHub Issues (`gh` CLI).
> To use a different tracker (Jira, Linear, GitLab), `/bootstrap`
> will update this file and the corresponding permissions in settings.local.json.

## Operations Reference

| Operation | Command |
|-----------|---------|
| List open issues | `gh issue list --state open --limit 200 --json number,title,body,labels,milestone,assignees` |
| View issue | `gh issue view NUMBER --json number,title,body,labels,milestone,assignees` |
| View issue body | `gh issue view NUMBER --json body --jq '.body'` |
| Create issue | `gh issue create --title "TITLE" --body "BODY" --label "LABEL" [--assignee "USER"]` |
| Edit issue body | `gh issue edit NUMBER --body "BODY"` |
| Close issue | `gh issue close NUMBER` |
| Comment on issue | `gh issue comment NUMBER --body "COMMENT"` |
| Add label | `gh issue edit NUMBER --add-label "LABEL"` |
| Remove label | `gh issue edit NUMBER --remove-label "LABEL"` |
| Assign to milestone | `gh issue edit NUMBER --milestone "NAME"` |
| Create milestone | `gh api repos/{owner}/{repo}/milestones --method POST -f title="NAME" -f state="open"` |
| List milestones | `gh api repos/{owner}/{repo}/milestones` |
| Check milestone progress | `gh api repos/{owner}/{repo}/milestones/{number}` |
| Resolve current user | `gh api user --jq '.login'` |
| List collaborators | `gh api repos/{owner}/{repo}/collaborators --jq '.[].login'` |
