package species

import (
	"strings"

	"worldgen/internal/mind"
)

// A people is made of exactly one substrate and carries zero or more
// modifiers. Both are entries in the registry below, each carrying its own
// odds, vocabulary, profile and hooks, so a new one is a new file and not a
// new switch arm anywhere else. Nothing outside this package names an
// entry by switch: the sim reads Species.Profile(), Species.Is and
// Species.Sub.

// Substrate is what a people is made of.
type Substrate uint8

const (
	Biological Substrate = iota
	Machine
	Eldritch
	Parasite
)

func (s Substrate) String() string { return Substrates[s].Key }

// Def is the substrate's registry entry.
func (s Substrate) Def() *SubstrateDef { return Substrates[s] }

// SubstrateByKey finds a substrate by key; ok is false for an unknown one.
func SubstrateByKey(key string) (Substrate, bool) {
	for _, d := range Substrates {
		if d.Key == key {
			return d.Sub, true
		}
	}
	return 0, false
}

// Mod is a set of modifiers: how a people is shaped.
type Mod uint8

const (
	Planetary Mod = 1 << iota
	Hive
	Unconscious
	Replicator
	Antimemetic
	Evolver
)

// Has says whether every modifier in o is in m.
func (m Mod) Has(o Mod) bool { return m&o == o }

// Defs lists the entries of the carried modifiers in registry order.
func (m Mod) Defs() []*ModDef {
	var out []*ModDef
	for _, d := range Mods {
		if m.Has(d.Mod) {
			out = append(out, d)
		}
	}
	return out
}

// String lists the carried modifiers by key: "planetary, hive"; "" for none.
func (m Mod) String() string {
	var keys []string
	for _, d := range m.Defs() {
		keys = append(keys, d.Key)
	}
	return strings.Join(keys, ", ")
}

// ModByKey finds a modifier by key; ok is false for an unknown one.
func ModByKey(key string) (Mod, bool) {
	for _, d := range Mods {
		if d.Key == key {
			return d.Mod, true
		}
	}
	return 0, false
}

// Flavour is what a people calls the things it builds. An entry leaves a
// field empty to take the substrate's word.
type Flavour struct {
	Colony  string
	Ship    string
	Station string
}

// Setting picks which numbers the generator draws with. Legacy reproduces
// the old kind table so that histories do not move; Proposed is the kinds
// proposal's first setting, switched on when the modifier rules land.
type Setting uint8

const (
	Legacy Setting = iota
	Proposed
)

// Draw is an entry's place in the chain of rolls under one setting.
type Draw struct {
	Base   float64            // a weight for a substrate; a chance for a modifier or the swarm roll
	Tilts  map[string]float64 // multipliers on the odds of later rolls, by their key (a modifier, "swarming", or a trait); missing is 1
	Skip   []string           // trait groups not rolled for a people that carries the entry
	Groups map[string]float64 // multipliers on the chance of rolling a trait group; missing is 1
}

// Entry is what a substrate and a modifier have in common.
type Entry struct {
	Key      string
	Portrait string   // the sentence the portrait opens with; "" for the plain case
	Arising  string   // how the legends say it arose; "" for the plain case
	Flavour  Flavour  // vocabulary; empty fields fall back to the substrate's
	Legacy   Draw     // the numbers that reproduce the old kind table
	Draws    Draw     // the proposal's first setting
	Own      []string // trait groups only a people with this entry rolls
	Profile  Profile  // what it does to the sim
	Hooks    Hooks
}

func (e *Entry) draw(s Setting) *Draw {
	if s == Proposed {
		return &e.Draws
	}
	return &e.Legacy
}

// SubstrateDef is a substrate's registry entry.
type SubstrateDef struct {
	Entry
	Sub Substrate
}

// ModDef is a modifier's registry entry.
type ModDef struct {
	Entry
	Mod        Mod
	Cradle     string  // an archetype the modifier is usually born on, "" for none
	CradleOdds float64 // how often, when the home world is rolled
}

// Hooks are the few behaviours that are not a dial. They grow as the steps
// that need them land; every hook is typed on species values only.
type Hooks struct{}

// Ability is something a people can do unless an entry it carries denies it.
type Ability uint16

const (
	Fields        Ability = 1 << iota // grows food
	Works                             // builds structures
	Trades                            // sends and takes trade
	Launches                          // sends expeditions
	SettlesByShip                     // colonises with ships
	Stiffens                          // ossifies
	CivilWars                         // splits under pressure: schism, civil war
	HoldsGrudges                      // takes wrongs to heart
	Flees                             // takes to the sky when its last world is lost
	Sickens                           // catches biological plagues
)

// Profile is what an entry does to the sim. The sim reads the composed
// profile of a species: multipliers multiply, adds add, denials union. An
// entry leaves a multiplier at zero to mean one and a map nil to mean
// nothing.
type Profile struct {
	Mil, Sur, Soc float64            // adds to the base levels
	Reach         float64            // multiplier on reach
	Rate          float64            // multiplier on research
	Expand        float64            // multiplier on the colony rate
	Memory        float64            // multiplier on the wear of tales; below one keeps them
	Endure        float64            // multiplier on how long a failing sun can be borne
	Env           int                // widens the habitable envelope
	Dom           M                  // research tilt by domain
	Dials         mind.Dials         // what it does to temperament
	FilterDiff    map[string]float64 // what it does to each filter's difficulty
	Cannot        Ability            // what it cannot do
	// means: see the flow package and history's flow.go
	Cradle          float64 // multiplier on what the cradle world yields the people that arose on it
	Upkeep          M       // multiplier on the upkeep of each domain's nodes
	OrganicAsEnergy bool    // every cost in organic matter is paid in energy instead
}

// Can says whether the profile allows an ability.
func (p Profile) Can(a Ability) bool { return p.Cannot&a == 0 }

func mul(m float64) float64 {
	if m == 0 {
		return 1
	}
	return m
}

// Compose multiplies the entries' profiles together.
func Compose(ps ...Profile) Profile {
	out := Profile{Reach: 1, Rate: 1, Expand: 1, Memory: 1, Endure: 1, Cradle: 1, Dom: M{}, Upkeep: M{}, FilterDiff: map[string]float64{}}
	for _, p := range ps {
		out.Mil, out.Sur, out.Soc = out.Mil+p.Mil, out.Sur+p.Sur, out.Soc+p.Soc
		out.Reach *= mul(p.Reach)
		out.Cradle *= mul(p.Cradle)
		out.OrganicAsEnergy = out.OrganicAsEnergy || p.OrganicAsEnergy
		for k, v := range p.Upkeep {
			out.Upkeep[k] = mul(out.Upkeep[k]) * v
		}
		out.Rate *= mul(p.Rate)
		out.Expand *= mul(p.Expand)
		out.Memory *= mul(p.Memory)
		out.Endure *= mul(p.Endure)
		out.Env += p.Env
		for k, v := range p.Dom {
			out.Dom[k] = mul(out.Dom[k]) * v
		}
		out.Dials.Add(p.Dials)
		for k, v := range p.FilterDiff {
			out.FilterDiff[k] += v
		}
		out.Cannot |= p.Cannot
	}
	return out
}

// Substrates is the registry of substrates, indexed by Substrate.
var Substrates = []*SubstrateDef{biological, machine, eldritch, parasite}

// Mods is the registry of modifiers in the order the chain rolls them.
var Mods = []*ModDef{planetary, hive, unconscious, replicator, antimemetic, evolver}

func init() {
	for i, d := range Substrates {
		if d.Sub != Substrate(i) {
			panic("species: substrate registry out of order at " + d.Key)
		}
	}
}
