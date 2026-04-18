---
stepsCompleted: [1, 2, 3, 4]
inputDocuments: []
session_topic: '为 ccctx 添加 bash session 模式，实现环境变量在子进程中持久化'
session_goals: '设计通用、可分发、无需依赖 ccctx 的脚本运行方案'
selected_approach: 'ai-recommended'
techniques_used: ['Constraint Mapping', 'SCAMPER Method', 'First Principles Thinking']
ideas_generated: 17
technique_execution_complete: true
session_active: false
workflow_completed: true
context_file: ''
---

# Brainstorming Session Results

**Facilitator:** dsdashun
**Date:** 2026-04-17

## Session Overview

**Topic:** 为 ccctx 添加 bash session 模式，实现环境变量在子进程中持久化
**Goals:** 设计通用、可分发、无需依赖 ccctx 的脚本运行方案

### Session Setup

当前痛点：`ccctx run <provider>` 直接启动 Claude Code 进程，但内部 bash 子 shell 中环境变量未继承，导致 `claude -p` 调用失败。需要一种方式让 ccctx 启动预配置好环境变量的 bash session，使所有子进程都能继承正确的模型配置。

## Technique Selection

**Approach:** AI-Recommended Techniques
**Analysis Context:** 为 ccctx 添加 bash session 模式 with focus on 环境变量持久化与脚本可分发性

**Recommended Techniques:**

- **Constraint Mapping:** 系统识别真实约束 vs 想象约束，界定解决方案空间
- **SCAMPER Method:** 用 7 个透镜系统化改造现有 ccctx run 命令
- **First Principles Thinking:** 回到本质问题，发现全新实现路径

**AI Rationale:** 先用约束映射明确边界，再用 SCAMPER 系统探索改造方向，最后用第一性原理突破现有架构思维局限。三层递进从边界到本质。

## Technique 1: Constraint Mapping — Results

### Constraint Map

| Constraint | Type | Status |
|------|------|------|
| New `exec` subcommand, `--` followed by command to execute | Design Decision | ✅ Confirmed |
| No `--` defaults to launching `$SHELL` | Design Decision | ✅ Confirmed |
| `run` behavior unchanged | Hard Constraint | ✅ Confirmed |
| Provider config env vars override existing values | Design Decision | ✅ Confirmed |
| Child process exit code passthrough | Design Decision | ✅ Confirmed |
| All child processes inherit env vars | Hard Constraint | ✅ Guaranteed by exec |
| Interactive TUI selector when no provider specified | Feature Parity | ✅ Confirmed |

### Key Insight

Original assumption was "env vars not inherited" — testing revealed env vars ARE inherited by Bash tool, but `claude -p` hangs when called inside an existing Claude Code session. The real use case is running Ralph loop in a **separate terminal** outside Claude Code, with provider env vars pre-configured.

## Technique 2: SCAMPER Method — Results

- **S (Substitute):** No `.env` file, no removing `--` separator — keep it simple and unambiguous
- **C (Combine):** No multi-provider merge for `exec` — single provider only
- **A (Adapt):** Borrow from `env` command pattern; `-e` flag noted as low-priority future enhancement
- **M (Modify):** Env var handling logic from `run.go:157-178` can be directly reused in `exec`
- **P (Put to other uses):** `exec` naturally supports CI/CD, provider comparison testing, other AI CLI tools
- **E (Eliminate):** `exec` eliminates the `claude` binary dependency (`LookPath` not needed)
- **R (Reverse):** `eval "$(ccctx init)"` approach rejected — conflicts with distributability goal

## Technique 3: First Principles Thinking — Results

### Core Insight

ccctx is NOT "a tool that launches Claude Code" — it is a **provider config → env vars → exec** bridge.

`run` = `exec -- claude` (convenience syntax sugar for the most common case).

### Shared Kernel

```
resolveProvider(name) → config → envVars → exec(command)
```

| Step | `run` | `exec` |
|------|-------|--------|
| Parse provider | shared | shared |
| Read config | shared | shared |
| Build env vars | shared | shared |
| Determine target | hardcoded `claude` | `--` args or `$SHELL` |
| Execute | shared | shared |

### Parameter Override Design

Override config-level parameters (not raw env vars) — user-friendly:
- `--model <value>` → overrides provider config's model field
- Place flags **before** `--` separator: `ccctx exec [provider] [flags] -- [command]`
- Only `--model` for now; other overrides can be added later via shared kernel extension

## Final Design Summary

### New `exec` subcommand

```
ccctx exec <provider>                   # Launch $SHELL with provider env vars
ccctx exec <provider> -- bash           # Launch bash with provider env vars
ccctx exec <provider> -- bash script.sh # Run script with provider env vars
ccctx exec <provider> --model glm-4.7-flash -- bash  # Override model
ccctx exec                              # TUI selector → $SHELL (interactive)
```

### Architecture

- Shared kernel extracted from `run` for provider resolution, env var building, process execution
- `run` becomes thin wrapper calling shared kernel with `claude` as target
- `exec` calls shared kernel with user-specified command (or `$SHELL`)

### Decisions Log

| Decision | Rationale |
|----------|-----------|
| `--` required for exec command | Zero ambiguity |
| Default to `$SHELL` when no command | Frictionless UX |
| `--model` flag only for now | YAGNI — add others when needed |
| Single provider only | No real multi-provider use case |
| Exit code passthrough | Natural Unix behavior |
| Provider config overrides env vars | Provider takes priority |
| TUI selector for `exec` too | Feature parity with `run` |

## Idea Organization and Action Plan

### Thematic Organization

**Theme 1: Core Architecture**
- Shared kernel pattern: `resolveProvider → config → envVars → exec(command)`
- `run` refactored as thin wrapper over shared kernel
- Env var logic extracted from `run.go:157-178`

**Theme 2: CLI Interface Design**
- `exec` subcommand with `--` separator
- Default to `$SHELL` when no command specified
- TUI interactive selector support
- `--model` config-level flag for parameter override

**Theme 3: Extensibility**
- Shared kernel enables future config-level flags (`--base-url`, etc.)
- Natural CI/CD support without additional tooling
- Compatible with any AI CLI tool using `ANTHROPIC_*` env vars

### Priority Action Plan

| Priority | Task | Impact | Difficulty |
|----------|------|--------|------------|
| P0 | Extract shared kernel from `run.go` into `internal/runner/` | High (foundation) | Medium |
| P0 | Implement `exec` subcommand (`cmd/exec.go`) | High (core feature) | Low |
| P1 | TUI interactive selector for `exec` | Medium (feature parity) | Low (reuse) |
| P1 | `--model` flag override | High (Ralph loop scenario) | Low |
| P2 | Exit code passthrough | Medium (correctness) | Low |

### Implementation Steps

1. **Refactor:** Extract provider parsing, TUI selection, env var building, process execution from `run.go` into `internal/runner/`
2. **New command:** Add `cmd/exec.go`, register `exec` subcommand with Cobra
3. **Arg parsing:** `[provider] [--model value] -- [command args...]`
4. **Default behavior:** No `--` → use `os.Getenv("SHELL")` as command
5. **Testing:** Table-driven tests covering all argument combinations

## Session Summary

### Key Achievements

- Identified ccctx's true abstraction: provider config → env vars → exec
- Designed `exec` subcommand with clean separation from `run`
- Discovered config-level parameter override pattern (user-friendly alternative to `-e` env var flags)
- Validated architecture supports future extensibility without over-engineering

### Breakthrough Moments

- Testing `claude -p` inside Claude Code session revealed the real constraint (not env var inheritance)
- First Principles reframing: `run` = `exec -- claude` → shared kernel architecture
- User insight on parameter override: users think in config names, not env var names
