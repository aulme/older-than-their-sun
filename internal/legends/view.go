// Package legends renders a run directory as readable text: the debug
// view of the record, and the proof that nothing the prose said was
// lost in the files. It reads the dossier, the state, the chronicle and
// the tellings, the lookups in data/ and the names in the state, and
// never the simulation.
package legends

import (
	"fmt"
	"sort"
	"strconv"

	"worldgen/internal/galaxy"
	"worldgen/internal/names"
	"worldgen/internal/record"
	"worldgen/internal/species"
)

// Year is the record's.
type Year = record.Year

// view is a loaded run with what the rendering needs at hand: the
// galaxy rebuilt from the stars, every blood as a species, the events by
// id, each people's tales, each remain's testament, and the names.
type view struct {
	r       *record.Run
	d       *record.Dossier
	st      *record.State
	g       *galaxy.Galaxy
	sp      []*species.Species
	ev      map[int]*record.Event
	lore    map[int][]*record.Tale // a people's living tales, in the order held
	walls   map[int][]*record.Tale // a remain's testament
	names   *names.Table
	present Year
}

// load builds the view of a run.
func load(r *record.Run) *view {
	v := &view{r: r, d: r.Dossier, st: r.State, ev: map[int]*record.Event{}, lore: map[int][]*record.Tale{}, walls: map[int][]*record.Tale{}, present: r.Dossier.Present}
	rg, err := galaxy.RegionByName(r.Dossier.Config.At)
	if err != nil {
		panic("legends: " + err.Error())
	}
	stars := make([]galaxy.Star, len(r.State.Stars))
	sys := make([]*galaxy.System, len(r.State.Stars))
	for i, s := range r.State.Stars {
		stars[i] = galaxy.Star{ID: s.ID, Name: s.Designation, Desig: s.Kind, Real: s.Real, Alt: s.Alt, Class: s.Class[0], Note: s.Note, Remnant: s.Remnant, X: s.X, Y: s.Y, Z: s.Z, Hab: s.Hab, Mult: s.Mult, Lifetime: s.Lifetime, DiesAt: int64(s.DiesAt), Failing: s.Failing, Mag: s.Mag}
		sys[i] = s.System
	}
	v.g = galaxy.Rebuild(stars, sys, r.Dossier.Place.Sol, r.Dossier.Config.Radius, r.Dossier.Config.Thickness, rg)
	v.sp = make([]*species.Species, len(r.State.Species))
	for i, s := range r.State.Species {
		v.sp[i] = species.Rebuild(s.ID, s.Sub, s.Mods, s.Channel, s.Powers, s.World, s.Traits, s.Made)
	}
	for i, s := range r.State.Species {
		if s.Parent >= 0 {
			v.sp[i].Parent = v.sp[s.Parent]
		}
	}
	for _, e := range r.Chronicle {
		v.ev[e.ID] = e
	}
	for _, t := range r.Tellings {
		if t.Remain >= 0 {
			v.walls[t.Remain] = append(v.walls[t.Remain], t)
		} else {
			v.lore[t.Civ] = append(v.lore[t.Civ], t)
		}
	}
	v.names = names.FromRows(r.State.Names,
		func(id int) (string, bool) {
			if id >= 0 && id < len(r.State.Stars) {
				return r.State.Stars[id].Designation, true
			}
			return "", false
		},
		func(id int) bool {
			return id >= 0 && id < len(r.State.Plagues) && r.State.Plagues[id].Kind == "memetic"
		},
		func(id int) int {
			if id >= 0 && id < len(r.State.Species) {
				return r.State.Species[id].First
			}
			return -1
		})
	return v
}

// The record's objects by id.
func (v *view) civ(id int) *record.Civ                 { return v.st.Civs[id] }
func (v *view) species(c *record.Civ) *species.Species { return v.sp[c.Species] }
func (v *view) remain(id int) *record.Remain           { return v.st.Remains[id] }
func (v *view) plague(id int) *record.Plague           { return v.st.Plagues[id] }
func (v *view) source(id int) *record.Source           { return v.st.Sources[id] }
func (v *view) war(id int) *record.War                 { return v.st.Wars[id] }
func (v *view) held(star int) int                      { return v.st.Stars[star].Held }
func (v *view) event(id int) *record.Event             { return v.ev[id] }

// elder finds an elder by id.
func (v *view) elder(id int) *record.Elder {
	for _, e := range v.st.Elders {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// The stages of a people.
func active(c *record.Civ) bool {
	return c.Stage == "emergent" || c.Stage == "interstellar" || c.Stage == "zenith"
}
func living(c *record.Civ) bool  { return c.Stage != "dead" }
func remnant(c *record.Civ) bool { return c.Stage == "remnant" }

// The tokens the view prints; the names table resolves them.
func tok(id int) string { return "{civ:" + itoa(id) + "}" }
func tokBy(id, by int, tone string) string {
	return "{civ:" + itoa(id) + "@" + itoa(by) + ":" + tone + "}"
}
func star(id int) string                { return "{star:" + itoa(id) + "}" }
func title(id int) string               { return "{title:" + itoa(id) + "}" }
func word(id int) string                { return "{word:" + itoa(id) + "}" }
func plagueTok(id int) string           { return "{plague:" + itoa(id) + "}" }
func warTok(id int) string              { return "{war:" + itoa(id) + "}" }
func elderTok(id int) string            { return "{elder:" + itoa(id) + "}" }
func speciesTok(id int) string          { return "{species:" + itoa(id) + "}" }
func makersTok(l *record.Remain) string { return "{makers:" + itoa(l.ID) + "}" }
func sourceTok(id int) string           { return "{source:" + itoa(id) + "}" }

// A remain's predicates.
func transmitter(l *record.Remain) bool    { return l.Kind == "threat" && l.People < 0 }
func speakingRemain(l *record.Remain) bool { return transmitter(l) && l.State == "unleashed" }

// shipsIn is what is left in a field to fly: nothing at ruin.
func shipsIn(l *record.Remain) int {
	if l.Kind != "field" || l.Cond == "ruin" {
		return 0
	}
	return l.Wrecks + l.Derelicts
}

// fleetsOf lists a people's fleets not over.
func (v *view) fleetsOf(c *record.Civ) []*record.Fleet {
	var out []*record.Fleet
	for _, x := range v.st.Fleets {
		if !x.Over && x.Owner == c.ID {
			out = append(out, x)
		}
	}
	return out
}

// shipsOf is a people's ships in being, its fleets, and its ships laid up.
func (v *view) shipsOf(c *record.Civ) (ships, fleets, laidUp int) {
	for _, x := range v.fleetsOf(c) {
		if x.Ships == 0 {
			continue
		}
		fleets++
		ships += x.Ships
		if x.LaidUp {
			laidUp += x.Ships
		}
	}
	return
}

// gunsOf is a people's guns standing and the worlds they stand over.
func (v *view) gunsOf(c *record.Civ) (guns, worlds int) {
	for _, s := range c.Systems {
		if g := v.st.Stars[s].Guns; g > 0 && v.st.Stars[s].Held == c.ID {
			guns += g
			worlds++
		}
	}
	return
}

// hostsOf lists the living peoples a parasite is in: ridden, or fighting
// its plague.
func (v *view) hostsOf(p *record.Civ) []*record.Civ {
	if p.Own < 0 {
		return nil
	}
	var out []*record.Civ
	for _, c := range v.st.Civs {
		if c != p && living(c) {
			_, sick := c.Infections[p.Own]
			if (c.Master == p.ID && !c.Vassal) || sick {
				out = append(out, c)
			}
		}
	}
	return out
}

// telling is a people's tales as it would tell them: the ones it holds
// dearest, up to a limit, in the order it believes they happened.
func (v *view) telling(c *record.Civ, limit int) []*record.Tale {
	var out []*record.Tale
	for _, t := range v.lore[c.ID] {
		if !t.Forgot && t.Rank > 0 && t.Rank <= limit {
			out = append(out, t)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Rank < out[j].Rank })
	sort.SliceStable(out, func(i, j int) bool { return v.ev[out[i].Fact].Year < v.ev[out[j].Fact].Year })
	return out
}

// perceives says whether a people can hold another in mind.
func (v *view) perceives(c *record.Civ, id int) bool {
	for _, x := range c.Knowledge.Perceives {
		if x == id {
			return true
		}
	}
	return false
}

func has(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func hasKey(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func itoa(n int) string { return strconv.Itoa(n) }

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic("legends: bad id in key: " + s)
	}
	return n
}

func sprintf(format string, args ...any) string { return fmt.Sprintf(format, args...) }
