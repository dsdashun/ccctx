---
tested_plan: _bmad-output/implementation-artifacts/3-1-runner-and-command-layer-robustness.md
tested_stories: _bmad-output/implementation-artifacts/3-1-runner-and-command-layer-robustness.md
tester: kimi
latest_round: 1
test_rounds:
  - round: 1
    date: 2026-04-20
---

# User Story Test: 3.1 — Runner & Command Layer Robustness

Tested plan: `_bmad-output/implementation-artifacts/3-1-runner-and-command-layer-robustness.md`
Tested stories: Same file (Story + Acceptance Criteria embedded in plan)

## Round 1

**Context:** Initial simulation of all acceptance criteria against the design and actual codebase. The plan has already undergone 3 rounds of code review; this test verifies behavioral coverage from a user perspective.

---

### Story: AC1 — BaseURL format validation in runner.New()

**User story:** `New()` validates that `BaseURL` is a valid URL (has scheme, no spaces). Invalid URLs return a clear validation error before creating a Runner.

**Simulated input:** User configures a context with an invalid base_url and runs `ccctx run badurl`.

```toml
[context.badurl]
base_url = "not a valid url"
auth_token = "token123"
```

**Expected behavior:** The tool returns a clear validation error about the invalid BaseURL format *before* attempting to launch claude.

**Traced behavior:**
1. `cmd/run.go` calls `runner.ParseArgs(["badurl"])` -> provider="badurl", targetArgs=[], useTUI=false
2. `exec.LookPath("claude")` succeeds
3. `runner.New(Options{ContextName: "badurl", Target: [claudePath]})` is called
4. `config.GetContext("badurl")` resolves the context
5. Existing checks pass: BaseURL non-empty (line 31), AuthToken non-empty (line 34)
6. **New logic (Task 1):** `validateURL("not a valid url")` is called
   - `strings.Contains(rawURL, " ")` -> true, or `url.Parse()` then scheme check catches it
   - Returns error: `"invalid base_url: contains spaces"` or `"invalid base_url: missing scheme"`
7. `New()` returns the error upward
8. `cmd/run.go` prints `"Error: invalid base_url: ..."` to stderr and exits with code 1

**Result:** PASS — The design specifies exact validation logic (`net/url.Parse` + scheme check + space check), insertion point (after AuthToken check, before Target check), and error messages. Task 1 follows Red-Green-Refactor with explicit test cases.

---

### Story: AC2 — No double-printing of non-ExitError failures

**User story:** When `runner.Run()` returns a non-ExitError (e.g., binary not found), the error message is printed exactly once by the caller. The runner itself never prints to stderr/stdout.

**Simulated input:** User runs `ccctx run mycontext` but the claude binary has been removed or has permission issues after `LookPath` succeeded (race condition or filesystem change).

**Expected behavior:** The error message appears exactly once on stderr.

**Traced behavior:**
1. `runner.New()` succeeds (context resolved, URL valid)
2. `r.Run()` calls `exec.Command(claudePath, targetArgs...).Run()`
3. `cmd.Run()` fails with a non-ExitError (e.g., `"permission denied"`, `"text file busy"`)
4. `errors.As(err, &exitErr)` -> false (not an ExitError)
5. `Run()` returns `(1, err)` — **no printing occurs inside runner**
6. `cmd/run.go` receives `err != nil`, executes `fmt.Fprintf(os.Stderr, "Error: %v\n", err)` — **exactly one print**
7. `os.Exit(1)`

**Result:** PASS — Current `runner.Run()` (lines 44-60) already returns `(int, error)` without any I/O. Task 2 verifies this contract with an audit + documentation comment. Both `cmd/run.go` and `cmd/exec.go` have exactly one error-print path per `r.Run()` call. No double-printing risk exists.

---

### Story: AC3 — ParseArgs rejects flag-like args in provider position

**User story:** When `ParseArgs` receives args starting with `-` in the provider position (before `--`), it rejects them with a clear error. This prevents flags from being treated as context names.

**Simulated input 1:** User mistakenly runs `ccctx run --model foo`

**Traced behavior:**
1. `runner.ParseArgs(["--model", "foo"])` — no `--` separator
2. Current behavior would return provider="--model", targetArgs=["foo"]
3. **New logic (Task 3):** `args[0] = "--model"` starts with `"-"` -> return error `"flag-like argument '--model' not allowed in provider position"`
4. `cmd/run.go` prints the error and exits

**Simulated input 2:** User runs `ccctx run --model -- foo`

**Traced behavior:**
1. `runner.ParseArgs(["--model", "--", "foo"])` — `--` at index 1
2. contextArgs = ["--model"], provider = "--model"
3. **New logic:** provider starts with `"-"` -> return same error
4. User gets clear error instead of confusing "context '--model' not found"

**Expected behavior:** Clear error message instead of confusing "context not found".

**Result:** PASS — Task 3 specifies exact rejection logic for both code paths (with and without `--` separator), with consistent error messages verified across the GREEN step and Dev Notes. Test cases in Task 3 RED cover all three scenarios.

---

### Story: AC4 — No direct os.Exit in cmd/exec.go error paths

**User story:** Error handling in `cmd/exec.go` uses `return` with exit code propagation instead of direct `os.Exit()`, ensuring any deferred cleanup functions execute. `os.Exit()` is called only at the top level of the `Run` function.

**Simulated input:** User runs `ccctx exec nonexistent` where "nonexistent" is not a configured context.

**Traced behavior (after refactor):**
1. `Run: func(cmd *cobra.Command, args []string) { os.Exit(execRun(args)) }` — single `os.Exit` at top level
2. `execRun(["nonexistent"])`:
   - `ParseArgs` succeeds -> provider="nonexistent"
   - `config.GetContext("nonexistent")` fails
   - `execRun` returns `1` (no `os.Exit` inside)
   - Any `defer` in `execRun` executes before return
3. `Run` receives `1`, calls `os.Exit(1)`

**Expected behavior:** Deferred functions in `execRun` execute before program exit.

**Result:** PASS — Task 4 specifies the exact refactor pattern: extract body into `execRun(args) int`, return codes from all error paths, single `os.Exit` in the `Run` closure. Same pattern applied to `cmd/run.go`. Currently no defers exist, but the pattern is future-proof for Story 3.2 cleanup.

---

### Story: AC5 — cmd/run.go table-driven tests

**User story:** New test file `cmd/run_test.go` with table-driven tests covering success path, context not found, TUI cancellation, and LookPath failure. Uses testify framework.

**Simulated input:** Developer runs `go test ./cmd/...`

**Traced behavior:**
1. `cmd/run.go` is refactored (Task 4) to extract `runRun(args []string) int`
2. New file `cmd/run_test.go` with `package cmd` (internal test, not external)
3. Tests call unexported `runRun` directly:
   - **Success path:** Temp config with valid context + temp PATH with fake `claude` binary -> expect exit code 0
   - **Context not found:** Temp config without the requested context -> expect exit code 1
   - **LookPath failure:** PATH without `claude` -> expect exit code 1
   - **ParseArgs error:** Args like `["a", "b", "c"]` (multiple before `--`) -> expect exit code 1
4. TUI cancel path is explicitly excluded from unit tests (requires interactive terminal)

**Expected behavior:** All test cases pass, providing regression coverage for the run command.

**Result:** PASS — Task 5 specifies exact test strategy: `package cmd` for unexported access, `CCCTX_CONFIG_PATH` for config isolation, temp PATH for LookPath control, testify framework, table-driven pattern. The note about TUI cancellation being excluded is appropriate and documented.

---

### Story: AC6 — All tests pass

**User story:** `go test ./...` passes, `go vet ./...` clean, `make build` succeeds.

**Simulated input:** Developer runs the verification commands after implementing all tasks.

**Traced behavior:**
1. Task 6 lists explicit verification steps
2. `go test ./...` — runs all existing tests + new `cmd/run_test.go` + updated `runner_test.go` + updated `args_test.go`
3. `go vet ./...` — static analysis clean
4. `make build` — binary builds successfully

**Expected behavior:** All verification commands succeed.

**Result:** PASS — Task 6 is a standard verification step. No new dependencies are introduced (URL validation uses stdlib `net/url`). The only new file is a test file, which doesn't affect the build.

---

### Story: AC7 — Existing behavior unchanged

**User story:** `ccctx run <provider>`, `ccctx exec <provider>`, `ccctx run` (TUI), `ccctx exec` (TUI) all work identically to current behavior.

**Simulated input 1:** `ccctx run myprovider` with a valid configuration

**Traced behavior:**
- ParseArgs: provider="myprovider" (no `-` prefix) -> passes flag check
- URL validation: valid URL with scheme -> passes
- All other logic unchanged

**Simulated input 2:** `ccctx run` (no args, TUI mode)

**Traced behavior:**
- ParseArgs: no args -> useTUI=true
- No provider to validate, no flag check triggered
- TUI flow identical to current

**Simulated input 3:** `ccctx exec myprovider -- python3 script.py`

**Traced behavior:**
- ParseArgs: provider="myprovider", targetArgs=["python3", "script.py"]
- Provider doesn't start with `-`, passes flag check
- exec flow unchanged

**Expected behavior:** All existing use cases continue to work exactly as before.

**Result:** PASS — The design's changes are all additive/structural:
- URL validation only triggers on invalid URLs (valid URLs pass through)
- ParseArgs flag check only triggers when provider starts with `-` (normal names unaffected)
- os.Exit refactor is purely structural (same behavior, different code organization)
- New tests don't affect runtime behavior

---

### Cross-Story Consistency Check

| Interaction | Verified | Notes |
|-------------|----------|-------|
| AC1 URL validation + AC3 flag rejection | OK | Both are input validation; they operate at different layers (config vs. CLI args) with no overlap |
| AC3 flag rejection + Story 3.2 `--model` flag | OK | Design explicitly notes Story 3.2 will change this behavior; AC3 establishes the pattern |
| AC4 os.Exit refactor + AC5 testability | OK | Refactor enables testing `runRun` without `os.Exit` killing the test process |
| AC5 tests + AC6 build | OK | Test file uses `package cmd` (internal test), doesn't affect `make build` |
| AC7 behavior preservation + all other ACs | OK | All changes are non-breaking for valid inputs |

---

### Design-to-Code Accuracy Verification

| Design Reference | Actual Code Location | Match |
|-----------------|---------------------|-------|
| `runner.go:31-36` — `ctx.AuthToken == ""` check | Line 34: `if ctx.AuthToken == ""` | **CORRECT** (minor line drift noted and fixed in review) |
| `runner.go:37` — `len(opts.Target) == 0` | Line 37: `if len(opts.Target) == 0` | **CORRECT** |
| `runner.go:44-60` — `Run()` returns `(int, error)` | Lines 44-60: three return paths | **CORRECT** |
| `args.go:30-34` — ParseArgs no-args path | Lines 30-34: `len(args) == 0` through `return args[0], args[1:]` | **CORRECT** |
| `cmd/exec.go:19-69` — exec with os.Exit calls | Lines 19-69: Run closure with multiple `os.Exit` | **CORRECT** |
| `cmd/run.go:20-74` — run with os.Exit calls | Lines 20-74: Run closure with multiple `os.Exit` | **CORRECT** |
| `config/config.go` respects `CCCTX_CONFIG_PATH` | Verified via grep: `os.Getenv("CCCTX_CONFIG_PATH")` | **CORRECT** |

---

### Issue Summary

No issues found. All acceptance criteria are covered by the design with clear implementation paths, accurate code references, and consistent specifications.

### Verdict

**All user stories are satisfied by the design.** No blocking, high, or medium-severity issues remain. The design has undergone prior review rounds that resolved TDD ordering, test feasibility, ambiguous descriptions, line number drift, and minor inconsistencies. The behavioral coverage for all 7 acceptance criteria is complete and correct.
