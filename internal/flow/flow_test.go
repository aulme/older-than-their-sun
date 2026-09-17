package flow

import (
	"reflect"
	"testing"
)

func use(key string, cat Category, era int, o, e, m float64) Use {
	return Use{Key: key, Cat: cat, Era: era, Need: Income{o, e, m}}
}

func TestDirectFillsInOrder(t *testing.T) {
	uses := []Use{
		use("weapons", Arms, 1, 0, 0, 1),
		use("farms", Fields, 1, 1, 0, 0),
		use("mills", Works, 1, 0, 1, 0),
	}
	a := Direct(Income{1, 1, 1}, uses, DefaultOrder)
	if want := []string{"farms", "mills", "weapons"}; !reflect.DeepEqual(a.Working, want) {
		t.Fatalf("working %v, want %v", a.Working, want)
	}
	if len(a.Dormant) != 0 || a.Surplus != (Income{}) || a.Want != (Income{}) {
		t.Fatalf("dormant %v surplus %v want %v", a.Dormant, a.Surplus, a.Want)
	}
	// arms first: the weapons take the metal
	a = Direct(Income{1, 1, 1}, append(uses, use("forges", Works, 2, 0, 0, 1)), Order{Arms, Fields, Works})
	if want := []string{"weapons", "farms", "mills"}; !reflect.DeepEqual(a.Working, want) {
		t.Fatalf("arms first: working %v, want %v", a.Working, want)
	}
	if want := []string{"forges"}; !reflect.DeepEqual(a.Dormant, want) {
		t.Fatalf("arms first: dormant %v, want %v", a.Dormant, want)
	}
}

func TestDirectShedsNewestFirst(t *testing.T) {
	uses := []Use{
		use("late", Works, 3, 0, 1, 0),
		use("early", Works, 1, 0, 1, 0),
		use("middle", Works, 2, 0, 1, 0),
		use("small", Works, 4, 0, 0.1, 0),
	}
	a := Direct(Income{0, 2.1, 0}, uses, DefaultOrder)
	// the late use fails and takes nothing; the small one after it is fed
	// from what is left
	if want := []string{"early", "middle", "small"}; !reflect.DeepEqual(a.Working, want) {
		t.Fatalf("working %v, want %v", a.Working, want)
	}
	if want := []string{"late"}; !reflect.DeepEqual(a.Dormant, want) {
		t.Fatalf("dormant %v, want %v", a.Dormant, want)
	}
	// a use short of one kind is shed while another kind's uses go on
	a = Direct(Income{5, 0, 0}, []Use{use("x", Works, 1, 0, 1, 0), use("y", Works, 2, 1, 0, 0)}, DefaultOrder)
	if want := []string{"y"}; !reflect.DeepEqual(a.Working, want) {
		t.Fatalf("mixed kinds: working %v, want %v", a.Working, want)
	}
	// a category the order leaves out is not fed at all
	a = Direct(Income{0, 9, 0}, uses, Order{Fields, Mind})
	if len(a.Working) != 0 || len(a.Dormant) != 4 {
		t.Fatalf("unnamed category: working %v dormant %v", a.Working, a.Dormant)
	}
}

func TestDirectKeepsFleetsInFlight(t *testing.T) {
	fleet := Use{Key: "fleet", Cat: Arms, Era: 3, Need: Income{0, 1, 1}, Flight: true}
	uses := []Use{
		use("farms", Fields, 1, 1, 0, 0),
		use("mills", Works, 1, 0, 1, 1),
		use("thought", Mind, 2, 0, 1, 0),
		fleet,
	}
	// arms are last, but the fleet is fed right after the fields
	a := Direct(Income{1, 1, 1}, uses, DefaultOrder)
	if want := []string{"farms", "fleet"}; !reflect.DeepEqual(a.Working, want) {
		t.Fatalf("working %v, want %v", a.Working, want)
	}
	// only the fields come before it: with no organic matter the farms fail
	// and the fleet still flies
	a = Direct(Income{0, 1, 1}, uses, DefaultOrder)
	if want := []string{"fleet"}; !reflect.DeepEqual(a.Working, want) {
		t.Fatalf("no fields fed: working %v, want %v", a.Working, want)
	}
	// and the fields do come first: with energy for one thing only the farms
	// take nothing of it, so the fleet has it; with metal short the fleet is shed
	a = Direct(Income{1, 1, 0}, uses, DefaultOrder)
	if want := []string{"fleet"}; !reflect.DeepEqual(a.Dormant[:1], want) {
		t.Fatalf("metal short: dormant %v, want fleet first", a.Dormant)
	}
	// an order with no fields feeds the fleet first of all
	a = Direct(Income{0, 1, 1}, uses, Order{Works, Mind, Arms})
	if want := []string{"fleet"}; !reflect.DeepEqual(a.Working, want) {
		t.Fatalf("no fields in order: working %v, want %v", a.Working, want)
	}
	// a fleet is never cut by another arms use failing
	a = Direct(Income{0, 3, 3}, []Use{use("guns", Arms, 1, 0, 9, 0), fleet}, Order{Arms})
	if want := []string{"fleet"}; !reflect.DeepEqual(a.Working, want) {
		t.Fatalf("arms failure: working %v, want %v", a.Working, want)
	}
}

func TestDirectSurplusAndWantAddUp(t *testing.T) {
	uses := []Use{
		use("a", Fields, 1, 1, 0, 0),
		use("b", Works, 2, 0, 2, 1),
		use("c", Works, 3, 0, 2, 1),
		use("d", Mind, 3, 0, 1, 0),
		use("e", Road, 4, 0, 1, 3),
	}
	income := Income{2, 3, 1}
	a := Direct(income, uses, DefaultOrder)
	var working, total Income
	for i := range uses {
		total.Add(uses[i].Need)
		for _, k := range a.Working {
			if k == uses[i].Key {
				working.Add(uses[i].Need)
			}
		}
	}
	for k := range income {
		if working[k]+a.Surplus[k] != income[k] {
			t.Fatalf("kind %s: working %v + surplus %v != income %v", Kind(k), working[k], a.Surplus[k], income[k])
		}
		if want := max(0, total[k]-income[k]); a.Want[k] != want {
			t.Fatalf("kind %s: want %v, expected %v", Kind(k), a.Want[k], want)
		}
	}
	if len(a.Working)+len(a.Dormant) != len(uses) {
		t.Fatalf("every use is working or dormant: %v %v", a.Working, a.Dormant)
	}
}
