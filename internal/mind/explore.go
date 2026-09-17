package mind

import (
	"fmt"
	"math"
)

// Exploration policy: how many surveyors to keep out, where the next one
// goes, when a tour ends, and when the Sight is turned outward.

// SurveyInput is a people's want for surveyors.
type SurveyInput struct {
	Free  bool // not held, not aloft
	Reach float64
	Mil   float64
	AtWar bool
	Era   int
	Dials Dials
	// these cost a scan of the stars, so they are asked only when the
	// answer would change the want
	ToSettle  func() bool // a read, free, livable star lies within reach
	Unread    func() bool // an unread or stale star lies within reach
	NeverRead func() bool // a star within reach was never read
}

// SurveyWant is how many surveyors a people wants out.
type SurveyWant struct {
	Want   int
	Reason string
}

// Why says the want.
func (s SurveyWant) Why() string { return fmt.Sprintf("wants %d surveyors out: %s", s.Want, s.Reason) }

// Survey is how many surveyors hunger and greed want and the level can
// spare: none in wartime, at least one when there is nothing read to
// settle and stars still unread, one when only stale stars are left.
func Survey(in SurveyInput, t *Tuning) SurveyWant {
	p := &t.Survey
	if !in.Free || in.Reach < p.MinReach || in.Mil < p.MinMil {
		return SurveyWant{Reason: "no ships to spare"}
	}
	s := SurveyWant{Want: int(math.Round(p.HungerWeight*in.Dials.Hunger + p.GreedWeight*in.Dials.Greed)), Reason: "hunger and greed"}
	if in.AtWar {
		s.Want, s.Reason = 0, "at war"
	}
	if s.Want == 0 && in.Era >= p.Necessity && !in.ToSettle() && in.Unread() {
		s.Want, s.Reason = 1, "nothing read to settle"
	}
	if s.Want > 1 && !in.NeverRead() {
		s.Want, s.Reason = 1, "only stale charts to renew"
	}
	if keep := int(in.Mil - p.KeepHome); s.Want > keep {
		s.Want, s.Reason = keep, "a level must stay home"
	}
	return s
}

// SurveyHop is how far a tour's next star may lie from the last: a hop,
// or the whole reach with the Door.
func SurveyHop(reach float64, ftl bool, t *Tuning) float64 {
	if ftl {
		return reach
	}
	return min(max(reach, t.Survey.HopMin), t.Survey.HopMax)
}

// Star is a star as a survey or a colony ship weighs it, nearest first.
type Star struct {
	ID      int
	Marked  bool // the Sight marked it: something is there
	Charted bool // read before
	Fresh   bool // read lately
}

// SurveyTarget picks the next star for surveyors from candidates in order
// of distance: a marked star first, else the nearest never read, else the
// nearest stale. -1 when nothing is left.
func SurveyTarget(stars []Star) int {
	best, stale := -1, -1
	for _, s := range stars {
		if s.Marked {
			return s.ID
		}
		if !s.Charted {
			if best < 0 {
				best = s.ID
			}
		} else if stale < 0 && !s.Fresh {
			stale = s.ID
		}
	}
	if best < 0 {
		return stale
	}
	return best
}

// TourOn says whether surveyors go on to another star: not recalled, and
// the tour not yet at its length in stars or years.
func TourOn(tour int, years float64, recalled bool, t *Tuning) bool {
	return !recalled && tour < t.Survey.MaxTour && years < t.Survey.MaxTourYears
}

// SightInput is whether the Sight is turned outward.
type SightInput struct {
	AtWar bool
	Fear  float64
	// asked only when the answer would change the mode, since each costs
	// a scan
	Menaced func() bool // a hostile neighbour in reach
	Unread  func() bool // stars within the Sight's range are unread or stale
}

// SightMode is the Sight's mode.
type SightMode struct {
	Outward bool
	Threat  bool // it is on the borders because something threatens
}

// Why says the mode.
func (s SightMode) Why() string {
	switch {
	case s.Outward:
		return "the Sight looks outward"
	case s.Threat:
		return "the Sight watches the borders"
	}
	return "nothing left to look for"
}

// Sight turns the Sight outward when nothing threatens and something is
// unread; a war, or for the fearful a hostile neighbour, turns it back.
func Sight(in SightInput, t *Tuning) SightMode {
	want := !in.AtWar
	if want && in.Fear > t.Sight.FearBar && in.Menaced() {
		want = false
	}
	m := SightMode{Threat: !want}
	if want && !in.Unread() {
		want = false
	}
	m.Outward = want
	return m
}

// Menaces says whether a neighbour counts as hostile to the Sight: it
// strikes first, is held a monster, or is deeply resented.
func Menaces(hostile, monster bool, grudge float64, t *Tuning) bool {
	return hostile || grudge > t.Sight.GrudgeBar || monster
}

// SightRange is how far the Sight reads outward.
func SightRange(reach float64, t *Tuning) float64 {
	return max(t.Sight.RangeMul*reach, t.Sight.RangeMin)
}
