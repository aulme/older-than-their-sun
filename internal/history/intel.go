package history

import (
	"worldgen/internal/mind"
	"worldgen/internal/tech"
)

// A people never reads the world directly about anyone else. It reads its
// own intelligence: what it saw, where and when, with noise that widens as
// the report ages. Fog is not a special case in each decision but the only
// way anyone sees anyone.

// Intel is one people's latest belief about another.
type Intel struct {
	Mil    float64 // the level seen, with the noise of the seeing
	Ships  int     // the guard seen in the sky at Star
	Guns   int     // the guns seen over Star
	Total  int     // the ships seen manned everywhere, as far as the look could tell
	Relief float64 // ships of others seen standing with them at Star
	Star   int     // where the look was taken
	Year   Year
	Sick   string // a plague seen raging in them, by name; "" for none
}

// look takes an observation without storing it.
func (w *World) look(c, e *Civ, star int, noise float64) *Intel {
	i := &Intel{
		Mil:    e.Mil + w.R.NormFloat64()*noise,
		Sick:   w.sickSeen(e),
		Guns:   w.gunsAt(e, star),
		Total:  w.standing(e),
		Relief: float64(w.reliefAt(e, star)),
		Star:   star,
		Year:   w.Now,
	}
	if g := w.guardAt(e, star); g != nil && !g.LaidUp {
		i.Ships = g.Ships
	}
	return i
}

// observe stores a fresh observation of e by c.
func (w *World) observe(c, e *Civ, star int, noise float64) *Intel {
	i := w.look(c, e, star, noise)
	c.Intel[e.ID] = i
	return i
}

// observeDark is what a people learns of another from a battle between
// fleets with no world in sight: the level, and the ships manned
// everywhere as far as the fight could tell.
func (w *World) observeDark(c, e *Civ) *Intel {
	i := &Intel{Mil: e.Mil + w.R.NormFloat64()*0.3, Total: w.standing(e), Star: -1, Year: w.Now}
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
	in := mind.BeliefInput{EnemyEra: e.Era}
	if i := c.Intel[e.ID]; i != nil {
		in.Seen, in.Mil, in.AgeKyr = true, i.Mil, float64(w.Now-i.Year)/1000
	}
	return mind.Believe(in, w.Cfg.Tuning)
}

// believeShips is how many ships c thinks e has: the last count, or a
// guess by era.
func (w *World) believeShips(c, e *Civ) float64 {
	in := mind.BeliefInput{EnemyEra: e.Era}
	if i := c.Intel[e.ID]; i != nil {
		in.Seen, in.Ships = true, i.Total
	}
	return mind.BelieveShips(in, w.Cfg.Tuning)
}

// believeSky is what c thinks stands in e's sky at a star: what was seen
// there, or elsewhere the whole believed force and the guns a world of
// that era is expected to have.
func (w *World) believeSky(c, e *Civ, star int) (ships, guns, relief float64) {
	body := float64(w.bodyGuns(e)) // a world that is the people is plain to see
	if i := c.Intel[e.ID]; i != nil && i.Star == star {
		return float64(i.Ships), max(float64(i.Guns), body), i.Relief
	}
	return w.believeShips(c, e), mind.BelieveGuns(e.Era, w.Cfg.Tuning) + body, 0
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
			in := mind.ForwardInput{AllyAtWar: m.Wars[about.ID] || p.Target == about.ID, AllyMet: m.Met[about.ID], Hostile: c.hostile(), Reported: i.Mil, Mil: c.Mil, Dials: c.Dials}
			if ok, _ := mind.Forward(in, w.Cfg.Tuning); ok {
				w.send(c, m, &Message{Kind: MsgIntel, About: about.ID, Intel: i})
			}
		}
	}
}

// watchTable is how far a people's worlds see fleets, in light years, by
// the tree; a grid or a swarm sees farther, but only where it stands
// (tech.Structure.Watch), and an observatory farthest of all.
var watchTable = []struct {
	node  string
	watch float64
}{
	{"astronomy", 2}, {"rocketry", 3}, {"computers", 5}, {"orbital_habitats", 8},
}

// watchRange is how far out a people's worlds see a fleet coming, by the
// tree alone; watchAt adds the works.
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
