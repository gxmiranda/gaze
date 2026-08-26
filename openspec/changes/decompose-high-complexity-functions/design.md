## Context

Five functions exceed the project's CC ≤15 target. They span three
files across two packages:

| Function | File | CC | CRAP | Coverage | Lines |
|----------|------|----|------|----------|-------|
| `writeOneResult` | `internal/report/text.go:63` | 32 | 32.3 | 93.2% | 63-212 (~150) |
| `runQuality` | `cmd/gaze/main.go:1052` | 32 | 40.5 | 76.3% | 1052-1230 (~178) |
| `detectASTReceiverMutations` | `internal/analysis/mutation.go:458` | 24 | 67.0 | 44.3% | 458-553 (~95) |
| `runCrap` | `cmd/gaze/main.go:493` | 19 | 38.0 | 68.2% | 493-599 (~106) |
| `isPointerArgStore` | `internal/analysis/mutation.go:303` | 13 | 34.1 | 0.0% | 303-341 (~38) |

The project has successfully completed 5+ similar decompositions (q4-complexity-reduction,
crapload-decompose-pr2a/pr2b/pr2c, extract-violation-helper), establishing proven patterns
for helper extraction, testing with synthetic data, and DI-based testability.

## Goals / Non-Goals

### Goals
- Reduce all 5 functions to CC ≤15
- Make extracted helpers independently testable with unit tests
- Preserve all existing behavior — zero behavioral changes
- Remove structurally unreachable branches in `isPointerArgStore`
- All new tests run without `testing.Short()` guard
- CRAPload and GazeCRAPload do not increase

### Non-Goals
- Decomposing other high-complexity functions not in the #200 scope
- Changing confidence values, matching behavior, or classification logic
- Adding new CLI flags, output formats, or user-facing features
- Modifying any CI gate thresholds or acceptance criteria values

## Decisions

### D1: Decompose `writeOneResult` via row-building and signal extraction

**Decision**: Extract two helpers: `writeEffectRows` (builds `[][]string`
row data with truncation) and `writeVerboseSignals` (prints signal
breakdown for verbose mode).

**Rationale**: The function has two major branches (showClassify/else)
that share identical row-building logic with minor column differences.
Extracting row construction into a shared helper eliminates the
duplication. The verbose signal loop is an independent concern that
can be called after table rendering in both branches.

The table construction itself (lipgloss table setup, StyleFunc) remains
inline because it depends on branch-specific column counts and style
rules. The key complexity reduction comes from extracting the
data-preparation loop, not the rendering call.

**Signatures**:
```go
func writeEffectRows(effects []taxonomy.SideEffect, maxDesc int, showClassify bool) [][]string
func writeVerboseSignals(w io.Writer, effects []taxonomy.SideEffect)
```

### D2: Decompose `runQuality` via per-package helper and empty-result handler

**Decision**: Extract two helpers: `runQualityPerPackage` (encapsulates
the per-package loop body: analyze → classify → load test pkg → assess)
and `writeQualityEmptyResults` (handles the zero-result output path
including JSON/text formatting and skipped test display).

**Rationale**: The per-package loop body (lines 1115-1165) is a
self-contained pipeline of 4 operations with clear inputs (package path,
options, config) and outputs (reports, summary). Extracting it makes
each operation's error handling independently testable.

The empty-result handler (lines 1174-1214) has its own complexity from
the JSON/text format switch and the skipped test truncation logic.
Separating it from the threshold check (which follows) makes both
concerns independently testable.

`runQuality` retains: format validation, external analyzer rejection,
package resolution, config loading, module loading, AI mapper wiring,
the for-loop orchestration, summary merging, and threshold evaluation.

**Signatures**:
```go
func runQualityPerPackage(
    pkgPath string,
    p qualityParams,
    opts analysis.Options,
    cfg *config.Config,
    modPkgs []*packages.Package,
    aiMapperFn quality.AIMapperFunc,
) ([]taxonomy.QualityReport, *taxonomy.PackageSummary, error)

func writeQualityEmptyResults(
    w io.Writer,
    format string,
    merged *taxonomy.PackageSummary,
) error
```

### D3: Decompose `detectASTReceiverMutations` via per-node-type handlers

**Decision**: Extract three handlers: `handleReceiverAssignStmt`,
`handleReceiverIncDecStmt`, and `handleReceiverCallExpr`. Each takes
the AST node and receiver name, returns `(found bool, pos token.Pos)`.

**Rationale**: The `ast.Inspect` closure contains a 3-case type switch
with each case following the same pattern: check if the node references
the receiver, verify it's a field access (not bare receiver), return
position. This is the same decomposition pattern used successfully in
`AnalyzeP1Effects`/`AnalyzeP2Effects` (spec 009).

The handlers are pure functions operating on AST nodes with no external
dependencies — they can be tested with synthetic AST constructed via
`go/parser` without any DI.

**Signatures**:
```go
func handleReceiverAssignStmt(node *ast.AssignStmt, receiverName string) (bool, token.Pos)
func handleReceiverIncDecStmt(node *ast.IncDecStmt, receiverName string) (bool, token.Pos)
func handleReceiverCallExpr(node *ast.CallExpr, receiverName string) (bool, token.Pos)
```

### D4: Decompose `runCrap` via baseline, output, and gate helpers

**Decision**: Extract three helpers: `resolveBaselineAndCompare`
(baseline path resolution + load + compare), `writeCrapOutputAndSummary`
(comparison or normal report + CI summary), and `evaluateCrapGates`
(baseline gate then threshold gate).

**Rationale**: `runCrap` has three sequential phases after the initial
analysis call: (1) baseline handling (lines 562-571), (2) output
writing (lines 573-584), and (3) gate evaluation (lines 586-598).
Each phase is logically independent and produces clear outputs.

**Critical invariant**: The gate ordering (baseline gate before
threshold gate, per D7 in the original design) MUST be preserved in
`evaluateCrapGates`. This ordering ensures comparison output is always
visible before a threshold failure. After extraction, the ordering is
explicitly visible in the helper's control flow rather than implicit
in the parent function's line ordering.

**Signatures**:
```go
func resolveBaselineAndCompare(
    baselinePath string,
    moduleDir string,
    stderr io.Writer,
    rpt *crap.Report,
    baselineExplicitlySet bool,
) (*crap.ComparisonResult, error)

func writeCrapOutputAndSummary(
    stdout, stderr io.Writer,
    format string,
    rpt *crap.Report,
    comparisonResult *crap.ComparisonResult,
    maxCrapload, maxGazeCrapload *int,
) error

func evaluateCrapGates(
    rpt *crap.Report,
    comparisonResult *crap.ComparisonResult,
    stderr io.Writer,
    maxCrapload, maxGazeCrapload *int,
) error
```

### D5: Simplify `isPointerArgStore` by removing unreachable branches

**Decision**: Remove the `UnOp`, `FieldAddr`, and `IndexAddr` type
checks after the initial `tracesToParam(addr, param)` call. These
branches are structurally unreachable because `tracesToParam` (via
`tracesToParamVisited`) already recursively walks `FieldAddr`, `IndexAddr`,
and `UnOp` chains to find the root parameter.

**Rationale**: The `tracesToParam` function at line 345 delegates to
`tracesToParamVisited` at line 353, which handles:
- `*ssa.FieldAddr` → recurses on `val.X` (line 364)
- `*ssa.IndexAddr` → recurses on `val.X` (line 366)
- `*ssa.UnOp` → recurses on `val.X` (line 368)
- `*ssa.Phi` → recurses on each edge (line 371)

So when `isPointerArgStore` checks `tracesToParam(addr, param)` at line
307, it already walks the full chain `addr → FieldAddr.X → UnOp.X →
param`. The subsequent checks (e.g., "is addr a FieldAddr? if so, does
FieldAddr.X trace to param?") are checking subsets of what
`tracesToParam` already checked.

This is confirmed by learning `crapload-di-coverage-20260703T151625`
and by the 0% coverage on these branches — they are never reached in
any test suite because the first `tracesToParam` call catches all cases.

After removal, the function becomes a simple loop: for each param,
call `tracesToParam`; return the first match. CC drops from 13 to ~3.

Unit tests MUST be added to verify the simplified function still
correctly identifies pointer argument stores for all SSA instruction
patterns (direct store, store through FieldAddr, store through
IndexAddr, store through UnOp).

### D6: No DI needed — all functions use explicit parameters or existing DI

**Decision**: Do not add new DI structs.

**Rationale**: `writeOneResult` takes `io.Writer` + data. The AST
handlers take AST nodes + strings. `runQuality` already has
`qualityParams` (a DI struct). `runCrap` already has `crapParams`
(a DI struct). No additional dependency injection is needed because
the extractable concerns operate on data already passed as parameters.

### D7: Task ordering — `isPointerArgStore` first, `writeOneResult` last

**Decision**: Decompose in order: `isPointerArgStore` (simplest, branch
removal) → `detectASTReceiverMutations` (same file) →
`runCrap` → `runQuality` (same file) → `writeOneResult` (different
package, most complex).

**Rationale**: `isPointerArgStore` and `detectASTReceiverMutations` are
in the same file (`mutation.go`) and should be done together. `runCrap`
and `runQuality` are in the same file (`main.go`) and should be done
together. `writeOneResult` is in a different package and is independent.
Starting with the simplest change (branch removal) builds momentum and
establishes the testing pattern before tackling the higher-complexity
extractions.

## Risks / Trade-offs

### Risk: `resolveBaselineAndCompare` duplicates parameter threading

The baseline resolution currently reads `p.baselinePath` and
`p.moduleDir` from the `crapParams` struct. Extracting to a helper
requires passing these as explicit parameters, which is slightly
more verbose. Acceptable because the params are well-typed values
and the helper is unexported.

### Risk: `runQualityPerPackage` has many parameters

The per-package helper needs the package path, params struct, analysis
options, config, module packages, and AI mapper function — 6 parameters.
This is at the boundary of acceptable Go style. However, the params
struct already encapsulates most CLI-level config, and the remaining
parameters are per-invocation context that can't be folded in. The
alternative (a new options struct) would add ceremony without reducing
cognitive load.

### Risk: `isPointerArgStore` simplification may miss edge cases

If `tracesToParam` has a bug or doesn't cover a specific SSA pattern
that the explicit type checks would have caught, removing the explicit
checks could introduce a false negative. Mitigated by:
1. The explicit checks have 0% coverage — they are never reached
2. The existing test suite (`TestIsPointerArgStore`) covers all
   five subtests and would catch regressions
3. `tracesToParam` handles `FieldAddr`, `IndexAddr`, `UnOp`, and `Phi`
   exhaustively

### Trade-off: More functions per file

Each file gains 2-3 new unexported helpers. `mutation.go` (638 lines)
gains 3. `main.go` (~1979 lines) gains 5. `text.go` (212 lines) gains
2. The function count increase is acceptable because each helper has a
clear single responsibility and the parent functions become
significantly more readable.
