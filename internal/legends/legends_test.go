package legends

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"worldgen/internal/history"
	"worldgen/internal/record"
	"worldgen/internal/writer"
)

// sameLegendsDigest is the digest of the full legends of seed 5 at 200
// stars, pinned at the events step (specs/plan.md step 2), where the
// view began rendering the chronicle from the record, and re-pinned at
// names-translated (step 4), where the translated names and the
// technical terms replaced the stubs and the miracle names while the
// history stayed the same. At the writer step (step 5) the view began
// reading the run directory instead of the world, and the digest held.
// A step that must not change what the view prints keeps it green; one
// that changes the view or the history on purpose re-pins it and says
// so in its commit.
const sameLegendsDigest = "836a8a186c3e5fb1ef35c41597202e54f9c731c9b185931e5ae1fd4060a9a372"

// run is seed 5 at 200 stars, written to a directory and read back, so
// that the view is tested on the files alone.
func run(t *testing.T) *record.Run {
	t.Helper()
	cfg := history.DefaultConfig()
	cfg.Stars = 200
	w := history.Generate(5, cfg)
	dir := filepath.Join(t.TempDir(), "5")
	if err := writer.Write(dir, w, "sol"); err != nil {
		t.Fatal(err)
	}
	r, err := record.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// TestResolved: the legends of a seed, rendered from the files, carry
// no unresolved name token and no placeholder; every {kind:id} went
// through the names table. And the legends are what they were.
func TestResolved(t *testing.T) {
	r := run(t)
	var sb strings.Builder
	Write(&sb, r, true)
	out := sb.String()
	for _, k := range []string{"{civ:", "{star:", "{plague:", "{war:", "{elder:", "{word:", "{title:", "{species:", "{makers:", "{source:", "{^"} {
		if i := strings.Index(out, k); i >= 0 {
			t.Errorf("an unresolved token: %q", out[max(0, i-40):min(i+60, len(out))])
		}
	}
	v := load(r)
	for _, e := range r.Chronicle {
		if line := v.line(e); strings.Contains(line, "{") && !strings.Contains(line, ":") {
			t.Errorf("%s renders with a placeholder left: %s", e.Kind, line)
		}
	}
	sum := sha256.Sum256([]byte(out))
	if got := hex.EncodeToString(sum[:]); got != sameLegendsDigest {
		if path := os.Getenv("LEGENDS_OUT"); path != "" {
			os.WriteFile(path, []byte(out), 0o644)
		}
		t.Errorf("the legends of seed 5 changed: digest %s, pinned %s", got, sameLegendsDigest)
	}
}
