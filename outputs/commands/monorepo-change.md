---
description: Apply changes across monorepo subprojects with planning, implementation, and review
argument-hint: [change description]
---

# Monorepo Change Request

$ARGUMENTS

---

## Instructions

### Step 1: Understand Monorepo Structure

1. Analyze the monorepo structure if not already familiar
2. Identify subprojects and their relationships
3. Common monorepo patterns:
   - Language-specific directories (e.g., `go/`, `node/`, `python/`)
   - Proto/schema directories that generate code for other subprojects
   - Shared libraries or packages
   - Multiple applications/services

### Step 2: Plan the Changes

1. **Analyze the change request** and determine:
   - Type of change: feature, bugfix, refactoring, configuration, dependency update, etc.
   - Which subprojects are affected
   - Dependencies between subprojects (e.g., proto changes affect multiple consumers)

2. **Create a detailed plan**:
   - Break down tasks by subproject
   - Identify execution order (sequential vs parallel)
   - **Determine appropriate Claude Code Skills** to use for each subproject:
     - `golang` skill for Go projects
     - `nextjs` skill for Next.js applications
     - `typescript` skill for TypeScript projects
     - `protobuf` skill for Protocol Buffer files
     - `github-actions` skill for GitHub Actions workflows
   - Consider testing strategy for each subproject

3. **Use TodoWrite** to create a task list with clear phases

4. **Present the plan** to the user and wait for approval before proceeding

### Step 3: Implement Changes

Execute changes respecting dependency order and using appropriate patterns:

#### Phase 1: Foundation Changes (if needed)

Handle changes that other subprojects depend on first:
- **Schemas/Protos**: Update `.proto` files, regenerate code
- **Shared libraries**: Update shared packages or modules
- **Configuration**: Update config files that affect multiple projects

For each foundation change:
- Invoke appropriate skill using the Skill tool (e.g., `protobuf` for .proto files)
- Complete the changes
- Regenerate or rebuild dependent code
- Mark the todo as completed

#### Phase 2: Dependent Projects

After foundation changes are complete, update dependent subprojects:

**Parallel Execution Pattern:**
- When multiple subprojects can be updated independently, launch Task agents **in parallel**
- **CRITICAL**: Launch all parallel agents in a single message with multiple Task tool calls
- Each agent should invoke appropriate skills for their language/framework

**Sequential Execution Pattern:**
- When changes must be made in a specific order, complete them sequentially
- Update TodoWrite as each step completes

**For each subproject Task agent, provide detailed instructions to:**

**CRITICAL: Each agent MUST ensure all quality checks pass before returning.**

1. Invoke appropriate skill(s) using the Skill tool:
   - `golang` for Go projects - skill ensures tests and lints pass
   - `typescript` for TypeScript projects - skill ensures tests and lints pass
   - `nextjs` for Next.js applications - skill ensures tests and build pass
   - `protobuf` for Protocol Buffer files
   - Others as applicable

2. Implement the change:
   - Features: Add new functionality with tests
   - Bugfixes: Fix the issue and add regression tests
   - Refactoring: Improve code structure while maintaining behavior
   - Configuration: Update settings, build configs, or dependencies
   - Dependencies: Update package versions and handle breaking changes

3. **MANDATORY: Fix all issues until checks pass**

   The skill will guide the agent through verification, but the agent MUST:
   - Run all quality checks (tests, lints, builds, type checks)
   - Fix any failures immediately
   - Iterate until ALL checks pass
   - Update tests if implementing new features or fixing bugs

   **DO NOT return until:**
   - ✅ All tests pass
   - ✅ All lints pass
   - ✅ Build succeeds
   - ✅ Type checking passes (for typed languages)
   - ✅ No errors or warnings

4. Return a summary of changes made and verification results
   - List files modified
   - Confirm all checks passed
   - Note any issues encountered and how they were resolved

**Key Points:**
- Use parallel Task agents when possible to improve efficiency
- Each agent handles implementation details for their domain
- Agents run verification independently
- Wait for all agents to complete before proceeding
- Update TodoWrite as tasks complete

### Step 4: Integration and Review

After all subproject changes are complete:

1. **Verify all quality checks passed** in each subproject:
   - Confirm all tests passed
   - Confirm all lints passed
   - Confirm all builds succeeded
   - Confirm type checking passed
   - If any failed, go back and fix them

2. **Verify integration** between changed subprojects:
   - Do API contracts match between services?
   - Are shared types/schemas consistent?
   - Do configuration changes work across projects?

3. **Review code quality**:
   - Launch Task agents to review changes if needed
   - Check for security issues
   - Verify test coverage
   - Ensure documentation is updated

4. **Run integration tests** if available:
   - End-to-end tests
   - Integration test suites
   - Manual testing instructions for the user

5. **Provide a final summary**:
   - List all changed files by subproject
   - **Explicitly confirm all quality checks passed** (tests, lints, builds)
   - Highlight important changes or breaking changes
   - Note any issues encountered and how they were resolved
   - Note any manual steps needed
   - Suggest next steps (e.g., creating a PR, deploying)

---

## Examples

### Example 1: Adding a Feature
```
/monorepo-change Add user profile photo support with upload and display
```

Workflow:
1. Plan: Identify protos, backend API, frontend UI, and storage config changes
2. Phase 1: Update protos for new image fields, regenerate code
3. Phase 2: Launch parallel Task agents for backend and frontend
4. Review: Verify end-to-end flow works

### Example 2: Fixing a Bug
```
/monorepo-change Fix rate limiting bug in API gateway that affects all services
```

Workflow:
1. Plan: Identify bug is in Go gateway, affects rate limiter middleware
2. Phase 1: Fix bug in gateway, add regression test
3. Phase 2: Update integration tests
4. Review: Verify fix works and doesn't break other services

### Example 3: Updating Dependencies
```
/monorepo-change Update Node.js dependencies to latest versions and fix breaking changes
```

Workflow:
1. Plan: Identify all Node.js subprojects and their dependencies
2. Phase 1: Update package.json files, check for breaking changes
3. Phase 2: Launch parallel Task agents to fix breaking changes in each subproject
4. Review: Verify all builds and tests pass

### Example 4: Refactoring
```
/monorepo-change Refactor shared error handling across Go services for consistency
```

Workflow:
1. Plan: Create shared error package, update all Go services to use it
2. Phase 1: Create shared error handling package with tests
3. Phase 2: Launch parallel Task agents to update each Go service
4. Review: Verify consistent error handling across services

### Example 5: Configuration Change
```
/monorepo-change Update logging configuration to use structured JSON logging
```

Workflow:
1. Plan: Update logging config in all subprojects
2. Phase 1: Update shared logging configuration
3. Phase 2: Launch parallel Task agents to update each subproject
4. Review: Verify logs are properly formatted in all services

---

## Change Type Guidelines

Different types of changes may require different approaches:

- **Features**: Focus on tests, documentation, and user-facing changes
- **Bugfixes**: Add regression tests, verify fix doesn't introduce new issues
- **Refactoring**: Ensure behavior is unchanged, improve code quality
- **Configuration**: Verify settings work in all environments
- **Dependencies**: Handle breaking changes, update lock files
- **Security**: Audit changes carefully, verify no vulnerabilities introduced
- **Performance**: Include benchmarks, verify improvements

---

## Success Criteria

**ALL of these must be true before completing the task:**

- ✅ All affected subprojects updated with appropriate skills
- ✅ **All tests pass in each subproject** (MANDATORY)
- ✅ **All lints pass in each subproject** (MANDATORY)
- ✅ **All builds succeed in each subproject** (MANDATORY)
- ✅ **All type checks pass in typed languages** (MANDATORY)
- ✅ **Tests updated for new or modified code** (MANDATORY)
- ✅ Integration between subprojects verified
- ✅ Changes reviewed for quality, security, and consistency
- ✅ User receives comprehensive summary of all changes made

**If any check fails, you MUST fix it before considering the task complete.**
