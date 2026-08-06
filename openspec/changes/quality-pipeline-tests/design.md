## Context

Three functions in `internal/quality/` have elevated CRAP scores (issue #198):

- `WriteText` (CC=32, CRAP=39.9) — monolithic report renderer with 11 distinct output sections, each with its own conditional logic and style selection.
- `traceForwardDataFlow` (CC=32, CRAP=32.4) — deeply nested AST walker with three distinct concerns (RHS reference checking, transformation call handling, LHS extraction) interleaved in a single function.
- `generateSuggestion` (CC=6, 0% direct coverage) — small switch statement with no direct tests.

The crap package already established the decomposition pattern in `crap/compare_report.go` (extracted `writeScoreTable`, `writeSummarySection`, `writeQuadrantSection`, etc.) and in `crap/report.go` (extracted 6 section helpers). The quality package's `mapping.go` was partially decomposed in Phase 2b (#166) — `matchContainerUnwrap` dropped from CC=50 to CC=8, but `traceForwardDataFlow` absorbed the complexity at CC=32.

This design follows the same helper-extraction pattern, adapting it to the quality package's simpler style system.

## Goals / Non-Goals

### Goals

- Reduce `WriteText` CC from 32 to ≤ 5
- Reduce `traceForwardDataFlow` CC from 32 to ≤ 12
- Add direct unit tests for `generateSuggestion` (all 6 branches)
- Add targeted tests for extracted helpers covering previously untested paths
- All new tests run without `testing.Short()` guards (synthetic data only)

### Non-Goals

- Reusing the shared `report.Styles` struct from `internal/report/` — the quality report uses a simpler 5-style palette (header/good/warn/bad/muted) vs the crap report's 20+ field struct. Adding quality-specific fields to the shared struct would bloat it for all consumers.
- Changing `WriteText` output format — byte-for-byte identical output after decomposition.
- Decomposing `WriteJSON` — it is CC=1 (delegates to `json.NewEncoder`).
- Further decomposing helpers extracted from `traceForwardDataFlow` — the extracted helpers will have CC ≤ 12 each, which is acceptable.

## Decisions

### D1: Local `qualityStyles` struct

Define a package-local `qualityStyles` struct in `report.go` bundling the 5 lipgloss styles currently created as local variables in `WriteText`:

```go
type qualityStyles struct {
    header lipgloss.Style
    good   lipgloss.Style
    warn   lipgloss.Style
    bad    lipgloss.Style
    muted  lipgloss.Style
}
```

**Rationale**: The `crap/report.go` pattern uses `report.Styles` (shared), but the quality report's palette is fundamentally different (good/warn/bad threshold coloring vs tier-based coloring). A local struct keeps the styles co-located with their consumers and avoids coupling to the crap report's style system.

### D2: 11 section helpers for `WriteText`

Each helper follows the signature pattern `func writeXxx(w io.Writer, ..., s qualityStyles)` and handles one logical output section:

| Helper | Lines | CC | Responsibility |
|--------|-------|----|----------------|
| `writeReportHeader` | 44-51 | 1 | `=== Test -> Target ===` header |
| `writeContractCoverage` | 53-64 | 3 | Coverage percentage with threshold coloring |
| `writeOverSpecification` | 66-76 | 3 | Over-spec count with threshold coloring |
| `writeDetectionConfidence` | 78-88 | 3 | Confidence percentage with threshold coloring |
| `writeGapsSection` | 90-101 | 5 | Gap list with hints |
| `writeDiscardedReturns` | 103-113 | 5 | Discarded returns with hints |
| `writeSuggestionsSection` | 115-121 | 3 | Over-spec suggestions |
| `writeAmbiguousEffects` | 123-131 | 3 | Ambiguous effects list |
| `writeUnmappedAssertions` | 133-146 | 4 | Unmapped assertions with reasons |
| `writeSSADiagnostics` | 149-159 | 6 | SSA degradation warnings |
| `writePackageSummary` | 161-183 | 6 | Package-level summary footer |

After extraction, `WriteText` becomes a ~20-line orchestrator (CC=3): style initialization, report loop with separator, then two trailing section calls.

### D3: 3 helpers for `traceForwardDataFlow`

Extract three concern-specific helpers:

1. **`rhsReferencesAnyTracked`** — checks whether an RHS expression references any tracked variable (direct `containsObject` check + `resolveExprRoot` fallback). Absorbs lines 998-1023.
2. **`handleTransformationCalls`** — handles the transformation call detection and pointer destination extraction. Absorbs the inner `ast.Inspect` closure at lines 1029-1063.
3. **`extractDataFlowLHS`** — handles the non-transformation LHS extraction with `isDataExtraction` gating. Absorbs lines 1069-1090.

After extraction, `traceForwardDataFlow` becomes an iteration loop that calls these three helpers sequentially per RHS element, with convergence checking. Target CC: ~10.

### D4: Test file placement

| Tests for | File | Package |
|-----------|------|---------|
| `generateSuggestion` | `container_unwrap_internal_test.go` | `quality` (internal) |
| `traceForwardDataFlow` helpers | `container_unwrap_internal_test.go` | `quality` (internal) |
| `WriteText` section helpers | `report_internal_test.go` (new) | `quality` (internal) |

**Rationale**: The extracted helpers are all unexported, requiring internal package tests. `container_unwrap_internal_test.go` already contains tests for mapping helpers and has the `parseAndTypeCheck` infrastructure. Report helpers need a separate file because they test formatting output (different concern, different test patterns).

### D5: Test patterns

- **`generateSuggestion`**: Table-driven with 6 cases (5 switch arms + default), asserting both `strings.Contains` for key phrases and `strings.HasPrefix` for format consistency.
- **Report helpers**: Construct synthetic `taxonomy.QualityReport` structs, call each helper, assert output contains expected strings. Focus on threshold boundary values (49/50/79/80 for coverage, 0/1/4 for over-spec, 49/50/69/70 for confidence).
- **Data flow helpers**: Use `parseAndTypeCheck` to create synthetic AST from Go source strings, following the established pattern in `container_unwrap_internal_test.go`.

## Risks / Trade-offs

### R1: Increased function count

Adding 14 new functions (11 + 3 helpers) increases the function count in `report.go` and `mapping.go`. This is an accepted trade-off — each function is small, focused, and independently testable. The `crap/report.go` decomposition (6 helpers) and `crap/compare_report.go` decomposition (6 helpers) established this pattern as project convention.

### R2: `traceForwardDataFlow` may not reach CC ≤ 12

The Dewey learning from Phase 2b notes: "traceForwardDataFlow landed at complexity 32 despite the design estimating ~12." The outer iteration loop, AST inspection, and convergence check contribute inherent complexity. After extracting 3 helpers, the residual CC depends on how much branching remains in the orchestrator. If CC lands at ~10-12, that is acceptable. If it lands higher, we accept the result — the value is in making each concern independently testable, not in hitting an exact CC target.

### R3: Style threshold tests are fragile to format changes

Tests that assert on specific output strings (e.g., checking that coverage renders with a specific color code) will break if the output format changes. This is acceptable — the tests catch unintended format regressions, and intentional changes should update the tests.
