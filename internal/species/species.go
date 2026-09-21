// Package species generates a species: a substrate (what it is made of),
// modifiers (how it is shaped), a home world that sets physical facts and
// base traits, and random traits layered on top. Substrates and modifiers
// are registry entries (registry.go, one file each); the generator rolls
// them as a chain of tilted rolls (generate.go).
//
// Traits are few and asymmetric: they move the three levels a little, tilt
// research toward some domains, and (in the history package) change the
// difficulty of particular filters. A species should read in three to five
// words in the legends.
package species

import (
	"strings"
)

// M is a domain weight map, keyed by tech domain name.
type M = map[string]float64

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
	Miracle       string                // tech node key of the miracle this species is born to
	Quiet         bool                  // the common case of its group; the legends do not say it
	Flavour       *Flavour              // vocabulary the trait brings, over the substrate's; nil for none
	Fit           func(s *Species) bool // rolled only for a species this is true of; nil for any
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
	// biology quirks; swarming is drawn by its own tilted roll (TraitRolls), never from the group
	{Key: "swarming", Name: "a swarm, a million small bodies that think as one when they gather", Group: "bio", Mil: 0.5, Sur: 0.5, Soc: 0.5, Domains: M{"computation": 0.8, "industry": 1.2}, Flavour: &Flavour{"nest", "seed-cloud", "hive-moon"}},
	{Key: "shortlived", Name: "short-lived", Group: "bio", Weight: 20, Soc: -0.5, Rate: 1.2},
	{Key: "longlived", Name: "very long-lived", Group: "bio", Weight: 20, Soc: 0.5, Rate: 0.85},
	{Key: "dormancy", Name: "given to cyclical dormancy", Group: "bio", Weight: 15, Sur: 1, Soc: 0.5, Rate: 0.9},
	{Key: "manysexes", Name: "of many sexes", Group: "bio", Weight: 15, Soc: 0.5, Domains: M{"biology": 1.3}},
	{Key: "sessile", Name: "sessile as adults", Group: "bio", Weight: 10, Reach: 0.6, Mil: -0.5, Sur: 0.5, Domains: M{"society": 1.2}},
	{Key: "amphibious", Name: "amphibious", Group: "bio", Weight: 10, Sur: 0.5},
	{Key: "eusocial", Name: "eusocial", Group: "bio", Weight: 10, Soc: 0.5, Mil: 0.5},
	{Key: "symbiosis", Name: "bonded to their machines", Group: "bio", Weight: 8, Mil: 0.5, Domains: M{"computation": 1.5}},
	{Key: "memory", Name: "of unbroken memory across generations", Group: "bio", Weight: 8, Soc: 1, Domains: M{"society": 1.2}},
	{Key: "kinfed", Name: "who eat their own; a caste is bred for the table", Group: "bio", Weight: 4, Soc: -0.5,
		Fit: func(s *Species) bool { return s.Sub == Biological && (s.Has("caste") || s.Is(Hive)) }},
	// senses: what they have beyond, or instead of, the usual five; the
	// eldritch draw from a table of senses nobody else has
	{Key: "eyeless", Name: "eyeless, who see by sound", Group: "sense", Weight: 6, Domains: M{"exotic": 0.8}, Fit: bodily},
	{Key: "deaf", Name: "without hearing", Group: "sense", Weight: 4, Fit: bodily},
	{Key: "thermal", Name: "who see heat", Group: "sense", Weight: 12, Fit: bodily},
	{Key: "magnetic", Name: "who feel the pull of the world", Group: "sense", Weight: 12, Domains: M{"propulsion": 1.1}, Fit: bodily},
	{Key: "electric", Name: "who sense the living current", Group: "sense", Weight: 10, Domains: M{"energy": 1.1}, Fit: bodily},
	{Key: "radiation", Name: "who feel radiation as warmth", Group: "sense", Weight: 6, Domains: M{"energy": 1.1, "exotic": 1.1}, Fit: bodily},
	{Key: "chemical", Name: "who taste the world in the air", Group: "sense", Weight: 12, Domains: M{"biology": 1.1}, Fit: bodily},
	{Key: "timesight", Name: "who see time", Group: "sense", Weight: 10, Fit: eldritchOnly},
	{Key: "masssense", Name: "who feel mass", Group: "sense", Weight: 10, Fit: eldritchOnly},
	{Key: "lighthearing", Name: "who hear light", Group: "sense", Weight: 8, Fit: eldritchOnly},
	{Key: "fartaste", Name: "who taste distance", Group: "sense", Weight: 8, Fit: eldritchOnly},
	{Key: "deathsense", Name: "who feel every death within reach", Group: "sense", Weight: 6, Fit: eldritchOnly},
	// what a parasite rides: only parasites get one
	{Key: "bodyrider", Name: "who live in the flesh of others", Group: "rider", Weight: 60},
	{Key: "mindrider", Name: "who live as an idea in the minds of others", Group: "rider", Weight: 40, Domains: M{"society": 1.2, "computation": 1.1}},
	// the way: nomads take to the sky when they can
	{Key: "nomadic", Name: "nomads, who will not stay", Group: "way", Weight: 1, Domains: M{"propulsion": 1.3, "industry": 0.8}},
	// the seat of a hive: only hives get one; a nomad hive's throne moves
	{Key: "onequeen", Name: "of one queen", Group: "seat", Weight: 50},
	{Key: "noqueen", Name: "of no queen", Group: "seat", Weight: 50},
	{Key: "throne", Name: "of a moving throne", Group: "seat"},
	// born to a miracle
	{Key: "born_voice", Name: "minds that speak across any distance", Group: "power", Weight: 25, Miracle: "ansible"},
	{Key: "born_flesh", Name: "masters of their own flesh", Group: "power", Weight: 25, Miracle: "directed_evolution"},
	{Key: "born_sight", Name: "touched by foresight", Group: "power", Weight: 20, Miracle: "foresight"},
	{Key: "born_chorus", Name: "whose thought takes root in any mind", Group: "power", Weight: 15, Miracle: "chorus"},
	{Key: "born_door", Name: "who walk between the stars", Group: "power", Weight: 15, Miracle: "ftl"},
	{Key: "born_manna", Name: "who keep something that feeds them", Group: "power", Weight: 10, Miracle: "manna"},
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
	{Key: "table", Name: "grown for the table", Group: "made", Soc: -0.5, Sur: 0.5},
}

// bodily fits a species with a body the usual senses could be in; the
// eldritch have their own table.
func bodily(s *Species) bool { return s.Sub != Eldritch }

func eldritchOnly(s *Species) bool { return s.Sub == Eldritch }

var byKey = map[string]*Trait{}

func init() {
	for _, t := range Traits {
		byKey[t.Key] = t
	}
}

// Get returns a trait by key.
func Get(key string) *Trait { return byKey[key] }

// Species is a people's blood: what it is made of, how it is shaped, where
// it arose and what it is like. The world holds one entry per species and
// several peoples may share one; a people changes species only by a made
// path (an uplift, a machine successor, a remaking, the Brood's change),
// which makes a new entry with the old as Parent.
type Species struct {
	ID      int
	Sub     Substrate
	Channel string // how the people communicates: sound, light, electromagnetic, chemical, touch, none; see Channels
	Mods    Mod
	Powers  []string // the eldritch pool's powers held, by key, in the order gained; see pool.go
	World   *Archetype
	Traits  []*Trait
	Made    string   // who made them, "" if they arose naturally
	Parent  *Species // the species this one was made from, nil if none

	profSub  Substrate
	profMods Mod
	prof     *Profile // cleared by AddPower and StripPower
}

// Is says whether the species carries every modifier in m.
func (s *Species) Is(m Mod) bool { return s.Mods.Has(m) }

// entries lists the carried registry entries: the substrate, then the
// modifiers in registry order.
func (s *Species) entries() []*Entry {
	out := []*Entry{&s.Sub.Def().Entry}
	for _, d := range s.Mods.Defs() {
		out = append(out, &d.Entry)
	}
	return out
}

// Profile is the composed profile of the carried entries and the powers
// held, cached.
func (s *Species) Profile() Profile {
	if s.prof == nil || s.profSub != s.Sub || s.profMods != s.Mods {
		var ps []Profile
		for _, e := range s.entries() {
			ps = append(ps, e.Profile)
		}
		for _, k := range s.Powers {
			if p := powerByKey[k]; p != nil {
				ps = append(ps, p.Grants)
			}
		}
		p := Compose(ps...)
		s.prof, s.profSub, s.profMods = &p, s.Sub, s.Mods
	}
	return *s.prof
}

// Flavour is the vocabulary: the substrate's, overridden by each modifier's
// in order and then by any trait that brings its own.
func (s *Species) Flavour() Flavour {
	f := s.Sub.Def().Flavour
	over := func(o Flavour) {
		if o.Colony != "" {
			f.Colony = o.Colony
		}
		if o.Ship != "" {
			f.Ship = o.Ship
		}
		if o.Station != "" {
			f.Station = o.Station
		}
	}
	for _, d := range s.Mods.Defs() {
		over(d.Flavour)
	}
	for _, t := range s.Traits {
		if t.Flavour != nil {
			over(*t.Flavour)
		}
	}
	return f
}

// Portrait is the sentences the portrait opens with: the substrate's, then
// one per modifier, then the powers if any; none for a plain biological
// people.
func (s *Species) Portrait() []string {
	var out []string
	for _, e := range s.entries() {
		if e.Portrait != "" {
			out = append(out, e.Portrait)
		}
	}
	if p := s.PowerPortrait(); p != "" {
		out = append(out, p)
	}
	if s.Sub == Biological && !s.Is(Hive) && s.Channel != Sound && s.Channel != "" {
		if d := ChannelDesc[s.Channel]; d != "" {
			out = append(out, strings.ToUpper(d[:1])+d[1:]+".")
		}
	}
	return out
}

// Channels are the ways a people communicates. The channel decides the
// voice the names pass gives the people: sound is transcribed, anything
// else translated. It is rolled at generation from the substrate and the
// senses (rollChannel) and shown in the portrait through the lookup.
const (
	Sound           = "sound"
	Light           = "light"
	Electromagnetic = "electromagnetic"
	Chemical        = "chemical"
	Touch           = "touch"
	NoChannel       = "none"
)

// Channels lists them, for the tables.
var Channels = []string{Sound, Light, Electromagnetic, Chemical, Touch, NoChannel}

// ChannelDesc is how the portrait says a channel other than sound.
var ChannelDesc = map[string]string{
	Light:           "they speak in light",
	Electromagnetic: "they speak in pulses of the electromagnetic field",
	Chemical:        "they speak in scents",
	Touch:           "they speak by touch",
	NoChannel:       "they have no language",
}

// Arising is how the legends say the people came to be: "arise on" for the
// plain case.
func (s *Species) Arising() string {
	if a := s.Sub.Def().Arising; a != "" {
		return a
	}
	return "arise on"
}

// Nature names the substrate and the modifiers: "biological", "machine, hive".
func (s *Species) Nature() string {
	if s.Mods == 0 {
		return s.Sub.String()
	}
	return s.Sub.String() + ", " + s.Mods.String()
}

// Kin says whether two species are the same blood: the same entry, or one
// made from the other, or both made from the same one.
func (s *Species) Kin(o *Species) bool {
	return s == o || s.Parent == o || o.Parent == s || (s.Parent != nil && s.Parent == o.Parent)
}

// Branch is a copy of the species with this one as Parent: the same blood
// changed, for a people that is made from it or that changes.
func (s *Species) Branch() *Species {
	b := *s
	b.ID, b.Parent, b.prof = 0, s, nil
	b.Traits = append([]*Trait(nil), s.Traits...)
	b.Powers = append([]string(nil), s.Powers...)
	return &b
}

// Fixed makes a known species for tests: a biological people of a lush
// world with exactly the traits named, in that order, and no roll. Unknown
// keys are ignored.
func Fixed(traits ...string) *Species {
	s := &Species{World: ArchetypeByKey("lush"), Channel: Sound}
	for _, t := range traits {
		s.Add(t)
	}
	return s
}

// Add gives the species a trait by key if it does not have it.
func (s *Species) Add(key string) {
	if t := byKey[key]; t != nil && !s.Has(key) {
		s.Traits = append(s.Traits, t)
	}
}

// Replace swaps one trait for another in place, keeping the order; nothing
// if the old one is not there.
func (s *Species) Replace(old, key string) {
	t := byKey[key]
	if t == nil {
		return
	}
	for i, x := range s.Traits {
		if x.Key == old {
			s.Traits[i] = t
			return
		}
	}
}

// Remove takes a trait away by key; nothing if it is not there.
func (s *Species) Remove(key string) {
	for i, x := range s.Traits {
		if x.Key == key {
			s.Traits = append(s.Traits[:i:i], s.Traits[i+1:]...)
			return
		}
	}
}

// Of lists the species' traits of a group, in the order they were gained.
func (s *Species) Of(group string) []*Trait {
	var out []*Trait
	for _, t := range s.Traits {
		if t.Group == group {
			out = append(out, t)
		}
	}
	return out
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

// Base returns the base levels from world, profile and traits.
func (s *Species) Base() (mil, sur, soc float64) {
	p := s.Profile()
	mil, sur, soc = s.World.Mil+p.Mil, s.World.Sur+p.Sur, s.World.Soc+p.Soc
	for _, t := range s.Traits {
		mil, sur, soc = mil+t.Mil, sur+t.Sur, soc+t.Soc
	}
	return
}

// ReachMul is the species' multiplier on reach.
func (s *Species) ReachMul() float64 {
	m := s.Profile().Reach
	for _, t := range s.Traits {
		if t.Reach > 0 {
			m *= t.Reach
		}
	}
	return m
}

// Rate is the species' multiplier on research speed.
func (s *Species) Rate() float64 {
	m := s.Profile().Rate
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
	if v, ok := s.Profile().Dom[d]; ok {
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
