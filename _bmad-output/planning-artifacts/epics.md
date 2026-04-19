---
stepsCompleted:
  - step-01-validate-prerequisites
  - step-02-design-epics
  - step-03-create-stories
  - step-04-final-validation
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/planning-artifacts/architecture.md
---

# ccctx - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for ccctx, decomposing the requirements from the PRD and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: Users can define multiple named provider contexts in a TOML configuration file, each with `base_url` and `auth_token` fields
FR2: Users can optionally specify `model` and `small_fast_model` per context to override default model selection
FR3: Users can reference environment variables in `auth_token` using the `env:` prefix for secure token resolution
FR4: The system auto-creates the config directory and example configuration file when none exists
FR5: Users can override the default config path (`~/.ccctx/config.toml`) via the `CCCTX_CONFIG_PATH` environment variable
FR6: Users can list all configured context names via `ccctx list`
FR7: Users can view context names output to stdout, one per line
FR8: Users can run Claude Code with a specified provider context via `ccctx run <provider>`
FR9: Users can forward additional arguments to Claude Code using the `--` separator
FR10: The system injects the provider's `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, and optional model env vars into the Claude process
FR11: The system strips any pre-existing `ANTHROPIC_*` environment variables before injecting provider-specific values
FR12: The system propagates Claude's exit code back to the caller
FR13: Users can launch an interactive shell session with provider env vars via `ccctx exec <provider>`
FR14: Users can execute an arbitrary command with provider env vars via `ccctx exec <provider> -- <command> [args...]`
FR15: The system defaults to `$SHELL` when no command is specified after the provider name
FR16: All child processes of the launched shell or command inherit the provider environment variables
FR17: The system propagates the child process exit code back to the caller
FR18: The system overrides any existing `ANTHROPIC_*` environment variables with the provider's values
FR19: Users can launch an interactive context selector via `ccctx exec` (no provider specified)
FR20: Users can navigate the context list using arrow keys (up/down) and vim keys (j/k)
FR21: Users can select a context by pressing Enter or clicking on an item
FR22: Users can cancel the selection by pressing ESC
FR23: The TUI selector is available in both `run` and `exec` commands when no provider is specified
FR24: The `run` and `exec` commands share a common provider resolution, env var construction, and process execution pipeline
FR25: The `run` command uses the shared pipeline with `claude` as the hardcoded execution target
FR26: The `exec` command uses the shared pipeline with a user-specified command or `$SHELL` as the execution target
FR27: Users can override the provider's configured model via `--model <value>` flag on both `run` and `exec` commands
FR28: The system applies CLI flag values with higher priority than config file values, allowing temporary overrides without modifying the config

### NonFunctional Requirements

NFR1: Auth tokens are never written to stdout, stderr, or log output — resolved internally only
NFR2: Existing `ANTHROPIC_*` environment variables are stripped before injecting provider-specific values, preventing credential leakage across contexts
NFR3: The `env:` prefix resolution reads environment variables at runtime only — tokens are never persisted in expanded form
NFR4: Config file permissions should be user-readable only (mode 0600) to protect stored credentials
NFR5: ccctx works with any POSIX-compliant shell (`bash`, `zsh`, `sh`, etc.) as the `$SHELL` default
NFR6: ccctx works with any AI CLI tool that reads `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN` environment variables, not just Claude Code
NFR7: The tool builds as a static binary (`CGO_ENABLED=0`) with zero runtime dependencies

### Additional Requirements

- Shared kernel extraction from `run.go` into `internal/runner/` using Runner struct pattern (`New()`, `Run()`, `buildEnv()`, `ParseArgs()`)
- `exec` is the core command; `run` is an ultra-thin wrapper (run ≡ exec -- claude)
- Runner struct carries state between provider resolution and env var construction (Phase 2 ready)
- TUI selector reusable across both `run` and `exec` commands
- Table-driven test pattern using testify framework for all new tests
- Exit code passthrough via `Run()` returning `(int, error)` — start failures distinguishable from non-zero exits
- No `LookPath` in `exec` — command is user-specified directly; `LookPath("claude")` stays in `run.go` only
- Environment filtering uses `strings.HasPrefix("ANTHROPIC_")` instead of hardcoded length checks
- `ParseArgs()` is command-agnostic — default target resolved at command level, not inside ParseArgs
- All existing tests must pass with zero regressions after refactoring
- Static binary build maintained (CGO_ENABLED=0)

### Breaking Changes (from Architecture)

These intentional behavior changes are documented in the architecture and must be reflected in story acceptance criteria:

1. **Unknown first arg hard-errors instead of TUI fallback.** `ccctx run nonexistent` → error instead of launching TUI. ParseArgs always treats first arg as provider name.
2. **TUI cancellation exit code 0 → 1.** ESC cancellation is an abort, not a success. `os.Exit(1)`.
3. **"No contexts found" → stderr + exit 1.** No configured contexts is an error condition. Output: `Error: no contexts found\n` to stderr.

### UX Design Requirements

No standalone UX design document — this is a CLI tool.

### Implementation Status

#### Already Implemented (FR1-FR12, FR20-FR22, NFR1-NFR3, NFR5-NFR7)

These requirements are satisfied by the existing codebase and require no additional work:

- FR1-FR7: Context configuration and discovery (config.go, cmd/list.go)
- FR8-FR12: `run` command with full argument parsing, env var management, exit code passthrough (cmd/run.go)
- FR20-FR22: TUI selector navigation and selection (internal/ui/selector.go)
- NFR1-NFR3, NFR5-NFR7: Security, compatibility, static build

#### Remaining Work (FR13-FR19, FR23-FR28, NFR4)

- FR13-FR18: `exec` subcommand — new command for shell/command execution with provider context
- FR19, FR23: TUI selector sharing with `exec`
- FR24-FR26: Shared execution kernel extraction
- FR27-FR28: `--model` CLI override flag
- NFR4: Config file permissions (0600)

### FR Coverage Map

| FR | Epic | Status | Brief Description |
|----|------|--------|-------------------|
| FR1 | — | ✅ Done | TOML multi-context config |
| FR2 | — | ✅ Done | Optional model/small_fast_model |
| FR3 | — | ✅ Done | env: prefix resolution |
| FR4 | — | ✅ Done | Auto-create config dir & file |
| FR5 | — | ✅ Done | CCCTX_CONFIG_PATH override |
| FR6 | — | ✅ Done | ccctx list command |
| FR7 | — | ✅ Done | Context names to stdout |
| FR8 | — | ✅ Done | ccctx run <provider> |
| FR9 | — | ✅ Done | -- separator forwarding |
| FR10 | — | ✅ Done | Inject ANTHROPIC_* env vars |
| FR11 | — | ✅ Done | Strip existing ANTHROPIC_* vars |
| FR12 | — | ✅ Done | Exit code passthrough |
| FR13 | Epic 1 | ❌ New | exec launches shell with provider env |
| FR14 | Epic 1 | ❌ New | exec runs arbitrary command |
| FR15 | Epic 1 | ❌ New | exec defaults to $SHELL |
| FR16 | Epic 1 | ❌ New | Child processes inherit env vars |
| FR17 | Epic 1 | ❌ New | exec exit code passthrough |
| FR18 | Epic 1 | ❌ New | exec overrides ANTHROPIC_* vars |
| FR19 | Epic 2 | ❌ New | exec no-arg triggers TUI |
| FR20 | — | ✅ Done | Arrow/vim key navigation |
| FR21 | — | ✅ Done | Enter/click selection |
| FR22 | — | ✅ Done | ESC cancellation |
| FR23 | Epic 2 | ❌ New | TUI shared between run and exec |
| FR24 | Epic 1 | ❌ New | Shared execution pipeline |
| FR25 | Epic 1 | ❌ New | run as thin wrapper |
| FR26 | Epic 1 | ❌ New | exec uses shared pipeline |
| FR27 | Epic 3 | ❌ New | --model CLI flag |
| FR28 | Epic 3 | ❌ New | CLI flag priority over config |

**Epic 3 also includes Story 3.1 (Runner & Command Layer Robustness)** — deferred tech debt that is a prerequisite for the flag implementation.

| NFR | Status | Notes |
|-----|--------|-------|
| NFR1 | ✅ Done | Tokens not leaked |
| NFR2 | ✅ Done | ANTHROPIC_* stripped |
| NFR3 | ✅ Done | Runtime-only resolution |
| NFR4 | Epic 1 | ❌ Current perms are 0644 |
| NFR5 | ✅ Done | POSIX shell compatible |
| NFR6 | ✅ Done | Multi-tool compatible |
| NFR7 | ✅ Done | Static binary (CGO_ENABLED=0) |

## Epic List

### Epic 1: Shared Execution Kernel & Flexible Provider Execution
Extract shared execution logic from run.go into `internal/runner/` using the Runner struct pattern, refactor run as an ultra-thin wrapper, and build the new exec subcommand (the core command). Users can run any command or launch a shell with provider context — not limited to Claude Code.
**FRs covered:** FR13, FR14, FR15, FR16, FR17, FR18, FR24, FR25, FR26
**NFRs covered:** NFR4
**Architecture decisions:** AD1 (Runner struct), AD2 (HasPrefix filtering), AD4 (config permissions), AD5 (testing)

### Epic 2: Interactive Provider Selection for All Commands
Share the existing TUI selector between run and exec commands. Users can interactively choose providers in either command without memorizing names.
**FRs covered:** FR19, FR23

### Epic 3: Model Override Flags (Phase 2)
Add --model and --small-fast-model CLI flags to both run and exec, allowing temporary overrides without editing config. Known conflict: registering these as Cobra flags will consume the `--` separator — must be resolved in Phase 2 design.
**FRs covered:** FR27, FR28

---

## Epic 1: Shared Execution Kernel & Flexible Provider Execution

Extract shared execution logic from run.go into `internal/runner/` using the Runner struct pattern, refactor run as an ultra-thin wrapper, and build the new exec subcommand (the core command). Users can run any command or launch a shell with provider context — not limited to Claude Code.

### Story 1.1: Extract Runner Struct and ParseArgs into internal/runner/

As a **developer**,
I want **run.go 中的环境变量构建和进程执行逻辑被提取到 internal/runner/ 包，使用 Runner struct 模式（New/Run/buildEnv）和命令无关的 ParseArgs 函数**,
So that **run 和 exec 命令可以共享同一套执行管道，且 Runner struct 天然支持 Phase 2 的 --model 状态传递**.

**Acceptance Criteria:**

**Given** run.go 当前包含环境变量构建和进程执行等全部逻辑
**When** 创建 `internal/runner/` 包，包含：
- `runner.go` — `Options` struct, `Runner` struct, `New(opts Options) (*Runner, error)`, `(*Runner) Run() (int, error)`, unexported `buildEnv()`
- `args.go` — `ParseArgs(args []string) (provider string, targetArgs []string, useTUI bool, err error)`
**Then** Runner struct 从 `run.go` 提取以下逻辑：
- env var building → `buildEnv()` — 使用 `strings.HasPrefix("ANTHROPIC_")` 过滤（AD2）
- process execution → `Run()` — 返回 `(exitCode int, err error)`，区分正常退出和启动失败
- argument parsing → `ParseArgs()` — 命令无关，default target 在 command 层决定
**And** `New()` 验证 `base_url` 和 `auth_token` 非空（空值返回 error）
**And** `buildEnv()` 注入顺序：ANTHROPIC_BASE_URL（必填）→ ANTHROPIC_AUTH_TOKEN（必填）→ ANTHROPIC_MODEL（可选）→ ANTHROPIC_SMALL_FAST_MODEL（可选），空值不注入
**And** Runner 包无 I/O 副作用（无 os.Exit，无 stderr 输出）
**And** 所有现有测试 `go test ./...` 通过，行为无变化
**And** FR24, FR25 得到满足

**Implementation file mapping (from architecture):**

| Current (run.go) | New location | Function |
|---|---|---|
| Lines 154-178 (env building) | `internal/runner/runner.go` → `buildEnv()` | ANTHROPIC_* env var construction |
| Lines 180-196 (exec + exit code) | `internal/runner/runner.go` → `Run()` | Process execution with exit code + error passthrough |
| Lines 22-137 (arg parsing) | `internal/runner/args.go` → `ParseArgs()` | Command-agnostic argument parsing |
| Lines 147-151 (LookPath) | `cmd/run.go` only | Binary discovery (exec doesn't use it) |

### Story 1.2: Refactor run Command as Ultra-Thin Wrapper

As a **developer**,
I want **run 命令重构为调用 internal/runner/ 的超薄包装器（run ≡ exec -- claude）**,
So that **run 的核心行为不变，但代码量从 ~200 行降至 ~30 行，所有业务逻辑在 runner 包中**.

**Acceptance Criteria:**

**Given** internal/runner/ 包已提供 Runner struct 和 ParseArgs
**When** 重构 cmd/run.go：
- 调用 `runner.ParseArgs(args)` 解析参数
- `useTUI=true` 时调用 `ui.RunContextSelector()`（含 breaking changes）
- 调用 `exec.LookPath("claude")` 查找 claude 二进制（runner 不含此逻辑）
- 构建 `runner.Options{ContextName: provider, Target: []string{claudePath} + targetArgs}`
- 调用 `runner.New(opts)` + `r.Run()` 获取退出码
**Then** `ccctx run <provider>` 行为与重构前完全一致
**And** `ccctx run <provider> -- --help` 参数转发正常
**And** `ccctx run`（无参数）TUI 选择器正常工作

**Breaking changes from architecture:**
- `ccctx run nonexistent` → hard error `context 'nonexistent' not found`（不再 fallback 到 TUI）
- TUI 取消（ESC）退出码从 0 改为 1
- "No contexts found" → stderr `Error: no contexts found` + exit 1

**And** `go test ./...` 通过，FR25 得到满足

### Story 1.3: Implement exec Subcommand (Core Command)

As a **ccctx 用户**,
I want **通过 `ccctx exec <provider>` 启动带 provider 环境变量的 shell，或通过 `ccctx exec <provider> -- <command>` 执行任意命令**,
So that **我可以在任何 AI CLI 工具或自动化脚本中使用 provider 上下文，不再局限于 Claude Code**.

**Acceptance Criteria:**

**Given** 一个已配置的 provider（如 provider-A）
**When** 运行 `ccctx exec provider-A`
**Then** 启动 `$SHELL`，shell 中包含 provider-A 的 ANTHROPIC_* 环境变量
**And** 子进程继承所有 provider 环境变量

**Given** 一个已配置的 provider 和一个命令
**When** 运行 `ccctx exec provider-A -- env | grep ANTHROPIC`
**Then** 输出正确的 ANTHROPIC_BASE_URL 和 ANTHROPIC_AUTH_TOKEN 值
**And** 命令的退出码被正确透传

**Given** 已有 ANTHROPIC_* 环境变量的 shell
**When** 运行 `ccctx exec provider-A -- some-command`
**Then** 已有的 ANTHROPIC_* 变量被 provider-A 的值覆盖

**Given** exec 子命令参数解析
**When** 运行 `ccctx exec provider-A -- bash -c "echo hello"`
**Then** `--` 后的完整命令（含参数）被正确传递执行

**Given** $SHELL 环境变量未设置
**When** 运行 `ccctx exec provider-A`（无 -- 分隔符，target 默认为 $SHELL）
**Then** 输出 `Error: SHELL environment variable not set` 到 stderr，退出码 1

**And** FR13, FR14, FR15, FR16, FR17, FR18, FR26 得到满足
**And** 不使用 LookPath，命令由用户直接指定
**And** exec 注册到 main.go（`rootCmd.AddCommand(cmd.ExecCmd)`）
**And** 新增 `internal/runner/runner_test.go` 和 `internal/runner/args_test.go`（testify 表驱动测试）
**And** 错误格式：`Error: <lowercase description without trailing period>`

### Story 1.4: Fix Config File Permissions

As a **ccctx 用户**,
I want **配置文件权限为 0600（仅用户可读）**,
So that **存储在配置文件中的 auth token 和凭据不会被其他用户读取**.

**Acceptance Criteria:**

**Given** ccctx 首次运行，配置目录和文件不存在
**When** 自动创建 config.toml 文件
**Then** 文件权限为 0600（仅所有者可读写）
**And** `config.go` 中 `os.WriteFile` 的权限参数从 0644 改为 0600

**Given** 已存在的配置文件
**When** ccctx 加载配置
**Then** 不修改已有文件的权限（避免影响用户自定义设置）

**And** NFR4 得到满足

---

## Epic 2: Interactive Provider Selection for All Commands

Share the existing TUI selector between run and exec commands. Users can interactively choose providers in either command without memorizing names.

### Story 2.1: exec Command Integrates TUI Interactive Selector

As a **ccctx 用户**,
I want **运行 `ccctx exec`（不带 provider 参数）时弹出交互式 TUI 选择器**,
So that **我不需要记忆 provider 名称，可以快速选择并直接进入配置好的 shell 环境**.

**Acceptance Criteria:**

**Given** 配置文件中有多个 provider（如 provider-A, provider-B）
**When** 运行 `ccctx exec`（无参数）
**Then** `ParseArgs` 返回 `useTUI=true`，显示 TUI 交互式选择器，列出所有可用 provider

**Given** TUI 选择器已显示
**When** 用方向键或 j/k 选择 provider-B 并按 Enter
**Then** 启动 `$SHELL`，shell 中包含 provider-B 的 ANTHROPIC_* 环境变量

**Given** TUI 选择器已显示
**When** 按 ESC 键
**Then** 输出 `Operation cancelled.` 到 stdout，退出码为 1（breaking change: 原为 0）

**Given** `ccctx run`（无参数）
**When** TUI 选择器显示
**Then** 行为与现有完全一致（无回归，除退出码变更）

**Given** 配置文件中没有任何 provider
**When** 运行 `ccctx exec`（无参数）
**Then** 输出 `Error: no contexts found` 到 stderr，退出码 1

**And** FR19, FR23 得到满足
**And** TUI 选择器在 run 和 exec 间共享同一 `internal/ui/selector.go` 实现
**And** TUI 调用仅在 `cmd/*.go` 中，不在 runner 包中（runner 不导入 internal/ui）

---

## Epic 3: Model Override Flags (Phase 2)

Add --model and --small-fast-model CLI flags to both run and exec, allowing temporary overrides without editing config. **Known conflict:** registering these as Cobra flags will cause Cobra to consume the `--` separator, breaking `ParseArgs`. Phase 2 design must resolve this (likely via `cmd.DisableFlagParsing = true` with manual flag parsing, or restricting `--model` to appear only before the provider name).

### Story 3.1: Runner & Command Layer Robustness

As a **developer**,
I want **runner 包和 cmd 包的健壮性问题得到修复，包括输入验证、错误处理、flag 解析和测试覆盖**,
So that **后续 Story 3.2 的 --model flag 添加能在稳固的基础上进行，且 ParseArgs 的 flag 处理方式不会与 Cobra flag 冲突**.

**Deferred items addressed:**
- runner.New does not validate BaseURL format [internal/runner/runner.go:31-36]
- runner.Run double-prints non-ExitError failures [internal/runner/runner.go:51-57, cmd/exec.go:62-66]
- ParseArgs treats flags before `--` as context names [internal/runner/args.go:30-34] — **critical prerequisite for Story 3.2**
- os.Exit calls bypass future defer cleanup [cmd/exec.go]
- cmd/run.go has no automated tests [cmd/run.go]

**Acceptance Criteria:**

**Given** runner.New 接收一个 URL 格式的 BaseURL
**When** BaseURL 不是合法 URL（如缺少 scheme、包含空格）
**Then** 返回明确的验证错误，不创建 Runner

**Given** runner.Run 执行的子进程因信号终止（非 ExitError）
**When** Run 返回错误
**Then** 错误消息只由调用方（cmd/*.go）打印一次，不会双重输出

**Given** ParseArgs 收到 `--model` 等类 flag 参数（以 `-` 开头）
**When** 参数出现在 provider 位置（`--` 之前）
**Then** 明确拒绝或按设计处理，为 Story 3.2 的 `--model` flag 预留正确行为

**Given** cmd/exec.go 中的错误处理路径
**When** 发生错误需要退出
**Then** 使用 `return` + 顶层 os.Exit 而非直接 os.Exit，确保 defer 正常执行

**Given** cmd/run.go 的核心路径
**When** 运行测试
**Then** 表驱动测试覆盖：成功路径、context 不存在、TUI 取消、LookPath 失败

### Story 3.2: Add --model Override Flags to run and exec

As a **ccctx 用户**,
I want **通过 `--model <value>` 和 `--small-fast-model <value>` 标志临时覆盖配置中的模型设置**,
So that **我可以在不修改配置文件的情况下快速切换模型，适应不同的使用场景**.

**Acceptance Criteria:**

**Given** provider-A 配置中 model 为 `claude-sonnet-4-6`
**When** 运行 `ccctx run provider-A --model claude-opus-4-7`
**Then** Claude 进程的 `ANTHROPIC_MODEL` 环境变量为 `claude-opus-4-7`

**Given** provider-A 配置中未设置 model
**When** 运行 `ccctx exec provider-A --model claude-opus-4-7 -- env | grep ANTHROPIC_MODEL`
**Then** 输出 `ANTHROPIC_MODEL=claude-opus-4-7`

**Given** provider-A 配置中 small_fast_model 为 `claude-haiku-4-5`
**When** 运行 `ccctx run provider-A --small-fast-model claude-sonnet-4-6`
**Then** `ANTHROPIC_SMALL_FAST_MODEL` 环境变量为 `claude-sonnet-4-6`

**Given** 同时指定 CLI 标志和配置文件中的模型设置
**When** 运行命令
**Then** CLI 标志值优先于配置文件值（Options.Model > ctx.Model > omit）

**Given** 未指定 --model 或 --small-fast-model 标志
**When** 运行命令
**Then** 使用配置文件中的模型设置（行为不变）

**And** FR27, FR28 得到满足
**And** --model 和 --small-fast-model 标志在 run 和 exec 命令中均可用
**And** `Options` struct 中的 `Model` 和 `SmallFastModel` 字段已在 Epic 1 中创建，Phase 2 只需连接 CLI flag → Options
**And** 新增表驱动测试覆盖标志优先级逻辑

---

## Epic 4: Tech Debt & Code Quality

Address deferred issues from code reviews across Epic 1 and Epic 2. These items were identified during reviews but deferred as pre-existing or non-blocking. Grouped by package boundary to minimize context-switching.

### Story 4.1: Config Loading Robustness

As a **developer**,
I want **config 包的加载逻辑更健壮，能正确处理各种文件系统异常和边界情况**,
So that **用户不会在特殊环境下（符号链接、只读文件系统、损坏的配置文件）遇到静默失败或误导性错误**.

**Deferred items addressed:**
- LoadConfig uses global viper singleton [config/config.go:82-91]
- 非 "not exist" 的 os.Stat 错误被静默忽略 [config/config.go:68]
- 符号链接/TOCTOU/目录/空文件/只读fs 等边界情况 [config/config.go]

**Acceptance Criteria:**

**Given** 配置文件路径指向一个不存在的路径
**When** os.Stat 返回错误且不是 os.ErrNotExist
**Then** 返回错误而非静默忽略（如权限错误、IO 错误）

**Given** 配置文件路径指向一个符号链接
**When** LoadConfig 读取配置
**Then** 正确解析符号链接指向的文件内容

**Given** 配置文件存在但内容为空
**When** LoadConfig 解析配置
**Then** 返回空 context 列表而非 panic 或静默返回 nil

**Given** 配置目录为只读文件系统
**When** 尝试自动创建配置文件
**Then** 返回明确的权限错误

**Given** LoadConfig 被多次调用
**When** 使用不同的配置路径
**Then** 每次调用独立加载，不受前一次 viper 全局状态影响

### Story 4.2: TUI Selector Quality Polish

As a **developer**,
I want **TUI 选择器的魔法数字消除、边界 case 修复、测试覆盖完善**,
So that **代码可维护性提升，空字符串等边界情况不会导致用户困惑**.

**Deferred items addressed:**
- maxItems 魔法数字 10 无解释 [internal/ui/selector.go:59]
- SetRect Y 坐标硬编码为 1 无解释 [internal/ui/selector.go:68]
- 缺少对新增代码路径的测试 [internal/ui/selector_test.go]
- 空字符串上下文名被误认为取消 [internal/ui/selector.go]
- flexHeightPadding 值与实际 tview 布局无验证 [internal/ui/selector.go:25]

**Acceptance Criteria:**

**Given** selector.go 中的常量定义
**When** 审查代码
**Then** maxItems(10)、flexHeightPadding(4)、SetRect Y 坐标(1) 等值有命名常量和注释说明选取依据

**Given** 配置中存在一个名为空字符串的 context
**When** TUI 显示该 context 且用户选择它
**Then** 正确返回空字符串名称，不与取消操作混淆

**Given** selector 的测试文件
**When** 运行测试
**Then** 表驱动测试覆盖：正常选择、取消、panic recovery、空列表

**Given** flexHeightPadding 值
**When** 验证 tview 实际布局
**Then** padding 值与 tview Flex 实际使用的边距一致或添加注释说明偏差原因
