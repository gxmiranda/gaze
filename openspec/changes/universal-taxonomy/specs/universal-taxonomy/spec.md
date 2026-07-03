## ADDED Requirements

### Requirement: Universal Side Effect Types

The taxonomy package MUST define the following new `SideEffectType` constants for language-neutral side effect categories that have no existing Go-specific equivalent:

| Constant | String Value | Tier |
|----------|-------------|------|
| `GeneratorYield` | `"GeneratorYield"` | P1 |
| `AsyncGeneratorYield` | `"AsyncGeneratorYield"` | P2 |
| `MetaprogrammingMutation` | `"MetaprogrammingMutation"` | P2 |
| `DescriptorEffect` | `"DescriptorEffect"` | P2 |
| `ResourceManagement` | `"ResourceManagement"` | P2 |
| `ImportSideEffect` | `"ImportSideEffect"` | P2 |
| `MonkeyPatch` | `"MonkeyPatch"` | P2 |
| `ContainerMutation` | `"ContainerMutation"` | P1 |
| `StreamOutput` | `"StreamOutput"` | P1 |
| `ErrorSignal` | `"ErrorSignal"` | P0 |

Each constant MUST have its own unique string value distinct from existing Go-specific types. Constants MUST be declared in a clearly commented block (e.g., "Universal types — language-neutral").

#### Scenario: External analyzer reports GeneratorYield

- **GIVEN** an external Python analyzer is connected via the JSON-RPC protocol
- **WHEN** the analyzer reports a side effect with `type: "GeneratorYield"`
- **THEN** gaze's adapter converts it to `taxonomy.SideEffect{Type: taxonomy.GeneratorYield}`
- **AND** `taxonomy.TierOf(taxonomy.GeneratorYield)` returns `TierP1`
- **AND** the effect flows through classification and scoring normally

#### Scenario: External analyzer reports ErrorSignal

- **GIVEN** an external analyzer reports a side effect with `type: "ErrorSignal"`
- **WHEN** gaze processes the effect through the classification engine
- **THEN** `tierBoost` returns +25 (P0 boost) for `ErrorSignal`
- **AND** the effect receives the same default contractual classification as `ErrorReturn`

### Requirement: Language-Neutral Aliases

The taxonomy package MUST define the following language-neutral aliases as `SideEffectType` constants whose string values equal the Go-specific constant they alias:

| Alias | Aliases | String Value |
|-------|---------|-------------|
| `AsyncTaskSpawn` | `GoroutineSpawn` | `"GoroutineSpawn"` |
| `AsyncMessageSend` | `ChannelSend` | `"ChannelSend"` |
| `AsyncChannelClose` | `ChannelClose` | `"ChannelClose"` |
| `BarrierOp` | `WaitGroupOp` | `"WaitGroupOp"` |
| `PanicRecovery` | `RecoverBehavior` | `"RecoverBehavior"` |
| `FFICall` | `CgoCall` | `"CgoCall"` |
| `ObjectPoolOp` | `SyncPoolOp` | `"SyncPoolOp"` |
| `DeferredMutation` | `DeferredReturnMutation` | `"DeferredReturnMutation"` |
| `ArgumentMutation` | `PointerArgMutation` | `"PointerArgMutation"` |
| `ProcessTermination` | `Panic` | `"Panic"` |
| `SentinelErrorDecl` | `SentinelError` | `"SentinelError"` |

Aliases MUST be declared in a separate, clearly commented block from the canonical Go types. Aliases MUST NOT be added to `tierMap` (they resolve to the same string value as their target, which is already in `tierMap`).

#### Scenario: Alias resolves to canonical string value

- **GIVEN** an external analyzer uses `"GoroutineSpawn"` as the side effect type
- **WHEN** Go code references `taxonomy.AsyncTaskSpawn`
- **THEN** `taxonomy.AsyncTaskSpawn == taxonomy.GoroutineSpawn` is `true`
- **AND** `taxonomy.TierOf(taxonomy.AsyncTaskSpawn)` returns `TierP2` (same as `GoroutineSpawn`)

### Requirement: SideEffect Detail Field

The `taxonomy.SideEffect` struct MUST include a `Detail` field typed as `map[string]any` with JSON tag `json:"detail,omitempty"`.

The `Detail` field MUST be opaque to gaze's scoring, classification, and CRAP computation logic. It MUST be passed through to JSON output and AI report pipelines without interpretation.

When `Detail` is nil or empty, it MUST be omitted from JSON output (ensured by `omitempty`).

#### Scenario: External analyzer provides language detail

- **GIVEN** an external analyzer reports a side effect with detail metadata
- **WHEN** gaze serializes the analysis result to JSON
- **THEN** the `detail` field appears in the JSON output with the analyzer's metadata
- **AND** the `detail` field is not consulted by `classify.ComputeScore` or `crap.Analyze`

#### Scenario: Go analysis produces no detail

- **GIVEN** gaze's built-in Go analyzer produces a side effect
- **WHEN** gaze serializes the analysis result to JSON
- **THEN** the `detail` field is absent from the JSON output (omitempty)

### Requirement: ValidTypes Registry

The taxonomy package MUST export an `IsKnownType(t SideEffectType) bool` function that returns `true` for every canonical (non-alias) `SideEffectType` constant.

The taxonomy package SHOULD export a `ValidTypes` variable or function that returns the complete set of known types for enumeration.

`IsKnownType` MUST return `true` for all new universal types and all existing Go-specific types. Aliases are NOT separately added to `ValidTypes` — they resolve to the same string value as their canonical target, which IS in `ValidTypes`, so `IsKnownType` returns `true` for aliases. `IsKnownType` MUST return `false` for unknown/arbitrary strings not matching any canonical type's string value.

#### Scenario: Known type lookup

- **GIVEN** all canonical `SideEffectType` constants
- **WHEN** `IsKnownType` is called with each constant
- **THEN** it returns `true` for every one

#### Scenario: Unknown type lookup

- **GIVEN** a string `"FooBarEffect"` that is not a defined constant
- **WHEN** `IsKnownType(SideEffectType("FooBarEffect"))` is called
- **THEN** it returns `false`

#### Scenario: Alias lookup

- **GIVEN** the alias `AsyncTaskSpawn` (string value `"GoroutineSpawn"`)
- **WHEN** `IsKnownType(AsyncTaskSpawn)` is called
- **THEN** it returns `true` (because `"GoroutineSpawn"` is a canonical type's string value)

### Requirement: Detail Field Protocol Passthrough

The `protocol.AnalyzedSideEffect` struct in `internal/protocol/types.go` MUST include a `Detail` field typed as `map[string]any` with JSON tag `json:"detail,omitempty"`.

The adapter's `convertAnalysisResults` function in `internal/adapter/sideeffect.go` MUST copy the `Detail` field from `protocol.AnalyzedSideEffect` to `taxonomy.SideEffect` during conversion. This ensures the end-to-end passthrough: external analyzer → protocol → adapter → taxonomy → JSON output.

#### Scenario: Detail flows through protocol to taxonomy

- **GIVEN** an external analyzer returns a side effect with `detail: {"language_type": "RaiseException", "exception_class": "ValueError"}`
- **WHEN** the adapter converts the protocol response to taxonomy types
- **THEN** the resulting `taxonomy.SideEffect.Detail` map contains `{"language_type": "RaiseException", "exception_class": "ValueError"}`
- **AND** the detail appears in the final JSON output

### Requirement: Tier Map Completeness

Every canonical `SideEffectType` constant MUST have an entry in `tierMap`. Every entry in `tierMap` MUST correspond to a canonical constant.

There MUST be a test that asserts `len(tierMap)` equals the number of canonical types and that every canonical type appears in `tierMap`.

#### Scenario: Tier map matches valid types

- **GIVEN** the `ValidTypes` set and `tierMap`
- **WHEN** the completeness test runs
- **THEN** every key in `ValidTypes` appears in `tierMap`
- **AND** every key in `tierMap` appears in `ValidTypes`

### Requirement: JSON Schema Update

The `SideEffect.type` enum in `internal/report/schema.go` MUST include all new universal type constants.

The `SideEffect` schema definition MUST include an optional `detail` property of type `object`.

#### Scenario: JSON output validates against updated schema

- **GIVEN** analysis JSON output containing a `GeneratorYield` side effect with detail
- **WHEN** validated against the updated JSON Schema
- **THEN** validation passes

### Requirement: Protocol Documentation Update

`docs/protocol.md` MUST include a comprehensive universal type table showing:
- Universal type name
- Tier assignment
- Semantic definition
- Example mappings for at least Go and Python

The table MUST include all 10 new universal types and all 38 existing types (48 total). The existing phantom types `FileSystemRead` and `NetworkCall` in the current protocol documentation MUST be removed (they are not defined in the taxonomy).

#### Scenario: Analyzer author references documentation

- **GIVEN** a developer building a new external analyzer
- **WHEN** they read `docs/protocol.md`
- **THEN** they find a complete type reference table with language-specific mapping examples
- **AND** they can determine which type to use for their language's concepts

## MODIFIED Requirements

### Requirement: Adapter Unknown Type Warning

The `isKnownP4` function in `internal/adapter/sideeffect.go` MUST be replaced by a call to `taxonomy.IsKnownType`. The warning logic ("unknown side effect type from external analyzer") MUST use `IsKnownType` instead of a hardcoded switch statement.

Previously: `isKnownP4` was a local switch statement listing the 6 P4 types. The adapter warned on any type that was P4-by-default AND not in the switch.

Now: The adapter warns on any type where `!taxonomy.IsKnownType(effectType)`. This is simpler and automatically stays current as new types are added.

#### Scenario: Known P2 type does not trigger warning

- **GIVEN** an external analyzer reports `MetaprogrammingMutation`
- **WHEN** the adapter converts the response
- **THEN** no warning is logged (it is a known type at P2)

#### Scenario: Unknown type triggers warning

- **GIVEN** an external analyzer reports `"CustomFooEffect"`
- **WHEN** the adapter converts the response
- **THEN** a warning is logged to stderr: `unknown side effect type "CustomFooEffect" from external analyzer (defaulting to P4)`

## REMOVED Requirements

None. All existing types, tier assignments, and behaviors are preserved.
