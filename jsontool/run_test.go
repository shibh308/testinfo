package jsontool_test

import (
	"fmt"
	"github.com/shibh308/testinfo/jsontool"
	"testing"
)

func TestRun(t *testing.T) {
	cases := []struct {
		Name string
		File string
		Pos  int
		Want int
	}{
		{"sample1", "./run.go", 400, 10},
		{"sample1", "./run.go", 500, 10},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			jsontool.SetFlags(jsontool.Flags{
				Parse:  false,
				Path:   c.File,
				Offset: c.Pos,
			})
			str, err := jsontool.Run()
			fmt.Println(str)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
