## ADDED Requirements

### Requirement: Environment variable recursion guard

`runGoTestCoverage` MUST set `GAZE_COVERAGE_RUN=1` in the subprocess environment of the spawned `go test` process. This variable MUST be inherited by all child processes of `go test`.

#### Scenario: Env var is set on subprocess
- **GIVEN** `GoLineCoverageProvider.Coverage` is called with an empty `coverProfile`
- **WHEN** `runGoTestCoverage` spawns `go test -coverprofile=...`
- **THEN** the subprocess environment SHALL contain exactly one `GAZE_COVERAGE_RUN=1` entry

#### Scenario: Pre-existing env var does not duplicate
- **GIVEN** `GAZE_COVERAGE_RUN` is already set in the parent process environment
- **WHEN** `runGoTestCoverage` spawns `go test`
- **THEN** the subprocess environment SHALL contain exactly one `GAZE_COVERAGE_RUN=1` entry (not duplicated)

#### Scenario: Self-check test skips under env var
- **GIVEN** `GAZE_COVERAGE_RUN` is set in the environment
- **WHEN** `TestRunSelfCheck_TextFormat` or `TestRunSelfCheck_JSONFormat` is executed
- **THEN** the test MUST skip with a message referencing the recursion guard

#### Scenario: Normal tests unaffected by env var
- **GIVEN** `GAZE_COVERAGE_RUN` is set in the environment
- **WHEN** a user test that does not check `GAZE_COVERAGE_RUN` is executed
- **THEN** the test MUST run normally and contribute to coverage

### Requirement: `--test-short` CLI flag

`gaze crap`, `gaze report`, and `gaze self-check` MUST accept a `--test-short` flag that passes `-short` to the internal `go test` invocation when generating coverage.

The flag MUST default to `false`.

#### Scenario: Default behavior omits `-short`
- **GIVEN** `gaze crap ./...` is invoked without `--test-short`
- **WHEN** coverage is generated internally (no `--coverprofile`)
- **THEN** the `go test` subprocess MUST NOT include `-short` in its arguments

#### Scenario: Explicit `--test-short` passes `-short`
- **GIVEN** `gaze crap --test-short ./...` is invoked
- **WHEN** coverage is generated internally
- **THEN** the `go test` subprocess MUST include `-short` in its arguments

#### Scenario: `--test-short` ignored when `--coverprofile` provided
- **GIVEN** `gaze crap --test-short --coverprofile=cover.out ./...` is invoked
- **WHEN** coverage is loaded from the pre-generated profile
- **THEN** `--test-short` SHALL have no effect (no `go test` is spawned)

### Requirement: `GoLineCoverageProvider.Short` field

`GoLineCoverageProvider` MUST expose a `Short bool` field that controls whether `-short` is passed to the internal `go test` invocation. The field MUST default to `false`.

#### Scenario: Provider with `Short=true`
- **GIVEN** a `GoLineCoverageProvider` with `Short: true`
- **WHEN** `Coverage` is called with an empty `coverProfile`
- **THEN** the spawned `go test` MUST include `-short`

#### Scenario: Provider with `Short=false` (default)
- **GIVEN** a `GoLineCoverageProvider` with `Short: false`
- **WHEN** `Coverage` is called with an empty `coverProfile`
- **THEN** the spawned `go test` MUST NOT include `-short`

## MODIFIED Requirements

### Requirement: Coverage generation accuracy

`runGoTestCoverage` MUST NOT pass `-short` to `go test` by default. Coverage MUST reflect the project's full test suite unless the user explicitly opts in to `-short` via `--test-short`.

Previously: `-short` was hardcoded unconditionally, causing `testing.Short()`-gated tests to skip and report 0% coverage for the code they exercise.

### Requirement: Self-check test skip guards

The E2E self-check tests `TestRunSelfCheck_TextFormat` (line 1432) and `TestRunSelfCheck_JSONFormat` (line 1450) in `cmd/gaze/main_test.go` MUST check `os.Getenv("GAZE_COVERAGE_RUN") != ""` as the primary skip condition. The existing `testing.Short()` check MUST remain as a secondary skip guard. Unit test variants that use injected `runCrapFunc` do NOT need the env-var guard (they do not trigger recursion).

Previously: Only `testing.Short()` was checked, coupling gaze's recursion prevention to the standard Go `-short` mechanism.

## REMOVED Requirements

None.
