## Why

`BuildContractCoverageFunc` in `internal/provider/goprovider/contract.go` has
complexity 22 and CRAP 506 (0% line coverage) — the single worst CRAPload
offender in the codebase. Its complexity comes from interleaving three concerns
in one function: effects discovery, coverage map construction with reason
computation, and the returned closure's fallback logic.

Additionally, two utility functions are duplicated across `goprovider` and
`aireport`:
- `loadTestPackage` / `loadTestPackageForQuality` — near-identical test package
  loaders (the aireport copy is missing `pkg.Errors` validation, a latent bug)
- `loadGazeConfigBestEffort` — identical config-load-with-fallback in both
  packages

This is Phase 2a of issue #166 (CRAPload fragility reduction). Phases 1a (#183)
and 1b (#187) reduced CRAPload from 38 to 30. This phase addresses the worst
remaining offender and eliminates cross-package duplication introduced during
the provider-interfaces refactoring.

## What Changes

### Decomposition

Extract three pure/near-pure helpers from `BuildContractCoverageFunc`:

1. **`buildEffectsSet`** — per-package analysis loop that populates the
   effects tracking set (currently lines 124–140)
2. **`buildCoverageMap`** — per-package quality pipeline + report-to-coverage
   conversion (currently lines 142–192)
3. **`computeCoverageReason`** — per-report `ContractCoverageInfo` construction
   including ambiguous-effects confidence range (currently lines 160–190)

After extraction, `BuildContractCoverageFunc` becomes a thin orchestrator
(~30 lines): resolve paths → build effects set → build coverage map →
construct closure.

### Deduplication

1. Export `loadTestPackage` as `goprovider.LoadTestPackage` and replace
   `loadTestPackageForQuality` in aireport with a reference to the exported
   function. Fixes the missing `pkg.Errors` validation in the aireport copy.
2. Add `config.LoadFromDir(moduleDir string) *GazeConfig` to the config
   package. Replace both private `loadGazeConfigBestEffort` copies.

## Capabilities

### New Capabilities
- `config.LoadFromDir`: Best-effort config loading by directory path (joins
  `.gaze.yaml`, falls back to `DefaultConfig()` on any error)
- `goprovider.LoadTestPackage`: Exported test package loader with `pkg.Errors`
  validation (previously unexported)

### Modified Capabilities
- `BuildContractCoverageFunc`: Identical behavior, reduced complexity (22 → ~6–8)
  via helper extraction. No signature or return value changes.

### Removed Capabilities
- `loadTestPackageForQuality` (aireport, unexported): Replaced by
  `goprovider.LoadTestPackage`
- `loadGazeConfigBestEffort` (goprovider + aireport, both unexported): Replaced
  by `config.LoadFromDir`

## Impact

- **Files modified**: `internal/provider/goprovider/contract.go`,
  `internal/aireport/runner_steps.go`, `internal/config/config.go`
- **Test files modified**: `contract_internal_test.go`,
  `runner_steps_test.go`, `config_test.go`
- **No API surface changes**: All modified functions are internal (unexported
  or package-internal). `GoContractCoverageProvider.Build()` and
  `BuildContractCoverageFunc` retain identical signatures and behavior.
- **No behavior changes**: Pure refactoring + dedup. All existing tests must
  pass without modification (error message strings may change for dedup).
- **CRAPload impact**: `BuildContractCoverageFunc` CRAP drops from 506 to
  <30 (helpers get their own test coverage). CRAPload expected: 30 → ~29.

## Constitution Alignment

Assessed against the Gaze project constitution (v1.3.0).

### I. Accuracy

**Assessment**: PASS

No analysis logic changes. Side effect detection, classification, and
contract coverage computation are unchanged. The three extracted helpers
preserve identical data flow. Existing regression tests verify accuracy
is maintained.

### II. Minimal Assumptions

**Assessment**: N/A

This change does not alter assumptions about host projects, test frameworks,
or coding styles. The `loadTestPackage` dedup fixes a missing `pkg.Errors`
check — strictly an improvement in assumption enforcement.

### III. Actionable Output

**Assessment**: N/A

No changes to output formats, report content, or metric computation. The
`ContractCoverageInfo` struct and its `Reason`/confidence fields are
constructed identically by the extracted `computeCoverageReason` helper.

### IV. Testability

**Assessment**: PASS

This change directly improves testability. Three extracted helpers are
pure or near-pure functions testable with synthetic data (no `go/packages`
loading required). The `testing.Short()` constraint is respected — all new
tests run without `-short` guard so they contribute to CRAPload coverage.
Coverage strategy: unit tests on each helper with synthetic inputs covering
all branches.
