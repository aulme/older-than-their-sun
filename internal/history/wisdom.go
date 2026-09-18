package history

import "worldgen/internal/mind"

// Wisdom is a people's capacity to see the counterintuitive: a fourth
// level, derived like the other three, that decides whether a people
// fathoms another and makes its acts fall nearer its own estimate. It
// makes nothing else better. A wise people wants what its posture and
// morality say it wants, and gets it more surely.
//
// Fathoming is a state between met and understood, one way: the X may
// understand the Y while the Y do not understand the X. Until c fathoms
// e, c reads e as a monster, its appraisal of e carries an extra level of
// spread, and nothing e says means anything. One-sided understanding lets
// the wise side see the other coming, make itself understood, speak for
// it to others, and sue it for a truce. Mutual understanding opens trade,
// pacts, tales, messages both ways and peace with terms.

// The level.
const (
	wisBase           = 2.5
	wisExpEach        = 0.1 // per woe or folly remembered
	wisExpCap         = 2
	wisRenaissance    = 0.5
	wisRenaissanceCap = 1
	wisCommunion      = 1
	wisSight          = 1
	wisCycle          = 1
	wisOssified       = -1
	wisVassal         = -0.25
	wisSlave          = -0.5
)

// wisTable is what each trait does to Wisdom, the same shape as traitDiff.
var wisTable = map[string]float64{
	"contemplative": 1.5,
	"memory":        1,
	"longlived":     1,
	"curious":       0.5,
	"cautious":      0.5,
	"pragmatic":     0.5,
	"shortlived":    -1,
	"herd":          -1,
	"xenophobic":    -0.5, // it sees less because it looks less
	"swarming":      -1,
}

// The roll.
const (
	fathomBase        = 3
	fathomNoise       = 1.5
	fathomSignal      = 1    // contact by signal only, no face yet
	fathomFamiliar    = 0.1  // per ten thousand years since the meeting
	fathomFamiliarMax = 2    // and no more than this
	fathomWar         = 1    // a war fought and ended: you learn what a people is by fighting it
	fathomTaught      = 2    // the other fathoms this one and is making itself understood
	fathomBroker      = 3    // a broker stands for this one, for the brokered attempt
	fathomRetry       = 0.02 // retries per thousand years
	fathomRetryTaught = 0.05 // while taught
	fathomForget      = 0.3  // each fathoming lost in a dark age
)

// wisInput is what Wisdom is derived from.
type wisInput struct {
	Traits       []string // the species' trait keys
	Profile      float64  // what the substrate and the shape add
	Tech         float64  // the working nodes' Wis summed
	Experience   int      // woes and follies the people remembers going through
	Renaissances int
	Communion    bool // spoke with a sleeper
	Sight        bool
	Cycle        bool // knows the shape of the cycle
	Ossified     bool
	Vassal       bool
	Slave        bool
}

// wisParts is where a people's Wisdom comes from, for the batch.
type wisParts struct {
	Species, Tech, Experience, Boons, Scars float64
}

// wisdom derives the level and its parts, unclamped.
func wisdom(in wisInput) (float64, wisParts) {
	var p wisParts
	p.Species = wisBase + in.Profile
	for _, t := range in.Traits {
		p.Species += wisTable[t]
	}
	p.Tech = in.Tech
	p.Experience = min(wisExpCap, wisExpEach*float64(in.Experience))
	p.Boons = min(wisRenaissanceCap, wisRenaissance*float64(in.Renaissances))
	if in.Communion {
		p.Boons += wisCommunion
	}
	if in.Sight {
		p.Boons += wisSight
	}
	if in.Cycle {
		p.Boons += wisCycle
	}
	if in.Ossified {
		p.Scars += wisOssified
	}
	switch {
	case in.Slave:
		p.Scars += wisSlave
	case in.Vassal:
		p.Scars += wisVassal
	}
	return p.Species + p.Tech + p.Experience + p.Boons + p.Scars, p
}

// setWisdom derives Wisdom for a people from the working tree's sum and
// the rest; called from recompute after the three levels. Morale does not
// touch it, and neither does age by itself: the old are wise only through
// what they remember.
func (w *World) setWisdom(c *Civ, tree float64) {
	in := wisInput{
		Tech: tree, Profile: c.Species.Profile().Wis, Experience: c.experience, Renaissances: c.Renaissances,
		Communion: c.Boons[BoonCommunion], Sight: c.miracle("foresight"), Cycle: c.KnowsCycle,
		Ossified: c.Scars[ScarOssified], Vassal: !c.Free() && c.Vassal, Slave: !c.Free() && !c.Vassal,
	}
	for _, t := range c.Species.Traits {
		in.Traits = append(in.Traits, t.Key)
	}
	x, parts := wisdom(in)
	c.Wis, c.WisFrom = clamp(x, 0, 10), parts
	c.PeakWis = max(c.PeakWis, c.Wis)
}

// experienced counts the tales a people holds of its own woes and
// follies: what it went through, whatever it tells itself about whose
// fault it was. Called from reckon, which walks the same tales.
func (w *World) experienced(c *Civ) {
	n := 0
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Facts[t.Fact]
		if f.Subject == c.ID && (f.sort() == Woe || f.sort() == Folly) {
			n++
		}
	}
	c.experience = n
}

// WisdomParts is where a people's Wisdom comes from, as last derived:
// species, tech, experience, boons, scars.
func WisdomParts(c *Civ) [5]float64 {
	p := c.WisFrom
	return [5]float64{p.Species, p.Tech, p.Experience, p.Boons, p.Scars}
}

// Difference is how alien two peoples are to each other, for the batch.
func Difference(a, b *Civ) float64 { return difference(a.Species, b.Species) }

// WisdomWord is the portrait's word for a people's Wisdom, at the ends
// only: nothing between.
func WisdomWord(wis float64) string {
	switch {
	case wis >= 8:
		return "a wise people"
	case wis <= 2:
		return "a foolish people"
	}
	return ""
}

// Fathoming is one people coming to understand another, for the batch.
type Fathoming struct {
	Year     Year
	Who      int // the one who understands
	Whom     int
	How      string  // meeting, kin, chorus, familiarity, war, taught, broker
	Since    Year    // when the two met
	Diff     float64 // how alien the two are
	Mutual   bool    // this made the pair mutual
	Reversed bool    // the other already understood this one: the taught side following
}

// fathomRoll is the roll: Wis + N(0, 1.5) against 3, how alien the other
// is, and the adjustments.
func fathomRoll(wis, noise, diff, adj float64) bool {
	return wis+noise*fathomNoise >= fathomBase+diff+adj
}

// fathoms says whether c understands e.
func (w *World) fathoms(c, e *Civ) bool { return c.Fathomed[e.ID] }

// mutual says whether two peoples understand each other.
func (w *World) mutual(a, b *Civ) bool { return a.Fathomed[b.ID] && b.Fathomed[a.ID] }

// unfathomed says whether c has met e and does not understand it: e is a
// monster to c in every rule that asks. Watching a young people from
// orbit is not contact, so it is not this.
func (w *World) unfathomed(c, e *Civ) bool {
	_, tried := c.FathomTried[e.ID]
	return tried && !c.Fathomed[e.ID]
}

// kin says whether two peoples understand each other at once: a branch,
// a made successor, an uplift, or the reverse. Step 15's kinship of a
// broken people extends this.
func (w *World) kin(a, b *Civ) bool {
	return a.Species.Kin(b.Species) || a.Sire == b.ID || b.Sire == a.ID
}

// fathomAdj is the adjustment on c's roll to fathom e from how they
// stand: signal only, familiarity, a war fought and ended, e making
// itself understood.
func (w *World) fathomAdj(c, e *Civ) float64 {
	adj := 0.0
	if !c.Reached[e.ID] {
		adj += fathomSignal
	}
	if since, ok := c.FathomTried[e.ID]; ok {
		adj -= min(fathomFamiliarMax, fathomFamiliar*float64(w.Now-since)/10_000)
	}
	fought := c.Fought[e.ID]
	if c.Wars[e.ID] {
		fought--
	}
	if fought > 0 {
		adj -= fathomWar
	}
	if w.taught(c, e) {
		adj -= fathomTaught
	}
	return adj
}

// taught says whether e fathoms c and is making itself understood to c.
func (w *World) taught(c, e *Civ) bool {
	return e.Fathomed[c.ID] && w.teaches(e, c)
}

// teaches is whether c, which fathoms e, wants e to understand it: unless
// it hates e, is at war with e, or strikes first and reads e as prey.
func (w *World) teaches(c, e *Civ) bool {
	mil, _ := w.believe(c, e)
	g := mind.Teach(mind.TeachInput{
		Hates: c.hates(e), AtWar: c.Wars[e.ID], Posture: c.posture(),
		Fixed: c.fixed(Conquest) || c.fixed(Holding), Weaker: mil < c.Mil-w.Cfg.Tuning.Wisdom.TeachWeaker,
	}, w.Cfg.Tuning)
	return g.Teach
}

// tryFathom is one attempt by c to understand e: automatic for the Chorus
// and for kin, else the roll with the standing adjustments and any extra
// (a broker's). Returns whether c fathoms e after it.
func (w *World) tryFathom(c, e *Civ, how string, extra float64) bool {
	if c.Fathomed[e.ID] {
		return true
	}
	if _, ok := c.FathomTried[e.ID]; !ok {
		c.FathomTried[e.ID] = w.Now
	}
	switch {
	case c.miracle("chorus"):
		how = "chorus"
	case w.kin(c, e):
		how = "kin"
	case !fathomRoll(c.Wis, w.R.NormFloat64(), difference(c.Species, e.Species), w.fathomAdj(c, e)+extra):
		return false
	}
	w.fathomed(c, e, how)
	return true
}

// fathomed is c coming to understand e: the record, the fact, the line,
// and what opens when the pair is mutual.
func (w *World) fathomed(c, e *Civ, how string) {
	c.Fathomed[e.ID] = true
	c.Tally.Fathomed++
	since := c.FathomTried[e.ID]
	rec := Fathoming{Year: w.Now, Who: c.ID, Whom: e.ID, How: how, Since: since, Diff: difference(c.Species, e.Species), Mutual: e.Fathomed[c.ID], Reversed: e.Fathomed[c.ID]}
	w.Fathomings = append(w.Fathomings, rec)
	w.fact(FFathomed, c, e, -1)
	ago := span(w.Now - since)
	wars := c.Fought[e.ID]
	switch {
	case how == "kin" && rec.Mutual:
		w.log("The %s and the %s, of one blood, understand each other at once.", c.Name, e.Name)
	case how == "kin", how == "meeting":
		// the meeting's own line stands for it
	case how == "chorus":
		w.log("The %s, who hold the Chorus, understand the %s at once.", c.Name, e.Name)
	case how == "taught" && rec.Mutual:
		w.log("The %s, long spoken to, at last understand the %s.", c.Name, e.Name)
	case how == "broker":
		w.log("Within a generation the %s understand the %s.", c.Name, e.Name)
	case rec.Mutual && wars > 0:
		w.log("After %s of silence and %s, the %s come to understand the %s, and find the %s had been talking the whole time.", ago, warsOf(wars), c.Name, e.Name, e.Name)
	case rec.Mutual:
		w.log("After %s of silence, the %s come to understand the %s, and find the %s had been talking the whole time.", ago, c.Name, e.Name, e.Name)
	case wars > 0:
		w.log("After %s of silence and %s, the %s come to understand the %s. The %s do not understand them.", ago, warsOf(wars), c.Name, e.Name, e.Name)
	default:
		w.log("After %s of silence, the %s come to understand the %s. The %s do not understand them.", ago, c.Name, e.Name, e.Name)
	}
	if rec.Mutual {
		w.openPair(c, e)
	}
}

func warsOf(n int) string {
	switch n {
	case 1:
		return "a war"
	case 2:
		return "two wars"
	}
	return sprintf("%d wars", n)
}

// fathomPair is a meeting: each side rolls once at once, and thereafter
// each retries on its own.
func (w *World) fathomPair(a, b *Civ) {
	w.tryFathom(a, b, "meeting", 0)
	w.tryFathom(b, a, "meeting", 0)
}

// openPair is what mutual understanding opens: the tales each holds, and
// trade if nobody strikes and neither remembers the other as a monster.
// At a meeting it runs after the councils; on a retry, on its own.
func (w *World) openPair(a, b *Civ) {
	if !a.Active() || !b.Active() {
		return
	}
	w.exchange(a, b)
	if a.Wars[b.ID] || w.monster(a, b) || w.monster(b, a) || a.Trade[b.ID] {
		return
	}
	a.Trade[b.ID], b.Trade[a.ID] = true, true
	w.fact(FTrade, a, b, -1)
	if !a.Reached[b.ID] {
		return // heard only: the trade is by signal, and nothing crosses with it
	}
	w.log("Slow messages cross the dark between the %s and the %s for generations, and then trade.", a.Name, b.Name)
	if (a.Faced["plague"] || b.Faced["plague"]) && w.R.Float64() < 0.3 {
		a.Plagued, b.Plagued = true, true
		w.log("Something crosses with the messages and the trade. Both the %s and the %s begin to sicken.", a.Name, b.Name)
	}
}

// fathoming is the per-tick pass for one people: the retries on each
// people it has met and does not understand, at the taught cadence where
// the other is making itself understood, and the unpaid brokering of
// those it understands to those it is sworn to.
func (w *World) fathoming(c *Civ) {
	if !c.Active() {
		return
	}
	for _, eid := range sortedInts(c.FathomTried) {
		if c.Fathomed[eid] {
			continue
		}
		e := w.Civs[eid]
		if !e.Active() {
			continue
		}
		rate, how := fathomRetry, "familiarity"
		switch {
		case w.taught(c, e):
			rate, how = fathomRetryTaught, "taught"
		case c.Fought[eid] > 0:
			how = "war"
		}
		if w.chance(rate) {
			w.tryFathom(c, e, how, 0)
		}
	}
	w.brokering(c)
}

// brokering is a people that understands two others speaking for one to
// the other for nothing: when it holds a pact with both, or is
// confederate by posture. A go-between that merely trades with both does
// nothing; its position is the middle, and it keeps it.
func (w *World) brokering(z *Civ) {
	for _, cid := range sortedInts(z.Fathomed) {
		c := w.Civs[cid]
		if !c.Active() || !c.Fathomed[z.ID] {
			continue // the buyer must understand the broker to be spoken to
		}
		for _, eid := range sortedInts(z.Fathomed) {
			if _, met := c.FathomTried[eid]; eid == cid || !met || c.Fathomed[eid] {
				continue
			}
			e := w.Civs[eid]
			if !e.Active() {
				continue
			}
			b := mind.Broker(mind.BrokerInput{SharedPact: w.allied(z, c) && w.allied(z, e), Posture: z.posture(), Xenophobic: z.Has("xenophobic"), AtWarWith: z.Wars[eid]}, w.Cfg.Tuning)
			if b.Rate == 0 || !w.chance(b.Rate) {
				continue
			}
			w.explain(z, "speaking for the "+c.Name+" to the "+e.Name, b)
			w.broker(z, c, e)
		}
	}
}

// broker is one brokered attempt: an immediate roll for c on e with the
// broker's help.
func (w *World) broker(z, c, e *Civ) {
	z.Tally.Brokered++
	w.log("The %s, who know both, speak for the %s to the %s.", z.Name, c.Name, e.Name)
	w.factOf(FBrokered, z, c, -1, e.Name)
	w.tryFathom(c, e, "broker", -fathomBroker)
}

// forgetFathomed is a dark age losing each understanding with a chance:
// they forgot how to speak to the Y.
func (w *World) forgetFathomed(c *Civ) {
	for _, eid := range sortedInts(c.Fathomed) {
		if w.R.Float64() >= fathomForget {
			continue
		}
		delete(c.Fathomed, eid)
		c.FathomTried[eid] = w.Now
		c.Tally.Unfathomed++
		if e := w.Civs[eid]; e.Living() {
			w.log("The %s forget how to speak to the %s.", c.Name, e.Name)
		}
	}
}
