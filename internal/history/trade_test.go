package history

import (
	"testing"

	"worldgen/internal/flow"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// partners raises two peoples on neighbouring stars, in each other's
// reach, trading, and gives the second a use for metal it cannot feed.
func partners(t *testing.T, seed uint64) (*World, *Civ, *Civ) {
	t.Helper()
	w := newTestWorld(t, seed, 30)
	a := spawnAt(w, 0, species.Fixed("cooperative"))
	b := spawnAt(w, w.G.Near(0, 1000)[0], species.Fixed("cooperative"))
	a.Trade[b.ID], b.Trade[a.ID] = true, true
	a.Met[b.ID], b.Met[a.ID] = true, true
	b.Known["firearms"] = true // weapons, era 1: 1 M, and b has no metal
	for _, c := range []*Civ{a, b} {
		w.flows(c)
		c.Reach = 1000
	}
	return w, a, b
}

// TestTradeMovesSurplus: what one people has spare goes to a partner that
// wants it, up to the cap; it is counted in the partner's next income and
// feeds what was dark, so the sum of working uses rises.
func TestTradeMovesSurplus(t *testing.T) {
	w, a, b := partners(t, 21)
	if !b.Shed["firearms"] {
		t.Fatalf("b's firearms are fed with no metal: income %v", b.Income)
	}
	a.Surplus, a.Reserved = flow.Income{flow.M: 8}, flow.Income{}
	w.trade()
	got := b.From[a.ID]
	if got[flow.M] != 1 {
		t.Fatalf("b got %v from a; want 1 M, the whole want, under a cap of 2", got)
	}
	if a.Tally.Sent[flow.M] != 1 || b.Tally.Got[flow.M] != 1 || a.Tally.Fed != 1 || a.Tally.Partners != 1 {
		t.Errorf("the tallies: sent %v got %v fed %d partners %d", a.Tally.Sent, b.Tally.Got, a.Tally.Fed, a.Tally.Partners)
	}
	w.flows(b)
	if b.Received[flow.M] != 1 || b.Shed["firearms"] {
		t.Errorf("after the sending b's firearms are dark: received %v, shed %v", b.Received, b.Shed)
	}
	if b.OwnWant[flow.M] != 1 {
		t.Errorf("b's own want is %v, want 1 M whatever was sent", b.OwnWant)
	}
	// a partner out of reach gets nothing
	a.Reach = 0
	a.Surplus = flow.Income{flow.M: 8}
	w.trade()
	if b.From[a.ID] != (flow.Income{}) {
		t.Errorf("a sent %v beyond its reach", b.From[a.ID])
	}
}

// TestEmbargo: a refusal that lasts is one fact, and it lasts.
func TestEmbargo(t *testing.T) {
	w, a, b := partners(t, 22)
	a.monsters = map[int]bool{b.ID: true}
	facts := func() int {
		n := 0
		for _, f := range w.Facts {
			if f.Kind == FEmbargo {
				n++
			}
		}
		return n
	}
	send := func() {
		a.Surplus, a.Reserved = flow.Income{flow.M: 8}, flow.Income{}
		w.trade()
	}
	send()
	if a.Refused[b.ID] != w.Now || a.Embargo[b.ID] || facts() != 0 {
		t.Fatalf("at the first refusal: refused since %d, embargo %v, facts %d", a.Refused[b.ID], a.Embargo[b.ID], facts())
	}
	w.Now += embargoYears
	send()
	if !a.Embargo[b.ID] || facts() != 1 {
		t.Fatalf("after the years: embargo %v, facts %d", a.Embargo[b.ID], facts())
	}
	w.Now += embargoYears
	send()
	if !a.Embargo[b.ID] || facts() != 1 {
		t.Errorf("an embargo that lasts wrote %d facts", facts())
	}
	if w.cause(b, a) != "the embargo" {
		t.Errorf("b's cause against a is %q", w.cause(b, a))
	}
	// relenting lifts it
	a.monsters = nil
	send()
	if a.Embargo[b.ID] || b.From[a.ID][flow.M] == 0 {
		t.Errorf("relenting: embargo %v, sent %v", a.Embargo[b.ID], b.From[a.ID])
	}
}

// TestCutOffDarkens: a people whose fed uses hang on a partner's sending
// is dependent; when the war comes, its uses go dark the same tick and
// the cut-off is a woe.
func TestCutOffDarkens(t *testing.T) {
	w, a, b := partners(t, 23)
	for range 2 {
		a.Surplus, a.Reserved = flow.Income{flow.M: 8}, flow.Income{}
		w.trade()
		w.flows(b)
	}
	if b.Shed["firearms"] {
		t.Fatal("b's firearms are dark with a's metal coming")
	}
	a.Surplus = flow.Income{flow.M: 8}
	w.trade()
	if !b.Dependent[a.ID] {
		t.Fatalf("b is not dependent on a: own income %v, working need %v", b.Income.Less(b.Received), b.WorkingNeed)
	}
	w.declare(a, b, "a test")
	if !b.Shed["firearms"] {
		t.Error("the war came and b's firearms stayed fed")
	}
	n := 0
	for _, f := range w.Facts {
		if f.Kind == FCutOff && f.Subject == b.ID && f.Object == a.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d cut-off facts, want one", n)
	}
	if len(b.From) != 0 || len(b.Dependent) != 0 {
		t.Errorf("after the cut b still has from %v dependent %v", b.From, b.Dependent)
	}
}

// TestPartnerSharesRarity: a partner has the use of an immobile rarity's
// grant and levels, and none of its yield.
func TestPartnerSharesRarity(t *testing.T) {
	w, a, b := partners(t, 24)
	n := tech.Get("causal_physics")
	full := w.price(b, n)
	soc := b.Soc
	w.addSource(&Source{Key: "horizon", Name: "a horizon", Kind: CosmicSource, Star: a.Home, Rarity: true, Grants: []string{"causal_physics"}, Levels: [3]float64{0, 0, 0.5}, Yield: flow.Income{flow.E: 5}, Holder: -1, Carried: -1, Legacy: -1})
	w.flows(a)
	w.flows(b)
	if got := w.price(b, n); got != full/2 {
		t.Errorf("with a partner's horizon causal physics costs b %v, want half of %v", got, full)
	}
	if b.Soc != soc+0.5 {
		t.Errorf("with a partner's horizon b's Soc is %v, was %v: want +0.5", b.Soc, soc)
	}
	if w.income(a)[flow.E] < 5 {
		t.Errorf("a's own horizon yields it %v", w.income(a))
	}
	if w.income(b)[flow.E] != b.Received[flow.E] {
		t.Errorf("b takes the yield of a's horizon: income %v", w.income(b))
	}
	// a mobile thing of a's is not shared
	w.addSource(&Source{Key: "artifact", Name: "a thing", Kind: MadeSource, Star: a.Home, Mobile: true, Rarity: true, Levels: [3]float64{1, 0, 0}, Holder: a.ID, Carried: -1, Legacy: -1})
	w.flows(b)
	if b.Rare["artifact"] {
		t.Error("b has the use of a's mobile artifact")
	}
}

var _ = tech.Get
