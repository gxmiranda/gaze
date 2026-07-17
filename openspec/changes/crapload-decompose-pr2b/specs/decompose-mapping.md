## ADDED Requirements

### Requirement: isByteLikeParam helper

`isByteLikeParam` MUST accept a `types.Type` and return `true` if the type
is `[]byte`, `string`, or implements `io.Reader` (has a method named `Read`
in its method set).

#### Scenario: []byte parameter

- **GIVEN** a `types.Type` representing `[]byte` (Slice with byte element)
- **WHEN** `isByteLikeParam` is called
- **THEN** it SHALL return `true`

#### Scenario: string parameter

- **GIVEN** a `types.Type` representing `string` (Basic with String kind)
- **WHEN** `isByteLikeParam` is called
- **THEN** it SHALL return `true`

#### Scenario: io.Reader parameter

- **GIVEN** a `types.Type` representing an interface with a `Read` method
- **WHEN** `isByteLikeParam` is called
- **THEN** it SHALL return `true`

#### Scenario: non-byte-like parameter

- **GIVEN** a `types.Type` representing `int`
- **WHEN** `isByteLikeParam` is called
- **THEN** it SHALL return `false`

### Requirement: isPointerDestParam helper

`isPointerDestParam` MUST accept a `types.Type` and return `true` if the type
is a pointer (`*T`) or an empty interface (`interface{}` / `any`).

#### Scenario: pointer parameter

- **GIVEN** a `types.Type` representing `*int`
- **WHEN** `isPointerDestParam` is called
- **THEN** it SHALL return `true`

#### Scenario: empty interface parameter

- **GIVEN** a `types.Type` representing `interface{}` (empty method set)
- **WHEN** `isPointerDestParam` is called
- **THEN** it SHALL return `true`

#### Scenario: non-pointer non-interface parameter

- **GIVEN** a `types.Type` representing `int`
- **WHEN** `isPointerDestParam` is called
- **THEN** it SHALL return `false`

### Requirement: matchDirect helper

`matchDirect` MUST walk the assertion expression tree via `ast.Inspect`,
look up each `*ast.Ident` in `types.Info.Uses` (falling back to `.Defs`),
and check for a match in `objToEffectID` (confidence 75) or `helperBridge`
→ `objToEffectID` (confidence 70). It MUST return the first match found,
or `nil` if no match exists.

#### Scenario: Direct identity match

- **GIVEN** an assertion expression containing an identifier whose
  `types.Object` is a key in `objToEffectID`
- **WHEN** `matchDirect` is called
- **THEN** it SHALL return an `AssertionMapping` with confidence 75

#### Scenario: Helper bridge match

- **GIVEN** an assertion expression containing an identifier whose
  `types.Object` maps through `helperBridge` to a key in `objToEffectID`
- **WHEN** `matchDirect` is called
- **THEN** it SHALL return an `AssertionMapping` with confidence 70

#### Scenario: No match

- **GIVEN** an assertion expression with no identifiers in `objToEffectID`
  or `helperBridge`
- **WHEN** `matchDirect` is called
- **THEN** it SHALL return `nil`

#### Scenario: Nil expression

- **GIVEN** a nil assertion expression
- **WHEN** `matchDirect` is called
- **THEN** it SHALL return `nil`

### Requirement: matchIndirectRoot helper

`matchIndirectRoot` MUST walk the assertion expression tree, identify
composite expressions (`SelectorExpr`, `IndexExpr`, `CallExpr`), resolve
them to root identifiers via `resolveExprRoot`, and check the root's
`types.Object` against `objToEffectID`. Matches SHALL have confidence 65.

#### Scenario: Selector expression match

- **GIVEN** an assertion expression `result.Name` where `result` is in
  `objToEffectID`
- **WHEN** `matchIndirectRoot` is called
- **THEN** it SHALL return an `AssertionMapping` with confidence 65

#### Scenario: No composite expressions

- **GIVEN** an assertion expression with only simple identifiers (no
  SelectorExpr, IndexExpr, or CallExpr)
- **WHEN** `matchIndirectRoot` is called
- **THEN** it SHALL return `nil`

### Requirement: collectTrackedVars helper

`collectTrackedVars` MUST iterate over `objToEffectID` and return all
`types.Object` keys whose effect ID matches `returnEffectID`.

#### Scenario: Multiple objects map to return effect

- **GIVEN** an `objToEffectID` map with 3 entries, 2 mapping to
  `returnEffectID`
- **WHEN** `collectTrackedVars` is called
- **THEN** the returned map SHALL contain exactly 2 entries

#### Scenario: No objects map to return effect

- **GIVEN** an `objToEffectID` map with entries but none matching
  `returnEffectID`
- **WHEN** `collectTrackedVars` is called
- **THEN** the returned map SHALL be empty

### Requirement: traceForwardDataFlow helper

`traceForwardDataFlow` MUST iterate over AST files in the test package,
walking assignment statements to trace tracked variables forward through
assignments, transformation calls (via `isTransformationCall`), and data
extractions (via `isDataExtraction`). It MUST run multiple iterations
until no new tracked variables are discovered or `maxContainerChainDepth`
is reached.

#### Scenario: Simple assignment chain

- **GIVEN** source code `x := target(); y := x.Field` with `x` in the
  tracked set
- **WHEN** `traceForwardDataFlow` is called
- **THEN** `y` SHALL be added to the tracked set

#### Scenario: Transformation call chain

- **GIVEN** source code with `json.Unmarshal(data, &result)` where `data`
  is tracked
- **WHEN** `traceForwardDataFlow` is called
- **THEN** `result` SHALL be added to the tracked set (via pointer
  destination extraction)

#### Scenario: Empty tracked set

- **GIVEN** an empty tracked set
- **WHEN** `traceForwardDataFlow` is called
- **THEN** it SHALL return an empty tracked set

#### Scenario: Convergence before max depth

- **GIVEN** source code where the tracked set stops growing after 2
  iterations (no new assignments reference tracked variables)
- **WHEN** `traceForwardDataFlow` is called with `maxContainerChainDepth` = 6
- **THEN** it SHALL terminate after 2 iterations, not 6

#### Scenario: Non-data-extraction gating

- **GIVEN** an assignment `got := s.Get("key")` where `s` is tracked but
  `Get` is a method call (not a field access, index, or type assertion)
- **WHEN** `traceForwardDataFlow` processes the assignment
- **THEN** `got` SHALL NOT be added to the tracked set (gated by
  `isDataExtraction`)

### Requirement: matchTrackedInExpr helper

`matchTrackedInExpr` MUST walk the assertion expression tree and return
`true` if any identifier's `types.Object` is in the tracked set. It MUST
check both direct identity and `resolveExprRoot` fallback for composite
expressions.

#### Scenario: Direct match

- **GIVEN** an expression containing an identifier in the tracked set
- **WHEN** `matchTrackedInExpr` is called
- **THEN** it SHALL return `true`

#### Scenario: Root resolution match

- **GIVEN** an expression `tracked.Field` where `tracked` is in the set
- **WHEN** `matchTrackedInExpr` is called
- **THEN** it SHALL return `true` (via `resolveExprRoot`)

#### Scenario: No match

- **GIVEN** an expression with no identifiers in the tracked set
- **WHEN** `matchTrackedInExpr` is called
- **THEN** it SHALL return `false`

## MODIFIED Requirements

### Requirement: isTransformationCall complexity

`isTransformationCall` MUST delegate byte-like detection to `isByteLikeParam`
and pointer destination detection to `isPointerDestParam`. The function's
cyclomatic complexity MUST be 8 or less.

Previously: All 5 type-checking patterns were inline in a single loop body
with complexity 26.

### Requirement: matchAssertionToEffect complexity

`matchAssertionToEffect` MUST delegate Pass 1 to `matchDirect` and Pass 2
to `matchIndirectRoot`. The function's cyclomatic complexity MUST be 10 or
less.

Previously: Both passes were inline with complexity 25.

### Requirement: matchContainerUnwrap complexity

`matchContainerUnwrap` MUST delegate to `collectTrackedVars`,
`traceForwardDataFlow`, and `matchTrackedInExpr`. The function's cyclomatic
complexity MUST be 12 or less.

Previously: All concerns were interleaved with complexity 50.

## REMOVED Requirements

None — pure decomposition, no functionality removed.
