package history

import "worldgen/internal/names"

func (w *World) spawnCiv(home int) *Civ {
	c := &Civ{
		ID: len(w.Civs), Name: names.Civ(w.R), Home: home, HomeName: names.Star(w.R),
		Born: w.Now, Temper: Temper(w.R.IntN(4)), Systems: []int{home}, Peak: 1,
		Wars: map[int]bool{}, Met: map[int]bool{},
		Faced: map[string]bool{}, Scars: map[string]bool{}, Boons: map[string]bool{},
	}
	w.Civs = append(w.Civs, c)
	w.Owner[home] = c.ID
	old := w.G.Stars[home].Name
	w.G.Stars[home].Name = c.HomeName
	prior := ""
	for _, o := range w.Civs {
		if o != c && o.Home == home {
			prior = sprintf(", among the ruins of the %s", o.Name)
		}
	}
	w.log("The %s arise on a world of %s%s. They call it %s. They are %s.", c.Name, old, prior, c.HomeName, c.Temper)
	return c
}

func (w *World) tickCivs() {
	for _, c := range w.Civs {
		if !c.Living() {
			continue
		}
		if c.Stage == Remnant {
			w.tickRemnant(c)
			continue
		}
		w.growTech(c)
		w.arrivals(c)
		w.expand(c)
		w.megastructures(c)
		w.ftl(c)
		w.war(c)
		w.faceFilters(c)
		if c.Active() {
			w.weightOfAges(c)
		}
	}
	if w.Now%(w.Cfg.FineStep*10) == 0 {
		w.contacts()
	}
}

func (w *World) growTech(c *Civ) {
	rate := map[Temper]float64{Curious: 0.012, Zealous: 0.010, Aggressive: 0.009, Insular: 0.007}[c.Temper]
	c.Tech += rate * c.techMul() * (0.5 + w.R.Float64()) / (1 + c.Tech/4)
	if c.Stage == Emergent && c.Tech >= 1 {
		c.Stage = Interstellar
		w.log("The %s reach the stars. The first slow ships leave %s.", c.Name, c.HomeName)
	}
	if c.Stage == Interstellar && len(c.Systems) >= 6 && c.Tech >= 2 {
		c.Stage = Zenith
		w.log("The %s enter their zenith: %d systems, and no rival in sight.", c.Name, len(c.Systems))
	}
}

func (w *World) arrivals(c *Civ) {
	keep := c.Voyages[:0]
	for _, v := range c.Voyages {
		if v.Arrive > w.Now {
			keep = append(keep, v)
			continue
		}
		if w.Owner[v.Target] >= 0 || w.Held[v.Target] >= 0 {
			w.trace(v.Target, "derelict colony ship", c.ID)
			w.log("A colony ship of the %s arrives at %s to find it already taken. It is never heard from again.", c.Name, w.star(v.Target))
			continue
		}
		w.Owner[v.Target] = c.ID
		c.Systems = append(c.Systems, v.Target)
		c.colonies++
		if len(c.Systems) > c.Peak {
			c.Peak = len(c.Systems)
		}
		switch n := len(c.Systems); {
		case c.colonies == 1:
			w.log("The %s settle %s, their first world beyond %s.", c.Name, w.star(v.Target), c.HomeName)
		case n == 5 || n == 10 || n == 20 || n == 40:
			w.log("The %s now hold %d systems.", c.Name, n)
		}
	}
	c.Voyages = keep
}

func (w *World) expand(c *Civ) {
	if c.Stage < Interstellar {
		return
	}
	p := min(0.3, 0.04*float64(len(c.Systems)))
	p *= map[Temper]float64{Curious: 1.1, Zealous: 1.0, Aggressive: 1.2, Insular: 0.4}[c.Temper] * c.expandMul()
	if !w.chance(p) {
		return
	}
	rng := 6 + c.Tech*4
	speed := 100.0 // years per light year, 0.01c
	if c.HasFTL {
		rng = 60
		speed = 1
	}
	from := c.Systems[w.R.IntN(len(c.Systems))]
	for _, t := range w.G.Near(from, rng) {
		if w.Owner[t] >= 0 || w.Held[t] >= 0 || w.targeted(c, t) {
			continue
		}
		d := w.G.Dist(from, t)
		c.Voyages = append(c.Voyages, Voyage{Target: t, Arrive: w.Now + Year(d*speed)})
		return
	}
}

func (w *World) targeted(c *Civ, t int) bool {
	for _, v := range c.Voyages {
		if v.Target == t {
			return true
		}
	}
	return false
}

func (w *World) megastructures(c *Civ) {
	if c.Tech < 2.2 || !w.chance(0.004) {
		return
	}
	s := c.Systems[w.R.IntN(len(c.Systems))]
	if contains(c.Enclosed, s) {
		return
	}
	c.Dyson++
	c.Enclosed = append(c.Enclosed, s)
	w.log("The %s enclose %s in a swarm of collectors. The star dims from outside.", c.Name, w.star(s))
}

// FTL is rare, and using it is not free.
func (w *World) ftl(c *Civ) {
	if c.HasFTL || c.Tech < 3 || !w.chance(0.0006) {
		return
	}
	c.HasFTL = true
	w.log("The %s tear a door in space. Faster-than-light travel is theirs.", c.Name)
	if w.chance(0.3) {
		s := c.Systems[w.R.IntN(len(c.Systems))]
		if w.chance(0.5) {
			h := w.spawnHorror(Elder, s, -1)
			h.Dormant = true
			w.log("Something on the other side of the door notices. %s now sleeps near %s.", h.Name, w.star(s))
		} else {
			h := w.spawnHorror(Beacon, s, c.ID)
			w.log("What came back through the door at %s speaks. It is called %s.", w.star(s), h.Name)
		}
	}
}

func (w *World) contacts() {
	for i, a := range w.Civs {
		if !a.Active() {
			continue
		}
		for j := i + 1; j < len(w.Civs); j++ {
			b := w.Civs[j]
			if !b.Living() || a.Met[b.ID] {
				continue
			}
			rng := 10 + max(a.Tech, b.Tech)*5
			if !w.within(a, b, rng) {
				continue
			}
			a.Met[b.ID], b.Met[a.ID] = true, true
			pw := 0.1
			for _, t := range []Temper{a.Temper, b.Temper} {
				if t == Aggressive {
					pw = max(pw, 0.6)
				} else if t == Zealous {
					pw = max(pw, 0.4)
				}
			}
			if w.chance(pw) {
				a.Wars[b.ID], b.Wars[a.ID] = true, true
				w.log("The %s and the %s find each other. The first relativistic strikes are launched within a century.", a.Name, b.Name)
			} else {
				w.log("The %s and the %s find each other. Slow messages cross the dark between them for generations.", a.Name, b.Name)
				if (a.Faced["plague"] || b.Faced["plague"]) && w.chance(0.3) {
					a.Plagued, b.Plagued = true, true
					w.log("Something crosses with the messages and the trade. Both the %s and the %s begin to sicken.", a.Name, b.Name)
				}
			}
		}
	}
}

func (w *World) within(a, b *Civ, rng float64) bool {
	for _, s := range a.Systems {
		for _, t := range b.Systems {
			if w.G.Dist(s, t) <= rng {
				return true
			}
		}
	}
	return false
}

// War between stars is slow and strange: strikes launched decades ahead,
// answered after the attacker may already be gone.
func (w *World) war(c *Civ) {
	for eid := range c.Wars {
		e := w.Civs[eid]
		if !e.Living() {
			delete(c.Wars, eid)
			continue
		}
		if w.chance(0.03) {
			delete(c.Wars, eid)
			delete(e.Wars, c.ID)
			w.log("The war between the %s and the %s ends. Neither side is sure who won.", c.Name, e.Name)
			continue
		}
		if !w.chance(0.12*c.warMul()) || len(e.Systems) == 0 {
			continue
		}
		t := e.Systems[w.R.IntN(len(e.Systems))]
		w.Bio[t] = BioNone
		if t == e.Home {
			w.log("A relativistic strike from the %s shatters %s, homeworld of the %s.", c.Name, w.star(t), e.Name)
		}
		w.loseSystem(e, t, "glassed world", sprintf("were annihilated in war with the %s", c.Name))
		if t == e.Home {
			w.endCiv(e, Extinct, sprintf("were annihilated in war with the %s", c.Name))
		} else if w.chance(0.4) {
			w.log("The %s glass %s, a world of the %s.", c.Name, w.star(t), e.Name)
		}
	}
}

func (w *World) tickRemnant(c *Civ) {
	c.Tech = max(0.5, c.Tech-0.001)
	if w.chance(0.0001) {
		w.endCiv(c, Extinct, "faded away, the last of them unremarked")
	}
}

// loseSystem removes a star from a civilisation and leaves a trace. A living
// civilisation with no worlds left is extinct; cause says why, in the
// caller's words.
func (w *World) loseSystem(c *Civ, s int, kind string, cause string) {
	if !contains(c.Systems, s) {
		return
	}
	c.Systems = remove(c.Systems, s)
	w.Owner[s] = -1
	w.trace(s, kind, c.ID)
	if c.Stage != Dead && len(c.Systems) == 0 {
		if cause == "" {
			cause = "lost their last world"
		}
		w.endCiv(c, Extinct, cause)
	}
}

// contract shrinks a civilisation to its home (or one world) as a remnant.
func (w *World) contract(c *Civ, cause string) {
	keep := c.Home
	if !contains(c.Systems, keep) && len(c.Systems) > 0 {
		keep = c.Systems[w.R.IntN(len(c.Systems))]
	}
	for _, s := range append([]int(nil), c.Systems...) {
		if s != keep {
			w.loseSystem(c, s, "abandoned colony", "")
		}
	}
	c.Stage, c.Fate, c.Cause, c.Ended = Remnant, Contracted, cause, w.Now
	c.Title = names.Title(w.R)
	c.Voyages = nil
	w.log("The %s %s. What remains of them lives on %s under %s. Once they held %s.", c.Name, cause, w.star(keep), c.Title, systems(c.Peak))
}

// endCiv finishes a civilisation as extinct or transformed. Systems become traces.
func (w *World) endCiv(c *Civ, f Fate, cause string) {
	if c.Stage == Dead {
		return
	}
	wasRemnant := c.Stage == Remnant
	c.Stage = Dead
	for _, s := range append([]int(nil), c.Systems...) {
		if f == Extinct {
			w.loseSystem(c, s, "dead cities", "")
		} else {
			w.loseSystem(c, s, "transformed world", "")
		}
	}
	for _, s := range c.Enclosed {
		w.trace(s, "dyson remnant", c.ID)
	}
	if wasRemnant {
		cause = c.Cause + ", and long after " + cause
	}
	c.Fate, c.Cause, c.Ended = f, cause, w.Now
	c.Voyages = nil
	if f == Extinct && wasRemnant {
		w.log("The last of the %s are gone from %s. They %s.", c.Name, c.HomeName, cause)
	} else if f == Extinct {
		w.log("The %s %s. They held %s at their height.", c.Name, cause, systems(c.Peak))
	}
}
