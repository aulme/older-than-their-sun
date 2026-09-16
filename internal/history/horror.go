package history

import "worldgen/internal/names"

func (w *World) spawnHorror(kind HorrorKind, origin int, fromCiv int) *Horror {
	h := &Horror{ID: len(w.Horrors), Kind: kind, Name: names.Horror(w.R), Origin: origin, Born: w.Now, FromCiv: fromCiv}
	w.Horrors = append(w.Horrors, h)
	if kind == Replicators || kind == RogueMind {
		w.horrorTake(h, origin)
	}
	return h
}

// horrorTake gives a star to a horror, evicting any civilisation there.
func (w *World) horrorTake(h *Horror, s int) {
	if s == w.G.Sol {
		return
	}
	if c := w.Owner[s]; c >= 0 {
		if h.Kind == Replicators {
			w.loseSystem(w.Civs[c], s, "stripped world", sprintf("were consumed by %s", h.Name))
		} else {
			w.loseSystem(w.Civs[c], s, "absorbed world", sprintf("were absorbed into %s", h.Name))
		}
	}
	w.Held[s] = h.ID
	h.Systems = append(h.Systems, s)
}

func (w *World) tickHorrors() {
	for _, h := range w.Horrors {
		switch h.Kind {
		case Replicators:
			w.tickReplicators(h)
		case RogueMind:
			w.tickRogueMind(h)
		case Beacon:
			w.tickBeacon(h)
		case Elder:
			w.tickElder(h)
		}
	}
}

// Replicators spread star to star, eating everything. Eventually they go
// quiet, whether by burnout, mutation, or something else.
func (w *World) tickReplicators(h *Horror) {
	if h.Dormant {
		return
	}
	if w.chance(0.0003) {
		h.Dormant = true
		w.log("%s falls silent across %d systems. Nobody knows why. The machines are still there.", h.Name, len(h.Systems))
		return
	}
	for _, s := range h.Systems {
		if !w.chance(0.02) {
			continue
		}
		for _, t := range w.G.Near(s, 12) {
			if w.Held[t] < 0 && t != w.G.Sol {
				w.horrorTake(h, t)
				break
			}
		}
	}
}

// A rogue mind expands slowly and mostly ignores everyone.
func (w *World) tickRogueMind(h *Horror) {
	for _, s := range h.Systems {
		if !w.chance(0.004) {
			continue
		}
		for _, t := range w.G.Near(s, 15) {
			if w.Held[t] < 0 && t != w.G.Sol {
				w.horrorTake(h, t)
				if w.chance(0.3) {
					w.log("%s extends itself to %s.", h.Name, w.star(t))
				}
				break
			}
		}
	}
}

// A beacon broadcasts something that destroys or converts whoever listens.
// Only civilisations advanced enough to listen, and close enough, are at risk.
func (w *World) tickBeacon(h *Horror) {
	for _, c := range w.Civs {
		if !c.Active() || c.Tech < 0.8 {
			continue
		}
		listen := 25 + c.Tech*12
		near := false
		for _, s := range c.Systems {
			if w.G.Dist(s, h.Origin) <= listen {
				near = true
				break
			}
		}
		if !near || !w.chance(0.0025) {
			continue
		}
		h.Victims++
		if w.chance(0.5) {
			w.endCiv(c, Transformed, sprintf("heard %s and were changed by it", h.Name))
			c.Into = "a cult of the signal"
			nb := w.spawnHorror(Beacon, c.Home, c.ID)
			w.log("From %s a new signal goes out, in the voice of the %s. It is called %s.", c.HomeName, c.Name, nb.Name)
		} else {
			w.endCiv(c, Extinct, sprintf("listened to %s", h.Name))
		}
	}
}

// An elder entity sleeps until someone settles too close, then unmakes
// everything nearby and goes back to sleep.
func (w *World) tickElder(h *Horror) {
	if !h.Dormant {
		h.Dormant = true
		return
	}
	disturbed := false
	for _, c := range w.Civs {
		if !c.Living() {
			continue
		}
		for _, s := range c.Systems {
			if w.G.Dist(s, h.Origin) <= 12 {
				disturbed = true
			}
		}
	}
	if !disturbed || !w.chance(0.01) {
		return
	}
	h.Dormant = false
	h.Wakings++
	w.log("%s wakes at %s.", h.Name, w.star(h.Origin))
	for _, c := range w.Civs {
		if !c.Living() {
			continue
		}
		lost := 0
		for _, s := range append([]int(nil), c.Systems...) {
			if w.G.Dist(s, h.Origin) <= 40 {
				w.loseSystem(c, s, "silent world", sprintf("were unmade by %s", h.Name))
				lost++
			}
		}
		if lost == 0 {
			continue
		}
		h.Victims++
		if !contains(c.Systems, c.Home) {
			w.endCiv(c, Extinct, sprintf("were unmade by %s", h.Name))
		} else if c.Active() {
			w.contract(c, sprintf("withdrew to %s after %s took %d of their worlds", c.HomeName, h.Name, lost))
		}
	}
	w.log("%s goes still again.", h.Name)
}
