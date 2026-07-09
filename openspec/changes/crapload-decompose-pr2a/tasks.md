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

## 1. Independent dedup targets

These tasks touch different packages and have no inter-dependencies.

- [x] 1.1 [P] Add `config.LoadFromDir` + tests
  - Add `LoadFromDir(moduleDir string) *GazeConfig` to
    `internal/config/config.go`
  - Add `path/filepath` to the import block
  - Construct path via `filepath.Join(moduleDir, ".gaze.yaml")`, call
    `config.Load`, return `DefaultConfig()` on any error
  - GoDoc comment MUST document best-effort semantics: errors are
    silently swallowed because callers use config as an optimization
    hint, not a hard requirement
  - Add 3 tests to `internal/config/config_test.go`:
    - `TestLoadFromDir_ValidConfig` — temp dir with valid `.gaze.yaml`
    - `TestLoadFromDir_MissingFile` — empty temp dir → `DefaultConfig()`
    - `TestLoadFromDir_InvalidYAML` — temp dir with malformed file →
      `DefaultConfig()`
  - **Files**: `internal/config/config.go`, `internal/config/config_test.go`

- [x] 1.2 [P] Export `LoadTestPackage` from goprovider
  - Rename `loadTestPackage` to `LoadTestPackage` in
    `internal/provider/goprovider/contract.go`
  - Update the internal caller at line 256 (`d.loadTestPkg = loadTestPackage`
    → `d.loadTestPkg = LoadTestPackage`)
  - Update test references in `contract_internal_test.go` (3 tests:
    `TestLoadTestPackage_*`) to use `LoadTestPackage`
  - Ensure GoDoc comment starts with `LoadTestPackage` (per CS-004 MUST)
  - No new tests needed — existing 3 tests cover all branches
  - **Files**: `internal/provider/goprovider/contract.go`,
    `internal/provider/goprovider/contract_internal_test.go`

## 2. Decompose `BuildContractCoverageFunc`

These tasks modify the same file (`contract.go` / `contract_internal_test.go`)
and MUST run sequentially.

- [x] 2.1 Extract `computeCoverageReason` helper + tests
  - Extract lines 160–190 from `BuildContractCoverageFunc` into
    `computeCoverageReason(report taxonomy.QualityReport) crap.ContractCoverageInfo`
  - Pure function: reads `ContractCoverage.TotalContractual`,
    `AmbiguousEffects`, `ContractCoverage.Percentage`
  - Replace inline logic in `BuildContractCoverageFunc` with call to
    `computeCoverageReason(report)`
  - Add 4 tests to `contract_internal_test.go`:
    - `TestComputeCoverageReason_WithContractualEffects` — Percentage
      passthrough, empty Reason
    - `TestComputeCoverageReason_AllAmbiguous` — Reason
      `"all_effects_ambiguous"`, correct min/max confidence
    - `TestComputeCoverageReason_NoEffects` — Reason
      `"no_effects_detected"`
    - `TestComputeCoverageReason_NilClassification` — entries with nil
      Classification skipped
  - Tests MUST NOT use `testing.Short()` guard
  - **Files**: `internal/provider/goprovider/contract.go`,
    `internal/provider/goprovider/contract_internal_test.go`

- [x] 2.2 Extract `buildEffectsSet` helper + tests
  - Extract lines 124–140 from `BuildContractCoverageFunc` into
    `buildEffectsSet(pkgPaths []string, loadAndAnalyzeFn func(string, analysis.Options) ([]taxonomy.AnalysisResult, error)) map[string]bool`
  - When `loadAndAnalyzeFn` is nil, default to `analysis.LoadAndAnalyze`
  - Replace inline logic in `BuildContractCoverageFunc` with
    `effectsSet := buildEffectsSet(pkgPaths, nil)`
  - Add 4 tests to `contract_internal_test.go`:
    - `TestBuildEffectsSet_WithEffects` — synthetic results with side
      effects → keys present
    - `TestBuildEffectsSet_NoEffects` — results with empty SideEffects →
      empty map
    - `TestBuildEffectsSet_AnalysisError` — loadAndAnalyze returns error →
      package silently skipped
    - `TestBuildEffectsSet_EmptyPaths` — empty pkgPaths → empty map
  - Tests MUST NOT use `testing.Short()` guard
  - **Files**: `internal/provider/goprovider/contract.go`,
    `internal/provider/goprovider/contract_internal_test.go`

- [x] 2.3 Extract `buildCoverageMap` helper + tests
  - Extract lines 142–192 (the per-package quality pipeline loop +
    report processing) from `BuildContractCoverageFunc` into
    `buildCoverageMap(pkgPaths []string, moduleDir string, gazeConfig *config.GazeConfig, stderr io.Writer, deps ...contractCoverageDeps) (map[string]crap.ContractCoverageInfo, []string)`
  - Must call `analyzePackageCoverage` per package and
    `computeCoverageReason` per report (uses helpers from 2.1)
  - Must skip reports with empty `TargetFunction.Function`
  - Must track degraded packages
  - Must keep highest-coverage entry when duplicates exist
  - Replace inline logic in `BuildContractCoverageFunc` with
    `coverageMap, degradedPkgs := buildCoverageMap(pkgPaths, moduleDir, gazeConfig, stderr, ccDeps)`
  - Add 5 tests to `contract_internal_test.go`:
    - `TestBuildCoverageMap_Success` — synthetic deps → map populated
    - `TestBuildCoverageMap_DegradedReport` — empty Function field → skipped
    - `TestBuildCoverageMap_SSADegradation` — degraded pkg path returned
    - `TestBuildCoverageMap_HigherCoverageWins` — two reports, higher pct wins
    - `TestBuildCoverageMap_EmptyPaths` — empty pkgPaths → empty map
  - Tests MUST NOT use `testing.Short()` guard
  - **Files**: `internal/provider/goprovider/contract.go`,
    `internal/provider/goprovider/contract_internal_test.go`

## 3. Wire dedup replacements

These tasks modify different files and can run in parallel, but depend
on Groups 1 and 2 being complete.

- [x] 3.1 [P] Replace `loadGazeConfigBestEffort` in goprovider with `config.LoadFromDir`
  - Replace call at `contract.go` line 113:
    `gazeConfig := loadGazeConfigBestEffort(moduleDir)` →
    `gazeConfig := config.LoadFromDir(moduleDir)`
  - Delete `loadGazeConfigBestEffort` function (lines 387–396)
  - Verify no other references remain in `contract.go`
  - **Files**: `internal/provider/goprovider/contract.go`

- [x] 3.2 [P] Replace `loadGazeConfigBestEffort` and `loadTestPackageForQuality` in aireport
  - Replace default in `resolveQualityDeps`:
    `d.loadTestPkg = loadTestPackageForQuality` →
    `d.loadTestPkg = goprovider.LoadTestPackage`
  - Replace default in `resolveQualityDeps`:
    `d.loadConfig = loadGazeConfigBestEffort` →
    `d.loadConfig = config.LoadFromDir`
  - Replace direct call at line 297:
    `cfg := loadGazeConfigBestEffort(moduleDir)` →
    `cfg := config.LoadFromDir(moduleDir)`
  - Delete `loadGazeConfigBestEffort` function (lines 354–363)
  - Delete `loadTestPackageForQuality` function (lines 365–392)
  - Update `runner_steps_test.go`:
    - Remove `TestLoadTestPackageForQuality_*` tests (3 tests — coverage
      now provided by goprovider's `TestLoadTestPackage_*`)
    - Delete `TestLoadGazeConfigBestEffort_AlwaysNonNil` — the function
      it tests is being removed; config loading is now tested by
      `TestLoadFromDir_*` tests in `config_test.go`
    - Update any error string assertions that changed due to
      `LoadTestPackage`'s different error messages
  - **Files**: `internal/aireport/runner_steps.go`,
    `internal/aireport/runner_steps_test.go`

## 4. Verification

- [x] 4.1 Build and test
  - Run `go build ./...` — MUST pass
  - Run `go test -race -count=1 -short ./...` — MUST pass
  - Run `golangci-lint run` — MUST pass with 0 issues
  - Verify no unused imports after function deletions
  - **Verify**: `BuildContractCoverageFunc` complexity ≤ 8 (check via
    `gocyclo -over 7 internal/provider/goprovider/contract.go`)
  - **Verify**: GoDoc comments on `LoadFromDir` and `LoadTestPackage`
    are complete and start with the function name (CS-004)
  - **Verify**: Net test delta — ~12 new tests added, ~4 removed

- [x] 4.2 Constitution alignment verification
  - **Accuracy (I)**: Confirm existing integration tests
    (`TestBuildContractCoverageFunc_WelltestedPackage`) still pass —
    coverage data is computed identically
  - **Testability (IV)**: Confirm all new helper tests run without
    `testing.Short()` guard — they contribute to CRAPload
  - **Minimal Assumptions (II)**: N/A — no analysis behavior changed
  - **Actionable Output (III)**: N/A — no output format changed

<!-- spec-review: passed -->

<!-- code-review: passed -->
