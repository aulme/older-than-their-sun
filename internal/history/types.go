// Package history simulates the rise and fall of life across the star field.
//
// Two passes: a coarse deep-time pass (life arising, sterilising events,
// precursors, elder things settling in) and a fine recent-history pass
// (civilisations, contact, war, horrors, doom). The present is year 0 and
// is defined as the aftermath: every civilisation ends as extinct,
// transformed, or contracted.
package history

import (
	"math/rand/v2"

	"worldgen/internal/galaxy"
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

// Temper is a civilisation's broad disposition.
type Temper uint8

const (
	Curious Temper = iota
	Insular
	Aggressive
	Zealous
)

func (t Temper) String() string {
	return [...]string{"curious", "insular", "aggressive", "zealous"}[t]
}

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

// Voyage is a sublight colony ship in flight.
type Voyage struct {
	Target int
	Arrive Year
}

// Civ is a civilisation.
type Civ struct {
	ID       int
	Name     string
	Home     int
	HomeName string
	Born     Year
	Ended    Year
	Stage    Stage
	Fate     Fate
	Cause    string // why it ended
	Into     string // what it became, if transformed
	Temper   Temper
	Tech     float64 // 1 = interstellar, 2 = megastructures, 3 = FTL possible
	Systems  []int
	Peak     int
	HasFTL   bool
	Dyson    int
	Enclosed []int // stars with Dyson swarms
	Wars     map[int]bool
	Met      map[int]bool
	Plagued  bool
	Title    string // ruler title once contracted
	Voyages  []Voyage
	colonies int // colonies founded, for log throttling
}

// Living is true for active and remnant civilisations.
func (c *Civ) Living() bool { return c.Stage != Dead }

// Active is true for civilisations still growing.
func (c *Civ) Active() bool { return c.Stage < Remnant }

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
	Victims int
	FromCiv int // -1 if not made by a civilisation
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
	DeepStart Year
	DeepEnd   Year
	DeepStep  Year
	FineStep  Year
}

// DefaultConfig is a small, fast world.
func DefaultConfig() Config {
	return Config{
		Stars: 400, Radius: 150, Thickness: 40,
		DeepStart: -3_000_000_000, DeepEnd: -5_000_000, DeepStep: 10_000_000,
		FineStep: 1_000,
	}
}

// World is the whole simulated history.
type World struct {
	Cfg     Config
	Seed    uint64
	G       *galaxy.Galaxy
	R       *rand.Rand
	Now     Year
	Bio     []BioState
	Owner   []int // civ id owning each star, -1 if none
	Held    []int // horror id holding each star, -1 if none
	Hazard  float64
	Civs    []*Civ
	Horrors []*Horror
	Traces  []Trace
	Events  []Event
	Dusk    int // civs forced to end by the Long Dusk
}

func (w *World) log(format string, args ...any) {
	w.Events = append(w.Events, Event{Year: w.Now, Text: sprintf(format, args...)})
}

func (w *World) trace(star int, kind string, civ int) {
	w.Traces = append(w.Traces, Trace{Star: star, Kind: kind, Civ: civ, Year: w.Now})
}

func (w *World) star(id int) string { return w.G.Stars[id].Name }

func (w *World) chance(p float64) bool { return w.R.Float64() < p }
