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
