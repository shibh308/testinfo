package analyzers_test

import (
	"github.com/shibh308/testinfo/analyzers"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzers.Analyzer, "a/test", "a/body")
}

// TestAnalyzer is a test for Analyzer with funcName flag.
func TestAnalyzerWithFuncNameFlag(t *testing.T) {
	testdata := analysistest.TestData()
	defer analyzers.ExportSetFuncNameFlag("TestF2")()
	analysistest.Run(t, testdata, analyzers.Analyzer, "b")
}

// TestAnalyzer is a test for Analyzer with fileName flag.
func TestAnalyzerWithFileNameFlag(t *testing.T) {
	testdata := analysistest.TestData()
	defer analyzers.ExportSetFileNameFlag("cx")()
	analysistest.Run(t, testdata, analyzers.Analyzer, "c")
}
