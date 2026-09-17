package species

// antimemetic: known only by the absence of information about it. A stub:
// the odds are the proposal's, perception lands with the anti-memetic step.
var antimemetic = &ModDef{Mod: Antimemetic, Entry: Entry{
	Key:      "antimemetic",
	Portrait: "Nothing that has met them remembers it.",
	Legacy:   Draw{Base: 0},
	Draws:    Draw{Base: 1.0 / 200},
}}
