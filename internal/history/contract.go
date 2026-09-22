package history

import (
	"math"
	"sort"
	"worldgen/internal/species"

	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/tech"
)

// Contracts: a timed payment for a discrete thing. Two peoples each give
// one term for a length: a flow, a rarity, access to one, a node taught,
// a fleet standing guard, a fleet sent against a world, a world taken and
// handed over, peace, a sighting, or a people spoken for. The buyer asks
// and pays; the seller does. An offer and its answer are messages at the
// lag of light; the contract forms when the answer arrives. A flow owed
// is a use in the direction order, in the word category, placed by
// honour. A term undelivered two ticks running breaks the contract, and
// the one that failed is the breaker, a betrayal on the world. A hired
// fleet may be bought off by the one it stands against. A war that ends
// by yielding may end in tribute instead of worlds. Each term kind is one
// entry in termKinds: what it is worth to the one that gets it and to
// the one that gives it, and what it does when the contract forms, each
// tick, and when it is done or lapses. The decisions are mind/contract.go.

// TermKind is what one side gives; the kinds are the mind's.
type TermKind = mind.TermKind

// Term is one side's part.
type Term struct {
	Kind   TermKind
	Res    flow.Kind // flow: the kind
	Amount float64   // flow: units per tick; guard, strike, deliver: ships
	Source int       // rarity, access: the source
	Node   string    // teach
	Star   int       // guard: where to stand; strike, deliver: the target world
	Work   string    // strike: a structure key at Star to burn, or "" for the world
	Target int       // guard: whom against; strike, deliver: whose; peace: with whom; broker: the people to be understood
	Fleet  int       // sighting: the fleet
}

// ContractState is where a contract is in its life.
type ContractState uint8

const (
	Offered  ContractState = iota
	Accepted               // the answer is on its way back
	Running
	Done
	Broken
	Lapsed
	Refused
)

func (s ContractState) String() string {
	return [...]string{"offered", "accepted", "running", "done", "broken", "lapsed", "refused"}[s]
}

// Contract binds two peoples.
type Contract struct {
	ID            int
	Buyer, Seller int // the buyer asks, the seller does; the buyer pays
	By            int // who proposed it
	Ask, Pay      Term
	Length        float64 // thousand years the pay runs
	Offered       Year
	Formed, Until Year // Until is 0 until the clock starts: a guard's starts when its fleet arrives
	Ended         Year
	State         ContractState
	Broke         int    // who broke it, or -1
	Failed        [2]int // ticks running each term has gone undelivered: the ask, the pay
	Missed        bool   // the pay has failed at least once
	AskDone       bool   // the ask is delivered; the pay may run on to Until
	Taught        bool   // teach: the node arrived
	Burned        bool   // strike: the work was burned
	Tribute       bool   // written at a war's end, not offered
	BoughtOff     bool   // ended by the seller taking a better offer from the other side
}

// terms are the ask, given by the seller, and the pay, given by the buyer.
func (k *Contract) terms() [2]Term { return [2]Term{k.Ask, k.Pay} }

// giver is who gives term i, and receiver who gets it.
func (k *Contract) giver(i int) (giver, receiver int) {
	if i == 0 {
		return k.Seller, k.Buyer
	}
	return k.Buyer, k.Seller
}

// wordKey is a flow term as a use: what the shed set names it by.
func wordKey(id int) string { return "word:" + itoa(id) }

// termDef is one term kind's rules. worthTo prices the term for the one
// that gets it, worthFrom for the one that gives it; start runs when the
// contract forms; tick says whether the term was delivered this tick;
// done says whether a one-off term is finished; lapse names why the term
// cannot go on, or "".
type termDef struct {
	worthTo   func(w *World, c *Civ, k *Contract, t Term, other *Civ) float64
	worthFrom func(w *World, c *Civ, k *Contract, t Term, other *Civ) float64
	start     func(w *World, k *Contract, giver, receiver *Civ, t Term)
	tick      func(w *World, k *Contract, giver, receiver *Civ, t Term) bool
	done      func(w *World, k *Contract, giver, receiver *Civ, t Term) bool
	lapse     func(w *World, k *Contract, giver, receiver *Civ, t Term) string
}

var termKinds map[TermKind]termDef

func init() {
	termKinds = map[TermKind]termDef{
		mind.TermFlow:     {worthTo: flowWorthTo, worthFrom: flowWorthFrom, tick: flowTick},
		mind.TermRarity:   {worthTo: rarityWorthTo, worthFrom: rarityWorthFrom, start: rarityStart, done: always},
		mind.TermAccess:   {worthTo: rarityWorthTo, worthFrom: rarityWorthFrom, tick: accessTick},
		mind.TermTeach:    {worthTo: teachWorthTo, worthFrom: teachWorthFrom, start: teachStart, tick: always, done: teachDone},
		mind.TermGuard:    {worthTo: guardWorthTo, worthFrom: guardWorthFrom, start: guardStart, tick: fleetTick, lapse: guardLapse},
		mind.TermStrike:   {worthTo: strikeWorthTo, worthFrom: strikeWorthFrom, start: strikeStart, tick: fleetTick, done: strikeDone},
		mind.TermDeliver:  {worthTo: strikeWorthTo, worthFrom: strikeWorthFrom, start: strikeStart, tick: fleetTick, done: deliverDone},
		mind.TermPeace:    {worthTo: peaceWorthTo, worthFrom: peaceWorthFrom, start: peaceStart, tick: peaceTick},
		mind.TermSighting: {worthTo: sightingWorthTo, worthFrom: nothing, start: sightingStart, done: always},
		mind.TermBroker:   {worthTo: brokerWorthTo, worthFrom: brokerWorthFrom, tick: brokerTick, done: brokerDone, lapse: brokerLapse},
	}
}

func always(*World, *Contract, *Civ, *Civ, Term) bool     { return true }
func nothing(*World, *Civ, *Contract, Term, *Civ) float64 { return 0 }
func (k *Contract) ticks(w *World) float64                { return k.Length }
func (w *World) spare(c *Civ) flow.Income                 { return c.Surplus.Less(c.Reserved) }
func (w *World) contractOf(x *Expedition) *Contract {
	if x.Contract < 0 {
		return nil
	}
	return w.Contracts[x.Contract]
}

// The worth of each term, to the one that gets it and to the one that
// gives it. A lasting term counts for the contract's length.

func flowWorthTo(w *World, c *Civ, k *Contract, t Term, _ *Civ) float64 {
	p := &w.Cfg.Tuning.Contract
	refill := 0.0
	for _, u := range w.uses(c) {
		if c.Shed[u.Key] && u.Need[t.Res] > 0 {
			refill += u.Need[t.Res]
		}
	}
	refill = min(refill, t.Amount)
	return (refill + p.FlowRest*(t.Amount-refill)) * k.ticks(w)
}

func flowWorthFrom(w *World, c *Civ, k *Contract, t Term, _ *Civ) float64 {
	return max(0, t.Amount-max(0, w.spare(c)[t.Res])) * k.ticks(w)
}

// rarityValue is what a rarity counts for to a people: the nodes it opens
// that the people lacks, at the prize's rate, plus its levels and reach.
func rarityValue(c *Civ, s *Source) float64 {
	v := 0.0
	for _, g := range s.Grants {
		if !c.Known[g] {
			v += rarityWorth
		}
	}
	return v + s.Levels[0] + s.Levels[1] + s.Levels[2] + s.Reach/5
}

func rarityWorthTo(w *World, c *Civ, _ *Contract, t Term, _ *Civ) float64 {
	return rarityValue(c, w.Sources[t.Source])
}

func rarityWorthFrom(w *World, c *Civ, _ *Contract, t Term, _ *Civ) float64 {
	v := rarityValue(c, w.Sources[t.Source])
	if c.fixed(OldThings) {
		v *= 3
	}
	return v
}

func rarityStart(w *World, k *Contract, giver, receiver *Civ, t Term) {
	s := w.Sources[t.Source]
	if s.Holder != giver.ID || s.Carried >= 0 {
		w.lapse(k, "what was promised was no longer theirs to give")
		return
	}
	w.transfer(s, giver, receiver)
	s.Star = receiver.Home
	w.event(KRarityPassed, giver, receiver, -1, P{"source": s.ID})
}

func accessTick(w *World, _ *Contract, giver, _ *Civ, t Term) bool {
	s := w.Sources[t.Source]
	return w.Owner[s.Star] == giver.ID
}

// saves is the ticks of research a node would save a people: its price
// less what is banked toward it, over its rate.
func (w *World) saves(c *Civ, key string) float64 {
	n := tech.Get(key)
	if n == nil || c.Known[key] {
		return 0
	}
	price := w.price(c, n)
	if c.Pursuit == key {
		price -= c.Progress
	}
	return max(0, price) / max(w.researchRate(c), 1e-6)
}

func teachWorthTo(w *World, c *Civ, _ *Contract, t Term, _ *Civ) float64 {
	return w.saves(c, t.Node)
}

func teachWorthFrom(w *World, c *Civ, _ *Contract, t Term, other *Civ) float64 {
	p := &w.Cfg.Tuning.Contract
	v := p.TeachGiver * w.saves(other, t.Node)
	if other.hostile() || w.threat(c) == other {
		v *= p.TeachHostile
	}
	return v
}

func teachStart(w *World, k *Contract, giver, receiver *Civ, t Term) {
	w.send(giver, receiver, &Message{Kind: MsgTeach, Contract: k.ID, Node: t.Node})
}

func teachDone(_ *World, k *Contract, _, _ *Civ, _ Term) bool { return k.Taught }

func guardWorthTo(w *World, c *Civ, k *Contract, t Term, _ *Civ) float64 {
	gap := 0.5
	if t.Target >= 0 {
		mil, _ := w.believe(c, w.Civs[t.Target])
		gap = max(gap, mil-c.Mil)
	}
	return t.Amount * gap * k.ticks(w)
}

func guardWorthFrom(w *World, c *Civ, k *Contract, t Term, _ *Civ) float64 {
	p := &w.Cfg.Tuning.Contract
	risk := p.RiskElse
	if t.Target >= 0 && contains(w.front(w.Civs[t.Target], c), t.Star) {
		risk = p.RiskFront
	}
	return t.Amount * p.GuardGiver * risk * k.ticks(w)
}

func guardStart(w *World, k *Contract, giver, receiver *Civ, t Term) {
	if x := w.launch(giver, Relief, receiver, t.Star, int(t.Amount)); x != nil {
		x.Contract = k.ID
	}
}

// fleetTick: a fleet term is delivered while the fleet is at its star or
// on its way there.
func fleetTick(w *World, k *Contract, _, _ *Civ, _ Term) bool {
	x := w.contractFleet(k)
	return x != nil && !x.Over && !x.Returning && x.Ships > 0
}

func guardLapse(w *World, k *Contract, _, receiver *Civ, t Term) string {
	if w.Owner[t.Star] != receiver.ID {
		return "the star it was to hold changed hands"
	}
	if x := w.contractFleet(k); x != nil && x.Over && x.Ships <= 0 {
		return "the fleet was lost"
	}
	return ""
}

func strikeWorthTo(w *World, c *Civ, _ *Contract, t Term, _ *Civ) float64 {
	p := &w.Cfg.Tuning.Contract
	v := w.prize(c, t.Star) * p.PrizeYears
	if t.Target >= 0 {
		ships, guns, relief := w.believeSky(c, w.Civs[t.Target], t.Star)
		v += ships + guns + relief
	}
	return v
}

func strikeWorthFrom(w *World, c *Civ, _ *Contract, t Term, _ *Civ) float64 {
	p := &w.Cfg.Tuning.Contract
	return t.Amount*p.RiskFront + p.StrikeGrudge
}

func strikeStart(w *World, k *Contract, giver, receiver *Civ, t Term) {
	e := w.Civs[t.Target]
	if !e.Active() {
		return
	}
	if w.warBetween(giver.ID, e.ID) == nil {
		if wr := w.declare(giver, e, because("coin").By(receiver)); wr != nil {
			wr.Hire = k.ID
		}
	}
	e.resent(receiver.ID, 1)
	e.resent(giver.ID, 0.5)
	if x := w.launch(giver, Campaign, e, t.Star, int(t.Amount)); x != nil {
		x.Contract = k.ID
	}
}

func strikeDone(w *World, k *Contract, _, _ *Civ, t Term) bool {
	if t.Work != "" {
		return k.Burned
	}
	x := w.contractFleet(k)
	return w.Owner[t.Star] != t.Target || (x != nil && x.Over)
}

func deliverDone(w *World, k *Contract, giver, receiver *Civ, t Term) bool {
	if w.Owner[t.Star] == giver.ID {
		w.handOver(giver, receiver, t.Star)
	}
	x := w.contractFleet(k)
	return w.Owner[t.Star] == receiver.ID || (x != nil && x.Over)
}

func peaceWorthTo(w *World, c *Civ, _ *Contract, t Term, other *Civ) float64 {
	if wr := w.warBetween(c.ID, other.ID); wr != nil {
		return wr.Will[wr.side(c.ID)] * w.Cfg.Tuning.Contract.WillWorth
	}
	return 0
}

func peaceWorthFrom(w *World, c *Civ, _ *Contract, t Term, other *Civ) float64 {
	if f := w.front(c, other); len(f) > 0 {
		return w.prize(c, f[0]) * w.Cfg.Tuning.Contract.PrizeYears
	}
	return 0
}

// peaceStart is the peace as a truce on both: no new war before the end.
func peaceStart(w *World, k *Contract, giver, receiver *Civ, _ Term) {
	giver.Truce[receiver.ID] = max(giver.Truce[receiver.ID], k.Until)
	receiver.Truce[giver.ID] = max(receiver.Truce[giver.ID], k.Until)
}

func peaceTick(w *World, k *Contract, giver, receiver *Civ, _ Term) bool {
	wr := w.warBetween(giver.ID, receiver.ID)
	return wr == nil || wr.Began < k.Formed
}

func sightingWorthTo(w *World, c *Civ, _ *Contract, t Term, other *Civ) float64 {
	x := w.Expeditions[t.Fleet]
	s := other.Sightings[t.Fleet]
	if s == nil || x.Over || x.Base >= 0 || x.Launched != s.Leg || x.Arrive <= w.Now {
		return 0
	}
	left := float64(x.Arrive-w.Now) / max(1, float64(x.Arrive-x.Launched))
	return s.Mil * w.Cfg.Tuning.Contract.SightingWorth * left
}

func sightingStart(w *World, k *Contract, giver, receiver *Civ, t Term) {
	x := w.Expeditions[t.Fleet]
	s := giver.Sightings[t.Fleet]
	if s == nil || x.Over || x.Base >= 0 || x.Launched != s.Leg {
		w.lapse(k, "the fleet had already arrived")
		return
	}
	x.SoldBy = giver.ID
	giver.Tally.SoldSightings++
	w.send(giver, receiver, &Message{Kind: MsgSighting, About: x.Owner, Target: x.ID, Sighting: s})
	w.event(KSightingSold, giver, receiver, -1, P{"owner": x.Owner, "years": Year(k.Length * 1000), "res": k.Pay.Res})
}

func brokerWorthTo(w *World, c *Civ, _ *Contract, t Term, _ *Civ) float64 {
	p := &w.Cfg.Tuning.Contract
	e := w.Civs[t.Target]
	v := p.BrokerBase + min(e.Surplus.Total(), c.Want.Total())*p.PrizeYears
	if c.Wars[e.ID] {
		v += p.BrokerBase
	}
	return v
}

func brokerWorthFrom(w *World, c *Civ, _ *Contract, t Term, other *Civ) float64 {
	p := &w.Cfg.Tuning.Contract
	return p.BrokerGiver*(c.From[other.ID].Total()+c.From[t.Target].Total()) + c.Dials.Greed
}

func brokerTick(w *World, k *Contract, giver, receiver *Civ, t Term) bool {
	if w.chance(w.Cfg.Tuning.Contract.BrokerRate) {
		w.broker(giver, receiver, w.Civs[t.Target])
	}
	return true
}

func brokerDone(_ *World, _ *Contract, _, receiver *Civ, t Term) bool {
	return receiver.Fathomed[t.Target]
}

func brokerLapse(w *World, k *Contract, giver, receiver *Civ, t Term) string {
	switch {
	case float64(w.Now-k.Formed) >= w.Cfg.Tuning.Contract.BrokerLapse:
		return "the words never got through"
	case !giver.Fathomed[receiver.ID] || !giver.Fathomed[t.Target]:
		return "the broker lost the thread"
	}
	return ""
}

// worthTo and worthFrom price a term for a people through the registry.
func (w *World) worthTo(c *Civ, k *Contract, t Term, other *Civ) float64 {
	return termKinds[t.Kind].worthTo(w, c, k, t, other)
}

func (w *World) worthFrom(c *Civ, k *Contract, t Term, other *Civ) float64 {
	return termKinds[t.Kind].worthFrom(w, c, k, t, other)
}

// refuses is what a people never gives, whatever the price: a rarity when
// fixed on old things; a weapons node to a people it reads as hostile; a
// fleet while at war or to be used on an ally; a broker's word when it is
// a xenophobe or at war with the one to be understood.
func (w *World) refuses(c *Civ, t Term, other *Civ) string {
	switch t.Kind {
	case mind.TermRarity:
		if c.fixed(OldThings) {
			return "they part with nothing old"
		}
	case mind.TermTeach:
		if n := tech.Get(t.Node); n != nil && n.Domain == tech.Weapons && (other.hostile() || w.threat(c) == other) {
			return "they teach no weapon to a people that strikes first"
		}
	case mind.TermGuard, mind.TermStrike, mind.TermDeliver:
		if len(c.Wars) > 0 {
			return "at war, they sell no fleet"
		}
		if t.Target >= 0 && (w.allied(c, w.Civs[t.Target]) || t.Target == c.ID) {
			return "not against their own"
		}
	case mind.TermBroker:
		if b := mind.Broker(mind.BrokerInput{SharedPact: true, Posture: c.posture(), Xenophobic: c.Has("xenophobic"), AtWarWith: c.Wars[t.Target]}, w.Cfg.Tuning); b.Rate == 0 {
			return b.Reason
		}
	}
	return ""
}

// crimeToGive says whether a people's morality counts giving a term a
// crime: a fleet sent against a world, by the war row.
func crimeToGive(c *Civ, t Term) bool {
	switch t.Kind {
	case mind.TermStrike, mind.TermDeliver:
		return moralTable[FWar][c.Morality.column()].Sort == Crime
	}
	return false
}

// contractFleet is the fleet doing a contract's guard or strike, or nil.
func (w *World) contractFleet(k *Contract) *Expedition {
	for _, x := range w.Expeditions {
		if x.Contract == k.ID {
			return x
		}
	}
	return nil
}

// Offers.

// contracting is the civ step: sightings for sale, the sellsword's offer,
// and the buyer's ask.
func (w *World) contracting(c *Civ) {
	if !c.Active() || !c.Free() || !c.Species.Profile().Can(species.Trades) {
		return // nothing the sim counts to bargain with
	}
	w.sellSightings(c)
	if standing := w.standing(c); standing > 0 && c.Upkeep.Total() > 0 {
		wantShare := c.Want.Total() / c.Upkeep.Total()
		spareShare := float64(w.idleShips(c)) / float64(standing)
		if w.chance(mind.MercenaryRate(wantShare, spareShare, w.Cfg.Tuning)) {
			if w.Cfg.TraceAI {
				w.event(KDebug, c, nil, -1, P{"text": sprintf("[the %s, in want with %s idle, look for someone to hold a gate for pay]", c.Tok(), shipsWord(w.idleShips(c)))})
			}
			w.offerGuard(c)
		}
	}
	if w.chance(mind.AskRate(c.Dials.Greed, c.fixed(Holding), w.Cfg.Tuning)) {
		w.ask(c)
	}
}

// idleShips is what a people could send away: its ships at bases beyond
// what its garrisons asked for, none while at war.
func (w *World) idleShips(c *Civ) int {
	if len(c.Wars) > 0 {
		return 0
	}
	return max(0, w.standing(c)-c.GarrisonWant)
}

// dealable says whether two peoples can deal: both free, understanding
// each other, at peace, neither a monster to the other, no deep grudge,
// and the asker's last offer far enough back.
func (w *World) dealable(c, e *Civ) bool {
	return e.Active() && e.Free() && w.mutual(c, e) && !c.Wars[e.ID] && !w.monster(c, e) && !w.monster(e, c) &&
		e.Grudge[c.ID] <= w.Cfg.Tuning.Trade.GrudgeBar && c.Asked[e.ID]+Year(w.Cfg.Tuning.Pact.AskAgain) <= w.Now
}

// ask is a buyer looking at its wants in order and asking for the first
// it finds a seller for: a guard, a strike, a node, a rarity, access, a
// flow, a broker.
func (w *World) ask(c *Civ) {
	for _, want := range []func(*Civ) (*Civ, Term){w.wantGuard, w.wantStrike, w.wantNode, w.wantRarity, w.wantAccess, w.wantFlow, w.wantBroker} {
		if seller, t := want(c); seller != nil {
			w.propose(c, seller, t)
			return
		}
	}
}

// fleetSeller is the met people with the largest idle fleet that could
// send so many ships against a star.
func (w *World) fleetSeller(c *Civ, star, ships int, against *Civ) *Civ {
	var best *Civ
	bestIdle := 0
	for _, eid := range sortedInts(c.Met) {
		e := w.Civs[eid]
		if e == against || !w.dealable(c, e) || w.allied(e, against) {
			continue
		}
		idle := w.idleShips(e)
		if idle < ships || idle <= bestIdle || w.guardWith(e, star, ships) == nil {
			continue
		}
		if _, d := w.nearest(e, star); d*e.Speed > w.Cfg.Tuning.Campaign.MaxLag {
			continue
		}
		best, bestIdle = e, idle
	}
	return best
}

// shipsFor is how many ships close a gap in levels: one a level, at least one.
func shipsFor(gap float64) int { return max(1, int(math.Ceil(gap))) }

func (w *World) wantGuard(c *Civ) (*Civ, Term) {
	th := w.threat(c)
	if th == nil {
		return nil, Term{}
	}
	mil, _ := w.believe(c, th)
	if mil-c.Mil < 0.5 {
		return nil, Term{}
	}
	star := c.Home
	if f := w.front(th, c); len(f) > 0 {
		star = f[0]
	}
	n := shipsFor(mil - c.Mil)
	if s := w.fleetSeller(c, star, n, th); s != nil {
		return s, Term{Kind: mind.TermGuard, Star: star, Amount: float64(n), Target: th.ID}
	}
	return nil, Term{}
}

func (w *World) wantStrike(c *Civ) (*Civ, Term) {
	if len(c.Wars) == 0 || c.WantShips == 0 {
		return nil, Term{}
	}
	for _, eid := range sortedInts(c.Wars) {
		e := w.Civs[eid]
		if !e.Active() || e.Aloft {
			continue
		}
		_, star := w.nearestEnemy(c, e)
		ships, guns, relief := w.believeSky(c, e, star)
		n := max(1, int(math.Ceil(ships+guns+relief)))
		work := ""
		for _, wk := range e.Works {
			if wk.Star == star && tech.Structures[wk.Key] != nil && tech.Structures[wk.Key].Guns > 0 {
				work = wk.Key
				break
			}
		}
		if s := w.fleetSeller(c, star, n, e); s != nil {
			return s, Term{Kind: mind.TermStrike, Star: star, Amount: float64(n), Target: eid, Work: work}
		}
	}
	return nil, Term{}
}

// teacher is the met people that knows a node and holds the smallest
// grudge against the asker.
func (w *World) teacher(c *Civ, key string) *Civ {
	var best *Civ
	for _, eid := range sortedInts(c.Met) {
		e := w.Civs[eid]
		if !e.Known[key] || !w.dealable(c, e) {
			continue
		}
		if best == nil || e.Grudge[c.ID] < best.Grudge[c.ID] {
			best = e
		}
	}
	return best
}

func (w *World) wantNode(c *Civ) (*Civ, Term) {
	if c.Pursuit != "" {
		if s := w.teacher(c, c.Pursuit); s != nil {
			return s, Term{Kind: mind.TermTeach, Node: c.Pursuit}
		}
	}
	var best *tech.Node
	var seller *Civ
	for _, n := range tech.Nodes {
		if n.Miracle || (best != nil && n.Era <= best.Era) || !w.canPursue(c, n) {
			continue
		}
		if s := w.teacher(c, n.Key); s != nil {
			best, seller = n, s
		}
	}
	if best == nil {
		return nil, Term{}
	}
	return seller, Term{Kind: mind.TermTeach, Node: best.Key}
}

// gated is the rarity that would halve a people's pursuit, held by a met
// people that has no node left to gain from it: mobile or not.
func (w *World) gated(c *Civ, mobile bool) (*Civ, *Source) {
	if c.Pursuit == "" || c.Grants[c.Pursuit] {
		return nil, nil
	}
	for _, s := range w.Sources {
		if !s.Rarity || s.Mobile != mobile || s.Holder < 0 || s.Holder == c.ID || !grants(s, c.Pursuit) {
			continue
		}
		if !mobile && w.Owner[s.Star] != s.Holder {
			continue
		}
		e := w.Civs[s.Holder]
		if !c.Met[e.ID] || !w.dealable(c, e) {
			continue
		}
		spare := true
		for _, g := range s.Grants {
			if !e.Known[g] {
				spare = false
			}
		}
		if spare {
			return e, s
		}
	}
	return nil, nil
}

// grants says whether a rarity halves a node.
func grants(s *Source, key string) bool {
	for _, g := range s.Grants {
		if g == key {
			return true
		}
	}
	return false
}

func (w *World) wantRarity(c *Civ) (*Civ, Term) {
	if e, s := w.gated(c, true); e != nil {
		return e, Term{Kind: mind.TermRarity, Source: s.ID}
	}
	return nil, Term{}
}

func (w *World) wantAccess(c *Civ) (*Civ, Term) {
	if e, s := w.gated(c, false); e != nil {
		return e, Term{Kind: mind.TermAccess, Source: s.ID}
	}
	return nil, Term{}
}

func (w *World) wantFlow(c *Civ) (*Civ, Term) {
	if len(c.Shed) == 0 {
		return nil, Term{}
	}
	for _, k := range flow.Kinds {
		short := c.OwnWant[k] - c.Received[k]
		if short <= 0 {
			continue
		}
		var best *Civ
		bestSpare := 0.0
		for _, eid := range sortedInts(c.Met) {
			e := w.Civs[eid]
			if !w.dealable(c, e) {
				continue
			}
			if sp := w.spare(e)[k]; sp > bestSpare {
				best, bestSpare = e, sp
			}
		}
		if best != nil {
			return best, Term{Kind: mind.TermFlow, Res: k, Amount: min(short, w.transportCap(best)*bestSpare)}
		}
	}
	return nil, Term{}
}

func (w *World) wantBroker(c *Civ) (*Civ, Term) {
	for _, eid := range sortedInts(c.FathomTried) {
		e := w.Civs[eid]
		if c.Fathomed[eid] || !e.Active() {
			continue
		}
		for _, zid := range sortedInts(c.Fathomed) {
			z := w.Civs[zid]
			if zid == eid || !z.Fathomed[c.ID] || !z.Fathomed[eid] || !w.dealable(c, z) {
				continue
			}
			if w.refuses(z, Term{Kind: mind.TermBroker, Target: eid}, c) != "" {
				continue
			}
			return z, Term{Kind: mind.TermBroker, Target: eid}
		}
	}
	return nil, Term{}
}

// transportCap is the share of a spare that can cross by a people's road.
func (w *World) transportCap(c *Civ) float64 {
	p := &w.Cfg.Tuning.Trade
	switch w.drive(c) {
	case 2:
		return p.CapDoor
	case 1:
		return p.CapFast
	}
	return p.CapBase
}

// payFor is what a buyer offers a seller for a thing of so many units:
// the kind the seller wants most that the buyer has spare, up to the
// spare and the road; failing that a mobile rarity it has no node for,
// access to an immobile one, or a node the seller could pursue, the
// highest era first. A buyer fixed on old things pays with no rarity;
// one fixed on holding with no metal.
func (w *World) payFor(buyer, seller *Civ) (Term, bool) {
	kinds := []flow.Kind{flow.O, flow.E, flow.M}
	sort.SliceStable(kinds, func(i, j int) bool { return seller.OwnWant[kinds[i]] > seller.OwnWant[kinds[j]] })
	spare := w.spare(buyer)
	for _, k := range kinds {
		if k == flow.M && buyer.fixed(Holding) {
			continue
		}
		if spare[k] <= 0.1 {
			continue
		}
		amount := spare[k] * w.transportCap(buyer)
		if seller.OwnWant[k] > 0 {
			amount = min(amount, seller.OwnWant[k])
		}
		if amount > 0.05 {
			return Term{Kind: mind.TermFlow, Res: k, Amount: amount}, true
		}
	}
	if !buyer.fixed(OldThings) {
		for _, id := range w.mobile {
			s := w.Sources[id]
			if s.Holder != buyer.ID || s.Carried >= 0 || rarityValue(buyer, s) > 0 || rarityValue(seller, s) == 0 {
				continue
			}
			return Term{Kind: mind.TermRarity, Source: s.ID}, true
		}
		for _, h := range w.holdings(buyer) {
			for _, id := range w.sourcesAt[h] {
				s := w.Sources[id]
				if s.Rarity && !s.Mobile && s.Star == h && rarityValue(seller, s) > 0 {
					return Term{Kind: mind.TermAccess, Source: s.ID}, true
				}
			}
		}
	}
	var best *tech.Node
	for _, n := range tech.Nodes {
		if n.Miracle || !buyer.Known[n.Key] || (best != nil && n.Era <= best.Era) || !w.canPursue(seller, n) {
			continue
		}
		best = n
	}
	if best != nil {
		return Term{Kind: mind.TermTeach, Node: best.Key}, true
	}
	return Term{}, false
}

// units is what a term counts for in the pay's length: ships for a fleet,
// a tenth of the ticks a node saves the buyer, a fifth of a rarity's
// value, one for anything else.
func (w *World) units(buyer *Civ, t Term) float64 {
	switch t.Kind {
	case mind.TermGuard, mind.TermStrike, mind.TermDeliver:
		return t.Amount
	case mind.TermTeach:
		return w.saves(buyer, t.Node) / 10
	case mind.TermRarity, mind.TermAccess:
		return rarityValue(buyer, w.Sources[t.Source]) / 5
	}
	return 1
}

// propose is a buyer asking a seller for a term, paying what it can.
func (w *World) propose(buyer, seller *Civ, ask Term) *Contract {
	pay, ok := w.payFor(buyer, seller)
	if !ok {
		return nil
	}
	buyer.Asked[seller.ID] = w.Now
	k := w.newContract(buyer, seller, ask, pay, buyer)
	if w.worthTo(buyer, k, ask, seller) <= 0 {
		k.State = Lapsed
		return nil
	}
	w.send(buyer, seller, &Message{Kind: MsgOffer, Contract: k.ID})
	return k
}

// newContract writes a contract as offered.
func (w *World) newContract(buyer, seller *Civ, ask, pay Term, by *Civ) *Contract {
	k := &Contract{ID: len(w.Contracts), Buyer: buyer.ID, Seller: seller.ID, By: by.ID, Ask: ask, Pay: pay, Offered: w.Now, Broke: -1,
		Length: mind.PayLength(w.units(buyer, ask), w.Cfg.Tuning)}
	w.Contracts = append(w.Contracts, k)
	buyer.Contracts = append(buyer.Contracts, k.ID)
	seller.Contracts = append(seller.Contracts, k.ID)
	return k
}

// offerGuard is the sellsword's side: a people with a want and an idle
// fleet offers a guard to the most threatened met people that has a
// surplus of a kind it wants, at its own margin.
func (w *World) offerGuard(c *Civ) {
	var best, against *Civ
	bestGap, bestStar := 0.0, -1
	for _, eid := range sortedInts(c.Met) {
		e := w.Civs[eid]
		if !w.dealable(c, e) || len(e.Wars) > 0 {
			continue
		}
		th := w.threat(e)
		if th == nil || th == c || w.allied(c, th) {
			continue
		}
		mil, _ := w.believe(e, th)
		gap := mil - e.Mil
		if gap < 0.5 || gap <= bestGap {
			continue
		}
		if !w.hasSpareFor(e, c) {
			continue
		}
		star := e.Home
		if f := w.front(th, e); len(f) > 0 {
			star = f[0]
		}
		if _, d := w.nearest(c, star); d*c.Speed > w.Cfg.Tuning.Campaign.MaxLag {
			continue
		}
		best, against, bestGap, bestStar = e, th, gap, star
	}
	if best == nil {
		return
	}
	n := min(shipsFor(bestGap), w.idleShips(c))
	if n <= 0 || w.guardWith(c, bestStar, n) == nil {
		return
	}
	ask := Term{Kind: mind.TermGuard, Star: bestStar, Amount: float64(n), Target: against.ID}
	pay, ok := w.payFor(best, c)
	if !ok {
		if w.Cfg.TraceAI {
			w.event(KDebug, c, best, bestStar, P{"text": sprintf("[the %s would hold %s for the %s, who have nothing they want]", c.Tok(), w.star(bestStar), best.Tok())})
		}
		return
	}
	c.Asked[best.ID] = w.Now
	k := w.newContract(best, c, ask, pay, c)
	w.send(c, best, &Message{Kind: MsgOffer, Contract: k.ID})
}

// hasSpareFor says whether a buyer has a surplus of a kind the seller wants.
func (w *World) hasSpareFor(buyer, seller *Civ) bool {
	spare := w.spare(buyer)
	for _, k := range flow.Kinds {
		if seller.OwnWant[k] > 0 && spare[k] > 0.1 {
			return true
		}
	}
	return false
}

// sellSightings offers what a people has seen of fleets in flight, by
// honour, to the people the fleet is coming for or to a partner at war
// with its owner.
func (w *World) sellSightings(c *Civ) {
	if len(c.Sightings) == 0 {
		return
	}
	for _, fid := range sortedInts(c.Sightings) {
		s := c.Sightings[fid]
		x := w.Expeditions[fid]
		if s.Offered || x.Over || x.Base >= 0 || x.Launched != s.Leg || x.Arrive <= w.Now || x.Owner == c.ID {
			continue
		}
		s.Offered = true
		owner := w.Civs[x.Owner]
		if !owner.Active() {
			continue
		}
		boundToAlly := x.Kind == Relief && x.Target >= 0 && w.allied(c, w.Civs[x.Target])
		if !mind.SellSighting(c.honour(), w.allied(c, owner), boundToAlly) {
			continue
		}
		buyer := w.sightingBuyer(c, x)
		if buyer == nil {
			continue
		}
		ask := Term{Kind: mind.TermSighting, Fleet: fid}
		pay, ok := w.payFor(buyer, c)
		if !ok {
			continue
		}
		c.Asked[buyer.ID] = w.Now
		k := w.newContract(buyer, c, ask, pay, c)
		w.send(c, buyer, &Message{Kind: MsgOffer, Contract: k.ID})
	}
}

// bound is the people a fleet is coming for: its target, or for a relief
// fleet the enemy its host is fighting; -1 if none.
func (w *World) bound(x *Expedition) int {
	switch x.Kind {
	case Campaign, Intercept, Scout:
		return x.Target
	case Relief:
		if x.Target >= 0 {
			for _, eid := range sortedInts(w.Civs[x.Target].Wars) {
				return eid
			}
		}
	}
	return -1
}

// sightingBuyer is who a sighting is offered to: the people the fleet is
// coming for, when the seller is neither its ally nor its enemy, else a
// partner at war with the fleet's owner; a partner now or once, in
// either case.
func (w *World) sightingBuyer(c *Civ, x *Expedition) *Civ {
	owner := w.Civs[x.Owner]
	traded := func(e *Civ) bool { return c.Trade[e.ID] || c.partners[e.ID] }
	if b := w.bound(x); b >= 0 {
		e := w.Civs[b]
		if e != c && e != owner && traded(e) && !w.allied(c, e) && !c.Wars[b] && w.dealable(c, e) {
			return e
		}
	}
	for _, eid := range sortedInts(c.Met) {
		e := w.Civs[eid]
		if e != owner && e.Wars[owner.ID] && traded(e) && w.dealable(c, e) {
			return e
		}
	}
	return nil
}

// Answers.

// answerOffer is the asked side weighing an offer: what it would give
// against what it would get, at its margin, through its folly.
func (w *World) answerOffer(to, from *Civ, m *Message) {
	k := w.Contracts[m.Contract]
	if k.State != Offered {
		return
	}
	if !to.Active() || !from.Active() || !w.mutual(to, from) || to.Wars[from.ID] || !to.Species.Profile().Can(species.Trades) {
		k.State = Lapsed
		return
	}
	gives, gets := k.Ask, k.Pay
	if to.ID == k.Buyer {
		gives, gets = k.Pay, k.Ask
	}
	what := "asked by the " + from.Tok() + " for " + w.termName(gives) + " for " + w.termName(gets)
	if why := w.refuses(to, gives, from); why != "" {
		k.State = Refused
		if w.Cfg.TraceAI {
			w.event(KReason, to, nil, -1, P{"what": what, "why": why})
		}
		return
	}
	in := mind.OfferInput{
		Gives: gives.Kind, Gets: gets.Kind,
		GiveWorth: w.worthFrom(to, k, gives, from), GetWorth: w.worthTo(to, k, gets, from),
		Greed: to.Dials.Greed, Fixation: to.Morality.Object, Crime: crimeToGive(to, gives),
		Xenophobe: to.Has("xenophobic"), Different: to.differs(from) >= 1, Ignores: to.Morality.Kind == Herd || to.Morality.Kind == Amoral,
		Wis: to.Wis, Noise: w.R.NormFloat64(),
	}
	d := mind.AnswerOffer(in, w.Cfg.Tuning)
	w.explain(to, what, d)
	if !d.Accept {
		k.State = Refused
		from.Tally.Refused++
		if w.R.Float64() < 0.3 {
			w.event(KOfferRefused, from, to, -1, P{"ask": gets})
		}
		return
	}
	k.State = Accepted
	w.send(to, from, &Message{Kind: MsgAnswer, Contract: k.ID})
}

// flowWord is a commodity as a payment reads: grain, power, metal.
func flowWord(k flow.Kind) string {
	return [...]string{"grain", "power", "metal"}[k]
}

// answered is the answer arriving back at the proposer: the contract forms.
func (w *World) answered(to, from *Civ, m *Message) {
	k := w.Contracts[m.Contract]
	if k.State != Accepted {
		return
	}
	if !to.Active() || !from.Active() || to.Wars[from.ID] {
		k.State = Lapsed
		return
	}
	w.form(k)
}

// form is a contract starting: the record, the line, the fact, and each
// term's start.
func (w *World) form(k *Contract) {
	b, s := w.Civs[k.Buyer], w.Civs[k.Seller]
	k.State, k.Formed = Running, w.Now
	if k.Ask.Kind != mind.TermGuard {
		k.Until = w.Now + Year(k.Length*1000)
	}
	b.Tally.Hired++
	s.Tally.Sold++
	switch k.Ask.Kind {
	case mind.TermGuard:
		w.fact(FHire, s, b, k.Ask.Star).with(P{"pay": k.Pay, "against": k.Ask.Target})
	case mind.TermStrike, mind.TermDeliver:
		w.fact(FStrikeBought, b, w.Civs[k.Ask.Target], k.Ask.Star).with(P{"seller": s.ID, "pay": k.Pay, "ask": k.Ask})
	case mind.TermTeach:
		w.event(KTeachAgreed, s, b, -1, P{"node": k.Ask.Node, "pay": k.Pay})
	case mind.TermBroker:
		w.event(KBrokerAgreed, s, b, -1, P{"target": k.Ask.Target, "pay": k.Pay})
	case mind.TermSighting:
		// the sale's own line
	case mind.TermFlow:
		if !b.dealt[s.ID] {
			w.event(KBargain, b, s, -1, P{"ask": k.Ask, "pay": k.Pay, "first": true})
		}
	default:
		w.event(KBargain, b, s, -1, P{"ask": k.Ask, "pay": k.Pay, "first": false})
	}
	if b.dealt == nil {
		b.dealt = map[int]bool{}
	}
	b.dealt[s.ID] = true
	for i, t := range k.terms() {
		g, r := k.giver(i)
		if def := termKinds[t.Kind]; def.start != nil {
			def.start(w, k, w.Civs[g], w.Civs[r], t)
		}
	}
}

// The phase.

// tickContracts is the phase, after trade: each running contract's terms
// are checked for delivery, flows paid, breaks and lapses written, done
// ones closed, hired fleets tempted by the other side, and the sellsword's
// name earned.
func (w *World) tickContracts() {
	for _, k := range w.Contracts {
		if k.State != Running {
			continue
		}
		w.runContract(k)
	}
	for _, k := range w.Contracts {
		if k.State == Running {
			w.tempt(k)
		}
	}
	for _, c := range w.Civs {
		if c.Active() {
			w.sellsword(c)
		}
	}
}

// runContract is one running contract's tick.
func (w *World) runContract(k *Contract) {
	t := w.Cfg.Tuning.Contract
	b, s := w.Civs[k.Buyer], w.Civs[k.Seller]
	switch {
	case !b.Active():
		w.lapse(k, "the fall of the "+b.Tok())
		return
	case !s.Active():
		w.lapse(k, "the fall of the "+s.Tok())
		return
	}
	for i, term := range k.terms() {
		g, r := k.giver(i)
		def := termKinds[term.Kind]
		if def.lapse != nil {
			if why := def.lapse(w, k, w.Civs[g], w.Civs[r], term); why != "" {
				w.lapse(k, why)
				return
			}
		}
	}
	for i, term := range k.terms() {
		g, r := k.giver(i)
		giver, receiver := w.Civs[g], w.Civs[r]
		def := termKinds[term.Kind]
		if i == 0 && k.AskDone {
			continue // delivered already; only the pay runs
		}
		delivered := true
		if def.tick != nil {
			delivered = def.tick(w, k, giver, receiver, term)
		}
		if term.Kind.Lasting() && k.Until > 0 && w.Now >= k.Until {
			continue // its time is up; nothing more is owed
		}
		if delivered {
			k.Failed[i] = 0
			continue
		}
		k.Failed[i]++
		if i == 1 {
			k.Missed = true
		}
		if k.Failed[i] >= t.GraceTicks {
			w.breakContract(k, giver, receiver, term)
			return
		}
	}
	if def := termKinds[k.Ask.Kind]; def.done != nil && !k.AskDone {
		g, r := k.giver(0)
		if def.done(w, k, w.Civs[g], w.Civs[r], k.Ask) {
			k.AskDone = true
			if x := w.contractFleet(k); x != nil {
				// a strike done is a fleet released, whatever the pay has left to run
				w.event(KWorkDone, w.Civs[k.Seller], w.Civs[k.Buyer], -1, P{})
				w.releaseFleet(k, false)
			}
		}
	}
	done := k.Until > 0 && w.Now >= k.Until
	if termKinds[k.Ask.Kind].done != nil {
		done = k.AskDone && (k.Until == 0 || w.Now >= k.Until || !k.Pay.Kind.Lasting())
	}
	if done {
		w.finish(k)
	}
}

// flowTick pays a flow if the giver fed the word this tick.
func flowTick(w *World, k *Contract, giver, receiver *Civ, t Term) bool {
	if giver.Shed[wordKey(k.ID)] {
		return false
	}
	receiver.Paid[t.Res] += t.Amount
	giver.Tally.Sent[t.Res] += t.Amount
	return true
}

// contractUses lists what a people owes under its running contracts as
// uses in the word category.
func (w *World) contractUses(c *Civ) []flow.Use {
	var out []flow.Use
	for _, id := range c.Contracts {
		k := w.Contracts[id]
		if k.State != Running || (k.Until > 0 && w.Now >= k.Until) {
			continue
		}
		for i, t := range k.terms() {
			if g, _ := k.giver(i); g == c.ID && t.Kind == mind.TermFlow {
				var need flow.Income
				need[t.Res] = t.Amount
				out = append(out, flow.Use{Key: wordKey(k.ID), Cat: flow.Word, Era: 3, Need: need})
			}
		}
	}
	return out
}

// accessed lists the immobile rarities a people has the use of by
// contract, with who holds them.
func (w *World) accessed(c *Civ) []had {
	var out []had
	for _, id := range c.Contracts {
		k := w.Contracts[id]
		if k.State != Running {
			continue
		}
		for i, t := range k.terms() {
			if g, r := k.giver(i); r == c.ID && t.Kind == mind.TermAccess {
				out = append(out, had{w.Sources[t.Source], g})
			}
		}
	}
	return out
}

// finish closes a contract done.
func (w *World) finish(k *Contract) {
	k.State, k.Ended = Done, w.Now
	b, s := w.Civs[k.Buyer], w.Civs[k.Seller]
	if k.Ask.Kind == mind.TermGuard && w.contractFleet(k) != nil {
		w.event(KTermDone, b, s, k.Ask.Star, P{})
	}
	w.releaseFleet(k, false)
}

// lapse ends a contract without blame.
func (w *World) lapse(k *Contract, why string) {
	if k.State == Lapsed {
		return
	}
	k.State, k.Ended = Lapsed, w.Now
	w.releaseFleet(k, false)
}

// breakContract is a term failing for good: the giver is the breaker, a
// betrayal on the world; a broken flow ends the guard at once, and a
// broken fleet term sends the fleet home.
func (w *World) breakContract(k *Contract, giver, receiver *Civ, t Term) {
	k.State, k.Ended, k.Broke = Broken, w.Now, giver.ID
	giver.Tally.Broke++
	shape := "unpaid"
	switch t.Kind {
	case mind.TermGuard:
		shape = "left_star"
	case mind.TermStrike, mind.TermDeliver:
		shape = "no_strike"
	case mind.TermTeach:
		shape = "untaught"
	case mind.TermAccess:
		shape = "closed"
	case mind.TermPeace:
		shape = "broke_peace"
	case mind.TermFlow:
		if k.Ask.Kind == mind.TermGuard && giver.ID == k.Buyer {
			shape = "stopped_paying"
		}
	}
	w.betray(giver, receiver, shape, "broke", 1)
	w.releaseFleet(k, true)
}

// releaseFleet sends a contract's fleet home and ends a war declared for
// the hire.
func (w *World) releaseFleet(k *Contract, broken bool) {
	if x := w.contractFleet(k); x != nil && !x.Over && !x.Returning {
		x.Contract = -1
		if x.Base >= 0 && x.Kind == Campaign {
			w.resolve(x)
		} else {
			w.goHome(x)
		}
	}
	for _, wr := range w.Wars {
		if !wr.Over && wr.Hire == k.ID {
			s := w.Civs[k.Seller]
			w.event(KHireEnded, s, nil, -1, P{})
			w.endWar(wr, "hire_ended")
		}
	}
}

// tempt is the other side of a hired fleet offering its people a better
// bargain: peace with itself for a pay that beats the employer's at the
// seller's margin. The faithful refuse; the practical take it once the
// employer has failed them; the faithless at a chance by greed. A guard
// bought off turns on the world it held when its new payer is at war with
// the old.
func (w *World) tempt(k *Contract) {
	if k.Ask.Kind != mind.TermGuard && k.Ask.Kind != mind.TermStrike && k.Ask.Kind != mind.TermDeliver {
		return
	}
	if k.Ask.Target < 0 || k.Pay.Kind != mind.TermFlow {
		return
	}
	p, s, b := w.Civs[k.Ask.Target], w.Civs[k.Seller], w.Civs[k.Buyer]
	if !p.Active() || !w.mutual(p, s) || !w.hasSpareFor(p, s) {
		return
	}
	pay, ok := w.payFor(p, s)
	if !ok || pay.Kind != mind.TermFlow {
		return
	}
	offer := &Contract{Buyer: p.ID, Seller: s.ID, By: p.ID, Ask: Term{Kind: mind.TermPeace, Target: p.ID}, Pay: pay, Length: k.Length, Broke: -1}
	in := mind.OfferInput{Gives: mind.TermPeace, Gets: pay.Kind, Greed: s.Dials.Greed, Fixation: s.Morality.Object, Xenophobe: s.Has("xenophobic"), Different: s.differs(p) >= 1, Ignores: s.Morality.Kind == Herd || s.Morality.Kind == Amoral}
	bo := mind.BuyOffRate(mind.BuyOffInput{
		Honour: s.honour(), Greed: s.Dials.Greed, PayFailed: k.Missed,
		NewWorth: w.worthTo(s, offer, pay, p), OldWorth: w.worthTo(s, k, k.Pay, b), Margin: mind.Margin(in, w.Cfg.Tuning),
		Wis: s.Wis, Noise: w.R.NormFloat64(),
	}, w.Cfg.Tuning)
	if bo.Rate == 0 || !w.chance(bo.Rate) {
		return
	}
	w.explain(s, "offered a better bargain by the "+p.Tok(), bo)
	w.buyOff(k, offer, p, s, b)
}

// buyOff is the fleet sold: the old contract broken by the seller, the
// new one running, the fact, and the fleet turned or sent home.
func (w *World) buyOff(k, offer *Contract, p, s, b *Civ) {
	k.BoughtOff = true
	p.Tally.Hired++
	s.Tally.Sold++
	offer.ID = len(w.Contracts)
	offer.Offered, offer.Formed, offer.Until, offer.State = w.Now, w.Now, w.Now+Year(offer.Length*1000), Running
	w.Contracts = append(w.Contracts, offer)
	p.Contracts = append(p.Contracts, offer.ID)
	s.Contracts = append(s.Contracts, offer.ID)
	x := w.contractFleet(k)
	k.State, k.Ended, k.Broke = Broken, w.Now, s.ID
	s.Tally.Broke++
	s.Tally.BoughtOff++
	w.betray(s, b, "bought_off", "bought_off", 1)
	w.told(FBoughtOff, s, b, k.Ask.Star).with(P{"buyer": p.ID})
	for _, wr := range w.Wars {
		if !wr.Over && wr.Hire == k.ID {
			w.endWar(wr, "hire_sold")
		}
	}
	if x == nil || x.Over || x.Returning {
		return
	}
	x.Contract = -1
	if k.Ask.Kind == mind.TermGuard && x.Base == k.Ask.Star && p.Wars[b.ID] && b.Active() {
		w.turn(x)
		return
	}
	if x.Base >= 0 && x.Kind == Campaign {
		w.resolve(x)
	} else {
		w.goHome(x)
	}
}

// sellsword is the name earned: a people whose income of a kind has been
// more than half contract pay for ten ticks running is sellswords in the
// telling.
func (w *World) sellsword(c *Civ) {
	t := w.Cfg.Tuning.Contract
	living := false
	for _, k := range flow.Kinds {
		if c.PaidIn[k] > 0 && c.PaidIn[k] > t.SellswordShare*c.Income[k] {
			living = true
		}
	}
	if !living {
		c.hiredRun = 0
		return
	}
	c.hiredRun++
	if c.hiredRun >= t.SellswordTicks && !c.Sellsword {
		c.Sellsword = true
		w.event(KSellsword, c, nil, -1, P{})
	}
}

// Teaching.

// taughtNode is a bought node arriving: learned if the taught can still
// pursue it, else the contract lapses.
func (w *World) taughtNode(to, from *Civ, m *Message) {
	k := w.Contracts[m.Contract]
	n := tech.Get(m.Node)
	if k.State != Running || n == nil {
		return
	}
	if !w.canPursue(to, n) {
		w.lapse(k, "the "+to.Tok()+" could not take in what they were taught")
		return
	}
	if to.Taught == nil {
		to.Taught = map[string]int{}
	}
	to.Taught[n.Key] = from.ID
	k.Taught = true
	w.learn(to, n, false)
	w.fact(FTaught, from, to, -1).with(P{"node": n.Key, "pay": k.Pay})
}

// Tribute.

// tribute is a war ending in tribute instead of worlds: the loser asks
// peace and pays a flow of its largest surplus kind for twenty thousand
// years, when the winner is the kind that takes tribute over worlds and
// the loser has anything to pay. Returns whether it was written.
func (w *World) tribute(wr *War, l, v *Civ) bool {
	t := w.Cfg.Tuning.Contract
	if t.TributeLength <= 0 || !mind.TakesTribute(mind.TributeInput{Posture: v.posture(), Fixation: v.Morality.Object}) {
		return false
	}
	if !l.Species.Profile().Can(species.Trades) || !v.Species.Profile().Can(species.Trades) {
		return false // nothing the sim counts to pay with, or to take
	}
	spare := w.spare(l)
	best, bestSpare := flow.O, 0.0
	for _, k := range flow.Kinds {
		if spare[k] > bestSpare {
			best, bestSpare = k, spare[k]
		}
	}
	if bestSpare <= 0.1 {
		return false
	}
	k := w.newContract(l, v, Term{Kind: mind.TermPeace, Target: v.ID}, Term{Kind: mind.TermFlow, Res: best, Amount: bestSpare * w.transportCap(l)}, v)
	k.Length = t.TributeLength
	k.State, k.Formed, k.Until, k.Tribute = Running, w.Now, w.Now+Year(t.TributeLength*1000), true
	l.Tally.Tributes++
	w.fact(FTribute, l, v, -1).with(P{"res": best, "for": k.Until - w.Now}).with(w.warSpanP(wr))
	w.endWar(wr, "tribute")
	peaceStart(w, k, v, l, k.Ask)
	return true
}

// Worlds and works changing hands under a contract.

// handOver moves a world from one people to another as agreed: the star,
// the works standing at it, and the guns; the guard in its sky goes home.
func (w *World) handOver(from, to *Civ, s int) {
	if w.Owner[s] != from.ID || s == from.Home || !to.Active() {
		return
	}
	if g := w.guardAt(from, s); g != nil {
		w.goHome(g)
	}
	from.Systems = remove(from.Systems, s)
	w.Owner[s] = to.ID
	to.Systems = append(to.Systems, s)
	to.Peak = max(to.Peak, len(to.Systems))
	keep := from.Works[:0]
	for _, wk := range from.Works {
		if wk.Star == s {
			from.Structures[wk.Key]--
			to.Works = append(to.Works, wk)
			to.Structures[wk.Key]++
		} else {
			keep = append(keep, wk)
		}
	}
	from.Works = keep
	if g, ok := from.Guns[s]; ok {
		delete(from.Guns, s)
		if to.Guns == nil {
			to.Guns = map[int]int{}
		}
		to.Guns[s] = g
	}
	w.event(KHanded, from, to, s, P{})
	w.recompute(from)
	w.recompute(to)
}

// strikeBurns is a hired fleet that won the sky over a world it was paid
// to burn a work at: the work is burned and the world left. Returns
// whether it did.
func (w *World) strikeBurns(x *Expedition, e *Civ, t int) bool {
	k := w.contractOf(x)
	if k == nil || k.Ask.Kind != mind.TermStrike || k.Ask.Work == "" || k.Ask.Star != t || k.Burned {
		return false
	}
	c := w.Civs[x.Owner]
	keep := e.Works[:0]
	burned := false
	for _, wk := range e.Works {
		if !burned && wk.Star == t && wk.Key == k.Ask.Work {
			e.Structures[wk.Key]--
			w.leaveRuin(e, wk, "burned")
			burned = true
			continue
		}
		keep = append(keep, wk)
	}
	e.Works = keep
	k.Burned = true
	if burned {
		delete(e.Guns, t)
		w.event(KBurnedForPay, c, e, t, P{"work": k.Ask.Work})
		w.recompute(e)
	}
	return true
}

// sold is the owner of a fleet learning, when the fleet is met in the
// dark, who sold its coming: a betrayal, at a chance.
func (w *World) sold(q *Expedition) {
	if q.SoldBy < 0 {
		return
	}
	z := w.Civs[q.SoldBy]
	q.SoldBy = -1
	c := w.Civs[q.Owner]
	if !z.Active() || !c.Active() || w.R.Float64() >= w.Cfg.Tuning.Contract.SoldTold {
		return
	}
	w.betray(z, c, "sold", "sold", 1)
}

// Contracts lists a people's contracts, for the reports.
func Contracts(w *World, c *Civ) []*Contract {
	var out []*Contract
	for _, id := range c.Contracts {
		out = append(out, w.Contracts[id])
	}
	return out
}
