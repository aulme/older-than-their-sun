package history

import (
	"fmt"
	"math"
	"strings"

	"worldgen/internal/galaxy"
)

// Geography: what the place in the galaxy does to a history. The laws are
// computed once for the field (see galaxy.Law) and read here: life arises
// at the law's rate, the sky burns at the law's rate, the deep physics
// comes easier where there are dead stars to study, and near the centre
// the great hole flares.

// flares: near the centre the hole wakes now and then, and the sky burns
// over the whole field at once. The rate follows the glare.
func (w *World) flares() {
	if w.Law.Glare < 4 {
		return
	}
	if !w.chance(0.00002 * w.Law.Glare) {
		return
	}
	w.log("The heart of the galaxy flares. For a century the sky is white, and every world in the field turns its face away.")
	origin := 0
	w.blastAll("the flaring of the centre", -1.5)
	_ = origin
}

// blastAll applies a cosmic filter to every civilisation in the field.
func (w *World) blastAll(what string, adj float64) {
	for _, c := range w.Civs {
		if !c.Living() {
			continue
		}
		for _, s := range c.Systems {
			if w.Bio[s] == BioComplex && w.R.Float64() < 0.3 {
				w.Bio[s] = BioSimple
			}
		}
		if !c.Active() {
			continue
		}
		w.blastWorlds = append([]int(nil), c.Systems...)
		w.blastWhat = what
		w.face(c, "cosmic", adj)
	}
}

// skyEvents adds the real sky's dated events, once the present is known:
// the supernovae and flares that Earth's sky recorded or that left their
// marks nearby. They are flavour, dated from the present, and only for a
// field that could have seen them.
func (w *World) skyEvents() {
	for _, f := range galaxy.Features {
		if f.When == 0 || f.Event == "" {
			continue
		}
		d := f.Pos.Dist(w.G.Region.Pos)
		far := 7.0 // kpc: a new star in the sky
		if f.Kind == galaxy.Magnetar {
			far = 15 // a giant flare is felt across the galaxy
		}
		if d > far {
			continue
		}
		y := w.Present - Year(f.When)
		if y < w.Cfg.Dawn {
			continue
		}
		w.Events = append(w.Events, Event{Year: y, Text: f.Event})
	}
}

// lawDiff bends a filter's difficulty by the place: a people born under a
// hard sky is built for it, and the burning sky is a smaller thing to them.
func (w *World) lawDiff(key string) float64 {
	if key == "cosmic" && w.Law.Glare > 2 {
		return -min(1.5, 0.4*math.Log2(w.Law.Glare))
	}
	return 0
}

// starDetail names a star with its class and what Earth knows of it.
func (w *World) starDetail(id int) string {
	s := &w.G.Stars[id]
	d := fmt.Sprintf("%s (%s", w.star(id), s.ClassName())
	if s.Real && s.Alt != "" {
		d += ", " + s.Alt
	}
	if s.Real && s.Mag < 6 && w.G.Sol >= 0 {
		d += fmt.Sprintf(", a naked-eye star from Earth")
	}
	return d + ")"
}

// systemLine describes a star's system for the legends.
func (w *World) systemLine(id int) string {
	sys := w.G.Sys[id]
	name := w.star(id)
	line := "  " + name + ": " + sys.Describe(name) + "."
	if sys.Missed {
		line += " The home world is one Earth's surveys never saw."
	}
	return line
}

// lawsInWords tells the laws of the place as a people living there would.
func LawsInWords(g *galaxy.Galaxy) []string {
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

// NearFeatures lists the catalogued things near the field, nearest first.
func NearFeatures(g *galaxy.Galaxy, max int) []string {
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

// SkyFeatures names what lies beyond the galaxy.
func SkyFeatures(g *galaxy.Galaxy) string {
	var names []string
	for _, nf := range g.Law.Near {
		if nf.F.Kind == galaxy.Sky {
			names = append(names, nf.F.Name)
		}
	}
	return strings.Join(names, ", ")
}
