---
stepsCompleted:
  - step-01-init
  - step-02-discovery
  - step-02b-vision
  - step-02c-executive-summary
  - step-03-success
  - step-04-journeys
  - step-05-domain-skipped
  - step-06-innovation-skipped
  - step-07-project-type
  - step-08-scoping
  - step-09-functional
  - step-10-nonfunctional
  - step-11-polish
  - step-12-complete
classification:
  projectType: cli_tool
  domain: developer_tools_general
  complexity: low
  projectContext: brownfield
inputDocuments:
  - _bmad-output/brainstorming/brainstorming-session-2026-04-17-0412.md
  - _bmad-output/project-context.md
  - docs/REQUIREMENT.md
  - docs/index.md
  - docs/project-overview.md
  - docs/architecture.md
  - docs/component-inventory.md
  - docs/development-guide.md
  - docs/source-tree-analysis.md
documentCounts:
  briefs: 0
  research: 0
  brainstorming: 1
  projectDocs: 7
  projectContext: 1
workflowType: 'prd'
---

# Product Requirements Document - ccctx

**Author:** dsdashun
**Date:** 2026-04-17

## Executive Summary

ccctx is a Go CLI tool for Claude Code users who frequently switch between different API providers and models. It replaces manual environment file editing and `source` with a single command: `ccctx run <provider>` or `ccctx exec <provider>`.

**Target Users:** Company developers using multiple Claude API providers (third-party proxies, AWS Bedrock, etc.) who need to quickly switch contexts without memorizing or manually exporting `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, and related environment variables. These users face daily provider quota limits and must switch providers when one is exhausted.

**Problem Solved:** Managing environment variables for multiple providers is error-prone, hard to remember, and incompatible with automation scripts. ccctx manages all contexts through a TOML configuration file and switches with a single command.

**Core Abstraction:** ccctx is a **provider config → env vars → exec** bridge. `run` is a convenience wrapper for Claude Code (equivalent to `exec -- claude`), while `exec` extends this capability to arbitrary shell sessions, scripts, and CI/CD pipelines.

### What Makes This Special

- **Minimal Interface** — One command replaces multiple `export` operations; users never need to remember env var names
- **Interactive Fallback** — TUI selector pops up when no provider is specified, with arrow key and vim key binding support
- **Automation-Native** — Pure CLI design fits script orchestration and CI/CD pipelines, unrestricted by GUI manual interaction
- **Shared Kernel Architecture** — `run` and `exec` share provider resolution, env var construction, and process execution logic

## Project Classification

| Field | Value |
|-------|-------|
| **Project Type** | CLI Tool (Go, Cobra framework) |
| **Domain** | Developer Tools |
| **Complexity** | Low |
| **Project Context** | Brownfield — existing codebase with `list` and `run` commands; adding `exec` subcommand and shared kernel refactoring |

## Success Criteria

### User Success

- `ccctx exec <provider>` launches a shell with all provider env vars pre-configured; child processes inherit correct configuration
- `ccctx exec` (no provider) opens the TUI selector for quick provider selection without memorizing names
- `ccctx run <provider>` behavior remains unchanged — existing users experience zero disruption
- `ccctx exec <provider> -- <command>` executes any command with provider context in one step
- Exit code passthrough ensures automation reliability

### Business Success

- `exec` subcommand expands use cases beyond Claude Code to any AI CLI tool or automation script
- Shared kernel architecture reduces maintenance burden — future features extend the kernel without touching individual commands
- Existing `run` users experience no regressions; refactoring is transparent

### Technical Success

- Shared kernel extracted from `run.go` into `internal/runner/` — reusable provider resolution, env var building, process execution
- `run` refactored as thin wrapper with `claude` as hardcoded target
- `exec` calls shared kernel with user-specified command or `$SHELL` default
- TUI selector reusable across both commands
- All existing tests pass; new table-driven tests cover `exec` argument combinations
- Exit code passthrough works for both interactive and non-interactive invocations

### Measurable Outcomes

- `ccctx exec <provider> -- env | grep ANTHROPIC` shows correct values for all configured env vars
- `ccctx exec <provider> -- some-command && echo "passed"` reflects the command's actual exit code
- `ccctx exec` (no args) launches TUI → user selects provider → lands in configured shell
- `ccctx run <provider> -- --help` produces identical output to current behavior
- `go test ./...` passes with zero regressions

## User Journeys

### Journey 1: Quota Exhausted — Quick Provider Switch

**Persona:** Company developer using Claude Code via third-party providers (direct Anthropic access unavailable)

**Scenario:** Developer works with provider-A in the morning. By afternoon, provider-A returns a quota-exceeded error. They switch to provider-B immediately.

**Journey:**
1. Developer sees quota-exceeded error and exits current Claude Code session
2. Runs `ccctx exec provider-B`, instantly landing in a shell with provider-B's env vars
3. Runs `claude` in the new shell and continues working seamlessly
4. If unsure which providers have remaining quota, runs `ccctx exec` for TUI selection

**Key Requirement:** Switch is instant — one command, no memorizing env var names

### Journey 2: Automation Script with Claude Code Invocation

**Persona:** Company developer writing a multi-step CLI script that invokes Claude Code

**Scenario:** Developer has a code review script where one step calls `claude -p` to analyze a diff. The script must run under the correct provider context.

**Journey:**
1. Developer writes `ccctx exec provider-A -- bash review.sh` in script or CI config
2. Inside `review.sh`, `claude -p` automatically inherits provider-A's env vars
3. Script completes with correct exit code passthrough — CI determines success/failure
4. If provider-A quota is exhausted, developer changes one provider name in the script

**Key Requirement:** Env vars persist in child processes, exit code passthrough, script-friendly CLI

### Journey Requirements Summary

| Capability | Journey 1 | Journey 2 |
|-----------|-----------|-----------|
| `ccctx exec <provider>` | ✓ | ✓ |
| `ccctx exec` (TUI selector) | ✓ | |
| `ccctx exec <provider> -- <command>` | | ✓ |
| Env var inheritance in child processes | ✓ | ✓ |
| Exit code passthrough | | ✓ |
| `$SHELL` default when no command | ✓ | |

## Project Scoping & Phased Development

### MVP Strategy

**Approach:** Problem-solving MVP — deliver the `exec` subcommand that solves provider switching in shell sessions and automation scripts. Build on existing `run` with shared kernel architecture for extensibility.

**Resources:** Solo developer, side project, no timeline pressure.

### Phase 1 — MVP

**Core User Journeys:** Quota-exhausted quick provider switch + automation script invocation

**Must-Have Capabilities:**
- Shared kernel extraction from `run.go` into `internal/runner/`
- `exec` subcommand (`cmd/exec.go`)
- `ccctx exec <provider>` — launches `$SHELL` with provider env vars
- `ccctx exec <provider> -- <command>` — executes command with provider env vars
- `ccctx exec` (no provider) — TUI interactive selector → `$SHELL`
- Exit code passthrough
- `run` refactored as thin wrapper, behavior unchanged
- Table-driven tests for all `exec` argument combinations

### Phase 2 — Growth

- `--model <value>` flag for config-level parameter override
- Additional config-level flags (`--base-url`, etc.) via shared kernel extension

### Phase 3 — Expansion

- Compatibility validation with any AI CLI tool using `ANTHROPIC_*` env vars
- CI/CD pipeline integration examples
- Multi-provider comparison testing support
- Shell completion (bash, zsh, fish)

### Risk Mitigation

**Technical:** Low — shared kernel extraction is a well-understood refactoring pattern. Existing tests serve as regression safety net.

**Resource:** Minimal — solo developer, no deadline. Phase 2 and 3 can be deferred indefinitely.

## CLI Tool Specific Requirements

### Command Structure

| Command | Usage | Behavior |
|---------|-------|----------|
| `ccctx list` | List available contexts | Outputs context names to stdout |
| `ccctx run [provider] [-- claude-args]` | Run Claude Code with provider context | Hardcoded `claude` target, env vars injected |
| `ccctx exec [provider] [-- command]` | Execute command or launch shell with provider context | User-specified target or `$SHELL` default |
| `ccctx exec` (no args) | Interactive TUI selector | Select provider, launch `$SHELL` |

**Argument parsing rules:**
- Arguments before `--` are treated as the provider name
- Arguments after `--` are forwarded as the command to execute
- No `--` and no provider: triggers interactive TUI selector
- No `--` with provider: `exec` defaults to `$SHELL`, `run` defaults to `claude`

### Output Formats

- **list:** Context names to stdout, one per line
- **run / exec (success):** Exit code passthrough from child process
- **run / exec (error):** Error messages to stderr via `fmt.Fprintf(os.Stderr, ...)`, exit code 1
- **TUI:** Terminal-based interactive selector using tview

### Configuration Schema

```toml
[context.<name>]
base_url = "https://..."           # Required
auth_token = "token" or "env:VAR"  # Required, supports env: prefix
model = "claude-..."               # Optional
small_fast_model = "claude-..."    # Optional
```

- Config path: `~/.ccctx/config.toml` (overridable via `CCCTX_CONFIG_PATH`)
- Auto-creates config directory and example file if missing
- `env:` prefix resolves auth_token from OS environment variables

### Scripting Support

- `ccctx exec <provider> -- <command>` enables full script integration
- Exit code from child process passed through (`exec.ExitError`)
- Env vars persist in all child processes (inherited via `exec.Cmd`)
- Provider config env vars override existing `ANTHROPIC_*` values
- No `LookPath` dependency in `exec` — command is user-specified directly

## Functional Requirements

### Context Configuration

- FR1: Users can define multiple named provider contexts in a TOML configuration file, each with `base_url` and `auth_token` fields
- FR2: Users can optionally specify `model` and `small_fast_model` per context to override default model selection
- FR3: Users can reference environment variables in `auth_token` using the `env:` prefix for secure token resolution
- FR4: The system auto-creates the config directory and example configuration file when none exists
- FR5: Users can override the default config path (`~/.ccctx/config.toml`) via the `CCCTX_CONFIG_PATH` environment variable

### Context Discovery

- FR6: Users can list all configured context names via `ccctx list`
- FR7: Users can view context names output to stdout, one per line

### Direct Execution (run command)

- FR8: Users can run Claude Code with a specified provider context via `ccctx run <provider>`
- FR9: Users can forward additional arguments to Claude Code using the `--` separator
- FR10: The system injects the provider's `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, and optional model env vars into the Claude process
- FR11: The system strips any pre-existing `ANTHROPIC_*` environment variables before injecting provider-specific values
- FR12: The system propagates Claude's exit code back to the caller

### Flexible Execution (exec command)

- FR13: Users can launch an interactive shell session with provider env vars via `ccctx exec <provider>`
- FR14: Users can execute an arbitrary command with provider env vars via `ccctx exec <provider> -- <command> [args...]`
- FR15: The system defaults to `$SHELL` when no command is specified after the provider name
- FR16: All child processes of the launched shell or command inherit the provider environment variables
- FR17: The system propagates the child process exit code back to the caller
- FR18: The system overrides any existing `ANTHROPIC_*` environment variables with the provider's values

### Interactive Selection (TUI)

- FR19: Users can launch an interactive context selector via `ccctx exec` (no provider specified)
- FR20: Users can navigate the context list using arrow keys (up/down) and vim keys (j/k)
- FR21: Users can select a context by pressing Enter or clicking on an item
- FR22: Users can cancel the selection by pressing ESC
- FR23: The TUI selector is available in both `run` and `exec` commands when no provider is specified

### Shared Architecture

- FR24: The `run` and `exec` commands share a common provider resolution, env var construction, and process execution pipeline
- FR25: The `run` command uses the shared pipeline with `claude` as the hardcoded execution target
- FR26: The `exec` command uses the shared pipeline with a user-specified command or `$SHELL` as the execution target

### Config-Level Parameter Overrides (Phase 2)

- FR27: Users can override the provider's configured model via `--model <value>` flag on both `run` and `exec` commands
- FR28: The system applies CLI flag values with higher priority than config file values, allowing temporary overrides without modifying the config

## Non-Functional Requirements

### Security

- NFR1: Auth tokens are never written to stdout, stderr, or log output — resolved internally only
- NFR2: Existing `ANTHROPIC_*` environment variables are stripped before injecting provider-specific values, preventing credential leakage across contexts
- NFR3: The `env:` prefix resolution reads environment variables at runtime only — tokens are never persisted in expanded form
- NFR4: Config file permissions should be user-readable only (mode 0600) to protect stored credentials

### Compatibility

- NFR5: ccctx works with any POSIX-compliant shell (`bash`, `zsh`, `sh`, etc.) as the `$SHELL` default
- NFR6: ccctx works with any AI CLI tool that reads `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN` environment variables, not just Claude Code
- NFR7: The tool builds as a static binary (`CGO_ENABLED=0`) with zero runtime dependencies
