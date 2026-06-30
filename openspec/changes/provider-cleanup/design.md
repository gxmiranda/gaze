## Context

Phase 1 (PR #165) extracted provider interfaces and kept `ContractCoverageFunc` and `SSADegradedPackages` in `crap.Options` for backward compatibility. Phase 2 (PR #178) added external analyzer adapters. All production callers now use the provider interfaces. The deprecated fields and the `BuildContractCoverageFunc` function in `internal/crap/` are the remaining Go-specific coupling in the scoring package.

Relevant code locations:
- `internal/crap/analyze.go:139-155` — fallback path checking deprecated fields
- `internal/crap/analyze.go:292-293` — `ContractCoverageFunc` call in `computeScores`
- `internal/crap/contract.go` — `BuildContractCoverageFunc` and helpers (Go-specific)
- `internal/provider/goprovider/sideeffect.go:113-120` — local `isMainPkg` copy
- `internal/loader/loader.go:181` — canonical `IsMainPkg`

## Goals / Non-Goals

### Goals
- Remove deprecated `ContractCoverageFunc` and `SSADegradedPackages` from `crap.Options`
- Remove the fallback code path in `crap.Analyze` that dispatches on these fields
- Deduplicate `isMainPkg` by using `loader.IsMainPkg`
- Move `BuildContractCoverageFunc` and helpers to `internal/provider/goprovider/`
- Achieve zero Go-specific imports in `internal/crap/`

### Non-Goals
- Taxonomy generalization — separate spec needed (research on language-specific effect types)
- Streaming protocol mode — separate spec needed (performance benchmarking)
- Removing the `crap.BuildContractCoverageFunc` export entirely — it may have test utility; move it, don't delete the symbol

## Decisions

### D1: Remove deprecated fields atomically

Remove `ContractCoverageFunc`, `SSADegradedPackages`, and the fallback path in a single commit. Update all callers simultaneously. This avoids an intermediate state where the fields exist but the fallback is gone (compile error).

### D2: Move BuildContractCoverageFunc to goprovider

`BuildContractCoverageFunc` and its helpers (`analyzePackageCoverage`, `loadTestPackage`, `classifyResults`, `loadGazeConfigBestEffort`) move to `internal/provider/goprovider/contract.go`, replacing the current thin-wrapper implementation. `GoContractCoverageProvider.Build()` calls the moved functions directly.

The `crap` package no longer imports `analysis`, `classify`, `quality`, `loader`, or `go/packages`.

### D3: isMainPkg uses loader.IsMainPkg

Replace the local `isMainPkg` in `goprovider/sideeffect.go` with `loader.IsMainPkg`. The `goprovider` package already imports `loader` for other functionality.

### D4: Remove precedence test

The mock test `TestMockProviders_ProviderPrecedence` validates that `ContractCoverageProvider` takes priority over `ContractCoverageFunc`. Once `ContractCoverageFunc` is removed, this test has no subject. Remove it.

## Risks / Trade-offs

### R1: Internal API break

Removing `crap.Options` fields is a breaking change to the internal API. All callers are within the module (`cmd/gaze/`, `internal/aireport/`, `internal/provider/`). No external consumers exist (`internal/` packages cannot be imported).

### R2: Large file move

Moving `contract.go` from `crap/` to `goprovider/` is a large diff that may be hard to review. Mitigated by keeping the function signatures identical — only the package location changes.