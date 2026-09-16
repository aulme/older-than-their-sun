package galaxy

import (
	"math"
	"strings"
)

// The catalogue: real stars within 150 light years of the Sun, with their
// known planets. Built by cmd/mkcatalog from the HYG database and the NASA
// Exoplanet Archive; see catalog_data.go. Nearby regions are real, distant
// ones statistical.

// CatStar is a catalogued star system.
type CatStar struct {
	Name, Alt, Spect string
	L, B, Dist       float64 // galactic longitude and latitude in degrees, distance in ly
	Mag, Lum         float64 // apparent magnitude from Earth, luminosity in suns (0 if unknown)
	Var              string  // variable star designation, if any
	Mult             int
	Comps            []string // spectral types of the companions
	Mass, Age, Met   float64  // stellar mass in suns, age in Gyr, [Fe/H]; 0 if unknown
	Teff             float64  // K, 0 if unknown
	Planets          []CatPlanet
}

// CatPlanet is a known planet.
type CatPlanet struct {
	Name                               string
	MassE, RadE, Period, SMA, Teq, Ecc float64 // Earth masses, Earth radii, days, AU, K
	Method                             string
	Year                               int
}

// Position in the field frame: heliocentric galactic Cartesian, light years,
// x toward the centre, y along rotation, z north.
func (c *CatStar) XYZ() (x, y, z float64) {
	l, b := c.L*math.Pi/180, c.B*math.Pi/180
	return c.Dist * math.Cos(b) * math.Cos(l), c.Dist * math.Cos(b) * math.Sin(l), c.Dist * math.Sin(b)
}

// Proper is true if the name is a proper name rather than a designation.
func (c *CatStar) Proper() bool {
	for _, p := range []string{"HD ", "HIP ", "GJ ", "Gl ", "G ", "L ", "LP ", "LHS ", "TOI-", "HAT", "WASP", "K2-", "Kepler", "Wolf ", "Ross ", "GJ", "BD", "CD-", "WISE", "2MASS", "TYC", "HR ", "LSPM", "LTT", "NLTT", "Luyten", "Lalande", "Lacaille", "Groombridge", "Struve", "Kruger", "COCONUTS"} {
		if strings.HasPrefix(c.Name, p) {
			return false
		}
	}
	return true
}

// Class reads the spectral class. Luminosity class and oddities come back
// as a note: giants, white dwarfs, brown dwarfs, subdwarfs.
func (c *CatStar) Class() (class byte, note string) {
	sp := strings.TrimSpace(c.Spect)
	sp = strings.TrimPrefix(sp, "sd")
	sp = strings.TrimPrefix(sp, "d")
	sp = strings.TrimPrefix(sp, "k")
	sp = strings.TrimPrefix(sp, "g")
	for _, p := range []string{"CFBDSIR", "WISE", "2MASS", "ULAS", "SDSS", "COCONUTS", "CWISE", "Luhman"} {
		if strings.HasPrefix(c.Name, p) || strings.HasPrefix(c.Alt, p) {
			return 'M', "brown dwarf"
		}
	}
	if sp == "" && c.Teff > 0 {
		switch t := c.Teff; {
		case t < 2600:
			return 'M', "brown dwarf"
		case t < 3700:
			return 'M', ""
		case t < 5200:
			return 'K', ""
		case t < 6000:
			return 'G', ""
		case t < 7500:
			return 'F', ""
		case t < 10000:
			return 'A', ""
		default:
			return 'B', ""
		}
	}
	if sp == "" {
		switch {
		case c.Lum > 0 && c.Lum < 0.01, c.Mag > 12:
			return 'M', ""
		case c.Lum < 0.5:
			return 'K', ""
		case c.Lum < 2:
			return 'G', ""
		case c.Lum < 6:
			return 'F', ""
		}
		return 'A', ""
	}
	switch sp[0] {
	case 'O', 'B', 'A', 'F', 'G', 'K', 'M':
		class = sp[0]
	case 'D':
		return 'W', "white dwarf"
	case 'L', 'T', 'Y':
		return 'M', "brown dwarf"
	case 'W':
		return 'O', "Wolf-Rayet star"
	case 'C', 'S', 'N', 'R':
		return 'M', "carbon star"
	default:
		return 'K', ""
	}
	if strings.HasPrefix(c.Spect, "sd") {
		note = "subdwarf"
	}
	body := sp[1:]
	body = strings.ReplaceAll(body, "IV", "4")
	switch {
	case strings.Contains(body, "III"), strings.Contains(body, "II"):
		note = "giant"
	case strings.Contains(body, "Ia"), strings.Contains(body, "Ib"):
		note = "supergiant"
	case strings.Contains(body, "I") && !strings.Contains(body, "V"):
		note = "supergiant"
	}
	return class, note
}

// KnownDiscs: stars with a well-known debris disc.
var knownDiscs = map[string]string{
	"Vega": "a wide cold ring of dust, the first ever seen around another star", "Fomalhaut": "a great eye of dust with a sharp inner edge",
	"Ran": "two belts of debris and a young system still settling", "Tau Cet": "a thick disc of comets, ten times the Sun's",
	"Beta Pic": "an edge-on disc of dust and falling comets, seen from Earth as a line through the star", "GJ 803": "an edge-on disc with clumps moving outward through it",
	"Epsilon Ind": "a faint far belt", "Alsafi": "", "HD 69830": "a warm belt of dust close in, where an asteroid belt would be",
}
