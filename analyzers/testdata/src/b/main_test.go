package b

import "testing"

func TestF1(t *testing.T) {
	tests := []struct {
		name string
		arg int
		want int
	}{
		{"A", 1, 5},
		{"B", 5, 25},
		{"C", 3, 15},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			if got := F1(c.arg); got != c.want {
				t.Errorf("F1() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestF2(t *testing.T) { // want `{type:Test, testFunc:"b:main_test.go:24:6 TestF2", targetFunc:"b:main.go:7:6 f2", CallPos:\["36:14"\]}`
	tests := []struct {
		name string
		arg int
		want int
	}{
		{"A", 1, 5},
		{"B", 5, 25},
		{"C", 3, 15},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			if got := f2(c.arg); got != c.want {
				t.Errorf("F1() = %v, want %v", got, c.want)
			}
		})
	}
}