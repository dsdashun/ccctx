---
reviewed_plan: _bmad-output/implementation-artifacts/4-2-cli-flag-extension-backward-compatibility.md
reviewer: Claude (plan-review skill)
latest_round: 2
review_rounds:
  - round: 1
    date: 2026-05-13
  - round: 2
    date: 2026-05-13
health:
  structural: green   # File paths, types, contracts exist
  procedural: green   # TDD ordering, verification steps present
  clerical: green     # Line numbers, counts consistent
---

# Review: 4-2-cli-flag-extension-backward-compatibility.md

Reviewed plan: `_bmad-output/implementation-artifacts/4-2-cli-flag-extension-backward-compatibility.md`

## Round 1

**Context:** Initial full three-dimension review (dry-run execution, spec-plan alignment, checkpoint/task granularity). Plan is for Story 4.2: CLI Flag Extension & Backward Compatibility — adding `--haiku-model`, `--sonnet-model`, `--opus-model` flags to ExtractFlags and converting `--small-fast-model` into an alias for `--haiku-model`.

### Consolidated Issue List

| # | Severity | Review Dimension | Issue | Suggested Fix |
|---|----------|-----------------|-------|---------------|
| 1 | HIGH | Dry-Run | **Integration tests written after implementation (TDD ordering violation).** Task 4 ("RED — add failing integration tests") is placed after Tasks 2+3 (GREEN implementation + call site updates). After Tasks 2+3, the code is fully functional — the integration tests will pass immediately on first run, never seeing a RED state. This undermines their value as a contract. | Reorder: Task 1 (unit RED) → Task 4 (integration RED) → Task 2 (unit GREEN) → Task 3 (call sites GREEN) → Task 5 (verify). Note: after Task 4 writes integration tests against the current 4-return ExtractFlags, those tests won't compile until Tasks 2+3 deliver the 6-return signature. Alternatively, keep the order but acknowledge in Task 4 that tests will be GREEN-on-arrival and add a manual verification step to cross-check expected values before implementing Tasks 2+3. |
| 2 | MEDIUM | Dry-Run | **`--small-fast-model` duplicate-flag behavior silently changes from "last-wins" to "first-wins".** The `if haikuModel == ""` guard means `--small-fast-model X --small-fast-model Y` sets haikuModel = "X" (first wins). The existing test case `"duplicate --small-fast-model last wins"` (args_test.go:99-104) expects `"bar"`. After mechanically renaming `wantSmallFast` → `wantHaikuModel`, this test will **fail**: expected `"bar"`, actual `"foo"`. The plan's Dev Notes claim "both orderings work correctly" without addressing duplicate alias flags. | Either (a) update the duplicate `--small-fast-model` test case to expect `"foo"` (first-wins), adding a comment that the guard intentionally changes this for the alias, or (b) add a separate `haikuExplicitlySet` boolean to preserve last-wins for duplicate alias flags while still giving `--haiku-model` priority. Option (a) is simpler and sufficient since using the deprecated alias twice in one command is practically nonexistent. |
| 3 | MEDIUM | Dry-Run | **Missing explicit migration steps for existing test fields (`wantSmallFast` → `wantHaikuModel`).** Task 1 says "add `wantHaikuModel`, `wantSonnetModel`, `wantOpusModel` fields" but doesn't explicitly say to *rename* the existing `wantSmallFast` field to `wantHaikuModel`. There are 4 existing test cases using `wantSmallFast` (args_test.go lines 51, 62, 102, 133). Task 4 has the same gap for `wantSFM` in integration test structs (run_test.go, exec_test.go). The corresponding assertion code (`assert.Equal(t, tt.wantSmallFast, sfm)` → for sonnet/opus, and index shifting in integration tests) also needs updating but isn't shown. | Add explicit bullets in Task 1: "Rename all existing `wantSmallFast` fields to `wantHaikuModel` in both struct definition and test case values." Add equivalent bullet in Task 4: "Rename `wantSFM` → `wantHaikuModel` in `modelFlagTest` struct and test cases. Update assertion blocks to use the new field name." |
| 4 | LOW | Dry-Run | **Integration test assertion logic for `lines[2]` (sonnet) and `lines[3]` (opus) not shown.** The plan shows the updated mock script (4 output lines) and updated test struct (5 fields), but not the corresponding assertion code in the `t.Run` loops. After the mock script outputs 4 lines, the test loops need new blocks: `assert.Equal(t, tt.wantSonnetModel, lines[2])` and `assert.Equal(t, tt.wantOpusModel, lines[3])`. The existing `lines[1]` assertion also changes from `wantSFM` to `wantHaikuModel`. | Show the updated assertion block in the plan's Dev Notes for Task 4, mirroring the level of detail already provided for the struct and mock script updates. |
| 5 | TRIVIAL | Dry-Run | **`exec_test.go` struct field `envSetup` dropped without comment.** The current `TestExecRun_ModelFlags` struct has an `envSetup func(...)` field. The plan's proposed replacement struct drops it. No existing test case uses this field, so the change is harmless. | Add a brief note: "The unused `envSetup` field from the current exec_test.go struct can be dropped." |

| Severity | Count |
|----------|-------|
| BLOCKING | 0 |
| HIGH     | 1 |
| MEDIUM   | 2 |
| LOW      | 1 |
| TRIVIAL  | 1 |

### Detailed Review Findings by Dimension

#### Dry-Run Execution Review

**File Path Verification**: All 6 files the plan references exist at their exact paths. The line numbers in the plan match the current codebase. Current code state (ExtractFlags signature, Options struct fields, buildEnv priority chains, call site code) all match the plan's assumptions.

**TDD Compliance**:

| Phase | Test-First? | Status |
|-------|-------------|--------|
| Unit tests (Task 1 → 2) | Yes | PASS — Task 1 writes failing tests before Task 2 implements |
| Integration tests (Task 4 → ?) | No | FAIL — Task 4 is placed after implementation; tests will pass immediately |
| Refactor step | N/A | No explicit REFACTOR step; minor concern but acceptable for this scope |
| Test failure verification | Partial | Task 1 implies RED state but doesn't explicitly state the verification step |

**Compilation Chain**: Tasks 1→2→3 form a compilation chain with two intermediate non-compilable states:
- After Task 1: `internal/runner` does NOT compile (tests reference 6-return signature that doesn't exist yet)
- After Task 2: `internal/runner` compiles; `cmd` does NOT compile (call sites still assign to 4 variables)

This is expected in TDD but means sub-agents for Tasks 1 and 2 cannot run `go test`/`go build` to verify their work in isolation.

**Code Logic Correctness**: Mental execution confirms all 13 acceptance criteria are correctly implemented by the plan's code snippets. The 6-value return patterns are consistent across all error and success paths. Priority logic, validation, and separator handling are all correct.

**Risk Summary**: No blocking issues. 1 HIGH, 2 MEDIUM, 1 LOW, 1 TRIVIAL.

**Verdict**: The plan is in good shape for execution. File paths, line numbers, function signatures, and code snippets are all accurate. The unit-level TDD cycle is correct. Integration test ordering is the main concern.

#### Spec-Plan Alignment

**Consistency**: All 34 spec items checked — 2 FRs, 13 ACs, 5 architecture flag mappings, 6 architecture patterns, 8 anti-patterns. All are fully covered and aligned. The plan's code snippets exactly match the Architecture spec's signatures, flag mapping table, and command file template.

**Completeness**: No missing requirements. No scope creep — all 5 plan tasks trace back to specific spec items.

**Coverage Matrix**:

| Spec Requirement | Plan Task Coverage |
|-----------------|-------------------|
| PRD FR27 (flags on both run and exec, alias) | Tasks 2, 3 |
| PRD FR28 (CLI priority over config) | Tasks 2, 3 |
| Architecture ExtractFlags signature | Task 2 |
| Architecture Flag Mapping Table (5 flags) | Task 2 |
| Architecture Command File Template | Task 3 |
| Architecture Anti-Patterns (8 items) | All respected |
| Epics AC1-AC8 (flag behavior) | Tasks 1, 2 |
| Epics AC9-AC10 (no regression) | Tasks 3, 5 |
| Epics AC11 (duplicate flags) | Task 1 |
| Epics AC12 (all tests pass) | Task 5 |
| Epics AC13 (new tests) | Tasks 1, 4 |

**Notable Observations** (not plan issues):
1. Architecture spec's haiku priority chain omits the `Options.SmallFastModel` rung that exists in actual code. After Story 4.2, this rung becomes unreachable from CLI but remains for programmatic use — correct behavior.
2. `--small-fast-model` duplicate flag behavior changes from last-wins to first-wins due to the guard. This is not defined by spec and practically never occurs. Noted for completeness (captured as MEDIUM issue #2 in dry-run findings).

**Verdict**: The plan is fully aligned with the spec. No gaps, no contradictions, no scope creep.

#### Checkpoint & Task Granularity Analysis

**Context Budget**: All 5 tasks are well within the 140K safety threshold:

| Task | Sub-agent Est. | Parent Est. | Status |
|------|---------------|-------------|--------|
| Task 1 | ~4.7K | ~16K | Pass |
| Task 2 | ~4.2K | ~16K | Pass |
| Task 3 | ~3.3K | ~15K | Pass |
| Task 4 | ~4.9K | ~17K | Pass |
| Task 5 | ~7.5K | ~20K | Pass |

All 5 tasks in a single parent session: ~65K total — well under 140K.

**Coupling**: Tasks 1→2→3→4→5 form a strict sequential chain. No parallelization possible. Tasks 1+2 and 2+3 have compilation dependencies that create two intermediate non-compilable states.

**Suggested Restructuring** (recommendation, not required):

| Original | Suggested Merge | Rationale |
|----------|----------------|-----------|
| Tasks 1+2 | Task A: RED-GREEN ExtractFlags | Single sub-agent writes tests AND implementation. Can run `go test` to verify. |
| Task 3 | Task B: Update call sites | Mechanical change, already small enough standalone. |
| Tasks 4+5 | Task C: Integration tests + verify | Single sub-agent adds tests and fixes failures. Full test suite passes at end. |

This 3-task structure eliminates compilation gaps and lets each sub-agent verify its work independently. If keeping the 5-task structure, sub-agent dispatch prompts for Tasks 1 and 2 must explicitly state that compilation failure IS the expected RED state.

**Verdict**: All tasks fit within context budget. The compilation gaps between Tasks 1→2 and 2→3 are the only structural concern, addressable with prompt instructions or the suggested merge.

### Plan Health

| Axis | Meaning | Score | Rationale |
|------|---------|-------|-----------|
| **structural** | File paths, types, contracts exist — can the plan be executed? | 🟢 green | 0 BLOCKING — all files exist, all signatures match, and all code snippets are correct against the real codebase |
| **procedural** | TDD ordering, verification steps present — is the process correct? | 🟡 yellow | 1 HIGH — integration tests written after implementation (Task 4 after Tasks 2+3); unit-level TDD is correctly ordered |
| **clerical** | Line numbers, counts consistent — are the details precise? | 🟢 green | 2 MEDIUM — missing explicit field migration notes and undocumented alias behavior change; both are executor-discoverable |

### Verdict

[ ] APPROVED — Plan is ready for agent-based execution
[x] APPROVED WITH CONDITIONS — Plan is executable after addressing critical issues
[ ] NEEDS REVISION — Plan requires significant rework before execution

**Rationale:** Structural health is green (0 BLOCKING — all files exist, all code snippets are accurate against the real codebase). Procedural health is yellow due to 1 HIGH issue: integration tests are written after implementation, which violates TDD ordering and weakens regression coverage. Clerical health is green with 2 MEDIUM issues (missing field migration notes, undocumented alias behavior change) that are executor-discoverable and won't block execution.

**Recommended actions before execution:**
1. **Reorder Task 4 before Tasks 2+3** (or acknowledge GREEN-on-arrival) — fixes the HIGH procedural issue
2. **Add explicit field migration notes** for `wantSmallFast` → `wantHaikuModel` in Tasks 1 and 4 — eliminates MEDIUM clerical issues
3. **Document the `--small-fast-model` duplicate behavior** (first-wins with guard) and update the existing test expectation — prevents execution-time debugging

## Round 1 Response

**Responded by:** Claude (plan author)
**Date:** 2026-05-13
**Plan file:** 4-2-cli-flag-extension-backward-compatibility.md (updated)

### Issue Responses

#### 1. [FIX] Integration tests written after implementation (TDD ordering violation)

**Change made:** Changed Task 4 title from "RED — add failing integration tests" to "Add integration tests for new flags". Added a subtask note acknowledging that these tests will be GREEN-on-arrival since implementation is complete from Tasks 2+3, with an explicit manual cross-check step for expected values.
**Location:** Task 4 title and first subtask bullet in the plan.

#### 2. [FIX] `--small-fast-model` duplicate-flag behavior silently changes from "last-wins" to "first-wins"

**Change made:** Replaced the `if haikuModel == ""` guard pattern with a separate `sfmAlias` variable. Each flag now independently follows last-wins semantics for duplicates within the switch loop, and priority between `--haiku-model` and `--small-fast-model` is resolved once after the loop: `if haikuModel == "" && sfmAlias != "" { haikuModel = sfmAlias }`. This satisfies both AC5 (priority) and AC11 (last-wins for duplicates). Updated the Flag Branch Logic code snippet and the surrounding explanation in Dev Notes.
**Location:** Dev Notes → Flag Branch Logic section.

#### 3. [FIX] Missing explicit migration steps for existing test fields

**Change made:** Updated Task 1 subtask to explicitly say "rename `wantSmallFast` field to `wantHaikuModel`, add `wantSonnetModel` and `wantOpusModel` fields; update all existing test case values from `wantSmallFast: ...` to `wantHaikuModel: ...`". Updated Task 4 subtask to explicitly say "rename `wantSFM` field to `wantHaikuModel`, add `wantSonnetModel` and `wantOpusModel` fields; update all existing test case values from `wantSFM: ...` to `wantHaikuModel: ...`".
**Location:** Task 1 first subtask bullet and Task 4 field migration subtask bullet.

#### 4. [FIX] Integration test assertion logic for `lines[2]` and `lines[3]` not shown

**Change made:** Added complete assertion code block to Dev Notes showing the updated `if tt.wantCode == 0` block with assertions for `wantHaikuModel` (lines[1]), `wantSonnetModel` (lines[2]), and `wantOpusModel` (lines[3]).
**Location:** Dev Notes → Mock Script Update section, after the test struct definition.

#### 5. [FIX] `exec_test.go` struct field `envSetup` dropped without comment

**Change made:** Added a note in the Mock Script Update Dev Notes section: "The unused `envSetup` field in the current `TestExecRun_ModelFlags` struct (exec_test.go) can be dropped when defining the new `modelFlagTest` struct, since no existing test case uses it."
**Location:** Dev Notes → Mock Script Update section, after the assertion block.

### Summary

All 5 issues from Round 1 have been addressed. The plan is ready for re-review.

---

## Round 2

**Context:** Re-review after all 5 Round 1 issues were addressed. Changes applied: Task 4 GREEN-on-arrival acknowledgment (Issue #1), sfmAlias variable approach for correct last-wins alias behavior (Issue #2), explicit field migration notes (Issue #3), assertion code block added to Dev Notes (Issue #4), envSetup field drop note added (Issue #5).

### Previous Round Status

| # | Severity | Issue | Status |
|---|----------|-------|--------|
| 1 | HIGH | Integration tests written after implementation (TDD ordering violation) | FIXED |
| 2 | MEDIUM | `--small-fast-model` duplicate-flag behavior silently changed from last-wins to first-wins | FIXED |
| 3 | MEDIUM | Missing explicit migration steps for existing test fields | FIXED |
| 4 | LOW | Integration test assertion logic for `lines[2]` (sonnet) and `lines[3]` (opus) not shown | FIXED |
| 5 | TRIVIAL | `exec_test.go` struct field `envSetup` dropped without comment | FIXED |

### Consolidated Issue List

| # | Severity | Review Dimension | Issue | Suggested Fix |
|---|----------|-----------------|-------|---------------|
| 1 | LOW | Dry-Run | **exec_test.go mock infrastructure adaptation requires more context than run_test.go.** exec_test.go stores its mock script in a `mockScript` Go string variable with conditional filename logic (`hasExplicitTarget` branching), not as a raw shell script like run_test.go. The plan's Dev Notes show the mock script in shell form but don't call out the exec-specific embedding pattern. | Add a note in Task 4: "exec_test.go stores mock scripts as Go string variables with conditional filename logic — apply the same 2-line→4-line expansion to the `mockScript` variable, preserving the existing `hasExplicitTarget` branching." |
| 2 | TRIVIAL | Dry-Run | **Existing `--model` branch error returns need 4→6 value update.** The plan's Flag Branch Logic code snippet (lines 112-159) shows only the new and modified branches. The existing `case "--model"` branch (`args.go:31-39`) contains two `return` statements each returning 4 values. After the signature change to 6 return values, both must be updated (e.g., `return "", "", "", "", []string{}, fmt.Errorf(...)`). | Self-healing — the compiler will flag the wrong number of return values at the exact line. Executor fixes mechanically by adding two empty strings to each return. |
| 3 | TRIVIAL | Dry-Run | **Final success return in ExtractFlags not shown with updated 6-value signature.** The plan's code snippets show the switch cases and post-loop resolution, but not the final `return model, smallFastModel, remaining, nil` line at `args.go:59`. This must be updated to `return model, haikuModel, sonnetModel, opusModel, remaining, nil`. | Self-healing via compiler. Executor fixes mechanically. |

| Severity | Count |
|----------|-------|
| BLOCKING | 0 |
| HIGH     | 0 |
| MEDIUM   | 0 |
| LOW      | 1 |
| TRIVIAL  | 2 |

No new issues of MEDIUM severity or above. All 3 findings are executor-discoverable and self-healing (compiler catches the 2 TRIVIAL issues; the 1 LOW issue is an adaptation note with the pattern already established in the plan).

### Detailed Review Findings by Dimension

#### Dry-Run Execution Review (Round 2)

**Round 1 Fix Verification:** All 5 Round 1 fixes verified against the updated plan file. Each fix is correctly applied with no regressions.

**Verification Against Real Codebase:**
- All 6 files the plan lists to modify exist at their exact paths.
- All 4 line number references match the real code exactly: `cmd/run.go:33`, `cmd/run.go:76-81`, `cmd/exec.go:32`, `cmd/exec.go:74`.
- Code state assumptions verified: ExtractFlags current 4-return signature, Options struct with HaikuModel/SonnetModel/OpusModel fields, buildEnv haiku priority chain, config layer fields, unused envSetup field in exec_test.go.
- Mock script consistency: Current mock scripts in both run_test.go and exec_test.go already output `ANTHROPIC_DEFAULT_HAIKU_MODEL` (from Story 4.1). The plan's 4-line extension is consistent.

**sfmAlias Logic Verification (mental trace across all 5 scenarios):**

| Scenario | Input | sfmAlias | haikuModel | Final Result | Correct? |
|----------|-------|----------|------------|-------------|----------|
| `--small-fast-model X` only | one alias | `"X"` | `""` | `haikuModel = "X"` (AC4) | Yes |
| `--haiku-model X` only | one explicit | `""` | `"X"` | `haikuModel = "X"` (AC1) | Yes |
| `--small-fast-model X --small-fast-model Y` | duplicate alias | `"Y"` (last wins) | `""` | `haikuModel = "Y"` (AC11) | Yes |
| `--haiku-model X --small-fast-model Y` | explicit, then alias | `"Y"` | `"X"` | `haikuModel = "X"` (AC5) | Yes |
| `--small-fast-model Y --haiku-model X` | alias, then explicit | `"Y"` | `"X"` | `haikuModel = "X"` (AC5) | Yes |

All scenarios produce correct results. The sfmAlias approach simultaneously satisfies AC4 (alias), AC5 (haiku-model priority), and AC11 (duplicate last-wins).

**TDD Compliance:**

| Phase | Test-First? | Status |
|-------|-------------|--------|
| Unit tests (Task 1 before Task 2) | Yes | PASS — Task 1 writes tests referencing 6-return signature that doesn't exist yet; LEGITIMATE RED STATE |
| Integration tests (Task 4) | No (acknowledged) | PASS with mitigation — explicit GREEN-on-arrival note + manual cross-check step |
| Test failure verification | Yes | PASS — Task 1 implies RED, Task 5 explicitly runs `go test ./...` |

**Compilation Chain:** Two intermediate non-compilable states (after Task 1, after Task 2) — expected in TDD with signature changes.

**Risk Summary:** 0 BLOCKING, 0 HIGH, 0 MEDIUM, 1 LOW, 2 TRIVIAL. All TRIVIAL issues are compiler-self-healing.

**Verdict:** Plan is ready for execution. No blocking, high, or medium-severity issues remain.

---

#### Spec-Plan Alignment (Round 2)

**Round 1 Fix Verification:** All 5 Round 1 fixes confirmed resolved in the plan. No regressions.

**Focus Area 1 — sfmAlias and FR28 (CLI Priority Over Config):**

The sfmAlias variable correctly routes `--small-fast-model` through `opts.HaikuModel`, which is the highest-priority rung in buildEnv's haiku priority chain. All 4 model variables satisfy CLI-over-config precedence:

| CLI Flag | ExtractFlow | Options Field | Priority vs Config |
|----------|-------------|---------------|-------------------|
| `--haiku-model X` | `haikuModel = X` directly | `opts.HaikuModel` | CLI wins |
| `--small-fast-model Y` | `sfmAlias = Y` then `haikuModel = Y` | `opts.HaikuModel` | CLI wins |
| `--sonnet-model X` | `sonnetModel = X` directly | `opts.SonnetModel` | CLI wins |
| `--opus-model X` | `opusModel = X` directly | `opts.OpusModel` | CLI wins |

When no CLI flag is provided, all return values are empty strings and buildEnv falls through to config-level values. FR28 is fully satisfied.

**Focus Area 2 — All 13 ACs Coverage:**

| AC | Description | Plan Coverage | Status |
|----|-------------|--------------|--------|
| AC1-AC3 | New flag basic behavior | Task 2 switch cases + Task 4 integration tests | Covered |
| AC4 | `--small-fast-model` alias | Task 2 sfmAlias → haikuModel resolution | Covered |
| AC5 | `--haiku-model` wins over alias | `if haikuModel == ""` guard | Covered |
| AC6-AC7 | Value requirement + newline validation | `i+1 >= len(preSep)` checks + `validateFlagValue` calls | Covered |
| AC8 | Flags only before `--` | Operates on `preSep` slice | Covered |
| AC9-AC10 | No regression | Anti-patterns + Task 5 verification | Covered |
| AC11 | Duplicate last-wins | Each branch overwrites its target variable | Covered |
| AC12-AC13 | Tests pass + new tests | Tasks 1, 4, 5 | Covered |

All 13 ACs fully covered. No gaps.

**Focus Area 3 — Architecture Flag Mapping Consistency:**

All 5 flags in the architecture spec's mapping table are correctly implemented in the plan's code snippets. The ExtractFlags signature matches exactly. The Command File Template patterns match both `cmd/run.go` and `cmd/exec.go` changes.

**Minor observation (not a plan issue):** The architecture spec's haiku priority chain omits `Options.SmallFastModel` as a rung, but the actual code from Story 4.1 includes it. After Story 4.2, `opts.SmallFastModel` is never set by any CLI flag, making the rung unreachable from CLI regardless. This is a pre-existing documentation gap from Story 4.1, not a Story 4.2 issue.

**Focus Area 4 — Scope Creep Check:**

The `sfmAlias` variable is a local, unexported implementation detail within ExtractFlags. It exists solely to correctly implement the interaction of AC4, AC5, and AC11 simultaneously. No new exported types, fields, or APIs. No scope creep.

**Coverage Summary:** 29/29 spec requirements fully covered. 0 gaps. 0 scope creep items.

**Verdict:** Plan is fully aligned with all three spec documents.

---

#### Checkpoint & Task Granularity Analysis (Round 2)

**Context Budget — Parent Session:** ~51K total (well under 140K safety threshold)

**Context Budget — Per Task:**

| Task | Description | Est. Tokens | Budget Status |
|------|-------------|-------------|---------------|
| Task 1 | RED tests in args_test.go | ~8.5K | Pass |
| Task 2 | GREEN ExtractFlags + sfmAlias | ~8.5K | Pass |
| Task 3 | run.go + exec.go call sites | ~7.3K | Pass |
| Task 4 | Integration tests | ~8.7K | Pass |
| Task 5 | Verify + fix | ~9.5K | Pass |

All 5 tasks are well within the 140K safety threshold. Largest task (Task 5) is only ~7% of the limit.

**sfmAlias Impact on Task Sizing:** The sfmAlias approach adds ~5 lines of code compared to the simpler guard approach. This translates to negligible token impact (~8 tokens of write cost). Task 2 estimate unchanged at ~8.5K.

**Assertion Code Block Impact:** The added assertion code block in Dev Notes is documentation that the sub-agent would need regardless. It reduces reasoning burden rather than increasing write burden. Task 4 estimate unchanged.

**Compilation Gaps:** Same two gaps as Round 1 — Gap 1 (after Task 1: tests expect 6-return but impl is 4-return) and Gap 2 (after Task 2: impl is 6-return but call sites destructure 4). Both are expected TDD intermediate states.

**Coupling:** All 5 tasks form a sequential chain (1→2→3→4→5). No parallelization possible. Circular dependencies: none.

**Optional Restructuring (unchanged from Round 1):** Merge Tasks 1+2 (RED-GREEN atomic) and Tasks 4+5 (integrate+verify) to eliminate compilation gaps. All merged tasks remain well under budget. Not required — the 5-task structure is also fine if sub-agent prompts note that compilation failure is the expected RED state.

**Verdict:** All tasks within budget. No granularity issues.

---

### Plan Health

| Axis | Meaning | Score | Rationale |
|------|---------|-------|-----------|
| **structural** | File paths, types, contracts exist — can the plan be executed? | 🟢 green | 0 BLOCKING — all files exist, all signatures match real codebase, all code snippets correct |
| **procedural** | TDD ordering, verification steps present — is the process correct? | 🟢 green | 0 HIGH — Round 1 HIGH issue resolved (Task 4 GREEN-on-arrival now documented with mitigation); unit-level TDD correctly ordered |
| **clerical** | Line numbers, counts consistent — are the details precise? | 🟢 green | 1 LOW + 2 TRIVIAL — all are executor-discoverable and compiler-self-healing; no stale line numbers or erroneous counts |

### Verdict

[x] APPROVED — Plan is ready for agent-based execution

**Rationale:** Structural health is green (0 BLOCKING). Procedural health is now green (Round 1's HIGH TDD-ordering issue resolved — Task 4 GREEN-on-arrival documented with manual cross-check mitigation). Clerical health is green (3 minor executor-discoverable findings, all self-healing via compiler or established patterns). All 5 Round 1 fixes verified correct against the real codebase. The sfmAlias approach correctly implements AC5 (haiku-model priority) and AC11 (duplicate last-wins) simultaneously. Full spec alignment: 29/29 requirements covered, 0 gaps, 0 scope creep. All tasks well within context budget. The plan is ready for agent-based execution.

## Round 2 Response

**Responded by:** Claude (plan author)
**Date:** 2026-05-13
**Plan file:** 4-2-cli-flag-extension-backward-compatibility.md (updated)

### Issue Responses

#### 1. [FIX] exec_test.go mock infrastructure adaptation requires more context than run_test.go

**Change made:** Added a note in Dev Notes → Mock Script Update section (after the `envSetup` drop note) explaining that exec_test.go stores its mock script as a Go `[]byte` variable with conditional filename logic via `hasExplicitTarget` branching, and that the executor should apply the 2-line→4-line expansion to the `mockScript` variable body while preserving the existing branching pattern. Explicitly contrasts this with run_test.go's hardcoded `claudePath` approach.
**Location:** Dev Notes → Mock Script Update section, after the `envSetup` note.

#### 2. [WONTFIX] Existing `--model` branch error returns need 4→6 value update

**Reason:** Self-healing via compiler. The two `return` statements in `case "--model"` (args.go:33 and args.go:36) will produce compile errors with wrong return count. The executor fixes mechanically by adding two empty strings. Adding an explicit note would be redundant — every compile error the executor hits during GREEN phase (Task 2) is immediately obvious and fixable.

#### 3. [WONTFIX] Final success return in ExtractFlags not shown with updated 6-value signature

**Reason:** Same as Issue 2 — self-healing via compiler. The `return model, smallFastModel, remaining, nil` at args.go:59 will fail to compile with wrong return count. Mechanical fix: add the three new variables. This is the most obvious line to update in the entire plan.

### Summary

- Fixed: 1 issue (exec_test.go mock script embedding pattern note added)
- Wontfix: 2 issues (both TRIVIAL, compiler-self-healing)
- Plan file: updated

The plan verdict is already APPROVED. All remaining findings are executor-discoverable. No re-review needed.
