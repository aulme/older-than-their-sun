package galaxy

import (
	"math"

	"worldgen/data"
)

// Features are the named things of the real galaxy: the great hole at the
// centre, the arms and the bar, the nurseries and the clusters, the dead
// stars and the ones about to die, the globular clusters of the halo, and
// the two clouds beyond the rim. A region of the galaxy is described by
// what it lies near, and the laws of the place are bent by them. Beyond
// their reach they are only sky.
//
// Positions are heliocentric galactic coordinates (l, b in degrees, d in
// kpc) as catalogued, converted at start. Distances are the commonly
// quoted ones; a few are argued over by a factor of two. It does not
// matter here.

type FeatureKind int

const (
	BlackHole FeatureKind = iota
	NeutronStar
	Magnetar
	Remnant // a supernova remnant
	Nebula  // a star-forming region
	Cluster // a young massive cluster or association
	Globular
	Giant     // a massive star near its end
	Structure // a large-scale feature: bar, bubbles, waves, rings
	Sky       // beyond the galaxy: only ever seen
)

// kindKeys are the kinds' keys in data/features.json, in order; the
// names the file gives them are kindNames.
var kindKeys = []string{"black_hole", "neutron_star", "magnetar", "remnant", "nebula", "cluster", "globular", "giant", "structure", "sky"}

var kindNames = map[FeatureKind]string{}

func (k FeatureKind) String() string { return kindNames[k] }

// Feature is one catalogued thing, a row of data/features.json.
type Feature struct {
	Key      string      `json:"key"`
	Name     string      `json:"name"`
	KindKey  string      `json:"kind"`
	Kind     FeatureKind `json:"-"`
	L, B, D  float64     // heliocentric galactic, degrees and kpc
	Pos      Vec         `json:"-"`               // galactocentric, computed
	Radius   float64     `json:"radius"`          // kpc, its physical extent or the range of its influence
	Strength float64     `json:"strength"`        // relative weight for the laws: 1 is ordinary
	When     int64       `json:"when,omitempty"`  // years before the present of a known event, 0 if none
	Event    string      `json:"event,omitempty"` // what the sky did then, if When is set
	Desc     string      `json:"desc"`            // how a people living near it would tell it
	Fact     string      `json:"fact"`            // what is actually known
}

// Reach is how far away the feature still counts as near.
func (f *Feature) Reach() float64 {
	switch f.Kind {
	case Structure:
		return math.Max(1.0, 1.5*f.Radius)
	case Globular, Magnetar, BlackHole:
		return 1.5
	case Sky:
		return math.Inf(1)
	}
	return 1.0
}

// Features is the catalogue, in the file's order.
var Features []*Feature

var byName = map[string]*Feature{}

type featuresFile struct {
	Kinds []struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"kinds"`
	Features []*Feature `json:"features"`
}

func init() {
	var f featuresFile
	data.Load("features.json", &f)
	if len(f.Kinds) != len(kindKeys) {
		panic("galaxy: the file's feature kinds are not the code's")
	}
	kindOf := map[string]FeatureKind{}
	for i, k := range f.Kinds {
		if k.Key != kindKeys[i] {
			panic("galaxy: the file's feature kind " + k.Key + " is not the code's " + kindKeys[i])
		}
		kindNames[FeatureKind(i)] = k.Name
		kindOf[k.Key] = FeatureKind(i)
	}
	Features = f.Features
	for _, f := range Features {
		k, ok := kindOf[f.KindKey]
		if !ok {
			panic("galaxy: feature " + f.Key + " has an unknown kind " + f.KindKey)
		}
		f.Kind = k
		f.Pos = FromSun(f.L, f.B, f.D)
		byName[f.Name] = f
	}
}

// FeatureByName looks a feature up, case-insensitively, by name or by a
// distinctive part of it.
func FeatureByName(name string) *Feature {
	n := lower(name)
	for _, f := range Features {
		if lower(f.Name) == n {
			return f
		}
	}
	for _, f := range Features {
		if contains(lower(f.Name), n) {
			return f
		}
	}
	return nil
}

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
