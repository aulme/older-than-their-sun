package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/tech"
)

// Reservations: a fleet, a scout, surveyors or a colony ship in flight is
// a use of the means for as long as it is out. Nothing is spent, so nothing
// is refunded: the reservation is released when the thing ends, is lost or
// comes home. A launch needs the reservation covered from the spare this
// tick, or it does not go. A structure reserves its upkeep and is built
// only when the spare covers it twice over.

// per-level and per-ship reservations
var (
	fleetNeed = flow.Income{flow.E: 1, flow.M: 1} // per level of a fleet, a scout, surveyors
	grownNeed = flow.Income{flow.O: 1, flow.E: 1} // the same for a people whose ships are bred
	shipNeed  = flow.Income{flow.E: 2, flow.M: 2} // a colony ship in flight
	sporeNeed = flow.Income{flow.O: 1}            // a swarm's seed-cloud
)

// fleetReservation is what a fleet of a strength reserves for a people:
// half the metal when it sails from a shipyard.
func (w *World) fleetReservation(c *Civ, mil float64, from int) flow.Income {
	need := fleetNeed
	if c.Known["living_ships"] {
		need = grownNeed
	}
	need = need.Scale(mil)
	if from >= 0 && c.hasWork("shipyard", from) {
		need[flow.M] /= 2
	}
	return w.bend(c, need)
}

// shipReservation is what a colony ship reserves.
func (w *World) shipReservation(c *Civ) flow.Income {
	need := shipNeed
	if c.Known["seed_clouds"] {
		need = sporeNeed
	}
	return w.bend(c, need)
}

// bend applies the profile to a need: a machine pays organic matter in energy.
func (w *World) bend(c *Civ, need flow.Income) flow.Income {
	if c.Species.Profile().OrganicAsEnergy {
		need[flow.E] += need[flow.O]
		need[flow.O] = 0
	}
	return need
}

// reservations lists a people's fleets and ships as uses, in flight.
func (w *World) reservations(c *Civ) []flow.Use {
	var out []flow.Use
	for _, x := range w.Expeditions {
		if x.Over || x.Owner != c.ID || x.Kind == Roam {
			continue
		}
		cat, what := flow.Arms, "the fleet"
		switch x.Kind {
		case Survey:
			cat, what = flow.Road, "the surveyors"
		case Scout:
			what = "the scout"
		}
		out = append(out, flow.Use{Key: "fleet:" + itoa(x.ID), Name: what + " bound for " + w.star(x.Star), Cat: cat, Era: 3, Need: w.fleetReservation(c, x.Mil, x.From), Flight: true})
	}
	for i, v := range c.Voyages {
		out = append(out, flow.Use{Key: "ship:" + itoa(i), Name: "the colony ship bound for " + w.star(v.Target), Cat: flow.Road, Era: 3, Need: w.shipReservation(c), Flight: true})
	}
	return out
}

// works lists a people's structures as uses.
func (w *World) works(c *Civ) []flow.Use {
	var out []flow.Use
	for _, wk := range c.Works {
		st := tech.Structures[wk.Key]
		if st == nil {
			continue
		}
		out = append(out, flow.Use{Key: wk.key(), Name: "the " + st.Name + " at " + w.star(wk.Star), Cat: st.Cat, Era: tech.Get(st.Node).Era, Need: w.bend(c, st.Upkeep)})
	}
	return out
}

// afford says whether the spare this tick covers a need, less what was
// already reserved this tick.
func (w *World) afford(c *Civ, need flow.Income) bool {
	return c.Surplus.Less(c.Reserved).Covers(need)
}

// reserve takes a need from this tick's spare.
func (w *World) reserve(c *Civ, need flow.Income) {
	c.Reserved.Add(need)
}
