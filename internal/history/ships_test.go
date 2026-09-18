package history

import (
	"testing"

	"worldgen/internal/battle"
	"worldgen/internal/flow"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// shipyardTick runs the flows and the shipwright by hand for a people
// whose income is set so that the spare beyond its ships' keep is the
// given amount: a people with a want, docks, and nothing else to feed.
func shipyardTick(w *World, c *Civ, spare flow.Income) {
	c.Reserved = flow.Income{}
	keep := w.keepOf(c).Scale(float64(w.ships(c)))
	c.Income = spare
	c.Income.Add(keep)
	w.direct(c, w.uses(c), false)
	w.shipwright(c)
	w.Now += w.Cfg.Step
}

// starfarer is a people with a first guard and no nodes to feed.
func starfarer(t *testing.T, seed uint64) (*World, *Civ) {
	t.Helper()
	w := newTestWorld(t, seed, 20)
	c := spawnAt(w, 0, species.Fixed("cooperative"))
	c.Known = map[string]bool{}
	w.Now = 1000 // the dawn is year zero, and a year of zero is never
	c.Starfaring = w.Now
	c.Dials.Fear = 0
	w.firstGuard(c)
	return w, c
}

// TestDocksBuild: a dock with 4 M 4 E spare builds one ship a thousand
// years, with 2 M 2 E one per two; a ship's build draws four times its
// keep in ship-kyr whatever the rate; a people at its want builds nothing.
func TestDocksBuild(t *testing.T) {
	w, c := starfarer(t, 41)
	for range 2000 {
		c.WantShips, c.WantSince = 1_000_000, w.Now
		shipyardTick(w, c, flow.Income{flow.O: 4, flow.E: 4, flow.M: 4})
	}
	if b := c.Tally.Built; b < 1900 || b > 2000 {
		t.Errorf("a full dock built %d ships in 2000 kyr, want about 2000", b)
	}
	if per := c.Tally.Building / float64(c.Tally.Built); per < 3.9 || per > 4.1 {
		t.Errorf("a ship cost %.2f ship-kyr of flow, want 4", per)
	}
	if w.ships(c) != 1+c.Tally.Built {
		t.Errorf("%d ships in being, want the first and %d built", w.ships(c), c.Tally.Built)
	}
	w, c = starfarer(t, 42)
	for range 2000 {
		c.WantShips, c.WantSince = 1_000_000, w.Now
		shipyardTick(w, c, flow.Income{flow.O: 4, flow.E: 2, flow.M: 2})
	}
	if b := c.Tally.Built; b < 900 || b > 1100 {
		t.Errorf("a half-fed dock built %d ships in 2000 kyr, want about 1000", b)
	}
	if per := c.Tally.Building / float64(c.Tally.Built); per < 3.7 || per > 4.3 {
		t.Errorf("a ship at half rate cost %.2f ship-kyr of flow, want 4", per)
	}
	w, c = starfarer(t, 43)
	for range 100 {
		shipyardTick(w, c, flow.Income{flow.O: 9, flow.E: 9, flow.M: 9})
	}
	if c.Tally.Built != 0 || len(c.DockRate) != 0 {
		t.Errorf("a people at its want of one built %d ships and works %d docks", c.Tally.Built, len(c.DockRate))
	}
	if c.Tally.AtWant != 100 {
		t.Errorf("at the want %d ticks of 100", c.Tally.AtWant)
	}
}

// TestLaidUp: a fleet the flow cannot keep is laid up: it does not fight,
// loses about one ship in ten a thousand years, and is manned again at
// the current level when the flow comes back.
func TestLaidUp(t *testing.T) {
	w, c := starfarer(t, 44)
	g := w.guardAt(c, c.Home)
	g.Ships = 1000
	c.Mil = 4
	shipyardTick(w, c, flow.Income{flow.O: -1000, flow.E: -1000, flow.M: -1000}) // nothing to keep them
	if !g.LaidUp || w.standing(c) != 0 || w.strikeForce(c) != 0 {
		t.Fatalf("a fleet with nothing to keep it is manned: laid up %v, standing %d", g.LaidUp, w.standing(c))
	}
	if d := w.defence(c, c.Home); d > 2*battle.Quality(c.Mil+3.5)+1e-9 {
		t.Errorf("a home with its fleet laid up is defended with %.2f, want the world alone", d)
	}
	if g.Ships < 860 || g.Ships > 940 {
		t.Errorf("a laid-up fleet of a thousand has %d after a thousand years, want about 900", g.Ships)
	}
	if c.Tally.LaidTick != 1 {
		t.Errorf("laid-up ticks %d, want 1", c.Tally.LaidTick)
	}
	w.learn(c, tech.Get("firearms"), false)
	shipyardTick(w, c, flow.Income{})
	if g.LaidUp || w.standing(c) != g.Ships {
		t.Errorf("a fleet fed again stays laid up: %v, standing %d of %d", g.LaidUp, w.standing(c), g.Ships)
	}
	if got, want := w.strikeForce(c), battle.Strength(g.Ships, w.quality(c)); got != want || w.quality(c) <= 1 {
		t.Errorf("manned again it fights at %.1f, want %.1f at the level today (quality %.2f)", got, want, w.quality(c))
	}
}

// TestQualityFollowsTheLevel: learning a weapons node raises the quality
// of the same ships, and the level is no longer a count of anything.
func TestQualityFollowsTheLevel(t *testing.T) {
	w, c := starfarer(t, 45)
	w.recompute(c)
	before, ships := w.quality(c), w.ships(c)
	w.learn(c, tech.Get("firearms"), false)
	w.recompute(c)
	if w.quality(c) <= before || w.ships(c) != ships {
		t.Errorf("firearms: quality %.2f from %.2f, ships %d from %d", w.quality(c), before, w.ships(c), ships)
	}
	if got := battle.Levels(w.quality(c) / before); got < 0.49 || got > 0.51 {
		t.Errorf("half a level of arms is %.2f levels of quality", got)
	}
	mil := c.Mil
	w.launch(c, Scout, c, 1, 1)
	w.recompute(c)
	if c.Mil != mil {
		t.Errorf("a ship out moved the level from %.2f to %.2f", mil, c.Mil)
	}
}

// TestHordeThins: a horde with no grazing cannot keep its ships, and thins.
func TestHordeThins(t *testing.T) {
	w, c := starfarer(t, 46)
	c.Species = species.Fixed("nomadic", "cooperative")
	w.guardAt(c, c.Home).Ships = 200
	if !w.takeSky(c, "") {
		t.Fatal("did not take to the sky")
	}
	if n := w.ships(c); n != 200 {
		t.Fatalf("the horde has %d ships, want the guard's 200", n)
	}
	for range 20 {
		shipyardTick(w, c, flow.Income{flow.O: -1000, flow.E: -1000, flow.M: -1000})
		w.roam(c)
	}
	if n := w.ships(c); n > 60 || n < 8 {
		t.Errorf("a horde with nothing to graze has %d ships after 20 kyr, want about 24", n)
	}
	if len(c.Record) == 0 || c.Record[len(c.Record)-1] != "took to the sky" {
		t.Errorf("record %v", c.Record)
	}
}

// TestFleetsMerge: a scout comes home into the guard, and a campaign's
// ships are taken out of it; a people cannot send what it has not got.
func TestFleetsMerge(t *testing.T) {
	w, c := starfarer(t, 47)
	e := spawnAt(w, 1, species.Fixed("cooperative"))
	w.guardAt(c, c.Home).Ships = 5
	if x := w.launch(c, Campaign, e, e.Home, 6); x != nil {
		t.Fatal("six ships sailed from a guard of five")
	}
	x := w.launch(c, Scout, e, e.Home, 2)
	if x == nil || w.guardAt(c, c.Home).Ships != 3 || w.ships(c) != 5 {
		t.Fatalf("the scout took %v; guard %d, in being %d", x != nil, w.guardAt(c, c.Home).Ships, w.ships(c))
	}
	w.goHome(x)
	w.Now = x.Arrive
	w.tickExpeditions()
	if !x.Over || w.guardAt(c, c.Home).Ships != 5 || len(w.fleetsOf(c)) != 1 {
		t.Errorf("the scout home: over %v, guard %d, fleets %d", x.Over, w.guardAt(c, c.Home).Ships, len(w.fleetsOf(c)))
	}
	// a war of fleets: the strike pays losses from the guards
	c.Mil, e.Mil = 6, 2
	w.addGuard(e, e.Home, 2)
	before := w.ships(e)
	wr := w.declare(c, e, "a test")
	w.strike(wr, c, e, e.Home)
	if w.ships(e) > before {
		t.Errorf("the defender gained ships in a strike")
	}
}
