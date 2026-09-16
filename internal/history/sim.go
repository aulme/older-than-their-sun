package history

import (
	"math/rand/v2"

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

	w.runDeep()
	w.runFine()
	w.Now = 0
	w.settle()
	return w
}

// runFine is the recent-history pass: agents act every FineStep years.
func (w *World) runFine() {
	for y := w.Cfg.DeepEnd; y < 0; y += w.Cfg.FineStep {
		w.Now = y
		w.fineLife()
		w.fineCosmic()
		w.tickHorrors()
		w.tickCivs()
		w.updateHazard()
	}
}

func (w *World) fineLife() {
	for i, b := range w.Bio {
		switch b {
		case BioSimple:
			if w.chance(0.00002) {
				w.Bio[i] = BioComplex
			}
		case BioComplex:
			if i != w.G.Sol && w.Owner[i] < 0 && w.Held[i] < 0 && w.chance(0.00016) {
				w.spawnCiv(i)
			}
		}
	}
}

func (w *World) fineCosmic() {
	if !w.chance(0.00015) {
		return
	}
	origin := w.R.IntN(len(w.G.Stars))
	radius := 15.0 + w.R.Float64()*15
	w.log("A gamma-ray burst lights the sky near %s.", w.star(origin))
	for _, s := range append(w.G.Near(origin, radius), origin) {
		if s == w.G.Sol {
			continue
		}
		w.Bio[s] = BioNone
		if cid := w.Owner[s]; cid >= 0 {
			c := w.Civs[cid]
			w.loseSystem(c, s, "scoured world", nil)
			if s == c.Home {
				w.endCiv(c, Extinct, "were sterilised by a gamma-ray burst")
			} else if len(c.Systems) == 0 {
				w.endCiv(c, Extinct, "were sterilised by a gamma-ray burst")
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
			held += len(h.Systems)
		case Beacon:
			beacons++
		}
	}
	w.Hazard = 1 + 0.01*float64(held) + 0.15*float64(beacons)
}

// settle enforces the aftermath rule at the present. Anything still active
// is pushed into decline by the Long Dusk. Ideally this rarely fires.
func (w *World) settle() {
	for _, c := range w.Civs {
		if !c.Active() {
			continue
		}
		w.Dusk++
		if w.chance(0.7) {
			w.contract(c, "dwindled through the Long Dusk")
		} else {
			w.endCiv(c, Extinct, "did not survive the Long Dusk")
		}
	}
}
