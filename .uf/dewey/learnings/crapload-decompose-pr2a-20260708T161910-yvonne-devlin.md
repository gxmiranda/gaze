---
tag: crapload-decompose-pr2a
author: yvonne-devlin
category: pattern
created_at: 2026-07-08T16:19:10Z
identity: crapload-decompose-pr2a-20260708T161910-yvonne-devlin
tier: draft
---

When decomposing a high-complexity Go function (CRAP 506, complexity 22) like BuildContractCoverageFunc, the most effective strategy is identifying distinct concerns interleaved in the same loop: effects discovery, coverage map construction, and reason computation were three independent concerns sharing a single for-loop iteration. Extracting each as a separate function with its own DI parameter reduced complexity from 22 to 7 in one PR. The key enabling pattern was that each helper could accept a function parameter (loadAndAnalyzeFn) or the existing contractCoverageDeps struct, allowing tests to inject synthetic data without loading real Go packages. This avoids the testing.Short() guard problem — tests contribute to CRAPload because they don't need the -short guard since they use synthetic data instead of calling packages.Load.
