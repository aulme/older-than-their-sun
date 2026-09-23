package history

import (
	"math"
	"testing"

	"worldgen/internal/plague"
	"worldgen/internal/species"
)

// TestLifespanBands: a mortal blood's span lies in the band its traits
// name, hashed and not drawn, so the same seed gives the same spans and
// the registering moves no stream; a blood that does not turn over has
// none; and a drift that makes a people short-lived moves the span into
// the new band.
func TestLifespanBands(t *testing.T) {
	w := newTestWorld(t, 5, 30)
	for _, tc := range []struct {
		traits []string
		lo, hi int
	}{
		{nil, 60, 300}, {[]string{"shortlived"}, 30, 60}, {[]string{"longlived"}, 300, 1000},
	} {
		for range 20 {
			sp := species.Fixed(tc.traits...)
			w.register(sp)
			if sp.Lifespan < tc.lo || sp.Lifespan > tc.hi {
				t.Fatalf("%v: a span of %d, outside %d to %d", tc.traits, sp.Lifespan, tc.lo, tc.hi)
			}
		}
	}
	again := newTestWorld(t, 5, 30)
	for i := range 3 {
		again.register(species.Fixed())
		if a, b := again.Species[i].Lifespan, w.Species[i].Lifespan; a != b {
			t.Fatalf("blood %d: a span of %d in one world and %d in the same seed's other", i, a, b)
		}
	}
	m := species.GenerateWith(w.R, 1, "lush", species.Machine, 0)
	w.register(m)
	if m.Lifespan != 0 || m.Mortal() {
		t.Fatalf("a machine blood with a span of %d", m.Lifespan)
	}
	p := species.GenerateWith(w.R, 1, "ocean", species.Biological, species.Planetary)
	w.register(p)
	if p.Lifespan != 0 {
		t.Fatalf("a living world with a span of %d", p.Lifespan)
	}
	sp := species.Fixed("longlived")
	w.register(sp)
	sp.Traits = nil
	sp.Add("shortlived")
	if sp.Lived() {
		t.Fatal("a span of the long-lived passes for the short-lived")
	}
	sp.Live(unit(w.Seed, "x"))
	if !sp.Lived() {
		t.Fatalf("a short-lived span of %d", sp.Lifespan)
	}
}

// TestContinuityEnds: continuity is what the generations and the memory
// arts leave. A short-lived people keeps less than a long-lived one of
// the same arts; writing, the archives and medicine each keep more; a
// dark age cuts it hard and the cut heals; a machine keeps all but the
// drift; a people in its heaven keeps everything. And the wearing reads
// it: memory is one at the reference continuity.
func TestContinuityEnds(t *testing.T) {
	w := newTestWorld(t, 6, 40)
	w.Now = 1_000_000
	people := func(star int, traits ...string) *Civ {
		c := spawnAt(w, star, species.Fixed(traits...))
		return c
	}
	short, mid, long := people(1, "shortlived"), people(2), people(3, "longlived")
	cs, cm, cl := w.continuity(short), w.continuity(mid), w.continuity(long)
	if !(cs < cm && cm < cl) {
		t.Fatalf("short %.3f, middle %.3f, long %.3f: not in the order of their spans", cs, cm, cl)
	}
	before := w.continuity(mid)
	mid.Known["writing"] = true
	lit := w.continuity(mid)
	if lit <= before {
		t.Fatalf("writing: %.3f from %.3f", lit, before)
	}
	mid.Works = append(mid.Works, Work{Key: "archive", Node: "writing", Star: mid.Home, Legacy: -1})
	if a := w.continuity(mid); a <= lit {
		t.Fatalf("an archive: %.3f from %.3f", a, lit)
	}
	mid.Works[len(mid.Works)-1].Dark = true
	if a := w.continuity(mid); a != lit {
		t.Fatalf("an archive gone dark still keeps: %.3f from %.3f", a, lit)
	}
	mid.Works = nil
	span := w.lifespanOf(mid)
	mid.Known["medicine"] = true
	if w.lifespanOf(mid) <= span || w.continuity(mid) <= lit {
		t.Fatalf("medicine: a span of %.0f from %.0f", w.lifespanOf(mid), span)
	}
	mid.Known["life_extension"] = true
	extended := w.lifespanOf(mid)
	mid.Scars[ScarMortality] = true
	if w.lifespanOf(mid) >= extended {
		t.Fatalf("a mortality creed lives %.0f years, the extended %.0f", w.lifespanOf(mid), extended)
	}
	delete(mid.Scars, ScarMortality)
	whole := w.continuity(mid)
	mid.DarkAges, mid.LastDark = 1, w.Now
	cut := w.continuity(mid)
	if cut > whole/2 {
		t.Fatalf("a dark age leaves %.3f of %.3f", cut, whole)
	}
	w.Now += 100_000
	if healing := w.continuity(mid); healing <= cut || healing >= whole {
		t.Fatalf("a hundred thousand years on: %.3f, between %.3f and %.3f", healing, cut, whole)
	}
	w.Now += 200_000
	if healed := w.continuity(mid); math.Abs(healed-whole) > 1e-12 {
		t.Fatalf("the cut not healed: %.3f against %.3f", healed, whole)
	}
	m := spawnAt(w, 4, species.GenerateWith(w.R, 1, "lush", species.Machine, 0))
	if got, want := w.continuity(m), math.Exp(-w.Cfg.Tuning.Continuity.Drift); math.Abs(got-want) > 1e-12 {
		t.Fatalf("a machine keeps %.4f, want the drift's %.4f", got, want)
	}
	for _, c := range []*Civ{short, mid, long, m} {
		if k := w.continuity(c); k <= 0 || k >= 1 {
			t.Fatalf("%s: a continuity of %v", c.Tok(), k)
		}
	}
	// the wearing reads it
	ref := w.Cfg.Tuning.Continuity.Ref
	if got := w.memory(short); math.Abs(got-(1-cs)/(1-ref)) > 1e-12 {
		t.Fatalf("memory %.3f for a continuity of %.3f", got, cs)
	}
	// and the ways setting: more past, faster
	if !(w.contStiff(short) < w.contStiff(long) && w.contStiff(long) <= w.contStiff(m)) {
		t.Fatalf("the ways set at %.2f, %.2f and %.2f for the short-lived, the long-lived and a machine", w.contStiff(short), w.contStiff(long), w.contStiff(m))
	}
}

// TestDarkDepthReadsContinuity: with the noise off, the same people
// falls further the less of its past it keeps.
func TestDarkDepthReadsContinuity(t *testing.T) {
	w := newTestWorld(t, 7, 30)
	w.Cfg.Tuning.Ossify.DepthNoise = 0
	w.Now = 1_000_000
	short := spawnAt(w, 1, species.Fixed("shortlived"))
	long := spawnAt(w, 2, species.Fixed("longlived"))
	for _, c := range []*Civ{short, long} {
		c.Known["writing"], c.Known["printing"] = true, true
	}
	if ds, dl := w.darkDepth(short), w.darkDepth(long); ds <= dl {
		t.Fatalf("the short-lived fall %.3f deep and the long-lived %.3f", ds, dl)
	}
}

// TestUpload: the filter's three outcomes. Overcome is a blood of its
// own in software, with no span and nothing but the drift to lose;
// scarred is the creed of the flesh; declined is the heaven — the realm
// contracted to the world the substrate runs on, a remnant whose telling
// does not wear, and when the substrate's world goes the remain holds
// more of the telling than a wall.
func TestUpload(t *testing.T) {
	w, c := realm(t, 31, 3)
	old := c.Species
	filters["upload"].Overcome(w, c)
	if c.Species == old || !c.Species.Software || c.Species.Lifespan != 0 || c.Species.Parent != old {
		t.Fatalf("overcome: blood %d, software %v, span %d", c.Species.ID, c.Species.Software, c.Species.Lifespan)
	}
	if w.generations(c) != 0 {
		t.Fatalf("a people in software turns over %.2f times a thousand years", w.generations(c))
	}
	if natureMul(c, plague.Biological) >= 1 || natureMul(c, plague.Memetic) != 1 {
		t.Fatalf("a people in software bears the plagues of the body at %.2f and of the mind at %.2f", natureMul(c, plague.Biological), natureMul(c, plague.Memetic))
	}

	w, c = realm(t, 32, 3)
	filters["upload"].Scar(w, c)
	if !c.Scars[ScarFlesh] {
		t.Fatal("scarred without the creed of the flesh")
	}

	w, c = realm(t, 33, 4)
	for i := range 3 {
		spawnAt(w, 20+i, species.Fixed()) // the others the telling is about
	}
	fillTelling(w, c, 60, 1)
	w.resum(c)
	filters["upload"].Decline(w, c)
	if c.Stage != Remnant || c.Cause != "heaven" || !c.inHeaven() {
		t.Fatalf("declined: stage %v, cause %q", c.Stage, c.Cause)
	}
	if len(c.Systems) != 1 || c.worksAt("heaven", c.Systems[0]) != 1 {
		t.Fatalf("the heaven: worlds %v, works %v", c.Systems, c.Works)
	}
	if w.continuity(c) != 1 || w.memory(c) != 0 {
		t.Fatalf("in heaven the telling wears: continuity %v, memory %v", w.continuity(c), w.memory(c))
	}
	legacies := len(w.Legacies)
	w.wreck = &Wreckage{Destroy: 0, Leave: Abandoned}
	w.loseSystem(c, c.Systems[0], "abandoned", reason{})
	w.wreck = nil
	if len(w.Legacies) != legacies+1 {
		t.Fatal("the substrate left no remain")
	}
	l := w.Legacies[legacies]
	if l.Portrait != "heaven" || len(l.Testament) <= wallTales {
		t.Fatalf("the substrate's remain: %q with %d tales", l.Portrait, len(l.Testament))
	}
}

// TestArchiveBuilt: a people that writes and has nothing better to raise
// raises an archive, and the builder counts its keeping.
func TestArchiveBuilt(t *testing.T) {
	w := newTestWorld(t, 34, 30)
	w.Now = 1_000_000
	c := spawnAt(w, 0, species.Fixed())
	c.Known["writing"] = true
	sites := w.sites(c)
	found := false
	for _, s := range sites {
		if s.Key == "archive" {
			found = s.Keeps > 0
		}
		if s.Key == "heaven" {
			t.Fatal("a heaven offered to the builder")
		}
	}
	if !found {
		t.Fatalf("no archive among %d sites", len(sites))
	}
}
