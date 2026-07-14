---
tag: crapload-di-coverage
author: yvonne-devlin
category: gotcha
created_at: 2026-07-03T15:16:25Z
identity: crapload-di-coverage-20260703T151625-yvonne-devlin
tier: draft
---

During Phase 1b of CRAPload reduction (issue #166), we discovered that `isPointerArgStore` in `internal/analysis/mutation.go` has structurally unreachable branches. The function checks Store addresses against pointer parameters using 6 patterns: direct tracesToParam, UnOp dereference, FieldAddr direct, FieldAddr+UnOp, IndexAddr direct, and IndexAddr+UnOp. However, the `tracesToParam` function at line 345 already walks through FieldAddr, IndexAddr, and UnOp chains internally. This means the first check `tracesToParam(addr, param)` catches all real SSA patterns produced by the Go compiler — the subsequent UnOp/FieldAddr/IndexAddr branches in `isPointerArgStore` are defensive code that never executes with real Go source. As a result, line coverage caps at 50% regardless of how many fixture functions are added. To improve coverage beyond 50%, the function needs structural decomposition (Phase 2) — either removing the unreachable branches or restructuring to separate the value-chain walking from the type-checking. This is a common pattern in SSA analysis code where `tracesToParam` is a Swiss Army knife that resolves most address patterns.
