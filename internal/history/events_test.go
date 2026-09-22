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
	if string(have) != KindsSource("history") {
		t.Fatal("kinds_gen.go is stale: run go generate ./internal/history")
	}
	have, err = os.ReadFile("../record/kinds_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(have) != KindsSource("record") {
		t.Fatal("record/kinds_gen.go is stale: run go generate ./internal/history")
	}
}

// TestParams: every event of a run carries the keys its kind declares
// and no others, is placed in the chronicle once, and renders.
func TestParams(t *testing.T) {
	t.Parallel()
	w := reference()
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
		seen[e.Kind] = true
	}
	t.Logf("%d events of %d kinds, %d kinds of %d in the table seen", len(w.Events), len(seen), len(seen), len(kindDefs))
}

// TestLinesUnread: the chronicle's templates and the renderers are the
// view's (internal/legends); nothing in the simulation reads them, so
// the text is free to change. The view imports the record and never the
// history, and the history never imports the view.
func TestLinesUnread(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(src), `"worldgen/internal/`+"legends"+`"`) {
			t.Errorf("%s imports the view", file)
		}
	}
}
