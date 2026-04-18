# ccctx - Development Guide

**Date:** 2026-04-17

## Prerequisites

- **Go:** 1.23.3 or later
- **Make:** For build automation (optional, can use Go commands directly)

## Setup

```bash
# Clone the repository
git clone https://github.com/dsdashun/ccctx.git
cd ccctx

# Install dependencies (handled automatically by Go)
go mod download
```

## Building

```bash
# Build using Make
make build

# Or build directly
CGO_ENABLED=0 go build -o ccctx .
```

## Installation

```bash
# Install to GOPATH/bin
make install

# Or install directly
CGO_ENABLED=0 go install .
```

## Running Locally

```bash
# Build first, then run
make build
./ccctx --help
./ccctx list
./ccctx run

# Or run without building
go run . --help
```

## Testing

```bash
# Run all tests
make test
# Or
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./config/
```

## Code Quality

```bash
# Format code
make fmt
# Or
go fmt ./...

# Static analysis
make vet
# Or
go vet ./...

# Tidy dependencies
make tidy
# Or
go mod tidy
```

## Project Structure

```
main.go           # Entry point
cmd/              # CLI subcommands (add new commands here)
config/           # Configuration management
internal/ui/      # Terminal UI components
examples/         # Example configuration files
```

## Common Development Tasks

### Adding a New Subcommand

1. Create a new file in `cmd/` (e.g., `cmd/switch.go`)
2. Define a Cobra command variable
3. Register it in `main.go`'s `init()` function: `rootCmd.AddCommand(cmd.SwitchCmd)`

### Adding a New Config Field

1. Add the field to the `Context` struct in `config/config.go` with a `mapstructure` tag
2. Update `GetContext()` to handle the new field
3. Update `examples/config.toml` with the new field
4. Add tests in `config/config_test.go`

### Modifying the Interactive Selector

1. Edit `internal/ui/selector.go`
2. The selector uses tview — refer to [tview documentation](https://pkg.go.dev/github.com/rivo/tview)
3. Key bindings are set via `SetInputCapture()`

## Configuration for Development

Create a test config at `~/.ccctx/config.toml`:

```toml
[context.dev]
base_url = "https://api.anthropic.com"
auth_token = "dev-test-token"

[context.staging]
base_url = "https://api.anthropic.com"
auth_token = "env:STAGING_TOKEN"
model = "claude-3-5-sonnet-20241022"
```

Or override the config path:

```bash
export CCCTX_CONFIG_PATH=/path/to/test/config.toml
```

## Testing Guidelines

- Use the standard Go `testing` package
- Prefer table-driven test patterns
- Tests should be in the same package (same directory) as the code being tested
- File naming convention: `*_test.go`

## Build Targets Reference

| Target | Command | Description |
|--------|---------|-------------|
| all | `make all` | Default, same as build |
| build | `make build` | Build binary (CGO_ENABLED=0) |
| install | `make install` | Install to GOPATH/bin |
| test | `make test` | Run tests |
| fmt | `make fmt` | Format code |
| vet | `make vet` | Static analysis |
| tidy | `make tidy` | Clean dependencies |
| clean | `make clean` | Remove build artifacts |

---
_Generated using BMAD Method `document-project` workflow_
