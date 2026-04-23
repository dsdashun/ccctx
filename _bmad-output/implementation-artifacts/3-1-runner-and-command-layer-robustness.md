# Story 3.1: Runner & Command Layer Robustness

Status: done

## Story

As a **developer**,
I want **runner 包和 cmd 包的健壮性问题得到修复，包括输入验证、错误处理、flag 解析和测试覆盖**,
So that **后续 Story 3.2 的 --model flag 添加能在稳固的基础上进行，且 ParseArgs 的 flag 处理方式不会与 Cobra flag 冲突**.

## Acceptance Criteria

1. **AC1: BaseURL format validation in runner.New()** — `New()` validates that `BaseURL` is a valid URL (has scheme, no spaces). Invalid URLs return a clear validation error before creating a Runner. [Source: deferred-work.md — "runner.New does not validate BaseURL format"]

2. **AC2: No double-printing of non-ExitError failures** — When `runner.Run()` returns a non-ExitError (e.g., binary not found), the error message is printed exactly once by the caller (`cmd/*.go`). The runner itself never prints to stderr/stdout. [Source: deferred-work.md — "runner.Run double-prints non-ExitError failures"]

3. **AC3: ParseArgs rejects or handles flag-like args in provider position** — When `ParseArgs` receives args starting with `-` in the provider position (before `--`), it either rejects them with a clear error or handles them in a defined way that Story 3.2 can build on. This is the **critical prerequisite** for adding `--model` flags. [Source: deferred-work.md — "ParseArgs treats flags before `--` as context names"]

4. **AC4: No direct os.Exit in cmd/exec.go error paths** — Error handling in `cmd/exec.go` uses `return` with exit code propagation instead of direct `os.Exit()`, ensuring any deferred cleanup functions execute. `os.Exit()` is called only at the top level of the `Run` function. [Source: deferred-work.md — "os.Exit calls bypass future defer cleanup"]

5. **AC5: cmd/run.go table-driven tests** — New test file `cmd/run_test.go` with table-driven tests covering: success path (provider found, claude found), context not found, TUI cancellation, LookPath failure. Uses testify framework. [Source: deferred-work.md — "cmd/run.go has no automated tests"]

6. **AC6: All tests pass** — `go test ./...` passes, `go vet ./...` clean, `make build` succeeds.

7. **AC7: Existing behavior unchanged** — `ccctx run <provider>`, `ccctx exec <provider>`, `ccctx run` (TUI), `ccctx exec` (TUI) all work identically to current behavior.

## Tasks / Subtasks

- [x] Task 1: Add URL validation in runner.New() (AC: #1)
  - [x] RED: Write table-driven tests in `runner_test.go` first — test `validateURL` directly (cases: valid URL, missing scheme, contains spaces, empty string) and optionally test `New()` with invalid URLs. For `New()` tests: create a temp TOML config file with contexts containing invalid `base_url` values, and use `t.Setenv("CCCTX_CONFIG_PATH", tempFile)` before calling `New()`. Alternatively, focus RED tests on `validateURL` directly and defer `New()` integration tests to a separate test function. Run tests, confirm failures (function doesn't exist yet).
  - [x] GREEN: Import `net/url`, implement `validateURL` helper (parse URL, check scheme non-empty, check no spaces), wire into `New()` after resolving context, before `buildEnv()`. Run tests, confirm all pass.
  - [x] REFACTOR: Clean up if needed.

- [x] Task 2: Verify and document single-print error contract (AC: #2)
  - [x] Audit `runner.Run()` — already correct: returns `(int, error)` without printing
  - [x] Audit `cmd/exec.go` and `cmd/run.go` — confirm each error path prints exactly once
  - [x] Minimal code change: verify the existing contract via audit, then add a documentation comment in `runner.go` near `Run()`
  - [x] Add a comment in `runner.go` near `Run()`: `// Run executes the target command. Returns (0, nil) on success, (exitCode, nil) for command exit errors, (1, error) for start failures. Caller is responsible for printing errors.`

- [x] Task 3: Update ParseArgs to handle flag-like args (AC: #3)
  - [x] RED: Add failing test cases in `args_test.go` for flag-like args in provider position: `["--model", "foo"]` (wantErr contains `"flag-like argument"`), `["-m"]` (wantErr contains `"flag-like argument '-m'"`), `["--model", "--", "foo"]` (wantErr contains `"flag-like argument '--model'"`). Each test expects a non-empty error containing the flag name. Run tests, confirm failures.
  - [x] GREEN: In `ParseArgs`, when the resolved provider argument (the single argument before `--`, or `args[0]` when `--` is absent) starts with `-`, return an error: `"flag-like argument '%s' not allowed in provider position"`. Run tests, confirm all pass.
  - [x] REFACTOR: Ensure error message consistency.
  - [x] Note: This makes `ccctx run --model` fail explicitly instead of treating `--model` as a provider name. Story 3.2 will change this behavior to accept and parse flags — but the current rejection establishes the pattern.

- [x] Task 4: Refactor cmd/exec.go to avoid direct os.Exit in error paths (AC: #4)
  - [x] Extract the `Run` function body into a helper that returns exit code (int)
  - [x] Call `os.Exit(result)` once at the top level
  - [x] Same pattern for `cmd/run.go` if it has the same issue
  - [x] Ensure deferred functions (if any) execute before exit

- [x] Task 5: Add cmd/run_test.go (AC: #5)
  - [x] Refactor `cmd/run.go` to extract `runRun(args []string) int` (same pattern as Task 4)
  - [x] Test `runRun` directly — it returns exit code, no `os.Exit` in test path
  - [x] Use `CCCTX_CONFIG_PATH` env var to point at a temp test config file with known contexts
  - [x] For LookPath testing: create a temp dir, symlink/write a fake `claude` binary, set `PATH` via `t.Setenv("PATH", tmpDir)`
  - [x] Test cases: success path (provider found, claude found), context not found, LookPath failure, ParseArgs error
  - [x] Note: TUI cancel path cannot be unit-tested because `ui.RunContextSelector` requires an interactive terminal (`tview.Application.Run()`). The cancel path is covered by manual smoke testing in Task 6. The `errors.Is(err, ui.ErrCancelled)` branch in `runRun` is a simple 3-line if-block — low risk.
  - [x] Note: Use `package cmd` (not `package cmd_test`) in `cmd/run_test.go` so tests can call the unexported `runRun` function.
  - [x] Note: cmd/exec.go tests can be deferred — this story focuses on run.go since it was explicitly called out

- [x] Task 6: Verify all tests and build (AC: #6, #7)
  - [x] `go test ./...` passes
  - [x] `go vet ./...` clean
  - [x] `make build` succeeds
  - [x] Manual smoke test: `ccctx run`, `ccctx exec`, argument forwarding

## Dev Notes

### Current State

Epic 1 and Epic 2 are complete. The runner package (`internal/runner/`) and command files (`cmd/run.go`, `cmd/exec.go`) work correctly for all current use cases. This story addresses deferred code review items that are prerequisites for adding `--model` flag support in Story 3.2.

### Key Files

| File | Role | Change Type |
|------|------|-------------|
| `internal/runner/runner.go` | Runner struct, New(), Run(), buildEnv() | MODIFY — add URL validation |
| `internal/runner/args.go` | ParseArgs() | MODIFY — reject flag-like args in provider position |
| `internal/runner/runner_test.go` | Tests for buildEnv | MODIFY — add URL validation tests |
| `internal/runner/args_test.go` | Tests for ParseArgs | MODIFY — add flag-like arg test cases |
| `cmd/exec.go` | exec subcommand | MODIFY — refactor os.Exit pattern |
| `cmd/run.go` | run subcommand | MODIFY — refactor os.Exit pattern |
| `cmd/run_test.go` | Tests for run command | NEW — table-driven tests |

### URL Validation Strategy

Use `net/url.Parse()` to validate BaseURL in `runner.New()`:

```go
func validateURL(rawURL string) error {
    if strings.Contains(rawURL, " ") {
        return fmt.Errorf("invalid base_url: contains spaces")
    }
    u, err := url.Parse(rawURL)
    if err != nil {
        return fmt.Errorf("invalid base_url: %w", err)
    }
    if u.Scheme == "" {
        return fmt.Errorf("invalid base_url: missing scheme (e.g., https://)")
    }
    return nil
}
```

Insert in `New()` between the `ctx.AuthToken == ""` check (line 34) and the `len(opts.Target) == 0` check (line 37). This validates the URL before `buildEnv()` uses it.

### ParseArgs Flag Handling Strategy

The critical design question: how should `ParseArgs` treat `--model foo`?

**Current behavior:** `ParseArgs(["--model", "foo"])` returns `provider="--model", targetArgs=["foo"]` — treats `--model` as a provider name, which will fail at `config.GetContext("--model")` with a confusing error.

**Story 3.1 approach:** Reject flag-like args explicitly. When the resolved provider argument (the single argument before `--`, or `args[0]` when `--` is absent) starts with `-`, return an error: `"flag-like argument '%s' not allowed in provider position"`. This:
- Gives a clear error message instead of confusing "context not found"
- Establishes the pattern that flags in provider position are special
- Story 3.2 will change this to parse `--model` and `--small-fast-model` before calling ParseArgs (likely via `cmd.DisableFlagParsing = true` with manual flag extraction)

### os.Exit Refactor Pattern

```go
// BEFORE (cmd/exec.go):
Run: func(cmd *cobra.Command, args []string) {
    // ... error handling with direct os.Exit(1) calls ...
    os.Exit(exitCode)
},

// AFTER:
Run: func(cmd *cobra.Command, args []string) {
    os.Exit(execRun(args))
},

func execRun(args []string) int {
    // ... all logic with return 1 on error, return exitCode on success ...
}
```

This pattern ensures any `defer` statements in `execRun` execute before `os.Exit`. Currently there are no defers, but this establishes the pattern for future use (e.g., cleanup in Story 3.2).

### cmd/run_test.go Testing Strategy

Testing `cmd/run.go` is challenging because it calls `os.Exit`. Options:

1. **Extract logic into testable function** (recommended) — same pattern as Task 4. Extract `runRun(args []string) int` and test that function directly.
2. **Use os/exec to run the binary** — integration test approach, slower but realistic.

Recommend approach 1: after Task 4 refactors `run.go` to extract `runRun(args) int`, test `runRun` directly by:
- Setting up test config via `CCCTX_CONFIG_PATH`
- Mocking `exec.LookPath` by controlling PATH
- Testing TUI paths by checking the error flow

### Deferred Work Items Source

All items in this story come from `deferred-work.md` and were promoted from code reviews of Stories 1.1-1.4:

| Deferred Item | Source Review | AC |
|---------------|---------------|-----|
| runner.New does not validate BaseURL format | Story 1-1 review | AC1 |
| runner.Run double-prints non-ExitError failures | Story 1-3 review | AC2 |
| ParseArgs treats flags before `--` as context names | Story 1-3 review | AC3 |
| os.Exit calls bypass future defer cleanup | Story 1-3 review | AC4 |
| cmd/run.go has no automated tests | Story 1-1 review | AC5 |

### Previous Story Intelligence

From Story 2-2 implementation:
- TUI selector uses sentinel error `ErrCancelled` — tested via `errors.Is` in both cmd files
- Cancellation messages moved to stderr in both `cmd/run.go:42` and `cmd/exec.go:39`
- All error output uses `fmt.Fprintf(os.Stderr, ...)` or `fmt.Fprintln(os.Stderr, ...)`

From Story 1-3 implementation:
- Runner struct carries `ctx`, `opts`, `env` — `New()` validates BaseURL and AuthToken non-empty
- `Run()` returns `(int, error)` — ExitError returns exit code, other errors return `(1, err)`
- `ParseArgs()` is command-agnostic — no cmdType parameter
- Both `cmd/run.go` and `cmd/exec.go` follow identical error handling pattern

From Story 1-2 implementation:
- `cmd/run.go` is ~74 lines — thin wrapper calling `runner.ParseArgs`, TUI, `LookPath`, `runner.New`, `runner.Run`
- Error format: `fmt.Fprintf(os.Stderr, "Error: %v\n", err)` — consistent across both commands

### Anti-Patterns (Do NOT)

- Do NOT add `--model` flag support in this story — that's Story 3.2
- Do NOT modify `internal/ui/selector.go` — not in scope
- Do NOT modify `config/config.go` — not in scope
- Do NOT call `os.Exit()` inside `internal/runner/` — breaks testability
- Do NOT use `fmt.Fprintf(os.Stderr, ...)` inside runner — commands own output
- Do NOT over-engineer the test mocking — keep tests simple and direct
- Do NOT change the `buildEnv` injection order — it's established and tested
- Do NOT change TUI cancellation exit code from 1 to 0 — this was an intentional breaking change applied in Epic 2

### Project Structure Notes

- All changes are in existing packages (`internal/runner/`, `cmd/`)
- New file `cmd/run_test.go` follows the established test naming convention
- URL validation uses stdlib `net/url` — no new dependencies
- Go 1.23.3 target — `min()`/`max()` builtins available if needed

### References

- [Source: internal/runner/runner.go:31-36] — New() validation, needs URL check
- [Source: internal/runner/runner.go:44-60] — Run() error return pattern
- [Source: internal/runner/args.go:30-34] — ParseArgs "no args" path, needs flag detection
- [Source: cmd/exec.go:19-69] — exec command, needs os.Exit refactor
- [Source: cmd/run.go:20-74] — run command, needs os.Exit refactor + tests
- [Source: internal/runner/runner_test.go] — existing buildEnv tests
- [Source: internal/runner/args_test.go] — existing ParseArgs tests
- [Source: _bmad-output/implementation-artifacts/deferred-work.md] — deferred items source
- [Source: _bmad-output/planning-artifacts/epics.md#Story 3.1] — epic story definition
- [Source: _bmad-output/planning-artifacts/architecture.md#AD1] — Runner struct design
- [Source: _bmad-output/planning-artifacts/architecture.md#Phase 2] — known --model vs -- conflict
- [Source: _bmad-output/project-context.md] — implementation rules and patterns

## Dev Agent Record

### Agent Model Used

GLM-5.1

### Debug Log References

### Completion Notes List

- Task 1: Added `validateURL()` helper using `net/url.Parse()` with scheme and space checks. Wired into `New()` between BaseURL empty check and AuthToken check. 7 table-driven test cases pass.
- Task 2: Audited `runner.Run()`, `cmd/exec.go`, `cmd/run.go` — single-print contract already correct. Added documentation comment on `Run()` method.
- Task 3: Added flag-like arg detection in `ParseArgs` — when provider position arg starts with `-`, returns clear error message. 3 new test cases (double dash, single dash, with separator).
- Task 4: Extracted `execRun(args) int` and `runRun(args) int` helpers from Cobra `Run` closures. `os.Exit()` called only at top level now.
- Task 5: Created `cmd/run_test.go` with 7 table-driven test cases covering ParseArgs error, context not found, LookPath failure, success, non-zero exit code, missing base_url, invalid URL format.
- Task 6: All 32 tests pass, `go vet` clean, `make build` succeeds.

### Review Findings

- [x] [Review][Patch] `cmd/run_test.go` 缺少带 `--` 分隔符和转发参数的成功路径测试 [cmd/run_test.go] — 应添加 `args: []string{"test", "--", "--model", "foo"}` 等用例验证参数正确转发到 claude
- [x] [Review][Patch] `cmd/run_test.go` 缺少 auth_token 缺失的测试用例 [cmd/run_test.go] — 现有测试覆盖 missing base_url 和 invalid base_url，但未覆盖 auth_token 为空的情况
- [x] [Review][Defer] 跨平台兼容性：Windows 上假 claude 脚本使用 shebang 不工作 [cmd/run_test.go] — deferred, pre-existing
- [x] [Review][Defer] buildEnv 不验证注入环境变量值中的换行或空字节 [internal/runner/runner.go:82-99] — deferred, pre-existing
- [x] [Review][Defer] config 对 base_url 不支持 env: 前缀解析（与 auth_token 不一致） [config/config.go] — deferred, pre-existing
- [x] [Review][Defer] SHELL 环境变量可执行性未验证 [cmd/exec.go:52-58] — deferred, pre-existing
- [x] [Review][Defer] 目标命令被信号终止时退出码丢失 [internal/runner/runner.go:64-80] — deferred, pre-existing
- [x] [Review][Defer] 空 provider 名未在 ParseArgs 中显式拒绝 [internal/runner/args.go] — deferred, pre-existing
- [x] [Review][Defer] 通用 "Error: %v" 消息模式不利于程序化错误区分 [cmd/run.go, cmd/exec.go] — deferred, pre-existing

### File List

- `internal/runner/runner.go` — MODIFIED: added `validateURL()`, imported `net/url`, wired into `New()`, added `Run()` doc comment
- `internal/runner/args.go` — MODIFIED: added flag-like arg detection in provider position, imported `strings`
- `internal/runner/runner_test.go` — MODIFIED: added `TestValidateURL` with 7 test cases, added `require` import
- `internal/runner/args_test.go` — MODIFIED: added 3 flag-like arg test cases
- `cmd/exec.go` — MODIFIED: extracted `execRun()` helper, `os.Exit()` only at top level
- `cmd/run.go` — MODIFIED: extracted `runRun()` helper, `os.Exit()` only at top level
- `cmd/run_test.go` — NEW: 7 table-driven test cases for run command

### Change Log

- 2026-04-20: Implemented all 6 tasks — URL validation (AC1), single-print contract audit (AC2), flag-like arg rejection (AC3), os.Exit refactor (AC4), run_test.go (AC5), full validation (AC6/AC7)
