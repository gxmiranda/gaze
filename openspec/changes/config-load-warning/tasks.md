<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Core: Change `LoadFromDir` signature and behavior

- [x] 1.1 Change `LoadFromDir` in `internal/config/config.go` to accept `stderr io.Writer` as second parameter. When `config.Load` returns an error, write `warning: config %q rejected, using defaults: %v\n` to `stderr` before returning `DefaultConfig()`. Update the doc comment to document the warning behavior.
- [x] 1.2 Update existing tests in `internal/config/config_test.go` to pass `io.Writer` (e.g., `io.Discard` or `bytes.Buffer`) to `LoadFromDir`. Add 4 new tests: `TestLoadFromDir_InvalidConfig_EmitsWarning` (contractual=500), `TestLoadFromDir_MalformedYAML_EmitsWarning` (unparseable YAML), `TestLoadFromDir_MissingFile_NoWarning` (empty dir, buffer empty), `TestLoadFromDir_ValidConfig_NoWarning` (valid config, buffer empty).

## 2. Update call sites

- [x] 2.1 [P] Update `qualityPipelineDeps.loadConfig` type in `internal/aireport/runner_steps.go` from `func(string) *config.GazeConfig` to `func(string, io.Writer) *config.GazeConfig`. Update `resolveQualityDeps` default assignment. Update all invocations of `d.loadConfig` (in `runQualityStep` and `runClassifyStep`) to pass `stderr`. Update `runDocscanStep` signature to accept `stderr io.Writer` and pass it to `config.LoadFromDir`. Update `pipelineStepFuncs.docscanStep` function type in `runner.go` from `func(string) (json.RawMessage, error)` to `func(string, io.Writer) (json.RawMessage, error)`. Update the `docscanStep` assignment in `runProductionPipeline` to pass `opts.Stderr`.
- [x] 2.2 [P] Update `buildContractCoverageFuncDeps.loadConfig` type in `internal/provider/goprovider/contract.go` from `func(string) *config.GazeConfig` to `func(string, io.Writer) *config.GazeConfig`. Update `buildContractCoverageFuncImpl` default assignment. Update the invocation of `loadCfg` to pass `stderr`.
- [x] 2.3 [P] Update `config.LoadFromDir` call sites in `cmd/gaze/main.go`: `initExternalSession` (line 653 — already has `stderr io.Writer`), `resolveBaselinePath` (line 726 — add `stderr io.Writer` parameter), `loadAndCompare` (line 777 — add `stderr io.Writer` parameter). Update callers of `resolveBaselinePath` and `loadAndCompare` to pass `p.stderr`.

## 3. Test updates for call site changes

- [x] 3.1 [P] Update any existing tests for `runDocscanStep` or `qualityPipelineDeps` in `internal/aireport/` to match the new function signatures (pass `io.Writer` to DI mocks and direct calls). Update `pipelineStepFuncs` fakes in `pipeline_internal_test.go` (including `fakeSteps()` and all individual test overrides of `steps.docscanStep`) to match the new `func(string, io.Writer) (json.RawMessage, error)` type.
- [x] 3.2 [P] Update any existing tests for `buildContractCoverageFuncDeps` in `internal/provider/goprovider/` to match the new `loadConfig` type (pass `io.Writer` to DI mocks).
- [x] 3.3 [P] Update any existing tests for `resolveBaselinePath` or `loadAndCompare` in `cmd/gaze/` to match the new signatures.

## 4. Verification

- [x] 4.1 Run `go build ./cmd/gaze` — all call sites compile.
- [x] 4.2 Run `go test -race -count=1 -short ./...` — all tests pass.
- [x] 4.3 Run `golangci-lint run` — no lint violations.
- [x] 4.4 Constitution alignment verification: Confirm Principle II (no silent assumptions — warning makes fallback explicit), Principle III (actionable output — warning includes file path and error), Principle IV (testability — all warnings captured via `io.Writer`, no `os.Stderr`).

<!-- spec-review: passed -->
<!-- code-review: passed -->
