# Sprint Change Proposal: Claude Code Model Environment Variable Update

**Date:** 2026-05-13
**Trigger:** Epic 3 Retrospective — Significant Discovery
**Scope Classification:** Minor (direct implementation by Developer agent)
**Status:** Approved

---

## 1. Issue Summary

### Problem Statement

Claude Code has updated its model-related environment variable naming convention. `ANTHROPIC_SMALL_FAST_MODEL` is now marked as `[DEPRECATED]` in official Anthropic documentation. Three new variables have been introduced: `ANTHROPIC_DEFAULT_HAIKU_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`, and `ANTHROPIC_DEFAULT_OPUS_MODEL`.

ccctx currently only supports `ANTHROPIC_MODEL` and `ANTHROPIC_SMALL_FAST_MODEL`, meaning users cannot configure haiku, sonnet, or opus model overrides through ccctx.

### Discovery Context

Identified during Epic 3 retrospective (2026-05-13) as a Significant Discovery. Confirmed via official Anthropic documentation at https://docs.anthropic.com/en/docs/claude-code/settings.

### Evidence

Official documentation environment variable table shows:
- `ANTHROPIC_SMALL_FAST_MODEL` → `[DEPRECATED] Name of Haiku-class model for background tasks`
- `ANTHROPIC_DEFAULT_HAIKU_MODEL` → See Model configuration
- `ANTHROPIC_DEFAULT_SONNET_MODEL` → See Model configuration
- `ANTHROPIC_DEFAULT_OPUS_MODEL` → See Model configuration

---

## 2. Impact Analysis

### Epic Impact

| Epic | Impact |
|------|--------|
| Epic 1 (done) | No impact |
| Epic 2 (done) | No impact |
| Epic 3 (done) | No impact — completed before discovery |
| Epic 4 → renumbered to Epic 5 | Priority lowered — deferred tech debt items remain valid but not urgent |
| **New Epic 4** | **Required** — environment variable update for user-facing correctness |

### Story Impact

No in-progress or completed stories are affected. All changes are additive.

### Artifact Conflicts

| Artifact | Impact | Severity |
|----------|--------|----------|
| PRD (FR2, FR10, FR27, FR28, Config Schema) | Update required | Medium — documentation updates |
| Architecture (AD1 Options struct, buildEnv, ExtractFlags, Config struct) | Update required | Medium — design doc updates |
| Epics | New Epic 4 inserted, old Epic 4 renumbered to Epic 5 | Medium — planning update |
| `config/config.go` | Context struct gains 3 fields; GetContext passes them through | Low — additive change |
| `internal/runner/runner.go` | Options struct gains 3 fields; buildEnv gains 3 injection blocks | Low — additive change |
| `internal/runner/args.go` | ExtractFlags gains 4 new flag cases; signature changes | Medium — breaking API change within internal package |
| Tests | runner_test.go, args_test.go, config_test.go need new test cases | Low — additive |
| `examples/config.toml` | Update comments to show new fields | Low |
| `project-context.md` | Update model-related rules | Low |

### Technical Impact

- **No breaking changes for existing users** — `small_fast_model` config field and `--small-fast-model` flag continue to work, mapping to `ANTHROPIC_DEFAULT_HAIKU_MODEL`
- **`ANTHROPIC_SMALL_FAST_MODEL` is no longer injected** — when `ANTHROPIC_DEFAULT_HAIKU_MODEL` is present, the deprecated variable is redundant
- **Config backward compatibility** — old config files with `small_fast_model` still work; the field maps to haiku via priority chain

---

## 3. Recommended Approach

**Selected path: Direct Adjustment (Option 1)**

Insert a new Epic 4 before the existing tech debt epic (renumbered to Epic 5). The change is purely additive — no rollbacks, no scope reduction needed.

### Rationale

- Implementation effort: **Medium** — extending existing patterns (struct fields, flag extraction, env injection)
- Technical risk: **Low** — no architectural changes, just field additions
- Timeline impact: **Minimal** — 2 stories, solo developer, no external dependencies
- User impact: **High** — without this update, users cannot configure haiku/sonnet/opus models

### Key Design Decisions

1. **`ANTHROPIC_SMALL_FAST_MODEL` is NOT injected** — the new `ANTHROPIC_DEFAULT_HAIKU_MODEL` supersedes it; setting both is useless
2. **`small_fast_model` config field retained** — backward compatible; maps to haiku via priority chain `HaikuModel > SmallFastModel`
3. **`--small-fast-model` is an alias for `--haiku-model`** — handled in ExtractFlags, downstream code unaware
4. **When both `--haiku-model` and `--small-fast-model` are used** — `--haiku-model` takes priority, `--small-fast-model` is silently ignored

---

## 4. Detailed Change Proposals

### 4.1 PRD Changes

**FR2 (Context Configuration):**
> Users can optionally specify `model`, `haiku_model`, `sonnet_model`, and `opus_model` per context to override default model selection. The `small_fast_model` field is retained for backward compatibility — when set, it maps to `ANTHROPIC_DEFAULT_HAIKU_MODEL` (the deprecated `ANTHROPIC_SMALL_FAST_MODEL` is no longer injected).

**FR10 (Environment Variable Injection):**
> The system injects the provider's `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, and optional model env vars (`ANTHROPIC_MODEL`, `ANTHROPIC_DEFAULT_HAIKU_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`, `ANTHROPIC_DEFAULT_OPUS_MODEL`) into the Claude process.

**FR27 (CLI Flags):**
> Users can override the provider's configured model via `--model`, `--haiku-model`, `--sonnet-model`, and `--opus-model` flags on both `run` and `exec` commands. The `--small-fast-model` flag is retained as an alias for `--haiku-model` for backward compatibility.

**FR28 (Priority):**
> For each model variable (`model`, `haiku_model`, `sonnet_model`, `opus_model`), the system applies CLI flag values with higher priority than config file values, allowing temporary overrides without modifying the config. When `--small-fast-model` is used, it sets `ANTHROPIC_DEFAULT_HAIKU_MODEL` (equivalent to `--haiku-model`).

**Config Schema:**
```toml
[context.<name>]
base_url = "https://..."           # Required
auth_token = "token" or "env:VAR"  # Required, supports env: prefix
model = "claude-..."               # Optional — sets ANTHROPIC_MODEL
haiku_model = "claude-..."         # Optional — sets ANTHROPIC_DEFAULT_HAIKU_MODEL
sonnet_model = "claude-..."        # Optional — sets ANTHROPIC_DEFAULT_SONNET_MODEL
opus_model = "claude-..."          # Optional — sets ANTHROPIC_DEFAULT_OPUS_MODEL
small_fast_model = "claude-..."    # Optional (deprecated) — alias for haiku_model, sets ANTHROPIC_DEFAULT_HAIKU_MODEL
```

### 4.2 Architecture Changes

**Options struct:**
```go
type Options struct {
    ContextName string   // resolved provider name
    Target      []string // command to execute
    Model       string   // optional override — sets ANTHROPIC_MODEL
    HaikuModel  string   // optional override — sets ANTHROPIC_DEFAULT_HAIKU_MODEL
    SonnetModel string   // optional override — sets ANTHROPIC_DEFAULT_SONNET_MODEL
    OpusModel   string   // optional override — sets ANTHROPIC_DEFAULT_OPUS_MODEL
}
```

**Config Context struct:**
```go
type Context struct {
    BaseURL        string `mapstructure:"base_url"`
    AuthToken      string `mapstructure:"auth_token"`
    Model          string `mapstructure:"model"`
    HaikuModel     string `mapstructure:"haiku_model"`
    SonnetModel    string `mapstructure:"sonnet_model"`
    OpusModel      string `mapstructure:"opus_model"`
    SmallFastModel string `mapstructure:"small_fast_model"` // deprecated, maps to haiku in buildEnv
}
```

**Model priority chain:**
```
ANTHROPIC_MODEL: Options.Model > ctx.Model > omit
ANTHROPIC_DEFAULT_HAIKU_MODEL: Options.HaikuModel > ctx.HaikuModel > ctx.SmallFastModel > omit
ANTHROPIC_DEFAULT_SONNET_MODEL: Options.SonnetModel > ctx.SonnetModel > omit
ANTHROPIC_DEFAULT_OPUS_MODEL: Options.OpusModel > ctx.OpusModel > omit
```

**Injection order:**
1. `ANTHROPIC_BASE_URL` — required
2. `ANTHROPIC_AUTH_TOKEN` — required
3. `ANTHROPIC_MODEL` — optional
4. `ANTHROPIC_DEFAULT_HAIKU_MODEL` — optional
5. `ANTHROPIC_DEFAULT_SONNET_MODEL` — optional
6. `ANTHROPIC_DEFAULT_OPUS_MODEL` — optional

**ExtractFlags signature:**
```go
func ExtractFlags(args []string) (model, haikuModel, sonnetModel, opusModel string, remaining []string, err error)
```

**Flag mapping:**

| Flag | Maps to |
|------|---------|
| `--model` | `model` |
| `--haiku-model` | `haikuModel` |
| `--sonnet-model` | `sonnetModel` |
| `--opus-model` | `opusModel` |
| `--small-fast-model` | `haikuModel` (alias, silently ignored if `--haiku-model` also present) |

### 4.3 Epics Changes

**New Epic 4: Claude Code Model Environment Variable Update**

- Story 4.1: Config & Runner Model Field Expansion
- Story 4.2: CLI Flag Extension & Backward Compatibility
- Epic 4 Retrospective: optional

**Old Epic 4 → Epic 5: Tech Debt & Code Quality** (content unchanged, renumbered)

- Story 5.1 (was 4-1): Config Loading Robustness
- Story 5.2 (was 4-2): TUI Selector Quality Polish
- Epic 5 Retrospective: optional

---

## 5. Implementation Handoff

### Scope Classification: Minor

Direct implementation by Developer agent. No backlog reorganization or strategic replan needed.

### Handoff Plan

**Responsible agent:** Developer (Amelia)

**Execution sequence:**
1. Update `config/config.go` — add 3 new fields to Context struct, update GetContext
2. Update `internal/runner/runner.go` — add 3 new fields to Options struct, update buildEnv with new priority chains
3. Update `internal/runner/args.go` — extend ExtractFlags signature and flag extraction logic
4. Update `cmd/run.go` and `cmd/exec.go` — adapt to new ExtractFlags signature
5. Update `examples/config.toml` — add new field examples
6. Update tests — runner_test.go, args_test.go, config_test.go
7. Update `project-context.md` — revise model-related rules

**Prerequisite:** Run `/bmad-sprint-planning` to generate sprint plan for new Epic 4 before starting implementation.

**Success criteria:**
- `ccctx run provider --haiku-model X` injects `ANTHROPIC_DEFAULT_HAIKU_MODEL=X`
- `ccctx run provider --sonnet-model X` injects `ANTHROPIC_DEFAULT_SONNET_MODEL=X`
- `ccctx run provider --opus-model X` injects `ANTHROPIC_DEFAULT_OPUS_MODEL=X`
- `ccctx run provider --small-fast-model X` injects `ANTHROPIC_DEFAULT_HAIKU_MODEL=X` (NOT `ANTHROPIC_SMALL_FAST_MODEL`)
- Config with `haiku_model` set → `ANTHROPIC_DEFAULT_HAIKU_MODEL` injected
- Config with only `small_fast_model` set → `ANTHROPIC_DEFAULT_HAIKU_MODEL` injected
- Config with both `haiku_model` and `small_fast_model` → `haiku_model` takes priority
- `ANTHROPIC_SMALL_FAST_MODEL` is never injected under any circumstances
- All existing tests pass with zero regressions
- `go vet ./...` clean
