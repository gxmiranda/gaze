## ADDED Requirements

None — this is a cleanup change.

## MODIFIED Requirements

### Requirement: crap.Options deprecated field removal

`crap.Options` MUST NOT contain `ContractCoverageFunc` or `SSADegradedPackages` fields. All contract coverage data MUST flow through `ContractCoverageProvider.Build()`. The fallback path in `crap.Analyze` that checks these deprecated fields MUST be removed.

#### Scenario: Options without deprecated fields

- **GIVEN** `crap.Options` constructed with `ContractCoverageProvider` set
- **WHEN** `crap.Analyze` runs
- **THEN** contract coverage data MUST come from the provider and no deprecated field access occurs

#### Scenario: Backward-compat fallback removed

- **GIVEN** the `crap.Analyze` function
- **WHEN** inspecting its source code
- **THEN** there MUST be no reference to `opts.ContractCoverageFunc` or `opts.SSADegradedPackages`

---

### Requirement: isMainPkg deduplication

The `internal/provider/goprovider/` package MUST use `loader.IsMainPkg` instead of a local `isMainPkg` copy. No duplicate `isMainPkg` implementations MUST exist in the codebase.

#### Scenario: goprovider uses loader.IsMainPkg

- **GIVEN** `GoSideEffectAnalyzer.Analyze` in `goprovider/sideeffect.go`
- **WHEN** it checks whether a package is a main package
- **THEN** it MUST call `loader.IsMainPkg`, not a local function

---

### Requirement: BuildContractCoverageFunc relocation

`BuildContractCoverageFunc` and its helpers MUST reside in `internal/provider/goprovider/`, not in `internal/crap/`. The `internal/crap/` package MUST NOT import `internal/analysis`, `internal/classify`, `internal/quality`, `internal/loader`, or `golang.org/x/tools/go/packages`.

#### Scenario: crap package import isolation

- **GIVEN** the `internal/crap/` package
- **WHEN** its import list is inspected
- **THEN** it MUST contain zero imports from `internal/analysis`, `internal/classify`, `internal/quality`, `internal/loader`, or `golang.org/x/tools/go/packages`

## REMOVED Requirements

### Requirement: ContractCoverageFunc callback pattern

The `ContractCoverageFunc` field on `crap.Options` is removed. Callers MUST use `ContractCoverageProvider` instead. The removal is justified because all production callers were migrated in Phase 1 (PR #165).

#### Scenario: Compile-time verification

- **GIVEN** a Go source file that references `crap.Options{ContractCoverageFunc: ...}`
- **WHEN** compiled
- **THEN** compilation MUST fail (field does not exist)