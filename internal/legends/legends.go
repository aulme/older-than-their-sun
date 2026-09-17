// Package legends renders a simulated history as readable text.
package legends

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"worldgen/internal/history"
	"worldgen/internal/species"
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
	p("seed %d: %d stars within %.0f ly of %s", w.Seed, len(w.G.Stars), w.G.Radius, w.G.Anchor())
	p("The place: %s.", w.G.Region.Describe())
	for _, line := range history.LawsInWords(w.G) {
		p("  %s", line)
	}
	real := 0
	for i := range w.G.Stars {
		if w.G.Stars[i].Real {
			real++
		}
	}
	if real > 0 {
		p("  %d of the stars are real, with the worlds Earth knows of; the rest are drawn to the laws of the place.", real)
	}
	if near := history.NearFeatures(w.G, 14); len(near) > 0 {
		p("Near:")
		for _, n := range near {
			p("  %s", n)
		}
	}
	if sky := history.SkyFeatures(w.G); sky != "" {
		p("Beyond: %s.", sky)
	}
	held, complex := 0, 0
	for i := range w.G.Stars {
		if w.Held[i] >= 0 {
			held++
		}
		if w.Bio[i] == history.BioComplex {
			complex++
		}
	}
	p("at the present: galactic hazard %.2f, %d stars held by horrors, %d worlds with complex life; the wall between this and what is beneath it is %s", w.Hazard, held, complex, w.ThinWord())
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
			if c.Active() && c.Aloft {
				p("  The %s, aloft, seated for now at %s, %s, in %d fleets. %s. Now: %s.", c.Name, c.HomeName, tech.EraNames[c.Era], fleetsOf(w, c), c.Species.Describe(), levels(c))
			} else if c.Active() {
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
		where := c.HomeName
		if len(c.Systems) > 0 {
			where = w.G.Stars[c.Systems[0]].Name
		} else {
			where = "no world of their own, drifting near " + c.HomeName
		}
		p("  The %s on %s, ruled by %s. %s. Once %s, %s. They %s (%s).%s", c.Name, where, c.Title, c.Species.Describe(), systems(c.Peak), tech.EraNames[c.Era], c.Cause, year(c.Ended), sc)
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
	wars(p, w)
	p("")
	tellings(p, w, full)
	p("")
	gazetteer(p, w)
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
				carries := ""
				if n := len(l.Testament); n > 0 {
					carries = fmt.Sprintf(" It carries a telling of %d things.", n)
				}
				p("  %-10s %-12s %s, at %s.%s", l.Kind, l.State, l.Describe(), w.G.Stars[l.Star].Name, carries)
				if full {
					for _, in := range l.Testament {
						p("  %-10s %-12s   \"%s\"", "", "", in.Text)
					}
				}
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
		if c.Word != "" {
			into += " They called it " + c.Word + "."
		}
		made := ""
		if c.Species.Made != "" {
			made = " (" + c.Species.Made + ")"
		} else if c.Origin != "" {
			made = " (" + c.Origin + ")"
		}
		kind := ""
		if c.Species.Sub != species.Biological || c.Species.Mods != 0 {
			kind = "; " + c.Species.Nature()
		}
		if c.Species.Name != c.Name {
			kind += "; the " + c.Species.Name + " by blood"
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

// wars lists the wars worth remembering: those that took a world or ended
// a people, and every war fought more than once; then the pacts and the
// promises broken.
func wars(p func(string, ...any), w *history.World) {
	p("Wars of the age:")
	n := 0
	for _, wr := range w.Wars {
		moved := wr.Taken[0] + wr.Taken[1] + wr.Glassed[0] + wr.Glassed[1]
		if moved == 0 && wr.Nth == 1 && wr.Result == "peace" {
			continue
		}
		a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
		name := wr.Name
		if name == "" {
			name = "a war"
		}
		nth := ""
		if wr.Nth > 1 {
			nth = fmt.Sprintf(", their %s", ordinalOf(wr.Nth))
		}
		result := wr.Result
		if !wr.Over {
			result = "still fought"
		}
		length := ""
		if wr.Over {
			length = fmt.Sprintf(", %s", spanOf(wr.Ended-wr.Began))
		}
		p("  %s: the %s against the %s%s, over %s (%s%s). %s taken, %d burned; %s.", name, a.Name, b.Name, nth, wr.Cause, year(wr.Began), length, worldsOf(wr.Taken[0]+wr.Taken[1]), wr.Glassed[0]+wr.Glassed[1], result)
		n++
		if n >= 60 {
			p("  (and %d more)", len(w.Wars)-n)
			break
		}
	}
	if n == 0 {
		p("  none worth the telling")
	}
	p("")
	p("Pacts sworn:")
	n = 0
	for _, pc := range w.Pacts {
		var names []string
		for _, m := range pc.Members {
			names = append(names, w.Civs[m].Name)
		}
		against := "whoever came"
		if pc.Target >= 0 {
			against = "the " + w.Civs[pc.Target].Name
		}
		state := "still held"
		if pc.Over {
			state = "broken " + year(pc.Ended)
		}
		p("  the %s, a pact of %s against %s (%s; %s).", strings.Join(names, " and the "), pc.Kind, against, year(pc.Formed), state)
		n++
	}
	if n == 0 {
		p("  none")
	}
	p("")
	p("Promises remembered:")
	n = 0
	for _, b := range w.Betrayals {
		if b.Weight <= 0 {
			continue
		}
		p("  The %s %s the %s (%s).", w.Civs[b.By].Name, b.Shape, w.Civs[b.Against].Name, year(b.Year))
		n++
	}
	if n == 0 {
		p("  none broken")
	}
}

func fleetsOf(w *history.World, c *history.Civ) int {
	n := 0
	for _, x := range w.Expeditions {
		if !x.Over && x.Kind == history.Roam && x.Owner == c.ID {
			n++
		}
	}
	return n
}

func worldsOf(n int) string {
	switch n {
	case 0:
		return "no worlds"
	case 1:
		return "one world"
	}
	return fmt.Sprintf("%d worlds", n)
}

func ordinalOf(n int) string {
	switch n {
	case 2:
		return "second"
	case 3:
		return "third"
	case 4:
		return "fourth"
	case 5:
		return "fifth"
	}
	return fmt.Sprintf("%dth", n)
}

func spanOf(y history.Year) string {
	switch {
	case y < 1000:
		return "under a thousand years"
	default:
		return fmt.Sprintf("%d thousand years", y/1000)
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
		end := c.Fell
		if c.Active() {
			end = w.Present
		}
		lived := float64(end-c.Born) / 1e6
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
	fmt.Fprintf(out, "%s seed %d: age %.1f Myr, fade %.0f Myr, fertility %.1f%%, %d civs (lived <0.5/<1/<3/<10/10+ Myr: %d/%d/%d/%d/%d), standing %d, remnants %d, knowers %d, whole tree %d, miracles born/leap/found/wielded %d/%d/%d/%d, horrors %d, remains %d (mastered %d, wielded %d, sealed %d, unleashed %d, crumbled %d; still buried abandoned/derelict/wreck/ruin %d/%d/%d/%d), wall %.1f, capped %v\n",
		w.G.Region.Code,
		w.Seed, float64(w.Present-w.Cfg.Dawn)/1e6, float64(w.Cycle.Fade)/1e6, 100*w.FertilityNow(), len(w.Civs), b[0], b[1], b[2], b[3], b[4], standing, remnants, knowers, whole, miracles["born"], miracles["leap"], miracles["found"], miracles["wielded"], len(w.Horrors),
		nRuins, ruins[history.Mastered], ruins[history.Wielded], ruins[history.Sealed], ruins[history.Unleashed], ruins[history.Lost],
		conds[history.Abandoned], conds[history.Derelict], conds[history.Wreck], conds[history.Ruin], w.Thin, w.Capped)
}
