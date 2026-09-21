package names

import (
	"regexp"
	"sort"
	"strings"

	"worldgen/internal/galaxy"
	"worldgen/internal/history"
	"worldgen/internal/plague"
	"worldgen/internal/species"
)

// Object is a thing that can be named: a kind and the id the simulation
// gave it.
type Object struct {
	Kind string `json:"kind"` // civ, star, plague, war, elder, legacy, source
	ID   int    `json:"id"`
}

// Human is the namer of a designation row.
const Human = -2

// Row is one name for one object by one culture.
type Row struct {
	Object Object       `json:"object"`
	By     int          `json:"by"` // a people's id, or Human for the catalogue
	Name   string       `json:"name"`
	Mode   string       `json:"mode"` // transcribed, translated, designation, adopted
	Tone   string       `json:"tone"` // self, stranger, friend, enemy, monster, sky, none
	Coined history.Year `json:"coined"`
	From   int          `json:"from"`            // the culture an adopted name was learned from, or -1
	Gloss  string       `json:"gloss,omitempty"` // the meaning a transcribed name would also be given
	Stub   bool         `json:"stub,omitempty"`  // a placeholder from the old generators until the translated pass lands
}

// Voice is how a culture names: none, transcribed or translated.
type Voice string

const (
	None        Voice = "none"
	Transcribed Voice = "transcribed"
	Translated  Voice = "translated"
)

// Book is the names of one run: every row the record entitles, and the
// voice and phonology of every people. It is built once from the record
// and answers the view.
type Book struct {
	w     *history.World
	rows  map[Object][]Row
	voice map[int]Voice
	phon  map[int]*Phonology
}

// Of builds the book of a world.
func Of(w *history.World) *Book {
	b := &Book{w: w, rows: map[Object][]Row{}, voice: map[int]Voice{}, phon: map[int]*Phonology{}}
	for _, c := range w.Civs {
		b.voice[c.ID] = VoiceOf(c.Species)
	}
	for _, c := range w.Civs {
		b.phonology(c)
	}
	b.selfRows()
	b.designations()
	b.factRows()
	b.stubs()
	return b
}

// Rows lists an object's rows in the order they were coined.
func (b *Book) Rows(o Object) []Row { return b.rows[o] }

// All lists every row, by object kind, id and coining year.
func (b *Book) All() []Row {
	var out []Row
	keys := make([]Object, 0, len(b.rows))
	for k := range b.rows {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Kind != keys[j].Kind {
			return keys[i].Kind < keys[j].Kind
		}
		return keys[i].ID < keys[j].ID
	})
	for _, k := range keys {
		out = append(out, b.rows[k]...)
	}
	return out
}

// VoiceOf derives a culture's voice from its species: a consequence of
// what the people is, never rolled.
func VoiceOf(sp *species.Species) Voice {
	switch {
	case sp.Is(species.Hive), sp.Is(species.Unconscious), sp.Is(species.Planetary), sp.Is(species.Replicator), sp.Sub == species.Eldritch:
		return None
	case sp.Sub == species.Machine:
		return Translated
	case sp.Sub == species.Parasite:
		return Transcribed // through its host
	case sp.Has("deaf"), sp.Channel != "" && sp.Channel != species.Sound:
		return Translated
	}
	return Transcribed
}

// Voice is a people's voice.
func (b *Book) Voice(c *history.Civ) Voice { return b.voice[c.ID] }

// Phonology is a people's voice as sounds, or nil for a voiceless one.
func (b *Book) Phonology(c *history.Civ) *Phonology { return b.phon[c.ID] }

func traitKeys(sp *species.Species) []string {
	var out []string
	for _, t := range sp.Traits {
		out = append(out, t.Key)
	}
	if sp.World != nil {
		out = append(out, sp.World.Key)
	}
	return out
}

// phonology is a people's voice as sounds: a first people of a species
// rolls it from the seed and the species; an heir, a branch or a shard
// inherits its parent's with one step of drift; a made people takes its
// maker's with two; a parasite speaks in its first host's.
func (b *Book) phonology(c *history.Civ) *Phonology {
	if p, ok := b.phon[c.ID]; ok {
		return p
	}
	w := b.w
	var p *Phonology
	b.phon[c.ID] = nil // against a cycle in the record
	switch {
	case len(c.Line) > 0:
		if parent := b.phonology(w.Civs[c.Line[len(c.Line)-1]]); parent != nil {
			p = Drift(stream(w.Seed, "phon", "civ", itoa(c.ID)), parent, 1)
		}
	case c.Species.Sub == species.Parasite && c.Own >= 0:
		if host := w.Plagues[c.Own].FirstHost; host >= 0 && host != c.ID {
			if hp := b.phonology(w.Civs[host]); hp != nil {
				p = Drift(stream(w.Seed, "phon", "civ", itoa(c.ID)), hp, 1)
			}
		}
	case c.Sire >= 0 && c.Sire != c.ID:
		if mp := b.phonology(w.Civs[c.Sire]); mp != nil {
			p = Drift(stream(w.Seed, "phon", "civ", itoa(c.ID)), mp, 2)
		}
	}
	if p == nil {
		// the first people of a species, or a branch of it, rolls from the species
		first := c.ID
		for _, o := range w.Civs {
			if o.Species == c.Species {
				first = o.ID
				break
			}
		}
		if first != c.ID && w.Civs[first] != c {
			if fp := b.phonology(w.Civs[first]); fp != nil {
				p = Drift(stream(w.Seed, "phon", "civ", itoa(c.ID)), fp, 1)
			}
		}
		if p == nil {
			p = NewPhonology(stream(w.Seed, "phon", "species", itoa(c.Species.ID)), traitKeys(c.Species))
		}
	}
	b.phon[c.ID] = p
	return p
}

func (b *Book) add(r Row) {
	rs := b.rows[r.Object]
	for _, x := range rs {
		if x.By == r.By && x.Tone == r.Tone {
			return // one name per (culture, object, tone)
		}
	}
	b.rows[r.Object] = append(rs, r)
}

// selfRows is every people's own name for itself, in its own voice; a
// voiceless people has none. A translated voice's endonym is a stub
// until the translated pass: the syllabary stands in.
func (b *Book) selfRows() {
	w := b.w
	for _, c := range w.Civs {
		v := b.voice[c.ID]
		if v == None {
			continue
		}
		p := b.phon[c.ID]
		r := stream(w.Seed, itoa(c.ID), "civ", itoa(c.ID), "self")
		row := Row{Object: Object{"civ", c.ID}, By: c.ID, Name: p.Word(r, 0), Mode: string(v), Tone: "self", Coined: c.Born, From: -1}
		if v == Translated {
			row.Stub = true
		}
		b.add(row)
	}
}

// designations are the human catalogue's rows: a proper name is a strong
// human name, a catalogue label a weak one.
func (b *Book) designations() {
	for i := range b.w.G.Stars {
		s := &b.w.G.Stars[i]
		if !s.Real || s.Name == "" {
			continue
		}
		mode := "designation"
		if s.Proper() {
			mode = "proper"
		}
		b.add(Row{Object: Object{"star", s.ID}, By: Human, Name: s.Name, Mode: mode, Tone: "none", Coined: 0, From: -1})
	}
}

// factRows walks the facts for the relationships that entitle a name:
// a cradle and a settled star are named by their people; a meeting
// adopts the other's endonym.
func (b *Book) factRows() {
	w := b.w
	for _, f := range w.Events {
		switch f.Kind {
		case history.FArise:
			b.starRow(f.Subject, f.Star, f.Year)
		case history.FSettle:
			b.starRow(f.Subject, f.Star, f.Year)
		case history.FMet:
			if f.Object >= 0 {
				b.adopt(f.Subject, f.Object, f.Year)
				if f.P["how"] != "noticed" { // a one-sided finding: the other never knew
					b.adopt(f.Object, f.Subject, f.Year)
				}
			}
		}
	}
}

// starRow is a people's name for a star of its own: one or two syllables
// of its voice; a voiceless people names nothing.
func (b *Book) starRow(by, star int, y history.Year) {
	if star < 0 || by < 0 {
		return
	}
	v := b.voice[by]
	if v == None {
		return
	}
	p := b.phon[by]
	r := stream(b.w.Seed, itoa(by), "star", itoa(star), "self")
	row := Row{Object: Object{"star", star}, By: by, Name: p.Word(r, 1+r.IntN(2)), Mode: string(v), Tone: "self", Coined: y, From: -1}
	if v == Translated {
		row.Stub = true
	}
	b.add(row)
}

// adopt is a people learning another's endonym at a meeting: transcribed
// through its own voice. A voiceless learner adopts nothing; a voiceless
// other has nothing to adopt, and gets a stranger row from the old
// syllabary instead, a stub until the translated pass.
func (b *Book) adopt(by, other int, y history.Year) {
	if by == other || b.voice[by] == None {
		return
	}
	self := b.selfName(other)
	r := stream(b.w.Seed, itoa(by), "civ", itoa(other), "adopted")
	if self == "" {
		b.add(Row{Object: Object{"civ", other}, By: by, Name: b.phon[by].Word(r, 0), Mode: "translated", Tone: "stranger", Coined: y, From: -1, Stub: true})
		return
	}
	name := self
	if b.voice[by] == Transcribed && b.voice[other] == Transcribed {
		name = b.phon[by].Adopt(r, self)
	}
	b.add(Row{Object: Object{"civ", other}, By: by, Name: name, Mode: "adopted", Tone: "stranger", Coined: y, From: other})
}

func (b *Book) selfName(id int) string {
	for _, r := range b.rows[Object{"civ", id}] {
		if r.By == id && r.Tone == "self" {
			return r.Name
		}
	}
	return ""
}

// stubs are the descriptive names the translated pass will make from
// recipes, meanwhile made by the old generators on the hash stream so
// the legends read as before: plagues, titles, wars, the words for the
// state beneath, finder names for elders and for makers nobody knew.
func (b *Book) stubs() {
	w := b.w
	for _, p := range w.Plagues {
		b.plagueRow(p)
	}
	for _, c := range w.Civs {
		if c.Fate == history.Contracted {
			r := stream(w.Seed, itoa(c.ID), "title", itoa(c.ID), "self")
			b.add(Row{Object: Object{"title", c.ID}, By: c.ID, Name: title(r), Mode: "translated", Tone: "self", Coined: c.Ended, From: -1, Stub: true})
		}
	}
	for _, f := range w.Events {
		switch f.Kind {
		case history.FWord:
			c := w.Civs[f.Subject]
			// an heir keeps the old people's word: the row hashes on the first of the line to have one
			root := f.Subject
			for _, a := range c.Line {
				if b.hasWord(a) {
					root = a
					break
				}
			}
			r := stream(w.Seed, itoa(root), "word", itoa(root), "self")
			b.add(Row{Object: Object{"word", c.ID}, By: c.ID, Name: pick(r, beneathWords), Mode: "translated", Tone: "self", Coined: f.Year, From: -1, Stub: true})
		case history.FFind:
			if f.Legacy < 0 {
				continue
			}
			l := w.Legacies[f.Legacy]
			if l.Elder != nil {
				r := stream(w.Seed, itoa(f.Subject), "elder", itoa(l.Elder.ID), "stranger")
				b.add(Row{Object: Object{"elder", l.Elder.ID}, By: f.Subject, Name: pick(r, finderNames[l.Kind]), Mode: "translated", Tone: "stranger", Coined: f.Year, From: -1, Stub: true})
			} else if !f.P["known"].(bool) {
				r := stream(w.Seed, itoa(f.Subject), "makers", itoa(l.ID), "stranger")
				b.add(Row{Object: Object{"makers", l.ID}, By: f.Subject, Name: pick(r, ruinNames), Mode: "translated", Tone: "stranger", Coined: f.Year, From: -1, Stub: true})
			}
		}
	}
	for _, wr := range w.Wars {
		if wr.Named >= 0 {
			b.add(Row{Object: Object{"war", wr.ID}, By: wr.Sides[0], Name: "the war of " + b.Default(Object{"star", wr.Named}), Mode: "translated", Tone: "self", Coined: wr.Began, From: -1, Stub: true})
		}
	}
	for _, s := range w.Sources {
		if s.Kind == history.MadeSource && s.Maker >= 0 && b.voice[s.Maker] != None {
			r := stream(w.Seed, itoa(s.Maker), "source", itoa(s.ID), "self")
			b.add(Row{Object: Object{"source", s.ID}, By: s.Maker, Name: b.phon[s.Maker].Word(r, 1+r.IntN(2)), Mode: string(b.voice[s.Maker]), Tone: "self", Coined: s.Made, From: -1})
		}
	}
}

func (b *Book) hasWord(id int) bool {
	for _, f := range b.w.Events {
		if f.Kind == history.FWord && f.Subject == id {
			return true
		}
	}
	return false
}

// plagueRow is the old plague generator on the hash stream: a made
// plague is its maker's gift or lie, in order; a born one is named by
// its first host, one time in three for the host or the host's star.
func (b *Book) plagueRow(p *history.Plague) {
	w := b.w
	by := p.FirstHost
	if p.Maker >= 0 && p.Made {
		by = p.Maker
	}
	if by < 0 {
		by = p.Maker
	}
	r := stream(w.Seed, itoa(by), "plague", itoa(p.ID), "self")
	var name string
	switch {
	case p.Made && p.Maker >= 0:
		word := "Gift"
		if p.Kind == plague.Memetic {
			word = "Lie"
		}
		n := 1
		for _, q := range w.Plagues {
			if q.ID < p.ID && q.Maker == p.Maker && q.Made && q.Kind == p.Kind {
				n++
			}
		}
		name = "the " + b.Default(Object{"civ", p.Maker}) + " " + word
		if n > 1 {
			name = "the " + ordinal(n) + " " + b.Default(Object{"civ", p.Maker}) + " " + word
		}
	default:
		host := ""
		if p.FirstHost >= 0 && !p.Made {
			h := w.Civs[p.FirstHost]
			host = b.Default(Object{"civ", h.ID})
			if r.IntN(2) == 0 {
				host = b.Default(Object{"star", h.Cradle})
			}
		}
		name = plagueName(r, p.Kind == plague.Memetic, host)
	}
	b.add(Row{Object: Object{"plague", p.ID}, By: max(by, -1), Name: name, Mode: "translated", Tone: "self", Coined: p.Born, From: -1, Stub: true})
}

// Default is the debug view's one name for a thing: the human proper
// name; else the earliest self row; else the earliest row of any tone;
// else the designation; else the id.
func (b *Book) Default(o Object) string {
	rows := b.rows[o]
	for _, r := range rows {
		if r.By == Human && r.Mode == "proper" {
			return r.Name
		}
	}
	if r := earliest(rows, "self"); r != nil {
		return r.Name
	}
	if r := earliest(rows, ""); r != nil {
		return r.Name
	}
	for _, r := range rows {
		if r.By == Human {
			return r.Name
		}
	}
	return b.fallback(o)
}

// By is what one culture calls a thing: its own row for it, else the
// default.
func (b *Book) By(o Object, by int) string {
	var best *Row
	for i := range b.rows[o] {
		r := &b.rows[o][i]
		if r.By != by {
			continue
		}
		if best == nil || r.Tone == "self" || (best.Tone != "self" && r.Coined < best.Coined) {
			best = r
		}
	}
	if best != nil {
		return best.Name
	}
	return b.Default(o)
}

func earliest(rows []Row, tone string) *Row {
	var best *Row
	for i := range rows {
		r := &rows[i]
		if r.By == Human || (tone != "" && r.Tone != tone) {
			continue
		}
		if best == nil || r.Coined < best.Coined {
			best = r
		}
	}
	return best
}

// fallback is what a thing with no name is called: a star by its
// catalogue label, anything else by its kind and id.
func (b *Book) fallback(o Object) string {
	switch o.Kind {
	case "star":
		if o.ID >= 0 && o.ID < len(b.w.G.Stars) {
			return b.w.G.Stars[o.ID].Name
		}
	case "war", "word", "title":
		return ""
	case "species":
		for _, c := range b.w.Civs {
			if c.Species.ID == o.ID {
				return b.Default(Object{"civ", c.ID})
			}
		}
	case "elder":
		return "the Ones Before"
	case "makers":
		return "the Forgotten"
	}
	return "unnamed #" + itoa(o.ID)
}

var tokenRe = regexp.MustCompile(`\{(\^?)([a-z]+):(-?\d+)(?:@(-?\d+))?\}`)

// Text resolves every name token in a text: {kind:id} to the default
// name, {kind:id@by} to what that culture calls it, {^kind:id} with the
// first letter raised. Names may hold tokens themselves, so it runs to
// a fixed point.
func (b *Book) Text(s string) string {
	for range 4 {
		if !strings.Contains(s, "{") {
			return s
		}
		s = tokenRe.ReplaceAllStringFunc(s, func(m string) string {
			g := tokenRe.FindStringSubmatch(m)
			o := Object{Kind: g[2], ID: atoi(g[3])}
			var name string
			if g[4] != "" {
				name = b.By(o, atoi(g[4]))
			} else {
				name = b.Default(o)
			}
			if g[1] == "^" {
				name = capital(name)
			}
			return name
		})
	}
	return s
}

func atoi(s string) int {
	n, neg := 0, false
	for i := 0; i < len(s); i++ {
		if s[i] == '-' {
			neg = true
			continue
		}
		n = n*10 + int(s[i]-'0')
	}
	if neg {
		return -n
	}
	return n
}

// Star is the view's name for a star, with the human label after an
// alien name where the star has one and it is not a proper name:
// "Ur-Kesh, HIP 56601 on human maps".
func (b *Book) Star(s *galaxy.Star) string {
	o := Object{"star", s.ID}
	name := b.Default(o)
	if s.Real && s.Name != "" && name != s.Name {
		return name + " (" + s.Name + " on human maps)"
	}
	return name
}
