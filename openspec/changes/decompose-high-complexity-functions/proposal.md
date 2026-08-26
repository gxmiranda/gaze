## Why

Five functions identified in issue #166 scope reconciliation (items #7, #9, #14, #15) exceed the project's cyclomatic complexity target of CC ≤15. High-complexity functions resist effective unit testing because branch coverage requires exponentially more test cases. Decomposing these functions into focused helpers enables better test isolation, reduces CRAP scores, and continues the CRAPload reduction effort that has already successfully decomposed 10+ functions across prior PRs (q4-complexity-reduction, crapload-decompose-pr2a/pr2b/pr2c, extract-violation-helper).

Verified CC values (gocyclo):

| Function | File | CC | CRAP | Coverage |
|----------|------|----|------|----------|
| `writeOneResult` | `internal/report/text.go:63` | 32 | 32.3 | 93.2% |
| `runQuality` | `cmd/gaze/main.go:1052` | 32 | 40.5 | 76.3% |
| `detectASTReceiverMutations` | `internal/analysis/mutation.go:458` | 24 | 67.0 | 44.3% |
| `runCrap` | `cmd/gaze/main.go:493` | 19 | 38.0 | 68.2% |
| `isPointerArgStore` | `internal/analysis/mutation.go:303` | 13 | 34.1 | 0.0% |

Note: `runQuality` measured at CC 32, higher than the ~15 estimate in issue #200. `runCrap` measured at CC 19, slightly below the issue's CC 20 claim. `isPointerArgStore` at CC 13 already meets the ≤15 target but has 0% coverage and structurally unreachable branches requiring cleanup.

## What Changes

Decompose each of the 5 functions into smaller helpers with CC ≤15 each. Extract logical phases into named functions, add unit tests for extracted helpers, and verify all existing tests pass without modification.

This is a pure refactoring change — no behavioral modifications, no API surface changes, no new dependencies.

## Capabilities

### New Capabilities
- None — this is a structural refactoring with no user-facing changes.

### Modified Capabilities
- `writeOneResult`: Decomposed into per-section helpers (`formatEffectLine`, `formatClassification`, `writeEffectsList`) for better test isolation.
- `runQuality`: Per-package loop and threshold evaluation extracted into helper functions.
- `detectASTReceiverMutations`: Per-AST-node-type handlers extracted following the established `AnalyzeP1Effects`/`AnalyzeP2Effects` pattern.
- `runCrap`: Baseline resolution, output writing, and threshold evaluation extracted into named helpers.
- `isPointerArgStore`: Structurally unreachable branches removed (confirmed by prior analysis that `tracesToParam` already walks FieldAddr/IndexAddr/UnOp chains internally), then unit tests added for the simplified function.

### Removed Capabilities
- None.

## Impact

- **Files modified**: `cmd/gaze/main.go`, `internal/analysis/mutation.go`, `internal/report/text.go`
- **Test files**: New `*_test.go` entries or additions for extracted helpers in all 3 packages
- **Risk**: Low — all 5 functions have existing test coverage providing regression safety. The project has 5+ successful prior decompositions using identical patterns.
- **CI gates**: CRAPload must not increase. All existing tests must pass with `-race -count=1`.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change is an internal structural refactoring. It does not affect artifact-based communication, inter-component interfaces, or self-describing outputs.

### II. Composability First

**Assessment**: PASS

All decomposed helpers remain internal to their respective packages. No new cross-package dependencies are introduced. Each package continues to function independently.

### III. Observable Quality

**Assessment**: PASS

Decomposition preserves all existing output formats (JSON, text). The extracted helpers do not alter machine-parseable output or provenance metadata. CRAP scores improve, which is itself an observable quality improvement.

### IV. Testability

**Assessment**: PASS

This change directly serves the Testability principle. Decomposing high-CC functions into focused helpers enables unit testing of individual code paths that were previously buried inside monolithic functions. Extracted helpers get dedicated unit tests covering their primary branches. Coverage ratchets are enforced (CRAPload must not increase).
