## Why

Phases 1 and 2 of Issue #95 introduced provider interfaces and the external analyzer protocol. During those phases, several backward-compatibility shims and code duplications were introduced intentionally to keep the changes safe and incremental:

1. **Deprecated `ContractCoverageFunc` and `SSADegradedPackages` fields** in `crap.Options` — kept for backward compatibility during Phase 1 transition. All production callers now use `ContractCoverageProvider`. The deprecated fields add confusion and a fallback code path in `crap.Analyze` that is no longer exercised in production.

2. **Duplicated `isMainPkg`** — `internal/provider/goprovider/sideeffect.go` has a local copy despite `loader.IsMainPkg` being the canonical exported version (consolidated by the `analyze-multi-package` change).

3. **`BuildContractCoverageFunc` in `internal/crap/`** — this Go-specific function imports `analysis`, `classify`, `quality`, `loader`, and `go/packages`, keeping the `crap` package coupled to Go-specific tooling. `GoContractCoverageProvider` already wraps it, so it can move to `internal/provider/goprovider/` where the Go-specific code belongs.

Additionally, the taxonomy has Go-specific assumptions that limit multi-language usability:

4. **`GoVersion` field in `Metadata`** — hardcoded field name (`go_version` in JSON). Should be `LanguageVersion` or similar so external analyzers can report their own runtime version (e.g., Python 3.12).

5. **9 Go-specific `SideEffectType` constants** — `ChannelSend`, `ChannelClose`, `DeferredReturnMutation`, `GoroutineSpawn`, `WaitGroupOp`, `RecoverBehavior`, `CgoCall`, `SyncPoolOp`, and `SentinelError` have no equivalents in most languages. External analyzers mapping Python/TypeScript concepts to these types is confusing. The taxonomy needs language-neutral aliases or a mapping layer.

6. **No streaming protocol mode** — the `analyze` method returns all results in a single batch response. For large codebases (10K+ functions), this could produce multi-megabyte JSON responses. A streaming mode (JSONL, one function per line) would allow incremental processing.

This change removes the shims, deduplicates the code, generalizes the taxonomy, and adds streaming support — completing the external analyzer protocol initiative.

## What Changes

### Deprecated Field Removal

Remove `ContractCoverageFunc` and `SSADegradedPackages` from `crap.Options`. Remove the fallback path in `crap.Analyze` that checks these fields. Update all remaining callers (internal tests, mock precedence test) to use `ContractCoverageProvider` exclusively.

### isMainPkg Deduplication

Replace the local `isMainPkg` in `goprovider/sideeffect.go` with a call to `loader.IsMainPkg`.

### BuildContractCoverageFunc Relocation

Move `BuildContractCoverageFunc` and its helpers (`analyzePackageCoverage`, `loadTestPackage`, `classifyResults`, `loadGazeConfigBestEffort`) from `internal/crap/contract.go` to `internal/provider/goprovider/contract.go`. The `GoContractCoverageProvider.Build()` method calls these directly instead of through the `crap` package. This removes Go-specific imports from `internal/crap/`.

### Taxonomy Generalization

Rename `GoVersion` to `LanguageVersion` in the `Metadata` struct (with `language_version` JSON tag). Add a `Language` field to `Metadata` so analyzers can declare what language was analyzed (e.g., `"go"`, `"python"`). The Go analysis path sets `Language: "go"` and `LanguageVersion: runtime.Version()`. External analyzers set their own values via the `initialize` response.

Add language-neutral aliases for the 9 Go-specific `SideEffectType` constants. For example, `AsyncTaskSpawn` as an alias for `GoroutineSpawn`, `FFICall` for `CgoCall`. External analyzers can use either the Go-specific or neutral name. The existing Go names remain valid — this is additive, not a rename.

### Streaming Protocol Mode

Add a `analyze/stream` protocol method as an alternative to the batch `analyze` method. When the analyzer declares `capabilities.streaming: true`, gaze calls `analyze/stream` instead of `analyze`. The analyzer writes one JSON object per line (JSONL format) to stdout, each representing one function's analysis results. Gaze processes results incrementally. This is optional — batch mode remains the default.

## Capabilities

### New Capabilities
- `taxonomy-language-field`: `Metadata.Language` field declaring the analyzed language
- `taxonomy-neutral-aliases`: Language-neutral `SideEffectType` aliases (e.g., `AsyncTaskSpawn`, `FFICall`)
- `streaming-protocol`: `analyze/stream` method for incremental JSONL analysis results

### Modified Capabilities
- `crap.Options`: `ContractCoverageFunc` and `SSADegradedPackages` fields removed
- `internal/crap/`: Go-specific imports (`analysis`, `classify`, `quality`, `loader`, `go/packages`) removed from package
- `Metadata.GoVersion`: renamed to `LanguageVersion` (JSON tag: `language_version`)
- `protocol.Capabilities`: new `Streaming` boolean field

### Removed Capabilities
- `ContractCoverageFunc` callback pattern (replaced by `ContractCoverageProvider` interface)
- `SSADegradedPackages` direct field (degradation data now flows through `ContractCoverageProvider.Build` return value)

## Impact

- **Modified files**: `internal/crap/analyze.go` (remove deprecated fields + fallback), `internal/crap/crap.go` (remove fields from Options), `internal/provider/goprovider/sideeffect.go` (use loader.IsMainPkg), `internal/provider/goprovider/contract.go` (absorb BuildContractCoverageFunc), `internal/taxonomy/types.go` (rename GoVersion, add Language, add neutral aliases), `internal/protocol/types.go` (add Streaming capability), `internal/protocol/client.go` (streaming support), `internal/adapter/sideeffect.go` (consume streaming), `docs/protocol.md` (document streaming + neutral types)
- **Removed file**: `internal/crap/contract.go` (moved to goprovider)
- **Test updates**: `internal/crap/analyze_internal_test.go`, `internal/crap/crap_test.go`, `internal/provider/mockprovider/mock_test.go` (precedence test removed or adapted), `internal/taxonomy/` tests, `internal/protocol/` tests
- **Breaking change to internal API**: `crap.Options` fields removed, `Metadata.GoVersion` renamed. All callers are internal (`cmd/gaze/`, `internal/aireport/`, `internal/provider/`). No external consumers.
- **JSON output change**: `go_version` renamed to `language_version`, new `language` field added. This is a breaking change to the JSON schema for consumers parsing `gaze analyze --format=json` output.

## Constitution Alignment

### I. Accuracy
**Assessment**: PASS — Pure cleanup. No scoring, detection, or classification logic changes. All tests pass.

### II. Minimal Assumptions
**Assessment**: PASS — One JSON output field rename (`go_version` → `language_version`) is an explicit, documented breaking change to the JSON schema. No annotation or restructuring of user code is required. The assumption that consumers will update their field reference is documented in Risk R3.

### III. Actionable Output
**Assessment**: PASS — Output unchanged in structure. The `language_version` field rename and new `language` field improve cross-run comparability for multi-language environments.

### IV. Testability
**Assessment**: PASS — Removing deprecated code paths simplifies the test surface. The mock precedence test (which tested the deprecated fallback) is removed since the fallback no longer exists.