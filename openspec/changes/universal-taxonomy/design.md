## Context

Gaze's taxonomy (`internal/taxonomy/types.go`) defines 38 `SideEffectType` constants organized into P0-P4 tiers. These constants use Go-specific names (`GoroutineSpawn`, `CgoCall`, `ChannelSend`, etc.) because gaze was originally a Go-only tool. With the external analyzer protocol now merged (Issue #95), non-Go analyzers must map their language concepts to these types. Several important effect categories — generators, decorators, context managers, monkey-patching — have no representation at all.

The provider-cleanup PR (#180, in progress) added 7 language-neutral aliases as a first step. This design completes the generalization.

## Goals / Non-Goals

### Goals
- Define a complete set of universal side effect types covering Go, Python, Rust, and TypeScript semantics
- Provide language-neutral aliases for all Go-specific type names
- Add a `Detail` field for language-specific metadata passthrough
- Add a `ValidTypes` registry for programmatic type validation
- Update JSON Schema and protocol documentation
- Maintain full backward compatibility for existing Go analysis

### Non-Goals
- Renaming existing Go-specific constants (they remain the primary constants for Go analysis)
- Changing scoring, classification, or CRAP computation logic
- Implementing language-specific detection logic (that's the analyzer's job)
- Adding new protocol methods or changing the JSON-RPC message format
- Supporting custom/user-defined effect types (external types default to P4 with a warning, as they do today)

## Decisions

### D1: Additive constants, not renames

New universal types are added as new `SideEffectType` constants with their own string values (e.g., `GeneratorYield = "GeneratorYield"`). Existing Go-specific constants (`GoroutineSpawn = "GoroutineSpawn"`) are unchanged. Language-neutral aliases point to the Go-specific string values (e.g., `AsyncTaskSpawn SideEffectType = GoroutineSpawn`).

**Rationale**: No breaking changes to JSON output. Go analysis continues to emit the same type strings. External analyzers can use either the Go name or the neutral alias — both resolve to the same string value, so `tierMap` lookups and equality checks work identically.

**Constitution alignment**: Minimal Assumptions (Principle II) — existing Go functionality is unaffected; no new assumptions imposed.

### D2: Detail as `map[string]any`

The `Detail` field on `SideEffect` is typed as `map[string]any` with `json:"detail,omitempty"`. It is opaque to gaze's scoring engine — passed through to reports and AI pipelines without interpretation.

**Rationale**: A typed struct per language would couple the taxonomy package to language-specific schemas. A `map[string]any` is the simplest representation that supports arbitrary metadata. The `omitempty` tag ensures Go analysis output is unchanged (Detail is nil/empty for Go effects).

**Trade-off**: No compile-time validation of detail contents. This is acceptable because detail is informational, not used for scoring.

**Constitution alignment**: Actionable Output (Principle III) — detail is machine-parseable JSON that enhances report actionability; Minimal Assumptions (Principle II) — analyzers produce self-describing detail without runtime coupling to gaze's type system.

### D3: ValidTypes as a package-level set

Add a `ValidTypes` variable of type `map[SideEffectType]bool` that contains every canonical type constant (not aliases, since aliases resolve to the same string value). Add `IsKnownType(t SideEffectType) bool` as the public API.

**Rationale**: Replaces the hardcoded `isKnownP4` switch in `adapter/sideeffect.go` with a single source of truth. Enables table-driven tests that verify `tierMap` completeness against `ValidTypes`.

**Constitution alignment**: Testability (Principle IV) — enables completeness assertions in tests.

### D4: Tier assignments follow semantic priority

New universal types are assigned tiers based on the same semantic criteria as existing types:

- **P0**: Core function contract (return values, error signals, receiver/argument mutation)
- **P1**: High-value common effects (container mutation, stream output, generators, channel ops)
- **P2**: Infrastructure and language-specific effects (metaprogramming, descriptors, resource management, import effects, monkey-patching)

`ErrorSignal` is P0 because it generalizes `ErrorReturn` (which is P0). `GeneratorYield` is P1 because it generalizes `ReturnValue` semantics for lazy iteration — a core function contract element. `ContainerMutation` is P1 because it generalizes `SliceMutation`/`MapMutation` (both P1). `StreamOutput` is P1 because it generalizes `WriterOutput` (P1).

### D5: Relationship between generic and specific types

Generic types like `ContainerMutation` and specific types like `SliceMutation`/`MapMutation` coexist. Go analysis continues to emit the specific types. External analyzers use whichever granularity matches their language:

- Python analyzer reports `ContainerMutation` for `list.append()` (no need to distinguish slice vs. map)
- Go analyzer continues to report `SliceMutation` and `MapMutation` (more precise)

Both are valid. `tierMap` has entries for all of them. There is no inheritance or subsumption relationship in the type system — they are independent constants that happen to be semantically related.

### D6: No changes to classification or scoring

The `classify.ComputeScore` function and `tierBoost` logic are unchanged. New types flow through the same scoring pipeline:

1. External analyzer reports `GeneratorYield` → adapter converts to `taxonomy.SideEffect{Type: GeneratorYield}`
2. `TierOf(GeneratorYield)` returns `TierP1` (from updated `tierMap`)
3. `tierBoost` returns +10 for P1
4. Classification proceeds normally

No special-casing for new types in the scoring engine.

### D7: Detail field flows through the protocol layer

The `Detail` field must exist at three layers: `protocol.AnalyzedSideEffect` (JSON-RPC response from analyzer), `taxonomy.SideEffect` (internal representation), and JSON output (report). The adapter's `convertAnalysisResults` copies `Detail` from protocol to taxonomy. This ensures external analyzers can send detail metadata and it reaches reports without any layer dropping it.

### D8: Detail field size risk

The `Detail` map has no size limit. The protocol client reads entire JSON-RPC messages into memory, so a malicious analyzer could already send arbitrarily large `side_effects` arrays. The `Detail` field adds no new blast radius beyond what already exists. Protocol documentation will advise analyzer authors to keep detail payloads small (< 1KB per effect).

## Coverage Strategy

All new code is covered by unit tests:
- `ValidTypes`/`IsKnownType`: completeness test (task 2.1a-c) serves as the regression ratchet — ensures every canonical type has a tier mapping and vice versa.
- Alias equivalence: table-driven test (task 2.1d) verifies all 11 aliases.
- `Detail` serialization: round-trip test (task 2.2) with specific value assertions.
- Adapter `Detail` passthrough: integration test (task 3.2) verifying protocol → taxonomy conversion.
- No integration or e2e tests needed because the new types flow through existing, already-tested pipelines (classification, CRAP scoring) without any code path changes.
- Expected coverage: near 100% for new code in `types.go` and `priority.go` (constants + lookup function).

## Risks / Trade-offs

### R1: Alias collision with PR #180

PR #180 adds 7 aliases. This change adds a superset (11 aliases). If #180 merges first, this change will have a merge conflict in the alias block. **Mitigation**: The conflict is trivially resolvable — this change's alias block replaces #180's.

### R2: JSON Schema enum expansion

Adding types to the `SideEffect.type` enum in the JSON Schema is technically a breaking change for strict schema validators that reject unknown enum values. **Mitigation**: Schema version will be bumped. Gaze's own schema validation tests use `additionalProperties: true` style — new enum values are additive. The `type` field already receives unknown values from external analyzers (which default to P4), so consumers should already handle unknown types gracefully.

### R3: ValidTypes maintenance burden

Every new `SideEffectType` constant must be added to both `tierMap` and `ValidTypes`. Forgetting one creates an inconsistency. **Mitigation**: Add a test that asserts `len(tierMap) == len(ValidTypes)` and that every key in one appears in the other.

### R4: Detail field size

See design decision D8 for full analysis. The `Detail` field adds no new blast radius beyond the existing protocol message size risk. Protocol documentation will advise analyzer authors to keep detail payloads small (< 1KB per effect).

### R5: Semantic overlap between generic and specific types

Having both `ErrorSignal` and `ErrorReturn` could confuse analyzer authors about which to use. **Mitigation**: Protocol documentation will explicitly state that Go analyzers use the specific types and external analyzers use whichever matches their semantics. The mapping table in `docs/protocol.md` will list both with clear guidance. Note: `ProcessTermination` aliases `Panic` (P2) rather than `ProcessExit` (P3) because Go's `panic` is the closest semantic match for abrupt, potentially-catchable termination (Python's `SystemExit`, Rust's `panic!`). `ProcessExit` covers clean exit (`os.Exit`, `sys.exit`). This distinction will be documented in the protocol type table.
