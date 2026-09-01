## Context

`gaze quality --analyzer <binary>` is currently hard-blocked with an error ("--analyzer is not yet supported for 'gaze quality'"). Even when the external analyzer path is used via `gaze crap` or `gaze report`, the `ExternalSideEffectAnalyzer` only calls `analyze` (or `analyze_stream`) — it never calls `classify_signals`, even when the external analyzer advertises `classify_signals: true` in its `initialize` response. The protocol types, method constant, and capability flag all exist; the conversion code in `convertAnalysisResults` already handles non-nil `Classification` fields. Only the call site and merge logic are missing.

Because unclassified effects are treated conservatively as contractual by `ComputeContractCoverage` (internal/quality/coverage.go lines 47-48), every detected side effect inflates the contract coverage denominator, producing deflated coverage scores and excessive gap counts.

## Goals / Non-Goals

### Goals
- Call `classify_signals` protocol method when the external analyzer advertises `classify_signals: true`
- Merge returned signals through `classify.ComputeScore` to produce per-effect classifications
- Attach classifications to cached `taxonomy.SideEffect` objects before downstream consumers access them
- Remove the `--analyzer is not yet supported` guard from `gaze quality`
- Gracefully degrade when classify_signals fails (warn to stderr, continue with unclassified effects)

### Non-Goals
- Modifying `ComputeContractCoverage` behavior for unclassified effects (the conservative-as-contractual default is correct for analyzers without classification support)
- Adding a new `gaze classify` subcommand for external analyzers
- Streaming classify_signals responses (batch-only for now; streaming can be added later if needed)
- Passing `classify_signals` results through the `gaze analyze` or `gaze report` classify step (those paths already have Go-native classification via `classify.Classify`)

## Decisions

### D1: Classify signals as a post-load enrichment step in ExternalSideEffectAnalyzer

Classification happens inside `loadBatch()` and `loadStreaming()` after `a.cached` is populated, rather than in a separate provider or in Session.Initialize. This keeps the classify_signals call co-located with the analyze call it depends on.

**Rationale**: The `ExternalSideEffectAnalyzer` already owns the protocol client, capabilities, root directory, and patterns. Adding a `classifyAndMerge()` method that runs after `convertAnalysisResults` keeps the mutation of `a.cached` in one place. The alternative — adding a standalone `ClassifyProvider` to Session.Providers — would require threading a new dependency through callers and break the existing composition pattern where `ExternalContractCoverageProvider` uses `ExternalSideEffectAnalyzer.AllResults()` to get effects that are already classified.

### D2: Feed signals through classify.ComputeScore, not direct label assignment

The `classify_signals` protocol returns `ClassifySignalData` records (Source, Weight, Reasoning) that map 1:1 to `taxonomy.Signal`. These are fed into `classify.ComputeScore` to produce a `taxonomy.Classification` (label + confidence) using gaze's scoring engine (tier boost, threshold logic).

**Rationale**: `ComputeScore` encapsulates the label-assignment logic (tier boosts, contractual/ambiguous/incidental thresholds from `.gaze.yaml`). Using it ensures external analyzer classifications follow the same rules as Go-native classifications. The alternative — having the external analyzer return final labels — would bypass gaze's scoring engine and create behavioral divergence between Go-native and external paths. The protocol type names (`ClassifySignalData`, not `ClassifyResultData`) confirm this design intent.

### D3: Graceful degradation on classify_signals failure

If `classify_signals` fails (transport error, protocol error, unmarshal error), the adapter warns to stderr and continues with unclassified effects. This follows the same pattern used by `fetchTestMappings` in `contract.go`.

**Rationale**: Classification is an accuracy improvement, not a correctness prerequisite. Unclassified effects are still usable (treated conservatively as contractual). A hard error would break the entire pipeline for a non-essential enrichment step. Constitution Principle II (Minimal Assumptions) requires that optional capabilities degrade gracefully. On failure, `fetchClassifySignals` returns `(nil, nil)` — nil signals and nil error — so callers do not need to distinguish between "no signals available" and "classify_signals failed."

### D4: New file `internal/adapter/classify.go` for classify-signals logic

The `fetchClassifySignals` function and `mergeClassifications` helper live in a new file, following the adapter package's existing file-per-concern pattern (`call.go`, `contract.go`, `session.go`, `sideeffect.go`).

**Rationale**: The classify logic is a distinct concern from side effect analysis (fetching) and contract coverage (computing). Keeping it in `sideeffect.go` would add unrelated imports (`classify`, `config`) and increase file complexity. The new file also gets its own `_test.go` for isolated testing.

### D5: Config threading for ComputeScore

`classify.ComputeScore` requires a `*config.GazeConfig` parameter (for classification thresholds). The `ExternalSideEffectAnalyzer` does not currently hold a config reference. A config field will be added to the struct, passed through `NewExternalSideEffectAnalyzer` and populated from `Session.Initialize`.

**Rationale**: `ComputeScore` uses config for the contractual/incidental confidence thresholds. Without config, we'd need to either hardcode defaults (fragile) or pass nil (uses `config.DefaultConfig()` internally via `classifyLabel`). Since the session already constructs the analyzer, adding config there is natural. The `*config.GazeConfig` loaded by `initExternalSession` via `config.LoadFromDir` is passed to `NewSession` (or `Session.Initialize`) and threaded to `NewExternalSideEffectAnalyzer` — no separate config load is needed in the adapter layer.

### D6: Signal grouping by function and effect type

`classify_signals` returns a flat list of `ClassifySignalData` entries, each tagged with `(Function, Package, SideEffectType)`. The `mergeClassifications` function groups these by `(Function, Package, SideEffectType)` and matches them to cached `taxonomy.SideEffect` objects. Only effects that are currently unclassified (nil Classification) receive the computed classification — pre-classified effects from the `analyze` response are preserved. When a function has multiple effects of the same `SideEffectType`, the computed classification is applied to all matching effects that have nil Classification. Signals for `(Package, Function, SideEffectType)` tuples with no corresponding cached effect are silently discarded.

**Rationale**: The `analyze` response may include inline classifications (handled at sideeffect.go:211-216). If the analyzer returns both inline classifications and classify_signals data, the inline classification takes precedence as it is more specific (the analyzer explicitly chose to classify that effect). This avoids double-classification and respects the analyzer's choice.

### D7: Remove quality --analyzer guard and wire external providers

The `--analyzer is not yet supported` error in `runQuality()` (cmd/gaze/main.go:1104-1107) is removed. When `--analyzer` is specified for `gaze quality`, external providers from `Session.Initialize()` are used instead of Go-native providers. This follows the same pattern already used in `runCrap()`.

**Rationale**: With classify_signals wired, external analyzers can now provide meaningful quality assessment data. The guard was a placeholder from the initial external analyzer protocol implementation (spec 019). Removing it enables `gaze quality --analyzer snake-eyes` and completes the external analyzer integration story. Closes #246.

## Coverage Strategy

Unit tests for `fetchClassifySignals` and `mergeClassifications` in `classify_test.go` (10+ cases covering success, error, edge, and no-op paths). Integration test for end-to-end `Analyze()` → classified effects flow via fake analyzer in `sideeffect_test.go`. Target: 100% branch coverage on new `classify.go` functions. No e2e tests needed — the fake analyzer provides a controlled test environment sufficient for verifying the full protocol flow.

## Risks / Trade-offs

### R1: Config availability in adapter layer
Adding `*config.GazeConfig` to `ExternalSideEffectAnalyzer` introduces a dependency from the adapter layer to the config package. This is acceptable because the adapter already imports `protocol` and `taxonomy`, and config is a lightweight data package with no circular dependency risk.

### R2: classify_signals timeout
Using the same `protocol.AnalysisTimeout` as `analyze` may be too generous or too tight for classification. Classification typically involves less computation than full analysis (it operates on already-detected effects, not raw source code). If this becomes an issue, a separate timeout constant can be added later.

### R3: Unclassified effects in mixed scenarios
When an analyzer returns inline classifications for some effects (via the `analyze` response) but not others, and also supports `classify_signals`, the unclassified effects will be enriched by `classify_signals` data while the pre-classified ones are preserved. This is the correct behavior but may surprise users who expect uniform classification provenance. The risk is low — analyzers are unlikely to partially classify effects inline.

### R4: Fake analyzer test data fidelity
The fake analyzer's `classify_signals` handler already returns one signal (`divide`/`ErrorReturn`, weight 15, source `docstring`), but `classify_signals` capability is `false`. Enabling it and adding additional signals to produce both Contractual and Incidental outcomes requires careful weight selection (P0 effects get +25 tier boost, P1 get +10). Integration testing with a real analyzer (e.g., snake-eyes) is recommended after implementation.

### R5: Security model inherited from existing protocol
The `classify_signals` call uses the same subprocess communication pattern as `analyze` and other protocol methods. The external binary is trusted at the same level as the existing protocol — it receives the project root path and returns free-text fields (`Reasoning`). No additional attack surface is introduced beyond what the existing protocol already exposes.
<!-- scaffolded by uf vdev -->
