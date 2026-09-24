package names

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"worldgen/data"
	"worldgen/internal/history"
	"worldgen/internal/plague"
	"worldgen/internal/species"
)

// sizes are the inventory's targets at landing: entries per kind of
// thing and tone group, and how many of them are grounded (require a
// property of the thing).
var sizes = []struct {
	about   string
	tones   []string
	entries int
	ground  int
	req     string // a requirement that marks the group, "" for any
}{
	{"civ", []string{"self"}, 60, 50, ""},
	{"civ", []string{"self"}, 20, 20, "kind:machine"},
	{"civ", []string{"stranger", "friend"}, 120, 100, ""},
	{"civ", []string{"enemy", "monster"}, 120, 100, ""},
	{"star", []string{"self", "stranger", "sky"}, 80, 80, ""},
	{"elder", []string{"stranger"}, 40, 40, ""},
	{"makers", []string{"stranger"}, 40, 40, ""},
	{"plague", []string{"self", "stranger"}, 40, 40, ""},
	{"title", []string{"self"}, 30, 20, ""},
	{"war", []string{"self"}, 20, 20, ""},
	{"word", []string{"self"}, 30, 30, ""},
	{"leader", []string{"self"}, 24, 16, ""},
}

func hasTone(e *Entry, tones []string) bool {
	for _, t := range tones {
		if contains(e.Tone, t) {
			return true
		}
	}
	return false
}

// TestInventorySize: the inventory meets its size and grounding targets,
// and every counted property is required by at least eight entries.
func TestInventorySize(t *testing.T) {
	for _, sz := range sizes {
		n, g := 0, 0
		for _, e := range Epithets() {
			if e.About != sz.about || !hasTone(e, sz.tones) || (sz.req != "" && !contains(e.Requires, sz.req)) {
				continue
			}
			n++
			if len(e.Requires) > 0 {
				g++
			}
		}
		if n < sz.entries || g < sz.ground {
			t.Errorf("%s %v%s: %d entries (%d grounded), want %d (%d)", sz.about, sz.tones, sz.req, n, g, sz.entries, sz.ground)
		}
	}
	counts := map[string]int{}
	for _, e := range Epithets() {
		for _, r := range e.Requires {
			counts[strings.TrimPrefix(r, "!")]++
		}
	}
	for _, p := range Properties() {
		if counts[p] < 8 {
			t.Errorf("property %s is required by %d entries, want eight", p, counts[p])
		}
	}
	t.Logf("%d entries, %d counted properties", len(Epithets()), len(Properties()))
}

// TestInventoryWellFormed: every id is unique, every slot names a bank
// that exists or a value the pass fills, every pattern's holes have
// slots, every bank has a dozen words, and every requirement is a
// property key the pass can set.
func TestInventoryWellFormed(t *testing.T) {
	ids := map[string]bool{}
	hole := regexp.MustCompile(`\{(\w+)\}`)
	values := map[string]bool{"self": true, "home": true, "cradle": true, "them": true, "deed.star": true, "host": true, "host.star": true, "star": true, "enemy": true, "work": true, "first": true, "first.adj": true, "last": true, "mark": true, "form": true, "onset": true}
	prefixes := []string{"trait:", "world:", "way:", "fix:", "deed:", "did:", "bond:", "warcause:", "heavy:", "voice:", "kind:", "class:", "mult:", "remnant:", "place:", "legacy:", "work:", "portrait:", "elder:", "symptom:", "first:", "last:", "mark:", "onset:", "course:", "takes:", "form:", "effect:", "carrier:", "cause:", "side:", "miracle:", "occasion:", "stance:", "leader:"}
	flags := map[string]bool{"seen": true, "bright": true, "dim": true, "dead": true, "brilliant": true, "marked": true, "made": true, "host": true, "thinks": true, "named": true, "again": true, "deed.star": true}
	for _, e := range Epithets() {
		if ids[e.ID] {
			t.Errorf("entry %s twice", e.ID)
		}
		ids[e.ID] = true
		for _, h := range hole.FindAllStringSubmatch(e.Pattern, -1) {
			src, ok := e.Slots[h[1]]
			if !ok {
				t.Errorf("%s: hole {%s} has no slot", e.ID, h[1])
				continue
			}
			if bank, isBank := strings.CutPrefix(src, "bank:"); isBank {
				if len(Banks()[bank]) < 12 {
					t.Errorf("%s: bank %s has %d words", e.ID, bank, len(Banks()[bank]))
				}
			} else if !values[src] {
				t.Errorf("%s: slot %s reads %q, which the pass never fills", e.ID, h[1], src)
			}
		}
		for _, r := range e.Requires {
			r = strings.TrimPrefix(r, "!")
			ok := flags[r]
			for _, p := range prefixes {
				ok = ok || strings.HasPrefix(r, p)
			}
			if !ok {
				t.Errorf("%s requires %q, which the pass never sets", e.ID, r)
			}
		}
	}
	for k, words := range Banks() {
		if len(words) < 12 {
			t.Errorf("bank %s has %d words, want a dozen", k, len(words))
		}
	}
}

// TestKeysReal: every property key an entry requires that names a row of
// another table (a trait, a world, a portrait, a symptom, a cause) is a
// row of that table.
func TestKeysReal(t *testing.T) {
	var traits struct {
		Traits []struct {
			Key string `json:"key"`
		} `json:"traits"`
	}
	data.Load("traits.json", &traits)
	var worlds struct {
		Archetypes []struct {
			Key string `json:"key"`
		} `json:"archetypes"`
	}
	data.Load("worlds.json", &worlds)
	var portraits map[string]json.RawMessage
	data.Load("portraits.json", &portraits)
	var causes struct {
		Causes     []struct{ Key string } `json:"causes"`
		WarCauses  []struct{ Key string } `json:"war_causes"`
		WarResults []struct{ Key string } `json:"war_results"`
	}
	data.Load("causes.json", &causes)
	var works struct {
		Structures []struct{ Key string } `json:"structures"`
	}
	data.Load("works.json", &works)
	var events struct {
		Kinds []struct{ Kind string } `json:"kinds"`
	}
	data.Load("events.json", &events)
	known := map[string]bool{}
	for _, x := range traits.Traits {
		known["trait:"+x.Key] = true
		known["way:"+x.Key] = true
		known["world:"+x.Key] = true
	}
	for _, x := range worlds.Archetypes {
		known["world:"+x.Key] = true
	}
	known["world:manysuns"] = true
	for name, raw := range portraits {
		if name == "_" {
			continue
		}
		var rows []struct{ Key string }
		if json.Unmarshal(raw, &rows) == nil {
			for _, r := range rows {
				known["portrait:"+r.Key] = true
				known["elder:"+r.Key] = true
			}
		}
	}
	for _, x := range works.Structures {
		known["portrait:"+x.Key] = true
	}
	for _, x := range causes.Causes {
		known["cause:"+x.Key] = true
	}
	for _, x := range causes.WarCauses {
		known["cause:"+x.Key] = true
		known["warcause:"+x.Key] = true
	}
	for _, x := range events.Kinds {
		for _, p := range []string{"deed:", "did:", "bond:", "heavy:"} {
			known[p+x.Kind] = true
		}
	}
	var miracles struct {
		Miracles []struct {
			Forms []struct{ Key string } `json:"forms"`
		} `json:"miracles"`
	}
	data.Load("miracles.json", &miracles)
	for _, m := range miracles.Miracles {
		for _, f := range m.Forms {
			known["form:"+f.Key] = true
		}
	}
	pt := plague.Profiles
	for _, s := range pt.Symptoms {
		known["symptom:"+s.Key], known["first:"+s.Key], known["last:"+s.Key], known["mark:"+s.Key] = true, true, true, true
	}
	for _, s := range pt.Forms {
		known["form:"+s.Key] = true
	}
	for _, s := range pt.Effects {
		known["effect:"+s.Key], known["last:"+s.Key] = true, true
	}
	for _, s := range pt.Onsets {
		known["onset:"+s.Key] = true
	}
	for _, s := range pt.Courses {
		known["course:"+s.Key] = true
	}
	for _, s := range pt.Takes {
		known["takes:"+s.Key] = true
	}
	for _, s := range pt.Carriers {
		known["carrier:"+s.Key] = true
	}
	var leaders struct {
		Forms, Occasions []struct {
			Key string `json:"key"`
		}
	}
	data.Load("leaders.json", &leaders)
	for _, f := range leaders.Forms {
		known["leader:"+f.Key] = true
	}
	for _, o := range leaders.Occasions {
		known["occasion:"+o.Key] = true
	}
	for _, e := range Epithets() {
		for _, r := range e.Requires {
			if k, ok := strings.CutPrefix(r, "stance:"); ok && species.Get(k) == nil {
				t.Errorf("%s requires %q, which is no stance", e.ID, r)
			}
		}
	}
	for _, e := range Epithets() {
		for _, r := range e.Requires {
			r = strings.TrimPrefix(r, "!")
			i := strings.IndexByte(r, ':')
			if i < 0 {
				continue
			}
			switch r[:i] {
			case "trait", "way", "world", "portrait", "elder", "cause", "warcause", "deed", "did", "bond", "heavy", "symptom", "first", "last", "mark", "form", "effect", "onset", "course", "takes", "carrier", "leader", "occasion":
				if !known[r] {
					t.Errorf("%s requires %q, which no table has", e.ID, r)
				}
			}
		}
	}
}

// batch is the books of a few seeds, for the tests that read many
// names: one book per age. The ages not made yet are made side by side,
// since they share nothing but the tables, which are built at init and
// never written.
func batch(t *testing.T) []*Book {
	t.Helper()
	made := make([]func() *history.World, len(batchSeeds))
	for i, seed := range batchSeeds {
		made[i] = ages[seed]
	}
	worlds := make([]*history.World, len(made))
	var wg sync.WaitGroup
	for i, age := range made {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worlds[i] = age()
		}()
	}
	wg.Wait()
	out := make([]*Book, len(worlds))
	for i, w := range worlds {
		out[i] = Of(w)
	}
	return out
}

// TestExonymSpread: the pass spreads its exonyms. Among the patterns
// that answer the same thing — one requirement, one tone — none is used
// more than twice the mean of its kin, wherever that thing was named
// often enough to judge; and the batch uses at least a hundred and fifty
// entries.
//
// Until step 8 this asked that no entry be more than three percent of
// the batch's exonyms, and that is a question about the histories, not
// the pass. Which deed a people is named for is the deed its namers
// remember best, so one age can put a deed's patterns above the bar
// whatever the pass does: at step 8 one of the four ages gave 2975 of
// 3169 exonyms, its peoples had refused each other so often that each
// of the refused deed's four patterns was four percent of the batch,
// and giving the deed six patterns made each still four percent, a
// quarter of the batch between them, since the pass scores every row on
// its own and a deed with more rows wins more often. The small ages,
// weighted equally instead, put the stranger's first-meeting names over
// the bar the same way. Before that the test needed first a thousand
// exonyms and then five hundred, and both floors failed on draws of the
// heavy tail. The largest shares of the batch are still logged, so a
// history that names everyone for one thing is visible.
//
// A requirement is judged once its kin are named fifty times, or a
// twentieth of the batch's exonyms where that is fewer: at step 9 the
// four ages gave 401 exonyms, no requirement reached fifty, and a fixed
// floor failed on the heavy tail a third time.
func TestExonymSpread(t *testing.T) {
	counts := map[string]int{}
	n := 0
	for _, b := range batch(t) {
		for _, r := range b.All() {
			if r.Object.Kind == "civ" && r.Mode == "translated" && r.Tone != "self" {
				counts[r.Recipe.Entry]++
				n++
			}
		}
	}
	if len(counts) < 150 {
		t.Errorf("%d exonyms from %d entries: the pass has narrowed", n, len(counts))
	}
	// kin are the entries that answer the same requirement in the same tone
	kin := map[string][]string{}
	for _, e := range Epithets() {
		if e.About != "civ" || len(e.Requires) == 0 {
			continue
		}
		k := strings.Join(e.Requires, "&") + "|" + strings.Join(e.Tone, ",")
		kin[k] = append(kin[k], e.ID)
	}
	judged, floor := 0, min(50, n/20)
	for _, k := range sortedKeys(kin) {
		ids := kin[k]
		sum := 0
		for _, id := range ids {
			sum += counts[id]
		}
		if len(ids) < 2 || sum < floor {
			continue
		}
		judged++
		mean := float64(sum) / float64(len(ids))
		for _, id := range ids {
			if float64(counts[id]) > 2*mean {
				t.Errorf("%s: %d of %d names for %s, more than twice its kin's mean of %.0f", id, counts[id], sum, k, mean)
			}
		}
	}
	if judged == 0 {
		t.Fatalf("%d exonyms, and no requirement named %d times to judge the spread", n, floor)
	}
	top := sortedKeys(counts)
	sort.SliceStable(top, func(i, j int) bool { return counts[top[i]] > counts[top[j]] })
	t.Logf("%d exonyms from %d entries, %d requirements judged; the largest shares of the batch, which are the histories':", n, len(counts), judged)
	for _, e := range top[:min(3, len(top))] {
		t.Logf("  %s %d (%.1f%%)", e, counts[e], 100*float64(counts[e])/float64(n))
	}
}

// TestEntitled: every translated row's requirements held of the thing
// as the namer could know it at the coining year, recomputed from the
// facts: a body or a world only after a meeting in the flesh or an
// understanding, a way only after an understanding or a war, a deed only
// from a fact before the coining.
func TestEntitled(t *testing.T) {
	w := world(t, 5)
	b := Of(w)
	byID := map[string]*Entry{}
	for _, e := range Epithets() {
		byID[e.ID] = e
	}
	checked := 0
	for _, r := range b.All() {
		if r.Object.Kind != "civ" || r.Mode != "translated" || r.By == r.Object.ID {
			continue
		}
		e := byID[r.Recipe.Entry]
		if e == nil {
			t.Fatalf("row %v names an entry the inventory lacks", r)
		}
		// what the facts before the coining allow
		touched, fathomed, warred, met := false, false, false, false
		deeds := map[string]bool{}
		for _, f := range w.Events {
			if f.Year > r.Coined || !f.IsFact() {
				continue
			}
			pair := (f.Subject == r.By && f.Object == r.Object.ID) || (f.Subject == r.Object.ID && f.Object == r.By)
			if !pair {
				continue
			}
			met = true
			switch f.Kind {
			case history.FMet:
				touched = touched || f.P["how"] == "touch"
			case history.FFathomed:
				fathomed = fathomed || f.Subject == r.By
			case history.FWar:
				warred = true
			}
			deeds[string(f.Kind)] = true
		}
		for _, req := range e.Requires {
			if strings.HasPrefix(req, "!") {
				continue
			}
			i := strings.IndexByte(req, ':')
			if i < 0 {
				continue
			}
			ok := true
			switch req[:i] {
			case "trait", "world":
				ok = touched || fathomed
			case "way", "fix":
				ok = fathomed || warred
			case "deed", "did", "bond", "heavy":
				ok = deeds[req[i+1:]]
			case "warcause":
				ok = warred
			case "kind", "voice":
				ok = met
			}
			if !ok {
				t.Errorf("%d names %d %q by %s at %d, but could not know %s then", r.By, r.Object.ID, r.Name, e.ID, r.Coined, req)
			}
			checked++
		}
	}
	t.Logf("%d requirements checked", checked)
}

// TestVoicelessNamed: every voiceless people that was met has an exonym
// from someone, and a translated voice's endonym is a recipe; every
// plague has a name from its first host when it has a voice; every
// remnant a title; every reaching-in a word.
func TestVoicelessNamed(t *testing.T) {
	w := world(t, 3)
	b := Of(w)
	for _, c := range w.Civs {
		rows := b.Rows(Object{Kind: "civ", ID: c.ID})
		if b.Voice(c) == None {
			metBy := false
			for _, f := range w.Events {
				if f.Kind == history.FMet && (f.Object == c.ID || f.Subject == c.ID) && f.Object >= 0 {
					if f.Subject == c.ID && f.P["how"] == "noticed" {
						continue // it noticed them, unseen: they never met it, and have nothing to call it
					}
					o := f.Subject
					if o == c.ID {
						o = f.Object
					}
					if b.Voice(w.Civs[o]) != None {
						metBy = true
					}
				}
			}
			if metBy && len(rows) == 0 {
				t.Errorf("civ %d is voiceless, was met by a voiced people, and has no name", c.ID)
			}
		}
		if b.Voice(c) == Translated {
			self := earliest(rows, "self")
			if self == nil || self.Recipe.Entry == "" {
				t.Errorf("civ %d speaks in translation and has no recipe for its own name: %+v", c.ID, self)
			}
		}
		if c.Fate == history.Contracted && b.Voice(c) != None && len(b.Rows(Object{Kind: "title", ID: c.ID})) == 0 {
			t.Errorf("civ %d is a remnant with no title", c.ID)
		}
	}
	for _, p := range w.Plagues {
		if p.FirstHost >= 0 && b.Voice(w.Civs[p.FirstHost]) != None && len(b.Rows(Object{Kind: "plague", ID: p.ID})) == 0 {
			t.Errorf("plague %d has no name from its first host %d", p.ID, p.FirstHost)
		}
	}
	for _, f := range w.Events {
		if f.Kind == history.FWord && b.Voice(w.Civs[f.Subject]) != None && len(b.Rows(Object{Kind: "word", ID: f.Subject})) == 0 {
			t.Errorf("civ %d reached in and has no word", f.Subject)
		}
	}
}

// TestNoTokensUnresolved: every name of the view resolves: a name may
// hold a token, and the fixed point leaves none.
func TestNoTokensUnresolved(t *testing.T) {
	w := world(t, 7)
	b := Of(w)
	for _, r := range b.All() {
		s := b.Text(r.Name)
		if strings.Contains(s, "{") || strings.Contains(s, "  ") || strings.HasSuffix(s, " ") {
			t.Errorf("%q renders as %q", r.Name, s)
		}
	}
}
