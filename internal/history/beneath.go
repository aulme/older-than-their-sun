package history

import (
	"math"

	"worldgen/internal/tech"
)

// The state beneath. Every miracle that breaks causality (the Door, the
// Voice, the Sight, the Unmaking) reaches into the same thing: not a place
// but a state, the one the universe is in underneath this one, where
// distance and sequence are not features. Nothing born to this age
// understands it, and the code does not either: there is no name for it
// here, only the words each people coins. Nobody is ever consciously
// inside it. A crossing has no duration from inside; one moment here, the
// next there. Anyone who perceives it as something real is perceiving a
// leak, and that is always bad.
//
// The mechanic is the wall. Using the causal miracles wears it thin, for
// everyone in the field; a thin wall means more comes through: sleepers
// wake, transmitters start, the miracles' own filters bite harder. Nobody
// in the field is told this. The reader is.

// causal miracles are the ones that reach into the state.
var causal = map[string]bool{"ftl": true, "ansible": true, "foresight": true, "unmaking": true}

// wear per thousand years of holding each causal miracle
var wear = map[string]float64{"ftl": 0.0015, "ansible": 0.001, "foresight": 0.0005, "unmaking": 0.001}

// words a people might coin for the state; each people gets its own
var beneathWords = []string{"the Grain", "the Quiet", "the Sea Beneath", "the Underneath", "the Blank", "the Between", "the Still",
	"the Floor", "the Hollow", "the Elsewhere", "the Interval", "the Ground", "the Unplace", "the Undertow", "the Low", "the Deep Water",
	"the Other Side of the Page", "the Back of the Sky", "the White", "the Absence", "the Long Now", "the Nothing", "the Unlit", "the Lull"}

// naming: what each miracle's holders say about the state when they first
// reach into it. The people's word is the %s.
var beneathNames = map[string]string{
	"ftl":       "Whatever the ships pass through, nobody is in it. From inside, the crossing has no duration: one moment here, the next there. The ones who tried to stay awake for it did not come back as one person. The %s call it %s and do not look at it.",
	"ansible":   "The Voice does not cross space; it goes under it. The %s call what it goes under %s. The speakers do not hear it. Anyone who begins to hear it is taken off the line.",
	"foresight": "The Sight is not a looking forward. It is a leaning on something under time, where before and after are one thing. The %s call it %s, and the ones who lean too hard do not come back up.",
	"unmaking":  "What the Unmaking does is not destruction. Matter is put back the way it was before it was matter. The %s have a word for that state, %s, and it is a word they say once.",
	"wound":     "There is a place in the %s where the wall is not. They call what shows through it %s, and it is the only thing they are afraid of.",
}

// name gives a people its word for the state, once, on first reaching in.
func (w *World) name(c *Civ, key string) {
	if c.Word != "" {
		return
	}
	used := map[string]bool{}
	for _, o := range w.Civs {
		used[o.Word] = true
	}
	var free []string
	for _, s := range beneathWords {
		if !used[s] {
			free = append(free, s)
		}
	}
	if len(free) == 0 {
		free = beneathWords
	}
	c.Word = free[w.R.IntN(len(free))]
	if t, ok := beneathNames[key]; ok {
		w.log(t, c.Name, c.Word)
	}
}

// tickBeneath wears the wall and lets things through.
func (w *World) tickBeneath() {
	for _, c := range w.Civs {
		if !c.Living() {
			continue
		}
		for _, k := range c.held() {
			if r, ok := wear[k]; ok && c.Miracles[k] != "born" && c.Miracles[k] != "deepening" { // what is evolved does not wear it, as it does not leak; what an eldritch thing is does not either: the wound is its wear
				w.Thin += r * w.dt
			}
		}
	}
	for _, l := range w.Legacies {
		if l.Kind == Law && (l.State == Mastered || l.State == Wielded) {
			w.Thin += 0.0003 * w.dt
		}
	}
	w.Thin *= math.Max(0, 1-0.002*w.dt) // it heals; half-life about 350 kyr
	if st := w.thinStage(); st > w.ThinStage {
		w.ThinStage = st
		switch st {
		case 1:
			w.log("Something has changed in the field, and nobody in it can say what. Doors are used a little less carefully than they were.")
		case 2:
			w.log("The wall between this and what is under it has worn thin. Things come through more easily now, for everyone, and no one knows to blame anyone.")
		case 3:
			w.log("The wall is torn. What leaks through no longer needs a door.")
		}
	} else if st < w.ThinStage {
		w.ThinStage = st
	}
	if w.Thin < 0.5 || !w.chance(0.0001*w.Thin) {
		return
	}
	w.leak()
}

// thinStage: 0 whole, 1 worn, 2 thin, 3 torn.
func (w *World) thinStage() int {
	switch {
	case w.Thin >= 5:
		return 3
	case w.Thin >= 2:
		return 2
	case w.Thin >= 0.5:
		return 1
	}
	return 0
}

// ThinWord says how the wall stands.
func (w *World) ThinWord() string {
	return []string{"whole", "worn", "thin", "torn"}[w.thinStage()]
}

// leak: something comes through, somewhere.
func (w *World) leak() {
	// a people that reaches in begins to see it as a place
	var holders []*Civ
	for _, c := range w.Civs {
		if !c.Active() {
			continue
		}
		if c.Species.HasPower("wound") {
			holders = append(holders, c) // the wall is open there
			continue
		}
		for _, k := range c.held() {
			if causal[k] && c.Miracles[k] != "born" { // what is evolved does not leak
				holders = append(holders, c)
				break
			}
		}
	}
	x := w.R.Float64()
	switch {
	case x < 0.5 && len(holders) > 0:
		c := holders[w.R.IntN(len(holders))]
		var keys []string
		for _, k := range c.held() {
			if causal[k] && c.Miracles[k] != "born" {
				keys = append(keys, k)
			}
		}
		filter := "door" // what comes through a wound comes as through a door
		if len(keys) > 0 {
			filter = tech.Get(keys[w.R.IntN(len(keys))]).Filter
		}
		word := c.Word
		if word == "" {
			word = "it"
		}
		w.log("Some of the %s begin to see %s as a place, with a shore and a weather. That is never good; it means something is coming through.", c.Name, word)
		delete(c.Faced, filter) // a leak is a second facing
		w.face(c, filter, 1)
	case x < 0.85:
		var sleeping []*Civ
		for _, c := range w.Civs {
			if c.Active() && c.Asleep {
				sleeping = append(sleeping, c)
			}
		}
		if len(sleeping) > 0 {
			c := sleeping[w.R.IntN(len(sleeping))]
			w.log("The wall is thin near %s now, and something that slept there notices.", c.HomeName)
			w.rouse(c, nil)
			return
		}
		fallthrough
	default:
		var free []int
		for i := range w.G.Stars {
			if w.Owner[i] < 0 && i != w.G.Sol {
				free = append(free, i)
			}
		}
		if len(free) == 0 {
			return
		}
		s := free[w.R.IntN(len(free))]
		w.makeTransmitter(s, -1, true)
		w.log("Something speaks from %s in no language, in a voice that did not cross space to get there. It came through.", w.star(s))
	}
}

// thinDiff is what a thin wall does to the causal miracles' own filters.
func (w *World) thinDiff(key string) float64 {
	switch key {
	case "door", "openline", "sight", "unmaking":
		return math.Min(2, 0.3*w.Thin)
	}
	return 0
}

// tear is the wear of one bad moment: a scar or a fall on a causal filter,
// a broken law, a world unmade.
func (w *World) tear(amount float64) { w.Thin += amount }
