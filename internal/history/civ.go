package history

import (
	"worldgen/internal/names"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// spawnCiv raises a civilisation at a star. sp is nil for a natural species;
// made species pass their own and the maker's id.
func (w *World) spawnCiv(home int, sp *species.Species, maker int) *Civ {
	st := &w.G.Stars[home]
	sys := w.G.Sys[home]
	sys.EnsureHome(w.R, st)
	if sp == nil {
		sp = species.GenerateOn(w.R, st.Mult, sys.Arch)
		if w.Law.Glare > 3 && !sp.Has("hardy") && !sp.Has("skyless") {
			sp.Add("hardy") // born under a hard sky
		}
	}
	c := &Civ{
		ID: len(w.Civs), Name: sp.Name, Species: sp, Home: home, HomeName: names.Star(w.R),
		Cradle: home, Born: w.Now, Renewed: w.Now, Systems: []int{home}, Peak: 1, Master: maker,
		Known: map[string]bool{}, Focus: map[string]float64{}, Locked: map[string]bool{},
		Structures: map[string]int{}, Found: map[int]bool{}, Heard: map[int]bool{},
		Wars: map[int]bool{}, Met: map[int]bool{}, Trade: map[int]bool{},
		Faced: map[string]bool{}, Scars: map[string]bool{}, Boons: map[string]bool{}, Miracles: map[string]string{},
	}
	if st.Real {
		c.HomeName = st.Name // a real star keeps the name Earth knows it by
	}
	for _, d := range sp.World.Locked {
		c.Locked[d] = true
	}
	w.Civs = append(w.Civs, c)
	w.Owner[home] = c.ID
	old := st.Name
	if !st.Real {
		st.Name = c.HomeName
	}
	c.CradleName = c.HomeName
	prior := ""
	for _, o := range w.Civs {
		if o != c && o.Home == home {
			prior = sprintf(", among the ruins of the %s", o.Name)
		}
	}
	w.recompute(c)
	if maker < 0 {
		w.log("The %s arise on %s, %s, around %s%s, %.0f ly from %s. They are %s.",
			c.Name, sys.HomeName(c.HomeName), sp.World.Desc, w.starDetail(home, old), prior, w.G.FromCentre(home), w.G.Anchor(), sp.Describe())
		w.log("%s", w.systemLine(home))
	}
	if f := sp.Kind.Flavour(); f.Portrait != "" {
		w.log("%s", f.Portrait)
	}
	if m := sp.Miracle(); m != "" {
		c.Miracles[m] = "born"             // the surge begins when they can first use it: see tickCivs
		c.Faced[tech.Get(m).Filter] = true // what is evolved is not a leap; nothing to fall from
		w.log("They are born to a miracle: %s. What others will spend ages reaching for, they have from the first.", miracleNames[m])
	}
	if st.Failing {
		w.log("Their sun is already failing. They were born under a dying star.")
	}
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
		if len(c.Systems) == 0 {
			panic(sprintf("active civ %s with no worlds: record %v, cause %q, last events: %v", c.Name, c.Record, c.Cause, w.Events[len(w.Events)-4:]))
		}
		w.recompute(c)
		if c.Ascended == 0 && c.Reach >= 1 && len(c.held()) > 0 {
			c.Ascended = w.Now // the born reach the stars, and the miracle begins to matter
		}
		for _, step := range []func(*Civ){w.arrivals, w.research, w.expand, w.build, w.dyingSun, w.find, w.war, w.revolt, w.ambientFilters, w.uplift} {
			if !c.Active() {
				break
			}
			step(c)
		}
		if c.Active() {
			// morale drifts back toward zero in peace
			if len(c.Wars) == 0 {
				c.Morale *= 1 - 0.02*w.dt
			}
			for d, f := range c.Focus {
				c.Focus[d] = 1 + (f-1)*(1-0.01*w.dt)
			}
		}
	}
	if w.count(0.1) > 0 {
		w.contacts()
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
			w.trace(v.Target, "derelict "+c.Species.Kind.Flavour().Ship, c.ID)
			w.log("A %s of the %s arrives at %s to find it already taken. It is never heard from again.", c.Species.Kind.Flavour().Ship, c.Name, w.star(v.Target))
			continue
		}
		w.settle(c, v.Target)
	}
	c.Voyages = keep
}

func (w *World) settle(c *Civ, t int) {
	w.Owner[t] = c.ID
	c.Systems = append(c.Systems, t)
	c.colonies++
	w.takeOver(c, t)
	if len(c.Systems) > c.Peak {
		c.Peak = len(c.Systems)
	}
	switch n := len(c.Systems); {
	case c.colonies == 1:
		w.log("The %s settle %s, their first %s beyond %s.", c.Name, w.star(t), c.Species.Kind.Flavour().Colony, c.HomeName)
	case n == 5 || n == 10 || n == 20 || n == 40:
		w.log("The %s now hold %d systems.", c.Name, n)
	}
	if len(c.Systems) >= 6 && c.Era >= 3 && c.Stage == Interstellar {
		c.Stage = Zenith
		w.log("The %s enter their zenith: %d systems, and no rival in sight.", c.Name, len(c.Systems))
	}
}

// canLive says whether a star is inside the civilisation's habitable envelope.
func (w *World) canLive(c *Civ, t int) bool {
	s := &w.G.Stars[t]
	if s.Dead() {
		return c.Envelope >= 3
	}
	return s.Hostility() <= c.Envelope && !w.mindDead(t)
}

// mindDead is true inside a "region where minds do not work" law.
func (w *World) mindDead(t int) bool {
	for _, l := range w.Legacies {
		if l.Kind == Law && l.Desc == lawDescs[0] && l.State != Mastered && w.G.Dist(l.Star, t) <= 8 {
			return true
		}
	}
	return false
}

// expand launches colony ships within reach. A target must be within reach
// of home and within a ship's hop of a held system.
func (w *World) expand(c *Civ) {
	if !c.Free() && !c.Vassal {
		return
	}
	// necessity: a people with nowhere to go works on ships, whatever else it was doing
	if c.Era >= 2 && c.Reach < 40 && w.nothingNear(c) {
		c.Focus[tech.Propulsion] = max(c.Focus[tech.Propulsion], 4)
		if p := tech.Get(c.Pursuit); p == nil || (p.Domain != tech.Propulsion && !p.Miracle) {
			if k := w.cheapest(c, tech.Propulsion); k != "" {
				c.Pursuit = k
			}
		}
	}
	if c.Reach < 1 {
		return
	}
	p := min(0.3, 0.04*float64(len(c.Systems))) * c.expandMul(w)
	if !w.chance(p) {
		return
	}
	hop := min(c.Reach, 20)
	if c.miracle("ftl") {
		hop = c.Reach // a door does not care how far
	}
	from := w.pick(c.Systems)
	for _, t := range w.G.Near(from, hop) {
		if w.Owner[t] >= 0 || w.Held[t] >= 0 || w.targeted(c, t) || w.G.Dist(c.Home, t) > c.Reach || !w.canLive(c, t) {
			continue
		}
		d := w.G.Dist(from, t)
		c.Voyages = append(c.Voyages, Voyage{Target: t, Arrive: w.Now + Year(d*c.Speed)})
		return
	}
}

// cheapest returns the cheapest node of a domain open to a people, or "".
func (w *World) cheapest(c *Civ, domain string) string {
	best := ""
	for _, n := range tech.Nodes {
		if n.Domain == domain && !n.Miracle && w.canPursue(c, n) && (best == "" || n.Price() < tech.Get(best).Price()) {
			best = n.Key
		}
	}
	return best
}

// nothingNear is true when no star a people could live on lies within reach.
func (w *World) nothingNear(c *Civ) bool {
	for _, t := range w.G.Near(c.Home, max(c.Reach, 1)) {
		if w.Owner[t] < 0 && w.Held[t] < 0 && w.canLive(c, t) {
			return false
		}
	}
	return true
}

func (w *World) targeted(c *Civ, t int) bool {
	for _, v := range c.Voyages {
		if v.Target == t {
			return true
		}
	}
	return false
}

// build raises a structure within reach. Great works are few: a people
// raises one every few hundred thousand years, and at most two of a kind.
func (w *World) build(c *Civ) {
	if !w.chance(0.004) {
		return
	}
	var can []string
	for k := range c.Known {
		if s := tech.Get(k).Structure; s != "" && c.Structures[s] < 2 {
			can = append(can, s)
		}
	}
	if len(can) == 0 {
		return
	}
	key := can[w.R.IntN(len(can))]
	st := tech.Structures[key]
	s := w.pick(c.Systems)
	node := ""
	for k := range c.Known {
		if tech.Get(k).Structure == key {
			node = k
		}
	}
	if key == "dyson" {
		for _, wk := range c.Works {
			if wk.Key == "dyson" && wk.Star == s {
				return
			}
		}
	}
	c.Works = append(c.Works, Work{Key: key, Node: node, Star: s, Legacy: -1})
	c.Structures[key]++
	if c.Structures[key] == 1 || key == "dyson" {
		if key == "shipyard" {
			w.log(st.Text, w.star(s), c.Name)
		} else {
			w.log(st.Text, c.Name, w.star(s))
		}
	}
}

func (w *World) tickRemnant(c *Civ) {
	if w.chance(0.00003) {
		w.endCiv(c, Extinct, "faded away, the last of them unremarked")
	}
}

// loseSystem removes a star from a civilisation and leaves a trace. A living
// civilisation with no worlds left is extinct; cause says why.
func (w *World) loseSystem(c *Civ, s int, kind string, cause string) {
	if !contains(c.Systems, s) {
		return
	}
	c.Systems = remove(c.Systems, s)
	w.Owner[s] = -1
	w.trace(s, kind, c.ID)
	keep := c.Works[:0]
	for _, wk := range c.Works {
		if wk.Star == s {
			c.Structures[wk.Key]--
			w.leaveRuin(c, wk, kind)
		} else {
			keep = append(keep, wk)
		}
	}
	c.Works = keep
	if c.Stage != Dead && len(c.Systems) == 0 {
		if cause == "" {
			cause = "lost their last world"
		}
		w.endCiv(c, Extinct, cause)
		return
	}
	if s == c.Home && c.Stage != Dead {
		w.reseat(c)
	}
}

// reseat moves the home to the nearest remaining world after the old one is lost.
func (w *World) reseat(c *Civ) {
	old := c.Home
	best, bd := -1, 1e9
	for _, x := range c.Systems {
		if d := w.G.Dist(old, x); d < bd {
			best, bd = x, d
		}
	}
	if best < 0 {
		return
	}
	c.Home = best
	c.HomeName = w.star(best)
	c.Dying = false
	w.log("What is left of the %s gathers on %s. It is home now.", c.Name, c.HomeName)
}

// contract shrinks a civilisation to its home (or one world) as a remnant.
func (w *World) contract(c *Civ, cause string) {
	keep := c.Home
	if !contains(c.Systems, keep) && len(c.Systems) > 0 {
		keep = w.pick(c.Systems)
	}
	for _, s := range append([]int(nil), c.Systems...) {
		if s != keep {
			w.loseSystem(c, s, "abandoned "+c.Species.Kind.Flavour().Colony, "")
		}
	}
	c.Stage, c.Fate, c.Cause, c.Ended = Remnant, Contracted, cause, w.Now
	c.Fell = w.Now
	c.Title = names.Title(w.R)
	c.Voyages = nil
	c.Wars = map[int]bool{}
	w.dropWielded(c, 0.5)
	w.log("The %s %s. What remains of them lives on %s under %s. Once they held %s.", c.Name, cause, w.star(keep), c.Title, systems(c.Peak))
}

// endCiv finishes a civilisation as extinct or transformed. Systems become traces.
func (w *World) endCiv(c *Civ, f Fate, cause string) {
	if c.Stage == Dead {
		return
	}
	wasRemnant := c.Stage == Remnant
	c.Stage = Dead
	if f == Extinct && len(c.Systems) > 0 && w.R.Float64() < 0.4 {
		w.leaveRelic(c, w.lateNode(c), c.Home)
	}
	for _, s := range append([]int(nil), c.Systems...) {
		if f == Extinct {
			w.loseSystem(c, s, "dead cities", "")
		} else {
			w.loseSystem(c, s, "transformed world", "")
		}
	}
	if wasRemnant {
		cause = c.Cause + ", and long after " + cause
	}
	c.Fate, c.Cause, c.Ended = f, cause, w.Now
	if !wasRemnant {
		c.Fell = w.Now
	}
	c.Voyages = nil
	c.Wars = map[int]bool{}
	w.dropWielded(c, 1)
	if f == Extinct && wasRemnant {
		w.log("The last of the %s are gone from %s. They %s.", c.Name, c.HomeName, cause)
	} else if f == Extinct {
		w.log("The %s %s. They held %s at their height.", c.Name, cause, systems(c.Peak))
	}
}

// darkAge is a non-terminal decline: tech and reach are lost. A third one is fatal.
func (w *World) darkAge(c *Civ, why string) {
	if !c.Active() {
		return
	}
	c.DarkAges++
	c.Morale -= 1
	c.Voyages = nil
	if w.wreck == nil {
		w.wreck = &defaultWreckage
		defer func() { w.wreck = nil }()
	}
	forgotten := w.forget(c, 0.3)
	// what is forgotten is not always destroyed: a relic of the lost art may wait at home
	if len(forgotten) > 0 && w.R.Float64() < 0.6 {
		best := forgotten[0]
		for _, k := range forgotten {
			if tech.Get(k).Era > tech.Get(best).Era {
				best = k
			}
		}
		w.leaveRelic(c, best, c.Home)
	}
	lost := 0
	for _, s := range append([]int(nil), c.Systems...) {
		if s != c.Home && w.R.Float64() < 0.5 {
			w.loseSystem(c, s, "abandoned "+c.Species.Kind.Flavour().Colony, "")
			lost++
		}
	}
	w.dropWielded(c, 0.5)
	w.recompute(c)
	if c.Reach < 10 {
		c.Stage = Emergent
	} else if c.Stage == Zenith {
		c.Stage = Interstellar
	}
	if c.DarkAges >= 3 {
		w.endCiv(c, Extinct, why+", and a third dark age was one too many")
		return
	}
	if lost > 0 {
		w.log("The %s %s. A dark age follows. %d %ss go silent.", c.Name, why, lost, c.Species.Kind.Flavour().Colony)
	} else {
		w.log("The %s %s. A dark age follows.", c.Name, why)
	}
}

// forget drops a fraction of known nodes, leaves first, so the tree stays
// consistent. It returns what was forgotten.
func (w *World) forget(c *Civ, frac float64) []string {
	var forgotten []string
	n := int(float64(len(c.Known))*frac + 0.5)
	for i := 0; i < n; i++ {
		var leaves []string
		for k := range c.Known {
			if tech.Get(k).Era == 0 {
				continue
			}
			leaf := true
			for o := range c.Known {
				for _, p := range tech.Get(o).Prereqs {
					if p == k {
						leaf = false
					}
				}
			}
			if leaf {
				leaves = append(leaves, k)
			}
		}
		if len(leaves) == 0 {
			break
		}
		k := leaves[w.R.IntN(len(leaves))]
		delete(c.Known, k)
		forgotten = append(forgotten, k)
	}
	return forgotten
}

func (w *World) schism(c *Civ) {
	if c.Has("hive") {
		w.log("The %s cannot split; a hive has no factions. The pressure goes elsewhere.", c.Name)
		c.Morale -= 1
		return
	}
	if len(c.Systems) < 2 {
		w.log("Unrest among the %s on %s. It passes, this time.", c.Name, c.HomeName)
		c.Morale -= 0.5
		return
	}
	lost := len(c.Systems) / 2
	var gone []int
	for i := 0; i < lost; i++ {
		s := w.pick(c.Systems)
		if s != c.Home {
			w.loseSystem(c, s, "abandoned "+c.Species.Kind.Flavour().Colony, "")
			gone = append(gone, s)
		}
	}
	// a branch of the people goes its own way, if the split was clean
	if len(gone) > 0 && w.R.Float64() < 0.4 && !c.Has("hive") {
		sp := *c.Species
		sp.Name = names.Civ(w.R)
		sp.Traits = append([]*species.Trait(nil), c.Species.Traits...)
		sp.Add("branch")
		sp.Made = "a branch of the " + c.Name
		nc := w.spawnCiv(gone[0], &sp, -1)
		nc.Master = -1
		for k := range c.Known {
			nc.Known[k] = true
		}
		w.forget(nc, 0.2)
		w.recompute(nc)
		w.log("Schism among the %s. Half their worlds go dark, and at %s the %s declare themselves a new people.", c.Name, w.star(gone[0]), nc.Name)
		return
	}
	w.log("Schism among the %s. Half their worlds go dark or go their own way.", c.Name)
}

func (c *Civ) expandMul(w *World) float64 {
	m := 1.0
	if c.Has("expansionist") {
		m *= 1.3
	}
	if c.Has("contemplative") || c.Has("cautious") {
		m *= 0.7
	}
	if c.Scars[ScarStewardship] {
		m *= 0.6
	}
	if c.Scars[ScarCentralism] {
		m *= 0.7
	}
	if c.Scars[ScarQuarantine] {
		m *= 0.5
	}
	if c.Boons[BoonSwarm] {
		m *= 1.4
	}
	if c.Scars[ScarFatalism] {
		m *= 0.6
	}
	if c.miracle("directed_evolution") {
		m *= 2
	}
	if c.miracle("ftl") {
		m *= 1.5
	}
	if c.surging(w.Now) {
		m *= 2.5
	}
	if c.Dying {
		m *= 3
	}
	return m
}
