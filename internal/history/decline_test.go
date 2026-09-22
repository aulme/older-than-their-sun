package history

import (
	"testing"
)

// TestDeclineAtTheDawn: an age that has not had a height has not fallen
// from one. At the first tick every term is at its own running peak, so
// the index is 0 — and a galaxy that never fills stays at 0 rather than
// reading as declined from a height it never had.
func TestDeclineAtTheDawn(t *testing.T) {
	w := newTestWorld(t, 3, 60)
	w.tickDecline()
	if w.Decline.Index != 0 {
		t.Errorf("the index at the dawn is %.3f, want 0 (%+v)", w.Decline.Index, w.Decline)
	}
}

// TestDeclineWithNothingLeft: the index is 1 when the age had a height
// and nothing of it is left — nothing held, nothing rising, nothing
// born.
func TestDeclineWithNothingLeft(t *testing.T) {
	w := newTestWorld(t, 3, 60)
	d := &w.Decline
	d.PeakHeld, d.PeakRising, d.PeakBirths = 0.5, 20, 4
	w.bornMark = len(w.Civs) // nobody inside the window
	w.tickDecline()
	if d.HeldNow != 0 || d.RisingNow != 0 || d.BirthsNow != 0 {
		t.Fatalf("the empty galaxy is not empty: %+v", *d)
	}
	if d.Index != 1 {
		t.Errorf("the index with nothing left is %.3f, want 1 (%+v)", d.Index, *d)
	}
}

// TestDeclinePeaksNeverFall: the peaks are running maxima taken as the
// age goes and never reset. A height the age does not return to is the
// point of the measure, so a peak that could fall would hide exactly
// what the index is for.
func TestDeclinePeaksNeverFall(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stars = 120
	cfg.Until = 6_000_000
	var peakHeld, peakBirths float64
	var peakRising int
	var maxIndex float64
	sawHeld := false
	cfg.Sample = func(w *World) {
		d := w.Decline
		if d.PeakHeld < peakHeld || d.PeakRising < peakRising || d.PeakBirths < peakBirths {
			t.Fatalf("a peak fell at year %d: %+v", w.Now, d)
		}
		if d.Index < 0 || d.Index > 1 {
			t.Fatalf("the index is %.3f at year %d, outside 0 to 1", d.Index, w.Now)
		}
		for _, x := range []float64{d.Held, d.Rising, d.Births} {
			if x < 0 || x > 1 {
				t.Fatalf("a term is %.3f at year %d, outside 0 to 1", x, w.Now)
			}
		}
		peakHeld, peakRising, peakBirths = d.PeakHeld, d.PeakRising, d.PeakBirths
		maxIndex = max(maxIndex, d.Index)
		sawHeld = sawHeld || d.HeldNow > 0
	}
	Generate(4, cfg)
	if !sawHeld {
		t.Fatal("nothing was ever held: the test is checking nothing")
	}
	if peakHeld == 0 || peakRising == 0 || peakBirths == 0 {
		t.Errorf("a peak stayed at zero through the age: held %.3f, rising %d, births %.2f", peakHeld, peakRising, peakBirths)
	}
	if maxIndex == 0 {
		t.Error("the index never left zero over six million years")
	}
}

// TestDeclineBarMustHold: one tick over the bar does not declare an age.
// The bar must stand for HoldMyr, and a dip back under starts the count
// again; once it has stood, a good tick does not unsay it.
func TestDeclineBarMustHold(t *testing.T) {
	w := newTestWorld(t, 3, 60)
	bar := w.Cfg.Tuning.Decline.WaningBar
	hold := Year(w.Cfg.Tuning.Decline.HoldMyr * 1e6)
	step := func(index float64, dt Year) Year {
		w.Now += dt
		return w.stood(index, bar, &w.overWaning, w.Decline.Crossed)
	}
	if got := step(bar+0.1, 1000); got != 0 {
		t.Fatalf("the bar was believed on the first tick over it, at %d", got)
	}
	if got := step(bar+0.1, hold/2); got != 0 {
		t.Fatalf("the bar was believed after half the hold, at %d", got)
	}
	if got := step(bar-0.01, 1000); got != 0 || w.overWaning != 0 {
		t.Fatalf("a tick under the bar did not start the count again: %d, since %d", got, w.overWaning)
	}
	began := w.Now + 1000
	if got := step(bar+0.1, 1000); got != 0 {
		t.Fatalf("the count did not start again: %d", got)
	}
	got := step(bar+0.1, hold)
	if got != began {
		t.Fatalf("the bar was crossed at %d, want %d", got, began)
	}
	w.Decline.Crossed = got
	if again := step(0, 1000); again != got {
		t.Errorf("a good tick unsaid the waning: %d, was %d", again, got)
	}
}

// TestDeclineReadsTheMap: the index reads what is held, which is what
// the fertility clock it replaced could not. Take half the held worlds
// away with every people still standing and the index climbs; fertility
// does not move at all, because fertility is about who is born.
func TestDeclineReadsTheMap(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stars = 120
	cfg.Until = 8_000_000
	w := Generate(4, cfg)
	before, fertility := w.Decline.Index, w.fertility()
	held := 0
	for i, o := range w.Owner {
		if o >= 0 && w.Civs[o].Active() && w.G.Stars[i].Hab > 0 {
			held++
		}
	}
	if held < 4 {
		t.Fatalf("only %d habitable worlds held: the test is checking nothing", held)
	}
	// half the worlds go dark, everyone still standing
	gone := 0
	for i, o := range w.Owner {
		if o >= 0 && w.Civs[o].Active() && w.G.Stars[i].Hab > 0 && gone < held/2 {
			w.Owner[i] = -1
			gone++
		}
	}
	w.tickDecline()
	if w.Decline.Index <= before {
		t.Errorf("half the map went dark and the index did not move: %.3f, was %.3f", w.Decline.Index, before)
	}
	if w.fertility() != fertility {
		t.Errorf("fertility moved with the map: %.4f, was %.4f", w.fertility(), fertility)
	}
}
