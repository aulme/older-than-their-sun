package main

import (
	"fmt"
	"io"
	"sort"
)

// loreReport reads how history was remembered: how much each kind of
// people held at the end and how much of it was myth or gone, where the
// tales came from, how often they were retold or the blame moved, who
// was remembered as a monster, and what the telling did to temperament.
func loreReport(out io.Writer, recs []Rec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Tellings")
	p("")
	sum := func(rs []Rec, f func(Rec) int) int {
		n := 0
		for _, r := range rs {
			n += f(r)
		}
		return n
	}
	mean := func(rs []Rec, f func(Rec) float64) float64 {
		if len(rs) == 0 {
			return 0
		}
		x := 0.0
		for _, r := range rs {
			x += f(r)
		}
		return x / float64(len(rs))
	}
	var old []Rec
	for _, r := range recs {
		if r.Lived >= 1 {
			old = append(old, r)
		}
	}
	tales := sum(recs, func(r Rec) int { return r.Tally.Tales })
	p("%d tales learned in all, %.1f per people. By source: witnessed %s, told %s, read in ruins and relics %s, handed down %s.",
		tales, float64(tales)/float64(max(1, len(recs))),
		pct(sum(recs, func(r Rec) int { return r.Tally.Witnessed }), tales),
		pct(sum(recs, func(r Rec) int { return r.Tally.Told }), tales),
		pct(sum(recs, func(r Rec) int { return r.Tally.Read }), tales),
		pct(sum(recs, func(r Rec) int { return r.Tally.Inherited }), tales))
	p("Retold because a regard changed: %d. Blame moved to the enemy of the day: %d. Tellings left in remains: %d; tales restored from a people's own relic: %d.",
		sum(recs, func(r Rec) int { return r.Tally.Revised }), sum(recs, func(r Rec) int { return r.Tally.Blamed }),
		sum(recs, func(r Rec) int { return r.Tally.Testaments }), sum(recs, func(r Rec) int { return r.Tally.Restored }))
	p("")
	p("### Memory by kind, peoples that lived a million years or more")
	p("")
	p("| kind | peoples | held at the end | of them myth | forgotten over the life | monsters remembered |")
	p("|---|---|---|---|---|---|")
	kinds := map[string][]Rec{}
	for _, r := range old {
		kinds[r.kind()] = append(kinds[r.kind()], r)
	}
	var ks []string
	for k := range kinds {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		rs := kinds[k]
		held := sum(rs, func(r Rec) int { return r.Held })
		p("| %s | %d | %.1f | %s | %.1f | %.2f |", k, len(rs), float64(held)/float64(len(rs)), pct(sum(rs, func(r Rec) int { return r.Myth }), held),
			float64(sum(rs, func(r Rec) int { return r.Tally.Forgot }))/float64(len(rs)), mean(rs, func(r Rec) float64 { return float64(r.Monsters) }))
	}
	p("")
	p("### What the telling did to temperament")
	p("")
	p("Mean shift of each dial from the tales held at the end, over all peoples; the cap is 0.3 either way.")
	p("")
	p("| dial | mean shift | mean size of shift |")
	p("|---|---|---|")
	dials := []struct {
		name string
		f    func(Rec) float64
	}{
		{"aggression", func(r Rec) float64 { return r.LoreDials.Aggression }},
		{"risk", func(r Rec) float64 { return r.LoreDials.Risk }},
		{"greed", func(r Rec) float64 { return r.LoreDials.Greed }},
		{"fear", func(r Rec) float64 { return r.LoreDials.Fear }},
		{"loyalty", func(r Rec) float64 { return r.LoreDials.Loyalty }},
		{"hunger", func(r Rec) float64 { return r.LoreDials.Hunger }},
		{"patience", func(r Rec) float64 { return r.LoreDials.Patience }},
		{"hate", func(r Rec) float64 { return r.LoreDials.Hate }},
	}
	for _, d := range dials {
		p("| %s | %+.3f | %.3f |", d.name, mean(recs, d.f), mean(recs, func(r Rec) float64 {
			x := d.f(r)
			if x < 0 {
				x = -x
			}
			return x
		}))
	}
	with := 0
	for _, r := range recs {
		if r.Monsters > 0 {
			with++
		}
	}
	p("")
	p("Peoples that held somebody to be a monster at the end: %d of %d (%s).", with, len(recs), pct(with, len(recs)))
	moralityReport(out, recs)
}

// moralityReport: peoples by what they count as wrong, the monsters each
// morality holds, and how often the same fact is a crime to one people
// and a deed to another.
func moralityReport(out io.Writer, recs []Rec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("### Morality")
	p("")
	p("What each people counts as wrong, rolled at birth from its nature. Monsters are the peoples it remembered as things that do harm at the end, over peoples that lived a million years or more; excused is tales held whose fact is a crime and the holder's judgment is not, condemned the reverse.")
	p("")
	p("| Morality | Peoples | Share | Lived 1 Myr+ | Monsters per people | Held a monster | Excused per people | Condemned per people |")
	p("|---|---|---|---|---|---|---|---|")
	by := map[string][]Rec{}
	for _, r := range recs {
		by[r.Morality] = append(by[r.Morality], r)
	}
	for _, k := range sortedKeys(by) {
		rs := by[k]
		var old []Rec
		for _, r := range rs {
			if r.Lived >= 1 {
				old = append(old, r)
			}
		}
		monsters, with, excused, condemned := 0, 0, 0, 0
		for _, r := range old {
			monsters += r.Monsters
			if r.Monsters > 0 {
				with++
			}
		}
		for _, r := range rs {
			excused += r.Excused
			condemned += r.Condemned
		}
		n := float64(max(1, len(old)))
		p("| %s | %d | %s | %d | %.2f | %s | %.1f | %.1f |", k, len(rs), pct(len(rs), len(recs)), len(old), float64(monsters)/n, pct(with, len(old)),
			float64(excused)/float64(len(rs)), float64(condemned)/float64(len(rs)))
	}
	split, shared := 0, 0
	for _, r := range recs {
		split += r.Split
		shared += r.Shared
	}
	p("")
	p("Crimes known to two peoples or more: %d; of them a crime to one and a deed to another: %d (%s).", shared, split, pct(split, shared))
}
