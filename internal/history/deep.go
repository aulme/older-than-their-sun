package history

// runDeep is the coarse pass: billions of years in big steps. Nothing is
// simulated as an agent here; it only seeds the board for the fine pass.
func (w *World) runDeep() {
	for y := w.Cfg.DeepStart; y < w.Cfg.DeepEnd; y += w.Cfg.DeepStep {
		w.Now = y
		w.deepLife()
		w.deepCosmic()
		w.deepPrecursors()
		w.deepElders()
	}
}

func (w *World) deepLife() {
	for i, s := range w.G.Stars {
		switch w.Bio[i] {
		case BioNone:
			if s.Hab > 0 && w.chance(s.Hab*0.0015) {
				w.Bio[i] = BioSimple
				w.log("Life arises on the worlds of %s.", w.G.Describe(i))
			}
		case BioSimple:
			if w.chance(0.012) {
				w.Bio[i] = BioComplex
				w.log("Complex life flourishes at %s.", w.star(i))
			}
		}
	}
}

// deepCosmic: gamma-ray bursts and similar sterilise a region. Sol is
// protected by fiat so that the player exists.
func (w *World) deepCosmic() {
	if !w.chance(0.02) {
		return
	}
	origin := w.R.IntN(len(w.G.Stars))
	radius := 20.0 + w.R.Float64()*30
	killed := 0
	for _, s := range w.G.Near(origin, radius) {
		if s != w.G.Sol && w.Bio[s] != BioNone {
			w.Bio[s] = BioNone
			killed++
		}
	}
	if w.Bio[origin] != BioNone && origin != w.G.Sol {
		w.Bio[origin] = BioNone
		killed++
	}
	if killed > 0 {
		w.log("A gamma-ray burst near %s sterilises %d living worlds within %.0f ly.", w.star(origin), killed, radius)
	}
}

// deepPrecursors: a civilisation rose and vanished before the fine record
// begins. It is not simulated, only its remains are placed.
func (w *World) deepPrecursors() {
	if !w.chance(0.012) {
		return
	}
	var candidates []int
	for i, b := range w.Bio {
		if b == BioComplex && i != w.G.Sol {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) == 0 {
		return
	}
	s := candidates[w.R.IntN(len(candidates))]
	switch w.R.IntN(3) {
	case 0:
		w.trace(s, "precursor vault", -1)
		w.Bio[s] = BioNone
		w.log("Something rose at %s, built, and was gone. Sealed vaults remain. The world is dead.", w.star(s))
	case 1:
		w.trace(s, "dyson remnant", -1)
		w.log("A shell of ancient machinery still turns around %s. Whoever built it is not there.", w.star(s))
	case 2:
		h := w.spawnHorror(Beacon, s, -1)
		w.log("At %s a beacon begins to broadcast. It has never stopped. It will later be called %s.", w.star(s), h.Name)
	}
}

// deepElders: something old settles into the dark. Dormant until disturbed.
func (w *World) deepElders() {
	if !w.chance(0.006) {
		return
	}
	s := w.R.IntN(len(w.G.Stars))
	if s == w.G.Sol || w.Held[s] >= 0 {
		return
	}
	h := w.spawnHorror(Elder, s, -1)
	h.Dormant = true
	w.log("Something settles into the dark around %s and goes still. Later ages will name it %s.", w.star(s), h.Name)
}
