---
tag: test-duplication-avoidance
author: jay-flowers
category: gotcha
created_at: 2026-08-17T19:06:46Z
identity: test-duplication-avoidance-20260817T190646-jay-flowers
tier: draft
---

When updating existing test functions to cover new behavior (e.g., adding `bytes.Buffer` capture to existing `TestLoadFromDir_*` tests), do NOT also create separate new test functions that test the same scenario. The spec's "required tests" list should be mapped to existing tests first — if an existing test already covers the scenario, update it rather than creating a duplicate. In the config-load-warning implementation, 3 duplicate test functions had to be removed during code review because the existing tests were updated to include the new assertions, making the new tests redundant.
