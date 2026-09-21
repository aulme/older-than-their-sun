package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"worldgen/internal/history"
)

// Ossification: how the old fall now that nobody dies of age. Facings of
// the filter and their outcomes, the stiffness at the break, civil wars
// by heir count, dark ages by depth, shatterings by shard count, and the
// share of peoples ended by cause, so the "died of nothing" count can be
// watched at zero.

// OssRec is one world's ossification.
type OssRec struct {
	Seed      uint64
	Faced     int
	Renewed   int
	Set       int
	Broke     int
	StiffSum  float64     // stiffness summed over facings
	CivilWars map[int]int // by heir count
	Depths    []float64   // each dark age's depth
	Shatters  map[int]int // by shard count
	Reclaimed int
	KinMeets  int
	Standing  int // active at the present
	StandOss  int // of them ossified
	StandSet  int // of them at stiffness one or more
}

func flattenOss(w *history.World) OssRec {
	r := OssRec{Seed: w.Seed, CivilWars: map[int]int{}, Shatters: map[int]int{}}
	for _, c := range w.Civs {
		r.Faced += c.Tally.OssFaced
		r.Renewed += c.Tally.OssRenewed
		r.Set += c.Tally.OssSet
		r.Broke += c.Tally.OssBroke
		r.StiffSum += c.Tally.OssStiff
		if c.Active() {
			r.Standing++
			if c.Ossified {
				r.StandOss++
			} else if c.Stiff >= 1 {
				r.StandSet++
			}
		}
	}
	seen := map[int]bool{}
	for _, f := range w.Events {
		switch f.Kind {
		case history.FSundered:
			if !seen[f.Subject] {
				seen[f.Subject] = true
				r.CivilWars[f.N]++
			}
		case history.FShattered:
			if !seen[f.Subject] {
				seen[f.Subject] = true
				r.Shatters[f.N]++
			}
		case history.FDarkAge:
			r.Depths = append(r.Depths, float64(f.N)/10)
		case history.FReclaimed:
			r.Reclaimed++
		}
	}
	for _, e := range w.Events {
		if e.Kind == history.KKinMet {
			r.KinMeets++
		}
	}
	return r
}

func ossReport(out io.Writer, oss []OssRec, recs []Rec, seeds int) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	fs := float64(seeds)
	var faced, renewed, set, broke, reclaimed, kin, standing, standOss, standSet int
	var stiff float64
	civil := map[int]int{}
	shatter := map[int]int{}
	var depths []float64
	for _, r := range oss {
		faced += r.Faced
		renewed += r.Renewed
		set += r.Set
		broke += r.Broke
		stiff += r.StiffSum
		reclaimed += r.Reclaimed
		kin += r.KinMeets
		standing += r.Standing
		standOss += r.StandOss
		standSet += r.StandSet
		for k, v := range r.CivilWars {
			civil[k] += v
		}
		for k, v := range r.Shatters {
			shatter[k] += v
		}
		depths = append(depths, r.Depths...)
	}
	p("")
	p("## Ossification")
	p("")
	p("Nobody dies of age. A people's ways set (stiffness), and past one it faces Ossification: overcome is a renaissance, the near miss sets it, the bad miss breaks it into a civil war or a dark age of variable depth, which shatters it if the forgetting takes the stars. Heirs and shards are new peoples of the old line.")
	p("")
	meanStiff := 0.0
	if faced > 0 {
		meanStiff = stiff / float64(faced)
	}
	p("Facings %.1f per world at mean stiffness %.2f: renaissance %s, set %s, the break %s.", float64(faced)/fs, meanStiff, pct(renewed, faced), pct(set, faced), pct(broke, faced))
	p("Civil wars %.1f per world (%s); shatterings %.1f per world (%s); dark ages %.1f per world at median depth %.2f (quartiles %.2f to %.2f); worlds reclaimed %.1f per world; kin meeting again %.1f per world.",
		float64(sum(civil))/fs, byCount(civil, "heirs"), float64(sum(shatter))/fs, byCount(shatter, "shards"), float64(len(depths))/fs, median(depths), quantile(depths, 0.25), quantile(depths, 0.75), float64(reclaimed)/fs, float64(kin)/fs)
	p("Standing at the present %.1f per world: %s ossified, %s set in their ways, the rest still rising.", float64(standing)/fs, pct(standOss, standing), pct(standSet, standing))
	lines := 0
	for _, r := range recs {
		if r.Line > 0 {
			lines++
		}
	}
	p("Peoples of a line: %s of all peoples.", pct(lines, len(recs)))
	p("")
	p("How peoples ended, by cause (the ended only):")
	p("")
	p("| Cause | Peoples | Share |")
	p("|---|---|---|")
	causes := map[string]int{}
	ended := 0
	for _, r := range recs {
		if r.Standing || r.Fate == "active" {
			continue
		}
		ended++
		k := killer(r.Cause)
		if r.Fate == "sundered" || r.Fate == "shattered" {
			k = r.Fate
		}
		causes[k]++
	}
	var ks []string
	for k := range causes {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool { return causes[ks[i]] > causes[ks[j]] })
	for _, k := range ks {
		p("| %s | %d | %s |", k, causes[k], pct(causes[k], ended))
	}
}

func sum(m map[int]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}

func byCount(m map[int]int, word string) string {
	var ks []int
	for k := range m {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	var parts []string
	for _, k := range ks {
		parts = append(parts, fmt.Sprintf("%d in %d %s", m[k], k, word))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}
