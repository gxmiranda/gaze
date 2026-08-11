package bddstyle_test

import (
	"testing"
)

// These tests simulate BDD-style test suites where the test function
// only calls framework registration functions (not in the target package),
// making InferTargets return zero candidates. This is the pattern seen
// with Ginkgo: TestSuite only calls RegisterFailHandler and RunSpecs.

// register simulates a BDD framework's test registration. This function
// is in the test package (bddstyle_test), not the target package
// (bddstyle), so walkCalls won't count it as a candidate target.
func register(name string, fn func()) {
	_ = name
	fn()
}

// TestCalculatorSuite simulates a Ginkgo-style test suite entry point.
// It only calls register() which is in the _test package, not the
// target package — so InferTargets finds no candidates.
func TestCalculatorSuite(t *testing.T) {
	register("Calculator.Add", func() {
		// The actual target calls are here, inside a closure registered
		// with a framework function — invisible to InferTargets because
		// walkCalls only walks the top-level test function and helpers.
		if 2+3 != 5 {
			t.Error("math is broken")
		}
	})
}

// TestFormatSuite is another test function that only calls register().
func TestFormatSuite(t *testing.T) {
	register("Calculator.Format", func() {
		_ = "formatted"
	})
}
