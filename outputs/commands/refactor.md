---
description: Refactor the codebase following structured workflow with architecture design and review
argument-hint: "describe what improvements to make"
allowed-tools: ["*"]
---

# Refactor

$ARGUMENTS

## Workflow

### 1. Analyze Existing Codebase and Architecture (Architect Agent)

Analyze the current codebase to understand:
- Current structure and patterns
- Areas that need refactoring
- Potential improvements

### 2. Clean Up Unnecessary Code (Architect Agent)

Before planning implementation, clean up unnecessary code in areas where changes will be made:
- Remove dead code and unused imports
- Remove redundant logic
- Identify and document technical debt

### 3. Create Design and Plan Changes (Architect Agent)

Design the refactoring including:
- Architecture improvements
- Code structure changes
- Implementation plan with specific tasks
- Identify opportunities for parallel work

### 4. Get Design Review (Code Reviewer Agent)

Have the refactoring plan reviewed for:
- Soundness of approach
- Potential risks or issues
- Better alternatives

### 5. Confirm Plan with User

**IMPORTANT**: Present the design, whether backward compatibility is required or not, and plan to the user. Confirm the plan is good before proceeding. Do not start implementation until you get approval from the user.

### 6. Update Local Main Branch

Once you get approval, ensure the local main branch is the same as the remote main branch. If not, recreate the local main branch from the remote.

### 7. Set Up Development Environment

Based on the plan phases:

**If there are multiple phases:**
1. Create an epic PR to the default branch with an empty commit
2. For each phase, create a git worktree under `../worktrees`
3. Create sub PRs targeting the epic PR branch for each phase

**If there is a single phase:**
1. Create a git worktree under `../worktrees` for the implementation

The worktree names must include the ticket number provided.

### 8. Implementation (Software Engineer Agents)

Subagents make each change with review by other agents in the new worktrees. For each task, follow this process:

a. **Write Code with Tests**
   - The appropriate software engineer agent (Golang, TypeScript, Next.js, etc.) implements using Claude Code Skills
   - Implement refactoring following language/framework skill guidelines
   - Maintain existing functionality
   - Improve code structure and readability
   - Remove duplication (DRY principle)
   - Add or update tests as needed

b. **Review Changes**
   - A different agent reviews the refactoring
   - Validates improvements made
   - Ensures no functionality broken
   - Verifies test coverage

c. **Commit Changes**
   - Commit the incremental change before moving to next task

### 9. Create GitHub PR

Once all changes are completed in the worktrees, create a GitHub Pull Request. Then fix any CI errors until CI passes.

**IMPORTANT - CI Wait Times**: CI is slow and requires patience:
- Wait for at least 1 minute for CI jobs to start
- Wait for at least 5 minutes between checks for job completion
- Do not assume CI has failed if it hasn't started or completed yet

## Guidelines

- Clean up unnecessary code before refactoring
- Maintain existing functionality (tests must pass)
- Follow general coding guidelines (DRY, fail-fast, simplicity)
- Adhere to project-specific guidelines from `.claude/docs/guideline.md`
- Prefer breaking backward compatibility unless explicitly prohibited
- Keep refactoring focused and incremental
- Ensure all tests pass after refactoring
- Include ticket number in commit messages
- **CI takes time**: Wait at least 1 minute for CI to start, and at least 5 minutes between completion checks
