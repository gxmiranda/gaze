## Context

The SSA panic recovery paths in `internal/analysis/mutation.go` and `internal/quality/pairing.go` use `log.Printf` from the standard library `log` package. Convention pack rule CS-008 requires all application logging to use `github.com/charmbracelet/log`. These four calls were introduced in spec 021 before CS-008 was codified. A similar fix in `internal/loader/loader.go` was completed in PR #159 and serves as a precedent.

## Goals / Non-Goals

### Goals
- Replace all four stdlib `log.Printf` calls with `charmbracelet/log` structured equivalents
- Update import statements in both affected files
- Maintain identical diagnostic behavior (same log levels, same information conveyed)

### Non-Goals
- Refactoring the `safeSSABuild` recovery pattern itself
- Adding log-level configuration or filtering
- Changing the log output destination (both packages already write to stderr)
- Modifying any test code or test assertions for logging output

## Decisions

**D1: Use package-level `log.Warn`/`log.Debug` functions, not a logger instance**

Both affected files currently use `log.Printf` at the package level. The replacement will use `charmbracelet/log`'s package-level functions (`log.Warn`, `log.Debug`) which is consistent with how `charmbracelet/log` is used elsewhere in the project (e.g., the fix in `internal/loader/loader.go`).

Rationale: No need to thread a logger through function signatures for diagnostic logging in recovery paths. The `charmbracelet/log` default logger writes to stderr, matching the stdlib `log` behavior.

**D2: Map log levels from Printf prefixes to structured calls**

The current calls embed level in the format string (`"warning: ..."`, `"debug: ..."`). The replacement uses the appropriate structured method:

| Current | Replacement |
|---------|-------------|
| `log.Printf("warning: SSA build skipped for %s: ...")` | `log.Warn("SSA build skipped: internal panic recovered", "pkg", pkg.PkgPath)` |
| `log.Printf("debug: SSA panic value for %s: %v")` | `log.Debug("SSA panic value", "pkg", pkg.PkgPath, "panic", r)` |

Rationale: Structured key-value pairs improve machine-parseability (Observable Quality principle) and let users filter by log level.

**D3: Import alias not needed**

Both files currently import `"log"`. The replacement import `"github.com/charmbracelet/log"` also resolves to the package name `log`, so no alias is required and all call sites remain `log.Xxx(...)`.

## Coverage Strategy

No new tests required. Existing `safeSSABuild` tests in both packages verify the recovery behavior. The logging format change is not asserted in tests (operational output, not contractual behavior). Verification: existing test suite passes unchanged (task 2.2), plus grep check confirms no stdlib `"log"` import remains (task 2.4).

## Risks / Trade-offs

**Low risk**: The stdlib `log` and `charmbracelet/log` both default to stderr output. The only observable difference is the output format (structured vs. unstructured), which is an improvement. No tests assert on the format of these log messages.

**No dependency risk**: `charmbracelet/log` is already in `go.mod` — no new dependencies are introduced.
