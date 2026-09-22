package writer

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"worldgen/data"
	"worldgen/internal/history"
	"worldgen/internal/record"
)

func generate(seed uint64, stars int, until history.Year) *history.World {
	cfg := history.DefaultConfig()
	cfg.Stars = stars
	cfg.Until = until
	return history.Generate(seed, cfg)
}

// TestRoundtrip: a run written to a directory and read back is the run.
func TestRoundtrip(t *testing.T) {
	w := generate(7, 120, 0)
	r := Run(w, "sol")
	dir := filepath.Join(t.TempDir(), "7")
	if err := record.Write(dir, r); err != nil {
		t.Fatal(err)
	}
	back, err := record.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(r)
	b, _ := json.Marshal(back)
	if !bytes.Equal(a, b) {
		t.Fatal("the run read back is not the run written")
	}
	if len(r.Chronicle) == 0 || len(r.Tellings) == 0 || len(r.State.Names) == 0 {
		t.Fatalf("an empty run: %d events, %d tales, %d names", len(r.Chronicle), len(r.Tellings), len(r.State.Names))
	}
	for i := 1; i < len(r.Chronicle); i++ {
		if r.Chronicle[i].Year < r.Chronicle[i-1].Year {
			t.Fatalf("the chronicle is out of year order at %d", i)
		}
	}
	for _, l := range r.State.Remains {
		for _, id := range l.Testament {
			if tl := r.Tellings[id]; tl.Remain != l.ID || tl.Frozen == nil {
				t.Fatalf("remain %d's testament points at tale %d, which is %+v", l.ID, id, tl)
			}
		}
	}
}

// TestUntil: a run stopped at its own present is the plain run; one
// stopped earlier is the same history to that year, marked truncated.
func TestUntil(t *testing.T) {
	plain := Run(generate(7, 120, 0), "sol")
	same := Run(generate(7, 120, plain.Dossier.Present), "sol")
	a, _ := json.Marshal(plain)
	b, _ := json.Marshal(same)
	if !bytes.Equal(a, b) {
		t.Fatal("-until at the present is not the plain run")
	}
	until := plain.Dossier.Present / 2
	early := Run(generate(7, 120, until), "sol")
	if !early.Dossier.Truncated || early.Dossier.Present != until {
		t.Fatalf("truncated %v, present %d, want %d", early.Dossier.Truncated, early.Dossier.Present, until)
	}
	var before, sky []*record.Event
	for _, e := range plain.Chronicle {
		if e.Kind == record.KSkyFeature {
			continue // placed from the present, which moved
		}
		if e.Year <= until {
			before = append(before, e)
		}
	}
	for _, e := range early.Chronicle {
		if e.Kind == record.KSkyFeature {
			sky = append(sky, e)
			continue
		}
		if e.Year > until {
			t.Fatalf("an event after the cut: %+v", e)
		}
	}
	got := 0
	for _, e := range early.Chronicle {
		if e.Kind != record.KSkyFeature {
			got++
		}
	}
	if got != len(before) {
		t.Fatalf("%d events to year %d, the plain run has %d", got, until, len(before))
	}
	silent := silentKinds()
	for i, e := range before {
		f := early.Chronicle[i+skyBefore(early.Chronicle, i)]
		// a silent kind's id comes after every told event's, so it moves with the cut
		if (!silent[e.Kind] && e.ID != f.ID) || e.Kind != f.Kind || e.Year != f.Year || e.Subject != f.Subject {
			t.Fatalf("event %d differs: plain %+v, early %+v", i, e, f)
		}
	}
}

// silentKinds reads which kinds are silent from the events table.
func silentKinds() map[record.Kind]bool {
	var f struct {
		Kinds []struct {
			Kind   record.Kind `json:"kind"`
			Silent bool        `json:"silent"`
		} `json:"kinds"`
	}
	data.Load("events.json", &f)
	out := map[record.Kind]bool{}
	for _, k := range f.Kinds {
		if k.Silent {
			out[k.Kind] = true
		}
	}
	return out
}

// skyBefore counts the sky events among the first i+1 records that are
// not sky events, so the two chronicles can be walked together.
func skyBefore(es []*record.Event, i int) int {
	n, seen := 0, 0
	for _, e := range es {
		if e.Kind == record.KSkyFeature {
			n++
			continue
		}
		if seen == i {
			break
		}
		seen++
	}
	return n
}
