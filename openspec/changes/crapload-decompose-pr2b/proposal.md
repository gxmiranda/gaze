## Why

Three functions in `internal/quality/mapping.go` have the highest cyclomatic
complexities in the codebase: `matchContainerUnwrap` (50), `isTransformationCall`
(26), and `matchAssertionToEffect` (25). Together they represent 101 complexity
points — the top 3 in a 1575-line file with 32 functions.

`matchContainerUnwrap` at complexity 50 is the single most complex function in
the entire project. It interleaves 5-6 distinct concerns (tracked variable
collection, forward data flow tracing, transformation call detection,
data-extraction gating, assertion expression matching, and mapping construction)
in a 201-line function with nested `ast.Inspect` closures.

Direct unit test coverage is minimal: `isTransformationCall` has 5 unit tests
(with gaps for `io.Reader` and empty interface patterns), `matchAssertionToEffect`
has zero direct tests, and `matchContainerUnwrap` has 4 integration tests but
no synthetic unit tests.

This is Phase 2b of issue #166 (CRAPload fragility reduction). Phase 2a (#192)
decomposed `BuildContractCoverageFunc` and reduced CRAPload from 30 to ~29.
This phase addresses the three highest-complexity functions remaining.

## What Changes

### Decomposition

Extract concern-specific helpers from each function:

1. **`isTransformationCall` (26 → ≤8)**: Extract `isByteLikeParam` and
   `isPointerDestParam` as type-classification helpers. The main function
   becomes a simple parameter loop calling two classifiers.

2. **`matchAssertionToEffect` (25 → ≤10)**: Extract `matchDirect` (Pass 1:
   identity matching at confidence 75 + helper bridge at 70) and
   `matchIndirectRoot` (Pass 2: root resolution at confidence 65).

3. **`matchContainerUnwrap` (50 → ≤12)**: Extract `collectTrackedVars`,
   `traceForwardDataFlow`, and `matchTrackedInExpr` as data-flow tracing
   helpers.

### Test gap coverage

Add ~20 new unit tests using the existing `parseAndTypeCheck` synthetic
AST helper, plus fill 3 gaps in `isTransformationCall` coverage (`io.Reader`,
empty interface, mixed parameter ordering).

## Capabilities

### New Capabilities
- `isByteLikeParam`: Detects `[]byte`, `string`, and `io.Reader` parameter types
- `isPointerDestParam`: Detects `*T` and empty interface parameter types
- `matchDirect`: Direct `types.Object` identity matching with helper bridge
- `matchIndirectRoot`: Indirect root resolution matching via `resolveExprRoot`
- `collectTrackedVars`: Collects initial tracked variables from effect ID map
- `traceForwardDataFlow`: Multi-iteration forward data flow tracing through AST
- `matchTrackedInExpr`: Checks if any tracked variable appears in an expression

### Modified Capabilities
- `isTransformationCall`: Same behavior, reduced complexity (26 → ≤8)
- `matchAssertionToEffect`: Same behavior, reduced complexity (25 → ≤10)
- `matchContainerUnwrap`: Same behavior, reduced complexity (50 → ≤12)

### Removed Capabilities
- None

## Impact

- **Files modified**: `internal/quality/mapping.go`,
  `internal/quality/container_unwrap_internal_test.go`
- **No API surface changes**: All three functions are unexported. All extracted
  helpers are also unexported.
- **No behavior changes**: Pure decomposition. The `TestSC003_MappingAccuracy`
  ratchet (85.0% floor) must pass unchanged.
- **Total complexity reduction**: 101 → ≤30 across the three functions.

## Constitution Alignment

Assessed against the Gaze project constitution (v1.3.0).

### I. Accuracy

**Assessment**: PASS

No analysis or mapping logic changes. The `TestSC003_MappingAccuracy` ratchet
(85.0% floor across 66 assertion sites) serves as the regression gate. All
4 existing `matchContainerUnwrap` integration tests and 5 existing
`isTransformationCall` unit tests remain unchanged.

### II. Minimal Assumptions

**Assessment**: N/A

No changes to assumptions about host projects, test frameworks, or coding
styles.

### III. Actionable Output

**Assessment**: N/A

No changes to output formats, report content, or metric computation.
Confidence values (75, 70, 65, 55) are preserved identically.

### IV. Testability

**Assessment**: PASS

Directly improves testability. Seven new helpers are each independently
testable with synthetic AST data via the existing `parseAndTypeCheck` helper.
No `testing.Short()` guards needed — tests use synthetic Go source strings,
not real package loading. Coverage strategy: unit tests on each helper
covering all branches.
