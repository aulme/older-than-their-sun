package history

import (
	"sort"

	"worldgen/internal/species"
)

// Cosmic filters: supernovae, gamma-ray bursts, passing dark masses, and the
// slow death of a home star. Each has a blast radius; every civilisation
// inside faces the filter at once. Overcoming can mean migration.

func (w *World) cosmic() {
	w.starDeaths()
	w.flares()
	if w.chance(0.00015 * min(w.Law.Youth, 5)) {
		origin := w.R.IntN(len(w.G.Stars))
		w.blast(origin, 15+w.R.Float64()*15, "a gamma-ray burst", "A gamma-ray burst lights the sky near %s.", 0)
	}
	if w.chance(0.00003 * min(w.Law.Crowd, 30)) {
		origin := w.R.IntN(len(w.G.Stars))
		w.blast(origin, 6, "a passing dark mass", "Something heavy and unlit passes through the worlds of %s, and orbits come apart.", 1)
	}
}

// starDeaths kills stars scheduled to die this tick and marks failing ones.
func (w *World) starDeaths() {
	for i := range w.G.Stars {
		s := &w.G.Stars[i]
		if s.Dead() {
			continue
		}
		if Year(s.DiesAt) <= w.Now {
			if s.Massive() {
				s.Kill()
				w.blast(i, 30, "the supernova of "+w.star(i), "%s goes supernova.", 1)
				continue
			}
			cid := w.Owner[i]
			s.Kill()
			w.Bio[i] = BioNone
			if cid >= 0 {
				c := w.Civs[cid]
				if i == c.Home && c.Active() {
					w.leaveHome(c, sprintf("the death of %s", c.HomeName))
				} else {
					w.loseSystem(c, i, "frozen world", sprintf("lost their last world to the death of %s", w.star(i)))
				}
			} else if w.Bio[i] != BioNone {
				w.log("%s dies. Its worlds freeze.", w.star(i))
			}
			continue
		}
		if !s.Failing && Year(s.DiesAt)-w.Now < 5_000_000 && s.Hab > 0 {
			s.Failing = true
		}
	}
}

// blast applies a cosmic filter to everyone inside a radius.
func (w *World) blast(origin int, radius float64, what, text string, adj float64) {
	w.log(text, w.star(origin))
	inside := append(w.G.Near(origin, radius), origin)
	hit := map[int][]int{}
	for _, s := range inside {
		if s == w.G.Sol {
			continue
		}
		if w.Bio[s] == BioComplex {
			w.Bio[s] = BioSimple
		}
		if cid := w.Owner[s]; cid >= 0 {
			hit[cid] = append(hit[cid], s)
		}
	}
	for _, x := range w.Expeditions {
		if !x.Over && x.Kind == Roam && x.Base >= 0 && contains(inside, x.Base) {
			x.Over = true
			w.log("A fleet of the %s at %s is caught in it and is gone.", w.Civs[x.Owner].Name, w.star(x.Base))
		}
	}
	var cids []int
	for cid := range hit {
		cids = append(cids, cid)
	}
	sort.Ints(cids)
	for _, cid := range cids {
		worlds := hit[cid]
		c := w.Civs[cid]
		if !c.Living() {
			continue
		}
		if !c.Active() {
			for _, s := range worlds {
				w.loseSystem(c, s, "scoured world", "were sterilised by "+what)
			}
			continue
		}
		homeHit := contains(worlds, c.Home)
		if homeHit {
			adj++
		}
		w.blastWorlds = worlds
		w.blastWhat = what
		w.face(c, "cosmic", adj)
	}
}

// leaveHome moves a civilisation's home to another of its worlds, or ends it.
func (w *World) leaveHome(c *Civ, why string) {
	old := c.Home
	best, bd := -1, 1e9
	for _, s := range c.Systems {
		if s != old && w.G.Dist(old, s) < bd {
			best, bd = s, w.G.Dist(old, s)
		}
	}
	if best < 0 {
		w.loseSystem(c, old, "burned cradle", sprintf("died with their star, %s", c.HomeName))
		return
	}
	c.Systems = remove(c.Systems, old)
	w.Owner[old] = -1
	w.trace(old, "burned cradle", c.ID)
	c.Home = best
	oldName := c.HomeName
	c.HomeName = w.star(best)
	c.Dying = false
	c.Morale -= 1
	w.log("The %s leave %s to %s. %s is home now, and always a little less than the one before.", c.Name, oldName, why, c.HomeName)
}

// dyingSun is the slow filter: a failing home star degrades the world every
// tick, and Survival sets how long the species endures it.
func (w *World) dyingSun(c *Civ) {
	if !c.Active() || c.Home < 0 {
		return
	}
	s := &w.G.Stars[c.Home]
	if !s.Failing {
		return
	}
	if !c.Dying {
		c.Dying = true
		c.Endure = 300 * pow(1+c.Sur, 1.5) * (1 - 0.3*c.traitDiff("dying"))
		if c.Known["star_lifting"] {
			c.Endure *= 4
		}
		if c.Known["deep_root"] {
			c.Endure *= 2 // the mind goes down into the crust
		}
		if c.Species.Kind == species.MachineBorn {
			c.Endure *= 3 // cold is only cold
		}
		c.focus("propulsion", 2)
		c.focus("biology", 1.5)
		w.log("The sun of the %s is failing. %s grows harsher with every century. They have, perhaps, %d thousand years.", c.Name, c.HomeName, int(c.Endure))
	}
	if c.Known["star_lifting"] && !c.Boons["star kept"] {
		c.Boons["star kept"] = true
		w.log("The %s reach into %s and hold it together. Their sun will fail, but not yet.", c.Name, c.HomeName)
		c.Endure += 5000
	}
	c.Endure -= w.dt
	if c.Endure > 0 {
		if len(c.Systems) > 1 && c.Endure < 200 && w.chance(0.2) {
			w.leaveHome(c, "the failing of its sun")
		}
		return
	}
	if len(c.Systems) > 1 {
		w.leaveHome(c, "the failing of its sun")
		return
	}
	w.log("The %s endure under the failing sun of %s until they cannot. The last of them die looking up.", c.Name, c.HomeName)
	w.loseSystem(c, c.Home, "burned cradle", sprintf("died with their star, %s", c.HomeName))
}

func pow(x, y float64) float64 {
	r := 1.0
	for i := 0; i < int(y); i++ {
		r *= x
	}
	if y != float64(int(y)) {
		r *= sqrt(x)
	}
	return r
}

func sqrt(x float64) float64 {
	z := x / 2
	if z == 0 {
		return 0
	}
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func init() {
	def(&Filter{
		Key: "cosmic", Name: "the burning sky", Levels: []string{"sur"}, Diff: 4.5, Repeat: true, Domain: "biology",
		Overcome: func(w *World, c *Civ) {
			w.log("The sky burns over the worlds of the %s. Deep shelters hold. They come out to a dead surface and rebuild.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarBurningSky] = true
			for _, s := range w.blastWorlds {
				if s != c.Home {
					w.loseSystem(c, s, "scoured world", "")
				}
			}
			w.log("The sky burns over the worlds of the %s. %s holds; nothing else does. They never trust the sky again.", c.Name, c.HomeName)
		},
		Decline: func(w *World, c *Civ) {
			homeHit := contains(w.blastWorlds, c.Home)
			for _, s := range w.blastWorlds {
				if s != c.Home {
					w.loseSystem(c, s, "scoured world", "were sterilised by "+w.blastWhat)
				}
			}
			if !c.Active() {
				return
			}
			if homeHit {
				if len(c.Systems) > 1 && c.Reach >= 10 {
					w.leaveHome(c, "the burning sky")
				} else {
					w.loseSystem(c, c.Home, "scoured world", "were sterilised by "+w.blastWhat)
				}
			} else {
				w.contract(c, "lost their colonies to "+w.blastWhat+" and drew in")
			}
		},
	})
	def(&Filter{
		Key: "dying", Name: "the dying sun", Levels: []string{"sur"}, Diff: 5,
		Overcome: func(w *World, c *Civ) {}, Scar: func(w *World, c *Civ) {}, Decline: func(w *World, c *Civ) {},
	})
}
