package flow

import (
	"encoding/json"
	"fmt"
)

// The tables in data/ write an income as an object of the kinds it has,
// by symbol ({"E": 1, "M": 2}), and a category by its name.

// MarshalJSON writes the kinds that are not zero, by symbol.
func (a Income) MarshalJSON() ([]byte, error) {
	m := map[string]float64{}
	for k := range Kinds {
		if a[k] != 0 {
			m[Kind(k).Symbol()] = a[k]
		}
	}
	return json.Marshal(m)
}

// UnmarshalJSON reads an object by symbol; a missing kind is zero.
func (a *Income) UnmarshalJSON(b []byte) error {
	var m map[string]float64
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	*a = Income{}
	for s, v := range m {
		found := false
		for k := range Kinds {
			if Kind(k).Symbol() == s {
				a[k], found = v, true
			}
		}
		if !found {
			return fmt.Errorf("flow: unknown kind %q", s)
		}
	}
	return nil
}

// MarshalJSON writes the category's name.
func (c Category) MarshalJSON() ([]byte, error) { return json.Marshal(c.String()) }

// UnmarshalJSON reads a category by name.
func (c *Category) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	for i := Fields; i <= Word; i++ {
		if i.String() == s {
			*c = i
			return nil
		}
	}
	return fmt.Errorf("flow: unknown category %q", s)
}

// MarshalText writes a kind by its symbol, as an income's keys are.
func (k Kind) MarshalText() ([]byte, error) { return []byte(k.Symbol()), nil }

// UnmarshalText reads a kind by its symbol.
func (k *Kind) UnmarshalText(b []byte) error {
	for _, x := range Kinds {
		if x.Symbol() == string(b) {
			*k = x
			return nil
		}
	}
	return fmt.Errorf("flow: unknown kind %q", b)
}
