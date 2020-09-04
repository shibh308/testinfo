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
		{"a_f", "./testdata/a/a.go", 40, 10},
		{"a_G", "./testdata/a/a.go", 80, 10},
		{"a_A", "./testdata/a/a.go", 110, 10},
		{"a_B", "./testdata/a/a.go", 130, 10},
		{"a_testF", "./testdata/a/a_test.go", 320, 10},
		{"a_testG", "./testdata/a/a_test.go", 30, 10},
		{"a_testA", "./testdata/a/another_test.go", 110, 10},
		{"a_testB", "./testdata/a/another_test.go", 250, 10},
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
