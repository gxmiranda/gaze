## 1. Deprecated Field Removal

- [x] 1.1 Remove `ContractCoverageFunc` and `SSADegradedPackages` fields from `crap.Options` in `internal/crap/analyze.go`. Remove the `// Deprecated:` GoDoc comments. Remove the fallback dispatch path in `crap.Analyze` (lines 139-155) that checks `ContractCoverageProvider` then falls back to `ContractCoverageFunc`. Remove the `ContractCoverageFunc` call in `computeScores` (lines 292-293) — contract coverage now comes exclusively through `ContractCoverageProvider` which sets up the scoring path via its `Build()` return value.
- [x] 1.2 Update `internal/crap/analyze_internal_test.go` — replace all `opts.ContractCoverageFunc = ...` usages with mock `ContractCoverageProvider` implementations. There are 4 occurrences (lines 154, 221, 376, 413).
- [x] 1.3 Update `internal/crap/crap_test.go` — replace `testProviderOpts()` or any test helper that sets `ContractCoverageFunc`. Update the `lookupCoverage` and `Analyze` tests.
- [x] 1.4 Remove `TestMockProviders_ProviderPrecedence` from `internal/provider/mockprovider/mock_test.go` — this test validates precedence between `ContractCoverageProvider` and `ContractCoverageFunc`, which no longer exists. Remove the `//nolint:staticcheck` directive that was on the deprecated field access.

## 2. isMainPkg Deduplication

- [x] 2.1 In `internal/provider/goprovider/sideeffect.go`, replace the local `isMainPkg` function with a call to `loader.IsMainPkg`. Remove the local function definition and its `// NOTE: keep in sync` comment. Add `loader` to the import if not already present.

## 3. BuildContractCoverageFunc Relocation

- [x] 3.1 Move `BuildContractCoverageFunc`, `analyzePackageCoverage`, `loadTestPackage`, `classifyResults`, and `loadGazeConfigBestEffort` from `internal/crap/contract.go` to `internal/provider/goprovider/contract.go`. Update `GoContractCoverageProvider.Build()` to call the moved functions directly instead of through `crap.BuildContractCoverageFunc`. Keep the function signatures identical.
- [x] 3.2 Remove `internal/crap/contract.go` (now empty after the move). Verify `internal/crap/` no longer imports `internal/analysis`, `internal/classify`, `internal/quality`, `internal/loader`, or `golang.org/x/tools/go/packages`.
- [x] 3.3 Update any callers of `crap.BuildContractCoverageFunc` outside of `goprovider/` — search `cmd/gaze/main.go` and `internal/aireport/` for remaining references. These should now call through `GoContractCoverageProvider` or the moved function in `goprovider`.

## 4. Taxonomy Generalization

- [x] 4.1 Rename `GoVersion` to `LanguageVersion` in `internal/taxonomy/types.go` (JSON tag `language_version`). Add `Language string` field (JSON tag `language`). Update `MarshalJSON`/`UnmarshalJSON` if they handle these fields.
- [x] 4.2 Update Go analysis paths to set `Language: "go"` and `LanguageVersion: runtime.Version()`: `internal/analysis/analyzer.go` and `internal/quality/quality.go`.
- [x] 4.3 Update `internal/adapter/sideeffect.go` to set `Language` and `LanguageVersion` from the `initialize` response's `language` field.
- [x] 4.4 Update JSON Schema in `internal/report/schema.go` — rename `go_version` to `language_version`, add `language` field.
- [x] 4.5 Add neutral `SideEffectType` aliases to `internal/taxonomy/types.go`: `AsyncTaskSpawn = GoroutineSpawn`, `AsyncMessageSend = ChannelSend`, `AsyncChannelClose = ChannelClose`, `BarrierOp = WaitGroupOp`, `PanicRecovery = RecoverBehavior`, `FFICall = CgoCall`, `ObjectPoolOp = SyncPoolOp`. Add GoDoc comments noting these are language-neutral aliases.
- [x] 4.6 Add tests verifying alias equivalence (`AsyncTaskSpawn == GoroutineSpawn` etc.) in `internal/taxonomy/`.

## 5. Streaming Protocol Mode

- [x] 5.1 Add `Streaming bool` to `protocol.Capabilities` in `internal/protocol/types.go`. Add `MethodAnalyzeStream = "analyze/stream"` constant.
- [x] 5.2 Add `CallStream(ctx context.Context, method string, params any) (*bufio.Scanner, error)` to `internal/protocol/client.go` — sends the request, returns a scanner for reading JSONL lines from stdout. The caller reads lines until EOF or context cancellation.
- [x] 5.3 Update `ExternalSideEffectAnalyzer` in `internal/adapter/sideeffect.go` — when `Capabilities.Streaming` is true, call `analyze/stream` via `CallStream` and collect `AnalyzedFunction` objects line by line. When false, use existing batch `Call`.
- [x] 5.4 Add `--hang-stream` mode to fake analyzer (`internal/protocol/testdata/fake_analyzer/main.go`) that writes JSONL for `analyze/stream`.
- [x] 5.5 Add tests: (1) streaming session produces same results as batch, (2) streaming with context timeout cancels mid-stream, (3) malformed JSONL line is reported as error.
- [x] 5.6 Update `docs/protocol.md` with `analyze/stream` method documentation and JSONL format specification.

## 6. Non-Regression Verification

- [x] 6.1 Run `go build ./...` — MUST compile with zero errors.
- [x] 6.2 Run `go test -race -count=1 -short ./...` — all tests MUST pass.
- [ ] 6.3 Run `go test -race -count=1 -run TestRunSelfCheck -timeout 30m ./cmd/gaze/...` — E2E self-check MUST produce identical output.
- [x] 6.4 Verify `internal/crap/` import list contains zero Go-specific analysis imports.

## 7. Documentation

- [x] 7.1 Update `AGENTS.md` — add to Recent Changes. Note `BuildContractCoverageFunc` relocation, taxonomy generalization, and streaming protocol.
- [x] 7.2 Update GoDoc comments on any moved or modified exported functions.
- [x] 7.3 Update `docs/protocol.md` with neutral side effect type aliases and `language` / `language_version` fields.

## 8. Constitution Alignment Verification

- [x] 8.1 Verify all four principles: Accuracy (no scoring changes, E2E passes), Minimal Assumptions (no user-facing changes beyond JSON field rename), Actionable Output (output improved with language field), Testability (deprecated test paths removed, streaming tested with fake analyzer).

<!-- spec-review: passed -->