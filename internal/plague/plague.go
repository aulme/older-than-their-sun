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

	"worldgen/internal/names"
)

// Kind is where a plague lives.
type Kind uint8

const (
	Biological Kind = iota // in bodies: moves with goods and people
	Memetic                // in minds: moves with messages
)

func (k Kind) String() string { return [...]string{"biological", "memetic"}[k] }

// Plague is one sickness: a name, a kind, a contagion and a lethality.
// Contagion is the chance a channel carries it and how hard it is to
// shake; lethality is the damage once it is in. The two are drawn apart,
// so they work against each other by construction: what kills its hosts
// in three ticks has three ticks to cross a trade link, and what kills
// nobody has the age.
type Plague struct {
	Name       string
	Kind       Kind
	Contagion  float64 // c, in (0, 1]
	Lethality  float64 // l, in (0, 1]
	Named      bool    // named for its first host, which that host remembers
	Conscious  bool    // a parasite people waiting to happen; read from step 14
	Engineered bool    // made as a weapon; step 14
	Band       int     // the people it was tailored to catch, -1 for any body; step 14
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
}

// Default is today's numbers.
func Default() Tuning {
	return Tuning{
		BaseBio: 0.002, BaseMeme: 0.0005, PerWorlds: 8, Rung: 3, Halving: 2, Siege: 2, SiegeCap: 4,
		Shed: 2, Dark: 3, DarkYears: 100_000, Taken: 2, TakenYears: 10_000, Faith: 2, Beacon: 2,
		Hygiene: 1.5, Difficulty: 3, PerContagion: 6, Spread: 1.5, RungCure: 0.5, Dirt: 1, PartnerCured: 2, Contain: -2,
		Power: 3, ContainedToll: 0.5, Conscious: 0.05, CreedTicks: 20, Wildfire: 10, ExtinctYears: 100_000, ReservoirMyr: 2, WallsShare: 0.5,
		CultChance: 0.4, Weakened: 0.2,
	}
}

// New draws a plague of a kind: contagion and lethality uniform in
// (0, 1] and independent, a name, and whether it thinks. host is what the
// neighbours might name it for, "" for nothing.
func New(r *rand.Rand, kind Kind, host string, t *Tuning) Plague {
	p := Plague{Kind: kind, Contagion: 1 - r.Float64(), Lethality: 1 - r.Float64(), Band: -1}
	p.Name, p.Named = names.Plague(r, kind == Memetic, host)
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
	if f.Mul > 0 {
		p *= f.Mul
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
