# Story 2.1: exec Command Integrates TUI Interactive Selector

Status: done

## Story

As a **ccctx user**,
I want **the TUI interactive selector to work reliably for both `run` and `exec` commands, with bug fixes and proper test coverage**,
so that **I don't need to memorize provider names and can quickly select a provider in either command without encountering terminal corruption, race conditions, or misleading errors**.

## Acceptance Criteria

1. **AC1: exec TUI integration works end-to-end** — `ccctx exec` (no args) shows TUI selector, user selects provider, launches `$SHELL` with provider env vars. `ccctx exec` (no args, no contexts) outputs `Error: no contexts found` to stderr, exits 1. FR19 satisfied. [Source: epics.md#Story 2.1]

2. **AC2: TUI shared between run and exec** — Both commands call the same `internal/ui/selector.go` implementation. TUI calls only exist in `cmd/*.go`, never in runner package. FR23 satisfied. [Source: architecture.md#Boundary Rules]

3. **AC3: TUI selector signature refactored** — `RunContextSelector()` changed to accept `contexts []string` parameter, eliminating the redundant `config.ListContexts()` call inside the selector. Command files pass contexts directly. [Source: deferred-work.md — "Redundant config.ListContexts() call in TUI path"]

4. **AC4: Sentinel error for TUI cancellation** — Replace fragile `err.Error() == "operation cancelled"` string comparison with a sentinel error `var ErrCancelled = errors.New("operation cancelled")` in the `ui` package. Both `cmd/run.go` and `cmd/exec.go` use `errors.Is(err, ui.ErrCancelled)`. [Source: deferred-work.md — "Fragile string comparison for TUI cancellation detection"]

5. **AC5: Race condition on cancelled flag fixed** — The `cancelled` bool in `runTviewSelector` is set in a callback and read after `app.Stop()` — potential race condition. Eliminate by using the sentinel error return pattern instead of the mutable `cancelled` flag. [Source: deferred-work.md — "Race condition on TUI cancelled flag"]

6. **AC6: Panic recovery in TUI selector** — Wrap `app.SetRoot(flex, false).SetFocus(list).Run()` with `defer` + `recover()` to restore terminal state if a panic occurs inside tview. [Source: deferred-work.md — "No panic recovery in TUI selector may leave terminal corrupted"]

7. **AC7: Misleading bounds check error fixed** — `runTviewSelector` returns `"no context selected"` when `selectedContext == ""` after the app stops. This can never happen in practice (ESC sets cancelled=true, Enter/click sets selectedContext). Remove or clarify the dead-code branch. [Source: deferred-work.md — "TUI list selection bounds check returns misleading error"]

8. **AC8: ESC cancellation exits with code 1 (verify no regression)** — `ccctx exec` or `ccctx run` → ESC → outputs `Operation cancelled.` to stdout, exits with code 1. Already implemented in Epic 1 — verify no regression. [Source: architecture.md#Breaking Changes]

9. **AC9: `ccctx run` TUI unchanged (no regression)** — `ccctx run` (no args) TUI selector works identically to current behavior (except exit code change from 0→1 on cancel). [Source: epics.md#Story 2.1]

10. **AC10: Automated tests for TUI selector** — New `internal/ui/selector_test.go` with table-driven tests covering: sentinel error matching (`errors.Is(err, ErrCancelled)`), empty contexts error. Full TUI rendering tests (tview `app.Run()`) are out of scope — focus on testable error paths. Tests use testify. [Source: retro — "Write automated tests for TUI selector"]

11. **AC11: All tests pass** — `go test ./...` passes, `go vet ./...` clean, `make build` succeeds.

## Tasks / Subtasks

- [x] Task 1: Refactor TUI selector signature (AC: #3)
  - [x] Change `RunContextSelector()` to `RunContextSelector(contexts []string)` in `internal/ui/selector.go`
  - [x] Remove internal `config.ListContexts()` call — callers now pass contexts directly
  - [x] Keep `len(contexts) == 0` validation as defensive input parameter checking (callers also check, but selector guards its own contract)
  - [x] Remove `"github.com/dsdashun/ccctx/config"` import from `internal/ui/selector.go`
  - [x] Update `cmd/run.go`: pass `contexts` to `ui.RunContextSelector(contexts)` (already has `contexts` from `config.ListContexts()`)
  - [x] Update `cmd/exec.go`: pass `contexts` to `ui.RunContextSelector(contexts)` (already has `contexts` from `config.ListContexts()`)

- [x] Task 2: Write failing tests for sentinel error (AC: #10) [RED phase]
  - [x] Create `internal/ui/selector_test.go`
  - [x] Test: `ErrCancelled` is defined and matches `errors.Is(err, ErrCancelled)` — will fail because `ErrCancelled` doesn't exist yet
  - [x] Test: `RunContextSelector([]string{})` returns error containing "no contexts found" — validates defensive input check
  - [x] Full TUI rendering tests (tview `app.Run()`) are out of scope — focus on testable error paths only
  - [x] Note: The actual `runTviewSelector` cancellation return path (ESC → `selectedContext == ""` → returns `ErrCancelled`) is not directly testable without mocking tview's `Application.Run()`. The sentinel error test verifies the contract (`errors.Is` works); the return logic is trivially correct (3-line if/else). If a future refactor breaks this path, the manual verification in Task 5 will catch it.
  - [x] Use testify (`require`/`assert`), table-driven pattern

- [x] Task 3: Add sentinel error and fix cancellation detection (AC: #4, #5, #7) [GREEN phase]
  - [x] Add `var ErrCancelled = errors.New("operation cancelled")` in `internal/ui/selector.go`
  - [x] Add `"errors"` import to `cmd/run.go` and `cmd/exec.go`
  - [x] Replace `fmt.Errorf("operation cancelled")` with `ErrCancelled` in `runTviewSelector`
  - [x] Remove `cancelled` bool variable — ESC handler calls `app.Stop()` directly, post-Stop check: if `selectedContext == ""` return `ErrCancelled`
  - [x] All 3 exit paths covered by `selectedContext == ""` check: ESC (stays `""`), SetDoneFunc (sets context from current item directly), SetSelectedFunc (sets context from `mainText`)
  - [x] Remove bounds check in `SetDoneFunc` — replace `if currentIndex >= 0 && currentIndex < len(contexts) { selectedContext = contexts[currentIndex] }` with `selectedContext = contexts[list.GetCurrentItem()]`. Safe because `len(contexts) > 0` is guaranteed by both callers and selector's own defensive check, and `GetCurrentItem()` returns a valid index when items exist. Removing it makes the claim "`selectedContext == ""` always means cancellation" strictly true.
  - [x] Remove `"no context selected"` dead-code branch (AC: #7) — `selectedContext == ""` now always means cancellation, never an unreachable "no selection" case
  - [x] Update `cmd/run.go`: replace `err.Error() == "operation cancelled"` with `errors.Is(err, ui.ErrCancelled)`
  - [x] Update `cmd/exec.go`: replace `err.Error() == "operation cancelled"` with `errors.Is(err, ui.ErrCancelled)`
  - [x] Run tests from Task 2 — all should now pass

- [x] Task 4: Add panic recovery in TUI selector (AC: #6)
  - [x] Change `runTviewSelector` signature to use named return values: `(result string, err error)` — required for defer/recover to modify returns
  - [x] Add deferred recovery with nil check: `if app != nil { app.Stop() }` to restore terminal state, then sets `err = fmt.Errorf("TUI error: %v", r)`
  - [x] The nil check guards against panics during or before `tview.NewApplication()` when `app` would still be nil

- [x] Task 5: Verify end-to-end behavior (AC: #1, #2, #8, #9, #11)
  - [x] `go test ./...` passes
  - [x] `go vet ./...` clean
  - [x] `make build` succeeds
  - [x] Manual: `./ccctx exec` shows TUI, select provider, lands in shell with env vars
  - [x] Manual: `./ccctx run` shows TUI (regression check)
  - [x] Manual: ESC cancels with exit code 1 in both commands

## Dev Notes

### Current State

The basic TUI integration for `exec` is already implemented in `cmd/exec.go:25-44`. Both `run` and `exec` call `ui.RunContextSelector()` when `ParseArgs` returns `useTUI=true`. The functional behavior is working.

This story focuses on **quality improvements** identified during Epic 1 code reviews and the retrospective:
- TUI selector refactoring (signature change)
- Bug fixes (sentinel error, race condition, panic recovery, dead code)
- Automated test coverage

### Sentinel Error Pattern

```go
// internal/ui/selector.go
package ui

import "errors"

var ErrCancelled = errors.New("operation cancelled")

func RunContextSelector(contexts []string) (string, error) {
    return runTviewSelector(contexts)
}
```

```go
// cmd/exec.go or cmd/run.go
if errors.Is(err, ui.ErrCancelled) {
    fmt.Println("Operation cancelled.")
    os.Exit(1)
}
```

### Race Condition Fix Strategy

The current code uses a mutable `cancelled` bool set in the ESC callback and read after `app.Stop()`. Fix by eliminating the flag entirely:

```go
// BEFORE:
var selectedContext string
var cancelled bool

// ESC handler:
cancelled = true
app.Stop()

// AFTER (no cancelled flag, no dead-code branch, no bounds check):
var selectedContext string

// ESC handler:
app.Stop()  // selectedContext stays ""

// SetDoneFunc (Enter/Tab) — no bounds check, direct access:
selectedContext = contexts[list.GetCurrentItem()]
app.Stop()

// Post-Stop (replaces both cancelled check and "no context selected" branch):
if selectedContext == "" {
    return "", ErrCancelled
}
```

This works because all 3 exit paths are covered:
- ESC → `app.Stop()` → `selectedContext` remains `""` → return `ErrCancelled`
- `SetDoneFunc` (Enter/Tab) → `selectedContext = contexts[list.GetCurrentItem()]` (no bounds check — safe given `len(contexts) > 0` guarantee) → `app.Stop()` → return selected context
- `SetSelectedFunc` (click) → `selectedContext = mainText` → `app.Stop()` → return selected context

### Panic Recovery Pattern

```go
func runTviewSelector(contexts []string) (result string, err error) {
    defer func() {
        if r := recover(); r != nil {
            if app != nil {
                app.Stop()  // restore terminal state before returning
            }
            result = ""
            err = fmt.Errorf("TUI error: %v", r)
        }
    }()
    // ... existing tview setup and app.Run()
}
```

Note: Named return values `(result string, err error)` are required — the deferred function can only modify named returns. The `app != nil` check guards against panics during or before `tview.NewApplication()`.

### Signature Change Impact

Current call chain:
1. `cmd/*.go` calls `config.ListContexts()` → checks empty → calls `ui.RunContextSelector()`
2. `ui.RunContextSelector()` calls `config.ListContexts()` again internally → redundant

New call chain:
1. `cmd/*.go` calls `config.ListContexts()` → checks empty → passes contexts to `ui.RunContextSelector(contexts)`
2. `ui.RunContextSelector(contexts)` validates `len(contexts) == 0` as defensive check → uses passed contexts directly

This also means `internal/ui/selector.go` no longer imports `config/` package — cleaner dependency graph.

### Previous Story Intelligence

From Epic 1 retrospective:
- TUI-related issues surfaced repeatedly in Stories 1.1, 1.2, and 1.3 code reviews but were deferred to Epic 2
- Deferred-work.md has 5 distinct TUI-specific issues (6 deferred items, fragile string comparison appears twice from Stories 1.1 and 1.2 reviews)
- Team agreement: "TUI-related issues concentrated in Epic 2 for resolution — no further individual defers"

From Story 1-3 deferred items:
- Fragile string comparison for TUI cancellation — sentinel error fixes this
- Redundant `config.ListContexts()` — signature change fixes this
- No panic recovery — explicit fix in this story
- Race condition on cancelled flag — eliminated by removing the flag
- Misleading bounds check error — dead code removal

### Project Structure Notes

- Modified: `internal/ui/selector.go` — signature change, sentinel error, panic recovery, race fix, dead code removal (AC7 merged into sentinel error task)
- New: `internal/ui/selector_test.go` — automated tests
- Modified: `cmd/exec.go` — updated TUI call (pass contexts, use sentinel error)
- Modified: `cmd/run.go` — updated TUI call (pass contexts, use sentinel error)
- No changes to: `internal/runner/`, `config/`, `main.go`
- `runner.Run` double-print issue for non-ExitError failures remains out of scope (deferred, not TUI-specific)

### References

- [Source: internal/ui/selector.go] — current TUI implementation to refactor
- [Source: cmd/exec.go:25-43] — exec TUI integration (already working)
- [Source: cmd/run.go:26-46] — run TUI integration (already working)
- [Source: _bmad-output/implementation-artifacts/deferred-work.md] — 5 TUI-specific deferred items
- [Source: _bmad-output/implementation-artifacts/epic-1-retro-2026-04-19.md#Technical Debt] — TUI bug fixes targeted for Epic 2
- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.1] — acceptance criteria
- [Source: _bmad-output/planning-artifacts/architecture.md#Boundary Rules] — TUI calls only in cmd/*.go
- [Source: _bmad-output/project-context.md#TUI (tview)] — "ESC cancels and returns operation cancelled error"
- [Source: _bmad-output/project-context.md#Critical Don't-Miss Rules] — "TUI cancellation exits with code 1, not 0"

## Dev Agent Record

### Agent Model Used

Claude Opus 4.7 (claude-opus-4-7)

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- Refactored `RunContextSelector()` to accept `contexts []string` parameter, removing redundant `config.ListContexts()` call and `config` import from ui package
- Added `ErrCancelled` sentinel error replacing fragile string comparison in both `cmd/run.go` and `cmd/exec.go`
- Eliminated `cancelled` bool race condition — ESC handler now simply calls `app.Stop()`, post-Stop check uses `selectedContext == ""` to determine cancellation
- Removed dead-code `"no context selected"` branch (AC7) and unnecessary bounds check in `SetDoneFunc`
- Added panic recovery in `runTviewSelector` with named returns and `app.Stop()` in defer to restore terminal state
- Created `internal/ui/selector_test.go` with table-driven tests for sentinel error matching and empty contexts error
- Updated `min()` usage replacing manual `if maxItems > 10` pattern (Go 1.21+ builtin)
- All 19 tests pass, go vet clean, make build succeeds

### File List

- internal/ui/selector.go (modified)
- internal/ui/selector_test.go (new)
- cmd/exec.go (modified)
- cmd/run.go (modified)

## Change Log

- 2026-04-19: Implemented TUI quality improvements — signature refactor, sentinel error, race condition fix, panic recovery, dead code removal, automated tests

### Review Findings

#### decision-needed
(None)

#### patch
(None)

#### defer
- [x] [Review][Defer] Terminal may be left in raw mode if app.Run() returns error [internal/ui/selector.go:91] — deferred, pre-existing
- [x] [Review][Defer] Hardcoded flex width 50 truncates long context names [internal/ui/selector.go:53] — deferred, pre-existing
- [x] [Review][Defer] Magic number maxItems+4 in SetRect lacks explanation [internal/ui/selector.go:53] — deferred, pre-existing
- [x] [Review][Defer] Divergent selection sources between SetDoneFunc (contexts[index]) and SetSelectedFunc (mainText) [internal/ui/selector.go:82,87] — deferred, pre-existing
- [x] [Review][Defer] Cancellation message goes to stdout instead of stderr (inconsistent with other errors) [cmd/exec.go:39, cmd/run.go:42] — deferred, pre-existing
- [x] [Review][Defer] Panic recovery converts panic to plain error, losing stack trace [internal/ui/selector.go:24-32] — deferred, pre-existing
