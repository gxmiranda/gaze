## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: SSA recovery logging uses charmbracelet/log (CS-008 compliance)

All diagnostic logging in SSA panic recovery paths MUST use `github.com/charmbracelet/log` structured logging functions. The standard library `log` package MUST NOT be imported in `internal/analysis/mutation.go` or `internal/quality/pairing.go`.

Previously: SSA recovery paths used `log.Printf` from the standard library `log` package with level prefixes embedded in format strings (e.g., `"warning: ..."`, `"debug: ..."`).

#### Scenario: SSA build panics in analysis package

- **GIVEN** a Go package with types that trigger an SSA builder panic
- **WHEN** `BuildSSA` is called for that package
- **THEN** a warning-level structured log entry MUST be emitted with key `"pkg"` containing the package path
- **AND** a debug-level structured log entry MUST be emitted with keys `"pkg"` and `"panic"` containing the package path and recovered panic value
- **AND** the function MUST return nil (existing behavior, unchanged)

#### Scenario: SSA build panics in quality package

- **GIVEN** a Go test package with types that trigger an SSA builder panic
- **WHEN** `BuildTestSSA` is called for that package
- **THEN** a warning-level structured log entry MUST be emitted with key `"pkg"` containing the package path
- **AND** a debug-level structured log entry MUST be emitted with keys `"pkg"` and `"panic"` containing the package path and recovered panic value
- **AND** the function MUST return a non-nil error (existing behavior, unchanged)

#### Scenario: No stdlib log import remains

- **GIVEN** the completed change
- **WHEN** the import blocks of `internal/analysis/mutation.go` and `internal/quality/pairing.go` are inspected
- **THEN** neither file MUST contain an import of `"log"` (the standard library log package)
- **AND** both files MUST import `"github.com/charmbracelet/log"`

## REMOVED Requirements

None.
