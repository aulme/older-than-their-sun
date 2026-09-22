package history

import (
	"testing"

	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// circle is three peoples: a supplier b that sends p a share of its
// income, and an attacker a that has met both.
func circle(t *testing.T, seed uint64, attacker *species.Species) (w *World, a, b, p *Civ) {
	t.Helper()
	w = newTestWorld(t, seed, 30)
	w.Now = 100_000
	a, b, p = spawnAt(w, 0, attacker), spawnAt(w, 1, species.Fixed("cooperative")), spawnAt(w, 2, species.Fixed("cooperative"))
	for _, c := range []*Civ{a, b, p} {
		c.Reach, c.Speed = 40, 20
		c.Income = flow.Income{10, 10, 10}
		w.addGuard(c, c.Home, 2)
	}
	for _, x := range [][2]*Civ{{a, b}, {a, p}, {b, p}} {
		x[0].Met[x[1].ID], x[1].Met[x[0].ID] = true, true
		x[0].Reached[x[1].ID], x[1].Reached[x[0].ID] = true, true
		x[0].Fathomed[x[1].ID], x[1].Fathomed[x[0].ID] = true, true
	}
	b.Trade[p.ID], p.Trade[b.ID] = true, true
	p.From = map[int]flow.Income{b.ID: {flow.O: 2}} // a fifteenth of p's income
	return
}

// TestSlightSized: a war on a partner's supplier writes a grudge on the
// attacker sized by the flow, a dependent partner takes double, and a
// people that trades with nobody takes none.
func TestSlightSized(t *testing.T) {
	w, a, b, p := circle(t, 1, species.Fixed("conqueror"))
	q := spawnAt(w, 3, species.Fixed("cooperative"))
	q.Income = flow.Income{10, 10, 10}
	wr := w.declare(a, b, because("border"))
	if wr == nil {
		t.Fatal("no war")
	}
	if got := p.Grudge[a.ID]; got < 0.13 || got > 0.14 {
		t.Errorf("the slight is %.2f, want 0.13 (two of thirty at weight two)", got)
	}
	if q.Grudge[a.ID] != 0 {
		t.Error("a people that trades with nobody took a slight")
	}
	if wr.SlightTold[p.ID] || p.Tally.Slights != wr.Slighted[p.ID] {
		t.Errorf("told %v at %.2f, tally %.2f", wr.SlightTold[p.ID], wr.Slighted[p.ID], p.Tally.Slights)
	}
	// dependence doubles it
	w2, a2, b2, p2 := circle(t, 2, species.Fixed("conqueror"))
	p2.Dependent = map[int]bool{b2.ID: true}
	w2.declare(a2, b2, because("conquest"))
	if got := p2.Grudge[a2.ID]; got < 0.26 || got > 0.27 {
		t.Errorf("a dependent partner's slight is %.2f, want 0.27", got)
	}
	if !w2.Wars[0].SlightTold[p2.ID] {
		t.Error("a slight past the bar wrote no fact")
	}
	if f := lastFacts(w2, 1); f[0] == nil || f[0].Kind != FSlight {
		t.Error("the last fact is not the slight")
	}
}

// TestSlightRepeats: each tick the supplier sends less than before the
// war, the partner takes a tenth of the slight again, to the cap, and
// only one fact is written.
func TestSlightRepeats(t *testing.T) {
	w, a, b, p := circle(t, 3, species.Fixed("conqueror"))
	wr := w.declare(a, b, because("border"))
	first := p.Grudge[a.ID]
	p.From = map[int]flow.Income{}
	for range 5 {
		w.slightTick(wr)
	}
	if got := p.Grudge[a.ID]; got < first*1.49 || got > first*1.51 {
		t.Errorf("after five starved ticks the grudge is %.2f, want %.2f", got, first*1.5)
	}
	p.From = map[int]flow.Income{b.ID: {flow.O: 2}}
	held := p.Grudge[a.ID]
	w.slightTick(wr)
	if got := p.Grudge[a.ID]; got != held {
		t.Error("a partner sent as before still took a slight")
	}
	p.From = map[int]flow.Income{}
	for range 200 {
		w.slightTick(wr)
	}
	if got := wr.Slighted[p.ID]; got > 1.0001 {
		t.Errorf("the slight of one war passed the cap: %.2f", got)
	}
	told := 0
	for _, f := range w.Events {
		if f.Kind == FSlight {
			told++
		}
	}
	if told != 1 {
		t.Errorf("%d slight facts for one war, want one", told)
	}
}

// TestOffenceRaisesTheBar: the appraisal's bar rises with the offence, a
// conqueror's by half as much, and a war that would slight nobody the
// attacker minds is judged as before.
func TestOffenceRaisesTheBar(t *testing.T) {
	tn := mind.Default()
	base := mind.AppraiseInput{Strength: 5, Believed: 4, Spread: 0.3, Risk: 0.5, Prize: 2}
	plain := mind.Appraise(base, tn)
	with := base
	with.Slights = []mind.Slighted{{Name: "X", Slight: 0.3}, {Name: "Y", Slight: 0.1}}
	off := mind.Appraise(with, tn)
	if off.Prize >= plain.Prize {
		t.Errorf("the offence did not raise the bar: prize %.3f with, %.3f without", off.Prize, plain.Prize)
	}
	if want := max(-tn.Appraise.PrizeMax, tn.Appraise.PrizeWeight*(2-5*0.4)); off.Prize-want > 1e-9 || want-off.Prize > 1e-9 {
		t.Errorf("prize with the offence %.3f, want %.3f", off.Prize, want)
	}
	if off.Alone != plain.Prize {
		t.Error("the prize alone is not the plain prize")
	}
	with.Conqueror = true
	conq := mind.Appraise(with, tn)
	if conq.Prize <= off.Prize || conq.Prize >= plain.Prize {
		t.Errorf("a conqueror's offence: %.3f, want between %.3f and %.3f", conq.Prize, off.Prize, plain.Prize)
	}
	if why := off.Why(); why == plain.Why() {
		t.Error("the appraisal does not name the slighted")
	}
	// the verdict knows when the offence alone held them
	v := mind.Judge(mind.JudgeInput{Appraisal: mind.Appraisal{Acted: 0.44, Low: 0.3, High: 0.6, Prize: -0.1, Alone: 0.04}, Bar: 0.4, Front: 1}, tn)
	if v.Action == mind.Strike || !v.Deterred {
		t.Errorf("acting on 0.44 against 0.5 with the offence and 0.36 without: %s, deterred %v", v.Action, v.Deterred)
	}
	// care: partners and allies in full, a strong neighbour by half and double by fear, nothing for monsters
	if mind.Care(mind.CareInput{Partner: true}, tn) != 1 || mind.Care(mind.CareInput{Strong: true}, tn) != 0.5 || mind.Care(mind.CareInput{Strong: true, Fear: 0.9}, tn) != 1 || mind.Care(mind.CareInput{Partner: true, Monster: true}, tn) != 0 {
		t.Error("the care weights")
	}
}

// TestOffenceInTheWorld: the attacker's appraisal of a supplier names
// the partner it trades with, and the second war on the same supplier is
// answered colder in the pact.
func TestOffenceInTheWorld(t *testing.T) {
	w, a, b, p := circle(t, 4, species.Fixed("conqueror"))
	a.Trade[p.ID], p.Trade[a.ID] = true, true
	w.observe(a, b, b.Home, 0)
	ap := w.appraise(a, b, -1)
	if len(ap.Slights) != 1 || ap.Slights[0].Name != p.Tok() || ap.Slights[0].Slight < 0.13 {
		t.Fatalf("the offence: %+v", ap.Slights)
	}
	delete(a.Trade, p.ID)
	delete(p.Trade, a.ID)
	if ap := w.appraise(a, b, -1); len(ap.Slights) != 0 {
		t.Error("a people the attacker does not mind is in the offence")
	}
	// the pact answer reads the grudge the slight left
	cold := func() bool {
		in := mind.AnswerInput{Posture: p.posture(), Mil: 3, ProposerMil: 3, Dials: p.Dials, Wis: 10, ProposerGrudge: p.Grudge[a.ID], Target: true, Believed: 3}
		return mind.AnswerPact(in, w.Cfg.Tuning).Score < mind.AnswerPact(mind.AnswerInput{Posture: p.posture(), Mil: 3, ProposerMil: 3, Dials: p.Dials, Wis: 10, Target: true, Believed: 3}, w.Cfg.Tuning).Score
	}
	w.declare(a, b, because("border"))
	w.endWar(w.Wars[0], "peace")
	if p.Grudge[a.ID] <= 0 {
		t.Fatal("no grudge from the slight")
	}
	if !cold() {
		t.Error("the slighted people answers the attacker's pact as warmly as before")
	}
	w.declare(a, b, because("border"))
	if p.Grudge[a.ID] < 0.26 {
		t.Errorf("the second war left a grudge of %.2f, want two slights", p.Grudge[a.ID])
	}
}

// TestSourceSlight: a world taken that held a source the partner drew on
// is a slight of its own.
func TestSourceSlight(t *testing.T) {
	w, a, b, p := circle(t, 5, species.Fixed("conqueror"))
	star := 3
	b.Systems = append(b.Systems, star)
	w.Owner[star] = b.ID
	s := w.addSource(&Source{Key: "diamond", Star: star, Rarity: true, Grants: []string{"exotic_matter"}, Holder: b.ID, Carried: -1, Legacy: -1})
	_ = s
	wr := w.declare(a, b, because("border"))
	before := p.Grudge[a.ID]
	w.sourceSlight(wr, a, b, star)
	if got := p.Grudge[a.ID] - before; got < 0.29 || got > 0.31 {
		t.Errorf("the source slight is %.2f, want 0.3", got)
	}
	w.sourceSlight(wr, a, b, b.Home)
	if got := p.Grudge[a.ID] - before; got > 0.31 {
		t.Error("a world with no shared source slighted")
	}
}
