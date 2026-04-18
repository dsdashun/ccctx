---
project_name: 'ccctx'
user_name: 'dsdashun'
date: '2026-04-18'
sections_completed:
  ['technology_stack', 'language_specific_rules', 'framework_specific_rules', 'testing_rules', 'code_quality_rules', 'development_workflow_rules', 'critical_dont_miss_rules']
status: 'complete'
rule_count: 53
optimized_for_llm: true
---

# Project Context for AI Agents

_This file contains critical rules and patterns that AI agents must follow when implementing code in this project. Focus on unobvious details that agents might otherwise miss._

---

## Technology Stack & Versions

- **Language:** Go 1.23.3
- **CLI Framework:** github.com/spf13/cobra v1.9.1
- **Config Management:** github.com/spf13/viper v1.20.1 (TOML format)
- **TUI Framework:** github.com/rivo/tview v0.0.0-20250625164341 + github.com/gdamore/tcell/v2 v2.8.1
- **Testing:** github.com/stretchr/testify (assert + require) — table-driven pattern
- **Build:** Makefile, CGO_ENABLED=0 (static binary)

## Critical Implementation Rules

### Language-Specific Rules (Go)

- Build with `CGO_ENABLED=0` for static binary — no C dependencies
- Use `fmt.Fprintf(os.Stderr, ...)` for error output, never `fmt.Println` for errors
- Exit with `os.Exit(1)` on errors; propagate child exit code via `errors.As(err, &exitErr)` + `exec.ExitError` (not type assertion `.(*exec.ExitError)`)
- Use `mapstructure` struct tags for Viper config binding (e.g., `mapstructure:"base_url"`)
- Error wrapping: use `fmt.Errorf("context: %w", err)` for wrapped errors
- Package naming: lowercase single word, no underscores (e.g., `config`, `cmd`, `runner`)
- Exported package-level Cobra commands as `var` (e.g., `var RunCmd = &cobra.Command{...}`)
- Register subcommands in `init()` function of each command file
- `internal/` package for encapsulated internals not importable by external projects
- Use `strings.HasPrefix(e, "ANTHROPIC_")` for env var filtering — catches all current and future ANTHROPIC_* vars
- Never inject env vars with empty values — check `!= ""` before appending optional vars

### Framework-Specific Rules

#### Cobra CLI
- Each command lives in its own file under `cmd/` package
- Commands are exported vars (e.g., `var RunCmd`, `var ListCmd`, `var ExecCmd`)
- Use `cobra.ArbitraryArgs` for commands that accept flexible arguments
- Argument parsing: `--` separator splits context args from forwarded args
- Never use `cobra.PersistentFlags` unless the flag must propagate to subcommands
- `exec` is the core command; `run` is an ultra-thin wrapper (`run` ≡ `exec -- claude`)

#### Viper Configuration
- Config format is TOML; path via `CCCTX_CONFIG_PATH` env or `~/.ccctx/config.toml`
- Auto-create config directory and default config file if missing
- Use `viper.Unmarshal()` with mapstructure-tagged structs — never read keys individually
- `env:` prefix is a project convention, not a Viper feature — `resolveEnvVar()` must be called explicitly
- Config struct uses nested map: `map[string]Context` under `context` key
- Config file permissions: `0600` on creation (protects auth tokens)

#### TUI (tview)
- Interactive selector in `internal/ui/` package
- Support both arrow keys and vim bindings (j/k) for navigation
- ESC cancels and returns `"operation cancelled"` error
- Compact layout: position at top of screen, max 10 visible items with scrolling
- TUI functions return `(string, error)` for clean integration with CLI flow

#### Shared Runner (internal/runner/)
- `Runner` struct-based design: `New(opts Options) (*Runner, error)` + `Run() (int, error)`
- `ParseArgs(args) (provider, targetArgs, useTUI, error)` — command-agnostic argument parsing
- Runner must be free of I/O side effects: no `os.Exit`, no `fmt.Fprintf(os.Stderr, ...)`
- Commands own output and exit; runner only returns values and errors
- Error message format: `Error: <lowercase description without trailing period>`

### Testing Rules

- Use **testify** framework (`github.com/stretchr/testify`) — `require` for fatal assertions, `assert` for non-fatal
- Use table-driven test pattern: `[]struct{ name string; ... }` with `t.Run()` for subtests
- Name test functions as `Test<FunctionName>` (e.g., `TestResolveEnvVar`, `TestParseArgs`)
- Error assertions: `require.NoError(t, err)` / `require.Error(t, err)` + `assert.Contains(t, err.Error(), "expected text")`
- Equality checks: `assert.Equal(t, expected, actual)`
- Use `t.Setenv()` for environment variable tests (automatic cleanup, replaces manual defer pattern)
- Run tests with `go test ./...` or `make test`
- Test scope for shared kernel: `runner_test.go` covers `buildEnv()`, `Run()` exit codes; `args_test.go` covers `ParseArgs()` edge cases
- Command files (`cmd/*.go`) are thin wrappers — no dedicated unit tests needed
- Keep tests self-contained within their package — no shared test utilities

### Code Quality & Style Rules

- Format with `go fmt ./...` or `make fmt` — no custom formatting config
- Run `go vet ./...` or `make vet` before committing
- File organization: one package per directory, matching package name
- Naming conventions:
  - Files: lowercase with underscores (e.g., `config.go`, `config_test.go`)
  - Functions: PascalCase (exported) or camelCase (unexported)
  - Acronyms stay uppercase: `BaseURL`, `AuthToken`, `RunCmd`
- No comments unless the WHY is non-obvious — code should be self-documenting
- No doc comments on unexported functions
- Error messages: lowercase, no trailing period, no capitalized first letter
- Use `make tidy` to clean up go.mod/go.sum after dependency changes

### Development Workflow Rules

- Build before testing: `make build` produces the `ccctx` binary
- Commit message style: conventional commits (e.g., `feat:`, `fix:`, `chore:`)
- Branch: main branch for all work, feature branches for PRs
- No CI/CD pipeline configured — rely on local `make test` and `make vet`
- Install locally with `make install` for manual testing
- Config file for testing: `~/.ccctx/config.toml` or override with `CCCTX_CONFIG_PATH`
- Implementation sequence: extract runner → refactor run → add exec → fix config permissions → add tests

### Critical Don't-Miss Rules

- NEVER hardcode config paths — always use `GetConfigPath()` which respects `CCCTX_CONFIG_PATH`
- NEVER leak ANTHROPIC environment variables — strip all existing `ANTHROPIC_*` vars before injecting new ones using `strings.HasPrefix`
- NEVER expose auth tokens in logs or stdout — tokens are resolved internally only
- NEVER call `os.Exit()` inside `internal/runner/` — breaks testability
- NEVER use `fmt.Fprintf(os.Stderr, ...)` inside runner — commands own output
- NEVER hardcode `"claude"` in runner — it's a run-specific concern
- NEVER use `LookPath` for exec command — target is user-specified directly
- NEVER duplicate logic between run.go and exec.go — exec is core, run is thin wrapper
- The `env:` prefix is a project convention, not a standard Viper feature — `resolveEnvVar()` must be called explicitly
- `resolveEnvVar()` only matches exact prefix `env:` at the start of the string — `some-env:text` is treated as a literal value
- The `--` separator is critical: args before it are context names, args after are forwarded. Missing this breaks argument forwarding
- `ParseArgs` forwards remaining args without `--` when no separator present; rejects multiple args before `--`
- TUI cancellation exits with code 1, not 0 — cancellation is an abort, not a success
- Validate `$SHELL != ""` before using it as default target for exec
- First arg without `--` is always treated as provider name — no TUI fallback for unknown names (breaking change from current behavior)
- `Context` struct fields `Model` and `SmallFastModel` are optional — always check for empty string before setting env vars
- Runner boundary: may import `config/` and stdlib only; must NOT import `internal/ui/` or `cmd/`

---

## Usage Guidelines

**For AI Agents:**

- Read this file before implementing any code
- Follow ALL rules exactly as documented
- When in doubt, prefer the more restrictive option
- Update this file if new patterns emerge

**For Humans:**

- Keep this file lean and focused on agent needs
- Update when technology stack changes
- Review quarterly for outdated rules
- Remove rules that become obvious over time

Last Updated: 2026-04-18
