---
reviewed_plan: _bmad-output/implementation-artifacts/2-1-exec-command-integrates-tui-interactive-selector.md
reviewer: glm
latest_round: 3
review_rounds:
  - round: 1
    date: 2026-04-19
  - round: 2
    date: 2026-04-19
  - round: 3
    date: 2026-04-19
---

# Review: 2-1-exec-command-integrates-tui-interactive-selector

Reviewed plan: `_bmad-output/implementation-artifacts/2-1-exec-command-integrates-tui-interactive-selector.md`

## Round 1

**Context:** Initial full dry-run of Story 2.1 plan against the actual codebase. Verified all file paths, line numbers, code claims, task ordering, and TDD compliance.

### Issues Found

#### 1. [HIGH] Task 1 contradicts Task 5: empty-contexts check removal vs. test expectation

- **Plan says (Task 1):** "Remove internal `config.ListContexts()` call **and `len(contexts) == 0` check** from selector (caller handles this)"
- **Plan says (Task 5):** "Test: `RunContextSelector([]string{})` returns error containing 'no contexts found'"
- **Actual:** These two tasks are mutually exclusive. If Task 1 removes the empty check from `RunContextSelector`, then `RunContextSelector([]string{})` will pass the empty slice to `runTviewSelector`, which will attempt to launch a tview application with 0 items — likely hanging or crashing in a test environment. It will NOT return "no contexts found".
- **Impact:** The test in Task 5 will fail (or hang indefinitely) because the behavior it tests was removed in Task 1. The executor will be stuck trying to reconcile these conflicting instructions.
- **Suggested fix:** Choose one approach:
  - **Option A (recommended):** Keep the `len(contexts) == 0` validation in `RunContextSelector` as a defensive input check. Task 1 should only remove the `config.ListContexts()` call, not the empty check. Reword to: "Remove internal `config.ListContexts()` call. Keep the `len(contexts) == 0` validation as input parameter checking."
  - **Option B:** Remove both as Task 1 says, and change the Task 5 test to verify the callers (`cmd/run.go`, `cmd/exec.go`) handle empty contexts correctly. But this moves the test scope outside `internal/ui/selector_test.go`.

#### 2. [HIGH] TDD violation: all implementation before all tests

- **Plan says:** Tasks 1–4 are implementation changes. Task 5 is tests. Task 6 is verification.
- **Actual:** The plan follows the "implement everything first, then write tests" pattern. Tasks 1–4 modify production code (`selector.go`, `cmd/exec.go`, `cmd/run.go`), and Task 5 adds tests after all changes are complete. This is the classic test-after anti-pattern — tests written after implementation are biased toward what was built, not what was required, and they pass immediately without proving they can catch bugs.
- **Impact:** Tests won't catch bugs introduced in Tasks 1–4 because they're written against the already-changed code. For example, the sentinel error test in Task 5 will pass trivially because `ErrCancelled` was already defined in Task 2 — it never verified that the old string comparison was actually replaced.
- **Suggested fix:** Restructure into Red-Green-Refactor cycles:
  - **Task 1 (RED):** Write failing test for `ErrCancelled` sentinel (`errors.Is`). **(GREEN):** Add `ErrCancelled` var, update selector to return it. **(REFACTOR):** Remove `cancelled` bool, simplify post-Stop logic.
  - **Task 2 (RED):** Write failing test for `RunContextSelector(contexts []string)` signature. **(GREEN):** Refactor signature, update callers.
  - **Task 3:** Add panic recovery (hard to test without mocking tview — acceptable as implementation-only).
  - **Task 4:** Clean up dead code.
  - **Task 5:** Integration verification (`go test`, `go vet`, `make build`).

#### 3. [MEDIUM] Task 3 panic recovery: `app.Stop()` called on potentially nil `app`

- **Plan says:** Deferred recovery calls `app.Stop()` to restore terminal state.
- **Actual:** The proposed code pattern is:
  ```go
  func runTviewSelector(contexts []string) (result string, err error) {
      defer func() {
          if r := recover(); r != nil {
              app.Stop()  // app is declared below
              ...
          }
      }()
      app := tview.NewApplication()
      ...
  }
  ```
  While Go closures capture variables by reference (so `app` will be assigned by the time the defer runs in most cases), if a panic occurs during or before `tview.NewApplication()`, `app` will be nil and `app.Stop()` will nil-pointer panic inside the recovery handler — defeating the purpose of panic recovery.
- **Impact:** A panic during `tview.NewApplication()` or any code before `app` assignment would cause a secondary panic in the recovery handler, leaving the terminal in a corrupted state.
- **Suggested fix:** Add a nil check: `if app != nil { app.Stop() }`. Or move `app := tview.NewApplication()` before the `defer` statement.

#### 4. [MEDIUM] Task 4 is redundant with Task 2 — dead code is already handled

- **Plan says (Task 4):** "In `runTviewSelector`, the `selectedContext == ""` check after app.Stop is dead code (ESC returns ErrCancelled now). Remove it or convert to a safety assertion."
- **Actual:** Task 2 already replaces the current code structure:
  ```go
  // CURRENT (to be replaced by Task 2):
  if cancelled {
      return "", fmt.Errorf("operation cancelled")
  }
  if selectedContext == "" {
      return "", fmt.Errorf("no context selected")  // <-- AC7 dead code
  }
  ```
  With:
  ```go
  // AFTER TASK 2:
  if selectedContext == "" {
      return "", ErrCancelled
  }
  ```
  Task 2 already removes the `"no context selected"` branch and replaces both checks with a single `selectedContext == ""` → `ErrCancelled`. Task 4's instruction to "remove or clarify the dead-code branch" has already been fulfilled by Task 2. The executor may be confused about what additional work Task 4 requires.
- **Impact:** Executor may perform unnecessary work or be confused about whether Task 4 requires a separate change.
- **Suggested fix:** Merge Task 4 into Task 2 as a sub-step, or rephrase Task 4 to clarify it's just a verification step ("Verify that after Task 2 changes, no dead-code branches remain in `runTviewSelector`").

#### 5. [LOW] Dev Notes line range slightly off for exec.go

- **Plan says:** "The basic TUI integration for `exec` is already implemented in `cmd/exec.go:25-43`"
- **Actual:** The `if useTUI` block spans lines 25–44 (closing brace on line 44). Line 43 is the inner closing `}` of the `if err != nil` block.
- **Impact:** Minor — won't cause execution failure but could mislead the executor.
- **Suggested fix:** Change to `cmd/exec.go:25-44`.

#### 6. [TRIVIAL] Task 1 subtask wording: "Remove `config` import" is implied but should be explicit about import cleanup

- **Plan says:** "Remove `"github.com/dsdashun/ccctx/config"` import from `internal/ui/selector.go`" — this is listed as a subtask.
- **Actual:** This is correct and explicit. No issue.
- **Impact:** None. Recording for completeness.

### Verification Summary

| Item | Plan claims | Actual | Match |
|------|------------|--------|-------|
| `internal/ui/selector.go` exists | TUI selector implementation | File exists, 123 lines | **CORRECT** |
| `cmd/exec.go` TUI block | Lines 25-43 | Lines 25-44 | **OFF BY ONE** |
| `cmd/run.go` TUI block | Lines 26-46 | Lines 26-47 (closing `}` on 47) | **CLOSE** |
| `RunContextSelector()` takes no args | Current signature | Line 12: `func RunContextSelector() (string, error)` | **CORRECT** |
| `config.ListContexts()` called in selector | Line 14 | Line 14: `contexts, err := config.ListContexts()` | **CORRECT** |
| `cancelled` bool exists in selector | Mutable flag | Line 62: `var cancelled bool` | **CORRECT** |
| `fmt.Errorf("operation cancelled")` in selector | Error return | Line 114: `return "", fmt.Errorf("operation cancelled")` | **CORRECT** |
| `err.Error() == "operation cancelled"` in cmd files | Fragile string check | `cmd/exec.go:37`, `cmd/run.go:40` | **CORRECT** |
| `testify` in go.mod | Test dependency | Line 10: `github.com/stretchr/testify v1.10.0` | **CORRECT** |
| `config` import in selector.go | Should be removed | Line 6: `"github.com/dsdashun/ccctx/config"` | **CORRECT** |
| `app.SetRoot(flex, false).SetFocus(list).Run()` | Panic recovery target | Line 108: exact match | **CORRECT** |

### Risk Summary

| # | Severity | Issue | Task |
|---|----------|-------|------|
| 1 | **HIGH** | Task 1 vs Task 5 contradiction: empty check removed then tested | Task 1, Task 5 |
| 2 | **HIGH** | TDD violation: all implementation before tests | Tasks 1-5 ordering |
| 3 | **MEDIUM** | Panic recovery: `app.Stop()` on nil `app` if panic before assignment | Task 3 |
| 4 | **MEDIUM** | Task 4 redundant with Task 2 — dead code already handled | Task 2, Task 4 |
| 5 | **LOW** | Line range off by one for exec.go TUI block | Dev Notes |
| 6 | **TRIVIAL** | Import cleanup already explicit | Task 1 |

### Verdict

The plan is **NOT ready for execution**. Fix the two high-severity issues first:

1. **Issue #1 (Task 1 vs Task 5 contradiction)** is the most critical — the executor will hit a wall at Task 5 when the test they're told to write can't pass because the behavior it tests was removed in Task 1. Decide whether to keep the empty-contexts validation in `RunContextSelector` (recommended) or change the test.

2. **Issue #2 (TDD violation)** means tests won't actually catch bugs. Restructure into Red-Green-Refactor cycles so each behavior has a failing test before implementation.

The medium issues (#3 nil `app` in recovery, #4 redundant task) should also be addressed but won't block execution.

## Round 1 Response

**Responded by:** claude
**Date:** 2026-04-19
**Plan file:** 2-1-exec-command-integrates-tui-interactive-selector.md (updated)

### Issue Responses

#### 1. [FIX] Task 1 contradicts Task 5: empty-contexts check removal vs. test expectation

**Change made:** Updated Task 1 to keep `len(contexts) == 0` validation as defensive input parameter checking. Only the `config.ListContexts()` call is removed. Added explicit note: "callers also check, but selector guards its own contract". Updated Signature Change Impact in Dev Notes to reflect the defensive check remains. Task 5's test for `RunContextSelector([]string{})` returning "no contexts found" is now consistent.

**Location:** Tasks → Task 1, Dev Notes → Signature Change Impact

#### 2. [FIX] TDD violation: all implementation before all tests

**Change made:** Restructured tasks into RED-GREEN-Refactor pattern:
- Task 1: Signature refactor (mechanical change, no new behavior to test)
- Task 2 [RED]: Write failing tests for `ErrCancelled` sentinel and empty contexts — tests fail because `ErrCancelled` doesn't exist yet
- Task 3 [GREEN]: Implement sentinel error, remove `cancelled` bool, update callers — tests from Task 2 now pass
- Task 4: Panic recovery (hard to test without mocking tview, accepted as implementation-only)
- Task 5: Integration verification

**Location:** Tasks → Tasks 2–5 (renumbered from original Tasks 2–6)

#### 3. [FIX] Task 3 panic recovery: `app.Stop()` called on potentially nil `app`

**Change made:** Updated Panic Recovery Pattern in Dev Notes to include `if app != nil { app.Stop() }` nil check. Updated Task 4 (new numbering) subtask to explicitly mention the nil check and its rationale. Added note explaining the guard protects against panics during or before `tview.NewApplication()`.

**Location:** Tasks → Task 4, Dev Notes → Panic Recovery Pattern

#### 4. [FIX] Task 4 redundant with Task 2 — dead code already handled

**Change made:** Merged original Task 4 (AC7 dead code cleanup) into Task 3 (sentinel error). Added explicit subtask: "Remove `"no context selected"` dead-code branch (AC: #7) — `selectedContext == ""` now always means cancellation, never an unreachable 'no selection' case". Updated Race Condition Fix Strategy code comment to `// Post-Stop (replaces both cancelled check and "no context selected" branch):`. Updated Project Structure Notes to note AC7 is merged into sentinel error task.

**Location:** Tasks → Task 3, Dev Notes → Race Condition Fix Strategy, Dev Notes → Project Structure Notes

#### 5. [FIX] Dev Notes line range slightly off for exec.go

**Change made:** Changed `cmd/exec.go:25-43` to `cmd/exec.go:25-44` in Dev Notes.

**Location:** Dev Notes → Current State

#### 6. [WONTFIX] Task 1 subtask wording: import cleanup already explicit

**Reason:** Reviewer confirmed no issue — recording for completeness only.

### Summary

All issues from Round 1 have been addressed. The plan is ready for re-review.

---

## Round 2

**Context:** Re-verified the updated plan against the actual codebase. Checked all Round 1 issue resolutions and scanned for new issues.

### Round 1 Issue Status

| # | Severity | Issue | Round 2 Status |
|---|----------|-------|----------------|
| 1 | HIGH | Task 1 vs Task 5 contradiction: empty check removed then tested | **FIXED** — Task 1 now explicitly keeps `len(contexts) == 0` as defensive input parameter checking |
| 2 | HIGH | TDD violation: all implementation before tests | **FIXED** — Restructured into RED (Task 2) / GREEN (Task 3) phases |
| 3 | MEDIUM | Panic recovery: `app.Stop()` on nil `app` | **FIXED** — Task 4 includes `if app != nil { app.Stop() }` nil check with rationale |
| 4 | MEDIUM | Task 4 redundant with Task 2 — dead code already handled | **FIXED** — Dead code removal merged into Task 3 as explicit substep |
| 5 | LOW | Line range off by one for exec.go TUI block | **PARTIALLY FIXED** — Dev Notes Current State updated to `cmd/exec.go:25-44`, but References section still shows `cmd/exec.go:25-43` |
| 6 | TRIVIAL | Import cleanup already explicit | **WONTFIX** — Confirmed no action needed |

### Verification Summary

| Item | Plan claims | Actual | Match |
|------|------------|--------|-------|
| `RunContextSelector()` takes no args | Current signature to be refactored | Line 12: `func RunContextSelector() (string, error)` | **CORRECT** |
| `config.ListContexts()` called in selector | To be removed | Line 14: `contexts, err := config.ListContexts()` | **CORRECT** |
| `cancelled` bool in selector | To be removed | Line 62: `var cancelled bool` | **CORRECT** |
| `fmt.Errorf("operation cancelled")` in selector | To be replaced with sentinel | Line 114: `return "", fmt.Errorf("operation cancelled")` | **CORRECT** |
| `err.Error() == "operation cancelled"` in cmd files | Fragile string to be replaced | `cmd/exec.go:37`, `cmd/run.go:40` | **CORRECT** |
| `"no context selected"` dead code branch | To be removed | Line 118-119 in selector.go | **CORRECT** |
| `config` import in selector.go | To be removed | Line 6: `"github.com/dsdashun/ccctx/config"` | **CORRECT** |
| `testify` in go.mod | Test dependency | go.mod line 10: `github.com/stretchr/testify v1.10.0` | **CORRECT** |
| `errors` import in cmd files | To be added | Neither cmd/run.go nor cmd/exec.go currently import "errors" | **CORRECT** — confirmed import is absent and needs adding |
| References: `cmd/exec.go:25-43` | Informational reference | Actual TUI block is lines 25-44 | **OFF BY ONE** |
| References: `cmd/run.go:26-46` | Informational reference | Actual TUI block is lines 26-47 | **OFF BY ONE** |

### Issues Found

#### 1. [LOW] References section line ranges not updated after Round 1 fixes

- **Plan says (References section):** `cmd/exec.go:25-43` and `cmd/run.go:26-46`
- **Actual:** Dev Notes Current State correctly shows `cmd/exec.go:25-44`, but the References section at the bottom still shows the old `25-43`. `cmd/run.go` TUI block spans lines 26-47, but the plan shows `26-46` — this was never updated in either location.
- **Impact:** Misleading for the executor tracing references. Not executable, so won't cause failures.
- **Suggested fix:** Update References section to `cmd/exec.go:25-44` and `cmd/run.go:26-47`.

#### 2. [LOW] Task 2 second test is not truly RED

- **Plan says (Task 2 [RED phase]):** "Test: `RunContextSelector([]string{})` returns error containing 'no contexts found'"
- **Actual:** After Task 1 keeps the `len(contexts) == 0` defensive check, this test will pass immediately — the behavior it tests already exists. Only the first test (`ErrCancelled` sentinel) is genuinely RED.
- **Impact:** The executor may be confused why one of the "RED phase" tests passes immediately. No execution failure risk.
- **Suggested fix:** Either move this test to Task 1 (as a validation of the refactored signature) or add a note: "This test validates existing defensive behavior kept from Task 1 — it will pass immediately, which is expected."

### Risk Summary

| # | Severity | Issue | Location |
|---|----------|-------|----------|
| 1 | LOW | References section line ranges outdated | References section |
| 2 | LOW | Task 2 second test not truly RED | Task 2 |

### Verdict

**The plan is ready for execution.** All Round 1 HIGH and MEDIUM issues are properly resolved. The two remaining issues are LOW — neither will cause execution failure or incorrect output. The TDD structure is sound (genuine RED phase for sentinel error, GREEN phase implementation, panic recovery as accepted implementation-only change). Code logic traces correctly through all exit paths.

---

## Round 3

**Context:** User story simulation. Assume all plan tasks are fully implemented. Walk through the user story as an end-user, exercising every acceptance criterion against the post-implementation system. Trace the full call chain from CLI input through `ParseArgs` → TUI selector → `runner.New` → `runner.Run` for each scenario.

### User Story

> As a **ccctx user**, I want **the TUI interactive selector to work reliably for both `run` and `exec` commands**, so that **I don't need to memorize provider names and can quickly select a provider in either command without encountering terminal corruption, race conditions, or misleading errors**.

### Scenario Walkthrough

#### Scenario 1: `ccctx exec` (no args, has contexts) → AC1 happy path

**Setup:** Config has 3 contexts: `work`, `personal`, `staging`. `$SHELL=/bin/zsh`.

1. User runs `ccctx exec`
2. `ParseArgs([]string{})` → `provider="", targetArgs=[], useTUI=true, err=nil`
3. `cmd/exec.go`: `useTUI=true` → `config.ListContexts()` → `["work", "personal", "staging"]`
4. `len(contexts) != 0` → passes caller-side check
5. `ui.RunContextSelector(["work", "personal", "staging"])` → defensive `len != 0` check passes → `runTviewSelector(["work", "personal", "staging"])`
6. TUI renders list with 3 items
7. User presses `j` twice → highlights `staging` (input capture intercepts, returns nil)
8. User presses Enter → `SetSelectedFunc` fires → `selectedContext = "staging"` → `app.Stop()`
9. Post-Stop: `selectedContext != ""` → returns `("staging", nil)`
10. Back in `cmd/exec.go`: `err == nil` → `provider = "staging"`
11. `len(targetArgs) == 0` → `shell = "/bin/zsh"` → `targetArgs = ["/bin/zsh"]`
12. `runner.New({ContextName: "staging", Target: ["/bin/zsh"]})` → `config.GetContext("staging")` → resolves env vars → validates `BaseURL` and `AuthToken` → builds env slice
13. `r.Run()` → `exec.Command("/bin/zsh")` with filtered env + `ANTHROPIC_BASE_URL` + `ANTHROPIC_AUTH_TOKEN`
14. **User lands in zsh with staging's env vars.** ✓

**Expected:** User gets a shell with the selected context's environment.
**Actual:** Trace matches expected behavior. AC1 satisfied.

#### Scenario 2: `ccctx exec` (no args, no contexts) → AC1 error path

**Setup:** Config has 0 contexts.

1. User runs `ccctx exec`
2. `ParseArgs([]string{})` → `useTUI=true`
3. `config.ListContexts()` → `[]string{}`
4. `len(contexts) == 0` → `fmt.Fprintf(os.Stderr, "Error: no contexts found\n")` → `os.Exit(1)`

**Expected:** "Error: no contexts found" on stderr, exit 1.
**Actual:** Trace matches. The selector is never invoked. AC1 satisfied.

#### Scenario 3: `ccctx exec` → ESC → AC8

1. User runs `ccctx exec`
2. TUI appears with context list
3. User presses ESC → input capture: `app.Stop()`, returns `nil` (event consumed)
4. `selectedContext` stays `""`
5. Post-Stop: `selectedContext == ""` → returns `("", ui.ErrCancelled)`
6. Back in `cmd/exec.go`: `errors.Is(err, ui.ErrCancelled)` → `true`
7. `fmt.Println("Operation cancelled.")` → stdout
8. `os.Exit(1)`

**Expected:** "Operation cancelled." on stdout, exit code 1.
**Actual:** Trace matches. AC8 satisfied.

#### Scenario 4: `ccctx run` → select → AC9 regression check

1. User runs `ccctx run`
2. TUI appears → user selects "work"
3. Returns `("work", nil)` → `provider = "work"`
4. `exec.LookPath("claude")` → finds `/usr/local/bin/claude`
5. `target = ["/usr/local/bin/claude"]`
6. `runner.New({ContextName: "work", Target: ["/usr/local/bin/claude"]})` → runs claude with work's env

**Expected:** Claude runs with selected context, identical to pre-plan behavior.
**Actual:** Trace matches. Only changes are: selector receives `contexts` param, sentinel error instead of string comparison. No behavioral change for the happy path. AC9 satisfied.

#### Scenario 5: `ccctx run` → ESC → AC8

1-6. Same as Scenario 3 but through `cmd/run.go`
7. `errors.Is(err, ui.ErrCancelled)` → `true`
8. `fmt.Println("Operation cancelled.")` → stdout, exit 1

**Expected:** Same as exec. Exit code 1.
**Actual:** Trace matches. AC8 satisfied for both commands.

#### Scenario 6: `ccctx exec -- vim` → TUI + command forwarding

1. `ParseArgs(["--", "vim"])` → `separatorIndex=0`, `contextArgs=[]`, `targetArgs=["vim"]`, `useTUI=true`
2. TUI appears → user selects "personal"
3. `provider = "personal"`, `targetArgs = ["vim"]` (non-empty, no SHELL fallback)
4. `runner.New({ContextName: "personal", Target: ["vim"]})` → runs vim with personal's env

**Expected:** vim opens with selected context's environment.
**Actual:** Trace matches. SHELL fallback only triggers when `len(targetArgs) == 0`. ✓

#### Scenario 7: tview panic during TUI → AC6

1. User runs `ccctx exec`
2. TUI launches → something panics inside tview
3. Deferred recovery catches panic
4. `app != nil` → `app.Stop()` restores terminal state
5. Returns `("", fmt.Errorf("TUI error: %v", r))`
6. Back in `cmd/exec.go`: `errors.Is(err, ui.ErrCancelled)` → `false` → falls through
7. `fmt.Fprintf(os.Stderr, "Error: %v\n", err)` → "Error: TUI error: ..."
8. `os.Exit(1)`

**Expected:** Terminal state restored, meaningful error message, exit 1.
**Actual:** Trace matches. AC6 satisfied.

#### Scenario 8: Tab key navigation in TUI

1. TUI is showing with 3 contexts
2. User presses Tab → input capture does not consume Tab → event propagates to list
3. `SetDoneFunc` fires: `currentIndex = list.GetCurrentItem()` → bounds check passes → `selectedContext = contexts[currentIndex]` → `app.Stop()`
4. Post-Stop: `selectedContext != ""` → returns `(selectedContext, nil)`

**Expected:** Tab selects the currently highlighted item and closes TUI.
**Actual:** Trace matches. ✓

### Issues Found

#### 3. [MEDIUM] Test coverage gap: `RunContextSelector` returning `ErrCancelled` on ESC is untested

- **Plan says (Task 2):** "Full TUI rendering tests (tview `app.Run()`) are out of scope — focus on testable error paths only."
- **Actual:** The test for `ErrCancelled` only verifies that the sentinel variable exists and `errors.Is` works against it. No test verifies that `RunContextSelector` actually *returns* `ErrCancelled` when the user cancels. The cancellation path (ESC → input capture → `app.Stop()` → `selectedContext == ""` → return `ErrCancelled`) is the core behavioral change in this story, and it has zero automated coverage.
- **Impact:** If a future refactor accidentally changes the cancellation path to return a different error (e.g., someone adds back the `cancelled` bool and forgets to return `ErrCancelled`), no test will catch it. The sentinel error exists but nothing verifies it's used at the right place.
- **Suggested fix:** Consider adding a targeted test that calls `runTviewSelector` directly with a context that has at least 1 item, and verify the `selectedContext == "" → ErrCancelled` logic path. This could be done by extracting the post-Stop decision logic into a testable helper:

  ```go
  func resolveSelection(selectedContext string) (string, error) {
      if selectedContext == "" {
          return "", ErrCancelled
      }
      return selectedContext, nil
  }
  ```

  Then test `resolveSelection("")` returns `ErrCancelled` and `resolveSelection("work")` returns `"work"`. This tests the core logic without needing tview. However, this introduces a new helper — the plan author should decide if the added coverage is worth the extra function. If not, at minimum add a comment in the test file acknowledging the untested path.

#### 4. [LOW] SetDoneFunc bounds-check failure would return misleading `ErrCancelled`

- **Plan says:** After removing `cancelled` bool, `selectedContext == ""` "now always means cancellation" (Dev Notes → Race Condition Fix Strategy).
- **Actual:** This is almost always true, but there's a theoretical edge case in `SetDoneFunc`:

  ```go
  list.SetDoneFunc(func() {
      currentIndex := list.GetCurrentItem()
      if currentIndex >= 0 && currentIndex < len(contexts) {
          selectedContext = contexts[currentIndex]
      }
      app.Stop()
  })
  ```

  If the bounds check `currentIndex >= 0 && currentIndex < len(contexts)` fails (e.g., tview returns -1), `selectedContext` stays `""` and the user gets `ErrCancelled` — implying they cancelled, when actually they pressed Tab and a tview edge case prevented selection. The user would see "Operation cancelled." despite having taken a selection action.
- **Impact:** Extremely unlikely in practice — `list.GetCurrentItem()` returns a valid index when items exist, and we guard against empty contexts. But the claim "always means cancellation" is not strictly true.
- **Suggested fix:** Either (a) remove the bounds check entirely since it can't fail given the empty-context guard, making the intent clearer, or (b) change the bounds-check failure to return a distinct error:

  ```go
  // In SetDoneFunc, remove bounds check entirely:
  selectedContext = contexts[list.GetCurrentItem()]
  app.Stop()
  ```

  This is safe because `len(contexts) > 0` is guaranteed before calling `runTviewSelector`, and `GetCurrentItem()` returns `[0, len(items))` when items exist.

### Scenario Verification Summary

| # | Scenario | AC | Result |
|---|----------|----|--------|
| 1 | `ccctx exec` → select → shell | AC1 | **PASS** — lands in $SHELL with env vars |
| 2 | `ccctx exec` → no contexts | AC1 | **PASS** — "Error: no contexts found" stderr, exit 1 |
| 3 | `ccctx exec` → ESC | AC8 | **PASS** — "Operation cancelled." stdout, exit 1 |
| 4 | `ccctx run` → select | AC9 | **PASS** — claude runs, no regression |
| 5 | `ccctx run` → ESC | AC8 | **PASS** — "Operation cancelled." stdout, exit 1 |
| 6 | `ccctx exec -- vim` → select | AC1 | **PASS** — vim opens with selected context env |
| 7 | tview panic during TUI | AC6 | **PASS** — terminal restored, meaningful error |
| 8 | Tab key in TUI | — | **PASS** — selects highlighted item |
| — | `ccctx exec mycontext` (no TUI) | AC1 | **PASS** — direct context, no TUI invoked |
| — | TUI shared between run/exec | AC2 | **PASS** — both call same `ui.RunContextSelector` |
| — | Sentinel error in callers | AC4 | **PASS** — both use `errors.Is(err, ui.ErrCancelled)` |
| — | Race condition eliminated | AC5 | **PASS** — no `cancelled` bool, single `selectedContext == ""` check |
| — | Dead code removed | AC7 | **PASS** — no "no context selected" branch |

### Risk Summary

| # | Severity | Issue | Scenario |
|---|----------|-------|----------|
| 1 | LOW | References section line ranges outdated | N/A |
| 2 | LOW | Task 2 second test not truly RED | Task 2 |
| 3 | MEDIUM | Cancellation return path untested | ESC → ErrCancelled |
| 4 | LOW | SetDoneFunc bounds-check false → misleading ErrCancelled | Tab key edge case |

### Verdict

**The plan is ready for execution.** All user story scenarios trace correctly through the full call chain. The user-facing behavior matches expectations for every AC. The two new issues are MEDIUM and LOW — the MEDIUM item (untested cancellation return path) is worth considering but doesn't block execution since the plan explicitly acknowledges TUI rendering tests are out of scope. The LOW item (SetDoneFunc bounds check) is theoretical.

---

## Round 3 Response

**Responded by:** claude
**Date:** 2026-04-19
**Plan file:** 2-1-exec-command-integrates-tui-interactive-selector.md (updated)

### Issue Responses

#### 3. [WONTFIX] Test coverage gap: RunContextSelector returning ErrCancelled on ESC is untested

**Reason:** Adding a `resolveSelection` helper function solely to test a 3-line if/else (`if selectedContext == "" { return "", ErrCancelled } else { return selectedContext, nil }`) is over-engineering. The plan explicitly scopes out TUI rendering tests, and the sentinel error test in Task 2 verifies the contract that `errors.Is(err, ErrCancelled)` works — which is the meaningful assertion. The return logic itself is trivially correct; if it were broken, the manual verification in Task 5 would catch it. Rather than adding a helper, I've added a note in Task 2 acknowledging the untested path and explaining why it's acceptable.

**Location:** Tasks → Task 2 (new note added)

#### 4. [FIX] SetDoneFunc bounds-check failure would return misleading ErrCancelled

**Change made:** Updated Task 3 to explicitly remove the bounds check in `SetDoneFunc`. Added a new subtask: "Remove bounds check in `SetDoneFunc` — replace `if currentIndex >= 0 && currentIndex < len(contexts) { selectedContext = contexts[currentIndex] }` with `selectedContext = contexts[list.GetCurrentItem()]`" with rationale that `len(contexts) > 0` is guaranteed and `GetCurrentItem()` returns a valid index when items exist. Updated the "All 3 exit paths" subtask to remove "bounds-checked" qualifier. Updated Dev Notes → Race Condition Fix Strategy to show the direct access pattern without bounds check.

**Location:** Tasks → Task 3, Dev Notes → Race Condition Fix Strategy

### Summary

- Fixed: 1 issue (Issue 4 — SetDoneFunc bounds check removal)
- Wontfix: 1 issue (Issue 3 — untested cancellation return path)
- Plan file: updated

All issues from Round 3 have been addressed. The plan is ready for re-review.
