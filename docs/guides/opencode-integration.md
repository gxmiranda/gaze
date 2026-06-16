# OpenCode Integration

Gaze integrates with [OpenCode](https://opencode.ai) to provide AI-assisted quality reporting directly from your editor. The `gaze init` command scaffolds agent definitions, command files, and reference materials into your project so that OpenCode can run Gaze analysis and produce human-readable reports.

## What `gaze init` Does

The [`gaze init`](../reference/cli/init.md) command writes OpenCode configuration files into the `.opencode/` directory of your project:

```bash
gaze init
```

Output:

```text
Gaze OpenCode integration initialized:
  created: .opencode/agents/gaze-reporter.md
  created: .opencode/agents/gaze-test-generator.md
  created: .opencode/agents/reviewer-testing.md
  created: .opencode/commands/gaze.md
  created: .opencode/commands/gaze-fix.md
  created: .opencode/commands/speckit.testreview.md
  created: .opencode/references/example-report.md
  created: .opencode/references/doc-scoring-model.md

Run /gaze for quality reports and /speckit.testreview for testability analysis.
```

Use `--force` to overwrite all existing files:

```bash
gaze init --force
```

## Scaffolded Files

The scaffold creates 8 files organized into three directories:

### Agents

| File | Description | Ownership |
|------|-------------|-----------|
| `agents/gaze-reporter.md` | Quality report agent — interprets Gaze JSON output and produces formatted markdown reports with actionable recommendations | User-owned |
| `agents/gaze-test-generator.md` | Test generation agent — generates Go test functions from Gaze quality data (gap hints, gaps, fix strategies) | Tool-owned |
| `agents/reviewer-testing.md` | Testing persona for the review council — audits test quality and testability | Tool-owned (via `gaze-fix.md` command) |

### Commands

| File | Description | Ownership |
|------|-------------|-----------|
| `commands/gaze.md` | `/gaze` command — runs Gaze CLI with `--format=json` and delegates to the gaze-reporter agent for formatting | User-owned |
| `commands/gaze-fix.md` | `/gaze fix` command — batch remediation using the gaze-test-generator agent | Tool-owned |
| `commands/speckit.testreview.md` | `/speckit.testreview` command — read-only spec testability analysis | Tool-owned |

### References

| File | Description | Ownership |
|------|-------------|-----------|
| `references/example-report.md` | Canonical example of the expected report output format | Tool-owned |
| `references/doc-scoring-model.md` | Document-enhanced classification scoring model for the gaze-reporter agent | Tool-owned |

## Tool-Owned vs User-Owned Files

Gaze distinguishes between two ownership models for scaffolded files:

**User-owned files** (`agents/gaze-reporter.md`, `commands/gaze.md`) are written once and never overwritten on subsequent `gaze init` runs. You can customize these files — add project-specific instructions, change formatting preferences, adjust the agent's behavior. Your changes are preserved.

**Tool-owned files** (everything under `references/`, plus `agents/gaze-test-generator.md`, `agents/reviewer-testing.md`, `commands/gaze-fix.md`, `commands/speckit.testreview.md`) use overwrite-on-diff behavior. When you run `gaze init`, these files are compared against the version embedded in the Gaze binary. If the content differs (e.g., because you upgraded Gaze), the file is silently updated. If the content is identical, the file is skipped. This ensures reference materials and tool-managed agents stay current without requiring `--force`.

The `--force` flag overrides both behaviors and overwrites all files unconditionally.

## Using the `/gaze` Command

After running `gaze init`, the `/gaze` command is available in OpenCode:

```text
/gaze ./...                     # Full report: CRAP + quality + classification
/gaze crap ./internal/store     # CRAP scores only
/gaze quality ./pkg/api         # Test quality metrics only
```

The `/gaze` command:

1. Runs the appropriate `gaze` CLI command with `--format=json`
2. Passes the JSON output to the gaze-reporter agent
3. The agent interprets the data and produces a formatted markdown report

The report includes:

- Module health summary with letter grades
- [CRAP](../reference/glossary.md) and [GazeCRAP](../reference/glossary.md) score analysis
- [Quadrant](../reference/glossary.md) distribution (Q1 Safe through Q4 Dangerous)
- Worst offenders with [fix strategy](../reference/glossary.md) labels
- [Contract coverage](../reference/glossary.md) gaps with specific unasserted effects
- Actionable recommendations prioritized by impact

## What the Gaze-Reporter Agent Does

The gaze-reporter agent is an OpenCode agent definition that instructs the AI model how to interpret Gaze's JSON output. It:

- Reads the JSON payload from Gaze CLI commands
- Applies the scoring consistency rules (CRAPload threshold from `summary.crap_threshold`, contract coverage as module-wide average)
- Uses emoji-rich formatting with section markers and severity indicators
- References the canonical example output from `references/example-report.md`
- Applies the document-enhanced classification scoring model from `references/doc-scoring-model.md`
- Produces a structured markdown report with clear remediation guidance

The agent does not perform any analysis — all metrics are computed by Gaze. The agent only formats and interprets the pre-computed data.

## What the Gaze-Test-Generator Agent Does

The gaze-test-generator agent generates Go test functions based on Gaze quality data. It reads:

- **GapHints**: Specific unasserted side effects for each function
- **Gaps**: Functions with zero contract coverage
- **FixStrategy**: The recommended remediation action (`add_tests`, `add_assertions`, `decompose`, `decompose_and_test`)

The agent supports five actions:

| Action | Description |
|--------|-------------|
| `add_tests` | Generate new test functions for functions with zero coverage |
| `add_assertions` | Add assertions to existing tests that execute code but don't verify behavior |
| `add_docs` | Add GoDoc comments to shift classification confidence |
| `decompose_and_test` | Split complex functions and generate tests for the extracted pieces |
| `decompose` | Reduce cyclomatic complexity by extracting helper functions |

Use the `/gaze fix` command to trigger batch remediation:

```text
/gaze fix ./internal/store
```

## Using `/speckit.testreview`

The `/speckit.testreview` command performs read-only testability analysis on spec artifacts. It examines `spec.md`, `plan.md`, and `tasks.md` for:

- Testability gaps in acceptance criteria
- Missing coverage strategy in the plan
- Tasks that lack test hooks or verification steps

This command does not modify any files — it produces a report with recommendations.

## Upgrading

When you upgrade Gaze (e.g., via `brew upgrade gaze`), run `gaze init` again to update tool-owned files:

```bash
gaze init
```

Tool-owned files with changed content are automatically updated. User-owned files are skipped (your customizations are preserved). The output shows what was updated:

```text
Gaze OpenCode integration already up to date:
  skipped: .opencode/agents/gaze-reporter.md (already exists)
  updated: .opencode/references/example-report.md (content changed)
  skipped: .opencode/commands/gaze.md (already exists)
```

## Version Marker

Every scaffolded file includes a version marker comment:

```html
<!-- scaffolded by gaze v1.2.3 -->
```

This marker is inserted after the YAML frontmatter (if present) or appended to the end of the file. It identifies which version of Gaze produced the file, which is useful for debugging when tool-owned files are not updating as expected.
