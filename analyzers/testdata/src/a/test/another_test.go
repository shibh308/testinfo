package body_test

import (
	"a/body"
	"testing"
)

func TestEval1(t *testing.T) { // want `{type:Test, testFunc:"body_test:another_test.go:8:6 TestEval1", targetFunc:"body:body.go:4:6 Eval1", CallPos:\["9:10"\]}`

	if body.Eval1(2) != 10 {
		t.Errorf("NG")
	}
}
