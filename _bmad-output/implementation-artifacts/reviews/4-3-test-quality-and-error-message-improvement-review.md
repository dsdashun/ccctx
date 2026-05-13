---
reviewed_plan: _bmad-output/implementation-artifacts/4-3-test-quality-and-error-message-improvement.md
reviewer: plan-review skill (Claude)
latest_round: 2
review_rounds:
  - round: 1
    date: 2026-05-13
  - round: 2
    date: 2026-05-13
health:
  structural: green
  procedural: green
  clerical: green
---

# Review: 4-3-test-quality-and-error-message-improvement.md

Reviewed plan: `_bmad-output/implementation-artifacts/4-3-test-quality-and-error-message-improvement.md`

## Round 1

**Context:** Initial full three-dimension review (dry-run execution feasibility, spec-plan alignment, task granularity). Story 4.3 is a code-quality cleanup story — no new user-facing features, purely test improvements and edge-case hardening on top of Stories 4.1/4.2.

### Consolidated Issue List

| # | Severity | Review Dimension | Issue | Suggested Fix |
|---|----------|-----------------|-------|---------------|
| 1 | HIGH | Dry-Run | **TDD ordering violation in Task 2.** Sub-bullets list implementation ("track whether `--haiku-model` was explicitly set") before test writing ("Add test case for the edge case regardless"). This is the only production code change in the plan and the sub-bullet order would lead to test-after-implementation. | Reorder Task 2 sub-bullets: (1) Decide fix vs document, (2) Write failing test first, (3) Implement `haikuModelSet` fix, (4) Verify test passes. Or add explicit instruction: "Write the failing test FIRST before implementing the fix." |
| 2 | MEDIUM | Granularity | **Tasks 1 and 2 both target `args.go`** — if dispatched as separate sub-agents, the second agent reads pre-change state and produces conflicting edits. Task 1 is also mostly verification-only (AC2 likely already satisfied per plan's own Dev Notes at lines 82-84). | Merge Tasks 1 and 2 into a single "Args improvements (AC2 + AC5)" task. The combined task reads `args.go` once, produces coordinated changes, and stays well under budget (~9K tokens). Renumber remaining tasks from 3/4/5/6 → 2/3/4/5. |
| 3 | LOW | Dry-Run | **Dev Notes example inconsist with existing code.** The exec error-path test table (line 132) shows `base_url = "not-a-url"`, but the existing `TestRunRun` in `run_test.go` uses `base_url = "not-a-valid-url"`. Both trigger the same validation error (no behavioral impact), but the divergence adds unnecessary variance between test files. | Change the Dev Notes table entry to `"not-a-valid-url"` to match the canonical pattern in `TestRunRun`, or add a note: "Mirror the existing invalid-url string from TestRunRun." |
| 4 | LOW | Dry-Run | **Stale test count ("34+ existing ExtractFlags tests").** The plan references "34+ existing ExtractFlags tests" at lines 153 and 164, but `args_test.go` currently contains 43 test cases in `TestExtractFlags`. "34+" is technically true but misleading. | Update counts to approximately "43 existing ExtractFlags tests" or simply say "all existing ExtractFlags tests." Self-healing during execution, but correcting improves plan precision. |

| Severity | Count |
|----------|-------|
| BLOCKING | 0 |
| HIGH     | 1 |
| MEDIUM   | 1 |
| LOW      | 2 |
| TRIVIAL  | 0 |

### Detailed Review Findings by Dimension

#### Dry-Run Execution Review

**Verified against real codebase.** All file paths in the plan exist and are correctly identified:
- `internal/runner/args.go` (145 lines) — `validateFlagValue` at line 8, `ExtractFlags` at line 29
- `internal/runner/args_test.go` (599 lines) — `TestExtractFlags` with 43 test cases
- `cmd/run_test.go` (339 lines) — `TestRunRun` (error paths) + `TestRunRun_ModelFlags`
- `cmd/exec_test.go` (233 lines) — `TestExecRun_ModelFlags` only
- `internal/runner/runner.go` (139 lines) — `New()` validation at lines 30-48

**Function signatures verified:**
- `validateFlagValue(name, value string) error` — matches plan at line 76
- `ExtractFlags(args []string) (model, haikuModel, sonnetModel, opusModel string, remaining []string, err error)` — matches architecture spec
- `runner.New(opts Options) (*Runner, error)` — validates required fields, inputs plan references correctly

**Proposed code changes feasible:**
- `haikuModelSet` boolean addition to `ExtractFlags` — syntactically and semantically sound
- Alias resolution change from `haikuModel == ""` to `!haikuModelSet` — correctly preserves all existing behavior while fixing edge case
- Unicode/long value test cases — `ExtractFlags` uses `strings.Contains(value, "\n")` only, no encoding-dependent logic, so unicode and long strings process without crash. Tests confirming this behavior are well-scoped.

**Test pattern compliance:** Plan-prescribed testify + table-driven patterns match codebase conventions. No deviation from CLAUDE.md testing guidelines.

**Risk summary:** 1 HIGH, 1 MEDIUM, 2 LOW. No blocking issues.

**Verdict:** APPROVED WITH CONDITIONS — Plan is executable. Only the TDD ordering concern (HIGH) and args.go file contention (MEDIUM) need attention before dispatch.

---

#### Spec-Plan Alignment

**Spec files analyzed:**
1. `/home/dsdashun/src/github.com/dsdashun/ccctx/_bmad-output/planning-artifacts/prd.md`
2. `/home/dsdashun/src/github.com/dsdashun/ccctx/_bmad-output/planning-artifacts/architecture.md`
3. `/home/dsdashun/src/github.com/dsdashun/ccctx/_bmad-output/planning-artifacts/epics.md`

**Coverage matrix:**

| Spec Requirement | Plan Coverage | Status |
|---|---|---|
| Epics AC1: Shared test helpers | Plan AC1 + Task 4 (4 subtasks) | Fully covered |
| Epics AC2: Error messages include flag name | Plan AC2 + Task 1 (3 subtasks) | Fully covered |
| Epics AC3: Error-path tests for exec | Plan AC3 + Task 5 (3 subtasks) | Fully covered |
| Epics AC4: Unicode + long value tests | Plan AC4 + Task 3 (3 subtasks) | Fully covered |
| Epics AC5: `--haiku-model ""` edge case | Plan AC5 + Task 2 (4 subtasks) | Fully covered |
| Epics AC6: All tests pass | Plan AC6 + Task 6 (3 subtasks) | Fully covered |
| Architecture: File boundaries | Anti-patterns + Files Changed table | Compliant |
| Architecture: Test patterns | testify + table-driven throughout | Compliant |

**Consistency:** No contradictions between plan and any spec document. All anti-patterns in the plan derive from the Architecture document. The "Files Changed" table in the plan matches the Epics spec exactly.

**Completeness:** All 6 acceptance criteria from the Epics Story 4.3 definition map to plan tasks. No missing requirements. No scope creep — Dev Notes, anti-patterns, and references are execution infrastructure, not new features.

**Note on AC2 format detail:** The Epics spec shows error format `"--haiku-model: value cannot contain newline"` (with colon), while the plan and actual code use `"--haiku-model value cannot contain newline"` (no colon). Both satisfy the core requirement — the flag name is included. The plan correctly takes a "verify and improve only if inconsistent" approach.

**Verdict:** Full alignment. All 6 ACs covered with concrete subtasks. No spec gaps. No scope creep.

---

#### Checkpoint & Task Granularity Analysis

**Context budget estimates:**

| Component | Est. Tokens | Status |
|-----------|-------------|--------|
| All 6 tasks from single parent session | ~52K | Pass (well under 140K safety) |
| Task 1: validateFlagValue (AC2) | ~8K | Pass |
| Task 2: --haiku-model "" edge case (AC5) | ~8K | Pass |
| Task 3: Unicode + long value tests (AC4) | ~8K | Pass |
| Task 4: Shared test helpers (AC1) | ~9K | Pass |
| Task 5: Error-path tests for exec (AC3) | ~8K | Pass |
| Task 6: Verify all tests pass (AC6) | ~6K | Pass |

**No oversized tasks.** All individual tasks are 6K-9K tokens. Combined parent session is ~52K.

**File contention identified:**
- Tasks 1 and 2 both modify `internal/runner/args.go` → merge recommended
- Tasks 4 and 5 both modify `cmd/exec_test.go` → ordering enforced (Task 4 before 5, plan already respects this)

**Dependency chain:**
```
Tasks 1+2 (args.go) → Task 3 (args_test.go additions) ─┐
                                                         ├─→ Task 6 (verify)
Tasks 4 (shared helpers) → Task 5 (exec error tests) ──┘
```

The `runner/` cluster (1+2 then 3) and `cmd/` cluster (4 then 5) operate on disjoint files — independently parallelizable.

**Suggested restructuring:** Merge Tasks 1 and 2 into "Args improvements (AC2 + AC5)." Renumber tasks 3/4/5/6 → 2/3/4/5. Combined task is ~9K tokens, well within budget.

**No concern mixing detected.** Each task maps cleanly to one AC or a closely related pair.

**Verdict:** Structurally sound. One coupling optimization recommended (merge Tasks 1+2).

---

### Plan Health

| Axis | Meaning | Health | Detail |
|------|---------|--------|--------|
| **structural** | File paths, types, contracts exist — can the plan be executed? | 🟢 Green | 0 BLOCKING. All file paths verified. All function signatures match. All proposed changes syntactically feasible. |
| **procedural** | TDD ordering, verification steps present — is the process correct? | 🟡 Yellow | 1 HIGH (TDD ordering in Task 2 sub-bullets). One production code change with test written after implementation per current sub-bullet order. Trivial to fix. |
| **clerical** | Line numbers, counts consistent — are the details precise? | 🟢 Green | 2 LOW. Stale test count ("34+" vs 43) and Dev Notes example string mismatch. Both self-healing during execution. |

### Verdict

**[x] APPROVED WITH CONDITIONS** — Plan is executable after addressing one critical issue and one recommendation.

Conditions for approval:
1. **Fix TDD ordering in Task 2** (Issue #1): Reorder sub-bullets so the failing test is written before the `haikuModelSet` implementation.
2. **Consider merging Tasks 1 and 2** (Issue #2): Both modify `args.go` and Task 1 is mostly verification-only. Merging eliminates file-contention risk.

**Rationale:** Structural health is green (all file paths and signatures verified). Procedural health is yellow due to the single TDD ordering concern — trivial to fix with a one-line plan edit. Clerical health is green (minor stale counts, self-healing). The plan is faithful to all three spec documents with full AC coverage and zero architecture violations. All tasks are well-scoped and within context budget.

## Round 1 Response

**Responded by:** glm
**Date:** 2026-05-13
**Plan file:** 4-3-test-quality-and-error-message-improvement.md (updated)

### Issue Responses

#### 1. [FIX] TDD ordering violation in Task 2

**Change made:** Merged former Tasks 1 and 2 into a single "Task 1: Args improvements (AC: #2, #5)". Reordered sub-bullets so the failing test is written before any implementation: (1) verify error paths, (2) decide fix vs document, (3) write failing test FIRST, (4) implement fix or add comment, (5) verify test passes. Added bold emphasis on "Write a failing test FIRST" to prevent ambiguity. Renumbered Tasks 3/4/5/6 → 2/3/4/5.

**Location:** Tasks section, former Tasks 1–2 → new Task 1

#### 2. [FIX] Tasks 1 and 2 both target `args.go`

**Change made:** Merged into the combined Task 1 described above. Both AC2 (verify error messages) and AC5 (`--haiku-model ""` edge case) now live in a single task that reads `args.go` once and produces coordinated changes. Renumbered remaining tasks.

**Location:** Tasks section, Tasks renumbered 1–5 (was 1–6)

#### 3. [FIX] Dev Notes example inconsistent with existing code

**Change made:** Updated Dev Notes table from `base_url = "not-a-url"` to `base_url = "not-a-valid-url"` to match the canonical string in `cmd/run_test.go:95`.

**Location:** Dev Notes → "Exec Error-Path Tests" table, row "invalid base_url"

#### 4. [FIX] Stale test count ("34+ existing ExtractFlags tests")

**Change made:** Updated both occurrences of "34+" to "43" — in the Anti-Patterns section ("all 43 existing ExtractFlags tests must still pass") and in References ("43 existing test cases to preserve").

**Location:** Anti-Patterns section (line ~153) and References section (line ~166)

### Summary

All 4 issues from Round 1 have been fixed. The plan file has been updated. The plan is ready for re-review.

---

## Round 2

**Context:** Re-review after all 4 Round 1 fixes applied. Focus: verify fixes are correct, detect new issues introduced by the task merging/renumbering, and re-assess all three dimensions.

### Previous Round Status

| # | Severity | Issue | Status |
|---|----------|-------|--------|
| 1 | HIGH | TDD ordering violation in Task 2 — test listed after implementation | FIXED — Tasks 1+2 merged; sub-bullets reordered so "Write a failing test FIRST" (bold) precedes implementation |
| 2 | MEDIUM | Tasks 1 and 2 both target `args.go` — file contention risk | FIXED — Merged into single Task 1; coordinated changes to `args.go` |
| 3 | LOW | Dev Notes example `"not-a-url"` inconsistent with `"not-a-valid-url"` in run_test.go | FIXED — Table now uses `"not-a-valid-url"` (line 129) |
| 4 | LOW | Stale test count "34+" vs actual 43 ExtractFlags tests | FIXED — Both occurrences updated to "43" (lines 150, 164) |

### Consolidated Issue List

| # | Severity | Review Dimension | Issue | Suggested Fix |
|---|----------|-----------------|-------|---------------|
| 1 | MEDIUM | Granularity | **Tasks 1 and 2 both write to `args_test.go` — undocumented file-contention coupling.** Task 1 adds a `--haiku-model ""` edge case test; Task 2 adds unicode/long value tests. The dependency on sequential dispatch (Task 2 must read post-Task-1 state) is implicit in task ordering but not documented. | Either: **(A)** Add an explicit dependency note to Task 2: `**Depends on:** Task 1 (reads post-Task-1 state of args_test.go)`. Or **(B)** Merge Tasks 1 and 2 into a single "Args improvements and boundary tests (AC2, AC4, AC5)" task (~11K tokens, well within budget). Option B is cleaner and eliminates the coupling entirely. |
| 2 | LOW | Dry-Run | **Exec error-path test table (lines 124-129) is underspecified on required args.** The table shows `Test Case \| Config \| Expected` but exec tests need `-- <command>` in args to bypass `$SHELL` lookup and exercise `runner.New()` validation. Without this, a test of `execRun([]string{"badurl"})` hits `$SHELL unset` error instead of the intended validation error — a false-positive test. Dev Notes lines 131 and 156 warn about this, but the distance between table and note creates execution risk. | Add a note directly below the table: "All exec error-path tests must include `-- <command>` in args (e.g., `args: []string{"badurl", "--", "env"}`) to bypass `$SHELL` lookup and exercise `runner.New()` validation." Or add an `args` column to the table. |
| 3 | LOW | Dry-Run | **"invalid base_url" table row (line 129) omits the context section header.** The other rows specify the context name (e.g., `[context.badctx]`). This row only shows `base_url = "not-a-valid-url"` without the enclosing `[context.badurl]` section. Self-healing — the executor follows `TestRunRun` pattern which includes it. | Optionally add the context section header: `` `[context.badurl] base_url = "not-a-valid-url" auth_token = "t"` `` for consistency with other table rows. |
| 4 | TRIVIAL | Dry-Run | **Granularity reviewer noted Tasks 3 and 4 share `exec_test.go` — pre-existing coupling.** Both already ordered sequentially (Task 3 → Task 4). No new issue; informational only. | No action needed. Already handled by existing task ordering. |

| Severity | Count |
|----------|-------|
| BLOCKING | 0 |
| HIGH     | 0 |
| MEDIUM   | 1 |
| LOW      | 2 |
| TRIVIAL  | 1 |

### Detailed Review Findings by Dimension

#### Dry-Run Execution Review

**Round 1 fix verification:** All four Round 1 fixes verified correct against the plan file:
- TDD ordering: Task 1 sub-bullet 3 is "**Write a failing test FIRST**" — precedes implementation sub-bullets ✓
- Task merging: Single Task 1 covers both AC2 and AC5 ✓
- Dev Notes string: `"not-a-valid-url"` matches run_test.go:95 ✓
- Test count: Both references say "43" ✓

**Codebase verification (all passed):**
- All 7 referenced file paths exist (`internal/runner/args.go`, `internal/runner/args_test.go`, `cmd/run_test.go`, `cmd/exec_test.go`, `cmd/run.go`, `cmd/exec.go`, `internal/runner/runner.go`)
- All function signatures match: `validateFlagValue`, `ExtractFlags`, `New`, `ParseArgs`, `WantsHelp`
- 43 test cases confirmed in `TestExtractFlags`
- Baseline: `go test ./...` passes, `go vet ./...` clean

**Issues found:** 2 LOW, 1 TRIVIAL (see consolidated list #2-#3).

**New issue detail — Exec error-path test args:** The Dev Notes table on lines 124-129 describes four error-path test cases for exec. For these tests to exercise `runner.New()` validation (which validates base_url, auth_token, context existence), the args must include `-- <command>` (e.g., `["nonexistent", "--", "env"]`). Without this, calling `execRun([]string{"nonexistent"})` would:
1. Not hit `runner.New()` at all if `ParseArgs` returns early with an error, OR
2. Default to `$SHELL` which is unset in the test environment → exit code 1 for wrong reason.

The plan is aware of this (Dev Notes line 131: "Exec tests use `-- <command>` to specify the target explicitly"; line 156: "error-path tests need explicit target commands"). The issue is that the table itself doesn't show the args column, and the notes are several paragraphs away. Severity: LOW (self-healing for an attentive executor).

**Risk summary:** 0 BLOCKING, 0 HIGH, 0 MEDIUM, 2 LOW, 1 TRIVIAL. No execution blockers.

**Verdict:** Plan is executable. All paths verified against the real codebase. No structural defects.

---

#### Spec-Plan Alignment

**Round 1 re-verification:** The Round 1 fixes (task merging, renumbering) introduced zero misalignments. Full re-verification of all AC-to-task mappings:

| AC | Round 1 Task | Round 2 Task | Verified |
|----|-------------|-------------|----------|
| AC1 (shared helpers) | Task 4 | Task 3 | Yes — heading: "Extract shared test helpers for model flag tests (AC: #1)" |
| AC2 (error messages) | Task 1 | Task 1 (combined) | Yes — heading: "Args improvements ... (AC: #2, #5)" |
| AC3 (exec error tests) | Task 5 | Task 4 | Yes — heading: "Add error-path tests for exec command (AC: #3)" |
| AC4 (unicode/long) | Task 3 | Task 2 | Yes — heading: "Add unicode and long value tests to args_test.go (AC: #4)" |
| AC5 (haiku-model "") | Task 2 | Task 1 (combined) | Yes — heading: "Args improvements ... (AC: #2, #5)" |
| AC6 (all tests pass) | Task 6 | Task 5 | Yes — heading: "Verify all tests pass (AC: #6)" |

**Coverage matrix (re-verified):**
- 6/6 ACs fully covered — no gaps
- 0 plan items without spec basis — no scope creep
- All 8 anti-patterns trace to architecture boundary rules
- "Files Changed" table matches Epics spec exactly

**Architecture boundary compliance:**
- Runner package: only `args.go` modified — no `os.Exit`, no stderr, no `cmd/` imports
- Cmd package: test files only — no production code changes
- No cross-boundary violations introduced

**Issues found:** None. Full alignment maintained through all Round 1 fixes.

**Verdict:** Full alignment. All 6 ACs covered. Zero spec gaps. Zero scope creep. Zero architecture violations.

---

#### Checkpoint & Task Granularity Analysis

**Context budget re-estimate (parent session, 5 tasks):**

| Component | Est. Tokens | Status |
|-----------|-------------|--------|
| Fixed overhead | ~7K | — |
| Task 1: Args improvements (AC2+AC5) | ~10K | Pass |
| Task 2: Unicode/long value tests (AC4) | ~8K | Pass |
| Task 3: Shared test helpers (AC1) | ~11K | Pass |
| Task 4: Exec error-path tests (AC3) | ~9K | Pass |
| Task 5: Verify all tests pass (AC6) | ~5K | Pass |
| Inter-task conversation | ~5K | — |
| **Total** | **~55K** | Well under 140K safety |

**Individual task budgets:**

| Task | Files Read | Files Written | Est. Tokens | Status |
|------|-----------|---------------|-------------|--------|
| Task 1 | `args.go` (145L), `args_test.go` (599L) | `args.go` (~5 lines), `args_test.go` (~30 lines) | ~9K | Pass |
| Task 2 | `args_test.go` (599L) | `args_test.go` (~60 lines added) | ~8K | Pass |
| Task 3 | `run_test.go` (339L), `exec_test.go` (233L) | Both files refactored (~200 lines changed) | ~11K | Pass |
| Task 4 | `exec_test.go`, `run_test.go`, `exec.go`, `run.go`, `runner.go` | `exec_test.go` (~80 lines added) | ~9K | Pass |
| Task 5 | None (commands only) | None | ~5K | Pass |

All tasks well within budget. No oversized items.

**Coupling analysis:**
- **Tasks 1 & 2:** File-contention on `args_test.go` — both make additive changes. Sequential dispatch resolves it. **Dependency is implicit, not documented.** → MEDIUM issue (see consolidated list #1).
- **Tasks 3 & 4:** File-contention on `exec_test.go` — pre-existing, already ordered sequentially. → TRIVIAL (no action needed).
- **Parallel clusters:** Runner cluster (Tasks 1→2) and Cmd cluster (Tasks 3→4) operate on disjoint file sets. Can run in parallel.

**Dependency chain (verified):**
```
Task 1 (args.go + args_test.go) ──→ Task 2 (args_test.go additions) ─┐
                                                                       ├──→ Task 5 (verify all)
Task 3 (run_test.go + exec_test.go) → Task 4 (exec_test.go additions) ┘
```

**Issues found:** 1 MEDIUM (new coupling from renumbering), 1 TRIVIAL (pre-existing, informational).

**Verdict:** Structurally sound. One coupling optimization recommended — merge Tasks 1+2 or document the implicit dependency.

### Plan Health

| Axis | Meaning | Health | Detail |
|------|---------|--------|--------|
| **structural** | File paths, types, contracts exist — can the plan be executed? | 🟢 Green | 0 BLOCKING. All file paths verified. All function signatures match. All proposed changes syntactically feasible. Round 1 fixes confirmed correct. |
| **procedural** | TDD ordering, verification steps present — is the process correct? | 🟢 Green | 0 HIGH. TDD ordering corrected (Round 1 fix). Test writing precedes implementation in all tasks. Verification step (Task 5) present. |
| **clerical** | Line numbers, counts consistent — are the details precise? | 🟢 Green | 0-2 MEDIUM (1 MEDIUM coupling issue, 2 LOW table underspecification). Stale counts fixed. Line number precision adequate. |

### Verdict

**[x] APPROVED** — Plan is ready for agent-based execution.

**Rationale:** Structural health is green (0 BLOCKING — all file paths, function signatures, and proposed changes verified against real codebase). Procedural health is green (0 HIGH — TDD ordering corrected in Round 1, all tasks have test-before-implementation ordering). Clerical health is green (1 MEDIUM + 2 LOW — the MEDIUM coupling issue between Tasks 1 and 2 is resolved by sequential dispatch; adding an explicit dependency note would strengthen the plan but is not required for execution). All 4 Round 1 fixes verified correct; zero regressions. Full AC coverage maintained through renumbering.

The one actionable improvement (not required for approval): merge Tasks 1 and 2 or add `**Depends on:** Task 1` to Task 2's description. This eliminates the implicit file-contention coupling on `args_test.go`. Either approach takes ~1 minute to apply.

## Round 2 Response

**Responded by:** glm
**Date:** 2026-05-13
**Plan file:** 4-3-test-quality-and-error-message-improvement.md (updated)

### Issue Responses

#### 1. [FIX] Tasks 1 and 2 both write to `args_test.go` — undocumented file-contention coupling

**Change made:** Added explicit `**Depends on:** Task 1 (reads post-Task-1 state of args_test.go)` to Task 2's heading line. Chose option A over merging to keep tasks focused — Task 1 handles production code changes + edge case test, Task 2 handles pure additive test cases. The dependency is now explicit for any executor.

**Location:** Tasks section, Task 2 heading

#### 2. [FIX] Exec error-path test table underspecified on required args

**Change made:** Replaced the loose sentence after the table with a blockquote note directly below the table: "All exec error-path tests must include `-- <command>` in args (e.g., `args: []string{"badurl", "--", "env"}`) to bypass `$SHELL` lookup and exercise `runner.New()` validation." The note now sits adjacent to the table instead of paragraphs away.

**Location:** Dev Notes → "Exec Error-Path Tests" section, below the test case table

#### 3. [FIX] "invalid base_url" table row omits context section header

**Change made:** Updated the row from `base_url = "not-a-valid-url"` to `` `[context.badurl] base_url = "not-a-valid-url" auth_token = "t"` ``, consistent with the other rows that include the context section header.

**Location:** Dev Notes → "Exec Error-Path Tests" table, row "invalid base_url"

#### 4. [WONTFIX] Tasks 3 and 4 share `exec_test.go` — pre-existing coupling

**Reason:** As the reviewer noted, this is informational only. Sequential ordering (Task 3 → Task 4) already handles it correctly. No action needed.

### Summary

All issues from Round 2 have been addressed (3 fixed, 1 wontfix). The plan is ready for execution — no re-review needed as the verdict was already APPROVED.
