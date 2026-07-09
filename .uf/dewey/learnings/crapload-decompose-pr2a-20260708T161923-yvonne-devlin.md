---
tag: crapload-decompose-pr2a
author: yvonne-devlin
category: gotcha
created_at: 2026-07-08T16:19:23Z
identity: crapload-decompose-pr2a-20260708T161923-yvonne-devlin
tier: draft
---

The spec review council's most consistently raised finding across 5 reviewers for the crapload-decompose-pr2a change was the ambiguity in task 3.2's test cleanup instruction: "Update TestLoadGazeConfigBestEffort to reference config.LoadFromDir (or delete if the test only tested the local copy)." Four of five reviewers flagged the "or delete" phrasing as ambiguous. The fix was straightforward — change it to a definitive "Delete TestLoadGazeConfigBestEffort_AlwaysNonNil" since the function being tested is being removed. Lesson: task specs should be definitive, not conditional, especially for deletion/cleanup steps. When a function is being deleted, tests for it should always be deleted too (unless they test behavior preserved by the replacement), and the task should say so explicitly.
