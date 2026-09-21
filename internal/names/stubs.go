package names

import (
	"math/rand/v2"

	"worldgen/internal/history"
)

// The old descriptive generators, kept on the hash stream until the
// translated pass makes these names from grounded recipes. Every row
// they make is marked Stub.

var titleAdj = []string{"Eternal", "Last", "Undying", "Hollow", "Radiant", "Silent", "Grey", "Thousandth", "Sleepless", "Forgotten", "Ashen", "Unbroken", "Deathless", "Weeping"}
var titleNoun = []string{"Emperor", "Regent", "Matriarch", "Hierarch", "Custodian", "Speaker", "Archon", "Sovereign", "Steward", "Oracle", "Warden", "Autarch"}

// title is a ruler title for a contracted remnant.
func title(r *rand.Rand) string { return "the " + pick(r, titleAdj) + " " + pick(r, titleNoun) }

var plagueAdj = []string{"Red", "Grey", "Weeping", "Glass", "Silent", "Sweating", "Hollow", "Slow", "Quick", "Black", "White", "Wandering", "Crystal"}
var plagueBody = []string{"Fever", "Rot", "Blight", "Wasting", "Sleep", "Cough", "Bloom", "Pox", "Fade", "Sweat", "Ague"}
var plagueMind = []string{"Song", "Question", "Doctrine", "Laugh", "Silence", "Certainty", "Word", "Number", "Joke", "Prayer", "Dream", "Argument"}

// plagueName names a sickness: "the Grey Rot" for one of the body, "the
// Silent Question" for one of the mind; one in three for its first host
// instead, "the Qaosh Sweat" or "the Fever of Wolf 359".
func plagueName(r *rand.Rand, memetic bool, host string) string {
	noun := pick(r, plagueBody)
	if memetic {
		noun = pick(r, plagueMind)
	}
	if host != "" && r.IntN(3) == 0 {
		if r.IntN(2) == 0 {
			return "the " + host + " " + noun
		}
		return "the " + noun + " of " + host
	}
	return "the " + pick(r, plagueAdj) + " " + noun
}

// beneathWords are the words a people might coin for the state beneath.
var beneathWords = []string{"the Grain", "the Quiet", "the Sea Beneath", "the Underneath", "the Blank", "the Between", "the Still",
	"the Floor", "the Hollow", "the Elsewhere", "the Interval", "the Ground", "the Unplace", "the Undertow", "the Low", "the Deep Water",
	"the Other Side of the Page", "the Back of the Sky", "the White", "the Absence", "the Long Now", "the Nothing", "the Unlit", "the Lull"}

// finderNames are what a finder calls the makers of an elder work, by
// the kind of the work found.
var finderNames = map[history.LegacyKind][]string{
	history.Artifact:  {"the Ones Who Left Things", "the Makers of Small Suns", "the Toolmakers", "the Careful Dead"},
	history.Structure: {"the Ones Who Moved the Star", "the Makers of the Hollow Sun", "the Builders", "the Ones Who Bent the Dark"},
	history.Threat:    {"the Ones Who Made the Hunger", "the Ones Who Lost Control", "the Warmakers"},
	history.Sleeper:   {"the Ones Who Dream", "the Sleepers", "the Ones Who Would Not Die"},
	history.Law:       {"the Lawgivers", "the Ones Who Changed the Rules"},
	history.Bounty:    {"the Providers", "the Ones Who Left the Table Laid", "the Gardeners"},
	history.Field:     {"the Ones Who Fell Here", "the Lost Fleet"},
}

// ruinNames are what a finder calls the makers of a remain of this age
// when it does not know who they were.
var ruinNames = []string{"the Old Builders", "the Ones Before", "the First People", "the Builders of the Halls", "the Ones Who Left the Lights On"}

func ordinal(n int) string {
	switch n {
	case 2:
		return "second"
	case 3:
		return "third"
	case 4:
		return "fourth"
	case 5:
		return "fifth"
	case 6:
		return "sixth"
	}
	return itoa(n) + "th"
}
