package legends

import (
	"worldgen/data"
	"worldgen/internal/record"
)

// The text the view reads from the lookups in data/: the words beside
// the keys the record holds. The simulation reads the same files for
// their mechanics and never these fields.

type (
	keyText struct {
		Key  string `json:"key"`
		Text string `json:"text"`
	}
	keyName struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}
	miracleText struct {
		Key    string    `json:"key"`
		Term   string    `json:"term"`
		Object string    `json:"object"`
		Forms  []keyDesc `json:"forms"`
	}
	keyDesc struct {
		Key  string `json:"key"`
		Desc string `json:"desc"`
	}
	sourceText struct {
		Key   string `json:"key"`
		Name  string `json:"name"`
		Frame string `json:"frame"`
	}
	band struct {
		Word  string  `json:"word"`
		Below float64 `json:"below"`
	}
	stageWord struct {
		From float64 `json:"from"`
		Word string  `json:"word"`
	}
	kindShape struct {
		Kind   record.Kind `json:"kind"`
		Sort   string      `json:"sort"`
		Weight float64     `json:"weight"`
	}
)

var tables = struct {
	scars, boons, filters map[string]string
	portraits             map[string]map[string]string
	conditions            []keyText
	condText              map[string]string
	traces                map[string]string
	miracles              map[string]*miracleText
	sources               map[string]*sourceText
	causes, blasts        map[string]string
	becomings, warCauses  map[string]string
	warResults, betrayals map[string]string
	origins               map[string]string
	bands                 []band
	stiffness, wall       []stageWord
	aptitudes             map[string]string
	kinds                 map[record.Kind]kindShape
	leaderForms           map[string]leaderForm
	leaderOccasions       map[string]string
}{}

// leaderForm is a row of leaders.json's forms: what the view calls one,
// and what a telling worn to myth calls it.
type leaderForm struct {
	Key       string `json:"key"`
	Word      string `json:"word"`
	Archetype string `json:"archetype"`
}

func init() {
	names := func(rows []keyName) map[string]string {
		m := map[string]string{}
		for _, r := range rows {
			m[r.Key] = r.Name
		}
		return m
	}
	texts := func(rows []keyText) map[string]string {
		m := map[string]string{}
		for _, r := range rows {
			m[r.Key] = r.Text
		}
		return m
	}
	var sf struct {
		Scars []keyName `json:"scars"`
	}
	data.Load("scars.json", &sf)
	tables.scars = names(sf.Scars)
	var bf struct {
		Boons []keyName `json:"boons"`
	}
	data.Load("boons.json", &bf)
	tables.boons = names(bf.Boons)
	var ff struct {
		Filters []keyName `json:"filters"`
	}
	data.Load("filters.json", &ff)
	tables.filters = names(ff.Filters)
	var pf map[string]any
	data.Load("portraits.json", &pf)
	tables.portraits = map[string]map[string]string{}
	for list, rows := range pf {
		if list == "_" {
			continue
		}
		m := map[string]string{}
		for _, row := range rows.([]any) {
			r := row.(map[string]any)
			text, _ := r["text"].(string)
			m[r["key"].(string)] = text
		}
		tables.portraits[list] = m
	}
	var cf struct {
		Conditions []keyText `json:"conditions"`
		Traces     []keyText `json:"traces"`
	}
	data.Load("conditions.json", &cf)
	tables.conditions = cf.Conditions
	tables.condText = texts(cf.Conditions)
	tables.traces = texts(cf.Traces)
	var mf struct {
		Miracles []*miracleText `json:"miracles"`
	}
	data.Load("miracles.json", &mf)
	tables.miracles = map[string]*miracleText{}
	for _, m := range mf.Miracles {
		tables.miracles[m.Key] = m
	}
	var srf struct {
		Sources []*sourceText `json:"sources"`
	}
	data.Load("sources.json", &srf)
	tables.sources = map[string]*sourceText{}
	for _, s := range srf.Sources {
		tables.sources[s.Key] = s
	}
	var cf2 struct {
		Causes     []keyText `json:"causes"`
		Blasts     []keyText `json:"blasts"`
		Becomings  []keyText `json:"becomings"`
		WarCauses  []keyText `json:"war_causes"`
		WarResults []keyText `json:"war_results"`
		Betrayals  []keyText `json:"betrayals"`
	}
	data.Load("causes.json", &cf2)
	tables.causes, tables.blasts, tables.becomings = texts(cf2.Causes), texts(cf2.Blasts), texts(cf2.Becomings)
	tables.warCauses, tables.warResults, tables.betrayals = texts(cf2.WarCauses), texts(cf2.WarResults), texts(cf2.Betrayals)
	var of struct {
		Origins []keyText `json:"origins"`
	}
	data.Load("origins.json", &of)
	tables.origins = texts(of.Origins)
	var lf struct {
		Bands     []band      `json:"bands"`
		Stiffness []stageWord `json:"stiffness"`
		Wall      []stageWord `json:"wall"`
	}
	data.Load("levels.json", &lf)
	tables.bands, tables.stiffness, tables.wall = lf.Bands, lf.Stiffness, lf.Wall
	var af struct {
		Aptitudes []keyText `json:"aptitudes"`
	}
	data.Load("aptitudes.json", &af)
	tables.aptitudes = map[string]string{}
	for _, a := range af.Aptitudes {
		if a.Key != "" {
			tables.aptitudes[a.Key] = a.Text
		}
	}
	var ef struct {
		Kinds []kindShape `json:"kinds"`
	}
	data.Load("events.json", &ef)
	tables.kinds = map[record.Kind]kindShape{}
	for _, k := range ef.Kinds {
		tables.kinds[k.Kind] = k
	}
	var ldf struct {
		Forms     []leaderForm `json:"forms"`
		Occasions []struct {
			Key  string `json:"key"`
			Line string `json:"line"`
		} `json:"occasions"`
	}
	data.Load("leaders.json", &ldf)
	tables.leaderForms, tables.leaderOccasions = map[string]leaderForm{}, map[string]string{}
	for _, f := range ldf.Forms {
		tables.leaderForms[f.Key] = f
	}
	for _, o := range ldf.Occasions {
		tables.leaderOccasions[o.Key] = o.Line
	}
}

// sortOf is a fact's own sort; weightOf its weight. A kind that is no
// fact is "nothing" at weight zero.
func sortOf(e *record.Event) string {
	if k, ok := tables.kinds[e.Kind]; ok && k.Weight > 0 {
		return k.Sort
	}
	return "nothing"
}

// textOf is a keyed text from a table, or the key.
func textOf(m map[string]string, key string) string {
	if t, ok := m[key]; ok {
		return t
	}
	return key
}

// portraitText is the text of a portrait key in a list, or the key.
func portraitText(list, key string) string {
	if t, ok := tables.portraits[list][key]; ok {
		return t
	}
	return key
}

// filterName is how the legends say a filter, or the key if unknown.
func filterName(key string) string { return textOf(tables.filters, key) }

// scarName is how the legends say a scar; boonName a boon.
func scarName(key string) string { return textOf(tables.scars, key) }
func boonName(key string) string { return textOf(tables.boons, key) }

// conditionText says a remain's description in its condition.
func conditionText(cond, desc string) string {
	return expand(textOf(tables.condText, cond), map[string]string{"desc": desc})
}

// traceText is how a trace kind reads.
func traceText(key string, vars map[string]string) string {
	return expand(textOf(tables.traces, key), vars)
}

// aptitudeText is the text of a birthright block, by its row's key.
func aptitudeText(key string) string { return textOf(tables.aptitudes, key) }

// rarityFrame is what a rarity of a key is called in a note.
func rarityFrame(key string) string {
	if s := tables.sources[key]; s != nil && s.Frame != "" {
		return s.Frame
	}
	return key
}

// levelName turns a level into words: the first band of data/levels.json
// it is below, the last for anything above.
func levelName(x float64) string {
	for _, b := range tables.bands {
		if b.Below == 0 || x < b.Below {
			return b.Word
		}
	}
	return ""
}

// stiffWord is the portrait's word for a people's stiffness: the last
// row for one ossified, else the row its stiffness begins at.
func stiffWord(c *record.Civ) string {
	rows := tables.stiffness
	if c.Ossified {
		return rows[len(rows)-1].Word
	}
	for i := len(rows) - 2; i > 0; i-- {
		if c.Stiff >= rows[i].From {
			return rows[i].Word
		}
	}
	return rows[0].Word
}

// wisdomWord is the portrait's word for a people's Wisdom, at the ends
// only: nothing between.
func wisdomWord(wis float64) string {
	switch {
	case wis >= 8:
		return "a wise people"
	case wis <= 2:
		return "a foolish people"
	}
	return ""
}

// expand fills {name} holes in a table's text from vars; a hole with no
// value is left as it is.
func expand(text string, vars map[string]string) string {
	out := []byte{}
	for i := 0; i < len(text); i++ {
		if text[i] == '{' {
			if j := indexByte(text[i:], '}'); j > 0 {
				if v, ok := vars[text[i+1:i+j]]; ok {
					out = append(out, v...)
					i += j
					continue
				}
			}
		}
		out = append(out, text[i])
	}
	return string(out)
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func hasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }
