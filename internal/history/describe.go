package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// The view's descriptions: what a remain, a source, a trace or an elder
// looks like in words, from the keys the record holds and the tables in
// data/. Nothing in the simulation reads these; they move out with the
// view at the writer step.

// ElderPortrait is an elder's portrait in words.
func (w *World) ElderPortrait(e *Elder) string { return portraitText("elders", e.Portrait) }

// AgeEnder is what swept up an age's remains, in words.
func (w *World) AgeEnder(a *AgeRecord) string { return portraitText("enders", a.Ender) }

// legacyDesc is a remain's bare description, without its condition: the
// portrait the record holds, filled with its star, its makers and its art.
func (w *World) legacyDesc(l *Legacy) string {
	vars := map[string]string{"T": w.star(l.Star)}
	if l.Maker >= 0 {
		vars["M"] = w.Civs[l.Maker].Tok()
	}
	if n := tech.Get(l.Node); n != nil {
		vars["node"] = n.Name
	}
	switch l.Kind {
	case Field:
		if l.Derelicts > 0 && l.Wrecks == 0 {
			return expand(portraitText("fields", "derelicts"), vars)
		}
		return expand(portraitText("fields", "wrecks"), vars)
	case Bounty:
		return portraitText("bounties", l.Portrait)
	case Threat:
		return expand(portraitText("threats", l.Portrait), vars)
	case Sleeper:
		return expand(portraitText("sleepers", l.Portrait), vars)
	case Law:
		return portraitText("laws", l.Portrait)
	case Structure:
		if l.Maker < 0 {
			return portraitText("structures", l.Portrait)
		}
		return expand(tech.Structures[l.Portrait].Remain, vars)
	case Artifact:
		if f := tables.miracleByKey[l.Node]; f != nil && f.Object != "" {
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

// Describe gives a remain's description with its condition; a field with
// the ships in it.
func (w *World) Describe(l *Legacy) string {
	return w.describeAt(l, l.Cond, l.ships(), l.Adrift)
}

// describeAt is Describe as the remain stood at some moment: the
// condition, and for a field the ships in it, as an event captured them.
func (w *World) describeAt(l *Legacy, cond Condition, ships int, adrift bool) string {
	desc := w.legacyDesc(l)
	if l.Maker < 0 {
		return desc
	}
	if l.Kind == Field {
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
func (w *World) sourceName(s *Source) string {
	key := s.Key
	if hasPrefix(key, "bounty:") {
		key = "bounty"
	}
	row := tables.sourceByKey[key]
	if row == nil {
		return s.Key
	}
	vars := map[string]string{}
	if s.Star >= 0 {
		vars["T"] = w.star(s.Star)
	}
	if s.Feature != nil {
		vars["F"] = s.Feature.Name
	}
	if s.Legacy >= 0 {
		vars["L"] = w.legacyDesc(w.Legacies[s.Legacy])
	}
	if s.Form != "" {
		// an object is its form, whatever remain it was found as
		if m := tables.miracleByKey[s.Key]; m != nil {
			for _, f := range m.Forms {
				if f.Key == s.Form {
					vars["L"] = f.Desc
				}
			}
		}
	}
	return expand(row.Name, vars)
}

// SourceName is sourceName for the view.
func (w *World) SourceName(s *Source) string { return w.sourceName(s) }

// traceName is a trace's kind in words.
func (w *World) traceName(t Trace) string {
	vars := map[string]string{}
	if t.Civ >= 0 {
		vars["C"] = w.Civs[t.Civ].Tok()
	}
	if t.Species >= 0 {
		f := w.Species[t.Species].Flavour()
		vars["colony"], vars["ship"] = f.Colony, f.Ship
	}
	return traceText(t.Kind, vars)
}

// TraceName is traceName for the view.
func (w *World) TraceName(t Trace) string { return w.traceName(t) }

func hasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }

// textOf is a keyed text from one of the causes tables, or the key.
func textOf(m map[string]*textDef, key string) string {
	if d := m[key]; d != nil {
		return d.Text
	}
	return key
}

// blastText is what a blast was, as a cause says it: from the blast
// event's way, its star and the remain that failed.
func (w *World) blastText(id int) string {
	b := w.Events[id]
	vars := map[string]string{"T": w.star(b.Star)}
	if b.Legacy >= 0 {
		vars["L"] = w.legacyDesc(w.Legacies[b.Legacy])
	}
	return expand(textOf(tables.blasts, b.P["way"].(string)), vars)
}

// whyText renders a reason a fact carries under name, with the parties
// beside it: by, at, plague and blast.
func (w *World) whyText(e *Event, name string) string {
	return w.reasonText(e, name, true)
}

// reasonText is whyText, composing a remnant's end with its fall only
// when asked, since CauseText composes it itself.
func (w *World) reasonText(e *Event, name string, compose bool) string {
	key, _ := e.P[name].(string)
	if key == "" {
		return ""
	}
	if compose && e.Kind == FEnd && name == "cause" && e.P["remnant"].(bool) {
		return w.CauseText(w.Civs[e.Subject]) // the fall's reason, and long after it the end's
	}
	vars := map[string]string{}
	by := -1
	if v, ok := e.P["by"].(int); ok && v >= 0 {
		by = v
		vars["C"] = w.Civs[v].Tok()
		other := e.Subject
		if other == v {
			other = e.Object
		}
		if other >= 0 {
			vars["O"] = w.Civs[other].Tok()
		}
	}
	if v, ok := e.P["at"].(int); ok && v >= 0 {
		vars["T"] = w.star(v)
	}
	if v, ok := e.P["plague"].(int); ok && v >= 0 {
		vars["P"] = w.Plagues[v].Tok()
	}
	if v, ok := e.P["blast"].(int); ok && v >= 0 {
		vars["B"] = w.blastText(v)
	} else if by >= 0 {
		vars["B"] = "the " + w.Civs[by].Tok()
	}
	if hasPrefix(key, "filter_") {
		vars["filter"] = filters[key[len("filter_"):]].Name
	}
	if tables.causes[key] == nil && tables.warCauses[key] != nil {
		return expand(textOf(tables.warCauses, key), vars) // a war's cause, on its declaration
	}
	return expand(textOf(tables.causes, key), vars)
}

// CauseText is why a people fell or ended, as the aftermath says it: the
// fall's reason, and long after it the end's, for a remnant that ended.
func (w *World) CauseText(c *Civ) string {
	if c.FallEvent < 0 && c.EndEvent < 0 {
		return textOf(tables.causes, c.Cause) // sundered or shattered: no fall, no end
	}
	out := ""
	if c.FallEvent >= 0 {
		out = w.reasonText(w.Events[c.FallEvent], "cause", false)
	}
	if c.EndEvent >= 0 {
		end := w.Events[c.EndEvent]
		s := w.reasonText(end, "cause", false)
		if q, _ := end.P["queen"].(bool); q {
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

// IntoText is what a transformed people became.
func (w *World) IntoText(c *Civ) string {
	if c.Into == "" {
		return ""
	}
	if c.Into == "species" {
		return expand(textOf(tables.becomings, c.Into), map[string]string{"C": speciesTok(w.Species[c.IntoCivs[0]])})
	}
	if len(c.IntoCivs) == 0 {
		return textOf(tables.becomings, c.Into)
	}
	var parts []string
	for _, id := range c.IntoCivs {
		parts = append(parts, expand(textOf(tables.becomings, c.Into), map[string]string{"C": w.Civs[id].Tok()}))
	}
	return join(parts, ", ")
}

// OriginText is how a people or a blood came to be, from its origin.
func (w *World) OriginText(o Origin) string {
	if o.Key == "" {
		return ""
	}
	vars := map[string]string{}
	if o.By >= 0 {
		vars["C"] = w.Civs[o.By].Tok()
	}
	if o.From >= 0 {
		vars["S"] = w.Civs[o.From].Tok()
	}
	if o.Legacy >= 0 {
		l := w.Legacies[o.Legacy]
		vars["L"] = w.legacyDesc(l)
		vars["T"] = w.star(l.Star)
		if l.Elder != nil {
			vars["E"] = w.ElderPortrait(l.Elder)
		}
	}
	if o.Plague >= 0 {
		vars["P"] = w.Plagues[o.Plague].Tok()
	}
	if d := tables.origins[o.Key]; d != nil {
		return expand(d.Text, vars)
	}
	return o.Key
}

// WarCause is why a war was declared; WarResult how it ended.
func (w *World) WarCause(wr *War) string {
	vars := map[string]string{}
	if wr.CauseOf >= 0 {
		vars["C"] = w.Civs[wr.CauseOf].Tok()
	}
	return expand(textOf(tables.warCauses, wr.Cause), vars)
}

func (w *World) WarResult(wr *War) string { return textOf(tables.warResults, wr.Result) }

// BetrayalText is the shape of a broken faith, said of its object.
func (w *World) BetrayalText(shape string, against int) string {
	return expand(textOf(tables.betrayals, shape), map[string]string{"C": w.Civs[against].Tok()})
}

// RecordText is one entry of a people's record in words.
func (w *World) RecordText(r Record) string {
	switch r.Kind {
	case "faced":
		s := r.Outcome.String() + " " + filters[r.Filter].Name
		if r.Narrow {
			s += " (narrowly)"
		}
		return s
	case "foresaw":
		return "foresaw " + filters[r.Filter].Name
	case "sky":
		return "took to the sky"
	case "rest":
		return "came to rest"
	}
	maker := w.makerNameIf(w.Legacies[r.Legacy], r.Known)
	switch r.Kind {
	case "bounty":
		return "put a bounty of " + maker + " to use"
	case "crewed":
		return "crewed the wrecks of " + maker
	}
	return r.Kind + " a legacy of " + maker
}

// useName is a use in words, from a went-dark event's key and the ids
// it captured: a node, a work at a star, a fleet or a ship in flight, a
// dock, a plague programme, or a word owed.
func (w *World) useName(e *Event) string {
	key := e.P["use"].(string)
	star := func() string { return w.star(e.P["star"].(int)) }
	switch {
	case hasPrefix(key, "work:"):
		rest := key[len("work:"):]
		i := indexByte(rest, ':')
		return "the " + tech.Structures[rest[:i]].Name + " at " + w.star(atoi(rest[i+1:]))
	case hasPrefix(key, "weapon:"):
		return "the programme that keeps " + w.Plagues[e.Plague].Tok()
	case hasPrefix(key, "fleet:"):
		what := map[string]string{"fleet": "the fleet", "surveyors": "the surveyors", "scout": "the scout", "picket": "the picket", "interceptors": "the interceptors", "guard": "the guard", "horde": "the horde's fleet"}[e.P["what"].(string)]
		if e.P["based"].(bool) {
			return what + " at " + star()
		}
		return what + " bound for " + star()
	case hasPrefix(key, "ship:"):
		return "the colony ship bound for " + star()
	case hasPrefix(key, "dock:"):
		if e.P["what"] == "grounds" {
			return "the breeding grounds at " + star()
		}
		return "the yards at " + star()
	case hasPrefix(key, "word:"):
		return "the " + flowWord(e.P["flow"].(flow.Kind)) + " owed to the " + w.Civs[e.P["to"].(int)].Tok()
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
func (w *World) driftText(e *Event) string {
	gained, _ := e.P["gained"].(string)
	lost, _ := e.P["lost"].(string)
	how := ""
	if e.P["world"].(bool) {
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

// PortraitLines is a people's portrait as its blood and the powers it
// held at the time say it: the substrate's sentence, the modifiers', the
// powers' and the channel's.
func (w *World) PortraitLines(sp *species.Species, powers []string) []string {
	return sp.PortraitWith(powers)
}

// scarName is how the legends say a scar; boonName a boon.
func scarName(key string) string {
	if s := tables.scarByKey[key]; s != nil {
		return s.Name
	}
	return key
}

func boonName(key string) string {
	if b := tables.boonByKey[key]; b != nil {
		return b.Name
	}
	return key
}

// ScarName and BoonName are the same for the view.
func ScarName(key string) string { return scarName(key) }
func BoonName(key string) string { return boonName(key) }

// portraitText is the text of a portrait key in a list, or the key.
func portraitText(list, key string) string {
	if p := tables.portraitByKey[list][key]; p != nil {
		return p.Text
	}
	return key
}

// ConditionName says a remain's description in its condition.
func conditionText(c Condition, desc string) string {
	return expand(tables.conditions[c].Text, map[string]string{"desc": desc})
}

// traceText is how a trace kind reads.
func traceText(key string, vars map[string]string) string {
	if t := tables.traceByKey[key]; t != nil {
		return expand(t.Text, vars)
	}
	return key
}

// termName is a term in a phrase.
func (w *World) termName(t Term) string {
	switch t.Kind {
	case mind.TermFlow:
		return flowWord(t.Res)
	case mind.TermRarity:
		return w.sourceName(w.Sources[t.Source])
	case mind.TermAccess:
		return "the use of " + w.sourceName(w.Sources[t.Source])
	case mind.TermTeach:
		return "the making of " + tech.Get(t.Node).Name
	case mind.TermGuard:
		return sprintf("%s to hold %s", shipsWord(int(t.Amount)), w.star(t.Star))
	case mind.TermStrike:
		if t.Work != "" {
			return sprintf("%s against the %s at %s", shipsWord(int(t.Amount)), tech.Structures[t.Work].Name, w.star(t.Star))
		}
		return sprintf("%s against %s", shipsWord(int(t.Amount)), w.star(t.Star))
	case mind.TermDeliver:
		return sprintf("%s to take %s", shipsWord(int(t.Amount)), w.star(t.Star))
	case mind.TermPeace:
		return "peace"
	case mind.TermSighting:
		return "word of a fleet in flight"
	case mind.TermBroker:
		return "a word with the " + w.Civs[t.Target].Tok()
	}
	return t.Kind.String()
}
