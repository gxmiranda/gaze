## ADDED Requirements

### Requirement: LoadFromDir stderr parameter

`config.LoadFromDir` MUST accept an `io.Writer` parameter for diagnostic output. The function signature SHALL change from `func LoadFromDir(moduleDir string) *GazeConfig` to `func LoadFromDir(moduleDir string, stderr io.Writer) *GazeConfig`.

#### Scenario: Signature includes io.Writer
- **GIVEN** a caller invoking `config.LoadFromDir`
- **WHEN** the caller provides a `moduleDir` and an `io.Writer`
- **THEN** the function compiles and operates as before for valid configs

### Requirement: Warning on config validation error

`config.LoadFromDir` MUST write a warning message to the provided `io.Writer` when `config.Load` returns an error. The warning MUST include the config file path and the error message. The function MUST still return `DefaultConfig()` after emitting the warning.

#### Scenario: Invalid classification threshold
- **GIVEN** a `.gaze.yaml` file with `contractual: 500` in the module directory
- **WHEN** `LoadFromDir` is called with that directory and a `bytes.Buffer` as `io.Writer`
- **THEN** the buffer contains a warning message including the file path and "must be in [1, 99]"
- **AND** the returned config equals `DefaultConfig()`

#### Scenario: Malformed YAML
- **GIVEN** a `.gaze.yaml` file with unparseable YAML content in the module directory
- **WHEN** `LoadFromDir` is called with that directory and a `bytes.Buffer` as `io.Writer`
- **THEN** the buffer contains a warning message including the file path and the parse error
- **AND** the returned config equals `DefaultConfig()`

### Requirement: No warning on missing config

`config.LoadFromDir` MUST NOT write any output when the `.gaze.yaml` file does not exist. Missing config files are a normal condition (the file is optional).

#### Scenario: No config file present
- **GIVEN** a module directory with no `.gaze.yaml` file
- **WHEN** `LoadFromDir` is called with that directory and a `bytes.Buffer` as `io.Writer`
- **THEN** the buffer is empty
- **AND** the returned config equals `DefaultConfig()`

### Requirement: No warning on valid config

`config.LoadFromDir` MUST NOT write any output when the `.gaze.yaml` file is valid. Warnings are reserved for error conditions.

#### Scenario: Valid config file present
- **GIVEN** a module directory with a valid `.gaze.yaml` file
- **WHEN** `LoadFromDir` is called with that directory and a `bytes.Buffer` as `io.Writer`
- **THEN** the buffer is empty
- **AND** the returned config reflects the values from the file

## MODIFIED Requirements

### Requirement: qualityPipelineDeps.loadConfig function type

The `loadConfig` field in `qualityPipelineDeps` SHALL change from `func(string) *config.GazeConfig` to `func(string, io.Writer) *config.GazeConfig`. All callers MUST pass the available `io.Writer` when invoking `loadConfig`.

Previously: `loadConfig func(string) *config.GazeConfig`

#### Scenario: DI default uses updated LoadFromDir
- **GIVEN** a `qualityPipelineDeps` with nil `loadConfig`
- **WHEN** `resolveQualityDeps` assigns the default
- **THEN** the default is `config.LoadFromDir` (matching the new 2-parameter signature)

### Requirement: buildContractCoverageFuncDeps.loadConfig function type

The `loadConfig` field in `buildContractCoverageFuncDeps` SHALL change from `func(string) *config.GazeConfig` to `func(string, io.Writer) *config.GazeConfig`. All callers MUST pass the available `io.Writer` when invoking `loadConfig`.

Previously: `loadConfig func(string) *config.GazeConfig`

### Requirement: runDocscanStep stderr parameter

`runDocscanStep` SHALL accept an additional `stderr io.Writer` parameter and pass it to `config.LoadFromDir`.

Previously: `func runDocscanStep(moduleDir string) (json.RawMessage, error)`

#### Scenario: Docscan step propagates config warning
- **GIVEN** a module directory with an invalid `.gaze.yaml`
- **WHEN** `runDocscanStep` is called with that directory and a `bytes.Buffer`
- **THEN** the warning from `LoadFromDir` appears in the buffer

### Requirement: cmd/gaze call sites pass stderr

All direct `config.LoadFromDir` call sites in `cmd/gaze/main.go` (`initExternalSession`, `resolveBaselinePath`, `loadAndCompare`) MUST pass an `io.Writer` to `LoadFromDir`. Functions that lack an `io.Writer` parameter MUST be updated to accept one.

Previously: `config.LoadFromDir(moduleDir)` with no `io.Writer` argument.

#### Scenario: resolveBaselinePath propagates config warning
- **GIVEN** a `.gaze.yaml` with invalid baseline epsilon in the module directory
- **WHEN** `resolveBaselinePath` is called with a `bytes.Buffer` as stderr
- **THEN** the warning from `LoadFromDir` appears in the buffer

## REMOVED Requirements

None.
