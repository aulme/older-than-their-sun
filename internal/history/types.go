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
	"worldgen/internal/record"

	"worldgen/internal/flow"
	"worldgen/internal/galaxy"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Year is years since the dawn of the current age (negative = the
// ages of myth): the record's own type.
type Year = record.Year

// Kind is what an event is: a key in data/events.json, where its
// parameters, its meaning and, for a fact, its sort and weight are
// declared. The constants are generated from the file (kinds_gen.go):
// F-kinds are facts, the ones a people can hold a tale of; K-kinds are
// the rest of the chronicle. The type is the record's, so the readers of
// the files compare against the same constants.
type Kind = record.Kind

// P is an event's parameters: the keys its kind declares, each an id, a
// number, a key or a flag. What is text today is a debt to the lookups
// step (specs/plan.md step 3).
type P map[string]any

// Event is one thing that happened, as it happened: a typed record. The
// chronicle is every event in year order; the view renders a line per
// kind from the record (lines.go), the simulation writes no prose. A fact
// is an event whose kind has a shape, a sort and a weight (lore.go).
type Event struct {
	ID      int
	Year    Year
	Kind    Kind
	Subject int          // the people it is about, or -1
	Object  int          // the other people, or -1
	Star    int          // where, or -1
	Legacy  int          // the remain in it, or -1
	Plague  int          // the plague in it, or -1
	N       int          // a count: worlds
	P       P            // the kind's parameters
	sh      *shape       // the kind's shape, looked up once; see lore.go
	row     *[8]judgment // the kind's row of the moral table, or nil; see morality.go
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
	Sundered  // tore itself into heirs; see sunder.go
	Shattered // forgot the stars and became one people per world
)

func (f Fate) String() string {
	return [...]string{"active", "extinct", "transformed", "contracted", "sundered", "shattered"}[f]
}

// Voyage is a colony ship in flight.
type Voyage struct {
	Target int
	Arrive Year
	Blind  bool // sent to a star nobody has read, on a guess
}

// Origin is how a people came to be when it did not arise: a key of
// data/origins.json and the parties. It is the same shape as a blood's
// making (species.Making), read through the same table.
type Origin = species.Making

// Civ is a civilisation: a species on a home world with a history.
type Civ struct {
	ID        int              // the people's id; every name it carries is a row of the names pass, keyed by it
	Species   *species.Species // its blood, one of w.Species; kin share it, and it changes only by a made path
	Origin    Origin           // how the people came to be when not by arising; the zero key for a cradle people
	Home      int              // current seat; moves if the cradle is lost
	Cradle    int              // the world the people arose on; never changes
	Born      Year
	Ended     Year
	Fell      Year // when it stopped being active
	Stage     Stage
	Fate      Fate
	Cause     string // why it fell or ended: a key of data/causes.json; the fall and end facts carry the parties
	Into      string // what it became, if transformed: a key of data/causes.json's becomings; IntoCivs the peoples it became
	IntoCivs  []int
	FallEvent int             // the fall fact, -1 for none
	EndEvent  int             // the end fact, -1 for none
	Named     bool            // has a word for the state beneath: an FWord fact of its own or its line's; see beneath.go
	Hosts     int             // for parasites: peoples ridden or fighting its plague
	Morality  Morality        // what the people counts as wrong; see morality.go
	Lifted    map[string]bool // world blocks lifted by a colony: sea, sky, fire
	Systems   []int
	Peak      int
	Voyages   []Voyage
	colonies  int

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
	Wis           float64 // see wisdom.go: the capacity to see the counterintuitive
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
	// useNodes is what this people knows that the upkeep pass walks, in
	// the order it learned them: the answer to a scan of every node of
	// the tree and a sort, which the pass wanted afresh every tick
	// though it changes only when a node is learned or lost. know,
	// forgetNode and learn drop it; uses rebuilds it. The needs
	// themselves are still read every tick, so a people whose profile
	// bends its upkeep is unaffected.
	useNodes   []*tech.Node
	useNodesOK bool
	LastDark   Year
	Summoned   bool // an event calls the council this tick
	Aloft      bool // a nomad people living as fleets, with no worlds
	Rested     bool // a nomad people that came to rest, and will not rise again

	// ships: see ships.go
	DockRate   map[int]float64 // the rate each dock worked at last tick, by star: its draw this tick
	WantShips  int             // the need of the campaign the council last sized and could not man
	WantSince  Year            // when; the want stands for a while
	SurveyWant int             // surveyors the exploring policy asked for last tick, before the ships were counted
	PeakShips  int             // the most ships ever in being, for the batch
	firstShip  bool            // the dock's first ship was written

	// guns: see guns.go
	Guns       map[int]int  // guns standing over each star, of the gun structures there
	GridBroken map[int]bool // stars whose grid was shot to nothing, until it is whole again

	// garrisons and the muster: see garrison.go
	GarrisonWant int     // what the garrison policy asked for last tick, summed over the holdings
	Muster       *Muster // the campaign gathering, if one is

	// wisdom: see wisdom.go
	Sire        int          // the people that raised or bred this one, -1: kin understand each other at once
	Fathomed    map[int]bool // the peoples this one understands
	FathomTried map[int]Year // when this one first tried to understand each: the meeting, or the dark age that lost it
	PeakWis     float64      // the most Wisdom ever had, for the batch
	WisFrom     wisParts     // where the Wisdom comes from, as last derived
	experience  int          // woes and follies remembered as one's own, counted by reckon

	// contracts: see contract.go
	Contracts []int          // every contract this people was party to
	Taught    map[string]int // nodes bought: the node to the people that taught it
	Paid      flow.Income    // what contracts paid this tick; counted in the next tick's income
	PaidIn    flow.Income    // what contracts paid into this tick's income
	Sellsword bool           // earned: lived on contract pay for ten ticks running
	hiredRun  int            // ticks running that contract pay was more than half an income
	dealt     map[int]bool   // peoples a bargain has been struck with, for the first line

	// plagues: see plague.go
	Infections  map[int]*Infection // the plagues this people has, by plague
	Immune      map[int]bool       // the plagues it cannot catch again
	Suspect     map[int]bool       // the peoples it thinks are sick
	Closed      map[int]bool       // the suspects it has closed its ears and its ports to
	LastTaken   Year               // when a world last changed hands with it on either side
	FirstPlague Year               // when it first had one, for the batch
	Own         int                // for parasites: the plague it is; -1 for a people that is not one
	Weapons     map[string]*Weapon // the plagues it has made and holds, by the craft that made them
	Barred      map[int]bool       // the peoples whose goods and messages it refuses for good: those that tried to poison it
	Gross       flow.Income        // the income before a rider's tithe

	// ossification and lines: see ossify.go and sunder.go
	Stiff    float64      // how far the people's ways have set: zero young, one an empire whose ways are fixed, past two Trantor
	Ossified bool         // set: acting every other tick, half the research, twice the building
	Still    Year         // when something new last happened: a war, a world, a node, a meeting
	Line     []int        // the peoples this one came out of by a sundering or a shattering, oldest first
	Claim    map[int]bool // the worlds of the old realm an heir holds itself owed

	// kinds: see eldritch.go and waking.go
	Asleep     bool               // the long sleep: it sits every tick out but its guns until disturbed
	Slept      Year               // when it last went to sleep
	LastUnmade Year               // when the unmaking was last turned on a world
	demanded   map[int]Year       // when a living world last told each people to leave its neighbourhood
	tithedBy   map[int]bool       // the peoples whose tithe this one has paid, for the first line
	Drifts     int                // for an evolver: how many times its shape has drifted; see evolver.go
	FathomedAt map[int]int        // the drift count of each people fathomed, at the fathoming: the people understood is that one
	heldKyr    map[string]float64 // thousand years of holding worlds of each type, for the drift

	// sightings and salvage: see sighting.go, field.go
	Sightings    map[int]*Sighting // what this people has seen of fleets in flight, by fleet
	Salvage      int               // ships of others' make in the guards, crewed from a field
	SalvageTaken int               // ships taken from fields since the last were lost: a tenth is lost each thousand years

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
	PeakTrade    []int // the partners at the height of means, for the bloc reading; see slight.go
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
	Record       []Record // what it did that its record keeps, in order; see Record
	DarkAges     int
	KnowsCycle   bool // learned the shape of the cycle
	Ascended     Year // when a miracle was last gained; the surge runs from here
	Renewed      Year
	Renaissances int
	NextDrift    int     // size at which the Distance is faced again
	Dying        bool    // home star is failing
	Endure       float64 // kyr left under the failing star
}

// Tally counts what a people did in war and peace, for the batch reports:
// a record type, written to the state as it is.
type Tally = record.Tally

// Living is true for active and remnant civilisations.
func (c *Civ) Living() bool { return c.Stage != Dead }

// Active is true for civilisations still growing.
func (c *Civ) Active() bool { return c.Stage < Remnant }

// Free is true for civilisations not held by another.
func (c *Civ) Free() bool { return c.Master < 0 }

// Has reports a species trait.
func (c *Civ) Has(trait string) bool { return c.Species.Has(trait) }

// Payload is what a transmitter carries: a corruption, which is the
// Signal as a map hazard; or a seed, a conscious memetic plague that wakes
// a mind-rider in whoever's home goes over to it. See transmitter.go.
type Payload uint8

const (
	Corruption Payload = iota
	Seed
)

func (p Payload) String() string { return [...]string{"corruption", "seed"}[p] }

// LegacyKind is what an age leaves behind.
type LegacyKind uint8

const (
	Artifact LegacyKind = iota
	Structure
	Threat
	Sleeper
	Law
	Bounty // a thing still doing what it was made to do: a yield for whoever puts it to use
	Field  // the wrecks and derelicts of a fleet, where they fell; see field.go
)

func (k LegacyKind) String() string {
	return [...]string{"artifact", "structure", "threat", "sleeper", "law", "bounty", "field"}[k]
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
	Portrait  string // what it looks like: a key of data/portraits.json by kind (an elder's list, a relic, a threat, a sleeper, a bounty), a work's key for a remain, a form's for an object; the view renders it
	State     LegacyState
	People    int     // the people a threat or a sleeper is: asleep at the star until disturbed; -1 for none
	Payload   Payload // what a transmitter carries
	Listeners int     // peoples a transmitter has taken: scarred, ended, or seeded
	Woken     int     // parasites a transmitter's seed woke
	Finder    int     // civ that last acted on it, -1 if none
	Level     string  // for wielded artifacts: which level it lifts, or "miracle"
	Cond      Condition
	Hardy     float64       // multiplier on the rate of decay; 0 never decays
	Source    int           // the source record it is, for a bounty or a wielded artifact; -1 if none
	Testament []Inscription // what its makers told of their age, as they left it
	Plague    int           // the sickness of the mind its makers had when they wrote, -1 for none; see plague.go
	// a field of wrecks
	Wrecks, Derelicts int
	At                vec  // where it is, for a field adrift
	Adrift            bool // between stars: Star is the nearer end
}

// Elder is a civilisation of an earlier age. No traits, only a portrait.
type Elder struct {
	ID       int // across the ages, in order of making; finders' names for it are rows of the names pass
	Age      int
	Portrait string // a key of data/portraits.json's elders
	Rose     Year
	Fell     Year
	Legacies []*Legacy
}

// Record is one entry of a people's record: a filter faced with its
// outcome (narrow if the margin was slight), one foreseen, a remain
// mastered, wielded, sealed or unleashed, a bounty used, wrecks crewed,
// the sky taken, a rest. Kind is one of faced, foresaw, mastered,
// bounty, crewed, wielded, sealed, unleashed, sky, rest; Filter and
// Legacy are set where the kind names one, Known whether the people could
// place the remain's makers then.
type Record struct {
	Kind    string
	Filter  string
	Outcome Outcome
	Narrow  bool
	Legacy  int
	Known   bool
}

// AgeRecord is one earlier age of the galaxy.
type AgeRecord struct {
	Index  int
	Start  Year   // the surge
	End    Year   // fertility below the floor
	Ender  string // what swept up the remains: a key of data/portraits.json's enders
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
	Star    int
	Kind    string // a key of data/conditions.json's traces
	Civ     int    // -1 if none
	Species int    // the blood that lost the world, for the word its kind uses; -1 if none
	Year    Year
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
	// the waning: declared when this few are still rising (zero for no bar) and fertility is this low
	FineActive    int
	FineFertility float64
	// the present: stop when this few are still rising (zero for no bar) and fertility is
	// below a threshold drawn per world between EndFertilityLow and EndFertility, then
	// linger a while
	EndActive       int
	EndFertility    float64
	EndFertilityLow float64
	Linger          Year
	MaxFades        float64 // give up after this many fades and flag it
	Until           Year    // stop at this year of the age instead of the age's own end, for a state as of then; 0 for none
	// WearEvery is how many ticks apart a people's telling is put
	// through the wearing. 1 is every tick. Above that the rate of each
	// tale is compounded over the gap, so a tale wears as often as it
	// did; what changes is that it can only do so on the ticks its
	// people is due, and the peoples are staggered so no tick bears
	// them all. 0 means 1. See specs/plan.md, step 6.
	WearEvery int
	Debug     bool         // log the state of the galaxy every million years
	TraceAI   bool         // log every council's reasoning
	Profile   bool         // log each phase's time every million years
	Tuning    *mind.Tuning // every number the decisions use; nil means mind.Default()
}

// DefaultConfig is a small, fast world.
func DefaultConfig() Config {
	return Config{
		Stars: 400, Radius: 150, Thickness: 40,
		DeepStart: -galaxy.Age, Dawn: 0,
		DeepStep: 10_000_000, Step: 1_000,
		FineActive: 0, FineFertility: 0.5,
		EndActive: 0, EndFertility: 0.2, EndFertilityLow: 0.05, Linger: 2_000_000,
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
	Truncated bool    // stopped at Cfg.Until, before the age's own end
	Ticks     int     // ticks run in the current age
	dt        float64 // current tick in kyr
	phases    []phase // the tick, in order; see sim.go
	// regard held for the length of one retelling: revise asks how a
	// people regards the other party of every tale it holds, and a
	// people holds hundreds of tales about a handful of others. The
	// stamp saves clearing the table; nothing revise writes is anything
	// regard reads, so the answer cannot change under it.
	regardOf   []int8
	regardSeen []uint32
	regardGen  uint32

	phaseTime  map[string]time.Duration
	stepTime   map[string]time.Duration // under -phases: what each civ step cost
	stepDraws  map[string]uint64        // under -phases: what each civ step drew from the RNG
	draws      *drawCount               // the tally behind R, under -phases
	Bio        []BioState
	Owner      []int // civ id owning each star, -1 if none
	Hazard     float64
	Thin       float64 // how worn the wall between this and the state beneath is; see beneath.go
	ThinStage  int
	Civs       []*Civ
	Species    []*species.Species // every blood that has arisen or been made, by ID
	Sources    []*Source          // everything with a yield, by ID; see sources.go
	sourcesAt  [][]int            // the sources that yield at each star, ranged ones included
	mobile     []int              // the sources that move with a holder, by ID; see rarity.go
	Ages       []*AgeRecord
	Cycle      *Cycle
	Fathomings []Fathoming // every understanding reached, for the batch; see wisdom.go
	Legacies   []*Legacy
	Traces     []Trace
	Events     []*Event      // what happened, as it happened, by id; see lore.go
	Chronicle  []*Event      // the same in the order it is told, and by year once the run is over
	notes      []*Event      // the silent events of the run, placed after the told ones at the end; see notes.go
	factsAt    map[int][]int // facts by star
	// war and diplomacy
	Wars        []*War
	Battles     []*Battle   // every battle at a world, for the batch; see battle.go
	Meetings    []*Meeting  // every battle in the dark, for the batch; see intercept.go
	Watch       []*Sighting // every sighting, for the batch; see sighting.go
	pending     []*Sighting // sightings queued on the timetable, not yet happened
	Expeditions []*Expedition
	fleetsBy    map[int][]*Expedition // the fleets not over, by owner: an index kept by addExpedition, reown and fleetsOf; see ships.go
	live        []*Expedition         // the fleets not over, all owners: kept by addExpedition and liveFleets
	Pacts       []*Pact
	Contracts   []*Contract        // every contract offered; see contract.go
	Plagues     []*Plague          // every plague born; see plague.go
	Reservoir   map[int]*Reservoir // plagues waiting in dead cities, by star
	wokeOnMaker bool               // the plague that woke on its maker has been told: once per galaxy
	Messages    []*Message
	Betrayals   []Betrayal // and faith kept, with negative weight
	scratch
}

// scratch passed from a trigger to its filter's outcome functions
type scratch struct {
	blastWorlds []int
	blastRef    reason       // the blast being suffered, for the causes it gives; for a waking, the waker
	transmitter *Legacy      // the transmitter being listened to, for the Signal's filter
	wreck       *Wreckage    // set while a filter's outcome runs
	finding     bool         // set while the Find teaches a civilisation what it mastered
	fleet       *Expedition  // the fleet taking a world, while it does; what it carries off rides with it
	emptySky    bool         // the world being taken had nothing in its sky
	foughtAt    map[int]Year // the last tick a battle was fought at each star: one a tick
	loose       []int        // objects that got loose in a taking this tick, by source; see objects.go
}
