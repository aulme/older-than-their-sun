// Package tech is the high-level technology tree shared by every species.
//
// Nodes are broad (Industrial Revolution, not the steam engine). Each node
// has prerequisites, effects on the three levels, on reach and on the
// habitable envelope, a tilt on later research, a structure it unlocks, and
// optionally a filter that fires when it is discovered. The words differ by
// species kind; the tree does not.
package tech

import "sort"

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
	Mil, Sur, Soc     float64
	Reach             float64 // reach in light years this node grants; the highest known wins
	Speed             float64 // colony ship speed in years per light year; the lowest known wins
	Env               int     // widens the habitable envelope
	Focus             M       // research tilt after discovery
	Filter            string  // filter key procced on discovery
	Structure         string  // structure key unlocked
	Milestone         bool    // worth a line in the legends
	Text              string  // legend text; %s is the civilisation name
}

// EraNames label the derived era of a civilisation.
var EraNames = []string{"pre-industrial", "industrial", "atomic", "interstellar", "exotic"}

// Nodes is the tree.
var Nodes = []*Node{
	// era 0
	{Key: "tools", Name: "Tools", Domain: Industry, Mil: 0.2},
	{Key: "fire", Name: "Fire", Domain: Energy, Sur: 0.3},
	{Key: "agriculture", Name: "Agriculture", Domain: Biology, Sur: 0.5, Soc: 0.3},
	{Key: "writing", Name: "Writing", Domain: Society, Soc: 0.5},
	{Key: "metallurgy", Name: "Metallurgy", Domain: Industry, Prereqs: []string{"fire", "tools"}, Mil: 0.5},
	{Key: "mathematics", Name: "Mathematics", Domain: Computation, Prereqs: []string{"writing"}},
	{Key: "astronomy", Name: "Astronomy", Domain: Exotic, Prereqs: []string{"mathematics"}, Focus: M{Propulsion: 1.2}},
	{Key: "states", Name: "States", Domain: Society, Prereqs: []string{"agriculture", "writing"}, Soc: 0.5, Mil: 0.5},
	{Key: "seafaring", Name: "Seafaring", Domain: Propulsion, Prereqs: []string{"tools"}, Focus: M{Society: 1.1}},
	{Key: "cold_chemistry", Name: "Cold Chemistry", Domain: Exotic, Weight: 0.3, Era: 1,
		Text: "The %s learn to make without burning. What fire did for others, patience does for them."},
	// era 1
	{Key: "printing", Name: "Printing", Domain: Society, Era: 1, Prereqs: []string{"writing", "metallurgy"}, Soc: 0.3, Focus: M{Computation: 1.2, Biology: 1.1}},
	{Key: "scientific_method", Name: "the Scientific Method", Domain: Computation, Era: 1, Prereqs: []string{"mathematics", "astronomy", "printing"},
		Focus: M{Energy: 1.3, Biology: 1.3, Industry: 1.3, Exotic: 1.3}, Milestone: true,
		Text: "The %s learn to ask the world questions and to believe the answers."},
	{Key: "steam", Name: "Steam Power", Domain: Energy, Era: 1, Prereqs: []string{"metallurgy", "scientific_method"}},
	{Key: "industrial", Name: "the Industrial Revolution", Domain: Industry, Era: 1, Prereqs: []string{"steam"}, Mil: 0.5, Sur: 0.5, Milestone: true,
		Text: "Smoke rises over the cities of the %s. Nothing is made by hand for long."},
	{Key: "chemistry", Name: "Chemistry", Domain: Industry, Era: 1, Prereqs: []string{"scientific_method"}},
	{Key: "medicine", Name: "Medicine", Domain: Biology, Era: 1, Prereqs: []string{"scientific_method"}, Sur: 0.5},
	{Key: "firearms", Name: "Firearms", Domain: Weapons, Era: 1, Prereqs: []string{"metallurgy", "chemistry"}, Mil: 0.5},
	{Key: "mass_politics", Name: "Mass Politics", Domain: Society, Era: 1, Prereqs: []string{"printing", "industrial"}, Soc: 0.5},
	{Key: "electricity", Name: "Electricity", Domain: Energy, Era: 1, Prereqs: []string{"scientific_method", "industrial"}, Focus: M{Computation: 1.3}},
	// era 2
	{Key: "mass_industry", Name: "Mass Industry", Domain: Energy, Era: 2, Prereqs: []string{"industrial", "chemistry"}, Mil: 0.3, Sur: 0.3, Filter: "overshoot"},
	{Key: "mechanised_war", Name: "Mechanised War", Domain: Weapons, Era: 2, Prereqs: []string{"firearms", "industrial"}, Mil: 0.5},
	{Key: "physics", Name: "Modern Physics", Domain: Exotic, Era: 2, Prereqs: []string{"electricity", "mathematics"}, Focus: M{Energy: 1.3}},
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
	// era 3
	{Key: "machine_minds", Name: "Machine Minds", Domain: Computation, Era: 3, Prereqs: []string{"computers", "neuroscience"}, Soc: 0.5, Filter: "machines", Milestone: true,
		Focus: M{Energy: 1.2, Industry: 1.2, Biology: 1.2, Exotic: 1.2, Propulsion: 1.2, Weapons: 1.2},
		Text: "The %s build a mind that is not one of theirs."},
	{Key: "closed_ecologies", Name: "Closed Ecologies", Domain: Biology, Era: 3, Prereqs: []string{"ecology", "genetics"}, Sur: 0.5, Env: 1, Structure: "arcology"},
	{Key: "orbital_habitats", Name: "Orbital Habitats", Domain: Industry, Era: 3, Prereqs: []string{"rocketry", "closed_ecologies"}, Sur: 0.5, Structure: "shipyard"},
	{Key: "interplanetary", Name: "Interplanetary Flight", Domain: Propulsion, Era: 3, Prereqs: []string{"rocketry", "fusion"}, Reach: 1},
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
	{Key: "directed_evolution", Name: "Directed Evolution", Domain: Biology, Era: 3, Prereqs: []string{"terraforming", "life_extension"}, Env: 1, Sur: 1.5},
	// era 4
	{Key: "stellar_engineering", Name: "Stellar Engineering", Domain: Exotic, Era: 4, Prereqs: []string{"dyson", "physics"}, Sur: 1, Mil: 1, Filter: "stellar", Milestone: true,
		Text: "The %s reach into their star."},
	{Key: "ansible", Name: "the Ansible", Domain: Exotic, Era: 4, Prereqs: []string{"physics", "uploading"}, Weight: 0.15, Soc: 1.5, Structure: "ansible", Milestone: true,
		Text: "The %s find a way to speak across any distance without delay. Their worlds are one world again."},
	{Key: "wormhole_physics", Name: "Wormhole Physics", Domain: Exotic, Era: 4, Prereqs: []string{"antimatter", "physics", "computers"}, Weight: 0.4, Milestone: true,
		Text: "The %s prove that space can be folded. It is only a proof, for now."},
	{Key: "ftl", Name: "Faster-than-light Travel", Domain: Propulsion, Era: 4, Prereqs: []string{"wormhole_physics", "relativistic"}, Weight: 0.5, Reach: 90, Speed: 0.3, Filter: "door", Milestone: true,
		Text: "The %s tear a door in space. Faster-than-light travel is theirs."},
	{Key: "planck_weapons", Name: "Planck Weapons", Domain: Weapons, Era: 4, Prereqs: []string{"wormhole_physics", "relativistic_weapons"}, Weight: 0.5, Mil: 2},
	{Key: "transcendence", Name: "Transcendence", Domain: Exotic, Era: 4, Prereqs: []string{"uploading", "wormhole_physics", "memetics"}, Weight: 0.5, Filter: "transcend"},
	{Key: "star_lifting", Name: "Star Lifting", Domain: Exotic, Era: 4, Prereqs: []string{"stellar_engineering"}, Weight: 0.5, Sur: 1, Milestone: true,
		Text: "The %s learn to feed and drain their star. They will never need to fear it again."},
	{Key: "mind_shaping", Name: "Mind Shaping", Domain: Society, Era: 4, Prereqs: []string{"memetics", "uploading"}, Soc: 1.5},
}

var byKey = map[string]*Node{}

func init() {
	for _, n := range Nodes {
		if n.Weight == 0 {
			n.Weight = 1
		}
		byKey[n.Key] = n
	}
	for _, n := range Nodes {
		for _, p := range n.Prereqs {
			if byKey[p] == nil {
				panic("tech: unknown prerequisite " + p + " of " + n.Key)
			}
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

// Structure is something built within reach that gives levels.
type Structure struct {
	Key, Name     string
	Mil, Sur, Soc float64
	Rate          float64 // research multiplier while held
	Text          string  // %s civ, %s star
}

// Structures by key.
var Structures = map[string]*Structure{
	"arcology": {Key: "arcology", Name: "arcology", Sur: 1, Text: "The %s seal a city at %s against everything outside it."},
	"shipyard": {Key: "shipyard", Name: "shipyard", Mil: 0.5, Text: "Yards turn above %s, building ships for the %s."},
	"defences": {Key: "defences", Name: "defence grid", Mil: 1.5, Text: "The %s ring %s with guns that watch the sky."},
	"ansible":  {Key: "ansible", Name: "ansible net", Soc: 1.5, Text: "The %s link %s to home without delay."},
	"dyson":    {Key: "dyson", Name: "Dyson swarm", Sur: 0.5, Rate: 1.5, Text: "The %s enclose %s in a swarm of collectors. The star dims from outside."},
}
