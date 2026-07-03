## Why

Gaze's side effect taxonomy (`internal/taxonomy/types.go`) defines 38 effect types rooted in Go semantics. With the external analyzer protocol (Issue #95, PR #178) now merged, external analyzers for other languages need to map their language concepts to this taxonomy. The current type names are Go-centric (`GoroutineSpawn`, `CgoCall`, `DeferredReturnMutation`, `WaitGroupOp`, `SyncPoolOp`, etc.), which creates friction for analyzer authors and confuses users reading reports for non-Go codebases.

Issue #96 proposes a universal abstract type system that covers side effects across languages while preserving backward compatibility for Go analysis. The provider-cleanup PR (#180, not yet merged) added 7 language-neutral aliases and renamed `GoVersion` to `LanguageVersion` as a first step. This change is designed to work whether #180 merges first or not. It completes the taxonomy generalization by:

1. Adding the remaining universal types that have no Go equivalent (Python generators, decorators, context managers, monkey-patching, etc.)
2. Adding language-neutral aliases for all Go-specific names
3. Adding a `Detail` field to `SideEffect` for language-specific metadata
4. Updating tier mappings, the JSON Schema, and protocol documentation
5. Adding a `ValidTypes` registry for programmatic type validation

This is a prerequisite for the Python analyzer ([snake-eyes](https://github.com/zero-dot-force/snake-eyes)) to report its full range of side effects using gaze's taxonomy rather than inventing ad-hoc type strings that default to P4.

## What Changes

### New Universal Effect Types

Add 10 new `SideEffectType` constants for effects that exist in other languages but have no Go equivalent:

| Type | Tier | Semantic Definition |
|------|------|---------------------|
| `GeneratorYield` | P1 | Producing a value via lazy iteration (Python `yield`, JS generators) |
| `AsyncGeneratorYield` | P2 | Async lazy iteration with side effects |
| `MetaprogrammingMutation` | P2 | Runtime code/class modification (decorators, metaclasses) |
| `DescriptorEffect` | P2 | Attribute access protocol side effects (`__set__`, `__delete__`) |
| `ResourceManagement` | P2 | Scoped resource acquire/release (`__enter__`/`__exit__`, `using`) |
| `ImportSideEffect` | P2 | Code executed at module load time (Go `init()`, Python top-level) |
| `MonkeyPatch` | P2 | Runtime modification of external module/class attributes |
| `ContainerMutation` | P1 | In-place mutation of a collection (generic; subsumes `SliceMutation`/`MapMutation` for non-Go) |
| `StreamOutput` | P1 | Writing to an output stream (generic; subsumes `WriterOutput`/`StdoutWrite`/`StderrWrite` for non-Go) |
| `ErrorSignal` | P0 | Function signals a failure condition (generic; subsumes `ErrorReturn` for non-Go — exceptions, panics, Result::Err) |

### Language-Neutral Aliases for Go-Specific Types

Add aliases so external analyzers can use semantic names instead of Go-specific ones. These are additive — the Go names remain valid and continue to be the primary constants used by Go analysis:

| Alias | Maps To | Rationale |
|-------|---------|-----------|
| `AsyncTaskSpawn` | `GoroutineSpawn` | Generic concurrent task creation |
| `AsyncMessageSend` | `ChannelSend` | Generic message passing |
| `AsyncChannelClose` | `ChannelClose` | Generic channel/queue close |
| `BarrierOp` | `WaitGroupOp` | Generic synchronization barrier |
| `PanicRecovery` | `RecoverBehavior` | Generic panic/exception recovery |
| `FFICall` | `CgoCall` | Generic foreign function interface |
| `ObjectPoolOp` | `SyncPoolOp` | Generic object pool operation |
| `DeferredMutation` | `DeferredReturnMutation` | Generic deferred cleanup mutation |
| `ArgumentMutation` | `PointerArgMutation` | Generic argument mutation |
| `ProcessTermination` | `Panic` | Generic process termination (maps Panic for cross-language use) |
| `SentinelErrorDecl` | `SentinelError` | Generic sentinel error/exception class declaration |

### SideEffect Detail Field

Add an optional `Detail` field to `taxonomy.SideEffect` for language-specific metadata that gaze's scoring engine does not interpret but passes through to reports and AI pipelines:

```go
type SideEffect struct {
    // ... existing fields ...
    Detail map[string]any `json:"detail,omitempty"`
}
```

Example: a Python analyzer reports `ErrorSignal` with detail `{"language_type": "RaiseException", "exception_class": "ValueError"}`.

### ValidTypes Registry

Add a `ValidTypes` set (or `IsKnownType` function) to `taxonomy` so that the adapter layer and protocol documentation can programmatically enumerate all valid types instead of hardcoding switch statements.

### Tier Mappings

Update `tierMap` in `priority.go` to include all new universal types at their assigned tiers. The existing Go-specific types retain their current tier assignments.

### JSON Schema Updates

Update the `SideEffect.type` enum in `internal/report/schema.go` to include all new universal types. Add the `detail` property to the `SideEffect` schema definition.

### Protocol Documentation

Update `docs/protocol.md` to document the full universal type table with cross-language mapping examples.

## Capabilities

### New Capabilities
- `taxonomy-universal-types`: 10 new language-neutral `SideEffectType` constants for non-Go effects
- `taxonomy-neutral-aliases`: 11 language-neutral aliases for Go-specific constants
- `taxonomy-detail-field`: `SideEffect.Detail` map for language-specific metadata passthrough
- `taxonomy-valid-types`: `IsKnownType` function and `ValidTypes` set for programmatic type validation

### Modified Capabilities
- `internal/taxonomy/priority.go`: `tierMap` expanded with new universal types
- `internal/report/schema.go`: `SideEffect.type` enum expanded; `detail` property added
- `internal/adapter/sideeffect.go`: `isKnownP4` replaced by `IsKnownType` from taxonomy
- `docs/protocol.md`: Universal type table with cross-language mappings

### Removed Capabilities
- None. All existing Go-specific type constants and their wire values remain unchanged.

## Impact

- **Modified files**: `internal/taxonomy/types.go` (new types, aliases, Detail field, ValidTypes), `internal/taxonomy/priority.go` (expanded tierMap), `internal/report/schema.go` (schema enum + detail), `internal/adapter/sideeffect.go` (use IsKnownType, propagate Detail), `internal/protocol/types.go` (add Detail to AnalyzedSideEffect), `docs/protocol.md` (universal type table), `docs/porting/taxonomy-reference.md`, `docs/porting/contracts.md`, `docs/porting/requirements.md` (updated type counts and tier tables)
- **New files**: `internal/taxonomy/types_test.go` (if not existing; tests for ValidTypes, aliases, tier completeness)
- **No breaking changes**: All existing `SideEffectType` constants retain their string values. JSON output for Go analysis is unchanged. The `detail` field uses `omitempty` and is absent when not populated.
- **JSON output addition**: New `detail` field on `SideEffect` (optional, omitempty). Consumers ignoring unknown fields are unaffected.
- **Coordination with PR #180**: This change is based on `main` and is independent of `opsx/provider-cleanup`. If #180 merges first, the aliases it added will conflict — the resolution is straightforward (this change's alias block supersedes #180's, since it's a superset).

## Constitution Alignment

Assessed against the Gaze project constitution (v1.3.0).

### I. Accuracy

**Assessment**: PASS

New universal types are additive. Go analysis continues to use the same type constants with the same string values — no detection, classification, or scoring logic changes. External analyzers gain the ability to report effects accurately using semantically correct type names rather than forcing Go-specific names onto non-Go concepts.

### II. Minimal Assumptions

**Assessment**: PASS

The taxonomy makes no assumptions about which language produced the data. Types are defined by semantic meaning ("function signals a failure condition"), not language syntax ("function returns error"). External analyzers use whichever type name matches their semantics. The `Detail` field is fully optional and opaque to the scoring engine.

### III. Actionable Output

**Assessment**: PASS

Reports using universal types are equally actionable. The `Detail` field enhances actionability by providing language-specific context (e.g., exception class names, decorator details) that AI report pipelines can surface. Cross-run comparability is maintained because type string values are stable.

### IV. Testability

**Assessment**: PASS

Universal types can be tested with language-neutral fixtures. `IsKnownType` enables table-driven tests that verify completeness (every constant in tierMap is in ValidTypes and vice versa). Alias equivalence is trivially testable. The `Detail` field is a plain map — no special serialization logic to test beyond standard JSON marshaling.
