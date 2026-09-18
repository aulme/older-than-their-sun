// Package history simulates the rise and fall of life across the star field.
//
// Two passes. The age generator writes the myth of earlier ages as coarse
// events and leaves their legacies on the substrate. The current age then
// runs from its dawn at one tick of a thousand years, as a pipeline of named
// phases. The present is year 0 and is the aftermath: every civilisation ends
// extinct, transformed or contracted.
package history

import (
	"math/rand/v2"
	"time"
	"worldgen/internal/mind"

	"worldgen/internal/flow"
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
	Blind  bool // sent to a star nobody has read, on a guess
}

// Civ is a civilisation: a species on a home world with a history.
type Civ struct {
	ID         int
	Name       string           // the people's own name; a cradle people is named for its species
	Species    *species.Species // its blood, one of w.Species; kin share it, and it changes only by a made path
	Origin     string           // how the people came to be when not by arising: "a branch of the X"; "" for a cradle people
	Home       int              // current seat; moves if the cradle is lost
	HomeName   string
	Cradle     int // the world the people arose on; never changes
	CradleName string
	Born       Year
	Ended      Year
	Fell       Year // when it stopped being active
	Stage      Stage
	Fate       Fate
	Cause      string          // why it ended
	Into       string          // what it became, if transformed
	Word       string          // the people's word for the state beneath, once they have reached into it
	Hosts      int             // for parasites: host species ridden, the local one included
	Morality   Morality        // what the people counts as wrong; see morality.go
	Lifted     map[string]bool // world blocks lifted by a colony: sea, sky, fire
	Systems    []int
	Peak       int
	Voyages    []Voyage
	colonies   int

	// research
	Known    map[string]bool
	Learned  map[string]Year // when each node was learned by pursuit or find; inherited nodes are absent
	Era      int
	Focus    map[string]float64 // temporary research tilt, decays to 1
	Locked   map[string]bool    // domains closed by world or scar
	Pursuit  string             // the node being worked toward
	Progress float64            // research points banked toward it
	Miracles map[string]string  // miracle key -> how it was gained: born, leap, found, wielded

	// derived each tick
	Mil, Sur, Soc float64
	Reach         float64
	Speed         float64 // years per light year
	Envelope      int
	Morale        float64 // dynamic part of Social

	// structures and works
	Structures map[string]int // structure key -> count
	Works      []Work         // where the structures stand
	Wielded    []*Legacy
	Found      map[int]bool // legacies attempted
	Heard      map[int]bool // beacons already faced
	Uplifts    int
	Ruled      int // peoples this one has held as slaves or vassals

	// relations
	Wars     map[int]bool // peoples this one is at war with; the wars themselves are on the World
	Met      map[int]bool // known of: by touch or by signal
	Reached  map[int]bool // met in the flesh: territories touched
	Trade    map[int]bool
	Master   int // civ that holds this one, -1 if free
	Vassal   bool
	Declines int // declines suffered, watched by slaves for revolt
	Seen     int // master declines this civ has reacted to

	// war and diplomacy: see dials.go, intel.go, war.go, expedition.go, pact.go
	Away       float64         // Military out with fleets and scouts, subtracted from the level
	Dials      Dials           // temperament as numbers
	Intel      map[int]*Intel  // what this people believes about each other people
	Grudge     map[int]float64 // what each other people has done to them
	Truce      map[int]Year    // no new war with each before this
	Fought     map[int]int     // wars fought with each
	Watched    map[int]bool    // looked at hard and left alone, until beliefs change
	Asked      map[int]Year    // when each was last offered a pact
	Scouted    map[int]Year    // when a scout last reported on each
	Ridden     map[int]bool    // for parasites: peoples taken as hosts
	Charted    map[int]Year    // stars read: worlds and who is on them known; see explore.go
	Marked     map[int]bool    // stars the Sight showed something at, for the surveyors to visit
	Searching  bool            // the Sight is turned outward, reading stars, not watching borders
	Starfaring Year            // when reach first touched another star; 0 if never
	Pacts      []int
	Tally      Tally
	Lore       []*Tale      // what this people knows of what happened; see lore.go
	LoreDials  Dials        // what the telling does to the temperament
	lore       map[int]bool // facts held, forgotten or not
	inscribed  map[int]bool // remains whose testament this people has read
	foeNow     int          // the enemy of the day, or -1; a new one gets the old blame
	monsters   map[int]bool // peoples remembered as things that do harm
	LastDark   Year
	Summoned   bool    // an event calls the council this tick
	Aloft      bool    // a nomad people living as fleets, with no worlds
	Rested     bool    // a nomad people that came to rest, and will not rise again
	Quality    float64 // for the aloft: the Military the tree would give, which the fleets grow toward

	// flows: see flow.go and sources.go
	Income       flow.Income     // this tick's yield by kind
	Upkeep       flow.Income     // the needs of every use, fed or not
	Surplus      flow.Income     // income less what the fed uses take
	Want         flow.Income     // what more would feed everything
	Order        flow.Order      // the direction this tick
	Shed         map[string]bool // nodes dormant this tick
	DormantSince map[string]Year // when each dormant node went dark
	ShedSince    Year            // when the current stretch of shedding began; a stretch is running while Shed is not empty
	Wanted       bool            // the lean years were written for this stretch
	ShedTicks    map[string]int  // for the batch: ticks each node has spent dark
	HighIncome   flow.Income     // income, upkeep and want at the people's height of means
	HighUpkeep   flow.Income
	HighWant     flow.Income
	highUpkeep   float64
	Loot         flow.Income // taken once, added to the next tick's income: what a horde strips from a world
	Reserved     flow.Income // what launches and builds took of the spare this tick
	WorkingNeed  flow.Income // the needs of the uses fed this tick

	// trade: see trade.go
	From          map[int]flow.Income // what each partner sent last tick; counted in this tick's income
	Received      flow.Income         // the sum of From as it was counted
	OwnWant       flow.Income         // the want less what partners send: what trade is asked for
	Dependent     map[int]bool        // partners whose sending keeps the fed uses fed
	Refused       map[int]Year        // since when each partner has been refused while it wanted
	Embargo       map[int]bool        // partners this people has closed its ports to
	FellDependent bool                // was dependent on a partner at the moment it fell
	partners      map[int]bool        // partners ever traded with, for the batch
	fed           map[int]bool        // partners ever sent anything, for the batch
	Remade        map[string]Year     // when the object of a miracle was last lost; another comes a million years on

	// rarities: see rarity.go
	Rare      map[string]bool // the rarities had this tick, by key
	Grants    map[string]bool // the nodes those rarities grant at half cost
	Had       map[string]bool // every rarity ever had, by key
	Harnessed map[string]bool // every source kind ever harnessed, by key
	Built     map[string]int  // structures raised, by key, for the batch
	Granted   map[string]bool // nodes learned while their grant was had

	// filters
	Faced        map[string]bool
	Scars        map[string]bool
	Boons        map[string]bool
	Record       []string
	DarkAges     int
	KnowsCycle   bool // learned the shape of the cycle
	Ascended     Year // when a miracle was last gained; the surge runs from here
	Renewed      Year
	Renaissances int
	NextDrift    int // size at which the Distance is faced again
	Plagued      bool
	Dying        bool    // home star is failing
	Endure       float64 // kyr left under the failing star
	Title        string  // ruler title once contracted
}

// Tally counts what a people did in war and peace, for the batch reports.
type Tally struct {
	Declared, Fought, Taken, Lost, Glassed int
	Fleets, Native, Scouts, Relief         int
	Pacts, Refused, Betrayals, Called      int
	Capitulated                            bool
	// exploration
	Surveys, Charted, Blind, BlindLost          int
	FindSurvey, FindSettle, FindChance, FindOwn int
	MetTouch, MetHeard, MetSurvey, MetShip      int
	Searched, Sighted                           float64 // kyr with the Sight turned outward; kyr holding it
	// tellings
	Tales, Witnessed, Told, Read, Inherited              int
	Forgot, Myths, Revised, Blamed, Testaments, Restored int
	// flows: ticks lived, ticks shedding, both by era, and both alone (one system)
	Ticks, Lean          int
	TicksAt, LeanAt      [5]int
	AloneAt, LeanAloneAt [5]int
	// trade: what was sent and what came, over the life; partners ever, and partners ever sent to
	Sent, Got     flow.Income
	Partners, Fed int
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
	Replicators   HorrorKind = iota // self-copying machines
	Beacon                          // memetic hazard broadcast
	SleeperHorror                   // an elder that withdrew and went still; the legacy kind is Sleeper
	RogueMind                       // machine intelligence that outgrew its makers
)

func (k HorrorKind) String() string {
	return [...]string{"replicator swarm", "memetic beacon", "sleeper", "rogue intelligence"}[k]
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
	Bounty // a thing still doing what it was made to do: a yield for whoever puts it to use
)

func (k LegacyKind) String() string {
	return [...]string{"artifact", "structure", "threat", "sleeper", "law", "bounty"}[k]
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

// Condition is the state of repair of a legacy, best to worst. Below Ruin is Lost.
type Condition uint8

const (
	Abandoned Condition = iota // whole; everything works more or less
	Derelict                   // bad shape, but usable
	Wreck                      // repairable with a lot of work
	Ruin                       // nothing usable; something may still be learned
)

func (c Condition) String() string {
	return [...]string{"abandoned", "derelict", "wreck", "ruin"}[c]
}

// Wreckage is what an ending does to the works of the fallen: the fraction
// destroyed outright, and the condition the rest are left in.
type Wreckage struct {
	Destroy float64
	Leave   Condition
}

// Legacy is something an earlier age left on the substrate. The current age
// writes the same record type for what it leaves.
type Legacy struct {
	ID        int
	Age       int // index into World.Ages, or -1 for the current age
	Elder     *Elder
	Maker     int // civ that made it, -1 for the elder ages
	Kind      LegacyKind
	Star      int
	Node      string // tech node, for artifacts and structures
	Desc      string // "a ring of black metal around a dead star"
	Name      string // given by the finder
	State     LegacyState
	Horror    int    // horror id for threats and sleepers, -1 if none
	Finder    int    // civ that last acted on it, -1 if none
	Level     string // for wielded artifacts: which level it lifts, or "miracle"
	Cond      Condition
	Hardy     float64       // multiplier on the rate of decay; 0 never decays
	Source    int           // the source record it is, for a bounty or a wielded artifact; -1 if none
	Testament []Inscription // what its makers told of their age, as they left it
}

// Elder is a civilisation of an earlier age. No traits, only a portrait.
type Elder struct {
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
	Elders []*Elder
}

// Work is one structure standing at a star. Legacy is set if it was inherited.
type Work struct {
	Key    string
	Node   string
	Star   int
	Legacy int
	Dark   bool // shed this tick: no levels, no yield
}

// key is the work as a use: what the shed set names it by.
func (wk Work) key() string { return "work:" + wk.Key + ":" + itoa(wk.Star) }

// Trace is something left behind for the player to find.
type Trace struct {
	Star int
	Kind string
	Civ  int // -1 if none
	Year Year
}

// Config tunes the simulation.
type Config struct {
	Region    string // where in the galaxy: a preset, a feature name, or x,y,z in kpc
	Stars     int
	Radius    float64
	Thickness float64
	DeepStart Year // substrate begins
	Dawn      Year // the current age dawns; the engine runs from here
	DeepStep  Year
	Step      Year // the tick of the current age, dawn to present; every rate is per thousand years
	// the waning: declared when this few are active and fertility is this low
	FineActive    int
	FineFertility float64
	// the present: stop when this few are active and fertility is below a threshold drawn
	// per world between EndFertilityLow and EndFertility, then linger a while
	EndActive       int
	EndFertility    float64
	EndFertilityLow float64
	Linger          Year
	MaxFades        float64      // give up after this many fades and flag it
	Debug           bool         // log the state of the galaxy every million years
	TraceAI         bool         // log every council's reasoning
	Profile         bool         // log each phase's time every million years
	Tuning          *mind.Tuning // every number the decisions use; nil means mind.Default()
}

// DefaultConfig is a small, fast world.
func DefaultConfig() Config {
	return Config{
		Stars: 400, Radius: 150, Thickness: 40,
		DeepStart: -galaxy.Age, Dawn: 0,
		DeepStep: 10_000_000, Step: 1_000,
		FineActive: 12, FineFertility: 0.5,
		EndActive: 5, EndFertility: 0.2, EndFertilityLow: 0.05, Linger: 2_000_000,
		MaxFades: 8,
		Tuning:   mind.Default(),
	}
}

// World is the whole simulated history.
type World struct {
	Cfg       Config
	Seed      uint64
	G         *galaxy.Galaxy
	Law       galaxy.Law // the laws of the place
	R         *rand.Rand
	Now       Year
	Present   Year    // when the simulation stopped; years are printed relative to this
	Waning    Year    // when the waning was declared
	Capped    bool    // the age never ended on its own; stopped at MaxFades
	Ticks     int     // ticks run in the current age
	dt        float64 // current tick in kyr
	phases    []phase // the tick, in order; see sim.go
	phaseTime map[string]time.Duration
	Bio       []BioState
	Owner     []int // civ id owning each star, -1 if none
	Held      []int // horror id holding each star, -1 if none
	Hazard    float64
	Thin      float64 // how worn the wall between this and the state beneath is; see beneath.go
	ThinStage int
	Civs      []*Civ
	Species   []*species.Species // every blood that has arisen or been made, by ID
	Sources   []*Source          // everything with a yield, by ID; see sources.go
	sourcesAt [][]int            // the sources that yield at each star, ranged ones included
	mobile    []int              // the sources that move with a holder, by ID; see rarity.go
	Horrors   []*Horror
	Ages      []*AgeRecord
	Cycle     *Cycle
	Legacies  []*Legacy
	Traces    []Trace
	Events    []Event
	Facts     []*Fact       // what happened, as it happened; see lore.go
	factsAt   map[int][]int // facts by star
	// war and diplomacy
	Wars        []*War
	Expeditions []*Expedition
	Pacts       []*Pact
	Messages    []*Message
	Betrayals   []Betrayal // and faith kept, with negative weight
	scratch
}

// scratch passed from a trigger to its filter's outcome functions
type scratch struct {
	blastWorlds []int
	blastWhat   string
	beacon      *Horror
	incursionAt int
	incursionBy *Horror
	wreck       *Wreckage   // set while a filter's outcome runs
	finding     bool        // set while the Find teaches a civilisation what it mastered
	fleet       *Expedition // the fleet taking a world, while it does; what it carries off rides with it
	loose       []int       // objects that got loose in a taking this tick, by source; see objects.go
}
