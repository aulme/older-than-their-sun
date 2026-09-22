package species

// parasite: information that rides others, in flesh or in pure form. The
// rider group says which.
var parasite = &SubstrateDef{Sub: Parasite, Entry: Entry{
	Key:    "parasite",
	Legacy: Draw{Base: 0, Tilts: map[string]float64{"swarming": 0, "planetary": 0, "evolver": 0}}, // a plague that woke makes them, never a cradle: see plagues.md
	Draws:  Draw{Base: 0, Tilts: map[string]float64{"swarming": 0.5, "planetary": 0.3, "hive": 0.5, "antimemetic": 2, "evolver": 0.5}},
	Own:    []string{"rider"},
	Profile: Profile{
		Sur:    1,
		Body:   "organic",
		Wis:    0.5, // knows its hosts from inside
		Dom:    M{"biology": 1.4, "society": 1.2, "industry": 0.8},
		Morals: [4]float64{2, 1, 1, 1.5},
	},
}}
