package mind

import (
	"fmt"
	"strings"
)

// The slight: a war on a people's partner is a wrong done to that people,
// sized by the trade between them; and before a war the council weighs
// the slights it would cause in the peoples whose opinion it cares about.

// SlightInput is a partner's stake in the people about to be attacked.
type SlightInput struct {
	Flow      float64 // what the pair moves each tick, both ways
	Income    float64 // what the slighted people takes in each tick
	Dependent bool    // the partner's sending keeps its fed uses fed
}

// Slight is the size of the wrong: the pair's flow as a share of the
// slighted people's income, weighted, doubled by dependence, capped.
func Slight(in SlightInput, t *Tuning) float64 {
	p := &t.Slight
	s := p.Weight * in.Flow / max(in.Income, 1)
	if in.Dependent {
		s *= p.Dependent
	}
	return min(p.Cap, s)
}

// CareInput is whether an attacker minds what one people thinks of it.
type CareInput struct {
	Partner bool // trades with it
	Allied  bool // sworn to it
	Strong  bool // at least its own level, and in reach of its holdings
	Ruler   bool // its master or its vassal
	Hates   bool // too different to be let live
	Monster bool // remembered as a thing that does harm, or not understood
	Fear    float64
}

// Care is the weight of a people's opinion to the attacker: partners,
// allies, master and vassal in full, a strong neighbour by half, doubled
// by fear; nothing for the hated and monsters.
func Care(in CareInput, t *Tuning) float64 {
	p := &t.Slight
	if in.Hates || in.Monster {
		return 0
	}
	if in.Partner || in.Allied || in.Ruler {
		return 1
	}
	if in.Strong {
		if in.Fear > p.FearBar {
			return 2 * p.StrongWeight
		}
		return p.StrongWeight
	}
	return 0
}

// Slighted is one people the war would wrong, weighed.
type Slighted struct {
	Name   string
	Slight float64 // care × slight × stake
}

// Offence sums the slights a war would cause, halved for a conqueror.
func Offence(slights []Slighted, conqueror bool, t *Tuning) float64 {
	o := 0.0
	for _, s := range slights {
		o += s.Slight
	}
	if conqueror {
		o *= t.Slight.ConquerorShare
	}
	return o
}

// slightsWhy names the slighted.
func slightsWhy(slights []Slighted) string {
	if len(slights) == 0 {
		return ""
	}
	var parts []string
	for _, s := range slights {
		parts = append(parts, fmt.Sprintf("the %s (%.2f)", s.Name, s.Slight))
	}
	return ", would slight " + strings.Join(parts, " and ")
}
