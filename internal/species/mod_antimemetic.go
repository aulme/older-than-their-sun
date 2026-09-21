package species

// antimemetic: known only by the absence of information about it. Nothing
// conscious can hold it in mind, live or in records: no conscious people
// meets it, sights its fleets, reads its worlds as anything but empty or
// receives its messages, and its deeds against a conscious people are
// facts with no doer. It perceives everyone. The unconscious perceive it
// as anyone; a conscious people that holds Antimemetic Resilience does
// too, for as long as it holds it. It can be fought by deduction, from
// the shape of the hole its losses leave: see history's antimemetic.go
// and gap.go.
var antimemetic = &ModDef{Mod: Antimemetic, Entry: Entry{
	Key:      "antimemetic",
	Portrait: "Nothing that has met them remembers it.",
	Legacy:   Draw{Base: 0},
	Draws:    Draw{Base: 1.0 / 200},
	Profile: Profile{
		Cannot: Believes, // nothing about it can be in a mind, its own ideas included: no memetic plague takes
	},
}}
