package body_test

import (
	"testinfo/testdata/src/a/body"
	"testing"
)

func TestEval1(t *testing.T) {
	if body.Eval1(2) != 10 {
		t.Errorf("NG")
	}
}
