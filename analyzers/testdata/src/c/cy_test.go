package c

import "testing"

func Test_Cy1(t *testing.T) {
	if cy1(3) != 30 {
		t.Errorf("not correct")
	}
	if cy1(5) != 50 {
		t.Errorf("not correct")
	}
}

func TestCy2(t *testing.T) {
	if Cy2(3) != 9 {
		t.Errorf("not correct")
	}
	if Cy2(5) != 15 {
		t.Errorf("not correct")
	}
}
