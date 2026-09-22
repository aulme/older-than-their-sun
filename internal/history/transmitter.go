package history

import (
	"worldgen/internal/plague"
	"worldgen/internal/species"
)

// The transmitter: a beacon is not a people. It is a remain at a star, a
// Legacy of kind Threat with the memetics node, made by the Door's
// decline, by a thin wall, by an unleashed exotic artifact, by a
// deep-pass legacy, by an eldritch people's door (the same filter), and
// by a people that became one voice. It speaks while its state is
// Unleashed and is silent otherwise, with a range that grows with the
// listener's era, and it fires on any people that can listen: era two
// or more, a holding within range, a mind an idea can take. Each carries
// one of two payloads, drawn at its making. A corruption is the Signal
// as it was: forbid listening, a cult scar, or the decline, which is
// extinction or a transformation into a new transmitter in the
// listener's voice at its home; the scar also seeds a memetic plague, so
// the corruption can travel on a message (the decline leaves nobody to
// carry it). A seed is
// a conscious memetic plague in the plagues' sense: the listener fights
// the cure contest instead of the Signal, and if its home goes over a
// mind-rider wakes riding it, with the transmitter as its origin; cured
// before that, it never wakes.

// listenRange is how far a people of an era hears a transmitter from.
func listenRange(era int) float64 { return 25 + float64(era)*10 }

// makeTransmitter is a transmitter arising at a star: by a people, or by
// nothing (maker -1); speaking, or silent until the Find unleashes it.
// The payload is drawn here and never changes.
func (w *World) makeTransmitter(star, maker int, live bool) *Legacy {
	l := &Legacy{ID: len(w.Legacies), Age: -1, Maker: maker, Kind: Threat, Star: star, Node: "memetics", People: -1, Finder: -1, Source: -1, Plague: -1, Cond: Abandoned}
	l.Portrait = "transmitter"
	if w.R.Float64() < w.Cfg.Tuning.Kinds.SeedShare {
		l.Payload = Seed
	}
	if live {
		l.State = Unleashed
	}
	w.Legacies = append(w.Legacies, l)
	if maker >= 0 {
		w.factL(FUnleashed, w.Civs[maker], l).with(P{"way": "made"})
	}
	return l
}

// Transmitter says whether a remain is one: a threat that is not a
// sleeping people.
func (l *Legacy) Transmitter() bool { return l.Kind == Threat && l.People < 0 }

// Speaking says whether a transmitter is live.
func (l *Legacy) Speaking() bool { return l.Transmitter() && l.State == Unleashed }

// transmitters counts the live ones, for the hazard.
func (w *World) transmitters() int {
	n := 0
	for _, l := range w.Legacies {
		if l.Speaking() {
			n++
		}
	}
	return n
}

// listens says whether a people can hear a transmitter at all: era two
// or more, a mind an idea can take, and nobody home is nobody to hear.
func (w *World) listens(c *Civ) bool {
	p := c.Species.Profile()
	return c.Active() && c.Era >= 2 && p.Can(species.Believes) && !p.NoOne
}

// tickTransmitters is the phase: every live transmitter, every people
// within its range that has not heard it, at a small chance a tick.
func (w *World) tickTransmitters() {
	t := &w.Cfg.Tuning.Kinds
	for _, l := range w.Legacies {
		if !l.Speaking() {
			continue
		}
		for _, c := range w.Civs {
			if !w.listens(c) || c.Heard[l.ID] {
				continue
			}
			near := false
			for _, s := range c.Systems {
				if w.G.Dist(s, l.Star) <= listenRange(c.Era) {
					near = true
					break
				}
			}
			if !near || !w.chance(t.Listen) {
				continue
			}
			w.listen(c, l)
		}
	}
}

// listen is a people hearing a transmitter: the Signal for a corruption, the seed's
// plague for a seed.
func (w *World) listen(c *Civ, l *Legacy) {
	c.Heard[l.ID] = true
	if l.Payload == Seed {
		w.seeded(c, l)
		return
	}
	w.transmitter = l
	w.face(c, "beacon", 0)
	w.transmitter = nil
}

// seeded is the seed heard: a conscious memetic plague in the listener,
// with the transmitter as where it came from. The cure contest is the
// plagues' own; what wakes if the home goes over is a mind-rider (see
// parasite.go's wake).
func (w *World) seeded(c *Civ, l *Legacy) {
	if !w.bears(c, plague.Memetic) || w.plagued(c) {
		return
	}
	p := w.newPlague(plague.Memetic, c, "signal")
	p.Conscious = true
	p.Transmitter = l.ID
	l.Listeners++
	w.infect(c, p, nil, "signal")
	ev := w.event(KSignalPlague, c, nil, l.Star, P{})
	ev.Legacy, ev.Plague = l.ID, p.ID
}

// corrupted is the Signal's scar seeding a memetic plague in the
// listener, so the corruption can travel on a message. The proposal
// asked for contagion one, one hop and then dead; in the plagues'
// arithmetic contagion one is a wildfire and the hardest thing to cure
// (the first batch grew a fifth more cults), so it is an ordinary memetic
// plague, drawn as any is, and spreads as ideas do.
func (w *World) corrupted(c *Civ, l *Legacy) {
	if !c.Active() || !w.bears(c, plague.Memetic) || w.plagued(c) {
		return
	}
	p := w.newPlague(plague.Memetic, c, "signal")
	p.Conscious = false
	w.infect(c, p, nil, "signal")
}

func init() {
	def(&Filter{
		Key: "beacon",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "beacon", "overcome", "", w.transmitter.Star)
		},
		Scar: func(w *World, c *Civ) {
			l := w.transmitter
			c.Scars[ScarSignal] = true
			c.Morale -= 1
			l.Listeners++
			w.faced(c, "beacon", "scarred", "", l.Star)
			w.corrupted(c, l)
		},
		Decline: func(w *World, c *Civ) {
			l := w.transmitter
			l.Listeners++
			if w.R.Float64() < 0.3 {
				home := c.Home
				w.endCiv(c, Transformed, because("transmitter_changed").At(l.Star))
				c.Into = "cult"
				w.makeTransmitter(home, c.ID, true)
				w.event(KNewSignal, c, nil, home, P{})
				return
			}
			w.endCiv(c, Extinct, because("transmitter_listened").At(l.Star))
		},
	})
}
