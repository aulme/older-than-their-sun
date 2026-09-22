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

	"worldgen/data"
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

// Node is one discovery. The tree is data/tech.json; the fields are its
// columns.
type Node struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Domain     string   `json:"domain"`
	Era        int      `json:"era"`
	Prereqs    []string `json:"prereqs,omitempty"`
	Weight     float64  `json:"weight,omitempty"`   // discovery weight, 0 means 1
	Cost       float64  `json:"cost,omitempty"`     // research points to reach it; 0 means the era's default
	Patience   float64  `json:"patience,omitempty"` // kyr a civilisation must have lived before it can pursue this
	Chance     float64  `json:"chance,omitempty"`   // if set, a finished pursuit only succeeds this often; failure wastes the work
	Miracle    bool     `json:"miracle,omitempty"`  // a power apart from the tree: rare, potent, dangerous
	For        string   `json:"for,omitempty"`      // only for these: "sub:machine", "mod:planetary", "world:iceshell", "trait:fireless"; "" for everyone
	Mil        float64  `json:"mil,omitempty"`
	Sur        float64  `json:"sur,omitempty"`
	Soc        float64  `json:"soc,omitempty"`
	Wis        float64  `json:"wis,omitempty"`        // what the node adds to Wisdom; ten nodes set it
	Reach      float64  `json:"reach,omitempty"`      // reach in light years this node grants; the highest known wins
	Speed      float64  `json:"speed,omitempty"`      // colony ship speed in years per light year; the lowest known wins
	Env        int      `json:"env,omitempty"`        // widens the habitable envelope
	Focus      M        `json:"focus,omitempty"`      // research tilt after discovery
	Filter     string   `json:"filter,omitempty"`     // filter key procced on discovery
	Structures []string `json:"structures,omitempty"` // structure keys unlocked; Structure() is the first
	Gated      string   `json:"gated,omitempty"`      // a rarity without which the node cannot be learned at all; none is, and TestNoCatch22 keeps it so
	Milestone  bool     `json:"milestone,omitempty"`  // worth a line in the legends
	Idx        int      `json:"-"`                    // this node's place in Nodes, for callers that index by node
	Ladder     string   `json:"ladder,omitempty"`     // a rung of a plague ladder: Bio or Mind; see plague.go
	Immune     bool     `json:"immune,omitempty"`     // the top of its ladder: no plague of the kind is born in or caught by a people working it
	Clean      bool     `json:"clean,omitempty"`      // halves biological births only: sewers, sealed cities
	Cure       float64  `json:"cure,omitempty"`       // what the rung adds to the cure roll; 0 means the usual half
	Weapon     *Weapon  `json:"weapon,omitempty"`     // a plague-making craft; nil for none
	Text       string   `json:"text,omitempty"`       // the line the chronicle says of a milestone; {S} is the people
	Desc       string   `json:"desc,omitempty"`       // what it is, in one line
}

// EraNames label the derived era of a civilisation, from data/tech.json.
var EraNames []string

// Eras are the five ages of a people's arts: a key, a name and the
// default research cost of a node of the era.
var Eras []Era

// Era is one row of the eras table.
type Era struct {
	Key  string  `json:"key"`
	Name string  `json:"name"`
	Cost float64 `json:"cost"`
}

// Nodes is the tree, in the file's order.
var Nodes []*Node

// EraCosts is the default research cost of a node by era.
var EraCosts []float64

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
var Subs map[string][]string

// Miracles lists the miracle nodes.
var Miracles []*Node

// Weapon is what a plague-making node allows: the kind it makes, the top
// of the band its contagion and lethality may be chosen in, whether the
// plague is tailored to one blood, whether the maker is immune to what it
// makes, and whether it may make one that thinks. Rung is its place on
// the craft, from one, which the containment filter adds to its difficulty
// and the vial's scar locks above.
type Weapon struct {
	Memetic   bool    `json:"memetic,omitempty"`
	Band      float64 `json:"band"`
	Tailored  bool    `json:"tailored,omitempty"`
	Immune    bool    `json:"immune,omitempty"`
	Conscious bool    `json:"conscious,omitempty"`
	Rung      int     `json:"rung"`
}

// Crafts lists the plague-making nodes, filled at init.
var Crafts []*Node

// The plague ladders: the nodes that divide births, harden against
// catching and add to the cure, by kind; the Immune node at the top of
// each is the end of the kind for whoever works it. Ladders is filled from
// the nodes' markers at init; see plague.go in history for the reading.
const (
	Bio  = "bio"
	Mind = "mind"
)

var Ladders = map[string][]*Node{}

var byKey = map[string]*Node{}

// file is the shape of data/tech.json.
type file struct {
	Domains []string            `json:"domains"`
	Eras    []Era               `json:"eras"`
	Nodes   []*Node             `json:"nodes"`
	Subs    map[string][]string `json:"subs"`
	Upkeep  upkeepFile          `json:"upkeep"`
}

// worksFile is the shape of data/works.json.
type worksFile struct {
	Structures []*Structure `json:"structures"`
}

func init() {
	var f file
	data.Load("tech.json", &f)
	if len(f.Domains) != len(Domains) {
		panic("tech: the file's domains are not the code's")
	}
	for i, d := range f.Domains {
		if d != Domains[i] {
			panic("tech: the file's domain " + d + " is not the code's " + Domains[i])
		}
	}
	Eras, Nodes, Subs = f.Eras, f.Nodes, f.Subs
	for _, e := range Eras {
		EraNames = append(EraNames, e.Name)
		EraCosts = append(EraCosts, e.Cost)
	}
	var wf worksFile
	data.Load("works.json", &wf)
	for _, st := range wf.Structures {
		Structures[st.Key] = st
		StructureKeys = append(StructureKeys, st.Key)
	}
	loadUpkeep(f.Upkeep)
	for i, n := range Nodes {
		n.Idx = i
		if n.Desc == "" {
			panic("tech: no description for " + n.Key)
		}
		if n.Weight == 0 {
			n.Weight = 1
		}
		byKey[n.Key] = n
		if n.Miracle {
			Miracles = append(Miracles, n)
		}
		if n.Ladder != "" {
			Ladders[n.Ladder] = append(Ladders[n.Ladder], n)
		}
		if n.Weapon != nil {
			Crafts = append(Crafts, n)
		}
	}
	for _, n := range Nodes {
		for _, p := range n.Prereqs {
			if byKey[p] == nil {
				panic("tech: unknown prerequisite " + p + " of " + n.Key)
			}
		}
		for _, k := range n.Structures {
			if Structures[k] == nil {
				panic("tech: unknown structure " + k + " of " + n.Key)
			}
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

// Structure is the first structure a node unlocks, or "".
func (n *Node) Structure() string {
	if len(n.Structures) == 0 {
		return ""
	}
	return n.Structures[0]
}

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
	Key    string        `json:"key"`
	Name   string        `json:"name"`
	Node   string        `json:"node"` // the node that unlocks it
	Mil    float64       `json:"mil,omitempty"`
	Sur    float64       `json:"sur,omitempty"`
	Soc    float64       `json:"soc,omitempty"`
	Upkeep flow.Income   `json:"upkeep"`           // per tick, while it works
	Cat    flow.Category `json:"cat"`              // what it is fed under
	Per    string        `json:"per,omitempty"`    // what it is one of: "star" (one per star), "belt" (one per belt), "" (two per people, one per star)
	Hardy  float64       `json:"hardy"`            // multiplier on decay once abandoned; less is hardier
	Guns   int           `json:"guns,omitempty"`   // guns it stands in the sky of its star: an immobile fleet at the builder's quality; one per star, no cap per people
	Modern bool          `json:"modern,omitempty"` // one gun more per weapons era the builder knows beyond the node's
	Repair bool          `json:"repair,omitempty"` // its guns are remade a thousand years each while the world is held, a siege or no; else only between sieges
	Dug    bool          `json:"dug,omitempty"`    // raised by dig, at every world a people holds, never by the pick
	Watch  float64       `json:"watch,omitempty"`  // how far it sees a fleet in flight, in light years; an eye where it stands
	Max    int           `json:"max,omitempty"`    // the most a people may raise; 0 is the usual two
	Text   string        `json:"text"`             // its raising: {S} the people, {T} the star
	Remain string        `json:"remain"`           // what is left of it once abandoned; {M} the makers
}

// Structures by key.
var Structures = map[string]*Structure{}

// StructureKeys is every structure in the file's order, for loops that
// must not range over the map.
var StructureKeys []string

// Yields says whether a structure harnesses a source rather than giving
// levels: the mines, the collectors, the swarm, the tap and the lifter.
func (s *Structure) Yields() bool { return s.Per != "" }
