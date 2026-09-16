package galaxy

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
)

// Star systems: what orbits each star. Known planets come from the
// catalogue; the rest is drawn from the star's class, the metals of the
// place, and a little astrophysics (snow lines, tidal locking, hot
// Jupiters where the metals are). The point is that every star a story
// touches has a system worth describing, and that a species' home world
// is a real world in it.

type PlanetKind int

const (
	Rock PlanetKind = iota
	SuperEarth
	SubNeptune
	IceGiant
	GasGiant
	HotJupiter
	Dwarf // ice dwarfs and belt objects
)

var kindWords = map[PlanetKind]string{Rock: "rocky world", SuperEarth: "super-Earth", SubNeptune: "sub-Neptune", IceGiant: "ice giant", GasGiant: "gas giant", HotJupiter: "hot Jupiter", Dwarf: "ice dwarf"}

// Planet is one world.
type Planet struct {
	Name      string // the catalogued name, "" for procedural worlds
	Kind      PlanetKind
	MassE     float64 // Earth masses
	RadE      float64 // Earth radii
	Period    float64 // days
	SMA       float64 // AU
	Teq       float64 // K, equilibrium
	Ecc       float64
	Known     bool
	Method    string
	Year      int
	Temperate bool // rock or super-Earth in the temperate band
	Moons     int
	Tag       string // one detail: "tidally locked", "in a cloud of its own dust", ...
}

// System is what orbits a star.
type System struct {
	Planets []Planet
	Belts   []float64 // AU
	Disc    string    // a debris disc, described; "" if none
	Comp    string    // a companion note: "a white dwarf companion", ...
	Home    int       // index of the habitable world, -1 if none
	Arch    string    // archetype key of the habitable world: lush, twilight, floater...
	Missed  bool      // the habitable world is one the old surveys did not see
}

// lum, mass by class, for procedural stars
func classLum(c byte) float64 {
	return map[byte]float64{'O': 3e4, 'B': 800, 'A': 15, 'F': 3, 'G': 1, 'K': 0.3, 'M': 0.02, 'W': 0.001, 'N': 0}[c]
}
func classMass(c byte) float64 {
	return map[byte]float64{'O': 30, 'B': 8, 'A': 2, 'F': 1.3, 'G': 1, 'K': 0.7, 'M': 0.3, 'W': 0.6, 'N': 1.4}[c]
}

// teq is the equilibrium temperature of a world at a AU around a star of
// luminosity lum suns, Earth-like albedo.
func teq(lum, a float64) float64 { return 278 * math.Pow(lum, 0.25) / math.Sqrt(a) }

// temperate is the band where liquid water can stand on a rock.
func temperate(k PlanetKind, t float64) bool {
	return (k == Rock || k == SuperEarth) && t >= 185 && t <= 320
}

// tempChance is the chance that a star of a class has a temperate world
// of rock at all; Hab given one is the class weight divided by this, so
// the mean is kept.
var tempChance = map[byte]float64{'M': 0.55, 'K': 0.8, 'G': 0.85, 'F': 0.6, 'A': 0.1}

// genSystem draws a system for a star. rocky is the law's metal factor,
// metals its [Fe/H]; cat is the catalogue entry or nil.
func genSystem(r *rand.Rand, s *Star, cat *CatStar, rocky, metals float64) *System {
	sys := &System{Home: -1}
	lum := classLum(s.Class)
	mass := classMass(s.Class)
	if cat != nil {
		if cat.Lum > 0 {
			lum = cat.Lum
		}
		if cat.Mass > 0 {
			mass = cat.Mass
		}
		if cat.Met != 0 {
			metals = cat.Met
		}
		for _, p := range cat.Planets {
			sys.Planets = append(sys.Planets, fromCat(p, lum))
		}
		if d, ok := knownDiscs[cat.Name]; ok {
			if d == "" {
				d = "a debris disc"
			}
			sys.Disc = d
		}
		for _, c := range cat.Comps {
			switch {
			case strings.HasPrefix(c, "D"):
				sys.Comp = "a white dwarf companion"
			case strings.HasPrefix(c, "L"), strings.HasPrefix(c, "T"):
				sys.Comp = "a brown dwarf companion"
			}
		}
	}
	if s.Class == 'N' || s.Class == 'W' {
		// a remnant keeps a survivor or two, far out and cold
		for i := 0; i < r.IntN(3); i++ {
			a := 5 + 40*r.Float64()
			sys.Planets = append(sys.Planets, Planet{Kind: IceGiant, MassE: 5 + 20*r.Float64(), SMA: a, Period: period(a, mass), Teq: 30, Tag: "a survivor of the star's death"})
		}
		sortPlanets(sys.Planets)
		return sys
	}
	if s.Note == "giant" || s.Note == "supergiant" || s.Note == "brown dwarf" {
		n := r.IntN(3)
		for i := 0; i < n; i++ {
			a := 2 + 20*r.Float64()
			k := IceGiant
			if r.Float64() < 0.4 {
				k = GasGiant
			}
			sys.Planets = append(sys.Planets, Planet{Kind: k, MassE: 10 + 300*r.Float64(), SMA: a, Period: period(a, mass), Teq: teq(lum, a)})
		}
		sortPlanets(sys.Planets)
		return sys
	}
	// how many worlds, and are there giants
	mean := map[byte]float64{'M': 4, 'K': 5, 'G': 5, 'F': 4, 'A': 2.5, 'B': 1.2, 'O': 0.4}[s.Class] * (0.5 + 0.5*rocky)
	n := 0
	for r.Float64() < mean/(mean+1) && n < 9 {
		n++
	}
	if cat != nil {
		n = max(0, n-len(sys.Planets)) // the surveys found some of them
		n = min(n, 3)
	}
	giant := r.Float64() < 0.12*math.Pow(10, 1.4*metals)*math.Sqrt(mass)
	snow := 2.7 * math.Sqrt(lum)
	outer := math.Max(0.5, math.Min(60, 30*math.Sqrt(mass)))
	for i := 0; i < n; i++ {
		a := math.Exp(math.Log(0.03)+r.Float64()*(math.Log(outer)-math.Log(0.03))) * math.Sqrt(mass)
		if cat != nil && len(cat.Planets) > 0 {
			// unseen worlds sit farther out than the seen ones, mostly
			amax := 0.0
			for _, p := range cat.Planets {
				amax = math.Max(amax, p.SMA)
			}
			if r.Float64() < 0.8 {
				a = amax * (1.5 + 4*r.Float64())
			}
		}
		p := Planet{SMA: a, Period: period(a, mass), Teq: teq(lum, a)}
		x := r.Float64()
		if a < snow {
			switch {
			case x < 0.6:
				p.Kind, p.MassE = Rock, 0.1+1.4*r.Float64()
			case x < 0.85:
				p.Kind, p.MassE = SuperEarth, 1.5+6*r.Float64()
			default:
				p.Kind, p.MassE = SubNeptune, 3+12*r.Float64()
			}
		} else if mass < 0.5 {
			// small stars make small worlds, even out in the cold
			switch {
			case giant && x < 0.15:
				p.Kind, p.MassE = IceGiant, 10+20*r.Float64()
				giant = false
			case x < 0.5:
				p.Kind, p.MassE = Rock, 0.1+1.4*r.Float64()
			case x < 0.7:
				p.Kind, p.MassE = SuperEarth, 1.5+4*r.Float64()
			case x < 0.85:
				p.Kind, p.MassE = SubNeptune, 3+6*r.Float64()
			default:
				p.Kind, p.MassE = Dwarf, 0.002+0.1*r.Float64()
			}
		} else {
			switch {
			case giant && x < 0.35:
				p.Kind, p.MassE = GasGiant, 50+600*r.Float64()
				giant = false
			case x < 0.55:
				p.Kind, p.MassE = IceGiant, 8+25*r.Float64()
			case x < 0.75:
				p.Kind, p.MassE = SubNeptune, 3+8*r.Float64()
			default:
				p.Kind, p.MassE = Dwarf, 0.002+0.1*r.Float64()
			}
		}
		sys.Planets = append(sys.Planets, p)
	}
	if giant && r.Float64() < 0.25 {
		// a hot Jupiter, migrated in
		a := 0.02 + 0.06*r.Float64()
		sys.Planets = append(sys.Planets, Planet{Kind: HotJupiter, MassE: 100 + 800*r.Float64(), SMA: a, Period: period(a, mass), Teq: teq(lum, a)})
	}
	for i := range sys.Planets {
		p := &sys.Planets[i]
		if p.RadE == 0 {
			p.RadE = radius(p.Kind, p.MassE)
		}
		p.Temperate = temperate(p.Kind, p.Teq)
		if p.Kind == GasGiant || p.Kind == IceGiant {
			p.Moons = 2 + r.IntN(30)
		}
	}
	// belts: one inside the snow line if there is a gap, one beyond the last giant
	if r.Float64() < 0.5 {
		sys.Belts = append(sys.Belts, snow*(0.6+0.6*r.Float64()))
	}
	if r.Float64() < 0.6 {
		sys.Belts = append(sys.Belts, outer*(0.8+0.6*r.Float64()))
	}
	if sys.Disc == "" && r.Float64() < 0.08 {
		sys.Disc = "a faint disc of dust"
	}
	if sys.Comp == "" && r.Float64() < 0.05 {
		sys.Comp = "a brown dwarf companion"
	}
	sortPlanets(sys.Planets)

	// the habitable world
	hab := s.Hab
	known := -1
	for i, p := range sys.Planets {
		if p.Known && p.Temperate {
			known = i
		}
	}
	if known >= 0 {
		sys.Home = known
		s.Hab = math.Max(hab, 0.7) * rocky
	} else if tc, ok := tempChance[s.Class]; ok && hab > 0 && r.Float64() < tc*rocky {
		s.Hab = math.Min(1, hab/tc)
		for i, p := range sys.Planets {
			if p.Temperate && !p.Known {
				sys.Home = i
				break
			}
		}
		if sys.Home < 0 && !surveyed(cat, math.Pow(278*math.Pow(lum, 0.25)/265, 2)) {
			// put one in: a world the surveys would have missed
			a := math.Pow(278*math.Pow(lum, 0.25)/(230+70*r.Float64()), 2)
			p := Planet{Kind: Rock, MassE: 0.3 + 1.7*r.Float64(), SMA: a, Period: period(a, mass), Teq: teq(lum, a), Temperate: true}
			if r.Float64() < 0.3 {
				p.Kind, p.MassE = SuperEarth, 1.5+4*r.Float64()
			}
			p.RadE = radius(p.Kind, p.MassE)
			sys.Planets = append(sys.Planets, p)
			sortPlanets(sys.Planets)
			for i := range sys.Planets {
				if sys.Planets[i].SMA == a {
					sys.Home = i
				}
			}
			sys.Missed = cat != nil && len(cat.Planets) > 0
		}
		if sys.Home < 0 {
			s.Hab = 0 // the surveys would have seen it: there is none
		}
	} else {
		s.Hab = 0
	}
	if sys.Home >= 0 {
		sys.Arch = pickArch(r, s, sys, lum)
		p := &sys.Planets[sys.Home]
		if s.Class == 'M' && p.SMA < 0.25 {
			p.Tag = "tidally locked"
		}
		if sys.Arch == "floater" || sys.Arch == "volcanic" {
			// the home is a giant, or a moon of one
			for i, q := range sys.Planets {
				if (q.Kind == GasGiant || q.Kind == IceGiant) && (sys.Arch != "floater" || q.Teq > 120 && q.Teq < 350) {
					sys.Home = i
					break
				}
			}
		}
	}
	return sys
}

// surveyed is true if a world at a AU would have been found already: a
// known planet lies within a factor of two of it in orbit.
func surveyed(cat *CatStar, a float64) bool {
	if cat == nil {
		return false
	}
	for _, p := range cat.Planets {
		q := p.SMA
		if q == 0 && p.Period > 0 {
			q = math.Pow(p.Period/365.25, 2.0/3)
		}
		if q > a/2 && q < a*2 {
			return true
		}
	}
	return false
}

// EnsureHome gives a system a habitable world if it has none: the home of
// a people that was seeded there, or that made one.
func (sys *System) EnsureHome(r *rand.Rand, s *Star) {
	if sys.Home >= 0 {
		return
	}
	lum := classLum(s.Class)
	if lum <= 0 {
		lum = 0.001
	}
	mass := classMass(s.Class)
	a := math.Pow(278*math.Pow(lum, 0.25)/(240+60*r.Float64()), 2)
	p := Planet{Kind: Rock, MassE: 0.5 + 1.5*r.Float64(), SMA: a, Period: period(a, mass), Teq: teq(lum, a), Temperate: true, Tag: "made habitable"}
	p.RadE = radius(p.Kind, p.MassE)
	sys.Planets = append(sys.Planets, p)
	sortPlanets(sys.Planets)
	for i := range sys.Planets {
		if sys.Planets[i].SMA == a {
			sys.Home = i
		}
	}
	sys.Arch = pickArch(r, s, sys, lum)
	if sys.Arch == "floater" || sys.Arch == "volcanic" {
		sys.Arch = "lush"
	}
}

func fromCat(p CatPlanet, lum float64) Planet {
	q := Planet{Name: p.Name, MassE: p.MassE, RadE: p.RadE, Period: p.Period, SMA: p.SMA, Teq: p.Teq, Ecc: p.Ecc, Known: true, Method: p.Method, Year: p.Year}
	if q.SMA == 0 && q.Period > 0 {
		q.SMA = math.Pow(q.Period/365.25, 2.0/3)
	}
	if q.Teq == 0 && q.SMA > 0 {
		q.Teq = teq(lum, q.SMA)
	}
	m := q.MassE
	if m == 0 && q.RadE > 0 {
		m = math.Pow(q.RadE, 3.5)
	}
	q.MassE = m
	switch {
	case m > 50 && q.Teq > 800:
		q.Kind = HotJupiter
	case m > 50:
		q.Kind = GasGiant
	case m > 10, q.RadE > 3.5:
		q.Kind = IceGiant
	case m > 4 || q.RadE > 1.8:
		q.Kind = SubNeptune
	case m > 1.5:
		q.Kind = SuperEarth
	default:
		q.Kind = Rock
	}
	if q.RadE == 0 {
		q.RadE = radius(q.Kind, m)
	}
	q.Temperate = temperate(q.Kind, q.Teq)
	return q
}

func radius(k PlanetKind, m float64) float64 {
	switch k {
	case Rock, SuperEarth:
		return math.Pow(math.Max(m, 0.01), 0.27)
	case SubNeptune:
		return 2 + 0.1*m
	case IceGiant:
		return 3.5 + 0.02*m
	case Dwarf:
		return math.Pow(math.Max(m, 0.001), 0.33) * 1.2
	}
	return 10 + 2*math.Log10(math.Max(m/300, 0.1))
}

func period(a, mass float64) float64 { return 365.25 * math.Sqrt(a*a*a/mass) }

func sortPlanets(ps []Planet) {
	for i := 1; i < len(ps); i++ {
		for j := i; j > 0 && ps[j].SMA < ps[j-1].SMA; j-- {
			ps[j], ps[j-1] = ps[j-1], ps[j]
		}
	}
}

// pickArch chooses the home world archetype the system supports. The
// weights are the species table's, bent by what the world actually is.
func pickArch(r *rand.Rand, s *Star, sys *System, lum float64) string {
	p := sys.Planets[sys.Home]
	hasGiant, warmGiant, coldGiant := false, false, false
	for _, q := range sys.Planets {
		if q.Kind == GasGiant || q.Kind == IceGiant {
			hasGiant = true
			if q.Teq > 120 && q.Teq < 350 {
				warmGiant = true
			} else if q.Teq <= 120 {
				coldGiant = true
			}
		}
	}
	locked := s.Class == 'M' && p.SMA < 0.25
	w := map[string]float64{"lush": 30, "ocean": 15, "arid": 12, "twilight": 10, "superterran": 8, "lowg": 6, "hothouse": 5, "iceshell": 5, "floater": 3, "volcanic": 3, "dim": 3}
	if locked {
		w["twilight"] *= 6
		w["lush"] *= 0.3
		w["ocean"] *= 0.5
	} else {
		w["twilight"] = 0
	}
	if p.MassE > 2.5 {
		w["superterran"] *= 4
	} else {
		w["superterran"] *= 0.15
	}
	if p.MassE < 0.5 {
		w["lowg"] *= 4
	} else {
		w["lowg"] *= 0.1
	}
	if p.Teq > 290 {
		w["hothouse"] *= 3
		w["arid"] *= 2
	} else if p.Teq < 230 {
		w["hothouse"] = 0
		w["ocean"] *= 0.5
	}
	if coldGiant {
		w["iceshell"] *= 3
	} else {
		w["iceshell"] *= 0.3
	}
	if warmGiant {
		w["floater"] *= 4
	} else {
		w["floater"] = 0
	}
	if hasGiant {
		w["volcanic"] *= 2
	} else {
		w["volcanic"] = 0
	}
	if sys.Comp == "a brown dwarf companion" {
		w["dim"] *= 6
	} else {
		w["dim"] = 0
	}
	if p.Known {
		// a world Earth has seen stays the home; no moving off to a moon
		w["floater"], w["volcanic"] = 0, 0
	}
	total := 0.0
	for _, v := range w {
		total += v
	}
	x := r.Float64() * total
	for _, k := range []string{"lush", "ocean", "arid", "twilight", "superterran", "lowg", "hothouse", "iceshell", "floater", "volcanic", "dim"} {
		if x < w[k] {
			return k
		}
		x -= w[k]
	}
	return "lush"
}

// ordinal words
var ordinals = []string{"first", "second", "third", "fourth", "fifth", "sixth", "seventh", "eighth", "ninth", "tenth", "eleventh", "twelfth"}

// WorldName names a world of a star: its catalogued name, or "the third world of X".
func (sys *System) WorldName(i int, star string) string {
	if i < 0 || i >= len(sys.Planets) {
		return star
	}
	p := sys.Planets[i]
	if p.Name != "" {
		return p.Name
	}
	if i < len(ordinals) {
		return "the " + ordinals[i] + " world of " + star
	}
	return fmt.Sprintf("the %dth world of %s", i+1, star)
}

// HomeName names the habitable world, with the archetype's twist for
// moons and cloud decks.
func (sys *System) HomeName(star string) string {
	if sys.Home < 0 {
		return star
	}
	n := sys.WorldName(sys.Home, star)
	switch sys.Arch {
	case "floater":
		return "the clouds of " + n
	case "volcanic", "iceshell":
		return "a moon of " + n
	}
	return n
}

// DescribePlanet gives one world in a phrase.
func (p Planet) Describe() string {
	var b strings.Builder
	if p.Name != "" {
		b.WriteString(p.Name + ", ")
	}
	if p.Kind == Rock || p.Kind == SuperEarth {
		switch {
		case p.Temperate:
			b.WriteString("a temperate ")
		case p.Teq > 600:
			b.WriteString("a molten ")
		case p.Teq > 320:
			b.WriteString("a scorched ")
		case p.Teq < 185:
			b.WriteString("a frozen ")
		default:
			b.WriteString("a ")
		}
	} else if p.Kind == Dwarf {
		b.WriteString("an ")
	} else if p.Kind == IceGiant {
		b.WriteString("an ")
	} else {
		b.WriteString("a ")
	}
	b.WriteString(kindWords[p.Kind])
	if p.Kind != Dwarf && p.Kind != HotJupiter {
		if p.MassE >= 20 {
			fmt.Fprintf(&b, " of %.0f Earth masses", p.MassE)
		} else if p.MassE > 0 {
			fmt.Fprintf(&b, " of %.1f Earth masses", p.MassE)
		}
	}
	if p.SMA > 0 {
		switch {
		case p.Period < 100:
			fmt.Fprintf(&b, " with a %.0f-day year", p.Period)
		case p.Period < 2*365.25:
			fmt.Fprintf(&b, " with a %.0f-day year", p.Period)
		default:
			fmt.Fprintf(&b, " at %.1f AU", p.SMA)
		}
	}
	if p.Moons == 1 {
		b.WriteString(" with a moon")
	} else if p.Moons > 1 {
		fmt.Fprintf(&b, " with %d moons", p.Moons)
	}
	if p.Tag != "" {
		b.WriteString(", " + p.Tag)
	}
	if p.Known && p.Year > 0 {
		fmt.Fprintf(&b, " (found from Earth in %d)", p.Year)
	}
	return b.String()
}

// Describe gives the system in a sentence or two.
func (sys *System) Describe(star string) string {
	var parts []string
	n := len(sys.Planets)
	switch n {
	case 0:
		parts = append(parts, "no worlds")
	case 1:
		parts = append(parts, "one world: "+sys.Planets[0].Describe())
	default:
		var ws []string
		for i, p := range sys.Planets {
			d := p.Describe()
			if i == sys.Home {
				d = "*" + d
			}
			ws = append(ws, d)
		}
		parts = append(parts, fmt.Sprintf("%d worlds: %s", n, strings.Join(ws, "; ")))
	}
	if len(sys.Belts) > 0 {
		var bs []string
		for _, a := range sys.Belts {
			if a < 1 {
				bs = append(bs, fmt.Sprintf("%.2f AU", a))
			} else {
				bs = append(bs, fmt.Sprintf("%.1f AU", a))
			}
		}
		parts = append(parts, "belts at "+strings.Join(bs, " and "))
	}
	if sys.Disc != "" {
		parts = append(parts, sys.Disc)
	}
	if sys.Comp != "" {
		parts = append(parts, sys.Comp)
	}
	return strings.Join(parts, "; ")
}
