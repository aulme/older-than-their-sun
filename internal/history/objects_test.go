package history

import (
	"testing"

	"worldgen/internal/flow"
	"worldgen/internal/species"
)

// TestEmberDims: a pocket-star Ember yields eight energy to its holder and
// dims by two each time it is moved.
func TestEmberDims(t *testing.T) {
	w := newTestWorld(t, 25, 30)
	c := spawnAt(w, 0, species.Fixed("cooperative"))
	e := spawnAt(w, 1, species.Fixed("cooperative"))
	s := w.makeObject(c, "ember", nil, &Source{Key: "ember", Form: "pocket_star"}, "made")
	if got := w.mobileYield(c)[flow.E]; got != emberYield {
		t.Fatalf("the Ember yields %v E, want %v", got, emberYield)
	}
	if !c.miracle("ember") || !w.holdsObject(c, "ember") {
		t.Fatal("the maker does not hold the Ember")
	}
	w.transfer(s, c, e)
	if got := w.mobileYield(e)[flow.E]; got != emberYield-pocketDim {
		t.Errorf("moved once, the Ember yields %v E, want %v", got, emberYield-pocketDim)
	}
	if w.mobileYield(c)[flow.E] != 0 || c.Remade["ember"] != w.Now {
		t.Errorf("the loser still draws on it, or was not marked: %v %v", w.mobileYield(c), c.Remade)
	}
	if !e.miracle("ember") {
		t.Error("the taker does not hold the Ember by wielding it")
	}
}

// TestObjectRemade: a people that knows the Ember and holds none makes
// one, and after losing it, another a million years on and not before.
func TestObjectRemade(t *testing.T) {
	w := newTestWorld(t, 26, 30)
	c := spawnAt(w, 0, species.Fixed("cooperative"))
	c.Known["ember"] = true
	w.objects(c)
	if !w.holdsObject(c, "ember") {
		t.Fatal("knowing the Ember, none was made")
	}
	var s *Source
	for _, id := range w.mobile {
		if w.Sources[id].Holder == c.ID {
			s = w.Sources[id]
		}
	}
	w.lose(s, c, "test")
	w.objects(c)
	if w.holdsObject(c, "ember") {
		t.Fatal("another was made at once")
	}
	w.Now += remakeYears
	w.objects(c)
	if !w.holdsObject(c, "ember") {
		t.Error("a million years on, no other was made")
	}
}

// TestMannaCutting: a cutting given to a partner in want of organic matter
// is an instance of the giver's form, and the giver keeps its own.
func TestMannaCutting(t *testing.T) {
	w, a, b := partners(t, 27)
	s := w.makeObject(a, "manna", nil, &Source{Key: "manna", Form: "mould"}, "made")
	b.OwnWant = flow.Income{flow.O: 3}
	w.dt = 1000 // the chance is certain
	w.cutting(a, b)
	if !w.holdsObject(b, "manna") || !w.holdsObject(a, "manna") {
		t.Fatalf("after the cutting a holds %v, b holds %v", w.holdsObject(a, "manna"), w.holdsObject(b, "manna"))
	}
	if s.Given != 1 {
		t.Errorf("the giver's count is %d", s.Given)
	}
	for _, id := range w.mobile {
		if x := w.Sources[id]; x.Holder == b.ID && x.Form != "mould" {
			t.Errorf("b's Manna is %s, want the giver's mould", x.Form)
		}
	}
	if !b.miracle("manna") {
		t.Error("b does not hold the Manna by its cutting")
	}
	// nobody gives a second to a people that has one
	w.cutting(a, b)
	if s.Given != 1 {
		t.Errorf("a second cutting went to a people that had one: %d", s.Given)
	}
}
