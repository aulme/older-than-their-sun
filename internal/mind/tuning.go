// Package mind is every decision a people makes: the appraisal of a fight,
// the bar a posture needs, the council's choice, scouting and survey
// policy, pact offers and answers, the answer to a call, forwarding a
// report, the Find's choice, the research pick, the colony target, a
// relief fleet's turn and a nomad fleet's next star. Each decision is a
// pure function over an input struct the caller assembles (beliefs, dials,
// geometry already computed) and returns a value that can say Why. Every
// number a decision uses is a field of one Tuning; nothing here holds a
// literal a designer might want to move. The package knows nothing of the
// world: no star, no civilisation, only values.
package mind

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"worldgen/internal/plague"
)

// Tuning is every number the decisions use. Default() is today's sim.
type Tuning struct {
	Belief    BeliefTuning
	Appraise  AppraiseTuning
	Bar       BarTuning
	Council   CouncilTuning
	Scout     ScoutTuning
	Campaign  CampaignTuning
	Pact      PactTuning
	Call      CallTuning
	Forward   ForwardTuning
	Survey    SurveyTuning
	Sight     SightTuning
	Find      FindTuning
	Research  ResearchTuning
	Expand    ExpandTuning
	Build     BuildTuning
	Direction DirectionTuning
	Turn      TurnTuning
	Roam      RoamTuning
	Trade     TradeTuning
	Want      WantTuning
	Garrison  GarrisonTuning
	Intercept InterceptTuning
	Picket    PicketTuning
	Wisdom    WisdomTuning
	Contract  ContractTuning
	Slight    SlightTuning
	Refuse    RefuseTuning
	Ossify    OssifyTuning
	Plague    plague.Tuning // the plagues' own numbers, kept here so one table tunes everything
}

// BeliefTuning: what a people thinks another's level is from its
// intelligence, and how sure.
type BeliefTuning struct {
	UnknownBase        float64 // never seen: the lights of their cities are a guess at their age
	UnknownPerEra      float64
	UnknownSpread      float64
	Spread             float64 // a fresh report's spread, in levels
	SpreadPerKyr       float64 // widening per thousand years since the look
	MaxSpread          float64
	UnknownShips       float64 // never seen: the ships guessed, plus this per era
	UnknownShipsPerEra float64
	UnknownGuns        float64 // over a world not looked at: the silos nearly every world has
	UnknownGunsEra     int     // from this era on
}

// AppraiseTuning: the estimate of a fight.
type AppraiseTuning struct {
	Weakened    float64 // the enemy plagued or in a dark age
	OtherWar    float64 // per other war either side is in
	DarkAge     float64 // years after a dark age a people still counts as weakened
	AllyShare   float64 // of an ally's level that would come
	Scale       float64 // levels of margin per unit of the normal
	PrizeWeight float64 // the bar drops this much per unit of prize: the target's yield the attacker wants, and its rarities
	PrizeMax    float64 // and never by more than this
	Stiff       float64 // stiffness at which a people is a target of opportunity: its ways have set
}

// BarTuning: the odds a posture needs before it strikes.
type BarTuning struct {
	Hate           float64 // a xenophobe against the different
	Opportunist    float64
	Conqueror      float64
	Vengeful       float64 // against the grudge target
	GrudgeDiscount float64 // off the bar when there is a grudge
	StiffNew       float64 // what a stiff people adds to the bar per point of stiffness past one against a people it never fought: it prefers the wars it knows
	StiffNoFirst   float64 // stiffness from which a people that never sent a fleet sends none
}

// CouncilTuning: when a people decides, and how boldly.
type CouncilTuning struct {
	Cadence    float64 // councils per thousand years, when nothing summons one
	Compulsion float64 // chance a conqueror's honour compels it to a lower bar
	Compelled  float64 // the bar it is compelled to
}

// ScoutTuning: when a report is worth a ship.
type ScoutTuning struct {
	KeepHome   float64 // ships that must stay after the scout leaves
	FearBar    float64 // above this fear a small people does not scout
	FearShips  int     // small means fewer ships than this
	SightNoise float64 // the Sight's reading is nearly exact
}

// CampaignTuning: sizing a fleet.
type CampaignTuning struct {
	Floor        float64 // share of the ships in being a fleet is at least
	MinFloor     int     // and at least this many ships
	FearCap      float64 // how much of the ships fear keeps home, at full fear
	MaxLag       float64 // years of crossing beyond which nobody sends a fleet
	ConquerorLag float64 // but a conqueror will
	MusterMax    float64 // years a muster waits for its ships before it stands down
}

// PactTuning: offers and answers.
type PactTuning struct {
	Rate            float64 // proposals per thousand years, by default
	Confederate     float64
	Defensive       float64
	Aggressor       float64 // conquerors and the vengeful, seeking partners in war
	AskAgain        float64 // years before the same people is asked again
	ThreatSlack     float64 // a threat is at least this much below one's own level
	ThreatMargin    float64 // light years beyond the two reaches a threat still counts
	Accept          float64 // the score an offer needs
	Conqueror       float64 // aggression: posture terms
	OpportunistWeak float64
	VengefulGrudge  float64
	VengefulNone    float64
	Unyielding      float64
	GreedWeight     float64
	FearBase        float64 // defence: fear times a term in the enemy's excess
	FearPerLevel    float64
	FearMax         float64
	EnemyNear       float64
	AtWar           float64
	Grudge          float64
	DefConfederate  float64
	DefDefensive    float64
	DefMeek         float64 // pacifists, the submissive, the vengeful
	DefOpportunist  float64 // for or against, by who is stronger
	DefConqueror    float64
	Difference      float64 // per point of difference
	Infamy          float64
	ProposerGrudge  float64 // per point of grudge the asked holds against the proposer
	Renown          float64
	Loyalty         float64 // times loyalty less a half
}

// CallTuning: whether an ally comes.
type CallTuning struct {
	Share       float64 // of the ships that could sail, sent as relief
	Floor       float64 // of the ships in being, at least
	MinFloor    int
	HelpSlack   float64 // relief helps when it and the host reach the attacker less this, in levels
	SafeLeft    float64 // ships that must stay home
	SafeFear    float64 // unless fear is below this
	Base        float64 // want: loyalty less fear times FearWeight plus this
	FearWeight  float64
	Confederate float64
	Betrayed    float64 // off the want when the caller broke faith before
	Want        float64 // the want it takes to come
	BlameAbove  float64 // not coming is a betrayal when the ships are at least this
}

// ForwardTuning: passing a report to an ally.
type ForwardTuning struct {
	NeedStranger float64
	NeedMet      float64
	NeedEnemy    float64 // the ally is at war with the subject, or the pact names it
	Designs      float64 // greed's weight when the forwarder has its own designs
	Bar          float64
}

// SurveyTuning: how many surveyors a people keeps out, and their tour.
type SurveyTuning struct {
	HungerWeight float64
	GreedWeight  float64
	Necessity    int     // the era from which a people with nothing to settle sends one
	KeepHome     float64 // ships that must stay home
	MinReach     float64 // no surveyors below this reach
	NearMin      float64 // the least range within which unread stars are looked for
	Rate         float64 // launches per thousand years when short
	HopMin       float64 // the next star on a tour lies within a hop
	HopMax       float64
	MaxTour      int // stars per tour
	MaxTourYears float64
}

// SightTuning: the Sight's outward mode.
type SightTuning struct {
	FearBar   float64 // above this fear a hostile neighbour in reach turns it back
	GrudgeBar float64 // a grudge this deep makes a neighbour hostile
	Reads     float64 // stars read per thousand years
	RangeMul  float64 // of the reach
	RangeMin  float64
}

// FindTuning: what a people attempts with a remain, as weights.
type FindTuning struct {
	Master     float64
	Wield      float64
	Seal       float64
	Curious    float64 // more mastery
	Reaching   float64 // expansionists and symbiotes
	Cautious   float64 // more sealing
	Wary       float64 // xenophobes and the contemplative
	Practical  float64 // pragmatists and conquerors wield
	ThreatSeal float64 // a threat or a sleeper is sealed, never wielded
	Plain      float64 // a miracle artifact is plainly a thing to be used
	Own        float64 // one's own lost work is mastered
}

// ResearchTuning: the pick of the next pursuit.
type ResearchTuning struct {
	DepthBonus float64 // per node already known in the domain
	Unfed      float64 // multiplier on a node whose upkeep the spare does not cover
}

// ExpandTuning: colony ships.
type ExpandTuning struct {
	Rate           float64 // per system, per thousand years
	MaxRate        float64
	ParasiteReach  float64 // a rider settles nothing on its own
	Hop            float64 // a ship goes this far at most, unless the Door
	Blind          float64 // chance a ship goes on a guess when nothing read is free
	NeedShipsBelow float64 // reach below which a people with nowhere to go works on ships
	NeedShipsEra   int
	ShipFocus      float64
}

// BuildTuning: structures.
type BuildTuning struct {
	Rate        float64 // per thousand years
	Cover       float64 // the spare must cover a structure's upkeep this many times over before it is built
	LevelWeight float64 // a level lifted is worth this much yield per tick, when nothing is wanting
	DockWeight  float64 // a dock is worth this much yield per tick per ship wanting
	GunWeight   float64 // a gun standing over a world is worth this much yield per tick
	WatchWeight float64 // a light year of watch is worth this much yield per tick, times WatchBase plus fear
	WatchBase   float64
}

// InterceptTuning: meeting a fleet in the dark.
type InterceptTuning struct {
	Muster  float64 // years from the sighting before an interceptor can sail
	Samples int     // points along the rest of the quarry's line tried for a meeting
}

// PicketTuning: scouts that stay and watch.
type PicketTuning struct {
	Tour     float64 // years a picket holds its post
	FearBar  float64 // above this fear a picket is kept against any hostile neighbour in reach, in peacetime
	KeepHome float64 // ships that must stay home
	Rate     float64 // launches per thousand years when one is wanted
}

// WisdomTuning: how Wisdom moves an act toward the estimate. Nothing here
// changes what is believed or wanted.
type WisdomTuning struct {
	Tail          float64 // the tail term of the acted-on odds shrinks by this per point of Wisdom
	GrudgeFade    float64 // the Wisdom at which the grudge discount on the bar is gone
	VengefulBelow float64 // below this Wisdom the vengeful act on their hope against the grudge target
	Compulsion    float64 // a conqueror's compulsion shrinks by this per point
	Folly         float64 // the noise on a worth or a score, per point below ten
	ThreatSeal    float64 // per point, on the seal of a sleeper or a threat
	AboveEras     int     // a remain this many eras above the finder's is beyond it
	AboveWield    float64 // per point, off the wield of what is beyond
	AboveSeal     float64 // per point, on its seal
	LeapMargin    float64 // below this expected margin a people asks whether it can
	LeapBar       float64 // and seals instead when Wisdom and the roll clear this
	LeapNoise     float64 // the roll's spread
	TeachWeaker   float64 // a hostile or fixed posture does not explain itself to a people this many levels weaker
	BrokerRate    float64 // unpaid brokering, per thousand years, by a shared pact or a confederate
}

// ContractTuning: offers, margins, lengths and the sellsword's trade.
type ContractTuning struct {
	AskRate         float64 // how often a people looks for something to buy, per thousand years
	AskRateGreedy   float64 // for greed above GreedBar or a fixation on holding
	GreedBar        float64
	GreedMargin     float64 // the margin of the greedy
	EagerMargin     float64 // of a fixation selling what it wants sold anyway: conquest a strike, knowing a teach
	CrimeMargin     float64 // times, when the ask is a crime by the seller's morality
	XenoMargin      float64 // times, for a xenophobe against the different
	LengthPerUnit   float64 // thousand years of pay per ship asked or per rarity
	LengthMin       float64
	LengthMax       float64
	FlowRest        float64 // a flow's worth beyond the uses it refills, per unit
	TeachGiver      float64 // a teach costs the giver this share of what it saves the taught
	TeachHostile    float64 // times, when the taught is the giver's threat or strikes first
	GuardGiver      float64 // a guard costs the giver this per ship, times the risk
	RiskFront       float64 // the risk of loss at a star in a hostile front
	RiskElse        float64
	StrikeGrudge    float64 // what a strike costs the giver beyond its ships: the grudge it earns
	PrizeYears      float64 // a world's prize per tick counts for this many ticks in a worth
	WillWorth       float64 // a people's war will counts this much per point in the worth of peace
	SightingWorth   float64 // a sighting's worth per level of the fleet seen, at the start of its crossing
	BrokerGiver     float64 // a broker's cost: this share of the trade it has with both
	BrokerBase      float64 // and what understanding is worth to the buyer before the trade
	MercenaryRate   float64 // offers of a guard per thousand years at full want and full spare fleet
	FaithlessBuyOff float64 // a faithless hired fleet is bought at this times greed, per thousand years
	GraceTicks      int     // ticks a term may go undelivered before the contract breaks
	TributeLength   float64 // thousand years a tribute runs
	BrokerRate      float64 // brokered attempts per thousand years under a broker contract
	BrokerLapse     float64 // years after which a broker contract lapses
	SoldTold        float64 // chance the owner of a sold fleet learns who sold it, when it is met
	SellswordTicks  int     // ticks living on contract pay before a people is called sellswords
	SellswordShare  float64 // the share of a kind's income that has to be contract pay
}

// SlightTuning: a war on a partner's partner as a wrong, and its weight
// before the council.
type SlightTuning struct {
	Weight         float64 // the slight per unit of the pair's flow as a share of the slighted people's income
	Cap            float64 // the most one war slights one people
	Tick           float64 // the share of the slight taken again each tick the war starves the partner
	Source         float64 // the slight of a world taken that held a source the partner drew on
	Fact           float64 // the slight past which it is a fact
	Dependent      float64 // times, when the partner's sending kept the slighted people's uses fed
	OffenceWeight  float64 // units of prize one unit of offence costs on the bar
	StrongWeight   float64 // a strong neighbour's opinion, against a partner's one
	FearBar        float64 // fear above this doubles the strong neighbour's weight
	ConquerorShare float64 // a conqueror weighs the whole offence by this
}

// WantTuning: how many ships a people builds toward.
type WantTuning struct {
	Floor      int     // ships kept whatever else is wanted
	FearWeight float64 // more by fear, rounded
}

// GarrisonTuning: where a people keeps its ships.
type GarrisonTuning struct {
	HomeBase    float64 // the home wants the strongest threat's ships times this plus fear
	HomeMin     int     // and at least this many
	ColonyShare float64 // a world in an enemy's reach wants this share of what the home would
}

// TradeTuning: what a people sends a partner.
type TradeTuning struct {
	CapBase        float64 // the share of the spare that can cross by slow ships
	CapFast        float64 // with beamed sails or near-light travel
	CapDoor        float64 // with the Door or wormholes
	CapNomad       float64 // to or from a people that lives as fleets, whatever the drive
	GrudgeBar      float64 // a grudge above this and nothing is sent
	DifferentShare float64 // what a xenophobe sends a people it counts as different
	FearBar        float64 // fear above this sends no metal to a stronger, hostile partner
	SpawnOrganic   float64 // a people fixed on spawning gives organic matter at this times its cap
	GraspWant      float64 // a partner fixed on holding has its want weighed this much more in the share
	UsesOnly       bool    // goods that came by trade feed uses only: the spare that launches and builds is capped at what the people's own income would leave
}

// DirectionTuning: what bends the order the means are fed in.
type DirectionTuning struct {
	FearBar   float64 // fear above this, with a hostile neighbour in reach, puts arms first
	HungerBar float64 // hunger above this puts the mind before the works
	GreedBar  float64 // greed above this puts the road before the mind
}

// TurnTuning: a relief fleet seizing the world it keeps.
type TurnTuning struct {
	Faithful  float64 // per thousand years, by honour
	Practical float64
	Faithless float64
	Hostile   float64 // opportunists and conquerors, times
	Vengeful  float64 // the vengeful against a betrayer, times
	Opening   float64 // the fleet must exceed the hold by this
	GreedBase float64 // the chance is times greed plus this
}

// RoamTuning: a nomad fleet's hop.
type RoamTuning struct {
	HopMin float64
	HopMax float64
}

// RefuseTuning: the chance a people that suspects a sender of sickness
// closes its ears and its ports to it.
type RefuseTuning struct {
	Creed   float64 // what the quarantine creed adds
	Caution float64 // what a cautious nature adds
	Cap     float64
	Censor  float64 // censorship: the chance at this multiple
}

// OssifyTuning: how a people's ways set, and when that comes for it. See
// history's ossify.go for the growth table these feed and the filter.
type OssifyTuning struct {
	Base         float64 // stiffness per thousand years for a lone cradle world under a fresh sky
	PerWorld     float64 // the size term: 1 + worlds times this
	Fade         float64 // the fading galaxy: 1 + this times one less the fertility
	Still        float64 // times this when nothing new has happened in the window
	StillKyr     float64 // the window, thousand years
	Ossified     float64 // times this once the people has set
	IronScar     float64 // times this per iron answer held: centralism, stewardship, quarantine, fatalism
	Chance       float64 // facings per thousand years per point of stiffness past one
	Renaissance  float64 // added per renaissance
	Foresight    float64 // taken off for a people that sees it coming
	ForeseeAt    float64 // stiffness from which the foresighted tilt their research to society
	ForeseeTilt  float64 // by this
	Renewal      float64 // research at this multiple after a renaissance
	RenewalKyr   float64 // for this long
	Opportunity  float64 // stiffness at which the neighbours read a people as slow to answer
	CivilWar     float64 // the break: chance of a civil war per world held, capped
	CivilWarCap  float64
	DepthBase    float64 // a dark age's depth: base, plus this per unit of stiffness over three, plus this per earlier dark age, plus noise
	DepthStiff   float64
	DepthPrior   float64
	DepthNoise   float64
	DepthMin     float64
	DepthMax     float64
	Shards       int     // the most peoples a shattering makes; the rest of the worlds are abandoned
	GrudgeDecay  float64 // what a grudge keeps per thousand years
	GrudgeFloor  float64 // below this a grudge is gone
	HeirGrudge   float64 // what an heir keeps of the old people's grudges
	HeldGrudge   float64 // what others keep against an heir of what they held against the old people
	SunderGrudge float64 // what the heirs of a civil war hold against each other
	KinLoyalty   float64 // the loyalty term in a pact answer between kin, times
	KinLine      float64 // the grudge above which kinship's warmth is suspended
}

// Default is today's numbers.
func Default() *Tuning {
	return &Tuning{
		Belief:   BeliefTuning{UnknownBase: 1, UnknownPerEra: 1.2, UnknownSpread: 3, Spread: 0.3, SpreadPerKyr: 0.1, MaxSpread: 3, UnknownShips: 1, UnknownShipsPerEra: 1, UnknownGuns: 2, UnknownGunsEra: 2},
		Appraise: AppraiseTuning{Weakened: 1, OtherWar: 0.3, AllyShare: 0.5, Scale: 2, DarkAge: 50_000, PrizeWeight: 0.02, PrizeMax: 0.15, Stiff: 1.5},
		Bar:      BarTuning{Hate: 0.35, Opportunist: 0.75, Conqueror: 0.4, Vengeful: 0.3, GrudgeDiscount: 0.1, StiffNew: 0.2, StiffNoFirst: 2},
		Council:  CouncilTuning{Cadence: 0.3, Compulsion: 0.1, Compelled: 0.25},
		Scout:    ScoutTuning{KeepHome: 1, FearBar: 0.8, FearShips: 3, SightNoise: 0.1},
		Campaign: CampaignTuning{Floor: 0.1, MinFloor: 1, FearCap: 0.4, MaxLag: 20_000, ConquerorLag: 40_000, MusterMax: 30_000},
		Pact: PactTuning{
			Rate: 0.08, Confederate: 0.5, Defensive: 0.15, Aggressor: 0.2, AskAgain: 30_000,
			ThreatSlack: 0.5, ThreatMargin: 10, Accept: 0.45,
			Conqueror: 0.3, OpportunistWeak: 0.2, VengefulGrudge: 0.4, VengefulNone: 0.3, Unyielding: 0.1, GreedWeight: 0.5,
			FearBase: 0.5, FearPerLevel: 0.25, FearMax: 1.5, EnemyNear: 0.2, AtWar: 0.3, Grudge: 0.2,
			DefConfederate: 0.3, DefDefensive: 0.2, DefMeek: 0.1, DefOpportunist: 0.2, DefConqueror: 0.1,
			Difference: 0.1, Infamy: 0.3, ProposerGrudge: 0.5, Renown: 0.15, Loyalty: 0.2,
		},
		Call:      CallTuning{Share: 0.3, Floor: 0.1, MinFloor: 1, HelpSlack: 1, SafeLeft: 2, SafeFear: 0.3, Base: 0.3, FearWeight: 0.6, Confederate: 0.2, Betrayed: 1, Want: 0.4, BlameAbove: 3},
		Forward:   ForwardTuning{NeedStranger: 0.3, NeedMet: 0.5, NeedEnemy: 1, Designs: 0.6, Bar: 0.3},
		Survey:    SurveyTuning{HungerWeight: 2, GreedWeight: 1, Necessity: 2, KeepHome: 1, MinReach: 1, NearMin: 5, Rate: 0.3, HopMin: 3, HopMax: 20, MaxTour: 6, MaxTourYears: 40_000},
		Sight:     SightTuning{FearBar: 0.6, GrudgeBar: 0.5, Reads: 2, RangeMul: 2, RangeMin: 10},
		Find:      FindTuning{Master: 1, Wield: 1.5, Seal: 1, Curious: 3, Reaching: 1, Cautious: 3, Wary: 1.5, Practical: 2, ThreatSeal: 2, Plain: 2, Own: 3},
		Research:  ResearchTuning{DepthBonus: 0.25, Unfed: 0.25},
		Expand:    ExpandTuning{Rate: 0.04, MaxRate: 0.3, ParasiteReach: 0.3, Hop: 20, Blind: 0.2, NeedShipsBelow: 40, NeedShipsEra: 2, ShipFocus: 4},
		Build:     BuildTuning{Rate: 0.004, Cover: 2, LevelWeight: 2, DockWeight: 1, GunWeight: 1, WatchWeight: 0.05, WatchBase: 0.5},
		Direction: DirectionTuning{FearBar: 0.6, HungerBar: 0.6, GreedBar: 0.6},
		Turn:      TurnTuning{Faithful: 0.0005, Practical: 0.01, Faithless: 0.05, Hostile: 2, Vengeful: 3, Opening: 1, GreedBase: 0.5},
		Roam:      RoamTuning{HopMin: 3, HopMax: 20},
		Trade:     TradeTuning{CapBase: 0.25, CapFast: 0.5, CapDoor: 1, CapNomad: 0.5, GrudgeBar: 0.3, DifferentShare: 0.5, FearBar: 0.6, SpawnOrganic: 2, GraspWant: 2, UsesOnly: false},
		Want:      WantTuning{Floor: 1, FearWeight: 1},
		Garrison:  GarrisonTuning{HomeBase: 0.5, HomeMin: 1, ColonyShare: 0.3},
		Intercept: InterceptTuning{Muster: 50, Samples: 64},
		Picket:    PicketTuning{Tour: 20_000, FearBar: 0.6, KeepHome: 1, Rate: 0.3},
		Contract: ContractTuning{
			AskRate: 0.05, AskRateGreedy: 0.15, GreedBar: 0.6, GreedMargin: 1.5, EagerMargin: 0.5, CrimeMargin: 2, XenoMargin: 2,
			LengthPerUnit: 5, LengthMin: 5, LengthMax: 50, FlowRest: 0.1, TeachGiver: 0.1, TeachHostile: 5,
			GuardGiver: 0.1, RiskFront: 0.5, RiskElse: 0.1, StrikeGrudge: 3, PrizeYears: 20, WillWorth: 10, SightingWorth: 5,
			BrokerGiver: 0.1, BrokerBase: 10, MercenaryRate: 0.02, FaithlessBuyOff: 0.3, GraceTicks: 2, TributeLength: 20,
			BrokerRate: 0.1, BrokerLapse: 30_000, SoldTold: 0.5, SellswordTicks: 10, SellswordShare: 0.5,
		},
		Slight: SlightTuning{Weight: 2, Cap: 1, Tick: 0.1, Source: 0.3, Fact: 0.2, Dependent: 2, OffenceWeight: 5, StrongWeight: 0.5, FearBar: 0.6, ConquerorShare: 0.5},
		Refuse: RefuseTuning{Creed: 0.5, Caution: 0.3, Cap: 0.9, Censor: 2},
		Ossify: OssifyTuning{
			Base: 0.00025, PerWorld: 1.0 / 8, Fade: 2, Still: 1.5, StillKyr: 200, Ossified: 1.5, IronScar: 1.25,
			Chance: 0.0015, Renaissance: 1, Foresight: 1, ForeseeAt: 0.7, ForeseeTilt: 1.5, Renewal: 1.5, RenewalKyr: 300, Opportunity: 1.5,
			CivilWar: 1.0 / 6, CivilWarCap: 0.8,
			DepthBase: 0.1, DepthStiff: 0.3, DepthPrior: 0.1, DepthNoise: 0.1, DepthMin: 0.1, DepthMax: 0.8, Shards: 8,
			GrudgeDecay: 0.995, GrudgeFloor: 0.05, HeirGrudge: 0.25, HeldGrudge: 0.5, SunderGrudge: 3, KinLoyalty: 2, KinLine: 0.5,
		},
		Plague: plague.Default(),
		Wisdom: WisdomTuning{Tail: 0.08, GrudgeFade: 10, VengefulBelow: 7, Compulsion: 0.08, Folly: 0.04, ThreatSeal: 0.3, AboveEras: 2, AboveWield: 0.07, AboveSeal: 0.2, LeapMargin: -1.5, LeapBar: 6, LeapNoise: 1.5, TeachWeaker: 0, BrokerRate: 0.02},
	}
}

// Load reads a tuning file: JSON with the shape of Tuning, fields left out
// keeping their defaults.
func Load(path string) (*Tuning, error) {
	t := Default()
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, t); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return t, nil
}

// Set applies one override, "Council.Cadence=0.5", by field name.
func (t *Tuning) Set(kv string) error {
	key, val, ok := strings.Cut(kv, "=")
	if !ok {
		return fmt.Errorf("tune %q: want key=value", kv)
	}
	v := reflect.ValueOf(t).Elem()
	for _, part := range strings.Split(key, ".") {
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("tune %q: %s is not a group", kv, key)
		}
		v = v.FieldByName(part)
		if !v.IsValid() {
			return fmt.Errorf("tune %q: no field %s", kv, key)
		}
	}
	switch v.Kind() {
	case reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("tune %q: %w", kv, err)
		}
		v.SetFloat(f)
	case reflect.Int:
		n, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("tune %q: %w", kv, err)
		}
		v.SetInt(int64(n))
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("tune %q: %w", kv, err)
		}
		v.SetBool(b)
	default:
		return fmt.Errorf("tune %q: %s is a group, not a number", kv, key)
	}
	return nil
}

// Configure is what the commands do with their flags: the file if named,
// then each override in order.
func Configure(path string, sets []string) (*Tuning, error) {
	t := Default()
	if path != "" {
		var err error
		if t, err = Load(path); err != nil {
			return nil, err
		}
	}
	for _, kv := range sets {
		if err := t.Set(kv); err != nil {
			return nil, err
		}
	}
	return t, nil
}
