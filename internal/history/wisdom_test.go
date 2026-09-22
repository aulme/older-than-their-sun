package history

import (
	"math/rand/v2"
	"testing"

	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// alien is a species as far from species.Fixed as the difference counts:
// a machine hive with four senses the other lacks, from another world.
func alien() *species.Species {
	s := &species.Species{Sub: species.Machine, Mods: species.Hive, World: species.ArchetypeByKey("iceshell")}
	for _, k := range []string{"thermal", "magnetic", "electric", "chemical"} {
		s.Add(k)
	}
	return s
}

// pair is two peoples that have met in the flesh and rolled once, with
// Wisdom set by hand, driven by the per-tick pass alone.
func pair(t *testing.T, seed uint64, a, b *species.Species, wa, wb float64) (w *World, c, e *Civ) {
	t.Helper()
	w = newTestWorld(t, seed, 20)
	w.Now = 1000
	c, e = spawnAt(w, 0, a), spawnAt(w, 1, b)
	c.Met[e.ID], e.Met[c.ID], c.Reached[e.ID], e.Reached[c.ID] = true, true, true, true
	c.Wis, e.Wis = wa, wb
	w.fathomPair(c, e)
	return
}

// pass is a thousand years of the retries for both, each on its own
// people's stream, which is what the civs phase does (streams.go).
func pass(w *World, c, e *Civ) {
	w.Now += 1000
	w.R = w.civStream(c)
	w.fathoming(c)
	w.R = w.civStream(e)
	w.fathoming(e)
}

// TestWisdomLevel: the core adds up as the table says, and the
// experience and renaissance terms cap.
func TestWisdomLevel(t *testing.T) {
	x, p := wisdom(wisInput{Traits: []string{"contemplative", "curious", "herd"}, Profile: 1, Tech: 1.5, Experience: 30, Renaissances: 3, Communion: true, Cycle: true, Ossified: true, Slave: true})
	if p.Species != 2.5+1.5+0.5-1+1 || p.Tech != 1.5 || p.Experience != 2 || p.Boons != 1+1+1 || p.Scars != -1.5 {
		t.Errorf("parts %+v", p)
	}
	if want := p.Species + p.Tech + p.Experience + p.Boons + p.Scars; x != want {
		t.Errorf("wisdom %.2f, want %.2f", x, want)
	}
	if d := difference(species.Fixed("cooperative"), alien()); d != 6 {
		t.Errorf("the alien pair differs by %.1f, want 6", d)
	}
}

// TestCousinsFathom: a branch of a people and its parent understand each
// other at the meeting, whatever their Wisdom.
func TestCousinsFathom(t *testing.T) {
	sp := species.Fixed("cooperative")
	w, c, e := pair(t, 3, sp, sp.Branch(), 0, 0)
	if !w.mutual(c, e) {
		t.Fatalf("cousins do not understand each other: %v %v", c.Fathomed, e.Fathomed)
	}
	if !c.Trade[e.ID] {
		t.Error("mutual understanding at the meeting did not open trade")
	}
}

// TestAlienPairStaysDark: a difference-6 pair on a fixed seed stays
// unfathomed fifty thousand years, then goes one-sided when the wiser
// side gets through, and the taught side follows within twenty thousand.
func TestAlienPairStaysDark(t *testing.T) {
	w, c, e := pair(t, 12, species.Fixed("cooperative"), alien(), 8, 7)
	if w.fathoms(c, e) || w.fathoms(e, c) {
		t.Fatal("an alien pair understood each other at the meeting")
	}
	oneSided, mutual := Year(0), Year(0)
	for w.Now < 1_000_000 && mutual == 0 {
		pass(w, c, e)
		if oneSided == 0 && (w.fathoms(c, e) || w.fathoms(e, c)) {
			oneSided = w.Now
		}
		if w.mutual(c, e) {
			mutual = w.Now
		}
	}
	if oneSided == 0 || oneSided-1000 < 50_000 {
		t.Fatalf("one-sided at %d, want after 50 kyr", oneSided-1000)
	}
	if !w.fathoms(c, e) {
		t.Error("the wiser side was not the one to get through")
	}
	if mutual == 0 || mutual-oneSided > 20_000 {
		t.Errorf("the taught side followed at %d after %d kyr, want within 20", mutual, (mutual-oneSided)/1000)
	}
	if !c.Trade[e.ID] || !e.Trade[c.ID] {
		t.Error("mutual understanding did not open trade")
	}
}

// TestHostileDoesNotTeach: a conqueror that fathoms weaker prey does not
// make itself understood, and the prey's retries run cold.
func TestHostileDoesNotTeach(t *testing.T) {
	w, c, e := pair(t, 5, species.Fixed("conqueror"), alien(), 10, 7)
	c.Fathomed[e.ID] = true
	c.Mil, e.Mil = 8, 2
	w.observe(c, e, e.Home, 0)
	if w.taught(e, c) {
		t.Error("a conqueror explains itself to prey")
	}
	if adj := w.fathomAdj(e, c); adj != 0 {
		t.Errorf("the prey's adjustment is %.1f, want 0 with no teaching", adj)
	}
	e.Mil = 9
	w.observe(c, e, e.Home, 0)
	if !w.taught(e, c) {
		t.Error("a conqueror does not explain itself to an equal")
	}
	d := species.Fixed("cooperative")
	w2, c2, e2 := pair(t, 6, d, alien(), 10, 7)
	c2.Fathomed[e2.ID] = true
	c2.Wars[e2.ID] = true
	if w2.taught(e2, c2) {
		t.Error("a people at war explains itself")
	}
}

// TestBrokerCutsTheWait: over a hundred seeds of the core, a brokered
// attempt gets through far more often than a cold one.
func TestBrokerCutsTheWait(t *testing.T) {
	r := rand.New(rand.NewPCG(9, 9))
	cold, brokered := 0, 0
	for range 100 {
		if fathomRoll(5, r.NormFloat64(), 5, 0) {
			cold++
		}
		if fathomRoll(5, r.NormFloat64(), 5, -fathomBroker) {
			brokered++
		}
	}
	if cold > 15 || brokered < 40 || brokered < 3*cold {
		t.Errorf("cold %d, brokered %d of 100", cold, brokered)
	}
	// and in the world: a confederate that understands both speaks for one
	sp := species.Fixed("cooperative")
	w, c, e := pair(t, 8, sp, alien(), 4, 4)
	z := spawnAt(w, 2, species.Fixed("confederate"))
	for _, x := range []*Civ{c, e} {
		z.Fathomed[x.ID], x.Fathomed[z.ID] = true, true
		z.Met[x.ID], x.Met[z.ID] = true, true
	}
	before := len(w.Fathomings)
	for range 300 {
		w.Now += 1000
		w.brokering(z)
	}
	if z.Tally.Brokered == 0 {
		t.Fatal("a confederate that knows both never spoke for either")
	}
	spoke := 0
	for _, f := range w.Fathomings[before:] {
		if f.How == "broker" {
			spoke++
		}
	}
	if spoke == 0 {
		t.Errorf("%d brokered attempts and none got through", z.Tally.Brokered)
	}
}

// TestNoPeaceUnfathomed: a war between two peoples neither of whom
// fathoms the other does not end when one side's will runs out, and ends
// unsigned when both do.
func TestNoPeaceUnfathomed(t *testing.T) {
	w, c, e := pair(t, 4, species.Fixed("cooperative"), alien(), 0, 0)
	c.Reach, e.Reach = 30, 30
	wr := w.declare(c, e, because("border"))
	if wr == nil {
		t.Fatal("no war")
	}
	wr.Will[0] = 0
	w.judge(wr)
	if wr.Over {
		t.Fatalf("the war ended by %q with the loser not understanding the winner", wr.Result)
	}
	wr.Will[1] = 0
	w.judge(wr)
	if !wr.Over || wr.Result != "exhaustion" {
		t.Errorf("both wills gone: over %v, result %q, want exhaustion", wr.Over, wr.Result)
	}
	// one-sided: the loser that understands the winner sues for a truce
	w2, c2, e2 := pair(t, 4, species.Fixed("cooperative"), alien(), 0, 0)
	c2.Reach, e2.Reach = 30, 30
	c2.Fathomed[e2.ID] = true
	wr2 := w2.declare(c2, e2, because("border"))
	wr2.Will[0] = 0
	w2.judge(wr2)
	if !wr2.Over || wr2.Result != "truce" {
		t.Errorf("a loser that understands: over %v, result %q, want truce", wr2.Over, wr2.Result)
	}
}

// TestWiseSeal: a people at Wisdom 9 with a Sleeper seals it at least
// four times in five over a hundred runs of the Find's choice.
func TestWiseSeal(t *testing.T) {
	r := rand.New(rand.NewPCG(2, 2))
	tn := mind.Default()
	sealed := 0
	for range 100 {
		a := mind.Find(mind.FindInput{Threat: true, Wis: 9}, tn)
		if a.Pick(r.Float64()) == 2 {
			sealed++
		}
	}
	if sealed < 80 {
		t.Errorf("sealed %d of 100", sealed)
	}
	// and the look before the leap: a poor margin and a wise people
	if !mind.Leap(mind.LeapInput{Margin: -3, Wis: 9, Noise: 0}, tn) {
		t.Error("a wise people leaps at a hopeless roll")
	}
	if mind.Leap(mind.LeapInput{Margin: 1, Wis: 9, Noise: 0}, tn) {
		t.Error("a wise people balks at a fair roll")
	}
	if mind.Leap(mind.LeapInput{Margin: -3, Wis: 2, Noise: 0}, tn) {
		t.Error("a fool looks before it leaps")
	}
}

// TestUnfathomedMessageDropped: a message from a sender the recipient
// does not fathom is dropped unread; one from a sender it does is read.
func TestUnfathomedMessageDropped(t *testing.T) {
	w, c, e := pair(t, 4, species.Fixed("cooperative"), alien(), 0, 0)
	c.Trade[e.ID], e.Trade[c.ID] = true, true
	w.send(c, e, &Message{Kind: MsgPact, PactKind: Defensive, Target: -1, Pact: -1})
	w.Now += 1_000_000
	w.tickMessages()
	if e.Tally.Dropped != 1 || len(w.Pacts) != 0 {
		t.Errorf("dropped %d, pacts %d", e.Tally.Dropped, len(w.Pacts))
	}
	e.Fathomed[c.ID], c.Fathomed[e.ID] = true, true
	w.send(c, e, &Message{Kind: MsgPact, PactKind: Defensive, Target: -1, Pact: -1})
	w.Now += 1_000_000
	w.tickMessages()
	if e.Tally.Dropped != 1 {
		t.Errorf("a message from a fathomed sender was dropped")
	}
}

// TestJudgmentShrinks: the tail of the acted-on odds, the grudge's
// discount and the conqueror's compulsion all fall with Wisdom, and the
// vengeful act on their hope only below the bar.
func TestJudgmentShrinks(t *testing.T) {
	tn := mind.Default()
	in := mind.AppraiseInput{Strength: 5, Believed: 6, Spread: 2, Risk: 0.9}
	fool := mind.Appraise(in, tn)
	in.Wis = 10
	wise := mind.Appraise(in, tn)
	if fool.Acted <= fool.Odds || wise.Acted <= fool.Odds-1e-9 || wise.Acted >= fool.Acted {
		t.Errorf("fool acts on %.2f, wise on %.2f, the mean %.2f", fool.Acted, wise.Acted, fool.Odds)
	}
	if b, _, _ := mind.Bar(mind.BarInput{Posture: mind.Opportunist, Grudge: true, Wis: 10}, tn); b != tn.Bar.Opportunist {
		t.Errorf("a wise people's bar with a grudge %.2f, want %.2f", b, tn.Bar.Opportunist)
	}
	if b, _, _ := mind.Bar(mind.BarInput{Posture: mind.Opportunist, Grudge: true, Wis: 0}, tn); b != tn.Bar.Opportunist-tn.Bar.GrudgeDiscount {
		t.Errorf("a fool's bar with a grudge %.2f", b)
	}
	if c := mind.Compulsion(true, 10, tn); c >= tn.Council.Compulsion/2 {
		t.Errorf("a wise conqueror's compulsion %.3f", c)
	}
	ap := mind.Appraisal{Odds: 0.3, Low: 0.1, High: 0.6, Acted: 0.3}
	if v := mind.Judge(mind.JudgeInput{Appraisal: ap, Bar: 0.5, Front: 1, Vengeful: true, Wis: 6}, tn); v.Action != mind.Strike {
		t.Error("a vengeful fool does not act on its hope")
	}
	if v := mind.Judge(mind.JudgeInput{Appraisal: ap, Bar: 0.5, Front: 1, Vengeful: true, Wis: 8}, tn); v.Action == mind.Strike {
		t.Error("a wise vengeful people acts on its hope")
	}
	if x := mind.Folly(1, 10, 2, tn); x != 1 {
		t.Errorf("folly at ten %.2f", x)
	}
	if x := mind.Folly(1, 2.5, 1, tn); x != 1.3 {
		t.Errorf("folly at two and a half %.2f, want 1.30", x)
	}
}
