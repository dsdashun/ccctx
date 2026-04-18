# Architecture User Story Test Report

**Architecture Document:** `_bmad-output/planning-artifacts/architecture.md`
**Epic/Story Document:** `_bmad-output/planning-artifacts/epics.md`
**Method:** Mental simulation — trace each User Story's input through the designed architecture, verify the output matches acceptance criteria.

---

## Round 1

**Date:** 2026-04-18

### Summary

| Severity | Count | Description |
|----------|-------|-------------|
| **Critical** | 2 | Design flaw that would produce incorrect output or break existing behavior |
| **High** | 3 | Ambiguity or missing specification that blocks implementation |
| **Medium** | 3 | Edge case or inconsistency that could cause subtle bugs |
| **Low** | 2 | Minor gap, easily resolved during implementation |

---

### Story-by-Story Test Results

---

#### Story 1.1: Extract Shared Execution Kernel into `internal/runner/`

**Status:** PASS with notes

**Trace:** Developer extracts logic from `run.go` into `internal/runner/runner.go` and `internal/runner/args.go`.

The extraction map is clear and complete:

| Source (run.go) | Target | Status |
|---|---|---|
| Lines 154-178 (env building) | `runner.buildEnv()` | Clear |
| Lines 147-151 (LookPath) | Stays in `cmd/run.go` | Clear |
| Lines 180-196 (exec + exit code) | `runner.Run()` | **See Issue M1** |
| Lines 22-137 (arg parsing + TUI) | `runner.ParseArgs()` | **See Issue C2** |

**Notes:**
- The extraction map covers all logic in run.go. No logic is orphaned.
- The boundary rule "runner never exits, commands never contain business logic" is well-defined.
- `config.GetContext()` is called inside `runner.New()`, which imports `config/` — this respects the boundary rules.

---

#### Story 1.2: Refactor `run` Command as Thin Wrapper

**Status:** FAIL — Critical behavioral regression

##### Test Case 1.2.1: `ccctx run provider-A`

**Trace:**
1. `runner.ParseArgs(["provider-A"])` → no `--`, has provider → `provider="provider-A"`, `targetArgs=[]`, `useTUI=false`
2. No TUI needed
3. `LookPath("claude")` → `/usr/local/bin/claude`
4. `targetArgs` empty → `targetArgs = ["/usr/local/bin/claude"]`
5. `runner.New({ContextName:"provider-A", Target:["/usr/local/bin/claude"]})` → resolves config, builds env
6. `r.Run()` → executes claude with provider env vars

**Expected:** Claude launches with provider-A env vars
**Result:** PASS

##### Test Case 1.2.2: `ccctx run provider-A -- --help`

**Trace:**
1. `runner.ParseArgs(["provider-A", "--", "--help"])` → `provider="provider-A"`, `targetArgs=["--help"]`, `useTUI=false`
2. `LookPath("claude")` → `/usr/local/bin/claude`
3. `targetArgs` not empty → `targetArgs = ["/usr/local/bin/claude", "--help"]`
4. `runner.New(...)` → resolves, builds env
5. `r.Run()` → executes `claude --help` with provider env vars

**Expected:** Claude receives `--help` flag
**Result:** PASS

##### Test Case 1.2.3: `ccctx run provider-A somearg` (no `--` separator)

**Current behavior (run.go:64-84):**
1. No `--`, check if "provider-A" matches a context → yes
2. `contextName = "provider-A"`, `claudeArgs = ["somearg"]`
3. Claude runs with `["somearg"]`

**New behavior (architecture ParseArgs):**
1. `runner.ParseArgs(["provider-A", "somearg"])` → no `--`, "provider arg" case per table → `provider="provider-A"`, `targetArgs=[]`, `useTUI=false`
2. `targetArgs` empty → `targetArgs = ["/usr/local/bin/claude"]`
3. Claude runs with NO args. **`somearg` is LOST.**

**Expected:** Claude receives `somearg`
**Result:** FAIL — **Critical behavioral regression**

> **Issue C1: Argument forwarding without `--` is broken**
>
> The architecture's ParseArgs table specifies that for the "No `--`, provider arg" case, `targetArgs=[]`. This means all arguments after the provider name are silently dropped when no `--` separator is present.
>
> The current `run.go` (lines 64-84) forwards remaining args to Claude when the first arg matches a known context. The new design loses this behavior.
>
> **Impact:** Any user relying on `ccctx run provider-A somearg` (without `--`) will have their args silently ignored. This is a breaking change not called out in the architecture.
>
> **Possible fix:** Either (a) update ParseArgs to return remaining args as targetArgs when a provider is found without `--`, or (b) explicitly document this as a deliberate behavioral change with a migration note.

##### Test Case 1.2.4: `ccctx run` (no args, TUI mode)

**Trace:**
1. `runner.ParseArgs([])` → no args → `provider=""`, `targetArgs=[]`, `useTUI=true`
2. `ui.RunContextSelector()` → user selects "provider-A"
3. `LookPath("claude")` → `/usr/local/bin/claude`
4. `targetArgs = ["/usr/local/bin/claude"]`
5. `runner.New(...)` → resolves, builds env
6. `r.Run()` → executes claude

**Expected:** Claude launches after TUI selection
**Result:** PASS

##### Test Case 1.2.5: `ccctx run nonexistent` (provider not in config)

**Current behavior (run.go:73-84):**
1. No `--`, check if "nonexistent" matches a context → no
2. All args treated as claude args, TUI launched
3. User selects a real provider from TUI
4. Claude runs with `["nonexistent"]` forwarded

**New behavior (architecture ParseArgs):**
1. `runner.ParseArgs(["nonexistent"])` → no `--`, provider arg case → `provider="nonexistent"`, `targetArgs=[]`, `useTUI=false`
2. No TUI (useTUI=false)
3. `runner.New({ContextName:"nonexistent", ...})` → calls `config.GetContext("nonexistent")` → returns error: `context 'nonexistent' not found`
4. Command prints error, exits 1

**Expected (current):** TUI fallback, then forward arg to Claude
**Result (new):** Hard error. **Different behavior.**

> **Issue C2: ParseArgs cannot distinguish provider names from forwarded args**
>
> The current `run.go` validates the first arg against known context names (lines 73-84). If it doesn't match, it falls back to TUI mode and forwards all args to Claude. The new `ParseArgs` is command-agnostic and has no access to config, so it always treats the first arg as a provider name. This removes the "unknown arg → TUI fallback" behavior.
>
> **Impact:** Users who accidentally typo a context name will get a hard error instead of TUI fallback. This may be acceptable as a deliberate simplification, but it should be explicitly documented.

---

#### Story 1.3: Implement `exec` Subcommand

**Status:** PASS with issues

##### Test Case 1.3.1: `ccctx exec provider-A` → shell with env vars

**Trace:**
1. `runner.ParseArgs(["provider-A"])` → `provider="provider-A"`, `targetArgs=[]`, `useTUI=false`
2. `targetArgs` empty → `targetArgs = [os.Getenv("SHELL")]` → `["/bin/bash"]`
3. `runner.New({ContextName:"provider-A", Target:["/bin/bash"]})` → resolves config, builds env
4. `r.Run()` → `exec.Command("/bin/bash")` with modified env, stdin/stdout/stderr connected
5. User gets interactive shell with ANTHROPIC_* env vars

**Expected:** Shell with provider-A env vars
**Result:** PASS

##### Test Case 1.3.2: `ccctx exec provider-A -- env | grep ANTHROPIC`

**Trace:**
1. Shell parses: `ccctx exec provider-A -- env` with stdout piped to `grep ANTHROPIC`
2. `runner.ParseArgs(["provider-A", "--", "env"])` → `provider="provider-A"`, `targetArgs=["env"]`, `useTUI=false`
3. `targetArgs` not empty → stays `["env"]`
4. `runner.New({ContextName:"provider-A", Target:["env"]})` → builds env
5. `r.Run()` → executes `env` with modified env, stdout goes to parent shell's pipe → `grep ANTHROPIC`

**Expected:** `ANTHROPIC_BASE_URL=...` and `ANTHROPIC_AUTH_TOKEN=...` in output
**Result:** PASS

##### Test Case 1.3.3: `ccctx exec provider-A -- bash -c "echo hello"`

**Trace:**
1. Shell parses: `ccctx exec provider-A -- bash -c "echo hello"` (shell handles quoting)
2. `runner.ParseArgs(["provider-A", "--", "bash", "-c", "echo hello"])` → `provider="provider-A"`, `targetArgs=["bash", "-c", "echo hello"]`, `useTUI=false`
3. `runner.New(...)` → builds env
4. `r.Run()` → `exec.Command("bash", "-c", "echo hello")` with modified env
5. Output: `hello`

**Expected:** `hello`
**Result:** PASS

##### Test Case 1.3.4: `$SHELL` not set

**Trace:**
1. `ccctx exec provider-A`
2. `targetArgs` empty → `targetArgs = [os.Getenv("SHELL")]` → `[""]` (empty string)
3. `exec.Command("")` → error: `exec: already running`

**Expected:** Graceful error message
**Result:** FAIL — unhandled edge case

> **Issue H1: Empty `$SHELL` environment variable not handled**
>
> `os.Getenv("SHELL")` returns empty string if not set. `exec.Command("")` will fail with a confusing error. The architecture doesn't specify error handling for this case.
>
> **Impact:** Users without `$SHELL` set get an obscure Go runtime error instead of a clear message.
>
> **Fix:** Add validation in `cmd/exec.go`: if `os.Getenv("SHELL") == ""`, print `"Error: SHELL environment variable not set\n"` and exit 1.

##### Test Case 1.3.5: Command not found

**Trace:**
1. `ccctx exec provider-A -- nonexistent-cmd`
2. `runner.New(...)` succeeds (config resolves fine)
3. `r.Run()` → `exec.Command("nonexistent-cmd")` → `cmd.Run()` returns error (NOT `*exec.ExitError`)
4. Run() needs to return an exit code, but the error isn't an ExitError

**Expected:** Clear error message, exit code 1
**Result:** AMBIGUOUS — architecture doesn't specify this case

> **Issue H2: `Run()` return type can't distinguish command start failure from exit code**
>
> `Run()` returns `int` (exit code). When the command fails to start (e.g., binary not found), `cmd.Run()` returns a generic `error`, not `*exec.ExitError`. The runner can't print the error (no I/O rule) and can't return it (signature is `int` only).
>
> The current `run.go` handles this correctly (lines 188-195) because it has access to stderr. But in the new design, this logic moves to the runner, which can't output errors.
>
> **Impact:** Command start failures are silently converted to exit code 1 with no error message. User sees `ccctx` exit with 1 but doesn't know why.
>
> **Fix:** Change `Run()` signature to `Run() (int, error)`. Return `(0, nil)` for success, `(exitCode, nil)` for normal exit, `(1, err)` for start failure. Command files check for error and print it.

##### Test Case 1.3.6: Exit code passthrough

**Trace:**
1. `ccctx exec provider-A -- bash -c "exit 42"`
2. `r.Run()` → `exec.Command("bash", "-c", "exit 42")` → `cmd.Run()` returns `*exec.ExitError` with code 42
3. `Run()` extracts exit code via `exitError.ExitCode()` → returns 42
4. `cmd/exec.go` calls `os.Exit(42)`

**Expected:** Exit code 42
**Result:** PASS (assuming Issue H2 is resolved for ExitError case)

##### Test Case 1.3.7: ANTHROPIC_* override (existing vars in environment)

**Trace:**
1. Parent shell has `ANTHROPIC_BASE_URL=https://old.example.com`
2. `ccctx exec provider-A -- env`
3. `buildEnv()`:
   - Start with `os.Environ()` → includes `ANTHROPIC_BASE_URL=https://old.example.com`
   - Filter: `strings.HasPrefix(e, "ANTHROPIC_")` → strips all ANTHROPIC_* vars
   - Inject: `ANTHROPIC_BASE_URL=<provider-A's url>`, `ANTHROPIC_AUTH_TOKEN=<provider-A's token>`
4. `cmd.Env = filteredEnv + injectedVars`
5. `env` output shows only provider-A's values

**Expected:** Old ANTHROPIC_* values replaced by provider-A's
**Result:** PASS

---

#### Story 1.4: Fix Config File Permissions

**Status:** PASS with note

##### Test Case 1.4.1: New config file creation

**Trace:**
1. `~/.ccctx/` doesn't exist
2. `config.LoadConfig()` → `os.MkdirAll(dir, 0755)` → creates directory
3. `os.WriteFile(configPath, content, 0600)` → creates file with 0600 perms

**Expected:** File permissions 0600
**Result:** PASS

##### Test Case 1.4.2: Existing config file (no permission change)

**Trace:**
1. `~/.ccctx/config.toml` exists with 0644
2. `config.LoadConfig()` → file exists, skips WriteFile → permissions unchanged

**Expected:** Permissions not modified
**Result:** PASS

##### Test Case 1.4.3: Config directory permissions

**Trace:**
1. `os.MkdirAll(dir, 0755)` → directory created with 0755 (world-readable)
2. Other users can `ls ~/.ccctx/` to see that `config.toml` exists
3. But can't read the file (0600) → token is safe

**Result:** PASS (file content is protected; directory listing is a minor info leak but acceptable)

---

#### Story 2.1: `exec` Command Integrates TUI Interactive Selector

**Status:** FAIL — exit code conflict

##### Test Case 2.1.1: `ccctx exec` (no args) → TUI → select provider → shell

**Trace:**
1. `runner.ParseArgs([])` → `provider=""`, `targetArgs=[]`, `useTUI=true`
2. TUI launched → user selects "provider-B"
3. `targetArgs` empty → `targetArgs = [os.Getenv("SHELL")]`
4. `runner.New({ContextName:"provider-B", Target:["/bin/bash"]})` → builds env
5. `r.Run()` → executes shell with provider-B env vars

**Expected:** Shell with provider-B's ANTHROPIC_* vars
**Result:** PASS

##### Test Case 2.1.2: TUI ESC → cancel

**Trace:**
1. `runner.ParseArgs([])` → `useTUI=true`
2. TUI launched → user presses ESC
3. `ui.RunContextSelector()` returns `("", fmt.Errorf("operation cancelled"))`
4. `cmd/exec.go` handles the error

**Architecture says (Consistency Rules):**
> `"operation cancelled" from TUI → print "Operation cancelled." to stdout, return (no error exit)`

**Story 2.1 acceptance criteria says:**
> "按 ESC 键 → 取消操作，退出码为 1，不执行任何命令"

**Contradiction:** Architecture implies exit code 0 (just return), story requires exit code 1.

**Result:** FAIL

> **Issue H3: TUI cancellation exit code is contradictory**
>
> The architecture's "Cancellation handling" rule says `"operation cancelled" → print, return (no error exit)`, which implies exit code 0 (default for Cobra commands). But Story 2.1 explicitly requires exit code 1.
>
> **Impact:** Implementation ambiguity — different developers could implement either behavior.
>
> **Fix:** Align the architecture rule with the story: cancellation should print `"Operation cancelled."` and exit with code 1. Update the consistency rule to: `"operation cancelled" from TUI → print "Operation cancelled." to stdout, os.Exit(1)`.

##### Test Case 2.1.3: `ccctx run` (no args) → TUI regression test

**Trace:**
1. Same flow as 2.1.1 but via `cmd/run.go`
2. TUI launches, user selects, claude runs
3. Behavior identical to pre-refactor

**Expected:** No regression
**Result:** PASS (assuming ParseArgs issue C1 is resolved or accepted)

---

#### Story 3.1: Add `--model` Override Flags (Phase 2)

**Status:** FAIL — architectural conflict

##### Test Case 3.1.1: `ccctx run provider-A --model claude-opus-4-7`

**Trace:**
1. Cobra parses `--model claude-opus-4-7` as a registered flag → `modelFlag = "claude-opus-4-7"`
2. Remaining args passed to Run: `["provider-A"]`
3. `runner.ParseArgs(["provider-A"])` → `provider="provider-A"`, `targetArgs=[]`, `useTUI=false`
4. `runner.New({..., Model:"claude-opus-4-7"})` → in `buildEnv()`, Model overrides config
5. Claude runs with `ANTHROPIC_MODEL=claude-opus-4-7`

**Expected:** `ANTHROPIC_MODEL=claude-opus-4-7`
**Result:** PASS (structurally supported, priority logic is inferable)

##### Test Case 3.1.2: `ccctx exec provider-A --model claude-opus-4-7 -- env | grep ANTHROPIC_MODEL`

**Trace:**
1. Cobra parses flags: `--model claude-opus-4-7` consumed as flag
2. Cobra encounters `--` → flag terminator, all remaining args are positional
3. Cobra passes to Run: `["provider-A", "env"]` (the `--` is consumed by Cobra, not passed through)
4. `runner.ParseArgs(["provider-A", "env"])` → no `--` found, "provider arg" case → `provider="provider-A"`, `targetArgs=[]`
5. **`"env"` is LOST**

**Expected:** `ANTHROPIC_MODEL=claude-opus-4-7` in output
**Result:** FAIL

> **Issue C3: `--model` Cobra flag breaks `--` separator mechanism**
>
> When `--model` is registered as a Cobra flag, Cobra treats `--` as a standard flag terminator and **consumes it** — it is NOT passed through to the Run function's args. This means `ParseArgs` never sees the `--` separator and cannot split provider from command.
>
> This is a fundamental conflict between Cobra's flag parsing and the architecture's manual `--` separator handling.
>
> **Impact:** Phase 2 is not implementable as designed. Any use of `--` after `--model` will break.
>
> **Fix options:**
> 1. Use a non-`--` separator (e.g., `//` or `---`) that Cobra won't consume
> 2. Disable Cobra's flag parsing entirely (`cmd.DisableFlagParsing = true`) and parse flags manually
> 3. Use Cobra's `FlagParsing` with `TraverseChildren` and custom arg handling
> 4. Restrict `--model` to appear ONLY before the provider name, and parse everything after the provider as command args
>
> This needs resolution before Phase 2 design is finalized.

##### Test Case 3.1.3: CLI flag priority over config

**Trace:**
1. Provider-A config has `model = "claude-sonnet-4-6"`
2. `ccctx run provider-A --model claude-opus-4-7`
3. `buildEnv()` checks: Options.Model ("claude-opus-4-7") != "" → use it
4. Config's model ("claude-sonnet-4-6") is ignored

**Expected:** CLI flag wins
**Result:** PASS (inferable from struct design, but logic is not explicitly written out)

> **Issue M2: `buildEnv()` model priority logic not explicitly specified**
>
> The architecture defines the Options struct with Model/SmallFastModel fields and the rule "CLI flag > config", but doesn't show the actual code logic inside `buildEnv()`. Specifically: what happens when Options.Model is empty and config has a model? The intended behavior is "use config value", but this should be explicitly stated in the injection order pattern.
>
> **Fix:** Add to "Environment Variable Construction Patterns":
> ```
> Priority: Options.Model > ctx.Model > omit
> If Options.Model != "" → inject Options.Model
> Else if ctx.Model != "" → inject ctx.Model
> Else → don't set ANTHROPIC_MODEL
> ```

##### Test Case 3.1.4: `--model` with no config model

**Trace:**
1. Provider-A config has no `model` field
2. `ccctx exec provider-A --model claude-opus-4-7 -- env | grep ANTHROPIC_MODEL`
3. (Assuming Issue C3 is resolved) → `ANTHROPIC_MODEL=claude-opus-4-7`

**Expected:** `ANTHROPIC_MODEL=claude-opus-4-7`
**Result:** PASS (if C3 resolved)

##### Test Case 3.1.5: No `--model` flag, no config model

**Trace:**
1. Provider-A config has no `model` field
2. `ccctx exec provider-A -- env`
3. `buildEnv()` → Options.Model = "", ctx.Model = "" → don't inject ANTHROPIC_MODEL
4. `env` output does NOT include ANTHROPIC_MODEL

**Expected:** No ANTHROPIC_MODEL in output
**Result:** PASS

---

### Cross-Story Issues

#### Issue M3: `handleError` helper referenced but undefined

The command file templates reference `handleError(err)` but this function is never defined. The architecture acknowledges this as a "minor gap" but it affects error handling consistency across both `exec.go` and `run.go`.

Without `handleError`, each command file needs inline error handling, which duplicates the pattern from the current `run.go`. Since the architecture's goal is to reduce duplication, a small helper in each command file (or a shared internal package) would be better.

**Recommendation:** Define inline in each command file as a closure:
```go
handleError := func(err error) {
    if err.Error() == "operation cancelled" {
        fmt.Println("Operation cancelled.")
        os.Exit(1) // per Story 2.1
        return
    }
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

#### Issue L1: Signal handling not discussed

If the child process is killed by a signal (SIGTERM, SIGKILL), `exec.ExitError.ExitCode()` returns -1 on some platforms. The architecture doesn't specify how to handle this.

**Impact:** Minor — exit code -1 is unusual but `os.Exit(-1)` is valid in Go. Most users won't notice.

#### Issue L2: ParseArgs behavior with multiple args before `--`

The parsing rules table covers 4 cases but doesn't address: multiple args before `--` separator.

Example: `ccctx exec foo bar -- cmd`
- ParseArgs receives `["foo", "bar", "--", "cmd"]`
- Table says "provider before --" but doesn't specify: is provider "foo" or "foo bar"?
- Likely behavior: provider = "foo", "bar" is ignored → confusing

**Recommendation:** Either (a) treat only the first arg before `--` as provider and error if more exist, or (b) document that only the first arg is used and others are silently ignored.

---

### Requirements Coverage Audit

#### Functional Requirements

| FR | Covered? | Notes |
|----|----------|-------|
| FR1-FR12 | ✅ Preserved | Architecture maintains all existing behavior |
| FR13 | ✅ | `ccctx exec <provider>` → $SHELL with env vars |
| FR14 | ✅ | `ccctx exec <provider> -- <command>` |
| FR15 | ⚠️ | $SHELL default, but empty $SHELL not handled (Issue H1) |
| FR16 | ✅ | Child inherits env via `cmd.Env = newEnv` |
| FR17 | ⚠️ | Exit code passthrough works for normal exits, but command start failure is ambiguous (Issue H2) |
| FR18 | ✅ | `strings.HasPrefix("ANTHROPIC_")` strips all, then injects |
| FR19 | ✅ | `ccctx exec` (no args) → TUI |
| FR20-FR22 | ✅ | Existing TUI, no changes |
| FR23 | ✅ | Same TUI selector used by both commands |
| FR24 | ✅ | Shared runner package |
| FR25 | ⚠️ | Run as thin wrapper, but behavioral regression for args without `--` (Issue C1) |
| FR26 | ✅ | exec uses shared pipeline |
| FR27 | ⚠️ | Phase 2 — struct ready but `--` conflict (Issue C3) |
| FR28 | ⚠️ | Phase 2 — priority logic inferable but not explicit (Issue M2) |

#### Non-Functional Requirements

| NFR | Covered? | Notes |
|-----|----------|-------|
| NFR1 | ✅ | Tokens never in stdout/stderr/logs |
| NFR2 | ✅ | `strings.HasPrefix("ANTHROPIC_")` |
| NFR3 | ✅ | `resolveEnvVar()` at runtime only |
| NFR4 | ✅ | `os.WriteFile(..., 0600)` |
| NFR5 | ✅ | Any POSIX shell via $SHELL |
| NFR6 | ✅ | Generic ANTHROPIC_* env vars |
| NFR7 | ✅ | CGO_ENABLED=0 |

---

### Final Verdict

**Architecture is NOT ready for implementation as-is.** Two critical issues must be resolved first:

#### Must Fix Before Implementation

1. **Issue C2/H2 — `Run()` return type:** Change signature from `Run() int` to `Run() (int, error)` so command start failures can be reported. This is a structural change that affects the interface contract.

2. **Issue H3 — TUI cancellation exit code:** Decide whether cancellation exits 0 or 1, then update both the architecture and story to match.

#### Should Fix Before Implementation

3. **Issue C1 — Arg forwarding without `--`:** Decide whether this is an intentional behavioral change or a bug. If intentional, document it. If not, update ParseArgs to handle it.

4. **Issue H1 — Empty `$SHELL` handling:** Add validation in `cmd/exec.go`.

#### Can Fix During Implementation

5. **Issue M3 — `handleError` helper:** Define during implementation.
6. **Issue L2 — Multiple args before `--`:** Add validation or documentation.
7. **Issue M2 — Model priority logic:** Clarify in buildEnv() comments.

#### Must Fix Before Phase 2

8. **Issue C3 — `--model` flag breaks `--` separator:** This is the most critical long-term issue. The architecture needs a concrete solution for reconciling Cobra flag parsing with the `--` separator mechanism.

---

## Round 2

**Date:** 2026-04-18
**Context:** Second review pass after architecture was updated to address Round 1 issues. All stories re-traced against the updated architecture. Current source code (run.go, config.go, selector.go, main.go) was cross-referenced to verify extraction accuracy and identify undocumented behavioral changes.

### Round 1 Issues Resolution Status

| Issue | Severity | Status | Evidence |
|-------|----------|--------|----------|
| C1: Arg forwarding without `--` | Critical | ✅ Resolved | ParseArgs now returns `args[1:]` as targetArgs for "No `--`, multiple args" case (architecture line 328) |
| C2: Provider name vs forwarded args | Critical | ✅ Resolved | Explicitly documented as breaking change — first arg always treated as provider (line 745) |
| C3: `--model` vs `--` separator | Critical | ⚠️ Acknowledged | Deferred to Phase 2 with resolution strategies noted (lines 146-147). Acceptable for Phase 1. |
| H1: Empty `$SHELL` | High | ✅ Resolved | Validation added: `shell == ""` check with clear error message (lines 384-388) |
| H2: `Run()` return type | High | ✅ Resolved | Signature changed to `(int, error)` (line 637) |
| H3: TUI cancellation exit code | High | ✅ Resolved | Explicitly set to `os.Exit(1)` (lines 305, 640) |
| M1: Related to H2 | Medium | ✅ Resolved | Covered by H2 |
| M2: Model priority logic | Medium | ✅ Resolved | Explicit priority chain: `Options.Model > ctx.Model > omit` (lines 343-348) |
| M3: `handleError` helper | Medium | ✅ Resolved | Replaced with inline error handling (line 684) |
| L1: Signal handling (exit code -1) | Low | ❌ Unresolved | No mention in updated architecture |
| L2: Multiple args before `--` | Low | ✅ Resolved | Explicitly returns error (line 221, 639) |

**Summary:** 9/11 resolved, 1 acknowledged (C3 deferred), 1 unresolved (L1).

---

### Round 2 Test Results

---

#### Story 1.1: Extract Shared Execution Kernel into `internal/runner/`

**Status:** PASS with notes

**Extraction map verification against current `run.go`:**

| Current location (run.go) | Target in architecture | Verified |
|---|---|---|
| Lines 155-178 (env building) | `runner.buildEnv()` | ✅ Covers ANTHROPIC_* filtering + injection |
| Lines 147-151 (LookPath) | Stays in `cmd/run.go` | ✅ Correct — exec doesn't use LookPath |
| Lines 180-196 (exec + exit code) | `runner.Run()` | ✅ ExitError handling extracted (see Issue R2-4) |
| Lines 22-137 (arg parsing + TUI) | `runner.ParseArgs()` + TUI in cmd files | ✅ TUI call stays in cmd files, arg parsing extracted |

**Notes:**
- The extraction covers all logic in run.go. No orphaned code.
- The `config.GetContext()` call (run.go:140) moves into `runner.New()`, which is correct — it imports `config/` as permitted by boundary rules.
- The `resolveEnvVar()` call inside `config.GetContext()` (config.go:123) stays in the config package. Correct.

---

#### Story 1.2: Refactor `run` Command as Thin Wrapper

**Status:** PASS with notes

##### Test Case 1.2.1: `ccctx run provider-A`

**Trace:**
1. `runner.ParseArgs(["provider-A"])` → no `--`, one arg → `provider="provider-A"`, `targetArgs=[]`, `useTUI=false`
2. `LookPath("claude")` → `/usr/local/bin/claude`
3. `targetArgs` empty → `targetArgs = ["/usr/local/bin/claude"]`
4. `runner.New({ContextName:"provider-A", Target:["/usr/local/bin/claude"]})` → resolves config, builds env
5. `r.Run()` → `(0, nil)` → `os.Exit(0)`

**Expected:** Claude launches with provider-A env vars
**Result:** PASS

##### Test Case 1.2.2: `ccctx run provider-A -- --help`

**Trace:**
1. `runner.ParseArgs(["provider-A", "--", "--help"])` → `provider="provider-A"`, `targetArgs=["--help"]`, `useTUI=false`
2. `LookPath("claude")` → `/usr/local/bin/claude`
3. `targetArgs` not empty → `targetArgs = ["/usr/local/bin/claude", "--help"]`
4. `runner.New(...)` → builds env
5. `r.Run()` → Claude receives `--help`

**Expected:** Claude receives `--help` flag
**Result:** PASS

##### Test Case 1.2.3: `ccctx run provider-A somearg` (no `--` separator)

**Current behavior (run.go:81-84):**
- First arg matches known context → `contextName="provider-A"`, `claudeArgs=["somearg"]`
- Claude receives `somearg`

**New behavior (architecture ParseArgs):**
1. `runner.ParseArgs(["provider-A", "somearg"])` → no `--`, multiple args → `provider="provider-A"`, `targetArgs=["somearg"]`, `useTUI=false`
2. `LookPath("claude")` → `/usr/local/bin/claude`
3. `targetArgs` not empty → `targetArgs = ["/usr/local/bin/claude", "somearg"]`
4. Claude receives `somearg`

**Expected:** Claude receives `somearg`
**Result:** PASS — Round 1 Issue C1 is resolved

##### Test Case 1.2.4: `ccctx run` (no args, TUI mode)

**Trace:**
1. `runner.ParseArgs([])` → no args → `provider=""`, `targetArgs=[]`, `useTUI=true`
2. `ui.RunContextSelector()` → user selects "provider-A"
3. `LookPath("claude")` → `/usr/local/bin/claude`
4. `targetArgs = ["/usr/local/bin/claude"]`
5. `runner.New(...)` → resolves, builds env
6. `r.Run()` → Claude launches

**Expected:** Claude launches after TUI selection
**Result:** PASS

##### Test Case 1.2.5: `ccctx run nonexistent` (unknown provider)

**Current behavior (run.go:85-111):**
- First arg doesn't match known context → TUI fallback → all args forwarded to Claude

**New behavior (architecture):**
1. `runner.ParseArgs(["nonexistent"])` → `provider="nonexistent"`, `targetArgs=[]`, `useTUI=false`
2. No TUI (useTUI=false)
3. `runner.New({ContextName:"nonexistent", ...})` → `config.GetContext("nonexistent")` → error
4. `fmt.Fprintf(os.Stderr, "Error: context 'nonexistent' not found\n")`, `os.Exit(1)`

**Expected (current):** TUI fallback
**Result (new):** Hard error

**Verdict:** Documented as intentional breaking change (architecture line 745, 329). PASS — accepted design decision.

##### Test Case 1.2.6: `ccctx run -- --help` (no provider, with `--`)

**Trace:**
1. `runner.ParseArgs(["--", "--help"])` → `--` present, no provider before `--` → `provider=""`, `targetArgs=["--help"]`, `useTUI=true`
2. TUI launches → user selects provider
3. `LookPath("claude")` → `/usr/local/bin/claude`
4. `targetArgs` not empty → `targetArgs = ["/usr/local/bin/claude", "--help"]`
5. Claude receives `--help`

**Expected:** User picks provider interactively, then Claude gets `--help`
**Result:** PASS

##### Test Case 1.2.7: `ccctx run --` (just `--`, nothing after)

**Trace:**
1. `runner.ParseArgs(["--"])` → `--` present, no provider before `--`, nothing after → `provider=""`, `targetArgs=[]`, `useTUI=true`
2. TUI launches → user selects provider
3. `LookPath("claude")` → `/usr/local/bin/claude`
4. `targetArgs = ["/usr/local/bin/claude"]`
5. Claude launches normally

**Expected:** Equivalent to `ccctx run` (no args)
**Result:** PASS

##### Test Case 1.2.8: TUI cancellation in run

**Current behavior (run.go:56-57):** `fmt.Println("Operation cancelled.")` → `return` → exit code 0

**New behavior (architecture):** `fmt.Println("Operation cancelled.")` → `os.Exit(1)` → exit code 1

**Expected (story 2.1):** Exit code 1
**Result:** PASS — but this is a behavioral change from current code (exit 0 → exit 1). See Issue R2-2.

---

#### Story 1.3: Implement `exec` Subcommand

**Status:** PASS with issues

##### Test Case 1.3.1: `ccctx exec provider-A` → shell with env vars

**Trace:**
1. `runner.ParseArgs(["provider-A"])` → `provider="provider-A"`, `targetArgs=[]`, `useTUI=false`
2. `targetArgs` empty → `shell := os.Getenv("SHELL")` → `"/bin/bash"` → `targetArgs = ["/bin/bash"]`
3. `runner.New({ContextName:"provider-A", Target:["/bin/bash"]})` → resolves config, builds env
4. `r.Run()`:
   - `exec.Command("/bin/bash")` with modified env
   - `cmd.Stdin/Stdout/Stderr = os.Stdin/Stdout/Stderr`
   - `cmd.Run()` → interactive shell with ANTHROPIC_* env vars
   - Return `(exitCode, nil)` or `(exitCode, nil)` for ExitError

**Expected:** Shell with provider-A env vars
**Result:** PASS

##### Test Case 1.3.2: `ccctx exec provider-A -- env | grep ANTHROPIC`

**Trace:**
1. Shell parses: `ccctx exec provider-A -- env` piped to `grep ANTHROPIC`
2. `runner.ParseArgs(["provider-A", "--", "env"])` → `provider="provider-A"`, `targetArgs=["env"]`, `useTUI=false`
3. `runner.New({ContextName:"provider-A", Target:["env"]})` → builds env
4. `r.Run()` → `exec.Command("env")` with modified env → stdout piped through grep

**Expected:** ANTHROPIC_BASE_URL and ANTHROPIC_AUTH_TOKEN in output
**Result:** PASS

##### Test Case 1.3.3: `ccctx exec provider-A -- bash -c "echo hello"`

**Trace:**
1. `runner.ParseArgs(["provider-A", "--", "bash", "-c", "echo hello"])` → `targetArgs=["bash", "-c", "echo hello"]`
2. `r.Run()` → `exec.Command("bash", "-c", "echo hello")` with modified env
3. Output: `hello`

**Expected:** `hello`
**Result:** PASS

##### Test Case 1.3.4: `$SHELL` not set

**Trace:**
1. `ccctx exec provider-A` → `targetArgs=[]` → `os.Getenv("SHELL")` → `""`
2. `shell == ""` → `fmt.Fprintf(os.Stderr, "Error: SHELL environment variable not set\n")` → `os.Exit(1)`

**Expected:** Graceful error message
**Result:** PASS — Round 1 Issue H1 is resolved

##### Test Case 1.3.5: Command not found

**Trace:**
1. `ccctx exec provider-A -- nonexistent-cmd`
2. `runner.New(...)` succeeds (config resolves fine)
3. `r.Run()` → `exec.Command("nonexistent-cmd")` → `cmd.Run()` returns error (NOT `*exec.ExitError`)
4. Run() returns `(1, err)` (start failure)
5. Command file: `err != nil` → `fmt.Fprintf(os.Stderr, "Error: %v\n", err)` → `os.Exit(1)`

**Expected:** Clear error message, exit code 1
**Result:** PASS — Round 1 Issue H2 is resolved

**Note:** The actual `Run()` implementation logic for distinguishing `*exec.ExitError` from start failures is not shown in the architecture. The architecture specifies the interface `(int, error)` but not the internal type assertion logic. See Issue R2-4.

##### Test Case 1.3.6: Exit code passthrough

**Trace:**
1. `ccctx exec provider-A -- bash -c "exit 42"`
2. `r.Run()` → `exec.Command("bash", "-c", "exit 42")` → `cmd.Run()` returns `*exec.ExitError`
3. Run() detects ExitError → `exitError.ExitCode()` → returns `(42, nil)`
4. Command file: `os.Exit(42)`

**Expected:** Exit code 42
**Result:** PASS

##### Test Case 1.3.7: ANTHROPIC_* override

**Trace:**
1. Parent has `ANTHROPIC_BASE_URL=https://old.example.com`
2. `ccctx exec provider-A -- env`
3. `buildEnv()`: filter `strings.HasPrefix(e, "ANTHROPIC_")` → strip all → inject provider-A's values
4. `env` output shows only provider-A's values

**Expected:** Old values replaced
**Result:** PASS

##### Test Case 1.3.8: `ccctx exec -- bash` (no provider, with command)

**Trace:**
1. `runner.ParseArgs(["--", "bash"])` → `--` present, no provider before `--` → `provider=""`, `targetArgs=["bash"]`, `useTUI=true`
2. TUI launches → user selects provider
3. `targetArgs` not empty → stays `["bash"]`
4. `runner.New({ContextName: selectedProvider, Target:["bash"]})`
5. `r.Run()` → `exec.Command("bash")` with selected provider's env

**Expected:** User picks provider interactively, then bash runs with that provider's env vars
**Result:** PASS

##### Test Case 1.3.9: `ccctx exec --` (just `--`, nothing after)

**Trace:**
1. `runner.ParseArgs(["--"])` → `--` present, no provider before `--`, nothing after → `provider=""`, `targetArgs=[]`, `useTUI=true`
2. TUI launches → user selects provider
3. `targetArgs` empty → `shell := os.Getenv("SHELL")` → `targetArgs = ["/bin/bash"]`
4. Shell launches with provider's env vars

**Expected:** Equivalent to `ccctx exec` (no args)
**Result:** PASS

##### Test Case 1.3.10: Multiple args before `--`

**Trace:**
1. `ccctx exec foo bar -- bash`
2. `runner.ParseArgs(["foo", "bar", "--", "bash"])` → `--` present, 2 args before `--`
3. ParseArgs returns error: ambiguous input

**Expected:** Error, reject ambiguous input
**Result:** PASS — Round 1 Issue L2 is resolved

---

#### Story 1.4: Fix Config File Permissions

**Status:** PASS with notes

##### Test Case 1.4.1: New config file creation

**Trace:**
1. `~/.ccctx/` doesn't exist
2. `config.LoadConfig()` → `os.MkdirAll(dir, 0755)` → `os.WriteFile(configPath, content, 0600)`

**Expected:** File permissions 0600
**Result:** PASS

**Note:** Current code uses 0644 (config.go:77). Architecture specifies changing to 0600. This is a one-line change.

##### Test Case 1.4.2: Existing config file

**Trace:**
1. File exists with 0644
2. `config.LoadConfig()` → `os.Stat(configPath)` → exists → skips WriteFile → permissions unchanged

**Expected:** Permissions not modified
**Result:** PASS

##### Test Case 1.4.3: Config directory permissions

**Trace:**
1. `os.MkdirAll(dir, 0755)` → directory world-readable
2. Other users can `ls ~/.ccctx/` to see `config.toml` exists
3. But can't read file content (0600)

**Result:** PASS — file content protected; directory listing is a minor info leak but acceptable.

---

#### Story 2.1: `exec` Command Integrates TUI Interactive Selector

**Status:** PASS with notes

##### Test Case 2.1.1: `ccctx exec` (no args) → TUI → select provider → shell

**Trace:**
1. `runner.ParseArgs([])` → `provider=""`, `targetArgs=[]`, `useTUI=true`
2. TUI launched → user selects "provider-B"
3. `targetArgs` empty → `targetArgs = [os.Getenv("SHELL")]` → `["/bin/bash"]`
4. `runner.New({ContextName:"provider-B", Target:["/bin/bash"]})` → builds env
5. `r.Run()` → shell with provider-B env vars

**Expected:** Shell with provider-B's ANTHROPIC_* vars
**Result:** PASS

##### Test Case 2.1.2: TUI ESC → cancel

**Trace:**
1. ParseArgs → `useTUI=true`
2. TUI → ESC → `ui.RunContextSelector()` returns `("", fmt.Errorf("operation cancelled"))`
3. `err.Error() == "operation cancelled"` → `fmt.Println("Operation cancelled.")` → `os.Exit(1)`

**Expected (Story 2.1):** Cancel, exit code 1
**Result:** PASS — Round 1 Issue H3 is resolved

##### Test Case 2.1.3: `ccctx run` (no args) → TUI regression test

**Trace:**
1. Same flow as 2.1.1 but via `cmd/run.go`
2. TUI launches, user selects, claude runs
3. Behavior identical to pre-refactor

**Expected:** No regression
**Result:** PASS

##### Test Case 2.1.4: TUI with no contexts configured

**Trace:**
1. `ccctx exec` → `useTUI=true`
2. `resolveViaTUI()` → `ui.RunContextSelector()` → `config.ListContexts()` → returns `[]` → error `"no contexts found"`
3. `err.Error()` != `"operation cancelled"` → general error branch
4. `fmt.Fprintf(os.Stderr, "Error: no contexts found\n")` → `os.Exit(1)`

**Current behavior (run.go:48-49, 97-99, 119-124):** Prints `"No contexts found."` to stdout, returns (exit 0).

**Expected (new):** Error to stderr, exit 1
**Result:** Functional behavior correct, but differs from current code. See Issue R2-3.

---

#### Story 3.1: Add `--model` Override Flags (Phase 2)

**Status:** NOT TESTABLE — deferred to Phase 2

##### Test Case 3.1.1: `ccctx run provider-A --model claude-opus-4-7`

**Trace:**
1. If `--model` registered as Cobra flag → Cobra parses it → `modelFlag = "claude-opus-4-7"`
2. Remaining args: `["provider-A"]`
3. `runner.New({..., Model:"claude-opus-4-7"})` → buildEnv uses CLI value
4. `ANTHROPIC_MODEL=claude-opus-4-7`

**Expected:** `ANTHROPIC_MODEL=claude-opus-4-7`
**Result:** PASS (structurally supported)

##### Test Case 3.1.2: `ccctx exec provider-A --model claude-opus-4-7 -- env | grep ANTHROPIC_MODEL`

**Trace:**
1. Cobra parses `--model claude-opus-4-7` as flag
2. Cobra encounters `--` → flag terminator → remaining args: `["provider-A", "env"]`
3. `runner.ParseArgs(["provider-A", "env"])` → no `--` found, multiple args → `provider="provider-A"`, `targetArgs=["env"]`
4. **Wait — Cobra consumed the `--`, so ParseArgs never sees it. `"env"` becomes a forwarded arg, not post-separator content.**
5. Actually: `"env"` is the first arg after the provider, so `targetArgs=["env"]` — this happens to work since `env` is a valid command.
6. But for `ccctx exec provider-A --model x -- bash -c "echo hi"`:
   - Cobra consumes `--model x` and `--` → args to Run: `["provider-A", "bash", "-c", "echo hi"]`
   - ParseArgs: no `--` → `provider="provider-A"`, `targetArgs=["bash", "-c", "echo hi"]`
   - This actually works in this specific case because `bash` is the first forwarded arg.

**Result:** UNRESOLVED — Round 1 Issue C3 remains. The architecture acknowledges this as a Phase 2 concern with documented resolution strategies. Acceptable for Phase 1 implementation.

##### Test Case 3.1.3: CLI flag priority over config

**Trace:**
1. Config has `model = "claude-sonnet-4-6"`
2. CLI `--model claude-opus-4-7` → `Options.Model = "claude-opus-4-7"`
3. `buildEnv()`: `Options.Model != ""` → use `"claude-opus-4-7"` → ignore config value

**Expected:** CLI flag wins
**Result:** PASS — priority chain is clear (architecture lines 343-348)

##### Test Case 3.1.4: No flag, no config model

**Trace:**
1. No `--model`, config has no `model` → `Options.Model = ""`, `ctx.Model = ""`
2. `buildEnv()`: both empty → don't inject `ANTHROPIC_MODEL`

**Expected:** No ANTHROPIC_MODEL in env
**Result:** PASS

---

### New Issues Found (Round 2)

#### Issue R2-1: [MEDIUM] `Run()` internal logic not specified

**Location:** Architecture → `internal/runner/runner.go` → `Run()` method

**Description:**
The architecture defines the `Run()` signature as `(int, error)` and describes what it returns for different cases, but does not show the actual implementation logic for distinguishing `*exec.ExitError` (normal exit with code) from start failures (binary not found, permission denied, etc.).

The current `run.go` (lines 188-195) implements this correctly:
```go
if exitError, ok := err.(*exec.ExitError); ok {
    os.Exit(exitError.ExitCode())
} else {
    fmt.Fprintf(os.Stderr, "Error executing claude: %v\n", err)
    os.Exit(1)
}
```

But in the new architecture, this logic moves to `runner.Run()`, which cannot print errors or exit. The architecture needs to specify that Run() uses `errors.As(err, &exitErr)` to distinguish the two cases:

```go
func (r *Runner) Run() (int, error) {
    // ... setup cmd ...
    err := cmd.Run()
    if err != nil {
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            return exitErr.ExitCode(), nil  // normal non-zero exit
        }
        return 1, err  // start failure
    }
    return 0, nil
}
```

**Impact:** Without this specification, a developer might return `(0, err)` for ExitError cases, swallowing non-zero exit codes. Or return `(code, err)` for all errors, causing command files to print error messages for normal non-zero exits.

**Recommendation:** Add the implementation logic to the Runner struct design section, or add a "Runner implementation notes" subsection.

---

#### Issue R2-2: [LOW] TUI cancellation exit code changed (0 → 1) but not documented as breaking change

**Location:** Architecture → Cancellation handling (line 305)

**Description:**
The current `run.go` handles TUI cancellation with `return` (exit code 0):
```go
if selectorErr.Error() == "operation cancelled" {
    fmt.Println("Operation cancelled.")
    return  // exit 0
}
```

The architecture changes this to `os.Exit(1)`:
```go
if err.Error() == "operation cancelled" {
    fmt.Println("Operation cancelled.")
    os.Exit(1)
}
```

This behavioral change is not listed in the "Behavioral change from current code" note (architecture line 329), which only documents the ParseArgs first-arg-always-provider change.

**Impact:** Scripts that check exit codes after cancellation (`ccctx run; if [ $? -eq 0 ]; then ...`) will behave differently.

**Recommendation:** Add this to the breaking changes documentation alongside the ParseArgs behavioral change.

---

#### Issue R2-3: [LOW] "No contexts found" handling differs from current code

**Location:** Command file templates → `resolveViaTUI()` error handling

**Description:**
The current code handles "no contexts found" specially in three places (run.go:47-49, 97-99, 119-124):
```go
if len(contexts) == 0 {
    fmt.Println("No contexts found.")
    return  // stdout, exit 0
}
```

In the new architecture, `ui.RunContextSelector()` returns `fmt.Errorf("no contexts found")` when no contexts exist. The command template handles this through the general error path:
```go
fmt.Fprintf(os.Stderr, "Error: no contexts found\n")
os.Exit(1)
```

This changes:
- Output destination: stdout → stderr
- Message format: `"No contexts found."` → `"Error: no contexts found"`
- Exit code: 0 → 1

**Impact:** Minor UX change. Users who parse output or check exit codes may notice.

**Recommendation:** Either (a) add explicit handling for "no contexts found" error (print to stdout, exit 0), or (b) document this as an intentional behavioral change.

---

#### Issue R2-4: [LOW] `resolveViaTUI()` function referenced but undefined

**Location:** Command file templates (architecture lines 371-380, 412-423)

**Description:**
Both command templates reference `resolveViaTUI()` but this function is never defined in the architecture. Similar to the old `handleError` issue (Round 1, Issue M3).

The function would need to:
1. Call `config.ListContexts()` to check for contexts
2. Call `ui.RunContextSelector()` to display TUI
3. Return `(string, error)`

**Impact:** Minor — easily resolved during implementation. The logic is clear from context.

**Recommendation:** Either inline the TUI logic in command templates (like current run.go does), or define `resolveViaTUI` as a closure in each command file.

---

#### Issue R2-5: [LOW] `BuildEnv()` vs `buildEnv()` casing in test section

**Location:** Architecture → AD5: Testing Strategy (line 252)

**Description:**
The test scope lists `BuildEnv()` (capital B) as a test target, but the function is defined as `buildEnv()` (lowercase b) in the export patterns section (line 288). Since the test file is in the same package (`internal/runner/`), it can access the unexported function.

**Impact:** Trivial — the test file will reference `buildEnv()` correctly. Just a documentation typo.

**Recommendation:** Change `BuildEnv()` to `buildEnv()` in the test scope section.

---

#### Issue R2-6: [MEDIUM] "Always inject" vs "empty value rule" contradiction for BASE_URL/AUTH_TOKEN

**Location:** Architecture → Environment Variable Construction Patterns (lines 337-341)

**Description:**
The injection order states:
> 1. `ANTHROPIC_BASE_URL` — **always**
> 2. `ANTHROPIC_AUTH_TOKEN` — **always**

But the "Empty value rule" states:
> Never inject env vars with empty values. Check `!= ""` before appending.

These rules contradict each other when a context has empty `base_url` or `auth_token`. For example, a malformed config:
```toml
[context.broken]
base_url = ""
auth_token = ""
```

Should the system inject `ANTHROPIC_BASE_URL=` (empty) or skip injection entirely?

**Impact:** A child process expecting `ANTHROPIC_BASE_URL` to be set would behave differently depending on which rule is followed. With "always" → empty string is set. With "empty value" → not set at all. Some tools may treat empty differently from unset.

**Recommendation:** Clarify that BASE_URL and AUTH_TOKEN should be validated as non-empty in `runner.New()`, returning an error for contexts with missing required fields. This is preferable to silently injecting empty values.

---

#### Issue R2-7: [LOW] Signal-killed child processes produce exit code -1

**Location:** `internal/runner/runner.go` → `Run()` method

**Description:**
Round 1 Issue L1 remains unresolved. When a child process is killed by signal (SIGTERM, SIGKILL), `exitError.ExitCode()` returns -1 on Linux. `os.Exit(-1)` is valid in Go but unusual.

**Impact:** Minor. Most users won't notice exit code -1. Shell `$?` would show 255 (since -1 mod 256 = 255 in most shells).

**Recommendation:** Document this behavior or normalize -1 to 1 in `Run()`.

---

### Cross-Story Verification

#### ParseArgs comprehensive test matrix

| Input | provider | targetArgs | useTUI | Result |
|-------|----------|------------|--------|--------|
| `[]` | `""` | `[]` | true | PASS |
| `["provider-A"]` | `"provider-A"` | `[]` | false | PASS |
| `["provider-A", "somearg"]` | `"provider-A"` | `["somearg"]` | false | PASS |
| `["provider-A", "a", "b"]` | `"provider-A"` | `["a", "b"]` | false | PASS |
| `["--", "cmd"]` | `""` | `["cmd"]` | true | PASS |
| `["--"]` | `""` | `[]` | true | PASS |
| `["provider-A", "--", "cmd"]` | `"provider-A"` | `["cmd"]` | false | PASS |
| `["provider-A", "--"]` | `"provider-A"` | `[]` | false | PASS |
| `["provider-A", "--", "cmd", "arg"]` | `"provider-A"` | `["cmd", "arg"]` | false | PASS |
| `["foo", "bar", "--", "cmd"]` | — | — | — | PASS (error) |

All 10 ParseArgs cases produce correct results.

#### Behavioral Changes Summary

| Scenario | Current Behavior | New Architecture Behavior | Documented? |
|----------|-----------------|--------------------------|-------------|
| `ccctx run nonexistent` | TUI fallback | Hard error | ✅ Yes (line 329) |
| TUI cancellation | Exit 0 | Exit 1 | ❌ No |
| No contexts found | stdout, exit 0 | stderr, exit 1 | ❌ No |
| `ccctx run provider-A somearg` | Forwards `somearg` | Forwards `somearg` | N/A (preserved) |

---

### Updated Requirements Coverage Audit

| FR | Covered? | Round 1 | Round 2 | Notes |
|----|----------|---------|---------|-------|
| FR1-FR12 | ✅ | ✅ | ✅ | Preserved |
| FR13 | ✅ | ✅ | ✅ | `ccctx exec <provider>` → $SHELL |
| FR14 | ✅ | ✅ | ✅ | `ccctx exec <provider> -- <command>` |
| FR15 | ✅ | ⚠️ | ✅ | $SHELL validation added (H1 resolved) |
| FR16 | ✅ | ✅ | ✅ | Child inherits env |
| FR17 | ✅ | ⚠️ | ✅ | (int, error) return (H2 resolved) |
| FR18 | ✅ | ✅ | ✅ | ANTHROPIC_* override via HasPrefix |
| FR19 | ✅ | ✅ | ✅ | TUI for exec |
| FR20-FR22 | ✅ | ✅ | ✅ | Existing TUI unchanged |
| FR23 | ✅ | ✅ | ✅ | TUI shared |
| FR24 | ✅ | ✅ | ✅ | Shared runner |
| FR25 | ✅ | ⚠️ | ✅ | Arg forwarding fixed (C1 resolved) |
| FR26 | ✅ | ✅ | ✅ | exec uses shared pipeline |
| FR27 | ⚠️ | ⚠️ | ⚠️ | Phase 2 — struct ready, -- conflict noted |
| FR28 | ⚠️ | ⚠️ | ⚠️ | Phase 2 — priority logic clear |

| NFR | Covered? | Notes |
|-----|----------|-------|
| NFR1 | ✅ | Tokens never in stdout/stderr |
| NFR2 | ✅ | `strings.HasPrefix("ANTHROPIC_")` |
| NFR3 | ✅ | `resolveEnvVar()` at runtime only |
| NFR4 | ✅ | 0600 on file creation |
| NFR5 | ✅ | Any POSIX shell via $SHELL |
| NFR6 | ✅ | Generic ANTHROPIC_* env vars |
| NFR7 | ✅ | CGO_ENABLED=0 |

---

### Round 2 Final Verdict

**Architecture is READY FOR IMPLEMENTATION with minor improvements recommended.**

All critical and high-severity issues from Round 1 have been resolved. The remaining issues are low to medium severity and can be addressed during implementation:

#### Recommended Before Implementation

1. **R2-1 [MEDIUM]:** Specify `Run()` internal logic — add the `errors.As` / `*exec.ExitError` type check pattern to the architecture document so implementers get this right.

2. **R2-6 [MEDIUM]:** Resolve "always inject" vs "empty value" contradiction — either validate required fields in `runner.New()` or clarify the rule.

#### Can Fix During Implementation

3. **R2-2 [LOW]:** Document TUI cancellation exit code change (0 → 1) as a breaking change.
4. **R2-3 [LOW]:** Decide on "no contexts found" behavior (current stdout/exit-0 vs new stderr/exit-1).
5. **R2-4 [LOW]:** Define or inline `resolveViaTUI()` during implementation.
6. **R2-5 [LOW]:** Fix `BuildEnv()` → `buildEnv()` casing in test section.
7. **R2-7 [LOW]:** Document or normalize exit code -1 for signal-killed processes.

#### Deferred to Phase 2

8. **R2-8 (was C3):** `--model` vs `--` separator conflict — architecture has documented resolution strategies.
