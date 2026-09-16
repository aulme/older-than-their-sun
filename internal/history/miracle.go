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

// miracleNames is how the legends say each miracle.
var miracleNames = map[string]string{
	"ansible":            "the Voice, minds that speak across any distance",
	"directed_evolution": "the Flesh, a body for every world",
	"ftl":                "the Door, a way between the stars faster than light",
	"unmaking":           "the Unmaking, the end of matter at any distance",
	"chorus":             "the Chorus, thought that takes root in any mind",
	"foresight":          "the Sight, knowledge of what is coming",
}

// gain records a miracle gained by leap, find or wielding, and starts the
// surge: for a while the holder grows and learns at a rate nothing else can
// match. Gaining one also renews a people; whatever they were tired of, they
// are not tired of this.
func (w *World) gain(c *Civ, key, how string) {
	if c.Miracles[key] == "" {
		c.Miracles[key] = how
	}
	if causal[key] {
		w.name(c, key)
	}
	if c.Ascended == 0 {
		c.Ascended = w.Now // the surge comes once
		c.Renewed = w.Now
		c.Morale += 1
	}
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
		if c.Has("collective") || c.Has("hive") {
			m *= 2
		}
	case "directed_evolution":
		if c.Species.Kind == species.Evolver || c.Species.Kind == species.Parasite {
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
		case "weight", "hold":
			d -= 2
		}
	}
	if c.miracle("ansible") {
		switch key {
		case "weight", "hold":
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
	if m.Species.Kind == species.Parasite {
		// a rider is not thrown off; it is cured, and the cure is a science
		d += 4
		switch {
		case m.Has("mindrider") && c.Known["memetics"]:
			d -= 4
		case m.Has("mindrider") && c.Known["religion"]:
			d -= 1 // the old prayers turn out to be worth something
		case !m.Has("mindrider") && c.Known["medicine"] && c.Known["genetics"]:
			d -= 4
		}
	}
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
	delete(c.Known, key)
	keep := c.Wielded[:0]
	for _, l := range c.Wielded {
		if l.Level == "miracle" && l.Node == key {
			l.State = Lost
			continue
		}
		keep = append(keep, l)
	}
	c.Wielded = keep
	w.recompute(c)
}

// remake turns a people into a successor species on the same worlds, the
// Flesh gone wrong or gone too far.
func (w *World) remake(c *Civ, why string) *Civ {
	sp := species.Generate(w.R, w.G.Stars[c.Home].Mult)
	if w.R.Float64() < 0.5 {
		sp.Kind = species.Evolver
	}
	sp.Made = "what the " + c.Name + " made of themselves"
	worlds := append([]int(nil), c.Systems...)
	known := knownOf(c)
	home := c.Home
	w.endCiv(c, Transformed, why)
	nc := w.spawnCiv(home, sp, -1)
	nc.Master = -1
	for _, s := range worlds {
		if s != home && w.Owner[s] < 0 {
			w.Owner[s] = nc.ID
			nc.Systems = append(nc.Systems, s)
		}
	}
	nc.Peak = len(nc.Systems)
	for _, k := range known {
		if w.R.Float64() < 0.6 || tech.Get(k).Miracle {
			nc.Known[k] = true
		}
	}
	nc.Scars[ScarChanged] = true
	if nc.Known["directed_evolution"] {
		nc.Miracles["directed_evolution"] = "found"
		nc.Faced["brood"] = true
	}
	w.recompute(nc)
	c.Into = "the " + nc.Name
	w.log("The %s are gone. What they made of themselves holds their worlds and calls itself the %s: %s.", c.Name, nc.Name, sp.Describe())
	return nc
}

func init() {
	def(&Filter{
		Key: "openline", Name: "the Open Line", Levels: []string{"soc"}, Diff: 6, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			w.log("The line carries only the voices of the %s. They keep it that way.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarOtherVoices] = true
			c.Morale -= 1
			w.tear(0.3)
			w.log("There are other voices on the line, older, and some of the %s listen. They never quite stop.", c.Name)
		},
		Decline: func(w *World, c *Civ) {
			w.tear(0.6)
			if w.R.Float64() < 0.65 {
				w.loseMiracle(c, "ansible")
				w.contract(c, "heard what else was on the line, closed it, and forgot how to open it")
				return
			}
			home := c.Home
			w.endCiv(c, Transformed, "became one voice")
			c.Into = "one voice"
			h := w.spawnHorror(Beacon, home, c.ID)
			w.log("The %s stop being many. From %s one voice goes out that used to be all of theirs. It is called %s.", c.Name, w.star(home), h.Name)
		},
	})
	def(&Filter{
		Key: "brood", Name: "the Brood", Levels: []string{"soc"}, Diff: 6, Domain: "biology",
		Overcome: func(w *World, c *Civ) {
			w.log("The %s change, and stay themselves. It is a matter of law with them what may not be altered.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			sp := *c.Species
			sp.Traits = append([]*species.Trait(nil), c.Species.Traits...)
			c.Species = &sp
			t := species.Pick(w.R, "bio")
			sp.Add(t.Key)
			c.Scars[ScarChanged] = true
			w.log("The %s come out the other side of the change %s. They did not mean to.", c.Name, t.Name)
		},
		Decline: func(w *World, c *Civ) {
			w.remake(c, "remade themselves once too often")
		},
	})
	def(&Filter{
		Key: "unmaking", Name: "the Unmaking", Levels: []string{"soc"}, Diff: 6.5, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			w.log("The %s build it and do not use it. Everyone within reach knows they have it. That is enough.", c.Name)
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
			c.Scars[ScarBurningSky] = true
			w.tear(0.3)
			w.blast(s, 4, "a test of the Unmaking", "The first test of the %s' Unmaking takes a world with it.", 1)
		},
		Decline: func(w *World, c *Civ) {
			home := c.Home
			w.tear(0.6)
			w.log("The %s turn the Unmaking on something too close.", c.Name)
			w.blast(home, 6, "the Unmaking turned inward", "Everything around %s stops being matter for a while.", 2)
			if c.Active() && contains(c.Systems, home) {
				w.loseSystem(c, home, "unmade world", "unmade their own world")
			}
			if c.Active() {
				w.contract(c, "unmade their own world and fled what was left")
			}
		},
	})
	def(&Filter{
		Key: "chorus", Name: "the Chorus", Levels: []string{"soc"}, Diff: 6, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			w.log("The %s think one thought and remain many people. It can be done.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarOssified] = true
			w.log("The %s think one thought, and it is the same thought every year after. Nothing new is ever said.", c.Name)
		},
		Decline: func(w *World, c *Civ) {
			if w.R.Float64() < 0.65 {
				w.loseMiracle(c, "chorus")
				w.contract(c, "lost themselves in their own chorus")
				return
			}
			home := c.Home
			w.endCiv(c, Transformed, "became the thought they were thinking")
			c.Into = "a chorus"
			h := w.spawnHorror(Beacon, home, c.ID)
			w.log("The thought of the %s gets loose. From %s it goes out to whoever will hear it. It is called %s.", c.Name, w.star(home), h.Name)
		},
	})
	def(&Filter{
		Key: "sight", Name: "the Sight", Levels: []string{"soc"}, Diff: 6, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			w.log("The %s see how it ends, and go on anyway.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarFatalism] = true
			c.Morale -= 1
			w.tear(0.3)
			w.log("The %s see how it ends. A fatalism settles on them that never lifts.", c.Name)
		},
		Decline: func(w *World, c *Civ) {
			w.tear(1)
			w.contract(c, "saw what was coming and sat down to wait for it")
		},
	})
}
