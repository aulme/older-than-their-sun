package species

// biological: mobile bodies, many members, conscious or not. The standard
// people as it stands, with nothing added; the swarm is its swarming trait.
var biological = &SubstrateDef{Sub: Biological, Deep: 55, Entry: Entry{
	Key:     "biological",
	Flavour: Flavour{"colony", "colony ship", "station"},
	Legacy:  Draw{Base: 93}, // the old standard, swarm, planetary mind and evolver together
	Draws:   Draw{Base: 92},
	Profile: Profile{Body: "organic"},
}}
