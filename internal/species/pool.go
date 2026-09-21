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

// Pool is the table, in the order the proposal lists it.
var Pool = []*Power{
	{Key: "thought", Name: "the long thought", Domain: "society", Weight: 10, Grants: Profile{Wis: 1, Soc: 1},
		Line: "{S} have begun to think about something, and the thinking will take an age."},
	{Key: "presence", Name: "a second presence", Domain: "propulsion", Weight: 10, Grants: Profile{Range: 10, Sur: 0.5, Mil: 0.5},
		Line: "There is another of {S} now, at a star nothing was seen to cross to."},
	{Key: "reach", Name: "the wide reach", Domain: "propulsion", Weight: 8, Grants: Profile{Range: 20, Sur: 0.5, Mil: 0.5},
		Line: "{S} can feel further than they could."},
	{Key: "door", Name: "the door in it", Domain: "exotic", Weight: 6, Grants: Profile{Soc: 0.5, Era: 4}, Node: "ftl",
		Line: "There is a door in {S} now, and things pass through it without a ship."},
	{Key: "sight", Name: "the sight", Domain: "exotic", Weight: 7, Grants: Profile{Era: 4}, Node: "foresight",
		Line: "{S} have begun to see every star within their reach as if it were under them."},
	{Key: "voice", Name: "the voice", Domain: "computation", Weight: 7, Grants: Profile{}, Node: "ansible",
		Line: "{S} have begun to speak, and the speaking crosses any distance at once."},
	{Key: "unmaking", Name: "the unmaking", Domain: "weapons", Weight: 5, Grants: Profile{}, Node: "unmaking",
		Line: "{S} have learned to end a world without touching it."},
	{Key: "shell", Name: "the shell", Domain: "weapons", Weight: 8, Grants: Profile{HomeDefence: 3, Mil: 1.5},
		Line: "{S} have grown a shell, and nothing that comes for them gets in."},
	{Key: "tithe", Name: "the tithe", Domain: "industry", Weight: 7, Grants: Profile{Sur: 1},
		Line: "{S} have begun to take a share of every harvest within their reach. Nobody agreed to it."},
	{Key: "sleep", Name: "the long sleep", Domain: "biology", Weight: 8, Grants: Profile{Sur: 1.5, Dormant: true},
		Line: "{S} have learned to sleep, and will sleep when there is nothing left they want."},
	{Key: "mirror", Name: "the mirror", Domain: "society", Weight: 7, Grants: Profile{Soc: 1},
		Line: "{S} have begun to answer every message in the sender's own voice."},
	{Key: "hunger", Name: "the hunger", Domain: "biology", Weight: 6, Grants: Profile{Expand: 5, Sur: 1}, Filter: "overshoot",
		Line: "{S} have begun to grow."},
	{Key: "making", Name: "the making", Domain: "industry", Weight: 7, Grants: Profile{Allows: Works, Cradle: 2, Sur: 1},
		Line: "{S} have begun to build."},
	{Key: "wound", Name: "the wound", Domain: "exotic", Weight: 5, Grants: Profile{Mil: 1, Soc: -0.5, Era: 4},
		Line: "There is a wound in {S}, and what is underneath everything is open there."},
}

var powerByKey = map[string]*Power{}

func init() {
	for _, p := range Pool {
		powerByKey[p.Key] = p
	}
}

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
func (s *Species) PowerPortrait() string {
	ns := s.PowerNames()
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
