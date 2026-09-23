package history

import (
	"math"

	"worldgen/internal/plague"
)

// Continuity: how much of a people's past reaches its present
// (specs/proposals/continuity.md). A people's bodies live a span its blood
// and its medicine set (lifespanOf); at a thousand-year tick a thirty-year
// kind turns over thirty-three times and a thousand-year kind once, and
// each turning loses a share of what the people knew of itself, less for
// a people that writes, prints, networks and keeps archives. What is left
// across the thousand years is its continuity, on nought to one, never
// quite one: a mind that has run ten million years has drifted or chosen
// to forget.
//
// It is derived, never stored: the wearing reads it for how fast a tale
// is lost (memory), the ways setting read it for how fast they set, and a
// dark age for how deep it goes. Both ends fail, differently. A people
// that remembers everything cannot change, and sets faster; one that
// remembers nothing has nothing to climb back with, and falls further.
// Continuity is not stiffness: stiffness is whether the present can move,
// continuity whether the past is known, and a people can be either
// without the other.

// lifeTech is what the arts do to a body's span: medicine and clean water
// stop the young dying, life extension stops the old, and a germline
// remade lives longer still. A mortality creed refuses the last two.
var lifeTech = []struct {
	node   string
	times  float64
	cheats bool // cheats death: a mortality creed will not have it
}{
	{"medicine", 1.3, false}, {"sanitation", 1.2, false}, {"life_extension", 5, true}, {"germline", 1.5, true},
}

// memoryTable is the arts that keep the past outside the skull, each a
// multiplier on what a generation loses.
var memoryTable = []struct {
	node string
	keep float64
}{
	{"writing", 0.6}, {"printing", 0.7}, {"networks", 0.6}, {"substrate_minds", 0.3}, {"long_thought", 0.3},
}

// lifespanOf is the span in years a body of the people lives now: its
// blood's, times its medicine, cut while a plague of the body is in it.
// Zero is no natural turnover at all.
func (w *World) lifespanOf(c *Civ) float64 {
	sp := c.Species
	if !sp.Mortal() || sp.Lifespan <= 0 {
		return 0
	}
	l := float64(sp.Lifespan)
	for _, row := range lifeTech {
		if c.Known[row.node] && !(row.cheats && c.Scars[ScarMortality]) {
			l *= row.times
		}
	}
	for pid := range c.Infections {
		if w.Plagues[pid].Kind != plague.Memetic {
			l *= w.Cfg.Tuning.Continuity.SickCut // any one of them: the order is not read
			break
		}
	}
	return l
}

// generations is how many times the people turns over in a thousand years.
func (w *World) generations(c *Civ) float64 {
	if l := w.lifespanOf(c); l > 0 {
		return 1000 / l
	}
	return 0
}

// continuityLoss is what a thousand years costs the people's past, the x
// of exp(-x): the generations times what each loses, the drift, and the
// cut of a dark age still close.
func (w *World) continuityLoss(c *Civ) float64 {
	t := &w.Cfg.Tuning.Continuity
	per := t.Loss * c.Species.Profile().Memory
	if c.Has("memory") {
		per *= 0.3 // of unbroken memory across generations
	}
	if c.Has("swarming") {
		per *= 0.8
	}
	if c.Has("collective") {
		per *= 0.7
	}
	for _, row := range memoryTable {
		if c.Known[row.node] {
			per *= row.keep
		}
	}
	if n := min(c.archives(), t.Archives); n > 0 {
		per *= math.Pow(t.Archive, float64(n))
	}
	if c.Boons[BoonCommunion] {
		per *= 0.5
	}
	x := w.generations(c)*per + t.Drift
	if l := c.Leader; l != nil && l.Mad >= 1 {
		x += w.Cfg.Tuning.Leaders.MadPurge // the immortal at the top, the amnesia below
	}
	if c.DarkAges > 0 {
		if since := float64(w.Now-c.LastDark) / 1000; since < t.DarkKyr {
			x += t.Dark * (1 - since/t.DarkKyr)
		}
	}
	return x
}

// continuity is the share of a people's past that reaches across a
// thousand years. A people gone into its heaven keeps all of it: nothing
// in there is born or dies.
func (w *World) continuity(c *Civ) float64 {
	if c.inHeaven() {
		return 1
	}
	return math.Exp(-w.continuityLoss(c))
}

// memory is how well a people keeps a tale: a multiplier on the rate of
// wear, one at the continuity the wearing was tuned at and falling with
// what is lost, so a people that loses half as much of its past a
// thousand years forgets its tales half as fast.
func (w *World) memory(c *Civ) float64 {
	return (1 - w.continuity(c)) / (1 - w.Cfg.Tuning.Continuity.Ref)
}

// doublings is how many times over the middle a people loses of its past
// a thousand years, on a log scale: continuity bunches up under one for
// every people with medicine and letters, and what separates them is the
// loss, a machine's a third of a mayfly's. Negative keeps more than most.
func (w *World) doublings(c *Civ) float64 {
	loss := max(1-w.continuity(c), 1e-4)
	return math.Log2(loss / w.Cfg.Tuning.Continuity.MidLoss)
}

// contStiff is continuity's term in the rate a people's ways set at: the
// past having authority. One at the middle loss, more as less is lost.
func (w *World) contStiff(c *Civ) float64 {
	t := &w.Cfg.Tuning.Continuity
	return clamp(math.Exp2(-t.StiffPow*w.doublings(c)), t.StiffMin, t.StiffMax)
}

// archives counts the archives a people keeps standing and fed.
func (c *Civ) archives() int {
	n := 0
	for _, wk := range c.Works {
		if wk.Key == "archive" && !wk.Dark {
			n++
		}
	}
	return n
}

// inHeaven says whether the people uploaded into a paradise of its own
// and did not come out: what is left of it is the substrate it runs on.
func (c *Civ) inHeaven() bool { return c.Stage == Remnant && c.Cause == "heaven" }

// The upload, the filter on Mind Uploading: machine deathlessness was
// free and biological deathlessness cost the Long Silence, so this is
// the other road. Overcome, the people goes on in software: a blood of
// its own with no turnover at all, so its continuity is the drift and
// its ways set the faster for it, and biology's plagues are traded for
// computation's: the body's at a quarter, the mind's as any mind's.
// Scarred, the flesh is made sacred. Declined, it goes into its heaven:
// most of the people uploads into a paradise and stops acting, the realm
// contracts to the world the substrate runs on, and what is left there
// is a running substrate full of minds that are still there and are not
// coming out, with their telling frozen at the moment they went in
// (continuity is one; nothing wears). When the substrate goes at last
// its remain carries that telling for a finder.
func init() {
	def(&Filter{
		Key: "upload",
		Adjust: func(w *World, c *Civ) ([]string, float64, string) {
			f := filters["upload"]
			return f.Levels, w.Cfg.Tuning.Continuity.UploadStiff * min(c.Stiff, 3), f.Domain
		},
		Overcome: func(w *World, c *Civ) {
			sp := c.Species.Branch() // the same people, run on something else; its heirs and its cults keep the blood
			sp.Software, sp.Lifespan = true, 0
			w.register(sp)
			w.setSpecies(c, sp)
			w.recompute(c)
			w.faced(c, "upload", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarFlesh)
			w.faced(c, "upload", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			w.faced(c, "upload", "declined", "", c.Home)
			w.contract(c, because("heaven"))
			if c.Stage == Remnant && len(c.Systems) > 0 {
				s := c.Home
				if !contains(c.Systems, s) {
					s = c.Systems[0]
				}
				w.raise(c, "heaven", "uploading", s)
			}
		},
	})
}
