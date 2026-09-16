package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"worldgen/internal/history"
)

// WarRec is one war, for wars.jsonl.
type WarRec struct {
	Seed    uint64  `json:"seed"`
	A       string  `json:"a"`
	B       string  `json:"b"`
	PostA   string  `json:"posture_a"`
	PostB   string  `json:"posture_b"`
	Cause   string  `json:"cause"`
	Nth     int     `json:"nth"`
	Began   float64 `json:"began_myr"`
	Length  float64 `json:"length_kyr"`
	Taken   int     `json:"taken"`
	Glassed int     `json:"glassed"`
	Result  string  `json:"result"`
	Fine    bool    `json:"fine"` // fought in the fine pass, where duration means something
}

func flattenWars(w *history.World) []WarRec {
	var out []WarRec
	for _, wr := range w.Wars {
		a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
		r := WarRec{Seed: w.Seed, A: a.Name, B: b.Name, PostA: posture(a), PostB: posture(b), Cause: wr.Cause, Nth: wr.Nth,
			Began: float64(wr.Began-w.Cfg.Dawn) / 1e6, Taken: wr.Taken[0] + wr.Taken[1], Glassed: wr.Glassed[0] + wr.Glassed[1],
			Result: wr.Result, Fine: wr.Began >= w.Waning}
		if wr.Over {
			r.Length = float64(wr.Ended-wr.Began) / 1000
		} else {
			r.Result = "unfinished"
			r.Length = float64(w.Present-wr.Began) / 1000
		}
		out = append(out, r)
	}
	return out
}

func posture(c *history.Civ) string {
	for _, t := range c.Species.Traits {
		if t.Group == "stance" {
			return t.Key
		}
	}
	return ""
}

func writeWars(path string, wars []WarRec) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, r := range wars {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}

// warReport is the war and peace section of the report.
func warReport(out io.Writer, recs []Rec, wars []WarRec, seeds int) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## War and peace")
	p("")
	p("%d wars over %d worlds, %.1f per world; %d fought in the fine pass, where a duration means something.", len(wars), seeds, float64(len(wars))/float64(seeds), count2(wars, func(r WarRec) bool { return r.Fine }))
	p("")
	p("### How wars end")
	p("")
	results := map[string]int{}
	for _, r := range wars {
		results[r.Result]++
	}
	p("| Result | Wars | Share |")
	p("|---|---|---|")
	for _, k := range sortedKeys(results) {
		p("| %s | %d | %s |", k, results[k], pct(results[k], len(wars)))
	}
	p("")
	var lengths, moved []float64
	nth := 0
	for _, r := range wars {
		if r.Fine && r.Result != "unfinished" {
			lengths = append(lengths, r.Length)
		}
		moved = append(moved, float64(r.Taken+r.Glassed))
		if r.Nth > 1 {
			nth++
		}
	}
	if len(lengths) > 0 {
		p("Fine-pass wars last %.0f kyr at the median (quartiles %.0f to %.0f, longest %.0f).", median(lengths), quantile(lengths, 0.25), quantile(lengths, 0.75), quantile(lengths, 1))
	}
	p("Worlds changing hands or burned per war: median %.0f, mean %.1f. Wars that were the second or later between the same two: %s.", median(moved), mean(moved), pct(nth, len(wars)))
	p("")
	causes := map[string]int{}
	for _, r := range wars {
		causes[r.Cause]++
	}
	p("| Cause | Wars |")
	p("|---|---|")
	for _, k := range sortedByCount(causes) {
		p("| %s | %d |", k, causes[k])
	}
	p("")
	p("### By posture")
	p("")
	p("| Posture | Peoples | Met | Declared | Fought | Taken | Lost | Fleets | Went native | Scouts | Pacts | Betrayals | Ruled | Capitulated | Median life (Myr) |")
	p("|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|")
	byPost := map[string][]Rec{}
	for _, r := range recs {
		byPost[r.Posture] = append(byPost[r.Posture], r)
	}
	for _, k := range sortedKeys(byPost) {
		rs := byPost[k]
		var lives []float64
		var t history.Tally
		met, ruled, cap := 0, 0, 0
		for _, r := range rs {
			lives = append(lives, r.Lived)
			t.Declared += r.Tally.Declared
			t.Fought += r.Tally.Fought
			t.Taken += r.Tally.Taken
			t.Lost += r.Tally.Lost
			t.Fleets += r.Tally.Fleets
			t.Native += r.Tally.Native
			t.Scouts += r.Tally.Scouts
			t.Pacts += r.Tally.Pacts
			t.Betrayals += r.Tally.Betrayals
			met += r.Met
			ruled += r.Ruled
			if r.Tally.Capitulated {
				cap++
			}
		}
		n := float64(len(rs))
		p("| %s | %d | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %d | %.2f | %.2f | %d | %.2f | %s | %.2f |", k, len(rs), float64(met)/n, float64(t.Declared)/n, float64(t.Fought)/n, float64(t.Taken)/n, float64(t.Lost)/n, float64(t.Fleets)/n, t.Native, float64(t.Scouts)/n, float64(t.Pacts)/n, t.Betrayals, float64(ruled)/n, pct(cap, len(rs)), median(lives))
	}
	p("")
	p("### By honour")
	p("")
	p("| Honour | Peoples | Pacts | Relief sent | Called | Betrayals | Refused pacts | Median life (Myr) |")
	p("|---|---|---|---|---|---|---|---|")
	byHon := map[string][]Rec{}
	for _, r := range recs {
		byHon[r.Honour] = append(byHon[r.Honour], r)
	}
	for _, k := range sortedKeys(byHon) {
		rs := byHon[k]
		var lives []float64
		var t history.Tally
		for _, r := range rs {
			lives = append(lives, r.Lived)
			t.Pacts += r.Tally.Pacts
			t.Relief += r.Tally.Relief
			t.Called += r.Tally.Called
			t.Betrayals += r.Tally.Betrayals
			t.Refused += r.Tally.Refused
		}
		n := float64(len(rs))
		p("| %s | %d | %.2f | %d | %d | %d | %d | %.2f |", k, len(rs), float64(t.Pacts)/n, t.Relief, t.Called, t.Betrayals, t.Refused, median(lives))
	}
	p("")
	p("### Contact")
	p("")
	var mets []float64
	never := 0
	for _, r := range recs {
		mets = append(mets, float64(r.Met))
		if r.Met == 0 {
			never++
		}
	}
	p("A people meets %.1f others on average; %s never meet anyone. %d peoples were ruled by another at some point (%s).", mean(mets), pct(never, len(recs)), count(recs, func(r Rec) bool { return r.Ruled > 0 }), "as masters")
	p("")
	p("### Nomads")
	p("")
	nomads := count(recs, func(r Rec) bool { return r.Nomad })
	aloft := count(recs, func(r Rec) bool { return r.Aloft })
	rested := count(recs, func(r Rec) bool { return r.Rested })
	var lives, settled []float64
	for _, r := range recs {
		if r.Aloft {
			lives = append(lives, r.Lived)
		} else if r.Nomad {
			settled = append(settled, r.Lived)
		}
	}
	p("%d peoples born with the way (%s); %d took to the sky (%s of them), %d came to rest again. Those who flew lived %.2f Myr at the median; nomads who never reached the sky %.2f.", nomads, pct(nomads, len(recs)), aloft, pct(aloft, nomads), rested, median(lives), median(settled))
}

func count2(ws []WarRec, f func(WarRec) bool) int {
	n := 0
	for _, w := range ws {
		if f(w) {
			n++
		}
	}
	return n
}

func count(rs []Rec, f func(Rec) bool) int {
	n := 0
	for _, r := range rs {
		if f(r) {
			n++
		}
	}
	return n
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func sortedKeys[T any](m map[string]T) []string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func sortedByCount(m map[string]int) []string {
	ks := sortedKeys(m)
	sort.SliceStable(ks, func(i, j int) bool { return m[ks[i]] > m[ks[j]] })
	return ks
}
