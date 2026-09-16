package history

import (
	"worldgen/internal/tech"
)

// Ruins and relics: what the current age leaves for its own successors.
// A civilisation that loses a world leaves its works there as ruins. One
// that falls, or forgets, leaves relics of what it knew. Both are Legacy
// records, the same as the elder ages leave, and the Find handles them the
// same way: a young people living in the halls of the old without knowing
// how to raise them (wielding), or a renaissance that reads the old works
// and recovers the arts behind them (mastering). A people that went through
// a dark age can find its own works, and something in them remembers.

var ruinDescs = map[string]string{
	"arcology": "a sealed city of the %s, its air still good",
	"shipyard": "the yards of the %s, hanging dark",
	"defences": "the guns of the %s, still watching the sky",
	"ansible":  "a relay of the %s that answers, faintly, when spoken to",
	"dyson":    "the swarm of the %s, half its mirrors still turned to the star",
}

var relicDescs = []string{
	"a vault of the %s",
	"the archives of the %s, in a script nobody reads",
	"an engine of the %s, still warm",
	"a machine of the %s that no one dares to switch off",
	"the last workshop of the %s, sealed when its makers went quiet",
}

var ruinNames = []string{"the Old Builders", "the Ones Before", "the First People", "the Builders of the Halls", "the Ones Who Left the Lights On"}

// leaveRuin turns a work at a lost star into a structure legacy.
func (w *World) leaveRuin(c *Civ, wk Work) {
	if wk.Legacy >= 0 {
		// an inherited work goes back to being a ruin
		l := w.Legacies[wk.Legacy]
		l.State = Buried
		l.Star = wk.Star
		return
	}
	// most of what is abandoned is stripped or falls within a few centuries
	if w.R.Float64() > 0.1 {
		return
	}
	l := &Legacy{ID: len(w.Legacies), Age: -1, Maker: c.ID, Kind: Structure, Star: wk.Star, Node: wk.Node, Horror: -1, Finder: -1}
	l.Desc = sprintf(ruinDescs[wk.Key], c.Name)
	w.Legacies = append(w.Legacies, l)
}

// leaveRelic leaves an artifact of one late thing a civilisation knew, at a
// star it holds, so that a successor or its own descendants can find it.
func (w *World) leaveRelic(c *Civ, node string, star int) {
	n := tech.Get(node)
	if n == nil || n.Era < 2 {
		return
	}
	l := &Legacy{ID: len(w.Legacies), Age: -1, Maker: c.ID, Kind: Artifact, Star: star, Node: node, Horror: -1, Finder: -1}
	l.Desc = sprintf(relicDescs[w.R.IntN(len(relicDescs))], c.Name)
	w.Legacies = append(w.Legacies, l)
}

// lateNode picks something a civilisation knows from its highest era.
func (w *World) lateNode(c *Civ) string {
	var best []string
	era := -1
	for k := range c.Known {
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

// tickLegacies: what this age leaves erodes far faster than what the elder
// ages left, since it was built to last centuries, not aeons.
func (w *World) tickLegacies() {
	for _, l := range w.Legacies {
		if l.Maker < 0 || l.State != Buried {
			continue
		}
		if w.chance(0.0005) {
			l.State = Lost
		}
	}
}

// makerName is what a finder calls the makers of a legacy.
func (w *World) makerName(l *Legacy) string {
	if l.Elder != nil {
		return l.Elder.Name
	}
	if l.Name != "" {
		return l.Name
	}
	return "the " + w.Civs[l.Maker].Name
}

// kinship: 2 for one's own works, 1 for those of the same species, else 0.
func (w *World) kinship(c *Civ, l *Legacy) int {
	if l.Maker < 0 {
		return 0
	}
	m := w.Civs[l.Maker]
	switch {
	case m == c:
		return 2
	case m.Species.Name == c.Species.Name || m.Species.Made == "a branch of the "+c.Name || c.Species.Made == "a branch of the "+m.Name:
		return 1
	}
	return 0
}

// takeOver: a people settling a star put any ruin there whose art they know
// back to work. No Find, no test; it is simply theirs now.
func (w *World) takeOver(c *Civ, star int) {
	for _, l := range w.Legacies {
		if l.Maker < 0 || l.Kind != Structure || l.Star != star || l.State != Buried || !c.Known[l.Node] {
			continue
		}
		l.State = Wielded
		l.Finder = c.ID
		key := tech.Get(l.Node).Structure
		c.Works = append(c.Works, Work{Key: key, Node: l.Node, Star: star, Legacy: l.ID})
		c.Structures[key]++
		if w.kinship(c, l) == 2 {
			w.log("The %s return to %s and put their own old works there back to use.", c.Name, w.star(star))
		} else if w.R.Float64() < 0.2 {
			w.log("The %s find %s at %s, and put it back to work.", c.Name, l.Desc, w.star(star))
		}
	}
}
