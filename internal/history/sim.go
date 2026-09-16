package history

import (
	"math/rand/v2"
	"sort"

	"worldgen/internal/galaxy"
)

// Generate runs the whole history for a seed and returns the world at the present.
func Generate(seed uint64, cfg Config) *World {
	r := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
	rg, err := galaxy.RegionByName(cfg.Region)
	if err != nil {
		rg, _ = galaxy.RegionByName("sol")
	}
	g := galaxy.GenerateAt(r, rg, cfg.Stars, cfg.Radius, cfg.Thickness)
	w := &World{Cfg: cfg, Seed: seed, G: g, R: r, Law: g.Law, Hazard: g.Law.Hazard()}
	n := len(g.Stars)
	w.Bio = make([]BioState, n)
	w.Owner = make([]int, n)
	w.Held = make([]int, n)
	for i := range n {
		w.Owner[i], w.Held[i] = -1, -1
	}
	if g.Sol >= 0 {
		w.Bio[g.Sol] = BioSimple
	}

	w.makeCycle()
	w.runDeep()
	w.runAges()
	w.runAge()
	w.skyEvents()
	sort.SliceStable(w.Events, func(i, j int) bool { return w.Events[i].Year < w.Events[j].Year })
	return w
}

// runAge is the civilisation engine. It runs from the dawn at the coarse
// tick, switches to the fine tick when the waning sets in, and stops when
// decline has truly set in: few enough still active, fertility low enough,
// and then a little longer so the present lands somewhere in the waning.
func (w *World) runAge() {
	cfg := w.Cfg
	step := cfg.MidStep
	w.dt = float64(step) / 1000
	y := cfg.Dawn
	var stopAt Year
	ended := false
	for {
		w.Now = y
		w.life()
		w.cosmic()
		w.tickHorrors()
		w.tickCivs()
		w.tickLegacies()
		w.updateHazard()
		active := w.activeCount()
		f := w.fertility()
		if w.Cfg.Debug && (y-cfg.Dawn)%1_000_000 == 0 {
			w.log("[debug: %d active, %d remnants, fertility %.2f, hazard %.2f]", active, len(w.Civs)-active-w.deadCount(), f, w.Hazard)
		}
		if step != cfg.FineStep && active <= cfg.FineActive && f < cfg.FineFertility {
			step = cfg.FineStep
			w.dt = float64(step) / 1000
			w.Waning = y
			w.log("The age is waning. Few still rise, and those that stand are old.")
		}
		if !ended && active <= cfg.EndActive && f < w.Cycle.Ends {
			ended = true
			stopAt = y + Year(w.R.Float64()*float64(cfg.Linger))
		}
		if ended && y >= stopAt && active <= cfg.EndActive {
			break
		}
		if float64(y-cfg.Dawn) > cfg.MaxFades*float64(w.Cycle.Fade) {
			w.Capped = true
			break
		}
		y += step
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
func (w *World) life() {
	for i, b := range w.Bio {
		switch b {
		case BioNone:
			if w.G.Stars[i].Hab > 0 && w.chance(w.G.Stars[i].Hab*0.000002*w.Law.Life()) {
				w.Bio[i] = BioSimple
			}
		case BioSimple:
			if w.chance(0.00005 * w.Law.Life()) {
				w.Bio[i] = BioComplex
			}
		case BioComplex:
			if i != w.G.Sol && w.Owner[i] < 0 && w.Held[i] < 0 && !w.G.Stars[i].Dead() && w.chance(0.0004*w.fertility()) {
				w.spawnCiv(i, nil, -1)
			}
		}
	}
}

// updateHazard: the galaxy gets more dangerous as horrors accumulate.
// This is what makes the aftermath the natural outcome rather than a forced one.
func (w *World) updateHazard() {
	held := 0
	beacons := 0
	for _, h := range w.Horrors {
		switch h.Kind {
		case Replicators, RogueMind:
			if !h.Dormant {
				held += len(h.Systems)
			}
		case Beacon:
			if !h.Dormant {
				beacons++
			}
		}
	}
	w.Hazard = min(2.5, w.Law.Hazard()+0.005*float64(held)+0.06*float64(beacons))
}
