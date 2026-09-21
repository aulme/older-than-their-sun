package history

import (
	"math"

	"worldgen/internal/species"
)

// A filter is anything that may push a civilisation into decline. In v1 a
// filter tests one or two levels against a difficulty: roll plus level minus
// difficulty gives a margin, and the margin decides overcome, scarred or
// declined. Tech filters fire on discovery; ambient and external ones have
// their own triggers. Traits move the difficulty asymmetrically.

// Outcome of facing a filter.
type Outcome uint8

const (
	Overcome Outcome = iota
	Scarred
	Declined
)

func (o Outcome) String() string { return [...]string{"overcame", "scarred by", "fell to"}[o] }

// Scar and boon keys. Kept as strings so the legends can print them directly.
const (
	ScarAtomicTaboo  = "an atomic taboo"
	ScarChurch       = "a church that outranks the state"
	ScarStewardship  = "a stewardship creed"
	ScarNoMachines   = "a prohibition on thinking machines"
	ScarCentralism   = "iron centralism"
	ScarMortality    = "a mortality creed"
	ScarNoSelfCopies = "the law that no machine may make itself"
	ScarStarFear     = "a fear of their own star"
	ScarLeftBehind   = "being the ones left behind"
	ScarQuarantine   = "a quarantine creed"
	ScarBurningSky   = "the memory of the burning sky"
	ScarSignal       = "a cult of the signal"
	ScarDoor         = "a dread of doors"
	ScarChains       = "the memory of chains"
	ScarFatalism     = "a fatalist creed"
	ScarOtherVoices  = "the other voices on the line"
	ScarChanged      = "having been something else"
	BoonAligned      = "aligned minds"
	BoonSwarm        = "swarm industry"
	BoonUnity        = "unity forged in the atomic age"
	BoonCommunion    = "communion with something older"
)

// Filter describes one hurdle.
type Filter struct {
	Key, Name string
	Levels    []string // "mil", "sur", "soc"; averaged
	Diff      float64
	Repeat    bool
	Domain    string                                                                // research pushed while facing it
	Adjust    func(w *World, c *Civ) (levels []string, diff float64, domain string) // levels, difficulty and domain set by the occasion; nil for the fixed ones
	Overcome  func(w *World, c *Civ)
	Scar      func(w *World, c *Civ)
	Decline   func(w *World, c *Civ)
}

var filters = map[string]*Filter{}

// FilterName is how the legends say a filter, or the key if unknown.
func FilterName(key string) string {
	if f := filters[key]; f != nil {
		return f.Name
	}
	return key
}

// FilterDiff is the base difficulty of a filter and the levels it tests.
func FilterDiff(key string) (float64, []string) {
	if f := filters[key]; f != nil {
		return f.Diff, f.Levels
	}
	return 0, nil
}

func def(f *Filter) { filters[f.Key] = f }

// traitDiff is the asymmetry: how each trait changes each filter's difficulty.
var traitDiff = map[string]map[string]float64{
	"memory":        {"silence": 2, "find": -1},
	"swarming":      {"cosmic": -1, "beacon": 2}, // they scatter; gathered, they hear with one ear
	"unyielding":    {"atomic": 1, "hold": -1},
	"opportunist":   {"hold": 0.5},
	"vengeful":      {"hold": -0.5},
	"confederate":   {"distance": -0.5},
	"pacifist":      {"atomic": -2, "overshoot": -1},
	"conqueror":     {"atomic": 0.5, "machines": 0.5},
	"expansionist":  {"overshoot": 1, "distance": 1},
	"contemplative": {"overshoot": -1, "transcend": 1, "machines": -0.5},
	"curious":       {"machines": 1, "replication": 0.5, "door": 0.5},
	"cautious":      {"machines": -1, "replication": -1, "door": -1, "find": 0.5},
	"collective":    {"atomic": -1, "overshoot": -1},
	"individualist": {"distance": 1, "hold": 1},
	"caste":         {"hold": -0.5},
	"shortlived":    {"silence": -1},
	"longlived":     {"silence": 1},
	"radiation":     {"atomic": -1},
	"solitary":      {"distance": -2, "beacon": -2, "hold": 1},
	"herd":          {"beacon": 2, "atomic": -1, "distance": 1, "hold": -1},
	"dormancy":      {"cosmic": -1, "dying": -1},
	"symbiosis":     {"machines": -1.5, "replication": -0.5},
	"xenophobic":    {"beacon": -1, "find": 1},
	"submissive":    {"revolt": 1, "hold": -0.5},
	"skyless":       {"cosmic": -1},
	"nomadic":       {"overshoot": -1, "distance": -3},
}

func (c *Civ) traitDiff(key string) float64 {
	d := 0.0
	for _, t := range c.Species.Traits {
		d += traitDiff[t.Key][key]
	}
	return d
}

// face resolves a filter against a civilisation and returns the outcome.
// diffAdj lets the caller raise or lower the difficulty for the occasion.
func (w *World) face(c *Civ, key string, diffAdj float64) Outcome {
	f := filters[key]
	if f == nil {
		panic("history: unknown filter " + key)
	}
	if !c.Active() {
		return Declined
	}
	if c.Faced[key] && !f.Repeat {
		return Overcome
	}
	if c.neverFaces(key) {
		return Overcome // nothing in it for the filter to test: no belief, no boredom, no institutions
	}
	again := c.Faced[key]
	c.Faced[key] = true
	master := c.Master
	w.recompute(c)
	if c.miracle("foresight") && key != "sight" && w.R.Float64() < 0.5 {
		w.event(KForesaw, c, nil, -1, P{"filter": key})
		c.Record = append(c.Record, "foresaw "+f.Name)
		return Overcome
	}
	levels, domain := f.Levels, f.Domain
	if f.Adjust != nil {
		if l, d, dom := f.Adjust(w, c); l != nil {
			levels, diffAdj, domain = l, diffAdj+d, dom
		}
	}
	lvl := c.level(levels...)
	diff := f.Diff + diffAdj + 0.25*float64(len(c.Scars)) + 1.5*(w.Hazard-1) + c.traitDiff(key) + c.natureDiff(key) + c.miracleDiff(key) + w.lawDiff(key) + w.thinDiff(key)
	roll := w.R.NormFloat64() * 1.5
	margin := lvl + roll - diff
	if domain != "" {
		c.focus(domain, 1.5)
	}
	var out Outcome
	var how string
	w.wreck = wreckOf(key)
	defer func() { w.wreck = nil }()
	switch {
	case margin >= 0.5:
		out = Overcome
		how = ""
		f.Overcome(w, c)
	case margin >= -2:
		out = Scarred
		c.Morale -= 0.5
		f.Scar(w, c)
		if c.Aloft && c.Active() && w.R.Float64() < 0.2 {
			w.rest(c, f.Name)
		}
	default:
		out = Declined
		c.Morale -= 1
		c.Declines++
		f.Decline(w, c)
	}
	if math.Abs(margin) < 0.5 {
		how = " (narrowly)"
	}
	c.Record = append(c.Record, sprintf("%s %s%s", out, f.Name, how))
	if !again || out == Declined || key == "revolt" {
		w.recordFilter(c, key, f, out, master) // a filter faced again is not a new story unless it wins
	}
	return out
}

// recordFilter is the tale of a filter faced: the outcome, and for a
// revolt who rose against whom, and for the Signal where it came from.
func (w *World) recordFilter(c *Civ, key string, f *Filter, out Outcome, master int) {
	if key == "revolt" && master >= 0 {
		m := w.Civs[master]
		if out == Overcome {
			w.fact(FFreed, c, m, c.Home).with(P{"way": "revolt"})
		} else {
			w.fact(FCrushed, m, c, c.Home)
		}
		return
	}
	kind := FOvercome
	switch out {
	case Scarred:
		kind = FScarred
	case Declined:
		kind = FDeclined
	}
	star := c.Home
	if key == "beacon" && w.transmitter != nil {
		star = w.transmitter.Star // remembered where it came from, so nobody surveys there
	}
	ff := w.fact(kind, c, nil, star).with(P{"filter": key})
	if key == "beacon" && w.transmitter != nil {
		ff.Legacy = w.transmitter.ID
	}
}

// faced is the line of a filter's outcome, told where it happens; the
// fact of it, when it is a new story, is recordFilter's and silent. Way
// tells the lines of one outcome apart.
func (w *World) faced(c *Civ, key, outcome, way string, star int) *Event {
	return w.event(KFaced, c, nil, star, P{"filter": key, "outcome": outcome, "way": way})
}

// ambientFilters are the ones with their own triggers rather than a tech node.
func (w *World) ambientFilters(c *Civ) {
	if !c.Active() {
		return
	}
	if c.NextDrift == 0 {
		c.NextDrift = 6
	}
	if len(c.Systems) >= c.NextDrift && w.chance(0.05) {
		adj := 0.3 * float64(len(c.Systems)-6)
		c.NextDrift *= 2
		if !c.miracle("ansible") && !c.Has("swarming") { // nothing drifts when every world is in the room, or there is no centre
			w.face(c, "distance", adj)
		}
	}
	w.tickStiff(c) // the ways setting, and the filter that comes for whoever survives the rest: see ossify.go
}

func init() {
	def(&Filter{
		Key: "atomic", Name: "the Atomic Age", Levels: []string{"soc"}, Diff: 3.5, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			c.Boons[BoonUnity] = true
			w.faced(c, "atomic", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarAtomicTaboo] = true
			w.faced(c, "atomic", "scarred", "", c.Home)
		},
		Decline: func(w *World, c *Civ) {
			x := w.R.Float64()
			switch {
			case x < 0.5:
				w.darkAge(c, "burned their world to ash")
			case x < 0.8:
				w.endCiv(c, Extinct, "burned themselves out in a single afternoon")
			default:
				w.contract(c, "burned their world and never rose from the ash")
			}
		},
	})
	def(&Filter{
		Key: "overshoot", Name: "Overshoot", Levels: []string{"sur", "soc"}, Diff: 4, Domain: "biology",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "overshoot", "overcome", "", c.Home)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarStewardship] = true
			w.faced(c, "overshoot", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			if w.R.Float64() < 0.6 {
				w.darkAge(c, "exhausted their world")
			} else {
				w.endCiv(c, Extinct, "exhausted their world and starved on it")
			}
		},
	})
	def(&Filter{
		Key: "machines", Name: "Thinking Machines", Levels: []string{"soc"}, Diff: 4.5, Domain: "computation",
		Overcome: func(w *World, c *Civ) {
			c.Boons[BoonAligned] = true
			w.faced(c, "machines", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarNoMachines] = true
			c.Locked["computation"] = true
			w.faced(c, "machines", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			x := w.R.Float64()
			if x < 0.25 {
				w.darkAge(c, "pulled the plug on their own machines, too late and at great cost")
				return
			}
			w.machinePeople(c) // every time it does not pull the plug: what it built thinks on without it
		},
	})
	def(&Filter{
		Key: "distance", Name: "the Distance", Levels: []string{"soc"}, Diff: 4.5, Repeat: true, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "distance", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarCentralism] = true
			w.faced(c, "distance", "scarred", "", c.Home)
		},
		Decline: func(w *World, c *Civ) {
			if !c.Species.Profile().Can(species.CivilWars) && len(c.Systems) > 1 {
				// a people with no factions does not split: the far world is simply cut from the seat, and is its own people after
				far, fd := -1, 0.0
				for _, s := range c.Systems {
					if d := w.G.Dist(c.Home, s); s != c.Home && d > fd {
						far, fd = s, d
					}
				}
				if w.cutOff(c, far) != nil {
					return
				}
			}
			if w.R.Float64() < 0.6 && w.civilWar(c) { // the same story from the colonies' side: strangers, then enemies
				return
			}
			w.contract(c, "watched their colonies become strangers, and then enemies, and then silence")
		},
	})
	def(&Filter{
		Key: "silence", Name: "the Long Silence", Levels: []string{"soc"}, Diff: 4.5, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "silence", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarMortality] = true
			w.faced(c, "silence", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			w.contract(c, "stopped dying, and then stopped being born")
		},
	})
	def(&Filter{
		Key: "replication", Name: "Self-Replication", Levels: []string{"mil"}, Diff: 5.5, Domain: "weapons",
		Overcome: func(w *World, c *Civ) {
			c.Boons[BoonSwarm] = true
			w.faced(c, "replication", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarNoSelfCopies] = true
			w.faced(c, "replication", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			s := w.aWorld(c)
			if w.R.Float64() < 0.3 {
				w.loseSystem(c, s, "stripped world", "were consumed by their own machines")
				w.darkAge(c, "lost "+w.star(s)+" to their own machines and burned the rest to stop it spreading")
				return
			}
			// the eaten world is the new people's, and the old people are gone
			w.faced(c, "replication", "declined", "", s)
			w.loseSystem(c, s, "stripped world", "were consumed by their own machines")
			w.endCiv(c, Extinct, "were consumed by their own machines")
			if nc := w.replicatorAt(s, species.Machine, "the machines of the "+c.Tok()+", copying themselves", false); nc != nil {
				nc.Species.Parent = c.Species
				w.event(KNamedItself, nc, nil, s, P{"way": "eats", "desc": nc.Species.Describe()})
			}
		},
	})
	def(&Filter{
		Key: "stellar", Name: "Stellar Engineering", Levels: []string{"sur"}, Diff: 7, Domain: "exotic",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "stellar", "overcome", "", c.Home)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarStarFear] = true
			c.Locked["exotic"] = true
			w.faced(c, "stellar", "scarred", "", c.Home)
		},
		Decline: func(w *World, c *Civ) {
			w.faced(c, "stellar", "declined", "", c.Home)
			w.loseSystem(c, c.Home, "wounded star", "broke their own star")
			w.Bio[c.Home] = BioNone
			if len(c.Systems) == 0 || w.R.Float64() < 0.5 {
				w.endCiv(c, Extinct, "broke their own star")
			} else {
				w.contract(c, "broke their own star and fled to a lesser one")
			}
		},
	})
	def(&Filter{
		Key: "transcend", Name: "Transcendence", Levels: []string{"soc"}, Diff: 7, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "transcend", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarLeftBehind] = true
			w.faced(c, "transcend", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			if w.R.Float64() < 0.4 {
				w.contract(c, "mostly went elsewhere, leaving a few to mind the ruins")
				return
			}
			for _, s := range c.Systems {
				w.trace(s, "silent machinery", c.ID)
			}
			w.endCiv(c, Transformed, "went elsewhere")
			c.Into = "something that left"
			w.faced(c, "transcend", "declined", "", -1)
		},
	})
	def(&Filter{
		Key: "door", Name: "the Door", Levels: []string{"soc"}, Diff: 4.5, Domain: "exotic",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "door", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarDoor] = true
			w.tear(0.3)
			s := w.sleeperStar(c)
			if s < 0 {
				w.faced(c, "door", "scarred", "noticed", -1)
				return
			}
			w.faced(c, "door", "scarred", "came", s)
			w.sleeperAt(s, "what came through the door of the "+c.Tok())
		},
		Decline: func(w *World, c *Civ) {
			w.tear(0.6)
			s := w.aWorld(c)
			w.makeTransmitter(s, c.ID, true)
			w.faced(c, "door", "declined", "", s)
		},
	})
	// hold together after losing a war's battle
	def(&Filter{
		Key: "hold", Name: "the strain of war", Levels: []string{"soc"}, Diff: 3, Repeat: true,
		Overcome: func(w *World, c *Civ) {},
		Scar: func(w *World, c *Civ) {
			c.Morale -= 0.5
		},
		Decline: func(w *World, c *Civ) {
			w.faced(c, "hold", "declined", "", -1)
			if !w.civilWar(c) {
				c.Morale -= 1
			}
		},
	})
	def(&Filter{
		Key: "revolt", Name: "Revolt", Levels: []string{"soc"}, Diff: 4, Repeat: true, Domain: "weapons",
		Overcome: func(w *World, c *Civ) {
			m := w.Civs[c.Master]
			c.Master = -1
			c.Vassal = false
			c.Scars[ScarChains] = true
			w.event(KFaced, c, m, -1, P{"filter": "revolt", "outcome": "overcome", "way": ""})
			if !m.Living() && w.Owner[m.Home] < 0 {
				w.Owner[m.Home] = c.ID
				c.Systems = append(c.Systems, m.Home)
				w.faced(c, "revolt", "overcome", "home", m.Home)
			}
		},
		Scar: func(w *World, c *Civ) {
			w.faced(c, "revolt", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			m := w.Civs[c.Master]
			if !m.Living() {
				w.endCiv(c, Extinct, sprintf("fell with their masters the %s", m.Tok()))
			} else {
				w.faced(c, "revolt", "declined", "", -1)
				c.Morale -= 2
			}
		},
	})
}

// neverFaces says whether a filter is one the people's nature puts it
// past: the unconscious never face belief, boredom or ossification.
func (c *Civ) neverFaces(key string) bool {
	for _, k := range c.Species.Profile().NeverFaces {
		if k == key {
			return true
		}
	}
	return false
}

// natureDiff is what the substrate and the modifiers do to each filter,
// from the profile, plus the one rule that reads a people's state: an
// empty field is a slow death for a rider.
func (c *Civ) natureDiff(key string) float64 {
	d := c.Species.Profile().FilterDiff[key]
	if c.Own >= 0 && key == "silence" && c.Hosts <= 1 {
		d++
	}
	return d
}

func init() {
	def(&Filter{
		Key: "faith", Name: "the Wars of Faith", Levels: []string{"soc"}, Diff: 2.5, Domain: "society",
		Overcome: func(w *World, c *Civ) {}, // most peoples manage it; the legends only note the ones that did not
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarChurch] = true
			w.faced(c, "faith", "scarred", "", -1)
			w.churchMorality(c)
		},
		Decline: func(w *World, c *Civ) {
			if w.R.Float64() < 0.5 {
				w.darkAge(c, "tore themselves apart over the nature of god")
			} else {
				w.contract(c, "fought over god until there was nothing left to fight with")
			}
		},
	})
}
