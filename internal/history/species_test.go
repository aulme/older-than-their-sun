package history

import (
	"testing"

	"worldgen/internal/species"
)

// TestHeirsShareBlood: the heirs of a civil war keep the old people's
// species, take names of their own, and are of its line.
func TestHeirsShareBlood(t *testing.T) {
	w := newTestWorld(t, 1, 30)
	c := spawnAt(w, 0, species.Fixed("collective", "defensive", "practical", "curious"))
	w.Owner[1] = c.ID
	c.Systems = append(c.Systems, 1)
	if !w.civilWar(c) {
		t.Fatal("two worlds did not make a civil war")
	}
	if len(w.Civs) != 3 {
		t.Fatalf("%d peoples after the sundering", len(w.Civs))
	}
	for _, nc := range w.Civs[1:] {
		if nc.Species != c.Species {
			t.Fatal("an heir has its own species")
		}
		if nc.Name == c.Name || nc.Name == c.Species.Name {
			t.Fatalf("an heir is named %q, its parent %q", nc.Name, c.Name)
		}
		if nc.Origin == "" || len(w.Species) != 1 || c.Species.ID != 0 {
			t.Fatalf("origin %q, %d species in the world", nc.Origin, len(w.Species))
		}
		if w.kinship(nc, &Legacy{Maker: c.ID}) != 2 {
			t.Fatal("an heir should find the old people's works its own")
		}
	}
}

// TestMachineSuccessorIsNewBlood: what a people builds that outgrows them
// is a new species, a machine one, under its own name.
func TestMachineSuccessorIsNewBlood(t *testing.T) {
	w := newTestWorld(t, 3, 30)
	c := spawnAt(w, 0, species.Fixed("collective", "defensive", "practical", "curious"))
	nc := w.machinePeople(c)
	if nc.Species == c.Species || nc.Species.Sub != species.Machine {
		t.Fatalf("the successor is %s and shares blood: %v", nc.Species.Nature(), nc.Species == c.Species)
	}
	if nc.Name != nc.Species.Name || nc.Species.ID != 1 || len(w.Species) != 2 {
		t.Fatalf("name %q for species %q (id %d of %d)", nc.Name, nc.Species.Name, nc.Species.ID, len(w.Species))
	}
	if w.kinship(nc, &Legacy{Maker: c.ID}) != 0 {
		t.Fatal("machines are not their makers' blood")
	}
}

// TestDifferenceOrder: the order term is what it was when the hive was an
// organisation trait, and nobody home adds two.
func TestDifferenceOrder(t *testing.T) {
	hive := species.Fixed("defensive")
	hive.Mods = species.Hive
	ind := species.Fixed("individualist", "defensive")
	if d := difference(hive, ind); d != 1 {
		t.Fatalf("hive against individualist: %g, want 1", d)
	}
	hive2 := species.Fixed("defensive")
	hive2.Mods = species.Hive
	if d := difference(hive, hive2); d != 0 {
		t.Fatalf("hive against hive: %g, want 0", d)
	}
	un := species.Fixed("defensive")
	un.Mods = species.Unconscious
	if d := difference(hive, un); d != 3 {
		t.Fatalf("hive against unconscious: %g, want the order's 1 and no one home's 2", d)
	}
	swarm := species.Fixed("individualist", "defensive", "swarming")
	if d := difference(swarm, ind); d != 2.5 {
		t.Fatalf("swarm against its like: %g, want 2.5", d)
	}
}
