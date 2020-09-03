package body

import (
	"testing"
)

func TestEval1(t *testing.T) { // want `{type:Test, testFunc:"body:body_test.go:7:6 TestEval1", targetFunc:"body:body.go:4:6 Eval1", CallPos:\["8:5", "14:38"\]}`
	if Eval1(5) != 25 {
		t.Errorf("NG")
	}
	if (func () func (int) int { return eval2 }()(3)) != 15 {
		t.Errorf("NG")
	}
	if (func () func (int) int { return Eval1 }()(3)) != 15 {
		t.Errorf("NG")
	}
}

func BenchmarkEval2(b *testing.B) { // want `{type:Benchmark, testFunc:"body:body_test.go:19:6 BenchmarkEval2", targetFunc:"body:body.go:8:6 eval2", CallPos:\["22:3"\]}`
	b.StartTimer()
	for i := 0; i < 100; i++ {
		eval2(i)
	}
	b.StopTimer()
}