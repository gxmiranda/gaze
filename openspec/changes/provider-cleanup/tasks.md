## 1. Deprecated Field Removal

- [ ] 1.1 Remove `ContractCoverageFunc` and `SSADegradedPackages` fields from `crap.Options` in `internal/crap/analyze.go`. Remove the `// Deprecated:` GoDoc comments. Remove the fallback dispatch path in `crap.Analyze` (lines 139-155) that checks `ContractCoverageProvider` then falls back to `ContractCoverageFunc`. Remove the `ContractCoverageFunc` call in `computeScores` (lines 292-293) — contract coverage now comes exclusively through `ContractCoverageProvider` which sets up the scoring path via its `Build()` return value.
- [ ] 1.2 Update `internal/crap/analyze_internal_test.go` — replace all `opts.ContractCoverageFunc = ...` usages with mock `ContractCoverageProvider` implementations. There are 4 occurrences (lines 154, 221, 376, 413).
- [ ] 1.3 Update `internal/crap/crap_test.go` — replace `testProviderOpts()` or any test helper that sets `ContractCoverageFunc`. Update the `lookupCoverage` and `Analyze` tests.
- [ ] 1.4 Remove `TestMockProviders_ProviderPrecedence` from `internal/provider/mockprovider/mock_test.go` — this test validates precedence between `ContractCoverageProvider` and `ContractCoverageFunc`, which no longer exists. Remove the `//nolint:staticcheck` directive that was on the deprecated field access.

## 2. isMainPkg Deduplication

- [ ] 2.1 In `internal/provider/goprovider/sideeffect.go`, replace the local `isMainPkg` function with a call to `loader.IsMainPkg`. Remove the local function definition and its `// NOTE: keep in sync` comment. Add `loader` to the import if not already present.

## 3. BuildContractCoverageFunc Relocation

- [ ] 3.1 Move `BuildContractCoverageFunc`, `analyzePackageCoverage`, `loadTestPackage`, `classifyResults`, and `loadGazeConfigBestEffort` from `internal/crap/contract.go` to `internal/provider/goprovider/contract.go`. Update `GoContractCoverageProvider.Build()` to call the moved functions directly instead of through `crap.BuildContractCoverageFunc`. Keep the function signatures identical.
- [ ] 3.2 Remove `internal/crap/contract.go` (now empty after the move). Verify `internal/crap/` no longer imports `internal/analysis`, `internal/classify`, `internal/quality`, `internal/loader`, or `golang.org/x/tools/go/packages`.
- [ ] 3.3 Update any callers of `crap.BuildContractCoverageFunc` outside of `goprovider/` — search `cmd/gaze/main.go` and `internal/aireport/` for remaining references. These should now call through `GoContractCoverageProvider` or the moved function in `goprovider`.

## 4. Non-Regression Verification

- [ ] 4.1 Run `go build ./...` — MUST compile with zero errors.
- [ ] 4.2 Run `go test -race -count=1 -short ./...` — all tests MUST pass.
- [ ] 4.3 Run `go test -race -count=1 -run TestRunSelfCheck -timeout 30m ./cmd/gaze/...` — E2E self-check MUST produce identical output.
- [ ] 4.4 Verify `internal/crap/` import list contains zero Go-specific analysis imports.

## 5. Documentation

- [ ] 5.1 Update `AGENTS.md` — add to Recent Changes. Note that `BuildContractCoverageFunc` moved from `internal/crap/` to `internal/provider/goprovider/`.
- [ ] 5.2 Update GoDoc comments on any moved or modified exported functions.

## 6. Constitution Alignment Verification

- [ ] 6.1 Verify all four principles: Accuracy (no scoring changes, E2E passes), Minimal Assumptions (no user-facing changes), Actionable Output (output unchanged), Testability (deprecated test paths removed, remaining tests cover provider-only path).