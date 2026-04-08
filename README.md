# Evoloop

A CLI tool that runs self-improvement loops on local Git projects.

Evoloop inspects your project, detects quality issues, generates patch proposals via LLM, evaluates them against configurable policies, and tracks the results — all locally.

## How It Works

```
evoloop init       # Set up .evoloop/ config
evoloop inspect    # Detect project structure and available commands
evoloop analyze    # Run quality checks and generate improvement issues
evoloop propose    # Generate a patch proposal for an issue using LLM
evoloop evaluate   # Evaluate a proposal against tests, lint, and policy
evoloop history    # View execution history
```

### Loop

```
inspect → analyze → propose → evaluate → history
```

1. **inspect** detects the Git project, branch, dirty state, and available test/lint/typecheck commands.
2. **analyze** runs those commands and generates `ImplementationIssue` entries for any failures.
3. **propose** takes an issue, builds a prompt, calls the LLM, and saves the resulting patch.
4. **evaluate** applies the patch in a temporary directory, runs quality checks, and decides Accept or Reject.
5. **history** shows all issues, executions, and evaluations stored in SQLite.

## Installation

```bash
go install github.com/kiwamoto1987/evoloop@latest
```

Or build from source:

```bash
git clone https://github.com/kiwamoto1987/evoloop.git
cd evoloop
go build -o evoloop .
```

## Quick Start

```bash
# Initialize in your project
cd your-project
evoloop init

# Edit config if needed
vim .evoloop/config.yaml

# Run the loop
evoloop inspect
evoloop analyze
evoloop propose --issue <ISSUE_ID>
evoloop evaluate --execution <EXECUTION_ID>
evoloop history
```

## Configuration

`evoloop init` generates `.evoloop/config.yaml`:

```yaml
project_name: my-project

llm:
  provider: claude
  model: sonnet
  command: "claude"

evaluation:
  test_command: "go test ./..."
  lint_command: "golangci-lint run"
  typecheck_command: "go build ./..."

policies:
  max_changed_files: 5
  max_changed_lines: 200
  deny_paths:
    - ".github/**"
    - ".evoloop/**"
```

## Evaluation Policy

Every proposal is evaluated against:

- **TestPassed** — all tests pass
- **LintPassed** — no lint violations
- **TypeCheckPassed** — type checks pass
- **ChangedFileCount** — within `max_changed_files` limit
- **ChangedLineCount** — within `max_changed_lines` limit

All criteria must pass for a proposal to be accepted.

## Requirements

- Go 1.22+
- Git
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-cli) (for `propose` command)

## License

[MIT](LICENSE)
