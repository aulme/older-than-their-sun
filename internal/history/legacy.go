package history

import (
	"strings"

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

var remainDescs = map[string]string{
	"arcology":    "a sealed city of the %s",
	"shipyard":    "the yards of the %s",
	"defences":    "the guns of the %s",
	"silos":       "the silos of the %s",
	"ansible":     "a relay of the %s",
	"dyson":       "the swarm of the %s",
	"mine":        "the mines of the %s",
	"collectors":  "the collectors of the %s",
	"tap":         "the tap of the %s, still ringing the dead star",
	"lifter":      "the lifter of the %s",
	"observatory": "the mirrors of the %s",
}

// relics: description and hardiness
var relicKinds = []struct {
	Desc  string
	Hardy float64
}{
	{"a vault of the %s", 0.4},
	{"the archives of the %s", 1.0},
	{"an engine of the %s", 0.8},
	{"a machine of the %s that no one dares to switch off", 0.8},
	{"the last workshop of the %s", 1.3},
}

// wreckages by filter: what a failed filter does to the works of the fallen.
var filterWreckage = map[string]Wreckage{
	"atomic":       {0.6, Wreck},
	"overshoot":    {0.4, Derelict},
	"machines":     {0.5, Derelict},
	"distance":     {0.1, Abandoned},
	"silence":      {0.3, Abandoned},
	"replication":  {0.8, Wreck},
	"stellar":      {1, Ruin},
	"transcend":    {0.1, Abandoned},
	"ossification": {0.2, Abandoned},
	"door":         {0.5, Wreck},
	"hold":         {0.3, Derelict},
	"revolt":       {0.5, Wreck},
	"containment":  {0.3, Derelict},
	"beacon":       {0.2, Abandoned},
	"incursion":    {0.3, Derelict},
	"elder":        {0.2, Abandoned},
}

// wreckages by manner of loss, used when no filter is running (war, cosmic
// events, things that eat). Keyed by the trace kind that loseSystem records.
var lossWreckage = map[string]Wreckage{
	"abandoned":               {0.1, Abandoned},
	"dead cities":             {0.3, Derelict},
	"transformed world":       {0.2, Abandoned},
	"host-world":              {0.3, Derelict},
	"glassed world":           {0.7, Wreck},
	"frozen world":            {0.6, Wreck},
	"scoured world":           {1, Ruin},
	"burned cradle":           {1, Ruin},
	"stripped world":          {1, Ruin},
	"wounded star":            {1, Ruin},
	"unmade world":            {1, Ruin},
	"absorbed world":          {0.5, Derelict},
	"silent world":            {0.2, Abandoned},
	"quarantined dead cities": {0.15, Abandoned},
	"world that believes":     {0.1, Abandoned},
}

var defaultWreckage = Wreckage{0.3, Derelict}

func wreckOf(filter string) *Wreckage {
	if wk, ok := filterWreckage[filter]; ok {
		return &wk
	}
	return nil
}

// wreckage decides what happens to works at a star being lost in this manner.
func (w *World) wreckage(kind string) Wreckage {
	if w.wreck != nil {
		return *w.wreck
	}
	if strings.HasPrefix(kind, "abandoned") {
		kind = "abandoned"
	}
	if wk, ok := lossWreckage[kind]; ok {
		return wk
	}
	return defaultWreckage
}

// leaveRuin decides the fate of a work at a star being lost.
func (w *World) leaveRuin(c *Civ, wk Work, kind string) {
	wr := w.wreckage(kind)
	if wk.Legacy >= 0 {
		// an inherited work goes back to the substrate, in whatever shape the ending left it
		l := w.Legacies[wk.Legacy]
		if w.R.Float64() < wr.Destroy {
			l.State = Lost
			return
		}
		l.State = Buried
		l.Star = wk.Star
		l.Cond = max(l.Cond, wr.Leave)
		return
	}
	st := tech.Structures[wk.Key]
	if st.Dug || w.R.Float64() < wr.Destroy {
		return // what was dug is spent: holes in the ground are nobody's find
	}
	l := &Legacy{ID: len(w.Legacies), Age: -1, Maker: c.ID, Kind: Structure, Star: wk.Star, Node: wk.Node, People: -1, Finder: -1, Source: -1, Plague: -1, Cond: wr.Leave, Hardy: st.Hardy}
	l.Desc = sprintf(remainDescs[wk.Key], c.Tok())
	w.Legacies = append(w.Legacies, l)
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
	rk := relicKinds[w.R.IntN(len(relicKinds))]
	l := &Legacy{ID: len(w.Legacies), Age: -1, Maker: c.ID, Kind: Artifact, Star: star, Node: node, People: -1, Finder: -1, Source: -1, Plague: -1, Cond: wr.Leave, Hardy: rk.Hardy}
	l.Desc = sprintf(rk.Desc, c.Tok())
	if n.Miracle {
		l.Desc = sprintf("what the %s left of %s", c.Tok(), n.Name)
		l.Hardy = 0.5
	}
	w.Legacies = append(w.Legacies, l)
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
				l.State = Lost
			} else {
				l.Cond++
			}
		}
	}
}

// Describe gives a legacy's description with its condition; a field with
// the ships in it.
func (l *Legacy) Describe() string {
	if l.Maker < 0 {
		return l.Desc
	}
	if l.Kind == Field {
		s := l.Desc
		if n := l.ships(); n > 0 {
			s += ", " + shipsWord(n)
		} else {
			s = "what is left of " + l.Desc
		}
		if l.Adrift {
			s += ", adrift"
		}
		return s
	}
	switch l.Cond {
	case Abandoned:
		return l.Desc + ", abandoned but whole"
	case Derelict:
		return l.Desc + ", derelict"
	case Wreck:
		return "the wreck of " + l.Desc
	default:
		return "the ruin of " + l.Desc
	}
}

// condAdj is the Find's difficulty adjustment for a legacy's condition:
// negative is easier.
func (l *Legacy) condAdj() float64 {
	if l.Maker < 0 {
		return 0
	}
	switch l.Cond {
	case Abandoned:
		return -1
	case Wreck:
		return 1.5
	case Ruin:
		return 3
	}
	return 0
}

// makerName is what a finder calls the makers of a legacy: the elder's
// finder name, the maker's own if the finder knows of them, else the
// finder's name for whoever they were.
func (w *World) makerName(c *Civ, l *Legacy) string {
	if l.Elder != nil {
		return l.Elder.Tok()
	}
	m := w.Civs[l.Maker]
	if c.Met[m.ID] || m.Living() || w.kinship(c, l) > 0 {
		return "the " + m.Tok()
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
		l.State = Wielded
		l.Finder = c.ID
		key := tech.Get(l.Node).Structure()
		c.Works = append(c.Works, Work{Key: key, Node: l.Node, Star: star, Legacy: l.ID})
		c.Structures[key]++
		if w.kinship(c, l) == 2 {
			w.log("The %s return to %s and put their own old works there back to use.", c.Tok(), w.star(star))
		} else if w.R.Float64() < 0.2 {
			w.log("The %s find %s at %s, and put it back to work.", c.Tok(), l.Describe(), w.star(star))
		}
	}
}
