package mind

import "testing"

// TestWarCouncil: on fixed inputs, a side that believes it is winning
// presses, one that fears it sues, one whose aim is met offers peace on
// the lines, and one with the odds and not the ships holds; the side
// declared on presses to retake at a lower bar; the unyielding never sue.
func TestWarCouncil(t *testing.T) {
	tn := Default()
	base := WarInput{Aim: AimWorld, Declarer: true, Posture: Opportunist, Acted: 0.8, Reach: true, Ready: true, Will: 1, Fear: 0.5, Treats: true}
	cases := []struct {
		name string
		mod  func(*WarInput)
		want WarChoice
	}{
		{"winning presses", func(in *WarInput) {}, Press},
		{"a campaign out holds", func(in *WarInput) { in.Out = true }, Hold},
		{"the odds and not the ships holds", func(in *WarInput) { in.Ready = false }, Hold},
		{"afraid after a battle sues", func(in *WarInput) { in.Acted, in.Fought = 0.05, true }, Sue},
		{"afraid before any fighting holds", func(in *WarInput) { in.Acted = 0.05 }, Hold},
		{"the aim met sues for the lines", func(in *WarInput) { in.Met = true }, Sue},
		{"the will nearly gone sues", func(in *WarInput) { in.Will = 0.1 }, Sue},
		{"idle and not pressing sues", func(in *WarInput) { in.Acted, in.Idle = 0.3, 4 }, Sue},
		{"idle, unyielding, holds", func(in *WarInput) { in.Acted, in.Idle, in.Posture = 0.3, 4, Unyielding }, Hold},
		{"the unyielding do not sue when spent", func(in *WarInput) { in.Posture, in.Will = Unyielding, 0.1 }, Press},
		{"no terms between them: no suing", func(in *WarInput) { in.Treats, in.Will = false, 0.1 }, Press},
		{"a pacifist never presses", func(in *WarInput) { in.Posture = Pacifist }, Hold},
		{"the side declared on retakes at a lower bar", func(in *WarInput) {
			in.Aim, in.Declarer, in.Posture, in.Acted, in.Lost = AimHold, false, Defensive, PressBarOf(Defensive, AimHold, 1)+0.01, 1
		}, Press},
		{"and holds at the same odds with nothing lost", func(in *WarInput) {
			in.Aim, in.Declarer, in.Posture, in.Acted = AimHold, false, Defensive, PressBarOf(Defensive, AimHold, 1)+0.01
		}, Hold},
	}
	for _, c := range cases {
		in := base
		c.mod(&in)
		if v := WarCouncil(in, tn); v.Choice != c.want {
			t.Errorf("%s: %s, want %s", c.name, v.Why(), c.want)
		}
	}
}

// PressBarOf is the bar a posture presses at for an aim, having lost so
// many worlds and taken none.
func PressBarOf(posture, aim string, lost int) float64 {
	b, _ := PressBar(WarInput{Posture: posture, Aim: aim, Lost: lost}, Default())
	return b
}

// TestAnswerTerms: terms that meet the aim are taken unless the side is
// winning big and greedy; terms short of it are refused by a side that
// believes it can take the rest and would press, and taken by one that
// is tired or would not; the hating take nothing while they can go on.
func TestAnswerTerms(t *testing.T) {
	tn := Default()
	cases := []struct {
		name string
		in   TermsInput
		want bool
	}{
		{"the aim, at even odds", TermsInput{Worth: 1, Acted: 0.5, Will: 1, Posture: Defensive, Presses: true}, true},
		{"half the aim, winning and pressing", TermsInput{Worth: 0.5, Acted: 0.9, Will: 1, Posture: Opportunist, Presses: true}, false},
		{"half the aim, winning and not pressing", TermsInput{Worth: 0.5, Acted: 0.9, Will: 1, Posture: Opportunist}, true},
		{"half the aim, winning and tired", TermsInput{Worth: 0.5, Acted: 0.9, Will: 0, Posture: Opportunist, Presses: true}, true},
		{"the hating, anything", TermsInput{Worth: 1, Acted: 0.1, Will: 1, Posture: Defensive, Hates: true}, false},
		{"the unyielding, short of the aim", TermsInput{Worth: 0.9, Acted: 0, Will: 0, Posture: Unyielding}, false},
	}
	for _, c := range cases {
		if a := AnswerTerms(c.in, tn); a.Accept != c.want {
			t.Errorf("%s: %s", c.name, a.Why())
		}
	}
}

// TestAnswerYoke: a small people bends when the war it would fight is
// believed lost, sooner for fear and for the submissive; the unyielding,
// the conqueror and the hating do not.
func TestAnswerYoke(t *testing.T) {
	tn := Default()
	cases := []struct {
		in   YokeInput
		want bool
	}{
		{YokeInput{Acted: 0.1, Posture: Defensive}, true},
		{YokeInput{Acted: 0.5, Posture: Defensive}, false},
		{YokeInput{Acted: 0.4, Posture: Defensive, Fear: 1}, true},
		{YokeInput{Acted: 0.45, Posture: Submissive}, true},
		{YokeInput{Acted: 0.1, Posture: Unyielding}, false},
		{YokeInput{Acted: 0.1, Posture: Conqueror}, false},
		{YokeInput{Acted: 0.1, Posture: Defensive, Hates: true}, false},
	}
	for _, c := range cases {
		if a := AnswerYoke(c.in, tn); a.Accept != c.want {
			t.Errorf("%+v: %s", c.in, a.Why())
		}
	}
}

// TestEscalate: between old enemies with a grudge standing, the second
// war is for redress and the third for the whole when the declarer is
// twice the other's size, redress again when it is not; without a
// grudge, or for a total aim or an ally's, the aim stays.
func TestEscalate(t *testing.T) {
	tn := Default()
	cases := []struct {
		aim    string
		nth    int
		grudge bool
		size   float64
		want   string
	}{
		{AimWorld, 1, true, 3, AimWorld},
		{AimWorld, 2, true, 3, AimRedress},
		{AimTribute, 2, true, 1, AimRedress},
		{AimWorld, 3, true, 3, AimSubmission},
		{AimRedress, 5, true, 2, AimSubmission},
		{AimWorld, 3, true, 1.5, AimRedress}, // rivals of a size: redress again
		{AimWorld, 3, false, 3, AimWorld},
		{AimEnding, 3, true, 3, AimEnding},
		{AimDefence, 3, true, 3, AimDefence},
	}
	for _, c := range cases {
		if got := Escalate(c.aim, c.nth, c.grudge, c.size, tn); got != c.want {
			t.Errorf("%s, war %d, grudge %v, size %.1f: got %s, want %s", c.aim, c.nth, c.grudge, c.size, got, c.want)
		}
	}
}

// TestAllyStays: an ally that holds its alliance dear stays in a war it
// would leave on its own account, and one whose principal broke faith
// with it does not; and it weighs peace alone as the betrayal it is.
func TestAllyStays(t *testing.T) {
	tn := Default()
	dear := Bound(BoundInput{Age: 200_000, Members: 5, Menace: true, Renown: 1}, tn)
	cheap := Bound(BoundInput{Age: 0, Members: 2, Betrayed: true}, tn)
	if dear < 0.9 || cheap > 0 {
		t.Fatalf("worth: dear %.2f, cheap %.2f", dear, cheap)
	}
	in := WarInput{Aim: AimDefence, Declarer: true, Posture: Defensive, Acted: 0.05, Reach: true, Fought: true, Will: 0.5, Treats: true}
	if v := WarCouncil(in, tn); v.Choice != Sue {
		t.Errorf("an ally with nothing to lose by leaving, afraid: %s", v.Why())
	}
	in.Bound = dear
	if v := WarCouncil(in, tn); v.Choice == Sue {
		t.Errorf("an ally holding its alliance dear, as afraid: %s", v.Why())
	}
	terms := TermsInput{Worth: 0.5, Acted: 0.3, Will: 1, Posture: Defensive}
	if a := AnswerTerms(terms, tn); !a.Accept {
		t.Errorf("peace offered an ally that would not press: %s", a.Why())
	}
	terms.Bound = dear
	if a := AnswerTerms(terms, tn); a.Accept {
		t.Errorf("and to one that holds the alliance dear: %s", a.Why())
	}
}
