package history

import (
	"math"
	"strings"
	"testing"

	"worldgen/internal/mind"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// realm raises a people holding n worlds at the first n stars, with the
// stars' reach and the tree an interstellar people has.
func realm(t *testing.T, seed uint64, n int) (*World, *Civ) {
	t.Helper()
	w := newTestWorld(t, seed, 40)
	w.Now = 100_000
	c := spawnAt(w, 0, species.Fixed("cooperative", "defensive", "practical"))
	for s := 1; s < n; s++ {
		w.Owner[s] = c.ID
		c.Systems = append(c.Systems, s)
	}
	for _, k := range []string{"writing", "agriculture", "metallurgy", "mathematics", "printing", "scientific_method", "industry", "electricity", "atomic_power", "rocketry", "orbital_industry", "fusion", "interstellar_flight"} {
		if tech.Get(k) != nil {
			c.Known[k] = true
		}
	}
	c.Peak = n
	w.recompute(c)
	return w, c
}

// TestStiffGrowth: the growth table, term by term.
func TestStiffGrowth(t *testing.T) {
	tn := &mind.Default().Ossify
	base := stiffInput{Fertility: 1, Nature: 1, Traits: 1}
	if g := stiffGrowth(base, tn); math.Abs(g-tn.Base) > 1e-12 {
		t.Fatalf("a lone cradle under a fresh sky: %g against %g", g, tn.Base)
	}
	cases := []struct {
		name string
		in   stiffInput
		mul  float64
	}{
		{"eight worlds", stiffInput{Worlds: 8, Fertility: 1, Nature: 1, Traits: 1}, 2},
		{"a dead sky", stiffInput{Fertility: 0, Nature: 1, Traits: 1}, 3},
		{"still", stiffInput{Fertility: 1, Still: true, Nature: 1, Traits: 1}, 1.5},
		{"ossified", stiffInput{Fertility: 1, Ossified: true, Nature: 1, Traits: 1}, 1.5},
		{"two iron answers", stiffInput{Fertility: 1, Iron: 2, Nature: 1, Traits: 1}, 1.25 * 1.25},
		{"a machine", stiffInput{Fertility: 1, Nature: 1.5, Traits: 1}, 1.5},
		{"longlived and shortlived", stiffInput{Fertility: 1, Nature: 1, Traits: 1.4 * 0.6}, 0.84},
	}
	for _, cs := range cases {
		if g := stiffGrowth(cs.in, tn); math.Abs(g/tn.Base-cs.mul) > 1e-9 {
			t.Errorf("%s: %g times the base, want %g", cs.name, g/tn.Base, cs.mul)
		}
	}
	// a mid-sized people of eight worlds under half fertility is at one after a million years
	mid := stiffGrowth(stiffInput{Worlds: 8, Fertility: 0.5, Nature: 1, Traits: 1}, tn)
	if years := 1 / mid * 1000; years < 800_000 || years > 1_200_000 {
		t.Errorf("eight worlds under half fertility reach one in %.0f years", years)
	}
	sp := species.Fixed("longlived", "caste")
	c := &Civ{Species: sp}
	if m := c.traitStiff(); math.Abs(m-1.4*1.3) > 1e-9 {
		t.Errorf("the trait table: %g", m)
	}
}

// TestLoweringHooks: what is new takes stiffness off: a first meeting, a
// war fought to its end, a world lost, a people made, an art mastered
// from a find; a miracle and a dark age reset it.
func TestLoweringHooks(t *testing.T) {
	w, c := realm(t, 30, 3)
	e := spawnAt(w, 20, species.Fixed("cooperative", "defensive"))
	for k := range c.Known {
		e.Known[k] = true
	}
	w.recompute(e)
	for _, x := range []*Civ{c, e} {
		x.Reach, x.Speed, x.Stiff = 60, 20, 1
	}
	w.meet(c, e, -1)
	if c.Stiff != 0.9 || e.Stiff != 0.9 {
		t.Errorf("a first meeting: %g and %g", c.Stiff, e.Stiff)
	}
	wr := w.declare(c, e, "a test")
	if c.Still != w.Now {
		t.Error("a war is not still")
	}
	w.endWar(wr, "peace")
	if math.Abs(c.Stiff-0.8) > 1e-9 {
		t.Errorf("a war fought to peace: %g", c.Stiff)
	}
	w.loseSystem(c, 2, "abandoned colony", "")
	if math.Abs(c.Stiff-0.75) > 1e-9 {
		t.Errorf("a world lost: %g", c.Stiff)
	}
	sp := species.Fixed("cooperative")
	sp.Made = "made"
	w.spawnCiv(30, sp, c.ID, "")
	if math.Abs(c.Stiff-0.55) > 1e-9 {
		t.Errorf("a people made: %g", c.Stiff)
	}
	w.gain(c, "ansible", "leap")
	if c.Stiff != 0 || c.Still != w.Now {
		t.Errorf("a miracle gained: %g", c.Stiff)
	}
	c.Stiff, c.Ossified = 2, true
	w.darkAge(c, "fell")
	if c.Stiff != 0 || c.Ossified {
		t.Error("a dark age did not reset the institutions")
	}
}

// TestOffTicks: an ossified people banks no research and launches nothing
// on its off ticks, and banks half on its on ticks.
func TestOffTicks(t *testing.T) {
	w, c := realm(t, 31, 4)
	c.Pursuit = ""
	w.Now = 1_000_000 + Year(c.ID%2)*1000 // an off tick for this people once ossified
	c.Ossified = true
	if !w.offTick(c) {
		w.Now += 1000
	}
	if !w.offTick(c) {
		t.Fatal("no off tick found")
	}
	before, voyages := c.Progress, len(c.Voyages)
	w.tickCivs()
	if c.Progress != before || len(c.Voyages) != voyages {
		t.Errorf("an off tick banked research (%g to %g) or launched (%d to %d)", before, c.Progress, voyages, len(c.Voyages))
	}
	w.Now += 1000
	if w.offTick(c) {
		t.Fatal("two off ticks running")
	}
	c.Ossified = false
	w.recompute(c)
	full := w.researchRate(c)
	c.Ossified = true
	w.recompute(c)
	if half := w.researchRate(c); half/full > 0.5 || half/full < 0.45 { // a set people's society is half a level lower too
		t.Errorf("an on tick banks %g of the rate, want a half", half/full)
	}
}

// TestSecondNearMiss: a people set once breaks at the next near miss;
// renaissance, civil war and dark age all clear the state.
func TestSecondNearMiss(t *testing.T) {
	w, c := realm(t, 32, 1)
	f := filters["ossification"]
	f.Scar(w, c)
	if !c.Ossified {
		t.Fatal("the near miss did not set the people")
	}
	f.Scar(w, c)
	if c.Ossified || c.DarkAges != 1 {
		t.Errorf("a second near miss on one world: ossified %v, dark ages %d", c.Ossified, c.DarkAges)
	}
	c.Ossified = true
	f.Overcome(w, c)
	if c.Ossified || c.Renaissances != 1 || c.Stiff != 0 {
		t.Error("a renaissance did not clear the state")
	}
	w, c = realm(t, 33, 6)
	c.Ossified = true
	if !w.civilWar(c) {
		t.Fatal("six worlds did not make a civil war")
	}
	for _, h := range w.Civs[1:] {
		if h.Ossified || h.Stiff != 0 {
			t.Error("an heir is born ossified")
		}
	}
}

// TestStiffFaces: a people at stiffness three faces within a few hundred
// thousand years, and the difficulty carries its stiffness.
func TestStiffFaces(t *testing.T) {
	w, c := realm(t, 34, 2)
	c.Stiff = 3
	for i := 0; i < 500 && c.Tally.OssFaced == 0; i++ {
		w.tickStiff(c)
		w.Now += 1000
	}
	if c.Tally.OssFaced == 0 {
		t.Fatal("stiffness three was not faced in half a million years")
	}
	c.Stiff, c.Renaissances = 2, 2
	if _, d, _ := filters["ossification"].Adjust(w, c); d != 2+2*w.Cfg.Tuning.Ossify.Renaissance {
		t.Errorf("the difficulty's adjustment: %g", d)
	}
}

// TestDarkDepth: the depth formula's bounds, and the deepening.
func TestDarkDepth(t *testing.T) {
	w, c := realm(t, 35, 1)
	for range 200 {
		if d := w.darkDepth(c); d < 0.1 || d > 0.3 {
			t.Fatalf("a fresh people's depth %g", d)
		}
	}
	c.Stiff, c.DarkAges = 3, 4
	for range 200 {
		if d := w.darkDepth(c); d < 0.7 || d > 0.8 {
			t.Fatalf("Trantor's depth %g", d)
		}
	}
	c.DarkAges = 20
	if d := w.darkDepth(c); d != 0.8 {
		t.Errorf("the cap: %g", d)
	}
}

// TestNoAgeDeaths: a small world with the filter forced every tick records
// no ends by age: nobody contracts and nobody dies of the filter.
func TestNoAgeDeaths(t *testing.T) {
	w, c := realm(t, 36, 5)
	w.Cfg.Tuning.Ossify.Chance = 1
	for range 400 {
		for _, x := range w.Civs {
			if x.Active() {
				x.Stiff = max(x.Stiff, 2)
			}
		}
		w.tick()
	}
	for _, x := range w.Civs {
		if x.Fate == Contracted || strings.Contains(x.Cause, "hardened") || strings.Contains(x.Cause, "weight") {
			t.Errorf("the %s ended %s: %s", x.Name, x.Fate, x.Cause)
		}
	}
	if c.Tally.OssFaced == 0 {
		t.Error("the filter was never faced")
	}
}

// TestForgive: grudges decay and clear; a grudge above the enemy line
// suspends kinship's warmth and its decay restores it.
func TestForgive(t *testing.T) {
	w, c := realm(t, 37, 6)
	if !w.civilWar(c) {
		t.Fatal("no civil war")
	}
	a, b := w.Civs[1], w.Civs[2]
	if !w.kin(a, b) || w.regard(a, b.ID) != -1 {
		t.Fatalf("heirs: kin %v, regard %d", w.kin(a, b), w.regard(a, b.ID))
	}
	if wr := w.warBetween(a.ID, b.ID); wr != nil {
		w.endWar(wr, "peace")
	}
	a.Grudge[b.ID], b.Grudge[a.ID] = 3, 3
	if w.regard(a, b.ID) != -1 {
		t.Error("a grudge above the line does not suspend the warmth")
	}
	for range 400 {
		w.forgive(a)
		w.forgive(b)
	}
	if g := a.Grudge[b.ID]; g > 0.5 {
		t.Fatalf("after 400 kyr the grudge is %g", g)
	}
	if w.regard(a, b.ID) != 1 {
		t.Errorf("the warmth did not come back: regard %d, grudge %g", w.regard(a, b.ID), a.Grudge[b.ID])
	}
	for range 600 {
		w.forgive(a)
	}
	if _, ok := a.Grudge[b.ID]; ok {
		t.Error("the grudge did not clear")
	}
}
