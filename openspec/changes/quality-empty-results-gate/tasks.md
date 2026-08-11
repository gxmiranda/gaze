## 1. Track Skipped Tests in quality.Assess

- [x] 1.1 Add `SkippedTests int` and `SkippedTestNames []string` fields to `taxonomy.PackageSummary` in `internal/taxonomy/types.go`. Use `json:"skipped_tests"` and `json:"skipped_test_names,omitempty"` tags. Add GoDoc comments.
- [x] 1.2 In `quality.Assess` (`internal/quality/quality.go`), add `skippedTests int` and `skippedTestNames []string` accumulators before the test function loop in the normal (SSA available) path. At the `len(targets) == 0` continue site (line ~192), increment `skippedTests` and append `tf.Name` to `skippedTestNames`. Keep the existing stderr warning. After the loop, set `summary.SkippedTests` and `summary.SkippedTestNames`.
- [x] 1.3 Verify that `BuildPackageSummary` (`internal/quality/quality.go`) does NOT need changes — `Assess` populates `SkippedTests`/`SkippedTestNames` directly on the summary it returns, so `BuildPackageSummary` (which aggregates from reports only) should not require modification. Document the verification.
- [x] 1.4 Add unit test: `quality.Assess` with test functions that have no resolvable targets — verify `summary.SkippedTests` and `summary.SkippedTestNames` are populated. Use the existing `BuildSSAFunc` injection pattern (`quality.Options`) to create a test fixture where `InferTargets` returns zero targets, or create a `testdata/src/bddpkg/` fixture with tests that call through interfaces.
- [x] 1.5 Add unit test: `quality.Assess` with all tests resolvable — verify `summary.SkippedTests` is 0 and `summary.SkippedTestNames` is empty.
- [x] 1.6 Add unit test: SSA degraded path — verify `summary.SkippedTests` is 0 (degraded path produces partial reports, does not skip).

## 2. Empty-Result Stdout Summary and Gate in runQuality

- [x] 2.1 In `runQuality` (`cmd/gaze/main.go`), replace the `len(allReports) == 0` block (lines ~1166-1169) with: (a) call `mergeSummaries(allSummaries)` first (move above the empty check), (b) write structured summary to `p.stdout` (total test count, paired count, skipped names, `--target` hint), (c) if `p.minContractCoverage > 0 || p.maxOverSpecification > 0`, return an error indicating thresholds cannot be evaluated (the `> 0` check is intentional — both thresholds have zero-means-disabled semantics). Otherwise return nil. (d) When `p.format == "json"`, emit valid JSON (empty reports array + summary with skipped fields) instead of plain text.
- [x] 2.2 Update `mergeSummaries` (`cmd/gaze/main.go`) to aggregate `SkippedTests` (sum) and `SkippedTestNames` (concatenate) across package summaries, consistent with existing aggregation of `TotalTests`, `TotalOverSpecifications`, `SSADegradedPackages`.
- [x] 2.3 Add unit test for `runQuality`: zero reports, no threshold flags — verify stdout contains summary with skipped test names and hint, verify exit 0.
- [x] 2.4 Add unit test for `runQuality`: zero reports, `--min-contract-coverage` set — verify non-zero exit and stdout still contains summary.
- [x] 2.5 Add unit test for `mergeSummaries`: multiple summaries with varying skipped test counts — verify aggregation.
- [x] 2.6 Add unit test for `runQuality`: zero reports, `--max-over-specification` set (without `--min-contract-coverage`) — verify non-zero exit. This tests the second branch of the OR condition.
- [x] 2.7 Add unit test for `runQuality`: zero reports, `--format=json` — verify valid JSON output with empty reports array and skipped test data in summary.

## 3. Skipped Test Section in Text Report

- [x] 3.1 In `quality.WriteText` (`internal/quality/report.go`), after writing the main quality table, check `summary.SkippedTests > 0`. If true, append a "Skipped Tests" section listing the names and a `--target` hint. Truncate to first 20 names if the list is long (with "... and N more" message).
- [x] 3.2 Add unit test: `WriteText` with `summary.SkippedTests > 0` — verify section appears with names and hint.
- [x] 3.3 Add unit test: `WriteText` with `summary.SkippedTests == 0` — verify no skipped section appears.
- [x] 3.4 Add unit test: `WriteText` with 25 skipped tests — verify truncation to first 20 names and "... and 5 more" message.

## 4. JSON Schema Update

- [x] 4.1 Update the quality JSON Schema in `internal/report/schema.go` to include `skipped_tests` (integer) and `skipped_test_names` (array of strings) in the `PackageSummary` definition.
- [x] 4.2 Verify existing JSON Schema validation tests pass with the new fields.

## 5. Report Pipeline Propagation

- [x] 5.1 Modify `runQualityForPackage` (`internal/aireport/runner_steps.go`) to return skipped test metadata from the `Assess`-returned `PackageSummary`. Currently returns `([]taxonomy.QualityReport, string)` — add skipped test count and names to the return values or introduce a result struct.
- [x] 5.2 Add `SkippedTests int` and `SkippedTestNames []string` to `qualityStepResult` (`internal/aireport/runner_steps.go`). Populate in `runQualityStep` alongside the existing `AvgContractCoverage` and `SSADegraded` extraction (line ~189-196).
- [x] 5.3 Add `SkippedTests int` to `ReportSummary` in `internal/aireport/payload.go`. Populate from `qualityStepResult` in `runProductionPipeline` (`internal/aireport/runner.go`).
- [x] 5.4 Add `SkippedTests int` field with JSON tag to `compactSummary` in `internal/aireport/compact.go`. Wire it in `CompactForAI()` so the AI reporter payload includes skipped test data.
- [x] 5.5 Update any existing pipeline tests that construct `qualityStepResult` or `ReportSummary` to include the new fields.

## 6. Cross-Cutting Verification

- [x] 6.1 Run `go test -race -count=1 -short ./...` — all existing tests must pass.
- [x] 6.2 Run `golangci-lint run` — no new lint violations.
- [x] 6.3 Verify constitution alignment: Principle III (Actionable Output) — confirm empty-result case now produces stdout output with clear guidance. Principle IV (Testability) — confirm all new code paths are covered by unit tests.
- [x] 6.4 Update AGENTS.md "Recent Changes" section with a `quality-empty-results-gate` entry summarizing the changes.
- [x] 6.5 Assess whether README.md needs updating (new stdout behavior for empty results, `--target` hint visibility).
