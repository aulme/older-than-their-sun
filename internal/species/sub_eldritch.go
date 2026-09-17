package species

// eldritch: no clear biology and no clear hardware; one or a few of a kind.
// A stub until the pool of powers lands: the odds and the flags are the
// proposal's, the behaviour behind the flags is not yet read.
var eldritch = &SubstrateDef{Sub: Eldritch, Entry: Entry{
	Key:      "eldritch",
	Portrait: "They are made of nothing that has a name: not flesh, not machine, and no one who has looked can say what.",
	Arising:  "are first noticed on",
	Flavour:  Flavour{"presence", "second presence", "locus"},
	Legacy:   Draw{Base: 0, Tilts: map[string]float64{"swarming": 0, "planetary": 0, "evolver": 0}},
	Draws: Draw{Base: 6,
		Tilts:  map[string]float64{"swarming": 0.3, "planetary": 15, "hive": 2, "unconscious": 10, "replicator": 0.2, "antimemetic": 5, "evolver": 0.1},
		Groups: map[string]float64{"power": 10}},
	Profile: Profile{Cannot: Fields | Works | Trades | SettlesByShip | Sickens},
}}
