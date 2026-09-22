package history

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// sameHistoryDigest is the digest of every event of seed 11 at 200 stars,
// pinned at the start of step 2 (the mind), re-pinned at names-ids
// (specs/plan.md step 1), where the history stopped drawing for names
// and every seed shifted once, re-pinned at events (step 2), where the
// digest went from the prose to the typed record, at lookups (step 3),
// where the record's words became keys and ids, and at the streams (step
// 7), where the one random source became a stream per phase and per
// people, the decline index took over the waning and the end, and the
// wearing went to every fourth tick: every history was renumbered once. A refactor that must not
// change behaviour keeps it green; a step that changes the sim on purpose
// re-pins it and says so in its commit.
const sameHistoryDigest = "2312728f2718c58ff9e61d5a75778ee2d0b362a1c43972b98d60532047007f9c"

func historyDigest(w *World) string {
	h := sha256.New()
	for _, e := range w.Events {
		if !e.Silent() {
			fmt.Fprintln(h, e.String()) // the told events; the silent kinds are the fold test's
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

// TestSameHistory: the events of one seed are what they were.
func TestSameHistory(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Stars = 200
	w := Generate(11, cfg)
	if got := historyDigest(w); got != sameHistoryDigest {
		t.Fatalf("history of seed 11 changed: digest %s, pinned %s (%d events)", got, sameHistoryDigest, len(w.Events))
	}
}
