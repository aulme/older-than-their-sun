package species

// eldritch: no clear biology and no clear hardware; one or a few of a
// kind. It works in ways the sim does not model, so most of the sim's
// machinery is absent for it and what remains is strange: no fields, no
// works, no upkeep, no trade, no ships and no tree. In place of research
// it draws powers from the pool (pool.go): a few at birth, more as the
// ages pass. Everyone finds it other. Its reach is not a ship's but a
// range it feels, and another of it is simply there (history's
// eldritch.go). It is never young: it reads as interstellar from the
// first, and as exotic once it holds an exotic power.
var eldritch = &SubstrateDef{Sub: Eldritch, Deep: 25, Entry: Entry{
	Key:      "eldritch",
	Portrait: "They are made of nothing that has a name: not flesh, not machine, and no one who has looked can say what.",
	Arising:  "are first noticed on",
	Flavour:  Flavour{"presence", "second presence", "locus"},
	Legacy:   Draw{Base: 0, Tilts: map[string]float64{"swarming": 0, "planetary": 0, "evolver": 0}},
	Draws: Draw{Base: 6,
		Tilts:  map[string]float64{"swarming": 0.3, "planetary": 15, "hive": 2, "unconscious": 10, "replicator": 0.2, "antimemetic": 5, "evolver": 0.1},
		Groups: map[string]float64{"power": 10, "sense": 2.5}}, // a miracle from birth at ten times the rate; a sense always, from its own table
	Profile: Profile{
		Mil: 1, Sur: 3, Soc: 1,
		Wis:         1,   // the hardest thing to understand, and it reads minds it cannot model: the same term back
		Range:       10,  // what it feels, before the powers widen it
		Env:         3,   // it is wherever it is: rock, vacuum, a dead star
		Era:         3,   // never young
		Alien:       3,   // everyone finds it other
		HomeDefence: 1.5, // everything about it resists
		Expand:      0.1, // most eldritch peoples hold one world all their lives
		Cannot:      Fields | Works | Trades | SettlesByShip | Launches | Sickens | Believes | Researches | Pays,
		Morals:      [4]float64{3, 1, 1, 2},
	},
}}
