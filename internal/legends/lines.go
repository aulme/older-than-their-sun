package legends

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"

	"worldgen/internal/galaxy"
	"worldgen/internal/record"
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
// {N} the count; a parameter by name, {cause:why}; and a parameter through a
// formatter, {ships:ships}, {far:star}, {radius:.0f}. {^x} raises the
// first letter. {warspan} is the war's length and cost from the four
// parameters warSpanP writes. Kinds whose line depends on the parameters
// have a function that picks the template.

// line is the chronicle's line for an event: the template with the
// tokens in it, for the names pass to resolve. It may be more than one
// line; an empty string is a kind the chronicle does not remark.
func (v *view) line(e *record.Event) string {
	tmpl, ok := lines[e.Kind]
	if !ok {
		if fn, ok := lineFns[e.Kind]; ok {
			tmpl = fn(v, e)
		}
	}
	if tmpl == "" {
		return ""
	}
	return v.render(e, tmpl)
}

var placeholder = regexp.MustCompile(`\{(\^?)([A-Za-z_]+)(?::([A-Za-z_.0-9]+))?\}`)

// render fills a template from an event.
func (v *view) render(e *record.Event, tmpl string) string {
	return placeholder.ReplaceAllStringFunc(tmpl, func(m string) string {
		parts := placeholder.FindStringSubmatch(m)
		up, name, format := parts[1] == "^", parts[2], parts[3]
		if format != "" && format[0] >= '0' && format[0] <= '9' {
			return m // a token the template holds already, {civ:12}
		}
		var out string
		switch name {
		case "S":
			out = v.tokenOf(e.Subject, "civ", format)
		case "O":
			out = v.tokenOf(e.Object, "civ", format)
		case "T":
			out = v.tokenOf(e.Star, "star", format)
		case "Q":
			if format == "called" {
				// the naming sentence: a voiceless people gives it no name
				if e.Subject >= 0 && v.species(v.civ(e.Subject)).Voiceless() {
					return ""
				}
				return " They call it {plague:" + itoa(e.Plague) + "}."
			}
			out = v.tokenOf(e.Plague, "plague", format)
		case "L":
			out = v.legacyOf(e, format)
		case "N":
			out = v.format(e.N, format, e)
		case "warspan":
			out = warSpan(e)
		default:
			x, ok := e.P[name]
			if !ok {
				return m
			}
			switch format {
			case "why":
				out = v.whyText(e, name)
			case "term":
				out = v.termName(v.term(e, name))
			default:
				out = v.format(x, format, e)
			}
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
func (v *view) tokenOf(id int, kind, format string) string {
	switch format {
	case "":
		return "{" + kind + ":" + itoa(id) + "}"
	case "system":
		return v.systemLine(id)
	case "title":
		return "{title:" + itoa(id) + "}"
	case "word":
		return "{word:" + itoa(id) + "}"
	}
	return v.format(id, format, nil)
}

// legacyOf is the remain in an event as the line says it.
func (v *view) legacyOf(e *record.Event, format string) string {
	l := v.remain(e.Legacy)
	switch format {
	case "makers":
		return makersTok(l)
	case "at":
		// as it stood when the event captured it
		return v.describeAt(l, tables.conditions[e.Int("cond")].Key, e.Int("ships"), e.Bool("adrift"))
	}
	return v.legacyDesc(l)
}

// format is a parameter as a line says it: a number, a word for a
// number, a key's name, an id's token.
func (v *view) format(x any, format string, e *record.Event) string {
	switch format {
	case "civ", "star", "plague", "elder", "species", "makers":
		return "{" + format + ":" + itoa(toInt(x)) + "}"
	case "other":
		if toInt(x) == e.Subject {
			return "{civ:" + itoa(e.Object) + "}"
		}
		return "{civ:" + itoa(e.Subject) + "}"
	case "ships":
		return shipsWord(toInt(x))
	case "span":
		return span(Year(toInt(x)))
	case "systems":
		return systems(toInt(x))
	case "worlds":
		return worlds(toInt(x))
	case "number":
		return numberWord(toInt(x))
	case "ordinal":
		return ordinal(toInt(x))
	case "depth":
		return depthWord(toFloat(x))
	case "flow":
		return flowWord(flowKind(x.(string)))
	case "flowkind":
		return flowKind(x.(string)).String()
	case "category":
		return categoryPhrase(x.(string))
	case "percent":
		return percent(toFloat(x))
	case "share":
		return shareWord(toInt(x), e.Int("total"))
	case "node":
		return tech.Get(x.(string)).Name
	case "structure":
		return tech.Structures[x.(string)].Name
	case "miracle":
		return tables.miracles[x.(string)].Term
	case "object":
		return tables.miracles[x.(string)].Object
	case "filter":
		return filterName(x.(string))
	case "trait":
		return species.Get(x.(string)).Name
	case "traits":
		return species.DescribeTraits(toStrs(x))
	case "source":
		return v.sourceName(v.source(toInt(x)))
	case "named":
		return "{source:" + itoa(toInt(x)) + "}" // the object's proper name, a row of the names pass
	case "colony":
		return v.sp[toInt(x)].Flavour().Colony
	case "ship":
		return v.sp[toInt(x)].Flavour().Ship
	case "betrayal":
		return v.betrayalText(x.(string), e.Object)
	case "use":
		return v.useName(e)
	case "maker":
		return v.makerNameIf(v.remain(e.Legacy), x.(bool))
	case "drift":
		return v.driftText(e)
	case "portrait":
		return strings.Join(v.sp[toInt(x)].PortraitWith(e.Strs("powers")), "\n")
	case "elder_portrait":
		return v.elderPortrait(v.elder(toInt(x)))
	case "knower":
		return portraitText("knowers", x.(string))
	case "ender":
		return v.ageEnder(v.st.Ages[toInt(x)])
	case "feature":
		for _, f := range galaxy.Features {
			if f.Key == x.(string) {
				return f.Event
			}
		}
		return ""
	case "lifted":
		switch x.(string) {
		case "sea":
			return "put to sea"
		case "sky":
			return "look up, and see stars"
		}
		return "make fire"
	case "lines":
		return strings.Join(toStrs(x), "\n")
	}
	switch x := x.(type) {
	case string:
		return x
	case json.Number:
		if format == "" {
			return x.String() // an integer, as the simulation counted it
		}
		f, _ := x.Float64()
		return sprintf("%"+format, f)
	case bool:
		if x {
			return "true"
		}
		return "false"
	}
	return sprintf("%v", x)
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
var lines = map[record.Kind]string{
	record.KAgeDawn:     "The dawn of an age. Everywhere at once, things start to think.",
	record.KElderRose:   "Somewhere, {elder:elder_portrait} rises.",
	record.KAgeKnower:   "{line:knower}",
	record.KAgeWaned:    "The age wanes. Nothing new rises, and what remains dwindles. What is left is swept up by {age:ender}.",
	record.KElderLeft:   "It leaves {L}.",
	record.KLifeComplex: "Complex life flourishes at {T}.",
	record.KBurst:       "A gamma-ray burst near {T} sterilises {killed} living worlds within {radius:.0f} ly.",
	record.KSupernova:   "{T} goes supernova. {killed} living worlds within 30 ly are sterilised.",
	record.KStarSwelled: "{T} swells and dies, and the life on its worlds with it.",
	record.KReason:      "[the {S}, {what}: {why:why}]",
	record.KQuarry:      "The hunt of the {S} has a quarry now: the {O}.",
	record.KVeiled:      "The {S} forget the {O}, and this time there is no learning them again.",
	record.KLifted:      "The {S} of {T} {need:lifted}, to the bafflement of home. It never comes naturally to them.",

	record.FEnslaved:         "The {O} are enslaved by the {S}. They keep {T} and little else.",
	record.KMasterGone:       "The {S}, who held the {O}, are gone. The question of freedom answers itself, one way or the other.",
	record.FUplift:           "The {S} raise the {O} from the beasts of {T}. They are {traits:traits}, and grateful, for now.",
	record.FBred:             "The {S} remake the {O} into the {into:civ}: {traits:traits}.",
	record.KDebug:            "{text}",
	record.KSystem:           "{T:system}",
	record.KPortrait:         "{species:portrait}",
	record.KBornMiracle:      "They are born to a miracle: {miracle:miracle}. What others will spend ages reaching for, they have from the first.",
	record.KBornFailing:      "Their sun is already failing. They were born under a dying star.",
	record.FShipLost:         "A {species:ship} of the {S} arrives at {T}, which every reading said was empty, and is never heard from again. The {O} were there.",
	record.FZenith:           "The {S} enter their zenith: {N} systems, and no rival in sight.",
	record.KReseated:         "What is left of the {S} gathers on {T}. It is home now.",
	record.FFall:             "The {S} {cause:why}. What remains of them lives on {T} under {S:title}. Once they held {peak:systems}.",
	record.KOutgrown:         "The {S} are gone. What they built at {T} thinks on without them, and calls itself the {O}: {traits:traits}.",
	record.KForesaw:          "The {S} see {filter:filter} coming and step around it.",
	record.KStarDead:         "{T} dies. Its worlds freeze.",
	record.KFleetCaught:      "A fleet of the {S} at {T} is caught in it and is gone.",
	record.KCannotLeave:      "The {S} cannot leave {T}; they are it.",
	record.FLeftStar:         "The {S} leave {T} to {why:why}. {to:star} is home now, and always a little less than the one before.",
	record.FDoom:             "The sun of the {S} is failing. {T} grows harsher with every century. They have, perhaps, {endure} thousand years.",
	record.KStarKept:         "The {S} reach into {T} and hold it together. Their sun will fail, but not yet.",
	record.KEndured:          "The {S} endure under the failing sun of {T} until they cannot. The last of them die looking up.",
	record.KCentreFlared:     "The heart of the galaxy flares. For a century the sky is white, and every world in the field turns its face away.",
	record.KSkyFeature:       "{feature:feature}",
	record.KStayedHome:       "The {S} look hard at the {O}, and stay home.",
	record.FAppeared:         "Another of the {S} is at {T}. Nothing was seen to cross.",
	record.FTithed:           "Something is taken from every harvest of the {O} within reach of the {S}. Nobody agreed to it, and nothing can be found to refuse.",
	record.KMirrored:         "The {S} speak to the {O}, and what answers is their own voice, older than they are. A cult of the signal grows among them and is never quite rooted out.",
	record.KSlept:            "The {S} go still. There is nothing left they want, and nothing near them moves. They sleep.",
	record.KTakenDear:        "The {S} take {T}, and lose half their fleet doing it.",
	record.KLeftToGuns:       "The ships over {T} withdraw and leave it to its guns.",
	record.KHeldBehindGuns:   "The ships of the {S} hold their ground over {T} behind its guns.",
	record.KGunsSilent:       "The guns over {T} fall silent.",
	record.KFleetBroken:      "The {S} break a fleet of the {O} at {T}.",
	record.KFellBack:         "The fleet of the {S} falls back from {T}, and comes again.",
	record.KSightMarked:      "The Sight shows the {S} something at {T} that nobody made in this age. They mean to go and see.",
	record.KLurkerSeen:       "Surveyors of the {S} find {T} held by something that is not a people as they know one, and do not go closer.",
	record.KFirstSurvey:      "The {S} send their first surveyors out: a ship of a few, bound for {T}, to see what the stars hold.",
	record.KPicket:           "The {S} send a ship to {T} to sit and watch the sky toward the {O}.",
	record.KRarityPassed:     "{^source:source} passes from the {S} to the {O}, as agreed.",
	record.KSightingSold:     "The {S} sell the coming of the {owner:civ}'s fleet to the {O}, for {years:span} of {res:flowkind}.",
	record.KOfferRefused:     "The {S} ask the {O} for {ask:term}, and are refused.",
	record.FStrikeBought:     "The {S} pay the {seller:civ} {pay:term} to send {ask:term}.",
	record.KTeachAgreed:      "The {S} agree to teach the {O} {node:node}, for {pay:term}.",
	record.KBrokerAgreed:     "The {S}, who know both, will speak for the {O} to the {target:civ}, for {pay:term}.",
	record.KWorkDone:         "The {S}'s work for the {O} is done.",
	record.KTermDone:         "The {S}'s term at {T} is done, and the {O}'s fleet goes home.",
	record.KHireEnded:        "The {S}, their hire ended, leave the war.",
	record.FBoughtOff:        "The {S}, paid by the {O} to hold {T}, are paid more by the {buyer:civ}, and sell it.",
	record.KSellsword:        "The {S} live by their fleet now. Others call them sellswords.",
	record.FTaught:           "The {S} teach the {O} {node:node}, for {pay:term}.",
	record.FTribute:          "The {S} yield to the {O} and pay tribute in {res:flow} for {for:span}, after {warspan}.",
	record.KHanded:           "The {S} hand {T} to the {O}, as agreed.",
	record.KBurnedForPay:     "The {S} burn the {work:structure} at {T}, as they were paid to, and go.",
	record.KFindLeap:         "The {S}, who are wise, look hard at it and count what it would take of them, and do not try.",
	record.FSealed:           "The {S} seal it, and post a watch, and the watch holds.",
	record.KSealFailed:       "The {S} seal it. Someone opens it.",
	record.KSignalPlague:     "The {S} hear the transmitter at {T}. Something comes down the signal with it and begins to move through their minds.{Q:called}",
	record.KNewSignal:        "From {T} a new signal goes out, in the voice of the {S}.",
	record.FHunt:             "The ledger of the {S} shows a hole around {T}: {losses} losses inside {radius:.0f} light years, and nothing in any record to say what took them. The council declares a hunt on the region.",
	record.KHuntOn:           "The {S} are at war with something they can no longer name. What they have is the ledger, and the ledger says {T}.",
	record.KHuntEmpty:        "The hunt of the {S} finds nothing at {T}, and nothing, and nothing. Whatever was there is not, and the ledger is closed.",
	record.KDriftCured:       "What {Q} lived in is gone from under it: the {S} changed, and it did not change with them.",
	record.KReliefSent:       "The {S} send {ships:ships} to stand with the {O} at {T}, {away:span} away.",
	record.KArrivedLate:      "The fleet of the {S} arrives at {T} to find the war over.",
	record.FRelief:           "A fleet of the {S} arrives at {T} to stand with the {O}.",
	record.KFleetWasted:      "The fleet of the {S} wastes away at {T}, far from anything it could live on.",
	record.KFleetStayed:      "The fleet of the {S} never comes home. At {T} its captains rule as their own people.",
	record.KInterceptSent:    "The {S} send {ships:ships} from {T} to meet the fleet of the {O} in the dark.",
	record.KSalvaged:         "The {S} crew what will fly of it: {ships:ships}, turned for {T}.",
	record.KWentDark:         "The {S} let {use:use} go dark to keep {kept:category} fed.",
	record.FWant:             "The {S} have gone without for a hundred thousand years. They call them the lean years.",
	record.KGarrisoned:       "The {S} send {ships:ships} to hold {T}.",
	record.KGathered:         "The {S} gather their ships at {T}.",
	record.KGridRebuilt:      "The {S} rebuild the grid over {T}.",
	record.KClaimForgot:      "Among the {S} the sundering has become a story told to children. Nobody speaks of the old realm as theirs any more.",
	record.KRemade:           "The {S} are gone. What they made of themselves holds their worlds and calls itself the {O}: {traits:traits}.",
	record.KMovedOn:          "The fleets of the {S} move on, to {T}.",
	record.KPocketDimmed:     "The pocket star dims as the {S} move it. It gives {gives:.0f} now.",
	record.KSingularityLoose: "The captive singularity at {T} gets loose in the taking.",
	record.FVacuumHole:       "The hole in the vacuum at {T} has grown past holding. The star begins to go out.",
	record.FRise:             "What the {O} grew for the table at {T} has been thinking for a long time. It rises, and calls itself the {S}: {traits:traits}.",
	record.KCutting:          "The {S} give the {O} a cutting of {source:source}. It takes.",
	record.FRenaissance:      "The {S} grow old and tired, and then, unexpectedly, young again. A renaissance.",
	record.KSet:              "The {S} stop changing. Every year is like the last. It works, for a while.",
	record.KMadeToThink:      "The {S} made {Q} to think, and it does. It answers to them, for now.",
	record.FWoke:             "Something in {Q} has begun to think. At {T} it takes the {O} for its own, and calls itself the {S}.",
	record.KBornRidden:       "The {S} have never known a time before {Q}. They grew up ridden.",
	record.KRidden:           "{^Q} takes {T}. World by world, the {O} were ridden, and now they are.",
	record.FWildfire:         "{^Q} is everywhere now.",
	record.KRiderGone:        "The {S}, who rode the {O}, are gone. There is nothing left in them to fight.",
	record.KSealedDoors:      "The {S} seal every door against {Q}.",
	record.KQuarantineCreed:  "Twenty thousand years behind sealed doors, and the {S} no longer remember how to open them. It is a creed now.",
	record.KCult:             "At {T} the believers in {Q} declare themselves a people: the {O}.",
	record.KReservoirWoke:    "The {S} come to {T} and wake {Q} in its dead cities.",
	record.KLeftBehind:       "{^source:source} is left behind at {T}.",
	record.KWentDown:         "{^source:source} went down with the fleet, and lies at {T} for whoever finds it.",
	record.KPursuit:          "The {S} turn everything they have toward {node:node}. It will take ages, and it may not come.",
	record.KFellShort:        "The {S} come close to {node:node} and fall short. The work of ages goes for nothing.",
	record.FCycle:            "The {S} find their place in the turn: the age dawned {dawned:.0f} million years ago, the galaxy is {fertility:percent} as fertile as it was then, and the next dawn is {next:.0f} million years away. They will not see it.",
	record.FStars:            "The {S} reach the stars.",
	record.KReplicating:      "At {T} the {S} have begun to make more of themselves out of what is there.",
	record.KFirstShip:        "The yards at {T} launch their first ship for the {S}.",
	record.KLaidUp:           "The {S} lay up {ships:ships} at {T}: the ships stay where they are, and nothing keeps them.",
	record.KManned:           "The {S} man the ships at {T} again.",
	record.KScoutSeen:        "The {S} see a ship of the {O} in their sky at {T}, looking, and do not forget it.",
	record.FSlight:           "The {O} trade with the {partner:civ}, and take the {S}'s war on them as a wrong done to themselves.",
	record.FReclaimed:        "The {S} call {T} restored to the realm.",
	record.KRealmWhole:       "The {S} hold every world the {O} held. There is nothing left to claim, and they are the {S} that hold what the {O} held.",
	record.KKinMet:           "The {S} and the {O}, both of the line of the {line:civ}, find each other again.",
	record.FSevered:          "{T} is too far from {seat:star} for one mind to hold. What is there is the {O} now: of one blood with the {S}, and no longer one of them.",
	record.KPortsOpened:      "The {S} open their ports to the {O} again.",
	record.FEmbargo:          "The {S} have what the {O} want, and will not send it. The {O} call it an embargo.",
	record.KTired:            "The {S} tire of the {O}, who take and send nothing back, and the trade between them ends.",
	record.FCutOff:           "The {S} go dark when the {O} stop sending, with {why:why}.",
	record.KWeaponBurned:     "The {S} burn {Q}, having nobody left to give it to.",
	record.KWeaponLeaked:     "The programme that keeps {Q} is not kept well enough.",
	record.FWaking:           "The {S} wake. {seat:star} and everything near it is theirs, and the {O} are on it.",
	record.KStillAgain:       "The {S} go still again.",
	record.KRiddenWar:        "The {S} are still there, and still themselves, mostly. They do what the {O} want now, and what they knew, the {O} know.",
	record.FUnmade:           "The {S} unmake {T}, a world of the {O}. It is not there any more.",
	record.KSlowTrade:        "Slow messages cross the dark between the {S} and the {O} for generations, and then trade.",
	record.FBrokered:         "The {S}, who know both, speak for the {O} to the {to:civ}.",
	record.KUnfathomed:       "The {S} forget how to speak to the {O}.",
	record.KWaning:           "The age is waning. Few still rise, and those that stand are old.",
}

// lineFns pick a template by the event's parameters.
var lineFns = map[record.Kind]func(v *view, e *record.Event) string{
	record.KLifeArose: func(v *view, e *record.Event) string {
		st := v.g.Stars[e.Star]
		st.Class = e.Str("class")[0]
		return "Life arises on the worlds of " + v.g.DescribeStar(&st, e.Star) + "."
	},
	record.KElderFell: func(v *view, e *record.Event) string {
		switch {
		case e.Bool("late"):
			return "It ends, the last of its age, long after the others."
		case e.Bool("remembered"):
			return "It is gone. Its works remain."
		}
		return "It ends."
	},
	record.KBirthright: func(v *view, e *record.Event) string {
		var parts []string
		for _, k := range e.Strs("blocks") {
			parts = append(parts, aptitudeText(k))
		}
		if cheap := nodeNames(e.Strs("cheap")); len(cheap) > 0 {
			parts = append(parts, "They take to "+list(cheap)+" as if born to it.")
		}
		if costly := nodeNames(e.Strs("costly")); len(costly) > 0 {
			parts = append(parts, strings.ToUpper(list(costly)[:1])+list(costly)[1:]+" will come hard to them.")
		}
		return strings.Join(parts, " ")
	},
	record.FMet: func(v *view, e *record.Event) string {
		way := e.Str("way")
		switch e.Str("how") {
		case "signal":
			if !e.Bool("told") {
				return ""
			}
			d := e.Float("distance")
			return sprintf("The {S} hear the {O} across %.0f light years: a signal, then a conversation %.0f years to the answer. Neither can reach the other yet.", d, 2*d)
		case "noticed":
			switch way {
			case "unseen":
				if e.Bool("hidden") {
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
			case e.Star >= 0 && e.Bool("heard"):
				return "Ships of the {S} come upon the {O} at {T}, and the long conversation across the dark has a face at last."
			case e.Star >= 0:
				return "Ships of the {S} come upon the {O} at {T}."
			case e.Bool("heard"):
				return "The " + a + " and the " + b + ", who have heard each other for a long time, at last meet in the flesh."
			case e.Bool("watched"):
				return "The " + b + ", long watched from orbit, look up and find the " + a + "."
			}
			return "The " + a + " and the " + b + " find each other."
		}
		return ""
	},
	record.FArise: func(v *view, e *record.Event) string {
		if e.Has("host") {
			return sprintf("The {S} %s the {civ:%d}, at %s. They are {traits:traits}.", v.species(v.civ(e.Subject)).Arising(), e.Int("host"), v.g.Sys[e.Star].HomeName(star(e.Star)))
		}
		if e.Bool("made") {
			return ""
		}
		sp := v.species(v.civ(e.Subject))
		st := v.g.Stars[e.Star]
		st.Class = e.Str("class")[0]
		prior := ""
		if p := e.Int("prior"); p >= 0 {
			prior = sprintf(", among the ruins of the {civ:%d}", p)
		}
		they := "They are {traits:traits}."
		if e.Bool("shared") {
			they = "They are a people of the {species:species}."
		}
		return sprintf("The {S} %s %s, %s, around %s%s, %.0f ly from %s. %s",
			sp.Arising(), v.g.Sys[e.Star].HomeName(star(e.Star)), sp.World.Desc, v.starDetail(&st, e.Star), prior, v.g.FromCentre(e.Star), v.g.Anchor(), they)
	},
	record.KArrivalLost: func(v *view, e *record.Event) string {
		switch e.Str("why") {
		case "held":
			return "A {species:ship} of the {S} arrives at {T} to find the {O} already there."
		case "taken":
			return "A {species:ship} of the {S} arrives at {T} to find it already taken. It is never heard from again."
		}
		return "A {species:ship} of the {S} reaches {T} on a guess and finds nothing there it can live on. What it learned is sent home. The ship is not."
	},
	record.FSettle: func(v *view, e *record.Event) string {
		if e.Bool("first") {
			return "The {S} settle {T}, their first {species:colony} beyond {home:star}."
		}
		return "The {S} now hold {N} systems."
	},
	record.KBuilt: func(v *view, e *record.Event) string {
		return tech.Structures[e.Str("work")].Text
	},
	record.KHomeLost: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "was":
			return "The {S} were {T}, and {T} is gone. What they held elsewhere dies with it."
		case "queen":
			return "The queen of the {S} dies with {T}. A hive without its queen is only bodies, and the bodies stop."
		}
		return "The {S} have no queen to gather to when {T} is lost, and no seat. Every world of theirs is on its own."
	},
	record.FEnd: func(v *view, e *record.Event) string {
		switch {
		case e.Str("fate") != "extinct":
			return ""
		case e.Bool("remnant"):
			return "The last of the {S} are gone from {T}. They {cause:why}."
		}
		return "The {S} {cause:why}. They held {peak:systems} at their height."
	},
	record.FDarkAge: func(v *view, e *record.Event) string {
		if e.Int("lost") > 0 {
			return "The {S} {cause:why}. A dark age follows, and {depth:depth} of what they knew is forgotten. {lost} {species:colony}s go silent."
		}
		return "The {S} {cause:why}. A dark age follows, and {depth:depth} of what they knew is forgotten."
	},
	record.KFaced: func(v *view, e *record.Event) string {
		key := e.Str("filter") + "/" + e.Str("outcome")
		if way := e.Str("way"); way != "" {
			key += "/" + way
		}
		return facedLines[key]
	},
	record.KNamedItself: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "eats":
			return "What eats {T} calls itself the {S}, if it calls itself anything: {traits:traits}."
		case "growth":
			return "What is at {T} now is a growth that eats worlds, and it calls itself the {S}, if it calls itself anything: {traits:traits}."
		case "called":
			return "It is called the {S}, by those who have to call it something: {traits:traits}."
		case "mind":
			return "It calls itself the {S}: {traits:traits}."
		}
		return "It calls itself the {S}, if it calls itself anything: {traits:traits}."
	},
	record.KBlast: func(v *view, e *record.Event) string { return blastLines[e.Str("way")] },
	record.KWallStage: func(v *view, e *record.Event) string {
		switch e.Int("stage") {
		case 1:
			return "Something has changed in the field, and nobody in it can say what. Doors are used a little less carefully than they were."
		case 2:
			return "The wall between this and what is under it has worn thin. Things come through more easily now, for everyone, and no one knows to blame anyone."
		}
		return "The wall is torn. What leaks through no longer needs a door."
	},
	record.KLeak: func(v *view, e *record.Event) string {
		switch e.Str("mechanism") {
		case "shore":
			word := "it"
			if e.Bool("named") {
				word = "{S:word}"
			}
			return "Some of the {S} begin to see " + word + " as a place, with a shore and a weather. That is never good; it means something is coming through."
		case "sleeper":
			return "The wall is thin near {T} now, and something that slept there notices."
		}
		return "Something speaks from {T} in no language, in a voice that did not cross space to get there. It came through."
	},
	record.KRoused: func(v *view, e *record.Event) string {
		if e.Object >= 0 {
			return "Something came too close, and the {S} wake."
		}
		return "The {S} wake."
	},
	record.FWord: func(v *view, e *record.Event) string {
		route := e.Str("route")
		t, ok := beneathNames[route]
		if !ok {
			return ""
		}
		if v.species(v.civ(e.Subject)).Voiceless() {
			return beneathDesc[route] + " The {S} have no word for it, having no words."
		}
		return sprintf(t, "{S}", "{S:word}")
	},
	record.FDeepened: func(v *view, e *record.Event) string {
		return capitalise(strings.ReplaceAll(species.PowerByKey(e.Str("power")).Line, "{S}", "the {S}"))
	},
	record.FDefeat: func(v *view, e *record.Event) string {
		if e.Str("way") == "withdrew" {
			return "The fleet of the {S} withdraws from {T} and turns for home."
		}
		return "The fleet of the {S} is broken at {T}."
	},
	record.KSightTurned: func(v *view, e *record.Event) string {
		if e.Bool("outward") {
			return "The {S} turn their precognition outward, to the stars nobody has visited."
		}
		return "The {S} turn their precognition back to their own borders."
	},
	record.FSurveyLost: func(v *view, e *record.Event) string {
		if e.Bool("unseen") {
			return "The surveyors of the {S} do not come back from {T}. What they sent before the end says the star is empty. The {O} are there."
		}
		return "The surveyors of the {S} do not come back from {T}. What they sent before the end says enough: the {O} are there."
	},
	record.FBetrayal: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "broke":
			return "The {S} {shape:betrayal}, and the {O} remember it."
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
	record.FHire: func(v *view, e *record.Event) string {
		against := ""
		if e.Int("against") >= 0 {
			against = " against the {against:civ}"
		}
		return "The {S} take the {O}'s {pay:term} to hold {T}" + against + "."
	},
	record.KBargain: func(v *view, e *record.Event) string {
		if e.Bool("first") {
			return "The {S} and the {O} strike a bargain: {ask:term} for {pay:term}. It is the first of many."
		}
		return "The {S} and the {O} strike a bargain: {ask:term} for {pay:term}."
	},
	record.FFind: func(v *view, e *record.Event) string {
		who := "The {S}"
		if e.Str("how") == "survey" {
			who = "Surveyors of the {S}"
		}
		where := "beneath their own cities on {T}"
		if !e.Bool("own") {
			where = "at {T}"
		}
		if e.Str("how") == "settle" {
			where += ", under the feet of the first colonists"
		}
		switch {
		case e.Int("elder") >= 0:
			return who + " find {L:at} " + where + ". It is older than their sun. They call its makers {elder:elder}."
		case e.Int("kin") == 2:
			return who + " find {L:at} " + where + ". It is their own, from before the dark age. Something in them remembers it."
		case e.Int("kin") == 1:
			return who + " find {L:at} " + where + ". The hands that made it were like their hands."
		case !e.Bool("known"):
			return who + " find {L:at} " + where + ". They do not know who the {O} were. They call them {L:makers}."
		case e.Float("ago") < 0.1:
			return who + " find {L:at} " + where + ", not long after the {O} left it."
		}
		return who + " find {L:at} " + where + ", {ago:.1f} million years after the {O} left it."
	},
	record.FMastered: func(v *view, e *record.Event) string {
		switch e.Str("way") {
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
	record.KMasterFailed: func(v *view, e *record.Event) string {
		if e.Str("way") == "ruin" {
			return "The {S} pick over it for centuries and learn nothing. There is not enough left."
		}
		return "The {S} try to understand it and cannot."
	},
	record.KWieldFailed: func(v *view, e *record.Event) string {
		switch e.Str("way") {
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
	record.KWielded: func(v *view, e *record.Event) string {
		switch e.Str("way") {
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
	record.FUnleashed: func(v *view, e *record.Event) string {
		switch e.Str("way") {
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
	record.KWieldedDropped: func(v *view, e *record.Event) string {
		if e.Str("way") == "buried" {
			return "What the {S} wielded of {known:maker} lies where they left it, on {T}."
		}
		return "What the {S} wielded of {known:maker} is broken, and nobody knows how to mend it."
	},
	record.FDrifted: func(v *view, e *record.Event) string {
		if !e.Bool("told") {
			return ""
		}
		return "The {S} have changed again: {gained:drift}. Whoever knew them knew something else."
	},
	record.KFleetSent: func(v *view, e *record.Event) string {
		if e.Int("hunt") >= 0 {
			return "The {S} send {ships:share} of their ships into the hole around {hunt:star}, where the ledger says something is: a fleet of {ships:ships} bound for {T}, {away:span} away."
		}
		return "The {S} send {ships:share} of their ships against the {O}: a fleet of {ships:ships} bound for {T}, {away:span} away."
	},
	record.KFleetArrived: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "hunt":
			return "The hunting fleet of the {S} arrives at {T} and finds nothing there, which is what it was told it would find."
		case "hired":
			return "A fleet of the {S} arrives at {T} to hold it for the {O}, as paid."
		}
		return "The fleet of the {S} arrives at {T}, {out:span} after it set out."
	},
	record.FIntercept: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "broken":
			return "The {S} meet the fleet of the {O} between {T} and {far:star}, and break it. Nothing of it arrives."
		case "turned":
			return "The {S} meet the fleet of the {O} between {T} and {far:star}, and turn it back."
		}
		weaker := ""
		if lost := e.Int("lost"); lost > 0 {
			weaker = ", " + shareWord(lost, lost+e.Int("left")) + " weaker"
		}
		return "The fleet of the {S}, met in the dark between {T} and {far:star} by the {O}, goes on" + weaker + "."
	},
	record.KTakenOver: func(v *view, e *record.Event) string {
		if e.Bool("own") {
			return "The {S} return to {T} and put their own old works there back to use."
		}
		return "The {S} find {L:at} at {T}, and put it back to work."
	},
	record.KJudged: func(v *view, e *record.Event) string {
		f := v.event(e.Int("about"))
		switch e.Str("verdict") {
		case "deed":
			return "The {S} hear that the {O} " + v.deedOf(f) + ", and count it a deed."
		case "nothing":
			return "The {S} hear that the {O} " + v.deedOf(f) + ", and count it no crime."
		}
		return "The {S} hear what the {O} did, and call it a crime."
	},
	record.KReadWalls: func(v *view, e *record.Event) string {
		if e.Bool("own") {
			return "In what they left at {T} the {S} read their own story in their own words, and remember."
		}
		return "What the {S} read in {L} at {T} is the telling of {known:maker}, and they have no other."
	},
	record.KBlamed: func(v *view, e *record.Event) string {
		return "The {S} now tell that it was the {O} who " + v.blameOf(v.civ(e.Subject), v.event(e.Int("about"))) + ". It was not."
	},
	record.KMyth: func(v *view, e *record.Event) string {
		return "Among the {S}, " + v.mythOf(v.civ(e.Subject), v.event(e.Int("about")), e.Bool("nameless")) + " has become a story told to children."
	},
	record.KScapegoat: func(v *view, e *record.Event) string {
		c := v.civ(e.Subject)
		var deeds []string
		facts := e.Ints("facts")
		for _, id := range facts {
			if d := v.blameOf(c, v.event(id)); len(deeds) < 3 && !slices.Contains(deeds, d) {
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
	record.KMorality: func(v *view, e *record.Event) string {
		var m record.Morality
		e.Obj("morality", &m)
		switch e.Str("way") {
		case "branch":
			return "The {S} have gone their own way in what they count as wrong. " + moralityPortrait(m)
		case "church":
			return "The church of the {S} teaches what is good, and it is one thing. " + moralityPortrait(m)
		case "taught":
			return "The {S} were taught what the {O} call wrong. " + moralityPortrait(m)
		case "machine":
			return "What the {S} hold good is what their makers were doing when they were outgrown. " + moralityPortrait(m)
		}
		return moralityPortrait(m)
	},
	record.FExodus: func(v *view, e *record.Event) string {
		switch {
		case e.Str("way") == "fled":
			return "The {S} {why:why}. What got away is a fleet at {base:star}, and it is all of them now."
		case e.Str("why") == "":
			return "The {S} take to the sky. {T} is left empty behind them, and everything they are is in the fleets now."
		}
		return "The {S} take to the sky rather than {why:why}. {T} is left empty behind them."
	},
	record.KCarried: func(v *view, e *record.Event) string {
		if e.Str("way") == "brought" {
			return "The fleets of the {S} bring the {O} {node:node}."
		}
		return "The {S} learn {node:node} from the {O}, and carry it on."
	},
	record.FStripped: func(v *view, e *record.Event) string {
		if e.Bool("home") {
			return "The horde of the {S} strips {T}, the home of the {O}, of its ships and its people."
		}
		return "The {S} strip {T} of its ships and its people. The horde grows."
	},
	record.FRest: func(v *view, e *record.Event) string {
		if e.Bool("nomad") {
			return "The {S} come to rest at {T}, and are nomads no longer. It was {why:why} that did it."
		}
		return "The {S}, refugees no longer, settle {T}. It is home now."
	},
	record.KObjectMade: func(v *view, e *record.Event) string {
		switch e.Str("how") {
		case "found":
			return "The {S} put it to use. It is {object:object}: {source:source}, and it feeds them a swarm's worth."
		case "born":
			return "The {S} have kept {object:object} since before they had a name for it: {source:source}. It feeds them, and it is theirs to carry."
		case "wield":
			return "The {S} put it to use. It is {object:object}, and it feeds them a swarm's worth."
		}
		if e.Str("object") == "ember" {
			return "The {S} kindle an exotic energy source, {source:source}, and call it {source:named}. It gives what a swarm gives, and it is theirs to carry."
		}
		return "The {S} grow a self-sustaining food organism, {source:source}, and call it {source:named}. It feeds them, and it does not stop."
	},
	record.FManna: func(v *view, e *record.Event) string {
		if e.Str("way") == "eat" {
			return "It thinks. The {S} eat it anyway."
		}
		return ""
	},
	record.KThrough: func(v *view, e *record.Event) string {
		if e.Bool("burned") {
			return "Something comes through the energy source at {T}. The {S} burn it off."
		}
		return "Something comes through the energy source at {T}, and what lived there is lost to it."
	},
	record.FFreed: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "unthink":
			return "The {S} learn to unthink the {O}. What was in their heads is gone, and they are free, and never again quite trust a new idea."
		case "drug":
			return "The {S} find a drug that kills what rides them. They are free, and careful about their blood ever after."
		}
		return ""
	},
	record.FPlague: func(v *view, e *record.Event) string {
		road := e.Str("road")
		switch {
		case road == "born" && v.plague(e.Plague).Kind == "memetic":
			return "Something moves through the minds of the {S}.{Q:called}"
		case road == "born":
			return "Something moves through the worlds of the {S}.{Q:called}"
		case e.Object >= 0 && roadPhrases[road] != "":
			return "{^Q} comes to the {S} " + roadPhrases[road] + "."
		}
		return ""
	},
	record.FCured: func(v *view, e *record.Event) string {
		if e.Bool("unknowing") {
			return "The {S} are rid of {Q}, and never know it had begun to think."
		}
		return "The {S} are rid of {Q}."
	},
	record.FPlagueWorld: func(v *view, e *record.Event) string {
		if e.Bool("home") {
			return ""
		}
		return "{^Q} empties {T}, a {species:colony} of the {S}. The cities are sealed and left."
	},
	record.FBelieved: func(v *view, e *record.Event) string {
		switch {
		case e.Bool("cult"):
			return ""
		case e.Bool("home"):
			return "{^Q} takes {T}. The {S} listen to it and are changed by it, and nobody there answers to anyone now."
		}
		return "{^Q} takes {T}, a {species:colony} of the {S}. Nobody there answers to them any more."
	},
	record.FRefused: func(v *view, e *record.Event) string {
		if e.Bool("traded") {
			return "The {S} stop the trade with the {O} for fear of {Q}, and hear nothing from them."
		}
		return "The {S} close their ears to the {O} for fear of {Q}."
	},
	record.KWallsPlague: func(v *view, e *record.Event) string {
		maker := "someone"
		if e.Int("maker") >= 0 {
			maker = "the {maker:civ}"
		}
		return "On the walls of " + maker + " {Q} is written, and the {S} read it."
	},
	record.KRelicWoke: func(v *view, e *record.Event) string {
		if e.Bool("took") {
			return "It does what it was made to do, to the {S}.{Q:called}"
		}
		return "It does what it was made to do, and finds nothing in the {S} to do it to."
	},
	record.KPactRefused: func(v *view, e *record.Event) string {
		if e.Int("against") >= 0 {
			return "The {S} ask the {O} for a pact against the {against:civ}, and are refused."
		}
		return "The {S} ask the {O} for a pact, and are refused."
	},
	record.FPact: func(v *view, e *record.Event) string {
		against := "whoever comes"
		if e.Int("against") >= 0 {
			against = "the {against:civ}"
		}
		return "The {S} and the {O} swear a pact of {pact} against " + against + "."
	},
	record.FHarness: func(v *view, e *record.Event) string {
		if e.Has("source") {
			return harnessLines[v.source(e.Int("source")).Key]
		}
		return ""
	},
	record.KRarityHad: func(v *view, e *record.Event) string {
		s := v.source(e.Int("source"))
		if e.Bool("via") {
			return "The {S} have the use of {source:source}, by the grace of the {O}."
		}
		return rarityLines[s.Key]
	},
	record.KCarriedOff: func(v *view, e *record.Event) string {
		if e.Int("fleet") >= 0 {
			return "The {S} carry off {source:source} with the fleet."
		}
		return "The {S} carry off {source:source} to {T}."
	},
	record.KNodeLearned: func(v *view, e *record.Event) string {
		n := tech.Get(e.Str("node"))
		if n.Milestone && n.Text != "" {
			return n.Text
		}
		return ""
	},
	record.KFleetSeen: func(v *view, e *record.Event) string {
		switch e.Str("eye") {
		case "works":
			return "From the {eye_work:structure} at {eye_star:star} the {S} see the fleet of the {O} coming, {out:span} out."
		case "fleet":
			return "A fleet of the {S} in flight sees the fleet of the {O} coming toward {T}, {out:span} out."
		case "picket":
			return "The pickets of the {S} see the fleet of the {O} coming, {out:span} out."
		}
		return "The {S} see the fleet of the {O} coming, {out:span} out."
	},
	record.KUnrest: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "hive":
			return "The {S} cannot split; a hive has no factions. The pressure goes elsewhere."
		case "fleets":
			return "Unrest in the fleets of the {S}. It passes, this time."
		}
		return "Unrest among the {S} on {T}. It passes, this time."
	},
	record.FSundered: func(v *view, e *record.Event) string {
		if !e.Bool("first") {
			return ""
		}
		heirs := e.Ints("heirs")
		var ns []string
		for _, h := range heirs {
			ns = append(ns, "the {civ:"+itoa(h)+"}")
		}
		seatLine := ""
		if seat := e.Int("seat"); seat >= 0 {
			seatLine = " The {civ:" + itoa(seat) + "} hold the old seat."
		}
		return "The {S} tear themselves in " + numberWord(len(heirs)) + ": " + listOf(ns) + ", each the true {S} by its own telling, each holding the others traitors." + seatLine
	},
	record.FShattered: func(v *view, e *record.Event) string {
		if !e.Bool("first") {
			return ""
		}
		var ns []string
		stars := e.Ints("stars")
		for i, h := range e.Ints("shards") {
			ns = append(ns, "the {civ:"+itoa(h)+"} on {star:"+itoa(stars[i])+"}")
		}
		return "The {S} forget how to reach the stars. On " + numberWord(e.N) + " worlds " + numberWord(e.N) + " peoples wake up alone: " + listOf(ns) + "."
	},
	record.KWeaponMade: func(v *view, e *record.Event) string {
		line := "The {S} breed a sickness for the {O}, and call it {Q} among themselves."
		if e.Bool("memetic") {
			line = "The {S} shape an idea to break the {O}, and call it {Q} among themselves."
		}
		if e.Bool("conscious") {
			line += "\\nThey have made it to think."
		}
		return line
	},
	record.KBreakout: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "nothing":
			return "It gets out, and finds nothing in the {S} to be in."
		case "carrier":
			return "It gets out. {^Q} does nothing to the {S}, who made it; it goes with everything they send."
		}
		return "It gets out. {^Q} is loose among the {S}, who made it."
	},
	record.FPoisoned: func(v *view, e *record.Event) string {
		if e.Bool("took") {
			return "The {O} find {Q} was hidden in what the {S} sent them, and made for them."
		}
		return "The {O} catch the {S} trying to hide {Q} in what they sent. They take nothing from them again."
	},
	record.FDemand: func(v *view, e *record.Event) string {
		line := "The {S} make it known to the {O} that {T} is theirs, and the {O} are to leave it. The {O} "
		switch {
		case e.Str("outcome") == "left":
			return line + "go, and do not say why."
		case e.Bool("home"):
			return line + "have nowhere to go; {seat:star} is home."
		}
		return line + "stay."
	},
	record.FWar: func(v *view, e *record.Event) string {
		switch {
		case e.Bool("hunt"):
			return ""
		case e.Bool("unseen"):
			return "The {S} declare war on the {O}, over {cause:why}. The {O} will never know by whom."
		case e.Int("nth") > 1:
			return "The {S} go to war with the {O} again, the {nth:ordinal} time, over {cause:why}."
		}
		return "The {S} declare war on the {O}, over {cause:why}."
	},
	record.FBurned: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "unmade":
			return "The {S} unmake {T}, a {species:colony} of the {O}. There is nothing left to glass."
		case "glassed":
			if !e.Bool("told") {
				return ""
			}
			return "The {S} glass {T}, a {species:colony} of the {O}."
		}
		return "The {S} burn {T} to be rid of what the {O} put there."
	},
	record.FTaken: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "host":
			return "{^Q} takes {T}, a {species:colony} of the {O}. It is a host-world of the {S} now."
		case "host_war":
			return "The {O} of {T} are riders now. The {S} wear them."
		case "stripped":
			return "The {S} take {T} from the {O} and strip it. Nothing that was there is left; what is there now is more of the {S}."
		case "overrun":
			if !e.Bool("told") {
				return ""
			}
			return "The {S} overrun {T}. Where the {O} were there is a nest."
		}
		switch {
		case !e.Bool("told"):
			return ""
		case e.Bool("empty_sky"):
			return "The {S} take {T} from the {O}. There was nothing in its sky."
		}
		return "The {S} take {T} from the {O}."
	},
	record.FHomeBroken: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "unmade":
			return "The {S} unmake {T}, homeworld of the {O}. It is not there any more."
		case "shattered":
			return "A relativistic strike from the {S} shatters {T}, homeworld of the {O}. They never surrendered."
		case "world":
			return "The {S} take {T}, and there is nothing to rule. The {O} were the world, and the world is dead."
		}
		return "The {S} take {T}, and the queen of the {O} with it. A hive without its queen is only bodies, and the bodies stop."
	},
	record.FScoured: func(v *view, e *record.Event) string {
		switch e.Str("way") {
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
	record.FYield: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "spared":
			return "The {S} defeat the {O} and, having no use for a conquest, leave them be."
		case "ceded":
			return "The {O} yield to the {S} and cede {N:worlds}, after {warspan}."
		}
		return "The {O} yield to the {S}, who take {N:worlds} and want nothing more, after {warspan}."
	},
	record.KDefencesBroken: func(v *view, e *record.Event) string {
		if e.Str("outcome") == "vassal" {
			return "The {S} break the last defences of {T}, and the {O} bend the knee. They are vassals now."
		}
		return "The {S} break the last defences of {T}."
	},
	record.KWarEnded: func(v *view, e *record.Event) string {
		switch e.Str("way") {
		case "hunt":
			return "The hunt of the {S} ends, the will for it spent, after {warspan}. Nothing was found that could be named, and the {O} go on being there."
		case "spent":
			return "The {S} and the {O} stop fighting, both sides spent, after {warspan}. Nothing is signed; there is nobody on one side to sign it."
		}
		return "The {S} and the {O} stop fighting, both sides tired of it, after {warspan}. Neither ever understood what the other wanted, and nothing is signed."
	},
	record.FPeace: func(v *view, e *record.Event) string {
		if e.Str("way") == "truce" {
			return "The {S} sue the {O} for a truce, which is all they know how to ask for, after {warspan}. The fighting stops; nothing is settled."
		}
		terms := "Neither side is sure who won."
		switch net := e.Int("net"); {
		case net > 0:
			terms = "The {S} keep what they took."
		case net < 0:
			terms = "The {O} keep what they took."
		}
		return "The {S} and the {O} make peace, {why:why}, after {warspan}. " + terms
	},
	record.KYielded: func(v *view, e *record.Event) string {
		if e.Str("outcome") == "vassal" {
			return "The {S} yield to the {O} and bend the knee, after {warspan}. They are vassals now."
		}
		return "The {S} yield to the {O}, after {warspan}."
	},
	record.FFathomed: func(v *view, e *record.Event) string {
		how, mutual, wars := e.Str("how"), e.Bool("mutual"), e.Int("wars")
		ago := span(e.YearOf("since"))
		switch {
		case how == "kin" && mutual && !e.Bool("kin"):
			return "The {S} and the {O}, of one blood, understand each other at once."
		case how == "kin", how == "meeting":
			return ""
		case how == "chorus":
			return "The {S}, whose thought takes root in any mind, understand the {O} at once."
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
	"unmaking/declined":    "The {S} turn their annihilation on something too close.",
	"chorus/overcome":      "The {S} think one thought and remain many people. It can be done.",
	"chorus/scarred":       "The {S} think one thought, and it is the same thought every year after. Nothing new is ever said.",
	"chorus/declined":      "The thought of the {S} gets loose. From {T} it goes out to whoever will hear it.",
	"sight/overcome":       "The {S} see how it ends, and go on anyway.",
	"sight/scarred":        "The {S} see how it ends. A fatalism settles on them that never lifts.",
	"containment/scarred":  "Something gets out of the vial among the {S} and is burned with the building. They keep the craft and never use it, and go no further with it.",
	"waking/overcome":      "The {S} hide in the deep places while the {waker:civ} passes over them. It does not notice.",
	"waking/scarred":       "The {S} lose every world near the {waker:civ} but their own.",
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
	"ember":           "The energy source at {T} burns, once, and takes the star with it.",
	"manna":           "The food organism at {T} gets out, and eats.",
	"manna_grown":     "What the {S} grew for the table at {T} gets out, and eats.",
	"unmaking_test":   "The first test of the {T}' Unmaking takes a world with it.",
	"unmaking_inward": "Everything around {T} stops being matter for a while.",
}

// harnessLines are what a people says when it first harnesses a source
// kind worth a line, by the source's key.
var harnessLines = map[string]string{
	"belt":            "The {S} begin to work {source:source}: ice and iron, by the shipload.",
	"brown_companion": "The {S} burn {source:source} for fuel: a star that never lit, lit at last.",
	"giant":           "The {S} skim {source:source} for fuel. The small suns of their fusion plants never go out now.",
	"terraformed":     "The {S} bring {source:source} to life. It feeds them.",
	"heavy":           "The {S} dig {source:source} and find it rich in what splits.",
	"nebula":          "The {S} grow their food in the gas of {source:source} itself.",
	"doomed_giant":    "The {S} catch the light of {source:source}. It will not shine long, and they know it.",
	"comets":          "The {S} harvest {source:source}, and eat ice older than their sun.",
}

// rarityLines are what a people says when it first has a rarity, by key.
var rarityLines = map[string]string{
	"horizon":         "The {S} hold {source:source} now. What falls in comes out as understanding: the deep physics come cheap to them.",
	"beam":            "The {S} live under {source:source}. Its flares light their sky, and they learn from it how to unmake.",
	"heavy_star":      "The {S} hold {source:source}. Its skin is metal a mile deep, and its weight is a lesson in stellar weapons.",
	"beacon":          "The {S} hold a star within reach of {source:source}. They steer by its ticking, and their ships go further for it.",
	"diamond":         "The {S} hold {source:source}. It is a diamond the size of a world, and they are never done singing about it.",
	"colours":         "The {S} live in {source:source}. The sky is a painting, and the painting hides them.",
	"ash":             "The {S} hold a star in {source:source}. The matter there was made in a death, and some of it is not on any table.",
	"heart":           "The {S} hold a star within reach of {source:source}. Everything falls toward it, and so does thought.",
	"dwarf_companion": "The {S} hold {source:source}. It is a thing of great value that does nothing, which is what makes it valuable.",
	"dust":            "The {S} hold {source:source}. At dusk the whole sky is a ring, and they put it on their flags.",
	"moon":            "The {S} hold {source:source}. The tides and the calendar are its, and the songs.",
}

// roadPhrases say how a plague came, by its road; {O} is the giver.
var roadPhrases = map[string]string{
	"goods":      "with the goods of the {O}",
	"trade":      "along the trade with the {O}",
	"occupation": "with the taking of worlds from the {O}",
	"settling":   "from the worlds of the {O} nearby",
	"fleet":      "with the ships of the {O}",
	"landing":    "with the first landing of the {O}",
	"message":    "in a message from the {O}",
	"signal":     "on the signals of the {O}",
	"poison":     "hidden in the goods of the {O}",
	"whisper":    "hidden in a message from the {O}",
}

// beneathDesc is what each miracle's holders find when they first reach
// into the state beneath; beneathNames adds what they call it, the
// people's word being the second %s. A voiceless people gets the
// description alone.
var beneathDesc = map[string]string{
	"ftl":       "Whatever the ships pass through, nobody is in it. From inside, the crossing has no duration: one moment here, the next there. The ones who tried to stay awake for it did not come back as one person.",
	"ansible":   "The ansible does not cross space; it goes under it. The speakers do not hear what it goes under. Anyone who begins to hear it is taken off the line.",
	"foresight": "Precognition is not a looking forward. It is a leaning on something under time, where before and after are one thing, and the ones who lean too hard do not come back up.",
	"unmaking":  "What their annihilation does is not destruction. Matter is put back the way it was before it was matter.",
	"wound":     "There is a place in them where the wall is not, and what shows through it is the only thing they are afraid of.",
}

var beneathNames = map[string]string{
	"ftl":       "Whatever the ships pass through, nobody is in it. From inside, the crossing has no duration: one moment here, the next there. The ones who tried to stay awake for it did not come back as one person. The %s call it %s and do not look at it.",
	"ansible":   "The ansible does not cross space; it goes under it. The %s call what it goes under %s. The speakers do not hear it. Anyone who begins to hear it is taken off the line.",
	"foresight": "Precognition is not a looking forward. It is a leaning on something under time, where before and after are one thing. The %s call it %s, and the ones who lean too hard do not come back up.",
	"unmaking":  "What their annihilation does is not destruction. Matter is put back the way it was before it was matter. The %s have a word for that state, %s, and it is a word they say once.",
	"wound":     "There is a place in the %s where the wall is not. They call what shows through it %s, and it is the only thing they are afraid of.",
}

// numberWord is a small count in words.
func numberWord(n int) string {
	if n >= 0 && n < len(numberWords) {
		return numberWords[n]
	}
	return sprintf("%d", n)
}

var numberWords = []string{"no", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}

// listOf joins names with commas and an and.
func listOf(ns []string) string {
	switch len(ns) {
	case 0:
		return ""
	case 1:
		return ns[0]
	}
	out := ""
	for i, n := range ns {
		switch {
		case i == 0:
			out = n
		case i == len(ns)-1:
			out += " and " + n
		default:
			out += ", " + n
		}
	}
	return out
}
