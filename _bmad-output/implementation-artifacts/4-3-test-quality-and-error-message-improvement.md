# Story 4.3: Test Quality & Error Message Improvements

Status: done

## Story

As a **developer**,
I want **to clean up test quality issues deferred from Stories 4.1/4.2 and improve error message clarity**,
so that **the codebase meets consistent quality standards for test deduplication, error diagnosability, and boundary coverage, providing a clean foundation for Epic 5 and beyond**.

## Acceptance Criteria

1. **AC1: Shared test helpers for model flag tests** — `TestRunRun_ModelFlags` (run_test.go) and `TestExecRun_ModelFlags` (exec_test.go) share nearly identical test table structures, mock script construction, output parsing, and assertion logic. Extract common mock script setup, output file reading, and model-value assertions into shared helpers within each test file (or a test helper file in `cmd/`), eliminating near-duplicate code while keeping test tables separate. [Source: 4-2 review finding: "Near-identical test tables duplicated"]

2. **AC2: `validateFlagValue` error messages include flag name** — `validateFlagValue("--haiku-model", "foo\nbar")` currently returns `"--haiku-model value cannot contain newline"` which already includes the flag name. Verify ALL error paths from `ExtractFlags` consistently include the flag name in error messages. If any path omits it (e.g., the `"requires a value"` errors), ensure they include the flag name for consistency and diagnosability. [Source: epics.md#Story 4.3 AC2, args.go:8-13]

3. **AC3: Error-path tests for exec command** — `cmd/exec_test.go` currently only has `TestExecRun_ModelFlags` (model flag integration tests). Add a `TestExecRun` function covering the critical error paths that `cmd/run_test.go:TestRunRun` already covers: context not found, base_url missing, auth_token missing, invalid base_url format. These are the error paths exercised by `runner.New()` that exec shares with run. [Source: 4-2 review finding: "exec_test.go missing general error-path tests"]

4. **AC4: Unicode and long value tests for ExtractFlags** — Add test cases in `args_test.go` for `ExtractFlags` handling unicode flag values (e.g., `"claude-日本語"`) and long values (1000+ characters). These should process without crash and have test coverage confirming correct behavior. [Source: 4-2 review finding: "No tests for unicode or extremely long flag values"]

5. **AC5: `--haiku-model ""` edge case documented** — When `--haiku-model ""` (explicit empty string) and `--small-fast-model "X"` are both specified, the current implementation sets `haikuModel = ""` from `--haiku-model` (last-wins for haiku-model), and the post-loop alias resolution `if haikuModel == "" && sfmAlias != ""` then overrides it to `"X"`. Document this as a known limitation in a code comment in `args.go` near the alias resolution, or fix the behavior so `--haiku-model ""` explicitly wins over `--small-fast-model`. [Source: 4-2 review finding: "explicit empty can be silently overridden"]

6. **AC6: All tests pass** — `go test ./...` passes, `go vet ./...` clean, `make build` succeeds. No regressions. [Source: project-context.md#Testing Rules]

## Tasks / Subtasks

- [x] Task 1: Args improvements — `validateFlagValue` verification + `--haiku-model ""` edge case (AC: #2, #5)
  - [x] Review all error paths in `ExtractFlags` for flag-name inclusion; verify `"requires a value"` and `"value cannot contain newline"` are consistent
  - [x] Decide: fix the `--haiku-model ""` edge case (explicit empty wins) or document as known limitation
  - [x] **Write a failing test FIRST** for the `--haiku-model ""` edge case in `args_test.go` (e.g., `--haiku-model "" --small-fast-model "X"` should result in `haikuModel == ""`)
  - [x] If fix: implement `haikuModelSet` boolean in `ExtractFlags` to track whether `--haiku-model` was explicitly set; change alias resolution to `if !haikuModelSet && sfmAlias != ""`
  - [x] If document: add a code comment near the alias resolution explaining the known limitation
  - [x] Verify the failing test now passes; add/update tests if error message format changed during AC2 verification

- [x] Task 2: Add unicode and long value tests to `args_test.go` (AC: #4) **Depends on:** Task 1 (reads post-Task-1 state of args_test.go)
  - [x] Add test case: `--model` with unicode value (e.g., `"claude-日本語"`)
  - [x] Add test case: `--haiku-model` with 1000+ character value
  - [x] Add test case: `--sonnet-model` with unicode + special characters

- [x] Task 3: Extract shared test helpers for model flag tests (AC: #1)
  - [x] Identify duplicated logic between `TestRunRun_ModelFlags` and `TestExecRun_ModelFlags`: mock script, output parsing, assertion block
  - [x] Extract common helpers: `setupMockAndConfig`, `assertModelEnvVars`, etc.
  - [x] Refactor both test functions to use shared helpers
  - [x] Verify both test functions still pass with identical behavior

- [x] Task 4: Add error-path tests for exec command (AC: #3)
  - [x] Add `TestExecRun` function in `exec_test.go`
  - [x] Test cases: context not found, base_url missing, auth_token missing, invalid base_url format
  - [x] Use same pattern as `TestRunRun` in `run_test.go` but with explicit target command (no LookPath needed)

- [x] Task 5: Verify all tests pass (AC: #6)
  - [x] `go test ./...` — all pass
  - [x] `go vet ./...` — clean
  - [x] `make build` — succeeds

## Dev Notes

### What This Story Is About

Stories 4.1 and 4.2 added haiku/sonnet/opus model support across config, runner, and CLI layers. The code review for 4.2 identified several quality issues that were deferred to this story:

- Near-identical test tables duplicated between run_test.go and exec_test.go
- Error messages from `validateFlagValue` that could be more specific
- Missing error-path tests in exec_test.go
- No tests for unicode or long flag values
- Edge case: explicit empty `--haiku-model ""` silently overridden by `--small-fast-model`

### Current State of `validateFlagValue`

The function at `args.go:8-13` already includes the flag name in the error message:

```go
func validateFlagValue(name, value string) error {
    if strings.Contains(value, "\n") {
        return fmt.Errorf("%s value cannot contain newline", name)
    }
    return nil
}
```

This produces `"--haiku-model value cannot contain newline"` which is clear. The `"requires a value"` errors in each `case` branch also include the flag name (e.g., `"--haiku-model requires a value"`). AC2 may already be satisfied — verify during implementation.

### The `--haiku-model ""` Edge Case

Current implementation in `args.go:29,76,85-87`:

```go
var sfmAlias string
// ...
case "--small-fast-model":
    sfmAlias = preSep[i+1]    // last-wins for sfmAlias
    i += 2
case "--haiku-model":
    haikuModel = preSep[i+1]  // last-wins for haikuModel (including "")
    i += 2
// ...
if haikuModel == "" && sfmAlias != "" {
    haikuModel = sfmAlias
}
```

When `--haiku-model ""` is passed, `haikuModel` is set to `""` (explicit empty). Then if `--small-fast-model "X"` is also present, `sfmAlias = "X"`. Post-loop: `haikuModel == ""` is true, so `haikuModel = "X"` — the explicit empty is overridden.

**Recommended fix:** Track whether `--haiku-model` was explicitly used via a boolean `haikuModelSet`, and change the alias resolution to: `if !haikuModelSet && sfmAlias != ""`. This preserves all existing behavior while fixing the edge case.

### Test Deduplication Strategy

The shared logic between `TestRunRun_ModelFlags` and `TestExecRun_ModelFlags`:

1. **Config file setup** — identical (write TOML to temp config)
2. **Mock script** — identical 4-line shell script writing model env vars to `$MOCK_OUTPUT_FILE`
3. **Output parsing** — identical (read file, split lines, assert each model var)
4. **Differences** — run_test.go writes mock to hardcoded `claudePath`; exec_test.go has `hasExplicitTarget` branching for mock name

**Approach:** Extract helpers within `cmd/` package (same package, no import issues):
- `writeModelMockScript(t, dir, name)` — writes the mock script
- `assertModelOutput(t, outputFile, wantModel, wantHaiku, wantSonnet, wantOpus)` — reads and asserts
- Keep test tables and config setup inline (they define test cases)

### Exec Error-Path Tests

`TestRunRun` already covers these error paths via run_test.go. The exec command shares the same `runner.New()` validation, so equivalent tests are needed:

| Test Case | Config | Expected |
|-----------|--------|----------|
| context not found | `nonexistent` provider | exit code 1 |
| base_url missing | `[context.badctx] auth_token="t"` | exit code 1 |
| auth_token missing | `[context.noauth] base_url="https://x"` | exit code 1 |
| invalid base_url | `[context.badurl] base_url = "not-a-valid-url" auth_token = "t"` | exit code 1 |

> **Note:** All exec error-path tests must include `-- <command>` in args (e.g., `args: []string{"badurl", "--", "env"}`) to bypass `$SHELL` lookup and exercise `runner.New()` validation. Without this, tests hit `$SHELL unset` errors for the wrong reason.

### Files Changed

| File | Change | Scope |
|------|--------|-------|
| `internal/runner/args.go` | Fix `--haiku-model ""` edge case OR add documentation comment | MODIFY |
| `internal/runner/args_test.go` | Add unicode/long value tests + `--haiku-model ""` edge case test | MODIFY |
| `cmd/run_test.go` | Extract shared helpers, refactor `TestRunRun_ModelFlags` | MODIFY |
| `cmd/exec_test.go` | Extract shared helpers, refactor `TestExecRun_ModelFlags`, add `TestExecRun` error-path tests | MODIFY |

### Anti-Patterns (Do NOT)

- Do NOT change `ExtractFlags` signature or return values
- Do NOT change `ParseArgs`, `WantsHelp`, or `buildEnv` functions
- Do NOT change `runner.go` — no runner changes needed
- Do NOT change `config/config.go` — no config changes needed
- Do NOT change `cmd/run.go` or `cmd/exec.go` — no production code changes except potentially `args.go` for AC5
- Do NOT create a separate test helper package — keep helpers within `cmd/` package
- Do NOT break existing test cases — all 43 existing ExtractFlags tests must still pass
- Do NOT change error message format for existing working errors — only verify/improve if inconsistent

### Project Structure Notes

- Test helpers stay within `cmd/` package — Go convention allows test helpers in same package without exports
- `exec_test.go` error-path tests need explicit target commands (`-- env` or `-- bash`) to avoid `$SHELL` dependency, OR mock `$SHELL` like the model tests do
- The mock script for exec tests has `hasExplicitTarget` branching — the shared helper should accommodate both patterns

### References

- [Source: _bmad-output/implementation-artifacts/4-2-cli-flag-extension-backward-compatibility.md#Review Findings] — deferred items driving this story
- [Source: internal/runner/args.go] — ExtractFlags implementation, validateFlagValue, alias resolution
- [Source: internal/runner/args_test.go] — 43 existing test cases to preserve
- [Source: cmd/run_test.go] — TestRunRun (error paths) + TestRunRun_ModelFlags (model integration)
- [Source: cmd/exec_test.go] — TestExecRun_ModelFlags only (missing TestExecRun error paths)
- [Source: cmd/run.go] — runRun function (reference for error paths)
- [Source: cmd/exec.go] — execRun function (reference for error paths)
- [Source: internal/runner/runner.go:30-48] — New() validation logic (source of error-path tests)
- [Source: _bmad-output/planning-artifacts/epics.md#Story 4.3] — story definition and AC
- [Source: _bmad-output/project-context.md#Testing Rules] — testify + table-driven patterns

## Dev Agent Record

### Agent Model Used

Claude Opus 4.7

### Debug Log References

### Completion Notes List

- AC2: Verified all ExtractFlags error paths already include flag name in messages — no changes needed.
- AC5: Fixed `--haiku-model ""` edge case by adding `haikuModelSet` boolean tracker. Explicit empty `--haiku-model ""` now wins over `--small-fast-model` instead of being silently overridden. Two test cases added (both orderings).
- AC4: Added 3 new test cases — unicode model value (`claude-日本語`), 1000+ character haiku value, unicode + special characters for sonnet.
- AC1: Extracted `writeModelMock` and `assertModelOutput` shared helpers into `cmd/run_test.go` (same package). Refactored both `TestRunRun_ModelFlags` and `TestExecRun_ModelFlags` to use them, eliminating ~30 lines of duplicated mock setup and assertion code.
- AC3: Added `TestExecRun` in `exec_test.go` with 4 error-path test cases (context not found, base_url missing, auth_token missing, invalid base_url format), matching `TestRunRun` coverage.
- AC6: All tests pass (`go test ./...`), `go vet ./...` clean, `make build` succeeds. Total ExtractFlags tests: 48 (was 43).

### Review Findings

- [x] [Review][Defer] `TestExecRun` 错误路径测试缺少 mock 隔离 [cmd/exec_test.go] — deferred，与 TestRunRun 错误路径测试模式一致；`-- "env"` 正是用于绕过 `$SHELL` 查找的，runner.New() 校验先于 exec.LookPath
- [x] [Review][Defer] `TestExecRun` 未覆盖 `ParseArgs` 错误路径 [cmd/exec_test.go] — deferred，AC3 明确限定了 4 种错误路径
- [x] [Review][Defer] `TestExecRun` 未覆盖 `$SHELL` 默认目标路径 [cmd/exec_test.go] — deferred，不在 AC3 范围内
- [x] [Review][Defer] `validateFlagValue` 仅拒绝 `\n`，不拒绝 `\r` 等控制字符 [internal/runner/args.go:9] — deferred，已有行为，非本次改动引入

### File List

- `internal/runner/args.go` — Added `haikuModelSet` boolean to fix `--haiku-model ""` edge case
- `internal/runner/args_test.go` — Added 5 new test cases: 2 edge case + 3 unicode/long value
- `cmd/run_test.go` — Added `modelMockScript` constant, `writeModelMock` and `assertModelOutput` helpers; refactored `TestRunRun_ModelFlags` to use them
- `cmd/exec_test.go` — Refactored `TestExecRun_ModelFlags` to use shared helpers; added `TestExecRun` error-path tests
