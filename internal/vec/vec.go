// Package vec has the vector math retrieval is built on.
package vec

import "math"

func Dot(a, b []float64) float64 {
	var s float64
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

func Norm(a []float64) float64 {
	return math.Sqrt(Dot(a, a))
}

// Cosine is the angle-based similarity between two vectors:
// 1 = same direction, 0 = unrelated, -1 = opposite.
func Cosine(a, b []float64) float64 {
	na, nb := Norm(a), Norm(b)
	if na == 0 || nb == 0 {
		return 0
	}
	return Dot(a, b) / (na * nb)
}
