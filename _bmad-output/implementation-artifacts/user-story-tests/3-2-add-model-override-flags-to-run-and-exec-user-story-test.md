---
tested_plan: _bmad-output/implementation-artifacts/3-2-add-model-override-flags-to-run-and-exec.md
tested_stories: _bmad-output/implementation-artifacts/3-2-add-model-override-flags-to-run-and-exec.md
tester: glm
latest_round: 2
test_rounds:
  - round: 1
    date: 2026-04-23
  - round: 2
    date: 2026-04-23
---

# User Story Test: 3-2-add-model-override-flags-to-run-and-exec

Tested plan: `_bmad-output/implementation-artifacts/3-2-add-model-override-flags-to-run-and-exec.md`
Tested stories: `_bmad-output/implementation-artifacts/3-2-add-model-override-flags-to-run-and-exec.md` (embedded AC1-AC10)

## Round 1

**Context:** Initial behavioral simulation of all acceptance criteria (AC1-AC10) against the design. The plan has been through 3 review rounds with all blocking/high issues resolved. This test focuses on user-facing behavior, not plan quality.

### Story: AC1 — --model flag overrides config model

**User story:** Given provider-A has `model = "claude-sonnet-4-6"` in config, `ccctx run provider-A --model claude-opus-4-7` sets `ANTHROPIC_MODEL=claude-opus-4-7` in the child process.

**Simulated input:** User types `ccctx run provider-A --model claude-opus-4-7` in their terminal, where provider-A's config has `model = "claude-sonnet-4-6"`.

**Expected behavior:** The child `claude` process receives `ANTHROPIC_MODEL=claude-opus-4-7` (CLI flag value), not `claude-sonnet-4-6` (config value).

**Traced behavior:**
1. Cobra's `DisableFlagParsing: true` passes raw args `["provider-A", "--model", "claude-opus-4-7"]` to Run closure
2. `WantsHelp(args)` → false (no `--help`/`-h`)
3. `runRun(args)` called
4. `ExtractFlags(args)`: scans all args (no `--` separator), finds `--model` at index 1 with value `"claude-opus-4-7"` at index 2 → model="claude-opus-4-7", remaining=["provider-A"]
5. `ParseArgs(["provider-A"])` → provider="provider-A", targetArgs=[], useTUI=false
6. `LookPath("claude")` → finds claude binary
7. target = ["claude"]
8. `New(Options{ContextName: "provider-A", Target: ["claude"], Model: "claude-opus-4-7"})`
9. In `buildEnv(ctx, opts)`: opts.Model="claude-opus-4-7" (non-empty) → `model = "claude-opus-4-7"` → appends `ANTHROPIC_MODEL=claude-opus-4-7`
10. Child process launched with `ANTHROPIC_MODEL=claude-opus-4-7`

**Result:** PASS

---

### Story: AC2 — --model flag sets model when config has none

**User story:** Given provider-A has no `model` configured, `ccctx exec provider-A --model claude-opus-4-7 -- env | grep ANTHROPIC_MODEL` outputs `ANTHROPIC_MODEL=claude-opus-4-7`.

**Simulated input:** User types `ccctx exec provider-A --model claude-opus-4-7 -- env | grep ANTHROPIC_MODEL` in their terminal. The outer shell handles the `|` pipe operator, so ccctx receives args `["provider-A", "--model", "claude-opus-4-7", "--", "env"]`.

**Expected behavior:** `env` runs with `ANTHROPIC_MODEL=claude-opus-4-7` set in its environment, and `grep ANTHROPIC_MODEL` filters to show that line.

**Traced behavior:**
1. `DisableFlagParsing: true` → raw args to Run closure
2. `WantsHelp(args)` → false
3. `execRun(args)` called
4. `ExtractFlags(["provider-A", "--model", "claude-opus-4-7", "--", "env"])`: sepIdx=3, preSep=["provider-A", "--model", "claude-opus-4-7"]. Scans pre-sep: index 0 → "provider-A" (remaining), index 1 → `--model` with value "claude-opus-4-7". model="claude-opus-4-7", remaining=["provider-A", "--", "env"]
5. `ParseArgs(["provider-A", "--", "env"])`: separatorIndex=1, provider="provider-A", targetArgs=["env"], useTUI=false
6. targetArgs=["env"] (non-empty) → no $SHELL fallback
7. `New(Options{ContextName: "provider-A", Target: ["env"], Model: "claude-opus-4-7"})`
8. In `buildEnv(ctx, opts)`: ctx.Model="" (not configured), opts.Model="claude-opus-4-7" → model="claude-opus-4-7" → `ANTHROPIC_MODEL=claude-opus-4-7`
9. `exec.Command("env")` launched with ANTHROPIC_MODEL=claude-opus-4-7 in env
10. `env` prints all env vars, outer shell pipes to `grep ANTHROPIC_MODEL` → outputs `ANTHROPIC_MODEL=claude-opus-4-7`

**Result:** PASS

---

### Story: AC3 — --small-fast-model flag works

**User story:** Given provider-A has `small_fast_model = "claude-haiku-4-5"`, `ccctx run provider-A --small-fast-model claude-sonnet-4-6` sets `ANTHROPIC_SMALL_FAST_MODEL=claude-sonnet-4-6`.

**Simulated input:** User types `ccctx run provider-A --small-fast-model claude-sonnet-4-6`.

**Expected behavior:** Child process receives `ANTHROPIC_SMALL_FAST_MODEL=claude-sonnet-4-6` (CLI flag), not `claude-haiku-4-5` (config).

**Traced behavior:**
1. `ExtractFlags(["provider-A", "--small-fast-model", "claude-sonnet-4-6"])`: finds `--small-fast-model` at index 1, value "claude-sonnet-4-6" → smallFastModel="claude-sonnet-4-6", remaining=["provider-A"]
2. `ParseArgs(["provider-A"])` → provider="provider-A"
3. `buildEnv(ctx, opts)`: opts.SmallFastModel="claude-sonnet-4-6" (non-empty) → sfm="claude-sonnet-4-6" → `ANTHROPIC_SMALL_FAST_MODEL=claude-sonnet-4-6`

**Result:** PASS

---

### Story: AC4 — CLI flags take priority over config

**User story:** When both CLI flag and config value exist, CLI flag wins. Priority chain: `Options.Model > ctx.Model > omit`.

**Simulated input:** User types `ccctx run provider-A --model claude-opus-4-7 --small-fast-model claude-sonnet-4-6`, where provider-A config has `model = "claude-sonnet-4-6"` and `small_fast_model = "claude-haiku-4-5"`.

**Expected behavior:** `ANTHROPIC_MODEL=claude-opus-4-7` and `ANTHROPIC_SMALL_FAST_MODEL=claude-sonnet-4-6` (both CLI values win).

**Traced behavior:**
1. `ExtractFlags`: model="claude-opus-4-7", smallFastModel="claude-sonnet-4-6", remaining=["provider-A"]
2. `buildEnv(ctx, opts)`:
   - model = opts.Model = "claude-opus-4-7" (non-empty, skips ctx.Model="claude-sonnet-4-6") → `ANTHROPIC_MODEL=claude-opus-4-7`
   - sfm = opts.SmallFastModel = "claude-sonnet-4-6" (non-empty, skips ctx.SmallFastModel="claude-haiku-4-5") → `ANTHROPIC_SMALL_FAST_MODEL=claude-sonnet-4-6`

**Additional traces — all priority combinations:**

| opts.Model | ctx.Model | Result |
|------------|-----------|--------|
| "opus" | "sonnet" | ANTHROPIC_MODEL=opus |
| "opus" | "" | ANTHROPIC_MODEL=opus |
| "" | "sonnet" | ANTHROPIC_MODEL=sonnet |
| "" | "" | (no ANTHROPIC_MODEL injected) |

Same pattern verified for SmallFastModel.

**Result:** PASS — all priority combinations produce correct behavior.

---

### Story: AC5 — No flags = config values used

**User story:** When no flags specified, config file values are used (existing behavior unchanged).

**Simulated input:** User types `ccctx run provider-A`, where provider-A has `model = "claude-sonnet-4-6"` and `small_fast_model = "claude-haiku-4-5"`.

**Expected behavior:** Same behavior as before this story was implemented — config values are used.

**Traced behavior:**
1. `ExtractFlags(["provider-A"])`: no `--model`/`--small-fast-model` found → model="", smallFastModel="", remaining=["provider-A"]
2. `ParseArgs(["provider-A"])` → provider="provider-A"
3. `buildEnv(ctx, opts)`: opts.Model="" → falls through to ctx.Model="claude-sonnet-4-6" → `ANTHROPIC_MODEL=claude-sonnet-4-6`. Same for SmallFastModel.

**Result:** PASS — existing behavior preserved. Also verified that existing tests `TestBuildEnv`, `TestBuildEnv_SkipsEmptyOptional`, `TestBuildEnv_InjectionOrder` continue to pass because they use `Options{}` (zero values), which triggers the ctx fallback path.

---

### Story: AC6 — Flags available on both commands

**User story:** `--model` and `--small-fast-model` work identically on both `run` and `exec`.

**Simulated input:** User runs both commands with model flags:
- `ccctx run provider-A --model claude-opus-4-7`
- `ccctx exec provider-A --model claude-opus-4-7`

**Expected behavior:** Both commands extract the flag and set `ANTHROPIC_MODEL=claude-opus-4-7`.

**Traced behavior:**

**run command:**
1. RunCmd has `DisableFlagParsing: true`
2. Run closure checks `WantsHelp(args)` → false
3. `runRun(args)` calls `ExtractFlags` → `ParseArgs` → `New(Options{Model: "claude-opus-4-7", ...})`
4. `buildEnv` uses opts.Model → `ANTHROPIC_MODEL=claude-opus-4-7`

**exec command:**
1. ExecCmd has `DisableFlagParsing: true`
2. Run closure checks `WantsHelp(args)` → false
3. `execRun(args)` calls `ExtractFlags` → `ParseArgs` → `New(Options{Model: "claude-opus-4-7", ...})`
4. `buildEnv` uses opts.Model → `ANTHROPIC_MODEL=claude-opus-4-7`

Both follow the identical pattern: ExtractFlags → ParseArgs → Options → buildEnv. The flag handling is shared via `runner.ExtractFlags` (not duplicated per command).

**Result:** PASS

---

### Story: AC7 — --model with TUI mode

**User story:** `ccctx run --model claude-opus-4-7` (no provider) opens TUI selector, and the selected provider runs with the overridden model.

**Simulated input:** User types `ccctx run --model claude-opus-4-7` (no provider specified).

**Expected behavior:** TUI selector appears, user selects a provider, claude runs with `ANTHROPIC_MODEL=claude-opus-4-7`.

**Traced behavior:**
1. `WantsHelp(["--model", "claude-opus-4-7"])` → false (no `--help`/`-h`)
2. `runRun(["--model", "claude-opus-4-7"])`
3. `ExtractFlags(["--model", "claude-opus-4-7"])`: model="claude-opus-4-7", remaining=[]
4. `ParseArgs([])` → provider="", targetArgs=[], useTUI=true
5. TUI flow: `config.ListContexts()` → `ui.RunContextSelector(contexts)` → user selects "provider-A"
6. `LookPath("claude")` → finds claude binary
7. `New(Options{ContextName: "provider-A", Target: ["claude"], Model: "claude-opus-4-7"})`
8. `buildEnv(ctx, opts)`: opts.Model="claude-opus-4-7" → `ANTHROPIC_MODEL=claude-opus-4-7`

Key insight: `model` is captured before `ParseArgs` and passed to `Options` after TUI selection. The flag value flows through the TUI path correctly because `ExtractFlags` runs at the top of `runRun`, before the TUI branch.

**Result:** PASS

---

### Story: AC8 — -- separator not consumed by flag parsing

**User story:** `ccctx exec provider-A --model claude-opus-4-7 -- env | grep ANTHROPIC_MODEL` correctly forwards `env | grep ANTHROPIC_MODEL` as the target command.

**Simulated input:** User types `ccctx exec provider-A --model claude-opus-4-7 -- env | grep ANTHROPIC_MODEL`. The outer shell handles piping, so ccctx receives args `["provider-A", "--model", "claude-opus-4-7", "--", "env"]`.

**Expected behavior:** `--model` is extracted, `--` separator is preserved for `ParseArgs`, and `env` becomes the target command.

**Traced behavior:**
1. `ExtractFlags(["provider-A", "--model", "claude-opus-4-7", "--", "env"])`:
   - sepIdx = 3 (the `--` at index 3)
   - preSep = ["provider-A", "--model", "claude-opus-4-7"]
   - Scan pre-sep: "provider-A" → remaining; `--model` at index 1 → extract with value "claude-opus-4-7"
   - remaining = ["provider-A"]; post-separator = ["--", "env"]
   - Full remaining = ["provider-A", "--", "env"]; model = "claude-opus-4-7"
2. `ParseArgs(["provider-A", "--", "env"])`:
   - separatorIndex = 1 (the `--`)
   - provider = "provider-A", targetArgs = ["env"]
3. `exec.Command("env")` launched with correct env vars

**Additional edge case — flag after `--`:**
`ccctx run provider-A -- --model foo`:
- ExtractFlags: sepIdx=1, preSep=["provider-A"]. No flags in pre-sep.
- remaining = ["provider-A", "--", "--model", "foo"]
- ParseArgs: provider="provider-A", targetArgs=["--model", "foo"]
- `--model foo` forwarded to claude, not extracted.

**Result:** PASS — `--` separator handling is correct in both directions.

---

### Story: AC9 — All existing tests pass

**User story:** `go test ./...` passes, `go vet ./...` clean, `make build` succeeds.

**Simulated input:** Developer runs `go test ./...` after implementing all tasks.

**Expected behavior:** All tests pass without modification (except the one explicitly updated test).

**Traced behavior:** Traced through every existing test case:

**internal/runner/args_test.go — TestParseArgs:** ParseArgs is unchanged. All 11 cases pass. ✓

**internal/runner/runner_test.go:**
- `TestBuildEnv`: Uses `Options{}` (zero Model/SmallFastModel). After priority chain: opts.Model="" → falls through to ctx.Model="claude-sonnet-4-6". Same result as current code. ✓
- `TestBuildEnv_SkipsEmptyOptional`: ctx.Model="" and ctx.SmallFastModel="". opts.Model="" and opts.SmallFastModel="". Priority chain: both empty → no injection. Same result. ✓
- `TestBuildEnv_InjectionOrder`: Priority chain preserves BASE_URL → AUTH_TOKEN → MODEL → SMALL_FAST_MODEL order. ✓
- `TestValidateURL`: Unchanged. ✓

**cmd/run_test.go — TestRunRun:**
- "ParseArgs error - flag-like arg": Updated in Task 5 to use `--unknown-flag foo`. ExtractFlags doesn't extract it → ParseArgs rejects it. ✓
- "context not found": `ExtractFlags(["nonexistent"])` → remaining=["nonexistent"] → ParseArgs → provider="nonexistent" → GetContext fails. ✓
- "claude not found in PATH": `ExtractFlags(["test"])` → remaining=["test"] → ParseArgs → provider="test" → LookPath fails. ✓
- "success - provider found and claude found": `ExtractFlags(["test"])` → remaining=["test"] → ParseArgs → provider="test" → LookPath finds mock → success. ✓
- "success - claude exits with non-zero code": Same pattern as above. ✓
- "missing base_url", "invalid base_url", "missing auth_token": All work through the same ExtractFlags → ParseArgs → GetContext path. ✓
- "success - provider with forwarded args after separator": `ExtractFlags(["test", "--", "--model", "foo"])` → sepIdx=1, preSep=["test"], no flags extracted → remaining=["test", "--", "--model", "foo"] → ParseArgs → provider="test", targetArgs=["--model", "foo"]. ✓

All existing tests pass. ✓

**Result:** PASS

---

### Story: AC10 — New table-driven tests

**User story:** Tests cover flag extraction, priority chain, and integration with both commands.

**Simulated input:** Developer runs the new tests.

**Expected behavior:** Tests cover ExtractFlags (11 cases), WantsHelp (6 cases), buildEnv priority (7+ cases), run integration (4 cases), exec integration (3 cases).

**Traced behavior:** Verified coverage against acceptance criteria:

| Test Location | What It Tests | Coverage |
|---------------|---------------|----------|
| args_test.go: ExtractFlags | 11 cases including no flags, both flags, separator, errors, edge cases | AC8, AC10 |
| args_test.go: WantsHelp | 6 cases: --help, -h, after --, empty, with provider | Prevents --help regression |
| runner_test.go: buildEnv | Priority chain for Model and SmallFastModel, all 4 combinations each | AC1-AC5, AC10 |
| run_test.go: integration | --model with provider, --model with -- separator, --model without value, both flags | AC6, AC7, AC8, AC10 |
| exec_test.go: integration | --model with provider (shell mock), --model without value, --model with explicit target | AC6, AC10 |

**Result:** PASS — comprehensive test coverage for all acceptance criteria.

---

### Issue Summary

| # | Severity | Story | Issue |
|---|----------|-------|-------|
| 1 | MEDIUM | AC10 (Task 6) | Exec mock setup omits SHELL env var for no-target-args tests |

### Issues

#### 1. [MEDIUM] Exec mock setup for no-target-args tests doesn't mention setting SHELL

- **User story expects:** AC10 tests verify env var propagation end-to-end through the actual `exec.Command` path.
- **Design says:** "Env-Capturing Mock Pattern" Dev Notes step 4: "Set `PATH` to temp dir containing the mock. Mock naming rules: ...exec tests with no target args: mock named as `$SHELL` basename (e.g., `bash`) — exec falls back to `$SHELL`."
- **Gap/Mismatch/Ambiguity:** The instructions say to name the mock after `$SHELL`'s basename and add it to PATH, but don't mention setting `SHELL` itself. Without setting `SHELL`, the exec code does `shell := os.Getenv("SHELL")` → gets the real shell path (e.g., `/bin/bash`), then `exec.Command("/bin/bash")` runs the real shell, not the mock. PATH lookup is never triggered because the target contains a path separator.
- **Suggested fix:** Add to step 4: "For exec tests with no target args, also set `SHELL` to the mock's full path: `t.Setenv("SHELL", filepath.Join(tmpDir, "bash"))`." Alternatively, set `SHELL` to just `"bash"` (no path separator) so `exec.Command("bash")` does PATH lookup — both approaches work, but the full path approach is more representative of production behavior.

---

### Verification Summary

| AC | Description | Result | Notes |
|----|-------------|--------|-------|
| AC1 | --model overrides config model | PASS | Priority chain: opts.Model > ctx.Model |
| AC2 | --model sets model when config has none | PASS | -- separator preserved for ParseArgs |
| AC3 | --small-fast-model works | PASS | Same mechanism as --model |
| AC4 | CLI flags take priority over config | PASS | All 4 priority combinations verified |
| AC5 | No flags = config values used | PASS | Existing behavior preserved, existing tests unaffected |
| AC6 | Flags on both commands | PASS | Shared ExtractFlags, identical pattern |
| AC7 | --model with TUI mode | PASS | Flag captured before ParseArgs, passed after TUI selection |
| AC8 | -- separator not consumed | PASS | ExtractFlags scans only before --, preserves separator |
| AC9 | All existing tests pass | PASS | Every existing test traced through updated code |
| AC10 | New table-driven tests | PASS (with 1 MEDIUM issue) | Mock setup gap for exec no-target-args tests |

### Verdict

All 10 user stories are satisfied by the design. 9 of 10 pass cleanly. One medium-severity issue: the exec mock setup instructions for no-target-args tests don't mention setting the `SHELL` env var, which is required for the mock to actually be invoked (without it, `exec.Command` runs the real shell via the full `$SHELL` path, bypassing the mock). This is not a design flaw — the feature behavior is correct — but the test implementation instructions have a gap that the executor will need to resolve.

---

## Round 2

**Date:** 2026-04-23
**Context:** Response to Round 1. Addressing the single MEDIUM issue found during initial testing.

### Round 1 Issue Resolution

| # | Severity | Issue | Status | Evidence |
|---|----------|-------|--------|----------|
| 1 | MEDIUM | Exec mock setup for no-target-args tests doesn't mention setting `SHELL` env var | ✅ Fixed | Plan "Env-Capturing Mock Pattern" step 4 updated to include `t.Setenv("SHELL", filepath.Join(tmpDir, "bash"))` with explanation of why it's needed |

### Resolution Detail

**Issue #1 — SHELL env var for exec mock:**

The plan's "Env-Capturing Mock Pattern" (Dev Notes step 4) instructed the executor to name the mock after `$SHELL`'s basename (e.g., `bash`) and place it in PATH. However, without explicitly setting `SHELL` in the test environment, `os.Getenv("SHELL")` returns the real shell path (e.g., `/bin/bash`), which contains a path separator. Go's `exec.Command("/bin/bash")` resolves the full path directly — it never triggers PATH lookup, so the mock is bypassed entirely.

**Fix applied:** Added `t.Setenv("SHELL", filepath.Join(tmpDir, "bash"))` to the exec no-target-args mock setup instructions, with a note explaining the root cause: without setting `SHELL`, the real shell path bypasses the mock.

This is a test-instruction gap only — the feature code itself is correct.

### Updated Verdict

All 10 user stories pass cleanly. The single MEDIUM issue has been resolved in the plan. The design is ready for implementation.
