package plague

import (
	"hash/fnv"
	"math/rand/v2"
	"strconv"

	"worldgen/data"
)

// Profile is what a sickness is, as keys of data/plagues.json: for one of
// the body, the symptoms in the order they come, the onset, the course
// and whom it takes; for one of the mind, its form, its effects in the
// order they come and what carries it. It is drawn at the plague's birth
// by a hash of the seed and the plague's id, never from the history's
// rolls, so it costs the history nothing and is a property of the thing.
// The names pass reads it; the flavour is the list read in order.
type Profile struct {
	Symptoms []string `json:"symptoms,omitempty"`
	Onset    string   `json:"onset,omitempty"`
	Course   string   `json:"course,omitempty"`
	Takes    string   `json:"takes,omitempty"`
	Form     string   `json:"form,omitempty"`
	Effects  []string `json:"effects,omitempty"`
	Carrier  string   `json:"carrier,omitempty"`
}

// Symptom is one row of the symptoms table.
type Symptom struct {
	Key     string `json:"key"`
	Noun    string `json:"noun"`
	Adj     string `json:"adj"`
	Visible bool   `json:"visible"`
	End     string `json:"end"` // deadly, wasting, either: which bank the last symptom comes from
	Desc    string `json:"desc"`
}

// Form is one row of the forms table.
type Form struct {
	Key    string `json:"key"`
	Noun   string `json:"noun"`
	Hosted bool   `json:"hosted"` // wants a host: a conscious plague takes one of these
	Desc   string `json:"desc"`
}

// Effect is one row of the effects table.
type Effect struct {
	Key      string `json:"key"`
	Severity int    `json:"severity"`
	Desc     string `json:"desc"`
}

// Keyed is a row with a key and a description.
type Keyed struct {
	Key  string `json:"key"`
	Adj  string `json:"adj,omitempty"`
	Desc string `json:"desc"`
}

// Table is data/plagues.json.
type Table struct {
	Symptoms []Symptom `json:"symptoms"`
	Onsets   []Keyed   `json:"onsets"`
	Courses  []Keyed   `json:"courses"`
	Takes    []Keyed   `json:"takes"`
	Forms    []Form    `json:"forms"`
	Effects  []Effect  `json:"effects"`
	Carriers []Keyed   `json:"carriers"`
}

// Profiles is the table, loaded once.
var Profiles = loadProfiles()

func loadProfiles() *Table {
	t := &Table{}
	data.Load("plagues.json", t)
	return t
}

// SymptomOf finds a symptom row, or nil.
func (t *Table) SymptomOf(key string) *Symptom {
	for i := range t.Symptoms {
		if t.Symptoms[i].Key == key {
			return &t.Symptoms[i]
		}
	}
	return nil
}

// FormOf finds a form row, or nil.
func (t *Table) FormOf(key string) *Form {
	for i := range t.Forms {
		if t.Forms[i].Key == key {
			return &t.Forms[i]
		}
	}
	return nil
}

// EffectOf finds an effect row, or nil.
func (t *Table) EffectOf(key string) *Effect {
	for i := range t.Effects {
		if t.Effects[i].Key == key {
			return &t.Effects[i]
		}
	}
	return nil
}

// ProfileOf draws the profile of a plague from the seed and its id.
// Lethality picks the last symptom's bank and the list's length;
// contagion the onset; an engineered plague keeps the visible marks
// out; a tailored one takes one blood. A memetic plague that thinks
// takes a form that wants a host; lethality picks the last effect's
// severity and the list's length. The rest is the hash's.
func ProfileOf(seed uint64, id int, p Plague) Profile {
	h := fnv.New64a()
	h.Write([]byte(strconv.FormatUint(seed, 10) + "|plague|" + strconv.Itoa(id)))
	s := h.Sum64()
	r := rand.New(rand.NewPCG(s, s^0x9e3779b97f4a7c15))
	t := Profiles
	if p.Kind == Memetic {
		return memeticProfile(r, t, p)
	}
	n := 1 + int(p.Lethality*3.99) // one to four, by lethality
	var pool []Symptom
	for _, sy := range t.Symptoms {
		if p.Engineered && sy.Visible {
			continue
		}
		pool = append(pool, sy)
	}
	end := "either"
	switch {
	case p.Lethality > 0.6:
		end = "deadly"
	case p.Lethality < 0.3:
		end = "wasting"
	}
	var last []Symptom
	for _, sy := range pool {
		if sy.End == end || (end == "either" && sy.End != "deadly") {
			last = append(last, sy)
		}
	}
	pr := Profile{}
	taken := map[string]bool{}
	final := last[r.IntN(len(last))]
	taken[final.Key] = true
	for len(pr.Symptoms) < n-1 {
		sy := pool[r.IntN(len(pool))]
		if taken[sy.Key] {
			continue
		}
		taken[sy.Key] = true
		pr.Symptoms = append(pr.Symptoms, sy.Key)
	}
	pr.Symptoms = append(pr.Symptoms, final.Key)
	pr.Onset = "slow"
	if p.Contagion > 0.5 {
		pr.Onset = "quick"
	}
	pr.Course = t.Courses[r.IntN(len(t.Courses))].Key
	switch {
	case p.Band >= 0:
		pr.Takes = "species"
	default:
		pr.Takes = []string{"young", "old", "everyone", "everyone", "caste"}[r.IntN(5)]
	}
	return pr
}

func memeticProfile(r *rand.Rand, t *Table, p Plague) Profile {
	var forms []Form
	for _, f := range t.Forms {
		if !p.Conscious || f.Hosted {
			forms = append(forms, f)
		}
	}
	pr := Profile{Form: forms[r.IntN(len(forms))].Key}
	n := 1 + int(p.Lethality*2.99) // one to three
	sev := 1 + int(p.Lethality*2.99)
	var last []Effect
	for _, e := range t.Effects {
		if e.Severity == sev {
			last = append(last, e)
		}
	}
	final := last[r.IntN(len(last))]
	taken := map[string]bool{final.Key: true}
	for len(pr.Effects) < n-1 {
		e := t.Effects[r.IntN(len(t.Effects))]
		if taken[e.Key] {
			continue
		}
		taken[e.Key] = true
		pr.Effects = append(pr.Effects, e.Key)
	}
	pr.Effects = append(pr.Effects, final.Key)
	pr.Carrier = t.Carriers[r.IntN(len(t.Carriers))].Key
	return pr
}

// Visible is the first visible mark among the symptoms, or "".
func (pr Profile) Visible() string {
	for _, k := range pr.Symptoms {
		if sy := Profiles.SymptomOf(k); sy != nil && sy.Visible {
			return k
		}
	}
	return ""
}

// Last is the last symptom or effect: what the sickness comes to.
func (pr Profile) Last() string {
	if len(pr.Symptoms) > 0 {
		return pr.Symptoms[len(pr.Symptoms)-1]
	}
	if len(pr.Effects) > 0 {
		return pr.Effects[len(pr.Effects)-1]
	}
	return ""
}

// First is the first symptom or effect: what is noticed.
func (pr Profile) First() string {
	if len(pr.Symptoms) > 0 {
		return pr.Symptoms[0]
	}
	if len(pr.Effects) > 0 {
		return pr.Effects[0]
	}
	return ""
}
