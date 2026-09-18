package mind

import (
	"fmt"

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
