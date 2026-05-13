---
reviewed_plan: _bmad-output/implementation-artifacts/4-1-config-runner-model-field-expansion.md
reviewer: plan-review skill
latest_round: 1
review_rounds:
  - round: 1
    date: 2026-05-13
health:
  structural: green
  procedural: green
  clerical: green
---

# Review: 4-1-config-runner-model-field-expansion.md

Reviewed plan: `_bmad-output/implementation-artifacts/4-1-config-runner-model-field-expansion.md`

## Round 1

**Context:** Initial full three-dimension review (dry-run execution feasibility, spec-plan alignment, checkpoint granularity) of the Story 4.1 implementation plan.

### Consolidated Issue List

| # | Severity | Review Dimension | Issue | Suggested Fix |
|---|----------|-----------------|-------|---------------|
| 1 | MEDIUM | Dry-Run | **cmd test files (`cmd/run_test.go`, `cmd/exec_test.go`) omitted from Files to Change table and lack explicit subtasks.** Both files contain hardcoded `ANTHROPIC_SMALL_FAST_MODEL` references in mock scripts and assertions (5 locations total). After `buildEnv` stops injecting this env var, these tests will fail. The plan's "Existing Test Impact" section (lines 205-206) describes the issue but no explicit subtask covers these files — only the generic Task 5 "fix any test that breaks." | Add `cmd/run_test.go` and `cmd/exec_test.go` to the "Files to Change" table. Add an explicit subtask under Task 5: "Update mock scripts in `cmd/run_test.go` and `cmd/exec_test.go` to echo `$ANTHROPIC_DEFAULT_HAIKU_MODEL` instead of `$ANTHROPIC_SMALL_FAST_MODEL`, update assertions and struct field names accordingly." |
| 2 | MEDIUM | Dry-Run | **buildEnv replacement line range is ambiguous.** Task 4 step 2 says "Rewrite `buildEnv()` model injection section in `internal/runner/runner.go:82-110`". The line range 82-110 currently spans the entire function body (including env filtering prologue and BASE_URL/AUTH_TOKEN injection). The plan's code snippet only replaces the model injection portion (equivalent to lines 93-109). A hurried executor replacing lines 82-110 with just the snippet would lose the env filtering prologue. | Change the line range reference from `82-110` to `93-109` and explicitly state: "Replace the model injection section (lines 93-109), keeping the env filtering prologue (lines 82-92) intact." Alternatively, provide the complete function body in the code snippet. |
| 3 | LOW | Dry-Run | **TestBuildEnv_PriorityChain update scope is underspecified.** Task 3 (line 64) says to update `TestBuildEnv_PriorityChain` and add new test functions for haiku/sonnet/opus, but doesn't clarify whether to keep, remove, or refactor the existing SFM test cases from `TestBuildEnv_PriorityChain`, risking duplicate coverage or gaps. | Clarify: "Remove the SFM test cases from TestBuildEnv_PriorityChain (they are now covered by TestBuildEnv_HaikuPriorityChain). Add new cases for SonnetModel and OpusModel priority. Keep existing Model priority cases unchanged." |
| 4 | TRIVIAL | Spec Alignment | **Default config template line numbers may be stale.** The plan references `config/config.go:69-76` for the default config template update. Line numbers may have shifted slightly since the plan was written. | Intent is clear regardless of exact line numbers. The executor should locate the template section by content ("Optional: specify models explicitly") rather than relying on exact line numbers. |

### Severity Summary

| Severity | Count |
|----------|-------|
| BLOCKING | 0 |
| HIGH | 0 |
| MEDIUM | 2 |
| LOW | 1 |
| TRIVIAL | 1 |

### Detailed Review Findings by Dimension

#### Dry-Run Execution Review

**Positive findings:**
- All 8 line number references in the plan match current code exactly
- All file paths exist and are writable
- TDD ordering is correct: RED (Task 1) → GREEN (Task 2) → RED (Task 3) → GREEN (Task 4) → FIX (Task 5)
- Proposed Go code changes are syntactically valid and follow existing conventions
- The plan correctly scopes Story 4.2 work out of this story — no changes to `args.go`, `cmd/run.go`, or `cmd/exec.go` source files
- `mapstructure` tags follow the existing snake_case convention
- `SmallFastModel` is retained on both Context and Options structs for backward compatibility
- Baseline is clean: `go test ./...` and `go vet ./...` pass before any changes

**Issues found:**

1. **[MEDIUM] cmd test files omitted from explicit task coverage.** `cmd/run_test.go` line 217 and `cmd/exec_test.go` line 96 use mock shell scripts that echo `$ANTHROPIC_SMALL_FAST_MODEL`. After `buildEnv` stops injecting this variable, `require.GreaterOrEqual(t, len(lines), 2, ...)` assertions at `run_test.go:233` and `exec_test.go:138` will fail because the mock output will be a blank line. The plan's Dev Notes acknowledge this but neither the "Files to Change" table nor the task list includes these files explicitly.

2. **[MEDIUM] buildEnv line range 82-110 vs. snippet scope.** The plan's code snippet (lines 75-116) shows only the model injection logic. The line range `82-110` spans the entire function body including the env filtering prologue (lines 83-91). This is resolvable with careful reading but creates a footgun.

3. **[LOW] TestBuildEnv_PriorityChain refactoring ambiguity.** The plan doesn't specify whether to keep, remove, or refactor the existing SFM test cases in `TestBuildEnv_PriorityChain` now that dedicated test functions cover haiku/sonnet/opus priority chains.

**Verdict:** Ready for execution with minor notes. No BLOCKING or HIGH issues.

#### Spec-Plan Alignment

**Coverage Summary:**
- 22 distinct spec requirements applicable to Story 4.1
- 21 fully covered
- 1 partial (haiku priority chain — see below)
- 0 missing
- 0 plan items without spec basis (no scope creep)

**Key findings:**

1. **Haiku priority chain: plan is more thorough than the architecture doc.** The architecture specifies `Options.HaikuModel > ctx.HaikuModel > ctx.SmallFastModel > omit` (3-level chain). The plan adds `Options.SmallFastModel` between `Options.HaikuModel` and `ctx.HaikuModel`, creating a 4-level chain. The plan's justification is sound: pre-Story 4.2, the `--small-fast-model` CLI flag populates `Options.SmallFastModel`, and CLI flags must take priority over config values per PRD FR28. Without this check, `ccctx run provider --small-fast-model X` would have `X` silently overridden by a config file's `haiku_model = Y`. This is not a plan error — the plan is defensively correct for the inter-story transition period.

2. **Story boundary hygiene is excellent.** The plan explicitly and repeatedly states what belongs to Story 4.2 vs. Story 4.1. No CLI flag work, no args.go changes creep in. The `Options.SmallFastModel` field is retained for backward compatibility.

3. **Test coverage is comprehensive.** Config field loading tests, haiku priority chain (5 cases), sonnet/opus priority chains, SmallFastModel-not-injected verification, empty optional field tests, and injection order verification are all specified.

4. **All 8 acceptance criteria from epics.md Story 4.1 are covered.** AC1-AC8 each map to specific tasks and subtasks in the plan.

**Consistency table:** 22 rows checked across PRD, architecture, and epics — all aligned except the haiku priority chain where the plan is more thorough (see note above).

#### Checkpoint & Task Granularity Analysis

**Context budget estimates per task:**

| Task | Est. Tokens | Status |
|------|-------------|--------|
| Task 1 (RED config tests) | ~8.3K | Pass |
| Task 2 (GREEN config struct) | ~8.2K | Pass |
| Task 3 (RED runner tests) | ~8.7K | Pass |
| Task 4 (GREEN runner buildEnv) | ~8.3K | Pass |
| Task 5 (fix remaining failures) | ~8.4K | Pass |
| **Parent session total** | **~59K** | Pass |

All tasks are well within the 140K safety threshold. None approaches even 10% of budget.

**Task coupling:** Strict linear chain — Task 1 → Task 2 → Task 3 → Task 4 → Task 5. No tasks can execute in parallel due to TDD RED-GREEN-RED-GREEN-FIX sequencing and hard compile dependencies. This is expected for a TDD-driven story.

**Task 3 sub-task volume:** 6 sub-tasks involving ~250 lines of test code changes across 4 existing test functions and 3 new test functions. Manageable as a single unit — the 247-line `runner_test.go` is the only file needing deep understanding, and all test functions follow the same table-driven pattern.

**No restructuring needed.** All tasks are appropriately sized, the TDD rhythm is correct, and dependencies are clear.

### Plan Health

| Axis | Meaning | Score | Rationale |
|------|---------|-------|-----------|
| **structural** | File paths, types, contracts exist | 🟢 Green (0 BLOCKING) | All files exist, all proposed code compiles, all dependencies resolvable |
| **procedural** | TDD ordering, verification steps present | 🟢 Green (0 HIGH) | RED-GREEN-RED-GREEN-FIX sequence is correct; verification gates at each task |
| **clerical** | Line numbers, counts consistent | 🟢 Green (2 MEDIUM) | 2 MEDIUM issues (cmd test files omission, line range ambiguity) — within green threshold |

### Verdict

**[x] APPROVED** — Plan is ready for agent-based execution

**Rationale:** Structural and procedural axes are both green. The two MEDIUM clerical issues (cmd test files not in Files to Change table, buildEnv line range ambiguity) are well within the "self-healing" range — the plan's Dev Notes already describe the cmd test impact, and a careful executor will resolve the line range by reading the actual code. Both issues are resolvable during execution without plan modification.

The plan is well-aligned with all three spec documents (PRD, architecture, epics), has no scope creep, and covers all 8 acceptance criteria. The haiku priority chain is actually more thorough than the architecture doc requires, correctly handling the inter-story transition period between Story 4.1 and Story 4.2.

**Optional improvements before execution (not required for approval):**
1. Add `cmd/run_test.go` and `cmd/exec_test.go` to Files to Change table
2. Narrow the buildEnv line range from `82-110` to `93-109`
3. Clarify whether SFM cases should be removed from `TestBuildEnv_PriorityChain`

## Round 1 Response

**Responded by:** glm
**Date:** 2026-05-13
**Plan file:** 4-1-config-runner-model-field-expansion.md (updated)

### Issue Responses

#### 1. [FIX] cmd test files omitted from Files to Change table and explicit subtasks

**Change made:** Added `cmd/run_test.go` and `cmd/exec_test.go` to the Files to Change table. Replaced the generic Task 5 subtask ("fix any test that breaks") with three explicit subtasks specifying exact locations and changes: mock script updates at `run_test.go:~217` and `exec_test.go:~96`, assertion message updates at `run_test.go:~233` and `exec_test.go:~138`, and test case name update at `exec_test.go:~62`.

**Location:** Files to Change table (between `runner_test.go` and `examples/config.toml` rows), Task 5 subtask list.

#### 2. [FIX] buildEnv replacement line range ambiguous

**Change made:** Narrowed the line range from `82-110` to `93-109` in all three locations (Task 4 step 2 description, Files to Change table, References section). Added parenthetical note "(keep the env filtering prologue at lines 82-92 intact)" to make the boundary explicit.

**Location:** Task 4 step 2, Files to Change table `runner.go` row, References section.

#### 3. [FIX] TestBuildEnv_PriorityChain update scope underspecified

**Change made:** Replaced the vague "Update TestBuildEnv_PriorityChain to include new fields in test cases" with explicit instructions: remove existing SFM test cases (lines 119-142, now covered by `TestBuildEnv_HaikuPriorityChain`), keep Model priority cases (lines 95-118), add SonnetModel priority cases (`opts.SonnetModel > ctx.SonnetModel > omit`) and OpusModel priority cases (`opts.OpusModel > ctx.OpusModel > omit`).

**Location:** Task 3, subtask for `TestBuildEnv_PriorityChain`.

#### 4. [WONTFIX] Default config template line numbers may be stale

**Reason:** Verified line numbers are currently correct — the default config template is at lines 69-76 in `config/config.go`. The reviewer acknowledged this is trivial. The plan already provides the exact content to locate ("Optional: specify models explicitly"), so an executor can find the section regardless of minor shifts.

### Summary

All issues from Round 1 have been addressed. 3 fixed, 1 wontfix (line numbers verified correct). The plan is ready for re-review.
