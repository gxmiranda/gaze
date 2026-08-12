<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Add effect-type guards to incidental signals

- [x] 1.1 [P] Convert `incidentalPrefixes` in `internal/classify/naming.go` from `[]string` to a typed slice with `appliesTo []taxonomy.SideEffectType`. Each entry (`log`, `Log`, `debug`, `Debug`, `trace`, `Trace`, `print`, `Print`) gets `appliesTo: []taxonomy.SideEffectType{taxonomy.LogWrite, taxonomy.StdoutWrite, taxonomy.StderrWrite}`. Update the loop in `AnalyzeNamingSignal` to check `effectType` against `appliesTo` before returning the incidental signal — skip the entry if the effect type doesn't match. Add a comment referencing issue #105.
- [x] 1.2 [P] Convert `incidentalKeywords` in `internal/classify/godoc.go` from `[]string` to a typed slice with `appliesTo []taxonomy.SideEffectType`. Each entry (`logs`, `prints`, `traces`, `debugs`) gets `appliesTo: []taxonomy.SideEffectType{taxonomy.LogWrite, taxonomy.StdoutWrite, taxonomy.StderrWrite}`. Update the loop in `AnalyzeGodocSignal` to check `effectType` against `appliesTo` before returning the incidental signal — skip the entry if the effect type doesn't match. Add a comment referencing issue #105.

## 2. Update existing tests and add regression tests

- [x] 2.1 [P] Update existing tests in `internal/classify/classify_test.go` and add new tests in `internal/classify/naming_test.go` for the type-guarded incidental naming behavior. **Existing test to update**: `TestNamingSignal_IncidentalPrefixes` (classify_test.go:113) currently passes `taxonomy.ReturnValue` and asserts `Weight < 0` — change the effect type to `taxonomy.LogWrite` (or split into two cases: `LogWrite` expects negative weight, `ReturnValue` expects zero). **New tests**: Cover all four spec scenarios: (a) `ReturnValue` on `LogAndCompute` → no penalty, (b) `LogWrite` on `LogAndCompute` → penalty applied, (c) `ErrorReturn` on `DebugAndFetch` → no penalty, (d) `StdoutWrite` on `PrintSummary` → penalty applied. Use table-driven tests with `t.Run`.
- [x] 2.2 [P] Update existing tests and add new tests in `internal/classify/godoc_test.go` for the type-guarded incidental godoc behavior. **Existing tests to update**: (1) `TestAnalyzeGodocSignal_IncidentalPriority` (godoc_test.go:214) passes `taxonomy.ReturnValue` with godoc containing both "logs" and "returns", asserting incidental wins at `-15` — after the fix, "logs" won't apply to `ReturnValue`, so update to use `taxonomy.LogWrite` or update expected behavior. (2) `TestAnalyzeGodocSignal_IncidentalOnly` (godoc_test.go:258) passes `taxonomy.ReturnValue` with "logs" keyword and asserts `Weight == -15` — change effect type to `taxonomy.LogWrite` or update expected weight to `0`. **New tests**: Cover all four spec scenarios: (a) `ReturnValue` with "logs" in godoc → no penalty, (b) `LogWrite` with "logs" in godoc → penalty applied, (c) `StderrWrite` with "prints" in godoc → penalty applied, (d) `ReceiverMutation` with "traces" in godoc → no penalty. Use table-driven tests with `t.Run`.

## 3. Verify end-to-end classification fix

- [x] 3.1 Write an automated integration test (e.g., `TestComputeScore_Issue105_LogAndComputeReturnValue`) in `internal/classify/classify_test.go` that verifies the full scoring path for the issue #105 scenario: construct signals for a function named `LogAndCompute` with `ReturnValue` effect, exported, with godoc containing only "logs" (no contractual keywords). Assert that the final classification is `contractual` with confidence ≥ 80 (expected: `75 (base + P0 boost) + 8 (visibility) = 83`). This locks down the exact bug scenario as a regression test through `ComputeScore` where the contradiction penalty previously amplified the misclassification. Run `go test -race -count=1 -short ./internal/classify/...` to confirm all tests pass.

## 4. Verification gates

- [x] 4.1 Run `go test -race -count=1 -short ./...` — all tests must pass.
- [x] 4.2 Run `golangci-lint run` — no lint errors.
- [x] 4.3 Verify constitution alignment: Principle I (Accuracy) — false classification bug is fixed. Principle IV (Testability) — regression tests verify observable behavior. Both PASS.

<!-- spec-review: passed -->
