## Why

Three functions in `internal/quality/` have high CRAP scores due to a combination of high cyclomatic complexity and incomplete test coverage. Issue [#198](https://github.com/unbound-force/gaze/issues/198) tracks them:

| Function | File | CC | CRAP | Coverage |
|----------|------|----|------|----------|
| `WriteText` | `report.go:31-186` | 32 | 39.9 | 80.3% |
| `traceForwardDataFlow` | `mapping.go:980-1107` | 32 | 32.4 | 92.8% |
| `generateSuggestion` | `overspec.go:71-95` | 6 | 42 | 0% direct |

`WriteText` and `traceForwardDataFlow` are too complex to test thoroughly as monoliths — decomposing them into focused helpers makes each piece independently testable. `generateSuggestion` is small but has zero direct test coverage (only indirect through `ComputeOverSpecification`).

This continues the CRAPload reduction effort from issues #166 (phases 2a/2b/2c).

## What Changes

### Decomposition

1. **`WriteText`** (CC 32 → 3): Extract 11 section-rendering helpers (`writeReportHeader`, `writeContractCoverage`, `writeOverSpecification`, `writeDetectionConfidence`, `writeGapsSection`, `writeDiscardedReturns`, `writeSuggestionsSection`, `writeAmbiguousEffects`, `writeUnmappedAssertions`, `writeSSADiagnostics`, `writePackageSummary`). Define a local `qualityStyles` struct to bundle the 5 lipgloss styles.

2. **`traceForwardDataFlow`** (CC 32 → ~10): Extract 3 helpers (`rhsReferencesAnyTracked`, `handleTransformationCalls`, `extractDataFlowLHS`).

### New Tests

3. **`generateSuggestion`**: Table-driven test covering all 5 switch cases plus the default fallback.

4. **Extracted helper tests**: Style threshold boundary tests, SSA diagnostics rendering, package summary rendering, transformation call bridging, multi-iteration convergence.

## Capabilities

### New Capabilities

- None — this is a chore (internal code quality improvement).

### Modified Capabilities

- `WriteText`: Identical output behavior, reduced CC from 32 to 3 via helper extraction.
- `traceForwardDataFlow`: Identical behavior, reduced CC from 32 to ~10 via helper extraction.

### Removed Capabilities

- None.

## Impact

- **Files modified**: `internal/quality/report.go`, `internal/quality/mapping.go`
- **Files added**: `internal/quality/report_internal_test.go` (new test file for `WriteText` helpers)
- **Files modified (tests)**: `internal/quality/container_unwrap_internal_test.go` (additional tests for `generateSuggestion` and `traceForwardDataFlow` helpers)
- **No API surface changes**: All extracted helpers are unexported. `WriteText` and `traceForwardDataFlow` signatures and behavior are unchanged.
- **No behavioral changes**: Pure refactoring + test additions.

## Constitution Alignment

Assessed against the Gaze project constitution (`.specify/memory/constitution.md` v1.3.0).

### I. Accuracy

**Assessment**: PASS

Decomposition extracts helpers without changing any analysis logic. All existing tests pass unchanged, confirming output equivalence. New tests increase coverage of previously untested paths (SSA diagnostics rendering, style threshold boundaries, transformation call bridging), reducing the risk of latent inaccuracies.

### II. Minimal Assumptions

**Assessment**: N/A

No new assumptions introduced. The change is internal refactoring of report formatting and data flow tracing — no changes to how host projects are analyzed.

### III. Actionable Output

**Assessment**: PASS

Report output is byte-for-byte identical after decomposition. The extracted helpers make each output section independently testable, improving confidence that report output remains correct as the codebase evolves.

### IV. Testability

**Assessment**: PASS

This change directly improves testability — it decomposes two CC=32 functions into independently testable helpers (all CC ≤ 6) and adds targeted tests for previously uncovered paths. Coverage strategy: unit tests using synthetic data (no `testing.Short()` guards needed). The `parseAndTypeCheck` pattern from `container_unwrap_internal_test.go` is reused for AST-level tests.
