package adapter

import (
	"os/exec"

	"github.com/unbound-force/gaze/internal/config"
)

// Discover resolves the external analyzer binary and arguments using
// a three-tier discovery mechanism (design decision D5):
//
//  1. CLI flag: --analyzer <name> overrides everything.
//  2. Config: .gaze.yaml analyzers.<language>.command.
//  3. PATH convention: gaze-analyzer-<language>.
//
// Returns the binary name/path and args to pass to NewSession. If no
// analyzer is found at any tier, returns empty strings and nil args
// (no error) — the caller should fall back to Go providers.
//
// The language parameter is required for tiers 2 and 3. If
// analyzerFlag is non-empty, language is not needed (tier 1 wins).
func Discover(analyzerFlag, language string, cfg *config.GazeConfig) (binary string, args []string, err error) {
	// Tier 1: CLI flag overrides everything.
	if analyzerFlag != "" {
		return analyzerFlag, []string{"--stdio"}, nil
	}

	// Tier 2: Config-based lookup.
	if language != "" && cfg != nil && cfg.Analyzers != nil {
		if entry, ok := cfg.Analyzers[language]; ok && entry.Command != "" {
			return entry.Command, entry.Args, nil
		}
	}

	// Tier 3: PATH convention — gaze-analyzer-<language>.
	if language != "" {
		conventionName := "gaze-analyzer-" + language
		if _, lookErr := exec.LookPath(conventionName); lookErr == nil {
			return conventionName, []string{"--stdio"}, nil
		}
	}

	// No analyzer found — caller should use Go providers.
	return "", nil, nil
}
