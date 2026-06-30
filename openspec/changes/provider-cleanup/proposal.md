## Why

Phases 1 and 2 of Issue #95 introduced provider interfaces and the external analyzer protocol. During those phases, several backward-compatibility shims and code duplications were introduced intentionally to keep the changes safe and incremental:

1. **Deprecated `ContractCoverageFunc` and `SSADegradedPackages` fields** in `crap.Options` — kept for backward compatibility during Phase 1 transition. All production callers now use `ContractCoverageProvider`. The deprecated fields add confusion and a fallback code path in `crap.Analyze` that is no longer exercised in production.

2. **Duplicated `isMainPkg`** — `internal/provider/goprovider/sideeffect.go` has a local copy despite `loader.IsMainPkg` being the canonical exported version (consolidated by the `analyze-multi-package` change).

3. **`BuildContractCoverageFunc` in `internal/crap/`** — this Go-specific function imports `analysis`, `classify`, `quality`, `loader`, and `go/packages`, keeping the `crap` package coupled to Go-specific tooling. `GoContractCoverageProvider` already wraps it, so it can move to `internal/provider/goprovider/` where the Go-specific code belongs.

This change removes the shims and deduplicates the code, completing the provider interface cleanup.

## What Changes

### Deprecated Field Removal

Remove `ContractCoverageFunc` and `SSADegradedPackages` from `crap.Options`. Remove the fallback path in `crap.Analyze` that checks these fields. Update all remaining callers (internal tests, mock precedence test) to use `ContractCoverageProvider` exclusively.

### isMainPkg Deduplication

Replace the local `isMainPkg` in `goprovider/sideeffect.go` with a call to `loader.IsMainPkg`.

### BuildContractCoverageFunc Relocation

Move `BuildContractCoverageFunc` and its helpers (`analyzePackageCoverage`, `loadTestPackage`, `classifyResults`, `loadGazeConfigBestEffort`) from `internal/crap/contract.go` to `internal/provider/goprovider/contract.go`. The `GoContractCoverageProvider.Build()` method calls these directly instead of through the `crap` package. This removes Go-specific imports from `internal/crap/`.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `crap.Options`: `ContractCoverageFunc` and `SSADegradedPackages` fields removed
- `internal/crap/`: Go-specific imports (`analysis`, `classify`, `quality`, `loader`, `go/packages`) removed from package

### Removed Capabilities
- `ContractCoverageFunc` callback pattern (replaced by `ContractCoverageProvider` interface)
- `SSADegradedPackages` direct field (degradation data now flows through `ContractCoverageProvider.Build` return value)

## Impact

- **Modified files**: `internal/crap/analyze.go` (remove deprecated fields + fallback), `internal/crap/crap.go` (remove fields from Options), `internal/provider/goprovider/sideeffect.go` (use loader.IsMainPkg), `internal/provider/goprovider/contract.go` (absorb BuildContractCoverageFunc)
- **Removed file**: `internal/crap/contract.go` (moved to goprovider)
- **Test updates**: `internal/crap/analyze_internal_test.go`, `internal/crap/crap_test.go`, `internal/provider/mockprovider/mock_test.go` (precedence test removed or adapted)
- **Breaking change to internal API**: `crap.Options` fields removed. All callers are internal (`cmd/gaze/`, `internal/aireport/`, `internal/provider/`). No external consumers.
- **Zero user-facing changes**: CLI behavior, output formats, and exit codes are identical.

## Constitution Alignment

### I. Accuracy
**Assessment**: PASS — Pure cleanup. No scoring, detection, or classification logic changes. All tests pass.

### II. Minimal Assumptions
**Assessment**: PASS — No user-facing changes.

### III. Actionable Output
**Assessment**: PASS — Output unchanged.

### IV. Testability
**Assessment**: PASS — Removing deprecated code paths simplifies the test surface. The mock precedence test (which tested the deprecated fallback) is removed since the fallback no longer exists.