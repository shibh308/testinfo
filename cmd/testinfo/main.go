package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"
	"testinfo"
)

func main() {
	unitchecker.Main(testinfo.Analyzer)
}

