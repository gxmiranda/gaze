## Why

Four `log.Printf` calls in `internal/analysis/mutation.go` and `internal/quality/pairing.go` use the standard library `log` package instead of `github.com/charmbracelet/log`. This violates convention pack rule CS-008 ("Use `github.com/charmbracelet/log` for all application logging. Do not use the standard library `log` package or bare `fmt.Println` for operational output.").

These calls were introduced in spec 021 (SSA panic recovery) before CS-008 was codified and have persisted since. A similar violation in `internal/loader/loader.go` was already fixed in PR #159.

Closes #174.

## What Changes

Replace all four `log.Printf` calls in SSA recovery paths with `charmbracelet/log` structured logging equivalents. Update the `"log"` import in both files to `"github.com/charmbracelet/log"`.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- SSA panic recovery logging: switches from unstructured `log.Printf` to structured `log.Warn`/`log.Debug` with key-value pairs (`"pkg"`, `"panic"`)

### Removed Capabilities
- None

## Impact

- **`internal/analysis/mutation.go`** (lines 50-51): Two `log.Printf` calls in `BuildSSA` replaced with `log.Warn` and `log.Debug`
- **`internal/quality/pairing.go`** (lines 136-137): Two `log.Printf` calls in `BuildTestSSA` replaced with `log.Warn` and `log.Debug`
- Both files change their `"log"` import from stdlib to `"github.com/charmbracelet/log"`
- No functional behavior change — both packages write to stderr by default
- `charmbracelet/log` is already a project dependency (no new dependency added)

## Constitution Alignment

Assessed against the Gaze project constitution (v1.3.0).

### I. Accuracy

**Assessment**: N/A

No change to side effect detection or analysis behavior. Logging format is not part of the accuracy contract.

### II. Minimal Assumptions

**Assessment**: N/A

No new assumptions introduced. No new dependencies (`charmbracelet/log` is already in `go.mod`).

### III. Actionable Output

**Assessment**: PASS

Structured logging with key-value pairs (`"pkg"`, `"panic"`) improves machine-parseability of diagnostic output compared to the previous unstructured `Printf` format strings.

### IV. Testability

**Assessment**: N/A

No testable behavior changes. The `safeSSABuild` recovery pattern and its tests are unaffected. Logging output is not asserted in tests.
