package adapter

import (
	"testing"

	"github.com/unbound-force/gaze/internal/taxonomy"
)

// ---------------------------------------------------------------------------
// confidenceRange tests
// ---------------------------------------------------------------------------

func TestConfidenceRange_AllNilClassification(t *testing.T) {
	effects := []taxonomy.SideEffect{
		{Classification: nil},
		{Classification: nil},
	}
	minC, maxC, found := confidenceRange(effects)
	if found {
		t.Errorf("found = true, want false")
	}
	if minC != 0 {
		t.Errorf("minConf = %d, want 0", minC)
	}
	if maxC != 0 {
		t.Errorf("maxConf = %d, want 0", maxC)
	}
}

func TestConfidenceRange_AllClassified(t *testing.T) {
	effects := []taxonomy.SideEffect{
		{Classification: &taxonomy.Classification{Confidence: 78}},
		{Classification: &taxonomy.Classification{Confidence: 85}},
		{Classification: &taxonomy.Classification{Confidence: 79}},
	}
	minC, maxC, found := confidenceRange(effects)
	if !found {
		t.Errorf("found = false, want true")
	}
	if minC != 78 {
		t.Errorf("minConf = %d, want 78", minC)
	}
	if maxC != 85 {
		t.Errorf("maxConf = %d, want 85", maxC)
	}
}

func TestConfidenceRange_MixedNilAndClassified(t *testing.T) {
	effects := []taxonomy.SideEffect{
		{Classification: nil},
		{Classification: &taxonomy.Classification{Confidence: 60}},
		{Classification: nil},
		{Classification: &taxonomy.Classification{Confidence: 72}},
	}
	minC, maxC, found := confidenceRange(effects)
	if !found {
		t.Errorf("found = false, want true")
	}
	if minC != 60 {
		t.Errorf("minConf = %d, want 60", minC)
	}
	if maxC != 72 {
		t.Errorf("maxConf = %d, want 72", maxC)
	}
}

func TestConfidenceRange_SingleEffect(t *testing.T) {
	effects := []taxonomy.SideEffect{
		{Classification: &taxonomy.Classification{Confidence: 50}},
	}
	minC, maxC, found := confidenceRange(effects)
	if !found {
		t.Errorf("found = false, want true")
	}
	if minC != 50 {
		t.Errorf("minConf = %d, want 50", minC)
	}
	if maxC != 50 {
		t.Errorf("maxConf = %d, want 50", maxC)
	}
}

func TestConfidenceRange_EmptySlice(t *testing.T) {
	minC, maxC, found := confidenceRange(nil)
	if found {
		t.Errorf("found = true, want false")
	}
	if minC != 0 {
		t.Errorf("minConf = %d, want 0", minC)
	}
	if maxC != 0 {
		t.Errorf("maxConf = %d, want 0", maxC)
	}
}

// ---------------------------------------------------------------------------
// deriveCoverageReason tests
// ---------------------------------------------------------------------------

func TestDeriveCoverageReason_NoEffects(t *testing.T) {
	reason, minC, maxC := deriveCoverageReason(nil, taxonomy.ContractCoverage{})
	if reason != "no_effects_detected" {
		t.Errorf("reason = %q, want %q", reason, "no_effects_detected")
	}
	if minC != 0 || maxC != 0 {
		t.Errorf("confidence = (%d, %d), want (0, 0)", minC, maxC)
	}
}

func TestDeriveCoverageReason_AllNilClassification(t *testing.T) {
	effects := []taxonomy.SideEffect{
		{Classification: nil},
		{Classification: nil},
	}
	cc := taxonomy.ContractCoverage{TotalContractual: 0}
	reason, minC, maxC := deriveCoverageReason(effects, cc)
	if reason != "no_effects_detected" {
		t.Errorf("reason = %q, want %q", reason, "no_effects_detected")
	}
	if minC != 0 || maxC != 0 {
		t.Errorf("confidence = (%d, %d), want (0, 0)", minC, maxC)
	}
}

func TestDeriveCoverageReason_AllAmbiguous(t *testing.T) {
	effects := []taxonomy.SideEffect{
		{Classification: &taxonomy.Classification{Confidence: 78}},
		{Classification: &taxonomy.Classification{Confidence: 79}},
	}
	cc := taxonomy.ContractCoverage{TotalContractual: 0}
	reason, minC, maxC := deriveCoverageReason(effects, cc)
	if reason != "all_effects_ambiguous" {
		t.Errorf("reason = %q, want %q", reason, "all_effects_ambiguous")
	}
	if minC != 78 {
		t.Errorf("minConf = %d, want 78", minC)
	}
	if maxC != 79 {
		t.Errorf("maxConf = %d, want 79", maxC)
	}
}

func TestDeriveCoverageReason_WithContractual(t *testing.T) {
	effects := []taxonomy.SideEffect{
		{Classification: &taxonomy.Classification{Confidence: 90}},
	}
	cc := taxonomy.ContractCoverage{TotalContractual: 1}
	reason, minC, maxC := deriveCoverageReason(effects, cc)
	if reason != "" {
		t.Errorf("reason = %q, want empty string", reason)
	}
	if minC != 0 || maxC != 0 {
		t.Errorf("confidence = (%d, %d), want (0, 0)", minC, maxC)
	}
}
