package battle

import (
	"math"
	"math/rand/v2"
	"testing"
)

// TestQuality: three levels are twice the ship, ten nine times, and
// Levels inverts it.
func TestQuality(t *testing.T) {
	if q := Quality(3); math.Abs(q-1.953) > 0.001 {
		t.Errorf("three levels: %.3f", q)
	}
	if q := Quality(10); math.Abs(q-9.31) > 0.01 {
		t.Errorf("ten levels: %.3f", q)
	}
	if l := Levels(Quality(4.5)); math.Abs(l-4.5) > 1e-9 {
		t.Errorf("levels of the quality of 4.5: %.3f", l)
	}
	if s := Strength(4, Quality(2)); math.Abs(s-6.25) > 1e-9 {
		t.Errorf("four ships two levels up: %.3f", s)
	}
}

// TestRollTable: four ships a side; the attacker wins about half at no gap,
// about nineteen in twenty three levels better, always five better; and a
// strength of nothing never wins.
func TestRollTable(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 3))
	const n = 20000
	wins := func(gap float64) float64 {
		w := 0
		for range n {
			if Roll(r, Strength(4, Quality(gap)), Strength(4, 1)) {
				w++
			}
		}
		return float64(w) / n
	}
	cases := []struct {
		gap      float64
		lo, hi   float64
		describe string
	}{
		{0, 0.47, 0.53, "even"},
		{1, 0.65, 0.80, "one better"},
		{3, 0.93, 0.995, "three better"},
		{5, 0.99, 1.0001, "five better"},
	}
	for _, c := range cases {
		if p := wins(c.gap); p < c.lo || p > c.hi {
			t.Errorf("%s: the attacker wins %.3f, want %.2f to %.2f", c.describe, p, c.lo, c.hi)
		}
	}
	// three to one at equal arts loses; five levels better makes it even
	if p := func() float64 {
		w := 0
		for range n {
			if Roll(r, Strength(1, Quality(5)), Strength(3, 1)) {
				w++
			}
		}
		return float64(w) / n
	}(); p < 0.45 || p > 0.60 {
		t.Errorf("one ship five levels better against three: %.3f", p)
	}
	if Roll(r, 0, 1) || !Roll(r, 1, 0) {
		t.Error("nothing won, or lost to nothing")
	}
}

// TestLosses: each side's loss is uniform to half the other's strength,
// and is paid in ships at its own quality, never more than it has.
func TestLosses(t *testing.T) {
	r := rand.New(rand.NewPCG(4, 4))
	const n = 20000
	sa, sd := 0.0, 0.0
	for range n {
		la, ld := Losses(r, 10, 4)
		if la < 0 || la > 2 || ld < 0 || ld > 5 {
			t.Fatalf("losses out of range: %.2f %.2f", la, ld)
		}
		sa += la
		sd += ld
	}
	if m := sa / n; math.Abs(m-1) > 0.05 {
		t.Errorf("the attacker's mean loss %.3f, want 1", m)
	}
	if m := sd / n; math.Abs(m-2.5) > 0.1 {
		t.Errorf("the defender's mean loss %.3f, want 2.5", m)
	}
	sum := 0
	for range n {
		sum += ToShips(r, 1.5, 1, 10)
	}
	if m := float64(sum) / n; math.Abs(m-1.5) > 0.05 {
		t.Errorf("a loss of 1.5 at quality one is %.3f ships on average, want 1.5", m)
	}
	if k := ToShips(r, 100, 1, 3); k != 3 {
		t.Errorf("a loss past the fleet takes %d ships, want 3", k)
	}
	if k := ToShips(r, 3, Quality(3), 10); k > 2 {
		t.Errorf("a loss of three at three levels up is %d ships, want at most 2", k)
	}
}

// TestSilosHold: two silos at equal arts hold against one ship about nine
// times in ten and against three about one in eight; against one ship
// four levels better about one in three. A grid of three repaired a gun
// a thousand years holds against four ships at equal arts nearly always.
func TestSilosHold(t *testing.T) {
	r := rand.New(rand.NewPCG(9, 9))
	const n = 10000
	holds := func(ships int, gap float64, guns int, repair bool) float64 {
		h := 0
		for range n {
			if Hold(r, ships, Quality(gap), guns, 1, repair) {
				h++
			}
		}
		return float64(h) / n
	}
	cases := []struct {
		ships    int
		gap      float64
		guns     int
		repair   bool
		lo, hi   float64
		describe string
	}{
		{1, 0, 2, false, 0.84, 0.92, "two silos against one ship"},
		{3, 0, 2, false, 0.09, 0.18, "two silos against three"},
		{1, 4, 2, false, 0.28, 0.42, "two silos against one ship four levels better"},
		{4, 0, 3, true, 0.88, 0.97, "a grid of three, repaired, against four"},
	}
	for _, c := range cases {
		if p := holds(c.ships, c.gap, c.guns, c.repair); p < c.lo || p > c.hi {
			t.Errorf("%s: holds %.3f, want %.2f to %.2f", c.describe, p, c.lo, c.hi)
		}
	}
}
