package species

// unconscious: intelligence with no one home. The organisation group is
// not rolled; the rest of its rules land with the modifier rules.
var unconscious = &ModDef{Mod: Unconscious, Entry: Entry{
	Key:      "unconscious",
	Portrait: "There is no one inside; they only act as if.",
	Legacy:   Draw{Base: 5.0 / 92, Skip: []string{"org"}}, // the old org-group weight, after the hive
	Draws:    Draw{Base: 1.0 / 20, Tilts: map[string]float64{"replicator": 3, "antimemetic": 3}, Skip: []string{"org"}},
	Profile: Profile{
		Mil: -0.5, Sur: 1, Soc: 1.5,
		Wis:        -1.5, // no intuition to see past, and nothing to read another mind with
		Dom:        M{"exotic": 0.6, "society": 0.4, "biology": 1.3},
		Memory:     1.5,
		FilterDiff: map[string]float64{"beacon": -3, "transcend": 2, "machines": -1, "silence": -2},
		Cannot:     Believes | Stiffens, // no idea can take a mind that is not there, and nothing in it sets
	},
}}
