package names

import (
	"math"
	"math/rand/v2"
	"sort"
	"strings"

	"worldgen/data"
	"worldgen/internal/history"
	"worldgen/internal/plague"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Entry is one recipe of the epithet inventory (data/epithets.json): a
// pattern with slots, for the objects and tones it names, allowed only
// where every required property holds of the thing as the namer knows
// it, and weighted by what the namer notices.
type Entry struct {
	ID       string            `json:"id"`
	About    string            `json:"about"` // civ, star, elder, makers, plague, title, war, word, source
	Tone     []string          `json:"tone"`
	Requires []string          `json:"requires,omitempty"` // property predicates; a leading ! negates
	Pattern  string            `json:"pattern"`
	Slots    map[string]string `json:"slots,omitempty"`  // hole -> source: bank:<name>, or a property of the thing
	Notice   []string          `json:"notice,omitempty"` // the semantic fields the pattern uses
	Weight   float64           `json:"weight,omitempty"`
}

type inventory struct {
	Notice     map[string]map[string]float64 `json:"notice"`     // a namer's trait, world or organisation -> field -> weight
	Properties []string                      `json:"properties"` // the properties the floor test counts
	Entries    []*Entry                      `json:"entries"`
}

var epithets = loadEpithets()
var banks = loadBanks()

func loadEpithets() *inventory {
	inv := &inventory{}
	data.Load("epithets.json", inv)
	for _, e := range inv.Entries {
		if e.Weight == 0 {
			e.Weight = 1
		}
	}
	return inv
}

func loadBanks() map[string][]string {
	var t struct {
		Banks map[string][]string `json:"banks"`
	}
	data.Load("banks.json", &t)
	return t.Banks
}

// Epithets lists the inventory's entries, for the tests.
func Epithets() []*Entry { return epithets.Entries }

// Properties lists the counted properties, for the tests.
func Properties() []string { return epithets.Properties }

// Banks is the word lists by field, for the tests.
func Banks() map[string][]string { return banks }

// props is what a namer knows of a thing at the year it names it: the
// property keys that hold, and the values a slot may read (a star, a
// people, a word of the profile).
type props struct {
	has  map[string]bool
	vals map[string]string
}

func newProps() *props { return &props{has: map[string]bool{}, vals: map[string]string{}} }

func (p *props) set(k string)    { p.has[k] = true }
func (p *props) val(k, v string) { p.vals[k] = v }
func (p *props) holds(req string) bool {
	if strings.HasPrefix(req, "!") {
		return !p.has[req[1:]]
	}
	return p.has[req]
}

// Knows says whether the property holds, for the tests.
func (p *props) Knows(k string) bool { return p.has[k] }

// notice is a namer's weights per semantic field, derived from its
// senses, its world and its organisation by the inventory's header.
func (b *Book) notice(by int) map[string]float64 {
	if n, ok := b.noticeOf[by]; ok {
		return n
	}
	n := map[string]float64{}
	if by >= 0 {
		c := b.w.Civs[by]
		keys := traitKeys(c.Species)
		keys = append(keys, c.Species.Sub.String())
		for _, d := range c.Species.Mods.Defs() {
			keys = append(keys, d.Key)
		}
		for _, k := range keys {
			for field, w := range epithets.Notice[k] {
				if cur, ok := n[field]; ok {
					n[field] = cur * w
				} else {
					n[field] = w
				}
			}
		}
	}
	b.noticeOf[by] = n
	return n
}

// choose picks an entry for the thing: those whose requirements hold,
// scored by weight and by what the namer notices, the top few drawn
// from by score.
func (b *Book) choose(r *rand.Rand, by int, about, tone string, k *props) *Entry {
	n := b.notice(by)
	type cand struct {
		e     *Entry
		score float64
	}
	var cands []cand
	for _, e := range epithets.Entries {
		if e.About != about || !contains(e.Tone, tone) {
			continue
		}
		ok := true
		for _, req := range e.Requires {
			if !k.holds(req) {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		s := e.Weight
		for _, f := range e.Notice {
			if w, has := n[f]; has {
				s *= w
			}
		}
		if s > 0 {
			cands = append(cands, cand{e, s})
		}
	}
	if len(cands) == 0 {
		return nil
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
	if len(cands) > 6 {
		cut := cands[5].score
		n := 6
		for n < len(cands) && cands[n].score == cut {
			n++ // a tie is not cut by the order of the file
		}
		cands = cands[:n]
	}
	total := 0.0
	for _, c := range cands {
		total += c.score
	}
	x := r.Float64() * total
	for _, c := range cands {
		x -= c.score
		if x < 0 {
			return c.e
		}
	}
	return cands[len(cands)-1].e
}

// render fills an entry's pattern: a bank word drawn from the stream, or
// a value of the thing as the namer knows it.
func (b *Book) render(r *rand.Rand, e *Entry, by int, k *props) (string, Recipe) {
	name := e.Pattern
	var filled []string
	holes := sortedKeys(e.Slots)
	used := map[string]bool{}
	for _, hole := range holes {
		src := e.Slots[hole]
		var v string
		if bank, ok := strings.CutPrefix(src, "bank:"); ok {
			words := banks[bank]
			v = pick(r, words)
			for i := 0; i < 4 && used[v] && len(words) > 1; i++ {
				v = pick(r, words)
			}
			used[v] = true
		} else {
			v = k.vals[src]
		}
		name = strings.ReplaceAll(name, "{"+hole+"}", v)
		filled = append(filled, hole+"="+v)
	}
	return name, Recipe{Entry: e.ID, Pattern: e.Pattern, Slots: strings.Join(filled, "; ")}
}

// translate makes one translated row, or none when no entry fits.
func (b *Book) translate(by int, o Object, tone string, y history.Year, k *props) (Row, bool) {
	r := stream(b.w.Seed, itoa(by), o.Kind, itoa(o.ID), tone)
	e := b.choose(r, by, o.Kind, tone, k)
	if e == nil {
		return Row{}, false
	}
	name, rec := b.render(r, e, by, k)
	return Row{Object: o, By: by, Name: name, Mode: "translated", Tone: tone, Coined: y, From: -1, Recipe: rec}, true
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// tok is a name token: what the namer calls the thing.
func tok(kind string, id, by int) string { return "{" + kind + ":" + itoa(id) + "@" + itoa(by) + "}" }

// ---- what a namer knows ----

// between is every fact with the two as its parties, in order, indexed
// once; the deeds between two peoples are read from it.
func (b *Book) between(a, c int) []*history.Event {
	if a > c {
		a, c = c, a
	}
	return b.pairs[[2]int{a, c}]
}

func (b *Book) indexPairs() {
	b.pairs = map[[2]int][]*history.Event{}
	for _, e := range b.w.Events {
		if !e.IsFact() || e.Subject < 0 || e.Object < 0 || e.Subject == e.Object {
			continue
		}
		a, c := e.Subject, e.Object
		if a > c {
			a, c = c, a
		}
		b.pairs[[2]int{a, c}] = append(b.pairs[[2]int{a, c}], e)
	}
}

// selfProps is everything a people knows of itself: its body, its
// world, its ways, its kind and its voice.
func (b *Book) selfProps(c *history.Civ) *props {
	k := newProps()
	b.bodyProps(k, c)
	b.worldProps(k, c)
	b.wayProps(k, c)
	b.kindProps(k, c)
	k.val("home", tok("star", c.Home, c.ID))
	k.val("cradle", tok("star", c.Cradle, c.ID))
	k.val("self", tok("civ", c.ID, c.ID))
	return k
}

func (b *Book) bodyProps(k *props, c *history.Civ) {
	for _, t := range c.Species.Traits {
		if t.Group == "bio" || t.Group == "sense" {
			k.set("trait:" + t.Key)
		}
	}
}

func (b *Book) worldProps(k *props, c *history.Civ) {
	if c.Species.World != nil {
		k.set("world:" + c.Species.World.Key)
	}
	for _, t := range c.Species.Traits {
		if t.Group != "world" {
			continue
		}
		switch t.Key {
		case "skyless", "fireless", "threesuns":
			k.set("world:" + t.Key)
		default:
			k.set("trait:" + t.Key) // what the world made of the body: robust, fragile, hardy, cooperative
		}
	}
	if s := &b.w.G.Stars[c.Cradle]; s.Mult > 1 {
		k.set("world:manysuns")
	}
}

func (b *Book) wayProps(k *props, c *history.Civ) {
	for _, t := range c.Species.Traits {
		switch t.Group {
		case "org", "stance", "honour", "drive", "way":
			k.set("way:" + t.Key)
		}
	}
	switch c.Morality.Kind {
	case history.Fixation:
		k.set("fix:" + strings.ReplaceAll(c.Morality.Object, " ", "_"))
	default:
		k.set("fix:" + c.Morality.Kind.String())
	}
}

func (b *Book) kindProps(k *props, c *history.Civ) {
	sp := c.Species
	k.set("kind:" + sp.Sub.String())
	for _, d := range sp.Mods.Defs() {
		k.set("kind:" + d.Key)
	}
	if sp.Sub == species.Biological && sp.Mods == 0 {
		k.set("kind:plain")
	}
	ch := sp.Channel
	if ch == "" {
		ch = species.Sound
	}
	k.set("voice:" + ch)
	k.set("voice:" + string(b.voice[c.ID]))
}

// civProps is what a namer knows of another people at a year: its kind
// and voice from any meeting, its body and world from a meeting in the
// flesh or an understanding, its ways from an understanding or a war,
// and every deed between them, the heaviest first for the slots.
func (b *Book) civProps(by int, o *history.Civ, y history.Year) *props {
	k := newProps()
	b.kindProps(k, o)
	k.val("cradle", tok("star", o.Cradle, by))
	k.val("home", tok("star", b.w.Civs[by].Home, by))
	k.val("self", tok("civ", by, by))
	k.val("them", tok("civ", o.ID, by))
	touched, fathomed, warred := false, false, false
	var heavy *history.Event
	for _, e := range b.between(by, o.ID) {
		if e.Year > y {
			break
		}
		switch e.Kind {
		case history.FMet:
			if e.P["how"] == "touch" {
				touched = true
			}
		case history.FFathomed:
			if e.Subject == by {
				fathomed = true
			}
		case history.FWar:
			warred = true
		}
		var key string
		switch {
		case e.Sort() == history.Bond:
			key = "bond:" + string(e.Kind)
		case e.Subject == o.ID:
			key = "deed:" + string(e.Kind)
		default:
			key = "did:" + string(e.Kind)
		}
		k.set(key)
		if e.Sort() != history.Bond && (heavy == nil || e.Weight() > heavy.Weight()) {
			heavy = e
		}
	}
	if touched || fathomed {
		b.bodyProps(k, o)
		b.worldProps(k, o)
	}
	if fathomed || warred {
		b.wayProps(k, o)
	}
	if heavy != nil {
		k.set("heavy:" + string(heavy.Kind))
		if heavy.Star >= 0 {
			k.set("deed.star")
			k.val("deed.star", tok("star", heavy.Star, by))
		}
	}
	return k
}

// starProps is what a namer knows of a star: its class and company,
// and what the namer did there.
func (b *Book) starProps(by, star int, place string) *props {
	k := newProps()
	s := &b.w.G.Stars[star]
	k.set("class:" + string(s.Class))
	switch s.Class {
	case 'O', 'B', 'A', 'F':
		k.set("bright")
	case 'M':
		k.set("dim")
	}
	if s.Class == 'N' || s.Class == 'W' {
		k.set("dead")
		if s.Remnant != "" {
			k.set("remnant:" + s.Remnant)
		}
	}
	if s.Mult > 1 {
		k.set("mult:many")
	}
	if s.Mult == 2 {
		k.set("mult:2")
	}
	if s.Mult == 3 {
		k.set("mult:3")
	}
	if place != "" {
		k.set("place:" + place)
	}
	c := b.w.Civs[by]
	k.val("self", tok("civ", by, by))
	k.val("home", tok("star", c.Home, by))
	k.val("cradle", tok("star", c.Cradle, by))
	return k
}

// workProps is what a finder knows of a remain: its kind, the node of
// the work, its look.
func (b *Book) workProps(by int, l *history.Legacy) *props {
	k := newProps()
	k.set("legacy:" + strings.ToLower(l.Kind.String()))
	if l.Node != "" {
		k.set("work:" + l.Node)
		if n := tech.Get(l.Node); n != nil {
			k.val("work", n.Name)
		}
	}
	if l.Portrait != "" {
		k.set("portrait:" + l.Portrait)
	}
	if l.Elder != nil {
		k.set("elder:" + l.Elder.Portrait)
	}
	k.val("star", tok("star", l.Star, by))
	return k
}

// plagueProps is what a namer knows of a sickness: its profile, and
// the host it came from when it knows one.
func (b *Book) plagueProps(by int, p *history.Plague, host int) *props {
	k := newProps()
	k.set("kind:" + p.Kind.String())
	pr := p.Profile
	t := plague.Profiles
	for i, sy := range pr.Symptoms {
		k.set("symptom:" + sy)
		if i == 0 {
			k.set("first:" + sy)
		}
	}
	if len(pr.Symptoms) > 0 {
		last := pr.Symptoms[len(pr.Symptoms)-1]
		k.set("last:" + last)
		if s := t.SymptomOf(last); s != nil {
			k.val("last", s.Noun)
		}
		if s := t.SymptomOf(pr.Symptoms[0]); s != nil {
			k.val("first", s.Noun)
			k.val("first.adj", s.Adj)
		}
		if m := pr.Visible(); m != "" {
			k.set("marked")
			k.set("mark:" + m)
			k.val("mark", t.SymptomOf(m).Adj)
		}
		k.set("onset:" + pr.Onset)
		k.set("course:" + pr.Course)
		k.set("takes:" + pr.Takes)
		if pr.Onset != "" {
			for _, o := range t.Onsets {
				if o.Key == pr.Onset {
					k.val("onset", o.Adj)
				}
			}
		}
	}
	if pr.Form != "" {
		k.set("form:" + pr.Form)
		if f := t.FormOf(pr.Form); f != nil {
			k.val("form", f.Noun)
		}
		for _, ef := range pr.Effects {
			k.set("effect:" + ef)
		}
		if len(pr.Effects) > 0 {
			k.set("last:" + pr.Effects[len(pr.Effects)-1])
		}
		k.set("carrier:" + pr.Carrier)
	}
	if p.Made && p.Maker >= 0 && p.Maker != by {
		k.set("made")
	}
	if host >= 0 && host != by {
		k.set("host")
		k.val("host", tok("civ", host, by))
		k.val("host.star", tok("star", b.w.Civs[host].Cradle, by))
	}
	if p.Conscious {
		k.set("thinks")
	}
	if by >= 0 {
		k.val("cradle", tok("star", b.w.Civs[by].Cradle, by))
		k.val("self", tok("civ", by, by))
	}
	return k
}

// ---- the rows ----

// translatedRows is every translated name the record entitles: endonyms
// and glosses, exonyms in their tones, stars by relationship and sky,
// elders and remains by work, plagues by profile, titles, wars per side,
// the words for the state beneath, the objects a translated voice made.
func (b *Book) translatedRows() {
	w := b.w
	// the glosses: a transcribed name carries a meaning one time in three
	for _, c := range w.Civs {
		if b.voice[c.ID] != Transcribed {
			continue
		}
		k := b.selfProps(c)
		o := Object{Kind: "civ", ID: c.ID}
		r := stream(w.Seed, itoa(c.ID), "gloss", itoa(c.ID), "self")
		if r.IntN(3) == 0 {
			if e := b.choose(r, c.ID, "civ", "self", k); e != nil {
				gloss, _ := b.render(r, e, c.ID, k)
				b.gloss(o, c.ID, gloss)
			}
		}
	}
	for _, f := range w.Events {
		if !f.IsFact() {
			continue
		}
		switch f.Kind {
		case history.FMet:
			if f.Object >= 0 {
				b.exonym(f.Subject, f.Object, "stranger", f.Year)
				if f.P["how"] != "noticed" {
					b.exonym(f.Object, f.Subject, "stranger", f.Year)
				}
			}
		case history.FPact, history.FTrade, history.FRelief:
			if f.Object >= 0 {
				b.friend(f.Subject, f.Object, f.Year)
				b.friend(f.Object, f.Subject, f.Year)
			}
		case history.FWar:
			if f.Object >= 0 {
				b.enemy(f.Subject, f.Object, f.Year, false)
				b.enemy(f.Object, f.Subject, f.Year, false)
				b.warRows(f)
			}
		case history.FSettle:
			b.starTranslated(f.Subject, f.Star, "self", "settled", f.Year)
		case history.FTaken:
			b.starTranslated(f.Subject, f.Star, "stranger", "taken", f.Year)
			if f.Object >= 0 {
				b.starTranslated(f.Object, f.Star, "stranger", "lost", f.Year)
			}
		case history.FBurned:
			if f.Object >= 0 {
				b.starTranslated(f.Object, f.Star, "stranger", "burned", f.Year)
			}
		case history.FSurveyLost:
			b.starTranslated(f.Subject, f.Star, "stranger", "silent", f.Year)
		case history.FFind:
			b.findRows(f)
		case history.FPlague:
			if p := w.Plagues[f.Plague]; p != nil {
				b.plagueTranslated(f.Subject, p, f.Object, "self", f.Year)
			}
		case history.FFall:
			b.title(f.Subject, f.Year)
		case history.FWord:
			b.word(f)
		}
		if f.Sort() == history.Crime && f.Object >= 0 && f.Subject >= 0 && f.Subject != f.Object {
			b.enemy(f.Object, f.Subject, f.Year, f.Weight() >= 4)
		}
	}
	// a war names the star it is fought over, for the sides
	for _, wr := range w.Wars {
		if wr.Named >= 0 {
			for _, side := range wr.Sides {
				b.starTranslated(side, wr.Named, "stranger", "fought", wr.Began)
			}
		}
	}
	// plagues heard of: a people that holds a tale of another's catching
	for _, c := range w.Civs {
		for _, t := range c.Lore {
			f := w.Events[t.Fact]
			if f.Kind != history.FPlague || f.Subject == c.ID || f.Plague < 0 {
				continue
			}
			if _, had := c.Infections[f.Plague]; had {
				continue
			}
			b.plagueTranslated(c.ID, w.Plagues[f.Plague], f.Subject, "stranger", t.Learned)
		}
	}
	b.skyRows()
	b.objectRows()
}

// gloss puts a meaning beside a transcribed row.
func (b *Book) gloss(o Object, by int, gloss string) {
	rs := b.rows[o]
	for i := range rs {
		if rs[i].By == by && rs[i].Tone == "self" {
			rs[i].Gloss = gloss
		}
	}
}

// exonym is a people's name for another it has met, in a tone.
func (b *Book) exonym(by, other int, tone string, y history.Year) {
	if by == other || b.voice[by] == None || b.has(Object{Kind: "civ", ID: other}, by, tone) {
		return
	}
	k := b.civProps(by, b.w.Civs[other], y)
	if row, ok := b.translate(by, Object{Kind: "civ", ID: other}, tone, y, k); ok {
		b.add(row)
	}
}

// friend is the name for a voiceless people at the first bond with it:
// there is no endonym to adopt. A voiced people's friend row is its
// endonym, adopted at the meeting.
func (b *Book) friend(by, other int, y history.Year) {
	if b.voice[other] != None {
		return
	}
	b.exonym(by, other, "friend", y)
}

// enemy is the name at the first war or crime: monster when the crime
// was a heavy one.
func (b *Book) enemy(by, other int, y history.Year, monster bool) {
	o := Object{Kind: "civ", ID: other}
	if b.has(o, by, "enemy") || b.has(o, by, "monster") {
		return
	}
	tone := "enemy"
	if monster {
		tone = "monster"
	}
	b.exonym(by, other, tone, y)
}

func (b *Book) has(o Object, by int, tone string) bool {
	for _, r := range b.rows[o] {
		if r.By == by && r.Tone == tone {
			return true
		}
	}
	return false
}

// starTranslated is a people's name for a star by what it did there; a
// transcribed voice names its own stars in its own sounds (starRow) and
// the others by relationship.
func (b *Book) starTranslated(by, star int, tone, place string, y history.Year) {
	if by < 0 || star < 0 || b.voice[by] == None {
		return
	}
	o := Object{Kind: "star", ID: star}
	if b.has(o, by, tone) {
		return
	}
	k := b.starProps(by, star, place)
	if row, ok := b.translate(by, o, tone, y, k); ok {
		b.add(row)
	}
}

// findRows is what a finder calls what it found: the makers of an elder
// work by the work; the makers of a remain of this age when it does not
// know them; and, when it does and they had a voice, their own names for
// themselves and for the star, read from the walls.
func (b *Book) findRows(f *history.Event) {
	w := b.w
	if f.Legacy < 0 || b.voice[f.Subject] == None {
		return
	}
	l := w.Legacies[f.Legacy]
	by := f.Subject
	k := b.workProps(by, l)
	switch {
	case l.Elder != nil:
		if row, ok := b.translate(by, Object{Kind: "elder", ID: l.Elder.ID}, "stranger", f.Year, k); ok {
			b.add(row)
		}
	case l.Maker < 0:
	case f.P["known"] == true:
		if l.Maker == by {
			return
		}
		b.adoptFrom(by, l.Maker, Object{Kind: "civ", ID: l.Maker}, f.Year)
		b.adoptFrom(by, l.Maker, Object{Kind: "star", ID: l.Star}, f.Year)
	default:
		if row, ok := b.translate(by, Object{Kind: "makers", ID: l.ID}, "stranger", f.Year, k); ok {
			b.add(row)
		}
	}
}

// adoptFrom is a name read from a wall: the makers' own name for the
// thing, in the reader's mouth.
func (b *Book) adoptFrom(by, from int, o Object, y history.Year) {
	if b.has(o, by, "stranger") || b.has(o, by, "self") {
		return
	}
	var src *Row
	for i := range b.rows[o] {
		r := &b.rows[o][i]
		if r.By == from && r.Tone == "self" && r.Coined <= y {
			src = r
			break
		}
	}
	if src == nil {
		return
	}
	name := src.Name
	r := stream(b.w.Seed, itoa(by), o.Kind, itoa(o.ID), "read")
	if b.voice[by] == Transcribed && src.Mode == "transcribed" {
		name = b.phon[by].Adopt(r, name)
	}
	b.add(Row{Object: o, By: by, Name: name, Mode: "adopted", Tone: "stranger", Coined: y, From: from, Gloss: src.Gloss})
}

// plagueTranslated is a sufferer's name for a sickness, or a hearer's.
func (b *Book) plagueTranslated(by int, p *history.Plague, host int, tone string, y history.Year) {
	if by < 0 || b.voice[by] == None {
		return
	}
	o := Object{Kind: "plague", ID: p.ID}
	if b.has(o, by, tone) {
		return
	}
	if host < 0 && p.FirstHost >= 0 && p.FirstHost != by {
		host = p.FirstHost
	}
	if p.Made && p.Maker >= 0 && p.Maker != by {
		host = p.Maker
	}
	k := b.plagueProps(by, p, host)
	if row, ok := b.translate(by, o, tone, y, k); ok {
		b.add(row)
	}
}

// title is a remnant's ruler's title, from its fate's cause and its ways.
func (b *Book) title(id int, y history.Year) {
	c := b.w.Civs[id]
	if b.voice[id] == None || c.Fate != history.Contracted {
		return
	}
	k := b.selfProps(c)
	if c.Cause != "" {
		k.set("cause:" + c.Cause)
	}
	if row, ok := b.translate(id, Object{Kind: "title", ID: id}, "self", y, k); ok {
		b.add(row)
	}
}

// warRows is each side's name for a war: from its cause, and the star
// it is fought over.
func (b *Book) warRows(f *history.Event) {
	w := b.w
	id, _ := f.P["war"].(int)
	if id < 0 || id >= len(w.Wars) {
		return
	}
	wr := w.Wars[id]
	for i, side := range wr.Sides {
		if b.voice[side] == None {
			continue
		}
		k := newProps()
		k.set("cause:" + wr.Cause)
		if i == 0 {
			k.set("side:declared")
		} else {
			k.set("side:attacked")
		}
		if wr.Named >= 0 {
			k.set("named")
			k.val("star", tok("star", wr.Named, side))
		}
		if wr.Nth > 1 {
			k.set("again")
		}
		other := wr.Sides[1-i]
		k.val("enemy", tok("civ", other, side))
		k.val("self", tok("civ", side, side))
		if row, ok := b.translate(side, Object{Kind: "war", ID: wr.ID}, "self", wr.Began, k); ok {
			b.add(row)
		}
	}
}

// word is the word a people coins for the state beneath, from the
// miracle it reached in through; an heir keeps the old people's word.
func (b *Book) word(f *history.Event) {
	w := b.w
	c := w.Civs[f.Subject]
	if b.voice[c.ID] == None {
		return
	}
	root := f.Subject
	for _, a := range c.Line {
		if b.hasWord(a) {
			root = a
			break
		}
	}
	k := b.selfProps(c)
	route, _ := f.P["route"].(string)
	k.set("miracle:" + route)
	r := stream(w.Seed, itoa(root), "word", itoa(root), "self")
	e := b.choose(r, c.ID, "word", "self", k)
	if e == nil {
		return
	}
	name, rec := b.render(r, e, c.ID, k)
	b.add(Row{Object: Object{Kind: "word", ID: c.ID}, By: c.ID, Name: name, Mode: "translated", Tone: "self", Coined: f.Year, From: -1, Recipe: rec})
}

func (b *Book) hasWord(id int) bool {
	for _, f := range b.w.Events {
		if f.Kind == history.FWord && f.Subject == id {
			return true
		}
	}
	return false
}

// objectRows names the objects the miracles make: in the maker's voice,
// transcribed or by a recipe on the form; a cutting carries the giver's
// name down the chain, adopted.
func (b *Book) objectRows() {
	w := b.w
	for _, s := range w.Sources {
		if s.Kind != history.MadeSource || s.Maker < 0 || s.Form == "" || b.voice[s.Maker] == None {
			continue
		}
		o := Object{Kind: "source", ID: s.ID}
		if b.voice[s.Maker] == Transcribed {
			r := stream(w.Seed, itoa(s.Maker), "source", itoa(s.ID), "self")
			b.add(Row{Object: o, By: s.Maker, Name: b.phon[s.Maker].Word(r, 1+r.IntN(2)), Mode: "transcribed", Tone: "self", Coined: s.Made, From: -1})
			continue
		}
		k := newProps()
		k.set("form:" + s.Form)
		k.set("miracle:" + s.Key)
		if s.Sentient {
			k.set("thinks")
		}
		k.val("self", tok("civ", s.Maker, s.Maker))
		if row, ok := b.translate(s.Maker, o, "self", s.Made, k); ok {
			b.add(row)
		}
	}
	for _, f := range w.Events {
		if f.Kind != history.KCutting || f.Object < 0 {
			continue
		}
		from, _ := f.P["source"].(int)
		if from < 0 || from >= len(w.Sources) {
			continue
		}
		for _, s := range w.Sources {
			if s.Maker != f.Object || s.Made != f.Year || s.Key != "manna" || s.Form == "" {
				continue
			}
			o := Object{Kind: "source", ID: s.ID}
			if b.has(o, f.Object, "self") || b.voice[f.Object] == None {
				continue
			}
			var src *Row
			for i := range b.rows[Object{Kind: "source", ID: from}] {
				r := &b.rows[Object{Kind: "source", ID: from}][i]
				if r.By == f.Subject {
					src = r
					break
				}
			}
			if src == nil {
				continue
			}
			name := src.Name
			r := stream(w.Seed, itoa(f.Object), "source", itoa(s.ID), "adopted")
			if b.voice[f.Object] == Transcribed && src.Mode == "transcribed" {
				name = b.phon[f.Object].Adopt(r, name)
			}
			b.add(Row{Object: o, By: f.Object, Name: name, Mode: "adopted", Tone: "self", Coined: f.Year, From: f.Subject})
		}
	}
}

// skyRows are the old names: the stars a people could see with the naked
// eye from its cradle before it reached them, the brightest twelve, named
// by their colour and their company. A real star's brightness from the
// cradle follows from its catalogue magnitude and its distance from the
// Sun; a synthetic star has no magnitude and is not named from afar.
const skyNames = 12

func (b *Book) skyRows() {
	w := b.w
	g := w.G
	type seen struct {
		star int
		mag  float64
	}
	for _, c := range w.Civs {
		if b.voice[c.ID] == None || (c.Species.Has("skyless") || c.Species.Has("eyeless")) {
			continue
		}
		var sky []seen
		for i := range g.Stars {
			s := &g.Stars[i]
			if i == c.Cradle || !s.Real || s.Mag >= 99 {
				continue
			}
			dSun := g.Dist(g.Sol, i)
			dHere := g.Dist(c.Cradle, i)
			if dSun <= 0 || dHere <= 0 {
				continue
			}
			mag := s.Mag + 5*math.Log10(dHere/dSun)
			if mag <= 6 {
				sky = append(sky, seen{i, mag})
			}
		}
		sort.Slice(sky, func(i, j int) bool {
			if sky[i].mag != sky[j].mag {
				return sky[i].mag < sky[j].mag
			}
			return sky[i].star < sky[j].star
		})
		if len(sky) > skyNames {
			sky = sky[:skyNames]
		}
		for _, x := range sky {
			o := Object{Kind: "star", ID: x.star}
			if b.has(o, c.ID, "self") {
				continue
			}
			k := b.starProps(c.ID, x.star, "")
			k.set("seen")
			if x.mag < 1 {
				k.set("brilliant")
			}
			if row, ok := b.translate(c.ID, o, "sky", c.Born, k); ok {
				b.add(row)
			}
		}
	}
}
