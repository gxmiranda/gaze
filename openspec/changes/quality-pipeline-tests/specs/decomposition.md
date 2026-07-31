## ADDED Requirements

### Requirement: WriteText Helper Extraction

`WriteText` in `internal/quality/report.go` MUST be decomposed into section-rendering helpers. Each helper MUST have cyclomatic complexity ≤ 6. The resulting `WriteText` function MUST have cyclomatic complexity ≤ 5. The output of `WriteText` MUST remain byte-for-byte identical after decomposition.

A local `qualityStyles` struct MUST bundle the 5 lipgloss styles (header, good, warn, bad, muted). Each helper MUST accept an `io.Writer` and the data it needs, plus the `qualityStyles` struct.

#### Scenario: WriteText output equivalence

- **GIVEN** a set of `taxonomy.QualityReport` values and a `*taxonomy.PackageSummary`
- **WHEN** `WriteText` is called before and after decomposition
- **THEN** the output bytes MUST be identical

#### Scenario: Style threshold boundary for contract coverage

- **GIVEN** a `QualityReport` with `ContractCoverage.Percentage` at boundary values (49, 50, 79, 80)
- **WHEN** `writeContractCoverage` renders the coverage
- **THEN** values < 50 MUST use the `bad` style, values >= 50 and < 80 MUST use `warn`, values >= 80 MUST use `good`

#### Scenario: Style threshold boundary for over-specification

- **GIVEN** a `QualityReport` with `OverSpecification.Count` at boundary values (0, 1, 3, 4)
- **WHEN** `writeOverSpecification` renders the count
- **THEN** value 0 MUST use the `good` style, values > 0 and <= 3 MUST use `warn`, values > 3 MUST use `bad`

#### Scenario: Style threshold boundary for detection confidence

- **GIVEN** a `QualityReport` with `AssertionDetectionConfidence` at boundary values (49, 50, 69, 70)
- **WHEN** `writeDetectionConfidence` renders the confidence
- **THEN** values < 50 MUST use the `bad` style, values >= 50 and < 70 MUST use `warn`, values >= 70 MUST use `good`

#### Scenario: SSA diagnostics rendering

- **GIVEN** a `PackageSummary` with `SSADegraded=true` and `SSADegradedPackages=["pkg/a", "pkg/b"]`
- **WHEN** `writeSSADiagnostics` renders the section
- **THEN** the output MUST contain the warning indicator, the package count, and each package name

#### Scenario: Package summary with worst coverage

- **GIVEN** a `PackageSummary` with `TotalTests > 0` and populated `WorstCoverageTests`
- **WHEN** `writePackageSummary` renders the section
- **THEN** the output MUST contain the "Lowest coverage tests" sub-section with each test name and coverage percentage

#### Scenario: Multi-report separator

- **GIVEN** two `QualityReport` values
- **WHEN** `WriteText` renders both reports
- **THEN** a blank line MUST separate the two reports

### Requirement: traceForwardDataFlow Helper Extraction

`traceForwardDataFlow` in `internal/quality/mapping.go` MUST be decomposed into concern-specific helpers. The resulting `traceForwardDataFlow` function SHOULD have cyclomatic complexity ≤ 12.

Three helpers MUST be extracted:
- `rhsReferencesAnyTracked`: MUST check whether an RHS expression references any tracked variable via direct identity and `resolveExprRoot` fallback.
- `handleTransformationCalls`: MUST detect transformation calls in an RHS expression and extract the pointer destination as a new tracked variable.
- `extractDataFlowLHS`: MUST extract the LHS variable from a non-transformation assignment, gated on `isDataExtraction`.

#### Scenario: RHS reference detection via resolveExprRoot

- **GIVEN** an assignment `x := result.Field.SubField` where `result` is tracked
- **WHEN** `rhsReferencesAnyTracked` checks the RHS
- **THEN** it MUST return true via the `resolveExprRoot` fallback path

#### Scenario: Transformation call bridging

- **GIVEN** an assignment containing `json.Unmarshal(data, &target)` where `data` is tracked
- **WHEN** `handleTransformationCalls` processes the RHS
- **THEN** it MUST return `target` as a new tracked variable

#### Scenario: Non-data-extraction gating

- **GIVEN** an assignment `got := s.Get("key")` where `s` is tracked
- **WHEN** `extractDataFlowLHS` processes the assignment
- **THEN** it MUST return nil (method call is not a data extraction)

#### Scenario: Multi-iteration convergence

- **GIVEN** a chain `a := result.Data; b := a.Items[0]; c := b.Name` where `result` is tracked
- **WHEN** `traceForwardDataFlow` processes with `maxContainerChainDepth >= 3`
- **THEN** `a`, `b`, and `c` MUST all be in the tracked set after convergence

### Requirement: generateSuggestion Direct Tests

`generateSuggestion` in `internal/quality/overspec.go` MUST have direct unit tests covering all 6 branches (5 switch cases + default).

#### Scenario: Each switch case produces a type-specific suggestion

- **GIVEN** each of the 5 effect types: `LogWrite`, `StdoutWrite`, `GoroutineSpawn`, `ContextCancellation`, `CallbackInvocation`
- **WHEN** `generateSuggestion` is called with that type and a description
- **THEN** the returned string MUST contain the description AND MUST contain type-specific guidance text

#### Scenario: Default case produces generic suggestion

- **GIVEN** an effect type not in the 5 enumerated cases (e.g., `MapMutation`)
- **WHEN** `generateSuggestion` is called
- **THEN** the returned string MUST contain both the effect type name and the description

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
