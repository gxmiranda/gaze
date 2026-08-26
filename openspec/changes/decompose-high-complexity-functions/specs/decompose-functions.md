## ADDED Requirements

### Requirement: writeEffectRows helper

`writeEffectRows` MUST accept a slice of `taxonomy.SideEffect`, a maximum
description length, and a `showClassify` boolean. It MUST return a
`[][]string` of table rows with description truncation applied. When
`showClassify` is true, each row SHALL have 4 columns (TIER, TYPE,
DESCRIPTION, CLASSIFICATION). When false, each row SHALL have 3 columns
(TIER, TYPE, DESCRIPTION).

#### Scenario: Classify mode with long description

- **GIVEN** a side effect with a 40-character description and `showClassify=true`
  with `maxDesc=26`
- **WHEN** `writeEffectRows` is called
- **THEN** the description column SHALL be truncated to 23 characters plus "..."
  AND the row SHALL have 4 columns including a classification cell

#### Scenario: Non-classify mode with short description

- **GIVEN** a side effect with a 10-character description and `showClassify=false`
  with `maxDesc=42`
- **WHEN** `writeEffectRows` is called
- **THEN** the description column SHALL be the full original text
  AND the row SHALL have 3 columns

#### Scenario: Classification cell formatting

- **GIVEN** a side effect with `Classification.Label="contractual"` and
  `Classification.Confidence=85` and `showClassify=true`
- **WHEN** `writeEffectRows` is called
- **THEN** the classification cell SHALL be `"contractual/85%"`

#### Scenario: Nil classification

- **GIVEN** a side effect with `Classification=nil` and `showClassify=true`
- **WHEN** `writeEffectRows` is called
- **THEN** the classification cell SHALL be `"—"`

### Requirement: writeVerboseSignals helper

`writeVerboseSignals` MUST accept an `io.Writer` and a slice of
`taxonomy.SideEffect`. It MUST iterate over side effects that have
non-nil Classification with non-empty Signals, printing the signal
breakdown for each. It MUST be a no-op when no side effects have
signals.

#### Scenario: Side effect with signals

- **GIVEN** a side effect with 2 signals, each with Source, Weight, and
  optional Reasoning/SourceFile/Excerpt
- **WHEN** `writeVerboseSignals` is called
- **THEN** it SHALL write a "Signals for TYPE (location):" header
  followed by one line per signal with weight and optional reasoning

#### Scenario: No signals

- **GIVEN** side effects with nil Classification or empty Signals
- **WHEN** `writeVerboseSignals` is called
- **THEN** it SHALL produce no output

### Requirement: runQualityPerPackage helper

`runQualityPerPackage` MUST encapsulate the per-package analysis loop
body from `runQuality`: analyze, classify, load test package, and
assess quality. It MUST return the quality reports, package summary,
and any error.

#### Scenario: Successful package analysis

- **GIVEN** a valid package path with test files
- **WHEN** `runQualityPerPackage` is called
- **THEN** it SHALL return non-empty reports and a non-nil summary

#### Scenario: Package without tests

- **GIVEN** a package path where `loadTestPackage` fails
- **WHEN** `runQualityPerPackage` is called
- **THEN** it SHALL return nil reports, nil summary, and nil error
  (graceful skip)

#### Scenario: Classification failure

- **GIVEN** a package path where classification fails
- **WHEN** `runQualityPerPackage` is called
- **THEN** it SHALL return a non-nil error wrapping the classification error

### Requirement: writeQualityEmptyResults helper

`writeQualityEmptyResults` MUST handle the empty-results output path
from `runQuality`: write JSON or text output showing zero mapped tests,
skipped test details (truncated at `MaxSkippedTestDisplay`), and the
`--target` hint. It MUST NOT evaluate thresholds (that is the caller's
responsibility).

#### Scenario: Text format with skipped tests

- **GIVEN** format "text", merged summary with 5 skipped tests
- **WHEN** `writeQualityEmptyResults` is called
- **THEN** it SHALL write the "0 of N test functions mapped" header,
  list the skipped test names, and include the target hint

#### Scenario: JSON format

- **GIVEN** format "json", merged summary
- **WHEN** `writeQualityEmptyResults` is called
- **THEN** it SHALL produce valid JSON with an empty reports array

### Requirement: handleReceiverAssignStmt helper

`handleReceiverAssignStmt` MUST accept an `*ast.AssignStmt` and a
receiver name string. It MUST return `true` and the statement position
if any LHS expression has a root identifier matching the receiver name
AND the LHS is not a bare identifier (i.e., it must be a field access).

#### Scenario: Field assignment

- **GIVEN** an assignment `recv.Field = value` with `receiverName="recv"`
- **WHEN** `handleReceiverAssignStmt` is called
- **THEN** it SHALL return `(true, node.Pos())`

#### Scenario: Bare receiver assignment

- **GIVEN** an assignment `recv = value` with `receiverName="recv"`
- **WHEN** `handleReceiverAssignStmt` is called
- **THEN** it SHALL return `(false, 0)` — bare receiver assignment is
  not a mutation

#### Scenario: Unrelated assignment

- **GIVEN** an assignment `x.Field = value` with `receiverName="recv"`
- **WHEN** `handleReceiverAssignStmt` is called
- **THEN** it SHALL return `(false, 0)`

### Requirement: handleReceiverIncDecStmt helper

`handleReceiverIncDecStmt` MUST accept an `*ast.IncDecStmt` and a
receiver name string. It MUST return `true` and the statement position
if the operand has a root identifier matching the receiver name AND
the operand is not a bare identifier.

#### Scenario: Field increment

- **GIVEN** a statement `recv.count++` with `receiverName="recv"`
- **WHEN** `handleReceiverIncDecStmt` is called
- **THEN** it SHALL return `(true, node.Pos())`

#### Scenario: Bare receiver increment

- **GIVEN** a statement `recv++` with `receiverName="recv"`
- **WHEN** `handleReceiverIncDecStmt` is called
- **THEN** it SHALL return `(false, 0)`

### Requirement: handleReceiverCallExpr helper

`handleReceiverCallExpr` MUST accept an `*ast.CallExpr` and a receiver
name string. It MUST return `true` and the call position if the call
is a method call on a receiver field (e.g., `recv.field.Method()`)
but NOT a direct method call on the receiver itself
(e.g., `recv.Method()`).

#### Scenario: Method on receiver field

- **GIVEN** a call `recv.index.Delete(key)` with `receiverName="recv"`
- **WHEN** `handleReceiverCallExpr` is called
- **THEN** it SHALL return `(true, node.Pos())`

#### Scenario: Direct method on receiver

- **GIVEN** a call `recv.Method()` with `receiverName="recv"`
- **WHEN** `handleReceiverCallExpr` is called
- **THEN** it SHALL return `(false, 0)`

### Requirement: resolveBaselineAndCompare helper

`resolveBaselineAndCompare` MUST encapsulate the baseline resolution
and comparison logic from `runCrap`: resolve the baseline path, load
and compare if present, return the comparison result or error. This
consolidates lines 562-571 of the current `runCrap`.

#### Scenario: No baseline configured

- **GIVEN** an empty baseline path and no `.gaze/baseline.json` in the
  module directory
- **WHEN** `resolveBaselineAndCompare` is called
- **THEN** it SHALL return nil comparison result and nil error

#### Scenario: Baseline present with regressions

- **GIVEN** a baseline path pointing to a valid baseline file with
  regressions
- **WHEN** `resolveBaselineAndCompare` is called
- **THEN** it SHALL return a non-nil ComparisonResult with
  `Summary.Passed=false`

### Requirement: writeCrapOutputAndSummary helper

`writeCrapOutputAndSummary` MUST encapsulate the output writing phase
from `runCrap`: write the comparison report or normal report based on
whether a comparison result exists, then print the CI summary to
stderr. This consolidates lines 573-584 of the current `runCrap`.

#### Scenario: With comparison result

- **GIVEN** a non-nil comparison result
- **WHEN** `writeCrapOutputAndSummary` is called
- **THEN** it SHALL call `writeCrapComparisonReport`

#### Scenario: Without comparison result

- **GIVEN** a nil comparison result
- **WHEN** `writeCrapOutputAndSummary` is called
- **THEN** it SHALL call `writeCrapReport`

### Requirement: evaluateCrapGates helper

`evaluateCrapGates` MUST encapsulate the two sequential gate checks
from `runCrap`: baseline comparison gate (D7) followed by threshold
gate. The baseline gate MUST be evaluated first so its output is
always visible before a threshold failure.

#### Scenario: Baseline regression

- **GIVEN** a comparison result with `Summary.Passed=false`
- **WHEN** `evaluateCrapGates` is called
- **THEN** it SHALL return a non-nil error describing the regression
  WITHOUT evaluating thresholds

#### Scenario: Threshold violation

- **GIVEN** a nil comparison result (or `Summary.Passed=true`) and a
  threshold violation
- **WHEN** `evaluateCrapGates` is called
- **THEN** it SHALL return the threshold error

#### Scenario: All gates pass

- **GIVEN** no baseline regression and no threshold violation
- **WHEN** `evaluateCrapGates` is called
- **THEN** it SHALL return nil

## MODIFIED Requirements

### Requirement: writeOneResult complexity

`writeOneResult` MUST delegate row construction to `writeEffectRows`
and verbose signal output to `writeVerboseSignals`. The function's
cyclomatic complexity MUST be 15 or less.

Previously: Row construction and signal output were inline with
complexity 32.

### Requirement: runQuality complexity

`runQuality` MUST delegate the per-package loop body to
`runQualityPerPackage` and the empty-results output to
`writeQualityEmptyResults`. The function's cyclomatic complexity
MUST be 15 or less.

Previously: All concerns were interleaved with complexity 32.

### Requirement: detectASTReceiverMutations complexity

`detectASTReceiverMutations` MUST delegate per-AST-node-type handling
to `handleReceiverAssignStmt`, `handleReceiverIncDecStmt`, and
`handleReceiverCallExpr`. The function's cyclomatic complexity MUST
be 15 or less.

Previously: All three node type handlers were inline with complexity 24.

### Requirement: runCrap complexity

`runCrap` MUST delegate baseline handling to `resolveBaselineAndCompare`,
output writing to `writeCrapOutputAndSummary`, and gate evaluation to
`evaluateCrapGates`. The function's cyclomatic complexity MUST be 15
or less.

Previously: All concerns were inline with complexity 19.

### Requirement: isPointerArgStore complexity

`isPointerArgStore` MUST remove the structurally unreachable branches
after the initial `tracesToParam(addr, param)` check. The `UnOp`,
`FieldAddr`, and `IndexAddr` unwrapping within the loop body duplicates
logic that `tracesToParam` already performs recursively via
`tracesToParamVisited`. The function's cyclomatic complexity MUST be
5 or less after branch removal.

Previously: Redundant branches produced complexity 13 with 0% coverage
on the unreachable paths.

## REMOVED Requirements

None — pure decomposition, no functionality removed.
