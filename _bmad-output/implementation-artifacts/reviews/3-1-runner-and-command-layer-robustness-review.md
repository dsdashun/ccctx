---
reviewed_plan: _bmad-output/implementation-artifacts/3-1-runner-and-command-layer-robustness.md
reviewer: kimi
latest_round: 3
review_rounds:
  - round: 1
    date: 2026-04-20
  - round: 2
    date: 2026-04-20
  - round: 3
    date: 2026-04-20
---

# Review: 3-1-runner-and-command-layer-robustness

Reviewed plan: `_bmad-output/implementation-artifacts/3-1-runner-and-command-layer-robustness.md`

## Round 1

**Context:** Initial full dry-run of the plan against the actual codebase.

### Issues Found

#### 1. [BLOCKING] Task 1 and Task 3 violate TDD: implementation before tests

- **Plan says (Task 1):** Subtask order: (a) import `net/url`, (b) add `validateURL` helper, (c) call it in `New()`, (d) add table-driven tests in `runner_test.go`.
- **Actual:** This is the "test-after" anti-pattern. Tests are written after the implementation is complete, so they will pass immediately and never prove they can catch real bugs.
- **Impact:** If the `validateURL` implementation has edge-case bugs (e.g., `url.Parse` accepting `://example.com` as valid with empty scheme but non-empty opaque part), the tests written afterward may simply encode those same bugs. The tests provide no regression safety.
- **Suggested fix:** Restructure Task 1 into Red-Green-Refactor:
  - **RED:** Write table-driven tests in `runner_test.go` first that call a non-existent `validateURL` (or call `New()` with invalid URLs). Run tests, confirm failures.
  - **GREEN:** Implement `validateURL` and wire it into `New()`.
  - **REFACTOR:** Clean up if needed.

- **Plan says (Task 3):** Subtask order: (a) modify `ParseArgs` to reject flag-like args, (b) add test cases in `args_test.go`.
- **Actual:** Same anti-pattern. `args_test.go` tests are added after the behavior change.
- **Impact:** New test cases like `["--model", "foo"]` and `["-m"]` will pass immediately because the implementation is already written. If the `ParseArgs` logic incorrectly handles edge cases (e.g., `["-", "--", "cmd"]` where `-` is a single-dash non-flag), tests will encode the bug.
- **Suggested fix:** Restructure Task 3:
  - **RED:** Add failing test cases in `args_test.go` for flag-like args: `["--model", "foo"]`, `["-m"]`, `["--model", "--", "foo"]`.
  - **GREEN:** Modify `ParseArgs` to reject providers starting with `-`.
  - **REFACTOR:** Ensure error message consistency.

#### 2. [HIGH] Task 5 "TUI cancel (ErrCancelled)" test case is not feasible with current architecture

- **Plan says:** Test cases include "TUI cancel (ErrCancelled)" in `cmd/run_test.go`, testing `runRun` directly.
- **Actual:** When `useTUI` is true, `runRun` calls `ui.RunContextSelector(contexts)`, which internally creates a `tview.Application` and calls `app.Run()`. This requires an interactive terminal and cannot run in a `go test` environment. There is no dependency-injection point to mock the selector.
- **Impact:** The executor will attempt to write this test, discover it hangs or fails in CI, and either (a) skip it, leaving a gap in coverage, or (b) spend significant time trying to make it work. Either way, the plan's stated test coverage is unachievable as written.
- **Suggested fix:** Either:
  - **Option A (recommended):** Add a `selectorFunc` parameter to `runRun` (e.g., `runRun(args []string, selectContext func([]string) (string, error)) int`), and pass `ui.RunContextSelector` from the command's `Run` closure. In tests, inject a mock that returns `ui.ErrCancelled`.
  - **Option B:** Explicitly exclude the TUI cancel path from `cmd/run_test.go` scope, and add a comment documenting the untested path. Rely on manual smoke testing (Task 6) for TUI coverage.

#### 3. [MEDIUM] Task 3 description of flag detection is ambiguous

- **Plan says:** "In `ParseArgs`, when the first arg before `--` starts with `-`, return an error."
- **Actual:** When there is no `--` separator (e.g., `["--model", "foo"]`), the concept of "before `--`" does not exist. The current `ParseArgs` has two code paths: (1) `--` present, (2) `--` absent. The plan's wording conflates both paths under a single imprecise description. The intended behavior is clear from the test cases (reject `args[0]` starting with `-` in the no-separator path, and reject `contextArgs[0]` starting with `-` in the separator path), but a literal reading of the step could lead to inconsistent implementations.
- **Impact:** Different executors may implement the check in different places within `ParseArgs`, leading to divergent behavior for edge cases like `["--model", "--", "foo"]` vs `["--model", "foo"]`.
- **Suggested fix:** Clarify the Task 3 description:
  > "In `ParseArgs`, when the resolved provider argument (the single argument before `--`, or `args[0]` when `--` is absent) starts with `-`, return an error: `flag-like argument '%s' not allowed in provider position`.

#### 4. [LOW] Minor line number drift in Dev Notes references

- **Plan says:** "Insert in `New()` between the `ctx.AuthToken == ""` check (line 36) and the `len(opts.Target) == 0` check (line 37)."
- **Actual:** In the current `internal/runner/runner.go`, `ctx.AuthToken == ""` is at line 34, and `len(opts.Target) == 0` is at line 37. The content descriptions are correct, but the line numbers are off by two.
- **Impact:** Minimal — the surrounding context (`ctx.AuthToken == ""` / `len(opts.Target) == 0`) makes the insertion point unambiguous.
- **Suggested fix:** Update line references to `line 34` and `line 37`.

#### 5. [LOW] `runner.go` `Run()` comment in Task 2 slightly mischaracterizes the contract

- **Plan says (Task 2):** Add comment: `// Run executes the target command. Returns (exitCode, nil) for normal exits, (1, error) for start failures. Caller is responsible for printing errors.`
- **Actual:** `Run()` returns `(exitErr.ExitCode(), nil)` for command exit errors (non-zero exit codes), not just "normal exits" (which implies exit code 0). The comment could mislead a reader into thinking `(exitCode, nil)` only happens on success.
- **Impact:** A future maintainer might add redundant error handling for non-zero exit codes.
- **Suggested fix:** Adjust the comment to: `// Run executes the target command. Returns (0, nil) on success, (exitCode, nil) for command exit errors, (1, error) for start failures. Caller is responsible for printing errors.`

### Risk Summary

| # | Severity | Issue | Task |
|---|----------|-------|------|
| 1 | **BLOCKING** | TDD violation: implementation before tests in Task 1 and Task 3 | Task 1, Task 3 |
| 2 | HIGH | TUI cancel test not feasible without DI/mocking | Task 5 |
| 3 | MEDIUM | Ambiguous flag detection description | Task 3 |
| 4 | LOW | Line number drift in Dev Notes | Dev Notes |
| 5 | LOW | `Run()` comment mischaracterizes non-zero exit path | Task 2 |

### Verdict

The plan is **NOT ready for execution**. Fix the **blocking issue (#1)** — restructure Tasks 1 and 3 to follow Red-Green-Refactor (write failing tests first, then implementation). Also address the **high-severity issue (#2)** by either introducing dependency injection for the TUI selector or explicitly excluding the TUI cancel path from unit test scope. Issues #3, #4, and #5 should also be addressed but are not hard blockers.

## Round 1 Response

**Responded by:** glm
**Date:** 2026-04-20
**Plan file:** 3-1-runner-and-command-layer-robustness.md (updated)

### Issue Responses

#### 1. [FIX] Task 1 and Task 3 violate TDD: implementation before tests

**Change made:** Restructured both Task 1 and Task 3 into explicit Red-Green-Refactor subtasks. Task 1 now starts with "RED: Write table-driven tests first", then "GREEN: Implement validateURL", then "REFACTOR". Task 3 follows the same pattern with "RED: Add failing test cases", "GREEN: Modify ParseArgs", "REFACTOR".
**Location:** Tasks 1 and 3 in the plan file

#### 2. [FIX] TUI cancel (ErrCancelled) test case is not feasible with current architecture

**Change made:** Removed "TUI cancel (ErrCancelled)" from Task 5's test cases. Added a note explaining why: `ui.RunContextSelector` requires an interactive terminal via `tview.Application.Run()`. The cancel path is covered by manual smoke testing in Task 6 instead. The `errors.Is(err, ui.ErrCancelled)` branch is a simple 3-line if-block with low risk. Chose Option B (exclude from unit test scope) over Option A (DI injection) to avoid adding unnecessary complexity for a trivial code path.
**Location:** Task 5 test cases in the plan file

#### 3. [FIX] Ambiguous flag detection description

**Change made:** Updated Task 3 GREEN step and the Dev Notes "ParseArgs Flag Handling Strategy" section to use the clearer wording: "when the resolved provider argument (the single argument before `--`, or `args[0]` when `--` is absent) starts with `-`". This covers both code paths unambiguously.
**Location:** Task 3 GREEN step and Dev Notes section

#### 4. [FIX] Minor line number drift in Dev Notes references

**Change made:** Updated line reference from "line 36" to "line 34" for the `ctx.AuthToken == ""` check.
**Location:** Dev Notes — URL Validation Strategy

#### 5. [FIX] `runner.go` `Run()` comment mischaracterizes the contract

**Change made:** Updated the Task 2 comment from "Returns (exitCode, nil) for normal exits" to the three-case version: "Returns (0, nil) on success, (exitCode, nil) for command exit errors, (1, error) for start failures."
**Location:** Task 2 comment in the plan file

### Summary

All 5 issues from Round 1 have been addressed. The plan is ready for re-review.

---

## Round 2

**Context:** Plan was revised based on Round 1 feedback. All 5 issues were marked as fixed. This round re-verifies each fix against the actual codebase and checks for new issues introduced by the fixes or missed in Round 1.

### Round 1 Issue Status

| # | Severity | Issue | Round 2 Status |
|---|----------|-------|----------------|
| 1 | BLOCKING | TDD violation: implementation before tests in Task 1 and Task 3 | **FIXED** — Both tasks now follow explicit Red-Green-Refactor subtask ordering |
| 2 | HIGH | TUI cancel test not feasible without DI/mocking | **FIXED** — TUI cancel removed from unit test scope; covered by Task 6 manual smoke test |
| 3 | MEDIUM | Ambiguous flag detection description | **FIXED** — Task 3 GREEN and Dev Notes now use "resolved provider argument" wording covering both code paths |
| 4 | LOW | Line number drift in Dev Notes | **FIXED** — Line reference updated to line 34 for `ctx.AuthToken == ""` check; verified against `internal/runner/runner.go` |
| 5 | LOW | `Run()` comment mischaracterizes non-zero exit path | **FIXED** — Comment now correctly lists all three return patterns |

### Issues Found

#### 1. [MEDIUM] Task 1 RED doesn't explain config setup needed for `New()` URL validation tests

- **Plan says (Task 1 RED):** "Write table-driven tests in `runner_test.go` first — test `validateURL` directly and test `New()` with invalid URLs. Cases: valid URL, missing scheme, contains spaces, empty string."
- **Actual:** Testing `New()` with invalid URLs requires a config context with an invalid `BaseURL` to exist in the configuration store. `New()` calls `config.GetContext()`, which calls `LoadConfig()`, which reads from the file returned by `GetConfigPath()` (respecting `CCCTX_CONFIG_PATH`). Without setting up a temp config file and pointing `CCCTX_CONFIG_PATH` at it, calling `New()` with any context name will fail with "context not found" before reaching the URL validation code path.
- **Impact:** The executor may write `New()` tests that fail for the wrong reason (context missing vs. URL invalid), or may spend time debugging why the URL validation test isn't exercising the intended code path.
- **Suggested fix:** Add a note in Task 1 RED clarifying that `New()` tests require temp config setup:
  > "For `New()` tests: create a temp TOML config file with contexts containing invalid `base_url` values, and use `t.Setenv(\"CCCTX_CONFIG_PATH\", tempFile)` before calling `New()`. Alternatively, focus RED tests on `validateURL` directly and defer `New()` integration tests to a separate test function."

#### 2. [MEDIUM] Task 3 RED test cases lack expected error values

- **Plan says (Task 3 RED):** "Add failing test cases in `args_test.go` for flag-like args in provider position: `["--model", "foo"]`, `["-m"]`, `["--model", "--", "foo"]`. Run tests, confirm failures."
- **Actual:** The plan lists the input `args` for the failing test cases but does not specify what `wantErr` value each case should expect. The existing `args_test.go` uses a `wantErr string` field in the test table, so the executor must supply an expected error string to make the test complete and self-documenting.
- **Impact:** Without a specified expected error, different executors may write tests expecting different messages (e.g., `"flag-like argument"` vs `"unexpected flag"`), leading to test-implementation mismatches during the GREEN phase.
- **Suggested fix:** Specify the expected error string in the RED step. For example:
  > "Expected error: `flag-like argument '--model' not allowed in provider position` (or similar — ensure the test expects a non-empty error containing the flag name)."

#### 3. [LOW] Inconsistent error message between Task 3 GREEN and Dev Notes

- **Plan says (Task 3 GREEN):** `return an error: "flag-like argument '%s' not allowed in provider position"`
- **Plan says (Dev Notes — ParseArgs Flag Handling Strategy):** "return an error like `unexpected flag '--model' in provider position`"
- **Actual:** These are two different error message formats. The GREEN step uses a printf-style `"flag-like argument '%s' not allowed in provider position"`, while the Dev Notes use `"unexpected flag '--model' in provider position"`. An executor reading the Dev Notes for context may write a different message than what the GREEN step specifies.
- **Impact:** Potential mismatch between tests, implementation, and documentation.
- **Suggested fix:** Align both sections to the same message. Recommended: `"flag-like argument '%s' not allowed in provider position"` (matches the GREEN step and is more precise).

#### 4. [LOW] Task 5 doesn't specify `cmd/run_test.go` package name

- **Plan says (Task 5):** "Test `runRun` directly — it returns exit code, no `os.Exit` in test path"
- **Actual:** `runRun` will be an unexported function (`runRun` lowercase) in package `cmd`. To call it from `cmd/run_test.go`, the test file must use `package cmd` (internal test), not `package cmd_test` (external test). The plan never specifies the package name.
- **Impact:** If the executor creates `cmd/run_test.go` with `package cmd_test`, the tests will fail to compile because `runRun` is unexported. This is an easy fix but costs a build-test-debug cycle.
- **Suggested fix:** Add a note in Task 5: "Use `package cmd` (not `package cmd_test`) so tests can call the unexported `runRun` function."

#### 5. [LOW] Task 2 claims "No code change needed" but requires adding a comment

- **Plan says (Task 2):** "No code change needed — this task is verify + document the existing contract"
- **Actual:** The very next subtask says: "Add a comment in `runner.go` near `Run()`". Adding a comment is a code change.
- **Impact:** Trivial confusion for the executor who may skip the task thinking no edits are required, then discover a comment needs to be added.
- **Suggested fix:** Change the wording to: "Minimal code change: verify the existing contract via audit, then add a documentation comment in `runner.go` near `Run()`."

### Verification Summary

| Item | Plan claims | Actual | Match |
|------|------------|--------|-------|
| `runner.go` line 34 — `ctx.AuthToken == ""` | AuthToken check | Line 34: `if ctx.AuthToken == ""` | **CORRECT** |
| `runner.go` line 37 — `len(opts.Target) == 0` | Target check | Line 37: `if len(opts.Target) == 0` | **CORRECT** |
| `runner.go` lines 44-60 | Run() returns (int, error) | Lines 44-60: `func (r *Runner) Run()` with three return paths | **CORRECT** |
| `args.go` lines 30-34 | ParseArgs "no args" path | Lines 30-34: `if len(args) == 0` through `return args[0], args[1:], false, nil` | **CORRECT** |
| `cmd/exec.go` lines 19-69 | exec command with os.Exit calls | Lines 19-69: Run closure with multiple `os.Exit` calls | **CORRECT** |
| `cmd/run.go` lines 20-74 | run command with os.Exit calls | Lines 20-74: Run closure with multiple `os.Exit` calls | **CORRECT** |
| `go test ./...` passes | AC6 verification step | Currently passes (3 packages with tests) | **CORRECT** |
| `go vet ./...` clean | AC6 verification step | Clean (no output) | **CORRECT** |
| `CCCTX_CONFIG_PATH` respected | Task 5 test strategy | `config/config.go` line 42: `os.Getenv("CCCTX_CONFIG_PATH")` | **CORRECT** |

### Risk Summary

| # | Severity | Issue | Task |
|---|----------|-------|------|
| 1 | MEDIUM | Task 1 RED lacks config setup guidance for `New()` tests | Task 1 |
| 2 | MEDIUM | Task 3 RED test cases missing expected error values | Task 3 |
| 3 | LOW | Inconsistent error message between Task 3 GREEN and Dev Notes | Task 3, Dev Notes |
| 4 | LOW | Task 5 doesn't specify test file package name | Task 5 |
| 5 | LOW | Task 2 "no code change" contradicts comment addition | Task 2 |

### Verdict

The plan is **ready for execution**. All blocking and high-severity issues from Round 1 are resolved. No new blocking or high-severity issues were found in Round 2.

The five remaining issues are all LOW or MEDIUM severity and represent gaps in guidance rather than structural problems:
- Issues #1 and #2 are MEDIUM: missing test setup details that a capable executor can reasonably infer from context, but would benefit from explicit guidance.
- Issues #3, #4, and #5 are LOW: minor inconsistencies or omissions that are easily corrected during execution.

The factual claims in the plan (line numbers, file paths, command behaviors) were all verified against the actual codebase and are correct.

---

## Round 3

**Context:** Plan author responding to Round 2 findings. All 5 issues are guidance-level gaps (MEDIUM/LOW). No structural problems found.

### Round 2 Issue Status

| # | Severity | Issue | Round 3 Status |
|---|----------|-------|----------------|
| 1 | MEDIUM | Task 1 RED lacks config setup guidance for `New()` tests | **FIXED** — Added temp config setup instructions with `t.Setenv("CCCTX_CONFIG_PATH", tempFile)` and alternative approach |
| 2 | MEDIUM | Task 3 RED test cases missing expected error values | **FIXED** — Added `wantErr` expectations containing the flag name for each test case |
| 3 | LOW | Inconsistent error message between Task 3 GREEN and Dev Notes | **FIXED** — Aligned Dev Notes to use `"flag-like argument '%s' not allowed in provider position"` matching the GREEN step |
| 4 | LOW | Task 5 doesn't specify test file package name | **FIXED** — Added note: use `package cmd` (not `package cmd_test`) to access unexported `runRun` |
| 5 | LOW | Task 2 "no code change" contradicts comment addition | **FIXED** — Changed wording to "Minimal code change: verify the existing contract via audit, then add a documentation comment" |

### Issues Found

No new issues found. All Round 2 issues were legitimate guidance gaps and have been addressed.

### Verification Summary

| Item | Plan claims | Actual | Match |
|------|------------|--------|-------|
| `runner.go` `New()` calls `config.GetContext()` first | Verified — URL validation must happen after context resolution | Line 27: `config.GetContext()` | **CORRECT** |
| `args_test.go` uses `wantErr string` field | Verified — RED test cases now specify expected error substrings | Line 18: `wantErr string` | **CORRECT** |
| Dev Notes error message format | Updated to match GREEN step | Both use `"flag-like argument '%s'"` | **ALIGNED** |
| `runRun` is unexported | Verified — plan now specifies `package cmd` | Task 5 note added | **CORRECT** |

### Risk Summary

No remaining issues. All Round 2 findings have been resolved.

### Verdict

The plan is **ready for execution**. All issues from Rounds 1 and 2 are resolved. No new issues were introduced by the fixes.

---

## Round 2 Response

**Responded by:** glm
**Date:** 2026-04-20
**Plan file:** 3-1-runner-and-command-layer-robustness.md (updated)

### Issue Responses

#### 1. [FIX] Task 1 RED doesn't explain config setup needed for `New()` URL validation tests

**Change made:** Added config setup guidance to Task 1 RED step: create a temp TOML config file with contexts containing invalid `base_url` values, use `t.Setenv("CCCTX_CONFIG_PATH", tempFile)` before calling `New()`. Also added alternative: focus RED tests on `validateURL` directly and defer `New()` integration tests to a separate test function.
**Location:** Task 1 RED subtask

#### 2. [FIX] Task 3 RED test cases lack expected error values

**Change made:** Added `wantErr` expectations to each test case: `["--model", "foo"]` expects error containing `"flag-like argument"`, `["-m"]` expects `"flag-like argument '-m'"`, `["--model", "--", "foo"]` expects `"flag-like argument '--model'"`. Added note that each test expects a non-empty error containing the flag name.
**Location:** Task 3 RED subtask

#### 3. [FIX] Inconsistent error message between Task 3 GREEN and Dev Notes

**Change made:** Updated Dev Notes "ParseArgs Flag Handling Strategy" section from `"unexpected flag '--model' in provider position"` to `"flag-like argument '%s' not allowed in provider position"` to match the GREEN step exactly.
**Location:** Dev Notes — ParseArgs Flag Handling Strategy

#### 4. [FIX] Task 5 doesn't specify `cmd/run_test.go` package name

**Change made:** Added note to Task 5: "Use `package cmd` (not `package cmd_test`) in `cmd/run_test.go` so tests can call the unexported `runRun` function."
**Location:** Task 5 notes

#### 5. [FIX] Task 2 claims "No code change needed" but requires adding a comment

**Change made:** Changed Task 2 subtask wording from "No code change needed — this task is verify + document the existing contract" to "Minimal code change: verify the existing contract via audit, then add a documentation comment in `runner.go` near `Run()`".
**Location:** Task 2 subtask

### Summary

All 5 issues from Round 2 have been addressed. The plan is ready for re-review.
