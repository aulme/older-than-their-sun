package history

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// sameHistoryDigest is the digest of every event of seed 11 at 200 stars,
// pinned at the start of step 2 (the mind). A refactor that must not change
// behaviour keeps it green; a step that changes the sim on purpose re-pins
// it and says so in its commit.
const sameHistoryDigest = "3732e6b62949c43728b2963123af568414b041aa5d4aeebc886c6737d85f1830"

func historyDigest(w *World) string {
	h := sha256.New()
	for _, e := range w.Events {
		fmt.Fprintf(h, "%d\t%s\n", e.Year, e.Text)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// TestSameHistory: the events of one seed are what they were.
func TestSameHistory(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stars = 200
	w := Generate(11, cfg)
	if got := historyDigest(w); got != sameHistoryDigest {
		t.Fatalf("history of seed 11 changed: digest %s, pinned %s (%d events)", got, sameHistoryDigest, len(w.Events))
	}
}
