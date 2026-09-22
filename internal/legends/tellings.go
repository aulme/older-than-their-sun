package legends

import (
	"fmt"

	"worldgen/internal/record"
)

// tellings prints history as each living people tells it: the tales it
// holds dearest, in the order it believes they happened, with the years
// it believes, the names it uses and the slant it puts on them. Dead
// peoples are printed only in full, as they told it at the end.
func (v *view) tellings(p func(string, ...any), full bool) {
	p("=== AS THEY TELL IT ===")
	shown := 0
	for _, c := range v.st.Civs {
		if !living(c) && !full {
			continue
		}
		limit := 30
		if !living(c) {
			limit = 10
		}
		tales := v.telling(c, limit)
		m := c.Knowledge.Memory
		state := "tell it so"
		if !living(c) {
			state = "told it so, at the end"
		}
		p("The %s %s (%d things held, %d of them myth, %d forgotten, %d retold):", tok(c.ID), state, m.Held, m.Myth, m.Forgot, m.Revised)
		if len(tales) == 0 {
			p("  nothing; they have no story yet")
		}
		for _, t := range tales {
			src := ""
			switch t.Source {
			case "told":
				src = fmt.Sprintf(" [told by the %s]", tok(t.From))
			case "read":
				src = " [read in a ruin]"
			case "inherited":
				src = " [handed down]"
			}
			p("  %-18s %s%s", v.believed(t), v.tell(c, t), src)
		}
		shown++
	}
	if shown == 0 {
		p("  nobody is left to tell it")
	}
}

// believed is the year a people puts on a tale: exact while it is fresh,
// rough once worn, none at all once it is myth. The record carries it.
func (v *view) believed(t *record.Tale) string {
	switch {
	case t.Believed == nil:
		return "long ago"
	case t.Wear == 1:
		return "about " + v.year(*t.Believed)
	}
	return v.year(*t.Believed)
}
