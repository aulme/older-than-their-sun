package mind

import "fmt"

// Contracts: a timed payment for a discrete thing. The world prices each
// term for each side (what it is worth to the one that gets it, what it
// costs the one that gives it); the mind says whether the asked accepts,
// at what margin, how often a people asks, how long it pays, whether a
// hired fleet lets itself be bought off, and whether a sighting is for
// sale by honour.

// TermKind is what one side of a contract gives.
type TermKind uint8

const (
	TermFlow     TermKind = iota // so much of a kind per tick
	TermRarity                   // a mobile rarity handed over
	TermAccess                   // an immobile rarity's grants and levels counted as the receiver's
	TermTeach                    // a node passed as a message
	TermGuard                    // a fleet standing at a star
	TermStrike                   // a fleet against a world or a work
	TermDeliver                  // a fleet that takes a world and hands it over
	TermPeace                    // no war before the end
	TermSighting                 // a fleet seen in flight, passed on
	TermBroker                   // spoken for, to a people not understood
)

func (k TermKind) String() string {
	return [...]string{"flow", "rarity", "access", "teach", "guard", "strike", "deliver", "peace", "sighting", "broker"}[k]
}

// Lasting says whether a term runs for a length rather than being done once.
func (k TermKind) Lasting() bool {
	switch k {
	case TermFlow, TermAccess, TermGuard, TermPeace:
		return true
	}
	return false
}

// OfferInput is an offer as the asked side sees it: what it would give
// and what it would get, each already priced by the world for it.
type OfferInput struct {
	Gives, Gets TermKind
	GiveWorth   float64 // what the term it gives costs it
	GetWorth    float64 // what the term it gets is worth to it
	Greed       float64
	Fixation    string // the morality's object, or ""
	Crime       bool   // its morality counts what it is asked to give a crime
	Xenophobe   bool
	Different   bool // the other is different to it
	Ignores     bool // a herd or amoral morality: who asks does not matter
	Wis, Noise  float64
}

// Deal is the verdict on an offer.
type Deal struct {
	Accept bool
	Margin float64
	Priced float64 // what it gets, read through its folly
	Reason string
}

// Why says the verdict.
func (a Deal) Why() string {
	if a.Accept {
		return fmt.Sprintf("accepts at a margin of %.1f: gets %.1f against the asking", a.Margin, a.Priced)
	}
	return fmt.Sprintf("refuses at a margin of %.1f: %s", a.Margin, a.Reason)
}

// Margin is what a people wants over the cost of what it gives: one for
// the faithful and the pragmatic, more for the greedy, less for a
// fixation that wants the thing anyway, double for a crime, double for a
// xenophobe against the different. Nothing gets a margin of its own by
// kind or way.
func Margin(in OfferInput, t *Tuning) float64 {
	p := &t.Contract
	m := 1.0
	if in.Greed > p.GreedBar {
		m = p.GreedMargin
	}
	if in.Fixation == "conquest" && in.Gives == TermStrike {
		m = p.EagerMargin
	}
	if in.Fixation == "knowing" && (in.Gives == TermTeach || in.Gets == TermTeach) {
		m = p.EagerMargin
	}
	if in.Crime {
		m *= p.CrimeMargin
	}
	if in.Xenophobe && in.Different && !in.Ignores {
		m *= p.XenoMargin
	}
	return m
}

// AnswerOffer: the asked accepts when what it gets, read through its
// folly, covers what it gives at its margin, and there is something in
// it at all.
func AnswerOffer(in OfferInput, t *Tuning) Deal {
	a := Deal{Margin: Margin(in, t)}
	a.Priced = Folly(in.GetWorth, in.Wis, in.Noise, t)
	switch {
	case in.GetWorth <= 0:
		a.Reason = "nothing in it"
	case a.Priced < in.GiveWorth*a.Margin:
		a.Reason = fmt.Sprintf("gets %.1f against %.1f asked", a.Priced, in.GiveWorth*a.Margin)
	default:
		a.Accept = true
	}
	return a
}

// AskRate is how often a people looks for something to buy, per thousand
// years: more for the greedy and for a fixation on holding.
func AskRate(greed float64, holding bool, t *Tuning) float64 {
	if greed > t.Contract.GreedBar || holding {
		return t.Contract.AskRateGreedy
	}
	return t.Contract.AskRate
}

// PayLength is how long a flow is paid for a thing of so many units: five
// thousand years a unit, within bounds.
func PayLength(units float64, t *Tuning) float64 {
	p := &t.Contract
	return max(p.LengthMin, min(p.LengthMax, p.LengthPerUnit*units))
}

// MercenaryRate is how often a people offers its idle fleet for hire:
// the base times its want share times its spare fleet share, so a people
// with no want or no idle fleet never offers.
func MercenaryRate(wantShare, spareShare float64, t *Tuning) float64 {
	return t.Contract.MercenaryRate * max(0, min(1, wantShare)) * max(0, min(1, spareShare))
}

// BuyOffInput is a hired fleet's people weighing a better offer from the
// one it stands against.
type BuyOffInput struct {
	Honour     string
	Greed      float64
	PayFailed  bool    // the employer's pay has failed once already
	NewWorth   float64 // what the new payer offers, to it
	OldWorth   float64 // what the employer pays, to it
	Margin     float64 // its own
	Wis, Noise float64
}

// BuyOff is whether the fleet is bought: never the faithful; the
// practical when the employer has already failed it; the faithless at a
// chance by greed. In every case the new pay must beat the old by the
// margin.
type BuyOff struct {
	Rate   float64 // per thousand years; 0 never, 1 at once
	Reason string
}

// Why says the answer.
func (b BuyOff) Why() string {
	if b.Rate > 0 {
		return fmt.Sprintf("may be bought at %.2f a thousand years", b.Rate)
	}
	return "keeps its word: " + b.Reason
}

// BuyOffRate decides.
func BuyOffRate(in BuyOffInput, t *Tuning) BuyOff {
	if Folly(in.NewWorth, in.Wis, in.Noise, t) < in.OldWorth*in.Margin {
		return BuyOff{Reason: "the new pay does not beat the old"}
	}
	switch in.Honour {
	case Faithful:
		return BuyOff{Reason: "faithful"}
	case Faithless:
		return BuyOff{Rate: t.Contract.FaithlessBuyOff * in.Greed}
	}
	if in.PayFailed {
		return BuyOff{Rate: 1}
	}
	return BuyOff{Reason: "the employer has not failed them yet"}
}

// SellSighting is whose fleets are for sale, by honour: the faithful
// sell no pact member's fleet and none bound to stand with one; the
// practical no pact member's; the faithless anyone's.
func SellSighting(honour string, ownerAllied, boundToAlly bool) bool {
	switch honour {
	case Faithful:
		return !ownerAllied && !boundToAlly
	case Faithless:
		return true
	}
	return !ownerAllied
}

// TributeInput is a winner deciding between worlds and tribute.
type TributeInput struct {
	Posture  string
	Fixation string
}

// TakesTribute: a defensive, confederate or opportunist winner, or one
// fixed on holding, takes tribute over worlds; a conqueror takes worlds.
func TakesTribute(in TributeInput) bool {
	switch in.Posture {
	case Defensive, Confederate, Opportunist:
		return true
	}
	return in.Fixation == "holding"
}
