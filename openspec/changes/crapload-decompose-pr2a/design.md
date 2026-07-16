## Context

`BuildContractCoverageFunc` (complexity 22, CRAP 506) in
`internal/provider/goprovider/contract.go` is the worst CRAPload offender.
It interleaves three concerns: effects discovery, coverage map construction
with reason computation, and closure assembly. Two utility functions
(`loadTestPackage`, `loadGazeConfigBestEffort`) are duplicated across
`goprovider` and `aireport`.

Phase 1 (PRs #183, #187) reduced CRAPload from 38 to 30 via DI and tests
on simpler functions. This phase targets the worst remaining function and
eliminates cross-package duplication.

### Current structure (contract.go lines 96–216)

```
BuildContractCoverageFunc(patterns, moduleDir, stderr, aiMapperFn)
  ├─ ResolvePackagePaths(patterns, moduleDir)
  ├─ loadGazeConfigBestEffort(moduleDir)          ← DUPLICATED in aireport
  ├─ for each pkgPath:
  │   ├─ analysis.LoadAndAnalyze(pkgPath)         ← effects discovery
  │   │   └─ populate effectsSet[key] = true
  │   ├─ analyzePackageCoverage(pkgPath, ...)      ← quality pipeline
  │   └─ for each report:                          ← coverage map build
  │       ├─ skip degraded (empty Function)
  │       ├─ compute ContractCoverageInfo           ← reason computation
  │       │   └─ ambiguous confidence range loop
  │       └─ coverageMap[key] = info
  ├─ early return if empty
  └─ return closure(pkg, function):
      ├─ check coverageMap
      ├─ check effectsSet → "no_test_coverage"
      └─ default → "no_effects_detected"
```

## Goals / Non-Goals

### Goals
- Reduce `BuildContractCoverageFunc` complexity from 22 to ~6–8
- Make each extracted concern independently testable with synthetic data
- Eliminate `loadTestPackageForQuality` (fixes missing `pkg.Errors` bug)
- Eliminate duplicate `loadGazeConfigBestEffort` across two packages
- All new tests run without `testing.Short()` guard (CRAPload-visible)

### Non-Goals
- Unifying the `classifyResults` signature divergence between
  `contractCoverageDeps` and `qualityPipelineDeps` (acknowledged design
  decision D2 from PR #183 — out of scope)
- Decomposing `analyzePackageCoverage` further (already has DI + 9 tests)
- Changing the `GoContractCoverageProvider` interface or `Build()` method
- Reducing complexity of other Phase 2 targets (`matchContainerUnwrap`,
  `detectASTReceiverMutations`, etc.)

## Decisions

### D1: Extract 3 helpers from `BuildContractCoverageFunc`

**Decision**: Decompose into `buildEffectsSet`, `buildCoverageMap`, and
`computeCoverageReason`.

**Rationale**: Each corresponds to a distinct concern:
1. `buildEffectsSet` — data acquisition (side effect analysis per package)
2. `buildCoverageMap` — data transformation (quality reports → coverage info)
3. `computeCoverageReason` — value construction (report → CoverageInfo)

`computeCoverageReason` is a pure function (no I/O, no package loading). The
other two need DI for testability but are logically self-contained.

**Alternatives rejected**:
- *2 helpers only* (merge reason computation into buildCoverageMap): Keeps
  complexity higher in `buildCoverageMap` and makes the reason logic harder
  to test in isolation.
- *Full DI struct on BuildContractCoverageFunc*: Over-engineered — the
  function already delegates to `analyzePackageCoverage` which has its own
  DI. Adding another DI layer creates indirection without proportional
  benefit.

### D2: Export `LoadTestPackage` from goprovider

**Decision**: Export the existing `loadTestPackage` as
`goprovider.LoadTestPackage` and have `aireport` import it.

**Rationale**: The goprovider version has `pkg.Errors` validation that the
aireport version lacks. `goprovider` already imports `quality` (for
`HasTestSyntax`), so no new dependency edges are introduced. The function
signature is unchanged: `func(string) (*packages.Package, error)`.

**Alternatives rejected**:
- *Move to `loader` package*: Would create circular dependency
  (`loader` → `quality` for `HasTestSyntax`).
- *New `testloader` package*: Unnecessary package proliferation for a single
  function.
- *Callback parameter for HasTestSyntax*: Over-abstracted — `HasTestSyntax`
  is a stable API unlikely to change.

### D3: Add `config.LoadFromDir` to the config package

**Decision**: Add `config.LoadFromDir(moduleDir string) *GazeConfig` that
constructs the `.gaze.yaml` path, calls `config.Load`, and returns
`DefaultConfig()` on any error. Replace both private copies.

**Rationale**: `config.Load` already handles the missing-file case (returns
`DefaultConfig()` when file doesn't exist). `LoadFromDir` adds the
directory→path join and the error-swallowing behavior. This belongs in the
config package because it's config-loading logic, not package-loading logic.

**Signature**:
```go
// LoadFromDir loads the GazeConfig from the given module directory,
// looking for .gaze.yaml. Returns DefaultConfig() on any error.
func LoadFromDir(moduleDir string) *GazeConfig
```

### D4: DI approach for `buildEffectsSet` and `buildCoverageMap`

**Decision**: Use the existing `contractCoverageDeps` struct. The
`buildEffectsSet` helper accepts a `loadAndAnalyze` function parameter
directly (it only needs one dependency). `buildCoverageMap` accepts the
full `contractCoverageDeps` (it delegates to `analyzePackageCoverage`
which already uses it).

**Rationale**: Follows the established nil-means-default DI pattern from
Phase 1a (PR #183). No new DI struct needed — reuse existing one.

### D5: `computeCoverageReason` — pure function, no DI

**Decision**: `computeCoverageReason` takes a `taxonomy.QualityReport` and
returns `crap.ContractCoverageInfo`. No DI needed.

**Rationale**: The function only reads struct fields and computes derived
values. No I/O, no package loading, no external dependencies. Test directly
with struct literals.

## Risks / Trade-offs

### Risk: aireport now imports goprovider

The `loadTestPackage` dedup creates a new import edge:
`internal/aireport` → `internal/provider/goprovider`. Currently aireport
already imports goprovider (line 18: `runner_steps.go`), so this is an
existing edge, not a new one. **No new coupling introduced.**

### Risk: Error message changes in test assertions

The `loadTestPackageForQuality` error messages differ slightly from
`LoadTestPackage` (e.g., `"no test package found"` vs `"no test files
found"`). Tests in `runner_steps_test.go` that assert on specific error
strings will need updating. **Low risk — mechanical change.**

### Trade-off: Extracted helpers are package-internal

The 3 new helpers (`buildEffectsSet`, `buildCoverageMap`,
`computeCoverageReason`) remain unexported in the `goprovider` package.
They exist for decomposition and testability, not for external
consumption. This keeps the public API surface unchanged.

### Trade-off: `buildEffectsSet` duplicates the analysis call

`BuildContractCoverageFunc` calls `analysis.LoadAndAnalyze` twice per
package: once in `buildEffectsSet` and once inside
`analyzePackageCoverage` (via `buildCoverageMap`). This duplication
exists in the current code and is intentional — the effects set must be
built even when `loadTestPackage` fails (no tests), while
`analyzePackageCoverage` short-circuits on test-load failure. Merging
these would couple the two concerns. The performance cost is negligible
(SSA build is cached by `go/packages`).
