package history

import (
	"math"
	"testing"

	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// led is a realm of n worlds that has reached the stars, with its
// leaders' occasions made certain, so a tick of its leaders' step raises
// one.
func led(t *testing.T, seed uint64, n int) (*World, *Civ) {
	t.Helper()
	w, c := realm(t, seed, n)
	c.Starfaring = w.Now
	c.DarkAges, c.LastDark = 1, w.Now // a dark age just past: the crisis occasion
	w.Cfg.Tuning.Leaders.Crisis = 1e6
	return w, c
}

// TestLeaderRises: an occasion and the chance make a leader, named by a
// fact with its occasion; the council reads its stance; and the draws
// are the leaders' own stream's, so the people's stream is untouched.
func TestLeaderRises(t *testing.T) {
	w, c := led(t, 31, 4)
	w2, c2 := led(t, 31, 4)
	w2.leaderStep(c2)
	if w.civStream(c).Uint64() != w2.civStream(c2).Uint64() {
		t.Fatal("raising a leader drew from the people's own stream")
	}
	l := c2.Leader
	if l == nil {
		t.Fatal("no leader rose on a certain occasion")
	}
	f := lastFacts(w2, 1)[0]
	if f.Kind != FLeader || f.P["leader"] != l.ID || f.P["occasion"] != "crisis" {
		t.Fatalf("the rising's fact: %s %v", f.Kind, f.P)
	}
	if c2.posture() != l.Stance || l.Own != "defensive" {
		t.Fatalf("the council reads %s; the leader's stance is %s and the people's own %s", c2.posture(), l.Stance, l.Own)
	}
	if l.Deathless || l.Until <= w2.Now {
		t.Fatalf("a mortal people's leader with no span: deathless %v, until %d", l.Deathless, l.Until)
	}
	if b := c2.leaderBent(); b != l.Bent {
		t.Fatalf("the dials' term is %+v, the leader's bent %+v", b, l.Bent)
	}
}

// TestLeaderTurnsAndSnapsBack: a leader whose stance is war turns a
// peaceful people to it, as the numbers the council reads; lost, the
// people is itself again and owes the succession, which is faced the
// next time it acts.
func TestLeaderTurnsAndSnapsBack(t *testing.T) {
	w, c := realm(t, 32, 4)
	l := &Leader{ID: 0, Civ: c.ID, Rose: w.Now, Occasion: "war", Form: "person", Stance: mind.Conqueror, Own: c.ownPosture(), Fleet: -1, Star: c.Home, Bent: Dials{Aggression: 0.45}, Push: 0.8}
	w.Leaders = append(w.Leaders, l)
	c.Leader = l
	w.recompute(c)
	if c.posture() != mind.Conqueror {
		t.Fatalf("under a conqueror the council reads %s", c.posture())
	}
	if _, wants, _ := mind.Bar(mind.BarInput{Posture: c.posture()}, w.Cfg.Tuning); !wants {
		t.Fatal("a defensive people under a conqueror wants no war")
	}
	w.loseLeader(c, "died", c.Home)
	if c.posture() != "defensive" || c.Leader != nil || c.Bereft != l {
		t.Fatalf("after the loss: posture %s, leader %v, bereft %v", c.posture(), c.Leader, c.Bereft)
	}
	snapped := false
	for _, e := range w.Events {
		snapped = snapped || (e.Kind == KSnappedBack && e.P["leader"] == l.ID)
	}
	if !snapped {
		t.Fatal("a people turned against its bent did not snap back")
	}
	w.leaderStep(c)
	if l.Faced == "" || !c.Faced["succession"] || c.Bereft != nil {
		t.Fatalf("the succession was not faced: %q", l.Faced)
	}
}

// TestSuccessionReadsContinuity: the succession is harder the less the
// realm keeps of its past, harder the further the leader pushed it, and
// easier with a backup copy.
func TestSuccessionReadsContinuity(t *testing.T) {
	w, c := realm(t, 33, 4)
	l := &Leader{Push: 0}
	base := w.successionAdj(c, l)
	c.DarkAges, c.LastDark = 1, w.Now // the cut of a dark age: much more is lost
	if dark := w.successionAdj(c, l); dark <= base+0.4 {
		t.Fatalf("a realm just out of a dark age: %.2f against %.2f", dark, base)
	}
	c.DarkAges, c.LastDark = 0, 0
	if pushed := w.successionAdj(c, &Leader{Push: 1}); math.Abs(pushed-base-w.Cfg.Tuning.Leaders.Push) > 1e-9 {
		t.Fatalf("a leader that turned its people: %.2f against %.2f", pushed, base)
	}
	c.Known["forking"] = true
	if forked := w.successionAdj(c, l); math.Abs(base-forked-w.Cfg.Tuning.Leaders.Forking) > 1e-9 {
		t.Fatalf("a backup copy: %.2f against %.2f", forked, base)
	}
}

// TestLeaderFallsWithItsPlace: a leader at the seat falls with the
// capital; one that runs runs once, diminished, and falls the next time;
// a people that falls or ends takes its leader with it and owes nothing.
func TestLeaderFallsWithItsPlace(t *testing.T) {
	w, c := led(t, 34, 5)
	w.leaderStep(c)
	l := c.Leader
	l.Front, l.Flees = false, false
	home := c.Home
	w.loseSystem(c, home, "taken", reason{})
	if c.Leader != nil || l.End != "fell_capital" || c.Bereft != l {
		t.Fatalf("the capital lost: leader %v, end %q", c.Leader, l.End)
	}
	c.Bereft = nil // the succession is another test's
	w.leaderStep(c)
	m := c.Leader
	if m == nil {
		t.Fatal("no second leader")
	}
	m.Front, m.Flees = false, true
	w.loseSystem(c, c.Home, "taken", reason{})
	if c.Leader != m || !m.Fled || c.leaderBent() == m.Bent && m.Bent != (Dials{}) {
		t.Fatalf("a leader that runs: reigns %v, fled %v", c.Leader == m, m.Fled)
	}
	if mil, soc := w.leaderLevels(c); mil != 0 || soc != 0 {
		t.Fatal("a leader that ran still lifts the realm")
	}
	w.loseSystem(c, c.Home, "taken", reason{})
	if c.Leader != nil || m.End != "fell_capital" {
		t.Fatalf("a leader that ran once falls the second time: end %q", m.End)
	}
	c.Bereft = nil
	w.leaderStep(c)
	n := c.Leader
	if n == nil {
		t.Fatal("no third leader")
	}
	w.contract(c, because("ossified"))
	if c.Leader != nil || n.End != "with_people" || c.Bereft != nil {
		t.Fatalf("the people fell under its leader: end %q, bereft %v", n.End, c.Bereft)
	}
}

// TestDeathlessSlide: a deathless leader narrows what its people counts
// as wrong, to conquest and then to nothing; mad, it makes war on
// everything and its purges eat the realm's continuity. A people that
// made death sacred has no deathless ruler.
func TestDeathlessSlide(t *testing.T) {
	w, c := led(t, 35, 4)
	m := species.GenerateWith(w.R, 1, "lush", species.Machine, 0)
	w.register(m)
	w.setSpecies(c, m)
	w.recompute(c)
	w.leaderStep(c)
	l := c.Leader
	if l == nil || !l.Deathless || l.Until != 0 {
		t.Fatalf("a machine people's leader: %+v", l)
	}
	before := w.continuityLoss(c)
	l.Mad = 0.6
	w.slide(c, l)
	if c.Morality != (Morality{Kind: Fixation, Object: Conquest}) {
		t.Fatalf("at a half the good is %v", c.Morality)
	}
	l.Mad = 1
	w.slide(c, l)
	if c.Morality.Kind != Amoral || c.posture() != mind.Conqueror {
		t.Fatalf("mad: %v, %s", c.Morality, c.posture())
	}
	if after := w.continuityLoss(c); after-before < w.Cfg.Tuning.Leaders.MadPurge-1e-9 {
		t.Fatalf("the purges took nothing: %.3f then %.3f", before, after)
	}
	w2, c2 := led(t, 35, 4)
	m2 := species.GenerateWith(w2.R, 1, "lush", species.Machine, 0)
	w2.register(m2)
	w2.setSpecies(c2, m2)
	c2.Scars[ScarMortality] = true
	w2.recompute(c2)
	w2.leaderStep(c2)
	if l := c2.Leader; l == nil || l.Deathless || l.Until == 0 {
		t.Fatalf("a people that made death sacred has a deathless ruler: %+v", l)
	}
}

// TestNobodyToLead: the unconscious and a living world raise no leaders.
func TestNobodyToLead(t *testing.T) {
	for _, mod := range []species.Mod{species.Unconscious, species.Planetary} {
		w, c := led(t, 36, 3)
		sp := species.GenerateWith(w.R, 1, "lush", species.Biological, mod)
		w.register(sp)
		w.setSpecies(c, sp)
		w.recompute(c)
		w.leaderStep(c)
		if c.Leader != nil {
			t.Fatalf("%v raised a leader", mod)
		}
	}
}
