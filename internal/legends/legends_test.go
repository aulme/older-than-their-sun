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
// It was re-pinned at the optimisation step, where a party that is the
// teller and the subject of its clause began to say "we" where it said
// "us" ({OS} and the reflexives in telling.go); that moved 840 lines of
// this seed and nothing else. It was re-pinned again at step 7, where
// the one random source became a stream per phase and per people, the
// decline index took over the waning and the end (and the waning's line
// began to say what it read), and the wearing went to every fourth
// tick: every history was renumbered once.
// It was re-pinned a third time at step 7's summaries, and the move is
// worth the note, since almost nothing moved: three lines of this seed's
// forty-nine thousand, all of them a people reading "a wise people" at
// the end that did not before. Wisdom counts a people's own woes and
// follies, and that count used to be left by the reckoning in the lore
// step, which runs after the levels are derived — so the level read a
// count a tick old. Kept as the telling changes, it is the count now,
// and three peoples of this seed cross eight. Not an event of the
// chronicle moved, here or in seed 11. Re-pinned at continuity (step
// 8), which changes every history on purpose, and at leaders (step 9),
// which does too and gives the chronicle and the tellings named figures.
// A step that must not change what the view prints keeps it green; one
// that changes the view or the history on purpose re-pins it and says
// so in its commit.
const sameLegendsDigest = "1b8861baf042dafef3907d164d0060c00840ac08193b3cf8b1cf1465aa67d73f"

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
