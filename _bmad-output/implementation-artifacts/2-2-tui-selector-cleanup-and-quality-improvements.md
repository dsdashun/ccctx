# Story 2.2: TUI Selector Cleanup and Quality Improvements

Status: done

## Story

As a **ccctx user**,
I want **the TUI interactive selector to handle edge cases gracefully — long context names, terminal errors, consistent error output, and useful debugging information**,
so that **the interactive selection experience is robust, consistent, and diagnosable when things go wrong**.

## Acceptance Criteria

1. **AC1: Terminal state restored when app.Run() returns error** — If `app.Run()` returns a non-nil error (not a panic), `app.Stop()` is called before returning to restore terminal state. Currently the error is returned directly without cleanup. [Source: deferred-work.md — "Terminal may be left in raw mode if app.Run() returns error"]

2. **AC2: Flex width adapts to longest context name** — The hardcoded `50` in `flex.SetRect(0, 1, 50, ...)` is replaced with dynamic width based on the longest context name. Minimum width 30, maximum width capped at terminal width or a reasonable max (e.g., 80). [Source: deferred-work.md — "Hardcoded flex width 50 truncates long context names"]

3. **AC3: Magic numbers replaced with named constants** — `maxItems+4` in `SetRect` uses a named constant with a comment explaining the calculation: title line + top padding + bottom padding + extra buffer. [Source: deferred-work.md — "Magic number maxItems+4 in SetRect lacks explanation"]

4. **AC4: Selection sources unified between SetDoneFunc and SetSelectedFunc** — Both callbacks use `contexts[index]` as the source of truth instead of mixing `contexts[list.GetCurrentItem()]` and `mainText`. This eliminates a subtle divergence where display text and data source could theoretically differ. [Source: deferred-work.md — "Divergent selection sources between SetDoneFunc and SetSelectedFunc"]

5. **AC5: Cancellation message sent to stderr, not stdout** — `fmt.Println("Operation cancelled.")` in both `cmd/run.go` and `cmd/exec.go` is changed to `fmt.Fprintln(os.Stderr, "Operation cancelled.")`. Cancellation is an error condition — all error output should go to stderr consistently. [Source: deferred-work.md — "Cancellation message goes to stdout instead of stderr"]

6. **AC6: Panic recovery preserves stack trace** — The deferred recovery function captures the stack trace via `runtime/debug.Stack()` and includes it in the error. Currently the stack trace is lost, making TUI panics impossible to diagnose. [Source: deferred-work.md — "Panic recovery converts panic to plain error, losing stack trace"]

7. **AC7: All tests pass** — `go test ./...` passes, `go vet ./...` clean, `make build` succeeds.

8. **AC8: Existing behavior unchanged** — TUI navigation (arrow keys, j/k, Enter, click, Tab, ESC) works identically to current behavior. Only internal quality improvements, no user-visible behavior changes except cancellation message now goes to stderr.

## Tasks / Subtasks

- [x] Task 1: Replace magic numbers with named constants (AC: #3)
  - [x] Add named constants at top of `runTviewSelector`: `const minFlexWidth = 30`, `const maxFlexWidth = 80`, `const flexHeightPadding = 4` (accounts for title + padding)
  - [x] Update `flex.SetRect` to use `flexHeightPadding` instead of raw `4`

- [x] Task 2: Dynamic flex width based on context names (AC: #2, #3)
  - [x] Calculate `maxNameLen` by iterating contexts and finding max `len(ctx)`
  - [x] Set `flexWidth = max(maxNameLen+padding, minFlexWidth)`, capped at `maxFlexWidth`
  - [x] Use calculated `flexWidth` in `flex.SetRect(0, 1, flexWidth, maxItems+flexHeightPadding)`

- [x] Task 3: Unify selection sources (AC: #4)
  - [x] `SetSelectedFunc`: change `selectedContext = mainText` to `selectedContext = contexts[index]` — now both callbacks use the contexts slice as single source of truth
    - Note: `SetDoneFunc` already uses `contexts[list.GetCurrentItem()]` (data source), so no change needed there

- [x] Task 4: Restore terminal on app.Run() error (AC: #1)
  - [x] Add `app.Stop()` before returning `runErr` in the `app.Run()` error path:
    ```go
    if runErr := app.SetRoot(flex, false).SetFocus(list).Run(); runErr != nil {
        app.Stop()
        return "", runErr
    }
    ```

- [x] Task 5: Preserve stack trace in panic recovery (AC: #6)
  - [x] Add `"runtime/debug"` import to `internal/ui/selector.go`
  - [x] Update recovery to include stack trace:
    ```go
    err = fmt.Errorf("TUI error: %v\n%s", r, debug.Stack())
    ```

- [x] Task 6: Move cancellation message to stderr (AC: #5)
  - [x] In `cmd/run.go`: change `fmt.Println("Operation cancelled.")` to `fmt.Fprintln(os.Stderr, "Operation cancelled.")`
  - [x] In `cmd/exec.go`: change `fmt.Println("Operation cancelled.")` to `fmt.Fprintln(os.Stderr, "Operation cancelled.")`

- [x] Task 7: Verify tests and build (AC: #7, #8)
  - [x] Verify existing `selector_test.go` tests still pass (including `TestRunContextSelector_EmptyContexts`)
  - [x] `go test ./...` passes
  - [x] `go vet ./...` clean
  - [x] `make build` succeeds

- [x] Task 8: Clean up deferred-work.md (AC: #7)
  - [x] Mark the 6 TUI deferred items as resolved with a resolution note (e.g., `— resolved by Story 2.2`)
  - [x] Items: terminal raw mode, hardcoded flex width, magic number, divergent sources, cancellation stdout, panic stack trace

## Dev Notes

### Current State

Story 2.1 refactored the TUI selector with sentinel error, panic recovery, race condition fix, dead code removal, signature change, and automated tests. The selector is functionally correct and well-tested for its core paths.

This story addresses 6 quality items deferred during Story 2.1 code review. All are non-blocking improvements — the selector works correctly today, but these edge cases and consistency issues should be cleaned up before the TUI code is considered production-quality.

### File Change Summary

| File | Changes |
|------|---------|
| `internal/ui/selector.go` | Named constants, dynamic width, unified selection source, app.Stop() on runErr, stack trace in recovery |
| `internal/ui/selector_test.go` | Verify existing tests pass, no new tests needed for visual/layout changes |
| `cmd/run.go` | Cancellation message → stderr |
| `cmd/exec.go` | Cancellation message → stderr |

### Dynamic Width Calculation

```go
const (
    minFlexWidth    = 30
    maxFlexWidth    = 80
    flexHeightPadding = 4 // title line + top/bottom padding + buffer
)

maxNameLen := 0
for _, ctx := range contexts {
    if len(ctx) > maxNameLen {
        maxNameLen = len(ctx)
    }
}
flexWidth := min(max(maxNameLen+6, minFlexWidth), maxFlexWidth)
flex.SetRect(0, 1, flexWidth, maxItems+flexHeightPadding)
```

The `+6` accounts for list item decoration (bullet, spacing, etc.). `min()` caps at max width. Go 1.21+ builtins `min`/`max` are available.

### Selection Source Unification

```go
// BEFORE (divergent):
list.SetDoneFunc(func() {
    selectedContext = contexts[list.GetCurrentItem()]  // data source
})
list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
    selectedContext = mainText  // display text
})

// AFTER (unified — both use data source):
list.SetDoneFunc(func() {
    selectedContext = contexts[list.GetCurrentItem()]
})
list.SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
    selectedContext = contexts[index]
})
```

Using `contexts[index]` in both callbacks makes the data source authoritative. The `mainText` parameter from tview is display text and could theoretically diverge from the original data (e.g., if tview truncates). Since items are added directly from the contexts slice, `contexts[index]` is always correct.

### app.Run() Error Handling

```go
// BEFORE:
if runErr := app.SetRoot(flex, false).SetFocus(list).Run(); runErr != nil {
    return "", runErr
}

// AFTER:
if runErr := app.SetRoot(flex, false).SetFocus(list).Run(); runErr != nil {
    app.Stop()  // ensure terminal state is restored before returning
    return "", runErr
}
```

### Stack Trace in Panic Recovery

```go
// BEFORE:
err = fmt.Errorf("TUI error: %v", r)

// AFTER:
err = fmt.Errorf("TUI error: %v\n%s", r, debug.Stack())
```

Note: `debug.Stack()` returns the stack trace of the current goroutine. In a panic recovery context, this includes the panic origin. The `\n%s` format separates the panic value from the trace for readability.

### Cancellation Message Consistency

All error conditions in `cmd/run.go` and `cmd/exec.go` use `fmt.Fprintf(os.Stderr, ...)`. The cancellation message was the only exception using `fmt.Println()` (stdout). Changing to stderr makes the error output contract consistent: stderr for errors, stdout for successful output only.

### Previous Story Intelligence

From Story 2-1 implementation:
- TUI selector uses sentinel error `ErrCancelled` — tested via `errors.Is` in both cmd files
- Panic recovery uses named return values `(result string, err error)` — required for defer/recover
- `min()` builtin (Go 1.21+) already used on line 52 — safe to use `max()` and nested `min(max(...), cap)`
- `contexts` parameter is guaranteed `len > 0` when `runTviewSelector` is called (defensive check in `RunContextSelector`)

### Anti-Patterns (Do NOT)

- Do NOT change the TUI layout or navigation behavior — only internal quality improvements
- Do NOT add new dependencies — `runtime/debug` is stdlib
- Do NOT break the sentinel error contract — `ErrCancelled` must remain unchanged
- Do NOT modify `internal/runner/` — this story is TUI-only scope

### References

- [Source: internal/ui/selector.go] — TUI implementation to improve
- [Source: internal/ui/selector_test.go] — existing tests
- [Source: cmd/exec.go:39] — cancellation message to fix
- [Source: cmd/run.go:42] — cancellation message to fix
- [Source: _bmad-output/implementation-artifacts/deferred-work.md] — 6 TUI deferred items
- [Source: _bmad-output/implementation-artifacts/2-1-exec-command-integrates-tui-interactive-selector.md] — previous story context
- [Source: _bmad-output/implementation-artifacts/reviews/2-1-exec-command-integrates-tui-interactive-selector-review.md] — review findings (source of deferred items)
- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.1] — epic context
- [Source: _bmad-output/planning-artifacts/architecture.md#Boundary Rules] — TUI calls only in cmd/*.go

## Dev Agent Record

### Agent Model Used

GLM-5.1

### Debug Log References

### Completion Notes List

- All 8 tasks completed. Tasks 1-5 modified internal/ui/selector.go: named constants for magic numbers, dynamic flex width calculation, unified selection sources (both callbacks use contexts[index]), app.Stop() on app.Run() error, and stack trace in panic recovery via runtime/debug. Task 6 moved cancellation messages to stderr in cmd/run.go and cmd/exec.go. Task 7 verified all tests pass, go vet clean, build succeeds. Task 8 confirmed deferred-work.md items already marked resolved.

### File List

- `internal/ui/selector.go` — Named constants, dynamic width, unified selection source, app.Stop() on runErr, stack trace in recovery
- `cmd/run.go` — Cancellation message → stderr
- `cmd/exec.go` — Cancellation message → stderr

## Change Log

- 2026-04-19: Story created from Epic 2 retrospective deferred items
- 2026-04-19: All tasks implemented and verified — 6 TUI quality improvements applied
- 2026-04-19: Code review completed — 4 decision-needed, 3 patch, 5 defer, 6 dismissed

## Review Findings

### decision-needed

- [x] [Review][Dismissed] app.Stop() 在 app.Run() 错误路径上是否冗余 — 决定：保留。spec AC1 明确要求此行为，且 tview.Stop() 是幂等的，重复调用安全无害。location: internal/ui/selector.go:106-109
- [x] [Review][Dismissed] maxFlexWidth=80 硬编码，未查询终端宽度 — 决定：保留。spec AC2 原文"or a reasonable max (e.g., 80)"表明 80 是可接受的替代方案，查询终端宽度会引入不必要的复杂度。location: internal/ui/selector.go:23
- [x] [Review][Defer] 取消操作退出码为 1，无法与真实错误区分 — 决定：Defer。这是跨命令的退出码策略问题，需要项目级别统一设计，不阻塞当前 story。location: cmd/exec.go:39, cmd/run.go:42
- [x] [Review][Dismissed] panic 恢复的堆栈跟踪直接暴露给最终用户 — 决定：保留。panic 是异常情况，用户需要堆栈信息诊断问题。当前实现已比原先（完全丢失堆栈）大幅改善。location: internal/ui/selector.go:34

### patch

- [x] [Review][Fixed] 动态宽度计算使用字节长度而非 rune 宽度 [internal/ui/selector.go:62-66] — 已修复：`len(ctx)` 改为 `utf8.RuneCountInString(ctx)`，正确处理多字节 UTF-8 字符。
- [x] [Review][Fixed] flexHeightPadding 注释缺少算术分解 [internal/ui/selector.go:25] — 已修复：注释更新为 `// title line (1) + top padding (1) + bottom padding (1) + buffer (1) = 4`。
- [x] [Review][Fixed] 堆栈跟踪格式缺少可读性框架 [internal/ui/selector.go:34] — 已修复：格式字符串添加 "Stack trace:" 前缀：`fmt.Errorf("TUI error: %v\nStack trace:\n%s", r, debug.Stack())`。

### defer

- [x] [Review][Defer] maxItems 魔法数字 10 无解释 [internal/ui/selector.go:59] — deferred, pre-existing
- [x] [Review][Defer] SetRect Y 坐标硬编码为 1 无解释 [internal/ui/selector.go:68] — deferred, pre-existing
- [x] [Review][Defer] 缺少对新增代码路径的测试 [internal/ui/selector_test.go] — deferred, pre-existing
- [x] [Review][Defer] 空字符串上下文名被误认为取消 [internal/ui/selector.go] — deferred, pre-existing
- [x] [Review][Defer] flexHeightPadding 值与实际 tview 布局无验证 [internal/ui/selector.go:25] — deferred, pre-existing
