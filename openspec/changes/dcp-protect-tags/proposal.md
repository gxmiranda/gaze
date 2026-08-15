## Why

OpenCode's DCP (Dynamic Context Pruning) plugin compresses user-injected content during long sessions to manage context windows. Slash command files are delivered as user messages and are eligible for compression, unlike skills which are delivered as tool_use/tool_result pairs and implicitly protected.

When DCP compresses slash command content, critical execution instructions — guardrails, multi-step pipelines, exit conditions, and session-resume guards — can be lost or degraded. This causes agent drift, skipped steps, and incorrect behavior mid-session.

DCP supports `<protect>...</protect>` tags that mark sections as compression-exempt. Adding these tags to the three scaffolded slash command files ensures critical sections survive context pruning without requiring changes to the DCP plugin itself.

GitHub Issue: #221

## What Changes

Add `<protect>` tags to the three slash command files in `internal/scaffold/assets/commands/` around sections that contain execution-critical content.

### Files affected:

1. **`gaze-fix.md`** (highest priority — complex multi-step pipeline with workflow detection, 5-step batch remediation, and error handling)
2. **`speckit.testreview.md`** (6-step execution pipeline with constitution authority, read-only constraints, and delegation pattern)
3. **`gaze.md`** (simple delegation command — minimal protection needed)

### What gets protected:

- **Guardrails**: Operating constraints, read-only mandates, constitution authority declarations
- **Execution checklists**: Numbered step sequences that must be followed in order
- **Exit conditions**: Error handling, abort conditions, fallback behavior
- **Session-resume guards**: Branch validation, prerequisite checks, workflow detection logic

### What does NOT get protected:

- Usage examples and option tables (reconstructible from context)
- Descriptions and background (non-critical for execution)
- Report format templates (cosmetic, not behavioral)

## Capabilities

### New Capabilities
- `DCP compression resilience`: Slash command files retain their critical execution instructions during long sessions when DCP compresses context

### Modified Capabilities
- `gaze-fix.md`: Workflow detection logic, batch remediation pipeline steps, and error handling wrapped in protect tags
- `speckit.testreview.md`: Execution steps, operating constraints, and delegation instructions wrapped in protect tags
- `gaze.md`: Core delegation instruction wrapped in protect tag

### Removed Capabilities
- None

## Impact

- **3 files modified**: `internal/scaffold/assets/commands/gaze-fix.md`, `gaze.md`, `speckit.testreview.md`
- **No Go code changes**: These are embedded Markdown assets; the scaffold Go code is unchanged
- **No behavioral change**: The protect tags are invisible to OpenCode's command execution — they are metadata consumed only by DCP
- **User opt-in required**: Users must enable `protectTags` in their DCP plugin configuration for the tags to take effect
- **Scaffold delivery**: Changes propagate to new projects via `gaze init`. For existing projects, `gaze-fix.md` and `speckit.testreview.md` are tool-owned and overwritten on diff automatically. `gaze.md` is user-owned (skip-if-present) — existing projects retain their current version; only new projects receive the protect-tagged `gaze.md`

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent prompt content (slash command files), not inter-hero communication artifacts or runtime coupling. No artifact-based collaboration patterns are affected.

### II. Composability First

**Assessment**: PASS

The protect tags are additive metadata that DCP consumes optionally. Commands continue to function identically without DCP enabled. No new dependencies are introduced — the tags are plain XML-like markers in Markdown.

### III. Observable Quality

**Assessment**: N/A

This change does not affect machine-parseable output, provenance metadata, or quality metrics. The slash commands produce the same analysis results regardless of protect tag presence.

### IV. Testability

**Assessment**: PASS

The protect tags are static Markdown content embedded via `embed.FS`. Existing scaffold tests that verify file presence and content remain valid. A new `TestProtectTagPlacement` test verifies protect tag count, balanced pairs, frontmatter exclusion, no nesting, and own-line placement for each command file — providing automated regression protection for the tag structure.
<!-- scaffolded by uf vdev -->
