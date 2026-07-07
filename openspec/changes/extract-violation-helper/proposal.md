## Why

The new-function violation check in the baseline comparison pipeline evaluates
the same boolean expression at three separate sites: counting violations in
`buildComparisonSummary`, assigning `StatusNewViolation` in `WriteComparisonJSON`,
and separating violations from informational entries in
`writeComparisonNewFunctions`. All three inline identical logic:

```go
crapViolation := s.CRAP > crapThreshold
gazeViolation := s.GazeCRAP != nil && *s.GazeCRAP > gazeCRAPThreshold
if crapViolation || gazeViolation { ... }
```

The duplication is a maintenance hazard: any future change to violation criteria
(e.g., adding a third metric like quality score) requires updating three
locations. If one site drifts, violation classification becomes inconsistent
between counting, JSON output, and text output. This violates CS-004 (DRY --
identical logic in more than two locations MUST be extracted).

Identified during review of PR #169 (fix: evaluate GazeCRAP in baseline
new-function threshold). Closes #179.

## What Changes

### New Helper Function

A single `isNewFunctionViolation(s Score, crapThreshold, gazeCRAPThreshold float64) bool`
function centralizes the violation check. All three call sites replace their
inline logic with a call to the helper.

### No Behavioral Change

The logic is identical -- this is a pure mechanical extraction. No scoring,
detection, or classification behavior changes. The function is unexported
(package-internal) since all three call sites are within `internal/crap/`.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- None

### Removed Capabilities
- None

## Impact

- **Modified files**: `internal/crap/compare.go` (1 call site + new helper),
  `internal/crap/compare_report.go` (2 call sites)
- **Risk**: Minimal. The current code is correct and all three sites are
  consistent. This change improves maintainability only.
- **Test impact**: Existing tests cover all three code paths. A dedicated unit
  test for the extracted helper verifies the truth table directly.

## Constitution Alignment

Assessed against the Gaze project constitution (`.specify/memory/constitution.md`).

### I. Accuracy

**Assessment**: PASS

No detection, scoring, or classification logic changes. The extracted helper
encapsulates the exact same boolean expression. Existing tests validate
identical behavior through all three call sites.

### II. Minimal Assumptions

**Assessment**: N/A

No user-facing changes. No new assumptions introduced.

### III. Actionable Output

**Assessment**: N/A

No output format or content changes.

### IV. Testability

**Assessment**: PASS

The extracted helper is directly testable with synthetic `Score` values and
scalar threshold inputs. A dedicated table-driven unit test covers all
branches of the truth table, improving test granularity over the current
state where the logic is only tested indirectly through higher-level functions.
