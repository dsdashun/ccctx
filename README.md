# ccctx - Claude-Code Context Switcher

A CLI tool to manage and switch between different Claude-Code contexts.

## Features

- List available contexts from a configuration file
- Switch between contexts interactively or by name
- Run claude-code with a specific context temporarily

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
auth_token = "personal-token-here"
```

## Usage

```bash
# List available contexts
ccctx list

# Switch to a context (interactive mode)
ccctx switch

# Switch to a specific context
ccctx switch work

# Run claude-code with a context (interactive mode)
ccctx run

# Run claude-code with a specific context
ccctx run personal
```

## How It Works

When you use `ccctx switch`, it outputs the necessary export commands to set the environment variables in your current shell. You need to use it with `eval`:

```bash
eval $(ccctx switch work)
```

This way, the environment variables `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN` will be set in your current shell session.

For temporary context switching, you can use the `run` command which will execute claude-code with the specified context without affecting your current shell environment.

## Environment Variables

- `CCCTX_CONFIG_PATH`: Override the default config file path (`~/.ccctx/config.toml`)