## Context

CRAPload is at 29/38 after Phases 1a–2b of issue #166. Three functions
have high CRAP scores that are resolvable with test coverage improvements
alone — no structural decomposition needed. Two functions
(`BuildContractCoverageFunc`, `loadStreaming`) have 0% coverage because
their tests are gated by `testing.Short()` or absent. One function
(`ResolvePackagePaths`) has 10.5% coverage because 5 of its 6 tests are
gated by `testing.Short()`.

`gaze crap` generates coverage profiles via `go test -short`, so
`-short`-guarded tests are skipped and contribute zero coverage to
CRAPload. The Dewey learning `crapload-di-coverage-20260702T081523`
documents this pattern and confirms that DI-based tests without `-short`
guards are the proven solution.

## Goals / Non-Goals

### Goals
- Reduce CRAPload from 29 to ~26 by adding test coverage to 3 functions
- Remove unnecessary `testing.Short()` guards from `ResolvePackagePaths` tests
- Add DI to `BuildContractCoverageFunc` following the established
  `contractCoverageDeps` pattern in the same file
- Extract the JSONL scanner loop from `loadStreaming` into a testable helper
- Zero behavioral change to production code paths
- Close issue #166 after this PR merges

### Non-Goals
- Decomposing any of the 3 target functions (all have complexity ≤10)
- Covering remaining `decompose` targets from issue #166 (tracked separately)
- Changing the `--max-crapload` threshold (deferred to a separate issue)
- Achieving 100% coverage (≥70% is sufficient to drop below CRAP 15)
- Modifying any exported API surface

## Decisions

### D1: Remove `testing.Short()` guards from `ResolvePackagePaths` tests

The 5 gated tests use `packages.NeedName` load mode, which resolves
package names without parsing, type-checking, or building. This is
sub-second even on cold runs. The `-short` guards were applied as a
blanket policy for any test that calls `packages.Load`, but the
lightweight `NeedName` mode does not warrant the guard.

**Alternative considered**: Adding new DI-based tests instead of
ungating existing ones. Rejected — `ResolvePackagePaths` calls
`packages.Load` directly and that call *is* the function's purpose.
There is nothing meaningful to inject. The existing tests already
cover the important branches; they just need to run during CRAPload
measurement.

**New tests**: Add 2 tests for uncovered branches:
- `TestResolvePackagePaths_NilStderrWithErrors` — passes `nil` stderr
  with an invalid pattern; verifies no panic and the function silently
  skips the bad package (covers the `stderr != nil` guard, line 183)
- `TestResolvePackagePaths_DuplicatePatterns` — passes the same valid
  pattern twice; verifies deduplication produces a single result
  (covers the `seen[pkg.PkgPath]` guard, line 176)

### D2: Deps struct for `BuildContractCoverageFunc` with nil-means-default

Add a `buildContractCoverageFuncDeps` struct with 3 injectable fields:

```go
type buildContractCoverageFuncDeps struct {
    resolvePackagePaths func([]string, string, io.Writer) ([]string, error)
    loadConfig          func(string) *config.GazeConfig
    buildEffectsSetFn   func([]string, loadAndAnalyzeFn) map[string]bool
}
```

When a field is nil, the real implementation is used. The function
signature changes to accept an optional variadic deps parameter:

```go
func BuildContractCoverageFunc(
    patterns []string,
    moduleDir string,
    stderr io.Writer,
    aiMapperFn ...quality.AIMapperFunc,
) (...)
```

becomes:

```go
func buildContractCoverageFuncImpl(
    patterns []string,
    moduleDir string,
    stderr io.Writer,
    deps buildContractCoverageFuncDeps,
    aiMapperFn ...quality.AIMapperFunc,
) (...)
```

with the public function wrapping the impl with zero-value deps.

**Rationale**: The same file already uses this pattern for
`analyzePackageCoverage` with `contractCoverageDeps`. The
`successDeps()` test helper demonstrates the pattern works.
`buildCoverageMap` already accepts `contractCoverageDeps` so it
doesn't need additional injection — only the three calls before
it need DI.

**Alternative considered**: Making `BuildContractCoverageFunc` a
method on a struct with injectable fields. Rejected — this would
change the caller API in `cmd/gaze/main.go` and `aireport/runner.go`,
which is unnecessary for a test-only improvement.

### D3: Extract `parseSideEffectStream` from `loadStreaming`

Extract the JSONL scanner loop (lines 148–166 of `sideeffect.go`)
into a standalone function:

```go
func parseSideEffectStream(scanner *bufio.Scanner) ([]protocol.AnalyzedFunction, error)
```

`loadStreaming` becomes:

```go
func (a *ExternalSideEffectAnalyzer) loadStreaming() error {
    ctx, cancel := context.WithTimeout(context.Background(), protocol.AnalysisTimeout)
    defer cancel()
    scanner, err := a.client.CallStream(ctx, protocol.MethodAnalyzeStream, ...)
    if err != nil {
        return fmt.Errorf("analyze/stream protocol call: %w", err)
    }
    funcs, err := parseSideEffectStream(scanner)
    if err != nil {
        return err
    }
    a.cached = convertAnalysisResults(funcs, a.stderr)
    return nil
}
```

Tests create a `bufio.Scanner` from a `bytes.Reader` with synthetic
JSONL content — no subprocess needed.

**Rationale**: The `protocol.Client.CallStream` method returns a
`*bufio.Scanner`, so the extraction boundary is natural. The parsing
logic (JSON unmarshalling, empty line skipping, error formatting) is
pure and stateless. This follows Principle IV (Testability) — the
extracted function has no external dependencies.

**Alternative considered**: Extending the fake analyzer binary with
error-mode flags (`--malformed-stream`, `--stream-io-error`). Rejected
— adding test-only modes to a binary that ships as a reference
implementation creates maintenance burden and doesn't test the parsing
logic in isolation.

### D4: Internal package tests

All new tests will use internal package test style (same package name)
to access unexported functions and deps structs:

- `internal/loader/loader_test.go` — already uses `package loader_test`
  (external), but existing tests are external. New tests can also be
  external since `ResolvePackagePaths` is exported.
- `internal/provider/goprovider/contract_internal_test.go` — already
  uses `package goprovider` (internal). New tests follow the same pattern.
- `internal/adapter/sideeffect_test.go` — already uses `package adapter`
  (internal). New tests follow the same pattern.

### D5: No `-short` guards on new tests

Per the Phase 1a learning, all new tests MUST run without `testing.Short()`
guards. Tests that use DI or synthetic data have no I/O dependencies that
would make them slow. Tests that call `packages.Load` with `NeedName` mode
(the ungated `ResolvePackagePaths` tests) are lightweight enough to run
in all modes.

## Risks / Trade-offs

### Risk: Ungated `ResolvePackagePaths` tests may be slow on some CI

`packages.NeedName` invokes the Go toolchain's package name resolution.
On slow CI runners or cold caches, this could add a few seconds per test.

**Mitigation**: The existing fixtures (`testdata/src/welltested`,
`testdata/src/undertested`) are minimal packages with no external imports.
If timing becomes an issue, the specific slow tests can be re-gated as a
fallback without losing the DI-based gains.

### Risk: `buildContractCoverageFuncDeps` adds a third deps struct to contract.go

The file already has `contractCoverageDeps`. Adding another struct
increases the type count.

**Mitigation**: The two structs serve different levels of the call
hierarchy — `contractCoverageDeps` injects into `analyzePackageCoverage`
(per-package analysis), while `buildContractCoverageFuncDeps` injects
into the orchestrator (pattern resolution + effects discovery). They
don't overlap.

### Trade-off: `loadStreaming` extraction changes a private method

Extracting `parseSideEffectStream` from `loadStreaming` is a small
refactor of a private method. The extraction is safe because:
- The function is unexported and has a single call site
- The extraction preserves identical error messages and return values
- The existing integration test (`TestExternalSideEffectAnalyzer_Streaming`)
  continues to exercise the full path
