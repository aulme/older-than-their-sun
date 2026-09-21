package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/galaxy"
)

// Rarities: a source with a grant, levels or reach, that a people has or
// does not have. Having it means holding or being based at its star, holding
// a star within its radius, holding it as a mobile thing, or a partner
// holding an immobile one. One instance is enough; a second is worth nothing
// to the same people, which is what makes them worth trading. A grant
// halves the price of a node and never bars it: no node is impossible
// without a rarity, and tech.TestNoCatch22 keeps it so.
//
// The natural ones are placed at world generation from what the field
// holds: dead stars, the catalogued things whose reach covers the field,
// companions and the odd luxury. The elder ones are bounties, legacies of a
// new kind found like the rest. A wielded artifact is a mobile rarity.

// magnetarReach is how far a field magnetar's beam counts as near, in
// light years; a catalogued one has its own radius.
const magnetarReach = 30

// naturalRarities places the rarities of a galaxy. It reads only the stars,
// the systems and the catalogue, and draws nothing.
func naturalRarities(g *galaxy.Galaxy) []*Source {
	var out []*Source
	at := func(i int, key, name string, s Source) {
		s.Key, s.Name, s.Star, s.Kind = key, name, i, CosmicSource
		s.Rarity = s.Yield == (flow.Income{}) // a companion burned for fuel is a source like any other
		out = append(out, &s)
	}
	for i := range g.Stars {
		st := &g.Stars[i]
		sys := g.Sys[i]
		name := "{star:" + itoa(i) + "}"
		switch {
		case st.Class == 'N' && (st.Remnant == "black hole" || st.Remnant == "the great hole"):
			at(i, "horizon", "the horizon at "+name, Source{Grants: []string{"causal_physics", "deep_time"}})
		case st.Class == 'N' && st.Remnant == "magnetar":
			at(i, "beam", "the beam of "+name, Source{Radius: magnetarReach, Grants: []string{"unmaking"}, Levels: [3]float64{0.5, 0, 0}})
		case st.Class == 'N':
			at(i, "heavy_star", "the heavy star at "+name, Source{Grants: []string{"stellar_weapons"}})
		case st.Class == 'W':
			at(i, "diamond", "the diamond star at "+name, Source{Levels: [3]float64{0, 0, 0.5}})
		}
		switch sys.Comp {
		case "a white dwarf companion":
			at(i, "dwarf_companion", "the small white companion of "+name, Source{Levels: [3]float64{0, 0, 0.3}})
		case "a brown dwarf companion":
			at(i, "brown_companion", "the brown dwarf beside "+name, Source{Yield: flow.Income{flow.E: 2}, Needs: []string{"fusion"}, Rarity: false})
		}
		if sys.Disc != "" {
			at(i, "dust", "the dust of "+name, Source{Levels: [3]float64{0, 0, 0.3}})
		}
		if sys.Home >= 0 && sys.Planets[sys.Home].Moons > 0 {
			at(i, "moon", "the moon of "+name, Source{Levels: [3]float64{0, 0, 0.3}})
		}
	}
	ranged := func(f *galaxy.Feature, key string, s Source) {
		s.Key, s.Name, s.Star, s.Feature, s.Radius, s.Kind, s.Rarity = key, f.Name, -1, f, f.Radius, CosmicSource, true
		out = append(out, &s)
	}
	// a catalogued black hole is a horizon only as the star it is, which
	// the field holds when it is anchored on it; its radius is the reach of
	// its laws, not of its grant. The Heart reaches.
	for _, f := range galaxy.Features {
		switch f.Kind {
		case galaxy.BlackHole:
			if f.Name == "Sagittarius A*" {
				ranged(f, "heart", Source{Grants: []string{"transcendence"}})
			}
		case galaxy.Magnetar:
			ranged(f, "beam", Source{Grants: []string{"unmaking"}, Levels: [3]float64{0.5, 0, 0}})
		case galaxy.NeutronStar:
			ranged(f, "beacon", Source{Reach: 3})
		case galaxy.Nebula:
			ranged(f, "colours", Source{Levels: [3]float64{0, 0, 0.3}})
		case galaxy.Remnant:
			ranged(f, "ash", Source{Grants: []string{"exotic_matter"}})
		}
	}
	return out
}

// rarityLines are what a people says when it first has a rarity, by key.
// %s is the people; a second %s, where there is one, is the source's name.
var rarityLines = map[string]string{
	"horizon":         "The %s hold %s now. What falls in comes out as understanding: the deep physics come cheap to them.",
	"beam":            "The %s live under %s. Its flares light their sky, and they learn from it how to unmake.",
	"heavy_star":      "The %s hold %s. Its skin is metal a mile deep, and its weight is a lesson in stellar weapons.",
	"beacon":          "The %s hold a star within reach of %s. They steer by its ticking, and their ships go further for it.",
	"diamond":         "The %s hold %s. It is a diamond the size of a world, and they are never done singing about it.",
	"colours":         "The %s live in %s. The sky is a painting, and the painting hides them.",
	"ash":             "The %s hold a star in %s. The matter there was made in a death, and some of it is not on any table.",
	"heart":           "The %s hold a star within reach of %s. Everything falls toward it, and so does thought.",
	"dwarf_companion": "The %s hold %s. It is a thing of great value that does nothing, which is what makes it valuable.",
	"dust":            "The %s hold %s. At dusk the whole sky is a ring, and they put it on their flags.",
	"moon":            "The %s hold %s. The tides and the calendar are its, and the songs.",
}

// rarityFrames are what a rarity is called in a gazetteer note.
var rarityFrames = map[string]string{
	"horizon": "a horizon", "beam": "the beam", "heavy_star": "the heavy star", "beacon": "a beacon", "diamond": "the diamond star",
	"colours": "the colours", "ash": "the ash", "heart": "the Heart", "dwarf_companion": "a white companion", "brown_companion": "a brown dwarf",
	"dust": "a ring of dust", "moon": "a moon of its own",
}

// RarityFrame is what a rarity of a key is called in a note.
func RarityFrame(key string) string {
	if f, ok := rarityFrames[key]; ok {
		return f
	}
	return key
}

// GrantedNodes lists every node some natural rarity grants, for the batch.
var GrantedNodes = []string{"causal_physics", "deep_time", "unmaking", "stellar_weapons", "exotic_matter", "transcendence"}

// harnessLines are what a people says when it first harnesses a source
// kind worth a line. %s is the people; the second %s the source's name.
var harnessLines = map[string]string{
	"belt":            "The %s begin to work %s: ice and iron, by the shipload.",
	"brown_companion": "The %s burn %s for fuel: a star that never lit, lit at last.",
	"giant":           "The %s skim %s for fuel. The small suns of their fusion plants never go out now.",
	"terraformed":     "The %s bring %s to life. It feeds them.",
	"heavy":           "The %s dig %s and find it rich in what splits.",
	"nebula":          "The %s grow their food in the gas of %s itself.",
	"doomed_giant":    "The %s catch the light of %s. It will not shine long, and they know it.",
	"comets":          "The %s harvest %s, and eat ice older than their sun.",
}

// firstHarness is the line and the deed for the first source of a kind a
// people harnesses; the kinds with no line in the table pass in silence.
// A structure announces itself when it is raised; see raise.
func (w *World) firstHarness(c *Civ, s *Source) {
	if c.Harnessed == nil {
		c.Harnessed = map[string]bool{}
	}
	if c.Harnessed[s.Key] {
		return
	}
	c.Harnessed[s.Key] = true
	if line, ok := harnessLines[s.Key]; !ok || line == "" {
		return
	}
	w.fact(FHarness, c, nil, max(s.Star, c.Home)).with(P{"source": s.ID})
}

// had is one rarity a people has, and the partner it has it through, or -1.
type had struct {
	s   *Source
	via int
}

// rarities lists what a people has, each key once: the rarities at its
// holdings and within their reach, the mobile things it holds, and its
// partners' immobile ones that grant or reach, since partners share access
// to what stands still and is worth a war, grants and levels only; the
// yield goes through the surplus, and a luxury is the sky its holder was
// born under.
func (w *World) rarities(c *Civ) []had {
	seen := map[string]bool{}
	var out []had
	take := func(s *Source, via int) {
		if !s.Rarity || seen[s.Key] {
			return
		}
		seen[s.Key] = true
		out = append(out, had{s, via})
	}
	for _, h := range w.holdings(c) {
		for _, id := range w.sourcesAt[h] {
			take(w.Sources[id], -1)
		}
	}
	for _, id := range w.mobile {
		if s := w.Sources[id]; s.Holder == c.ID {
			take(s, -1)
		}
	}
	for _, pid := range sortedInts(c.Trade) {
		p := w.Civs[pid]
		if !p.Active() {
			continue
		}
		for _, h := range w.holdings(p) {
			for _, id := range w.sourcesAt[h] {
				if s := w.Sources[id]; !s.Mobile && (len(s.Grants) > 0 || s.Reach > 0) {
					take(s, pid) // what grants or reaches; a partner's moon is its own
				}
			}
		}
	}
	for _, a := range w.accessed(c) {
		if w.Owner[a.s.Star] == a.via {
			take(a.s, a.via) // bought the use of
		}
	}
	return out
}

// rare is the rarity walk each tick: what a people has now, by key, and
// the nodes those grant. It marks the holder on the immobile rarities at
// its holdings, logs each kind the first time it is had, and says whether
// the set changed, so the levels are recomputed.
func (w *World) rare(c *Civ) bool {
	now := map[string]bool{}
	grants := map[string]bool{}
	var firsts []had
	for _, x := range w.rarities(c) {
		if !c.Had[x.s.Key] {
			firsts = append(firsts, x)
		}
		now[x.s.Key] = true
		for _, k := range x.s.Grants {
			grants[k] = true
		}
	}
	for _, h := range w.holdings(c) {
		for _, id := range w.sourcesAt[h] {
			if s := w.Sources[id]; s.Rarity && s.Star == h && !s.Mobile {
				s.Holder = c.ID
			}
		}
	}
	changed := len(now) != len(c.Rare)
	for k := range now {
		if !c.Rare[k] {
			changed = true
		}
	}
	c.Rare, c.Grants = now, grants
	if c.Had == nil {
		c.Had = map[string]bool{}
	}
	for _, x := range firsts {
		s := x.s
		c.Had[s.Key] = true
		// a luxury at the cradle is the sky they were born under; what
		// grants or reaches, or is come upon elsewhere, is worth a line
		matters := len(s.Grants) > 0 || s.Reach > 0
		switch {
		case x.via >= 0 && matters:
			w.event(KRarityHad, c, w.Civs[x.via], -1, P{"source": s.ID, "via": true})
		case x.via < 0:
			if line := rarityLines[s.Key]; line != "" && (s.Star != c.Cradle || matters) {
				w.event(KRarityHad, c, nil, -1, P{"source": s.ID, "via": false})
			}
		}
	}
	return changed
}

// has says whether a people has a rarity of a key this tick.
func (c *Civ) has(key string) bool { return c.Rare[key] }

// rarityAt is a rarity of a key at a star, or nil.
func (w *World) rarityAt(star int, key string) *Source {
	for _, id := range w.sourcesAt[star] {
		if s := w.Sources[id]; s.Key == key && s.Star == star {
			return s
		}
	}
	return nil
}

// levelsFromRarities sums what the rarities had give: levels and reach.
func (w *World) levelsFromRarities(c *Civ) (mil, sur, soc, reach float64) {
	for _, x := range w.rarities(c) {
		s := x.s
		mil, sur, soc, reach = mil+s.Levels[0], sur+s.Levels[1], soc+s.Levels[2], reach+s.Reach
	}
	return
}

// Mobility. A mobile rarity sits at a star of its holder's, or aboard a
// fleet. It changes hands when its star is taken, is buried where a fleet
// that carried it died, and rides with a nomad's greatest fleet.

// mobileHeld lists a people's mobile rarities at a star.
func (w *World) mobileHeld(c *Civ, star int) []*Source {
	var out []*Source
	for _, id := range w.mobile {
		if s := w.Sources[id]; s.Holder == c.ID && s.Star == star && s.Carried < 0 {
			out = append(out, s)
		}
	}
	return out
}

// carryOff is a star's mobile rarities changing hands: aboard the fleet
// that took the star, or straight home when the front took it.
func (w *World) carryOff(c, e *Civ, star int, x *Expedition) {
	for _, s := range w.mobileHeld(e, star) {
		if w.taken(c, e, s, star) {
			continue // it did not survive the taking
		}
		w.transfer(s, e, c)
		if x != nil {
			s.Carried, s.Star = x.ID, -1
			w.event(KCarriedOff, c, e, -1, P{"source": s.ID, "fleet": x.ID})
		} else {
			s.Star = c.Home
			w.event(KCarriedOff, c, e, c.Home, P{"source": s.ID, "fleet": -1})
		}
	}
}

// transfer moves a mobile rarity, and the remain it is, from one holder
// to another, or to nobody. An object lost is remembered by its loser,
// who may make another after a while; an object moved pays its form's
// price.
func (w *World) transfer(s *Source, from, to *Civ) {
	if from != nil && s.Legacy >= 0 {
		keep := from.Wielded[:0]
		for _, l := range from.Wielded {
			if l.ID != s.Legacy {
				keep = append(keep, l)
			}
		}
		from.Wielded = keep
	}
	if from != nil && s.Form != "" {
		if from.Remade == nil {
			from.Remade = map[string]Year{}
		}
		from.Remade[s.Key] = w.Now
	}
	s.Holder = -1
	if to == nil {
		return
	}
	s.Holder = to.ID
	if s.Legacy >= 0 {
		l := w.Legacies[s.Legacy]
		l.Finder = to.ID
		l.State = Wielded
		held := false
		for _, x := range to.Wielded {
			if x == l {
				held = true
			}
		}
		if !held {
			to.Wielded = append(to.Wielded, l)
		}
	}
	w.moved(to, s)
}

// bury leaves a mobile rarity where it is, for somebody to find.
func (w *World) bury(s *Source, from *Civ, star int) {
	w.transfer(s, from, nil)
	s.Star, s.Carried = star, -1
	if s.Legacy >= 0 {
		l := w.Legacies[s.Legacy]
		l.State, l.Star = Buried, star
	}
}

// dropRarities buries a people's mobile rarities at a star it is losing
// to anything but a conqueror, who carries them off instead.
func (w *World) dropRarities(c *Civ, star int) {
	for _, s := range w.mobileHeld(c, star) {
		w.bury(s, c, star)
		w.event(KLeftBehind, c, nil, star, P{"source": s.ID})
	}
}

// carriedBy lists what a fleet carries.
func (w *World) carriedBy(x *Expedition) []*Source {
	var out []*Source
	for _, id := range w.mobile {
		if s := w.Sources[id]; s.Carried == x.ID {
			out = append(out, s)
		}
	}
	return out
}

// land is a fleet coming home: what it carried sits at the star it landed at.
func (w *World) land(x *Expedition, star int) {
	for _, s := range w.carriedBy(x) {
		s.Carried, s.Star = -1, star
	}
}

// fleetLost is a fleet dying where it is: what it carried is buried there.
func (w *World) fleetLost(x *Expedition, star int) {
	c := w.Civs[x.Owner]
	for _, s := range w.carriedBy(x) {
		w.bury(s, c, star)
		w.event(KWentDown, c, nil, star, P{"source": s.ID, "fleet": x.ID})
	}
}

// stow puts a nomad people's mobile rarities aboard its greatest fleet.
func (w *World) stow(c *Civ, best *Expedition) {
	for _, id := range w.mobile {
		if s := w.Sources[id]; s.Holder == c.ID && s.Carried != best.ID {
			s.Carried, s.Star = best.ID, -1
			w.moved(c, s)
		}
	}
}

// wieldRarity registers a wielded artifact as a mobile rarity of its
// finder, with the levels the artifact gives, at the finder's seat.
func (w *World) wieldRarity(c *Civ, l *Legacy) {
	var s *Source
	if l.Source >= 0 {
		s = w.Sources[l.Source]
	} else {
		s = w.addSource(&Source{Key: "artifact", Name: l.Desc, Kind: ElderSource, Star: c.Home, Mobile: true, Rarity: true, Legacy: l.ID, Holder: -1, Carried: -1})
		if l.Maker >= 0 {
			s.Kind = MadeSource
		}
		l.Source = s.ID
	}
	switch l.Level {
	case "mil":
		s.Levels = [3]float64{2, 0, 0}
	case "sur":
		s.Levels = [3]float64{0, 1.5, 0}
	case "soc":
		s.Levels = [3]float64{0, 0, 1.5}
	case "reach":
		s.Reach = 15
	case "all":
		s.Levels = [3]float64{1, 1, 1}
	}
	s.Holder, s.Star, s.Carried = c.ID, c.Home, -1
	if c.Aloft {
		if best := w.greatestFleet(c); best != nil {
			s.Carried, s.Star = best.ID, -1
		}
	}
}

// bounties are the elder rarities: things still doing what they were made
// to do, for whoever puts them to use. Immobile; the yield goes to whoever
// holds the star once somebody has worked out how.
var bounties = []struct {
	Key, Desc string
	Yield     flow.Income
	Levels    [3]float64
	Wear      Year
}{
	{"lattice", "a lattice that turns starlight to metal, still turning", flow.Income{flow.M: 6}, [3]float64{}, 0},
	{"sea", "a sea that has been growing since before the age", flow.Income{flow.O: 6}, [3]float64{}, 0},
	{"battery", "a battery the size of a moon, a tenth full", flow.Income{flow.E: 6}, [3]float64{}, 1_000_000},
	{"seam", "a seam of a metal that is not on the table", flow.Income{flow.M: 4}, [3]float64{0.5, 0, 0}, 0},
	{"garden", "a garden that tends itself, under a roof of something clear", flow.Income{flow.O: 4}, [3]float64{0, 0.5, 0}, 0},
	{"mirror", "a mirror in orbit that never lost its polish", flow.Income{flow.E: 4}, [3]float64{}, 0},
	{"ring", "a ring of black metal around a dead star, still warm", flow.Income{flow.E: 8}, [3]float64{}, 0},
}

// leaveBounty makes an elder legacy a bounty with its source.
func (w *World) leaveBounty(l *Legacy, i int) {
	b := bounties[i]
	l.Kind = Bounty
	l.Desc = b.Desc
	s := w.addSource(&Source{Key: "bounty:" + b.Key, Name: b.Desc, Kind: ElderSource, Star: l.Star, Yield: b.Yield, Levels: b.Levels, Rarity: true, Legacy: l.ID, Holder: -1, Carried: -1, Wear: b.Wear})
	l.Source = s.ID
}

// useBounty is a people putting a bounty to use: it yields to the star's
// holder from now on.
func (w *World) useBounty(c *Civ, l *Legacy) {
	l.State = Wielded
	l.Finder = c.ID
	s := w.Sources[l.Source]
	if s.Since == 0 {
		s.Since = w.Now
	}
}

// capital upper-cases the first letter.
func capital(s string) string {
	if s == "" {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-'a'+'A') + s[1:]
	}
	return s
}
