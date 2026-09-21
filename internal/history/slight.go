package history

import "worldgen/internal/mind"

// The slight: a war on a people's partner is a wrong done to that people,
// sized by the trade between them, taken as a grudge on the attacker and
// read by everything that reads grudges. It repeats while the war starves
// the partner, and a world taken that held a source the partner drew on
// is a slight of its own. Before a war the council sums the slights it
// would give the peoples whose opinion it minds, and the bar rises by
// them. Blocs are what follows: nothing declared, only the partner graph
// as the slights make it expensive to attack from inside.

// slightOf is the wrong p would take from a war on b.
func (w *World) slightOf(p, b *Civ) float64 {
	return mind.Slight(mind.SlightInput{Flow: p.From[b.ID].Total() + b.From[p.ID].Total(), Income: p.Income.Total(), Dependent: p.Dependent[b.ID]}, w.Cfg.Tuning)
}

// slighted is the declaration's wrong: every partner of the target takes
// the slight as a grudge on the attacker, and the war remembers what the
// target was sending each, for the tick.
func (w *World) slighted(a, b *Civ, wr *War) {
	for _, pid := range sortedInts(b.Trade) {
		p := w.Civs[pid]
		if p == a || !p.Active() || !w.perceives(p, a) {
			continue // no wrong can be taken from what cannot be held in mind
		}
		s := w.slightOf(p, b)
		if s <= 0 {
			continue
		}
		wr.Slights[pid] = s
		wr.Sent[pid] = p.From[b.ID].Total()
		w.takeSlight(wr, p, a, b, s)
	}
}

// takeSlight adds a slight to a people's grudge, capped per war, and
// writes the fact once it is past the bar.
func (w *World) takeSlight(wr *War, p, a, b *Civ, s float64) {
	t := &w.Cfg.Tuning.Slight
	s = min(s, t.Cap-wr.Slighted[p.ID])
	if s <= 0 {
		return
	}
	wr.Slighted[p.ID] += s
	p.resent(a.ID, s)
	p.Tally.Slights += s
	if !wr.SlightTold[p.ID] && wr.Slighted[p.ID] >= t.Fact {
		wr.SlightTold[p.ID] = true
		w.told(FSlight, a, p, -1).with(P{"partner": b.ID})
	}
}

// slightTick is the war's each tick: a partner the target is still
// sending less than before the war takes a share of its slight again.
func (w *World) slightTick(wr *War) {
	a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
	for _, pid := range sortedInts(wr.Slights) {
		p := w.Civs[pid]
		if !p.Active() || p.From[b.ID].Total() >= wr.Sent[pid] {
			continue
		}
		w.takeSlight(wr, p, a, b, wr.Slights[pid]*w.Cfg.Tuning.Slight.Tick)
	}
}

// sourceSlight is a world taken from b that held a source a partner drew
// on: a slight of its own to that partner.
func (w *World) sourceSlight(wr *War, c, e *Civ, t int) {
	for _, pid := range sortedInts(e.Trade) {
		p := w.Civs[pid]
		if p == c || !p.Active() {
			continue
		}
		for _, h := range w.rarities(p) {
			if h.via == e.ID && h.s.Star == t {
				w.takeSlight(wr, p, c, e, w.Cfg.Tuning.Slight.Source)
				break
			}
		}
	}
}

// offence is what the council would wrong by a war on e: each partner of
// e the attacker minds, by its care, its slight and the attacker's own
// stake in it.
func (w *World) offence(c, e *Civ) []mind.Slighted {
	var out []mind.Slighted
	for _, pid := range sortedInts(e.Trade) {
		p := w.Civs[pid]
		if p == c || !p.Active() {
			continue
		}
		strong := false
		if i := c.Intel[pid]; i != nil && i.Mil >= c.Mil {
			strong = w.partnerInReach(c, p)
		}
		care := mind.Care(mind.CareInput{Partner: c.Trade[pid], Allied: w.allied(c, p), Strong: strong, Ruler: c.Master == pid || p.Master == c.ID,
			Hates: c.hates(p), Monster: w.monster(c, p), Fear: c.Dials.Fear}, w.Cfg.Tuning)
		if care == 0 {
			continue
		}
		s := w.slightOf(p, e)
		if s <= 0 {
			continue
		}
		stake := 1 + c.From[pid].Total()/max(c.Income.Total(), 1)
		out = append(out, mind.Slighted{Name: p.Tok(), Slight: care * s * stake})
	}
	return out
}
