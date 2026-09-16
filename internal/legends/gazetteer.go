package legends

import (
	"fmt"
	"sort"
	"strings"

	"worldgen/internal/history"
)

// gazetteer lists every star the history touched, with its system and
// what happened there. This is the lazy-detail layer's first form: the
// stars that matter get their worlds spelled out.
func gazetteer(p func(string, ...any), w *history.World) {
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
	for _, c := range w.Civs {
		e := get(c.Cradle)
		e.notes = append(e.notes, fmt.Sprintf("cradle of the %s (%s)", c.Name, year(c.Born)))
	}
	for i, cid := range w.Owner {
		if cid >= 0 {
			c := w.Civs[cid]
			if i != c.Cradle {
				get(i).notes = append(get(i).notes, fmt.Sprintf("held by the %s", c.Name))
			}
		}
	}
	for i, hid := range w.Held {
		if hid >= 0 {
			get(i).notes = append(get(i).notes, "held by "+w.Horrors[hid].Name)
		}
	}
	for _, l := range w.Legacies {
		if l.Star < 0 || l.State == history.Lost {
			continue
		}
		if l.Maker < 0 {
			get(l.Star).notes = append(get(l.Star).notes, l.Describe()+" of "+elderName(l))
		} else if l.State == history.Buried {
			get(l.Star).notes = append(get(l.Star).notes, l.Describe())
		}
	}
	if w.G.Sol >= 0 {
		get(w.G.Sol)
	}
	var ids []int
	for i := range touched {
		ids = append(ids, i)
	}
	sort.Slice(ids, func(a, b int) bool { return w.G.FromCentre(ids[a]) < w.G.FromCentre(ids[b]) })
	p("=== THE STARS (%d with a history, nearest first) ===", len(ids))
	for _, i := range ids {
		s := &w.G.Stars[i]
		sys := w.G.Sys[i]
		head := fmt.Sprintf("%s, %s", s.Name, s.ClassName())
		if s.Alt != "" {
			head += " (" + s.Alt + ")"
		}
		if i != w.G.Sol {
			head += fmt.Sprintf(", %.0f ly from %s", w.G.FromCentre(i), w.G.Anchor())
		}
		if s.Real && s.Mag < 6.5 && w.G.Sol >= 0 && i != w.G.Sol {
			head += fmt.Sprintf(", magnitude %.1f in Earth's sky", s.Mag)
		}
		if s.Real && w.G.Sol < 0 {
			head += ", a named star"
		}
		p("%s", head)
		p("  %s.", sys.Describe(s.Name))
		if sys.Home >= 0 {
			p("  Habitable: %s, %s.", sys.HomeName(s.Name), archDesc(sys.Arch))
		}
		if bio := w.Bio[i]; bio == history.BioComplex && w.Owner[i] < 0 {
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
