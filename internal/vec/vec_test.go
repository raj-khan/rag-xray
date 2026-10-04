package vec

import (
	"math"
	"testing"
)

func TestCosine(t *testing.T) {
	cases := []struct {
		name string
		a, b []float64
		want float64
	}{
		{"same direction", []float64{1, 2}, []float64{2, 4}, 1},
		{"orthogonal", []float64{1, 0}, []float64{0, 1}, 0},
		{"opposite", []float64{1, 1}, []float64{-1, -1}, -1},
		{"zero vector", []float64{0, 0}, []float64{1, 1}, 0},
	}
	for _, c := range cases {
		if got := Cosine(c.a, c.b); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s: Cosine = %v, want %v", c.name, got, c.want)
		}
	}
}
