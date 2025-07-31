# ccctx - Claude Code Context Switcher

A CLI tool to manage and run Claude Code with different contexts.

## Features

- List available contexts from a configuration file
- Run Claude Code with a specific context temporarily
- Support for environment variables in authentication tokens for enhanced security

## Installation

```bash
go install github.com/dsdashun/ccctx@latest
```

## Configuration

The tool looks for a configuration file at `~/.ccctx/config.toml`. If it doesn't exist, it will be created with a default example.

Example configuration:

```toml
[context.work]
base_url = "https://api.anthropic.com"
auth_token = "work-token-here"

[context.personal]
base_url = "https://api.anthropic.com"
auth_token = "env:ANTHROPIC_PERSONAL_TOKEN"

[context.staging]
base_url = "https://api.anthropic.com"
auth_token = "env:ANTHROPIC_STAGING_TOKEN"
```

## Usage

```bash
# List available contexts
ccctx list

# Run Claude with a context (interactive mode with arrow keys)
ccctx run

# Run Claude with a specific context
ccctx run personal

# Run Claude with a specific context and additional arguments
ccctx run personal -- --help

# Run Claude in interactive mode with additional arguments
ccctx run -- --version

# Run Claude with model specification
ccctx run -- --model=xxxxx
```

## How It Works

The `run` command executes Claude with the specified context environment variables without affecting your current shell environment.

In interactive mode (when no context name is provided), you can:
- Use arrow keys (↑ ↓) or vim keys (j/k) to navigate between contexts
- Press Enter to select the highlighted context
- Press ESC to cancel the operation
- The interface is now powered by tview for a richer terminal experience

All arguments after the `--` separator are forwarded to Claude, allowing you to use Claude's full functionality.
For example: `ccctx run personal -- --help` or `ccctx run -- --version`

## Environment Variables in Authentication

For enhanced security, you can use environment variables instead of hardcoding authentication tokens in your configuration file. Use the `env:` prefix followed by the environment variable name:

```toml
[context.production]
base_url = "https://api.anthropic.com"
auth_token = "env:ANTHROPIC_API_KEY"
```

When using this format:
- The tool will read the value from the specified environment variable
- If the environment variable is not set or empty, an error will be displayed
- This prevents accidentally committing sensitive tokens to version control

**Important:** Make sure to export the environment variable in your shell:
```bash
export ANTHROPIC_API_KEY="your-api-key-here"
# Or add to your ~/.bashrc file
```

## Environment Variables

- `CCCTX_CONFIG_PATH`: Override the default config file path (`~/.ccctx/config.toml`)
