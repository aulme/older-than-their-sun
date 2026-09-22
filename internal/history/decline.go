package history

// The decline index: what the waning and the end of an age are made of.
// See specs/proposals/decline.md.
//
// The clock it replaces read fertility, which governs who is born and
// says nothing about who is standing or what they hold; the legends'
// line about the waning ("few still rise, and those that stand are old")
// was true of nothing the sim could see. The index reads three things a
// player could see in the aftermath, each against the age's own running
// peak, because an age that never had a height has not fallen from one.
//
// It draws nothing, so a run that computes it is the run that does not.

// Decline is the index, its terms, the age's running peaks and the
// absolute numbers the terms are shares of. The peaks are running
// maxima, taken as the age goes and never reset: a height the age does
// not return to is exactly the point.
type Decline struct {
	Index  float64 // 0 at the height, 1 with nothing standing; the mean of the three
	Held   float64 // habitable systems held, over the highest that share has been
	Rising float64 // peoples still rising, over the highest that count has been
	Births float64 // peoples born per Myr in the window, over the highest that rate has been

	// what the terms are shares of, for a player comparing galaxies:
	// the relative reading hides whether a galaxy was ever full
	HeldNow   float64 // habitable systems held, of all habitable systems
	RisingNow int     // peoples active, not ossified, not yet set
	BirthsNow float64 // peoples born per million years, over the window

	PeakHeld   float64
	PeakRising int
	PeakBirths float64
	PeakHeldAt Year // when the held share was highest: the age's height

	AtWaning float64 // the index at the moment the waning was declared
	Crossed  Year    // when the index first stood at the waning bar; 0 for never
	Fell     Year    // and at the end bar
	ByIndex  bool    // the age ended on the index rather than falling through to the fertility floor
}

// tickDecline is the decline phase: after the legacies, before the
// hazard, so the hazard of the tick reads the index of the tick.
func (w *World) tickDecline() {
	d := &w.Decline
	// held and rising in one walk each: the systems by their holder, the
	// peoples by what they are
	habheld := 0
	for i, o := range w.Owner {
		if o >= 0 && w.Civs[o].Active() && w.G.Stars[i].Hab > 0 {
			habheld++
		}
	}
	rising := 0
	for _, c := range w.Civs {
		if c.Active() && !c.Ossified && c.Stiff < 1 {
			rising++
		}
	}
	d.HeldNow = 0
	if w.habitable > 0 {
		d.HeldNow = float64(habheld) / float64(w.habitable)
	}
	d.RisingNow = rising
	d.BirthsNow = w.birthRate()

	if d.HeldNow > d.PeakHeld {
		d.PeakHeld, d.PeakHeldAt = d.HeldNow, w.Now
	}
	d.PeakRising = max(d.PeakRising, d.RisingNow)
	d.PeakBirths = max(d.PeakBirths, d.BirthsNow)

	d.Held = share(d.HeldNow, d.PeakHeld)
	d.Rising = share(float64(d.RisingNow), float64(d.PeakRising))
	d.Births = share(d.BirthsNow, d.PeakBirths)
	d.Index = clamp(1-(d.Held+d.Rising+d.Births)/3, 0, 1)

	// a bar is believed only once it has held: one bad tick does not
	// declare an age, and an age that dips back under has not waned
	t := w.Cfg.Tuning.Decline
	d.Crossed = w.stood(d.Index, t.WaningBar, &w.overWaning, d.Crossed)
	d.Fell = w.stood(d.Index, t.EndBar, &w.overEnd, d.Fell)
}

// stood keeps the run of ticks the index has been at or over a bar and
// returns the year the bar was first crossed once it has stood there for
// HoldMyr, or 0 while it has not. Once set it stands: what the age has
// been is not unsaid by a good tick.
func (w *World) stood(index, bar float64, since *Year, was Year) Year {
	if was != 0 {
		return was
	}
	if index < bar {
		*since = 0
		return 0
	}
	if *since == 0 {
		*since = w.Now
	}
	if w.Now-*since >= Year(w.Cfg.Tuning.Decline.HoldMyr*1e6) {
		return *since
	}
	return 0
}

// share is a term against the age's running peak. With no peak yet
// there has been no height, so nothing has fallen from one: the term is
// 1 and the index it feeds is 0.
func share(now, peak float64) float64 {
	if peak <= 0 {
		return 1
	}
	return clamp(now/peak, 0, 1)
}

// birthRate is the peoples born inside the window, per million years.
// The peoples of the age are appended as they are born, so from the dawn
// on w.Civs is in birth order and the window's first is found by walking
// a mark forward, never by scanning the list. The myth's sleepers sit
// before them all, at years the window never reaches.
func (w *World) birthRate() float64 {
	win := Year(w.Cfg.Tuning.Decline.BirthWindow * 1e6)
	if win <= 0 {
		return 0
	}
	from := w.Now - win
	for w.bornMark < len(w.Civs) && w.Civs[w.bornMark].Born < from {
		w.bornMark++
	}
	return float64(len(w.Civs)-w.bornMark) / (float64(win) / 1e6)
}
