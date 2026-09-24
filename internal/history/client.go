package history

import (
	"worldgen/internal/battle"
	"worldgen/internal/mind"
)

// Clients (specs/proposals/war.md, stage 3). A vassal is an actor on a
// leash: it holds its own councils, may make war on free peoples and on
// other masters' vassals, and never on its master, its master's allies or
// its master's other vassals. Struck, it calls on its patron, whose
// council joins the war, sends ships to stand with it, or stays out; the
// attacker has wronged the patron whatever it does, and staying out is a
// betrayal. Every council weighing an attack on a client reckons the
// patron with it, by the patron's record of coming. A slave stays inert.
// The judgments are mind.AnswerClient, mind.Guarantee and mind.ComeRate;
// the tribute the vassal pays for all this is tribute.go's.

// mayWar says whether two peoples may be at war under the leash: free
// peoples and vassals may, a slave may not; a vassal never with its
// master, its master's allies or its master's other vassals; a master
// with its own vassal only over a tribute unpaid; and a hunt always,
// since the hunter cannot know whose it hunts.
func (w *World) mayWar(c, e *Civ, cause string) bool {
	if cause == "ledger_hole" {
		return true
	}
	if c.slave() || e.slave() {
		return false // what a slave had to fight with is its master's
	}
	if c.Master == e.ID || e.Master == c.ID {
		return cause == "unpaid"
	}
	if !c.Free() && !e.Free() && c.Master == e.Master {
		return false
	}
	if !c.Free() && w.allied(w.Civs[c.Master], e) || !e.Free() && w.allied(w.Civs[e.Master], c) {
		return false
	}
	return true
}

// protected says whether a world of e's is no target of c's: the home of
// another's client, which is its patron's to keep. A client is struck
// for its other worlds, or not at all (stage 3's second batch: homes
// taken and spared, and the taker back after every truce, five hundred
// times).
func (w *World) protected(c, e *Civ, s int) bool {
	return s == e.Home && e.Vassal && e.Master >= 0 && e.Master != c.ID
}

// slave says whether a people is held as a slave: no council, no fleet.
func (c *Civ) slave() bool { return !c.Free() && !c.Vassal }

// sits says whether a people holds its own councils: free, or a vassal
// on its leash.
func (c *Civ) sits() bool { return c.Free() || c.Vassal }

// callPatron is a vassal struck calling on its patron, by message, as an
// ally calls on its pact.
func (w *World) callPatron(v, a *Civ) {
	if !v.Vassal || v.Master < 0 || v.Master == a.ID {
		return
	}
	m := w.Civs[v.Master]
	if !m.Active() || m.Wars[a.ID] || !w.perceives(m, a) {
		return
	}
	w.send(v, m, &Message{Kind: MsgClient, Target: a.ID, Pact: -1})
	v.Tally.Called++
}

// answerClient is the patron's council on its client's call: join the
// war, send ships to stand with the client, or stay out. The attacker has
// wronged it whatever it does; staying out is a betrayal of the client,
// and the record of coming is what every other council reads of the
// patron's guarantee.
func (w *World) answerClient(m, v, a *Civ) {
	if !m.Active() || !v.Active() || !a.Active() || v.Master != m.ID || !v.Vassal || !v.Wars[a.ID] || m.Wars[a.ID] || !m.sits() {
		return
	}
	t := w.Cfg.Tuning
	m.ClientCalls++
	m.resent(a.ID, t.Client.Offence) // an attack on a client is an attack on its patron's honour
	ap := w.appraise(m, a, -1)
	milA, _ := w.believe(m, a)
	k := mind.AnswerCall(mind.CallInput{Ships: w.standing(m), Total: w.ships(m), Q: w.quality(m), Victim: battle.Strength(w.standing(v), w.quality(v)), Believed: milA + a.warBonus() + mind.ShipLevels(w.believeShips(m, a)), Dials: m.Dials}, t)
	spare := 0
	if k.Safe && k.Share <= w.standing(m) {
		spare = k.Share
	}
	ans := mind.AnswerClient(mind.ClientInput{Posture: m.posture(), Acted: ap.Acted, Worth: w.clientWorth(m, v), Dials: m.Dials, Spare: spare, Reach: m.launches() && (len(ap.Front) > 0 || w.inReach(m, a.Home))}, t)
	w.explain(m, "called by its client the "+v.Tok()+" against the "+a.Tok(), ans)
	switch ans.Choice {
	case mind.Join:
		if wr := w.declare(m, a, because("client").By(v)); wr != nil {
			wr.Principal = v.ID // its war is the client's: it ends with the client's peace
			m.ClientCame++
			return
		}
		fallthrough
	case mind.Back:
		if spare > 0 && w.launch(m, Relief, v, v.Home, spare) != nil {
			m.ClientCame++
			return
		}
	}
	w.betray(m, v, "abandoned", "abandoned", t.Client.Abandoned)
	v.Seen = -1 // a client abandoned watches its patron as a slave watches a master in decline: the question of revolt is put (revolt)
}

// guarantee is what c reckons of the patron standing behind e, when e is
// a client whose patron is not c: a share of the fleet it believes the
// patron has, by the patron's record of coming, as ships that would
// stand over the client's world.
func (w *World) guarantee(c, e *Civ) float64 {
	if !e.Vassal || e.Master < 0 || e.Master == c.ID {
		return 0
	}
	p := w.Civs[e.Master]
	if !p.Active() || !w.perceives(c, p) {
		return 0
	}
	return mind.Guarantee(w.believeShips(c, p), mind.ComeRate(p.ClientCame, p.ClientCalls, w.Cfg.Tuning), w.Cfg.Tuning)
}

// clientWorth is what a client is worth to its patron, 0 to 1: the
// tribute it pays at its rate, and the faith it has kept.
func (w *World) clientWorth(m, v *Civ) float64 {
	x := 0.3 + 2*v.Rate
	if v.PaidShort > 0 {
		x -= 0.1 * float64(min(v.PaidShort, 5))
	}
	return min(1, max(0, x))
}
