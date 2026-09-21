package history

import (
	"slices"
	"sort"
	"strings"
)

// Tellings. The chronicle is what happened. What a people knows of it is a
// telling: a set of tales, each a fact seen from one side, learned by
// witness, by word from another people, by reading a ruin or a relic, or
// by inheritance. Tales wear: a year is lost, then a name, then the thing
// is a myth, then it is gone. Machines and planetary minds wear slowly.
// A people slants what it tells: its own deeds grow, its enemies' deeds
// shrink and their crimes swell, and a friend that turns is retold. What a
// people leaves in a relic is its telling at that time.

// FactKind is what happened.
type FactKind uint8

const (
	FArise FactKind = iota
	FStars
	FSettle
	FZenith
	FDarkAge
	FFall
	FEnd
	FWar
	FTaken
	FBurned
	FHomeBroken
	FScoured
	FYield
	FPeace
	FEnslaved
	FVassal
	FFreed
	FCrushed
	FMet
	FTrade
	FPact
	FBetrayal
	FRelief
	FDefeat
	FFind
	FMastered
	FSealed
	FUnleashed
	FOvercome
	FScarred
	FDeclined
	FMiracle
	FUplift
	FBred
	FCosmic
	FDoom
	FExodus
	FRest
	FStripped
	FCycle
	FSurveyLost
	FWant         // the lean years: a stretch of shedding past a hundred thousand years
	FHarness      // a first source of a kind put to use: mines in a belt, a tap on a dead star
	FEmbargo      // a people closed its ports to a partner in want
	FCutOff       // a people's uses went dark when a partner stopped sending
	FManna        // a people eats something that thinks
	FRise         // what was grown for the table rose as a people
	FLoose        // what was grown for the table got out
	FIntercept    // a fleet met in the dark and beaten: the winner's
	FCaught       // the loser's
	FFathomed     // a people came to understand another
	FBrokered     // a people spoke for another to a third
	FHire         // a people took another's pay to hold a star; see contract.go
	FTaught       // a people taught another a node, for pay
	FStrikeBought // a people paid a third to send a fleet against another
	FBoughtOff    // a hired people sold what it was paid to hold
	FTribute      // a people paid tribute after a war
	FSlight       // a people made war on the partner of another; see slight.go
	FPlague       // a people caught a plague; see plague.go
	FPlagueGiven  // and it came from another, by goods, occupation or a fleet
	FPlagueWorld  // a world emptied by a plague
	FCured        // a people rid of one
	FRefused      // a people closed its ears and its ports to another for fear of one
	FBelieved     // a world went over to an idea; the object is the cult, if one formed
	FWildfire     // a plague in ten peoples at once
	FPoisoned     // a people put a plague in another by stealth, or was caught trying; see weapon.go
	FWoke         // a plague became a people: the subject is the rider, the object its first host; see parasite.go
	FRenaissance  // a people grew young again; see ossify.go
	FSundered     // a people tore itself into heirs: the subject is the old people, the object one heir; see sunder.go
	FReclaimed    // an heir took a world of the old realm from whoever held it
	FShattered    // a people forgot the stars and became one people per world: the subject the old people, the object one shard
	FSevered      // a world of a people with no factions was cut from its seat as a people of its own: the subject the old, the object the new; see sunder.go
	FDeepened     // an eldritch people drew another power; What is the power's name; see eldritch.go
	FAppeared     // another of an eldritch people is simply there, at the star
	FTithed       // an eldritch people takes a share of another's harvest: the subject the taker, the object the tithed
	FDemand       // a living world told a people to leave a world of its neighbourhood: the subject the world, the object the told; What "left" or "refused"; see waking.go
	FWaking       // a living world woke on a people's worlds: the subject the world, the object the people
	FUnmade       // a people ended a world by the unmaking: the subject the unmaker, the object the holder
)

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

// factShape is the sort and weight of each kind. Weight is how much a
// tale resists wear and how far news of it travels.
var factShape = [...]struct {
	Sort   Sort
	Weight float64
}{
	FArise: {Deed, 3}, FStars: {Deed, 2}, FSettle: {Deed, 1}, FZenith: {Deed, 2},
	FDarkAge: {Woe, 3}, FFall: {Woe, 4}, FEnd: {Woe, 5},
	FWar: {Crime, 2}, FTaken: {Crime, 2}, FBurned: {Crime, 3}, FHomeBroken: {Crime, 5}, FScoured: {Crime, 5},
	FYield: {Deed, 3}, FPeace: {Bond, 1}, FEnslaved: {Crime, 4}, FVassal: {Deed, 3}, FFreed: {Deed, 4}, FCrushed: {Crime, 3},
	FMet: {Bond, 1}, FTrade: {Bond, 1}, FPact: {Bond, 2}, FBetrayal: {Crime, 3}, FRelief: {Deed, 2}, FDefeat: {Woe, 2},
	FFind: {Deed, 2}, FMastered: {Deed, 3}, FSealed: {Deed, 2}, FUnleashed: {Folly, 4},
	FOvercome: {Deed, 2}, FScarred: {Woe, 2}, FDeclined: {Woe, 3},
	FMiracle: {Deed, 3}, FUplift: {Deed, 3}, FBred: {Crime, 4},
	FCosmic: {Woe, 3}, FDoom: {Woe, 3}, FExodus: {Woe, 3}, FRest: {Deed, 2}, FStripped: {Crime, 3}, FCycle: {Deed, 3}, FSurveyLost: {Woe, 2},
	FWant: {Woe, 1}, FHarness: {Deed, 1},
	FEmbargo: {Crime, 1}, FCutOff: {Woe, 2}, FManna: {Crime, 2}, FRise: {Deed, 3}, FLoose: {Folly, 4},
	FIntercept: {Deed, 1}, FCaught: {Woe, 1},
	FFathomed: {Bond, 1}, FBrokered: {Deed, 1},
	FHire: {Deed, 1}, FTaught: {Deed, 1}, FStrikeBought: {Crime, 1}, FBoughtOff: {Crime, 2}, FTribute: {Woe, 1},
	FSlight: {Crime, 1},
	FPlague: {Woe, 4}, FPlagueGiven: {Crime, 2}, FPlagueWorld: {Woe, 3}, FCured: {Deed, 2}, FRefused: {Crime, 1}, FBelieved: {Woe, 3}, FWildfire: {Woe, 3},
	FPoisoned: {Crime, 4}, FWoke: {Deed, 3},
	FRenaissance: {Deed, 3}, FSundered: {Woe, 4}, FReclaimed: {Deed, 2}, FShattered: {Woe, 4},
	FSevered: {Woe, 2}, FDeepened: {Deed, 2}, FAppeared: {Deed, 1}, FTithed: {Crime, 2}, FDemand: {Crime, 1}, FWaking: {Crime, 4}, FUnmade: {Crime, 5},
}

// Fact is one thing that happened, as it happened.
type Fact struct {
	ID      int
	Kind    FactKind
	Year    Year
	Subject int    // the people it is about
	Object  int    // the other people, or -1
	Star    int    // where, or -1
	Legacy  int    // the remain in it, or -1
	N       int    // a count: worlds
	What    string // a cause, a filter's name, a miracle, a shape of betrayal
}

func (f *Fact) sort() Sort      { return factShape[f.Kind].Sort }
func (f *Fact) weight() float64 { return factShape[f.Kind].Weight }

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

// Inscription is a tale as it stood when it was written down, with its text.
type Inscription struct {
	Tale Tale
	Text string
}

// fact records something that happened and lets the parties know it.
func (w *World) fact(k FactKind, c, e *Civ, star int) *Fact {
	f := &Fact{ID: len(w.Facts), Kind: k, Year: w.Now, Subject: c.ID, Object: -1, Star: star, Legacy: -1}
	if e != nil {
		f.Object = e.ID
	}
	w.Facts = append(w.Facts, f)
	if star >= 0 {
		w.factsAt[star] = append(w.factsAt[star], f.ID)
	}
	w.witness(c, f)
	if e != nil {
		w.witness(e, f)
	}
	w.spread(f)
	return f
}

// factAt is a fact at a year inside the tick.
func (w *World) factAt(y Year, k FactKind, c, e *Civ, star int) *Fact {
	f := w.fact(k, c, e, star)
	f.Year = y
	return f
}

// factN is a fact with a count.
func (w *World) factN(k FactKind, c, e *Civ, star, n int) *Fact {
	f := w.fact(k, c, e, star)
	f.N = n
	return f
}

// factOf is a fact with a word in it: a cause, a name.
func (w *World) factOf(k FactKind, c, e *Civ, star int, what string) *Fact {
	f := w.fact(k, c, e, star)
	f.What = what
	return f
}

// factL is a fact with a remain in it.
func (w *World) factL(k FactKind, c *Civ, l *Legacy) *Fact {
	var m *Civ
	if l.Maker >= 0 && l.Maker != c.ID {
		m = w.Civs[l.Maker]
	}
	f := w.fact(k, c, m, l.Star)
	f.Legacy = l.ID
	return f
}

// witness is a party to a fact learning it as it happened.
func (w *World) witness(c *Civ, f *Fact) {
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
	if id < 0 {
		return 0
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
func (c *Civ) other(f *Fact) int {
	if f.Subject != c.ID {
		return f.Subject
	}
	return f.Object
}

// learn adds a tale. A tale told by another people carries the teller's
// regard for the other party when the learner has none of its own; that
// is how a stranger comes to be a monster to peoples it never met.
func (w *World) hold(c *Civ, f *Fact, src Provenance, from int, slant int8, wear int8) *Tale {
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
	w.takeToHeart(c, f, t)
	if len(c.Lore) > 600 {
		w.prune(c)
	}
	return t
}

// takeToHeart is what learning a tale does at once: a crime against a
// friend is held against the doer, and a crime by a stranger against
// anyone is a thing to be feared.
func (w *World) takeToHeart(c *Civ, f *Fact, t *Tale) {
	w.judgeLine(c, f, t)
	s, wt := sortFor(c, f)
	if s != Crime || f.Subject == c.ID || t.Source == Witnessed && f.Object == c.ID {
		return
	}
	if f.Object >= 0 && f.Object != c.ID && w.regard(c, f.Object) > 0 {
		c.resent(f.Subject, 0.1*wt/3)
	}
}

// judgeLine is the note, now and then, that a people's judgment of a
// thing it has just learned is not the fact's own: a crime it counts no
// crime, a deed it calls one.
func (w *World) judgeLine(c *Civ, f *Fact, t *Tale) {
	if t.Source == Inherited || f.Subject == c.ID || f.Object == c.ID || !judges(c, f) || w.R.Float64() >= 0.05 {
		return
	}
	s, _ := sortFor(c, f)
	switch {
	case f.sort() == Crime && s == Deed:
		w.log("The %s hear that the %s %s, and count it a deed.", c.Name, w.Civs[f.Subject].Name, w.deedOf(f))
	case f.sort() == Crime && s == Nothing:
		w.log("The %s hear that the %s %s, and count it no crime.", c.Name, w.Civs[f.Subject].Name, w.deedOf(f))
	case f.sort() != Crime && s == Crime:
		w.log("The %s hear what the %s did, and call it a crime.", c.Name, w.Civs[f.Subject].Name)
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

// reckon works out whom a people remembers as monsters, and counts what
// it went through.
func (w *World) reckon(c *Civ) {
	w.experienced(c)
	x := map[int]float64{}
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Facts[t.Fact]
		s, wt := sortFor(c, f)
		if s == Woe {
			if t.Blamed < 0 || t.Blamed == c.ID {
				continue
			}
			v := 0.5
			if f.Subject == c.ID {
				v = 2
			}
			x[t.Blamed] += v * wt / 3 * (1 + 0.5*float64(t.Wear))
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
		who := f.Subject
		if t.Blamed >= 0 {
			who = t.Blamed
		}
		x[who] += v * wt / 3 * (1 + 0.5*float64(t.Wear))
	}
	c.monsters = map[int]bool{}
	for id, v := range x {
		if v >= 3 {
			c.monsters[id] = true
		}
	}
}

// spread carries news: each party tells the peoples it trades with or is
// sworn to, and anyone who can hear its signals learns of the heavier
// things; whoever had a holding within sight of the star saw it happen.
func (w *World) spread(f *Fact) {
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
	f := w.Facts[m.Fact]
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
	var pick []*Fact
	for _, id := range ids {
		f := w.Facts[id]
		if c.knows(id) || f.weight() < 2 || f.sort() == Bond {
			continue
		}
		switch f.Kind {
		case FSettle, FTaken, FBurned, FHomeBroken, FScoured, FEnd, FFall, FUnleashed, FWaking, FUnmade, FArise, FExodus, FCosmic:
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
		fi, fj := w.Facts[keep[i].Fact], w.Facts[keep[j].Fact]
		wi, wj := w.dearness(c, fi, keep[i]), w.dearness(c, fj, keep[j])
		if wi != wj {
			return wi > wj
		}
		return fi.Year > fj.Year
	})
	if len(keep) > 8 {
		keep = keep[:8]
	}
	sort.SliceStable(keep, func(i, j int) bool { return w.Facts[keep[i].Fact].Year < w.Facts[keep[j].Fact].Year })
	for _, t := range keep {
		l.Testament = append(l.Testament, Inscription{Tale: *t, Text: w.tell(c, t)})
	}
	c.Tally.Testaments++
	w.wallsWritten(c, l)
}

// dearness is how much a tale matters to its teller: the fact's weight,
// more for its own part in it, less for strangers' business.
func (w *World) dearness(c *Civ, f *Fact, t *Tale) float64 {
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
		f := w.Facts[td.Tale.Fact]
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
		if own {
			w.log("In what they left at %s the %s read their own story in their own words, and remember.", w.star(l.Star), c.Name)
		} else {
			w.log("What the %s read in %s at %s is the telling of %s, and they have no other.", c.Name, l.Desc, w.star(l.Star), w.makerName(l))
		}
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
		t.Forgot = false
		t.Wear = was.Wear
		t.Blamed = was.Blamed
		t.Source = Inherited
		c.Tally.Restored++
		return true
	}
	f := w.Facts[fact]
	t := w.hold(c, f, Inherited, c.ID, 0, was.Wear)
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
		nt := w.hold(nc, w.Facts[t.Fact], Inherited, parent.ID, t.Slant, min(2, t.Wear+wear))
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
		f := w.Facts[t.Fact]
		if t.Forgot || t.Wear >= 2 || f.weight() < 2 || to.knows(f.ID) {
			continue
		}
		if f.Subject == to.ID || f.Object == to.ID {
			continue
		}
		pick = append(pick, t)
	}
	sort.SliceStable(pick, func(i, j int) bool {
		return w.Facts[pick[i].Fact].weight() > w.Facts[pick[j].Fact].weight()
	})
	for i, t := range pick {
		if i >= 5 {
			break
		}
		to.Tally.Told++
		w.hold(to, w.Facts[t.Fact], Told, from.ID, t.Slant, t.Wear)
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
	m := w.memory(c)
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Facts[t.Fact]
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
		if !w.chance(rate) {
			continue
		}
		w.wearStep(c, t, f)
	}
}

// wearStep takes one tale one step toward myth, and at the myth step may
// hang the blame on whoever is the enemy now.
func (w *World) wearStep(c *Civ, t *Tale, f *Fact) {
	if t.Wear >= 2 {
		t.Forgot = true
		c.Tally.Forgot++
		return
	}
	t.Wear++
	if t.Wear < 2 {
		return
	}
	c.Tally.Myths++
	if f.Kind == FSundered && f.Object == c.ID && c.Claim != nil {
		c.Claim = nil // the sundering is a story now, and a faction is a people
		w.log("Among the %s the sundering has become a story told to children. Nobody speaks of the old realm as theirs any more.", c.Name)
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
				w.log("The %s now tell that it was the %s who %s. It was not.", c.Name, w.Civs[e].Name, w.blameOf(c, f))
			}
			return
		}
	}
	if wt >= 4 && (f.Subject == c.ID || f.Object == c.ID) && w.R.Float64() < 0.15 {
		w.log("Among the %s, %s has become a story told to children.", c.Name, w.mythOf(c, f))
	}
}

// foe is the enemy of the day: the people this one holds the most against,
// alive or dead; failing a grudge, whoever it is at war with, or remembers
// as a monster. -1 when there is nobody.
func (w *World) foe(c *Civ) int {
	best, bg := -1, 0.3
	for _, id := range sortedInts(c.Met) {
		if g := c.Grudge[id]; g > bg {
			best, bg = id, g
		}
	}
	if best >= 0 {
		return best
	}
	for _, id := range sortedInts(c.Wars) {
		if c.Wars[id] {
			return id
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
	var deeds []string
	n := 0
	for _, t := range c.Lore {
		if t.Forgot || t.Wear < 1 || t.Blamed >= 0 {
			continue
		}
		f := w.Facts[t.Fact]
		s, _ := sortFor(c, f)
		if s != Crime && s != Folly && s != Woe {
			continue
		}
		doer := f.Subject
		if s == Woe {
			doer = f.Object
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
		if d := w.blameOf(c, f); len(deeds) < 3 && !slices.Contains(deeds, d) {
			deeds = append(deeds, d)
		}
	}
	if n == 0 {
		return
	}
	line := deeds[0]
	if len(deeds) > 1 {
		line = strings.Join(deeds[:len(deeds)-1], ", ") + ", and " + deeds[len(deeds)-1]
	}
	more := ""
	if n > len(deeds) {
		more = sprintf(", and %d things besides", n-len(deeds))
	}
	w.log("With the %s for an enemy, the %s tell their history over: it was the %s who %s%s. It was not.", w.Civs[e].Name, c.Name, w.Civs[e].Name, line, more)
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
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Facts[t.Fact]
		o := c.other(f)
		if o < 0 {
			continue
		}
		r := w.regard(c, o)
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
			t.Wear++
		}
	}
}

// forgetting is a dark age's toll on memory: most tales take a step, and
// what was already myth is lost.
func (w *World) forgetting(c *Civ) {
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Facts[t.Fact]
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
		fi, fj := w.Facts[c.Lore[i].Fact], w.Facts[c.Lore[j].Fact]
		return w.dearness(c, fi, c.Lore[i]) > w.dearness(c, fj, c.Lore[j])
	})
	for _, t := range c.Lore[500:] {
		t.Forgot = true
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
		f := w.Facts[t.Fact]
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

// loreDials is what a people's telling does to its temperament: what it
// remembers suffering makes it fearful and hating, what it remembers
// winning makes it bold, remembered promises make it loyal and remembered
// betrayals the reverse. Myth counts for more than memory.
func (w *World) loreDials(c *Civ) Dials {
	var d Dials
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Facts[t.Fact]
		s, _ := sortFor(c, f)
		k := 1 + 0.5*float64(t.Wear)
		self := f.Subject == c.ID
		switch {
		case f.Kind == FBetrayal && f.Object == c.ID && s == Crime:
			d.Loyalty -= 0.10 * k
			d.Fear += 0.06 * k
		case s == Crime && f.Object == c.ID:
			d.Hate += 0.06 * k
			d.Fear += 0.04 * k
			d.Patience += 0.04 * k
		case self && (f.Kind == FTaken || f.Kind == FHomeBroken || f.Kind == FYield):
			d.Aggression += 0.06 * k
			d.Greed += 0.04 * k
		case self && s == Folly:
			d.Risk -= 0.10 * k
		case self && s == Woe:
			d.Fear += 0.04 * k
			d.Risk -= 0.04 * k
		case self && (f.Kind == FFind || f.Kind == FMastered || f.Kind == FCycle):
			d.Hunger += 0.06 * k
		case s == Bond && (self || f.Object == c.ID):
			d.Loyalty += 0.04 * k
		}
	}
	d.Aggression = clamp(d.Aggression, -0.3, 0.3)
	d.Risk = clamp(d.Risk, -0.3, 0.3)
	d.Greed = clamp(d.Greed, -0.3, 0.3)
	d.Fear = clamp(d.Fear, -0.3, 0.3)
	d.Loyalty = clamp(d.Loyalty, -0.3, 0.3)
	d.Hunger = clamp(d.Hunger, -0.3, 0.3)
	d.Patience = clamp(d.Patience, -0.3, 0.3)
	d.Hate = clamp(d.Hate, -0.3, 0.3)
	return d
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
