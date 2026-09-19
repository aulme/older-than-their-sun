package history

import (
	"sort"

	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Flows: each tick a people takes in the yield of its sources, lists what
// it knows as uses with an upkeep, chooses an order by temperament and
// state, and feeds the uses in that order. What is not fed is dormant:
// recompute gives it no levels, no reach and, after ten ticks, no
// envelope. The core is flow.Direct and the order is mind.Direct; this
// file reads the state, calls them and writes the result and the lines.

// leanYears is how long a stretch of shedding runs before it is a woe.
const leanYears Year = 100_000

// envelopeGrace is how many ticks a node may be dormant before the
// envelope shrinks: a bad millennium does not empty worlds.
const envelopeGrace = 10

// flows is the civ step: income, uses, order, direction.
func (w *World) flows(c *Civ) {
	c.Reserved = flow.Income{}
	rareChanged := w.rare(c)
	c.Income = w.sickIncome(c, w.income(c))
	w.direct(c, w.uses(c), rareChanged)
	w.tallyFlows(c)
}

// direct chooses the order and feeds the uses in it, and writes what came
// of it: the spare, the want, what went dark, and the levels if the
// working set changed. Under Trade.UsesOnly goods that came by trade feed
// uses only, so the spare is capped at what the people's own income would
// leave; the rule was step 6's restraint on the war rise from fed roads
// and is off since the slight took its place. The own want is the want
// without what partners sent, which is what trade is asked for; the
// working need is what the fed uses take, which is what a partner's
// sending may be keeping fed.
func (w *World) direct(c *Civ, uses []flow.Use, rareChanged bool) {
	d := w.order(c)
	if w.Cfg.TraceAI && !sameOrder(c.Order, d.Order) {
		w.log("[the %s, direction: %s]", c.Name, d.Why())
	}
	c.Order = d.Order
	a := flow.Direct(c.Income, uses, d.Order)
	c.Surplus, c.Want = a.Surplus, a.Want
	if w.Cfg.Tuning.Trade.UsesOnly && c.Received != (flow.Income{}) {
		// what partners sent feeds the uses and nothing else: the spare
		// that launches and builds draw on is never more than the people's
		// own income would leave
		own := flow.Direct(c.Income.Less(c.Received), uses, d.Order)
		for k := range c.Surplus {
			c.Surplus[k] = min(c.Surplus[k], own.Surplus[k])
		}
	}
	c.Upkeep, c.WorkingNeed = flow.Income{}, flow.Income{}
	for i := range uses {
		c.Upkeep.Add(uses[i].Need)
	}
	dormant := map[string]bool{}
	for _, k := range a.Dormant {
		dormant[k] = true
	}
	for i := range uses {
		if !dormant[uses[i].Key] {
			c.WorkingNeed.Add(uses[i].Need)
		}
	}
	own := c.Income.Less(c.Received)
	for k := range c.OwnWant {
		c.OwnWant[k] = max(0, c.Upkeep[k]-own[k])
	}
	if c.Upkeep.Total() > c.highUpkeep {
		c.highUpkeep = c.Upkeep.Total()
		c.HighIncome, c.HighUpkeep, c.HighWant = c.Income, c.Upkeep, c.Want
		c.PeakTrade = sortedInts(c.Trade)
	}
	changed := w.setShed(c, uses, a)
	if changed || rareChanged {
		w.recompute(c)
	}
}

// redirect is a people losing part of its income within the tick, when a
// partner's sending stops: what was counted is taken back and the uses are
// fed again from what is left, so what the sending kept fed goes dark now.
func (w *World) redirect(c *Civ, lost flow.Income) {
	c.Income = c.Income.Less(lost)
	c.Received = c.Received.Less(lost)
	w.direct(c, w.uses(c), false)
}

// setShed writes the dormant set and its timing, and says whether the
// working set changed. The first shed in a stretch is a line; a stretch
// past the lean years is a woe, once.
func (w *World) setShed(c *Civ, uses []flow.Use, a flow.Allocation) bool {
	changed := false
	wasShedding := len(c.Shed) > 0
	shed := make(map[string]bool, len(a.Dormant))
	for _, k := range a.Dormant {
		shed[k] = true
		if !c.Shed[k] {
			changed = true
			if c.DormantSince == nil {
				c.DormantSince = map[string]Year{}
			}
			c.DormantSince[k] = w.Now
		}
	}
	for k := range c.Shed {
		if !shed[k] {
			changed = true
			delete(c.DormantSince, k)
		}
	}
	c.Shed = shed
	for i := range c.Works {
		c.Works[i].Dark = shed[c.Works[i].key()]
	}
	if len(a.Dormant) == 0 {
		c.ShedSince, c.Wanted = 0, false
		return changed
	}
	if !wasShedding {
		c.ShedSince = w.Now
		first := a.Dormant[0]
		for i := range uses {
			if uses[i].Key == first {
				w.log("The %s let %s go dark to keep %s fed.", c.Name, uses[i].Name, c.Order[0].Phrase())
				break
			}
		}
	} else if !c.Wanted && w.Now-c.ShedSince >= leanYears {
		c.Wanted = true
		w.fact(FWant, c, nil, c.Home)
		w.log("The %s have gone without for a hundred thousand years. They call them the lean years.", c.Name)
	}
	return changed
}

func sameOrder(a, b flow.Order) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// tallyFlows counts the tick for the batch reports.
func (w *World) tallyFlows(c *Civ) {
	e := min(c.Era, 4)
	c.Tally.Ticks++
	c.Tally.TicksAt[e]++
	alone := len(c.Systems) <= 1
	if alone {
		c.Tally.AloneAt[e]++
	}
	if len(c.Shed) > 0 {
		c.Tally.Lean++
		c.Tally.LeanAt[e]++
		if alone {
			c.Tally.LeanAloneAt[e]++
		}
		if c.ShedTicks == nil {
			c.ShedTicks = map[string]int{}
		}
		for k := range c.Shed {
			c.ShedTicks[k]++
		}
	}
}

// shedKey is what the batch counts a shed use under: a node by its key, a
// structure by its kind whatever star it stands at, a fleet or a ship as
// one kind each.
func shedKey(k string) string {
	switch {
	case len(k) > 5 && k[:5] == "work:":
		rest := k[5:]
		for i := range rest {
			if rest[i] == ':' {
				return "work:" + rest[:i]
			}
		}
		return k
	case len(k) > 6 && k[:6] == "fleet:":
		return "fleet"
	case len(k) > 5 && k[:5] == "ship:":
		return "ship"
	case len(k) > 5 && k[:5] == "dock:":
		return "dock"
	}
	return k
}

// uses lists what a people spends on: every known node with an upkeep,
// in the order it was learned so the newest of an era goes dark first;
// then its structures, its fleets and its ships. The profile bends the
// costs: a machine pays organic matter in energy, an evolver's flesh is
// cheap and its metal dear; a people with no fields feeds those nodes
// with the works.
func (w *World) uses(c *Civ) []flow.Use {
	p := c.Species.Profile()
	var out []flow.Use
	for _, k := range knownOf(c) {
		n := tech.Get(k)
		need := w.needOf(c, n)
		if need == (flow.Income{}) {
			continue
		}
		cat := n.Cat()
		if cat == flow.Fields && !p.Can(species.Fields) {
			cat = flow.Works
		}
		out = append(out, flow.Use{Key: k, Name: n.Name, Cat: cat, Era: n.Era, Need: need})
	}
	learned := func(k string) Year {
		if y, ok := c.Learned[k]; ok {
			return y
		}
		return c.Born
	}
	sort.SliceStable(out, func(i, j int) bool { return learned(out[i].Key) < learned(out[j].Key) })
	out = append(out, w.works(c)...)
	out = append(out, w.reservations(c)...)
	return append(out, w.contractUses(c)...)
}

// needOf is what a node costs a people each tick: the table's upkeep bent
// by the profile, so a machine pays organic matter in energy and an
// evolver's flesh is cheap and its metal dear.
func (w *World) needOf(c *Civ, n *tech.Node) flow.Income {
	need := n.Upkeep()
	if need == (flow.Income{}) {
		return need
	}
	if m, ok := c.Species.Profile().Upkeep[n.Domain]; ok {
		need = need.Scale(m)
	}
	return w.bend(c, need)
}

// order asks the mind for the direction.
func (w *World) order(c *Civ) mind.Direction {
	fix, fixed := c.Morality.fixation()
	return mind.Direct(mind.DirectionInput{
		Fixed:       fixed,
		Fixation:    fix,
		AtWar:       len(c.Wars) > 0,
		Fear:        c.Dials.Fear,
		HostileNear: func() bool { return w.hostileNear(c) },
		Hunger:      c.Dials.Hunger,
		Greed:       c.Dials.Greed,
		NoFields:    !c.Species.Profile().Can(species.Fields),
		Nomad:       c.Aloft,
		Honour:      c.honour(),
	}, w.Cfg.Tuning)
}

// hostileNear is true when a people this one has met strikes first and
// can reach its seat.
func (w *World) hostileNear(c *Civ) bool {
	for _, id := range sortedInts(c.Met) {
		e := w.Civs[id]
		if e.Active() && e.hostile() && w.inReach(e, c.Home) {
			return true
		}
	}
	return false
}

// working says whether a known node is fed this tick. A node learned
// since the last direction is fed until the next one.
func (c *Civ) working(k string) bool { return !c.Shed[k] }

// starved says whether a node has been dormant long enough to lose its
// envelope.
func (c *Civ) starved(k string, now Year) bool {
	since, ok := c.DormantSince[k]
	return ok && now-since >= envelopeGrace*1000
}
