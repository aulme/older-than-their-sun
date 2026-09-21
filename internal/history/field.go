package history

import "worldgen/internal/tech"

// Wrecks. Every ship lost in a battle stays where it fell: a wreck (four
// in five, broken, its makers' arts still in it) or a derelict (one in
// five, bad shape, but a ship), gathered into a field: one remain per
// battle per side that lost a ship, a Legacy of kind Field with the loser
// as its maker and the best weapon or drive it knew as what the hulls
// teach. A second battle at the same star by the same loser adds to the
// field there. A field from a meeting in the dark is adrift: it sits at
// the nearer end of the line for the gazetteer and keeps the point, and
// is found by whoever reads that star, or at once by any line that
// passes within half a light year of it. Fields wear like other remains,
// slowly adrift and fast at a living world, and at Ruin no ship is left.
// The Find reads a field for hulls: mastered, the maker's art; wielded,
// the derelicts and a quarter of the wrecks crewed and flown home as
// salvage, which nobody keeps running for long; never sealed.

const (
	derelictShare = 0.2 // of the ships lost, the derelicts; the rest are wrecks
	hardyAdrift   = 0.3
	hardyLiving   = 0.8 // at a star with a living world, where a field is picked over
	hardyDead     = 0.5
	wreckSalvage  = 4   // one wreck in this many flies again
	salvageWear   = 0.1 // of the ships taken from a field, lost per thousand years
)

// ships is what is left in a field to fly: nothing at Ruin.
func (l *Legacy) ships() int {
	if l.Kind != Field || l.Cond >= Ruin {
		return 0
	}
	return l.Wrecks + l.Derelicts
}

// leaveField leaves the ships a people lost in a battle where they fell,
// growing the field there if the same people left one at the star.
func (w *World) leaveField(loser *Civ, ships int, star int, at vec, adrift bool) *Legacy {
	if ships <= 0 {
		return nil
	}
	derelicts := 0
	for range ships {
		if w.R.Float64() < derelictShare {
			derelicts++
		}
	}
	wrecks := ships - derelicts
	if !adrift {
		for _, l := range w.Legacies {
			if l.Kind == Field && !l.Adrift && l.Maker == loser.ID && l.Star == star && l.State == Buried && l.Cond < Ruin {
				l.Wrecks += wrecks
				l.Derelicts += derelicts
				if derelicts > 0 {
					l.Cond = Derelict
				}
				l.Desc = fieldDesc(loser, l)
				return l
			}
		}
	}
	l := &Legacy{ID: len(w.Legacies), Age: -1, Maker: loser.ID, Kind: Field, Star: star, Node: bestArt(loser), People: -1, Finder: -1, Source: -1, Plague: -1,
		Wrecks: wrecks, Derelicts: derelicts, At: at, Adrift: adrift, Cond: Wreck, Hardy: hardyDead}
	if derelicts > 0 {
		l.Cond = Derelict
	}
	switch {
	case adrift:
		l.Hardy = hardyAdrift
	case w.Owner[star] >= 0:
		l.Hardy = hardyLiving
	}
	l.Desc = fieldDesc(loser, l)
	w.Legacies = append(w.Legacies, l)
	w.testament(loser, l)
	return l
}

// fieldDesc names a field for its makers.
func fieldDesc(c *Civ, l *Legacy) string {
	if l.Derelicts > 0 && l.Wrecks == 0 {
		return "the derelicts of the fleet of the " + c.Name
	}
	return "the wrecks of the fleet of the " + c.Name
}

// bestArt is the best weapon or drive a people knows: what its hulls
// teach. The latest era, and within it the tree's order.
func bestArt(c *Civ) string {
	best, era := "", -1
	for _, k := range knownOf(c) {
		n := tech.Get(k)
		if (n.Domain == tech.Weapons || n.Domain == tech.Propulsion) && n.Era > era {
			best, era = k, n.Era
		}
	}
	if best == "" {
		for _, k := range knownOf(c) {
			if n := tech.Get(k); n.Era > era {
				best, era = k, n.Era
			}
		}
	}
	return best
}

// salvage is a wielded field: the derelicts and a quarter of the wrecks
// crewed, after a fashion, and flown to the nearest holding, where they
// join the guard as salvage; the field is a ruin after.
func (w *World) salvage(c *Civ, l *Legacy) {
	n := 0
	if l.ships() > 0 {
		n = l.Derelicts + l.Wrecks/wreckSalvage
	}
	l.Cond = Ruin
	if n <= 0 {
		w.log("The %s try to crew the hulls. Nothing in them will fly again.", c.Name)
		return
	}
	at := l.At
	if !l.Adrift {
		at = w.pos(l.Star)
	}
	to, _ := w.nearestTo(c, at)
	x := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: -1, Kind: Guard, Star: to, From: l.Star, Ships: n, Back: -1, Contract: -1, SoldBy: -1,
		Launched: w.Now, Out: w.Now, Base: -1, Returning: true, Manned: w.Now, Seen: map[int]bool{}}
	w.addExpedition(x)
	w.legFrom(x, at, to, w.Now)
	c.Salvage += n
	c.SalvageTaken += n
	c.Tally.Salvaged += n
	w.log("The %s crew what will fly of it: %s, turned for %s.", c.Name, shipsWord(n), w.star(to))
}

// salvageWorn is the tick's loss of salvaged ships: a tenth of what was
// taken, each thousand years, from the guards, until none is left.
func (w *World) salvageWorn(c *Civ) {
	if c.Salvage <= 0 {
		return
	}
	k := min(c.Salvage, w.count(salvageWear*float64(c.SalvageTaken)))
	for _, g := range w.fleetsOf(c) {
		if k == 0 {
			break
		}
		if !g.atBase() || g.Ships <= 0 {
			continue
		}
		take := min(k, g.Ships)
		g.Ships -= take
		k -= take
		c.Salvage -= take
		if g.Ships == 0 && g.Kind == Guard {
			g.Over = true
		}
	}
	if c.Salvage <= 0 {
		c.Salvage, c.SalvageTaken = 0, 0
	}
}

// nearestTo is a people's holding nearest a point, and the distance.
func (w *World) nearestTo(c *Civ, at vec) (int, float64) {
	best, bd := c.Home, w.pos(c.Home).dist(at)
	for _, s := range w.holdings(c) {
		if d := w.pos(s).dist(at); d < bd {
			best, bd = s, d
		}
	}
	return best, bd
}

// FieldsOf is what the fields hold, for the batch: fields left, ships in
// them, fields adrift.
func FieldsOf(w *World) (fields, ships, adrift int) {
	for _, l := range w.Legacies {
		if l.Kind != Field {
			continue
		}
		fields++
		ships += l.Wrecks + l.Derelicts
		if l.Adrift {
			adrift++
		}
	}
	return
}
