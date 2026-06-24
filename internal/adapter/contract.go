package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/unbound-force/gaze/internal/crap"
	"github.com/unbound-force/gaze/internal/protocol"
	"github.com/unbound-force/gaze/internal/quality"
	"github.com/unbound-force/gaze/internal/taxonomy"
)

// ExternalContractCoverageProvider implements
// crap.ContractCoverageProvider by using the external analyzer's
// test_mapping data combined with side effect analysis results.
//
// When the analyzer supports test_mapping (capability is true), the
// provider calls the test_mapping method, converts the response to
// taxonomy.AssertionMapping entries, and uses
// quality.ComputeContractCoverage to build the lookup function.
//
// When test_mapping is not supported, the provider returns a no-op
// lookup that always returns (zero, false) — GazeCRAP is unavailable.
//
// Design decision D4: This adapter uses ExternalSideEffectAnalyzer
// as a composition dependency to access the full analysis results.
type ExternalContractCoverageProvider struct {
	client         *protocol.Client
	caps           protocol.Capabilities
	sideEffects    *ExternalSideEffectAnalyzer
	rootDir        string
	patterns       []string
	stderr         io.Writer
}

// NewExternalContractCoverageProvider creates a contract coverage
// provider that delegates to the given protocol client and side
// effect analyzer.
func NewExternalContractCoverageProvider(
	client *protocol.Client,
	caps protocol.Capabilities,
	sideEffects *ExternalSideEffectAnalyzer,
	rootDir string,
	patterns []string,
	stderr io.Writer,
) *ExternalContractCoverageProvider {
	return &ExternalContractCoverageProvider{
		client:      client,
		caps:        caps,
		sideEffects: sideEffects,
		rootDir:     rootDir,
		patterns:    patterns,
		stderr:      stderr,
	}
}

// Build implements crap.ContractCoverageProvider. When test_mapping
// is supported, it fetches assertion mappings and computes contract
// coverage per function. When not supported, returns a no-op lookup.
func (p *ExternalContractCoverageProvider) Build(patterns []string, rootDir string) (func(pkg, function string) (crap.ContractCoverageInfo, bool), []string, error) {
	if !p.caps.TestMapping {
		return noopLookup(), nil, nil
	}

	mappings, err := p.fetchTestMappings(patterns, rootDir)
	if err != nil {
		// fetchTestMappings handles graceful degradation (D7)
		// and returns nil mappings on optional method failure.
		return noopLookup(), nil, nil
	}

	allResults, err := p.sideEffects.AllResults()
	if err != nil {
		return nil, nil, fmt.Errorf("fetching side effects for contract coverage: %w", err)
	}

	lookup := buildContractLookup(allResults, mappings)
	return lookup, nil, nil
}

// noopLookup returns a contract coverage lookup that always returns
// (zero, false). Used when test_mapping is unavailable or fails.
func noopLookup() func(pkg, function string) (crap.ContractCoverageInfo, bool) {
	return func(pkg, function string) (crap.ContractCoverageInfo, bool) {
		return crap.ContractCoverageInfo{}, false
	}
}

// fetchTestMappings calls the test_mapping protocol method and parses
// the response. Returns nil on any failure (graceful degradation per D7).
func (p *ExternalContractCoverageProvider) fetchTestMappings(patterns []string, rootDir string) ([]protocol.AssertionMappingData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), protocol.AnalysisTimeout)
	defer cancel()

	resp, err := p.client.Call(ctx, protocol.MethodTestMapping, protocol.TestMappingParams{
		RootPath: rootDir,
		Patterns: patterns,
	})
	if err != nil {
		p.warn("test_mapping failed: %v", err)
		return nil, err
	}
	if resp.Error != nil {
		p.warn("test_mapping error: %s", resp.Error.Message)
		return nil, fmt.Errorf("test_mapping: %s", resp.Error.Message)
	}

	var result protocol.TestMappingResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		p.warn("parsing test_mapping result: %v", err)
		return nil, err
	}
	return result.Mappings, nil
}

// warn emits a warning to stderr if a writer is configured.
func (p *ExternalContractCoverageProvider) warn(format string, args ...any) {
	if p.stderr != nil {
		_, _ = fmt.Fprintf(p.stderr, "warning: "+format+"\n", args...)
	}
}

// buildContractLookup creates a lookup function that returns
// contract coverage info for a given (pkg, function) pair. It
// groups side effects and assertion mappings by function, then
// computes contract coverage for each.
func buildContractLookup(
	results []taxonomy.AnalysisResult,
	mappings []protocol.AssertionMappingData,
) func(pkg, function string) (crap.ContractCoverageInfo, bool) {
	type funcKey struct {
		pkg      string
		function string
	}

	// Group side effects by function.
	effectsByFunc := make(map[funcKey][]taxonomy.SideEffect)
	for _, r := range results {
		key := funcKey{pkg: r.Target.Package, function: r.Target.Function}
		effectsByFunc[key] = append(effectsByFunc[key], r.SideEffects...)
	}

	// Convert protocol mappings to taxonomy mappings, grouped by
	// target function.
	mappingsByFunc := make(map[funcKey][]taxonomy.AssertionMapping)
	for _, m := range mappings {
		key := funcKey{pkg: m.TargetPackage, function: m.TargetFunction}
		mappingsByFunc[key] = append(mappingsByFunc[key], taxonomy.AssertionMapping{
			AssertionLocation: m.AssertionLocation,
			AssertionType:     taxonomy.AssertionType(m.AssertionType),
			SideEffectID:      findSideEffectID(effectsByFunc[key], m.SideEffectType),
			Confidence:        m.Confidence,
		})
	}

	// Pre-compute contract coverage for each function.
	coverageByFunc := make(map[funcKey]crap.ContractCoverageInfo)
	for key, effects := range effectsByFunc {
		cc := quality.ComputeContractCoverage(effects, mappingsByFunc[key])
		coverageByFunc[key] = crap.ContractCoverageInfo{
			Percentage: cc.Percentage,
		}
	}

	return func(pkg, function string) (crap.ContractCoverageInfo, bool) {
		info, ok := coverageByFunc[funcKey{pkg: pkg, function: function}]
		return info, ok
	}
}

// findSideEffectID finds the ID of the first side effect matching
// the given type in the effects slice. Returns empty string if not
// found. This bridges the protocol's type-based mapping to the
// taxonomy's ID-based mapping.
func findSideEffectID(effects []taxonomy.SideEffect, sideEffectType string) string {
	for _, e := range effects {
		if string(e.Type) == sideEffectType {
			return e.ID
		}
	}
	return ""
}
