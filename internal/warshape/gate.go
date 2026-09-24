package warshape

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Row is one line of the war gate (specs/proposals/war.md, "The gate"):
// what the batch shows against what is wanted, and whether it passes.
type Row struct {
	Key  string `json:"key"`
	Got  string `json:"got"`
	Want string `json:"want"`
	Pass bool   `json:"pass"`
}

// most is "most seeds": fourteen of twenty.
func most(n, of int) bool { return 10*n >= 7*of }

// Gate reads a batch against every row of the war gate, steps 10's and
// 11's together, so a batch says at once what it passes.
func Gate(ss []*Shape) []Row {
	var rows []Row
	add := func(key string, pass bool, want, got string, a ...any) {
		rows = append(rows, Row{Key: key, Got: fmt.Sprintf(got, a...), Want: want, Pass: pass})
	}
	var all []*War
	var sys []System
	var wv []int
	var pw, pf []int
	cold, px := 0, 0
	var took time.Duration
	for _, s := range ss {
		all = append(all, s.Wars...)
		sys = append(sys, s.Systems...)
		wv = append(wv, s.Waves...)
		pw = append(pw, s.PairWars...)
		pf = append(pf, s.PairFought...)
		cold += s.Cold
		px += s.Proxy
		took += s.Took
	}
	seeds := func(ok func(*Shape) bool) int {
		n := 0
		for _, s := range ss {
			if ok(s) {
				n++
			}
		}
		return n
	}
	n := len(ss)
	fw := fleetWars(all)
	f := fought(all)
	share := func(k, of int) float64 {
		if of == 0 {
			return 0
		}
		return float64(k) / float64(of)
	}
	fs := share(len(f), len(fw))
	add("fought", fs >= 0.6, "60% of fleet wars", "%s of fleet wars", frac(fs))
	sh, lg, _ := lengths(f)
	add("short/long", share(sh, len(f)) >= 0.25 && share(lg, len(f)) >= 0.25, "a quarter each", "%s / %s", pct(sh, len(f)), pct(lg, len(f)))
	quarter := func(long bool) func(*Shape) bool {
		return func(s *Shape) bool {
			f := fought(s.Wars)
			sh, lg, _ := lengths(f)
			if long {
				return share(lg, len(f)) >= 0.25
			}
			return share(sh, len(f)) >= 0.25
		}
	}
	ss0, sl := seeds(quarter(false)), seeds(quarter(true))
	add("short seeds", most(ss0, n), "most seeds", "%d of %d", ss0, n)
	add("long seeds", most(sl, n), "most seeds", "%d of %d", sl, n)
	lc := longCarried(f)
	add("long carried", lc >= 1.0/3, "a third", "%s (battles %s)", frac(lc), frac(longFought(f)))
	old, loops := 0, 0
	for i := range pw {
		if pf[i] >= oldWars && pw[i] <= loopWars {
			old++
		}
		if pw[i] > loopWars {
			loops++
		}
	}
	oe := seeds(func(s *Shape) bool { o, _ := enemies(s); return o > 0 })
	add("old enemies", old >= 100 && most(oe, n), "100+, most seeds", "%d pairs, %d seeds", old, oe)
	add("loops", loops == 0, "none", "%d", loops)
	ws := seeds(func(s *Shape) bool { return worldWars(s.Systems) > 0 })
	add("world wars", most(ws, n), "most seeds", "%d, %d seeds", worldWars(sys), ws)
	b := border(all)
	bs := share(b.border, b.large)
	commonest := true
	for k, v := range b.kind {
		if k != "border: short, a world or two taken" && v > b.border {
			commonest = false
		}
	}
	add("border", bs >= 1.0/3 && commonest, "a third, the most common", "%s (long %s)", frac(bs), pct(b.kind["long"], b.large))
	cs := seeds(func(s *Shape) bool { return s.Cold > 0 })
	add("cold wars", most(cs, n), "most seeds, with the build-up", "%d pairs, %d seeds", cold, cs)
	ps := seeds(func(s *Shape) bool { return s.Proxy > 0 })
	add("proxy wars", px >= 5 && ps >= 3, "5+, 3+ seeds", "%d, %d seeds", px, ps)
	vs := seeds(func(s *Shape) bool { return len(s.Waves) > 0 })
	add("waves", len(wv) >= 5 && vs >= 3, "5+, 3+ seeds", "%d, %d seeds", len(wv), vs)
	sb := share(countW(f, func(w *War) bool { return w.Campaigns[1] > 0 }), len(f))
	add("strike back", sb >= 1.0/3, "a third of fought wars", "%s", frac(sb))
	res := map[string]int{}
	for _, w := range all {
		res[w.Result]++
	}
	defeat := res["fall"] + res["enslaved"] + res["capitulation"] + res["vassal"]
	add("endings", share(res["truce"], len(all)) < 0.5 && share(defeat, len(all)) >= 0.1 && share(res["terms"], len(all)) >= 0.1, "truce a minority, defeat and terms common",
		"terms %s, defeat %s, truce %s", pct(res["terms"], len(all)), pct(defeat, len(all)), pct(res["truce"], len(all)))
	lo, hi, capped := 1e9, 0.0, 0
	for _, s := range ss {
		lo, hi = min(lo, s.Ages), max(hi, s.Ages)
		if s.Capped {
			capped++
		}
	}
	add("ages", lo >= 20 && hi <= 80 && capped == 0, "20 to 80 Myr, none capped", "%.0f to %.0f Myr, %d capped", lo, hi, capped)
	add("run time", took <= 3*time.Hour*time.Duration(n)/20, "within twice stage 0", "%s", took.Round(time.Minute))
	return rows
}

// GateLines prints the gate, beside a base batch's when one is given.
func GateLines(rows, base []Row) []string {
	was := map[string]Row{}
	for _, r := range base {
		was[r.Key] = r
	}
	out := []string{"| row | got | wanted | |"}
	if base != nil {
		out = []string{"| row | base | got | wanted | |"}
	}
	for _, r := range rows {
		mark := "✗"
		if r.Pass {
			mark = "✓"
		}
		if base != nil {
			out = append(out, fmt.Sprintf("| %s | %s | %s | %s | %s |", r.Key, was[r.Key].Got, r.Got, r.Want, mark))
			continue
		}
		out = append(out, fmt.Sprintf("| %s | %s | %s | %s |", r.Key, r.Got, r.Want, mark))
	}
	return out
}

// SaveGate writes a gate to a file, for a later batch to be read beside.
func SaveGate(path string, rows []Row) error {
	b, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// LoadGate reads one back.
func LoadGate(path string) ([]Row, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rows []Row
	return rows, json.Unmarshal(b, &rows)
}
