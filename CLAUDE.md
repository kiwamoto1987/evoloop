# Evoloop

A CLI tool that runs self-improvement loops on local Git projects.

## Overview

Self-improvement loop: `inspect → analyze → propose → evaluate → history`

## Tech Stack

- Language: Go
- CLI Framework: cobra
- Persistence: SQLite
- Config: YAML (.evoloop/config.yaml)
- ID Generation: ULID
- LLM: Claude CLI (subprocess invocation)

## Directory Structure

```
cmd/           # cobra command definitions
internal/
  domain/      # domain models (ProjectContext, ImplementationIssue, etc.)
  service/     # service layer (ProjectInspectionService, etc.)
  repository/  # SQLite repository implementations
  llm/         # LanguageModelClient implementation
  policy/      # ExecutionPolicy
  config/      # config.yaml loader
.evoloop/      # runtime data (generated at runtime)
```

## Language Policy

- The official language of this project is English
- All source code, comments, commit messages, and documentation must be written in English

## Naming Conventions

- OOP naming: clearly separate Service / Repository / Client
- Util / Helper classes are prohibited
- IDs are generated using ULID

## Implementation Policy

- Keep implementations small
- analyze is rule-based (no LLM)
- Only propose uses LLM (Claude CLI)
- evaluate is non-LLM (determined by test / lint / typecheck results)

## Development Rules

- Follow TDD: write tests first, then implement to make them pass
- Never commit directly to master; always create a feature branch and open a pull request

## Build & Test

```bash
go build ./...
go test ./...
golangci-lint run
```

