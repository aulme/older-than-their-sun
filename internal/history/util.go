package history

import "fmt"

func sprintf(format string, args ...any) string { return fmt.Sprintf(format, args...) }

func remove(xs []int, v int) []int {
	for i, x := range xs {
		if x == v {
			return append(xs[:i], xs[i+1:]...)
		}
	}
	return xs
}

func contains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func systems(n int) string {
	if n == 1 {
		return "a single world"
	}
	return sprintf("%d systems", n)
}
