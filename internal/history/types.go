// Package history simulates the rise and fall of life across the star field.
//
// Three passes. The age generator writes the myth of earlier ages as coarse
// events and leaves their legacies on the substrate. The middle pass runs the
// early current age at medium grain and the fine pass the late current age at
// fine grain, both with the same engine. The present is year 0 and is the
// aftermath: every civilisation ends extinct, transformed or contracted.
package history

import (
	"math/rand/v2"

	"worldgen/internal/galaxy"
	"worldgen/internal/species"
)

// Year is years relative to the present (negative = past).
type Year int64

// Event is one line of the legends log.
type Event struct {
	Year Year
	Text string
}

// BioState tracks life on a star's worlds.
type BioState uint8

const (
	BioNone BioState = iota
	BioSimple
	BioComplex
)

// Stage is where a civilisation is in its lifecycle.
type Stage uint8

const (
	Emergent Stage = iota
	Interstellar
	Zenith
	Remnant // contracted, still alive
	Dead    // extinct or transformed
)

// Fate is the end state. The present being the aftermath, every civ gets one.
type Fate uint8

const (
	FateNone Fate = iota
	Extinct
	Transformed
	Contracted
)

func (f Fate) String() string {
	return [...]string{"active", "extinct", "transformed", "contracted"}[f]
}

// Voyage is a colony ship in flight.
type Voyage struct {
	Target int
	Arrive Year
}

// Civ is a civilisation: a species on a home world with a history.
type Civ struct {
	ID         int
	Name       string
	Species    *species.Species
	Home       int // current seat; moves if the cradle is lost
	HomeName   string
	Cradle     int // the world the species arose on; never changes, and neither does the species
	CradleName string
	Born       Year
	Ended      Year
	Fell       Year // when it stopped being active
	Stage      Stage
	Fate       Fate
	Cause      string // why it ended
	Into       string // what it became, if transformed
	Systems    []int
	Peak       int
	Voyages    []Voyage
	colonies   int

	// research
	Known  map[string]bool
	Era    int
	Focus  map[string]float64 // temporary research tilt, decays to 1
	Locked map[string]bool    // domains closed by world or scar

	// derived each tick
	Mil, Sur, Soc float64
	Reach         float64
	Speed         float64 // years per light year
	Envelope      int
	Morale        float64 // dynamic part of Social

	// structures and works
	Structures map[string]int // structure key -> count
	Enclosed   []int          // stars with Dyson swarms
	Wielded    []*Legacy
	Found      map[int]bool // legacies attempted
	Heard      map[int]bool // beacons already faced
	Uplifts    int

	// relations
	Wars     map[int]bool
	Met      map[int]bool
	Trade    map[int]bool
	Master   int // civ that holds this one, -1 if free
	Vassal   bool
	Declines int // declines suffered, watched by slaves for revolt
	Seen     int // master declines this civ has reacted to

	// filters
	Faced        map[string]bool
	Scars        map[string]bool
	Boons        map[string]bool
	Record       []string
	DarkAges     int
	KnowsCycle   bool // learned the shape of the cycle
	Renewed      Year
	Renaissances int
	NextDrift    int // size at which the Distance is faced again
	Plagued      bool
	Dying        bool    // home star is failing
	Endure       float64 // kyr left under the failing star
	Title        string  // ruler title once contracted
}

// Living is true for active and remnant civilisations.
func (c *Civ) Living() bool { return c.Stage != Dead }

// Active is true for civilisations still growing.
func (c *Civ) Active() bool { return c.Stage < Remnant }

// Free is true for civilisations not held by another.
func (c *Civ) Free() bool { return c.Master < 0 }

// Has reports a species trait.
func (c *Civ) Has(trait string) bool { return c.Species.Has(trait) }

// HorrorKind is which alien horror category an actor belongs to.
type HorrorKind uint8

const (
	Replicators HorrorKind = iota // self-copying machines
	Beacon                        // memetic hazard broadcast
	Elder                         // dormant elder entity
	RogueMind                     // machine intelligence that outgrew its makers
)

func (k HorrorKind) String() string {
	return [...]string{"replicator swarm", "memetic beacon", "elder entity", "rogue intelligence"}[k]
}

// Horror is a non-civilisation actor.
type Horror struct {
	ID      int
	Kind    HorrorKind
	Name    string
	Origin  int
	Born    Year
	Systems []int
	Dormant bool
	Wakings int
	Sleep   Year // will not wake before this
	Victims int
	FromCiv int // -1 if not made by a civilisation
	Legacy  int // legacy record it belongs to, -1 if none
}

// LegacyKind is what an age leaves behind.
type LegacyKind uint8

const (
	Artifact LegacyKind = iota
	Structure
	Threat
	Sleeper
	Law
)

func (k LegacyKind) String() string {
	return [...]string{"artifact", "structure", "threat", "sleeper", "law"}[k]
}

// LegacyState is what has happened to a legacy.
type LegacyState uint8

const (
	Buried LegacyState = iota
	Sealed
	Wielded
	Mastered
	Unleashed
	Lost
)

func (s LegacyState) String() string {
	return [...]string{"undisturbed", "sealed", "wielded", "mastered", "unleashed", "lost"}[s]
}

// Legacy is something an earlier age left on the substrate. The current age
// writes the same record type for what it leaves.
type Legacy struct {
	ID     int
	Age    int // index into World.Ages, or -1 for the current age
	Elder  *ElderCiv
	Kind   LegacyKind
	Star   int
	Node   string // tech node, for artifacts and structures
	Desc   string // "a ring of black metal around a dead star"
	Name   string // given by the finder
	State  LegacyState
	Horror int    // horror id for threats and sleepers, -1 if none
	Finder int    // civ that last acted on it, -1 if none
	Level  string // for wielded artifacts: which level it lifts
}

// Elder is a civilisation of an earlier age. No traits, only a portrait.
type ElderCiv struct {
	Age      int
	Portrait string
	Name     string // given by finders, "" until found
	Rose     Year
	Fell     Year
	Legacies []*Legacy
}

// AgeRecord is one earlier age of the galaxy.
type AgeRecord struct {
	Index  int
	Start  Year   // the surge
	End    Year   // fertility below the floor
	Ender  string // what swept up the remains
	Elders []*ElderCiv
}

// Trace is something left behind for the player to find.
type Trace struct {
	Star int
	Kind string
	Civ  int // -1 if none
	Year Year
}

// Config tunes the simulation.
type Config struct {
	Stars     int
	Radius    float64
	Thickness float64
	DeepStart Year // substrate begins
	MidStart  Year // current age begins, medium grain
	FineStart Year // late current age, fine grain
	DeepStep  Year
	MidStep   Year
	FineStep  Year
}

// DefaultConfig is a small, fast world.
func DefaultConfig() Config {
	return Config{
		Stars: 400, Radius: 150, Thickness: 40,
		DeepStart: -galaxy.Age, MidStart: -60_000_000, FineStart: -5_000_000,
		DeepStep: 10_000_000, MidStep: 20_000, FineStep: 1_000,
	}
}

// World is the whole simulated history.
type World struct {
	Cfg      Config
	Seed     uint64
	G        *galaxy.Galaxy
	R        *rand.Rand
	Now      Year
	dt       float64 // current tick in kyr
	Bio      []BioState
	Owner    []int // civ id owning each star, -1 if none
	Held     []int // horror id holding each star, -1 if none
	Hazard   float64
	Civs     []*Civ
	Horrors  []*Horror
	Ages     []*AgeRecord
	Cycle    *Cycle
	Legacies []*Legacy
	Traces   []Trace
	Events   []Event
	Dusk     int // civs forced to end by the Long Dusk
	scratch
}

// scratch passed from a trigger to its filter's outcome functions
type scratch struct {
	blastWorlds []int
	blastWhat   string
	beacon      *Horror
	incursionAt int
	incursionBy *Horror
}
