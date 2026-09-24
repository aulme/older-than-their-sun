// Command warshape reads the war shape of runs written earlier (worldgen
// -out, or TestWarShape with WAR_OUT) without running them again: the
// report and the gate TestWarShape prints, and on request the widest
// systems of wars, war by war. A question about a batch costs seconds,
// not a batch.
//
//	go run ./cmd/warshape [-base gate.json] [-systems N] dir...
//
// A directory holding seed-N run directories stands for all of them.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"worldgen/internal/record"
	"worldgen/internal/warshape"
)

func main() {
	base := flag.String("base", "", "a gate saved by TestWarShape (WAR_SAVE) to print beside this one")
	systems := flag.Int("systems", 0, "print each run's widest N systems of wars, war by war")
	events := flag.String("events", "", "civ,from,to: print the chronicle's events with that people in the years given (from the first run named)")
	census := flag.Bool("census", false, "print for each run how many peoples sent fleets, and the most in any million years")
	pairs := flag.Int("pairs", 0, "print every pair with more than N wars between them, war by war")
	flag.Parse()
	var dirs []string
	for _, a := range flag.Args() {
		if subs, _ := filepath.Glob(filepath.Join(a, "seed-*")); len(subs) > 0 {
			dirs = append(dirs, subs...)
			continue
		}
		dirs = append(dirs, a)
	}
	if len(dirs) == 0 {
		fmt.Fprintln(os.Stderr, "usage: warshape [-base gate.json] [-systems N] dir...")
		os.Exit(2)
	}
	runs := make([]*record.Run, len(dirs))
	shapes := make([]*warshape.Shape, len(dirs))
	var wg sync.WaitGroup
	for i, d := range dirs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := record.Load(d)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			runs[i], shapes[i] = r, warshape.Read(r)
		}()
	}
	wg.Wait()
	order := make([]int, len(dirs))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool { return shapes[order[a]].Seed < shapes[order[b]].Seed })
	ss := make([]*warshape.Shape, len(order))
	for k, i := range order {
		ss[k] = shapes[i]
	}
	for _, l := range warshape.Report(ss) {
		fmt.Println(l)
	}
	var b []warshape.Row
	if *base != "" {
		var err error
		if b, err = warshape.LoadGate(*base); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	fmt.Println()
	for _, l := range warshape.GateLines(warshape.Gate(ss), b) {
		fmt.Println(l)
	}
	if *events != "" {
		var civ int
		var from, to int64
		if _, err := fmt.Sscanf(*events, "%d,%d,%d", &civ, &from, &to); err != nil {
			fmt.Fprintln(os.Stderr, "-events wants civ,from,to:", err)
			os.Exit(2)
		}
		printEvents(runs[order[0]], civ, record.Year(from), record.Year(to))
		return
	}
	if *census {
		for _, i := range order {
			printCensus(runs[i])
		}
	}
	if *pairs > 0 {
		for _, i := range order {
			printPairs(shapes[i], *pairs)
		}
	}
	if *systems > 0 {
		for _, i := range order {
			printSystems(runs[i], shapes[i], *systems)
		}
	}
}

// printSystems prints a run's widest systems of wars: at the widest
// moment, each war open then, its sides, cause, aim, the pact it was
// joined under and its battles.
func printSystems(r *record.Run, s *warshape.Shape, n int) {
	joined := 0
	for _, w := range s.Wars {
		if w.Principal >= 0 {
			joined++
		}
	}
	fmt.Printf("\nseed %d: %d peoples, %d wars; %d pacts (the largest %d members at the end), %d wars joined by pact, %d separate peaces, %d allies absent, %d relief fleets\n",
		s.Seed, s.Peoples, len(s.Wars), s.Pacts, s.PactMax, joined, s.Separate, s.Absent, s.Relief)
	byID := map[int]*warshape.War{}
	for _, w := range s.Wars {
		byID[w.ID] = w
	}
	for k, sys := range s.Systems {
		if k == n {
			break
		}
		fmt.Printf("  system %d: %d peoples at year %d on %d fought fronts, %d wars in all\n", k+1, sys.Peak, sys.At, sys.Fronts, sys.Wars)
		for _, id := range sys.IDs {
			w := byID[id]
			if w.Began > sys.At || w.End < sys.At {
				continue
			}
			jn := ""
			if w.Principal >= 0 {
				jn = fmt.Sprintf(" joined for civ%d (pact %d)", w.Principal, w.Pact)
			}
			fmt.Printf("    war %d: civ%d (%d worlds) on civ%d (%d worlds), %s for %s%s; %d ticks, %d battles, campaigns %d/%d, %s\n",
				w.ID, w.Sides[0], w.Worlds[0], w.Sides[1], w.Worlds[1], w.Cause, w.Aim, jn, w.Ticks, w.Battles, w.Campaigns[0], w.Campaigns[1], w.Result)
		}
	}
}

// printPairs prints the pairs with more than n wars between them, each
// war with its cause, aim, length, battles and ending.
func printPairs(s *warshape.Shape, n int) {
	by := map[[2]int][]*warshape.War{}
	for _, w := range s.Wars {
		k := [2]int{min(w.Sides[0], w.Sides[1]), max(w.Sides[0], w.Sides[1])}
		by[k] = append(by[k], w)
	}
	var keys [][2]int
	for k, ws := range by {
		if len(ws) > n {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(a, b int) bool {
		return keys[a][0] < keys[b][0] || keys[a][0] == keys[b][0] && keys[a][1] < keys[b][1]
	})
	for _, k := range keys {
		ws := by[k]
		sort.Slice(ws, func(a, b int) bool { return ws[a].Began < ws[b].Began })
		fmt.Printf("\nseed %d: civ%d and civ%d, %d wars\n", s.Seed, k[0], k[1], len(ws))
		for _, w := range ws {
			joined := ""
			if w.Principal >= 0 {
				joined = fmt.Sprintf(" joined for civ%d", w.Principal)
			}
			fmt.Printf("  %d: war %d civ%d on civ%d (%d/%d worlds), %s for %s%s; %d ticks, %d battles, taken %d, campaigns %d/%d, %s\n",
				w.Began, w.ID, w.Sides[0], w.Sides[1], w.Worlds[0], w.Worlds[1], w.Cause, w.Aim, joined, w.Ticks, w.Battles, w.Taken, w.Campaigns[0], w.Campaigns[1], w.Result)
		}
	}
}

// printEvents prints a people's traits and every event it is in within
// the years given, with their parameters.
func printEvents(r *record.Run, civ int, from, to record.Year) {
	c := r.State.Civs[civ]
	sp := r.State.Species[c.Species]
	fmt.Printf("civ%d: species %d %v, born %d, ended %d (%s, %s)\n", civ, sp.ID, sp.Traits, c.Born, c.Ended, c.Fate, c.Cause)
	for _, e := range r.Chronicle {
		if e.Year < from || e.Year > to || (e.Subject != civ && e.Object != civ) {
			continue
		}
		fmt.Printf("  %d %s civ%d → civ%d at %d %v\n", e.Year, e.Kind, e.Subject, e.Object, e.Star, e.P)
	}
}

// printCensus prints how many peoples ever sent a fleet of any kind, and
// of a campaign, and the most distinct senders in any million years: how
// many peoples a seed has to make its wars from.
func printCensus(r *record.Run) {
	any, camp := map[int]bool{}, map[int]bool{}
	byMyr := map[record.Year]map[int]bool{}
	for _, x := range r.State.Fleets {
		any[x.Owner] = true
		if x.Kind == "campaign" {
			camp[x.Owner] = true
		}
		m := x.Launched / 1_000_000
		if byMyr[m] == nil {
			byMyr[m] = map[int]bool{}
		}
		byMyr[m][x.Owner] = true
	}
	most, at := 0, record.Year(0)
	for m, o := range byMyr {
		if len(o) > most || len(o) == most && m < at {
			most, at = len(o), m
		}
	}
	fmt.Printf("seed %d: %d peoples; %d sent fleets, %d sent a campaign; the most senders in one million years %d (at %d Myr); %d pacts\n",
		r.Dossier.Seed, len(r.State.Civs), len(any), len(camp), most, at, len(r.State.Pacts))
}
