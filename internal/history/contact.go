package history

import (
	"worldgen/internal/names"
	"worldgen/internal/species"
)

// Contact happens when reach spheres overlap. What follows depends on stance
// traits and relative military: peace and trade, submission, war, and after
// war either extermination, enslavement or contraction.

func (w *World) contacts() {
	for i, a := range w.Civs {
		if !a.Active() {
			continue
		}
		for j := i + 1; j < len(w.Civs); j++ {
			b := w.Civs[j]
			if !b.Active() || (a.Met[b.ID] && b.Met[a.ID]) {
				continue
			}
			if w.G.Dist(a.Home, b.Home) > a.Reach+b.Reach || a.Reach+b.Reach < 1 {
				continue
			}
			// a young species found by an old one is not a contact between equals
			// a people holding a miracle is nobody's primitive, whatever their era
			if young, old := a, b; (young.Era < 2 && len(young.held()) == 0) || (old.Era < 2 && len(old.held()) == 0) {
				if old.Era < 2 && len(old.held()) == 0 {
					young, old = b, a
				}
				if (young.Era >= 2 || len(young.held()) > 0) || old.Met[young.ID] {
					continue // both young, or already watched
				}
				if !w.primitives(old, young) {
					continue // watched from orbit; they will meet properly later
				}
			}
			a.Met[b.ID], b.Met[a.ID] = true, true
			w.encounter(a, b)
		}
	}
}

// primitives: an old civilisation finds a pre-atomic one. Returns true if
// something happened that counts as their meeting.
func (w *World) primitives(old, young *Civ) bool {
	switch {
	case old.Has("xenophobic") && w.R.Float64() < 0.15:
		old.Met[young.ID], young.Met[old.ID] = true, true
		w.log("The %s find the %s on %s before they have looked up, and scour the world clean. They are thorough.", old.Name, young.Name, young.HomeName)
		w.Bio[young.Home] = BioSimple
		w.endCiv(young, Extinct, sprintf("were scoured from %s by the %s before they had looked up", young.HomeName, old.Name))
		return true
	case (old.Has("expansionist") || old.Has("martial")) && w.R.Float64() < 0.3:
		old.Met[young.ID], young.Met[old.ID] = true, true
		w.log("The %s find the %s on %s, still at the plough, and take them. There is no war to speak of.", old.Name, young.Name, young.HomeName)
		w.enslave(old, young)
		return true
	}
	if !old.Met[young.ID] {
		old.Met[young.ID] = true // one-sided: the old know, the young do not
		w.log("The %s find the %s on %s, still young, and watch from orbit.", old.Name, young.Name, young.HomeName)
	}
	return false
}

func hostile(c *Civ) bool { return c.Has("xenophobic") || c.Has("expansionist") || c.Has("martial") }

func warChance(c *Civ) float64 {
	p := 0.05
	if c.Has("xenophobic") {
		p += 0.35
	}
	if c.Has("expansionist") {
		p += 0.2
	}
	if c.Has("martial") {
		p += 0.2
	}
	if c.Has("fighttodeath") {
		p += 0.1
	}
	if c.Has("contemplative") || c.Has("submissive") {
		p -= 0.1
	}
	return p
}

func (w *World) encounter(a, b *Civ) {
	// the stronger side is the one with the initiative
	if b.Mil > a.Mil {
		a, b = b, a
	}
	if a.Species.Kind == species.Parasite || b.Species.Kind == species.Parasite {
		w.infection(a, b)
		return
	}
	gap := a.Mil - b.Mil
	switch {
	case a.miracle("chorus") && !b.miracle("chorus") && !b.Has("hive") && !b.Has("nonconscious") && w.R.Float64() < 0.6:
		w.log("The %s find the %s, and speak. Within a generation the %s ask to be ruled.", a.Name, b.Name, b.Name)
		w.vassal(a, b)
	case b.miracle("chorus") && !a.miracle("chorus") && !a.Has("hive") && !a.Has("nonconscious") && w.R.Float64() < 0.6:
		w.log("The %s find the %s, and the %s speak. Within a generation the %s ask to be ruled.", a.Name, b.Name, b.Name, a.Name)
		w.vassal(b, a)
	case a.miracle("unmaking") && hostile(b) && !b.miracle("unmaking"):
		w.log("The %s meet the %s and learn what they hold. There is no war. The %s bend the knee.", b.Name, a.Name, b.Name)
		w.vassal(a, b)
	case b.Has("pacifist") && hostile(a) && gap >= 1:
		w.log("The %s find the %s, who will not fight. They are taken without a war.", a.Name, b.Name)
		w.enslave(a, b)
	case a.Has("pacifist") && hostile(b) && b.Mil >= a.Mil-1:
		w.log("The %s find the %s, who will not fight. They are taken without a war.", b.Name, a.Name)
		w.enslave(b, a)
	case b.Has("submissive") && hostile(a) && gap >= 2:
		w.log("The %s meet the %s, and seeing what they face, bend the knee. They are vassals now.", b.Name, a.Name)
		w.vassal(a, b)
	case !a.Has("pacifist") && !b.Has("pacifist") && w.R.Float64() < warChance(a)+warChance(b):
		a.Wars[b.ID], b.Wars[a.ID] = true, true
		w.log("The %s and the %s find each other. The first strikes are launched within a century.", a.Name, b.Name)
	default:
		a.Trade[b.ID], b.Trade[a.ID] = true, true
		w.log("The %s and the %s find each other. Slow messages cross the dark between them for generations, and then trade.", a.Name, b.Name)
		if (a.Faced["plague"] || b.Faced["plague"]) && w.R.Float64() < 0.3 {
			a.Plagued, b.Plagued = true, true
			w.log("Something crosses with the messages and the trade. Both the %s and the %s begin to sicken.", a.Name, b.Name)
		}
	}
}

// infection: a parasite meets a host species.
func (w *World) infection(a, b *Civ) {
	p, h := a, b
	if p.Species.Kind != species.Parasite {
		p, h = b, a
	}
	if h.Species.Kind == species.Parasite || h.Species.Kind == species.MachineBorn {
		w.log("The %s and the %s find each other, and find nothing in the other worth having.", a.Name, b.Name)
		return
	}
	w.log("The %s find the %s. Within a generation the %s are inside them.", p.Name, h.Name, p.Name)
	switch w.face(h, "infection", 0) {
	case Overcome:
		h.Wars[p.ID], p.Wars[h.ID] = true, true
	case Scarred:
		lost := 0
		for _, s := range append([]int(nil), h.Systems...) {
			if s != h.Home && w.R.Float64() < 0.5 {
				w.loseSystem(h, s, "host-world", "")
				w.Owner[s] = p.ID
				p.Systems = append(p.Systems, s)
				lost++
			}
		}
		w.log("The %s burn %d of their own worlds to stop it. It stops.", h.Name, lost)
		h.Wars[p.ID], p.Wars[h.ID] = true, true
	case Declined:
		worlds := append([]int(nil), h.Systems...)
		w.endCiv(h, Transformed, sprintf("were taken from within by the %s", p.Name))
		h.Into = "hosts of the " + p.Name
		for _, s := range worlds {
			w.Owner[s] = p.ID
			p.Systems = append(p.Systems, s)
		}
		p.Peak = max(p.Peak, len(p.Systems))
		w.log("The %s are still there, but they are the %s now.", h.Name, p.Name)
	}
}

func (w *World) enslave(m, s *Civ) {
	m.Ruled++
	s.Master, s.Vassal = m.ID, false
	s.Seen = m.Declines
	s.Voyages = nil
	delete(m.Wars, s.ID)
	delete(s.Wars, m.ID)
	// colonies pass to the master
	for _, x := range append([]int(nil), s.Systems...) {
		if x != s.Home {
			s.Systems = remove(s.Systems, x)
			w.Owner[x] = m.ID
			m.Systems = append(m.Systems, x)
		}
	}
	m.Peak = max(m.Peak, len(m.Systems))
	s.Morale -= 1
	w.log("The %s are enslaved by the %s. They keep %s and little else.", s.Name, m.Name, s.HomeName)
}

func (w *World) vassal(m, s *Civ) {
	m.Ruled++
	s.Master, s.Vassal = m.ID, true
	s.Seen = m.Declines
	delete(m.Wars, s.ID)
	delete(s.Wars, m.ID)
}

// war: each side has a chance per tick to win a battle and take a world.
// Losing the last colony puts the home at stake and ends the war.
func (w *World) war(c *Civ) {
	for _, eid := range sortedInts(c.Wars) {
		e := w.Civs[eid]
		if !e.Active() {
			delete(c.Wars, eid)
			continue
		}
		c.Morale -= 0.03 * w.dt
		if w.chance(0.02) && len(c.Systems) > 1 && len(e.Systems) > 1 {
			delete(c.Wars, eid)
			delete(e.Wars, c.ID)
			w.log("The war between the %s and the %s ends. Neither side is sure who won.", c.Name, e.Name)
			continue
		}
		if !w.chance(0.04) {
			continue
		}
		home := 0.0
		if len(e.Systems) == 1 {
			home = 1.5 // the last world is defended like the last world
		}
		if c.Mil+c.warBonus()+w.R.NormFloat64()*1.5 < e.Mil+e.warBonus()+home+w.R.NormFloat64()*1.5 {
			continue // the strike is answered; the other side gets its own turn
		}
		w.battleWon(c, e)
	}
}

func (w *World) battleWon(c, e *Civ) {
	if len(e.Systems) > 1 {
		t := e.Systems[w.R.IntN(len(e.Systems))]
		if t == e.Home {
			return
		}
		if c.miracle("unmaking") {
			w.Bio[t] = BioNone
			w.loseSystem(e, t, "unmade world", "")
			w.log("The %s unmake %s, a %s of the %s. There is nothing left to glass.", c.Name, w.star(t), e.Species.Kind.Flavour().Colony, e.Name)
		} else {
			w.Bio[t] = BioSimple
			w.loseSystem(e, t, "glassed world", "")
			if w.R.Float64() < 0.4 {
				w.log("The %s glass %s, a %s of the %s.", c.Name, w.star(t), e.Species.Kind.Flavour().Colony, e.Name)
			}
		}
		w.face(e, "hold", 0)
		return
	}
	// the home is all that is left
	delete(c.Wars, e.ID)
	delete(e.Wars, c.ID)
	switch {
	case c.miracle("unmaking") && !c.Has("pacifist") && (e.Has("fighttodeath") || w.R.Float64() < 0.5):
		w.Bio[e.Home] = BioNone
		w.log("The %s unmake %s, homeworld of the %s. It is not there any more.", c.Name, e.HomeName, e.Name)
		w.endCiv(e, Extinct, sprintf("were unmade by the %s", c.Name))
	case e.Has("fighttodeath") || (c.Has("xenophobic") && w.R.Float64() < 0.7):
		w.Bio[e.Home] = BioNone
		w.log("A relativistic strike from the %s shatters %s, homeworld of the %s. They never surrendered.", c.Name, e.HomeName, e.Name)
		w.endCiv(e, Extinct, sprintf("were annihilated in war with the %s", c.Name))
	case c.Has("pacifist"):
		w.log("The %s defeat the %s and, having no use for a conquest, leave them be.", c.Name, e.Name)
	default:
		w.log("The %s break the last defences of %s.", c.Name, e.HomeName)
		w.enslave(c, e)
	}
}

// revolt: slaves and vassals watch their master. A master's decline is the
// slaves' chance; a master's death forces the question.
func (w *World) revolt(c *Civ) {
	if c.Free() || !c.Active() {
		return
	}
	m := w.Civs[c.Master]
	if m.Living() && m.Declines == c.Seen {
		return
	}
	c.Seen = m.Declines
	adj := 0.0
	if c.Vassal {
		adj = -1
	}
	if !m.Living() {
		adj -= 1
		w.log("The %s, who held the %s, are gone. The question of freedom answers itself, one way or the other.", m.Name, c.Name)
	}
	w.face(c, "revolt", adj+w.holdDiff(c))
}

// uplift: a strong civilisation makes a new species from complex life
// within its reach. The client relationship goes the way of vassalage.
func (w *World) uplift(c *Civ) {
	if !c.Active() || c.Era < 3 || c.Soc < 5 || !c.Free() || c.Uplifts >= 2 || !(c.Has("curious") || c.Has("collective") || c.Has("contemplative")) || !w.chance(0.0001) {
		return
	}
	for _, t := range w.G.Near(c.Home, c.Reach) {
		if w.Bio[t] == BioComplex && w.Owner[t] < 0 && w.Held[t] < 0 && t != w.G.Sol {
			sp := species.Generate(w.R, w.G.Stars[t].Mult)
			sp.Add("uplifted")
			sp.Made = "uplifted by the " + c.Name
			c.Uplifts++
			nc := w.spawnCiv(t, sp, c.ID)
			nc.Vassal = true
			nc.Seen = c.Declines
			for _, k := range knownOf(c) {
				if w.R.Float64() < 0.5 {
					nc.Known[k] = true
				}
			}
			w.forget(nc, 0.3)
			w.recompute(nc)
			w.log("The %s raise the %s from the beasts of %s. They are %s, and grateful, for now.", c.Name, nc.Name, w.star(t), sp.Describe())
			return
		}
	}
}

// breed turns a slave people into something the master wants: a made species.
func (w *World) breed(m, s *Civ) {
	sp := *s.Species
	sp.Name = names.Civ(w.R)
	sp.Traits = append([]*species.Trait(nil), s.Species.Traits...)
	sp.Add("bred")
	sp.Made = "bred by the " + m.Name + " from the " + s.Name
	home := s.Home
	w.endCiv(s, Transformed, sprintf("were bred by the %s into something else", m.Name))
	s.Into = "the " + sp.Name
	nc := w.spawnCiv(home, &sp, m.ID)
	nc.Seen = m.Declines
	w.log("The %s remake the %s into the %s: %s.", m.Name, s.Name, nc.Name, sp.Describe())
}

func init() {
	def(&Filter{
		Key: "infection", Name: "Infection", Levels: []string{"sur"}, Diff: 5, Domain: "biology",
		Overcome: func(w *World, c *Civ) {
			w.log("The %s find the thing inside them and cut it out. Then they go looking for where it came from.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarQuarantine] = true
		},
		Decline: func(w *World, c *Civ) {},
	})
}
