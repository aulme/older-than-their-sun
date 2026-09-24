package mind

import "testing"

// TestTributeRate: a surrender costs more than an offer, an offer more
// than a people that sought its patron; greed, conquest and hopelessness
// raise the rate; it stays in its bounds.
func TestTributeRate(t *testing.T) {
	tn := Default()
	base := RateInput{Greed: 0.5, Hopeless: 0.5}
	rate := func(how string, mod func(*RateInput)) float64 {
		in := base
		in.How = how
		if mod != nil {
			mod(&in)
		}
		return TributeRate(in, tn)
	}
	if !(rate("surrender", nil) > rate("offer", nil) && rate("offer", nil) > rate("sought", nil)) {
		t.Errorf("surrender %.3f, offer %.3f, sought %.3f", rate("surrender", nil), rate("offer", nil), rate("sought", nil))
	}
	if rate("offer", func(in *RateInput) { in.Conqueror, in.Greed, in.Hopeless = true, 1, 1 }) <= rate("offer", nil) {
		t.Error("a greedy conqueror asks no more of a hopeless people")
	}
	if r := rate("surrender", func(in *RateInput) { in.Noise = 50 }); r != tn.Client.RateMax {
		t.Errorf("the rate is out of bounds: %.3f", r)
	}
	if r := rate("sought", nil); r < 0.05 || r > 0.15 {
		t.Errorf("a sought bond at %.3f, want near a tenth", r)
	}
}

// TestReviewRate: a loyal vassal's rate is lightened, one that paid short
// raised, one in distress under a wise patron lightened.
func TestReviewRate(t *testing.T) {
	tn := Default()
	in := ReviewInput{Rate: 0.12, Target: 0.12}
	loyal, short := in, in
	loyal.Loyal, short.Short = true, true
	if ReviewRate(loyal, tn) >= 0.12 || ReviewRate(short, tn) <= 0.12 {
		t.Errorf("loyal %.3f, short %.3f from 0.12", ReviewRate(loyal, tn), ReviewRate(short, tn))
	}
	distress := in
	distress.Distress, distress.Wise = true, true
	if ReviewRate(distress, tn) >= 0.12 {
		t.Error("a wise patron squeezes a vassal in distress")
	}
	distress.Wise = false
	if ReviewRate(distress, tn) != 0.12 {
		t.Error("a foolish patron spares the goose")
	}
}

// TestAnswerClient: with the will and the odds a patron joins; with the
// will and not the odds it sends ships; without the will it stays out; a
// pacifist sends ships and strikes nobody.
func TestAnswerClient(t *testing.T) {
	tn := Default()
	loyal := Dials{Loyalty: 0.8, Fear: 0.2}
	cases := []struct {
		in   ClientInput
		want ClientChoice
	}{
		{ClientInput{Posture: Defensive, Acted: 0.8, Worth: 0.5, Dials: loyal, Spare: 3, Reach: true}, Join},
		{ClientInput{Posture: Defensive, Acted: 0.2, Worth: 0.5, Dials: loyal, Spare: 3, Reach: true}, Back},
		{ClientInput{Posture: Defensive, Acted: 0.8, Worth: 0, Dials: Dials{Loyalty: 0, Fear: 1}, Spare: 3, Reach: true}, StayOut},
		{ClientInput{Posture: Pacifist, Acted: 0.9, Worth: 0.5, Dials: loyal, Spare: 3, Reach: true}, Back},
		{ClientInput{Posture: Defensive, Acted: 0.2, Worth: 0.5, Dials: loyal, Spare: 0, Reach: true}, StayOut},
	}
	for _, c := range cases {
		if a := AnswerClient(c.in, tn); a.Choice != c.want {
			t.Errorf("%+v: %s, want %s", c.in, a.Why(), c.want)
		}
	}
}

// TestDeterrence: the arms race builds toward a share of the rival's
// fleet, more for the fearful, never more than the cap.
func TestDeterrence(t *testing.T) {
	tn := Default()
	if Deterrence(0, 20, tn) != 0 || Deterrence(0.5, 20, tn) <= 0 || Deterrence(10, 20, tn) != int(tn.Client.DeterMax*20+0.5) {
		t.Errorf("deterrence: fearless %d, fearful %d, capped %d", Deterrence(0, 20, tn), Deterrence(0.5, 20, tn), Deterrence(10, 20, tn))
	}
	if ComeRate(0, 0, tn) != 0.5 || ComeRate(4, 4, tn) <= 0.8 || ComeRate(0, 4, tn) >= 0.2 {
		t.Errorf("the record of coming: new %.2f, faithful %.2f, faithless %.2f", ComeRate(0, 0, tn), ComeRate(4, 4, tn), ComeRate(0, 4, tn))
	}
}
