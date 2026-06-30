package crap

import (
	"testing"
)

func BenchmarkFormula(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Formula(10, 50)
	}
}

func BenchmarkClassifyQuadrant(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ClassifyQuadrant(20, 10, 15, 15)
	}
}

func BenchmarkBuildSummary(b *testing.B) {
	scores := make([]Score, 100)
	for i := range scores {
		scores[i] = Score{
			Function:     "Func",
			Complexity:   i + 1,
			LineCoverage: float64(i),
			CRAP:         Formula(i+1, float64(i)),
		}
	}
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildSummary(scores, opts, nil)
	}
}

func BenchmarkBuildCoverMap(b *testing.B) {
	coverages := make([]FuncCoverage, 200)
	for i := range coverages {
		coverages[i] = FuncCoverage{
			File:       "file.go",
			StartLine:  i + 1,
			Percentage: float64(i) / 2.0,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildCoverMap(coverages)
	}
}

func BenchmarkIsGeneratedFile_NotGenerated(b *testing.B) {
	// Use this test file itself as a non-generated file.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isGeneratedFile("bench_test.go")
	}
}
