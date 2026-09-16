package history

import "worldgen/internal/names"

func (w *World) spawnHorror(kind HorrorKind, origin int, fromCiv int) *Horror {
	h := &Horror{ID: len(w.Horrors), Kind: kind, Name: names.Horror(w.R), Origin: origin, Born: w.Now, FromCiv: fromCiv, Legacy: -1}
	w.Horrors = append(w.Horrors, h)
	if kind == Replicators || kind == RogueMind {
		w.horrorTake(h, origin)
	}
	return h
}

// horrorTake gives a star to a horror, evicting any civilisation there.
// A civilisation may hold its world by force.
func (w *World) horrorTake(h *Horror, s int) {
	if s == w.G.Sol || w.Held[s] >= 0 {
		return
	}
	if cid := w.Owner[s]; cid >= 0 {
		c := w.Civs[cid]
		if c.Active() && h.Kind != RogueMind || c.Active() && h.FromCiv != c.ID {
			w.incursionAt = s
			w.incursionBy = h
			if w.face(c, "incursion", 0) == Overcome {
				return
			}
			if !contains(c.Systems, s) {
				return // resolved otherwise
			}
		}
		if h.Kind == Replicators {
			w.loseSystem(c, s, "stripped world", sprintf("were consumed by %s", h.Name))
		} else {
			w.loseSystem(c, s, "absorbed world", sprintf("were absorbed into %s", h.Name))
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
		case SleeperHorror:
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
	if w.chance(0.0005) {
		h.Dormant = true
		w.log("%s falls silent across %d systems. Nobody knows why. The machines are still there.", h.Name, len(h.Systems))
		return
	}
	for _, s := range append([]int(nil), h.Systems...) {
		if !w.chance(0.005) {
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
	if len(h.Systems) >= 8 {
		return // it has what it needs
	}
	for _, s := range append([]int(nil), h.Systems...) {
		if !w.chance(0.004) {
			continue
		}
		for _, t := range w.G.Near(s, 15) {
			if w.Held[t] < 0 && t != w.G.Sol {
				w.horrorTake(h, t)
				if w.R.Float64() < 0.3 {
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
	if h.Dormant {
		return
	}
	for _, c := range w.Civs {
		if !c.Active() || c.Era < 2 {
			continue
		}
		listen := 25 + float64(c.Era)*10
		near := false
		for _, s := range c.Systems {
			if w.G.Dist(s, h.Origin) <= listen {
				near = true
				break
			}
		}
		if !near || c.Heard[h.ID] || !w.chance(0.001) {
			continue
		}
		c.Heard[h.ID] = true
		w.beacon = h
		w.face(c, "beacon", 0)
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
	if !disturbed || w.Now < h.Sleep || !w.chance(0.003) {
		return
	}
	w.wakeElder(h)
}

func (w *World) wakeElder(h *Horror) {
	h.Dormant = false
	h.Wakings++
	h.Sleep = w.Now + Year(500_000+w.R.Float64()*2_500_000)
	w.log("%s wakes at %s.", h.Name, w.star(h.Origin))
	for _, c := range w.Civs {
		if !c.Living() {
			continue
		}
		var worlds []int
		for _, s := range c.Systems {
			if w.G.Dist(s, h.Origin) <= 30 {
				worlds = append(worlds, s)
			}
		}
		if len(worlds) == 0 {
			continue
		}
		h.Victims++
		if !c.Active() {
			for _, s := range worlds {
				w.loseSystem(c, s, "silent world", sprintf("were unmade by %s", h.Name))
			}
			continue
		}
		w.blastWorlds = worlds
		w.blastWhat = h.Name
		adj := 1.0
		if c.Reach >= 12 {
			adj = 0
		}
		w.face(c, "elder", adj)
	}
	w.log("%s goes still again.", h.Name)
}

func init() {
	def(&Filter{
		Key: "beacon", Name: "the Signal", Levels: []string{"soc"}, Diff: 5.5, Repeat: true, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			w.log("The %s hear %s, and do not answer, and forbid anyone to listen again.", c.Name, w.beacon.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarSignal] = true
			c.Morale -= 1
			w.beacon.Victims++
			w.log("Some of the %s hear %s and are changed. A cult of the signal grows among them and is never quite rooted out.", c.Name, w.beacon.Name)
		},
		Decline: func(w *World, c *Civ) {
			h := w.beacon
			h.Victims++
			if w.R.Float64() < 0.3 {
				w.endCiv(c, Transformed, sprintf("heard %s and were changed by it", h.Name))
				c.Into = "a cult of the signal"
				nb := w.spawnHorror(Beacon, c.Home, c.ID)
				w.log("From %s a new signal goes out, in the voice of the %s. It is called %s.", c.HomeName, c.Name, nb.Name)
			} else {
				w.endCiv(c, Extinct, sprintf("listened to %s", h.Name))
			}
		},
	})
	def(&Filter{
		Key: "incursion", Name: "the machines at the door", Levels: []string{"mil"}, Diff: 5.5, Repeat: true, Domain: "weapons",
		Overcome: func(w *World, c *Civ) {
			if w.R.Float64() < 0.3 {
				w.log("%s reaches %s and is burned off it by the %s.", w.incursionBy.Name, w.star(w.incursionAt), c.Name)
			}
		},
		Scar: func(w *World, c *Civ) {
			c.Morale -= 0.5
		},
		Decline: func(w *World, c *Civ) {
			c.Morale -= 1
		},
	})
	def(&Filter{
		Key: "elder", Name: "the waking", Levels: []string{"sur"}, Diff: 5, Repeat: true, Domain: "propulsion",
		Overcome: func(w *World, c *Civ) {
			w.log("The %s hide in the deep places while %s passes over them. It does not notice.", c.Name, w.blastWhat)
		},
		Scar: func(w *World, c *Civ) {
			for _, s := range w.blastWorlds {
				if s != c.Home {
					w.loseSystem(c, s, "silent world", "")
				}
			}
			w.log("The %s lose every world near %s but their own.", c.Name, w.blastWhat)
		},
		Decline: func(w *World, c *Civ) {
			homeHit := contains(w.blastWorlds, c.Home)
			for _, s := range w.blastWorlds {
				if s != c.Home {
					w.loseSystem(c, s, "silent world", sprintf("were unmade by %s", w.blastWhat))
				}
			}
			if !c.Active() {
				return
			}
			if homeHit {
				if len(c.Systems) > 1 && c.Reach >= 12 {
					w.leaveHome(c, "flee "+w.blastWhat)
				} else {
					w.loseSystem(c, c.Home, "silent world", sprintf("were unmade by %s", w.blastWhat))
				}
			} else {
				w.contract(c, sprintf("withdrew to %s after %s took their worlds", c.HomeName, w.blastWhat))
			}
		},
	})
}
