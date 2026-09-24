package history

import (
	"encoding/json"
	"sort"

	"worldgen/internal/galaxy"
	"worldgen/internal/record"
	"worldgen/internal/species"
)

// Export is the run as records: the dossier, the state with every
// people's knowledge, the chronicle and the tellings. The names are
// added by the writer (internal/writer), which runs the names pass; the
// code revision is the writer's too. Nothing here decides anything, and
// the world is not changed: what a record holds is what the state holds,
// as keys and ids.
func (w *World) Export(at string) *record.Run {
	r := &record.Run{}
	r.State = w.exportState()
	r.Chronicle = w.exportChronicle()
	r.Tellings = w.exportTellings(r.State)
	r.Dossier = w.exportDossier(at, r)
	return r
}

func (w *World) exportDossier(at string, r *record.Run) *record.Dossier {
	g := w.G
	d := &record.Dossier{
		Format: record.Format, Code: "unknown", Seed: w.Seed,
		Present: w.Present, Truncated: w.Truncated, Waning: w.Waning, DeepStart: w.Cfg.DeepStart, Dawn: w.Cfg.Dawn, Step: w.Cfg.Step,
		Ticks: w.Ticks, Capped: w.Capped, MaxFades: w.Cfg.MaxFades,
		Config:    record.Config{At: at, Stars: w.Cfg.Stars, Radius: w.Cfg.Radius, Thickness: w.Cfg.Thickness, Tuning: w.Cfg.Tuning},
		Cycle:     record.Cycle{Period: w.Cycle.Period, Fade: w.Cycle.Fade, Surges: w.Cycle.Surges, Floor: w.Cycle.Floor, Ends: w.Cycle.Ends},
		Fertility: w.FertilityNow(), NextDawn: w.NextSurge(), Hazard: w.Hazard,
		Wall: record.Wall{Value: w.Thin, Stage: w.thinStage()},
		Decline: record.Decline{
			Index: w.Decline.Index, Held: w.Decline.Held, Rising: w.Decline.Rising, Births: w.Decline.Births,
			HeldNow: w.Decline.HeldNow, RisingNow: w.Decline.RisingNow, BirthsNow: w.Decline.BirthsNow,
			PeakHeld: w.Decline.PeakHeld, PeakRising: w.Decline.PeakRising, PeakBirths: w.Decline.PeakBirths,
			PeakHeldAt: w.Decline.PeakHeldAt, AtWaning: w.Decline.AtWaning,
			Crossed: w.Decline.Crossed, Fell: w.Decline.Fell, ByIndex: w.Decline.ByIndex,
		},
	}
	l := g.Law
	p := record.Place{
		Name: g.Region.Name, Code: g.Region.Code, Anchor: g.Anchor(), Pos: g.Region.Pos, HasSol: g.Region.HasSol, Sol: g.Sol,
		Zone: l.Zone, Laws: record.Laws{R: l.R, Z: l.Z, Density: l.Density, Youth: l.Youth, Metals: l.Metals, Glare: l.Glare, Crowd: l.Crowd, Exotic: l.Exotic},
		ArmDist: l.ArmDist, MeanGap: galaxy.MeanSpacing(len(g.Stars), g.Radius, g.Thickness), FieldRad: g.Radius,
		Near: []record.NearFeature{}, Beyond: []string{},
	}
	if l.Arm != nil {
		p.Arm = l.Arm.Name
	}
	for _, nf := range l.Near {
		if nf.F.Kind == galaxy.Sky {
			p.Beyond = append(p.Beyond, nf.F.Key)
		} else {
			p.Near = append(p.Near, record.NearFeature{Key: nf.F.Key, Dist: nf.Dist})
		}
	}
	for i := range g.Stars {
		if g.Stars[i].Real {
			p.Real++
		}
	}
	d.Place = p
	c := record.Counts{Civs: len(w.Civs), Traces: len(w.Traces), Events: len(r.Chronicle), Tales: len(r.Tellings), Plagues: len(w.Plagues), Wars: len(w.Wars), Leaders: len(w.Leaders), Fleets: len(w.Expeditions), Contracts: len(w.Contracts)}
	for _, cv := range w.Civs {
		if cv.Active() {
			c.Standing++
			if cv.Species.Profile().Eats {
				c.Eaten += len(cv.Systems)
			}
		}
		if cv.Stage == Remnant {
			c.Remnants++
		}
	}
	for i := range g.Stars {
		if w.Bio[i] == BioComplex {
			c.Complex++
		}
	}
	for _, l := range w.Legacies {
		if l.Speaking() {
			c.Speaking++
		}
		if l.Maker >= 0 {
			c.Remains++
			if l.State == Lost {
				c.Crumbled++
			}
		}
	}
	d.Counts = c
	return d
}

func (w *World) exportState() *record.State {
	st := &record.State{Voices: map[int]record.Voice{}}
	for i := range w.G.Stars {
		s := &w.G.Stars[i]
		st.Stars = append(st.Stars, &record.Star{
			ID: s.ID, Designation: s.Name, Kind: s.Desig, Real: s.Real, Alt: s.Alt, Class: string(s.Class), Note: s.Note, Remnant: s.Remnant,
			X: s.X, Y: s.Y, Z: s.Z, Hab: s.Hab, Mult: s.Mult, Lifetime: s.Lifetime, DiesAt: Year(s.DiesAt), Failing: s.Failing, Mag: s.Mag,
			System: w.G.Sys[i], Bio: [...]string{"none", "simple", "complex"}[w.Bio[i]], Held: w.Owner[i],
		})
	}
	for _, s := range st.Stars {
		if s.Held >= 0 {
			c := w.Civs[s.Held]
			s.Guns = w.gunsAt(c, s.ID)
			s.GridBroken = c.GridBroken[s.ID]
		}
	}
	for _, sp := range w.Species {
		r := &record.Species{ID: sp.ID, Sub: sp.Sub.String(), Mods: sp.ModKeys(), Channel: sp.Channel, Powers: sp.Powers, Made: sp.Made, Parent: -1, First: -1, Lifespan: sp.Lifespan, Software: sp.Software}
		if r.Mods == nil {
			r.Mods = []string{}
		}
		if r.Powers == nil {
			r.Powers = []string{}
		}
		if sp.World != nil {
			r.World = sp.World.Key
		}
		r.Traits = traitKeysOf(sp)
		if sp.Parent != nil {
			r.Parent = sp.Parent.ID
		}
		for _, c := range w.Civs {
			if c.Species == sp {
				r.First = c.ID
				break
			}
		}
		st.Species = append(st.Species, r)
	}
	for _, c := range w.Civs {
		st.Civs = append(st.Civs, w.exportCiv(c))
	}
	for _, a := range w.Ages {
		r := &record.Age{Index: a.Index, Start: a.Start, End: a.End, Ender: a.Ender, Elders: []int{}}
		for _, e := range a.Elders {
			r.Elders = append(r.Elders, e.ID)
			er := &record.Elder{ID: e.ID, Age: e.Age, Portrait: e.Portrait, Rose: e.Rose, Fell: e.Fell, Legacies: []int{}}
			for _, l := range e.Legacies {
				er.Legacies = append(er.Legacies, l.ID)
			}
			st.Elders = append(st.Elders, er)
		}
		st.Ages = append(st.Ages, r)
	}
	for _, l := range w.Legacies {
		r := &record.Remain{
			ID: l.ID, Age: l.Age, Elder: -1, Maker: l.Maker, Kind: l.Kind.String(), Star: l.Star, Node: l.Node, Portrait: l.Portrait, State: l.State.String(),
			People: l.People, Payload: l.Payload.String(), Listeners: l.Listeners, Woken: l.Woken, Finder: l.Finder, Level: l.Level, Cond: l.Cond.String(), Hardy: l.Hardy,
			Source: l.Source, Testament: []int{}, Plague: l.Plague, Wrecks: l.Wrecks, Derelicts: l.Derelicts, At: [3]float64(l.At), Adrift: l.Adrift,
		}
		if l.Elder != nil {
			r.Elder = l.Elder.ID
		}
		st.Remains = append(st.Remains, r)
	}
	for _, s := range w.Sources {
		r := &record.Source{
			ID: s.ID, Key: s.Key, Kind: s.Kind.String(), Star: s.Star, Radius: s.Radius, Yield: s.Yield, Needs: orEmpty(s.Needs), With: orEmpty(s.With),
			Cradle: s.Cradle, Rarity: s.Rarity, Grants: orEmpty(s.Grants), Levels: s.Levels, Reach: s.Reach, Mobile: s.Mobile, Holder: s.Holder, Carried: s.Carried,
			Legacy: s.Legacy, Since: s.Since, Wear: s.Wear, Form: s.Form, Sentient: s.Sentient, Maker: s.Maker, Made: s.Made, Given: s.Given, Fate: s.Fate,
		}
		if s.Feature != nil {
			r.Feature = s.Feature.Key
		}
		st.Sources = append(st.Sources, r)
	}
	st.Traces = []record.Trace{}
	for _, t := range w.Traces {
		st.Traces = append(st.Traces, record.Trace{Star: t.Star, Kind: t.Kind, Civ: t.Civ, Species: t.Species, Year: t.Year})
	}
	for _, p := range w.Plagues {
		st.Plagues = append(st.Plagues, &record.Plague{
			ID: p.ID, Kind: p.Kind.String(), Contagion: p.Contagion, Lethality: p.Lethality, Band: p.Band, Engineered: p.Engineered, Conscious: p.Conscious, Profile: p.Profile,
			Born: p.Born, FirstHost: p.FirstHost, Cause: p.Cause, Hosts: p.Hosts, Peak: p.Peak, Caught: p.Caught, Worlds: p.Worlds, Peoples: p.Peoples, Cults: p.Cults,
			Cures: p.Cures, Refusals: p.Refusals, Woken: p.Woken, LastHost: p.LastHost, Extinct: p.Extinct, Wildfire: p.Wildfire, Maker: p.Maker, Made: p.Made,
			Rider: p.Rider, Transmitter: p.Transmitter, Poisonings: p.Poisonings,
		})
	}
	st.Reservoirs = []record.Reservoir{}
	for _, s := range sortedInts(w.Reservoir) {
		st.Reservoirs = append(st.Reservoirs, record.Reservoir{Star: s, Plague: w.Reservoir[s].Plague, Until: w.Reservoir[s].Until})
	}
	for _, wr := range w.Wars {
		r := &record.War{
			ID: wr.ID, Sides: wr.Sides, Began: wr.Began, Ended: wr.Ended, Over: wr.Over, Cause: wr.Cause, CauseOf: wr.CauseOf, Aim: wr.Aim, Named: wr.Named, Nth: wr.Nth,
			Will: wr.Will, Taken: wr.Taken, Glassed: wr.Glassed, Lost: wr.Lost, Result: wr.Result, Pact: wr.Pact, Principal: wr.Principal, Hire: wr.Hire,
		}
		if wr.Gap != nil {
			r.Hunt = &record.Hunt{X: wr.Gap.X, Y: wr.Gap.Y, Radius: wr.Gap.Radius, Losses: wr.Gap.Losses, Star: wr.Gap.Star, Since: wr.Gap.Since, Empty: wr.Gap.Empty}
		}
		st.Wars = append(st.Wars, r)
	}
	st.Leaders = []*record.Leader{}
	for _, l := range w.Leaders {
		st.Leaders = append(st.Leaders, &record.Leader{
			ID: l.ID, Civ: l.Civ, Rose: l.Rose, Ended: l.Ended, End: l.End, Occasion: l.Occasion, Form: l.Form, Stance: l.Stance, Own: l.Own,
			Bent: l.Bent, Push: l.Push, Front: l.Front, Flees: l.Flees, Fled: l.Fled, Fleet: l.Fleet, Star: l.Star, Until: l.Until,
			Deathless: l.Deathless, Mad: l.Mad, Faced: l.Faced, Doublings: l.Doublings, Battles: l.Battles,
		})
	}
	for _, p := range w.Pacts {
		st.Pacts = append(st.Pacts, &record.Pact{ID: p.ID, Members: orEmpty(p.Members), Kind: p.Kind.String(), Target: p.Target, Formed: p.Formed, Ended: p.Ended, Over: p.Over})
	}
	st.Betrayals = []record.Betrayal{}
	for _, b := range w.Betrayals {
		st.Betrayals = append(st.Betrayals, record.Betrayal{By: b.By, Against: b.Against, Year: b.Year, Shape: b.Shape, Weight: b.Weight})
	}
	for _, x := range w.Expeditions {
		st.Fleets = append(st.Fleets, &record.Fleet{
			ID: x.ID, Owner: x.Owner, Target: x.Target, Kind: x.Kind.String(), Star: x.Star, From: x.From, Ships: x.Ships, Launched: x.Launched, Arrive: x.Arrive,
			Base: x.Base, Returning: x.Returning, Over: x.Over, LaidUp: x.LaidUp, Laid: x.Laid, Manned: x.Manned, Held: orEmpty(x.Held), Seen: idsOfBools(x.Seen), Battles: x.Battles, Wins: x.Wins,
			Turned: x.Turned, Out: x.Out, Back: x.Back, Sieges: x.Sieges, Drive: x.Drive, Warning: x.Warning, Quarry: x.Quarry, Picket: x.Picket, Contract: x.Contract,
		})
	}
	for _, k := range w.Contracts {
		st.Contracts = append(st.Contracts, &record.Contract{
			ID: k.ID, Buyer: k.Buyer, Seller: k.Seller, By: k.By, Ask: exportTerm(k.Ask), Pay: exportTerm(k.Pay), Length: k.Length, Offered: k.Offered, Formed: k.Formed,
			Until: k.Until, Ended: k.Ended, State: k.State.String(), Broke: k.Broke, Failed: k.Failed, Missed: k.Missed, AskDone: k.AskDone, Taught: k.Taught,
			Burned: k.Burned, Tribute: k.Tribute, BoughtOff: k.BoughtOff,
		})
	}
	st.Fathomings = []record.Fathoming{}
	for _, f := range w.Fathomings {
		st.Fathomings = append(st.Fathomings, record.Fathoming{Year: f.Year, Who: f.Who, Whom: f.Whom, How: f.How, Since: f.Since, Diff: f.Diff, Mutual: f.Mutual, Reversed: f.Reversed})
	}
	st.Battles = []record.Battle{}
	for _, b := range w.Battles {
		st.Battles = append(st.Battles, record.Battle{Year: b.Year, Star: b.Star, Attacker: b.Attacker, Defender: b.Defender, Ships: b.Ships, Held: b.Held, Gap: b.Gap, Won: b.Won, Outcome: b.Outcome, Relief: b.Relief})
	}
	st.Meetings = []record.Meeting{}
	for _, m := range w.Meetings {
		st.Meetings = append(st.Meetings, record.Meeting{Year: m.Year, Quarry: m.Quarry, Interceptor: m.Interceptor, Owner: m.Owner, Seer: m.Seer, Ships: m.Ships, Sent: m.Sent, Won: m.Won, Broken: m.Broken, Lost: m.Lost})
	}
	st.Sightings = []record.Sighting{}
	for i, s := range w.Watch {
		st.Sightings = append(st.Sightings, record.Sighting{
			ID: i, Fleet: s.Fleet, Owner: s.Owner, Seer: s.Seer, Kind: s.Kind.String(), From: s.From, Star: s.Star, Launched: s.Launched, Arrive: s.Arrive, Leg: s.Leg,
			Year: s.Year, Ships: s.Ships, Mil: s.Mil, Speed: s.Speed, Eye: s.Eye, EyeWork: s.EyeName, EyeStar: s.EyeStar, Feasible: s.Feasible, Intercept: s.Intercept, Offered: s.Offered,
		})
	}
	return st
}

// MarshalJSON writes a term as the record holds it, so that a term in an
// event's parameters reads as one in a contract.
func (t Term) MarshalJSON() ([]byte, error) { return json.Marshal(exportTerm(t)) }

func exportTerm(t Term) record.Term {
	r := record.Term{Kind: t.Kind.String(), Amount: t.Amount, Source: t.Source, Node: t.Node, Star: t.Star, Work: t.Work, Target: t.Target, Fleet: t.Fleet}
	if t.Kind == 0 { // a flow: the kind
		r.Res = t.Res.Symbol()
	}
	return r
}

func traitKeysOf(sp *species.Species) []string {
	out := []string{}
	for _, t := range sp.Traits {
		out = append(out, t.Key)
	}
	return out
}

func orEmpty[T any](xs []T) []T {
	if xs == nil {
		return []T{}
	}
	return xs
}

func keysOfBools(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func idsOfBools(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k, v := range m {
		if v {
			out = append(out, k)
		}
	}
	sort.Ints(out)
	return out
}

func (w *World) exportCiv(c *Civ) *record.Civ {
	r := &record.Civ{
		ID: c.ID, Species: c.Species.ID, Origin: c.Origin, Home: c.Home, Cradle: c.Cradle, Born: c.Born, Ended: c.Ended, Fell: c.Fell,
		Stage: [...]string{"emergent", "interstellar", "zenith", "remnant", "dead"}[c.Stage], Fate: c.Fate.String(), Cause: c.Cause, Into: c.Into, IntoCivs: c.IntoCivs,
		FallEvent: c.FallEvent, EndEvent: c.EndEvent, Named: c.Named, Hosts: c.Hosts,
		Morality: record.Morality{Kind: c.Morality.Kind.String(), Object: c.Morality.Object},
		Lifted:   keysOfBools(c.Lifted), Systems: orEmpty(c.Systems), Peak: c.Peak, Voyages: []record.Voyage{},
		Known: keysOfBools(c.Known), Dormant: keysOfBools(c.Shed), Learned: c.Learned, Era: c.Era, Pursuit: c.Pursuit, Progress: c.Progress, Miracles: c.Miracles, Remade: c.Remade,
		Levels: record.Levels{Mil: c.Mil, Sur: c.Sur, Soc: c.Soc, Wis: c.Wis}, Reach: c.Reach, Speed: c.Speed, Envelope: c.Envelope, Morale: c.Morale, Dials: c.Dials,
		Order: []string{}, Income: c.Income, Upkeep: c.Upkeep, Surplus: c.Surplus, Want: c.Want,
		Structures: c.Structures, Works: []record.Work{}, Wielded: []int{}, Found: idsOfBools(c.Found), Heard: idsOfBools(c.Heard), Uplifts: c.Uplifts, Ruled: c.Ruled,
		Docks: Docks(w, c), Salvage: c.Salvage,
		Wars: idsOfBools(c.Wars), Trade: idsOfBools(c.Trade), Master: c.Master, Vassal: c.Vassal, Pacts: orEmpty(c.Pacts), Truce: c.Truce, Fought: c.Fought, Grudge: c.Grudge,
		Ridden: idsOfBools(c.Ridden), Contracts: orEmpty(c.Contracts), Taught: c.Taught, Sellsword: c.Sellsword, Dependent: idsOfBools(c.Dependent), Embargo: idsOfBools(c.Embargo),
		Refused: c.Refused, Barred: idsOfBools(c.Barred),
		Infections: map[int]record.Infection{}, Immune: idsOfBools(c.Immune), Suspect: idsOfBools(c.Suspect), Closed: idsOfBools(c.Closed), Own: c.Own, Weapons: map[string]record.Weapon{},
		Stiff: c.Stiff, Continuity: w.continuity(c), Lifespan: w.lifespanOf(c), Leader: leaderID(c.Leader), Ossified: c.Ossified, Still: c.Still, Line: orEmpty(c.Line), Claim: idsOfBools(c.Claim), Asleep: c.Asleep, Slept: c.Slept, Aloft: c.Aloft, Rested: c.Rested,
		Drifts: c.Drifts, Searching: c.Searching, Starfaring: c.Starfaring, Faced: keysOfBools(c.Faced), Scars: keysOfBools(c.Scars), Boons: keysOfBools(c.Boons), Record: []record.Record{},
		DarkAges: c.DarkAges, KnowsCycle: c.KnowsCycle, Ascended: c.Ascended, Renewed: c.Renewed, Renaissances: c.Renaissances, Dying: c.Dying, Endure: c.Endure, Rare: keysOfBools(c.Rare),
		LastDark: c.LastDark, LastTaken: c.LastTaken, LastUnmade: c.LastUnmade,
	}
	for _, v := range c.Voyages {
		r.Voyages = append(r.Voyages, record.Voyage{Target: v.Target, Arrive: v.Arrive, Blind: v.Blind})
	}
	for _, o := range c.Order {
		r.Order = append(r.Order, o.String())
	}
	for _, wk := range c.Works {
		r.Works = append(r.Works, record.Work{Key: wk.Key, Node: wk.Node, Star: wk.Star, Legacy: wk.Legacy, Dark: wk.Dark})
	}
	for _, l := range c.Wielded {
		r.Wielded = append(r.Wielded, l.ID)
	}
	for _, pid := range sortedInts(c.Infections) {
		inf := c.Infections[pid]
		r.Infections[pid] = record.Infection{Since: inf.Since, From: inf.From, Road: inf.Road, Contained: inf.Contained, Held: inf.Held, Carrier: inf.Carrier}
	}
	for k, wp := range c.Weapons {
		r.Weapons[k] = record.Weapon{Plague: wp.Plague, Target: wp.Target, Node: wp.Node, Made: wp.Made}
	}
	for _, rec := range c.Record {
		x := record.Record{Kind: rec.Kind, Filter: rec.Filter, Narrow: rec.Narrow, Legacy: rec.Legacy, Known: rec.Known}
		if rec.Kind == "faced" {
			x.Outcome = [...]string{"overcame", "scarred", "declined"}[rec.Outcome]
		}
		if rec.Kind == "faced" || rec.Kind == "foresaw" {
			x.Legacy = -1
		}
		r.Record = append(r.Record, x)
	}
	for _, m := range []*map[string]int{&r.Structures, &r.Taught} {
		if *m == nil {
			*m = map[string]int{}
		}
	}
	if r.Learned == nil {
		r.Learned = map[string]Year{}
	}
	if r.Miracles == nil {
		r.Miracles = map[string]string{}
	}
	if r.Remade == nil {
		r.Remade = map[string]Year{}
	}
	if r.Truce == nil {
		r.Truce = map[int]Year{}
	}
	if r.Fought == nil {
		r.Fought = map[int]int{}
	}
	if r.Grudge == nil {
		r.Grudge = map[int]float64{}
	}
	if r.Refused == nil {
		r.Refused = map[int]Year{}
	}
	if r.IntoCivs == nil {
		r.IntoCivs = []int{}
	}
	r.Knowledge = w.exportKnowledge(c)
	r.Batch = w.exportBatch(c)
	return r
}

func (w *World) exportKnowledge(c *Civ) record.Knowledge {
	if c.monsters == nil {
		w.reckon(c)
	}
	k := record.Knowledge{
		Met: idsOfBools(c.Met), Reached: idsOfBools(c.Reached), Fathomed: idsOfBools(c.Fathomed), FathomTried: c.FathomTried, FathomedAt: c.FathomedAt,
		Perceives: []int{}, Alien: map[int]float64{}, Charted: c.Charted, Marked: idsOfBools(c.Marked), Scouted: c.Scouted, Intel: map[int]record.Intel{}, Regards: map[int]int8{},
		Monsters: idsOfBools(c.monsters), Dread: []int{}, Scapegoat: c.foeNow, Watched: idsOfBools(c.Watched), Sightings: map[int]int{},
		Memory: record.Memory{Forgot: c.Tally.Forgot, Revised: c.Tally.Revised, Blamed: c.Tally.Blamed}, LoreDials: c.LoreDials,
	}
	for _, t := range c.Lore {
		if !t.Forgot {
			k.Memory.Held++
			if t.Wear >= 2 {
				k.Memory.Myth++
			}
		}
	}
	for _, e := range w.Civs {
		if w.perceives(c, e) {
			k.Perceives = append(k.Perceives, e.ID)
		}
		if e != c {
			if r := w.regard(c, e.ID); r != 0 {
				k.Regards[e.ID] = r
			}
		}
	}
	for id := range c.FathomTried {
		k.Alien[id] = Difference(c, w.Civs[id])
	}
	for i := range w.G.Stars {
		if w.dread(c, i) {
			k.Dread = append(k.Dread, i)
		}
	}
	for _, id := range sortedInts(c.Intel) {
		in := c.Intel[id]
		k.Intel[id] = record.Intel{Mil: in.Mil, Ships: in.Ships, Guns: in.Guns, Total: in.Total, Relief: in.Relief, Star: in.Star, Year: in.Year, Sick: in.Sick}
	}
	index := map[*Sighting]int{}
	for i, s := range w.Watch {
		index[s] = i
	}
	for fleet, s := range c.Sightings {
		if i, ok := index[s]; ok {
			k.Sightings[fleet] = i
		}
	}
	if k.FathomTried == nil {
		k.FathomTried = map[int]Year{}
	}
	if k.FathomedAt == nil {
		k.FathomedAt = map[int]int{}
	}
	if k.Charted == nil {
		k.Charted = map[int]Year{}
	}
	if k.Scouted == nil {
		k.Scouted = map[int]Year{}
	}
	return k
}

func (w *World) exportBatch(c *Civ) record.Batch {
	b := record.Batch{
		Tally: c.Tally, HighIncome: c.HighIncome, HighUpkeep: c.HighUpkeep, HighWant: c.HighWant, PeakTrade: orEmpty(c.PeakTrade), ShedTicks: c.ShedTicks, Built: c.Built,
		Had: keysOfBools(c.Had), Harnessed: keysOfBools(c.Harnessed), Granted: keysOfBools(c.Granted), FellDependent: c.FellDependent, PeakWis: c.PeakWis, WisFrom: WisdomParts(c),
		PeakShips: c.PeakShips, FirstPlague: c.FirstPlague, Sire: c.Sire, DockRate: c.DockRate, WantShips: c.WantShips, GarrisonWant: c.GarrisonWant,
	}
	if b.ShedTicks == nil {
		b.ShedTicks = map[string]int{}
	}
	if b.Built == nil {
		b.Built = map[string]int{}
	}
	if b.DockRate == nil {
		b.DockRate = map[int]float64{}
	}
	return b
}

// exportChronicle is the chronicle in the order it is told, with every
// parameter as JSON holds it.
func (w *World) exportChronicle() []*record.Event {
	out := make([]*record.Event, 0, len(w.Chronicle))
	for _, e := range w.Chronicle {
		out = append(out, exportEvent(e))
	}
	return out
}

func exportEvent(e *Event) *record.Event {
	return &record.Event{ID: e.ID, Year: e.Year, Kind: e.Kind, Subject: e.Subject, Object: e.Object, Star: e.Star, Legacy: e.Legacy, Plague: e.Plague, N: e.N, P: record.Normalise(e.P), Fact: e.IsFact()}
}

// exportTellings is every people's tales in the order it holds them,
// with what the view derives (the believed year, the teller's sort and
// weight, the rank among its dearest), then every remain's testament as
// frozen tales, whose ids the remains point at.
func (w *World) exportTellings(st *record.State) []*record.Tale {
	var out []*record.Tale
	for _, c := range w.Civs {
		ranks := w.ranks(c)
		for i, t := range c.Lore {
			r := exportTale(len(out), c.ID, -1, t, w.Events[t.Fact], w.Present)
			r.Sort, r.Weight = w.judgment(c, w.Events[t.Fact])
			r.Rank = ranks[i]
			out = append(out, r)
		}
	}
	for _, l := range w.Legacies {
		for _, in := range l.Testament {
			r := exportTale(len(out), l.Maker, l.ID, &in.Tale, w.Events[in.Tale.Fact], w.Present)
			r.Sort, r.Weight = in.Sort.String(), in.Weight
			fz := in.Frozen
			r.Frozen = &record.Frozen{Perceived: orEmpty(fz.Perceived), Met: orEmpty(fz.Met), Active: orEmpty(fz.Active), Living: orEmpty(fz.Living), Ours: fz.Ours}
			st.Remains[l.ID].Testament = append(st.Remains[l.ID].Testament, r.ID)
			out = append(out, r)
		}
	}
	if out == nil {
		out = []*record.Tale{}
	}
	return out
}

func exportTale(id, civ, remain int, t *Tale, f *Event, present Year) *record.Tale {
	r := &record.Tale{ID: id, Civ: civ, Remain: remain, Fact: t.Fact, Learned: t.Learned, Source: t.Source.String(), From: t.From, Slant: t.Slant, Wear: t.Wear, Blamed: t.Blamed, Revised: t.Revised, Forgot: t.Forgot}
	switch t.Wear {
	case 0:
		y := f.Year
		r.Believed = &y
	case 1:
		y := f.Year - present
		const rough = Year(100_000)
		y = (y-rough/2)/rough*rough + present
		r.Believed = &y
	}
	return r
}

// judgment is a fact as one people judges it, for the tale record.
func (w *World) judgment(c *Civ, f *Event) (string, float64) {
	s, x := sortFor(c, f)
	return s.String(), x
}

// ranks is each tale's place among a people's dearest, by index into
// its lore: the dearest first, ties by the later fact, then by the order
// held; a forgotten tale has none.
func (w *World) ranks(c *Civ) []int {
	out := make([]int, len(c.Lore))
	var kept []int
	for i, t := range c.Lore {
		if !t.Forgot {
			kept = append(kept, i)
		}
	}
	sort.SliceStable(kept, func(a, b int) bool {
		ta, tb := c.Lore[kept[a]], c.Lore[kept[b]]
		fa, fb := w.Events[ta.Fact], w.Events[tb.Fact]
		da, db := w.dearness(c, fa, ta), w.dearness(c, fb, tb)
		if da != db {
			return da > db
		}
		return fa.Year > fb.Year
	})
	for rank, i := range kept {
		out[i] = rank + 1
	}
	return out
}

// leaderID is a leader's id, -1 for none.
func leaderID(l *Leader) int {
	if l == nil {
		return -1
	}
	return l.ID
}
