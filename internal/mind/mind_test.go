package mind

import (
	"math"
	"math/rand/v2"
	"testing"

	"worldgen/internal/flow"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// TestAppraise: the closed form. Strength against believed level plus the
// world's defence, as odds through the normal; the home and a grid add;
// risk moves what is acted on between the tails.
func TestAppraise(t *testing.T) {
	tn := Default()
	cases := []struct {
		name   string
		in     AppraiseInput
		margin float64
	}{
		{"even worlds", AppraiseInput{Strength: 4, Believed: 3, Spread: 0.3, Risk: 0.5}, 0},
		{"the home", AppraiseInput{Strength: 4, Believed: 3, Spread: 0.3, Risk: 0.5, AtHome: true}, -2.5},
		{"a grid and relief", AppraiseInput{Strength: 6, Believed: 3, Spread: 1, Risk: 0.5, Grid: true, Relief: 1}, 0.5},
		{"a weakened enemy in two other wars", AppraiseInput{Strength: 4, Believed: 4, Spread: 1, Risk: 0.5, Weakened: true, OtherWars: 2}, 0.6},
	}
	for _, c := range cases {
		a := Appraise(c.in, tn)
		if !near(a.Margin, c.margin) {
			t.Errorf("%s: margin %.3f, want %.3f", c.name, a.Margin, c.margin)
		}
		if want := phi(c.margin / 2); !near(a.Odds, want) {
			t.Errorf("%s: odds %.3f, want %.3f", c.name, a.Odds, want)
		}
		if !near(a.Acted, a.Odds) {
			t.Errorf("%s: at risk 0.5 acted %.3f should be the mean %.3f", c.name, a.Acted, a.Odds)
		}
		if !(a.Low <= a.Odds && a.Odds <= a.High) {
			t.Errorf("%s: tails %.3f..%.3f do not bracket %.3f", c.name, a.Low, a.High, a.Odds)
		}
	}
	bold := Appraise(AppraiseInput{Strength: 4, Believed: 3, Spread: 1, Risk: 1}, tn)
	shy := Appraise(AppraiseInput{Strength: 4, Believed: 3, Spread: 1, Risk: 0}, tn)
	if !near(bold.Acted, bold.High) || !near(shy.Acted, shy.Low) {
		t.Errorf("risk 1 acts on the high tail (%.3f vs %.3f), risk 0 on the low (%.3f vs %.3f)", bold.Acted, bold.High, shy.Acted, shy.Low)
	}
	if a := Appraise(AppraiseInput{Strength: 4, Believed: 3, Dist: 20, Speed: 100}, tn); a.Lag != 2000 {
		t.Errorf("lag %.0f, want 2000", a.Lag)
	}
}

func TestStrengthAndBelief(t *testing.T) {
	tn := Default()
	if s := Strength(StrengthInput{Mil: 4, Bonus: 1, Allies: []float64{2, 4}, OtherWars: 1}, tn); !near(s, 4+1+1+2-0.3) {
		t.Errorf("strength %.2f", s)
	}
	if mil, spread := Believe(BeliefInput{EnemyEra: 2}, tn); mil != 3.4 || spread != 3 {
		t.Errorf("unseen: %.2f ± %.2f", mil, spread)
	}
	if mil, spread := Believe(BeliefInput{Seen: true, Mil: 5, AgeKyr: 10}, tn); mil != 5 || !near(spread, 1.3) {
		t.Errorf("seen ten thousand years ago: %.2f ± %.2f", mil, spread)
	}
	if _, spread := Believe(BeliefInput{Seen: true, Mil: 5, AgeKyr: 1000}, tn); spread != 3 {
		t.Errorf("an old report caps at three: %.2f", spread)
	}
}

// TestBar: the odds each posture needs, and whether it sends fleets.
func TestBar(t *testing.T) {
	tn := Default()
	cases := []struct {
		in    BarInput
		bar   float64
		wants bool
		far   bool
	}{
		{BarInput{Posture: Pacifist}, 0, false, false},
		{BarInput{Posture: Pacifist, Hates: true}, 0, false, false}, // a pacifist is a pacifist first
		{BarInput{Posture: Defensive, Hates: true}, 0.35, true, true},
		{BarInput{Posture: Defensive}, 0, false, false},
		{BarInput{Posture: Opportunist}, 0.75, true, false},
		{BarInput{Posture: Opportunist, Grudge: true}, 0.65, true, false},
		{BarInput{Posture: Conqueror}, 0.4, true, true},
		{BarInput{Posture: Conqueror, Aloft: true}, 0.4, true, false},
		{BarInput{Posture: Vengeful}, 0, false, false},
		{BarInput{Posture: Vengeful, Grudge: true}, 0.3, true, true},
	}
	for _, c := range cases {
		bar, wants, far := Bar(c.in, tn)
		if !near(bar, c.bar) || wants != c.wants || far != c.far {
			t.Errorf("%+v: got %.2f %v %v, want %.2f %v %v", c.in, bar, wants, far, c.bar, c.wants, c.far)
		}
	}
}

// TestCouncil: strike over the bar, scout inside the band, watch below
// it; the council takes the best margin.
func TestCouncil(t *testing.T) {
	tn := Default()
	ap := func(margin, spread float64) Appraisal {
		return Appraise(AppraiseInput{Strength: margin + 1, Believed: 0, Spread: spread, Risk: 0.5}, tn)
	}
	cases := []struct {
		name string
		in   JudgeInput
		want Action
	}{
		{"clear", JudgeInput{Appraisal: ap(2, 0.3), Bar: 0.4, Front: 1}, Strike},
		{"straddles", JudgeInput{Appraisal: ap(0, 1), Bar: 0.6, Front: 1}, ScoutFirst},
		{"hopeless", JudgeInput{Appraisal: ap(-3, 0.3), Bar: 0.4, Front: 1}, Watch},
		{"no front, no fleet", JudgeInput{Appraisal: ap(2, 0.3), Bar: 0.4}, Nothing},
		{"no front, a fleet", JudgeInput{Appraisal: ap(2, 0.3), Bar: 0.4, Far: true}, Strike},
		{"vengeful acts on hope", JudgeInput{Appraisal: ap(-0.5, 2), Bar: 0.3, Front: 1, Vengeful: true}, Strike},
		{"compelled", JudgeInput{Appraisal: ap(-1, 0.3), Bar: 0.4, Front: 1, Compelled: true}, Strike},
	}
	for _, c := range cases {
		if v := Judge(c.in, tn); v.Action != c.want {
			t.Errorf("%s: %s, want %s (%s)", c.name, v.Action, c.want, v.Why())
		}
	}
	verdicts := []Verdict{
		Judge(JudgeInput{Appraisal: ap(1, 0.3), Bar: 0.4, Front: 1}, tn),
		Judge(JudgeInput{Appraisal: ap(3, 0.3), Bar: 0.4, Front: 1}, tn),
		Judge(JudgeInput{Appraisal: ap(3, 0.3), Bar: 0.9, Front: 1}, tn),
		Judge(JudgeInput{Appraisal: ap(-2, 0.3), Bar: 0.4, Front: 1}, tn),
	}
	if i := Council(verdicts); i != 1 {
		t.Errorf("council chose %d, want 1 (the widest margin)", i)
	}
	if i := Council(verdicts[3:]); i != -1 {
		t.Errorf("council with nothing over the bar chose %d", i)
	}
}

func TestScoutAndCampaign(t *testing.T) {
	tn := Default()
	if s := Scout(ScoutInput{Sight: true, Mil: 1}, tn); !s.Look {
		t.Error("the Sight reads for free")
	}
	if s := Scout(ScoutInput{Mil: 1.5, Fear: 0.2}, tn); !s.Kept {
		t.Error("a level must stay home")
	}
	if s := Scout(ScoutInput{Mil: 3, Fear: 0.9}, tn); !s.Kept {
		t.Error("the fearful and small do not scout")
	}
	if s := Scout(ScoutInput{Mil: 3, Fear: 0.2, Out: true}, tn); s.Send || s.Kept {
		t.Error("one scout at a time")
	}
	if s := Scout(ScoutInput{Mil: 3, Fear: 0.2}, tn); !s.Send {
		t.Error("a scout goes")
	}
	a := Appraise(AppraiseInput{Strength: 6, Believed: 3, Spread: 0.5, Risk: 0.5, Dist: 10, Speed: 100}, tn)
	k := SizeCampaign(CampaignInput{Appraisal: a, Strength: 6, Mil: 6, Away: 0, Risk: 0.5, Fear: 0.2}, tn)
	if !k.Send || !near(k.Need, 4) || !near(k.Share, 4) {
		t.Errorf("even odds at need 4: %s", k.Why())
	}
	k = SizeCampaign(CampaignInput{Appraisal: a, Strength: 6, Mil: 6, Risk: 0.5, Fear: 1}, tn)
	if k.Send {
		t.Errorf("full fear caps the fleet below the need: %s", k.Why())
	}
	far := Appraise(AppraiseInput{Strength: 6, Believed: 3, Spread: 0.5, Risk: 0.5, Dist: 300, Speed: 100}, tn)
	if k := SizeCampaign(CampaignInput{Appraisal: far, Strength: 6, Mil: 6, Risk: 0.5, Fear: 0.2}, tn); k.Send {
		t.Error("nobody crosses for thirty thousand years")
	}
	if k := SizeCampaign(CampaignInput{Appraisal: far, Strength: 6, Mil: 6, Risk: 0.5, Fear: 0.2, Conqueror: true}, tn); !k.Send {
		t.Error("a conqueror crosses for thirty thousand years")
	}
}

// honours as dials: the loyalty a faithful, practical and faithless people
// is born with.
var honours = []struct {
	name    string
	loyalty float64
}{{Faithful, 0.9}, {Practical, 0.5}, {Faithless, 0.1}}

// TestAnswerPact: an offer of defence against a neighbour of one's own
// level is taken by the defensive unless faithless, and refused by a
// conqueror; an offer of war is refused out of hand by the meek and taken
// by a conqueror.
func TestAnswerPact(t *testing.T) {
	tn := Default()
	for _, h := range honours {
		d := Dials{Fear: 0.6, Greed: 0.5, Loyalty: h.loyalty}
		in := AnswerInput{Posture: Defensive, Target: true, Believed: 3, Mil: 3, ProposerMil: 4, Dials: d}
		a := AnswerPact(in, tn)
		want := h.name != Faithless
		if a.Accept != want {
			t.Errorf("%s defensive people offered defence: %s, want accept %v", h.name, a.Why(), want)
		}
		in.Posture = Conqueror
		if a := AnswerPact(in, tn); a.Accept {
			t.Errorf("%s conqueror offered defence: %s", h.name, a.Why())
		}
		war := AnswerInput{Aggressive: true, Posture: Pacifist, Target: true, Dials: d}
		if a := AnswerPact(war, tn); a.Accept || a.Reason == "" {
			t.Errorf("%s pacifist offered war: %s", h.name, a.Why())
		}
		war.Posture = Conqueror
		if a := AnswerPact(war, tn); !a.Accept {
			t.Errorf("%s conqueror offered war: %s", h.name, a.Why())
		}
	}
	in := AnswerInput{Posture: Defensive, Target: true, Believed: 5, Mil: 3, EnemyNear: true, Dials: Dials{Fear: 0.6, Loyalty: 0.5}}
	if a := AnswerPact(in, tn); !a.Accept {
		t.Fatalf("baseline: %s", a.Why())
	}
	in.Infamy = 2
	if a := AnswerPact(in, tn); a.Accept {
		t.Errorf("an infamous proposer: %s", a.Why())
	}
	in.Infamy, in.Difference = 0, 6
	if a := AnswerPact(in, tn); a.Accept {
		t.Errorf("a very different proposer: %s", a.Why())
	}
}

// TestAnswerCall: the faithful come when it helps and home is safe; the
// faithless do not; nobody sends what would not help; a small people is
// not blamed for staying.
func TestAnswerCall(t *testing.T) {
	tn := Default()
	for _, h := range honours {
		in := CallInput{Mil: 6, VictimMil: 3, Believed: 5, Dials: Dials{Fear: 0.4, Loyalty: h.loyalty}}
		k := AnswerCall(in, tn)
		want := h.name != Faithless
		if k.Come != want {
			t.Errorf("%s called: %s, want come %v", h.name, k.Why(), want)
		}
		if !near(k.Share, 1.8) {
			t.Errorf("%s share %.2f, want a third of six", h.name, k.Share)
		}
	}
	if k := AnswerCall(CallInput{Mil: 6, VictimMil: 1, Believed: 9, Dials: Dials{Fear: 0.4, Loyalty: 0.9}}, tn); k.Come || k.Helps {
		t.Errorf("relief that cannot help: %s", k.Why())
	}
	if k := AnswerCall(CallInput{Mil: 2, VictimMil: 3, Believed: 3, Dials: Dials{Fear: 0.5, Loyalty: 0.9}}, tn); k.Come || k.Blame {
		t.Errorf("a small people keeps its level and is not blamed: %s", k.Why())
	}
	if k := AnswerCall(CallInput{Mil: 6, VictimMil: 3, Believed: 5, Betrayed: true, Dials: Dials{Fear: 0.4, Loyalty: 0.9}}, tn); k.Come {
		t.Errorf("a betrayer calling: %s", k.Why())
	}
	if k := AnswerCall(CallInput{Mil: 6, VictimMil: 3, Believed: 5, Confederate: true, Dials: Dials{Fear: 0.9, Loyalty: 0.4}}, tn); k.Come {
		t.Errorf("a fearful confederate that cannot leave home safe: %s", k.Why())
	}
}

func TestForwardAndTurn(t *testing.T) {
	tn := Default()
	loyal := Dials{Loyalty: 0.9, Greed: 0.5}
	if ok, _ := Forward(ForwardInput{AllyMet: true, Dials: loyal}, tn); !ok {
		t.Error("the loyal forward what an ally knows of")
	}
	if ok, _ := Forward(ForwardInput{AllyMet: true, Hostile: true, Reported: 2, Mil: 5, Dials: Dials{Loyalty: 0.5, Greed: 0.8}}, tn); ok {
		t.Error("the greedy withhold what serves their own designs")
	}
	if ok, _ := Forward(ForwardInput{AllyAtWar: true, Hostile: true, Reported: 2, Mil: 5, Dials: Dials{Loyalty: 0.9, Greed: 0.8}}, tn); !ok {
		t.Error("an ally at war is told by the loyal even so")
	}
	u := Turn(TurnInput{Honour: Faithless, Posture: Opportunist, HostMil: 2, Mil: 6, Greed: 0.5}, tn)
	if !near(u.Rate, 0.05*2*1) {
		t.Errorf("faithless opportunist with an opening: %s", u.Why())
	}
	if u := Turn(TurnInput{Honour: Faithless, Posture: Opportunist, HostMil: 12, Mil: 6, Greed: 0.5}, tn); u.Rate != 0 {
		t.Errorf("no opening: %s", u.Why())
	}
	if u := Turn(TurnInput{Honour: Faithless, Posture: Vengeful, HostMil: 2, Mil: 6, Greed: 0.5}, tn); u.Rate != 0 {
		t.Errorf("the vengeful keep faith with the faithful: %s", u.Why())
	}
	if u := Turn(TurnInput{Honour: Faithless, Posture: Vengeful, Betrayed: true, HostMil: 2, Mil: 6, Greed: 0.5}, tn); !near(u.Rate, 0.05*3) {
		t.Errorf("and turn on a betrayer: %s", u.Why())
	}
}

func TestSurveyAndSight(t *testing.T) {
	tn := Default()
	yes, no := func() bool { return true }, func() bool { return false }
	in := SurveyInput{Free: true, Reach: 10, Mil: 5, Dials: Dials{Hunger: 0.8, Greed: 0.6}, ToSettle: yes, Unread: yes, NeverRead: yes}
	if s := Survey(in, tn); s.Want != 2 {
		t.Errorf("hunger and greed: %s", s.Why())
	}
	in.AtWar = true
	if s := Survey(in, tn); s.Want != 0 {
		t.Errorf("wartime: %s", s.Why())
	}
	in.AtWar, in.NeverRead = false, no
	if s := Survey(in, tn); s.Want != 1 {
		t.Errorf("stale only: %s", s.Why())
	}
	in.NeverRead, in.Mil = yes, 2.5
	if s := Survey(in, tn); s.Want != 1 {
		t.Errorf("a level stays home: %s", s.Why())
	}
	in = SurveyInput{Free: true, Reach: 10, Mil: 5, Era: 2, Dials: Dials{Hunger: 0.1, Greed: 0.1}, ToSettle: no, Unread: yes, NeverRead: yes}
	if s := Survey(in, tn); s.Want != 1 {
		t.Errorf("necessity: %s", s.Why())
	}
	if got := SurveyTarget([]Star{{ID: 1, Charted: true, Fresh: true}, {ID: 2, Charted: true}, {ID: 3}, {ID: 4, Marked: true}}); got != 4 {
		t.Errorf("a marked star first: %d", got)
	}
	if got := SurveyTarget([]Star{{ID: 1, Charted: true, Fresh: true}, {ID: 2, Charted: true}, {ID: 3}}); got != 3 {
		t.Errorf("the nearest never read: %d", got)
	}
	if got := SurveyTarget([]Star{{ID: 1, Charted: true, Fresh: true}, {ID: 2, Charted: true}}); got != 2 {
		t.Errorf("else the stale: %d", got)
	}
	if got := SurveyTarget(nil); got != -1 {
		t.Errorf("nothing: %d", got)
	}
	if m := Sight(SightInput{Fear: 0.9, Menaced: yes, Unread: yes}, tn); m.Outward || !m.Threat {
		t.Errorf("the fearful with a hostile neighbour: %s", m.Why())
	}
	if m := Sight(SightInput{Fear: 0.3, Menaced: yes, Unread: yes}, tn); !m.Outward {
		t.Errorf("the bold ignore it: %s", m.Why())
	}
	if m := Sight(SightInput{Fear: 0.3, Menaced: no, Unread: no}, tn); m.Outward || m.Threat {
		t.Errorf("nothing left to read: %s", m.Why())
	}
}

func TestFindAndChoose(t *testing.T) {
	tn := Default()
	a := Find(FindInput{Curious: true}, tn)
	if a.Master != 4 || a.Wield != 1.5 || a.Seal != 1 {
		t.Errorf("curious: %s", a.Why())
	}
	if a := Find(FindInput{Pragmatic: true, Threat: true}, tn); a.Wield != 0 || a.Seal != 3 {
		t.Errorf("a threat is never wielded: %s", a.Why())
	}
	if a := Find(FindInput{Cautious: true, Own: true}, tn); a.Seal != 0 || a.Master != 4 {
		t.Errorf("one's own work: %s", a.Why())
	}
	if a := Find(FindInput{Law: true}, tn); a.Master != 0 || a.Seal != 0 || a.Wield != 1.5 {
		t.Errorf("a law: %s", a.Why())
	}
	a = Attempt{1, 2, 1}
	if a.Pick(0.1) != 0 || a.Pick(0.5) != 1 || a.Pick(0.9) != 2 {
		t.Error("the pick by the roll")
	}
	r := rand.New(rand.NewPCG(1, 2))
	open := []Pursuit{{Weight: 1, Domain: 1, Focus: 1, Aptitude: 1}, {Weight: 1, Domain: 1, Focus: 3, Aptitude: 1}}
	counts := [2]int{}
	for range 10_000 {
		i, _ := Choose(open, r, tn)
		counts[i]++
	}
	if counts[1] < 7000 || counts[1] > 8000 {
		t.Errorf("focus three against one: %v, want about three in four", counts)
	}
	if i, _ := Choose(nil, r, tn); i != -1 {
		t.Error("nothing open")
	}
	// a node the spare cannot feed is pursued a quarter as often, not never
	open = []Pursuit{{Weight: 1, Domain: 1, Focus: 1, Aptitude: 1, Unfed: true}, {Weight: 1, Domain: 1, Focus: 1, Aptitude: 1}}
	counts = [2]int{}
	for range 10_000 {
		i, _ := Choose(open, r, tn)
		counts[i]++
	}
	if counts[0] < 1500 || counts[0] > 2500 {
		t.Errorf("unfed against fed: %v, want about one in five", counts)
	}
}

func TestExpandAndRoam(t *testing.T) {
	tn := Default()
	yes := func() bool { return true }
	e := Expand(ExpandInput{Systems: 3, Mul: 1.5, Era: 2, Reach: 30, Nowhere: yes}, tn)
	if !near(e.Rate, 0.18) || !e.Ships || e.Hop != 20 || e.Reach != 30 {
		t.Errorf("%+v", e)
	}
	if e := Expand(ExpandInput{Systems: 20, Mul: 1, Era: 2, Reach: 50, Nowhere: yes, FTL: true, Parasite: true}, tn); e.Rate != 0.3 || e.Ships || e.Reach != 15 || e.Hop != 15 {
		t.Errorf("%+v", e)
	}
	r := rand.New(rand.NewPCG(3, 4))
	stars := []Colony{{ID: 1, Guess: true}, {ID: 2, Read: true}, {ID: 3, Read: true, Livable: true}}
	if s, blind := Target(stars, r, tn); s != 3 || blind {
		t.Errorf("the nearest livable read star: %d %v", s, blind)
	}
	n := 0
	for range 10_000 {
		if s, blind := Target(stars[:2], r, tn); s == 1 && blind {
			n++
		} else if s != -1 {
			t.Fatalf("a ship to %d", s)
		}
	}
	if n < 1800 || n > 2200 {
		t.Errorf("guesses %d in ten thousand, want about a fifth", n)
	}
	if h := RoamHop(1, tn); h != 3 {
		t.Errorf("hop %.0f", h)
	}
	if h := RoamHop(50, tn); h != 20 {
		t.Errorf("hop %.0f", h)
	}
	ports := []Port{{ID: 1, Held: true, Good: true}, {ID: 2}, {ID: 3, Good: true}}
	if s := NextStar(ports, r); s != 3 {
		t.Errorf("the good star: %d", s)
	}
	if s := NextStar(ports[:2], r); s != 2 {
		t.Errorf("else any: %d", s)
	}
	if s := NextStar(ports[:1], r); s != -1 {
		t.Errorf("nowhere: %d", s)
	}
}

// TestTuning: Set changes a field by name; unknown keys and groups error.
func TestTuning(t *testing.T) {
	tn := Default()
	if err := tn.Set("Council.Cadence=0.5"); err != nil || tn.Council.Cadence != 0.5 {
		t.Errorf("set: %v, cadence %v", err, tn.Council.Cadence)
	}
	if err := tn.Set("Survey.MaxTour=9"); err != nil || tn.Survey.MaxTour != 9 {
		t.Errorf("set an int: %v, %d", err, tn.Survey.MaxTour)
	}
	for _, bad := range []string{"Council.Nope=1", "Nope.Cadence=1", "Council=1", "Council.Cadence", "Council.Cadence=x"} {
		if err := tn.Set(bad); err == nil {
			t.Errorf("%q: no error", bad)
		}
	}
	got, err := Configure("", []string{"Bar.Conqueror=0.2", "Find.Own=9"})
	if err != nil || got.Bar.Conqueror != 0.2 || got.Find.Own != 9 || got.Council.Cadence != 0.3 {
		t.Errorf("configure: %v %+v", err, got.Bar)
	}
	if _, err := Configure("/nonexistent/tuning.json", nil); err == nil {
		t.Error("a missing file errors")
	}
}

// TestBuild: the structure that meets most of the want; with nothing
// wanting, the one worth most; nothing the spare cannot cover twice over.
func TestBuild(t *testing.T) {
	tn := Default()
	sites := []Site{
		{Key: "defences", Star: 0, Upkeep: flow.Income{0, 1, 1}, Levels: 1.5},
		{Key: "mine", Star: 0, Yield: flow.Income{0, 0, 4}, Upkeep: flow.Income{0, 1, 0}},
		{Key: "collectors", Star: 1, Yield: flow.Income{0, 3, 0}, Upkeep: flow.Income{0, 0, 1}},
	}
	// short of metal: the mine
	b := Build(BuildInput{Sites: sites, Want: flow.Income{0, 0, 3}, Spare: flow.Income{0, 4, 4}}, tn)
	if b.Pick != 1 || !b.Wanted {
		t.Errorf("short of metal: picked %d (%v), want the mine", b.Pick, b)
	}
	// short of energy: the collectors
	b = Build(BuildInput{Sites: sites, Want: flow.Income{0, 2, 0}, Spare: flow.Income{0, 4, 4}}, tn)
	if b.Pick != 2 {
		t.Errorf("short of energy: picked %d, want the collectors", b.Pick)
	}
	// nothing wanting: the mine is worth most (4 against 3 and 1.5 levels at 2 each)
	b = Build(BuildInput{Sites: sites, Spare: flow.Income{0, 4, 4}}, tn)
	if b.Pick != 1 || b.Wanted {
		t.Errorf("nothing wanting: picked %d (%v), want the mine", b.Pick, b)
	}
	// the spare must cover twice the upkeep: with 1 E only the collectors go
	b = Build(BuildInput{Sites: sites, Want: flow.Income{0, 0, 3}, Spare: flow.Income{0, 1, 4}}, tn)
	if b.Pick != 2 {
		t.Errorf("1 E spare: picked %d, want the collectors, the only thing covered twice", b.Pick)
	}
	b = Build(BuildInput{Sites: sites, Spare: flow.Income{0, 1, 1}}, tn)
	if b.Pick != -1 {
		t.Errorf("no spare: picked %d, want nothing", b.Pick)
	}
}

// TestPrize: a prize lowers the bar, to a cap, and never below zero.
func TestPrize(t *testing.T) {
	tn := Default()
	in := AppraiseInput{Strength: 4, Believed: 3, Spread: 0.3, Risk: 0.5}
	plain := Appraise(in, tn)
	in.Prize = 5
	rich := Appraise(in, tn)
	if plain.Prize != 0 || !near(rich.Prize, 0.1) {
		t.Errorf("prize off the bar: plain %.3f rich %.3f, want 0 and 0.1", plain.Prize, rich.Prize)
	}
	in.Prize = 100
	if capped := Appraise(in, tn); !near(capped.Prize, tn.Appraise.PrizeMax) {
		t.Errorf("a prize of 100 takes %.3f off, want the cap %.3f", capped.Prize, tn.Appraise.PrizeMax)
	}
	v := Judge(JudgeInput{Appraisal: rich, Bar: 0.5, Front: 1}, tn)
	if !near(v.Bar, 0.4) {
		t.Errorf("bar with the prize %.3f, want 0.4", v.Bar)
	}
	if v = Judge(JudgeInput{Appraisal: rich, Bar: 0.05, Front: 1}, tn); v.Bar != 0 {
		t.Errorf("bar below zero: %.3f", v.Bar)
	}
}

// TestTrade: the will and the road. A monster or a grudge gets nothing; a
// xenophobe sends half to the different; fear withholds metal from a
// stronger, hostile partner; the cap follows the drive, halves for a
// nomad, and is nothing out of reach.
func TestTrade(t *testing.T) {
	tn := Default()
	if c := Trade(TradeInput{Monster: true, InReach: true}, tn); !c.Refuse {
		t.Error("sent to a monster")
	}
	if c := Trade(TradeInput{Grudge: 0.5, InReach: true}, tn); !c.Refuse {
		t.Error("sent over a grudge")
	}
	if c := Trade(TradeInput{Grudge: 0.2, InReach: true}, tn); c.Refuse || c.Cap != 0.25 {
		t.Errorf("a small grudge: %s", c.Why())
	}
	if c := Trade(TradeInput{Xenophobe: true, Different: true, InReach: true, Drive: 1}, tn); c.Mul != [3]float64{0.5, 0.5, 0.5} || c.Cap != 0.5 {
		t.Errorf("a xenophobe to the different: %s %v", c.Why(), c.Mul)
	}
	if c := Trade(TradeInput{Fear: 0.8, Stronger: true, Hostile: true, InReach: true, Drive: 2}, tn); c.Mul != [3]float64{1, 1, 0} || c.Cap != 1 {
		t.Errorf("fear withholds metal: %s %v", c.Why(), c.Mul)
	}
	if c := Trade(TradeInput{Fear: 0.8, Stronger: true, InReach: true}, tn); c.Mul[2] != 1 {
		t.Error("fear withheld metal from a people that does not strike first")
	}
	if c := Trade(TradeInput{InReach: true, Nomad: true, Drive: 2}, tn); c.Cap != 0.5 {
		t.Errorf("a nomad partner: %s", c.Why())
	}
	if c := Trade(TradeInput{Drive: 2}, tn); c.Cap != 0 || c.Refuse {
		t.Errorf("out of reach: %s", c.Why())
	}
}
