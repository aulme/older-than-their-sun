package plague

import (
	"math/rand/v2"
	"sort"
	"testing"
)

// TestLethality: an l = 0.9 plague ends a four-world people within five
// ticks uncured at the median over a hundred runs; an l = 0.05 one lasts
// fifty ticks with no world lost at the median.
func TestLethality(t *testing.T) {
	tn := Default()
	r := rand.New(rand.NewPCG(1, 2))
	runs := func(l float64, ticks int) []int {
		var lost []int
		for range 100 {
			worlds, gone, at := 4, 0, ticks+1
			for tick := 1; tick <= ticks; tick++ {
				for i := 0; i < worlds-gone; i++ {
					if r.Float64() < TollChance(l, 1, false, &tn) {
						gone++
					}
				}
				if gone == worlds && at > ticks {
					at = tick
				}
			}
			if ticks == 50 {
				lost = append(lost, gone)
			} else {
				lost = append(lost, at)
			}
		}
		sort.Ints(lost)
		return lost
	}
	if m := runs(0.9, 20)[50]; m > 5 {
		t.Errorf("an l = 0.9 plague takes %d ticks to end four worlds at the median, want five or fewer", m)
	}
	if m := runs(0.05, 50)[50]; m != 0 {
		t.Errorf("an l = 0.05 plague takes %d worlds in fifty ticks at the median, want none", m)
	}
	if TollChance(0.5, 1, true, &tn) != TollChance(0.5, 1, false, &tn)/2 {
		t.Error("containment does not halve the toll")
	}
	if TollChance(0.5, 2, false, &tn) != 1 {
		t.Error("a frail body at l 0.5 does not lose a world each tick")
	}
}

// TestBirth: a cradle's birth rate against an era-3 people's with the
// ladder paid is about two hundred and fifty to one; the top of the
// ladder bears none; dirt multiplies and the siege is capped.
func TestBirth(t *testing.T) {
	tn := Default()
	cradle := BirthChance(Biological, Factors{Worlds: 1}, &tn)
	paid := BirthChance(Biological, Factors{Worlds: 1, Rungs: 5}, &tn)
	if ratio := cradle / paid; ratio < 200 || ratio > 300 {
		t.Errorf("cradle to ladder-paid births %.0f to one, want about 250", ratio)
	}
	if BirthChance(Biological, Factors{Worlds: 1, Immune: true}, &tn) != 0 {
		t.Error("a people at the top of the ladder bore a plague")
	}
	dirty := BirthChance(Biological, Factors{Worlds: 1, Sieged: 3, Shed: true, Dark: true, Taken: true}, &tn)
	if want := cradle * 4 * 2 * 3 * 2; dirty < want*0.999 || dirty > want*1.001 {
		t.Errorf("a besieged, shedding, fallen, conquered people bears at %.5f, want %.5f", dirty, want)
	}
	if once := 1000 / cradle; once < 400_000 || once > 500_000 {
		t.Errorf("a cradle bears once in %.0f years, want about four hundred and fifty thousand", once)
	}
	// an idea needs a medium and is a quarter as common; the wars of faith double it
	meme := BirthChance(Memetic, Factors{Worlds: 1, Faith: true, Taken: true}, &tn)
	if want := cradle / 4 * 2; meme < want*0.999 || meme > want*1.001 {
		t.Errorf("a memetic birth after the wars of faith %.5f, want %.5f", meme, want)
	}
	if BirthChance(Memetic, Factors{Worlds: 1, Halvings: 2}, &tn) != BirthChance(Memetic, Factors{Worlds: 1}, &tn) {
		t.Error("sanitation halves the birth of ideas")
	}
}

// TestCure: the contest is set so that a cradle with Survival 3 against
// a c = 0.5 plague is cured about one tick in forty and contained about
// one in four, and an era-3 people with the ladder paid is cured nineteen
// in twenty; against c = 0.9 the cradle is almost never cured and the
// era-3 people about one tick in two.
func TestCure(t *testing.T) {
	tn := Default()
	r := rand.New(rand.NewPCG(3, 4))
	rate := func(level, ladder, c float64) (cured, contained float64) {
		n := 20000
		for range n {
			switch Band(CureMargin(level, ladder, r.NormFloat64()*tn.Spread, c, 0, 0, &tn), &tn) {
			case Cured:
				cured++
			case Contained:
				contained++
			}
		}
		return cured / float64(n), contained / float64(n)
	}
	if cured, contained := rate(3, 0, 0.5); cured < 0.015 || cured > 0.04 || contained < 0.18 || contained > 0.32 {
		t.Errorf("a cradle against c 0.5: cured %.3f, contained %.3f, want about 0.025 and 0.25", cured, contained)
	}
	if cured, _ := rate(6, 2.5, 0.5); cured < 0.9 {
		t.Errorf("an era-3 people against c 0.5: cured %.3f, want about 0.95", cured)
	}
	if cured, _ := rate(3, 0, 0.9); cured > 0.002 {
		t.Errorf("a cradle against c 0.9: cured %.4f, want almost never", cured)
	}
	if cured, _ := rate(6, 2.5, 0.9); cured < 0.4 || cured > 0.6 {
		t.Errorf("an era-3 people against c 0.9: cured %.3f, want about a half", cured)
	}
	if Band(-1, &tn) != Contained || Band(-2.5, &tn) != Raging || Band(0, &tn) != Cured {
		t.Error("the bands")
	}
	if CureMargin(3, 0, 0, 0.5, 2, 2, &tn) != 3-6 {
		t.Error("dirt and a partner's cure do not cancel")
	}
}

// TestCatch: the chance a channel carries a plague is its contagion by
// the weight, divided by hygiene, capped at one; and the toll's shape.
func TestCatch(t *testing.T) {
	tn := Default()
	if got := CatchChance(0.8, 1, Hygiene(2, &tn), 0); got < 0.355 || got > 0.356 {
		t.Errorf("c 0.8 on a link with goods against two rungs: %.3f, want 0.356", got)
	}
	if CatchChance(0.8, 1, 1, 2) != 1 {
		t.Error("a doubled chance is not capped at one")
	}
	if CatchChance(0.5, 0.3, 1, 0) != 0.15 {
		t.Error("the weight")
	}
	p := Plague{Kind: Biological, Lethality: 0.5}
	if toll := TollOf(p, false, false); toll.Sur != 1 || toll.Soc != 1 || toll.Organic != 0.5 || toll.Research != 0.75 || toll.Energy != 1 {
		t.Errorf("the biological toll at l 0.5: %+v", toll)
	}
	if toll := TollOf(p, true, false); toll.Sur != 0.5 || toll.Organic != 0.75 {
		t.Errorf("the contained toll: %+v", toll)
	}
	p.Kind = Memetic
	if toll := TollOf(p, false, true); toll.Sur != 0 || toll.Soc != 1.5 || toll.Energy != 0.75 || toll.Organic != 1 {
		t.Errorf("the memetic toll on a machine: %+v", toll)
	}
}

// TestNew: a plague's shape is drawn in (0, 1], named, and one in three
// named for its host when there is one.
func TestNew(t *testing.T) {
	tn := Default()
	r := rand.New(rand.NewPCG(5, 6))
	named, conscious := 0, 0
	for range 3000 {
		p := New(r, Biological, "Qaosh", &tn)
		if p.Contagion <= 0 || p.Contagion > 1 || p.Lethality <= 0 || p.Lethality > 1 || p.Name == "" || p.Band != -1 {
			t.Fatalf("a malformed plague: %+v", p)
		}
		if p.Named {
			named++
		}
		if p.Conscious {
			conscious++
		}
	}
	if named < 900 || named > 1100 {
		t.Errorf("%d of 3000 named for the host, want about a third", named)
	}
	if conscious < 100 || conscious > 200 {
		t.Errorf("%d of 3000 conscious, want about one in twenty", conscious)
	}
	if p := New(r, Memetic, "", &tn); p.Named || p.Kind != Memetic {
		t.Errorf("a plague with no host to name it for: %+v", p)
	}
}
