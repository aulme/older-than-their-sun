package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/mind"
)

// Tribute (specs/proposals/war.md, stage 3): vassalage is protection for
// a price. A vassal pays its patron a standing tribute, a share of its
// whole income at a rate the bond sets from the patron's character, how
// the bond came about and how hopeless the vassal's position was; it is a
// use in the vassal's flows like any other. Paid first, it is fed after
// the fields and before everything else, so in a lean tick what goes
// unfed is the vassal's own works, fleets and road; paid short, it stands
// where the direction puts the word, and a tick it goes unfed angers the
// patron by the share missing, until the patron makes war on its client
// for it. The rate is reviewed as the bond goes on, lightened for a
// vassal that paid in full or is in real distress under a patron no fool,
// raised for one that paid short or grew strong. What the patron takes is
// income to it, and replaces the levels a vassal once gave its master.
// The judgments are mind.TributeRate, mind.ReviewRate and mind.PaysFirst.

// tributeKey is the tribute's use in the vassal's flows.
const tributeKey = "tribute"

// tributary says whether a people pays a standing tribute: a vassal with
// a patron that takes one, and something to pay it in.
func (w *World) tributary(c *Civ) bool {
	return c.Vassal && c.Master >= 0 && c.Rate > 0 && w.Civs[c.Master].Active() && w.Civs[c.Master].Own < 0
}

// bond sets the tribute of a new vassal: its rate, where it stands in the
// vassal's flows, and a note of both in the record. How is surrender (at
// the end of a war), offer (vassalage accepted without one) or sought (a
// people that came to a patron of itself: an uplifted client, a meeting
// that bent the knee).
func (w *World) bond(m, s *Civ, how string) {
	if m.Own >= 0 || !s.Vassal {
		return // a rider's hosts pay in themselves
	}
	back := w.appraise(s, m, -1)
	in := mind.RateInput{Greed: m.Dials.Greed, Conqueror: m.posture() == mind.Conqueror, Different: m.differs(s) >= 1, How: how, Hopeless: 2 * max(0, 0.5-back.Acted), Noise: w.R.NormFloat64()}
	s.Rate = mind.TributeRate(in, w.Cfg.Tuning)
	in.Noise = 0
	s.RateTarget = mind.TributeRate(in, w.Cfg.Tuning)
	s.PaidShort, s.RateReviewed = 0, w.Now
	s.PaysFirst = w.paysFirst(s, m)
	w.note(KBond, s, m, -1, P{"rate": s.Rate, "how": how, "first": s.PaysFirst})
}

// paysFirst is the vassal's council on where its tribute stands: its
// loyalty, its awe of its patron and the patron's nearness against the
// hard times.
func (w *World) paysFirst(s, m *Civ) bool {
	back := w.appraise(s, m, -1)
	return mind.PaysFirst(mind.PayInput{Loyalty: s.Dials.Loyalty, Awe: 1 - back.Acted, Near: w.inReach(m, s.Home), Lean: len(s.Shed) > 0}, w.Cfg.Tuning)
}

// tributeUses is the tribute as a use: paid first it stands with the
// fields, fed after every one of them; otherwise with the word.
func (w *World) tributeUses(c *Civ) []flow.Use {
	if !w.tributary(c) {
		return nil
	}
	cat, era := flow.Word, 3
	if c.PaysFirst {
		cat, era = flow.Fields, 99 // after the vassal's own fields, before everything else
	}
	return []flow.Use{{Key: tributeKey, Cat: cat, Era: era, Need: c.Income.Scale(c.Rate)}}
}

// payTribute is the tick's tribute after the direction: paid, it is
// income to the patron next tick; unfed, it is paid short, and the
// patron's anger grows by the share missing.
func (w *World) payTribute(c *Civ) {
	if !w.tributary(c) {
		return
	}
	m := w.Civs[c.Master]
	c.Tally.TributeTicks++
	if c.Shed[tributeKey] {
		c.PaidShort++
		c.Tally.TributeShort++
		m.resent(c.ID, w.Cfg.Tuning.Client.Anger)
		return
	}
	m.Paid.Add(c.Income.Scale(c.Rate))
}

// reviewTribute is the bond reviewed, every so often: the rate moves
// part of the way to its target, lightened or raised by how the vassal
// has done, and the vassal decides again where the tribute stands.
func (w *World) reviewTribute(c *Civ) {
	t := &w.Cfg.Tuning.Client
	if !w.tributary(c) || float64(w.Now-c.RateReviewed) < t.ReviewEvery {
		return
	}
	m := w.Civs[c.Master]
	was := c.Rate
	c.Rate = mind.ReviewRate(mind.ReviewInput{
		Rate: c.Rate, Target: c.RateTarget, Loyal: c.PaidShort == 0, Short: c.PaidShort > 0, Strong: c.Mil >= m.Mil,
		Distress: c.PaysFirst && len(c.Shed) > 0, Wise: m.Wis >= t.Wise,
	}, w.Cfg.Tuning)
	c.PaidShort, c.RateReviewed = 0, w.Now
	c.PaysFirst = w.paysFirst(c, m)
	if d := c.Rate - was; d > 0.005 || d < -0.005 {
		w.note(KBondRate, c, m, -1, P{"rate": c.Rate, "was": was, "first": c.PaysFirst})
	}
}

// punish is a patron's council on its clients: one whose tribute it has
// been paid short of long enough to anger it past bearing is made war on,
// for the tribute.
func (w *World) punish(c *Civ) bool {
	for _, id := range sortedInts(c.Grudge) {
		o := w.Civs[id]
		if !o.Active() || o.Master != c.ID || !o.Vassal || c.Grudge[id] < w.Cfg.Tuning.Client.Punish || c.Wars[id] || c.Truce[id] > w.Now {
			continue
		}
		ap := w.appraise(c, o, -1)
		if w.strikeFirst(c, o, ap, true, because("unpaid").By(o)) {
			return true
		}
	}
	return false
}
