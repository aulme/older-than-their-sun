package history

import (
	"math"
	"math/rand/v2"

	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// Leaders (specs/proposals/leaders.md). Everything else in the history is
// done by a whole people; a leader is a named figure a people follows for
// a while, and what its loss costs. It is rare and conditional: the
// conditions make one possible (a war, a dark age just past, a
// renaissance's surge, an heir's first years) and when one arises it is
// named, so a leader is evidence about the state of the galaxy rather
// than a die rolled on top of it.
//
// A leader moves the disposition, never the decision procedure. Its bent
// is a term in the dials larger than the telling's, drawn mostly along
// the people's own lean and against it often enough to be noticed, and
// the stance the council reads is read off it: a peaceful people can be
// led to conquest and a warlike one to its walls, and when the leader is
// gone it snaps back. The council still decides, on other inputs.
//
// It is somewhere: with the greatest fleet at war, where it adds a large
// bonus to that fleet and dies with it, or at the seat, where it adds a
// small one to the whole realm and dies only with the capital. Which is
// its temperament, drawn from the dials with noise; one in four or so
// runs rather than falls, once, and reigns on diminished. Immortality
// is the absence of a natural death, not invulnerability.
//
// Its loss is the Succession, a filter faced the next time the people
// acts: overcome it holds, scarred it holds drawn in on the seat,
// declined it breaks as ossification breaks, a civil war or a dark age.
// Its difficulty is continuity's (a realm with no institution behind the
// person comes apart) and how far the leader had pushed the people from
// its own bent. It is ossification's punctual sibling: the same three
// outcomes and the same break, the opposite timing.
//
// A deathless leader awake at war and at its people's crimes long enough
// narrows what its people counts as wrong, to the taking of worlds and
// then to nothing, and then it is mad: war on everything, purges that eat
// the realm's continuity, and the realm breaking under it once there is
// nobody left to fight. A people that made death sacred has no deathless
// ruler.

// Leader is one named figure a people follows. Its name is a row of the
// names pass, keyed by its id; leaders belong to the world, like wars.
type Leader struct {
	ID        int
	Civ       int
	Rose      Year
	Ended     Year    // 0 while it reigns
	End       string  // how it ended: a key of data/leaders.json's ends, "" while it reigns
	Occasion  string  // what made it possible: a key of leaders.json's occasions
	Form      string  // what it is: a key of leaders.json's forms
	Stance    string  // the stance it brings, which the council reads in place of the people's own
	Own       string  // the people's own stance at the rising, what it snaps back to
	Bent      Dials   // its term in the dials: at most Size a dial either way
	Push      float64 // how far it turned the people against its own bent, nought to one
	Front     bool    // the temperament: with the greatest fleet at war, or at the seat
	Flees     bool    // runs rather than falls, once
	Fled      bool    // has run: the bent halved, no bonus, and the next fall is the last
	Fleet     int     // the fleet it rides with, -1 at the seat
	Star      int     // where it is: the fleet's star, or the seat
	Until     Year    // when its span, or its line, runs out; 0 for never
	Deathless bool    // no natural end: a mind that does not turn over, not one kept on ice
	Mad       float64 // the slide: at a half its people's good narrows to conquest, at one to nothing
	Faced     string  // the succession after it: overcome, scarred or declined; "" if none was faced
	Doublings float64 // continuity's term when the succession was faced: the loss over the middle, on a log scale
	Battles   int     // battles fought at its side
	Stood     Year    // when it last stood at a world fought over
}

// Reigns says whether the leader still leads.
func (l *Leader) Reigns() bool { return l.End == "" }

// stanceLadder is the stances from peace to conquest, by rung; rungOf is
// where each of the eight stands.
var stanceLadder = []string{mind.Pacifist, mind.Defensive, mind.Unyielding, mind.Opportunist, mind.Conqueror}

func rungOf(stance string) int {
	switch stance {
	case mind.Pacifist:
		return 0
	case mind.Submissive, mind.Defensive:
		return 1
	case mind.Opportunist:
		return 3
	case mind.Conqueror:
		return 4
	}
	return 2 // unyielding, confederate, vengeful: the middle, where posture alone starts no war
}

// dialsOf is the eight dials in a fixed order, for the draws.
func dialsOf(d *Dials) [8]*float64 {
	return [8]*float64{&d.Aggression, &d.Risk, &d.Greed, &d.Fear, &d.Loyalty, &d.Hunger, &d.Patience, &d.Hate}
}

// leaderStream is a people's leaders' own stream: whether one rises, its
// bent and its temperament are drawn from it, so a people's history is
// the same draw for draw until it has a leader.
func (w *World) leaderStream(c *Civ) *rand.Rand { return w.stream("leader:" + itoa(c.ID)) }

// leaderStep is the people's turn at its leader: the succession of the
// one lost since it last acted, the reign of the one it has, or the
// chance of one.
func (w *World) leaderStep(c *Civ) {
	if c.Bereft != nil {
		w.succession(c)
		return
	}
	if l := c.Leader; l != nil {
		w.reign(c, l)
		return
	}
	w.maybeRaise(c)
}

// occasion is what makes a leader possible now, and the rate a thousand
// years it arises at: an heir's first years (the winner of a civil war,
// the founder of a line), a dark age just past, a renaissance's surge, a
// war. None is no leader.
func (w *World) occasion(c *Civ) (string, float64) {
	t := &w.Cfg.Tuning.Leaders
	switch {
	case len(c.Line) > 0 && float64(w.Now-c.Born)/1000 < t.FoundingKyr:
		return "founding", t.Founding
	case c.DarkAges > 0 && float64(w.Now-c.LastDark)/1000 < t.CrisisKyr:
		return "crisis", t.Crisis
	case w.renewing(c):
		return "reform", t.Reform
	case len(c.Wars) > 0:
		return "war", t.War
	}
	return "", 0
}

// leads says whether a people can have a leader: one that has reached
// the stars, is its own master, and has someone to follow.
func (w *World) leads(c *Civ) bool {
	return c.Active() && c.Free() && c.Starfaring != 0 && c.Species.Profile().Can(species.Leads) && !c.Asleep
}

func (w *World) maybeRaise(c *Civ) {
	if !w.leads(c) {
		return
	}
	why, rate := w.occasion(c)
	if rate <= 0 {
		return
	}
	r := w.leaderStream(c)
	if r.Float64() >= 1-math.Exp(-rate*w.dt) {
		return
	}
	w.raiseLeader(c, why, r)
}

// raiseLeader names a leader: its bent drawn along the people's own lean, the
// stance read off it, its temperament from the dials it makes, its span
// from the people's bodies.
func (w *World) raiseLeader(c *Civ, why string, r *rand.Rand) *Leader {
	t := &w.Cfg.Tuning.Leaders
	life := w.lifespanOf(c)
	l := &Leader{ID: len(w.Leaders), Civ: c.ID, Rose: w.Now, Occasion: why, Form: w.formOf(c, life), Own: c.ownPosture(), Fleet: -1, Star: c.Home}
	own := c.Dials
	with := own
	ownP, withP, bentP := dialsOf(&own), dialsOf(&with), dialsOf(&l.Bent)
	against, rest := 0.0, math.Sqrt(1-t.Corr*t.Corr)
	var lean [8]float64
	for i := range ownP {
		u := clamp((*ownP[i]-0.5)/0.45, -1, 1) // the people's own lean on the dial
		b := clamp(t.Corr*u+rest*r.NormFloat64(), -1, 1)
		lean[i] = b
		*bentP[i] = t.Size * b
		*withP[i] = clamp(*withP[i]+t.Size*b, 0.05, 0.95)
		if u*b < 0 {
			against += math.Abs(b) / 8 // the part of the bent against the people's own lean
		}
	}
	rung := min(4, max(0, rungOf(l.Own)+int(math.Round(t.Rungs*lean[0]))))
	l.Stance = l.Own
	if rung != rungOf(l.Own) {
		l.Stance = stanceLadder[rung]
	}
	flip := 0.0
	if mind.Hostile(l.Stance) != mind.Hostile(l.Own) {
		flip = 0.5
	}
	l.Push = clamp(against+flip, 0, 1)
	l.Front = c.launches() && with.Aggression+with.Risk-with.Fear+0.2*r.NormFloat64() > t.Front
	l.Flees = r.Float64() < t.Flee*with.Fear/0.5
	switch {
	case life == 0 && !c.Scars[ScarMortality]:
		l.Deathless = true // nothing in it turns over: it reigns until it is ended
	default:
		reign := t.LineKyr * 1000 * math.Exp(t.LineSpread*r.NormFloat64())
		if life > 0 {
			reign = max(reign, t.SpanShare*life*(0.5+r.Float64())) // a long-lived ruler outlasts a short-lived line
		}
		if c.Known["hibernation"] {
			reign *= t.Hibernate // kept on ice between the crises
		}
		l.Until = w.Now + Year(reign)
	}
	w.Leaders = append(w.Leaders, l)
	c.Leader = l
	w.told(FLeader, c, nil, c.Home).with(P{"leader": l.ID, "occasion": why})
	w.recompute(c)
	return l
}

// formOf is what a leader is, by the people's nature and the span of its
// bodies: a line where the founder dies inside a thousand years.
func (w *World) formOf(c *Civ, life float64) string {
	switch {
	case c.Species.Is(species.Hive):
		return "brood"
	case c.Species.Sub == species.Machine:
		return "directive"
	case c.Species.Sub == species.Eldritch:
		return "doctrine"
	case life > 0 && life < 1000:
		return "line"
	}
	return "person"
}

// reign is a leader's tick: its span, where it stands, and a deathless
// one's slide.
func (w *World) reign(c *Civ, l *Leader) {
	t := &w.Cfg.Tuning.Leaders
	if l.Until > 0 && w.Now >= l.Until {
		w.loseLeader(c, "died", c.Home)
		w.succession(c)
		return
	}
	l.Fleet, l.Star = -1, c.Home
	if l.Front && !l.Fled {
		if x := w.greatestAtWar(c); x != nil {
			l.Fleet, l.Star = x.ID, x.Star
			if x.Base >= 0 {
				l.Star = x.Base
			}
		}
	}
	if !l.Deathless {
		return
	}
	if len(c.Wars) > 0 {
		l.Mad += t.MadWar * w.dt
	}
	w.slide(c, l)
	if l.Mad >= 1 && len(c.Wars) == 0 && w.leaderStream(c).Float64() < 1-math.Exp(-t.MadBreak*w.dt) {
		// nobody left to fight: the purges turn inward until the realm breaks
		w.loseLeader(c, "overthrown", c.Home)
		c.Bereft = nil // the break is the succession
		w.breakDown(c, because("mad_leader"))
	}
}

// slide narrows a deathless leader's people's good as the madness grows:
// at a half to the taking of worlds, at one to nothing, and then the
// leader is mad and makes war on everything.
func (w *World) slide(c *Civ, l *Leader) {
	t := &w.Cfg.Tuning.Leaders
	narrowed := Morality{Kind: Fixation, Object: Conquest}
	switch {
	case l.Mad >= 1 && c.Morality.Kind != Amoral:
		c.think(Morality{Kind: Amoral})
		w.event(KMorality, c, nil, -1, P{"way": "mad", "morality": c.Morality, "leader": l.ID})
		l.Stance = mind.Conqueror
		l.Bent.Aggression, l.Bent.Hate, l.Bent.Fear, l.Bent.Loyalty = t.Size, t.Size, t.Size, -t.Size
		w.recompute(c)
	case l.Mad >= 0.5 && l.Mad < 1 && c.Morality != narrowed && c.Morality.Kind != Amoral:
		c.think(narrowed)
		w.event(KMorality, c, nil, -1, P{"way": "narrowed", "morality": c.Morality, "leader": l.ID})
	}
}

// crime is a crime of the people under a deathless leader: a step of its
// slide.
func (w *World) crime(c *Civ) {
	if l := c.Leader; l != nil && l.Deathless {
		l.Mad += w.Cfg.Tuning.Leaders.MadCrime
	}
}

// joins is a campaign launched: a leader whose place is the front goes
// with it if it is the greatest out.
func (w *World) joins(c *Civ, x *Expedition) {
	l := c.Leader
	if l == nil || !l.Front || l.Fled {
		return
	}
	if l.Fleet >= 0 {
		if cur := w.Expeditions[l.Fleet]; !cur.Over && cur.Ships >= x.Ships {
			return
		}
	}
	l.Fleet, l.Star = x.ID, x.From
}

// greatestAtWar is a people's greatest fleet campaigning: the one a
// leader at the front rides with.
func (w *World) greatestAtWar(c *Civ) *Expedition {
	var best *Expedition
	for _, x := range w.fleetsOf(c) {
		if (x.Kind == Campaign || x.Kind == Roam) && x.Ships > 0 && (best == nil || x.Ships > best.Ships) {
			best = x
		}
	}
	return best
}

// ownPosture is the people's own stance, from its traits, with no leader.
func (c *Civ) ownPosture() string {
	for _, t := range c.Species.Traits {
		if t.Group == "stance" {
			return t.Key
		}
	}
	return mind.Defensive
}

// leaderBent is the leader's term in the dials: halved once it has run.
func (c *Civ) leaderBent() Dials {
	l := c.Leader
	if l == nil {
		return Dials{}
	}
	b := l.Bent
	if l.Fled {
		for _, p := range dialsOf(&b) {
			*p *= 0.5
		}
	}
	return b
}

// leaderLevels is what a leader at the seat adds across the realm.
func (w *World) leaderLevels(c *Civ) (mil, soc float64) {
	l := c.Leader
	if l == nil || l.Fled || l.Fleet >= 0 {
		return 0, 0
	}
	t := &w.Cfg.Tuning.Leaders
	return t.CapitalMil, t.CapitalSoc
}

// fleetQuality is a fleet's quality in a battle: its owner's, with the
// levels of a leader riding with it.
func (w *World) fleetQuality(c *Civ, x *Expedition) float64 {
	if l := c.Leader; l != nil && l.Fleet == x.ID && !l.Fled {
		return w.qualityAt(c, w.Cfg.Tuning.Leaders.FrontLevels)
	}
	return w.quality(c)
}

// fought is a battle the fleet a leader rides with has fought: a broken
// fleet takes it always, a beaten one with a chance by the ships lost,
// a winning one seldom.
func (w *World) fought(c *Civ, x *Expedition, star int, won bool, lost, had int) {
	l := c.Leader
	if l == nil || l.Fleet != x.ID {
		return
	}
	l.Battles++
	t := &w.Cfg.Tuning.Leaders
	switch {
	case x.Ships <= 0:
		w.falls(c, "fell_field", star)
	case !won && had > 0 && w.R.Float64() < t.FieldRisk*float64(lost)/float64(had):
		w.falls(c, "fell_field", star)
	case won && w.R.Float64() < t.WonRisk:
		w.falls(c, "fell_field", star)
	}
}

// stands is a battle at one of a people's worlds: a leader whose place is
// the front and who rides with no campaign comes to it, the first fought
// over in the tick, and stands with the guns and the guard there.
func (w *World) stands(e *Civ, t int) {
	l := e.Leader
	if l == nil || !l.Front || l.Fled {
		return
	}
	if l.Fleet >= 0 {
		if x := w.Expeditions[l.Fleet]; !x.Over && x.Ships > 0 {
			return // out with a campaign
		}
	}
	if l.Fleet < 0 && l.Star != e.Home && l.Star != t && l.Stood == w.Now {
		return // standing at another world this tick
	}
	l.Fleet, l.Star, l.Stood = -1, t, w.Now
}

// skyQuality is the quality of what holds a people's world: its own, with
// the levels of a leader standing there.
func (w *World) skyQuality(e *Civ, t int) float64 {
	if l := e.Leader; l != nil && l.Front && !l.Fled && l.Fleet < 0 && l.Star == t && l.Stood == w.Now {
		return w.qualityAt(e, w.Cfg.Tuning.Leaders.FrontLevels)
	}
	return w.quality(e)
}

// heldWith is a battle fought at a world a leader stands at: a beaten sky
// with a chance by what was lost, a held one seldom. A world taken takes
// it always (seatFalls).
func (w *World) heldWith(e *Civ, t int, held bool, lost, had int) {
	l := e.Leader
	if l == nil || !l.Front || l.Fleet >= 0 || l.Star != t || l.Stood != w.Now {
		return
	}
	l.Battles++
	tn := &w.Cfg.Tuning.Leaders
	switch {
	case !held && had > 0 && w.R.Float64() < tn.FieldRisk*float64(lost)/float64(had):
		w.falls(e, "fell_field", t)
	case held && w.R.Float64() < tn.WonRisk:
		w.falls(e, "fell_field", t)
	}
}

// fleetGone is a fleet lost with a leader aboard.
func (w *World) fleetGone(x *Expedition, star int) {
	c := w.Civs[x.Owner]
	if l := c.Leader; l != nil && l.Fleet == x.ID {
		w.falls(c, "fell_field", star)
	}
}

// seatFalls is a world lost with a leader in it: the capital, or the
// world a leader at the front was standing at.
func (w *World) seatFalls(c *Civ, s int) {
	l := c.Leader
	if l == nil || l.Fleet >= 0 || !c.Active() {
		return
	}
	switch {
	case l.Front && l.Star == s && l.Stood == w.Now:
		w.falls(c, "fell_field", s)
	case s == c.Home:
		w.falls(c, "fell_capital", s)
	}
}

// falls is a leader's place taken or destroyed: it runs if it is the kind
// that runs and has not run before, and otherwise dies with it.
func (w *World) falls(c *Civ, end string, star int) {
	l := c.Leader
	if l.Flees && !l.Fled {
		l.Fled, l.Fleet, l.Star = true, -1, c.Home
		w.event(KLeaderFled, c, nil, star, P{"leader": l.ID})
		return
	}
	w.loseLeader(c, end, star)
}

// loseLeader is a leader lost: the fact, the people itself again, and
// the succession owed the next time the people acts.
func (w *World) loseLeader(c *Civ, end string, star int) {
	l := c.Leader
	if l == nil {
		return
	}
	l.Ended, l.End = w.Now, end
	c.Leader = nil
	c.Bereft = l
	w.told(FLeaderLost, c, nil, star).with(P{"leader": l.ID, "end": end})
	for _, eid := range sortedInts(c.Wars) {
		if wr := w.warBetween(c.ID, eid); wr != nil {
			wr.Will[wr.side(c.ID)] -= w.Cfg.Tuning.War.LeaderLost // the succession is its own crisis
			w.summon(wr)
		}
	}
	if mind.Hostile(l.Stance) != mind.Hostile(l.Own) {
		w.event(KSnappedBack, c, nil, -1, P{"leader": l.ID}) // it had turned the people to war or from it
	}
	w.setDials(c)
}

// endLeader is a leader ended with no story of its own: the people fell
// or ended under it, or was taken and kept. No succession is owed.
func (w *World) endLeader(c *Civ, end string) {
	l := c.Leader
	if l == nil {
		return
	}
	l.Ended, l.End = w.Now, end
	c.Leader = nil
	c.Bereft = nil
	w.note(KLeaderEnded, c, nil, -1, P{"leader": l.ID, "end": end})
}

// succession is the filter faced for the leader lost: its difficulty is
// what the realm has behind the person, continuity, and how far the
// leader had turned the people from itself.
func (w *World) succession(c *Civ) {
	l := c.Bereft
	c.Bereft = nil
	if l == nil || !c.Active() {
		return
	}
	w.succeeding = l
	defer func() { w.succeeding = nil }()
	l.Doublings = w.doublings(c)
	l.Faced = [...]string{"overcome", "scarred", "declined"}[w.face(c, "succession", w.successionAdj(c, l))]
}

// successionAdj is what the succession after a leader adds to the
// filter's difficulty: the base, how far it pushed the people from its
// own bent, and continuity's term; less for a mind with a backup copy.
func (w *World) successionAdj(c *Civ, l *Leader) float64 {
	t := &w.Cfg.Tuning.Leaders
	adj := t.Diff + t.Push*l.Push + t.Doubling*w.doublings(c)
	if c.Known["forking"] {
		adj -= t.Forking
	}
	return adj
}

func init() {
	def(&Filter{
		Key: "succession",
		Overcome: func(w *World, c *Civ) {
			w.faced(c, "succession", "overcome", "", -1).with(P{"leader": w.succeeding.ID})
		},
		Scar: func(w *World, c *Civ) {
			w.scar(c, ScarCentralism) // it holds, drawn in on the seat and frightened
			w.faced(c, "succession", "scarred", "", c.Home).with(P{"leader": w.succeeding.ID})
		},
		Decline: func(w *World, c *Civ) {
			w.faced(c, "succession", "declined", "", c.Home).with(P{"leader": w.succeeding.ID})
			if !c.Species.Profile().Can(species.CivilWars) && len(c.Systems) > 1 {
				// a people with no factions does not split: the far world is cut from the seat
				far, fd := -1, 0.0
				for _, s := range c.Systems {
					if d := w.G.Dist(c.Home, s); s != c.Home && d > fd {
						far, fd = s, d
					}
				}
				if w.cutOff(c, far) != nil {
					return
				}
			}
			w.breakDown(c, because("succession"))
		},
	})
}

// LeadersOf is the batch's reading of the leaders: how many rose, how
// many reigned deathless, how many went mad, and how each ended, by key.
func LeadersOf(w *World) (rose, deathless, mad int, ends map[string]int) {
	ends = map[string]int{}
	for _, l := range w.Leaders {
		rose++
		if l.Deathless {
			deathless++
		}
		if l.Mad >= 1 {
			mad++
		}
		if l.End != "" {
			ends[l.End]++
		}
	}
	return
}
