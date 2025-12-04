---
description: Split a large PR into a parent PR with multiple child PRs for easier review
argument-hint: "PR number or URL to split"
allowed-tools: ["Bash", "Read", "Write", "Edit", "Glob", "Grep", "Task", "WebFetch", "TodoWrite"]
---

# Split PR

Split a large Pull Request into a parent PR with multiple child PRs for easier code review.

## Prerequisites

**Required MCP Servers:**
- None required, but ensure GitHub CLI (`gh`) is installed and authenticated

**Required Tools:**
- GitHub CLI (`gh`) - Install via: `brew install gh` or `sudo apt install gh`
- Authenticate with: `gh auth login`

## Arguments

- **$1**: PR number or URL to split (e.g., `123` or `https://github.com/owner/repo/pull/123`)

## Workflow

### 1. Analyze the Original PR

First, gather information about the PR to be split:

```bash
# Get PR details
gh pr view $1 --json number,title,body,baseRefName,headRefName,commits,files

# Get the list of commits
gh pr view $1 --json commits --jq '.commits[] | "\(.oid) \(.messageHeadline)"'

# Get the list of changed files
gh pr view $1 --json files --jq '.files[].path'
```

Analyze:
- The PR title and description
- The base branch (target branch)
- The commits and their logical groupings
- The files changed and their relationships

### 2. Plan the Split Strategy

Based on the analysis, determine how to split the PR:

**Option A: Split by commits** (preferred when commits are clean and logical)
- Group related commits together
- Each child PR contains a set of meaningful commits

**Option B: Split by files** (when commits are noisy or mixed)
- Group related files together
- Each child PR focuses on a specific area/feature

**Option C: Hybrid approach**
- Use commits where they make sense
- Cherry-pick specific changes for cleaner separation

Present the split plan to the user with:
- Number of child PRs proposed
- Contents of each child PR (commits or files)
- Order of child PRs (dependencies)
- Estimated review complexity for each

**IMPORTANT**: Wait for user approval before proceeding.

### 3. Create the Parent PR

Once the user approves the plan:

1. **Ensure local main branch is up to date:**
   ```bash
   git fetch origin
   git checkout main
   git reset --hard origin/main
   ```

2. **Create the parent branch with an empty commit:**
   ```bash
   git checkout -b epic/$1-split
   git commit --allow-empty -m "Epic: Split of PR #$1

   This is a parent PR for splitting #$1 into smaller, reviewable chunks.

   Child PRs will be listed below once created."
   git push -u origin epic/$1-split
   ```

3. **Create the parent PR:**
   ```bash
   gh pr create --title "Epic: [Original PR Title] (Split)" \
     --body "$(cat <<'EOF'
   ## Summary

   This is a parent PR for splitting #$1 into smaller, reviewable chunks.

   **Original PR:** #$1

   ## Child PRs

   _Child PRs will be added here as they are created._

   ## Review Order

   Please review child PRs in the order listed above.

   ---
   This PR should be merged after all child PRs are merged.
   EOF
   )" \
     --base main
   ```

4. **Record the parent PR number** for use in child PRs.

### 4. Create Child PRs

For each planned child PR:

1. **Create a worktree for the child PR:**
   ```bash
   mkdir -p ../worktrees
   git worktree add ../worktrees/split-$1-child-N epic/$1-split
   cd ../worktrees/split-$1-child-N
   git checkout -b split/$1-child-N
   ```

2. **Apply changes based on split strategy:**

   **If splitting by commits:**
   ```bash
   # Cherry-pick the relevant commits
   git cherry-pick <commit-hash-1> <commit-hash-2> ...
   ```

   **If splitting by files:**
   ```bash
   # Checkout specific files from the original PR branch
   git checkout origin/<original-pr-branch> -- path/to/file1 path/to/file2
   git add .
   git commit -m "Child N: <description of changes>"
   ```

3. **Push and create the child PR:**
   ```bash
   git push -u origin split/$1-child-N
   gh pr create --title "Child N/M: <descriptive title>" \
     --body "$(cat <<'EOF'
   ## Summary

   <Brief description of what this child PR contains>

   ## Parent PR

   Part of epic PR #<parent-pr-number>

   ## Changes

   <List of changes in this child PR>

   ## Dependencies

   <Any child PRs that must be merged before this one, if applicable>
   EOF
   )" \
     --base epic/$1-split
   ```

4. **Repeat for all child PRs.**

### 5. Update Parent PR Description

After all child PRs are created, update the parent PR with the list of child PRs:

```bash
gh pr edit <parent-pr-number> --body "$(cat <<'EOF'
## Summary

This is a parent PR for splitting #$1 into smaller, reviewable chunks.

**Original PR:** #$1

## Child PRs

1. #<child-1-number> - <title>
2. #<child-2-number> - <title>
3. #<child-3-number> - <title>
...

## Review Order

Please review child PRs in the order listed above.

---
This PR should be merged after all child PRs are merged.
EOF
)"
```

### 6. Provide Summary to User

Present to the user:
- Parent PR number and URL
- List of all child PR numbers and URLs
- Recommended review order
- Instructions for merging (merge child PRs first, then parent PR)

## Guidelines

- Always analyze the original PR thoroughly before planning the split
- Ensure each child PR is independently reviewable and testable
- Maintain logical groupings - don't split related changes across multiple PRs
- Keep child PRs focused on a single concern when possible
- Preserve commit messages and authorship when cherry-picking
- Document dependencies between child PRs clearly
- The parent PR should remain empty (or near-empty) - all code goes in child PRs
- After all child PRs are merged into the parent branch, the parent PR merges to main

## Example

For PR `#456`:

**Parent PR:**
- Branch: `epic/456-split`
- PR: `#500` - "Epic: Add user authentication (Split)"

**Child PRs:**
- `#501` - "Child 1/3: Add user model and database migrations" (base: `epic/456-split`)
- `#502` - "Child 2/3: Implement authentication service" (base: `epic/456-split`)
- `#503` - "Child 3/3: Add login/logout UI components" (base: `epic/456-split`)

**Merge order:**
1. Merge `#501` into `epic/456-split`
2. Merge `#502` into `epic/456-split`
3. Merge `#503` into `epic/456-split`
4. Merge `#500` into `main`
