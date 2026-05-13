# Story 4.2: CLI Flag Extension & Backward Compatibility

Status: done

## Story

As a **user**,
I want **to override model selection via `--haiku-model`, `--sonnet-model`, and `--opus-model` CLI flags, with `--small-fast-model` continuing to work as an alias for `--haiku-model`**,
so that **I can temporarily switch models without editing the config file, using either the new canonical flags or the legacy alias for backward compatibility**.

## Acceptance Criteria

1. **AC1: `--haiku-model` flag works** — Given `ccctx run provider --haiku-model claude-haiku-4-5-20251001`, ExtractFlags sets `haikuModel = "claude-haiku-4-5-20251001"`, does not trigger TUI, and `ANTHROPIC_DEFAULT_HAIKU_MODEL=claude-haiku-4-5-20251001` is injected. [Source: epics.md#Story 4.2]

2. **AC2: `--sonnet-model` flag works** — Given `ccctx run provider --sonnet-model claude-sonnet-4-6`, ExtractFlags sets `sonnetModel = "claude-sonnet-4-6"`, and `ANTHROPIC_DEFAULT_SONNET_MODEL=claude-sonnet-4-6` is injected. [Source: epics.md#Story 4.2]

3. **AC3: `--opus-model` flag works** — Given `ccctx run provider --opus-model claude-opus-4-7`, ExtractFlags sets `opusModel = "claude-opus-4-7"`, and `ANTHROPIC_DEFAULT_OPUS_MODEL=claude-opus-4-7` is injected. [Source: epics.md#Story 4.2]

4. **AC4: `--small-fast-model` is alias for `--haiku-model`** — Given `ccctx run provider --small-fast-model claude-haiku-4-5-20251001`, behavior is identical to `--haiku-model claude-haiku-4-5-20251001`. Both set the haikuModel return value. [Source: architecture.md#Argument Parsing Patterns]

5. **AC5: `--haiku-model` wins over `--small-fast-model`** — Given both `--haiku-model X` and `--small-fast-model Y` specified, `haikuModel = "X"` (haiku-model takes priority, small-fast-model silently ignored). [Source: epics.md#Story 4.2]

6. **AC6: All new flags require a value** — `--haiku-model`, `--sonnet-model`, `--opus-model` without a following value return an error (same pattern as `--model`). [Source: architecture.md#Argument Parsing Patterns]

7. **AC7: All new flags validate against newlines** — Values containing `\n` are rejected with `"<flag> value cannot contain newline"` error (same pattern as `--model`). [Source: args.go:validateFlagValue]

8. **AC8: Flags only extracted before `--` separator** — `--haiku-model` after `--` is NOT extracted, passed through as a target argument. [Source: args.go:ExtractFlags design]

9. **AC9: `ccctx list` unaffected** — No change to list command output or behavior. [Source: epics.md#Story 4.2]

10. **AC10: No-flag commands unchanged** — `ccctx run provider` and `ccctx exec provider` without model flags behave identically to current behavior. [Source: epics.md#Story 4.2]

11. **AC11: Duplicate flags — last wins** — Same pattern as existing `--model`: `--haiku-model foo --haiku-model bar` → `haikuModel = "bar"`. [Source: args_test.go:TestExtractFlags "duplicate flags last wins"]

12. **AC12: All existing tests pass** — `go test ./...` passes, `go vet ./...` clean, `make build` succeeds. No regressions. [Source: project-context.md#Testing Rules]

13. **AC13: New table-driven tests** — Tests cover: new flag extraction, alias behavior (`--small-fast-model` → haikuModel), `--haiku-model` priority over `--small-fast-model`, all flag combinations, missing value errors, newline validation, flags after `--` not extracted. [Source: project-context.md#Testing Rules]

## Tasks / Subtasks

- [x] Task 1: RED — add failing tests for new flags in args_test.go (AC: #1-#8, #11, #13)
  - [x] Update `TestExtractFlags` test struct: rename `wantSmallFast` field to `wantHaikuModel`, add `wantSonnetModel` and `wantOpusModel` fields; update all existing test case values from `wantSmallFast: ...` to `wantHaikuModel: ...`; update call signature to unpack 6 return values
  - [x] Add test cases for each new flag: `--haiku-model`, `--sonnet-model`, `--opus-model` with provider
  - [x] Add test case: `--small-fast-model` sets `haikuModel` (alias verification)
  - [x] Add test case: both `--haiku-model` and `--small-fast-model` → `--haiku-model` wins
  - [x] Add test cases: new flags without value return error
  - [x] Add test cases: new flags with newline value return error
  - [x] Add test cases: new flags after `--` separator NOT extracted
  - [x] Add test cases: duplicate new flags — last wins
  - [x] Add test cases: all 5 flags combined

- [x] Task 2: GREEN — update ExtractFlags signature and implementation in args.go (AC: #1-#8, #11)
  - [x] Change signature from `(model, smallFastModel string, remaining []string, err error)` to `(model, haikuModel, sonnetModel, opusModel string, remaining []string, err error)`
  - [x] Add `case "--haiku-model"` branch: validate + set `haikuModel`
  - [x] Add `case "--sonnet-model"` branch: validate + set `sonnetModel`
  - [x] Add `case "--opus-model"` branch: validate + set `opusModel`
  - [x] Change `case "--small-fast-model"`: instead of setting `smallFastModel`, set `haikuModel` but only if `haikuModel` is still empty (so `--haiku-model` wins when both present)
  - [x] Remove `smallFastModel` return value entirely

- [x] Task 3: Update cmd/run.go and cmd/exec.go call sites (AC: #10)
  - [x] In `cmd/run.go:33`: change `model, smallFastModel, args, err := runner.ExtractFlags(args)` to `model, haikuModel, sonnetModel, opusModel, args, err := runner.ExtractFlags(args)`
  - [x] In `cmd/run.go:76-81`: update `runner.Options` construction to use `HaikuModel: haikuModel, SonnetModel: sonnetModel, OpusModel: opusModel` instead of `SmallFastModel: smallFastModel`
  - [x] In `cmd/exec.go:32`: same ExtractFlags signature change
  - [x] In `cmd/exec.go:74`: same Options construction change

- [x] Task 4: Add integration tests for new flags in cmd/run_test.go and cmd/exec_test.go (AC: #1-#5, #13)
  - [x] Note: These tests will be GREEN-on-arrival (implementation is complete from Tasks 2+3). Manually cross-check expected values before writing tests to ensure they capture correct behavior.
  - [x] In `TestRunRun_ModelFlags`: add test cases for `--haiku-model`, `--sonnet-model`, `--opus-model` flags (verify env vars via mock script)
  - [x] In `TestRunRun_ModelFlags`: add test case for `--small-fast-model` alias (verify it sets `ANTHROPIC_DEFAULT_HAIKU_MODEL`)
  - [x] In `TestRunRun_ModelFlags`: add test case for `--haiku-model` priority over `--small-fast-model`
  - [x] Update mock script to also capture `$ANTHROPIC_DEFAULT_SONNET_MODEL` and `$ANTHROPIC_DEFAULT_OPUS_MODEL`
  - [x] Update test struct: rename `wantSFM` field to `wantHaikuModel`, add `wantSonnetModel` and `wantOpusModel` fields; update all existing test case values from `wantSFM: ...` to `wantHaikuModel: ...`
  - [x] Mirror same changes in `TestExecRun_ModelFlags`

- [x] Task 5: Fix remaining test failures and verify (AC: #12)
  - [x] `go test ./...` — all tests pass
  - [x] `go vet ./...` — clean
  - [x] `make build` — static binary builds

## Dev Notes

### What Changed and Why

Story 4.1 expanded the config and runner layers with HaikuModel/SonnetModel/OpusModel fields and updated buildEnv() priority chains. This story connects the CLI flags to those fields.

**ExtractFlags signature change is a breaking internal API change.** The return value `(model, smallFastModel string, remaining []string, err error)` becomes `(model, haikuModel, sonnetModel, opusModel string, remaining []string, err error)`. Only two call sites exist: `cmd/run.go` and `cmd/exec.go`.

**Key backward compatibility behavior:** `--small-fast-model` no longer sets `opts.SmallFastModel`. It now sets `haikuModel` directly in ExtractFlags. The runner's buildEnv haiku priority chain (`opts.HaikuModel > opts.SmallFastModel > ctx.HaikuModel > ctx.SmallFastModel`) still works because:
- When `--haiku-model` is used → `opts.HaikuModel` is set
- When `--small-fast-model` is used → `haikuModel` return value is set → caller passes it as `opts.HaikuModel`
- `opts.SmallFastModel` is no longer set by any CLI flag, but the field remains in the Options struct for future programmatic use

### Files to Change

| File | Change | Scope |
|------|--------|-------|
| `internal/runner/args.go` | Expand ExtractFlags: new signature + 3 new flag branches + alias logic | MODIFY |
| `internal/runner/args_test.go` | Update TestExtractFlags struct and add new test cases | MODIFY |
| `cmd/run.go` | Update ExtractFlags call site + Options construction | MODIFY |
| `cmd/exec.go` | Update ExtractFlags call site + Options construction | MODIFY |
| `cmd/run_test.go` | Add new flag integration tests, update mock script | MODIFY |
| `cmd/exec_test.go` | Add new flag integration tests, update mock script | MODIFY |

### ExtractFlags New Signature

```go
func ExtractFlags(args []string) (model, haikuModel, sonnetModel, opusModel string, remaining []string, err error)
```

### Flag Branch Logic (args.go switch statement)

```go
// Declare a separate variable for the --small-fast-model alias
var sfmAlias string

// ... in the switch:
case "--haiku-model":
    if i+1 >= len(preSep) {
        return "", "", "", "", []string{}, fmt.Errorf("--haiku-model requires a value")
    }
    if err := validateFlagValue("--haiku-model", preSep[i+1]); err != nil {
        return "", "", "", "", []string{}, err
    }
    haikuModel = preSep[i+1]
    i += 2
case "--sonnet-model":
    if i+1 >= len(preSep) {
        return "", "", "", "", []string{}, fmt.Errorf("--sonnet-model requires a value")
    }
    if err := validateFlagValue("--sonnet-model", preSep[i+1]); err != nil {
        return "", "", "", "", []string{}, err
    }
    sonnetModel = preSep[i+1]
    i += 2
case "--opus-model":
    if i+1 >= len(preSep) {
        return "", "", "", "", []string{}, fmt.Errorf("--opus-model requires a value")
    }
    if err := validateFlagValue("--opus-model", preSep[i+1]); err != nil {
        return "", "", "", "", []string{}, err
    }
    opusModel = preSep[i+1]
    i += 2
case "--small-fast-model":
    if i+1 >= len(preSep) {
        return "", "", "", "", []string{}, fmt.Errorf("--small-fast-model requires a value")
    }
    if err := validateFlagValue("--small-fast-model", preSep[i+1]); err != nil {
        return "", "", "", "", []string{}, err
    }
    // Alias: collect into sfmAlias (last-wins for duplicates)
    sfmAlias = preSep[i+1]
    i += 2

// ... after the loop, before the return:
// Resolve alias: --haiku-model wins over --small-fast-model
if haikuModel == "" && sfmAlias != "" {
    haikuModel = sfmAlias
}
```

**Priority resolution uses a separate `sfmAlias` variable** so that each flag independently follows last-wins semantics for duplicates, and priority between `--haiku-model` and `--small-fast-model` is resolved once after the loop. This satisfies both AC5 (`--haiku-model` wins over `--small-fast-model` regardless of order) and AC11 (duplicate `--small-fast-model` follows last-wins).

### cmd/run.go Changes

```go
// Line 33: old
model, smallFastModel, args, err := runner.ExtractFlags(args)
// Line 33: new
model, haikuModel, sonnetModel, opusModel, args, err := runner.ExtractFlags(args)

// Lines 76-81: old
r, err := runner.New(runner.Options{
    ContextName:    provider,
    Target:         target,
    Model:          model,
    SmallFastModel: smallFastModel,
})
// Lines 76-81: new
r, err := runner.New(runner.Options{
    ContextName: provider,
    Target:      target,
    Model:       model,
    HaikuModel:  haikuModel,
    SonnetModel: sonnetModel,
    OpusModel:   opusModel,
})
```

### cmd/exec.go Changes

Same pattern as run.go at line 32 (ExtractFlags call) and line 74 (Options construction).

### Mock Script Update (Integration Tests)

Update mock script in `cmd/run_test.go` and `cmd/exec_test.go` to capture all model env vars:

```sh
#!/bin/sh
echo "$ANTHROPIC_MODEL" > "$MOCK_OUTPUT_FILE"
echo "$ANTHROPIC_DEFAULT_HAIKU_MODEL" >> "$MOCK_OUTPUT_FILE"
echo "$ANTHROPIC_DEFAULT_SONNET_MODEL" >> "$MOCK_OUTPUT_FILE"
echo "$ANTHROPIC_DEFAULT_OPUS_MODEL" >> "$MOCK_OUTPUT_FILE"
exit 0
```

Update test struct to verify all four model env vars:

```go
type modelFlagTest struct {
    name            string
    args            []string
    configTOML      string
    wantCode        int
    wantModel       string
    wantHaikuModel  string
    wantSonnetModel string
    wantOpusModel   string
}
```

Assertion block in the `t.Run` loop (after reading mock output):

```go
if tt.wantCode == 0 {
    data, err := os.ReadFile(outputFile)
    require.NoError(t, err)
    lines := strings.Split(string(data), "\n")
    if tt.wantModel != "" {
        require.GreaterOrEqual(t, len(lines), 1, "mock did not write ANTHROPIC_MODEL")
        assert.Equal(t, tt.wantModel, lines[0])
    }
    if tt.wantHaikuModel != "" {
        require.GreaterOrEqual(t, len(lines), 2, "mock did not write ANTHROPIC_DEFAULT_HAIKU_MODEL")
        assert.Equal(t, tt.wantHaikuModel, lines[1])
    }
    if tt.wantSonnetModel != "" {
        require.GreaterOrEqual(t, len(lines), 3, "mock did not write ANTHROPIC_DEFAULT_SONNET_MODEL")
        assert.Equal(t, tt.wantSonnetModel, lines[2])
    }
    if tt.wantOpusModel != "" {
        require.GreaterOrEqual(t, len(lines), 4, "mock did not write ANTHROPIC_DEFAULT_OPUS_MODEL")
        assert.Equal(t, tt.wantOpusModel, lines[3])
    }
}
```

Note: The unused `envSetup` field in the current `TestExecRun_ModelFlags` struct (exec_test.go) can be dropped when defining the new `modelFlagTest` struct, since no existing test case uses it.

Note: exec_test.go stores its mock script as a Go `[]byte` variable (`mockScript` at line 96) with conditional filename logic — the `hasExplicitTarget` branching (lines 99-122) determines whether the mock file is named after the first target arg after `--` or falls back to `"bash"`. Apply the same 2-line→4-line expansion to the `mockScript` variable body, preserving the existing `hasExplicitTarget` branching. This differs from run_test.go which writes the mock directly to a hardcoded `claudePath`.

### TestExtractFlags Struct Update

```go
tests := []struct {
    name              string
    args              []string
    wantModel         string
    wantHaikuModel    string
    wantSonnetModel   string
    wantOpusModel     string
    wantRemaining     []string
    wantErr           string
}{...}
```

Call changes from:
```go
model, sfm, remaining, err := ExtractFlags(tt.args)
```
To:
```go
model, haiku, sonnet, opus, remaining, err := ExtractFlags(tt.args)
```

### Story 4.1 Learnings

- Story 4.1 already updated `cmd/run_test.go` and `cmd/exec_test.go` to use `ANTHROPIC_DEFAULT_HAIKU_MODEL` instead of `ANTHROPIC_SMALL_FAST_MODEL` in mock scripts
- The `Options.SmallFastModel` field still exists in runner.go but is no longer set by any CLI flag after this story
- buildEnv haiku priority chain: `opts.HaikuModel > opts.SmallFastModel > ctx.HaikuModel > ctx.SmallFastModel > omit` — the `opts.SmallFastModel` rung becomes unreachable via CLI but remains for programmatic API completeness
- Config layer already has `HaikuModel`, `SonnetModel`, `OpusModel` fields with mapstructure tags

### Anti-Patterns (Do NOT)

- Do NOT remove `SmallFastModel` field from `Options` struct — backward compat for programmatic use
- Do NOT change `buildEnv()` in runner.go — it already handles the new fields correctly from Story 4.1
- Do NOT change `ParseArgs()` — argument parsing is unchanged
- Do NOT change `config/config.go` — config fields already exist
- Do NOT call `os.Exit()` inside `internal/runner/` — breaks testability
- Do NOT inject `ANTHROPIC_SMALL_FAST_MODEL` — it is deprecated, replaced by `ANTHROPIC_DEFAULT_HAIKU_MODEL`
- Do NOT change error message format — follow `"<flag> requires a value"` pattern
- Do NOT add Cobra flags — use manual flag parsing via `ExtractFlags` (avoids `--` separator conflict)
- Do NOT modify `cmd/list.go` — it doesn't use flags

### Project Structure Notes

- ExtractFlags is the ONLY function in args.go that needs changes (ParseArgs and WantsHelp are unchanged)
- The `--small-fast-model` alias behavior is implemented in ExtractFlags, not in runner.go — the conversion from "small fast model" to "haiku model" happens at the CLI boundary
- `Options.SmallFastModel` in runner.go becomes dead code from CLI perspective but is kept for API stability

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 4.2] — story definition and AC
- [Source: _bmad-output/planning-artifacts/architecture.md#Argument Parsing Patterns] — ExtractFlags signature and flag mapping table
- [Source: _bmad-output/planning-artifacts/architecture.md#Command File Template] — cmd template with new flag names
- [Source: _bmad-output/implementation-artifacts/4-1-config-runner-model-field-expansion.md] — previous story learnings
- [Source: internal/runner/args.go] — current ExtractFlags implementation (2 flags)
- [Source: internal/runner/args_test.go] — current test structure
- [Source: cmd/run.go] — call site to update
- [Source: cmd/exec.go] — call site to update
- [Source: internal/runner/runner.go:14-22] — Options struct (HaikuModel/SonnetModel/OpusModel already exist)
- [Source: internal/runner/runner.go:105-136] — buildEnv already handles new fields (no changes needed)

## Dev Agent Record

### Agent Model Used

glm-5.1

### Debug Log References

### Completion Notes List

- ✅ Task 1: RED — Updated TestExtractFlags struct from wantSmallFast→wantHaikuModel, added wantSonnetModel/wantOpusModel fields. Added 18 new test cases covering all new flags, alias, priority, errors, newline validation, separator boundary, duplicates, and combined scenarios. Tests confirmed failing before implementation.
- ✅ Task 2: GREEN — Updated ExtractFlags signature to return 6 values (model, haikuModel, sonnetModel, opusModel, remaining, err). Added --haiku-model, --sonnet-model, --opus-model branches with validation. Changed --small-fast-model to use sfmAlias with post-loop priority resolution (haiku-model wins). All 34 args_test.go tests pass.
- ✅ Task 3: Updated cmd/run.go and cmd/exec.go ExtractFlags call sites (4→6 return values) and Options construction (SmallFastModel→HaikuModel/SonnetModel/OpusModel). Build succeeds.
- ✅ Task 4: Added integration tests in run_test.go (6 new cases) and exec_test.go (5 new cases). Updated mock scripts to capture ANTHROPIC_DEFAULT_SONNET_MODEL and ANTHROPIC_DEFAULT_OPUS_MODEL. All integration tests pass.
- ✅ Task 5: go test ./... PASS, go vet ./... clean, make build succeeds. Zero regressions.

### File List

- internal/runner/args.go — MODIFIED: Expanded ExtractFlags with 3 new flag branches + alias logic
- internal/runner/args_test.go — MODIFIED: Updated test struct + added 18 new test cases
- cmd/run.go — MODIFIED: Updated ExtractFlags call site + Options construction
- cmd/exec.go — MODIFIED: Updated ExtractFlags call site + Options construction
- cmd/run_test.go — MODIFIED: Added integration tests for new flags, updated mock script
- cmd/exec_test.go — MODIFIED: Added integration tests for new flags, updated mock script

### Review Findings

- [x] [Review][Patch] Integration tests missing error-path coverage for new flags [cmd/run_test.go, cmd/exec_test.go]
- [x] [Review][Patch] Flag-value-looking-like-flag tests only cover --model consuming --small-fast-model, not the new flag names [internal/runner/args_test.go]
- [x] [Review][Patch] all 5 flags combined integration tests omit --small-fast-model (alias coverage gap in end-to-end) [cmd/run_test.go, cmd/exec_test.go]
- [x] [Review][Patch] No test for nil or empty args slice to ExtractFlags [internal/runner/args_test.go]
- [x] [Review][Patch] No test for flag at end of argument list with missing value (e.g., `{"provider-A", "--haiku-model"}`) [internal/runner/args_test.go]
- [x] [Review][Patch] AC5 reverse ordering test missing (--small-fast-model before --haiku-model) [internal/runner/args_test.go]
- [x] [Review][Defer] Near-identical test tables duplicated across exec_test.go and run_test.go — deferred, pre-existing
- [x] [Review][Defer] validateFlagValue returns ambiguous error messages without flag context — deferred, pre-existing
- [x] [Review][Defer] exec_test.go missing general error-path tests (pre-existing, not caused by this change) — deferred, pre-existing
- [x] [Review][Defer] --haiku-model "" (explicit empty) can be silently overridden by --small-fast-model alias — deferred, extremely unlikely edge case
- [x] [Review][Defer] Mock script line ordering fragile to future env var injection changes — deferred, design observation
- [x] [Review][Defer] No tests for unicode or extremely long flag values — deferred, low priority
