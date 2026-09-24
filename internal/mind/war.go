package mind

import "fmt"

// The war council is the question "what do we do about the war we are
// in", asked of each side of each running war; the ordinary council is
// the question "whom should we start one with". A war has an aim, which
// decides what winning is, when pressing stops and what terms are asked;
// the council reads the balance as believed against the next target the
// aim names, the aim's progress, its will and its fear, and presses,
// holds or sues. Terms offered are answered by the other side's own
// council, and a small people offered vassalage by a large one answers
// from its reading of the war it would otherwise fight.

// The aims. A limited aim (a world, tribute, redress) makes a short war;
// a total one (submission, ending) a long one. Defence is an ally's; hold
// is the side declared on's, and it may press to retake what it lost.
const (
	AimWorld      = "world"
	AimTribute    = "tribute"
	AimSubmission = "submission"
	AimRedress    = "redress"
	AimEnding     = "ending"
	AimDefence    = "defence"
	AimHold       = "hold"
)

// Limited says whether an aim can be met short of the enemy's end.
func Limited(aim string) bool {
	return aim == AimWorld || aim == AimTribute || aim == AimRedress
}

// WarChoice is what a war council decides for its side.
type WarChoice uint8

const (
	Hold  WarChoice = iota // keep the ships home; the campaigns out go on
	Press                  // send a campaign at the next target the aim names
	Sue                    // offer terms
)

func (c WarChoice) String() string { return [...]string{"hold", "press", "sue"}[c] }

// WarInput is one side of one war before its council.
type WarInput struct {
	Aim      string
	Declarer bool
	Posture  string
	Hates    bool    // a xenophobe against the different: nothing short of the end
	Acted    float64 // the odds acted on against the next target the aim names
	Reach    bool    // a fleet could be sent at it
	Ready    bool    // and the ships the sizing wants for it could sail, or be gathered
	Out      bool    // a campaign of its own is out or gathering in this war
	Met      bool    // the aim is met: the world taken, the home fallen, what was lost won back
	Fought   bool    // the war has seen a battle
	Idle     float64 // ticks since the last battle, or since the war began
	Will     float64
	Fear     float64
	Lost     int     // worlds it has lost in this war
	Taken    int     // and taken
	Appetite float64 // worlds taken of late, as the bar's discount; see Appetite
	Wary     float64 // the wars it came off worst in against this enemy, fading
	Treats   bool    // terms can pass between the two: each understands the other and both make terms
	Bound    float64 // an ally's: what the alliance is worth to it, 0 to 1, which a separate peace would throw away; see Bound
}

// WarVerdict is the council's answer, with what it weighed.
type WarVerdict struct {
	Choice WarChoice
	Acted  float64
	Bar    float64 // the odds it needed to press
	Sue    float64 // below which, afraid, it sues
	Key    string  // the reason in a word, for a count: met, will, afraid, posture, out, reach, odds, ships, press, and stale- before a hold turned to a suit
	Reason string
}

// Why says the verdict in a line.
func (v WarVerdict) Why() string {
	return fmt.Sprintf("%s: acting on %.2f against a bar of %.2f to press and %.2f to sue; %s", v.Choice, v.Acted, v.Bar, v.Sue, v.Reason)
}

// PressBar is the odds a side needs to send a campaign in a war it is
// already in, by its posture and its place in the war: the declarer at
// its posture's own, the side declared on higher unless it has lost
// worlds to win back, an ally at the defence bar. Pacifists and the
// submissive never carry a war, and the hating always would.
func PressBar(in WarInput, t *Tuning) (bar float64, ok bool) {
	p := &t.War
	switch {
	case in.Hates:
		return p.PressHate, true
	case in.Posture == Pacifist || in.Posture == Submissive:
		return 0, false
	}
	switch in.Aim {
	case AimDefence:
		bar = p.PressAlly
	case AimHold:
		bar = p.PressDefend
		if in.Lost > in.Taken {
			bar -= p.Retake
		}
	default:
		bar = p.PressDeclarer
	}
	switch in.Posture {
	case Conqueror, Vengeful:
		bar -= p.Bold
	case Opportunist:
		if !in.Declarer {
			bar += p.Bold // it folds on the first real loss
		}
	case Unyielding:
		bar -= p.Bold
	}
	return min(1, max(0, bar-in.Appetite+min(p.WaryMax, p.WaryBar*in.Wary))), true
}

// WarCouncil is press, hold or sue for one side of one war.
//
// A side whose aim is met offers peace on the lines as they stand. A side
// that has lost the will, or that is afraid (the balance believed against
// it, the more so the more it fears) and has seen the war begin in
// earnest, sues. A side that believes it can take the next target the aim
// names, and has no campaign out, presses. The rest hold. The unyielding
// and the hating never sue; a war between peoples who cannot treat has
// no terms to offer.
func WarCouncil(in WarInput, t *Tuning) WarVerdict {
	p := &t.War
	v := WarVerdict{Acted: in.Acted, Sue: min(p.SueMax, p.SueBase+p.SueFear*in.Fear)}
	bar, presses := PressBar(in, t)
	v.Bar = bar
	proud := in.Posture == Unyielding || in.Hates || in.Aim == AimEnding
	stay := 1 - p.AllyStay*in.Bound // an ally sues alone only for the stronger reasons, the more so the more the alliance is worth
	switch {
	case in.Treats && in.Met:
		v.Choice, v.Key, v.Reason = Sue, "met", "the aim is met: peace on the lines"
		return v
	case in.Treats && !proud && in.Will < p.SueWill*stay:
		v.Choice, v.Key, v.Reason = Sue, "will", "the will is nearly gone"
		return v
	case in.Treats && !proud && (in.Fought || in.Lost > 0) && in.Acted < v.Sue*stay && !in.Out:
		v.Choice, v.Key, v.Reason = Sue, "afraid", "afraid, and the balance is against it"
		return v
	}
	stale := in.Treats && !in.Out && in.Idle*stay >= p.IdleSue && in.Posture != Unyielding && !in.Hates
	building := presses && in.Reach && in.Acted >= bar && !in.Ready // the odds are there and the ships are coming
	if building && in.Idle < p.IdleSue*p.BuildWait {
		stale = false
	}
	switch {
	case !presses:
		v.Key, v.Reason = "posture", "its posture does not carry a war"
	case in.Out:
		v.Key, v.Reason = "out", "a campaign is out"
		return v
	case !in.Reach:
		v.Key, v.Reason = "reach", "nothing in reach"
	case in.Acted < bar:
		v.Key, v.Reason = "odds", "the odds are not"
	case !in.Ready:
		v.Key, v.Reason = "ships", "the odds are good, and the ships are too few"
	default:
		v.Choice, v.Key, v.Reason = Press, "press", "the odds are good enough"
		return v
	}
	if stale {
		v.Choice, v.Key, v.Reason = Sue, "stale-"+v.Key, "nobody has fought for a while, and it will not go: "+v.Reason
	}
	return v
}

// BoundInput is what an ally in a war joined by pact reads of the
// alliance.
type BoundInput struct {
	Age      float64 // years since the pact was made
	Members  int     // its members
	Menace   bool    // the enemy menaces the ally itself
	Renown   float64 // the principal's
	Betrayed bool    // the principal has broken faith with the ally
}

// Bound is what an alliance is worth to an ally in its principal's war,
// from 0 to 1: the more, the older and larger the pact, the more the
// enemy menaces the ally on its own account and the more the principal
// has kept faith. A separate peace throws it away, and the record counts
// it a betrayal every member weighs; so the ally that holds it dear
// stays in a war it would leave on its own account.
func Bound(in BoundInput, t *Tuning) float64 {
	p := &t.War
	b := p.AllyBase + p.AllyAge*min(1, in.Age/100_000) + p.AllyMembers*float64(min(3, max(0, in.Members-2))) + p.AllyRenown*min(1, in.Renown)
	if in.Menace {
		b += p.AllyMenace
	}
	if in.Betrayed {
		b -= p.AllyBetrayed
	}
	return min(1, max(0, b))
}

// Escalate is the aim of a war between old enemies: the second war
// between two with a grudge standing is at least for redress, and the
// third and after for the whole when the declarer has grown to many
// times the other (the Punic pattern: Rome was the larger by the third
// war). Between rivals of a size the third is redress again: a war for
// the whole between equals ended the rivalries it was meant to carry
// (step 11's batches: old enemies 104 pairs with it, 188 without). A war
// for a total aim, or an ally's, stays as it is. Size is the declarer's
// worlds over the other's.
func Escalate(aim string, nth int, grudge bool, size float64, t *Tuning) string {
	if !grudge || !Limited(aim) {
		return aim
	}
	switch {
	case nth >= t.War.Escalate && size >= t.War.EscalateSize:
		return AimSubmission
	case nth >= 2 && aim != AimRedress:
		return AimRedress
	}
	return aim
}

// TermsInput is an offer of terms before the council of the side it is
// made to.
type TermsInput struct {
	Worth   float64 // how much of its aim the terms give it, 0 to 1
	Acted   float64 // its odds acted on of taking the rest by pressing on
	Will    float64
	Posture string
	Hates   bool
	Greed   float64
	Presses bool    // its council would press on: a side that would not has nothing to refuse with
	Bound   float64 // an ally's: what the alliance is worth to it; peace alone would be a separate peace
}

// TermsAnswer is the answer.
type TermsAnswer struct {
	Accept bool
	Score  float64
}

// Why says the answer.
func (a TermsAnswer) Why() string {
	if a.Accept {
		return fmt.Sprintf("accepted: the terms outweigh pressing on by %.2f", a.Score)
	}
	return fmt.Sprintf("refused: pressing on outweighs the terms by %.2f", -a.Score)
}

// AnswerTerms weighs terms against pressing on: what they give of the
// aim, against the odds of taking the rest, which greed and a
// conqueror's hunger weigh up and a spent will weighs down. The
// unyielding take nothing short of the aim, and the hating nothing at
// all while they have the will to go on.
func AnswerTerms(in TermsInput, t *Tuning) TermsAnswer {
	p := &t.War
	tired := max(0, p.AcceptWill-in.Will) * p.Tired
	want := 1 + p.AcceptGreed*in.Greed
	if in.Posture == Conqueror {
		want += p.AcceptConqueror
	}
	a := TermsAnswer{Score: in.Worth + p.AcceptSlack + tired - in.Acted*want}
	if !in.Presses {
		a.Score += p.NoPress
	}
	a.Score -= p.AllyTerms * in.Bound
	switch {
	case in.Hates && in.Will > 0:
		a.Score = min(a.Score, -1)
	case in.Posture == Unyielding && in.Worth < 1:
		a.Score = min(a.Score, -1)
	}
	a.Accept = a.Score >= 0
	return a
}

// YokeInput is a small people offered vassalage by a large one without a
// war: what it believes of the war it would fight instead.
type YokeInput struct {
	Acted   float64 // its odds acted on of holding its home against them
	Fear    float64
	Posture string
	Hates   bool // it hates them: it would rather die
}

// YokeAnswer is the answer.
type YokeAnswer struct {
	Accept bool
	Bar    float64
}

// Why says the answer.
func (a YokeAnswer) Why() string {
	if a.Accept {
		return fmt.Sprintf("bends the knee: its odds are below %.2f", a.Bar)
	}
	return fmt.Sprintf("refuses: its odds clear %.2f, or it will not bend", a.Bar)
}

// AnswerYoke accepts vassalage when the war it would otherwise fight is
// believed lost, the more readily the more it fears. The unyielding, the
// conqueror and the hating do not bend; the submissive bend at better
// odds than anyone.
func AnswerYoke(in YokeInput, t *Tuning) YokeAnswer {
	p := &t.War
	a := YokeAnswer{Bar: p.YokeBar + p.YokeFear*in.Fear}
	switch in.Posture {
	case Unyielding, Conqueror:
		return a
	case Submissive:
		a.Bar += p.YokeMeek
	}
	if in.Hates {
		return a
	}
	a.Accept = in.Acted < a.Bar
	return a
}

// OffersYoke says whether a people that would strike a smaller one offers
// it vassalage first: a conqueror wants homes, and a client is one; an
// opportunist takes what it can hold; a people that hates the other does
// not want it kept.
func OffersYoke(posture string, hates bool) bool {
	return !hates && (posture == Conqueror || posture == Opportunist)
}

// Appetite is how much worlds taken of late lower the bar to the next
// war and to pressing the current one.
func Appetite(worlds float64, t *Tuning) float64 {
	return min(t.War.AppetiteMax, t.War.Appetite*worlds)
}
