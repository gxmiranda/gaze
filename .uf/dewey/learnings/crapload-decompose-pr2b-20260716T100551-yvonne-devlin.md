---
tag: crapload-decompose-pr2b
author: yvonne-devlin
category: pattern
created_at: 2026-07-16T10:05:51Z
identity: crapload-decompose-pr2b-20260716T100551-yvonne-devlin
tier: draft
---

When decomposing complex AST-walking functions in Go (like matchContainerUnwrap at complexity 50), the key insight is identifying the phases: setup (collect initial state), process (iterate and transform), and match (check result). Each phase can be extracted as a separate function. However, the process phase often retains high complexity even after extraction because nested ast.Inspect closures, multi-iteration convergence loops, and transformation call detection involve inherent branching. traceForwardDataFlow landed at complexity 32 despite the design estimating ~12. The parent function (matchContainerUnwrap) still dropped from 50 to 8 because the complexity was moved, not eliminated. For future decomposition work on AST-heavy functions, expect the extracted "core loop" to retain ~60-70% of the original complexity — the wins come from isolating concerns, not from reducing total branch count.
