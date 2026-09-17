package main

import (
	"fmt"
	"io"
	"sort"

	"worldgen/internal/flow"
	"worldgen/internal/tech"
)

// meansReport reads the flows: what peoples took in and what they owed at
// their height of means, how much of their lives they spent shedding,
// by era and alone against with colonies, and what went dark most. The
// calibration target is a cradle that runs era 2 in comfort, strains at
// era 3 and needs colonies or trade for era 4.
func meansReport(out io.Writer, recs []Rec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Means")
	p("")
	p("Income and upkeep are per tick, by kind (O organic matter, E energy, M metal), at the tick the people owed the most. A tick is shedding when any known node is dormant for want of its upkeep; alone means holding one system that tick.")
	p("")
	p("| Deepest era | Peoples | Income O / E / M | Upkeep O / E / M | Want O / E / M | Ticks shedding |")
	p("|---|---|---|---|---|---|")
	for e := 0; e <= 4; e++ {
		var in, up, want flow.Income
		n, ticks, lean := 0, 0, 0
		for _, r := range recs {
			if r.Era != e {
				continue
			}
			n++
			in.Add(r.Income)
			up.Add(r.Upkeep)
			want.Add(r.Want)
			ticks += r.Tally.Ticks
			lean += r.Tally.Lean
		}
		if n == 0 {
			continue
		}
		f := 1 / float64(n)
		p("| %d %s | %d | %.1f / %.1f / %.1f | %.1f / %.1f / %.1f | %.1f / %.1f / %.1f | %s |", e, tech.EraNames[e], n,
			in[flow.O]*f, in[flow.E]*f, in[flow.M]*f, up[flow.O]*f, up[flow.E]*f, up[flow.M]*f, want[flow.O]*f, want[flow.E]*f, want[flow.M]*f, pct(lean, ticks))
	}
	p("")
	p("Share of ticks shedding by the era the people was in at the time, over every people:")
	p("")
	p("| Era at the time | Ticks | Shedding | Ticks alone | Shedding alone | Ticks with colonies | Shedding with colonies |")
	p("|---|---|---|---|---|---|---|")
	for e := 0; e <= 4; e++ {
		ticks, lean, alone, leanAlone := 0, 0, 0, 0
		for _, r := range recs {
			ticks += r.Tally.TicksAt[e]
			lean += r.Tally.LeanAt[e]
			alone += r.Tally.AloneAt[e]
			leanAlone += r.Tally.LeanAloneAt[e]
		}
		if ticks == 0 {
			continue
		}
		p("| %d %s | %d | %s | %d | %s | %d | %s |", e, tech.EraNames[e], ticks, pct(lean, ticks), alone, pct(leanAlone, alone), ticks-alone, pct(lean-leanAlone, ticks-alone))
	}
	p("")
	shed := map[string]int{}
	total := 0
	for _, r := range recs {
		for k, n := range r.Shed {
			shed[k] += n
			total += n
		}
	}
	var keys []string
	for k := range shed {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if shed[keys[i]] != shed[keys[j]] {
			return shed[keys[i]] > shed[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > 12 {
		keys = keys[:12]
	}
	p("What went dark most, as a share of all node-ticks dormant:")
	p("")
	p("| Node | Era | Category | Upkeep O / E / M | Share |")
	p("|---|---|---|---|---|")
	for _, k := range keys {
		n := tech.Get(k)
		u := n.Upkeep()
		p("| %s | %d | %s | %.0f / %.0f / %.0f | %s |", n.Name, n.Era, n.Cat(), u[flow.O], u[flow.E], u[flow.M], pct(shed[k], total))
	}
}
