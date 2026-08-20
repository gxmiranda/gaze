## Why

`confidenceRange` in `internal/adapter/contract.go` initializes min/max
sentinels as `(100, 0)` and skips effects with `nil` Classification. When
all effects have nil Classification (the normal state for external analyzers
with `classify_signals: false`), the inverted sentinels `(100, 0)` flow
through `deriveCoverageReason` into CRAP output as `"confidence 100-0"`.

This violates Principle I (Accuracy) — the output is nonsensical and
misleading. Reported in #181.

## What Changes

### `confidenceRange` returns `found bool`

A third `found bool` return value indicates whether any classified effects
were seen. When `!found`, the function returns `(0, 0, false)` instead of
the inverted sentinels `(100, 0)`.

### `deriveCoverageReason` maps `!found` to `"all_effects_unclassified"`

When effects exist but none carry classification data (the external analyzer
`classify_signals: false` case), the reason string is set to
`"all_effects_unclassified"` — a new value distinct from
`"no_effects_detected"` (which means "function has no side effects").

### Doc comment corrections

- Fixed incorrect path reference (`internal/crap/contract.go` →
  `internal/provider/goprovider/contract.go`)
- Corrected parity claim: the adapter iterates all effects, not the
  pre-filtered `AmbiguousEffects` from the quality pipeline

### Table-driven tests

Consolidated 5 structurally identical `confidenceRange` tests and 4
`deriveCoverageReason` tests into two table-driven test functions. Added
a missing test case for `TotalContractual > 0` with all-nil Classification.

## Capabilities

### New Capabilities
- New `"all_effects_unclassified"` reason string in `ContractCoverageInfo`

### Modified Capabilities
- `confidenceRange` signature: `(int, int)` → `(int, int, bool)`
- `deriveCoverageReason` now returns `"all_effects_unclassified"` instead
  of `"no_effects_detected"` when effects exist but are unclassified

### Removed Capabilities
- None

## Impact

- **Modified files**: `internal/adapter/contract.go` (logic fix + doc
  corrections), `internal/adapter/contract_internal_test.go` (table-driven
  rewrite + new test case), `internal/crap/analyze.go` (Reason doc update)
- **Risk**: Low. The change only affects the external analyzer code path.
  The Go-native path is unmodified.
- **Test impact**: 10 test cases across 2 table-driven functions (was 9
  across 9 individual functions). New test case covers `TotalContractual > 0`
  with all-nil Classification.

## Constitution Alignment

Assessed against the Gaze project constitution (`.specify/memory/constitution.md`).

### I. Accuracy

**Assessment**: PASS

Eliminates a false output value (`"confidence 100-0"`) and a semantically
incorrect reason string (`"no_effects_detected"` for functions that have
detected effects). The new `"all_effects_unclassified"` reason accurately
describes the state: effects exist but carry no classification data.

### II. Minimal Assumptions

**Assessment**: N/A

No new assumptions about host project structure. The fix applies to the
existing external analyzer protocol path.

### III. Actionable Output

**Assessment**: PASS

Replaces a confusing output value with a clear, machine-readable reason
string that correctly guides users. The `"all_effects_unclassified"` label
tells the user their external analyzer needs to provide classification data
for GazeCRAP to be meaningful.

### IV. Testability

**Assessment**: PASS

Both modified functions (`confidenceRange`, `deriveCoverageReason`) are
directly testable with synthetic inputs. Table-driven tests cover all
branches. The `found bool` return is more testable than the previous
implicit sentinel approach.
