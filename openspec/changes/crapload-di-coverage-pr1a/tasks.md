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

## 1. Add contractCoverageDeps to internal/crap/

- [x] 1.1 Define `contractCoverageDeps` struct in `internal/crap/contract.go`
  with injectable fields: `loadAndAnalyze`, `classifyResults`, `loadTestPkg`,
  `assess`. Absorb the existing `aiMapperFn` variadic parameter into the struct
  as a field (Go does not allow two variadic parameters). Add a nil-means-default
  resolution block at the top of `analyzePackageCoverage`. Update
  `analyzePackageCoverage` signature to accept variadic `...contractCoverageDeps`.
  Update the call site in `BuildContractCoverageFunc` to pass the `aiMapperFn`
  through the deps struct (no behavioral change).

- [x] 1.2 Add unit tests for `analyzePackageCoverage` in
  `internal/crap/contract_test.go` using injected deps. Test cases:
  - Success path (all deps return valid data)
  - `loadAndAnalyze` returns error → returns `(nil, "")`
  - `loadAndAnalyze` returns empty results → returns `(nil, "")`
  - `classifyResults` returns nil → returns `(nil, "")`
  - `loadTestPkg` returns error → returns `(nil, "")`
  - `assess` returns error → returns `(nil, "")`
  - `assess` returns `SSADegraded=true` → returns `(reports, pkgPath)`
  Tests MUST NOT be guarded by `testing.Short()`.

- [x] 1.3 Add unit tests for `loadTestPackage` in
  `internal/crap/contract_test.go`. Test cases:
  - Package with test files (`internal/quality/testdata/src/welltested`) → success
  - Package without test files (`internal/analysis/testdata/src/returns`) → error
  - Non-existent package → error
  Tests MUST NOT be guarded by `testing.Short()`.

## 2. Add qualityPipelineDeps to internal/aireport/

- [x] 2.1 Define `qualityPipelineDeps` struct in
  `internal/aireport/runner_steps.go` with injectable fields:
  `resolvePackagePaths`, `loadAndAnalyze`, `classifyResults`, `loadTestPkg`,
  `assess`, `resolveModulePkgs`, `loadConfig`. Add nil-means-default resolution.
  Update signatures of `runQualityStep`, `runQualityForPackage`, and
  `runClassifyStep` to accept variadic `...qualityPipelineDeps`. Update call
  sites: `runQualityStep` calling `runQualityForPackage` must forward deps.
  `runProductionPipeline` call sites pass zero-value (no behavioral change).

- [x] 2.2 Add unit tests for `runQualityForPackage` in
  `internal/aireport/runner_steps_test.go` using injected deps. Test cases:
  - Success path (all deps return valid data) → non-nil reports
  - `loadAndAnalyze` returns error → returns `(nil, "")`
  - `loadAndAnalyze` returns empty results → returns `(nil, "")`
  - `classifyResults` returns error → returns `(nil, "")`
  - `loadTestPkg` returns error → returns `(nil, "")`
  - `assess` returns error → returns `(nil, "")`
  - `assess` returns `SSADegraded=true` → returns `(reports, pkgPath)`
  Tests MUST NOT be guarded by `testing.Short()`.

- [x] 2.3 Add unit tests for `runQualityStep` in
  `internal/aireport/runner_steps_test.go` using injected deps. Test cases:
  - Success with single package → valid `qualityStepResult`
  - Success with multiple packages → aggregated reports
  - `resolvePackagePaths` returns error → error
  - `resolvePackagePaths` returns empty → error
  - SSA degradation from one package propagates to result
  Tests MUST NOT be guarded by `testing.Short()`.

- [x] 2.4 Add unit tests for `runClassifyStep` in
  `internal/aireport/runner_steps_test.go` using injected deps. Test cases:
  - Success with label counts → correct Contractual/Ambiguous/Incidental
  - `resolvePackagePaths` returns error → error
  - `resolvePackagePaths` returns empty → error
  - `loadAndAnalyze` returns error for one package → skip, continue
  - `loadAndAnalyze` returns empty results for one package → skip, continue
  - `classifyResults` returns error for one package → skip, continue
  Tests MUST NOT be guarded by `testing.Short()`.

- [x] 2.5 Add unit tests for `loadTestPackageForQuality` in
  `internal/aireport/runner_steps_test.go`. Test cases:
  - Package with test files (`internal/quality/testdata/src/welltested`) → success
  - Package without test files (`internal/analysis/testdata/src/returns`) → error
  - Non-existent package → error
  Tests MUST NOT be guarded by `testing.Short()`.

## 3. Update pipelineStepFuncs call sites

- [x] 3.1 Update the `pipelineStepFuncs` function type signatures in
  `internal/aireport/runner.go` to match the new variadic signatures of
  `runQualityStep` and `runClassifyStep`. Update the fake step closures in
  `pipeline_internal_test.go` to match the new function types (minimal
  signature adaptation — add the variadic deps parameter, ignore it in
  the fake body). Verify that the nil-defaults in `runProductionPipeline`
  still correctly wire to the real functions. Existing
  `pipeline_internal_test.go` tests MUST continue to pass with the same
  behavioral assertions.

## 4. Verification

- [x] 4.1 Run `go test -race -count=1 -short ./internal/crap/... ./internal/aireport/...`
  and verify all new and existing tests pass.

- [x] 4.2 Run `go build ./cmd/gaze` and verify the binary builds without errors.

- [x] 4.3 Run `golangci-lint run ./internal/crap/... ./internal/aireport/...`
  and verify zero lint issues.

- [x] 4.4 Run `./gaze crap ./...` and verify CRAPload has decreased from 38
  toward the target of ~32. Record the new CRAPload value: **32**.
  Documentation impact: none (all changes are internal/unexported).

- [x] 4.5 Verify constitution alignment (Principle IV: Testability): confirm
  all new tests verify observable side effects (return values, error conditions,
  degradation signals) rather than implementation details. Confirm no new tests
  are guarded by `testing.Short()`.

<!-- spec-review: passed -->
<!-- code-review: passed -->
