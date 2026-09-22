package mind

// Dials are a people's temperament as numbers. They are set from traits at
// birth and nudged by scars, morale and memory, all in the history package;
// every decision here reads dials, never traits.
type Dials struct {
	Aggression float64 `json:"aggression"` // readiness to strike first
	Risk       float64 `json:"risk"`       // acts on hope (high) or on the pessimistic tail (low)
	Greed      float64 `json:"greed"`      // wants worlds
	Fear       float64 `json:"fear"`       // wants the home kept safe
	Loyalty    float64 `json:"loyalty"`    // keeps promises
	Hunger     float64 `json:"hunger"`     // wants to know before acting
	Patience   float64 `json:"patience"`   // holds a course, and a grudge
	Hate       float64 `json:"hate"`       // sees the different as a thing to end
}

// Add sums another set of dials into this one.
func (d *Dials) Add(o Dials) {
	d.Aggression += o.Aggression
	d.Risk += o.Risk
	d.Greed += o.Greed
	d.Fear += o.Fear
	d.Loyalty += o.Loyalty
	d.Hunger += o.Hunger
	d.Patience += o.Patience
	d.Hate += o.Hate
}

// Postures, as the stance trait names them.
const (
	Pacifist    = "pacifist"
	Defensive   = "defensive"
	Submissive  = "submissive"
	Opportunist = "opportunist"
	Conqueror   = "conqueror"
	Vengeful    = "vengeful"
	Confederate = "confederate"
	Unyielding  = "unyielding"
)

// Honours, as the honour trait names them.
const (
	Faithful  = "faithful"
	Practical = "practical"
	Faithless = "faithless"
)

// Hostile is true of postures that strike first.
func Hostile(posture string) bool { return posture == Opportunist || posture == Conqueror }
