package history

import (
	"math"
	"testing"

	"worldgen/internal/species"
)

// place puts a star at a point, for tests that want geometry by hand.
// The distance table is not rebuilt: sightings and meetings read the
// positions, not the table.
func place(w *World, star int, x, y, z float64) {
	s := &w.G.Stars[star]
	s.X, s.Y, s.Z = x, y, z
}

// flight is a fleet of c in flight from one point to another at c's
// speed, launched at Now, of a kind, with a target.
func flight(w *World, c *Civ, kind ExpKind, target int, from, to int, a, b vec, ships int) *Expedition {
	x := &Expedition{ID: len(w.Expeditions), Owner: c.ID, Target: target, Kind: kind, Star: to, From: from, Ships: ships, Back: -1,
		Launched: w.Now, Out: w.Now, Arrive: w.Now + Year(a.dist(b)*c.Speed), Base: -1, Manned: w.Now, Seen: map[int]bool{}, Drive: c.Speed,
		Path: &[2]vec{a, b}}
	w.Expeditions = append(w.Expeditions, x)
	return x
}

// TestEntry: the geometry core. A point crossing a still sphere enters
// it where the arithmetic says; one that misses never does; a moving eye
// is taken where it will be.
func TestEntry(t *testing.T) {
	// from x=10 toward the origin at one light year a year, an eye of
	// radius 2 at the origin: entered at year 8
	y, ok := entry(vec{10, 0, 0}, vec{-1, 0, 0}, vec{}, vec{}, 0, 10, 2)
	if !ok || math.Abs(y-8) > 1e-9 {
		t.Errorf("entry at %.2f, %v; want 8", y, ok)
	}
	// passing 3 away from an eye of radius 2: never
	if _, ok := entry(vec{10, 3, 0}, vec{-1, 0, 0}, vec{}, vec{}, 0, 10, 2); ok {
		t.Error("a line 3 off enters a sphere of 2")
	}
	// already inside at the start: at once
	if y, ok := entry(vec{1, 0, 0}, vec{-1, 0, 0}, vec{}, vec{}, 5, 10, 2); !ok || y != 5 {
		t.Errorf("inside at the start: %.2f %v", y, ok)
	}
	// an eye moving toward the point meets it halfway
	if y, ok := entry(vec{10, 0, 0}, vec{-1, 0, 0}, vec{}, vec{1, 0, 0}, 0, 10, 0); !ok || math.Abs(y-5) > 1e-9 {
		t.Errorf("a moving eye: %.2f %v; want 5", y, ok)
	}
	// outside the window: not seen
	if _, ok := entry(vec{10, 0, 0}, vec{-1, 0, 0}, vec{}, vec{}, 0, 5, 2); ok {
		t.Error("seen after the window closed")
	}
	if d := segmentDist(vec{5, 0.3, 0}, vec{0, 0, 0}, vec{10, 0, 0}); math.Abs(d-0.3) > 1e-9 {
		t.Errorf("segment distance %.2f, want 0.3", d)
	}
	if d := segmentDist(vec{12, 0, 0}, vec{0, 0, 0}, vec{10, 0, 0}); math.Abs(d-2) > 1e-9 {
		t.Errorf("segment distance past the end %.2f, want 2", d)
	}
}

// watcher is a starfarer at the origin with orbital habitats (a watch of
// eight light years) and an enemy at war with it, both at a fixed level.
func watcher(t *testing.T, seed uint64) (w *World, c, e *Civ) {
	t.Helper()
	w, c = starfarer(t, seed)
	e = spawnAt(w, 1, species.Fixed("cooperative"))
	e.Known = map[string]bool{}
	e.Starfaring = w.Now
	c.Known["orbital_habitats"] = true
	c.Speed, e.Speed = 100, 100
	c.Reach, e.Reach = 30, 30
	c.Mil, e.Mil = 6, 6
	place(w, c.Home, 0, 0, 0)
	place(w, e.Home, 60, 0, 0)
	w.firstGuard(e)
	return
}

// TestObservatorySees: an observatory at a held world sees a fleet of
// two ships forty light years out; the world alone sees it at eight.
func TestObservatorySees(t *testing.T) {
	w, c, e := watcher(t, 61)
	x := flight(w, e, Campaign, c.ID, e.Home, c.Home, vec{60, 0, 0}, vec{0, 0, 0}, 2)
	y, ey, ok := w.firstEntry(x, c)
	if !ok || x.Arrive-y != 800 || ey.kind != eyeWorld {
		t.Errorf("the world alone: seen %v with %d years to go by %s, want 800 by the world", ok, x.Arrive-y, ey.kind)
	}
	c.Works = append(c.Works, Work{Key: "observatory", Node: "orbital_habitats", Star: c.Home, Legacy: -1})
	y, ey, ok = w.firstEntry(x, c)
	if !ok || x.Arrive-y != 4000 || ey.kind != eyeWorks || ey.name != "observatory" {
		t.Errorf("with the mirrors: seen %v with %d years to go by %s %q, want 4000 by the observatory", ok, x.Arrive-y, ey.kind, ey.name)
	}
	// a torch is seen from twice as far, a fleet of eight at twice again
	far := flight(w, e, Campaign, c.ID, e.Home, c.Home, vec{100, 0, 0}, vec{0, 0, 0}, 2)
	far.Drive = 4
	if y, _, ok := w.firstEntry(far, c); !ok || far.Arrive-y != 8000 {
		t.Errorf("a torch: %d years to go, want 8000", far.Arrive-y)
	}
	far.Drive, far.Ships = 100, 8
	if y, _, ok := w.firstEntry(far, c); !ok || far.Arrive-y != 8000 {
		t.Errorf("eight ships: %d years to go, want 8000", far.Arrive-y)
	}
}

// TestMineSees: a mine at a stranger's belt sees a fleet passing through
// that star and nothing a light year off.
func TestMineSees(t *testing.T) {
	w, c, e := watcher(t, 62)
	belt := 5
	place(w, belt, 30, 0, 0)
	c.Works = append(c.Works, Work{Key: "mine", Node: "orbital_habitats", Star: belt, Legacy: -1})
	through := flight(w, e, Campaign, -1, e.Home, 6, vec{30, 20, 0}, vec{30, -20, 0}, 1)
	if _, ey, ok := w.firstEntry(through, c); !ok || ey.kind != eyeWorks || ey.star != belt {
		t.Errorf("a fleet through the belt: seen %v by %s at %d", ok, ey.kind, ey.star)
	}
	off := flight(w, e, Campaign, -1, e.Home, 6, vec{31, 20, 0}, vec{31, -20, 0}, 1)
	if _, _, ok := w.firstEntry(off, c); ok {
		t.Error("a fleet a light year off the belt was seen")
	}
}

// TestPicketSees: a picket at the midpoint sees a fleet passing where a
// world would not, and sees a fleet of two twice as far as one of one is
// half as far again.
func TestPicketSees(t *testing.T) {
	w, c, e := watcher(t, 63)
	x := flight(w, e, Campaign, -1, e.Home, 7, vec{40, 12, 0}, vec{0, 12, 0}, 4)
	if _, _, ok := w.firstEntry(x, c); ok {
		t.Error("a world of eight saw a fleet of four twelve light years off")
	}
	post := 8
	place(w, post, 20, 8, 0)
	p := w.launch(c, Scout, e, post, 1)
	p.Picket = true
	p.Base, p.Fed = post, w.Now
	if _, ey, ok := w.firstEntry(x, c); !ok || ey.kind != eyePicket {
		t.Errorf("the picket four light years off the line: seen %v by %s", ok, ey.kind)
	}
}

// TestTimetable: a fleet launched is seen on its timetable, inside the
// tick, by the world on its line; the sighting is held by the seer with
// the count and the line exact; the fleet does not know.
func TestTimetable(t *testing.T) {
	w, c, e := watcher(t, 64)
	place(w, e.Home, 4, 0, 0) // four light years: the whole line inside the world's watch, so seen as it leaves
	w.guardAt(e, e.Home).Ships = 3
	w.declare(e, c, because("border"))
	x := w.launch(e, Campaign, c, c.Home, 3)
	if x == nil {
		t.Fatal("no fleet")
	}
	if len(w.pending) != 1 || w.pending[0].Seer != c.ID || w.pending[0].Year != w.Now {
		t.Fatalf("pending %+v", w.pending)
	}
	w.tickExpeditions()
	s := c.Sightings[x.ID]
	if s == nil || s.Ships != 3 || s.Star != c.Home || s.Speed != 100 || s.Eye != eyeWorld {
		t.Fatalf("sighting %+v", s)
	}
	if c.Tally.Sightings != 1 || !x.Seen[c.ID] || len(w.Watch) != 1 {
		t.Errorf("tallies: %d sightings, seen %v, watch %d", c.Tally.Sightings, x.Seen[c.ID], len(w.Watch))
	}
	if e.Sightings[x.ID] != nil {
		t.Error("the fleet's own people hold a sighting of it")
	}
}

// TestTorchNotMet: a torch at four years a light year with thirty light
// years of warning yields no meeting from a holding twenty light years
// off the line; the guard at the destination fights at the world.
func TestTorchNotMet(t *testing.T) {
	w, c, e := watcher(t, 65)
	e.Speed, c.Speed = 4, 4
	colony := 9
	place(w, colony, 15, 20, 0)
	w.Owner[colony] = c.ID
	c.Systems = append(c.Systems, colony)
	w.addGuard(c, colony, 4)
	w.guardAt(c, c.Home).Ships = 4
	x := flight(w, e, Campaign, c.ID, e.Home, c.Home, vec{30, 0, 0}, vec{0, 0, 0}, 2)
	if _, _, ok := w.meetingPoint(c, x, w.Now+50); ok {
		t.Error("a torch thirty light years out was met from twenty light years off its line")
	}
}

// TestSlowFleetMet: a slow fleet seen five light years out is met by a
// guard three light years from its line, at a point the arithmetic
// allows, before it arrives.
func TestSlowFleetMet(t *testing.T) {
	w, c, e := watcher(t, 66)
	colony := 9
	place(w, colony, 2.5, 3, 0)
	w.Owner[colony] = c.ID
	c.Systems = append(c.Systems, colony)
	w.addGuard(c, colony, 4)
	x := flight(w, e, Campaign, c.ID, e.Home, c.Home, vec{5, 0, 0}, vec{0, 0, 0}, 2)
	m := w.Now + 50
	star, y, ok := w.meetingPoint(c, x, m)
	if !ok || star != colony || y >= x.Arrive {
		t.Fatalf("meeting: %v from %d at %d (arrives %d)", ok, star, y, x.Arrive)
	}
	if d := w.pos(colony).dist(w.posAtF(x, float64(y))); float64(m)+d*c.Speed > float64(y)+1 {
		t.Errorf("the interceptor cannot be there: %.2f light years, %d years", d, y-m)
	}
}

// TestMeetInDark: a slow fleet sighted at war is met by the council's
// interceptor on the timetable; the loser of the roll turns into the
// back hemisphere or goes back where it came from, and every ship lost
// lies adrift at the meeting point.
func TestMeetInDark(t *testing.T) {
	for seed := uint64(67); seed < 90; seed++ {
		w, c, e := watcher(t, seed)
		colony := 9
		place(w, colony, 2.5, 3, 0)
		w.Owner[colony] = c.ID
		c.Systems = append(c.Systems, colony)
		w.addGuard(c, colony, 4)
		c.Known["slow_interstellar"], e.Known["slow_interstellar"] = true, true
		w.declare(e, c, because("border"))
		x := flight(w, e, Campaign, c.ID, e.Home, c.Home, vec{5, 0, 0}, vec{0, 0, 0}, 2)
		w.timetable(x)
		w.tickExpeditions()
		s := c.Sightings[x.ID]
		if s == nil || !s.Feasible {
			t.Fatalf("seed %d: sighting %+v", seed, s)
		}
		if s.Intercept < 0 {
			continue // fear or the sizing kept it home; the dice again
		}
		y := w.Expeditions[s.Intercept]
		if y.Kind != Intercept || y.Quarry != x.ID || y.Meet >= x.Arrive || c.Tally.Intercepts != 1 {
			t.Fatalf("seed %d: interceptor %+v", seed, y)
		}
		if len(w.Meetings) != 1 {
			t.Fatalf("seed %d: %d meetings, want the one inside the tick", seed, len(w.Meetings))
		}
		m := w.Meetings[0]
		at := w.posAtF(y, float64(y.Meet))
		heading := vec{-1, 0, 0}
		if m.Won && !m.Broken {
			if !x.Returning {
				t.Errorf("seed %d: the beaten fleet keeps its course", seed)
			}
			if x.Star != e.Home && w.pos(x.Star).sub(at).dot(heading) >= 0 {
				t.Errorf("seed %d: the beaten fleet turned to %d, ahead of it", seed, x.Star)
			}
			if c.Tally.Caught != 0 || e.Tally.Caught != 1 {
				t.Errorf("seed %d: caught %d and %d", seed, c.Tally.Caught, e.Tally.Caught)
			}
		}
		if !m.Won && !x.Returning && x.Star != c.Home {
			t.Errorf("seed %d: the fleet that won the roll changed course", seed)
		}
		lost := m.Lost[0] + m.Lost[1]
		inFields := 0
		for _, l := range w.Legacies {
			if l.Kind == Field && l.Adrift {
				inFields += l.Wrecks + l.Derelicts
				if l.At != at {
					t.Errorf("seed %d: a field adrift at %v, the meeting at %v", seed, l.At, at)
				}
			}
		}
		if inFields != lost {
			t.Errorf("seed %d: %d ships lost in the dark, %d in fields adrift", seed, lost, inFields)
		}
		if !y.Returning && !y.Over {
			t.Errorf("seed %d: the interceptor is neither home-bound nor lost", seed)
		}
		return
	}
	t.Fatal("no seed launched an interceptor")
}

// TestTurnBack: a fleet turned back from a point lands at a star behind
// it, its own holding first, or where it came from.
func TestTurnBack(t *testing.T) {
	w, c, e := watcher(t, 91)
	x := flight(w, e, Campaign, c.ID, e.Home, c.Home, vec{60, 0, 0}, vec{0, 0, 0}, 2)
	at := vec{30, 0, 0}
	w.turnBack(x, at, w.Now)
	if !x.Returning || x.Path == nil || x.Path[0] != at {
		t.Errorf("turned back: %+v", x)
	}
	if x.Star != e.Home && w.pos(x.Star).sub(at).dot(vec{-1, 0, 0}) >= 0 {
		t.Errorf("turned to %d, which is ahead", x.Star)
	}
	// its own colony behind it within a hop is preferred over any star
	colony := 9
	place(w, colony, 40, 5, 0)
	w.Owner[colony] = e.ID
	e.Systems = append(e.Systems, colony)
	y := flight(w, e, Campaign, c.ID, e.Home, c.Home, vec{60, 0, 0}, vec{0, 0, 0}, 2)
	w.turnBack(y, at, w.Now)
	if y.Star != colony {
		t.Errorf("turned to %d, want the colony %d", y.Star, colony)
	}
	// and it lands there
	w.Now = y.Arrive
	w.tickExpeditions()
	if g := w.guardAt(e, colony); g == nil || g.Ships != 2 || !y.Over {
		t.Errorf("the turned-back fleet did not land at the colony")
	}
}

// TestSeenScout: a scout arriving at a world with orbital habitats is
// seen point-blank: the people is summoned and holds it, unless the two
// trade.
func TestSeenScout(t *testing.T) {
	w, c, e := watcher(t, 92)
	c.Summoned = false
	x := w.launch(e, Scout, c, c.Home, 1)
	w.Now = x.Arrive
	w.tickExpeditions()
	if !c.Summoned || c.Grudge[e.ID] != 0.5 {
		t.Errorf("summoned %v, grudge %.1f", c.Summoned, c.Grudge[e.ID])
	}
	c.Trade[e.ID] = true
	c.Grudge[e.ID] = 0
	w.addGuard(e, e.Home, 2)
	x = w.launch(e, Scout, c, c.Home, 1)
	w.Now = x.Arrive
	w.tickExpeditions()
	if c.Grudge[e.ID] != 0 {
		t.Errorf("a partner's scout drew a grudge of %.1f", c.Grudge[e.ID])
	}
}
