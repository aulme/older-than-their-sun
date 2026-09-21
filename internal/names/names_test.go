package names

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"worldgen/internal/history"
)

// TestSource: the history holds ids, never names. It imports nothing
// from this package, and none of its nameable types has a Name field;
// the plague package likewise.
func TestSource(t *testing.T) {
	nameable := map[string]bool{"Civ": true, "Legacy": true, "Elder": true, "War": true, "Plague": true, "World": true, "Species": true}
	for _, dir := range []string{"../history", "../plague", "../species"} {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, parser.ImportsOnly|parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		for _, pkg := range pkgs {
			for name, f := range pkg.Files {
				for _, imp := range f.Imports {
					if strings.Contains(imp.Path.Value, "internal/names") {
						t.Errorf("%s imports the names pass; the simulation must not", name)
					}
				}
			}
		}
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, file, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				ts, ok := n.(*ast.TypeSpec)
				if !ok || !nameable[ts.Name.Name] {
					return true
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					return true
				}
				for _, fld := range st.Fields.List {
					for _, id := range fld.Names {
						switch id.Name {
						case "Name", "HomeName", "CradleName", "Title", "Word":
							t.Errorf("%s: %s.%s is a name field; names are rows of the pass", file, ts.Name.Name, id.Name)
						}
					}
				}
				return true
			})
		}
	}
}

func world(t *testing.T, seed uint64) *history.World {
	t.Helper()
	cfg := history.DefaultConfig()
	cfg.Stars = 120 // small: the pass is milliseconds, the run is the cost
	return history.Generate(seed, cfg)
}

// TestDeterministic: the same record gives the same rows, and adding a
// relationship changes no row that was there.
func TestDeterministic(t *testing.T) {
	w := world(t, 7)
	a := Of(w)
	b := Of(w)
	if len(a.All()) != len(b.All()) {
		t.Fatalf("%d rows, then %d", len(a.All()), len(b.All()))
	}
	for i, r := range a.All() {
		if r != b.All()[i] {
			t.Fatalf("row %d differs: %+v / %+v", i, r, b.All()[i])
		}
	}
	// two peoples that never met, meeting now
	var x, y *history.Civ
	for _, c := range w.Civs {
		if a.Voice(c) != Transcribed {
			continue
		}
		for _, d := range w.Civs {
			if d != c && a.Voice(d) == Transcribed && !c.Met[d.ID] && !d.Met[c.ID] {
				x, y = c, d
				break
			}
		}
		if x != nil {
			break
		}
	}
	if x == nil {
		t.Skip("every pair has met")
	}
	w.Facts = append(w.Facts, &history.Fact{ID: len(w.Facts), Kind: history.FMet, Year: w.Present, Subject: x.ID, Object: y.ID, Star: -1, Legacy: -1, Plague: -1, What: "touch"})
	c := Of(w)
	before := map[Row]bool{}
	for _, r := range c.All() {
		before[r] = true
	}
	for _, r := range a.All() {
		if !before[r] {
			t.Errorf("a row changed when a meeting was added: %+v", r)
		}
	}
	if n := len(c.All()) - len(a.All()); n != 2 {
		t.Errorf("a meeting added %d rows, want two adopted names", n)
	}
}

// TestVoice: a voiceless people has no name of its own; a voiced one
// has exactly one; every name meets the convention.
func TestVoice(t *testing.T) {
	w := world(t, 3)
	b := Of(w)
	for _, c := range w.Civs {
		self := 0
		for _, r := range b.Rows(Object{"civ", c.ID}) {
			if r.Tone == "self" {
				self++
			}
			if r.Mode == "transcribed" || r.Mode == "adopted" {
				if f := b.phonologyOf(r.By); f == nil || !Conventional(r.Name, f.family().Exempt) {
					t.Errorf("%q by %d is not a conventional transcription", r.Name, r.By)
				}
			}
		}
		switch v := b.Voice(c); {
		case v == None && self != 0:
			t.Errorf("civ %d (%s) is voiceless and names itself", c.ID, c.Species.Describe())
		case v != None && self != 1:
			t.Errorf("civ %d has %d names for itself", c.ID, self)
		}
	}
}

// Phonology by civ id, for the tests.
func (b *Book) phonologyOf(id int) *Phonology { return b.phon[id] }

// TestFamilies: a thousand names per family meet the convention, and a
// letter-class classifier tells the families apart well above chance.
func TestFamilies(t *testing.T) {
	type sample struct {
		family int
		feats  []float64
	}
	var samples []sample
	for fi, f := range families {
		r := rand.New(rand.NewPCG(uint64(fi), 99))
		for i := 0; i < 1000; i++ {
			p := NewPhonology(r, append([]string(nil), f.Motivated...))
			if p.Family != f.Key {
				// the roll is by weight; force the family under test
				p = &Phonology{Family: f.Key, Onsets: subset(r, f.Onsets), Nuclei: subset(r, f.Nuclei), Codas: subset(r, f.Codas), Syllables: f.Syllables}
			}
			n := p.Word(r, 0)
			if !Conventional(n, f.Exempt) {
				t.Fatalf("%s: %q breaks the convention", f.Key, n)
			}
			samples = append(samples, sample{fi, Features(n)})
		}
	}
	// nearest centroid on z-scored features
	k := len(samples[0].feats)
	mean, sd := make([]float64, k), make([]float64, k)
	for _, s := range samples {
		for j, v := range s.feats {
			mean[j] += v
		}
	}
	for j := range mean {
		mean[j] /= float64(len(samples))
	}
	for _, s := range samples {
		for j, v := range s.feats {
			sd[j] += (v - mean[j]) * (v - mean[j])
		}
	}
	for j := range sd {
		sd[j] = math.Sqrt(sd[j]/float64(len(samples))) + 1e-9
	}
	z := func(f []float64) []float64 {
		out := make([]float64, k)
		for j, v := range f {
			out[j] = (v - mean[j]) / sd[j]
		}
		return out
	}
	cent := make([][]float64, len(families))
	cnt := make([]int, len(families))
	for i := range cent {
		cent[i] = make([]float64, k)
	}
	for _, s := range samples {
		for j, v := range z(s.feats) {
			cent[s.family][j] += v
		}
		cnt[s.family]++
	}
	for i := range cent {
		for j := range cent[i] {
			cent[i][j] /= float64(cnt[i])
		}
	}
	right := make([]int, len(families))
	confused := map[[2]int]int{}
	for _, s := range samples {
		zf := z(s.feats)
		best, bd := -1, math.Inf(1)
		for i, c := range cent {
			d := 0.0
			for j := range c {
				d += (zf[j] - c[j]) * (zf[j] - c[j])
			}
			if d < bd {
				best, bd = i, d
			}
		}
		if best == s.family {
			right[s.family]++
		} else {
			confused[[2]int{s.family, best}]++
		}
	}
	total := 0
	for i, f := range families {
		acc := float64(right[i]) / 1000
		t.Logf("%-16s %.0f%%", f.Key, acc*100)
		total += right[i]
		if acc < 0.4 {
			t.Errorf("%s is told from the others only %.0f%% of the time", f.Key, acc*100)
		}
	}
	if acc := float64(total) / float64(len(samples)); acc < 0.6 {
		t.Errorf("the families are told apart %.0f%% of the time, want well above the %.0f%% of chance", acc*100, 100/float64(len(families)))
	}
	for pair, n := range confused {
		if n > 350 {
			t.Errorf("%s is taken for %s %d times in a thousand", families[pair[0]].Key, families[pair[1]].Key, n)
		}
	}
}

// TestAdopt: a name in another mouth is conventional and not empty, and
// a translated voice copies it unchanged.
func TestAdopt(t *testing.T) {
	r := rand.New(rand.NewPCG(4, 4))
	for _, f := range families {
		p := &Phonology{Family: f.Key, Onsets: f.Onsets, Nuclei: f.Nuclei, Codas: f.Codas, Syllables: f.Syllables}
		for _, src := range []string{"Kailoa", "Krv-thak", "Tsu'o", "Ulan-teru-mek", "Ssesh", "Ngoma", "Tiki-tiki", "Aeioa", "Lirala", "Xkrrth'qv"} {
			got := p.Adopt(r, src)
			if got == "" || !Conventional(got, f.Exempt) {
				t.Errorf("%s adopts %q as %q", f.Key, src, got)
			}
		}
	}
}
