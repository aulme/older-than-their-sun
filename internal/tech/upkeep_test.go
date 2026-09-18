package tech

import (
	"testing"

	"worldgen/internal/flow"
)

func TestUpkeep(t *testing.T) {
	cases := []struct {
		key  string
		want flow.Income
	}{
		{"tools", flow.Income{}},             // era 0
		{"gathering", flow.Income{}},         // a kind node
		{"steam", flow.Income{}},             // the named exception
		{"agriculture", flow.Income{}},       // a producer
		{"medicine", flow.Income{1, 0, 0}},   // biology era 1
		{"printing", flow.Income{0, 1, 0}},   // society era 1
		{"industrial", flow.Income{0, 0, 1}}, // industry era 1
		{"breach", flow.Income{0, 0, 1}},     // a world node: one of its domain's kind
		{"cold_chemistry", flow.Income{0, 1, 0}},
		{"rocketry", flow.Income{0, 1, 1}},           // propulsion era 2
		{"closed_ecologies", flow.Income{1, 1, 0}},   // biology era 3
		{"antimatter", flow.Income{0, 2, 0}},         // energy era 3
		{"relativistic", flow.Income{0, 1, 2}},       // propulsion era 3
		{"panspermia", flow.Income{2, 1, 0}},         // biology era 4
		{"substrate_minds", flow.Income{0, 3, 0}},    // computation era 4
		{"near_light", flow.Income{0, 2, 3}},         // propulsion era 4
		{"ansible", flow.Income{0, 3, 0}},            // a miracle of the exotic
		{"directed_evolution", flow.Income{3, 0, 0}}, // a miracle of biology
		{"ftl", flow.Income{0, 0, 3}},                // a miracle of propulsion
		{"living_ships", flow.Income{2, 1, 0}},       // grown, not built
		{"fusion", flow.Income{}},
	}
	for _, c := range cases {
		if got := Get(c.key).Upkeep(); got != c.want {
			t.Errorf("%s: upkeep %v, want %v", c.key, got, c.want)
		}
	}
}

func TestCat(t *testing.T) {
	cases := map[string]flow.Category{
		"agriculture": flow.Fields, "medicine": flow.Fields, "life_extension": flow.Fields,
		"genetics": flow.Works, "terraforming": flow.Works, "industrial": flow.Works, "fusion": flow.Works,
		"firearms": flow.Arms, "nova_bombs": flow.Arms,
		"computers": flow.Mind, "printing": flow.Mind, "physics": flow.Mind,
		"rocketry": flow.Road, "ftl": flow.Road,
	}
	for k, want := range cases {
		if got := Get(k).Cat(); got != want {
			t.Errorf("%s: category %s, want %s", k, got, want)
		}
	}
	for _, n := range Nodes {
		if _, ok := groupOf[n.Domain]; !ok {
			t.Errorf("%s: domain %q has no upkeep group", n.Key, n.Domain)
		}
	}
}

// TestStructureTable: every structure has a node that unlocks it, an
// upkeep, a category, and the yielders are one per what they harness.
func TestStructureTable(t *testing.T) {
	for _, k := range StructureKeys {
		st := Structures[k]
		if st.Key != k {
			t.Errorf("structure %s has key %s", k, st.Key)
		}
		if Get(st.Node) == nil {
			t.Errorf("structure %s: unknown node %s", k, st.Node)
		}
		if st.Upkeep == (flow.Income{}) {
			t.Errorf("structure %s costs nothing", k)
		}
		if st.Yields() && st.Per != "star" && st.Per != "belt" {
			t.Errorf("structure %s is per %q", k, st.Per)
		}
	}
	if u := Get("vacuum_energy").Upkeep(); u != (flow.Income{flow.M: 1}) {
		t.Errorf("the vacuum tap costs %v, want 1 M", u)
	}
	if u := Get("dyson").Upkeep(); u != (flow.Income{}) {
		t.Errorf("Dyson swarms cost %v, want nothing: the swarm pays", u)
	}
}

// TestNoCatch22: harnessing anything needs a structure behind a node, and
// nothing in that node's closure may be impossible without a rarity, or a
// people could never reach the rarity that would let it reach the rarity.
// No node is hard-gated; this keeps it so.
func TestNoCatch22(t *testing.T) {
	for _, k := range StructureKeys {
		for _, n := range Closure(Structures[k].Node) {
			if g := Get(n).Gated; g != "" {
				t.Errorf("structure %s needs %s, which is gated by %s", k, n, g)
			}
		}
	}
	for _, n := range Nodes {
		if n.Gated != "" {
			t.Errorf("%s is gated by %s: a grant halves the price, it never bars the door", n.Key, n.Gated)
		}
	}
}
