## ADDED Requirements

### Requirement: analyzeModel.Update test coverage

The `(analyzeModel).Update` method in `cmd/gaze/interactive.go` MUST have unit
tests covering all message-type branches. Tests MUST NOT be guarded by
`testing.Short()`.

#### Scenario: First WindowSizeMsg initializes viewport

- **GIVEN** a freshly constructed `analyzeModel` via `newAnalyzeModel(results)`
  where `ready` is false
- **WHEN** `Update` is called with a `tea.WindowSizeMsg{Width: 80, Height: 24}`
- **THEN** the returned model has `ready == true`, viewport width 80, viewport
  height equal to `24 - footerHeight`, and the viewport content matches the
  pre-rendered content string

#### Scenario: Subsequent WindowSizeMsg resizes viewport

- **GIVEN** an `analyzeModel` that has already processed one `WindowSizeMsg`
  (i.e., `ready == true`)
- **WHEN** `Update` is called with a second `tea.WindowSizeMsg{Width: 120, Height: 40}`
- **THEN** the returned model has viewport width 120 and viewport height equal
  to `40 - footerHeight`, and `ready` remains true

#### Scenario: Quit key returns tea.Quit command

- **GIVEN** a ready `analyzeModel`
- **WHEN** `Update` is called with a `tea.KeyMsg` matching the quit binding
  (key "q")
- **THEN** the returned command is `tea.Quit`

#### Scenario: Help key toggles help visibility

- **GIVEN** a ready `analyzeModel` with `help.ShowAll == false`
- **WHEN** `Update` is called with a `tea.KeyMsg` matching the help binding
  (key "?")
- **THEN** the returned model has `help.ShowAll == true`

#### Scenario: Unhandled message type passes through

- **GIVEN** a ready `analyzeModel`
- **WHEN** `Update` is called with an unrecognized message type
- **THEN** the returned model is unchanged and no error or panic occurs

### Requirement: isPointerArgStore test coverage

The `isPointerArgStore` function in `internal/analysis/mutation.go` MUST have
unit tests covering all branch paths. An export shim `IsPointerArgStore` MUST
be added to `internal/analysis/export_test.go` following the project's
established pattern. Tests MUST NOT be guarded by `testing.Short()`.

#### Scenario: Direct trace to pointer parameter

- **GIVEN** a `*ssa.Store` instruction whose `Addr` directly traces to a pointer
  parameter
- **WHEN** `IsPointerArgStore` is called with the store and a `ptrParams` map
  containing that parameter
- **THEN** it returns the parameter name and `true`

#### Scenario: UnOp dereference through pointer parameter

- **GIVEN** a `*ssa.Store` instruction whose `Addr` is a `*ssa.UnOp`
  (dereference) of a pointer parameter (e.g., `*dst = ...`)
- **WHEN** `IsPointerArgStore` is called
- **THEN** it returns the parameter name and `true`

#### Scenario: FieldAddr through pointer parameter

- **GIVEN** a `*ssa.Store` instruction whose `Addr` is a `*ssa.FieldAddr` whose
  base traces to a pointer parameter (e.g., `cfg.Timeout = val`)
- **WHEN** `IsPointerArgStore` is called
- **THEN** it returns the parameter name and `true`

#### Scenario: IndexAddr through pointer parameter

- **GIVEN** a `*ssa.Store` instruction whose `Addr` is a `*ssa.IndexAddr` whose
  base traces to a pointer parameter (e.g., `v[0] = 1.0`)
- **WHEN** `IsPointerArgStore` is called
- **THEN** it returns the parameter name and `true`

#### Scenario: Store to local variable returns false

- **GIVEN** a `*ssa.Store` instruction whose `Addr` traces to a local variable,
  not a pointer parameter
- **WHEN** `IsPointerArgStore` is called with a `ptrParams` map that does not
  include the local variable
- **THEN** it returns `("", false)`

#### Scenario: Empty ptrParams returns false

- **GIVEN** any `*ssa.Store` instruction
- **WHEN** `IsPointerArgStore` is called with an empty `ptrParams` map
- **THEN** it returns `("", false)`

#### Scenario: Multiple pointer parameters identifies correct one

- **GIVEN** a function with multiple pointer parameters and a `*ssa.Store`
  instruction that writes through the second parameter
- **WHEN** `IsPointerArgStore` is called with a `ptrParams` map containing
  both parameters
- **THEN** it returns the name of the second parameter and `true`

### Requirement: Testdata fixture for FieldAddr branch

A new fixture function MUST be added to
`internal/analysis/testdata/src/mutation/mutation.go` that writes to a field of
a struct pointer parameter. The fixture MUST use only Go stdlib types (no
external imports). The fixture MUST produce an `*ssa.FieldAddr` store address
when compiled to SSA.

#### Scenario: SetTimeout fixture produces FieldAddr store

- **GIVEN** the mutation testdata package is loaded and built into SSA
- **WHEN** the SSA instructions for `SetTimeout` are inspected
- **THEN** at least one `*ssa.Store` instruction has an `*ssa.FieldAddr` address
  whose base traces to the `cfg` parameter

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
