---
tag: crapload-di-coverage
author: yvonne-devlin
category: gotcha
created_at: 2026-07-02T08:15:17Z
identity: crapload-di-coverage-20260702T081517-yvonne-devlin
tier: draft
---

When adding dependency injection to Go functions that already have a variadic parameter (like `aiMapperFn ...quality.AIMapperFunc`), the existing variadic must be absorbed into the new deps struct because Go does not allow two variadic parameters. This was discovered during spec review of the crapload-di-coverage-pr1a change where `analyzePackageCoverage` already had a variadic `aiMapperFn`. The solution was to add `aiMapperFn` as a regular field on the `contractCoverageDeps` struct and update the caller (`BuildContractCoverageFunc`) to pass it through the struct. Five spec reviewers independently flagged the `pipelineStepFuncs` signature cascade as an ambiguity — when step functions gain variadic deps, the function types stored in `pipelineStepFuncs` must also change, and test fake closures must be updated to match.
