## Context

`config.LoadFromDir` (`internal/config/config.go:151-158`) is a best-effort config loader that calls `config.Load` and returns `DefaultConfig()` on any error. Since PR #154 added validation to `config.Load` (threshold ranges, ordering, baseline parameters, timeout parsing), `LoadFromDir` now silently discards user-authored config that fails validation. Users see no indication that their `.gaze.yaml` was rejected.

This is inconsistent with the direct `config.Load` callers (`gaze analyze`, `gaze quality`, `gaze docscan`), which surface validation errors. It also creates a CI gate integrity risk: `gaze report` and `gaze crap` use `LoadFromDir` internally, so classification thresholds and baseline parameters may silently differ from the user's intent.

Proposal alignment: Principles II (Minimal Assumptions — silent config rejection) and III (Actionable Output — no guidance to user) are violated.

## Goals / Non-Goals

### Goals
- Emit a `warning:` message to stderr when `LoadFromDir` falls back to defaults due to a validation or parse error
- Distinguish file-not-found (silent fallback — config is optional) from real errors (warned fallback)
- Thread `io.Writer` through `LoadFromDir` following the project's testable CLI pattern
- Update all call sites and DI function types to pass stderr
- Add regression tests for warning output

### Non-Goals
- Changing the fallback behavior itself — `LoadFromDir` still returns `DefaultConfig()` on error
- Making `LoadFromDir` return an error — that would change the function's contract and require callers to handle errors (larger change, tracked as future work under #139)
- Adding validation logic — PR #154 already handles validation in `config.Load`
- Consolidating duplicate config loading patterns — tracked separately in #139

## Decisions

### D1: Add `io.Writer` parameter to `LoadFromDir`, not change the return type

**Decision**: Change `LoadFromDir(moduleDir string)` to `LoadFromDir(moduleDir string, stderr io.Writer)`. Write warning to `stderr` and still return `*GazeConfig`.

**Alternatives considered**:
- Return `(*GazeConfig, error)`: Most architecturally sound long-term, but changes the function's contract. All 6 callers would need error handling logic. Better suited for the #139 consolidation effort.
- Use `log.Warn` via the package logger: Avoids signature change, but the project does not use a package logger in `internal/config`. The warning pattern in this codebase uses injected `io.Writer` (21+ instances in `cmd/gaze/main.go`).
- Write to `os.Stderr` directly: Violates the testable CLI pattern. Would be the only direct `os.Stderr` reference in production code.

**Rationale**: The `io.Writer` approach is the smallest change that follows existing conventions. It preserves the best-effort return contract while adding observability. Compatible with a future migration to `(*GazeConfig, error)` under #139.

### D2: No `os.IsNotExist` check needed in `LoadFromDir`

**Decision**: `LoadFromDir` does not need to add its own file-not-found check.

**Rationale**: `config.Load` already handles `os.IsNotExist` at line 166-168, returning `(DefaultConfig(), nil)`. Any error that reaches `LoadFromDir`'s `if err != nil` branch is always a real error (YAML parse failure, validation failure, or permission error), never file-not-found. The warning is always appropriate for any error that reaches this branch.

### D3: Warning message format

**Decision**: Use `fmt.Fprintf(stderr, "warning: config %q rejected, using defaults: %v\n", cfgPath, err)`.

**Rationale**: Matches the project's existing `warning:` prefix pattern (used 20+ times). Includes the file path (actionable — user knows which file to fix) and the error message from `config.Load` (specific — user knows what's wrong). The `using defaults` phrasing makes the fallback behavior explicit.

### D4: DI function type changes

**Decision**: Update the `loadConfig` function type in both `qualityPipelineDeps` and `buildContractCoverageFuncDeps` from `func(string) *config.GazeConfig` to `func(string, io.Writer) *config.GazeConfig`.

**Rationale**: The DI types must match the new `LoadFromDir` signature for the default assignment to compile. Both DI structs already have `stderr io.Writer` or `io.Writer` available in their callers, so threading it through adds no complexity.

### D5: `runDocscanStep` signature change

**Decision**: Add `stderr io.Writer` parameter to `runDocscanStep` and pass it to `LoadFromDir`.

**Rationale**: `runDocscanStep` is the only call site that uses `LoadFromDir` directly (not through DI). It currently has no `stderr` parameter. The caller (`runProductionPipeline`) has `opts.Stderr` available.

## Coverage Strategy

Unit tests only. All test cases use `bytes.Buffer` to capture warning output, following the testable CLI pattern. No integration or e2e tests needed — `LoadFromDir` is a pure function with file I/O as its only external dependency. The 4 new tests cover all branches of the modified `LoadFromDir` (error path with warning, no-error path, file-not-found via `Load`). Expected branch coverage of modified code: 100%.

Required tests:
1. **`TestLoadFromDir_InvalidConfig_EmitsWarning`**: Write `.gaze.yaml` with `contractual: 500`, verify warning emitted to `io.Writer` and `DefaultConfig()` returned.
2. **`TestLoadFromDir_MalformedYAML_EmitsWarning`**: Write unparseable YAML, verify warning emitted.
3. **`TestLoadFromDir_MissingFile_NoWarning`**: Empty temp dir, verify no warning and `DefaultConfig()` returned.
4. **`TestLoadFromDir_ValidConfig_NoWarning`**: Valid `.gaze.yaml`, verify no warning and config returned.

Existing tests (`TestLoadFromDir_ValidConfig`, `TestLoadFromDir_MissingFile`, `TestLoadFromDir_InvalidYAML`) update to pass `io.Writer` parameter.

## Risks / Trade-offs

### R1: Signature change breaks all callers (Low risk, contained)

All 6 call sites must update to pass `stderr`. This is mechanical — each site already has an `io.Writer` available in scope or in its caller. The compiler catches any missed sites.

### R2: Warning may be noisy for CI users who intentionally use defaults (Very low risk)

A user who has no `.gaze.yaml` sees no warning (file-not-found is handled in `Load`). A user with an invalid `.gaze.yaml` deserves the warning — they authored a config file and it's being silently rejected. The warning is emitted once per `LoadFromDir` call, not per-function.

### R3: Multiple warnings per command invocation (Accepted trade-off)

`gaze crap` calls `LoadFromDir` up to 3 times (lines 653, 726, 777). With an invalid config, the user sees 3 warnings. This is acceptable because each call site uses the config for a different purpose (analyzer discovery, baseline path, comparison options), and seeing the warning repeated reinforces that the config is being rejected at multiple points. A future consolidation (#139) would reduce this to one call.
