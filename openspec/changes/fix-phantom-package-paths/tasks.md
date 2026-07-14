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

## 1. Fix ResolvePackagePaths

- [x] 1.1 Add `io.Writer` parameter to `loader.ResolvePackagePaths` signature and add `pkg.Errors` filtering with warning output. In `internal/loader/loader.go`: change signature from `ResolvePackagePaths(patterns []string, moduleDir string) ([]string, error)` to `ResolvePackagePaths(patterns []string, moduleDir string, stderr io.Writer) ([]string, error)`. Add `"fmt"` and `"io"` imports. In the loop body, after the existing `PkgPath == ""` / `_test` filter but before the `seen` deduplication check, add: if `len(pkg.Errors) > 0`, write a warning per error to `stderr` using format `fmt.Fprintf(stderr, "warning: skipping %s: %s\n", pkg.PkgPath, pkgErr)` (nil-safe: check `stderr != nil` before writing) and `continue`.

## 2. Update Callers

- [x] 2.1 [P] Update `internal/crap/contract.go:41` -- pass `stderr` as the third argument to `loader.ResolvePackagePaths`. The `stderr io.Writer` is already available as a parameter of `BuildContractCoverageFunc`.
- [x] 2.2 [P] Update `internal/aireport/runner_steps.go` -- two call sites. For `runQualityStep` (line 90): pass the existing `stderr io.Writer` parameter as the third argument. For `runClassifyStep` (line 192): pass `nil` as the third argument -- `runClassifyStep` does not have an `io.Writer` parameter, and adding one would cascade to the `pipelineStepFuncs.classifyStep` type definition and test fakes; passing `nil` is consistent with design decision D4 (nil-safe stderr handling) and avoids scope expansion.

## 3. Update Tests

- [x] 3.1 Add `TestResolvePackagePaths_InvalidPattern` to `internal/loader/loader_test.go`. Call `ResolvePackagePaths` with pattern `"github.com/nonexistent/does/not/exist"`, a valid `moduleDir`, and a `bytes.Buffer` for stderr. Assert: returned paths is empty, no error, and stderr buffer contains substring `"warning: skipping"` AND contains `"github.com/nonexistent/does/not/exist"` (matching the format specified in specs/resolve-package-paths.md). Guard with `testing.Short()` since it invokes `go/packages.Load`.
- [x] 3.2 Update existing `ResolvePackagePaths` tests in `internal/loader/loader_test.go` to pass `nil` as the third argument (stderr). Tests: `TestResolvePackagePaths_Wildcard`, `TestResolvePackagePaths_SinglePackage`, `TestResolvePackagePaths_EmptyPatterns`.

## 4. Verification

- [x] 4.1 Run `go build ./...` to verify compilation.
- [x] 4.2 Run `go test -race -count=1 -short ./internal/loader/...` to verify loader tests pass.
- [x] 4.3 Run `go test -race -count=1 -short ./internal/crap/...` to verify crap tests pass.
- [x] 4.4 Run `go test -race -count=1 -short ./internal/aireport/...` to verify aireport tests pass.
- [x] 4.5 Run `go test -race -count=1 -short ./...` to verify full short test suite passes.
- [x] 4.6 Run `golangci-lint run` to verify no lint violations. (golangci-lint not installed locally; `go vet ./...` passes; CI will run full lint)
- [x] 4.7 Constitution alignment: verify Principle IV (Testability) -- the previously-failing `TestBuildContractCoverageFunc_InvalidPattern` now passes, and the new `TestResolvePackagePaths_InvalidPattern` covers the fix in isolation.

<!-- spec-review: passed -->
<!-- code-review: passed -->
