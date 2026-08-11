## Why

`gaze quality` produces no output and exits 0 when every test function in a
package has unresolvable targets — the common case for Ginkgo/BDD suites and
any framework that uses interface-based or dynamic dispatch for test
registration. This is a silent failure: CI gates pass, the user sees nothing,
and there is no indication that gaze was unable to analyze anything.

The root cause is a four-link silent-skip chain:

1. **`resolveCallee` returns nil** for `call.Call.IsInvoke()` (interface/dynamic
   dispatch). Ginkgo's closure registry uses dynamic dispatch, making its test
   targets invisible to SSA.

2. **`InferTargets` returns nil** when the candidates map is empty — no target
   function could be identified for the test.

3. **`quality.Assess` silently skips** tests with no targets. A warning is
   printed to `opts.Stderr`, but the test is `continue`d over. The resulting
   `QualityReport` has `TotalTests` reflecting only successfully-paired tests.

4. **`runQuality` returns nil** (exit 0) when `allReports` is empty. A log
   message goes to stderr via the logger, but nothing is written to stdout.
   There is no gate failure mechanism.

This violates Constitution Principle III (Actionable Output): the user receives
no guidance toward improvement. A CI gate that silently passes on zero data is
worse than no gate at all — it creates false confidence.

Issue: #103

## What Changes

### Empty-result gate for `gaze quality` and `gaze report` quality step

When `gaze quality` (or the quality step in `gaze report`) produces zero
quality reports, the command must:

- **Always**: Write a structured summary to stdout listing how many test
  functions were found, how many were successfully paired, and the names of
  skipped tests. This ensures the user always sees what happened, even in
  non-CI contexts.
- **Gate mode** (`--min-contract-coverage` or `--max-over-specification` set):
  Exit non-zero with a clear error explaining that thresholds cannot be
  evaluated because no test-target pairs were resolved.
- **Non-gate mode** (no threshold flags): Warn on stderr and exit 0, matching
  the existing `gaze crap` pattern.

### Skipped-test visibility in `quality.Assess`

Track skipped test functions (those with unresolvable targets) in the
`PackageSummary` output so the information flows through the data pipeline
rather than being lost to stderr. Add `SkippedTests int` and
`SkippedTestNames []string` fields to `PackageSummary`.

### Skipped-test summary in text output

The quality text report must include a "Skipped Tests" section when any tests
were skipped, listing their names and a brief explanation (e.g., "target
function could not be inferred — use `--target` to specify manually").

## Capabilities

### New Capabilities
- `quality-empty-result-gate`: `gaze quality` exits non-zero when threshold flags are set but zero test-target pairs were resolved, preventing silent CI pass on unanalyzable packages.
- `quality-skipped-test-summary`: Stdout output lists skipped test names with guidance when test-target pairing fails, even in non-gate mode.

### Modified Capabilities
- `gaze quality`: Writes a summary to stdout even when no quality reports are generated (previously: no stdout output).
- `gaze quality --min-contract-coverage`: Fails when no tests are paired (previously: silently passed because there were no reports to check thresholds against).
- `gaze report` quality step: Propagates skipped-test metadata through the report pipeline.
- `PackageSummary` JSON: Adds `skipped_tests` and `skipped_test_names` fields.

### Removed Capabilities
- None.

## Impact

- **Files modified**: `internal/quality/quality.go`, `internal/taxonomy/types.go`, `internal/report/schema.go`, `cmd/gaze/main.go`, `internal/aireport/runner_steps.go`
- **Exit code changes**: Users with CI pipelines using `gaze quality --min-contract-coverage` on packages where all tests have unresolvable targets will see exit 1 where they previously saw exit 0. This is the intended fix.
- **JSON output changes**: `PackageSummary` gains `skipped_tests` (int) and `skipped_test_names` ([]string) fields. Additive change; existing consumers are unaffected.
- **No new dependencies**: All changes use standard library only.
- **Not in scope**: BDD-aware target heuristics. Attempting to infer targets through Ginkgo's dynamic dispatch would create false coverage claims, violating Principle I (Accuracy). The correct mitigation is the existing `--target` flag, and the new skipped-test guidance will direct users to it.

## Constitution Alignment

Assessed against the Gaze project constitution (v1.3.0).

### I. Accuracy

**Assessment**: PASS

No new analysis heuristics are introduced. The change does not attempt to guess
BDD targets — that would risk false positives. Instead, it makes the existing
limitation visible and actionable. `TotalTests` in `PackageSummary` already
counts only paired tests; the new `SkippedTests` field makes the gap explicit
without changing any existing accuracy guarantees.

### II. Minimal Assumptions

**Assessment**: PASS

No new assumptions about the host project's test framework. The change works
identically regardless of whether the package uses stdlib testing, Ginkgo,
testify, or any other framework. The `--target` flag (existing) remains the
explicit override for when automatic inference fails.

### III. Actionable Output

**Assessment**: PASS

This is the primary principle served. The change ensures that:
- "No quality reports generated" is accompanied by a list of skipped tests and
  a pointer to `--target` as the manual override.
- Gate commands fail explicitly when thresholds cannot be evaluated, rather than
  silently passing on zero data.
- The stdout summary is present even in non-gate mode, so interactive users
  also see what happened.

### IV. Testability

**Assessment**: PASS

All changes are testable in isolation:
- `quality.Assess`: Unit test with synthetic test functions that have no
  resolvable targets — verify `SkippedTests`/`SkippedTestNames` populated.
- `runQuality`: Unit test with empty reports + threshold flags — verify exit
  code and stdout content.
- `PackageSummary` JSON: Schema validation test for new fields.
- Text output: Unit test verifying skipped-test section appears in report.

Coverage strategy: unit tests for each modified function, plus integration
test verifying end-to-end exit code behavior.
