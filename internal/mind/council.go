package mind

import "fmt"

// The council is where a people decides whom to strike, whether to scout
// first, or to watch. Each candidate enemy is judged on its own; the
// council then takes the best of those that clear their bar.

// Action is what a council does about one enemy.
type Action uint8

const (
	Nothing    Action = iota
	Strike            // the odds clear the bar
	ScoutFirst        // the spread straddles the bar: a report would change the decision
	Watch             // below the bar
)

func (a Action) String() string { return [...]string{"nothing", "strike", "scout", "watch"}[a] }

// JudgeInput is one enemy before the council.
type JudgeInput struct {
	Appraisal Appraisal
	Bar       float64
	Far       bool // the posture would send a fleet beyond the front
	Front     int  // enemy worlds within strike reach
	Vengeful  bool // against the grudge target the cost is not counted
	Compelled bool // honour compels a lower bar this council
}

// Verdict is the judgment on one enemy.
type Verdict struct {
	Action Action
	Acted  float64 // what the people acted on
	Bar    float64 // the bar it faced
	Low    float64
	High   float64
	Margin float64 // acted less the bar; the council takes the largest
}

// Why says the verdict in a line.
func (v Verdict) Why() string {
	if v.Action == Nothing {
		return "nothing: no front to strike and no fleet to send"
	}
	return fmt.Sprintf("%s: acting on %.2f (%.2f to %.2f) against a bar of %.2f", v.Action, v.Acted, v.Low, v.High, v.Bar)
}

// Judge is the council's view of one enemy: strike when the acted odds
// clear the bar, scout when the spread straddles it, watch below it.
// Nothing when there is no front and the posture sends no fleet. A prize
// lowers the bar: a rich neighbour is worth more to a wanting people; a
// partner that sends raises it.
func Judge(in JudgeInput, t *Tuning) Verdict {
	v := Verdict{Bar: min(1, max(0, in.Bar-in.Appraisal.Prize)), Acted: in.Appraisal.Acted, Low: in.Appraisal.Low, High: in.Appraisal.High}
	if in.Compelled {
		v.Bar = min(v.Bar, t.Council.Compelled)
	}
	if in.Front == 0 && !in.Far {
		return v
	}
	if in.Vengeful {
		v.Acted = in.Appraisal.High
	}
	v.Margin = v.Acted - v.Bar
	switch {
	case v.Acted >= v.Bar:
		v.Action = Strike
	case v.Low < v.Bar && v.High > v.Bar:
		v.Action = ScoutFirst
	default:
		v.Action = Watch
	}
	return v
}

// Council picks whom to strike among judged enemies: the largest margin
// over its bar, or -1 when none clears it. Ties go to the earlier.
func Council(verdicts []Verdict) int {
	best := -1
	for i, v := range verdicts {
		if v.Action != Strike {
			continue
		}
		if best < 0 || v.Margin > verdicts[best].Margin {
			best = i
		}
	}
	return best
}

// ScoutInput is whether a report is worth a level.
type ScoutInput struct {
	Sight bool    // holds the Sight and it is on the borders: it reads for free
	Mil   float64 // level at home
	Fear  float64
	Out   bool // a scout is already on its way to them
}

// ScoutChoice is the answer.
type ScoutChoice struct {
	Look bool // read them with the Sight
	Send bool // send a scout
	Kept bool // the level could not be spared
}

// Why says the choice.
func (s ScoutChoice) Why() string {
	switch {
	case s.Look:
		return "the Sight reads them"
	case s.Send:
		return "a scout is sent"
	case s.Kept:
		return "no level can be spared for a scout"
	}
	return "a scout is already out"
}

// Scout sends a scout when the level can be spared and none is out; the
// Sight reads instead.
func Scout(in ScoutInput, t *Tuning) ScoutChoice {
	s := &t.Scout
	if in.Sight {
		return ScoutChoice{Look: true}
	}
	if in.Mil-1 < s.KeepHome || (in.Fear > s.FearBar && in.Mil < s.FearMil) {
		return ScoutChoice{Kept: true}
	}
	if in.Out {
		return ScoutChoice{}
	}
	return ScoutChoice{Send: true}
}

// CampaignInput is the sizing of a fleet against one world.
type CampaignInput struct {
	Appraisal Appraisal
	Strength  float64 // the attacker's, as appraised
	Mil       float64 // level at home
	Away      float64 // level already out
	Risk      float64
	Fear      float64
	Conqueror bool
}

// Campaign is the sized fleet.
type Campaign struct {
	Send  bool
	Share float64 // the level to send
	Need  float64 // what even odds would take, by belief and risk
	Floor float64
	Cap   float64
	LagOK bool
}

// Why says the sizing.
func (c Campaign) Why() string {
	verdict := "and it goes"
	if !c.Send {
		verdict = "and it stays home"
	}
	return fmt.Sprintf("a fleet would need %.1f, at least %.1f, at most %.1f, crossing in time %v, %s", c.Need, c.Floor, c.Cap, c.LagOK, verdict)
}

// SizeCampaign sizes a fleet: what the belief says even odds would take,
// tilted by risk, at least a share of the whole, at most what fear leaves.
// Nobody sends a fleet that cannot take its first world, or one that
// would cross for too long.
func SizeCampaign(in CampaignInput, t *Tuning) Campaign {
	p := &t.Campaign
	def := in.Strength - in.Appraisal.Margin
	c := Campaign{Need: def - (2*in.Risk-1)*in.Appraisal.Spread}
	total := in.Mil + in.Away
	c.Floor = max(p.Floor*total, p.MinFloor)
	c.Cap = in.Mil * (1 - p.FearCap*in.Fear)
	c.Share = min(max(c.Need, c.Floor), c.Cap)
	c.LagOK = in.Appraisal.Lag < p.MaxLag || (in.Conqueror && in.Appraisal.Lag < p.ConquerorLag)
	c.Send = !(c.Share < c.Need || c.Share < c.Floor || !c.LagOK)
	return c
}
