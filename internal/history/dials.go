package history

import (
	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// Dials are a people's temperament as numbers, set from traits at birth and
// nudged by scars and morale. Decisions read dials, never traits, so the
// table below is the one place a temperament is tuned. The type is the
// mind's; the mind reads dials and never traits.
type Dials = mind.Dials

// dialTable is what each trait does to the dials, as deltas from a half.
var dialTable = map[string]Dials{
	// posture
	"pacifist":    {Aggression: -0.5, Fear: 0.2, Loyalty: 0.1},
	"defensive":   {Aggression: -0.3, Fear: 0.2},
	"submissive":  {Aggression: -0.4, Fear: 0.4, Risk: -0.3},
	"opportunist": {Aggression: 0.3, Greed: 0.3, Risk: -0.3, Loyalty: -0.2, Hunger: 0.3},
	"conqueror":   {Aggression: 0.5, Greed: 0.3, Risk: 0.3, Fear: -0.2, Patience: -0.2},
	"vengeful":    {Aggression: 0.1, Patience: 0.3, Loyalty: 0.2},
	"confederate": {Loyalty: 0.2, Fear: 0.1, Hunger: 0.2},
	"unyielding":  {Risk: 0.3, Fear: -0.3},
	// honour
	"faithful":  {Loyalty: 0.4},
	"faithless": {Loyalty: -0.4, Greed: 0.1},
	// drive
	"xenophobic":    {Hate: 0.5, Loyalty: -0.1},
	"expansionist":  {Greed: 0.2, Aggression: 0.1},
	"cautious":      {Risk: -0.3, Hunger: 0.2, Fear: 0.1},
	"curious":       {Hunger: 0.2},
	"contemplative": {Aggression: -0.2, Patience: 0.2},
	"pragmatic":     {Risk: -0.1, Greed: 0.1},
	// order
	"solitary":   {Loyalty: -0.2},
	"collective": {Loyalty: 0.1},
	"herd":       {Fear: 0.2},
	"hive":       {Loyalty: 0.1, Fear: -0.1},
}

// scarDials is what a scar does, in a fixed order so sums are reproducible.
var scarDials = []struct {
	scar string
	d    Dials
}{
	{ScarChains, Dials{Fear: 0.2}},
	{ScarBurningSky, Dials{Aggression: -0.2}},
	{ScarFatalism, Dials{Risk: 0.2, Fear: -0.2}},
	{ScarCentralism, Dials{Aggression: 0.1}},
	{ScarQuarantine, Dials{Hate: 0.2}},
	{ScarChurch, Dials{Hate: 0.1, Aggression: 0.1}},
	{ScarStewardship, Dials{Greed: -0.2}},
}

// setDials derives the dials; called from recompute.
func (w *World) setDials(c *Civ) {
	d := Dials{Aggression: 0.5, Risk: 0.5, Greed: 0.5, Fear: 0.5, Loyalty: 0.5, Hunger: 0.5, Patience: 0.5, Hate: 0.5}
	for _, t := range c.Species.Traits {
		d.Add(dialTable[t.Key])
	}
	for _, sd := range scarDials {
		if c.Scars[sd.scar] {
			d.Add(sd.d)
		}
	}
	d.Aggression += 0.1 * clamp(c.Morale, -2, 2)
	c.LoreDials = w.loreDials(c)
	d.Add(c.LoreDials)
	d.Aggression = clamp(d.Aggression, 0.05, 0.95)
	d.Risk = clamp(d.Risk, 0.05, 0.95)
	d.Greed = clamp(d.Greed, 0.05, 0.95)
	d.Fear = clamp(d.Fear, 0.05, 0.95)
	d.Loyalty = clamp(d.Loyalty, 0.05, 0.95)
	d.Hunger = clamp(d.Hunger, 0.05, 0.95)
	d.Patience = clamp(d.Patience, 0.05, 0.95)
	d.Hate = clamp(d.Hate, 0.05, 0.95)
	c.Dials = d
}

// posture is the stance trait: how a people makes war.
func (c *Civ) posture() string {
	for _, t := range c.Species.Traits {
		if t.Group == "stance" {
			return t.Key
		}
	}
	return "defensive"
}

// honour is whether a promise binds.
func (c *Civ) honour() string {
	for _, t := range c.Species.Traits {
		if t.Group == "honour" {
			return t.Key
		}
	}
	return "practical"
}

// hostile is true of postures that strike first.
func (c *Civ) hostile() bool {
	return mind.Hostile(c.posture()) || c.Has("xenophobic")
}

// difference is how alien two species are to each other: another kind
// counts most, then the order of their societies, then senses and body,
// then the world they came from.
func difference(a, b *species.Species) float64 {
	d := 0.0
	if a.Kind != b.Kind {
		d += 2.5
	}
	oa, ob := orgOf(a), orgOf(b)
	if oa != ob {
		d += 0.5
		if oa == "hive" || ob == "hive" || oa == "nonconscious" || ob == "nonconscious" {
			d += 0.5
		}
	}
	for _, t := range a.Traits {
		if t.Group == "sense" && !b.Has(t.Key) {
			d += 0.5
		}
	}
	for _, t := range b.Traits {
		if t.Group == "sense" && !a.Has(t.Key) {
			d += 0.5
		}
	}
	if a.World.Key != b.World.Key {
		d += 0.5
	}
	return d
}

func orgOf(s *species.Species) string {
	for _, t := range s.Traits {
		if t.Group == "org" {
			return t.Key
		}
	}
	return ""
}

// hates is true when a xenophobe finds another people too different to
// treat as anything but a thing to end.
func (c *Civ) hates(e *Civ) bool {
	return c.Has("xenophobic") && difference(c.Species, e.Species) >= 2.5
}
