package history

import (
	"worldgen/internal/species"
)

// The anti-memetic: known only by the absence of information about it.
// Nothing conscious can hold it in mind, live or in records, so no
// conscious people meets it, sights its fleets, reads its worlds as
// anything but empty or receives its messages, and its deeds against a
// conscious people are facts with no doer on the victim's side. It
// perceives everyone. The unconscious perceive it as anyone, and so does
// a conscious people that holds Antimemetic Resilience, for as long as it
// holds it: a dark age that forgets the node forgets the people too. The
// chronicle names it, since the chronicle is not a mind. What a conscious
// people can do about it is in gap.go: fight it by the shape of the hole.

// resilience is the node that lets a conscious people hold the
// anti-memetic in mind.
const resilience = "antimemetic_resilience"

// perceives says whether a people can hold another in mind.
func (w *World) perceives(c, e *Civ) bool {
	if c == e || !e.Species.Is(species.Antimemetic) {
		return true
	}
	return c.Species.Profile().NoOne || c.Known[resilience]
}

// seen is a party of a fact as a people can hold it: the people, or -1
// for one it cannot perceive.
func (w *World) seen(c *Civ, id int) int {
	if id < 0 || w.perceives(c, w.Civs[id]) {
		return id
	}
	return -1
}

// veiled says whether a fact has a party this people cannot perceive.
func (w *World) veiled(c *Civ, f *Event) bool {
	return w.seen(c, f.Subject) < 0 || w.seen(c, f.Object) < 0
}

// keeps says whether a people can hold a tale at all: one with a party it
// cannot perceive only when the people is the other party, since it can
// perceive its own losses and nothing about the thing.
func (w *World) keeps(c *Civ, f *Event) bool {
	if !w.veiled(c, f) {
		return true
	}
	return f.Subject == c.ID || f.Object == c.ID
}

// notice is a one-sided meeting: the side that perceives learns the
// other is there, and the other learns nothing. What follows is the
// seer's council's matter alone.
func (w *World) notice(seer, unseen *Civ) {
	if seer.Met[unseen.ID] {
		return
	}
	addMet(seer, unseen.ID)
	seer.Reached[unseen.ID] = true
	seer.Tally.MetTouch++
	w.noticed(seer, unseen, -1).with(P{"way": "unseen", "hidden": seer.Species.Is(species.Antimemetic)})
	w.observe(seer, unseen, unseen.Home, 0.5)
	w.tryFathom(seer, unseen, "meeting", 0)
}

// unveil is a conscious people coming to hold the node: every
// anti-memetic people that has found it is found in turn, and the wars
// it was hunting by the shape of the hole are wars on a people now.
func (w *World) unveil(c *Civ) {
	if c.Species.Profile().NoOne {
		return
	}
	for _, e := range w.Civs {
		if !e.Active() || e == c || !e.Species.Is(species.Antimemetic) || c.Met[e.ID] {
			continue
		}
		if e.Met[c.ID] || w.touch(c, e) || w.hear(c, e) {
			addMet(c, e.ID)
			c.Reached[e.ID] = e.Reached[c.ID]
			w.meeting(c, e, -1, "touch").with(P{"way": "unveiled"})
			w.observe(c, e, e.Home, 0.5)
			w.tryFathom(c, e, "meeting", 0)
		}
		if wr := w.warBetween(c.ID, e.ID); wr != nil && wr.Gap != nil {
			wr.Gap = nil
			w.event(KQuarry, c, e, -1, P{})
		}
	}
}

// veil is a conscious people losing the node: the anti-memetic peoples it
// held in mind are gone from it, its tales of them go doerless again and
// its wars on them become hunts.
func (w *World) veil(c *Civ) {
	if c.Species.Profile().NoOne {
		return
	}
	for _, e := range w.Civs {
		if e == c || !e.Species.Is(species.Antimemetic) || !c.Met[e.ID] {
			continue
		}
		if c.Trade[e.ID] {
			w.cutTrade(c, e, because("forgetting"))
			w.cutTrade(e, c, because("forgetting"))
			endTrade(c, e)
		}
		dropMet(c, e.ID)
		delete(c.Reached, e.ID)
		delete(c.Fathomed, e.ID)
		delete(c.Intel, e.ID)
		delete(c.Watched, e.ID)
		delete(c.Grudge, e.ID)
		if e.Active() {
			w.event(KVeiled, c, e, -1, P{})
		}
		if wr := w.warBetween(c.ID, e.ID); wr != nil && wr.Gap == nil {
			w.huntOn(c, wr)
		}
	}
}
