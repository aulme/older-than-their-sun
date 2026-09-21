package species

import "worldgen/internal/mind"

// replicator: a thing that makes more of itself out of what it finds: a
// machine slurry, an insect swarm that turns all flesh into itself. It
// lives on its abilities: research at a fifth of anyone's, self-replication
// innate, the Brood never faced. It grows by eating: every world it holds
// turns what its body is made of into ships and nothing else, and a world
// it takes in war is stripped and held empty (history's replicator.go).
// It trades nothing, since it has nothing to give but itself, makes peace
// only by exhaustion, and when its will is spent it goes quiet: it sleeps,
// and wakes when someone settles inside its reach or the Find unleashes
// it. A monster to everyone by rule; usually a hive and often unconscious.
var replicator = &ModDef{Mod: Replicator, Entry: Entry{
	Key:      "replicator",
	Portrait: "They make more of themselves out of whatever they find.",
	Flavour:  Flavour{"growth", "spore", "mass"},
	Legacy:   Draw{Base: 0},
	Draws: Draw{Base: 1.0 / 100, Tilts: map[string]float64{"evolver": 2, "dormancy": 10, "hive": 4, "unconscious": 3,
		// its stance is its appetite: it comes for what it can reach, and never for terms
		"conqueror": 20, "opportunist": 3, "unyielding": 3, "vengeful": 0.5, "defensive": 0.1, "pacifist": 0, "submissive": 0, "confederate": 0,
		"expansionist": 3, "contemplative": 0}},
	Profile: Profile{
		Mil: 1, Soc: -1,
		Rate:       0.2,
		Range:      12,  // it crawls to what it can reach, ships or no ships
		Expand:     1.5, // it spreads: what it lands on it eats
		Env:        3,   // rock and vacuum are food enough
		Innate:     []string{"self_replication"},
		NeverFaces: []string{"brood", "replication"},
		Cannot:     Trades | Pays | Believes | HoldsGrudges | Stiffens | CivilWars, // nothing to give, nothing to feed, nothing an idea or a wrong could take hold of, no institutions to set and no factions to split
		Eats:       true,
		Monster:    true,
		NoTerms:    true,
		Dormant:    true,
		Dials:      mind.Dials{Aggression: 0.5, Greed: 0.5, Hate: 0.5},
		Morals:     [4]float64{5, 0, 0, 0}, // there is nothing in it to hold a wrong
	},
}}
