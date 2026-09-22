package main

import (
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"

	"worldgen/internal/legends"
	"worldgen/internal/record"
	"worldgen/internal/species"
)

// Kinds: substrate by modifier, with counts born, made and living at the
// end; the eldritch pool (powers held at birth and gained, deepenings);
// the living worlds' demands and wakings; the unmaking; the tithe; the
// hive's severings; and how each nature fares in war, so an eldritch
// thing can be read as neither trivial nor unbeatable.

// KindRec is one world's kinds.
type KindRec struct {
	Seed     uint64
	Powers   map[string]int // powers held at the end, by key, over every eldritch people
	Born     map[string]int // powers had at birth
	Deepened int
	Appeared int
	Demands  int
	Left     int // demands heeded
	Wakings  int
	Unmade   int
	Tithed   int // pairs
	Severed  int
	Sleeps   int
	Asleep   int // asleep at the present
	// the transmitter
	Transmitters int // made in this age, by a people or by the wall
	Elder        int // left by the deep pass
	Seeds        int // of all of them, carrying a seed
	Listeners    int // peoples taken: scarred, ended, or seeded
	SeedsWoke    int // parasites woken down a seed
	Speaking     int // live at the present
	// what eats
	Eaten    int // ships grown of what was eaten
	Consumed int // worlds stripped and held empty
	// what cannot be seen, and what does not stay put
	Hunts      int     // hunts declared on a hole in the ledger
	HuntWorlds int     // worlds a hunt took
	HuntFound  int     // hunts that fought at least one battle
	HuntEmpty  int     // hunts that ended with the hole closed
	Antimem    int     // anti-memetic peoples in the age
	AntimemEnd int     // of them ended
	Drifts     int     // drifts, over every evolver
	Evolvers   int     // evolvers that lived long enough to drift or not
	DriftKyr   float64 // thousand years lived by evolvers, for the rate
	DriftCures int
}

func flattenKinds(w *world) KindRec {
	r := KindRec{Seed: w.seed(), Powers: map[string]int{}, Born: map[string]int{}}
	for _, c := range w.State.Civs {
		sp := w.rd.Species(c)
		for _, k := range sp.Powers {
			r.Powers[k]++
		}
		t := c.Batch.Tally
		r.Deepened += t.Deepened
		r.Appeared += t.Appeared
		r.Demands += t.Demands
		r.Wakings += t.Wakings
		r.Unmade += t.Unmade
		r.Tithed += t.Tithed
		r.Sleeps += t.Sleeps
		if legends.Active(c) && c.Asleep {
			r.Asleep++
		}
		r.Eaten += t.Eaten
		r.Consumed += t.Consumed
		r.Hunts += t.Hunts
		if sp.Is(species.Antimemetic) {
			r.Antimem++
			if !legends.Active(c) {
				r.AntimemEnd++
			}
		}
		if sp.Is(species.Evolver) {
			r.Evolvers++
			r.Drifts += t.Drifts
			r.DriftKyr += float64(t.Ticks) * float64(w.Dossier.Step) / 1000
		}
	}
	for _, wr := range w.State.Wars {
		if wr.Hunt == nil {
			continue
		}
		r.HuntWorlds += wr.Taken[0]
		if wr.Result == "hole_closed" {
			r.HuntEmpty++
		}
		for _, b := range w.State.Battles {
			if b.Attacker == wr.Sides[0] && b.Defender == wr.Sides[1] && b.Year >= wr.Began && (b.Year <= wr.Ended || !wr.Over) {
				r.HuntFound++
				break
			}
		}
	}
	for _, l := range w.State.Remains {
		if !(l.Kind == "threat" && l.People < 0) { // a transmitter
			continue
		}
		if l.Age < 0 {
			r.Transmitters++
		} else {
			r.Elder++
		}
		if l.Payload == "seed" {
			r.Seeds++
		}
		r.Listeners += l.Listeners
		r.SeedsWoke += l.Woken
		if l.State == "unleashed" {
			r.Speaking++
		}
	}
	gained := map[string]int{} // by the power's key
	for _, f := range w.Chronicle {
		switch f.Kind {
		case record.FDemand:
			if f.Str("outcome") == "left" {
				r.Left++
			}
		case record.FSevered:
			r.Severed++
		case record.FDeepened:
			gained[f.Str("power")]++
		}
	}
	for _, pw := range species.Pool {
		r.Born[pw.Key] = r.Powers[pw.Key] - gained[pw.Key] // what is held and was not gained was there from birth
	}
	return r
}

func kindsReport(out io.Writer, ks []KindRec, recs []Rec, seeds int) {
	p := func(format string, a ...any) { fmt.Fprintf(out, format+"\n", a...) }
	p("")
	p("## Kinds")
	p("")
	p("Substrate by modifier: peoples born at a cradle, made (by a maker, a plague that woke, a sundering), and living at the present. A swarm is a biological people with the swarming trait, counted with its nature.")
	p("")
	type cell struct{ born, made, living int }
	table := map[string]map[string]*cell{}
	subs := map[string]bool{}
	mods := map[string]bool{}
	get := func(sub, mod string) *cell {
		if table[sub] == nil {
			table[sub] = map[string]*cell{}
		}
		if table[sub][mod] == nil {
			table[sub][mod] = &cell{}
		}
		return table[sub][mod]
	}
	for _, r := range recs {
		mod := "none"
		if len(r.Mods) > 0 {
			mod = strings.Join(r.Mods, "+")
		}
		subs[r.Sub], mods[mod] = true, true
		c := get(r.Sub, mod)
		if r.Made != "" || r.Origin != "" {
			c.made++
		} else {
			c.born++
		}
		if r.Standing {
			c.living++
		}
	}
	subList, modList := sortedKeysOf(subs), sortedKeysOf(mods)
	p("| Nature | %s |", strings.Join(subList, " | "))
	p("|---|%s", strings.Repeat("---|", len(subList)))
	for _, m := range modList {
		var cells []string
		for _, s := range subList {
			c := table[s][m]
			if c == nil {
				cells = append(cells, "")
				continue
			}
			cells = append(cells, fmt.Sprintf("%d born, %d made, %d living", c.born, c.made, c.living))
		}
		p("| %s | %s |", m, strings.Join(cells, " | "))
	}
	p("")
	// the pool
	powers := map[string]int{}
	born := map[string]int{}
	deep, appeared, demands, left, wakings, unmade, tithed, severed, sleeps, asleep := 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
	for _, k := range ks {
		for key, n := range k.Powers {
			powers[key] += n
		}
		for key, n := range k.Born {
			born[key] += n
		}
		deep += k.Deepened
		appeared += k.Appeared
		demands += k.Demands
		left += k.Left
		wakings += k.Wakings
		unmade += k.Unmade
		tithed += k.Tithed
		severed += k.Severed
		sleeps += k.Sleeps
		asleep += k.Asleep
	}
	eld := count(recs, func(r Rec) bool { return r.Sub == "eldritch" })
	p("The eldritch pool: %d eldritch peoples over %d worlds, %d deepenings (%.1f per eldritch people), %d second presences appeared.", eld, seeds, deep, float64(deep)/float64(max(eld, 1)), appeared)
	p("")
	p("| Power | Held at the end | Of them from birth |")
	p("|---|---|---|")
	for _, pw := range species.Pool {
		p("| %s | %d | %d |", pw.Name, powers[pw.Key], max(0, born[pw.Key]))
	}
	p("")
	p("Living worlds: %d demands made, %d heeded (%s), %d wakings. The unmaking turned on a world %d times. The tithe fell on %d peoples. Hive worlds severed from their seat: %d. Sleeps: %d, asleep at the present: %d.",
		demands, left, pct(left, demands), wakings, unmade, tithed, severed, sleeps, asleep)
	p("")
	// the transmitter and what eats
	made, elder, seeds, listeners, woke, live, eaten, consumed := 0, 0, 0, 0, 0, 0, 0, 0
	for _, k := range ks {
		made += k.Transmitters
		elder += k.Elder
		seeds += k.Seeds
		listeners += k.Listeners
		woke += k.SeedsWoke
		live += k.Speaking
		eaten += k.Eaten
		consumed += k.Consumed
	}
	p("Transmitters: %d made in the age and %d left by the deep pass, %d of them carrying a seed; %d listeners taken, %d things woken down a seed; %d speaking at the present.", made, elder, seeds, listeners, woke, live)
	var reps []Rec
	for _, r := range recs {
		if slices.Contains(r.Mods, "replicator") {
			reps = append(reps, r)
		}
	}
	repLiving := count(reps, func(r Rec) bool { return r.Standing })
	repBorn := count(reps, func(r Rec) bool { return r.Made == "" && r.Origin == "" })
	ends := newCounter()
	for _, r := range reps {
		if !r.Standing {
			ends.add(r.Cause, r.Lived)
		}
	}
	p("Things that eat: %d replicator peoples, %d born at a cradle and %d made; %d living at the present; %d ships grown of what was eaten, %d worlds stripped and held empty.", len(reps), repBorn, len(reps)-repBorn, repLiving, eaten, consumed)
	if len(ends.sorted()) > 0 {
		var parts []string
		for _, k := range ends.sorted() {
			parts = append(parts, fmt.Sprintf("%s (%d)", k, ends.n[k]))
		}
		p("What ended them: %s.", strings.Join(parts, "; "))
	}
	// what cannot be seen, and what does not stay put
	hunts, huntWorlds, huntFound, huntEmpty, antimem, antimemEnd, drifts, evolvers, driftKyr := 0, 0, 0, 0, 0, 0, 0, 0, 0.0
	for _, k := range ks {
		hunts += k.Hunts
		huntWorlds += k.HuntWorlds
		huntFound += k.HuntFound
		huntEmpty += k.HuntEmpty
		antimem += k.Antimem
		antimemEnd += k.AntimemEnd
		drifts += k.Drifts
		evolvers += k.Evolvers
		driftKyr += k.DriftKyr
	}
	worldsWith, hunting := 0, 0
	for _, k := range ks {
		if k.Antimem > 0 {
			worldsWith++
		}
		if k.Hunts > 0 {
			hunting++
		}
	}
	p("What cannot be seen: %d anti-memetic peoples in %d of %d worlds, %d of them ended; %d hunts declared on a hole in the ledger in %d worlds, %d of them found something to fight and took %d worlds between them, %d ended with the hole closed.", antimem, worldsWith, len(ks), antimemEnd, hunts, hunting, huntFound, huntWorlds, huntEmpty)
	rate := 0.0
	if driftKyr > 0 {
		rate = float64(drifts) / driftKyr * 1000
	}
	p("What does not stay put: %d evolvers, %d drifts, %.1f per million years of evolver life.", evolvers, drifts, rate)
	p("")
	// the natures in war
	p("| Nature | Peoples | Wars fought | Won per war | Lost per war | Ended by war | Median life (Myr) |")
	p("|---|---|---|---|---|---|---|")
	natures := newCounter()
	for _, r := range recs {
		natures.add(r.kind(), r.Lived)
	}
	for _, k := range natures.sorted() {
		fought, won, lost, byWar := 0, 0, 0, 0
		for _, r := range recs {
			if r.kind() != k {
				continue
			}
			fought += r.Tally.Fought
			won += r.Tally.Taken
			lost += r.Tally.Lost
			if !r.Standing && (strings.Contains(r.Cause, "war") || strings.Contains(r.Cause, "taken by") || strings.Contains(r.Cause, "scoured") || strings.Contains(r.Cause, "unmade") || strings.Contains(r.Cause, "annihilated")) {
				byWar++
			}
		}
		p("| %s | %d | %d | %.2f | %.2f | %s | %.2f |", k, natures.n[k], fought, float64(won)/float64(max(fought, 1)), float64(lost)/float64(max(fought, 1)), pct(byWar, natures.n[k]), median(natures.sub[k]))
	}
}

func sortedKeysOf(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
