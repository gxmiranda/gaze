## ADDED Requirements

### Requirement: contractCoverageDeps injection for analyzePackageCoverage

`analyzePackageCoverage` in `internal/provider/goprovider/contract.go` MUST accept an optional
variadic `contractCoverageDeps` parameter. When no deps argument is supplied, the
function MUST use the real implementations (`analysis.LoadAndAnalyze`,
`classifyResults`, `loadTestPackage`, `quality.Assess`). When a deps argument is
supplied, the function MUST use the injected implementations.

The `contractCoverageDeps` struct MUST contain the following injectable fields:
- `loadAndAnalyze` — replaces `analysis.LoadAndAnalyze`
- `classifyResults` — replaces `classifyResults`
- `loadTestPkg` — replaces `loadTestPackage`
- `assess` — replaces `quality.Assess`

The struct MUST be unexported (package-internal).

#### Scenario: Production call site unchanged
- **GIVEN** `analyzePackageCoverage` is called without a deps argument
- **WHEN** the function executes
- **THEN** it MUST use the real implementations and produce identical results to
  the current behavior

#### Scenario: Test with injected deps — success path
- **GIVEN** `analyzePackageCoverage` is called with a deps argument providing
  synthetic implementations that return valid results
- **WHEN** the function executes
- **THEN** it MUST return non-nil quality reports using the injected data

#### Scenario: Test with injected deps — analysis failure
- **GIVEN** `analyzePackageCoverage` is called with a deps argument where
  `loadAndAnalyze` returns an error
- **WHEN** the function executes
- **THEN** it MUST return `(nil, "")` without calling subsequent pipeline steps

#### Scenario: Test with injected deps — SSA degradation
- **GIVEN** `analyzePackageCoverage` is called with a deps argument where
  `assess` returns a summary with `SSADegraded=true`
- **WHEN** the function executes
- **THEN** the second return value MUST be the package path (non-empty string)

### Requirement: qualityPipelineDeps injection for aireport orchestration functions

`runQualityStep`, `runQualityForPackage`, and `runClassifyStep` in
`internal/aireport/runner_steps.go` MUST accept an optional variadic
`qualityPipelineDeps` parameter. When no deps argument is supplied, the functions
MUST use the real implementations. When a deps argument is supplied, the
functions MUST use the injected implementations.

The `qualityPipelineDeps` struct MUST contain injectable fields for all heavy
I/O operations called by these functions:
- `resolvePackagePaths` — replaces `loader.ResolvePackagePaths`
- `loadAndAnalyze` — replaces `analysis.LoadAndAnalyze`
- `classifyResults` — replaces `runClassifyResults`
- `loadTestPkg` — replaces `loadTestPackageForQuality`
- `assess` — replaces `quality.Assess`
- `resolveModulePkgs` — replaces `resolveModulePackages`
- `loadConfig` — replaces `loadGazeConfigBestEffort`

The struct MUST be unexported (package-internal).

#### Scenario: runQualityStep — success with multiple packages
- **GIVEN** `runQualityStep` is called with deps where `resolvePackagePaths`
  returns 2 package paths and `runQualityForPackage` returns valid reports
  for each
- **WHEN** the function executes
- **THEN** it MUST return a `qualityStepResult` with aggregated reports from
  both packages and non-nil JSON

#### Scenario: runQualityStep — resolve failure
- **GIVEN** `runQualityStep` is called with deps where `resolvePackagePaths`
  returns an error
- **WHEN** the function executes
- **THEN** it MUST return an error without calling any subsequent pipeline steps

#### Scenario: runQualityStep — SSA degradation propagation
- **GIVEN** `runQualityStep` is called with deps where one package's quality
  assessment returns `SSADegraded=true`
- **WHEN** the function executes
- **THEN** the result MUST have `SSADegraded=true` and the degraded package path
  in `SSADegradedPackages`

#### Scenario: runClassifyStep — success with label counts
- **GIVEN** `runClassifyStep` is called with deps where analysis and
  classification succeed for a package
- **WHEN** the function executes
- **THEN** the result MUST contain correct `Contractual`, `Ambiguous`, and
  `Incidental` counts from `classify.CountLabels`

#### Scenario: runClassifyStep — resolve failure
- **GIVEN** `runClassifyStep` is called with deps where `resolvePackagePaths`
  returns an error
- **WHEN** the function executes
- **THEN** it MUST return an error

#### Scenario: runQualityForPackage — success path
- **GIVEN** `runQualityForPackage` is called with deps where all sub-steps
  succeed
- **WHEN** the function executes
- **THEN** it MUST return non-nil quality reports and an empty degraded
  package path

#### Scenario: runQualityForPackage — no test files
- **GIVEN** `runQualityForPackage` is called with deps where `loadTestPkg`
  returns an error (no test files found)
- **WHEN** the function executes
- **THEN** it MUST return `(nil, "")`

### Requirement: loadTestPackage unit tests

`loadTestPackage` in `internal/provider/goprovider/contract.go` MUST have unit tests that
exercise both the success and error paths. Tests MUST NOT be guarded by
`testing.Short()`.

#### Scenario: Package with test files
- **GIVEN** a package path pointing to a testdata fixture with test files
  (e.g., `internal/quality/testdata/src/welltested`)
- **WHEN** `loadTestPackage` is called
- **THEN** it MUST return a non-nil `*packages.Package` with test syntax

#### Scenario: Package without test files
- **GIVEN** a package path pointing to a testdata fixture without test files
  (e.g., `internal/analysis/testdata/src/returns`)
- **WHEN** `loadTestPackage` is called
- **THEN** it MUST return an error containing "no test files found"

#### Scenario: Non-existent package
- **GIVEN** a package path that does not exist
- **WHEN** `loadTestPackage` is called
- **THEN** it MUST return an error

### Requirement: loadTestPackageForQuality unit tests

`loadTestPackageForQuality` in `internal/aireport/runner_steps.go` MUST have
unit tests that exercise both the success and error paths. Tests MUST NOT be
guarded by `testing.Short()`.

#### Scenario: Package with test files
- **GIVEN** a package path pointing to a testdata fixture with test files
- **WHEN** `loadTestPackageForQuality` is called
- **THEN** it MUST return a non-nil `*packages.Package` with test syntax

#### Scenario: Package without test files
- **GIVEN** a package path pointing to a testdata fixture without test files
- **WHEN** `loadTestPackageForQuality` is called
- **THEN** it MUST return an error containing "no test package found"

## MODIFIED Requirements

### Requirement: analyzePackageCoverage signature

`analyzePackageCoverage` MUST accept an optional variadic `contractCoverageDeps`
parameter as the last argument. Previously: the function accepted only
`pkgPath`, `moduleDir`, `gazeConfig`, `stderr`, and optional `aiMapperFn`.

### Requirement: runQualityStep signature

`runQualityStep` MUST accept an optional variadic `qualityPipelineDeps`
parameter as the last argument. Previously: the function accepted only
`patterns`, `moduleDir`, and `stderr`.

### Requirement: runClassifyStep signature

`runClassifyStep` MUST accept an optional variadic `qualityPipelineDeps`
parameter as the last argument. Previously: the function accepted only
`patterns` and `moduleDir`.

### Requirement: runQualityForPackage signature

`runQualityForPackage` MUST accept an optional variadic `qualityPipelineDeps`
parameter as the last argument. Previously: the function accepted only
`pkgPath`, `gazeConfig`, `modPkgs`, and `stderr`.

## REMOVED Requirements

None.
