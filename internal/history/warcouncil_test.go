package history

import (
	"testing"

	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// atWar is twoPeoples with the two understanding each other and each
// knowing the other's sky, c's guard at home holding cShips and e's
// eShips, and c's war declared on e for the cause.
func atWar(t *testing.T, seed uint64, cShips, eShips int, cause string) (w *World, c, e *Civ, colony int, wr *War) {
	t.Helper()
	w, c, e, colony = twoPeoples(t, seed)
	c.Fathomed[e.ID], e.Fathomed[c.ID] = true, true
	c.Met[e.ID], e.Met[c.ID] = true, true
	w.guardAt(c, c.Home).Ships = cShips
	w.addGuard(e, e.Home, eShips)
	wr = w.declare(c, e, because(cause))
	if wr == nil {
		t.Fatal("no war")
	}
	w.observe(c, e, e.Home, 0)
	w.observe(e, c, c.Home, 0)
	return
}

// campaignOf is c's campaign against e, or its muster, or nil.
func campaignOf(w *World, c, e *Civ) bool {
	for _, x := range w.fleetsOf(c) {
		if x.Kind == Campaign && x.Target == e.ID {
			return true
		}
	}
	return c.Muster != nil && c.Muster.Target == e.ID
}

// TestDefenderStrikesBack: the side declared on, far the stronger, carries
// the war to the declarer.
func TestDefenderStrikesBack(t *testing.T) {
	w, c, e, _, wr := atWar(t, 61, 1, 30, "border")
	e.Dials.Fear, e.Dials.Risk = 0, 0.5
	w.warCouncil(wr, e, c, 1)
	if wr.Verdict[1] != mind.Press || !campaignOf(w, e, c) {
		in, _ := w.warInput(wr, e, c, 1)
		t.Fatalf("the stronger side declared on did not strike back: verdict %s, %+v", wr.Verdict[1], in)
	}
}

// TestInheritedWarGetsCampaigns: a war declared with no fleet behind it,
// as an heir's is at a sundering, gets one from the war council.
func TestInheritedWarGetsCampaigns(t *testing.T) {
	w, c, e, _, wr := atWar(t, 62, 30, 1, "sundering")
	if campaignOf(w, c, e) {
		t.Fatal("a campaign before the council sat")
	}
	c.Dials.Fear, c.Dials.Risk = 0, 0.5
	w.warCouncil(wr, c, e, 0)
	if wr.Verdict[0] != mind.Press || !campaignOf(w, c, e) {
		t.Fatalf("the inherited war got no campaign: verdict %s", wr.Verdict[0])
	}
}

// TestLimitedWarEndsOnTerms: a war for a world ends once the world is
// taken: the declarer offers peace on the lines and the side declared
// on, too weak to win it back, accepts.
func TestLimitedWarEndsOnTerms(t *testing.T) {
	w, c, e, colony, wr := atWar(t, 63, 30, 1, "border")
	if wr.Aim != mind.AimWorld {
		t.Fatalf("a border war's aim is %q", wr.Aim)
	}
	w.takeWorld(wr, c, e, colony)
	if !wr.aimMet(0) || wr.Over {
		t.Fatalf("the world taken: aim met %v, over %v", wr.aimMet(0), wr.Over)
	}
	w.warCouncil(wr, c, e, 0)
	if !wr.Over || wr.Result != "terms" {
		t.Fatalf("the war went on: verdict %s, result %q", wr.Verdict[0], wr.Result)
	}
	f := lastFacts(w, 1)[0]
	if f.Kind != FSettled || f.P["terms"] != "lines" || f.P["net"] != 1 {
		t.Errorf("the settlement: %v", f)
	}
}

// TestFearSues: a side that has lost a world and believes the balance
// against it offers what the other's aim wants, and the terms are taken.
func TestFearSues(t *testing.T) {
	w, c, e, colony, wr := atWar(t, 64, 30, 1, "border")
	other := -1
	for _, s := range w.G.Near(e.Home, fleetHop) {
		if w.Owner[s] < 0 && s != c.Home && s != colony {
			other = s
			break
		}
	}
	if other < 0 {
		t.Skip("no second free star near the enemy")
	}
	w.Owner[other] = e.ID
	e.Systems = append(e.Systems, other)
	wr.Battles, wr.Fought, wr.Lost[1] = 1, w.Now, 1 // it has fought and lost
	e.Dials.Fear = 1
	c.Grudge[e.ID] = 1 // an old quarrel
	w.warCouncil(wr, e, c, 1)
	if wr.Verdict[1] != mind.Sue {
		t.Fatalf("the losing side's verdict %s", wr.Verdict[1])
	}
	if !wr.Over || wr.Result != "terms" {
		t.Fatalf("the terms were not taken: result %q", wr.Result)
	}
	if f := lastFacts(w, 1)[0]; f.Kind != FSettled || f.P["terms"] != "worlds" || f.P["n"] != 1 {
		t.Errorf("the settlement: %v", f)
	}
	// the side that took the terms won the war, and has no quarrel left
	// to take up again when the truce runs out
	if wr.Winner != c.ID || c.Grudge[e.ID] > 0 {
		t.Errorf("winner %d (want %d), grudge %.2f", wr.Winner, c.ID, c.Grudge[e.ID])
	}
}

// TestSideDeclaredOnWins: the declarer's will spent while the other's
// holds is a peace the side declared on wins, keeping what it took.
func TestSideDeclaredOnWins(t *testing.T) {
	w, c, e, _, wr := atWar(t, 65, 1, 1, "border")
	wr.Will = [2]float64{0, 1}
	wr.Taken[1] = 1
	w.judge(wr)
	if !wr.Over || wr.Result != "peace" {
		t.Fatalf("result %q", wr.Result)
	}
	if f := lastFacts(w, 1)[0]; f.Kind != FPeace || f.P["net"] != -1 || f.P["by"] != c.ID {
		t.Errorf("the peace: %v", f)
	}
	_ = e
}

// TestWillFollowsBattles: a battle lost costs the loser will beyond the
// roll's tenth, by the ships it lost.
func TestWillFollowsBattles(t *testing.T) {
	lost := 0
	for seed := uint64(70); seed < 80; seed++ {
		w, c, e, colony := twoPeoples(t, seed)
		w.addGuard(e, colony, 40)
		x := campaignAt(w, c, e, colony, 2)
		wr := w.warBetween(c.ID, e.ID)
		wr.Will = [2]float64{2, 2}
		w.fight(x, colony)
		if wr.Will[0] < 2-0.1-1e-9 {
			lost++
		}
		if wr.Will[1] < 2 {
			t.Errorf("seed %d: the defender who held lost will: %.2f", seed, wr.Will[1])
		}
	}
	if lost < 8 {
		t.Errorf("a hopeless attack cost its side more than a tenth of its will in %d of 10", lost)
	}
}

// TestYoke: a large conqueror about to strike a small people it would
// clearly beat offers vassalage first, and the small people, believing
// the war lost, bends the knee.
func TestYoke(t *testing.T) {
	w, c, e, _ := twoPeoples(t, 66)
	c.Species = species.Fixed("conqueror")
	c.Fathomed[e.ID], e.Fathomed[c.ID] = true, true
	c.Met[e.ID], e.Met[c.ID] = true, true
	for _, s := range w.G.Near(c.Home, 40) {
		if len(c.Systems) >= 8 {
			break
		}
		if w.Owner[s] < 0 && s != e.Home {
			w.Owner[s] = c.ID
			c.Systems = append(c.Systems, s)
		}
	}
	if len(c.Systems) < 8 {
		t.Skip("no room for a large realm")
	}
	w.guardAt(c, c.Home).Ships = 40
	w.observe(c, e, e.Home, 0)
	w.observe(e, c, c.Home, 0)
	ap := w.appraise(c, e, -1)
	if !w.yoke(c, e, ap) {
		t.Fatalf("no offer: acting on %.2f", ap.Acted)
	}
	if e.Free() || !e.Vassal || e.Master != c.ID {
		t.Errorf("the small people did not bend: master %d", e.Master)
	}
}

// TestStrengthCountsShipsOut: a people's strength against an enemy counts
// the ships it has out on campaign against it, so a conqueror that sends
// most of its fleet does not read itself beaten the tick it sails.
func TestStrengthCountsShipsOut(t *testing.T) {
	w, c, e, colony, _ := atWar(t, 66, 10, 3, "border")
	before := w.strength(c, e)
	if w.launch(c, Campaign, e, colony, 8) == nil {
		t.Fatal("no fleet")
	}
	if after := w.strength(c, e); after < before-0.01 {
		t.Errorf("strength %.2f before the fleet sailed, %.2f after", before, after)
	}
	if w.strength(c, w.Civs[e.ID]) <= w.strength(e, c) {
		t.Error("the stronger side reads itself the weaker")
	}
}

// TestBoughtOffWins: a side that sues and buys the war off with less
// than the other's whole aim has still lost it: the side that took the
// terms is the winner, and its quarrel is ended rather than renewed. The
// loop was a vengeful people bought off with half its redress, its grudge
// left standing and fed again at the war's end, declaring again every
// twenty-five ticks.
func TestBoughtOffWins(t *testing.T) {
	w, c, e, colony, wr := atWar(t, 67, 30, 1, "revenge")
	if wr.Aim != mind.AimRedress {
		t.Fatalf("a revenge war's aim is %q", wr.Aim)
	}
	c.Grudge[e.ID] = 2
	o := offer{kind: "worlds", worlds: []int{colony}}
	paid := worth(wr, 0, o)
	if paid >= 1 {
		t.Fatalf("one world pays a redress in full (%.2f)", paid)
	}
	w.settleTerms(wr, e, c, o, paid)
	if !wr.Over || wr.Result != "terms" {
		t.Fatalf("not settled: result %q", wr.Result)
	}
	if wr.Winner != c.ID || c.Grudge[e.ID] > 0 {
		t.Errorf("winner %d (want %d), grudge %.2f", wr.Winner, c.ID, c.Grudge[e.ID])
	}
}
