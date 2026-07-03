package adapter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/unbound-force/gaze/internal/protocol"
	"github.com/unbound-force/gaze/internal/taxonomy"
)

// TestConvertAnalysisResults_UniversalType verifies that a universal
// taxonomy type (GeneratorYield) is converted without a warning and
// receives the correct tier assignment (P1).
func TestConvertAnalysisResults_UniversalType(t *testing.T) {
	funcs := []protocol.AnalyzedFunction{
		{
			Name:    "generate_values",
			Package: "generators",
			File:    "generators/gen.py",
			Line:    5,
			SideEffects: []protocol.AnalyzedSideEffect{
				{
					Type:        "GeneratorYield",
					Description: "yields computed value",
					Location:    "generators/gen.py:10:5",
					Target:      "value",
				},
			},
		},
	}

	var stderr bytes.Buffer
	results := convertAnalysisResults(funcs, &stderr)

	// No warning should be logged for a known universal type.
	if stderr.Len() != 0 {
		t.Errorf("unexpected warning for known universal type: %s", stderr.String())
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if len(results[0].SideEffects) != 1 {
		t.Fatalf("got %d effects, want 1", len(results[0].SideEffects))
	}

	effect := results[0].SideEffects[0]
	if effect.Type != taxonomy.GeneratorYield {
		t.Errorf("effect type = %q, want %q", effect.Type, taxonomy.GeneratorYield)
	}
	if effect.Tier != taxonomy.TierP1 {
		t.Errorf("effect tier = %q, want %q", effect.Tier, taxonomy.TierP1)
	}
}

// TestConvertAnalysisResults_UnknownTypeWarning verifies that a
// truly unknown type produces a warning (regression test for the
// isKnownP4 → IsKnownType migration).
func TestConvertAnalysisResults_UnknownTypeWarning(t *testing.T) {
	funcs := []protocol.AnalyzedFunction{
		{
			Name:    "do_something",
			Package: "pkg",
			File:    "pkg/mod.py",
			Line:    1,
			SideEffects: []protocol.AnalyzedSideEffect{
				{
					Type:        "CompletelyMadeUpEffect",
					Description: "not a real type",
				},
			},
		},
	}

	var stderr bytes.Buffer
	results := convertAnalysisResults(funcs, &stderr)

	// A warning should be logged for the unknown type.
	if !strings.Contains(stderr.String(), "CompletelyMadeUpEffect") {
		t.Errorf("expected warning about unknown type, got: %q", stderr.String())
	}

	// The effect should still be included with P4 default tier.
	if len(results) != 1 || len(results[0].SideEffects) != 1 {
		t.Fatalf("expected 1 result with 1 effect, got %d results", len(results))
	}
	if results[0].SideEffects[0].Tier != taxonomy.TierP4 {
		t.Errorf("unknown type tier = %q, want %q", results[0].SideEffects[0].Tier, taxonomy.TierP4)
	}
}

// TestConvertAnalysisResults_DetailPassthrough verifies that the
// Detail metadata map from the protocol response is preserved
// through conversion to taxonomy.SideEffect.
func TestConvertAnalysisResults_DetailPassthrough(t *testing.T) {
	detail := map[string]any{
		"language_type": "RaiseException",
		"confidence":    0.95,
		"framework":     "django",
	}

	funcs := []protocol.AnalyzedFunction{
		{
			Name:    "handle_request",
			Package: "views",
			File:    "views/api.py",
			Line:    42,
			SideEffects: []protocol.AnalyzedSideEffect{
				{
					Type:        "ErrorSignal",
					Description: "raises HTTP 404",
					Location:    "views/api.py:50:9",
					Target:      "Http404",
					Detail:      detail,
				},
			},
		},
	}

	var stderr bytes.Buffer
	results := convertAnalysisResults(funcs, &stderr)

	if stderr.Len() != 0 {
		t.Errorf("unexpected warning: %s", stderr.String())
	}

	if len(results) != 1 || len(results[0].SideEffects) != 1 {
		t.Fatalf("expected 1 result with 1 effect")
	}

	effect := results[0].SideEffects[0]

	// Verify Detail is non-nil and contains expected keys.
	if effect.Detail == nil {
		t.Fatal("effect.Detail is nil, want non-nil")
	}

	// Verify string value.
	langType, ok := effect.Detail["language_type"]
	if !ok {
		t.Fatal("Detail missing key \"language_type\"")
	}
	if langType != "RaiseException" {
		t.Errorf("Detail[\"language_type\"] = %v, want %q", langType, "RaiseException")
	}

	// Verify numeric value (float64 in Go's map[string]any).
	conf, ok := effect.Detail["confidence"]
	if !ok {
		t.Fatal("Detail missing key \"confidence\"")
	}
	confFloat, ok := conf.(float64)
	if !ok {
		t.Fatalf("Detail[\"confidence\"] type = %T, want float64", conf)
	}
	if confFloat != 0.95 {
		t.Errorf("Detail[\"confidence\"] = %g, want 0.95", confFloat)
	}

	// Verify third key.
	fw, ok := effect.Detail["framework"]
	if !ok {
		t.Fatal("Detail missing key \"framework\"")
	}
	if fw != "django" {
		t.Errorf("Detail[\"framework\"] = %v, want %q", fw, "django")
	}
}

// TestConvertAnalysisResults_NilDetail verifies that a nil Detail
// map is preserved as nil (not converted to an empty map).
func TestConvertAnalysisResults_NilDetail(t *testing.T) {
	funcs := []protocol.AnalyzedFunction{
		{
			Name:    "simple_func",
			Package: "pkg",
			File:    "pkg/mod.py",
			Line:    1,
			SideEffects: []protocol.AnalyzedSideEffect{
				{
					Type:        "ReturnValue",
					Description: "returns result",
					// Detail intentionally omitted (nil).
				},
			},
		},
	}

	results := convertAnalysisResults(funcs, nil)

	if len(results) != 1 || len(results[0].SideEffects) != 1 {
		t.Fatalf("expected 1 result with 1 effect")
	}
	if results[0].SideEffects[0].Detail != nil {
		t.Errorf("Detail = %v, want nil", results[0].SideEffects[0].Detail)
	}
}
