package history

import (
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Guns are an immobile fleet in a world's sky: the defence grid's three and
// more, and the silos every atomic people digs before it has ships. They
// stand at the builder's quality, are broken before any fleet at the world
// and never withdraw; a world with a gun standing cannot be taken. A gun
// is remade in place, one a thousand years, with no dock and no build
// cost: the grid through a siege, silos only between sieges, since a silo
// fired is a silo spent.

// silosHome and silosColony are how fast a people that knows the art digs
// silos, per thousand years: within a millennium at the home, a few at a
// colony.
const (
	silosHome   = 1.0
	silosColony = 0.3
)

// gunRepair is guns remade per thousand years at a world still held.
const gunRepair = 1.0

// gunsOf is how many guns a gun structure stands for a people: its count,
// and for a modern one a gun more per weapons era known beyond its node.
func (w *World) gunsOf(c *Civ, st *tech.Structure) int {
	n := st.Guns
	if !st.Modern {
		return n
	}
	era := tech.Get(st.Node).Era
	eras := map[int]bool{}
	for _, k := range knownOf(c) {
		if nd := tech.Get(k); nd.Domain == tech.Weapons && nd.Era > era {
			eras[nd.Era] = true
		}
	}
	return n + len(eras)
}

// bodyGuns is the body as guns: for a people that is its world (or wears
// a shell), so many guns over every world it holds per level of Military,
// regrown a gun a thousand years like a grid, through a siege.
func (w *World) bodyGuns(c *Civ) int {
	d := c.Species.Profile().HomeDefence
	if d == 0 {
		return 0
	}
	return int(d*(1+c.Mil) + 0.5)
}

// gunCap is the guns that could stand over a star: the fed gun structures
// there, and the body for a people that is its world. With repairing
// set, only those that may be remade now: the grid and the body always,
// silos only while no enemy fleet is in the sky.
func (w *World) gunCap(c *Civ, star int, repairing bool) int {
	n := w.bodyGuns(c)
	for _, wk := range c.Works {
		if wk.Star != star || wk.Dark {
			continue
		}
		st := tech.Structures[wk.Key]
		if st == nil || st.Guns == 0 {
			continue
		}
		if repairing && !st.Repair && w.sieged(c, star) {
			continue
		}
		n += w.gunsOf(c, st)
	}
	return n
}

// gunsAt is the guns standing over a star: what was not shot away, of
// what the fed structures there stand.
func (w *World) gunsAt(c *Civ, star int) int {
	return min(c.Guns[star], w.gunCap(c, star, false))
}

// gridAt says whether a working defence grid stands at a star.
func (w *World) gridAt(c *Civ, star int) bool {
	for _, wk := range c.Works {
		if wk.Star == star && wk.Key == "defences" && !wk.Dark {
			return true
		}
	}
	return false
}

// addGuns raises the guns standing over a star.
func (w *World) addGuns(c *Civ, star, n int) {
	if c.Guns == nil {
		c.Guns = map[int]int{}
	}
	c.Guns[star] += n
}

// sieged says whether an enemy's campaign fleet is in a star's sky.
func (w *World) sieged(c *Civ, star int) bool {
	for _, x := range w.liveFleets() {
		if x.Kind == Campaign && x.Target == c.ID && x.Base == star && !x.LaidUp && x.Ships > 0 {
			return true
		}
	}
	return false
}

// dig is the civ step that raises silos: a people that knows the art digs
// them at its home within a millennium and at every colony within a few
// thousand years, never by the pick.
func (w *World) dig(c *Civ) {
	st := tech.Structures["silos"]
	if c.Aloft || !c.Known[st.Node] || !c.working(st.Node) || !c.Species.Profile().Can(species.Works) {
		return
	}
	for _, s := range c.Systems {
		if c.worksAt("silos", s) > 0 {
			continue
		}
		rate := silosColony
		if s == c.Home {
			rate = silosHome
		}
		if w.count(rate) == 0 {
			continue
		}
		w.reserve(c, st.Upkeep)
		w.raise(c, "silos", st.Node, s)
	}
}

// guns is the civ step that keeps the guns: what stands is never more
// than the fed structures there, and a gun is remade a thousand years
// where the world is still held, the grid through a siege and silos only
// between them. A grid shot to nothing and whole again is a line.
func (w *World) guns(c *Civ) {
	if c.Aloft {
		return
	}
	for _, s := range c.Systems {
		full := w.gunCap(c, s, false)
		if c.Guns[s] > full {
			c.Guns[s] = full
		}
		allowed := w.gunCap(c, s, true)
		if c.Guns[s] >= allowed {
			continue
		}
		w.addGuns(c, s, min(allowed-c.Guns[s], w.count(gunRepair)))
		if c.Guns[s] == full && c.GridBroken[s] {
			delete(c.GridBroken, s)
			w.log("The %s rebuild the grid over %s.", c.Tok(), w.star(s))
		}
	}
}

// GunsOf is a people's guns standing and the worlds they stand over, for
// the legends.
func GunsOf(w *World, c *Civ) (guns, worlds int) {
	for _, s := range c.Systems {
		if g := w.gunsAt(c, s); g > 0 {
			guns += g
			worlds++
		}
	}
	return
}
