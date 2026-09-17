package species

// machine: minds on hardware, copied and backed up. Machines never arise on
// their own except at a vanishing weight; they are what is left when
// someone builds a mind that outgrows them.
var machine = &SubstrateDef{Sub: Machine, Entry: Entry{
	Key:      "machine",
	Portrait: "They are machines, and do not remember who built them.",
	Arising:  "wake on",
	Flavour:  Flavour{"node", "probe", "array"},
	Legacy:   Draw{Base: 0, Tilts: map[string]float64{"swarming": 0, "planetary": 0, "evolver": 0}},
	Draws:    Draw{Base: 2, Tilts: map[string]float64{"swarming": 0.5, "planetary": 0.5, "hive": 1.5, "unconscious": 1.5, "replicator": 5, "evolver": 0.2}},
	Profile: Profile{
		Mil: 0.5, Sur: 1.5,
		Dom:        M{"computation": 1.3, "biology": 0.6, "industry": 1.1},
		Env:        2,    // rock and vacuum are enough
		Memory:     0.05, // backups on every world
		Endure:     3,    // cold is only cold
		FilterDiff: map[string]float64{"cosmic": -1, "weight": 1, "replication": -1},
		Cannot:     Sickens, // nothing lives in them that they did not put there
	},
}}
