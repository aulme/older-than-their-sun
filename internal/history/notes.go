package history

import (
	"worldgen/internal/galaxy"
	"worldgen/internal/species"
)

// Notes: the silent kinds of the chronicle. A durable change the prose
// never remarked (a world passing to a master, a remain crumbling a step
// unseen, a people's stage moving with its reach) is still a happening,
// and the chronicle is complete about the durable: the fold test
// (internal/writer) rebuilds every durable field of the state from the
// chronicle alone, and a change without an event fails it. A note is
// written by the setter that makes the change, so the two cannot
// drift. Notes have no line in the view and no weight; they take their
// ids after every told event of the run, so the ids the tellings hash
// stay where they were.

// note records a silent event at the present tick.
func (w *World) note(k Kind, c, e *Civ, star int, p P) *Event {
	ev := &Event{Year: w.Now, Kind: k, Subject: -1, Object: -1, Star: star, Legacy: -1, Plague: -1, P: p}
	if c != nil {
		ev.Subject = c.ID
	}
	if e != nil {
		ev.Object = e.ID
	}
	if ev.P == nil {
		ev.P = P{}
	}
	w.notes = append(w.notes, ev)
	return ev
}

// placeNotes gives the notes their ids, after every told event, and
// puts them in the chronicle; the caller sorts it by year.
func (w *World) placeNotes() {
	for _, n := range w.notes {
		n.ID = len(w.Events)
		w.Events = append(w.Events, n)
		w.Chronicle = append(w.Chronicle, n)
	}
	w.notes = nil
}

// setOwner is a star changing hands: lost by its holder, held by the
// next. A star nobody held to nobody holding it is nothing.
func (w *World) setOwner(s, id int) {
	was := w.Owner[s]
	if was == id {
		return
	}
	w.Owner[s] = id
	if was >= 0 {
		w.note(KWorldLost, w.Civs[was], nil, s, nil)
	}
	if id >= 0 {
		w.note(KWorldHeld, w.Civs[id], nil, s, nil)
	}
}

// setBio is life on a star's worlds changing.
func (w *World) setBio(s int, b BioState) {
	if w.Bio[s] == b {
		return
	}
	w.Bio[s] = b
	w.note(KBio, nil, nil, s, P{"bio": [...]string{"none", "simple", "complex"}[b]})
}

// killStar turns a star into its remnant.
func (w *World) killStar(s *galaxy.Star) {
	s.Kill()
	w.note(KStarClass, nil, nil, s.ID, P{"class": string(s.Class)})
}

// setStage is a people's stage moving.
func (w *World) setStage(c *Civ, st Stage) {
	if c.Stage == st {
		return
	}
	c.Stage = st
	w.note(KStage, c, nil, -1, P{"stage": [...]string{"emergent", "interstellar", "zenith", "remnant", "dead"}[st]})
}

// setFate is a people's fate and cause set, at its fall or its end.
func (w *World) setFate(c *Civ, f Fate, cause string) {
	c.Fate, c.Cause = f, cause
	w.note(KFate, c, nil, -1, P{"fate": f.String(), "cause": cause})
}

// setHome is a people's seat moving.
func (w *World) setHome(c *Civ, s int) {
	if c.Home == s {
		return
	}
	c.Home = s
	w.note(KSeat, c, nil, s, nil)
}

// know is a node coming into a people's holding by any road.
func (w *World) know(c *Civ, k string) {
	if c.Known[k] {
		return
	}
	c.Known[k] = true
	w.note(KNodeHeld, c, nil, -1, P{"node": k})
}

// forgetNode is a node lost.
func (w *World) forgetNode(c *Civ, k string) {
	if !c.Known[k] {
		return
	}
	delete(c.Known, k)
	w.note(KNodeLost, c, nil, -1, P{"node": k})
}

// scar and boon are a mark taken.
func (w *World) scar(c *Civ, key string) {
	if c.Scars[key] {
		return
	}
	c.Scars[key] = true
	w.note(KScar, c, nil, -1, P{"scar": key})
}

func (w *World) boon(c *Civ, key string) {
	if c.Boons[key] {
		return
	}
	c.Boons[key] = true
	w.note(KBoon, c, nil, -1, P{"boon": key})
}

// holdMiracle is a miracle coming into a people's holding.
func (w *World) holdMiracle(c *Civ, key, how string) {
	c.Miracles[key] = how
	w.note(KMiracleHeld, c, nil, -1, P{"miracle": key, "route": how})
}

// dropMiracle is a miracle lost.
func (w *World) dropMiracle(c *Civ, key string) {
	if _, had := c.Miracles[key]; !had {
		return
	}
	delete(c.Miracles, key)
	w.note(KMiracleLost, c, nil, -1, P{"miracle": key})
}

// setMaster is a people's master changing: slave, vassal, or free.
func (w *World) setMaster(c *Civ, id int, vassal bool) {
	if c.Master == id && c.Vassal == vassal {
		return
	}
	c.Master, c.Vassal = id, vassal
	w.note(KMaster, c, nil, -1, P{"master": id, "vassal": vassal})
}

// setAsleep is the long sleep begun or ended.
func (w *World) setAsleep(c *Civ, asleep bool) {
	if c.Asleep == asleep {
		return
	}
	c.Asleep = asleep
	w.note(KSleep, c, nil, -1, P{"asleep": asleep})
}

// setLine is an heir's line set at its birth.
func (w *World) setLine(c *Civ, line []int) {
	c.Line = line
	w.note(KLine, c, nil, -1, P{"line": line})
}

// setNamed is a people having a word for the state beneath.
func (w *World) setNamed(c *Civ, named bool) {
	if c.Named == named {
		return
	}
	c.Named = named
	w.note(KNamed, c, nil, -1, P{"named": named})
}

// addLegacy is a remain left, with its state and condition as left.
func (w *World) addLegacy(l *Legacy) {
	l.ID = len(w.Legacies)
	w.Legacies = append(w.Legacies, l)
	n := w.note(KRemainLeft, nil, nil, l.Star, P{"kind": l.Kind.String(), "state": l.State.String(), "cond": l.Cond.String(), "finder": l.Finder, "maker": l.Maker})
	n.Legacy = l.ID
}

// left says whether a remain is in the record yet; a change to one that
// is not is part of its leaving, and remain_left carries it.
func (w *World) left(l *Legacy) bool {
	return l.ID < len(w.Legacies) && w.Legacies[l.ID] == l
}

// moveRemain is a remain put down at another star: a wielded thing
// buried where its holder sits, a work going back to the substrate.
func (w *World) moveRemain(l *Legacy, star int) {
	if l.Star == star {
		return
	}
	l.Star = star
	if w.left(l) {
		w.note(KRemainMoved, nil, nil, star, nil).Legacy = l.ID
	}
}

// setState is a remain's state changing.
func (w *World) setState(l *Legacy, st LegacyState) {
	if l.State == st {
		return
	}
	l.State = st
	if w.left(l) {
		w.note(KRemainState, nil, nil, l.Star, P{"state": st.String()}).Legacy = l.ID
	}
}

// setCond is a remain's condition changing: decay, or a repair.
func (w *World) setCond(l *Legacy, cond Condition) {
	if l.Cond == cond {
		return
	}
	l.Cond = cond
	if w.left(l) {
		w.note(KRemainCond, nil, nil, l.Star, P{"cond": cond.String()}).Legacy = l.ID
	}
}

// setFinder is the people that last acted on a remain.
func (w *World) setFinder(l *Legacy, id int) {
	if l.Finder == id {
		return
	}
	l.Finder = id
	if w.left(l) {
		w.note(KRemainFinder, nil, nil, l.Star, P{"finder": id}).Legacy = l.ID
	}
}

// listened is a transmitter taking a people.
func (w *World) listened(l *Legacy, c *Civ) {
	l.Listeners++
	w.note(KListened, c, nil, l.Star, nil).Legacy = l.ID
}

// addWar is a war declared or inherited: its sides and its cause.
func (w *World) addWar(wr *War) {
	w.Wars = append(w.Wars, wr)
	w.note(KWarOpened, w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]], -1, P{"war": wr.ID, "cause": wr.Cause})
}

// warOver is a war ended, with its result.
func (w *World) warOver(wr *War, result string) {
	w.note(KWarOver, w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]], -1, P{"war": wr.ID, "result": result})
}

// infected and cleared are a plague's hold on a people taken and lost.
func (w *World) infected(c *Civ, p *Plague) {
	w.note(KInfected, c, nil, -1, nil).Plague = p.ID
}

func (w *World) cleared(c *Civ, pid int) {
	if _, had := c.Infections[pid]; !had {
		return
	}
	delete(c.Infections, pid)
	w.note(KCleared, c, nil, -1, nil).Plague = pid
}

// gainedPower is a power added to a blood.
func (w *World) gainedPower(c *Civ, key string) {
	w.note(KPowerHeld, c, nil, -1, P{"power": key, "species": c.Species.ID})
}

// setSpecies is a people's blood changing: a branch of its own, on a
// drift or a brood scar.
func (w *World) setSpecies(c *Civ, sp *species.Species) {
	c.Species = sp
	w.note(KBlood, c, nil, -1, P{"species": sp.ID, "powers": append([]string{}, sp.Powers...)})
}

// born is a people's birth: its blood, its cradle and its master; and
// the powers its blood holds, for a blood new to the record.
func (w *World) born(c *Civ) {
	w.note(KCivBorn, c, nil, c.Cradle, P{"species": c.Species.ID, "master": c.Master, "powers": append([]string{}, c.Species.Powers...)})
}
