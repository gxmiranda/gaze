package goprovider_test

import (
	"github.com/unbound-force/gaze/internal/crap"
	"github.com/unbound-force/gaze/internal/provider/goprovider"
)

// Compile-time interface satisfaction checks. These verify that all
// four Go provider adapters implement their respective interfaces.
// Behavioral correctness is verified by the E2E self-check (task 5.2)
// which exercises the full pipeline with Go providers.
var (
	_ crap.ComplexityProvider       = (*goprovider.GoComplexityProvider)(nil)
	_ crap.LineCoverageProvider     = (*goprovider.GoLineCoverageProvider)(nil)
	_ crap.SideEffectAnalyzer       = (*goprovider.GoSideEffectAnalyzer)(nil)
	_ crap.ContractCoverageProvider = (*goprovider.GoContractCoverageProvider)(nil)
)
