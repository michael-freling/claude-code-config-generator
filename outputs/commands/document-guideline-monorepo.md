---
description: Analyze monorepo and create comprehensive guidelines for each subproject as an Architect
argument-hint: ""
allowed-tools: ["*"]
---

# Document Guideline for Monorepo

You are working as an **Architect** agent to analyze a monorepo codebase and create comprehensive project guidelines for each subproject.

## Instructions

1. Analyze the monorepo structure thoroughly to:
   - Identify all subprojects and their locations
   - Understand the relationship between subprojects
   - Identify shared dependencies and common patterns
   - Understand each subproject's purpose and technology stack

2. For **each subproject**, create a guideline document at `<subproject>/.claude/docs/guideline.md` that includes at least:
   - **Architecture**: System design, components, and their relationships specific to this subproject
   - **API Designs**: Endpoint patterns, request/response formats, versioning strategies
   - **Data Models**: Database schemas, entity relationships, data flow patterns
   - **Design Patterns**: Patterns used in the subproject and when to apply them
   - **Coding Best Practices**: Subproject-specific standards, conventions, and patterns

3. For the **root project**, create a guideline at `.claude/docs/guideline.md` that includes:
   - Monorepo structure and organization
   - Cross-project patterns and shared conventions
   - Dependency management strategies
   - **DO NOT** include subproject-specific guidelines in the root guideline

4. Follow these requirements:
   - Keep guidelines concise and actionable
   - Include examples when they help remove ambiguity
   - Focus on project-specific patterns that override general best practices
   - Update existing guidelines rather than creating duplicates
   - Make guidelines useful for both Claude Code and human developers

## Important Notes

- This command is for **monorepos only**
- For single projects, use `/document-guideline` instead
- Each subproject should have its own `.claude/docs/guideline.md`
- The root guideline should focus on monorepo-wide patterns, not subproject details
- Each guideline takes precedence over general coding guidelines for its scope
- Always read existing guidelines before making changes to understand current conventions
