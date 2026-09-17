package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/galaxy"
)

// Sources: what the galaxy yields. A source is anything with a yield: a
// world, a belt, a giant, the star itself, a comet field, a nebula or a
// doomed giant whose reach covers the star. The natural ones are placed
// once at world generation from the systems the galaxy made; later steps
// add structures, elder remains and made things to the same record. A
// source yields to whoever holds its star and knows one of the nodes it
// needs, and to a nomad fleet based there.

// SourceKind is what sort of thing a source is.
type SourceKind uint8

const (
	WorldSource SourceKind = iota
	BeltSource
	GiantSource
	StarSource
	CosmicSource
	StructureSource
	ElderSource
	MadeSource
)

func (k SourceKind) String() string {
	return [...]string{"world", "belt", "giant", "star", "cosmic", "structure", "elder", "made"}[k]
}

// Source is one thing with a yield.
type Source struct {
	ID      int
	Key     string // what it is, for the reports: "habitable", "wood", "fuels", "atom", "fusion", "rocky", "terraformed", "heavy", "belt", "giant", "star", "comets", "nebula", "doomed_giant"
	Name    string // for a line: "the belt at X"
	Kind    SourceKind
	Star    int             // the star it is at, or -1 for a ranged source
	Feature *galaxy.Feature // for a ranged source: the thing whose reach it is
	Radius  float64         // kpc, for a ranged source
	Yield   flow.Income     // per tick
	Needs   []string        // any one of these harnesses it; none means it needs nothing
	With    []string        // and all of these
	Cradle  bool            // the habitable world: the people that arose on it needs nothing to farm it, and its profile's cradle multiplier applies
	Mobile  bool            // moves with its holder; see step 5
	Holder  int             // the people that harnessed it, or -1
	Carried int             // the expedition carrying it, or -1
}

// worldYield is what a habitable world yields in organic matter by archetype.
var worldYield = map[string]float64{
	"lush": 9, "ocean": 8, "superterran": 7, "arid": 6, "twilight": 6, "lowg": 6, "iceshell": 6,
	"hothouse": 5, "floater": 5, "volcanic": 5, "dim": 5,
}

// starYield is what a star's light gives in energy by class, once there
// are habitats in orbit to catch it. Dead stars give nothing.
var starYield = map[byte]float64{'M': 1, 'K': 2, 'G': 3, 'F': 4, 'A': 4, 'B': 4, 'O': 4}

// The yields and needs of the natural sources. These, with the upkeep
// table in tech, are the calibration: a cradle alone should run era 2 in
// comfort, strain at era 3 and need colonies or trade for era 4.
const (
	rockyYield      = 5.0  // metal per rocky world, times the metallicity
	richRockyBonus  = 3.0  // more for a superterran or volcanic home
	terraformYield  = 4.0  // organic matter from a world remade
	heavyYield      = 2.0  // energy from a world rich in heavy elements, with the atom
	woodYield       = 5.0  // energy from a habitable world's wood, wind and water, with fire
	fuelsYield      = 8.0  // energy from a habitable world's coal, with steam
	atomYield       = 5.0  // energy from a habitable world's fissile ore, with the atom
	fusionYield     = 14.0 // energy from a habitable world's seas, with fusion
	beltYield       = 4.0  // metal per belt
	giantYield      = 3.0  // energy per giant, with fusion and habitats
	cometYield      = 1.0  // organic matter from comets shaken loose in a crowded field
	crowded         = 3.0  // the crowd law above which comets are shaken loose
	nebulaYield     = 1.0  // organic matter per held star inside a nebula
	doomedYield     = 2.0  // energy per held star inside a doomed giant's reach
	remnantMetalMul = 2.0  // rocky worlds inside a supernova remnant
	globularMetal   = 0.5  // rocky worlds inside a globular cluster
	globularStar    = 2.0  // and the star's light there
)

// naturalSources places the natural sources of a galaxy. It reads only
// geometry and the systems, and draws nothing. The second return is the
// index by star: which sources yield there, ranged ones included.
func naturalSources(g *galaxy.Galaxy) ([]*Source, [][]int) {
	var out []*Source
	at := make([][]int, len(g.Stars))
	add := func(s *Source) *Source {
		s.ID = len(out)
		s.Holder, s.Carried = -1, -1
		out = append(out, s)
		return s
	}
	inside := func(star int, kind galaxy.FeatureKind) bool {
		pos := g.Position(star)
		for _, f := range galaxy.Features {
			if f.Kind == kind && f.Pos.Dist(pos) <= f.Radius {
				return true
			}
		}
		return false
	}
	metals := g.Law.IndustryMul()
	for i := range g.Stars {
		st := &g.Stars[i]
		sys := g.Sys[i]
		name := st.Name
		var here []*Source
		if sys.Home >= 0 {
			here = append(here,
				&Source{Key: "habitable", Name: "the fields of " + name, Kind: WorldSource, Star: i, Yield: flow.Income{flow.O: worldYield[sys.Arch]}, Needs: []string{"agriculture"}, Cradle: true},
				&Source{Key: "wood", Name: "the woods and rivers of " + name, Kind: WorldSource, Star: i, Yield: flow.Income{flow.E: woodYield}, Needs: []string{"fire", "cold_chemistry"}},
				&Source{Key: "fuels", Name: "the coal of " + name, Kind: WorldSource, Star: i, Yield: flow.Income{flow.E: fuelsYield}, Needs: []string{"steam"}},
				&Source{Key: "atom", Name: "the ore of " + name, Kind: WorldSource, Star: i, Yield: flow.Income{flow.E: atomYield}, Needs: []string{"atomic"}},
				&Source{Key: "fusion", Name: "the seas of " + name, Kind: WorldSource, Star: i, Yield: flow.Income{flow.E: fusionYield}, Needs: []string{"fusion"}})
		}
		rocky, dead, giants, heavy := 0.0, 0, 0, 0
		for j, p := range sys.Planets {
			switch p.Kind {
			case galaxy.Rock, galaxy.SuperEarth:
				rocky += rockyYield * metals
				if j == sys.Home && (sys.Arch == "superterran" || sys.Arch == "volcanic") {
					rocky += richRockyBonus
				}
				if !p.Temperate {
					dead++
				}
				if p.Tag == galaxy.HeavyTag {
					heavy++
				}
			case galaxy.GasGiant, galaxy.IceGiant:
				giants++
			}
		}
		if rocky > 0 {
			if inside(i, galaxy.Remnant) {
				rocky *= remnantMetalMul
			}
			if inside(i, galaxy.Globular) {
				rocky *= globularMetal
			}
			here = append(here, &Source{Key: "rocky", Name: "the rocky worlds of " + name, Kind: WorldSource, Star: i, Yield: flow.Income{flow.M: rocky}, Needs: []string{"metallurgy"}})
		}
		if dead > 0 {
			here = append(here, &Source{Key: "terraformed", Name: "the remade world of " + name, Kind: WorldSource, Star: i, Yield: flow.Income{flow.O: terraformYield}, Needs: []string{"terraforming"}})
		}
		if heavy > 0 {
			here = append(here, &Source{Key: "heavy", Name: "the heavy ore of " + name, Kind: WorldSource, Star: i, Yield: flow.Income{flow.E: heavyYield * float64(heavy)}, Needs: []string{"atomic"}})
		}
		for range sys.Belts {
			here = append(here, &Source{Key: "belt", Name: "the belt at " + name, Kind: BeltSource, Star: i, Yield: flow.Income{flow.M: beltYield}, Needs: []string{"interplanetary"}})
		}
		if giants > 0 {
			here = append(here, &Source{Key: "giant", Name: "the giants of " + name, Kind: GiantSource, Star: i, Yield: flow.Income{flow.E: giantYield * float64(giants)}, Needs: []string{"fusion"}, With: []string{"orbital_habitats"}})
		}
		if y := starYield[st.Class]; y > 0 {
			if inside(i, galaxy.Globular) {
				y *= globularStar
			}
			here = append(here, &Source{Key: "star", Name: "the light of " + name, Kind: StarSource, Star: i, Yield: flow.Income{flow.E: y}, Needs: []string{"orbital_habitats"}})
		}
		if g.Law.Crowd >= crowded {
			here = append(here, &Source{Key: "comets", Name: "the comets of " + name, Kind: CosmicSource, Star: i, Yield: flow.Income{flow.O: cometYield}, Needs: []string{"interplanetary"}})
		}
		for _, s := range here {
			at[i] = append(at[i], add(s).ID)
		}
	}
	// ranged: a nebula feeds, a doomed giant lights, every star in reach
	for _, f := range galaxy.Features {
		var s *Source
		switch f.Kind {
		case galaxy.Nebula:
			s = &Source{Key: "nebula", Name: f.Name, Kind: CosmicSource, Star: -1, Feature: f, Radius: f.Radius, Yield: flow.Income{flow.O: nebulaYield}, Needs: []string{"synthetic_biology"}, With: []string{"orbital_habitats"}}
		case galaxy.Giant:
			s = &Source{Key: "doomed_giant", Name: f.Name, Kind: CosmicSource, Star: -1, Feature: f, Radius: f.Radius, Yield: flow.Income{flow.E: doomedYield}, Needs: []string{"orbital_habitats"}}
		default:
			continue
		}
		var covers []int
		for i := range g.Stars {
			if f.Pos.Dist(g.Position(i)) <= f.Radius {
				covers = append(covers, i)
			}
		}
		if len(covers) == 0 {
			continue
		}
		add(s)
		for _, i := range covers {
			at[i] = append(at[i], s.ID)
		}
	}
	return out, at
}

// harnessed says whether a people knows what a source needs.
func (c *Civ) harnessed(s *Source) bool {
	if s.Cradle && s.Star == c.Cradle {
		return true
	}
	if len(s.Needs) > 0 {
		ok := false
		for _, k := range s.Needs {
			if c.Known[k] {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	for _, k := range s.With {
		if !c.Known[k] {
			return false
		}
	}
	return true
}

// yieldAt is what one star gives a people this tick: every source there
// that it has harnessed, with the profile's cradle multiplier on the
// world it arose on and the swarm's half on its nests.
func (w *World) yieldAt(c *Civ, star int) flow.Income {
	var in flow.Income
	for _, id := range w.sourcesAt[star] {
		s := w.Sources[id]
		if !c.harnessed(s) {
			continue
		}
		y := s.Yield
		if s.Cradle {
			if star == c.Cradle {
				y = y.Scale(c.Species.Profile().Cradle)
			}
			if c.Has("swarming") {
				y = y.Scale(0.5) // a nest is a small thing
			}
		}
		in.Add(y)
	}
	return in
}

// income is what a people takes in this tick: the sources at every
// holding, and for a parasite riding hosts, the hosts' income too.
func (w *World) income(c *Civ) flow.Income {
	var in flow.Income
	for _, s := range w.holdings(c) {
		in.Add(w.yieldAt(c, s))
	}
	if len(c.Ridden) > 0 {
		for _, h := range sortedInts(c.Ridden) {
			if o := w.Civs[h]; o.Living() && o.Master == c.ID {
				in.Add(o.Income)
			}
		}
	}
	return in
}
