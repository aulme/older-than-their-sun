package legends

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"worldgen/internal/galaxy"
	"worldgen/internal/record"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// scarNames is a people's scars as the legends say them, in the order
// of their names.
func scarNames(keys []string) []string {
	var out []string
	for _, k := range keys {
		out = append(out, scarName(k))
	}
	sort.Strings(out)
	return out
}

func (v *view) levels(c *record.Civ) string {
	s := fmt.Sprintf("military %s, survival %s, social %s", levelName(c.Levels.Mil), levelName(c.Levels.Sur), levelName(c.Levels.Soc))
	if word := wisdomWord(c.Levels.Wis); word != "" {
		s += ", " + word
	}
	if c.Sellsword {
		s += ", sellswords"
	}
	return s
}

// arms is the military line of the portrait: the level, and the ships in
// being, in how many fleets, how many laid up, and the docks at work.
func (v *view) arms(c *record.Civ) string {
	ships, fleets, laid := v.shipsOf(c)
	if ships == 0 {
		if g, worlds := v.gunsOf(c); g > 0 {
			return fmt.Sprintf("no ships, %d guns over %d worlds", g, worlds)
		}
		return "no ships"
	}
	s := fmt.Sprintf("%d ships in %d fleets", ships, fleets)
	if fleets == 1 {
		s = fmt.Sprintf("%d ships in one fleet", ships)
	}
	if ships == 1 {
		s = "one ship"
	}
	if laid > 0 {
		s += fmt.Sprintf(", %d laid up", laid)
	}
	switch g, worlds := v.gunsOf(c); {
	case g == 0:
	case worlds == 1:
		s += fmt.Sprintf(", %d guns over one world", g)
	default:
		s += fmt.Sprintf(", %d guns over %d worlds", g, worlds)
	}
	switch d := c.Docks; d {
	case 0:
	case 1:
		s += ", one dock at work"
	default:
		s += fmt.Sprintf(", %d docks at work", d)
	}
	return s
}

// Write prints the full legends of a run: the ages of myth, deep time,
// the current age at both grains, and the present. The records hold ids
// and the templates print name tokens; every line goes out through the
// names table, which resolves them with the debug view's default rule.
func Write(out io.Writer, r *record.Run, full bool) {
	v := load(r)
	p := func(f string, a ...any) { fmt.Fprint(out, v.names.Text(fmt.Sprintf(f+"\n", a...))) }
	v.write(p, full)
}

func (v *view) write(p func(string, ...any), full bool) {
	d, st, g := v.d, v.st, v.g
	events := func(from, to Year) {
		for _, e := range v.r.Chronicle {
			if e.Year < from || e.Year >= to {
				continue
			}
			for _, line := range strings.Split(v.line(e), "\n") {
				if line != "" {
					p("  %-16s %s", v.year(e.Year), line)
				}
			}
		}
	}

	p("=== THE GALAXY ===")
	p("seed %d: %d stars within %.0f ly of %s", d.Seed, len(g.Stars), g.Radius, g.Anchor())
	p("The place: %s.", g.Region.Describe())
	for _, line := range lawsInWords(g) {
		p("  %s", line)
	}
	if d.Place.Real > 0 {
		p("  %d of the stars are real, with the worlds Earth knows of; the rest are drawn to the laws of the place.", d.Place.Real)
	}
	if near := nearFeatures(g, 14); len(near) > 0 {
		p("Near:")
		for _, n := range near {
			p("  %s", n)
		}
	}
	if sky := skyFeatures(g); sky != "" {
		p("Beyond: %s.", sky)
	}
	p("at the present: galactic hazard %.2f, %d stars held by things that eat them, %d transmitters speaking, %d worlds with complex life; the wall between this and what is beneath it is %s", d.Hazard, d.Counts.Eaten, d.Counts.Speaking, d.Counts.Complex, tables.wall[d.Wall.Stage].Word)
	cy := d.Cycle
	if d.Capped {
		p("WARNING: the age never wound down on its own; stopped after %.0f fades", d.MaxFades)
	}
	p("the cycle: period %.0f Myr, fade %.0f Myr; the current age dawned %s, fertility now %.1f%% of its dawn, next dawn in %.0f Myr",
		float64(cy.Period)/1e6, float64(cy.Fade)/1e6, v.year(cy.Surges[len(cy.Surges)-1]), 100*d.Fertility, float64(d.NextDawn-d.Present)/1e6)
	p("")
	p("=== THE AGES OF MYTH (%s to %s) ===", v.year(d.DeepStart), v.year(d.Dawn))
	events(d.DeepStart, d.Dawn)
	p("")
	p("=== THE YOUTH OF THE AGE (%s to %s) ===", v.year(d.Dawn), v.year(d.Waning))
	events(d.Dawn, d.Waning)
	p("")
	p("=== THE WANING (%s to present) ===", v.year(d.Waning))
	events(d.Waning, d.Present+1)
	p("")
	p("=== THE PRESENT: AFTERMATH ===")
	fates := map[string]int{}
	for _, c := range st.Civs {
		fates[c.Fate]++
	}
	p("Civilisations: %d arose. %d extinct, %d transformed, %d contracted, %d sundered into heirs, %d shattered into shards.", len(st.Civs), fates["extinct"], fates["transformed"], fates["contracted"], fates["sundered"], fates["shattered"])
	standing := 0
	for _, c := range st.Civs {
		if active(c) {
			standing++
		}
	}
	if standing > 0 {
		p("Still standing in the waning of the age: %d.", standing)
		for _, c := range st.Civs {
			sp := v.species(c)
			sick := ""
			if word := v.sickWord(c); word != "" {
				sick = "; " + word
			}
			if word := v.rideWord(c); word != "" {
				sick += "; " + word
			}
			line := ""
			if n := len(c.Line); n > 0 {
				line = ", heirs of the " + tok(c.Line[n-1])
			}
			ways := "; " + stiffWord(c)
			if len(c.Claim) > 0 {
				ways += ", claiming the old realm"
			}
			if c.Asleep {
				ways += ", asleep"
			}
			if ps := sp.PowerNames(); len(ps) > 0 {
				ways += "; it has " + strings.Join(ps, ", ")
			}
			if active(c) && c.Aloft {
				p("  The %s%s, aloft, seated for now at %s, %s, in %d fleets. %s. Now: %s; %s%s%s.", tok(c.ID), line, star(c.Home), tech.EraNames[c.Era], v.roaming(c), sp.Describe(), v.levels(c), v.arms(c), ways, sick)
			} else if active(c) {
				p("  The %s%s on %s, %s, holding %s. %s. Now: %s; %s%s%s.", tok(c.ID), line, star(c.Home), tech.EraNames[c.Era], systems(len(c.Systems)), sp.Describe(), v.levels(c), v.arms(c), ways, sick)
			}
		}
	}
	p("")
	p("Remnant civilisations still living:")
	any := false
	for _, c := range st.Civs {
		if !remnant(c) {
			continue
		}
		any = true
		sc := ""
		if scars := scarNames(c.Scars); len(scars) > 0 {
			sc = " They live under " + strings.Join(scars, " and ") + "."
		}
		where := star(c.Home)
		if len(c.Systems) > 0 {
			where = star(c.Systems[0])
		} else {
			where = "no world of their own, drifting near " + star(c.Home)
		}
		p("  The %s on %s, ruled by %s. %s. Once %s, %s. They %s (%s).%s", tok(c.ID), where, title(c.ID), v.species(c).Describe(), systems(c.Peak), tech.EraNames[c.Era], v.causeText(c), v.year(c.Ended), sc)
	}
	if !any {
		p("  none")
	}
	p("")
	p("What is still there:")
	any = false
	for _, c := range st.Civs {
		sp := v.species(c)
		pr := sp.Profile()
		if !active(c) || !(pr.Eats || pr.Monster || c.Asleep || !pr.Can(species.Researches) || sp.Is(species.Antimemetic)) {
			continue
		}
		any = true
		state := fmt.Sprintf("holding %s", systems(len(c.Systems)))
		switch {
		case c.Asleep:
			state = fmt.Sprintf("asleep at %s since %s, woke %d times", star(c.Home), v.year(c.Slept), c.Batch.Tally.Wakings)
		case pr.Eats:
			state += fmt.Sprintf(", %d ships grown of what it ate, %d worlds stripped", c.Batch.Tally.Eaten, c.Batch.Tally.Consumed)
		case sp.Is(species.Antimemetic):
			state += fmt.Sprintf(", which nobody who met them remembers; hunted %d times", v.hunted(c))
		}
		made := ""
		if c.Origin.Key != "" {
			made = ", " + v.originText(c.Origin)
		} else if sp.Made.Key != "" {
			made = ", " + v.originText(sp.Made)
		}
		p("  The %s, %s, %s%s.", tok(c.ID), sp.Describe(), state, made)
	}
	for _, l := range st.Remains {
		if !transmitter(l) || l.State == "lost" {
			continue
		}
		any = true
		state := fmt.Sprintf("silent at %s", star(l.Star))
		if speakingRemain(l) {
			state = fmt.Sprintf("speaking from %s, %d listeners taken", star(l.Star), l.Listeners)
			if l.Woken > 0 {
				state += fmt.Sprintf(", %d things woken down it", l.Woken)
			}
		}
		made := ""
		if l.Maker >= 0 {
			made = fmt.Sprintf(", in the voice of the %s", tok(l.Maker))
		} else if l.Elder >= 0 {
			made = fmt.Sprintf(", a legacy of %s", v.elderName(l))
		}
		p("  A transmitter carrying a %s, %s%s.", l.Payload, state, made)
	}
	if !any {
		p("  nothing")
	}
	p("")
	v.lineage(p)
	p("")
	v.plagues(p)
	p("")
	v.wars(p)
	p("")
	v.tellings(p, full)
	p("")
	v.gazetteer(p)
	p("")
	p("Legacies of the elder ages:")
	for _, l := range st.Remains {
		if l.Maker < 0 {
			p("  %-10s %-12s %s, at %s, %s.", l.Kind, l.State, v.describe(l), star(l.Star), v.elderName(l))
		}
	}
	p("")
	p("Remains of this age:")
	ruins := map[string]int{}
	for _, l := range st.Remains {
		if l.Maker >= 0 {
			ruins[l.State]++
			if l.State != "lost" {
				carries := ""
				if n := len(l.Testament); n > 0 {
					carries = fmt.Sprintf(" It carries a telling of %d things.", n)
				}
				p("  %-10s %-12s %s, at %s.%s", l.Kind, l.State, v.describe(l), star(l.Star), carries)
				if full {
					for _, t := range v.walls[l.ID] {
						p("  %-10s %-12s   \"%s\"", "", "", v.tell(v.civ(l.Maker), t))
					}
				}
			}
		}
	}
	p("  (%d more have crumbled)", ruins["lost"])
	p("")
	kinds := map[string]int{}
	for _, t := range st.Traces {
		kinds[v.traceName(t)]++
	}
	keys := make([]string, 0, len(kinds))
	for k := range kinds {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	p("Traces left for whoever comes next: %d", len(st.Traces))
	for _, k := range keys {
		p("  %-28s %d", k, kinds[k])
	}
	p("")
	p("Fates of all civilisations:")
	for _, c := range st.Civs {
		sp := v.species(c)
		into := ""
		if c.Into != "" {
			into = " Became " + v.intoText(c) + "."
		}
		if c.KnowsCycle {
			into += " They knew the shape of the cycle."
		}
		if len(c.Miracles) > 0 {
			var ms []string
			for _, k := range sortedStrings(c.Miracles) {
				ms = append(ms, tech.Get(k).Name+" ("+c.Miracles[k]+")")
			}
			into += " Miracles: " + strings.Join(ms, ", ") + "."
		}
		if c.Named {
			if !sp.Voiceless() {
				into += " They called it " + word(c.ID) + "."
			}
		}
		if len(c.Taught) > 0 {
			var ts []string
			var ks []string
			for k := range c.Taught {
				ks = append(ks, k)
			}
			sort.Strings(ks)
			for _, k := range ks {
				ts = append(ts, tech.Get(k).Name+" (the "+tok(c.Taught[k])+")")
			}
			into += " Taught: " + strings.Join(ts, ", ") + "."
		}
		made := ""
		if sp.Made.Key != "" {
			made = " (" + v.originText(sp.Made) + ")"
		} else if c.Origin.Key != "" {
			made = " (" + v.originText(c.Origin) + ")"
		}
		kind := ""
		if sp.Sub != species.Biological || sp.Mods != 0 {
			kind = "; " + sp.Nature()
		}
		if first := st.Species[c.Species].First; first != c.ID {
			kind += "; the " + tok(first) + " by blood"
		}
		kind += "; " + moralityWord(c.Morality)
		seat := ""
		if c.Home != c.Cradle {
			seat = ", later seated on " + star(c.Home)
		}
		ruled := ""
		if c.Ruled > 0 {
			ruled = fmt.Sprintf(" and %d peoples", c.Ruled)
		}
		p("  %-14s %-11s from %s (%s of a %s)%s%s, %s, lived %.2f Myr, peak %d systems%s, %s. They %s.%s", tok(c.ID), c.Fate, star(c.Cradle), sp.World.Desc, g.Stars[c.Cradle].ClassName(), seat, made, v.year(c.Born), float64(c.Fell-c.Born)/1e6, c.Peak, ruled, tech.EraNames[c.Era], v.causeText(c), into)
		p("  %-14s %s%s. At the end: %s.", "", sp.Describe(), kind, v.levels(c))
		if len(c.Record) > 0 {
			var recs []string
			for _, r := range c.Record {
				recs = append(recs, v.recordText(r))
			}
			p("  %-14s filters: %s", "", strings.Join(recs, "; "))
		}
		if scars := scarNames(c.Scars); len(scars) > 0 {
			p("  %-14s scars: %s", "", strings.Join(scars, ", "))
		}
		if full {
			p("  %-14s known: %s", "", strings.Join(c.Known, " "))
		}
	}
}

// lineage is the family tree of every empire that broke: each people that
// ended in a sundering or a shattering with its heirs, their heirs in
// turn, and who holds the old seat now.
func (v *view) lineage(p func(string, ...any)) {
	p("Lines: the empires that broke, and their heirs.")
	any := false
	var tree func(c *record.Civ, depth int)
	tree = func(c *record.Civ, depth int) {
		pad := strings.Repeat("    ", depth)
		state := ""
		switch {
		case active(c):
			state = fmt.Sprintf("standing, %s", stiffWord(c))
		case remnant(c):
			state = "a remnant"
		case c.Fate == "sundered":
			state = fmt.Sprintf("sundered %s", v.year(c.Ended))
		case c.Fate == "shattered":
			state = fmt.Sprintf("shattered %s", v.year(c.Ended))
		default:
			state = fmt.Sprintf("%s %s: %s", c.Fate, v.year(c.Ended), v.causeText(c))
		}
		seat := ""
		if o := v.held(c.Cradle); o >= 0 && o != c.ID && len(v.civ(o).Line) > 0 && contains(v.civ(o).Line, c.ID) {
			seat = fmt.Sprintf("; the old seat is held by the %s", tok(o))
		}
		p("  %sThe %s (%s), %s%s.", pad, tok(c.ID), v.year(c.Born), state, seat)
		for _, h := range v.st.Civs {
			if n := len(h.Line); n > 0 && h.Line[n-1] == c.ID {
				tree(h, depth+1)
			}
		}
	}
	for _, c := range v.st.Civs {
		if (c.Fate == "sundered" || c.Fate == "shattered") && len(c.Line) == 0 {
			any = true
			tree(c, 0)
		}
	}
	if !any {
		p("  none")
	}
}

func contains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// plagues lists every plague alive at the present, endemic or raging,
// with its kind, its road so far and its toll, and every dead one that
// took more than three peoples.
func (v *view) plagues(p func(string, ...any)) {
	p("Plagues:")
	any := false
	for _, pl := range v.st.Plagues {
		if pl.Hosts == 0 && pl.Peoples <= 3 {
			continue
		}
		any = true
		kind := "of the body"
		if pl.Kind == "memetic" {
			kind = "of the mind"
		}
		first := "nobody"
		if pl.FirstHost >= 0 {
			first = "the " + tok(pl.FirstHost)
		}
		if pl.Maker >= 0 {
			first += fmt.Sprintf(", made by the %s", tok(pl.Maker))
		}
		if pl.Rider >= 0 {
			first += fmt.Sprintf("; it thinks, and is the %s", tok(pl.Rider))
		}
		state := "dead"
		if pl.Hosts > 0 {
			var in []string
			for _, c := range v.st.Civs {
				inf, ok := c.Infections[pl.ID]
				if !active(c) || !ok {
					continue
				}
				how := "raging"
				switch {
				case inf.Carrier:
					how = "carried"
				case v.present-inf.Since > 1_000_000:
					how = fmt.Sprintf("endemic these %s, and they no longer notice it", spanOf(v.present-inf.Since))
				case inf.Contained:
					how = "contained"
				}
				in = append(in, fmt.Sprintf("the %s (%s)", tok(c.ID), how))
			}
			state = "in " + strings.Join(in, ", ")
		}
		p("  %s, %s, contagion %.2f, lethality %.2f: born %s among %s, %s. Has been in %d peoples, at most %d at once; took %d worlds and %d peoples, made %d cults, was cured %s.",
			upperFirst(plagueTok(pl.ID)), kind, pl.Contagion, pl.Lethality, v.year(pl.Born), first, state, pl.Caught, pl.Peak, pl.Worlds, pl.Peoples, pl.Cults, times(pl.Cures))
	}
	if !any {
		p("  none alive, and none that took more than three peoples")
	}
}

func times(n int) string {
	switch n {
	case 0:
		return "never"
	case 1:
		return "once"
	case 2:
		return "twice"
	}
	return fmt.Sprintf("%d times", n)
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// wars lists the wars worth remembering: those that took a world or ended
// a people, and every war fought more than once; then the pacts and the
// promises broken.
func (v *view) wars(p func(string, ...any)) {
	p("Wars of the age:")
	n := 0
	for _, wr := range v.st.Wars {
		moved := wr.Taken[0] + wr.Taken[1] + wr.Glassed[0] + wr.Glassed[1]
		if moved == 0 && wr.Nth == 1 && wr.Result == "peace" {
			continue
		}
		a, b := wr.Sides[0], wr.Sides[1]
		name := warTok(wr.ID) // a side's name for it, or "a war" where no side has a voice
		nth := ""
		if wr.Nth > 1 {
			nth = fmt.Sprintf(", their %s", ordinalOf(wr.Nth))
		}
		result := v.warResult(wr)
		if !wr.Over {
			result = "still fought"
		}
		length := ""
		if wr.Over {
			length = fmt.Sprintf(", %s", spanOf(wr.Ended-wr.Began))
		}
		p("  %s: the %s against the %s%s, over %s (%s%s). %s taken, %d burned; %s.", name, tok(a), tok(b), nth, v.warCause(wr), v.year(wr.Began), length, worldsOf(wr.Taken[0]+wr.Taken[1]), wr.Glassed[0]+wr.Glassed[1], result)
		n++
		if n >= 60 {
			p("  (and %d more)", len(v.st.Wars)-n)
			break
		}
	}
	if n == 0 {
		p("  none worth the telling")
	}
	p("")
	p("Pacts sworn:")
	n = 0
	for _, pc := range v.st.Pacts {
		var names []string
		for _, m := range pc.Members {
			names = append(names, tok(m))
		}
		against := "whoever came"
		if pc.Target >= 0 {
			against = "the " + tok(pc.Target)
		}
		state := "still held"
		if pc.Over {
			state = "broken " + v.year(pc.Ended)
		}
		p("  the %s, a pact of %s against %s (%s; %s).", strings.Join(names, " and the "), pc.Kind, against, v.year(pc.Formed), state)
		n++
	}
	if n == 0 {
		p("  none")
	}
	p("")
	p("Promises remembered:")
	n = 0
	for _, b := range v.st.Betrayals {
		if b.Weight <= 0 {
			continue
		}
		p("  The %s %s the %s (%s).", tok(b.By), v.betrayalText(b.Shape, b.Against), tok(b.Against), v.year(b.Year))
		n++
	}
	if n == 0 {
		p("  none broken")
	}
}

// roaming counts a nomad people's fleets in being.
func (v *view) roaming(c *record.Civ) int {
	n := 0
	for _, x := range v.st.Fleets {
		if !x.Over && x.Kind == "roam" && x.Owner == c.ID {
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

func spanOf(y Year) string {
	switch {
	case y < 1000:
		return "under a thousand years"
	default:
		return fmt.Sprintf("%d thousand years", y/1000)
	}
}

func sortedStrings(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (v *view) elderName(l *record.Remain) string {
	if l.Elder < 0 {
		return "no one remembered"
	}
	return elderTok(l.Elder) + ", " + v.elderPortrait(v.elder(l.Elder))
}

// sickWord is the portrait's line: what a people has, and for how long.
func (v *view) sickWord(c *record.Civ) string {
	var parts []string
	for _, pid := range sortedIDs(c.Infections) {
		inf := c.Infections[pid]
		state := "raging"
		switch {
		case inf.Carrier:
			state = "carried"
		case inf.Contained:
			state = "contained"
		}
		parts = append(parts, sprintf("sick with %s these %s, %s", plagueTok(pid), span(v.present-inf.Since), state))
	}
	return list(parts)
}

// rideWord is the portrait's line for a parasite: what it rides.
func (v *view) rideWord(c *record.Civ) string {
	if c.Own < 0 {
		return ""
	}
	var names []string
	for _, h := range v.hostsOf(c) {
		if h.Master == c.ID && !h.Vassal {
			names = append(names, "the "+tok(h.ID))
		}
	}
	if len(names) == 0 {
		return "riding nobody"
	}
	return "riding " + list(names)
}

func sortedIDs[V any](m map[int]V) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

// hunted counts the hunts declared on a people by the shape of the hole.
func (v *view) hunted(c *record.Civ) int {
	n := 0
	for _, wr := range v.st.Wars {
		if wr.Hunt != nil && wr.Sides[1] == c.ID {
			n++
		}
	}
	return n
}

// Stats prints one line of numbers about a run, for tuning across seeds.
func Stats(out io.Writer, r *record.Run) {
	d, st := r.Dossier, r.State
	var b [5]int
	standing, remnants, knowers := 0, 0, 0
	for _, c := range st.Civs {
		end := c.Fell
		if active(c) {
			end = d.Present
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
		if active(c) {
			standing++
		}
		if remnant(c) {
			remnants++
		}
		if c.KnowsCycle {
			knowers++
		}
	}
	miracles := map[string]int{}
	whole := 0
	for _, c := range st.Civs {
		for _, how := range c.Miracles {
			miracles[how]++
		}
		if len(c.Known) >= len(tech.Nodes)-len(tech.Miracles)-1 {
			whole++
		}
	}
	ruins := map[string]int{}
	conds := map[string]int{}
	nRuins := 0
	for _, l := range st.Remains {
		if l.Maker >= 0 {
			ruins[l.State]++
			nRuins++
			if l.State == "undisturbed" {
				conds[l.Cond]++
			}
		}
	}
	fmt.Fprintf(out, "%s seed %d: age %.1f Myr, fade %.0f Myr, fertility %.1f%%, %d civs (lived <0.5/<1/<3/<10/10+ Myr: %d/%d/%d/%d/%d), standing %d, remnants %d, knowers %d, whole tree %d, miracles born/leap/found/wielded %d/%d/%d/%d, transmitters %d, remains %d (mastered %d, wielded %d, sealed %d, unleashed %d, crumbled %d; still buried abandoned/derelict/wreck/ruin %d/%d/%d/%d), wall %.1f, capped %v\n",
		d.Place.Code,
		d.Seed, float64(d.Present-d.Dawn)/1e6, float64(d.Cycle.Fade)/1e6, 100*d.Fertility, len(st.Civs), b[0], b[1], b[2], b[3], b[4], standing, remnants, knowers, whole, miracles["born"], miracles["leap"], miracles["found"], miracles["wielded"], d.Counts.Speaking,
		nRuins, ruins["mastered"], ruins["wielded"], ruins["sealed"], ruins["unleashed"], ruins["lost"],
		conds["abandoned"], conds["derelict"], conds["wreck"], conds["ruin"], d.Wall.Value, d.Capped)
}

// starDetail names a star with its class and what Earth knows of it.
func (v *view) starDetail(s *galaxy.Star, id int) string {
	d := fmt.Sprintf("%s (%s", star(id), s.ClassName())
	if s.Real && s.Alt != "" {
		d += ", " + s.Alt
	}
	if s.Real && s.Mag < 6 && v.g.Sol >= 0 {
		d += ", a naked-eye star from Earth"
	}
	return d + ")"
}

// systemLine describes a star's system for the legends.
func (v *view) systemLine(id int) string {
	sys := v.g.Sys[id]
	name := star(id)
	line := "  " + name + ": " + sys.Describe(name) + "."
	if sys.Missed {
		line += " The home world is one Earth's surveys never saw."
	}
	return line
}

// lawsInWords tells the laws of the place as a people living there would.
func lawsInWords(g *galaxy.Galaxy) []string {
	l := g.Law
	var out []string
	sp := galaxy.MeanSpacing(len(g.Stars), g.Radius, g.Thickness)
	switch {
	case l.Density > 8:
		out = append(out, fmt.Sprintf("The stars stand close here, %.0f light years apart in this thinned field; no people is alone for long.", sp))
	case l.Density > 2:
		out = append(out, fmt.Sprintf("The stars stand closer than around the Sun, %.0f light years apart in this thinned field.", sp))
	case l.Density < 0.3:
		out = append(out, fmt.Sprintf("The stars are far apart, %.0f light years in this thinned field; a people that cannot cross that is alone.", sp))
	default:
		out = append(out, fmt.Sprintf("The stars stand about as they do around the Sun, %.0f light years apart in this thinned field.", sp))
	}
	switch {
	case l.Youth > 8:
		out = append(out, "The sky is full of giants: a star dies within sight every few thousand years, and the young suns burn hard.")
	case l.Youth > 2.5:
		out = append(out, "Young suns burn among the old and die young; the sky is never long without a new star.")
	case l.Youth < 0.25:
		out = append(out, "The sky is quiet and old; no star here will die for an age.")
	}
	switch {
	case l.Metals > 0.2:
		out = append(out, "The stars are rich in metal, and their worlds are heavy with it.")
	case l.Metals < -0.8:
		out = append(out, "The stars are ancient and poor in metal; worlds of rock are few, small and dry.")
	case l.Metals < -0.35:
		out = append(out, "The stars are poorer in metal than the Sun; worlds of rock are fewer.")
	}
	switch {
	case l.Glare > 10:
		out = append(out, "The heart of the galaxy fills the sky. Nothing keeps a thin skin for long; life that lasts lives underground or under ice.")
	case l.Glare > 3:
		out = append(out, "The sky is hard, and a biosphere on an open surface is a short-lived thing.")
	case l.Glare < 0.5:
		out = append(out, "The sky is soft and dark, the gentlest in the galaxy.")
	}
	if l.Crowd > 5 {
		out = append(out, "Other stars pass close enough to shake the comets loose, again and again.")
	}
	switch {
	case l.Exotic > 5:
		out = append(out, "Dead and collapsed stars are near, and those who study them learn the deep physics early, and things that should not be learned.")
	case l.Exotic > 2:
		out = append(out, "A dead star bends the light nearby; those who study it learn the deep physics sooner.")
	}
	switch {
	case l.R < 3:
		// the arms do not reach here; the zone says it all
	case l.Arm != nil:
		out = append(out, fmt.Sprintf("This is %s: %s.", l.Arm.Name, l.Arm.Desc))
	default:
		out = append(out, "This is the space between the arms, where nothing is born and little dies.")
	}
	return out
}

// nearFeatures lists the catalogued things near the field, nearest first.
func nearFeatures(g *galaxy.Galaxy, max int) []string {
	var out []string
	for _, nf := range g.Law.Near {
		if nf.F.Kind == galaxy.Sky {
			continue
		}
		if len(out) >= max {
			break
		}
		d := nf.Dist * galaxy.LyPerKpc
		where := fmt.Sprintf("%.0f ly", d)
		if d < 1 {
			where = "here"
		}
		desc := nf.F.Desc
		if desc == "" {
			desc = nf.F.Fact
		}
		out = append(out, fmt.Sprintf("%s (%s, %s): %s.", nf.F.Name, nf.F.Kind, where, desc))
	}
	return out
}

// skyFeatures names what lies beyond the galaxy.
func skyFeatures(g *galaxy.Galaxy) string {
	var names []string
	for _, nf := range g.Law.Near {
		if nf.F.Kind == galaxy.Sky {
			names = append(names, nf.F.Name)
		}
	}
	return strings.Join(names, ", ")
}
