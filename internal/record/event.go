package record

import (
	"bytes"
	"encoding/json"
)

// Kind is what an event is: a key of data/events.json, where its
// parameters, its meaning and, for a fact, its sort and weight are
// declared. The constants are generated from the file (kinds_gen.go):
// F-kinds are facts, the ones a people can hold a tale of; K-kinds are
// the rest of the chronicle.
type Kind string

// P is an event's parameters as JSON holds them: numbers are
// json.Number (the literal, so an integer stays one), lists are []any,
// objects map[string]any. The accessors on Event read them by the type
// the kind declares.
type P map[string]any

// UnmarshalJSON reads the parameters with numbers kept as literals.
func (p *P) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	m := map[string]any{}
	if err := dec.Decode(&m); err != nil {
		return err
	}
	*p = m
	return nil
}

// Event is one thing that happened, as it happened: a record of
// chronicle.jsonl.
type Event struct {
	ID      int  `json:"id"`
	Year    Year `json:"year"`
	Kind    Kind `json:"kind"`
	Subject int  `json:"subject"` // the people it is about, or -1
	Object  int  `json:"object"`  // the other people, or -1
	Star    int  `json:"star"`    // where, or -1
	Legacy  int  `json:"legacy"`  // the remain in it, or -1
	Plague  int  `json:"plague"`  // the plague in it, or -1
	N       int  `json:"n"`       // a count: worlds
	P       P    `json:"p"`       // the kind's parameters
	Fact    bool `json:"fact"`    // the kind has a sort and a weight: a people can hold a tale of it
}

// Has says whether a parameter is set.
func (e *Event) Has(k string) bool { _, ok := e.P[k]; return ok }

// Int reads an integer parameter; a missing one is zero.
func (e *Event) Int(k string) int { return toInt(e.P[k]) }

// Float reads a number.
func (e *Event) Float(k string) float64 { return toFloat(e.P[k]) }

// Str reads a string or a key; a missing one is "".
func (e *Event) Str(k string) string { s, _ := e.P[k].(string); return s }

// Bool reads a flag; a missing one is false.
func (e *Event) Bool(k string) bool { b, _ := e.P[k].(bool); return b }

// Year reads a year.
func (e *Event) YearOf(k string) Year { return Year(toInt64(e.P[k])) }

// Ints reads a list of integers.
func (e *Event) Ints(k string) []int {
	xs, _ := e.P[k].([]any)
	out := make([]int, 0, len(xs))
	for _, x := range xs {
		out = append(out, toInt(x))
	}
	return out
}

// Strs reads a list of strings.
func (e *Event) Strs(k string) []string {
	xs, _ := e.P[k].([]any)
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		s, _ := x.(string)
		out = append(out, s)
	}
	return out
}

// Obj reads an object parameter into v, as JSON would.
func (e *Event) Obj(k string, v any) bool {
	m, ok := e.P[k]
	if !ok {
		return false
	}
	b, err := json.Marshal(m)
	if err != nil {
		return false
	}
	return json.Unmarshal(b, v) == nil
}

func toInt(v any) int { return int(toInt64(v)) }

func toInt64(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	case int8:
		return int64(x)
	case json.Number:
		n, _ := x.Int64()
		return n
	}
	return 0
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		f, _ := x.Float64()
		return f
	}
	return 0
}

// Normalise turns a parameter map of Go-typed values into what JSON
// holds: every value goes through a marshal and an unmarshal, so that
// a record built in memory reads as one read from the file.
func Normalise(p map[string]any) P {
	if p == nil {
		return P{}
	}
	b, err := json.Marshal(p)
	if err != nil {
		panic("record: parameters: " + err.Error())
	}
	var out P
	if err := out.UnmarshalJSON(b); err != nil {
		panic("record: parameters: " + err.Error())
	}
	return out
}

// Tale is one thing a people holds of what happened: a record of
// tellings.jsonl. A testament is the same record with Remain set, frozen
// when the remain was left.
type Tale struct {
	ID       int     `json:"id"`
	Civ      int     `json:"civ"`
	Remain   int     `json:"remain"` // the remain whose testament it is, or -1 for a living memory
	Fact     int     `json:"fact"`
	Learned  Year    `json:"learned"`
	Source   string  `json:"source"` // witnessed, told, read, inherited
	From     int     `json:"from"`
	Slant    int8    `json:"slant"`
	Wear     int8    `json:"wear"`
	Blamed   int     `json:"blamed"`
	Revised  int8    `json:"revised"`
	Forgot   bool    `json:"forgot"`
	Believed *Year   `json:"believed_year,omitempty"` // exact at wear 0, rough at wear 1, absent at myth
	Sort     string  `json:"sort"`                    // as this teller judges it: deed, crime, woe, bond, folly, nothing
	Weight   float64 `json:"weight"`                  // as this teller weighs it
	Rank     int     `json:"rank"`                    // its place among the teller's dearest, 1 first; 0 for a forgotten tale
	Frozen   *Frozen `json:"frozen,omitempty"`        // for a testament: what the maker's telling read of the world when it wrote
}

// Frozen is what a testament's telling read of the world at the moment
// it was written, of the parties in the tale (its subject, object and
// blamed) and its star: which the maker could hold in mind, which it
// had met, which were still rising, which still lived; and whether the
// star was its own. Everything else a telling reads is the tale's own
// fields and the names of the maker's time.
type Frozen struct {
	Perceived []int `json:"perceived"`
	Met       []int `json:"met"`
	Active    []int `json:"active"`
	Living    []int `json:"living"`
	Ours      bool  `json:"ours"`
}
