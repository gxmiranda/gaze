## Why

`config.LoadFromDir` (`internal/config/config.go:151-158`) silently discards all errors from `config.Load` and returns `DefaultConfig()`. Before PR #154 added threshold validation, the only errors were file-not-found (legitimate fallback) and YAML parse errors (rare). Now that `config.Load` validates threshold ranges, ordering, baseline parameters, and timeout parsing, `LoadFromDir` silently rejects user-authored `.gaze.yaml` files that fail validation — substituting defaults with no warning.

This creates a behavioral inconsistency: commands using `config.Load` directly (`gaze analyze`, `gaze quality`, `gaze docscan`) surface validation errors clearly, while commands using `LoadFromDir` (`gaze crap`, `gaze report`) silently use defaults. A user with `contractual: 500` in `.gaze.yaml` sees an error from `gaze quality` but gets silently wrong results from `gaze crap`.

The impact is not cosmetic. `gaze report` runs in CI with threshold flags (`--max-gaze-crapload`, `--min-contract-coverage`) that evaluate against GazeCRAP scores computed from classification thresholds. Silent default substitution means CI gates evaluate against incorrectly-computed scores — a CI gate integrity issue.

Reported in [#155](https://github.com/unbound-force/gaze/issues/155). Triage confirmed by all five Divisor panel agents (Adversary, Architect, Guard, SRE, Testing) as HIGH severity. Constitutional violations: Principle II (silent assumption) and Principle III (no actionable guidance).

## What Changes

Change `LoadFromDir` to distinguish file-not-found from validation/parse errors. File-not-found remains a silent fallback to `DefaultConfig()` (the config file is optional). Validation and parse errors emit a warning to stderr before falling back, telling the user their config was rejected and why.

The function signature changes from `func LoadFromDir(moduleDir string) *GazeConfig` to `func LoadFromDir(moduleDir string, stderr io.Writer) *GazeConfig`. This follows the project's testable CLI pattern — all stderr output goes through injected `io.Writer`, never `os.Stderr` directly.

All call sites update to pass their existing `stderr` writer. DI function types in `qualityPipelineDeps.loadConfig` and `buildContractCoverageFuncDeps.loadConfig` update to match the new signature.

## Capabilities

### New Capabilities
- `config.LoadFromDir warning output`: When `config.Load` returns a validation or parse error, `LoadFromDir` writes a `warning:` message to the provided `io.Writer` before falling back to `DefaultConfig()`. The warning includes the file path and the error message from `config.Load`.

### Modified Capabilities
- `config.LoadFromDir`: Signature changes to accept `io.Writer` parameter. File-not-found behavior unchanged (silent fallback). Validation/parse error behavior changes from silent fallback to warned fallback.
- `qualityPipelineDeps.loadConfig`: Function type changes from `func(string) *GazeConfig` to `func(string, io.Writer) *GazeConfig` to thread stderr through the DI boundary.
- `buildContractCoverageFuncDeps.loadConfig`: Same function type change as above.

### Removed Capabilities
- None.

## Impact

- **`internal/config/config.go`**: Change `LoadFromDir` signature to accept `io.Writer`, emit warning on errors (file-not-found is already handled by `config.Load`, so any error reaching `LoadFromDir` is always a real error) (~10 lines changed).
- **`internal/config/config_test.go`**: Add `TestLoadFromDir_InvalidConfig_EmitsWarning`, `TestLoadFromDir_MissingFile_NoWarning`, update `TestLoadFromDir_InvalidYAML` to verify warning output (~3 new/modified tests).
- **`internal/aireport/runner_steps.go`**: Update `qualityPipelineDeps.loadConfig` type, update `initDeps` default, update `runDocscanStep` call site to pass `stderr` (~5 lines changed across 3 sites).
- **`internal/provider/goprovider/contract.go`**: Update `buildContractCoverageFuncDeps.loadConfig` type, update default assignment (~3 lines changed).
- **`cmd/gaze/main.go`**: Update direct `LoadFromDir` call sites in `initExternalSession`, `resolveBaselinePath`, `loadAndCompare` to pass `stderr` (~3 lines changed).
- **No API surface changes beyond `LoadFromDir`**: All changes are internal. The `config.Load` function is unchanged. The `DefaultConfig()` fallback behavior is preserved — only the silence is broken.

## Constitution Alignment

Assessed against the Gaze project constitution (v1.3.0).

### I. Accuracy

**Assessment**: PASS

This change does not alter classification, scoring, or analysis logic. The fallback to `DefaultConfig()` is preserved. The change adds observability (warnings) without changing computed results.

### II. Minimal Assumptions

**Assessment**: PASS (remediates existing violation)

The current behavior violates this principle: `LoadFromDir` silently ignores validation errors, making an implicit assumption that the user intended defaults. The fix makes the assumption explicit by warning the user that their config was rejected and defaults are in use. File-not-found remains a legitimate silent fallback (the config file is optional by design).

### III. Actionable Output

**Assessment**: PASS (remediates existing violation)

The warning message names the config file path and includes the validation error from `config.Load`, giving the user specific, actionable information: which file failed, why it failed, and what defaults are being used instead.

### IV. Testability

**Assessment**: PASS

`LoadFromDir` accepts `io.Writer` for warning output, making warnings testable without subprocess execution. Each test case captures output to a `bytes.Buffer` and asserts on content. The DI function types propagate the `io.Writer` through the dependency injection boundary, maintaining testability at all call sites.
