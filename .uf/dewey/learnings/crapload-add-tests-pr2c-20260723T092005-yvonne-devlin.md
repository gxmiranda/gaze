---
tag: crapload-add-tests-pr2c
author: yvonne-devlin
category: pattern
created_at: 2026-07-23T09:20:05Z
identity: crapload-add-tests-pr2c-20260723T092005-yvonne-devlin
tier: draft
---

When adding DI to a Go function that calls multiple internal helpers in sequence (like BuildContractCoverageFunc calling buildEffectsSet then buildCoverageMap), ensure ALL internal helpers that produce data consumed by later logic are injectable — not just the external dependencies. In the crapload-add-tests-pr2c change, the original design proposed a deps struct with 3 fields (resolvePackagePaths, loadConfig, buildEffectsSetFn), but during implementation a 4th field (buildCoverageMapFn) was needed because the closure-path tests needed to control both the effects set AND the coverage map. Without injecting buildCoverageMap, it was impossible to test the closure's three branches (found, no_test_coverage, no_effects_detected) with synthetic data. The lesson: when planning DI for orchestrator functions, trace the full data flow from inputs to the returned value and ensure every data-producing call is injectable, not just the ones that cross package boundaries.
