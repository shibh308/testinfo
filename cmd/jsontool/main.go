package main

import (
	"fmt"
	"github.com/shibh308/testinfo/jsontool"
	"os"
)

func main() {
	str, err := jsontool.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(str)
	os.Exit(0)
}
