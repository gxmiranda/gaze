## Why

The CI CRAPload threshold is `--max-crapload=38` and `main` is at exactly
`CRAPload: 38/38`. Six functions in `internal/crap/` and `internal/aireport/`
have 0% line coverage and CRAP scores between 30 and 110, contributing to this
fragile boundary. Any complexity increase to these functions risks breaking CI,
as already happened in #131 / PR #160.

These functions are currently untestable in isolation because they hardcode calls
to heavy I/O operations (`packages.Load`, `analysis.LoadAndAnalyze`,
`quality.Assess`). The only existing tests are guarded by `testing.Short()`,
which means they are **skipped** during `gaze crap`'s internal coverage
generation (`go test -short -coverprofile=...`) and contribute zero coverage
to the CRAPload calculation.

Adding dependency injection to these functions enables fast unit tests that run
without the `-short` guard, producing real coverage data that reduces CRAP scores
and creates headroom below the threshold.

This is Phase 1a of issue #166 (CRAPload fragility reduction).

## What Changes

### Production code (light refactoring)

Add injectable dependency structs to 4 orchestration functions, following the
existing `pipelineStepFuncs` pattern in `internal/aireport/runner.go`:

- `analyzePackageCoverage` in `internal/crap/contract.go`
- `runQualityForPackage` in `internal/aireport/runner_steps.go`
- `runQualityStep` in `internal/aireport/runner_steps.go`
- `runClassifyStep` in `internal/aireport/runner_steps.go`

Each function gains an optional deps parameter (or uses a struct with default
constructors) so that tests can substitute synthetic implementations for heavy
I/O calls while production call sites remain unchanged.

### Test code

Add unit tests for all 6 target functions:

- `loadTestPackage` (crap) -- test with small `testdata/` fixtures, no `-short` guard
- `loadTestPackageForQuality` (aireport) -- same approach
- `analyzePackageCoverage` (crap) -- test via injected deps
- `runQualityForPackage` (aireport) -- test via injected deps
- `runQualityStep` (aireport) -- test via injected deps
- `runClassifyStep` (aireport) -- test via injected deps

## Capabilities

### New Capabilities
- None (internal testability improvement only)

### Modified Capabilities
- `analyzePackageCoverage`: Accepts optional injectable dependencies for testing
- `runQualityForPackage`: Accepts optional injectable dependencies for testing
- `runQualityStep`: Accepts optional injectable dependencies for testing
- `runClassifyStep`: Accepts optional injectable dependencies for testing

### Removed Capabilities
- None

## Impact

- **Files modified**: `internal/crap/contract.go`, `internal/aireport/runner_steps.go`
- **Test files modified**: `internal/crap/contract_test.go`, `internal/aireport/runner_steps_test.go`
- **No API surface changes**: All functions are unexported; no callers outside their package
- **No behavioral changes**: Production code paths remain identical; DI structs default to the real implementations
- **Expected CRAPload reduction**: ~6 functions drop below the CRAP threshold (15), reducing CRAPload from 38 to ~32

## Constitution Alignment

Assessed against the Gaze project constitution (v1.3.0).

### I. Accuracy

**Assessment**: N/A

This change does not modify side effect detection, classification, or reporting
logic. All production code paths remain identical.

### II. Minimal Assumptions

**Assessment**: PASS

No new assumptions are introduced. The DI pattern uses zero-value defaults that
wire to the real implementations, requiring no configuration from users or
callers.

### III. Actionable Output

**Assessment**: N/A

This change does not modify output formats, report content, or metrics. CRAPload
numbers will change as a consequence of increased coverage, but the metric
definition and computation are unchanged.

### IV. Testability

**Assessment**: PASS

This change directly advances Principle IV. It makes 6 functions testable in
isolation by decoupling them from heavy I/O dependencies. Tests verify observable
side effects (return values, error conditions, degradation signals) rather than
implementation details. Coverage strategy: unit tests with synthetic dependencies,
no `-short` guard, targeting >= 50% line coverage per function.
