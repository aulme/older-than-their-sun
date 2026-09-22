package galaxy

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strings"

	"worldgen/data"
)

// A field is the star field a history runs in: a few hundred stars around
// a point in the galaxy. Around the Sun the field is seeded with the real
// catalogue and filled in statistically; anywhere else it is drawn from
// the laws of the place, with the named feature it is anchored to at its
// centre. The field's radius scales with the density of the place so a
// fixed number of stars stand at the right average distance: crowded
// toward the centre, lonely at the rim.

// Region is a place in the galaxy to run a history in.
type Region struct {
	Name   string // "the Sun's neighbourhood", "Cygnus X-1", "the Rim"
	Code   string // designation prefix for procedural stars
	Pos    Vec
	Law    Law
	HasSol bool
	Anchor *Feature // a feature at the field's centre, or nil
}

// Preset is a named place, besides every feature in the catalogue: a
// row of data/laws.json.
type Preset struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Code string `json:"code"`
	Pos  Vec    `json:"pos"`
	Desc string `json:"desc"`
}

// Presets are the named places, in the file's order.
var Presets []*Preset

type lawsFile struct {
	Arms    []*Arm    `json:"arms"`
	Presets []*Preset `json:"presets"`
}

func init() {
	var f lawsFile
	data.Load("laws.json", &f)
	Arms, Presets = f.Arms, f.Presets
}

// RegionByName resolves a preset key, a feature name, or "x,y,z" in kpc.
func RegionByName(name string) (Region, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "sol"
	}
	for _, p := range Presets {
		if strings.EqualFold(p.Key, name) {
			return newRegion(p.Name, p.Code, p.Pos, nil, p.Key == "sol"), nil
		}
	}
	if f := FeatureByName(name); f != nil {
		if f.Kind == Sky {
			return Region{}, fmt.Errorf("%s is beyond the galaxy", f.Name)
		}
		return newRegion(f.Name, code(f.Name), f.Pos, f, false), nil
	}
	var x, y, z float64
	if n, _ := fmt.Sscanf(name, "%f,%f,%f", &x, &y, &z); n == 3 {
		return newRegion(fmt.Sprintf("(%.1f, %.1f, %.1f)", x, y, z), "FLD", Vec{x, y, z}, nil, false), nil
	}
	return Region{}, fmt.Errorf("unknown region %q", name)
}

func newRegion(name, code string, pos Vec, anchor *Feature, sol bool) Region {
	return Region{Name: name, Code: code, Pos: pos, Law: At(pos), HasSol: sol, Anchor: anchor}
}

func code(name string) string {
	var b []byte
	for _, w := range strings.Fields(strings.ToUpper(name)) {
		if w == "THE" {
			continue
		}
		for i := 0; i < len(w) && len(b) < 3; i++ {
			if w[i] >= 'A' && w[i] <= 'Z' {
				b = append(b, w[i])
			}
		}
		if len(b) >= 3 {
			break
		}
	}
	if len(b) == 0 {
		return "FLD"
	}
	return string(b)
}

// Describe the region in a line.
func (rg Region) Describe() string {
	l, b, d := rg.Pos.ToSun()
	if math.Abs(b) < 0.5 {
		b = 0
	}
	where := fmt.Sprintf("%.1f kpc from the centre", rg.Law.R)
	if rg.Pos.Len() < 0.5 {
		where = fmt.Sprintf("%.0f ly from the centre", rg.Pos.Len()*LyPerKpc)
	}
	if math.Abs(rg.Law.Z) >= 0.3 {
		where += fmt.Sprintf(", %.1f kpc %s the plane", math.Abs(rg.Law.Z), map[bool]string{true: "above", false: "below"}[rg.Law.Z > 0])
	}
	if rg.HasSol {
		return fmt.Sprintf("%s, in %s on the inner edge of %s, %s", rg.Name, rg.Law.Zone, Arms[2].Name, where)
	}
	arm := ", between the arms"
	if rg.Law.Arm != nil {
		arm = ", in " + rg.Law.Arm.Name
	}
	if rg.Law.R < 3 {
		arm = ""
	}
	return fmt.Sprintf("%s, in %s%s, %s; from Earth %.0f ly toward l=%.0f b=%.0f", rg.Name, rg.Law.Zone, arm, where, d*LyPerKpc, l, b)
}

// GenerateAt builds the star field of a region.
func GenerateAt(r *rand.Rand, rg Region, n int, radius, thickness float64) *Galaxy {
	law := rg.Law
	radius *= law.Spacing()
	thickness *= math.Max(0.5, math.Min(2, scaleHeight(law.R)/0.3)) * law.Spacing()
	thickness = math.Min(thickness, radius)
	g := &Galaxy{Radius: radius, Thickness: thickness, Region: rg, Law: law, Sol: -1}
	g.Stars = make([]Star, 0, n)
	rocky := law.Rocky()

	if rg.HasSol {
		g.Sol = 0
		g.Stars = append(g.Stars, Star{ID: 0, Name: "Sol", Class: 'G', Hab: 1.0, Mult: 1, Lifetime: 1e10, DiesAt: 5e9, Real: true, Mag: -26.7})
		g.addCatalogue(r, n, radius, rocky)
	} else if a := rg.Anchor; a != nil {
		if s, ok := anchorStar(a); ok {
			g.Stars = append(g.Stars, s)
		}
	}
	// fill
	massive := law.Massive()
	for len(g.Stars) < n {
		rr := radius * math.Sqrt(r.Float64())
		if rg.HasSol && rr < 20 {
			continue // everything within 20 ly of the Sun is already known
		}
		th := r.Float64() * 2 * math.Pi
		z := r.NormFloat64() * thickness / 2
		c, hab, life := pickClassIn(r, massive, law.Youth)
		mult := 1
		switch x := r.Float64(); {
		case x < 0.07:
			mult = 3
		case x < 0.45:
			mult = 2
		}
		s := Star{
			ID: len(g.Stars), Name: fmt.Sprintf("%s-%d", rg.Code, 1000+r.IntN(90000)),
			Class: c, X: rr * math.Cos(th), Y: rr * math.Sin(th), Z: z, Hab: hab,
			Mult: mult, Lifetime: life, DiesAt: diesAt(r, c, life), Mag: 99,
		}
		if law.Youth < 0.2 && (c == 'F' || c == 'G') && r.Float64() < 0.5 {
			// an old population: the brighter dwarfs have already gone
			s.Class, s.Hab, s.Lifetime = 'K', 0.85, 3e10
		}
		if s.DiesAt < -Age {
			s.Kill()
			if s.Class == 'N' {
				s.Remnant = remnantOf(r, c)
			}
		}
		g.Stars = append(g.Stars, s)
	}
	// systems
	g.Sys = make([]*System, len(g.Stars))
	for i := range g.Stars {
		s := &g.Stars[i]
		if s.Real && s.Name == "Sol" {
			g.Sys[i] = solarSystem()
			continue
		}
		g.Sys[i] = genSystem(r, s, s.cat, rocky, law.Metals)
	}
	g.index()
	return g
}

// addCatalogue seeds the field with the real stars within its radius, the
// best-known first, up to a budget so that the field stays sparse enough
// to simulate.
func (g *Galaxy) addCatalogue(r *rand.Rand, n int, radius, rocky float64) {
	type cand struct {
		c     *CatStar
		score float64
	}
	var cs []cand
	for i := range catalogStars {
		c := &catalogStars[i]
		if c.Dist > radius {
			continue
		}
		sc := 0.0
		if c.Proper() {
			sc += 3
		}
		if len(c.Planets) > 0 {
			sc += 2 + math.Min(2, 0.5*float64(len(c.Planets)))
		}
		if c.Dist < 25 {
			sc += 2
		} else if c.Dist < 50 {
			sc += 1
		}
		if c.Mag < 4 {
			sc += 1
		}
		cs = append(cs, cand{c, sc})
	}
	sort.SliceStable(cs, func(i, j int) bool {
		if cs[i].score != cs[j].score {
			return cs[i].score > cs[j].score
		}
		return cs[i].c.Dist < cs[j].c.Dist
	})
	budget := n * 55 / 100
	if budget > len(cs) {
		budget = len(cs)
	}
	for _, cd := range cs[:budget] {
		c := cd.c
		x, y, z := c.XYZ()
		class, note := c.Class()
		hab := 0.0
		life := 1e10
		for _, cw := range classWeights {
			if cw.c == class {
				hab, life = cw.hab, cw.life
			}
		}
		if note != "" && note != "subdwarf" {
			hab = 0
		}
		s := Star{ID: len(g.Stars), Name: c.Name, Alt: c.Alt, Class: class, X: x, Y: y, Z: z, Hab: hab, Mult: max(1, c.Mult),
			Lifetime: life, DiesAt: 5e8, Real: true, Mag: c.Mag, Note: note, cat: c}
		if class == 'W' {
			s.Class, s.Hab = 'W', 0
			s.DiesAt = -Age - 1
		}
		if note == "giant" || note == "supergiant" {
			s.DiesAt = 1e8 + int64(r.Float64()*1e9) // after the present
		}
		g.Stars = append(g.Stars, s)
	}
}

// anchorStar makes the object a region is anchored on into a star of the field.
func anchorStar(f *Feature) (Star, bool) {
	s := Star{Name: f.Name, Real: true, Mag: 99, Mult: 1, DiesAt: 1e15, Lifetime: 1e10}
	switch f.Kind {
	case BlackHole:
		s.Class, s.Remnant = 'N', "black_hole"
		if f.Name == "Sagittarius A*" {
			s.Remnant = "great_hole"
		}
	case NeutronStar:
		s.Class, s.Remnant = 'N', "neutron_star"
	case Magnetar:
		s.Class, s.Remnant = 'N', "magnetar"
	case Giant:
		s.Class, s.Note, s.Mult = 'B', "supergiant", 1
		s.Lifetime, s.DiesAt = 1e7, 5e8 // dies after the present
	default:
		return s, false
	}
	return s, true
}

func remnantOf(r *rand.Rand, was byte) string {
	if was == 'O' && r.Float64() < 0.6 || r.Float64() < 0.25 {
		return "black_hole"
	}
	if r.Float64() < 0.1 {
		return "magnetar"
	}
	return "neutron_star"
}

// pickClassIn draws a class with the massive share scaled by the place's
// youth, and the population aged where nothing is young.
func pickClassIn(r *rand.Rand, massive, youth float64) (byte, float64, float64) {
	total := 0.0
	ws := make([]float64, len(classWeights))
	for i, cw := range classWeights {
		w := cw.w
		if cw.c == 'O' || cw.c == 'B' || cw.c == 'A' {
			w *= massive
		}
		ws[i] = w
		total += w
	}
	x := r.Float64() * total
	for i, cw := range classWeights {
		if x < ws[i] {
			return cw.c, cw.hab, cw.life
		}
		x -= ws[i]
	}
	return 'M', 0.3, 1e12
}

// solarSystem is the one system we know.
func solarSystem() *System {
	mk := func(name string, k PlanetKind, m, a, t float64, moons int, tag string) Planet {
		return Planet{Name: name, Kind: k, MassE: m, RadE: radius(k, m), SMA: a, Period: period(a, 1), Teq: t, Known: true, Moons: moons, Tag: tag}
	}
	sys := &System{Home: 2, Arch: "lush", Belts: []float64{2.7, 40}}
	sys.Planets = []Planet{
		mk("Mercury", Rock, 0.055, 0.39, 440, 0, "airless"),
		mk("Venus", Rock, 0.815, 0.72, 328, 0, "under a crushing, poisonous sky"),
		mk("Earth", Rock, 1, 1, 255, 1, "the cradle"),
		mk("Mars", Rock, 0.107, 1.52, 210, 2, "cold and dry, with old riverbeds"),
		mk("Jupiter", GasGiant, 318, 5.2, 110, 95, "with an ocean moon and a volcanic one"),
		mk("Saturn", GasGiant, 95, 9.5, 81, 146, "ringed"),
		mk("Uranus", IceGiant, 14.5, 19.2, 58, 28, "tipped on its side"),
		mk("Neptune", IceGiant, 17.1, 30.1, 46, 16, ""),
	}
	sys.Planets[2].Temperate = true
	sys.Planets[2].RadE = 1
	return sys
}
