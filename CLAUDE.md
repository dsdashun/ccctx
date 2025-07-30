# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`ccctx` is a Go CLI tool that manages and runs Claude Code with different contexts. It allows users to switch between different Claude API configurations (base URLs and auth tokens) using a simple TOML configuration file.

## Build and Development Commands

### Common Development Tasks
- `make build` - Build the binary (CGO_ENABLED=0)
- `make install` - Install to GOPATH/bin
- `make test` - Run tests
- `make fmt` - Format code with gofmt
- `make vet` - Run go vet for static analysis
- `make tidy` - Clean up dependencies
- `make clean` - Remove build artifacts

### Direct Go Commands
- `go build -o ccctx .` - Build binary
- `go install .` - Install to GOPATH/bin
- `go test ./...` - Run tests
- `go fmt ./...` - Format code
- `go vet ./...` - Static analysis

## Architecture

### Core Components
- **main.go** - Entry point using Cobra CLI framework
- **cmd/list.go** - Lists available contexts from configuration
- **cmd/run.go** - Main execution logic with argument parsing and Claude invocation
- **config/config.go** - TOML configuration management using Viper
- **internal/ui/selector.go** - Interactive TUI for context selection using tview

### Key Design Patterns
- **Cobra CLI Framework** - Structured command-line interface with subcommands
- **Viper Configuration** - TOML config file management with environment variable override support
- **Tview UI** - Rich terminal interface for interactive context selection
- **Environment Isolation** - Clean environment management when executing Claude

### Configuration System
- Default config path: `~/.ccctx/config.toml`
- Override with `CCCTX_CONFIG_PATH` environment variable
- Auto-creates config file with example if missing
- Structure: `[context.<name>]` with `base_url` and `auth_token` fields

### Argument Processing Logic
The `run` command has sophisticated argument parsing:
- Arguments before `--` are treated as context names
- Arguments after `--` are forwarded to Claude
- Interactive mode triggered when no context specified
- Supports both direct context specification and interactive selection

### UI/UX Features
- Interactive context selector with arrow keys and vim bindings (j/k)
- ESC to cancel operation
- Compact TUI layout positioned at top of screen
- Maximum 10 items displayed with scrolling

## Environment Variables
- `CCCTX_CONFIG_PATH` - Override default config file location
- `ANTHROPIC_BASE_URL` - Set by tool when executing Claude
- `ANTHROPIC_AUTH_TOKEN` - Set by tool when executing Claude

## Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/rivo/tview` - Terminal UI framework
- `github.com/gdamore/tcell/v2` - Terminal interface (tview dependency)

## Testing the Tool
1. Build the tool: `make build`
2. Create test contexts in `~/.ccctx/config.toml`
3. Test listing: `./ccctx list`
4. Test interactive mode: `./ccctx run`
5. Test direct context: `./ccctx run <context-name>`
6. Test argument forwarding: `./ccctx run <context-name> -- --help`

## Testing Guidelines
When adding unit tests, use the testify framework and use table-driven pattern as much as possible.