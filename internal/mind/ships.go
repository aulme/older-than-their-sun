package mind

import (
	"fmt"
	"math"

	"worldgen/internal/battle"
)

// Ships: a level is how good a people's arms are, ships are how many it
// has. The decisions weigh both in one unit, levels: a count of ships is
// worth its logarithm at the battle base, so four ships at level six
// against two at level nine is even. The want is how many ships a people
// builds toward, the sum of what its policies ask for.

// ShipLevels is what a count of ships is worth in levels: the log at the
// battle base. Nothing, or less than one ship, is worth one ship, since
// the world itself fights as one until guns land.
func ShipLevels(ships float64) float64 {
	return battle.Levels(max(ships, 1))
}

// WantInput is what asks for ships.
type WantInput struct {
	Garrisons int     // the garrison wants of every holding, summed; zero until garrisons land
	Campaign  int     // the need of the campaign the council last sized and could not man
	Explorers int     // one per exploring policy that wants a ship out
	Fear      float64 // the floor rises with fear
}

// Want is how many ships a people builds toward, and why.
type Want struct {
	Ships     int
	Garrisons int
	Campaign  int
	Explorers int
	Floor     int
}

// Why says the want in a line.
func (w Want) Why() string {
	return fmt.Sprintf("wants %d ships: %d for the garrisons, %d for the campaign, %d for the explorers, a floor of %d", w.Ships, w.Garrisons, w.Campaign, w.Explorers, w.Floor)
}

// WantShips sums the wants: the garrisons, the campaign, the explorers,
// and a floor of one plus fear rounded.
func WantShips(in WantInput, t *Tuning) Want {
	p := &t.Want
	w := Want{Garrisons: in.Garrisons, Campaign: in.Campaign, Explorers: in.Explorers, Floor: p.Floor + int(p.FearWeight*in.Fear+0.5)}
	w.Ships = w.Garrisons + w.Campaign + w.Explorers + w.Floor
	return w
}

// Holding is one of a people's stars as the garrison policy sees it: what
// threatens it, what stands over it, what is there and what is coming.
type Holding struct {
	Star   int
	Home   bool
	Threat float64 // the strongest fleet a hostile or unknown neighbour whose reach covers it is believed to have, in the people's own ships
	Guns   int     // guns standing over it
	Ships  int     // the guard there, manned
	Coming int     // ships already on their way to it
}

// GarrisonInput is every holding, and the fear that weighs the threats.
type GarrisonInput struct {
	Holdings []Holding
	Fear     float64
}

// Garrison is the want of every holding and the one move the council
// orders: from the holding most over its want to the one most under it.
type Garrison struct {
	Wants    []int // by holding
	Total    int   // the wants summed: what the docks are asked for
	From, To int   // indexes into the holdings, or -1 for no move
	Ships    int   // the ships moved
}

// Why says the move.
func (g Garrison) Why() string {
	if g.From < 0 {
		return fmt.Sprintf("the garrisons want %d ships and nothing moves", g.Total)
	}
	return fmt.Sprintf("the garrisons want %d ships; %d move from holding %d to holding %d", g.Total, g.Ships, g.From, g.To)
}

// Garrisons places the ships. The home wants the strongest threat's ships
// times a base plus fear, and at least a floor; a world in an enemy's
// reach wants a share of that; guns over a world stand in for ships;
// everything else wants none. One move per council: the largest excess
// to the largest shortfall, what is already coming counted.
func Garrisons(in GarrisonInput, t *Tuning) Garrison {
	p := &t.Garrison
	g := Garrison{From: -1, To: -1, Wants: make([]int, len(in.Holdings))}
	for i, h := range in.Holdings {
		want := h.Threat * (p.HomeBase + in.Fear)
		if !h.Home {
			want *= p.ColonyShare
		}
		n := int(math.Ceil(want-1e-9)) - h.Guns
		if h.Home {
			n = max(n, p.HomeMin)
		}
		g.Wants[i] = max(n, 0)
		g.Total += g.Wants[i]
	}
	excess, short := 0, 0
	for i, h := range in.Holdings {
		if e := h.Ships - g.Wants[i]; e > excess {
			excess, g.From = e, i
		}
		if s := g.Wants[i] - h.Ships - h.Coming; s > short {
			short, g.To = s, i
		}
	}
	if g.From < 0 || g.To < 0 || g.From == g.To {
		g.From, g.To = -1, -1
		return g
	}
	g.Ships = min(excess, short)
	return g
}

// MusterInput is a campaign gathering at a holding.
type MusterInput struct {
	Have   int     // ships in the guard at the muster star, manned
	Need   int     // the campaign's ships
	Coming int     // ships on their way to it
	Age    float64 // years since the order
	Holds  bool    // the sizing still says send
}

// MusterChoice is what becomes of it this tick.
type MusterChoice struct {
	Launch    bool
	StandDown bool
	Reason    string
}

// Why says the choice.
func (m MusterChoice) Why() string {
	switch {
	case m.Launch:
		return "the fleet is gathered and sails"
	case m.StandDown:
		return "the muster stands down: " + m.Reason
	}
	return "the muster waits: " + m.Reason
}

// Muster launches when the guard has the ships, stands down when the
// odds no longer hold, when nothing more is coming, or when it has waited
// too long, and waits otherwise.
func Muster(in MusterInput, t *Tuning) MusterChoice {
	switch {
	case !in.Holds:
		return MusterChoice{StandDown: true, Reason: "the odds no longer hold"}
	case in.Have >= in.Need:
		return MusterChoice{Launch: true}
	case in.Age > t.Campaign.MusterMax:
		return MusterChoice{StandDown: true, Reason: "it waited too long"}
	case in.Coming == 0:
		return MusterChoice{StandDown: true, Reason: "no ships are coming"}
	}
	return MusterChoice{Reason: fmt.Sprintf("%d of %d, %d on the way", in.Have, in.Need, in.Coming)}
}
