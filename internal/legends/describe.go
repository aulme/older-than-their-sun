package legends

import (
	"worldgen/internal/flow"
	"worldgen/internal/galaxy"
	"worldgen/internal/record"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// The view's descriptions: what a remain, a source, a trace or an elder
// looks like in words, from the keys the record holds and the tables in
// data/.

// elderPortrait is an elder's portrait in words.
func (v *view) elderPortrait(e *record.Elder) string { return portraitText("elders", e.Portrait) }

// ageEnder is what swept up an age's remains, in words.
func (v *view) ageEnder(a *record.Age) string { return portraitText("enders", a.Ender) }

// legacyDesc is a remain's bare description, without its condition: the
// portrait the record holds, filled with its star, its makers and its art.
func (v *view) legacyDesc(l *record.Remain) string {
	vars := map[string]string{"T": star(l.Star)}
	if l.Maker >= 0 {
		vars["M"] = tok(l.Maker)
	}
	if n := tech.Get(l.Node); n != nil {
		vars["node"] = n.Name
	}
	switch l.Kind {
	case "field":
		if l.Derelicts > 0 && l.Wrecks == 0 {
			return expand(portraitText("fields", "derelicts"), vars)
		}
		return expand(portraitText("fields", "wrecks"), vars)
	case "bounty":
		return portraitText("bounties", l.Portrait)
	case "threat":
		return expand(portraitText("threats", l.Portrait), vars)
	case "sleeper":
		return expand(portraitText("sleepers", l.Portrait), vars)
	case "law":
		return portraitText("laws", l.Portrait)
	case "structure":
		if l.Maker < 0 {
			return portraitText("structures", l.Portrait)
		}
		return expand(tech.Structures[l.Portrait].Remain, vars)
	case "artifact":
		if f := tables.miracles[l.Node]; f != nil && f.Object != "" {
			for _, fm := range f.Forms {
				if fm.Key == l.Portrait {
					return fm.Desc
				}
			}
		}
		if l.Maker < 0 {
			return portraitText("artifacts", l.Portrait)
		}
		return expand(portraitText("relics", l.Portrait), vars)
	}
	return l.Portrait
}

// describe gives a remain's description with its condition; a field with
// the ships in it.
func (v *view) describe(l *record.Remain) string {
	return v.describeAt(l, l.Cond, shipsIn(l), l.Adrift)
}

// describeAt is describe as the remain stood at some moment: the
// condition, and for a field the ships in it, as an event captured them.
func (v *view) describeAt(l *record.Remain, cond string, ships int, adrift bool) string {
	desc := v.legacyDesc(l)
	if l.Maker < 0 {
		return desc
	}
	if l.Kind == "field" {
		s := desc
		if ships > 0 {
			s += ", " + shipsWord(ships)
		} else {
			s = "what is left of " + desc
		}
		if adrift {
			s += ", adrift"
		}
		return s
	}
	return conditionText(cond, desc)
}

// sourceName is a source in words: the natural ones by their key and
// star or feature, a remain's by the remain.
func (v *view) sourceName(s *record.Source) string {
	key := s.Key
	if hasPrefix(key, "bounty:") {
		key = "bounty"
	}
	row := tables.sources[key]
	if row == nil {
		return s.Key
	}
	vars := map[string]string{}
	if s.Star >= 0 {
		vars["T"] = star(s.Star)
	}
	if s.Feature != "" {
		vars["F"] = featureName(s.Feature)
	}
	if s.Legacy >= 0 {
		vars["L"] = v.legacyDesc(v.remain(s.Legacy))
	}
	if s.Form != "" {
		// an object is its form, whatever remain it was found as
		if m := tables.miracles[s.Key]; m != nil {
			for _, f := range m.Forms {
				if f.Key == s.Form {
					vars["L"] = f.Desc
				}
			}
		}
	}
	return expand(row.Name, vars)
}

// traceName is a trace's kind in words.
func (v *view) traceName(t record.Trace) string {
	vars := map[string]string{}
	if t.Civ >= 0 {
		vars["C"] = tok(t.Civ)
	}
	if t.Species >= 0 {
		f := v.sp[t.Species].Flavour()
		vars["colony"], vars["ship"] = f.Colony, f.Ship
	}
	return traceText(t.Kind, vars)
}

// blastText is what a blast was, as a cause says it: from the blast
// event's way, its star and the remain that failed.
func (v *view) blastText(id int) string {
	b := v.event(id)
	vars := map[string]string{"T": star(b.Star)}
	if b.Legacy >= 0 {
		vars["L"] = v.legacyDesc(v.remain(b.Legacy))
	}
	return expand(textOf(tables.blasts, b.Str("way")), vars)
}

// whyText renders a reason a fact carries under name, with the parties
// beside it: by, at, plague and blast.
func (v *view) whyText(e *record.Event, name string) string {
	return v.reasonText(e, name, true)
}

// reasonText is whyText, composing a remnant's end with its fall only
// when asked, since causeText composes it itself.
func (v *view) reasonText(e *record.Event, name string, compose bool) string {
	key := e.Str(name)
	if key == "" {
		return ""
	}
	if compose && e.Kind == record.FEnd && name == "cause" && e.Bool("remnant") {
		return v.causeText(v.civ(e.Subject)) // the fall's reason, and long after it the end's
	}
	vars := map[string]string{}
	by := -1
	if e.Has("by") && e.Int("by") >= 0 {
		by = e.Int("by")
		vars["C"] = tok(by)
		other := e.Subject
		if other == by {
			other = e.Object
		}
		if other >= 0 {
			vars["O"] = tok(other)
		}
	}
	if e.Has("at") && e.Int("at") >= 0 {
		vars["T"] = star(e.Int("at"))
	}
	if e.Has("plague") && e.Int("plague") >= 0 {
		vars["P"] = plagueTok(e.Int("plague"))
	}
	if e.Has("blast") && e.Int("blast") >= 0 {
		vars["B"] = v.blastText(e.Int("blast"))
	} else if by >= 0 {
		vars["B"] = "the " + tok(by)
	}
	if hasPrefix(key, "filter_") {
		vars["filter"] = filterName(key[len("filter_"):])
	}
	if _, ok := tables.causes[key]; !ok {
		if _, war := tables.warCauses[key]; war {
			return expand(textOf(tables.warCauses, key), vars) // a war's cause, on its declaration
		}
	}
	return expand(textOf(tables.causes, key), vars)
}

// causeText is why a people fell or ended, as the aftermath says it: the
// fall's reason, and long after it the end's, for a remnant that ended.
func (v *view) causeText(c *record.Civ) string {
	if c.FallEvent < 0 && c.EndEvent < 0 {
		return textOf(tables.causes, c.Cause) // sundered or shattered: no fall, no end
	}
	out := ""
	if c.FallEvent >= 0 {
		out = v.reasonText(v.event(c.FallEvent), "cause", false)
	}
	if c.EndEvent >= 0 {
		end := v.event(c.EndEvent)
		s := v.reasonText(end, "cause", false)
		if end.Bool("queen") {
			s += ", and their queen with it"
		}
		if out != "" {
			out += ", and long after " + s
		} else {
			out = s
		}
	}
	return out
}

// intoText is what a transformed people became.
func (v *view) intoText(c *record.Civ) string {
	if c.Into == "" {
		return ""
	}
	if c.Into == "species" {
		return expand(textOf(tables.becomings, c.Into), map[string]string{"C": speciesTok(c.IntoCivs[0])})
	}
	if len(c.IntoCivs) == 0 {
		return textOf(tables.becomings, c.Into)
	}
	var parts []string
	for _, id := range c.IntoCivs {
		parts = append(parts, expand(textOf(tables.becomings, c.Into), map[string]string{"C": tok(id)}))
	}
	return join(parts, ", ")
}

// originText is how a people or a blood came to be, from its origin.
func (v *view) originText(o species.Making) string {
	if o.Key == "" {
		return ""
	}
	vars := map[string]string{}
	if o.By >= 0 {
		vars["C"] = tok(o.By)
	}
	if o.From >= 0 {
		vars["S"] = tok(o.From)
	}
	if o.Legacy >= 0 {
		l := v.remain(o.Legacy)
		vars["L"] = v.legacyDesc(l)
		vars["T"] = star(l.Star)
		if l.Elder >= 0 {
			vars["E"] = v.elderPortrait(v.elder(l.Elder))
		}
	}
	if o.Plague >= 0 {
		vars["P"] = plagueTok(o.Plague)
	}
	if t, ok := tables.origins[o.Key]; ok {
		return expand(t, vars)
	}
	return o.Key
}

// warCause is why a war was declared; warResult how it ended.
func (v *view) warCause(wr *record.War) string {
	vars := map[string]string{}
	if wr.CauseOf >= 0 {
		vars["C"] = tok(wr.CauseOf)
	}
	return expand(textOf(tables.warCauses, wr.Cause), vars)
}

func (v *view) warResult(wr *record.War) string { return textOf(tables.warResults, wr.Result) }

// betrayalText is the shape of a broken faith, said of its object.
func (v *view) betrayalText(shape string, against int) string {
	return expand(textOf(tables.betrayals, shape), map[string]string{"C": tok(against)})
}

// recordText is one entry of a people's record in words.
func (v *view) recordText(r record.Record) string {
	switch r.Kind {
	case "faced":
		s := map[string]string{"overcame": "overcame", "scarred": "scarred by", "declined": "fell to"}[r.Outcome] + " " + filterName(r.Filter)
		if r.Narrow {
			s += " (narrowly)"
		}
		return s
	case "foresaw":
		return "foresaw " + filterName(r.Filter)
	case "sky":
		return "took to the sky"
	case "rest":
		return "came to rest"
	}
	maker := v.makerNameIf(v.remain(r.Legacy), r.Known)
	switch r.Kind {
	case "bounty":
		return "put a bounty of " + maker + " to use"
	case "crewed":
		return "crewed the wrecks of " + maker
	}
	return r.Kind + " a legacy of " + maker
}

// makerNameIf is what a finder calls the makers, given whether it can
// place them.
func (v *view) makerNameIf(l *record.Remain, known bool) string {
	if l.Elder >= 0 {
		return elderTok(l.Elder)
	}
	if known {
		return "the " + tok(l.Maker)
	}
	return makersTok(l)
}

// useName is a use in words, from a went-dark event's key and the ids
// it captured: a node, a work at a star, a fleet or a ship in flight, a
// dock, a plague programme, or a word owed.
func (v *view) useName(e *record.Event) string {
	key := e.Str("use")
	at := func() string { return star(e.Int("star")) }
	switch {
	case hasPrefix(key, "work:"):
		rest := key[len("work:"):]
		i := indexByte(rest, ':')
		return "the " + tech.Structures[rest[:i]].Name + " at " + star(atoi(rest[i+1:]))
	case hasPrefix(key, "weapon:"):
		return "the programme that keeps " + plagueTok(e.Plague)
	case hasPrefix(key, "fleet:"):
		what := map[string]string{"fleet": "the fleet", "surveyors": "the surveyors", "scout": "the scout", "picket": "the picket", "interceptors": "the interceptors", "guard": "the guard", "horde": "the horde's fleet"}[e.Str("what")]
		if e.Bool("based") {
			return what + " at " + at()
		}
		return what + " bound for " + at()
	case hasPrefix(key, "ship:"):
		return "the colony ship bound for " + at()
	case hasPrefix(key, "dock:"):
		if e.Str("what") == "grounds" {
			return "the breeding grounds at " + at()
		}
		return "the yards at " + at()
	case hasPrefix(key, "word:"):
		return "the " + flowWord(flowKind(e.Str("flow"))) + " owed to the " + tok(e.Int("to"))
	}
	if n := tech.Get(key); n != nil {
		return n.Name
	}
	return key
}

func join(parts []string, sep string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}

// driftText is what a drift changed, in words.
func (v *view) driftText(e *record.Event) string {
	gained, lost := e.Str("gained"), e.Str("lost")
	how := ""
	if e.Bool("world") {
		how = ", as the worlds they hold made them"
	}
	switch {
	case gained != "" && lost != "":
		return "where they were " + species.Get(lost).Name + " they are " + species.Get(gained).Name + how
	case gained != "":
		return "they are " + species.Get(gained).Name + " now" + how
	}
	return "they are no longer " + species.Get(lost).Name
}

// termName is a term in a phrase.
func (v *view) termName(t record.Term) string {
	switch t.Kind {
	case "flow":
		return flowWord(flowKind(t.Res))
	case "rarity":
		return v.sourceName(v.source(t.Source))
	case "access":
		return "the use of " + v.sourceName(v.source(t.Source))
	case "teach":
		return "the making of " + tech.Get(t.Node).Name
	case "guard":
		return sprintf("%s to hold %s", shipsWord(int(t.Amount)), star(t.Star))
	case "strike":
		if t.Work != "" {
			return sprintf("%s against the %s at %s", shipsWord(int(t.Amount)), tech.Structures[t.Work].Name, star(t.Star))
		}
		return sprintf("%s against %s", shipsWord(int(t.Amount)), star(t.Star))
	case "deliver":
		return sprintf("%s to take %s", shipsWord(int(t.Amount)), star(t.Star))
	case "peace":
		return "peace"
	case "sighting":
		return "word of a fleet in flight"
	case "broker":
		return "a word with the " + tok(t.Target)
	}
	return t.Kind
}

// term reads a term from an event's parameters.
func (v *view) term(e *record.Event, name string) record.Term {
	var t record.Term
	e.Obj(name, &t)
	return t
}

// flowKind reads a commodity by its symbol.
func flowKind(sym string) flow.Kind {
	for _, k := range flow.Kinds {
		if k.Symbol() == sym {
			return k
		}
	}
	return flow.O
}

// flowWord is a commodity as a payment reads: grain, power, metal.
func flowWord(k flow.Kind) string {
	return [...]string{"grain", "power", "metal"}[k]
}

// categoryPhrase is a category of use by its name, as a line says it.
func categoryPhrase(name string) string {
	var c flow.Category
	if err := c.UnmarshalJSON([]byte(`"` + name + `"`)); err != nil {
		return name
	}
	return c.Phrase()
}

// featureName is a catalogued feature's name, by key.
func featureName(key string) string {
	for _, f := range galaxy.Features {
		if f.Key == key {
			return f.Name
		}
	}
	return key
}

// morality is a people's judgment as the record holds it.
func moralityWord(m record.Morality) string {
	if m.Kind == "fixation" {
		return "fixed on " + m.Object
	}
	return m.Kind
}

// moralityPortrait is the sentence the legends give a morality at birth.
func moralityPortrait(m record.Morality) string {
	switch m.Kind {
	case "amoral":
		return "They have no word for wrong."
	case "individual":
		return "They hold each life a good in itself."
	case "herd":
		return "For them the many are everything and the one is nothing."
	}
	switch m.Object {
	case "conquest":
		return "For them the only good is the taking of worlds."
	case "spawning":
		return "For them the only good is more of themselves."
	case "knowing":
		return "For them the only good is knowing."
	case "old things":
		return "For them the only good is what the old ones left."
	}
	return "For them the only good is holding what they have."
}
