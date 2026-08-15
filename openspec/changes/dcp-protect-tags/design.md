## Context

OpenCode's DCP plugin compresses user-injected content during long sessions. Slash command files are user messages and eligible for compression, unlike skills (tool_use/tool_result pairs). The three scaffolded command files in `internal/scaffold/assets/commands/` contain multi-step execution pipelines that break when critical sections are compressed away.

DCP supports `<protect>...</protect>` tags that mark content as compression-exempt. This is an opt-in feature requiring users to enable `protectTags` in their DCP configuration.

The proposal (constitution alignment: all PASS/N/A) establishes that this change is additive, composable, and introduces no new runtime behavior.

## Goals / Non-Goals

### Goals
- Wrap execution-critical sections in `<protect>` tags across all three command files
- Prioritize protection by command complexity: `gaze-fix.md` > `speckit.testreview.md` > `gaze.md`
- Ensure protect tags do not alter command behavior when DCP is disabled
- Keep protect boundaries at logical section boundaries (not mid-paragraph)

### Non-Goals
- Modifying the DCP plugin itself
- Adding protect tags to agent prompts (`.opencode/agents/`) — agents are not injected as user messages
- Adding protect tags to skill files — skills are already protected by delivery mechanism
- Restructuring command file content for better compression — this is a tag-only change
- Protecting usage examples, option tables, or report format templates — these are reconstructible

## Decisions

### D1: Protect granularity — section-level, not line-level

Protect tags wrap entire logical sections (e.g., "### When no arguments are provided" through the end of its content) rather than individual lines or paragraphs. This keeps the Markdown readable and avoids tag proliferation.

**Rationale**: DCP compresses at the block level. Protecting individual lines within a compressible block provides no benefit — the entire block is either kept or compressed.

### D2: What to protect

Each command file has different criticality:

**`gaze-fix.md`** — 3 protect regions:
1. Workflow detection logic (lines ~41-82): Branch detection, Speckit/OpenSpec checks, fallback prompt — if compressed, the agent loses its routing logic
2. Batch remediation pipeline (lines ~84-153): Steps 1-4 with exact commands and filtering logic — if compressed, the agent skips steps or uses wrong parameters
3. Error handling (lines ~175-184): Abort conditions and graceful degradation — if compressed, failures go unreported

**`speckit.testreview.md`** — 2 protect regions:
1. Operating constraints (lines ~17-21): Read-only mandate and constitution authority — if compressed, the agent may modify files it shouldn't
2. Execution steps and operating principles (lines ~23-133): 6-step pipeline with delegation pattern, plus analysis guidelines ("NEVER modify files", "Missing coverage strategy is CRITICAL") — if compressed, analysis is incomplete, incorrectly formatted, or guardrails are lost

**`gaze.md`** — 1 protect region:
1. Instructions section (lines ~41-48): Core delegation instruction — if compressed, the agent doesn't know to delegate to gaze-reporter

### D3: Tags placed outside Markdown structure

Protect tags are placed on their own lines, outside of Markdown headings and code blocks. They do not wrap YAML frontmatter.

**Rationale**: YAML frontmatter is parsed by OpenCode's command loader, not rendered to the agent. Wrapping it in protect tags could interfere with parsing. The tags should be invisible to Markdown rendering.

### D4: No nesting of protect tags

Each protect region is a flat `<protect>...</protect>` pair. No nested protect tags.

**Rationale**: DCP does not define nesting semantics. Flat tags are simpler and sufficient.

## Risks / Trade-offs

### R1: Token overhead from protected content
Protected sections cannot be compressed, consuming more context window space during long sessions. Mitigated by protecting only execution-critical sections and leaving descriptions, examples, and templates unprotected.

### R2: Protect tags visible in non-DCP environments
When DCP is not enabled, the `<protect>` tags appear as literal text in the command content. OpenCode's Markdown renderer treats unknown HTML tags as passthrough, so they are invisible in rendered output but present in raw content.

### Coverage Strategy

- **Unit test**: `TestProtectTagPlacement` in `internal/scaffold/scaffold_test.go` — reads each command file from `embed.FS` and asserts: (a) correct number of `<protect>`/`</protect>` pairs per file (3, 2, 1), (b) balanced open/close tags, (c) no tags within YAML frontmatter, (d) no nested tags, (e) each tag on its own line
- **No integration or e2e tests needed**: Tags are static Markdown content with no runtime behavior
- **Existing tests**: `TestEmbeddedAssetsMatchSource` (drift detection) and `TestRun_CreatesFiles` (file delivery) provide additional regression protection

### R3: Scaffold delivery asymmetry
`gaze-fix.md` and `speckit.testreview.md` are tool-owned (overwritten on diff by `gaze init`), so existing projects receive the protect-tagged versions automatically. `gaze.md` is user-owned (skip-if-present), so existing projects retain their current version — only new projects receive the protect-tagged `gaze.md`. This is an accepted trade-off: `gaze.md` is the lowest-priority file (simple delegation, 1 protect region) and its protection is a minor improvement.
<!-- scaffolded by uf vdev -->
