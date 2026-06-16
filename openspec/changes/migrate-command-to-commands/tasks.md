## 1. Rename embedded assets directory

- [x] 1.1 `git mv internal/scaffold/assets/command/ internal/scaffold/assets/commands/` — rename the embedded assets directory from singular to plural. This automatically changes all `gaze init` output paths since they are derived from `embed.FS` directory structure.

## 2. Update internal cross-references in embedded assets

- [x] 2.1 In `internal/scaffold/assets/commands/gaze-fix.md`, update line 50: `.opencode/command/speckit.implement.md` → `.opencode/commands/speckit.implement.md`.
- [x] 2.2 In `internal/scaffold/assets/commands/gaze-fix.md`, update line 66: `.opencode/command/opsx-apply.md` → `.opencode/commands/opsx-apply.md`.

## 3. Update scaffold Go code

- [x] 3.1 In `internal/scaffold/scaffold.go`, update `isToolOwned` switch cases: `"command/speckit.testreview.md"` → `"commands/speckit.testreview.md"`, `"command/gaze-fix.md"` → `"commands/gaze-fix.md"`.
- [x] 3.2 In `internal/scaffold/scaffold.go`, update GoDoc comments referencing `.opencode/command/` to `.opencode/commands/` (~4 locations: lines 64, 70, 139, 244).

## 4. Update scaffold tests

- [x] 4.1 In `internal/scaffold/scaffold_test.go`, update `TestRun_CreatesFiles` expected paths: `".opencode/command/gaze.md"` → `".opencode/commands/gaze.md"`, `".opencode/command/speckit.testreview.md"` → `".opencode/commands/speckit.testreview.md"` (lines 51-52).
- [x] 4.2 In `internal/scaffold/scaffold_test.go`, update `TestRun_VersionMarker_Dev` prefix check: `"command/"` → `"commands/"` (line 283).
- [x] 4.3 In `internal/scaffold/scaffold_test.go`, update `TestAssetPaths_Returns8Files` expected map keys: `"command/gaze-fix.md"` → `"commands/gaze-fix.md"`, `"command/gaze.md"` → `"commands/gaze.md"`, `"command/speckit.testreview.md"` → `"commands/speckit.testreview.md"` (lines 372-374).
- [x] 4.4 In `internal/scaffold/scaffold_test.go`, update `TestRun_OverwriteOnDiff_ToolOwned` path construction: `"command"` → `"commands"` and asset content key `"command/speckit.testreview.md"` → `"commands/speckit.testreview.md"` (lines 438-490).
- [x] 4.5 In `internal/scaffold/scaffold_test.go`, update `TestRun_OverwriteOnDiff_SkipsIdentical` path construction and expected skip entry: `"command"` → `"commands"` (lines 548-560).
- [x] 4.6 In `internal/scaffold/scaffold_test.go`, update `TestIsToolOwned` test cases: `"command/speckit.testreview.md"` → `"commands/speckit.testreview.md"`, `"command/gaze.md"` → `"commands/gaze.md"`, `"command/speckit.analyze.md"` → `"commands/speckit.analyze.md"` (lines 580-587).

## 5. Move live gaze-scaffolded command files

- [x] 5.1 Create `.opencode/commands/` directory: `mkdir -p .opencode/commands/`.
- [x] 5.2 `git mv .opencode/command/gaze.md .opencode/commands/gaze.md`.
- [x] 5.3 `git mv .opencode/command/gaze-fix.md .opencode/commands/gaze-fix.md`.
- [x] 5.4 `git mv .opencode/command/speckit.testreview.md .opencode/commands/speckit.testreview.md`.
- [x] 5.5 Update live `.opencode/commands/gaze-fix.md` cross-references to match embedded copy (same changes as tasks 2.1 and 2.2).

## 6. Update documentation

- [x] 6.1 In `docs/reference/cli/init.md`, update all occurrences of `command/` to `commands/` — including the directory tree diagram, ownership model section, and example output lines.
- [x] 6.2 In `docs/guides/opencode-integration.md`, update all occurrences of `command/` to `commands/` — including example output, the Commands table, and ownership descriptions.
- [x] 6.3 In `AGENTS.md`, update the `gaze-test-generation` entry in "Recent Changes" that references `.opencode/command/gaze-fix.md` to `.opencode/commands/gaze-fix.md`.

## 7. Verification

- [x] 7.1 Run `go build ./cmd/gaze` — verify build succeeds.
- [x] 7.2 Run `go test -race -count=1 -short ./internal/scaffold/...` — verify all scaffold tests pass.
- [x] 7.3 Run `go test -race -count=1 -short ./...` — verify full test suite passes.
- [x] 7.4 Run `golangci-lint run` — verify no lint issues (1 pre-existing staticcheck issue in ai_mapper.go unrelated to this change).
- [x] 7.5 Constitution alignment: verify Testability (Principle IV) — all existing scaffold tests continue to validate the same behavioral contracts with updated paths; no coverage regression.

<!-- spec-review: passed -->
<!-- code-review: passed -->
