package history

import (
	"testing"

	"worldgen/internal/species"
)

// TestCivilWarDeals: every holding goes to exactly one heir and every heir
// has a world; the old people ends Sundered with no trace left and no heir
// sharing its ID; the heirs are pairwise at war with claims on the whole
// realm; the works and the guards go with their worlds.
func TestCivilWarDeals(t *testing.T) {
	w, c := realm(t, 40, 7)
	c.Works = append(c.Works, Work{Key: "collectors", Node: "fusion", Star: 3, Legacy: -1})
	c.Structures["collectors"]++
	g := w.addGuard(c, 5, 4)
	x := spawnAt(w, 30, species.Fixed("cooperative", "defensive"))
	x.Grudge[c.ID], x.Truce[c.ID] = 2, w.Now+50_000
	x.Met[c.ID], c.Met[x.ID] = true, true
	traces := len(w.Traces)
	if !w.civilWar(c) {
		t.Fatal("seven worlds made no civil war")
	}
	heirs := w.Civs[2:]
	if len(heirs) < 2 || len(heirs) > 3 {
		t.Fatalf("%d heirs", len(heirs))
	}
	if c.Fate != Sundered || c.Living() || len(c.Systems) != 0 || len(w.Traces) != traces {
		t.Errorf("the old people: fate %s, living %v, worlds %d, traces %d to %d", c.Fate, c.Living(), len(c.Systems), traces, len(w.Traces))
	}
	seen := map[int]int{}
	for _, h := range heirs {
		if h.ID == c.ID || len(h.Systems) == 0 || h.Peak != len(h.Systems) {
			t.Errorf("the %s: id %d, worlds %v, peak %d", h.Tok(), h.ID, h.Systems, h.Peak)
		}
		for _, s := range h.Systems {
			seen[s]++
			if w.Owner[s] != h.ID {
				t.Errorf("%s dealt to the %s is owned by %d", w.star(s), h.Tok(), w.Owner[s])
			}
		}
		for s := 0; s < 7; s++ {
			if !h.Claim[s] {
				t.Errorf("the %s hold no claim on %s", h.Tok(), w.star(s))
			}
		}
		if len(h.Line) != 1 || h.Line[0] != c.ID || h.Origin.Key == "" {
			t.Errorf("the %s: line %v, origin %v", h.Tok(), h.Line, h.Origin)
		}
		if x.Grudge[h.ID] != 1 || x.Truce[h.ID] != x.Truce[c.ID] || !x.Met[h.ID] {
			t.Errorf("the stranger's grudge on the %s is %g, truce %d, met %v", h.Tok(), x.Grudge[h.ID], x.Truce[h.ID], x.Met[h.ID])
		}
	}
	for s := 0; s < 7; s++ {
		if seen[s] != 1 {
			t.Errorf("%s dealt %d times", w.star(s), seen[s])
		}
	}
	for i, a := range heirs {
		for _, b := range heirs[i+1:] {
			if !a.Wars[b.ID] || w.warBetween(a.ID, b.ID) == nil || a.Grudge[b.ID] != 3 {
				t.Errorf("the %s and the %s are not at war over the sundering", a.Tok(), b.Tok())
			}
		}
	}
	for _, h := range heirs {
		if w.Owner[3] == h.ID && (len(h.Works) != 1 || h.Structures["collectors"] != 1) {
			t.Error("the work did not go with its world")
		}
		if w.Owner[5] == h.ID && g.Owner != h.ID {
			t.Error("the guard did not go with its world")
		}
		if w.Owner[3] != h.ID && len(h.Works) != 0 {
			t.Error("a work went to the wrong heir")
		}
	}
	// the sundering is in every heir's telling as its own, and the old people's deeds are its deeds
	for _, h := range heirs {
		if w.regard(h, c.ID) != 2 {
			t.Error("an heir does not hold the old people as itself")
		}
	}
}

// TestReclaim: a claimed world taken is restored to the realm, and an heir
// holding the whole old realm has no claims left.
func TestReclaim(t *testing.T) {
	w, c := realm(t, 41, 5)
	if !w.civilWar(c) {
		t.Fatal("no civil war")
	}
	a, b := w.Civs[1], w.Civs[2]
	if len(b.Systems) < 2 {
		a, b = b, a
	}
	if len(b.Systems) < 2 {
		t.Skip("the dealing gave nobody a colony")
	}
	wr := w.warBetween(a.ID, b.ID)
	var t2 int
	for _, s := range b.Systems {
		if s != b.Home {
			t2 = s
		}
	}
	facts := len(w.Events)
	w.takeWorld(wr, a, b, t2)
	if w.Owner[t2] != a.ID {
		t.Fatalf("the world was not taken")
	}
	found := false
	for _, f := range w.Events[facts:] {
		found = found || f.Kind == FReclaimed
	}
	if !found || a.Claim == nil {
		t.Error("no reclaiming was written, or the claims went early")
	}
	// the rest of the old realm changes hands: the claims run out
	for _, h := range w.Civs[1:] {
		if h != a {
			for _, s := range append([]int(nil), h.Systems...) {
				w.Owner[s] = a.ID
				a.Systems = append(a.Systems, s)
			}
		}
	}
	w.reclaimed(a, b, b.Home)
	if a.Claim != nil {
		t.Error("an heir holding the whole old realm still holds claims")
	}
}

// TestShatter: a dark age that takes the stars shatters a people of more
// than one world into one people per world, capped at eight, kin to each
// other and at war with nobody; a people that keeps the stars, or holds
// one world, does not shatter.
func TestShatter(t *testing.T) {
	w, c := realm(t, 42, 12)
	x := spawnAt(w, 30, species.Fixed("cooperative", "defensive"))
	x.Grudge[c.ID] = 2
	w.shatter(c, because("ossified"), nil)
	shards := w.Civs[2:]
	if len(shards) != 8 || c.Fate != Shattered {
		t.Fatalf("%d shards, fate %s", len(shards), c.Fate)
	}
	for _, s := range shards {
		if len(s.Systems) != 1 || w.Owner[s.Home] != s.ID || len(s.Wars) != 0 || s.Claim != nil {
			t.Errorf("the %s: worlds %v, wars %d, claims %d", s.Tok(), s.Systems, len(s.Wars), len(s.Claim))
		}
		if x.Grudge[s.ID] != 1 {
			t.Errorf("the stranger's grudge on the %s is %g", s.Tok(), x.Grudge[s.ID])
		}
	}
	for i, a := range shards {
		for _, b := range shards[i+1:] {
			if !w.kin(a, b) || w.regard(a, b.ID) != 1 || a.Grudge[b.ID] != 0 {
				t.Errorf("the %s and the %s: kin %v, regard %d", a.Tok(), b.Tok(), w.kin(a, b), w.regard(a, b.ID))
			}
		}
	}
	abandoned := 0
	for _, tr := range w.Traces {
		if tr.Civ == c.ID {
			abandoned++
		}
	}
	if abandoned != 4 {
		t.Errorf("%d worlds abandoned past the cap", abandoned)
	}
	// a dark age shatters only when the forgetting takes the stars
	w, c = realm(t, 43, 3)
	w.darkAge(c, because("ossified"))
	if (c.Fate == Shattered) != (c.Reach < 10) {
		t.Errorf("reach %.0f after the dark age, fate %s", c.Reach, c.Fate)
	}
	// one world cannot shatter
	w, c = realm(t, 44, 1)
	c.Known = map[string]bool{}
	w.darkAge(c, because("ossified"))
	if c.Fate == Shattered || !c.Active() {
		t.Error("a one-world people shattered or ended")
	}
}

// TestKinDepth: kin three steps apart are still kin, and regard for the
// line is self at every step.
func TestKinDepth(t *testing.T) {
	w, c := realm(t, 45, 6)
	if !w.civilWar(c) {
		t.Fatal("no civil war")
	}
	a := w.Civs[1]
	w.Owner[20], w.Owner[21] = a.ID, a.ID
	a.Systems = append(a.Systems, 20, 21)
	n := len(w.Civs)
	if !w.civilWar(a) {
		t.Fatal("no second civil war")
	}
	b := w.Civs[n]
	w.Owner[22] = b.ID
	b.Systems = append(b.Systems, 22)
	n = len(w.Civs)
	if !w.civilWar(b) {
		t.Fatal("no third civil war")
	}
	d := w.Civs[n]
	if len(d.Line) != 3 || !w.kin(d, w.Civs[2]) || w.regard(d, c.ID) != 2 || w.regard(d, a.ID) != 2 {
		t.Errorf("three steps down: line %v, kin to the first heir %v", d.Line, w.kin(d, w.Civs[2]))
	}
	if w.kinship(d, &Legacy{Maker: c.ID}) != 2 {
		t.Error("the old realm's works are not the shard's own")
	}
}

// TestClaimsFade: the claims clear when the sundering wears to myth.
func TestClaimsFade(t *testing.T) {
	w, c := realm(t, 46, 2)
	if !w.civilWar(c) {
		t.Fatal("no civil war")
	}
	a := w.Civs[1]
	for _, tl := range a.Lore {
		f := w.Events[tl.Fact]
		if f.Kind == FSundered && f.Object == a.ID {
			tl.Wear = 1
			w.wearStep(a, tl, f)
		}
	}
	if a.Claim != nil {
		t.Error("the claims did not clear at myth")
	}
}

// TestKinMeet: kin with no feud between them open trade at once when they
// meet and skip the council; a feud keeps the council.
func TestKinMeet(t *testing.T) {
	w, c := realm(t, 47, 12)
	w.shatter(c, because("ossified"), nil)
	a, b := w.Civs[1], w.Civs[2]
	for _, x := range []*Civ{a, b} {
		x.Reach, x.Speed = 60, 20
	}
	w.meet(a, b, -1)
	if !a.Trade[b.ID] || !a.Fathomed[b.ID] {
		t.Errorf("kin met: trade %v, fathomed %v", a.Trade[b.ID], a.Fathomed[b.ID])
	}
	if _, wants, _ := w.bar(a, b); wants {
		t.Error("kin with no grudge are wanted on posture alone")
	}
}
