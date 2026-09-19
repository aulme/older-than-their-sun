package history

import (
	"math"

	"worldgen/internal/mind"
	"worldgen/internal/tech"
)

// Sightings. A fleet in flight is a point on a line at a known speed, and
// every people has eyes: its worlds, the works it holds at any star, its
// fleets at base and in flight, its pickets. Each eye has a radius, wider
// for a bigger fleet and for a torch. The sighting is computed on the
// fleet's timetable when it launches, not at tick boundaries: for every
// eye, the year the line enters its sphere, and the earliest per people
// is queued as an event at that year. A crossing shorter than a tick is
// seen like any other. What a sighting does: the will and the call to
// allies as before, a report to pact members, and the council's choice
// to meet the fleet in the dark (intercept.go).

// Eye kinds, for the telling and the batch.
const (
	eyeWorld  = "world"
	eyeWorks  = "works"
	eyeFleet  = "fleet"
	eyePicket = "picket"
	eyeSight  = "sight"
)

// How far a fleet is seen: the eye's radius times the fleet's size,
// sqrt(ships / sizeDiv) clamped, doubled for a torch.
const (
	sizeDiv    = 2.0
	sizeMin    = 0.5
	sizeMax    = 3.0
	torchSpeed = 4.0 // years per light year at or below which a drive is a torch
	fleetEye   = 0.5 // a ship's eye is half the tree's range, at least this
	sightNoise = 0.4 // the noise on the level seen from a fleet's drive and hull
	adriftFind = 0.5 // a line this close to a field adrift finds it
)

// eye is one thing that watches the sky for a people.
type eye struct {
	star  int         // where it is; -1 for a fleet in flight
	r     float64     // its plain radius in light years
	fleet *Expedition // a fleet in flight, moving with it
	kind  string
	name  string // the structure's name when a work is the eye
}

// Sighting is one people's sight of one fleet in flight.
type Sighting struct {
	Fleet, Owner, Seer int
	Kind               ExpKind
	From, Star         int
	Launched, Arrive   Year
	Leg                Year    // the fleet's Launched when seen: the leg it was on
	Year               Year    // when it was seen
	Ships              int     // exact: a few years of watching give the count
	Mil                float64 // the level seen, with noise
	Speed              float64 // years per light year, exact
	Eye                string  // the kind of eye
	EyeName            string  // the work, for the telling
	EyeStar            int     // where the eye was; -1 for a fleet
	Feasible           bool    // a meeting point existed against a fleet worth meeting
	Intercept          int     // the interceptor sent, or -1
	Offered            bool    // put up for sale, or judged not for sale; see contract.go
}

// watchAt is how far a people sees fleets from a star: the tree's range
// at a holding, and the widest work standing there, whoever holds it.
func (w *World) watchAt(c *Civ, star int) float64 {
	r := 0.0
	if contains(w.holdings(c), star) {
		r = c.watchRange()
	}
	for _, wk := range c.Works {
		if wk.Star == star && !wk.Dark {
			r = max(r, tech.Structures[wk.Key].Watch)
		}
	}
	return r
}

// eyes lists everything that watches the sky for a people: each holding
// at the tree's range or the widest work there, each work elsewhere at
// its own, each fleet at base or in flight and each picket at half the
// tree's range, at least half a light year.
func (w *World) eyes(o *Civ) []eye {
	var out []eye
	base := o.watchRange()
	seen := map[int]bool{}
	worksEye := func(s int, e eye) eye {
		for _, wk := range o.Works {
			if wk.Star != s || wk.Dark {
				continue
			}
			if st := tech.Structures[wk.Key]; st.Watch > e.r {
				e.r, e.kind, e.name = st.Watch, eyeWorks, st.Name
			}
		}
		return e
	}
	for _, s := range w.holdings(o) {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, worksEye(s, eye{star: s, r: base, kind: eyeWorld}))
	}
	for _, wk := range o.Works {
		if seen[wk.Star] {
			continue
		}
		seen[wk.Star] = true
		if e := worksEye(wk.Star, eye{star: wk.Star, kind: eyeWorks}); e.r > 0 {
			out = append(out, e)
		}
	}
	fr := max(fleetEye, base/2)
	for _, x := range w.fleetsOf(o) {
		if x.Ships <= 0 || x.LaidUp {
			continue
		}
		switch {
		case x.Base >= 0:
			kind := eyeFleet
			if x.Kind == Scout && x.Picket {
				kind = eyePicket
			}
			out = append(out, eye{star: x.Base, r: fr, kind: kind})
		case x.Kind == Survey || x.Kind == Scout:
			// a ship of a few, on its way: it looks where it is going
		default:
			out = append(out, eye{star: -1, r: fr, fleet: x, kind: eyeFleet})
		}
	}
	return out
}

// seenRadius is an eye's radius against a fleet: scaled by the fleet's
// size, doubled for a torch.
func (w *World) seenRadius(r float64, x *Expedition) float64 {
	r *= clamp(math.Sqrt(float64(x.Ships)/sizeDiv), sizeMin, sizeMax)
	if w.driveOf(x) <= torchSpeed {
		r *= 2
	}
	return r
}

// driveOf is a fleet's speed in years per light year: what it sailed with,
// or its owner's today.
func (w *World) driveOf(x *Expedition) float64 {
	if x.Drive > 0 {
		return x.Drive
	}
	return w.Civs[x.Owner].Speed
}

// watched says whether a fleet is the kind that is seen and met: a
// campaign, a relief, a horde's hop or an interceptor, outbound.
func watched(x *Expedition) bool {
	if x.Over || x.Base >= 0 || x.Returning || x.Ships <= 0 {
		return false
	}
	return x.Kind == Campaign || x.Kind == Relief || x.Kind == Roam || x.Kind == Intercept
}

// posAtF is where a fleet is at a fractional year, clamped to its leg.
func (w *World) posAtF(x *Expedition, t float64) vec {
	a, b := w.line(x)
	if x.Arrive <= x.Launched {
		return b
	}
	return a.lerp(b, clamp((t-float64(x.Launched))/float64(x.Arrive-x.Launched), 0, 1))
}

// entryYear is when an eye first sees a fleet on its current leg, if it
// does. A fleet in flight as an eye is taken over the years both are in
// flight, where it will be then.
func (w *World) entryYear(x *Expedition, e eye) (Year, bool) {
	r := w.seenRadius(e.r, x)
	t0, t1 := float64(x.Launched), float64(x.Arrive)
	var e0, ve vec
	if e.fleet != nil {
		f := e.fleet
		t0, t1 = max(t0, float64(f.Launched)), min(t1, float64(f.Arrive))
		if t1 < t0 {
			return 0, false
		}
		e0, ve = w.posAtF(f, t0), w.velocity(f)
	} else {
		e0 = w.pos(e.star)
	}
	t, ok := entry(w.posAtF(x, t0), w.velocity(x), e0, ve, t0, t1, r)
	if !ok {
		return 0, false
	}
	return Year(math.Ceil(t - 1e-6)), true
}

// firstEntry is the earliest an eye of a people sees a fleet, and which.
func (w *World) firstEntry(x *Expedition, o *Civ) (Year, eye, bool) {
	var best Year
	var be eye
	ok := false
	for _, e := range w.eyes(o) {
		if y, hit := w.entryYear(x, e); hit && (!ok || y < best) {
			best, be, ok = y, e, true
		}
	}
	return best, be, ok
}

// timetable is run when a fleet sets out on a leg: whatever adrift its
// line passes is found, and for every other people the year its eyes
// first see the fleet is queued as a sighting. The Sight's holder sees a
// fleet bound for its worlds as it leaves.
func (w *World) timetable(x *Expedition) {
	w.adriftFinds(x)
	if !watched(x) {
		return
	}
	c := w.Civs[x.Owner]
	for _, o := range w.Civs {
		if !o.Active() || o.ID == c.ID {
			continue
		}
		if o.miracle("foresight") && !o.Searching && w.Owner[x.Star] == o.ID {
			w.queueSighting(x, o, x.Launched, eye{star: x.Star, kind: eyeSight})
			continue
		}
		if y, e, ok := w.firstEntry(x, o); ok {
			w.queueSighting(x, o, y, e)
		}
	}
}

// newEye is an eye appearing after fleets have launched: a fleet setting
// out or a picket at its post checks every fleet in flight.
func (w *World) newEye(o *Civ, e eye) {
	for _, x := range w.liveFleets() {
		if !watched(x) || x.Owner == o.ID {
			continue
		}
		if y, ok := w.entryYear(x, e); ok {
			w.queueSighting(x, o, y, e)
		}
	}
}

// queueSighting queues a sighting at a year, once per fleet, seer and leg.
func (w *World) queueSighting(x *Expedition, o *Civ, y Year, e eye) {
	for _, s := range w.pending {
		if s.Fleet == x.ID && s.Seer == o.ID && s.Leg == x.Launched {
			return
		}
	}
	if s := o.Sightings[x.ID]; s != nil && s.Leg == x.Launched {
		return
	}
	w.pending = append(w.pending, &Sighting{
		Fleet: x.ID, Owner: x.Owner, Seer: o.ID, Kind: x.Kind, From: x.From, Star: x.Star,
		Launched: x.Launched, Arrive: x.Arrive, Leg: x.Launched, Year: y,
		Eye: e.kind, EyeName: e.name, EyeStar: e.star, Intercept: -1,
	})
}

// sightingsDue resolves every sighting due before the next tick, in year
// order, launches included: an interceptor sent is itself seen.
func (w *World) sightingsDue() {
	until := w.Now + w.Cfg.Step
	for {
		bi := -1
		for i, s := range w.pending {
			if s.Year >= until {
				continue
			}
			if bi < 0 || s.Year < w.pending[bi].Year || (s.Year == w.pending[bi].Year && (s.Fleet < w.pending[bi].Fleet || (s.Fleet == w.pending[bi].Fleet && s.Seer < w.pending[bi].Seer))) {
				bi = i
			}
		}
		if bi < 0 {
			return
		}
		s := w.pending[bi]
		w.pending = append(w.pending[:bi], w.pending[bi+1:]...)
		w.sighted(s)
	}
}

// sighted is a sighting happening: the seer holds it, the fleet's owner
// does not know, and the seer's council meets at once.
func (w *World) sighted(s *Sighting) {
	x := w.Expeditions[s.Fleet]
	o, c := w.Civs[s.Seer], w.Civs[x.Owner]
	if x.Over || x.Base >= 0 || x.Launched != s.Leg || !o.Active() || !c.Living() {
		return
	}
	s.Ships = x.Ships
	s.Mil = c.Mil + w.R.NormFloat64()*sightNoise
	s.Speed = w.driveOf(x)
	if o.Sightings == nil {
		o.Sightings = map[int]*Sighting{}
	}
	o.Sightings[x.ID] = s
	o.Tally.Sightings++
	w.Watch = append(w.Watch, s)
	if warn := x.Arrive - s.Year; warn > x.Warning {
		x.Warning = warn
	}
	if !x.Seen[o.ID] {
		x.Seen[o.ID] = true
		w.fleetSeen(x, o, s)
	}
	w.forwardSighting(o, s)
	w.considerIntercept(o, s)
}

// fleetSeen is what a people does with a fleet sighted, the first time:
// a fleet coming for it is a line, a focus, a summons, will and a call
// to allies; relief bound to an enemy is counted in the sky it goes to.
func (w *World) fleetSeen(x *Expedition, o *Civ, s *Sighting) {
	c := w.Civs[x.Owner]
	switch {
	case x.Kind == Campaign && x.Target == o.ID:
		out := span(x.Arrive - s.Year)
		switch s.Eye {
		case eyeWorks:
			w.logAt(s.Year, "From the %s at %s the %s see the fleet of the %s coming, %s out.", s.EyeName, w.star(s.EyeStar), o.Name, c.Name, out)
		case eyeFleet:
			w.logAt(s.Year, "A fleet of the %s in flight sees the fleet of the %s coming toward %s, %s out.", o.Name, c.Name, w.star(x.Star), out)
		case eyePicket:
			w.logAt(s.Year, "The pickets of the %s see the fleet of the %s coming, %s out.", o.Name, c.Name, out)
		default:
			w.logAt(s.Year, "The %s see the fleet of the %s coming, %s out.", o.Name, c.Name, out)
		}
		o.Focus[tech.Weapons] = max(o.Focus[tech.Weapons], 3)
		o.Summoned = true
		if wr := w.warBetween(c.ID, o.ID); wr != nil {
			wr.Will[wr.side(o.ID)] += 0.5
			w.callAllies(o, c, wr)
		}
	case x.Kind == Relief && x.Target >= 0 && o.Wars[x.Target]:
		h := w.Civs[x.Target]
		i := w.observe(o, h, x.Star, 0.5)
		i.Relief += float64(x.Ships)
	}
}

// forwardSighting passes a sighting to pact members by the rule reports
// go by, and always to a member whose world the fleet is bound for.
func (w *World) forwardSighting(o *Civ, s *Sighting) {
	c := w.Civs[s.Owner]
	for _, pid := range o.Pacts {
		p := w.Pacts[pid]
		if p.Over {
			continue
		}
		for _, mid := range p.Members {
			m := w.Civs[mid]
			if m == o || !m.Active() {
				continue
			}
			ok := w.Owner[s.Star] == m.ID
			if !ok {
				in := mind.ForwardInput{AllyAtWar: m.Wars[c.ID] || p.Target == c.ID, AllyMet: m.Met[c.ID], Hostile: o.hostile(), Reported: s.Mil, Mil: o.Mil, Dials: o.Dials}
				ok, _ = mind.Forward(in, w.Cfg.Tuning)
			}
			if ok {
				w.send(o, m, &Message{Kind: MsgSighting, About: c.ID, Target: s.Fleet, Sighting: s})
			}
		}
	}
}

// receiveSighting is a sighting arriving from a pact member: held if it
// is news, and the council meets on it.
func (w *World) receiveSighting(to *Civ, s *Sighting) {
	if old := to.Sightings[s.Fleet]; old != nil && old.Year >= s.Year {
		return
	}
	if to.Sightings == nil {
		to.Sightings = map[int]*Sighting{}
	}
	t := *s
	t.Intercept = -1
	to.Sightings[s.Fleet] = &t
	w.considerIntercept(to, &t)
}

// adriftFinds: a fleet, a scout or surveyors whose line passes within
// half a light year of a field adrift come upon it.
func (w *World) adriftFinds(x *Expedition) {
	c := w.Civs[x.Owner]
	if !c.Active() {
		return
	}
	a, b := w.line(x)
	for _, l := range w.Legacies {
		if l.Kind != Field || !l.Adrift || l.State != Buried || c.Found[l.ID] {
			continue
		}
		if c.Known[l.Node] && l.ships() == 0 {
			continue
		}
		if segmentDist(l.At, a, b) <= adriftFind {
			w.discover(c, l, "ship")
			if !c.Active() {
				return
			}
		}
	}
}

// scoutSeen is a scout arriving at a world whose people has orbital
// habitats or a defence grid: it is seen point-blank, the people is
// summoned, and it holds the looking against the scout's people unless
// the two are partners.
func (w *World) scoutSeen(x *Expedition) {
	c := w.Civs[x.Owner]
	oid := w.Owner[x.Star]
	if oid < 0 || oid == c.ID {
		return
	}
	o := w.Civs[oid]
	if !o.Active() || !(o.Known["orbital_habitats"] || o.Known["defence_grid"]) {
		return
	}
	o.Summoned = true
	if !o.Trade[c.ID] && !w.allied(o, c) {
		o.resent(c.ID, 0.5)
		if w.R.Float64() < 0.1 {
			w.log("The %s see a ship of the %s in their sky at %s, looking, and do not forget it.", o.Name, c.Name, w.star(x.Star))
		}
	}
}
