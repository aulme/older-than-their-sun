package species

// evolver: a people whose shape does not stay put. Under the legacy numbers
// it is the old evolver kind; the drift lands with the evolver step.
var evolver = &ModDef{Mod: Evolver, Entry: Entry{
	Key:      "evolver",
	Portrait: "They shape their own flesh, and breed what they need instead of building it.",
	Flavour:  Flavour{"brood", "vacuum-whale", "grown moon"},
	Legacy:   Draw{Base: 8.0 / 82}, // after the swarm and the planetary rolls
	Draws:    Draw{Base: 1.0 / 12},
	Profile: Profile{
		Sur:        1,
		Dom:        M{"biology": 1.5, "industry": 0.8},
		Env:        1,                                // they change themselves instead of the world
		Upkeep:     M{"biology": 0.5, "industry": 2}, // flesh is cheap to them and metal dear
		FilterDiff: map[string]float64{"brood": -1, "overshoot": 1},
	},
}}
