package species

// machine: minds on hardware, copied and backed up. Machines never arise on
// their own except at a vanishing weight; they are what is left when
// someone builds a mind that outgrows them. Because their minds are
// copied, a dark age forgets half as much for them and a shattering
// leaves each shard the whole tree: the backups are on every world.
var machine = &SubstrateDef{Sub: Machine, Deep: 20, Entry: Entry{
	Key:    "machine",
	Legacy: Draw{Base: 0, Tilts: map[string]float64{"swarming": 0, "planetary": 0, "evolver": 0}},
	Draws:  Draw{Base: 2, Tilts: map[string]float64{"swarming": 0.5, "planetary": 0.5, "hive": 1.5, "unconscious": 1.5, "replicator": 5, "evolver": 0.2}},
	Profile: Profile{
		Mil: 0.5, Sur: 1.5,
		Body:            "metal",
		Dom:             M{"computation": 1.3, "biology": 0.6, "industry": 1.1},
		Env:             2,    // rock and vacuum are enough
		Memory:          0.05, // backups on every world
		Endure:          3,    // cold is only cold
		FilterDiff:      map[string]float64{"cosmic": -1, "replication": -1},
		Stiffen:         1.5,              // machines ossify
		Forgets:         0.5,              // and forget half as much: the minds are backed up
		Backups:         true,             // on every world, so a shattering leaves each shard the whole tree
		Cannot:          Sickens | Fields, // nothing lives in them that they did not put there; nothing to farm
		PlagueMeme:      2,                // they copy exactly
		OrganicAsEnergy: true,             // what flesh pays in food, they pay in power
		Morals:          [4]float64{2, 1, 1, 2},
	},
}}
