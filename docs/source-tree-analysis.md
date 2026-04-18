# ccctx - Source Tree Analysis

**Date:** 2026-04-17

## Overview

ccctx is a single-part Go CLI application with a flat, focused directory structure. The codebase follows Go's standard package layout with Cobra CLI conventions.

## Complete Directory Structure

```
ccctx/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── Makefile                # Build automation
├── README.md               # User documentation
├── CLAUDE.md               # Claude Code instructions
├── LICENSE                 # MIT License
├── .gitignore              # Git ignore rules
├── cmd/                    # CLI command implementations
│   ├── list.go             # `ccctx list` command
│   └── run.go              # `ccctx run` command
├── config/                 # Configuration management
│   ├── config.go           # TOML config loading, context CRUD
│   └── config_test.go      # Unit tests for config package
├── internal/               # Internal packages (not importable externally)
│   └── ui/
│       └── selector.go     # Interactive TUI context selector
├── examples/               # Example configuration files
│   └── config.toml         # Sample TOML configuration
└── docs/                   # Project documentation
    └── REQUIREMENT.md      # Original product requirements
```

## Critical Directories

### `cmd/`

**Purpose:** CLI subcommand implementations using Cobra framework.
**Contains:** Two command files — `list.go` (context listing) and `run.go` (context execution with argument forwarding).
**Entry Points:** Commands are registered in `main.go` via `rootCmd.AddCommand()`.

### `config/`

**Purpose:** Configuration management layer using Viper for TOML parsing.
**Contains:** Configuration struct definitions, file loading, context CRUD operations, environment variable resolution.
**Entry Points:** `LoadConfig()`, `ListContexts()`, `GetContext()` are the primary exported functions.

### `internal/ui/`

**Purpose:** Terminal UI components for interactive context selection.
**Contains:** Tview-based selector with arrow key and vim key (j/k) navigation.
**Entry Points:** `RunContextSelector()` is the single exported function.

### `examples/`

**Purpose:** Example configuration files for new users.
**Contains:** A sample `config.toml` demonstrating context definitions.

### `docs/`

**Purpose:** Project documentation and requirements.
**Contains:** Original product requirements document.

## Entry Points

- **Main Entry:** `main.go` — Initializes Cobra root command, registers `list` and `run` subcommands, executes.

## File Organization Patterns

- **Standard Go layout:** `cmd/` for commands, `internal/` for non-exportable packages, `config/` for configuration
- **Cobra convention:** One file per subcommand in `cmd/`
- **Flat structure:** No nested packages beyond one level — appropriate for the project's small size

## Key File Types

### Go Source Files

- **Pattern:** `*.go` in `cmd/`, `config/`, `internal/ui/`
- **Purpose:** Application logic
- **Examples:** `cmd/run.go`, `config/config.go`, `internal/ui/selector.go`

### Go Test Files

- **Pattern:** `*_test.go`
- **Purpose:** Unit tests
- **Examples:** `config/config_test.go`

### Configuration Files

- **Pattern:** `config.toml`
- **Purpose:** User-defined context configurations
- **Examples:** `examples/config.toml`

## Configuration Files

- **`go.mod`**: Go module definition (github.com/dsdashun/ccctx, Go 1.23.3)
- **`Makefile`**: Build targets (build, install, test, fmt, vet, tidy, clean)
- **`.gitignore`**: Ignores built binary, .serena, swap files
- **`examples/config.toml`**: Sample user configuration

## Notes for Development

- The project is small enough that all source files can be understood in a single reading session
- No intermediate abstractions — commands directly call config and UI functions
- The `internal/` package prevents external imports of the UI selector
- Adding a new subcommand requires: creating a file in `cmd/`, registering it in `main.go`

---
_Generated using BMAD Method `document-project` workflow_
