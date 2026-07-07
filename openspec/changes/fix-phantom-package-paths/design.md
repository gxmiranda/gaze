## Context

`loader.ResolvePackagePaths` (`internal/loader/loader.go:154-177`) resolves package patterns to fully-qualified paths using `go/packages.Load`. The `go/packages` API returns `err == nil` at the call level even when individual packages fail to load -- per-package errors are stored in `pkg.Errors`. The current implementation never checks `pkg.Errors`, so synthetic packages for nonexistent patterns pass through the filter and are returned as valid paths.

This was identified in issue #104 with a confirmed failing test on `main`. The issue comment (#104, council analysis) provides a thorough critique of the naive fix, identifying five gaps that this design addresses.

## Goals / Non-Goals

### Goals

- Filter out packages with load errors in `ResolvePackagePaths`, preventing phantom paths from entering downstream analysis
- Surface warnings for skipped packages so users understand why a pattern was dropped
- Add a test for the invalid-pattern case in `internal/loader/loader_test.go`
- Fix the currently-failing `TestBuildContractCoverageFunc_InvalidPattern` test
- Close #104

### Non-Goals

- Consolidating the duplicate `resolvePackagePaths` copies across packages (tracked in #139; both copies now delegate to `loader.ResolvePackagePaths`, so fixing the shared function fixes all callers)
- Differentiating error kinds (`ListError` vs `TypeError` vs `ParseError`) -- the council comment suggests this but it adds complexity for marginal benefit in the `ResolvePackagePaths` context, which uses `NeedName` mode only (type errors are not produced at this load level)
- Checking `pkg.Name == ""` as a complementary defense -- while noted in the critique, `pkg.Errors` is the canonical signal; `Name == ""` is an implementation detail of `go/packages` that could change

## Decisions

### D1: Add `io.Writer` parameter for warning output

`ResolvePackagePaths` currently has no way to surface diagnostics. Rather than returning a `[]error` or using a global logger, adding an `io.Writer` parameter follows the project's established pattern (see `BuildContractCoverageFunc`, `analyzePackageCoverage`, `quality.Options.Stderr`). Four of five callers already have an `io.Writer` available (`BuildContractCoverageFunc`, `runQualityStep`, `runAnalyze`, `runQuality`). The fifth (`runClassifyStep`) does not; it passes `nil` per D4 to avoid cascading signature changes.

**Signature change**: `ResolvePackagePaths(patterns []string, moduleDir string) ([]string, error)` becomes `ResolvePackagePaths(patterns []string, moduleDir string, stderr io.Writer) ([]string, error)`.

This aligns with Constitution Principle IV (Testability) -- tests can capture warnings via `bytes.Buffer`.

### D2: Skip all errored packages, warn per-error

For each package with `len(pkg.Errors) > 0`, emit one warning line per error to `stderr` and skip the package. This is consistent with `LoadModule` (lines 113-120) which also silently excludes errored packages, though `LoadModule` does not log warnings.

The council critique notes that silently dropping packages is "arguably worse" than the current behavior. The warning output addresses this -- the package is dropped from the result but the user is informed why.

### D3: Do not return an error for partial resolution

If some patterns resolve and others don't, return the valid paths with warnings for the invalid ones. Only the existing behavior of returning `(nil, nil)` for completely empty input is preserved. This matches the best-effort semantics already documented in `BuildContractCoverageFunc`: "if the quality pipeline fails for any package, those packages are silently skipped."

If ALL patterns fail to resolve (non-empty input producing zero valid paths), the function returns `(nil, nil)` -- not an error. The caller (`BuildContractCoverageFunc`) already handles this case correctly by returning `nil`. Callers cannot distinguish "no input" from "all input invalid" via return values alone; the warning output to `stderr` is the only differentiator.

### D4: Nil-safe stderr handling

If `stderr` is nil, warnings are silently discarded. This avoids forcing callers that don't care about diagnostics to pass a writer. The function checks `stderr != nil` before writing.

## Coverage Strategy

Unit test in `internal/loader/loader_test.go` (`TestResolvePackagePaths_InvalidPattern`) covers the new `pkg.Errors` filtering logic directly. Existing `TestBuildContractCoverageFunc_InvalidPattern` in `internal/crap/contract_test.go` serves as integration-level regression guard through the `BuildContractCoverageFunc` -> `ResolvePackagePaths` call chain. Both guarded by `testing.Short()` since they invoke `go/packages.Load`. No e2e test needed -- the fix is internal to the loader package.

## Risks / Trade-offs

### R1: Signature change is breaking for callers

Adding `io.Writer` changes the exported function signature. All known callers are within this module (`internal/crap/contract.go`, `internal/aireport/runner_steps.go`), so the blast radius is contained. External consumers cannot import `internal/` packages.

### R2: NeedName mode limits error kinds

`ResolvePackagePaths` uses `packages.NeedName` -- a lightweight mode. At this load level, errors are primarily `ListError` (package not found). The blanket `len(pkg.Errors) > 0` check is safe here because `TypeError` and `ParseError` are not produced without `NeedTypes`/`NeedSyntax`. If the load mode were ever expanded, this decision would need revisiting.

### R3: No error aggregation

The function does not aggregate errors into a returned `error`. This is intentional (D3) but means a caller passing `stderr=nil` would silently get fewer results than expected. Callers that need strict validation should check result count against input count.
