package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"worldgen/internal/history"
	"worldgen/internal/species"
)

// Kinds: substrate by modifier, with counts born, made and living at the
// end; the eldritch pool (powers held at birth and gained, deepenings);
// the living worlds' demands and wakings; the unmaking; the tithe; the
// hive's severings; and how each nature fares in war, so an eldritch
// thing can be read as neither trivial nor unbeatable.

// KindRec is one world's kinds.
type KindRec struct {
	Seed     uint64
	Powers   map[string]int // powers held at the end, by key, over every eldritch people
	Born     map[string]int // powers had at birth
	Deepened int
	Appeared int
	Demands  int
	Left     int // demands heeded
	Wakings  int
	Unmade   int
	Tithed   int // pairs
	Severed  int
	Sleeps   int
	Asleep   int // asleep at the present
}

func flattenKinds(w *history.World) KindRec {
	r := KindRec{Seed: w.Seed, Powers: map[string]int{}, Born: map[string]int{}}
	for _, c := range w.Civs {
		for _, k := range c.Species.Powers {
			r.Powers[k]++
		}
		r.Deepened += c.Tally.Deepened
		r.Appeared += c.Tally.Appeared
		r.Demands += c.Tally.Demands
		r.Wakings += c.Tally.Wakings
		r.Unmade += c.Tally.Unmade
		r.Tithed += c.Tally.Tithed
		r.Sleeps += c.Tally.Sleeps
		if c.Active() && c.Asleep {
			r.Asleep++
		}
	}
	gained := map[string]int{} // by the power's name
	for _, f := range w.Facts {
		switch f.Kind {
		case history.FDemand:
			if f.What == "left" {
				r.Left++
			}
		case history.FSevered:
			r.Severed++
		case history.FDeepened:
			gained[f.What]++
		}
	}
	for _, pw := range species.Pool {
		r.Born[pw.Key] = r.Powers[pw.Key] - gained[pw.Name] // what is held and was not gained was there from birth
	}
	return r
}

func kindsReport(out io.Writer, ks []KindRec, recs []Rec, seeds int) {
	p := func(format string, a ...any) { fmt.Fprintf(out, format+"\n", a...) }
	p("")
	p("## Kinds")
	p("")
	p("Substrate by modifier: peoples born at a cradle, made (by a maker, a plague that woke, a sundering), and living at the present. A swarm is a biological people with the swarming trait, counted with its nature.")
	p("")
	type cell struct{ born, made, living int }
	table := map[string]map[string]*cell{}
	subs := map[string]bool{}
	mods := map[string]bool{}
	get := func(sub, mod string) *cell {
		if table[sub] == nil {
			table[sub] = map[string]*cell{}
		}
		if table[sub][mod] == nil {
			table[sub][mod] = &cell{}
		}
		return table[sub][mod]
	}
	for _, r := range recs {
		mod := "none"
		if len(r.Mods) > 0 {
			mod = strings.Join(r.Mods, "+")
		}
		subs[r.Sub], mods[mod] = true, true
		c := get(r.Sub, mod)
		if r.Made != "" || r.Origin != "" {
			c.made++
		} else {
			c.born++
		}
		if r.Standing {
			c.living++
		}
	}
	subList, modList := sortedKeysOf(subs), sortedKeysOf(mods)
	p("| Nature | %s |", strings.Join(subList, " | "))
	p("|---|%s", strings.Repeat("---|", len(subList)))
	for _, m := range modList {
		var cells []string
		for _, s := range subList {
			c := table[s][m]
			if c == nil {
				cells = append(cells, "")
				continue
			}
			cells = append(cells, fmt.Sprintf("%d born, %d made, %d living", c.born, c.made, c.living))
		}
		p("| %s | %s |", m, strings.Join(cells, " | "))
	}
	p("")
	// the pool
	powers := map[string]int{}
	born := map[string]int{}
	deep, appeared, demands, left, wakings, unmade, tithed, severed, sleeps, asleep := 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
	for _, k := range ks {
		for key, n := range k.Powers {
			powers[key] += n
		}
		for key, n := range k.Born {
			born[key] += n
		}
		deep += k.Deepened
		appeared += k.Appeared
		demands += k.Demands
		left += k.Left
		wakings += k.Wakings
		unmade += k.Unmade
		tithed += k.Tithed
		severed += k.Severed
		sleeps += k.Sleeps
		asleep += k.Asleep
	}
	eld := count(recs, func(r Rec) bool { return r.Sub == "eldritch" })
	p("The eldritch pool: %d eldritch peoples over %d worlds, %d deepenings (%.1f per eldritch people), %d second presences appeared.", eld, seeds, deep, float64(deep)/float64(max(eld, 1)), appeared)
	p("")
	p("| Power | Held at the end | Of them from birth |")
	p("|---|---|---|")
	for _, pw := range species.Pool {
		p("| %s | %d | %d |", pw.Name, powers[pw.Key], max(0, born[pw.Key]))
	}
	p("")
	p("Living worlds: %d demands made, %d heeded (%s), %d wakings. The unmaking turned on a world %d times. The tithe fell on %d peoples. Hive worlds severed from their seat: %d. Sleeps: %d, asleep at the present: %d.",
		demands, left, pct(left, demands), wakings, unmade, tithed, severed, sleeps, asleep)
	p("")
	// the natures in war
	p("| Nature | Peoples | Wars fought | Won per war | Lost per war | Ended by war | Median life (Myr) |")
	p("|---|---|---|---|---|---|---|")
	natures := newCounter()
	for _, r := range recs {
		natures.add(r.kind(), r.Lived)
	}
	for _, k := range natures.sorted() {
		fought, won, lost, byWar := 0, 0, 0, 0
		for _, r := range recs {
			if r.kind() != k {
				continue
			}
			fought += r.Tally.Fought
			won += r.Tally.Taken
			lost += r.Tally.Lost
			if !r.Standing && (strings.Contains(r.Cause, "war") || strings.Contains(r.Cause, "taken by") || strings.Contains(r.Cause, "scoured") || strings.Contains(r.Cause, "unmade") || strings.Contains(r.Cause, "annihilated")) {
				byWar++
			}
		}
		p("| %s | %d | %d | %.2f | %.2f | %s | %.2f |", k, natures.n[k], fought, float64(won)/float64(max(fought, 1)), float64(lost)/float64(max(fought, 1)), pct(byWar, natures.n[k]), median(natures.sub[k]))
	}
}

func sortedKeysOf(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
