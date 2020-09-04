package main

import (
	"github.com/shibh308/testinfo/analyzers"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(analyzers.Analyzer)
}

