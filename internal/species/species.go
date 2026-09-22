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

	"worldgen/data"
)

// M is a domain weight map, keyed by tech domain name.
type M = map[string]float64

// Archetype is a home world type, a row of data/worlds.json.
type Archetype struct {
	Key     string   `json:"key"`
	Desc    string   `json:"desc"` // "a temperate, lush world"
	Weight  float64  `json:"weight"`
	Mil     float64  `json:"mil"` // base levels
	Sur     float64  `json:"sur"`
	Soc     float64  `json:"soc"`
	Domains M        `json:"domains,omitempty"` // research tilt
	Traits  []string `json:"traits,omitempty"`  // traits every species from this world gets
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
var Archetypes []*Archetype

// Trait is one facet of a species, a row of data/traits.json.
type Trait struct {
	Key     string   `json:"key"`
	Name    string   `json:"name"`  // how the legends say it
	Group   string   `json:"group"` // org, stance, honour, drive, bio, sense, power, world, made
	Weight  float64  `json:"weight,omitempty"`
	Mil     float64  `json:"mil,omitempty"`
	Sur     float64  `json:"sur,omitempty"`
	Soc     float64  `json:"soc,omitempty"`
	Reach   float64  `json:"reach,omitempty"` // multiplier, 0 means 1
	Rate    float64  `json:"rate,omitempty"`  // research rate multiplier, 0 means 1
	Domains M        `json:"domains,omitempty"`
	Miracle string   `json:"miracle,omitempty"` // tech node key of the miracle this species is born to
	Quiet   bool     `json:"quiet,omitempty"`   // the common case of its group; the legends do not say it
	Flavour *Flavour `json:"flavour,omitempty"` // vocabulary the trait brings, over the substrate's; nil for none
	Fit     string   `json:"fit,omitempty"`     // rolled only for a species the named predicate is true of (fits); "" for any
}

// Traits is the pool. Groups org, stance, honour and drive are common and
// every species gets one of each; bio is uncommon; power is very rare and means
// the species is born to a miracle, with no tree beneath it and no filter
// on it.
var Traits []*Trait

// fits are the predicates a trait's Fit names.
var fits = map[string]func(s *Species) bool{
	"bodily":   bodily,
	"eldritch": eldritchOnly,
	"kinfed":   func(s *Species) bool { return s.Sub == Biological && (s.Has("caste") || s.Is(Hive)) },
}

// fits says whether the species may roll the trait.
func (t *Trait) fits(s *Species) bool {
	if t.Fit == "" {
		return true
	}
	return s != nil && fits[t.Fit](s)
}

// bodily fits a species with a body the usual senses could be in; the
// eldritch have their own table.
func bodily(s *Species) bool { return s.Sub != Eldritch }

func eldritchOnly(s *Species) bool { return s.Sub == Eldritch }

var byKey = map[string]*Trait{}

// groupRoll is one trait group in the order the chain draws them, with
// the chance of drawing one at all; an own group is drawn only for a
// people whose entries list it.
type groupRoll struct {
	Group  string  `json:"group"`
	Chance float64 `json:"chance"`
	Own    bool    `json:"own,omitempty"`
}

var groupRolls []groupRoll

// ChannelDesc is how the portrait says a channel other than sound.
var ChannelDesc = map[string]string{}

type traitsFile struct {
	Groups   []groupRoll `json:"groups"`
	Traits   []*Trait    `json:"traits"`
	Channels []struct {
		Key  string `json:"key"`
		Desc string `json:"desc,omitempty"`
	} `json:"channels"`
}

type worldsFile struct {
	Archetypes []*Archetype `json:"archetypes"`
}

func init() {
	var tf traitsFile
	data.Load("traits.json", &tf)
	groupRolls, Traits = tf.Groups, tf.Traits
	for _, t := range Traits {
		if t.Fit != "" && fits[t.Fit] == nil {
			panic("species: trait " + t.Key + " names an unknown predicate " + t.Fit)
		}
		byKey[t.Key] = t
	}
	if len(tf.Channels) != len(Channels) {
		panic("species: the file's channels are not the code's")
	}
	for i, ch := range tf.Channels {
		if ch.Key != Channels[i] {
			panic("species: the file's channel " + ch.Key + " is not the code's " + Channels[i])
		}
		if ch.Desc != "" {
			ChannelDesc[ch.Key] = ch.Desc
		}
	}
	var wf worldsFile
	data.Load("worlds.json", &wf)
	Archetypes = wf.Archetypes
	for _, a := range Archetypes {
		for _, k := range a.Traits {
			if byKey[k] == nil {
				panic("species: world " + a.Key + " gives an unknown trait " + k)
			}
		}
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
	Made    Making   // how they were made, when they did not arise; the zero value for a cradle blood
	Parent  *Species // the species this one was made from, nil if none

	profSub  Substrate
	profMods Mod
	prof     *Profile // cleared by AddPower and StripPower
}

// Making is how a blood or a people came to be when it did not arise:
// a key of data/origins.json and the parties, as ids the history
// resolves (a people, the people it was made from, a remain, a plague);
// -1 for none. The zero key is a cradle blood.
type Making struct {
	Key    string `json:"key"`
	By     int    `json:"by"`
	From   int    `json:"from"`
	Legacy int    `json:"legacy"`
	Plague int    `json:"plague"`
}

// Made is a making with no parties.
func Made(key string) Making { return Making{Key: key, By: -1, From: -1, Legacy: -1, Plague: -1} }

// MadeBy is a making by a people.
func MadeBy(key string, by int) Making {
	return Making{Key: key, By: by, From: -1, Legacy: -1, Plague: -1}
}

// Is says whether the species carries every modifier in m.
func (s *Species) Is(m Mod) bool { return s.Mods.Has(m) }

// Voiceless says whether the people has no language a human could render:
// a hive, an unconscious people, a living world, a replicator, an eldritch
// thing. It coins no names and is known only by what others call it; the
// names pass and the view read this, the simulation never.
func (s *Species) Voiceless() bool {
	return s.Is(Hive) || s.Is(Unconscious) || s.Is(Planetary) || s.Is(Replicator) || s.Sub == Eldritch
}

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
func (s *Species) Portrait() []string { return s.PortraitWith(s.Powers) }

// PortraitWith is the portrait with the powers given in place of the
// ones held now: a record captures the keys at its time.
func (s *Species) PortraitWith(powers []string) []string {
	var out []string
	for _, e := range s.entries() {
		if e.Portrait != "" {
			out = append(out, e.Portrait)
		}
	}
	if p := powerPortrait(powers); p != "" {
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

// TraitKeys lists the traits as Describe says them, by key: the ones a
// line names first, then the world-given, in order; a record captures
// it at the time, since a blood can drift.
func (s *Species) TraitKeys() []string {
	var out []string
	for _, t := range s.Traits {
		if t.Group == "world" || t.Group == "made" || t.Quiet {
			continue
		}
		out = append(out, t.Key)
	}
	for _, t := range s.Traits {
		if t.Group == "world" {
			out = append(out, t.Key)
		}
	}
	return out
}

// DescribeTraits is Describe over captured keys.
func DescribeTraits(keys []string) string {
	var parts []string
	for _, k := range keys {
		parts = append(parts, byKey[k].Name)
	}
	return strings.Join(parts, ", ")
}

// ModKeys lists a species' modifiers by key, in the registry's order.
func (s *Species) ModKeys() []string {
	var out []string
	for _, d := range s.Mods.Defs() {
		out = append(out, d.Key)
	}
	return out
}

// Rebuild makes a species from the keys a record holds: the substrate,
// the modifiers, the channel, the powers, the world and the traits, and
// its making. The parent is set by the caller once every blood exists.
// A key the registry lacks panics, since the record and the binary
// disagree.
func Rebuild(id int, sub string, mods []string, channel string, powers []string, world string, traits []string, made Making) *Species {
	s := &Species{ID: id, Channel: channel, Made: made}
	var ok bool
	if s.Sub, ok = SubstrateByKey(sub); !ok {
		panic("species: unknown substrate " + sub)
	}
	for _, k := range mods {
		m, ok := ModByKey(k)
		if !ok {
			panic("species: unknown modifier " + k)
		}
		s.Mods |= m
	}
	if world != "" {
		if s.World = ArchetypeByKey(world); s.World == nil {
			panic("species: unknown world " + world)
		}
	}
	for _, k := range traits {
		t := byKey[k]
		if t == nil {
			panic("species: unknown trait " + k)
		}
		s.Traits = append(s.Traits, t)
	}
	s.Powers = append(s.Powers, powers...)
	return s
}
