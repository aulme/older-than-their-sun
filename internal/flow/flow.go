// Package flow is the allocation of a people's means: three commodities as
// per-tick income, a set of uses each with an upkeep, and an order of
// categories that says what is fed first. Direct fills the uses in that
// order and sheds the rest. It is pure: it takes values and returns values,
// and knows nothing of worlds, stars or peoples.
package flow

import "sort"

// Kind of commodity.
type Kind int

const (
	O Kind = iota // organic matter
	E             // energy
	M             // metal
)

// Kinds in order, for loops that must not range over a map.
var Kinds = [3]Kind{O, E, M}

func (k Kind) String() string {
	return [...]string{"organic matter", "energy", "metal"}[k]
}

// Symbol is the one-letter name.
func (k Kind) Symbol() string { return [...]string{"O", "E", "M"}[k] }

// Income is a yield per tick by kind, indexed by Kind.
type Income [3]float64

// Add sums another income into this one.
func (a *Income) Add(b Income) {
	for k := range a {
		a[k] += b[k]
	}
}

// Scale multiplies every kind.
func (a Income) Scale(m float64) Income {
	for k := range a {
		a[k] *= m
	}
	return a
}

// Total is the sum over kinds.
func (a Income) Total() float64 { return a[O] + a[E] + a[M] }

// Category is what a use is for. The order of categories is the direction:
// what a people feeds first when the income runs short.
type Category int

const (
	Fields Category = iota // what keeps people alive
	Arms                   // weapons, defences, fleets
	Works                  // industry and energy, and the structures that harness sources
	Mind                   // computation, society, the exotic
	Road                   // propulsion, colony ships, surveyors
	Word                   // contracts and messages; added at a later step
)

func (c Category) String() string {
	return [...]string{"fields", "arms", "works", "mind", "road", "word"}[c]
}

// Phrase is how the category reads in a sentence: "to keep the fleets fed".
func (c Category) Phrase() string {
	return [...]string{"the fields", "the fleets", "the works", "the mind", "the road", "the word"}[c]
}

// Order is the categories in the order they are fed. A category left out
// is not fed at all.
type Order []Category

// DefaultOrder is the order with nothing pressing.
var DefaultOrder = Order{Fields, Works, Mind, Road, Arms}

// Has says whether the order feeds a category.
func (o Order) Has(c Category) bool {
	for _, x := range o {
		if x == c {
			return true
		}
	}
	return false
}

// Use is one thing a people spends its means on.
type Use struct {
	Key    string
	Name   string // for the line when it goes dark
	Cat    Category
	Era    int
	Need   Income
	Flight bool // a fleet in flight: shed for nothing but the fields
}

// Allocation is what Direct decided: what works, what is dormant, what is
// left over by kind, and what more by kind would run everything.
type Allocation struct {
	Working []string // in the order they were fed
	Dormant []string // in the order they were shed
	Surplus Income   // income less the needs of the working uses
	Want    Income   // the needs of every use less the income, floored at zero
}

// Direct fills the uses from the income in the order's sequence of
// categories, lowest era first within a category, and sheds what cannot
// be fed when its turn comes: the newest things go dark first, and only
// for the kind that is short, since a use that fails takes nothing and the
// next is tried. A use in flight is fed as soon as the fields are, or first
// of all when the order has no fields. A category the order does not name
// is not fed.
func Direct(income Income, uses []Use, order Order) Allocation {
	byCat := map[Category][]int{}
	for i := range uses {
		byCat[uses[i].Cat] = append(byCat[uses[i].Cat], i)
	}
	for _, idx := range byCat {
		sort.SliceStable(idx, func(a, b int) bool { return uses[idx[a]].Era < uses[idx[b]].Era })
	}
	left := income
	done := make([]bool, len(uses))
	out := Allocation{}
	feed := func(i int) bool {
		u := &uses[i]
		done[i] = true
		for k := range left {
			if u.Need[k] > left[k]+1e-9 {
				out.Dormant = append(out.Dormant, u.Key)
				return false
			}
		}
		for k := range left {
			left[k] -= u.Need[k]
		}
		out.Working = append(out.Working, u.Key)
		return true
	}
	flights := func() {
		for _, c := range []Category{Fields, Arms, Works, Mind, Road, Word} {
			for _, i := range byCat[c] {
				if uses[i].Flight && !done[i] {
					feed(i)
				}
			}
		}
	}
	if !order.Has(Fields) {
		flights()
	}
	for _, c := range order {
		for _, i := range byCat[c] {
			if !done[i] {
				feed(i)
			}
		}
		if c == Fields {
			flights()
		}
	}
	for i := range uses {
		if !done[i] {
			out.Dormant = append(out.Dormant, uses[i].Key)
		}
	}
	out.Surplus = left
	var total Income
	for i := range uses {
		total.Add(uses[i].Need)
	}
	for k := range total {
		out.Want[k] = max(0, total[k]-income[k])
	}
	return out
}
