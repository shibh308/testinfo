package c

import "testing"

func Test_Cx1(t *testing.T) { // want `{type:Test, testFunc:"c:cx_test.go:5:6 Test_Cx1", targetFunc:"c:cx.go:3:6 cx1", CallPos:\["6:5", "9:5"\]}`
	if cx1(3) != 30 {
		t.Errorf("not correct")
	}
	if cx1(5) != 50 {
		t.Errorf("not correct")
	}
}

func TestCx2(t *testing.T) { // want `{type:Test, testFunc:"c:cx_test.go:14:6 TestCx2", targetFunc:"c:cx.go:7:6 Cx2", CallPos:\["15:5", "18:5"\]}`
	if Cx2(3) != 9 {
		t.Errorf("not correct")
	}
	if Cx2(5) != 15 {
		t.Errorf("not correct")
	}
}
