// Package bddstyle is a test fixture that simulates BDD framework patterns
// where test registration uses interface dispatch (dynamic calls invisible
// to SSA analysis). This causes InferTargets to return no targets.
package bddstyle

import "fmt"

// Runner is an interface that simulates a BDD test runner's registration
// mechanism. Calls through this interface use dynamic dispatch, which
// SSA's resolveCallee cannot resolve.
type Runner interface {
	It(description string, fn func())
}

// Calculator performs basic arithmetic with observable side effects.
type Calculator struct {
	LastResult int
}

// Add adds two numbers and stores the result.
func (c *Calculator) Add(a, b int) int {
	c.LastResult = a + b
	return c.LastResult
}

// Subtract subtracts b from a and stores the result.
func (c *Calculator) Subtract(a, b int) int {
	c.LastResult = a - b
	return c.LastResult
}

// Format returns a formatted string of the last result.
func (c *Calculator) Format() string {
	return fmt.Sprintf("result: %d", c.LastResult)
}
