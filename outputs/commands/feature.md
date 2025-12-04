---
description: Add or update a feature following a structured workflow with architecture design and review
argument-hint: "describe the feature to add or update"
allowed-tools: ["*"]
---

# Feature

$ARGUMENTS

## Workflow

### 1. Analyze Existing Codebase and Architecture (Architect Agent)

Analyze the current codebase structure, patterns, and architecture to understand the context for the new feature.

### 2. Create Design and Plan Changes (Architect Agent)

Design the new feature including:
- Architecture changes required
- API designs (if applicable)
- Data models (if applicable)
- Implementation plan with specific tasks
- Opportunities for parallel development

### 3. Get Design Review (Code Reviewer Agent)

Have the design and plan reviewed for:
- Architectural soundness
- Alignment with existing patterns
- Potential issues or improvements

### 4. Confirm Plan with User

**IMPORTANT**: Present the design, whether backward compatibility is required or not, and plan to the user. Confirm the plan is good before proceeding. Do not start implementation until you get approval from the user.

### 5. Update Local Main Branch

Once you get approval, ensure the local main branch is the same as the remote main branch. If not, recreate the local main branch from the remote.

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
   - Implement feature code and corresponding tests
   - Maintain consistency with existing patterns
   - Keep changes simple and focused

b. **Review Changes**
   - A different agent reviews the implementation
   - Validates code quality and standards
   - Suggests improvements if needed
   - Ensures tests are adequate

c. **Commit Changes**
   - Commit the incremental change before moving to next task

### 8. Create GitHub PR

Once all changes are completed, create a GitHub Pull Request. Then fix any CI errors until CI passes.

**IMPORTANT - CI Wait Times**: CI is slow, so you must wait appropriately:
- Wait for at least 1 minute for CI jobs to start
- Wait for at least 5 minutes between checks for CI job completion
- Do not assume CI has finished without waiting and checking the actual results

## Guidelines

- Follow general coding guidelines (DRY, fail-fast, simplicity)
- Adhere to project-specific guidelines from `.claude/docs/guideline.md`
- Use appropriate language/framework skills
- Maintain test coverage
- Prefer breaking backward compatibility unless explicitly prohibited
- Include ticket number in commit messages
- When waiting for CI: wait at least 1 minute for jobs to start, and at least 5 minutes between completion checks
