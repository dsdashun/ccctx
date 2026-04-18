# Story 1.2: Refactor run Command as Ultra-Thin Wrapper

Status: done

## Story

As a **developer**,
I want **run 命令重构为调用 internal/runner/ 的超薄包装器，应用架构定义的三项 breaking changes（未知参数硬错误、TUI 取消退出码 1、无上下文时 stderr 错误）**,
So that **run 的行为与架构规范完全一致，为 exec 命令提供可参照的模板，且代码量精简到最小**.

## Acceptance Criteria

1. **AC1: TUI 取消退出码从 0 改为 1** — `ccctx run`（无参数）→ TUI → 按 ESC → 输出 "Operation cancelled." 到 stdout，退出码为 1（不再 `return`）

2. **AC2: "No contexts found" 改为 stderr + exit 1** — `ccctx run`（无参数且配置为空）→ 输出 `Error: no contexts found\n` 到 stderr，退出码为 1（不再 `fmt.Println` + `return`）

3. **AC3: 未知参数硬错误** — `ccctx run nonexistent` → `Error: context 'nonexistent' not found` 到 stderr，退出码 1（不再 fallback 到 TUI）。**注意：此行为已在 Story 1.1 中生效**（ParseArgs 始终将首参数视为 provider，`runner.New()` 调用 `config.GetContext()` 返回错误），本 Story 仅需确认无需额外工作。

4. **AC4: run.go 为超薄包装器** — `cmd/run.go` 不再包含业务逻辑，仅做：调用 `runner.ParseArgs` → TUI 选择（如需）→ `exec.LookPath("claude")` → `runner.New()` + `r.Run()` → 处理退出码。无中间变量 `contextName`，直接使用 `provider`。

5. **AC5: 所有现有测试通过** — `go test ./...` 通过，`go vet ./...` 无警告

6. **AC6: FR25 得到满足** — run 命令使用共享执行管道，是 exec -- claude 的等价物

## Tasks / Subtasks

- [x] Task 1: Apply breaking change — TUI cancellation exit code (AC: #1)
  - [x] 将 `cmd/run.go:42-44` 的 `fmt.Println("Operation cancelled."); return` 改为 `fmt.Println("Operation cancelled."); os.Exit(1)`

- [x] Task 2: Apply breaking change — "No contexts found" error handling (AC: #2)
  - [x] 将 `cmd/run.go:35-37` 的 `fmt.Println("No contexts found."); return` 改为 `fmt.Fprintf(os.Stderr, "Error: no contexts found\n"); os.Exit(1)`

- [x] Task 3: Simplify run.go — remove contextName intermediate variable (AC: #4)
  - [x] 消除 `contextName` 变量，直接使用 `provider` 覆盖（TUI 选择后 `provider = selected`）
  - [x] 简化 TUI 和非 TUI 分支：统一使用 `provider` 传给 `runner.Options{ContextName: provider}`

- [x] Task 4: Align exit code handling with architecture template (AC: #4)
  - [x] 将 `if exitCode != 0 { os.Exit(exitCode) }` 改为无条件的 `os.Exit(exitCode)`，与架构模板一致

- [x] Task 5: Run tests and verify (AC: #5, #6)
  - [x] `go test ./...` 全部通过
  - [x] `go vet ./...` 无警告
  - [x] 手动验证三项 breaking changes 的行为

## Dev Notes

### 当前状态分析

Story 1.1 已完成核心提取工作：`cmd/run.go` 从 ~200 行降至 79 行，已委托给 `runner.ParseArgs()`、`runner.New()`、`r.Run()`。本 Story 的工作量较小，主要是应用架构定义的三项 breaking changes 和代码清理。

### 当前 run.go 关键行号（需修改的位置）

| 行号 | 当前行为 | 目标行为 | Breaking Change |
|------|---------|---------|-----------------|
| 35-37 | `fmt.Println("No contexts found."); return` | `fmt.Fprintf(os.Stderr, "Error: no contexts found\n"); os.Exit(1)` | 是 |
| 41-44 | `fmt.Println("Operation cancelled."); return` | `fmt.Println("Operation cancelled."); os.Exit(1)` | 是 |
| 26 | `var contextName string` | 消除，直接用 `provider` | 否（清理） |
| 48 | `contextName = selected` | `provider = selected` | 否（清理） |
| 50 | `contextName = provider` | 删除此分支 | 否（清理） |
| 61-64 | `runner.Options{ContextName: contextName, ...}` | `runner.Options{ContextName: provider, ...}` | 否（清理） |
| 75-77 | `if exitCode != 0 { os.Exit(exitCode) }` | `os.Exit(exitCode)` | 否（对齐模板） |

### 目标代码结构（参考架构模板）

```go
var RunCmd = &cobra.Command{
    Use:   "run [context] [-- claude-args...]",
    Short: "Run claude with a context",
    Long:  "Run claude with the specified context or interactively select one. Arguments after '--' are passed to claude.",
    Args:  cobra.ArbitraryArgs,
    Run: func(cmd *cobra.Command, args []string) {
        provider, targetArgs, useTUI, err := runner.ParseArgs(args)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }

        if useTUI {
            contexts, err := config.ListContexts()
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
            }
            if len(contexts) == 0 {
                fmt.Fprintf(os.Stderr, "Error: no contexts found\n")
                os.Exit(1)
            }

            provider, err = ui.RunContextSelector()
            if err != nil {
                if err.Error() == "operation cancelled" {
                    fmt.Println("Operation cancelled.")
                    os.Exit(1)
                }
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
            }
        }

        claudePath, err := exec.LookPath("claude")
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: claude not found in PATH\n")
            os.Exit(1)
        }

        target := append([]string{claudePath}, targetArgs...)

        r, err := runner.New(runner.Options{
            ContextName: provider,
            Target:      target,
        })
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }

        exitCode, err := r.Run()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        os.Exit(exitCode)
    },
}
```

### 注意事项

1. **TUI selector 签名不变** — 当前 `ui.RunContextSelector()` 无参数，内部调用 `config.ListContexts()`。架构模板显示未来签名改为 `ui.RunContextSelector(contexts)`，但此变更是 Story 2.1 的范围（TUI 共享化）。本 Story 保持当前签名。

2. **deferred-work 中的脆弱字符串比较** — `selectorErr.Error() == "operation cancelled"` 是预存的技术债，不在本 Story 范围内修复。

3. **AC3 无需额外工作** — 未知参数硬错误已在 Story 1.1 中生效（`runner.ParseArgs` 将首参数视为 provider，`runner.New()` 调用 `config.GetContext()` 返回 not found 错误）。只需确认行为正确即可。

4. **`os.Exit(exitCode)` vs 条件退出** — 架构模板使用无条件 `os.Exit(exitCode)`。当 exitCode 为 0 时，`os.Exit(0)` 也会跳过 deferred 函数。但由于 Cobra 的 Run 函数内没有 defer，这是安全的。使用无条件形式与架构模板保持一致。

### 与 exec.go 模板的关系

本 Story 完成后，`cmd/run.go` 将成为 `cmd/exec.go`（Story 1.3）的参照模板。两者的区别仅在于：
- `run.go` 使用 `exec.LookPath("claude")` + `target := append([]string{claudePath}, targetArgs...)`
- `exec.go` 使用 `os.Getenv("SHELL")`（当 targetArgs 为空时）+ 无 LookPath

### Anti-Patterns (Do NOT)

- Do NOT 在 run.go 中添加新的业务逻辑 — 它是超薄包装器
- Do NOT 修改 `internal/runner/` 包 — Story 1.1 已完成提取
- Do NOT 修改 `internal/ui/selector.go` — TUI 签名变更是 Story 2.1
- Do NOT 修改 `config/` 包 — 本 Story 无配置变更
- Do NOT 在 run.go 中调用 `config.GetContext()` — 由 `runner.New()` 内部处理
- Do NOT 修改 error message format — 保持 `Error: <lowercase, no trailing period>`

### References

- [Source: _bmad-output/planning-artifacts/architecture.md#Command File Template] — run.go 目标模板
- [Source: _bmad-output/planning-artifacts/architecture.md#Breaking changes from current code] — 三项 breaking changes 定义
- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.2] — Story 需求和验收标准
- [Source: _bmad-output/implementation-artifacts/1-1-extract-runner-struct-and-parseargs-into-internal-runner.md] — 前序 Story 完成记录和 Review findings
- [Source: _bmad-output/implementation-artifacts/deferred-work.md] — 脆弱字符串比较的延迟项
- [Source: cmd/run.go] — 当前待修改的文件
- [Source: internal/runner/runner.go] — Runner 包（不需修改，仅调用）
- [Source: internal/runner/args.go] — ParseArgs（不需修改，仅调用）

## Dev Agent Record

### Agent Model Used

Claude (GLM-5.1)

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

- ✅ Applied breaking change: TUI cancellation now exits with code 1 instead of returning (AC1)
- ✅ Applied breaking change: "No contexts found" now outputs to stderr with "Error: no contexts found" and exits with code 1 (AC2)
- ✅ AC3 (unknown argument hard error) already working from Story 1.1 — ParseArgs treats first arg as provider, runner.New() returns not found error
- ✅ Removed contextName intermediate variable; provider is used directly throughout (AC4)
- ✅ Changed conditional exit `if exitCode != 0 { os.Exit(exitCode) }` to unconditional `os.Exit(exitCode)` aligning with architecture template (AC4)
- ✅ All tests pass (go test ./...), no vet warnings (go vet ./...), build succeeds (AC5)
- ✅ run.go is now an ultra-thin wrapper matching architecture template, ready as reference for exec.go (AC6/FR25)

### Change Log

- 2026-04-19: Refactored cmd/run.go — applied three breaking changes (TUI cancel exit 1, no contexts stderr+exit 1, unknown args hard error confirmed), removed contextName variable, aligned exit code handling with architecture template

### File List

- cmd/run.go (modified)

### Review Findings

- [x] [Review][Patch] `r.Run()` error message uses non-compliant format [cmd/run.go:68] — fixed: changed to `Error: %v\n` — Uses `"Error executing claude: %v\n"` but architecture template requires `"Error: %v\n"` (lowercase, no trailing period)

- [x] [Review][Defer] Fragile string comparison for TUI cancellation detection [cmd/run.go:40] — deferred, pre-existing
- [x] [Review][Defer] Redundant config.ListContexts() call in TUI path [cmd/run.go:27, internal/ui/selector.go:14] — deferred, pre-existing; Story 2.1 will address TUI shared contexts
- [x] [Review][Defer] TUI list selection bounds check returns misleading error [internal/ui/selector.go:93-99,118] — deferred, pre-existing
- [x] [Review][Defer] No panic recovery in TUI selector may leave terminal corrupted [internal/ui/selector.go:27-123] — deferred, pre-existing
- [x] [Review][Defer] Race condition on TUI cancelled flag [internal/ui/selector.go:62,68,113] — deferred, pre-existing
- [x] [Review][Defer] cmd/run.go has no automated tests [cmd/run.go] — deferred, pre-existing; breaking changes are manually verified only
- [x] [Review][Defer] runner.New does not validate BaseURL format [internal/runner/runner.go:31-36] — deferred, pre-existing
- [x] [Review][Defer] Environment variable injection edge case with empty values [internal/runner/runner.go:66-77] — deferred, pre-existing
- [x] [Review][Defer] Default config file created with world-readable permissions [config/config.go:77] — deferred, pre-existing
- [x] [Review][Defer] LoadConfig uses global viper singleton [config/config.go:82-91] — deferred, pre-existing
