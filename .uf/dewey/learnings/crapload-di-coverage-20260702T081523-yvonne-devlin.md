---
tag: crapload-di-coverage
author: yvonne-devlin
category: pattern
created_at: 2026-07-02T08:15:23Z
identity: crapload-di-coverage-20260702T081523-yvonne-devlin
tier: draft
---

The `gaze crap` CRAPload metric internally runs `go test -short -coverprofile=<tmpfile>` to generate coverage data. Tests guarded by `testing.Short()` are skipped during this coverage generation and contribute zero coverage to the CRAPload calculation. This means that adding tests with `-short` guards will NOT reduce CRAPload. To reduce CRAPload for functions that call heavy I/O operations (like `packages.Load`, `analysis.LoadAndAnalyze`), dependency injection is required so tests can use synthetic data and run without the `-short` guard. The existing `pipelineStepFuncs` pattern in `internal/aireport/runner.go` proves this approach works. After adding DI to 4 orchestration functions and direct fixture tests for 2 leaf functions, CRAPload dropped from 38 to 32 (6-function reduction, matching the target).
