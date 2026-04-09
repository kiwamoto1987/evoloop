# Evoloop

A CLI tool that orchestrates self-improvement loops on local Git projects.

Evoloop provides the **workflow** for automated improvement: issue detection, LLM-driven patch generation, evaluation, and application. The domain-specific logic (what to monitor, how to validate, what to restart) is injected by the user through configuration and external scripts.

## Operating Modes

### Mode A: Code Quality

Detects test/lint/typecheck failures and generates code fixes.

```
inspect → analyze → propose → evaluate(sandbox) → apply
```

### Mode B: Operational Improvement

External scripts inject issues (e.g., KPI degradation). Evoloop proposes config/code changes, validates them, applies, and runs post-apply hooks.

```
[external script] → issue create → propose → evaluate(validate_only) → apply → hook
```

## Commands

```
evoloop init                    # Set up .evoloop/ config
evoloop run                     # Run the full improvement loop
evoloop inspect                 # Detect project structure
evoloop analyze                 # Run quality checks, generate issues
evoloop propose --issue <ID>    # Generate a patch proposal via LLM
evoloop evaluate --exec <ID>    # Evaluate a proposal
evoloop history                 # View execution history
evoloop issue create            # Inject an issue from external source
```

### Automated Loop

```bash
evoloop run                                  # single iteration, review only
evoloop run --auto-apply                     # apply accepted patches
evoloop run --auto-apply --max-iterations 3  # up to 3 iterations
evoloop run --max-failures 2                 # stop after 2 consecutive failures
```

`evoloop run` follows this logic:

1. Check DB for existing Open issues (from `issue create` or previous runs)
2. If none, run inspect + analyze to detect new issues
3. Select the best candidate (priority, retry count, cooldown)
4. Propose a patch via LLM
5. Evaluate in sandbox or validate_only mode
6. If accepted and `--auto-apply`: apply patch, run post-apply hook
7. Record all results in SQLite

### External Issue Injection

```bash
evoloop issue create \
  --title "High slippage on Arbitrum" \
  --description "Avg slippage 2.1% over last hour" \
  --category kpi_degradation \
  --remediation config_patch \
  --priority 1 \
  --source check_kpi.sh \
  --source-ref "rule:slippage_threshold" \
  --dedup-key "kpi:slippage:arbitrum"
```

Features:
- **Validation**: category allowlist, priority range, description length limit
- **Dedup**: same `--dedup-key` updates the existing Open issue instead of creating a duplicate
- **Source tracking**: `--source` and `--source-ref` record where the issue came from

### Step-by-Step (Mode A)

1. **inspect** detects the Git project, branch, dirty state, and available commands
2. **analyze** runs quality checks and generates `ImplementationIssue` entries
3. **propose** builds a prompt, calls the LLM, and saves the resulting patch
4. **evaluate** applies the patch in a sandbox, runs checks, and decides Accept or Reject
5. **history** shows all issues, executions, and evaluations

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

### Mode A: Code Quality

```bash
cd your-project
evoloop init
evoloop run --auto-apply
```

### Mode B: Operational Improvement

```bash
cd your-project
evoloop init
# Edit .evoloop/config.yaml (see Configuration below)

# External script injects an issue
evoloop issue create \
  --title "KPI degraded" \
  --description "Details..." \
  --category kpi_degradation \
  --remediation config_patch

# Evoloop picks it up and runs the loop
evoloop run --auto-apply
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
  # validate_commands: []  # For validate_only mode

policies:
  max_changed_files: 5
  max_changed_lines: 200
  deny_paths:
    - ".github/**"
    - ".evoloop/**"
  evaluation_mode: "sandbox"  # sandbox | validate_only
  max_attempts: 3             # max proposal attempts per issue
  cooldown_minutes: 60        # wait time after failure before retry

# hooks:
#   post_apply:
#     command: "systemctl"
#     args: ["restart", "my-service"]
#     timeout_sec: 30
#     allowlist: ["systemctl"]

# issues:
#   allowed_categories:
#     - "kpi_degradation"
#     - "config_tuning"
#   max_priority: 10
#   max_description_length: 5000
```

### Mode B Example Config

```yaml
project_name: dex-bot

evaluation:
  validate_commands:
    - "yamllint config.yaml"
    - "./scripts/validate_config.sh"

policies:
  evaluation_mode: "validate_only"
  max_attempts: 3
  cooldown_minutes: 30

hooks:
  post_apply:
    command: "systemctl"
    args: ["restart", "trade-bot"]
    timeout_sec: 30
    allowlist: ["systemctl"]

issues:
  allowed_categories:
    - "kpi_degradation"
    - "config_tuning"
  max_priority: 10
  max_description_length: 5000
```

## Evaluation Modes

### sandbox (default)

Copies the project to a temp directory, applies the patch, and runs test/lint/typecheck. Accept requires all checks to pass.

### validate_only

Applies the patch in a temp directory and runs `validate_commands` only. Test/lint/typecheck are skipped. Useful for config file changes where full test suites are unnecessary or infeasible.

> **Note**: `validate_only` is minimum syntax verification, not runtime safety validation. It cannot detect semantically dangerous values.

## Issue Lifecycle

```
Open → Proposed → Accepted → Applied → Completed
                → Rejected ──→ Open (retry, within max_attempts)
                               → Closed (max_attempts exceeded)
       Accepted → ApplyFailed → Open (retry)
       Applied  → HookFailed  → Open (retry)
```

## Post-Apply Hooks

Hooks execute after a patch is applied to the project. They are structured commands (not shell strings) with safety constraints:

- **Allowlist**: command must be in the configured allowlist
- **Timeout**: enforced via context; killed if exceeded
- **Audit trail**: exit code, stdout, stderr, duration recorded in SQLite

## Architecture

Evoloop is a **workflow orchestrator**. It manages the loop and provides safety/state/retry infrastructure. All domain-specific logic is external:

| Concern | Responsibility |
|---------|---------------|
| KPI collection / monitoring | External scripts |
| Issue detection criteria | External scripts → `issue create` |
| Change proposal | LLM (Claude CLI) |
| Validation commands | User-defined in config |
| Post-apply actions | User-defined hooks |
| Workflow orchestration | **Evoloop** |
| State management (SQLite) | **Evoloop** |
| Safety (allowlist, timeout, policy, dedup, retry) | **Evoloop** |

## Requirements

- Go 1.22+
- Git
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-cli) (for `propose` command)

## License

[MIT](LICENSE)
