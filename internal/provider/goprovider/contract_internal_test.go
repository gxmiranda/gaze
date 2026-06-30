package goprovider

import (
	"bytes"
	"testing"

	"github.com/unbound-force/gaze/internal/config"
)

// ---------------------------------------------------------------------------
// extractShortPkgName tests
// ---------------------------------------------------------------------------

func TestExtractShortPkgName_WithSlash(t *testing.T) {
	got := extractShortPkgName("github.com/unbound-force/gaze/internal/crap")
	if got != "crap" {
		t.Errorf("extractShortPkgName(...crap) = %q, want %q", got, "crap")
	}
}

func TestExtractShortPkgName_NoSlash(t *testing.T) {
	got := extractShortPkgName("main")
	if got != "main" {
		t.Errorf("extractShortPkgName(main) = %q, want %q", got, "main")
	}
}

func TestExtractShortPkgName_TrailingSlash(t *testing.T) {
	got := extractShortPkgName("github.com/user/repo/")
	if got != "" {
		t.Errorf("extractShortPkgName(.../repo/) = %q, want %q (empty)", got, "")
	}
}

func TestExtractShortPkgName_Empty(t *testing.T) {
	got := extractShortPkgName("")
	if got != "" {
		t.Errorf("extractShortPkgName('') = %q, want %q", got, "")
	}
}

// ---------------------------------------------------------------------------
// analyzePackageCoverage tests
// ---------------------------------------------------------------------------

func TestAnalyzePackageCoverage_ValidPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: loads real packages via analysis pipeline")
	}
	gazeConfig := config.DefaultConfig()
	var stderr bytes.Buffer
	reports, _ := analyzePackageCoverage(
		"github.com/unbound-force/gaze/internal/quality/testdata/src/welltested",
		".",
		gazeConfig,
		&stderr,
	)
	if len(reports) == 0 {
		t.Error("expected non-nil quality reports for well-tested package")
	}
}

func TestAnalyzePackageCoverage_InvalidPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: invokes go/packages.Load via analysis.LoadAndAnalyze")
	}
	gazeConfig := config.DefaultConfig()
	var stderr bytes.Buffer
	reports, _ := analyzePackageCoverage(
		"github.com/nonexistent/does/not/exist",
		".",
		gazeConfig,
		&stderr,
	)
	if reports != nil {
		t.Error("expected nil reports for non-existent package")
	}
}

// ---------------------------------------------------------------------------
// BuildContractCoverageFunc tests
// ---------------------------------------------------------------------------

func TestBuildContractCoverageFunc_InvalidPattern(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: invokes go/packages.Load via resolvePackagePaths")
	}
	var buf bytes.Buffer
	fn, _ := BuildContractCoverageFunc(
		[]string{"github.com/nonexistent/package/does/not/exist"},
		t.TempDir(),
		&buf,
	)
	if fn != nil {
		_, ok := fn("nonexistent", "Foo")
		if ok {
			t.Error("expected ok=false for unknown pkg:func key")
		}
	}
}

func TestBuildContractCoverageFunc_WelltestedPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: runs quality pipeline (package loading)")
	}

	pattern := "github.com/unbound-force/gaze/internal/quality/testdata/src/welltested"

	var buf bytes.Buffer
	fn, _ := BuildContractCoverageFunc([]string{pattern}, ".", &buf)

	if fn == nil {
		t.Fatal("BuildContractCoverageFunc returned nil; expected non-nil closure for well-tested package")
	}

	info, ok := fn("welltested", "Add")
	t.Logf("welltested:Add contract coverage: %.1f%% (found=%v, reason=%q)", info.Percentage, ok, info.Reason)
	if !ok {
		t.Fatal("expected ok=true for welltested:Add, got ok=false")
	}
	if info.Percentage <= 0 {
		t.Errorf("expected pct > 0 for welltested:Add (well-tested fixture should have non-zero coverage), got %.1f", info.Percentage)
	}
}
