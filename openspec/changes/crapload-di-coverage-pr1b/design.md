## Context

Phase 1a (PR #183) reduced CRAPload from 38 to 32 by adding dependency injection
and tests to 6 pipeline orchestration functions. Phase 1b targets 2 remaining
`add_tests` functions that need no DI — they operate on in-memory data structures
and can be tested directly.

Both functions currently show 0% line coverage in `gaze crap` output:
- `(analyzeModel).Update` (CRAP 42) — Bubble Tea model, pure state transformation
- `isPointerArgStore` (CRAP 34) — SSA instruction classifier, pure predicate

## Goals / Non-Goals

### Goals
- Add direct unit tests for `(analyzeModel).Update` covering all message-type
  branches (WindowSizeMsg, KeyMsg quit, KeyMsg help, fallthrough)
- Add direct unit tests for `isPointerArgStore` covering all 6 positive branch
  paths (direct trace, UnOp, FieldAddr, FieldAddr+UnOp, IndexAddr,
  IndexAddr+UnOp) plus the negative fallthrough path and additional negative
  cases
- Add export shim for `isPointerArgStore` following the project's
  `export_test.go` pattern
- Add 1-2 testdata fixture functions for untested SSA instruction patterns
- Reduce CRAPload from 32 to ~30
- All tests run without `testing.Short()` guard (Constitution IV: Testability)

### Non-Goals
- No production code changes (except the export shim in `export_test.go`)
- No dependency injection refactoring (these functions don't need it)
- No coverage of `writeOneResult` (moved to Phase 2 decompose targets)
- No threshold tightening (deferred to Phase 3)
- No decomposition of high-complexity functions

## Decisions

### D1: Export shim vs. internal test for `isPointerArgStore`

**Decision**: Export shim in `export_test.go`.

**Rationale**: The project has an established pattern in
`internal/analysis/export_test.go` with 6 existing export shims (`FindFuncDecl`,
`FindMethodDecl`, `FindSSAFunction`, `BaseTypeName`, `SafeSSABuild`,
`ExprRootIdent`). Adding `IsPointerArgStore` follows this convention. Tests
remain in the external test package (`package analysis_test`) consistent with
all other mutation tests in `mutation_test.go`.

**Alternative rejected**: Internal package test (`package analysis`). Would work
but creates mixed test package style in a directory that currently uses only
external tests.

### D2: Test fixture strategy for `isPointerArgStore`

**Decision**: Add 1-2 new functions to the existing `testdata/src/mutation/mutation.go`
fixture, then load with `loadTestPackageWithSSA` to extract real `*ssa.Store`
instructions.

**Rationale**: Tests that construct real SSA from Go source code are more
accurate and maintainable than hand-building SSA instruction trees. The project
already uses this pattern extensively in `analysis_test.go`. The mutation fixture
already has `Normalize` (IndexAddr pattern), `FillSlice` (UnOp dereference), and
`ReadOnly` (no mutation). Missing: a function that writes to a field of a struct
pointer parameter (FieldAddr pattern).

**New fixture**: `SetTimeout(cfg *Config, val int) { cfg.Timeout = val }` — writes
to a field through a pointer parameter, producing an `*ssa.FieldAddr` store
address that traces to the parameter.

### D3: `(analyzeModel).Update` test approach

**Decision**: Direct construction with `newAnalyzeModel(results)` + synthetic
`tea.Msg` values.

**Rationale**: `Update` is a pure function — it takes a model and a message,
returns a new model and a command. No I/O, no external dependencies. The existing
`interactive_test.go` already has `TestMain` forcing ASCII rendering mode, and 7
tests for `renderAnalyzeContent` demonstrating the pattern. Testing `Update`
follows the same style: construct model, send message, assert on returned model
state and command.

**Key assertion targets**:
- `m.ready` flag (set by first WindowSizeMsg)
- `m.viewport.Width`/`m.viewport.Height` (set/updated by WindowSizeMsg)
- Command identity (`tea.Quit` vs nil)
- `m.help.ShowAll` toggle

### D4: No `testing.Short()` guards

**Decision**: All new tests run unconditionally (no `-short` guard).

**Rationale**: The PR 1a learning confirmed that `gaze crap` runs
`go test -short -coverprofile=<tmpfile>` internally. Tests guarded by
`testing.Short()` are skipped and contribute zero coverage. Since the purpose
of this change is to reduce CRAPload, all tests must run during coverage
generation. Both functions under test are fast (< 1ms per test) so there is
no performance reason to guard them.

## Risks / Trade-offs

### Low risk: SSA instruction patterns may vary across Go versions

The FieldAddr/IndexAddr patterns emitted by the SSA builder could theoretically
change between Go versions. Mitigation: tests assert on the observable behavior
(does `isPointerArgStore` return the correct param name?) rather than on specific
SSA instruction sequences. The project already runs CI on Go 1.24 and 1.25.

### Low risk: Export shim adds to test-only API surface

Adding `IsPointerArgStore` to `export_test.go` creates a test-only export that
could be misused. Mitigation: the file is clearly named `export_test.go` (Go
convention for test-only exports), and the project has 6 existing shims in the
same file with the same pattern.

### Accepted trade-off: Partial branch coverage for `isPointerArgStore`

Some branches (e.g., FieldAddr+UnOp) may not be reachable with a simple testdata
fixture because the Go compiler may optimize the SSA differently. If a branch
is unreachable with real Go code, the test will document this and still achieve
>80% line coverage, which is sufficient to drop CRAP below the threshold.
