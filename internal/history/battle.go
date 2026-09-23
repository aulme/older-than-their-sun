package history

import (
	"worldgen/internal/battle"
	"worldgen/internal/mind"
)

// A battle is a campaign fleet at a world against what is in the world's
// sky: its guns, the guard there and the relief standing with it, each at
// its owner's quality. One battle a tick at a world, so a siege is several
// ticks and a withdrawal, a re-appraisal and a relief on its way all have
// somewhere to happen. The roll decides who won the day; the losses rule
// what each paid, the defender's falling on its guns first and then on
// its fleets in proportion. The loser of the roll withdraws if it lost
// ships; a fleet that lost the roll and nothing else holds its ground. A
// world whose fleets have gone and no gun stands over is taken; a world
// with nothing in its sky is taken without a battle. No battle happens
// without a fleet at the world: the front is only where a fleet can be
// sent. The arithmetic is internal/battle.

// fleetHop is how far a fleet moves between worlds in one leg, in light
// years: the next world of a campaign, a withdrawal, a relief's move.
const fleetHop = 20

// Battle is one battle at a world, for the batch.
type Battle struct {
	Year               Year
	Star               int
	Attacker, Defender int
	Ships              int     // the attacker's
	Held               int     // ships and guns in the sky
	Gap                float64 // the attacker's levels less the defender's
	Won                bool    // the attacker won the roll
	Outcome            string  // empty (taken without a battle), taken, withdrew (the fleets did, the guns stand), guns (a gun stands and any fleet stayed behind it), held (the defender won the roll), broken (the attacker is gone)
}

// sky is what holds a world: its guns, the guard, the relief, and their
// strength summed.
type sky struct {
	guns     int
	guard    *Expedition
	relief   []*Expedition
	strength float64
}

func (s sky) empty() bool { return s.strength <= 0 }

// ships is everything in the sky as a count, guns included.
func (s sky) ships() int {
	n := s.guns
	if s.guard != nil {
		n += s.guard.Ships
	}
	for _, x := range s.relief {
		n += x.Ships
	}
	return n
}

// skyAt is what holds a star for a people: the guns standing, the guard
// manned, and the relief at it.
func (w *World) skyAt(e *Civ, t int) sky {
	q := w.skyQuality(e, t)
	s := sky{guns: w.gunsAt(e, t)}
	s.strength = battle.Strength(s.guns, q)
	if g := w.guardAt(e, t); g != nil && !g.LaidUp && g.Ships > 0 {
		s.guard = g
		s.strength += battle.Strength(g.Ships, q)
	}
	for _, x := range w.reliefsAt(e, t) {
		s.relief = append(s.relief, x)
		s.strength += battle.Strength(x.Ships, w.quality(w.Civs[x.Owner]))
	}
	return s
}

// defence is what a world is held with, as a strength.
func (w *World) defence(e *Civ, t int) float64 { return w.skyAt(e, t).strength }

// holds says whether a people holds a star: its world, or for a horde a
// fleet based there.
func (w *World) holds(e *Civ, t int) bool {
	if e.Aloft {
		return w.guardAt(e, t) != nil
	}
	return w.Owner[t] == e.ID
}

// levelGap is the attacker's fighting levels less the defender's.
func (w *World) levelGap(c, e *Civ) float64 {
	return (c.Mil + c.warBonus()) - (e.Mil + e.warBonus())
}

// fight is one battle at a world: a campaign fleet at its base against
// the sky there.
func (w *World) fight(x *Expedition, t int) {
	c, e := w.Civs[x.Owner], w.Civs[x.Target]
	wr := w.warBetween(c.ID, e.ID)
	if wr == nil || !e.Active() {
		w.resolve(x)
		return
	}
	if w.foughtAt == nil {
		w.foughtAt = map[int]Year{}
	}
	if w.foughtAt[t] == w.Now {
		return // one battle a tick at a world
	}
	w.foughtAt[t] = w.Now
	if e.Asleep {
		w.rouse(e, c) // a fleet in its sky is a disturbance
	}
	i := wr.side(c.ID)
	wr.Contested[t]++
	w.stands(e, t) // a leader at the front with no campaign out comes to the world fought over
	s := w.skyAt(e, t)
	rec := &Battle{Year: w.Now, Star: t, Attacker: c.ID, Defender: e.ID, Ships: x.Ships, Held: s.ships(), Gap: w.levelGap(c, e)}
	w.Battles = append(w.Battles, rec)
	x.Battles++
	c.Tally.Battles++
	w.observe(c, e, t, 0.3)
	w.observe(e, c, c.Home, 0.3)
	if s.empty() {
		rec.Outcome, rec.Won = "empty", true
		x.Wins++
		c.Tally.Won++
		c.Tally.EmptySky++
		if g := w.guardAt(e, t); g != nil {
			w.fleetBroken(wr, c, g, t) // laid up, and nothing to save it
		}
		w.take(wr, x, c, e, t, true)
		return
	}
	had := x.Ships
	atk := battle.Strength(x.Ships, w.fleetQuality(c, x))
	won, la, ld := battle.Fight(w.R, atk, s.strength)
	rec.Won = won
	lostA := w.payAttacker(c, x, t, la)
	gunsD, lostD := w.payDefender(wr, c, e, t, s, ld)
	w.fought(c, x, t, won, lostA, had) // a leader riding with it may fall with the field
	w.heldWith(e, t, !won, gunsD+lostD, s.ships())
	if x.Ships <= 0 {
		rec.Outcome = "broken"
		w.fact(FDefeat, c, e, t).with(P{"way": "broken"})
		wr.Will[i] -= 0.2
		wr.Will[1-i] += 0.2
		w.resolve(x)
		return
	}
	if !won {
		wr.Will[i] -= 0.1
		wr.Will[1-i] += 0.1
		rec.Outcome = "held"
		if lostA > 0 {
			w.withdraw(x, t)
		}
		return
	}
	x.Wins++
	c.Tally.Won++
	wr.Will[i] += 0.1
	wr.Will[1-i] -= 0.1
	fleets := s.guard != nil || len(s.relief) > 0
	if fleets && lostD > 0 {
		w.withdrawSky(e, t, s)
		fleets = w.guardAt(e, t) != nil || len(w.reliefsAt(e, t)) > 0
	}
	guns := w.gunsAt(e, t)
	switch {
	case e.Aloft:
		rec.Outcome = "withdrew"
		if w.guardAt(e, t) != nil {
			rec.Outcome = "held"
		}
	case !fleets && guns == 0 && w.strikeBurns(x, e, t):
		rec.Outcome = "burned"
	case !fleets && guns == 0:
		rec.Outcome = "taken"
		if lostA > 0 && lostA >= x.Ships {
			w.event(KTakenDear, c, e, t, P{})
		}
		w.take(wr, x, c, e, t, false)
	case !fleets && s.guard == nil && len(s.relief) == 0:
		rec.Outcome = "guns" // the guns alone, and a gun still stands
	case !fleets:
		rec.Outcome = "withdrew"
		if w.chance(0.3) {
			w.event(KLeftToGuns, e, c, t, P{})
		}
	default:
		rec.Outcome = "guns"
		if x.Battles == 1 || w.chance(0.1) {
			w.event(KHeldBehindGuns, e, c, t, P{})
		}
	}
}

// payAttacker is a campaign fleet's losses, in ships at its owner's
// quality; the fleet pays them all, and they lie where they fell.
func (w *World) payAttacker(c *Civ, x *Expedition, t int, loss float64) int {
	k := battle.ToShips(w.R, loss, w.fleetQuality(c, x), x.Ships)
	x.Ships -= k
	c.Tally.ShipsLost += k
	w.leaveField(c, k, t, w.pos(t), false)
	return k
}

// payDefender is the losses of what holds a world: the guns first, then
// the fleets in the sky in proportion to their strength, each at its
// owner's quality. Returns guns and ships lost.
func (w *World) payDefender(wr *War, c, e *Civ, t int, s sky, loss float64) (guns, ships int) {
	q := w.skyQuality(e, t)
	if s.guns > 0 {
		guns = battle.ToShips(w.R, loss, q, s.guns)
		e.Guns[t] -= guns
		loss -= float64(guns) * q
		if guns < s.guns {
			loss = 0 // a gun still stands: the guns took the whole blow
		}
		if guns > 0 && w.gunsAt(e, t) == 0 && w.gridAt(e, t) && !e.GridBroken[t] {
			if e.GridBroken == nil {
				e.GridBroken = map[int]bool{}
			}
			e.GridBroken[t] = true
			w.event(KGunsSilent, e, c, t, P{})
		}
	}
	if loss <= 0 {
		return
	}
	var fl []*Expedition
	if s.guard != nil {
		fl = append(fl, s.guard)
	}
	fl = append(fl, s.relief...)
	total := 0.0
	for _, f := range fl {
		total += battle.Strength(f.Ships, w.quality(w.Civs[f.Owner]))
	}
	for _, f := range fl {
		o := w.Civs[f.Owner]
		qf := w.quality(o)
		k := battle.ToShips(w.R, loss*battle.Strength(f.Ships, qf)/total, qf, f.Ships)
		f.Ships -= k
		ships += k
		o.Tally.ShipsLost += k
		w.leaveField(o, k, t, w.pos(t), false)
		if f.Ships <= 0 {
			w.fleetBroken(wr, c, f, t)
		}
	}
	return
}

// fleetBroken is a fleet in a world's sky with nothing left: a horde's
// fleet broken is a line and a loss the war counts.
func (w *World) fleetBroken(wr *War, c *Civ, f *Expedition, t int) {
	o := w.Civs[f.Owner]
	w.leaveField(o, f.Ships, t, w.pos(t), false)
	o.Tally.ShipsLost += f.Ships
	f.Ships = 0
	f.Over = true
	w.fleetLost(f, t)
	if f.Kind == Roam {
		i := wr.side(c.ID)
		wr.Glassed[i]++
		wr.Will[i] += 0.2
		wr.Will[1-i] -= 0.2
		w.event(KFleetBroken, c, o, t, P{})
	}
}

// take is a world with nothing left to hold it: the attacker's by its
// nature, or stripped by a horde, or for a horde's base nothing, since a
// horde with nothing there has moved on.
func (w *World) take(wr *War, x *Expedition, c, e *Civ, t int, empty bool) {
	if e.Aloft {
		return
	}
	w.fleet, w.emptySky = x, empty
	if c.Aloft {
		x.Kind, x.Base, x.Fed = Roam, t, w.Now
		w.strip(wr, c, e, t)
		w.fleet, w.emptySky = nil, false
		return
	}
	w.takeWorld(wr, c, e, t)
	w.fleet, w.emptySky = nil, false
	if w.Owner[t] == c.ID {
		x.Held = append(x.Held, t)
		x.Base = t
	}
	if wr.Over || !e.Active() {
		w.resolve(x)
	}
}

// withdrawSky is the defender's fleets leaving a world they lost ships
// over: the guard to the nearest own holding within a hop, else an empty
// star, else home, and it wants to come back; a horde's fleet moves on;
// a relief to the nearest world of its host within a hop, else home.
func (w *World) withdrawSky(e *Civ, t int, s sky) {
	if g := s.guard; g != nil && g.Ships > 0 {
		switch g.Kind {
		case Roam:
			if n := w.nextStar(e, g, mind.RoamHop(e.Reach, w.Cfg.Tuning)); n >= 0 {
				w.moveFleet(e, g, n)
			}
		default:
			if dest := w.fallback(e, t); dest >= 0 {
				w.sail(g, dest)
				g.Returning = true
			}
		}
	}
	for _, x := range s.relief {
		if x.Ships <= 0 {
			continue
		}
		h := w.Civs[x.Target]
		if dest := w.nearestOf(h, t, fleetHop, t); dest >= 0 {
			w.sail(x, dest)
			x.Withdrawn = true
		} else {
			w.goHome(x)
		}
	}
}

// fallback is where a fleet withdraws to from a star: the nearest own
// holding within a hop, else the nearest empty star within a hop, else
// home; -1 when the star is home and nowhere else is within a hop.
func (w *World) fallback(c *Civ, t int) int {
	if s := w.nearestOf(c, t, fleetHop, t); s >= 0 {
		return s
	}
	for _, s := range w.G.Near(t, fleetHop) {
		if w.Owner[s] < 0 {
			return s
		}
	}
	if c.Home != t {
		return c.Home
	}
	return -1
}

// withdraw is an attacking fleet that lost ships falling back a hop. It
// re-appraises against what it now knows is at the world: with the ships
// the sizing needs it comes again, which is a siege; without, it goes
// home, and the campaign is lost.
func (w *World) withdraw(x *Expedition, t int) {
	c, e := w.Civs[x.Owner], w.Civs[x.Target]
	need := w.needAt(c, e, t, x.Ships)
	dest := w.fallback(c, t)
	if x.Ships < need || dest < 0 {
		w.fact(FDefeat, c, e, t).with(P{"way": "withdrew"})
		w.goHome(x)
		return
	}
	if x.Sieges == 0 {
		w.event(KFellBack, c, e, t, P{})
	}
	x.Sieges++
	w.sail(x, dest)
	x.Back = t
}

// needAt is the ships the sizing says even odds at a world would take,
// for a fleet of so many in the field.
func (w *World) needAt(c, e *Civ, t, ships int) int {
	ap := w.appraise(c, e, t)
	k := mind.SizeCampaign(mind.CampaignInput{
		Appraisal: ap.Appraisal, Strength: w.strength(c, e), Mil: c.Mil, Bonus: c.warBonus(), Ships: ships, Total: ships,
		Risk: c.Dials.Risk, Fear: c.Dials.Fear, Conqueror: c.posture() == mind.Conqueror,
	}, w.Cfg.Tuning)
	return k.Need
}

// sail puts a fleet on a new leg to a star from where it is.
func (w *World) sail(x *Expedition, t int) {
	c := w.Civs[x.Owner]
	from := x.Base
	if from < 0 {
		from = x.Star
	}
	x.From, x.Star = from, t
	x.Base = -1
	x.Path = nil
	x.Launched = w.Now
	x.Arrive = w.Now + Year(w.G.Dist(from, t)*c.Speed)
	x.Drive = c.Speed
	w.timetable(x)
}

// reliefsAt is the relief fleets of others standing with a people at a
// world, manned.
func (w *World) reliefsAt(h *Civ, t int) []*Expedition {
	var out []*Expedition
	for _, x := range w.liveFleets() {
		if x.Kind != Relief || x.Base != t || x.Target != h.ID || x.Returning || x.LaidUp || x.Ships <= 0 {
			continue
		}
		out = append(out, x)
	}
	return out
}

// reliefAt is the ships of others standing with a people at a world.
func (w *World) reliefAt(h *Civ, t int) int {
	r := 0
	for _, x := range w.reliefsAt(h, t) {
		r += x.Ships
	}
	return r
}

// nearestOf is a people's holding nearest a star within a range, not the
// star itself when it is named, or -1.
func (w *World) nearestOf(e *Civ, star int, within float64, not int) int {
	best, bd := -1, within
	for _, s := range w.holdings(e) {
		if s == not {
			continue
		}
		if d := w.G.Dist(star, s); d <= bd {
			best, bd = s, d
		}
	}
	return best
}

// fewerWon says whether the side with fewer ships won a battle: the
// attacker's ships against everything in the sky, guns counted.
func (b *Battle) fewerWon() bool {
	if b.Ships == b.Held {
		return false
	}
	return b.Won == (b.Ships < b.Held)
}
