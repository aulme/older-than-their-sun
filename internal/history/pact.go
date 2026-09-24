package history

import (
	"math"

	"worldgen/internal/battle"
	"worldgen/internal/mind"
	"worldgen/internal/species"
)

// Pacts, messages and reputation. Diplomacy travels at light speed: a
// proposal, a call for help or a report is a message that arrives after
// the distance in years, unless the sender holds the Voice. A pact has a
// kind and a target. Defence is a relief fleet. Betrayal is recorded on the
// world, and everyone weighs it.

// PactKind is what a pact binds its members to.
type PactKind uint8

const (
	Defensive PactKind = iota
	Aggressive
	Both
)

func (k PactKind) String() string { return [...]string{"defence", "war", "defence and war"}[k] }

// Pact is an alliance.
type Pact struct {
	ID      int
	Members []int
	Kind    PactKind
	Target  int // the people it is against, -1 for whoever comes
	Formed  Year
	Ended   Year
	Over    bool
}

// MsgKind is what a message carries.
type MsgKind uint8

const (
	MsgPact MsgKind = iota
	MsgCall
	MsgIntel
	MsgNews     // a fact, with the teller's slant
	MsgSighting // a fleet seen in flight, passed on
	MsgOffer    // a contract proposed; see contract.go
	MsgAnswer   // the answer to one, when it was yes
	MsgTeach    // a node taught under one
)

// Message is one thing said across the dark.
type Message struct {
	From, To int
	Kind     MsgKind
	Sent     Year
	Arrive   Year
	Pact     int      // an existing pact to join, -1
	PactKind PactKind // for proposals
	Target   int      // the enemy, for proposals and calls
	About    int      // the subject of a report
	Intel    *Intel
	Sighting *Sighting
	Fact     int    // for news
	Slant    int8   // the teller's regard for the other party in it
	Contract int    // for offers, answers and teaching
	Node     string // for teaching
}

// Betrayal is a promise broken, or with negative weight, kept at a cost.
type Betrayal struct {
	By, Against int
	Year        Year
	Shape       string // a key of data/causes.json's betrayals
	Weight      float64
}

// send queues a message with the lag of light.
func (w *World) send(from, to *Civ, m *Message) {
	m.From, m.To, m.Sent = from.ID, to.ID, w.Now
	lag := 0.0
	if !from.miracle("ansible") {
		_, lag = w.nearest(from, to.Home)
	}
	m.Arrive = w.Now + Year(lag)
	w.Messages = append(w.Messages, m)
}

// tickMessages delivers what has arrived. A message from a sender the
// recipient does not fathom is dropped unread: intel, news, a call, a
// pact, an offer.
func (w *World) tickMessages() {
	pending := w.Messages
	w.Messages = nil
	var keep []*Message
	for _, m := range pending {
		if m.Arrive > w.Now {
			keep = append(keep, m)
			continue
		}
		from, to := w.Civs[m.From], w.Civs[m.To]
		if !to.Active() || (!from.Living() && m.Kind != MsgNews) {
			continue
		}
		if !to.Fathomed[from.ID] {
			to.Tally.Dropped++
			continue // a message from a people not understood means nothing on arrival
		}
		if w.shutTo(to, from) {
			continue // dropped unread: nothing in it is heard, and nothing in it is caught
		}
		if from.Active() {
			w.expose(from, to, "message")
		}
		switch m.Kind {
		case MsgNews:
			w.news(to, from, m)
		case MsgIntel:
			to.receive(m.About, m.Intel)
		case MsgSighting:
			w.receiveSighting(to, m.Sighting)
		case MsgPact:
			w.answerPact(to, from, m)
		case MsgCall:
			if m.Target >= 0 {
				w.answerCall(to, from, w.Civs[m.Target])
			}
		case MsgOffer:
			w.answerOffer(to, from, m)
		case MsgAnswer:
			w.answered(to, from, m)
		case MsgTeach:
			w.taughtNode(to, from, m)
		}
	}
	w.Messages = append(w.Messages, keep...)
}

// allied says whether two peoples share a pact.
func (w *World) allied(c, e *Civ) bool {
	for _, pid := range c.Pacts {
		p := w.Pacts[pid]
		if !p.Over && contains(p.Members, e.ID) {
			return true
		}
	}
	return false
}

// pactWith finds the pact two peoples share, or nil.
func (w *World) pactWith(c, e *Civ) *Pact {
	for _, pid := range c.Pacts {
		p := w.Pacts[pid]
		if !p.Over && contains(p.Members, e.ID) {
			return p
		}
	}
	return nil
}

// threat is the people c most fears in reach, or nil.
func (w *World) threat(c *Civ) *Civ {
	var worst *Civ
	worstMil := 0.0
	for _, eid := range metOf(c) {
		e := w.Civs[eid]
		if !e.Active() || !e.Free() || w.allied(c, e) || e.Master == c.ID {
			continue
		}
		mil, _ := w.believe(c, e)
		if !mind.Threatens(mind.ThreatInput{Believed: mil, Mil: c.Mil, Hostile: e.hostile(), Hates: e.hates(c), AtWar: len(e.Wars) > 0, Rules: e.Ruled > 0, Grudge: c.Grudge[eid] > 0}, w.Cfg.Tuning) {
			continue
		}
		if !w.inReach(e, c.Home) && len(w.front(e, c)) == 0 && w.G.Dist(c.Home, e.Home) > e.Reach+c.Reach+w.Cfg.Tuning.Pact.ThreatMargin {
			continue
		}
		if worst == nil || mil > worstMil {
			worst, worstMil = e, mil
		}
	}
	return worst
}

// proposePact is a council's diplomacy: confederates seek defence against
// a threat both can see, conquerors and the vengeful seek partners in war.
// Only a people that understands and is understood is asked.
func (w *World) proposePact(c *Civ) {
	plan := mind.ProposePact(c.posture(), w.Cfg.Tuning)
	kind := Defensive
	if plan.Aggressive {
		kind = Aggressive
	}
	var target *Civ
	if w.R.Float64() > plan.Rate {
		return
	}
	if c.Stiff > 1 && !w.swornKind(c, kind) {
		return // a stiff people makes no new kind of promise
	}
	if kind == Aggressive {
		for _, eid := range metOf(c) {
			e := w.Civs[eid]
			if !e.Active() || !e.Free() || w.allied(c, e) || c.Truce[eid] > w.Now {
				continue
			}
			if _, wants, _ := w.bar(c, e); wants {
				target = e
				break
			}
		}
	} else {
		target = w.threat(c)
	}
	if target == nil {
		return
	}
	for _, fid := range metOf(c) {
		f := w.Civs[fid]
		if f == target || !f.Active() || !f.Free() || f.Wars[c.ID] || w.allied(c, f) || c.hates(f) || f.hates(c) || !w.mutual(c, f) {
			continue
		}
		if c.Asked[fid]+Year(w.Cfg.Tuning.Pact.AskAgain) > w.Now {
			continue
		}
		if !mind.Partner(plan.Aggressive, f.hostile(), f.Grudge[target.ID] > 0) {
			continue
		}
		if w.allied(f, target) {
			continue
		}
		c.Asked[fid] = w.Now
		pid := -1
		for _, x := range c.Pacts {
			if p := w.Pacts[x]; !p.Over && p.Kind == kind && p.Target == target.ID {
				pid = x
			}
		}
		w.send(c, f, &Message{Kind: MsgPact, PactKind: kind, Target: target.ID, Pact: pid})
		return
	}
}

// swornKind says whether a people has ever held a pact of a kind.
func (w *World) swornKind(c *Civ, kind PactKind) bool {
	for _, pid := range c.Pacts {
		if p := w.Pacts[pid]; p.Kind == kind || p.Kind == Both {
			return true
		}
	}
	return false
}

// answerPact is a people weighing an offer: the appraisal with posture on
// top, less the proposer's infamy, the score read through its folly.
func (w *World) answerPact(f, c *Civ, m *Message) {
	if !f.Active() || !c.Active() || f.hates(c) || c.hates(f) || f.Wars[c.ID] || w.allied(f, c) || !w.mutual(f, c) {
		return
	}
	var e *Civ
	if m.Target >= 0 {
		e = w.Civs[m.Target]
		if !e.Living() {
			return
		}
	}
	in := mind.AnswerInput{
		Aggressive: m.PactKind == Aggressive, Posture: f.posture(), Target: e != nil, Mil: f.Mil, ProposerMil: c.Mil,
		Difference: f.differs(c), Infamy: w.infamy(c), ProposerGrudge: f.Grudge[c.ID], Renown: w.renown(c), Dials: f.Dials,
		Wis: f.Wis, Noise: w.R.NormFloat64(), Kin: w.kin(f, c),
	}
	against := "whoever comes"
	if e != nil {
		against = "the " + e.Tok()
		in.Believed, _ = w.believe(f, e)
		in.Grudge = f.Grudge[e.ID] > 0
		in.AlliedEnemy = w.allied(f, e)
		in.EnemyNear = f.Met[e.ID] && (w.inReach(e, f.Home) || len(w.front(e, f)) > 0)
		in.AtWar = f.Wars[e.ID]
	}
	ans := mind.AnswerPact(in, w.Cfg.Tuning)
	w.explain(f, "asked by the "+c.Tok()+" for a pact of "+m.PactKind.String()+" against "+against, ans)
	if ans.Reason != "" {
		return
	}
	if !ans.Accept {
		c.Tally.Refused++
		if w.R.Float64() < 0.3 {
			against := -1
			if e != nil {
				against = e.ID
			}
			w.event(KPactRefused, c, f, -1, P{"against": against})
		}
		return
	}
	w.formPact(c, f, m.PactKind, m.Target, m.Pact)
}

// formPact makes or joins a pact.
func (w *World) formPact(c, f *Civ, kind PactKind, target int, pid int) {
	var p *Pact
	if pid >= 0 && pid < len(w.Pacts) && !w.Pacts[pid].Over && contains(w.Pacts[pid].Members, c.ID) {
		p = w.Pacts[pid]
		p.Members = append(p.Members, f.ID)
		f.Pacts = append(f.Pacts, p.ID)
	} else {
		p = &Pact{ID: len(w.Pacts), Members: []int{c.ID, f.ID}, Kind: kind, Target: target, Formed: w.Now}
		w.Pacts = append(w.Pacts, p)
		c.Pacts = append(c.Pacts, p.ID)
		f.Pacts = append(f.Pacts, p.ID)
	}
	c.Tally.Pacts++
	f.Tally.Pacts++
	if c.Species.Profile().Can(species.Trades) && f.Species.Profile().Can(species.Trades) {
		startTrade(c, f) // a pact opens the road, for two peoples that have anything the sim counts to give
	}
	w.fact(FPact, c, f, -1).with(P{"pact": kind.String(), "against": target})
	if target >= 0 && c.Wars[target] {
		w.answerCall(f, c, w.Civs[target])
	}
}

// callAllies is the attacked calling on its pacts of defence.
func (w *World) callAllies(v, a *Civ, wr *War) {
	for _, pid := range v.Pacts {
		p := w.Pacts[pid]
		if p.Over || p.Kind == Aggressive {
			continue
		}
		for _, mid := range p.Members {
			m := w.Civs[mid]
			if mid == v.ID || !m.Active() || m.Wars[a.ID] || wr.Called[mid] || !w.perceives(m, a) {
				continue // an ally cannot be told about what it cannot hold in mind
			}
			wr.Called[mid] = true
			w.send(v, m, &Message{Kind: MsgCall, Target: a.ID, Pact: pid})
			v.Tally.Called++
		}
	}
}

// joinAllies is the attacker's pacts of war coming in with it.
func (w *World) joinAllies(c, e *Civ, wr *War) {
	for _, pid := range c.Pacts {
		p := w.Pacts[pid]
		if p.Over || p.Kind == Defensive || (p.Target >= 0 && p.Target != e.ID) {
			continue
		}
		for _, mid := range p.Members {
			m := w.Civs[mid]
			if mid == c.ID || !m.Active() || m.Wars[e.ID] || len(w.front(m, e)) == 0 || !w.perceives(m, e) || !w.arms(m, e) {
				continue
			}
			if wr2 := w.declare(m, e, because("pact").By(c)); wr2 != nil {
				wr2.Pact, wr2.Principal = pid, c.ID
			}
		}
	}
}

// arms says whether an ally goes to war in its own name beside its
// principal: it has ships to send, and has not come off worst against the
// enemy so often that it will not face it again. Otherwise it answers as
// an ally with no front does. (Stage 2 gives the ally a council of its
// own; this keeps allies with nothing to fight with from filling the
// cascade with empty wars.)
func (w *World) arms(m, e *Civ) bool {
	t := w.Cfg.Tuning
	return w.standing(m) > 0 && mind.Wariness(m.Wary[e.ID], t) < t.War.WaryMax
}

// answerCall is an ally deciding whether to come: join at the front if it
// has one, send relief if that helps and home stays safe, or not come.
// An ally at war holds its own war council on it (warcouncil.go).
func (w *World) answerCall(m, v, a *Civ) {
	if !m.Active() || !v.Active() || !a.Active() || m.Wars[a.ID] || !w.allied(m, v) || !v.Wars[a.ID] {
		return
	}
	p := w.pactWith(m, v)
	if len(w.front(m, a)) > 0 && w.arms(m, a) {
		if wr := w.declare(m, a, because("pact").By(v)); wr != nil && p != nil {
			wr.Pact, wr.Principal = p.ID, v.ID
		}
		return
	}
	milA, _ := w.believe(m, a)
	believed := milA + a.warBonus() + mind.ShipLevels(w.believeShips(m, a))
	k := mind.AnswerCall(mind.CallInput{Ships: w.standing(m), Total: w.ships(m), Q: w.quality(m), Victim: battle.Strength(w.standing(v), w.quality(v)), Believed: believed, Confederate: m.posture() == mind.Confederate, Betrayed: w.betrayed(m, v), Dials: m.Dials}, w.Cfg.Tuning)
	w.explain(m, "called by the "+v.Tok()+" against the "+a.Tok(), k)
	if k.Come {
		w.launch(m, Relief, v, v.Home, k.Share)
		return
	}
	if k.Blame {
		w.betray(m, v, "absent", "absent", 0.5)
	}
}

// betray records a promise broken.
// Way is the line's key: how the chronicle tells it, where the betrayal
// is not told by another fact's line.
func (w *World) betray(by, against *Civ, shape, way string, weight float64) *Event {
	w.Betrayals = append(w.Betrayals, Betrayal{By: by.ID, Against: against.ID, Year: w.Now, Shape: shape, Weight: weight})
	f := w.told(FBetrayal, by, against, -1).with(P{"shape": shape, "way": way})
	by.Tally.Betrayals++
	against.resent(by.ID, 2*weight)
	return f
}

// faith records a promise kept at a cost.
func (w *World) faith(by, forWhom *Civ, weight float64) {
	w.Betrayals = append(w.Betrayals, Betrayal{By: by.ID, Against: forWhom.ID, Year: w.Now, Shape: "came", Weight: -weight})
}

// betrayed says whether one people has broken faith with another.
func (w *World) betrayed(v, by *Civ) bool {
	for _, b := range w.Betrayals {
		if b.By == by.ID && b.Against == v.ID && b.Weight > 0 {
			return true
		}
	}
	return false
}

// infamy is what the galaxy remembers against a people, halved every two
// hundred thousand years.
func (w *World) infamy(c *Civ) float64 {
	x := 0.0
	for _, b := range w.Betrayals {
		if b.By == c.ID && b.Weight > 0 {
			x += b.Weight * math.Pow(0.5, float64(w.Now-b.Year)/200_000)
		}
	}
	return x
}

// renown is faith kept, remembered the same way.
func (w *World) renown(c *Civ) float64 {
	x := 0.0
	for _, b := range w.Betrayals {
		if b.By == c.ID && b.Weight < 0 {
			x -= b.Weight * math.Pow(0.5, float64(w.Now-b.Year)/200_000)
		}
	}
	return x
}

// breakPacts ends every pact two peoples share.
func (w *World) breakPacts(c, h *Civ) {
	for _, pid := range c.Pacts {
		p := w.Pacts[pid]
		if p.Over || !contains(p.Members, h.ID) {
			continue
		}
		p.Members = remove(p.Members, c.ID)
		if len(p.Members) < 2 {
			p.Over, p.Ended = true, w.Now
		}
	}
	w.cutTrade(c, h, because("pact_left"))
	w.cutTrade(h, c, because("pact_left"))
	endTrade(c, h)
}

// warEnded is the pact side of a war's end: a principal's peace binds its
// allies, and an ally's own peace is a separate peace.
func (w *World) warEnded(wr *War) {
	a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
	if wr.Principal < 0 {
		for _, o := range w.Wars {
			if o.Over || o.Principal < 0 {
				continue
			}
			if (o.Principal == a.ID && (o.Sides[0] == b.ID || o.Sides[1] == b.ID)) || (o.Principal == b.ID && (o.Sides[0] == a.ID || o.Sides[1] == a.ID)) {
				w.endWar(o, "pact_peace")
			}
		}
		return
	}
	if wr.Result != "peace" && wr.Result != "capitulation" && wr.Result != "terms" {
		return
	}
	pr := w.Civs[wr.Principal]
	ally, enemy := a, b
	if w.pactWith(pr, b) != nil && w.pactWith(pr, a) == nil {
		ally, enemy = b, a
	}
	if pr.Active() && pr.Wars[enemy.ID] {
		w.betray(ally, pr, "separate_peace", "separate", 0.3).P["enemy"] = enemy.ID
	}
}

// endWars closes every war a people is in, when it falls.
func (w *World) endWars(c *Civ, result string) {
	for _, wr := range w.Wars {
		if !wr.Over && (wr.Sides[0] == c.ID || wr.Sides[1] == c.ID) {
			w.endWar(wr, result)
		}
	}
}
