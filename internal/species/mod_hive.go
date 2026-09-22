package species

import "worldgen/internal/mind"

// hive: the whole people is one mind, with or without a queen. The seat
// group says which; the organisation group is not rolled. No civil war
// and no stiffness ever; prone to sundering: the Distance's decline for a
// people with no factions is a world cut from the seat as a people of its
// own (history's sunder.go). Memetic plagues at twice the odds: one mind is
// one infection. Born to the voice at five times the rate.
var hive = &ModDef{Mod: Hive, Entry: Entry{
	Key:    "hive",
	Legacy: Draw{Base: 10.0 / 102, Tilts: map[string]float64{"unconscious": 0}, Skip: []string{"org"}}, // the old org-group weight
	Draws:  Draw{Base: 1.0 / 10, Tilts: map[string]float64{"unconscious": 2, "replicator": 3, "born_voice": 5}, Skip: []string{"org"}},
	Own:    []string{"seat"},
	Profile: Profile{
		Mil: 0.5, Soc: 2.5,
		Dom:        M{"society": 0.6, "computation": 0.8},
		Memory:     0.7, // one memory
		Dials:      mind.Dials{Loyalty: 0.1, Fear: -0.1},
		FilterDiff: map[string]float64{"distance": -3, "beacon": 3, "silence": -1, "machines": -1},
		Cannot:     CivilWars | Stiffens, // a hive has no factions, and no institutions to set
		PlagueMeme: 2,                    // one mind: what takes it takes all of it
		Morals:     [4]float64{2, 0.04, 3, 1},
	},
}}
