package history

import (
	"bytes"
	"os"
	"testing"
)

// TestWatchSameHistory: a watched run is the same history as an
// unwatched one, and a watch with low thresholds fires and dumps.
func TestWatchSameHistory(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Stars = 200
	done := make(chan string)
	go func(cfg Config) { done <- historyDigest(Generate(11, cfg)) }(cfg)
	var out bytes.Buffer
	wt := &Watch{PairWars: 3, MasterChanges: 2, WarTicks: 20, WarJump: 5, Out: &out}
	wt.Attach(&cfg)
	w := Generate(11, cfg)
	plain := <-done
	if got := historyDigest(w); got != plain {
		t.Fatalf("the watch changed the history: %s against %s", got, plain)
	}
	if len(wt.Fired) == 0 || out.Len() == 0 {
		t.Fatalf("nothing fired at low thresholds over %d wars", len(w.Wars))
	}
	if os.Getenv("WATCH_V") != "" {
		t.Log(out.String())
	}
}
