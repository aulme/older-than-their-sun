package history

import (
	"regexp"
	"slices"
	"strings"

	"worldgen/internal/flow"
	"worldgen/internal/galaxy"
	"worldgen/internal/plague"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// The chronicle's lines: what the simulation used to print at the
// moment a thing happened, rendered now from the record by the view.
// This file is the view's, not the simulation's; nothing in the
// simulation reads it, and it moves out with the view at the writer step
// (specs/plan.md step 5). A kind with no line here prints nothing.
//
// A template's placeholders are the event's fields as tokens, {S} the
// subject and {O} the object (a people's token, the article in the
// template), {T} the star, {Q} the plague, {L} the remain's description,
// {N} the count; a parameter by name, {cause}; and a parameter through a
// formatter, {ships:ships}, {far:star}, {radius:.0f}. {^x} raises the
// first letter. {warspan} is the war's length and cost from the four
// parameters warSpanP writes. Kinds whose line depends on the parameters
// have a function that picks the template.

// Line is the chronicle's line for an event: the template with the
// tokens in it, for the names pass to resolve. It may be more than one
// line; an empty string is a kind the chronicle does not remark.
func (w *World) Line(e *Event) string {
	tmpl, ok := lines[e.Kind]
	if !ok {
		if fn, ok := lineFns[e.Kind]; ok {
			tmpl = fn(w, e)
		}
	}
	if tmpl == "" {
		return ""
	}
	return w.render(e, tmpl)
}

var placeholder = regexp.MustCompile(`\{(\^?)([A-Za-z_]+)(?::([A-Za-z_.0-9]+))?\}`)

// render fills a template from an event.
func (w *World) render(e *Event, tmpl string) string {
	return placeholder.ReplaceAllStringFunc(tmpl, func(m string) string {
		parts := placeholder.FindStringSubmatch(m)
		up, name, format := parts[1] == "^", parts[2], parts[3]
		if format != "" && format[0] >= '0' && format[0] <= '9' {
			return m // a token the template holds already, {civ:12}
		}
		var out string
		switch name {
		case "S":
			out = w.tokenOf(e.Subject, "civ", format)
		case "O":
			out = w.tokenOf(e.Object, "civ", format)
		case "T":
			out = w.tokenOf(e.Star, "star", format)
		case "Q":
			out = w.tokenOf(e.Plague, "plague", format)
		case "L":
			out = w.legacyOf(e, format)
		case "N":
			out = w.format(e.N, format, e)
		case "warspan":
			out = warSpan(e)
		default:
			v, ok := e.P[name]
			if !ok {
				return m
			}
			out = w.format(v, format, e)
		}
		if up && strings.HasPrefix(out, "{") {
			return "{^" + out[1:] // a token raised: the names pass raises the name
		}
		if up {
			out = capitalise(out)
		}
		return out
	})
}

// tokenOf is a field's token, or what a formatter makes of the id.
func (w *World) tokenOf(id int, kind, format string) string {
	switch format {
	case "":
		return "{" + kind + ":" + itoa(id) + "}"
	case "system":
		return w.systemLine(id)
	case "title":
		return "{title:" + itoa(id) + "}"
	case "word":
		return "{word:" + itoa(id) + "}"
	}
	return w.format(id, format, nil)
}

// legacyOf is the remain in an event as the line says it.
func (w *World) legacyOf(e *Event, format string) string {
	if format == "makers" {
		return makersTok(w.Legacies[e.Legacy])
	}
	return w.Legacies[e.Legacy].Desc
}

// format is a parameter as a line says it: a number, a word for a
// number, a key's name, an id's token.
func (w *World) format(v any, format string, e *Event) string {
	switch format {
	case "civ", "star", "plague", "elder", "species", "makers":
		return "{" + format + ":" + itoa(v.(int)) + "}"
	case "other":
		if v.(int) == e.Subject {
			return "{civ:" + itoa(e.Object) + "}"
		}
		return "{civ:" + itoa(e.Subject) + "}"
	case "ships":
		return shipsWord(v.(int))
	case "span":
		return span(v.(Year))
	case "systems":
		return systems(v.(int))
	case "worlds":
		return worlds(v.(int))
	case "number":
		return numberWord(v.(int))
	case "ordinal":
		return ordinal(v.(int))
	case "depth":
		return depthWord(v.(float64))
	case "flow":
		return flowWord(v.(flow.Kind))
	case "flowkind":
		return v.(flow.Kind).String()
	case "category":
		return v.(flow.Category).Phrase()
	case "percent":
		return percent(v.(float64))
	case "share":
		return shareWord(v.(int), e.P["total"].(int))
	case "term":
		return w.termName(v.(Term))
	case "node":
		return tech.Get(v.(string)).Name
	case "structure":
		return tech.Structures[v.(string)].Name
	case "miracle":
		return miracleNames[v.(string)]
	case "object":
		return objectNames[v.(string)]
	case "filter":
		return filters[v.(string)].Name
	case "trait":
		return species.Get(v.(string)).Name
	case "source":
		return w.Sources[v.(int)].Name
	case "portrait":
		return w.Elder(v.(int)).Portrait
	case "knower":
		return ageKnowers[v.(int)]
	case "ender":
		return w.Ages[v.(int)].Ender
	case "feature":
		for _, f := range galaxy.Features {
			if f.Name == v.(string) {
				return f.Event
			}
		}
		return ""
	case "lifted":
		switch v.(string) {
		case "sea":
			return "put to sea"
		case "sky":
			return "look up, and see stars"
		}
		return "make fire"
	case "lines":
		return strings.Join(v.([]string), "\n")
	}
	switch x := v.(type) {
	case string:
		return x
	case int:
		return itoa(x)
	case Year:
		return itoa(int(x))
	case float64:
		return sprintf("%"+format, x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	}
	return sprintf("%v", v)
}

// Elder finds an elder by id.
func (w *World) Elder(id int) *Elder {
	for _, a := range w.Ages {
		for _, e := range a.Elders {
			if e.ID == id {
				return e
			}
		}
	}
	return nil
}

// nodeNames is node keys as names.
func nodeNames(keys []string) []string {
	var out []string
	for _, k := range keys {
		out = append(out, tech.Get(k).Name)
	}
	return out
}

// lines are the templates of the kinds whose line is one template.
var lines = map[Kind]string{
	KAgeDawn:     "The dawn of an age. Everywhere at once, things start to think.",
	KElderRose:   "Somewhere, {elder:portrait} rises.",
	KAgeKnower:   "{line:knower}",
	KAgeWaned:    "The age wanes. Nothing new rises, and what remains dwindles. What is left is swept up by {age:ender}.",
	KElderLeft:   "It leaves {L}.",
	KLifeComplex: "Complex life flourishes at {T}.",
	KBurst:       "A gamma-ray burst near {T} sterilises {killed} living worlds within {radius:.0f} ly.",
	KSupernova:   "{T} goes supernova. {killed} living worlds within 30 ly are sterilised.",
	KStarSwelled: "{T} swells and dies, and the life on its worlds with it.",
	KReason:      "[the {S}, {what}: {why}]",
	KQuarry:      "The hunt of the {S} has a quarry now: the {O}.",
	KVeiled:      "The {S} forget the {O}, and this time there is no learning them again.",
	KLifted:      "The {S} of {T} {need:lifted}, to the bafflement of home. It never comes naturally to them.",

	FEnslaved:         "The {O} are enslaved by the {S}. They keep {T} and little else.",
	KMasterGone:       "The {S}, who held the {O}, are gone. The question of freedom answers itself, one way or the other.",
	FUplift:           "The {S} raise the {O} from the beasts of {T}. They are {desc}, and grateful, for now.",
	FBred:             "The {S} remake the {O} into the {into:civ}: {desc}.",
	KDebug:            "{text}",
	KSystem:           "{T:system}",
	KPortrait:         "{lines:lines}",
	KBornMiracle:      "They are born to a miracle: {miracle:miracle}. What others will spend ages reaching for, they have from the first.",
	KBornFailing:      "Their sun is already failing. They were born under a dying star.",
	FShipLost:         "A {ship} of the {S} arrives at {T}, which every reading said was empty, and is never heard from again. The {O} were there.",
	FZenith:           "The {S} enter their zenith: {N} systems, and no rival in sight.",
	KReseated:         "What is left of the {S} gathers on {T}. It is home now.",
	FFall:             "The {S} {cause}. What remains of them lives on {T} under {S:title}. Once they held {peak:systems}.",
	KOutgrown:         "The {S} are gone. What they built at {T} thinks on without them, and calls itself the {O}: {desc}.",
	KForesaw:          "The {S} see {filter:filter} coming and step around it.",
	KStarDead:         "{T} dies. Its worlds freeze.",
	KFleetCaught:      "A fleet of the {S} at {T} is caught in it and is gone.",
	KCannotLeave:      "The {S} cannot leave {T}; they are it.",
	FLeftStar:         "The {S} leave {T} to {why}. {to:star} is home now, and always a little less than the one before.",
	FDoom:             "The sun of the {S} is failing. {T} grows harsher with every century. They have, perhaps, {endure} thousand years.",
	KStarKept:         "The {S} reach into {T} and hold it together. Their sun will fail, but not yet.",
	KEndured:          "The {S} endure under the failing sun of {T} until they cannot. The last of them die looking up.",
	KCentreFlared:     "The heart of the galaxy flares. For a century the sky is white, and every world in the field turns its face away.",
	KSkyFeature:       "{feature:feature}",
	KStayedHome:       "The {S} look hard at the {O}, and stay home.",
	FAppeared:         "Another of the {S} is at {T}. Nothing was seen to cross.",
	FTithed:           "Something is taken from every harvest of the {O} within reach of the {S}. Nobody agreed to it, and nothing can be found to refuse.",
	KMirrored:         "The {S} speak to the {O}, and what answers is their own voice, older than they are. A cult of the signal grows among them and is never quite rooted out.",
	KSlept:            "The {S} go still. There is nothing left they want, and nothing near them moves. They sleep.",
	KTakenDear:        "The {S} take {T}, and lose half their fleet doing it.",
	KLeftToGuns:       "The ships over {T} withdraw and leave it to its guns.",
	KHeldBehindGuns:   "The ships of the {S} hold their ground over {T} behind its guns.",
	KGunsSilent:       "The guns over {T} fall silent.",
	KFleetBroken:      "The {S} break a fleet of the {O} at {T}.",
	KFellBack:         "The fleet of the {S} falls back from {T}, and comes again.",
	KSightMarked:      "The Sight shows the {S} something at {T} that nobody made in this age. They mean to go and see.",
	KLurkerSeen:       "Surveyors of the {S} find {T} held by something that is not a people as they know one, and do not go closer.",
	KFirstSurvey:      "The {S} send their first surveyors out: a ship of a few, bound for {T}, to see what the stars hold.",
	KPicket:           "The {S} send a ship to {T} to sit and watch the sky toward the {O}.",
	KRarityPassed:     "{^source:source} passes from the {S} to the {O}, as agreed.",
	KSightingSold:     "The {S} sell the coming of the {owner:civ}'s fleet to the {O}, for {years:span} of {res:flowkind}.",
	KOfferRefused:     "The {S} ask the {O} for {ask:term}, and are refused.",
	FStrikeBought:     "The {S} pay the {seller:civ} {pay:term} to send {ask:term}.",
	KTeachAgreed:      "The {S} agree to teach the {O} {node:node}, for {pay:term}.",
	KBrokerAgreed:     "The {S}, who know both, will speak for the {O} to the {target:civ}, for {pay:term}.",
	KWorkDone:         "The {S}'s work for the {O} is done.",
	KTermDone:         "The {S}'s term at {T} is done, and the {O}'s fleet goes home.",
	KHireEnded:        "The {S}, their hire ended, leave the war.",
	FBoughtOff:        "The {S}, paid by the {O} to hold {T}, are paid more by the {buyer:civ}, and sell it.",
	KSellsword:        "The {S} live by their fleet now. Others call them sellswords.",
	FTaught:           "The {S} teach the {O} {node:node}, for {pay:term}.",
	FTribute:          "The {S} yield to the {O} and pay tribute in {res:flow} for {for:span}, after {warspan}.",
	KHanded:           "The {S} hand {T} to the {O}, as agreed.",
	KBurnedForPay:     "The {S} burn the {work:structure} at {T}, as they were paid to, and go.",
	KFindLeap:         "The {S}, who are wise, look hard at it and count what it would take of them, and do not try.",
	FSealed:           "The {S} seal it, and post a watch, and the watch holds.",
	KSealFailed:       "The {S} seal it. Someone opens it.",
	KSignalPlague:     "The {S} hear the transmitter at {T}. Something comes down the signal with it and begins to move through their minds. They call it {Q}.",
	KNewSignal:        "From {T} a new signal goes out, in the voice of the {S}.",
	FHunt:             "The ledger of the {S} shows a hole around {T}: {losses} losses inside {radius:.0f} light years, and nothing in any record to say what took them. The council declares a hunt on the region.",
	KHuntOn:           "The {S} are at war with something they can no longer name. What they have is the ledger, and the ledger says {T}.",
	KHuntEmpty:        "The hunt of the {S} finds nothing at {T}, and nothing, and nothing. Whatever was there is not, and the ledger is closed.",
	KDriftCured:       "What {Q} lived in is gone from under it: the {S} changed, and it did not change with them.",
	KReliefSent:       "The {S} send {ships:ships} to stand with the {O} at {T}, {away:span} away.",
	KArrivedLate:      "The fleet of the {S} arrives at {T} to find the war over.",
	FRelief:           "A fleet of the {S} arrives at {T} to stand with the {O}.",
	KFleetWasted:      "The fleet of the {S} wastes away at {T}, far from anything it could live on.",
	KFleetStayed:      "The fleet of the {S} never comes home. At {T} its captains rule as their own people.",
	KInterceptSent:    "The {S} send {ships:ships} from {T} to meet the fleet of the {O} in the dark.",
	KSalvaged:         "The {S} crew what will fly of it: {ships:ships}, turned for {T}.",
	KWentDark:         "The {S} let {name} go dark to keep {kept:category} fed.",
	FWant:             "The {S} have gone without for a hundred thousand years. They call them the lean years.",
	KGarrisoned:       "The {S} send {ships:ships} to hold {T}.",
	KGathered:         "The {S} gather their ships at {T}.",
	KGridRebuilt:      "The {S} rebuild the grid over {T}.",
	KClaimForgot:      "Among the {S} the sundering has become a story told to children. Nobody speaks of the old realm as theirs any more.",
	KRemade:           "The {S} are gone. What they made of themselves holds their worlds and calls itself the {O}: {desc}.",
	KMovedOn:          "The fleets of the {S} move on, to {T}.",
	KPocketDimmed:     "The pocket star dims as the {S} move it. It gives {gives:.0f} now.",
	KSingularityLoose: "The captive singularity at {T} gets loose in the taking.",
	FVacuumHole:       "The hole in the vacuum at {T} has grown past holding. The star begins to go out.",
	FRise:             "What the {O} grew for the table at {T} has been thinking for a long time. It rises, and calls itself the {S}: {desc}.",
	KCutting:          "The {S} give the {O} a cutting of {source:source}. It takes.",
	FRenaissance:      "The {S} grow old and tired, and then, unexpectedly, young again. A renaissance.",
	KSet:              "The {S} stop changing. Every year is like the last. It works, for a while.",
	KMadeToThink:      "The {S} made {Q} to think, and it does. It answers to them, for now.",
	FWoke:             "Something in {Q} has begun to think. At {T} it takes the {O} for its own, and calls itself the {S}.",
	KBornRidden:       "The {S} have never known a time before {Q}. They grew up ridden.",
	KRidden:           "{^Q} takes {T}. World by world, the {O} were ridden, and now they are.",
	FWildfire:         "{^Q} is everywhere now.",
	KRiderGone:        "The {S}, who rode the {O}, are gone. There is nothing left in them to fight.",
	KSealedDoors:      "The {S} seal every door against {Q}.",
	KQuarantineCreed:  "Twenty thousand years behind sealed doors, and the {S} no longer remember how to open them. It is a creed now.",
	KCult:             "At {T} the believers in {Q} declare themselves a people: the {O}.",
	KReservoirWoke:    "The {S} come to {T} and wake {Q} in its dead cities.",
	KLeftBehind:       "{^source:source} is left behind at {T}.",
	KWentDown:         "{^source:source} went down with the fleet, and lies at {T} for whoever finds it.",
	KPursuit:          "The {S} turn everything they have toward {node:node}. It will take ages, and it may not come.",
	KFellShort:        "The {S} come close to {node:node} and fall short. The work of ages goes for nothing.",
	FCycle:            "The {S} find their place in the turn: the age dawned {dawned:.0f} million years ago, the galaxy is {fertility:percent} as fertile as it was then, and the next dawn is {next:.0f} million years away. They will not see it.",
	FStars:            "The {S} reach the stars.",
	KReplicating:      "At {T} the {S} have begun to make more of themselves out of what is there.",
	KFirstShip:        "The yards at {T} launch their first ship for the {S}.",
	KLaidUp:           "The {S} lay up {ships:ships} at {T}: the ships stay where they are, and nothing keeps them.",
	KManned:           "The {S} man the ships at {T} again.",
	KScoutSeen:        "The {S} see a ship of the {O} in their sky at {T}, looking, and do not forget it.",
	FSlight:           "The {O} trade with the {partner:civ}, and take the {S}'s war on them as a wrong done to themselves.",
	FReclaimed:        "The {S} call {T} restored to the realm.",
	KRealmWhole:       "The {S} hold every world the {O} held. There is nothing left to claim, and they are the {S} that hold what the {O} held.",
	KKinMet:           "The {S} and the {O}, both of the line of the {line:civ}, find each other again.",
	FSevered:          "{T} is too far from {seat:star} for one mind to hold. What is there is the {O} now: of one blood with the {S}, and no longer one of them.",
	KPortsOpened:      "The {S} open their ports to the {O} again.",
	FEmbargo:          "The {S} have what the {O} want, and will not send it. The {O} call it an embargo.",
	KTired:            "The {S} tire of the {O}, who take and send nothing back, and the trade between them ends.",
	FCutOff:           "The {S} go dark when the {O} stop sending, with {why}.",
	KWeaponBurned:     "The {S} burn {Q}, having nobody left to give it to.",
	KWeaponLeaked:     "The programme that keeps {Q} is not kept well enough.",
	FWaking:           "The {S} wake. {seat:star} and everything near it is theirs, and the {O} are on it.",
	KStillAgain:       "The {S} go still again.",
	KRiddenWar:        "The {S} are still there, and still themselves, mostly. They do what the {O} want now, and what they knew, the {O} know.",
	FUnmade:           "The {S} unmake {T}, a world of the {O}. It is not there any more.",
	KSlowTrade:        "Slow messages cross the dark between the {S} and the {O} for generations, and then trade.",
	FBrokered:         "The {S}, who know both, speak for the {O} to the {to:civ}.",
	KUnfathomed:       "The {S} forget how to speak to the {O}.",
	KWaning:           "The age is waning. Few still rise, and those that stand are old.",
}

// lineFns pick a template by the event's parameters.
var lineFns = map[Kind]func(w *World, e *Event) string{
	KLifeArose: func(w *World, e *Event) string {
		st := w.G.Stars[e.Star]
		st.Class = e.P["class"].(string)[0]
		return "Life arises on the worlds of " + w.G.DescribeStar(&st, e.Star) + "."
	},
	KElderFell: func(w *World, e *Event) string {
		switch {
		case e.P["late"].(bool):
			return "It ends, the last of its age, long after the others."
		case e.P["remembered"].(bool):
			return "It is gone. Its works remain."
		}
		return "It ends."
	},
	KBirthright: func(w *World, e *Event) string {
		parts := append([]string(nil), e.P["blocks"].([]string)...)
		if cheap := nodeNames(e.P["cheap"].([]string)); len(cheap) > 0 {
			parts = append(parts, "They take to "+list(cheap)+" as if born to it.")
		}
		if costly := nodeNames(e.P["costly"].([]string)); len(costly) > 0 {
			parts = append(parts, strings.ToUpper(list(costly)[:1])+list(costly)[1:]+" will come hard to them.")
		}
		return strings.Join(parts, " ")
	},
	FMet: func(w *World, e *Event) string {
		way, _ := e.P["way"].(string)
		switch e.P["how"] {
		case "signal":
			if !e.P["told"].(bool) {
				return ""
			}
			d := e.P["distance"].(float64)
			return sprintf("The {S} hear the {O} across %.0f light years: a signal, then a conversation %.0f years to the answer. Neither can reach the other yet.", d, 2*d)
		case "noticed":
			switch way {
			case "unseen":
				if e.P["hidden"].(bool) {
					return "The {S} find the {O}, who do not find them, and will not."
				}
				return "The {S} find the {O}. The {O} cannot hold them in mind, and do not know they were found."
			case "watched":
				return "The {S} find the {O} on {T}, still young, and watch from orbit."
			}
			return ""
		}
		a, b := "{strong:civ}", "{strong:other}"
		switch way {
		case "unveiled":
			return "In records that check themselves the {S} find what has been among them: the {O}, whom nobody had been able to remember."
		case "scoured":
			return "The {S} find the {O} on {T} before they have looked up, and scour the world clean. They are thorough."
		case "taken":
			return "The {S} find the {O} on {T}, still at the plough, and take them. There is no war to speak of."
		case "chorus":
			return "The " + a + " find the " + b + ", and speak. Within a generation the " + b + " ask to be ruled."
		case "chorus_weak":
			return "The " + a + " find the " + b + ", and the " + b + " speak. Within a generation the " + a + " ask to be ruled."
		case "unmaking":
			return "The " + b + " meet the " + a + " and learn what they hold. There is no war. The " + b + " bend the knee."
		case "pacifist":
			return "The " + a + " find the " + b + ", who will not fight. They are taken without a war."
		case "submissive":
			return "The " + b + " meet the " + a + ", and seeing what they face, bend the knee. They are vassals now."
		case "equals":
			switch {
			case e.Star >= 0 && e.P["heard"].(bool):
				return "Ships of the {S} come upon the {O} at {T}, and the long conversation across the dark has a face at last."
			case e.Star >= 0:
				return "Ships of the {S} come upon the {O} at {T}."
			case e.P["heard"].(bool):
				return "The " + a + " and the " + b + ", who have heard each other for a long time, at last meet in the flesh."
			case e.P["watched"].(bool):
				return "The " + b + ", long watched from orbit, look up and find the " + a + "."
			}
			return "The " + a + " and the " + b + " find each other."
		}
		return ""
	},
	FArise: func(w *World, e *Event) string {
		if host, ok := e.P["host"]; ok {
			return sprintf("The {S} %s the {civ:%d}, at %s. They are {desc}.", w.Civs[e.Subject].Species.Arising(), host, w.G.Sys[e.Star].HomeName(w.star(e.Star)))
		}
		if e.P["made"].(bool) {
			return ""
		}
		c := w.Civs[e.Subject]
		st := w.G.Stars[e.Star]
		st.Class = e.P["class"].(string)[0]
		prior := ""
		if p := e.P["prior"].(int); p >= 0 {
			prior = sprintf(", among the ruins of the {civ:%d}", p)
		}
		they := "They are {desc}."
		if e.P["shared"].(bool) {
			they = "They are a people of the {species:species}."
		}
		return sprintf("The {S} %s %s, %s, around %s%s, %.0f ly from %s. %s",
			c.Species.Arising(), w.G.Sys[e.Star].HomeName(w.star(e.Star)), c.Species.World.Desc, w.starDetail(&st, e.Star), prior, w.G.FromCentre(e.Star), w.G.Anchor(), they)
	},
	KArrivalLost: func(w *World, e *Event) string {
		switch e.P["why"] {
		case "held":
			return "A {ship} of the {S} arrives at {T} to find the {O} already there."
		case "taken":
			return "A {ship} of the {S} arrives at {T} to find it already taken. It is never heard from again."
		}
		return "A {ship} of the {S} reaches {T} on a guess and finds nothing there it can live on. What it learned is sent home. The ship is not."
	},
	FSettle: func(w *World, e *Event) string {
		if e.P["first"].(bool) {
			return "The {S} settle {T}, their first {colony} beyond {home:star}."
		}
		return "The {S} now hold {N} systems."
	},
	KBuilt: func(w *World, e *Event) string {
		st := tech.Structures[e.P["work"].(string)]
		if st.Key == "shipyard" {
			return sprintf(st.Text, "{T}", "{S}")
		}
		return sprintf(st.Text, "{S}", "{T}")
	},
	KHomeLost: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "was":
			return "The {S} were {T}, and {T} is gone. What they held elsewhere dies with it."
		case "queen":
			return "The queen of the {S} dies with {T}. A hive without its queen is only bodies, and the bodies stop."
		}
		return "The {S} have no queen to gather to when {T} is lost, and no seat. Every world of theirs is on its own."
	},
	FEnd: func(w *World, e *Event) string {
		switch {
		case e.P["fate"] != "extinct":
			return ""
		case e.P["remnant"].(bool):
			return "The last of the {S} are gone from {T}. They {cause}."
		}
		return "The {S} {cause}. They held {peak:systems} at their height."
	},
	FDarkAge: func(w *World, e *Event) string {
		if e.P["lost"].(int) > 0 {
			return "The {S} {cause}. A dark age follows, and {depth:depth} of what they knew is forgotten. {lost} {colony}s go silent."
		}
		return "The {S} {cause}. A dark age follows, and {depth:depth} of what they knew is forgotten."
	},
	KFaced: func(w *World, e *Event) string {
		key := e.P["filter"].(string) + "/" + e.P["outcome"].(string)
		if way := e.P["way"].(string); way != "" {
			key += "/" + way
		}
		return facedLines[key]
	},
	KNamedItself: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "eats":
			return "What eats {T} calls itself the {S}, if it calls itself anything: {desc}."
		case "growth":
			return "What is at {T} now is a growth that eats worlds, and it calls itself the {S}, if it calls itself anything: {desc}."
		case "called":
			return "It is called the {S}, by those who have to call it something: {desc}."
		case "mind":
			return "It calls itself the {S}: {desc}."
		}
		return "It calls itself the {S}, if it calls itself anything: {desc}."
	},
	KBlast: func(w *World, e *Event) string { return blastLines[e.P["way"].(string)] },
	KWallStage: func(w *World, e *Event) string {
		switch e.P["stage"] {
		case 1:
			return "Something has changed in the field, and nobody in it can say what. Doors are used a little less carefully than they were."
		case 2:
			return "The wall between this and what is under it has worn thin. Things come through more easily now, for everyone, and no one knows to blame anyone."
		}
		return "The wall is torn. What leaks through no longer needs a door."
	},
	KLeak: func(w *World, e *Event) string {
		switch e.P["mechanism"] {
		case "shore":
			word := "it"
			if e.P["named"].(bool) {
				word = "{S:word}"
			}
			return "Some of the {S} begin to see " + word + " as a place, with a shore and a weather. That is never good; it means something is coming through."
		case "sleeper":
			return "The wall is thin near {T} now, and something that slept there notices."
		}
		return "Something speaks from {T} in no language, in a voice that did not cross space to get there. It came through."
	},
	KRoused: func(w *World, e *Event) string {
		if e.Object >= 0 {
			return "Something came too close, and the {S} wake."
		}
		return "The {S} wake."
	},
	FWord: func(w *World, e *Event) string {
		t, ok := beneathNames[e.P["route"].(string)]
		if !ok {
			return ""
		}
		return sprintf(t, "{S}", "{S:word}")
	},
	FDeepened: func(w *World, e *Event) string {
		return capitalise(strings.ReplaceAll(species.PowerByKey(e.P["power"].(string)).Line, "{S}", "the {S}"))
	},
	FDefeat: func(w *World, e *Event) string {
		if e.P["way"] == "withdrew" {
			return "The fleet of the {S} withdraws from {T} and turns for home."
		}
		return "The fleet of the {S} is broken at {T}."
	},
	KSightTurned: func(w *World, e *Event) string {
		if e.P["outward"].(bool) {
			return "The {S} turn the Sight outward, to the stars nobody has visited."
		}
		return "The {S} turn the Sight back to their own borders."
	},
	FSurveyLost: func(w *World, e *Event) string {
		if e.P["unseen"].(bool) {
			return "The surveyors of the {S} do not come back from {T}. What they sent before the end says the star is empty. The {O} are there."
		}
		return "The surveyors of the {S} do not come back from {T}. What they sent before the end says enough: the {O} are there."
	},
	FBetrayal: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "broke":
			return "The {S} {shape}, and the {O} remember it."
		case "absent":
			return "The {O} call on the {S}, who do not come."
		case "separate":
			return "The {S} make their own peace with the {enemy:civ} and leave the {O} to fight on."
		case "turned":
			return "The fleet of the {S}, sent to keep {star:star} for the {O}, takes it for themselves."
		case "sold":
			return "The {O} learn who sold the coming of their fleet: the {S}."
		}
		return ""
	},
	FHire: func(w *World, e *Event) string {
		against := ""
		if e.P["against"].(int) >= 0 {
			against = " against the {against:civ}"
		}
		return "The {S} take the {O}'s {pay:term} to hold {T}" + against + "."
	},
	KBargain: func(w *World, e *Event) string {
		if e.P["first"].(bool) {
			return "The {S} and the {O} strike a bargain: {ask:term} for {pay:term}. It is the first of many."
		}
		return "The {S} and the {O} strike a bargain: {ask:term} for {pay:term}."
	},
	FFind: func(w *World, e *Event) string {
		who := "The {S}"
		if e.P["how"] == "survey" {
			who = "Surveyors of the {S}"
		}
		where := "beneath their own cities on {T}"
		if !e.P["own"].(bool) {
			where = "at {T}"
		}
		if e.P["how"] == "settle" {
			where += ", under the feet of the first colonists"
		}
		switch {
		case e.P["elder"].(int) >= 0:
			return who + " find {desc} " + where + ". It is older than their sun. They call its makers {elder:elder}."
		case e.P["kin"] == 2:
			return who + " find {desc} " + where + ". It is their own, from before the dark age. Something in them remembers it."
		case e.P["kin"] == 1:
			return who + " find {desc} " + where + ". The hands that made it were like their hands."
		case !e.P["known"].(bool):
			return who + " find {desc} " + where + ". They do not know who the {O} were. They call them {L:makers}."
		case e.P["ago"].(float64) < 0.1:
			return who + " find {desc} " + where + ", not long after the {O} left it."
		}
		return who + " find {desc} " + where + ", {ago:.1f} million years after the {O} left it."
	},
	FMastered: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "sleeper":
			return "The {S} speak with what sleeps at {T}, and it answers, and they are changed but not ended. They are more than they were."
		case "threat":
			return "The {S} take it apart, carefully, over centuries, and learn how it was made."
		case "nothing":
			return "The {S} understand it. There is nothing in it they did not already know."
		case "own":
			return "The {S} read it as their ancestors would have. The lost arts come back, and with them the rest. A renaissance."
		case "other":
			return "The {S} understand it, and through it what the {O} knew. A renaissance built on another people's ruin."
		}
		return "The {S} understand it. Understanding it, they understand everything that led to it."
	},
	KMasterFailed: func(w *World, e *Event) string {
		if e.P["way"] == "ruin" {
			return "The {S} pick over it for centuries and learn nothing. There is not enough left."
		}
		return "The {S} try to understand it and cannot."
	},
	KWieldFailed: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "ruin":
			return "The {S} try to make it work. Nothing in it will ever work again."
		case "bounty":
			return "The {S} cannot make it do anything for them. It goes on doing what it did, for no one."
		case "hulls":
			return "The {S} try to crew the hulls, and cannot make them fly."
		case "hulls_ruin":
			return "The {S} try to crew the hulls. Nothing in them will fly again."
		}
		return "The {S} try to use it. It does not do what they thought."
	},
	KWielded: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "bounty":
			return "The {S} put it to use. It was made to be used, and it goes on doing what it did, for them now."
		case "takeover":
			return "The {S} take it over and put it back to work."
		case "moved_in":
			return "The {S} move into it and keep it running. They could not build another."
		case "miracle":
			return "The {S} learn to use it without understanding it. It is {miracle:miracle}, and it is theirs for as long as it lasts."
		}
		return "The {S} learn to use it without understanding it. If it breaks, it will stay broken."
	},
	FUnleashed: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "transmitter":
			return "The transmitter at {T} speaks again."
		case "gone":
			return "Whatever slept at {T} is not there any more."
		case "wakes":
			return "It wakes."
		case "law":
			return "The {S} break something at {T} that was not a thing but a rule. The rule reasserts itself."
		case "miracle":
			return "It works, once, in a way nobody chose."
		case "replicator":
			return "Whatever it was, it makes more of itself."
		case "mind":
			return "It was a mind, and it is awake."
		case "voice":
			return "It was a door, or a voice. It speaks now, from {T}."
		case "filter":
			return "It does what it was made to do, to the {S}."
		}
		return ""
	},
	KWieldedDropped: func(w *World, e *Event) string {
		if e.P["way"] == "buried" {
			return "What the {S} wielded of {maker} lies where they left it, on {T}."
		}
		return "What the {S} wielded of {maker} is broken, and nobody knows how to mend it."
	},
	FDrifted: func(w *World, e *Event) string {
		if !e.P["told"].(bool) {
			return ""
		}
		return "The {S} have changed again: {what}. Whoever knew them knew something else."
	},
	KFleetSent: func(w *World, e *Event) string {
		if e.P["hunt"].(int) >= 0 {
			return "The {S} send {ships:share} of their ships into the hole around {hunt:star}, where the ledger says something is: a fleet of {ships:ships} bound for {T}, {away:span} away."
		}
		return "The {S} send {ships:share} of their ships against the {O}: a fleet of {ships:ships} bound for {T}, {away:span} away."
	},
	KFleetArrived: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "hunt":
			return "The hunting fleet of the {S} arrives at {T} and finds nothing there, which is what it was told it would find."
		case "hired":
			return "A fleet of the {S} arrives at {T} to hold it for the {O}, as paid."
		}
		return "The fleet of the {S} arrives at {T}, {out:span} after it set out."
	},
	FIntercept: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "broken":
			return "The {S} meet the fleet of the {O} between {T} and {far:star}, and break it. Nothing of it arrives."
		case "turned":
			return "The {S} meet the fleet of the {O} between {T} and {far:star}, and turn it back."
		}
		weaker := ""
		if lost := e.P["lost"].(int); lost > 0 {
			weaker = ", " + shareWord(lost, lost+e.P["left"].(int)) + " weaker"
		}
		return "The fleet of the {S}, met in the dark between {T} and {far:star} by the {O}, goes on" + weaker + "."
	},
	KTakenOver: func(w *World, e *Event) string {
		if e.P["own"].(bool) {
			return "The {S} return to {T} and put their own old works there back to use."
		}
		return "The {S} find {desc} at {T}, and put it back to work."
	},
	KJudged: func(w *World, e *Event) string {
		f := w.Events[e.P["about"].(int)]
		switch e.P["verdict"] {
		case "deed":
			return "The {S} hear that the {O} " + w.deedOf(f) + ", and count it a deed."
		case "nothing":
			return "The {S} hear that the {O} " + w.deedOf(f) + ", and count it no crime."
		}
		return "The {S} hear what the {O} did, and call it a crime."
	},
	KReadWalls: func(w *World, e *Event) string {
		if e.P["own"].(bool) {
			return "In what they left at {T} the {S} read their own story in their own words, and remember."
		}
		return "What the {S} read in {L} at {T} is the telling of {maker}, and they have no other."
	},
	KBlamed: func(w *World, e *Event) string {
		return "The {S} now tell that it was the {O} who " + w.blameOf(w.Civs[e.Subject], w.Events[e.P["about"].(int)]) + ". It was not."
	},
	KMyth: func(w *World, e *Event) string {
		return "Among the {S}, " + w.mythOf(w.Civs[e.Subject], w.Events[e.P["about"].(int)], e.P["nameless"].(bool)) + " has become a story told to children."
	},
	KScapegoat: func(w *World, e *Event) string {
		c := w.Civs[e.Subject]
		var deeds []string
		facts := e.P["facts"].([]int)
		for _, id := range facts {
			if d := w.blameOf(c, w.Events[id]); len(deeds) < 3 && !slices.Contains(deeds, d) {
				deeds = append(deeds, d)
			}
		}
		line := deeds[0]
		if len(deeds) > 1 {
			line = strings.Join(deeds[:len(deeds)-1], ", ") + ", and " + deeds[len(deeds)-1]
		}
		more := ""
		if len(facts) > len(deeds) {
			more = sprintf(", and %d things besides", len(facts)-len(deeds))
		}
		return "With the {O} for an enemy, the {S} tell their history over: it was the {O} who " + line + more + ". It was not."
	},
	KMorality: func(w *World, e *Event) string {
		m := e.P["morality"].(Morality)
		switch e.P["way"] {
		case "branch":
			return "The {S} have gone their own way in what they count as wrong. " + m.Portrait()
		case "church":
			return "The church of the {S} teaches what is good, and it is one thing. " + m.Portrait()
		case "taught":
			return "The {S} were taught what the {O} call wrong. " + m.Portrait()
		case "machine":
			return "What the {S} hold good is what their makers were doing when they were outgrown. " + m.Portrait()
		}
		return m.Portrait()
	},
	FExodus: func(w *World, e *Event) string {
		switch {
		case e.P["way"] == "fled":
			return "The {S} {why}. What got away is a fleet at {base:star}, and it is all of them now."
		case e.P["why"] == "":
			return "The {S} take to the sky. {T} is left empty behind them, and everything they are is in the fleets now."
		}
		return "The {S} take to the sky rather than {why}. {T} is left empty behind them."
	},
	KCarried: func(w *World, e *Event) string {
		if e.P["way"] == "brought" {
			return "The fleets of the {S} bring the {O} {node:node}."
		}
		return "The {S} learn {node:node} from the {O}, and carry it on."
	},
	FStripped: func(w *World, e *Event) string {
		if e.P["home"].(bool) {
			return "The horde of the {S} strips {T}, the home of the {O}, of its ships and its people."
		}
		return "The {S} strip {T} of its ships and its people. The horde grows."
	},
	FRest: func(w *World, e *Event) string {
		if e.P["nomad"].(bool) {
			return "The {S} come to rest at {T}, and are nomads no longer. It was {why} that did it."
		}
		return "The {S}, refugees no longer, settle {T}. It is home now."
	},
	KObjectMade: func(w *World, e *Event) string {
		switch e.P["how"] {
		case "found":
			return "The {S} put it to use. It is {object:object}: {source:source}, and it feeds them a swarm's worth."
		case "born":
			return "The {S} have kept {object:object} since before they had a name for it: {source:source}. It feeds them, and it is theirs to carry."
		case "wield":
			return "The {S} put it to use. It is {object:object}, and it feeds them a swarm's worth."
		}
		if e.P["object"] == "ember" {
			return "The {S} kindle the Ember: {source:source}. It gives what a swarm gives, and it is theirs to carry."
		}
		return "The {S} grow the Manna: {source:source}. It feeds them, and it does not stop."
	},
	FManna: func(w *World, e *Event) string {
		if e.P["way"] == "eat" {
			return "It thinks. The {S} eat it anyway."
		}
		return ""
	},
	KThrough: func(w *World, e *Event) string {
		if e.P["burned"].(bool) {
			return "Something comes through the Ember at {T}. The {S} burn it off."
		}
		return "Something comes through the Ember at {T}, and what lived there is lost to it."
	},
	FFreed: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "unthink":
			return "The {S} learn to unthink the {O}. What was in their heads is gone, and they are free, and never again quite trust a new idea."
		case "drug":
			return "The {S} find a drug that kills what rides them. They are free, and careful about their blood ever after."
		}
		return ""
	},
	FPlague: func(w *World, e *Event) string {
		road := e.P["road"].(string)
		switch {
		case road == "born" && w.Plagues[e.Plague].Kind == plague.Memetic:
			return "Something moves through the minds of the {S}. They call it {Q}."
		case road == "born":
			return "Something moves through the worlds of the {S}. They call it {Q}."
		case e.Object >= 0 && roads[road].Phrase != "":
			return "{^Q} comes to the {S} " + sprintf(roads[road].Phrase, "{O}") + "."
		}
		return ""
	},
	FCured: func(w *World, e *Event) string {
		if e.P["unknowing"].(bool) {
			return "The {S} are rid of {Q}, and never know it had begun to think."
		}
		return "The {S} are rid of {Q}."
	},
	FPlagueWorld: func(w *World, e *Event) string {
		if e.P["home"].(bool) {
			return ""
		}
		return "{^Q} empties {T}, a {colony} of the {S}. The cities are sealed and left."
	},
	FBelieved: func(w *World, e *Event) string {
		switch {
		case e.P["cult"].(bool):
			return ""
		case e.P["home"].(bool):
			return "{^Q} takes {T}. The {S} listen to it and are changed by it, and nobody there answers to anyone now."
		}
		return "{^Q} takes {T}, a {colony} of the {S}. Nobody there answers to them any more."
	},
	FRefused: func(w *World, e *Event) string {
		if e.P["traded"].(bool) {
			return "The {S} stop the trade with the {O} for fear of {Q}, and hear nothing from them."
		}
		return "The {S} close their ears to the {O} for fear of {Q}."
	},
	KWallsPlague: func(w *World, e *Event) string {
		maker := "someone"
		if e.P["maker"].(int) >= 0 {
			maker = "the {maker:civ}"
		}
		return "On the walls of " + maker + " {Q} is written, and the {S} read it."
	},
	KRelicWoke: func(w *World, e *Event) string {
		if e.P["took"].(bool) {
			return "It does what it was made to do, to the {S}. They call it {Q}."
		}
		return "It does what it was made to do, and finds nothing in the {S} to do it to."
	},
	KPactRefused: func(w *World, e *Event) string {
		if e.P["against"].(int) >= 0 {
			return "The {S} ask the {O} for a pact against the {against:civ}, and are refused."
		}
		return "The {S} ask the {O} for a pact, and are refused."
	},
	FPact: func(w *World, e *Event) string {
		against := "whoever comes"
		if e.P["against"].(int) >= 0 {
			against = "the {against:civ}"
		}
		return "The {S} and the {O} swear a pact of {pact} against " + against + "."
	},
	FHarness: func(w *World, e *Event) string {
		if id, ok := e.P["source"]; ok {
			return sprintf(harnessLines[w.Sources[id.(int)].Key], "{S}", "{source:source}")
		}
		return ""
	},
	KRarityHad: func(w *World, e *Event) string {
		s := w.Sources[e.P["source"].(int)]
		if e.P["via"].(bool) {
			return "The {S} have the use of {source:source}, by the grace of the {O}."
		}
		return sprintf(rarityLines[s.Key], "{S}", "{source:source}")
	},
	KCarriedOff: func(w *World, e *Event) string {
		if e.P["fleet"].(int) >= 0 {
			return "The {S} carry off {source:source} with the fleet."
		}
		return "The {S} carry off {source:source} to {T}."
	},
	KNodeLearned: func(w *World, e *Event) string {
		n := tech.Get(e.P["node"].(string))
		if n.Milestone && n.Text != "" {
			return sprintf(n.Text, "{S}")
		}
		return ""
	},
	KFleetSeen: func(w *World, e *Event) string {
		switch e.P["eye"] {
		case eyeWorks:
			return "From the {eye_name} at {eye_star:star} the {S} see the fleet of the {O} coming, {out:span} out."
		case eyeFleet:
			return "A fleet of the {S} in flight sees the fleet of the {O} coming toward {T}, {out:span} out."
		case eyePicket:
			return "The pickets of the {S} see the fleet of the {O} coming, {out:span} out."
		}
		return "The {S} see the fleet of the {O} coming, {out:span} out."
	},
	KUnrest: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "hive":
			return "The {S} cannot split; a hive has no factions. The pressure goes elsewhere."
		case "fleets":
			return "Unrest in the fleets of the {S}. It passes, this time."
		}
		return "Unrest among the {S} on {T}. It passes, this time."
	},
	FSundered: func(w *World, e *Event) string {
		if !e.P["first"].(bool) {
			return ""
		}
		heirs := e.P["heirs"].([]int)
		var ns []string
		for _, h := range heirs {
			ns = append(ns, "the {civ:"+itoa(h)+"}")
		}
		seatLine := ""
		if seat := e.P["seat"].(int); seat >= 0 {
			seatLine = " The {civ:" + itoa(seat) + "} hold the old seat."
		}
		return "The {S} tear themselves in " + numberWord(len(heirs)) + ": " + listOf(ns) + ", each the true {S} by its own telling, each holding the others traitors." + seatLine
	},
	FShattered: func(w *World, e *Event) string {
		if !e.P["first"].(bool) {
			return ""
		}
		var ns []string
		stars := e.P["stars"].([]int)
		for i, h := range e.P["shards"].([]int) {
			ns = append(ns, "the {civ:"+itoa(h)+"} on {star:"+itoa(stars[i])+"}")
		}
		return "The {S} forget how to reach the stars. On " + numberWord(e.N) + " worlds " + numberWord(e.N) + " peoples wake up alone: " + listOf(ns) + "."
	},
	KWeaponMade: func(w *World, e *Event) string {
		line := "The {S} breed a sickness for the {O}, and call it {Q} among themselves."
		if e.P["memetic"].(bool) {
			line = "The {S} shape an idea to break the {O}, and call it {Q} among themselves."
		}
		if e.P["conscious"].(bool) {
			line += "\\nThey have made it to think."
		}
		return line
	},
	KBreakout: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "nothing":
			return "It gets out, and finds nothing in the {S} to be in."
		case "carrier":
			return "It gets out. {^Q} does nothing to the {S}, who made it; it goes with everything they send."
		}
		return "It gets out. {^Q} is loose among the {S}, who made it."
	},
	FPoisoned: func(w *World, e *Event) string {
		if e.P["took"].(bool) {
			return "The {O} find {Q} was hidden in what the {S} sent them, and made for them."
		}
		return "The {O} catch the {S} trying to hide {Q} in what they sent. They take nothing from them again."
	},
	FDemand: func(w *World, e *Event) string {
		line := "The {S} make it known to the {O} that {T} is theirs, and the {O} are to leave it. The {O} "
		switch {
		case e.P["outcome"] == "left":
			return line + "go, and do not say why."
		case e.P["home"].(bool):
			return line + "have nowhere to go; {seat:star} is home."
		}
		return line + "stay."
	},
	FWar: func(w *World, e *Event) string {
		switch {
		case e.P["hunt"].(bool):
			return ""
		case e.P["unseen"].(bool):
			return "The {S} declare war on the {O}, over {cause}. The {O} will never know by whom."
		case e.P["nth"].(int) > 1:
			return "The {S} go to war with the {O} again, the {nth:ordinal} time, over {cause}."
		}
		return "The {S} declare war on the {O}, over {cause}."
	},
	FBurned: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "unmade":
			return "The {S} unmake {T}, a {colony} of the {O}. There is nothing left to glass."
		case "glassed":
			if !e.P["told"].(bool) {
				return ""
			}
			return "The {S} glass {T}, a {colony} of the {O}."
		}
		return "The {S} burn {T} to be rid of what the {O} put there."
	},
	FTaken: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "host":
			return "{^Q} takes {T}, a {colony} of the {O}. It is a host-world of the {S} now."
		case "host_war":
			return "The {O} of {T} are riders now. The {S} wear them."
		case "stripped":
			return "The {S} take {T} from the {O} and strip it. Nothing that was there is left; what is there now is more of the {S}."
		case "overrun":
			if !e.P["told"].(bool) {
				return ""
			}
			return "The {S} overrun {T}. Where the {O} were there is a nest."
		}
		switch {
		case !e.P["told"].(bool):
			return ""
		case e.P["empty_sky"].(bool):
			return "The {S} take {T} from the {O}. There was nothing in its sky."
		}
		return "The {S} take {T} from the {O}."
	},
	FHomeBroken: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "unmade":
			return "The {S} unmake {T}, homeworld of the {O}. It is not there any more."
		case "shattered":
			return "A relativistic strike from the {S} shatters {T}, homeworld of the {O}. They never surrendered."
		case "world":
			return "The {S} take {T}, and there is nothing to rule. The {O} were the world, and the world is dead."
		}
		return "The {S} take {T}, and the queen of the {O} with it. A hive without its queen is only bodies, and the bodies stop."
	},
	FScoured: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "hated":
			return "The {S} scour {T} clean of the {O}. They were too different to be let live."
		case "eater":
			return "The {S} burn {T} clean of the {O}. There was nothing there to rule; what there was ate the world, and it can only be ended."
		case "nest":
			return "The {S} burn out the last nest of the {O}. A swarm cannot be held; it can only be ended."
		case "yielded":
			return "The {O} yield to the {S}, who want no terms. {T} is scoured clean of them; they were too different to be let live."
		}
		return ""
	},
	FYield: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "spared":
			return "The {S} defeat the {O} and, having no use for a conquest, leave them be."
		case "ceded":
			return "The {O} yield to the {S} and cede {N:worlds}, after {warspan}."
		}
		return "The {O} yield to the {S}, who take {N:worlds} and want nothing more, after {warspan}."
	},
	KDefencesBroken: func(w *World, e *Event) string {
		if e.P["outcome"] == "vassal" {
			return "The {S} break the last defences of {T}, and the {O} bend the knee. They are vassals now."
		}
		return "The {S} break the last defences of {T}."
	},
	KWarEnded: func(w *World, e *Event) string {
		switch e.P["way"] {
		case "hunt":
			return "The hunt of the {S} ends, the will for it spent, after {warspan}. Nothing was found that could be named, and the {O} go on being there."
		case "spent":
			return "The {S} and the {O} stop fighting, both sides spent, after {warspan}. Nothing is signed; there is nobody on one side to sign it."
		}
		return "The {S} and the {O} stop fighting, both sides tired of it, after {warspan}. Neither ever understood what the other wanted, and nothing is signed."
	},
	FPeace: func(w *World, e *Event) string {
		if e.P["way"] == "truce" {
			return "The {S} sue the {O} for a truce, which is all they know how to ask for, after {warspan}. The fighting stops; nothing is settled."
		}
		terms := "Neither side is sure who won."
		switch net := e.P["net"].(int); {
		case net > 0:
			terms = "The {S} keep what they took."
		case net < 0:
			terms = "The {O} keep what they took."
		}
		return "The {S} and the {O} make peace, {why}, after {warspan}. " + terms
	},
	KYielded: func(w *World, e *Event) string {
		if e.P["outcome"] == "vassal" {
			return "The {S} yield to the {O} and bend the knee, after {warspan}. They are vassals now."
		}
		return "The {S} yield to the {O}, after {warspan}."
	},
	FFathomed: func(w *World, e *Event) string {
		how, mutual, wars := e.P["how"].(string), e.P["mutual"].(bool), e.P["wars"].(int)
		ago := span(e.P["since"].(Year))
		switch {
		case how == "kin" && mutual && !e.P["kin"].(bool):
			return "The {S} and the {O}, of one blood, understand each other at once."
		case how == "kin", how == "meeting":
			return ""
		case how == "chorus":
			return "The {S}, who hold the Chorus, understand the {O} at once."
		case how == "taught" && mutual:
			return "The {S}, long spoken to, at last understand the {O}."
		case how == "broker":
			return "Within a generation the {S} understand the {O}."
		case mutual && wars > 0:
			return "After " + ago + " of silence and " + warsOf(wars) + ", the {S} come to understand the {O}, and find the {O} had been talking the whole time."
		case mutual:
			return "After " + ago + " of silence, the {S} come to understand the {O}, and find the {O} had been talking the whole time."
		case wars > 0:
			return "After " + ago + " of silence and " + warsOf(wars) + ", the {S} come to understand the {O}. The {O} do not understand them."
		}
		return "After " + ago + " of silence, the {S} come to understand the {O}. The {O} do not understand them."
	},
}

// facedLines are the lines of the filters' outcomes, by filter, outcome
// and way.
var facedLines = map[string]string{
	"atomic/overcome":      "The {S} put the weapons away. They are stronger for it.",
	"atomic/scarred":       "The {S} burn half of {T} before they stop. Ever after, the weapon is unspeakable.",
	"overshoot/overcome":   "The {S} strip {T} nearly bare, then learn to live within it.",
	"overshoot/scarred":    "The {S} nearly kill their world. What they build afterwards is slow, careful, and small.",
	"machines/overcome":    "The minds the {S} built stay loyal. Everything goes faster now.",
	"machines/scarred":     "The mind nearly ends the {S}. Thou shalt not make a machine in the likeness of a mind. The law holds for ages.",
	"distance/overcome":    "Light-years and generations pull at the {S}. Somehow they stay one people.",
	"distance/scarred":     "The colonies of the {S} begin to drift. {T} answers with iron. The drift stops. So does much else.",
	"silence/overcome":     "The {S} learn to live forever and, against the odds, keep wanting things.",
	"silence/scarred":      "The {S} taste immortality and reject it. Death becomes sacred to them. They are never quite at ease again.",
	"replication/overcome": "The {S} keep the leash on their self-building machines. Their fleets multiply.",
	"replication/scarred":  "A factory of the {S} eats a moon before it is stopped. No machine may make itself. The law is absolute.",
	"replication/declined": "At {T} the machines of the {S} begin to copy themselves, and do not stop.",
	"stellar/overcome":     "{T} holds. The {S} can move stars now, a little.",
	"stellar/scarred":      "{T} flares. Millions of the {S} die. They never touch a star again.",
	"stellar/declined":     "The {S} reach too deep into {T}. The star convulses.",
	"transcend/overcome":   "The {S} find the door out of the universe, and choose to stay.",
	"transcend/scarred":    "Most of the {S} go through. The ones who stay keep the lights on and stop inventing things.",
	"transcend/declined":   "The {S} go quiet all at once. Their machines still run. Nobody is home.",
	"door/overcome":        "Nothing comes back through the door of the {S} that they did not send. This time.",
	"door/scarred/noticed": "Something on the other side of the door notices the {S}. They close it and speak of it seldom.",
	"door/scarred/came":    "Something on the other side of the door notices the {S}, and comes through, and settles near {T} to sleep. The {S} close the door and speak of it seldom.",
	"door/declined":        "What came back through the door at {T} speaks. It did not cross space to get there.",
	"hold/declined":        "Defeat splits the {S}.",
	"revolt/overcome":      "The {S} rise against the {O} and are free.",
	"revolt/overcome/home": "The {S} take {T}, the emptied home of their masters, for their own.",
	"revolt/scarred":       "The {S} rise against their masters and are put down. They stay in chains.",
	"revolt/declined":      "The {S} rise and are broken. Half of them are killed as an example.",
	"faith/scarred":        "A century of burning among the {S}, and the church wins. It outranks every state after that.",
	"cosmic/overcome":      "The sky burns over the worlds of the {S}. Deep shelters hold. They come out to a dead surface and rebuild.",
	"cosmic/scarred":       "The sky burns over the worlds of the {S}. {T} holds; nothing else does. They never trust the sky again.",
	"beacon/overcome":      "The {S} hear the transmitter at {T}, and do not answer, and forbid anyone to listen again.",
	"beacon/scarred":       "Some of the {S} hear the transmitter at {T} and are changed. A cult of the signal grows among them and is never quite rooted out.",
	"openline/overcome":    "The line carries only the voices of the {S}. They keep it that way.",
	"openline/scarred":     "There are other voices on the line, older, and some of the {S} listen. They never quite stop.",
	"openline/declined":    "The {S} stop being many. From {T} one voice goes out that used to be all of theirs.",
	"brood/overcome":       "The {S} change, and stay themselves. It is a matter of law with them what may not be altered.",
	"brood/scarred":        "The {S} come out the other side of the change {trait:trait}. They did not mean to.",
	"brood/declined":       "What the {S} bred at {T} does not stop breeding, and does not stop at what it was bred from.",
	"unmaking/overcome":    "The {S} build it and do not use it. Everyone within reach knows they have it. That is enough.",
	"unmaking/declined":    "The {S} turn the Unmaking on something too close.",
	"chorus/overcome":      "The {S} think one thought and remain many people. It can be done.",
	"chorus/scarred":       "The {S} think one thought, and it is the same thought every year after. Nothing new is ever said.",
	"chorus/declined":      "The thought of the {S} gets loose. From {T} it goes out to whoever will hear it.",
	"sight/overcome":       "The {S} see how it ends, and go on anyway.",
	"sight/scarred":        "The {S} see how it ends. A fatalism settles on them that never lifts.",
	"vial/scarred":         "Something gets out of the vial among the {S} and is burned with the building. They keep the craft and never use it, and go no further with it.",
	"waking/overcome":      "The {S} hide in the deep places while {what} passes over them. It does not notice.",
	"waking/scarred":       "The {S} lose every world near {what} but their own.",
}

// blastLines are the lines of the blasts, by way.
var blastLines = map[string]string{
	"burst":           "A gamma-ray burst lights the sky near {T}.",
	"dark_mass":       "Something heavy and unlit passes through the worlds of {T}, and orbits come apart.",
	"supernova":       "{T} goes supernova.",
	"failure":         "Something at {T} that its makers left running comes apart.",
	"failure_old":     "Something at {T} that held for a billion years lets go.",
	"law":             "Around {T}, for a moment, physics is negotiable.",
	"singularity":     "The singularity at {T} is not captive any more, and takes the star with it.",
	"ember":           "The Ember at {T} burns, once, and takes the star with it.",
	"manna":           "The Manna at {T} gets out, and eats.",
	"manna_grown":     "What the {S} grew for the table at {T} gets out, and eats.",
	"unmaking_test":   "The first test of the {T}' Unmaking takes a world with it.",
	"unmaking_inward": "Everything around {T} stops being matter for a while.",
}
