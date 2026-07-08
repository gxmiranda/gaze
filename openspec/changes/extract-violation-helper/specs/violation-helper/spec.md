## ADDED Requirements

### Requirement: Centralized new-function violation check

The `internal/crap` package MUST provide a single `isNewFunctionViolation`
function that determines whether a new function (not present in the baseline)
violates either the CRAP threshold or the GazeCRAP threshold.

The function MUST return `true` when `Score.CRAP` is strictly greater than
`crapThreshold`, OR when `Score.GazeCRAP` is non-nil and strictly greater
than `gazeCRAPThreshold`.

The function MUST return `false` when `Score.GazeCRAP` is nil and
`Score.CRAP` does not exceed the threshold.

#### Scenario: CRAP-only violation

- **GIVEN** a Score with CRAP = 35.0 and GazeCRAP = nil
- **WHEN** `isNewFunctionViolation` is called with crapThreshold = 30.0 and gazeCRAPThreshold = 30.0
- **THEN** it MUST return `true`

#### Scenario: GazeCRAP-only violation

- **GIVEN** a Score with CRAP = 10.0 and GazeCRAP = 40.0
- **WHEN** `isNewFunctionViolation` is called with crapThreshold = 30.0 and gazeCRAPThreshold = 30.0
- **THEN** it MUST return `true`

#### Scenario: No violation

- **GIVEN** a Score with CRAP = 10.0 and GazeCRAP = 5.0
- **WHEN** `isNewFunctionViolation` is called with crapThreshold = 30.0 and gazeCRAPThreshold = 30.0
- **THEN** it MUST return `false`

#### Scenario: Nil GazeCRAP with low CRAP

- **GIVEN** a Score with CRAP = 10.0 and GazeCRAP = nil
- **WHEN** `isNewFunctionViolation` is called with crapThreshold = 30.0 and gazeCRAPThreshold = 30.0
- **THEN** it MUST return `false`

#### Scenario: Both thresholds exceeded

- **GIVEN** a Score with CRAP = 35.0 and GazeCRAP = 40.0
- **WHEN** `isNewFunctionViolation` is called with crapThreshold = 30.0 and gazeCRAPThreshold = 30.0
- **THEN** it MUST return `true`

#### Scenario: Exact threshold boundary

- **GIVEN** a Score with CRAP = 30.0 and GazeCRAP = 30.0
- **WHEN** `isNewFunctionViolation` is called with crapThreshold = 30.0 and gazeCRAPThreshold = 30.0
- **THEN** it MUST return `false` (violation requires strictly greater than)

## MODIFIED Requirements

### Requirement: Violation classification consistency

All sites that classify a new function as a violation MUST use the centralized
`isNewFunctionViolation` helper instead of inline logic. Previously, three
separate sites inlined identical boolean expressions independently.

Affected sites:
- `buildComparisonSummary` (counting `NewViolations`)
- `WriteComparisonJSON` (assigning `StatusNewViolation`)
- `writeComparisonNewFunctions` (separating violations from informational)

## REMOVED Requirements

None.
