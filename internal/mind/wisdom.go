package mind

import "fmt"

// Wisdom is the capacity to see the counterintuitive and make the right
// call when the obvious call is wrong. It touches how a belief is used,
// never what is believed or wanted: the tail of the appraisal, the
// grudge's discount, the conqueror's compulsion, the noise on a worth, the
// Find's choice, and whether a people that fathoms another lets itself be
// understood or speaks for a third.

// Folly reads a worth or a score through the noise of a people's
// judgment: a fool at two and a half misprices a bargain by a third one
// time in three, a people at ten prices it exactly. The noise is a
// standard normal draw made by the caller.
func Folly(worth, wis, noise float64, t *Tuning) float64 {
	return worth * (1 + noise*t.Wisdom.Folly*max(0, 10-wis))
}

// LeapInput is a people about to try its hand at a remain.
type LeapInput struct {
	Margin float64 // the expected margin on the roll it would make: its level less the difficulty
	Wis    float64
	Noise  float64 // a standard normal draw
}

// Leap says whether a people looks before it leaps: when the expected
// margin is poor and Wisdom with the roll clears the bar, it seals
// instead of trying.
func Leap(in LeapInput, t *Tuning) bool {
	p := &t.Wisdom
	return in.Margin < p.LeapMargin && in.Wis+in.Noise*p.LeapNoise >= p.LeapBar
}

// TeachInput is whether a people that fathoms another lets itself be
// understood in return.
type TeachInput struct {
	Hates   bool   // it finds the other too different to let live
	AtWar   bool   // with the other
	Posture string // its own
	Fixed   bool   // a morality fixed on conquest or holding
	Weaker  bool   // it reads the other as prey: weaker by the margin
}

// Teaching is the answer.
type Teaching struct {
	Teach  bool
	Reason string
}

// Why says the answer.
func (g Teaching) Why() string {
	if g.Teach {
		return "makes itself understood"
	}
	return "does not explain itself: " + g.Reason
}

// Teach: a people wants to be understood unless it hates the other, is
// at war with it, or strikes first and reads it as prey; a fixation on
// conquest or holding refuses prey the same way.
func Teach(in TeachInput, t *Tuning) Teaching {
	switch {
	case in.Hates:
		return Teaching{Reason: "too different to be let live"}
	case in.AtWar:
		return Teaching{Reason: "at war"}
	case (Hostile(in.Posture) || in.Fixed) && in.Weaker:
		return Teaching{Reason: "a people that strikes first does not explain itself to prey"}
	}
	return Teaching{Teach: true}
}

// BrokerInput is whether a people that knows two others speaks for one to
// the other for nothing.
type BrokerInput struct {
	SharedPact bool // it holds a pact with both
	Posture    string
	Xenophobic bool
	AtWarWith  bool // with the one to be understood
}

// Brokering is the answer.
type Brokering struct {
	Rate   float64 // attempts per thousand years
	Reason string
}

// Why says the answer.
func (b Brokering) Why() string {
	if b.Rate > 0 {
		return fmt.Sprintf("speaks for one to the other at %.2f a thousand years", b.Rate)
	}
	return "keeps its position: " + b.Reason
}

// Broker: allies that cannot speak to each other are no use in a war, so
// a people sworn to both speaks for one to the other, and a confederate
// does it by nature; a go-between that merely trades with both keeps the
// middle. Nobody brokers to a people it is at war with, and a xenophobe
// never speaks for anyone.
func Broker(in BrokerInput, t *Tuning) Brokering {
	switch {
	case in.Xenophobic:
		return Brokering{Reason: "a xenophobe speaks for no one"}
	case in.AtWarWith:
		return Brokering{Reason: "at war with the one to be understood"}
	case in.SharedPact || in.Posture == Confederate:
		return Brokering{Rate: t.Wisdom.BrokerRate}
	}
	return Brokering{Reason: "the middle is its position"}
}
