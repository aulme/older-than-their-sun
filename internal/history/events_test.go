package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestKindsGenerated: kinds_gen.go is what data/events.json says.
func TestKindsGenerated(t *testing.T) {
	have, err := os.ReadFile("kinds_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(have) != KindsSource() {
		t.Fatal("kinds_gen.go is stale: run go generate ./internal/history")
	}
}

// TestParams: every event of a run carries the keys its kind declares
// and no others, is placed in the chronicle once, and renders.
func TestParams(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stars = 200
	w := Generate(5, cfg)
	placed := map[int]int{}
	for i, e := range w.Chronicle {
		if e == nil {
			t.Fatalf("chronicle slot %d was reserved and never filled", i)
		}
		placed[e.ID]++
	}
	if len(w.Chronicle) != len(w.Events) {
		t.Errorf("%d events, %d in the chronicle", len(w.Events), len(w.Chronicle))
	}
	seen := map[Kind]bool{}
	for i, e := range w.Events {
		if e.ID != i {
			t.Fatalf("event %d has id %d", i, e.ID)
		}
		if placed[e.ID] != 1 {
			t.Errorf("event %d (%s) is in the chronicle %d times", e.ID, e.Kind, placed[e.ID])
		}
		sh, ok := shapes[e.Kind]
		if !ok {
			t.Fatalf("event %d is of a kind the table lacks: %q", e.ID, e.Kind)
		}
		declared := map[string]bool{}
		for _, p := range sh.Params {
			key, optional := strings.CutSuffix(p, "?")
			declared[key] = true
			if _, has := e.P[key]; !has && !optional {
				t.Errorf("%s lacks %q: %s", e.Kind, key, e)
			}
		}
		for k := range e.P {
			if !declared[k] {
				t.Errorf("%s carries %q, which it does not declare: %s", e.Kind, k, e)
			}
		}
		if e.IsFact() && e.Subject < 0 {
			t.Errorf("fact %s with no subject: %s", e.Kind, e)
		}
		if line := w.Line(e); strings.Contains(line, "{") && !strings.Contains(line, ":") {
			t.Errorf("%s renders with a placeholder left: %s", e.Kind, line)
		}
		seen[e.Kind] = true
	}
	t.Logf("%d events of %d kinds, %d kinds of %d in the table seen", len(w.Events), len(seen), len(seen), len(kindDefs))
}

// TestLinesUnread: the chronicle's templates and the view's renderers
// are the view's (lines.go, describe.go); nothing in the simulation
// reads them, so the text is free to change. The tellings (telling.go)
// render too, and the simulation reads only their weights.
func TestLinesUnread(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") || file == "lines.go" || file == "describe.go" || file == "telling.go" {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, ident := range []string{"w.Line(", "lines[", "lineFns[", "facedLines[", "blastLines[", "harnessLines[", "rarityLines[", "roadPhrases[", "beneathNames[",
			"legacyDesc(", "describeAt(", "whyText(", "reasonText(", "sourceName(", "traceName(", "OriginText(", "CauseText(", "IntoText(", "WarCause(", "WarResult(", "BetrayalText(", "RecordText(", "useName(", "driftText(", "blastText(", "portraitText(", "conditionText(", "traceText("} {
			if strings.Contains(string(src), ident) {
				t.Errorf("%s reads the chronicle's templates (%s)", file, ident)
			}
		}
	}
}
