## Deferred from: code review of 1-1-extract-runner-struct-and-parseargs-into-internal-runner (2026-04-18)

- Fragile string equality for TUI cancellation check [cmd/run.go] — deferred, pre-existing

## Deferred from: code review of 1-2-refactor-run-command-as-ultra-thin-wrapper (2026-04-19)

- Fragile string comparison for TUI cancellation detection [cmd/run.go:40] — deferred, pre-existing
- Redundant config.ListContexts() call in TUI path [cmd/run.go:27, internal/ui/selector.go:14] — deferred, pre-existing; Story 2.1 will address TUI shared contexts
- TUI list selection bounds check returns misleading error [internal/ui/selector.go:93-99,118] — deferred, pre-existing
- No panic recovery in TUI selector may leave terminal corrupted [internal/ui/selector.go:27-123] — deferred, pre-existing
- Race condition on TUI cancelled flag [internal/ui/selector.go:62,68,113] — deferred, pre-existing
- cmd/run.go has no automated tests [cmd/run.go] — deferred, pre-existing; breaking changes are manually verified only
- runner.New does not validate BaseURL format [internal/runner/runner.go:31-36] — deferred, pre-existing
- Environment variable injection edge case with empty values [internal/runner/runner.go:66-77] — deferred, pre-existing
- Default config file created with world-readable permissions [config/config.go:77] — deferred, pre-existing
- LoadConfig uses global viper singleton [config/config.go:82-91] — deferred, pre-existing

## Deferred from: code review of 1-3-implement-exec-subcommand-core-command (2026-04-19)

- runner.Run double-prints non-ExitError failures [internal/runner/runner.go:51-57, cmd/exec.go:62-66] — deferred, pre-existing
- ParseArgs treats flags before `--` as context names [internal/runner/args.go:30-34] — deferred, pre-existing
- os.Exit calls bypass future defer cleanup [cmd/exec.go] — deferred, pre-existing

## Deferred from: code review of 1-4-fix-config-file-permissions (2026-04-19)

- 非 "not exist" 的 `os.Stat` 错误被静默忽略 [config/config.go:68] — deferred, pre-existing
- 符号链接/TOCTOU/目录/空文件/只读fs 等边界情况 — deferred, pre-existing
