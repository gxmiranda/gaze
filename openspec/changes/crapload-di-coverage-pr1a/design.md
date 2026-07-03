## Context

Six functions in `internal/crap/` and `internal/aireport/` have 0% line coverage
and CRAP scores between 30 and 110, contributing to CRAPload sitting at exactly
38/38 (the CI threshold). These functions are orchestration functions that
hardcode calls to heavy I/O operations (`packages.Load`,
`analysis.LoadAndAnalyze`, `quality.Assess`).

`gaze crap` generates coverage profiles via `go test -short`, so tests guarded
by `testing.Short()` are skipped and contribute zero coverage. To reduce
CRAPload, new tests must run without the `-short` guard, which requires
dependency injection so tests can substitute synthetic implementations for
the expensive I/O calls.

The existing `pipelineStepFuncs` pattern in `internal/aireport/runner.go:243`
proves this approach works in the same codebase.

## Goals / Non-Goals

### Goals
- Make 6 zero-coverage functions testable in isolation (Principle IV: Testability)
- Add unit tests that run without `-short` guard and produce coverage data
- Reduce CRAPload from 38 toward ~32 (drop ~6 functions below CRAP threshold)
- Follow the established `pipelineStepFuncs` DI pattern
- Zero behavioral change to production code paths

### Non-Goals
- Decomposing `BuildContractCoverageFunc` (deferred to PR 2a)
- Deduplicating `loadTestPackage` / `loadTestPackageForQuality` (deferred to PR 2a)
- Modifying any exported API surface
- Changing the CRAP threshold value (deferred to Phase 3)
- Achieving 100% coverage on target functions (50%+ is sufficient to drop below CRAP 15)

## Decisions

### D1: Deps struct pattern with nil-means-default

Each orchestration function that needs DI will accept an optional deps struct
parameter. When a field is nil, the real implementation is used. This mirrors the
existing `pipelineStepFuncs` pattern exactly.

**Rationale**: Proven pattern in the same package. Zero-value struct means
production behavior — no risk of accidentally using a test double in production.
Callers that don't need injection pass zero-value or omit the parameter.

### D2: Two deps structs, not four

The 4 orchestration functions form a call hierarchy:

```
runQualityStep → runQualityForPackage → (leaf calls)
runClassifyStep → (leaf calls)
```

Rather than 4 separate deps structs, use 2:

- **`qualityPipelineDeps`** in `internal/aireport/runner_steps.go` — shared by
  `runQualityStep`, `runQualityForPackage`, and `runClassifyStep`. Contains
  injectable fields for `resolvePackagePaths`, `loadAndAnalyze`,
  `runClassifyResults`, `loadTestPkg`, `qualityAssess`, `resolveModulePkgs`,
  and `loadConfig`.

- **`contractCoverageDeps`** in `internal/provider/goprovider/contract.go` — used by
  `analyzePackageCoverage`. Contains injectable fields for `loadAndAnalyze`,
  `classifyResults`, `loadTestPkg`, and `qualityAssess`.

**Rationale**: Fewer types to maintain. The aireport functions share the same
dependency set because they're all part of the same pipeline. The crap package
has its own copy of some dependencies (e.g., `classifyResults` differs from
`runClassifyResults`) so it needs a separate struct.

### D3: Variadic deps parameter for backward compatibility

Functions will accept deps as a variadic parameter:

```go
func analyzePackageCoverage(..., deps ...contractCoverageDeps) (...)
```

When no deps argument is supplied, the function uses its default (real)
implementations. This preserves all existing call sites without modification.

**Rationale**: Minimizes diff size and risk. Production callers don't change at
all. Only test callers pass the deps struct.

### D4: Leaf functions tested with real fixtures, no DI

`loadTestPackage` and `loadTestPackageForQuality` are leaf functions that call
`packages.Load` directly. There's nothing meaningful to inject — the function
*is* the `packages.Load` call plus error handling and test-syntax detection.

These will be tested with small `testdata/` fixture packages that load in
sub-second time:
- `internal/quality/testdata/src/welltested` — success path (has test files)
- `internal/analysis/testdata/src/returns` — error path (no test files)

**Rationale**: The fixtures already exist, load quickly, and have no external
imports that would require `go mod download` state. Testing with real packages
exercises the actual `packages.Config` setup, which is the main thing to verify.

### D5: Internal package tests (same package, not `_test` suffix)

New tests will be in the same package (internal tests) to access unexported
functions and deps structs. This matches the existing test file patterns:
- `contract_test.go` uses `package crap`
- `pipeline_internal_test.go` uses `package aireport`

**Rationale**: The target functions are all unexported. External package tests
cannot access them.

## Risks / Trade-offs

### Risk: Deps structs increase API surface within the package

Adding deps structs introduces new unexported types. If poorly designed, they
could become a maintenance burden as function signatures evolve.

**Mitigation**: Keep deps structs minimal — only inject the calls that are
actually expensive I/O operations. Pure functions (`classify.CountLabels`,
`quality.BuildPackageSummary`) are not injected.

### Risk: Leaf function tests may be slow on some CI environments

`loadTestPackage` tests call `packages.Load`, which invokes the Go compiler
toolchain. On slow CI runners, even small fixtures could take several seconds.

**Mitigation**: The existing `testdata/` fixtures are minimal Go packages (4-10
source files, no external imports). If timing becomes an issue, these specific
tests can be `-short`-guarded as a fallback without losing the DI-based coverage
gains on the orchestration functions.

### Trade-off: Two similar deps patterns in two packages

`contractCoverageDeps` (crap) and `qualityPipelineDeps` (aireport) share some
field types (e.g., `loadAndAnalyze`, `loadTestPkg`). Deduplication would require
a shared package for the deps type, which introduces coupling between `crap` and
`aireport`. We accept the duplication as the lesser cost.

### Trade-off: Coverage may not reach 50% on all functions

Some functions have high cyclomatic complexity with many branches (e.g.,
`analyzePackageCoverage` at complexity 10). Achieving 50% coverage requires
exercising at least 5 distinct paths. The test plan targets the most impactful
paths (success, analysis failure, no results, classify failure, no test files,
quality failure, SSA degradation) which should exceed 50%.
