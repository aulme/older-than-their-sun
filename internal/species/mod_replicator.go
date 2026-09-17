package species

// replicator: a thing that makes more of itself out of what it finds. A
// stub: the odds are the proposal's, the conversion and the dormancy land
// with the replicator step.
var replicator = &ModDef{Mod: Replicator, Entry: Entry{
	Key:      "replicator",
	Portrait: "They make more of themselves out of whatever they find.",
	Legacy:   Draw{Base: 0},
	Draws:    Draw{Base: 1.0 / 100, Tilts: map[string]float64{"evolver": 2, "dormancy": 10}},
}}
