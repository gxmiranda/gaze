## Context

`runGoTestCoverage` in `internal/provider/goprovider/coverage.go:88` hardcodes `-short` when spawning `go test -coverprofile`. This was added to prevent infinite recursion when gaze analyzes its own source tree: `TestRunSelfCheck_*` tests call `runCrap` → `generateCoverProfile` → `go test ./...`, which would re-execute those same tests. The `-short` flag causes all `testing.Short()`-gated tests to skip, zeroing their coverage contribution and inflating CRAP scores for any project that uses this standard Go idiom.

The `--coverprofile` flag already bypasses `generateCoverProfile` entirely, but the default path (no pre-generated profile) is the most common usage and should produce accurate results.

## Goals / Non-Goals

### Goals
- Remove `-short` from the default `go test` invocation so all tests contribute to coverage
- Prevent the specific gaze self-check recursion scenario without penalizing user tests
- Provide opt-in `--test-short` flag for users who want faster coverage generation at the cost of accuracy
- Thread the `short` option through `GoLineCoverageProvider` for testability (Constitution IV)

### Non-Goals
- General `--test-flags` passthrough (too broad for this fix; can be added later)
- Changing CI workflow behavior (CI already uses `--coverprofile=coverage.out`)
- Changing how `--coverprofile` works (it already bypasses the problem)

## Decisions

### D1: Environment variable guard over `-short`

**Decision**: Set `GAZE_COVERAGE_RUN=1` on the `go test` subprocess environment. Replace the `testing.Short()` guard in gaze's self-check tests with an env-var check.

**Rationale**: The recursion risk is narrow — it only occurs when gaze runs `go test ./...` on its own module. An env var is a precise, targeted guard that prevents exactly this scenario. The subprocess inherits the env var, so any test that checks it will skip, while all other `testing.Short()`-gated tests run normally. This aligns with Constitution II (Composability): gaze users are not affected by gaze's internal recursion concern.

**Alternatives considered**:
- *Keep `-short` as default, add `--no-short` flag*: Inverts the correct default. Users who don't know about the flag get degraded coverage.
- *Build tag*: Requires users to understand Go build tags. Too invasive.
- *Test binary detection*: Fragile; Go test binaries don't have a reliable signature.

### D2: `--test-short` flag defaults to `false`

**Decision**: Add `--test-short` to `gaze crap`, `gaze report`, and `gaze self-check`. Defaults to `false`.

**Rationale**: The default behavior should produce the most accurate results. Users with large test suites gated behind `testing.Short()` who want faster coverage runs can opt in. This reverses the current implicit `-short` behavior, which is the root cause of #106.

### D3: Thread `Short` through `GoLineCoverageProvider` via struct field

**Decision**: Add `Short bool` field to `GoLineCoverageProvider`. Callers set the field after construction (`p.Short = true`) rather than changing the `NewLineCoverageProvider` constructor signature. The `Short` value is threaded through `generateCoverProfile` → `runGoTestCoverage` to conditionally include `-short` in args.

**Rationale**: The provider struct is the natural place for this configuration — it's already the layer between CLI flags and subprocess execution. This keeps the `crap.LineCoverageProvider` interface unchanged (no signature changes to `Coverage`) and supports injection in tests (Constitution IV: Testability). Using a struct field rather than a constructor parameter avoids breaking existing callers (3 test call sites in `goprovider_test.go`, plus production callers).

**Threading path (report pipeline)**: `RunnerOptions.TestShort` → `runProductionPipeline` (new `testShort bool` parameter) → `pipelineStepFuncs.crapStep` (updated type signature with `short bool` parameter) → `runCRAPStep` (new `short bool` parameter) → `GoLineCoverageProvider{Stderr: stderr, Short: short}`.

### D4: Self-check tests use env-var guard, keep `testing.Short()` as secondary

**Decision**: The E2E self-check tests in `cmd/gaze/main_test.go` (`TestRunSelfCheck_TextFormat` at line 1432 and `TestRunSelfCheck_JSONFormat` at line 1450) add `if os.Getenv("GAZE_COVERAGE_RUN") != "" { t.Skip(...) }` as the *first* guard, before the existing `testing.Short()` check. The check uses `!= ""` (not `== "1"`) intentionally — any non-empty value is treated as "gaze is generating coverage" for robustness.

**Rationale**: The env-var guard prevents recursion. The `testing.Short()` guard remains as a secondary skip for when users (or CI) explicitly run with `-short` for speed. Both guards are documented in the skip message. Only the E2E test variants need the guard — unit test variants that use injected `runCrapFunc` do not trigger recursion.

### D5: Extract command construction for testability

**Decision**: Extract a `buildGoTestCmd(profilePath, moduleDir string, patterns []string, short bool) *exec.Cmd` helper from `runGoTestCoverage`. This function constructs the `*exec.Cmd` with args and env vars, returning it for inspection in unit tests without executing.

**Rationale**: The current `runGoTestCoverage` constructs and executes the command inline, making it impossible to unit test arg/env construction without spawning a real `go test` subprocess. Extracting the construction into a pure function enables direct inspection of `cmd.Args` and `cmd.Env` in tests (Constitution IV: Testability).

## Coverage Strategy

- **Unit tests (provider)**: Test `buildGoTestCmd` helper — verify `-short` inclusion/exclusion, `GAZE_COVERAGE_RUN` in env, arg ordering. Test `GoLineCoverageProvider.Short` field threading. Test `Coverage` with `Short=true` + valid `coverProfile` does NOT spawn subprocess.
- **Unit tests (CLI wiring)**: Verify `--test-short` threads to provider `Short` field in all three commands using the existing `runXxx(params)` injection pattern with captured options.
- **Unit tests (pipeline)**: Verify `RunnerOptions.TestShort` reaches `GoLineCoverageProvider.Short` through the pipeline using `pipelineStepFuncs` injection.
- **Integration (existing)**: The E2E self-check tests (`TestRunSelfCheck_TextFormat/JSONFormat`) exercise the full `generateCoverProfile` → `runGoTestCoverage` path and will validate the env-var recursion guard. No new E2E tests needed.
- **No new E2E tests**: The existing self-check tests are the recursion scenario this fix addresses.

## Risks / Trade-offs

### R1: Longer default `go test` run time

Without `-short`, `go test` may take longer for projects with heavyweight `testing.Short()`-gated tests. The `--test-short` flag mitigates this for users who prefer speed over accuracy.

### R2: Env var leaks into user test environment

`GAZE_COVERAGE_RUN=1` is set on the subprocess environment. If a user test reads this env var for its own purposes (unlikely — it's gaze-specific), it could cause unexpected behavior. The variable name is namespaced with `GAZE_` to minimize collision risk.

### R3: Self-check recursion if env var is cleared

If a test explicitly clears `GAZE_COVERAGE_RUN` from its environment, the recursion guard breaks. This is an intentional-adversarial scenario, not a realistic risk. The `testing.Short()` fallback remains as defense in depth.
