// Package legends renders a simulated history as readable text.
package legends

import (
	"fmt"
	"io"
	"sort"

	"worldgen/internal/history"
)

func year(y history.Year) string {
	switch {
	case y == 0:
		return "present"
	case y <= -1_000_000:
		return fmt.Sprintf("%.1f Myr ago", -float64(y)/1e6)
	default:
		return fmt.Sprintf("%s ago", commas(-int64(y)))
	}
}

func commas(n int64) string {
	s := fmt.Sprint(n)
	out := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += ","
		}
		out += string(c)
	}
	return out
}

// Write prints the full legends: deep time, the recent age, and the present.
func Write(out io.Writer, w *history.World) {
	p := func(f string, a ...any) { fmt.Fprintf(out, f+"\n", a...) }

	p("=== THE GALAXY ===")
	p("seed %d, %d stars within %.0f ly of Sol", w.Seed, len(w.G.Stars), w.Cfg.Radius)
	p("")
	p("=== DEEP TIME (%s to %s) ===", year(w.Cfg.DeepStart), year(w.Cfg.DeepEnd))
	for _, e := range w.Events {
		if e.Year < w.Cfg.DeepEnd {
			p("  %-16s %s", year(e.Year), e.Text)
		}
	}
	p("")
	p("=== THE RECENT AGE (%s to present) ===", year(w.Cfg.DeepEnd))
	for _, e := range w.Events {
		if e.Year >= w.Cfg.DeepEnd {
			p("  %-16s %s", year(e.Year), e.Text)
		}
	}
	p("")
	p("=== THE PRESENT: AFTERMATH ===")
	fates := map[history.Fate]int{}
	for _, c := range w.Civs {
		fates[c.Fate]++
	}
	p("Civilisations: %d arose. %d extinct, %d transformed, %d contracted.", len(w.Civs), fates[history.Extinct], fates[history.Transformed], fates[history.Contracted])
	if w.Dusk > 0 {
		p("  (%d were still active at the present and were pushed into the Long Dusk)", w.Dusk)
	}
	p("")
	p("Remnant civilisations still living:")
	any := false
	for _, c := range w.Civs {
		if c.Stage != history.Remnant {
			continue
		}
		any = true
		p("  The %s on %s, ruled by %s. Once %d systems, now tech %.1f. They %s (%s).", c.Name, w.G.Stars[c.Systems[0]].Name, c.Title, c.Peak, c.Tech, c.Cause, year(c.Ended))
	}
	if !any {
		p("  none")
	}
	p("")
	p("Things that are not civilisations:")
	for _, h := range w.Horrors {
		state := ""
		switch h.Kind {
		case history.Replicators, history.RogueMind:
			state = fmt.Sprintf("holding %d systems", len(h.Systems))
			if h.Dormant {
				state += ", silent"
			}
		case history.Beacon:
			state = fmt.Sprintf("still broadcasting from %s, %d listeners lost", w.G.Stars[h.Origin].Name, h.Victims)
		case history.Elder:
			state = fmt.Sprintf("sleeping near %s, woke %d times", w.G.Stars[h.Origin].Name, h.Wakings)
		}
		made := ""
		if h.FromCiv >= 0 {
			made = fmt.Sprintf(", made by the %s", w.Civs[h.FromCiv].Name)
		}
		p("  %s, a %s, %s%s.", h.Name, h.Kind, state, made)
	}
	p("")
	kinds := map[string]int{}
	for _, t := range w.Traces {
		kinds[t.Kind]++
	}
	keys := make([]string, 0, len(kinds))
	for k := range kinds {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	p("Traces left for whoever comes next: %d", len(w.Traces))
	for _, k := range keys {
		p("  %-24s %d", k, kinds[k])
	}
	p("")
	p("Fates of all civilisations:")
	for _, c := range w.Civs {
		into := ""
		if c.Into != "" {
			into = " Became " + c.Into + "."
		}
		p("  %-14s %-11s %s from %s, %s, peak %d systems, tech %.1f. They %s.%s", c.Name, c.Fate, c.Temper, c.HomeName, year(c.Born), c.Peak, c.Tech, c.Cause, into)
	}
}
