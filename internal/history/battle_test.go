package history

import (
	"testing"

	"worldgen/internal/species"
)

// twoPeoples is a starfarer and an enemy with a colony within a hop of
// its home, both at a fixed level, the attacker knowing the enemy's sky.
func twoPeoples(t *testing.T, seed uint64) (w *World, c, e *Civ, colony int) {
	t.Helper()
	w, c = starfarer(t, seed)
	e = spawnAt(w, 1, species.Fixed("cooperative"))
	e.Known = map[string]bool{}
	e.Starfaring = w.Now
	for _, s := range w.G.Near(e.Home, fleetHop) {
		if w.Owner[s] < 0 && s != c.Home {
			colony = s
			break
		}
	}
	if colony == 0 {
		t.Fatal("no free star within a hop of the enemy's home")
	}
	w.Owner[colony] = e.ID
	e.Systems = append(e.Systems, colony)
	c.Speed, e.Speed = 100, 100
	c.Reach, e.Reach = 30, 30
	c.Mil, e.Mil = 6, 6
	return
}

// guns puts a fed gun structure at a star with its guns standing.
func guns(w *World, c *Civ, key string, star int) {
	w.raise(c, key, "defence_grid", star)
}

// campaignAt is a campaign fleet of n ships of c at an enemy's world,
// with the war declared.
func campaignAt(w *World, c, e *Civ, star, n int) *Expedition {
	w.guardAt(c, c.Home).Ships += n
	w.declare(c, e, "a test")
	x := w.launch(c, Campaign, e, star, n)
	x.Base = star
	return x
}

// TestEmptySky: a world with nothing in its sky is taken without a
// battle: nobody loses a ship.
func TestEmptySky(t *testing.T) {
	w, c, e, colony := twoPeoples(t, 51)
	x := campaignAt(w, c, e, colony, 2)
	w.fight(x, colony)
	if len(w.Battles) != 1 || w.Battles[0].Outcome != "empty" {
		t.Fatalf("battles %d: %+v", len(w.Battles), w.Battles)
	}
	if w.Owner[colony] == e.ID {
		t.Errorf("the world with an empty sky is still the enemy's")
	}
	if c.Tally.ShipsLost != 0 || e.Tally.ShipsLost != 0 || x.Ships != 2 {
		t.Errorf("ships were lost taking an empty sky: %d and %d, fleet %d", c.Tally.ShipsLost, e.Tally.ShipsLost, x.Ships)
	}
	if c.Tally.EmptySky != 1 {
		t.Errorf("empty skies %d", c.Tally.EmptySky)
	}
}

// TestGunsAbsorb: the defender's losses fall on its guns first, and a
// guard that lost nothing stays where it is. A world with a gun standing
// is not taken.
func TestGunsAbsorb(t *testing.T) {
	w, c, e, colony := twoPeoples(t, 52)
	guns(w, e, "defences", colony)
	guns(w, e, "silos", colony)
	if g := w.gunsAt(e, colony); g != 5 {
		t.Fatalf("a grid and silos stand %d guns, want 5", g)
	}
	w.addGuard(e, colony, 2)
	c.Mil = 4 // two ships at four levels: a loss of at most 2.44, well within five guns
	x := campaignAt(w, c, e, colony, 2)
	for i := 0; i < 3 && !x.Over && x.Base == colony; i++ {
		w.fight(x, colony)
		w.Now += w.Cfg.Step
	}
	g := w.guardAt(e, colony)
	if g == nil || g.Ships != 2 || g.Base != colony {
		t.Errorf("the guard behind five guns lost ships or moved: %+v", g)
	}
	if w.gunsAt(e, colony) < 2 {
		t.Errorf("two ships shot away %d of five guns", 5-w.gunsAt(e, colony))
	}
	if w.Owner[colony] != e.ID {
		t.Errorf("a world with a gun standing was taken")
	}
	for _, b := range w.Battles {
		if b.Outcome == "taken" || b.Outcome == "withdrew" {
			t.Errorf("outcome %s with the guns standing", b.Outcome)
		}
	}
	// guns alone, and a gun still standing: not taken, whoever won the roll
	w, c, e, colony = twoPeoples(t, 53)
	guns(w, e, "defences", colony)
	guns(w, e, "silos", colony)
	x = campaignAt(w, c, e, colony, 1)
	w.fight(x, colony)
	if w.Owner[colony] != e.ID || w.gunsAt(e, colony) < 3 {
		t.Errorf("one ship against five guns: owner %d, guns %d", w.Owner[colony], w.gunsAt(e, colony))
	}
}

// TestGuardWithdraws: a guard that lost ships to a fleet that won the roll
// withdraws to the nearest own world within a hop, and the world, with no
// gun over it, is taken as it leaves.
func TestGuardWithdraws(t *testing.T) {
	// the dice: the first seed whose roll the attacker wins with the guard
	// losing some ships and not all
	for seed := uint64(54); seed < 80; seed++ {
		w, c, e, colony := twoPeoples(t, seed)
		g := w.addGuard(e, colony, 6)
		c.Mil = 8 // three ships at eight levels: about 17 against 6, a loss of at most 8 ships on the guard
		x := campaignAt(w, c, e, colony, 3)
		w.fight(x, colony)
		b := w.Battles[0]
		if !b.Won || g.Ships == 0 || g.Ships == 6 {
			continue
		}
		if g.Base != -1 || !g.Returning || g.Star != e.Home {
			t.Errorf("the guard did not withdraw to the home a hop away: base %d, returning %v, bound for %d", g.Base, g.Returning, g.Star)
		}
		if b.Outcome != "taken" || w.Owner[colony] == e.ID {
			t.Errorf("outcome %s, owner %d: the world should be taken as the guard leaves", b.Outcome, w.Owner[colony])
		}
		// and it lands in the guard at home
		w.Now = g.Arrive
		w.tickExpeditions()
		if home := w.guardAt(e, e.Home); home == nil || home.Ships != g.Ships || !g.Over {
			t.Errorf("the withdrawn guard did not land at home")
		}
		return
	}
	t.Fatal("no seed in the range gave a won roll with a guard that lost some")
}

// TestGunsRepair: a broken grid is remade a gun a thousand years while the
// world is held, through a siege; silos are not re-dug while an enemy
// fleet is in the sky, and are after.
func TestGunsRepair(t *testing.T) {
	w, c, e, colony := twoPeoples(t, 55)
	guns(w, e, "defences", colony)
	e.Guns[colony] = 0
	e.GridBroken = map[int]bool{colony: true}
	x := campaignAt(w, c, e, colony, 1) // a siege
	for i := 1; i <= 3; i++ {
		w.guns(e)
		if g := w.gunsAt(e, colony); g != i {
			t.Errorf("after %d thousand years the grid has %d guns, want %d", i, g, i)
		}
	}
	w.guns(e)
	if g := w.gunsAt(e, colony); g != 3 || e.GridBroken[colony] {
		t.Errorf("a whole grid has %d guns, broken %v", g, e.GridBroken[colony])
	}
	guns(w, e, "silos", colony)
	e.Guns[colony] = 3 // the silos fired
	w.guns(e)
	if g := w.gunsAt(e, colony); g != 3 {
		t.Errorf("silos re-dug under siege: %d guns", g)
	}
	x.Over = true
	w.guns(e)
	w.guns(e)
	if g := w.gunsAt(e, colony); g != 5 {
		t.Errorf("silos not re-dug after the siege: %d guns, want 5", g)
	}
	w.guns(e)
	if g := w.gunsAt(e, colony); g != 5 {
		t.Errorf("guns beyond what stands there: %d", g)
	}
}

// TestMuster: a campaign no guard can man alone gathers at the holding
// nearest the target and sails at ship speed when the ships are there;
// one whose odds no longer hold stands down into the guard.
func TestMuster(t *testing.T) {
	w, c, e, colony := twoPeoples(t, 57)
	w.guardAt(c, c.Home).Ships = 2
	w.addGuard(c, colony, 0) // c holds a colony too, a hop from the enemy
	w.Owner[colony] = c.ID
	e.Systems = e.Systems[:1]
	c.Systems = append(c.Systems, colony)
	w.addGuard(c, colony, 2)
	w.addGuard(e, e.Home, 3)
	c.Intel[e.ID] = &Intel{Mil: 6, Ships: 3, Total: 3, Star: e.Home, Year: w.Now, Sick: -1}
	k := w.sizeAt(c, e, e.Home)
	if !k.Send || k.Share != 3 {
		t.Fatalf("sizing: %s", k.Why())
	}
	if w.launch(c, Campaign, e, e.Home, 3) != nil {
		t.Fatal("three ships sailed from guards of two and two")
	}
	w.muster(c, e, e.Home, "a test", 3)
	m := c.Muster
	if m == nil || m.Star != colony || m.Ships != 3 || c.Tally.Musters != 1 {
		t.Fatalf("muster %+v", m)
	}
	var coming *Expedition
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Guard && x.Base < 0 {
			coming = x
		}
	}
	if coming == nil || coming.Star != colony || coming.Ships != 1 || coming.Arrive <= w.Now { // the shortfall, not the whole guard
		t.Fatalf("nothing gathers: %+v", coming)
	}
	for w.Now < coming.Arrive {
		w.musterStep(c)
		if c.Muster == nil {
			t.Fatal("the muster ended before its ships arrived")
		}
		w.tickExpeditions()
		w.Now += w.Cfg.Step
	}
	w.tickExpeditions()
	w.musterStep(c)
	if c.Muster != nil {
		t.Fatalf("the ships arrived and the muster stands: %+v", c.Muster)
	}
	sailed := false
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Campaign && x.Ships == 3 && x.From == colony {
			sailed = true
		}
	}
	if !sailed {
		t.Errorf("the campaign did not sail from the muster")
	}
	// a muster whose odds go stands down
	w, c, e, colony = twoPeoples(t, 57)
	w.guardAt(c, c.Home).Ships = 2
	w.Owner[colony] = c.ID
	e.Systems = e.Systems[:1]
	c.Systems = append(c.Systems, colony)
	w.addGuard(c, colony, 2)
	w.addGuard(e, e.Home, 3)
	c.Intel[e.ID] = &Intel{Mil: 6, Ships: 3, Total: 3, Star: e.Home, Year: w.Now, Sick: -1}
	w.muster(c, e, e.Home, "a test", 3)
	w.guardAt(e, e.Home).Ships = 30
	c.Intel[e.ID].Ships = 30
	w.musterStep(c)
	if c.Muster != nil {
		t.Errorf("the odds went and the muster stands")
	}
}

// TestNoFleetDrainsDouble: a war with nobody's fleet against the other
// drains will twice as fast; one with a fleet at a base half as fast.
func TestNoFleetDrainsDouble(t *testing.T) {
	w, c, e, colony := twoPeoples(t, 58)
	wr := w.declare(c, e, "a test")
	wr.Will = [2]float64{5, 5}
	w.drain(wr, 0)
	if d := 5 - wr.Will[0]; d < 0.0999 || d > 0.1001 {
		t.Errorf("a war with no fleet drained %.3f, want 0.1: twice the base", d)
	}
	x := campaignAt(w, c, e, colony, 1)
	wr.Will = [2]float64{5, 5}
	w.drain(wr, 0)
	if d := 5 - wr.Will[0]; d < 0.0249 || d > 0.0251 {
		t.Errorf("a war with a fleet at a base drained %.3f, want 0.025", d)
	}
	x.Over = true
	c.Muster = &Muster{Target: e.ID}
	wr.Will = [2]float64{5, 5}
	w.drain(wr, 0)
	if d := 5 - wr.Will[0]; d < 0.0249 || d > 0.0251 {
		t.Errorf("a war with a muster gathering drained %.3f, want 0.025", d)
	}
}
