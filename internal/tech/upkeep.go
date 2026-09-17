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

var groupOf = map[string]group{
	Biology: bioGroup,
	Energy:  mindGroup, Computation: mindGroup, Exotic: mindGroup, Society: mindGroup,
	Industry: makerGroup, Weapons: makerGroup, Propulsion: makerGroup,
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

// upkeepTable is the cost by era and group, as (O, E, M).
var upkeepTable = [5][3]flow.Income{
	{},
	{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
	{{1, 0, 0}, {0, 1, 0}, {0, 1, 1}},
	{{1, 1, 0}, {0, 2, 0}, {0, 1, 2}},
	{{2, 1, 0}, {0, 3, 0}, {0, 2, 3}},
}

// MiracleUpkeep is what a miracle costs, of its domain's kind.
const MiracleUpkeep = 3

// Free are the nodes that cost nothing: the producers, which are what the
// sources need, and the first engines.
var Free = map[string]bool{
	"agriculture": true, "metallurgy": true, "steam": true, "electricity": true,
	"atomic": true, "fusion": true, "interplanetary": true, "orbital_habitats": true,
	"terraforming": true, "synthetic_biology": true, "dyson": true, "vacuum_energy": true,
	"stellar_engineering": true, "star_lifting": true,
}

// Grown are the nodes whose metal is flesh: a living ship or a seed-cloud
// is bred, not built, so what the table asks in metal is asked in organic
// matter.
var Grown = map[string]bool{"living_ships": true, "seed_clouds": true}

// Upkeep is what the node costs each tick to keep working. Era 0 and the
// kind nodes cost nothing; the world nodes cost one of their domain's kind,
// which is what the table gives them.
func (n *Node) Upkeep() flow.Income {
	if Free[n.Key] || n.Era == 0 {
		return flow.Income{}
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
var fields = map[string]bool{
	"agriculture": true, "medicine": true, "ecology": true, "closed_ecologies": true, "life_extension": true,
}

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
