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
