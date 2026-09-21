package history

import (
	"strings"
	"testing"

	"worldgen/internal/plague"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// The horrors dissolved: what was a ticker is a people or a remain now.

// starfaring gives a people the tree to the stars, for tests that need
// an era and a reach without a run.
func starfaring(w *World, c *Civ) {
	for _, k := range []string{"writing", "agriculture", "metallurgy", "mathematics", "printing", "scientific_method", "industry", "electricity", "atomic_power", "rocketry", "orbital_industry", "fusion", "interstellar_flight"} {
		if tech.Get(k) != nil {
			c.Known[k] = true
		}
	}
	w.recompute(c)
}

// TestMachinesDeclineMakesAPeople: Thinking Machines' decline pulls the
// plug or makes a machine people, and nothing else, every time.
func TestMachinesDeclineMakesAPeople(t *testing.T) {
	w := newTestWorld(t, 50, 120)
	for i := range 100 {
		star := i + 1
		if star == w.G.Sol {
			continue
		}
		c := spawnAt(w, star, species.Fixed("defensive", "practical", "curious"))
		starfaring(w, c)
		before := len(w.Civs)
		filters["machines"].Decline(w, c)
		switch {
		case c.DarkAges == 1 && c.Active():
		case !c.Active() && c.Fate == Transformed && len(w.Civs) == before+1:
			nc := w.Civs[before]
			if nc.Species.Sub != species.Machine || nc.Species.Made == "" || !nc.Active() || nc.Home != star {
				t.Fatalf("run %d: what the decline made is %s (%s), made %q, active %v at %d", i, nc.Tok(), nc.Species.Sub, nc.Species.Made, nc.Active(), nc.Home)
			}
		default:
			t.Fatalf("run %d: the decline left %s active %v, fate %v, dark ages %d, civs %d to %d", i, c.Tok(), c.Active(), c.Fate, c.DarkAges, before, len(w.Civs))
		}
	}
}

// TestReplicatorEats: a replicator's world yields ships and nothing
// else: no uses, no yards, and every ship it has was grown of what it
// ate.
func TestReplicatorEats(t *testing.T) {
	w := newTestWorld(t, 51, 40)
	c := spawnAt(w, 1, fixedWith(species.Machine, species.Replicator, nil, "defensive", "practical", "curious", "dormancy"))
	starfaring(w, c)
	if !c.Known["self_replication"] || !c.Faced["replication"] {
		t.Fatal("self-replication is not innate")
	}
	w.ticks(600)
	if !c.Active() {
		t.Fatalf("ended: %s", c.Cause)
	}
	if w.ships(c) == 0 || c.Tally.Eaten == 0 {
		t.Fatalf("nothing grown: ships %d, eaten %d", w.ships(c), c.Tally.Eaten)
	}
	if c.Tally.Eaten != c.Tally.Built || len(c.DockRate) > 0 || len(w.docks(c)) > 0 {
		t.Fatalf("ships from somewhere else: eaten %d, built %d, docks %d", c.Tally.Eaten, c.Tally.Built, len(w.docks(c)))
	}
	if c.Upkeep.Total() > 0 || len(w.uses(c)) > 0 {
		t.Fatalf("a replicator with uses: %v", c.Upkeep)
	}
	if c.Pursuit != "" && c.Progress > 0 && c.Species.Profile().Rate != 0.2 {
		t.Fatal("research at the full rate")
	}
}

// TestReplicatorNoTerms: a war with a replicator ends only by
// exhaustion: one will spent is nothing, both spent is the end, and
// nothing is signed.
func TestReplicatorNoTerms(t *testing.T) {
	w := newTestWorld(t, 52, 40)
	r := spawnAt(w, 0, fixedWith(species.Biological, species.Replicator, nil, "conqueror", "practical", "curious", "dormancy"))
	e := spawnAt(w, 1, species.Fixed("defensive", "faithful", "cautious"))
	starfaring(w, r)
	starfaring(w, e)
	r.Met[e.ID], e.Met[r.ID] = true, true
	r.Fathomed[e.ID], e.Fathomed[r.ID] = true, true
	wr := w.declare(e, r, "a test")
	if wr == nil {
		t.Fatal("no war")
	}
	wr.Will[0], wr.Will[1] = 0, 1
	w.judge(wr)
	if wr.Over {
		t.Fatalf("the war ended with one will spent: %s", wr.Result)
	}
	wr.Will[1] = 0
	w.judge(wr)
	if !wr.Over || wr.Result != "exhaustion" {
		t.Fatalf("over %v, result %q", wr.Over, wr.Result)
	}
	if !r.Asleep {
		t.Fatal("the replicator did not go quiet with its will spent")
	}
	if e.Tally.Capitulated || len(w.Pacts) > 0 {
		t.Fatal("terms were made")
	}
}

// TestSleeperWakes: a sleeper is asleep from the making, wakes on a
// people that settles inside its neighbourhood, and sleeps again when
// the war it woke to is spent.
func TestSleeperWakes(t *testing.T) {
	w := newTestWorld(t, 53, 40)
	s := w.sleeperAt(0, "a test")
	if s == nil || !s.Asleep || !s.Species.HasPower("sleep") || !s.Species.Is(species.Planetary) || !s.Species.Is(species.Unconscious) {
		t.Fatalf("the sleeper: %v", s)
	}
	near := -1
	for _, x := range w.G.Near(0, 12) {
		if x != w.G.Sol {
			near = x
			break
		}
	}
	if near < 0 {
		t.Skip("no star inside the neighbourhood")
	}
	e := spawnAt(w, 1, species.Fixed("defensive", "faithful", "cautious"))
	starfaring(w, e)
	w.ticks(3)
	if !s.Asleep {
		t.Fatal("woke with nobody inside")
	}
	w.holdWorld(e, near)
	w.afterHold(e, near)
	if s.Asleep || s.Tally.Wakings == 0 {
		t.Fatalf("asleep %v, wakings %d", s.Asleep, s.Tally.Wakings)
	}
	if !e.Faced["waking"] {
		t.Fatal("no waking")
	}
	wr := w.warBetween(s.ID, e.ID)
	if !e.Active() || wr == nil {
		return // the waking ended them, or the war with it; the sleeping again is TestReplicatorNoTerms's
	}
	wr.Will[0], wr.Will[1] = 0, 0
	w.judge(wr)
	if !wr.Over || !s.Asleep {
		t.Fatalf("over %v, asleep again %v", wr.Over, s.Asleep)
	}
}

// TestTransmitterPayloads: a corruption runs the Signal; a seed puts a
// conscious memetic plague in the listener, and a mind-rider wakes down
// it when the listener's home goes over.
func TestTransmitterPayloads(t *testing.T) {
	w := newTestWorld(t, 54, 40)
	l := w.makeTransmitter(0, -1, true)
	if !l.Transmitter() || !l.Speaking() {
		t.Fatal("not a live transmitter")
	}
	a := spawnAt(w, 1, species.Fixed("defensive", "faithful", "cautious"))
	b := spawnAt(w, 2, species.Fixed("defensive", "faithful", "cautious"))
	starfaring(w, a)
	starfaring(w, b)
	if !w.listens(a) {
		t.Fatalf("era %d does not listen", a.Era)
	}
	l.Payload = Corruption
	w.listen(a, l)
	if !a.Faced["beacon"] || !a.Heard[l.ID] {
		t.Fatal("the corruption did not run the Signal")
	}
	l.Payload = Seed
	w.listen(b, l)
	var p *Plague
	for _, x := range w.Infected(b) {
		if x.Transmitter == l.ID {
			p = x
		}
	}
	if p == nil || !p.Conscious || p.Kind != plague.Memetic || b.Faced["beacon"] {
		t.Fatalf("the seed: plague %v, faced the Signal %v", p, b.Faced["beacon"])
	}
	civs := len(w.Civs)
	w.worldLost(b, p, b.Home)
	if len(w.Civs) != civs+1 || l.Woken != 1 {
		t.Fatalf("nothing woke: civs %d to %d, woken %d", civs, len(w.Civs), l.Woken)
	}
	nc := w.Civs[civs]
	if nc.Species.Sub != species.Parasite || !nc.Has("mindrider") || !strings.Contains(nc.Origin, "came down the signal") || b.Master != nc.ID {
		t.Fatalf("what woke: %s, %s, origin %q, rides %v", nc.Tok(), nc.Species.Describe(), nc.Origin, b.Master == nc.ID)
	}
	an := spawnAt(w, 3, fixedWith(species.Biological, species.Unconscious, nil, "defensive", "practical"))
	starfaring(w, an)
	if w.listens(an) {
		t.Fatal("nobody home, and it listened")
	}
}

// TestHazardOfWhatIsThere: the hazard rises with the worlds held by
// what eats them and with the transmitters speaking.
func TestHazardOfWhatIsThere(t *testing.T) {
	w := newTestWorld(t, 55, 40)
	w.updateHazard()
	base := w.Hazard
	r := spawnAt(w, 0, fixedWith(species.Machine, species.Replicator, nil, "conqueror", "practical", "dormancy"))
	for s := 1; s < 4; s++ {
		w.Owner[s] = r.ID
		r.Systems = append(r.Systems, s)
	}
	w.updateHazard()
	eaten := w.Hazard
	if eaten <= base {
		t.Fatalf("hazard %g with four worlds eaten, %g with none", eaten, base)
	}
	w.makeTransmitter(5, -1, true)
	w.updateHazard()
	speaking := w.Hazard
	if speaking <= eaten {
		t.Fatalf("hazard %g with a transmitter speaking, %g without", speaking, eaten)
	}
	w.sleep(r)
	w.updateHazard()
	if w.Hazard >= speaking {
		t.Fatal("a sleeping replicator still counts")
	}
}
