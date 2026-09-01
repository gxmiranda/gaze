## ADDED Requirements

### Requirement: classify_signals protocol call

The `ExternalSideEffectAnalyzer` MUST call the `classify_signals` JSON-RPC method when the external analyzer's `Capabilities.ClassifySignals` is `true`. The call MUST occur after `analyze` (or `analyze_stream`) results are cached and before any downstream consumer accesses the results.

#### Scenario: Analyzer supports classify_signals
- **GIVEN** an external analyzer that returns `classify_signals: true` in its `initialize` response
- **WHEN** `ExternalSideEffectAnalyzer.Analyze()` is called for the first time (triggering `loadBatch()` or `loadStreaming()`)
- **THEN** the adapter MUST call `classify_signals` with the same `RootPath` and `Patterns` used for `analyze`
- **AND** the returned signals MUST be merged into the cached `taxonomy.SideEffect` objects

#### Scenario: Analyzer does not support classify_signals
- **GIVEN** an external analyzer that returns `classify_signals: false` in its `initialize` response
- **WHEN** `ExternalSideEffectAnalyzer.Analyze()` is called for the first time (triggering `loadBatch()` or `loadStreaming()`)
- **THEN** the adapter MUST NOT call `classify_signals`
- **AND** all cached effects MUST retain their existing classification state (nil or inline)

### Requirement: Signal-to-classification conversion

Returned `ClassifySignalData` records MUST be converted to `taxonomy.Signal` structs and passed through `classify.ComputeScore` to produce a `taxonomy.Classification` for each side effect. The adapter MUST NOT assign classification labels directly.

#### Scenario: Signals scored through ComputeScore
- **GIVEN** `classify_signals` returns signals for a function's `ContainerMutation` effect with Source="type_annotation", Weight=30
- **WHEN** the adapter processes the response
- **THEN** the signals MUST be grouped by `(Package, Function, SideEffectType)`
- **AND** each group MUST be passed to `classify.ComputeScore` with the matching `SideEffectType` and config
- **AND** the resulting `taxonomy.Classification` MUST be attached to the corresponding `taxonomy.SideEffect`
- **AND** the classification MUST have label `contractual` with confidence 90 (base 50 + P1 tier boost 10 + weight 30)

#### Scenario: Signals for unmatched SideEffectType
- **GIVEN** `classify_signals` returns signals for a `(Package, Function, SideEffectType)` tuple that has no corresponding cached effect
- **WHEN** the adapter merges classifications
- **THEN** the unmatched signals MUST be silently discarded
- **AND** no cached effects MUST be modified

#### Scenario: Multiple effects of same type on same function
- **GIVEN** a function has two `ContainerMutation` effects (on different targets) with nil Classification
- **AND** `classify_signals` returns signals for that function's `ContainerMutation`
- **WHEN** the adapter merges classifications
- **THEN** the computed classification MUST be applied to all matching effects with nil Classification

### Requirement: Pre-classified effect preservation

Effects that already have a non-nil `Classification` from the `analyze` response MUST NOT be overwritten by `classify_signals` results.

#### Scenario: Inline classification takes precedence
- **GIVEN** the `analyze` response includes a `ContainerMutation` effect with an inline classification of `contractual` at confidence 85
- **AND** `classify_signals` returns signals for the same effect
- **WHEN** the adapter merges classifications
- **THEN** the inline classification (contractual, 85) MUST be preserved
- **AND** the classify_signals data for that effect MUST be ignored

### Requirement: Graceful degradation on classify_signals failure

If the `classify_signals` protocol call fails (transport error, protocol error, or unmarshal error), the adapter MUST warn to stderr and continue with unclassified effects. The failure MUST NOT propagate as a hard error.

#### Scenario: classify_signals failure (transport, protocol, or unmarshal error)
- **GIVEN** an external analyzer that advertises `classify_signals: true`
- **WHEN** the `classify_signals` call fails with a transport error, protocol error, or unmarshal error
- **THEN** the adapter MUST emit a warning to stderr in the format `"warning: classify_signals failed: %v"`
- **AND** `fetchClassifySignals` MUST return `(nil, nil)` — nil signals and nil error
- **AND** all cached effects MUST retain their existing classification state
- **AND** `Analyze()` and `AllResults()` MUST return successfully with unclassified effects

#### Scenario: classify_signals with empty/nil signals response
- **GIVEN** an external analyzer that advertises `classify_signals: true`
- **WHEN** `classify_signals` returns an empty signals list
- **THEN** `mergeClassifications` MUST be a no-op — no cached effects are modified

### Requirement: Config threading for ComputeScore

The `ExternalSideEffectAnalyzer` MUST accept a `*config.GazeConfig` parameter for use with `classify.ComputeScore`. When nil, `config.DefaultConfig()` MUST be used.

#### Scenario: Custom config thresholds
- **GIVEN** a `.gaze.yaml` with custom classification thresholds
- **WHEN** the adapter runs `ComputeScore` with signals from `classify_signals`
- **THEN** the config thresholds MUST be applied (not hardcoded defaults)

## MODIFIED Requirements

### Requirement: gaze quality --analyzer support

The `runQuality()` function in `cmd/gaze/main.go` MUST support the `--analyzer` flag. When `--analyzer` is specified, external providers from `Session.Initialize()` MUST be used instead of Go-native providers. The existing hard-block error ("--analyzer is not yet supported for 'gaze quality'") MUST be removed.

Previously: `runQuality()` returned an error when `--analyzer` was specified (lines 1104-1107).

#### Scenario: External quality assessment
- **GIVEN** `gaze quality --analyzer snake-eyes ./...`
- **WHEN** the command runs
- **THEN** it MUST initialize an external analyzer session
- **AND** use external providers for side effect analysis and contract coverage
- **AND** produce quality reports with classified effects (when classify_signals is supported)

#### Scenario: External quality without classify_signals
- **GIVEN** `gaze quality --analyzer <binary>` where the binary has `classify_signals: false`
- **WHEN** the command runs
- **THEN** it MUST still produce quality reports
- **AND** all effects MUST be treated as contractual (conservative default, existing behavior)

### Requirement: Fake analyzer classify_signals capability

The fake analyzer at `internal/protocol/testdata/fake_analyzer/main.go` MUST advertise `classify_signals: true` in its `initialize` response and return realistic classification signals when the `classify_signals` method is called.

Previously: The fake analyzer had `classify_signals: false` and the handler returned empty results.

#### Scenario: Fake analyzer classification data
- **GIVEN** the fake analyzer receives a `classify_signals` request
- **WHEN** it processes the request
- **THEN** it MUST return `ClassifySignalData` entries with realistic Source, Weight, and SideEffectType values
- **AND** the signals MUST cover at least one Contractual and one Incidental classification outcome

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
