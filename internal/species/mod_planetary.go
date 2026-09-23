package species

// planetary: a people that is, or fills, one celestial body: a living
// ocean, a conscious sun, a station the size of a moon run by a mind. It
// cannot move: its neighbourhood is its reach, fixed and small, and
// within it it is godlike (the body as guns over every world, its strikes
// faced as the waking: history's waking.go); outside it it does nothing.
// Numbers small, no society, makes nothing, and a taken home is the end.
var planetary = &ModDef{Mod: Planetary, Cradle: "ocean", CradleOdds: 0.6, Entry: Entry{
	Key:    "planetary",
	Legacy: Draw{Base: 4.0 / 86, Tilts: map[string]float64{"evolver": 0}, Skip: []string{"way"}},
	Draws:  Draw{Base: 1.0 / 25, Tilts: map[string]float64{"hive": 3, "unconscious": 2, "replicator": 0.1, "evolver": 0.5, "nomadic": 0.05}, Skip: []string{"org"}},
	Profile: Profile{
		Mil: -1, Sur: 1, Soc: 2,
		Wis:           1,   // slow, and whole
		Rate:          1.3, // one vast mind
		Dom:           M{"propulsion": 0.4, "biology": 1.5, "society": 1.3},
		PlagueBio:     2,  // one body to sicken
		Frail:         2,  // and one body to die
		Neighbourhood: 12, // its reach, fixed and small
		Worlds:        4,  // a handful
		Expand:        0.1,
		HomeDefence:   4,                                           // the body over every world it holds, at several times Military
		Cannot:        Flees | Fields | Works | Launches | Reseats, // a taken home is the end; it does not farm or build, it is the world; it sends nothing
		Cradle:        2,                                           // and the world feeds it as no field could
		Morals:        [4]float64{3, 0, 0, 2},
	},
}}
