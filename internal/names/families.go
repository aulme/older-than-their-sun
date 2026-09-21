package names

import (
	"encoding/json"
	"math/rand/v2"
	"sort"
	"strings"

	"worldgen/data"
)

// Family is a sound pattern a reader can tell from one name: an inventory
// of onsets, nuclei and codas, a syllable range, and the devices the
// human transcription uses for it. The table is data/families.json.
type Family struct {
	Key       string             `json:"key"`
	Signature string             `json:"signature"`
	Examples  []string           `json:"examples"`
	Weight    float64            `json:"weight"`
	Onsets    []string           `json:"onsets"`
	Nuclei    []string           `json:"nuclei"`
	Codas     []string           `json:"codas"`
	Syllables [2]int             `json:"syllables"`
	Join      string             `json:"join"`      // the device between syllables: "", "'" or "-"
	JoinOdds  float64            `json:"join_odds"` // how often a boundary gets it
	NoOnset   float64            `json:"no_onset"`  // how often a syllable has no onset
	CodaOdds  float64            `json:"coda_odds"` // how often a syllable has a coda
	Redup     bool               `json:"reduplicate"`
	VowelRun  bool               `json:"vowel_run"` // the family is one long vowel run; nuclei may follow each other
	Exempt    bool               `json:"exempt"`    // from the vowel rules of the convention
	Motivated []string           `json:"motivated"` // drawn only for a people with one of these traits
	Bias      map[string]float64 `json:"bias"`      // trait or archetype key -> weight multiplier
	Subs      map[string]string  `json:"subs"`      // what a sound the family lacks becomes in its mouth
}

var families []*Family
var familyByKey = map[string]*Family{}

func init() {
	raw, err := data.FS.ReadFile("families.json")
	if err != nil {
		panic("names: families.json: " + err.Error())
	}
	var t struct {
		Families []*Family `json:"families"`
	}
	if err := json.Unmarshal(raw, &t); err != nil {
		panic("names: families.json: " + err.Error())
	}
	families = t.Families
	for _, f := range families {
		familyByKey[f.Key] = f
	}
}

// Families lists the families in table order.
func Families() []*Family { return families }

// FamilyByKey finds a family, or nil.
func FamilyByKey(key string) *Family { return familyByKey[key] }

// Phonology is one people's voice: a family and the subset of its
// inventory this people uses, with its own syllable range. A first
// people of a species rolls it from the seed and the species; heirs,
// branches and shards inherit their parent's with a small drift, so a
// line sounds like a line and a schism like a dialect.
type Phonology struct {
	Family    string   `json:"family"`
	Onsets    []string `json:"onsets"`
	Nuclei    []string `json:"nuclei"`
	Codas     []string `json:"codas"`
	Syllables [2]int   `json:"syllables"`
}

func (p *Phonology) family() *Family { return familyByKey[p.Family] }

// drawFamily picks a family for a people by weight, tilted by the body
// and the world, the unpronounceable only where motivated.
func drawFamily(r *rand.Rand, traits []string) *Family {
	has := map[string]bool{}
	for _, t := range traits {
		has[t] = true
	}
	total := 0.0
	ws := make([]float64, len(families))
	for i, f := range families {
		w := f.Weight
		if len(f.Motivated) > 0 {
			ok := false
			for _, m := range f.Motivated {
				ok = ok || has[m]
			}
			if !ok {
				w = 0
			}
		}
		for k, m := range f.Bias {
			if has[k] {
				w *= m
			}
		}
		ws[i] = w
		total += w
	}
	x := r.Float64() * total
	for i, f := range families {
		x -= ws[i]
		if x < 0 {
			return f
		}
	}
	return families[0]
}

// subset keeps about seven in ten of an inventory, at least two, in the
// family's order.
func subset(r *rand.Rand, xs []string) []string {
	if len(xs) <= 2 {
		return append([]string(nil), xs...)
	}
	var out []string
	for _, x := range xs {
		if r.Float64() < 0.7 {
			out = append(out, x)
		}
	}
	for len(out) < 2 {
		out = append(out, xs[r.IntN(len(xs))])
	}
	return out
}

// NewPhonology rolls a people's voice from its own stream: a family for
// the body, a subset of its inventory, the family's range.
func NewPhonology(r *rand.Rand, traits []string) *Phonology {
	f := drawFamily(r, traits)
	return &Phonology{Family: f.Key, Onsets: subset(r, f.Onsets), Nuclei: subset(r, f.Nuclei), Codas: subset(r, f.Codas), Syllables: f.Syllables}
}

// Drift is an heir's voice: the parent's, with one subset re-chosen and
// the range shifted by one; a made people's (uplifted, bred, built)
// drifts twice.
func Drift(r *rand.Rand, parent *Phonology, steps int) *Phonology {
	f := parent.family()
	p := &Phonology{Family: parent.Family, Onsets: parent.Onsets, Nuclei: parent.Nuclei, Codas: parent.Codas, Syllables: parent.Syllables}
	for range steps {
		switch r.IntN(3) {
		case 0:
			p.Onsets = subset(r, f.Onsets)
		case 1:
			p.Nuclei = subset(r, f.Nuclei)
		default:
			p.Codas = subset(r, f.Codas)
		}
		d := r.IntN(2)*2 - 1
		lo, hi := p.Syllables[0]+d, p.Syllables[1]+d
		if lo >= 1 && lo <= f.Syllables[1] && hi >= f.Syllables[0] && hi <= f.Syllables[1]+1 {
			p.Syllables = [2]int{lo, hi}
		}
	}
	return p
}

// syllable is one syllable of a word in this voice.
func (p *Phonology) syllable(r *rand.Rand, first, last bool) string {
	f := p.family()
	s := ""
	if !(first && r.Float64() < f.NoOnset) && !(f.VowelRun && r.Float64() < f.NoOnset) {
		s += pick(r, p.Onsets)
	}
	s += pick(r, p.Nuclei)
	if len(p.Codas) > 0 && r.Float64() < f.CodaOdds {
		s += pick(r, p.Codas)
	}
	_ = last
	return s
}

// Word builds one word of the voice with the given syllable count, or a
// count from the range when n is 0. The result meets the transcription
// convention: the draw is retried until it does, and a last resort is
// the plainest word the inventory allows.
func (p *Phonology) Word(r *rand.Rand, n int) string {
	f := p.family()
	for range 24 {
		w := p.tryWord(r, n)
		if Conventional(w, f.Exempt) {
			return capital(w)
		}
	}
	w := pick(r, p.Onsets) + pick(r, p.Nuclei)
	if !Conventional(w, f.Exempt) {
		w = repair(w)
	}
	return capital(w)
}

func (p *Phonology) tryWord(r *rand.Rand, n int) string {
	f := p.family()
	if n == 0 {
		n = p.Syllables[0] + r.IntN(p.Syllables[1]-p.Syllables[0]+1)
	}
	if f.Exempt {
		return unpronounceable(r, p, n)
	}
	var parts []string
	for i := 0; i < n; i++ {
		parts = append(parts, p.syllable(r, i == 0, i == n-1))
	}
	if f.Redup {
		parts = append([]string{parts[0]}, parts...)
	}
	var b strings.Builder
	for i, s := range parts {
		if i > 0 && f.Join != "" && r.Float64() < f.JoinOdds {
			b.WriteString(f.Join)
		}
		b.WriteString(s)
	}
	return b.String()
}

// unpronounceable is a name a human cannot say: clusters with no vowel,
// stops joined by apostrophes, or a repetition with a tempo.
func unpronounceable(r *rand.Rand, p *Phonology, n int) string {
	switch r.IntN(3) {
	case 0: // Xkrrth'qv
		var parts []string
		for i := 0; i < n; i++ {
			parts = append(parts, pick(r, p.Onsets)+pick(r, p.Codas))
		}
		return strings.Join(parts, "'")
	case 1: // K-k-k-th
		k := pick(r, p.Onsets)
		if len(k) > 1 {
			k = k[:1]
		}
		return strings.Repeat(k+"-", 1+r.IntN(3)) + pick(r, p.Codas)
	default: // Tk'qv
		return pick(r, p.Onsets) + "'" + pick(r, p.Onsets) + pick(r, p.Codas)
	}
}

func isVowel(c byte) bool { return strings.IndexByte("aeiou", c) >= 0 }

// Conventional says whether a transcription meets the convention: ASCII
// letters plus apostrophe and hyphen, never two devices together nor one
// at either end; at most one run of vowels; no doubled vowel; no letter
// three times running; a cluster of at most three consonants; two to
// twelve letters. exempt drops the vowel rules (the unpronounceable
// family).
func Conventional(name string, exempt bool) bool {
	if name == "" {
		return false
	}
	letters, runs, run, cons, same := 0, 0, 0, 0, 0
	var prev byte
	for i := 0; i < len(name); i++ {
		c := name[i] | 0x20
		switch {
		case c >= 'a' && c <= 'z':
			letters++
			if c == prev {
				same++
				if same >= 2 {
					return false
				}
			} else {
				same = 0
			}
			if isVowel(c) {
				run++
				cons = 0
				if run == 2 {
					runs++
				}
				if !exempt && prev == c {
					return false
				}
			} else {
				run = 0
				cons++
				if cons > 3 {
					return false
				}
			}
			prev = c
		case name[i] == '\'' || name[i] == '-':
			if i == 0 || i == len(name)-1 || prev == 0 {
				return false
			}
			run, cons, prev, same = 0, 0, 0, 0
		default:
			return false
		}
	}
	if letters > 12 || letters < 2 {
		return false
	}
	if !exempt && (runs > 1 || strings.IndexAny(name, "aeiouAEIOU") < 0) {
		return false
	}
	return true
}

// repair makes a string conventional by force: doubled vowels collapsed,
// later vowel runs cut to their first vowel, clusters cut to three,
// letters cut to twelve.
func repair(s string) string {
	var b strings.Builder
	runs, run, cons, letters, same := 0, 0, 0, 0, 0
	var prev byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		l := c | 0x20
		switch {
		case l >= 'a' && l <= 'z':
			if l == prev {
				same++
				if same >= 2 {
					continue
				}
			} else {
				same = 0
			}
			if isVowel(l) {
				if prev == l {
					continue
				}
				run++
				cons = 0
				if run == 2 {
					runs++
				}
				if run >= 2 && runs > 1 {
					continue
				}
			} else {
				run = 0
				cons++
				if cons > 3 {
					continue
				}
			}
			letters++
			if letters > 12 {
				break
			}
			prev = l
			b.WriteByte(c)
		case c == '\'' || c == '-':
			if prev == 0 {
				continue // no device at the start or after another
			}
			run, cons, prev, same = 0, 0, 0, 0
			b.WriteByte(c)
		}
	}
	out := strings.Trim(b.String(), "'-")
	if strings.IndexAny(out, "aeiouAEIOU") < 0 {
		out += "a"
	}
	if len(out) < 2 {
		if strings.IndexAny(out, "aeiouAEIOU") >= 0 {
			out += "n"
		} else {
			out += "a"
		}
	}
	return out
}

// Adopt renders another people's name in this voice: the source is cut
// into syllables and each onset, nucleus and coda is kept if the voice
// has it, substituted by the family's table if not, and dropped
// otherwise; the family's devices are applied. Kailoa in a cluster
// mouth is Kaloth; in a glottal one Ka'i'o.
func (p *Phonology) Adopt(r *rand.Rand, source string) string {
	f := p.family()
	src := strings.ToLower(source)
	if f.Exempt {
		// no vowels: keep the consonants, joined as stops
		var cons []string
		for _, part := range strings.FieldsFunc(src, func(c rune) bool { return isVowel(byte(c)) || c == '\'' || c == '-' }) {
			if part != "" {
				cons = append(cons, part)
			}
		}
		if len(cons) == 0 {
			cons = []string{pick(r, p.Onsets)}
		}
		w := strings.Join(cons, "'")
		if !Conventional(w, true) {
			w = repair(w)
		}
		return capital(w)
	}
	var syls []string
	for _, syl := range syllabify(src) {
		on, nu, co := syl[0], syl[1], syl[2]
		on = p.fit(r, on, p.Onsets, f)
		if len(p.Codas) == 0 || co == "" {
			co = ""
		} else {
			co = p.fit(r, co, p.Codas, f)
		}
		// a long vowel run the voice has no nucleus for is broken into
		// nuclei with a glide between them: Aeioa in an open mouth is Aeyiho
		if len(nu) > 2 && !p.hasNucleus(nu) && !f.VowelRun {
			for i := 0; i < len(nu); i += 2 {
				chunk := nu[i:min(i+2, len(nu))]
				o := on
				if i > 0 {
					o = p.glide(r)
				}
				c := ""
				if i+2 >= len(nu) {
					c = co
				}
				syls = append(syls, o+p.fitNucleus(r, chunk)+c)
			}
			continue
		}
		syls = append(syls, on+p.fitNucleus(r, nu)+co)
	}
	if len(syls) == 0 {
		syls = []string{pick(r, p.Onsets) + pick(r, p.Nuclei)}
	}
	if f.Redup {
		syls = append([]string{syls[0]}, syls...)
	}
	var b strings.Builder
	for i, s := range syls {
		if i > 0 && f.Join != "" && r.Float64() < f.JoinOdds {
			b.WriteString(f.Join)
		}
		b.WriteString(s)
	}
	w := b.String()
	if !Conventional(w, false) {
		w = repair(w)
	}
	return capital(w)
}

// fit is a consonant cluster in this voice: kept if the voice has it;
// else the family's substitute; else the voice's nearest sound, one
// that starts with the same letter or with the substitute's; else the
// first letter alone; else whatever the voice says in that place.
func (p *Phonology) fit(r *rand.Rand, c string, inv []string, f *Family) string {
	if c == "" {
		return ""
	}
	has := func(x string) bool {
		for _, y := range inv {
			if y == x {
				return true
			}
		}
		return false
	}
	starts := func(x string) string {
		if x == "" {
			return ""
		}
		for _, y := range inv {
			if strings.HasPrefix(y, x[:1]) {
				return y
			}
		}
		return ""
	}
	if has(c) {
		return c
	}
	if s, ok := f.Subs[c]; ok {
		if s == "" || has(s) || strings.ContainsAny(s, "'-") {
			return s
		}
		if y := starts(s); y != "" {
			return y
		}
	}
	if y := starts(c); y != "" {
		return y
	}
	if s, ok := f.Subs[c[:1]]; ok {
		if s == "" || has(s) || strings.ContainsAny(s, "'-") {
			return s
		}
		if y := starts(s); y != "" {
			return y
		}
	}
	return pick(r, inv)
}

func (p *Phonology) hasNucleus(v string) bool {
	for _, y := range p.Nuclei {
		if y == v {
			return true
		}
	}
	return false
}

// glide is the softest onset the voice has, for breaking a vowel run.
func (p *Phonology) glide(r *rand.Rand) string {
	for _, g := range []string{"y", "w", "h", "l", "r"} {
		for _, o := range p.Onsets {
			if o == g {
				return g
			}
		}
	}
	return pick(r, p.Onsets)
}

func (p *Phonology) fitNucleus(r *rand.Rand, v string) string {
	for _, y := range p.Nuclei {
		if y == v {
			return v
		}
	}
	if len(v) > 1 {
		for _, y := range p.Nuclei {
			if y == v[:1] {
				return y
			}
		}
	}
	return pick(r, p.Nuclei)
}

// syllabify cuts a lower-case transcription into (onset, nucleus, coda)
// triples: each vowel run is a nucleus; the consonants before the first
// are its onset; consonants between two nuclei go to the next onset
// except the first, which stays as a coda when there are two or more.
func syllabify(s string) [][3]string {
	s = strings.NewReplacer("'", "", "-", "").Replace(s)
	var out [][3]string
	i := 0
	var pending string // consonants before the current nucleus
	for i < len(s) {
		j := i
		for j < len(s) && !isVowel(s[j]) {
			j++
		}
		cons := s[i:j]
		k := j
		for k < len(s) && isVowel(s[k]) {
			k++
		}
		nucleus := s[j:k]
		if nucleus == "" {
			// trailing consonants: the last syllable's coda
			if n := len(out); n > 0 {
				out[n-1][2] += cons
			} else if cons != "" {
				out = append(out, [3]string{cons, "", ""})
			}
			break
		}
		if len(out) > 0 && len(cons) >= 2 {
			out[len(out)-1][2] = cons[:1]
			cons = cons[1:]
		}
		_ = pending
		out = append(out, [3]string{cons, nucleus, ""})
		i = k
	}
	return out
}

// Classify assigns a transcription to the family it most resembles by
// its letter classes, for the test that keeps the families distinct.
// Features: vowel share, longest consonant cluster, apostrophes,
// hyphens, doubled letters, length, sibilant share, nasal share, liquid
// share, a repeated syllable.
func Features(name string) []float64 {
	s := strings.ToLower(name)
	letters, vowels, sib, nas, liq, dbl, apos, hyph := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	cluster, longest := 0, 0
	var prev byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\'':
			apos++
			cluster = 0
		case c == '-':
			hyph++
			cluster = 0
		case isVowel(c):
			letters++
			vowels++
			cluster = 0
		default:
			letters++
			cluster++
			longest = max(longest, cluster)
			if strings.IndexByte("szx", c) >= 0 || (c == 'h' && (prev == 's' || prev == 'z' || prev == 't')) {
				sib++
			}
			if c == 'm' || c == 'n' {
				nas++
			}
			if strings.IndexByte("lrwy", c) >= 0 {
				liq++
			}
		}
		if c == prev && c != '\'' && c != '-' {
			dbl++
		}
		prev = c
	}
	redup := 0.0
	if n := len(s); n >= 4 {
		half := strings.NewReplacer("-", "", "'", "").Replace(s)
		for l := 2; l <= len(half)/2; l++ {
			if half[:l] == half[l:2*l] {
				redup = 1
			}
		}
	}
	if letters == 0 {
		letters = 1
	}
	return []float64{vowels / letters, float64(longest), apos, hyph, dbl, letters, sib / letters, nas / letters, liq / letters, redup}
}

// sortedKeys is for deterministic walks of a map.
func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
