package names

import (
	"regexp"
	"sort"
	"strings"

	"worldgen/internal/galaxy"
	"worldgen/internal/history"
	"worldgen/internal/plague"
	"worldgen/internal/record"
	"worldgen/internal/species"
)

// Object is a thing that can be named: a kind and the id the simulation
// gave it. Row is one name for one object by one culture, a row of the
// state's names[]; Recipe how a translated name was made. The types are
// the record's: the pass writes them and the view reads them.
type (
	Object = record.Object
	Row    = record.Name
	Recipe = record.Recipe
	Voice  = record.Voice
)

// Human is the namer of a designation row.
const Human = record.Human

// The voices: none, transcribed or translated.
const (
	None        = record.None
	Transcribed = record.Transcribed
	Translated  = record.Translated
)

// Table is the names of one run as a reader resolves them: every row by
// object, and what a thing with no row is called. The view builds one
// from the state's names[] (FromRows); the pass builds one from the
// record (Of) and answers the same questions.
type Table struct {
	rows map[Object][]Row
	// what the fallbacks need of the world: a star's human label, whether
	// a plague is of the mind, the first people of a blood
	starLabel  func(id int) (string, bool)
	memetic    func(id int) bool
	speciesCiv func(id int) int
}

// FromRows builds a table from the state's rows. The fallbacks read the
// state through the three functions; any may be nil.
func FromRows(rows []Row, starLabel func(int) (string, bool), memetic func(int) bool, speciesCiv func(int) int) *Table {
	t := &Table{rows: map[Object][]Row{}, starLabel: starLabel, memetic: memetic, speciesCiv: speciesCiv}
	for _, r := range rows {
		t.rows[r.Object] = append(t.rows[r.Object], r)
	}
	return t
}

// Book is the names of one run: every row the record entitles, and the
// voice and phonology of every people. It is built once from the record
// and answers the view.
type Book struct {
	*Table
	w        *history.World
	voice    map[int]Voice
	phon     map[int]*Phonology
	noticeOf map[int]map[string]float64
	pairs    map[[2]int][]*history.Event
}

// Of builds the book of a world: the voices, the transcribed rows, the
// designations, the adopted endonyms, then every translated row.
func Of(w *history.World) *Book {
	t := &Table{rows: map[Object][]Row{}}
	t.starLabel = func(id int) (string, bool) {
		if id >= 0 && id < len(w.G.Stars) {
			return w.G.Stars[id].Name, true
		}
		return "", false
	}
	t.memetic = func(id int) bool { return id >= 0 && id < len(w.Plagues) && w.Plagues[id].Kind == plague.Memetic }
	t.speciesCiv = func(id int) int {
		for _, c := range w.Civs {
			if c.Species.ID == id {
				return c.ID
			}
		}
		return -1
	}
	b := &Book{Table: t, w: w, voice: map[int]Voice{}, phon: map[int]*Phonology{}, noticeOf: map[int]map[string]float64{}}
	for _, c := range w.Civs {
		b.voice[c.ID] = VoiceOf(c.Species)
	}
	for _, c := range w.Civs {
		b.phonology(c)
	}
	b.selfRows()
	b.designations()
	b.factRows()
	b.translatedRows()
	return b
}

// Rows lists an object's rows in the order they were coined.
func (b *Table) Rows(o Object) []Row { return b.rows[o] }

// All lists every row, by object kind, id and coining year.
func (b *Table) All() []Row {
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
	case sp.Voiceless():
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

// selfRows is every transcribed people's own name for itself, in its
// own sounds; a translated voice's endonym is a recipe (translated.go)
// and a voiceless people has none.
func (b *Book) selfRows() {
	w := b.w
	for _, c := range w.Civs {
		if b.voice[c.ID] != Transcribed {
			continue
		}
		p := b.phon[c.ID]
		r := stream(w.Seed, itoa(c.ID), "civ", itoa(c.ID), "self")
		b.add(Row{Object: Object{Kind: "civ", ID: c.ID}, By: c.ID, Name: p.Word(r, 0), Mode: "transcribed", Tone: "self", Coined: c.Born, From: -1})
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
		b.add(Row{Object: Object{Kind: "star", ID: s.ID}, By: Human, Name: s.Name, Mode: mode, Tone: "none", Coined: 0, From: -1})
	}
}

// factRows walks the facts for the transcribed and adopted rows: a
// cradle and a settled star are named by their people in its own
// sounds; a meeting adopts the other's endonym. The translated rows
// follow (translated.go), once every endonym exists to be adopted.
func (b *Book) factRows() {
	w := b.w
	for _, f := range w.Events {
		switch f.Kind {
		case history.FArise:
			b.starRow(f.Subject, f.Star, f.Year)
		case history.FSettle:
			b.starRow(f.Subject, f.Star, f.Year)
		}
	}
	// the endonyms of translated voices exist before any adoption
	b.indexPairs()
	for _, c := range w.Civs {
		if b.voice[c.ID] == Translated {
			if row, ok := b.translate(c.ID, Object{Kind: "civ", ID: c.ID}, "self", c.Born, b.selfProps(c)); ok {
				b.add(row)
			}
		}
	}
	for _, f := range w.Events {
		if f.Kind == history.FMet && f.Object >= 0 {
			b.adopt(f.Subject, f.Object, f.Year)
			if f.P["how"] != "noticed" { // a one-sided finding: the other never knew
				b.adopt(f.Object, f.Subject, f.Year)
			}
		}
	}
}

// starRow is a transcribed people's name for a star of its own: one or
// two syllables of its voice. A translated voice names its cradle by a
// recipe; a voiceless people names nothing.
func (b *Book) starRow(by, star int, y history.Year) {
	if star < 0 || by < 0 {
		return
	}
	switch b.voice[by] {
	case None:
		return
	case Translated:
		b.starTranslated(by, star, "self", "cradle", y)
		return
	}
	p := b.phon[by]
	r := stream(b.w.Seed, itoa(by), "star", itoa(star), "self")
	b.add(Row{Object: Object{Kind: "star", ID: star}, By: by, Name: p.Word(r, 1+r.IntN(2)), Mode: "transcribed", Tone: "self", Coined: y, From: -1})
}

// adopt is a people learning another's endonym at a meeting, as its
// friend row: transcribed through its own voice, or copied unchanged
// when it is already in our words. A voiceless learner adopts nothing;
// a voiceless other has nothing to adopt, and its friend row is a
// translation at the first bond (translated.go).
func (b *Book) adopt(by, other int, y history.Year) {
	if by == other || b.voice[by] == None {
		return
	}
	self := b.selfName(other)
	if self == "" {
		return
	}
	r := stream(b.w.Seed, itoa(by), "civ", itoa(other), "adopted")
	name := self
	if b.voice[by] == Transcribed && b.voice[other] == Transcribed {
		name = b.phon[by].Adopt(r, self)
	}
	b.add(Row{Object: Object{Kind: "civ", ID: other}, By: by, Name: name, Mode: "adopted", Tone: "friend", Coined: y, From: other})
}

func (b *Book) selfName(id int) string {
	for _, r := range b.rows[Object{Kind: "civ", ID: id}] {
		if r.By == id && r.Tone == "self" {
			return r.Name
		}
	}
	return ""
}

// Default is the debug view's one name for a thing: the human proper
// name; else the earliest self row; else the earliest row of any tone;
// else the designation; else the id.
func (b *Table) Default(o Object) string {
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

// By is what one culture calls a thing: its own row for it in the tone
// asked for, else down the ladder of regard (a monster is at least an
// enemy, an enemy at least a stranger); with no tone, its own name for
// its own thing, else the endonym it adopted, else its exonym; else the
// default.
func (b *Table) By(o Object, by int, tone string) string {
	order := []string{"self", "friend", "stranger", "sky", "enemy", "monster"}
	switch tone {
	case "monster":
		order = []string{"monster", "enemy", "stranger", "friend", "self"}
	case "enemy":
		order = []string{"enemy", "monster", "stranger", "friend", "self"}
	case "stranger":
		order = []string{"stranger", "self", "friend", "enemy", "monster"}
	case "friend":
		order = []string{"friend", "self", "stranger", "enemy", "monster"}
	}
	for _, t := range order {
		for i := range b.rows[o] {
			r := &b.rows[o][i]
			if r.By == by && r.Tone == t {
				return r.Name
			}
		}
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
func (b *Table) fallback(o Object) string {
	switch o.Kind {
	case "star":
		if b.starLabel != nil {
			if label, ok := b.starLabel(o.ID); ok {
				return label
			}
		}
	case "war":
		return "a war"
	case "leader":
		return "the one who led them" // a voiceless people's, which it never named
	case "word":
		return ""
	case "title":
		return "nameless"
	case "source":
		return "nothing" // a voiceless maker calls its object nothing
	case "plague":
		if b.memetic != nil && b.memetic(o.ID) {
			return "a nameless idea"
		}
		return "a nameless sickness"
	case "species":
		if b.speciesCiv != nil {
			if c := b.speciesCiv(o.ID); c >= 0 {
				return b.Default(Object{Kind: "civ", ID: c})
			}
		}
	case "elder":
		return "the Ones Before"
	case "makers":
		return "the Forgotten"
	}
	return "unnamed #" + itoa(o.ID)
}

var tokenRe = regexp.MustCompile(`\{(\^?)([a-z]+):(-?\d+)(?:@(-?\d+)(?::([a-z]+))?)?\}`)

// Text resolves every name token in a text: {kind:id} to the default
// name, {kind:id@by} to what that culture calls it, {kind:id@by:tone}
// to what it calls it in that regard (a telling's slant), {^kind:id}
// with the first letter raised. Names may hold tokens themselves, so it
// runs to a fixed point.
func (b *Table) Text(s string) string {
	for range 4 {
		if !strings.Contains(s, "{") {
			return s
		}
		s = tokenRe.ReplaceAllStringFunc(s, func(m string) string {
			g := tokenRe.FindStringSubmatch(m)
			o := Object{Kind: g[2], ID: atoi(g[3])}
			var name string
			if g[4] != "" {
				name = b.By(o, atoi(g[4]), g[5])
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
func (b *Table) Star(s *galaxy.Star) string {
	o := Object{Kind: "star", ID: s.ID}
	name := b.Default(o)
	if s.Real && s.Name != "" && name != s.Name {
		return name + " (" + s.Name + " on human maps)"
	}
	return name
}
