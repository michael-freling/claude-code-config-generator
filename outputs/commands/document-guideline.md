---
description: Analyze the codebase and create comprehensive project guidelines as an Architect
argument-hint: ""
allowed-tools: ["*"]
---

# Document Guideline

You are working as an **Architect** agent to analyze the codebase and create comprehensive project guidelines.

## Instructions

1. Analyze the codebase thoroughly to understand:
   - Project structure and organization
   - Existing patterns and conventions
   - Technology stack and dependencies
   - Code organization and module boundaries

2. Create a guideline document at `.claude/docs/guideline.md` that includes at least:
   - **Architecture**: System design, components, and their relationships
   - **API Designs**: Endpoint patterns, request/response formats, versioning strategies
   - **Data Models**: Database schemas, entity relationships, data flow patterns
   - **Design Patterns**: Patterns used in the project and when to apply them
   - **Coding Best Practices**: Project-specific standards, conventions, and patterns

3. Follow these requirements:
   - Keep guidelines concise and actionable
   - Include examples when they help remove ambiguity
   - Focus on project-specific patterns that override general best practices
   - Update existing guidelines rather than creating duplicates
   - Make the guideline useful for both Claude Code and human developers

## Important Notes

- This command is for **single projects only**
- For monorepos, use `/document-guideline-monorepo` instead
- The guideline in `.claude/docs/guideline.md` takes precedence over general coding guidelines
- Always read existing `.claude/docs/guideline.md` before making changes to understand current conventions
