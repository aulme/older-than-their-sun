package mind

// Interception: whether a people that holds a sighting of a fleet tries
// to meet it in the dark, and pickets: whether it keeps a scout out
// watching an enemy. The geometry of the meeting is the caller's; this is
// the will to try.

// InterceptInput is the sighting as the council sees it.
type InterceptInput struct {
	Feasible      bool   // a meeting point exists before the fleet arrives
	AtWar         bool   // with the fleet's people
	ReliefAgainst bool   // the fleet is relief bound to stand with a people this one is at war with
	BoundForUs    bool   // the fleet is bound for one of this people's worlds
	Sending       bool   // this people already has a campaign or relief fleet in flight
	Posture       string // this people's posture
}

// InterceptChoice is the answer.
type InterceptChoice struct {
	Try    bool
	Reason string
}

// Why says the answer.
func (c InterceptChoice) Why() string {
	if c.Try {
		return "meets it in the dark: " + c.Reason
	}
	return "lets it pass: " + c.Reason
}

// Intercept decides whether to meet a sighted fleet: at war with its
// people, or relief bound against this one when this one's posture would
// declare on the sender; a pacifist only for a fleet bound for its own
// worlds; a people already sending a fleet only for one bound for its
// own worlds; and only when a meeting point exists at all.
func Intercept(in InterceptInput, t *Tuning) InterceptChoice {
	switch {
	case !in.Feasible:
		return InterceptChoice{Reason: "nothing could reach its line in time"}
	case !in.AtWar && !(in.ReliefAgainst && Hostile(in.Posture)):
		return InterceptChoice{Reason: "no war with them"}
	case in.Posture == Pacifist && !in.BoundForUs:
		return InterceptChoice{Reason: "it is not coming here"}
	case in.Sending && !in.BoundForUs:
		return InterceptChoice{Reason: "a fleet is already out"}
	case in.BoundForUs:
		return InterceptChoice{Try: true, Reason: "it is bound for one of ours"}
	case in.AtWar:
		return InterceptChoice{Try: true, Reason: "an enemy fleet in the open"}
	}
	return InterceptChoice{Try: true, Reason: "relief for an enemy"}
}

// PicketInput is one enemy as the picket policy sees it.
type PicketInput struct {
	AtWar   bool    // with this enemy
	Front   bool    // this people has a front with it: worlds in strike reach
	Caught  bool    // this people has been caught by a fleet before
	Hostile bool    // the enemy's posture is hostile
	InReach bool    // the enemy can strike this people's home
	Fear    float64 // this people's fear
	Ships   int     // ships manned at home
	Rival   bool    // the enemy is the rival this people builds against: it is watched harder
}

// PicketWant is the answer.
type PicketWant struct {
	Send   bool
	Reason string
}

// Why says the want.
func (p PicketWant) Why() string {
	if p.Send {
		return "keeps a picket: " + p.Reason
	}
	return "keeps no picket: " + p.Reason
}

// WantPicket says whether a people keeps a picket against an enemy: at
// war with one it has no front with, or with any enemy once it has been
// caught by a fleet, or for the fearful against any hostile neighbour in
// reach in peacetime, or against its rival; and a ship must stay home.
func WantPicket(in PicketInput, t *Tuning) PicketWant {
	p := &t.Picket
	var why string
	switch {
	case in.AtWar && !in.Front:
		why = "a war with no front"
	case in.AtWar && in.Caught:
		why = "caught by a fleet before"
	case in.Fear > p.FearBar && in.Hostile && in.InReach:
		why = "a hostile neighbour in reach"
	case in.Rival:
		why = "the rival, watched"
	default:
		return PicketWant{Reason: "nothing to watch for"}
	}
	if float64(in.Ships)-p.KeepHome < 1 {
		return PicketWant{Reason: "a ship must stay home"}
	}
	return PicketWant{Send: true, Reason: why}
}
