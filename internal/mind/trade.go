package mind

import (
	"fmt"

	"worldgen/internal/flow"
)

// Trade is what a people is willing to send a partner, and how much can
// cross. The arithmetic of what goes is flow.Share; this is the will and
// the road.

// TradeInput is one partner as the sender sees it.
type TradeInput struct {
	Monster   bool    // the partner is remembered as a thing that does harm
	Grudge    float64 // what the partner has done to them
	Xenophobe bool    // the sender's drive
	Different bool    // the partner counts as different to it
	Fear      float64 // the sender's fear dial
	Stronger  bool    // the sender's intel reads the partner as stronger
	Hostile   bool    // the partner strikes first, by posture
	InReach   bool    // a holding of the sender is within its reach of a holding of the partner
	Nomad     bool    // either side lives as fleets
	Drive     int     // the sender's road: 0 slow, 1 sails or near-light, 2 the Door or wormholes
	Embargoed bool    // the partner has closed its ports to the sender
	Holding   bool    // the sender is fixed on holding what it has
	Spawning  bool    // the sender is fixed on more of itself
	Grasping  bool    // the partner is fixed on holding: its want is filled first
}

// TradeChoice is the answer: a cap on what crosses, a multiplier on each
// kind's want, and whether the sender refuses outright.
type TradeChoice struct {
	Cap     float64
	Organic float64    // the cap on organic matter, when it differs
	Mul     [3]float64 // by kind: organic matter, energy, metal
	Refuse  bool
	Holds   bool // refuses because it holds what it has: not a closing of ports, a way of being
	reason  string
}

// CapOf is the cap for a kind.
func (c TradeChoice) CapOf(k flow.Kind) float64 {
	if k == flow.O && c.Organic > 0 {
		return c.Organic
	}
	return c.Cap
}

// Why says the choice in a line.
func (c TradeChoice) Why() string {
	if c.Refuse {
		return "sends nothing: " + c.reason
	}
	s := fmt.Sprintf("sends up to %.0f%% of the spare", 100*c.Cap)
	if c.reason != "" {
		s += ", " + c.reason
	}
	return s
}

// Trade decides. A sender refuses a monster or a grudge above the bar;
// sends half to a people its xenophobia counts as different; a fearful
// people sends no metal to a partner it reads as stronger and hostile. The
// cap is the drive's, half for a nomad on either side, nothing out of reach.
func Trade(in TradeInput, t *Tuning) TradeChoice {
	p := &t.Trade
	c := TradeChoice{Mul: [3]float64{1, 1, 1}}
	switch {
	case in.Monster:
		return TradeChoice{Refuse: true, reason: "they are remembered as monsters"}
	case in.Grudge > p.GrudgeBar:
		return TradeChoice{Refuse: true, reason: fmt.Sprintf("a grudge of %.1f is held against them", in.Grudge)}
	case in.Holding:
		return TradeChoice{Refuse: true, Holds: true, reason: "they hold what they have"}
	case in.Embargoed:
		return TradeChoice{Refuse: true, reason: "the partner has closed its ports to them"}
	}
	switch {
	case !in.InReach:
		c.Cap = 0
		c.reason = "nothing of theirs is within reach"
	case in.Nomad:
		c.Cap = p.CapNomad
	case in.Drive >= 2:
		c.Cap = p.CapDoor
	case in.Drive == 1:
		c.Cap = p.CapFast
	default:
		c.Cap = p.CapBase
	}
	if in.Xenophobe && in.Different {
		for k := range c.Mul {
			c.Mul[k] *= p.DifferentShare
		}
		c.reason = "half to a people so unlike them"
	}
	if in.Fear > p.FearBar && in.Stronger && in.Hostile {
		c.Mul[2] = 0
		c.reason = "no metal to a stronger people that strikes first"
	}
	if in.Spawning {
		c.Organic = min(1, c.Cap*p.SpawnOrganic)
	}
	if in.Grasping {
		c.Mul[0] *= p.GraspWant
		c.Mul[1] *= p.GraspWant
		c.Mul[2] *= p.GraspWant
	}
	return c
}
