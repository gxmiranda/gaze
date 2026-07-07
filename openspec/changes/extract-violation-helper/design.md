## Context

The baseline comparison pipeline in `internal/crap/` evaluates new-function
violation status at three sites:

1. `compare.go:186-188` -- `buildComparisonSummary` counts `NewViolations`
2. `compare_report.go:119-122` -- `WriteComparisonJSON` assigns `StatusNewViolation`
3. `compare_report.go:274-276` -- `writeComparisonNewFunctions` separates
   violations from informational entries in text output

All three evaluate the same expression with the same nil guard on `*float64`.
The pattern was introduced in the original baseline-comparison change and
extended consistently in PR #169 when GazeCRAP threshold support was added.

Per the proposal's constitution alignment: this is a pure internal refactoring
with no accuracy, output, or assumption changes (Principles II, III: N/A;
Principles I, IV: PASS).

## Goals / Non-Goals

### Goals
- Extract a single helper function for the violation check
- Replace all three inline instances with calls to the helper
- Add a dedicated unit test for the helper's truth table
- Zero behavioral change

### Non-Goals
- Changing violation criteria or adding new metrics (separate concern)
- Exporting the helper (all callers are package-internal)
- Refactoring other duplication in the comparison pipeline

## Decisions

### D1: Helper location in compare.go

Place `isNewFunctionViolation` in `internal/crap/compare.go` alongside
`buildComparisonSummary`, which is the primary consumer and where the violation
concept is defined. The function is unexported since all callers are within
`internal/crap/`.

### D2: Function signature

```go
func isNewFunctionViolation(s Score, crapThreshold, gazeCRAPThreshold float64) bool
```

Accepts a `Score` value and both thresholds. Returns true if either threshold
is exceeded. The nil guard on `GazeCRAP *float64` is encapsulated inside the
helper.

This signature avoids coupling to `CompareOptions` -- the caller extracts the
threshold values. This keeps the function maximally reusable and testable with
simple scalar inputs.

### D3: Test strategy

Add `TestIsNewFunctionViolation` to `internal/crap/compare_test.go` as a
table-driven test covering:
- CRAP above threshold, no GazeCRAP -- true
- CRAP below threshold, no GazeCRAP -- false
- CRAP below threshold, GazeCRAP above threshold -- true
- CRAP below threshold, GazeCRAP below threshold -- false
- Both above threshold -- true
- GazeCRAP nil (no contract coverage data) -- only CRAP evaluated

Existing integration and unit tests for `Compare`, `WriteComparisonJSON`, and
`writeComparisonNewFunctions` provide regression coverage -- they exercise the
helper indirectly through all three call sites.

## Risks / Trade-offs

### R1: Trivial change overhead

The implementation is approximately 10 lines of production code and 40 lines
of test code. The OpenSpec overhead is proportional. Accepted because the
spec-first requirement applies to all refactoring that changes function
signatures (AGENTS.md).
