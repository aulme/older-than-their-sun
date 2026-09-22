package main

import (
	"fmt"
	"io"
	"math"

	"worldgen/internal/legends"
)

// DeclineRec is one world's occupancy at the present, as the decline
// index reads it (specs/proposals/decline.md). Everything here is
// carried by the dossier but the reach, which is walked over the state:
// the held share over the reach is the real occupancy, and it separates
// worlds that the held share alone does not — a last empire holding
// everything it can cross reads nothing like a galaxy that can be
// crossed end to end and is a tenth full.
type DeclineRec struct {
	Seed     uint64
	Age      float64 // the age's length, Myr
	Waning   float64 // when the waning was declared, Myr
	Capped   bool
	Index    float64
	AtWaning float64
	Held     float64 // the term: the held share over the age's height
	Rising   float64
	Births   float64
	HeldNow  float64 // and the absolute numbers the terms are shares of
	PeakHeld float64
	HeightAt float64 // when the height was, Myr
	Rise     int
	PeakRise int
	Born     float64
	PeakBorn float64
	Reach    float64 // habitable systems within some living people's reach, of all habitable
	Standing int
	ByIndex  bool // the age ended on the index rather than on the fertility floor
}

func flattenDecline(w *world) DeclineRec {
	d, st := w.Dossier, w.State
	dec := d.Decline
	r := DeclineRec{
		Seed: d.Seed, Age: float64(d.Present-d.Dawn) / 1e6, Waning: float64(d.Waning-d.Dawn) / 1e6, Capped: d.Capped,
		Index: dec.Index, AtWaning: dec.AtWaning, Held: dec.Held, Rising: dec.Rising, Births: dec.Births,
		HeldNow: dec.HeldNow, PeakHeld: dec.PeakHeld, HeightAt: float64(dec.PeakHeldAt-d.Dawn) / 1e6,
		Rise: dec.RisingNow, PeakRise: dec.PeakRising, Born: dec.BirthsNow, PeakBorn: dec.PeakBirths,
		ByIndex: dec.ByIndex,
	}
	// The reach: a star is covered when it lies inside some living
	// people's reach of one of that people's own holdings, not only of
	// its seat. A realm reaches from its whole territory, and reading it
	// from the seat alone puts held worlds outside the reach that holds
	// them, which made the held share come out larger than the reach it
	// is read against.
	hab, covered := 0, 0
	within := make([]bool, len(st.Stars))
	for _, c := range st.Civs {
		if !legends.Active(c) {
			continue
		}
		r.Standing++
		if c.Reach < 1 {
			continue
		}
		for _, h := range c.Systems {
			from := st.Stars[h]
			for i := range st.Stars {
				if within[i] {
					continue
				}
				s := st.Stars[i]
				dx, dy, dz := s.X-from.X, s.Y-from.Y, s.Z-from.Z
				if math.Sqrt(dx*dx+dy*dy+dz*dz) <= c.Reach {
					within[i] = true
				}
			}
		}
	}
	for i := range st.Stars {
		if st.Stars[i].Hab > 0 {
			hab++
			if within[i] {
				covered++
			}
		}
	}
	if hab > 0 {
		r.Reach = float64(covered) / float64(hab)
	}
	return r
}

// declineReport is the occupancy section: where each world stands
// against its own height. The whole trajectory — the still span, the
// churn, where the height fell in the age — is not in a run's records,
// which hold the present; it is measured by the sampler that step 6
// built for it (internal/history/declineshape_test.go, DECLINE=<dir>).
func declineReport(out io.Writer, decs []DeclineRec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Decline")
	p("")
	p("The index is three terms, each against the age's own running peak, and their mean taken from 1: **held** (habitable systems held by living peoples), **rising** (peoples active, not ossified and not yet set) and **births** (peoples born per Myr in the window). 0 is an age at its height, 1 an age with nothing standing. The whole trajectory is the sampler's (`DECLINE=<dir> go test ./internal/history -run TestDeclineShape`); this table is the present.")
	p("")
	p("| Seed | Age Myr | Waning | Index | at waning | held | rising | births | held now | height | when | held/reach | standing | ended by | capped |")
	p("|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|")
	for _, d := range decs {
		over := 0.0
		if d.Reach > 0 {
			over = d.HeldNow / d.Reach
		}
		by := "the floor"
		if d.ByIndex {
			by = "the index"
		}
		p("| %d | %.1f | %.1f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.0f%% | %.2f | %d | %s | %v |",
			d.Seed, d.Age, d.Waning, d.Index, d.AtWaning, d.Held, d.Rising, d.Births,
			d.HeldNow, d.PeakHeld, 100*d.HeightAt/max(d.Age, 0.001), over, d.Standing, by, d.Capped)
	}
	p("")
	var idx, keep, age float64
	late, capped, byIndex := 0, 0, 0
	for _, d := range decs {
		idx += d.Index
		keep += d.Held
		age += d.Age
		if d.HeightAt > 2*d.Age/3 {
			late++
		}
		if d.Capped {
			capped++
		}
		if d.ByIndex {
			byIndex++
		}
	}
	n := float64(len(decs))
	if n == 0 {
		return
	}
	height := fmt.Sprintf("**%d of %d worlds have their height in the last third of the age** — the galaxy at its fullest at the moment the legends call it the waning.", late, len(decs))
	if late == 0 {
		height = fmt.Sprintf("No world of the %d has its height in the last third of its age.", len(decs))
	}
	p("Mean index at the present %.2f; the present holds %.0f%% of the height on average; mean age %.1f Myr. %s %d capped.",
		idx/n, 100*keep/n, age/n, height, capped)
	p("")
	p("**%d of %d ages ended on the index; %d fell through to the fertility floor.** The floor stays until the force lands and takes worlds from the old (`specs/proposals/decline.md` stage 3, carried by `empires.md`); that second count is the number the force drives to zero.",
		byIndex, len(decs), len(decs)-byIndex)
}
