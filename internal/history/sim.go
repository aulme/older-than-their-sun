package history

import (
	"math/rand/v2"
	"sort"

	"worldgen/internal/galaxy"
)

// Generate runs the whole history for a seed and returns the world at the present.
func Generate(seed uint64, cfg Config) *World {
	r := rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
	g := galaxy.Generate(r, cfg.Stars, cfg.Radius, cfg.Thickness)
	w := &World{Cfg: cfg, Seed: seed, G: g, R: r, Hazard: 1}
	n := len(g.Stars)
	w.Bio = make([]BioState, n)
	w.Owner = make([]int, n)
	w.Held = make([]int, n)
	for i := range n {
		w.Owner[i], w.Held[i] = -1, -1
	}
	w.Bio[g.Sol] = BioSimple

	w.runAges()
	w.runDeep()
	w.runEngine(cfg.MidStart, cfg.FineStart, cfg.MidStep)
	w.runEngine(cfg.FineStart, 0, cfg.FineStep)
	w.Now = 0
	w.dt = 1
	w.longDusk()
	sort.SliceStable(w.Events, func(i, j int) bool { return w.Events[i].Year < w.Events[j].Year })
	return w
}

// runEngine is the civilisation engine, run at whatever grain is asked for.
func (w *World) runEngine(from, to, step Year) {
	w.dt = float64(step) / 1000
	for y := from; y < to; y += step {
		w.Now = y
		w.life()
		w.cosmic()
		w.tickHorrors()
		w.tickCivs()
		w.updateHazard()
	}
}

func (w *World) life() {
	for i, b := range w.Bio {
		switch b {
		case BioNone:
			if w.G.Stars[i].Hab > 0 && w.chance(w.G.Stars[i].Hab*0.000002) {
				w.Bio[i] = BioSimple
			}
		case BioSimple:
			if w.chance(0.00005) {
				w.Bio[i] = BioComplex
			}
		case BioComplex:
			if i != w.G.Sol && w.Owner[i] < 0 && w.Held[i] < 0 && !w.G.Stars[i].Dead() && w.chance(0.00016) {
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
	w.Hazard = min(2.5, 1+0.005*float64(held)+0.08*float64(beacons))
}

// longDusk enforces the aftermath rule at the present. Anything still active
// is pushed into decline by the Long Dusk. Ideally this rarely fires.
func (w *World) longDusk() {
	for _, c := range w.Civs {
		if !c.Active() {
			continue
		}
		w.Dusk++
		if w.R.Float64() < 0.7 {
			w.contract(c, "dwindled through the Long Dusk")
		} else {
			w.endCiv(c, Extinct, "did not survive the Long Dusk")
		}
	}
}
