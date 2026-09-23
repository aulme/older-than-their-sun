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

// Scar and boon keys: rows of data/scars.json and data/boons.json, which
// say what each is; the view says them through scarName and boonName.
const (
	ScarAtomicTaboo  = "atomic_taboo"
	ScarChurch       = "church"
	ScarStewardship  = "stewardship"
	ScarNoMachines   = "no_machines"
	ScarCentralism   = "centralism"
	ScarMortality    = "mortality"
	ScarNoSelfCopies = "no_self_copies"
	ScarFlesh        = "flesh"
	ScarStarFear     = "star_fear"
	ScarLeftBehind   = "left_behind"
	ScarQuarantine   = "quarantine"
	ScarBurningSky   = "burning_sky"
	ScarSignal       = "signal"
	ScarDoor         = "door"
	ScarChains       = "chains"
	ScarFatalism     = "fatalism"
	ScarOtherVoices  = "other_voices"
	ScarChanged      = "changed"
	BoonAligned      = "aligned"
	BoonSwarm        = "swarm"
	BoonUnity        = "unity"
	BoonCommunion    = "communion"
	BoonStarKept     = "star_kept"
)

// Filter describes one hurdle: its numbers are its row in
// data/filters.json, its outcomes are code.
type Filter struct {
	Key, Name string
	Levels    []string // "mil", "sur", "soc"; averaged
	Diff      float64
	Repeat    bool
	Domain    string                                                                // research pushed while facing it
	Wreckage  *Wreckage                                                             // what a decline does to the works of the fallen; nil for the default
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

// def registers a filter's outcomes under its key and reads its numbers
// from the table.
func def(f *Filter) {
	d := tables.filters[f.Key]
	if d == nil {
		panic("history: filter " + f.Key + " is not in data/filters.json")
	}
	f.Name, f.Levels, f.Diff, f.Repeat, f.Domain = d.Name, d.Levels, d.Diff, d.Repeat, d.Domain
	if d.Wreckage != nil {
		f.Wreckage = &Wreckage{d.Wreckage.Destroy, conditionOf(d.Wreckage.Leave)}
	}
	filters[f.Key] = f
}

// traitDiff is the asymmetry: how each trait changes each filter's
// difficulty, from the table.
var traitDiff = tables.traitDiff

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
		c.Record = append(c.Record, Record{Kind: "foresaw", Filter: key, Legacy: -1})
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
	w.wreck = wreckOf(key)
	defer func() { w.wreck = nil }()
	switch {
	case margin >= 0.5:
		out = Overcome
		f.Overcome(w, c)
	case margin >= -2:
		out = Scarred
		c.Morale -= 0.5
		f.Scar(w, c)
		if c.Aloft && c.Active() && w.R.Float64() < 0.2 {
			w.rest(c, because("filter_"+key))
		}
	default:
		out = Declined
		c.Morale -= 1
		c.Declines++
		f.Decline(w, c)
	}
	c.Record = append(c.Record, Record{Kind: "faced", Filter: key, Outcome: out, Narrow: math.Abs(margin) < 0.5, Legacy: -1})
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
		adj := 0.3*float64(len(c.Systems)-6) + w.Cfg.Tuning.Continuity.Distance*w.doublings(c) // nobody left who remembers why the colonies are ours
		c.NextDrift *= 2
		if !c.miracle("ansible") && !c.Has("swarming") { // nothing drifts when every world is in the room, or there is no centre
			w.face(c, "distance", adj)
		}
	}
	w.tickStiff(c) // the ways setting, and the filter that comes for whoever survives the rest: see ossify.go
}

func init() {
	def(&Filter{
		Key: "atomic",
		Overcome: func(w *World, c *Civ) {
			w.boon(c, BoonUnity)
			w.faced(c, "atomic", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarAtomicTaboo)
			w.faced(c, "atomic", "scarred", "", c.Home)
		},
		Decline: func(w *World, c *Civ) {
			x := w.R.Float64()
			switch {
			case x < 0.5:
				w.darkAge(c, because("atomic"))
			case x < 0.8:
				w.endCiv(c, Extinct, because("atomic_afternoon"))
			default:
				w.contract(c, because("atomic_ash"))
			}
		},
	})
	def(&Filter{
		Key: "overshoot",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "overshoot", "overcome", "", c.Home)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarStewardship)
			w.faced(c, "overshoot", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			if w.R.Float64() < 0.6 {
				w.darkAge(c, because("overshoot"))
			} else {
				w.endCiv(c, Extinct, because("overshoot_starved"))
			}
		},
	})
	def(&Filter{
		Key: "machines",
		Overcome: func(w *World, c *Civ) {
			w.boon(c, BoonAligned)
			w.faced(c, "machines", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarNoMachines)
			c.Locked["computation"] = true
			w.faced(c, "machines", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			x := w.R.Float64()
			if x < 0.25 {
				w.darkAge(c, because("machines_unplugged"))
				return
			}
			w.machinePeople(c) // every time it does not pull the plug: what it built thinks on without it
		},
	})
	def(&Filter{
		Key: "distance",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "distance", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarCentralism)
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
			w.contract(c, because("distance_silence"))
		},
	})
	def(&Filter{
		Key: "silence",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "silence", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarMortality)
			w.faced(c, "silence", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			w.contract(c, because("silence_unborn"))
		},
	})
	def(&Filter{
		Key: "replication",
		Overcome: func(w *World, c *Civ) {
			w.boon(c, BoonSwarm)
			w.faced(c, "replication", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarNoSelfCopies)
			w.faced(c, "replication", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			s := w.aWorld(c)
			if w.R.Float64() < 0.3 {
				w.loseSystem(c, s, "stripped", because("machines_consumed"))
				w.darkAge(c, because("machines_burned").At(s))
				return
			}
			// the eaten world is the new people's, and the old people are gone
			w.faced(c, "replication", "declined", "", s)
			w.loseSystem(c, s, "stripped", because("machines_consumed"))
			w.endCiv(c, Extinct, because("machines_consumed"))
			if nc := w.replicatorAt(s, species.Machine, species.MadeBy("copies", c.ID), false); nc != nil {
				nc.Species.Parent = c.Species
				w.event(KNamedItself, nc, nil, s, P{"way": "eats", "traits": nc.Species.TraitKeys()})
			}
		},
	})
	def(&Filter{
		Key: "stellar",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "stellar", "overcome", "", c.Home)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarStarFear)
			c.Locked["exotic"] = true
			w.faced(c, "stellar", "scarred", "", c.Home)
		},
		Decline: func(w *World, c *Civ) {
			w.faced(c, "stellar", "declined", "", c.Home)
			w.loseSystem(c, c.Home, "wounded_star", because("star_broken"))
			w.setBio(c.Home, BioNone)
			if len(c.Systems) == 0 || w.R.Float64() < 0.5 {
				w.endCiv(c, Extinct, because("star_broken"))
			} else {
				w.contract(c, because("star_broken_fled"))
			}
		},
	})
	def(&Filter{
		Key: "transcend",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "transcend", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarLeftBehind)
			w.faced(c, "transcend", "scarred", "", -1)
		},
		Decline: func(w *World, c *Civ) {
			if w.R.Float64() < 0.4 {
				w.contract(c, because("transcend_most"))
				return
			}
			for _, s := range c.Systems {
				w.trace(s, "silent_machinery", c)
			}
			w.endCiv(c, Transformed, because("transcend"))
			c.Into = "left"
			w.faced(c, "transcend", "declined", "", -1)
		},
	})
	def(&Filter{
		Key: "door",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "door", "overcome", "", -1)
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarDoor)
			w.tear(0.3)
			s := w.sleeperStar(c)
			if s < 0 {
				w.faced(c, "door", "scarred", "noticed", -1)
				return
			}
			w.faced(c, "door", "scarred", "came", s)
			w.sleeperAt(s, species.MadeBy("door", c.ID))
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
		Key:      "hold",
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
		Key: "revolt",
		Overcome: func(w *World, c *Civ) {
			m := w.Civs[c.Master]
			w.setMaster(c, -1, false)
			w.scar(c, ScarChains)
			w.event(KFaced, c, m, -1, P{"filter": "revolt", "outcome": "overcome", "way": ""})
			if !m.Living() && w.Owner[m.Home] < 0 {
				w.setOwner(m.Home, c.ID)
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
				w.endCiv(c, Extinct, because("with_masters").By(m))
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
		Key:      "faith",
		Overcome: func(w *World, c *Civ) {}, // most peoples manage it; the legends only note the ones that did not
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarChurch)
			w.faced(c, "faith", "scarred", "", -1)
			w.churchMorality(c)
		},
		Decline: func(w *World, c *Civ) {
			if w.R.Float64() < 0.5 {
				w.darkAge(c, because("faith"))
			} else {
				w.contract(c, because("faith_war"))
			}
		},
	})
}
