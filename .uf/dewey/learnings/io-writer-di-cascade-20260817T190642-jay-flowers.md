---
tag: io-writer-di-cascade
author: jay-flowers
category: pattern
created_at: 2026-08-17T19:06:42Z
identity: io-writer-di-cascade-20260817T190642-jay-flowers
tier: draft
---

When adding an `io.Writer` parameter to a function that is used as a DI function type (e.g., `qualityPipelineDeps.loadConfig`), the change cascades to: (1) the struct field type, (2) all default assignments, (3) all invocation sites, (4) all test fakes, and (5) any higher-level function type wrappers like `pipelineStepFuncs`. In this project, `config.LoadFromDir` gaining `stderr io.Writer` required updating 10 call sites across 4 files, 2 DI struct types, 2 pipelineStepFuncs types, and ~12 test fakes. The Go compiler catches all mismatches, making it safe but tedious.
