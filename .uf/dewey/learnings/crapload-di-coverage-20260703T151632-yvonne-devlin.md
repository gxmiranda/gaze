---
tag: crapload-di-coverage
author: yvonne-devlin
category: pattern
created_at: 2026-07-03T15:16:32Z
identity: crapload-di-coverage-20260703T151632-yvonne-devlin
tier: draft
---

When testing Bubble Tea model Update methods, the key pattern is to treat Update as a pure function: construct a model, send a message, assert on the returned model's observable state. For tea.Cmd assertions, the correct approach is to execute the command (`cmd()`) and check the returned message type (e.g., `tea.QuitMsg`), since Go functions are not comparable. The `analyzeModel.Update` tests achieved 100% line coverage with 5 subtests covering WindowSizeMsg init/resize, KeyMsg quit/help, and unhandled message passthrough. The errcheck linter requires type assertions from `result.(analyzeModel)` to use the two-value form `result.(analyzeModel)` → `model, ok := result.(analyzeModel)` with an explicit `if !ok` check. The `TestMain` function in the test file already sets `lipgloss.DefaultRenderer().SetColorProfile(termenv.Ascii)` for deterministic rendering across environments.
