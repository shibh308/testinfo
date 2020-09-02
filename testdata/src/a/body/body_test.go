package body

import (
	"testing"
)

func TestEval1(t *testing.T) {
	if Eval1(5) != 25 {
		t.Errorf("NG")
	}
	if (func () func (int) int { return eval2 }()(3)) != 15 {
		t.Errorf("NG")
	}
}

func BenchmarkEval2(b *testing.B) {
	b.StartTimer()
	for i := 0; i < 100; i++ {
		eval2(i)
	}
	b.StopTimer()
}