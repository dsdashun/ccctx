## Deferred from: code review of 1-1-extract-runner-struct-and-parseargs-into-internal-runner (2026-04-18)

- ~~Fragile string equality for TUI cancellation check [cmd/run.go]~~ — resolved, now uses sentinel error `errors.Is(err, ui.ErrCancelled)`

## Deferred from: code review of 1-2-refactor-run-command-as-ultra-thin-wrapper (2026-04-19)

- ~~Fragile string comparison for TUI cancellation detection [cmd/run.go:40]~~ — resolved, now uses sentinel error `errors.Is(err, ui.ErrCancelled)`
- ~~Redundant config.ListContexts() call in TUI path [cmd/run.go:27, internal/ui/selector.go:14]~~ — resolved, selector now receives contexts as parameter
- ~~TUI list selection bounds check returns misleading error [internal/ui/selector.go:93-99,118]~~ — resolved, problematic bounds check code removed
- ~~No panic recovery in TUI selector may leave terminal corrupted [internal/ui/selector.go:27-123]~~ — resolved, defer/recover with stack trace added
- ~~Race condition on TUI cancelled flag [internal/ui/selector.go:62,68,113]~~ — resolved, replaced with selectedContext string pattern
- ~~cmd/run.go has no automated tests [cmd/run.go]~~ — promoted to Story 3.1
- ~~runner.New does not validate BaseURL format [internal/runner/runner.go:31-36]~~ — promoted to Story 3.1
- ~~Environment variable injection edge case with empty values [internal/runner/runner.go:66-77]~~ — resolved, New() validates BaseURL and AuthToken non-empty before buildEnv
- ~~Default config file created with world-readable permissions [config/config.go:77]~~ — resolved, now uses 0600
- ~~LoadConfig uses global viper singleton [config/config.go:82-91]~~ — promoted to Story 4.1

## Deferred from: code review of 1-3-implement-exec-subcommand-core-command (2026-04-19)

- ~~runner.Run double-prints non-ExitError failures [internal/runner/runner.go:51-57, cmd/exec.go:62-66]~~ — promoted to Story 3.1
- ~~ParseArgs treats flags before `--` as context names [internal/runner/args.go:30-34]~~ — promoted to Story 3.1
- ~~os.Exit calls bypass future defer cleanup [cmd/exec.go]~~ — promoted to Story 3.1

## Deferred from: code review of 1-4-fix-config-file-permissions (2026-04-19)

- ~~非 "not exist" 的 `os.Stat` 错误被静默忽略 [config/config.go:68]~~ — promoted to Story 4.1
- ~~符号链接/TOCTOU/目录/空文件/只读fs 等边界情况~~ — promoted to Story 4.1

## Deferred from: code review of 2-1-exec-command-integrates-tui-interactive-selector (2026-04-19)

- ~~Terminal may be left in raw mode if app.Run() returns error [internal/ui/selector.go:91]~~ — resolved by Story 2.2
- ~~Hardcoded flex width 50 truncates long context names [internal/ui/selector.go:53]~~ — resolved by Story 2.2
- ~~Magic number maxItems+4 in SetRect lacks explanation [internal/ui/selector.go:53]~~ — resolved by Story 2.2
- ~~Divergent selection sources between SetDoneFunc and SetSelectedFunc [internal/ui/selector.go:82,87]~~ — resolved by Story 2.2
- ~~Cancellation message goes to stdout instead of stderr [cmd/exec.go:39, cmd/run.go:42]~~ — resolved by Story 2.2
- ~~Panic recovery converts panic to plain error, losing stack trace [internal/ui/selector.go:24-32]~~ — resolved by Story 2.2

## Deferred from: code review of 2-2-tui-selector-cleanup-and-quality-improvements (2026-04-19)

- ~~maxItems 魔法数字 10 无解释 [internal/ui/selector.go:59]~~ — promoted to Story 4.2
- ~~SetRect Y 坐标硬编码为 1 无解释 [internal/ui/selector.go:68]~~ — promoted to Story 4.2
- ~~缺少对新增代码路径的测试 [internal/ui/selector_test.go]~~ — promoted to Story 4.2
- ~~空字符串上下文名被误认为取消 [internal/ui/selector.go]~~ — promoted to Story 4.2
- ~~flexHeightPadding 值与实际 tview 布局无验证 [internal/ui/selector.go:25]~~ — promoted to Story 4.2

## Deferred from: code review of 3-1-runner-and-command-layer-robustness (2026-04-23)

- 跨平台兼容性：Windows 上假 claude 脚本使用 shebang 不工作 [cmd/run_test.go]
- buildEnv 不验证注入环境变量值中的换行或空字节 [internal/runner/runner.go:82-99]
- config 对 base_url 不支持 env: 前缀解析（与 auth_token 不一致）[config/config.go]
- SHELL 环境变量可执行性未验证 [cmd/exec.go:52-58]
- 目标命令被信号终止时退出码丢失 [internal/runner/runner.go:64-80]
- 空 provider 名未在 ParseArgs 中显式拒绝 [internal/runner/args.go]
- 通用 "Error: %v" 消息模式不利于程序化错误区分 [cmd/run.go, cmd/exec.go]

## Deferred from: code review of 3-2-add-model-override-flags-to-run-and-exec (2026-04-23)

- `claude` 二进制存在但不可执行时，`exec.LookPath` 成功但 `exec.Command` 会失败 [cmd/run.go:66-70]
- `SHELL` 环境变量设为空字符串时会被当作目标命令传递 [cmd/exec.go:63-70]

## Deferred from: code review of 4-1-config-runner-model-field-expansion (2026-05-13)

- 集成测试 mock 脚本只捕获 ANTHROPIC_MODEL 和 ANTHROPIC_DEFAULT_HAIKU_MODEL，缺少 ANTHROPIC_DEFAULT_SONNET_MODEL 和 ANTHROPIC_DEFAULT_OPUS_MODEL 的端到端覆盖。待 Story 4.2 添加 CLI flags 后补充 [cmd/run_test.go, cmd/exec_test.go]

## Deferred from: code review of 4-2-cli-flag-extension-backward-compatibility (2026-05-13)

- Near-identical test tables duplicated across exec_test.go and run_test.go — pre-existing pattern
- validateFlagValue returns ambiguous error messages without flag context — pre-existing (same pattern was used for --model, --small-fast-model)
- exec_test.go missing general error-path tests (pre-existing, not caused by this change)
- --haiku-model "" (explicit empty) can be silently overridden by --small-fast-model alias — extremely unlikely edge case
- Mock script line ordering fragile to future env var injection changes — design observation
- No tests for unicode or extremely long flag values — low priority
