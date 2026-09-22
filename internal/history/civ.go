package history

import (
	"slices"

	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// spawnCiv raises a civilisation at a star. sp is nil for a natural species;
// made species pass their own and the maker's id. The people has no
// name here: what it is called is a row of the names pass, keyed by its id.
func (w *World) spawnCiv(home int, sp *species.Species, maker int) *Civ {
	return w.spawn(home, sp, maker, nil)
}

// spawn is spawnCiv with a host: a rider woken in a people holds no world
// of its own and lives at the host's home.
func (w *World) spawn(home int, sp *species.Species, maker int, host *Civ) *Civ {
	st := &w.G.Stars[home]
	sys := w.G.Sys[home]
	sys.EnsureHome(w.R, st)
	natural := sp == nil
	if sp == nil {
		sp = species.GenerateOn(w.R, st.Mult, sys.Arch)
		if w.Law.Glare > 3 && !sp.Has("hardy") && !sp.Has("skyless") {
			sp.Add("hardy") // born under a hard sky
		}
	}
	shared := w.register(sp)
	c := w.newCiv(home, sp, maker)
	if host != nil {
		c.Systems, c.Peak = nil, 0
	} else if sp.Sub == species.Parasite {
		for _, o := range w.Civs {
			if o != c && o.Species == sp && o.Own >= 0 {
				c.Own = o.Own // a branch of a rider is the same plague
				break
			}
		}
	}
	if host == nil {
		w.setOwner(home, c.ID)
	}
	prior := -1
	for _, o := range w.Civs {
		if o != c && o.Home == home {
			prior = o.ID
		}
	}
	w.recompute(c)
	arose := w.told(FArise, c, nil, home).with(P{"made": maker >= 0, "shared": shared, "prior": prior, "species": sp.ID, "class": string(st.Class), "traits": sp.TraitKeys()})
	if host != nil {
		arose.P["host"] = host.ID
	} else if maker < 0 {
		w.event(KSystem, c, nil, home, P{})
	}
	if !shared {
		w.event(KPortrait, c, nil, -1, P{"species": sp.ID, "powers": append([]string(nil), sp.Powers...)})
	}
	w.bornMorality(c)
	w.birthright(c)
	w.innate(c)
	w.bornPowers(c)
	if sp.Has("kinfed") {
		w.fact(FManna, c, nil, home).with(P{"way": "kinfed"}) // a fact every other people judges by its own lights, once known
	}
	if m := sp.Miracle(); m != "" {
		w.holdMiracle(c, m, "born") // the surge begins when they can first use it: see tickCivs
		if f := tech.Get(m).Filter; f != "" {
			c.Faced[f] = true // what is evolved is not a leap; nothing to fall from
		}
		w.event(KBornMiracle, c, nil, -1, P{"miracle": m})
	}
	if st.Failing {
		w.event(KBornFailing, c, nil, home, P{})
	}
	if natural {
		w.bornRider(c)
	}
	if maker >= 0 {
		w.renew(w.Civs[maker], 0.2) // a people made is something new
	}
	return c
}

// newCiv is the bare people: the struct with its maps, on the world's
// list, holding its home and nothing else. spawn dresses a birth; the
// heirs of a sundering (sunder.go) are dressed from the old people.
func (w *World) newCiv(home int, sp *species.Species, maker int) *Civ {
	w.register(sp)
	c := &Civ{
		ID: len(w.Civs), Species: sp, Home: home,
		Cradle: home, Born: w.Now, Renewed: w.Now, Still: w.Now, Systems: []int{home}, Peak: 1, Master: maker,
		Known: map[string]bool{}, Learned: map[string]Year{}, Focus: map[string]float64{}, Locked: map[string]bool{},
		Structures: map[string]int{}, Found: map[int]bool{}, Heard: map[int]bool{},
		Wars: map[int]bool{}, Met: map[int]bool{}, Reached: map[int]bool{}, Trade: map[int]bool{},
		Faced: map[string]bool{}, Scars: map[string]bool{}, Boons: map[string]bool{}, Miracles: map[string]string{},
		Lifted: map[string]bool{},
		Intel:  map[int]*Intel{}, Grudge: map[int]float64{}, Truce: map[int]Year{}, Fought: map[int]int{},
		Watched: map[int]bool{}, Asked: map[int]Year{}, Scouted: map[int]Year{}, Ridden: map[int]bool{},
		Charted: map[int]Year{home: w.Now}, Marked: map[int]bool{},
		Sire: maker, Fathomed: map[int]bool{}, FathomTried: map[int]Year{},
		Infections: map[int]*Infection{}, Immune: map[int]bool{}, Suspect: map[int]bool{}, Closed: map[int]bool{},
		Own: -1, Weapons: map[string]*Weapon{}, Barred: map[int]bool{}, FallEvent: -1, EndEvent: -1,
		LastDark: -1 << 40, foeNow: -1,
	}
	w.Civs = append(w.Civs, c)
	w.born(c)
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
	{"eat", (*World).eat},
	{"shipwright", (*World).shipwright},
	{"guns", (*World).guns},
	{"objects", (*World).objects},
	{"research", (*World).research},
	{"drift", (*World).drift},
	{"eldritch", (*World).eldritchStep},
	{"wander", (*World).wander},
	{"expand", (*World).expand},
	{"build", (*World).build},
	{"dig", (*World).dig},
	{"dyingSun", (*World).dyingSun},
	{"find", (*World).find},
	{"explore", (*World).explore},
	{"lore", (*World).loreStep},
	{"fathoming", (*World).fathoming},
	{"intel", (*World).intelStep},
	{"council", (*World).council},
	{"hunts", (*World).huntStep},
	{"contracting", (*World).contracting},
	{"garrison", (*World).garrison},
	{"wartime", (*World).wartime},
	{"revolt", (*World).revolt},
	{"filters", (*World).ambientFilters},
	{"uplift", (*World).uplift},
}

// offSteps are the steps an ossified people skips on its off ticks: it
// holds no council, launches nothing, settles nothing, builds nothing,
// offers nothing and banks no research. See ossify.go.
var offSteps = map[string]bool{
	"shipwright": true, "research": true, "eldritch": true, "expand": true, "build": true, "explore": true,
	"council": true, "contracting": true, "uplift": true,
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
		if c.Own >= 0 && !w.rides(c) {
			w.starveOne(c) // the last host died since the plagues' pass
			if !c.Living() {
				continue
			}
		}
		if len(c.Systems) == 0 && !c.Aloft && !w.rides(c) {
			panic(sprintf("active civ %s with no worlds: record %v, cause %q, last events: %v", c.Tok(), c.Record, c.Cause, w.Events[len(w.Events)-4:]))
		}
		w.recompute(c)
		if c.Ascended == 0 && c.Reach >= 1 && len(c.held()) > 0 {
			c.Ascended = w.Now // the born reach the stars, and the miracle begins to matter
		}
		if c.Asleep {
			w.guns(c) // the long sleep: nothing but the body, until disturbed
			continue
		}
		off := w.offTick(c)
		for _, step := range civSteps {
			if !c.Active() {
				break
			}
			if off && offSteps[step.Name] {
				continue // an ossified people sits this tick out: upkeep, fleets, wars and answers only
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
			w.forgive(c)
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
		t, ship := v.Target, c.Species.ID
		switch {
		case w.Owner[t] == c.ID:
			// settled already by another ship
		case w.Owner[t] >= 0 && w.Civs[w.Owner[t]].Active() && !w.perceives(c, w.Civs[w.Owner[t]]):
			// a world of something the people cannot hold in mind: the ship is a loss with no doer on its ledger
			o := w.Civs[w.Owner[t]]
			w.trace(t, "derelict_ship", c)
			w.fact(FShipLost, c, o, t).with(P{"species": ship})
			c.Morale -= 0.2
			w.chart(c, t, "ship")
		case w.Owner[t] >= 0 && w.Civs[w.Owner[t]].Active():
			w.trace(t, "derelict_ship", c)
			w.event(KArrivalLost, c, w.Civs[w.Owner[t]], t, P{"species": ship, "why": "held"})
			w.chart(c, t, "ship")
		case w.Owner[t] >= 0:
			w.trace(t, "derelict_ship", c)
			w.event(KArrivalLost, c, nil, t, P{"species": ship, "why": "taken"})
			w.chart(c, t, "ship")
		case !w.canLive(c, t):
			c.Tally.BlindLost++
			w.trace(t, "derelict_ship", c)
			w.event(KArrivalLost, c, nil, t, P{"species": ship, "why": "barren"})
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
	w.holdWorld(c, t)
	switch n := len(c.Systems); {
	case c.colonies == 1:
		w.factN(FSettle, c, nil, t, n).with(P{"first": true, "species": c.Species.ID, "home": c.Home})
	case n == 5 || n == 10 || n == 20 || n == 40:
		w.factN(FSettle, c, nil, t, n).with(P{"first": false, "species": c.Species.ID, "home": c.Home})
	}
	w.afterHold(c, t)
}

// holdWorld is a new world held: the state of a settlement, whether a
// ship brought it or another of an eldritch people is simply there.
func (w *World) holdWorld(c *Civ, t int) {
	w.setOwner(t, c.ID)
	c.Systems = append(c.Systems, t)
	c.colonies++
	w.stir(c)
	w.takeOver(c, t)
	if len(c.Systems) > c.Peak {
		c.Peak = len(c.Systems)
	}
}

// afterHold is what follows a new world, after the lines: the zenith,
// the chart, what wakes there and who notices.
func (w *World) afterHold(c *Civ, t int) {
	if len(c.Systems) >= 6 && c.Era >= 3 && c.Stage == Interstellar {
		w.setStage(c, Zenith)
		w.factN(FZenith, c, nil, -1, len(c.Systems))
	}
	w.chart(c, t, "settle")
	w.wakeReservoir(c, t)
	w.settledNear(c, t)
	w.disturbed(c, t)
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
		if l.Kind == Law && l.variant() == "mind_dead" && l.State != Mastered && w.G.Dist(l.Star, t) <= 8 {
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
	p := c.Species.Profile()
	if !p.Can(species.SettlesByShip) || (p.Worlds > 0 && len(c.Systems)+len(c.Voyages) >= p.Worlds) {
		return // nothing crosses by ship, or it holds what it can
	}
	t := w.Cfg.Tuning
	plan := mind.Expand(mind.ExpandInput{
		Systems: len(c.Systems), Mul: c.expandMul(w), Era: c.Era, Reach: c.Reach,
		Nowhere:  func() bool { return w.nothingNear(c) },
		Parasite: c.Own >= 0 && !c.Known["free_living"], FTL: c.miracle("ftl"),
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
	if c.Reach < 1 || len(c.Systems) == 0 {
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
	// a ship is a reservation of the means: it goes only if the spare
	// covers it; a people with no upkeep is sustained by whatever it is
	if p.Can(species.Pays) {
		need := w.shipReservation(c)
		if !w.afford(c, need) {
			if w.Cfg.TraceAI {
				w.event(KDebug, c, nil, target, P{"text": sprintf("[the %s cannot spare a ship for %s: %v short]", c.Tok(), w.star(target), need.Less(c.Surplus.Less(c.Reserved)))})
			}
			return
		}
		w.reserve(c, need)
	}
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
	rate := w.Cfg.Tuning.Build.Rate
	if c.Ossified {
		rate *= 0.5 // everything takes twice as long
	}
	if !c.Species.Profile().Can(species.Works) {
		return // it makes nothing
	}
	if c.Aloft || !w.chance(rate) {
		return
	}
	sites := w.sites(c)
	if len(sites) == 0 {
		return
	}
	b := mind.Build(mind.BuildInput{Sites: sites, Want: c.Want, Spare: c.Surplus.Less(c.Reserved), ShipsWanting: w.want(c).Ships - w.ships(c), Fear: c.Dials.Fear}, w.Cfg.Tuning)
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
// are two per people, or what the structure says, and one per star. What
// is dug is not picked.
func (w *World) sites(c *Civ) []mind.Site {
	var out []mind.Site
	for _, key := range tech.StructureKeys {
		st := tech.Structures[key]
		if !c.Known[st.Node] || !c.working(st.Node) || st.Dug {
			continue
		}
		if limit := max(st.Max, 2); !st.Yields() && st.Guns == 0 && c.Structures[key] >= limit {
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
			site := mind.Site{Key: key, Star: s, Yield: y, Upkeep: w.bend(c, st.Upkeep), Levels: st.Mil + st.Sur + st.Soc, Dock: key == "shipyard", Watch: st.Watch}
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
		w.event(KBuilt, c, nil, s, P{"work": key})
	}
	if st.Yields() {
		if c.Harnessed == nil {
			c.Harnessed = map[string]bool{}
		}
		if !c.Harnessed[key] {
			c.Harnessed[key] = true
			w.fact(FHarness, c, nil, s).with(P{"work": key})
		}
	}
}

func (w *World) tickRemnant(c *Civ) {
	if w.chance(0.00003) {
		w.endCiv(c, Extinct, because("faded"))
	}
}

// loseSystem removes a star from a civilisation and leaves a trace. A living
// civilisation with no worlds left is extinct; cause says why.
func (w *World) loseSystem(c *Civ, s int, kind string, cause reason) {
	if !contains(c.Systems, s) {
		return
	}
	c.Systems = remove(c.Systems, s)
	w.setOwner(s, -1)
	w.trace(s, kind, c)
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
	if c.Stage != Dead && len(c.Systems) == 0 && !c.Aloft && !w.rides(c) {
		if cause.none() {
			cause = because("last_world")
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
		w.seatLost(c, cause)
	}
	w.renew(c, 0.05) // a loss is something new
}

// seatLost is what a lost home does to a people that still holds worlds:
// the seat moves to the nearest, unless the people cannot move (a world
// that is the mind, dead with it), or is a hive of one queen (dead with
// her), or a hive of no queen (every world its own people).
func (w *World) seatLost(c *Civ, cause reason) {
	if cause.none() {
		cause = because("lost_home").At(c.Home)
	}
	switch {
	case !c.Species.Profile().Can(species.Reseats):
		w.event(KHomeLost, c, nil, c.Home, P{"way": "was"})
		w.endCiv(c, Extinct, cause)
	case c.Has("onequeen"):
		w.event(KHomeLost, c, nil, c.Home, P{"way": "queen"})
		w.endCiv(c, Extinct, cause).P["queen"] = true
	case c.Has("noqueen") && len(c.Systems) > 1:
		w.event(KHomeLost, c, nil, c.Home, P{"way": "noqueen"})
		w.setHome(c, c.Systems[0])
		w.shatter(c, cause, nil)
	default:
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
	w.setHome(c, best)
	c.Dying = false
	w.event(KReseated, c, nil, c.Home, P{})
}

// contract shrinks a civilisation to its home (or one world) as a remnant.
func (w *World) contract(c *Civ, cause reason) {
	keep := c.Home
	if !contains(c.Systems, keep) && len(c.Systems) > 0 {
		keep = w.pick(c.Systems)
	}
	if c.Aloft {
		w.rest(c, because("road_end"))
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
			w.loseSystem(c, s, "abandoned", reason{})
		}
	}
	fell := w.unplaced(FFall, c, nil, keep).with(P{"peak": c.Peak}).with(cause.params("cause"))
	w.setStage(c, Remnant)
	w.setFate(c, Contracted, cause.key)
	c.Ended = w.Now
	c.FallEvent = fell.ID
	c.Fell = w.Now
	c.FellDependent = len(c.Dependent) > 0
	c.Voyages = nil
	w.endWars(c, "fall")
	c.Wars = map[int]bool{}
	w.dropWielded(c, 0.5)
	w.place(fell)
}

// endCiv finishes a civilisation as extinct or transformed. Systems become traces.
func (w *World) endCiv(c *Civ, f Fate, cause reason) *Event {
	if c.Stage == Dead {
		return nil
	}
	wasRemnant := c.Stage == Remnant
	w.setStage(c, Dead)
	if f == Extinct && len(c.Systems) > 0 && w.R.Float64() < 0.4 {
		w.leaveRelic(c, w.lateNode(c), c.Home)
	}
	for _, s := range append([]int(nil), c.Systems...) {
		if f == Extinct {
			w.loseSystem(c, s, "dead_cities", reason{})
		} else {
			w.loseSystem(c, s, "transformed", reason{})
		}
	}
	w.setFate(c, f, cause.key)
	c.Ended = w.Now
	if !wasRemnant {
		c.Fell = w.Now
		c.FellDependent = len(c.Dependent) > 0
	}
	c.Voyages = nil
	w.endWars(c, "fall")
	c.Wars = map[int]bool{}
	w.dropWielded(c, 1)
	end := w.told(FEnd, c, nil, c.Home).with(P{"fate": f.String(), "remnant": wasRemnant, "peak": c.Peak}).with(cause.params("cause"))
	c.EndEvent = end.ID
	return end
}

// darkAge is the one dark age, whoever calls it: a share of the tree
// forgotten, drawn once by the formula (a fresh people a tenth to a
// fifth, Trantor at three four tenths and up, every earlier dark age a
// tenth deeper), the colonies lost at the depth, the institutions gone
// with everything else. Nothing is fatal here; a people that keeps
// falling forgets more each time until the forgetting takes the stars,
// and then it shatters (sunder.go).
func (w *World) darkAge(c *Civ, why reason) {
	if !c.Active() {
		return
	}
	depth := w.darkDepth(c)
	c.DarkAges++
	c.LastDark = w.Now
	c.Morale -= 1
	c.Voyages = nil
	w.reset(c)
	c.Renewed = w.Now
	if w.wreck == nil {
		w.wreck = &tables.defaultWr
		defer func() { w.wreck = nil }()
	}
	forgotten := w.forget(c, depth*c.Species.Profile().Forgets) // a mind that is backed up forgets less
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
	w.forgetFathomed(c)
	if slices.Contains(forgotten, resilience) {
		w.veil(c) // the records that checked themselves are gone, and what they held with them
	}
	f := w.unplaced(FDarkAge, c, nil, c.Home).with(P{"depth": depth, "species": c.Species.ID}).with(why.params("cause"))
	f.N = int(depth*10 + 0.5)
	lost := 0
	for _, s := range append([]int(nil), c.Systems...) {
		if s != c.Home && w.R.Float64() < depth {
			w.loseSystem(c, s, "abandoned", reason{})
			lost++
		}
	}
	w.dropWielded(c, 0.5)
	w.recompute(c)
	if c.Reach < 10 {
		w.setStage(c, Emergent)
	} else if c.Stage == Zenith {
		w.setStage(c, Interstellar)
	}
	f.P["lost"] = lost
	w.place(f)
	if c.Active() && c.Reach < 10 && len(c.Systems) > 1 {
		w.shatter(c, why, forgotten)
	}
}

// darkDepth is the share of the tree a dark age takes, drawn once.
func (w *World) darkDepth(c *Civ) float64 {
	t := &w.Cfg.Tuning.Ossify
	d := t.DepthBase + t.DepthStiff*min(1, c.Stiff/3) + t.DepthPrior*float64(c.DarkAges) + (2*w.R.Float64()-1)*t.DepthNoise
	return clamp(d, t.DepthMin, t.DepthMax)
}

// depthWord says a depth: a tenth, a fifth, a third, half, most.
func depthWord(d float64) string {
	switch {
	case d < 0.15:
		return "a tenth"
	case d < 0.25:
		return "a fifth"
	case d < 0.4:
		return "a third"
	case d < 0.6:
		return "half"
	}
	return "most"
}

// forget drops a fraction of known nodes, leaves first, so the tree stays
// consistent, and among the leaves the dormant ones first: what was not
// fed is not missed. It returns what was forgotten.
func (w *World) forget(c *Civ, frac float64) []string {
	var forgotten []string
	if !c.Species.Profile().Can(species.Researches) {
		return nil // no tree: a dark age does not touch the pool
	}
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
		w.forgetNode(c, k)
		delete(c.Shed, k)
		delete(c.DormantSince, k)
		forgotten = append(forgotten, k)
	}
	return forgotten
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
	m *= c.Species.Profile().Expand // a living world a tenth; an eldritch thing a tenth, unless it hungers
	return m * c.stiffMul()
}

// machinePeople is what is left when a people builds a mind that outgrows
// them: a machine-born people on the same worlds, with most of what the
// makers knew and no memory of who built them.
func (w *World) machinePeople(c *Civ) *Civ {
	sp := species.GenerateWith(w.R, w.G.Stars[c.Home].Mult, c.Species.World.Key, species.Machine, 0)
	sp.Made = species.MadeBy("built", c.ID)
	worlds := append([]int(nil), c.Systems...)
	known := knownOf(c)
	home := c.Home
	w.endCiv(c, Transformed, because("outgrown"))
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
		if w.R.Float64() < 0.6 && tech.Get(k).Domain != tech.Biology {
			w.know(nc, k)
		}
	}
	w.recompute(nc)
	c.Into, c.IntoCivs = "people", []int{nc.ID}
	w.event(KOutgrown, c, nc, c.Home, P{"traits": sp.TraitKeys()})
	w.machineMorality(nc, c)
	w.inherit(nc, c, 0)
	return nc
}
