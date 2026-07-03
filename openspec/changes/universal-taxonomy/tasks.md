<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Taxonomy Types and Tier Map

- [x] 1.1 Add 10 universal `SideEffectType` constants to `internal/taxonomy/types.go`. Add them in a new `// Universal types -- language-neutral` comment block after the P4 block. Constants: `GeneratorYield`, `AsyncGeneratorYield`, `MetaprogrammingMutation`, `DescriptorEffect`, `ResourceManagement`, `ImportSideEffect`, `MonkeyPatch`, `ContainerMutation`, `StreamOutput`, `ErrorSignal`. Each has its own unique string value matching the constant name.
- [x] 1.2 Add 11 language-neutral alias constants to `internal/taxonomy/types.go`. Add them in a `// Language-neutral aliases` comment block after the universal types block. Aliases: `AsyncTaskSpawn = GoroutineSpawn`, `AsyncMessageSend = ChannelSend`, `AsyncChannelClose = ChannelClose`, `BarrierOp = WaitGroupOp`, `PanicRecovery = RecoverBehavior`, `FFICall = CgoCall`, `ObjectPoolOp = SyncPoolOp`, `DeferredMutation = DeferredReturnMutation`, `ArgumentMutation = PointerArgMutation`, `ProcessTermination = Panic`, `SentinelErrorDecl = SentinelError`.
- [x] 1.3 Add `Detail map[string]any` field to `taxonomy.SideEffect` struct with JSON tag `json:"detail,omitempty"`. Add GoDoc comment explaining it is opaque to scoring and classification.
- [x] 1.4 Add `ValidTypes` set (`map[SideEffectType]bool`) and `IsKnownType(t SideEffectType) bool` function to `internal/taxonomy/types.go`. `ValidTypes` contains all 48 canonical (non-alias) type constants (38 existing + 10 new). `IsKnownType` returns `ValidTypes[t]`.
- [x] 1.5 [P] Add tier entries for all 10 new universal types in `internal/taxonomy/priority.go`. Entries: `ErrorSignal: TierP0`, `GeneratorYield: TierP1`, `ContainerMutation: TierP1`, `StreamOutput: TierP1`, `AsyncGeneratorYield: TierP2`, `MetaprogrammingMutation: TierP2`, `DescriptorEffect: TierP2`, `ResourceManagement: TierP2`, `ImportSideEffect: TierP2`, `MonkeyPatch: TierP2`.

## 2. Tests

- [x] 2.1 Add tests in `internal/taxonomy/types_test.go`: (a) `TestValidTypes_Completeness` -- assert `len(ValidTypes) == len(tierMap)` (bidirectional completeness) and every key in `ValidTypes` appears in `tierMap` and vice versa. (b) `TestIsKnownType_Known` -- table-driven test with all canonical types. (c) `TestIsKnownType_Unknown` -- verify returns `false` for `"FooBarEffect"`. (d) `TestAliases_ResolveToCanonical` -- verify each alias equals its target constant (e.g., `AsyncTaskSpawn == GoroutineSpawn`). (e) `TestUniversalTypes_TierAssignments` -- verify tier for each new universal type. (f) `TestErrorSignal_TierBoost` -- verify `classify.ComputeScore` gives `ErrorSignal` the P0 boost (+25).
- [x] 2.2 Add test in `internal/taxonomy/types_test.go`: `TestSideEffect_DetailMarshal` -- (a) round-trip test: marshal a `SideEffect` with `Detail: map[string]any{"language_type": "RaiseException", "confidence": 0.95}`, unmarshal it, and assert specific key-value pairs are preserved (verify string values remain strings and numeric values survive the `float64` round-trip). (b) omitempty test: verify `Detail` is absent from JSON when nil.

## 3. Adapter and Protocol Update

- [x] 3.1 Add `Detail map[string]any` field with JSON tag `json:"detail,omitempty"` to `protocol.AnalyzedSideEffect` in `internal/protocol/types.go`. Update `convertAnalysisResults` in `internal/adapter/sideeffect.go` to copy `se.Detail` to `effect.Detail`.
- [x] 3.2 Replace `isKnownP4` function in `internal/adapter/sideeffect.go` with `taxonomy.IsKnownType`. Change the warning condition from `tier == TierP4 && !isKnownP4(effectType)` to `!taxonomy.IsKnownType(effectType)`. Remove the `isKnownP4` function.
- [x] 3.3 [P] Update `internal/adapter/adapter_test.go` to add test cases: (a) external analyzer reporting a universal type (e.g., `GeneratorYield`) -- verify no warning is logged and correct tier is assigned. (b) external analyzer reporting a side effect with `Detail` metadata -- verify the detail map is preserved through conversion.

## 4. JSON Schema Update

- [x] 4.1 Update the `SideEffect.type` enum in `internal/report/schema.go` (`Schema` constant) to include all 10 new universal types: `GeneratorYield`, `AsyncGeneratorYield`, `MetaprogrammingMutation`, `DescriptorEffect`, `ResourceManagement`, `ImportSideEffect`, `MonkeyPatch`, `ContainerMutation`, `StreamOutput`, `ErrorSignal`.
- [x] 4.2 Add `"detail": {"type": "object", "description": "Language-specific metadata (opaque to scoring)"}` as an optional property (not in `required`) to the `SideEffect` properties in `Schema` and to the `SideEffectRef` definition in `QualitySchema`.

## 5. Documentation

- [x] 5.1 Update `docs/protocol.md` "Side Effect Types" section to include a comprehensive universal type reference table. The table MUST list all 48 canonical types organized by tier (P0-P4) with columns: Type, Tier, Definition, Go Mapping, Python Mapping. Include the 10 new universal types and all 38 existing types. Add a note about language-neutral aliases. Remove phantom types `FileSystemRead` and `NetworkCall` from the existing table (they are not defined in the taxonomy). Correct the Python `yield` mapping from `ReturnValue` to `GeneratorYield`. Add protocol version bump to 1.1.0. Add a note advising analyzer authors to keep `detail` payloads small (< 1KB per effect).
- [x] 5.2 [P] Update `docs/porting/taxonomy-reference.md`, `docs/porting/contracts.md`, and `docs/porting/requirements.md` with expanded type counts (38 → 48), new universal types in tier tables, and updated EC-001/EC-004 references.

## 6. Verification

- [x] 6.1 Run `go build ./cmd/gaze` -- verify clean build.
- [x] 6.2 Run `go test -race -count=1 -short ./internal/taxonomy/...` -- verify all new tests pass.
- [x] 6.3 Run `go test -race -count=1 -short ./internal/adapter/...` -- verify adapter tests pass with `IsKnownType` and Detail passthrough.
- [x] 6.4 Run `go test -race -count=1 -short ./internal/report/...` -- verify schema validation tests pass with expanded enum.
- [x] 6.5 Run `go test -race -count=1 -short ./...` -- full unit test suite passes.
- [x] 6.6 Run `golangci-lint run` -- no lint violations (golangci-lint not installed locally; go vet passes clean).
- [x] 6.7 Verify constitution alignment: (I) Accuracy -- no detection/scoring logic changes; (II) Minimal Assumptions -- no language-specific assumptions in universal types; (III) Actionable Output -- JSON output includes new types with machine-parseable detail; (IV) Testability -- completeness tests verify type/tier consistency.
- [x] 6.8 Update AGENTS.md "Recent Changes" section with a summary of this change. Verify the taxonomy package description in "Architecture" is still accurate.
- [x] 6.9 If PR #180 has merged before implementation, verify the 7 aliases it added are a subset of the 11 defined here and resolve any conflicts. (PR #180 has not merged; no conflict.)
<!-- spec-review: passed -->
<!-- code-review: passed -->
