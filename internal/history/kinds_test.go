package history

import (
	"testing"

	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// fixedWith is a known species of a substrate and modifiers with the
// traits named, for tests: no roll, and the powers given.
func fixedWith(sub species.Substrate, mods species.Mod, powers []string, traits ...string) *species.Species {
	s := species.Fixed(traits...)
	s.Sub, s.Mods = sub, mods
	for _, p := range powers {
		s.AddPower(p)
	}
	return s
}

// runSmall runs a small world for n ticks with the peoples given at the
// first stars, and everything else born as it comes.
func runSmall(t *testing.T, seed uint64, stars, ticks int, sps ...*species.Species) (*World, []*Civ) {
	t.Helper()
	w := newTestWorld(t, seed, stars)
	var cs []*Civ
	for i, sp := range sps {
		c := spawnAt(w, i, sp)
		c.Reach, c.Speed = 40, 20
		for _, k := range []string{"writing", "agriculture", "metallurgy", "mathematics", "printing", "scientific_method", "industry", "electricity", "atomic_power", "rocketry", "orbital_industry", "fusion", "interstellar_flight"} {
			if !sp.Profile().Can(species.Researches) {
				break
			}
			if tech.Get(k) != nil {
				c.Known[k] = true
			}
		}
		w.recompute(c)
		cs = append(cs, c)
	}
	for range ticks {
		w.tick()
	}
	return w, cs
}

// TestHiveNeverStiffens: a hive has no institutions to set; its stiffness
// stays at zero through a run, and it never faces Ossification.
func TestHiveNeverStiffens(t *testing.T) {
	w, cs := runSmall(t, 40, 40, 1500, fixedWith(species.Biological, species.Hive, nil, "defensive", "practical", "curious", "noqueen"))
	c := cs[0]
	for _, x := range w.Civs {
		if x.Species == c.Species && (x.Stiff > 0 || x.Tally.OssFaced > 0 || x.Ossified) {
			t.Fatalf("the hive %s: stiffness %g, faced %d", x.Tok(), x.Stiff, x.Tally.OssFaced)
		}
	}
}

// TestUnconsciousHoldsNothing: an unconscious people ends a run with an
// empty grudge map and a morale of zero, whatever was done to it, and is
// amoral.
func TestUnconsciousHoldsNothing(t *testing.T) {
	w := newTestWorld(t, 41, 40)
	u := spawnAt(w, 0, fixedWith(species.Biological, species.Unconscious, nil, "conqueror", "faithless", "expansionist"))
	o := spawnAt(w, 1, fixedWith(species.Biological, 0, nil, "vengeful", "faithful", "xenophobic"))
	if w.face(u, "faith", 0) != Overcome || u.Faced["faith"] {
		t.Fatal("the unconscious faced the Wars of Faith")
	}
	if difference(u.Species, o.Species) < 2 {
		t.Fatal("no one home is not a large difference")
	}
	u.resent(o.ID, 3)
	u.Morale -= 2
	for range 1500 {
		w.tick()
	}
	for _, x := range w.Civs {
		if x.Species != u.Species {
			continue
		}
		if len(x.Grudge) != 0 || x.Morale != 0 || x.Morality.Kind != Amoral {
			t.Fatalf("the unconscious %s: grudges %v, morale %g, morality %s", x.Tok(), x.Grudge, x.Morale, x.Morality.Word())
		}
	}
}

// TestPlanetaryStaysHome: a living world never launches an expedition,
// never holds more than the cap, and has a fixed reach.
func TestPlanetaryStaysHome(t *testing.T) {
	w, cs := runSmall(t, 42, 40, 2000, fixedWith(species.Biological, species.Planetary, nil, "conqueror", "practical", "expansionist"))
	c := cs[0]
	cap := c.Species.Profile().Worlds
	for _, x := range w.Civs {
		if x.Species != c.Species {
			continue
		}
		if len(w.fleetsOf(x)) > 0 || x.Tally.Fleets > 0 || x.Tally.Scouts > 0 || x.Tally.Surveys > 0 {
			t.Fatalf("the world %s launched: %d fleets, %d/%d/%d", x.Tok(), len(w.fleetsOf(x)), x.Tally.Fleets, x.Tally.Scouts, x.Tally.Surveys)
		}
		if x.Peak > cap {
			t.Fatalf("the world held %d, the cap is %d", x.Peak, cap)
		}
		if x.Active() && x.Reach != c.Species.Profile().Neighbourhood {
			t.Fatalf("the world's reach is %g", x.Reach)
		}
	}
	if w.bodyGuns(c) == 0 {
		t.Fatal("the body stands no guns")
	}
}

// TestWaking: a conscious living world with a people inside its
// neighbourhood demands first, and wakes when the demand has stood
// refused; the waking is faced, not fought.
func TestWaking(t *testing.T) {
	w := newTestWorld(t, 43, 40)
	g := spawnAt(w, 0, fixedWith(species.Biological, species.Planetary, nil, "conqueror", "practical", "xenophobic"))
	g.Mil = 6
	near := -1
	for _, s := range w.G.Near(0, 12) {
		if s != w.G.Sol {
			near = s
			break
		}
	}
	if near < 0 {
		t.Skip("no star inside the neighbourhood")
	}
	e := spawnAt(w, near, species.Fixed("unyielding", "faithful", "cautious"))
	e.Reach = 20
	for _, s := range w.G.Near(near, 15) {
		if w.Owner[s] < 0 && s != w.G.Sol && s != 0 {
			w.Owner[s] = e.ID
			e.Systems = append(e.Systems, s)
			break
		}
	}
	w.recompute(g)
	w.recompute(e)
	g.Met[e.ID], e.Met[g.ID] = true, true
	g.Reached[e.ID], e.Reached[g.ID] = true, true
	worlds := w.inside(g, e)
	if len(worlds) == 0 {
		t.Fatal("nothing inside the neighbourhood")
	}
	if w.presenceStrike(g, e, worlds) {
		t.Fatal("a war on the first word")
	}
	if g.Tally.Demands != 1 || len(g.demanded) != 1 && e.Active() && len(e.Systems) > 1 {
		t.Fatalf("demands %d, standing %v", g.Tally.Demands, g.demanded)
	}
	if _, ok := g.demanded[e.ID]; !ok {
		return // they left; the demand did its work
	}
	w.Now += Year(w.Cfg.Tuning.Kinds.Demand+1) * w.Cfg.Step
	fleets := len(w.Expeditions)
	if !w.presenceStrike(g, e, worlds) {
		t.Fatal("the refused demand did not wake the world")
	}
	if g.Tally.Wakings != 1 || len(w.Expeditions) != fleets || g.Tally.Declared != 1 {
		t.Fatalf("wakings %d, fleets %d to %d, wars declared %d", g.Tally.Wakings, fleets, len(w.Expeditions), g.Tally.Declared)
	}
	if !e.Faced["waking"] {
		t.Fatal("the waking was not faced")
	}
}

// TestEldritchLives: an eldritch people never has a colony ship in flight
// and never researches, holds a power from birth and gains more over a
// long run; it reads as interstellar from the first.
func TestEldritchLives(t *testing.T) {
	sp := fixedWith(species.Eldritch, 0, []string{"presence"}, "defensive", "practical", "curious", "timesight")
	w, cs := runSmall(t, 44, 40, 8000, sp)
	c := cs[0]
	for _, x := range w.Civs {
		if x.Species != c.Species {
			continue
		}
		if len(x.Voyages) > 0 || x.Tally.Blind > 0 || x.Pursuit != "" || x.Progress > 0 {
			t.Fatalf("the eldritch %s: voyages %d, pursuit %q, progress %g", x.Tok(), len(x.Voyages), x.Pursuit, x.Progress)
		}
		for k := range x.Known {
			if p := powerNode(k); p == "" {
				t.Fatalf("the eldritch know %s", k)
			}
		}
	}
	if c.Era < 3 {
		t.Fatalf("an eldritch thing reads as era %d", c.Era)
	}
	if len(c.Species.Powers) < 2 {
		t.Fatalf("no deepening in eight million years: %v", c.Species.Powers)
	}
	if c.Tally.Deepened == 0 {
		t.Fatal("the deepening was not tallied")
	}
}

// powerNode is the power a miracle node stands for, or "".
func powerNode(node string) string {
	for _, p := range species.Pool {
		if p.Node == node {
			return p.Key
		}
	}
	return ""
}

// TestAppear: with a second presence another of it is there within a
// long run, and nothing crossed.
func TestAppear(t *testing.T) {
	sp := fixedWith(species.Eldritch, 0, []string{"presence", "hunger"}, "defensive", "practical", "expansionist", "masssense")
	w, cs := runSmall(t, 45, 40, 12000, sp)
	c := cs[0]
	if c.Tally.Appeared == 0 {
		t.Fatalf("no second presence in twelve million years: %d worlds", len(c.Systems))
	}
	if len(w.fleetsOf(c)) > 0 {
		t.Fatal("something crossed")
	}
}

// TestCutOff: a hive world too far from the seat becomes a people of its
// own, kin to the old, holding the world and what stood on it.
func TestCutOff(t *testing.T) {
	w := newTestWorld(t, 46, 40)
	c := spawnAt(w, 0, fixedWith(species.Biological, species.Hive, nil, "defensive", "practical", "noqueen"))
	for s := 1; s < 4; s++ {
		w.Owner[s] = c.ID
		c.Systems = append(c.Systems, s)
	}
	w.addGuns(c, 3, 2)
	w.recompute(c)
	before := len(w.Civs)
	h := w.cutOff(c, 3)
	if h == nil || len(w.Civs) != before+1 {
		t.Fatal("no people was cut off")
	}
	if w.Owner[3] != h.ID || contains(c.Systems, 3) || len(h.Systems) != 1 || h.Guns[3] != 2 {
		t.Fatalf("the world went wrong: owner %d, old holds %v, new holds %v, guns %v", w.Owner[3], c.Systems, h.Systems, h.Guns)
	}
	if !w.kin(c, h) || h.Species != c.Species || !c.Active() {
		t.Fatal("the cut people is not kin, or the old one did not go on")
	}
	if w.civilWar(c) {
		t.Fatal("a hive made a civil war")
	}
	// the Distance's decline for a hive is the cut
	filters["distance"].Decline(w, c)
	if len(w.Civs) != before+2 || !c.Active() {
		t.Fatalf("the Distance did %s to the hive; %d peoples", c.Fate, len(w.Civs))
	}
}

// TestSeatRules: a hive of one queen dies with its home; a hive of no
// queen shatters when the seat is lost; a world cannot reseat.
func TestSeatRules(t *testing.T) {
	w := newTestWorld(t, 47, 40)
	q := spawnAt(w, 0, fixedWith(species.Biological, species.Hive, nil, "defensive", "practical", "onequeen"))
	w.Owner[1] = q.ID
	q.Systems = append(q.Systems, 1)
	w.loseSystem(q, 0, "test", "lost the seat")
	if q.Active() {
		t.Fatal("a hive of one queen lived on without her")
	}
	n := spawnAt(w, 2, fixedWith(species.Biological, species.Hive, nil, "defensive", "practical", "noqueen"))
	for _, s := range []int{3, 4} {
		w.Owner[s] = n.ID
		n.Systems = append(n.Systems, s)
	}
	w.recompute(n)
	w.loseSystem(n, 2, "test", "lost the seat")
	if n.Fate != Shattered {
		t.Fatalf("a hive of no queen %s", n.Fate)
	}
	p := spawnAt(w, 5, fixedWith(species.Biological, species.Planetary, nil, "defensive", "practical"))
	w.Owner[6] = p.ID
	p.Systems = append(p.Systems, 6)
	w.loseSystem(p, 5, "test", "lost the world")
	if p.Active() {
		t.Fatal("a world lived on without itself")
	}
}

// TestMachineBackups: a machine people's dark age forgets half as much,
// and its shards wake with the whole tree.
func TestMachineBackups(t *testing.T) {
	w := newTestWorld(t, 48, 40)
	m := spawnAt(w, 0, fixedWith(species.Machine, 0, nil, "defensive", "practical", "curious"))
	for _, k := range []string{"writing", "agriculture", "metallurgy", "mathematics", "printing", "scientific_method", "industry", "electricity", "atomic_power", "rocketry", "orbital_industry", "fusion", "interstellar_flight"} {
		if tech.Get(k) != nil {
			m.Known[k] = true
		}
	}
	for s := 1; s < 4; s++ {
		w.Owner[s] = m.ID
		m.Systems = append(m.Systems, s)
	}
	w.recompute(m)
	known := len(m.Known)
	m.Stiff = 3
	m.DarkAges = 4 // the depth at its cap, halved
	w.darkAge(m, "fell")
	heirs := 0
	for _, x := range w.Civs[1:] {
		heirs++
		if len(x.Known) != known {
			t.Fatalf("a shard woke with %d of %d nodes", len(x.Known), known)
		}
	}
	if m.Fate != Shattered && len(m.Known) < known-int(float64(known)*0.45) {
		t.Fatalf("a machine people forgot %d of %d", known-len(m.Known), known)
	}
}
