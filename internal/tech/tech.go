// Package tech is the high-level technology tree shared by every species.
//
// Nodes are broad (Industrial Revolution, not the steam engine). Each node
// has prerequisites, effects on the three levels, on reach and on the
// habitable envelope, a tilt on later research, a structure it unlocks, and
// optionally a filter that fires when it is discovered. The words differ by
// species kind; the tree does not.
package tech

import (
	"sort"

	"worldgen/internal/flow"
)

// Domains of research.
const (
	Energy      = "energy"
	Industry    = "industry"
	Computation = "computation"
	Biology     = "biology"
	Society     = "society"
	Propulsion  = "propulsion"
	Weapons     = "weapons"
	Exotic      = "exotic"
)

// Domains lists every domain.
var Domains = []string{Energy, Industry, Computation, Biology, Society, Propulsion, Weapons, Exotic}

// M is a domain weight map.
type M = map[string]float64

// Node is one discovery.
type Node struct {
	Key, Name, Domain string
	Era               int
	Prereqs           []string
	Weight            float64 // discovery weight, 0 means 1
	Cost              float64 // research points to reach it; 0 means the era's default
	Patience          float64 // kyr a civilisation must have lived before it can pursue this
	Chance            float64 // if set, a finished pursuit only succeeds this often; failure wastes the work
	Miracle           bool    // a power apart from the tree: rare, potent, dangerous
	For               string  // only for these: "sub:machine", "mod:planetary", "world:iceshell", "trait:fireless"; "" for everyone
	Mil, Sur, Soc     float64
	Reach             float64 // reach in light years this node grants; the highest known wins
	Speed             float64 // colony ship speed in years per light year; the lowest known wins
	Env               int     // widens the habitable envelope
	Focus             M       // research tilt after discovery
	Filter            string  // filter key procced on discovery
	Structure         string  // structure key unlocked
	Gated             string  // a rarity without which the node cannot be learned at all; none is, and TestNoCatch22 keeps it so
	Milestone         bool    // worth a line in the legends
	Text              string  // legend text; %s is the civilisation name
	Desc              string  // what it is, in one line; see desc.go
}

// EraNames label the derived era of a civilisation.
var EraNames = []string{"pre-industrial", "industrial", "atomic", "interstellar", "exotic"}

// Nodes is the tree.
var Nodes = []*Node{
	// era 0
	{Key: "tools", Name: "Tools", Domain: Industry, Mil: 0.2},
	{Key: "fire", Name: "Fire", Domain: Energy, Sur: 0.3},
	{Key: "agriculture", Name: "Agriculture", Domain: Biology, Sur: 0.5, Soc: 0.1},
	{Key: "writing", Name: "Writing", Domain: Society, Soc: 0.2},
	{Key: "metallurgy", Name: "Metallurgy", Domain: Industry, Prereqs: []string{"fire", "tools"}, Mil: 0.5},
	{Key: "mathematics", Name: "Mathematics", Domain: Computation, Prereqs: []string{"writing"}},
	{Key: "star_gazing", Name: "Star Gazing", Domain: Exotic, Focus: M{Propulsion: 1.2}},
	{Key: "states", Name: "States", Domain: Society, Prereqs: []string{"agriculture", "writing"}, Soc: 0.3, Mil: 0.5},
	{Key: "seafaring", Name: "Seafaring", Domain: Propulsion, Prereqs: []string{"tools"}, Focus: M{Society: 1.1}},
	{Key: "burial", Name: "Burial", Domain: Society, Soc: 0.1},
	{Key: "song", Name: "Song", Domain: Society, Soc: 0.1},
	{Key: "religion", Name: "Religion", Domain: Society, Prereqs: []string{"burial"}, Soc: 0.2},
	{Key: "philosophy", Name: "Philosophy", Domain: Computation, Prereqs: []string{"writing"}, Focus: M{Computation: 1.2, Society: 1.2}},
	{Key: "law", Name: "Law", Domain: Society, Prereqs: []string{"states", "writing"}, Soc: 0.2},
	{Key: "organised_religion", Name: "Organised Religion", Domain: Society, Prereqs: []string{"religion", "states"}, Soc: 0.2, Mil: 0.2},
	// era 0, only for some
	{Key: "gathering", Name: "the Gathering", Domain: Society, For: "trait:swarming", Soc: 0.5, Mil: 0.5,
		Text: "The %s learn to gather on purpose: a million bodies, one mind, when it is wanted."},
	{Key: "host_craft", Name: "Host-craft", Domain: Biology, For: "sub:parasite", Sur: 0.5, Soc: 0.5},
	{Key: "maintenance", Name: "Maintenance", Domain: Industry, For: "sub:machine", Sur: 0.5},
	{Key: "husbandry", Name: "Husbandry of the Self", Domain: Biology, For: "mod:evolver", Sur: 0.5, Focus: M{Biology: 1.2}},
	{Key: "cold_chemistry", Name: "Cold Chemistry", Domain: Exotic, Era: 1, For: "trait:fireless",
		Text: "The %s learn to make without burning. What fire did for others, patience does for them."},
	// era 1
	{Key: "printing", Name: "Printing", Domain: Society, Era: 1, Prereqs: []string{"writing", "metallurgy"}, Soc: 0.3, Focus: M{Computation: 1.2, Biology: 1.1}},
	{Key: "scientific_method", Name: "the Scientific Method", Domain: Computation, Era: 1, Prereqs: []string{"mathematics", "philosophy", "printing"},
		Focus: M{Energy: 1.3, Biology: 1.3, Industry: 1.3, Exotic: 1.3}, Milestone: true,
		Text: "The %s learn to ask the world questions and to believe the answers."},
	{Key: "astronomy", Name: "Astronomy", Domain: Exotic, Era: 1, Prereqs: []string{"scientific_method", "star_gazing"}, Focus: M{Propulsion: 1.2, Exotic: 1.1}},
	{Key: "doubt", Name: "Doubt", Domain: Society, Era: 1, Prereqs: []string{"philosophy", "printing", "organised_religion"}, Soc: 0.1, Filter: "faith"},
	{Key: "steam", Name: "Steam Power", Domain: Energy, Era: 1, Prereqs: []string{"metallurgy", "scientific_method"}},
	{Key: "breach", Name: "the Breach", Domain: Industry, Era: 1, For: "world:iceshell", Prereqs: []string{"steam"}, Milestone: true,
		Text: "The %s drill up through the ice and break the shell of the world. There is a sky. There was always a sky."},
	{Key: "high_air", Name: "High Air", Domain: Propulsion, Era: 1, For: "world:hothouse", Prereqs: []string{"scientific_method"}, Milestone: true,
		Text: "A balloon of the %s rises above the poison and, for the first time, someone sees the stars."},
	{Key: "industrial", Name: "the Industrial Revolution", Domain: Industry, Era: 1, Prereqs: []string{"steam"}, Mil: 0.5, Sur: 0.5, Milestone: true,
		Text: "Smoke rises over the cities of the %s. Nothing is made by hand for long."},
	{Key: "chemistry", Name: "Chemistry", Domain: Industry, Era: 1, Prereqs: []string{"scientific_method"}},
	{Key: "medicine", Name: "Medicine", Domain: Biology, Era: 1, Prereqs: []string{"scientific_method"}, Sur: 0.5},
	{Key: "firearms", Name: "Firearms", Domain: Weapons, Era: 1, Prereqs: []string{"metallurgy", "chemistry"}, Mil: 0.5},
	{Key: "mass_politics", Name: "Mass Politics", Domain: Society, Era: 1, Prereqs: []string{"printing", "industrial", "doubt"}, Soc: 0.5},
	{Key: "electricity", Name: "Electricity", Domain: Energy, Era: 1, Prereqs: []string{"scientific_method", "industrial"}, Focus: M{Computation: 1.3}},
	// era 2
	{Key: "mass_industry", Name: "Mass Industry", Domain: Energy, Era: 2, Prereqs: []string{"industrial", "chemistry"}, Mil: 0.3, Sur: 0.3, Filter: "overshoot"},
	{Key: "mechanised_war", Name: "Mechanised War", Domain: Weapons, Era: 2, Prereqs: []string{"firearms", "industrial"}, Mil: 0.5},
	{Key: "physics", Name: "Modern Physics", Domain: Exotic, Era: 2, Prereqs: []string{"electricity", "mathematics", "astronomy"}, Focus: M{Energy: 1.3}},
	{Key: "atomic", Name: "Atomic Power", Domain: Energy, Era: 2, Prereqs: []string{"physics"}, Mil: 1, Filter: "atomic", Milestone: true,
		Text: "The %s split the atom."},
	{Key: "rocketry", Name: "Rocketry", Domain: Propulsion, Era: 2, Prereqs: []string{"chemistry", "mechanised_war"}, Reach: 0.1, Milestone: true,
		Text: "The %s put a machine into orbit and see their world whole for the first time."},
	{Key: "computers", Name: "Computers", Domain: Computation, Era: 2, Prereqs: []string{"electricity", "mathematics"}, Focus: M{Energy: 1.1, Biology: 1.1, Industry: 1.1}},
	{Key: "networks", Name: "Global Networks", Domain: Computation, Era: 2, Prereqs: []string{"computers", "mass_politics"}, Soc: 0.5},
	{Key: "genetics", Name: "Genetics", Domain: Biology, Era: 2, Prereqs: []string{"medicine", "chemistry"}, Sur: 0.5},
	{Key: "ecology", Name: "Ecology", Domain: Biology, Era: 2, Prereqs: []string{"medicine", "mass_industry"}, Sur: 0.5, Focus: M{Society: 1.2}},
	{Key: "orbital_weapons", Name: "Orbital Weapons", Domain: Weapons, Era: 2, Prereqs: []string{"rocketry", "atomic"}, Mil: 0.5},
	{Key: "fusion", Name: "Fusion Power", Domain: Energy, Era: 2, Prereqs: []string{"atomic", "computers"}, Mil: 0.3, Sur: 0.3, Milestone: true,
		Text: "The %s light a small star of their own and keep it burning."},
	{Key: "neuroscience", Name: "Neuroscience", Domain: Biology, Era: 2, Prereqs: []string{"medicine", "computers"}, Focus: M{Computation: 1.2}},
	{Key: "broodline", Name: "Broodline", Domain: Biology, Era: 2, For: "sub:parasite", Prereqs: []string{"host_craft", "chemistry"}, Sur: 0.5, Focus: M{Biology: 1.2}},
	{Key: "forking", Name: "Forking", Domain: Computation, Era: 2, For: "sub:machine", Prereqs: []string{"maintenance", "electricity"}, Sur: 0.5, Soc: 0.5},
	// era 3
	{Key: "machine_minds", Name: "Machine Minds", Domain: Computation, Era: 3, Prereqs: []string{"computers", "neuroscience"}, Soc: 0.5, Filter: "machines", Milestone: true,
		Focus: M{Energy: 1.2, Industry: 1.2, Biology: 1.2, Exotic: 1.2, Propulsion: 1.2, Weapons: 1.2},
		Text:  "The %s build a mind that is not one of theirs."},
	{Key: "closed_ecologies", Name: "Closed Ecologies", Domain: Biology, Era: 3, Prereqs: []string{"ecology", "genetics"}, Sur: 0.5, Env: 1, Structure: "arcology"},
	{Key: "orbital_habitats", Name: "Orbital Habitats", Domain: Industry, Era: 3, Prereqs: []string{"rocketry", "closed_ecologies"}, Sur: 0.5, Structure: "shipyard"},
	{Key: "interplanetary", Name: "Interplanetary Flight", Domain: Propulsion, Era: 3, Prereqs: []string{"rocketry", "fusion", "astronomy"}, Reach: 1},
	{Key: "slow_interstellar", Name: "Slow Interstellar Travel", Domain: Propulsion, Era: 3, Prereqs: []string{"interplanetary", "closed_ecologies"}, Reach: 12, Speed: 100, Milestone: true,
		Text: "The %s reach the stars. The first slow ships leave home, and will not arrive for centuries."},
	{Key: "self_replication", Name: "Self-Replicating Industry", Domain: Industry, Era: 3, Prereqs: []string{"machine_minds", "orbital_habitats"}, Mil: 0.5, Sur: 0.3, Filter: "replication", Milestone: true,
		Text: "The %s teach their machines to build themselves."},
	{Key: "life_extension", Name: "Life Extension", Domain: Biology, Era: 3, Prereqs: []string{"genetics", "neuroscience"}, Sur: 0.5, Filter: "silence"},
	{Key: "terraforming", Name: "Terraforming", Domain: Biology, Era: 3, Prereqs: []string{"closed_ecologies", "fusion"}, Env: 1, Sur: 0.5, Milestone: true,
		Text: "The %s remake a dead world in the image of their own."},
	{Key: "antimatter", Name: "Antimatter", Domain: Energy, Era: 3, Prereqs: []string{"fusion", "physics"}, Mil: 0.5},
	{Key: "defence_grid", Name: "Planetary Defence", Domain: Weapons, Era: 3, Prereqs: []string{"orbital_weapons", "computers"}, Mil: 0.5, Structure: "defences"},
	{Key: "memetics", Name: "Memetic Engineering", Domain: Society, Era: 3, Prereqs: []string{"networks", "neuroscience"}, Soc: 1},
	{Key: "relativistic", Name: "Relativistic Travel", Domain: Propulsion, Era: 3, Prereqs: []string{"antimatter", "slow_interstellar"}, Reach: 25, Speed: 4, Milestone: true,
		Text: "The ships of the %s now cross the dark at a good fraction of the speed of light."},
	{Key: "relativistic_weapons", Name: "Relativistic Weapons", Domain: Weapons, Era: 3, Prereqs: []string{"relativistic"}, Mil: 1.5, Milestone: true,
		Text: "The %s learn that a fast enough rock is the end of any world. Everything is a weapon now."},
	{Key: "uploading", Name: "Mind Uploading", Domain: Computation, Era: 3, Prereqs: []string{"machine_minds", "neuroscience"}, Soc: 0.5, Sur: 0.5},
	{Key: "dyson", Name: "Dyson Swarms", Domain: Industry, Era: 3, Prereqs: []string{"self_replication", "antimatter"}, Sur: 0.5, Structure: "dyson", Milestone: true,
		Text: "The %s begin to take their star apart for the light."},
	{Key: "germline", Name: "Germline Engineering", Domain: Biology, Era: 3, Prereqs: []string{"terraforming", "life_extension"}, Env: 1, Sur: 1},
	{Key: "quantum_computing", Name: "Quantum Computing", Domain: Computation, Era: 3, Prereqs: []string{"computers", "physics"}, Focus: M{Exotic: 1.3, Computation: 1.2}},
	{Key: "synthetic_biology", Name: "Synthetic Biology", Domain: Biology, Era: 3, Prereqs: []string{"genetics", "closed_ecologies"}, Sur: 0.5, Focus: M{Biology: 1.3}},
	{Key: "deep_governance", Name: "Deep Governance", Domain: Society, Era: 3, Prereqs: []string{"memetics", "networks", "law"}, Soc: 1},
	{Key: "beamed_sails", Name: "Beamed Sails", Domain: Propulsion, Era: 3, Prereqs: []string{"slow_interstellar", "orbital_habitats"}, Reach: 18, Speed: 30},
	{Key: "hibernation", Name: "Hibernation", Domain: Biology, Era: 3, Prereqs: []string{"medicine", "slow_interstellar"}, Sur: 0.5, Reach: 5},
	// era 3, only for some: other ways to the stars, other ways to last
	{Key: "seed_clouds", Name: "Seed-clouds", Domain: Propulsion, Era: 3, For: "trait:swarming", Prereqs: []string{"interplanetary", "gathering"}, Reach: 12, Speed: 300, Milestone: true,
		Text: "The %s reach the stars as spores: clouds of seed cast into the dark, and most of it lost, and enough of it not."},
	{Key: "living_ships", Name: "Living Ships", Domain: Propulsion, Era: 3, For: "mod:evolver", Prereqs: []string{"interplanetary", "synthetic_biology"}, Reach: 12, Speed: 100, Milestone: true,
		Text: "The %s breed the vacuum-whale: a body that crosses the dark on its own, with them asleep inside it."},
	{Key: "grafting", Name: "Grafting", Domain: Biology, Era: 3, For: "mod:planetary", Prereqs: []string{"slow_interstellar", "synthetic_biology"}, Sur: 0.5, Milestone: true,
		Text: "The %s learn to grow a piece of themselves on another world. It is the same mind. It always was."},
	{Key: "deep_root", Name: "Deep Root", Domain: Biology, Era: 3, For: "mod:planetary", Prereqs: []string{"closed_ecologies"}, Sur: 1},
	{Key: "free_living", Name: "Free-living", Domain: Biology, Era: 3, Cost: 60, For: "sub:parasite", Prereqs: []string{"broodline", "closed_ecologies"}, Sur: 0.5, Milestone: true,
		Text: "The %s learn to live without a host. It is a poorer life, and it can be lived anywhere."},
	// era 4: the deep tree. Each domain has a spine that costs a great deal
	// to climb, and no one climbs all of them.
	{Key: "stellar_engineering", Name: "Stellar Engineering", Domain: Exotic, Era: 4, Prereqs: []string{"dyson", "physics"}, Sur: 1, Mil: 1, Filter: "stellar", Structure: "tap", Milestone: true,
		Text: "The %s reach into their star."},
	{Key: "wormhole_physics", Name: "Wormhole Physics", Domain: Exotic, Era: 4, Prereqs: []string{"antimatter", "physics", "quantum_computing"}, Milestone: true,
		Text: "The %s prove that space can be folded. It is only a proof, for now."},
	{Key: "exotic_matter", Name: "Exotic Matter", Domain: Exotic, Era: 4, Cost: 500, Prereqs: []string{"wormhole_physics", "antimatter"}, Mil: 0.5, Sur: 0.5},
	{Key: "causal_physics", Name: "Causal Physics", Domain: Exotic, Era: 4, Cost: 600, Prereqs: []string{"wormhole_physics", "quantum_computing"}, Soc: 0.5, Focus: M{Exotic: 1.3}},
	{Key: "transcendence", Name: "Transcendence", Domain: Exotic, Era: 4, Cost: 500, Prereqs: []string{"uploading", "wormhole_physics", "memetics"}, Filter: "transcend"},
	{Key: "star_lifting", Name: "Star Lifting", Domain: Exotic, Era: 4, Cost: 500, Prereqs: []string{"stellar_engineering"}, Sur: 1, Structure: "lifter", Milestone: true,
		Text: "The %s learn to feed and drain their star. They will never need to fear it again."},
	{Key: "deep_time", Name: "Deep Time", Domain: Exotic, Era: 4, Cost: 600, Prereqs: []string{"star_lifting", "causal_physics"}, Patience: 4000, Chance: 0.3, Soc: 0.5, Milestone: true,
		Text: "The %s read the ages in the ash of dead stars and learn that the galaxy has done this before."},
	{Key: "vacuum_energy", Name: "Vacuum Energy", Domain: Energy, Era: 4, Prereqs: []string{"antimatter", "quantum_computing"}, Sur: 1, Mil: 0.5, Focus: M{Exotic: 1.2}},
	{Key: "matter_compilers", Name: "Matter Compilers", Domain: Industry, Era: 4, Prereqs: []string{"self_replication", "vacuum_energy"}, Sur: 1, Mil: 1},
	{Key: "world_engines", Name: "World Engines", Domain: Industry, Era: 4, Cost: 500, Prereqs: []string{"matter_compilers", "terraforming"}, Env: 1, Sur: 1},
	{Key: "substrate_minds", Name: "Substrate Minds", Domain: Computation, Era: 4, Prereqs: []string{"uploading", "quantum_computing"}, Soc: 1, Sur: 0.5, Focus: M{Society: 1.2}},
	{Key: "panspermia", Name: "Panspermia", Domain: Biology, Era: 4, Prereqs: []string{"synthetic_biology", "germline"}, Env: 1, Sur: 1},
	{Key: "posthuman_law", Name: "Posthuman Law", Domain: Society, Era: 4, Prereqs: []string{"deep_governance", "uploading"}, Soc: 1.5},
	{Key: "long_thought", Name: "the Long Thought", Domain: Society, Era: 4, Cost: 500, Prereqs: []string{"posthuman_law", "substrate_minds"}, Soc: 1},
	{Key: "near_light", Name: "Near-light Travel", Domain: Propulsion, Era: 4, Prereqs: []string{"relativistic", "vacuum_energy"}, Reach: 35, Speed: 1.5, Milestone: true,
		Text: "The ships of the %s run so close to light that a voyage is an afternoon inside and a lifetime outside."},
	{Key: "nova_bombs", Name: "Nova Bombs", Domain: Weapons, Era: 4, Prereqs: []string{"relativistic_weapons", "antimatter"}, Mil: 1.5},
	{Key: "stellar_weapons", Name: "Stellar Weapons", Domain: Weapons, Era: 4, Cost: 500, Prereqs: []string{"nova_bombs", "stellar_engineering"}, Mil: 2},

	// miracles: powers apart from the tree. Reached by a conscious leap
	// after a deep spine, or found and mastered, or born with. Each carries
	// its own filter, faced on the leap or the find but never by the born.
	{Key: "ansible", Name: "the Voice", Domain: Exotic, Era: 4, Miracle: true, Cost: 600, Prereqs: []string{"causal_physics", "substrate_minds"}, Structure: "ansible", Filter: "openline", Milestone: true,
		Text: "The %s make the leap. They can speak across any distance without delay. Their worlds are one world, and every mind among them is in the room."},
	{Key: "directed_evolution", Name: "the Flesh", Domain: Biology, Era: 4, Miracle: true, Cost: 600, Prereqs: []string{"panspermia", "life_extension"}, Filter: "brood", Milestone: true,
		Text: "The %s make the leap. They can remake themselves in a generation, and they do: for every world, a body."},
	{Key: "ftl", Name: "the Door", Domain: Propulsion, Era: 4, Miracle: true, Cost: 600, Prereqs: []string{"exotic_matter", "near_light"}, Reach: 45, Speed: 0.3, Filter: "door", Milestone: true,
		Text: "The %s make the leap. They tear a door in space, and the stars are next door."},
	{Key: "unmaking", Name: "the Unmaking", Domain: Weapons, Era: 4, Miracle: true, Cost: 600, Prereqs: []string{"exotic_matter", "stellar_weapons"}, Filter: "unmaking", Milestone: true,
		Text: "The %s make the leap. They can unmake matter at any distance they can see. Nothing can be defended against them."},
	{Key: "chorus", Name: "the Chorus", Domain: Society, Era: 4, Miracle: true, Cost: 600, Prereqs: []string{"long_thought", "memetics"}, Filter: "chorus", Milestone: true,
		Text: "The %s make the leap. Their thought takes root in any mind that hears it. Whoever meets them joins them."},
	{Key: "foresight", Name: "the Sight", Domain: Exotic, Era: 4, Miracle: true, Cost: 600, Prereqs: []string{"causal_physics", "long_thought"}, Filter: "sight", Milestone: true,
		Text: "The %s make the leap. They see what is coming, and they are never surprised again."},
	// the Ember and the Manna: miracles that make an object, with a form
	// and a price; the object is the thing, and its consequences are its
	// filter. See history/objects.go.
	{Key: "ember", Name: "the Ember", Domain: Energy, Era: 4, Miracle: true, Cost: 600, Prereqs: []string{"antimatter", "exotic_matter"}, Milestone: true,
		Text: "The %s make the leap. They kindle something that should not burn, and it burns for them."},
	{Key: "manna", Name: "the Manna", Domain: Biology, Era: 4, Miracle: true, Cost: 600, Prereqs: []string{"synthetic_biology", "germline"}, Milestone: true,
		Text: "The %s make the leap. They grow something that feeds them, and it does not stop."},
}

// EraCosts is the default research cost of a node by era.
var EraCosts = []float64{0.6, 2, 6, 30, 250}

// Price is the research cost of a node.
func (n *Node) Price() float64 {
	if n.Cost > 0 {
		return n.Cost
	}
	return EraCosts[n.Era]
}

// Subs are the stand-ins: a prerequisite is met by the node itself or by
// any of these. Cold Chemistry does what Fire does for the fireless; the
// swarm's Gathering is its state; a spore cloud or a living ship is a ship.
var Subs = map[string][]string{
	"fire":              {"cold_chemistry"},
	"states":            {"gathering"},
	"slow_interstellar": {"seed_clouds", "living_ships"},
	"genetics":          {"broodline", "forking"},
	"neuroscience":      {"forking"},
}

// Miracles lists the miracle nodes.
var Miracles []*Node

var byKey = map[string]*Node{}

func init() {
	for _, n := range Nodes {
		if n.Weight == 0 {
			n.Weight = 1
		}
		byKey[n.Key] = n
		if n.Miracle {
			Miracles = append(Miracles, n)
		}
	}
	for _, n := range Nodes {
		for _, p := range n.Prereqs {
			if byKey[p] == nil {
				panic("tech: unknown prerequisite " + p + " of " + n.Key)
			}
		}
		if n.Structure != "" && Structures[n.Structure] == nil {
			panic("tech: unknown structure " + n.Structure + " of " + n.Key)
		}
	}
	for _, st := range Structures {
		if byKey[st.Node] == nil {
			panic("tech: unknown node " + st.Node + " of structure " + st.Key)
		}
	}
}

// Get returns a node by key.
func Get(key string) *Node { return byKey[key] }

// Closure returns the node and everything it depends on, prerequisites first.
func Closure(key string) []string {
	seen := map[string]bool{}
	var out []string
	var walk func(k string)
	walk = func(k string) {
		if seen[k] {
			return
		}
		seen[k] = true
		for _, p := range byKey[k].Prereqs {
			walk(p)
		}
		out = append(out, k)
	}
	walk(key)
	sort.SliceStable(out, func(i, j int) bool { return byKey[out[i]].Era < byKey[out[j]].Era })
	return out
}

// Structure is something built at a star a people holds. It gives levels,
// or a yield from what is there, or both; it has an upkeep like any use
// and goes dark when it is shed. The yields themselves are calibration
// and live with the natural sources in history.
type Structure struct {
	Key, Name     string
	Node          string // the node that unlocks it
	Mil, Sur, Soc float64
	Upkeep        flow.Income   // per tick, while it works
	Cat           flow.Category // what it is fed under
	Per           string        // what it is one of: "star" (one per star), "belt" (one per belt), "" (two per people, one per star)
	Hardy         float64       // multiplier on decay once abandoned; less is hardier
	Text          string        // %s civ, %s star
}

// Structures by key.
var Structures = map[string]*Structure{
	"arcology":   {Key: "arcology", Name: "arcology", Node: "closed_ecologies", Sur: 1, Upkeep: flow.Income{flow.E: 1}, Cat: flow.Fields, Hardy: 1.2, Text: "The %s seal a city at %s against everything outside it."},
	"shipyard":   {Key: "shipyard", Name: "shipyard", Node: "orbital_habitats", Upkeep: flow.Income{flow.E: 1, flow.M: 1}, Cat: flow.Arms, Hardy: 1.6, Text: "Yards turn above %s, building ships for the %s."},
	"defences":   {Key: "defences", Name: "defence grid", Node: "defence_grid", Upkeep: flow.Income{flow.E: 1, flow.M: 1}, Cat: flow.Arms, Hardy: 0.7, Text: "The %s ring %s with guns that watch the sky."},
	"ansible":    {Key: "ansible", Name: "ansible net", Node: "ansible", Soc: 1.5, Upkeep: flow.Income{flow.E: 2}, Cat: flow.Mind, Hardy: 0.9, Text: "The %s link %s to home without delay."},
	"dyson":      {Key: "dyson", Name: "Dyson swarm", Node: "dyson", Sur: 0.5, Upkeep: flow.Income{flow.M: 2}, Cat: flow.Works, Per: "star", Hardy: 0.3, Text: "The %s enclose %s in a swarm of collectors. The star dims from outside."},
	"mine":       {Key: "mine", Name: "mines", Node: "orbital_habitats", Upkeep: flow.Income{flow.E: 1}, Cat: flow.Works, Per: "belt", Hardy: 1.4, Text: "The %s put mines in the belt at %s."},
	"collectors": {Key: "collectors", Name: "collectors", Node: "orbital_habitats", Upkeep: flow.Income{flow.M: 1}, Cat: flow.Works, Per: "star", Hardy: 0.5, Text: "The %s ring %s with collectors, and live on its light."},
	"tap":        {Key: "tap", Name: "accretion tap", Node: "stellar_engineering", Upkeep: flow.Income{flow.E: 1, flow.M: 2}, Cat: flow.Works, Per: "star", Hardy: 0.4, Text: "The %s ring the dead star at %s with a tap and draw on what falls in."},
	"lifter":     {Key: "lifter", Name: "star lifter", Node: "star_lifting", Upkeep: flow.Income{flow.M: 3}, Cat: flow.Works, Per: "star", Hardy: 0.4, Text: "The %s set a lifter on %s and take the star itself for metal."},
}

// StructureKeys is every structure in a fixed order, for loops that must
// not range over the map.
var StructureKeys = func() []string {
	var out []string
	for k := range Structures {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}()

// Yields says whether a structure harnesses a source rather than giving
// levels: the mines, the collectors, the swarm, the tap and the lifter.
func (s *Structure) Yields() bool { return s.Per != "" }
