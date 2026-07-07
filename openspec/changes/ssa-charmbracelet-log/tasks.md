<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Replace stdlib log with charmbracelet/log

- [x] 1.1 [P] In `internal/analysis/mutation.go`: change import from `"log"` to `"github.com/charmbracelet/log"`. Replace `log.Printf("warning: SSA build skipped for %s: internal panic recovered", pkg.PkgPath)` with `log.Warn("SSA build skipped: internal panic recovered", "pkg", pkg.PkgPath)`. Replace `log.Printf("debug: SSA panic value for %s: %v", pkg.PkgPath, r)` with `log.Debug("SSA panic value", "pkg", pkg.PkgPath, "panic", r)`.
- [x] 1.2 [P] In `internal/quality/pairing.go`: change import from `"log"` to `"github.com/charmbracelet/log"`. Replace `log.Printf("warning: SSA build skipped for %s: internal panic recovered", pkg.PkgPath)` with `log.Warn("SSA build skipped: internal panic recovered", "pkg", pkg.PkgPath)`. Replace `log.Printf("debug: SSA panic value for %s: %v", pkg.PkgPath, r)` with `log.Debug("SSA panic value", "pkg", pkg.PkgPath, "panic", r)`.
- [x] 1.3 In `internal/analysis/mutation_test.go` line 69: update comment from "the log.Printf calls are co-located" to "the log.Warn/log.Debug calls are co-located" to match the updated call sites.

## 2. Verification

- [x] 2.1 Run `go build ./...` to confirm no compilation errors
- [x] 2.2 Run `go test -race -count=1 -short ./internal/analysis/... ./internal/quality/...` to confirm existing tests pass
- [x] 2.3 Run `golangci-lint run ./internal/analysis/... ./internal/quality/...` to confirm no lint violations
- [x] 2.4 Verify no stdlib `"log"` import remains in `internal/analysis/mutation.go` or `internal/quality/pairing.go` (grep check; note: `testdata/` fixtures legitimately use stdlib `log` as analyzed code — these are not violations)
<!-- spec-review: passed -->
<!-- code-review: passed -->
