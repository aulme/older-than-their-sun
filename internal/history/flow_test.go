package history

import (
	"testing"

	"worldgen/internal/flow"
	"worldgen/internal/species"
)

// starve leaves a people with no sources at its home, so every use with an
// upkeep goes dark.
func starve(w *World, c *Civ) {
	w.sourcesAt[c.Home] = nil
}

// TestMachinePaysOrganicInEnergy: a machine people's costs in organic
// matter are costs in energy.
func TestMachinePaysOrganicInEnergy(t *testing.T) {
	w := newTestWorld(t, 7, 30)
	m := spawnAt(w, 0, species.GenerateWith(w.R, 1, "lush", species.Machine, 0))
	b := spawnAt(w, 1, species.Fixed("cooperative"))
	for _, c := range []*Civ{m, b} {
		c.Known["medicine"] = true // biology, era 1: 1 O
	}
	need := func(c *Civ) flow.Income {
		for _, u := range w.uses(c) {
			if u.Key == "medicine" {
				return u.Need
			}
		}
		t.Fatalf("%s has no use for medicine", c.Tok())
		return flow.Income{}
	}
	if got := need(m); got != (flow.Income{flow.E: 1}) {
		t.Errorf("the machine's medicine costs %v, want 1 E", got)
	}
	if got := need(b); got != (flow.Income{flow.O: 1}) {
		t.Errorf("the biological's medicine costs %v, want 1 O", got)
	}
	// and the machine has no fields: its medicine is fed with the works
	for _, u := range w.uses(m) {
		if u.Key == "medicine" && u.Cat != flow.Works {
			t.Errorf("the machine's medicine is fed under %s, want the works", u.Cat)
		}
	}
	if w.order(m).Order.Has(flow.Fields) {
		t.Error("a machine's order names the fields")
	}
}

// TestPlanetaryCradleYieldsDouble: the world a planetary mind arose on
// feeds it twice what it would feed anyone else.
func TestPlanetaryCradleYieldsDouble(t *testing.T) {
	w := newTestWorld(t, 8, 30)
	star := -1
	for i := range w.G.Stars {
		if w.G.Sys[i].Home >= 0 {
			star = i
			break
		}
	}
	if star < 0 {
		t.Skip("no habitable world in the field")
	}
	base := worldYield[w.G.Sys[star].Arch]
	p := spawnAt(w, star, species.GenerateWith(w.R, 1, w.G.Sys[star].Arch, species.Biological, species.Planetary))
	if got := w.yieldAt(p, star)[flow.O]; got != 2*base {
		t.Errorf("the planetary cradle yields %v O, want %v", got, 2*base)
	}
	w.Owner[star] = -1
	b := spawnAt(w, star, species.Fixed("cooperative"))
	if got := w.yieldAt(b, star)[flow.O]; got != base {
		t.Errorf("the plain cradle yields %v O, want %v", got, base)
	}
	// a colonist there, not born to it, must know how to farm
	c := spawnAt(w, (star+1)%len(w.G.Stars), species.Fixed("cooperative"))
	delete(c.Known, "agriculture")
	if got := w.yieldAt(c, star)[flow.O]; got != 0 {
		t.Errorf("a stranger's fields yield %v O before agriculture", got)
	}
	c.Known["agriculture"] = true
	if got := w.yieldAt(c, star)[flow.O]; got != base {
		t.Errorf("a stranger's fields yield %v O with agriculture, want %v", got, base)
	}
}

// TestDormantEnvelope: a node dormant ten ticks stops widening the
// envelope; one dormant nine does not.
func TestDormantEnvelope(t *testing.T) {
	w := newTestWorld(t, 9, 30)
	c := spawnAt(w, 0, species.Fixed("cooperative"))
	starve(w, c)
	c.Known["closed_ecologies"] = true // Env 1, 1 O 1 E
	w.recompute(c)
	fed := c.Envelope
	w.ticks(9)
	if !c.Shed["closed_ecologies"] {
		t.Fatal("closed ecologies are fed with no income")
	}
	w.recompute(c)
	if c.Envelope != fed {
		t.Errorf("after nine dark ticks the envelope is %d, was %d", c.Envelope, fed)
	}
	w.tick()
	w.recompute(c)
	if c.Envelope != fed-1 {
		t.Errorf("after ten dark ticks the envelope is %d, want %d", c.Envelope, fed-1)
	}
}

// TestLeanYearsOnce: a stretch of shedding past a hundred thousand years
// is one fact, and one only, however long it goes on.
func TestLeanYearsOnce(t *testing.T) {
	w := newTestWorld(t, 10, 30)
	c := spawnAt(w, 0, species.Fixed("cooperative"))
	starve(w, c)
	c.Known["medicine"] = true
	w.ticks(99)
	wants := func() int {
		n := 0
		for _, f := range w.Events {
			if f.Kind == FWant && f.Subject == c.ID {
				n++
			}
		}
		return n
	}
	if len(c.Shed) == 0 || wants() != 0 {
		t.Fatalf("after 99 ticks: shedding %v, %d lean-years facts", c.Shed, wants())
	}
	w.ticks(101)
	if wants() != 1 {
		t.Fatalf("after 200 ticks of shedding: %d lean-years facts, want 1", wants())
	}
	// a first shed is one line, and only one for the stretch
	n := 0
	for _, e := range w.Events {
		if e.Kind == KWentDark {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d lines of going dark over one stretch, want 1", n)
	}
}
