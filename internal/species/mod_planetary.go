package species

// planetary: a people that is, or fills, one celestial body. Under the
// legacy numbers it is the old planetary mind: a living ocean, usually.
var planetary = &ModDef{Mod: Planetary, Cradle: "ocean", CradleOdds: 0.6, Entry: Entry{
	Key:      "planetary",
	Portrait: "They are one mind, spread through the living substance of their world.",
	Flavour:  Flavour{"graft", "spore-ark", "living moon"},
	Legacy:   Draw{Base: 4.0 / 86, Tilts: map[string]float64{"evolver": 0}, Skip: []string{"way"}},
	Draws:    Draw{Base: 1.0 / 25, Tilts: map[string]float64{"hive": 3, "unconscious": 2, "replicator": 0.1, "evolver": 0.5, "nomadic": 0.05}, Skip: []string{"org"}},
	Profile: Profile{
		Mil: -1, Sur: 1, Soc: 2,
		Wis:       1,   // slow, and whole
		Reach:     0.3, // a third until it learns to graft
		Rate:      1.3, // one vast mind
		Dom:       M{"propulsion": 0.4, "biology": 1.5, "society": 1.3},
		Memory:    0.15,
		PlagueBio: 2,              // one body to sicken
		Frail:     2,              // and one body to die
		Cannot:    Flees | Fields, // a taken home is the end; it does not farm, it is the world
		Cradle:    2,              // and the world feeds it as no field could
	},
}}
