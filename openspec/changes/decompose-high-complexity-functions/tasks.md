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

## 1. Simplify `isPointerArgStore` (13 → ≤5)

All tasks modify the same file and MUST run sequentially.

- [x] 1.1 Remove unreachable branches from `isPointerArgStore`
  - Remove the `UnOp`, `FieldAddr`, and `IndexAddr` type-switch branches
    after the initial `tracesToParam(addr, param)` call (lines 310-337
    of `internal/analysis/mutation.go`). These branches duplicate logic
    that `tracesToParam` → `tracesToParamVisited` already performs
    recursively (handles `FieldAddr`, `IndexAddr`, `UnOp`, and `Phi`
    at lines 362-376).
  - The simplified function becomes: for each param in `ptrParams`,
    call `tracesToParam(addr, param)`; return the first match.
  - Add GoDoc comment noting that `tracesToParam` handles chain walking.
  - **Files**: `internal/analysis/mutation.go`

- [x] 1.2 Add unit tests for simplified `isPointerArgStore`
  - The existing `TestIsPointerArgStore` (5 subtests) provides
    regression coverage. Add additional tests to verify the simplified
    function handles all SSA instruction patterns:
    - `TestIsPointerArgStore_DirectStore` — store directly to parameter
    - `TestIsPointerArgStore_FieldAddrChain` — store via FieldAddr→param
    - `TestIsPointerArgStore_IndexAddrChain` — store via IndexAddr→param
    - `TestIsPointerArgStore_UnOpChain` — store via UnOp (dereference)→param
    - `TestIsPointerArgStore_NoMatch` — store to unrelated address
  - Tests MUST NOT use `testing.Short()` guard
  - Use real SSA construction via `ssatest` or `golang.org/x/tools`
    test patterns established in `mutation_test.go`
  - **Files**: `internal/analysis/mutation_test.go`

## 2. Decompose `detectASTReceiverMutations` (24 → ≤15)

- [x] 2.1 Extract per-node-type handlers
  - Extract `handleReceiverAssignStmt(node *ast.AssignStmt,
    receiverName string) (bool, token.Pos)` from the `*ast.AssignStmt`
    case (lines 494-508). Handles: iterate LHS expressions, find root
    ident matching receiver, verify it's a field access not bare ident.
  - Extract `handleReceiverIncDecStmt(node *ast.IncDecStmt,
    receiverName string) (bool, token.Pos)` from the `*ast.IncDecStmt`
    case (lines 509-517). Handles: check operand root ident, verify
    not bare ident.
  - Extract `handleReceiverCallExpr(node *ast.CallExpr,
    receiverName string) (bool, token.Pos)` from the `*ast.CallExpr`
    case (lines 518-534). Handles: check selector expression, verify
    method call on receiver field (not direct receiver method).
  - Replace inline cases in `detectASTReceiverMutations` with calls
    to these handlers.
  - Add GoDoc comments on all three helpers.
  - **Files**: `internal/analysis/mutation.go`

- [x] 2.2 Add unit tests for receiver mutation handlers
  - Add tests to `internal/analysis/mutation_test.go` using
    `go/parser.ParseFile` for synthetic AST construction:
    - `TestHandleReceiverAssignStmt_FieldAssign` — `recv.Field = v` → true
    - `TestHandleReceiverAssignStmt_BareAssign` — `recv = v` → false
    - `TestHandleReceiverAssignStmt_Unrelated` — `x.Field = v` → false
    - `TestHandleReceiverIncDecStmt_FieldIncrement` — `recv.count++` → true
    - `TestHandleReceiverIncDecStmt_BareIncrement` — `recv++` → false
    - `TestHandleReceiverCallExpr_FieldMethod` — `recv.f.Delete(k)` → true
    - `TestHandleReceiverCallExpr_DirectMethod` — `recv.Method()` → false
  - Tests MUST NOT use `testing.Short()` guard
  - **Files**: `internal/analysis/mutation_test.go`

## 3. Decompose `runCrap` (19 → ≤15)

All tasks modify the same file and MUST run sequentially.

- [x] 3.1 Extract baseline and gate helpers from `runCrap`
  - Extract `resolveBaselineAndCompare(baselinePath, moduleDir string,
    stderr io.Writer, rpt *crap.Report, baselineExplicit bool)
    (*crap.ComparisonResult, error)` from lines 562-571. Handles:
    resolve baseline path via `resolveBaselinePath`, load and compare
    via `loadAndCompare`.
  - Extract `writeCrapOutputAndSummary(stdout, stderr io.Writer,
    format string, rpt *crap.Report, cr *crap.ComparisonResult,
    maxCrapload, maxGazeCrapload int) error` from lines 573-584.
    Handles: write comparison or normal report, print CI summary.
    Note: threshold params forwarded to `printCISummary` for display.
  - Extract `evaluateCrapGates(rpt *crap.Report, cr *crap.ComparisonResult,
    stderr io.Writer, maxCrapload, maxGazeCrapload int) error` from
    lines 586-598. Handles: baseline comparison gate (D7 — must be
    first), then threshold gate via `checkCIThresholds`.
  - Replace inline logic in `runCrap` with calls to these helpers.
  - **Critical invariant**: `evaluateCrapGates` MUST evaluate the
    baseline gate before the threshold gate (D7 ordering).
  - Add GoDoc comments on all three helpers.
  - **Files**: `cmd/gaze/main.go`

- [x] 3.2 Add unit tests for `runCrap` helpers
  - Add tests to `cmd/gaze/main_test.go`:
    - `TestResolveBaselineAndCompare_NoBaseline` — empty path, no default
      file → nil result, nil error
    - `TestResolveBaselineAndCompare_BaselinePresent` — valid baseline
      file → non-nil ComparisonResult
    - `TestResolveBaselineAndCompare_LoadError` — invalid/corrupt file →
      error returned
    - `TestEvaluateCrapGates_BaselineRegression` — comparison with
      `Passed=false` → error before thresholds
    - `TestEvaluateCrapGates_BaselinePassThenThresholds` — comparison
      with `Passed=true` + thresholds pass → nil
    - `TestEvaluateCrapGates_ThresholdViolation` — no baseline, threshold
      exceeded → threshold error
    - `TestEvaluateCrapGates_AllPass` — nil comparison, no threshold
      violation → nil
    - `TestWriteCrapOutputAndSummary_WithComparison` — comparison result
      present → comparison report written
    - `TestWriteCrapOutputAndSummary_WithoutComparison` — nil comparison
      → normal report written
  - Tests MUST NOT use `testing.Short()` guard
  - Use the existing `crapParams` DI pattern for test setup
  - **Files**: `cmd/gaze/main_test.go`

## 4. Decompose `runQuality` (32 → ≤15)

- [x] 4.1 Extract per-package and empty-result helpers from `runQuality`
  - Extract `runQualityPerPackage(pkgPath string, p qualityParams,
    opts analysis.Options, cfg *config.Config,
    modPkgs []*packages.Package, aiMapperFn quality.AIMapperFunc)
    ([]taxonomy.QualityReport, *taxonomy.PackageSummary, error)` from
    lines 1115-1165. Handles: analyze, classify, load test package
    (skip gracefully), assess quality.
  - Extract `writeQualityEmptyResults(w io.Writer, format string,
    merged *taxonomy.PackageSummary) error` from lines 1174-1214.
    Handles: JSON empty array output, text summary with skipped test
    names (truncated at `MaxSkippedTestDisplay`), target hint.
    Does NOT evaluate thresholds (caller responsibility).
  - Replace inline logic in `runQuality` with calls to these helpers.
  - Add GoDoc comments on both helpers.
  - **Files**: `cmd/gaze/main.go`

- [x] 4.2 Add unit tests for `runQuality` helpers
  - Add tests to `cmd/gaze/main_test.go`:
    - `TestRunQualityPerPackage_Success` — valid package with tests →
      non-empty reports, non-nil summary
    - `TestRunQualityPerPackage_NoTests` — package where loadTestPackage
      fails → nil reports, nil summary, nil error (graceful skip)
    - `TestRunQualityPerPackage_ClassifyError` — classification failure →
      non-nil error wrapping classify error
    - `TestWriteQualityEmptyResults_TextFormat` — text format with
      skipped tests → writes summary, names, hint
    - `TestWriteQualityEmptyResults_JSONFormat` — JSON format → writes
      valid JSON with empty array
    - `TestWriteQualityEmptyResults_TextNoSkipped` — text format with
      0 skipped → writes summary only, no skipped section
    - `TestWriteQualityEmptyResults_TextTruncation` — text format with
      > MaxSkippedTestDisplay skipped → truncates with "... and N more"
  - Note: `runQualityPerPackage` tests use the existing `qualityParams`
    DI pattern with real but minimal package fixtures from `testdata/src/`.
    These are integration-style tests since D6 (no new DI) means the
    internal calls (`LoadAndAnalyze`, `runClassify`, etc.) are real.
    The existing `TestRunQuality_BadPackage`, `TestRunQuality_MultiPackage_SkipsNoTests`
    also provide transitive coverage.
  - Tests MUST NOT use `testing.Short()` guard
  - **Files**: `cmd/gaze/main_test.go`

## 5. Decompose `writeOneResult` (32 → ≤15) [P]

This group touches a different package and CAN run in parallel with
groups 1-4.

- [x] 5.1 Extract row-building and signal helpers from `writeOneResult`
  - Extract `writeEffectRows(effects []taxonomy.SideEffect, maxDesc int,
    showClassify bool) [][]string` from the row-building loops in both
    branches (lines 82-103 and 159-170). Consolidates the duplicate
    row-building logic. When `showClassify=true`, adds a 4th column
    for classification cell with truncation. When `showClassify=false`,
    produces 3-column rows.
  - Extract `writeVerboseSignals(w io.Writer,
    effects []taxonomy.SideEffect)` from the verbose signal loop
    (lines 131-152). Iterates effects, prints signal breakdown for
    those with non-nil Classification and non-empty Signals.
  - Replace inline logic in both branches of `writeOneResult` with
    calls to `writeEffectRows`. Call `writeVerboseSignals` after the
    classify-mode table render.
  - Add GoDoc comments on both helpers.
  - **Files**: `internal/report/text.go`

- [x] 5.2 Add unit tests for `writeOneResult` helpers
  - Add tests to `internal/report/text_test.go`:
    - `TestWriteEffectRows_ClassifyMode` — showClassify=true with
      classification → 4-column rows, correct classification cell
    - `TestWriteEffectRows_NonClassifyMode` — showClassify=false →
      3-column rows
    - `TestWriteEffectRows_DescriptionTruncation` — long description
      → truncated to maxDesc-3 + "..."
    - `TestWriteEffectRows_NilClassification` — showClassify=true,
      nil classification → "—" cell
    - `TestWriteEffectRows_ClassificationTruncation` — long classification
      label → truncated
    - `TestWriteVerboseSignals_WithSignals` — effects with signals →
      formatted output
    - `TestWriteVerboseSignals_NoSignals` — no signals → empty output
  - Tests MUST NOT use `testing.Short()` guard
  - **Files**: `internal/report/text_test.go`

## 6. Verification

- [x] 6.1 Build and test
  - Run `go build ./...` — MUST pass
  - Run `go test -race -count=1 -short ./...` — MUST pass
  - Run `golangci-lint run` — MUST pass with 0 issues
  - **Verify** complexity targets:
    - `isPointerArgStore` ≤ 5
      (`gocyclo -over 5 internal/analysis/mutation.go | grep isPointerArgStore`)
    - `detectASTReceiverMutations` ≤ 15
      (`gocyclo -over 15 internal/analysis/mutation.go | grep detectASTReceiverMutations`)
    - `runCrap` ≤ 15
      (`gocyclo -over 15 cmd/gaze/main.go | grep runCrap`)
    - `runQuality` ≤ 15
      (`gocyclo -over 15 cmd/gaze/main.go | grep runQuality`)
    - `writeOneResult` ≤ 15
      (`gocyclo -over 15 internal/report/text.go | grep writeOneResult`)
  - **Verify** all existing tests pass without modification
  - **Verify** GoDoc comments on all new helpers start with function name
  - **Verify** net test delta: ~35 new tests added, 0 removed

- [x] 6.2 CRAPload verification
  - Measure current CRAPload and GazeCRAPload before implementation
    (record in PR description as baseline)
  - Run `gaze crap ./...` and verify CRAPload does not increase
  - Run `gaze crap --max-gaze-crapload=<baseline_value> ./...` and
    verify GazeCRAPload does not increase
  - Record before/after CRAPload values in the PR description

- [x] 6.3 Constitution alignment verification
  - **Accuracy (I)**: All existing tests pass — behavior unchanged
  - **Testability (IV)**: All new tests run without `testing.Short()`
    guard — they contribute to CRAPload reduction
  - **Minimal Assumptions (II)**: N/A — no analysis behavior changed
  - **Actionable Output (III)**: N/A — no output format changed

## 7. Documentation

- [x] 7.1 Update AGENTS.md Recent Changes
  - Add entry for `decompose-high-complexity-functions` describing
    the 5-function decomposition with before/after CC values
  - **Files**: `AGENTS.md`

- [x] 7.2 Confirm no README/website updates needed
  - This is internal-only refactoring with no user-facing changes
  - No new CLI commands, flags, or output formats
  - No website issue required

<!-- spec-review: passed -->

<!-- code-review: passed -->
