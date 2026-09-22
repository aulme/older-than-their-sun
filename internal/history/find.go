package history

import (
	"worldgen/internal/mind"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// The Find: a civilisation's reach touches a legacy of an earlier age. The
// species chooses what to attempt from its traits; the levels decide whether
// it succeeds; failure slides down the list to unleashing.

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
		if l.People == c.ID {
			continue // the remain is what it is: a sleeper does not find itself
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
	// what the finder can say of the makers: an elder's, its own line's,
	// its own blood's, a people it knows of, or nobody it can place (whose
	// name for them is then a row of the names pass)
	p := P{"how": how, "cond": int(l.Cond), "ships": l.ships(), "adrift": l.Adrift, "kin": w.kinship(c, l), "elder": -1, "known": true, "ago": 0.0, "own": l.Star == c.Home}
	switch {
	case l.Elder != nil:
		p["elder"] = l.Elder.ID
	case p["kin"] == 0:
		m := w.Civs[l.Maker]
		p["ago"] = float64(w.Now-m.Fell) / 1e6
		p["known"] = c.Met[m.ID] || m.Living()
	}
	w.factL(FFind, c, l).with(p)
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
	if !c.Species.Profile().Can(species.Researches) {
		a.Master = 0 // no tree to master it into: it wields what it finds, or seals it
	}
	w.explain(c, "weighing what to do with "+w.Describe(l), a)
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
			w.event(KFindLeap, c, nil, l.Star, P{}).Legacy = l.ID
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
		c.Record = append(c.Record, Record{Kind: "mastered", Legacy: l.ID, Known: w.knowsMaker(c, l)})
		f := w.toldL(FMastered, c, l)
		if l.Kind == Sleeper {
			c.Boons[BoonCommunion] = true
			f.P["way"] = "sleeper"
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
			f.P["way"] = "threat"
		case gained == 0:
			f.P["way"] = "nothing"
		case w.kinship(c, l) == 2:
			f.P["way"] = "own"
		case l.Maker >= 0:
			f.P["way"] = "other"
		default:
			f.P["way"] = "all"
		}
		w.finding = true
		for _, k := range tech.Closure(l.Node) {
			if !w.had(c, k) && c.Active() {
				w.learn(c, tech.Get(k), true)
			}
		}
		w.finding = false
		if gained > 0 {
			w.renew(c, 0.2) // an art mastered from a find is something new
		}
		return
	}
	if l.Maker >= 0 && l.Cond == Ruin {
		w.event(KMasterFailed, c, nil, l.Star, P{"way": "ruin"}).Legacy = l.ID
		return
	}
	w.event(KMasterFailed, c, nil, l.Star, P{"way": "cannot"}).Legacy = l.ID
	w.attemptWield(c, l)
}

func (w *World) attemptWield(c *Civ, l *Legacy) {
	if l.Kind == Sleeper || l.Kind == Threat {
		w.unleash(c, l)
		return
	}
	if l.Maker >= 0 && l.Cond == Ruin {
		w.event(KWieldFailed, c, nil, l.Star, P{"way": "ruin"}).Legacy = l.ID
		return
	}
	if l.Kind == Bounty {
		if c.Sur+w.R.NormFloat64()*1.5 >= 2.5+c.traitDiff("find") {
			w.useBounty(c, l)
			c.Record = append(c.Record, Record{Kind: "bounty", Legacy: l.ID, Known: w.knowsMaker(c, l)})
			w.event(KWielded, c, nil, l.Star, P{"way": "bounty"}).Legacy = l.ID
			return
		}
		w.event(KWieldFailed, c, nil, l.Star, P{"way": "bounty"}).Legacy = l.ID
		return
	}
	diff := w.wieldDiff(c, l)
	if l.Kind == Field {
		// hulls: crewed, or not; nothing in a field gets loose
		if c.Mil+w.R.NormFloat64()*1.5 >= diff {
			l.State = Wielded
			c.Record = append(c.Record, Record{Kind: "crewed", Legacy: l.ID, Known: w.knowsMaker(c, l)})
			w.salvage(c, l)
		} else {
			w.event(KWieldFailed, c, nil, l.Star, P{"way": "hulls"}).Legacy = l.ID
		}
		return
	}
	if c.Mil+w.R.NormFloat64()*1.5 >= diff {
		l.State = Wielded
		if l.Maker >= 0 && l.Cond == Wreck {
			l.Cond = Derelict // repaired, after a fashion
		}
		c.Record = append(c.Record, Record{Kind: "wielded", Legacy: l.ID, Known: w.knowsMaker(c, l)})
		if l.Kind == Structure && l.Maker >= 0 && (contains(c.Systems, l.Star) || (w.Owner[l.Star] < 0 && w.canLive(c, l.Star))) {
			if !contains(c.Systems, l.Star) {
				w.settle(c, l.Star)
			}
			key := tech.Get(l.Node).Structure()
			c.Works = append(c.Works, Work{Key: key, Node: l.Node, Star: l.Star, Legacy: l.ID})
			c.Structures[key]++
			if c.Known[l.Node] {
				w.event(KWielded, c, nil, l.Star, P{"way": "takeover"}).Legacy = l.ID
			} else {
				w.event(KWielded, c, nil, l.Star, P{"way": "moved_in"}).Legacy = l.ID
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
				w.event(KWielded, c, nil, l.Star, P{"way": "miracle", "miracle": n.Key}).Legacy = l.ID
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
		w.event(KWielded, c, nil, l.Star, P{"way": "blind"}).Legacy = l.ID
		w.wieldRarity(c, l)
		w.recompute(c)
		if n := l.node(); n != nil && n.Filter != "" && l.Kind == Artifact {
			w.face(c, n.Filter, 0)
		}
		return
	}
	w.event(KWieldFailed, c, nil, l.Star, P{"way": "wrong"}).Legacy = l.ID
	w.unleash(c, l)
}

func (w *World) attemptSeal(c *Civ, l *Legacy) {
	if c.Soc+w.R.NormFloat64()*1.5 >= 3+c.traitDiff("find")+min(l.condAdj(), 0) {
		l.State = Sealed
		c.Record = append(c.Record, Record{Kind: "sealed", Legacy: l.ID, Known: w.knowsMaker(c, l)})
		w.toldL(FSealed, c, l)
		return
	}
	w.event(KSealFailed, c, nil, l.Star, P{}).Legacy = l.ID
	w.unleash(c, l)
}

// unleash is the failure: the legacy acts on its own terms.
func (w *World) unleash(c *Civ, l *Legacy) {
	l.State = Unleashed
	c.Record = append(c.Record, Record{Kind: "unleashed", Legacy: l.ID, Known: w.knowsMaker(c, l)})
	f := w.toldL(FUnleashed, c, l).with(P{"way": ""})
	switch l.Kind {
	case Sleeper, Threat:
		if l.People < 0 {
			f.P["way"] = "transmitter"
			break
		}
		p := w.Civs[l.People]
		if !p.Active() {
			f.P["way"] = "gone"
			break
		}
		f.P["way"] = "wakes"
		w.rouse(p, c)
	case Structure:
		if l.Maker >= 0 {
			w.blast(l.Star, 3, "failure", nil, 1, l.ID)
		} else {
			w.blast(l.Star, 12, "failure_old", nil, 2, l.ID)
		}
	case Law:
		w.tear(0.6)
		f.P["way"] = "law"
		w.blast(l.Star, 5, "law", nil, 3, l.ID)
	case Artifact:
		n := l.node()
		if n.Miracle && objectForms[n.Key] != nil {
			w.unleashObject(c, l) // the thing is the miracle, and it gets out
			return
		}
		if n.Miracle {
			// the miracle's own danger, at its worst
			f.P["way"] = "miracle"
			w.face(c, n.Filter, 2)
			return
		}
		switch n.Domain {
		case tech.Industry, tech.Weapons:
			f.P["way"] = "replicator"
			if nc := w.replicatorAt(l.Star, species.Machine, Origin{Key: "breakout", By: -1, From: -1, Legacy: l.ID, Plague: -1}, false); nc != nil {
				w.event(KNamedItself, nc, nil, l.Star, P{"way": "called", "traits": nc.Species.TraitKeys()})
			}
		case tech.Computation:
			f.P["way"] = "mind"
			sp := species.GenerateWith(w.R, w.G.Stars[l.Star].Mult, w.G.Sys[l.Star].Arch, species.Machine, 0)
			if nc := w.ariseAt(l.Star, sp, Origin{Key: "woke_in", By: -1, From: -1, Legacy: l.ID, Plague: -1}); nc != nil {
				nc.Origin = sp.Made
				w.event(KNamedItself, nc, nil, l.Star, P{"way": "mind", "traits": sp.TraitKeys()})
			}
		case tech.Exotic, tech.Propulsion:
			w.makeTransmitter(l.Star, -1, true)
			f.P["way"] = "voice"
		default:
			if n.Filter != "" {
				f.P["way"] = "filter"
				w.face(c, n.Filter, 3)
			} else {
				w.wakeRelic(c, l)
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
			w.event(KWieldedDropped, c, nil, c.Home, P{"way": "buried", "known": w.knowsMaker(c, l)}).Legacy = l.ID
		} else {
			l.State = Lost
			w.event(KWieldedDropped, c, nil, -1, P{"way": "broken", "known": w.knowsMaker(c, l)}).Legacy = l.ID
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
