package legends

import (
	"strings"
	"unicode"

	"worldgen/internal/record"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Telling a tale: the fact's plain line with the teller's names for the
// parties, then the slant, then the wear. The same fact told by two
// peoples is two different stories, and the same people tells it
// differently as the ages pass.

// templates are the plain lines. S and O are the parties, T the star, L
// the remain, X the word (what, below, by kind), N the count. A letter's
// case is the party's grammatical case where the teller is one of them
// and renders as a pronoun: {S} "we" against {s} "us", {OS} "we" against
// {O} "us". A named people reads the same either way.
var templates = map[record.Kind]string{
	record.FArise:        "{S} arose on {T}.",
	record.FStars:        "{S} reached the stars.",
	record.FSettle:       "{S} settled {T}.",
	record.FZenith:       "{S} held {N} and feared no one.",
	record.FDarkAge:      "{S} {X}, and a dark age followed.",
	record.FFall:         "{S} {X}, and were a remnant after.",
	record.FEnd:          "{S} {X}.",
	record.FWar:          "{S} made war on {O}, over {X}.",
	record.FTaken:        "{S} took {T} from {O}.",
	record.FBurned:       "{S} burned {T}, a world of {O}.",
	record.FHomeBroken:   "{S} broke {T}, the home of {O}.",
	record.FScoured:      "{S} scoured {O} from {T}, and left none.",
	record.FYield:        "{OS} yielded to {s}.",
	record.FPeace:        "{S} and {O} made peace.",
	record.FSettled:      "{S} sued {O} for peace, and had it.",
	record.FEnslaved:     "{S} took {O} and kept {o}.",
	record.FVassal:       "{OS} bent the knee to {s}.",
	record.FFreed:        "{S} rose against {O} and were free.",
	record.FCrushed:      "{S} put down the rising of {O}.",
	record.FMet:          "{S} and {O} found each other.",
	record.FTrade:        "{S} and {O} traded across the dark.",
	record.FPact:         "{S} and {O} swore a pact of {X}.",
	record.FBetrayal:     "{S} {X}, and {OS} paid for it.",
	record.FRelief:       "{S} stood with {O} at {T}.",
	record.FDefeat:       "{P} fleet was broken at {T} by {O}.",
	record.FIntercept:    "{S} met the fleet of {O} in the dark near {T} and beat it.",
	record.FCaught:       "{P} fleet was met in the dark near {T} by {O} and beaten.",
	record.FFind:         "{S} found {L} at {T}.",
	record.FMastered:     "{S} understood {L}, and how it was made.",
	record.FSealed:       "{S} sealed {L} at {T} and set a watch on it.",
	record.FUnleashed:    "{S} let {L} loose at {T}.",
	record.FOvercome:     "{S} faced {X} and came through.",
	record.FScarred:      "{S} faced {X} and were marked by it.",
	record.FDeclined:     "{X} broke {s}.",
	record.FMiracle:      "{S} gained {X}.",
	record.FUplift:       "{S} raised {O} from the beasts of {T}.",
	record.FBred:         "{S} remade {O} into something else.",
	record.FStarDied:     "{T} died, and the worlds of {S} with it.",
	record.FLeftStar:     "{S} left {T} to {X}.",
	record.FVacuumHole:   "{S} opened a hole in the vacuum at {T}, and the star went out.",
	record.FDoom:         "{P} sun began to fail.",
	record.FExodus:       "{S} left {T} and took to the sky.",
	record.FRest:         "{S} came to rest at {T}.",
	record.FStripped:     "{S} stripped {T} of its ships and its people, and {O} with it.",
	record.FCycle:        "{S} learned that the galaxy had done all this before, and would again.",
	record.FSurveyLost:   "{P} surveyors did not come back from {T}. Something is there.",
	record.FWant:         "{S} went without, and called them the lean years.",
	record.FHarness:      "{S} put {X} to use.",
	record.FEmbargo:      "{S} closed their ports to {O}.",
	record.FCutOff:       "{S} went dark when {OS} stopped sending.",
	record.FManna:        "{S} ate what thought.",
	record.FRise:         "{S} were grown for the table of {O}, and rose.",
	record.FLoose:        "{S} let loose what they grew for the table, and it ate {T}.",
	record.FFathomed:     "{S} came to understand {O}.",
	record.FBrokered:     "{S} spoke for {O} to the {X}.",
	record.FHire:         "{S} took {X} from {O} to hold {T}.",
	record.FTaught:       "{S} taught {O} {X}.",
	record.FStrikeBought: "{S} paid the {X} to strike at {T}, a world of {O}.",
	record.FBoughtOff:    "{S}, paid by {O} to hold {T}, sold it to the {X}.",
	record.FTribute:      "{S} paid {O} in {X} for twenty thousand years.",
	record.FSlight:       "{S} made war on the {X}, with whom {O} traded.",
	record.FPlague:       "{X} came to {s}.",
	record.FPlagueGiven:  "{S} brought {X} to {o}.",
	record.FPlagueWorld:  "{X} emptied {T}, and {S} sealed it.",
	record.FCured:        "{S} were rid of {X}.",
	record.FRefused:      "{S} closed their doors to {O} for fear of {X}.",
	record.FBelieved:     "{T} went over to {X}, and was lost to {s}.",
	record.FWildfire:     "{X} was everywhere.",
	record.FPoisoned:     "{S} made {X} for {O} and hid it in what they sent.",
	record.FWoke:         "{X} began to think, and took {O}, and was {s}.",
	record.FRenaissance:  "{S} grew old, and then young again.",
	record.FSundered:     "{S} tore {RS} apart, and {OS} declared {RO} {X}.",
	record.FReclaimed:    "{S} took {T} back from {O}, and called it restored to the realm.",
	record.FShattered:    "{S} forgot how to reach the stars, and on {T} {OS} woke up alone.",
	record.FSevered:      "{T} was too far from the seat of {S} for one mind to hold, and what was there was {O} after.",
	record.FDeepened:     "{S} changed: {X} was in them after.",
	record.FAppeared:     "Another of {s} was at {T}, and nothing was seen to cross.",
	record.FTithed:       "{S} took a share of every harvest of {O}, and nobody agreed to it.",
	record.FDemand:       "{S} told {O} to leave {T}, and they {X}.",
	record.FWaking:       "{S} woke, and the worlds of {O} near {T} were unmade.",
	record.FUnmade:       "{S} unmade {T}, a world of {O}, without touching it.",
	record.FShipLost:     "{P} {X} to {T} was never heard from again.",
	record.FHunt:         "{P} ledger showed a hole around {T}, and {s} declared a hunt on it.",
	record.FDrifted:      "{S} changed again: {X}.",
	record.FWord:         "{S} reached into what lies beneath, and {X}.",
	record.FLeader:       "{X} rose among {s}.",
	record.FLeaderLost:   "{S} lost {X} at {T}.",
}

// blamedTemplates are the woes that name their own cause, retold once
// somebody else is blamed for them.
var blamedTemplates = map[record.Kind]string{
	record.FDarkAge:  "{S} were brought low, and a dark age followed.",
	record.FFall:     "{S} were brought low, and were a remnant after.",
	record.FEnd:      "{S} were ended.",
	record.FDeclined: "{S} were broken.",
	record.FScarred:  "{S} were marked, and it did not heal.",
	record.FWant:     "{S} were made to go without.",
	record.FCutOff:   "{S} were starved.",
}

// archetypes are what a people calls an enemy whose name it has lost.
var archetypes = []string{"the ones from the dark", "the eaters of worlds", "the faithless ones", "the old enemy", "the ones who came in ships"}

// tell renders a tale in its teller's voice. A testament (a tale with
// a Frozen record) is told as its maker told it when it wrote: the
// frozen record stands in for what the telling would read of the world.
func (v *view) tell(c *record.Civ, t *record.Tale) string {
	f := v.event(t.Fact)
	k := v.knowing(c, t)
	subj, obj := k.seen(f.Subject), k.seen(f.Object) // a party that cannot be held in mind has no name in the telling
	fsort := sortOf(f)
	if t.Blamed >= 0 && fsort != "woe" {
		subj = t.Blamed
	}
	line := templates[f.Kind]
	if t.Blamed >= 0 {
		if bl, ok := blamedTemplates[f.Kind]; ok {
			line = bl // the cause it gave itself is gone; the blame sentence carries it
		}
	}
	sort := t.Sort
	if sort == "bond" && obj == c.ID {
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
	if f.Kind == record.FArise && we == 1 && t.Wear >= 2 {
		return "In the beginning we were on " + star(f.Star) + ", and there was nothing else."
	}
	what := v.what(f)
	if t.Wear >= 2 && f.Has("leader") {
		// at myth the name is gone and the figure is an archetype: the deed survives the name
		what = v.leaderArchetype(f)
	}
	if we == 1 {
		what = ours(what)
	}
	// the object of our own crime, at myth, is no longer a people
	oName := v.partyName(c, t, k, obj, false)
	if we == 1 && sort == "crime" && t.Wear >= 2 && obj >= 0 {
		oName = "the ones who deserved it"
	}
	n := f.N
	switch {
	case we == 1 && sort == "deed", we == 2 && sort == "crime" && sl < 0:
		n *= 1 + int(t.Wear)
	case we != 1 && sort == "deed" && sl < 0:
		n = max(1, n/2)
	}
	sName := v.partyName(c, t, k, subj, true)
	poss := sName + "'s"
	switch {
	case we == 1:
		poss = "our"
	case subj < 0:
		poss = "its"
	}
	r := strings.NewReplacer(
		"{S}", sName,
		"{s}", v.partyName(c, t, k, subj, false),
		"{P}", poss,
		// {OS} before {O}: the replacer matches in argument order, and
		// {O} is a prefix of {OS}. {O} is the object party where the
		// clause makes it an object ("took us"), {OS} where the clause
		// makes it the subject ("we paid for it"); the two differ only
		// when that party is the teller, since a name does not inflect.
		"{OS}", v.partyName(c, t, k, obj, true),
		"{O}", oName,
		// the reflexives, which follow whichever party is the teller
		"{RS}", reflexive(we == 1),
		"{RO}", reflexive(we == 2),
		"{o}", pronoun(we == 2),
		"{T}", v.starName(c, t, k, f.Star),
		"{L}", v.remainName(f),
		"{X}", what,
		"{N}", systems(n),
	)
	s := r.Replace(line)
	s = v.frame(c, t, k, f, s, we, sort, sl)
	if we != 1 && sort != fsort {
		s += judged(fsort, sort)
	}
	if strings.HasPrefix(s, "{") && !strings.HasPrefix(s, "{^") {
		s = "{^" + s[1:] // a name opens the sentence: the names pass raises it
	}
	return sentences(s)
}

// knowing is what a telling reads of the world besides the tale: for a
// living memory, the teller's knowledge and the state now; for a
// testament, the frozen record.
type knowing struct {
	perceived func(id int) bool
	met       func(id int) bool
	active    func(id int) bool
	living    func(id int) bool
	ours      func(star int) bool
}

func (k *knowing) seen(id int) int {
	if id < 0 || k.perceived(id) {
		return id
	}
	return -1
}

func (v *view) knowing(c *record.Civ, t *record.Tale) *knowing {
	if fz := t.Frozen; fz != nil {
		return &knowing{
			perceived: func(id int) bool { return has(fz.Perceived, id) },
			met:       func(id int) bool { return has(fz.Met, id) },
			active:    func(id int) bool { return has(fz.Active, id) },
			living:    func(id int) bool { return has(fz.Living, id) },
			ours:      func(int) bool { return fz.Ours },
		}
	}
	return &knowing{
		perceived: func(id int) bool { return v.perceives(c, id) },
		met:       func(id int) bool { return has(c.Knowledge.Met, id) },
		active:    func(id int) bool { return active(v.civ(id)) },
		living:    func(id int) bool { return living(v.civ(id)) },
		ours:      func(s int) bool { return s == c.Home || s == c.Cradle || has(c.Systems, s) || v.held(s) == c.ID },
	}
}

// judged is what a teller adds when its judgment of another's act is not
// the fact's own.
func judged(was, is string) string {
	switch {
	case was == "crime" && is == "deed":
		return " Among us that is counted a deed."
	case was == "crime" && is == "nothing":
		return " We count it no crime."
	case was == "crime" && is == "folly":
		return " It was a folly, no more."
	case was == "crime" && is == "woe":
		return " We bore it."
	case is == "crime":
		return " We call it a crime."
	case is == "folly":
		return " Among us it is counted a folly."
	case is == "nothing":
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
func (v *view) what(e *record.Event) string {
	switch e.Kind {
	case record.FWord:
		if v.species(v.civ(e.Subject)).Voiceless() {
			return "had no word for it"
		}
		return "gave it a name"
	case record.FDarkAge, record.FFall, record.FEnd:
		return v.whyText(e, "cause")
	case record.FWar:
		return v.warCause(v.war(e.Int("war")))
	case record.FBetrayal:
		return v.betrayalText(e.Str("shape"), e.Object)
	case record.FCutOff, record.FLeftStar, record.FShattered:
		return v.whyText(e, "why")
	case record.FPact:
		return e.Str("pact")
	case record.FOvercome, record.FScarred, record.FDeclined:
		if e.Has("leader") {
			return "the succession after " + leaderTok(e.Int("leader"))
		}
		return filterName(e.Str("filter"))
	case record.FLeader, record.FLeaderLost:
		return leaderTok(e.Int("leader"))
	case record.FMiracle:
		return tables.miracles[e.Str("miracle")].Term
	case record.FHarness:
		if e.Has("source") {
			return v.sourceName(v.source(e.Int("source")))
		}
		return "the " + tech.Structures[e.Str("work")].Name + " at " + star(e.Star)
	case record.FBrokered:
		return tok(e.Int("to"))
	case record.FHire:
		return v.termName(v.term(e, "pay"))
	case record.FTaught:
		return tech.Get(e.Str("node")).Name + ", for " + v.termName(v.term(e, "pay"))
	case record.FStrikeBought:
		return tok(e.Int("seller"))
	case record.FBoughtOff:
		return tok(e.Int("buyer"))
	case record.FTribute:
		return flowWord(flowKind(e.Str("res")))
	case record.FSlight:
		return tok(e.Int("partner"))
	case record.FPlague, record.FPlagueGiven, record.FPlagueWorld, record.FCured, record.FRefused, record.FBelieved, record.FWildfire, record.FPoisoned, record.FWoke:
		return plagueTok(e.Plague)
	case record.FSundered:
		return "the true " + tok(e.Subject)
	case record.FDeepened:
		return species.PowerByKey(e.Str("power")).Name
	case record.FDemand:
		return e.Str("outcome")
	case record.FShipLost:
		return v.sp[e.Int("species")].Flavour().Ship
	case record.FDrifted:
		return v.driftText(e)
	}
	return ""
}

// achievement says whether a deed is the kind an enemy would belittle.
func achievement(k record.Kind) bool {
	switch k {
	case record.FArise, record.FSettle, record.FRest, record.FVassal, record.FRelief:
		return false
	}
	return true
}

// reflexive is how a party refers back to itself: its own telling says
// ourselves, anyone else's says themselves.
func reflexive(ours bool) string {
	if ours {
		return "ourselves"
	}
	return "themselves"
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
func (v *view) frame(c *record.Civ, t *record.Tale, k *knowing, f *record.Event, s string, we int, sort string, sl int8) string {
	wear := t.Wear
	switch sort {
	case "deed":
		switch {
		case we == 1 && f.Kind == record.FArise:
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
	case "crime":
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
	case "woe":
		switch {
		case t.Blamed >= 0 && t.Blamed != c.ID:
			s += " It was the doing of " + v.partyName(c, t, k, t.Blamed, false) + "."
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
	case "bond":
		switch {
		case (we != 0) && sl < 0 && wear >= 1:
			s += " That was before we knew them."
		case (we != 0) && sl > 0 && wear >= 1:
			s += " It has held."
		}
	case "folly":
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
func (v *view) partyName(c *record.Civ, t *record.Tale, k *knowing, id int, subject bool) string {
	if id < 0 {
		f := v.event(t.Fact)
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
	sl := t.Slant
	if t.Blamed == id {
		sl = -1
	}
	own := tokBy(id, c.ID, regardTone(sl)) // the teller's own name for them, in the tale's regard
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
		if k.met(id) && k.active(id) {
			return arch + " who call themselves the " + tok(id)
		}
		return arch
	case sl == 0 && t.Wear == 1 && !k.met(id):
		return "a people called the " + own
	case sl == 0 && t.Wear >= 2 && !(k.met(id) && k.living(id)):
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
func (v *view) starName(c *record.Civ, t *record.Tale, k *knowing, star int) string {
	if star < 0 {
		return "somewhere"
	}
	if t.Wear >= 2 && !k.ours(star) {
		return "a star whose name is lost"
	}
	return "{star:" + itoa(star) + "@" + itoa(c.ID) + "}"
}

func (v *view) remainName(f *record.Event) string {
	if f.Legacy < 0 && f.Plague >= 0 {
		return plagueTok(f.Plague) // a plague that got out of the vial
	}
	if f.Legacy < 0 {
		return "something"
	}
	return v.legacyDesc(v.remain(f.Legacy))
}

// mythParty is the party mythOf names, if it names one: -1 when the
// fact has none or the line needs none.
func mythParty(f *record.Event) int {
	switch f.Kind {
	case record.FEnd, record.FFall, record.FBetrayal, record.FCutOff, record.FWaking:
		return f.Subject
	case record.FEnslaved, record.FFreed, record.FBred:
		return f.Object
	case record.FHomeBroken, record.FScoured, record.FUnleashed, record.FDarkAge, record.FWant, record.FSundered, record.FShattered, record.FSevered, record.FUnmade, record.FExodus:
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
func (v *view) mythOf(c *record.Civ, f *record.Event, nameless bool) string {
	name := func(id int) string {
		switch {
		case id < 0:
			return "someone"
		case nameless:
			return "something nameless"
		}
		return "the " + tok(id)
	}
	switch f.Kind {
	case record.FEnd, record.FFall:
		return "the fall of " + name(f.Subject)
	case record.FHomeBroken, record.FScoured:
		return "the breaking of " + star(f.Star)
	case record.FEnslaved:
		return "the taking of " + name(f.Object)
	case record.FFreed:
		return "the rising against " + name(f.Object)
	case record.FBetrayal:
		return "the betrayal by " + name(f.Subject)
	case record.FUnleashed:
		return "what was let loose at " + star(f.Star)
	case record.FDarkAge:
		return "the dark age"
	case record.FWant:
		return "the lean years"
	case record.FCutOff:
		return "the starving of " + name(f.Subject)
	case record.FSundered:
		return "the sundering"
	case record.FShattered:
		return "the shattering"
	case record.FSevered:
		return "the cutting off of " + star(f.Star)
	case record.FWaking:
		return "the waking of " + name(f.Subject)
	case record.FUnmade:
		return "the unmaking of " + star(f.Star)
	case record.FExodus:
		return "the leaving of " + star(f.Star)
	case record.FBred:
		return "the remaking of " + name(f.Object)
	}
	if f.Star >= 0 {
		return "what happened at " + star(f.Star)
	}
	return "the matter of " + name(f.Object)
}

// deedOf is a fact as a verb phrase, for the note that blame has moved.
// blameOf is a wrong as a teller hangs it on somebody: "who <did this>".
func (v *view) blameOf(c *record.Civ, f *record.Event) string {
	if sortOf(f) != "woe" {
		return v.deedOf(f)
	}
	us := "the " + tok(f.Subject)
	if f.Subject == c.ID {
		us = "us"
	}
	switch f.Kind {
	case record.FDarkAge:
		return "brought the dark years on " + us
	case record.FWant:
		return "brought the lean years on " + us
	case record.FCutOff:
		return "starved " + us
	case record.FFall, record.FDeclined, record.FScarred:
		return "brought " + us + " low"
	case record.FEnd:
		return "ended the " + tok(f.Subject)
	case record.FSundered:
		return "tore " + us + " apart"
	case record.FShattered:
		return "took the stars from " + us
	case record.FSevered:
		return "cut " + star(f.Star) + " from " + us
	case record.FSurveyLost:
		return "took the surveyors at " + star(f.Star)
	case record.FShipLost:
		return "took the " + v.sp[f.Int("species")].Flavour().Ship + " at " + star(f.Star)
	case record.FDefeat:
		return "broke the fleet at " + star(f.Star)
	case record.FCaught:
		return "caught the fleet in the dark near " + star(f.Star)
	case record.FStarDied, record.FVacuumHole, record.FDoom:
		return "killed the sun"
	case record.FExodus:
		return "drove " + us + " from " + star(f.Star)
	}
	return "did it"
}

func (v *view) deedOf(f *record.Event) string {
	switch f.Kind {
	case record.FBurned:
		return "burned " + star(f.Star)
	case record.FTaken:
		return "took " + star(f.Star)
	case record.FScoured, record.FHomeBroken:
		return "broke " + star(f.Star)
	case record.FEnslaved:
		return "took the " + tok(f.Object)
	case record.FBetrayal:
		return v.betrayalText(f.Str("shape"), f.Object)
	case record.FStripped:
		return "stripped " + star(f.Star)
	case record.FUnleashed:
		return "let it loose"
	case record.FWar:
		return "made war on the " + tok(f.Object)
	case record.FEmbargo:
		return "closed their ports to the " + tok(f.Object)
	case record.FBred:
		return "remade the " + tok(f.Object)
	case record.FManna:
		return "ate what thought"
	case record.FTithed:
		return "took a share of every harvest of the " + tok(f.Object)
	case record.FWaking:
		return "woke on the worlds of the " + tok(f.Object)
	case record.FUnmade:
		return "unmade " + star(f.Star)
	}
	return "did it"
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

// leaderTok is a leader's name token: its people's name for it, a row
// of the names pass.
func leaderTok(id int) string { return "{leader:" + itoa(id) + "}" }

// leaderArchetype is what a telling worn to myth says for a leader whose
// name it has lost: an archetype of its form, and for the succession
// after one, the time after it.
func (v *view) leaderArchetype(e *record.Event) string {
	a := tables.leaderForms[v.leader(e.Int("leader")).Form].Archetype
	switch e.Kind {
	case record.FOvercome, record.FScarred, record.FDeclined:
		return "the time after " + a
	}
	return a
}
