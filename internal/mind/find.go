package mind

import "fmt"

// The Find's choice: what a people attempts with a remain, as weights
// over mastering it, wielding it and sealing it away.

// FindInput is the people and the remain.
type FindInput struct {
	Curious, Expansionist, Symbiotic, Cautious, Xenophobic, Contemplative, Pragmatic, Conqueror bool
	Threat                                                                                      bool // a threat or a sleeper: never wielded
	Plain                                                                                       bool // a miracle artifact: plainly a thing to be used
	Own                                                                                         bool // the people's own lost work
	Ruin                                                                                        bool // a ruin: nothing to use, nothing to guard
	Law                                                                                         bool // a place where the state beneath shows through: only to be used
	OldThings                                                                                   bool // the people is fixed on what the old ones left: eager to master
}

// Attempt is the weights.
type Attempt struct{ Master, Wield, Seal float64 }

// Why says the weights.
func (a Attempt) Why() string {
	return fmt.Sprintf("master %.1f, wield %.1f, seal %.1f", a.Master, a.Wield, a.Seal)
}

// Total is the sum of the weights.
func (a Attempt) Total() float64 { return a.Master + a.Wield + a.Seal }

// Find weighs the three attempts by temperament and by what the remain is.
func Find(in FindInput, t *Tuning) Attempt {
	p := &t.Find
	a := Attempt{p.Master, p.Wield, p.Seal}
	if in.Curious || in.OldThings {
		a.Master += p.Curious
	}
	if in.Expansionist || in.Symbiotic {
		a.Master += p.Reaching
	}
	if in.Cautious {
		a.Seal += p.Cautious
	}
	if in.Xenophobic || in.Contemplative {
		a.Seal += p.Wary
	}
	if in.Pragmatic || in.Conqueror {
		a.Wield += p.Practical
	}
	if in.Threat {
		a.Wield = 0
		a.Seal += p.ThreatSeal
	}
	if in.Plain {
		a.Wield += p.Plain
	}
	if in.Own {
		a.Master += p.Own
		a.Seal = 0
	}
	if in.Ruin {
		a.Wield, a.Seal = 0, 0
		a.Master = p.Master
	}
	if in.Law {
		a.Master, a.Seal = 0, 0
	}
	return a
}

// Pick turns a roll in [0,1) into the attempt: 0 master, 1 wield, 2 seal.
func (a Attempt) Pick(roll float64) int {
	x := roll * a.Total()
	switch {
	case x < a.Master:
		return 0
	case x < a.Master+a.Wield:
		return 1
	}
	return 2
}
