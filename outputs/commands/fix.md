---
description: Fix a bug by reproducing the error, understanding root cause, and planning fixes
argument-hint: "describe the error and what fix is needed"
allowed-tools: ["*"]
---

# Fix

$ARGUMENTS

## Workflow

### 1. Analyze Existing Codebase (Architect Agent)

Analyze the codebase to understand where the error happens:
- Identify affected components
- Understand code flow leading to the error
- Map out relevant parts of the codebase

### 2. Reproduce Error and Find Root Cause (Software Engineer Agent)

A software engineer with the appropriate tech stack must:
- Reproduce the error to confirm the issue
- Understand the root cause of the error
- Document findings and analysis

### 3. Plan Changes (Architect Agent)

Based on the analysis and root cause:
- Plan the necessary fixes
- Identify opportunities for parallel changes
- Ensure the fix addresses the root cause, not just symptoms

### 4. Confirm Plan with User

**IMPORTANT**: Present the analysis, root cause findings, and planned fixes to the user. Wait for user approval before proceeding with any implementation.

### 5. Update Local Default Branch

Once you get approval, ensure the local default branch is the same as the remote one. If not, recreate the local default branch from the remote.

### 6. Set Up Development Environment

Based on the plan phases:

**If there are multiple phases:**
1. Create an epic PR to the default branch with an empty commit
2. For each phase, create a git worktree under `../worktrees`
3. Create sub PRs targeting the epic PR branch for each phase

**If there is a single phase:**
1. Create a git worktree under `../worktrees` for the implementation

The worktree names must include the ticket number you provided.

### 7. Implementation (Software Engineer Agents)

Subagents must make each change with review by other agents in the new worktrees. For each task, follow this process:

a. **Write Code with Tests**
   - The appropriate software engineer agent (Golang, TypeScript, Next.js, etc.) implements using Claude Code Skills
   - Follow language/framework skill guidelines
   - Implement fix and corresponding tests
   - Maintain consistency with existing patterns
   - Keep changes simple and focused

b. **Review Changes**
   - A different agent reviews the implementation
   - Validates code quality and standards
   - Suggests improvements if needed
   - Ensures tests are adequate

c. **Commit Changes**
   - Commit the incremental change before moving to next task

### 8. Create GitHub PR and Fix CI Errors

Once all changes are completed in the worktrees, create a GitHub Pull Request.

**IMPORTANT - CI Error Handling:**
- Fix any CI errors until CI passes
- CI is slow, so wait appropriately:
  - Wait for at least a minute for a job to start
  - Wait for at least every 5 minutes between checks to allow jobs to complete
- Do not rush CI checks - be patient and thorough

## Guidelines

- Always reproduce the error first before attempting fixes
- Understand root cause before fixing
- Follow general coding guidelines (DRY, fail-fast, simplicity)
- Adhere to project-specific guidelines from `.claude/docs/guideline.md`
- Use appropriate language/framework skills
- Maintain test coverage
- Include ticket number in commit messages
- Be patient with CI - wait for jobs to start and complete before re-checking
