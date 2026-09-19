package mind

import (
	"fmt"

	"worldgen/internal/plague"
)

// Making a plague, and choosing to ride: the council's part of the
// plagues draft. A people that knows the craft makes a plague when it has
// a reason, and the reason sets the shape. A parasite people is a plague
// that chooses when to try; on every channel event it could ride, its
// council says whether it does, under its posture.

// WeaponInput is one people before the council of another that could
// make a plague for it.
type WeaponInput struct {
	Posture  string
	Hates    bool    // a xenophobe against the different
	AtWar    bool    // with the target now
	Grudge   float64 // held against the target
	Bar      float64 // the war bar the posture sets; a grudge above it is a reason
	Worth    bool    // the conqueror's appraisal: the worlds are worth more than the war would cost
	Vengeful bool    // the vengeful want the rival gone, not its worlds
}

// Weapon is whether to make one, and to what end.
type Weapon struct {
	Make   bool
	Aim    plague.Aim
	reason string
}

func (v Weapon) Why() string {
	if !v.Make {
		return "no plague: no reason"
	}
	return fmt.Sprintf("a plague for %s: %s", v.Aim, v.reason)
}

// MakeWeapon is the council's reason to make a plague and the shape that
// follows: a xenophobe makes the worst it can for the different; a people
// at war or holding a grudge above its bar wants the rival gone, unless
// it is a conqueror that wants the worlds; a conqueror whose appraisal
// says the worlds are worth more than the war softens them for the
// taking.
func MakeWeapon(in WeaponInput) Weapon {
	switch {
	case in.Hates:
		return Weapon{Make: true, Aim: plague.AimAll, reason: "too different to be let live"}
	case in.AtWar && in.Posture == Conqueror:
		return Weapon{Make: true, Aim: plague.AimWorlds, reason: "at war, and the worlds are wanted"}
	case in.AtWar:
		return Weapon{Make: true, Aim: plague.AimGone, reason: "at war"}
	case in.Grudge > in.Bar && in.Bar > 0:
		return Weapon{Make: true, Aim: plague.AimGone, reason: "a grudge above the bar"}
	case in.Worth && (in.Posture == Conqueror || in.Posture == Opportunist):
		return Weapon{Make: true, Aim: plague.AimWorlds, reason: "the worlds are worth more than the war"}
	}
	return Weapon{}
}

// RideInput is a parasite's council on one people it could try.
type RideInput struct {
	Posture string
	Hates   bool // a xenophobe against the different
	AtWar   bool
	Prey    bool // raging with another plague, or a dark age within memory
	Odds    bool // the appraisal likes the odds: acted above the bar
	Sworn   bool // a pact partner: never tried unless the betrayal rule already fires
	Hurry   bool // the hosts are dwindling: every people is prey
}

// Ride is whether to try.
type Ride struct {
	Try    bool
	reason string
}

func (r Ride) Why() string {
	if !r.Try {
		return "no attempt: " + r.reason
	}
	return "an attempt: " + r.reason
}

// TryRide is the parasite's choice: conquerors and opportunists try where
// the appraisal likes the odds; the rest try at war and where the target
// is prey; xenophobes try the different; a people in a hurry tries
// anyone; a pact partner is never tried.
func TryRide(in RideInput) Ride {
	switch {
	case in.Sworn:
		return Ride{reason: "sworn"}
	case in.AtWar:
		return Ride{Try: true, reason: "at war"}
	case in.Hates:
		return Ride{Try: true, reason: "too different to be let be"}
	case in.Hurry:
		return Ride{Try: true, reason: "nothing left to wear"}
	case in.Prey:
		return Ride{Try: true, reason: "prey"}
	case Hostile(in.Posture) && in.Odds:
		return Ride{Try: true, reason: "the odds"}
	case Hostile(in.Posture):
		return Ride{reason: "the odds are against it"}
	}
	return Ride{reason: "no cause"}
}
