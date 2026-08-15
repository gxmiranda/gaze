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

## 1. Add protect tags to command files

Each task modifies a different file and can run in parallel.

- [x] 1.1 [P] Add protect tags to `gaze-fix.md` — wrap 3 regions: (1) workflow detection logic ("When no arguments are provided" section, lines ~41-82), (2) batch remediation pipeline ("When arguments are provided" section, lines ~84-153), (3) error handling section (lines ~175-184). Place `<protect>` on its own line before each section's heading and `</protect>` on its own line after the section's last content line. Do NOT wrap YAML frontmatter, usage/options table, or report template. File: `internal/scaffold/assets/commands/gaze-fix.md`. [DCP-PROTECT-001, DCP-PROTECT-004]

- [x] 1.2 [P] Add protect tags to `speckit.testreview.md` — wrap 2 regions: (1) operating constraints ("Operating Constraints" section, lines ~17-21), (2) execution steps ("Execution Steps" through "Operating Principles", lines ~23-133). Place tags on own lines at section boundaries. Do NOT wrap YAML frontmatter or "User Input" section. File: `internal/scaffold/assets/commands/speckit.testreview.md`. [DCP-PROTECT-002, DCP-PROTECT-004]

- [x] 1.3 [P] Add protect tags to `gaze.md` — wrap 1 region: Instructions section (lines ~41-48). Place `<protect>` before "## Instructions" and `</protect>` after the last line. Do NOT wrap YAML frontmatter, usage examples, or mode table. File: `internal/scaffold/assets/commands/gaze.md`. [DCP-PROTECT-003, DCP-PROTECT-004]

## 2. Verification

- [x] 2.1 Verify scaffold builds — run `go build ./cmd/gaze` to confirm embedded assets compile. Run `go test -race -count=1 -short ./internal/scaffold/...` to confirm scaffold tests pass. [DCP-PROTECT-005]

- [x] 2.2 Add `TestProtectTagPlacement` to `internal/scaffold/scaffold_test.go` — read each command file from `embed.FS` and assert: (a) correct number of `<protect>`/`</protect>` pairs per file (3 for gaze-fix.md, 2 for speckit.testreview.md, 1 for gaze.md), (b) balanced open/close tags, (c) no tags within YAML frontmatter (between `---` delimiters), (d) no nested protect tags. Run `go test -race -count=1 -short -run TestProtectTagPlacement ./internal/scaffold/...` to verify. [DCP-PROTECT-004, DCP-PROTECT-006]

- [x] 2.3 Constitution alignment verification — confirm the change satisfies the constitution alignment assessment from the proposal: Composability (tags are additive, no new dependencies), Testability (automated test provides regression protection). [Constitution]
<!-- scaffolded by uf vdev -->
