package history

import (
	"worldgen/internal/tech"
)

// Remains: what the current age leaves for its own successors. A
// civilisation that loses a world leaves its works there. One that falls,
// or forgets, leaves relics of what it knew. Both are Legacy records, the
// same as the elder ages leave, and the Find handles them the same way: a
// young people living in the halls of the old without knowing how to raise
// them (wielding), or a renaissance that reads the old works and recovers
// the arts behind them (mastering). A people that went through a dark age
// can find its own works, and something in them remembers.
//
// How a thing ended decides how much of it survives and in what condition;
// time then wears the survivors down a step at a time, faster for the
// fragile and the precariously placed. What is hardy enough outlasts the
// age and becomes, by survivorship, an elder legacy of the next.

// wreckOf is what a filter's decline does to the works of the fallen,
// from the filter's row; nil for the default.
func wreckOf(filter string) *Wreckage {
	if f := filters[filter]; f != nil {
		return f.Wreckage
	}
	return nil
}

// wreckage decides what happens to works at a star being lost in this manner.
func (w *World) wreckage(kind string) Wreckage {
	if w.wreck != nil {
		return *w.wreck
	}
	if wk, ok := tables.losses[kind]; ok {
		return wk
	}
	return tables.defaultWr
}

// leaveRuin decides the fate of a work at a star being lost.
func (w *World) leaveRuin(c *Civ, wk Work, kind string) {
	wr := w.wreckage(kind)
	if wk.Legacy >= 0 {
		// an inherited work goes back to the substrate, in whatever shape the ending left it
		l := w.Legacies[wk.Legacy]
		if w.R.Float64() < wr.Destroy {
			w.setState(l, Lost)
			return
		}
		w.moveRemain(l, wk.Star)
		w.setState(l, Buried)
		w.setCond(l, max(l.Cond, wr.Leave))
		return
	}
	st := tech.Structures[wk.Key]
	if st.Dug || w.R.Float64() < wr.Destroy {
		return // what was dug is spent: holes in the ground are nobody's find
	}
	l := &Legacy{ID: len(w.Legacies), Age: -1, Maker: c.ID, Kind: Structure, Star: wk.Star, Node: wk.Node, People: -1, Finder: -1, Source: -1, Plague: -1, Cond: wr.Leave, Hardy: st.Hardy}
	l.Portrait = wk.Key
	w.addLegacy(l)
	w.testament(c, l)
}

// leaveRelic leaves an artifact of one late thing a civilisation knew, at a
// star it holds, so that a successor or its own descendants can find it.
func (w *World) leaveRelic(c *Civ, node string, star int) {
	n := tech.Get(node)
	if n == nil || n.Era < 2 {
		return
	}
	wr := w.wreckage("")
	if w.R.Float64() < wr.Destroy {
		return
	}
	relics := tables.portraits.Relics[:len(tables.portraits.Relics)-1] // the last is the miracle's
	rk := relics[w.R.IntN(len(relics))]
	if n.Miracle {
		rk = tables.portraits.Relics[len(tables.portraits.Relics)-1]
	}
	l := &Legacy{ID: len(w.Legacies), Age: -1, Maker: c.ID, Kind: Artifact, Star: star, Node: node, People: -1, Finder: -1, Source: -1, Plague: -1, Cond: wr.Leave, Hardy: rk.Hardy, Portrait: rk.Key}
	w.addLegacy(l)
	w.testament(c, l)
}

// lateNode picks something a civilisation knows from its highest era.
func (w *World) lateNode(c *Civ) string {
	var best []string
	era := -1
	for _, k := range knownOf(c) {
		n := tech.Get(k)
		if n.Era > era {
			best, era = nil, n.Era
		}
		if n.Era == era {
			best = append(best, k)
		}
	}
	if era < 3 {
		return ""
	}
	return best[w.R.IntN(len(best))]
}

// tickLegacies wears the remains down a step at a time. The base rate is
// one step per 2.5 Myr on average; hardiness scales it. Elder legacies have
// hardiness 0 and never decay, which is what makes them elder.
func (w *World) tickLegacies() {
	for _, l := range w.Legacies {
		if l.Hardy <= 0 || l.State != Buried {
			continue
		}
		if w.chance(0.0004 * l.Hardy) {
			if l.Cond == Ruin {
				w.setState(l, Lost)
			} else {
				w.setCond(l, l.Cond+1)
			}
		}
	}
}

// condAdj is the Find's difficulty adjustment for a legacy's condition:
// negative is easier.
func (l *Legacy) condAdj() float64 {
	if l.Maker < 0 {
		return 0
	}
	return tables.conditions[l.Cond].Find
}

// variant is what a law does: its portrait's variant in the table.
func (l *Legacy) variant() string {
	if p := tables.portraitByKey["laws"][l.Portrait]; p != nil {
		return p.Variant
	}
	return ""
}

// makerName is what a finder calls the makers of a legacy: the elder's
// finder name, the maker's own if the finder knows of them, else the
// finder's name for whoever they were.
func (w *World) makerName(c *Civ, l *Legacy) string {
	return w.makerNameIf(l, w.knowsMaker(c, l))
}

// knowsMaker says whether a finder can place a remain's makers: an elder
// always has a name, a people if the finder has met them, they live, or
// the finder is of their line or blood.
func (w *World) knowsMaker(c *Civ, l *Legacy) bool {
	if l.Elder != nil {
		return true
	}
	m := w.Civs[l.Maker]
	return c.Met[m.ID] || m.Living() || w.kinship(c, l) > 0
}

// makerNameIf is what a finder calls the makers, given whether it can
// place them.
func (w *World) makerNameIf(l *Legacy, known bool) string {
	if l.Elder != nil {
		return l.Elder.Tok()
	}
	if known {
		return "the " + w.Civs[l.Maker].Tok()
	}
	return makersTok(l)
}

// kinship: 2 for one's own works and the works of one's line (the design
// is theirs), 1 for those of the same blood, else 0.
func (w *World) kinship(c *Civ, l *Legacy) int {
	if l.Maker < 0 {
		return 0
	}
	m := w.Civs[l.Maker]
	switch {
	case c.ofLine(m.ID):
		return 2
	case m.Species.Kin(c.Species):
		return 1
	}
	return 0
}

// takeOver: a people settling a star put any remains there whose art they
// know back to work, if they are in a state to be used. No Find, no test.
func (w *World) takeOver(c *Civ, star int) {
	for _, l := range w.Legacies {
		if l.Maker < 0 || l.Kind != Structure || l.Star != star || l.State != Buried || !c.Known[l.Node] || l.Cond > Derelict {
			continue
		}
		w.setState(l, Wielded)
		w.setFinder(l, c.ID)
		key := tech.Get(l.Node).Structure()
		c.Works = append(c.Works, Work{Key: key, Node: l.Node, Star: star, Legacy: l.ID})
		c.Structures[key]++
		if w.kinship(c, l) == 2 {
			w.event(KTakenOver, c, nil, star, P{"own": true, "cond": int(l.Cond)}).Legacy = l.ID
		} else if w.R.Float64() < 0.2 {
			w.event(KTakenOver, c, nil, star, P{"own": false, "cond": int(l.Cond)}).Legacy = l.ID
		}
	}
}
