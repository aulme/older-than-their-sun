package history

import (
	"testing"

	"worldgen/internal/species"
)

// TestCascadeUnwinds: when a war ends, the wars joined to it end too,
// whether it was a principal's war or itself an ally's (stage 3's first
// batch: allies joined against an ally's war fought on alone).
func TestCascadeUnwinds(t *testing.T) {
	w := newTestWorld(t, 75, 40)
	p := func(star int) *Civ { return spawnAt(w, star, species.Fixed("defensive")) }
	a, b, ally, counter := p(1), p(2), p(3), p(4)
	principal := w.declare(a, b, because("border"))
	joined := w.declare(ally, b, because("pact").By(a))
	joined.Principal = a.ID
	against := w.declare(counter, ally, because("pact").By(b))
	against.Principal = b.ID
	if principal == nil || joined == nil || against == nil {
		t.Fatal("no wars")
	}
	w.endWar(joined, "terms")
	if !against.Over {
		t.Error("the war joined against an ally's war outlived it")
	}
	if principal.Over {
		t.Error("an ally's peace ended its principal's war")
	}
}
