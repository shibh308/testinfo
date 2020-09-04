package analyzers

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"path/filepath"
	"strings"
)

const doc = "analyzers is ..."

var Analyzer = &analysis.Analyzer{
	Name: "analyzers",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

var flags struct {
	funcName string
	fileName string
}

func init() {
	Analyzer.Flags.StringVar(&flags.funcName, "testfunc", flags.funcName, "test function name")
	Analyzer.Flags.StringVar(&flags.fileName, "testfile", flags.fileName, "target testfile name")
}

func run(pass *analysis.Pass) (interface{}, error) {

	filterFunc := func(path string) bool{return true}
	if flags.fileName != "" {
		filterFunc = func(path string) bool{return strings.HasPrefix(filepath.Base(path), flags.fileName)}
	}
	testInfo, err := New(pass, filterFunc)

	if err != nil {
		return nil, err
	}

	if flags.funcName != "" {
		x := testInfo.GetFuncDataFromName(flags.funcName)
		if x != nil {
			pass.Reportf(x.TestDecl.Pos(), testInfo.Format(*x))
		}
	} else {
		for _, x := range testInfo.FuncData {
			pass.Reportf(x.TestDecl.Pos(), testInfo.Format(*x))
		}
	}

	return nil, nil
}

