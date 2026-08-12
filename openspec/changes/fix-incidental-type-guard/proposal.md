## Why

GitHub Issue #105: The classifier's naming signal (`internal/classify/naming.go`) and godoc signal (`internal/classify/godoc.go`) apply incidental penalties to **all** side effect types indiscriminately. When a function name contains I/O-describing prefixes like `log`, `debug`, `trace`, or `print`, or its godoc contains substrings like `logs`, `prints`, `traces`, or `debugs`, the penalty hits every effect — including P0 effects like `ReturnValue` and `ErrorReturn`.

This causes a cascade failure:

1. A function like `LogAndCompute` returning an `int` gets its `ReturnValue` effect penalized by naming (-10) and godoc (-15).
2. The tier boost (+25 for P0) and visibility (+8 for exported) are insufficient to overcome the penalties plus the contradiction penalty (-20) that fires when conflicting signals exist.
3. Final confidence: 38% → classified as `incidental` (below the 50 threshold).
4. `gaze quality` then flags the correct return-value assertion as **over-specification** — telling users to remove a test they should keep.

This is a **Principle I (Accuracy)** violation: a false classification that erodes trust in gaze's output.

## What Changes

Add effect-type guards to incidental naming prefixes and godoc keywords so they only apply to I/O-related side effect types, not to P0 contractual effects like return values.

## Capabilities

### New Capabilities

- None — this is a bug fix, not a new feature.

### Modified Capabilities

- `classify`: Incidental naming signals (`log`, `debug`, `trace`, `print` prefixes) now only penalize I/O effect types (`LogWrite`, `StdoutWrite`, `StderrWrite`), not return values or mutations.
- `classify`: Incidental godoc keywords (`logs`, `prints`, `traces`, `debugs` substrings) now only penalize I/O effect types, not return values or mutations.

### Removed Capabilities

- None.

## Impact

- **Files**: `internal/classify/naming.go`, `internal/classify/godoc.go`, and their test files.
- **Behavioral change**: Functions with I/O-describing names (e.g., `LogAndCompute`) will no longer have their return values misclassified as incidental. Return values will be classified based on their actual signals (tier boost, visibility, caller patterns) without interference from naming heuristics that describe the function's I/O behavior, not its return contract.
- **No regression risk to I/O classification**: The incidental signals still apply to `LogWrite`, `StdoutWrite`, and `StderrWrite` effects — the fix only prevents them from leaking onto unrelated effect types.
- **Downstream**: `gaze quality` will stop flagging correct return-value assertions as over-specification for functions with I/O-describing names.

## Constitution Alignment

Assessed against the Gaze project constitution.

### I. Accuracy

**Assessment**: PASS

This change directly fixes a false classification bug. P0 return values on functions with I/O-describing names are currently misclassified as incidental, which violates the constitutional mandate that false positives "erode trust and MUST be treated as bugs." The fix ensures naming/godoc signals only affect the effect types they semantically describe.

### II. Minimal Assumptions

**Assessment**: PASS

The fix reduces assumptions: currently the classifier assumes that if a function name starts with `log`, all of its effects are incidental. The fix narrows this to only assume that I/O effects (`LogWrite`, `StdoutWrite`, `StderrWrite`) are incidental when the name suggests I/O — a more precise assumption that matches reality.

### III. Actionable Output

**Assessment**: PASS

After the fix, `gaze quality` will stop producing misleading "over-specification" warnings for correct return-value assertions. This directly improves the actionability of gaze's output by eliminating a class of false recommendations.

### IV. Testability

**Assessment**: PASS

The fix uses the same typed-struct pattern already established by `contractualPrefixes` and `contractualKeywords`. Each incidental entry's `appliesTo` field is testable in isolation. Regression tests will verify that P0 effects are not penalized by I/O-describing names and that I/O effects still receive the correct penalty.
