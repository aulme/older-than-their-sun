package history

import (
	"math"
	"math/rand/v2"
	"strings"
	"testing"

	"worldgen/internal/species"
)

// TestDeterminism: one seed gives one history. Three worlds from the same
// seed have byte-identical events. Any range over a map on a path that draws
// from the RNG breaks this.
func TestDeterminism(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stars = 200
	var first []Event
	for i := range 3 {
		w := Generate(7, cfg)
		if i == 0 {
			first = w.Events
			continue
		}
		if len(w.Events) != len(first) {
			t.Fatalf("run %d: %d events, first run %d", i, len(w.Events), len(first))
		}
		for j := range first {
			if first[j] != w.Events[j] {
				t.Fatalf("run %d: event %d differs:\n  %d %s\n  %d %s", i, j, first[j].Year, first[j].Text, w.Events[j].Year, w.Events[j].Text)
			}
		}
	}
}

// TestOneStep: the age runs at Step from the dawn to the present, and the
// waning is declared once and logged once.
func TestOneStep(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stars = 200
	w := Generate(3, cfg)
	span := w.Present - cfg.Dawn
	if span%cfg.Step != 0 {
		t.Errorf("present %d is not a whole number of steps of %d from the dawn", w.Present, cfg.Step)
	}
	if want := int(span/cfg.Step) + 1; w.Ticks != want {
		t.Errorf("ran %d ticks over %d years at step %d; want %d", w.Ticks, span, cfg.Step, want)
	}
	if w.Waning < cfg.Dawn || w.Waning > w.Present {
		t.Errorf("waning at %d outside the age %d..%d", w.Waning, cfg.Dawn, w.Present)
	}
	n := 0
	for _, e := range w.Events {
		if strings.HasPrefix(e.Text, "The age is waning.") {
			n++
			if e.Year != w.Waning {
				t.Errorf("waning logged at %d, set at %d", e.Year, w.Waning)
			}
		}
	}
	if n != 1 {
		t.Errorf("the waning was logged %d times", n)
	}
}

// TestChanceRate: a rate given per thousand years fires its expected number
// of times over a span, through both chance and count.
func TestChanceRate(t *testing.T) {
	w := &World{R: rand.New(rand.NewPCG(11, 22)), dt: 1}
	const ticks = 200_000
	for _, p := range []float64{0.001, 0.01, 0.2} {
		hits := 0
		for range ticks {
			if w.chance(p) {
				hits++
			}
		}
		want := p * ticks
		// four standard deviations of a binomial
		tol := 4 * math.Sqrt(want*(1-p))
		if math.Abs(float64(hits)-want) > tol {
			t.Errorf("chance(%g): %d hits over %d ticks, want %.0f ± %.0f", p, hits, ticks, want, tol)
		}
	}
	for _, rate := range []float64{0.3, 1.5} {
		sum := 0
		for range ticks {
			sum += w.count(rate)
		}
		want := rate * ticks
		frac := rate - math.Floor(rate)
		tol := 4 * math.Sqrt(ticks*frac*(1-frac))
		if math.Abs(float64(sum)-want) > tol+1 {
			t.Errorf("count(%g): %d over %d ticks, want %.0f ± %.0f", rate, sum, ticks, want, tol)
		}
	}
}

// TestHarness: a known people raised on a small world lives through a few
// ticks of the pipeline and the world stays consistent.
func TestHarness(t *testing.T) {
	w := newTestWorld(t, 6, 60)
	c := spawnAt(w, 0, species.Fixed("cooperative"))
	if c.Home != 0 || w.Owner[0] != c.ID || !c.Has("cooperative") {
		t.Fatalf("spawnAt did not raise the people at star 0: home %d owner %d", c.Home, w.Owner[0])
	}
	w.ticks(50)
	if w.Ticks != 50 || w.Now != w.Cfg.Dawn+50*w.Cfg.Step {
		t.Errorf("after 50 ticks: Ticks %d, Now %d", w.Ticks, w.Now)
	}
	if !c.Living() {
		t.Errorf("the people died within 50 kyr: %q", c.Cause)
	}
}
