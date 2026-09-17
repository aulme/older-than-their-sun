package mind

import "fmt"

// Pacts: whom to ask, whether to answer yes, whether to come when called,
// whether to pass a report on.

// PactPlan is what a posture proposes, and how often.
type PactPlan struct {
	Aggressive bool    // partners in war rather than defence
	Rate       float64 // proposals per thousand years
}

// ProposePact is a council's diplomacy: confederates seek defence against
// a threat both can see, conquerors and the vengeful seek partners in war.
func ProposePact(posture string, t *Tuning) PactPlan {
	p := &t.Pact
	switch posture {
	case Confederate:
		return PactPlan{Rate: p.Confederate}
	case Defensive:
		return PactPlan{Rate: p.Defensive}
	case Conqueror, Vengeful:
		return PactPlan{Aggressive: true, Rate: p.Aggressor}
	}
	return PactPlan{Rate: p.Rate}
}

// ThreatInput is one people weighed as a menace.
type ThreatInput struct {
	Believed float64 // its level as believed
	Mil      float64 // one's own
	Hostile  bool    // a posture that strikes first
	Hates    bool    // a xenophobe that finds us different
	AtWar    bool    // in any war
	Rules    bool    // holds slaves or vassals
	Grudge   bool    // we hold a grudge against it
}

// Threatens says whether a people counts as a threat before geometry: at
// least nearly one's own level, and of a kind that strikes.
func Threatens(in ThreatInput, t *Tuning) bool {
	if in.Believed < in.Mil-t.Pact.ThreatSlack {
		return false
	}
	return in.Hostile || in.Hates || in.AtWar || in.Rules || in.Grudge
}

// Near says whether a threat is close enough to matter: it reaches our
// home, has a front against us, or its home is within the two reaches and
// a margin of ours.
func Near(inReach, front bool, homeDist, reaches float64, t *Tuning) bool {
	return inReach || front || homeDist <= reaches+t.Pact.ThreatMargin
}

// Partner says whether a people is worth asking into a pact of a kind: for
// war, only those who strike first or hold a grudge against the target.
func Partner(aggressive, hostile, grudge bool) bool {
	return !aggressive || hostile || grudge
}

// AnswerInput is an offer as the asked sees it.
type AnswerInput struct {
	Aggressive  bool
	Posture     string
	Target      bool    // the offer names an enemy
	Believed    float64 // the enemy's level as believed
	Mil         float64 // one's own
	Grudge      bool    // against the enemy
	AlliedEnemy bool    // already sworn to the enemy
	EnemyNear   bool    // the enemy is met and reaches our home or has a front against us
	AtWar       bool    // with the enemy
	ProposerMil float64
	Difference  float64 // between the two species
	Infamy      float64 // the proposer's
	Renown      float64
	Dials       Dials
}

// Answer is the verdict on an offer.
type Answer struct {
	Accept bool
	Score  float64
	Reason string // when refused out of hand
}

// Why says the answer.
func (a Answer) Why() string {
	if a.Reason != "" {
		return "refused: " + a.Reason
	}
	if a.Accept {
		return fmt.Sprintf("accepted at %.2f", a.Score)
	}
	return fmt.Sprintf("refused at %.2f", a.Score)
}

// AnswerPact weighs an offer: the appraisal with posture on top, less the
// proposer's infamy and the difference between them.
func AnswerPact(in AnswerInput, t *Tuning) Answer {
	p := &t.Pact
	score := 0.0
	if in.Aggressive {
		switch in.Posture {
		case Pacifist, Defensive, Submissive:
			return Answer{Reason: "they make no wars"}
		case Confederate:
			if !in.Target || !in.Grudge {
				return Answer{Reason: "no grudge of their own"}
			}
		}
		if in.Target && in.AlliedEnemy {
			return Answer{Reason: "sworn to the other side"}
		}
		switch in.Posture {
		case Conqueror:
			score += p.Conqueror
		case Opportunist:
			if in.Believed < in.Mil {
				score += p.OpportunistWeak
			} else {
				score -= p.OpportunistWeak
			}
		case Vengeful:
			if in.Target && in.Grudge {
				score += p.VengefulGrudge
			} else {
				score -= p.VengefulNone
			}
		case Unyielding:
			score += p.Unyielding
		}
		score += p.GreedWeight * in.Dials.Greed
	} else {
		if in.Target {
			score += in.Dials.Fear * min(max(p.FearBase+p.FearPerLevel*(in.Believed-in.Mil), 0), p.FearMax)
			if in.EnemyNear {
				score += p.EnemyNear
			}
			if in.AtWar {
				score += p.AtWar
			}
			if in.Grudge {
				score += p.Grudge
			}
		}
		switch in.Posture {
		case Confederate:
			score += p.DefConfederate
		case Defensive:
			score += p.DefDefensive
		case Pacifist, Submissive, Vengeful:
			score += p.DefMeek
		case Opportunist:
			if in.ProposerMil > in.Mil {
				score += p.DefOpportunist
			} else {
				score -= p.DefOpportunist
			}
		case Conqueror:
			score -= p.DefConqueror
		}
	}
	score -= p.Difference * in.Difference
	score -= p.Infamy * in.Infamy
	score += p.Renown * in.Renown
	score += p.Loyalty * (in.Dials.Loyalty - 0.5)
	return Answer{Accept: score > p.Accept, Score: score}
}

// CallInput is a call for help as the ally sees it.
type CallInput struct {
	Mil         float64 // the ally's level at home
	Away        float64 // and out
	VictimMil   float64
	Believed    float64 // the attacker's level as believed
	Confederate bool
	Betrayed    bool // the caller broke faith with us before
	Dials       Dials
}

// Coming is the answer to a call.
type Coming struct {
	Come  bool
	Share float64 // the relief sent
	Want  float64
	Helps bool // relief and the host together would reach the attacker
	Safe  bool // home stays safe without it
	Blame bool // not coming counts as a betrayal
}

// Why says the answer.
func (c Coming) Why() string {
	if c.Come {
		return fmt.Sprintf("comes with %.1f: want %.2f", c.Share, c.Want)
	}
	return fmt.Sprintf("does not come: want %.2f, helps %v, safe %v", c.Want, c.Helps, c.Safe)
}

// AnswerCall is an ally deciding whether to send relief: loyalty against
// fear, when the relief would help and home stays safe.
func AnswerCall(in CallInput, t *Tuning) Coming {
	p := &t.Call
	total := in.Mil + in.Away
	c := Coming{Share: max(p.Share*in.Mil, max(p.Floor*total, p.MinFloor))}
	c.Helps = c.Share+in.VictimMil >= in.Believed-p.HelpSlack
	c.Safe = in.Mil-c.Share >= p.SafeLeft || in.Dials.Fear < p.SafeFear
	c.Want = in.Dials.Loyalty - p.FearWeight*in.Dials.Fear + p.Base
	if in.Confederate {
		c.Want += p.Confederate
	}
	if in.Betrayed {
		c.Want -= p.Betrayed
	}
	c.Come = c.Want > p.Want && c.Helps && c.Safe && c.Share <= in.Mil
	c.Blame = !c.Come && in.Mil >= p.BlameAbove
	return c
}

// ForwardInput is a report weighed for one ally.
type ForwardInput struct {
	AllyAtWar bool    // the ally is at war with the subject, or the pact names it
	AllyMet   bool    // the ally knows the subject
	Hostile   bool    // the forwarder strikes first
	Reported  float64 // the level in the report
	Mil       float64 // the forwarder's own
	Dials     Dials
}

// Forward says whether a report goes to an ally: loyalty times the ally's
// need against greed times the forwarder's own designs on the subject.
func Forward(in ForwardInput, t *Tuning) (bool, float64) {
	p := &t.Forward
	need := p.NeedStranger
	if in.AllyAtWar {
		need = p.NeedEnemy
	} else if in.AllyMet {
		need = p.NeedMet
	}
	designs := 0.0
	if in.Hostile && in.Reported < in.Mil {
		designs = 1
	}
	score := in.Dials.Loyalty*need - in.Dials.Greed*designs*p.Designs
	return score > p.Bar, score
}
