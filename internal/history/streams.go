package history

import "math/rand/v2"

// The random streams.
//
// One stream per phase of the tick and one per people, each seeded by
// hashing a stable name together with the world's seed, never by drawing
// sub-seeds in turn. That is the whole point: with one stream, any change
// to what is drawn or when slides every later draw, so a change to the
// wearing changes the galaxy, the bloods and when the first peoples
// arise, and no two runs can be compared. Hashing a name keeps the
// domains apart — the same galaxy, the same bloods, the same deep pass
// and the same early arisings whichever way a mechanic is set.
//
// It cannot remove the coupling that is real: a tale reaching myth stops
// dread keeping ships from a star, so expansion moves and with it who
// meets whom. That is the point. A change having an effect nobody
// expected is fine and is usually the interesting part; a change having
// an effect only because it reshuffled the draws is not interesting at
// all, and these streams are what tells the two apart.
//
// It costs nothing at the call sites: w.R is reassigned at the top of
// each phase and each people's turn, so the many places that draw need
// no edit. The rule it sets is that a phase's draws belong to that
// phase, and a people's to that people.

// mix is splitmix64's finaliser: it takes a word with structure in it
// and returns one without.
func mix(x uint64) uint64 {
	x += 0x9E3779B97F4A7C15
	x = (x ^ (x >> 30)) * 0xBF58476D1CE4E5B9
	x = (x ^ (x >> 27)) * 0x94D049BB133111EB
	return x ^ (x >> 31)
}

// streamSeed hashes a stream's name with the world's seed. FNV-1a over
// the name, then splitmix over the two together, so neighbouring names
// ("civ:11", "civ:12") give unrelated streams.
func streamSeed(seed uint64, name string) uint64 {
	h := uint64(14695981039346656037)
	for i := range len(name) {
		h ^= uint64(name[i])
		h *= 1099511628211
	}
	return mix(seed ^ h)
}

// newStream makes the stream of a name. Under -phases every stream is
// counted through the one tally, so the draws of a step are still the
// source's own and no draw escapes the count.
func (w *World) newStream(name string) *rand.Rand {
	h := streamSeed(w.Seed, name)
	var src rand.Source = rand.NewPCG(h, mix(h))
	if w.draws != nil {
		src = &counted{src: src, d: w.draws}
	}
	return rand.New(src)
}

// stream is the named domain's stream, made on first ask and kept for
// the run. Making it on first ask is what lets a subsystem be added
// without moving anything else.
func (w *World) stream(name string) *rand.Rand {
	if r := w.streams[name]; r != nil {
		return r
	}
	r := w.newStream(name)
	w.streams[name] = r
	return r
}

// civStream is a people's own stream. A people's draws are its own, so
// what one people does cannot renumber another's history, and the
// peoples can be stepped in any order or beside each other.
func (w *World) civStream(c *Civ) *rand.Rand {
	for len(w.civR) <= c.ID {
		w.civR = append(w.civR, nil)
	}
	if w.civR[c.ID] == nil {
		w.civR[c.ID] = w.newStream("civ:" + itoa(c.ID))
	}
	return w.civR[c.ID]
}
