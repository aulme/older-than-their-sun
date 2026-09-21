package history

import (
	"math/rand/v2"
	"testing"

	"worldgen/internal/galaxy"
	"worldgen/internal/species"
)

// newTestWorld builds a small real galaxy and an empty world on it, with the
// cycle set so fertility works and the tick's phases in place. The deep pass
// and the earlier ages are not run: the substrate is bare, life is only at
// Sol, and no legacies are buried. Now is the dawn and dt is one tick.
func newTestWorld(t *testing.T, seed uint64, stars int) *World {
	t.Helper()
	cfg := DefaultConfig()
	cfg.Stars = stars
	w := newWorld(seed, cfg)
	w.dt = float64(cfg.Step) / 1000
	w.Now = cfg.Dawn
	return w
}

// tick runs the phases once at Now and advances Now by a step.
func (w *World) tick() {
	w.runPhases()
	w.Now += w.Cfg.Step
}

// ticks runs n ticks.
func (w *World) ticks(n int) {
	for range n {
		w.tick()
	}
}

// spawnAt raises a people of a known species at a star. The star's worlds
// are made fit to live on as the sim does for a natural birth.
func spawnAt(w *World, star int, sp *species.Species) *Civ {
	w.Bio[star] = BioComplex
	return w.spawnCiv(star, sp, -1)
}

// testGalaxy is a galaxy alone, for tests that want geometry and nothing else.
func testGalaxy(seed uint64, stars int) *galaxy.Galaxy {
	r := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
	rg, _ := galaxy.RegionByName("sol")
	return galaxy.GenerateAt(r, rg, stars, DefaultConfig().Radius, DefaultConfig().Thickness)
}

// lastFacts is the last n facts recorded, oldest first; nil where there
// are fewer.
func lastFacts(w *World, n int) []*Event {
	out := make([]*Event, n)
	i := n - 1
	for j := len(w.Events) - 1; j >= 0 && i >= 0; j-- {
		if w.Events[j].IsFact() {
			out[i] = w.Events[j]
			i--
		}
	}
	return out
}
