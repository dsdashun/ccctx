# ccctx - Project Overview

**Date:** 2026-04-17
**Type:** CLI Tool
**Architecture:** Command Pattern (Cobra CLI)

## Executive Summary

ccctx is a Go CLI tool that manages and runs Claude Code with different API contexts. It solves the problem of switching between multiple Claude API configurations (different base URLs, auth tokens, and model preferences) by providing a simple TOML-based configuration system with both direct and interactive context selection.

## Project Classification

- **Repository Type:** Monolith
- **Project Type:** CLI Tool
- **Primary Language:** Go
- **Architecture Pattern:** Command Pattern (Cobra CLI)

## Technology Stack

| Category | Technology | Version | Purpose |
|----------|-----------|---------|---------|
| Language | Go | 1.23.3 | Primary language |
| CLI Framework | Cobra | v1.9.1 | Command structure and argument parsing |
| Configuration | Viper | v1.20.1 | TOML config file management |
| Terminal UI | tview | (latest) | Interactive context selector |
| Build Tool | Make | - | Build automation |

## Key Features

1. **Context Management:** Define multiple Claude API contexts in a TOML configuration file
2. **Direct Execution:** Run Claude with a specific context via `ccctx run <name>`
3. **Interactive Selection:** TUI-based context picker with arrow keys and vim bindings
4. **Environment Variable Auth:** Support `env:` prefix for secure token resolution
5. **Model Configuration:** Optional per-context model and small_fast_model settings
6. **Argument Forwarding:** Full Claude argument passthrough via `--` separator
7. **Clean Environment:** Executes Claude with isolated environment, no shell side effects

## Architecture Highlights

- **Cobra Command Pattern:** Two subcommands (`list`, `run`) registered on a root command
- **Configuration Layer:** Viper-based TOML parsing with auto-creation of example config
- **Interactive UI:** tview-based selector for context picking
- **Environment Isolation:** Modified environment copy for Claude execution
- **Exit Code Propagation:** Claude's exit code is passed through to the caller

## Development Overview

### Prerequisites

- Go 1.23.3+

### Getting Started

```bash
git clone https://github.com/dsdashun/ccctx.git
cd ccctx
make build
```

### Key Commands

- **Install:** `make install`
- **Build:** `make build`
- **Test:** `make test`
- **Format:** `make fmt`

## Repository Structure

```
ccctx/
├── main.go              # Entry point
├── cmd/                 # CLI subcommands (list, run)
├── config/              # TOML config management + tests
├── internal/ui/         # Interactive TUI selector
├── examples/            # Example configuration
├── docs/                # Documentation
├── go.mod / go.sum      # Go module definition
└── Makefile             # Build automation
```

## Documentation Map

For detailed information, see:

- [index.md](./index.md) - Master documentation index
- [architecture.md](./architecture.md) - Detailed architecture and data flow
- [source-tree-analysis.md](./source-tree-analysis.md) - Directory structure analysis
- [development-guide.md](./development-guide.md) - Development workflow and setup
- [component-inventory.md](./component-inventory.md) - Component catalog

---
_Generated using BMAD Method `document-project` workflow_
