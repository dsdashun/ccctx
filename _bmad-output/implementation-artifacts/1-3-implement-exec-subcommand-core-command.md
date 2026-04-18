# Story 1.3: Implement exec Subcommand (Core Command)

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **ccctx user**,
I want **to launch a shell with provider env vars via `ccctx exec <provider>`, or execute an arbitrary command via `ccctx exec <provider> -- <command>`**,
so that **I can use any AI CLI tool or automation script with provider context, not just Claude Code**.

## Acceptance Criteria

1. **AC1: Shell launch with provider env** — `ccctx exec provider-A` starts `$SHELL` with provider-A's `ANTHROPIC_*` env vars; child processes inherit all provider env vars (FR13, FR15, FR16)

2. **AC2: Arbitrary command execution** — `ccctx exec provider-A -- env | grep ANTHROPIC` outputs correct `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN` values; command exit code is passed through (FR14, FR17)

3. **AC3: ANTHROPIC_* override** — `ccctx exec provider-A -- some-command` overrides any existing `ANTHROPIC_*` env vars with provider-A's values (FR18)

4. **AC4: Multi-arg command forwarding** — `ccctx exec provider-A -- bash -c "echo hello"` correctly passes the full command with arguments (FR14)

5. **AC5: $SHELL not set error** — When `$SHELL` is empty, `ccctx exec provider-A` (no `--`) outputs `Error: SHELL environment variable not set` to stderr with exit code 1 (FR15)

6. **AC6: No LookPath** — exec command does not use `exec.LookPath`; command is user-specified directly

7. **AC7: Registered in main.go** — `rootCmd.AddCommand(cmd.ExecCmd)` added to `main.go` init()

8. **AC8: Error format** — All errors use `Error: <lowercase description without trailing period>` format, output to stderr via `fmt.Fprintf(os.Stderr, ...)`

9. **AC9: All tests pass** — `go test ./...` passes, `go vet ./...` clean (FR13-FR18, FR26)

## Tasks / Subtasks

- [x] Task 1: Create `cmd/exec.go` (AC: #1-#6, #8)
  - [x] Create exported `ExecCmd` var with `&cobra.Command{Use: "exec [context] [-- command...]", Args: cobra.ArbitraryArgs}`
  - [x] Call `runner.ParseArgs(args)` for argument parsing
  - [x] Handle `useTUI` path: `config.ListContexts()` → empty check → `ui.RunContextSelector()` → cancellation handling
  - [x] Handle default target: when `targetArgs` empty, resolve `$SHELL` with emptiness validation
  - [x] Construct `runner.Options{ContextName: provider, Target: targetArgs}` and call `runner.New()` + `r.Run()`
  - [x] Handle exit code and errors with correct format

- [x] Task 2: Register exec command in main.go (AC: #7)
  - [x] Add `rootCmd.AddCommand(cmd.ExecCmd)` to `init()` in `main.go`

- [x] Task 3: Run tests and verify (AC: #9)
  - [x] `go test ./...` passes
  - [x] `go vet ./...` clean
  - [x] `make build` succeeds

### Review Findings

- [x] [Review][Defer] Fragile string comparison for TUI cancellation detection [cmd/exec.go:37] — deferred, pre-existing
- [x] [Review][Defer] TUI cancellation uses same exit code (1) as errors [cmd/exec.go:38-40] — deferred, pre-existing
- [x] [Review][Defer] Duplicate error handling pattern across commands [cmd/exec.go] — deferred, pre-existing
- [x] [Review][Defer] runner.Run double-prints non-ExitError failures [internal/runner/runner.go:51-57, cmd/exec.go:62-66] — deferred, pre-existing
- [x] [Review][Defer] ParseArgs treats flags before `--` as context names [internal/runner/args.go:30-34] — deferred, pre-existing
- [x] [Review][Defer] os.Exit calls bypass future defer cleanup [cmd/exec.go] — deferred, pre-existing
- [x] [Review][Defer] runner.New does not validate target executable exists [internal/runner/runner.go:37-39] — deferred, pre-existing

## Dev Notes

### Architecture Context

**exec is the core command; run is an ultra-thin wrapper (`run` ≡ `exec -- claude`).** The exec command is the generic provider→env→exec bridge. run.go exists only as a convenience wrapper that hardcodes `claude` as the target.

### cmd/exec.go — Full Template

Follow this template exactly (from architecture.md#Command File Template):

```go
package cmd

import (
    "fmt"
    "os"

    "github.com/dsdashun/ccctx/config"
    "github.com/dsdashun/ccctx/internal/runner"
    "github.com/dsdashun/ccctx/internal/ui"
    "github.com/spf13/cobra"
)

var ExecCmd = &cobra.Command{
    Use:   "exec [context] [-- command...]",
    Short: "Execute a command or launch a shell with a context",
    Long:  "Execute a command or launch a shell with the specified context. If no command is given, launches $SHELL. If no context is given, opens the interactive selector.",
    Args:  cobra.ArbitraryArgs,
    Run: func(cmd *cobra.Command, args []string) {
        provider, targetArgs, useTUI, err := runner.ParseArgs(args)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }

        if useTUI {
            contexts, err := config.ListContexts()
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
            }
            if len(contexts) == 0 {
                fmt.Fprintf(os.Stderr, "Error: no contexts found\n")
                os.Exit(1)
            }
            provider, err = ui.RunContextSelector()
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
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }

        exitCode, err := r.Run()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        os.Exit(exitCode)
    },
}
```

### Key Differences from cmd/run.go

| Aspect | cmd/run.go (wrapper) | cmd/exec.go (core) |
|--------|---------------------|---------------------|
| Default target | `exec.LookPath("claude")` | `os.Getenv("SHELL")` |
| LookPath | Yes (claude binary) | **NEVER** |
| Target construction | `append([]string{claudePath}, targetArgs...)` | `targetArgs` directly (or `[]string{shell}`) |
| $SHELL validation | Not needed | **Required** — must check `shell != ""` |

### main.go Modification

Add ONE line in `init()`:

```go
func init() {
    rootCmd.AddCommand(cmd.ListCmd)
    rootCmd.AddCommand(cmd.RunCmd)
    rootCmd.AddCommand(cmd.ExecCmd)  // ADD THIS LINE
}
```

### Project Structure Notes

- New file: `cmd/exec.go` — follows existing `cmd/run.go` pattern exactly
- Modified file: `main.go` — single line addition for command registration
- No changes to: `internal/runner/`, `config/`, `internal/ui/` — all shared logic already implemented in Stories 1.1 and 1.2

### Deferred Items (from previous story reviews)

These pre-existing issues are NOT in scope for this story:
- Fragile string comparison for TUI cancellation detection (`err.Error() == "operation cancelled"`) — pre-existing, deferred
- Redundant `config.ListContexts()` call in TUI path — deferred, Story 2.1 will address
- TUI list selection bounds check — deferred, pre-existing
- No panic recovery in TUI selector — deferred, pre-existing
- Race condition on TUI cancelled flag — deferred, pre-existing
- cmd/*.go has no automated tests — by design (architecture: "Command files are thin wrappers — no dedicated unit tests needed")
- runner.New does not validate BaseURL format — deferred, pre-existing
- Config file permissions 0644 — Story 1.4 scope

### References

- [Source: _bmad-output/planning-artifacts/architecture.md#AD3: exec Argument Parsing] — exec-as-core pattern, command-level differences
- [Source: _bmad-output/planning-artifacts/architecture.md#Command File Template] — exact exec.go template with error handling
- [Source: _bmad-output/planning-artifacts/architecture.md#Integration Points] — data flow diagram for exec command
- [Source: _bmad-output/planning-artifacts/architecture.md#Anti-Patterns] — no LookPath in exec, no I/O in runner
- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.3] — story requirements and acceptance criteria
- [Source: _bmad-output/project-context.md#Critical Don't-Miss Rules] — NEVER use LookPath for exec, NEVER duplicate logic between run.go and exec.go
- [Source: cmd/run.go] — reference implementation (ultra-thin wrapper pattern to mirror)
- [Source: internal/runner/runner.go] — Runner.New() and Run() API contract
- [Source: internal/runner/args.go] — ParseArgs() signature and behavior

## Dev Agent Record

### Agent Model Used

Claude Opus 4.7 (claude-opus-4-7)

### Debug Log References

### Completion Notes List

- ✅ Created `cmd/exec.go` following the exact template from architecture.md. The exec command is the core provider→env→exec bridge — no LookPath, uses $SHELL as default target, validates $SHELL emptiness.
- ✅ Registered `ExecCmd` in `main.go` init() with `rootCmd.AddCommand(cmd.ExecCmd)`.
- ✅ All tests pass, `go vet` clean, build succeeds.

### File List

- cmd/exec.go (new)
- main.go (modified)
- _bmad-output/implementation-artifacts/1-3-implement-exec-subcommand-core-command.md (modified)
- _bmad-output/implementation-artifacts/sprint-status.yaml (modified)
