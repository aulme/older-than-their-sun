package history

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// runWith is Generate with the world handed over once before it runs, so
// a test can put a phase in the tick.
func runWith(seed uint64, cfg Config, tweak func(*World)) *World {
	w := newWorld(seed, cfg)
	if tweak != nil {
		tweak(w)
	}
	w.runDeep()
	w.runAges()
	w.runAge()
	return w
}

// fieldDigest is the star field: where the stars are, what they are and
// what could live on them. It is drawn once, from the galaxy stream.
func fieldDigest(w *World) string {
	h := sha256.New()
	fmt.Fprintf(h, "%v\n", w.Law)
	for i := range w.G.Stars {
		s := &w.G.Stars[i]
		fmt.Fprintf(h, "%d %s %c %.6f %.6f %.6f %.4f %d %.0f\n", s.ID, s.Name, s.Class, s.X, s.Y, s.Z, s.Hab, s.Mult, s.Lifetime)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// TestStreamsAddAPhase is the gate the streams exist to buy (streams.go).
// A phase put in the tick that draws from its own stream and does
// nothing else changes no history at all: not a star, not a blood, not
// an event. With one stream its draws slid every later draw of every
// tick, so adding a subsystem — or moving one draw inside an existing
// one — made a different galaxy, and no two runs of the simulation could
// be compared.
func TestStreamsAddAPhase(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Stars = 120
	cfg.Until = 4_000_000
	plain := runWith(4, cfg, nil)
	noisy := runWith(4, cfg, func(w *World) {
		w.insertPhase("plagues", phase{"noise", func(w *World) {
			for range 1 + w.R.IntN(20) {
				w.R.Float64()
			}
		}})
	})
	if len(plain.Events) == 0 {
		t.Fatal("no events: the test is checking nothing")
	}
	if a, b := fieldDigest(plain), fieldDigest(noisy); a != b {
		t.Errorf("the field moved with a phase that only drew: %s against %s", a, b)
	}
	if len(plain.Events) != len(noisy.Events) {
		t.Fatalf("%d events against %d", len(plain.Events), len(noisy.Events))
	}
	for i := range plain.Events {
		if a, b := plain.Events[i].String(), noisy.Events[i].String(); a != b {
			t.Fatalf("event %d differs:\n  %s\n  %s", i, a, b)
		}
	}
}

// TestCivStreamIsTheIdAlone: a people's stream is a function of its id
// and the world's seed and of nothing else — not of how many peoples
// came before it, not of what any of them drew. That is what lets the
// peoples be stepped in any order, or beside each other, and what keeps
// one people's doings from renumbering another's history.
func TestCivStreamIsTheIdAlone(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stars = 20
	want := make([]float64, 6)
	for id := range want {
		w := newWorld(3, cfg)
		want[id] = w.civStream(&Civ{ID: id}).Float64()
	}
	w := newWorld(3, cfg)
	for id := range want {
		for range id * 7 { // whatever the peoples before it drew
			w.civStream(&Civ{ID: id - 1}).Float64()
		}
		if got := w.civStream(&Civ{ID: id}).Float64(); got != want[id] {
			t.Errorf("people %d: %v alone, %v in company", id, want[id], got)
		}
	}
}

// TestStreamNames: neighbouring names give unrelated streams, which is
// what hashing a name rather than drawing sub-seeds in turn is for.
func TestStreamNames(t *testing.T) {
	seen := map[uint64]string{}
	for _, name := range []string{"galaxy", "deep", "ages", "cycle", "age", "life", "plagues", "civs", "civ:0", "civ:1", "civ:2", "civ:11", "civ:12"} {
		s := streamSeed(7, name)
		if other, ok := seen[s]; ok {
			t.Errorf("%q and %q share a stream seed", name, other)
		}
		seen[s] = name
	}
	if streamSeed(7, "civs") == streamSeed(8, "civs") {
		t.Error("the same stream of two seeds is the same stream")
	}
}
