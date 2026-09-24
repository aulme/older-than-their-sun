// Package record is the output of a run as records: the shape of
// dossier.json, state.json, chronicle.jsonl and tellings.jsonl, and the
// reading and writing of a run directory. Nothing here decides
// anything; the simulation writes these through the writer and the
// view and the batch reports read them. FORMAT.md at the repository
// root is the contract these types keep, field for field.
//
// Every id is the simulation's small integer, stable within a run, -1
// for none. Every year is years since the dawn of the age, negative in
// the ages of myth; the dossier gives the present. Every enum is a key
// that resolves in a table under data/. No record carries a sentence
// the simulation wrote, and no record carries a name: names are the
// names[] table of the state, one row per (object, namer, tone).
package record

import (
	"worldgen/internal/flow"
	"worldgen/internal/galaxy"
	"worldgen/internal/mind"
	"worldgen/internal/plague"
	"worldgen/internal/species"
)

// Format is the version of the contract these types keep; the dossier
// carries it, and FORMAT.md describes it.
const Format = 1

// Year is years since the dawn of the current age.
type Year int64

// Run is one run's files as loaded: the dossier, the state, the
// chronicle in year order, the tellings, and the names.
type Run struct {
	Dossier   *Dossier
	State     *State
	Chronicle []*Event
	Tellings  []*Tale
}

// Dossier is dossier.json: the run's header, small enough to sit in a
// prompt whole.
type Dossier struct {
	Format    int     `json:"format"`
	Code      string  `json:"code"` // the git revision that produced the run, or "unknown"
	Seed      uint64  `json:"seed"`
	Present   Year    `json:"present"`   // when the run stopped
	Truncated bool    `json:"truncated"` // written as of a year before the run's own end (-until)
	Waning    Year    `json:"waning"`    // when the waning was declared
	DeepStart Year    `json:"deep_start"`
	Dawn      Year    `json:"dawn"`
	Step      Year    `json:"step"` // the tick of the current age, in years
	Ticks     int     `json:"ticks"`
	Capped    bool    `json:"capped"` // the age never wound down on its own
	MaxFades  float64 `json:"max_fades"`
	Config    Config  `json:"config"`
	Place     Place   `json:"place"`
	Cycle     Cycle   `json:"cycle"`
	Fertility float64 `json:"fertility"` // now, as a fraction of the dawn's
	NextDawn  Year    `json:"next_dawn"`
	Hazard    float64 `json:"hazard"`
	Wall      Wall    `json:"wall"`
	Decline   Decline `json:"decline"`
	Counts    Counts  `json:"counts"`
}

// Decline is the decline index at the present, its three terms, the
// age's own running peaks and the absolute numbers the terms are shares
// of. Every term is relative, against the peak the age itself reached,
// because an age that never had a height has not fallen from one; the
// absolute numbers are here beside them for a reader comparing two
// galaxies, which the relative reading hides.
type Decline struct {
	Index      float64 `json:"index"`  // 0 at the height, 1 with nothing standing: the mean of the three terms
	Held       float64 `json:"held"`   // habitable systems held, over the highest that share has been
	Rising     float64 `json:"rising"` // peoples still rising, over the highest that count has been
	Births     float64 `json:"births"` // peoples born per Myr, over the highest that rate has been
	HeldNow    float64 `json:"held_now"`
	RisingNow  int     `json:"rising_now"`
	BirthsNow  float64 `json:"births_now"`
	PeakHeld   float64 `json:"peak_held"`
	PeakRising int     `json:"peak_rising"`
	PeakBirths float64 `json:"peak_births"`
	PeakHeldAt Year    `json:"peak_held_at"` // when the held share was highest: the age's height
	AtWaning   float64 `json:"at_waning"`    // the index at the moment the waning was declared
	Crossed    Year    `json:"crossed"`      // when the index first stood at the waning bar; 0 for never
	Fell       Year    `json:"fell"`         // and at the end bar
	ByIndex    bool    `json:"by_index"`     // the age ended on the index rather than falling through to the fertility floor
}

// Config is what the run was asked for, besides the seed.
type Config struct {
	At        string       `json:"at"` // the place, as given: a preset, a feature name, or x,y,z in kpc
	Stars     int          `json:"stars"`
	Radius    float64      `json:"radius"`
	Thickness float64      `json:"thickness"`
	Tuning    *mind.Tuning `json:"tuning"`
}

// Place is where in the galaxy the field is and the laws there.
type Place struct {
	Name     string        `json:"name"`
	Code     string        `json:"code"` // the designation prefix of the field's own stars
	Anchor   string        `json:"anchor"`
	Pos      galaxy.Vec    `json:"pos"` // kpc, galactocentric
	HasSol   bool          `json:"has_sol"`
	Sol      int           `json:"sol"` // the star that is the Sun, or -1
	Real     int           `json:"real"`
	Zone     string        `json:"zone"`
	Laws     Laws          `json:"laws"`
	Arm      string        `json:"arm,omitempty"` // the arm the place is in
	ArmDist  float64       `json:"arm_dist"`
	Near     []NearFeature `json:"near"`
	Beyond   []string      `json:"beyond"` // the sky features, by key
	MeanGap  float64       `json:"mean_gap"`
	FieldRad float64       `json:"field_radius"`
}

// Laws are the numbers of the place, each relative to the Sun's
// neighbourhood but Metals.
type Laws struct {
	R       float64 `json:"r"`
	Z       float64 `json:"z"`
	Density float64 `json:"density"`
	Youth   float64 `json:"youth"`
	Metals  float64 `json:"metals"`
	Glare   float64 `json:"glare"`
	Crowd   float64 `json:"crowd"`
	Exotic  float64 `json:"exotic"`
}

// NearFeature is a catalogued thing near the field.
type NearFeature struct {
	Key  string  `json:"key"`
	Dist float64 `json:"dist"` // kpc
}

// Cycle is the galaxy's cycle of dawns.
type Cycle struct {
	Period Year    `json:"period"`
	Fade   Year    `json:"fade"`
	Surges []Year  `json:"surges"`
	Floor  float64 `json:"floor"`
	Ends   float64 `json:"ends"`
}

// Wall is the wall between this and what is beneath it.
type Wall struct {
	Value float64 `json:"value"`
	Stage int     `json:"stage"`
}

// Counts are the numbers the aftermath's header prints.
type Counts struct {
	Civs      int `json:"civs"`
	Standing  int `json:"standing"`
	Remnants  int `json:"remnants"`
	Eaten     int `json:"eaten"`    // stars held by things that eat them
	Speaking  int `json:"speaking"` // transmitters live
	Complex   int `json:"complex"`  // worlds with complex life
	Remains   int `json:"remains"`
	Crumbled  int `json:"crumbled"`
	Traces    int `json:"traces"`
	Events    int `json:"events"`
	Tales     int `json:"tales"`
	Names     int `json:"names"`
	Plagues   int `json:"plagues"`
	Wars      int `json:"wars"`
	Leaders   int `json:"leaders"`
	Fleets    int `json:"fleets"`
	Contracts int `json:"contracts"`
}

// State is state.json: the galaxy at the present as objects, by id.
type State struct {
	Stars      []*Star       `json:"stars"`
	Civs       []*Civ        `json:"civs"`
	Species    []*Species    `json:"species"`
	Ages       []*Age        `json:"ages"`
	Elders     []*Elder      `json:"elders"`
	Remains    []*Remain     `json:"remains"`
	Sources    []*Source     `json:"sources"`
	Traces     []Trace       `json:"traces"`
	Plagues    []*Plague     `json:"plagues"`
	Reservoirs []Reservoir   `json:"reservoirs"`
	Wars       []*War        `json:"wars"`
	Leaders    []*Leader     `json:"leaders"`
	Pacts      []*Pact       `json:"pacts"`
	Betrayals  []Betrayal    `json:"betrayals"`
	Fleets     []*Fleet      `json:"fleets"`
	Contracts  []*Contract   `json:"contracts"`
	Fathomings []Fathoming   `json:"fathomings"`
	Battles    []Battle      `json:"battles"`
	Meetings   []Meeting     `json:"meetings"`
	Sightings  []Sighting    `json:"sightings"`
	Names      []Name        `json:"names"`
	Voices     map[int]Voice `json:"voices"` // how each people names, by id
}

// Star is one star of the field with its system, and what is there now.
type Star struct {
	ID          int            `json:"id"`
	Designation string         `json:"designation"` // the human catalogue's label for a real star, a code for a synthetic one
	Kind        string         `json:"kind"`        // what the designation is: proper, bayer, flamsteed, hd, hip, gj, gaia, 2mass, catalogue, code
	Real        bool           `json:"real"`
	Alt         string         `json:"alt,omitempty"`
	Class       string         `json:"class"` // O B A F G K M; W a white dwarf, N a neutron star or black hole
	Note        string         `json:"note,omitempty"`
	Remnant     string         `json:"remnant,omitempty"`
	X           float64        `json:"x"`
	Y           float64        `json:"y"`
	Z           float64        `json:"z"`
	Hab         float64        `json:"hab"`
	Mult        int            `json:"mult"`
	Lifetime    float64        `json:"lifetime"`
	DiesAt      Year           `json:"dies_at"`
	Failing     bool           `json:"failing"`
	Mag         float64        `json:"mag"`
	System      *galaxy.System `json:"system"`
	Bio         string         `json:"bio"`  // none, simple, complex
	Held        int            `json:"held"` // the people holding it, or -1
	Guns        int            `json:"guns"`
	GridBroken  bool           `json:"grid_broken"`
}

// Species is a blood.
type Species struct {
	ID      int            `json:"id"`
	Sub     string         `json:"sub"`
	Mods    []string       `json:"mods"`
	Channel string         `json:"channel"`
	Powers  []string       `json:"powers"`
	World   string         `json:"world"`
	Traits  []string       `json:"traits"`
	Made    species.Making `json:"made"`
	Parent  int            `json:"parent"`
	First   int            `json:"first"` // the first people of the blood, whose name it goes by
	// Lifespan is the years a body of the blood lives before its
	// medicine; 0 for a blood that does not turn over.
	Lifespan int  `json:"lifespan"`
	Software bool `json:"software,omitempty"` // uploaded and kept going: deathless, in machines
}

// Civ is a people.
type Civ struct {
	ID        int            `json:"id"`
	Species   int            `json:"species"`
	Origin    species.Making `json:"origin"`
	Home      int            `json:"home"`
	Cradle    int            `json:"cradle"`
	Born      Year           `json:"born"`
	Ended     Year           `json:"ended"`
	Fell      Year           `json:"fell"`
	Stage     string         `json:"stage"` // emergent, interstellar, zenith, remnant, dead
	Fate      string         `json:"fate"`  // active, extinct, transformed, contracted, sundered, shattered
	Cause     string         `json:"cause"`
	Into      string         `json:"into,omitempty"`
	IntoCivs  []int          `json:"into_civs,omitempty"`
	FallEvent int            `json:"fall_event"`
	EndEvent  int            `json:"end_event"`
	Named     bool           `json:"named"`
	Hosts     int            `json:"hosts"`
	Morality  Morality       `json:"morality"`
	Lifted    []string       `json:"lifted"`
	Systems   []int          `json:"systems"`
	Peak      int            `json:"peak"`
	Voyages   []Voyage       `json:"voyages"`

	Known    []string          `json:"known"`
	Dormant  []string          `json:"dormant"`
	Learned  map[string]Year   `json:"learned"`
	Era      int               `json:"era"`
	Pursuit  string            `json:"pursuit,omitempty"`
	Progress float64           `json:"progress"`
	Miracles map[string]string `json:"miracles"`
	Remade   map[string]Year   `json:"remade"`

	Levels   Levels      `json:"levels"`
	Reach    float64     `json:"reach"`
	Speed    float64     `json:"speed"`
	Envelope int         `json:"envelope"`
	Morale   float64     `json:"morale"`
	Dials    mind.Dials  `json:"dials"`
	Order    []string    `json:"order"`
	Income   flow.Income `json:"income"`
	Upkeep   flow.Income `json:"upkeep"`
	Surplus  flow.Income `json:"surplus"`
	Want     flow.Income `json:"want"`

	Structures map[string]int `json:"structures"`
	Works      []Work         `json:"works"`
	Wielded    []int          `json:"wielded"`
	Found      []int          `json:"found"`
	Heard      []int          `json:"heard"`
	Uplifts    int            `json:"uplifts"`
	Ruled      int            `json:"ruled"`
	Docks      int            `json:"docks"`
	Salvage    int            `json:"salvage"`

	Wars      []int           `json:"wars"`
	Trade     []int           `json:"trade"`
	Master    int             `json:"master"`
	Vassal    bool            `json:"vassal"`
	Pacts     []int           `json:"pacts"`
	Truce     map[int]Year    `json:"truce"`
	Fought    map[int]int     `json:"fought"`
	Grudge    map[int]float64 `json:"grudge"`
	Ridden    []int           `json:"ridden"`
	Contracts []int           `json:"contracts"`
	Taught    map[string]int  `json:"taught"`
	Sellsword bool            `json:"sellsword"`
	Dependent []int           `json:"dependent"`
	Embargo   []int           `json:"embargo"`
	Refused   map[int]Year    `json:"refused"`
	Barred    []int           `json:"barred"`

	Infections map[int]Infection `json:"infections"`
	Immune     []int             `json:"immune"`
	Suspect    []int             `json:"suspect"`
	Closed     []int             `json:"closed"`
	Own        int               `json:"own"`
	Weapons    map[string]Weapon `json:"weapons"`

	Stiff        float64  `json:"stiff"`
	Continuity   float64  `json:"continuity"` // the share of its past that reaches across a thousand years
	Lifespan     float64  `json:"lifespan"`   // the years a body lives now, with its medicine and its sickness; 0 for none
	Leader       int      `json:"leader"`     // the leader it follows now, -1 for none
	Ossified     bool     `json:"ossified"`
	Still        Year     `json:"still"`
	Line         []int    `json:"line"`
	Claim        []int    `json:"claim"`
	Asleep       bool     `json:"asleep"`
	Slept        Year     `json:"slept"`
	Aloft        bool     `json:"aloft"`
	Rested       bool     `json:"rested"`
	Drifts       int      `json:"drifts"`
	Searching    bool     `json:"searching"`
	Starfaring   Year     `json:"starfaring"`
	Faced        []string `json:"faced"`
	Scars        []string `json:"scars"`
	Boons        []string `json:"boons"`
	Record       []Record `json:"record"`
	DarkAges     int      `json:"dark_ages"`
	KnowsCycle   bool     `json:"knows_cycle"`
	Ascended     Year     `json:"ascended"`
	Renewed      Year     `json:"renewed"`
	Renaissances int      `json:"renaissances"`
	Dying        bool     `json:"dying"`
	Endure       float64  `json:"endure"`
	Rare         []string `json:"rare"`
	LastDark     Year     `json:"last_dark"`
	LastTaken    Year     `json:"last_taken"`
	LastUnmade   Year     `json:"last_unmade"`

	Knowledge Knowledge `json:"knowledge"`
	Batch     Batch     `json:"batch"`
}

// Levels are a people's derived levels this tick.
type Levels struct {
	Mil float64 `json:"military"`
	Sur float64 `json:"survival"`
	Soc float64 `json:"social"`
	Wis float64 `json:"wisdom"`
}

// Morality is what a people counts as wrong.
type Morality struct {
	Kind   string `json:"kind"` // amoral, individual, herd, fixation
	Object string `json:"object,omitempty"`
}

// Voyage is a colony ship in flight.
type Voyage struct {
	Target int  `json:"target"`
	Arrive Year `json:"arrive"`
	Blind  bool `json:"blind"`
}

// Work is one structure standing at a star.
type Work struct {
	Key    string `json:"key"`
	Node   string `json:"node"`
	Star   int    `json:"star"`
	Legacy int    `json:"legacy"`
	Dark   bool   `json:"dark"`
}

// Record is one entry of a people's record.
type Record struct {
	Kind    string `json:"kind"`
	Filter  string `json:"filter,omitempty"`
	Outcome string `json:"outcome,omitempty"` // overcame, scarred, declined
	Narrow  bool   `json:"narrow,omitempty"`
	Legacy  int    `json:"legacy"`
	Known   bool   `json:"known,omitempty"`
}

// Infection is a plague a people has.
type Infection struct {
	Since     Year   `json:"since"`
	From      int    `json:"from"`
	Road      string `json:"road"`
	Contained bool   `json:"contained"`
	Held      int    `json:"held"`
	Carrier   bool   `json:"carrier"`
}

// Weapon is a plague a people made and holds.
type Weapon struct {
	Plague int    `json:"plague"`
	Target int    `json:"target"`
	Node   string `json:"node"`
	Made   Year   `json:"made"`
}

// Knowledge is what a people knows, believes and has forgotten: the only
// place ignorance lives in the output.
type Knowledge struct {
	Met         []int           `json:"met"`
	Reached     []int           `json:"reached"`
	Fathomed    []int           `json:"fathomed"`
	FathomTried map[int]Year    `json:"fathom_tried"`
	FathomedAt  map[int]int     `json:"fathomed_at"`
	Alien       map[int]float64 `json:"alien"`     // how alien each people tried is, as the fathoming counts it
	Perceives   []int           `json:"perceives"` // the peoples it can hold in mind, of every people there is
	Charted     map[int]Year    `json:"charted"`
	Marked      []int           `json:"marked"`
	Scouted     map[int]Year    `json:"scouted"`
	Intel       map[int]Intel   `json:"intel"`
	Regards     map[int]int8    `json:"regards"` // -2 a monster, -1 an enemy, 0 a stranger, 1 a friend, 2 its own line
	Monsters    []int           `json:"monsters"`
	Dread       []int           `json:"dread"` // stars it will not go to
	Scapegoat   int             `json:"scapegoat"`
	Watched     []int           `json:"watched"`
	Sightings   map[int]int     `json:"sightings"` // fleets seen, to the sighting record
	Memory      Memory          `json:"memory"`
	LoreDials   mind.Dials      `json:"lore_dials"`
}

// Memory is the standing count of a people's tales.
type Memory struct {
	Held    int `json:"held"`
	Myth    int `json:"myth"`
	Forgot  int `json:"forgot"`
	Revised int `json:"revised"`
	Blamed  int `json:"blamed"`
}

// Intel is what a people believes of another, as of a year.
type Intel struct {
	Mil    float64 `json:"mil"`
	Ships  int     `json:"ships"`
	Guns   int     `json:"guns"`
	Total  int     `json:"total"`
	Relief float64 `json:"relief"`
	Star   int     `json:"star"`
	Year   Year    `json:"year"`
	Sick   int     `json:"sick"`
}

// Batch is what the batch reports read of a people and nothing else
// does: counters kept for tuning, snapshot only.
type Batch struct {
	Tally         Tally           `json:"tally"`
	HighIncome    flow.Income     `json:"high_income"`
	HighUpkeep    flow.Income     `json:"high_upkeep"`
	HighWant      flow.Income     `json:"high_want"`
	PeakTrade     []int           `json:"peak_trade"`
	ShedTicks     map[string]int  `json:"shed_ticks"`
	Built         map[string]int  `json:"built"`
	Had           []string        `json:"had"`
	Harnessed     []string        `json:"harnessed"`
	Granted       []string        `json:"granted"`
	FellDependent bool            `json:"fell_dependent"`
	PeakWis       float64         `json:"peak_wis"`
	WisFrom       [5]float64      `json:"wis_from"`
	PeakShips     int             `json:"peak_ships"`
	FirstPlague   Year            `json:"first_plague"`
	Sire          int             `json:"sire"`
	DockRate      map[int]float64 `json:"dock_rate"`
	WantShips     int             `json:"want_ships"`
	GarrisonWant  int             `json:"garrison_want"`
}

// Age is one earlier age of the galaxy.
type Age struct {
	Index  int    `json:"index"`
	Start  Year   `json:"start"`
	End    Year   `json:"end"`
	Ender  string `json:"ender"`
	Elders []int  `json:"elders"`
}

// Elder is a civilisation of an earlier age.
type Elder struct {
	ID       int    `json:"id"`
	Age      int    `json:"age"`
	Portrait string `json:"portrait"`
	Rose     Year   `json:"rose"`
	Fell     Year   `json:"fell"`
	Legacies []int  `json:"legacies"`
}

// Remain is something an age left on the substrate.
type Remain struct {
	ID        int        `json:"id"`
	Age       int        `json:"age"`
	Elder     int        `json:"elder"`
	Maker     int        `json:"maker"`
	Kind      string     `json:"kind"`
	Star      int        `json:"star"`
	Node      string     `json:"node,omitempty"`
	Portrait  string     `json:"portrait"`
	State     string     `json:"state"`
	People    int        `json:"people"`
	Payload   string     `json:"payload"`
	Listeners int        `json:"listeners"`
	Woken     int        `json:"woken"`
	Finder    int        `json:"finder"`
	Level     string     `json:"level,omitempty"`
	Cond      string     `json:"cond"`
	Hardy     float64    `json:"hardy"`
	Source    int        `json:"source"`
	Testament []int      `json:"testament"` // tale ids in tellings.jsonl
	Plague    int        `json:"plague"`
	Wrecks    int        `json:"wrecks"`
	Derelicts int        `json:"derelicts"`
	At        [3]float64 `json:"at"`
	Adrift    bool       `json:"adrift"`
}

// Source is everything with a yield.
type Source struct {
	ID       int         `json:"id"`
	Key      string      `json:"key"`
	Kind     string      `json:"kind"`
	Star     int         `json:"star"`
	Feature  string      `json:"feature,omitempty"`
	Radius   float64     `json:"radius"`
	Yield    flow.Income `json:"yield"`
	Needs    []string    `json:"needs"`
	With     []string    `json:"with"`
	Cradle   bool        `json:"cradle"`
	Rarity   bool        `json:"rarity"`
	Grants   []string    `json:"grants"`
	Levels   [3]float64  `json:"levels"`
	Reach    float64     `json:"reach"`
	Mobile   bool        `json:"mobile"`
	Holder   int         `json:"holder"`
	Carried  int         `json:"carried"`
	Legacy   int         `json:"legacy"`
	Since    Year        `json:"since"`
	Wear     Year        `json:"wear"`
	Form     string      `json:"form,omitempty"`
	Sentient bool        `json:"sentient"`
	Maker    int         `json:"maker"`
	Made     Year        `json:"made"`
	Given    int         `json:"given"`
	Fate     string      `json:"fate,omitempty"`
}

// Trace is something left behind for whoever comes next.
type Trace struct {
	Star    int    `json:"star"`
	Kind    string `json:"kind"`
	Civ     int    `json:"civ"`
	Species int    `json:"species"`
	Year    Year   `json:"year"`
}

// Plague is a sickness of the body or the mind.
type Plague struct {
	ID          int            `json:"id"`
	Kind        string         `json:"kind"` // biological, memetic
	Contagion   float64        `json:"contagion"`
	Lethality   float64        `json:"lethality"`
	Band        int            `json:"band"`
	Engineered  bool           `json:"engineered"`
	Conscious   bool           `json:"conscious"`
	Profile     plague.Profile `json:"profile"`
	Born        Year           `json:"born"`
	FirstHost   int            `json:"first_host"`
	Cause       string         `json:"cause"`
	Hosts       int            `json:"hosts"`
	Peak        int            `json:"peak"`
	Caught      int            `json:"caught"`
	Worlds      int            `json:"worlds"`
	Peoples     int            `json:"peoples"`
	Cults       int            `json:"cults"`
	Cures       int            `json:"cures"`
	Refusals    int            `json:"refusals"`
	Woken       int            `json:"woken"`
	LastHost    Year           `json:"last_host"`
	Extinct     bool           `json:"extinct"`
	Wildfire    bool           `json:"wildfire"`
	Maker       int            `json:"maker"`
	Made        bool           `json:"made"`
	Rider       int            `json:"rider"`
	Transmitter int            `json:"transmitter"`
	Poisonings  int            `json:"poisonings"`
}

// Reservoir is a plague waiting in dead cities.
type Reservoir struct {
	Star   int  `json:"star"`
	Plague int  `json:"plague"`
	Until  Year `json:"until"`
}

// Leader is a named figure a people followed; its name is the names[]
// rows of kind leader.
type Leader struct {
	ID        int        `json:"id"`
	Civ       int        `json:"civ"`
	Rose      Year       `json:"rose"`
	Ended     Year       `json:"ended"`
	End       string     `json:"end"`      // how it ended, a key of leaders.json's ends; "" while it reigns
	Occasion  string     `json:"occasion"` // what made it possible, a key of leaders.json's occasions
	Form      string     `json:"form"`     // what it is, a key of leaders.json's forms
	Stance    string     `json:"stance"`   // the stance it brought, a stance trait's key
	Own       string     `json:"own"`      // the people's own stance at the rising
	Bent      mind.Dials `json:"bent"`     // its term in the dials
	Push      float64    `json:"push"`     // how far it turned the people against its own bent, 0 to 1
	Front     bool       `json:"front"`    // its place: with the greatest fleet at war, else the seat
	Flees     bool       `json:"flees"`
	Fled      bool       `json:"fled"`
	Fleet     int        `json:"fleet"` // the fleet it rides with, -1
	Star      int        `json:"star"`  // where it is, or was at its end
	Until     Year       `json:"until"` // when its span or its line runs out; 0 for never
	Deathless bool       `json:"deathless"`
	Mad       float64    `json:"mad"`       // the slide: a half narrowed, one mad
	Faced     string     `json:"faced"`     // the succession after it: overcome, scarred, declined; "" for none
	Doublings float64    `json:"doublings"` // continuity's term at that succession
	Battles   int        `json:"battles"`   // battles fought at its side
}

// War is one war.
type War struct {
	ID        int        `json:"id"`
	Sides     [2]int     `json:"sides"`
	Began     Year       `json:"began"`
	Ended     Year       `json:"ended"`
	Over      bool       `json:"over"`
	Cause     string     `json:"cause"`
	CauseOf   int        `json:"cause_of"`
	Aim       string     `json:"aim,omitempty"`
	Named     int        `json:"named"`
	Nth       int        `json:"nth"`
	Will      [2]float64 `json:"will"`
	Taken     [2]int     `json:"taken"`
	Glassed   [2]int     `json:"glassed"`
	Lost      [2]int     `json:"lost"`
	Result    string     `json:"result,omitempty"`
	Pact      int        `json:"pact"`
	Principal int        `json:"principal"`
	Hire      int        `json:"hire"`
	Hunt      *Hunt      `json:"hunt,omitempty"`
}

// Hunt is a war's target when it is a region, not a people.
type Hunt struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Radius float64 `json:"radius"`
	Losses int     `json:"losses"`
	Star   int     `json:"star"`
	Since  Year    `json:"since"`
	Empty  Year    `json:"empty"`
}

// Pact is an alliance.
type Pact struct {
	ID      int    `json:"id"`
	Members []int  `json:"members"`
	Kind    string `json:"kind"`
	Target  int    `json:"target"`
	Formed  Year   `json:"formed"`
	Ended   Year   `json:"ended"`
	Over    bool   `json:"over"`
}

// Betrayal is a promise broken, or kept with negative weight.
type Betrayal struct {
	By      int     `json:"by"`
	Against int     `json:"against"`
	Year    Year    `json:"year"`
	Shape   string  `json:"shape"`
	Weight  float64 `json:"weight"`
}

// Fleet is an expedition: ships in flight or at a star.
type Fleet struct {
	ID        int     `json:"id"`
	Owner     int     `json:"owner"`
	Target    int     `json:"target"`
	Kind      string  `json:"kind"`
	Star      int     `json:"star"`
	From      int     `json:"from"`
	Ships     int     `json:"ships"`
	Launched  Year    `json:"launched"`
	Arrive    Year    `json:"arrive"`
	Base      int     `json:"base"`
	Returning bool    `json:"returning"`
	Over      bool    `json:"over"`
	LaidUp    bool    `json:"laid_up"`
	Laid      Year    `json:"laid"`
	Manned    Year    `json:"manned"`
	Held      []int   `json:"held"`
	Seen      []int   `json:"seen"` // the peoples that have seen it
	Battles   int     `json:"battles"`
	Wins      int     `json:"wins"`
	Turned    bool    `json:"turned"`
	Out       Year    `json:"out"`
	Back      int     `json:"back"`
	Sieges    int     `json:"sieges"`
	Drive     float64 `json:"drive"`
	Warning   Year    `json:"warning"`
	Quarry    int     `json:"quarry"`
	Picket    bool    `json:"picket"`
	Contract  int     `json:"contract"`
}

// Contract is a bargain between two peoples.
type Contract struct {
	ID        int     `json:"id"`
	Buyer     int     `json:"buyer"`
	Seller    int     `json:"seller"`
	By        int     `json:"by"`
	Ask       Term    `json:"ask"`
	Pay       Term    `json:"pay"`
	Length    float64 `json:"length"`
	Offered   Year    `json:"offered"`
	Formed    Year    `json:"formed"`
	Until     Year    `json:"until"`
	Ended     Year    `json:"ended"`
	State     string  `json:"state"`
	Broke     int     `json:"broke"`
	Failed    [2]int  `json:"failed"`
	Missed    bool    `json:"missed"`
	AskDone   bool    `json:"ask_done"`
	Taught    bool    `json:"taught"`
	Burned    bool    `json:"burned"`
	Tribute   bool    `json:"tribute"`
	BoughtOff bool    `json:"bought_off"`
}

// Term is one side of a bargain.
type Term struct {
	Kind   string  `json:"kind"`
	Res    string  `json:"res,omitempty"`
	Amount float64 `json:"amount,omitempty"`
	Source int     `json:"source"`
	Node   string  `json:"node,omitempty"`
	Star   int     `json:"star"`
	Work   string  `json:"work,omitempty"`
	Target int     `json:"target"`
	Fleet  int     `json:"fleet"`
}

// Fathoming is one people coming to understand another.
type Fathoming struct {
	Year     Year    `json:"year"`
	Who      int     `json:"who"`
	Whom     int     `json:"whom"`
	How      string  `json:"how"`
	Since    Year    `json:"since"`
	Diff     float64 `json:"diff"`
	Mutual   bool    `json:"mutual"`
	Reversed bool    `json:"reversed"`
}

// Battle is one battle at a world.
type Battle struct {
	Year     Year    `json:"year"`
	Star     int     `json:"star"`
	Attacker int     `json:"attacker"`
	Defender int     `json:"defender"`
	Ships    int     `json:"ships"`
	Held     int     `json:"held"`
	Gap      float64 `json:"gap"`
	Won      bool    `json:"won"`
	Outcome  string  `json:"outcome"`
}

// Meeting is one battle in the dark.
type Meeting struct {
	Year        Year   `json:"year"`
	Quarry      int    `json:"quarry"`
	Interceptor int    `json:"interceptor"`
	Owner       int    `json:"owner"`
	Seer        int    `json:"seer"`
	Ships       int    `json:"ships"`
	Sent        int    `json:"sent"`
	Won         bool   `json:"won"`
	Broken      bool   `json:"broken"`
	Lost        [2]int `json:"lost"`
}

// Sighting is a fleet seen in flight.
type Sighting struct {
	ID        int     `json:"id"`
	Fleet     int     `json:"fleet"`
	Owner     int     `json:"owner"`
	Seer      int     `json:"seer"`
	Kind      string  `json:"kind"`
	From      int     `json:"from"`
	Star      int     `json:"star"`
	Launched  Year    `json:"launched"`
	Arrive    Year    `json:"arrive"`
	Leg       Year    `json:"leg"`
	Year      Year    `json:"year"`
	Ships     int     `json:"ships"`
	Mil       float64 `json:"mil"`
	Speed     float64 `json:"speed"`
	Eye       string  `json:"eye"`
	EyeWork   string  `json:"eye_work,omitempty"`
	EyeStar   int     `json:"eye_star"`
	Feasible  bool    `json:"feasible"`
	Intercept int     `json:"intercept"`
	Offered   bool    `json:"offered"`
}

// Object is a thing that can be named: a kind and the id the simulation
// gave it.
type Object struct {
	Kind string `json:"kind"` // civ, star, plague, war, elder, makers, source, word, title, species
	ID   int    `json:"id"`
}

// Human is the namer of a designation row.
const Human = -2

// Name is one name for one object by one culture: a row of names[].
type Name struct {
	Object Object `json:"object"`
	By     int    `json:"by"`
	Name   string `json:"name"`
	Mode   string `json:"mode"` // transcribed, translated, designation, proper, adopted
	Tone   string `json:"tone"` // self, stranger, friend, enemy, monster, sky, none
	Coined Year   `json:"coined"`
	From   int    `json:"from"`
	Gloss  string `json:"gloss,omitempty"`
	Recipe Recipe `json:"recipe,omitzero"`
}

// Recipe is how a translated name was made.
type Recipe struct {
	Entry   string `json:"entry"`
	Pattern string `json:"pattern"`
	Slots   string `json:"slots,omitempty"` // hole=value pairs, joined with "; "
}

// Voice is how a culture names: none, transcribed or translated.
type Voice string

const (
	None        Voice = "none"
	Transcribed Voice = "transcribed"
	Translated  Voice = "translated"
)
