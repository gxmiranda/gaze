## Why

OpenCode documentation canonicalizes `.opencode/commands/` (plural) as the
standard directory for slash commands. The `unbound-force` meta-repo has
already migrated its scaffold engine (`unbound-force/unbound-force#164`),
and OpenSpec has switched (`Fission-AI/OpenSpec#748`).

`gaze init` currently creates command files in `.opencode/command/`
(singular), producing a split-directory situation when used alongside
`uf init` (which now writes to `.opencode/commands/`). While OpenCode's
runtime loads both directories via a `{command,commands}` glob so nothing
is functionally broken, the inconsistency is confusing for users and
contributors.

Closes: https://github.com/unbound-force/gaze/issues/94

## What Changes

1. Rename the embedded assets directory from `internal/scaffold/assets/command/`
   to `internal/scaffold/assets/commands/` — since output paths are derived from
   `embed.FS` directory structure, this automatically changes all `gaze init`
   output paths.
2. Update `isToolOwned` switch cases and GoDoc comments in `scaffold.go` to
   reference `commands/` instead of `command/`.
3. Update internal cross-references in embedded `gaze-fix.md` that point to
   `uf init`-managed command files (now at `commands/` after their migration).
4. Move the 3 live gaze-scaffolded command files from `.opencode/command/` to
   `.opencode/commands/` to keep `TestEmbeddedAssetsMatchSource` passing.
5. Update documentation example output in `docs/`.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `gaze init`: Writes command files to `.opencode/commands/` (plural) instead
  of `.opencode/command/` (singular), aligning with OpenCode's canonical
  directory convention.

### Removed Capabilities
- None

## Impact

- **Scaffold package**: `internal/scaffold/scaffold.go` (path references,
  `isToolOwned`), `internal/scaffold/scaffold_test.go` (path expectations),
  `internal/scaffold/assets/` (directory rename).
- **Live command files**: 3 files moved from `.opencode/command/` to
  `.opencode/commands/` (gaze.md, gaze-fix.md, speckit.testreview.md).
- **Documentation**: `docs/reference/cli/init.md`,
  `docs/guides/opencode-integration.md`, `AGENTS.md` (recent changes section).
- **Not changed**: Historical spec artifacts under `specs/`, non-gaze command
  files in `.opencode/command/` (those belong to `uf init`), no migration
  logic added (handled by `uf init`).

## Constitution Alignment

Assessed against the Gaze project constitution (`.specify/memory/constitution.md`).

### I. Accuracy

**Assessment**: N/A

This change does not alter side effect detection, classification, or any
analysis output. It only changes the filesystem path where scaffold files
are written.

### II. Minimal Assumptions

**Assessment**: PASS

The change aligns with OpenCode's canonical convention, reducing the
assumption that users must understand the singular/plural directory
distinction. No new assumptions are introduced.

### III. Actionable Output

**Assessment**: N/A

Report output, JSON schemas, and all analysis output are unchanged.

### IV. Testability

**Assessment**: PASS

All existing scaffold tests continue to verify the same behavior — only
path strings change. `TestEmbeddedAssetsMatchSource` continues to enforce
byte-identical content between embedded and live copies. No test coverage
is reduced.
