## ADDED Requirements

None — no new functionality is introduced.

## MODIFIED Requirements

### Requirement: Scaffold output directory for command files

`gaze init` MUST write command files to `.opencode/commands/` (plural)
instead of `.opencode/command/` (singular).

Previously: `gaze init` wrote command files to `.opencode/command/`
(singular).

#### Scenario: Fresh project initialization

- **GIVEN** a Go project with no `.opencode/` directory
- **WHEN** the user runs `gaze init`
- **THEN** command files MUST be created under `.opencode/commands/`
  (e.g., `.opencode/commands/gaze.md`, `.opencode/commands/gaze-fix.md`,
  `.opencode/commands/speckit.testreview.md`)

#### Scenario: Re-initialization with existing files at old path

- **GIVEN** a Go project with command files at `.opencode/command/gaze.md`
  from a previous `gaze init` run
- **WHEN** the user runs `gaze init`
- **THEN** new command files MUST be created under `.opencode/commands/`
- **AND** the old files at `.opencode/command/` MUST NOT be modified or
  removed (migration is handled by `uf init`, not `gaze init`)

### Requirement: Tool-owned file identification

The `isToolOwned` function MUST recognize tool-owned command files under the
`commands/` prefix (plural).

Previously: `isToolOwned` recognized tool-owned files under the `command/`
prefix (singular).

#### Scenario: Tool-owned command file detection

- **GIVEN** a relative path `"commands/speckit.testreview.md"`
- **WHEN** `isToolOwned` is called with this path
- **THEN** it MUST return `true`

#### Scenario: User-owned command file detection

- **GIVEN** a relative path `"commands/gaze.md"`
- **WHEN** `isToolOwned` is called with this path
- **THEN** it MUST return `false`

### Requirement: Internal cross-references

Scaffolded command files that reference other command files MUST use
the `.opencode/commands/` path (plural).

Previously: Cross-references used `.opencode/command/` (singular).

#### Scenario: gaze-fix.md cross-references

- **GIVEN** the embedded `gaze-fix.md` asset
- **WHEN** it references the Speckit implement command
- **THEN** the path MUST be `.opencode/commands/speckit.implement.md`
- **AND** the OpenSpec apply reference MUST be
  `.opencode/commands/opsx-apply.md`

### Requirement: Embedded asset source-of-truth test

`TestEmbeddedAssetsMatchSource` MUST verify byte-identical content between
embedded assets and live `.opencode/` files using the `commands/` directory.

Previously: Test compared against `.opencode/command/` paths.

#### Scenario: Embedded-to-live file comparison

- **GIVEN** embedded assets under `internal/scaffold/assets/commands/`
- **WHEN** `TestEmbeddedAssetsMatchSource` runs
- **THEN** each embedded command file MUST match its counterpart at
  `.opencode/commands/<filename>`

## REMOVED Requirements

None — no requirements are removed.
