package legends

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"worldgen/internal/history"
)

// sameLegendsDigest is the digest of the full legends of seed 5 at 200
// stars, pinned at the events step (specs/plan.md step 2), where the
// view began rendering the chronicle from the record, and re-pinned at
// names-translated (step 4), where the translated names and the
// technical terms replaced the stubs and the miracle names while the
// history stayed the same. A step that must not change what the view
// prints keeps it green; one that changes the view or the history on
// purpose re-pins it and says so in its commit.
const sameLegendsDigest = "836a8a186c3e5fb1ef35c41597202e54f9c731c9b185931e5ae1fd4060a9a372"

// TestResolved: the legends of a seed carry no unresolved name token;
// every {kind:id} the history printed went through the names pass. And
// the legends are what they were.
func TestResolved(t *testing.T) {
	cfg := history.DefaultConfig()
	cfg.Stars = 200
	w := history.Generate(5, cfg)
	var sb strings.Builder
	Write(&sb, w, true)
	out := sb.String()
	for _, k := range []string{"{civ:", "{star:", "{plague:", "{war:", "{elder:", "{word:", "{title:", "{species:", "{makers:", "{source:", "{^"} {
		if i := strings.Index(out, k); i >= 0 {
			t.Errorf("an unresolved token: %q", out[max(0, i-40):min(i+60, len(out))])
		}
	}
	sum := sha256.Sum256([]byte(out))
	if got := hex.EncodeToString(sum[:]); got != sameLegendsDigest {
		t.Errorf("the legends of seed 5 changed: digest %s, pinned %s", got, sameLegendsDigest)
	}
}
