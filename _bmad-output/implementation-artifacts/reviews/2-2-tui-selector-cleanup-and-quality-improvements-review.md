---
reviewed_plan: 2-2-tui-selector-cleanup-and-quality-improvements.md
reviewer: glm
latest_round: 1
review_rounds:
  - round: 1
    date: 2026-04-19
---

# Review: 2-2-tui-selector-cleanup-and-quality-improvements

Reviewed plan: `_bmad-output/implementation-artifacts/2-2-tui-selector-cleanup-and-quality-improvements.md`

## Round 1

**Context:** Initial full dry-run of the plan against the actual codebase. Verified all file paths, line numbers, code claims, and test references.

### Issues Found

#### 1. [MEDIUM] Task 7 claims to "add" a test that already exists

- **Plan says:** Task 7 subtask: "Add test for `RunContextSelector([]string{})` returning 'no contexts found' — validates defensive check still works after changes"
- **Actual:** `TestRunContextSelector_EmptyContexts` already exists at `internal/ui/selector_test.go:50-54` with identical behavior: calls `RunContextSelector([]string{})`, asserts error containing "no contexts found".
- **Impact:** Executor will either write a duplicate test or waste time investigating whether the existing test counts. Creates confusion about what's actually needed.
- **Suggested fix:** Change to "Verify existing `TestRunContextSelector_EmptyContexts` still passes after changes" or remove the subtask entirely since the test already exists and is already covered by the `go test ./...` step.

#### 2. [MEDIUM] Constant naming inconsistency between Tasks and Dev Notes

- **Plan says:** Task 1 defines the constant as `flexHeightPadding` (lines 32, 33, 38). Dev Notes "Dynamic Width Calculation" code snippet defines it as `flexHeightPad` (lines 98, 108).
- **Actual:** Two different names for the same constant appear in the plan itself.
- **Impact:** Executor won't know which name to use. If they follow Task 1 literally, they'll write `flexHeightPadding`; if they reference the Dev Notes snippet, they'll write `flexHeightPad`. Inconsistent code if different parts of the implementation follow different sections.
- **Suggested fix:** Pick one name and use it consistently. Update all occurrences — either change Dev Notes to use `flexHeightPadding` or change Task 1 to use `flexHeightPad`.

#### 3. [MEDIUM] Dev Notes contradict Task 7 on test needs

- **Plan says:** Dev Notes (File Change Summary, `selector_test.go` row): "Verify existing tests pass, no new tests needed for visual/layout changes". But Task 7 says "Add test for `RunContextSelector([]string{})`".
- **Actual:** These two statements conflict — one says no new tests, the other says add a test.
- **Impact:** Executor must resolve the contradiction themselves, which may lead to skipping a needed test or writing an unnecessary one.
- **Suggested fix:** Align Dev Notes and Task 7. Since the test already exists (see issue #1), both should say "verify existing `TestRunContextSelector_EmptyContexts` passes."

#### 4. [LOW] TDD ordering: tests after implementation

- **Plan says:** Tasks 1-6 are implementation changes. Task 7 is "Update and add tests" — run after all implementation.
- **Actual:** This is the test-after pattern. However, acknowledging the context: these are quality improvements to TUI internals (mechanical renames, error path cleanup, stderr redirect), not new features. Most changes are internal refactors that existing tests already indirectly cover. The TUI layer is inherently hard to unit test without mocking tview internals.
- **Impact:** Minor — these are refactors with existing test coverage. Risk of regression is low since the plan includes `go test ./...` and `go vet ./...` as verification gates.
- **Suggested fix:** Consider moving the "verify existing tests pass" step to run after each task or after a logical group (e.g., after Tasks 1-3 for selector.go changes), rather than deferring all verification to Task 7. This catches regressions earlier. Not blocking given the nature of the changes.

#### 5. [LOW] Task 8 "Remove or mark as resolved" is ambiguous

- **Plan says:** Task 8: "Remove or mark as resolved the 6 TUI deferred items addressed by this story"
- **Actual:** "Remove or mark" gives two options without guidance on which to choose. Should the items be deleted from deferred-work.md, or annotated with a resolution note?
- **Impact:** Executor must make a judgment call. Either approach works, but the plan should be explicit.
- **Suggested fix:** Pick one approach. Recommended: mark items with a resolution note (e.g., `— resolved by Story 2.2`) rather than deleting, to preserve history.

#### 6. [TRIVIAL] Task 3 lists a no-op subtask

- **Plan says:** Task 3 subtask: "SetDoneFunc: keep `contexts[list.GetCurrentItem()]` — already correct (uses data source)"
- **Actual:** This subtask describes doing nothing ("keep" existing code). It's documentation, not an action item.
- **Impact:** No functional impact, but adds noise to the task list.
- **Suggested fix:** Either remove this subtask or convert it to a note/rationale rather than a checklist item. The "keep" wording is fine as context but shouldn't be a checkable action.

### Verification Summary

| Item | Plan claims | Actual | Match |
|------|------------|--------|-------|
| `cmd/exec.go:39` cancellation message | `fmt.Println("Operation cancelled.")` | Line 39: `fmt.Println("Operation cancelled.")` | CORRECT |
| `cmd/run.go:42` cancellation message | `fmt.Println("Operation cancelled.")` | Line 42: `fmt.Println("Operation cancelled.")` | CORRECT |
| `internal/ui/selector.go:53` hardcoded width | `flex.SetRect(0, 1, 50, maxItems+4)` | Line 53: `flex.SetRect(0, 1, 50, maxItems+4)` | CORRECT |
| `selector.go:87` divergent source | `selectedContext = mainText` | Line 87: `selectedContext = mainText` | CORRECT |
| `selector.go:82` SetDoneFunc source | `contexts[list.GetCurrentItem()]` | Line 82: `selectedContext = contexts[list.GetCurrentItem()]` | CORRECT |
| `selector.go:30` panic recovery format | `fmt.Errorf("TUI error: %v", r)` | Line 30: `err = fmt.Errorf("TUI error: %v", r)` | CORRECT |
| `runtime/debug` not imported | Need to add import | Grep confirms no `runtime/debug` in selector.go | CORRECT |
| `min()` builtin available | Go 1.21+ builtins | go.mod: `go 1.23.3` | CORRECT |
| `TestRunContextSelector_EmptyContexts` exists | Plan says "add" this test | Already exists at selector_test.go:50-54 | **CONFLICT** — see issue #1 |
| 6 deferred-work.md items match | Listed in Task 8 | All 6 items confirmed in deferred-work.md lines 31-36 | CORRECT |
| Constant name in Tasks | `flexHeightPadding` | Dev Notes use `flexHeightPad` | **MISMATCH** — see issue #2 |

### Risk Summary

| # | Severity | Issue | Task |
|---|----------|-------|------|
| 1 | MEDIUM | Test already exists, plan says "add" | Task 7 |
| 2 | MEDIUM | Constant name inconsistency (`flexHeightPadding` vs `flexHeightPad`) | Task 1 vs Dev Notes |
| 3 | MEDIUM | Dev Notes vs Task 7 contradiction on test needs | Dev Notes vs Task 7 |
| 4 | LOW | Test-after ordering | Tasks 1-6 vs Task 7 |
| 5 | LOW | "Remove or mark" ambiguity in deferred-work cleanup | Task 8 |
| 6 | TRIVIAL | No-op subtask listed as action item | Task 3 |

### Verdict

The plan is in **good shape for execution**. No blocking or high-severity issues found. All file paths, line numbers, and code references verified correctly against the codebase.

Three medium-severity issues should be addressed before execution:
1. **Remove the duplicate "add test" subtask** — the test already exists; change to "verify existing test passes."
2. **Unify constant naming** — pick `flexHeightPadding` or `flexHeightPad` and use it consistently across Tasks and Dev Notes.
3. **Align Dev Notes with Task 7** — both should agree on whether new tests are needed (they aren't; the test exists).

Once these consistency fixes are made, the plan is ready for execution. The implementation changes themselves (Tasks 1-6) are well-specified with correct code snippets and verified references.

## Round 1 Response

**Responded by:** kimi
**Date:** 2026-04-19
**Plan file:** 2-2-tui-selector-cleanup-and-quality-improvements.md (updated)

### Issue Responses

#### 1. [FIX] Task 7 claims to "add" a test that already exists

**Change made:** Removed the "Add test for `RunContextSelector([]string{})`" subtask from Task 7. Replaced with "Verify existing `selector_test.go` tests still pass (including `TestRunContextSelector_EmptyContexts`)". Verified `TestRunContextSelector_EmptyContexts` exists at `internal/ui/selector_test.go:50-54` and already covers the defensive check.
**Location:** Tasks section, Task 7

#### 2. [FIX] Constant naming inconsistency between Tasks and Dev Notes

**Change made:** Unified constant name to `flexHeightPadding` everywhere. Updated Dev Notes "Dynamic Width Calculation" snippet (lines 98 and 108) from `flexHeightPad` to `flexHeightPadding`. Task 1 already used `flexHeightPadding`, which is the more descriptive name.
**Location:** Dev Notes, Dynamic Width Calculation code block

#### 3. [FIX] Dev Notes contradict Task 7 on test needs

**Change made:** Aligned both sections. Task 7 now says "Verify existing tests pass" with no mention of adding new tests. The Dev Notes File Change Summary already stated "no new tests needed for visual/layout changes" — this is now consistent with the updated Task 7.
**Location:** Tasks section, Task 7

#### 4. [WONTFIX] TDD ordering: tests after implementation

**Reason:** These changes are internal refactors (mechanical renames, error path cleanup, stderr redirect) with existing test coverage. The reviewer correctly noted this is "not blocking given the nature of the changes." Deferring all verification to Task 7 is acceptable here because the changes don't add new behavior — they modify implementation details that existing tests already indirectly cover. The plan includes `go test ./...` and `go vet ./...` as verification gates. Adding per-task test runs would add noise without significant risk reduction.

#### 5. [FIX] Task 8 "Remove or mark as resolved" is ambiguous

**Change made:** Changed Task 8 from "Remove or mark as resolved" to "Mark the 6 TUI deferred items as resolved with a resolution note (e.g., `— resolved by Story 2.2`)". This preserves history while clearly indicating resolution.
**Location:** Tasks section, Task 8

#### 6. [FIX] Task 3 lists a no-op subtask

**Change made:** Removed the no-op "SetDoneFunc: keep..." checklist item. Converted the information into a note under the remaining `SetSelectedFunc` subtask: "Note: `SetDoneFunc` already uses `contexts[list.GetCurrentItem()]` (data source), so no change needed there." This preserves the rationale without creating a checkable no-op action.
**Location:** Tasks section, Task 3

### Summary

All issues from Round 1 have been addressed. The plan is ready for re-review.
