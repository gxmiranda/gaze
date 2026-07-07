<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Extract Helper Function

- [x] 1.1 Add `isNewFunctionViolation(s Score, crapThreshold, gazeCRAPThreshold float64) bool` to `internal/crap/compare.go`, placed after `buildComparisonSummary`. The function returns `true` if `s.CRAP > crapThreshold` OR (`s.GazeCRAP != nil && *s.GazeCRAP > gazeCRAPThreshold`). Add GoDoc comment explaining it centralizes the new-function violation check used by summary counting, JSON output, and text output.

## 2. Replace Inline Logic

- [x] 2.1 In `buildComparisonSummary` (`internal/crap/compare.go:185-192`), replace the inline `crapViolation`/`gazeViolation` variables and `if` check with `if isNewFunctionViolation(s, opts.NewFunctionThreshold, opts.NewFunctionGazeCRAPThreshold)`. File: `internal/crap/compare.go`.
- [x] 2.2 In `WriteComparisonJSON` (`internal/crap/compare_report.go:119-123`), replace the inline `crapViolation`/`gazeViolation` variables and `if` check with `if isNewFunctionViolation(nf, result.Summary.NewFunctionThreshold, result.Summary.NewFunctionGazeCRAPThreshold)`. File: `internal/crap/compare_report.go`.
- [x] 2.3 In `writeComparisonNewFunctions` (`internal/crap/compare_report.go:274-276`), replace the inline `crapViolation`/`gazeViolation` variables and `if` check with `if isNewFunctionViolation(s, crapThreshold, gazeCRAPThreshold)`. File: `internal/crap/compare_report.go`.

## 3. Testing

- [x] 3.1 Add `TestIsNewFunctionViolation` table-driven test to `internal/crap/compare_test.go`. Cover all scenarios from the spec: CRAP-only violation, GazeCRAP-only violation, both thresholds violated, neither violated, nil GazeCRAP (no contract coverage data), exact-threshold boundary (not a violation -- strictly greater than).
- [x] 3.2 Run `go test -race -count=1 -short ./internal/crap/...` -- all existing comparison tests MUST pass unchanged, confirming zero behavioral regression.

## 4. Verification

- [x] 4.1 Run `go build ./...` -- MUST compile with zero errors.
- [x] 4.2 Run `go test -race -count=1 -short ./...` -- all tests MUST pass.
- [x] 4.3 Run `golangci-lint run` -- MUST pass with zero findings.

## 5. Documentation

- [x] 5.1 Update `AGENTS.md` Recent Changes section with `extract-violation-helper` entry describing the DRY extraction.
- [x] 5.2 No README, CLI docs, or website issue needed -- pure internal refactoring with no user-facing changes.

## 6. Constitution Alignment Verification

- [x] 6.1 Verify all four principles: Accuracy (no logic change, existing tests pass), Minimal Assumptions (no user-facing change), Actionable Output (no output change), Testability (helper directly tested with table-driven test covering all spec scenarios).
<!-- spec-review: passed -->
<!-- code-review: passed -->
