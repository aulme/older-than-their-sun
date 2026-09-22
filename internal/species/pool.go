package species

import "math/rand/v2"

// The eldritch pool. An eldritch people has no tree; in place of research
// it draws powers from this table, a few at birth and more as the ages
// pass (history's eldritch.go). Each power names the domain it counts as
// a level in, grants what a node would grant as a profile delta, and
// brings what a node would bring: the miracle it stands for, with that
// miracle's filter, or a hook the sim calls by key. Nothing is ever
// forgotten, since there is nothing to forget; a filter's decline can
// strip the power that brought it.

// Power is one entry of the pool.
type Power struct {
	Key, Name string
	Domain    string  // the tech domain it counts as a level in
	Weight    float64 // its weight in the draw
	Grants    Profile // what it adds to the people's profile
	Node      string  // the miracle node it stands for, "" for none; the node brings its own filter
	Filter    string  // a filter it brings without a node, "" for none
	Line      string  // the deepening, as the legends say it; {S} is the people
	Sense     string  // the portrait's phrase: "the sight", "a second presence"
}

// Pool is the table, in the order the proposal lists it; the words and
// the plain numbers of each power are its row in data/kinds.json, read
// by the registry's init, and what it grants is here.
var Pool = []*Power{
	{Key: "thought", Grants: Profile{Wis: 1, Soc: 1}},
	{Key: "presence", Grants: Profile{Range: 10, Sur: 0.5, Mil: 0.5}},
	{Key: "reach", Grants: Profile{Range: 20, Sur: 0.5, Mil: 0.5}},
	{Key: "door", Grants: Profile{Soc: 0.5, Era: 4}},
	{Key: "sight", Grants: Profile{Era: 4}},
	{Key: "voice", Grants: Profile{}},
	{Key: "unmaking", Grants: Profile{}},
	{Key: "shell", Grants: Profile{HomeDefence: 3, Mil: 1.5}},
	{Key: "tithe", Grants: Profile{Sur: 1}},
	{Key: "sleep", Grants: Profile{Sur: 1.5, Dormant: true}},
	{Key: "mirror", Grants: Profile{Soc: 1}},
	{Key: "hunger", Grants: Profile{Expand: 5, Sur: 1}},
	{Key: "making", Grants: Profile{Allows: Works, Cradle: 2, Sur: 1}},
	{Key: "wound", Grants: Profile{Mil: 1, Soc: -0.5, Era: 4}},
}

var powerByKey = map[string]*Power{}

// PowerByKey finds a power, or nil.
func PowerByKey(key string) *Power { return powerByKey[key] }

// HasPower says whether the species holds a power.
func (s *Species) HasPower(key string) bool {
	for _, k := range s.Powers {
		if k == key {
			return true
		}
	}
	return false
}

// AddPower gives the species a power it does not have; ok is false if it
// had it or the key is unknown.
func (s *Species) AddPower(key string) bool {
	if powerByKey[key] == nil || s.HasPower(key) {
		return false
	}
	s.Powers = append(s.Powers, key)
	s.prof = nil
	return true
}

// StripPower takes a power away.
func (s *Species) StripPower(key string) {
	for i, k := range s.Powers {
		if k == key {
			s.Powers = append(s.Powers[:i], s.Powers[i+1:]...)
			s.prof = nil
			return
		}
	}
}

// DrawPower draws a power the species does not have, by weight; nil when
// the pool is exhausted.
func (s *Species) DrawPower(r *rand.Rand) *Power {
	var open []*Power
	for _, p := range Pool {
		if !s.HasPower(p.Key) {
			open = append(open, p)
		}
	}
	if len(open) == 0 {
		return nil
	}
	return pickWeighted(r, open, func(p *Power) float64 { return p.Weight })
}

// PowerNames lists the held powers as the legends say them.
func (s *Species) PowerNames() []string {
	var out []string
	for _, k := range s.Powers {
		if p := powerByKey[k]; p != nil {
			out = append(out, p.Name)
		}
	}
	return out
}

// PowerPortrait is the portrait's sentence on the powers, "" for none:
// "It has the sight and the long sleep."
func (s *Species) PowerPortrait() string { return powerPortrait(s.Powers) }

func powerPortrait(powers []string) string {
	var ns []string
	for _, k := range powers {
		if p := powerByKey[k]; p != nil {
			ns = append(ns, p.Name)
		}
	}
	switch len(ns) {
	case 0:
		return ""
	case 1:
		return "It has " + ns[0] + "."
	}
	return "It has " + joinAnd(ns) + "."
}

func joinAnd(ns []string) string {
	out := ""
	for i, n := range ns {
		switch {
		case i == 0:
		case i == len(ns)-1:
			out += " and "
		default:
			out += ", "
		}
		out += n
	}
	return out
}
