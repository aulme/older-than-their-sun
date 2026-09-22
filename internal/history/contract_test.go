package history

import (
	"testing"

	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// dealers is two peoples that understand each other, met in the flesh,
// each with a guard of so many ships at home, and a third at war with
// nobody as a threat to the first.
func dealers(t *testing.T, seed uint64, buyer, seller *species.Species, ships int) (w *World, b, s *Civ) {
	t.Helper()
	w = newTestWorld(t, seed, 30)
	w.Now = 100_000
	b, s = spawnAt(w, 0, buyer), spawnAt(w, 1, seller)
	for _, c := range []*Civ{b, s} {
		c.Reach, c.Speed = 40, 20
		w.addGuard(c, c.Home, ships)
		c.Surplus = flow.Income{2, 2, 2}
	}
	b.Met[s.ID], s.Met[b.ID], b.Reached[s.ID], s.Reached[b.ID] = true, true, true, true
	b.Fathomed[s.ID], s.Fathomed[b.ID] = true, true
	return
}

// TestOfferAccepted: an offer with a positive margin is accepted, the
// answer crosses, the guard fleet arrives, stands, and the clock starts.
func TestOfferAccepted(t *testing.T) {
	w, b, s := dealers(t, 1, species.Fixed("cooperative"), species.Fixed("faithful"), 4)
	th := spawnAt(w, 2, species.Fixed("conqueror"))
	th.Mil, b.Mil, th.Reach = 6, 3, 40
	b.Met[th.ID], th.Met[b.ID] = true, true
	w.observe(b, th, th.Home, 0)
	k := w.propose(b, s, Term{Kind: mind.TermGuard, Star: b.Home, Amount: 2, Target: th.ID})
	if k == nil || k.State != Offered || k.Pay.Kind != mind.TermFlow {
		t.Fatalf("no offer: %+v", k)
	}
	if got := w.worthFrom(s, k, k.Ask, b); got <= 0 {
		t.Errorf("a guard costs the seller %.2f", got)
	}
	for i := 0; i < 20 && k.State != Running; i++ {
		w.Now += 1000
		w.tickMessages()
	}
	if k.State != Running {
		t.Fatalf("the contract is %s, not running", k.State)
	}
	x := w.contractFleet(k)
	if x == nil || x.Kind != Relief || x.Ships != 2 {
		t.Fatalf("no guard fleet launched: %+v", x)
	}
	for i := 0; i < 50 && x.Base < 0; i++ {
		w.Now += 1000
		w.tickExpeditions()
	}
	if x.Base != b.Home || w.reliefAt(b, b.Home) != 2 {
		t.Errorf("the guard stands at %d with %d relief, want %d and 2", x.Base, w.reliefAt(b, b.Home), b.Home)
	}
	if k.Until != w.Now+Year(k.Length*1000) {
		t.Errorf("the clock did not start at arrival: until %d, now %d, length %.0f", k.Until, w.Now, k.Length)
	}
	if b.Tally.Hired != 1 || s.Tally.Sold != 1 {
		t.Errorf("tally hired %d sold %d", b.Tally.Hired, s.Tally.Sold)
	}
	w.Now = k.Until
	w.tickContracts()
	if k.State != Done || !x.Returning {
		t.Errorf("at the end the contract is %s and the fleet returning is %v", k.State, x.Returning)
	}
}

// TestShedFlowBreaks: a flow the buyer cannot feed is shed, and after
// two ticks the contract breaks with a betrayal on the buyer; the guard
// goes home at once.
func TestShedFlowBreaks(t *testing.T) {
	w, b, s := dealers(t, 2, species.Fixed("cooperative"), species.Fixed("faithful"), 4)
	k := w.newContract(b, s, Term{Kind: mind.TermGuard, Star: b.Home, Amount: 1, Target: -1}, Term{Kind: mind.TermFlow, Res: flow.M, Amount: 1000}, b)
	w.form(k)
	x := w.contractFleet(k)
	if x == nil {
		t.Fatal("no fleet")
	}
	x.Base, x.Arrive = b.Home, w.Now
	k.Until = w.Now + 50_000
	for range 3 {
		s.Loot = flow.Income{10, 10, 10} // the seller keeps its own fleets fed; a laid-up ship rots on a roll
		w.tick()
	}
	if k.State != Broken || k.Broke != b.ID {
		t.Fatalf("the contract is %s, broken by %d; want broken by the buyer %d (failed %v)", k.State, k.Broke, b.ID, k.Failed)
	}
	if !w.betrayed(s, b) {
		t.Error("no betrayal written on the buyer")
	}
	if !x.Returning && !x.Over {
		t.Error("the guard did not go home when the pay stopped")
	}
}

// TestFlowPaid: a flow the buyer can feed is paid each tick into the
// seller's income.
func TestFlowPaid(t *testing.T) {
	w, b, s := dealers(t, 3, species.Fixed("cooperative"), species.Fixed("faithful"), 1)
	k := w.newContract(b, s, Term{Kind: mind.TermPeace, Target: s.ID}, Term{Kind: mind.TermFlow, Res: flow.O, Amount: 0.5}, b)
	w.form(k)
	w.ticks(2)
	if k.State != Running || k.Failed[1] != 0 {
		t.Fatalf("the contract is %s with the pay failed %d", k.State, k.Failed[1])
	}
	if s.PaidIn[flow.O] != 0.5 || s.Tally.Got[flow.O] != 0 && s.PaidIn[flow.O] == 0 {
		t.Errorf("the seller's income took in %.2f of contract pay, want 0.5", s.PaidIn[flow.O])
	}
	if !b.Order.Has(flow.Word) {
		t.Error("the buyer's direction has no word category")
	}
}

// TestBoughtOff: a faithless guard whose employer's pay has failed is
// bought off by the people it stands against, and turns on the world
// when its new payer is at war with the old.
func TestBoughtOff(t *testing.T) {
	w, b, s := dealers(t, 4, species.Fixed("cooperative"), species.Fixed("faithless", "greedy"), 4)
	s.Dials.Greed = 1
	p := spawnAt(w, 2, species.Fixed("conqueror"))
	p.Reach, p.Speed = 40, 20
	p.Surplus = flow.Income{4, 4, 4}
	p.Met[s.ID], s.Met[p.ID], p.Fathomed[s.ID], s.Fathomed[p.ID] = true, true, true, true
	p.Met[b.ID], b.Met[p.ID] = true, true
	s.OwnWant = flow.Income{flow.M: 3}
	k := w.newContract(b, s, Term{Kind: mind.TermGuard, Star: b.Home, Amount: 2, Target: p.ID}, Term{Kind: mind.TermFlow, Res: flow.M, Amount: 0.2}, b)
	w.form(k)
	x := w.contractFleet(k)
	x.Base, x.Arrive = b.Home, w.Now
	k.Until = w.Now + 50_000
	k.Missed = true
	w.declare(p, b, because("border"))
	bought := false
	for i := 0; i < 30 && !bought; i++ {
		w.Now += 1000
		w.tickContracts()
		bought = k.BoughtOff
	}
	if !bought || k.State != Broken || k.Broke != s.ID {
		t.Fatalf("not bought off in thirty ticks: %s by %d", k.State, k.Broke)
	}
	if !x.Turned || x.Kind != Campaign {
		t.Errorf("the bought guard did not turn: kind %s turned %v", x.Kind, x.Turned)
	}
	if !w.betrayed(b, s) {
		t.Error("no betrayal on the seller")
	}
	if n := len(w.Contracts); n != 2 || w.Contracts[1].State != Running || w.Contracts[1].Ask.Kind != mind.TermPeace {
		t.Errorf("the new bargain: %d contracts, last %+v", n, w.Contracts[n-1])
	}
	// a faithful guard is never bought
	w2, b2, s2 := dealers(t, 5, species.Fixed("cooperative"), species.Fixed("faithful"), 4)
	p2 := spawnAt(w2, 2, species.Fixed("conqueror"))
	p2.Surplus = flow.Income{4, 4, 4}
	p2.Met[s2.ID], s2.Met[p2.ID], p2.Fathomed[s2.ID], s2.Fathomed[p2.ID] = true, true, true, true
	s2.OwnWant = flow.Income{flow.M: 3}
	k2 := w2.newContract(b2, s2, Term{Kind: mind.TermGuard, Star: b2.Home, Amount: 2, Target: p2.ID}, Term{Kind: mind.TermFlow, Res: flow.M, Amount: 0.2}, b2)
	w2.form(k2)
	w2.contractFleet(k2).Base = b2.Home
	k2.Until, k2.Missed = w2.Now+50_000, true
	for range 30 {
		w2.Now += 1000
		w2.tickContracts()
	}
	if k2.BoughtOff {
		t.Error("a faithful guard was bought")
	}
}

// TestTribute: a defensive winner takes tribute over worlds; a conqueror
// takes worlds.
func TestTribute(t *testing.T) {
	for _, tc := range []struct {
		winner  string
		tribute bool
	}{{"defensive", true}, {"conqueror", false}} {
		w, l, v := dealers(t, 6, species.Fixed("cooperative"), species.Fixed(tc.winner), 1)
		l.Systems = append(l.Systems, 3)
		w.Owner[3] = l.ID
		l.Surplus = flow.Income{flow.O: 5}
		wr := w.declare(v, l, because("border"))
		if wr == nil || len(w.front(v, l)) == 0 {
			t.Fatalf("%s: no war or no front", tc.winner)
		}
		w.yield(wr, wr.side(l.ID))
		if !wr.Over {
			t.Fatalf("%s: the war did not end", tc.winner)
		}
		if got := wr.Result == "tribute"; got != tc.tribute {
			t.Errorf("%s: result %q, tribute %v want %v", tc.winner, wr.Result, got, tc.tribute)
		}
		if tc.tribute {
			k := w.Contracts[len(w.Contracts)-1]
			if !k.Tribute || k.State != Running || k.Buyer != l.ID || k.Pay.Kind != mind.TermFlow || k.Pay.Res != flow.O {
				t.Errorf("the tribute: %+v", k)
			}
			if l.Tally.Tributes != 1 || len(l.Systems) != 2 {
				t.Errorf("tributes %d, worlds kept %d", l.Tally.Tributes, len(l.Systems))
			}
		} else if len(l.Systems) == 2 && l.Active() {
			t.Error("the conqueror took no world")
		}
	}
}

// TestTeachLapses: a node the buyer cannot pursue at arrival lapses the
// contract; one it can is learned and remembered as taught.
func TestTeachLapses(t *testing.T) {
	w, b, s := dealers(t, 7, species.Fixed("cooperative"), species.Fixed("faithful"), 1)
	s.Known["fusion"] = true
	k := w.newContract(b, s, Term{Kind: mind.TermTeach, Node: "fusion"}, Term{Kind: mind.TermFlow, Res: flow.O, Amount: 0.5}, b)
	w.form(k)
	b.Locked["energy"] = true
	w.Now += 1_000_000
	w.tickMessages()
	if k.State != Lapsed || b.Known["fusion"] {
		t.Errorf("a node that cannot be pursued: %s, known %v", k.State, b.Known["fusion"])
	}
	w2, b2, s2 := dealers(t, 8, species.Fixed("cooperative"), species.Fixed("faithful"), 1)
	s2.Known["fusion"] = true
	for _, p := range []string{"atomic", "computers"} {
		b2.Known[p] = true
	}
	k2 := w2.newContract(b2, s2, Term{Kind: mind.TermTeach, Node: "fusion"}, Term{Kind: mind.TermFlow, Res: flow.O, Amount: 0.5}, b2)
	w2.form(k2)
	w2.Now += 1_000_000
	w2.tickMessages()
	if !k2.Taught || b2.Taught["fusion"] != s2.ID || !b2.Known["fusion"] {
		t.Errorf("taught %v, by %d (want %d), known %v", k2.Taught, b2.Taught["fusion"], s2.ID, b2.Known["fusion"])
	}
	w2.tickContracts()
	if k2.State != Done {
		t.Errorf("a taught contract is %s", k2.State)
	}
}

// TestSoldSighting: a sighting of a pact member's fleet is refused by the
// faithful and offered by the faithless, to the people the fleet is
// coming for.
func TestSoldSighting(t *testing.T) {
	for _, tc := range []struct {
		honour string
		sells  bool
	}{{"faithful", false}, {"practical", false}, {"faithless", true}} {
		w, buyer, seller := dealers(t, 9, species.Fixed("cooperative"), species.Fixed(tc.honour), 4)
		owner := spawnAt(w, 2, species.Fixed("conqueror"))
		owner.Reach, owner.Speed = 40, 5000 // slow ships: the fleet is still crossing when the sale forms
		w.addGuard(owner, owner.Home, 6)
		w.formPact(seller, owner, Defensive, -1, -1)
		seller.Trade[buyer.ID], buyer.Trade[seller.ID] = true, true
		w.declare(owner, buyer, because("border"))
		x := w.launch(owner, Campaign, buyer, buyer.Home, 3)
		if x == nil {
			t.Fatalf("%s: no fleet", tc.honour)
		}
		seller.Sightings = map[int]*Sighting{x.ID: {Fleet: x.ID, Owner: owner.ID, Seer: seller.ID, Kind: Campaign, Star: x.Star, Launched: x.Launched, Arrive: x.Arrive, Leg: x.Launched, Year: w.Now, Ships: 3, Mil: 4}}
		w.Now += 1000
		w.sellSightings(seller)
		offered := len(w.Contracts) == 1 && w.Contracts[0].Ask.Kind == mind.TermSighting && w.Contracts[0].Buyer == buyer.ID
		if offered != tc.sells {
			t.Errorf("%s: offered %v, want %v", tc.honour, offered, tc.sells)
		}
		if !offered {
			continue
		}
		k := w.Contracts[0]
		for i := 0; i < 30 && k.State != Running; i++ {
			w.Now += 1000
			w.tickMessages()
		}
		if k.State != Running || x.SoldBy != seller.ID {
			t.Fatalf("%s: the sale is %s, sold by %d", tc.honour, k.State, x.SoldBy)
		}
		w.Now += 1_000_000
		w.tickMessages()
		if buyer.Sightings[x.ID] == nil {
			t.Error("the buyer never got the sighting")
		}
	}
}

// TestBrokerTerm: a broker contract runs until the buyer fathoms the
// target, then is done; it lapses when the broker stops understanding.
func TestBrokerTerm(t *testing.T) {
	w, b, z := dealers(t, 10, species.Fixed("cooperative"), species.Fixed("faithful"), 1)
	e := spawnAt(w, 2, alien())
	b.Met[e.ID], e.Met[b.ID], z.Met[e.ID], e.Met[z.ID] = true, true, true, true
	b.FathomTried[e.ID] = w.Now
	z.Fathomed[e.ID] = true
	b.Wis = 8
	k := w.newContract(b, z, Term{Kind: mind.TermBroker, Target: e.ID}, Term{Kind: mind.TermFlow, Res: flow.O, Amount: 0.5}, b)
	w.form(k)
	for i := 0; i < 200 && k.State == Running; i++ {
		w.Now += 1000
		w.tickContracts()
	}
	if k.State != Done || !b.Fathomed[e.ID] {
		t.Errorf("the broker's term is %s, fathomed %v", k.State, b.Fathomed[e.ID])
	}
	w2, b2, z2 := dealers(t, 11, species.Fixed("cooperative"), species.Fixed("faithful"), 1)
	e2 := spawnAt(w2, 2, alien())
	b2.FathomTried[e2.ID] = w2.Now
	z2.Fathomed[e2.ID] = true
	k2 := w2.newContract(b2, z2, Term{Kind: mind.TermBroker, Target: e2.ID}, Term{Kind: mind.TermFlow, Res: flow.O, Amount: 0.5}, b2)
	w2.form(k2)
	delete(z2.Fathomed, e2.ID)
	w2.tickContracts()
	if k2.State != Lapsed {
		t.Errorf("a broker that lost the thread: %s", k2.State)
	}
}

// TestContractMind: the margins, the honour's order, the sighting rule
// and tribute's takers in the core.
func TestContractMind(t *testing.T) {
	tn := mind.Default()
	in := mind.OfferInput{Gives: mind.TermGuard, Gets: mind.TermFlow, GiveWorth: 10, GetWorth: 12, Wis: 10}
	if d := mind.AnswerOffer(in, tn); !d.Accept || d.Margin != 1 {
		t.Errorf("a plain offer: %+v", d)
	}
	in.Greed = 0.8
	if d := mind.AnswerOffer(in, tn); d.Accept || d.Margin != 1.5 {
		t.Errorf("the greedy at 1.2 over cost: %+v", d)
	}
	in.Greed, in.Crime = 0, true
	if d := mind.AnswerOffer(in, tn); d.Accept {
		t.Error("a crime at 1.2 over cost")
	}
	in.Crime, in.Fixation, in.Gives = false, "conquest", mind.TermStrike
	in.GetWorth = 6
	if d := mind.AnswerOffer(in, tn); !d.Accept || d.Margin != 0.5 {
		t.Errorf("a conqueror selling a strike: %+v", d)
	}
	if d := mind.AnswerOffer(mind.OfferInput{GetWorth: 0, GiveWorth: 0}, tn); d.Accept {
		t.Error("nothing for nothing accepted")
	}
	for honour, want := range map[string]flow.Order{
		mind.Faithful:  {flow.Fields, flow.Word, flow.Works, flow.Mind, flow.Road, flow.Arms},
		mind.Practical: {flow.Fields, flow.Works, flow.Word, flow.Mind, flow.Road, flow.Arms},
		mind.Faithless: {flow.Fields, flow.Works, flow.Mind, flow.Road, flow.Arms, flow.Word},
	} {
		if got := mind.Direct(mind.DirectionInput{Honour: honour}, tn).Order; !sameOrder(got, want) {
			t.Errorf("%s: order %v, want %v", honour, got, want)
		}
	}
	if !sameOrder(mind.Direct(mind.DirectionInput{Honour: mind.Faithful, AtWar: true}, tn).Order, flow.Order{flow.Arms, flow.Fields, flow.Word, flow.Works, flow.Mind, flow.Road}) {
		t.Error("a faithful people at war does not pay before its works")
	}
	if mind.SellSighting(mind.Faithful, false, true) || mind.SellSighting(mind.Practical, true, false) || !mind.SellSighting(mind.Practical, false, true) || !mind.SellSighting(mind.Faithless, true, true) {
		t.Error("the sighting rule by honour")
	}
	if !mind.TakesTribute(mind.TributeInput{Posture: mind.Opportunist}) || mind.TakesTribute(mind.TributeInput{Posture: mind.Conqueror}) || !mind.TakesTribute(mind.TributeInput{Posture: mind.Conqueror, Fixation: "holding"}) {
		t.Error("tribute's takers")
	}
	if b := mind.BuyOffRate(mind.BuyOffInput{Honour: mind.Practical, NewWorth: 10, OldWorth: 5, Margin: 1}, tn); b.Rate != 0 {
		t.Error("a practical guard not yet failed is bought")
	}
	if b := mind.BuyOffRate(mind.BuyOffInput{Honour: mind.Faithless, Greed: 0.5, NewWorth: 10, OldWorth: 5, Margin: 1}, tn); b.Rate != 0.15 {
		t.Errorf("a faithless guard at greed 0.5: rate %.2f", b.Rate)
	}
}
