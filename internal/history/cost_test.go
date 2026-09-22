package history

import (
	"os"
	"testing"
	"time"

	"worldgen/internal/species"
)

// TestCostShape measures what a tick costs when the peoples and the
// worlds they hold are varied on their own. The observed runs cannot
// answer this: in a real age the two grow together, so nothing there
// separates the cost of one more people from the cost of one more
// world. Here the world is built to order and ticked cold — no tech
// learned, no tales held, nobody met — so what it measures is the
// structural cost of the per-people and per-holding passes, which is a
// floor: the passes that need a history (the lore's walks, the sorted
// walks of who has been met, the pairwise ones) are all per people, and
// they only widen the gap.
//
// It is a measurement, not a gate, so it runs only when asked:
//
//	COST=1 go test ./internal/history -run TestCostShape -v
func TestCostShape(t *testing.T) {
	if os.Getenv("COST") == "" {
		t.Skip("a measurement, not a gate: set COST=1 to run it")
	}
	type shape struct{ peoples, per int }
	shapes := []shape{
		{8, 8}, {16, 4}, {32, 2}, {64, 1}, // 64 worlds, split four ways
		{8, 16}, {16, 8}, {32, 4}, {64, 2}, // 128 worlds, the same
		{8, 1}, {8, 2}, {8, 4}, // worlds alone, peoples fixed
		{16, 1}, {32, 1}, // peoples alone, one world each
	}
	t.Logf("%8s %8s %8s %12s %12s %12s", "peoples", "per", "worlds", "us/tick", "us/people", "us/world")
	for _, s := range shapes {
		us := costOf(t, s.peoples, s.per)
		worlds := s.peoples * s.per
		t.Logf("%8d %8d %8d %12.1f %12.2f %12.2f",
			s.peoples, s.per, worlds, us, us/float64(s.peoples), us/float64(worlds))
	}
}

// costOf builds a world of peoples holding per worlds each and returns
// the microseconds a tick takes, averaged over a run of them.
func costOf(t *testing.T, peoples, per int) float64 {
	t.Helper()
	w := newTestWorld(t, 4242, 900)
	star := 0
	next := func() int { // a star nobody holds yet
		for star < len(w.G.Stars) && w.Owner[star] >= 0 {
			star++
		}
		if star >= len(w.G.Stars) {
			t.Fatalf("the field ran out of stars at %d peoples of %d worlds", peoples, per)
		}
		star++
		return star - 1
	}
	for range peoples {
		c := spawnAt(w, next(), species.GenerateWith(w.R, 1, "lush", species.Biological, 0))
		for range per - 1 {
			w.holdWorld(c, next())
		}
	}
	const warm, runs = 20, 200
	w.ticks(warm) // let the first tick's one-off work fall out of the measure
	start := time.Now()
	w.ticks(runs)
	return float64(time.Since(start).Microseconds()) / runs
}
