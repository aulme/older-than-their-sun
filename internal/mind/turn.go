package mind

import "fmt"

// A relief fleet turning on its host: honour sets the base, posture
// multiplies it, and there has to be an opening.

// TurnInput is the fleet and the world it keeps.
type TurnInput struct {
	Honour   string
	Posture  string
	Betrayed bool    // the host broke faith with the fleet's people before
	HostMil  float64 // the host's level with its miracles
	Relief   float64 // others standing at the world
	Mil      float64 // the fleet's level
	AtHome   bool    // the world is the host's home
	Grid     bool
	Greed    float64
}

// Turning is the chance per thousand years, and why.
type Turning struct {
	Rate float64
	Hold float64 // what the host holds the world with
}

// Why says the chance.
func (u Turning) Why() string {
	if u.Rate == 0 {
		return "keeps faith"
	}
	return fmt.Sprintf("would turn at %.4f a millennium against a hold of %.1f", u.Rate, u.Hold)
}

// Turn is the chance a relief fleet seizes the world it was sent to keep.
func Turn(in TurnInput, t *Tuning) Turning {
	p := &t.Turn
	rate := 0.0
	switch in.Honour {
	case Faithful:
		rate = p.Faithful
	case Practical:
		rate = p.Practical
	case Faithless:
		rate = p.Faithless
	}
	switch in.Posture {
	case Opportunist, Conqueror:
		rate *= p.Hostile
	case Vengeful:
		if in.Betrayed {
			rate *= p.Vengeful
		} else {
			rate = 0
		}
	case Pacifist:
		rate = 0
	}
	if rate == 0 {
		return Turning{}
	}
	a := &t.Appraise
	hold := in.HostMil + in.Relief - in.Mil + a.Defence
	if in.AtHome {
		hold += a.HomeDefence
	}
	if in.Grid {
		hold += a.Grid
	}
	if in.Mil <= hold+p.Opening {
		return Turning{Hold: hold}
	}
	return Turning{Rate: rate * (p.GreedBase + in.Greed), Hold: hold}
}
