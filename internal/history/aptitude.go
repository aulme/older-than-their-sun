package history

import (
	"sort"
	"strings"

	"worldgen/data"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Aptitudes are what a people's body, world and nature do to the tree. A
// node can be dear or cheap to them, innate (had from birth, with its
// effects and no filter), moot (counted as had for prerequisites and
// otherwise not on their tree at all), never (impossible because of what
// they are, counted as had so the tree goes on, with the dependents made
// dear by their own rows), or blocked by the cradle world, which lifts once
// a colony offers what the cradle lacked, at a price for being
// counter-intuitive, or once a named node opens the way.

type aptMode int

const (
	aptDear   aptMode = iota // a price multiplier; the default with multiplier 1
	aptInnate                // had from birth with its effects; its filter never fires
	aptMoot                  // counts as had for prerequisites; no effects; not on their tree
	aptNever                 // cannot, ever; counts as had for prerequisites
	aptWorld                 // blocked by the cradle; lifts at a price once a colony offers the need, or when until is known
	aptAbsent                // not on their tree at all: a node only some peoples have, and they are not one of them
)

type apt struct {
	Node  string  `json:"node"`  // node key, or "domain:x" for every node of a domain
	When  string  `json:"when"`  // trait key, "trait:x", "world:x", "sub:x", "mod:x"; a leading "!" negates
	Mode  string  `json:"mode"`  // "" for dear, else innate, moot, never, world, absent
	Mult  float64 `json:"mult"`  // dear: the multiplier; world: the penalty once lifted
	Need  string  `json:"need"`  // world: what a colony must offer: sea, sky, fire
	Until string  `json:"until"` // world: a node that opens the way instead
	Key   string  `json:"key"`   // the row said at birth, for never and world rows; its text is the view's
	Text  string  `json:"text"`
	mode  aptMode
}

var aptModes = map[string]aptMode{"": aptDear, "innate": aptInnate, "moot": aptMoot, "never": aptNever, "world": aptWorld, "absent": aptAbsent}

// aptitudes is the table, data/aptitudes.json. Rows for the same node and
// people stack: dear rows multiply, and innate, never, moot and world rows
// take precedence in that order.
var aptitudes = func() []apt {
	var f struct {
		Aptitudes []apt `json:"aptitudes"`
	}
	data.Load("aptitudes.json", &f)
	for i := range f.Aptitudes {
		a := &f.Aptitudes[i]
		m, ok := aptModes[a.Mode]
		if !ok {
			panic("history: aptitude " + a.Node + " " + a.When + " has an unknown mode " + a.Mode)
		}
		a.mode = m
		if m == aptWorld {
			a.Mult = 3
		}
	}
	return f.Aptitudes
}()

// aptitudeText is the text of a birthright block, by its row's key.
func aptitudeText(key string) string {
	for _, a := range aptitudes {
		if a.Key == key {
			return a.Text
		}
	}
	return key
}

func init() {
	for _, a := range aptitudes {
		if strings.HasPrefix(a.Node, "domain:") {
			continue
		}
		if tech.Get(a.Node) == nil {
			panic("history: aptitude for unknown node " + a.Node)
		}
		if a.Until != "" && tech.Get(a.Until) == nil {
			panic("history: aptitude until unknown node " + a.Until)
		}
	}
}

// applies says whether a condition holds for a people.
func (c *Civ) applies(when string) bool {
	if strings.HasPrefix(when, "!") {
		return !c.applies(when[1:])
	}
	switch {
	case strings.HasPrefix(when, "world:"):
		return c.Species.World.Key == when[6:]
	case strings.HasPrefix(when, "sub:"):
		return c.Species.Sub.String() == when[4:]
	case strings.HasPrefix(when, "mod:"):
		m, ok := species.ModByKey(when[4:])
		return ok && c.Species.Is(m)
	case strings.HasPrefix(when, "trait:"):
		return c.Has(when[6:])
	}
	return c.Has(when)
}

// offers says whether a world of an archetype has what a need asks for. An
// unknown archetype is a made-habitable rock: air, ground and sky, and sea
// enough.
func offers(arch, need string) bool {
	switch need {
	case "sea":
		switch arch {
		case "ocean", "lush", "twilight", "superterran", "lowg", "dim", "iceshell", "":
			return true
		}
		return false
	case "sky":
		return arch != "iceshell" && arch != "hothouse"
	case "fire":
		return arch != "floater" && arch != "iceshell"
	}
	return false
}

// lifted says whether a colony beyond the cradle offers what the cradle
// lacked, and says so once in the legends.
func (w *World) lifted(c *Civ, need string) bool {
	if c.Lifted[need] {
		return true
	}
	for _, s := range c.Systems {
		if s == c.Cradle {
			continue
		}
		if offers(w.G.Sys[s].Arch, need) {
			c.Lifted[need] = true
			w.event(KLifted, c, nil, s, P{"need": need})
			return true
		}
	}
	return false
}

// rank orders the modes for precedence when rows stack.
func rank(m aptMode) int {
	switch m {
	case aptWorld:
		return 1
	case aptMoot:
		return 2
	case aptNever:
		return 3
	case aptInnate:
		return 4
	}
	return 0
}

// aptitude resolves the table for one people and one node: the mode and,
// for dear, the multiplier. A world block that has been lifted comes back
// as dear with its penalty; one that has not comes back as aptWorld.
func (w *World) aptitude(c *Civ, n *tech.Node) (aptMode, float64) {
	if n.For != "" && !c.applies(n.For) {
		return aptAbsent, 0
	}
	mode, mult := aptDear, 1.0
	domain := "domain:" + n.Domain
	for _, a := range aptitudes {
		if a.Node != n.Key && a.Node != domain {
			continue
		}
		if !c.applies(a.When) {
			continue
		}
		switch a.mode {
		case aptDear:
			if !n.Miracle {
				mult *= a.Mult // a miracle costs what it costs, whoever reaches for it
			}
		case aptWorld:
			if a.Until != "" && w.had(c, a.Until) {
				continue
			}
			if w.lifted(c, a.Need) {
				mult *= a.Mult
				continue
			}
			if rank(aptWorld) > rank(mode) {
				mode = aptWorld
			}
		default:
			if rank(a.mode) > rank(mode) {
				mode = a.mode
			}
		}
	}
	return mode, mult
}

// had says whether a node counts as known: known, or innate, moot or never.
func (w *World) had(c *Civ, key string) bool {
	if c.Known[key] {
		return true
	}
	n := tech.Get(key)
	if n == nil {
		return false
	}
	m, _ := w.aptitude(c, n)
	return m == aptInnate || m == aptMoot || m == aptNever
}

// met says whether a prerequisite is satisfied, by the node or a stand-in.
func (w *World) met(c *Civ, key string) bool {
	if w.had(c, key) {
		return true
	}
	for _, s := range tech.Subs[key] {
		if c.Known[s] { // a stand-in must really be held; what one can never have stands in for nothing
			return true
		}
	}
	return false
}

// price is what a node costs this people.
func (w *World) price(c *Civ, n *tech.Node) float64 {
	_, mult := w.aptitude(c, n)
	if c.Grants[n.Key] {
		mult *= 0.5 // a rarity had: the horizon for the deep physics, the ash for exotic matter
	}
	return n.Price() * mult
}

// birthright gives a people its innate nodes and says at birth what its
// body and world do to the tree.
func (w *World) birthright(c *Civ) {
	if !c.Species.Profile().Can(species.Researches) {
		return // no tree to take to
	}
	var blocks, cheap, costly []string
	type ranked struct {
		key  string
		mult float64
	}
	var rs []ranked
	for _, n := range tech.Nodes {
		if n.For != "" && !c.applies(n.For) {
			continue
		}
		mode, mult := w.aptitude(c, n)
		switch mode {
		case aptInnate:
			if !c.Known[n.Key] {
				w.learn(c, n, false)
			}
		case aptDear:
			if !n.Miracle && (mult <= 0.5 || mult >= 2) {
				rs = append(rs, ranked{n.Key, mult})
			}
		}
	}
	for _, a := range aptitudes {
		if a.Key == "" || !c.applies(a.When) {
			continue
		}
		if a.mode == aptNever || (a.mode == aptWorld && !(a.Until != "" && w.had(c, a.Until)) && !w.lifted(c, a.Need)) {
			blocks = append(blocks, a.Key)
		}
	}
	sort.Slice(rs, func(i, j int) bool {
		di, dj := rs[i].mult, rs[j].mult
		if di < 1 {
			di = 1 / di
		}
		if dj < 1 {
			dj = 1 / dj
		}
		return di > dj
	})
	for _, r := range rs {
		if r.mult < 1 && len(cheap) < 3 {
			cheap = append(cheap, r.key)
		}
		if r.mult > 1 && len(costly) < 3 {
			costly = append(costly, r.key)
		}
	}
	if len(blocks)+len(cheap)+len(costly) > 0 {
		w.event(KBirthright, c, nil, -1, P{"blocks": blocks, "cheap": cheap, "costly": costly})
	}
}

// list joins names as prose: "a, b and c".
func list(xs []string) string {
	switch len(xs) {
	case 0:
		return ""
	case 1:
		return xs[0]
	}
	return strings.Join(xs[:len(xs)-1], ", ") + " and " + xs[len(xs)-1]
}
