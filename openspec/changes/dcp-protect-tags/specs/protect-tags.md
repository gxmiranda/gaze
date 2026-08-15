## ADDED Requirements

### Requirement: DCP-PROTECT-001 — Protect tags on gaze-fix.md

The `gaze-fix.md` command file MUST wrap execution-critical sections in `<protect>...</protect>` tags to prevent DCP compression of:
1. Workflow detection logic (the "When no arguments are provided" section)
2. Batch remediation pipeline steps (Steps 1-4 in the "When arguments are provided" section)
3. Error handling section

The protect tags MUST NOT wrap:
- YAML frontmatter
- Usage examples or option tables
- Report format templates

#### Scenario: DCP compresses gaze-fix command during long session
- **GIVEN** a user has enabled `protectTags` in their DCP configuration
- **WHEN** DCP compresses the context window during a long session containing `/gaze fix`
- **THEN** the workflow detection logic, batch remediation pipeline, and error handling sections remain uncompressed
- **AND** usage examples and option tables MAY be compressed

#### Scenario: gaze-fix command without DCP enabled
- **GIVEN** a user has NOT enabled DCP or `protectTags`
- **WHEN** the `/gaze fix` command is executed
- **THEN** the command behaves identically to the pre-change version
- **AND** the `<protect>` tags are treated as passthrough HTML and do not affect Markdown rendering

### Requirement: DCP-PROTECT-002 — Protect tags on speckit.testreview.md

The `speckit.testreview.md` command file MUST wrap execution-critical sections in `<protect>...</protect>` tags to prevent DCP compression of:
1. Operating constraints (read-only mandate, constitution authority)
2. Execution steps (Steps 1-6, delegation pattern, report format, next actions)

The protect tags MUST NOT wrap:
- YAML frontmatter
- User input section

#### Scenario: DCP compresses testreview command during long session
- **GIVEN** a user has enabled `protectTags` in their DCP configuration
- **WHEN** DCP compresses the context window during a long session containing `/speckit.testreview`
- **THEN** the operating constraints and execution steps remain uncompressed
- **AND** the read-only mandate is preserved, preventing the agent from modifying files

#### Scenario: Constitution authority survives compression
- **GIVEN** a user has enabled `protectTags` in their DCP configuration
- **WHEN** DCP compresses the `/speckit.testreview` command content
- **THEN** the constitution authority declaration ("Constitution Principle IV violations are automatically CRITICAL") remains in the agent's context

### Requirement: DCP-PROTECT-003 — Protect tags on gaze.md

The `gaze.md` command file MUST wrap the Instructions section in a `<protect>...</protect>` tag to prevent DCP compression of the delegation instruction to the `gaze-reporter` agent.

#### Scenario: DCP compresses gaze command during long session
- **GIVEN** a user has enabled `protectTags` in their DCP configuration
- **WHEN** DCP compresses the context window during a long session containing `/gaze`
- **THEN** the delegation instruction to `gaze-reporter` agent remains uncompressed

### Requirement: DCP-PROTECT-004 — Tag placement rules

All `<protect>` tags MUST:
1. Be placed on their own line (not inline with other content)
2. Not wrap YAML frontmatter
3. Not be nested within other protect tags
4. Be placed at logical section boundaries (e.g., before a Markdown heading, after a section's last content)

#### Scenario: Protect tag does not interfere with YAML frontmatter
- **GIVEN** a command file with YAML frontmatter delimited by `---`
- **WHEN** the file is parsed by OpenCode's command loader
- **THEN** the YAML frontmatter is parsed correctly
- **AND** no `<protect>` tags appear within the frontmatter block

### Requirement: DCP-PROTECT-005 — Scaffold delivery

Updated command files with protect tags MUST be delivered to new projects via the `gaze init` scaffold mechanism. For existing projects, tool-owned command files (`gaze-fix.md`, `speckit.testreview.md`) SHALL be overwritten when the scaffold detects a content diff. The user-owned command file (`gaze.md`) follows skip-if-present behavior — existing projects retain their current version; only new projects receive the protect-tagged `gaze.md`.

#### Scenario: New project receives all protect-tagged files via gaze init
- **GIVEN** a new project with no existing `.opencode/` directory
- **WHEN** the user runs `gaze init`
- **THEN** all three command files are created with protect tags

#### Scenario: Existing project receives protect tags for tool-owned files
- **GIVEN** an existing project with older command files (without protect tags)
- **WHEN** the user runs `gaze init`
- **THEN** `gaze-fix.md` and `speckit.testreview.md` are overwritten with protect-tagged versions
- **AND** `gaze.md` is skipped (user-owned, already exists)

### Requirement: DCP-PROTECT-006 — Automated tag verification

The embedded command files MUST be verified by an automated test (`TestProtectTagPlacement`) that asserts:
1. Correct number of `<protect>`/`</protect>` pairs per file (3 for `gaze-fix.md`, 2 for `speckit.testreview.md`, 1 for `gaze.md`)
2. Balanced open/close tags (every `<protect>` has a matching `</protect>`)
3. No `<protect>` tags within YAML frontmatter blocks
4. No nested protect tags
5. Each `<protect>` and `</protect>` tag appears on its own line (not inline with other content)

#### Scenario: Embedded gaze-fix.md contains correct protect tags
- **GIVEN** the embedded asset `commands/gaze-fix.md` read from `embed.FS`
- **WHEN** its content is parsed for protect tags
- **THEN** it contains exactly 3 `<protect>` opening tags and 3 `</protect>` closing tags
- **AND** no protect tags appear within the YAML frontmatter block (between `---` delimiters)

#### Scenario: Embedded speckit.testreview.md contains correct protect tags
- **GIVEN** the embedded asset `commands/speckit.testreview.md` read from `embed.FS`
- **WHEN** its content is parsed for protect tags
- **THEN** it contains exactly 2 `<protect>` opening tags and 2 `</protect>` closing tags

#### Scenario: Embedded gaze.md contains correct protect tags
- **GIVEN** the embedded asset `commands/gaze.md` read from `embed.FS`
- **WHEN** its content is parsed for protect tags
- **THEN** it contains exactly 1 `<protect>` opening tag and 1 `</protect>` closing tag

## MODIFIED Requirements

None — this change adds new content (protect tags) to existing files without modifying any existing requirements.

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
