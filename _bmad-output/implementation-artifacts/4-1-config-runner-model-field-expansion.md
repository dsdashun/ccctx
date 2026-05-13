# Story 4.1: Config & Runner Model Field Expansion

Status: done

## Story

As a **developer**,
I want **config and runner layers expanded to support new model fields (haiku_model, sonnet_model, opus_model) and inject the updated Claude Code environment variables (ANTHROPIC_DEFAULT_HAIKU_MODEL, ANTHROPIC_DEFAULT_SONNET_MODEL, ANTHROPIC_DEFAULT_OPUS_MODEL)**,
so that **ccctx correctly injects Claude Code's current model environment variables and users can override haiku/sonnet/opus model selection through config or CLI flags**.

## Acceptance Criteria

1. **AC1: haiku_model config field injects correct env var** — Given a context configured with `haiku_model = "claude-haiku-4-5-20251001"`, `buildEnv()` injects `ANTHROPIC_DEFAULT_HAIKU_MODEL=claude-haiku-4-5-20251001`. [Source: epics.md#Story 4.1]

2. **AC2: haiku_model takes priority over small_fast_model** — Given a context configured with both `haiku_model` and `small_fast_model`, `buildEnv()` uses `haiku_model` value for `ANTHROPIC_DEFAULT_HAIKU_MODEL`, ignoring `small_fast_model`. [Source: epics.md#Story 4.1]

3. **AC3: small_fast_model fallback to new env var** — Given a context with only `small_fast_model` configured (no `haiku_model`), `buildEnv()` injects `ANTHROPIC_DEFAULT_HAIKU_MODEL` with the `small_fast_model` value. `ANTHROPIC_SMALL_FAST_MODEL` is **never** injected. [Source: epics.md#Story 4.1, architecture.md#Environment Variable Construction]

4. **AC4: no model fields = no injection** — Given a context with no model fields configured, `buildEnv()` does not inject any `ANTHROPIC_DEFAULT_*_MODEL` variables. [Source: epics.md#Story 4.1]

5. **AC5: CLI flag priority over config** — `Options.HaikuModel/SonnetModel/OpusModel` fields take priority over config values when set. Priority chain: `Options.HaikuModel > Options.SmallFastModel > ctx.HaikuModel > ctx.SmallFastModel > omit` for haiku. `Options.SonnetModel > ctx.SonnetModel > omit` for sonnet. `Options.OpusModel > ctx.OpusModel > omit` for opus. [Source: architecture.md#Model priority]

6. **AC6: sonnet_model and opus_model work** — Given `sonnet_model = "claude-sonnet-4-6"` or `opus_model = "claude-opus-4-7"` in config, `buildEnv()` injects `ANTHROPIC_DEFAULT_SONNET_MODEL` or `ANTHROPIC_DEFAULT_OPUS_MODEL` respectively. [Source: epics.md#Story 4.1]

7. **AC7: All existing tests pass** — `go test ./...` passes, `go vet ./...` clean, `make build` succeeds. No regressions. [Source: project-context.md#Testing Rules]

8. **AC8: New table-driven tests** — Tests cover: new config fields, haiku priority chain (opts.HaikuModel > opts.SmallFastModel > ctx.HaikuModel > ctx.SmallFastModel > omit), sonnet/opus priority chains, small_fast_model backward compat, ANTHROPIC_SMALL_FAST_MODEL no longer injected, updated injection order. [Source: project-context.md#Testing Rules]

## Tasks / Subtasks

- [x] Task 1: RED — add failing tests for new config fields in config_test.go (AC: #1-#4, #6, #8)
  - [x] Table-driven test `TestGetContext_ModelFields` verifying that `GetContext` correctly copies all model fields (Model, SmallFastModel, HaikuModel, SonnetModel, OpusModel) from config to resolved context
  - [x] Write a TOML config with all 5 model fields, call `GetContext`, verify all 5 values present in resolved Context
  - [x] Write a TOML config with only `small_fast_model`, verify HaikuModel is empty (config layer doesn't promote SmallFastModel to HaikuModel — that's runner's job)

- [x] Task 2: GREEN — add HaikuModel, SonnetModel, OpusModel to config.Context struct (AC: #1, #6)
  - [x] In `config/config.go:12-17`, add three fields to Context struct:
    ```go
    HaikuModel  string `mapstructure:"haiku_model"`
    SonnetModel string `mapstructure:"sonnet_model"`
    OpusModel   string `mapstructure:"opus_model"`
    ```
  - [x] In `config/config.go:128-134` (`GetContext` resolvedContext), add the new fields to the copy:
    ```go
    HaikuModel:  context.HaikuModel,
    SonnetModel: context.SonnetModel,
    OpusModel:   context.OpusModel,
    ```
  - [x] Update default config template in `config/config.go:69-76` to include commented-out examples for new fields
  - [x] Verify Task 1 tests pass

- [x] Task 3: RED — add failing tests for buildEnv model expansion in runner_test.go (AC: #1-#6, #8)
  - [x] Update `TestBuildEnv` to verify new env vars: assert `ANTHROPIC_DEFAULT_HAIKU_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`, `ANTHROPIC_DEFAULT_OPUS_MODEL` present; assert `ANTHROPIC_SMALL_FAST_MODEL` NOT present
  - [x] Update `TestBuildEnv_SkipsEmptyOptional` to verify no `ANTHROPIC_DEFAULT_*_MODEL` injected when all model fields empty
  - [x] Update `TestBuildEnv_InjectionOrder` to verify new order: `ANTHROPIC_BASE_URL → ANTHROPIC_AUTH_TOKEN → ANTHROPIC_MODEL → ANTHROPIC_DEFAULT_HAIKU_MODEL → ANTHROPIC_DEFAULT_SONNET_MODEL → ANTHROPIC_DEFAULT_OPUS_MODEL`
  - [x] Add `TestBuildEnv_HaikuPriorityChain` table-driven tests covering:
    - `opts.HaikuModel` set → uses it (regardless of other fields)
    - `opts.HaikuModel` empty, `opts.SmallFastModel` set → uses SmallFastModel as haiku
    - `opts.HaikuModel` empty, `opts.SmallFastModel` empty, `ctx.HaikuModel` set → uses ctx.HaikuModel
    - `opts.HaikuModel` empty, `opts.SmallFastModel` empty, `ctx.HaikuModel` empty, `ctx.SmallFastModel` set → uses ctx.SmallFastModel as haiku
    - All empty → no `ANTHROPIC_DEFAULT_HAIKU_MODEL`
  - [x] Add `TestBuildEnv_SonnetOpusPriorityChain` table-driven tests covering sonnet and opus priority chains (opts > ctx > omit)
  - [x] Add `TestBuildEnv_SmallFastModelNotInjected` verifying `ANTHROPIC_SMALL_FAST_MODEL` is never in the output env, even when `ctx.SmallFastModel` is set
  - [x] Update `TestBuildEnv_PriorityChain`: remove existing SFM test cases (lines 119-142, now covered by `TestBuildEnv_HaikuPriorityChain`), keep Model priority cases (lines 95-118), add SonnetModel priority cases (`opts.SonnetModel > ctx.SonnetModel > omit`) and OpusModel priority cases (`opts.OpusModel > ctx.OpusModel > omit`)

- [x] Task 4: GREEN — update Options struct and buildEnv in runner.go (AC: #1-#6)
  - [x] In `internal/runner/runner.go:14-19`, add fields to Options struct:
    ```go
    HaikuModel  string
    SonnetModel string
    OpusModel   string
    ```
    Keep existing `SmallFastModel string` for backward compat (used by `--small-fast-model` CLI flag until Story 4.2 maps it to HaikuModel).
  - [x] Rewrite `buildEnv()` model injection section in `internal/runner/runner.go:93-109` (keep the env filtering prologue at lines 82-92 intact):
    ```go
    // Model: opts > config > omit
    model := opts.Model
    if model == "" {
        model = ctx.Model
    }
    if model != "" {
        filtered = append(filtered, "ANTHROPIC_MODEL="+model)
    }

    // Haiku: opts.HaikuModel > opts.SmallFastModel > ctx.HaikuModel > ctx.SmallFastModel > omit
    haiku := opts.HaikuModel
    if haiku == "" {
        haiku = opts.SmallFastModel
    }
    if haiku == "" {
        haiku = ctx.HaikuModel
    }
    if haiku == "" {
        haiku = ctx.SmallFastModel
    }
    if haiku != "" {
        filtered = append(filtered, "ANTHROPIC_DEFAULT_HAIKU_MODEL="+haiku)
    }

    // Sonnet: opts > config > omit
    sonnet := opts.SonnetModel
    if sonnet == "" {
        sonnet = ctx.SonnetModel
    }
    if sonnet != "" {
        filtered = append(filtered, "ANTHROPIC_DEFAULT_SONNET_MODEL="+sonnet)
    }

    // Opus: opts > config > omit
    opus := opts.OpusModel
    if opus == "" {
        opus = ctx.OpusModel
    }
    if opus != "" {
        filtered = append(filtered, "ANTHROPIC_DEFAULT_OPUS_MODEL="+opus)
    }
    ```
  - [x] Verify Task 3 tests pass

- [x] Task 5: Fix remaining test failures (AC: #7)
  - [x] Update mock scripts in `cmd/run_test.go` line ~217 and `cmd/exec_test.go` line ~96: replace `$ANTHROPIC_SMALL_FAST_MODEL` with `$ANTHROPIC_DEFAULT_HAIKU_MODEL` in the shell script strings
  - [x] Update assertion messages in `cmd/run_test.go` line ~233 and `cmd/exec_test.go` line ~138: change error messages referencing `ANTHROPIC_SMALL_FAST_MODEL` to `ANTHROPIC_DEFAULT_HAIKU_MODEL`
  - [x] Update test case name in `cmd/exec_test.go` line ~62: change `--small-fast-model flag sets ANTHROPIC_SMALL_FAST_MODEL via shell` to `--small-fast-model flag sets ANTHROPIC_DEFAULT_HAIKU_MODEL via shell`
  - [x] `go test ./...` — fix any remaining test that breaks
  - [x] `go vet ./...` — ensure clean
  - [x] `make build` — verify static binary builds

## Dev Notes

### What Changed and Why

Claude Code updated its model environment variables. `ANTHROPIC_SMALL_FAST_MODEL` is deprecated. New canonical variables: `ANTHROPIC_DEFAULT_HAIKU_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`, `ANTHROPIC_DEFAULT_OPUS_MODEL`.

**Critical behavioral change:** `ANTHROPIC_SMALL_FAST_MODEL` is **never** injected. All model-typed env vars now use `ANTHROPIC_DEFAULT_*_MODEL` format.

### Files to Change

| File | Change | Lines |
|------|--------|-------|
| `config/config.go:12-17` | Add HaikuModel, SonnetModel, OpusModel to Context struct | MODIFY |
| `config/config.go:69-76` | Update default config template with new field examples | MODIFY |
| `config/config.go:128-134` | Copy new fields in GetContext resolvedContext | MODIFY |
| `config/config_test.go` | Add TestGetContext_ModelFields | MODIFY |
| `internal/runner/runner.go:14-19` | Add HaikuModel, SonnetModel, OpusModel to Options struct | MODIFY |
| `internal/runner/runner.go:93-109` | Rewrite buildEnv model injection section (keep env filtering prologue at 82-92 intact) | MODIFY |
| `internal/runner/runner_test.go` | Update existing tests, add haiku/sonnet/opus priority chain tests | MODIFY |
| `cmd/run_test.go` | Update mock script to echo `$ANTHROPIC_DEFAULT_HAIKU_MODEL` instead of `$ANTHROPIC_SMALL_FAST_MODEL` | MODIFY |
| `cmd/exec_test.go` | Update mock script and test case names for new env var names | MODIFY |
| `examples/config.toml` | Add commented examples for new fields | MODIFY |

### Haiku Priority Chain (the tricky part)

This is the most complex part of the story. The haiku model has 4 fallback sources:

```
Priority: opts.HaikuModel → opts.SmallFastModel → ctx.HaikuModel → ctx.SmallFastModel → omit
```

- `opts.HaikuModel` — will be set by `--haiku-model` CLI flag (Story 4.2)
- `opts.SmallFastModel` — currently set by `--small-fast-model` CLI flag (Epic 3). In Story 4.2, this flag becomes an alias for `--haiku-model`
- `ctx.HaikuModel` — config field `haiku_model`
- `ctx.SmallFastModel` — config field `small_fast_model` (backward compat)

The `opts.SmallFastModel` must be checked **before** `ctx.HaikuModel` because CLI flags always take priority over config values.

### buildEnv Injection Order (new)

```
ANTHROPIC_BASE_URL (required)
→ ANTHROPIC_AUTH_TOKEN (required)
→ ANTHROPIC_MODEL (optional: opts > ctx > omit)
→ ANTHROPIC_DEFAULT_HAIKU_MODEL (optional: opts.HaikuModel > opts.SmallFastModel > ctx.HaikuModel > ctx.SmallFastModel > omit)
→ ANTHROPIC_DEFAULT_SONNET_MODEL (optional: opts > ctx > omit)
→ ANTHROPIC_DEFAULT_OPUS_MODEL (optional: opts > ctx > omit)
```

### Config TOML Field Mapping

| TOML field | Context field | mapstructure tag | Env var injected |
|-----------|---------------|------------------|-----------------|
| `model` | `Model` | `"model"` | `ANTHROPIC_MODEL` |
| `haiku_model` | `HaikuModel` | `"haiku_model"` | `ANTHROPIC_DEFAULT_HAIKU_MODEL` |
| `sonnet_model` | `SonnetModel` | `"sonnet_model"` | `ANTHROPIC_DEFAULT_SONNET_MODEL` |
| `opus_model` | `OpusModel` | `"opus_model"` | `ANTHROPIC_DEFAULT_OPUS_MODEL` |
| `small_fast_model` | `SmallFastModel` | `"small_fast_model"` | Mapped to `ANTHROPIC_DEFAULT_HAIKU_MODEL` (fallback) |

### default config.go Template Update

Add commented-out examples for new fields in the default config template at `config/config.go:69-76`:

```toml
# Optional: specify models explicitly
# model = "claude-sonnet-4-6"
# haiku_model = "claude-haiku-4-5-20251001"
# sonnet_model = "claude-sonnet-4-6"
# opus_model = "claude-opus-4-7"
# small_fast_model = "claude-haiku-4-5-20251001"  # deprecated: use haiku_model instead
```

### Existing Test Impact

**`TestBuildEnv`** in `runner_test.go:12-35` — currently asserts `ANTHROPIC_SMALL_FAST_MODEL=claude-haiku-4-5`. Must change to `ANTHROPIC_DEFAULT_HAIKU_MODEL=claude-haiku-4-5`.

**`TestBuildEnv_SkipsEmptyOptional`** in `runner_test.go:37-53` — currently checks no `ANTHROPIC_MODEL=` and no `ANTHROPIC_SMALL_FAST_MODEL=`. Add checks for no `ANTHROPIC_DEFAULT_HAIKU_MODEL=`, no `ANTHROPIC_DEFAULT_SONNET_MODEL=`, no `ANTHROPIC_DEFAULT_OPUS_MODEL=`.

**`TestBuildEnv_InjectionOrder`** in `runner_test.go:55-83` — currently verifies order: `ANTHROPIC_BASE_URL → ANTHROPIC_AUTH_TOKEN → ANTHROPIC_MODEL → ANTHROPIC_SMALL_FAST_MODEL`. Must update to: `ANTHROPIC_BASE_URL → ANTHROPIC_AUTH_TOKEN → ANTHROPIC_MODEL → ANTHROPIC_DEFAULT_HAIKU_MODEL → ANTHROPIC_DEFAULT_SONNET_MODEL → ANTHROPIC_DEFAULT_OPUS_MODEL`. Also need to update the test context to include SonnetModel and OpusModel.

**`TestBuildEnv_PriorityChain`** in `runner_test.go:85-184` — currently tests Model and SmallFastModel priority chains. SmallFastModel tests must be updated to verify `ANTHROPIC_DEFAULT_HAIKU_MODEL` instead of `ANTHROPIC_SMALL_FAST_MODEL`. Add new test cases for HaikuModel, SonnetModel, OpusModel priority chains.

**`cmd/run_test.go`** and `cmd/exec_test.go`** — integration tests that use env-capturing mocks may assert `ANTHROPIC_SMALL_FAST_MODEL`. These must be updated to check `ANTHROPIC_DEFAULT_HAIKU_MODEL` instead.

### Story 4.2 Preparation

This story adds the Options struct fields and buildEnv priority chain. Story 4.2 will:
- Add `--haiku-model`, `--sonnet-model`, `--opus-model` flags to `ExtractFlags` in `args.go`
- Make `--small-fast-model` an alias for `--haiku-model` (set `Options.HaikuModel` instead of `Options.SmallFastModel`)
- Update cmd/run.go and cmd/exec.go to pass new flag values in Options

No changes to args.go, cmd/run.go, or cmd/exec.go in this story — those are Story 4.2 scope.

### Anti-Patterns (Do NOT)

- Do NOT inject `ANTHROPIC_SMALL_FAST_MODEL` — it is deprecated, replaced by `ANTHROPIC_DEFAULT_HAIKU_MODEL`
- Do NOT remove `SmallFastModel` from Context struct or Options struct — backward compat
- Do NOT remove `--small-fast-model` flag handling from ExtractFlags — backward compat (Story 4.2 maps it to haiku)
- Do NOT modify args.go, cmd/run.go, or cmd/exec.go — those are Story 4.2 scope
- Do NOT call `os.Exit()` inside `internal/runner/` — breaks testability
- Do NOT inject empty env var values — check `!= ""` before appending
- Do NOT change the ANTHROPIC_BASE_URL or ANTHROPIC_AUTH_TOKEN injection — those are required and unchanged
- Do NOT add CLI flags for new model fields — that's Story 4.2

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic 4 Story 4.1] — story definition and AC
- [Source: _bmad-output/planning-artifacts/architecture.md#Environment Variable Construction] — injection order, model priority chains
- [Source: _bmad-output/planning-artifacts/architecture.md#AD1 Runner struct] — Options struct design with model fields
- [Source: _bmad-output/planning-artifacts/prd.md#FR2] — model field configuration requirements
- [Source: _bmad-output/planning-artifacts/prd.md#FR10] — env var injection requirements
- [Source: _bmad-output/planning-artifacts/prd.md#FR27-FR28] — CLI flag override requirements (Story 4.2)
- [Source: config/config.go:12-17] — Context struct (add HaikuModel, SonnetModel, OpusModel)
- [Source: config/config.go:69-76] — default config template (add new field examples)
- [Source: config/config.go:128-134] — GetContext resolvedContext copy (add new fields)
- [Source: internal/runner/runner.go:14-19] — Options struct (add HaikuModel, SonnetModel, OpusModel)
- [Source: internal/runner/runner.go:93-109] — buildEnv model injection section (keep env filtering prologue at 82-92 intact)
- [Source: internal/runner/runner_test.go:12-35] — TestBuildEnv (update assertions)
- [Source: internal/runner/runner_test.go:55-83] — TestBuildEnv_InjectionOrder (update prefix list)
- [Source: internal/runner/runner_test.go:85-184] — TestBuildEnv_PriorityChain (update for new fields)
- [Source: _bmad-output/implementation-artifacts/3-2-add-model-override-flags-to-run-and-exec.md] — previous story learnings
- [Source: _bmad-output/project-context.md] — implementation rules and patterns

## Dev Agent Record

### Agent Model Used

GLM-5.1

### Debug Log References

No blockers encountered during implementation.

### Completion Notes List

- All 5 tasks completed via red-green-refactor cycle
- Added HaikuModel, SonnetModel, OpusModel to config.Context struct with mapstructure tags
- Updated GetContext to copy new fields into resolvedContext
- Updated default config template with new model field examples and deprecated note for small_fast_model
- Added HaikuModel, SonnetModel, OpusModel to runner.Options struct
- Rewrote buildEnv model injection: ANTHROPIC_SMALL_FAST_MODEL replaced by ANTHROPIC_DEFAULT_HAIKU_MODEL via 4-level priority chain (opts.HaikuModel > opts.SmallFastModel > ctx.HaikuModel > ctx.SmallFastModel)
- Added ANTHROPIC_DEFAULT_SONNET_MODEL and ANTHROPIC_DEFAULT_OPUS_MODEL injection with 2-level priority (opts > ctx > omit)
- Updated cmd/run_test.go and cmd/exec_test.go mock scripts to capture ANTHROPIC_DEFAULT_HAIKU_MODEL instead of ANTHROPIC_SMALL_FAST_MODEL
- All tests pass: go test ./... clean, go vet ./... clean, make build succeeds

### File List

- config/config.go — Added HaikuModel, SonnetModel, OpusModel to Context struct; updated GetContext resolvedContext; updated default config template
- config/config_test.go — Added TestGetContext_ModelFields table-driven test
- internal/runner/runner.go — Added HaikuModel, SonnetModel, OpusModel to Options struct; rewrote buildEnv model injection section
- internal/runner/runner_test.go — Updated TestBuildEnv, TestBuildEnv_SkipsEmptyOptional, TestBuildEnv_InjectionOrder, TestBuildEnv_PriorityChain; added TestBuildEnv_HaikuPriorityChain, TestBuildEnv_SonnetOpusPriorityChain, TestBuildEnv_SmallFastModelNotInjected
- cmd/run_test.go — Updated mock script and assertion messages for ANTHROPIC_DEFAULT_HAIKU_MODEL
- cmd/exec_test.go — Updated mock script, test case name, and assertion messages for ANTHROPIC_DEFAULT_HAIKU_MODEL

### Review Findings

- [x] [Review][Patch] TestBuildEnv_SmallFastModelNotInjected 应同时验证 ANTHROPIC_DEFAULT_HAIKU_MODEL 被正确注入（目前只验证 ANTHROPIC_SMALL_FAST_MODEL 不注入，未验证值被正确提升）[internal/runner/runner_test.go]
- [x] [Review][Patch] TestBuildEnv_SonnetOpusPriorityChain 缺少 Sonnet 和 Opus 同时设置的组合测试用例 [internal/runner/runner_test.go]
- [x] [Review][Patch] TestBuildEnv 未验证 ANTHROPIC_DEFAULT_HAIKU_MODEL / _SONNET / _OPUS 被 prologue 从初始环境中剥离 [internal/runner/runner_test.go]
- [x] [Review][Patch] wantHaikuValue / wantSonnetValue / wantOpusValue 与 wantXxx 字段冗余（值始终为 "ENV_VAR_NAME=" + wantXxx），存在维护风险 [internal/runner/runner_test.go]
- [x] [Review][Defer] 集成测试 mock 脚本只捕获 ANTHROPIC_MODEL 和 ANTHROPIC_DEFAULT_HAIKU_MODEL，缺少 SONNET/OPUS 端到端覆盖 [cmd/run_test.go, cmd/exec_test.go] — deferred，Story 4.2 会添加 CLI flag 后再补充

## Change Log

- 2026-05-13: Story 4.1 implementation complete — config and runner model field expansion with ANTHROPIC_DEFAULT_*_MODEL env vars
