# Claude Code Configuration Generator

A CLI tool to generate prompts for creating Claude Code skills, agents, and commands.

## Overview

This generator outputs PROMPTS to stdout that you can give to Claude to create skills, agents, and commands. It does NOT create the files directly - instead, it generates instructions that Claude can use.

This follows the same pattern as the bash scripts in `scripts/`:
- `coding-subagents.sh` - Outputs prompts for creating agent definitions
- `coding-commands.sh` - Outputs prompts for creating slash commands
- `coding-skills.sh` - Outputs prompts for creating skills

## Installation

```bash
go build -o generator cmd/generator/main.go
```

## Usage

### Generate prompts for all agents

```bash
./generator agents
```

This outputs prompts that tell Claude how to create agent definitions like:
- Software Architect
- Golang Engineer
- Golang Reviewer
- TypeScript Engineer
- TypeScript Reviewer
- etc.

### Generate prompts for all commands

```bash
./generator commands
```

This outputs prompts for creating slash commands like:
- `/feature` - Add or update a feature
- `/fix` - Fix a bug
- `/refactor` - Refactor code
- `/document-guideline` - Create project guidelines
- `/split-pr` - Split large PRs into smaller ones

### Generate prompts for all skills

```bash
./generator skills
```

This outputs prompts for creating skills like:
- `coding` - Iterative development with TDD
- `ci-error-fix` - Fix CI errors systematically

## How It Works

1. The generator uses Go templates in `internal/generator/templates/prompts/`
2. Shared rules (COMMON_RULES, CODING_RULES, GOLANG_RULES, TYPESCRIPT_RULES) are defined in `_partials.tmpl`
3. Each template includes the relevant shared rules using `{{template "RULE_NAME"}}`
4. Output is always to stdout - no files are written

## Template Structure

```
internal/generator/templates/prompts/
├── _partials.tmpl          # Shared rule definitions
├── agents/
│   ├── software-architect.tmpl
│   ├── golang-engineer.tmpl
│   ├── golang-code-reviewer.tmpl
│   └── ...
├── commands/
│   ├── feature.tmpl
│   ├── fix.tmpl
│   ├── refactor.tmpl
│   └── ...
└── skills/
    ├── coding.tmpl
    ├── ci-error-fix.tmpl
    └── ...
```

## Module Path

This project uses the module path `github.com/michael-freling/claude-code-config` to be at the root level of the repository.

## Development

### Build

```bash
go build ./...
```

### Verify

```bash
go vet ./...
go fmt ./...
```

### Test output

```bash
./generator agents | head -50
./generator commands | head -50
./generator skills
```
