package tech

import "worldgen/internal/flow"

// Upkeep and category: what a known node costs each tick to keep working,
// and which of the direction's categories it is fed under. Both are tables
// by domain and era with named exceptions, so a designer tunes them here
// and nowhere else.

// group is the upkeep column a domain falls in.
type group int

const (
	bioGroup   group = iota // biology: organic matter
	mindGroup               // energy, computation, exotic, society: energy
	makerGroup              // industry, weapons, propulsion: metal, and energy from era 2
)

var groupNames = map[string]group{"bio": bioGroup, "mind": mindGroup, "maker": makerGroup}

var groupOf = map[string]group{}

// upkeepFile is the upkeep section of data/tech.json.
type upkeepFile struct {
	Groups map[string]string        `json:"groups"`
	Table  []map[string]flow.Income `json:"table"`
	Free   []string                 `json:"free"`
	Custom map[string]flow.Income   `json:"custom"`
	Grown  []string                 `json:"grown"`
	Fields []string                 `json:"fields"`
}

func loadUpkeep(u upkeepFile) {
	for d, g := range u.Groups {
		gr, ok := groupNames[g]
		if !ok {
			panic("tech: unknown upkeep group " + g)
		}
		groupOf[d] = gr
	}
	for _, d := range Domains {
		if _, ok := groupOf[d]; !ok {
			panic("tech: no upkeep group for " + d)
		}
	}
	if len(u.Table) != 5 {
		panic("tech: the upkeep table has not five eras")
	}
	for e, row := range u.Table {
		for g, gr := range groupNames {
			upkeepTable[e][gr] = row[g]
		}
	}
	for _, k := range u.Free {
		Free[k] = true
	}
	Custom = u.Custom
	for _, k := range u.Grown {
		Grown[k] = true
	}
	for _, k := range u.Fields {
		fields[k] = true
	}
}

// KindOf is the commodity a domain lives on: what a miracle of the domain
// costs, and what a source of the domain yields.
func KindOf(domain string) flow.Kind {
	switch groupOf[domain] {
	case bioGroup:
		return flow.O
	case makerGroup:
		return flow.M
	}
	return flow.E
}

// upkeepTable is the cost by era and group, from the file.
var upkeepTable [5][3]flow.Income

// MiracleUpkeep is what a miracle costs, of its domain's kind.
const MiracleUpkeep = 3

// Free are the nodes that cost nothing: the producers, which are what the
// sources need, and the first engines.
var Free = map[string]bool{}

// Custom are the nodes whose upkeep is their own and not the table's: the
// vacuum tap costs a little metal for the energy it gives at every held star.
var Custom map[string]flow.Income

// Grown are the nodes whose metal is flesh: a living ship or a seed-cloud
// is bred, not built, so what the table asks in metal is asked in organic
// matter.
var Grown = map[string]bool{}

// Upkeep is what the node costs each tick to keep working. Era 0 and the
// kind nodes cost nothing; the world nodes cost one of their domain's kind,
// which is what the table gives them.
func (n *Node) Upkeep() flow.Income {
	if Free[n.Key] || n.Era == 0 {
		return flow.Income{}
	}
	if u, ok := Custom[n.Key]; ok {
		return u
	}
	if n.Miracle {
		var u flow.Income
		u[KindOf(n.Domain)] = MiracleUpkeep
		return u
	}
	u := upkeepTable[n.Era][groupOf[n.Domain]]
	if Grown[n.Key] {
		u[flow.O], u[flow.M] = u[flow.M], 0
	}
	return u
}

// fields are the biology nodes that keep people alive; the rest of biology
// is the making of bodies and worlds, and is fed with the works.
var fields = map[string]bool{}

// Cat is the category the node is fed under.
func (n *Node) Cat() flow.Category {
	if fields[n.Key] {
		return flow.Fields
	}
	switch n.Domain {
	case Weapons:
		return flow.Arms
	case Industry, Energy, Biology:
		return flow.Works
	case Propulsion:
		return flow.Road
	}
	return flow.Mind
}
