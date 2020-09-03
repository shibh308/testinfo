package testinfo_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	"testinfo"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, testinfo.Analyzer, "a/test", "a/body")
}

