package mind

import "fmt"

// Refusing: a people that suspects a sender of carrying a sickness may
// close its ears to every message from it and its ports to its goods,
// while the suspicion lasts. The cost is real: no news, no intel, no
// pacts, no calls, and a partner in want calls the closed ports an
// embargo.

// RefuseInput is the suspicious people's temperament.
type RefuseInput struct {
	Fear     float64
	Creed    bool // holds the quarantine creed
	Cautious bool
	Censor   bool // censorship working: refuses at twice the chance
}

// Refusal is the chance and its reason.
type Refusal struct {
	Chance float64
	reason string
}

func (r Refusal) Why() string { return fmt.Sprintf("closing at %.2f: %s", r.Chance, r.reason) }

// Refuse is the chance a suspicious people closes up: its fear, its creed
// and its caution, capped, doubled by censorship.
func Refuse(in RefuseInput, t *Tuning) Refusal {
	p := &t.Refuse
	r := Refusal{Chance: in.Fear, reason: "fear"}
	if in.Creed {
		r.Chance += p.Creed
		r.reason += ", the creed"
	}
	if in.Cautious {
		r.Chance += p.Caution
		r.reason += ", caution"
	}
	if in.Censor {
		r.Chance *= p.Censor
		r.reason += ", censorship"
	}
	r.Chance = min(p.Cap, r.Chance)
	return r
}
