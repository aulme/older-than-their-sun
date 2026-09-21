package history

import (
	"strings"
	"testing"

	"worldgen/internal/flow"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Kinds, stages 6 and 7: the evolver that drifts and the anti-memetic
// fought by the shape of the hole.

// TestEvolverDrifts: an evolver's trait set differs after a long run, its
// old acquaintances feel a quarter more difference per drift, and its
// kin keep the shape they had.
func TestEvolverDrifts(t *testing.T) {
	w := newTestWorld(t, 60, 40)
	c := spawnAt(w, 1, fixedWith(species.Biological, species.Evolver, nil, "defensive", "practical", "curious", "hardy"))
	e := spawnAt(w, 2, species.Fixed("defensive", "faithful", "cautious"))
	kin := spawnAt(w, 3, c.Species)
	starfaring(w, c)
	starfaring(w, e)
	starfaring(w, kin)
	w.fathomed(e, c, "test")
	before := c.Species.Describe()
	base := e.differs(c)
	w.ticks(2000)
	if !c.Active() {
		t.Skipf("the evolver ended: %s", c.Cause)
	}
	if c.Drifts == 0 || c.Tally.Drifts != c.Drifts {
		t.Fatalf("no drift in two million years: %d, tally %d", c.Drifts, c.Tally.Drifts)
	}
	if c.Species.Describe() == before {
		t.Fatalf("drifted %d times and read the same: %s", c.Drifts, before)
	}
	if kin.Species == c.Species || kin.Species.Describe() != before {
		t.Fatalf("the kin drifted with them: %s", kin.Species.Describe())
	}
	if e.Active() && e.Fathomed[c.ID] {
		if got := e.differs(c) - base; got < 0.25*float64(c.Drifts)-1e-9 || got > 0.25*float64(c.Drifts)+1.5 {
			t.Fatalf("difference grew by %g over %d drifts", got, c.Drifts)
		}
	}
	found := false
	for _, f := range w.Events {
		if f.Kind == FDrifted && f.Subject == c.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("no fact of the drift")
	}
}

// reaching gives a people the stars, and a well at home rich enough that
// nothing it has is ever laid up.
func reaching(w *World, c *Civ) {
	starfaring(w, c)
	c.Known["slow_interstellar"] = true
	w.addSource(&Source{Key: "well", Name: "a well", Kind: CosmicSource, Star: c.Home, Yield: flow.Income{1000, 1000, 1000}, Holder: -1, Carried: -1, Legacy: -1})
	w.recompute(c)
}

// antimemeticPair is a conscious people at star a and an anti-memetic one
// at star b, both reaching and in touch.
func antimemeticPair(t *testing.T, seed uint64, a, b int) (*World, *Civ, *Civ) {
	t.Helper()
	w := newTestWorld(t, seed, 40)
	c := spawnAt(w, a, species.Fixed("defensive", "practical", "curious"))
	x := spawnAt(w, b, fixedWith(species.Biological, species.Antimemetic, nil, "opportunist", "practical", "cautious"))
	reaching(w, c)
	reaching(w, x)
	return w, c, x
}

// TestUnseen: a conscious people never has Met set for an anti-memetic
// one and the anti-memetic side has; with the node it does, and a dark
// age that forgets the node forgets the people too.
func TestUnseen(t *testing.T) {
	w, c, x := antimemeticPair(t, 65, 1, 2)
	if !w.touch(c, x) && !w.hear(c, x) {
		t.Skip("not in touch")
	}
	w.ticks(5)
	if c.Met[x.ID] || c.Reached[x.ID] || c.Intel[x.ID] != nil {
		t.Fatal("the conscious side met what cannot be held in mind")
	}
	if !x.Met[c.ID] {
		t.Fatal("the anti-memetic side did not find them")
	}
	an := spawnAt(w, 3, fixedWith(species.Biological, species.Unconscious, nil, "defensive", "practical"))
	starfaring(w, an)
	if !w.perceives(an, x) {
		t.Fatal("nobody home, and still it cannot see")
	}
	w.learn(c, tech.Get(resilience), false)
	if !c.Met[x.ID] {
		t.Fatal("the node did not unveil them")
	}
	c.Known[resilience] = false
	delete(c.Known, resilience)
	w.veil(c)
	if c.Met[x.ID] || c.Fathomed[x.ID] {
		t.Fatal("the dark age did not forget them")
	}
}

// TestUnseenFleets: an anti-memetic people's fleets are never sighted by
// a conscious people without the node, and its message is not received.
func TestUnseenFleets(t *testing.T) {
	w, c, x := antimemeticPair(t, 62, 1, 2)
	w.addGuard(c, c.Home, 6)
	w.addGuard(x, x.Home, 20)
	x.Met[c.ID] = true
	if wr := w.declare(x, c, "a test"); wr == nil {
		t.Fatal("no war")
	}
	f := w.launch(x, Campaign, c, c.Home, 10)
	if f == nil {
		t.Fatal("no fleet")
	}
	for range 400 {
		w.tick()
		if f.Base >= 0 || f.Over {
			break
		}
	}
	if c.Tally.Sightings > 0 || len(c.Sightings) > 0 {
		t.Fatalf("the fleet was seen: %d sightings", c.Tally.Sightings)
	}
	if c.Met[x.ID] {
		t.Fatal("a war made the conscious side meet them")
	}
	if wr := w.warBetween(c.ID, x.ID); wr != nil && w.canTreat(wr) {
		t.Fatal("terms between what cannot hold each other in mind")
	}
}

// losses writes n doerless losses for c against x at stars around a
// centre, and returns the stars.
func losses(w *World, c, x *Civ, centre, n int) []int {
	var out []int
	for _, s := range append([]int{centre}, w.G.Near(centre, 8)...) {
		if len(out) == n {
			break
		}
		w.fact(FSurveyLost, c, x, s)
		out = append(out, s)
	}
	return out
}

// TestHunt: a hunt is declared after three doerless losses in one region
// and not after two; its strike lands on the anti-memetic people's
// garrison and is fought; the taking is a doerless tale.
func TestHunt(t *testing.T) {
	w, c, x := antimemeticPair(t, 63, 1, 2)
	w.addGuard(c, c.Home, 30)
	w.addGuard(x, x.Home, 3)
	losses(w, c, x, x.Home, 2)
	w.deduce(c)
	if wr := w.warBetween(c.ID, x.ID); wr != nil {
		t.Fatal("a hunt on two losses")
	}
	losses(w, c, x, x.Home, 3)
	w.deduce(c)
	wr := w.warBetween(c.ID, x.ID)
	if wr == nil || wr.Gap == nil || wr.Sides[0] != c.ID {
		t.Fatalf("no hunt: %v", wr)
	}
	if c.Tally.Hunts != 1 || !w.inGap(wr.Gap, x.Home) {
		t.Fatalf("hunts %d, the home inside the hole %v", c.Tally.Hunts, w.inGap(wr.Gap, x.Home))
	}
	if !w.hasFleetAgainst(c, x) {
		t.Fatal("no fleet sent into the region")
	}
	for range 600 {
		w.tick()
		if c.Tally.Battles > 0 || wr.Over {
			break
		}
	}
	if c.Tally.Battles == 0 {
		t.Fatalf("the hunt fought nothing: over %v (%s)", wr.Over, wr.Result)
	}
	if c.Met[x.ID] {
		t.Fatal("the hunt met them")
	}
	for _, tl := range c.Lore {
		f := w.Events[tl.Fact]
		if f.Kind == FTaken && f.Subject == c.ID && f.Object == x.ID {
			if line := w.tell(c, tl); !strings.Contains(line, "nameless") {
				t.Fatalf("the taking names them: %q", line)
			}
		}
		if f.Kind == FHunt && f.Subject == c.ID {
			if line := w.tell(c, tl); !strings.Contains(line, "hole") {
				t.Fatalf("the deduction: %q", line)
			}
		}
	}
}

// TestWarBecomesHunt: a dark age that forgets the node turns a war on an
// anti-memetic people into a hunt.
func TestWarBecomesHunt(t *testing.T) {
	w, c, x := antimemeticPair(t, 64, 1, 2)
	w.learn(c, tech.Get(resilience), false)
	c.Met[x.ID], x.Met[c.ID] = true, true
	wr := w.declare(c, x, "a test")
	if wr == nil || wr.Gap != nil {
		t.Fatalf("the war with the node: %v", wr)
	}
	delete(c.Known, resilience)
	w.veil(c)
	if wr.Gap == nil || c.Tally.Hunts != 1 || c.Met[x.ID] {
		t.Fatalf("the war did not become a hunt: gap %v, hunts %d, met %v", wr.Gap, c.Tally.Hunts, c.Met[x.ID])
	}
	w.learn(c, tech.Get(resilience), false)
	if wr.Gap != nil || !c.Met[x.ID] {
		t.Fatal("the node did not give the hunt its quarry back")
	}
}
