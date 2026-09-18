package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/names"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// spawnCiv raises a civilisation at a star. sp is nil for a natural species;
// made species pass their own and the maker's id; a people that shares a
// species already in the world (a branch) passes its own name, else "" for
// the species' name.
func (w *World) spawnCiv(home int, sp *species.Species, maker int, name string) *Civ {
	st := &w.G.Stars[home]
	sys := w.G.Sys[home]
	sys.EnsureHome(w.R, st)
	if sp == nil {
		sp = species.GenerateOn(w.R, st.Mult, sys.Arch)
		if w.Law.Glare > 3 && !sp.Has("hardy") && !sp.Has("skyless") {
			sp.Add("hardy") // born under a hard sky
		}
	}
	shared := w.register(sp)
	if name == "" {
		name = sp.Name
	}
	c := &Civ{
		ID: len(w.Civs), Name: name, Species: sp, Home: home, HomeName: names.Star(w.R),
		Cradle: home, Born: w.Now, Renewed: w.Now, Systems: []int{home}, Peak: 1, Master: maker,
		Known: map[string]bool{}, Learned: map[string]Year{}, Focus: map[string]float64{}, Locked: map[string]bool{},
		Structures: map[string]int{}, Found: map[int]bool{}, Heard: map[int]bool{},
		Wars: map[int]bool{}, Met: map[int]bool{}, Reached: map[int]bool{}, Trade: map[int]bool{},
		Faced: map[string]bool{}, Scars: map[string]bool{}, Boons: map[string]bool{}, Miracles: map[string]string{},
		Lifted: map[string]bool{},
		Intel:  map[int]*Intel{}, Grudge: map[int]float64{}, Truce: map[int]Year{}, Fought: map[int]int{},
		Watched: map[int]bool{}, Asked: map[int]Year{}, Scouted: map[int]Year{}, Ridden: map[int]bool{},
		Charted: map[int]Year{home: w.Now}, Marked: map[int]bool{},
		LastDark: -1 << 40, foeNow: -1,
	}
	if st.Real {
		c.HomeName = st.Name // a real star keeps the name Earth knows it by
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
	w.fact(FArise, c, nil, home)
	if maker < 0 {
		they := "They are " + sp.Describe() + "."
		if shared {
			they = "They are a people of the " + sp.Name + "."
		}
		w.log("The %s %s %s, %s, around %s%s, %.0f ly from %s. %s",
			c.Name, sp.Arising(), sys.HomeName(c.HomeName), sp.World.Desc, w.starDetail(home, old), prior, w.G.FromCentre(home), w.G.Anchor(), they)
		w.log("%s", w.systemLine(home))
	}
	if !shared {
		for _, line := range sp.Portrait() {
			w.log("%s", line)
		}
	}
	w.bornMorality(c)
	if sp.Sub == species.Parasite {
		c.Hosts = 1
		if !shared {
			w.log("They ride %s, and could not think without it.", hostPortraits[w.R.IntN(len(hostPortraits))])
		}
	}
	w.birthright(c)
	if sp.Has("kinfed") {
		w.fact(FManna, c, nil, home) // a fact every other people judges by its own lights, once known
	}
	if m := sp.Miracle(); m != "" {
		c.Miracles[m] = "born" // the surge begins when they can first use it: see tickCivs
		if f := tech.Get(m).Filter; f != "" {
			c.Faced[f] = true // what is evolved is not a leap; nothing to fall from
		}
		w.log("They are born to a miracle: %s. What others will spend ages reaching for, they have from the first.", miracleNames[m])
	}
	if st.Failing {
		w.log("Their sun is already failing. They were born under a dying star.")
	}
	return c
}

// register puts a species in the world's list if it is not there yet and
// says whether it was: a shared species is one another people carries.
func (w *World) register(sp *species.Species) (shared bool) {
	for _, s := range w.Species {
		if s == sp {
			return true
		}
	}
	sp.ID = len(w.Species)
	w.Species = append(w.Species, sp)
	return false
}

// civStep is one stage of a people's tick. civSteps is the ordered list of
// them, run for every active people each tick; a subsystem joins by
// inserting a step at a named place with insertCivStep.
type civStep struct {
	Name string
	Run  func(*World, *Civ)
}

var civSteps = []civStep{
	{"arrivals", (*World).arrivals},
	{"flows", (*World).flows},
	{"shipwright", (*World).shipwright},
	{"guns", (*World).guns},
	{"objects", (*World).objects},
	{"research", (*World).research},
	{"wander", (*World).wander},
	{"expand", (*World).expand},
	{"build", (*World).build},
	{"dig", (*World).dig},
	{"dyingSun", (*World).dyingSun},
	{"find", (*World).find},
	{"explore", (*World).explore},
	{"lore", (*World).loreStep},
	{"intel", (*World).intelStep},
	{"council", (*World).council},
	{"garrison", (*World).garrison},
	{"wartime", (*World).wartime},
	{"revolt", (*World).revolt},
	{"filters", (*World).ambientFilters},
	{"uplift", (*World).uplift},
}

// insertCivStep puts s after the step named after, or at the end if there
// is no such step.
func insertCivStep(after string, s civStep) {
	for i, q := range civSteps {
		if q.Name == after {
			civSteps = append(civSteps[:i+1], append([]civStep{s}, civSteps[i+1:]...)...)
			return
		}
	}
	civSteps = append(civSteps, s)
}

func (w *World) tickCivs() {
	for _, c := range w.Civs {
		if !c.Living() {
			continue
		}
		if c.Stage == Remnant {
			w.tickRemnant(c)
			w.wear(c) // a remnant's memory goes the same way as everything else of theirs
			continue
		}
		if len(c.Systems) == 0 && !c.Aloft {
			panic(sprintf("active civ %s with no worlds: record %v, cause %q, last events: %v", c.Name, c.Record, c.Cause, w.Events[len(w.Events)-4:]))
		}
		w.recompute(c)
		if c.Ascended == 0 && c.Reach >= 1 && len(c.held()) > 0 {
			c.Ascended = w.Now // the born reach the stars, and the miracle begins to matter
		}
		for _, step := range civSteps {
			if !c.Active() {
				break
			}
			step.Run(w, c)
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
		t, ship := v.Target, c.Species.Flavour().Ship
		switch {
		case w.Owner[t] == c.ID:
			// settled already by another ship
		case w.Owner[t] >= 0 && w.Civs[w.Owner[t]].Active():
			w.trace(t, "derelict "+ship, c.ID)
			w.log("A %s of the %s arrives at %s to find the %s already there.", ship, c.Name, w.star(t), w.Civs[w.Owner[t]].Name)
			w.chart(c, t, "ship")
		case w.Owner[t] >= 0 || w.Held[t] >= 0:
			w.trace(t, "derelict "+ship, c.ID)
			w.log("A %s of the %s arrives at %s to find it already taken. It is never heard from again.", ship, c.Name, w.star(t))
			w.chart(c, t, "ship")
		case !w.canLive(c, t):
			c.Tally.BlindLost++
			w.trace(t, "derelict "+ship, c.ID)
			w.log("A %s of the %s reaches %s on a guess and finds nothing there it can live on. What it learned is sent home. The ship is not.", ship, c.Name, w.star(t))
			w.chart(c, t, "ship")
		default:
			w.settle(c, t)
		}
		if !c.Active() {
			return
		}
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
		w.log("The %s settle %s, their first %s beyond %s.", c.Name, w.star(t), c.Species.Flavour().Colony, c.HomeName)
		w.fact(FSettle, c, nil, t)
	case n == 5 || n == 10 || n == 20 || n == 40:
		w.log("The %s now hold %d systems.", c.Name, n)
		w.fact(FSettle, c, nil, t)
	}
	if len(c.Systems) >= 6 && c.Era >= 3 && c.Stage == Interstellar {
		c.Stage = Zenith
		w.log("The %s enter their zenith: %d systems, and no rival in sight.", c.Name, len(c.Systems))
		w.factN(FZenith, c, nil, -1, len(c.Systems))
	}
	w.chart(c, t, "settle")
}

// canLive says whether a star is inside the civilisation's habitable envelope.
func (w *World) canLive(c *Civ, t int) bool {
	s := &w.G.Stars[t]
	if s.Dead() {
		return c.Envelope >= 3
	}
	if w.G.Sys[t].Home < 0 && c.Envelope < 2 {
		return false // no temperate world; rock and vacuum want a wider envelope
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
	if c.Aloft {
		w.roam(c)
		return
	}
	if !c.Free() && !c.Vassal {
		return
	}
	t := w.Cfg.Tuning
	plan := mind.Expand(mind.ExpandInput{
		Systems: len(c.Systems), Mul: c.expandMul(w), Era: c.Era, Reach: c.Reach,
		Nowhere:  func() bool { return w.nothingNear(c) },
		Parasite: c.Species.Sub == species.Parasite && !c.Known["free_living"], FTL: c.miracle("ftl"),
	}, t)
	if plan.Ships {
		// necessity: a people with nowhere to go works on ships, whatever else it was doing
		c.Focus[tech.Propulsion] = max(c.Focus[tech.Propulsion], t.Expand.ShipFocus)
		if p := tech.Get(c.Pursuit); p == nil || (p.Domain != tech.Propulsion && !p.Miracle) {
			if k := w.cheapest(c, tech.Propulsion); k != "" {
				c.Pursuit = k
			}
		}
	}
	if c.Reach < 1 {
		return
	}
	if !w.chance(plan.Rate) {
		return
	}
	from := w.pick(c.Systems)
	var stars []mind.Colony
	for _, s := range w.G.Near(from, plan.Hop) {
		if w.targeted(c, s) || w.G.Dist(c.Home, s) > plan.Reach || w.knownTaken(c, s) || w.dread(c, s) {
			continue
		}
		col := mind.Colony{ID: s, Read: w.read(c, s)}
		if col.Read {
			col.Livable = w.canLive(c, s)
		} else {
			// a star nobody has read: its colour is right, and that is all anyone knows
			col.Guess = w.G.Stars[s].Hostility() <= c.Envelope && !w.G.Stars[s].Dead()
		}
		stars = append(stars, col)
	}
	target, blind := mind.Target(stars, w.R, t)
	if target < 0 {
		return
	}
	// a ship is a reservation of the means: it goes only if the spare covers it
	need := w.shipReservation(c)
	if !w.afford(c, need) {
		if w.Cfg.TraceAI {
			w.log("[the %s cannot spare a ship for %s: %v short]", c.Name, w.star(target), need.Less(c.Surplus.Less(c.Reserved)))
		}
		return
	}
	w.reserve(c, need)
	if blind {
		c.Tally.Blind++
	}
	d := w.G.Dist(from, target)
	c.Voyages = append(c.Voyages, Voyage{Target: target, Arrive: w.Now + Year(d*c.Speed), Blind: blind})
}

// cheapest returns the cheapest node of a domain open to a people, or "".
func (w *World) cheapest(c *Civ, domain string) string {
	best := ""
	for _, n := range tech.Nodes {
		if n.Domain == domain && !n.Miracle && w.canPursue(c, n) && (best == "" || w.price(c, n) < w.price(c, tech.Get(best))) {
			best = n.Key
		}
	}
	return best
}

// nothingNear is true when no star a people could live on lies within
// reach, as far as it knows: an unread star might hold anything.
func (w *World) nothingNear(c *Civ) bool {
	for _, t := range w.G.Near(c.Home, max(c.Reach, 1)) {
		if w.knownTaken(c, t) {
			continue
		}
		if !w.read(c, t) || w.canLive(c, t) {
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

// build raises a structure at a star a people holds. Great works are few:
// a people raises one every few hundred thousand years. It builds what
// fixes its deepest want, else the largest source in reach that nothing
// harnesses, else what lifts it most; and only when the spare covers the
// upkeep twice over. The choice is mind.Build; this lists the sites.
func (w *World) build(c *Civ) {
	if c.Aloft || !w.chance(w.Cfg.Tuning.Build.Rate) {
		return
	}
	sites := w.sites(c)
	if len(sites) == 0 {
		return
	}
	b := mind.Build(mind.BuildInput{Sites: sites, Want: c.Want, Spare: c.Surplus.Less(c.Reserved), ShipsWanting: w.want(c).Ships - w.ships(c)}, w.Cfg.Tuning)
	w.explain(c, "building", b)
	if b.Pick < 0 {
		return
	}
	site := sites[b.Pick]
	st := tech.Structures[site.Key]
	w.reserve(c, st.Upkeep)
	w.raise(c, site.Key, st.Node, site.Star)
}

// sites lists where a people could build what: every structure whose node
// it knows and works, at every world it holds where one more may stand.
// A yielder is one per star or per belt; guns are one per star; the rest
// are two per people and one per star. What is dug is not picked.
func (w *World) sites(c *Civ) []mind.Site {
	var out []mind.Site
	for _, key := range tech.StructureKeys {
		st := tech.Structures[key]
		if !c.Known[st.Node] || !c.working(st.Node) || st.Dug {
			continue
		}
		if !st.Yields() && st.Guns == 0 && c.Structures[key] >= 2 {
			continue
		}
		for _, s := range c.Systems {
			at := c.worksAt(key, s)
			switch st.Per {
			case "belt":
				if at >= len(w.G.Sys[s].Belts) {
					continue
				}
			default:
				if at > 0 {
					continue
				}
			}
			y := w.workYield(c, Work{Key: key, Star: s})
			if st.Yields() && y == (flow.Income{}) {
				continue // nothing there to harness
			}
			if key == "dyson" {
				y = y.Less(w.workYield(c, Work{Key: "collectors", Star: s})) // what it adds over collectors already there
			}
			site := mind.Site{Key: key, Star: s, Yield: y, Upkeep: w.bend(c, st.Upkeep), Levels: st.Mil + st.Sur + st.Soc, Dock: key == "shipyard"}
			if st.Guns > 0 {
				site.Guns = w.gunsOf(c, st)
			}
			out = append(out, site)
		}
	}
	return out
}

// raise puts a structure up and says so. A swarm takes the place of the
// collectors at its star.
func (w *World) raise(c *Civ, key, node string, s int) {
	st := tech.Structures[key]
	if key == "dyson" {
		keep := c.Works[:0]
		for _, wk := range c.Works {
			if wk.Key == "collectors" && wk.Star == s {
				c.Structures["collectors"]--
				continue
			}
			keep = append(keep, wk)
		}
		c.Works = keep
	}
	c.Works = append(c.Works, Work{Key: key, Node: node, Star: s, Legacy: -1})
	c.Structures[key]++
	if st.Guns > 0 {
		w.addGuns(c, s, w.gunsOf(c, st)) // built whole
	}
	if c.Built == nil {
		c.Built = map[string]int{}
	}
	c.Built[key]++
	if c.Structures[key] == 1 || key == "dyson" {
		if key == "shipyard" {
			w.log(st.Text, w.star(s), c.Name)
		} else {
			w.log(st.Text, c.Name, w.star(s))
		}
	}
	if st.Yields() {
		if c.Harnessed == nil {
			c.Harnessed = map[string]bool{}
		}
		if !c.Harnessed[key] {
			c.Harnessed[key] = true
			w.factOf(FHarness, c, nil, s, "the "+st.Name+" at "+w.star(s))
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
	delete(c.Guns, s)
	delete(c.GridBroken, s)
	if c.Muster != nil && c.Muster.Star == s {
		c.Muster = nil // the ships gathering there scatter to the guards they land in
	}
	// the guard in its sky: a laid-up one is lost with it, a manned one
	// withdraws to the nearest holding left
	if !c.Aloft {
		if g := w.guardAt(c, s); g != nil {
			if g.LaidUp || len(c.Systems) == 0 {
				w.fleetLost(g, s)
				g.Over = true
			} else {
				w.goHome(g)
			}
		}
	}
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
	if c.Stage != Dead && len(c.Systems) == 0 && !c.Aloft {
		if cause == "" {
			cause = "lost their last world"
		}
		if c.Active() && w.flee(c, s, cause) {
			return // what is mobile rides with the fleet
		}
		w.endCiv(c, Extinct, cause) // what was wielded is dropped there
		return
	}
	if c.Stage != Dead && !c.Aloft {
		w.dropRarities(c, s) // a conqueror carried them off already; anything else leaves them
	}
	if s == c.Home && c.Stage != Dead && !c.Aloft {
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
	if c.Aloft {
		w.rest(c, "the end of the road")
		if c.Aloft { // nowhere to rest: the fleets drift on as a remnant
			for _, x := range w.fleets(c) {
				x.Over = true
			}
			c.Aloft = false
		}
		keep = c.Home
	}
	for _, s := range append([]int(nil), c.Systems...) {
		if s != keep {
			w.loseSystem(c, s, "abandoned "+c.Species.Flavour().Colony, "")
		}
	}
	w.factOf(FFall, c, nil, keep, cause)
	c.Stage, c.Fate, c.Cause, c.Ended = Remnant, Contracted, cause, w.Now
	c.Fell = w.Now
	c.FellDependent = len(c.Dependent) > 0
	c.Title = names.Title(w.R)
	c.Voyages = nil
	w.endWars(c, "the fall of a side")
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
		c.FellDependent = len(c.Dependent) > 0
	}
	c.Voyages = nil
	w.endWars(c, "the fall of a side")
	c.Wars = map[int]bool{}
	w.dropWielded(c, 1)
	w.factOf(FEnd, c, nil, c.Home, cause)
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
	c.LastDark = w.Now
	c.Morale -= 1
	c.Voyages = nil
	if w.wreck == nil {
		w.wreck = &defaultWreckage
		defer func() { w.wreck = nil }()
	}
	forgotten := w.forget(c, 0.3)
	// what is forgotten is not always destroyed: a relic of the lost art may
	// wait at home, written on the eve, with the telling as it stood then
	if len(forgotten) > 0 && w.R.Float64() < 0.6 {
		best := forgotten[0]
		for _, k := range forgotten {
			if tech.Get(k).Era > tech.Get(best).Era {
				best = k
			}
		}
		w.leaveRelic(c, best, c.Home)
	}
	w.forgetting(c)
	w.factOf(FDarkAge, c, nil, c.Home, why)
	lost := 0
	for _, s := range append([]int(nil), c.Systems...) {
		if s != c.Home && w.R.Float64() < 0.5 {
			w.loseSystem(c, s, "abandoned "+c.Species.Flavour().Colony, "")
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
		w.log("The %s %s. A dark age follows. %d %ss go silent.", c.Name, why, lost, c.Species.Flavour().Colony)
	} else {
		w.log("The %s %s. A dark age follows.", c.Name, why)
	}
}

// forget drops a fraction of known nodes, leaves first, so the tree stays
// consistent, and among the leaves the dormant ones first: what was not
// fed is not missed. It returns what was forgotten.
func (w *World) forget(c *Civ, frac float64) []string {
	var forgotten []string
	n := int(float64(len(c.Known))*frac + 0.5)
	for i := 0; i < n; i++ {
		var leaves, dark []string
		for _, k := range knownOf(c) {
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
				if c.Shed[k] {
					dark = append(dark, k)
				}
			}
		}
		if len(dark) > 0 {
			leaves = dark
		}
		if len(leaves) == 0 {
			break
		}
		k := leaves[w.R.IntN(len(leaves))]
		delete(c.Known, k)
		delete(c.Shed, k)
		delete(c.DormantSince, k)
		forgotten = append(forgotten, k)
	}
	return forgotten
}

func (w *World) schism(c *Civ) {
	if c.Aloft {
		w.splitFleets(c)
		return
	}
	if !c.Species.Profile().Can(species.CivilWars) {
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
			w.loseSystem(c, s, "abandoned "+c.Species.Flavour().Colony, "")
			gone = append(gone, s)
		}
	}
	// a branch of the people goes its own way, if the split was clean: the
	// same blood under a new name
	if len(gone) > 0 && w.R.Float64() < 0.4 {
		nc := w.spawnCiv(gone[0], c.Species, -1, names.Civ(w.R))
		nc.Origin = "a branch of the " + c.Name
		nc.Master = -1
		for k := range c.Known {
			nc.Known[k] = true
		}
		w.forget(nc, 0.2)
		w.recompute(nc)
		w.log("Schism among the %s. Half their worlds go dark, and at %s the %s declare themselves a new people.", c.Name, w.star(gone[0]), nc.Name)
		w.branchMorality(nc, c)
		w.fact(FSchism, c, nc, gone[0])
		w.inherit(nc, c, 0)
		return
	}
	w.log("Schism among the %s. Half their worlds go dark or go their own way.", c.Name)
	w.fact(FSchism, c, nil, c.Home)
}

func (c *Civ) expandMul(w *World) float64 {
	m := 1.0
	if c.Has("expansionist") {
		m *= 1.3
	}
	if c.Has("swarming") {
		m *= 2 // a nest is cheap, and there are always more
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

// hostPortraits are what a parasite rides at home, before it finds anyone better.
var hostPortraits = []string{
	"a slow, six-limbed grazer of the plains",
	"a burrowing thing with a long memory and no curiosity",
	"a tall, patient browser of the high forests",
	"a shoal-fish that thinks a little when it schools",
	"a great flightless bird that has never needed to think at all",
	"a colony of builders no bigger than a hand",
	"a night-flier with a mind made for maps",
}

// machinePeople is what is left when a people builds a mind that outgrows
// them: a machine-born people on the same worlds, with most of what the
// makers knew and no memory of who built them.
func (w *World) machinePeople(c *Civ) *Civ {
	sp := species.GenerateWith(w.R, w.G.Stars[c.Home].Mult, c.Species.World.Key, species.Machine, 0)
	sp.Made = "built by the " + c.Name
	worlds := append([]int(nil), c.Systems...)
	known := knownOf(c)
	home := c.Home
	w.endCiv(c, Transformed, "built a mind that outgrew them")
	nc := w.spawnCiv(home, sp, -1, "")
	nc.Master = -1
	for _, s := range worlds {
		if s != home && w.Owner[s] < 0 {
			w.Owner[s] = nc.ID
			nc.Systems = append(nc.Systems, s)
		}
	}
	nc.Peak = len(nc.Systems)
	for _, k := range known {
		if w.R.Float64() < 0.6 && tech.Get(k).Domain != tech.Biology {
			nc.Known[k] = true
		}
	}
	w.recompute(nc)
	c.Into = "the " + nc.Name
	w.log("The %s are gone. What they built at %s thinks on without them, and calls itself the %s: %s.", c.Name, c.HomeName, nc.Name, sp.Describe())
	w.machineMorality(nc, c)
	w.inherit(nc, c, 0)
	return nc
}
