package history

import (
	"worldgen/internal/mind"
	"worldgen/internal/tech"
)

// The Find: a civilisation's reach touches a legacy of an earlier age. The
// species chooses what to attempt from its traits; the levels decide whether
// it succeeds; failure slides down the list to unleashing.

var finderNames = map[LegacyKind][]string{
	Artifact:  {"the Ones Who Left Things", "the Makers of Small Suns", "the Toolmakers", "the Careful Dead"},
	Structure: {"the Ones Who Moved the Star", "the Makers of the Hollow Sun", "the Builders", "the Ones Who Bent the Dark"},
	Threat:    {"the Ones Who Made the Hunger", "the Ones Who Lost Control", "the Warmakers"},
	Sleeper:   {"the Ones Who Dream", "the Sleepers", "the Ones Who Would Not Die"},
	Law:       {"the Lawgivers", "the Ones Who Changed the Rules"},
	Bounty:    {"the Providers", "the Ones Who Left the Table Laid", "the Gardeners"},
}

// find is the lottery: what turns up with nobody looking. Surveyors,
// colony ships and fleets find things by visiting; see explore.go.
func (w *World) find(c *Civ) {
	if !c.Active() || !c.Free() {
		return
	}
	var cands []*Legacy
	var own *Legacy
	for _, l := range w.Legacies {
		if (l.State != Buried && l.State != Sealed) || c.Found[l.ID] {
			continue
		}
		if l.Maker >= 0 && c.Known[l.Node] && l.ships() == 0 {
			continue // nothing to learn; a structure is taken over on settling; a field with ships is worth the ships
		}
		d := w.G.Dist(c.Home, l.Star)
		for _, s := range c.Systems {
			d = min(d, w.G.Dist(s, l.Star))
		}
		if (d == 0 && c.Era >= 1) || (d > 0 && d <= c.Reach) {
			cands = append(cands, l)
			if own == nil && w.kinship(c, l) == 2 {
				own = l
			}
		}
	}
	if len(cands) == 0 {
		return
	}
	// one's own lost works are looked for, and found quickly; the rest turn up by chance
	var l *Legacy
	how := ""
	switch {
	case own != nil && w.chance(0.05):
		l = own
		how = "own"
	case w.chance(findChance(c)):
		l = cands[w.R.IntN(len(cands))]
	default:
		return
	}
	w.discover(c, l, how)
}

// findChance is the rate per thousand years at which a people comes upon
// a remain in its reach by chance; a people fixed on the old things looks
// harder.
func findChance(c *Civ) float64 {
	if c.fixed(OldThings) {
		return 0.0002
	}
	return 0.0001
}

// discover is the Find itself: a people comes upon a remain, by a survey,
// by settling the star, by a fleet basing there, or by chance ("").
func (w *World) discover(c *Civ, l *Legacy, how string) {
	c.Found[l.ID] = true
	l.Finder = c.ID
	switch how {
	case "survey":
		c.Tally.FindSurvey++
	case "settle", "fleet", "ship":
		c.Tally.FindSettle++
	case "own":
		c.Tally.FindOwn++
	default:
		c.Tally.FindChance++
	}
	who := "The " + c.Name
	if how == "survey" {
		who = "Surveyors of the " + c.Name
	}
	where := "beneath their own cities on " + c.HomeName
	if l.Star != c.Home {
		where = "at " + w.star(l.Star)
	}
	if how == "settle" {
		where += ", under the feet of the first colonists"
	}
	kin := w.kinship(c, l)
	switch {
	case l.Elder != nil:
		if l.Elder.Name == "" {
			ns := finderNames[l.Kind]
			l.Elder.Name = ns[w.R.IntN(len(ns))]
		}
		w.log("%s find %s %s. It is older than their sun. They call its makers %s.", who, l.Desc, where, l.Elder.Name)
	case kin == 2:
		w.log("%s find %s %s. It is their own, from before the dark age. Something in them remembers it.", who, l.Describe(), where)
	case kin == 1:
		w.log("%s find %s %s. The hands that made it were like their hands.", who, l.Describe(), where)
	default:
		m := w.Civs[l.Maker]
		ago := float64(w.Now-m.Fell) / 1e6
		if (c.Met[m.ID] || m.Living()) && ago < 0.1 {
			w.log("%s find %s %s, not long after the %s left it.", who, l.Describe(), where, m.Name)
		} else if c.Met[m.ID] || m.Living() {
			w.log("%s find %s %s, %.1f million years after the %s left it.", who, l.Describe(), where, ago, m.Name)
		} else {
			l.Name = ruinNames[w.R.IntN(len(ruinNames))]
			w.log("%s find %s %s. They do not know who the %s were. They call them %s.", who, l.Describe(), where, m.Name, l.Name)
		}
	}
	if ff := w.factL(FFind, c, l); l.Elder != nil {
		ff.What = l.Elder.Name
	}
	w.readTestament(c, l)
	if !c.Active() {
		return
	}

	// what to attempt; a bounty was made to be used, and that is all it is for
	if l.Kind == Bounty {
		w.attemptWield(c, l)
		return
	}
	n := l.node()
	t := w.Cfg.Tuning
	a := mind.Find(mind.FindInput{
		Curious: c.Has("curious"), Expansionist: c.Has("expansionist"), Symbiotic: c.Has("symbiosis"), Cautious: c.Has("cautious"),
		Xenophobic: c.Has("xenophobic"), Contemplative: c.Has("contemplative"), Pragmatic: c.Has("pragmatic"), Conqueror: c.Has("conqueror"),
		Threat: l.Kind == Sleeper || l.Kind == Threat, Plain: n != nil && n.Miracle && l.Kind == Artifact,
		Own: w.kinship(c, l) == 2, Ruin: l.Maker >= 0 && l.Cond == Ruin, Law: l.Kind == Law,
		OldThings: c.fixed(OldThings), Field: l.Kind == Field, Known: l.Node == "" || c.Known[l.Node],
		Above: n != nil && (l.Kind == Artifact || l.Kind == Structure) && n.Era-c.Era >= t.Wisdom.AboveEras, Wis: c.Wis,
	}, t)
	w.explain(c, "weighing what to do with "+l.Describe(), a)
	pick := a.Pick(w.R.Float64())
	// whichever way the die falls, a wise people asks whether it can
	// before it tries
	if pick < 2 && a.Seal > 0 {
		margin := c.level("mil", "sur", "soc") - w.masterDiff(c, l)
		if pick == 1 {
			margin = c.Mil - w.wieldDiff(c, l)
		}
		if mind.Leap(mind.LeapInput{Margin: margin, Wis: c.Wis, Noise: w.R.NormFloat64()}, t) {
			c.Tally.Leaps++
			w.log("The %s, who are wise, look hard at it and count what it would take of them, and do not try.", c.Name)
			pick = 2
		}
	}
	switch pick {
	case 0:
		w.attemptMaster(c, l)
	case 1:
		w.attemptWield(c, l)
	default:
		w.attemptSeal(c, l)
	}
}

// masterDiff is what understanding a remain takes: read against the
// mean of the three levels.
func (w *World) masterDiff(c *Civ, l *Legacy) float64 {
	diff := 7.5 + c.traitDiff("find")
	if l.Kind == Sleeper {
		diff = 9
	}
	if l.Maker >= 0 {
		diff -= 1 // made to be understood by minds of this age
		diff -= 1.5 * float64(w.kinship(c, l))
		diff += l.condAdj()
	}
	if n := l.node(); n != nil {
		diff += 0.5 * float64(n.Era-c.Era)
	}
	return diff
}

// wieldDiff is what using a remain without understanding it takes: read
// against the military level.
func (w *World) wieldDiff(c *Civ, l *Legacy) float64 {
	diff := 4.5 + c.traitDiff("find")
	if l.Maker >= 0 {
		diff -= 1 + 1.5*float64(w.kinship(c, l))
		diff += l.condAdj()
	}
	if n := l.node(); n != nil && n.Miracle && l.Kind != Field {
		diff -= 1.5 // a miracle is made to be used; that is what makes it a miracle
	}
	return diff
}

func (w *World) attemptMaster(c *Civ, l *Legacy) {
	if l.Maker >= 0 && l.Node == "" {
		w.attemptWield(c, l) // nothing to read in it
		return
	}
	diff := w.masterDiff(c, l)
	if c.level("mil", "sur", "soc")+w.R.NormFloat64()*1.5 >= diff {
		l.State = Mastered
		c.Record = append(c.Record, "mastered a legacy of "+w.makerName(l))
		w.factL(FMastered, c, l)
		if l.Kind == Sleeper {
			c.Boons[BoonCommunion] = true
			w.log("The %s speak with what sleeps at %s, and it answers, and they are changed but not ended. They are more than they were.", c.Name, w.star(l.Star))
			return
		}
		gained := 0
		for _, k := range tech.Closure(l.Node) {
			if !c.Known[k] {
				gained++
			}
		}
		switch {
		case l.Kind == Threat:
			w.log("The %s take it apart, carefully, over centuries, and learn how it was made.", c.Name)
		case gained == 0:
			w.log("The %s understand it. There is nothing in it they did not already know.", c.Name)
		case w.kinship(c, l) == 2:
			w.log("The %s read it as their ancestors would have. The lost arts come back, and with them the rest. A renaissance.", c.Name)
		case l.Maker >= 0:
			w.log("The %s understand it, and through it what the %s knew. A renaissance built on another people's ruin.", c.Name, w.Civs[l.Maker].Name)
		default:
			w.log("The %s understand it. Understanding it, they understand everything that led to it.", c.Name)
		}
		w.finding = true
		for _, k := range tech.Closure(l.Node) {
			if !w.had(c, k) && c.Active() {
				w.learn(c, tech.Get(k), true)
			}
		}
		w.finding = false
		return
	}
	if l.Maker >= 0 && l.Cond == Ruin {
		w.log("The %s pick over it for centuries and learn nothing. There is not enough left.", c.Name)
		return
	}
	w.log("The %s try to understand it and cannot.", c.Name)
	w.attemptWield(c, l)
}

func (w *World) attemptWield(c *Civ, l *Legacy) {
	if l.Kind == Sleeper || l.Kind == Threat {
		w.unleash(c, l)
		return
	}
	if l.Maker >= 0 && l.Cond == Ruin {
		w.log("The %s try to make it work. Nothing in it will ever work again.", c.Name)
		return
	}
	if l.Kind == Bounty {
		if c.Sur+w.R.NormFloat64()*1.5 >= 2.5+c.traitDiff("find") {
			w.useBounty(c, l)
			c.Record = append(c.Record, "put a bounty of "+w.makerName(l)+" to use")
			w.log("The %s put it to use. It was made to be used, and it goes on doing what it did, for them now.", c.Name)
			return
		}
		w.log("The %s cannot make it do anything for them. It goes on doing what it did, for no one.", c.Name)
		return
	}
	diff := w.wieldDiff(c, l)
	if l.Kind == Field {
		// hulls: crewed, or not; nothing in a field gets loose
		if c.Mil+w.R.NormFloat64()*1.5 >= diff {
			l.State = Wielded
			c.Record = append(c.Record, "crewed the wrecks of "+w.makerName(l))
			w.salvage(c, l)
		} else {
			w.log("The %s try to crew the hulls, and cannot make them fly.", c.Name)
		}
		return
	}
	if c.Mil+w.R.NormFloat64()*1.5 >= diff {
		l.State = Wielded
		if l.Maker >= 0 && l.Cond == Wreck {
			l.Cond = Derelict // repaired, after a fashion
		}
		c.Record = append(c.Record, "wielded a legacy of "+w.makerName(l))
		if l.Kind == Structure && l.Maker >= 0 && (contains(c.Systems, l.Star) || (w.Owner[l.Star] < 0 && w.Held[l.Star] < 0 && w.canLive(c, l.Star))) {
			if !contains(c.Systems, l.Star) {
				w.settle(c, l.Star)
			}
			key := tech.Get(l.Node).Structure()
			c.Works = append(c.Works, Work{Key: key, Node: l.Node, Star: l.Star, Legacy: l.ID})
			c.Structures[key]++
			if c.Known[l.Node] {
				w.log("The %s take it over and put it back to work.", c.Name)
			} else {
				w.log("The %s move into it and keep it running. They could not build another.", c.Name)
			}
			w.recompute(c)
			return
		}
		c.Wielded = append(c.Wielded, l)
		l.Level = "all"
		if n := l.node(); n != nil && n.Miracle {
			l.Level = "miracle"
			if objectForms[n.Key] != nil {
				w.wieldObject(c, l) // the thing is the miracle
			} else {
				w.log("The %s learn to use it without understanding it. It is %s, and it is theirs for as long as it lasts.", c.Name, miracleNames[n.Key])
			}
			w.gain(c, n.Key, "wielded")
			if n.Filter != "" {
				w.face(c, n.Filter, 0)
			}
			return
		} else if n != nil {
			switch n.Domain {
			case tech.Weapons, tech.Industry:
				l.Level = "mil"
			case tech.Biology, tech.Energy:
				l.Level = "sur"
			case tech.Society, tech.Computation:
				l.Level = "soc"
			case tech.Propulsion:
				l.Level = "reach"
			}
		}
		if l.Kind == Law {
			l.Level = "reach"
		}
		w.log("The %s learn to use it without understanding it. If it breaks, it will stay broken.", c.Name)
		w.wieldRarity(c, l)
		w.recompute(c)
		if n := l.node(); n != nil && n.Filter != "" && l.Kind == Artifact {
			w.face(c, n.Filter, 0)
		}
		return
	}
	w.log("The %s try to use it. It does not do what they thought.", c.Name)
	w.unleash(c, l)
}

func (w *World) attemptSeal(c *Civ, l *Legacy) {
	if c.Soc+w.R.NormFloat64()*1.5 >= 3+c.traitDiff("find")+min(l.condAdj(), 0) {
		l.State = Sealed
		c.Record = append(c.Record, "sealed a legacy of "+w.makerName(l))
		w.factL(FSealed, c, l)
		w.log("The %s seal it, and post a watch, and the watch holds.", c.Name)
		return
	}
	w.log("The %s seal it. Someone opens it.", c.Name)
	w.unleash(c, l)
}

// unleash is the failure: the legacy acts on its own terms.
func (w *World) unleash(c *Civ, l *Legacy) {
	l.State = Unleashed
	c.Record = append(c.Record, "unleashed a legacy of "+w.makerName(l))
	w.factL(FUnleashed, c, l)
	switch l.Kind {
	case Sleeper:
		h := w.Horrors[l.Horror]
		w.log("It wakes.")
		w.wakeElder(h)
	case Threat:
		h := w.Horrors[l.Horror]
		h.Dormant = false
		if h.Kind == Beacon {
			w.log("The transmitter at %s speaks again. It is called %s.", w.star(l.Star), h.Name)
		} else {
			w.log("The machines at %s wake, and begin to eat. They are called %s.", w.star(l.Star), h.Name)
		}
	case Structure:
		if l.Maker >= 0 {
			w.blast(l.Star, 3, "the failure of "+l.Desc, "Something at %s that its makers left running comes apart.", 1)
		} else {
			w.blast(l.Star, 12, "the failure of "+l.Desc, "Something at %s that held for a billion years lets go.", 2)
		}
	case Law:
		w.tear(0.6)
		w.log("The %s break something at %s that was not a thing but a rule. The rule reasserts itself.", c.Name, w.star(l.Star))
		w.blast(l.Star, 5, "a broken law", "Around %s, for a moment, physics is negotiable.", 3)
	case Artifact:
		n := l.node()
		if n.Miracle && objectForms[n.Key] != nil {
			w.unleashObject(c, l) // the thing is the miracle, and it gets out
			return
		}
		if n.Miracle {
			// the miracle's own danger, at its worst
			w.log("It works, once, in a way nobody chose.")
			w.face(c, n.Filter, 2)
			return
		}
		switch n.Domain {
		case tech.Industry, tech.Weapons:
			h := w.spawnHorror(Replicators, l.Star, -1)
			h.Legacy = l.ID
			w.log("Whatever it was, it makes more of itself. It is called %s.", h.Name)
		case tech.Computation:
			h := w.spawnHorror(RogueMind, l.Star, -1)
			h.Legacy = l.ID
			w.log("It was a mind, and it is awake, and it is called %s.", h.Name)
		case tech.Exotic, tech.Propulsion:
			h := w.spawnHorror(Beacon, l.Star, -1)
			h.Legacy = l.ID
			w.log("It was a door, or a voice. It is called %s now.", h.Name)
		default:
			w.log("It does what it was made to do, to the %s.", c.Name)
			if n.Filter != "" {
				w.face(c, n.Filter, 3)
			} else {
				w.face(c, "plague", 3)
			}
		}
	}
}

// dropWielded: artifacts are lost when their holder falls. Some return to
// the substrate for a successor to find; some are simply broken.
func (w *World) dropWielded(c *Civ, p float64) {
	keep := c.Wielded[:0]
	for _, l := range c.Wielded {
		if w.R.Float64() > p {
			keep = append(keep, l)
			continue
		}
		if w.R.Float64() < 0.5 {
			l.State = Buried
			l.Star = c.Home
			w.log("What the %s wielded of %s lies where they left it, on %s.", c.Name, w.makerName(l), c.HomeName)
		} else {
			l.State = Lost
			w.log("What the %s wielded of %s is broken, and nobody knows how to mend it.", c.Name, w.makerName(l))
		}
		if l.Source >= 0 {
			s := w.Sources[l.Source]
			s.Holder, s.Carried, s.Star = -1, -1, l.Star
			if l.State == Lost {
				s.Star = -1
			}
		}
	}
	c.Wielded = keep
	w.recompute(c)
}
