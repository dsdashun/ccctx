# Story 1.1: Extract Runner Struct and ParseArgs into internal/runner/

Status: done

## Story

As a **developer**,
I want **the environment variable construction and process execution logic extracted from run.go into an internal/runner/ package using the Runner struct pattern (New/Run/buildEnv) and a command-agnostic ParseArgs function**,
so that **the run and exec commands can share the same execution pipeline, and the Runner struct naturally supports Phase 2 --model state passing**.

## Acceptance Criteria

1. **AC1: Runner package created** — `internal/runner/` package exists with two files:
   - `runner.go` — `Options` struct (with `ContextName`, `Target`, `Model`, `SmallFastModel` fields — Model/SmallFastModel reserved for Phase 2), `Runner` struct, `New(opts Options) (*Runner, error)`, `(*Runner) Run() (int, error)`, unexported `buildEnv()`
   - `args.go` — `ParseArgs(args []string) (provider string, targetArgs []string, useTUI bool, err error)`

2. **AC2: Logic extracted from run.go** — Runner struct encapsulates:
   - Env var building → `buildEnv()` using `strings.HasPrefix("ANTHROPIC_")` filtering (AD2), replacing the hardcoded length checks at current run.go:159-165
   - Process execution → `Run()` returning `(exitCode int, err error)` using `errors.As(err, &exitErr)` instead of type assertion `.(*exec.ExitError)` at current run.go:189
   - Argument parsing → `ParseArgs()` command-agnostic; default target resolved at command level, not inside ParseArgs

3. **AC3: New() validates required fields** — `New()` validates that `base_url` and `auth_token` are non-empty after resolving the context; returns error if empty

4. **AC4: buildEnv() injection order** — Injects in order: `ANTHROPIC_BASE_URL` (required) → `ANTHROPIC_AUTH_TOKEN` (required) → `ANTHROPIC_MODEL` (optional, only if non-empty) → `ANTHROPIC_SMALL_FAST_MODEL` (optional, only if non-empty); empty values never injected

5. **AC5: Runner has no I/O side effects** — Runner package never calls `os.Exit` or writes to stderr/stdout; all errors returned as values

6. **AC6: Existing tests pass** — `go test ./...` passes with zero regressions

7. **AC7: FR24, FR25 satisfied** — Shared execution pipeline exists; run uses it

## Tasks / Subtasks

- [x] Task 1: Create `internal/runner/` package directory (AC: #1)
  - [x] Create `internal/runner/runner.go` with Options struct, Runner struct, New(), Run(), buildEnv()
  - [x] Create `internal/runner/args.go` with ParseArgs()

- [x] Task 2: Extract argument parsing into ParseArgs (AC: #2, #2c)
  - [x] Implement ParseArgs with -- separator detection, provider extraction, useTUI flag
  - [x] Follow the parsing table from architecture: no args → useTUI=true; one arg → provider=args[0]; multiple args before -- → error; etc.
  - [x] ParseArgs is command-agnostic — no default target resolution inside

- [x] Task 3: Extract env var building into buildEnv() (AC: #2a, #4)
  - [x] Replace hardcoded length checks with `strings.HasPrefix(e, "ANTHROPIC_")`
  - [x] Injection order: ANTHROPIC_BASE_URL → ANTHROPIC_AUTH_TOKEN → ANTHROPIC_MODEL → ANTHROPIC_SMALL_FAST_MODEL
  - [x] Skip optional vars when empty string

- [x] Task 4: Extract process execution into Run() (AC: #2b)
  - [x] Return `(int, error)` — exit code + possible start error
  - [x] Use `errors.As(err, &exitErr)` with `*exec.ExitError` (not type assertion)
  - [x] Normal non-zero exit: return exit code with nil error
  - [x] Start failure: return 1 with the error

- [x] Task 5: Implement New() with validation (AC: #3)
  - [x] Resolve context via `config.GetContext()`
  - [x] Validate base_url and auth_token non-empty after resolution
  - [x] Call buildEnv() internally, store env in Runner struct
  - [x] Return error for any validation failure

- [x] Task 6: Ensure runner package purity (AC: #5)
  - [x] No os.Exit calls in runner package
  - [x] No fmt.Fprintf(os.Stderr, ...) in runner package
  - [x] No import of internal/ui
  - [x] No hardcoded "claude" string

- [x] Task 7: Update cmd/run.go to use new runner package (AC: #6, #7)
  - [x] Call runner.ParseArgs(args) for argument parsing
  - [x] Handle useTUI=true with existing TUI selector flow
  - [x] Keep exec.LookPath("claude") in run.go (not in runner)
  - [x] Build runner.Options and call runner.New() + r.Run()
  - [x] All existing tests pass

### Review Findings

- [x] [Review][Dismiss] Test files scope creep — dismissed by reviewer, tests kept

- [x] [Review][Patch] Empty Target slice causes panic in Run() [internal/runner/runner.go:36-37] — fixed: added len(opts.Target) validation in New()
- [x] [Review][Patch] cmd/run.go hard-exits on success paths [cmd/run.go] — fixed: only os.Exit when exitCode != 0
- [x] [Review][Patch] ParseArgs inconsistent nil vs empty slice [internal/runner/args.go] — fixed: return []string{} consistently; updated test
- [x] [Review][Patch] buildEnv allocates without capacity hint [internal/runner/runner.go:54] — fixed: preallocate with make([]string, 0, len(env))
- [x] [Review][Dismiss] go.mod tcell/v2 promoted from indirect to direct [go.mod] — dismissed: selector.go directly imports tcell, so direct dependency is correct

- [x] [Review][Defer] Fragile string equality for TUI cancellation check [cmd/run.go] — deferred, pre-existing

## Dev Notes

### Architecture Compliance

- **Runner struct pattern (AD1):** `Options` struct carries input, `Runner` struct carries resolved state. `New()` resolves and validates; `Run()` executes. This naturally supports Phase 2 `--model` flag by adding fields to `Options`.
- **HasPrefix filtering (AD2):** Current run.go:159-165 uses hardcoded length checks like `len(e) >= 19 && e[:19] == "ANTHROPIC_BASE_URL="`. Replace with `strings.HasPrefix(e, "ANTHROPIC_")` — catches all current and future ANTHROPIC_* vars.
- **ParseArgs is command-agnostic (AD3):** No cmdType parameter. Default target (claude for run, $SHELL for exec) is resolved at the command level after ParseArgs returns.

### Extraction Map — Current run.go → New Location

| Current run.go lines | New location | What |
|---|---|---|
| 154-178 (env building) | `internal/runner/runner.go` → `buildEnv()` | ANTHROPIC_* env var construction with HasPrefix filtering |
| 180-196 (exec + exit code) | `internal/runner/runner.go` → `Run()` | Process execution with `(int, error)` return using `errors.As` |
| 22-137 (arg parsing + TUI) | `internal/runner/args.go` → `ParseArgs()` | Shared argument parsing (TUI handling stays in cmd/) |
| 147-151 (LookPath) | `cmd/run.go` only | Binary discovery — exec command does NOT use LookPath |

### ParseArgs Behavior Table

| Condition | provider | targetArgs | useTUI | err |
|-----------|----------|-----------|--------|-----|
| `--` present, one arg before `--` | args[0] | args after `--` | false | nil |
| `--` present, no arg before `--` | "" | args after `--` | true | nil |
| `--` present, multiple args before `--` | — | — | — | error: ambiguous |
| No `--`, one arg | args[0] | [] | false | nil |
| No `--`, multiple args | args[0] | args[1:] | false | nil |
| No args at all | "" | [] | true | nil |

### Package Boundaries

| Package | May Import | May NOT Import |
|---------|-----------|---------------|
| `internal/runner/` | `config/`, stdlib (`os`, `fmt`, `os/exec`, `strings`, `errors`) | `internal/ui/`, `cmd/` |

### Error Message Format

All errors from runner follow: `fmt.Errorf("context 'name' description")` — lowercase, no trailing period. Examples:
- `context 'foo' not found` (from config.GetContext)
- `context 'foo' is missing base_url` (from New() validation)
- `at most one argument allowed before --` (from ParseArgs)

### Run() Exit Code Logic

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
        return 1, err // start failure
    }
    return 0, nil
}
```

### New() Implementation Pattern

```go
type Options struct {
    ContextName    string   // resolved provider name
    Target         []string // command to execute
    Model          string   // optional override (Phase 2, not used yet)
    SmallFastModel string   // optional override (Phase 2, not used yet)
}

func New(opts Options) (*Runner, error) {
    ctx, err := config.GetContext(opts.ContextName)
    if err != nil {
        return nil, err
    }
    if ctx.BaseURL == "" {
        return nil, fmt.Errorf("context '%s' is missing base_url", opts.ContextName)
    }
    if ctx.AuthToken == "" {
        return nil, fmt.Errorf("context '%s' is missing auth_token", opts.ContextName)
    }
    env := buildEnv(ctx, opts)
    return &Runner{ctx: ctx, opts: opts, env: env}, nil
}
```

### buildEnv() Implementation Pattern

```go
func buildEnv(ctx *config.Context, opts Options) []string {
    env := os.Environ()
    var filtered []string
    for _, e := range env {
        if !strings.HasPrefix(e, "ANTHROPIC_") {
            filtered = append(filtered, e)
        }
    }
    filtered = append(filtered, "ANTHROPIC_BASE_URL="+ctx.BaseURL)
    filtered = append(filtered, "ANTHROPIC_AUTH_TOKEN="+ctx.AuthToken)
    if ctx.Model != "" {
        filtered = append(filtered, "ANTHROPIC_MODEL="+ctx.Model)
    }
    if ctx.SmallFastModel != "" {
        filtered = append(filtered, "ANTHROPIC_SMALL_FAST_MODEL="+ctx.SmallFastModel)
    }
    return filtered
}
```

### How run.go Should Change

After extraction, `cmd/run.go` becomes the consumer of the runner package. Key changes:
1. Replace inline arg parsing (lines 22-137) with `runner.ParseArgs(args)`
2. Keep TUI flow in run.go — runner never imports ui
3. Keep `exec.LookPath("claude")` in run.go
4. Replace inline env building (lines 154-178) with `runner.New(opts)` which calls buildEnv internally
5. Replace inline execution (lines 180-196) with `r.Run()` which returns `(int, error)`

### Anti-Patterns (Do NOT)

- Do NOT call `os.Exit()` inside `internal/runner/` — breaks testability
- Do NOT use `fmt.Fprintf(os.Stderr, ...)` inside runner — commands own output
- Do NOT import `internal/ui` from runner — TUI resolution stays in command files
- Do NOT hardcode `"claude"` in runner — it's a run-specific concern
- Do NOT use `LookPath` in runner — stays in cmd/run.go only
- Do NOT validate context names in ParseArgs — it's command-agnostic; context validity checked in New() via config.GetContext()
- Do NOT duplicate logic between run.go and the runner package — once extracted, run.go should delegate, not keep a copy

### Current Code Reference Points

- **config.Context struct** — `config/config.go:12-17` — has BaseURL, AuthToken, Model, SmallFastModel fields with mapstructure tags
- **config.GetContext()** — `config/config.go:111-136` — resolves context by name, calls resolveEnvVar on auth_token
- **config.ListContexts()** — `config/config.go:97-109` — returns context names from config
- **ui.RunContextSelector()** — `internal/ui/selector.go:12-25` — currently takes no args, calls config.ListContexts internally; returns (string, error)
- **Current exit code handling** — `run.go:189` uses type assertion `.(*exec.ExitError)` — must change to `errors.As`
- **Current env filtering** — `run.go:159-165` uses hardcoded length checks — must change to `strings.HasPrefix`

### Important: RunContextSelector Signature

The current `ui.RunContextSelector()` takes no arguments and calls `config.ListContexts()` internally. The architecture shows a future pattern where contexts are passed in: `ui.RunContextSelector(contexts)`. For this story, maintain compatibility with the current signature. The signature change can happen in Story 1.2 or 2.1 when the TUI is shared.

### Behavioral Change Note

When run.go is updated to use `runner.ParseArgs()`, the first arg is always treated as a provider name — ParseArgs does NOT validate against known context names. This means `ccctx run nonexistent` will produce a hard error (`context 'nonexistent' not found` from `config.GetContext()`) instead of falling back to TUI. This is an **intentional** change documented in the architecture as a breaking change. Existing tests (`config/config_test.go`) are unaffected since they don't test cmd-level behavior. The change is safe because the new behavior (explicit error) is more predictable than silent TUI fallback. Story 1.2's AC lists these breaking changes for visibility, but they take effect when run.go is updated in this story.

### Testing Notes

- Existing tests in `config/config_test.go` must pass unchanged
- New tests for the runner package will be added in a later story (runner_test.go, args_test.go per AD5)
- This story focuses on extraction; dedicated test creation is part of Story 1.3 scope

### Project Structure Notes

New files to create:
```
internal/runner/
├── runner.go    # Options, Runner, New(), Run(), buildEnv()
└── args.go      # ParseArgs()
```

Existing file to modify:
```
cmd/run.go       # Use runner.ParseArgs + runner.New + r.Run()
```

### References

- [Source: _bmad-output/planning-artifacts/architecture.md#AD1] — Runner struct pattern decision
- [Source: _bmad-output/planning-artifacts/architecture.md#AD2] — HasPrefix filtering decision
- [Source: _bmad-output/planning-artifacts/architecture.md#AD3] — exec argument parsing, ParseArgs signature
- [Source: _bmad-output/planning-artifacts/architecture.md#Implementation Patterns] — Error handling, export visibility, anti-patterns
- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.1] — Story requirements and acceptance criteria
- [Source: _bmad-output/project-context.md#Critical Don't-Miss Rules] — Never os.Exit in runner, never hardcoded "claude", etc.
- [Source: cmd/run.go:154-178] — Current env building code to extract
- [Source: cmd/run.go:180-196] — Current execution code to extract
- [Source: cmd/run.go:22-137] — Current arg parsing code to extract
- [Source: config/config.go:12-17] — Context struct definition
- [Source: config/config.go:111-136] — GetContext() function used by New()

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

- Created `internal/runner/` package with two files: `runner.go` (Options, Runner, New, Run, buildEnv) and `args.go` (ParseArgs)
- ParseArgs is fully command-agnostic, handles all 8 cases from the architecture parsing table
- buildEnv uses `strings.HasPrefix("ANTHROPIC_")` filtering, injection order enforced (BASE_URL → AUTH_TOKEN → MODEL → SMALL_FAST_MODEL), empty optional vars skipped
- Run() uses `errors.As(err, &exitErr)` instead of type assertion, returns (exitCode, nil) for normal exits and (1, err) for start failures
- New() validates base_url and auth_token after context resolution via config.GetContext()
- Runner package has zero I/O side effects: no os.Exit, no stderr writes, no ui import, no hardcoded "claude"
- Updated cmd/run.go to delegate to runner: ParseArgs for arg parsing, New() for context resolution + env building, Run() for execution
- All existing tests pass (config, runner), zero regressions, go vet clean

### File List

- `internal/runner/runner.go` — NEW: Options struct, Runner struct, New(), Run(), buildEnv()
- `internal/runner/args.go` — NEW: ParseArgs()
- `internal/runner/runner_test.go` — NEW: Tests for buildEnv (filtering, skip empty, injection order)
- `internal/runner/args_test.go` — NEW: Table-driven tests for ParseArgs (8 cases)
- `cmd/run.go` — MODIFIED: Refactored to use runner.ParseArgs, runner.New, r.Run()
- `go.mod` — MODIFIED: Added testify dependency
- `go.sum` — MODIFIED: Updated checksums

### Change Log

- 2026-04-18: Extracted execution kernel into internal/runner/ package — runner.go (New/Run/buildEnv), args.go (ParseArgs). Updated cmd/run.go to delegate to runner. All tests pass.
