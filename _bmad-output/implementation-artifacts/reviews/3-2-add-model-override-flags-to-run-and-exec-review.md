---
reviewed_plan: _bmad-output/implementation-artifacts/3-2-add-model-override-flags-to-run-and-exec.md
reviewer: glm
latest_round: 3
latest_response: 3
review_rounds:
  - round: 1
    date: 2026-04-23
  - round: 2
    date: 2026-04-23
  - round: 3
    date: 2026-04-23
---

# Review: 3-2-add-model-override-flags-to-run-and-exec

Reviewed plan: `_bmad-output/implementation-artifacts/3-2-add-model-override-flags-to-run-and-exec.md`

## Round 1

**Context:** Initial full dry-run of the plan against the actual codebase. Verified all file paths, line numbers, code snippets, task ordering, TDD compliance, and interaction with existing tests.

### Issues Found

#### 1. [BLOCKING] TDD violation: all implementation tasks (1-4) before test tasks (5-7)

- **Plan says:** Task 1 writes `ExtractFlags` (implementation), Task 2 updates `buildEnv` (implementation), Tasks 3-4 update command files (implementation). Tests come after in Tasks 5-7.
- **Actual:** This is the classic "test-after" anti-pattern. Tests written after implementation pass immediately and never prove they catch real bugs. The plan's own predecessor (Story 3-1) was restructured in review to follow Red-Green-Refactor — this plan should follow the same discipline.
- **Impact:** Tests will be biased toward the implementation. Edge cases in `ExtractFlags` (e.g., flag after `--`, duplicate flags, flag value starting with `-`) will never be proven to fail before the implementation is written.
- **Suggested fix:** Restructure into Red-Green-Refactor cycles:
  - Task 1: RED — write failing tests for `ExtractFlags` in `args_test.go`
  - Task 2: GREEN — implement `ExtractFlags` in `args.go`
  - Task 3: RED — write failing tests for `buildEnv` priority chain in `runner_test.go`
  - Task 4: GREEN — implement priority chain in `runner.go`
  - Task 5: RED — write failing integration tests for flags in `cmd/run_test.go` (and update existing breaking test)
  - Task 6: GREEN — update `cmd/run.go` and `cmd/exec.go` to use `ExtractFlags`
  - Task 7: Verify all tests pass

#### 2. [BLOCKING] Existing test will break — `"ParseArgs error - flag-like arg"` in `cmd/run_test.go`

- **Plan says:** AC9 claims "All existing tests pass". Task 8 says "`go test ./...` passes".
- **Actual:** `cmd/run_test.go` line 22-28 has a test case:
  ```go
  {name: "ParseArgs error - flag-like arg", args: []string{"--model", "foo"}, wantCode: 1}
  ```
  This test calls `runRun(["--model", "foo"])` and expects exit code 1 (ParseArgs rejects flag-like args). After the plan is implemented:
  1. `ExtractFlags(["--model", "foo"])` → model="foo", remaining=[]
  2. `ParseArgs([])` → useTUI=true
  3. TUI flow → `ui.RunContextSelector(["test"])` → blocks on terminal I/O → **test hangs indefinitely**

  The plan has no task to update this test. AC9 ("all existing tests pass") is unachievable without modifying it.
- **Impact:** `go test ./...` hangs on this test case. Executor will discover this during Task 8 verification and must fix it ad-hoc.
- **Suggested fix:** Add a task (or subtask) to update the existing test case. The test should now expect success (exit code 0 with a claude mock) or be updated to test a truly unknown flag like `--unknown-flag foo` that `ExtractFlags` would NOT extract and `ParseArgs` would still reject. Two options:
  - **Option A:** Change the test args to `[]string{"--unknown-flag", "foo"}` so ExtractFlags passes it through and ParseArgs still rejects it.
  - **Option B:** Change the test to expect TUI mode (but this requires mocking the selector — same issue as Story 3-1 review issue #2).

#### 3. [HIGH] `--help` regression from `DisableFlagParsing: true`

- **Plan says:** Task 3 adds `DisableFlagParsing: true` to `RunCmd`. Task 4 adds it to `ExecCmd`.
- **Actual:** With `DisableFlagParsing: true`, Cobra no longer processes `--help`. Currently `ccctx run --help` shows Cobra's help text. After the change:
  1. `ExtractFlags(["--help"])` → no match → remaining=["--help"]
  2. `ParseArgs(["--help"])` → error "flag-like argument '--help' not allowed in provider position"

  This is a user-visible regression. `ccctx exec --help` would break similarly.
- **Impact:** Users who run `ccctx run --help` or `ccctx exec --help` will see an error instead of help text. Note: `ccctx --help run` and `ccctx help run` still work (parent command handles them).
- **Suggested fix:** Add manual `--help` handling in `ExtractFlags` or in `runRun`/`execRun`. For example, in `ExtractFlags`, check if `--help` or `-h` appears before `--` and return a sentinel value or special error. Or handle it in `runRun`/`execRun` before calling `ExtractFlags`:
  ```go
  for _, arg := range args {
      if arg == "--" { break }
      if arg == "--help" || arg == "-h" {
          fmt.Print(helpText)
          return 0
      }
  }
  ```
  Alternatively, add `--help`/`-h` as recognized flags in `ExtractFlags` and return a boolean `showHelp`.

#### 4. [HIGH] No integration tests for `exec` command flags

- **Plan says:** AC6 requires "Flags available on both commands". Task 7 adds integration tests only in `cmd/run_test.go`. No `cmd/exec_test.go` exists.
- **Actual:** There are zero test cases for the `exec` command with flags. The plan claims the pattern is "identical to run.go" (Task 4), but this claim is unverified by any test. The exec command has a different default behavior (launches `$SHELL` when no target args), which introduces different edge cases:
  - `ccctx exec --model foo` → TUI mode with model override, then shell
  - `ccctx exec provider-A --model foo` → model override with shell fallback (no target args)
  - `ccctx exec provider-A --model foo -- env | grep MODEL` → model override with target command
- **Impact:** The exec command flag behavior is completely untested. Bugs in the exec path (e.g., `--model` value not being passed to `runner.Options`) would go undetected.
- **Suggested fix:** Either:
  - **Option A:** Add a parallel Task 8 for exec integration tests in `cmd/exec_test.go` (new file, using `package cmd` to access `execRun`).
  - **Option B:** Add exec test cases to the existing `cmd/run_test.go` alongside run tests (since both are in `package cmd`).

#### 5. [MEDIUM] Line reference `cmd/run.go:15-22` slightly off

- **Plan says:** "Source: cmd/run.go:15-22 — RunCmd definition"
- **Actual:** `RunCmd` spans lines 15-23 (closing brace is on line 23, not 22). The content is correct — `DisableFlagParsing: true` goes inside the struct literal — but the line range is off by one.
- **Impact:** Minimal. The insertion point (`DisableFlagParsing: true` before the `Run` field) is clear from context.
- **Suggested fix:** Update reference to `cmd/run.go:15-23`.

#### 6. [MEDIUM] Code snippets use spaces, actual code uses tabs

- **Plan says:** Dev Notes code snippets use 4-space indentation throughout.
- **Actual:** All Go source files use tab indentation (standard `gofmt`). The `ExtractFlags` algorithm snippet, `runRun` changes, and `buildEnv` changes all use spaces.
- **Impact:** Executor who copies code directly from the plan will introduce mixed indentation. `go fmt` fixes it, but it adds a step and could confuse.
- **Suggested fix:** Note in Dev Notes: "All code uses tab indentation — run `go fmt` after edits." Or use `[...]` placeholders to indicate where to add code rather than full snippets.

#### 7. [MEDIUM] ExtractFlags code has a subtle return value issue on error

- **Plan says:** `ExtractFlags` returns `("", "", nil, fmt.Errorf(...))` on error (remaining = nil).
- **Actual:** While functionally fine (callers check error first), this is inconsistent with Go convention where error returns typically use zero values, not `nil` for slices. More importantly, if a future caller forgets to check the error and uses `remaining`, they'll get a nil slice which may cause panics in `ParseArgs` (e.g., `len(nil)` works, but `args[0]` on nil panics).
- **Impact:** Low risk currently — both callers (`runRun`, `execRun`) check error first. But it's a defense-in-depth concern.
- **Suggested fix:** Change error returns to `("", "", []string{}, err)` for consistency, or add a comment noting that remaining is invalid on error.

### Verification Summary

| Item | Plan claims | Actual | Match |
|------|------------|--------|-------|
| `internal/runner/args.go` exists | File path | Exists, 45 lines | **CORRECT** |
| `internal/runner/runner.go:82-99` buildEnv | Line range | Lines 82-99: `func buildEnv` | **CORRECT** |
| `internal/runner/runner.go:14-19` Options struct | Line range | Lines 14-19: `type Options` | **CORRECT** |
| `cmd/run.go:15-22` RunCmd | Line range | Lines 15-23 (off by 1) | **OFF BY 1** |
| `cmd/exec.go:14-22` ExecCmd | Line range | Lines 14-22 | **CORRECT** |
| `cmd/run.go:25-78` runRun | Line range | Lines 25-78 | **CORRECT** |
| `cmd/exec.go:24-74` execRun | Line range | Lines 24-74 | **CORRECT** |
| `args.go:28-29,40-41` flag rejection | Line numbers | Lines 28-29, 40-41: flag rejection in ParseArgs | **CORRECT** |
| `Options.Model` and `Options.SmallFastModel` fields exist | Previous Story Intelligence | Lines 17-18: both fields present | **CORRECT** |
| `buildEnv` receives `opts Options` | Dev Notes | Line 82: `buildEnv(ctx *config.Context, opts Options)` | **CORRECT** |
| `cmd/run_test.go` uses `package cmd` | Previous Story Intelligence | Line 1: `package cmd` | **CORRECT** |
| `cmd/exec_test.go` does not exist | Implicit (not mentioned) | File does not exist | **CORRECT** |
| `DisableFlagParsing` not currently set | Plan's Task 3 | `grep -c` returns 0 for both files | **CORRECT** |

### Risk Summary

| # | Severity | Issue | Task |
|---|----------|-------|------|
| 1 | **BLOCKING** | TDD violation: implementation before tests | Tasks 1-7 |
| 2 | **BLOCKING** | Existing test breaks (`--model foo` enters TUI) | Not addressed |
| 3 | HIGH | `--help` regression from `DisableFlagParsing: true` | Tasks 3-4 |
| 4 | HIGH | No integration tests for exec command flags | Task 7 |
| 5 | MEDIUM | Line reference off by 1 for RunCmd | Dev Notes |
| 6 | MEDIUM | Code snippets use spaces, code uses tabs | Dev Notes |
| 7 | MEDIUM | Error return uses `nil` for remaining slice | Task 1 code |

### Verdict

The plan is **NOT ready for execution**. Fix the two blocking issues first:

1. **Restructure task ordering** to follow Red-Green-Refactor (issue #1). Write failing tests for each component before implementation, following the same pattern established in Story 3-1's review.
2. **Acknowledge and handle the breaking test** (issue #2). The existing `"ParseArgs error - flag-like arg"` test in `cmd/run_test.go` will hang after `ExtractFlags` is added. Either update the test to use a truly unknown flag, or restructure it.

Also address the two high-severity issues:
3. **Handle `--help` regression** (issue #3) by adding manual help handling when `DisableFlagParsing: true` is set.
4. **Add exec command integration tests** (issue #4) to verify AC6.

## Round 1 Response

**Responded by:** claude
**Date:** 2026-04-23
**Plan file:** 3-2-add-model-override-flags-to-run-and-exec.md (updated)

### Issue Responses

#### 1. [FIX] TDD violation: all implementation tasks (1-4) before test tasks (5-7)

**Change made:** Restructured all 8 tasks into Red-Green-Refactor cycles. Each component (ExtractFlags, buildEnv priority, command integration) now has a RED task (write failing tests) immediately followed by a GREEN task (implement to pass). Task ordering: RED ExtractFlags tests → GREEN ExtractFlags → RED buildEnv tests → GREEN buildEnv → RED run integration tests + fix breaking test → RED exec integration tests → GREEN command updates → verify all.
**Location:** Tasks section (Task 1 through Task 8, fully rewritten)

#### 2. [FIX] Existing test will break — "ParseArgs error - flag-like arg" in cmd/run_test.go

**Change made:** Added explicit subtask in Task 5 to update the existing test case. Changed args from `[]string{"--model", "foo"}` to `[]string{"--unknown-flag", "foo"}` so that ExtractFlags does NOT extract it and ParseArgs still rejects it. This preserves the test's intent (unknown flags rejected) while accommodating the new flag extraction layer.
**Location:** Task 5, first subtask

#### 3. [FIX] --help regression from DisableFlagParsing: true

**Change made:** Added a `WantsHelp(args []string) bool` helper function in `args.go` that checks for `--help`/`-h` before the `--` separator. Both Run closures now check `runner.WantsHelp(args)` before calling `runRun`/`execRun`, calling `cmd.Help()` if true. Added new Dev Notes section "Help Handling (prevents --help regression)" with the function code. Updated Command File Changes to show the WantsHelp check in the Run closure. Added edge cases #6 (help before separator) and #7 (unknown flag behavior). Added --help smoke tests to Task 8.
**Location:** New "Help Handling" Dev Notes section, updated Command File Changes, updated Edge Cases, Task 7 and Task 8

#### 4. [FIX] No integration tests for exec command flags

**Change made:** Added Task 6 as a dedicated RED task for exec integration tests. Creates new file `cmd/exec_test.go` (package cmd) with table-driven tests covering: `--model foo provider-A` with shell mock, `--model` without value, and `provider-A --model foo -- env | grep MODEL`. Added `cmd/exec_test.go` to Key Files table as CREATE.
**Location:** New Task 6, updated Key Files table

#### 5. [FIX] Line reference cmd/run.go:15-22 slightly off

**Change made:** Updated reference from `cmd/run.go:15-22` to `cmd/run.go:15-23` in References section.
**Location:** References section

#### 6. [FIX] Code snippets use spaces, actual code uses tabs

**Change made:** Added note to Project Structure Notes: "All Go source files use tab indentation — run `go fmt` after edits. Code snippets in this plan use spaces for readability only."
**Location:** Project Structure Notes section

#### 7. [FIX] ExtractFlags code has a subtle return value issue on error

**Change made:** Changed error returns in ExtractFlags code snippet from `nil` to `[]string{}` for the remaining slice. Also added explicit instruction in Task 2: "Error returns use `[]string{}` (not `nil`) for remaining slice".
**Location:** Dev Notes ExtractFlags code snippet, Task 2 subtask

### Summary

All 7 issues from Round 1 have been addressed. The plan is ready for re-review.

---

## Round 2

**Context:** Full re-verification of the updated plan after all Round 1 fixes. Traced every task step-by-step against the actual codebase, mentally executed all code snippets (ExtractFlags algorithm, WantsHelp, buildEnv priority chain, command file changes), and verified all file paths and line references.

### Round 1 Issue Status

| # | Severity | Issue | Round 2 Status |
|---|----------|-------|----------------|
| 1 | BLOCKING | TDD violation: implementation before tests | **FIXED** — Tasks restructured into Red-Green-Refactor cycles (Task 1 RED → 2 GREEN, Task 3 RED → 4 GREEN, Tasks 5-6 RED → 7 GREEN) |
| 2 | BLOCKING | Existing test breaks (`--model foo` enters TUI) | **FIXED** — Task 5 has explicit subtask to update test args from `--model` to `--unknown-flag` |
| 3 | HIGH | `--help` regression from `DisableFlagParsing: true` | **FIXED** — WantsHelp helper added with Dev Notes section, checked in Run closures, edge cases documented, smoke tests added to Task 8 |
| 4 | HIGH | No integration tests for exec command flags | **FIXED** — Task 6 creates `cmd/exec_test.go` with table-driven tests |
| 5 | MEDIUM | Line reference off by 1 for RunCmd | **FIXED** — Updated to `cmd/run.go:15-23` (verified correct) |
| 6 | MEDIUM | Code snippets use spaces, code uses tabs | **FIXED** — Note added to Project Structure Notes |
| 7 | MEDIUM | Error return uses `nil` for remaining slice | **FIXED** — Code snippet and Task 2 both use `[]string{}` |

### Issues Found

#### 1. [MEDIUM] WantsHelp has no unit tests

- **Plan says:** `WantsHelp(args []string) bool` is added to `args.go` (Dev Notes, "Help Handling" section). Verified by manual smoke tests in Task 8 (`ccctx run --help`, `ccctx exec --help`).
- **Actual:** WantsHelp is a pure function that could easily have table-driven unit tests in `args_test.go`. It's called in the Cobra `Run` closure — not inside `runRun`/`execRun` — so the cmd-level integration tests (which call `runRun`/`execRun` directly) cannot exercise it. The only automated coverage would come from integration tests that invoke the Run closure, but no such tests are specified.
- **Impact:** If WantsHelp has a bug (e.g., missing `-h` check, or not stopping at `--`), it would only be caught by manual testing. This is a regression-risky function since it replaces Cobra's built-in help handling.
- **Suggested fix:** Add WantsHelp test cases to Task 1 alongside ExtractFlags tests. Cases: `["--help"]` → true, `["-h"]` → true, `["provider-A", "--", "--help"]` → false, `[]` → false, `["provider-A", "--help"]` → true.

#### 2. [MEDIUM] Integration test env var verification mechanism unspecified

- **Plan says:** Task 5: "`--model foo provider-A` with claude mock → success, verify ANTHROPIC_MODEL env var". Task 6: "`--model foo provider-A` with shell mock → success, verify ANTHROPIC_MODEL env var".
- **Actual:** The existing mocks in `cmd/run_test.go` (lines 58-63) are simple shell scripts that just `exit 0` or `exit 42`. They don't capture environment variables. To verify `ANTHROPIC_MODEL=foo` in the child process, the executor needs to create a mock that writes its environment to a temp file and a mechanism to read that file back. The plan doesn't specify this mechanism.
- **Impact:** The executor must improvise the mock infrastructure. Two reasonable approaches exist: (a) mock writes `printenv ANTHROPIC_MODEL` to a file via an env-var-specified path, or (b) mock writes full `env` output to a file. Both work but the choice affects all test cases.
- **Suggested fix:** Add a Dev Notes subsection or Task 5 subtask specifying the env-capturing mock pattern. For example: "Create mock script that writes `printenv ANTHROPIC_MODEL` to `$MOCK_OUTPUT_FILE`, set `MOCK_OUTPUT_FILE` to a temp file in the test, read and assert after runRun returns."

#### 3. [LOW] Missing test cases for documented edge cases

- **Plan says:** Edge Cases section documents case #3: "`--model -` → model value is `-`. This is technically valid (though unusual). ExtractFlags should accept it." And case #4: "Duplicate flags: `--model foo --model bar` → last value wins."
- **Actual:** Neither edge case has a corresponding test case in Task 1's ExtractFlags tests.
- **Impact:** These edge cases are handled correctly by the ExtractFlags algorithm (verified by mental trace), but are untested. Low risk since the algorithm is simple and the cases are unusual.
- **Suggested fix:** Add two test cases to Task 1: `provider-A --model -` → model="-", remaining=["provider-A"]; `provider-A --model foo --model bar` → model="bar", remaining=["provider-A"].

#### 4. [LOW] Task 6 exec tests don't cover exec TUI mode with flags

- **Plan says:** AC7: "`ccctx run --model claude-opus-4-7` (no provider) opens TUI selector, and the selected provider runs with the overridden model."
- **Actual:** AC7 is only tested for the `run` command (in Task 5). Task 6's exec tests don't include a TUI-mode-with-flags test case (`ccctx exec --model foo` with no provider). The exec command has the same TUI flow but a different post-TUI behavior (launches `$SHELL` instead of looking for `claude`).
- **Impact:** The exec TUI + model override path is untested. If there's a bug in how flags flow through exec's TUI path, it won't be caught.
- **Suggested fix:** Add a test case to Task 6: `--model foo` (no provider) with shell mock → TUI mode with model override. This requires the same TUI-mocking approach as the run tests.

### Verification Summary

| Item | Plan claims | Actual | Match |
|------|------------|--------|-------|
| `internal/runner/args.go` exists | File path | Exists, 45 lines | **CORRECT** |
| `internal/runner/args_test.go` exists | File path | Exists, 118 lines | **CORRECT** |
| `internal/runner/runner.go:82-99` buildEnv | Line range | Lines 82-99: `func buildEnv` | **CORRECT** |
| `internal/runner/runner.go:14-19` Options struct | Line range | Lines 14-19: `type Options` with Model, SmallFastModel | **CORRECT** |
| `cmd/run.go:15-23` RunCmd | Line range (updated) | Lines 15-23: RunCmd struct literal | **CORRECT** |
| `cmd/exec.go:14-22` ExecCmd | Line range | Lines 14-22: ExecCmd struct literal | **CORRECT** |
| `cmd/run.go:25-78` runRun | Line range | Lines 25-78: `func runRun` | **CORRECT** |
| `cmd/exec.go:24-74` execRun | Line range | Lines 24-74: `func execRun` | **CORRECT** |
| `args.go:28-29,40-41` flag rejection | Line numbers | Lines 28-29, 40-41: flag rejection in ParseArgs | **CORRECT** |
| `cmd/run_test.go` uses `package cmd` | Previous Story Intelligence | Line 1: `package cmd` | **CORRECT** |
| `cmd/exec_test.go` does not exist | Implicit | File does not exist | **CORRECT** |
| `Options.Model` and `Options.SmallFastModel` fields | Lines 17-18 | Both fields present | **CORRECT** |
| `buildEnv` receives `opts Options` | Line 82 | `buildEnv(ctx *config.Context, opts Options)` | **CORRECT** |
| ExtractFlags algorithm correctness | Dev Notes code | Mental trace of all 9 test cases: all pass | **CORRECT** |
| WantsHelp algorithm correctness | Dev Notes code | Mental trace: --help/-h detected, stops at -- | **CORRECT** |
| buildEnv priority chain correctness | Dev Notes code | opts > ctx > omit: correct for both Model and SmallFastModel | **CORRECT** |
| Existing tests pass with buildEnv changes | Task 4 code | TestBuildEnv, TestBuildEnv_SkipsEmptyOptional, TestBuildEnv_InjectionOrder: all unchanged behavior | **CORRECT** |
| Breaking test update (--model → --unknown-flag) | Task 5 | `cmd/run_test.go:22`: `[]string{"--model", "foo"}` → plan updates to `[]string{"--unknown-flag", "foo"}` | **CORRECT** |

### Risk Summary

| # | Severity | Issue | Task |
|---|----------|-------|------|
| 1 | MEDIUM | WantsHelp has no unit tests | Task 1, Task 7 |
| 2 | MEDIUM | Integration test env var verification mechanism unspecified | Task 5, Task 6 |
| 3 | LOW | Missing test cases for documented edge cases (flag value `-`, duplicate flags) | Task 1 |
| 4 | LOW | Task 6 exec tests don't cover exec TUI mode with flags | Task 6 |

### Verdict

The plan is **in good shape for execution**. All Round 1 blocking and high-severity issues are resolved. The TDD restructuring is correct and complete, the --help regression is properly handled, and the exec test coverage gap is addressed.

The two medium-severity issues should be considered before starting:
1. **WantsHelp unit tests** (issue #1) — Adding a few table-driven test cases to Task 1 is low effort and provides important safety coverage for a function that replaces Cobra's built-in help handling.
2. **Env var mock mechanism** (issue #2) — Specifying the mock pattern upfront saves the executor from improvising test infrastructure mid-task.

Neither issue blocks execution — the executor can resolve them during implementation — but addressing them would improve the plan's completeness.

## Round 2 Response

**Responded by:** claude
**Date:** 2026-04-23
**Plan file:** 3-2-add-model-override-flags-to-run-and-exec.md (updated)

### Issue Responses

#### 1. [FIX] WantsHelp has no unit tests

**Change made:** Added WantsHelp table-driven test cases to Task 1 alongside ExtractFlags tests. Cases cover: `["--help"]` → true, `["-h"]` → true, `["provider-A", "--", "--help"]` → false (after separator), `[]` → false, `["provider-A", "--help"]` → true, `["provider-A"]` → false. Updated Task 1 title to mention WantsHelp.
**Location:** Task 1 (test cases and title)

#### 2. [FIX] Integration test env var verification mechanism unspecified

**Change made:** Added new Dev Notes section "Env-Capturing Mock Pattern" specifying the mock infrastructure: a shell script that writes `printenv ANTHROPIC_MODEL` to `$MOCK_OUTPUT_FILE`, with test setup instructions (create temp file, set env var, read back and assert). Updated Task 5 and Task 6 to reference this pattern. Noted that for exec tests, the mock should be named as the `$SHELL` basename (e.g., `bash`) since exec falls back to `$SHELL` when no target args.
**Location:** New "Env-Capturing Mock Pattern" Dev Notes section; Task 5 and Task 6 test descriptions

#### 3. [FIX] Missing test cases for documented edge cases

**Change made:** Added two test cases to Task 1's ExtractFlags table: `provider-A --model -` → model="-", remaining=["provider-A"] (edge case #3: flag value is `-`); `provider-A --model foo --model bar` → model="bar", remaining=["provider-A"] (edge case #4: duplicate flags, last wins).
**Location:** Task 1, ExtractFlags test cases (appended)

#### 4. [WONTFIX] Task 6 exec tests don't cover exec TUI mode with flags

**Reason:** TUI mode integration testing requires mocking `ui.RunContextSelector`, which is complex mock infrastructure that was explicitly deferred in Story 3-1 (which also didn't add TUI-mode integration tests). The flag handling code path is identical between TUI and non-TUI modes — both call `ExtractFlags` at the top of `execRun`/`runRun`, both pass the same `Options` struct. If the non-TUI path works correctly with flags (tested by Tasks 5-6), the TUI path will too since the flag extraction happens before the TUI branch. Adding TUI-mode tests is a worthwhile follow-up but out of scope for this story.

### Summary

All issues from Round 2 have been addressed. 3 fixed, 1 wontfix. The plan is ready for re-review.

---

## Round 3

**Context:** Full re-verification after Round 2 fixes. Mentally traced every task step-by-step against the actual codebase, executed all code snippets (ExtractFlags algorithm for all 11 test cases, WantsHelp for all 6 test cases, buildEnv priority chain, command file changes for both run.go and exec.go), verified all file paths and line references, and checked for new issues introduced by Round 2 fixes.

### Round 2 Issue Status

| # | Severity | Issue | Round 3 Status |
|---|----------|-------|----------------|
| 1 | MEDIUM | WantsHelp has no unit tests | **FIXED** — Table-driven WantsHelp tests added to Task 1 with 6 cases |
| 2 | MEDIUM | Integration test env var verification mechanism unspecified | **FIXED** — New "Env-Capturing Mock Pattern" Dev Notes section with mock script, test setup, and PATH instructions |
| 3 | LOW | Missing test cases for documented edge cases | **FIXED** — Added `--model -` and duplicate flags test cases to Task 1 |
| 4 | LOW | Task 6 exec tests don't cover exec TUI mode with flags | **WONTFIX** — Justified: TUI mocking is complex, flag extraction happens before TUI branch, non-TUI path covers the logic |

### Issues Found

#### 1. [MEDIUM] Task 6 third exec test case has mock naming ambiguity

- **Plan says:** Task 6 third test case: `provider-A --model foo -- env | grep MODEL` → success with model override. Dev Notes "Env-Capturing Mock Pattern" says: "Set `PATH` to temp dir containing the mock (named `claude` for run tests, `$SHELL` basename for exec tests)".
- **Actual:** For this test case, after ExtractFlags and ParseArgs, the target args are `["env", "|", "grep", "MODEL"]`. Since `len(targetArgs) > 0`, execRun does NOT fall back to `$SHELL` — it calls `exec.LookPath("env")`. The mock needs to be named `env`, not `bash` (the `$SHELL` basename). The Dev Notes pattern only applies to exec tests where no target args exist (shell fallback). This test has explicit target args, requiring a different mock name.
- **Impact:** Executor following the Dev Notes mock pattern literally will create a mock named `bash`, but LookPath("env") won't find it in the temp PATH. The test will fail with "env not found in PATH" rather than the intended verification. The executor must improvise: either (a) create a mock named `env` instead, or (b) create both `bash` and `env` mocks. Option (a) is correct for this test case.
- **Suggested fix:** Add a clarifying note to Task 6's third test case or to the Env-Capturing Mock Pattern section: "When exec tests have explicit target args after `--`, the mock should be named after the first target arg (e.g., `env`), not the `$SHELL` basename. The `$SHELL` basename pattern only applies to exec tests with no target args (shell fallback)." Alternatively, simplify the third test case to avoid the ambiguity: use `provider-A --model foo` (no target args, shell fallback) which is already covered by the first test case, or use `provider-A --model foo -- echo hello` where the mock is named `echo`.

#### 2. [LOW] Task 5 `--model` without value test case passes during RED phase

- **Plan says:** Task 5 includes test case `--model` without value → exit code 1, and "Verify new tests fail (runRun doesn't call ExtractFlags yet)".
- **Actual:** Without ExtractFlags, runRun passes `["--model"]` to ParseArgs, which rejects it as a flag-like arg (error: "flag-like argument '--model'"), returning exit code 1. The test expects exit code 1. So this test **passes immediately** — it is not a failing test during the RED phase. After implementation, ExtractFlags catches it first with a different error ("--model requires a value"), but the exit code is still 1. The test passes both before and after implementation.
- **Impact:** This test provides no RED/GREEN signal. It tests that "no value after --model → error", which is correct behavior, but the verification happens at the wrong layer (ParseArgs instead of ExtractFlags). The other three tests in Task 5 DO fail during RED, so the overall RED phase is still valid.
- **Suggested fix:** Acceptable as-is. The test validates important behavior (exit code 1 for `--model` without value) even though it passes during RED. If strict RED compliance is desired, change the test to verify the specific error message contains "requires a value" (which would fail during RED since ParseArgs says "flag-like argument" instead). But exit code verification is sufficient for integration tests.

#### 3. [TRIVIAL] No explicit REFACTOR pass

- **Plan says:** Tasks follow Red-Green pattern: RED (failing tests) → GREEN (implement).
- **Actual:** No task includes a REFACTOR step. The Anti-Patterns section correctly prevents duplication ("use shared ExtractFlags in runner package"), and the implementation is clean enough that no refactoring is obviously needed.
- **Impact:** None. The code is minimal and well-structured. ExtractFlags/WantsHelp are small pure functions, buildEnv change is a straightforward priority chain, command changes are mechanical.
- **Suggested fix:** No action needed.

### Verification Summary

| Item | Plan claims | Actual | Match |
|------|------------|--------|-------|
| `internal/runner/args.go` exists | File path | Exists, 44 lines | **CORRECT** |
| `internal/runner/args_test.go` exists | File path | Exists, 118 lines | **CORRECT** |
| `internal/runner/runner.go:82-99` buildEnv | Line range | Lines 82-99: `func buildEnv` through closing `}` | **CORRECT** |
| `internal/runner/runner.go:14-19` Options struct | Line range | Lines 14-19: `type Options` with Model, SmallFastModel | **CORRECT** |
| `cmd/run.go:15-23` RunCmd | Line range | Lines 15-23: RunCmd struct literal | **CORRECT** |
| `cmd/exec.go:14-22` ExecCmd | Line range | Lines 14-22: ExecCmd struct literal | **CORRECT** |
| `cmd/run.go:25-78` runRun | Line range | Lines 25-78: `func runRun` | **CORRECT** |
| `cmd/exec.go:24-74` execRun | Line range | Lines 24-74: `func execRun` | **CORRECT** |
| ExtractFlags algorithm (11 test cases) | Dev Notes code | Mental trace: all 11 cases produce expected outputs | **CORRECT** |
| WantsHelp algorithm (6 test cases) | Dev Notes code | Mental trace: all 6 cases produce expected outputs | **CORRECT** |
| buildEnv priority chain | Dev Notes code | opts > ctx > omit: correct for both Model and SmallFastModel | **CORRECT** |
| Existing tests unaffected by buildEnv change | Task 4 code | TestBuildEnv, TestBuildEnv_SkipsEmptyOptional, TestBuildEnv_InjectionOrder all use empty Options | **CORRECT** |
| DisableFlagParsing not currently set | Task 7 | grep returns 0 matches in both cmd files | **CORRECT** |
| ExtractFlags/WantsHelp don't exist yet | Task 1 RED | grep returns 0 matches in source files | **CORRECT** |
| `cmd/exec_test.go` does not exist | Task 6 CREATE | find returns no such file | **CORRECT** |
| Breaking test update `--model` → `--unknown-flag` | Task 5 | `cmd/run_test.go:22`: currently `[]string{"--model", "foo"}` | **CORRECT** |
| Env-Capturing Mock Pattern section exists | Round 2 fix | Dev Notes section present with mock script and setup instructions | **CORRECT** |

### Risk Summary

| # | Severity | Issue | Task |
|---|----------|-------|------|
| 1 | MEDIUM | Task 6 third exec test: mock naming ambiguity for explicit target args | Task 6 |
| 2 | LOW | Task 5 `--model` without value test passes during RED phase | Task 5 |
| 3 | TRIVIAL | No explicit REFACTOR pass | Tasks 1-7 |

### Verdict

The plan is **ready for execution**. No blocking or high-severity issues remain. All Round 1 and Round 2 issues are resolved.

One medium-severity item to consider:
- **Task 6 mock naming** (issue #1) — The executor will likely discover this during test writing and adapt (create a mock named `env` instead of `bash`). Adding a clarifying note to the Dev Notes would eliminate the ambiguity, but it's not a blocker since the fix is obvious in context.

The plan is well-structured with clean Red-Green cycles, comprehensive test coverage, and correct algorithms. All code references match the actual codebase.

## Round 3 Response

**Responded by:** claude
**Date:** 2026-04-23
**Plan file:** 3-2-add-model-override-flags-to-run-and-exec.md (updated)

### Issue Responses

#### 1. [FIX] Task 6 third exec test case has mock naming ambiguity

**Change made:** Updated the "Env-Capturing Mock Pattern" Dev Notes section to replace the single-line mock naming rule with an explicit list of three cases: (1) run tests → mock named `claude`, (2) exec tests with no target args → mock named as `$SHELL` basename, (3) exec tests with explicit target args after `--` → mock named after the first target arg. Also added inline clarification to Task 6's third test case: "Mock named `env` (first target arg after `--`), not `$SHELL` basename."
**Location:** Dev Notes "Env-Capturing Mock Pattern" section (step 4 expanded); Task 6 third test case

#### 2. [WONTFIX] Task 5 `--model` without value test case passes during RED phase

**Reason:** The reviewer's own analysis confirms this is acceptable. The test validates important behavior (`--model` without a value → exit code 1) that must work both before and after implementation. While it doesn't provide a RED signal, the other three tests in Task 5 DO fail during RED, so the overall RED phase is valid. Forcing this test to fail during RED (e.g., by asserting a specific error message like "requires a value") would couple the test to ExtractFlags' error wording, which is fragile for an integration test. Exit code verification is the correct abstraction level for cmd-level integration tests.

#### 3. [WONTFIX] No explicit REFACTOR pass

**Reason:** The reviewer confirmed "No action needed." The implementation is minimal and well-structured — ExtractFlags and WantsHelp are small pure functions, buildEnv change is a straightforward priority chain, and command changes are mechanical. No refactoring is anticipated.

### Summary

All issues from Round 3 have been addressed. 1 fixed, 2 wontfix. The plan is ready for execution.
