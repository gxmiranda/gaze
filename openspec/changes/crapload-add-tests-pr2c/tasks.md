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

## 1. ResolvePackagePaths — ungate tests and add branch coverage

Target: `internal/loader/loader_test.go`

- [x] 1.1 Remove `testing.Short()` guards from the 5 gated
  `TestResolvePackagePaths_*` tests: `_Wildcard`, `_SinglePackage`,
  `_InvalidPattern`, `_MixedValidAndInvalid`, and one additional test
  that may reference `ResolvePackagePaths` indirectly. Verify all 6
  tests pass under `go test -race -count=1 -short ./internal/loader/...`.

- [x] 1.2 Add `TestResolvePackagePaths_NilStderrWithErrors` — call
  `ResolvePackagePaths` with an invalid pattern and `stderr` set to
  `nil`. Assert no panic, assert the invalid package is excluded from
  results. This covers the `stderr != nil` guard at line 183.

- [x] 1.3 Add `TestResolvePackagePaths_DuplicatePatterns` — call
  `ResolvePackagePaths` with the same valid package path passed twice.
  Assert the result contains the package exactly once. This covers the
  `seen[pkg.PkgPath]` dedup guard at line 176.

## 2. BuildContractCoverageFunc — add DI and synthetic tests

Target: `internal/provider/goprovider/contract.go`,
`internal/provider/goprovider/contract_internal_test.go`

- [x] 2.1 Add `buildContractCoverageFuncDeps` struct with 3 injectable
  fields: `resolvePackagePaths`, `loadConfig`, `buildEffectsSetFn`.
  Add `buildContractCoverageFuncImpl` that accepts the deps struct.
  Update `BuildContractCoverageFunc` to call `buildContractCoverageFuncImpl`
  with zero-value deps. Each nil field defaults to the real
  implementation. No caller changes.

- [x] 2.2 Add test helper `successBCCFDeps()` returning a
  `buildContractCoverageFuncDeps` with synthetic implementations:
  `resolvePackagePaths` returns `["example.com/pkg"]`,
  `loadConfig` returns `config.DefaultConfig()`,
  `buildEffectsSetFn` returns `{"example.com/pkg:DoWork": true}`.

- [x] 2.3 Add 7 tests in `contract_internal_test.go`:
  - `TestBuildContractCoverageFuncImpl_ResolveError` — injected
    `resolvePackagePaths` returns error; assert result is `(nil, nil)`
    and stderr contains warning
  - `TestBuildContractCoverageFuncImpl_EmptyPkgPaths` — injected
    `resolvePackagePaths` returns empty slice; assert result is
    `(nil, nil)`
  - `TestBuildContractCoverageFuncImpl_BothMapsEmpty` — injected deps
    produce empty coverage map and empty effects set; assert result is
    `(nil, degradedPkgs)`
  - `TestBuildContractCoverageFuncImpl_ClosureFound` — injected deps
    produce coverage map with `"pkg:Func"` entry; call returned closure
    with `("pkg", "Func")`; assert `ok=true` and correct coverage info
  - `TestBuildContractCoverageFuncImpl_ClosureNoTestCoverage` — function
    in effects set but not coverage map; assert
    `Reason="no_test_coverage"` and `ok=false`
  - `TestBuildContractCoverageFuncImpl_ClosureNoEffects` — function in
    neither set; assert `Reason="no_effects_detected"` and `ok=false`
  - `TestBuildContractCoverageFuncImpl_HappyPath` — full synthetic
    pipeline; assert closure returns correct info, degraded packages
    are passed through, and stderr contains completion message

## 3. loadStreaming — extract parser and add tests

Target: `internal/adapter/sideeffect.go`,
`internal/adapter/sideeffect_test.go`

- [x] 3.1 [P] Extract `parseSideEffectStream(scanner *bufio.Scanner)
  ([]protocol.AnalyzedFunction, error)` from `loadStreaming`. Move the
  JSONL scanning loop (lines 148–166) into the new function. Update
  `loadStreaming` to call `parseSideEffectStream` and assign the result
  to `a.cached` via `convertAnalysisResults`. Verify existing
  integration test `TestExternalSideEffectAnalyzer_Streaming` still
  passes.

- [x] 3.2 [P] Add 6 tests in `sideeffect_test.go`:
  - `TestParseSideEffectStream_ValidJSONL` — 3 valid JSONL lines;
    assert 3 functions returned with correct package/name/effects
  - `TestParseSideEffectStream_EmptyStream` — empty input; assert
    nil/empty slice and no error
  - `TestParseSideEffectStream_EmptyLinesSkipped` — valid JSONL with
    empty lines interspersed; assert only valid lines parsed
  - `TestParseSideEffectStream_MalformedJSON` — valid line 1, bad JSON
    line 2; assert error contains "malformed JSONL on line 2"
  - `TestParseSideEffectStream_LongLineTruncated` — malformed JSON
    exceeding 200 bytes; assert error content is truncated with "..."
  - `TestParseSideEffectStream_ScannerError` — use an `io.Reader` that
    returns `io.ErrUnexpectedEOF` mid-stream; assert error contains
    "reading analyze/stream response"

Note: Tasks 3.1 and 3.2 are marked `[P]` relative to group 2 tasks
(different package: `internal/adapter/` vs `internal/provider/goprovider/`).
Within group 3, task 3.2 depends on 3.1 (the function must exist before
tests can reference it).

## 4. Verification

- [x] 4.1 Run `go build ./...` — must pass with zero errors.

- [x] 4.2 Run `go test -race -count=1 -short ./...` — all packages must
  pass. Verify the ungated `ResolvePackagePaths` tests run (not skipped)
  under `-short` mode.

- [x] 4.3 Run `golangci-lint run` — must report zero issues.

- [x] 4.4 Run `gaze crap --max-crapload=38 ./...` and verify:
  - `ResolvePackagePaths` CRAP < 15
  - `BuildContractCoverageFunc` CRAP < 15
  - `loadStreaming` CRAP < 15 (or `parseSideEffectStream` if renamed)
  - CRAPload ≤ 26

- [x] 4.5 Constitution alignment: confirm Principle IV (Testability) is
  satisfied — all 3 functions are testable in isolation with synthetic
  data (or lightweight fixtures) and tests verify observable side effects.

- [x] 4.6 Update `AGENTS.md` Recent Changes section with a summary of
  this change.

- [x] 4.7 Update issue #166 with a post-Phase 2c comment documenting
  final CRAPload, functions addressed, and recommendation to close.

<!-- spec-review: passed -->
<!-- code-review: passed -->
