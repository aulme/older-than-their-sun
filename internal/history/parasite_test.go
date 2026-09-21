package history

import (
	"math"
	"testing"

	"worldgen/internal/plague"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// woken is a conscious plague put in a people by hand and woken into a
// rider at once: the host ridden, the rider with no world of its own.
func woken(t *testing.T, seed uint64, kind plague.Kind) (w *World, host, rider *Civ, p *Plague) {
	t.Helper()
	w, host, _, p = sickPair(t, seed, kind, 0.8, 0.5)
	p.Conscious = true
	rider = w.wake(p, host)
	return
}

// TestWake: a conscious plague whose first host's home goes over becomes
// a people riding that host; one cured before that spawns nothing and
// never thinks again.
func TestWake(t *testing.T) {
	w, a, _, p := sickPair(t, 20, plague.Biological, 0.8, 1)
	p.Conscious = true
	before := len(w.Civs)
	w.worldLost(a, p, a.Home)
	if len(w.Civs) != before+1 {
		t.Fatal("the home going over did not wake the plague")
	}
	rider := w.Civs[before]
	if rider.Own != p.ID || p.Rider != rider.ID || rider.Species.Sub != species.Parasite || !rider.Has("bodyrider") {
		t.Fatalf("the rider: own %d, plague's rider %d, %s", rider.Own, p.Rider, rider.Species.Describe())
	}
	if a.Master != rider.ID || a.Vassal || !w.ridden(a) || len(rider.Systems) != 0 || w.Owner[a.Home] != a.ID {
		t.Errorf("the host is not ridden as it should be: master %d, rider's worlds %v, owner %d", a.Master, rider.Systems, w.Owner[a.Home])
	}
	if !w.rides(rider) || len(w.hostsOf(rider)) != 1 {
		t.Error("the rider has no host to be in")
	}
	written := false
	for _, f := range w.Facts {
		written = written || (f.Kind == FEnslaved && f.Subject == rider.ID && f.Object == a.ID)
	}
	if !written {
		t.Error("the riding was not written")
	}
	// cured first: nothing wakes, and it never thinks again
	w, a, _, p = sickPair(t, 21, plague.Biological, 0.8, 1)
	p.Conscious = true
	before = len(w.Civs)
	w.cure(a, p)
	if len(w.Civs) != before || p.Conscious || p.Rider >= 0 {
		t.Error("a plague cured before its host's home went over woke, or still thinks")
	}
}

// TestBornRider: one cradle in a hundred is born with a rider in it, its
// home already over; the host starts ridden.
func TestBornRider(t *testing.T) {
	w := newTestWorld(t, 22, 30)
	w.Cfg.Tuning.Plague.BornRider = 1
	w.Now = 100_000
	host := spawnAt(w, 0, nil)
	if !w.ridden(host) {
		t.Fatal("a cradle born with a rider does not start ridden")
	}
	rider := w.Civs[host.Master]
	if rider.Own < 0 || w.Plagues[rider.Own].FirstHost != host.ID || host.Infections[rider.Own] == nil || rider.Home != host.Home {
		t.Errorf("the born rider: own %d, host's infections %v, home %d against %d", rider.Own, host.Infections, rider.Home, host.Home)
	}
	if w.Plagues[rider.Own].Cause != "born rider" {
		t.Error("the birth is not a born rider's for the batch")
	}
}

// TestParasiteNoPlague: a parasite bears no plague and catches none on
// any channel.
func TestParasiteNoPlague(t *testing.T) {
	w, host, rider, _ := woken(t, 23, plague.Biological)
	w.Cfg.Tuning.Plague.BaseBio, w.Cfg.Tuning.Plague.BaseMeme = 1, 1
	rider.Known["writing"] = true
	for range 20 {
		w.bearPlague(rider)
	}
	if len(rider.Infections) != 0 {
		t.Fatal("a parasite bore a plague")
	}
	q := w.newPlague(plague.Biological, host, "clean")
	q.Contagion = 1
	w.infect(host, q, nil, "born")
	for _, road := range []string{"goods", "occupation", "landing", "fleet"} {
		w.offer(host, rider, q, road)
	}
	w.Reservoir[7] = &Reservoir{Plague: q.ID, Until: w.Now + 1e6}
	w.wakeReservoir(rider, 7)
	if rider.Infections[q.ID] != nil {
		t.Error("a parasite caught a plague")
	}
}

// TestStarved: a parasite whose last host dies dies the same tick, and a
// plague that empties a parasite's only host ends the parasite.
func TestStarved(t *testing.T) {
	w, host, rider, _ := woken(t, 24, plague.Biological)
	w.endCiv(host, Extinct, "test")
	w.starve()
	if rider.Living() || rider.Cause != "had nothing left to wear" {
		t.Fatalf("the rider outlived its last host: %v %q", rider.Living(), rider.Cause)
	}
	// a second plague through the host
	w, host, rider, _ = woken(t, 25, plague.Biological)
	q := w.newPlague(plague.Biological, host, "clean")
	q.Contagion, q.Lethality = 1, 1
	w.infect(host, q, nil, "born")
	host.Infections[q.ID].Since = 0
	w.tickPlagues()
	if host.Living() {
		t.Fatal("a lethality-1 plague did not empty the host's only world")
	}
	if rider.Living() {
		t.Error("the parasite outlived the plague that emptied its host")
	}
}

// TestRidden: a ridden people's rising is the cure contest at the rider's
// difficulty, and winning it frees and immunises; while ridden its home is
// not lost to the toll; a rider's plague spreads on no channel of its own.
func TestRidden(t *testing.T) {
	w, host, rider, p := woken(t, 26, plague.Biological)
	host.Infections[p.ID].Since = 0
	host.Sur = 20 // certain to win the contest
	w.fightPlagues(host)
	if w.ridden(host) || !host.Immune[p.ID] || host.Infections[p.ID] != nil || !host.Scars[ScarChains] || rider.Tally.Risen != 1 {
		t.Fatalf("the rising did not free the host: master %d immune %v", host.Master, host.Immune[p.ID])
	}
	w, host, rider, p = woken(t, 27, plague.Biological)
	host.Infections[p.ID].Since = 0
	p.Lethality = 1
	host.Sur = -20 // certain to lose
	for range 5 {
		w.fightPlagues(host)
	}
	if !host.Living() || w.Owner[host.Home] != host.ID {
		t.Error("a ridden people lost its home to the toll of what rides it")
	}
	other := spawnAt(w, 2, species.Fixed("cooperative"))
	host.Trade[other.ID], other.Trade[host.ID] = true, true
	p.Contagion = 1
	for range 20 {
		w.contagion(host)
	}
	if other.Infections[p.ID] != nil || rider.Tally.Attempts != 0 {
		t.Error("a rider's plague spread on the host's account")
	}
	w.expose(host, other, "occupation")
	if rider.Tally.Attempts != 1 {
		t.Error("the host's channel event was not the rider's choice")
	}
}

// TestRiderChooses: a rider tries where its council says so, and a pact
// partner is never tried; a rider that never tries writes no FPoisoned.
func TestRiderChooses(t *testing.T) {
	w, host, rider, p := woken(t, 29, plague.Biological)
	other := spawnAt(w, 2, species.Fixed("cooperative"))
	rider.Trade[other.ID], other.Trade[rider.ID] = true, true
	rider.Met[other.ID], other.Met[rider.ID] = true, true
	w.Pacts = append(w.Pacts, &Pact{ID: 0, Members: []int{rider.ID, other.ID}, Kind: Defensive, Target: -1})
	rider.Pacts, other.Pacts = []int{0}, []int{0}
	p.Contagion = 1
	for range 20 {
		w.tryRide(rider, other, "goods")
		w.expose(host, other, "landing")
	}
	if rider.Tally.Attempts != 0 || other.Infections[p.ID] != nil {
		t.Fatal("a rider tried a pact partner")
	}
	for _, f := range w.Facts {
		if f.Kind == FPoisoned {
			t.Fatal("a rider that never tried wrote FPoisoned")
		}
	}
	// the pact gone, and in a hurry: it tries, and the attempt is the crime
	w.Pacts[0].Over = true
	w.tryRide(rider, other, "goods")
	if rider.Tally.Attempts != 1 || other.Infections[p.ID] == nil || !other.Barred[rider.ID] {
		t.Fatalf("the attempt at contagion 1 did not take: attempts %d", rider.Tally.Attempts)
	}
	if n := len(w.Facts); w.Facts[n-1].Kind != FWar || w.Facts[n-2].Kind != FBetrayal || w.Facts[n-3].Kind != FPoisoned {
		t.Errorf("the crime was not written as it should be: %v %v %v", w.Facts[len(w.Facts)-3].Kind, w.Facts[len(w.Facts)-2].Kind, w.Facts[len(w.Facts)-1].Kind)
	}
	if len(w.hostsOf(rider)) != 2 {
		t.Error("the new host is not counted")
	}
}

// TestBreakout: a crude weapon's breakout leaves its maker a host; a
// tailored one's leaves it a carrier, immune; the leak from a held
// programme is the same story.
func TestBreakout(t *testing.T) {
	w, a, _, _ := sickPair(t, 30, plague.Biological, 0.1, 0.1)
	delete(a.Infections, 0)
	a.Known["plague_craft"] = true
	a.Learned["plague_craft"] = w.Now
	w.breakout(a, tech.Get("plague_craft"), nil)
	if len(w.Plagues) != 2 {
		t.Fatal("no plague got out")
	}
	p := w.Plagues[1]
	if inf := a.Infections[p.ID]; inf == nil || inf.Carrier || a.Immune[p.ID] || p.Maker != a.ID || p.Cause != "breakout" || !p.Engineered || p.Contagion > 0.5 || p.Lethality > 0.5 {
		t.Errorf("a crude breakout: %+v, plague %+v", a.Infections[p.ID], p.Plague)
	}
	if n := len(w.Facts); w.Facts[n-1].Kind != FUnleashed || w.Facts[n-1].What != p.Tok() {
		t.Error("the breakout is not FUnleashed with the plague's name")
	}
	b := spawnAt(w, 2, species.Fixed("cooperative"))
	b.Known["tailored_plague"] = true
	b.Learned["tailored_plague"] = w.Now
	w.breakout(b, tech.Get("tailored_plague"), nil)
	q := w.Plagues[2]
	if inf := b.Infections[q.ID]; inf == nil || !inf.Carrier || !b.Immune[q.ID] {
		t.Errorf("a tailored breakout did not leave its maker an immune carrier: %+v", b.Infections[q.ID])
	}
	if b.Tally.Breakouts != 1 || a.Tally.Breakouts != 1 {
		t.Error("the breakouts are not counted")
	}
}

// TestTailored: a tailored plague catches the target's blood and its
// branches and nothing else, over a run of every channel.
func TestTailored(t *testing.T) {
	w, a, b, _ := sickPair(t, 31, plague.Biological, 0.1, 0.1)
	delete(a.Infections, 0)
	a.Known["tailored_plague"] = true
	kin := spawnAt(w, 2, b.Species.Branch())
	other := spawnAt(w, 3, species.Fixed("cooperative"))
	for _, x := range []*Civ{kin, other} {
		x.Trade[b.ID], b.Trade[x.ID] = true, true
		x.Trade[a.ID], a.Trade[x.ID] = true, true
	}
	p := w.makePlague(a, b, tech.Get("tailored_plague"), plague.AimGone)
	if p.Band != b.Species.ID || p.Lethality != 0.8 || math.Abs(p.Contagion-0.6) > 1e-9 || !p.Made || p.Maker != a.ID {
		t.Fatalf("the shape: %+v", p.Plague)
	}
	if a.Weapons["tailored_plague"] == nil || len(w.weaponUses(a)) != 1 {
		t.Fatal("the weapon is not held, or costs nothing")
	}
	p.Contagion = 1
	w.attempt(a, other, p, "poison")
	if other.Infections[p.ID] != nil || a.Tally.Attempts != 0 {
		t.Fatal("a tailored plague was tried on, or took in, a people not of the blood")
	}
	if !w.attempt(a, b, p, "poison") {
		t.Fatal("the attempt at contagion 1 did not take")
	}
	if a.Weapons["tailored_plague"] == nil {
		t.Error("attempt alone should not drop the weapon; useWeapons does")
	}
	for range 30 {
		w.contagion(b)
		w.expose(b, kin, "occupation")
		w.expose(b, other, "occupation")
	}
	if kin.Infections[p.ID] == nil {
		t.Error("the target's kin did not catch it at contagion 1 over thirty ticks")
	}
	if other.Infections[p.ID] != nil {
		t.Error("a tailored plague spread past the target's kin")
	}
}

// TestVial: the containment filter reads the craft learned, its levels
// by kind and its difficulty by rung; the scar locks the craft's use.
func TestVial(t *testing.T) {
	w, a, _, _ := sickPair(t, 32, plague.Biological, 0.1, 0.1)
	delete(a.Infections, 0)
	a.Known["agitation"] = true
	a.Learned["agitation"] = w.Now
	levels, diff, domain := filters["containment"].Adjust(w, a)
	if len(levels) != 1 || levels[0] != "soc" || diff != 1 || domain != "society" {
		t.Errorf("agitation's vial: %v %v %v", levels, diff, domain)
	}
	a.Known["black_biology"] = true
	a.Learned["black_biology"] = w.Now + 1
	levels, diff, _ = filters["containment"].Adjust(w, a)
	if levels[0] != "sur" || diff != 3 {
		t.Errorf("black biology's vial: %v %v", levels, diff)
	}
	a.Scars[ScarVial] = true
	if w.craft(a, false) != nil || w.craft(a, true) != nil {
		t.Error("the vial's scar does not lock the craft")
	}
}
