package history

import (
	"sort"
	"strings"

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
	node  string // node key, or "domain:x" for every node of a domain
	when  string // trait key, "world:x", "kind:x"; a leading "!" negates
	mode  aptMode
	mult  float64 // dear: the multiplier; world: the penalty once lifted
	need  string  // world: what a colony must offer: sea, sky, fire
	until string  // world: a node that opens the way instead
	text  string  // said at birth, for never and world rows
}

func dear(node, when string, mult float64) apt { return apt{node: node, when: when, mult: mult} }
func innate(node, when string) apt             { return apt{node: node, when: when, mode: aptInnate} }
func moot(node, when string) apt               { return apt{node: node, when: when, mode: aptMoot} }
func never(node, when, text string) apt {
	return apt{node: node, when: when, mode: aptNever, text: text}
}
func world(node, when, need, until, text string) apt {
	return apt{node: node, when: when, mode: aptWorld, mult: 3, need: need, until: until, text: text}
}

// aptitudes is the table. Rows for the same node and people stack: dear
// rows multiply, and innate, never, moot and world rows take precedence in
// that order.
var aptitudes = []apt{
	// the cradle world
	world("seafaring", "world:arid", "sea", "", "There is no sea to put out on."),
	dear("agriculture", "world:arid", 1.5), dear("closed_ecologies", "world:arid", 0.8),
	dear("seafaring", "world:ocean", 0.5), dear("metallurgy", "world:ocean", 2), dear("industrial", "world:ocean", 1.5), dear("rocketry", "world:ocean", 1.3), dear("closed_ecologies", "world:ocean", 0.7),
	dear("star_gazing", "world:twilight", 0.7), dear("agriculture", "world:twilight", 1.3), dear("closed_ecologies", "world:twilight", 0.7), dear("terraforming", "world:twilight", 0.8),
	dear("rocketry", "world:superterran", 3), dear("interplanetary", "world:superterran", 1.5), dear("orbital_habitats", "world:superterran", 1.5),
	dear("rocketry", "world:lowg", 0.5), dear("interplanetary", "world:lowg", 0.7), dear("orbital_habitats", "world:lowg", 0.7), dear("closed_ecologies", "world:lowg", 0.8),
	world("star_gazing", "world:hothouse", "sky", "high_air", "The sky is a rumour under the poison, until someone rises above it."),
	dear("chemistry", "world:hothouse", 0.8), dear("closed_ecologies", "world:hothouse", 0.7),
	world("star_gazing", "world:iceshell", "sky", "breach", "There is no sky, and will be none until they dig for it."),
	dear("seafaring", "world:iceshell", 0.5), dear("rocketry", "world:iceshell", 2), dear("closed_ecologies", "world:iceshell", 0.6), dear("agriculture", "world:iceshell", 1.3),
	world("seafaring", "world:floater", "sea", "", "There is no sea and no ground, only the deeps."),
	dear("agriculture", "world:floater", 2), dear("metallurgy", "world:floater", 2), dear("rocketry", "world:floater", 3), dear("star_gazing", "world:floater", 1.5),
	dear("fire", "world:volcanic", 0.5), dear("metallurgy", "world:volcanic", 0.6), dear("chemistry", "world:volcanic", 0.8), dear("agriculture", "world:volcanic", 1.5), dear("star_gazing", "world:volcanic", 0.7),
	dear("star_gazing", "world:dim", 0.7), dear("fusion", "world:dim", 0.7), dear("stellar_engineering", "world:dim", 0.7), dear("agriculture", "world:dim", 1.5),
	dear("agriculture", "world:lush", 0.7),
	// what the world gave them
	world("fire", "fireless", "fire", "", "They will not make fire: nothing to burn, and nowhere to burn it."),
	dear("steam", "fireless", 1.5),
	dear("hibernation", "skyless", 0.8),
	dear("star_gazing", "threesuns", 0.5), dear("physics", "threesuns", 0.8), dear("stellar_engineering", "threesuns", 0.7), dear("agriculture", "threesuns", 1.3),
	dear("closed_ecologies", "hardy", 0.8), dear("states", "cooperative", 0.7), dear("deep_governance", "cooperative", 0.8),
	dear("rocketry", "robust", 1.2), dear("mechanised_war", "fragile", 1.3),
	// senses
	never("star_gazing", "eyeless", "They have no eyes for the stars."),
	dear("astronomy", "eyeless", 3), dear("physics", "eyeless", 2),
	moot("song", "deaf"), dear("networks", "deaf", 1.2),
	dear("fire", "thermal", 0.7), dear("star_gazing", "thermal", 1.3), dear("medicine", "thermal", 0.8),
	dear("seafaring", "magnetic", 0.5), dear("star_gazing", "magnetic", 0.8), dear("electricity", "magnetic", 0.7),
	dear("electricity", "electric", 0.5), dear("computers", "electric", 0.8), dear("networks", "electric", 0.8),
	dear("atomic", "radiation", 0.7), dear("fusion", "radiation", 0.8), dear("physics", "radiation", 0.8),
	dear("chemistry", "chemical", 0.6), dear("medicine", "chemical", 0.8), dear("genetics", "chemical", 0.8),
	// how they are organised
	dear("states", "solitary", 2), dear("mass_politics", "solitary", 2), dear("networks", "solitary", 1.5), dear("organised_religion", "solitary", 1.5), dear("deep_governance", "solitary", 1.5),
	dear("states", "individualist", 1.3), dear("deep_governance", "individualist", 1.5), dear("long_thought", "individualist", 1.5), dear("firearms", "individualist", 0.8),
	dear("states", "collective", 0.7), dear("mass_politics", "collective", 0.7), dear("deep_governance", "collective", 0.8),
	dear("states", "herd", 0.7), dear("mass_politics", "herd", 0.5), dear("organised_religion", "herd", 0.5), dear("memetics", "herd", 0.7),
	dear("states", "caste", 0.5), dear("mass_politics", "caste", 2), dear("posthuman_law", "caste", 0.7),
	moot("writing", "hive"), moot("printing", "hive"), moot("states", "hive"), moot("mass_politics", "hive"), moot("organised_religion", "hive"), moot("doubt", "hive"), moot("burial", "hive"),
	innate("networks", "hive"), dear("memetics", "hive", 2), dear("long_thought", "hive", 0.5),
	moot("song", "nonconscious"), moot("religion", "nonconscious"), moot("organised_religion", "nonconscious"), moot("doubt", "nonconscious"), moot("burial", "nonconscious"), moot("mass_politics", "nonconscious"), moot("memetics", "nonconscious"),
	dear("uploading", "nonconscious", 0.5), dear("transcendence", "nonconscious", 2), dear("philosophy", "nonconscious", 1.5),
	// stance
	dear("firearms", "pacifist", 2), dear("mechanised_war", "pacifist", 2), dear("orbital_weapons", "pacifist", 2), dear("relativistic_weapons", "pacifist", 3), dear("nova_bombs", "pacifist", 3),
	never("stellar_weapons", "pacifist", ""), never("unmaking", "pacifist", ""),
	dear("firearms", "conqueror", 0.6), dear("mechanised_war", "conqueror", 0.6), dear("orbital_weapons", "conqueror", 0.7), dear("defence_grid", "conqueror", 0.7),
	dear("defence_grid", "xenophobic", 0.5), dear("memetics", "xenophobic", 1.5),
	dear("defence_grid", "unyielding", 0.7), dear("firearms", "submissive", 1.3),
	// drive
	dear("star_gazing", "curious", 0.7), dear("scientific_method", "curious", 0.7), dear("wormhole_physics", "curious", 0.8),
	dear("machine_minds", "cautious", 1.5), dear("self_replication", "cautious", 1.5), dear("wormhole_physics", "cautious", 1.3), dear("stellar_engineering", "cautious", 1.5), dear("closed_ecologies", "cautious", 0.8), dear("defence_grid", "cautious", 0.8),
	dear("mathematics", "contemplative", 0.7), dear("philosophy", "contemplative", 0.5), dear("long_thought", "contemplative", 0.5), dear("deep_time", "contemplative", 0.7), dear("mechanised_war", "contemplative", 1.5),
	dear("seafaring", "expansionist", 0.7), dear("rocketry", "expansionist", 0.8), dear("slow_interstellar", "expansionist", 0.7), dear("beamed_sails", "expansionist", 0.8),
	dear("steam", "pragmatic", 0.8), dear("industrial", "pragmatic", 0.8), dear("mass_industry", "pragmatic", 0.8), dear("wormhole_physics", "pragmatic", 1.3), dear("transcendence", "pragmatic", 2), dear("philosophy", "pragmatic", 1.3),
	// bodies
	dear("seafaring", "sessile", 3), dear("rocketry", "sessile", 1.5), dear("slow_interstellar", "sessile", 1.5), dear("hibernation", "sessile", 0.5), dear("agriculture", "sessile", 0.5),
	dear("seafaring", "amphibious", 0.5), dear("closed_ecologies", "amphibious", 0.8),
	dear("genetics", "manysexes", 0.7), dear("germline", "manysexes", 0.7),
	dear("life_extension", "longlived", 0.5), dear("long_thought", "longlived", 0.6), dear("hibernation", "longlived", 1.5),
	dear("life_extension", "shortlived", 1.5), dear("long_thought", "shortlived", 2), dear("hibernation", "shortlived", 0.7),
	dear("hibernation", "dormancy", 0.3), dear("life_extension", "dormancy", 0.8),
	dear("writing", "memory", 0.5), dear("burial", "memory", 0.5), dear("deep_time", "memory", 0.7), dear("deep_governance", "memory", 0.7),
	dear("computers", "symbiosis", 0.6), dear("machine_minds", "symbiosis", 0.7), dear("uploading", "symbiosis", 0.7),
	dear("states", "eusocial", 0.7), dear("closed_ecologies", "eusocial", 0.8), dear("orbital_habitats", "eusocial", 0.8),
	// kinds
	moot("states", "kind:swarm"), moot("mass_politics", "kind:swarm"), moot("burial", "kind:swarm"),
	dear("networks", "kind:swarm", 0.5), dear("uploading", "kind:swarm", 2), dear("orbital_habitats", "kind:swarm", 0.7), dear("closed_ecologies", "kind:swarm", 0.8),
	moot("seafaring", "kind:planetary mind"), moot("states", "kind:planetary mind"), moot("mass_politics", "kind:planetary mind"), moot("burial", "kind:planetary mind"), moot("religion", "kind:planetary mind"),
	dear("terraforming", "kind:planetary mind", 0.5), dear("panspermia", "kind:planetary mind", 0.5), dear("life_extension", "kind:planetary mind", 0.5), dear("uploading", "kind:planetary mind", 2), dear("rocketry", "kind:planetary mind", 2), dear("networks", "kind:planetary mind", 0.3), dear("memetics", "kind:planetary mind", 2),
	moot("agriculture", "kind:parasite"), moot("states", "kind:parasite"), never("genetics", "kind:parasite", ""),
	dear("medicine", "kind:parasite", 0.6), dear("neuroscience", "kind:parasite", 0.6), dear("memetics", "kind:parasite", 0.6),
	innate("computers", "kind:machine-born"), innate("machine_minds", "kind:machine-born"), innate("hibernation", "kind:machine-born"),
	moot("fire", "kind:machine-born"), moot("agriculture", "kind:machine-born"), moot("medicine", "kind:machine-born"), moot("genetics", "kind:machine-born"), moot("neuroscience", "kind:machine-born"), moot("burial", "kind:machine-born"), moot("closed_ecologies", "kind:machine-born"),
	never("germline", "kind:machine-born", ""), never("directed_evolution", "kind:machine-born", ""),
	dear("synthetic_biology", "kind:machine-born", 1.5), dear("panspermia", "kind:machine-born", 2), dear("uploading", "kind:machine-born", 0.3), dear("self_replication", "kind:machine-born", 0.5), dear("terraforming", "kind:machine-born", 1.5),
	dear("domain:biology", "kind:evolver", 0.5), dear("metallurgy", "kind:evolver", 1.3), dear("industrial", "kind:evolver", 1.3), dear("self_replication", "kind:evolver", 1.5), dear("machine_minds", "kind:evolver", 1.5), dear("uploading", "kind:evolver", 2),
}

func init() {
	for _, a := range aptitudes {
		if strings.HasPrefix(a.node, "domain:") {
			continue
		}
		if tech.Get(a.node) == nil {
			panic("history: aptitude for unknown node " + a.node)
		}
		if a.until != "" && tech.Get(a.until) == nil {
			panic("history: aptitude until unknown node " + a.until)
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
	case strings.HasPrefix(when, "kind:"):
		return c.Species.Kind.String() == when[5:]
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
			var did string
			switch need {
			case "sea":
				did = "put to sea"
			case "sky":
				did = "look up, and see stars"
			case "fire":
				did = "make fire"
			}
			w.log("The %s of %s %s, to the bafflement of home. It never comes naturally to them.", c.Name, w.star(s), did)
			return true
		}
	}
	return false
}

// aptitude resolves the table for one people and one node: the mode and,
// for dear, the multiplier. A world block that has been lifted comes back
// as dear with its penalty; one that has not comes back as aptWorld.
func (w *World) aptitude(c *Civ, n *tech.Node) (aptMode, float64) {
	if n.For != "" && !c.applies(n.For) {
		return aptAbsent, 0
	}
	mode, mult := aptDear, 1.0
	rank := func(m aptMode) int {
		return map[aptMode]int{aptDear: 0, aptWorld: 1, aptMoot: 2, aptNever: 3, aptInnate: 4}[m]
	}
	for _, a := range aptitudes {
		if a.node != n.Key && a.node != "domain:"+n.Domain {
			continue
		}
		if !c.applies(a.when) {
			continue
		}
		switch a.mode {
		case aptDear:
			if !n.Miracle {
				mult *= a.mult // a miracle costs what it costs, whoever reaches for it
			}
		case aptWorld:
			if a.until != "" && w.had(c, a.until) {
				continue
			}
			if w.lifted(c, a.need) {
				mult *= a.mult
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
	return n.Price() * mult
}

// birthright gives a people its innate nodes and says at birth what its
// body and world do to the tree.
func (w *World) birthright(c *Civ) {
	var blocks, cheap, costly []string
	type ranked struct {
		name string
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
				rs = append(rs, ranked{n.Name, mult})
			}
		}
	}
	for _, a := range aptitudes {
		if a.text == "" || !c.applies(a.when) {
			continue
		}
		if a.mode == aptNever || (a.mode == aptWorld && !(a.until != "" && w.had(c, a.until)) && !w.lifted(c, a.need)) {
			blocks = append(blocks, a.text)
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
			cheap = append(cheap, r.name)
		}
		if r.mult > 1 && len(costly) < 3 {
			costly = append(costly, r.name)
		}
	}
	var parts []string
	parts = append(parts, blocks...)
	if len(cheap) > 0 {
		parts = append(parts, "They take to "+list(cheap)+" as if born to it.")
	}
	if len(costly) > 0 {
		parts = append(parts, strings.ToUpper(list(costly)[:1])+list(costly)[1:]+" will come hard to them.")
	}
	if len(parts) > 0 {
		w.log("%s", strings.Join(parts, " "))
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
