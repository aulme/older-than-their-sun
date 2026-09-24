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
// wearing went to every fourth tick: every history was renumbered once;
// and at continuity (step 8), where a people's bodies got a span and its
// past a continuity the wearing, the ways setting and a dark age read,
// and the upload got its filter; and at leaders (step 9), where a
// people can follow a named leader and faces the succession when it is
// lost; and at the war council (step 10, stage 1), where each side of
// every war presses, holds or sues, wars have aims and end on terms, and
// will follows the fighting; and at allies and old enemies (step 11,
// stage 2), where an ally prices a separate peace and stands with its
// principal, and old enemies want each other and ask more each war. A
// refactor that must not
// change behaviour keeps it green; a step that changes the sim on purpose
// re-pins it and says so in its commit.
const sameHistoryDigest = "183078843f52008a74144f637677f9e61e33f8e638d76a5362426ccb2c541d66"

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
