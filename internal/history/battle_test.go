package history

import (
	"math"
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
	w.declare(c, e, because("border"))
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
	w.muster(c, e, e.Home, because("border"), 3)
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
	w.muster(c, e, e.Home, because("border"), 3)
	w.guardAt(e, e.Home).Ships = 30
	c.Intel[e.ID].Ships = 30
	w.musterStep(c)
	if c.Muster != nil {
		t.Errorf("the odds went and the muster stands")
	}
	// a muster gathered for a war that has ended stands down, and declares
	// no new one when its ships arrive (step 11's batches, seed 12: a muster
	// outliving each war it was called in opened the next, twenty-nine times)
	w, c, e, colony = twoPeoples(t, 57)
	w.guardAt(c, c.Home).Ships = 2
	w.Owner[colony] = c.ID
	e.Systems = e.Systems[:1]
	c.Systems = append(c.Systems, colony)
	w.addGuard(c, colony, 2)
	w.addGuard(e, e.Home, 3)
	c.Intel[e.ID] = &Intel{Mil: 6, Ships: 3, Total: 3, Star: e.Home, Year: w.Now, Sick: -1}
	wr := w.declare(c, e, because("border"))
	w.muster(c, e, e.Home, because("border"), 3)
	if c.Muster == nil || c.Muster.War != wr.ID {
		t.Fatalf("the muster does not know its war: %+v", c.Muster)
	}
	w.endWar(wr, "peace")
	c.Truce[e.ID], e.Truce[c.ID] = 0, 0
	w.guardAt(c, colony).Ships = 3
	w.musterStep(c)
	if c.Muster != nil || w.warBetween(c.ID, e.ID) != nil {
		t.Errorf("a muster outlived its war: muster %+v, a new war %v", c.Muster, w.warBetween(c.ID, e.ID) != nil)
	}
}

// TestIdleDrains: a tick nobody fights in costs will; a tick with a
// battle, or with a fleet on its way, costs nothing; a home within the
// enemy's reach halves the cost.
func TestIdleDrains(t *testing.T) {
	w, c, e, colony := twoPeoples(t, 58)
	idle := w.Cfg.Tuning.War.Idle
	e.Reach = 0 // the declarer's home is safe
	wr := w.declare(c, e, because("border"))
	wr.Will = [2]float64{5, 5}
	w.drain(wr, 0)
	if d := 5 - wr.Will[0]; math.Abs(d-idle) > 1e-9 {
		t.Errorf("an idle tick drained %.3f, want %.3f", d, idle)
	}
	wr.Will = [2]float64{5, 5}
	w.drain(wr, 1) // the side declared on: its home is in the declarer's reach
	if d := 5 - wr.Will[1]; math.Abs(d-idle*w.Cfg.Tuning.War.HomeResolve) > 1e-9 {
		t.Errorf("an idle tick with the home threatened drained %.3f, want %.3f", d, idle*w.Cfg.Tuning.War.HomeResolve)
	}
	wr.Battles, wr.Fought = 1, w.Now
	wr.Will = [2]float64{5, 5}
	w.drain(wr, 0)
	if wr.Will[0] != 5 {
		t.Errorf("a tick with a battle drained %.3f", 5-wr.Will[0])
	}
	wr.Fought = w.Now - 5000
	w.guardAt(c, c.Home).Ships += 1
	if x := w.launch(c, Campaign, e, colony, 1); x == nil {
		t.Fatal("no fleet")
	}
	w.drain(wr, 0)
	if wr.Will[0] != 5 {
		t.Errorf("a tick with a fleet on its way drained %.3f", 5-wr.Will[0])
	}
}

// TestGoNativeOnce: a fleet that took a world, lost it and took it again
// holds it once, and the people it becomes holds it once (seed 37 of
// step 11's batches: a lost fleet's people held a star seven times over,
// and its heirs shattered onto it twice, one of them left with nothing).
func TestGoNativeOnce(t *testing.T) {
	w, c, e, colony := twoPeoples(t, 73)
	e.Systems = remove(e.Systems, colony)
	w.setOwner(colony, c.ID)
	c.Systems = append(c.Systems, colony)
	x := w.launch(c, Campaign, e, e.Home, 1)
	if x == nil {
		t.Fatal("no fleet")
	}
	other := -1
	for s := range w.G.Stars {
		if w.Owner[s] < 0 && s != colony {
			other = s
			break
		}
	}
	w.setOwner(other, c.ID)
	c.Systems = append(c.Systems, other)
	x.Held, x.Base = []int{other, other, other, colony}, colony // the last held is the new home
	w.goNative(x)
	nc := w.Civs[len(w.Civs)-1]
	if nc == c || nc == e {
		t.Fatal("no people of the lost fleet")
	}
	n := 0
	for _, s := range nc.Systems {
		if s == other {
			n++
		}
	}
	if n != 1 {
		t.Errorf("the lost fleet's people holds a world %d times: %v", n, nc.Systems)
	}
}
