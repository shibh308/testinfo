package main

import (
	"github.com/shibh308/testinfo"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(testinfo.Analyzer)
}

