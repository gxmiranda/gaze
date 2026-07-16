## Why

`runGoTestCoverage` in `internal/provider/goprovider/coverage.go:88` (called by `generateCoverProfile`) hardcodes `-short` when spawning `go test -coverprofile`. Any user test guarded by `testing.Short()` is skipped, producing 0% coverage for the code those tests exercise. CRAP scores are inflated accordingly, causing false CI gate failures.

The `-short` flag was added as a blunt recursion guard: when gaze runs `go test ./...` on its own source tree, self-check tests (`TestRunSelfCheck_*`) would re-invoke `runCrap` → `generateCoverProfile` → `go test ./...`, creating an infinite subprocess loop. But the fix over-corrects — it suppresses *all* `testing.Short()`-guarded tests, not just gaze's own recursive ones.

The `--coverprofile` flag already lets users bypass the problem, but the default path (no pre-generated profile) should produce accurate coverage. Closes [#106](https://github.com/unbound-force/gaze/issues/106).

## What Changes

Replace the blanket `-short` with a targeted environment-variable guard (`GAZE_COVERAGE_RUN=1`) that prevents only the specific recursion scenario. Add a `--test-short` CLI flag for users who explicitly want `-short` behavior.

## Capabilities

### New Capabilities
- `GAZE_COVERAGE_RUN` env var: Set by `runGoTestCoverage` on the spawned `go test` subprocess. Gaze's own self-check tests check this variable and skip when set, preventing recursive `go test` chains without affecting any other tests.
- `--test-short` flag on `gaze crap`, `gaze report`, `gaze self-check`: Opt-in flag that passes `-short` to the internal `go test` invocation. Defaults to `false`.

### Modified Capabilities
- `runGoTestCoverage`: No longer passes `-short` by default. Sets `GAZE_COVERAGE_RUN=1` on the subprocess environment instead. Accepts a `short` parameter for explicit opt-in.
- `GoLineCoverageProvider`: Gains a `Short bool` field threaded from CLI flags.

### Removed Capabilities
- None.

## Impact

- **Files modified**: `internal/provider/goprovider/coverage.go`, `cmd/gaze/main.go`, `cmd/gaze/main_test.go`, `internal/aireport/runner.go`, `internal/aireport/runner_steps.go`
- **Behavioral change**: `gaze crap ./...` and `gaze report ./...` (without `--coverprofile`) now run the full test suite instead of a `-short` subset. Coverage accuracy improves; run time may increase for projects with heavyweight `testing.Short()`-gated tests. When gaze analyzes its own source tree locally (without `--coverprofile`), 46 additional `testing.Short()`-gated tests will now execute, increasing self-analysis time. CI is unaffected since it uses `--coverprofile=coverage.out`.
- **Backward compatibility**: Users who relied on implicit `-short` for speed can restore it with `--test-short`. The `--coverprofile` path is unaffected.
- **Semver**: This is a bug fix (the old behavior produced inaccurate CRAP scores). Appropriate for a MINOR version bump. Migration: users who relied on implicit `-short` for speed should add `--test-short` to their invocation.
- **CI impact**: Gaze's own CI already uses `--coverprofile=coverage.out`, so this change does not affect CI run time or behavior. The self-check E2E tests already run without `-short`.

## Constitution Alignment

Assessed against the Gaze project constitution (`.specify/memory/constitution.md`).

### I. Accuracy

**Assessment**: PASS

This change directly improves accuracy: CRAP scores will reflect actual test coverage instead of artificially deflated values caused by `-short` skipping `testing.Short()`-gated tests. The fix eliminates a class of false positives (inflated CRAP scores) that eroded trust in gaze's output.

### II. Minimal Assumptions

**Assessment**: PASS

No new user requirements are introduced. The `--test-short` flag is optional with a sensible default (`false`). The env-var guard (`GAZE_COVERAGE_RUN`) is internal to gaze and does not require users to annotate or restructure their test code.

### III. Actionable Output

**Assessment**: PASS

Output formats (JSON, text) are unchanged. Coverage data is more accurate, making CRAP scores and remediation recommendations more actionable. The `--test-short` flag is discoverable via `--help`.

### IV. Testability

**Assessment**: PASS

The `GoLineCoverageProvider.Short` field is injectable in tests. The env-var guard consumer side is testable via `t.Setenv`. Command argument construction is extractable into a pure function for unit testing without subprocess execution.
