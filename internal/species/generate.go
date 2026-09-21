package species

import (
	"math/rand/v2"
)

// The generator rolls a chain: the substrate by weight, then the trait
// rolls and each modifier in registry order by its base odds times the
// tilt of everything already rolled, then the trait groups with the
// entries' skips and tilts. Tilts multiply the odds, not the chance, so
// nothing is certain and nothing is impossible unless a legacy tilt says
// so outright.

// swarm is the swarming trait's roll, before the modifiers so it can tilt
// the hive.
var swarm = &Entry{
	Key:    "swarming",
	Legacy: Draw{Base: 7.0 / 93, Tilts: map[string]float64{"planetary": 0, "evolver": 0}},
	Draws:  Draw{Base: 1.0 / 14, Tilts: map[string]float64{"hive": 3}},
}

// TraitRolls are traits drawn by their own tilted roll in the chain rather
// than from their group; each Entry's key is the trait's.
var TraitRolls = []*Entry{swarm}

// groupRolls is the trait groups in the order they are drawn, with the
// chance of drawing one at all. An own group is drawn only for a people
// whose entries list it.
var groupRolls = []struct {
	group  string
	chance float64
	own    bool
}{
	{"org", 1, false}, {"stance", 1, false}, {"honour", 1, false}, {"drive", 1, false},
	{"bio", 0.5, false}, {"sense", 0.4, false}, {"power", 0.006, false},
	{"rider", 1, true}, {"way", 0.08, false}, {"seat", 1, true},
}

// Options steer a roll. The zero value is a cradle roll with the
// proposal's numbers.
type Options struct {
	Arch   string    // home world archetype key, "" to roll one
	Fix    bool      // Sub is fixed rather than rolled
	Sub    Substrate // the fixed substrate
	Mods   Mod       // modifiers fixed on; the rest are rolled
	Deep   bool      // the deep pass's substrate weights: the old galaxy is where the wonders and the leftovers come from
	Legacy bool      // the numbers that reproduce the old kind table, for a test
}

// Powers an eldritch people draws at birth: one to three.
const (
	bornPowersMin = 1
	bornPowersMax = 3
)

// Generate rolls a species. mult is the home star's multiplicity.
func Generate(r *rand.Rand, mult int) *Species { return Roll(r, mult, Options{}) }

// GenerateDeep rolls a species with the deep pass's substrate weights,
// for the peoples of past ages.
func GenerateDeep(r *rand.Rand, mult int, arch string) *Species {
	return Roll(r, mult, Options{Arch: arch, Deep: true})
}

// GenerateOn generates a species for a home world of a given archetype
// key, or a random one if the key is empty.
func GenerateOn(r *rand.Rand, mult int, arch string) *Species {
	return Roll(r, mult, Options{Arch: arch})
}

// GenerateWith makes a people whose substrate and given modifiers the
// story fixes; the rest is rolled through the same chain.
func GenerateWith(r *rand.Rand, mult int, arch string, sub Substrate, mods Mod) *Species {
	return Roll(r, mult, Options{Arch: arch, Fix: true, Sub: sub, Mods: mods})
}

// Roll runs the chain.
func Roll(r *rand.Rand, mult int, o Options) *Species {
	s := &Species{}
	if o.Fix {
		s.Sub = o.Sub
	} else {
		d := pickWeighted(r, Substrates, func(d *SubstrateDef) float64 {
			if o.Deep {
				return d.Deep
			}
			return d.draw(o.Legacy).Base
		})
		s.Sub = d.Sub
	}
	carried := []*Entry{&s.Sub.Def().Entry}
	tilt := func(key string) float64 {
		t := 1.0
		for _, e := range carried {
			if v, ok := e.draw(o.Legacy).Tilts[key]; ok {
				t *= v
			}
		}
		return t
	}
	roll := func(e *Entry) bool { return tilted(r, e.draw(o.Legacy).Base, tilt(e.Key)) }
	for _, e := range TraitRolls {
		if roll(e) {
			s.Add(e.Key)
			carried = append(carried, e)
		}
	}
	for _, d := range Mods {
		if o.Mods.Has(d.Mod) || roll(&d.Entry) {
			s.Mods |= d.Mod
			carried = append(carried, &d.Entry)
		}
	}
	s.World = ArchetypeByKey(o.Arch)
	if s.World == nil {
		s.World = pickWeighted(r, Archetypes, func(a *Archetype) float64 { return a.Weight })
		for _, d := range s.Mods.Defs() {
			if d.Cradle != "" && r.Float64() < d.CradleOdds {
				s.World = ArchetypeByKey(d.Cradle)
			}
		}
	}
	for _, t := range s.World.Traits {
		s.Add(t)
	}
	if mult == 3 {
		s.Add("threesuns")
	} else if mult == 2 {
		s.Add("hardy")
	}
	skips := func(group string) bool {
		for _, e := range carried {
			for _, g := range e.draw(o.Legacy).Skip {
				if g == group {
					return true
				}
			}
		}
		return false
	}
	owned := func(group string) bool {
		for _, e := range carried {
			for _, g := range e.Own {
				if g == group {
					return true
				}
			}
		}
		return false
	}
	pick := func(group string) *Trait {
		return pickWeighted(r, poolFor(group, s), func(t *Trait) float64 { return t.Weight * tilt(t.Key) })
	}
	for _, g := range groupRolls {
		if (g.own && !owned(g.group)) || skips(g.group) {
			continue
		}
		chance := g.chance
		for _, e := range carried {
			if v, ok := e.draw(o.Legacy).Groups[g.group]; ok {
				chance *= v
			}
		}
		if chance < 1 && r.Float64() >= chance {
			continue
		}
		switch g.group {
		case "seat":
			if s.Has("nomadic") {
				s.Add("throne") // a nomad hive's queen is a fleet
				continue
			}
		case "sense":
			s.Add(pick("sense").Key)
			if r.Float64() < 0.15 {
				s.Add(pick("sense").Key)
			}
			continue
		}
		s.Add(pick(g.group).Key)
	}
	s.Channel = rollChannel(r, s)
	if !s.Profile().Can(Researches) {
		// no tree: the pool instead, a few at birth
		n := bornPowersMin + r.IntN(bornPowersMax-bornPowersMin+1)
		for i := 0; i < n; i++ {
			if p := s.DrawPower(r); p != nil {
				s.AddPower(p.Key)
			}
		}
	}
	return s
}

// tiltedOdds is the chance of a roll whose odds are the base's times the
// tilt. A base of one or more is certain; a tilt of zero is impossible.
func tiltedOdds(base, tilt float64) float64 {
	if base >= 1 {
		return 1
	}
	if base <= 0 || tilt <= 0 {
		return 0
	}
	odds := base / (1 - base) * tilt
	return odds / (1 + odds)
}

// tilted rolls a chance of base tilted by tilt. Nothing is drawn from r
// when the outcome is certain either way.
func tilted(r *rand.Rand, base, tilt float64) bool {
	p := tiltedOdds(base, tilt)
	switch {
	case p <= 0:
		return false
	case p >= 1:
		return true
	}
	return r.Float64() < p
}

// pool is the traits of a group that the group draw can pick: those with
// a weight and no fit, since nothing is known of the species.
func pool(group string) []*Trait { return poolFor(group, nil) }

// poolFor is the traits of a group a species could roll: those with a
// weight whose fit, if any, the species meets.
func poolFor(group string, s *Species) []*Trait {
	var out []*Trait
	for _, t := range Traits {
		if t.Group != group || t.Weight <= 0 {
			continue
		}
		if t.Fit != nil && (s == nil || !t.Fit(s)) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func pickWeighted[T any](r *rand.Rand, xs []T, weight func(T) float64) T {
	total := 0.0
	for _, x := range xs {
		total += weight(x)
	}
	v := r.Float64() * total
	for _, x := range xs {
		v -= weight(x)
		if v < 0 {
			return x
		}
	}
	return xs[len(xs)-1]
}

// Pick draws a trait from a group with no tilts.
func Pick(r *rand.Rand, group string) *Trait {
	return pickWeighted(r, pool(group), func(t *Trait) float64 { return t.Weight })
}

// PickFor draws a trait of a group that a species could roll and does not
// have, with no tilts; nil when there is none left to draw.
func PickFor(r *rand.Rand, group string, s *Species) *Trait {
	var xs []*Trait
	for _, t := range poolFor(group, s) {
		if !s.Has(t.Key) {
			xs = append(xs, t)
		}
	}
	if len(xs) == 0 {
		return nil
	}
	return pickWeighted(r, xs, func(t *Trait) float64 { return t.Weight })
}

// rollChannel is how a people communicates, from what it is: a machine
// in the field, a hive in scents, a planet or an eldritch thing not at
// all; a body by sound unless it is deaf, hears light, senses the
// current, tastes the air or is rooted, each of which tilts it elsewhere.
func rollChannel(r *rand.Rand, s *Species) string {
	switch {
	case s.Is(Planetary), s.Is(Unconscious), s.Sub == Eldritch:
		return NoChannel
	case s.Sub == Machine, s.Is(Replicator):
		return Electromagnetic
	case s.Is(Hive):
		if r.Float64() < 0.7 {
			return Chemical
		}
		return Sound
	case s.Sub == Parasite:
		return NoChannel // it speaks through its host
	}
	w := map[string]float64{Sound: 70, Light: 6, Electromagnetic: 4, Chemical: 6, Touch: 2}
	if s.Has("deaf") {
		w[Sound] = 0
	}
	if s.Has("eyeless") {
		w[Light] = 0
		w[Sound] *= 1.5
	}
	if s.Has("lighthearing") {
		w[Light] *= 6
	}
	if s.Has("electric") || s.Has("magnetic") {
		w[Electromagnetic] *= 6
	}
	if s.Has("chemical") {
		w[Chemical] *= 6
	}
	if s.Has("sessile") {
		w[Touch] *= 8
		w[Chemical] *= 2
	}
	total := 0.0
	for _, k := range Channels {
		total += w[k]
	}
	x := r.Float64() * total
	for _, k := range Channels {
		x -= w[k]
		if x < 0 {
			return k
		}
	}
	return Sound
}
