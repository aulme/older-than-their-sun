package legends

import (
	"fmt"

	"worldgen/internal/history"
)

// tellings prints history as each living people tells it: the tales it
// holds dearest, in the order it believes they happened, with the years
// it believes, the names it uses and the slant it puts on them. Dead
// peoples are printed only in full, as they told it at the end.
func tellings(p func(string, ...any), w *history.World, full bool) {
	p("=== AS THEY TELL IT ===")
	shown := 0
	for _, c := range w.Civs {
		if !c.Living() && !full {
			continue
		}
		limit := 30
		if !c.Living() {
			limit = 10
		}
		tales := w.Telling(c, limit)
		held, myth := 0, 0
		for _, t := range c.Lore {
			if !t.Forgot {
				held++
				if t.Wear >= 2 {
					myth++
				}
			}
		}
		state := "tell it so"
		if !c.Living() {
			state = "told it so, at the end"
		}
		p("The %s %s (%d things held, %d of them myth, %d forgotten, %d retold):", c.Name, state, held, myth, c.Tally.Forgot, c.Tally.Revised)
		if len(tales) == 0 {
			p("  nothing; they have no story yet")
		}
		for _, t := range tales {
			src := ""
			switch t.Source {
			case history.Told:
				src = fmt.Sprintf(" [told by the %s]", w.Civs[t.From].Name)
			case history.Read:
				src = " [read in a ruin]"
			case history.Inherited:
				src = " [handed down]"
			}
			p("  %-18s %s%s", believed(w, t), w.Tell(c, t), src)
		}
		shown++
	}
	if shown == 0 {
		p("  nobody is left to tell it")
	}
}

// believed is the year a people puts on a tale: exact while it is fresh,
// rough once worn, none at all once it is myth.
func believed(w *history.World, t *history.Tale) string {
	f := w.Facts[t.Fact]
	switch t.Wear {
	case 0:
		return year(f.Year)
	case 1:
		y := f.Year - present
		r := history.Year(100_000)
		y = (y - r/2) / r * r
		return "about " + year(y+present)
	}
	return "long ago"
}
