package history

import (
	"worldgen/internal/plague"
	"worldgen/internal/species"
)

// The evolver: a people whose shape does not stay put. At a slow rate a
// bio, sense or world trait is gained, lost or replaced, and every world
// type it has held long enough offers that world's trait, so an old
// evolver is at home everywhere. Each drift raises the difference those
// who fathomed it feel toward it, since the people they understood is
// not the one in front of them (dials.go), and a drift while a
// biological plague rages is a cure roll: what the plague lived in is
// gone from under it. What it is made of is not touched: a machine
// evolver rewrites its own minds, an eldritch one is never the same
// thing twice.

// driftGroups are the trait groups the drift moves, and how many of
// each a shape holds before a drift can only lose or replace.
var driftGroups = []string{"bio", "sense", "world"}

var driftRoom = map[string]int{"bio": 3, "sense": 2, "world": 3}

// opposites are the traits a shape cannot be at once: gaining one is
// leaving the other.
var opposites = map[string]string{"shortlived": "longlived", "longlived": "shortlived", "robust": "fragile", "fragile": "robust"}

// drift is the evolver's tick.
func (w *World) drift(c *Civ) {
	if !c.Active() || !c.Species.Is(species.Evolver) {
		return
	}
	k := &w.Cfg.Tuning.Kinds
	if c.heldKyr == nil {
		c.heldKyr = map[string]float64{}
	}
	for _, s := range c.Systems {
		if a := w.G.Sys[s].Arch; a != "" {
			c.heldKyr[a] += w.dt
		}
	}
	if !w.chance(k.Drift) {
		return
	}
	sp := c.Species
	if w.sharedSpecies(c) {
		sp = sp.Branch() // the drift is this people's; its kin keep the shape they had
		c.Species = sp
		w.register(sp)
	}
	what := w.driftOnce(c, sp)
	if what == "" {
		return
	}
	c.Drifts++
	c.Tally.Drifts++
	w.recompute(c)
	w.factOf(FDrifted, c, nil, c.Home, what)
	if w.R.Float64() < 0.3 || c.Drifts == 1 {
		w.log("The %s have changed again: %s. Whoever knew them knew something else.", c.Tok(), what)
	}
	w.driftCure(c)
}

// sharedSpecies says whether another living people carries this one's species.
func (w *World) sharedSpecies(c *Civ) bool {
	for _, e := range w.Civs {
		if e != c && e.Living() && e.Species == c.Species {
			return true
		}
	}
	return false
}

// driftOnce is one change of shape: a world's trait taken if one is
// owed, else a trait of a drawn group gained, lost or replaced; "" when
// the draw found nothing to change. It says what changed.
func (w *World) driftOnce(c *Civ, sp *species.Species) string {
	k := &w.Cfg.Tuning.Kinds
	// a world held long enough offers its trait first
	var owed []string
	for _, key := range sortedKeys(c.heldKyr) {
		if c.heldKyr[key] < k.DriftWorld {
			continue
		}
		if a := species.ArchetypeByKey(key); a != nil {
			for _, t := range a.Traits {
				if !sp.Has(t) {
					owed = append(owed, t)
				}
			}
		}
	}
	// gain is a trait taken on: in place of its opposite if the shape has one
	gain := func(t *species.Trait, how string) string {
		if o := opposites[t.Key]; o != "" && sp.Has(o) {
			old := species.Get(o)
			sp.Replace(o, t.Key)
			return "where they were " + old.Name + " they are " + t.Name + how
		}
		sp.Add(t.Key)
		return "they are " + t.Name + " now" + how
	}
	if len(owed) > 0 && w.R.Float64() < 0.5 {
		return gain(species.Get(owed[w.R.IntN(len(owed))]), ", as the worlds they hold made them")
	}
	group := driftGroups[w.R.IntN(len(driftGroups))]
	have := sp.Of(group)
	full := len(have) >= driftRoom[group]
	switch roll := w.R.Float64(); {
	case len(have) > 0 && (roll < 0.3 || full && roll < 0.5):
		old := have[w.R.IntN(len(have))]
		sp.Remove(old.Key)
		return "they are no longer " + old.Name
	case len(have) > 0 && (roll < 0.6 || full):
		old := have[w.R.IntN(len(have))]
		t := species.PickFor(w.R, group, sp)
		if t == nil {
			return ""
		}
		if o := opposites[t.Key]; o != "" && sp.Has(o) && o != old.Key {
			sp.Remove(o)
		}
		sp.Replace(old.Key, t.Key)
		return "where they were " + old.Name + " they are " + t.Name
	default:
		if group == "world" {
			return "" // a world's trait comes only from a world held
		}
		t := species.PickFor(w.R, group, sp)
		if t == nil {
			return ""
		}
		return gain(t, "")
	}
}

// driftCure is the cure a drift is: for every biological plague raging in
// the people, one roll with the drift's bonus on the ladder.
func (w *World) driftCure(c *Civ) {
	t := &w.Cfg.Tuning.Plague
	for _, pid := range sortedInts(c.Infections) {
		inf := c.Infections[pid]
		p := w.Plagues[pid]
		if p.Kind != plague.Biological || inf.Carrier {
			continue
		}
		_, ladder := w.rungs(c, p.Kind)
		margin := plague.CureMargin(c.Sur, ladder+w.Cfg.Tuning.Kinds.DriftCure, w.R.NormFloat64()*t.Spread, p.Contagion, w.dirt(c), 0, t)
		if plague.Band(margin, t) == plague.Cured {
			w.log("What %s lived in is gone from under it: the %s changed, and it did not change with them.", p.Tok(), c.Tok())
			w.cure(c, p)
		}
	}
}

// driftGap is what the drift adds to the difference a people feels
// toward another it fathomed: a quarter per drift since.
func driftGap(c, e *Civ) float64 {
	if !c.Fathomed[e.ID] {
		return 0
	}
	return 0.25 * float64(e.Drifts-c.FathomedAt[e.ID])
}
