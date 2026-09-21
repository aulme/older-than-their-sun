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

// naming: what each miracle's holders say about the state when they first
// reach into it. The people's word is the %s.
var beneathNames = map[string]string{
	"ftl":       "Whatever the ships pass through, nobody is in it. From inside, the crossing has no duration: one moment here, the next there. The ones who tried to stay awake for it did not come back as one person. The %s call it %s and do not look at it.",
	"ansible":   "The Voice does not cross space; it goes under it. The %s call what it goes under %s. The speakers do not hear it. Anyone who begins to hear it is taken off the line.",
	"foresight": "The Sight is not a looking forward. It is a leaning on something under time, where before and after are one thing. The %s call it %s, and the ones who lean too hard do not come back up.",
	"unmaking":  "What the Unmaking does is not destruction. Matter is put back the way it was before it was matter. The %s have a word for that state, %s, and it is a word they say once.",
	"wound":     "There is a place in the %s where the wall is not. They call what shows through it %s, and it is the only thing they are afraid of.",
}

// name is a people reaching into the state for the first time: the fact
// the names pass coins its word from, once. An heir has its line's word
// already (the pass reads the line), so it writes no fact of its own.
func (w *World) name(c *Civ, key string) {
	if c.Named {
		return
	}
	c.Named = true
	w.told(FWord, c, nil, -1).with(P{"route": key})
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
		w.event(KWallStage, nil, nil, -1, P{"stage": st})
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
		w.event(KLeak, c, nil, -1, P{"mechanism": "shore", "named": c.Named})
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
			w.event(KLeak, c, nil, c.Home, P{"mechanism": "sleeper"})
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
		w.event(KLeak, nil, nil, s, P{"mechanism": "voice"})
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
