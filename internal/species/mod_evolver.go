package species

// evolver: a people whose shape does not stay put: it breeds what it needs,
// and what it is drifts. Under the legacy numbers it is the old evolver
// kind; the drift is history's evolver.go: at a slow rate a bio, sense or
// world trait is gained, lost or replaced, every world type held long
// enough adds that world's trait, each drift raises the difference those
// who fathomed it feel, and a drift while a biological plague rages is a
// cure roll.
var evolver = &ModDef{Mod: Evolver, Entry: Entry{
	Key:    "evolver",
	Legacy: Draw{Base: 8.0 / 82}, // after the swarm and the planetary rolls
	Draws:  Draw{Base: 1.0 / 12},
	Profile: Profile{
		Sur:        1,
		Dom:        M{"biology": 1.5, "industry": 0.8},
		Env:        1,                                // they change themselves instead of the world
		Upkeep:     M{"biology": 0.5, "industry": 2}, // flesh is cheap to them and metal dear
		FilterDiff: map[string]float64{"brood": -1, "overshoot": 1},
		Stiffen:    0.7, // nothing in them stays put, their institutions included
	},
}}
