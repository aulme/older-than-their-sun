package flow

import (
	"math"
	"testing"
)

// TestShare: never more than the surplus times the cap goes out, the split
// follows the want share, nobody gets more than it wants, and a partner
// out of reach gets nothing.
func TestShare(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
	sum := func(xs []float64) float64 {
		s := 0.0
		for _, x := range xs {
			s += x
		}
		return s
	}
	// plenty wanted, little to give: the cap binds and the split is by share
	got := Share(10, []float64{6, 2}, []float64{0.25, 0.25})
	if !near(sum(got), 2.5) || !near(got[0], 2.5*0.75) || !near(got[1], 2.5*0.25) {
		t.Errorf("capped: %v", got)
	}
	// little wanted, plenty to give: each gets what it wants and no more
	got = Share(100, []float64{3, 1}, []float64{1, 1})
	if !near(got[0], 3) || !near(got[1], 1) {
		t.Errorf("wants met: %v", got)
	}
	// a partner out of reach gets nothing, and the other's share is not diluted by it
	got = Share(10, []float64{6, 2}, []float64{0.25, 0})
	if got[1] != 0 || !near(got[0], 2.5) {
		t.Errorf("out of reach: %v", got)
	}
	// a nomad's smaller cap clips its own take, and the total never passes the surplus
	got = Share(10, []float64{50, 50}, []float64{1, 0.5})
	if !near(got[0], 5) || !near(got[1], 5) || sum(got) > 10+1e-9 {
		t.Errorf("mixed caps: %v", got)
	}
	// nothing spare, nothing sent; nothing wanted, nothing sent
	if got = Share(0, []float64{1}, []float64{1}); got[0] != 0 {
		t.Error("sent from nothing")
	}
	if got = Share(5, []float64{0}, []float64{1}); got[0] != 0 {
		t.Error("sent to nobody wanting")
	}
}
