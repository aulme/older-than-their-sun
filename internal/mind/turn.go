package mind

import "fmt"

// A relief fleet turning on its host: honour sets the base, posture
// multiplies it, and there has to be an opening.

// TurnInput is the fleet and the world it keeps.
type TurnInput struct {
	Honour    string
	Posture   string
	Betrayed  bool    // the host broke faith with the fleet's people before
	HostMil   float64 // the host's level with its miracles
	HostShips float64 // the host's ships at the world
	Guns      float64 // the host's guns over it
	Relief    float64 // others' ships standing at the world
	Mil       float64 // the fleet's people's level with its miracles
	Ships     int     // the fleet's ships
	Greed     float64
}

// Turning is the chance per thousand years, and why.
type Turning struct {
	Rate float64
	Hold float64 // what the host holds the world with, in levels
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
	hold := in.HostMil + ShipLevels(in.HostShips+in.Guns+in.Relief)
	if in.Mil+ShipLevels(float64(in.Ships)) <= hold+p.Opening {
		return Turning{Hold: hold}
	}
	return Turning{Rate: rate * (p.GreedBase + in.Greed), Hold: hold}
}
