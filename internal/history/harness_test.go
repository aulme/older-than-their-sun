package history

import (
	"math/rand/v2"
	"sync"
	"testing"

	"worldgen/internal/galaxy"
	"worldgen/internal/species"
)

// The package's heavy tests read one generated run rather than each
// making its own: a 200-star age is a minute and a half of CPU, and
// most of what the tests ask of it is the same age. A test that reads
// the reference must not change it; one that needs a world of its own
// (the determinism check) makes it, and says why.
//
// reference is seed 7 at the full field, generated once however many
// tests ask for it. It is the seed the determinism check runs, so its
// first of three runs is this one: the age the lookups walk and the age
// the check compares are the same work, done once. Seed 7 is also the
// longest history of the seeds the suite reads, so the lookups and the
// parameter walk see the most kinds.
var reference = sync.OnceValue(func() *World {
	cfg := DefaultConfig()
	cfg.Stars = heavyStars()
	return Generate(referenceSeed, cfg)
})

const referenceSeed = 7

// heavyStars is the field a whole-run test spans. Under -short it is
// smaller, so that the suite can be run between edits; the gate before
// a commit is the full suite, and the tests that pin a digest ignore
// this and always run at 200, since their number is of that field.
func heavyStars() int {
	if testing.Short() {
		return 120
	}
	return 200
}

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
