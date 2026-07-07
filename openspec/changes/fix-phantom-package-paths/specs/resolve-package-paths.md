## ADDED Requirements

### Requirement: Phantom Package Filtering

`loader.ResolvePackagePaths` MUST skip packages where `len(pkg.Errors) > 0` and MUST NOT include their `PkgPath` in the returned slice. This prevents synthetic packages returned by `go/packages.Load` for nonexistent patterns from being treated as valid package paths.

#### Scenario: Nonexistent package pattern

- **GIVEN** `ResolvePackagePaths` is called with pattern `"github.com/nonexistent/does/not/exist"`
- **WHEN** `go/packages.Load` returns a synthetic package with `PkgPath` set to the pattern and `Errors` populated
- **THEN** the returned slice MUST be empty and no error MUST be returned

#### Scenario: Mix of valid and invalid patterns

- **GIVEN** `ResolvePackagePaths` is called with patterns `["github.com/unbound-force/gaze/internal/loader", "github.com/nonexistent/pkg"]`
- **WHEN** `go/packages.Load` returns one valid package and one errored package
- **THEN** the returned slice MUST contain only the valid package path and a warning MUST be written to `stderr` for the invalid pattern

### Requirement: Warning Output for Skipped Packages

`loader.ResolvePackagePaths` MUST accept an `io.Writer` parameter for diagnostic output. When a package is skipped due to load errors, the function MUST write a warning line per error to the writer in the format: `warning: skipping <PkgPath>: <error message>`.

#### Scenario: Warning output for invalid pattern

- **GIVEN** `ResolvePackagePaths` is called with a nonexistent pattern and a non-nil `io.Writer`
- **WHEN** the package is skipped due to load errors
- **THEN** the writer MUST contain a warning line including the package path and error detail

#### Scenario: Nil stderr writer

- **GIVEN** `ResolvePackagePaths` is called with `stderr` as `nil`
- **WHEN** a package has load errors
- **THEN** the package MUST still be skipped and no panic MUST occur

## MODIFIED Requirements

### Requirement: ResolvePackagePaths Signature

`loader.ResolvePackagePaths` signature changes from:

```go
func ResolvePackagePaths(patterns []string, moduleDir string) ([]string, error)
```

to:

```go
func ResolvePackagePaths(patterns []string, moduleDir string, stderr io.Writer) ([]string, error)
```

Previously: The function accepted only patterns and moduleDir. The new `io.Writer` parameter enables diagnostic output without global logger dependency.

## REMOVED Requirements

None.
