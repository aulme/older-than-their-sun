package history

import (
	"sort"
	"strings"
	"unicode"

	"worldgen/internal/flow"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Telling a tale: the fact's plain line with the teller's names for the
// parties, then the slant, then the wear. The same fact told by two
// peoples is two different stories, and the same people tells it
// differently as the ages pass.

// templates are the plain lines. S and O are the parties, T the star, L
// the remain, X the word (what, below, by kind), N the count.
var templates = map[Kind]string{
	FArise:        "{S} arose on {T}.",
	FStars:        "{S} reached the stars.",
	FSettle:       "{S} settled {T}.",
	FZenith:       "{S} held {N} and feared no one.",
	FDarkAge:      "{S} {X}, and a dark age followed.",
	FFall:         "{S} {X}, and were a remnant after.",
	FEnd:          "{S} {X}.",
	FWar:          "{S} made war on {O}, over {X}.",
	FTaken:        "{S} took {T} from {O}.",
	FBurned:       "{S} burned {T}, a world of {O}.",
	FHomeBroken:   "{S} broke {T}, the home of {O}.",
	FScoured:      "{S} scoured {O} from {T}, and left none.",
	FYield:        "{O} yielded to {s}.",
	FPeace:        "{S} and {O} made peace.",
	FEnslaved:     "{S} took {O} and kept {o}.",
	FVassal:       "{O} bent the knee to {s}.",
	FFreed:        "{S} rose against {O} and were free.",
	FCrushed:      "{S} put down the rising of {O}.",
	FMet:          "{S} and {O} found each other.",
	FTrade:        "{S} and {O} traded across the dark.",
	FPact:         "{S} and {O} swore a pact of {X}.",
	FBetrayal:     "{S} {X}, and {O} paid for it.",
	FRelief:       "{S} stood with {O} at {T}.",
	FDefeat:       "{P} fleet was broken at {T} by {O}.",
	FIntercept:    "{S} met the fleet of {O} in the dark near {T} and beat it.",
	FCaught:       "{P} fleet was met in the dark near {T} by {O} and beaten.",
	FFind:         "{S} found {L} at {T}.",
	FMastered:     "{S} understood {L}, and how it was made.",
	FSealed:       "{S} sealed {L} at {T} and set a watch on it.",
	FUnleashed:    "{S} let {L} loose at {T}.",
	FOvercome:     "{S} faced {X} and came through.",
	FScarred:      "{S} faced {X} and were marked by it.",
	FDeclined:     "{X} broke {s}.",
	FMiracle:      "{S} gained {X}.",
	FUplift:       "{S} raised {O} from the beasts of {T}.",
	FBred:         "{S} remade {O} into something else.",
	FStarDied:     "{T} died, and the worlds of {S} with it.",
	FLeftStar:     "{S} left {T} to {X}.",
	FVacuumHole:   "{S} opened a hole in the vacuum at {T}, and the star went out.",
	FDoom:         "{P} sun began to fail.",
	FExodus:       "{S} left {T} and took to the sky.",
	FRest:         "{S} came to rest at {T}.",
	FStripped:     "{S} stripped {T} of its ships and its people, and {O} with it.",
	FCycle:        "{S} learned that the galaxy had done all this before, and would again.",
	FSurveyLost:   "{P} surveyors did not come back from {T}. Something is there.",
	FWant:         "{S} went without, and called them the lean years.",
	FHarness:      "{S} put {X} to use.",
	FEmbargo:      "{S} closed their ports to {O}.",
	FCutOff:       "{S} went dark when {O} stopped sending.",
	FManna:        "{S} ate what thought.",
	FRise:         "{S} were grown for the table of {O}, and rose.",
	FLoose:        "{S} let loose what they grew for the table, and it ate {T}.",
	FFathomed:     "{S} came to understand {O}.",
	FBrokered:     "{S} spoke for {O} to the {X}.",
	FHire:         "{S} took {X} from {O} to hold {T}.",
	FTaught:       "{S} taught {O} {X}.",
	FStrikeBought: "{S} paid the {X} to strike at {T}, a world of {O}.",
	FBoughtOff:    "{S}, paid by {O} to hold {T}, sold it to the {X}.",
	FTribute:      "{S} paid {O} in {X} for twenty thousand years.",
	FSlight:       "{S} made war on the {X}, with whom {O} traded.",
	FPlague:       "{X} came to {s}.",
	FPlagueGiven:  "{S} brought {X} to {o}.",
	FPlagueWorld:  "{X} emptied {T}, and {S} sealed it.",
	FCured:        "{S} were rid of {X}.",
	FRefused:      "{S} closed their doors to {O} for fear of {X}.",
	FBelieved:     "{T} went over to {X}, and was lost to {s}.",
	FWildfire:     "{X} was everywhere.",
	FPoisoned:     "{S} made {X} for {O} and hid it in what they sent.",
	FWoke:         "{X} began to think, and took {O}, and was {S}.",
	FRenaissance:  "{S} grew old, and then young again.",
	FSundered:     "{S} tore themselves apart, and {O} declared themselves {X}.",
	FReclaimed:    "{S} took {T} back from {O}, and called it restored to the realm.",
	FShattered:    "{S} forgot how to reach the stars, and on {T} {O} woke up alone.",
	FSevered:      "{T} was too far from the seat of {S} for one mind to hold, and what was there was {O} after.",
	FDeepened:     "{S} changed: {X} was in them after.",
	FAppeared:     "Another of {s} was at {T}, and nothing was seen to cross.",
	FTithed:       "{S} took a share of every harvest of {O}, and nobody agreed to it.",
	FDemand:       "{S} told {O} to leave {T}, and they {X}.",
	FWaking:       "{S} woke, and the worlds of {O} near {T} were unmade.",
	FUnmade:       "{S} unmade {T}, a world of {O}, without touching it.",
	FShipLost:     "{P} {X} to {T} was never heard from again.",
	FHunt:         "{P} ledger showed a hole around {T}, and {s} declared a hunt on it.",
	FDrifted:      "{S} changed again: {X}.",
	FWord:         "{S} reached into what lies beneath, and {X}.",
}

// blamedTemplates are the woes that name their own cause, retold once
// somebody else is blamed for them.
var blamedTemplates = map[Kind]string{
	FDarkAge:  "{S} were brought low, and a dark age followed.",
	FFall:     "{S} were brought low, and were a remnant after.",
	FEnd:      "{S} were ended.",
	FDeclined: "{S} were broken.",
	FScarred:  "{S} were marked, and it did not heal.",
	FWant:     "{S} were made to go without.",
	FCutOff:   "{S} were starved.",
}

// archetypes are what a people calls an enemy whose name it has lost.
var archetypes = []string{"the ones from the dark", "the eaters of worlds", "the faithless ones", "the old enemy", "the ones who came in ships"}

// Tell renders a tale in its teller's voice.
func (w *World) Tell(c *Civ, t *Tale) string { return w.tell(c, t) }

func (w *World) tell(c *Civ, t *Tale) string {
	f := w.Events[t.Fact]
	subj, obj := w.seen(c, f.Subject), w.seen(c, f.Object) // a party that cannot be held in mind has no name in the telling
	if t.Blamed >= 0 && f.sort() != Woe {
		subj = t.Blamed
	}
	line := templates[f.Kind]
	if t.Blamed >= 0 {
		if bl, ok := blamedTemplates[f.Kind]; ok {
			line = bl // the cause it gave itself is gone; the blame sentence carries it
		}
	}
	sort, _ := sortFor(c, f)
	if sort == Bond && obj == c.ID {
		subj, obj = obj, subj // we come first in what we did together
	}
	we := 0 // 1 subject, 2 object
	switch c.ID {
	case subj:
		we = 1
	case obj:
		we = 2
	}
	sl := t.Slant
	if f.Kind == FArise && we == 1 && t.Wear >= 2 {
		return "In the beginning we were on " + w.star(f.Star) + ", and there was nothing else."
	}
	what := w.what(f)
	if we == 1 {
		what = ours(what)
	}
	// the object of our own crime, at myth, is no longer a people
	oName := w.partyName(c, t, obj, false)
	if we == 1 && sort == Crime && t.Wear >= 2 && obj >= 0 {
		oName = "the ones who deserved it"
	}
	n := f.N
	switch {
	case we == 1 && sort == Deed, we == 2 && sort == Crime && sl < 0:
		n *= 1 + int(t.Wear)
	case we != 1 && sort == Deed && sl < 0:
		n = max(1, n/2)
	}
	sName := w.partyName(c, t, subj, true)
	poss := sName + "'s"
	switch {
	case we == 1:
		poss = "our"
	case subj < 0:
		poss = "its"
	}
	r := strings.NewReplacer(
		"{S}", sName,
		"{s}", w.partyName(c, t, subj, false),
		"{P}", poss,
		"{O}", oName,
		"{o}", pronoun(we == 2),
		"{T}", w.starName(c, t, f.Star),
		"{L}", w.remainName(f),
		"{X}", what,
		"{N}", systems(n),
	)
	s := r.Replace(line)
	s = w.frame(c, t, f, s, we, sort, sl)
	if we != 1 && sort != f.sort() {
		s += judged(f.sort(), sort)
	}
	return sentences(s)
}

// judged is what a teller adds when its judgment of another's act is not
// the fact's own.
func judged(was, is Sort) string {
	switch {
	case was == Crime && is == Deed:
		return " Among us that is counted a deed."
	case was == Crime && is == Nothing:
		return " We count it no crime."
	case was == Crime && is == Folly:
		return " It was a folly, no more."
	case was == Crime && is == Woe:
		return " We bore it."
	case is == Crime:
		return " We call it a crime."
	case is == Folly:
		return " Among us it is counted a folly."
	case is == Nothing:
		return " It is nothing to us."
	}
	return ""
}

func pronoun(us bool) string {
	if us {
		return "us"
	}
	return "them"
}

// what is the word in a fact, X in its template: a cause, a filter's
// name, a miracle, a shape of betrayal, from the event's parameters.
func (w *World) what(e *Event) string {
	switch e.Kind {
	case FWord:
		if w.Civs[e.Subject].Species.Voiceless() {
			return "had no word for it"
		}
		return "gave it a name"
	case FDarkAge, FFall, FEnd:
		return w.whyText(e, "cause")
	case FWar:
		return w.WarCause(w.Wars[e.P["war"].(int)])
	case FBetrayal:
		return w.BetrayalText(e.P["shape"].(string), e.Object)
	case FCutOff, FLeftStar, FShattered:
		return w.whyText(e, "why")
	case FPact:
		return e.P["pact"].(string)
	case FOvercome, FScarred, FDeclined:
		return filters[e.P["filter"].(string)].Name
	case FMiracle:
		return tables.miracleByKey[e.P["miracle"].(string)].Term
	case FHarness:
		if id, ok := e.P["source"]; ok {
			return w.sourceName(w.Sources[id.(int)])
		}
		return "the " + tech.Structures[e.P["work"].(string)].Name + " at " + w.star(e.Star)
	case FBrokered:
		return w.Civs[e.P["to"].(int)].Tok()
	case FHire:
		return w.termName(e.P["pay"].(Term))
	case FTaught:
		return tech.Get(e.P["node"].(string)).Name + ", for " + w.termName(e.P["pay"].(Term))
	case FStrikeBought:
		return w.Civs[e.P["seller"].(int)].Tok()
	case FBoughtOff:
		return w.Civs[e.P["buyer"].(int)].Tok()
	case FTribute:
		return flowWord(e.P["res"].(flow.Kind))
	case FSlight:
		return w.Civs[e.P["partner"].(int)].Tok()
	case FPlague, FPlagueGiven, FPlagueWorld, FCured, FRefused, FBelieved, FWildfire, FPoisoned, FWoke:
		return w.Plagues[e.Plague].Tok()
	case FSundered:
		return "the true " + w.Civs[e.Subject].Tok()
	case FDeepened:
		return species.PowerByKey(e.P["power"].(string)).Name
	case FDemand:
		return e.P["outcome"].(string)
	case FShipLost:
		return w.Species[e.P["species"].(int)].Flavour().Ship
	case FDrifted:
		return w.driftText(e)
	}
	return ""
}

// achievement says whether a deed is the kind an enemy would belittle.
func achievement(k Kind) bool {
	switch k {
	case FArise, FSettle, FRest, FVassal, FRelief:
		return false
	}
	return true
}

// ours turns a cause written of a people into one told by it.
func ours(s string) string {
	r := strings.NewReplacer(" themselves", " ourselves", " their ", " our ", " they ", " we ", " them", " us")
	return r.Replace(s)
}

// sentences capitalises the first letter of each sentence.
func sentences(s string) string {
	s = capitalise(s)
	for i := 0; i+2 < len(s); i++ {
		if s[i] == '.' && s[i+1] == ' ' && s[i+2] >= 'a' && s[i+2] <= 'z' {
			s = s[:i+2] + strings.ToUpper(s[i+2:i+3]) + s[i+3:]
		}
	}
	return s
}

// frame adds what the teller thinks of it.
func (w *World) frame(c *Civ, t *Tale, f *Event, s string, we int, sort Sort, sl int8) string {
	wear := t.Wear
	switch sort {
	case Deed:
		switch {
		case we == 1 && f.Kind == FArise:
		case we == 1 && wear == 1:
			s += " It was a great thing."
		case we == 1 && wear >= 2:
			s = "in the age of heroes, " + lower(s)
		case we == 0 && sl > 0 && wear >= 1:
			s += " They were good friends to us then."
		case we != 1 && sl < 0 && !achievement(f.Kind):
		case we != 1 && sl < 0 && wear == 0:
			s += " They had help."
		case we != 1 && sl < 0:
			s = "it is said that " + lower(s) + " Few believe it."
		}
	case Crime:
		switch {
		case we == 1 && wear == 0:
			s += " There was no other way."
		case we == 1 && wear == 1:
			s += " They had it coming."
		case we == 1:
			s += " It was necessary."
		case we == 2 && sl < 0 && wear == 0:
			s += " We had done nothing to deserve it."
		case we == 2 && sl < 0 && wear == 1:
			s += " It is not forgotten."
		case we == 2 && sl < 0:
			s += " Every child knows it."
		case we == 2 && sl > 0:
			s += " It was long ago, and they have made it right."
		case we == 2 && wear >= 1:
			s += " Nobody now remembers why."
		case sl < 0 && wear >= 1:
			s += " That is what they are."
		case sl > 0 && wear >= 1:
			s += " It was a hard time, and they had no choice."
		}
	case Woe:
		switch {
		case t.Blamed >= 0 && t.Blamed != c.ID:
			s += " It was the doing of " + w.partyName(c, t, t.Blamed, false) + "."
		case we == 1 && f.Object >= 0 && sl < 0 && wear == 0:
			s += " We had done nothing to deserve it."
		case we == 1 && f.Object >= 0 && sl < 0:
			s += " It is not forgotten."
		case we == 1 && f.Object < 0 && wear == 1:
			s += " Those were dark years."
		case we == 1 && f.Object < 0 && wear >= 2:
			s += " The old songs are about it."
		case we == 0 && sl < 0 && wear >= 1:
			s += " It was no more than they deserved."
		}
	case Bond:
		switch {
		case (we != 0) && sl < 0 && wear >= 1:
			s += " That was before we knew them."
		case (we != 0) && sl > 0 && wear >= 1:
			s += " It has held."
		}
	case Folly:
		switch {
		case t.Blamed >= 0 && t.Blamed != c.ID:
			s += " That is what they are."
		case we == 1 && wear == 0:
			s += " It was a mistake."
		case we == 1:
			s += " Nobody now knows why."
		case sl < 0:
			s += " That is what they are."
		case sl > 0 && wear >= 1:
			s += " It was a hard time."
		}
	}
	return s
}

// partyName is what the teller calls a people, by regard and by wear.
func (w *World) partyName(c *Civ, t *Tale, id int, subject bool) string {
	if id < 0 {
		f := w.Events[t.Fact]
		if was := f.Object; (subject && f.Subject >= 0) || (!subject && was >= 0) {
			return "something nameless" // a party the fact has and the teller cannot hold in mind
		}
		return "someone"
	}
	if id == c.ID {
		if subject {
			return "we"
		}
		return "us"
	}
	e := w.Civs[id]
	sl := t.Slant
	if t.Blamed == id {
		sl = -1
	}
	own := e.tokBy(c, regardTone(sl)) // the teller's own name for them, in the tale's regard
	name := "the " + own
	arch := archetypes[(t.Fact+id)%len(archetypes)]
	switch {
	case sl >= 1 && t.Wear == 0:
		return "our friends " + name
	case sl >= 1 && t.Wear == 1:
		return "our old friends " + name
	case sl >= 1:
		return name // the sentence says what they were to us
	case sl == -1 && t.Wear == 1:
		return "the faithless " + own
	case sl == -1 && t.Wear >= 2:
		return "the treacherous " + own
	case sl <= -2 && t.Wear == 0:
		return "the monstrous " + own
	case sl <= -2 && t.Wear == 1:
		return "the monsters of the " + own
	case sl <= -2:
		if c.Met[id] && e.Active() {
			return arch + " who call themselves the " + e.Tok()
		}
		return arch
	case sl == 0 && t.Wear == 1 && !c.Met[id]:
		return "a people called the " + own
	case sl == 0 && t.Wear >= 2 && !(c.Met[id] && e.Living()):
		return "a people whose name is lost"
	}
	return name
}

// regardTone is the tone of name a slant asks for: the friend row for
// a friend, the exonym for a stranger, the enemy or monster row below.
func regardTone(sl int8) string {
	switch {
	case sl >= 1:
		return "friend"
	case sl == -1:
		return "enemy"
	case sl <= -2:
		return "monster"
	}
	return "stranger"
}

// starName is what the teller calls a star: its own it never forgets.
func (w *World) starName(c *Civ, t *Tale, star int) string {
	if star < 0 {
		return "somewhere"
	}
	if t.Wear >= 2 && star != c.Home && star != c.Cradle && !contains(c.Systems, star) && w.Owner[star] != c.ID {
		return "a star whose name is lost"
	}
	return "{star:" + itoa(star) + "@" + itoa(c.ID) + "}"
}

func (w *World) remainName(f *Event) string {
	if f.Legacy < 0 && f.Plague >= 0 {
		return w.Plagues[f.Plague].Tok() // a plague that got out of the vial
	}
	if f.Legacy < 0 {
		return "something"
	}
	return w.legacyDesc(w.Legacies[f.Legacy])
}

// mythParty is the party mythOf names, if it names one: -1 when the
// fact has none or the line needs none.
func mythParty(f *Event) int {
	switch f.Kind {
	case FEnd, FFall, FBetrayal, FCutOff, FWaking:
		return f.Subject
	case FEnslaved, FFreed, FBred:
		return f.Object
	case FHomeBroken, FScoured, FUnleashed, FDarkAge, FWant, FSundered, FShattered, FSevered, FUnmade, FExodus:
		return -1
	}
	if f.Star >= 0 {
		return -1
	}
	return f.Object
}

// mythOf is a short name for a fact, for the chronicle's note that it has
// become a story; nameless says whether the party in it could be held in
// mind when the note was made.
func (w *World) mythOf(c *Civ, f *Event, nameless bool) string {
	name := func(id int) string {
		switch {
		case id < 0:
			return "someone"
		case nameless:
			return "something nameless"
		}
		return "the " + w.Civs[id].Tok()
	}
	switch f.Kind {
	case FEnd, FFall:
		return "the fall of " + name(f.Subject)
	case FHomeBroken, FScoured:
		return "the breaking of " + w.star(f.Star)
	case FEnslaved:
		return "the taking of " + name(f.Object)
	case FFreed:
		return "the rising against " + name(f.Object)
	case FBetrayal:
		return "the betrayal by " + name(f.Subject)
	case FUnleashed:
		return "what was let loose at " + w.star(f.Star)
	case FDarkAge:
		return "the dark age"
	case FWant:
		return "the lean years"
	case FCutOff:
		return "the starving of " + name(f.Subject)
	case FSundered:
		return "the sundering"
	case FShattered:
		return "the shattering"
	case FSevered:
		return "the cutting off of " + w.star(f.Star)
	case FWaking:
		return "the waking of " + name(f.Subject)
	case FUnmade:
		return "the unmaking of " + w.star(f.Star)
	case FExodus:
		return "the leaving of " + w.star(f.Star)
	case FBred:
		return "the remaking of " + name(f.Object)
	}
	if f.Star >= 0 {
		return "what happened at " + w.star(f.Star)
	}
	return "the matter of " + name(f.Object)
}

// deedOf is a fact as a verb phrase, for the note that blame has moved.
// blameOf is a wrong as a teller hangs it on somebody: "who <did this>".
func (w *World) blameOf(c *Civ, f *Event) string {
	if f.sort() != Woe {
		return w.deedOf(f)
	}
	us := "the " + w.Civs[f.Subject].Tok()
	if f.Subject == c.ID {
		us = "us"
	}
	switch f.Kind {
	case FDarkAge:
		return "brought the dark years on " + us
	case FWant:
		return "brought the lean years on " + us
	case FCutOff:
		return "starved " + us
	case FFall, FDeclined, FScarred:
		return "brought " + us + " low"
	case FEnd:
		return "ended the " + w.Civs[f.Subject].Tok()
	case FSundered:
		return "tore " + us + " apart"
	case FShattered:
		return "took the stars from " + us
	case FSevered:
		return "cut " + w.star(f.Star) + " from " + us
	case FSurveyLost:
		return "took the surveyors at " + w.star(f.Star)
	case FShipLost:
		return "took the " + w.Species[f.P["species"].(int)].Flavour().Ship + " at " + w.star(f.Star)
	case FDefeat:
		return "broke the fleet at " + w.star(f.Star)
	case FCaught:
		return "caught the fleet in the dark near " + w.star(f.Star)
	case FStarDied, FVacuumHole, FDoom:
		return "killed the sun"
	case FExodus:
		return "drove " + us + " from " + w.star(f.Star)
	}
	return "did it"
}

func (w *World) deedOf(f *Event) string {
	switch f.Kind {
	case FBurned:
		return "burned " + w.star(f.Star)
	case FTaken:
		return "took " + w.star(f.Star)
	case FScoured, FHomeBroken:
		return "broke " + w.star(f.Star)
	case FEnslaved:
		return "took the " + w.Civs[f.Object].Tok()
	case FBetrayal:
		return w.BetrayalText(f.P["shape"].(string), f.Object)
	case FStripped:
		return "stripped " + w.star(f.Star)
	case FUnleashed:
		return "let it loose"
	case FWar:
		return "made war on the " + w.Civs[f.Object].Tok()
	case FEmbargo:
		return "closed their ports to the " + w.Civs[f.Object].Tok()
	case FBred:
		return "remade the " + w.Civs[f.Object].Tok()
	case FManna:
		return "ate what thought"
	case FTithed:
		return "took a share of every harvest of the " + w.Civs[f.Object].Tok()
	case FWaking:
		return "woke on the worlds of the " + w.Civs[f.Object].Tok()
	case FUnmade:
		return "unmade " + w.star(f.Star)
	}
	return "did it"
}

// Telling is a people's tales as it would tell them: the ones it holds
// dearest, up to a limit, in the order it believes they happened.
func (w *World) Telling(c *Civ, limit int) []*Tale {
	var out []*Tale
	for _, t := range c.Lore {
		if !t.Forgot {
			out = append(out, t)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		fi, fj := w.Events[out[i].Fact], w.Events[out[j].Fact]
		di, dj := w.dearness(c, fi, out[i]), w.dearness(c, fj, out[j])
		if di != dj {
			return di > dj
		}
		return fi.Year > fj.Year
	})
	if len(out) > limit {
		out = out[:limit]
	}
	sort.SliceStable(out, func(i, j int) bool { return w.Events[out[i].Fact].Year < w.Events[out[j].Fact].Year })
	return out
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func lower(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	// keep a name's capital: only "The", "We", "A", "It", "In" go down
	if strings.HasPrefix(s, "The ") || strings.HasPrefix(s, "We ") || strings.HasPrefix(s, "A ") || strings.HasPrefix(s, "It ") || strings.HasPrefix(s, "In ") || strings.HasPrefix(s, "Our ") {
		r[0] = unicode.ToLower(r[0])
	}
	return string(r)
}
