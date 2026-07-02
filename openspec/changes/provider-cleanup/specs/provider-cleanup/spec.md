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

### Requirement: Taxonomy Language Field

The `Metadata` struct MUST include a `Language string` field (JSON tag `language`) declaring the analyzed language. The `GoVersion` field MUST be renamed to `LanguageVersion` (JSON tag `language_version`). Go analysis MUST set `Language: "go"` and `LanguageVersion: runtime.Version()`.

#### Scenario: Go analysis metadata

- **GIVEN** a Go package analyzed with `gaze analyze --format=json`
- **WHEN** the JSON output is parsed
- **THEN** the `metadata` object MUST contain `"language": "go"` and `"language_version"` with a Go version string (e.g., `"go1.25.0"`)

#### Scenario: External analyzer metadata

- **GIVEN** an external analyzer that responds to `initialize` with `language: "python"`
- **WHEN** results flow through the scoring pipeline
- **THEN** the output `metadata` MUST contain `"language": "python"` and `"language_version"` from the analyzer

---

### Requirement: Neutral Side Effect Type Aliases

The `taxonomy` package MUST define language-neutral `SideEffectType` constants as aliases for Go-specific types: `AsyncTaskSpawn` (for `GoroutineSpawn`), `AsyncMessageSend` (for `ChannelSend`), `AsyncChannelClose` (for `ChannelClose`), `BarrierOp` (for `WaitGroupOp`), `PanicRecovery` (for `RecoverBehavior`), `FFICall` (for `CgoCall`), `ObjectPoolOp` (for `SyncPoolOp`). Each alias MUST have the same string value as the Go-specific constant. The Go-specific names MUST NOT be removed.

#### Scenario: Alias equivalence

- **GIVEN** the constants `AsyncTaskSpawn` and `GoroutineSpawn`
- **WHEN** compared
- **THEN** they MUST be equal (`AsyncTaskSpawn == GoroutineSpawn`)

---

### Requirement: Streaming Protocol Mode

The protocol MUST support an optional `analyze/stream` method. When an analyzer declares `capabilities.streaming: true` in the `initialize` response, gaze MUST call `analyze/stream` instead of `analyze`. The analyzer MUST write one JSON object per line (JSONL) to stdout, each representing one `AnalyzedFunction`. Gaze MUST process results incrementally. When `streaming` is false or absent, gaze MUST use the batch `analyze` method.

#### Scenario: Streaming analysis

- **GIVEN** an analyzer with `capabilities.streaming: true`
- **WHEN** gaze calls `analyze/stream`
- **THEN** the analyzer writes JSONL and gaze collects all lines into `[]AnalyzedFunction`

#### Scenario: Streaming not supported

- **GIVEN** an analyzer with `capabilities.streaming: false`
- **WHEN** gaze runs analysis
- **THEN** gaze MUST call the batch `analyze` method (existing behavior)

#### Scenario: Malformed JSONL line during streaming

- **GIVEN** an analyzer streaming JSONL results
- **WHEN** a line contains invalid JSON
- **THEN** gaze MUST report the error with the line number and content, and MUST stop processing (fail-fast, do not skip malformed lines)

---

## REMOVED Requirements

### Requirement: ContractCoverageFunc callback pattern

The `ContractCoverageFunc` field on `crap.Options` is removed. Callers MUST use `ContractCoverageProvider` instead. The removal is justified because all production callers were migrated in Phase 1 (PR #165).

#### Scenario: Compile-time verification

- **GIVEN** a Go source file that references `crap.Options{ContractCoverageFunc: ...}`
- **WHEN** compiled
- **THEN** compilation MUST fail (field does not exist)