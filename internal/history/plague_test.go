package history

import (
	"testing"

	"worldgen/internal/plague"
	"worldgen/internal/species"
)

// sickPair is two peoples that trade, with a plague of the kind put into
// the first by hand.
func sickPair(t *testing.T, seed uint64, kind plague.Kind, c, l float64) (w *World, a, b *Civ, p *Plague) {
	t.Helper()
	w = newTestWorld(t, seed, 30)
	w.Now = 100_000
	a, b = spawnAt(w, 0, species.Fixed("cooperative")), spawnAt(w, 1, species.Fixed("cooperative"))
	for _, x := range []*Civ{a, b} {
		x.Reach, x.Speed = 40, 20
	}
	a.Met[b.ID], b.Met[a.ID], a.Reached[b.ID], b.Reached[a.ID] = true, true, true, true
	a.Fathomed[b.ID], b.Fathomed[a.ID] = true, true
	a.Trade[b.ID], b.Trade[a.ID] = true, true
	p = w.newPlague(kind, a, "clean")
	p.Contagion, p.Lethality = c, l
	w.infect(a, p, nil, "born")
	a.Infections[p.ID].Since = 0 // it has been in them a while
	return
}

// TestContainedOffersNothing: a contained host offers its plague on no
// channel; a raging one offers it on every one, and a certain plague
// crosses with a taken world.
func TestContainedOffersNothing(t *testing.T) {
	w, a, b, p := sickPair(t, 1, plague.Biological, 1, 0.1)
	a.Infections[p.ID].Contained = true
	for range 50 {
		w.contagion(a)
		w.expose(a, b, "occupation")
		w.expose(a, b, "landing")
	}
	if b.Infections[p.ID] != nil {
		t.Fatal("a contained host gave its plague away")
	}
	a.Infections[p.ID].Contained = false
	w.expose(a, b, "occupation")
	if b.Infections[p.ID] == nil {
		t.Fatal("a plague at contagion 1 did not cross with a taken world")
	}
	if b.Infections[p.ID].From != a.ID || b.Infections[p.ID].Road != "occupation" {
		t.Errorf("the infection remembers %+v", b.Infections[p.ID])
	}
	if f := lastFacts(w, 2); f[1].Kind != FPlagueGiven || f[0].Kind != FPlague {
		t.Error("the catch and the blame were not written")
	}
	if w.Plagues[p.ID].Hosts != 2 || p.Caught != 2 {
		t.Errorf("hosts %d caught %d", p.Hosts, p.Caught)
	}
}

// TestRefusedCarriesNothing: a people that has closed its ears to a sender
// drops its messages unread, and a plague of the mind in them never
// crosses; with the ears open it crosses at once at contagion 1.
func TestRefusedCarriesNothing(t *testing.T) {
	w, a, b, p := sickPair(t, 2, plague.Memetic, 1, 0.1)
	b.Closed[a.ID] = true
	for range 20 {
		w.Messages = append(w.Messages, &Message{From: a.ID, To: b.ID, Kind: MsgNews, About: a.ID, Fact: 0, Arrive: w.Now})
		w.tickMessages()
	}
	if b.Infections[p.ID] != nil {
		t.Fatal("a refused message carried the plague")
	}
	if b.Tally.Shut != 20 {
		t.Errorf("%d messages dropped for the closed ears, want 20", b.Tally.Shut)
	}
	delete(b.Closed, a.ID)
	w.Messages = append(w.Messages, &Message{From: a.ID, To: b.ID, Kind: MsgNews, About: a.ID, Fact: 0, Arrive: w.Now})
	w.tickMessages()
	if inf := b.Infections[p.ID]; inf == nil || inf.Road != "message" {
		t.Fatalf("a read message did not carry the plague: %+v", inf)
	}
	// the message road is no crime: the catch is written, the blame is not
	if f := lastFacts(w, 1); f[0].Kind != FPlague {
		t.Error("the catch was not the last fact")
	}
}

// TestReservoir: a world emptied by a plague keeps it, a settler catches it
// at the contagion, and the reservoir runs dry in time.
func TestReservoir(t *testing.T) {
	w, a, b, p := sickPair(t, 3, plague.Biological, 0.5, 1)
	star := 5
	a.Systems = append(a.Systems, star)
	w.Owner[star] = a.ID
	w.worldLost(a, p, star)
	r := w.Reservoir[star]
	if r == nil || r.Plague != p.ID || r.Until != w.Now+1_000_000 || w.Owner[star] != -1 {
		t.Fatalf("no reservoir at the emptied world: %+v, owner %d", r, w.Owner[star])
	}
	if !contains(a.Systems, a.Home) || contains(a.Systems, star) {
		t.Error("the wrong world went dark")
	}
	caught := 0
	for range 400 {
		delete(b.Infections, p.ID)
		w.Reservoir[star] = r
		w.wakeReservoir(b, star)
		if b.Infections[p.ID] != nil {
			caught++
		}
	}
	if caught < 170 || caught > 230 {
		t.Errorf("%d of 400 settlers woke a c = 0.5 reservoir, want about 200", caught)
	}
	w.Now = r.Until
	w.plagueBooks()
	if w.Reservoir[star] != nil {
		t.Error("the reservoir did not run dry")
	}
	_ = b
}

// TestMachineCatches: a machine-born people never catches a plague of the
// body and catches one of the mind at twice the rate of a people of flesh.
func TestMachineCatches(t *testing.T) {
	catches := func(kind plague.Kind, machine bool) int {
		n := 0
		for i := range 300 {
			w := newTestWorld(t, uint64(500+i), 30)
			w.Now = 100_000
			a := spawnAt(w, 0, species.Fixed("cooperative"))
			sp := species.Fixed("cooperative")
			if machine {
				sp = species.GenerateWith(w.R, 1, "lush", species.Machine, 0)
			}
			b := spawnAt(w, 1, sp)
			p := w.newPlague(kind, a, "clean")
			p.Contagion = 0.2
			w.infect(a, p, nil, "born")
			w.expose(a, b, "landing")
			w.expose(a, b, "message")
			if b.Infections[p.ID] != nil {
				n++
			}
		}
		return n
	}
	if got := catches(plague.Biological, true); got != 0 {
		t.Errorf("a machine-born people caught a plague of the body %d times in 300", got)
	}
	flesh, machine := catches(plague.Memetic, false), catches(plague.Memetic, true)
	if flesh < 40 || flesh > 80 {
		t.Errorf("flesh caught an idea at c 0.2 %d times in 300, want about 60", flesh)
	}
	if machine < 95 || machine > 145 {
		t.Errorf("a machine-born people caught an idea at c 0.2 %d times in 300, want about 120 (twice the flesh's %d)", machine, flesh)
	}
}

// TestCult: a world gone over to an idea may declare itself a people that
// is immune to the idea, carries it without toll, never fights it, and
// passes it on.
func TestCult(t *testing.T) {
	w, a, b, p := sickPair(t, 4, plague.Memetic, 1, 0.5)
	star := 5
	a.Systems = append(a.Systems, star)
	w.Owner[star] = a.ID
	nc := w.cult(a, p, star)
	if nc == nil || w.Owner[star] != nc.ID || !nc.Immune[p.ID] || nc.Infections[p.ID] == nil || !nc.Infections[p.ID].Carrier {
		t.Fatalf("the cult: %+v", nc)
	}
	if p.Cults != 1 || nc.Species != a.Species {
		t.Error("the cult is not counted, or not of the blood")
	}
	sur, soc := w.sickLevels(nc)
	if sur != 0 || soc != 0 {
		t.Error("a carrier pays a toll")
	}
	w.fightPlagues(nc)
	if nc.Infections[p.ID] == nil || nc.Morale != 0 {
		t.Error("a carrier fought its idea, or paid for it")
	}
	nc.Met[b.ID], b.Met[nc.ID], nc.Fathomed[b.ID], b.Fathomed[nc.ID] = true, true, true, true
	w.Messages = append(w.Messages, &Message{From: nc.ID, To: b.ID, Kind: MsgNews, About: nc.ID, Fact: 0, Arrive: w.Now})
	w.tickMessages()
	if b.Infections[p.ID] == nil || b.Infections[p.ID].From != nc.ID {
		t.Error("the cult did not pass its idea on")
	}
	// a people that has it, or has had it, catches nothing and suspects nobody for it
	w.suspicion(b)
	if b.Suspect[nc.ID] || b.Suspect[a.ID] {
		t.Error("a host suspects another of what it has")
	}
}

// TestWalls: a testament written while a people has a plague of the mind
// carries it, and a reader catches it at half the contagion; the walls
// never lose it.
func TestWalls(t *testing.T) {
	w, a, b, p := sickPair(t, 5, plague.Memetic, 1, 0.1)
	w.fact(FArise, a, nil, a.Home)
	l := &Legacy{ID: len(w.Legacies), Age: -1, Maker: a.ID, Kind: Artifact, Star: 7, Node: "writing", People: -1, Finder: -1, Source: -1, Plague: -1}
	w.Legacies = append(w.Legacies, l)
	w.testament(a, l)
	if l.Plague != p.ID {
		t.Fatalf("the walls carry %d, want the plague %d", l.Plague, p.ID)
	}
	caught := 0
	for i := range 200 {
		w2, a2, b2, _ := sickPair(t, uint64(700+i), plague.Memetic, 1, 0.1)
		l2 := &Legacy{ID: len(w2.Legacies), Age: -1, Maker: a2.ID, Kind: Artifact, Star: 7, Node: "writing", People: -1, Finder: -1, Source: -1, Plague: 0, Testament: []Inscription{{Text: "x"}}}
		w2.Legacies = append(w2.Legacies, l2)
		w2.Plagues[0].Extinct = true
		w2.readTestament(b2, l2)
		if b2.Infections[0] != nil {
			caught++
			if w2.Plagues[0].Extinct {
				t.Fatal("a plague read off the walls is still extinct")
			}
		}
	}
	if caught < 80 || caught > 120 {
		t.Errorf("%d of 200 readers caught a c = 1 idea off the walls, want about 100", caught)
	}
	_ = b
}

// TestPlagueMind: the ladder and the top of it, the levels the toll takes,
// the research it slows and bends, the weakness the neighbours read, and
// the immunity a branch inherits.
func TestPlagueMind(t *testing.T) {
	w, a, b, p := sickPair(t, 6, plague.Biological, 0.5, 0.5)
	for _, k := range []string{"scientific_method", "medicine", "chemistry", "genetics"} {
		a.Known[k] = true
	}
	if n, cure := w.rungs(a, plague.Biological); n != 2 || cure != 1 {
		t.Errorf("rungs %d cure %.1f, want 2 and 1", n, cure)
	}
	a.Shed = map[string]bool{"genetics": true}
	if n, _ := w.rungs(a, plague.Biological); n != 1 {
		t.Error("a dormant rung counts")
	}
	a.Shed = nil
	if sur, soc := w.sickLevels(a); sur != 1 || soc != 1 {
		t.Errorf("the toll on the levels: %.1f %.1f, want 1 and 1", sur, soc)
	}
	if mul, focus := w.sickRate(a); mul != 0.75 || focus["biology"] != 1.5 {
		t.Errorf("research %.2f, focus %v", mul, focus)
	}
	if !w.plagued(a) {
		t.Error("a raging l 0.5 plague is not read as weakness")
	}
	a.Infections[p.ID].Contained = true
	if w.plagued(a) {
		t.Error("a contained plague is read as weakness")
	}
	if !w.bears(a, plague.Biological) {
		t.Error("a people below the top of the ladder cannot bear")
	}
	for _, k := range []string{"immunology", "closed_ecologies", "synthetic_biology", "designed_immunity", "germline", "bodily_sovereignty"} {
		a.Known[k] = true
	}
	if w.bears(a, plague.Biological) || w.factors(a, plague.Biological).Immune != true {
		t.Error("bodily sovereignty does not end the kind")
	}
	if w.offer(a, b, p, "goods") {
		t.Error("an immune people caught it")
	}
	w.cure(a, p)
	if !a.Immune[p.ID] || a.Infections[p.ID] != nil || p.Cures != 1 {
		t.Error("the cure")
	}
	nc := w.spawnCiv(3, a.Species, -1)
	w.inherit(nc, a, 0)
	if !nc.Immune[p.ID] {
		t.Error("a branch did not inherit its parent's immunity")
	}
}
