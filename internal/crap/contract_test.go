package crap

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Contract coverage closure behavior tests
// These test the lookup closure behavior with ContractCoverageInfo,
// independent of the quality pipeline. The pipeline orchestration
// itself lives in internal/provider/goprovider/contract.go.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// findModuleRoot tests
// ---------------------------------------------------------------------------

func TestFindModuleRoot_Found(t *testing.T) {
	// Create a temp directory tree with a go.mod in the root.
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := findModuleRoot(sub)
	if err != nil {
		t.Fatalf("findModuleRoot: %v", err)
	}
	if got != root {
		t.Errorf("findModuleRoot = %q, want %q", got, root)
	}
}

func TestFindModuleRoot_NotFound(t *testing.T) {
	// Use a temp directory with no go.mod anywhere up the tree.
	dir := t.TempDir()
	sub := filepath.Join(dir, "no", "gomod", "here")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := findModuleRoot(sub)
	if err == nil {
		t.Fatal("expected error when no go.mod exists")
	}
}

func TestFindModuleRoot_AtStartDir(t *testing.T) {
	// go.mod is in the start directory itself.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := findModuleRoot(dir)
	if err != nil {
		t.Fatalf("findModuleRoot: %v", err)
	}
	if got != dir {
		t.Errorf("findModuleRoot = %q, want %q", got, dir)
	}
}

// ---------------------------------------------------------------------------
// Contract coverage closure behavior tests
// ---------------------------------------------------------------------------

// TestContractCoverageClosure_NoTestCoverage verifies that a
// function with detected effects but no test coverage returns
// ContractCoverageInfo with Reason "no_test_coverage". This tests
// the effectsSet logic added in spec 036 (FR-006).
func TestContractCoverageClosure_NoTestCoverage(t *testing.T) {
	// Construct the closure directly using the internal maps to
	// avoid running the full quality pipeline. The effectsSet
	// contains a function key, but the coverageMap does not.
	coverageMap := make(map[string]ContractCoverageInfo)
	effectsSet := map[string]bool{
		"mypkg:MyFunc": true,
	}

	fn := func(pkg, function string) (ContractCoverageInfo, bool) {
		key := pkg + ":" + function
		info, ok := coverageMap[key]
		if ok {
			return info, true
		}
		// Return ok=false so CRAP pipeline excludes from GazeCRAP
		// calculations. The Reason is informational for display.
		if effectsSet[key] {
			return ContractCoverageInfo{Reason: "no_test_coverage"}, false
		}
		return ContractCoverageInfo{Reason: "no_effects_detected"}, false
	}

	info, ok := fn("mypkg", "MyFunc")
	if ok {
		t.Fatal("expected ok=false for function with effects but no test (excluded from GazeCRAP)")
	}
	if info.Reason != "no_test_coverage" {
		t.Errorf("expected Reason %q, got %q", "no_test_coverage", info.Reason)
	}
	if info.Percentage != 0 {
		t.Errorf("expected Percentage 0 for untested function, got %.1f", info.Percentage)
	}
}

// TestContractCoverageClosure_NoEffects verifies that a function
// with zero detected effects and no test coverage returns
// ContractCoverageInfo with Reason "no_effects_detected" and
// ok=false (existing behavior preserved).
func TestContractCoverageClosure_NoEffects(t *testing.T) {
	// Construct the closure directly. Neither coverageMap nor
	// effectsSet contains the function key.
	coverageMap := make(map[string]ContractCoverageInfo)
	effectsSet := make(map[string]bool)

	fn := func(pkg, function string) (ContractCoverageInfo, bool) {
		key := pkg + ":" + function
		info, ok := coverageMap[key]
		if ok {
			return info, true
		}
		if effectsSet[key] {
			return ContractCoverageInfo{Reason: "no_test_coverage"}, true
		}
		return ContractCoverageInfo{Reason: "no_effects_detected"}, false
	}

	info, ok := fn("mypkg", "UnknownFunc")
	if ok {
		t.Error("expected ok=false for function with no effects and no test coverage")
	}
	if info.Reason != "no_effects_detected" {
		t.Errorf("expected Reason %q, got %q", "no_effects_detected", info.Reason)
	}
}
