# gaze init

Scaffold OpenCode agent and command files into the current directory. After running this command, you can use the `/gaze` command in OpenCode to generate AI-powered quality reports.

## Synopsis

```
gaze init [flags]
```

## Arguments

This command takes no positional arguments. It always operates on the current working directory.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | `bool` | `false` | Overwrite existing files. Without this flag, existing user-owned files are skipped. |

## Scaffolded Files

The `init` command creates the following directory structure:

```
.opencode/
├── agents/
│   └── gaze-reporter.md          # AI agent prompt for quality reporting
├── commands/
│   ├── gaze.md                   # /gaze command definition
│   ├── review-council.md         # /review-council command definition
│   └── speckit.testreview.md     # /speckit.testreview command definition
└── references/
    ├── gaze-example-output.md    # Canonical example output for the reporter
    └── gaze-scoring-model.md     # Document-enhanced classification scoring model
```

### File Ownership Model

Gaze uses a mixed ownership model for scaffolded files:

- **User-owned files** (`agents/`, `commands/gaze.md`): Created once, never overwritten on subsequent `gaze init` runs (unless `--force` is used). Users can customize these files.
- **Tool-owned files** (`references/`, `commands/review-council.md`, `commands/speckit.testreview.md`): Updated automatically on `gaze init` if the content has changed. These files are maintained by Gaze and should not be manually edited.

## Configuration Interaction

This command does not read `.gaze.yaml`.

## Examples

### Initialize OpenCode integration

```bash
cd /path/to/your/go/project
gaze init
```

```
✓ Created .opencode/agents/gaze-reporter.md
✓ Created .opencode/commands/gaze.md
✓ Created .opencode/commands/review-council.md
✓ Created .opencode/commands/speckit.testreview.md
✓ Created .opencode/references/gaze-example-output.md
✓ Created .opencode/references/gaze-scoring-model.md
```

### Re-initialize (update tool-owned files)

```bash
gaze init
```

On subsequent runs, user-owned files are skipped and tool-owned files are updated if their content has changed.

### Force overwrite all files

```bash
gaze init --force
```

Overwrites all files, including user-customized agent prompts and commands.

## See Also

- [OpenCode Integration](../../guides/opencode-integration.md) — how to use `/gaze` after initialization
- [`gaze report`](report.md) — the CLI equivalent of the `/gaze` OpenCode command
