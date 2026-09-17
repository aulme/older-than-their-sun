package species

// parasite: information that rides others, in flesh or in pure form. The
// rider group says which.
var parasite = &SubstrateDef{Sub: Parasite, Entry: Entry{
	Key:      "parasite",
	Portrait: "They are a parasite, and need the bodies of others to think and to build.",
	Arising:  "first ride on",
	Flavour:  Flavour{"host-world", "carrier", "hive"},
	Legacy:   Draw{Base: 4, Tilts: map[string]float64{"swarming": 0, "planetary": 0, "evolver": 0}},
	Draws:    Draw{Base: 0, Tilts: map[string]float64{"swarming": 0.5, "planetary": 0.3, "hive": 0.5, "antimemetic": 2, "evolver": 0.5}}, // born-infected rolls make them, never a cradle
	Own:      []string{"rider"},
	Profile: Profile{
		Sur: 1,
		Dom: M{"biology": 1.4, "society": 1.2, "industry": 0.8},
	},
}}
