package main

import (
	"fmt"
	"io"
)

// Blocs: nothing declared; the connected components of the partner graph
// as each people had it at the height of its means, and the wars inside
// a component against the wars across.

// BlocRec is one world's reading.
type BlocRec struct {
	Seed      uint64
	Peoples   int // in a component of two or more
	Blocs     int // components of two or more
	Largest   int
	Inside    int // wars between two members of one component
	Across    int // wars between members of different components
	Loose     int // wars with a side in no component
	Slights   float64
	Judged    int
	Deterred  int
	SlightsBy int // peoples that took any slight
}

func flattenBlocs(w *world) BlocRec {
	n := len(w.State.Civs)
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	union := func(a, b int) { parent[find(a)] = find(b) }
	linked := make([]bool, n)
	for _, c := range w.State.Civs {
		for _, p := range c.Batch.PeakTrade {
			union(c.ID, p)
			linked[c.ID], linked[p] = true, true
		}
	}
	size := map[int]int{}
	for _, c := range w.State.Civs {
		if linked[c.ID] {
			size[find(c.ID)]++
		}
	}
	r := BlocRec{Seed: w.seed()}
	for _, s := range size {
		if s >= 2 {
			r.Blocs++
			r.Peoples += s
			r.Largest = max(r.Largest, s)
		}
	}
	for _, wr := range w.State.Wars {
		a, b := wr.Sides[0], wr.Sides[1]
		switch {
		case !linked[a] || !linked[b]:
			r.Loose++
		case find(a) == find(b):
			r.Inside++
		default:
			r.Across++
		}
	}
	for _, c := range w.State.Civs {
		t := c.Batch.Tally
		r.Slights += t.Slights
		r.Judged += t.Judged
		r.Deterred += t.Deterred
		if t.Slights > 0 {
			r.SlightsBy++
		}
	}
	return r
}

// blocReport reads the blocs. The calibration targets are wars between
// partners of a common partner falling by more than seed noise, with the
// total read by bloc and not as one number.
func blocReport(out io.Writer, blocs []BlocRec, recs []Rec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Blocs and slights")
	p("")
	p("A war on a people's partner is a wrong done to that people, sized by their trade, taken as a grudge on the attacker; the council weighs the slights a war would give the peoples it minds. Blocs are nothing declared: the connected components of the partner graph as each people had it at the height of its means.")
	p("")
	var t BlocRec
	for _, b := range blocs {
		t.Peoples += b.Peoples
		t.Blocs += b.Blocs
		t.Largest = max(t.Largest, b.Largest)
		t.Inside += b.Inside
		t.Across += b.Across
		t.Loose += b.Loose
		t.Slights += b.Slights
		t.Judged += b.Judged
		t.Deterred += b.Deterred
		t.SlightsBy += b.SlightsBy
	}
	seeds := float64(max(len(blocs), 1))
	p("Blocs of two or more per world %.1f, peoples in one %.1f per world, the largest %d. Wars inside a bloc %d, across blocs %d, with a side in none %d.", float64(t.Blocs)/seeds, float64(t.Peoples)/seeds, t.Largest, t.Inside, t.Across, t.Loose)
	p("Slights taken %.1f per world by %.1f peoples per world; councils' verdicts %d, of which the offence alone held back %s.", t.Slights/seeds, float64(t.SlightsBy)/seeds, t.Judged, pct(t.Deterred, max(t.Judged, 1)))
}
