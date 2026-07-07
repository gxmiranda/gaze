## Why

`loader.ResolvePackagePaths` does not check `pkg.Errors` on packages returned by `go/packages.Load`. When a pattern does not resolve to a real package, `go/packages` returns a synthetic `*packages.Package` with `PkgPath` set to the input pattern and `Errors` populated. Because the filter only rejects empty/duplicate/`_test` paths, the phantom path is returned as valid.

This violates the function's contract ("resolves package patterns to individual fully-qualified package paths") and causes a failing unit test on `main`. For a test-quality tool, shipping with red tests is a credibility problem (Constitution Principle IV).

Ref: https://github.com/unbound-force/gaze/issues/104

## What Changes

Add `pkg.Errors` checking to `loader.ResolvePackagePaths` so that packages with load errors are skipped rather than returned as valid paths. Log warnings for skipped packages. Add a test for the invalid-pattern case directly in `internal/loader/loader_test.go`.

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `loader.ResolvePackagePaths`: Now filters out packages with load errors (phantom paths) and logs warnings via a new `io.Writer` parameter for the caller to observe diagnostics.

### Removed Capabilities

- None

## Impact

- **`internal/loader/loader.go`**: `ResolvePackagePaths` gains `pkg.Errors` filtering. Signature changes to accept `io.Writer` for warning output.
- **`internal/loader/loader_test.go`**: New test for invalid/nonexistent pattern.
- **`internal/crap/contract.go`**: Caller updated to pass `stderr` to `ResolvePackagePaths`.
- **`internal/aireport/runner_steps.go`**: Two call sites updated. `runQualityStep` (line 90) passes its existing `stderr` parameter. `runClassifyStep` (line 192) passes `nil` since it lacks an `io.Writer` parameter -- warnings are discarded for classify-step resolution, which is acceptable per D4.
- **`cmd/gaze/main.go`**: Two callers updated -- `runAnalyze` (line 220) and `runQuality` (line 1069), both pass their existing `p.stderr` parameter.
- Fixes failing `TestBuildContractCoverageFunc_InvalidPattern` test on `main`.
- Closes #104.

## Constitution Alignment

Assessed against the Gaze project constitution.

### I. Accuracy

**Assessment**: PASS

This fix improves accuracy by preventing phantom package paths from entering the analysis pipeline. Currently, nonexistent patterns produce paths that silently fail downstream, dropping GazeCRAP/quality contributions without a clear error. After this fix, invalid patterns are detected early and surfaced as warnings.

### II. Minimal Assumptions

**Assessment**: PASS

The fix uses `pkg.Errors` -- the standard `go/packages` mechanism for reporting per-package load failures. It does not introduce new assumptions about project structure or coding style. The `go/packages` API documents that `Load` returns `err == nil` even when individual packages fail; checking `pkg.Errors` is the expected usage pattern.

### III. Actionable Output

**Assessment**: PASS

Warning messages for skipped packages give users clear, actionable feedback: the specific pattern that failed and the specific error from `go/packages`. This replaces silent failure downstream with early, explicit diagnostics.

### IV. Testability

**Assessment**: PASS

The fix directly restores a currently-failing test. The `io.Writer` parameter enables test-time capture of warning output without global state. A new dedicated test in `loader_test.go` verifies the invalid-pattern filtering behavior in isolation.
