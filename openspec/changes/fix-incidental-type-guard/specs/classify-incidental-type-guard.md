## ADDED Requirements

*None — this change modifies existing behavior, no new requirements added.*

## MODIFIED Requirements

### Requirement: Incidental naming signals MUST be scoped to applicable effect types

Previously: `incidentalPrefixes` was a `[]string` that applied the incidental naming penalty (`-maxNamingWeight`) to all side effect types regardless of semantic relevance.

`incidentalPrefixes` MUST be a typed slice where each entry specifies an `appliesTo []taxonomy.SideEffectType` list. The incidental naming penalty SHALL only be applied when the current effect type matches an entry in the prefix's `appliesTo` list. When the effect type does not match, `AnalyzeNamingSignal` MUST skip the incidental prefix and continue to contractual prefix checks.

The `appliesTo` list for prefixes `log`, `Log`, `debug`, `Debug`, `trace`, `Trace`, `print`, `Print` MUST include exactly: `LogWrite`, `StdoutWrite`, `StderrWrite`.

#### Scenario: P0 ReturnValue on a function with log prefix

- **GIVEN** a function named `LogAndCompute` with a `ReturnValue` side effect
- **WHEN** `AnalyzeNamingSignal("LogAndCompute", taxonomy.ReturnValue)` is called
- **THEN** the returned signal MUST have `Weight == 0` (no incidental penalty applied)

#### Scenario: LogWrite on a function with log prefix

- **GIVEN** a function named `LogAndCompute` with a `LogWrite` side effect
- **WHEN** `AnalyzeNamingSignal("LogAndCompute", taxonomy.LogWrite)` is called
- **THEN** the returned signal MUST have `Weight == -maxNamingWeight` and `Source == "naming"`

#### Scenario: ErrorReturn on a function with debug prefix

- **GIVEN** a function named `DebugAndFetch` with an `ErrorReturn` side effect
- **WHEN** `AnalyzeNamingSignal("DebugAndFetch", taxonomy.ErrorReturn)` is called
- **THEN** the returned signal MUST have `Weight == 0` (no incidental penalty applied)

#### Scenario: StdoutWrite on a function with print prefix

- **GIVEN** a function named `PrintSummary` with a `StdoutWrite` side effect
- **WHEN** `AnalyzeNamingSignal("PrintSummary", taxonomy.StdoutWrite)` is called
- **THEN** the returned signal MUST have `Weight == -maxNamingWeight` and `Source == "naming"`

### Requirement: Incidental godoc signals MUST be scoped to applicable effect types

Previously: `incidentalKeywords` was a `[]string` that applied the incidental godoc penalty (`-maxGodocWeight`) to all side effect types regardless of semantic relevance.

`incidentalKeywords` MUST be a typed slice where each entry specifies an `appliesTo []taxonomy.SideEffectType` list. The incidental godoc penalty SHALL only be applied when the current effect type matches an entry in the keyword's `appliesTo` list. When the effect type does not match, `AnalyzeGodocSignal` MUST skip the incidental keyword and continue to contractual keyword checks.

The `appliesTo` list for keywords `logs`, `prints`, `traces`, `debugs` MUST include exactly: `LogWrite`, `StdoutWrite`, `StderrWrite`.

#### Scenario: P0 ReturnValue when godoc contains "logs"

- **GIVEN** a function with godoc containing "logs the request" and a `ReturnValue` side effect
- **WHEN** `AnalyzeGodocSignal(funcDecl, taxonomy.ReturnValue)` is called
- **THEN** the returned signal MUST have `Weight == 0` (no incidental penalty applied)

#### Scenario: LogWrite when godoc contains "logs"

- **GIVEN** a function with godoc containing "logs the request" and a `LogWrite` side effect
- **WHEN** `AnalyzeGodocSignal(funcDecl, taxonomy.LogWrite)` is called
- **THEN** the returned signal MUST have `Weight == -maxGodocWeight` and `Source == "godoc"`

#### Scenario: StderrWrite when godoc contains "prints"

- **GIVEN** a function with godoc containing "prints diagnostics" and a `StderrWrite` side effect
- **WHEN** `AnalyzeGodocSignal(funcDecl, taxonomy.StderrWrite)` is called
- **THEN** the returned signal MUST have `Weight == -maxGodocWeight` and `Source == "godoc"`

#### Scenario: ReceiverMutation when godoc contains "traces"

- **GIVEN** a function with godoc containing "traces the execution" and a `ReceiverMutation` side effect
- **WHEN** `AnalyzeGodocSignal(funcDecl, taxonomy.ReceiverMutation)` is called
- **THEN** the returned signal MUST have `Weight == 0` (no incidental penalty applied)

## REMOVED Requirements

*None — no requirements are being removed.*
