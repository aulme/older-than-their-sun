package history

import (
	"testing"

	"worldgen/internal/flow"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// TestGrantHalvesOnce: a rarity had halves the price of what it grants;
// a second instance of the same rarity adds nothing, in price or levels.
func TestGrantHalvesOnce(t *testing.T) {
	w := newTestWorld(t, 14, 30)
	c := spawnAt(w, 0, species.Fixed("cooperative"))
	n := tech.Get("causal_physics")
	full := w.price(c, n)
	w.flows(c)
	soc := c.Soc
	w.addSource(&Source{Key: "horizon", Name: "a horizon", Kind: CosmicSource, Star: 0, Rarity: true, Grants: []string{"causal_physics"}, Levels: [3]float64{0, 0, 0.5}, Holder: -1, Carried: -1, Legacy: -1})
	w.flows(c)
	if got := w.price(c, n); got != full/2 {
		t.Errorf("with the horizon causal physics costs %v, want half of %v", got, full)
	}
	if c.Soc != soc+0.5 {
		t.Errorf("with the horizon Soc is %v, was %v: want +0.5", c.Soc, soc)
	}
	w.addSource(&Source{Key: "horizon", Name: "another horizon", Kind: CosmicSource, Star: 0, Rarity: true, Grants: []string{"causal_physics"}, Levels: [3]float64{0, 0, 0.5}, Holder: -1, Carried: -1, Legacy: -1})
	w.flows(c)
	if got := w.price(c, n); got != full/2 {
		t.Errorf("with two horizons causal physics costs %v, want still half", got)
	}
	if c.Soc != soc+0.5 {
		t.Errorf("with two horizons Soc is %v: a second instance is worth nothing", c.Soc)
	}
	if !c.Had["horizon"] {
		t.Error("the horizon was not written as had")
	}
}

// TestSourcesUngated: nothing a natural source needs is hard-gated by a
// rarity. tech.TestNoCatch22 checks the structures; this checks the rest.
func TestSourcesUngated(t *testing.T) {
	sources, _ := naturalSources(testGalaxy(15, 60))
	for _, s := range sources {
		for _, k := range append(append([]string{}, s.Needs...), s.With...) {
			for _, n := range tech.Closure(k) {
				if g := tech.Get(n).Gated; g != "" {
					t.Errorf("%s needs %s, gated by %s", s.Key, n, g)
				}
			}
		}
	}
}

// TestFleetLostBuriesRarity: a fleet lost with a mobile rarity aboard
// leaves it buried where the fleet died, and its holder no longer has it.
func TestFleetLostBuriesRarity(t *testing.T) {
	w := newTestWorld(t, 16, 30)
	c := spawnAt(w, 0, species.Fixed("cooperative"))
	e := spawnAt(w, 1, species.Fixed("cooperative"))
	l := &Legacy{ID: len(w.Legacies), Age: 0, Maker: -1, Kind: Artifact, Star: 0, Node: "fusion", Desc: "a seed of grey metal", Horror: -1, Finder: -1, Source: -1, State: Wielded, Level: "sur"}
	w.Legacies = append(w.Legacies, l)
	c.Wielded = append(c.Wielded, l)
	w.wieldRarity(c, l)
	w.flows(c)
	if !c.has("artifact") {
		t.Fatal("the wielded artifact is not had")
	}
	s := w.Sources[l.Source]
	x := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: e.ID, Kind: Campaign, Star: 1, From: 0, Mil: 0.5, Base: 1, Seen: map[int]bool{}}
	w.Expeditions = append(w.Expeditions, x)
	s.Carried, s.Star = x.ID, -1
	w.resolve(x)
	if !x.Over {
		t.Fatal("a fleet of half a level with nothing held did not end")
	}
	if l.State != Buried || l.Star != 1 {
		t.Errorf("the artifact is %s at %d, want buried at 1", l.State, l.Star)
	}
	if s.Holder != -1 || s.Star != 1 || s.Carried != -1 {
		t.Errorf("the source is held by %d at %d aboard %d, want nobody at 1 aboard none", s.Holder, s.Star, s.Carried)
	}
	if len(c.Wielded) != 0 {
		t.Error("the people still wields what went down with the fleet")
	}
	w.flows(c)
	if c.has("artifact") {
		t.Error("the people still has the artifact")
	}
}

// TestGrazing: a nomad fleet at an unowned star grazes half its yield, at
// a partner's a quarter, at a stranger's nothing.
func TestGrazing(t *testing.T) {
	w := newTestWorld(t, 17, 30)
	c := spawnAt(w, 0, species.Fixed("cooperative"))
	o := spawnAt(w, 1, species.Fixed("cooperative"))
	full := w.yieldAt(c, 0)
	if full == (flow.Income{}) {
		t.Fatal("the cradle yields nothing")
	}
	w.takeSky(c, "")
	if !c.Aloft {
		t.Fatal("the people did not take to the sky")
	}
	if got := w.yieldAt(c, 0); got != full.Scale(grazeFree) {
		t.Errorf("grazing the empty cradle gives %v, want half of %v", got, full)
	}
	x := w.fleets(c)[0]
	x.Base = 1
	if got := w.yieldAt(c, 1); got != (flow.Income{}) {
		t.Errorf("grazing a stranger's star gives %v, want nothing", got)
	}
	c.Trade[o.ID] = true
	if got, want := w.yieldAt(c, 1), w.yieldAt(o, 1).Scale(grazePartner); got[flow.O] != want[flow.O] {
		t.Errorf("grazing a partner's star gives %v O, want a quarter: %v", got[flow.O], want[flow.O])
	}
}
