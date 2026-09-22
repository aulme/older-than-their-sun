package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/tech"
)

// Reservations: every ship in being is a use of the means, at a base or
// in flight, and so is a colony ship in flight and a dock at work. Nothing
// is spent, so nothing is refunded: the reservation is released when the
// thing ends, is lost or comes home. A structure reserves its upkeep and
// is built only when the spare covers it twice over.

// per-ship reservations for the colony ships; a fleet's keep is in ships.go
var (
	shipNeed  = flow.Income{flow.E: 2, flow.M: 2} // a colony ship in flight
	sporeNeed = flow.Income{flow.O: 1}            // a swarm's seed-cloud
)

// fleetReservation is what a fleet of so many ships reserves: the keep of
// each.
func (w *World) fleetReservation(c *Civ, ships int) flow.Income {
	return w.keepOf(c).Scale(float64(ships))
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

// reservations lists a people's fleets, colony ships and docks as uses:
// every fleet's keep, as arms, and the surveyors on the road. A fleet's
// keep is fed as soon as the fields are, at a base as in flight: a people
// lets its ships rot after everything but its fields, since the first
// batch with the keep last in the peacetime order had research grow into
// what the guard would have taken and every fleet built, laid up and
// rotted in turn.
func (w *World) reservations(c *Civ) []flow.Use {
	var out []flow.Use
	for _, x := range w.fleetsOf(c) {
		if x.Ships == 0 {
			continue
		}
		cat := flow.Arms
		if x.Kind == Survey {
			cat = flow.Road
		}
		out = append(out, flow.Use{Key: "fleet:" + itoa(x.ID), Cat: cat, Era: 3, Need: w.fleetReservation(c, x.Ships), Flight: true})
	}
	for i := range c.Voyages {
		out = append(out, flow.Use{Key: "ship:" + itoa(i), Cat: flow.Road, Era: 3, Need: w.shipReservation(c), Flight: true})
	}
	return append(out, w.dockUses(c)...)
}

// works lists a people's structures as uses.
func (w *World) works(c *Civ) []flow.Use {
	var out []flow.Use
	for _, wk := range c.Works {
		st := tech.Structures[wk.Key]
		if st == nil {
			continue
		}
		out = append(out, flow.Use{Key: wk.key(), Cat: st.Cat, Era: tech.Get(st.Node).Era, Need: w.bend(c, st.Upkeep)})
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
