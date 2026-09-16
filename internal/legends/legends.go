// Package legends renders a simulated history as readable text.
package legends

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"worldgen/internal/history"
	"worldgen/internal/tech"
)

func year(y history.Year) string {
	switch {
	case y == 0:
		return "present"
	case y <= -1_000_000_000:
		return fmt.Sprintf("%.2f Gyr ago", -float64(y)/1e9)
	case y <= -1_000_000:
		return fmt.Sprintf("%.2f Myr ago", -float64(y)/1e6)
	default:
		return fmt.Sprintf("%s ago", commas(-int64(y)))
	}
}

func systems(n int) string {
	if n == 1 {
		return "a single world"
	}
	return fmt.Sprintf("%d systems", n)
}

func keysOf(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
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

func levels(c *history.Civ) string {
	return fmt.Sprintf("military %s, survival %s, social %s", history.LevelName(c.Mil), history.LevelName(c.Sur), history.LevelName(c.Soc))
}

// Write prints the full legends: the ages of myth, deep time, the current
// age at both grains, and the present.
func Write(out io.Writer, w *history.World, full bool) {
	p := func(f string, a ...any) { fmt.Fprintf(out, f+"\n", a...) }
	events := func(from, to history.Year) {
		for _, e := range w.Events {
			if e.Year >= from && e.Year < to {
				p("  %-16s %s", year(e.Year), e.Text)
			}
		}
	}

	p("=== THE GALAXY ===")
	p("seed %d, %d stars within %.0f ly of Sol", w.Seed, len(w.G.Stars), w.Cfg.Radius)
	held, complex := 0, 0
	for i := range w.G.Stars {
		if w.Held[i] >= 0 {
			held++
		}
		if w.Bio[i] == history.BioComplex {
			complex++
		}
	}
	p("at the present: galactic hazard %.2f, %d stars held by horrors, %d worlds with complex life", w.Hazard, held, complex)
	cy := w.Cycle
	p("the cycle: period %.0f Myr, fade %.0f Myr; the current age woke %s, fertility now %.1f%% of the surge, next surge in %.0f Myr",
		float64(cy.Period)/1e6, float64(cy.Fade)/1e6, year(cy.Surges[len(cy.Surges)-1]), 100*w.FertilityNow(), float64(w.NextSurge())/1e6)
	p("")
	p("=== THE AGES OF MYTH (%s to %s) ===", year(w.Cfg.DeepStart), year(w.Cfg.MidStart))
	events(w.Cfg.DeepStart, w.Cfg.MidStart)
	p("")
	p("=== THE CURRENT AGE, EARLY (%s to %s) ===", year(w.Cfg.MidStart), year(w.Cfg.FineStart))
	events(w.Cfg.MidStart, w.Cfg.FineStart)
	p("")
	p("=== THE CURRENT AGE, LATE (%s to present) ===", year(w.Cfg.FineStart))
	events(w.Cfg.FineStart, 1)
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
		sc := ""
		if scars := keysOf(c.Scars); len(scars) > 0 {
			sc = " They live under " + strings.Join(scars, " and ") + "."
		}
		p("  The %s on %s, ruled by %s. %s. Once %s, %s. They %s (%s).%s", c.Name, w.G.Stars[c.Systems[0]].Name, c.Title, c.Species.Describe(), systems(c.Peak), tech.EraNames[c.Era], c.Cause, year(c.Ended), sc)
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
			if h.Dormant {
				state = fmt.Sprintf("silent at %s", w.G.Stars[h.Origin].Name)
			} else {
				state = fmt.Sprintf("broadcasting from %s, %d listeners lost", w.G.Stars[h.Origin].Name, h.Victims)
			}
		case history.Elder:
			state = fmt.Sprintf("sleeping near %s, woke %d times", w.G.Stars[h.Origin].Name, h.Wakings)
		}
		made := ""
		if h.FromCiv >= 0 {
			made = fmt.Sprintf(", made by the %s", w.Civs[h.FromCiv].Name)
		} else if h.Legacy >= 0 {
			l := w.Legacies[h.Legacy]
			made = fmt.Sprintf(", a legacy of %s", elderName(l))
		}
		p("  %s, a %s, %s%s.", h.Name, h.Kind, state, made)
	}
	p("")
	p("Legacies of the elder ages:")
	for _, l := range w.Legacies {
		p("  %-10s %-12s %s, at %s, %s.", l.Kind, l.State, l.Desc, w.G.Stars[l.Star].Name, elderName(l))
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
		p("  %-28s %d", k, kinds[k])
	}
	p("")
	p("Fates of all civilisations:")
	for _, c := range w.Civs {
		into := ""
		if c.Into != "" {
			into = " Became " + c.Into + "."
		}
		if c.KnowsCycle {
			into += " They knew the shape of the cycle."
		}
		made := ""
		if c.Species.Made != "" {
			made = " (" + c.Species.Made + ")"
		}
		kind := ""
		if c.Species.Kind != 0 {
			kind = "; " + c.Species.Kind.String()
		}
		seat := ""
		if c.Home != c.Cradle {
			seat = ", later seated on " + c.HomeName
		}
		p("  %-14s %-11s from %s (%s of a %s)%s%s, %s, lived %.2f Myr, peak %d systems, %s. They %s.%s", c.Name, c.Fate, c.CradleName, c.Species.World.Desc, w.G.Stars[c.Cradle].ClassName(), seat, made, year(c.Born), float64(c.Fell-c.Born)/1e6, c.Peak, tech.EraNames[c.Era], c.Cause, into)
		p("  %-14s %s%s. At the end: %s.", "", c.Species.Describe(), kind, levels(c))
		if len(c.Record) > 0 {
			p("  %-14s filters: %s", "", strings.Join(c.Record, "; "))
		}
		if scars := keysOf(c.Scars); len(scars) > 0 {
			p("  %-14s scars: %s", "", strings.Join(scars, ", "))
		}
		if full {
			var known []string
			for k := range c.Known {
				known = append(known, k)
			}
			sort.Strings(known)
			p("  %-14s known: %s", "", strings.Join(known, " "))
		}
	}
}

func elderName(l *history.Legacy) string {
	if l.Elder == nil {
		return "no one remembered"
	}
	if l.Elder.Name == "" {
		return "makers unnamed, " + l.Elder.Portrait
	}
	return l.Elder.Name + ", " + l.Elder.Portrait
}
