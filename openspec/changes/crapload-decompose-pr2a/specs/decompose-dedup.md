## ADDED Requirements

### Requirement: computeCoverageReason helper

`computeCoverageReason` MUST be a pure function that takes a single
`taxonomy.QualityReport` and returns a `crap.ContractCoverageInfo` with
the correct `Percentage`, `Reason`, `MinConfidence`, and `MaxConfidence`
fields.

The function MUST produce identical output to the inline logic it replaces
(lines 160–190 of `BuildContractCoverageFunc`).

#### Scenario: Report with contractual effects

- **GIVEN** a `QualityReport` with `ContractCoverage.TotalContractual > 0`
  and `Percentage = 75.0`
- **WHEN** `computeCoverageReason` is called
- **THEN** the returned `ContractCoverageInfo` SHALL have `Percentage = 75.0`
  and `Reason = ""` (empty — no special reason needed)

#### Scenario: Report with all ambiguous effects

- **GIVEN** a `QualityReport` with `ContractCoverage.TotalContractual == 0`
  and `AmbiguousEffects` containing 2 entries with confidences 78 and 79
- **WHEN** `computeCoverageReason` is called
- **THEN** the returned `ContractCoverageInfo` SHALL have
  `Reason = "all_effects_ambiguous"`, `MinConfidence = 78`,
  `MaxConfidence = 79`

#### Scenario: Report with no effects detected

- **GIVEN** a `QualityReport` with `ContractCoverage.TotalContractual == 0`
  and empty `AmbiguousEffects`
- **WHEN** `computeCoverageReason` is called
- **THEN** the returned `ContractCoverageInfo` SHALL have
  `Reason = "no_effects_detected"`

#### Scenario: AmbiguousEffects with nil Classification

- **GIVEN** a `QualityReport` with `ContractCoverage.TotalContractual == 0`
  and `AmbiguousEffects` containing entries where `Classification` is nil
- **WHEN** `computeCoverageReason` is called
- **THEN** entries with nil `Classification` MUST be skipped in confidence
  range computation

### Requirement: buildEffectsSet helper

`buildEffectsSet` MUST iterate over package paths, run side effect analysis
on each, and return a `map[string]bool` of `"shortPkg:qualifiedName"` keys
for functions with one or more detected side effects.

The function MUST accept a `loadAndAnalyze` function parameter for
dependency injection. When nil, it MUST default to
`analysis.LoadAndAnalyze`.

Analysis errors for individual packages MUST be silently skipped (best-effort
semantics, matching current behavior).

#### Scenario: Package with effects

- **GIVEN** a `loadAndAnalyze` function that returns results with side effects
- **WHEN** `buildEffectsSet` is called
- **THEN** the returned map SHALL contain the key `"pkg:DoWork"` for a
  function `DoWork` in package `example.com/pkg` (key format:
  `"shortPkg:qualifiedName"`)

#### Scenario: Package with analysis error

- **GIVEN** a `loadAndAnalyze` function that returns an error
- **WHEN** `buildEffectsSet` is called
- **THEN** the package MUST be silently skipped and the returned map SHALL
  not contain entries for that package

#### Scenario: Empty package paths

- **GIVEN** an empty `pkgPaths` slice
- **WHEN** `buildEffectsSet` is called
- **THEN** the returned map SHALL be empty (not nil)

### Requirement: buildCoverageMap helper

`buildCoverageMap` MUST iterate over package paths, run the quality pipeline
via `analyzePackageCoverage` on each, convert reports to
`ContractCoverageInfo` entries using `computeCoverageReason`, and return
the coverage map and degraded packages list.

The function MUST skip reports with empty `TargetFunction.Function` (degraded
reports) to prevent phantom entries.

When multiple reports exist for the same function, the entry with the higher
`Percentage` MUST win.

#### Scenario: Single package with coverage data

- **GIVEN** a package path where `analyzePackageCoverage` returns reports
- **WHEN** `buildCoverageMap` is called
- **THEN** the returned map SHALL contain entries keyed by
  `"shortPkg:qualifiedName"` with correct coverage info

#### Scenario: Degraded report skipping

- **GIVEN** a report with `TargetFunction.Function == ""`
- **WHEN** `buildCoverageMap` processes the report
- **THEN** the report MUST be skipped and not added to the coverage map

#### Scenario: SSA degradation tracking

- **GIVEN** a package where `analyzePackageCoverage` returns a non-empty
  degraded package path
- **WHEN** `buildCoverageMap` is called
- **THEN** the degraded package path SHALL be included in the returned
  `degradedPkgs` slice

#### Scenario: Higher-coverage wins

- **GIVEN** two reports for the same function, encountered in either order
  (50.0 then 80.0, or 80.0 then 50.0)
- **WHEN** `buildCoverageMap` processes both
- **THEN** the coverage map entry SHALL have `Percentage = 80.0` regardless
  of encounter order

### Requirement: config.LoadFromDir

`config.LoadFromDir` MUST accept a module directory path, construct the
`.gaze.yaml` path via `filepath.Join(moduleDir, ".gaze.yaml")`, call
`config.Load`, and return the result. On any error from `config.Load`,
it MUST return `DefaultConfig()`.

#### Scenario: Valid directory with .gaze.yaml

- **GIVEN** a directory containing a valid `.gaze.yaml` file
- **WHEN** `LoadFromDir` is called with that directory
- **THEN** the returned config SHALL reflect the values from the file

#### Scenario: Directory without .gaze.yaml

- **GIVEN** a directory with no `.gaze.yaml` file
- **WHEN** `LoadFromDir` is called
- **THEN** it SHALL return `DefaultConfig()` (no error)

#### Scenario: Directory with invalid .gaze.yaml

- **GIVEN** a directory with a malformed `.gaze.yaml` file
- **WHEN** `LoadFromDir` is called
- **THEN** it SHALL return `DefaultConfig()` (error swallowed)

### Requirement: LoadTestPackage export

`goprovider.LoadTestPackage` MUST be the exported form of the existing
`loadTestPackage` function. It MUST retain the `pkg.Errors` validation
loop that checks for package-level errors before selecting a test package.

#### Scenario: Package with test syntax

- **GIVEN** a valid Go package path with test files
- **WHEN** `LoadTestPackage` is called
- **THEN** it SHALL return the package that has test syntax

#### Scenario: Package with errors

- **GIVEN** a Go package path where loaded packages have `Errors`
- **WHEN** `LoadTestPackage` is called
- **THEN** it SHALL return an error containing the package error messages

## MODIFIED Requirements

### Requirement: BuildContractCoverageFunc complexity

`BuildContractCoverageFunc` MUST delegate to `buildEffectsSet`,
`buildCoverageMap`, and `computeCoverageReason` instead of inlining
their logic. The function's cyclomatic complexity MUST be 8 or less.

Previously: All logic was inline in a single 120-line function with
complexity 22.

### Requirement: aireport test package loading

`qualityPipelineDeps.loadTestPkg` default MUST resolve to
`goprovider.LoadTestPackage` instead of the local
`loadTestPackageForQuality`.

Previously: Default was `loadTestPackageForQuality` (local copy without
`pkg.Errors` validation).

### Requirement: aireport config loading

`qualityPipelineDeps.loadConfig` default MUST resolve to
`config.LoadFromDir` instead of the local `loadGazeConfigBestEffort`.

Previously: Default was local `loadGazeConfigBestEffort` (duplicate of
goprovider's copy).

## REMOVED Requirements

### Requirement: loadTestPackageForQuality (aireport)

Removed — replaced by `goprovider.LoadTestPackage`. The aireport copy
lacked `pkg.Errors` validation, making it strictly inferior to the
goprovider version.

### Requirement: loadGazeConfigBestEffort (both packages)

Removed — replaced by `config.LoadFromDir`. Two identical copies
consolidated into the config package where the logic belongs.
