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

## 1. Decompose `isTransformationCall` (26 → ≤8)

All tasks modify the same files and MUST run sequentially.

- [x] 1.1 Extract `isByteLikeParam` + `isPointerDestParam` helpers
  - Extract `isByteLikeParam(typ types.Type) bool` from the loop body
    in `isTransformationCall` (lines 772-814). Consolidates 3 patterns:
    - `[]byte`: `types.Slice` with `types.Basic` byte element
    - `string`: `types.Basic` with `types.String` kind
    - `io.Reader`: interface with `Read` method via `types.NewMethodSet`
  - Extract `isPointerDestParam(typ types.Type) bool` from the loop body
    (lines 790-822). Consolidates 2 patterns:
    - `*T`: `types.Pointer`
    - Empty interface: `types.Interface` with zero methods
  - Replace inline logic in `isTransformationCall` with calls to
    `isByteLikeParam` and `isPointerDestParam`
  - Add GoDoc comments on both helpers
  - **Files**: `internal/quality/mapping.go`

- [x] 1.2 Add gap tests for `isTransformationCall`
  - Add 3 tests to `internal/quality/container_unwrap_internal_test.go`:
    - `TestIsTransformationCall_IoReaderAndPointer` — function with
      `io.Reader` + `*int` parameters → positive, correct indices
    - `TestIsTransformationCall_EmptyInterfaceAsPointerDest` — function
      with `[]byte` + `interface{}` parameters → positive
    - `TestIsTransformationCall_PointerBeforeByteSlice` — function with
      `*int` before `[]byte` → positive, correct indices (verifies
      parameter ordering doesn't matter)
  - Add 7 tests for the extracted helpers:
    - `TestIsByteLikeParam_ByteSlice` — `[]byte` → true
    - `TestIsByteLikeParam_String` — `string` → true
    - `TestIsByteLikeParam_IoReader` — interface with `Read` method → true
    - `TestIsByteLikeParam_NonReadInterface` — interface without `Read` → false
    - `TestIsByteLikeParam_Int` — `int` → false
    - `TestIsPointerDestParam_Pointer` — `*int` → true
    - `TestIsPointerDestParam_EmptyInterface` — `interface{}` → true
    - `TestIsPointerDestParam_Int` — `int` → false
  - Tests MUST NOT use `testing.Short()` guard
  - Use the existing `parseAndTypeCheck` helper for synthetic AST
  - **Files**: `internal/quality/container_unwrap_internal_test.go`

## 2. Decompose `matchAssertionToEffect` (25 → ≤10)

- [x] 2.1 Extract `matchDirect` + `matchIndirectRoot` helpers
  - Extract `matchDirect(site, objToEffectID, effectMap, info,
    helperBridge) *taxonomy.AssertionMapping` from the first
    `ast.Inspect` closure (lines 1212-1269). Handles:
    - Direct identity matching at confidence 75
    - Helper bridge matching at confidence 70
    - Skips nil/true/false literal identifiers
  - Extract `matchIndirectRoot(site, objToEffectID, effectMap,
    info) *taxonomy.AssertionMapping` from the second `ast.Inspect`
    closure (lines 1279-1324). Handles:
    - SelectorExpr/IndexExpr/CallExpr detection
    - `resolveExprRoot` unwinding
    - Indirect matching at confidence 65
  - Replace inline closures in `matchAssertionToEffect` with:
    ```
    if m := matchDirect(...); m != nil { return m }
    return matchIndirectRoot(...)
    ```
  - Add GoDoc comments on both helpers
  - **Files**: `internal/quality/mapping.go`

- [x] 2.2 Add unit tests for `matchAssertionToEffect` helpers
  - Add tests to `internal/quality/container_unwrap_internal_test.go`
    using `parseAndTypeCheck`:
    - `TestMatchDirect_IdentityMatch` — expression contains identifier
      in `objToEffectID` → mapping with confidence 75
    - `TestMatchDirect_HelperBridgeMatch` — identifier maps through
      `helperBridge` → mapping with confidence 70
    - `TestMatchDirect_NoMatch` — no identifiers in maps → nil
    - `TestMatchDirect_NilExpr` — nil expression → nil
    - `TestMatchIndirectRoot_SelectorMatch` — `result.Name` where
      `result` is in `objToEffectID` → mapping with confidence 65
    - `TestMatchIndirectRoot_NoComposite` — simple identifier only → nil
  - Tests MUST NOT use `testing.Short()` guard
  - **Files**: `internal/quality/container_unwrap_internal_test.go`

## 3. Decompose `matchContainerUnwrap` (50 → ≤12)

- [x] 3.1 Extract `collectTrackedVars` + `matchTrackedInExpr` helpers
  - Extract `collectTrackedVars(objToEffectID, returnEffectID)
    map[types.Object]bool` from lines 974-982.
    Pure function: filters `objToEffectID` entries by `returnEffectID`.
  - Extract `matchTrackedInExpr(expr, tracked, info) bool` from
    lines 1106-1151. Walks expression tree via `ast.Inspect`, checks
    both direct identity and `resolveExprRoot` fallback.
  - Replace inline logic in `matchContainerUnwrap` with calls to
    these helpers
  - Add GoDoc comments
  - **Files**: `internal/quality/mapping.go`

- [x] 3.2 Extract `traceForwardDataFlow` helper
  - Extract `traceForwardDataFlow(tracked, testPkg)
    map[types.Object]bool` from lines 987-1104.
    Multi-iteration forward data-flow tracer:
    - Walks AST assignment statements
    - Checks if RHS references tracked variables (via `containsObject`
      and `resolveExprRoot`)
    - Detects transformation calls via `isTransformationCall`
    - Extracts pointer destinations via `extractPointerDest`
    - Gates non-transformation assignments via `isDataExtraction`
    - Iterates until convergence or `maxContainerChainDepth`
  - Replace inline logic in `matchContainerUnwrap` with:
    ```
    tracked = traceForwardDataFlow(tracked, testPkg)
    ```
  - Add GoDoc comment
  - **Files**: `internal/quality/mapping.go`

- [x] 3.3 Add unit tests for `matchContainerUnwrap` helpers
  - Add tests to `internal/quality/container_unwrap_internal_test.go`:
    - `TestCollectTrackedVars_MultipleMatches` — 3 entries, 2 match
      returnEffectID → 2 in result
    - `TestCollectTrackedVars_NoMatches` — no entries match → empty map
    - `TestTraceForwardDataFlow_SimpleChain` — assignment `y := x.Field`
      with `x` tracked → `y` added. Construct `packages.Package` with
      `Syntax: []*ast.File{file}` and `TypesInfo: info` from
      `parseAndTypeCheck`
    - `TestTraceForwardDataFlow_EmptyTracked` — empty tracked set →
      returns empty
    - `TestTraceForwardDataFlow_NonDataExtraction` — method call
      `got := s.Get("key")` where `s` is tracked → `got` NOT added
      (gated by `isDataExtraction`)
    - `TestMatchTrackedInExpr_DirectMatch` — identifier in tracked → true
    - `TestMatchTrackedInExpr_RootResolution` — `tracked.Field` → true
    - `TestMatchTrackedInExpr_NoMatch` — unrelated identifier → false
  - Tests MUST NOT use `testing.Short()` guard
  - **Files**: `internal/quality/container_unwrap_internal_test.go`

## 4. Verification

- [x] 4.1 Build and test
  - Run `go build ./...` — MUST pass
  - Run `go test -race -count=1 -short ./...` — MUST pass
  - Run `golangci-lint run` — MUST pass with 0 issues
  - **Verify** complexity targets:
    - `isTransformationCall` ≤ 8
    - `matchAssertionToEffect` ≤ 10
    - `matchContainerUnwrap` ≤ 12
    (check via `gocyclo -over 12 internal/quality/mapping.go`)
  - **Verify** mapping accuracy ratchet:
    `go test -race -count=1 -run TestSC003_MappingAccuracy ./internal/quality/...`
    — MUST pass at 85.0% floor
  - **Verify** GoDoc comments on all new helpers start with function name
  - **Verify** net test delta: ~20 new tests added, 0 removed

- [x] 4.2 Constitution alignment verification
  - **Accuracy (I)**: Confirm `TestSC003_MappingAccuracy` passes at
    85.0% — mapping behavior unchanged
  - **Testability (IV)**: Confirm all new tests run without
    `testing.Short()` guard — they contribute to CRAPload
  - **Minimal Assumptions (II)**: N/A — no analysis behavior changed
  - **Actionable Output (III)**: N/A — no output format changed

## 5. Documentation

- [x] 5.1 Update AGENTS.md Recent Changes
  - Add entry for `crapload-decompose-pr2b` describing the decomposition
  - **Files**: `AGENTS.md`

- [x] 5.2 Confirm no README/website updates needed
  - This is internal-only refactoring with no user-facing changes
  - No new CLI commands, flags, or output formats
  - No website issue required

<!-- spec-review: passed -->

<!-- code-review: passed -->
