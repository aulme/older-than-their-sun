package mind

import (
	"strings"

	"worldgen/internal/flow"
)

// The direction: the order a people feeds its uses in when the means run
// short. Fields, works, mind, road, arms by default, bent by war, fear,
// hunger, greed and the shape of the people.

// DirectionInput is what the direction reads.
type DirectionInput struct {
	AtWar       bool
	Fear        float64
	HostileNear func() bool // a hostile neighbour in reach; asked only when fear is high
	Hunger      float64
	Greed       float64
	NoFields    bool          // a planetary mind, a machine: nothing to farm
	Nomad       bool          // aloft: no works, and the road first
	Fixed       bool          // the people is fixed on one good
	Fixation    flow.Category // and this is what it feeds first
	Honour      string        // where the word (what contracts owe) sits: the faithful pay before their works, the practical after, the faithless last
}

// Direction is the order and the reasons for it.
type Direction struct {
	Order     flow.Order
	ArmsFirst bool
	RoadFirst bool
	MindFirst bool // the mind before the works
	RoadAhead bool // the road before the mind
	FixFirst  bool // the fixation's category before everything
}

// Why explains the order.
func (d Direction) Why() string {
	var why []string
	if d.ArmsFirst {
		why = append(why, "arms first: threatened")
	}
	if d.RoadFirst {
		why = append(why, "the road first: aloft")
	}
	if d.MindFirst {
		why = append(why, "the mind before the works: hungry")
	}
	if d.RoadAhead {
		why = append(why, "the road before the mind: greedy")
	}
	if d.FixFirst {
		why = append(why, "the fixation first")
	}
	if len(why) == 0 {
		return "nothing pressing"
	}
	return strings.Join(why, "; ")
}

// Direct decides the order. The word, what contracts owe, sits by honour:
// the faithful pay before their works go dark, the practical after, the
// faithless last of all. War or fear with a hostile neighbour in reach
// puts arms first; hunger puts the mind before the works; greed the road
// before the mind; a people with no fields drops them; a nomad drops the
// works and puts the road first, after arms if arms come first. A
// fixation puts its category first before every other rule; only a fleet
// in flight, which flow.Direct keeps, comes before it.
func Direct(in DirectionInput, t *Tuning) Direction {
	p := &t.Direction
	d := Direction{}
	order := append(flow.Order(nil), flow.DefaultOrder...)
	switch in.Honour {
	case Faithful:
		order = before(append(order, flow.Word), flow.Word, flow.Works)
	case Faithless:
		order = append(order, flow.Word)
	default:
		order = before(append(order, flow.Word), flow.Word, flow.Mind)
	}
	if in.Hunger > p.HungerBar {
		d.MindFirst = true
		order = before(order, flow.Mind, flow.Works)
	}
	if in.Greed > p.GreedBar {
		d.RoadAhead = true
		order = before(order, flow.Road, flow.Mind)
	}
	if in.Nomad {
		d.RoadFirst = true
		order = without(order, flow.Works)
		order = before(order, flow.Road, order[0])
	}
	if in.AtWar || in.Fear > p.FearBar && in.HostileNear != nil && in.HostileNear() {
		d.ArmsFirst = true
		order = before(order, flow.Arms, order[0])
	}
	if in.NoFields {
		order = without(order, flow.Fields)
	}
	if in.Fixed && index(order, in.Fixation) >= 0 {
		d.FixFirst = true
		order = before(order, in.Fixation, order[0])
	}
	d.Order = order
	return d
}

// before moves a so that it sits just before b, if a is after b.
func before(o flow.Order, a, b flow.Category) flow.Order {
	ia, ib := index(o, a), index(o, b)
	if a == b || ia < 0 || ib < 0 || ia < ib {
		return o
	}
	o = without(o, a)
	ib = index(o, b)
	return append(o[:ib], append(flow.Order{a}, o[ib:]...)...)
}

func without(o flow.Order, c flow.Category) flow.Order {
	out := flow.Order{}
	for _, x := range o {
		if x != c {
			out = append(out, x)
		}
	}
	return out
}

func index(o flow.Order, c flow.Category) int {
	for i, x := range o {
		if x == c {
			return i
		}
	}
	return -1
}
