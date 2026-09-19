package history

import (
	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/names"
	"worldgen/internal/plague"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// Plagues: sickness as an actor with a name. A plague is born in a people
// at a rate its dirt raises and its medicine divides, carried along the
// same trade, occupation, landing and message graph the rest of the sim
// runs on, fought each tick by whoever has it in a contest of its level
// and its ladder against the contagion, paid for in worlds by the
// lethality, and remembered: in the dead cities of a world it emptied and
// on the walls of what a sick people left. The arithmetic is the plague
// package; this file is the world's side: who bears, who catches, who is
// rid of it, and what it writes.

// Plague is one sickness in the world: its shape, and its history here.
type Plague struct {
	plague.Plague
	ID        int
	Born      Year
	FirstHost int    // the people it was born in
	Cause     string // what the birth reads as, for the batch: clean, dirt, siege, dark age, relic
	Hosts     int    // peoples that have it now
	Peak      int    // the most at once
	Caught    int    // peoples that have had it, in all
	Worlds    int    // worlds lost to it
	Peoples   int    // peoples ended or brought low by it
	Cults     int    // peoples that formed around it
	Cures     int
	Refusals  int  // ears and ports closed for fear of it
	Woken     int  // times a reservoir or a wall gave it again
	LastHost  Year // when it last had a host
	Extinct   bool
	Wildfire  bool // had ten hosts at once, once
}

// Infection is a plague in one people.
type Infection struct {
	Since     Year
	From      int    // the people it came from, -1 for born or woken
	Road      string // how it came; a key of roads
	Contained bool   // this tick's cure roll held it
	Held      int    // ticks contained in a row
	Carrier   bool   // carries it without toll and never fights it: a cult
	sealed    bool   // the containment line was written
}

// Reservoir is a plague waiting in the dead cities of an emptied world.
type Reservoir struct {
	Plague int
	Until  Year
}

// road is one channel a plague travels: its weight, the kind that takes
// it, whether the giver is blamed, and the line's phrase with %s the giver.
type road struct {
	Weight float64
	Kind   plague.Kind
	Crime  bool
	Phrase string
}

var roads = map[string]road{
	"born":       {0, plague.Biological, false, ""},
	"goods":      {1, plague.Biological, true, "with the goods of the %s"},
	"trade":      {0.3, plague.Biological, true, "along the trade with the %s"},
	"occupation": {1, plague.Biological, true, "with the taking of worlds from the %s"},
	"settling":   {0.3, plague.Biological, true, "from the worlds of the %s nearby"},
	"fleet":      {0.5, plague.Biological, true, "with the ships of the %s"},
	"landing":    {0.5, plague.Biological, true, "with the first landing of the %s"},
	"message":    {1, plague.Memetic, false, "in a message from the %s"},
	"signal":     {0.1, plague.Memetic, false, "on the signals of the %s"},
	"reservoir":  {0, plague.Biological, false, ""},
	"walls":      {0, plague.Memetic, false, ""},
	"belief":     {0, plague.Memetic, false, ""},
}

// catchMul is what a trait does to catching: the short-lived and the
// nomadic catch at one and a half, a people that can sleep through it
// at less.
var catchMul = map[string]float64{"shortlived": 1.5, "nomadic": 1.5, "dormancy": 0.7}

// tickPlagues is the phase, after messages and before the peoples act:
// births, the cure rolls and tolls, then the spread on the channels that
// stand every tick (trade and signal; the rest hook where the contact
// happens), suspicion, and the books.
func (w *World) tickPlagues() {
	for _, c := range w.Civs {
		if c.Active() {
			w.bearPlague(c)
		}
	}
	for _, c := range w.Civs {
		if c.Active() {
			w.fightPlagues(c)
		}
	}
	for _, c := range w.Civs {
		if c.Active() {
			w.contagion(c)
		}
	}
	for _, c := range w.Civs {
		if c.Active() {
			w.suspicion(c)
		}
	}
	w.plagueBooks()
}

// bears says whether a people can bear or catch the kind at all: its
// nature, and the top of the ladder working.
func (w *World) bears(c *Civ, k plague.Kind) bool {
	p := c.Species.Profile()
	if k == plague.Biological && (!p.Can(species.Sickens) || c.miracle("directed_evolution")) {
		return false
	}
	if k == plague.Memetic && (!p.Can(species.Believes) || c.miracle("chorus")) {
		return false
	}
	for _, n := range tech.Ladders[ladderOf(k)] {
		if n.Immune && c.Known[n.Key] && c.working(n.Key) {
			return false
		}
	}
	return true
}

func ladderOf(k plague.Kind) string {
	if k == plague.Memetic {
		return tech.Mind
	}
	return tech.Bio
}

// natureMul is the profile's multiplier on bearing and catching the kind.
func natureMul(c *Civ, k plague.Kind) float64 {
	p := c.Species.Profile()
	if k == plague.Memetic {
		return p.PlagueMeme
	}
	return p.PlagueBio
}

// rungs counts the working rungs of a ladder, and what they add to the
// cure roll.
func (w *World) rungs(c *Civ, k plague.Kind) (n int, cure float64) {
	t := &w.Cfg.Tuning.Plague
	for _, nd := range tech.Ladders[ladderOf(k)] {
		if nd.Immune || !c.Known[nd.Key] || !c.working(nd.Key) {
			continue
		}
		n++
		if nd.Cure > 0 {
			cure += nd.Cure
		} else {
			cure += t.RungCure
		}
	}
	return
}

// halvings counts the working nodes that halve biological births only.
func (w *World) halvings(c *Civ) int {
	n := 0
	for _, k := range knownOf(c) {
		if tech.Get(k).Clean && c.working(k) {
			n++
		}
	}
	return n
}

// sieged counts the worlds with an enemy campaign fleet in the sky.
func (w *World) siegedWorlds(c *Civ) int {
	n := 0
	for _, s := range c.Systems {
		if w.sieged(c, s) {
			n++
		}
	}
	return n
}

// factors is a people's sanitation this tick as birth reads it.
func (w *World) factors(c *Civ, k plague.Kind) plague.Factors {
	t := &w.Cfg.Tuning.Plague
	n, _ := w.rungs(c, k)
	f := plague.Factors{
		Worlds: len(c.Systems), Rungs: n, Halvings: w.halvings(c), Immune: !w.bears(c, k),
		Sieged: w.siegedWorlds(c), Shed: len(c.Shed) > 0,
		Dark:  c.DarkAges > 0 && float64(w.Now-c.LastDark) <= t.DarkYears,
		Taken: c.LastTaken > 0 && float64(w.Now-c.LastTaken) <= t.TakenYears,
		Faith: c.Faced["faith"], Beacon: c.Faced["beacon"], Mul: natureMul(c, k),
	}
	return f
}

// causeOf reads the dirt a birth came of, for the batch.
func causeOf(f plague.Factors) string {
	switch {
	case f.Sieged > 0:
		return "siege"
	case f.Dark:
		return "dark age"
	case f.Shed || f.Taken:
		return "dirt"
	}
	return "clean"
}

// bearPlague is a people's chance each tick to bear a new plague of each
// kind: nothing while it has one.
func (w *World) bearPlague(c *Civ) {
	if len(c.Infections) > 0 {
		return
	}
	for _, k := range []plague.Kind{plague.Biological, plague.Memetic} {
		if k == plague.Memetic && !c.Known["writing"] {
			continue // an idea needs a medium
		}
		f := w.factors(c, k)
		if !w.chance(plague.BirthChance(k, f, &w.Cfg.Tuning.Plague)) {
			continue
		}
		p := w.newPlague(k, c, causeOf(f))
		w.infect(c, p, nil, "born")
		return
	}
}

// newPlague draws one, named for its first host one time in three.
func (w *World) newPlague(k plague.Kind, host *Civ, cause string) *Plague {
	name := host.Name
	if w.R.IntN(2) == 0 {
		name = host.HomeName
	}
	p := &Plague{Plague: plague.New(w.R, k, name, &w.Cfg.Tuning.Plague), ID: len(w.Plagues), Born: w.Now, FirstHost: host.ID, Cause: cause}
	w.Plagues = append(w.Plagues, p)
	return p
}

// infect puts a plague in a people: the fact, the line by its road, the
// blame where the road carries it, and the wildfire when it is the tenth
// host at once.
func (w *World) infect(c *Civ, p *Plague, from *Civ, roadKey string) {
	inf := &Infection{Since: w.Now, From: -1, Road: roadKey}
	if from != nil {
		inf.From = from.ID
	}
	c.Infections[p.ID] = inf
	p.Hosts++
	p.Caught++
	p.Peak = max(p.Peak, p.Hosts)
	p.LastHost = w.Now
	p.Extinct = false
	c.Tally.Sickened++
	if c.FirstPlague == 0 {
		c.FirstPlague = w.Now
	}
	w.factOf(FPlague, c, from, c.Home, p.Name)
	rd := roads[roadKey]
	switch {
	case roadKey == "born" && p.Kind == plague.Memetic:
		w.log("Something moves through the minds of the %s. They call it %s.", c.Name, p.Name)
	case roadKey == "born":
		w.log("Something moves through the worlds of the %s. They call it %s.", c.Name, p.Name)
	case from != nil && rd.Phrase != "":
		w.log("%s comes to the %s %s.", upper(p.Name), c.Name, sprintf(rd.Phrase, from.Name))
	}
	if from != nil && rd.Crime {
		w.factOf(FPlagueGiven, from, c, c.Home, p.Name)
	}
	if p.Hosts >= w.Cfg.Tuning.Plague.Wildfire && !p.Wildfire {
		p.Wildfire = true
		w.factOf(FWildfire, c, nil, c.Home, p.Name)
		w.log("%s is everywhere now.", upper(p.Name))
	}
}

func upper(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-'a'+'A') + s[1:]
}

// plagued says whether a people has a raging plague its neighbours would
// read as weakness.
func (w *World) plagued(c *Civ) bool {
	for _, pid := range sortedInts(c.Infections) {
		inf := c.Infections[pid]
		if !inf.Carrier && !inf.Contained && w.Plagues[pid].Lethality >= w.Cfg.Tuning.Plague.Weakened {
			return true
		}
	}
	return false
}

// dirt is what the cure contest adds to the difficulty: a siege, a shed
// use, a recent dark age.
func (w *World) dirt(c *Civ) float64 {
	t := &w.Cfg.Tuning.Plague
	d := 0.0
	if w.siegedWorlds(c) > 0 {
		d += t.Dirt
	}
	if len(c.Shed) > 0 {
		d += t.Dirt
	}
	if c.DarkAges > 0 && float64(w.Now-c.LastDark) <= t.DarkYears {
		d += t.Dirt
	}
	return d
}

// partnerCured says whether a fathomed trade partner has been rid of the
// plague: the cure travels with the goods.
func (w *World) partnerCured(c *Civ, pid int) bool {
	for _, eid := range sortedInts(c.Trade) {
		if c.Fathomed[eid] && w.Civs[eid].Immune[pid] {
			return true
		}
	}
	return false
}

// fightPlagues is a people's tick against what it has: one cure roll per
// plague, then the toll of each it did not shake.
func (w *World) fightPlagues(c *Civ) {
	t := &w.Cfg.Tuning.Plague
	for _, pid := range sortedInts(c.Infections) {
		if !c.Active() {
			return
		}
		inf := c.Infections[pid]
		p := w.Plagues[pid]
		if inf.Carrier || inf.Since == w.Now {
			continue // a carrier never fights it; what came this tick acts next
		}
		level := c.Sur
		if p.Kind == plague.Memetic {
			level = c.Soc
		}
		_, ladder := w.rungs(c, p.Kind)
		cured := 0.0
		if p.Kind == plague.Biological && w.partnerCured(c, pid) {
			cured = t.PartnerCured
		}
		margin := plague.CureMargin(level, ladder, w.R.NormFloat64()*t.Spread, p.Contagion, w.dirt(c), cured, t)
		switch plague.Band(margin, t) {
		case plague.Cured:
			w.cure(c, p)
			continue
		case plague.Contained:
			inf.Contained = true
			inf.Held++
			c.Tally.Contained++
			if !inf.sealed {
				inf.sealed = true
				w.log("The %s seal every door against %s.", c.Name, p.Name)
			}
			if inf.Held == t.CreedTicks && !c.Scars[ScarQuarantine] {
				c.Scars[ScarQuarantine] = true
				w.log("Twenty thousand years behind sealed doors, and the %s no longer remember how to open them. It is a creed now.", c.Name)
			}
		default:
			inf.Contained, inf.Held = false, 0
		}
		w.toll(c, p, inf)
	}
}

// cure is a people rid of a plague: immune for good.
func (w *World) cure(c *Civ, p *Plague) {
	delete(c.Infections, p.ID)
	c.Immune[p.ID] = true
	p.Hosts--
	p.Cures++
	c.Tally.Cured++
	w.factOf(FCured, c, nil, c.Home, p.Name)
	w.log("The %s are rid of %s.", c.Name, p.Name)
}

// toll is what a plague takes each tick: morale, and each held world with
// a chance by the lethality. The levels, the research and the income read
// the infection where they are derived.
func (w *World) toll(c *Civ, p *Plague, inf *Infection) {
	t := &w.Cfg.Tuning.Plague
	tl := plague.TollOf(p.Plague, inf.Contained, c.Species.Profile().OrganicAsEnergy)
	c.Morale -= tl.Morale * w.dt
	pc := plague.TollChance(p.Lethality, c.Species.Profile().Frail, inf.Contained, t)
	for _, s := range append([]int(nil), c.Systems...) {
		if !c.Active() || !contains(c.Systems, s) {
			return
		}
		if w.chance(pc) {
			w.worldLost(c, p, s)
		}
	}
}

// worldLost is a held world going to the plague: dark and quarantined
// for a sickness of the body, gone over for one of the mind.
func (w *World) worldLost(c *Civ, p *Plague, s int) {
	t := &w.Cfg.Tuning.Plague
	p.Worlds++
	c.Tally.WorldsSick++
	home := s == c.Home
	if p.Kind == plague.Biological {
		w.Reservoir[s] = &Reservoir{Plague: p.ID, Until: w.Now + Year(p.Contagion*t.ReservoirMyr*1e6)}
		w.factOf(FPlagueWorld, c, nil, s, p.Name)
		switch {
		case home && len(c.Systems) == 1:
			p.Peoples++
			w.endCiv(c, Extinct, "sickened and died of "+p.Name)
		case home:
			p.Peoples++
			w.loseSystem(c, s, "quarantined dead cities", "")
			w.contract(c, "were hollowed out by "+p.Name)
		default:
			w.loseSystem(c, s, "quarantined dead cities", "")
			w.log("%s empties %s, a %s of the %s. The cities are sealed and left.", upper(p.Name), w.star(s), c.Species.Flavour().Colony, c.Name)
		}
		return
	}
	cult := w.R.Float64() < t.CultChance
	f := w.factOf(FBelieved, c, nil, s, p.Name)
	if home {
		p.Peoples++
		w.endCiv(c, Transformed, "listened to "+p.Name+" and were changed by it")
		c.Into = "something that believed"
		if cult {
			nc := w.cult(c, p, s)
			c.Into = "the " + nc.Name
			f.Object = nc.ID
		}
		return
	}
	w.loseSystem(c, s, "world that believes", "")
	if cult {
		nc := w.cult(c, p, s)
		f.Object = nc.ID
		return
	}
	w.log("%s takes %s, a %s of the %s. Nobody there answers to them any more.", upper(p.Name), w.star(s), c.Species.Flavour().Colony, c.Name)
}

// cult is a world gone over declaring itself a people carrying the idea:
// the same blood under a new name, immune to it and a carrier of it, as
// schism's branch is made.
func (w *World) cult(c *Civ, p *Plague, s int) *Civ {
	nc := w.spawnCiv(s, c.Species, -1, names.Civ(w.R))
	nc.Origin = "the believers in " + p.Name + ", once of the " + c.Name
	nc.Master = -1
	for k := range c.Known {
		nc.Known[k] = true
	}
	w.forget(nc, 0.2)
	w.recompute(nc)
	w.branchMorality(nc, c)
	w.inherit(nc, c, 0)
	nc.Immune[p.ID] = true
	nc.Infections[p.ID] = &Infection{Since: w.Now, From: c.ID, Road: "belief", Carrier: true}
	p.Hosts++
	p.Cults++
	c.Tally.Cults++
	w.log("At %s the believers in %s declare themselves a people: the %s.", w.star(s), p.Name, nc.Name)
	return nc
}

// contagion is the spread on the channels that stand every tick: a
// sickness of the body along every trade link, at full weight where goods
// moved, and one of the mind on the signals of everyone in earshot.
func (w *World) contagion(c *Civ) {
	for _, pid := range sortedInts(c.Infections) {
		inf := c.Infections[pid]
		if inf.Contained {
			continue
		}
		p := w.Plagues[pid]
		if p.Kind == plague.Biological {
			for _, eid := range sortedInts(c.Trade) {
				e := w.Civs[eid]
				if !e.Active() {
					continue
				}
				rd := "trade"
				if c.From[eid].Total()+e.From[c.ID].Total() > 0 {
					rd = "goods"
				}
				w.offer(c, e, p, rd)
			}
			continue
		}
		for _, eid := range sortedInts(c.Met) {
			e := w.Civs[eid]
			if e.Active() && w.hear(c, e) {
				w.offer(c, e, p, "signal")
			}
		}
	}
}

// expose is a channel event between two peoples: every plague of the
// road's kind the first has and is not containing is offered to the
// second. Certain makes the crossing sure for a contagious plague, as the
// taking of a world is.
func (w *World) expose(a, b *Civ, roadKey string) {
	for _, pid := range sortedInts(a.Infections) {
		p := w.Plagues[pid]
		if p.Kind == roads[roadKey].Kind {
			w.offer(a, b, p, roadKey)
		}
	}
}

// offer is one plague on one channel to one people: the chance is the
// contagion by the road's weight, divided by the taker's hygiene,
// multiplied by its nature. A contained host offers nothing; an immune
// people cannot catch it.
func (w *World) offer(a, b *Civ, p *Plague, roadKey string) bool {
	if a == b || !b.Active() || b.Infections[p.ID] != nil || b.Immune[p.ID] {
		return false
	}
	if inf := a.Infections[p.ID]; inf == nil || inf.Contained {
		return false
	}
	if !w.bears(b, p.Kind) {
		return false
	}
	rd := roads[roadKey]
	weight := rd.Weight
	if roadKey == "occupation" && p.Contagion > 0.5 {
		weight = 1e9 // certain
	}
	mul := natureMul(b, p.Kind)
	for _, tr := range b.Species.Traits {
		if m, ok := catchMul[tr.Key]; ok {
			mul *= m
		}
	}
	n, _ := w.rungs(b, p.Kind)
	pc := plague.CatchChance(p.Contagion, weight, plague.Hygiene(n, &w.Cfg.Tuning.Plague), mul)
	if !w.chance(pc) {
		return false
	}
	w.infect(b, p, a, roadKey)
	return true
}

// shutTo says whether a people drops a message from a sender it has
// closed its ears to.
func (w *World) shutTo(to, from *Civ) bool {
	if !to.Closed[from.ID] {
		return false
	}
	to.Tally.Shut++
	return true
}

// suspicion is what a people thinks each tick of who is sick: those a
// tale it holds says a plague struck and no later tale says were rid of
// it, those its intel within ten thousand years read as sick, and every
// partner of a sick partner. A newly
// suspected sender is closed out with one roll, ears and ports both,
// while the suspicion lasts.
func (w *World) suspicion(c *Civ) {
	sick := map[int]string{}
	cured := map[int]Year{} // the newest cure known of each people
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		if f := w.Facts[t.Fact]; f.Kind == FCured && f.Year > cured[f.Subject] {
			cured[f.Subject] = f.Year
		}
	}
	for _, t := range c.Lore {
		if t.Forgot {
			continue
		}
		f := w.Facts[t.Fact]
		if f.Kind == FPlague && f.Subject != c.ID && cured[f.Subject] < f.Year {
			sick[f.Subject] = f.What
		}
	}
	for _, eid := range sortedInts(c.Intel) {
		if i := c.Intel[eid]; i.Sick != "" && float64(w.Now-i.Year) <= w.Cfg.Tuning.Plague.TakenYears {
			sick[eid] = i.Sick
		}
	}
	suspects := map[int]string{}
	for eid, name := range sick {
		suspects[eid] = name
	}
	for _, pid := range sortedInts(c.Trade) {
		if name, ok := sick[pid]; ok {
			for _, qid := range sortedInts(w.Civs[pid].Trade) {
				if qid != c.ID {
					if _, ok := suspects[qid]; !ok {
						suspects[qid] = name
					}
				}
			}
		}
	}
	for _, eid := range sortedInts(suspects) {
		e := w.Civs[eid]
		if c.Suspect[eid] || !e.Living() || !c.Met[eid] {
			continue // nothing to close against a people never heard from
		}
		if p := w.plagueNamed(suspects[eid]); p != nil && (c.Infections[p.ID] != nil || c.Immune[p.ID]) {
			continue // nothing to fear from what one has, or has had
		}
		c.Suspect[eid] = true
		r := mind.Refuse(mind.RefuseInput{Fear: c.Dials.Fear, Creed: c.Scars[ScarQuarantine], Cautious: c.Has("cautious"), Censor: c.Known["censorship"] && c.working("censorship")}, w.Cfg.Tuning)
		w.explain(c, "suspecting the "+e.Name, r)
		if !w.chance(r.Chance) {
			continue
		}
		c.Closed[eid] = true
		c.Tally.Refusals++
		w.factOf(FRefused, c, e, -1, suspects[eid])
		if p := w.plagueNamed(suspects[eid]); p != nil {
			p.Refusals++
		}
		if c.Trade[eid] {
			w.log("The %s stop the trade with the %s for fear of %s, and hear nothing from them.", c.Name, e.Name, suspects[eid])
		} else {
			w.log("The %s close their ears to the %s for fear of %s.", c.Name, e.Name, suspects[eid])
		}
	}
	for _, eid := range sortedInts(c.Suspect) {
		if _, ok := suspects[eid]; !ok {
			delete(c.Suspect, eid)
			delete(c.Closed, eid)
		}
	}
}

// plagueNamed finds a plague by its name, for the books.
func (w *World) plagueNamed(name string) *Plague {
	for _, p := range w.Plagues {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// plagueBooks keeps the counts: hosts standing, extinction after a
// hundred thousand years with none, reservoirs run dry.
func (w *World) plagueBooks() {
	t := &w.Cfg.Tuning.Plague
	for _, p := range w.Plagues {
		p.Hosts = 0
	}
	for _, c := range w.Civs {
		if !c.Active() {
			continue
		}
		for _, pid := range sortedInts(c.Infections) {
			w.Plagues[pid].Hosts++
		}
	}
	for _, p := range w.Plagues {
		if p.Hosts > 0 {
			p.LastHost = w.Now
		} else if !p.Extinct && float64(w.Now-p.LastHost) > t.ExtinctYears && !w.reservoired(p.ID) {
			p.Extinct = true
		}
	}
	for _, s := range sortedInts(w.Reservoir) {
		if w.Reservoir[s].Until <= w.Now {
			delete(w.Reservoir, s)
		}
	}
}

// reservoired says whether a plague waits somewhere.
func (w *World) reservoired(pid int) bool {
	for _, s := range sortedInts(w.Reservoir) {
		if w.Reservoir[s].Plague == pid {
			return true
		}
	}
	return false
}

// wakeReservoir is a people settling or taking a star with dead cities: it
// catches what waits there with the plague's contagion.
func (w *World) wakeReservoir(c *Civ, s int) {
	r := w.Reservoir[s]
	if r == nil || !c.Active() {
		return
	}
	p := w.Plagues[r.Plague]
	if c.Infections[p.ID] != nil || c.Immune[p.ID] || !w.bears(c, p.Kind) {
		return
	}
	if !w.chance(plague.CatchChance(p.Contagion, 1, 1, natureMul(c, p.Kind))) {
		return
	}
	p.Woken++
	w.infect(c, p, nil, "reservoir")
	w.log("The %s come to %s and wake %s in its dead cities.", c.Name, w.star(s), p.Name)
}

// wallsWritten marks a remain with the sickness of the mind its makers
// had when they wrote on it.
func (w *World) wallsWritten(c *Civ, l *Legacy) {
	for _, pid := range sortedInts(c.Infections) {
		if w.Plagues[pid].Kind == plague.Memetic {
			l.Plague = pid
			return
		}
	}
}

// readWallsPlague is a people reading walls that carry an idea: it catches
// it at half the contagion, for ever.
func (w *World) readWallsPlague(c *Civ, l *Legacy) {
	if l.Plague < 0 || !c.Active() {
		return
	}
	p := w.Plagues[l.Plague]
	if c.Infections[p.ID] != nil || c.Immune[p.ID] || !w.bears(c, p.Kind) {
		return
	}
	if !w.chance(plague.CatchChance(p.Contagion, w.Cfg.Tuning.Plague.WallsShare, 1, natureMul(c, p.Kind))) {
		return
	}
	p.Woken++
	w.infect(c, p, nil, "walls")
	maker := "someone"
	if l.Maker >= 0 {
		maker = "the " + w.Civs[l.Maker].Name
	}
	w.log("On the walls of %s %s is written, and the %s read it.", maker, p.Name, c.Name)
}

// wakeRelic is a relic that does what it was made to do: a plague of its
// domain, born in the finder.
func (w *World) wakeRelic(c *Civ, l *Legacy) {
	k := plague.Biological
	if n := l.node(); n != nil && n.Domain == tech.Society {
		k = plague.Memetic
	}
	if !w.bears(c, k) {
		w.log("It does what it was made to do, and finds nothing in the %s to do it to.", c.Name)
		return
	}
	p := w.newPlague(k, c, "relic")
	w.infect(c, p, nil, "born")
	w.log("It does what it was made to do, to the %s. They call it %s.", c.Name, p.Name)
}

// inheritImmunity gives a people born of another what the parent could
// not catch again.
func (w *World) inheritImmunity(nc, parent *Civ) {
	for _, pid := range sortedInts(parent.Immune) {
		nc.Immune[pid] = true
	}
}

// sickLevels is what the plagues take off a people's levels this tick.
func (w *World) sickLevels(c *Civ) (sur, soc float64) {
	for _, pid := range sortedInts(c.Infections) {
		inf := c.Infections[pid]
		if inf.Carrier {
			continue
		}
		tl := plague.TollOf(w.Plagues[pid].Plague, inf.Contained, false)
		sur, soc = sur+tl.Sur, soc+tl.Soc
	}
	return
}

// sickRate is the plagues' multiplier on research, and the focus they
// bend it with: biology against a sickness of the body, society against
// one of the mind.
func (w *World) sickRate(c *Civ) (mul float64, focus map[string]float64) {
	mul = 1
	for _, pid := range sortedInts(c.Infections) {
		inf := c.Infections[pid]
		if inf.Carrier {
			continue
		}
		p := w.Plagues[pid]
		mul *= plague.TollOf(p.Plague, inf.Contained, false).Research
		if inf.Contained {
			continue
		}
		if focus == nil {
			focus = map[string]float64{}
		}
		if p.Kind == plague.Memetic {
			focus[tech.Society] = 1.5
		} else {
			focus[tech.Biology] = 1.5
		}
	}
	return
}

// sickIncome is the plagues' cut on income: organic matter for a sickness
// of the body, energy for one of the mind in a machine-born people.
func (w *World) sickIncome(c *Civ, in flow.Income) flow.Income {
	for _, pid := range sortedInts(c.Infections) {
		inf := c.Infections[pid]
		if inf.Carrier {
			continue
		}
		tl := plague.TollOf(w.Plagues[pid].Plague, inf.Contained, c.Species.Profile().OrganicAsEnergy)
		in[flow.O] *= tl.Organic
		in[flow.E] *= tl.Energy
	}
	return in
}

// SickWord is the portrait's line: what a people has, and for how long.
func (w *World) SickWord(c *Civ) string {
	var parts []string
	for _, pid := range sortedInts(c.Infections) {
		inf := c.Infections[pid]
		p := w.Plagues[pid]
		state := "raging"
		switch {
		case inf.Carrier:
			state = "carried"
		case inf.Contained:
			state = "contained"
		}
		parts = append(parts, sprintf("sick with %s these %s, %s", p.Name, span(w.Now-inf.Since), state))
	}
	return list(parts)
}

// Infected lists a people's plagues, for the readers.
func (w *World) Infected(c *Civ) []*Plague {
	var out []*Plague
	for _, pid := range sortedInts(c.Infections) {
		out = append(out, w.Plagues[pid])
	}
	return out
}

// sickSeen is what a look at a people reads of its health: the name of a
// plague raging in it, or nothing.
func (w *World) sickSeen(e *Civ) string {
	for _, pid := range sortedInts(e.Infections) {
		if inf := e.Infections[pid]; !inf.Carrier && !inf.Contained {
			return w.Plagues[pid].Name
		}
	}
	return ""
}

// settledNear is a colony placed within a hop of a sick people's world:
// the settler is exposed once to each such neighbour.
func (w *World) settledNear(c *Civ, t int) {
	hop := w.Cfg.Tuning.Expand.Hop
	for _, a := range w.Civs {
		if a == c || !a.Active() || len(a.Infections) == 0 {
			continue
		}
		for _, s := range a.Systems {
			if w.G.Dist(s, t) <= hop {
				w.expose(a, c, "settling")
				break
			}
		}
	}
}

// landed is a fleet of one people standing at a world of another: each
// is exposed to what the other carries in bodies.
func (w *World) landed(c, h *Civ) {
	w.expose(c, h, "fleet")
	w.expose(h, c, "fleet")
}
