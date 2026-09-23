package history

import (
	"slices"
	"testing"

	"worldgen/internal/tech"
)

// knownWalk is what knownOf gives the long way, without the kept list.
func knownWalk(c *Civ) []string {
	var out []string
	for _, n := range tech.Nodes {
		if c.Known[n.Key] {
			out = append(out, n.Key)
		}
	}
	return out
}

// TestKeptOrders is what makes the kept orders safe to keep. A cache
// whose writers have to remember to drop it is the shape the
// architecture rules are wary of, so the remembering is not left to
// anyone: at every tick of a whole age, every kept order that claims to
// stand is held against a fresh sort of the set it came from. A write
// that goes round startTrade, endTrade, addMet, dropMet, know or
// forgetNode without dropping the order fails here on the tick after
// it, size or no size.
func TestKeptOrders(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Stars = 200
	checked := 0
	cfg.Sample = func(w *World) {
		for _, c := range w.Civs {
			if c.tradeOrder.ok {
				if want := sortedInts(c.Trade); !slices.Equal(c.tradeOrder.ids, want) {
					t.Fatalf("at %d, %s keeps partners %v and its roads are %v", w.Now, c.Tok(), c.tradeOrder.ids, want)
				}
				checked++
			}
			if c.knownOK {
				if want := knownWalk(c); !slices.Equal(c.knownKeys, want) {
					t.Fatalf("at %d, %s keeps %d nodes and knows %d", w.Now, c.Tok(), len(c.knownKeys), len(want))
				}
				checked++
			}
			if c.metOrder.ok {
				if want := sortedInts(c.Met); !slices.Equal(c.metOrder.ids, want) {
					t.Fatalf("at %d, %s keeps a roll of %d and has heard of %d", w.Now, c.Tok(), len(c.metOrder.ids), len(want))
				}
				checked++
			}
		}
	}
	Generate(11, cfg)
	if checked < 10_000 {
		t.Fatalf("checked %d kept orders: too few to hold anything", checked)
	}
}
