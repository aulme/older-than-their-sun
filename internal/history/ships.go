package history

import (
	"worldgen/internal/battle"
	"worldgen/internal/flow"
	"worldgen/internal/mind"
)

// Ships: a level is how good a people's arms are, ships are how many it
// has. Every ship is in a fleet (an Expedition); a Guard is the ships at
// one of the people's stars. Ships are built at docks over time at the
// rate the spare affords, a ship costing four thousand years of its own
// keep; a ship in being reserves its keep each tick, and a fleet the flow
// cannot keep is laid up, fights nothing and rots. The count of ships is
// bounded by what a people can build, feed and crew and by nothing else;
// what it builds toward is its want, which the mind sums.

// Keep per ship by kind: the crew in organic matter, the repair and the
// refit in metal and energy. A machine-born people pays its crew in energy
// through bend; a people whose ships are grown pays flesh alone.
var (
	shipKeep  = flow.Income{flow.O: 1, flow.E: 1, flow.M: 1}
	grownKeep = flow.Income{flow.O: 3}
)

// dockWork is how many times a ship's keep a dock draws at full rate: a
// ship costs this many thousand years of its keep to build.
const dockWork = 4

// dockRate is what a dock at full work makes per thousand years.
const dockRate = 1

// hordeDock is a horde's fleet at a base as a dock, as a share of one.
const hordeDock = 0.25

// rotRate is the share of a laid-up fleet lost per thousand years.
const rotRate = 0.1

// wantStands is how long a campaign's unmet need keeps asking for ships.
const wantStands Year = 20_000

// laidLineAfter is how many ticks a fleet must have been manned before its
// laying up is a line, and laid up before its manning again is: a fleet
// that flaps between the two each tick is not news.
const laidLineAfter = 5

// keepOf is what one of a people's ships reserves each tick.
func (w *World) keepOf(c *Civ) flow.Income {
	if c.Known["living_ships"] {
		return grownKeep
	}
	return w.bend(c, shipKeep)
}

// quality is what a level is worth in battle: the base to the level, with
// the miracles that fight as levels on top.
func (w *World) quality(c *Civ) float64 {
	return battle.Quality(c.Mil + c.warBonus())
}

// levelOf is what a people brings in levels: its level, its miracles and
// its manned ships as levels.
func (w *World) levelOf(c *Civ) float64 {
	return c.Mil + c.warBonus() + mind.ShipLevels(float64(w.standing(c)))
}

// fleetsOf is every fleet a people has, of any kind, not over.
func (w *World) fleetsOf(c *Civ) []*Expedition {
	var out []*Expedition
	for _, x := range w.Expeditions {
		if !x.Over && x.Owner == c.ID {
			out = append(out, x)
		}
	}
	return out
}

// ships is every ship a people has in being, in every fleet, laid up or not.
func (w *World) ships(c *Civ) int {
	n := 0
	for _, x := range w.fleetsOf(c) {
		n += x.Ships
	}
	return n
}

// standing is the ships a people has manned at its bases, as guards or a
// horde's fleets: what could sail, and what the appraisal counts.
func (w *World) standing(c *Civ) int {
	n := 0
	for _, x := range w.fleetsOf(c) {
		if x.atBase() && !x.LaidUp {
			n += x.Ships
		}
	}
	return n
}

// atBase says whether a fleet is at a star as a guard or a horde's fleet:
// not in flight, and not out on a campaign or an errand.
func (x *Expedition) atBase() bool {
	return x.Base >= 0 && (x.Kind == Guard || x.Kind == Roam)
}

// guardAt is a people's guard at a star, or nil.
func (w *World) guardAt(c *Civ, star int) *Expedition {
	for _, x := range w.fleetsOf(c) {
		if x.Base == star && (x.Kind == Guard || x.Kind == Roam) {
			return x
		}
	}
	return nil
}

// addGuard puts ships into the guard at a star, raising one if there is
// none. A horde's ships join its fleet there.
func (w *World) addGuard(c *Civ, star, n int) *Expedition {
	if g := w.guardAt(c, star); g != nil {
		g.Ships += n
		return g
	}
	kind := Guard
	if c.Aloft {
		kind = Roam
	}
	g := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: -1, Kind: kind, Star: star, From: star, Ships: n, Back: -1,
		Launched: w.Now, Out: w.Now, Arrive: w.Now, Base: star, Fed: w.Now, Manned: w.Now, Seen: map[int]bool{}}
	w.Expeditions = append(w.Expeditions, g)
	return g
}

// firstGuard is the ship every starfaring people begins with: a people
// that reached the stars built at least one.
func (w *World) firstGuard(c *Civ) {
	if c.Aloft || w.ships(c) > 0 {
		return
	}
	w.addGuard(c, c.Home, 1)
}

// mergeInto lands a fleet's ships into the guard at a star and ends it.
func (w *World) mergeInto(x *Expedition, star int) {
	c := w.Civs[x.Owner]
	w.land(x, star)
	if x.Ships > 0 {
		w.addGuard(c, star, x.Ships)
	}
	x.Over = true
}

// dock is where ships are made: a star and a rate in ships per thousand
// years at full work.
type dock struct {
	star int
	rate float64
}

// docks are a people's: its home, each shipyard, and for a horde each fleet
// at a base at a quarter; a vassal's work at half, a slave's not at all.
func (w *World) docks(c *Civ) []dock {
	if c.Starfaring == 0 || (!c.Free() && !c.Vassal) {
		return nil
	}
	mul := 1.0
	if c.Vassal {
		mul = 0.5
	}
	var out []dock
	add := func(star int, rate float64) {
		for i := range out {
			if out[i].star == star {
				out[i].rate += rate
				return
			}
		}
		out = append(out, dock{star, rate})
	}
	if c.Aloft {
		for _, x := range w.fleetsOf(c) {
			if x.Kind == Roam && x.Base >= 0 && !x.LaidUp {
				add(x.Base, hordeDock*dockRate*mul)
			}
		}
		return out
	}
	add(c.Home, dockRate*mul)
	for _, wk := range c.Works {
		if wk.Key == "shipyard" && !wk.Dark {
			add(wk.Star, dockRate*mul)
		}
	}
	return out
}

// dockUses lists the docks at work as uses: the draw of last tick's rate,
// the works in peace and arms at war.
func (w *World) dockUses(c *Civ) []flow.Use {
	var out []flow.Use
	cat := flow.Works
	if len(c.Wars) > 0 {
		cat = flow.Arms
	}
	for _, star := range sortedInts(c.DockRate) {
		rate := c.DockRate[star]
		if rate <= 0 {
			continue
		}
		name := "the yards at " + w.star(star)
		if c.Known["living_ships"] {
			name = "the breeding grounds at " + w.star(star)
		}
		out = append(out, flow.Use{Key: "dock:" + itoa(star), Name: name, Cat: cat, Era: 3, Need: w.keepOf(c).Scale(dockWork * rate)})
	}
	return out
}

// want is how many ships a people builds toward: the mind's sum of the
// garrisons' wants, the campaign it could not man, the explorers it keeps
// out, and a floor by fear.
func (w *World) want(c *Civ) mind.Want {
	in := mind.WantInput{Fear: c.Dials.Fear, Explorers: c.SurveyWant, Garrisons: c.GarrisonWant}
	if len(c.Met) > 0 {
		in.Explorers++ // a scout
	}
	if c.WantShips > 0 && w.Now-c.WantSince <= wantStands {
		in.Campaign = c.WantShips
	}
	return mind.WantShips(in, w.Cfg.Tuning)
}

// shipwright is the civ step after the flows: the keep, then the docks.
// A fleet the direction did not feed is laid up, and a laid-up fleet rots;
// a fleet fed again is manned at the people's level today. Each dock
// builds at the rate it was fed, and the rate for next tick is what the
// spare affords while the ships are under the want; a people with a
// fleet laid up builds nothing, since it could not keep what it built.
func (w *World) shipwright(c *Civ) {
	laid := w.keep(c)
	if c.Starfaring == 0 {
		return
	}
	c.Tally.StarTicks++
	want := w.want(c)
	have := w.ships(c)
	if have >= want.Ships {
		c.Tally.AtWant++
	}
	keep := w.keepOf(c)
	rates := map[int]float64{}
	spare := c.Surplus.Less(c.Reserved)
	for _, d := range w.docks(c) {
		if laid {
			break
		}
		fed := 0.0
		if r, ok := c.DockRate[d.star]; ok && r > 0 && c.working("dock:"+itoa(d.star)) {
			fed = r
		}
		if fed > 0 {
			c.Tally.Building += dockWork * fed * w.dt
			if n := w.count(fed); n > 0 {
				w.addGuard(c, d.star, n)
				have += n
				c.Tally.Built += n
				if !c.firstShip {
					c.firstShip = true
					w.log("The yards at %s launch their first ship for the %s.", w.star(d.star), c.Name)
				}
			}
		}
		if have >= want.Ships {
			continue // the want is met: the dock stands idle
		}
		// what more the spare would feed, on top of what was fed
		extra := d.rate - fed
		for k := range keep {
			if keep[k] > 0 {
				extra = min(extra, spare[k]/(dockWork*keep[k]))
			}
		}
		extra = max(extra, 0)
		rate := fed + extra
		if rate > 0 {
			rates[d.star] = rate
			w.reserve(c, keep.Scale(dockWork*extra))
			spare = c.Surplus.Less(c.Reserved)
		}
	}
	c.DockRate = rates
	if n := w.ships(c); n > c.PeakShips {
		c.PeakShips = n
	}
	if w.Cfg.TraceAI && want.Ships != have {
		w.explain(c, "ships", want)
	}
}

// keep is the lay-up and the rot: a fleet whose use went dark this tick
// is laid up at its base, fights nothing and moves nowhere, and loses a
// tenth of its ships a thousand years; manned again when fed. Says
// whether any fleet at a base is laid up.
func (w *World) keep(c *Civ) bool {
	w.salvageWorn(c)
	laid := false
	for _, x := range w.fleetsOf(c) {
		shed := c.Shed["fleet:"+itoa(x.ID)]
		switch {
		case shed && x.Base >= 0 && !x.LaidUp:
			x.LaidUp = true
			if x.Ships >= 2 && w.Now-x.Manned >= laidLineAfter*w.Cfg.Step {
				w.log("The %s lay up %s at %s: the ships stay where they are, and nothing keeps them.", c.Name, shipsWord(x.Ships), w.star(x.Base))
			}
			x.Laid = w.Now
		case !shed && x.LaidUp:
			x.LaidUp = false
			if x.Ships >= 2 && w.Now-x.Laid >= laidLineAfter*w.Cfg.Step {
				w.log("The %s man the ships at %s again.", c.Name, w.star(x.Base))
			}
			x.Manned = w.Now
		}
		if !shed {
			continue
		}
		laid = laid || x.LaidUp
		if lost := w.count(rotRate * float64(x.Ships)); lost > 0 {
			x.Ships -= lost
			c.Tally.Rotted += lost
			if x.Ships <= 0 {
				x.Ships = 0
				w.fleetLost(x, max(x.Base, x.Star))
				x.Over = true
			}
		}
	}
	if laid {
		c.Tally.LaidTick++
	}
	return laid
}

// shipsWord says a count of ships.
func shipsWord(n int) string {
	switch n {
	case 1:
		return "one ship"
	case 2:
		return "two ships"
	case 3:
		return "three ships"
	}
	return sprintf("%d ships", n)
}

// ShipsOf is a people's ships in being, its fleets, and its ships laid
// up, for the legends and the batch.
func ShipsOf(w *World, c *Civ) (ships, fleets, laidUp int) {
	for _, x := range w.fleetsOf(c) {
		if x.Ships == 0 {
			continue
		}
		fleets++
		ships += x.Ships
		if x.LaidUp {
			laidUp += x.Ships
		}
	}
	return
}

// Docks is how many docks a people has at work, for the legends.
func Docks(w *World, c *Civ) int {
	n := 0
	for _, r := range c.DockRate {
		if r > 0 {
			n++
		}
	}
	return n
}
