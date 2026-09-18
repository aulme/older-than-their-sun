package mind

import (
	"fmt"
	"math"
)

// The appraisal is one estimate of a fight against a target, made from
// beliefs, that every choice routes through: whether to declare, to send a
// fleet, to scout, to accept a pact, to come when called.

// phi is the normal cumulative distribution.
func phi(x float64) float64 { return 0.5 * math.Erfc(-x/math.Sqrt2) }

// BeliefInput is what a people holds about another's level.
type BeliefInput struct {
	Seen     bool    // any report at all
	Mil      float64 // the level in the report
	AgeKyr   float64 // thousands of years since the look
	EnemyEra int     // for the guess when nothing was seen
}

// Believe is what a people thinks another's level is, and how sure: the
// spread widens with the age of the report, to a cap.
func Believe(in BeliefInput, t *Tuning) (mil, spread float64) {
	b := &t.Belief
	if !in.Seen {
		return b.UnknownBase + b.UnknownPerEra*float64(in.EnemyEra), b.UnknownSpread
	}
	return in.Mil, min(b.MaxSpread, b.Spread+b.SpreadPerKyr*in.AgeKyr)
}

// StrengthInput is what a people brings against one enemy.
type StrengthInput struct {
	Mil       float64   // its level at home
	Bonus     float64   // its miracles
	Allies    []float64 // levels of the allies who would actually join
	OtherWars int       // its other wars
}

// Strength is the level a people brings: its own, a share of each ally's,
// less what its other wars take.
func Strength(in StrengthInput, t *Tuning) float64 {
	a := &t.Appraise
	s := in.Mil + in.Bonus
	for _, m := range in.Allies {
		s += a.AllyShare * m
	}
	for i := 0; i < in.OtherWars; i++ {
		s -= a.OtherWar
	}
	return s
}

// AppraiseInput is a fight as the attacker sees it.
type AppraiseInput struct {
	Strength   float64 // the attacker's, from Strength
	Believed   float64 // the enemy's level as believed
	Spread     float64 // and how sure
	EnemyBonus float64 // the enemy's miracles
	AtHome     bool    // the target is the enemy's home
	Grid       bool    // a defence grid was seen
	Relief     float64 // others seen standing at the target
	Weakened   bool    // the enemy is plagued or in a dark age
	OtherWars  int     // the enemy's other wars
	Risk       float64 // the attacker's risk dial
	Dist       float64 // to the target, in light years
	Speed      float64 // years per light year
	Prize      float64 // what the target is worth to the attacker: its yield the attacker wants, and rarities it lacks
}

// Appraisal is what a people thinks of a fight.
type Appraisal struct {
	Margin float64 // believed strength difference, attacker minus defender
	Spread float64 // uncertainty of the belief, in levels
	Odds   float64 // on the mean
	Low    float64 // on the pessimistic tail
	High   float64 // on the hopeful tail
	Acted  float64 // what this people acts on, by its risk dial
	Lag    float64 // years for a strike or a fleet to arrive
	Prize  float64 // what the bar drops by for the target's worth
}

// Why says the appraisal in a line.
func (a Appraisal) Why() string {
	s := fmt.Sprintf("odds %.2f (%.2f to %.2f), acting on %.2f, margin %.1f, %.0f years away", a.Odds, a.Low, a.High, a.Acted, a.Margin, a.Lag)
	if a.Prize > 0 {
		s += fmt.Sprintf(", a prize worth %.2f off the bar", a.Prize)
	}
	return s
}

// Appraise estimates a fight: the believed enemy level with terrain on top
// against the attacker's strength, as odds with a spread, and what the
// attacker acts on by its risk.
func Appraise(in AppraiseInput, t *Tuning) Appraisal {
	p := &t.Appraise
	d := in.Believed + in.EnemyBonus + p.Defence
	if in.AtHome {
		d += p.HomeDefence
	}
	if in.Grid {
		d += p.Grid
	}
	d += in.Relief
	if in.Weakened {
		d -= p.Weakened
	}
	for i := 0; i < in.OtherWars; i++ {
		d -= p.OtherWar
	}
	a := Appraisal{Margin: in.Strength - d, Spread: in.Spread, Lag: in.Dist * in.Speed, Prize: min(p.PrizeMax, p.PrizeWeight*in.Prize)}
	a.Odds = phi(a.Margin / p.Scale)
	a.Low = phi((a.Margin - in.Spread) / p.Scale)
	a.High = phi((a.Margin + in.Spread) / p.Scale)
	a.Acted = phi((a.Margin + (2*in.Risk-1)*in.Spread) / p.Scale)
	return a
}

// BarInput is what sets the odds a people needs before it strikes.
type BarInput struct {
	Posture string
	Hates   bool // a xenophobe against the different
	Grudge  bool // holds a grudge against the target
	Aloft   bool // a horde fights from where it is, never by fleet
}

// Bar is the odds a posture needs before it strikes, whether it would
// strike at all, and whether it would send a fleet beyond the front to do it.
func Bar(in BarInput, t *Tuning) (bar float64, wants, far bool) {
	b := &t.Bar
	if in.Posture == Pacifist {
		return 0, false, false
	}
	if in.Hates {
		return b.Hate, true, true
	}
	switch in.Posture {
	case Opportunist:
		bar, wants = b.Opportunist, true
	case Conqueror:
		bar, wants, far = b.Conqueror, true, true
	case Vengeful:
		if in.Grudge {
			bar, wants, far = b.Vengeful, true, true
		}
	}
	if wants && in.Grudge && in.Posture != Vengeful {
		bar -= b.GrudgeDiscount
	}
	if in.Aloft {
		far = false
	}
	return
}
