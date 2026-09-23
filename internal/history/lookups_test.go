package history

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"worldgen/internal/plague"

	"worldgen/internal/galaxy"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// The lookups: every key the record can carry resolves in data/, and the
// simulation neither stores nor compares the text beside it.

// keyParams says, for each event parameter that carries a key, which
// table resolves it. A parameter not here that carries a string must
// look like a key (no space, no token), so text cannot come back under
// a new name; enums says the ones whose values are the code's own.
var keyParams = map[string]func(w *World, e *Event, v string) bool{
	"cause": func(w *World, e *Event, v string) bool {
		if e.Kind == FWar || e.Kind == KWarOpened {
			return tables.warCauses[v] != nil
		}
		return tables.causes[v] != nil
	},
	"why": func(w *World, e *Event, v string) bool {
		if e.Kind == KArrivalLost || e.Kind == KReason {
			return true // arrival_lost's why is an enum; reason's is debug text
		}
		return v == "" || tables.causes[v] != nil
	},
	"shape":    func(w *World, e *Event, v string) bool { return tables.betrayals[v] != nil },
	"filter":   func(w *World, e *Event, v string) bool { return filters[v] != nil },
	"node":     func(w *World, e *Event, v string) bool { return tech.Get(v) != nil },
	"use":      func(w *World, e *Event, v string) bool { return tech.Get(v) != nil || strings.Contains(v, ":") },
	"work":     func(w *World, e *Event, v string) bool { return tech.Structures[v] != nil },
	"eye_work": func(w *World, e *Event, v string) bool { return v == "" || tech.Structures[v] != nil },
	"trait":    func(w *World, e *Event, v string) bool { return species.Get(v) != nil },
	"gained":   func(w *World, e *Event, v string) bool { return v == "" || species.Get(v) != nil },
	"lost":     func(w *World, e *Event, v string) bool { return v == "" || species.Get(v) != nil },
	"miracle":  func(w *World, e *Event, v string) bool { return tables.miracleByKey[v] != nil },
	"object":   func(w *World, e *Event, v string) bool { return tables.miracleByKey[v] != nil },
	"route": func(w *World, e *Event, v string) bool {
		if e.Kind == KMiracleHeld { // how it was gained: a row of miracles.json's routes
			for _, r := range tables.routes {
				if r.Key == v {
					return true
				}
			}
			return false
		}
		return tables.miracleByKey[v] != nil || v == "wound"
	},
	"form":    func(w *World, e *Event, v string) bool { return v != "" },
	"power":   func(w *World, e *Event, v string) bool { return species.PowerByKey(v) != nil },
	"feature": func(w *World, e *Event, v string) bool { return featureByKey(v) != nil },
	"line":    func(w *World, e *Event, v string) bool { return tables.portraitByKey["knowers"][v] != nil },
	"how": func(w *World, e *Event, v string) bool {
		if e.Kind == FMiracle {
			for _, r := range tables.routes {
				if r.Key == v {
					return true
				}
			}
			return false
		}
		return true
	},
	"class":    func(w *World, e *Event, v string) bool { return len(v) == 1 },
	"occasion": func(w *World, e *Event, v string) bool { return tables.leaders["occasions"][v] },
	"end":      func(w *World, e *Event, v string) bool { return tables.leaders["ends"][v] },
	"what":     func(w *World, e *Event, v string) bool { return e.Kind == KReason || e.Kind == KWentDark },
	"species":  func(w *World, e *Event, v string) bool { return false },
}

// keyLists are the parameters that carry lists of keys.
var keyLists = map[string]func(v string) bool{
	"traits":  func(v string) bool { return species.Get(v) != nil },
	"blocks":  func(v string) bool { return aptitudeText(v) != v },
	"cheap":   func(v string) bool { return tech.Get(v) != nil },
	"costly":  func(v string) bool { return tech.Get(v) != nil },
	"powers":  func(v string) bool { return species.PowerByKey(v) != nil },
	"lifted":  func(v string) bool { return v == "sea" || v == "sky" || v == "fire" },
	"facts":   func(v string) bool { return true },
	"nodes":   func(v string) bool { return tech.Get(v) != nil },
	"forgot":  func(v string) bool { return tech.Get(v) != nil },
	"learned": func(v string) bool { return tech.Get(v) != nil },
}

func featureByKey(key string) *galaxy.Feature {
	for _, f := range galaxy.Features {
		if f.Key == key {
			return f
		}
	}
	return nil
}

func looksLikeKey(v string) bool {
	return !strings.ContainsAny(v, " {}'\"") || v == ""
}

// TestKeysResolve: every key an event of the reference run carries, and
// every key its records hold, is a row of a table in data/.
func TestKeysResolve(t *testing.T) {
	t.Parallel()
	w := reference()
	bad := map[string]int{}
	fail := func(where string, v any) {
		bad[where]++
		if bad[where] <= 3 {
			t.Errorf("%s: %v does not resolve", where, v)
		}
	}
	for _, e := range w.Events {
		for k, v := range e.P {
			switch x := v.(type) {
			case string:
				if r, ok := keyParams[k]; ok {
					if !r(w, e, x) {
						fail(string(e.Kind)+"."+k, x)
					}
				} else if !looksLikeKey(x) {
					fail(string(e.Kind)+"."+k+" (text)", x)
				}
			case []string:
				r, ok := keyLists[k]
				if !ok {
					fail(string(e.Kind)+"."+k+" (undeclared list)", x)
					continue
				}
				for _, s := range x {
					if !r(s) {
						fail(string(e.Kind)+"."+k, s)
					}
				}
			}
		}
	}
	for _, c := range w.Civs {
		if c.Cause != "" && tables.causes[c.Cause] == nil {
			fail("Civ.Cause", c.Cause)
		}
		if c.Into != "" && tables.becomings[c.Into] == nil {
			fail("Civ.Into", c.Into)
		}
		if c.Origin.Key != "" && tables.origins[c.Origin.Key] == nil {
			fail("Civ.Origin", c.Origin.Key)
		}
		for k := range c.Scars {
			if tables.scarByKey[k] == nil {
				fail("Civ.Scars", k)
			}
		}
		for k := range c.Boons {
			if tables.boonByKey[k] == nil {
				fail("Civ.Boons", k)
			}
		}
		for k, how := range c.Miracles {
			if tables.miracleByKey[k] == nil {
				fail("Civ.Miracles", k)
			}
			found := false
			for _, r := range tables.routes {
				found = found || r.Key == how
			}
			if !found {
				fail("Civ.Miracles route", how)
			}
		}
		for _, r := range c.Record {
			if r.Filter != "" && filters[r.Filter] == nil {
				fail("Record.Filter", r.Filter)
			}
			switch r.Kind {
			case "faced", "foresaw", "sky", "rest", "bounty", "crewed", "mastered", "wielded", "sealed", "unleashed":
			default:
				fail("Record.Kind", r.Kind)
			}
		}
	}
	for _, sp := range w.Species {
		if sp.Made.Key != "" && tables.origins[sp.Made.Key] == nil {
			fail("Species.Made", sp.Made.Key)
		}
	}
	for _, p := range w.Plagues {
		pr := p.Profile
		if pr.First() == "" || pr.Last() == "" {
			fail("plague.profile", p.ID)
		}
		for _, k := range pr.Symptoms {
			if plague.Profiles.SymptomOf(k) == nil {
				fail("plague.profile.symptom", k)
			}
		}
		for _, k := range pr.Effects {
			if plague.Profiles.EffectOf(k) == nil {
				fail("plague.profile.effect", k)
			}
		}
		if pr.Form != "" && plague.Profiles.FormOf(pr.Form) == nil {
			fail("plague.profile.form", pr.Form)
		}
	}
	for _, l := range w.Legacies {
		list := map[LegacyKind]string{Bounty: "bounties", Threat: "threats", Sleeper: "sleepers", Law: "laws"}[l.Kind]
		switch {
		case l.Kind == Field:
			continue
		case l.Kind == Structure && l.Maker >= 0:
			if tech.Structures[l.Portrait] == nil {
				fail("Legacy.Portrait structure", l.Portrait)
			}
			continue
		case l.Kind == Structure:
			list = "structures"
		case l.Kind == Artifact && l.Maker < 0:
			list = "artifacts"
		case l.Kind == Artifact:
			list = "relics"
			if f := tables.miracleByKey[l.Node]; f != nil && f.Object != "" {
				found := false
				for _, fm := range f.Forms {
					found = found || fm.Key == l.Portrait
				}
				if found {
					continue
				}
			}
		}
		if tables.portraitByKey[list][l.Portrait] == nil {
			fail("Legacy.Portrait "+l.Kind.String(), l.Portrait)
		}
	}
	for _, a := range w.Ages {
		if tables.portraitByKey["enders"][a.Ender] == nil {
			fail("AgeRecord.Ender", a.Ender)
		}
		for _, e := range a.Elders {
			if tables.portraitByKey["elders"][e.Portrait] == nil {
				fail("Elder.Portrait", e.Portrait)
			}
		}
	}
	for _, l := range w.Leaders {
		if !tables.leaders["forms"][l.Form] {
			fail("Leader.Form", l.Form)
		}
		if !tables.leaders["occasions"][l.Occasion] {
			fail("Leader.Occasion", l.Occasion)
		}
		if l.End != "" && !tables.leaders["ends"][l.End] {
			fail("Leader.End", l.End)
		}
		if species.Get(l.Stance) == nil || species.Get(l.Own) == nil {
			fail("Leader.Stance", l.Stance+" "+l.Own)
		}
	}
	for _, wr := range w.Wars {
		if tables.warCauses[wr.Cause] == nil {
			fail("War.Cause", wr.Cause)
		}
		if wr.Result != "" && tables.warResults[wr.Result] == nil {
			fail("War.Result", wr.Result)
		}
	}
	for _, b := range w.Betrayals {
		if b.Weight > 0 && tables.betrayals[b.Shape] == nil {
			fail("Betrayal.Shape", b.Shape)
		}
	}
	for _, tr := range w.Traces {
		if tables.traceByKey[tr.Kind] == nil {
			fail("Trace.Kind", tr.Kind)
		}
	}
	for _, s := range w.Sources {
		key := s.Key
		if strings.HasPrefix(key, "bounty:") {
			key = "bounty"
		}
		if tables.sourceByKey[key] == nil {
			fail("Source.Key", s.Key)
		}
	}
	for _, p := range w.Plagues {
		if !looksLikeKey(p.Cause) {
			fail("Plague.Cause", p.Cause)
		}
	}
	for _, c := range w.Civs {
		for _, in := range c.Infections {
			if _, ok := roads[in.Road]; !ok {
				fail("Infection.Road", in.Road)
			}
		}
	}
	// the static keys: what the code names must be in the tables too
	for key := range filters {
		if tables.filters[key] == nil {
			fail("filter", key)
		}
	}
	for _, n := range tech.Nodes {
		if n.Filter != "" && filters[n.Filter] == nil {
			fail("Node.Filter", n.Filter)
		}
	}
	for _, p := range species.Pool {
		if p.Node != "" && tech.Get(p.Node) == nil {
			fail("Power.Node", p.Node)
		}
		if p.Filter != "" && filters[p.Filter] == nil {
			fail("Power.Filter", p.Filter)
		}
	}
	for k := range traitDiff {
		if species.Get(k) == nil {
			fail("filters.json traits", k)
		}
	}
	for _, a := range aptitudes {
		if !strings.HasPrefix(a.Node, "domain:") && tech.Get(a.Node) == nil {
			fail("aptitudes.json node", a.Node)
		}
	}
}

// recordTypes are the records the simulation holds: the structs that may
// carry no text but names. Every string field of theirs is a key or a
// token, and the test names the fields it knows are neither.
var recordTypes = map[string][]string{
	"types.go":    {"Civ", "Legacy", "Elder", "AgeRecord", "Trace", "Work", "Record"},
	"war.go":      {"War"},
	"plague.go":   {"Plague", "Infection"},
	"sources.go":  {"Source"},
	"pact.go":     {"Betrayal"},
	"garrison.go": {"Muster"},
	"lore.go":     {"Tale", "Inscription"},
	"contract.go": {"Contract"},
}

// textFields are the text a record still holds, each with the step that
// takes it out.
var textFields = map[string]string{
	"Inscription.Text": "the writer step: a testament rendered as it was written, until tellings.jsonl freezes the maker's knowledge instead",
}

var textNames = regexp.MustCompile(`^(Name|Text|Desc|Line|Lines|Meaning|What|Why|Sense|Arising)$`)

// TestNoTextOnRecords: the records carry keys, not words: no string
// field of theirs is named as text, and the ones that are text are
// listed with the step that removes them.
func TestNoTextOnRecords(t *testing.T) {
	fset := token.NewFileSet()
	for file, types := range recordTypes {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		want := map[string]bool{}
		for _, ty := range types {
			want[ty] = true
		}
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok || !want[ts.Name.Name] {
				return true
			}
			delete(want, ts.Name.Name)
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return true
			}
			for _, fd := range st.Fields.List {
				id, ok := fd.Type.(*ast.Ident)
				if !ok || id.Name != "string" {
					continue
				}
				for _, name := range fd.Names {
					full := ts.Name.Name + "." + name.Name
					if _, listed := textFields[full]; listed {
						continue
					}
					if textNames.MatchString(name.Name) {
						t.Errorf("%s is a text field on a record; a key, or a row of textFields with the step that removes it", full)
					}
				}
			}
			return true
		})
		for ty := range want {
			t.Errorf("%s: no type %s", file, ty)
		}
	}
}

// TestNoTextComparisons: no code of the simulation compares a string
// against words: every == and != on a string literal is on a key. The
// view's files are exempt, since words are theirs.
func TestNoTextComparisons(t *testing.T) {
	viewFiles := map[string]bool{"lines.go": true, "describe.go": true, "telling.go": true}
	words := regexp.MustCompile(`(==|!=)\s*"[^"]* [^"]*"|"[^"]* [^"]*"\s*(==|!=)`)
	for _, dir := range []string{".", "../species"} {
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		for _, file := range files {
			base := filepath.Base(file)
			if strings.HasSuffix(base, "_test.go") || viewFiles[base] {
				continue
			}
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			for i, line := range strings.Split(string(src), "\n") {
				if m := words.FindString(line); m != "" {
					t.Errorf("%s:%d compares against words: %s", file, i+1, strings.TrimSpace(line))
				}
			}
		}
	}
}

// TestTablesHeaders: every table in data/ opens with a header that says
// what it holds and the rule for its text.
func TestTablesHeaders(t *testing.T) {
	files, _ := filepath.Glob("../../data/*.json")
	if len(files) < 20 {
		t.Fatalf("%d tables in data/", len(files))
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(src), `"_": [`) {
			t.Errorf("%s has no header", filepath.Base(file))
		}
		if !strings.Contains(string(src), "technical term") {
			t.Errorf("%s's header does not state the term rule", filepath.Base(file))
		}
	}
}
