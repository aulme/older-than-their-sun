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
	ScarOssified     = "ossification"
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
	Domain    string // research pushed while facing it
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
	"memory":        {"silence": 2, "weight": 1, "find": -1},
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
	"individualist": {"distance": 1, "weight": -0.5, "hold": 1},
	"caste":         {"weight": 1, "hold": -0.5},
	"shortlived":    {"silence": -1, "weight": -1},
	"longlived":     {"weight": 1.5, "silence": 1},
	"radiation":     {"atomic": -1},
	"solitary":      {"distance": -2, "beacon": -2, "weight": 0.5, "hold": 1},
	"herd":          {"beacon": 2, "atomic": -1, "distance": 1, "hold": -1},
	"dormancy":      {"cosmic": -1, "dying": -1},
	"symbiosis":     {"machines": -1.5, "replication": -0.5},
	"xenophobic":    {"beacon": -1, "find": 1},
	"submissive":    {"revolt": 1, "hold": -0.5},
	"skyless":       {"cosmic": -1},
	"nomadic":       {"weight": -1, "overshoot": -1, "distance": -3},
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
	again := c.Faced[key]
	c.Faced[key] = true
	master := c.Master
	w.recompute(c)
	if c.miracle("foresight") && key != "sight" && w.R.Float64() < 0.5 {
		w.log("The %s see %s coming and step around it.", c.Name, f.Name)
		c.Record = append(c.Record, "foresaw "+f.Name)
		return Overcome
	}
	lvl := c.level(f.Levels...)
	diff := f.Diff + diffAdj + 0.25*float64(len(c.Scars)) + 1.5*(w.Hazard-1) + c.traitDiff(key) + c.natureDiff(key) + c.miracleDiff(key) + w.lawDiff(key) + w.thinDiff(key)
	roll := w.R.NormFloat64() * 1.5
	margin := lvl + roll - diff
	if f.Domain != "" {
		c.focus(f.Domain, 1.5)
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
// revolt who rose against whom, and for the horrors' filters what came.
func (w *World) recordFilter(c *Civ, key string, f *Filter, out Outcome, master int) {
	if key == "revolt" && master >= 0 {
		m := w.Civs[master]
		if out == Overcome {
			w.fact(FFreed, c, m, c.Home)
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
	var h *Horror
	switch key {
	case "beacon":
		h = w.beacon
	case "incursion":
		h = w.incursionBy
	}
	ff := w.factH(kind, c, c.Home, h)
	ff.What = f.Name
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
		if !c.miracle("ansible") && !c.Has("swarming") && !c.Species.Is(species.Planetary) { // nothing drifts when every world is in the room, or there is no centre, or it is all one mind
			w.face(c, "distance", adj)
		}
	}
	if !c.Active() {
		return
	}
	age := float64(w.Now-max(c.Born, c.Renewed)) / 1000
	lived := float64(w.Now-max(c.Born, c.Ascended)) / 1000 // a miracle makes a people young again
	// the fading of the age weighs on everyone still alive in it
	p := 0.0006 * (age / 2000) * (1 + lived/4000) * (1 + float64(len(c.Systems))/8) * w.Hazard * (1 + 2*(1-w.fertility()))
	if c.Scars[ScarOssified] {
		p *= 1.5
	}
	if w.chance(p) {
		// every renaissance is harder than the last, and age itself weighs
		w.face(c, "weight", lived/2500+0.5*float64(c.Renaissances))
	}
}

func init() {
	def(&Filter{
		Key: "atomic", Name: "the Atomic Age", Levels: []string{"soc"}, Diff: 3.5, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			c.Boons[BoonUnity] = true
			w.log("The %s put the weapons away. They are stronger for it.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarAtomicTaboo] = true
			w.log("The %s burn half of %s before they stop. Ever after, the weapon is unspeakable.", c.Name, c.HomeName)
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
			w.log("The %s strip %s nearly bare, then learn to live within it.", c.Name, c.HomeName)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarStewardship] = true
			w.log("The %s nearly kill their world. What they build afterwards is slow, careful, and small.", c.Name)
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
			w.log("The minds the %s built stay loyal. Everything goes faster now.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarNoMachines] = true
			c.Locked["computation"] = true
			w.log("The mind nearly ends the %s. Thou shalt not make a machine in the likeness of a mind. The law holds for ages.", c.Name)
		},
		Decline: func(w *World, c *Civ) {
			x := w.R.Float64()
			if x < 0.25 {
				w.darkAge(c, "pulled the plug on their own machines, too late and at great cost")
				return
			}
			if x < 0.75 {
				w.machinePeople(c)
				return
			}
			worlds := append([]int(nil), c.Systems...)
			w.endCiv(c, Transformed, "built a mind that outgrew them")
			h := w.spawnHorror(RogueMind, c.Home, c.ID)
			for _, s := range worlds {
				w.horrorTake(h, s)
			}
			c.Into = h.Name
			w.log("The %s are gone. What they built at %s thinks on without them. It is called %s.", c.Name, c.HomeName, h.Name)
		},
	})
	def(&Filter{
		Key: "distance", Name: "the Distance", Levels: []string{"soc"}, Diff: 4.5, Repeat: true, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			w.log("Light-years and generations pull at the %s. Somehow they stay one people.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarCentralism] = true
			w.log("The colonies of the %s begin to drift. %s answers with iron. The drift stops. So does much else.", c.Name, c.HomeName)
		},
		Decline: func(w *World, c *Civ) {
			if w.R.Float64() < 0.6 {
				w.schism(c)
			} else {
				w.contract(c, "watched their colonies become strangers, and then enemies, and then silence")
			}
		},
	})
	def(&Filter{
		Key: "silence", Name: "the Long Silence", Levels: []string{"soc"}, Diff: 4.5, Domain: "society",
		Overcome: func(w *World, c *Civ) {
			w.log("The %s learn to live forever and, against the odds, keep wanting things.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarMortality] = true
			w.log("The %s taste immortality and reject it. Death becomes sacred to them. They are never quite at ease again.", c.Name)
		},
		Decline: func(w *World, c *Civ) {
			w.contract(c, "stopped dying, and then stopped being born")
		},
	})
	def(&Filter{
		Key: "replication", Name: "Self-Replication", Levels: []string{"mil"}, Diff: 5.5, Domain: "weapons",
		Overcome: func(w *World, c *Civ) {
			c.Boons[BoonSwarm] = true
			w.log("The %s keep the leash on their self-building machines. Their fleets multiply.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarNoSelfCopies] = true
			w.log("A factory of the %s eats a moon before it is stopped. No machine may make itself. The law is absolute.", c.Name)
		},
		Decline: func(w *World, c *Civ) {
			s := w.aWorld(c)
			if w.R.Float64() < 0.3 {
				w.loseSystem(c, s, "stripped world", "were consumed by their own machines")
				w.darkAge(c, "lost "+w.star(s)+" to their own machines and burned the rest to stop it spreading")
				return
			}
			w.endCiv(c, Extinct, "were consumed by their own machines")
			h := w.spawnHorror(Replicators, s, c.ID)
			w.log("At %s the machines of the %s begin to copy themselves, and do not stop. This is %s.", w.star(s), c.Name, h.Name)
		},
	})
	def(&Filter{
		Key: "stellar", Name: "Stellar Engineering", Levels: []string{"sur"}, Diff: 7, Domain: "exotic",
		Overcome: func(w *World, c *Civ) {
			w.log("%s holds. The %s can move stars now, a little.", c.HomeName, c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarStarFear] = true
			c.Locked["exotic"] = true
			w.log("%s flares. Millions of the %s die. They never touch a star again.", c.HomeName, c.Name)
		},
		Decline: func(w *World, c *Civ) {
			w.log("The %s reach too deep into %s. The star convulses.", c.Name, c.HomeName)
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
			w.log("The %s find the door out of the universe, and choose to stay.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarLeftBehind] = true
			w.log("Most of the %s go through. The ones who stay keep the lights on and stop inventing things.", c.Name)
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
			w.log("The %s go quiet all at once. Their machines still run. Nobody is home.", c.Name)
		},
	})
	def(&Filter{
		Key: "weight", Name: "the Weight of Ages", Levels: []string{"soc"}, Diff: 5, Repeat: true,
		Overcome: func(w *World, c *Civ) {
			c.Renewed = w.Now
			c.Renaissances++
			c.Morale += 1
			w.log("The %s grow old and tired, and then, unexpectedly, young again. A renaissance.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarOssified] = true
			w.log("The %s stop changing. Every year is like the last. It works, for a while.", c.Name)
		},
		Decline: func(w *World, c *Civ) {
			x := w.R.Float64()
			switch {
			case x < 0.4:
				w.contract(c, "collapsed under their own weight")
			case x < 0.7:
				w.schism(c)
			default:
				w.endCiv(c, Extinct, "collapsed under their own weight and did not recover")
			}
		},
	})
	def(&Filter{
		Key: "door", Name: "the Door", Levels: []string{"soc"}, Diff: 4.5, Domain: "exotic",
		Overcome: func(w *World, c *Civ) {
			w.log("Nothing comes back through the door of the %s that they did not send. This time.", c.Name)
		},
		Scar: func(w *World, c *Civ) {
			c.Scars[ScarDoor] = true
			w.tear(0.3)
			s := w.aWorld(c)
			h := w.spawnHorror(SleeperHorror, s, -1)
			h.Dormant = true
			w.log("Something on the other side of the door notices the %s. %s now sleeps near %s. The %s close the door and speak of it seldom.", c.Name, h.Name, w.star(s), c.Name)
		},
		Decline: func(w *World, c *Civ) {
			w.tear(0.6)
			s := w.aWorld(c)
			h := w.spawnHorror(Beacon, s, c.ID)
			w.log("What came back through the door at %s speaks. It did not cross space to get there. It is called %s.", w.star(s), h.Name)
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
			w.log("Defeat splits the %s.", c.Name)
			w.schism(c)
		},
	})
	def(&Filter{
		Key: "revolt", Name: "Revolt", Levels: []string{"soc"}, Diff: 4, Repeat: true, Domain: "weapons",
		Overcome: func(w *World, c *Civ) {
			m := w.Civs[c.Master]
			c.Master = -1
			c.Vassal = false
			c.Scars[ScarChains] = true
			switch {
			case m.Species.Sub == species.Parasite && m.Has("mindrider"):
				w.log("The %s learn to unthink the %s. What was in their heads is gone, and they are free, and never again quite trust a new idea.", c.Name, m.Name)
			case m.Species.Sub == species.Parasite:
				w.log("The %s find a drug that kills what rides them. They are free, and careful about their blood ever after.", c.Name)
			default:
				w.log("The %s rise against the %s and are free.", c.Name, m.Name)
			}
			if !m.Living() && w.Owner[m.Home] < 0 {
				w.Owner[m.Home] = c.ID
				c.Systems = append(c.Systems, m.Home)
				w.log("The %s take %s, the emptied home of their masters, for their own.", c.Name, m.HomeName)
			}
		},
		Scar: func(w *World, c *Civ) {
			w.log("The %s rise against their masters and are put down. They stay in chains.", c.Name)
		},
		Decline: func(w *World, c *Civ) {
			m := w.Civs[c.Master]
			if !m.Living() {
				w.endCiv(c, Extinct, sprintf("fell with their masters the %s", m.Name))
			} else {
				w.log("The %s rise and are broken. Half of them are killed as an example.", c.Name)
				c.Morale -= 2
			}
		},
	})
}

// natureDiff is what the substrate and the modifiers do to each filter,
// from the profile, plus the one rule that reads a people's state: an
// empty field is a slow death for a rider.
func (c *Civ) natureDiff(key string) float64 {
	d := c.Species.Profile().FilterDiff[key]
	if c.Species.Sub == species.Parasite && (key == "silence" || key == "weight") && c.Hosts <= 1 {
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
			w.log("A century of burning among the %s, and the church wins. It outranks every state after that.", c.Name)
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
