## ADDED Requirements

### Requirement: ResolvePackagePaths tests run without -short guard

All `TestResolvePackagePaths_*` tests in `internal/loader/loader_test.go`
MUST run without `testing.Short()` guards. These tests use
`packages.NeedName` load mode which is lightweight and does not warrant
the guard.

#### Scenario: Ungated tests contribute to CRAPload coverage
- **GIVEN** `ResolvePackagePaths` has 5 tests gated by `testing.Short()`
- **WHEN** the `-short` guards are removed from those 5 tests
- **THEN** all 6 `ResolvePackagePaths` tests run during `go test -short -coverprofile`
- **AND** `ResolvePackagePaths` line coverage exceeds 70%
- **AND** CRAP score drops below the threshold (15)

### Requirement: ResolvePackagePaths nil-stderr branch coverage

A test MUST verify that `ResolvePackagePaths` handles `nil` stderr
gracefully when packages have load errors.

#### Scenario: Nil stderr with invalid pattern does not panic
- **GIVEN** an invalid package pattern that produces `pkg.Errors`
- **WHEN** `ResolvePackagePaths` is called with `stderr` set to `nil`
- **THEN** no panic occurs
- **AND** the invalid package is excluded from the result
- **AND** valid packages in the same call are still returned

### Requirement: ResolvePackagePaths deduplication branch coverage

A test MUST verify that `ResolvePackagePaths` deduplicates package paths
when the same pattern resolves to the same package multiple times.

#### Scenario: Duplicate patterns produce deduplicated results
- **GIVEN** a patterns list containing the same valid package path twice
- **WHEN** `ResolvePackagePaths` is called
- **THEN** the result contains each package path exactly once

### Requirement: BuildContractCoverageFunc testable with synthetic data

`BuildContractCoverageFunc` MUST accept an optional deps parameter that
enables testing with synthetic implementations of its three external
dependencies: `ResolvePackagePaths`, `config.LoadFromDir`, and
`buildEffectsSet`.

#### Scenario: DI defaults to production implementations
- **GIVEN** `BuildContractCoverageFunc` is called without deps (zero-value)
- **WHEN** the function executes
- **THEN** it uses the real `loader.ResolvePackagePaths`, `config.LoadFromDir`, and `buildEffectsSet` implementations
- **AND** behavior is identical to the current production path

#### Scenario: ResolvePackagePaths error returns nil
- **GIVEN** deps with `resolvePackagePaths` injected to return an error
- **WHEN** `BuildContractCoverageFunc` is called
- **THEN** it returns `(nil, nil)`
- **AND** writes a warning to stderr

#### Scenario: Empty package paths returns nil
- **GIVEN** deps with `resolvePackagePaths` injected to return an empty slice
- **WHEN** `BuildContractCoverageFunc` is called
- **THEN** it returns `(nil, nil)`

#### Scenario: Both maps empty returns nil with degraded packages
- **GIVEN** deps injected so that `buildEffectsSet` returns empty and `buildCoverageMap` returns empty
- **WHEN** `BuildContractCoverageFunc` is called
- **THEN** it returns `(nil, degradedPkgs)`

#### Scenario: Closure returns coverage info for known function
- **GIVEN** deps injected with a coverage map containing `"pkg:Func"` at 80% coverage
- **WHEN** the returned closure is called with `("pkg", "Func")`
- **THEN** it returns the `ContractCoverageInfo` with `ok=true`

#### Scenario: Closure returns no_test_coverage for function with effects
- **GIVEN** deps injected with `"pkg:Func"` in the effects set but not in coverage map
- **WHEN** the returned closure is called with `("pkg", "Func")`
- **THEN** it returns `ContractCoverageInfo{Reason: "no_test_coverage"}` with `ok=false`

#### Scenario: Closure returns no_effects_detected for unknown function
- **GIVEN** deps injected with neither effects nor coverage for `"pkg:Unknown"`
- **WHEN** the returned closure is called with `("pkg", "Unknown")`
- **THEN** it returns `ContractCoverageInfo{Reason: "no_effects_detected"}` with `ok=false`

### Requirement: loadStreaming JSONL parsing testable in isolation

The JSONL scanner loop in `loadStreaming` MUST be extracted into a
standalone `parseSideEffectStream` function that accepts a
`*bufio.Scanner` and returns `([]protocol.AnalyzedFunction, error)`.

#### Scenario: Valid JSONL produces correct results
- **GIVEN** a scanner containing 3 valid JSONL lines (one `AnalyzedFunction` per line)
- **WHEN** `parseSideEffectStream` is called
- **THEN** it returns a slice of 3 `AnalyzedFunction` values with correct fields

#### Scenario: Empty stream returns empty slice
- **GIVEN** a scanner containing no data (empty input)
- **WHEN** `parseSideEffectStream` is called
- **THEN** it returns an empty/nil slice and no error

#### Scenario: Empty lines are skipped
- **GIVEN** a scanner containing valid JSONL interspersed with empty lines
- **WHEN** `parseSideEffectStream` is called
- **THEN** empty lines are ignored and only valid JSONL lines are parsed

#### Scenario: Malformed JSON produces error with line context
- **GIVEN** a scanner containing valid JSONL on line 1 and malformed JSON on line 2
- **WHEN** `parseSideEffectStream` is called
- **THEN** it returns an error containing "malformed JSONL on line 2"
- **AND** the error includes a truncated excerpt of the bad content

#### Scenario: Long malformed line is truncated in error
- **GIVEN** a scanner containing a malformed JSON line exceeding 200 bytes
- **WHEN** `parseSideEffectStream` is called
- **THEN** the error message truncates the content excerpt at 200 bytes with "..."

#### Scenario: Scanner I/O error is propagated
- **GIVEN** a scanner backed by a reader that returns an I/O error mid-stream
- **WHEN** `parseSideEffectStream` is called
- **THEN** it returns an error containing "reading analyze/stream response"

## MODIFIED Requirements

### Requirement: loadStreaming delegates to parseSideEffectStream

`loadStreaming` MUST delegate its JSONL parsing to the extracted
`parseSideEffectStream` function. The `CallStream` protocol call and
`convertAnalysisResults` conversion remain in `loadStreaming`.

Previously: `loadStreaming` contained the JSONL scanner loop inline.

## REMOVED Requirements

None.
