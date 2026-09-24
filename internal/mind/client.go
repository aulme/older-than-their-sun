package mind

import (
	"fmt"
	"math"
)

// Clients and the cold war (specs/proposals/war.md, stage 3). A vassal
// is an actor on a leash: it holds its own council, and when it is
// struck it calls on its patron, whose council joins the war, sends
// ships to stand with it, or stays out. Every council weighing an attack
// on a client reckons the patron with it, by the patron's record of
// coming. The vassal pays for the protection: a standing tribute, a
// share of its whole income at a rate the bond sets and the patron
// reviews, which it pays first or pays short. Two powers that menace
// each other build fleets against each other: the deterrence term in the
// want. The history gathers and executes.

// ClientTuning: the leash, the patron's answer, the guarantee, the arms
// race and the tribute.
type ClientTuning struct {
	Deter         float64 // ships built toward the believed fleet of the strongest menacing neighbour, per point of fear
	DeterMax      float64 // at most this share of that fleet
	Join          float64 // the patron joins its client's war at these odds against the attacker, or better
	Base          float64 // the patron's will to answer at all, before its loyalty, fear and the client's worth
	FearWeight    float64 // off it per point of fear
	WorthWeight   float64 // on it per point of what the client is worth to it
	Back          float64 // the will above which it sends ships when it will not join
	Offence       float64 // the grudge an attack on its client gives the patron, whatever it does
	Abandoned     float64 // the betrayal's weight when the patron stays out
	Guarantee     float64 // the share of a patron's fleet an attacker reckons standing with its client, at a perfect record of coming
	ComePrior     float64 // calls counted as half answered before the patron's first
	RateMin       float64 // the tribute's rate, as a share of the vassal's whole income: at least
	RateMax       float64 // at most
	RateBase      float64 // to begin with
	RateGreed     float64 // more per point of the patron's greed
	RateConqueror float64 // more from a conqueror
	RateDifferent float64 // more from a patron that finds the vassal different
	RateSurrender float64 // more for a bond made in defeat
	RateSought    float64 // less for a people that sought its patron
	RateHopeless  float64 // more per point the vassal's odds fell short of holding its own
	RateNoise     float64 // the spread of the bond's rate
	Review        float64 // the share of the way to its target a rate moves at a review
	Lighten       float64 // off the target for a vassal that answered, fought beside its patron or paid through lean years
	Raise         float64 // on it for one that paid short, refused a call or grew strong enough to worry it
	Distress      float64 // off it for a vassal whose own fields went unfed, from a patron no fool
	PayFirst      float64 // the vassal pays first when its reasons read above this
	Anger         float64 // the patron's grudge per share of a tick's tribute missing
	Punish        float64 // the patron's grudge against its vassal past which it makes war on it
	ReviewEvery   float64 // years between a bond's reviews
	Wise          float64 // the Wisdom from which a patron does not kill the goose
}

// Deterrence is the ships a people builds toward against its rival: a
// share of the rival's believed fleet, growing with fear.
func Deterrence(fear, believed float64, t *Tuning) int {
	p := &t.Client
	return int(math.Round(min(p.DeterMax, p.Deter*fear) * believed))
}

// ClientChoice is what a patron does when its client is struck.
type ClientChoice uint8

const (
	StayOut ClientChoice = iota
	Back                 // send ships to stand with the client
	Join                 // go to war with the attacker
)

func (c ClientChoice) String() string { return [...]string{"stays out", "sends ships", "joins"}[c] }

// ClientInput is what a patron reads when its client calls.
type ClientInput struct {
	Posture string
	Acted   float64 // its odds acted on against the attacker
	Worth   float64 // what the client is worth to it: its tribute and its faith, 0 to 1
	Dials   Dials
	Spare   int  // ships it could send and keep its home safe
	Reach   bool // it could strike the attacker
}

// ClientAnswer is the answer.
type ClientAnswer struct {
	Choice ClientChoice
	Will   float64
}

// Why says it.
func (a ClientAnswer) Why() string {
	return fmt.Sprintf("%s: will %.2f", a.Choice, a.Will)
}

// AnswerClient is the patron's council on its client's call: its will to
// answer is its loyalty against its fear and what the client is worth to
// it; with the will and the odds against the attacker it joins the war;
// with the will and not the odds — a power it would not face directly —
// it sends ships to stand with the client, which is how a proxy war is
// fought; without the will it stays out, and pays for it in renown.
func AnswerClient(in ClientInput, t *Tuning) ClientAnswer {
	p := &t.Client
	a := ClientAnswer{Will: p.Base + in.Dials.Loyalty - p.FearWeight*in.Dials.Fear + p.WorthWeight*in.Worth}
	switch {
	case in.Posture == Pacifist && in.Spare > 0 && a.Will > p.Back:
		a.Choice = Back // a pacifist stands with its client and strikes nobody
	case a.Will > 0 && in.Reach && in.Acted >= p.Join && in.Posture != Pacifist:
		a.Choice = Join
	case a.Will > p.Back && in.Spare > 0:
		a.Choice = Back
	}
	return a
}

// Guarantee is what an attacker reckons of a client's patron: a share of
// the patron's fleet as it believes it, by the patron's record of
// coming, as ships that would stand with the client.
func Guarantee(ships, record float64, t *Tuning) float64 {
	return t.Client.Guarantee * ships * record
}

// ComeRate is a patron's record of answering its clients: the calls it
// came to (joining or sending ships) of those it heard, with half a call
// answered before the first.
func ComeRate(came, called int, t *Tuning) float64 {
	return (float64(came) + t.Client.ComePrior) / (float64(called) + 2*t.Client.ComePrior)
}

// RateInput is a bond being made: the patron's character, how it came
// about, and how hopeless the vassal's position was.
type RateInput struct {
	Greed     float64
	Conqueror bool
	Different bool
	How       string  // surrender (at the end of a war), offer (vassalage accepted without one), sought (the vassal came of itself)
	Hopeless  float64 // how far the vassal's odds of holding its own fell short of even, 0 to 1
	Noise     float64 // a normal draw
}

// TributeRate is the share of its whole income a vassal pays its patron,
// set at the bond: higher from a greedy, conquering or estranged patron,
// highest on a surrender and lowest for a people that sought a patron,
// the more the more hopeless the vassal's position, and noise; bounded.
func TributeRate(in RateInput, t *Tuning) float64 {
	return clampRate(rateTarget(in, t)+t.Client.RateNoise*in.Noise, t)
}

func rateTarget(in RateInput, t *Tuning) float64 {
	p := &t.Client
	r := p.RateBase + p.RateGreed*in.Greed + p.RateHopeless*in.Hopeless
	if in.Conqueror {
		r += p.RateConqueror
	}
	if in.Different {
		r += p.RateDifferent
	}
	switch in.How {
	case "surrender":
		r += p.RateSurrender
	case "sought":
		r -= p.RateSought
	}
	return r
}

func clampRate(r float64, t *Tuning) float64 { return min(t.Client.RateMax, max(t.Client.RateMin, r)) }

// ReviewInput is a bond under review: its rate, the target the patron's
// character set, and how the vassal has done since the last.
type ReviewInput struct {
	Rate, Target float64
	Loyal        bool // it answered its patron's calls, fought beside it, or paid in full through lean years
	Short        bool // it paid short
	Refused      bool // it refused a call
	Strong       bool // it has grown strong enough to worry its patron
	Distress     bool // its own fields went unfed for the tribute
	Wise         bool // the patron is no fool: it does not kill the goose
}

// ReviewRate moves a bond's rate part of the way toward its target, the
// target lightened for a loyal vassal and for one in real distress under
// a patron no fool, and raised for one that paid short, refused or grew
// strong.
func ReviewRate(in ReviewInput, t *Tuning) float64 {
	p := &t.Client
	target := in.Target
	if in.Loyal {
		target -= p.Lighten
	}
	if in.Short || in.Refused || in.Strong {
		target += p.Raise
	}
	if in.Distress && in.Wise {
		target -= p.Distress
	}
	return clampRate(in.Rate+p.Review*(clampRate(target, t)-in.Rate), t)
}

// PayInput is a vassal deciding where its tribute stands in its flows.
type PayInput struct {
	Loyalty float64 // its dial
	Awe     float64 // how little it believes it could hold against its patron, 0 to 1
	Near    bool    // the patron's fleets reach its home
	Lean    bool    // its own fields went unfed last tick
}

// PaysFirst says whether a vassal pays its tribute before its own works,
// fleets and road, or puts it where its direction puts the word and
// risks paying short: loyalty, awe of its patron and the patron's
// nearness against the hard times.
func PaysFirst(in PayInput, t *Tuning) bool {
	s := 0.5*in.Loyalty + 0.5*in.Awe
	if in.Near {
		s += 0.2
	}
	if in.Lean {
		s -= 0.3
	}
	return s > t.Client.PayFirst
}

// defaultClient is the stage's numbers.
func defaultClient() ClientTuning {
	return ClientTuning{
		Deter: 0.3, DeterMax: 0.5,
		Join: 0.5, Base: 0.5, FearWeight: 0.5, WorthWeight: 0.5, Back: 0.1, Offence: 0.5, Abandoned: 0.5,
		Guarantee: 0.5, ComePrior: 0.5,
		RateMin: 0.05, RateMax: 0.35, RateBase: 0.05, RateGreed: 0.05, RateConqueror: 0.03, RateDifferent: 0.02,
		RateSurrender: 0.05, RateSought: 0.03, RateHopeless: 0.05, RateNoise: 0.02,
		Review: 0.3, Lighten: 0.03, Raise: 0.04, Distress: 0.05, PayFirst: 0.5, Anger: 0.05, Punish: 3, ReviewEvery: 10_000, Wise: 5,
	}
}
