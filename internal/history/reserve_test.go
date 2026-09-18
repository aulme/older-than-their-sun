package history

import (
	"testing"

	"worldgen/internal/flow"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// TestLaunchNeedsSpare: a fleet is a reservation of the means. With no
// spare it does not go; with spare it goes and the spare is taken.
func TestLaunchNeedsSpare(t *testing.T) {
	w := newTestWorld(t, 12, 30)
	c := spawnAt(w, 0, species.Fixed("cooperative"))
	e := spawnAt(w, 1, species.Fixed("cooperative"))
	c.Mil = 4
	c.Surplus = flow.Income{}
	if x := w.launch(c, Scout, e, e.Home, 1); x != nil {
		t.Fatal("a scout sailed with nothing spare")
	}
	c.Surplus = flow.Income{flow.E: 1.5, flow.M: 1.5}
	x := w.launch(c, Scout, e, e.Home, 1)
	if x == nil {
		t.Fatal("a scout stayed home with the spare to cover it")
	}
	if c.Reserved != (flow.Income{flow.E: 1, flow.M: 1}) {
		t.Errorf("the scout reserved %v, want 1 E 1 M", c.Reserved)
	}
	// a second this tick finds the spare gone
	if y := w.launch(c, Scout, e, e.Home, 1); y != nil {
		t.Error("a second scout sailed on spare the first had taken")
	}
	// and the fleet out is a use in flight
	found := false
	for _, u := range w.uses(c) {
		if u.Key == "fleet:"+itoa(x.ID) {
			found = u.Flight && u.Need == (flow.Income{flow.E: 1, flow.M: 1})
		}
	}
	if !found {
		t.Error("the scout out is not a use in flight")
	}
}

// TestBuildByDeficit: a people short of metal with a belt at home puts
// mines in it; the mines double the belt; a second mine on one belt is
// refused.
func TestBuildByDeficit(t *testing.T) {
	w := newTestWorld(t, 13, 40)
	star := -1
	for i := range w.G.Stars {
		if len(w.G.Sys[i].Belts) == 1 && w.G.Sys[i].Home >= 0 {
			star = i
			break
		}
	}
	if star < 0 {
		t.Skip("no habitable star with one belt in the field")
	}
	c := spawnAt(w, star, species.Fixed("cooperative"))
	for _, k := range []string{"interplanetary", "orbital_habitats", "closed_ecologies", "defence_grid"} {
		c.Known[k] = true
	}
	w.recompute(c)
	before := w.yieldAt(c, star)[flow.M]
	c.Want = flow.Income{flow.M: 3}
	c.Surplus = flow.Income{flow.E: 4, flow.M: 4}
	w.Cfg.Tuning.Build.Rate = 1
	w.build(c)
	if c.Structures["mine"] != 1 {
		t.Fatalf("short of metal, built %v, want the mines", c.Structures)
	}
	if got := w.yieldAt(c, star)[flow.M]; got != before+beltYield {
		t.Errorf("with mines the belt yields %v M, was %v: want doubled", got, before)
	}
	for _, s := range w.sites(c) {
		if s.Key == "mine" {
			t.Errorf("a second mine is offered on the one belt")
		}
	}
	// short of energy instead: the collectors
	c.Want = flow.Income{flow.E: 3}
	w.build(c)
	if c.Structures["collectors"] != 1 {
		t.Fatalf("short of energy, built %v, want the collectors", c.Structures)
	}
	// the mines are a use with an upkeep, fed under the works
	for _, u := range w.uses(c) {
		if u.Key == "work:mine:"+itoa(star) && (u.Cat != flow.Works || u.Need != tech.Structures["mine"].Upkeep) {
			t.Errorf("the mines are fed under %s for %v", u.Cat, u.Need)
		}
	}
}
