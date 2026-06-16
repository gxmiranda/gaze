## Context

`gaze init` scaffolds 8 files into `.opencode/` via `embed.FS`. Three of
these are command files written to `.opencode/command/` (singular). OpenCode
has canonicalized `.opencode/commands/` (plural) as the standard directory,
and the `unbound-force` meta-repo scaffold engine has already migrated
(`unbound-force/unbound-force#164`). This creates an inconsistency when
`gaze init` and `uf init` are used together in the same project.

The scaffold's output paths are derived entirely from the `embed.FS`
directory structure — renaming `internal/scaffold/assets/command/` to
`internal/scaffold/assets/commands/` automatically changes all output paths
with no code logic changes needed.

## Goals / Non-Goals

### Goals
- `gaze init` writes command files to `.opencode/commands/` (plural)
- Internal cross-references in scaffolded files point to `commands/`
- All scaffold tests pass with updated path expectations
- Live gaze-scaffolded command files in this repo match embedded copies
- Documentation reflects the new output paths

### Non-Goals
- Migration logic for existing `.opencode/command/` files (`uf init` handles this)
- Updating historical spec artifacts under `specs/`
- Moving non-gaze command files in `.opencode/command/` (26 files owned by `uf init`)
- Changing OpenCode runtime behavior (it already loads both directories)

## Decisions

### D1: Rename embed directory, not path-mapping logic

**Decision**: Rename `internal/scaffold/assets/command/` →
`internal/scaffold/assets/commands/` and let the existing `fs.WalkDir`
derive output paths automatically.

**Rationale**: The scaffold uses `strings.TrimPrefix(p, "assets/")` to
compute relative paths and `filepath.Join(targetDir, ".opencode", relPath)`
for output. No path-mapping table or configuration exists. Renaming the
source directory is the simplest change with zero risk of path logic bugs.

### D2: Move only gaze-scaffolded live files

**Decision**: Move 3 files (`gaze.md`, `gaze-fix.md`,
`speckit.testreview.md`) from `.opencode/command/` to `.opencode/commands/`.
Leave the other 26 command files untouched.

**Rationale**: `TestEmbeddedAssetsMatchSource` enforces byte-identical
content between `internal/scaffold/assets/<relPath>` and
`.opencode/<relPath>`. After the directory rename, the test will look for
`.opencode/commands/gaze.md` (plural). The 3 live files must match. The
other 26 files are owned by `uf init` and are outside gaze's scope.

### D3: Update cross-references to `uf init` command paths

**Decision**: Update the 2 cross-references in `gaze-fix.md` that point to
`speckit.implement.md` and `opsx-apply.md` to use the `commands/` path.

**Rationale**: `uf init` has migrated to `commands/` (PR #164 merged).
Forward-looking references should use the canonical path. OpenCode loads
both directories, so the reference works either way at runtime.

### D4: Do not update historical spec artifacts

**Decision**: Leave all `specs/NNN-*/` artifacts unchanged even though they
reference `.opencode/command/`.

**Rationale**: Spec artifacts are historical planning records that document
what was true at the time of writing. Retroactively updating them erodes
their value as a traceable record.

## Risks / Trade-offs

### Split directory during transition

Until `uf init` is run in this repo, `.opencode/` will contain both
`command/` (26 uf-managed files) and `commands/` (3 gaze-managed files).
This is cosmetically inconsistent but functionally harmless — OpenCode loads
both. Running `uf init` will migrate the remaining files.

### Cross-reference fragility

The `gaze-fix.md` references to `.opencode/commands/speckit.implement.md`
and `.opencode/commands/opsx-apply.md` assume `uf init` has migrated those
files. If a project has `uf init` from before PR #164, the files will still
be at `command/` (singular). OpenCode resolves both paths, so the command
will still work — but the reference path won't match the actual file path.
This is acceptable: the migration is one-time and the glob fallback is
reliable.
