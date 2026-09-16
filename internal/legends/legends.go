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

// present is set by Write; years are printed relative to it.
var present history.Year

func year(y history.Year) string {
	y -= present
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

	present = w.Present
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
	if w.Capped {
		p("WARNING: the age never wound down on its own; stopped after %.0f fades", w.Cfg.MaxFades)
	}
	p("the cycle: period %.0f Myr, fade %.0f Myr; the current age dawned %s, fertility now %.1f%% of its dawn, next dawn in %.0f Myr",
		float64(cy.Period)/1e6, float64(cy.Fade)/1e6, year(cy.Surges[len(cy.Surges)-1]), 100*w.FertilityNow(), float64(w.NextSurge()-w.Present)/1e6)
	p("")
	p("=== THE AGES OF MYTH (%s to %s) ===", year(w.Cfg.DeepStart), year(w.Cfg.Dawn))
	events(w.Cfg.DeepStart, w.Cfg.Dawn)
	p("")
	p("=== THE YOUTH OF THE AGE (%s to %s) ===", year(w.Cfg.Dawn), year(w.Waning))
	events(w.Cfg.Dawn, w.Waning)
	p("")
	p("=== THE WANING (%s to present) ===", year(w.Waning))
	events(w.Waning, w.Present+1)
	p("")
	p("=== THE PRESENT: AFTERMATH ===")
	fates := map[history.Fate]int{}
	for _, c := range w.Civs {
		fates[c.Fate]++
	}
	p("Civilisations: %d arose. %d extinct, %d transformed, %d contracted.", len(w.Civs), fates[history.Extinct], fates[history.Transformed], fates[history.Contracted])
	standing := 0
	for _, c := range w.Civs {
		if c.Active() {
			standing++
		}
	}
	if standing > 0 {
		p("Still standing in the waning of the age: %d.", standing)
		for _, c := range w.Civs {
			if c.Active() {
				p("  The %s on %s, %s, holding %s. %s. Now: %s.", c.Name, c.HomeName, tech.EraNames[c.Era], systems(len(c.Systems)), c.Species.Describe(), levels(c))
			}
		}
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
		case history.SleeperHorror:
			state = fmt.Sprintf("sleeping near %s, woke %d times", w.G.Stars[h.Origin].Name, h.Wakings)
		}
		made := ""
		if h.FromCiv >= 0 {
			made = fmt.Sprintf(", made by the %s", w.Civs[h.FromCiv].Name)
		} else if h.Legacy >= 0 {
			l := w.Legacies[h.Legacy]
			if l.Maker >= 0 {
				made = fmt.Sprintf(", from a relic of the %s", w.Civs[l.Maker].Name)
			} else {
				made = fmt.Sprintf(", a legacy of %s", elderName(l))
			}
		}
		p("  %s, a %s, %s%s.", h.Name, h.Kind, state, made)
	}
	p("")
	p("Legacies of the elder ages:")
	for _, l := range w.Legacies {
		if l.Maker < 0 {
			p("  %-10s %-12s %s, at %s, %s.", l.Kind, l.State, l.Desc, w.G.Stars[l.Star].Name, elderName(l))
		}
	}
	p("")
	p("Remains of this age:")
	ruins := map[history.LegacyState]int{}
	for _, l := range w.Legacies {
		if l.Maker >= 0 {
			ruins[l.State]++
			if l.State != history.Lost {
				p("  %-10s %-12s %s, at %s.", l.Kind, l.State, l.Describe(), w.G.Stars[l.Star].Name)
			}
		}
	}
	p("  (%d more have crumbled)", ruins[history.Lost])
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
		if len(c.Miracles) > 0 {
			var ms []string
			for _, k := range keysOf2(c.Miracles) {
				ms = append(ms, tech.Get(k).Name+" ("+c.Miracles[k]+")")
			}
			into += " Miracles: " + strings.Join(ms, ", ") + "."
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
		ruled := ""
		if c.Ruled > 0 {
			ruled = fmt.Sprintf(" and %d peoples", c.Ruled)
		}
		p("  %-14s %-11s from %s (%s of a %s)%s%s, %s, lived %.2f Myr, peak %d systems%s, %s. They %s.%s", c.Name, c.Fate, c.CradleName, c.Species.World.Desc, w.G.Stars[c.Cradle].ClassName(), seat, made, year(c.Born), float64(c.Fell-c.Born)/1e6, c.Peak, ruled, tech.EraNames[c.Era], c.Cause, into)
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

func keysOf2(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
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

// Stats prints one line of numbers about a world, for tuning across seeds.
func Stats(out io.Writer, w *history.World) {
	var b [5]int
	standing, remnants, knowers := 0, 0, 0
	for _, c := range w.Civs {
		lived := float64(c.Fell-c.Born) / 1e6
		switch {
		case lived < 0.5:
			b[0]++
		case lived < 1:
			b[1]++
		case lived < 3:
			b[2]++
		case lived < 10:
			b[3]++
		default:
			b[4]++
		}
		if c.Active() {
			standing++
		}
		if c.Stage == history.Remnant {
			remnants++
		}
		if c.KnowsCycle {
			knowers++
		}
	}
	miracles := map[string]int{}
	whole := 0
	for _, c := range w.Civs {
		for _, how := range c.Miracles {
			miracles[how]++
		}
		if len(c.Known) >= len(tech.Nodes)-len(tech.Miracles)-1 {
			whole++
		}
	}
	ruins := map[history.LegacyState]int{}
	conds := map[history.Condition]int{}
	nRuins := 0
	for _, l := range w.Legacies {
		if l.Maker >= 0 {
			ruins[l.State]++
			nRuins++
			if l.State == history.Buried {
				conds[l.Cond]++
			}
		}
	}
	fmt.Fprintf(out, "seed %d: age %.1f Myr, fade %.0f Myr, fertility %.1f%%, %d civs (lived <0.5/<1/<3/<10/10+ Myr: %d/%d/%d/%d/%d), standing %d, remnants %d, knowers %d, whole tree %d, miracles born/leap/found/wielded %d/%d/%d/%d, horrors %d, remains %d (mastered %d, wielded %d, sealed %d, unleashed %d, crumbled %d; still buried abandoned/derelict/wreck/ruin %d/%d/%d/%d), capped %v\n",
		w.Seed, float64(w.Present-w.Cfg.Dawn)/1e6, float64(w.Cycle.Fade)/1e6, 100*w.FertilityNow(), len(w.Civs), b[0], b[1], b[2], b[3], b[4], standing, remnants, knowers, whole, miracles["born"], miracles["leap"], miracles["found"], miracles["wielded"], len(w.Horrors),
		nRuins, ruins[history.Mastered], ruins[history.Wielded], ruins[history.Sealed], ruins[history.Unleashed], ruins[history.Lost],
		conds[history.Abandoned], conds[history.Derelict], conds[history.Wreck], conds[history.Ruin], w.Capped)
}
