package history

import "math"

// Geometry in the dark: points and lines in light years, and the year a
// moving point enters a sphere. Nothing here reads the world; the
// adapters in sighting.go and intercept.go hand it positions and years.

// vec is a point or a displacement in light years.
type vec [3]float64

func (a vec) sub(b vec) vec             { return vec{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
func (a vec) add(b vec) vec             { return vec{a[0] + b[0], a[1] + b[1], a[2] + b[2]} }
func (a vec) scale(f float64) vec       { return vec{a[0] * f, a[1] * f, a[2] * f} }
func (a vec) dot(b vec) float64         { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func (a vec) norm() float64             { return math.Sqrt(a.dot(a)) }
func (a vec) dist(b vec) float64        { return a.sub(b).norm() }
func (a vec) lerp(b vec, f float64) vec { return a.add(b.sub(a).scale(f)) }

// entry is the earliest year in [t0, t1] at which a point moving from p0
// at velocity vp (per year, from t0) comes within r of an eye moving from
// e0 at velocity ve. A still eye has velocity zero. False when the point
// never comes that close in the window.
func entry(p0, vp, e0, ve vec, t0, t1, r float64) (float64, bool) {
	if t1 < t0 {
		return 0, false
	}
	d0, dv := p0.sub(e0), vp.sub(ve)
	a, b, c := dv.dot(dv), 2*d0.dot(dv), d0.dot(d0)-r*r
	if a < 1e-12 {
		if c <= 0 {
			return t0, true
		}
		return 0, false
	}
	disc := b*b - 4*a*c
	if disc < 0 {
		return 0, false
	}
	sq := math.Sqrt(disc)
	lo, hi := (-b-sq)/(2*a), (-b+sq)/(2*a)
	if hi < 0 || lo > t1-t0 {
		return 0, false
	}
	return t0 + math.Max(lo, 0), true
}

// segmentDist is the distance from a point to the segment ab.
func segmentDist(p, a, b vec) float64 {
	ab := b.sub(a)
	l2 := ab.dot(ab)
	if l2 < 1e-12 {
		return p.dist(a)
	}
	f := clamp(p.sub(a).dot(ab)/l2, 0, 1)
	return p.dist(a.lerp(b, f))
}

// pos is a star's position.
func (w *World) pos(star int) vec {
	s := &w.G.Stars[star]
	return vec{s.X, s.Y, s.Z}
}

// line is the two ends of a fleet's current leg: the stars it moves
// between, or the points a leg to or from a meeting in the dark runs
// between.
func (w *World) line(x *Expedition) (a, b vec) {
	if x.Path != nil {
		return x.Path[0], x.Path[1]
	}
	return w.pos(x.From), w.pos(x.Star)
}

// posAt is where a fleet in flight is at a year, clamped to its leg.
func (w *World) posAt(x *Expedition, t Year) vec {
	a, b := w.line(x)
	if x.Arrive <= x.Launched {
		return b
	}
	return a.lerp(b, clamp(float64(t-x.Launched)/float64(x.Arrive-x.Launched), 0, 1))
}

// velocity is a fleet's displacement per year on its leg.
func (w *World) velocity(x *Expedition) vec {
	a, b := w.line(x)
	if x.Arrive <= x.Launched {
		return vec{}
	}
	return b.sub(a).scale(1 / float64(x.Arrive-x.Launched))
}

// nearerEnd is the star at the end of a fleet's leg nearer a point, for
// the telling and the gazetteer: the star of a leg between stars, else
// the leg's From or Star by distance.
func (w *World) nearerEnd(x *Expedition, p vec) int {
	if w.pos(x.From).dist(p) <= w.pos(x.Star).dist(p) {
		return x.From
	}
	return x.Star
}
