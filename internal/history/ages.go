package history

import (
	"math"

	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// The age generator. Earlier ages are myth: one tick per rise and fall,
// elder civilisations with a portrait instead of traits, and legacies left
// on the substrate for the current age to find.

// The portraits of the elders, their ages' enders, what they learned and
// what they left are data/portraits.json; the record keeps the key drawn.
// The first law is special (see mindDead): its variant in the table.

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
			e := &Elder{ID: elders, Age: i, Portrait: tables.portraits.Elders[w.R.IntN(len(tables.portraits.Elders))].Key}
			elders++
			// the earlier in the age, the likelier to rise: fertility is falling
			span := float64(age.End - age.Start)
			e.Rose = age.Start + Year(span*w.R.Float64()*w.R.Float64()*0.8)
			e.Fell = e.Rose + Year(w.R.Float64()*float64(age.End-e.Rose)*1.3)
			age.Elders = append(age.Elders, e)
			w.eventAt(e.Rose, KElderRose, nil, nil, -1, P{"elder": e.ID})
			if w.R.Float64() < 0.2 {
				w.eventAt(e.Rose+Year(float64(e.Fell-e.Rose)*0.6), KAgeKnower, nil, nil, -1, P{"elder": e.ID, "line": tables.portraits.Knowers[w.R.IntN(len(tables.portraits.Knowers))].Key})
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
		age.Ender = tables.portraits.Enders[w.R.IntN(len(tables.portraits.Enders))].Key
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
		w.leaveBounty(l, w.R.IntN(len(tables.portraits.Bounties)))
	case x < 0.5:
		l.Kind = Artifact
		if w.R.Float64() < miracleShare {
			l.Node = w.elderMiracle()
		} else {
			l.Node = artifactNodes[w.R.IntN(len(artifactNodes))]
		}
		l.Portrait = tables.portraits.Artifacts[w.R.IntN(len(tables.portraits.Artifacts))].Key
	case x < 0.72:
		l.Kind = Structure
		l.Node = structureNodes[w.R.IntN(len(structureNodes))]
		l.Portrait = tables.portraits.Structures[w.R.IntN(len(tables.portraits.Structures))].Key
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
			l.Portrait = "rubble"
			if p := w.replicatorAt(s, species.Machine, Origin{Key: "rubble", By: -1, From: -1, Legacy: l.ID, Plague: -1}, true); p != nil {
				l.People = p.ID
			}
		} else {
			l.Node = "memetics"
			if w.R.Float64() < 0.5 {
				l.Portrait = "transmitter_silent"
			} else {
				l.Portrait = "transmitter_running"
				l.State = Unleashed
			}
			if w.R.Float64() < w.Cfg.Tuning.Kinds.SeedShare {
				l.Payload = Seed
			}
		}
	case x < 0.94:
		l.Kind = Sleeper
		w.Now = at
		l.Portrait = "withdrawn"
		if p := w.sleeperAt(s, Origin{Key: "withdrawn", By: -1, From: -1, Legacy: l.ID, Plague: -1}); p != nil {
			l.People = p.ID
		}
	default:
		l.Kind = Law
		l.Portrait = tables.portraits.Laws[w.R.IntN(len(tables.portraits.Laws))].Key
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
