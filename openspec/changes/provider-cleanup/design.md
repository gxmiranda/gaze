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
- Generalize taxonomy: rename `GoVersion` → `LanguageVersion`, add `Language` field, add neutral side effect type aliases
- Add streaming protocol mode (`analyze/stream` method with JSONL output)

### Non-Goals
- Removing the `crap.BuildContractCoverageFunc` export entirely — it may have test utility; move it, don't delete the symbol
- Changing existing Go-specific `SideEffectType` constant names — aliases are additive
- Making streaming mandatory — batch remains the default

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

### D5: Taxonomy generalization — rename GoVersion, add Language

Rename `Metadata.GoVersion` to `Metadata.LanguageVersion` (JSON tag `language_version`). Add `Metadata.Language string` (JSON tag `language`). The Go analysis path sets `Language: "go"` and `LanguageVersion: runtime.Version()`. External analyzers set values from the `initialize` response (`analyzer_name`, `language` fields).

This is a JSON schema breaking change — `go_version` becomes `language_version`. Consumers parsing the JSON output will need to update. The `gaze analyze --format=json` and `gaze quality --format=json` schemas both include `Metadata`.

### D6: Neutral side effect type aliases

Add language-neutral constants alongside Go-specific ones. The neutral names map to the same string values:

| Go-Specific | Neutral Alias | Shared Value |
|-------------|---------------|--------------|
| `GoroutineSpawn` | `AsyncTaskSpawn` | `"GoroutineSpawn"` |
| `ChannelSend` | `AsyncMessageSend` | `"ChannelSend"` |
| `ChannelClose` | `AsyncChannelClose` | `"ChannelClose"` |
| `DeferredReturnMutation` | (no alias — Go-only concept) | `"DeferredReturnMutation"` |
| `WaitGroupOp` | `BarrierOp` | `"WaitGroupOp"` |
| `RecoverBehavior` | `PanicRecovery` | `"RecoverBehavior"` |
| `CgoCall` | `FFICall` | `"CgoCall"` |
| `SyncPoolOp` | `ObjectPoolOp` | `"SyncPoolOp"` |
| `SentinelError` | (no alias — Go idiom) | `"SentinelError"` |

The string values stay the same so existing JSON output is unchanged. The aliases are convenience constants for external analyzer authors. `DeferredReturnMutation` and `SentinelError` have no neutral alias because they're uniquely Go concepts with no cross-language equivalent.

### D7: Streaming protocol — analyze/stream method

Add a new optional protocol method `analyze/stream`. When declared in `capabilities.streaming: true`, gaze calls `analyze/stream` instead of `analyze`. The response is JSONL (one JSON object per line on stdout), each representing one `AnalyzedFunction`. The protocol client reads lines incrementally and converts to `[]taxonomy.AnalysisResult`.

The `ExternalSideEffectAnalyzer` adapter checks the `Streaming` capability flag and calls the appropriate method. The batch `analyze` method remains available as the default.

Stream format per line:
```json
{"name":"add","package":"math","file":"add.go","line":10,"side_effects":[...]}
```

Each line is a complete, self-contained `AnalyzedFunction` object (same schema as the `functions` array elements in the batch response).

## Risks / Trade-offs

### R1: Internal API break

Removing `crap.Options` fields is a breaking change to the internal API. All callers are within the module (`cmd/gaze/`, `internal/aireport/`, `internal/provider/`). No external consumers exist (`internal/` packages cannot be imported).

### R2: Large file move

Moving `contract.go` from `crap/` to `goprovider/` is a large diff that may be hard to review. Mitigated by keeping the function signatures identical — only the package location changes.

### R3: JSON schema break (GoVersion → LanguageVersion)

Renaming `go_version` to `language_version` in JSON output breaks consumers parsing this field. Mitigated by: (a) this is a minor version bump (the field is metadata, not scoring data), (b) the new field name is more accurate for multi-language use.

### R4: Streaming complexity

The streaming protocol adds complexity to the protocol client and side effect adapter. Mitigated by keeping it optional — batch remains the default, and analyzers that don't declare `streaming: true` are unaffected.