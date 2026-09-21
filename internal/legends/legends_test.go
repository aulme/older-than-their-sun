package legends

import (
	"strings"
	"testing"

	"worldgen/internal/history"
)

// TestResolved: the legends of a seed carry no unresolved name token;
// every {kind:id} the history printed went through the names pass.
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
}
