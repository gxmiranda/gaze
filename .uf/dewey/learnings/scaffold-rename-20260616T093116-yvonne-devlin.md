---
tag: scaffold-rename
author: yvonne-devlin
category: gotcha
created_at: 2026-06-16T09:31:16Z
identity: scaffold-rename-20260616T093116-yvonne-devlin
tier: draft
---

When renaming the gaze scaffold's embedded assets directory (e.g., command/ to commands/), the rename propagates automatically through embed.FS walk — no path-mapping code changes needed. However, the Cobra CLI Long description in cmd/gaze/main.go contains a hardcoded path string that won't be caught by grepping Go source for the embed prefix. Always search for user-facing CLI help text strings alongside production code paths when doing directory renames. The TestEmbeddedAssetsMatchSource test in scaffold_test.go is the strongest regression test for scaffold path changes — it enforces byte-identical content between embedded and live copies, so the live files must be moved in sync with the embedded directory rename.
