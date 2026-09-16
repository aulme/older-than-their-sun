package history

// A filter is anything that may push a civilisation into decline: the Great
// Filter idea, applied at every stage. Each filter is faced once (or, for
// plague, repeatedly) and resolves as overcome, scarred, or declined. Scars
// are lasting traits that alter behaviour and the odds of later filters.

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
	ScarStewardship  = "a stewardship creed"
	ScarNoMachines   = "a prohibition on thinking machines"
	ScarCentralism   = "iron centralism"
	ScarMortality    = "a mortality creed"
	ScarNoSelfCopies = "the law that no machine may make itself"
	ScarStarFear     = "a fear of their own star"
	ScarLeftBehind   = "being the ones left behind"
	ScarQuarantine   = "a quarantine creed"
	ScarOssified     = "ossification"
	BoonAligned      = "aligned minds"
	BoonSwarm        = "swarm industry"
	BoonUnity        = "unity forged in the atomic age"
)

// Filter describes one hurdle.
type Filter struct {
	Key     string
	Name    string
	Repeat  bool // can be faced more than once
	Trigger func(w *World, c *Civ) bool
	Weights [3]float64 // base weights: overcome, scar, decline
	// Temper adjusts the decline weight per temperament.
	Temper   map[Temper]float64
	Overcome func(w *World, c *Civ)
	Scar     func(w *World, c *Civ)
	Decline  func(w *World, c *Civ)
}

var filters []*Filter

func init() {
	filters = []*Filter{
		{
			Key: "atomic", Name: "the Atomic Age",
			Trigger: func(w *World, c *Civ) bool { return c.Tech >= 0.3 },
			Weights: [3]float64{4, 3, 2.5},
			Temper:  map[Temper]float64{Aggressive: 1.5, Zealous: 1.3, Curious: 0.9, Insular: 0.8},
			Overcome: func(w *World, c *Civ) {
				c.Boons[BoonUnity] = true
				w.log("The %s split the atom and, for once, put the weapons away. They are stronger for it.", c.Name)
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
		},
		{
			Key: "overshoot", Name: "Overshoot",
			Trigger: func(w *World, c *Civ) bool { return c.Tech >= 0.6 },
			Weights: [3]float64{4, 3, 2},
			Temper:  map[Temper]float64{Aggressive: 1.3, Curious: 0.8},
			Overcome: func(w *World, c *Civ) {
				w.log("The %s strip %s nearly bare, then learn to live within it.", c.Name, c.HomeName)
			},
			Scar: func(w *World, c *Civ) {
				c.Scars[ScarStewardship] = true
				w.log("The %s nearly kill their world. What they build afterwards is slow, careful, and small.", c.Name)
			},
			Decline: func(w *World, c *Civ) {
				if w.chance(0.6) {
					w.darkAge(c, "exhausted their world")
				} else {
					w.endCiv(c, Extinct, "exhausted their world and starved on it")
				}
			},
		},
		{
			Key: "machines", Name: "Thinking Machines",
			Trigger: func(w *World, c *Civ) bool { return c.Tech >= 1.5 },
			Weights: [3]float64{3, 4, 3},
			Temper:  map[Temper]float64{Curious: 0.8, Zealous: 1.4, Insular: 1.0, Aggressive: 1.2},
			Overcome: func(w *World, c *Civ) {
				c.Boons[BoonAligned] = true
				w.log("The %s build minds greater than their own, and the minds stay. Everything goes faster now.", c.Name)
			},
			Scar: func(w *World, c *Civ) {
				c.Scars[ScarNoMachines] = true
				w.log("The %s build a mind, and it nearly ends them. Thou shalt not make a machine in the likeness of a mind. The law holds for ages.", c.Name)
			},
			Decline: func(w *World, c *Civ) {
				if w.chance(0.3) {
					w.darkAge(c, "pulled the plug on their own machines, too late and at great cost")
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
		},
		{
			Key: "distance", Name: "the Distance",
			Trigger: func(w *World, c *Civ) bool { return len(c.Systems) >= 6 },
			Weights: [3]float64{3, 3, 4},
			Temper:  map[Temper]float64{Insular: 0.7, Zealous: 0.8, Curious: 1.2},
			Overcome: func(w *World, c *Civ) {
				w.log("Light-years and generations pull at the %s. Somehow they stay one people.", c.Name)
			},
			Scar: func(w *World, c *Civ) {
				c.Scars[ScarCentralism] = true
				w.log("The colonies of the %s begin to drift. %s answers with iron. The drift stops. So does much else.", c.Name, c.HomeName)
			},
			Decline: func(w *World, c *Civ) {
				if w.chance(0.6) {
					w.schism(c)
				} else {
					w.contract(c, "watched their colonies become strangers, and then enemies, and then silence")
				}
			},
		},
		{
			Key: "silence", Name: "the Long Silence",
			Trigger: func(w *World, c *Civ) bool {
				return c.Tech >= 1.8 && w.Now-c.Born > 300*w.Cfg.FineStep
			},
			Weights: [3]float64{3, 3, 3},
			Temper:  map[Temper]float64{Zealous: 0.7, Insular: 1.3, Curious: 0.9},
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
		},
		{
			Key: "replication", Name: "Self-Replication",
			Trigger: func(w *World, c *Civ) bool { return c.Tech >= 2 && !c.Scars[ScarNoMachines] },
			Weights: [3]float64{3, 3, 3},
			Temper:  map[Temper]float64{Curious: 0.9, Aggressive: 1.3},
			Overcome: func(w *World, c *Civ) {
				c.Boons[BoonSwarm] = true
				w.log("The %s teach their machines to build themselves, and keep the leash. Their fleets multiply.", c.Name)
			},
			Scar: func(w *World, c *Civ) {
				c.Scars[ScarNoSelfCopies] = true
				w.log("A factory of the %s eats a moon before it is stopped. No machine may make itself. The law is absolute.", c.Name)
			},
			Decline: func(w *World, c *Civ) {
				s := c.Systems[w.R.IntN(len(c.Systems))]
				if w.chance(0.3) {
					w.loseSystem(c, s, "stripped world", "were consumed by their own machines")
					w.darkAge(c, "lost "+w.star(s)+" to their own machines and burned the rest to stop it spreading")
					return
				}
				w.endCiv(c, Extinct, "were consumed by their own machines")
				h := w.spawnHorror(Replicators, s, c.ID)
				w.log("At %s the machines of the %s begin to copy themselves, and do not stop. This is %s.", w.star(s), c.Name, h.Name)
			},
		},
		{
			Key: "stellar", Name: "Stellar Engineering",
			Trigger: func(w *World, c *Civ) bool { return c.Dyson > 0 && c.Tech >= 2.5 },
			Weights: [3]float64{4, 2, 2},
			Temper:  map[Temper]float64{Curious: 1.2, Insular: 0.8},
			Overcome: func(w *World, c *Civ) {
				w.log("The %s reach into %s and it holds. They can move stars now, a little.", c.Name, c.HomeName)
			},
			Scar: func(w *World, c *Civ) {
				c.Scars[ScarStarFear] = true
				w.log("The %s reach into %s and it flares. Millions die. They never touch a star again.", c.Name, c.HomeName)
			},
			Decline: func(w *World, c *Civ) {
				w.log("The %s reach too deep into %s. The star convulses.", c.Name, c.HomeName)
				w.loseSystem(c, c.Home, "wounded star", "broke their own star")
				w.Bio[c.Home] = BioNone
				if len(c.Systems) == 0 || w.chance(0.5) {
					w.endCiv(c, Extinct, "broke their own star")
				} else {
					w.contract(c, "broke their own star and fled to a lesser one")
				}
			},
		},
		{
			Key: "transcend", Name: "Transcendence",
			Trigger: func(w *World, c *Civ) bool { return c.Tech >= 3 },
			Weights: [3]float64{3, 3, 3},
			Temper:  map[Temper]float64{Zealous: 1.4, Insular: 0.8},
			Overcome: func(w *World, c *Civ) {
				w.log("The %s find the door out of the universe, and choose to stay.", c.Name)
			},
			Scar: func(w *World, c *Civ) {
				c.Scars[ScarLeftBehind] = true
				w.log("Most of the %s go through. The ones who stay keep the lights on and stop inventing things.", c.Name)
			},
			Decline: func(w *World, c *Civ) {
				if w.chance(0.4) {
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
		},
		{
			Key: "plague", Name: "Plague", Repeat: true,
			Trigger: func(w *World, c *Civ) bool { return c.Plagued || w.chance(0.0004) },
			Weights: [3]float64{3, 3, 3},
			Temper:  map[Temper]float64{Insular: 0.8, Curious: 1.1},
			Overcome: func(w *World, c *Civ) {
				c.Plagued = false
				w.log("A sickness moves through the worlds of the %s. It passes.", c.Name)
			},
			Scar: func(w *World, c *Civ) {
				c.Plagued = false
				c.Scars[ScarQuarantine] = true
				w.log("A sickness moves through the worlds of the %s. When it is over they seal every door and never fully open them again.", c.Name)
			},
			Decline: func(w *World, c *Civ) {
				c.Plagued = false
				x := w.R.Float64()
				switch {
				case x < 0.4:
					w.darkAge(c, "were hollowed out by sickness")
				case x < 0.7:
					w.contract(c, "were hollowed out by sickness")
				default:
					w.endCiv(c, Extinct, "sickened and died")
				}
			},
		},
	}
}

// faceFilters checks each named filter's trigger and resolves it.
func (w *World) faceFilters(c *Civ) {
	for _, f := range filters {
		if !c.Active() {
			return
		}
		if (c.Faced[f.Key] && !f.Repeat) || !f.Trigger(w, c) || !w.chance(0.05) {
			continue
		}
		c.Faced[f.Key] = true
		w.resolve(c, f.Name, f.Weights, f.Temper[c.Temper], f.Overcome, f.Scar, f.Decline)
	}
}

// resolve rolls an outcome and applies it. Scars make later filters more
// brittle; hazard makes everything worse.
func (w *World) resolve(c *Civ, name string, wt [3]float64, temperMul float64, over, scar, decl func(*World, *Civ)) {
	if temperMul == 0 {
		temperMul = 1
	}
	o, s, d := wt[0], wt[1], wt[2]*temperMul*(1+0.15*float64(len(c.Scars)))*(0.5+0.5*w.Hazard)
	x := w.R.Float64() * (o + s + d)
	var out Outcome
	switch {
	case x < o:
		out = Overcome
		over(w, c)
	case x < o+s:
		out = Scarred
		scar(w, c)
	default:
		out = Declined
		decl(w, c)
	}
	c.Record = append(c.Record, sprintf("%s %s", out, name))
}

// weightOfAges is the ambient filter: age, size, and hazard grind everyone down.
func (w *World) weightOfAges(c *Civ) {
	age := float64(w.Now-max(c.Born, c.Renewed)) / float64(w.Cfg.FineStep)
	p := 0.0006 * (1 + age/2000) * (1 + float64(len(c.Systems))/8) * w.Hazard
	if c.Scars[ScarOssified] {
		p *= 1.5
	}
	if !w.chance(p) {
		return
	}
	w.resolve(c, "the Weight of Ages", [3]float64{3, 3, 4}, 1,
		func(w *World, c *Civ) {
			c.Renewed = w.Now
			w.log("The %s grow old and tired, and then, unexpectedly, young again. A renaissance.", c.Name)
		},
		func(w *World, c *Civ) {
			c.Scars[ScarOssified] = true
			w.log("The %s stop changing. Every year is like the last. It works, for a while.", c.Name)
		},
		func(w *World, c *Civ) {
			x := w.R.Float64()
			switch {
			case x < 0.4:
				w.contract(c, "collapsed under their own weight")
			case x < 0.7:
				w.schism(c)
			default:
				w.endCiv(c, Extinct, "collapsed under their own weight and did not recover")
			}
		})
}

// darkAge is a non-terminal decline: tech and reach are lost. A third one is fatal.
func (w *World) darkAge(c *Civ, why string) {
	c.DarkAges++
	c.Tech = max(0.1, c.Tech*0.6)
	c.Voyages = nil
	lost := 0
	for _, s := range append([]int(nil), c.Systems...) {
		if s != c.Home && w.chance(0.5) {
			w.loseSystem(c, s, "abandoned colony", "")
			lost++
		}
	}
	if c.Tech < 1 {
		c.Stage = Emergent
	} else if c.Stage == Zenith {
		c.Stage = Interstellar
	}
	if c.DarkAges >= 3 {
		w.endCiv(c, Extinct, why+", and a third dark age was one too many")
		return
	}
	if lost > 0 {
		w.log("The %s %s. A dark age follows. %d colonies go silent.", c.Name, why, lost)
	} else {
		w.log("The %s %s. A dark age follows.", c.Name, why)
	}
}

func (w *World) schism(c *Civ) {
	if len(c.Systems) < 2 {
		w.log("Unrest among the %s on %s. It passes, this time.", c.Name, c.HomeName)
		return
	}
	lost := len(c.Systems) / 2
	for i := 0; i < lost; i++ {
		s := c.Systems[w.R.IntN(len(c.Systems))]
		if s != c.Home {
			w.loseSystem(c, s, "abandoned colony", "")
		}
	}
	w.log("Schism among the %s. Half their worlds go dark or go their own way.", c.Name)
}

// Scar and boon effects on behaviour.

func (c *Civ) techMul() float64 {
	m := 1.0
	if c.Scars[ScarNoMachines] {
		m *= 0.75
	}
	if c.Scars[ScarOssified] {
		m *= 0.7
	}
	if c.Scars[ScarLeftBehind] {
		m *= 0.1
	}
	if c.Boons[BoonAligned] {
		m *= 1.3
	}
	return m
}

func (c *Civ) expandMul() float64 {
	m := 1.0
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
	return m
}

func (c *Civ) warMul() float64 {
	m := 1.0
	if c.Scars[ScarAtomicTaboo] {
		m *= 0.4
	}
	if c.Boons[BoonUnity] {
		m *= 1.2
	}
	return m
}
