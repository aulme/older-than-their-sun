package mind

import "math/rand/v2"

// Colony ships: how often a people sends one, and where.

// ExpandInput is a people's readiness to settle.
type ExpandInput struct {
	Systems  int
	Mul      float64 // the temperament's and the moment's multiplier on the rate
	Era      int
	Reach    float64
	Nowhere  func() bool // nothing within reach to settle; asked only when it matters
	Parasite bool        // a rider that has not learned to live free
	FTL      bool
}

// ExpandPlan is the readiness.
type ExpandPlan struct {
	Rate  float64 // ships per thousand years
	Ships bool    // a people with nowhere to go works on ships
	Reach float64 // how far it will settle
	Hop   float64 // how far one ship goes
}

// Expand is the rate and range of colony ships: a chance per system to a
// cap, the reach cut for a rider, a hop unless the Door.
func Expand(in ExpandInput, t *Tuning) ExpandPlan {
	p := &t.Expand
	e := ExpandPlan{Rate: min(p.MaxRate, p.Rate*float64(in.Systems)) * in.Mul, Reach: in.Reach}
	e.Ships = in.Era >= p.NeedShipsEra && in.Reach < p.NeedShipsBelow && in.Nowhere()
	if in.Parasite {
		e.Reach *= p.ParasiteReach
	}
	e.Hop = min(e.Reach, p.Hop)
	if in.FTL {
		e.Hop = e.Reach
	}
	return e
}

// Colony is a star a colony ship weighs, nearest first, already within
// reach and not known taken, targeted or dreaded.
type Colony struct {
	ID      int
	Read    bool // its worlds are known
	Livable bool // and one can be lived on
	Guess   bool // unread, but its colour is right and it is not dead
}

// Target picks where a ship goes: the nearest read, livable star; failing
// that, now and then, a guess at the nearest star of the right colour.
// The second return says it was a guess. -1 when no ship goes.
func Target(stars []Colony, r *rand.Rand, t *Tuning) (int, bool) {
	blind := -1
	for _, s := range stars {
		if !s.Read {
			if blind < 0 && s.Guess {
				blind = s.ID
			}
			continue
		}
		if s.Livable {
			return s.ID, false
		}
	}
	if blind >= 0 && r.Float64() < t.Expand.Blind {
		return blind, true
	}
	return -1, false
}
