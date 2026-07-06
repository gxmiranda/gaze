## Why

PR 1a (`crapload-di-coverage-pr1a`, PR #183) reduced CRAPload from 38 to 32 by
adding DI and tests to 6 pipeline orchestration functions in `internal/crap/`
and `internal/aireport/`. This is Phase 1b of issue #166 — covering 2 additional
zero-coverage functions to push CRAPload further below the `--max-crapload=38`
threshold.

The two target functions have 0% line coverage and CRAP scores between 34 and 42.
Both are testable without dependency injection — they operate on in-memory data
structures (Bubble Tea messages, SSA values) rather than I/O-heavy pipelines:

| Function | Package | Complexity | CRAP | Strategy |
|----------|---------|-----------|------|----------|
| `(analyzeModel).Update` | `cmd/gaze` | 6 | 42.0 | add_tests |
| `isPointerArgStore` | `internal/analysis` | 13 | 34.1 | add_tests |

With ~80% coverage, both functions drop below CRAP 15, reducing CRAPload by 2
(from 32 to ~30) and reaching the issue's target buffer of 8 functions below
threshold.

A third function (`writeOneResult`, complexity 32, CRAP 32.3) was originally
scoped for PR 1b but moved to Phase 2 because its high complexity requires
structural decomposition before testing is practical.

## What Changes

### New Capabilities
- `IsPointerArgStore`: Export shim in `internal/analysis/export_test.go` enabling
  direct unit testing of the unexported `isPointerArgStore` function. Follows the
  existing project pattern (`SafeSSABuild`, `ExprRootIdent`, `FindFuncDecl`).

### Modified Capabilities
- `cmd/gaze/interactive_test.go`: Extended with 5-6 tests covering
  `(analyzeModel).Update` — all Bubble Tea message types (WindowSizeMsg,
  KeyMsg quit/help) and viewport initialization/resize.
- `internal/analysis/mutation_test.go`: Extended with 6-7 tests covering all
  branches of `isPointerArgStore` — direct trace, UnOp dereference, FieldAddr,
  IndexAddr, and negative cases.
- `internal/analysis/testdata/src/mutation/mutation.go`: 1-2 new fixture
  functions for FieldAddr branch coverage (struct pointer parameter with field
  write).

### Removed Capabilities
- None.

## Impact

- **Files changed**: 4 files (3 test files + 1 testdata fixture)
- **Production code**: No changes to production logic. One export shim added
  in `export_test.go` (test-only file, not compiled into the binary).
- **CRAPload**: 32 → ~30 (2-function reduction)
- **CI**: All tests run without `-short` guard, so they contribute to the
  coverage profile used by `gaze crap`.

## Constitution Alignment

Assessed against the Gaze project constitution (`.specify/memory/constitution.md`).

### I. Accuracy

**Assessment**: PASS

Tests verify that `isPointerArgStore` correctly identifies pointer argument
mutations across all SSA instruction patterns (direct, dereferenced, FieldAddr,
IndexAddr) and correctly rejects non-mutations (local variables, read-only
access). This directly supports Principle I by ensuring the mutation detection
engine has regression coverage for all known code paths.

### II. Minimal Assumptions

**Assessment**: N/A

No changes to analysis behavior or user-facing assumptions. Test fixtures use
only Go stdlib types — no external import requirements.

### III. Actionable Output

**Assessment**: N/A

No changes to output formats or reporting. This is a test-only change.

### IV. Testability

**Assessment**: PASS

This change directly advances Principle IV by adding isolated unit tests to two
functions that previously had zero coverage. The `isPointerArgStore` tests use
SSA values from testdata fixtures (no external services). The `Update` tests use
synthetic Bubble Tea messages (pure state transformation). Neither requires
`testing.Short()` guards, ensuring they contribute to CRAPload coverage profiles.
