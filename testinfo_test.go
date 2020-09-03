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

// TestAnalyzer is a test for Analyzer with funcName flag.
func TestAnalyzerWithFuncNameFlag(t *testing.T) {
	testdata := analysistest.TestData()
	defer testinfo.ExportSetFuncNameFlag("TestF2")()
	analysistest.Run(t, testdata, testinfo.Analyzer, "b")
}

// TestAnalyzer is a test for Analyzer with fileName flag.
func TestAnalyzerWithFileNameFlag(t *testing.T) {
	testdata := analysistest.TestData()
	defer testinfo.ExportSetFileNameFlag("cx")()
	analysistest.Run(t, testdata, testinfo.Analyzer, "c")
}
