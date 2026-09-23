package main

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"worldgen/internal/legends"
	"worldgen/internal/mind"
	"worldgen/internal/record"
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
	// continuity (specs/proposals/continuity.md)
	Spans     []float64 // each mortal blood's span, before medicine
	Deathless int       // bloods with no turnover
	Software  int       // of them, bloods that uploaded and kept going
	Cont      []float64 // the continuity of each mortal people standing at the present
	StandSoft int       // standing at the present in software
	StandDead int       // standing at the present that never turned over: machines, eldritch things, living worlds
	Heavens   int       // peoples that went into their heaven
	Archives  int       // archives standing at the present
	ArchRuin  int       // archives left as remains
	// leaders (specs/proposals/leaders.md)
	Leaders []*record.Leader
}

func flattenOss(w *world) OssRec {
	r := OssRec{Seed: w.seed(), CivilWars: map[int]int{}, Shatters: map[int]int{}}
	for _, c := range w.State.Civs {
		t := c.Batch.Tally
		r.Faced += t.OssFaced
		r.Renewed += t.OssRenewed
		r.Set += t.OssSet
		r.Broke += t.OssBroke
		r.StiffSum += t.OssStiff
		if legends.Active(c) {
			r.Standing++
			if c.Ossified {
				r.StandOss++
			} else if c.Stiff >= 1 {
				r.StandSet++
			}
		}
	}
	for _, sp := range w.State.Species {
		switch {
		case sp.Lifespan > 0:
			r.Spans = append(r.Spans, float64(sp.Lifespan))
		case sp.Software:
			r.Deathless++
			r.Software++
		default:
			r.Deathless++
		}
	}
	for _, c := range w.State.Civs {
		if legends.Active(c) {
			switch sp := w.State.Species[c.Species]; {
			case sp.Lifespan > 0:
				r.Cont = append(r.Cont, c.Continuity)
			case sp.Software:
				r.StandSoft++
			default:
				r.StandDead++
			}
		}
		if c.Cause == "heaven" {
			r.Heavens++
		}
		r.Archives += c.Structures["archive"]
	}
	for _, l := range w.State.Remains {
		if l.Portrait == "archive" {
			r.ArchRuin++
		}
	}
	r.Leaders = w.State.Leaders
	seen := map[int]bool{}
	for _, f := range w.Chronicle {
		switch f.Kind {
		case record.FSundered:
			if !seen[f.Subject] {
				seen[f.Subject] = true
				r.CivilWars[f.N]++
			}
		case record.FShattered:
			if !seen[f.Subject] {
				seen[f.Subject] = true
				r.Shatters[f.N]++
			}
		case record.FDarkAge:
			r.Depths = append(r.Depths, float64(f.N)/10)
		case record.FReclaimed:
			r.Reclaimed++
		}
	}
	for _, e := range w.Chronicle {
		if e.Kind == record.KKinMet {
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
	var spans, cont []float64
	var deathless, software, heavens, archives, archRuin, standSoft, standDead int
	for _, r := range oss {
		spans = append(spans, r.Spans...)
		cont = append(cont, r.Cont...)
		deathless += r.Deathless
		software += r.Software
		heavens += r.Heavens
		archives += r.Archives
		archRuin += r.ArchRuin
		standSoft += r.StandSoft
		standDead += r.StandDead
	}
	p("")
	p("Continuity, the share of a people's past that reaches across a thousand years (`continuity.md`): the bloods' spans before medicine run %.0f to %.0f years at the quartiles (median %.0f) over %d mortal bloods, beside %d that do not turn over, %d of them in software. Of the standing at the present, %d are mortal and keep %.3f of their past at the median (quartiles %.3f to %.3f), %d are in software, and %d never turned over (machines, eldritch things, living worlds); each of the last two keeps %.3f, the drift alone. %d archives stand at the present and %d are remains. %d peoples went into their heaven and are still there.",
		quantile(spans, 0.25), quantile(spans, 0.75), median(spans), len(spans), deathless, software, len(cont), median(cont), quantile(cont, 0.25), quantile(cont, 0.75), standSoft, standDead, math.Exp(-mind.Default().Continuity.Drift), archives, archRuin, heavens)
	var leaders []*record.Leader
	for _, r := range oss {
		leaders = append(leaders, r.Leaders...)
	}
	p("")
	p("%s", leaderParagraph(leaders, fs))
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
	for _, k := range sortedByCount(causes) {
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

// leaderParagraph is the leaders of the batch (leaders.md): how many rose
// and on what occasion, where they stood, how many turned their people
// against its own bent, how they ended, and the succession after them by
// the continuity of the realm that faced it.
func leaderParagraph(ls []*record.Leader, fs float64) string {
	if len(ls) == 0 {
		return "Leaders (`leaders.md`): none rose."
	}
	by := func(key func(*record.Leader) string, order ...string) string {
		m := map[string]int{}
		for _, l := range ls {
			m[key(l)]++
		}
		var parts []string
		for _, k := range order {
			if m[k] > 0 {
				parts = append(parts, fmt.Sprintf("%s %s", k, pct(m[k], len(ls))))
			}
		}
		return strings.Join(parts, ", ")
	}
	front, turned, deathless, mad := 0, 0, 0, 0
	for _, l := range ls {
		if l.Front {
			front++
		}
		if mind.Hostile(l.Stance) != mind.Hostile(l.Own) {
			turned++
		}
		if l.Deathless {
			deathless++
		}
		if l.Mad >= 1 {
			mad++
		}
	}
	type band struct {
		n, over, dec int
	}
	bands := make([]band, 4) // doublings under 0, 0 to 1, 1 to 2, 2 and over
	faced, over, scar, dec := 0, 0, 0, 0
	for _, l := range ls {
		if l.Faced == "" {
			continue
		}
		faced++
		i := min(3, max(0, int(math.Floor(l.Doublings))+1))
		bands[i].n++
		switch l.Faced {
		case "overcome":
			over++
			bands[i].over++
		case "scarred":
			scar++
		default:
			dec++
			bands[i].dec++
		}
	}
	var bs []string
	for i, name := range []string{"under 0", "0 to 1", "1 to 2", "2 and over"} {
		if bands[i].n > 0 {
			bs = append(bs, fmt.Sprintf("%s doublings %s declined of %d", name, pct(bands[i].dec, bands[i].n), bands[i].n))
		}
	}
	return fmt.Sprintf("Leaders (`leaders.md`): %d rose, %.1f per world; by occasion %s; as %s. %s stood at the front; %s were turned to or from war against their people's own stance. They ended: %s. %d were deathless, and %d of those went mad. Successions faced %d: overcome %s, scarred %s, declined %s; by the realm's continuity, %s.",
		len(ls), float64(len(ls))/fs,
		by(func(l *record.Leader) string { return l.Occasion }, "war", "crisis", "reform", "founding"),
		by(func(l *record.Leader) string { return l.Form }, "person", "line", "directive", "brood", "doctrine"),
		pct(front, len(ls)), pct(turned, len(ls)),
		by(func(l *record.Leader) string {
			if l.End == "" {
				return "reigning at the present"
			}
			return l.End
		}, "died", "fell_field", "fell_capital", "overthrown", "with_people", "deposed", "reigning at the present"),
		deathless, mad, faced, pct(over, faced), pct(scar, faced), pct(dec, faced), strings.Join(bs, ", "))
}
