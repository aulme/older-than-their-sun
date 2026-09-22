package history

import (
	"math/rand/v2"
	"sort"
	"time"

	"worldgen/internal/galaxy"
	"worldgen/internal/mind"
)

// Generate runs the whole history for a seed and returns the world at the present.
func Generate(seed uint64, cfg Config) *World {
	w := newWorld(seed, cfg)
	w.runDeep()
	w.runAges()
	w.runAge()
	w.skyEvents()
	w.placeNotes()
	sort.SliceStable(w.Chronicle, func(i, j int) bool { return w.Chronicle[i].Year < w.Chronicle[j].Year })
	return w
}

// newWorld builds the galaxy and an empty world on it, with the cycle set
// and the tick's phases in order. Nothing has happened yet: Generate runs
// the passes, and the test harness runs ticks by hand.
func newWorld(seed uint64, cfg Config) *World {
	if cfg.Tuning == nil {
		cfg.Tuning = mind.Default()
	}
	var src rand.Source = rand.NewPCG(seed, seed^0x9E3779B97F4A7C15)
	var draws *drawCount
	if cfg.Profile {
		// under -phases only: what each step draws, which is what says
		// whether it could ever be run beside another people's
		draws = &drawCount{src: src}
		src = draws
	}
	r := rand.New(src)
	rg, err := galaxy.RegionByName(cfg.Region)
	if err != nil {
		rg, _ = galaxy.RegionByName("sol")
	}
	g := galaxy.GenerateAt(r, rg, cfg.Stars, cfg.Radius, cfg.Thickness)
	w := &World{Cfg: cfg, Seed: seed, G: g, R: r, Law: g.Law, Hazard: g.Law.Hazard(), factsAt: map[int][]int{}, draws: draws}
	n := len(g.Stars)
	w.Bio = make([]BioState, n)
	w.Owner = make([]int, n)
	for i := range n {
		w.Owner[i] = -1
	}
	w.Now = cfg.DeepStart // the first notes are the substrate's, before the ages
	if g.Sol >= 0 {
		w.setBio(g.Sol, BioSimple)
	}
	w.Sources, w.sourcesAt = naturalSources(g)
	w.Reservoir = map[int]*Reservoir{}
	w.makeCycle()
	w.phases = []phase{
		{"life", (*World).life},
		{"cosmic", (*World).cosmic},
		{"transmitters", (*World).tickTransmitters},
		{"beneath", (*World).tickBeneath},
		{"messages", (*World).tickMessages},
		{"plagues", (*World).tickPlagues},
		{"civs", (*World).tickCivs},
		{"trade", (*World).trade},
		{"contracts", (*World).tickContracts},
		{"wars", (*World).tickWars},
		{"expeditions", (*World).tickExpeditions},
		{"objects", (*World).tickObjects},
		{"legacies", (*World).tickLegacies},
		{"hazard", (*World).updateHazard},
	}
	return w
}

// drawCount is the random source with a tally of what has been drawn
// from it. It wraps the source rather than the calls, so no draw can
// escape it, and it is only in the stream under -phases, where the
// count is what is wanted; the numbers it hands out are the source's
// own, so a counted run is the same history as an uncounted one.
type drawCount struct {
	src rand.Source
	n   uint64
}

func (d *drawCount) Uint64() uint64 {
	d.n++
	return d.src.Uint64()
}

// phase is one stage of the tick. The tick is the ordered list of them: a
// subsystem joins by inserting a phase at a named place, never by editing
// the loop.
type phase struct {
	Name string
	Run  func(*World)
}

// insertPhase puts p after the phase named after, or at the end if there is
// no such phase.
func (w *World) insertPhase(after string, p phase) {
	for i, q := range w.phases {
		if q.Name == after {
			w.phases = append(w.phases[:i+1], append([]phase{p}, w.phases[i+1:]...)...)
			return
		}
	}
	w.phases = append(w.phases, p)
}

// runPhases runs one tick of the current age at w.Now.
func (w *World) runPhases() {
	w.Ticks++
	if !w.Cfg.Profile {
		for _, p := range w.phases {
			p.Run(w)
		}
		return
	}
	if w.phaseTime == nil {
		w.phaseTime = map[string]time.Duration{}
		w.stepTime = map[string]time.Duration{}
		w.stepDraws = map[string]uint64{}
	}
	for _, p := range w.phases {
		start := time.Now()
		p.Run(w)
		w.phaseTime[p.Name] += time.Since(start)
	}
}

// profileLine logs the time each phase took since the last line and
// resets, and beside it what the tick was spent on: the peoples it
// stepped, the worlds they hold and the tales they carry. A phase's
// cost is read against those, not against the field, since the field
// only decides how many peoples there come to be.
func (w *World) profileLine() {
	line := "[phases:"
	for _, p := range w.phases {
		line += sprintf(" %s %dms", p.Name, w.phaseTime[p.Name].Milliseconds())
		w.phaseTime[p.Name] = 0
	}
	var living, worlds, tales int
	for _, c := range w.Civs {
		if !c.Living() {
			continue
		}
		living++
		worlds += len(c.Systems)
		tales += len(c.Lore)
	}
	line += sprintf(" | civs %d of %d, worlds %d, tales %d", living, len(w.Civs), worlds, tales)
	w.event(KDebug, nil, nil, -1, P{"text": line + "]"})
	steps := "[steps:"
	for _, s := range civSteps {
		steps += sprintf(" %s %dms/%dr", s.Name, w.stepTime[s.Name].Milliseconds(), w.stepDraws[s.Name])
		w.stepTime[s.Name], w.stepDraws[s.Name] = 0, 0
	}
	w.event(KDebug, nil, nil, -1, P{"text": steps + "]"})
}

// runAge is the civilisation engine. It runs from the dawn at one tick to
// the present, and stops when the age has truly faded: fertility low
// enough, few enough still rising if any measure of that is set, and then
// a little longer so the present lands somewhere in the waning. Nobody
// dies of age (ossify.go): an old empire stands, set in its ways or
// broken into heirs, until something else ends it, so the present finds
// successors and ruins, not blank stars, and the sky is what ends the
// age. The waning is declared once on the way, and changes nothing but
// the legends' chapter.
func (w *World) runAge() {
	cfg := w.Cfg
	w.dt = float64(cfg.Step) / 1000
	y := cfg.Dawn
	var stopAt Year
	ended := false
	for {
		w.Now = y
		w.runPhases()
		active := w.risingCount()
		f := w.fertility()
		if (y-cfg.Dawn)%1_000_000 == 0 {
			if w.Cfg.Debug {
				all := w.activeCount()
				w.event(KDebug, nil, nil, -1, P{"text": sprintf("[debug: %d active, %d of them rising, %d remnants, fertility %.2f, hazard %.2f, wall %.2f]", all, active, len(w.Civs)-all-w.deadCount(), f, w.Hazard, w.Thin)})
			}
			if w.Cfg.Profile {
				w.profileLine()
			}
		}
		if w.Waning == 0 && (cfg.FineActive == 0 || active <= cfg.FineActive) && f < cfg.FineFertility {
			w.Waning = y
			w.event(KWaning, nil, nil, -1, P{})
		}
		if !ended && (cfg.EndActive == 0 || active <= cfg.EndActive) && f < w.Cycle.Ends {
			ended = true
			stopAt = y + Year(w.R.Float64()*float64(cfg.Linger))
		}
		if ended && y >= stopAt {
			break
		}
		if float64(y-cfg.Dawn) > cfg.MaxFades*float64(w.Cycle.Fade) {
			w.Capped = true
			break
		}
		if cfg.Until > 0 && y >= cfg.Until {
			w.Truncated = true // asked for the age as of this year: the same ticks, stopped early
			break
		}
		y += cfg.Step
	}
	w.Present = y
	w.Now = y
	if w.Waning == 0 {
		w.Waning = y
	}
}

func (w *World) deadCount() int {
	n := 0
	for _, c := range w.Civs {
		if c.Stage == Dead {
			n++
		}
	}
	return n
}

func (w *World) activeCount() int {
	n := 0
	for _, c := range w.Civs {
		if c.Active() {
			n++
		}
	}
	return n
}

// risingCount is the peoples still rising: active, and their ways not
// yet set. Few still rise, and those that stand are old: that is the
// waning. The default config sets no bar on it (FineActive and EndActive
// zero), since under ossification the count never falls: the sky ends
// the age.
func (w *World) risingCount() int {
	n := 0
	for _, c := range w.Civs {
		if c.Active() && !c.Ossified && c.Stiff < 1 {
			n++
		}
	}
	return n
}

func (w *World) life() {
	for i, b := range w.Bio {
		switch b {
		case BioNone:
			if w.G.Stars[i].Hab > 0 && w.chance(w.G.Stars[i].Hab*0.000002*w.Law.Life()) {
				w.setBio(i, BioSimple)
			}
		case BioSimple:
			if w.chance(0.00005 * w.Law.Life()) {
				w.setBio(i, BioComplex)
			}
		case BioComplex:
			if i != w.G.Sol && w.Owner[i] < 0 && !w.G.Stars[i].Dead() && w.chance(0.0004*w.fertility()) {
				w.spawnCiv(i, nil, -1)
			}
		}
	}
}

// updateHazard: the galaxy gets more dangerous as what is still there
// accumulates: the worlds held by whatever eats them, the transmitters
// speaking, the tithes taken, the wall wearing thin. This is what makes
// the aftermath the natural outcome rather than a forced one.
func (w *World) updateHazard() {
	eaten := 0
	for _, c := range w.Civs {
		if c.Active() && !c.Asleep && c.Species.Profile().Eats {
			eaten += len(c.Systems)
		}
	}
	w.Hazard = min(2.5, w.Law.Hazard()+0.005*float64(eaten)+0.06*float64(w.transmitters())+w.Cfg.Tuning.Kinds.TitheHazard*float64(w.tithes())+min(0.5, 0.1*w.Thin))
}
