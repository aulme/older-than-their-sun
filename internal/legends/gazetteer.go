package legends

import (
	"fmt"
	"sort"
	"strings"
)

// gazetteer lists every star the history touched, with its system and
// what happened there. This is the lazy-detail layer's first form: the
// stars that matter get their worlds spelled out.
func (v *view) gazetteer(p func(string, ...any)) {
	st, g := v.st, v.g
	type entry struct {
		star  int
		notes []string
	}
	touched := map[int]*entry{}
	get := func(i int) *entry {
		e := touched[i]
		if e == nil {
			e = &entry{star: i}
			touched[i] = e
		}
		return e
	}
	for _, c := range st.Civs {
		e := get(c.Cradle)
		e.notes = append(e.notes, fmt.Sprintf("cradle of the %s (%s)", tok(c.ID), v.year(c.Born)))
	}
	for i, s := range st.Stars {
		if cid := s.Held; cid >= 0 {
			c := v.civ(cid)
			if i != c.Cradle {
				get(i).notes = append(get(i).notes, fmt.Sprintf("held by the %s", tok(c.ID)))
			}
		}
	}
	for _, c := range st.Civs {
		if !active(c) {
			continue
		}
		for _, s := range c.Claim { // sorted
			if v.held(s) != c.ID {
				get(s).notes = append(get(s).notes, "claimed by the "+tok(c.ID))
			}
		}
	}
	for _, s := range st.Sources {
		if !s.Rarity || s.Star < 0 || s.Legacy >= 0 || s.Mobile {
			continue // the natural rarities at a star; bounties and artifacts list as remains
		}
		note := rarityFrame(s.Key)
		if s.Holder >= 0 && v.held(s.Star) == s.Holder {
			note += ", held by the " + tok(s.Holder)
		}
		get(s.Star).notes = append(get(s.Star).notes, note)
	}
	for _, l := range st.Remains {
		if l.Star < 0 || l.State == "lost" {
			continue
		}
		if l.Maker < 0 {
			get(l.Star).notes = append(get(l.Star).notes, v.describe(l)+" of "+v.elderName(l))
		} else if l.State == "undisturbed" {
			get(l.Star).notes = append(get(l.Star).notes, v.describe(l))
		}
	}
	for _, r := range st.Reservoirs { // by star
		get(r.Star).notes = append(get(r.Star).notes, fmt.Sprintf("dead cities under quarantine: %s waits there until %s", plagueTok(r.Plague), v.year(r.Until)))
	}
	for _, l := range st.Remains {
		if l.Plague >= 0 && l.Star >= 0 && l.State != "lost" {
			get(l.Star).notes = append(get(l.Star).notes, fmt.Sprintf("walls that carry %s", plagueTok(l.Plague)))
		}
	}
	if g.Sol >= 0 {
		get(g.Sol)
	}
	var ids []int
	for i := range touched {
		ids = append(ids, i)
	}
	sort.Slice(ids, func(a, b int) bool {
		da, db := g.FromCentre(ids[a]), g.FromCentre(ids[b])
		if da != db {
			return da < db
		}
		return ids[a] < ids[b] // two stars at one distance: a fixed order, since ids came from a map
	})
	p("=== THE STARS (%d with a history, nearest first) ===", len(ids))
	for _, i := range ids {
		s := &g.Stars[i]
		sys := g.Sys[i]
		head := fmt.Sprintf("%s, %s", v.names.Star(s), s.ClassName())
		if s.Alt != "" {
			head += " (" + s.Alt + ")"
		}
		if i != g.Sol {
			head += fmt.Sprintf(", %.0f ly from %s", g.FromCentre(i), g.Anchor())
		}
		if s.Real && s.Mag < 6.5 && g.Sol >= 0 && i != g.Sol {
			head += fmt.Sprintf(", magnitude %.1f in Earth's sky", s.Mag)
		}
		if s.Real && g.Sol < 0 {
			head += ", a named star"
		}
		p("%s", head)
		p("  %s.", sys.Describe(star(i)))
		if sys.Home >= 0 {
			p("  Habitable: %s, %s.", sys.HomeName(star(i)), archDesc(sys.Arch))
		}
		if st.Stars[i].Bio == "complex" && st.Stars[i].Held < 0 {
			p("  Complex life, unowned.")
		}
		if e := touched[i]; len(e.notes) > 0 {
			p("  %s.", strings.Join(e.notes, "; "))
		}
	}
}

func archDesc(key string) string {
	switch key {
	case "lush":
		return "a temperate, lush world"
	case "ocean":
		return "an ocean world with scattered islands"
	case "arid":
		return "an arid world of salt flats and canyons"
	case "twilight":
		return "tidally locked, habitable along its twilight band"
	case "superterran":
		return "a heavy world of crushing gravity"
	case "lowg":
		return "a small, light world with a thin sky"
	case "hothouse":
		return "a hothouse under a crushing, poisonous sky"
	case "iceshell":
		return "an ocean sealed beneath a shell of ice"
	case "floater":
		return "the cloud decks of a gas giant"
	case "volcanic":
		return "a moon kneaded by tides, all fire and sulphur"
	case "dim":
		return "a dim world close to a brown dwarf"
	}
	return key
}
