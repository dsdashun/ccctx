# ccctx - Component Inventory

**Date:** 2026-04-17

## Component Catalog

### CLI Commands

| Component | File | Type | Description |
|-----------|------|------|-------------|
| rootCmd | main.go | Root Command | Cobra root command, registers subcommands |
| ListCmd | cmd/list.go | Subcommand | Lists available contexts from config |
| RunCmd | cmd/run.go | Subcommand | Runs Claude with specified context |

### Configuration Components

| Component | File | Type | Exported | Description |
|-----------|------|------|----------|-------------|
| Context | config/config.go | Struct | Yes | Represents a context (base_url, auth_token, model, small_fast_model) |
| Config | config/config.go | Struct | Yes | Top-level config containing context map |
| resolveEnvVar | config/config.go | Function | No | Resolves `env:` prefixed values to OS env vars |
| GetConfigPath | config/config.go | Function | Yes | Returns config file path (env override or default) |
| LoadConfig | config/config.go | Function | Yes | Loads and parses TOML config, creates if missing |
| ListContexts | config/config.go | Function | Yes | Returns list of context names |
| GetContext | config/config.go | Function | Yes | Returns resolved Context by name |

### UI Components

| Component | File | Type | Exported | Description |
|-----------|------|------|----------|-------------|
| RunContextSelector | internal/ui/selector.go | Function | Yes | Entry point for interactive selector |
| runTviewSelector | internal/ui/selector.go | Function | No | Tview-based selector implementation |

## Component Dependencies

```
main.go
├── cmd.ListCmd
│   └── config.ListContexts()
└── cmd.RunCmd
    ├── config.ListContexts()
    ├── config.GetContext()
    ├── ui.RunContextSelector()
    │   ├── config.ListContexts()
    │   └── tview/tcell (external)
    └── os/exec (stdlib)
```

## External Dependencies

| Dependency | Version | Used By | Purpose |
|-----------|---------|---------|---------|
| github.com/spf13/cobra | v1.9.1 | main.go, cmd/ | CLI framework |
| github.com/spf13/viper | v1.20.1 | config/ | Config management |
| github.com/rivo/tview | (latest) | internal/ui/ | Terminal UI |
| github.com/gdamore/tcell/v2 | v2.8.1 | internal/ui/ | Terminal interface |

## Reusable vs Application-Specific

| Component | Category | Reusability |
|-----------|----------|-------------|
| resolveEnvVar | Utility | Potentially reusable — generic `env:` prefix resolver |
| runTviewSelector | Application | Specific to context selection |
| Config/Context structs | Application | Specific to ccctx domain |
| Cobra commands | Application | Specific to ccctx CLI |

---
_Generated using BMAD Method `document-project` workflow_
