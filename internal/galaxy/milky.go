package galaxy

import (
	"encoding/json"
	"math"
)

// The Milky Way as a fixed substrate. Positions are galactocentric, in
// kiloparsecs: X from the Sun toward the centre (so the Sun sits at
// X = -R0), Y along the direction of galactic rotation at the Sun
// (longitude 90), Z toward the north galactic pole. This is the standard
// galactocentric frame; heliocentric galactic coordinates (l, b, d) map
// onto it directly.
//
// The structure is the textbook picture, not a fit to any survey: a boxy
// bulge and long bar, four logarithmic spiral arms plus the Local arm the
// Sun sits in, a thin disc that flares outward, a thick disc, a thin
// stellar halo, and a central molecular zone around Sagittarius A*. The
// numbers are the commonly quoted ones (R0 8.2 kpc, bar angle about 30
// degrees, pitch angles 11 to 13 degrees, scale length 2.6 kpc). They are
// good enough to say where a region is and what the sky is like there;
// they are not good enough to navigate by.

const (
	R0       = 8.2     // kpc, the Sun's distance from the centre
	Z0       = 0.02    // kpc, the Sun's height above the plane
	LyPerPc  = 3.26156 // light years per parsec
	LyPerKpc = 3261.56
)

// Vec is a galactocentric position in kpc.
type Vec struct{ X, Y, Z float64 }

// FromSun converts heliocentric galactic coordinates (degrees, degrees,
// kpc) to a galactocentric position.
func FromSun(l, b, d float64) Vec {
	lr, br := l*math.Pi/180, b*math.Pi/180
	return Vec{-R0 + d*math.Cos(br)*math.Cos(lr), d * math.Cos(br) * math.Sin(lr), Z0 + d*math.Sin(br)}
}

// Sun is the Sun's position.
var Sun = Vec{-R0, 0, Z0}

// R is the distance from the axis of the galaxy.
func (v Vec) R() float64 { return math.Hypot(v.X, v.Y) }

// Len is the distance from the centre.
func (v Vec) Len() float64 { return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z) }

// Az is the galactocentric azimuth in degrees: 180 at the Sun, increasing
// toward longitude 90 (the direction of rotation), so a trailing arm winds
// outward with increasing azimuth.
func (v Vec) Az() float64 {
	a := math.Atan2(v.Y, v.X) * 180 / math.Pi
	if a < 0 {
		a += 360
	}
	return a
}

// Sub, Dist.
func (v Vec) Sub(o Vec) Vec       { return Vec{v.X - o.X, v.Y - o.Y, v.Z - o.Z} }
func (v Vec) Dist(o Vec) float64  { return v.Sub(o).Len() }
func (v Vec) Add(o Vec) Vec       { return Vec{v.X + o.X, v.Y + o.Y, v.Z + o.Z} }
func (v Vec) Scale(k float64) Vec { return Vec{v.X * k, v.Y * k, v.Z * k} }

// ToSun gives heliocentric galactic coordinates (l, b in degrees, d in kpc).
func (v Vec) ToSun() (l, b, d float64) {
	h := v.Sub(Sun)
	d = h.Len()
	if d == 0 {
		return 0, 0, 0
	}
	l = math.Atan2(h.Y, h.X) * 180 / math.Pi
	if l < 0 {
		l += 360
	}
	b = math.Asin(h.Z/d) * 180 / math.Pi
	return
}

// Arm is a logarithmic spiral arm: R = RSun * exp((az - 180) * tan(pitch)),
// where RSun is its radius where it crosses the Sun's azimuth on the near
// side, and az is unwrapped over several turns. From and To bound the
// arm's unwrapped azimuth, and always include 180, the Sun's; an arm that
// starts at the far end of the bar starts at a negative azimuth.
type Arm struct {
	Key   string  `json:"key"`
	Name  string  `json:"name"`
	Short string  `json:"short"`
	RSun  float64 `json:"r_sun"` // kpc at the Sun's azimuth
	Pitch float64 `json:"pitch"` // degrees
	Width float64 `json:"width"` // kpc, roughly the half-width
	From  float64 `json:"from"`  // unwrapped azimuth range, degrees
	To    float64 `json:"to"`
	Major bool    `json:"major,omitempty"`
	Desc  string  `json:"desc"`
}

// Arms, inner to outer as they cross the Sun's azimuth, from
// data/laws.json. The two major arms (Scutum-Centaurus and Perseus) start
// at the ends of the bar; the two minor ones (Sagittarius-Carina and
// Norma-Outer) are gas and young stars more than a change in stellar
// density. The Local arm is a short spur.
var Arms []*Arm

func (a *Arm) rAt(az float64) float64 {
	return a.RSun * math.Exp((az-180)*math.Pi/180*math.Tan(a.Pitch*math.Pi/180))
}

// Dist is the approximate perpendicular distance from a point in the
// plane to the arm's ridge, or +Inf if the point is outside the arm's
// azimuth range on every turn.
func (a *Arm) Dist(v Vec) float64 {
	best := math.Inf(1)
	r, az := v.R(), v.Az()
	for k := -2.0; k <= 2; k++ {
		u := az + 360*k
		if u < a.From || u > a.To {
			continue
		}
		d := math.Abs(r-a.rAt(u)) * math.Cos(a.Pitch*math.Pi/180)
		if d < best {
			best = d
		}
	}
	return best
}

// armWeight is a Gaussian in the arm's distance, 1 on the ridge.
func (a *Arm) armWeight(v Vec) float64 {
	d := a.Dist(v)
	if math.IsInf(d, 1) {
		return 0
	}
	return math.Exp(-d * d / (2 * a.Width * a.Width))
}

// The bar: half-length 4.5 kpc, angled about 30 degrees from the Sun-centre
// line with its near end at positive longitude.
var barAngle = 33.0 * math.Pi / 180

func barFrame(v Vec) (along, across float64) {
	c, s := math.Cos(math.Pi-barAngle), math.Sin(math.Pi-barAngle)
	return v.X*c + v.Y*s, -v.X*s + v.Y*c
}

// InBar is true inside the long bar.
func InBar(v Vec) bool {
	al, ac := barFrame(v)
	return math.Abs(al) < 4.5 && math.Abs(ac) < 1.2 && math.Abs(v.Z) < 0.5
}

// scaleHeight of the thin disc, flaring outward.
func scaleHeight(R float64) float64 { return 0.3 * math.Exp(math.Max(0, R-R0)/4.5) }

// Density is the stellar number density relative to the Sun's
// neighbourhood. Thin disc, thick disc, boxy bulge, bar, nuclear cluster,
// halo, and a modest enhancement in the arms.
func Density(v Vec) float64 {
	R, z := v.R(), math.Abs(v.Z)
	thin := math.Exp(-(R-R0)/2.6) * math.Exp(-z/scaleHeight(R))
	thick := 0.12 * math.Exp(-(R-R0)/3.6) * math.Exp(-z/0.9)
	al, ac := barFrame(v)
	bulge := 60 * math.Exp(-math.Pow(math.Sqrt(al*al+(ac*ac)/0.36+(v.Z*v.Z)/0.25)/0.9, 1.2))
	bar := 4 * math.Exp(-math.Pow(al/4.5, 4)-math.Pow(ac/0.9, 2)-math.Pow(v.Z/0.35, 2))
	nucleus := 4e4 * math.Exp(-v.Len()/0.006)
	halo := 0.003 * math.Pow(math.Max(v.Len(), 0.5)/R0, -3)
	arm := 0.0
	for _, a := range Arms {
		arm = math.Max(arm, a.armWeight(v))
	}
	return (thin*(1+0.5*arm) + thick + bulge + bar + nucleus + halo) / solarDensity
}

var solarDensity = 1.0

// Youth is the rate of star formation relative to the Sun's neighbourhood:
// where the massive stars are, and so the supernovae, the nebulae and the
// hard sky. It follows the gas: the arms, the molecular ring around the
// bar's end, and the central molecular zone. The bar's interior and the
// halo are barren.
func Youth(v Vec) float64 {
	R, z := v.R(), math.Abs(v.Z)
	gas := math.Exp(-(R-R0)/2.0) * math.Exp(-z/(0.5*scaleHeight(R)))
	if R < 3 {
		gas *= 0.15 // swept clear by the bar
	}
	ring := 2.5 * math.Exp(-(R-4.5)*(R-4.5)/0.8) * math.Exp(-z/0.1)
	arm := 0.0
	for _, a := range Arms {
		w := a.armWeight(v)
		if a.Major {
			w *= 1.3
		}
		arm = math.Max(arm, w)
	}
	cmz := 60 * math.Exp(-math.Pow(v.Len()/0.25, 2)) * math.Exp(-z/0.05)
	y := gas*(0.25+2.5*arm) + ring + cmz
	for _, f := range Features {
		if f.Kind == Nebula || f.Kind == Cluster {
			if d := f.Pos.Dist(v); d < 3*f.Radius {
				y += f.Strength * math.Exp(-d*d/(2*f.Radius*f.Radius))
			}
		}
	}
	return y / solarYouth
}

var solarYouth = 1.0

// Metals is the metallicity [Fe/H] in dex: 0 at the Sun, richer inward
// and in the bar, poorer outward, in the thick disc and in the halo.
func Metals(v Vec) float64 {
	R, z := v.R(), math.Abs(v.Z)
	m := -0.06*(R-R0) - 0.25*math.Min(z, 2)
	if R < 3 {
		m = math.Max(m, 0.25) // the bar and bulge are old but metal-rich
	}
	if v.Len() > 20 || z > 2.5 {
		m = math.Min(m, -1.2) // the halo
	}
	for _, f := range Features {
		if f.Kind == Globular && f.Pos.Dist(v) < f.Radius {
			m = math.Min(m, -1.5)
		}
	}
	return math.Max(-2.2, math.Min(0.6, m))
}

// Glare is the hard sky relative to the Sun's: cosmic rays from young
// stars, the X-ray glow of the centre, the echoes of its outbursts. It
// is what makes the surface of a world a bad place to keep a biosphere.
func Glare(v Vec) float64 {
	r := v.Len()
	g := 0.4 + 0.6*Youth(v)
	g += 30*math.Exp(-r/0.4) + 3*math.Exp(-r/2.5)
	for _, f := range Features {
		if f.Kind == Remnant || f.Kind == Magnetar {
			if d := f.Pos.Dist(v); d < 4*f.Radius {
				g += 0.5 * f.Strength * math.Exp(-d*d/(2*f.Radius*f.Radius))
			}
		}
	}
	return g / solarGlare
}

var solarGlare = 1.0

// Crowd is the rate at which other stars pass close, relative to the Sun's
// neighbourhood: density times the velocity spread, which is high in the
// bulge and the halo, and enormous inside a globular cluster.
func Crowd(v Vec) float64 {
	d := Density(v)
	disp := 1.0
	if v.R() < 3 {
		disp = 2.5
	} else if math.Abs(v.Z) > 1.5 {
		disp = 3
	}
	for _, f := range Features {
		if f.Kind == Globular && f.Pos.Dist(v) < f.Radius {
			d *= 800
		}
	}
	return d * disp
}

// Exotic is how much the neighbourhood teaches about the deep physics:
// dead stars, collapsed stars, the great hole at the centre. Relative to
// the Sun's, which is a quiet corner with a few old pulsars a few hundred
// light years off.
func Exotic(v Vec) float64 {
	e := 0.0
	for _, f := range Features {
		w := 0.0
		switch f.Kind {
		case BlackHole:
			w = 1.0 * f.Strength
		case NeutronStar, Magnetar:
			w = 0.6 * f.Strength
		default:
			continue
		}
		e += w * math.Exp(-f.Pos.Dist(v)/0.3)
	}
	return e / solarExotic
}

var solarExotic = 1.0

// Zone names the part of the galaxy a position is in.
func Zone(v Vec) string {
	r, R, z := v.Len(), v.R(), math.Abs(v.Z)
	switch {
	case r < 0.03:
		return "the Heart"
	case r < 0.3:
		return "the Core"
	case z > 1.5 || r > 22:
		return "the Halo"
	case R < 2.5:
		return "the Bulge"
	case InBar(v):
		return "the Bar"
	case R < 5.5:
		return "the Inner Disc"
	case R < 10.5:
		return "the Middle Disc"
	case R < 14.5:
		return "the Outer Disc"
	default:
		return "the Rim"
	}
}

// Law is what the geography of the galaxy says about a place. Every value
// but Metals is relative to the Sun's neighbourhood, which is 1.
type Law struct {
	Pos     Vec
	R, Z    float64
	Density float64
	Youth   float64
	Metals  float64
	Glare   float64
	Crowd   float64
	Exotic  float64
	Zone    string
	Arm     *Arm    // the arm the place is in, or nil for the space between
	ArmDist float64 // kpc to that arm's ridge
	Near    []NearFeature
}

// NearFeature is a catalogued feature seen from a place.
type NearFeature struct {
	F    *Feature
	Dist float64 // kpc
}

// At computes the law at a position.
func At(v Vec) Law {
	l := Law{Pos: v, R: v.R(), Z: v.Z, Density: Density(v), Youth: Youth(v), Metals: Metals(v), Glare: Glare(v), Crowd: Crowd(v), Exotic: Exotic(v), Zone: Zone(v)}
	l.ArmDist = math.Inf(1)
	for _, a := range Arms {
		if d := a.Dist(v); d < l.ArmDist {
			l.Arm, l.ArmDist = a, d
		}
	}
	if l.ArmDist > 1.5*l.Arm.Width {
		l.Arm = nil
	}
	for _, f := range Features {
		d := f.Pos.Dist(v)
		if f.Kind == Sky || d <= f.Reach() {
			l.Near = append(l.Near, NearFeature{f, d})
		}
	}
	sortNear(l.Near)
	return l
}

func sortNear(ns []NearFeature) {
	for i := 1; i < len(ns); i++ {
		for j := i; j > 0 && ns[j].Dist < ns[j-1].Dist; j-- {
			ns[j], ns[j-1] = ns[j-1], ns[j]
		}
	}
}

// Spacing is the factor to apply to a star field's radius so that a fixed
// number of stars stand at the right average distance: sparser fields are
// wider. Clamped so a field is never smaller than a quarter or larger
// than two and a half times the Sun's.
func (l Law) Spacing() float64 {
	return math.Max(0.25, math.Min(2.5, math.Pow(l.Density, -1.0/3)))
}

// Rocky is the fraction of stars that can form worlds of rock and water,
// relative to the Sun's neighbourhood: it takes metals to make planets.
func (l Law) Rocky() float64 {
	if l.Metals < 0 {
		return math.Max(0.1, math.Pow(10, 0.7*l.Metals))
	}
	return math.Min(1.15, 1+0.3*l.Metals)
}

// Life is the rate at which life arises and grows complex, relative to the
// Sun's neighbourhood: worlds to arise on, and a sky gentle enough.
func (l Law) Life() float64 {
	return l.Rocky() / (1 + 0.2*math.Max(0, l.Glare-1))
}

// Hazard is the baseline galactic hazard for the history engine: 1 at the
// Sun, rising with the glare, never much below 1.
func (l Law) Hazard() float64 {
	return math.Max(0.9, math.Min(1.8, 1+0.15*math.Log2(math.Max(l.Glare, 0.5))))
}

// Massive is the factor on the share of massive stars in a field.
func (l Law) Massive() float64 { return math.Min(20, math.Max(0.05, l.Youth)) }

// ExoticMul is the research multiplier on the exotic domain.
func (l Law) ExoticMul() float64 {
	return math.Max(0.7, math.Min(2.5, 1+0.3*math.Log2(math.Max(l.Exotic, 0.2))))
}

// IndustryMul is the research multiplier on industry: metals are ore.
func (l Law) IndustryMul() float64 { return math.Max(0.6, math.Min(1.3, 1+0.25*l.Metals)) }

// Mean spacing in light years between the stars of a field of n stars.
func MeanSpacing(n int, radius, thickness float64) float64 {
	vol := math.Pi * radius * radius * thickness
	return 0.55 * math.Cbrt(vol/float64(n)) // 0.55: nearest-neighbour for a Poisson field
}

func init() {
	// normalise the laws to the Sun's neighbourhood
	solarDensity = Density(Sun)
	solarYouth = Youth(Sun)
	solarGlare = Glare(Sun)
	solarExotic = Exotic(Sun)
}

// MarshalJSON writes a position as [x, y, z].
func (v Vec) MarshalJSON() ([]byte, error) { return json.Marshal([3]float64{v.X, v.Y, v.Z}) }

// UnmarshalJSON reads [x, y, z].
func (v *Vec) UnmarshalJSON(b []byte) error {
	var a [3]float64
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*v = Vec{a[0], a[1], a[2]}
	return nil
}
