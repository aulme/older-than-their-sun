package history

import (
	"testing"

	"worldgen/internal/flow"
	"worldgen/internal/species"
)

// TestStandingTribute: a vassal owes a standing tribute from the bond; paid, it
// is income to the patron; paid short, it angers the patron; reviewed,
// a rate paid short rises; and the leash lets a patron make war on its
// client only for a tribute unpaid.
func TestStandingTribute(t *testing.T) {
	w, m, s, _ := twoPeoples(t, 71)
	w.vassal(m, s, "surrender")
	tn := &w.Cfg.Tuning.Client
	if s.Rate < tn.RateMin || s.Rate > tn.RateMax || !w.tributary(s) {
		t.Fatalf("the bond's rate %.3f, tributary %v", s.Rate, w.tributary(s))
	}
	s.Income = flow.Income{10, 10, 10}
	uses := w.tributeUses(s)
	if len(uses) != 1 || uses[0].Need[flow.O] != 10*s.Rate {
		t.Fatalf("the tribute's use: %+v", uses)
	}
	m.Paid = flow.Income{}
	s.Shed = map[string]bool{}
	w.payTribute(s)
	if m.Paid[flow.O] != 10*s.Rate || s.Tally.TributeTicks != 1 {
		t.Fatalf("paid %.2f to the patron, ticks owed %d", m.Paid[flow.O], s.Tally.TributeTicks)
	}
	g := m.Grudge[s.ID]
	s.Shed[tributeKey] = true
	w.payTribute(s)
	if s.Tally.TributeShort != 1 || m.Grudge[s.ID] <= g {
		t.Fatalf("paid short: ticks short %d, the patron's grudge %.2f from %.2f", s.Tally.TributeShort, m.Grudge[s.ID], g)
	}
	was := s.Rate
	s.RateReviewed = w.Now - Year(tn.ReviewEvery)
	w.reviewTribute(s)
	if s.Rate <= was || s.PaidShort != 0 {
		t.Errorf("reviewed after paying short: rate %.3f from %.3f", s.Rate, was)
	}
	if w.mayWar(m, s, "border") || !w.mayWar(m, s, "unpaid") {
		t.Error("the leash: a patron may make war on its client for its tribute and nothing else")
	}
}

// TestLeash: a vassal may make war on a free people and on another
// master's vassal, and not on its master, its master's ally or its
// master's other vassal; a slave on nobody.
func TestLeash(t *testing.T) {
	w := newTestWorld(t, 72, 40)
	people := func(star int) *Civ { return spawnAt(w, star, species.Fixed("defensive")) }
	m, v, o, f, m2, v2 := people(1), people(2), people(3), people(4), people(5), people(6)
	w.vassal(m, v, "offer")
	w.vassal(m, o, "offer")
	w.vassal(m2, v2, "offer")
	switch {
	case !w.mayWar(v, f, ""):
		t.Error("a vassal may not strike a free people")
	case !w.mayWar(v, v2, ""):
		t.Error("a vassal may not strike another master's vassal")
	case w.mayWar(v, m, ""):
		t.Error("a vassal may strike its master")
	case w.mayWar(v, o, ""):
		t.Error("a vassal may strike its master's other vassal")
	}
	w.formPact(m, f, Defensive, -1, -1)
	if w.mayWar(v, f, "") {
		t.Error("a vassal may strike its master's ally")
	}
	w.enslave(m2, v2)
	if w.mayWar(v, v2, "") {
		t.Error("a slave may be struck in its own name")
	}
}

// TestPunishOnce: a patron angered by its client's tribute paid short
// makes war on it at most once a truce; the war spends the anger, and a
// war on another's client is for worlds, not for the client itself
// (stage 3's first batch: a patron punishing the same client every two
// thousand years; two powers taking one client off each other a hundred
// and thirty-five times).
func TestPunishOnce(t *testing.T) {
	w, m, s, _ := twoPeoples(t, 74)
	w.vassal(m, s, "offer")
	m.Grudge[s.ID] = w.Cfg.Tuning.Client.Punish + 1
	m.Truce[s.ID] = w.Now + 5000
	if w.punish(m) {
		t.Fatal("punished within the truce")
	}
	wr := w.declare(m, s, because("unpaid"))
	if wr == nil {
		t.Fatal("no war for the tribute")
	}
	w.endWar(wr, "peace")
	if m.Grudge[s.ID] > 0 {
		t.Errorf("the punishment left the anger: %.2f", m.Grudge[s.ID])
	}
	x := starfarer2(w, 7)
	if wr := w.declare(x, s, because("conquest")); wr == nil || wr.Aim != "world" {
		t.Errorf("a conqueror's war on another's client is for %v, want a world", wr)
	}
}

// starfarer2 is a third people in a test world.
func starfarer2(w *World, star int) *Civ {
	c := spawnAt(w, star, species.Fixed("conqueror"))
	c.Met = map[int]bool{}
	return c
}
