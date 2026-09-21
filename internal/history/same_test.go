package history

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// sameHistoryDigest is the digest of every event of seed 11 at 200 stars,
// pinned at the start of step 2 (the mind) and re-pinned at names-ids
// (specs/plan.md step 1), where the history stopped drawing for names
// and every seed shifted once. A refactor that must not change behaviour
// keeps it green; a step that changes the sim on purpose re-pins it and
// says so in its commit.
const sameHistoryDigest = "9b97475c0264e4477984911174b5a1d0fc8ef2775d1b5e6df1f53ff623634dd9"

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
