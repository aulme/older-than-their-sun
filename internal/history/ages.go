package history

import (
	"math"

	"worldgen/internal/tech"
)

// The age generator. Earlier ages are myth: one tick per rise and fall,
// elder civilisations with a portrait instead of traits, and legacies left
// on the substrate for the current age to find.

var elderPortraits = []string{
	"something that thought in the convection cells of a red giant",
	"a mind spread through the magnetic field of a nebula",
	"a people who lived in the dark between stars and never saw a sun up close",
	"a single organism the size of a moon, slow as geology",
	"a civilisation of machines whose makers were already a myth to them",
	"a chorus that existed only while it was being sung",
	"something that used stars the way others use fire",
	"a species that kept its dead and let them vote",
	"a swarm that thought only while falling into a gravity well",
	"minds that ran on the decay of heavy elements and were therefore very patient",
	"a people who folded themselves into a smaller dimension to save on entropy",
	"a thing that was mostly a question",
}

var ageEnders = []string{
	"a beacon that spoke to every mind still listening",
	"a wave of self-copying machines let loose by someone with no successors to stop it",
	"a passage through a dense arm, and a rain of supernovae on worlds already half empty",
	"a war between the last two powers, fought over nothing either still needed",
	"a long silence, in which the remaining few forgot each other",
	"the slow failure of everything that had been built to outlast its builders",
}

var ageKnowers = []string{
	"It learned that the galaxy had done this before, and would again.",
	"It counted the dead ages and knew its own for what it was.",
	"It measured the fading and built accordingly, for no one.",
}

var structureDescs = []string{
	"a star that was moved, and the trail it left",
	"a black hole set in a ring of black metal",
	"a hollowed sun, lit from inside",
	"a corridor of darkness where no light crosses",
	"a world with a machined core that still hums",
	"a lattice of threads around a dead star",
	"a moon that is a single instrument, tuned to something",
}

var artifactDescs = []string{
	"a seed of grey metal that is warm to the touch",
	"a lens that shows the sky as it will be",
	"an engine with no fuel and no moving parts",
	"a library written in the arrangement of atoms",
	"a weapon shaped like a musical instrument",
	"a door that opens onto the same room",
	"a cradle for something that was never put in it",
}

var lawDescs = []string{
	"a region where minds do not work",
	"a standing signal that repeats one number",
	"a corridor along which ships arrive before they leave",
}

// artifact nodes: which discoveries an elder artifact can stand for
var artifactNodes = []string{
	"fusion", "machine_minds", "self_replication", "antimatter", "relativistic", "relativistic_weapons",
	"terraforming", "memetics", "dyson", "stellar_engineering", "ansible", "wormhole_physics",
	"ftl", "planck_weapons", "transcendence", "star_lifting", "directed_evolution",
}

var structureNodes = []string{"stellar_engineering", "dyson", "wormhole_physics", "star_lifting"}

// runAges writes the myth. Each earlier turn of the cycle is an age: a
// surge of elder civilisations that fades as the galaxy's fertility fades,
// with a cosmic event to sweep up what is left. The current age's surge is
// the last one and belongs to the engine.
func (w *World) runAges() {
	surges := w.Cycle.Surges
	for i, s := range surges[:len(surges)-1] {
		age := &AgeRecord{Index: i, Start: s, End: w.ageEnd(s)}
		w.Ages = append(w.Ages, age)
		w.logAt(age.Start, "The dawn of an age. Everywhere at once, things start to think.")
		nElders := 2 + w.R.IntN(4)
		for j := 0; j < nElders; j++ {
			e := &Elder{Age: i, Portrait: elderPortraits[w.R.IntN(len(elderPortraits))]}
			// the earlier in the age, the likelier to rise: fertility is falling
			span := float64(age.End - age.Start)
			e.Rose = age.Start + Year(span*w.R.Float64()*w.R.Float64()*0.8)
			e.Fell = e.Rose + Year(w.R.Float64()*float64(age.End-e.Rose)*1.3)
			age.Elders = append(age.Elders, e)
			w.logAt(e.Rose, "Somewhere, %s rises.", e.Portrait)
			if w.R.Float64() < 0.2 {
				w.logAt(e.Rose+Year(float64(e.Fell-e.Rose)*0.6), "%s", ageKnowers[w.R.IntN(len(ageKnowers))])
			}
			nLeg := 1 + w.R.IntN(4)
			for k := 0; k < nLeg; k++ {
				at := e.Rose + Year(w.R.Float64()*float64(e.Fell-e.Rose))
				// what is left erodes with deep time
				if w.R.Float64() < math.Exp(-float64(-at)/5e9) {
					w.leaveLegacy(e, at)
				}
			}
			if e.Fell > age.End {
				w.logAt(e.Fell, "It ends, the last of its age, long after the others.")
			} else if w.R.Float64() < 0.35 {
				w.logAt(e.Fell, "It is gone. Its works remain.")
			} else {
				w.logAt(e.Fell, "It ends.")
			}
		}
		age.Ender = ageEnders[w.R.IntN(len(ageEnders))]
		w.logAt(age.End, "The age wanes. Nothing new rises, and what remains dwindles. What is left is swept up by %s.", age.Ender)
	}
}

func (w *World) leaveLegacy(e *Elder, at Year) {
	s := w.R.IntN(len(w.G.Stars))
	if s == w.G.Sol {
		return
	}
	l := &Legacy{ID: len(w.Legacies), Age: e.Age, Elder: e, Maker: -1, Star: s, Horror: -1, Finder: -1, Cond: Condition(w.R.IntN(2))}
	x := w.R.Float64()
	switch {
	case x < 0.35:
		l.Kind = Artifact
		l.Node = artifactNodes[w.R.IntN(len(artifactNodes))]
		l.Desc = artifactDescs[w.R.IntN(len(artifactDescs))]
	case x < 0.6:
		l.Kind = Structure
		l.Node = structureNodes[w.R.IntN(len(structureNodes))]
		l.Desc = structureDescs[w.R.IntN(len(structureDescs))]
		if !w.G.Stars[s].Dead() && w.R.Float64() < 0.5 {
			// structures like dead stars
			for _, t := range w.G.Near(s, 40) {
				if w.G.Stars[t].Dead() {
					s, l.Star = t, t
					break
				}
			}
		}
	case x < 0.75:
		l.Kind = Threat
		w.Now = at
		if w.R.Float64() < 0.5 {
			h := w.spawnHorror(Replicators, s, -1)
			h.Dormant = true
			h.Legacy = l.ID
			l.Horror = h.ID
			l.Node = "self_replication"
			l.Desc = "machines that sleep in the rubble of " + w.star(s)
		} else {
			h := w.spawnHorror(Beacon, s, -1)
			h.Dormant = w.R.Float64() < 0.5
			h.Legacy = l.ID
			l.Horror = h.ID
			l.Node = "memetics"
			if h.Dormant {
				l.Desc = "a transmitter at " + w.star(s) + ", silent"
			} else {
				l.Desc = "a transmitter at " + w.star(s) + " that has never stopped"
				l.State = Unleashed
			}
		}
	case x < 0.92:
		l.Kind = Sleeper
		w.Now = at
		h := w.spawnHorror(SleeperHorror, s, -1)
		h.Dormant = true
		h.Legacy = l.ID
		l.Horror = h.ID
		l.Desc = "something that withdrew into the dark near " + w.star(s) + " and went still"
	default:
		l.Kind = Law
		l.Desc = lawDescs[w.R.IntN(len(lawDescs))]
		l.Node = "ftl"
	}
	w.Legacies = append(w.Legacies, l)
	e.Legacies = append(e.Legacies, l)
	w.logAt(at, "It leaves %s.", l.Desc)
}

// legacyNode returns the tech node an artifact or structure stands for.
func (l *Legacy) node() *tech.Node { return tech.Get(l.Node) }

// runDeep is the substrate-only pass over the interregna and earlier ages:
// life arising, gamma-ray bursts, scheduled star deaths.
func (w *World) runDeep() {
	w.dt = float64(w.Cfg.DeepStep) / 1000
	for y := w.Cfg.DeepStart; y < w.Cfg.Dawn; y += w.Cfg.DeepStep {
		w.Now = y
		w.deepLife()
		w.deepBurst()
		w.deepStars(y, y+w.Cfg.DeepStep)
	}
}

func (w *World) deepLife() {
	for i := range w.G.Stars {
		s := &w.G.Stars[i]
		switch w.Bio[i] {
		case BioNone:
			if s.Hab > 0 && w.R.Float64() < s.Hab*0.0015 {
				w.Bio[i] = BioSimple
				w.log("Life arises on the worlds of %s.", w.G.Describe(i))
			}
		case BioSimple:
			if w.R.Float64() < 0.012 {
				w.Bio[i] = BioComplex
				w.log("Complex life flourishes at %s.", w.star(i))
			}
		}
	}
}

// deepBurst: gamma-ray bursts sterilise a region. Sol is protected by fiat.
func (w *World) deepBurst() {
	if w.R.Float64() > 0.02 {
		return
	}
	origin := w.R.IntN(len(w.G.Stars))
	radius := 20.0 + w.R.Float64()*30
	killed := w.sterilise(origin, radius)
	if killed > 0 {
		w.log("A gamma-ray burst near %s sterilises %d living worlds within %.0f ly.", w.star(origin), killed, radius)
	}
}

func (w *World) sterilise(origin int, radius float64) int {
	killed := 0
	for _, s := range append(w.G.Near(origin, radius), origin) {
		if s != w.G.Sol && w.Bio[s] != BioNone {
			w.Bio[s] = BioNone
			killed++
		}
	}
	return killed
}

// deepStars kills stars scheduled to die in this step.
func (w *World) deepStars(from, to Year) {
	for i := range w.G.Stars {
		s := &w.G.Stars[i]
		if s.Dead() || Year(s.DiesAt) < from || Year(s.DiesAt) >= to {
			continue
		}
		if s.Massive() {
			killed := w.sterilise(i, 30)
			s.Kill()
			if killed > 0 {
				w.log("%s goes supernova. %d living worlds within 30 ly are sterilised.", w.star(i), killed)
			}
			continue
		}
		if w.Bio[i] != BioNone {
			w.log("%s swells and dies, and the life on its worlds with it.", w.star(i))
			w.Bio[i] = BioNone
		}
		s.Kill()
	}
}
