package history

import (
	"math"

	"worldgen/internal/species"
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

// law legacies are places where the state beneath shows through; the
// first is special (see mindDead).
var lawDescs = []string{
	"a region where minds do not work",
	"a standing signal that repeats one number",
	"a corridor along which ships arrive before they leave",
	"a volume where light arrives before it is sent",
	"a shore where the sky is thin, and things are seen in it that are not there",
	"a drift of dust that is in the same place wherever you go",
	"a star whose light is always a day old, from any distance",
}

// artifact nodes: which discoveries an elder artifact can stand for
var artifactNodes = []string{
	"fusion", "machine_minds", "self_replication", "antimatter", "relativistic", "relativistic_weapons",
	"terraforming", "memetics", "dyson", "stellar_engineering", "wormhole_physics", "near_light",
	"transcendence", "star_lifting", "matter_compilers", "substrate_minds", "world_engines", "stellar_weapons",
}

// miracleShare is the chance an elder artifact stands for a miracle rather
// than an art. This is the bargain of the elder legacy: what is found may be
// a better water purifier, or it may be the Door.
const miracleShare = 0.4

var structureNodes = []string{"stellar_engineering", "dyson", "wormhole_physics", "star_lifting"}

// elderMiracle picks the miracle an elder artifact stands for: one of the
// six, or now and then an object.
func (w *World) elderMiracle() string {
	if w.R.Float64() < objectShare {
		return objectKeys[w.R.IntN(len(objectKeys))]
	}
	var six []string
	for _, n := range tech.Miracles {
		if objectForms[n.Key] == nil {
			six = append(six, n.Key)
		}
	}
	return six[w.R.IntN(len(six))]
}

// runAges writes the myth. Each earlier turn of the cycle is an age: a
// surge of elder civilisations that fades as the galaxy's fertility fades,
// with a cosmic event to sweep up what is left. The current age's surge is
// the last one and belongs to the engine.
func (w *World) runAges() {
	surges := w.Cycle.Surges
	elders := 0
	for i, s := range surges[:len(surges)-1] {
		age := &AgeRecord{Index: i, Start: s, End: w.ageEnd(s)}
		w.Ages = append(w.Ages, age)
		w.eventAt(age.Start, KAgeDawn, nil, nil, -1, P{"age": i})
		nElders := 3 + w.R.IntN(4)
		for j := 0; j < nElders; j++ {
			e := &Elder{ID: elders, Age: i, Portrait: elderPortraits[w.R.IntN(len(elderPortraits))]}
			elders++
			// the earlier in the age, the likelier to rise: fertility is falling
			span := float64(age.End - age.Start)
			e.Rose = age.Start + Year(span*w.R.Float64()*w.R.Float64()*0.8)
			e.Fell = e.Rose + Year(w.R.Float64()*float64(age.End-e.Rose)*1.3)
			age.Elders = append(age.Elders, e)
			w.eventAt(e.Rose, KElderRose, nil, nil, -1, P{"elder": e.ID})
			if w.R.Float64() < 0.2 {
				w.eventAt(e.Rose+Year(float64(e.Fell-e.Rose)*0.6), KAgeKnower, nil, nil, -1, P{"elder": e.ID, "line": w.R.IntN(len(ageKnowers))})
			}
			nLeg := 2 + w.R.IntN(5)
			for k := 0; k < nLeg; k++ {
				at := e.Rose + Year(w.R.Float64()*float64(e.Fell-e.Rose))
				// what is left erodes with deep time
				if w.R.Float64() < math.Exp(-float64(-at)/5e9) {
					w.leaveLegacy(e, at)
				}
			}
			late := e.Fell > age.End
			w.eventAt(e.Fell, KElderFell, nil, nil, -1, P{"elder": e.ID, "late": late, "remembered": !late && w.R.Float64() < 0.35})
		}
		age.Ender = ageEnders[w.R.IntN(len(ageEnders))]
		w.eventAt(age.End, KAgeWaned, nil, nil, -1, P{"age": i})
	}
}

func (w *World) leaveLegacy(e *Elder, at Year) {
	// the elders lived where life is: half of what they left lies under
	// worlds that will be someone's cradle, the rest wherever
	s := w.R.IntN(len(w.G.Stars))
	if w.R.Float64() < 0.5 {
		var living []int
		for i, b := range w.Bio {
			if b == BioComplex && i != w.G.Sol {
				living = append(living, i)
			}
		}
		if len(living) > 0 {
			s = living[w.R.IntN(len(living))]
		}
	}
	if s == w.G.Sol {
		return
	}
	l := &Legacy{ID: len(w.Legacies), Age: e.Age, Elder: e, Maker: -1, Star: s, People: -1, Finder: -1, Source: -1, Plague: -1, Cond: Condition(w.R.IntN(2))}
	x := w.R.Float64()
	switch {
	case x < 0.1:
		w.leaveBounty(l, w.R.IntN(len(bounties)))
	case x < 0.5:
		l.Kind = Artifact
		if w.R.Float64() < miracleShare {
			l.Node = w.elderMiracle()
		} else {
			l.Node = artifactNodes[w.R.IntN(len(artifactNodes))]
		}
		l.Desc = artifactDescs[w.R.IntN(len(artifactDescs))]
	case x < 0.72:
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
	case x < 0.82:
		l.Kind = Threat
		w.Now = at
		if w.R.Float64() < 0.5 {
			// a dormant replicator people: machines that sleep in the rubble until something settles too close
			l.Node = "self_replication"
			l.Desc = "machines that sleep in the rubble of " + w.star(s)
			if p := w.replicatorAt(s, species.Machine, "left in the rubble of "+w.star(s)+" by "+e.Portrait, true); p != nil {
				l.People = p.ID
			}
		} else {
			l.Node = "memetics"
			if w.R.Float64() < 0.5 {
				l.Desc = "a transmitter at " + w.star(s) + ", silent"
			} else {
				l.Desc = "a transmitter at " + w.star(s) + " that has never stopped"
				l.State = Unleashed
			}
			if w.R.Float64() < w.Cfg.Tuning.Kinds.SeedShare {
				l.Payload = Seed
			}
		}
	case x < 0.94:
		l.Kind = Sleeper
		w.Now = at
		l.Desc = "something that withdrew into the dark near " + w.star(s) + " and went still"
		if p := w.sleeperAt(s, "something of "+e.Portrait+" that withdrew into the dark and went still"); p != nil {
			l.People = p.ID
		}
	default:
		l.Kind = Law
		l.Desc = lawDescs[w.R.IntN(len(lawDescs))]
		l.Node = "ftl"
	}
	w.Legacies = append(w.Legacies, l)
	e.Legacies = append(e.Legacies, l)
	w.eventAt(at, KElderLeft, nil, nil, l.Star, P{"elder": e.ID}).Legacy = l.ID
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
				w.event(KLifeArose, nil, nil, i, P{"class": string(s.Class)})
			}
		case BioSimple:
			if w.R.Float64() < 0.012 {
				w.Bio[i] = BioComplex
				w.event(KLifeComplex, nil, nil, i, P{})
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
		w.event(KBurst, nil, nil, origin, P{"killed": killed, "radius": radius})
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
				w.event(KSupernova, nil, nil, i, P{"killed": killed})
			}
			continue
		}
		if w.Bio[i] != BioNone {
			w.event(KStarSwelled, nil, nil, i, P{})
			w.Bio[i] = BioNone
		}
		s.Kill()
	}
}
