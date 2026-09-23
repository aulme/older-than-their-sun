package history

import (
	"math"
	"sort"
)

// Tellings. The chronicle is what happened. What a people knows of it is a
// telling: a set of tales, each a fact seen from one side, learned by
// witness, by word from another people, by reading a ruin or a relic, or
// by inheritance. Tales wear: a year is lost, then a name, then the thing
// is a myth, then it is gone. Machines and planetary minds wear slowly.
// A people slants what it tells: its own deeds grow, its enemies' deeds
// shrink and their crimes swell, and a friend that turns is retold. What a
// people leaves in a relic is its telling at that time.

// Sort is the moral shape of a fact from the subject's side.
type Sort uint8

const (
	Deed    Sort = iota // the subject did well; the object, if any, gained by it
	Crime               // the subject harmed the object
	Woe                 // the subject suffered; the object, if any, is the cause
	Bond                // the two did something together
	Folly               // the subject brought something on itself
	Nothing             // no judgment: what a morality passes over; never a fact's own sort
)

var sortNames = [...]string{"deed", "crime", "woe", "bond", "folly", "nothing"}

func (s Sort) String() string { return sortNames[s] }

// shape is what data/events.json declares of a kind: the parameters an
// event of it carries, and for a fact its sort and its weight. Weight is
// how much a tale resists wear and how far news of it travels; a kind
// with none is not a fact, and no people holds a tale of it.
type shape struct {
	Params  []string
	Sort    Sort
	Weight  float64
	Meaning string
	Silent  bool
}

// shapeOf is the kind's shape, found once per event: the tellings read
// it per tale per people per tick.
func (e *Event) shapeOf() *shape {
	if e.sh == nil {
		sh := shapes[e.Kind]
		e.sh = &sh
	}
	return e.sh
}

func (e *Event) sort() Sort      { return e.shapeOf().Sort }
func (e *Event) weight() float64 { return e.shapeOf().Weight }

// Silent says whether an event is a silent kind: a durable change the
// fold reads and nothing else.
func (e *Event) Silent() bool { return e.shapeOf().Silent }

// IsFact says whether an event is one a people can hold a tale of.
func (e *Event) IsFact() bool { return e.shapeOf().Weight > 0 }

// Weight is a fact's weight, 0 for an event that is not one; Sort its
// moral shape. The names pass reads them to rank the deeds between two
// peoples.
func (e *Event) Weight() float64 { return e.weight() }
func (e *Event) Sort() Sort      { return e.sort() }

// Provenance is how a people came to know a tale.
type Provenance uint8

const (
	Witnessed Provenance = iota
	Told                 // by another people, with their slant
	Read                 // from a ruin at a star, or a relic's testament
	Inherited            // from the people this one came out of, or its own relic
)

func (s Provenance) String() string { return [...]string{"witnessed", "told", "read", "inherited"}[s] }

// Tale is a fact as one people holds it.
type Tale struct {
	Fact    int
	Learned Year
	Source  Provenance
	From    int  // who told it, or whose relic
	Slant   int8 // the teller's regard for the other party when last told: -2 monsters, -1 enemies, 0 strangers, 1 friends
	Wear    int8 // 0 exact, 1 worn, 2 myth
	Blamed  int  // whom the tale now blames instead: the doer of a crime, the cause of a woe; or -1
	Revised int8 // times rewritten because a regard changed
	Forgot  bool
}

// Inscription is a tale as it stood when it was written down: the tale,
// the maker's judgment of it then, and what the telling read of the
// world then (Frozen), so that a reader renders it as the maker told it.
type Inscription struct {
	Tale   Tale
	Sort   Sort
	Weight float64
	Frozen Frozen
}

// Frozen is what a telling reads of the world besides the tale: of the
// parties in it (subject, object, blamed), which the teller can hold in
// mind, which it has met, which are rising, which live; and whether the
// star is its own. A testament keeps these as they were.
type Frozen struct {
	Perceived, Met, Active, Living []int
	Ours                           bool
}

// freeze reads what a telling of a tale would read of the world now.
func (w *World) freeze(c *Civ, t *Tale) Frozen {
	f := w.Events[t.Fact]
	var fz Frozen
	seen := map[int]bool{}
	for _, id := range []int{f.Subject, f.Object, t.Blamed} {
		if id < 0 || seen[id] {
			continue
		}
		seen[id] = true
		e := w.Civs[id]
		if w.perceives(c, e) {
			fz.Perceived = append(fz.Perceived, id)
		}
		if c.Met[id] {
			fz.Met = append(fz.Met, id)
		}
		if e.Active() {
			fz.Active = append(fz.Active, id)
		}
		if e.Living() {
			fz.Living = append(fz.Living, id)
		}
	}
	if s := f.Star; s >= 0 {
		fz.Ours = s == c.Home || s == c.Cradle || contains(c.Systems, s) || w.Owner[s] == c.ID
	}
	return fz
}

// event writes a happening the chronicle keeps and no people holds a
// tale of: the kind has no weight. The subject and the object may be
// nil; the parameters are the kind's declared ones.
func (w *World) event(k Kind, c, e *Civ, star int, p P) *Event {
	return w.eventAt(w.Now, k, c, e, star, p)
}

// eventAt is an event at a year other than now: the deep past, or a year
// inside the tick.
func (w *World) eventAt(y Year, k Kind, c, e *Civ, star int, p P) *Event {
	ev := w.recordAt(y, k, c, e, star, p)
	w.Chronicle = append(w.Chronicle, ev)
	return ev
}

// recordAt writes an event and gives it its id, without placing it in
// the chronicle: the record's order is when things happened, the
// chronicle's is when they are told, and a fact recorded in the middle
// of something is told where the something is. Every event is placed
// once; fill and the event helpers do it.
func (w *World) recordAt(y Year, k Kind, c, e *Civ, star int, p P) *Event {
	if p == nil {
		p = P{}
	}
	ev := &Event{ID: len(w.Events), Year: y, Kind: k, Subject: -1, Object: -1, Star: star, Legacy: -1, Plague: -1, P: p}
	if sh, ok := shapes[k]; ok {
		ev.sh = &sh
	}
	if row, ok := moralTable[k]; ok {
		ev.row = &row
	}
	if c != nil {
		ev.Subject = c.ID
	}
	if e != nil {
		ev.Object = e.ID
	}
	w.Events = append(w.Events, ev)
	return ev
}

// slot reserves the next place in the chronicle, for a fact that will be
// recorded once what it is told before has been.
func (w *World) slot() int {
	w.Chronicle = append(w.Chronicle, nil)
	return len(w.Chronicle) - 1
}

// place puts an unplaced event at the end of the chronicle: told after
// what it caused.
func (w *World) place(e *Event) *Event {
	w.Chronicle = append(w.Chronicle, e)
	return e
}

// fill places an unplaced event in a reserved slot.
func (w *World) fill(slot int, e *Event) *Event {
	w.Chronicle[slot] = e
	return e
}

// fact records something that happened, places it, and lets the parties
// know it: what anyone makes of it is told after it.
func (w *World) fact(k Kind, c, e *Civ, star int) *Event {
	f := w.record(k, c, e, star)
	w.Chronicle = append(w.Chronicle, f)
	if e != nil {
		w.witness(e, f)
	}
	w.spread(f)
	return f
}

// told is a fact placed after the parties know it: what anyone makes of
// it at once is told before it. The chronicle kept both orders, and
// keeps them.
func (w *World) told(k Kind, c, e *Civ, star int) *Event {
	return w.place(w.unplaced(k, c, e, star))
}

// unplaced is a fact recorded, witnessed and spread but not yet placed in
// the chronicle; place or fill places it.
func (w *World) unplaced(k Kind, c, e *Civ, star int) *Event {
	f := w.record(k, c, e, star)
	if e != nil {
		w.witness(e, f)
	}
	w.spread(f)
	return f
}

// record writes a fact and lets the subject know it, unplaced; fact adds
// the object and the spread. The parameters the line wants are the
// caller's to add, before the tick is out; nothing in between reads them.
func (w *World) record(k Kind, c, e *Civ, star int) *Event {
	f := w.recordAt(w.Now, k, c, e, star, nil)
	if star >= 0 {
		w.factsAt[star] = append(w.factsAt[star], f.ID)
	}
	w.witness(c, f)
	return f
}

// with adds parameters to an event, and is the event.
func (e *Event) with(p P) *Event {
	for k, v := range p {
		e.P[k] = v
	}
	return e
}

// meeting is the meeting fact, with how: "touch" for territories met in the
// flesh, "signal" for a hearing. Every name a people has for another
// starts from one of these, so every way of meeting writes one. The
// parameters the line wants are the caller's to add.
func (w *World) meeting(a, b *Civ, at int, how string) *Event {
	return w.told(FMet, a, b, at).with(P{"how": how})
}

// noticed is a one-sided meeting: the seer finds the other, who never
// knows. The seer alone holds the tale; nothing spreads from it.
func (w *World) noticed(seer, unseen *Civ, at int) *Event {
	f := w.record(FMet, seer, unseen, at).with(P{"how": "noticed"})
	w.Chronicle = append(w.Chronicle, f)
	return f
}

// factAt is a fact at a year inside the tick.
func (w *World) factAt(y Year, k Kind, c, e *Civ, star int) *Event {
	f := w.fact(k, c, e, star)
	f.Year = y
	return f
}

// factN is a fact with a count.
func (w *World) factN(k Kind, c, e *Civ, star, n int) *Event {
	f := w.fact(k, c, e, star)
	f.N = n
	return f
}

// factL is a fact with a remain in it, placed first.
func (w *World) factL(k Kind, c *Civ, l *Legacy) *Event {
	var m *Civ
	if l.Maker >= 0 && l.Maker != c.ID {
		m = w.Civs[l.Maker]
	}
	f := w.fact(k, c, m, l.Star)
	f.Legacy = l.ID
	return f
}

// toldL is a fact with a remain in it, told after.
func (w *World) toldL(k Kind, c *Civ, l *Legacy) *Event {
	var m *Civ
	if l.Maker >= 0 && l.Maker != c.ID {
		m = w.Civs[l.Maker]
	}
	f := w.told(k, c, m, l.Star)
	f.Legacy = l.ID
	return f
}

// witness is a party to a fact learning it as it happened.
func (w *World) witness(c *Civ, f *Event) {
	if !c.Living() || c.knows(f.ID) {
		return
	}
	c.Tally.Witnessed++
	w.hold(c, f, Witnessed, -1, 0, 0)
}

// knows says whether a people holds a tale of a fact, forgotten or not.
func (c *Civ) knows(fact int) bool { return c.lore[fact] }

// regard is how a people sees another right now: -2 a monster, -1 an
// enemy, 0 a stranger, 1 a friend. Itself is 2.
func (w *World) regard(c *Civ, id int) int8 {
	if id < 0 || !w.perceives(c, w.Civs[id]) {
		return 0 // nothing can be felt toward what cannot be held in mind
	}
	if c.ofLine(id) {
		return 2 // the line's deeds are ours
	}
	e := w.Civs[id]
	switch {
	case c.hates(e) || w.monster(c, e):
		return -2
	case c.Wars[id] || c.Grudge[id] > 0.5 || (c.Master == id && !c.Vassal):
		return -1
	case c.Trade[id] || w.allied(c, e) || (e.Master == c.ID && e.Vassal) || (c.Master == id && c.Vassal) || w.warm(c, e):
		return 1
	}
	return 0
}

// other is the party in a fact that is not this people, or -1.
func (c *Civ) other(f *Event) int {
	if f.Subject != c.ID {
		return f.Subject
	}
	return f.Object
}

// learn adds a tale. A tale told by another people carries the teller's
// regard for the other party when the learner has none of its own; that
// is how a stranger comes to be a monster to peoples it never met.
func (w *World) hold(c *Civ, f *Event, src Provenance, from int, slant int8, wear int8) *Tale {
	if !w.keeps(c, f) {
		return nil // nothing about what cannot be held in mind, unless it is our own loss; see antimemetic.go
	}
	if c.lore == nil {
		c.lore = map[int]bool{}
	}
	c.lore[f.ID] = true
	t := &Tale{Fact: f.ID, Learned: w.Now, Source: src, From: from, Wear: wear, Blamed: -1}
	o := c.other(f)
	t.Slant = w.regard(c, o)
	if src == Told && t.Slant == 0 && o >= 0 && o != from {
		switch {
		case slant < 0:
			t.Slant = -1
		case slant > 0:
			t.Slant = 1
		}
	}
	c.Lore = append(c.Lore, t)
	c.Tally.Tales++
	w.learned(c, f, t)
	w.takeToHeart(c, f, t)
	if len(c.Lore) > 600 {
		w.prune(c)
	}
	return t
}

// takeToHeart is what learning a tale does at once: a crime against a
// friend is held against the doer, and a crime by a stranger against
// anyone is a thing to be feared.
func (w *World) takeToHeart(c *Civ, f *Event, t *Tale) {
	w.judgeLine(c, f, t)
	s, wt := sortFor(c, f)
	if s != Crime || f.Subject == c.ID || t.Source == Witnessed && f.Object == c.ID {
		return
	}
	if f.Object >= 0 && f.Object != c.ID && w.regard(c, f.Object) > 0 && w.seen(c, f.Subject) >= 0 {
		c.resent(f.Subject, 0.1*wt/3)
	}
}

// judgeLine is the note, now and then, that a people's judgment of a
// thing it has just learned is not the fact's own: a crime it counts no
// crime, a deed it calls one.
func (w *World) judgeLine(c *Civ, f *Event, t *Tale) {
	if t.Source == Inherited || f.Subject == c.ID || f.Object == c.ID || !judges(c, f) || w.R.Float64() >= 0.05 {
		return
	}
	s, _ := sortFor(c, f)
	switch {
	case f.sort() == Crime && s == Deed:
		w.event(KJudged, c, w.Civs[f.Subject], -1, P{"about": f.ID, "verdict": "deed"})
	case f.sort() == Crime && s == Nothing:
		w.event(KJudged, c, w.Civs[f.Subject], -1, P{"about": f.ID, "verdict": "nothing"})
	case f.sort() != Crime && s == Crime:
		w.event(KJudged, c, w.Civs[f.Subject], -1, P{"about": f.ID, "verdict": "crime"})
	}
}

// monster says whether a people reads another as a thing that does harm:
// enough crimes held against it, weighed by whom they were done to and
// how far into myth they have gone, reckoned once a tick; or a people met
// and not yet fathomed (wisdom.go), which is a monster until it is
// understood; or a thing that is a monster by its nature, to everyone.
func (w *World) monster(c, e *Civ) bool {
	return c.monsters[e.ID] || w.unfathomed(c, e) || e.Species.Profile().Monster
}

// reckon works out whom a people remembers as monsters. It is the one
// summary of a telling that is still walked every tick: what a crime
// weighs turns on whom it was done to — a trading partner, an ally — and
// on whether the doer can be perceived at all, and those move under the
// telling from tick to tick, so a tale's part in it cannot be kept. See
// summaries.go for the two that can.
func (w *World) reckon(c *Civ) {
	x := &w.crimes
	x.start(len(w.Civs))
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Events[t.Fact]
		s, wt := sortFor(c, f)
		if s == Woe {
			if t.Blamed < 0 || t.Blamed == c.ID {
				continue
			}
			v := 0.5
			if f.Subject == c.ID {
				v = 2
			}
			x.add(t.Blamed, v*wt/3*(1+0.5*float64(t.Wear)))
			continue
		}
		if c.ofLine(f.Subject) || (s != Crime && s != Folly) {
			continue // our own crimes, and our line's, are not held against us
		}
		v := 0.5
		switch {
		case f.Object == c.ID:
			v = 2
		case f.Object >= 0 && (c.Trade[f.Object] || w.allied(c, w.Civs[f.Object])):
			v = 1
		}
		who := w.seen(c, f.Subject)
		if t.Blamed >= 0 {
			who = t.Blamed
		}
		if who < 0 {
			continue // a crime with no doer is held against nobody, until somebody is blamed for it
		}
		x.add(who, v*wt/3*(1+0.5*float64(t.Wear)))
	}
	if c.monsters == nil {
		c.monsters = map[int]bool{} // nil until the first reckoning; export.go reads that
	} else {
		clear(c.monsters)
	}
	for _, id := range x.touched {
		if x.sum[id] >= 3 {
			c.monsters[id] = true
		}
	}
}

// spread carries news: each party tells the peoples it trades with or is
// sworn to, and anyone who can hear its signals learns of the heavier
// things; whoever had a holding within sight of the star saw it happen.
func (w *World) spread(f *Event) {
	parties := []int{f.Subject}
	if f.Object >= 0 {
		parties = append(parties, f.Object)
	}
	for _, pid := range parties {
		p := w.Civs[pid]
		slant := w.regard(p, p.other(f))
		for _, eid := range sortedInts(p.Met) {
			e := w.Civs[eid]
			if !e.Active() || eid == f.Subject || eid == f.Object {
				continue
			}
			if p.Trade[eid] || w.allied(p, e) || (f.weight() >= 3 && w.hear(p, e)) {
				m := &Message{Kind: MsgNews, About: pid, Fact: f.ID, Slant: slant}
				if p.Living() {
					w.send(p, e, m)
				} else {
					// the last signal of a people that is gone
					m.From, m.To, m.Sent = pid, eid, w.Now
					m.Arrive = w.Now + Year(w.G.Dist(p.Home, e.Home))
					w.Messages = append(w.Messages, m)
				}
			}
		}
	}
	if f.Star < 0 || f.weight() < 3 || f.sort() == Bond {
		return
	}
	for _, e := range w.Civs {
		if !e.Active() || e.ID == f.Subject || e.ID == f.Object || e.knows(f.ID) {
			continue
		}
		r := max(e.watchRange(), 2)
		for _, s := range w.holdings(e) {
			if w.G.Dist(s, f.Star) <= r {
				w.hold(e, f, Witnessed, -1, 0, 0)
				e.Tally.Witnessed++
				break
			}
		}
	}
}

// news is a fact told by another people arriving.
func (w *World) news(to, from *Civ, m *Message) {
	if to.knows(m.Fact) {
		return
	}
	f := w.Events[m.Fact]
	wear := int8(0)
	if w.regard(to, from.ID) < 0 {
		wear = 1 // an enemy's word is a rumour
	}
	to.Tally.Told++
	w.hold(to, f, Told, from.ID, m.Slant, wear)
}

// readRuins is what a visit to a star teaches of what happened there: the
// heavier things, as fragments, with the names on the walls.
func (w *World) readRuins(c *Civ, star int) {
	w.readWalls(c, star)
	ids := w.factsAt[star]
	if len(ids) == 0 {
		return
	}
	var pick []*Event
	for _, id := range ids {
		f := w.Events[id]
		if c.knows(id) || f.weight() < 2 || f.sort() == Bond {
			continue
		}
		switch f.Kind {
		case FSettle, FTaken, FBurned, FHomeBroken, FScoured, FEnd, FFall, FUnleashed, FWaking, FUnmade, FArise, FExodus, FStarDied, FLeftStar, FVacuumHole:
			pick = append(pick, f)
		}
	}
	sort.SliceStable(pick, func(i, j int) bool { return pick[i].weight() > pick[j].weight() })
	for i, f := range pick {
		if i >= 3 {
			break
		}
		c.Tally.Read++
		w.hold(c, f, Read, -1, 0, 1)
	}
}

// testament writes a people's telling into a remain it leaves: the tales
// it holds dearest, as it tells them now.
func (w *World) testament(c *Civ, l *Legacy) {
	if len(c.Lore) == 0 {
		return
	}
	var keep []*Tale
	for _, t := range c.Lore {
		if !t.Forgot {
			keep = append(keep, t)
		}
	}
	sort.SliceStable(keep, func(i, j int) bool {
		fi, fj := w.Events[keep[i].Fact], w.Events[keep[j].Fact]
		wi, wj := w.dearness(c, fi, keep[i]), w.dearness(c, fj, keep[j])
		if wi != wj {
			return wi > wj
		}
		return fi.Year > fj.Year
	})
	if len(keep) > 8 {
		keep = keep[:8]
	}
	sort.SliceStable(keep, func(i, j int) bool { return w.Events[keep[i].Fact].Year < w.Events[keep[j].Fact].Year })
	for _, t := range keep {
		sort, weight := sortFor(c, w.Events[t.Fact])
		l.Testament = append(l.Testament, Inscription{Tale: *t, Sort: sort, Weight: weight, Frozen: w.freeze(c, t)})
	}
	c.Tally.Testaments++
	w.wallsWritten(c, l)
}

// mythParty is the party the chronicle's note of a myth names, if it
// names one: -1 when the fact has none or the line needs none. The view
// keeps the same rule for the line itself.
func mythParty(f *Event) int {
	switch f.Kind {
	case FEnd, FFall, FBetrayal, FCutOff, FWaking:
		return f.Subject
	case FEnslaved, FFreed, FBred:
		return f.Object
	case FHomeBroken, FScoured, FUnleashed, FDarkAge, FWant, FSundered, FShattered, FSevered, FUnmade, FExodus:
		return -1
	}
	if f.Star >= 0 {
		return -1
	}
	return f.Object
}

// dearness is how much a tale matters to its teller: the fact's weight,
// more for its own part in it, less for strangers' business.
func (w *World) dearness(c *Civ, f *Event, t *Tale) float64 {
	_, x := sortFor(c, f)
	switch {
	case f.Subject == c.ID || f.Object == c.ID:
		x *= 1.5
	case t.Slant == 0:
		x *= 0.5
	}
	return x
}

// readTestament is a finder learning what a relic's makers told. Its own
// relic gives its own memory back; another's carries the makers' slant
// into a people that had none of its own.
func (w *World) readTestament(c *Civ, l *Legacy) {
	if len(l.Testament) == 0 || c.inscribed[l.ID] {
		return
	}
	if c.inscribed == nil {
		c.inscribed = map[int]bool{}
	}
	c.inscribed[l.ID] = true
	w.readWallsPlague(c, l)
	own := w.kinship(c, l) == 2
	n := 0
	for _, td := range l.Testament {
		f := w.Events[td.Tale.Fact]
		if own {
			if restored := w.restore(c, f.ID, td.Tale); restored {
				n++
			}
			continue
		}
		if c.knows(f.ID) {
			continue
		}
		wear := min(2, td.Tale.Wear+1)
		src := Read
		w.hold(c, f, src, l.Maker, td.Tale.Slant, wear)
		c.Tally.Read++
		n++
	}
	if n > 0 && w.R.Float64() < 0.3 {
		w.event(KReadWalls, c, nil, l.Star, P{"own": own, "known": w.knowsMaker(c, l)}).Legacy = l.ID
	}
}

// restore gives a people back a tale from its own relic: one it had
// forgotten, one it never held of its own past, or one worn past what the
// wall says, which is exact. The blame stands as it was written. Returns
// whether anything came back.
func (w *World) restore(c *Civ, fact int, was Tale) bool {
	for _, t := range c.Lore {
		if t.Fact != fact {
			continue
		}
		if !t.Forgot && t.Wear <= was.Wear {
			return false
		}
		w.amend(c, t, func(t *Tale) { t.Forgot, t.Wear = false, was.Wear })
		t.Blamed = was.Blamed
		t.Source = Inherited
		c.Tally.Restored++
		return true
	}
	f := w.Events[fact]
	t := w.hold(c, f, Inherited, c.ID, 0, was.Wear)
	if t == nil {
		return false // what the wall says is about something that cannot be held in mind
	}
	t.Blamed = was.Blamed
	c.Tally.Restored++
	return true
}

// readWalls is what a visit to a star reads off the remains there: the
// tellings their makers wrote, whoever they were and whether or not the
// visitors could use the thing. Once per remain per people.
func (w *World) readWalls(c *Civ, star int) {
	for _, l := range w.Legacies {
		if l.Star != star || l.State == Lost || len(l.Testament) == 0 || c.inscribed[l.ID] {
			continue
		}
		w.readTestament(c, l)
	}
}

// inherit gives a people born of another that people's telling: a branch
// keeps its parent's tales as its own history, a made people is told what
// its makers want it to know.
func (w *World) inherit(nc, parent *Civ, wear int8) {
	w.inheritImmunity(nc, parent)
	for _, t := range parent.Lore {
		if t.Forgot || nc.knows(t.Fact) {
			continue
		}
		nt := w.hold(nc, w.Events[t.Fact], Inherited, parent.ID, t.Slant, min(2, t.Wear+wear))
		if nt == nil {
			continue
		}
		nc.Tally.Inherited++
		nt.Blamed = t.Blamed
	}
}

// exchange is two peoples at a first meeting trading what they know of
// others: the heaviest tales about third parties, each with its slant.
func (w *World) exchange(a, b *Civ) {
	w.tellOf(a, b)
	w.tellOf(b, a)
}

func (w *World) tellOf(from, to *Civ) {
	var pick []*Tale
	for _, t := range from.Lore {
		f := w.Events[t.Fact]
		if t.Forgot || t.Wear >= 2 || f.weight() < 2 || to.knows(f.ID) {
			continue
		}
		if f.Subject == to.ID || f.Object == to.ID {
			continue
		}
		pick = append(pick, t)
	}
	sort.SliceStable(pick, func(i, j int) bool {
		return w.Events[pick[i].Fact].weight() > w.Events[pick[j].Fact].weight()
	})
	for i, t := range pick {
		if i >= 5 {
			break
		}
		to.Tally.Told++
		w.hold(to, w.Events[t.Fact], Told, from.ID, t.Slant, t.Wear)
	}
}

// memory is how well a people keeps a tale: a multiplier on the rate of
// wear. What a people is made of and how it is shaped set the base (the
// profile: machines and living worlds barely wear, a hive shares one
// memory); the arts of writing, printing, networks and substrate minds
// each slow the loss.
func (w *World) memory(c *Civ) float64 {
	m := c.Species.Profile().Memory
	if c.Has("swarming") {
		m *= 0.8
	}
	if c.Has("collective") {
		m *= 0.7
	}
	for _, row := range memoryTable {
		if c.Known[row.node] {
			m *= row.keep
		}
	}
	if c.Boons[BoonCommunion] {
		m *= 0.5
	}
	return m
}

var memoryTable = []struct {
	node string
	keep float64
}{
	{"writing", 0.6}, {"printing", 0.7}, {"networks", 0.6}, {"substrate_minds", 0.3}, {"long_thought", 0.3},
}

// wear is the tick step: tales age into myth and out of memory. A tale's
// rate rises with its age and falls with its weight; a people's own part
// in a thing is kept longer than strangers' business, and its own crimes
// are the first it lets go.
func (w *World) wear(c *Civ) {
	if !c.Living() || len(c.Lore) == 0 {
		return
	}
	every := max(w.Cfg.WearEvery, 1)
	if every > 1 {
		// the peoples are staggered across the gap, so that no tick
		// carries every telling and none carries none
		if (w.Ticks+c.ID)%every != 0 {
			return
		}
	}
	m := w.memory(c)
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Events[t.Fact]
		age := float64(w.Now-f.Year) / 1e6
		if age < 0.05 {
			continue
		}
		s, wt := sortFor(c, f)
		rate := 0.002 * m * min(1+age, 3) / max(wt, 0.5)
		own := c.ofLine(f.Subject) || c.ofLine(f.Object) // the line's tales are kept as our own
		switch {
		case c.ofLine(f.Subject) && s == Crime, c.ofLine(f.Subject) && s == Folly:
			rate *= 2 // what we did is easier to forget
		case own:
			rate *= 0.6
		case t.Slant == 0:
			rate *= 1.5
		}
		if t.Wear >= 2 && own && wt >= 3 {
			rate *= 0.25 // the old songs
		}
		if !w.chance(over(rate, every)) {
			continue
		}
		w.wearStep(c, t, f)
	}
}

// over is the chance of a thing of per-tick chance p happening at least
// once in n ticks. A telling put through the wearing every n ticks with
// this in place of the rate wears as often as one put through it every
// tick; it is the same process read at a coarser step.
func over(p float64, n int) float64 {
	if n <= 1 || p <= 0 {
		return p
	}
	if p >= 1 {
		return 1
	}
	return 1 - math.Pow(1-p, float64(n))
}

// wearStep takes one tale one step toward myth, and at the myth step may
// hang the blame on whoever is the enemy now.
func (w *World) wearStep(c *Civ, t *Tale, f *Event) {
	if t.Wear >= 2 {
		w.amend(c, t, func(t *Tale) { t.Forgot = true })
		c.Tally.Forgot++
		return
	}
	w.amend(c, t, func(t *Tale) { t.Wear++ })
	if t.Wear < 2 {
		return
	}
	c.Tally.Myths++
	if f.Kind == FSundered && f.Object == c.ID && c.Claim != nil {
		c.Claim = nil // the sundering is a story now, and a faction is a people
		w.event(KClaimForgot, c, nil, -1, P{"about": f.ID})
	}
	blame := false
	s, wt := sortFor(c, f)
	switch {
	case s == Folly && f.Subject == c.ID:
		blame = w.R.Float64() < 0.4 // our own folly becomes somebody's doing
	case (s == Crime || s == Folly) && f.Subject != c.ID && t.Slant == 0:
		blame = w.R.Float64() < 0.3 // a stranger's crime is hung on the enemy of the day
	}
	if blame {
		if e := w.foe(c); e >= 0 && e != f.Subject {
			t.Blamed = e
			c.Tally.Blamed++
			if w.R.Float64() < 0.3 {
				w.event(KBlamed, c, w.Civs[e], -1, P{"about": f.ID})
			}
			return
		}
	}
	if wt >= 4 && (f.Subject == c.ID || f.Object == c.ID) && w.R.Float64() < 0.15 {
		who := mythParty(f)
		w.event(KMyth, c, nil, -1, P{"about": f.ID, "nameless": who >= 0 && w.seen(c, who) < 0})
	}
}

// foe is the enemy of the day: the people this one holds the most against,
// alive or dead; failing a grudge, whoever it is at war with, or remembers
// as a monster. -1 when there is nobody.
func (w *World) foe(c *Civ) int {
	best, bg := -1, 0.3
	for _, id := range sortedInts(c.Met) {
		if g := c.Grudge[id]; g > bg && w.perceives(c, w.Civs[id]) {
			best, bg = id, g
		}
	}
	if best >= 0 {
		return best
	}
	for _, id := range sortedInts(c.Wars) {
		if c.Wars[id] && w.perceives(c, w.Civs[id]) {
			return id // a war with what cannot be named has no enemy of the day in it
		}
	}
	for _, id := range sortedInts(c.monsters) {
		if c.monsters[id] {
			return id
		}
	}
	return -1
}

// scapegoat is what a new enemy inherits: the old wrongs nobody was blamed
// for. Every worn or mythic crime, folly or woe whose doer is not the new
// foe, and is ourselves, a stranger, a people dead or never met, or nobody
// at all, is hung on the foe: a quarter of them at worn, half at myth.
func (w *World) scapegoat(c *Civ, e int) {
	var blamed []int
	n := 0
	for _, t := range c.Lore {
		if t.Forgot || t.Wear < 1 || t.Blamed >= 0 {
			continue
		}
		f := w.Events[t.Fact]
		s, _ := sortFor(c, f)
		if s != Crime && s != Folly && s != Woe {
			continue
		}
		doer := w.seen(c, f.Subject)
		if s == Woe {
			doer = w.seen(c, f.Object)
		}
		if doer == e || f.Subject == e || f.Object == e {
			continue
		}
		fits := false
		switch {
		case doer < 0:
			fits = true // nobody's doing, until now
		case doer == c.ID:
			fits = s == Folly || s == Woe // our own follies and sufferings
		case t.Slant == 0, !c.Met[doer], !w.Civs[doer].Living():
			fits = true
		}
		if !fits {
			continue
		}
		p := 0.25
		if t.Wear >= 2 {
			p = 0.5
		}
		if w.R.Float64() >= p {
			continue
		}
		t.Blamed = e
		c.Tally.Blamed++
		n++
		blamed = append(blamed, f.ID)
	}
	if n == 0 {
		return
	}
	w.event(KScapegoat, c, w.Civs[e], -1, P{"facts": blamed})
}

// revise is the propaganda step: when a people's regard for another
// changes sign, every tale with that people in it is retold to fit. A
// retelling wears the tale a little; it is never quite the same after.
func (w *World) revise(c *Civ) {
	if len(c.Lore) == 0 {
		return
	}
	if e := w.foe(c); e != c.foeNow {
		c.foeNow = e
		if e >= 0 {
			w.scapegoat(c, e)
		}
	}
	held := w.regardHolder(c)
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Events[t.Fact]
		o := c.other(f)
		if o < 0 {
			continue
		}
		r := held(o)
		if r == t.Slant || r == 0 {
			continue // indifference is wear's business, not the retellers'
		}
		if t.Slant == 0 || (r > 0) == (t.Slant > 0) {
			t.Slant = r // a first opinion, or a deeper one the same way
			continue
		}
		t.Slant = r
		t.Revised++
		c.Tally.Revised++
		if t.Wear < 2 && w.R.Float64() < 0.5 {
			w.amend(c, t, func(t *Tale) { t.Wear++ })
		}
	}
}

// regardHolder answers regard for one people about others, holding each
// answer for as long as the caller keeps the function. It is for a walk
// of a whole telling, where the same few parties come up again and
// again; a caller must not use it across anything that changes what
// regard reads.
func (w *World) regardHolder(c *Civ) func(id int) int8 {
	if len(w.regardSeen) < len(w.Civs) {
		w.regardOf = make([]int8, len(w.Civs))
		w.regardSeen = make([]uint32, len(w.Civs))
		w.regardGen = 0
	}
	w.regardGen++
	if w.regardGen == 0 { // the stamp wrapped: no answer is fresh
		clear(w.regardSeen)
		w.regardGen = 1
	}
	gen := w.regardGen
	return func(id int) int8 {
		if id < 0 || id >= len(w.regardSeen) {
			return w.regard(c, id)
		}
		if w.regardSeen[id] == gen {
			return w.regardOf[id]
		}
		r := w.regard(c, id)
		w.regardSeen[id], w.regardOf[id] = gen, r
		return r
	}
}

// forgetting is a dark age's toll on memory: most tales take a step, and
// what was already myth is lost.
func (w *World) forgetting(c *Civ) {
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Events[t.Fact]
		p := 0.6 * w.memory(c)
		if f.weight() >= 4 {
			p *= 0.5
		}
		if w.R.Float64() < p {
			w.wearStep(c, t, f)
		}
	}
}

// prune keeps a telling within bounds: forgotten tales are dropped, then
// the lightest of strangers' business.
func (w *World) prune(c *Civ) {
	keep := c.Lore[:0]
	for _, t := range c.Lore {
		if !t.Forgot {
			keep = append(keep, t)
		}
	}
	c.Lore = keep
	if len(c.Lore) <= 500 {
		return
	}
	sort.SliceStable(c.Lore, func(i, j int) bool {
		fi, fj := w.Events[c.Lore[i].Fact], w.Events[c.Lore[j].Fact]
		return w.dearness(c, fi, c.Lore[i]) > w.dearness(c, fj, c.Lore[j])
	})
	for _, t := range c.Lore[500:] {
		w.amend(c, t, func(t *Tale) { t.Forgot = true })
		c.Tally.Forgot++
	}
	c.Lore = c.Lore[:500]
	sort.SliceStable(c.Lore, func(i, j int) bool { return c.Lore[i].Learned < c.Lore[j].Learned })
}

// dread is whether a people remembers a star as somewhere its surveyors
// do not come back from: something let loose there, a survey lost, a
// world emptied by a plague, the Signal heard from there, or a deed at
// it by a people it remembers as a monster. A tale that has worn to myth
// no longer keeps anyone away.
func (w *World) dread(c *Civ, star int) bool {
	for _, t := range c.Lore {
		if t.Forgot || t.Wear >= 2 {
			continue
		}
		f := w.Events[t.Fact]
		if f.Star != star {
			continue
		}
		switch f.Kind {
		case FUnleashed, FSurveyLost, FPlagueWorld:
			return true
		case FScarred, FDeclined:
			if f.Legacy >= 0 {
				return true // the Signal, from there
			}
		case FTaken, FBurned, FStripped, FHomeBroken, FScoured, FWaking, FUnmade:
			if f.Subject != c.ID && f.Subject >= 0 && w.monster(c, w.Civs[f.Subject]) {
				return true
			}
		}
	}
	return false
}

// lore is the tick step for a people's telling.
func (w *World) loreStep(c *Civ) {
	if !c.Active() {
		return
	}
	w.wear(c)
	w.reckon(c)
	w.revise(c)
}

// LoreCounts is what a people holds at the end: tales, of them myth, and
// peoples it remembers as monsters.
func LoreCounts(w *World, c *Civ) (held, myth, monsters int) {
	for _, t := range c.Lore {
		if !t.Forgot {
			held++
			if t.Wear >= 2 {
				myth++
			}
		}
	}
	if c.monsters == nil {
		w.reckon(c)
	}
	return held, myth, len(c.monsters)
}

// A reason is the reason a fact gives for what happened: a key of
// data/causes.json and the ids its words name. It goes into the fact's
// parameters as the named key with by, at, plague and blast beside it
// where they are set; the view renders it, and nothing reads it.
type reason struct {
	key    string
	by     int // a people
	at     int // a star
	plague int // a plague
	blast  int // the blast event; for a waking, the waker is by
}

// because is a reason with no parties.
func because(key string) reason { return reason{key: key, by: -1, at: -1, plague: -1, blast: -1} }

// By names the people the reason is about.
func (y reason) By(c *Civ) reason {
	if c != nil {
		y.by = c.ID
	}
	return y
}

// At names the star.
func (y reason) At(star int) reason { y.at = star; return y }

// Plague names the plague.
func (y reason) Plague(id int) reason { y.plague = id; return y }

// Blast names the blast, or what stands for one.
func (y reason) Blast(id int) reason { y.blast = id; return y }

// Of takes another reason's parties: what a blast did is said with the
// blast's own ids.
func (y reason) Of(o reason) reason {
	if o.by >= 0 {
		y.by = o.by
	}
	if o.at >= 0 {
		y.at = o.at
	}
	if o.plague >= 0 {
		y.plague = o.plague
	}
	if o.blast >= 0 {
		y.blast = o.blast
	}
	return y
}

// none says whether the reason is unset.
func (y reason) none() bool { return y.key == "" }

// params is the reason as a fact's parameters, under the name given.
func (y reason) params(name string) P {
	p := P{name: y.key}
	if y.by >= 0 {
		p["by"] = y.by
	}
	if y.at >= 0 {
		p["at"] = y.at
	}
	if y.plague >= 0 {
		p["plague"] = y.plague
	}
	if y.blast >= 0 {
		p["blast"] = y.blast
	}
	return p
}
