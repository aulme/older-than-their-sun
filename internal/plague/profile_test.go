package plague

import (
	"math/rand/v2"
	"testing"
)

// TestProfile: every profile's keys are rows of plagues.json; a deadly
// plague's last symptom is from the deadly bank and a mild one's from
// the wasting bank; an engineered plague shows no visible mark; a
// tailored one takes one blood; a thinking idea takes a form that wants
// a host; the last effect's severity follows lethality; and the same
// seed and id give the same profile.
func TestProfile(t *testing.T) {
	r := rand.New(rand.NewPCG(9, 9))
	tn := Default()
	tab := Profiles
	for i := 0; i < 3000; i++ {
		kind := Kind(i % 2)
		var p Plague
		switch i % 3 {
		case 0:
			p = New(r, kind, &tn)
		case 1:
			p = Make(kind, r.Float64(), r.Float64(), r.IntN(2) == 0)
			if r.IntN(2) == 0 {
				p.Band = 3
			}
		default:
			p = Loose(r, kind, 0.5+r.Float64()/2, &tn)
		}
		pr := ProfileOf(11, i, p)
		if again := ProfileOf(11, i, p); again.Last() != pr.Last() || again.First() != pr.First() || again.Onset != pr.Onset || again.Form != pr.Form {
			t.Fatalf("profile %d differs on a second draw", i)
		}
		if kind == Biological {
			if len(pr.Symptoms) < 1 || len(pr.Symptoms) > 4 || pr.Form != "" {
				t.Fatalf("biological profile %d: %+v", i, pr)
			}
			for _, s := range pr.Symptoms {
				sy := tab.SymptomOf(s)
				if sy == nil {
					t.Fatalf("symptom %q is not in the table", s)
				}
				if p.Engineered && sy.Visible {
					t.Errorf("engineered plague %d shows %s", i, s)
				}
			}
			last := tab.SymptomOf(pr.Last())
			if p.Lethality > 0.6 && last.End != "deadly" {
				t.Errorf("plague %d (l %.2f) ends in %s, want a deadly symptom", i, p.Lethality, last.Key)
			}
			if p.Lethality < 0.3 && last.End != "wasting" {
				t.Errorf("plague %d (l %.2f) ends in %s, want a wasting symptom", i, p.Lethality, last.Key)
			}
			if (pr.Onset == "quick") != (p.Contagion > 0.5) {
				t.Errorf("plague %d (c %.2f) onset %s", i, p.Contagion, pr.Onset)
			}
			if p.Band >= 0 && pr.Takes != "species" {
				t.Errorf("tailored plague %d takes %s", i, pr.Takes)
			}
			ok := false
			for _, x := range tab.Courses {
				ok = ok || x.Key == pr.Course
			}
			for _, x := range tab.Takes {
				if x.Key == pr.Takes {
					ok = ok && true
				}
			}
			if !ok {
				t.Errorf("plague %d course %q", i, pr.Course)
			}
		} else {
			if len(pr.Effects) < 1 || len(pr.Effects) > 3 || len(pr.Symptoms) > 0 {
				t.Fatalf("memetic profile %d: %+v", i, pr)
			}
			f := tab.FormOf(pr.Form)
			if f == nil {
				t.Fatalf("form %q is not in the table", pr.Form)
			}
			if p.Conscious && !f.Hosted {
				t.Errorf("thinking idea %d took the form %s, which wants no host", i, pr.Form)
			}
			for _, e := range pr.Effects {
				if tab.EffectOf(e) == nil {
					t.Fatalf("effect %q is not in the table", e)
				}
			}
			if sev := tab.EffectOf(pr.Last()).Severity; sev != 1+int(p.Lethality*2.99) {
				t.Errorf("idea %d (l %.2f) ends in severity %d", i, p.Lethality, sev)
			}
			ok := false
			for _, x := range tab.Carriers {
				ok = ok || x.Key == pr.Carrier
			}
			if !ok {
				t.Errorf("idea %d carrier %q", i, pr.Carrier)
			}
		}
	}
}
