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
// digest went from the prose to the typed record, and at lookups (step
// 3), where the record's words became keys and ids; the legends were
// byte-identical across both. A refactor that must not change behaviour
// keeps it green; a step that changes the sim on purpose re-pins it and
// says so in its commit.
const sameHistoryDigest = "742802754e000f38bbb6461e9aac550df5b57f6829e73972fac30ea569235c02"

func historyDigest(w *World) string {
	h := sha256.New()
	for _, e := range w.Events {
		fmt.Fprintln(h, e.String())
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
