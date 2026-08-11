## Context

`gaze quality` silently exits 0 with no stdout output when every test function
in a package has unresolvable targets. This is the default behavior for
Ginkgo/BDD suites and any framework using dynamic dispatch for test
registration. The proposal (proposal.md) identifies a four-link silent-skip
chain from `resolveCallee` → `InferTargets` → `Assess` → `runQuality` where
the "no targets" signal degrades to a clean exit.

Current state:
- `quality.Assess` prints warnings to `opts.Stderr` when tests have no targets,
  but `continue`s over them. The resulting `PackageSummary.TotalTests` counts
  only successfully paired tests — skipped tests are invisible.
- `runQuality` (`cmd/gaze/main.go:1166-1169`) returns `nil` (exit 0) when
  `allReports` is empty, with only a logger warning to stderr.
- `checkQualityThresholds` is never reached when reports are empty because
  `runQuality` returns before getting there.
- `gaze crap` already has a zero-result gate pattern
  (`cmd/gaze/main.go:542-552`) that this change should mirror for consistency.

## Goals / Non-Goals

### Goals
- Surface skipped test counts and names through `PackageSummary` so the
  information flows through the data pipeline (not lost to stderr).
- Print a structured summary to stdout when no quality reports are generated,
  listing how many tests were found, how many were paired, and which were
  skipped.
- Exit non-zero when threshold flags (`--min-contract-coverage`,
  `--max-over-specification`) are set but zero test-target pairs were resolved.
- Propagate skipped-test metadata through the `gaze report` pipeline.

### Non-Goals
- Implementing BDD-aware target inference heuristics. Guessing Ginkgo targets
  would create false coverage claims, violating Principle I (Accuracy).
- Changing the SSA-based target inference logic in `resolveCallee` or
  `InferTargets` — those are working correctly for their design scope.
- Adding a new `--fail-on-empty` flag. The existing threshold flags already
  provide the CI gate semantic: if you set a threshold, you expect it to be
  evaluated. Zero results should fail the gate, not silently pass.
- Modifying how `gaze crap` handles zero results — that gate already works
  correctly (spec 009 / ci-gate-integrity).

## Decisions

### D1: Track skipped tests in PackageSummary

**Decision**: Add `SkippedTests int` and `SkippedTestNames []string` fields
to `taxonomy.PackageSummary`.

**Rationale**: The skip information currently exists only as transient stderr
warnings in `quality.Assess`. By adding it to the structured output, it flows
through the entire pipeline: `quality.Assess` → `BuildPackageSummary` →
`runQuality` → `checkQualityThresholds` → `gaze report`. Consumers (humans,
CI, AI reporters) can all act on it.

**Design**:

```go
// In taxonomy.PackageSummary:
SkippedTests     int      `json:"skipped_tests"`
SkippedTestNames []string `json:"skipped_test_names,omitempty"`
```

Populated in `quality.Assess` — every time the `continue` at line 197 fires
(test with no resolved targets), increment a counter and record the test
function name. Pass these through to the returned `PackageSummary`.

The `TotalTests` field semantics remain unchanged (counts only paired tests).
A new `TotalTestFunctions` field is NOT needed — the total is
`TotalTests + SkippedTests`.

### D2: Empty-result stdout summary in runQuality

**Decision**: When `allReports` is empty, write a structured summary to
`p.stdout` before returning, listing the total test functions found, paired
count (0), and skipped test names.

**Rationale**: The current behavior (stderr logger warning only, no stdout)
violates Principle III — the user gets zero actionable output. Even in
non-gate mode, the user should see what happened and get a pointer to
`--target` as the manual override.

**Design**: In `runQuality` at the `len(allReports) == 0` check
(`cmd/gaze/main.go:1166`):

```go
if len(allReports) == 0 {
    // Collect skipped test info from summaries.
    var totalSkipped int
    var skippedNames []string
    for _, s := range allSummaries {
        totalSkipped += s.SkippedTests
        skippedNames = append(skippedNames, s.SkippedTestNames...)
    }

    // Always write summary to stdout.
    fmt.Fprintf(p.stdout, "0 of %d test functions mapped to a target\n", totalSkipped)
    if len(skippedNames) > 0 {
        fmt.Fprintln(p.stdout, "\nSkipped tests (no target resolved):")
        for _, name := range skippedNames {
            fmt.Fprintf(p.stdout, "  %s\n", name)
        }
        fmt.Fprintf(p.stdout, "\nHint: use --target=FuncName to specify the function under test manually\n")
    }

    // Gate check: if thresholds are set, this is a CI failure.
    // Note: We use `> 0` rather than the `thresholdSet` boolean pattern
    // used by `gaze crap`. This is intentional: both quality thresholds
    // have natural zero-means-disabled semantics (0% minimum coverage is
    // always met, 0 max over-specification would fail everything). The
    // `gaze crap` pattern uses `thresholdSet` because `--max-crapload=0`
    // is a meaningful threshold (no functions may exceed CRAP 0).
    if p.minContractCoverage > 0 || p.maxOverSpecification > 0 {
        return fmt.Errorf("no test-target pairs resolved — cannot evaluate thresholds (check package patterns or use --target)")
    }
    return nil
}
```

Note: The error message is the non-zero exit code mechanism (Cobra returns the
error, which triggers `os.Exit(1)` via `Execute()`). This mirrors the `gaze
crap` zero-result gate pattern. The `> 0` threshold detection is intentionally
different from `gaze crap`'s `thresholdSet` boolean — see comment above.

### D3: Propagate skipped tests through quality.Assess

**Decision**: Modify `quality.Assess` to track skipped tests in the summary
it returns, rather than only printing to stderr.

**Rationale**: The summary already carries `TotalTests`, `SSADegraded`, etc.
Adding skipped test tracking keeps the structured output self-describing.

**Design**: In the normal path (SSA available) of `Assess`, add a counter
and name accumulator before the test function loop:

```go
var skippedTests int
var skippedTestNames []string

for _, tf := range testFuncs {
    // ... existing target inference ...

    if len(targets) == 0 {
        skippedTests++
        skippedTestNames = append(skippedTestNames, tf.Name)
        if opts.Stderr != nil {
            _, _ = fmt.Fprintf(opts.Stderr,
                "warning: %s: no target function identified, skipping\n", tf.Name)
        }
        continue
    }
    // ... rest of loop ...
}
```

Then set the fields on the summary before returning:

```go
summary.SkippedTests = skippedTests
summary.SkippedTestNames = skippedTestNames
```

The degraded path (SSA unavailable) does not skip tests — it produces
partial reports for all test functions — so `SkippedTests` stays 0 there.

### D4: Update mergeSummaries for skipped test aggregation

**Decision**: The `mergeSummaries` function in `cmd/gaze/main.go` must
aggregate `SkippedTests` (sum) and `SkippedTestNames` (concatenate) across
packages.

**Design**:

```go
merged.SkippedTests += s.SkippedTests
merged.SkippedTestNames = append(merged.SkippedTestNames, s.SkippedTestNames...)
```

### D5: Update JSON Schema

**Decision**: Add `skipped_tests` and `skipped_test_names` to the quality
JSON Schema (`internal/report/schema.go`).

**Rationale**: JSON Schema validation tests verify output against the embedded
schema. New fields must be declared.

### D6: Report pipeline propagation (gaze report)

**Decision**: The report pipeline requires explicit structural changes to
propagate skipped-test metadata. The data does NOT flow automatically —
`runQualityForPackage` currently returns `([]taxonomy.QualityReport, string)`
and discards the `PackageSummary` returned by `quality.Assess`. Each
intermediary struct must be updated manually.

**Design**: The propagation chain requires changes at three levels:

1. **`runQualityForPackage`** (`internal/aireport/runner_steps.go`): Change
   return type to include skipped test count and names (either expand the
   return values or introduce a result struct). The `Assess`-returned
   `PackageSummary` contains `SkippedTests`/`SkippedTestNames` — extract
   and return them.

2. **`qualityStepResult`** (`internal/aireport/runner_steps.go`): Add
   `SkippedTests int` and `SkippedTestNames []string` fields. Populate in
   `runQualityStep` alongside the existing `AvgContractCoverage` and
   `SSADegraded` field extraction (line ~189-196).

3. **`ReportSummary`** (`internal/aireport/payload.go`): Add
   `SkippedTests int` to surface the count to AI reporters and threshold
   evaluation. Populate from `qualityStepResult` in `runProductionPipeline`.

4. **`compactSummary`** (`internal/aireport/compact.go`): Add
   `SkippedTests int` field with JSON tag. Wire it in `CompactForAI()` so
   the AI reporter payload includes skipped test data.

This follows the same manual-wiring pattern used for `SSADegraded` and
`AvgContractCoverage` propagation.

### D7: Text report skipped-test section

**Decision**: When the text report has skipped tests, add a section after the
main quality table listing them with guidance.

**Rationale**: Interactive users running `gaze quality ./...` should see which
tests were skipped even when some tests were successfully paired. The stdout
summary from D2 only fires on zero reports — this covers the partial-success
case.

**Design**: In `quality.WriteText` (`internal/quality/report.go`), check
`summary.SkippedTests > 0` and append a section:

```
Skipped Tests (3 of 15 not paired):
  TestGinkgoSuite_Login     no target function identified
  TestGinkgoSuite_Logout    no target function identified
  TestGinkgoSuite_Register  no target function identified

  Hint: use --target=FuncName to specify manually
```

## Coverage Strategy

Per Constitution Principle IV, coverage strategy for all new code:

- **quality.Assess skipped tracking** (D3): Unit test with synthetic test
  functions that have no resolvable targets (mock SSA returning empty
  candidates). Verify `summary.SkippedTests` and `summary.SkippedTestNames`
  are populated correctly.
- **runQuality empty-result gate** (D2): Unit test using `qualityParams`
  with `io.Writer` injection. Test cases: (a) empty reports + no thresholds →
  exit 0 with stdout summary, (b) empty reports + `--min-contract-coverage` →
  non-zero exit.
- **mergeSummaries** (D4): Unit test verifying skipped test aggregation across
  multiple summaries.
- **WriteText skipped section** (D7): Unit test verifying the skipped section
  appears in text output.
- **JSON Schema** (D5): Existing schema validation tests will catch missing
  fields.

## Risks / Trade-offs

### R1: Exit code change for existing CI pipelines

Users running `gaze quality --min-contract-coverage=10 ./bdd-pkg/...` will
start seeing exit 1 where they previously saw exit 0.

**Mitigation**: This is the intended fix — the previous exit 0 was a bug. The
error message is clear and actionable, pointing to `--target` as the override.
Users who don't set threshold flags are unaffected (non-gate mode stays exit 0).

### R2: SkippedTestNames could be large

A package with hundreds of test functions (e.g., generated tests) could produce
a long list of skipped names.

**Mitigation**: The text output truncates at 20 names maximum (with "... and N
more" suffix). Both the D2 empty-result stdout summary and the D7 text report
section apply the same truncation. The JSON output includes all names for
machine-readable consumption.

### R3: Partial-success case changes output

When some tests pair and some don't, the quality report now includes a skipped
section that wasn't there before. This is additive text output, not a format
break, but may surprise users who parse text output with regex.

**Mitigation**: JSON output is the stable machine-readable format. The text
format has never been guaranteed stable.
