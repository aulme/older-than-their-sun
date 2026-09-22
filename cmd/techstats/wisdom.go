package main

import (
	"fmt"
	"io"
	"slices"
	"sort"

	"worldgen/internal/legends"
	"worldgen/internal/record"
)

// Wisdom: where the level comes from, who came to understand whom and
// how long it took, and how far the councils' acts fell from their
// estimates.

// PairRec is one people's understanding of another it met: reached or
// not, how, and after how long.
type PairRec struct {
	Seed      uint64
	Who, Whom int
	Diff      float64 // how alien the two are
	Fathomed  bool
	How       string
	Years     float64 // from the meeting to the understanding, or to the end
	Mutual    bool    // the reverse was reached too
	Gap       float64 // years between the first understanding of the pair and the second, on the second
}

func flattenPairs(w *world) []PairRec {
	type key struct{ a, b int }
	when := map[key]record.Fathoming{}
	for _, f := range w.State.Fathomings {
		when[key{f.Who, f.Whom}] = f
	}
	var out []PairRec
	for _, c := range w.State.Civs {
		k := c.Knowledge
		ids := make([]int, 0, len(k.FathomTried))
		for eid := range k.FathomTried {
			ids = append(ids, eid)
		}
		sort.Ints(ids)
		for _, eid := range ids {
			since, e := k.FathomTried[eid], w.civ(eid)
			fathomed := slices.Contains(k.Fathomed, eid)
			r := PairRec{Seed: w.seed(), Who: c.ID, Whom: eid, Diff: k.Alien[eid], Fathomed: fathomed, Mutual: fathomed && slices.Contains(e.Knowledge.Fathomed, c.ID)}
			end := w.Dossier.Present
			if !legends.Living(c) {
				end = c.Fell
			}
			if f, ok := when[key{c.ID, eid}]; ok && r.Fathomed {
				r.How, r.Years = f.How, float64(f.Year-f.Since)
				if g, ok := when[key{eid, c.ID}]; ok && g.Year <= f.Year && r.Mutual {
					r.Gap = float64(f.Year - g.Year)
				}
			} else {
				r.Years = float64(end - since)
			}
			out = append(out, r)
		}
	}
	return out
}

// wisdomReport reads the level and the fathoming. The calibration
// targets are a share of pairs that never fathom each other small but
// not zero, the lifetime spread unmoved (Wisdom reads into no filter),
// and trade opening later on average.
func wisdomReport(out io.Writer, recs []Rec, pairs []PairRec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Wisdom")
	p("")
	p("A fourth level: the capacity to see the counterintuitive. It decides whether a people fathoms another and pulls its acts toward its own estimate, and makes nothing else better. Peoples that reached the stars, by the most Wisdom they ever had:")
	p("")
	p("| Wisdom | Peoples |")
	p("|---|---|")
	bands := []struct {
		name   string
		lo, hi float64
	}{{"foolish (2 or less)", -1, 2.0001}, {"2 to 4", 2.0001, 4.0001}, {"4 to 6", 4.0001, 6.0001}, {"6 to 8", 6.0001, 8}, {"wise (8 or more)", 8, 11}}
	n := 0
	var from [5]float64
	for _, r := range recs {
		if !r.Stars {
			continue
		}
		n++
		for i := range from {
			from[i] += r.WisFrom[i]
		}
	}
	for _, b := range bands {
		k := 0
		for _, r := range recs {
			if r.Stars && r.PeakWis >= b.lo && r.PeakWis < b.hi {
				k++
			}
		}
		p("| %s | %s |", b.name, pct(k, n))
	}
	p("")
	if n > 0 {
		p("Sources, mean per starfaring people: species %.2f, tech %.2f, experience %.2f, boons %.2f, scars %.2f.", from[0]/float64(n), from[1]/float64(n), from[2]/float64(n), from[3]/float64(n), from[4]/float64(n))
	}
	p("")
	p("| Understood | Contacts | Mean years from the meeting |")
	p("|---|---|---|")
	hows := []string{"meeting", "kin", "chorus", "familiarity", "war", "taught", "broker"}
	for _, h := range hows {
		k := 0
		var ys []float64
		for _, q := range pairs {
			if q.Fathomed && q.How == h {
				k++
				ys = append(ys, q.Years)
			}
		}
		p("| by %s | %s | %.0f |", h, pct(k, len(pairs)), mean(ys))
	}
	never := 0
	var ys []float64
	for _, q := range pairs {
		if !q.Fathomed {
			never++
			ys = append(ys, q.Years)
		}
	}
	p("| never | %s | %.0f (to the end) |", pct(never, len(pairs)), mean(ys))
	p("")
	p("| Difference | Contacts | Never understood | Mean years unfathomed, of those that were |")
	p("|---|---|---|---|")
	diffs := []struct {
		name   string
		lo, hi float64
	}{{"kin (0)", -1, 0.0001}, {"0.5 to 2", 0.0001, 2.0001}, {"2.5 to 4", 2.0001, 4.0001}, {"over 4", 4.0001, 99}}
	for _, d := range diffs {
		k, nv := 0, 0
		var ys []float64
		for _, q := range pairs {
			if q.Diff < d.lo || q.Diff >= d.hi {
				continue
			}
			k++
			if !q.Fathomed {
				nv++
			} else {
				ys = append(ys, q.Years)
			}
		}
		p("| %s | %d | %s | %.0f |", d.name, k, pct(nv, k), mean(ys))
	}
	p("")
	oneSided, mutualPairs := 0, 0
	var gaps []float64
	for _, q := range pairs {
		if q.Fathomed && !q.Mutual {
			oneSided++
		}
		if q.Mutual && q.Gap > 0 {
			mutualPairs++
			gaps = append(gaps, q.Gap)
		}
	}
	mis, brokered, leaps, dropped := 0, 0, 0, 0
	for _, r := range recs {
		mis += r.Tally.Misunderstood
		brokered += r.Tally.Brokered
		leaps += r.Tally.Leaps
		dropped += r.Tally.Dropped
	}
	p("One-sided at the end %d; pairs that went from one-sided to mutual %d, median years between %.0f. Wars that ended with a side not understood %d; brokered attempts made for nothing %d; remains sealed by looking before the leap %d; messages dropped unread %d.", oneSided, mutualPairs, median(gaps), mis/2, brokered, leaps, dropped)
	p("")
	p("| Wisdom | Councils | Acted-on odds' mean distance from the estimate |")
	p("|---|---|---|")
	for _, b := range bands {
		judged, gap := 0, 0.0
		for _, r := range recs {
			if r.PeakWis >= b.lo && r.PeakWis < b.hi {
				judged += r.Tally.Judged
				gap += r.Tally.ActedGap
			}
		}
		if judged > 0 {
			p("| %s | %d | %.3f |", b.name, judged, gap/float64(judged))
		} else {
			p("| %s | 0 | - |", b.name)
		}
	}
}
