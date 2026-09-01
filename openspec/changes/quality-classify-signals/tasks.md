<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Add config field to ExternalSideEffectAnalyzer

- [x] 1.1 Add `config *config.GazeConfig` field to `ExternalSideEffectAnalyzer` struct in `internal/adapter/sideeffect.go`. Update `NewExternalSideEffectAnalyzer` to accept a `*config.GazeConfig` parameter and store it. When nil, the field is stored as nil (callers handle default).
- [x] 1.2 Update all call sites of `NewExternalSideEffectAnalyzer` in `internal/adapter/session.go` to pass the config parameter. For now, pass nil (config threading from CLI comes in task group 4).

## 2. Implement classify_signals fetch and merge logic

- [x] 2.1 [P] Create `internal/adapter/classify.go` with:
  - `fetchClassifySignals(client *protocol.Client, rootDir string, patterns []string, stderr io.Writer) ([]protocol.ClassifySignalData, error)` — calls `classify_signals` protocol method using `callAndUnmarshal[protocol.ClassifySignalsResult]`, graceful degradation on failure (warn to stderr in format `"warning: classify_signals failed: %v"`, return nil signals and nil error). This is a standalone function (not a method) for testability — it can be tested with a mock client without constructing a full `ExternalSideEffectAnalyzer`.
  - `mergeClassifications(cached []taxonomy.AnalysisResult, signals []protocol.ClassifySignalData, cfg *config.GazeConfig)` — groups signals by `(Package, Function, SideEffectType)`, matches to cached effects, skips effects with non-nil Classification (design D6), applies classification to all matching effects of the same type (not just the first), silently discards signals for unmatched `(Package, Function, SideEffectType)` tuples, calls `classify.ComputeScore` for each group, attaches resulting `taxonomy.Classification` to the effect. Uses `config.DefaultConfig()` when cfg is nil.
- [x] 2.2 [P] Create `internal/adapter/classify_test.go` with tests for:
  - `fetchClassifySignals` success path (fake analyzer returns signals)
  - `fetchClassifySignals` transport error (graceful degradation, warning to stderr)
  - `fetchClassifySignals` protocol error (graceful degradation)
  - `mergeClassifications` with matching signals (effects get classified)
  - `mergeClassifications` with no matching signals (effects unchanged)
  - `mergeClassifications` preserves pre-classified effects (inline from analyze)
  - `mergeClassifications` with nil config (uses DefaultConfig)
  - `mergeClassifications` with empty/nil signals slice (no-op, effects unchanged)
  - `mergeClassifications` with SideEffectType mismatch (signals for types not in cached effects — silently ignored)
  - `mergeClassifications` with multiple effects of same type on same function (all get classified)

## 3. Wire classify_signals into ExternalSideEffectAnalyzer load path

- [x] 3.1 In `internal/adapter/sideeffect.go`, add `classifyAndMerge()` method on `ExternalSideEffectAnalyzer` that: checks `a.caps.ClassifySignals`, calls `fetchClassifySignals`, calls `mergeClassifications` on `a.cached`. Call `classifyAndMerge()` at the end of both `loadBatch()` and `loadStreaming()` after `a.cached` is populated.
- [x] 3.2 Add integration test in `internal/adapter/sideeffect_test.go` (or existing test file) verifying that `Analyze()` returns classified effects when the fake analyzer supports classify_signals. Also verify that calling `Analyze(pkgA)` then `Analyze(pkgB)` returns classified effects for both packages from a single `classify_signals` call (classification is not repeated).

## 4. Remove quality --analyzer guard and wire external providers

- [x] 4.1 In `cmd/gaze/main.go`, remove the `--analyzer is not yet supported for 'gaze quality'` guard from `runQuality()` (lines 1104-1107). Wire external analyzer session initialization and provider usage following the pattern in `runCrap()`: create `Session`, call `Initialize()`, use returned `Providers` for side effect analysis and contract coverage. The `*config.GazeConfig` loaded by `initExternalSession` via `config.LoadFromDir` is passed through — no separate config load needed. Note: `gaze quality --analyzer` produces a reduced report with contract coverage metrics only (no assertion mapping, over-specification, or gap hints — those require Go-native SSA analysis not available from external analyzers).
- [x] 4.2 Update `Session.Initialize()` in `internal/adapter/session.go` to accept or thread a `*config.GazeConfig` parameter for the `ExternalSideEffectAnalyzer`. The config loaded in `initExternalSession` is threaded through `NewSession` → `Session.Initialize()` → `NewExternalSideEffectAnalyzer`. This supersedes the nil-passing from task 1.2. Add test verifying `gaze quality --analyzer` with both `classify_signals: true` (effects get classified) and `classify_signals: false` (effects treated as contractual) scenarios.

## 5. Update fake analyzer

- [x] 5.1 In `internal/protocol/testdata/fake_analyzer/main.go`, set `classify_signals: true` in the capabilities returned by the `initialize` handler. Update the `classify_signals` handler to return realistic `ClassifySignalData` entries — at minimum, signals that produce one Contractual and one Incidental classification outcome when scored by `ComputeScore`. Weight guidance: a P0 effect (e.g., ErrorReturn) with weight 20 gets tier boost +25 → confidence ~95 (contractual). For an Incidental outcome, use a P2 effect with a negative weight (e.g., -15) to push confidence below the incidental threshold. Verify expected outcomes against `classify.ComputeScore` with `config.DefaultConfig()`. Audit existing integration tests that use the fake analyzer for behavioral changes when `classify_signals` flips from `false` to `true` — tests asserting on effect classification state or protocol call counts may need updating.

## 6. Verification

- [x] 6.1 Run `go test -race -count=1 -short ./internal/adapter/...` — all new and existing tests pass.
- [x] 6.2 Run `go test -race -count=1 -short ./cmd/gaze/...` — verify quality command accepts --analyzer flag.
- [x] 6.3 Run `go test -race -count=1 -short ./...` — full test suite passes with no regressions.
- [x] 6.4 Run `golangci-lint run` — no lint errors.
- [x] 6.5 Verify constitution alignment: Accuracy (fixes systematic accuracy error from unclassified effects), Minimal Assumptions (classify_signals is capability-gated, no mandatory dependencies), Actionable Output (accurate coverage metrics guide users to real gaps), Testability (all new code testable in isolation via fake analyzer and synthetic inputs, 100% branch coverage target on classify.go).
<!-- spec-review: passed -->
<!-- scaffolded by uf vdev -->
