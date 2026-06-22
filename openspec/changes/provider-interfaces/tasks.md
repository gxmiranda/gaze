## 1. Interface Definitions

- [ ] 1.1 Create `internal/crap/provider.go` with `FunctionComplexity` struct (Package, Function, File, Line, Complexity fields with JSON tags) and four provider interface definitions: `ComplexityProvider` (method `Analyze(patterns []string, rootDir string) ([]FunctionComplexity, error)`), `LineCoverageProvider` (method `Coverage(patterns []string, rootDir string, coverProfile string) ([]FuncCoverage, error)`), `SideEffectAnalyzer` (method `Analyze(pkgPath string) ([]taxonomy.AnalysisResult, error)`), `ContractCoverageProvider` (method `Build(patterns []string, rootDir string) (func(pkg, function string) (ContractCoverageInfo, bool), []string, error)`). File MUST NOT import `go/ast`, `go/types`, `go/packages`, `go/ssa`, or `github.com/fzipp/gocyclo`.
- [ ] 1.2 Add provider fields to `crap.Options` in `internal/crap/crap.go`: `ComplexityProvider ComplexityProvider`, `LineCoverageProvider LineCoverageProvider`, `ContractCoverageProvider ContractCoverageProvider`. Keep existing `ContractCoverageFunc` and `SSADegradedPackages` fields for backward compatibility.

## 2. CRAP Pipeline Providers

- [ ] 2.1 Create `internal/provider/goprovider/complexity.go` with `GoComplexityProvider` struct implementing `crap.ComplexityProvider`. The `Analyze` method wraps `gocyclo.Analyze()` and converts `[]gocyclo.Stat` to `[]crap.FunctionComplexity`. Move `resolvePatterns` and `testFileRegexp` logic from `crap/analyze.go` into this adapter. Include a constructor `NewComplexityProvider() *GoComplexityProvider`.
- [ ] 2.2 Create `internal/provider/goprovider/coverage.go` with `GoLineCoverageProvider` struct implementing `crap.LineCoverageProvider`. The `Coverage` method wraps `generateCoverProfile()` + `ParseCoverProfile()` + `buildCoverMap()`. When `coverProfile` is non-empty, skip profile generation and parse the provided file. Move coverage generation logic from `crap/analyze.go`. Include fields for `Stderr io.Writer` configuration. Include a constructor `NewLineCoverageProvider(stderr io.Writer) *GoLineCoverageProvider`.
- [ ] 2.3 Refactor `computeScores` in `internal/crap/analyze.go` to accept `[]crap.FunctionComplexity` instead of `[]gocyclo.Stat`. Update field access: `stat.PkgName` to `stat.Package`, `stat.FuncName` to `stat.Function`, `stat.Pos.Filename` to `stat.File`, `stat.Pos.Line` to `stat.Line`. `stat.Complexity` is unchanged.
- [ ] 2.4 Refactor `crap.Analyze` in `internal/crap/analyze.go` to use `ComplexityProvider` and `LineCoverageProvider` from `Options`. If nil, instantiate `goprovider.NewComplexityProvider()` and `goprovider.NewLineCoverageProvider(opts.Stderr)`. Remove direct `gocyclo` import from `analyze.go`. The `coverMaps` type and `buildCoverMap` MAY stay in `analyze.go` or move to `goprovider/coverage.go` depending on which side of the boundary needs it.
- [ ] 2.5 Update `internal/crap/analyze_internal_test.go` to use `crap.FunctionComplexity` instead of `gocyclo.Stat` in `computeScores` test cases. Field mapping is mechanical.

## 3. Quality Pipeline Provider

- [ ] 3.1 Create `internal/provider/goprovider/sideeffect.go` with `GoSideEffectAnalyzer` struct implementing `crap.SideEffectAnalyzer`. The `Analyze` method calls `analysis.LoadAndAnalyze(pkgPath, analysisOpts)` then `classify.Classify(results, classifyOpts)`. Include config fields for classification options (`*config.GazeConfig`, module packages, verbose). Include a constructor accepting these dependencies.
- [ ] 3.2 Create `internal/provider/goprovider/contract.go` with `GoContractCoverageProvider` struct implementing `crap.ContractCoverageProvider`. The `Build` method wraps the logic currently in `BuildContractCoverageFunc` -- package path resolution, per-package side effect analysis + classification + test loading + quality assessment, contract coverage computation. Returns the lookup function, degraded packages list, and error. Include fields for `Stderr io.Writer`, `AIMapperFunc`, and `SideEffectAnalyzer` (allowing composition with the Go side effect analyzer). Include a constructor.
- [ ] 3.3 Refactor `crap.Analyze` in `internal/crap/analyze.go` to use `ContractCoverageProvider` from `Options` when set. When `ContractCoverageProvider` is non-nil, call `Build()` and use the returned lookup function and degraded packages. When nil, fall back to the existing `ContractCoverageFunc`/`SSADegradedPackages` pattern. This ensures zero behavior change for existing callers.
- [ ] 3.4 Update callers of `BuildContractCoverageFunc` in `cmd/gaze/main.go` -- either construct a `GoContractCoverageProvider` and pass it via `crap.Options.ContractCoverageProvider`, or continue using `ContractCoverageFunc` directly (both paths are valid during transition). The `runReport` pipeline in `internal/aireport/runner.go` and `runner_steps.go` should also be checked for `ContractCoverageFunc` usage and updated if appropriate.

## 4. Mock Providers

- [ ] 4.1 Create `internal/provider/mockprovider/mock.go` with mock implementations: `MockComplexityProvider` (configurable `[]crap.FunctionComplexity`), `MockLineCoverageProvider` (configurable `[]crap.FuncCoverage`), `MockSideEffectAnalyzer` (configurable `[]taxonomy.AnalysisResult`), `MockContractCoverageProvider` (configurable lookup function and degraded packages). Each mock stores its configured data and returns it on the corresponding method call.
- [ ] 4.2 Create `internal/provider/mockprovider/mock_test.go` with tests: (1) Mock complexity + coverage providers produce CRAP scores matching `crap.Formula()` for 3 functions with known values. (2) All four mock providers produce a valid `crap.Report` with correct quadrant classifications and fix strategies for 5 functions spanning all four quadrants. (3) Edge case: empty provider results produce an empty report (no scores, zero-valued summary), not an error. (4) ContractCoverageProvider returning reason `"all_effects_ambiguous"` flows through to `Score.ContractCoverageReason`.
- [ ] 4.3 Create `internal/provider/goprovider/goprovider_test.go` with compile-time interface satisfaction checks: `var _ crap.ComplexityProvider = (*GoComplexityProvider)(nil)`, etc. for all four interfaces. This ensures Go adapters satisfy the interfaces even without running integration tests.

## 5. Non-Regression Verification

- [ ] 5.1 Run `go test -race -count=1 -short ./...` -- all tests MUST pass with zero failures.
- [ ] 5.2 Run `go test -race -count=1 -run TestRunSelfCheck -timeout 30m ./cmd/gaze/...` -- E2E self-check MUST produce identical scores and report output.
- [ ] 5.3 Run `golangci-lint run` -- MUST pass with no new warnings.
- [ ] 5.4 Verify `internal/crap/provider.go` has zero imports from `go/ast`, `go/types`, `go/packages`, `go/ssa`, or `github.com/fzipp/gocyclo`.

## 6. Documentation

- [ ] 6.1 Update `AGENTS.md` -- add `internal/provider/goprovider/` to the architecture section. Add a "Key Patterns" entry for the provider interface pattern. Add to Recent Changes.
- [ ] 6.2 Add GoDoc comments on all new exported types, interfaces, methods, and constructors in `provider.go`, `goprovider/*.go`, and `mockprovider/*.go`.

## 7. Constitution Alignment Verification

- [ ] 7.1 Verify all four principles: Accuracy (byte-identical output with default providers, all tests pass), Minimal Assumptions (no user-facing changes, no new CLI flags or config), Actionable Output (output formats unchanged, reports byte-identical), Testability (each interface independently testable via mocks, universal scoring core testable with synthetic data, coverage strategy documented).