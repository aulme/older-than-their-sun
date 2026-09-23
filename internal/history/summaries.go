package history

import "math"

// The kept summaries: what a people's telling says about the people,
// held as the telling changes instead of re-summed from every tale every
// tick.
//
// A telling is a few hundred tales and a people is stepped a few
// thousand times, so a walk per summary per people per tick is a large
// part of what the lore costs. Two of the three summaries do not need
// the walk. A tale's part in them is settled by the fact, by the people
// reading it and by how worn the tale is, and by nothing else — so each
// is kept as a running total, added to when a tale is learned and moved
// when a tale wears, is forgotten, or comes back off a wall.
//
// The third, reckon's tally of monsters, weighs a crime by whom it was
// done to: a trading partner, an ally. Those move under the telling
// every tick and a tale's part in the tally moves with them, so there is
// nothing to invalidate on and reckon still walks. See specs/plan.md.
//
// The dials are kept in whole four-hundredths of a dial point rather
// than in floats, because a running total is added to and taken from in
// whatever order the telling happened to change in, and floats do not
// give the same answer twice that way. Every coefficient in loreRules is
// a multiple of 0.005 and a tale's wear multiplies it by 1, 1.5 or 2, so
// a four-hundredth holds every one of them exactly: the kept total is
// the walked one to the bit, whatever order it was reached in, which is
// what lets the peoples be summed in any order at all.

// The rules: what a remembered thing does to a people's temperament,
// before the tale's wear multiplies it. What it remembers suffering
// makes it fearful and hating, what it remembers winning makes it bold,
// remembered promises make it loyal and remembered betrayals the
// reverse. Myth counts for more than memory.
//
// Every number here must be a multiple of 0.005; TestLoreRulesAreWhole
// holds it, and it is what lets the sum be kept as an integer.
const (
	loreBetrayal = iota // a promise to us broken, and we call it a crime
	loreCrime           // any other crime against us
	loreTaken           // our world taken, our home broken, our surrender
	loreFolly           // our own foolishness
	loreWoe             // our own suffering
	loreFind            // what we found, mastered, or came through
	loreBond            // a promise, ours or to us
	loreNothing
)

var loreRules = [loreNothing]Dials{
	loreBetrayal: {Loyalty: -0.10, Fear: 0.06},
	loreCrime:    {Hate: 0.06, Fear: 0.04, Patience: 0.04},
	loreTaken:    {Aggression: 0.06, Greed: 0.04},
	loreFolly:    {Risk: -0.10},
	loreWoe:      {Fear: 0.04, Risk: -0.04},
	loreFind:     {Hunger: 0.06},
	loreBond:     {Loyalty: 0.04},
}

// taleRule is which rule a tale falls under for one people, or
// loreNothing. It reads the fact and the people's judgment of it, both
// of which stand for as long as the people's morality does.
func taleRule(c *Civ, f *Event) int {
	s, _ := sortFor(c, f)
	self := f.Subject == c.ID
	switch {
	case f.Kind == FBetrayal && f.Object == c.ID && s == Crime:
		return loreBetrayal
	case s == Crime && f.Object == c.ID:
		return loreCrime
	case self && (f.Kind == FTaken || f.Kind == FHomeBroken || f.Kind == FYield):
		return loreTaken
	case self && s == Folly:
		return loreFolly
	case self && s == Woe:
		return loreWoe
	case self && (f.Kind == FFind || f.Kind == FMastered || f.Kind == FCycle):
		return loreFind
	case s == Bond && (self || f.Object == c.ID):
		return loreBond
	}
	return loreNothing
}

// dialUnits is a sum of dials in four-hundredths of a point, in the
// order of mind.Dials' fields.
type dialUnits [8]int32

// dialDenom is the unit: a dial point is four hundred of them.
const dialDenom = 400

// unitsOf turns a rule into the kept total's integers, with the tale's
// wear counted in. A tale worn counts for a half more and a tale gone to
// myth for double, which is 2, 3 or 4 halves; the halves are why the
// unit is a four-hundredth of a point and not a two-hundredth.
func unitsOf(d Dials, wear int8) dialUnits {
	k := int32(2 + wear)
	whole := func(x float64) int32 { return int32(math.Round(x*(dialDenom/2))) * k }
	return dialUnits{
		whole(d.Aggression), whole(d.Risk), whole(d.Greed), whole(d.Fear),
		whole(d.Loyalty), whole(d.Hunger), whole(d.Patience), whole(d.Hate),
	}
}

// add puts another sum in, or takes it out with sign -1.
func (u *dialUnits) add(o dialUnits, sign int32) {
	for i, v := range o {
		u[i] += sign * v
	}
}

// dials reads the kept total back as dials, clamped as the telling's
// pull on a temperament has always been.
func (u dialUnits) dials() Dials {
	at := func(i int) float64 { return clamp(float64(u[i])/dialDenom, -0.3, 0.3) }
	return Dials{
		Aggression: at(0), Risk: at(1), Greed: at(2), Fear: at(3),
		Loyalty: at(4), Hunger: at(5), Patience: at(6), Hate: at(7),
	}
}

// taleUnits is one tale's part in a people's dials.
func taleUnits(c *Civ, f *Event, t *Tale) dialUnits {
	r := taleRule(c, f)
	if r == loreNothing {
		return dialUnits{}
	}
	return unitsOf(loreRules[r], t.Wear)
}

// ownGrief is whether a tale counts toward what a people went through:
// its own woes and follies, whatever it tells itself about whose fault
// they were. It is the fact's own sort, not the people's reading of it,
// so it stands whatever the people comes to believe.
func ownGrief(c *Civ, f *Event) bool {
	return f.Subject == c.ID && (f.sort() == Woe || f.sort() == Folly)
}

// sickTale is whether a tale is one the suspicion pass reads: a people
// other than this one struck by a plague, or anyone rid of one. It asks
// the kind and the subject and nothing else, because a tale joins the
// list the moment it is learned and an event's other parameters are the
// caller's to fill after the record is made — f.Plague among them. So
// the list is a superset of what suspicion acts on, and suspicion asks
// for the named plague itself. Both are settled at the fact and by the
// people's own id, so a tale that qualifies once qualifies for as long
// as it is held, whatever the people comes to believe.
func sickTale(c *Civ, f *Event) bool {
	return f.Kind == FCured || (f.Kind == FPlague && f.Subject != c.ID)
}

// resick rebuilds the kept sickness tales from the telling. Only prune
// needs it: it drops tales and reorders what is left, and the kept list
// is in the telling's order. learned keeps the list in step the rest of
// the time, and resum builds it in the walk it already makes.
func (w *World) resick(c *Civ) {
	c.sickLore = c.sickLore[:0]
	for _, t := range c.Lore {
		if sickTale(c, w.Events[t.Fact]) {
			c.sickLore = append(c.sickLore, t)
		}
	}
}

// resum walks a whole telling and takes the kept summaries from it. It
// is what the kept numbers mean, and a test holds the kept against it at
// every tick (summaries_test.go). A people is resummed when its
// judgment moves, which is the one thing the rules are read through that
// is not the tale itself, and when a whole telling is moved at once.
func (w *World) resum(c *Civ) {
	var u dialUnits
	n := 0
	c.sickLore = c.sickLore[:0]
	for _, t := range c.Lore {
		f := w.Events[t.Fact]
		if sickTale(c, f) {
			c.sickLore = append(c.sickLore, t) // held or not: the suspicion pass skips the forgotten itself
		}
		if t.Forgot {
			continue
		}
		u.add(taleUnits(c, f, t), 1)
		if ownGrief(c, f) {
			n++
		}
	}
	c.loreUnits, c.experience, c.loreKept = u, n, true
}

// summaries makes the kept numbers good before they are read.
func (w *World) summaries(c *Civ) {
	if !c.loreKept {
		w.resum(c)
	}
}

// learned adds a tale to the kept summaries; hold calls it, and hold is
// the only way a tale joins a telling. The count of a people's own
// griefs is settled by the fact's own sort, which no change of judgment
// moves, so it is right at every moment and never has to be taken
// again; wisdom.go reads it where it stands.
func (w *World) learned(c *Civ, f *Event, t *Tale) {
	c.loreUnits.add(taleUnits(c, f, t), 1)
	if sickTale(c, f) {
		c.sickLore = append(c.sickLore, t)
	}
	if ownGrief(c, f) {
		c.experience++
	}
}

// amend is the one way a tale's wear or its forgetting changes. The
// tale's part in the kept summaries is taken off before the change and
// put back after, so the kept numbers stay what a walk of the whole
// telling would give. Nothing else may write t.Wear or t.Forgot.
func (w *World) amend(c *Civ, t *Tale, do func(*Tale)) {
	f := w.Events[t.Fact]
	held := !t.Forgot
	if held {
		c.loreUnits.add(taleUnits(c, f, t), -1)
	}
	do(t)
	if now := !t.Forgot; now {
		c.loreUnits.add(taleUnits(c, f, t), 1)
		if !held && ownGrief(c, f) {
			c.experience++ // come back off a wall
		}
	} else if held && ownGrief(c, f) {
		c.experience--
	}
}

// think sets a people's judgment. Every change of morality goes through
// here: the dials' rules are read through the morality's column
// (sortFor), so a people that changes its mind reads every tale it holds
// again.
func (c *Civ) think(m Morality) {
	c.Morality = m
	c.loreKept = false
}

// loreDials is what a people's telling does to its temperament.
func (w *World) loreDials(c *Civ) Dials {
	w.summaries(c)
	return c.loreUnits.dials()
}

// tally is a running sum by people, for a walk that adds to the same few
// peoples over and over. It is the world's and reused: a map made and
// thrown away per people per tick was a quarter of what the reckoning
// cost. A stamp per people says whether its sum belongs to this walk, so
// nothing is cleared between walks, and touched keeps the peoples in the
// order the walk reached them, so what is read back off it does not
// depend on a map's order. It is the world's and not each people's
// because one per people would be a square of the peoples in memory; a
// pass that steps the peoples beside each other wants one per worker,
// as it does for regardHolder's stamps.
type tally struct {
	sum     []float64
	seen    []uint32
	gen     uint32
	touched []int
}

// start begins a walk over the peoples of the world.
func (t *tally) start(n int) {
	if len(t.seen) < n {
		t.sum, t.seen, t.gen = make([]float64, n), make([]uint32, n), 0
	}
	t.gen++
	if t.gen == 0 { // the stamp wrapped: no sum is this walk's
		clear(t.seen)
		t.gen = 1
	}
	t.touched = t.touched[:0]
}

// add puts v on one people's sum.
func (t *tally) add(id int, v float64) {
	if t.seen[id] != t.gen {
		t.seen[id], t.sum[id] = t.gen, 0
		t.touched = append(t.touched, id)
	}
	t.sum[id] += v
}
