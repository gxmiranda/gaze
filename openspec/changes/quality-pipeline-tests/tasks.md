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

  NOTE: Line references in tasks 2.2 and 3.1-3.3 are
  approximate, based on the pre-decomposition state of
  the file. Use semantic descriptions (function names,
  section roles) to locate extraction targets, as line
  numbers shift when earlier tasks modify the file.
-->

## 1. `generateSuggestion` Direct Tests

- [x] 1.1 [P] Add `TestGenerateSuggestion` table-driven test in `internal/quality/container_unwrap_internal_test.go`. Include 6 cases: `LogWrite` (assert "log output" and "implementation detail"), `StdoutWrite` (assert "stdout"), `GoroutineSpawn` (assert "goroutine lifecycle" and "concurrency detail"), `ContextCancellation` (assert "context usage"), `CallbackInvocation` (assert "callback invocation"), and a default case using `MapMutation` (assert contains both the type name and description). Each case MUST verify the description appears in the output.

## 2. `WriteText` Decomposition

- [x] 2.1 Define `qualityStyles` struct in `internal/quality/report.go` with 5 fields: `header`, `good`, `warn`, `bad`, `muted` (all `lipgloss.Style`). Add a `newQualityStyles()` constructor that returns the struct with the same style definitions currently inline in `WriteText` (lines 33-37).
- [x] 2.2 Extract 11 section helpers in `internal/quality/report.go`. Each helper takes `(w io.Writer, ..., s qualityStyles)` and handles one output section. Move lines from `WriteText` into each helper without changing output. Helpers: `writeReportHeader` (lines 44-51), `writeContractCoverage` (lines 53-64), `writeOverSpecification` (lines 66-76), `writeDetectionConfidence` (lines 78-88), `writeGapsSection` (lines 90-101), `writeDiscardedReturns` (lines 103-113), `writeSuggestionsSection` (lines 115-121), `writeAmbiguousEffects` (lines 123-131), `writeUnmappedAssertions` (lines 133-146), `writeSSADiagnostics` (lines 149-159), `writePackageSummary` (lines 161-183).
- [x] 2.3 Rewrite `WriteText` as an orchestrator: call `newQualityStyles()`, loop over reports with separator, delegate to the 11 helpers. Verify all existing `TestWriteText_*` tests pass unchanged.
- [x] 2.4 Add tests in new file `internal/quality/report_internal_test.go` (package `quality`):
  - `TestWriteContractCoverage_Thresholds`: table-driven with boundary values 49 (bad), 50 (warn), 79 (warn), 80 (good).
  - `TestWriteOverSpecification_Thresholds`: table-driven with values 0 (good), 1 (warn), 3 (warn — boundary), 4 (bad).
  - `TestWriteDetectionConfidence_Thresholds`: table-driven with values 49 (bad), 50 (warn), 69 (warn), 70 (good).
  - `TestWriteSSADiagnostics_Rendering`: verify SSA warning text, package count, and each package name appear in output. Also test the nil/non-degraded case produces no output.
  - `TestWritePackageSummary_WithWorstCoverage`: verify "Lowest coverage tests" section renders with test names and percentages. Also test nil summary and zero-tests cases produce no output.
  - `TestWriteGapsSection_WithoutHint`: verify a gap renders without a hint line when `GapHints` is shorter than `Gaps`.
  - `TestWriteDiscardedReturns_WithoutHint`: verify a discarded return renders without a hint line when `DiscardedReturnHints` is shorter than `DiscardedReturns` (mirrors `writeGapsSection` conditional hint logic).
  - `TestWriteUnmappedAssertions_WithoutReason`: verify an assertion with empty `UnmappedReason` renders without the `[reason]` suffix.
  - `TestWriteText_MultiReport`: verify a blank line separator appears between two reports.

## 3. `traceForwardDataFlow` Decomposition

- [x] 3.1 Extract `rhsReferencesAnyTracked(rhs ast.Expr, tracked map[types.Object]bool, info *types.Info) bool` in `internal/quality/mapping.go`. Move lines 998-1023 (the direct `containsObject` loop + `resolveExprRoot` fallback) into this helper.
- [x] 3.2 Extract `handleTransformationCalls(rhs ast.Expr, tracked map[types.Object]bool, info *types.Info) (dest types.Object, handled bool)` in `internal/quality/mapping.go`. Move lines 1028-1063 (the inner `ast.Inspect` closure for transformation call detection) into this helper. Return the extracted pointer destination and whether a transformation was handled.
- [x] 3.3 Extract `extractDataFlowLHS(assign *ast.AssignStmt, rhsIdx int, rhs ast.Expr, info *types.Info) types.Object` in `internal/quality/mapping.go`. Move lines 1069-1090 (the `isDataExtraction` gate + LHS ident extraction) into this helper. Return the extracted object or nil.
- [x] 3.4 Rewrite `traceForwardDataFlow` to call the 3 helpers sequentially per RHS element. Verify all existing `TestTraceForwardDataFlow_*` tests pass unchanged.
- [x] 3.5 Add tests in `internal/quality/container_unwrap_internal_test.go`:
  - `TestRhsReferencesAnyTracked_DirectMatch`: verify direct `containsObject` path returns true.
  - `TestRhsReferencesAnyTracked_ResolveExprRootFallback`: verify `resolveExprRoot` path returns true for `result.Field.SubField` where `result` is tracked.
  - `TestRhsReferencesAnyTracked_NoMatch`: verify false when no tracked variable is referenced.
  - `TestHandleTransformationCalls_JsonUnmarshal`: verify pointer destination extraction for `json.Unmarshal(data, &target)` pattern where `data` is tracked. Use `parseAndTypeCheck` with synthetic source.
  - `TestHandleTransformationCalls_NoTrackedArg`: verify `handled=false` when no tracked variable flows into the call arguments.
  - `TestHandleTransformationCalls_NonTransform`: verify `handled=false` for a regular function call that is not a transformation.
  - `TestExtractDataFlowLHS_FieldAccess`: verify LHS extraction for `x := result.Field` (data extraction).
  - `TestExtractDataFlowLHS_MethodCall`: verify nil return for `got := s.Get("key")` (not data extraction).
  - `TestTraceForwardDataFlow_MultiIteration`: verify chain `a := result.Data; b := a.Items[0]; c := b.Name` tracks all three variables across multiple iterations.

## 4. Verification

- [x] 4.1 Run `go test -race -count=1 -short ./internal/quality/...` — all tests MUST pass.
- [x] 4.2 Run `golangci-lint run` — zero issues.
- [x] 4.3 Verify existing `TestWriteText_*` tests produce identical output (no behavioral change).
- [x] 4.4 Verify existing `TestTraceForwardDataFlow_*` tests pass unchanged.

<!-- spec-review: passed -->
<!-- code-review: passed -->
