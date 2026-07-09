---
tag: crapload-decompose-pr2a
author: yvonne-devlin
category: pattern
created_at: 2026-07-08T16:19:17Z
identity: crapload-decompose-pr2a-20260708T161917-yvonne-devlin
tier: draft
---

When deduplicating functions across packages (loadTestPackage in goprovider vs loadTestPackageForQuality in aireport), always compare the implementations line by line before choosing which to keep. In this case the goprovider version had pkg.Errors validation that the aireport version lacked — a latent bug. The aireport version would silently return packages with compilation errors, potentially producing incorrect quality data. By consolidating to the more robust copy and exporting it (goprovider.LoadTestPackage), we fixed a bug as a side effect of dedup. Similarly, the two copies of loadGazeConfigBestEffort were consolidated into config.LoadFromDir — the natural package for config-loading logic. When deciding where deduped functions should live, prefer the package that already has the supporting imports (goprovider already imported quality for HasTestSyntax) and avoid creating circular dependencies (loader couldn't import quality).
