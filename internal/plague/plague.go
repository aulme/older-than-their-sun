// Package plague is the shape of a sickness and the arithmetic of its
// life: born of dirt at a rate medicine divides, caught along a channel
// at a chance hygiene divides, fought each tick in a contest of the
// people's fitness against its contagion, paid for in worlds by its
// lethality. Nothing here reads the world; history hands it what it
// needs and rolls the chances it returns.
package plague

import (
	"math"
	"math/rand/v2"
)

// Kind is where a plague lives.
type Kind uint8

const (
	Biological Kind = iota // in bodies: moves with goods and people
	Memetic                // in minds: moves with messages
)

func (k Kind) String() string { return [...]string{"biological", "memetic"}[k] }

// Plague is one sickness: a kind, a contagion and a lethality. Its name
// is a row of the names pass, keyed by the history's id for it.
// Contagion is the chance a channel carries it and how hard it is to
// shake; lethality is the damage once it is in. The two are drawn apart,
// so they work against each other by construction: what kills its hosts
// in three ticks has three ticks to cross a trade link, and what kills
// nobody has the age.
type Plague struct {
	Kind       Kind
	Contagion  float64 // c, in (0, 1]
	Lethality  float64 // l, in (0, 1]
	Conscious  bool    // a parasite people waiting to happen: it wakes when its first host's home goes over
	Engineered bool    // made as a weapon
	Band       int     // the species it was tailored to catch, with its kin; -1 for any body
	Profile    Profile // what it is, as keys; see profile.go
}

// Tuning is every number the plagues use.
type Tuning struct {
	BaseBio       float64 // births per kyr for a cradle with nothing
	BaseMeme      float64 // the same for an idea; needs writing
	PerWorlds     float64 // the base grows by one for this many worlds held
	Rung          float64 // births divided by this per working rung of the ladder
	Halving       float64 // births divided by this per working node that halves only: sanitation, closed ecologies
	Siege         float64 // births multiplied by this per world under siege, to SiegeCap
	SiegeCap      float64
	Shed          float64 // any use shed this tick
	Dark          float64 // a dark age within DarkYears
	DarkYears     float64
	Taken         float64 // a world taken within TakenYears, either side; biological only
	TakenYears    float64
	Faith         float64 // the Wars of Faith faced, ever; memetic only
	Beacon        float64 // a beacon heard, ever; memetic only
	Hygiene       float64 // the catching chance divided by this per working rung
	Difficulty    float64 // the cure contest's flat difficulty
	PerContagion  float64 // and what contagion adds to it
	Spread        float64 // the roll's spread
	RungCure      float64 // what a working rung adds to the cure roll
	Dirt          float64 // what each of siege, shed and a recent dark age adds to the difficulty
	PartnerCured  float64 // what a fathomed partner's cure adds, biological only
	Contain       float64 // the margin below which the plague rages; above it, to nought, it is contained
	Power         float64 // a world is lost each tick with chance l to this power
	ContainedToll float64 // the toll's share while contained
	Conscious     float64 // the chance a plague thinks
	CreedTicks    int     // ticks contained in a row before the quarantine creed
	Wildfire      int     // hosts at once that make a wildfire
	ExtinctYears  float64 // years with no host before the plague is extinct
	ReservoirMyr  float64 // a reservoir lasts c times this many million years
	WallsShare    float64 // walls carry a memetic plague at c times this
	CultChance    float64 // a world gone over that declares itself a people
	Weakened      float64 // the lethality from which a raging host reads as weakened to its neighbours
	BornRider     float64 // the share of cradles born with a rider already in them
	Revolt        float64 // what riding adds to the cure contest's difficulty: a rider is not thrown off, it is cured
	Detect        float64 // the chance an attempt that failed is seen, per working rung of the ladder
	DetectCap     float64
	DetectCensor  float64 // what censorship adds against a memetic attempt
	Leak          float64 // per kyr, a held weapon gets out, times the dirt
	LeakShed      float64 // times this when the programme is unpaid
	WorldsAim     float64 // the lethality, as a share of the band, of a plague made to soften a rival for the taking
	GoneAim       float64 // the contagion, as a share of the band, of one made to end it
}

// Default is today's numbers.
func Default() Tuning {
	return Tuning{
		BaseBio: 0.002, BaseMeme: 0.0005, PerWorlds: 8, Rung: 3, Halving: 2, Siege: 2, SiegeCap: 4,
		Shed: 2, Dark: 3, DarkYears: 100_000, Taken: 2, TakenYears: 10_000, Faith: 2, Beacon: 2,
		Hygiene: 1.5, Difficulty: 3, PerContagion: 6, Spread: 1.5, RungCure: 0.5, Dirt: 1, PartnerCured: 2, Contain: -2,
		Power: 3, ContainedToll: 0.5, Conscious: 0.02, CreedTicks: 20, Wildfire: 10, ExtinctYears: 100_000, ReservoirMyr: 2, WallsShare: 0.5,
		CultChance: 0.4, Weakened: 0.2,
		BornRider: 0.01, Revolt: 4, Detect: 0.2, DetectCap: 0.9, DetectCensor: 0.3, Leak: 0.0002, LeakShed: 10, WorldsAim: 0.25, GoneAim: 0.75,
	}
}

// New draws a plague of a kind: contagion and lethality uniform in
// (0, 1] and independent, and whether it thinks.
func New(r *rand.Rand, kind Kind, t *Tuning) Plague {
	p := Plague{Kind: kind, Contagion: 1 - r.Float64(), Lethality: 1 - r.Float64(), Band: -1}
	p.Conscious = r.Float64() < t.Conscious
	return p
}

// Factors is the sanitation of one people this tick, as birth reads it.
type Factors struct {
	Worlds   int
	Rungs    int     // working rungs of the ladder for the kind
	Halvings int     // working nodes that halve births only
	Immune   bool    // the top of the ladder working, or the kind cannot be borne
	Sieged   int     // worlds with a campaign fleet in the sky
	Shed     bool    // any use shed this tick
	Dark     bool    // a dark age within the window
	Taken    bool    // a world taken within the window, either side
	Faith    bool    // the Wars of Faith faced, ever
	Beacon   bool    // a beacon heard, ever
	Mul      float64 // the profile's multiplier on bearing the kind; 0 means 1
}

// BirthChance is the chance per thousand years a people bears a plague
// of the kind: a base by its worlds, divided by its medicine, multiplied
// by its dirt.
func BirthChance(kind Kind, f Factors, t *Tuning) float64 {
	if f.Immune {
		return 0
	}
	base := t.BaseBio
	if kind == Memetic {
		base = t.BaseMeme
	}
	p := base * (1 + float64(f.Worlds)/t.PerWorlds)
	p /= math.Pow(t.Rung, float64(f.Rungs))
	if kind == Biological {
		p /= math.Pow(t.Halving, float64(f.Halvings))
	}
	p *= Dirt(kind, f, t)
	if f.Mul > 0 {
		p *= f.Mul
	}
	return p
}

// Dirt is the birth table's multipliers alone: what a siege, a shed use,
// a dark age, a taking, the Wars of Faith and a beacon do, by kind. One
// with nothing.
func Dirt(kind Kind, f Factors, t *Tuning) float64 {
	p := 1.0
	if f.Sieged > 0 {
		p *= min(t.SiegeCap, math.Pow(t.Siege, float64(f.Sieged)))
	}
	if f.Shed {
		p *= t.Shed
	}
	if f.Dark {
		p *= t.Dark
	}
	if kind == Biological && f.Taken {
		p *= t.Taken
	}
	if kind == Memetic {
		if f.Faith {
			p *= t.Faith
		}
		if f.Beacon {
			p *= t.Beacon
		}
	}
	return p
}

// Aim is what a maker wants of a plague it makes.
type Aim uint8

const (
	AimWorlds Aim = iota // the rival's worlds: a drag that softens it for the taking
	AimGone              // the rival gone: the top of the band in lethality
	AimAll               // the different, wholly: the top of both
)

func (a Aim) String() string { return [...]string{"the worlds", "the end of them", "everything"}[a] }

// Shape is the contagion and lethality a maker picks for an aim within
// the band its craft allows: high contagion and low lethality to soften,
// the top in lethality to end, both to the top for hate.
func Shape(aim Aim, band float64, t *Tuning) (c, l float64) {
	switch aim {
	case AimWorlds:
		return band, band * t.WorldsAim
	case AimGone:
		return band * t.GoneAim, band
	}
	return band, band
}

// Make is a plague made to a shape: the name is the maker's to give.
func Make(kind Kind, c, l float64, conscious bool) Plague {
	return Plague{Kind: kind, Contagion: c, Lethality: l, Engineered: true, Conscious: conscious, Band: -1}
}

// Loose is a plague that got out of the vial before it was shaped: drawn
// within the band, and it may think.
func Loose(r *rand.Rand, kind Kind, band float64, t *Tuning) Plague {
	p := New(r, kind, t)
	p.Contagion, p.Lethality, p.Engineered = p.Contagion*band, p.Lethality*band, true
	return p
}

// DetectChance is the chance a target sees an attempt that failed: a
// share per working rung of the ladder, capped, and censorship against a
// memetic one.
func DetectChance(rungs int, censor bool, t *Tuning) float64 {
	p := t.Detect * float64(rungs)
	if censor {
		p += t.DetectCensor
	}
	return min(t.DetectCap, p)
}

// LeakChance is the chance per kyr a held weapon gets out: the base by the
// maker's dirt, ten times if the programme is unpaid.
func LeakChance(dirt float64, shed bool, t *Tuning) float64 {
	p := t.Leak * dirt
	if shed {
		p *= t.LeakShed
	}
	return p
}

// Hygiene is what a people's working rungs divide the catching chance by.
func Hygiene(rungs int, t *Tuning) float64 { return 1 / math.Pow(t.Hygiene, float64(rungs)) }

// CatchChance is the chance per tick a channel carries the plague to a
// people: its contagion, the channel's weight, the people's hygiene and
// its nature's multiplier (0 means 1).
func CatchChance(c, weight, hygiene, mul float64) float64 {
	if mul == 0 {
		mul = 1
	}
	return min(1, c*weight*hygiene*mul)
}

// Outcome is what a tick's cure roll came to.
type Outcome uint8

const (
	Raging    Outcome = iota // the full toll, every channel
	Contained                // half the toll, the doors closed
	Cured                    // rid of it, and immune
)

func (o Outcome) String() string { return [...]string{"raging", "contained", "cured"}[o] }

// CureMargin is the cure contest: the people's level in the kind, its
// working rungs, a roll, and what a cured partner's goods bring, against
// the plague's contagion and the people's dirt.
func CureMargin(level, ladder, roll, c, dirt, partnerCured float64, t *Tuning) float64 {
	return level + ladder + roll - (t.Difficulty + t.PerContagion*c) - dirt + partnerCured
}

// Band reads the margin: cured at nought or better, contained down to the
// containment line, raging below it.
func Band(margin float64, t *Tuning) Outcome {
	switch {
	case margin >= 0:
		return Cured
	case margin >= t.Contain:
		return Contained
	}
	return Raging
}

// TollChance is the chance per tick a held world is lost to the plague:
// its lethality to the power, by the people's frailty (0 means 1), halved
// while contained.
func TollChance(l, frail float64, contained bool, t *Tuning) float64 {
	if frail == 0 {
		frail = 1
	}
	p := math.Pow(min(1, l*frail), t.Power)
	if contained {
		p *= t.ContainedToll
	}
	return p
}

// Toll is what a plague does to a people each tick it has it, in levels
// and morale, and the multipliers on its research and its income.
type Toll struct {
	Sur, Soc float64 // taken off the levels
	Morale   float64 // taken off the morale each tick
	Research float64 // multiplier
	Organic  float64 // multiplier on organic matter, biological
	Energy   float64 // multiplier on energy, memetic in a machine-born people
}

// TollOf is the toll of one plague on a people, halved while contained.
func TollOf(p Plague, contained, machine bool) Toll {
	l := p.Lethality
	t := Toll{Morale: l, Research: 1 - l/2, Organic: 1, Energy: 1}
	if p.Kind == Biological {
		t.Sur, t.Soc, t.Organic = 2*l, 2*l, 1-l
	} else {
		t.Soc = 3 * l
		if machine {
			t.Energy = 1 - l/2
		}
	}
	if contained {
		t.Sur, t.Soc, t.Morale = t.Sur/2, t.Soc/2, t.Morale/2
		t.Research, t.Organic, t.Energy = (1+t.Research)/2, (1+t.Organic)/2, (1+t.Energy)/2
	}
	return t
}
