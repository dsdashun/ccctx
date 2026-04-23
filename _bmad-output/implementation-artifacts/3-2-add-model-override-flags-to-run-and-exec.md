# Story 3.2: Add --model Override Flags to run and exec

Status: done

## Story

As a **ccctx user**,
I want **to override the provider's configured model via `--model <value>` and `--small-fast-model <value>` flags on `run` and `exec` commands**,
so that **I can quickly switch models without editing the config file**.

## Acceptance Criteria

1. **AC1: --model flag overrides config model** — Given provider-A has `model = "claude-sonnet-4-6"` in config, `ccctx run provider-A --model claude-opus-4-7` sets `ANTHROPIC_MODEL=claude-opus-4-7` in the child process. [Source: epics.md#Story 3.2, FR27]

2. **AC2: --model flag sets model when config has none** — Given provider-A has no `model` configured, `ccctx exec provider-A --model claude-opus-4-7 -- env | grep ANTHROPIC_MODEL` outputs `ANTHROPIC_MODEL=claude-opus-4-7`. [Source: epics.md#Story 3.2]

3. **AC3: --small-fast-model flag works** — Given provider-A has `small_fast_model = "claude-haiku-4-5"`, `ccctx run provider-A --small-fast-model claude-sonnet-4-6` sets `ANTHROPIC_SMALL_FAST_MODEL=claude-sonnet-4-6`. [Source: epics.md#Story 3.2]

4. **AC4: CLI flags take priority over config** — When both CLI flag and config value exist, CLI flag wins. Priority chain: `Options.Model > ctx.Model > omit` (same for SmallFastModel). [Source: architecture.md#Model priority, FR28]

5. **AC5: No flags = config values used** — When no flags specified, config file values are used (existing behavior unchanged). [Source: epics.md#Story 3.2]

6. **AC6: Flags available on both commands** — `--model` and `--small-fast-model` work identically on both `run` and `exec`. [Source: FR27, FR28]

7. **AC7: --model with TUI mode** — `ccctx run --model claude-opus-4-7` (no provider) opens TUI selector, and the selected provider runs with the overridden model.

8. **AC8: -- separator not consumed by flag parsing** — `ccctx exec provider-A --model claude-opus-4-7 -- env | grep ANTHROPIC_MODEL` correctly forwards `env | grep ANTHROPIC_MODEL` as the target command, not consumed by flag parsing. [Source: architecture.md#Phase 2 conflict]

9. **AC9: All existing tests pass** — `go test ./...` passes, `go vet ./...` clean, `make build` succeeds.

10. **AC10: New table-driven tests** — Tests cover flag extraction, priority chain, and integration with both commands.

## Tasks / Subtasks

- [x] Task 1: RED — write failing tests for ExtractFlags and WantsHelp in args_test.go (AC: #8, #10)
  - [x] Table-driven tests for `ExtractFlags` in `args_test.go` covering:
    - No flags → returns empty strings, args unchanged
    - `--model foo provider-A` → model="foo", remaining=["provider-A"]
    - `provider-A --model foo` → model="foo", remaining=["provider-A"]
    - `provider-A --model foo -- --help` → model="foo", remaining=["provider-A", "--", "--help"]
    - Both flags: `provider-A --model foo --small-fast-model bar -- cmd` → both extracted
    - `--model` without value → error
    - `--small-fast-model` without value → error
    - `--model` after `--` separator → NOT extracted (forwarded arg)
    - No provider with flag: `--model foo` → model="foo", remaining=[]
    - Flag value is `-`: `provider-A --model -` → model="-", remaining=["provider-A"]
    - Duplicate flags: `provider-A --model foo --model bar` → model="bar", remaining=["provider-A"]
  - [x] Table-driven tests for `WantsHelp` in `args_test.go` covering:
    - `["--help"]` → true
    - `["-h"]` → true
    - `["provider-A", "--", "--help"]` → false (after separator)
    - `[]` → false
    - `["provider-A", "--help"]` → true
    - `["provider-A"]` → false
  - [x] Verify tests fail to compile (ExtractFlags and WantsHelp don't exist yet)

- [x] Task 2: GREEN — implement ExtractFlags in internal/runner/args.go (AC: #8)
  - [x] Write `ExtractFlags(args []string) (model, smallFastModel string, remaining []string, err error)` in `args.go`
  - [x] Logic: find `--` separator, scan only args before it for `--model <value>` and `--small-fast-model <value>` pairs, remove matched pairs, return remaining args + flag values
  - [x] Error cases: `--model` without value → `"--model requires a value"`, same for `--small-fast-model`
  - [x] Error returns use `[]string{}` (not `nil`) for remaining slice
  - [x] Keep ParseArgs unchanged — its flag rejection catches unknown flags after extraction
  - [x] Verify Task 1 tests now pass

- [x] Task 3: RED — write failing tests for buildEnv priority chain in runner_test.go (AC: #1-#5, #10)
  - [x] Table-driven tests in `runner_test.go` covering:
    - opts.Model set, ctx.Model empty → uses opts.Model
    - opts.Model set, ctx.Model set → opts.Model wins
    - opts.Model empty, ctx.Model set → uses ctx.Model
    - Both empty → no ANTHROPIC_MODEL injected
    - Same 4 cases for SmallFastModel
    - Mixed case: opts.Model set, ctx.SmallFastModel set → correct priority for each
  - [x] Verify tests fail (buildEnv doesn't check opts.Model yet)

- [x] Task 4: GREEN — implement priority chain in runner.go:buildEnv (AC: #1-#5)
  - [x] Change model logic: `opts.Model > ctx.Model > omit` pattern
  - [x] Same pattern for SmallFastModel
  - [x] Verify Task 3 tests now pass

- [x] Task 5: RED — update breaking test + write failing integration tests for run command flags (AC: #6, #7, #10)
  - [x] Update existing test "ParseArgs error - flag-like arg" in `cmd/run_test.go`: change args from `[]string{"--model", "foo"}` to `[]string{"--unknown-flag", "foo"}` (since `--model` is now extracted by ExtractFlags instead of rejected by ParseArgs)
  - [x] Add table-driven tests in `cmd/run_test.go` covering:
    - `--model foo provider-A` with env-capturing claude mock → success, verify ANTHROPIC_MODEL env var (see "Env-Capturing Mock Pattern" in Dev Notes)
    - `--model foo provider-A -- --help` → args forwarded correctly, verify success
    - `--model` without value → exit code 1
    - Both flags: `provider-A --model foo --small-fast-model bar` → success, verify both env vars
  - [x] Verify new tests fail (runRun doesn't call ExtractFlags yet)

- [x] Task 6: RED — write failing integration tests for exec command flags in cmd/exec_test.go (AC: #6, #10)
  - [x] Create new file `cmd/exec_test.go` (package cmd, like run_test.go)
  - [x] Table-driven tests covering:
    - `--model foo provider-A` with env-capturing shell mock → success, verify ANTHROPIC_MODEL env var (see "Env-Capturing Mock Pattern" in Dev Notes; mock named as `$SHELL` basename, e.g. `bash`)
    - `--model` without value → exit code 1
    - `provider-A --model foo -- env | grep MODEL` → success with model override. Mock named `env` (first target arg after `--`), not `$SHELL` basename.
  - [x] Verify tests fail (execRun doesn't call ExtractFlags yet)

- [x] Task 7: GREEN — update cmd/run.go and cmd/exec.go to use ExtractFlags (AC: #6, #7, #9)
  - [x] Set `DisableFlagParsing: true` on RunCmd and ExecCmd
  - [x] Add `runner.WantsHelp(args)` check in Run closures before calling runRun/execRun — if true, call `cmd.Help()` and return (prevents `--help` regression from DisableFlagParsing)
  - [x] In `runRun()` and `execRun()`, call `runner.ExtractFlags(args)` before `runner.ParseArgs()`
  - [x] Pass `model` and `smallFastModel` into `runner.Options{...}`
  - [x] Error handling: print extraction error to stderr, return 1
  - [x] Verify Tasks 5-6 tests now pass

- [x] Task 8: Verify all tests and build (AC: #9)
  - [x] `go test ./...` passes
  - [x] `go vet ./...` clean
  - [x] `make build` succeeds
  - [ ] Manual smoke test: `ccctx run <provider> --model <model-name>`
  - [ ] Manual smoke test: `ccctx run --help` shows help (not error)
  - [ ] Manual smoke test: `ccctx exec --help` shows help (not error)

## Dev Notes

### Critical Design Decision: Flag Parsing vs -- Separator

The architecture doc explicitly warns: registering `--model` as a Cobra flag causes Cobra to consume the `--` separator, breaking `ParseArgs`. **Resolution: set `DisableFlagParsing: true` on both commands and parse flags manually.**

This means:
- Cobra passes raw args to our `Run` function (no flag processing)
- We call `runner.ExtractFlags(args)` to extract `--model` and `--small-fast-model` from args before `--`
- We pass remaining args to `runner.ParseArgs()` (unchanged)
- ParseArgs' existing flag rejection still catches unknown flags (e.g., `--unknown-flag`)

### Flag Extraction Algorithm

`ExtractFlags` scans args before the `--` separator for known flag-value pairs. After extraction, remaining args should contain at most one non-flag arg (the provider name) before `--`.

```go
func ExtractFlags(args []string) (model, smallFastModel string, remaining []string, err error) {
    sepIdx := len(args) // scan all args by default
    for i, a := range args {
        if a == "--" {
            sepIdx = i
            break
        }
    }

    // Build remaining by scanning pre-separator args
    remaining = make([]string, 0, len(args))
    preSep := args[:sepIdx]
    i := 0
    for i < len(preSep) {
        switch preSep[i] {
        case "--model":
            if i+1 >= len(preSep) {
                return "", "", []string{}, fmt.Errorf("--model requires a value")
            }
            model = preSep[i+1]
            i += 2
        case "--small-fast-model":
            if i+1 >= len(preSep) {
                return "", "", []string{}, fmt.Errorf("--small-fast-model requires a value")
            }
            smallFastModel = preSep[i+1]
            i += 2
        default:
            remaining = append(remaining, preSep[i])
            i++
        }
    }

    // Append separator and post-separator args unchanged
    if sepIdx < len(args) {
        remaining = append(remaining, args[sepIdx:]...)
    }

    return model, smallFastModel, remaining, nil
}
```

### Help Handling (prevents --help regression)

With `DisableFlagParsing: true`, Cobra no longer processes `--help`. Add a `WantsHelp` helper in `args.go` and check it in the Run closures before calling runRun/execRun:

```go
// WantsHelp checks if --help or -h appears before -- in args.
func WantsHelp(args []string) bool {
    for _, a := range args {
        if a == "--" {
            break
        }
        if a == "--help" || a == "-h" {
            return true
        }
    }
    return false
}
```

Note: `ccctx --help run` and `ccctx help run` still work (parent command handles them).

### Command File Changes

**cmd/run.go** (minimal diff):

```go
var RunCmd = &cobra.Command{
    Use:              "run [context] [-- claude-args...]",
    Short:            "Run claude with a context",
    Long:             "...",
    Args:             cobra.ArbitraryArgs,
    DisableFlagParsing: true,  // ADD THIS
    Run: func(cmd *cobra.Command, args []string) {
        if runner.WantsHelp(args) {  // ADD: handle --help before DisableFlagParsing
            cmd.Help()
            return
        }
        os.Exit(runRun(args))
    },
}

func runRun(args []string) int {
    // ADD: extract flags before ParseArgs
    model, smallFastModel, args, err := runner.ExtractFlags(args)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return 1
    }

    provider, targetArgs, useTUI, err := runner.ParseArgs(args)
    // ... (TUI flow unchanged) ...

    r, err := runner.New(runner.Options{
        ContextName:    provider,
        Target:         target,
        Model:          model,          // ADD
        SmallFastModel: smallFastModel, // ADD
    })
    // ... (rest unchanged) ...
}
```

**cmd/exec.go**: identical pattern — add `DisableFlagParsing: true`, add `WantsHelp` check in Run closure, call `ExtractFlags` at top of `execRun`, pass flags in Options.

### buildEnv Priority Chain

Current `buildEnv` in `runner.go:82-99` uses `ctx.Model` directly. Change to:

```go
func buildEnv(ctx *config.Context, opts Options) []string {
    // ... filtering unchanged ...

    filtered = append(filtered, "ANTHROPIC_BASE_URL="+ctx.BaseURL)
    filtered = append(filtered, "ANTHROPIC_AUTH_TOKEN="+ctx.AuthToken)

    // Priority: CLI flag > config value > omit
    model := opts.Model
    if model == "" {
        model = ctx.Model
    }
    if model != "" {
        filtered = append(filtered, "ANTHROPIC_MODEL="+model)
    }

    sfm := opts.SmallFastModel
    if sfm == "" {
        sfm = ctx.SmallFastModel
    }
    if sfm != "" {
        filtered = append(filtered, "ANTHROPIC_SMALL_FAST_MODEL="+sfm)
    }

    return filtered
}
```

This is a minimal change — only the model/small_fast_model injection logic changes. The existing injection order (BASE_URL → AUTH_TOKEN → MODEL → SMALL_FAST_MODEL) is preserved.

### Key Files

| File | Role | Change Type |
|------|------|-------------|
| `internal/runner/args.go` | ParseArgs, ExtractFlags, WantsHelp | MODIFY — add ExtractFlags and WantsHelp functions |
| `internal/runner/runner.go` | buildEnv | MODIFY — implement priority chain |
| `cmd/run.go` | run command | MODIFY — add DisableFlagParsing, WantsHelp check, ExtractFlags call, pass flags |
| `cmd/exec.go` | exec command | MODIFY — same as run.go |
| `internal/runner/args_test.go` | Tests for ExtractFlags | MODIFY — add test cases |
| `internal/runner/runner_test.go` | Tests for buildEnv priority | MODIFY — add test cases |
| `cmd/run_test.go` | Integration tests | MODIFY — update breaking test, add flag test cases |
| `cmd/exec_test.go` | Integration tests for exec | CREATE — new file for exec flag tests |

### Env-Capturing Mock Pattern

Integration tests need to verify that `ANTHROPIC_MODEL` (and optionally `ANTHROPIC_SMALL_FAST_MODEL`) are set in the child process environment. The existing mocks (`exit 0` / `exit 42`) don't capture env vars. Use this pattern:

**Mock script** (writes env vars to a file for assertion):
```sh
#!/bin/sh
echo "$ANTHROPIC_MODEL" > "$MOCK_OUTPUT_FILE"
echo "$ANTHROPIC_SMALL_FAST_MODEL" >> "$MOCK_OUTPUT_FILE"
exit 0
```

**Test setup**:
1. Create temp file for mock output: `outputFile := filepath.Join(t.TempDir(), "mock_output")`
2. Create mock script that writes to `$MOCK_OUTPUT_FILE`
3. Set `MOCK_OUTPUT_FILE` env var in test: `t.Setenv("MOCK_OUTPUT_FILE", outputFile)`
4. Set `PATH` to temp dir containing the mock. Mock naming rules:
   - **run tests**: mock named `claude` (runRun calls `exec.LookPath("claude")`)
   - **exec tests with no target args**: mock named as `$SHELL` basename (e.g., `bash`) — exec falls back to `$SHELL`. **Important:** must also set `SHELL` env var to mock's full path: `t.Setenv("SHELL", filepath.Join(tmpDir, "bash"))`. Without this, `os.Getenv("SHELL")` returns the real shell path (e.g., `/bin/bash`) and `exec.Command` runs the real shell instead of the mock.
   - **exec tests with explicit target args after `--`**: mock named after the first target arg (e.g., `env` for `-- env | grep MODEL`)
5. After `runRun(args)` returns, read `outputFile` and assert contents

This approach avoids modifying the runner or cmd code — the env var propagation is tested end-to-end through the actual `exec.LookPath` → `exec.Cmd.Env` → child process path.

### Previous Story Intelligence

From Story 3-1 implementation:
- `ParseArgs` now rejects flag-like args in provider position — this is the **critical prerequisite** that allows us to catch unknown flags after our extraction
- `cmd/run.go` and `cmd/exec.go` both use `runRun(args) int` / `execRun(args) int` pattern — `os.Exit()` only at top level
- `Options` struct already has `Model` and `SmallFastModel` fields — we just need to populate them
- `buildEnv` already receives `opts Options` — we just need to check `opts.Model` before `ctx.Model`
- `cmd/run_test.go` uses `package cmd` (not `cmd_test`) to access unexported `runRun`

From Story 2-2 implementation:
- TUI uses sentinel error `ErrCancelled` — checked via `errors.Is(err, ui.ErrCancelled)`
- All error output uses `fmt.Fprintf(os.Stderr, ...)` or `fmt.Fprintln(os.Stderr, ...)`

From Story 1-3 implementation:
- Runner struct carries `ctx`, `opts`, `env`
- `buildEnv` signature: `buildEnv(ctx *config.Context, opts Options) []string`
- Both commands follow identical error handling pattern

### Anti-Patterns (Do NOT)

- Do NOT register `--model` as a Cobra flag — it will consume the `--` separator
- Do NOT modify `ParseArgs` — keep its flag rejection for unknown flags
- Do NOT extract flags from after the `--` separator — those are forwarded args
- Do NOT call `os.Exit()` inside `internal/runner/` — breaks testability
- Do NOT use `fmt.Fprintf(os.Stderr, ...)` inside runner — commands own output
- Do NOT change the buildEnv injection order — BASE_URL → AUTH_TOKEN → MODEL → SMALL_FAST_MODEL
- Do NOT inject empty model values — check `!= ""` before appending
- Do NOT duplicate the flag extraction logic between run.go and exec.go — use shared `ExtractFlags` in runner package
- Do NOT add `-m` or other shorthand flags — only `--model` and `--small-fast-model` as specified
- Do NOT modify `config/config.go` or `internal/ui/selector.go` — not in scope

### Project Structure Notes

- All changes in existing packages (`internal/runner/`, `cmd/`)
- `ExtractFlags` and `WantsHelp` go in `args.go` alongside `ParseArgs` — all are argument processing functions
- `DisableFlagParsing: true` is a Cobra feature that prevents any flag processing, giving us raw args
- No new dependencies needed
- All Go source files use tab indentation — run `go fmt` after edits. Code snippets in this plan use spaces for readability only.

### Edge Cases to Handle

1. **Flag after --**: `ccctx run provider-A -- --model foo` → `--model foo` is forwarded to claude, NOT extracted. ExtractFlags only scans before `--`.
2. **No provider with flag**: `ccctx run --model foo` → TUI mode with model override. After extraction, remaining args are empty → useTUI=true.
3. **Flag value starts with dash**: `ccctx run provider-A --model -` → model value is `-`. This is technically valid (though unusual). ExtractFlags should accept it.
4. **Duplicate flags**: `ccctx run provider-A --model foo --model bar` → last value wins (second overwrites first in sequential scan).
5. **Flag before provider**: `ccctx run --model foo provider-A` → works, ExtractFlags reorders to extract flag, leaving `provider-A` in remaining.
6. **--help before separator**: `ccctx run --help` → WantsHelp returns true, shows Cobra help. `ccctx run provider-A -- --help` → `--help` after `--` is forwarded, not intercepted.
7. **Unknown flag**: `ccctx run --unknown-flag foo` → not extracted by ExtractFlags, passed to ParseArgs, rejected as flag-like arg (existing behavior).

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 3.2] — story definition and AC
- [Source: _bmad-output/planning-artifacts/architecture.md#Phase 2] — known --model vs -- conflict, model priority chain
- [Source: _bmad-output/planning-artifacts/architecture.md#AD1] — Runner struct design with Model/SmallFastModel fields
- [Source: _bmad-output/planning-artifacts/prd.md#FR27-FR28] — functional requirements for model override
- [Source: internal/runner/runner.go:82-99] — buildEnv, current model injection logic
- [Source: internal/runner/runner.go:14-19] — Options struct with Model/SmallFastModel fields
- [Source: internal/runner/args.go:28-29,40-41] — flag rejection in ParseArgs (prerequisite from Story 3.1)
- [Source: cmd/run.go:15-23] — RunCmd definition, needs DisableFlagParsing
- [Source: cmd/exec.go:14-22] — ExecCmd definition, needs DisableFlagParsing
- [Source: cmd/run.go:25-78] — runRun function, needs ExtractFlags call
- [Source: cmd/exec.go:24-74] — execRun function, needs ExtractFlags call
- [Source: _bmad-output/implementation-artifacts/3-1-runner-and-command-layer-robustness.md] — previous story learnings
- [Source: _bmad-output/project-context.md] — implementation rules and patterns

## Dev Agent Record

### Agent Model Used

Claude (glm-5.1)

### Debug Log References

### Completion Notes List

- Implemented `ExtractFlags` in `internal/runner/args.go` — extracts `--model` and `--small-fast-model` from args before `--` separator, handles all edge cases (flag after separator, missing value, duplicate flags, dash value)
- Implemented `WantsHelp` in `internal/runner/args.go` — checks for `--help`/`-h` before `--` separator, prevents regression from `DisableFlagParsing: true`
- Updated `buildEnv` in `internal/runner/runner.go` — priority chain `opts.Model > ctx.Model > omit` for both Model and SmallFastModel
- Updated `cmd/run.go` — added `DisableFlagParsing: true`, `WantsHelp` check, `ExtractFlags` call, passes model flags in Options
- Updated `cmd/exec.go` — identical pattern as run.go
- Updated `cmd/run_test.go` — changed breaking test from `--model` to `--unknown-flag`, added `TestRunRun_ModelFlags` with env-capturing mock pattern
- Created `cmd/exec_test.go` — new file with `TestExecRun_ModelFlags` using env-capturing mocks for shell and explicit target scenarios

### File List

- `internal/runner/args.go` — modified: added `ExtractFlags` and `WantsHelp` functions
- `internal/runner/args_test.go` — modified: added `TestExtractFlags` and `TestWantsHelp` table-driven tests
- `internal/runner/runner.go` — modified: updated `buildEnv` with priority chain for Model and SmallFastModel
- `internal/runner/runner_test.go` — modified: added `TestBuildEnv_PriorityChain` table-driven tests
- `cmd/run.go` — modified: added `DisableFlagParsing`, `WantsHelp` check, `ExtractFlags` call, model flags in Options
- `cmd/exec.go` — modified: same changes as run.go
- `cmd/run_test.go` — modified: updated breaking test args, added `TestRunRun_ModelFlags` integration tests
- `cmd/exec_test.go` — created: new file with `TestExecRun_ModelFlags` integration tests

### Review Findings

- [x] [Review][Patch] `run_test.go` 缺少 `--small-fast-model` 无值错误测试，`exec_test.go` 完全缺少 `--small-fast-model` 覆盖 [cmd/run_test.go, cmd/exec_test.go] — 已添加测试
- [x] [Review][Patch] 标志值看起来像另一个标志时（如 `--model --small-fast-model`）被误解析为值，缺少测试 [internal/runner/args.go:24-39] — 已添加测试记录此行为（规范允许 `--model -`，故不添加拒绝逻辑）
- [x] [Review][Patch] 缺少重复 `--small-fast-model` 标志测试（规范要求 last wins）[internal/runner/args_test.go] — 已添加测试
- [x] [Review][Patch] 缺少组合标志行为测试 [internal/runner/args_test.go] — 已添加测试
- [x] [Review][Patch] `ExtractFlags` 错误消息格式不一致（句子大小写 vs `ParseArgs` 小写）[internal/runner/args.go] — 经核实，错误消息已一致（均为小写开头），无需修改
- [x] [Review][Patch] `buildEnv` 缺少 `opts > ctx` 优先级规则的代码注释 [internal/runner/runner.go:93-107] — 已添加注释
- [x] [Review][Patch] 缺少 `--` 分隔符后无参数的测试（如 `--model foo --`）[internal/runner/args_test.go] — 已添加测试
- [x] [Review][Patch] `cmd.Help()` 错误被静默吞掉，未处理 [cmd/run.go:22-25, cmd/exec.go:21-24] — 已添加错误处理
- [x] [Review][Patch] model/SFM 值可能包含换行，导致环境变量注入 [internal/runner/runner.go:93-107] — 已在 `ExtractFlags` 中添加换行符验证（CLI 输入路径）；配置值路径已在 Story 3.1 中标记为 deferred
- [x] [Review][Patch] 测试输出文件读取逻辑在 wantSFM 为空时不会验证第二行不存在 [cmd/run_test.go:216-225, cmd/exec_test.go:99-108] — 已改用 `require` 断言，确保预期值存在时才验证
- [x] [Review][Defer] `claude` 二进制存在但不可执行时，`exec.LookPath` 成功但 `exec.Command` 会失败 [cmd/run.go:66-70] — deferred, pre-existing
- [x] [Review][Defer] `SHELL` 环境变量设为空字符串时会被当作目标命令传递 [cmd/exec.go:63-70] — deferred, pre-existing
