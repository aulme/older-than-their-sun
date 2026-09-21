package history

import (
	"math/rand/v2"
	"testing"

	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// TestRollMorality: the weights. A hive all but never rolls individual; a
// pacifist never rolls amoral; a conqueror leans to the taking of worlds.
func TestRollMorality(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 7))
	hive := species.Fixed("defensive")
	hive.Mods = species.Hive
	const n = 20000
	ind := 0
	for range n {
		if rollMorality(r, hive, noHints).Kind == Individual {
			ind++
		}
	}
	if float64(ind)/n > 0.01 {
		t.Errorf("a hive rolled individual %d of %d", ind, n)
	}
	pac := species.Fixed("pacifist", "individualist")
	for range n {
		if m := rollMorality(r, pac, noHints); m.Kind == Amoral {
			t.Fatal("a pacifist rolled amoral")
		}
	}
	conq := species.Fixed("conqueror", "individualist")
	fix, taking := 0, 0
	for range n {
		if m := rollMorality(r, conq, noHints); m.Kind == Fixation {
			fix++
			if m.Object == Conquest {
				taking++
			}
		}
	}
	if float64(taking)/float64(fix) < 0.35 {
		t.Errorf("a conqueror's fixations were on conquest %d of %d", taking, fix)
	}
	h := noHints
	h.FixationMul = 3
	h.Object = Knowing
	fix = 0
	for range n {
		if m := rollMorality(r, species.Fixed("defensive"), h); m.Kind == Fixation && m.Object == Knowing {
			fix++
		}
	}
	if float64(fix)/n < 0.4 {
		t.Errorf("a machine successor leaning to knowing got there %d of %d", fix, n)
	}
}

// TestSortFor: the table. An enslavement is a crime of four to an
// individual people, a deed of two to a conquest people, nothing to an
// amoral bystander and a woe to the amoral sufferer; a fact not in the
// table keeps its own sort.
func TestSortFor(t *testing.T) {
	w := newTestWorld(t, 31, 20)
	m := spawnAt(w, 0, species.Fixed("conqueror"))
	s := spawnAt(w, 1, species.Fixed("defensive"))
	o := spawnAt(w, 2, species.Fixed("defensive"))
	f := w.fact(FEnslaved, m, s, 1)
	judge := func(c *Civ, mo Morality) (Sort, float64) {
		c.Morality = mo
		return sortFor(c, f)
	}
	if s, wt := judge(o, Morality{Kind: Individual}); s != Crime || wt != 4 {
		t.Errorf("individual: %v %g", s, wt)
	}
	if s, wt := judge(o, Morality{Kind: Fixation, Object: Conquest}); s != Deed || wt != 2 {
		t.Errorf("conquest: %v %g", s, wt)
	}
	if s, wt := judge(o, Morality{Kind: Amoral}); s != Nothing || wt != 0 {
		t.Errorf("amoral bystander: %v %g", s, wt)
	}
	if s, wt := judge(s, Morality{Kind: Amoral}); s != Woe || wt != 4 {
		t.Errorf("amoral sufferer: %v %g", s, wt)
	}
	if s, wt := judge(o, Morality{Kind: Herd}); s != Crime || wt != 1 {
		t.Errorf("herd: %v %g", s, wt)
	}
	var arise *Event
	for _, e := range w.Events {
		if e.Kind == FArise {
			arise = e
			break
		}
	}
	if s, wt := judge(o, Morality{Kind: Amoral}); s != Nothing {
		t.Errorf("amoral on the enslavement again: %v %g", s, wt)
	}
	if s, wt := sortFor(o, arise); s != Deed || wt != 3 {
		t.Errorf("a fact off the table: %v %g", s, wt)
	}
}

// TestAmoralHoldsNoMonsters: the reckoning sums crimes, and an amoral
// people has none to sum. The same tales make a monster for an individual
// people.
func TestAmoralHoldsNoMonsters(t *testing.T) {
	w := newTestWorld(t, 32, 20)
	a := spawnAt(w, 0, species.Fixed("defensive"))
	e := spawnAt(w, 1, species.Fixed("conqueror"))
	a.Met[e.ID], e.Met[a.ID] = true, true
	for i := range 4 {
		w.fact(FBurned, e, a, 2+i)
	}
	a.Morality = Morality{Kind: Amoral}
	w.reckon(a)
	if len(a.monsters) != 0 {
		t.Errorf("an amoral people holds monsters: %v", a.monsters)
	}
	a.Morality = Morality{Kind: Individual}
	w.reckon(a)
	if !a.monsters[e.ID] {
		t.Errorf("an individual people burned four times holds no monster")
	}
	if !judges(a, lastFacts(w, 1)[0]) {
		a.Morality = Morality{Kind: Amoral}
		if !judges(a, lastFacts(w, 1)[0]) {
			t.Error("an amoral people's judgment of a burning is the fact's own")
		}
	}
	// a small world runs with every morality in it
	for _, m := range []Morality{{Kind: Amoral}, {Kind: Herd}, {Kind: Fixation, Object: Conquest}, {Kind: Fixation, Object: Holding}} {
		c := spawnAt(w, 5+int(m.Kind)+len(m.Object), species.Fixed("defensive"))
		c.Morality = m
	}
	w.ticks(50)
}

// TestHoldingSendsNothing: a people fixed on holding refuses every
// partner; its partners, once it has closed its ports, refuse it back.
func TestHoldingSendsNothing(t *testing.T) {
	if c := mind.Trade(mind.TradeInput{Holding: true, InReach: true, Drive: 2}, mind.Default()); !c.Refuse {
		t.Errorf("a holding people sends: %s", c.Why())
	}
	if c := mind.Trade(mind.TradeInput{Embargoed: true, InReach: true, Drive: 2}, mind.Default()); !c.Refuse {
		t.Errorf("a people sends to one that closed its ports to it: %s", c.Why())
	}
	if c := mind.Trade(mind.TradeInput{Spawning: true, InReach: true}, mind.Default()); c.CapOf(flow.O) != 0.5 || c.CapOf(flow.M) != 0.25 {
		t.Errorf("a spawning people's caps: O %g M %g", c.CapOf(flow.O), c.CapOf(flow.M))
	}
	if c := mind.Trade(mind.TradeInput{Grasping: true, InReach: true}, mind.Default()); c.Mul != [3]float64{2, 2, 2} {
		t.Errorf("a grasping partner's want: %v", c.Mul)
	}
	w, a, b := partners(t, 33)
	a.Morality = Morality{Kind: Fixation, Object: Holding}
	a.Surplus, a.Reserved = flow.Income{flow.M: 8}, flow.Income{}
	w.trade()
	if b.From[a.ID] != (flow.Income{}) {
		t.Errorf("a holding people sent %v", b.From[a.ID])
	}
	if _, ok := a.Refused[b.ID]; !ok {
		t.Error("the refusal was not written down")
	}
}

// TestMoralityDifference: morality adds to how alien two peoples are: a
// whole point between individual and herd, half to or from an amoral one,
// nothing between the same.
func TestMoralityDifference(t *testing.T) {
	w := newTestWorld(t, 34, 20)
	a := spawnAt(w, 0, species.Fixed("individualist", "defensive"))
	b := spawnAt(w, 1, species.Fixed("individualist", "defensive"))
	base := difference(a.Species, b.Species)
	a.Morality, b.Morality = Morality{Kind: Individual}, Morality{Kind: Herd}
	if d := a.differs(b); d != base+1 {
		t.Errorf("individual against herd: %g, want %g", d, base+1)
	}
	b.Morality = Morality{Kind: Amoral}
	if d := a.differs(b); d != base+0.5 {
		t.Errorf("against amoral: %g", d)
	}
	a.Morality, b.Morality = Morality{Kind: Fixation, Object: Knowing}, Morality{Kind: Fixation, Object: Holding}
	if d := a.differs(b); d != base+0.5 {
		t.Errorf("two fixations: %g", d)
	}
	b.Morality = a.Morality
	if d := a.differs(b); d != base {
		t.Errorf("the same fixation: %g", d)
	}
}

// TestFixationDirection: a conquest people feeds arms before everything;
// a holding people the works; a nomad fixed on holding has no works to
// feed and keeps its order.
func TestFixationDirection(t *testing.T) {
	tn := mind.Default()
	d := mind.Direct(mind.DirectionInput{Fixed: true, Fixation: flow.Arms}, tn)
	if d.Order[0] != flow.Arms || !d.FixFirst {
		t.Errorf("conquest: %v", d.Order)
	}
	d = mind.Direct(mind.DirectionInput{Fixed: true, Fixation: flow.Works, AtWar: true}, tn)
	if d.Order[0] != flow.Works || d.Order[1] != flow.Arms {
		t.Errorf("holding at war: %v", d.Order)
	}
	d = mind.Direct(mind.DirectionInput{Fixed: true, Fixation: flow.Works, Nomad: true}, tn)
	if d.FixFirst || d.Order[0] != flow.Road {
		t.Errorf("a nomad fixed on holding: %v", d.Order)
	}
}
