package history

import (
	"worldgen/internal/tech"
)

// A people never reads the world directly about anyone else. It reads its
// own intelligence: what it saw, where and when, with noise that widens as
// the report ages. Fog is not a special case in each decision but the only
// way anyone sees anyone.

// Intel is one people's latest belief about another.
type Intel struct {
	Mil    float64 // the level seen, with the noise of the seeing
	Grid   bool    // a defence grid was seen
	Relief float64 // strength of others seen standing with them at Star
	Star   int     // where the look was taken
	Year   Year
}

// look takes an observation without storing it.
func (w *World) look(c, e *Civ, star int, noise float64) *Intel {
	return &Intel{
		Mil:    e.Mil + w.R.NormFloat64()*noise,
		Grid:   e.Known["defence_grid"],
		Relief: w.reliefAt(e, star),
		Star:   star,
		Year:   w.Now,
	}
}

// observe stores a fresh observation of e by c.
func (w *World) observe(c, e *Civ, star int, noise float64) *Intel {
	i := w.look(c, e, star, noise)
	c.Intel[e.ID] = i
	return i
}

// receive stores a report from elsewhere if it is newer than what is held.
func (c *Civ) receive(about int, i *Intel) bool {
	if old := c.Intel[about]; old != nil && old.Year >= i.Year {
		return false
	}
	c.Intel[about] = i
	return true
}

// believe is what c thinks e's level is, and how sure: the spread widens
// half a level per five thousand years since the last look, to three.
func (w *World) believe(c, e *Civ) (mil, spread float64) {
	i := c.Intel[e.ID]
	if i == nil {
		// the lights of their cities are a guess at their age
		return 1 + 1.2*float64(e.Era), 3
	}
	age := float64(w.Now-i.Year) / 1000
	return i.Mil, min(3, 0.3+0.1*age)
}

// intelStep is what a people learns each tick without trying: trade
// partners see each other, and a people that watched a young species from
// orbit keeps watching.
func (w *World) intelStep(c *Civ) {
	for _, eid := range sortedInts(c.Met) {
		e := w.Civs[eid]
		if !e.Active() {
			continue
		}
		switch {
		case c.Trade[eid]:
			w.observe(c, e, e.Home, 0.3)
		case e.Era < 2 && len(e.held()) == 0 && w.G.Dist(c.Home, e.Home) <= c.Reach:
			w.observe(c, e, e.Home, 0.2)
		}
	}
}

// forward passes a report to allies, or not: loyalty times the ally's need
// against greed times the forwarder's own designs on the subject. The
// faithful forward nearly everything, the faithless only what serves them.
func (w *World) forward(c *Civ, about *Civ, i *Intel) {
	for _, pid := range c.Pacts {
		p := w.Pacts[pid]
		if p.Over {
			continue
		}
		for _, mid := range p.Members {
			m := w.Civs[mid]
			if m == c || !m.Active() {
				continue
			}
			need := 0.3
			if m.Wars[about.ID] || p.Target == about.ID {
				need = 1
			} else if m.Met[about.ID] {
				need = 0.5
			}
			designs := 0.0
			if c.hostile() && i.Mil < c.Mil {
				designs = 1
			}
			score := c.Dials.Loyalty*need - c.Dials.Greed*designs*0.6
			if score > 0.3 {
				w.send(c, m, &Message{Kind: MsgIntel, About: about.ID, Intel: i})
			}
		}
	}
}

// watchTable is how far a people sees fleets, in light years, by the tree.
var watchTable = []struct {
	node  string
	watch float64
}{
	{"astronomy", 2}, {"rocketry", 3}, {"computers", 5}, {"orbital_habitats", 8}, {"defence_grid", 12}, {"dyson", 15},
}

// watchRange is how far out a people sees a fleet coming.
func (c *Civ) watchRange() float64 {
	r := 0.0
	for _, row := range watchTable {
		if c.Known[row.node] {
			r = max(r, row.watch)
		}
	}
	if c.Known["star_gazing"] && r == 0 {
		r = 1
	}
	return r
}

var _ = tech.Get
