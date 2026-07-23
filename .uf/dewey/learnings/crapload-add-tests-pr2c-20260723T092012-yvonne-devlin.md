---
tag: crapload-add-tests-pr2c
author: yvonne-devlin
category: gotcha
created_at: 2026-07-23T09:20:12Z
identity: crapload-add-tests-pr2c-20260723T092012-yvonne-devlin
tier: draft
---

Removing testing.Short() guards from ResolvePackagePaths tests was the single highest-ROI change in the crapload-add-tests-pr2c work — CRAP dropped from 81.6 to 10.1 with zero production code changes and zero new test logic, just deleting 4 guard blocks (12 lines removed). The function already had comprehensive tests; they were simply invisible to CRAPload measurement because gaze crap runs go test -short internally. Before adding new tests or DI for CRAPload reduction, always check whether existing tests are gated by testing.Short() unnecessarily. Functions using packages.NeedName (the lightest go/packages load mode — no parsing, no type-checking) run in sub-second time and do not warrant the guard. The blanket policy of guarding all packages.Load callers was overly conservative for NeedName mode.
