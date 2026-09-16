package main

import (
	"fmt"
	"io"
)

// exploreReport reads how the stars were explored: surveyors, the Sight
// turned outward, what was found and how, who met whom by what means.
func exploreReport(out io.Writer, recs []Rec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Exploration")
	p("")
	var reach, surveyed []Rec
	for _, r := range recs {
		if r.Stars {
			reach = append(reach, r)
		}
		if r.Tally.Surveys > 0 {
			surveyed = append(surveyed, r)
		}
	}
	sum := func(rs []Rec, f func(Rec) int) int {
		n := 0
		for _, r := range rs {
			n += f(r)
		}
		return n
	}
	surveys := sum(recs, func(r Rec) int { return r.Tally.Surveys })
	charted := sum(recs, func(r Rec) int { return r.Tally.Charted })
	p("%d peoples of %d that reached the stars ever sent surveyors (%s); %d surveys in all, %.1f per surveying people, %.1f readings per survey. Readings of stars by any means: %d, %.1f per people.",
		len(surveyed), len(reach), pct(len(surveyed), len(reach)), surveys, float64(surveys)/float64(max(1, len(surveyed))), float64(sum(surveyed, func(r Rec) int { return r.Tally.Charted }))/float64(max(1, surveys)), charted, float64(charted)/float64(max(1, len(recs))))
	blind := sum(recs, func(r Rec) int { return r.Tally.Blind })
	lost := sum(recs, func(r Rec) int { return r.Tally.BlindLost })
	p("Colony ships sent on a guess to unread stars: %d, of which %d found nothing to live on.", blind, lost)
	p("")
	p("### Finds, by how")
	p("")
	fs, fst, fc, fo := sum(recs, func(r Rec) int { return r.Tally.FindSurvey }), sum(recs, func(r Rec) int { return r.Tally.FindSettle }), sum(recs, func(r Rec) int { return r.Tally.FindChance }), sum(recs, func(r Rec) int { return r.Tally.FindOwn })
	tot := fs + fst + fc + fo
	p("| How | Finds | Share |")
	p("|---|---|---|")
	p("| surveyors visiting | %d | %s |", fs, pct(fs, tot))
	p("| a ship or fleet arriving, or settling the star | %d | %s |", fst, pct(fst, tot))
	p("| a people's own lost works | %d | %s |", fo, pct(fo, tot))
	p("| bumped into by chance | %d | %s |", fc, pct(fc, tot))
	p("")
	p("### Meetings, by how")
	p("")
	mt, mh, ms, mship := sum(recs, func(r Rec) int { return r.Tally.MetTouch }), sum(recs, func(r Rec) int { return r.Tally.MetHeard }), sum(recs, func(r Rec) int { return r.Tally.MetSurvey }), sum(recs, func(r Rec) int { return r.Tally.MetShip })
	p("Counted once per side. Heard across the dark: %d. Met in the flesh at a border, knowing the other was there: %d. Surveyors coming upon a people: %d. A colony ship arriving to find a people: %d.", mh, mt, ms, mship)
	p("")
	p("### The Sight")
	p("")
	searched, sighted := 0.0, 0.0
	for _, r := range recs {
		searched += r.Tally.Searched
		sighted += r.Tally.Sighted
	}
	if sighted > 0 {
		p("Peoples holding the Sight spent %s of their time with it turned outward, reading stars instead of watching borders (%.0f of %.0f kyr).", pct(int(searched), int(sighted)), searched, sighted)
	} else {
		p("Nobody held the Sight.")
	}
	p("")
	p("### Surveys by posture")
	p("")
	p("| Posture | Peoples that reached the stars | Ever surveyed | Surveys per people | Finds by survey |")
	p("|---|---|---|---|---|")
	byPost := map[string][]Rec{}
	for _, r := range reach {
		byPost[r.Posture] = append(byPost[r.Posture], r)
	}
	for _, k := range sortedKeys(byPost) {
		rs := byPost[k]
		ever := count(rs, func(r Rec) bool { return r.Tally.Surveys > 0 })
		p("| %s | %d | %s | %.1f | %d |", k, len(rs), pct(ever, len(rs)), float64(sum(rs, func(r Rec) int { return r.Tally.Surveys }))/float64(max(1, len(rs))), sum(rs, func(r Rec) int { return r.Tally.FindSurvey }))
	}
}
