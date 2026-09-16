package history

import "math"

// The cycle. Every so often the galaxy becomes fertile for minds: a sudden,
// unexplained surge in the chance that a complex biosphere produces a
// spacefaring species. From the moment of the surge the chance decays, a
// little every tick, until it is near zero. Cosmic events only sweep up what
// the fading leaves. Then, a very long time later, it happens again.
//
// Nothing in the simulation knows why. A few very advanced species learn
// that it happens, and where in the turn they stand.

// Cycle is the shape of the galaxy's fertility over deep time.
type Cycle struct {
	Period Year    // time from one surge to the next
	Fade   Year    // e-folding time of the decay after a surge
	Surges []Year  // every surge in the history, oldest first; the last is the current age's
	Floor  float64 // fertility below which an age is said to have ended
}

// fertility is the current multiplier on the chance of a new spacefaring
// species arising: 1 at a surge, decaying toward zero.
func (w *World) fertility() float64 {
	return w.fertilityAt(w.Now)
}

func (w *World) fertilityAt(y Year) float64 {
	var last Year
	found := false
	for _, s := range w.Cycle.Surges {
		if s <= y {
			last, found = s, true
		}
	}
	if !found {
		return 0
	}
	return math.Exp(-float64(y-last) / float64(w.Cycle.Fade))
}

// NextSurge is when the galaxy will wake again, after the present.
func (w *World) NextSurge() Year {
	last := w.Cycle.Surges[len(w.Cycle.Surges)-1]
	return last + w.Cycle.Period
}

// ageEnd is when an age that surged at s is counted as over.
func (w *World) ageEnd(s Year) Year {
	return s + Year(float64(w.Cycle.Fade)*math.Log(1/w.Cycle.Floor))
}

// makeCycle chooses the period and the fade, and places the surges so that
// the current age's surge falls at MidStart.
func (w *World) makeCycle() {
	c := &Cycle{
		Period: Year(9e8 + w.R.Float64()*9e8),
		Fade:   Year(1.6e7 + w.R.Float64()*1.4e7),
		Floor:  0.02,
	}
	// walk back from the current surge, with a little jitter each turn
	var surges []Year
	y := w.Cfg.MidStart
	for y > w.Cfg.DeepStart+Year(2e8) {
		surges = append([]Year{y}, surges...)
		y -= Year(float64(c.Period) * (0.9 + w.R.Float64()*0.2))
	}
	c.Surges = surges
	w.Cycle = c
}

// FertilityNow is the fertility at the present, for the legends.
func (w *World) FertilityNow() float64 { return w.fertilityAt(0) }
