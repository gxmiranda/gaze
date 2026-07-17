## Context

`internal/quality/mapping.go` (1575 lines, 32 functions) is the assertion
mapping engine — it maps test assertion expressions to side effects via AST
and SSA analysis. Three functions dominate the complexity budget:

| Function | Complexity | Lines | Direct unit tests |
|----------|-----------|-------|-------------------|
| `matchContainerUnwrap` | 50 | 201 (959-1159) | 0 (4 integration) |
| `isTransformationCall` | 26 | 86 (745-830) | 5 (3 gaps) |
| `matchAssertionToEffect` | 25 | 143 (1185-1327) | 0 |

These functions already accept explicit parameters (AST nodes, type info,
maps) — no DI structs are needed. The existing `parseAndTypeCheck` helper
in `container_unwrap_internal_test.go` provides synthetic AST construction
for unit testing.

### Matching cascade (call order in `mapAssertionsToEffectsImpl`)

```text
mapAssertionsToEffectsImpl
  ├─ matchAssertionToEffect (primary, conf 75/70/65)
  ├─ matchInlineCall (inline calls, conf 60) — out of scope
  ├─ matchContainerUnwrap (container chains, conf 55)
  └─ aiMapperFn (AI fallback, conf 50) — out of scope
```

## Goals / Non-Goals

### Goals
- Reduce combined complexity from 101 to ≤30 across the three functions
- Make each interleaved concern independently testable
- Fill `isTransformationCall` test gaps (io.Reader, empty interface, ordering)
- Add direct unit tests for `matchAssertionToEffect` (currently zero)
- Add unit tests for `matchContainerUnwrap` helpers (currently integration only)
- All new tests run without `testing.Short()` guard

### Non-Goals
- Decomposing `matchInlineCall` (complexity 17, deferred to Phase 2c)
- Decomposing `findHelperCallInFuncDepth` (complexity 15, deferred)
- Decomposing `mapAssertionsToEffectsImpl` (complexity 14, deferred)
- Changing confidence values (75, 70, 65, 55) or matching behavior
- Modifying the `TestSC003_MappingAccuracy` ratchet floor (85.0%)

## Decisions

### D1: Extract `isByteLikeParam` and `isPointerDestParam` from `isTransformationCall`

**Decision**: Split the parameter-classification logic in the loop body into
two predicate functions that each return `bool`.

**Rationale**: The current loop body interleaves 5 type-checking patterns
([]byte, string, io.Reader, *T, empty interface) with index tracking in a
single iteration. Extracting two classifiers makes each pattern independently
testable and reduces `isTransformationCall` to: validate → loop → classify →
return indices.

**Signatures**:
```go
func isByteLikeParam(typ types.Type) bool
func isPointerDestParam(typ types.Type) bool
```

### D2: Extract `matchDirect` and `matchIndirectRoot` from `matchAssertionToEffect`

**Decision**: Extract the two `ast.Inspect` passes as separate functions.

**Rationale**: The current function has two distinct matching strategies
interleaved with nil guards and helper bridge construction. Each pass
produces an `*AssertionMapping` independently. Separating them allows
testing each matching strategy in isolation.

Both helpers accept `*types.Info` directly instead of `*packages.Package`
because they only use `TypesInfo`; the parent function extracts and passes it.

**Signatures**:
```go
func matchDirect(
    site AssertionSite,
    objToEffectID map[types.Object]string,
    effectMap map[string]*taxonomy.SideEffect,
    info *types.Info,
    helperBridge map[types.Object]types.Object,
) *taxonomy.AssertionMapping

func matchIndirectRoot(
    site AssertionSite,
    objToEffectID map[types.Object]string,
    effectMap map[string]*taxonomy.SideEffect,
    info *types.Info,
) *taxonomy.AssertionMapping
```

### D3: Extract `collectTrackedVars`, `traceForwardDataFlow`, and `matchTrackedInExpr` from `matchContainerUnwrap`

**Decision**: Decompose the 201-line function into three phases: setup,
trace, and match.

**Rationale**: The function's complexity comes from interleaving forward
data-flow tracing (multi-iteration AST walking with transformation call
detection) with assertion matching. These are fundamentally different
operations that happen to share a `tracked` set. Separating them makes
each phase testable with controlled inputs.

**Signatures**:
```go
func collectTrackedVars(
    objToEffectID map[types.Object]string,
    returnEffectID string,
) map[types.Object]bool

func traceForwardDataFlow(
    tracked map[types.Object]bool,
    testPkg *packages.Package,
) map[types.Object]bool

func matchTrackedInExpr(
    expr ast.Expr,
    tracked map[types.Object]bool,
    info *types.Info,
) bool
```

`traceForwardDataFlow` takes `*packages.Package` (for `Syntax` file
iteration) and extracts `info := testPkg.TypesInfo` internally for
type resolution. This differs from `matchTrackedInExpr` which takes
`*types.Info` directly — the asymmetry exists because `traceForwardDataFlow`
needs both file access (`Syntax`) and type info, while `matchTrackedInExpr`
only needs type info. The function mutates and returns the input `tracked`
map (same map, expanded with new entries).

`traceForwardDataFlow` absorbs the multi-iteration loop, transformation
call detection via `isTransformationCall`, and data-extraction gating via
`isDataExtraction`. Even extracted, this function will have inherent
complexity (~12) from the nested AST walking — hence the ≤12 target for
`matchContainerUnwrap` overall.

### D4: No DI needed — functions already take explicit parameters

**Decision**: Do not add DI structs or function parameters for testability.

**Rationale**: Unlike Phase 2a's `BuildContractCoverageFunc` (which needed
DI to avoid real package loading), these functions accept `*ast.File`,
`*types.Info`, and `*packages.Package` directly. The `parseAndTypeCheck`
test helper creates synthetic instances of these types from Go source
strings. No external I/O is involved.

### D5: Task ordering — `isTransformationCall` first

**Decision**: Decompose in order: `isTransformationCall` → 
`matchAssertionToEffect` → `matchContainerUnwrap`.

**Rationale**: `matchContainerUnwrap` calls `isTransformationCall` internally
(line 1037). Decomposing `isTransformationCall` first means the cleaner
helper is in place when `matchContainerUnwrap` is refactored.
`matchAssertionToEffect` is independent and can go second.

## Risks / Trade-offs

### Risk: `traceForwardDataFlow` remains complex

Even after extraction, the forward data-flow tracing involves nested
`ast.Inspect` closures, multi-iteration convergence, transformation call
detection, and data-extraction gating. Target complexity is ≤12 for this
function specifically. If it exceeds 12, a second decomposition pass may
be needed in Phase 2c.

### Risk: Test fixture complexity for `matchContainerUnwrap` helpers

The `parseAndTypeCheck` helper only provides AST and type info — it does
not create `packages.Package` instances with full dependency resolution.
`traceForwardDataFlow` takes `*packages.Package` (for file access).
Tests will need to construct minimal `packages.Package` structs with
synthetic `Syntax` (parsed files) and `TypesInfo`. This is feasible
(the existing `isTransformationCall` tests show the pattern) but requires
care.

### Trade-off: More functions in mapping.go

Seven new unexported helpers increase the function count from 32 to ~39.
This is acceptable because each helper has a clear single responsibility
and reduces the cognitive load of the three parent functions. The file
is already large (1575 lines) — decomposition makes it more navigable,
not less.

### Trade-off: `matchAssertionToEffect` → `matchDirect` passes `helperBridge` explicitly

The helper bridge map is computed inside `matchAssertionToEffect` and
consumed by `matchDirect`. After extraction, `matchAssertionToEffect`
computes the bridge and passes it to `matchDirect`. This exposes an
internal detail in the function signature. Acceptable because both
functions are unexported and the bridge map is a well-typed value
(`map[types.Object]types.Object`).
