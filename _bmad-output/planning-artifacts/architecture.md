---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
lastStep: 8
status: 'complete'
completedAt: '2026-04-18'
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/project-context.md
  - _bmad-output/brainstorming/brainstorming-session-2026-04-17-0412.md
  - _bmad-output/planning-artifacts/epics.md
  - docs/index.md
  - docs/architecture.md
  - docs/component-inventory.md
  - docs/development-guide.md
  - docs/project-overview.md
  - docs/source-tree-analysis.md
  - docs/REQUIREMENT.md
workflowType: 'architecture'
project_name: 'ccctx'
user_name: 'dsdashun'
date: '2026-04-18'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**

28 FRs total — 12 already implemented (FR1-FR12), 16 new (FR13-FR28):
- **Context Configuration (FR1-FR5):** TOML-based multi-context config with env: prefix resolution — ALREADY DONE
- **Context Discovery (FR6-FR7):** `ccctx list` command — ALREADY DONE
- **Direct Execution via run (FR8-FR12):** Claude Code launch with provider env vars, argument forwarding, exit code passthrough — ALREADY DONE
- **Flexible Execution via exec (FR13-FR18):** NEW — shell/command execution with provider context, $SHELL default, child process env inheritance, exit code passthrough
- **Interactive Selection (FR19-FR23):** TUI selector for exec (FR19, FR23 are NEW; FR20-FR22 already done)
- **Shared Architecture (FR24-FR26):** NEW — common execution pipeline for run and exec
- **Config-Level Overrides (FR27-FR28):** NEW (Phase 2) — --model CLI flag with priority over config values

**Non-Functional Requirements:**

7 NFRs — 6 satisfied, 1 new:
- NFR1-NFR3 (Security): Token protection, ANTHROPIC_* stripping, runtime-only resolution — DONE
- NFR4 (Security): Config file permissions 0600 — NEW, to be implemented in Epic 1
- NFR5-NFR7 (Compatibility): POSIX shell, multi-tool compatibility, static binary — DONE

**Scale & Complexity:**

- Primary domain: Developer CLI tools (Go)
- Complexity level: Low
- Estimated architectural components: 5 (shared kernel, run wrapper, exec command, TUI selector, config layer)
- Existing codebase: ~5 source files, focused scope

### Technical Constraints & Dependencies

- **Language:** Go 1.23.3, CGO_ENABLED=0 for static binary
- **CLI Framework:** Cobra v1.9.1 — command structure, argument parsing
- **Config:** Viper v1.20.1 — TOML format, mapstructure tags
- **TUI:** tview + tcell/v2 — interactive selector
- **Build:** Makefile, no CI/CD pipeline
- **Testing:** Standard Go testing, table-driven patterns
- **No LookPath in exec:** exec command receives user-specified binary directly
- **env: prefix:** Custom convention, not a Viper feature — must be resolved explicitly via resolveEnvVar()
- **-- separator:** Critical argument boundary — args before are context names, after are forwarded

### Cross-Cutting Concerns Identified

1. **Environment Variable Management** — core mechanism shared by run and exec: strip ANTHROPIC_*, inject provider values, inherit in child processes
2. **Exit Code Passthrough** — both commands must propagate child process exit codes via exec.ExitError
3. **TUI Selector Reuse** — internal/ui/selector.go must serve both run and exec without duplication
4. **Provider Resolution Pipeline** — shared flow: parse args → resolve provider → load config → build env vars → execute
5. **Security** — tokens never in stdout/stderr/logs, env: resolved at runtime only, config file permissions

## Starter Template Evaluation

### Primary Technology Domain

CLI Tool (Go) — brownfield project with existing codebase. No starter template required.

### Technology Stack (Established)

This is a brownfield project. All technology decisions are already in place:

| Category | Technology | Version | Status |
|----------|-----------|---------|--------|
| Language | Go | 1.23.3 | Current |
| CLI Framework | Cobra | v1.9.1 | Current, actively maintained |
| Configuration | Viper | v1.20.1 | Upgrade available (v1.21.1 — reduced dependencies) |
| Terminal UI | tview | latest commit | Current |
| Terminal Interface | tcell/v2 | v2.8.1 | Current |
| Build | Makefile | — | Static binary (CGO_ENABLED=0) |
| Testing | Go standard testing | — | Table-driven patterns |

### Dependency Upgrade Note

Viper v1.21.1 is available with major improvements including heavily reduced third-party dependencies and a new encoding layer. This upgrade should be considered as a separate task after the exec feature is implemented, to avoid conflating feature work with dependency updates.

### Existing Project Structure

```
ccctx/
├── main.go                 # Entry point (Cobra root command)
├── cmd/
│   ├── list.go             # ccctx list
│   └── run.go              # ccctx run (to be refactored)
├── config/
│   ├── config.go           # TOML config management
│   └── config_test.go      # Unit tests
├── internal/
│   └── ui/
│       └── selector.go     # TUI interactive selector
├── examples/
│   └── config.toml         # Sample configuration
├── docs/                   # Project documentation
├── go.mod / go.sum
└── Makefile
```

### Architectural Extension Plan

Rather than a starter template, the architecture extends the existing codebase:

1. **NEW:** `internal/runner/` — shared execution kernel (extracted from run.go)
2. **REFACTOR:** `cmd/run.go` — thin wrapper calling shared kernel
3. **NEW:** `cmd/exec.go` — exec subcommand using shared kernel
4. **EXISTING:** `config/`, `internal/ui/` — unchanged (TUI shared by both commands)
5. **NEW:** `internal/runner/runner_test.go` — tests for shared kernel

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
- AD1: Shared kernel uses Runner struct pattern (functional pipeline rejected)
- AD2: Environment variable filtering uses `strings.HasPrefix("ANTHROPIC_")` (replaces hardcoded length checks)
- AD3: exec argument parsing mirrors run's logic (consistent UX, no LookPath in exec)

**Important Decisions (Shape Architecture):**
- AD4: Config file permissions set to 0600 on creation only (NFR4)
- AD5: Test coverage for shared kernel + argument parsing functions

**Deferred Decisions (Post-MVP):**
- `--model` and `--small-fast-model` flags (Phase 2, Epic 3) — **Known issue:** registering these as Cobra flags will cause Cobra to consume the `--` separator, breaking `ParseArgs`. Phase 2 design must resolve this (likely via `cmd.DisableFlagParsing = true` with manual flag parsing, or restricting `--model` to appear only before the provider name).
- Viper upgrade to v1.21.1 (separate task after feature work)

### AD1: Shared Kernel — Runner Struct Pattern

**Decision:** Extract shared execution logic into `internal/runner/` using a struct-based design.

**Rationale:** The `--model` flag (Phase 2) needs to carry state between provider resolution and env var construction. A struct naturally holds this intermediate state, keeping `run.go` and `exec.go` as thin wrappers with a single call: `runner.New(opts).Run()`.

**Runner struct design:**

```go
package runner

type Options struct {
    ContextName    string   // resolved provider name
    Target         []string // command to execute (["claude"] for run, user-specified for exec)
    Model          string   // optional override (Phase 2)
    SmallFastModel string   // optional override (Phase 2)
}

type Runner struct {
    ctx  *config.Context
    opts Options
    env  []string
}

func New(opts Options) (*Runner, error)  // resolve provider + validate required fields + build env
func (r *Runner) Run() (int, error)      // execute target, return exit code + possible start error
```

**`Run()` implementation logic:**

```go
func (r *Runner) Run() (int, error) {
    cmd := exec.Command(r.opts.Target[0], r.opts.Target[1:]...)
    cmd.Env = r.env
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    if err != nil {
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            return exitErr.ExitCode(), nil // normal non-zero exit
        }
        return 1, err // start failure (binary not found, permission denied, etc.)
    }
    return 0, nil
}
```

**Key architectural insight:** `exec` is the core command, `run` is a specialization (`run` ≡ `exec -- claude`). `cmd/exec.go` contains the full command logic; `cmd/run.go` is an ultra-thin wrapper that hardcodes target to `["claude"]` and delegates to the same runner pipeline.

**Extraction map from current `run.go`:**

| Current location (run.go) | New location | Function |
|---|---|---|
| Lines 154-178 (env building) | `internal/runner/runner.go` → `buildEnv()` | Construct ANTHROPIC_* env vars |
| Lines 147-151 (LookPath) | `cmd/run.go` only | Binary discovery (exec doesn't use it) |
| Lines 180-196 (exec + exit code) | `internal/runner/runner.go` → `Run()` | Process execution with exit code + error passthrough |
| Lines 22-137 (arg parsing + TUI) | `internal/runner/args.go` → `ParseArgs()` | Shared argument parsing logic |

**Impact:** FR24, FR25, FR26

### AD2: Environment Variable Filtering

**Decision:** Replace hardcoded string length checks with `strings.HasPrefix("ANTHROPIC_")`.

**Current code (run.go:159-165):**
```go
if !(len(e) >= 19 && e[:19] == "ANTHROPIC_BASE_URL=") &&
   !(len(e) >= 20 && e[:20] == "ANTHROPIC_AUTH_TOKEN=") && ...
```

**New approach:**
```go
for _, e := range env {
    if !strings.HasPrefix(e, "ANTHROPIC_") {
        newEnv = append(newEnv, e)
    }
}
```

**Rationale:** Automatically covers future `ANTHROPIC_*` variables without modifying the filter list. Simpler, more maintainable, less error-prone than hardcoded length checks.

**Impact:** FR11, FR18, NFR2

### AD3: exec Argument Parsing — Core Command Pattern

**Decision:** `exec` is the core command with full argument handling. `run` is an ultra-thin wrapper that hardcodes `claude` as the target.

**Parsing rules (shared via `ParseArgs`):**
- `--` present: arguments before `--` → provider name (at most one), arguments after `--` → forwarded command/args
- No `--`, with args: first arg → provider name, remaining args → forwarded args (preserves `ccctx run provider somearg` behavior)
- No args at all → TUI interactive selector
- Multiple args before `--` → error (reject ambiguous input)

**Command-level differences (not in ParseArgs):**
- `exec` default target: `os.Getenv("SHELL")` (no LookPath)
- `run` default target: `["claude"]` (via LookPath)

**Argument parsing function:**
```go
func ParseArgs(args []string) (provider string, targetArgs []string, useTUI bool, err error)
```

**No cmdType parameter needed** — the default target is determined at the command level, not inside ParseArgs.

**Rationale:** `exec` is the generic provider→env→exec bridge. `run` is just `exec -- claude`. This eliminates code duplication and makes the relationship explicit.

**Impact:** FR13, FR14, FR15, FR19

### AD4: Config File Permissions

**Decision:** Set file permissions to 0600 when creating a new config file. Do not modify permissions of existing files.

**Implementation:** Change `os.WriteFile(configPath, content, 0644)` to `os.WriteFile(configPath, content, 0600)` in `config.go:77`.

**Rationale:** Auth tokens may be stored directly in config (not all users use `env:` prefix). Protecting credentials from other users is a security best practice. Not modifying existing files respects user customizations.

**Impact:** NFR4

### AD5: Testing Strategy

**Decision:** Test shared kernel functions + argument parsing. Command files remain thin wrappers without dedicated unit tests.

**Test scope:**
1. `internal/runner/runner_test.go` — `buildEnv()`, `(*Runner).Run()` exit code passthrough + start failure cases, env var override priority
2. `internal/runner/args_test.go` — `ParseArgs()` covering `--` separator, no args, TUI trigger conditions
3. Existing `config/config_test.go` — unchanged, `resolveEnvVar()` stays in config package

**Test pattern:** Table-driven tests with **testify** (`require` for fatal, `assert` for non-fatal assertions), consistent with CLAUDE.md guidelines.

**Impact:** All FRs — regression safety net for refactoring

### Decision Impact Analysis

**Implementation Sequence:**
1. AD2 (env filtering refactor) — extracted into `internal/runner/buildEnv()`
2. AD1 (Runner struct + extraction) — `internal/runner/` package created
3. AD4 (config permissions) — one-line change in `config.go`
4. AD3 (exec command) — new `cmd/exec.go` using shared kernel
5. AD5 (tests) — `runner_test.go` + `args_test.go`

**Cross-Component Dependencies:**
- AD1 depends on AD2 (env building is part of the runner)
- AD3 depends on AD1 (exec uses the shared runner)
- AD5 depends on AD1 and AD2 (tests cover extracted logic)
- AD4 is independent (can be done anytime)

## Implementation Patterns & Consistency Rules

### Critical Conflict Points Identified

6 areas where AI agents could make inconsistent choices during implementation.

### Export & Visibility Patterns

**Runner package (`internal/runner/`):**
- `Options` struct — exported (constructed by command files)
- `New(opts Options) (*Runner, error)` — exported (entry point)
- `(*Runner) Run() (int, error)` — exported (execution entry, returns exit code + possible start error)
- `buildEnv()` — unexported (internal implementation detail)
- `ParseArgs()` — exported (called by both command files)

**Command constants:**
- No command type constants needed — `ParseArgs()` is command-agnostic, default target resolved at command level

**Rule: Runner never exits, commands never contain business logic.**

| Layer | Responsibility | Error Output | Exit |
|-------|---------------|-------------|------|
| `runner.New()` | Returns `error` on failure (includes validation of required fields) | — | — |
| `runner.Run()` | Returns `(int, error)` — exit code and possible start error | — | — |
| `runner.ParseArgs()` | Returns `error` on invalid input | — | — |
| `cmd/*.go` | Calls runner, handles errors | `fmt.Fprintf(os.Stderr, ...)` | `os.Exit()` |

**Error message format:** `Error: <lowercase description without trailing period>` — matches existing `run.go` pattern.

**Cancellation handling:** `"operation cancelled"` from TUI → print `"Operation cancelled."` to stdout, `os.Exit(1)`. Cancellation is an abort, not a success.

### Argument Parsing Patterns

**ParseArgs signature:**
```go
func ParseArgs(args []string) (provider string, targetArgs []string, useTUI bool, err error)
```

**Parsing logic (command-agnostic):**

| Condition | provider | targetArgs | useTUI |
|-----------|----------|-----------|--------|
| `--` present, provider before `--` | args[0] before `--` | args after `--` | false |
| `--` present, no provider before `--` | `""` | args after `--` | true |
| `--` present, multiple args before `--` | — | — | — (error: ambiguous) |
| No `--`, one arg | args[0] | `[]` (empty — command decides default) | false |
| No `--`, multiple args | args[0] | args[1:] (forward remaining) | false |
| No args at all | `""` | `[]` (empty — command decides default) | true |

**Default target is resolved in command files:**
- `exec.go`: if `targetArgs` empty → `[]string{os.Getenv("SHELL")}` (must validate `SHELL != ""`)
- `run.go`: if `targetArgs` empty → `[]string{claudePath}` (from LookPath)

**Breaking changes from current code:**

1. **Unknown first arg now hard-errors instead of TUI fallback.** The current `run.go` validates the first arg against known context names and falls back to TUI mode if it doesn't match. The new `ParseArgs` always treats the first arg as a provider name, so `ccctx run nonexistent` will produce a hard error (`context 'nonexistent' not found`) instead of launching TUI. This is an intentional simplification — explicit error messages are better than silent TUI fallback.

2. **TUI cancellation exit code changed from 0 to 1.** The current `run.go` uses `return` (exit code 0) on ESC cancellation. The new architecture uses `os.Exit(1)`. Cancellation is an abort, not a success — exit code 1 is semantically correct.

3. **"No contexts found" output changed.** Current code prints `"No contexts found."` to stdout and exits 0. The new architecture reports this as an error to stderr (`Error: no contexts found`) and exits 1. No configured contexts is an error condition, not a normal state.

### Environment Variable Construction Patterns

**Required field validation:** `runner.New()` validates that `base_url` and `auth_token` are non-empty after resolving the context. If either is empty, `New()` returns an error (e.g., `context 'foo' is missing base_url`). This guarantees `buildEnv()` never encounters empty required fields, eliminating the ambiguity.

**Filtering:** `strings.HasPrefix(e, "ANTHROPIC_")` — strips all matching vars, catches future additions.

**Injection order:**
1. `ANTHROPIC_BASE_URL` — required (guaranteed non-empty by New() validation)
2. `ANTHROPIC_AUTH_TOKEN` — required (guaranteed non-empty by New() validation)
3. `ANTHROPIC_MODEL` — optional, only if non-empty
4. `ANTHROPIC_SMALL_FAST_MODEL` — optional, only if non-empty

**Empty value rule:** Never inject env vars with empty values. Check `!= ""` before appending. Required fields are enforced at the validation layer, not the injection layer.

**Model priority (Phase 2 ready):**
```
ANTHROPIC_MODEL: Options.Model > ctx.Model > omit
ANTHROPIC_SMALL_FAST_MODEL: Options.SmallFastModel > ctx.SmallFastModel > omit
```
If the higher-priority source is non-empty, use it. If empty, fall through to the next source. If all are empty, don't inject the variable at all.

### Testing Patterns

- Framework: **testify** (`github.com/stretchr/testify`) — `assert` and `require` packages
- Pattern: Table-driven `[]struct{ name string; ... }` with `t.Run()`
- Error assertions: `require.Error(t, err)` + `assert.Contains(t, err.Error(), "expected text")`
- No error assertions: `require.NoError(t, err)`
- Equality checks: `assert.Equal(t, expected, actual)`
- Env var tests: `t.Setenv()` with automatic cleanup

### Command File Template

**exec.go (core command):**
```go
var ExecCmd = &cobra.Command{
    Use:   "exec [context] [-- command...]",
    Short: "...",
    Args:  cobra.ArbitraryArgs,
    Run: func(cmd *cobra.Command, args []string) {
        provider, targetArgs, useTUI, err := runner.ParseArgs(args)
        if err != nil { fmt.Fprintf(os.Stderr, "Error: %v\n", err); os.Exit(1) }

        if useTUI {
            contexts, err := config.ListContexts()
            if err != nil { fmt.Fprintf(os.Stderr, "Error: %v\n", err); os.Exit(1) }
            if len(contexts) == 0 {
                fmt.Fprintf(os.Stderr, "Error: no contexts found\n")
                os.Exit(1)
            }
            provider, err = ui.RunContextSelector(contexts)
            if err != nil {
                if err.Error() == "operation cancelled" {
                    fmt.Println("Operation cancelled.")
                    os.Exit(1)
                }
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
            }
        }

        if len(targetArgs) == 0 {
            shell := os.Getenv("SHELL")
            if shell == "" {
                fmt.Fprintf(os.Stderr, "Error: SHELL environment variable not set\n")
                os.Exit(1)
            }
            targetArgs = []string{shell}
        }

        opts := runner.Options{ContextName: provider, Target: targetArgs}
        r, err := runner.New(opts)
        if err != nil { fmt.Fprintf(os.Stderr, "Error: %v\n", err); os.Exit(1) }

        exitCode, err := r.Run()
        if err != nil { fmt.Fprintf(os.Stderr, "Error: %v\n", err); os.Exit(1) }
        os.Exit(exitCode)
    },
}
```

**run.go (ultra-thin wrapper — exec -- claude):**
```go
var RunCmd = &cobra.Command{
    Use:   "run [context] [-- claude-args...]",
    Short: "...",
    Args:  cobra.ArbitraryArgs,
    Run: func(cmd *cobra.Command, args []string) {
        provider, targetArgs, useTUI, err := runner.ParseArgs(args)
        if err != nil { fmt.Fprintf(os.Stderr, "Error: %v\n", err); os.Exit(1) }

        if useTUI {
            contexts, err := config.ListContexts()
            if err != nil { fmt.Fprintf(os.Stderr, "Error: %v\n", err); os.Exit(1) }
            if len(contexts) == 0 {
                fmt.Fprintf(os.Stderr, "Error: no contexts found\n")
                os.Exit(1)
            }
            provider, err = ui.RunContextSelector(contexts)
            if err != nil {
                if err.Error() == "operation cancelled" {
                    fmt.Println("Operation cancelled.")
                    os.Exit(1)
                }
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
            }
        }

        claudePath, err := exec.LookPath("claude")
        if err != nil { fmt.Fprintf(os.Stderr, "Error: claude not found in PATH\n"); os.Exit(1) }

        if len(targetArgs) == 0 {
            targetArgs = []string{claudePath}
        } else {
            targetArgs = append([]string{claudePath}, targetArgs...)
        }

        opts := runner.Options{ContextName: provider, Target: targetArgs}
        r, err := runner.New(opts)
        if err != nil { fmt.Fprintf(os.Stderr, "Error: %v\n", err); os.Exit(1) }

        exitCode, err := r.Run()
        if err != nil { fmt.Fprintf(os.Stderr, "Error: %v\n", err); os.Exit(1) }
        os.Exit(exitCode)
    },
}
```

### Anti-Patterns (Do NOT)

- Do NOT call `os.Exit()` inside `internal/runner/` — breaks testability
- Do NOT use `fmt.Fprintf(os.Stderr, ...)` inside runner — commands own output
- Do NOT import `internal/ui` from runner — TUI resolution stays in command files
- Do NOT hardcode `"claude"` in runner — it's a run-specific concern
- Do NOT duplicate logic between run.go and exec.go — exec is core, run is thin wrapper
- Do NOT use standard testing package for assertions — use testify
- Do NOT use `LookPath` for exec — command is user-specified directly

### Enforcement Guidelines

**All AI Agents MUST:**

- Follow the runner export pattern exactly as defined above
- Keep runner package free of I/O side effects (no os.Exit, no stderr, no stdout)
- Use `strings.HasPrefix` for ANTHROPIC_* filtering, never exact string matching
- Use testify for all test assertions (`require` for fatal, `assert` for non-fatal)
- Use table-driven test pattern for all new tests
- Check for empty string before injecting optional env vars

## Project Structure & Boundaries

### Complete Project Directory Structure

```
ccctx/
├── main.go                          # Entry point — MODIFY: add exec registration
├── go.mod                           # Module definition — MODIFY: add testify dependency
├── go.sum                           # Dependency checksums — AUTO-UPDATED
├── Makefile                         # Build automation — NO CHANGE
├── README.md                        # User documentation — MODIFY: add exec usage
├── CLAUDE.md                        # Claude Code instructions — NO CHANGE
├── LICENSE                          # MIT License — NO CHANGE
├── .gitignore                       # Git ignore rules — NO CHANGE
│
├── cmd/                             # CLI subcommands
│   ├── list.go                      # ccctx list — NO CHANGE
│   ├── exec.go                      # ccctx exec — NEW: core command (generic provider→env→exec)
│   └── run.go                       # ccctx run — REFACTOR: ultra-thin wrapper (exec -- claude)
│
├── config/                          # Configuration management
│   ├── config.go                    # TOML config loading — MODIFY: fix file permissions (0600)
│   └── config_test.go               # Unit tests — NO CHANGE
│
├── internal/                        # Internal packages (not importable externally)
│   ├── ui/
│   │   └── selector.go              # TUI interactive selector — NO CHANGE
│   └── runner/                      # Shared execution kernel — NEW PACKAGE
│       ├── runner.go                # Runner struct, New(), Run(), buildEnv()
│       ├── args.go                  # ParseArgs() — command-agnostic argument parsing
│       ├── runner_test.go           # Tests for Run() and buildEnv() — NEW
│       └── args_test.go             # Tests for ParseArgs() — NEW
│
├── examples/
│   └── config.toml                  # Sample configuration — NO CHANGE
│
└── docs/                            # Project documentation — NO CHANGE
```

### Architectural Boundaries

```
┌──────────────────────────────────────────────────────────┐
│                       main.go                             │
│            Root command + subcommand registration         │
└──────────┬──────────┬──────────┬─────────────────────────┘
           │          │          │
     ┌─────▼─────┐ ┌──▼──────┐ ┌▼──────────┐
     │ cmd/list  │ │ cmd/run │ │ cmd/exec  │
     │  (noop)   │ │ THIN    │ │  CORE     │
     │           │ │ WRAPPER │ │  COMMAND  │
     └─────┬─────┘ └──┬──────┘ └───┬───────┘
           │          │            │
           │          │    ┌───────▼────────────┐
           │          └───►│  internal/runner/   │◄── exec is the primary consumer
           │               │  ParseArgs()        │
           │               │  New() → Run()      │
           │               │  buildEnv()         │
           │               └──┬──────────┬──────┘
           │                  │          │
     ┌─────▼──────────────────▼──┐  ┌───▼──────────┐
     │  config/                   │  │ internal/ui/ │
     │  LoadConfig()              │  │ selector.go  │
     │  GetContext()               │  │ (TUI)        │
     │  ListContexts()             │  │              │
     └────────────────────────────┘  └──────────────┘
```

**Boundary Rules:**

| Package | May Import | May NOT Import |
|---------|-----------|---------------|
| `cmd/exec.go` | `config/`, `internal/runner/`, `internal/ui/`, stdlib | — |
| `cmd/run.go` | Same as exec + `exec.LookPath("claude")` | Must not duplicate exec logic |
| `internal/runner/` | `config/`, stdlib (`os`, `fmt`, `os/exec`, `strings`) | `internal/ui/`, `cmd/` |
| `config/` | stdlib, `viper` | `internal/runner/`, `internal/ui/`, `cmd/` |
| `internal/ui/` | `config/`, `tview`, `tcell` | `internal/runner/`, `cmd/` |

### Requirements to Structure Mapping

**Epic 1: Shared Execution Kernel & exec Subcommand**

| Story | Files Changed | Action |
|-------|--------------|--------|
| 1.1 Extract shared kernel | `internal/runner/runner.go`, `internal/runner/args.go` | NEW — extract from run.go |
| 1.2 Refactor run as thin wrapper | `cmd/run.go` | REFACTOR — ~30 lines, delegates to runner |
| 1.3 Implement exec (core command) | `cmd/exec.go`, `main.go` | NEW — generic provider→env→exec |
| 1.4 Config permissions | `config/config.go` | MODIFY — 0644→0600 |

**Epic 2: Interactive Provider Selection for exec**

| Story | Files Changed | Action |
|-------|--------------|--------|
| 2.1 exec integrates TUI | `cmd/exec.go` | MODIFY — add TUI flow (uses existing selector) |

**Epic 3: Model Override Flags (Phase 2)**

| Story | Files Changed | Action |
|-------|--------------|--------|
| 3.1 --model flags | `internal/runner/runner.go`, `cmd/exec.go`, `cmd/run.go` | MODIFY — add flag support |

**Cross-Cutting:**

| Concern | Location | Notes |
|---------|----------|-------|
| Env var management | `internal/runner/runner.go` → `buildEnv()` | Shared by run and exec |
| Exit code passthrough | `internal/runner/runner.go` → `Run()` | Returns int exit code |
| TUI selector | `internal/ui/selector.go` | Shared, called from cmd/ files only |
| Argument parsing | `internal/runner/args.go` → `ParseArgs()` | Command-agnostic, forwards remaining args without `--`, rejects multiple args before `--` |
| Tests | `internal/runner/runner_test.go`, `internal/runner/args_test.go` | testify-based table-driven |

### Integration Points

**Internal Communication:**

```
cmd/exec.go (CORE)                    cmd/run.go (WRAPPER)
    │                                      │
    ├─ runner.ParseArgs(args)              ├─ runner.ParseArgs(args)
    │   → provider, target, useTUI         │   → provider, target, useTUI
    │                                      │
    ├─ if useTUI:                          ├─ if useTUI:
    │   ui.RunContextSelector()            │   ui.RunContextSelector()
    │                                      │
    ├─ if target empty:                    ├─ LookPath("claude")
    │   target = [os.Getenv("SHELL")]      │
    │                                      ├─ target = [claudePath] + targetArgs
    ├─ runner.New(opts)                    │
    │   → config.GetContext()              ├─ runner.New(opts)
    │   → buildEnv()                       │   (same pipeline)
    │                                      │
    └─ runner.Run()                        └─ runner.Run()
        → exec.Command()                       (same execution)
        → (exitCode, error)
```

**External Integrations:**
- `claude` binary — discovered via `LookPath` (run command only)
- User-specified commands — executed directly via `exec.Command` (exec command)
- `$SHELL` environment variable — default shell for exec command
- `~/.ccctx/config.toml` — configuration file (or `CCCTX_CONFIG_PATH` override)

### File Change Summary

| Type | Count | Files |
|------|-------|-------|
| NEW | 5 | `cmd/exec.go`, `internal/runner/runner.go`, `internal/runner/args.go`, `internal/runner/runner_test.go`, `internal/runner/args_test.go` |
| REFACTOR | 1 | `cmd/run.go` (from ~200 lines to ~30 lines) |
| MODIFY | 2 | `main.go` (add exec registration), `config/config.go` (permissions) |
| NO CHANGE | all others | `cmd/list.go`, `internal/ui/selector.go`, `config/config_test.go`, `Makefile`, etc. |

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility:**
- AD1 (Runner struct) + AD2 (HasPrefix filtering) + AD3 (exec-as-core) — work seamlessly together
- AD4 (config permissions) — independent, no conflicts
- AD5 (testing) — testify framework compatible with all decisions

**Pattern Consistency:**
- Error handling pattern (runner returns `(int, error)`, commands handle output and exit) supports all 5 decisions
- Export patterns align with boundary rules — no circular dependencies
- Testing patterns (testify + table-driven) consistent across all new test files

**Structure Alignment:**
- `internal/runner/` cleanly separated from `cmd/` and `config/`
- Boundary rules prevent circular imports
- exec-as-core pattern reduces duplication between cmd files

**Review-Driven Updates Applied:**
- `Run()` signature changed to `(int, error)` — command start failures are now reportable
- ParseArgs forwards remaining args without `--` — preserves `ccctx run provider somearg` behavior
- Multiple args before `--` rejected — prevents silent arg dropping
- TUI cancellation exits with code 1 — consistent with abort semantics
- `$SHELL` emptiness validated — clear error instead of obscure Go runtime error
- `--model` vs `--` conflict documented for Phase 2 — won't block Phase 1

### Requirements Coverage Validation ✅

**Functional Requirements:**
- FR1-FR12: Already implemented, architecture preserves existing behavior (except `ccctx run unknown-arg` now hard-errors instead of TUI fallback — intentional breaking change)
- FR13-FR18: All covered by cmd/exec.go + internal/runner/ pipeline
- FR15: `$SHELL` emptiness validated with clear error message
- FR17: Exit code passthrough via `(int, error)` return — start failures are reported
- FR19, FR23: TUI selector reuse confirmed via useTUI flag
- FR24-FR26: Shared pipeline architecture via internal/runner/
- FR27-FR28: Deferred to Phase 2 (Options struct has Model/SmallFastModel fields ready; `--` separator conflict noted for Phase 2 resolution)

**Non-Functional Requirements:**
- NFR1-NFR3: Unchanged, runner doesn't leak tokens
- NFR4: Covered by AD4 (0600 on file creation)
- NFR5-NFR7: Preserved (POSIX $SHELL, multi-tool exec, CGO_ENABLED=0)

### Implementation Readiness Validation ✅

**Decision Completeness:**
- All 5 critical/important decisions documented with rationale
- Code templates provided for both command files (with error handling inline, no `handleError` helper)
- Export patterns explicitly defined
- Anti-patterns clearly listed
- Known Phase 2 issue (`--model` vs `--` separator) documented with resolution guidance

**Structure Completeness:**
- All new/existing/refactored files identified
- Boundary rules prevent import violations
- Integration points mapped with data flow diagrams

**Pattern Completeness:**
- Error handling, env var construction, argument parsing patterns all specified
- Test patterns with concrete testify examples
- 6 anti-patterns documented

### Gap Analysis Results

**No critical gaps found.**

**Minor items:**
- `handleError` helper function replaced with inline error handling in command templates — no longer needed
- `--model` flag implementation details deferred to Phase 2 per PRD — Options struct already has fields; known `--` separator conflict documented for Phase 2 resolution

### Architecture Completeness Checklist

**✅ Requirements Analysis**
- [x] Project context thoroughly analyzed
- [x] Scale and complexity assessed (low)
- [x] Technical constraints identified (Go, Cobra, Viper, CGO_ENABLED=0)
- [x] Cross-cutting concerns mapped (5 identified)

**✅ Architectural Decisions**
- [x] 5 decisions documented with rationale and impact
- [x] Technology stack verified (Cobra active, Viper upgrade noted)
- [x] Integration patterns defined (runner pipeline)
- [x] exec-as-core pattern established

**✅ Implementation Patterns**
- [x] Error handling patterns established — `Run()` returns `(int, error)`, commands handle I/O and exit
- [x] Export/visibility rules defined
- [x] Argument parsing patterns specified — forwards remaining args without `--`, rejects multiple args before `--`
- [x] Env var construction patterns documented — includes model priority chain
- [x] Testing patterns (testify + table-driven) defined
- [x] Edge cases from architecture review addressed (empty `$SHELL`, TUI exit code, start failure reporting)

**✅ Project Structure**
- [x] Complete directory structure with change annotations
- [x] Component boundaries with import rules
- [x] Integration points mapped (internal + external)
- [x] Requirements-to-structure mapping (all 3 epics + cross-cutting)

### Architecture Readiness Assessment

**Overall Status:** READY FOR IMPLEMENTATION

**Confidence Level:** High — brownfield project with low complexity, well-defined scope, existing tests as regression safety net.

**Key Strengths:**
- exec-as-core pattern eliminates duplication
- Runner struct naturally supports Phase 2 flags
- Clear boundary rules prevent implementation conflicts
- Existing codebase is small and well-understood

**Areas for Future Enhancement:**
- Viper v1.21.1 upgrade (post-feature work)
- `--model` / `--small-fast-model` flags (Phase 2)
- Shell completion support (Phase 3)

### Implementation Handoff

**AI Agent Guidelines:**
- Follow all architectural decisions exactly as documented
- exec is the core command, run is an ultra-thin wrapper
- Runner must be free of I/O side effects (no os.Exit, no stderr)
- `Run()` returns `(int, error)` — commands check error and print it
- Use testify for all test assertions
- Use `strings.HasPrefix("ANTHROPIC_")` for env var filtering
- ParseArgs forwards remaining args when no `--` is present
- ParseArgs returns error for multiple args before `--`
- TUI cancellation exits with code 1, not 0
- Validate `$SHELL != ""` before using it as default target
- First arg without `--` is always treated as provider name (no TUI fallback for unknown names — breaking change from current behavior)

**First Implementation Priority:**
1. Create `internal/runner/` package (runner.go + args.go)
2. Refactor `cmd/run.go` to thin wrapper
3. Add `cmd/exec.go` + register in main.go
4. Fix config file permissions in `config/config.go`
5. Add tests (runner_test.go + args_test.go)
