package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/mind"
)

// Trade moves surplus. Each tick, after every people has fed its uses,
// each sends what it has spare to its partners against what they want,
// and what arrives is counted in the next tick's income. The will and the
// road are mind.Trade; the arithmetic is flow.Share. A refusal that lasts
// is an embargo, a crime the refused remembers and a cause for war. A
// people whose fed uses are kept fed by a partner's sending is dependent
// on it, and when the sending stops its uses go dark that tick.

// embargoYears is how long a refusal runs before it is called an embargo.
const embargoYears Year = 20_000

// trade is the phase: every people sends, then every people learns what
// it depends on.
func (w *World) trade() {
	for _, c := range w.Civs {
		if !c.Active() {
			continue
		}
		for _, pid := range sortedInts(c.From) {
			if p := w.Civs[pid]; !c.Trade[pid] || !p.Active() {
				w.cutTrade(c, p, "the fall of the "+p.Name)
			}
		}
		c.From = nil
	}
	for _, c := range w.Civs {
		if c.Active() {
			w.sendGoods(c)
		}
	}
	for _, c := range w.Civs {
		if c.Active() {
			w.depend(c)
		}
	}
}

// sendGoods is one people sending its spare to its partners: per kind,
// against their own wants, as far as it is willing and the road allows.
func (w *World) sendGoods(a *Civ) {
	spare := a.Surplus.Less(a.Reserved)
	for k := range spare {
		spare[k] = max(0, spare[k])
	}
	var partners []*Civ
	var choices []mind.TradeChoice
	for _, pid := range sortedInts(a.Trade) {
		b := w.Civs[pid]
		if !b.Active() {
			continue
		}
		if a.partners == nil {
			a.partners = map[int]bool{}
		}
		if !a.partners[pid] {
			a.partners[pid] = true
			a.Tally.Partners++
		}
		ch := w.willing(a, b)
		w.explain(a, "trading with the "+b.Name, ch)
		partners = append(partners, b)
		choices = append(choices, ch)
		w.embargoStep(a, b, ch, spare)
	}
	if len(partners) == 0 {
		return
	}
	wants := make([]float64, len(partners))
	caps := make([]float64, len(partners))
	for _, k := range flow.Kinds {
		for i, b := range partners {
			wants[i], caps[i] = 0, 0
			if choices[i].Refuse {
				continue
			}
			wants[i] = b.OwnWant[k] * choices[i].Mul[k]
			caps[i] = choices[i].Cap
		}
		sent := flow.Share(spare[k], wants, caps)
		for i, b := range partners {
			if sent[i] <= 0 {
				continue
			}
			if b.From == nil {
				b.From = map[int]flow.Income{}
			}
			got := b.From[a.ID]
			got[k] += sent[i]
			b.From[a.ID] = got
			a.Tally.Sent[k] += sent[i]
			b.Tally.Got[k] += sent[i]
			if a.fed == nil {
				a.fed = map[int]bool{}
			}
			if !a.fed[b.ID] {
				a.fed[b.ID] = true
				a.Tally.Fed++
			}
		}
	}
	for _, b := range partners {
		w.cutting(a, b)
	}
}

// willing asks the mind what a people sends a partner.
func (w *World) willing(a, b *Civ) mind.TradeChoice {
	in := mind.TradeInput{
		Monster: w.monster(a, b), Grudge: a.Grudge[b.ID],
		Xenophobe: a.Has("xenophobic"), Different: difference(a.Species, b.Species) >= 1,
		Fear: a.Dials.Fear, Hostile: b.hostile(),
		Nomad: a.Aloft || b.Aloft, Drive: w.drive(a), InReach: w.partnerInReach(a, b),
	}
	if i := a.Intel[b.ID]; i != nil {
		in.Stronger = i.Mil > a.Mil
	}
	return mind.Trade(in, w.Cfg.Tuning)
}

// drive is a people's road for goods: 2 with the Door or wormholes, 1 with
// sails or near-light travel, else 0.
func (w *World) drive(c *Civ) int {
	switch {
	case c.miracle("ftl") || c.Known["wormhole_physics"]:
		return 2
	case c.Known["beamed_sails"] || c.Known["near_light"]:
		return 1
	}
	return 0
}

// partnerInReach says whether a holding of a is within a's reach of a
// holding of b.
func (w *World) partnerInReach(a, b *Civ) bool {
	for _, s := range w.holdings(b) {
		if w.inReach(a, s) {
			return true
		}
	}
	return false
}

// embargoStep keeps the books on refusal: a refusal while the partner
// wants what the refuser has spare, once it has lasted, is an embargo.
func (w *World) embargoStep(a, b *Civ, ch mind.TradeChoice, spare flow.Income) {
	wanting := false
	for k := range spare {
		if b.OwnWant[k] > 0 && spare[k] > 0 {
			wanting = true
		}
	}
	if !ch.Refuse || !wanting {
		if a.Refused[b.ID] != 0 {
			delete(a.Refused, b.ID)
		}
		if a.Embargo[b.ID] {
			delete(a.Embargo, b.ID)
			w.log("The %s open their ports to the %s again.", a.Name, b.Name)
		}
		return
	}
	if a.Refused == nil {
		a.Refused = map[int]Year{}
	}
	since, ok := a.Refused[b.ID]
	if !ok {
		a.Refused[b.ID] = w.Now
		return
	}
	if a.Embargo[b.ID] || w.Now-since < embargoYears {
		return
	}
	if a.Embargo == nil {
		a.Embargo = map[int]bool{}
	}
	a.Embargo[b.ID] = true
	w.fact(FEmbargo, a, b, -1)
	w.log("The %s have what the %s want, and will not send it. The %s call it an embargo.", a.Name, b.Name, b.Name)
	w.cutTrade(b, a, "the embargo")
}

// depend is a people learning what it hangs on: when its own income does
// not cover the fed uses, every partner that sent it something this tick
// is one it depends on.
func (w *World) depend(c *Civ) {
	c.Dependent = nil
	own := c.Income.Less(c.Received)
	if own.Covers(c.WorkingNeed) {
		return
	}
	for _, pid := range sortedInts(c.From) {
		if c.From[pid] != (flow.Income{}) {
			if c.Dependent == nil {
				c.Dependent = map[int]bool{}
			}
			c.Dependent[pid] = true
		}
	}
}

// cutTrade is a partner's sending to c stopping, by war, a pact left, a
// fall or an embargo. What was counted is taken back; if c depended on it,
// the uses it kept fed go dark now and the cut-off is a woe.
func (w *World) cutTrade(c, from *Civ, why string) {
	lost, ok := c.From[from.ID]
	if !ok {
		return
	}
	delete(c.From, from.ID)
	if !c.Dependent[from.ID] {
		return
	}
	delete(c.Dependent, from.ID)
	if c.Active() && lost != (flow.Income{}) {
		w.redirect(c, lost)
	}
	w.fact(FCutOff, c, from, c.Home)
	w.log("The %s go dark when the %s stop sending, with %s.", c.Name, from.Name, why)
}

// tradeLoss is what a people would lose in a war on a partner: what the
// partner sent it last tick.
func (w *World) tradeLoss(c, e *Civ) float64 {
	return c.From[e.ID].Total()
}
