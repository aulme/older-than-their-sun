package history

import (
	"testing"

	"worldgen/internal/species"
)

// TestSchismSharesBlood: a branch that declares itself a new people keeps
// its parent's species and takes a name of its own.
func TestSchismSharesBlood(t *testing.T) {
	for seed := uint64(1); seed < 60; seed++ {
		w := newTestWorld(t, seed, 30)
		c := spawnAt(w, 0, species.Fixed("collective", "defensive", "practical", "curious"))
		w.Owner[1] = c.ID
		c.Systems = append(c.Systems, 1)
		w.schism(c)
		if len(w.Civs) < 2 {
			continue
		}
		nc := w.Civs[1]
		if nc.Species != c.Species {
			t.Fatalf("seed %d: the branch has its own species", seed)
		}
		if nc.Name == c.Name || nc.Name == c.Species.Name {
			t.Fatalf("seed %d: the branch is named %q, its parent %q", seed, nc.Name, c.Name)
		}
		if nc.Origin == "" || len(w.Species) != 1 || c.Species.ID != 0 {
			t.Fatalf("seed %d: origin %q, %d species in the world", seed, nc.Origin, len(w.Species))
		}
		if w.kinship(nc, &Legacy{Maker: c.ID}) != 1 {
			t.Fatal("a branch should be kin to its parent's works")
		}
		return
	}
	t.Fatal("no schism made a branch in sixty seeds")
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
// organisation trait.
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
	if d := difference(hive, un); d != 1 {
		t.Fatalf("hive against unconscious: %g, want 1", d)
	}
	swarm := species.Fixed("individualist", "defensive", "swarming")
	if d := difference(swarm, ind); d != 2.5 {
		t.Fatalf("swarm against its like: %g, want 2.5", d)
	}
}
