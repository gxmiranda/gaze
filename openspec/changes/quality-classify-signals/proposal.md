## Why

When `gaze quality --analyzer <binary>` runs against an external analyzer that advertises `classify_signals: true` in its `initialize` response, the `classify_signals` protocol method is never called. The `ExternalSideEffectAnalyzer` only calls `analyze` (or `analyze_stream`), leaving all returned side effects with a nil `Classification` field.

`ComputeContractCoverage` in `internal/quality/coverage.go` treats unclassified effects conservatively as contractual (lines 47-48). This means every detected side effect — including effects that should be Incidental (like logging, debug output, internal goroutine lifecycle) — counts toward the contract coverage denominator. The result is inflated gap counts and deflated coverage scores that misrepresent the actual test quality.

**Observed impact** (snake-eyes analyzer on a real project): 326 tests analyzed, only 28% average contract coverage. 729 CallbackInvocation gaps and 277 ContainerMutation gaps dominate — effects that would likely classify as Incidental with proper classification. With classification, coverage is expected to rise to 60-70%.

This is a correctness bug. The protocol method exists, the types exist, the capability flag exists, the conversion code already handles non-nil classifications — only the call site is missing.

Closes #246.

## What Changes

Wire the `classify_signals` JSON-RPC protocol method into the `ExternalSideEffectAnalyzer` so that when an external analyzer advertises `classify_signals: true`, the method is called after `analyze` (or `analyze_stream`) and the returned classification labels are attached to the cached `taxonomy.SideEffect` objects before downstream consumers (contract coverage, quality reports) access them.

## Capabilities

### New Capabilities
- `classify_signals integration`: When an external analyzer advertises `classify_signals: true`, gaze automatically calls the protocol method and attaches classification labels (Contractual/Incidental/Ambiguous) to detected side effects before computing contract coverage.

### Modified Capabilities
- `ExternalSideEffectAnalyzer`: After calling `analyze`/`analyze_stream`, conditionally calls `classify_signals` and merges classification data into cached results. Graceful degradation on error (warn to stderr, continue with unclassified effects).
- `gaze quality --analyzer`: Removes the hard-block error ("--analyzer is not yet supported for 'gaze quality'") and enables external analyzer usage for quality assessment.
- `gaze report --analyzer`: Quality step in the report pipeline benefits from classified effects, producing accurate contract coverage metrics.

### Removed Capabilities
- None.

## Impact

### Files Modified
- `internal/adapter/sideeffect.go` — Add `fetchClassifySignals()` method and call it from `loadBatch()`/`loadStreaming()` after analysis results are cached
- `internal/adapter/session.go` — Accept `*config.GazeConfig` parameter in `NewSession`/`Session.Initialize()` and pass it through to `NewExternalSideEffectAnalyzer`
- `cmd/gaze/main.go` — Remove the `--analyzer is not yet supported` guard in `runQuality()`; wire external providers into quality flow
- `internal/protocol/testdata/fake_analyzer/main.go` — Enable `classify_signals: true` in capabilities and return realistic classification data

### Files Added
- `internal/adapter/classify.go` — New file for `fetchClassifySignals()` and `mergeClassifications()` helpers (following the `contract.go` / `call.go` file-per-concern pattern)
- `internal/adapter/classify_test.go` — Tests for classification fetch and merge logic

### Downstream Benefits
- `ComputeContractCoverage` correctly excludes Incidental effects from the contractual denominator
- `deriveCoverageReason` returns accurate reasons (no longer "all_effects_unclassified" for analyzers that support classification)
- `gaze report` JSON output includes classification breakdown (Contractual/Incidental/Ambiguous counts) when using external analyzers
- `gaze quality` text and JSON output accurately reflects contract coverage with classified effects

## Constitution Alignment

Assessed against the Gaze project constitution (`.specify/memory/constitution.md` v1.3.0).

### I. Accuracy

**Assessment**: PASS

This change fixes a systematic accuracy error where unclassified side effects from external analyzers are treated conservatively as contractual, inflating gap counts and deflating coverage scores. With classification, Incidental effects (logging, debug output, internal goroutine lifecycle) are correctly excluded from the contractual denominator, producing accurate contract coverage metrics. False positives in gap reporting erode trust (Constitution Principle I) — this change eliminates them for classification-capable analyzers.

### II. Minimal Assumptions

**Assessment**: PASS

Classification is capability-gated (`classify_signals: true` in the initialize handshake). Analyzers that don't support classification continue to work exactly as before — effects remain unclassified and are treated conservatively as contractual. No mandatory dependency, annotation, or restructuring is introduced.

### III. Actionable Output

**Assessment**: PASS

Classification labels are machine-parseable (JSON output includes classification per effect). Accurate gap counts and coverage scores guide users to real test gaps rather than false positives from Incidental effects. Metrics become comparable across runs and between Go-native and external analyzer paths.

### IV. Testability

**Assessment**: PASS

The `fetchClassifySignals` helper is independently testable via the existing fake analyzer binary. The `mergeClassifications` function is a pure data transformation testable with synthetic inputs. The fake analyzer's `classify_signals` handler already exists (just needs capability enabled and realistic responses). Integration tests verify the end-to-end flow through the provider interface. Coverage strategy: unit tests for new functions (100% branch coverage target), integration test via fake analyzer.
<!-- scaffolded by uf vdev -->
