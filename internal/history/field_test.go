package history

import "testing"

// fieldsAt is the ships in a people's fields at a star, and how many
// fields there are.
func fieldsAt(w *World, c *Civ, star int) (ships, fields int) {
	for _, l := range w.Legacies {
		if l.Kind == Field && l.Maker == c.ID && l.Star == star && !l.Adrift {
			fields++
			ships += l.Wrecks + l.Derelicts
		}
	}
	return
}

// TestFieldsFromBattle: every ship lost in a battle at a world is in a
// field at that star, wrecks and derelicts, and a second battle there by
// the same loser grows the field instead of leaving another.
func TestFieldsFromBattle(t *testing.T) {
	both := false
	for seed := uint64(71); seed < 90 && !both; seed++ {
		w, c, e, colony := twoPeoples(t, seed)
		c.Known["slow_interstellar"], e.Known["slow_interstellar"] = true, true
		w.addGuard(e, colony, 6)
		x := campaignAt(w, c, e, colony, 6)
		lostC, lostE := 0, 0
		for i := 0; i < 6 && !x.Over && x.Base == colony && w.Owner[colony] == e.ID; i++ {
			w.fight(x, colony)
			w.Now += w.Cfg.Step
			lostC, lostE = c.Tally.ShipsLost, e.Tally.ShipsLost
			sc, fc := fieldsAt(w, c, colony)
			se, fe := fieldsAt(w, e, colony)
			if sc != lostC || se != lostE {
				t.Fatalf("seed %d: lost %d and %d, in fields %d and %d", seed, lostC, lostE, sc, se)
			}
			if fc > 1 || fe > 1 {
				t.Fatalf("seed %d: %d and %d fields at one star by one loser", seed, fc, fe)
			}
		}
		if lostC > 0 && lostE > 0 {
			both = true
		}
		for _, l := range w.Legacies {
			if l.Kind != Field {
				continue
			}
			if l.Wrecks+l.Derelicts == 0 || l.Node == "" || l.State != Buried || l.Hardy != hardyLiving {
				t.Errorf("seed %d: field %+v", seed, l)
			}
			if l.Derelicts > 0 && l.Cond != Derelict || l.Derelicts == 0 && l.Cond != Wreck {
				t.Errorf("seed %d: %d derelicts and condition %s", seed, l.Derelicts, l.Cond)
			}
		}
	}
	if !both {
		t.Fatal("no seed had both sides lose ships in a siege")
	}
}

// TestAdriftFound: surveyors whose line passes a third of a light year
// from a field adrift come upon it; two light years off, they do not.
func TestAdriftFound(t *testing.T) {
	w, c, e := watcher(t, 72)
	e.Known["slow_interstellar"] = true
	l := w.leaveField(e, 5, e.Home, vec{20, 0, 0}, true)
	near := flight(w, c, Survey, -1, c.Home, e.Home, vec{0, 0.3, 0}, vec{40, 0.3, 0}, 1)
	w.timetable(near)
	if !c.Found[l.ID] {
		t.Error("a line a third of a light year off a field adrift did not find it")
	}
	w, c, e = watcher(t, 73)
	e.Known["slow_interstellar"] = true
	l = w.leaveField(e, 5, e.Home, vec{20, 0, 0}, true)
	far := flight(w, c, Survey, -1, c.Home, e.Home, vec{0, 2, 0}, vec{40, 2, 0}, 1)
	w.timetable(far)
	if c.Found[l.ID] {
		t.Error("a line two light years off a field adrift found it")
	}
	if l.Hardy != hardyAdrift || !l.Adrift || l.Star != e.Home {
		t.Errorf("field adrift: %+v", l)
	}
}

// TestRuinNoShips: a field worn to ruin holds no ships: it is nobody's
// find when its art is known, and crewing it yields nothing.
func TestRuinNoShips(t *testing.T) {
	w, c, e := watcher(t, 74)
	e.Known["slow_interstellar"], c.Known["slow_interstellar"] = true, true
	l := w.leaveField(e, 5, c.Home, w.pos(c.Home), false)
	if l.ships() != 5 {
		t.Fatalf("a fresh field of five holds %d", l.ships())
	}
	l.Cond = Ruin
	if l.ships() != 0 {
		t.Errorf("a field at ruin holds %d ships", l.ships())
	}
	c.Dials.Fear = 0
	for range 200 {
		w.find(c)
	}
	if c.Found[l.ID] {
		t.Error("a ruined field of a known art was a find")
	}
	before := len(w.Expeditions)
	w.salvage(c, l)
	if len(w.Expeditions) != before || c.Salvage != 0 {
		t.Error("crewing a ruin launched ships")
	}
}

// TestSalvage: a crewed field flies its derelicts and a quarter of its
// wrecks home as salvage, which wears to nothing in about ten thousand
// years.
func TestSalvage(t *testing.T) {
	w, c, e := watcher(t, 75)
	e.Known["slow_interstellar"] = true
	l := &Legacy{ID: len(w.Legacies), Age: -1, Maker: e.ID, Kind: Field, Star: c.Home, Node: "slow_interstellar", Horror: -1, Finder: -1, Source: -1, Plague: -1, Wrecks: 8, Derelicts: 2, Cond: Derelict, Hardy: hardyLiving}
	w.Legacies = append(w.Legacies, l)
	w.salvage(c, l)
	if c.Salvage != 4 || l.Cond != Ruin {
		t.Fatalf("salvage %d, condition %s", c.Salvage, l.Cond)
	}
	x := w.Expeditions[len(w.Expeditions)-1]
	if x.Kind != Guard || !x.Returning || x.Ships != 4 || x.Star != c.Home {
		t.Fatalf("the salvage fleet: %+v", x)
	}
	w.Now = x.Arrive
	w.tickExpeditions()
	if g := w.guardAt(c, c.Home); g == nil || g.Ships != 5 {
		t.Fatalf("the salvage did not land in the guard: %+v", g)
	}
	for i := 0; i < 40 && c.Salvage > 0; i++ {
		w.salvageWorn(c)
		if i == 2 && c.Salvage == 4 {
			t.Error("nothing of the salvage was lost in three thousand years")
		}
	}
	if c.Salvage != 0 || c.SalvageTaken != 0 {
		t.Errorf("salvage %d after forty thousand years", c.Salvage)
	}
	if g := w.guardAt(c, c.Home); g == nil || g.Ships != 1 {
		t.Errorf("the guard after the salvage wore away: %+v", g)
	}
}
