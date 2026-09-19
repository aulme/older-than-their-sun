package species

import (
	"math"
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestTilted(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	n := 1_000_000
	fails, plain := 0, 0
	for i := 0; i < n; i++ {
		if !tilted(r, 1.0/20, 1000) {
			fails++
		}
		if tilted(r, 1.0/20, 1) {
			plain++
		}
	}
	if fails == 0 {
		t.Fatal("a tilt of a thousand on one in twenty never failed")
	}
	if want := 1.0 / 20; math.Abs(float64(plain)/float64(n)-want) > 0.002 {
		t.Fatalf("a tilt of one gave %d in %d, want the base rate %.3f", plain, n, want)
	}
	if p := tiltedOdds(1.0/20, 20); p < 0.5 || p > 0.55 {
		t.Fatalf("one in twenty tilted by twenty should come out a little over even, got %.3f", p)
	}
	if tiltedOdds(0.5, 0) != 0 || tiltedOdds(1, 0.001) != 1 {
		t.Fatal("a zero tilt is impossible and a base of one is certain")
	}
}

// oldKinds is the old kind table's shares, which the legacy numbers must
// reproduce, less the parasite, which no cradle rolls since the plagues
// draft: a rider is born of a plague that woke.
func TestLegacyDistribution(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	n := 100_000
	got := map[string]int{}
	for i := 0; i < n; i++ {
		s := Roll(r, 1, Options{Legacy: true})
		switch {
		case s.Sub == Parasite:
			got["parasite"]++
		case s.Has("swarming"):
			got["swarm"]++
		case s.Is(Planetary):
			got["planetary mind"]++
		case s.Is(Evolver):
			got["evolver"]++
		case s.Sub == Biological && s.Mods&^(Hive|Unconscious) == 0:
			got["standard"]++
		default:
			t.Fatalf("rolled a kind the old table did not have: %s", s.Nature())
		}
	}
	old := map[string]float64{"standard": 74, "swarm": 7, "planetary mind": 4, "parasite": 0, "evolver": 8}
	for k, w := range old {
		want := w / 93
		if have := float64(got[k]) / float64(n); math.Abs(have-want) > 0.01 {
			t.Errorf("%s: %.3f, want %.3f", k, have, want)
		}
	}
	// the hive and the unconscious were org traits at 10 and 5 of 102
	hives, un := 0, 0
	for i := 0; i < n; i++ {
		s := Roll(r, 1, Options{Legacy: true})
		if s.Is(Hive) {
			hives++
		}
		if s.Is(Unconscious) {
			un++
		}
	}
	if have := float64(hives) / float64(n); math.Abs(have-10.0/102) > 0.01 {
		t.Errorf("hives: %.3f, want %.3f", have, 10.0/102)
	}
	if have := float64(un) / float64(n); math.Abs(have-5.0/102) > 0.01 {
		t.Errorf("unconscious: %.3f, want %.3f", have, 5.0/102)
	}
}

// setOdds is the chance, from the table, that a cradle roll under the
// proposed numbers carries exactly the given modifiers.
func setOdds(mods Mod) float64 {
	total := 0.0
	for _, d := range Substrates {
		total += d.Draws.Base
	}
	p := 0.0
	for _, d := range Substrates {
		for _, swarming := range []bool{false, true} {
			carried := []*Entry{&d.Entry}
			tilt := func(key string) float64 {
				t := 1.0
				for _, e := range carried {
					if v, ok := e.Draws.Tilts[key]; ok {
						t *= v
					}
				}
				return t
			}
			ps := tiltedOdds(swarm.Draws.Base, tilt("swarming"))
			if swarming {
				carried = append(carried, swarm)
			} else {
				ps = 1 - ps
			}
			q := d.Draws.Base / total * ps
			for _, m := range Mods {
				pm := tiltedOdds(m.Draws.Base, tilt(m.Key))
				if mods.Has(m.Mod) {
					q *= pm
					carried = append(carried, &m.Entry)
				} else {
					q *= 1 - pm
				}
			}
			p += q
		}
	}
	return p
}

func TestProposedDistribution(t *testing.T) {
	r := rand.New(rand.NewPCG(5, 6))
	n := 100_000
	none := 0
	for i := 0; i < n; i++ {
		if Generate(r, 1).Mods == 0 {
			none++
		}
	}
	if none <= n/2 {
		t.Fatalf("zero modifiers in %d of %d, want more than half", none, n)
	}
	if want := setOdds(0); math.Abs(float64(none)/float64(n)-want) > 0.01 {
		t.Errorf("zero modifiers: %.3f, the table says %.3f", float64(none)/float64(n), want)
	}
	all := Mod(0)
	for _, d := range Mods {
		all |= d.Mod
	}
	sum := 0.0
	for set := Mod(0); set <= all; set++ {
		p := setOdds(set)
		sum += p
		if p <= 0 {
			t.Errorf("modifier set %q has no chance", set.String())
		}
	}
	if math.Abs(sum-1) > 1e-9 {
		t.Errorf("the sets' odds sum to %.6f", sum)
	}
	if p := setOdds(all); p >= 1e-7 {
		t.Errorf("the full set has odds %.3g, want below one in ten million", p)
	}
}

// TestCradleDistribution: the proposal's per-thousand figures at a
// cradle, within a fifth, over a hundred thousand draws.
func TestCradleDistribution(t *testing.T) {
	r := rand.New(rand.NewPCG(9, 10))
	n := 100_000
	got := map[string]int{}
	powers := 0
	for i := 0; i < n; i++ {
		s := Generate(r, 1)
		switch {
		case s.Sub == Biological && s.Mods == 0:
			got["plain biological"]++
		case s.Sub == Biological && s.Mods == Hive:
			got["biological hive"]++
		case s.Sub == Biological && s.Mods == Evolver:
			got["biological evolver"]++
		case s.Sub == Biological && s.Mods == Unconscious:
			got["unconscious biological"]++
		case s.Sub == Biological && s.Mods == Planetary:
			got["living world"]++
		case s.Sub == Biological && s.Mods == Replicator:
			got["biological replicator"]++
		case s.Sub == Eldritch && s.Mods == 0:
			got["plain eldritch"]++
		case s.Sub == Eldritch && s.Mods == Unconscious:
			got["unconscious eldritch"]++
		case s.Sub == Eldritch && s.Mods == Planetary:
			got["eldritch world"]++
		case s.Sub == Eldritch && s.Mods == Planetary|Unconscious:
			got["sleeper"]++
		case s.Sub == Machine && s.Mods == 0:
			got["machine"]++
		case s.Sub == Biological && s.Mods == Antimemetic:
			got["antimemetic"]++
		}
		if s.Sub == Eldritch {
			if len(s.Powers) < 1 || len(s.Powers) > 3 {
				t.Fatalf("an eldritch people drew %d powers", len(s.Powers))
			}
			powers++
			sense := false
			for _, tr := range s.Traits {
				sense = sense || tr.Group == "sense"
			}
			if !sense {
				t.Fatalf("an eldritch people without a sense: %s", s.Describe())
			}
		} else if len(s.Powers) > 0 {
			t.Fatalf("a %s people with powers", s.Sub)
		}
	}
	perThousand := map[string]float64{
		"plain biological": 720, "biological hive": 70, "biological evolver": 60, "unconscious biological": 33, "living world": 24,
		"plain eldritch": 20, "unconscious eldritch": 9, "eldritch world": 6.5, "sleeper": 6.5, "machine": 15, "biological replicator": 6, "antimemetic": 3.5,
	}
	for k, want := range perThousand {
		have := float64(got[k]) / float64(n) * 1000
		if math.Abs(have-want) > want/5 {
			t.Errorf("%s: %.1f per thousand, the proposal says %.1f", k, have, want)
		}
	}
	if powers == 0 {
		t.Fatal("no eldritch people drew powers")
	}
}

func TestProfileAndFixing(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 8))
	plain := GenerateWith(r, 1, "lush", Machine, 0)
	plain.Mods = 0
	if a, b := plain.Profile(), Compose(machine.Profile); !reflect.DeepEqual(a, b) {
		t.Fatalf("a people with no modifiers should have its substrate's profile: %+v vs %+v", a, b)
	}
	for i := 0; i < 200; i++ {
		s := GenerateWith(r, 1, "arid", Machine, Hive|Evolver)
		if s.Sub != Machine || !s.Is(Hive|Evolver) || s.World.Key != "arid" {
			t.Fatalf("GenerateWith did not fix what it was told: %s on %s", s.Nature(), s.World.Key)
		}
		if orgOf(s) != "" || !s.Has("onequeen") && !s.Has("noqueen") && !s.Has("throne") {
			t.Fatalf("a hive should skip the org group and roll a seat: %v", s.Describe())
		}
		if s.Has("throne") != s.Has("nomadic") {
			t.Fatalf("a moving throne is a nomad hive's and nobody else's: %v", s.Describe())
		}
		p := GenerateWith(r, 1, "", Parasite, 0)
		if !p.Has("bodyrider") && !p.Has("mindrider") {
			t.Fatalf("a parasite should roll a rider: %v", p.Describe())
		}
		_ = p
	}
	if !plain.Profile().Can(CivilWars) || plain.Profile().Can(Sickens) {
		t.Fatal("a machine people splits and does not sicken")
	}
	h := GenerateWith(r, 1, "", Biological, Hive)
	h.Mods, h.prof = Hive, nil // whatever else the roll gave it
	h.Traits = nil
	if h.Profile().Can(CivilWars) || h.Profile().Memory != 0.7 || h.Profile().FilterDiff["distance"] != -3 {
		t.Fatalf("the hive's profile did not carry the old rows: %+v", h.Profile())
	}
	if f := h.Flavour(); f.Colony != "colony" {
		t.Fatalf("a hive builds colonies, got %q", f.Colony)
	}
	h.Add("swarming")
	if f := h.Flavour(); f.Colony != "nest" || f.Ship != "seed-cloud" {
		t.Fatalf("a swarm builds nests and seed-clouds, got %+v", f)
	}
	e := GenerateWith(r, 1, "", Eldritch, 0)
	e.Powers = nil
	e.prof = nil
	if p := e.Profile(); p.Can(Works) || p.Can(Researches) || p.Era != 3 || p.Alien != 3 {
		t.Fatalf("the eldritch profile: %+v", p)
	}
	if !e.AddPower("making") || e.AddPower("making") {
		t.Fatal("adding a power twice")
	}
	if !e.Profile().Can(Works) {
		t.Fatal("the making did not give the works back")
	}
	if !e.AddPower("sight") || e.Profile().Era != 4 {
		t.Fatalf("an exotic power did not make it exotic: era %d", e.Profile().Era)
	}
	e.StripPower("making")
	if e.Profile().Can(Works) || e.Profile().Era != 4 {
		t.Fatal("stripping a power")
	}
	if e.PowerPortrait() != "It has the sight." {
		t.Fatalf("the portrait: %q", e.PowerPortrait())
	}
	u := GenerateWith(r, 1, "", Biological, Unconscious)
	if p := u.Profile(); p.Can(HoldsGrudges) || p.Can(Wavers) || !p.Amoral || !p.NoOne {
		t.Fatalf("the unconscious profile: %+v", p)
	}
	pl := GenerateWith(r, 1, "", Eldritch, Planetary)
	if p := pl.Profile(); p.HomeDefence != 6 || p.Neighbourhood != 12 || p.Worlds != 4 || p.Can(Launches) {
		t.Fatalf("an eldritch living world's profile: %+v", p)
	}
}

func orgOf(s *Species) string {
	for _, t := range s.Traits {
		if t.Group == "org" {
			return t.Key
		}
	}
	return ""
}
