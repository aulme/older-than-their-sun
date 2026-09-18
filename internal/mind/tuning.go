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
}

// BeliefTuning: what a people thinks another's level is from its
// intelligence, and how sure.
type BeliefTuning struct {
	UnknownBase   float64 // never seen: the lights of their cities are a guess at their age
	UnknownPerEra float64
	UnknownSpread float64
	Spread        float64 // a fresh report's spread, in levels
	SpreadPerKyr  float64 // widening per thousand years since the look
	MaxSpread     float64
}

// AppraiseTuning: the estimate of a fight.
type AppraiseTuning struct {
	Defence     float64 // the defender fights with the world's industry behind it
	HomeDefence float64 // and the home with everything it has
	Grid        float64 // a defence grid seen
	Weakened    float64 // the enemy plagued or in a dark age
	OtherWar    float64 // per other war either side is in
	DarkAge     float64 // years after a dark age a people still counts as weakened
	AllyShare   float64 // of an ally's level that would come
	Scale       float64 // levels of margin per unit of the normal
	PrizeWeight float64 // the bar drops this much per unit of prize: the target's yield the attacker wants, and its rarities
	PrizeMax    float64 // and never by more than this
}

// BarTuning: the odds a posture needs before it strikes.
type BarTuning struct {
	Hate           float64 // a xenophobe against the different
	Opportunist    float64
	Conqueror      float64
	Vengeful       float64 // against the grudge target
	GrudgeDiscount float64 // off the bar when there is a grudge
}

// CouncilTuning: when a people decides, and how boldly.
type CouncilTuning struct {
	Cadence    float64 // councils per thousand years, when nothing summons one
	Compulsion float64 // chance a conqueror's honour compels it to a lower bar
	Compelled  float64 // the bar it is compelled to
}

// ScoutTuning: when a report is worth a level.
type ScoutTuning struct {
	KeepHome   float64 // levels that must stay after the scout leaves
	FearBar    float64 // above this fear a small people does not scout
	FearMil    float64 // small means below this
	SightNoise float64 // the Sight's reading is nearly exact
}

// CampaignTuning: sizing a fleet.
type CampaignTuning struct {
	Floor        float64 // share of the whole strength a fleet is at least
	MinFloor     float64 // and at least this many levels
	FearCap      float64 // how much of the level fear keeps home, at full fear
	MaxLag       float64 // years of crossing beyond which nobody sends a fleet
	ConquerorLag float64 // but a conqueror will
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
	Renown          float64
	Loyalty         float64 // times loyalty less a half
}

// CallTuning: whether an ally comes.
type CallTuning struct {
	Share       float64 // of the level sent as relief
	Floor       float64 // of the whole strength, at least
	MinFloor    float64
	HelpSlack   float64 // relief helps when it and the host reach the attacker less this
	SafeLeft    float64 // levels that must stay home
	SafeFear    float64 // unless fear is below this
	Base        float64 // want: loyalty less fear times FearWeight plus this
	FearWeight  float64
	Confederate float64
	Betrayed    float64 // off the want when the caller broke faith before
	Want        float64 // the want it takes to come
	BlameAbove  float64 // not coming is a betrayal when the level is at least this
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
	KeepHome     float64 // levels that must stay home
	MinMil       float64 // no surveyors below this level
	MinReach     float64 // or this reach
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

// Default is today's numbers.
func Default() *Tuning {
	return &Tuning{
		Belief:   BeliefTuning{UnknownBase: 1, UnknownPerEra: 1.2, UnknownSpread: 3, Spread: 0.3, SpreadPerKyr: 0.1, MaxSpread: 3},
		Appraise: AppraiseTuning{Defence: 1, HomeDefence: 2.5, Grid: 0.5, Weakened: 1, OtherWar: 0.3, AllyShare: 0.5, Scale: 2, DarkAge: 50_000, PrizeWeight: 0.02, PrizeMax: 0.15},
		Bar:      BarTuning{Hate: 0.35, Opportunist: 0.75, Conqueror: 0.4, Vengeful: 0.3, GrudgeDiscount: 0.1},
		Council:  CouncilTuning{Cadence: 0.3, Compulsion: 0.1, Compelled: 0.25},
		Scout:    ScoutTuning{KeepHome: 1, FearBar: 0.8, FearMil: 4, SightNoise: 0.1},
		Campaign: CampaignTuning{Floor: 0.1, MinFloor: 1, FearCap: 0.4, MaxLag: 20_000, ConquerorLag: 40_000},
		Pact: PactTuning{
			Rate: 0.08, Confederate: 0.5, Defensive: 0.15, Aggressor: 0.2, AskAgain: 30_000,
			ThreatSlack: 0.5, ThreatMargin: 10, Accept: 0.45,
			Conqueror: 0.3, OpportunistWeak: 0.2, VengefulGrudge: 0.4, VengefulNone: 0.3, Unyielding: 0.1, GreedWeight: 0.5,
			FearBase: 0.5, FearPerLevel: 0.25, FearMax: 1.5, EnemyNear: 0.2, AtWar: 0.3, Grudge: 0.2,
			DefConfederate: 0.3, DefDefensive: 0.2, DefMeek: 0.1, DefOpportunist: 0.2, DefConqueror: 0.1,
			Difference: 0.1, Infamy: 0.3, Renown: 0.15, Loyalty: 0.2,
		},
		Call:      CallTuning{Share: 0.3, Floor: 0.1, MinFloor: 1, HelpSlack: 1, SafeLeft: 2, SafeFear: 0.3, Base: 0.3, FearWeight: 0.6, Confederate: 0.2, Betrayed: 1, Want: 0.4, BlameAbove: 3},
		Forward:   ForwardTuning{NeedStranger: 0.3, NeedMet: 0.5, NeedEnemy: 1, Designs: 0.6, Bar: 0.3},
		Survey:    SurveyTuning{HungerWeight: 2, GreedWeight: 1, Necessity: 2, KeepHome: 1, MinMil: 2, MinReach: 1, NearMin: 5, Rate: 0.3, HopMin: 3, HopMax: 20, MaxTour: 6, MaxTourYears: 40_000},
		Sight:     SightTuning{FearBar: 0.6, GrudgeBar: 0.5, Reads: 2, RangeMul: 2, RangeMin: 10},
		Find:      FindTuning{Master: 1, Wield: 1.5, Seal: 1, Curious: 3, Reaching: 1, Cautious: 3, Wary: 1.5, Practical: 2, ThreatSeal: 2, Plain: 2, Own: 3},
		Research:  ResearchTuning{DepthBonus: 0.25, Unfed: 0.25},
		Expand:    ExpandTuning{Rate: 0.04, MaxRate: 0.3, ParasiteReach: 0.3, Hop: 20, Blind: 0.2, NeedShipsBelow: 40, NeedShipsEra: 2, ShipFocus: 4},
		Build:     BuildTuning{Rate: 0.004, Cover: 2, LevelWeight: 2},
		Direction: DirectionTuning{FearBar: 0.6, HungerBar: 0.6, GreedBar: 0.6},
		Turn:      TurnTuning{Faithful: 0.0005, Practical: 0.01, Faithless: 0.05, Hostile: 2, Vengeful: 3, Opening: 1, GreedBase: 0.5},
		Roam:      RoamTuning{HopMin: 3, HopMax: 20},
		Trade:     TradeTuning{CapBase: 0.25, CapFast: 0.5, CapDoor: 1, CapNomad: 0.5, GrudgeBar: 0.3, DifferentShare: 0.5, FearBar: 0.6},
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
