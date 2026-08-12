## Context

The classifier in `internal/classify/` uses five signal analyzers to determine whether a side effect is contractual, incidental, or ambiguous. Two of these — naming (`naming.go`) and godoc (`godoc.go`) — have an asymmetry in how they apply signals:

- **Contractual signals** use typed structs with `impliesFor []taxonomy.SideEffectType`, so a `Get` prefix only boosts `ReturnValue`, not `LogWrite`.
- **Incidental signals** use plain `[]string` with no type guard, so a `log` prefix penalizes every effect type equally — including `ReturnValue`.

This asymmetry causes P0 return values on functions with I/O-describing names (e.g., `LogAndCompute`) to be misclassified as incidental (confidence 38%, well below the 50 incidental threshold), and the contradiction penalty (-20) amplifies the problem when both tier boost and naming signals conflict.

## Goals / Non-Goals

### Goals

- Add `appliesTo []taxonomy.SideEffectType` guards to incidental naming prefixes and godoc keywords, scoped to I/O effect types only.
- Mirror the existing `contractualPrefixes`/`contractualKeywords` typed-struct pattern for consistency.
- Ensure P0 effects (`ReturnValue`, `ErrorReturn`, `SentinelError`, `ReceiverMutation`, `PointerArgMutation`) are never penalized by I/O-describing naming/godoc signals.
- Add regression tests that verify the fix and prevent re-introduction of the bug.

### Non-Goals

- Redesigning the scoring formula or contradiction penalty logic — the current formula is correct when signals are properly scoped.
- Changing the incidental penalty weights (-10 naming, -15 godoc) — the weights are appropriate for I/O effects.
- Adding new incidental prefixes or keywords — scope limited to fixing the type guard.
- Changing classification behavior for effects that should be incidental (e.g., `LogWrite` on a function named `LogRequest`).

## Decisions

### D1: Reuse the existing typed-struct pattern

The `contractualPrefixes` slice already uses `struct { prefix string; impliesFor []taxonomy.SideEffectType }`. The incidental prefixes will adopt an identical structure with an `appliesTo` field (named differently from `impliesFor` for semantic clarity — "applies to" means "this penalty targets these types", while "implies for" means "this boost elevates these types").

**Rationale**: Consistent patterns reduce cognitive load. The existing contractual pattern already handles the `nil` case (meaning "all types"), which we won't need for incidental signals but is there for future use.

### D2: Scope incidental signals to I/O effect types only

The `appliesTo` list for all incidental prefixes (`log`, `debug`, `trace`, `print`) and keywords (`logs`, `prints`, `traces`, `debugs`) will be:

```go
[]taxonomy.SideEffectType{
    taxonomy.LogWrite,
    taxonomy.StdoutWrite,
    taxonomy.StderrWrite,
}
```

**Rationale**: These prefixes/keywords describe I/O behavior. A function named `LogAndCompute` logs something (incidental) **and** returns a value (contractual). The naming signal should only affect the I/O effect, not the return value. The three I/O types listed are the exhaustive set of output-related effects (LogWrite is P2, StdoutWrite and StderrWrite are P3) that naming heuristics semantically describe.

### D3: Keep the same short-circuit ordering

Both `AnalyzeNamingSignal` and `AnalyzeGodocSignal` currently check incidental signals first, then contractual. This ordering is preserved — the only change is that the incidental check now includes an effect-type match condition. If the effect type doesn't match, the incidental entry is skipped and execution falls through to contractual checks.

**Rationale**: The ordering was originally designed so that a function named `LogResult` gets an incidental signal for `LogWrite` effects even if `Log` were a contractual prefix (which it isn't currently). With the type guard, the ordering matters less, but preserving it avoids unnecessary behavioral changes.

## Risks / Trade-offs

### Risk: New I/O effect types added without updating the guard

If a future change adds a new I/O-related `SideEffectType` (e.g., a universal type from the external analyzer protocol), the `appliesTo` lists in naming.go and godoc.go would need to be updated. This is mitigated by the fact that `SideEffectType` additions are rare and always go through a spec process.

**Mitigation**: A comment above the `appliesTo` lists will reference issue #105 and note that new I/O types should be added here.

### Trade-off: Per-prefix type lists vs. shared constant

Each incidental prefix entry will have its own `appliesTo` slice. An alternative would be a package-level `incidentalIOTypes` constant shared across entries. We chose per-entry lists for consistency with the contractual pattern, where each prefix has its own `impliesFor` list (e.g., `Get` implies `ReturnValue` only, while `Save` implies three types).

**Accepted**: The slight duplication is worth the flexibility — if a future prefix like `emit` should apply to `ChannelSend` instead of `LogWrite`, per-entry lists make that trivial.

## Coverage Strategy

Unit tests for each modified function (`AnalyzeNamingSignal`, `AnalyzeGodocSignal`) using table-driven tests covering both matching and non-matching effect types (tasks 2.1, 2.2). Integration test through `ComputeScore` for the full issue #105 scenario to verify that the contradiction penalty no longer fires when incidental signals are correctly scoped (task 3.1). Existing tests (`TestNamingSignal_IncidentalPrefixes`, `TestAnalyzeGodocSignal_IncidentalOnly`, `TestAnalyzeGodocSignal_IncidentalPriority`) must be updated to use I/O effect types instead of `ReturnValue`. No e2e test needed — the fix is internal to the classify package.
