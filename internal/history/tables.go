package history

import (
	"encoding/json"

	"worldgen/data"
	"worldgen/internal/flow"
)

// The lookup tables the history reads its mechanics from: data/*.json,
// loaded once into tables before any init runs. The text beside the
// mechanics (a scar's name, a portrait's sentence, a condition's words)
// is the view's and the names pass's; nothing here reads it.

// Table rows.
type (
	scarDef struct {
		Key   string `json:"key"`
		Name  string `json:"name"`
		Desc  string `json:"desc"`
		Dials Dials  `json:"dials"`
		Iron  bool   `json:"iron"`
	}
	boonDef struct {
		Key  string `json:"key"`
		Name string `json:"name"`
		Desc string `json:"desc"`
	}
	filterDef struct {
		Key      string       `json:"key"`
		Name     string       `json:"name"`
		Levels   []string     `json:"levels"`
		Diff     float64      `json:"diff"`
		Repeat   bool         `json:"repeat"`
		Domain   string       `json:"domain"`
		Wreckage *wreckageDef `json:"wreckage"`
		What     string       `json:"what"`
		Overcome string       `json:"overcome"`
		Scarred  string       `json:"scarred"`
		Declined string       `json:"declined"`
	}
	wreckageDef struct {
		Destroy float64 `json:"destroy"`
		Leave   string  `json:"leave"`
	}
	portrait struct {
		Key     string      `json:"key"`
		Text    string      `json:"text"`
		Variant string      `json:"variant"` // laws: what the law does
		Hardy   float64     `json:"hardy"`   // relics
		Yield   flow.Income `json:"yield"`   // bounties
		Levels  [3]float64  `json:"levels"`
		Wear    Year        `json:"wear"`
	}
	conditionDef struct {
		Key  string  `json:"key"`
		Text string  `json:"text"`
		Find float64 `json:"find"`
	}
	traceDef struct {
		Key  string `json:"key"`
		Text string `json:"text"`
	}
	lossDef struct {
		Key     string  `json:"key"`
		Destroy float64 `json:"destroy"`
		Leave   string  `json:"leave"`
	}
	miracleDef struct {
		Key    string    `json:"key"`
		Term   string    `json:"term"`
		Desc   string    `json:"desc"`
		Causal bool      `json:"causal"`
		Wear   float64   `json:"wear"`
		Object string    `json:"object"`
		Forms  []formDef `json:"forms"`
	}
	formDef struct {
		Key    string      `json:"key"`
		Desc   string      `json:"desc"`
		Weight float64     `json:"weight"`
		Yield  flow.Income `json:"yield"`
		Levels [3]float64  `json:"levels"`
	}
	routeDef struct {
		Key  string `json:"key"`
		Desc string `json:"desc"`
	}
	band struct {
		Word  string  `json:"word"`
		Below float64 `json:"below"`
	}
	stiffDef struct {
		Key  string  `json:"key"`
		From float64 `json:"from"`
		Word string  `json:"word"`
	}
	wallDef struct {
		Stage int     `json:"stage"`
		From  float64 `json:"from"`
		Word  string  `json:"word"`
	}
	sourceDef struct {
		Key     string `json:"key"`
		Name    string `json:"name"`
		Kind    string `json:"kind"`
		Frame   string `json:"frame"`
		Harness bool   `json:"harness"` // its first harnessing is a deed
		Remark  bool   `json:"remark"`  // its first holding is remarked
		Desc    string `json:"desc"`
	}
	textDef struct {
		Key  string `json:"key"`
		Text string `json:"text"`
		Desc string `json:"desc"`
	}
	originDef struct {
		Key     string `json:"key"`
		Text    string `json:"text"`
		Desc    string `json:"desc"`
		Parties string `json:"parties"`
	}
)

// tableSet is every table the history reads, by file.
type tableSet struct {
	scars     []scarDef
	boons     []boonDef
	filters   map[string]*filterDef
	traitDiff map[string]map[string]float64
	portraits struct {
		Elders, Enders, Knowers, Structures, Artifacts, Laws, Relics, Bounties, Threats, Sleepers, Fields []portrait
	}
	conditions []conditionDef
	traces     []traceDef
	losses     map[string]Wreckage
	defaultWr  Wreckage
	miracles   []miracleDef
	routes     []routeDef
	bands      []band
	stiffness  []stiffDef
	wall       []wallDef
	sources    []sourceDef
	causes     map[string]*textDef
	blasts     map[string]*textDef
	becomings  map[string]*textDef
	warCauses  map[string]*textDef
	warResults map[string]*textDef
	betrayals  map[string]*textDef
	origins    map[string]*originDef
	leaders    map[string]map[string]bool // data/leaders.json's keys by section: forms, occasions, ends, places

	scarByKey     map[string]*scarDef
	boonByKey     map[string]*boonDef
	miracleByKey  map[string]*miracleDef
	traceByKey    map[string]*traceDef
	portraitByKey map[string]map[string]*portrait
	sourceByKey   map[string]*sourceDef
}

var tables = loadTables()

func loadTables() *tableSet {
	t := &tableSet{filters: map[string]*filterDef{}, scarByKey: map[string]*scarDef{}, boonByKey: map[string]*boonDef{}, miracleByKey: map[string]*miracleDef{}, traceByKey: map[string]*traceDef{}, portraitByKey: map[string]map[string]*portrait{}, sourceByKey: map[string]*sourceDef{}}
	var sf struct {
		Scars []scarDef `json:"scars"`
	}
	data.Load("scars.json", &sf)
	t.scars = sf.Scars
	for i := range t.scars {
		t.scarByKey[t.scars[i].Key] = &t.scars[i]
	}
	var bf struct {
		Boons []boonDef `json:"boons"`
	}
	data.Load("boons.json", &bf)
	t.boons = bf.Boons
	for i := range t.boons {
		t.boonByKey[t.boons[i].Key] = &t.boons[i]
	}
	var ff struct {
		Filters []*filterDef                  `json:"filters"`
		Traits  map[string]map[string]float64 `json:"traits"`
	}
	data.Load("filters.json", &ff)
	for _, f := range ff.Filters {
		t.filters[f.Key] = f
	}
	t.traitDiff = ff.Traits
	var pf struct {
		Elders, Enders, Knowers, Structures, Artifacts, Laws, Relics, Bounties, Threats, Sleepers, Fields []portrait
	}
	data.Load("portraits.json", &pf)
	t.portraits = pf
	for name, list := range map[string][]portrait{"elders": pf.Elders, "enders": pf.Enders, "knowers": pf.Knowers, "structures": pf.Structures, "artifacts": pf.Artifacts, "laws": pf.Laws, "relics": pf.Relics, "bounties": pf.Bounties, "threats": pf.Threats, "sleepers": pf.Sleepers, "fields": pf.Fields} {
		m := map[string]*portrait{}
		for i := range list {
			m[list[i].Key] = &list[i]
		}
		t.portraitByKey[name] = m
	}
	var cf struct {
		Conditions []conditionDef `json:"conditions"`
		Traces     []traceDef     `json:"traces"`
		Losses     []lossDef      `json:"losses"`
		Default    wreckageDef    `json:"default"`
	}
	data.Load("conditions.json", &cf)
	t.conditions = cf.Conditions
	for i, c := range t.conditions {
		if c.Key != Condition(i).String() {
			panic("history: the file's condition " + c.Key + " is not the code's " + Condition(i).String())
		}
	}
	t.traces = cf.Traces
	for i := range t.traces {
		t.traceByKey[t.traces[i].Key] = &t.traces[i]
	}
	t.losses = map[string]Wreckage{}
	for _, l := range cf.Losses {
		t.losses[l.Key] = Wreckage{l.Destroy, conditionOf(l.Leave)}
	}
	t.defaultWr = Wreckage{cf.Default.Destroy, conditionOf(cf.Default.Leave)}
	var mf struct {
		Miracles []miracleDef `json:"miracles"`
		Routes   []routeDef   `json:"routes"`
	}
	data.Load("miracles.json", &mf)
	t.miracles, t.routes = mf.Miracles, mf.Routes
	for i := range t.miracles {
		t.miracleByKey[t.miracles[i].Key] = &t.miracles[i]
	}
	var lf struct {
		Bands     []band     `json:"bands"`
		Stiffness []stiffDef `json:"stiffness"`
		Wall      []wallDef  `json:"wall"`
	}
	data.Load("levels.json", &lf)
	t.bands, t.stiffness, t.wall = lf.Bands, lf.Stiffness, lf.Wall
	var srf struct {
		Sources []sourceDef `json:"sources"`
	}
	data.Load("sources.json", &srf)
	t.sources = srf.Sources
	for i := range t.sources {
		t.sourceByKey[t.sources[i].Key] = &t.sources[i]
	}
	var cf2 struct {
		Causes     []textDef `json:"causes"`
		Blasts     []textDef `json:"blasts"`
		Becomings  []textDef `json:"becomings"`
		WarCauses  []textDef `json:"war_causes"`
		WarResults []textDef `json:"war_results"`
		Betrayals  []textDef `json:"betrayals"`
	}
	data.Load("causes.json", &cf2)
	index := func(rows []textDef) map[string]*textDef {
		m := map[string]*textDef{}
		for i := range rows {
			m[rows[i].Key] = &rows[i]
		}
		return m
	}
	t.causes, t.blasts, t.becomings = index(cf2.Causes), index(cf2.Blasts), index(cf2.Becomings)
	t.warCauses, t.warResults, t.betrayals = index(cf2.WarCauses), index(cf2.WarResults), index(cf2.Betrayals)
	var of struct {
		Origins []originDef `json:"origins"`
	}
	data.Load("origins.json", &of)
	t.origins = map[string]*originDef{}
	for i := range of.Origins {
		t.origins[of.Origins[i].Key] = &of.Origins[i]
	}
	var ldf map[string]json.RawMessage
	data.Load("leaders.json", &ldf)
	t.leaders = map[string]map[string]bool{}
	for sec, raw := range ldf {
		var rows []struct {
			Key string `json:"key"`
		}
		if sec == "_" || json.Unmarshal(raw, &rows) != nil {
			continue
		}
		t.leaders[sec] = map[string]bool{}
		for _, r := range rows {
			t.leaders[sec][r.Key] = true
		}
	}
	return t
}

// conditionOf reads a condition by its key.
func conditionOf(key string) Condition {
	for c := Abandoned; c <= Ruin; c++ {
		if c.String() == key {
			return c
		}
	}
	panic("history: unknown condition " + key)
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
