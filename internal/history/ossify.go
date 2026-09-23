package history

import (
	"math"

	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// Ossification: how an old people falls. Nobody dies of age. A people's
// ways set instead, at a rate its size, the fading sky, its stillness and
// its nature set (Civ.Stiff), and everything genuinely new takes some of
// it off. Stiffness shows before it breaks: research and expansion slow,
// the council prefers what it did last time, the neighbours read it as
// slow to answer. Past one the people faces Ossification, a social filter
// whose difficulty is its stiffness: overcome is a renaissance, the near
// miss is true ossification (the people sets, acts every other tick, and
// cannot be scarred twice), the bad miss is the break, a civil war or a
// dark age of variable depth (sunder.go, civ.go). Never contraction, never
// extinction.

// stiffInput is what sets the rate a people's ways set at.
type stiffInput struct {
	Worlds    int
	Fertility float64
	Still     bool    // nothing new in the window: no war, no world, no node, no meeting
	Ossified  bool    // already set: it works, for a while
	Iron      int     // the iron answers held: centralism, stewardship, quarantine, fatalism
	Nature    float64 // the profile's term: an evolver's institutions drift with it
	Traits    float64 // the trait table's product
	Past      float64 // continuity's term: a people that remembers everything sets (continuity.go)
}

// stiffTraits is what each trait does to the rate, the old weight table
// as rates instead of margins.
// The long-lived, the short-lived and the unbroken memory are not in it:
// what they did here is continuity's, which reads the span itself.
var stiffTraits = map[string]float64{
	"caste": 1.3, "solitary": 1.15,
	"individualist": 0.8, "nomadic": 0.6,
	"swarming": 0.5, // a nest has no institutions
}

// ironScars are the iron answers: each makes a people set faster.
var ironScars = func() []string {
	var out []string
	for _, s := range tables.scars {
		if s.Iron {
			out = append(out, s.Key)
		}
	}
	return out
}()

// stiffGrowth is the growth table: stiffness per thousand years.
func stiffGrowth(in stiffInput, t *mind.OssifyTuning) float64 {
	g := t.Base * (1 + t.PerWorld*float64(in.Worlds)) * (1 + t.Fade*(1-in.Fertility))
	if in.Still {
		g *= t.Still
	}
	if in.Ossified {
		g *= t.Ossified
	}
	g *= math.Pow(t.IronScar, float64(in.Iron))
	return g * in.Nature * in.Traits * in.Past
}

func (c *Civ) traitStiff() float64 {
	m := 1.0
	for _, tr := range c.Species.Traits {
		if v, ok := stiffTraits[tr.Key]; ok {
			m *= v
		}
	}
	return m
}

// stiffens says whether a people's ways can set at all: a hive has no
// institutions and an unconscious people nothing that could harden.
func (c *Civ) stiffens() bool { return c.Species.Profile().Can(species.Stiffens) }

// tickStiff is the tick step at the old Weight's place: the growth, the
// foresighted tilt, and the facing once stiffness is past one.
func (w *World) tickStiff(c *Civ) {
	t := &w.Cfg.Tuning.Ossify
	if !c.Active() || !c.stiffens() {
		return
	}
	iron := 0
	for _, s := range ironScars {
		if c.Scars[s] {
			iron++
		}
	}
	worlds := len(c.Systems)
	if c.Aloft {
		worlds = len(w.fleets(c))
	}
	c.Stiff += w.dt * stiffGrowth(stiffInput{
		Worlds: worlds, Fertility: w.fertility(),
		Still:    len(c.Wars) == 0 && float64(w.Now-c.Still)/1000 >= t.StillKyr,
		Ossified: c.Ossified, Iron: iron,
		Nature: c.Species.Profile().Stiffen, Traits: c.traitStiff(), Past: w.contStiff(c),
	}, t)
	if c.Stiff > t.ForeseeAt && w.foresees(c) {
		c.Focus["society"] = max(c.Focus["society"], t.ForeseeTilt)
	}
	if c.Stiff > 1 && w.chance(t.Chance*(c.Stiff-1)) {
		adj := 0.0
		if w.foresees(c) {
			adj -= t.Foresight
		}
		c.Tally.OssFaced++
		c.Tally.OssStiff += c.Stiff
		w.face(c, "ossification", adj)
	}
}

// foresees says whether a people sees the fall coming: deep governance,
// or the shape of the cycle. Knowing does not stop it; it changes what
// they do on the eve.
func (w *World) foresees(c *Civ) bool { return c.Known["deep_governance"] || c.KnowsCycle }

// renew is something genuinely new taking stiffness off: a node mastered
// from a find, an uplift, a branch, a first meeting, a war fought to
// peace, a world lost.
func (w *World) renew(c *Civ, by float64) {
	if !c.Living() {
		return
	}
	c.Stiff = max(0, c.Stiff-by)
	c.Still = w.Now
}

// stir marks something new without lowering: a war, a world, a node.
func (w *World) stir(c *Civ) { c.Still = w.Now }

// reset is the institutions gone: a renaissance, a dark age, a miracle.
func (w *World) reset(c *Civ) {
	c.Stiff, c.Ossified, c.Still = 0, false, w.Now
}

// offTick says whether an ossified people sits this tick out: half of them
// act on any given tick, and each acts every other one.
func (w *World) offTick(c *Civ) bool {
	return c.Ossified && (int(w.Now/1000)+c.ID)%2 == 1
}

// stiffMul is what stiffness does to research and expansion: a third
// slower at one, half at two.
func (c *Civ) stiffMul() float64 { return 1 / (1 + c.Stiff/2) }

// renewing says whether a people is in the surge after a renaissance.
func (w *World) renewing(c *Civ) bool {
	t := &w.Cfg.Tuning.Ossify
	return c.Renaissances > 0 && float64(w.Now-c.Renewed)/1000 < t.RenewalKyr
}

// stiffWord is the portrait's word for a people's stiffness.
func (c *Civ) stiffWord() string {
	return tables.stiffness[c.stiffStage()].Word
}

// stiffStage is where a people's ways stand, as data/levels.json divides
// it: the last row is ossified, the rest by the stiffness each begins at.
func (c *Civ) stiffStage() int {
	rows := tables.stiffness
	if c.Ossified {
		return len(rows) - 1
	}
	for i := len(rows) - 2; i > 0; i-- {
		if c.Stiff >= rows[i].From {
			return i
		}
	}
	return 0
}

// StiffWord is stiffWord for the legends.
func (c *Civ) StiffWord() string { return c.stiffWord() }

// renaissance is the filter overcome: the institutions remade, the people
// young again, and each one harder than the last.
func (w *World) renaissance(c *Civ) {
	w.reset(c)
	c.Renewed = w.Now
	c.Renaissances++
	c.Morale += 1
	c.Tally.OssRenewed++
	w.told(FRenaissance, c, nil, c.Home)
}

// set is the near miss: the people stops changing.
func (w *World) set(c *Civ) {
	c.Ossified = true
	c.Tally.OssSet++
	w.event(KSet, c, nil, -1, P{})
}

// breakDown is the bad miss: a civil war where the people has the worlds
// or the fleets and the factions for one, else a dark age for the reason
// given. Ossification is cleared by either. The Succession breaks the
// same way (leaders.go).
func (w *World) breakDown(c *Civ, why reason) {
	t := &w.Cfg.Tuning.Ossify
	parts := len(c.Systems)
	if c.Aloft {
		parts = len(w.fleets(c))
	}
	if parts >= 2 && c.Species.Profile().Can(species.CivilWars) && w.R.Float64() < min(t.CivilWarCap, t.CivilWar*float64(parts)) {
		if w.civilWar(c) {
			return
		}
	}
	w.darkAge(c, why)
}

// resent is every grudge write: what another people did, added to what is
// held against it. A people that holds no grudges (the unconscious) holds
// nothing: the map stays empty and nothing is taken to heart.
func (c *Civ) resent(id int, by float64) {
	if !c.Species.Profile().Can(species.HoldsGrudges) || by <= 0 {
		return
	}
	c.Grudge[id] += by
}

// forgive is the per-tick decay of grudges: a wrong is held only as long
// as it is remembered, and what is still told feeds it back.
func (w *World) forgive(c *Civ) {
	t := &w.Cfg.Tuning.Ossify
	keep := math.Pow(t.GrudgeDecay, w.dt)
	for id, g := range c.Grudge {
		g *= keep
		if g < t.GrudgeFloor {
			delete(c.Grudge, id)
		} else {
			c.Grudge[id] = g
		}
	}
}

func init() {
	def(&Filter{
		Key: "ossification",
		Adjust: func(w *World, c *Civ) ([]string, float64, string) {
			t := &w.Cfg.Tuning.Ossify
			return []string{"soc"}, c.Stiff + t.Renaissance*float64(c.Renaissances), "society"
		},
		Overcome: func(w *World, c *Civ) { w.renaissance(c) },
		Scar: func(w *World, c *Civ) {
			if c.Ossified {
				c.Tally.OssBroke++
				w.breakDown(c, because("ossified")) // it cannot set twice: the second near miss is the break
				return
			}
			w.set(c)
		},
		Decline: func(w *World, c *Civ) {
			c.Tally.OssBroke++
			w.breakDown(c, because("ossified"))
		},
	})
}
