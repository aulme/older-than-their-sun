package mind

import "math/rand/v2"

// A nomad fleet's next star.

// RoamHop is how far a fleet moves in one hop.
func RoamHop(reach float64, t *Tuning) float64 {
	return min(max(reach, t.Roam.HopMin), t.Roam.HopMax)
}

// Port is a star a fleet weighs.
type Port struct {
	ID   int
	Held bool // by a horror: never
	Good bool // unowned, or a trade partner's: preferred
}

// NextStar picks a star within the hop: one of the good ones at random,
// else any that is not held. -1 when there is nowhere to go.
func NextStar(ports []Port, r *rand.Rand) int {
	var good, any []int
	for _, p := range ports {
		if p.Held {
			continue
		}
		if p.Good {
			good = append(good, p.ID)
		} else {
			any = append(any, p.ID)
		}
	}
	if len(good) > 0 {
		return good[r.IntN(len(good))]
	}
	if len(any) > 0 {
		return any[r.IntN(len(any))]
	}
	return -1
}
