package history

// The kept orders: a set of people ids read in id order many times a
// tick and written rarely.
//
// Every loop that draws from the RNG reads a set through sortedInts,
// which allocates and sorts (util.go). A people's partners and the
// peoples it has met are read that way by the suspicion pass, the
// rarities, the slight, the intel and the garrison, once per reader per
// people per tick, and they change when a road opens or shuts and at no
// other time. So the order is kept beside the set and taken again only
// when the set moves.
//
// The duty this puts on a writer is the reason the architecture rules
// are wary of caches, so it is not left to memory: TestKeptOrders holds
// every kept order against a fresh sort at every tick of a whole age, so
// a write that forgets to drop the order fails the run that follows it.

// keptIDs is a set's ids in order, and whether the order still stands.
type keptIDs struct {
	ids []int
	ok  bool
}

// of is the set in id order, sorted again if the set has moved since.
// The size is checked beside the flag: the sim's writers all drop the
// order, but the small world tests put a people's roads and meetings
// into the maps by hand, and an order taken before that would be wrong
// for the rest of the test. The check costs a length and catches every
// add and every delete; a swap of the same size still needs the flag,
// which is what TestKeptOrders holds the sim to.
func (k *keptIDs) of(m map[int]bool) []int {
	if !k.ok || len(k.ids) != len(m) {
		k.ids, k.ok = sortedInts(m), true
	}
	return k.ids
}

// drop says the set has moved and the order must be taken again.
func (k *keptIDs) drop() { k.ok = false }

// startTrade and endTrade open and shut a road between two peoples, and
// are the only places the pair is written. A road is always two-sided:
// nobody trades with a people that does not trade back.
func startTrade(a, b *Civ) {
	a.Trade[b.ID], b.Trade[a.ID] = true, true
	a.tradeOrder.drop()
	b.tradeOrder.drop()
}

func endTrade(a, b *Civ) {
	delete(a.Trade, b.ID)
	delete(b.Trade, a.ID)
	a.tradeOrder.drop()
	b.tradeOrder.drop()
}

// tradeOf is a people's partners in id order.
func tradeOf(c *Civ) []int { return c.tradeOrder.of(c.Trade) }

// addMet and dropMet are the only places a people's roll of those it has
// heard of is written. Meeting is not always two-sided — the old may
// know the young without the young knowing them, and what is unseen is
// met by one party alone — so these take one people at a time, and a
// mutual meeting (World.meet, contact.go) calls each way.
func addMet(c *Civ, id int) {
	if !c.Met[id] {
		c.Met[id] = true
		c.metOrder.drop()
	}
}

func dropMet(c *Civ, id int) {
	if c.Met[id] {
		delete(c.Met, id)
		c.metOrder.drop()
	}
}

// metOf is everyone a people has heard of, in id order.
func metOf(c *Civ) []int { return c.metOrder.of(c.Met) }
