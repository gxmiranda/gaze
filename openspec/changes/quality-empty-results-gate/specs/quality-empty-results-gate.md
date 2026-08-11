## ADDED Requirements

### Requirement: Skipped Test Tracking in PackageSummary

`quality.Assess` MUST track test functions that are skipped due to unresolvable
targets. `taxonomy.PackageSummary` MUST include `SkippedTests int` and
`SkippedTestNames []string` fields. `SkippedTests` MUST be incremented for
every test function where target inference returns zero targets (the `continue`
path). `SkippedTestNames` MUST record the name of each skipped test function.

The existing `TotalTests` semantics MUST NOT change — it counts only
successfully paired test-target pairs.

#### Scenario: All tests have resolvable targets

- **GIVEN** a package where every test function calls the target directly
- **WHEN** `quality.Assess` runs
- **THEN** `summary.SkippedTests` is `0`
- **AND** `summary.SkippedTestNames` is empty
- **AND** `summary.TotalTests` equals the number of test functions

#### Scenario: Some tests have unresolvable targets

- **GIVEN** a package with 5 test functions, 2 using dynamic dispatch
- **WHEN** `quality.Assess` runs
- **THEN** `summary.SkippedTests` is `2`
- **AND** `summary.SkippedTestNames` contains the 2 skipped test names
- **AND** `summary.TotalTests` is `3` (only paired tests)

#### Scenario: All tests have unresolvable targets

- **GIVEN** a package where no test function has a resolvable target (e.g., Ginkgo suite)
- **WHEN** `quality.Assess` runs
- **THEN** `summary.SkippedTests` equals the total test function count
- **AND** `summary.SkippedTestNames` contains all test names
- **AND** `summary.TotalTests` is `0`
- **AND** the returned reports slice is empty

#### Scenario: SSA degraded path does not skip tests

- **GIVEN** a package where SSA construction fails
- **WHEN** `quality.Assess` runs in degraded mode
- **THEN** `summary.SkippedTests` is `0`
- **AND** partial reports are generated for all test functions (existing behavior)

---

### Requirement: Empty-Result Stdout Summary

When `gaze quality` produces zero quality reports, it MUST write a structured
summary to stdout listing:
- The total number of test functions found
- The number successfully paired (0)
- The names of skipped test functions
- A hint directing the user to `--target` as a manual override

This summary MUST be written regardless of whether threshold flags are set.

#### Scenario: Zero reports, no threshold flags

- **GIVEN** `gaze quality ./bdd-pkg/` where all 10 tests have unresolvable targets
- **WHEN** no threshold flags are provided
- **THEN** stdout contains "0 of 10 test functions mapped to a target"
- **AND** stdout lists all 10 skipped test names
- **AND** stdout contains a hint mentioning `--target`
- **AND** the command exits 0

#### Scenario: Zero reports, threshold flags set

- **GIVEN** `gaze quality --min-contract-coverage=10 ./bdd-pkg/`
- **WHEN** all tests have unresolvable targets
- **THEN** stdout contains the same summary as the no-threshold case
- **AND** the command exits non-zero
- **AND** stderr contains an error about thresholds not being evaluable

#### Scenario: Zero reports across multiple packages

- **GIVEN** `gaze quality ./pkg1/... ./pkg2/...` where both packages have only unresolvable tests
- **WHEN** the command runs
- **THEN** stdout lists skipped tests from both packages
- **AND** the total count reflects all packages

#### Scenario: Zero threshold explicitly set

- **GIVEN** `gaze quality --min-contract-coverage=0 ./bdd-pkg/`
- **WHEN** all tests have unresolvable targets
- **THEN** the command exits 0 (threshold of 0 is treated as "not set")
- **AND** stdout contains the skipped-test summary

#### Scenario: No test files in package (no summaries)

- **GIVEN** `gaze quality ./pkg-without-tests/` where the package has no `*_test.go` files
- **WHEN** the command runs
- **THEN** no `PackageSummary` is created for that package (no summary to aggregate)
- **AND** stdout indicates no quality reports generated
- **AND** the message does not claim tests were "skipped" (there were none)

#### Scenario: JSON format with zero reports

- **GIVEN** `gaze quality --format=json ./bdd-pkg/` where all tests are skipped
- **WHEN** the command runs
- **THEN** stdout contains valid JSON with an empty `quality_reports` array
- **AND** the JSON includes a `quality_summary` with `skipped_tests` and `skipped_test_names`
- **AND** no plain-text summary is intermixed with the JSON output

---

### Requirement: Empty-Result Gate Failure for Quality Thresholds

When threshold flags (`--min-contract-coverage`, `--max-over-specification`)
are set but zero test-target pairs are resolved, `gaze quality` MUST exit
non-zero. The error message MUST indicate that thresholds cannot be evaluated
and MUST suggest checking package patterns or using `--target`.

This mirrors the existing `gaze crap` zero-result gate pattern.

#### Scenario: min-contract-coverage set, zero pairs

- **GIVEN** `gaze quality --min-contract-coverage=10 ./pkg/`
- **WHEN** zero test-target pairs are resolved
- **THEN** the command exits non-zero
- **AND** the error mentions "cannot evaluate thresholds"

#### Scenario: max-over-specification set, zero pairs

- **GIVEN** `gaze quality --max-over-specification=3 ./pkg/`
- **WHEN** zero test-target pairs are resolved
- **THEN** the command exits non-zero
- **AND** the error mentions "cannot evaluate thresholds"

#### Scenario: No threshold flags, zero pairs

- **GIVEN** `gaze quality ./pkg/` (no threshold flags)
- **WHEN** zero test-target pairs are resolved
- **THEN** the command exits 0
- **AND** stdout contains the skipped-test summary

---

### Requirement: Skipped Test Section in Text Report

When some tests are successfully paired and others are skipped, the text
quality report MUST include a "Skipped Tests" section listing the names of
skipped tests with guidance. This section MUST appear after the main quality
table.

#### Scenario: Mixed paired and skipped tests

- **GIVEN** a package with 10 tests, 7 paired and 3 skipped
- **WHEN** `gaze quality ./pkg/` runs with text output
- **THEN** the text output includes the quality table for the 7 paired tests
- **AND** a "Skipped Tests" section lists the 3 skipped test names
- **AND** the section includes a hint about `--target`

#### Scenario: All tests paired

- **GIVEN** a package with 5 tests, all paired
- **WHEN** `gaze quality ./pkg/` runs with text output
- **THEN** no "Skipped Tests" section appears in the output

#### Scenario: Large number of skipped tests (truncation)

- **GIVEN** a package with 25 skipped tests
- **WHEN** text output is generated
- **THEN** the first 20 skipped test names are listed
- **AND** a "... and 5 more" message appears
- **AND** the total skipped count header reflects the full 25

---

### Requirement: Skipped Test Aggregation in mergeSummaries

`mergeSummaries` MUST aggregate `SkippedTests` (sum) and `SkippedTestNames`
(concatenate) across multiple `PackageSummary` values, consistent with how
`TotalTests`, `TotalOverSpecifications`, and `SSADegradedPackages` are
aggregated.

#### Scenario: Multiple packages with skipped tests

- **GIVEN** 3 package summaries with 2, 0, and 3 skipped tests respectively
- **WHEN** `mergeSummaries` is called
- **THEN** the merged `SkippedTests` is 5
- **AND** the merged `SkippedTestNames` contains all 5 names

---

### Requirement: Quality JSON Schema Update

The quality JSON Schema MUST include `skipped_tests` (integer) and
`skipped_test_names` (array of strings, optional) in the `PackageSummary`
schema definition. Existing schema validation tests MUST pass with the new
fields present.

---

### Requirement: Report Pipeline Skipped Test Propagation

`ReportSummary` in `internal/aireport/payload.go` MUST include a
`SkippedTests int` field that surfaces the skipped test count from the quality
step. This field MUST be populated from the merged `PackageSummary` so that
AI reporters and threshold evaluation have access to the information.

#### Scenario: Report pipeline with skipped tests

- **GIVEN** a `gaze report` run where some tests are skipped
- **WHEN** the quality step completes
- **THEN** `ReportSummary.SkippedTests` reflects the total skipped count
- **AND** the AI reporter payload includes the skipped test count

---

## MODIFIED Requirements

### Requirement: quality.Assess Warning Behavior

Previously: `quality.Assess` printed warnings to `opts.Stderr` for skipped
tests but did not record the skip in the returned `PackageSummary`.

Now: `quality.Assess` MUST continue printing warnings to `opts.Stderr` (for
backward compatibility with users parsing stderr) AND MUST record the skip
in the `PackageSummary` via `SkippedTests` and `SkippedTestNames` fields.

---

### Requirement: runQuality Empty Result Behavior

Previously: `runQuality` returned `nil` (exit 0) with no stdout output when
`allReports` was empty. Only a logger warning was printed to stderr.

Now: `runQuality` MUST write a structured summary to stdout AND MUST return
a non-zero exit code when threshold flags are set. When no threshold flags
are set, the command MUST still exit 0 (warn-and-exit-0 pattern).

---

## REMOVED Requirements

None.
