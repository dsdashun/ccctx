# Story 1.4: Fix Config File Permissions

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **ccctx user**,
I want **the config file permissions to be 0600 (user-readable only) when auto-created**,
so that **auth tokens and credentials stored in the config file are not readable by other users**.

## Acceptance Criteria

1. **AC1: New config file has 0600 permissions** — When ccctx auto-creates `config.toml`, the file is created with mode `0600` (owner read/write only). Change `os.WriteFile(configPath, []byte(defaultConfig), 0644)` to `os.WriteFile(configPath, []byte(defaultConfig), 0600)` in `config/config.go:77` (FR4, NFR4)

2. **AC2: Existing config file permissions unchanged** — When loading an already-existing config file, ccctx does NOT modify its permissions. No `os.Chmod` call anywhere (NFR4, architecture AD4)

3. **AC3: Test coverage** — New test in `config/config_test.go` using testify table-driven pattern verifies that auto-created config files have `0600` permissions. Test uses `os.Stat` + `ModePerm` to assert file mode (AD5)

4. **AC4: All tests pass** — `go test ./...` passes, `go vet ./...` clean, `make build` succeeds

## Tasks / Subtasks

- [x] Task 1: Fix permissions in config.go (AC: #1)
  - [x] Change `os.WriteFile(configPath, []byte(defaultConfig), 0644)` to `os.WriteFile(configPath, []byte(defaultConfig), 0600)` at line 77

- [x] Task 2: Add permission test (AC: #3)
  - [x] Add `TestConfigFilePermissions` to `config/config_test.go`
  - [x] Use `t.TempDir()` for isolation
  - [x] Set `CCCTX_CONFIG_PATH` to temp dir path via `t.Setenv()`
  - [x] Call `LoadConfig()` to trigger auto-creation
  - [x] Assert file mode is `0600` using `os.Stat` + `assert.Equal(t, os.FileMode(0600), info.Mode().Perm())`
  - [x] Also test that a pre-existing file with `0644` is NOT changed — create file with `0644` first, then call `LoadConfig()`, assert mode still `0644`

- [x] Task 3: Verify everything (AC: #4)
  - [x] `go test ./...` passes
  - [x] `go vet ./...` clean
  - [x] `make build` succeeds

### Review Findings

- [x] [Review][Patch] umask 与现有文件测试交互 — 非标准 umask 下测试可能失败 [config/config_test.go:138]
- [x] [Review][Patch] 现有文件权限测试未验证"保留"语义 — 缺少 before/after 对比 [config/config_test.go:124-148]
- [x] [Review][Patch] 现有文件测试仅覆盖 `0644` — 缺少其他权限状态 [config/config_test.go:126-128]
- [x] [Review][Defer] 非 "not exist" 的 `os.Stat` 错误被静默忽略 [config/config.go:68] — deferred, pre-existing
- [x] [Review][Defer] 符号链接/TOCTOU/目录/空文件/只读fs 等边界情况 — deferred, pre-existing

## Dev Notes

### The Change

Single-line fix in `config/config.go:77`:

```go
// BEFORE:
if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
// AFTER:
if err := os.WriteFile(configPath, []byte(defaultConfig), 0600); err != nil {
```

That's it. No other code changes needed.

### Why 0600

The config file stores `auth_token` values (some users store tokens directly, not all use `env:` prefix). Mode `0600` ensures only the file owner can read/write. This is a standard security practice for credential files (same as `~/.ssh/` files).

### Why NOT Modify Existing Files

The architecture (AD4) explicitly states: "Do not modify permissions of existing files." Users may have intentionally set different permissions (e.g., `0400` for read-only, or `0640` for group-readable in shared environments). Overwriting their choice would be presumptuous and could break legitimate setups.

### Directory Permissions

The config directory is created with `0755` (line 62). This is acceptable — the directory itself doesn't contain secrets, only the file does. The `0600` file mode is sufficient protection. Do NOT change directory permissions.

### Test Pattern

```go
func TestConfigFilePermissions(t *testing.T) {
    tests := []struct {
        name           string
        createFirst    bool
        initialPerm    os.FileMode
        expectedPerm   os.FileMode
    }{
        {
            name:         "auto-created file has 0600 permissions",
            createFirst:  false,
            expectedPerm: 0600,
        },
        {
            name:         "existing file permissions are not modified",
            createFirst:  true,
            initialPerm:  0644,
            expectedPerm: 0644,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // require := require.New(t)
            // assert := assert.New(t)
            tmpDir := t.TempDir()
            configPath := filepath.Join(tmpDir, "config.toml")

            if tt.createFirst {
                err := os.WriteFile(configPath, []byte("[context.test]\nbase_url = \"https://example.com\"\nauth_token = \"token\"\n"), tt.initialPerm)
                require.NoError(t, err)
            }

            t.Setenv("CCCTX_CONFIG_PATH", configPath)
            _, err := LoadConfig()
            require.NoError(t, err)

            info, err := os.Stat(configPath)
            require.NoError(t, err)
            assert.Equal(t, tt.expectedPerm, info.Mode().Perm())
        })
    }
}
```

### Previous Story Intelligence

From Story 1.3 deferred items:
- "Config file permissions 0644 — Story 1.4 scope" — confirmed this is our target

From Story 1.3 dev notes:
- Fragile string comparison for TUI cancellation, duplicate error handling patterns, etc. — all pre-existing and NOT in scope for this story

### Project Structure Notes

- Modified file: `config/config.go` — single permission constant change (line 77)
- Modified file: `config/config_test.go` — add new test function
- No other files touched

### References

- [Source: config/config.go:77] — current `0644` permission on `os.WriteFile`
- [Source: _bmad-output/planning-artifacts/architecture.md#AD4: Config File Permissions] — "Set file permissions to 0600 when creating a new config file. Do not modify permissions of existing files."
- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.4] — acceptance criteria
- [Source: _bmad-output/project-context.md#Critical Don't-Miss Rules] — "Config file permissions: `0600` on creation (protects auth tokens)"
- [Source: config/config_test.go] — existing test file (uses raw `testing` not testify for resolveEnvVar — new test should use testify per project rules)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.7 (claude-opus-4-7)

### Debug Log References

### Completion Notes List

- Changed `os.WriteFile` permission from `0644` to `0600` in `config/config.go` — auto-created config files now protect auth tokens from other users
- Added `TestConfigFilePermissions` table-driven test with two cases: auto-created file gets `0600`, existing file permissions are preserved
- Test uses testify (`require`/`assert`), `t.TempDir()`, `t.Setenv()` per project standards
- All existing tests pass, `go vet` clean, build succeeds

### File List

- config/config.go (modified — permission constant 0644→0600)
- config/config_test.go (modified — added TestConfigFilePermissions with imports)
