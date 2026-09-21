package history

import (
	"sort"

	"worldgen/internal/mind"
)

// Fighting what cannot be seen. A conscious people cannot perceive an
// anti-memetic one, but it can perceive its own losses: every loss with
// no doer is written on the victim's side with a place, and the council
// keeps a ledger of them. When enough fall close together it deduces
// that something is there, and the deduction is a Gap: a region and a
// will, not a people. A hunt is a war against the region. The fleets
// strike stars inside it that read as empty, on whatever holds them; a
// battle is fought as any battle, a won world is taken with a doerless
// fact, a lost fleet is another loss that moves the hole. The hunt ends
// when the will is spent or the region has held nothing long enough. The
// hunted side sees the hunt coming and can fight, move or stop the
// losses; its offers of peace cannot be received. See antimemetic.go.

// Gap is a hole in a people's ledger: where its doerless losses fall.
type Gap struct {
	mind.Gap
	Star   int  // the star nearest the centre, for the facts and the lines
	Since  Year // when the hole was deduced
	Empty  Year // since when the region has held nothing the hunter cannot account for; 0 while it does
	Struck map[int]bool
}

// losses is one loss with no doer: the kinds a hunt can be deduced from,
// each with a place. The victim is c.
func (w *World) lossOf(c *Civ, f *Fact) bool {
	if f.Star < 0 || !w.veiled(c, f) {
		return false
	}
	switch f.Kind {
	case FTaken, FBurned, FHomeBroken, FScoured, FStripped, FUnmade, FWaking:
		return f.Object == c.ID
	case FDefeat, FCaught, FSurveyLost, FShipLost:
		return f.Subject == c.ID
	}
	return false
}

// ledger reads a people's doerless losses of the window into the mind's
// form, and the peoples behind them, most losses first.
func (w *World) ledger(c *Civ) (losses []mind.Loss, stars []int, behind []int) {
	k := &w.Cfg.Tuning.Kinds
	count := map[int]int{}
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Facts[t.Fact]
		ago := float64(w.Now-f.Year) / 1000
		if ago > k.HuntWindow || !w.lossOf(c, f) {
			continue
		}
		s := &w.G.Stars[f.Star]
		for range max(1, f.N) { // a waking is as many losses as worlds it took
			losses = append(losses, mind.Loss{X: s.X, Y: s.Y, Ago: ago})
			stars = append(stars, f.Star)
		}
		if o := c.other(f); o >= 0 {
			count[o]++
		}
	}
	for _, id := range sortedInts(count) {
		behind = append(behind, id)
	}
	sort.SliceStable(behind, func(i, j int) bool { return count[behind[i]] > count[behind[j]] })
	return
}

// deduce is the council reading its ledger: a hole, if the losses make
// one, and a hunt declared on it against the people behind most of them
// if that people still stands and no hunt runs on it already.
func (w *World) deduce(c *Civ) {
	if !c.launches() || c.Species.Profile().NoOne || len(c.Lore) == 0 {
		return
	}
	losses, stars, behind := w.ledger(c)
	if len(behind) == 0 {
		return
	}
	g, ok := mind.Deduce(losses, w.Cfg.Tuning)
	if !ok {
		return
	}
	e := w.Civs[behind[0]]
	if !e.Active() || c.Truce[e.ID] > w.Now {
		return
	}
	if wr := w.warBetween(c.ID, e.ID); wr != nil {
		if wr.Gap != nil {
			w.regap(wr, g, stars) // the hole read again: the region moves with the losses
		} else {
			w.huntOn(c, wr) // a war brought to them by what they cannot name: the ledger is all they have
		}
		return
	}
	gap := &Gap{Gap: g, Star: w.nearestStar(g.X, g.Y, stars), Since: w.Now, Struck: map[int]bool{}}
	wr := w.declare(c, e, "the hole in the ledger")
	if wr == nil {
		return
	}
	wr.Gap = gap
	wr.Will[0] = max(wr.Will[0], w.huntWill(g)) // the will is the deduction's: what the losses are worth
	c.Tally.Hunts++
	w.fact(FHunt, c, nil, gap.Star)
	w.log("The ledger of the %s shows a hole around %s: %d losses inside %.0f light years, and nothing in any record to say what took them. The council declares a hunt on the region.", c.Name, w.star(gap.Star), g.Losses, g.Radius)
	w.huntFleet(c, wr)
}

// huntOn turns a war into a hunt: the side that can no longer perceive
// its enemy fights the region its losses fall in, or the enemy's worlds
// as last known if there are none yet.
func (w *World) huntOn(c *Civ, wr *War) {
	e := w.Civs[wr.Sides[1-wr.side(c.ID)]]
	losses, stars, _ := w.ledger(c)
	g, ok := mind.Deduce(losses, w.Cfg.Tuning)
	if !ok {
		s := w.aWorld(e)
		g = mind.Gap{X: w.G.Stars[s].X, Y: w.G.Stars[s].Y, Radius: w.Cfg.Tuning.Kinds.HuntRadius / 2}
		stars = []int{s}
	}
	if wr.Sides[0] != c.ID {
		wr.Sides[0], wr.Sides[1] = wr.Sides[1], wr.Sides[0]
		wr.Will[0], wr.Will[1] = wr.Will[1], wr.Will[0]
		wr.Taken[0], wr.Taken[1] = wr.Taken[1], wr.Taken[0]
		wr.Glassed[0], wr.Glassed[1] = wr.Glassed[1], wr.Glassed[0]
		wr.Lost[0], wr.Lost[1] = wr.Lost[1], wr.Lost[0]
	}
	wr.Gap = &Gap{Gap: g, Star: w.nearestStar(g.X, g.Y, stars), Since: w.Now, Struck: map[int]bool{}}
	wr.Will[0] = max(wr.Will[0], w.huntWill(g))
	c.Tally.Hunts++
	w.log("The %s are at war with something they can no longer name. What they have is the ledger, and the ledger says %s.", c.Name, w.star(wr.Gap.Star))
}

// huntWill is what a people brings to a hunt: one, and a quarter per
// loss the hole was read from.
func (w *World) huntWill(g mind.Gap) float64 { return 1 + 0.25*float64(g.Losses) }

// regap moves a hunt's region to the hole as read now.
func (w *World) regap(wr *War, g mind.Gap, stars []int) {
	old := wr.Gap
	if g.X == old.X && g.Y == old.Y && g.Radius == old.Radius {
		return
	}
	old.Gap = g
	old.Star = w.nearestStar(g.X, g.Y, stars)
	old.Struck = map[int]bool{}
}

// nearestStar is the star of a list nearest a point.
func (w *World) nearestStar(x, y float64, stars []int) int {
	best, bd := stars[0], 0.0
	for i, s := range stars {
		d := (w.G.Stars[s].X-x)*(w.G.Stars[s].X-x) + (w.G.Stars[s].Y-y)*(w.G.Stars[s].Y-y)
		if i == 0 || d < bd {
			best, bd = s, d
		}
	}
	return best
}

// hunter is the side of a war that fights a region, or nil.
func (w *World) hunter(wr *War) *Civ {
	if wr.Gap == nil {
		return nil
	}
	return w.Civs[wr.Sides[0]]
}

// inGap says whether a star lies inside a hunt's region.
func (w *World) inGap(g *Gap, s int) bool {
	st := &w.G.Stars[s]
	return (st.X-g.X)*(st.X-g.X)+(st.Y-g.Y)*(st.Y-g.Y) <= g.Radius*g.Radius
}

// unaccounted says whether a star in a hunter's eyes might be where the
// hole is: nobody it can perceive holds it.
func (w *World) unaccounted(c *Civ, s int) bool {
	o := w.Owner[s]
	return o < 0 || (o != c.ID && !w.perceives(c, w.Civs[o]))
}

// huntTargets is the stars of a region a hunter could strike, nearest
// the centre first: inside it, unaccounted for, not struck this reading,
// and within reach of a holding or of the fleet if one is given.
func (w *World) huntTargets(c *Civ, g *Gap, from int) []int {
	var out []int
	for s := range w.G.Stars {
		if !w.inGap(g, s) || g.Struck[s] || !w.unaccounted(c, s) {
			continue
		}
		if from >= 0 {
			if w.G.Dist(from, s) > fleetHop {
				continue
			}
		} else if !w.inReach(c, s) {
			continue
		}
		out = append(out, s)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := &w.G.Stars[out[i]], &w.G.Stars[out[j]]
		return (a.X-g.X)*(a.X-g.X)+(a.Y-g.Y)*(a.Y-g.Y) < (b.X-g.X)*(b.X-g.X)+(b.Y-g.Y)*(b.Y-g.Y)
	})
	return out
}

// huntFleet sends a fleet into the region: a third of the standing ships
// from the guard that has them, if one has, at the nearest unaccounted
// star. Nothing is sent that no guard can man.
func (w *World) huntFleet(c *Civ, wr *War) bool {
	ts := w.huntTargets(c, wr.Gap, -1)
	if len(ts) == 0 {
		return false
	}
	n := max(1, w.standing(c)/3)
	if w.guardWith(c, ts[0], n) == nil {
		return false
	}
	e := w.Civs[wr.Sides[1]]
	x := w.launch(c, Campaign, e, ts[0], n)
	if x == nil {
		return false
	}
	wr.Gap.Struck[ts[0]] = true
	return true
}

// huntNext is where a hunting fleet goes from a star that held nothing:
// the next unaccounted star of the region within a hop, or -1.
func (w *World) huntNext(c *Civ, wr *War, from int) int {
	wr.Gap.Struck[from] = true
	ts := w.huntTargets(c, wr.Gap, from)
	if len(ts) == 0 {
		return -1
	}
	wr.Gap.Struck[ts[0]] = true
	return ts[0]
}

// huntStep is the hunts' tick for a people: a fleet sent into each
// region that has none against it, now and then; and the region read for
// whether anything is still there.
func (w *World) huntStep(c *Civ) {
	for _, eid := range sortedInts(c.Wars) {
		wr := w.warBetween(c.ID, eid)
		if wr == nil || wr.Gap == nil || wr.Sides[0] != c.ID {
			continue
		}
		e := w.Civs[eid]
		held := false
		for _, s := range w.holdings(e) {
			if w.inGap(wr.Gap, s) {
				held = true
				break
			}
		}
		switch {
		case held:
			wr.Gap.Empty = 0
		case wr.Gap.Empty == 0:
			wr.Gap.Empty = w.Now
		case float64(w.Now-wr.Gap.Empty)/1000 >= w.Cfg.Tuning.Kinds.HuntEmpty:
			w.log("The hunt of the %s finds nothing at %s, and nothing, and nothing. Whatever was there is not, and the ledger is closed.", c.Name, w.star(wr.Gap.Star))
			w.endWar(wr, "the hole closed")
			continue
		}
		if !w.hasFleetAgainst(c, e) && w.chance(w.Cfg.Tuning.Council.Cadence) {
			w.huntFleet(c, wr)
		}
	}
}
