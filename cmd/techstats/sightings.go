package main

import (
	"fmt"
	"io"
	"sort"
)

// Sightings: what saw fleets and how early, what was met in the dark,
// and what the battles left where they fell.

// SightRec is one sighting.
type SightRec struct {
	Seed      uint64
	Eye       string
	Kind      string
	Warning   int     // years from the sighting to the arrival
	Speed     float64 // years per light year
	Feasible  bool
	Intercept bool
}

// MeetRec is one battle in the dark.
type MeetRec struct {
	Seed   uint64
	Won    bool // the interceptor won the roll
	Broken bool
	Lost   [2]int // the interceptor's and the quarry's
}

// FleetRec is one campaign or relief fleet, seen or not.
type FleetRec struct {
	Seed    uint64
	Kind    string
	Speed   float64
	Seen    bool
	Warning int
}

// FieldRec is one world's fields.
type FieldRec struct {
	Seed                    uint64
	Fields, Ships, Adrift   int
	Wielded, Mastered, Ruin int
}

func flattenSightings(w *world) (sights []SightRec, meets []MeetRec, fleets []FleetRec, fields FieldRec) {
	for _, s := range w.State.Sightings {
		sights = append(sights, SightRec{Seed: w.seed(), Eye: s.Eye, Kind: s.Kind, Warning: int(s.Arrive - s.Year), Speed: s.Speed, Feasible: s.Feasible, Intercept: s.Intercept >= 0})
	}
	for _, m := range w.State.Meetings {
		meets = append(meets, MeetRec{Seed: w.seed(), Won: m.Won, Broken: m.Broken, Lost: m.Lost})
	}
	for _, x := range w.State.Fleets {
		if x.Kind != "campaign" && x.Kind != "relief" {
			continue
		}
		speed := x.Drive
		if speed == 0 {
			speed = w.civ(x.Owner).Speed
		}
		fleets = append(fleets, FleetRec{Seed: w.seed(), Kind: x.Kind, Speed: speed, Seen: len(x.Seen) > 0, Warning: int(x.Warning)})
	}
	fields.Seed = w.seed()
	for _, l := range w.State.Remains {
		if l.Kind != "field" {
			continue
		}
		fields.Fields++
		fields.Ships += l.Wrecks + l.Derelicts
		if l.Adrift {
			fields.Adrift++
		}
		switch l.State {
		case "wielded":
			fields.Wielded++
		case "mastered":
			fields.Mastered++
		}
		if l.Cond == "ruin" {
			fields.Ruin++
		}
	}
	return
}

// sightingsReport reads the watch: fleets seen before arrival by drive,
// what saw them, meetings in the dark and their outcome, pickets, and
// the fields the battles left. The calibration targets are most slow
// fleets seen and few torches, meetings a theatre of the slow drives,
// and early wars not going all to the defender.
func sightingsReport(out io.Writer, recs []Rec, sights []SightRec, meets []MeetRec, fleets []FleetRec, fields []FieldRec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Sightings")
	p("")
	p("A fleet in flight is seen by eyes: worlds by the tree, works by their own range, fleets and pickets at half the tree's. The sighting is on the fleet's timetable, not the tick's. A people at war that holds a sighting sends an interceptor to where the timetables cross, if they do before the fleet arrives; a fleet beaten in the dark turns back. Every ship lost lies where it fell.")
	p("")
	p("| Drive | Fleets | Seen before arrival | Warning, median years |")
	p("|---|---|---|---|")
	buckets := []struct {
		name   string
		lo, hi float64
	}{
		{"torch (4 years a light year or less)", 0, 4.0001}, {"sails (to 30)", 4.0001, 30.0001}, {"slow (over 30)", 30.0001, 1e9},
	}
	for _, bk := range buckets {
		n, seen := 0, 0
		var warn []float64
		for _, f := range fleets {
			if f.Speed < bk.lo || f.Speed >= bk.hi {
				continue
			}
			n++
			if f.Seen {
				seen++
				warn = append(warn, float64(f.Warning))
			}
		}
		p("| %s | %d | %s | %.0f |", bk.name, n, pct(seen, n), median(warn))
	}
	p("")
	byEye := map[string]int{}
	for _, s := range sights {
		byEye[s.Eye]++
	}
	var eyes []string
	for k := range byEye {
		eyes = append(eyes, k)
	}
	sort.Strings(eyes)
	line := fmt.Sprintf("%d sightings", len(sights))
	for _, k := range eyes {
		line += fmt.Sprintf(", %d by %s", byEye[k], eyeWord(k))
	}
	p("%s.", line)
	p("")
	feasible, sent := 0, 0
	for _, s := range sights {
		if s.Feasible {
			feasible++
		}
		if s.Intercept {
			sent++
		}
	}
	won, broken := 0, 0
	lost := [2]int{}
	for _, m := range meets {
		if m.Won {
			won++
		}
		if m.Broken {
			broken++
		}
		lost[0] += m.Lost[0]
		lost[1] += m.Lost[1]
	}
	p("Sightings with a meeting point against an enemy %d, interceptors sent %d, meetings fought %d; the interceptor won %d, fleets turned back %d, broken in the dark %d; ships lost by interceptors %d and by the fleets met %d.",
		feasible, sent, len(meets), won, won-broken, broken, lost[0], lost[1])
	p("")
	pickets, caught, salvaged := 0, 0, 0
	for _, r := range recs {
		pickets += r.Tally.Pickets
		caught += r.Tally.Caught
		salvaged += r.Tally.Salvaged
	}
	nf, ns, na, nw, nm, nr := 0, 0, 0, 0, 0, 0
	for _, f := range fields {
		nf += f.Fields
		ns += f.Ships
		na += f.Adrift
		nw += f.Wielded
		nm += f.Mastered
		nr += f.Ruin
	}
	p("Pickets sent %d; peoples' fleets caught in the dark %d. Fields of wrecks %d holding %d ships, %d adrift; crewed %d (%d ships flown home as salvage), read for their art %d, worn to ruin %d.",
		pickets, caught, nf, ns, na, nw, salvaged, nm, nr)
}

func eyeWord(k string) string {
	switch k {
	case "world":
		return "worlds"
	case "works":
		return "works"
	case "fleet":
		return "fleets"
	case "picket":
		return "pickets"
	case "sight":
		return "precognition"
	}
	return k
}
