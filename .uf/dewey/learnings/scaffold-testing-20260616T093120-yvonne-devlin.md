---
tag: scaffold-testing
author: yvonne-devlin
category: pattern
created_at: 2026-06-16T09:31:20Z
identity: scaffold-testing-20260616T093120-yvonne-devlin
tier: draft
---

The gaze scaffold test file (internal/scaffold/scaffold_test.go) had multiple stale comments from previous changes: TestRun_CreatesFiles said "6 files" when checking 8, TestAssetPaths_Returns8Files GoDoc said "6 files", and TestRun_OverwriteOnDiff_SkipsIdentical said "7 files" when checking 8. These were all pre-existing from when the scaffold expanded from 2 to 4 to 6 to 8 files across specs 005, 012, 016, 017, and the gaze-test-generation change. The code review council (Architect and Testing reviewers) caught these and recommended fixing them alongside the rename. When expanding file counts in scaffold-based systems, always update test comments in the same commit.
