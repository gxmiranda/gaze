## Why

CRAPload is at 29/38 after Phases 1a–2b of issue #166. Three functions
remain with high CRAP scores that can be resolved with test coverage
alone — no structural decomposition needed:

| Function | CRAP | CC | Coverage | Package |
|----------|------|----|----------|---------|
| `ResolvePackagePaths` | 81.6 | 10 | 10.5% | `internal/loader` |
| `BuildContractCoverageFunc` | 56.0 | 7 | 0% | `provider/goprovider` |
| `loadStreaming` | 42.0 | 6 | 0% | `internal/adapter` |

**`ResolvePackagePaths`** already has 6 tests — 5 are gated by
`testing.Short()` and skipped during `gaze crap`'s internal
`go test -short -coverprofile` run, contributing zero coverage.
`packages.NeedName` is the lightest load mode (no parsing or
type-checking), so the guards are unnecessarily conservative.
Removing them is the cheapest CRAPload reduction in the project.

**`BuildContractCoverageFunc`** was decomposed from complexity 22 to 7
in Phase 2a (PR #192) but received no test coverage. The same file
already has a mature `contractCoverageDeps` DI pattern with
`successDeps()` helper — extending it to the orchestrator function
is straightforward.

**`loadStreaming`** is a 34-line JSONL parsing method on the external
analyzer adapter. It has zero tests because the only integration test
is gated by `testing.Short()`. The scanner loop can be tested by
extracting it into a helper that accepts a `*bufio.Scanner`, or by
extending the fake analyzer with error-mode flags.

This is Phase 2c (final phase) of issue #166. After this, the issue
will be closed and remaining decomposition work tracked in separate
focused issues.

## What Changes

### `internal/loader/` — remove `testing.Short()` guards

Remove `testing.Short()` guards from 5 existing `ResolvePackagePaths`
tests. These tests use `packages.NeedName` mode which resolves package
names without parsing or type-checking — they complete in sub-second
time and do not need the `-short` guard. Add 2 new tests for uncovered
branches (nil stderr with errors, duplicate pattern deduplication).

### `internal/provider/goprovider/` — add DI for `BuildContractCoverageFunc`

Add a `buildContractCoverageFuncDeps` struct with injectable function
fields for `resolvePackagePaths`, `loadConfig`, and `buildEffectsSetFn`,
following the nil-means-default pattern already used by
`contractCoverageDeps` in the same file. Add 6–7 synthetic unit tests
covering all 7 paths through the function (error, empty, both-maps-empty,
closure-found, closure-no-test, closure-no-effects, happy path).

### `internal/adapter/` — add scanner-level tests for `loadStreaming`

Extract the JSONL scanner loop into a `parseSideEffectStream` helper
that accepts a `*bufio.Scanner` and returns `([]protocol.AnalyzedFunction,
error)`. This isolates the parsing logic from the protocol client,
enabling 6 unit tests without spawning a subprocess: valid JSONL,
empty stream, empty lines skipped, malformed JSON, long-line truncation,
scanner I/O error.

## Capabilities

### New Capabilities
- None (internal testability improvement only)

### Modified Capabilities
- `BuildContractCoverageFunc`: Accepts optional deps struct for testing
  (nil defaults to production implementations, no caller changes)
- `loadStreaming`: Delegates JSONL parsing to extracted `parseSideEffectStream`
  helper (identical behavior, now independently testable)

### Removed Capabilities
- None

## Impact

- **Files modified**: `internal/loader/loader_test.go`,
  `internal/provider/goprovider/contract.go`,
  `internal/provider/goprovider/contract_internal_test.go`,
  `internal/adapter/sideeffect.go`,
  `internal/adapter/sideeffect_test.go`
- **No API surface changes**: All functions are unexported or the
  signature change is additive (optional variadic deps parameter)
- **No behavioral changes**: Production code paths remain identical.
  DI structs default to real implementations when nil.
- **Expected CRAPload reduction**: 29 → ~26 (3 functions drop below
  the CRAP threshold of 15)
- **Issue closure**: #166 will be closed after this PR merges. The
  original goal (CRAPload 38 → ~30 with 8-function buffer) was
  achieved in Phase 1; this phase extends the buffer to ~12 functions.

## Constitution Alignment

Assessed against the Gaze project constitution (v1.3.0).

### I. Accuracy

**Assessment**: N/A

This change does not modify side effect detection, classification, or
reporting logic. All production code paths remain identical.

### II. Minimal Assumptions

**Assessment**: PASS

No new assumptions are introduced. The DI patterns use zero-value
defaults that wire to real implementations, requiring no configuration
from users or callers. Removing `testing.Short()` guards from
`ResolvePackagePaths` tests is safe because `packages.NeedName` is the
lightest load mode — it does not assume anything about the host
project's build state.

### III. Actionable Output

**Assessment**: N/A

No changes to output formats, report content, or metric computation.

### IV. Testability

**Assessment**: PASS

This change directly advances Principle IV. It makes 3 functions testable
in isolation: `ResolvePackagePaths` by removing unnecessary test guards,
`BuildContractCoverageFunc` by adding DI for synthetic testing, and
`loadStreaming` by extracting the scanner loop into a pure parsing
function. Tests verify observable side effects (return values, error
conditions, closure behavior) rather than implementation details.
Coverage strategy: unit tests with synthetic data, no `-short` guard,
targeting ≥70% line coverage per function.
