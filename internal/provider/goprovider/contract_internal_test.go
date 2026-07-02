package goprovider

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/unbound-force/gaze/internal/analysis"
	"github.com/unbound-force/gaze/internal/config"
	"github.com/unbound-force/gaze/internal/quality"
	"github.com/unbound-force/gaze/internal/taxonomy"
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

// ---------------------------------------------------------------------------
// analyzePackageCoverage DI tests (Task 1.2)
// ---------------------------------------------------------------------------

// syntheticResult returns a minimal AnalysisResult for DI tests.
func syntheticResult() taxonomy.AnalysisResult {
	return taxonomy.AnalysisResult{
		Target: taxonomy.FunctionTarget{
			Package:  "example.com/pkg",
			Function: "DoWork",
		},
		SideEffects: []taxonomy.SideEffect{
			{
				ID:   "se-abc12345",
				Type: taxonomy.ReturnValue,
				Tier: taxonomy.TierP0,
			},
		},
	}
}

// successDeps returns a contractCoverageDeps that simulates a
// successful 4-step pipeline with synthetic data.
func successDeps() contractCoverageDeps {
	return contractCoverageDeps{
		loadAndAnalyze: func(_ string, _ analysis.Options) ([]taxonomy.AnalysisResult, error) {
			return []taxonomy.AnalysisResult{syntheticResult()}, nil
		},
		classifyResults: func(results []taxonomy.AnalysisResult, _ string, _ string, _ *config.GazeConfig) []taxonomy.AnalysisResult {
			return results
		},
		loadTestPkg: func(_ string) (*packages.Package, error) {
			return &packages.Package{Name: "test"}, nil
		},
		assess: func(_ []taxonomy.AnalysisResult, _ *packages.Package, _ quality.Options) ([]taxonomy.QualityReport, *taxonomy.PackageSummary, error) {
			return []taxonomy.QualityReport{
				{
					TestFunction: "TestDoWork",
					TargetFunction: taxonomy.FunctionTarget{
						Package:  "example.com/pkg",
						Function: "DoWork",
					},
					ContractCoverage: taxonomy.ContractCoverage{
						Percentage: 75.0,
					},
				},
			}, &taxonomy.PackageSummary{TotalTests: 1}, nil
		},
	}
}

func TestAnalyzePackageCoverage_DI_Success(t *testing.T) {
	var stderr bytes.Buffer
	reports, degradedPkg := analyzePackageCoverage(
		"example.com/pkg", ".", config.DefaultConfig(), &stderr,
		successDeps(),
	)
	if len(reports) == 0 {
		t.Fatal("expected non-empty reports for successful pipeline")
	}
	if reports[0].ContractCoverage.Percentage != 75.0 {
		t.Errorf("expected 75.0%% coverage, got %.1f%%", reports[0].ContractCoverage.Percentage)
	}
	if degradedPkg != "" {
		t.Errorf("expected empty degradedPkg, got %q", degradedPkg)
	}
}

func TestAnalyzePackageCoverage_DI_AnalysisError(t *testing.T) {
	deps := successDeps()
	deps.loadAndAnalyze = func(_ string, _ analysis.Options) ([]taxonomy.AnalysisResult, error) {
		return nil, errors.New("analysis failed")
	}
	var stderr bytes.Buffer
	reports, degradedPkg := analyzePackageCoverage(
		"example.com/pkg", ".", config.DefaultConfig(), &stderr, deps,
	)
	if reports != nil {
		t.Errorf("expected nil reports on analysis error, got %d reports", len(reports))
	}
	if degradedPkg != "" {
		t.Errorf("expected empty degradedPkg, got %q", degradedPkg)
	}
}

func TestAnalyzePackageCoverage_DI_EmptyResults(t *testing.T) {
	deps := successDeps()
	deps.loadAndAnalyze = func(_ string, _ analysis.Options) ([]taxonomy.AnalysisResult, error) {
		return []taxonomy.AnalysisResult{}, nil
	}
	var stderr bytes.Buffer
	reports, degradedPkg := analyzePackageCoverage(
		"example.com/pkg", ".", config.DefaultConfig(), &stderr, deps,
	)
	if reports != nil {
		t.Errorf("expected nil reports on empty results, got %d reports", len(reports))
	}
	if degradedPkg != "" {
		t.Errorf("expected empty degradedPkg, got %q", degradedPkg)
	}
}

func TestAnalyzePackageCoverage_DI_ClassifyNil(t *testing.T) {
	deps := successDeps()
	deps.classifyResults = func(_ []taxonomy.AnalysisResult, _ string, _ string, _ *config.GazeConfig) []taxonomy.AnalysisResult {
		return nil
	}
	var stderr bytes.Buffer
	reports, degradedPkg := analyzePackageCoverage(
		"example.com/pkg", ".", config.DefaultConfig(), &stderr, deps,
	)
	if reports != nil {
		t.Errorf("expected nil reports when classify returns nil, got %d reports", len(reports))
	}
	if degradedPkg != "" {
		t.Errorf("expected empty degradedPkg, got %q", degradedPkg)
	}
}

func TestAnalyzePackageCoverage_DI_LoadTestError(t *testing.T) {
	deps := successDeps()
	deps.loadTestPkg = func(_ string) (*packages.Package, error) {
		return nil, errors.New("no test files")
	}
	var stderr bytes.Buffer
	reports, degradedPkg := analyzePackageCoverage(
		"example.com/pkg", ".", config.DefaultConfig(), &stderr, deps,
	)
	if reports != nil {
		t.Errorf("expected nil reports on loadTestPkg error, got %d reports", len(reports))
	}
	if degradedPkg != "" {
		t.Errorf("expected empty degradedPkg, got %q", degradedPkg)
	}
}

func TestAnalyzePackageCoverage_DI_AssessError(t *testing.T) {
	deps := successDeps()
	deps.assess = func(_ []taxonomy.AnalysisResult, _ *packages.Package, _ quality.Options) ([]taxonomy.QualityReport, *taxonomy.PackageSummary, error) {
		return nil, nil, errors.New("assess failed")
	}
	var stderr bytes.Buffer
	reports, degradedPkg := analyzePackageCoverage(
		"example.com/pkg", ".", config.DefaultConfig(), &stderr, deps,
	)
	if reports != nil {
		t.Errorf("expected nil reports on assess error, got %d reports", len(reports))
	}
	if degradedPkg != "" {
		t.Errorf("expected empty degradedPkg, got %q", degradedPkg)
	}
}

func TestAnalyzePackageCoverage_DI_SSADegraded(t *testing.T) {
	deps := successDeps()
	deps.assess = func(_ []taxonomy.AnalysisResult, _ *packages.Package, _ quality.Options) ([]taxonomy.QualityReport, *taxonomy.PackageSummary, error) {
		reports := []taxonomy.QualityReport{
			{
				TestFunction: "TestDoWork",
				TargetFunction: taxonomy.FunctionTarget{
					Package:  "example.com/pkg",
					Function: "DoWork",
				},
			},
		}
		summary := &taxonomy.PackageSummary{
			TotalTests:  1,
			SSADegraded: true,
		}
		return reports, summary, nil
	}
	var stderr bytes.Buffer
	reports, degradedPkg := analyzePackageCoverage(
		"example.com/pkg", ".", config.DefaultConfig(), &stderr, deps,
	)
	if len(reports) == 0 {
		t.Fatal("expected non-empty reports when SSA is degraded")
	}
	if degradedPkg != "example.com/pkg" {
		t.Errorf("expected degradedPkg = %q, got %q", "example.com/pkg", degradedPkg)
	}
	// Verify the warning was emitted to stderr.
	if !strings.Contains(stderr.String(), "SSA degraded") {
		t.Errorf("expected SSA degradation warning in stderr, got %q", stderr.String())
	}
}

// ---------------------------------------------------------------------------
// loadTestPackage tests (Task 1.3)
// ---------------------------------------------------------------------------

func TestLoadTestPackage_WithTests(t *testing.T) {
	pkg, err := loadTestPackage("github.com/unbound-force/gaze/internal/quality/testdata/src/welltested")
	if err != nil {
		t.Fatalf("expected no error for package with tests, got: %v", err)
	}
	if pkg == nil {
		t.Fatal("expected non-nil package")
	}
}

func TestLoadTestPackage_WithoutTests(t *testing.T) {
	_, err := loadTestPackage("github.com/unbound-force/gaze/internal/analysis/testdata/src/returns")
	if err == nil {
		t.Fatal("expected error for package without test files")
	}
	if !strings.Contains(err.Error(), "no test files found") {
		t.Errorf("expected error containing %q, got %q", "no test files found", err.Error())
	}
}

func TestLoadTestPackage_NonExistent(t *testing.T) {
	_, err := loadTestPackage("github.com/nonexistent/does-not-exist")
	if err == nil {
		t.Fatal("expected error for non-existent package")
	}
}
