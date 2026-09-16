// Package species generates a species from its home world: the world sets
// physical facts and base traits, random traits are layered on top, and a
// body-plan kind sets the flavour of everything the species builds.
//
// Traits are few and asymmetric: they move the three levels a little, tilt
// research toward some domains, and (in the history package) change the
// difficulty of particular filters. A species should read in three to five
// words in the legends.
package species

import (
	"math/rand/v2"
	"strings"

	"worldgen/internal/names"
)

// M is a domain weight map, keyed by tech domain name.
type M = map[string]float64

// Kind is the body plan.
type Kind uint8

const (
	Standard Kind = iota
	Swarm
	PlanetaryMind
	Parasite
	MachineBorn
	Evolver
)

func (k Kind) String() string {
	return [...]string{"standard", "swarm", "planetary mind", "parasite", "machine-born", "evolver"}[k]
}

// Flavour is what a kind calls the things it builds.
type Flavour struct {
	Portrait string // "They are a swarm..."
	Colony   string
	Ship     string
	Station  string
}

var flavours = [...]Flavour{
	Standard:      {"", "colony", "colony ship", "station"},
	Swarm:         {"They are a swarm, a million small bodies that think as one when they gather.", "nest", "seed-cloud", "hive-moon"},
	PlanetaryMind: {"They are one mind, spread through the living substance of their world.", "graft", "spore-ark", "living moon"},
	Parasite:      {"They are a parasite, and need the bodies of others to think and to build.", "host-world", "carrier", "hive"},
	MachineBorn:   {"They are machines, and do not remember who built them.", "node", "probe", "array"},
	Evolver:       {"They shape their own flesh, and breed what they need instead of building it.", "brood", "vacuum-whale", "grown moon"},
}

// Flavour returns the kind's vocabulary.
func (k Kind) Flavour() Flavour { return flavours[k] }

// Machine-born peoples never arise on their own: they are what is left when
// someone else builds a mind that outgrows them.
var kindWeights = [...]float64{Standard: 74, Swarm: 7, PlanetaryMind: 4, Parasite: 4, MachineBorn: 0, Evolver: 8}

// kind modifiers: levels and domain tilt
var kindMods = [...]struct {
	mil, sur, soc, reach float64
	dom                  M
}{
	Standard:      {0, 0, 0, 1, nil},
	Swarm:         {0.5, 0.5, 0.5, 1, M{"computation": 0.8, "industry": 1.2}},
	PlanetaryMind: {-1, 1, 2, 0.3, M{"propulsion": 0.4, "biology": 1.5, "society": 1.3}},
	Parasite:      {0, 1, 0, 1, M{"biology": 1.4, "society": 1.2, "industry": 0.8}},
	MachineBorn:   {0.5, 1.5, 0, 1, M{"computation": 1.3, "biology": 0.6, "industry": 1.1}},
	Evolver:       {0, 1, 0, 1, M{"biology": 1.5, "industry": 0.8}},
}

// Archetype is a home world type.
type Archetype struct {
	Key           string
	Desc          string // "a temperate, lush world"
	Weight        float64
	Mil, Sur, Soc float64  // base levels
	Domains       M        // research tilt
	Traits        []string // traits every species from this world gets
}

// ArchetypeByKey finds an archetype, or nil.
func ArchetypeByKey(key string) *Archetype {
	for _, a := range Archetypes {
		if a.Key == key {
			return a
		}
	}
	return nil
}

// Archetypes is the home world table, lush most common, extreme rare.
var Archetypes = []*Archetype{
	{Key: "lush", Desc: "a temperate, lush world", Weight: 30, Mil: 1, Sur: 2, Soc: 2},
	{Key: "ocean", Desc: "an ocean world with scattered islands", Weight: 15, Mil: 0.5, Sur: 2, Soc: 3, Domains: M{"industry": 0.6, "propulsion": 0.7, "biology": 1.3}, Traits: []string{"cooperative"}},
	{Key: "arid", Desc: "an arid world of salt flats and canyons", Weight: 12, Mil: 1.5, Sur: 3, Soc: 1.5, Domains: M{"energy": 1.2, "biology": 0.8}},
	{Key: "twilight", Desc: "a tidally locked world, habitable only along its twilight band", Weight: 10, Mil: 1, Sur: 3, Soc: 2, Domains: M{"society": 1.2, "energy": 1.1}, Traits: []string{"hardy"}},
	{Key: "superterran", Desc: "a heavy world of crushing gravity", Weight: 8, Mil: 2, Sur: 3.5, Soc: 2, Domains: M{"propulsion": 0.5, "industry": 1.2}, Traits: []string{"robust"}},
	{Key: "lowg", Desc: "a small, light world with a thin sky", Weight: 6, Mil: 0.5, Sur: 1, Soc: 2, Domains: M{"propulsion": 1.5, "exotic": 1.1}, Traits: []string{"fragile"}},
	{Key: "hothouse", Desc: "a hothouse world under a crushing, poisonous sky", Weight: 5, Mil: 1, Sur: 3.5, Soc: 2, Domains: M{"biology": 1.2, "exotic": 0.8, "propulsion": 0.7}},
	{Key: "iceshell", Desc: "an ocean sealed beneath a shell of ice", Weight: 5, Mil: 0.5, Sur: 3, Soc: 2.5, Domains: M{"exotic": 0.4, "propulsion": 0.4, "biology": 1.3, "society": 1.2}, Traits: []string{"skyless", "fireless"}},
	{Key: "floater", Desc: "the cloud decks of a gas giant", Weight: 3, Mil: 0.5, Sur: 2, Soc: 2, Domains: M{"biology": 1.4, "energy": 0.6}, Traits: []string{"fireless"}},
	{Key: "volcanic", Desc: "a moon kneaded by tides, all fire and sulphur", Weight: 3, Mil: 1.5, Sur: 3, Soc: 1.5, Domains: M{"energy": 1.3, "industry": 1.2}, Traits: []string{"hardy"}},
	{Key: "dim", Desc: "a dim world huddled close to a brown dwarf companion", Weight: 3, Mil: 0.5, Sur: 3.5, Soc: 2.5, Domains: M{"exotic": 1.2, "energy": 0.8}, Traits: []string{"hardy"}},
}

// Trait is one facet of a species.
type Trait struct {
	Key           string
	Name          string // how the legends say it
	Group         string // org, stance, honour, drive, bio, sense, power, world, made
	Weight        float64
	Mil, Sur, Soc float64
	Reach         float64 // multiplier, 0 means 1
	Rate          float64 // research rate multiplier, 0 means 1
	Domains       M
	Miracle       string // tech node key of the miracle this species is born to
	Quiet         bool   // the common case of its group; the legends do not say it
}

// Traits is the pool. Groups org, stance, honour and drive are common and
// every species gets one of each; bio is uncommon; power is very rare and means
// the species is born to a miracle, with no tree beneath it and no filter
// on it.
var Traits = []*Trait{
	// social organisation
	{Key: "solitary", Name: "solitary by nature", Group: "org", Weight: 6, Soc: -1.5, Mil: 0.5, Domains: M{"society": 0.7, "computation": 1.1}},
	{Key: "individualist", Name: "individualists", Group: "org", Weight: 26, Soc: -1, Mil: 0.5, Domains: M{"computation": 1.1, "energy": 1.1}},
	{Key: "collective", Name: "a collective people", Group: "org", Weight: 30, Soc: 1, Domains: M{"society": 1.1}},
	{Key: "herd", Name: "a herd people, who move as one", Group: "org", Weight: 10, Soc: 1.5, Mil: -0.5, Domains: M{"society": 1.1, "weapons": 0.9}},
	{Key: "caste", Name: "a caste society", Group: "org", Weight: 15, Soc: 1, Mil: 0.5, Domains: M{"biology": 1.1, "computation": 0.9}},
	{Key: "hive", Name: "a hive mind", Group: "org", Weight: 10, Soc: 2.5, Mil: 0.5, Domains: M{"society": 0.6, "computation": 0.8}},
	{Key: "nonconscious", Name: "an intelligence without consciousness", Group: "org", Weight: 5, Soc: 1.5, Sur: 1, Mil: -0.5, Domains: M{"exotic": 0.6, "society": 0.4, "biology": 1.3}},
	// posture toward others: how a people makes war
	{Key: "pacifist", Name: "pacifists", Group: "stance", Weight: 12, Mil: -1.5, Soc: 1, Domains: M{"weapons": 0.3, "biology": 1.2, "society": 1.2}},
	{Key: "defensive", Name: "who keep to themselves", Group: "stance", Weight: 28},
	{Key: "submissive", Name: "quick to submit", Group: "stance", Weight: 10, Mil: -0.5, Soc: 0.5},
	{Key: "opportunist", Name: "opportunists, who prey on the weak", Group: "stance", Weight: 13, Mil: 0.5, Domains: M{"weapons": 1.1}},
	{Key: "conqueror", Name: "conquerors", Group: "stance", Weight: 13, Mil: 1, Domains: M{"weapons": 1.3, "society": 0.9}},
	{Key: "vengeful", Name: "who never forget a wrong", Group: "stance", Weight: 9, Mil: 0.5, Soc: 0.5},
	{Key: "confederate", Name: "makers of pacts", Group: "stance", Weight: 8, Soc: 0.5, Domains: M{"society": 1.1}},
	{Key: "unyielding", Name: "a people who do not surrender", Group: "stance", Weight: 7, Mil: 1, Sur: -0.5},
	// honour: whether a promise binds
	{Key: "faithful", Name: "true to their word", Group: "honour", Weight: 30},
	{Key: "practical", Name: "practical about promises", Group: "honour", Weight: 50, Quiet: true},
	{Key: "faithless", Name: "faithless", Group: "honour", Weight: 20},
	// drive
	{Key: "curious", Name: "curious", Group: "drive", Weight: 25, Rate: 1.1, Domains: M{"computation": 1.2, "exotic": 1.2}},
	{Key: "expansionist", Name: "expansionist", Group: "drive", Weight: 20, Reach: 1.25, Soc: -0.5, Domains: M{"propulsion": 1.3}},
	{Key: "contemplative", Name: "contemplative", Group: "drive", Weight: 15, Reach: 0.7, Domains: M{"society": 1.3, "exotic": 1.2, "weapons": 0.6, "industry": 0.7}},
	{Key: "xenophobic", Name: "xenophobic", Group: "drive", Weight: 15, Mil: 0.5, Soc: 0.5, Domains: M{"weapons": 1.2}},
	{Key: "cautious", Name: "cautious", Group: "drive", Weight: 15, Sur: 0.5, Domains: M{"computation": 0.8, "exotic": 0.7}},
	{Key: "pragmatic", Name: "pragmatic", Group: "drive", Weight: 10, Domains: M{"industry": 1.2, "energy": 1.1}},
	// biology quirks
	{Key: "shortlived", Name: "short-lived", Group: "bio", Weight: 20, Soc: -0.5, Rate: 1.2},
	{Key: "longlived", Name: "very long-lived", Group: "bio", Weight: 20, Soc: 0.5, Rate: 0.85},
	{Key: "dormancy", Name: "given to cyclical dormancy", Group: "bio", Weight: 15, Sur: 1, Soc: 0.5, Rate: 0.9},
	{Key: "manysexes", Name: "of many sexes", Group: "bio", Weight: 15, Soc: 0.5, Domains: M{"biology": 1.3}},
	{Key: "sessile", Name: "sessile as adults", Group: "bio", Weight: 10, Reach: 0.6, Mil: -0.5, Sur: 0.5, Domains: M{"society": 1.2}},
	{Key: "amphibious", Name: "amphibious", Group: "bio", Weight: 10, Sur: 0.5},
	{Key: "eusocial", Name: "eusocial", Group: "bio", Weight: 10, Soc: 0.5, Mil: 0.5},
	{Key: "symbiosis", Name: "bonded to their machines", Group: "bio", Weight: 8, Mil: 0.5, Domains: M{"computation": 1.5}},
	{Key: "memory", Name: "of unbroken memory across generations", Group: "bio", Weight: 8, Soc: 1, Domains: M{"society": 1.2}},
	// senses: what they have beyond, or instead of, the usual five
	{Key: "eyeless", Name: "eyeless, who see by sound", Group: "sense", Weight: 6, Domains: M{"exotic": 0.8}},
	{Key: "deaf", Name: "without hearing", Group: "sense", Weight: 4},
	{Key: "thermal", Name: "who see heat", Group: "sense", Weight: 12},
	{Key: "magnetic", Name: "who feel the pull of the world", Group: "sense", Weight: 12, Domains: M{"propulsion": 1.1}},
	{Key: "electric", Name: "who sense the living current", Group: "sense", Weight: 10, Domains: M{"energy": 1.1}},
	{Key: "radiation", Name: "who feel radiation as warmth", Group: "sense", Weight: 6, Domains: M{"energy": 1.1, "exotic": 1.1}},
	{Key: "chemical", Name: "who taste the world in the air", Group: "sense", Weight: 12, Domains: M{"biology": 1.1}},
	// what a parasite rides: only parasites get one
	{Key: "bodyrider", Name: "who live in the flesh of others", Group: "rider", Weight: 60},
	{Key: "mindrider", Name: "who live as an idea in the minds of others", Group: "rider", Weight: 40, Domains: M{"society": 1.2, "computation": 1.1}},
	// the way: nomads take to the sky when they can
	{Key: "nomadic", Name: "nomads, who will not stay", Group: "way", Weight: 1, Domains: M{"propulsion": 1.3, "industry": 0.8}},
	// born to a miracle
	{Key: "born_voice", Name: "minds that speak across any distance", Group: "power", Weight: 25, Miracle: "ansible"},
	{Key: "born_flesh", Name: "masters of their own flesh", Group: "power", Weight: 25, Miracle: "directed_evolution"},
	{Key: "born_sight", Name: "touched by foresight", Group: "power", Weight: 20, Miracle: "foresight"},
	{Key: "born_chorus", Name: "whose thought takes root in any mind", Group: "power", Weight: 15, Miracle: "chorus"},
	{Key: "born_door", Name: "who walk between the stars", Group: "power", Weight: 15, Miracle: "ftl"},
	// world-given
	{Key: "cooperative", Name: "cooperative by necessity", Group: "world", Soc: 0.5},
	{Key: "hardy", Name: "hardy", Group: "world", Sur: 0.5},
	{Key: "robust", Name: "heavy-boned and strong", Group: "world", Sur: 0.5, Mil: 0.5},
	{Key: "fragile", Name: "light and fragile", Group: "world", Sur: -0.5},
	{Key: "skyless", Name: "who never saw a sky", Group: "world", Soc: 0.5},
	{Key: "fireless", Name: "who never made fire", Group: "world"},
	{Key: "threesuns", Name: "born under three suns", Group: "world", Soc: 0.5, Domains: M{"exotic": 1.3}},
	// made
	{Key: "uplifted", Name: "uplifted", Group: "made", Soc: -0.5},
	{Key: "bred", Name: "bred to serve", Group: "made", Soc: -1, Sur: 1},
	{Key: "branch", Name: "a branch of an older people", Group: "made"},
}

var byKey = map[string]*Trait{}

func init() {
	for _, t := range Traits {
		byKey[t.Key] = t
	}
}

// Get returns a trait by key.
func Get(key string) *Trait { return byKey[key] }

// Species is a people.
type Species struct {
	Name   string
	Kind   Kind
	World  *Archetype
	Traits []*Trait
	Made   string // who made them, "" if they arose naturally
}

func pickWeighted[T any](r *rand.Rand, xs []T, weight func(T) float64) T {
	total := 0.0
	for _, x := range xs {
		total += weight(x)
	}
	v := r.Float64() * total
	for _, x := range xs {
		v -= weight(x)
		if v < 0 {
			return x
		}
	}
	return xs[len(xs)-1]
}

func pickGroup(r *rand.Rand, group string) *Trait {
	var pool []*Trait
	for _, t := range Traits {
		if t.Group == group {
			pool = append(pool, t)
		}
	}
	return pickWeighted(r, pool, func(t *Trait) float64 { return t.Weight })
}

// Pick draws a trait from a group.
func Pick(r *rand.Rand, group string) *Trait { return pickGroup(r, group) }

// Generate rolls a species. mult is the home star's multiplicity.
func Generate(r *rand.Rand, mult int) *Species { return GenerateOn(r, mult, "") }

// GenerateOn generates a species for a home world of a given archetype
// key, or a random one if the key is empty.
func GenerateOn(r *rand.Rand, mult int, arch string) *Species {
	s := &Species{Name: names.Civ(r)}
	s.Kind = Kind(pickWeighted(r, []int{0, 1, 2, 3, 4, 5}, func(i int) float64 { return kindWeights[i] }))
	s.World = ArchetypeByKey(arch)
	if s.World == nil {
		s.World = pickWeighted(r, Archetypes, func(a *Archetype) float64 { return a.Weight })
		if s.Kind == PlanetaryMind && r.Float64() < 0.6 {
			s.World = Archetypes[1] // living oceans are the usual planetary mind
		}
	}
	for _, t := range s.World.Traits {
		s.Add(t)
	}
	if mult == 3 {
		s.Add("threesuns")
	} else if mult == 2 {
		s.Add("hardy")
	}
	s.Traits = append(s.Traits, pickGroup(r, "org"), pickGroup(r, "stance"), pickGroup(r, "honour"), pickGroup(r, "drive"))
	if r.Float64() < 0.5 {
		s.Traits = append(s.Traits, pickGroup(r, "bio"))
	}
	if r.Float64() < 0.4 {
		s.Traits = append(s.Traits, pickGroup(r, "sense"))
		if r.Float64() < 0.15 {
			s.Add(pickGroup(r, "sense").Key)
		}
	}
	if r.Float64() < 0.006 {
		s.Traits = append(s.Traits, pickGroup(r, "power"))
	}
	if s.Kind == Parasite {
		s.Traits = append(s.Traits, pickGroup(r, "rider"))
	}
	if r.Float64() < 0.08 && s.Kind != PlanetaryMind {
		s.Add("nomadic")
	}
	return s
}

// Add gives the species a trait by key if it does not have it.
func (s *Species) Add(key string) {
	if t := byKey[key]; t != nil && !s.Has(key) {
		s.Traits = append(s.Traits, t)
	}
}

// Has reports whether the species has a trait.
func (s *Species) Has(key string) bool {
	for _, t := range s.Traits {
		if t.Key == key {
			return true
		}
	}
	return false
}

// Miracle returns the node key of the miracle the species is born to, or "".
func (s *Species) Miracle() string {
	for _, t := range s.Traits {
		if t.Miracle != "" {
			return t.Miracle
		}
	}
	return ""
}

// Base returns the base levels from world, kind and traits.
func (s *Species) Base() (mil, sur, soc float64) {
	k := kindMods[s.Kind]
	mil, sur, soc = s.World.Mil+k.mil, s.World.Sur+k.sur, s.World.Soc+k.soc
	for _, t := range s.Traits {
		mil, sur, soc = mil+t.Mil, sur+t.Sur, soc+t.Soc
	}
	return
}

// ReachMul is the species' multiplier on reach.
func (s *Species) ReachMul() float64 {
	m := kindMods[s.Kind].reach
	for _, t := range s.Traits {
		if t.Reach > 0 {
			m *= t.Reach
		}
	}
	return m
}

// Rate is the species' multiplier on research speed.
func (s *Species) Rate() float64 {
	m := 1.0
	for _, t := range s.Traits {
		if t.Rate > 0 {
			m *= t.Rate
		}
	}
	return m
}

// DomainMul is the research tilt toward a domain.
func (s *Species) DomainMul(d string) float64 {
	m := 1.0
	if v, ok := s.World.Domains[d]; ok {
		m *= v
	}
	if v, ok := kindMods[s.Kind].dom[d]; ok {
		m *= v
	}
	for _, t := range s.Traits {
		if v, ok := t.Domains[d]; ok {
			m *= v
		}
	}
	return m
}

// Describe lists the traits as the legends say them: "a hive mind, martial, curious, short-lived".
func (s *Species) Describe() string {
	var parts []string
	for _, t := range s.Traits {
		if t.Group == "world" || t.Group == "made" || t.Quiet {
			continue
		}
		parts = append(parts, t.Name)
	}
	for _, t := range s.Traits {
		if t.Group == "world" {
			parts = append(parts, t.Name)
		}
	}
	return strings.Join(parts, ", ")
}
