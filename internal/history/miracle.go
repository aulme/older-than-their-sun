package history

import (
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Miracles are the powers apart from the tree: the Voice, the Flesh, the
// Door, the Unmaking, the Chorus, the Sight. Each is the dominant fact about
// whoever holds it. There are three ways in. A people can climb a deep spine
// of the tree and then make a conscious leap, which is slow, uncertain and
// dangerous. A people can find what an earlier people left of a miracle and
// master it, which is the same leap made on someone else's work. Or a
// species can be born to one, with no tree beneath it, and since what is
// evolved was never reached for there is nothing to fall from: the born
// never face the miracle's filter. A miracle held by wielding lasts as long
// as the wielded thing does.

// gain records a miracle gained by leap, find or wielding, and starts the
// surge: for a while the holder grows and learns at a rate nothing else can
// match. Gaining one also renews a people; whatever they were tired of, they
// are not tired of this.
func (w *World) gain(c *Civ, key, how string) {
	if c.Miracles[key] == "" {
		w.holdMiracle(c, key, how)
		w.fact(FMiracle, c, nil, -1).with(P{"miracle": key, "how": how})
	}
	if causal[key] {
		w.name(c, key)
	}
	if c.Ascended == 0 {
		c.Ascended = w.Now // the surge comes once
		c.Renewed = w.Now
		c.Morale += 1
	}
	w.reset(c) // whatever they were tired of, they are not tired of this
	w.recompute(c)
}

// surging is true for the few million years after a miracle is gained.
func (c *Civ) surging(now Year) bool {
	return c.Ascended > 0 && now-c.Ascended < 3_000_000 && len(c.held()) > 0
}

// miracle reports whether a civilisation holds a miracle by any route.
func (c *Civ) miracle(key string) bool {
	if c.Known[key] || c.Species.Miracle() == key {
		return true
	}
	for _, l := range c.Wielded {
		if l.Level == "miracle" && l.Node == key {
			return true
		}
	}
	return false
}

// held lists the miracles a civilisation holds now.
func (c *Civ) held() []string {
	var out []string
	for _, n := range tech.Miracles {
		if c.miracle(n.Key) {
			out = append(out, n.Key)
		}
	}
	return out
}

// leapWeight is how likely a people is to choose a miracle as its pursuit
// once the spine beneath it is climbed. It is a conscious choice, so it
// follows temperament.
func (c *Civ) leapWeight(key string) float64 {
	if len(c.held()) > 0 {
		return 0 // one miracle is a people's whole shape; nobody reaches for a second
	}
	m := 0.12 * (1 + 0.1*float64(len(c.Systems))) // once the spine is climbed the leap is the obvious next thing; the wide reach for it sooner
	if c.Has("curious") {
		m *= 2.5
	}
	if c.Has("cautious") {
		m *= 0.2
	}
	switch key {
	case "ansible":
		if c.Has("collective") || c.Species.Is(species.Hive) {
			m *= 2
		}
	case "directed_evolution":
		if c.Species.Is(species.Evolver) || c.Own >= 0 {
			m *= 3
		}
	case "ftl":
		if c.Has("expansionist") {
			m *= 2.5
		}
	case "unmaking":
		if c.Has("conqueror") || c.Has("xenophobic") {
			m *= 2.5
		}
		if c.Has("pacifist") {
			m = 0
		}
	case "chorus":
		if c.Has("contemplative") || c.Has("collective") {
			m *= 2
		}
	case "foresight":
		if c.Has("contemplative") {
			m *= 2.5
		}
	case "ember":
		m *= objectLeap // an object is a rarer shape for a people than a power
		if c.Has("pragmatic") {
			m *= 2
		}
	case "manna":
		m *= objectLeap
		if c.Has("caste") || c.Species.Is(species.Hive) {
			m *= 2
		}
		if c.Species.Profile().OrganicAsEnergy {
			m = 0 // nothing to feed
		}
	}
	return m
}

// miracleDiff is how a held miracle changes the difficulty of a filter.
func (c *Civ) miracleDiff(key string) float64 {
	d := 0.0
	if c.miracle("foresight") && key != "sight" {
		d -= 2 // they saw it coming
		if key == "cosmic" || key == "dying" {
			d -= 2
		}
	}
	if c.miracle("chorus") {
		switch key {
		case "beacon":
			d -= 10 // nothing takes root in them that they did not plant
		case "ossification", "hold":
			d -= 2
		}
	}
	if c.miracle("ansible") {
		switch key {
		case "ossification", "hold":
			d -= 1
		case "beacon":
			d += 1 // everyone hears it at once
		}
	}
	if c.miracle("directed_evolution") && (key == "cosmic" || key == "dying" || key == "overshoot") {
		d -= 2
	}
	return d
}

// warBonus is what a miracle adds to a strike roll.
func (c *Civ) warBonus() float64 {
	b := 0.0
	if c.miracle("unmaking") {
		b += 3
	}
	if c.miracle("foresight") {
		b += 2
	}
	if c.miracle("ansible") {
		b += 1.5
	}
	if c.miracle("ftl") {
		b += 1
	}
	return b
}

// holdDiff is how much harder it is to rise against a master who holds a miracle.
func (w *World) holdDiff(c *Civ) float64 {
	if c.Master < 0 {
		return 0
	}
	m := w.Civs[c.Master]
	if !m.Living() {
		return 0
	}
	d := 0.0
	if m.miracle("chorus") {
		d += 3
	}
	if m.miracle("ansible") || m.miracle("unmaking") || m.miracle("foresight") {
		d += 2
	}
	return d
}

// loseMiracle takes a leapt or found miracle away, as a filter's price.
func (w *World) loseMiracle(c *Civ, key string) {
	w.forgetNode(c, key)
	keep := c.Wielded[:0]
	for _, l := range c.Wielded {
		if l.Level == "miracle" && l.Node == key {
			w.setState(l, Lost)
			continue
		}
		keep = append(keep, l)
	}
	c.Wielded = keep
	w.recompute(c)
}

// remake turns a people into a successor species on the same worlds, the
// Flesh gone wrong or gone too far.
func (w *World) remake(c *Civ, why reason) *Civ {
	var sp *species.Species
	if w.R.Float64() < 0.5 {
		sp = species.GenerateWith(w.R, w.G.Stars[c.Home].Mult, "", species.Biological, species.Evolver)
	} else {
		sp = species.Generate(w.R, w.G.Stars[c.Home].Mult)
	}
	sp.Made = species.MadeBy("remade", c.ID)
	sp.Parent = c.Species
	worlds := append([]int(nil), c.Systems...)
	known := knownOf(c)
	home := c.Home
	w.endCiv(c, Transformed, why)
	nc := w.spawnCiv(home, sp, -1)
	nc.Master = -1
	for _, s := range worlds {
		if s != home && w.Owner[s] < 0 {
			w.setOwner(s, nc.ID)
			nc.Systems = append(nc.Systems, s)
		}
	}
	nc.Peak = len(nc.Systems)
	for _, k := range known {
		if w.R.Float64() < 0.6 || tech.Get(k).Miracle {
			w.know(nc, k)
		}
	}
	w.scar(nc, ScarChanged)
	if nc.Known["directed_evolution"] {
		w.holdMiracle(nc, "directed_evolution", "found")
		nc.Faced["brood"] = true
	}
	w.recompute(nc)
	c.Into, c.IntoCivs = "people", []int{nc.ID}
	w.event(KRemade, c, nc, -1, P{"traits": sp.TraitKeys()})
	return nc
}

func init() {
	def(&Filter{
		Key: "openline",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "openline", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarOtherVoices)
			c.Morale -= 1
			w.tear(0.3)
			w.faced(c, "openline", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			w.tear(0.6)
			if w.R.Float64() < 0.65 {
				w.loseMiracle(c, "ansible")
				w.contract(c, because("openline_closed"))
				return
			}
			home := c.Home
			w.endCiv(c, Transformed, because("one_voice"))
			c.Into = "one_voice"
			w.makeTransmitter(home, c.ID, true)
			w.faced(c, "openline", "declined", "", home)
		},
	})
	def(&Filter{
		Key: "brood",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "brood", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			sp := c.Species.Branch() // the same people, no longer quite the same blood
			w.register(sp)
			w.setSpecies(c, sp)
			t := species.Pick(w.R, "bio")
			sp.Add(t.Key)
			w.scar(c, ScarChanged)
			w.faced(c, "brood", "scarred", "", -1).P["trait"] = t.Key
		},
		Decline: func(w *World, c *Civ) {
			if w.R.Float64() < 0.5 {
				w.remake(c, because("remade"))
				return
			}
			// what they bred eats them: a swarm of flesh that makes more of itself, at one of their worlds, and the old people gone
			s := w.aWorld(c)
			w.faced(c, "brood", "declined", "", s)
			w.loseSystem(c, s, "stripped", because("brood_eaten"))
			w.endCiv(c, Extinct, because("brood_eaten"))
			if nc := w.replicatorAt(s, species.Biological, species.MadeBy("bred", c.ID), false); nc != nil {
				nc.Species.Parent = c.Species
				w.event(KNamedItself, nc, nil, s, P{"way": "", "traits": nc.Species.TraitKeys()})
			}
		},
	})
	def(&Filter{
		Key: "unmaking",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "unmaking", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			s := c.Home
			if len(c.Systems) > 1 {
				for _, x := range c.Systems {
					if x != c.Home {
						s = x
						break
					}
				}
			}
			w.scar(c, ScarBurningSky)
			w.tear(0.3)
			w.blast(s, 4, "unmaking_test", c, 1, -1)
		},
		Decline: func(w *World, c *Civ) {
			home := c.Home
			w.tear(0.6)
			w.faced(c, "unmaking", "declined", "", -1)
			w.blast(home, 6, "unmaking_inward", nil, 2, -1)
			if c.Active() && contains(c.Systems, home) {
				w.loseSystem(c, home, "unmade", because("unmade_own"))
			}
			if c.Active() {
				w.contract(c, because("unmade_own_fled"))
			}
		},
	})
	def(&Filter{
		Key: "chorus",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "chorus", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			c.Ossified, c.Stiff = true, max(c.Stiff, 1) // set, and the filter will come for it
			w.faced(c, "chorus", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			if w.R.Float64() < 0.65 {
				w.loseMiracle(c, "chorus")
				w.contract(c, because("chorus_lost"))
				return
			}
			home := c.Home
			w.endCiv(c, Transformed, because("chorus_became"))
			c.Into = "chorus"
			w.makeTransmitter(home, c.ID, true)
			w.faced(c, "chorus", "declined", "", home)
		},
	})
	def(&Filter{
		Key: "sight",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "sight", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarFatalism)
			c.Morale -= 1
			w.tear(0.3)
			w.faced(c, "sight", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			w.tear(1)
			w.contract(c, because("sight_waited"))
		},
	})
}
