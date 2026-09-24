// Package warshape is the measurement behind the war gate (specs/plan.md
// steps 10 and 11; specs/proposals/war.md): it reads one run's record,
// and nothing else, for the seven patterns of war the rules are meant to
// make — short and long wars, old enemies, world wars, border disputes,
// cold wars, proxy wars and conquest waves — with the vassal's lot and
// the health of the whole beside them. Report turns a batch of reads into
// the lines techstats and TestWarShape print.
package warshape

import (
	"slices"
	"sort"
	"time"

	"worldgen/internal/record"
	"worldgen/internal/species"
)

// The thresholds the patterns are read at. They are the proposal's.
const (
	shortTicks = 3      // a war this long or shorter is short
	longTicks  = 12     // and this long or longer, long
	large      = 5      // worlds a realm holds to count as large
	coldSpan   = 50_000 // years at peace for a cold war
	waveSpan   = 50_000 // years a conquest wave takes its peoples in
	waveFrom   = 3      // peoples taken from, for a wave
	worldPeak  = 8      // peoples at war at once, for a world war
	worldFront = 3      // fronts fought on at that moment
	oldWars    = 3      // fought wars between one pair, for old enemies
	loopWars   = 20     // more wars than this between one pair is a loop
)

// War is one war as the record shows it, with what the patterns read
// joined to it: its battles, the campaigns each side sent, the size and
// the masters of its sides when it began.
type War struct {
	ID           int
	Sides        [2]int // the declarer first
	Began, End   record.Year
	Over         bool
	Ticks        int // its length in ticks, at least one
	Battles      int
	BattleTicks  int    // ticks with a battle in them
	CarriedTicks int    // ticks with a battle, or a campaign of either side in flight at the other: the war being carried
	Campaigns    [2]int // campaign fleets each side launched at the other while it lasted
	Taken        int
	Cause        string
	Aim          string // what the declarer went to war for, when the record says (stage 1 on)
	Fleets       bool   // both sides send fleets: a war of fleets, not a living world's or a sleeper's blow against a people that cannot answer it
	Result       string
	Worlds       [2]int  // each side's worlds when it began
	Masters      [2]int  // each side's master when it began, -1 for free
	Vassal       [2]bool // held as a vassal rather than a slave
	Pact         int
	Principal    int
}

// Fought says whether a battle was fought in it.
func (w *War) Fought() bool { return w.Battles > 0 }

// System is a set of wars joined by sharing a people while both were
// open: at its widest moment, the peoples at war in it and the fronts
// among them that saw a battle.
type System struct {
	Wars   int
	Peak   int
	Fronts int
}

// Bond is one people held by another, from the record's changes of
// master: how it began, and for how long it lasted.
type Bond struct {
	Vassal bool
	Origin string // war, meeting, born (or uplifted), ridden, other
	Years  record.Year
}

// Attack is a war declared on a vassal: whether the one who declared was
// smaller than the patron, the patron itself or a power its size or
// greater, and what the patron did.
type Attack struct {
	By     string // smaller, rival, patron
	Patron string // joined, sent ships, stayed out
}

// Shape is one run's read.
type Shape struct {
	Seed        uint64
	Took        time.Duration // the run's time, when the caller knows it
	Ages        float64       // Myr from the dawn to the present
	Capped      bool
	Decline     record.Decline
	Peoples     int
	PeopleTicks int
	Pairs       int // distinct pairs that knew of each other
	First       int // first wars between a pair

	Wars       []*War
	Stray      int   // battles in no war between their two
	PairWars   []int // wars per pair that fought, each pair once
	PairFought []int // fought wars per pair
	Systems    []System
	Cold       int   // pairs with a cold war: large, at peace long, with an incident
	Quiet      int   // pairs large and at peace as long with no incident
	VassalWars int   // wars between vassals of different masters
	MasterIn   int   // of them, with a master's ships in them
	Proxy      int   // of those, the masters not at war with each other
	Waves      []int // for each people that made a wave, the most peoples it took from in one span
	Bonds      []Bond
	Attacks    []Attack
	Tributes   int // tribute facts: a yielding people paying in a commodity
}

// span is a stretch of a quantity's history: from a year on, the value.
type span struct {
	at record.Year
	v  int
}

// at is the value before the year: what a change in the same tick has
// not yet touched.
func at(h []span, y record.Year, zero int) int {
	i := sort.Search(len(h), func(i int) bool { return h[i].at >= y })
	if i == 0 {
		return zero
	}
	return h[i-1].v
}

type pair [2]int

// change is a people's master changing: to whom, and as what.
type change struct {
	y      record.Year
	master int
	vassal bool
}

func pairOf(a, b int) pair { return pair{min(a, b), max(a, b)} }

// Read measures one run.
func Read(r *record.Run) *Shape {
	d, st := r.Dossier, r.State
	s := &Shape{Seed: d.Seed, Ages: float64(d.Present-d.Dawn) / 1e6, Capped: d.Capped, Decline: d.Decline, Peoples: len(st.Civs)}
	step := max(d.Step, 1)
	seen := map[pair]bool{}
	for _, c := range st.Civs {
		s.PeopleTicks += c.Batch.Tally.Ticks
		for _, o := range c.Knowledge.Met {
			if k := pairOf(c.ID, o); k[0] != k[1] && !seen[k] {
				seen[k] = true
				s.Pairs++
			}
		}
	}

	// The worlds each people held and the master it had, through time,
	// folded from the silent events that carry them.
	worlds := map[int][]span{}
	masters := map[int][]span{}
	vassal := map[int][]span{}
	changes := map[int][]change{}
	born := map[int]record.Year{}
	ridden := map[int][]record.Year{}
	met := map[pair][]record.Year{}
	warBound := map[pair][]record.Year{} // the years a war between the two ended
	var taken, sent []*record.Event
	for _, e := range r.Chronicle {
		switch e.Kind {
		case record.KFleetSent:
			sent = append(sent, e)
		case record.KWorldHeld, record.KWorldLost:
			h := worlds[e.Subject]
			n := 0
			if len(h) > 0 {
				n = h[len(h)-1].v
			}
			if e.Kind == record.KWorldHeld {
				n++
			} else {
				n--
			}
			worlds[e.Subject] = append(h, span{e.Year, n})
		case record.KMaster:
			m, v := e.Int("master"), e.Bool("vassal")
			masters[e.Subject] = append(masters[e.Subject], span{e.Year, m})
			vassal[e.Subject] = append(vassal[e.Subject], span{e.Year, b2i(v)})
			changes[e.Subject] = append(changes[e.Subject], change{e.Year, m, v})
		case record.KCivBorn:
			born[e.Subject] = e.Year
		case record.FUplift:
			born[e.Object] = e.Year
		case record.KRidden, record.KRiddenWar, record.KBornRidden:
			for _, c := range []int{e.Subject, e.Object} {
				if c >= 0 {
					ridden[c] = append(ridden[c], e.Year)
				}
			}
		case record.FMet:
			k := pairOf(e.Subject, e.Object)
			met[k] = append(met[k], e.Year)
		case record.KWarOver:
			k := pairOf(e.Subject, e.Object)
			warBound[k] = append(warBound[k], e.Year)
		case record.FTaken:
			if w := e.Str("way"); w == "war" || w == "host_war" {
				taken = append(taken, e)
			}
		case record.FTribute:
			s.Tributes++
		}
	}
	for _, h := range [](map[int][]span){worlds, masters, vassal} {
		for k := range h {
			sort.SliceStable(h[k], func(i, j int) bool { return h[k][i].at < h[k][j].at })
		}
	}
	worldsAt := func(c int, y record.Year) int { return at(worlds[c], y, 0) }
	launches := map[int]bool{}
	for _, sp := range st.Species {
		launches[sp.ID] = species.Rebuild(sp.ID, sp.Sub, sp.Mods, sp.Channel, sp.Powers, sp.World, sp.Traits, sp.Made).Profile().Can(species.Launches)
	}
	sails := func(c int) bool { return launches[st.Civs[c].Species] }
	masterAt := func(c int, y record.Year) int { return at(masters[c], y, -1) }
	ended := func(c int) record.Year {
		if cv := st.Civs[c]; cv.Ended > 0 {
			return cv.Ended
		}
		return d.Present
	}

	// The wars, with their battles and campaigns joined by the pair and
	// the years: a pair has one war open at a time.
	byPair := map[pair][]*War{}
	for _, wr := range st.Wars {
		w := &War{ID: wr.ID, Sides: wr.Sides, Began: wr.Began, End: wr.Ended, Over: wr.Over, Taken: wr.Taken[0] + wr.Taken[1],
			Cause: wr.Cause, Aim: wr.Aim, Result: wr.Result, Pact: wr.Pact, Principal: wr.Principal, Fleets: sails(wr.Sides[0]) && sails(wr.Sides[1])}
		if !wr.Over {
			w.End, w.Result = d.Present, "unfinished"
		}
		w.Ticks = max(1, int((w.End-w.Began)/step))
		for i, c := range wr.Sides {
			w.Worlds[i] = worldsAt(c, wr.Began)
			w.Masters[i] = masterAt(c, wr.Began)
			w.Vassal[i] = w.Masters[i] >= 0 && at(vassal[c], wr.Began, 0) == 1
		}
		if wr.Nth == 1 {
			s.First++
		}
		s.Wars = append(s.Wars, w)
		k := pairOf(wr.Sides[0], wr.Sides[1])
		byPair[k] = append(byPair[k], w)
	}
	for k := range byPair {
		sort.Slice(byPair[k], func(i, j int) bool { return byPair[k][i].Began < byPair[k][j].Began })
	}
	warOf := func(a, b int, y record.Year) *War {
		for _, w := range byPair[pairOf(a, b)] {
			if w.Began <= y && y <= w.End {
				return w
			}
		}
		return nil
	}
	ticks := map[*War]map[record.Year]bool{}
	stray := map[pair][]record.Year{}
	for _, b := range st.Battles {
		w := warOf(b.Attacker, b.Defender, b.Year)
		if w == nil {
			s.Stray++
			k := pairOf(b.Attacker, b.Defender)
			stray[k] = append(stray[k], b.Year)
			continue
		}
		w.Battles++
		if ticks[w] == nil {
			ticks[w] = map[record.Year]bool{}
		}
		ticks[w][b.Year/step] = true
	}
	for w, t := range ticks {
		w.BattleTicks = len(t)
	}
	carried := map[*War]map[record.Year]bool{}
	for w, t := range ticks {
		carried[w] = map[record.Year]bool{}
		for k := range t {
			carried[w][k] = true
		}
	}
	for _, e := range sent {
		w := warOf(e.Subject, e.Object, e.Year)
		if w == nil {
			continue
		}
		if carried[w] == nil {
			carried[w] = map[record.Year]bool{}
		}
		for k := e.Year / step; k <= min(e.Year+record.Year(e.Int("away")), w.End)/step; k++ {
			carried[w][k] = true
		}
	}
	for w, t := range carried {
		w.CarriedTicks = len(t)
	}
	for _, x := range st.Fleets {
		if x.Kind != "campaign" || x.Target < 0 {
			continue
		}
		if w := warOf(x.Owner, x.Target, x.Launched); w != nil {
			if w.Sides[0] == x.Owner {
				w.Campaigns[0]++
			} else {
				w.Campaigns[1]++
			}
		}
	}

	// Old enemies: the wars of each pair, and the fought ones.
	for _, ws := range byPair {
		n := 0
		for _, w := range ws {
			if w.Fought() {
				n++
			}
		}
		s.PairWars = append(s.PairWars, len(ws))
		s.PairFought = append(s.PairFought, n)
	}

	s.Systems = systems(s.Wars)
	s.Cold, s.Quiet = cold(byPair, stray, worldsAt, ended, step)

	// Proxy wars: two vassals of different masters, a master's ships in
	// it, the masters themselves not at war over its span.
	for _, w := range s.Wars {
		m0, m1 := w.Masters[0], w.Masters[1]
		if !w.Vassal[0] || !w.Vassal[1] || m0 == m1 {
			continue
		}
		s.VassalWars++
		if !mastersIn(st, w) {
			continue
		}
		s.MasterIn++
		atWar := false
		for _, o := range byPair[pairOf(m0, m1)] {
			if o.Began <= w.End && w.Began <= o.End {
				atWar = true
			}
		}
		if !atWar {
			s.Proxy++
		}
	}

	s.Waves = waves(taken)
	s.Bonds = bonds(changes, born, ridden, met, warBound, ended)

	// Attacks on vassals: the declarer against the patron's size, and
	// what the patron did while it lasted.
	for _, w := range s.Wars {
		if !w.Vassal[1] {
			continue
		}
		p, by := w.Masters[1], w.Sides[0]
		a := Attack{By: "rival", Patron: "stayed out"}
		switch {
		case by == p:
			a.By = "patron"
		case w.Worlds[0] < worldsAt(p, w.Began):
			a.By = "smaller"
		}
		if a.By != "patron" {
			for _, o := range byPair[pairOf(p, by)] {
				if o.Began >= w.Began && o.Began <= w.End {
					a.Patron = "joined"
				}
			}
			if a.Patron == "stayed out" && shipsFor(st, p, w.Sides[1], by, w) {
				a.Patron = "sent ships"
			}
		}
		s.Attacks = append(s.Attacks, a)
	}
	return s
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// systems joins wars that share a people and are open at once, and reads
// each set at its widest: the moment most peoples were at war in it.
func systems(ws []*War) []System {
	parent := make([]int, len(ws))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(i int) int {
		if parent[i] != i {
			parent[i] = find(parent[i])
		}
		return parent[i]
	}
	byPeople := map[int][]int{}
	for i, w := range ws {
		for _, c := range w.Sides {
			byPeople[c] = append(byPeople[c], i)
		}
	}
	for _, is := range byPeople {
		for x := range is {
			for y := x + 1; y < len(is); y++ {
				a, b := ws[is[x]], ws[is[y]]
				if a.Began <= b.End && b.Began <= a.End {
					parent[find(is[x])] = find(is[y])
				}
			}
		}
	}
	groups := map[int][]*War{}
	for i := range ws {
		groups[find(i)] = append(groups[find(i)], ws[i])
	}
	var out []System
	for _, g := range groups {
		sys := System{Wars: len(g)}
		for _, t := range g {
			peoples := map[int]bool{}
			fronts := 0
			for _, w := range g {
				if w.Began <= t.Began && t.Began <= w.End {
					peoples[w.Sides[0]], peoples[w.Sides[1]] = true, true
					if w.Fought() {
						fronts++
					}
				}
			}
			if len(peoples) > sys.Peak || len(peoples) == sys.Peak && fronts > sys.Fronts {
				sys.Peak, sys.Fronts = len(peoples), fronts
			}
		}
		out = append(out, sys)
	}
	sortSystems(out)
	return out
}

// sortSystems puts the widest first, by peoples, then fronts, then wars,
// so a tie reads the same every time.
func sortSystems(sys []System) {
	sort.Slice(sys, func(i, j int) bool {
		a, b := sys[i], sys[j]
		if a.Peak != b.Peak {
			return a.Peak > b.Peak
		}
		if a.Fronts != b.Fronts {
			return a.Fronts > b.Fronts
		}
		return a.Wars > b.Wars
	})
}

// cold reads the pairs that have fought, between their long wars: a
// stretch of fifty thousand years or more with both large and no war
// longer than a short one between them. With an incident in it — a short
// war, or a battle with no war — it is a cold war; without, a quiet one.
// The build-up the proposal asks of a cold war is not in the record
// until stage 3 gives a people a rival it watches.
func cold(byPair map[pair][]*War, stray map[pair][]record.Year, worldsAt func(int, record.Year) int, ended func(int) record.Year, step record.Year) (int, int) {
	coldN, quiet := 0, 0
	for k, ws := range byPair {
		var incidents []record.Year
		for _, y := range stray[k] {
			incidents = append(incidents, y)
		}
		type gap struct{ from, to record.Year }
		var gaps []gap
		from := record.Year(-1)
		for _, w := range ws {
			if w.Ticks <= shortTicks {
				incidents = append(incidents, w.Began)
				if from < 0 {
					from = w.End
				}
				continue
			}
			if from >= 0 {
				gaps = append(gaps, gap{from, w.Began})
			}
			from = w.End
		}
		if last := min(ended(k[0]), ended(k[1])); from >= 0 && last > from {
			gaps = append(gaps, gap{from, last})
		}
		slices.Sort(incidents)
		isCold, isQuiet := false, false
		for _, g := range gaps {
			run := record.Year(-1)
			for y := g.from; y <= g.to; y += step {
				both := worldsAt(k[0], y) >= large && worldsAt(k[1], y) >= large
				if both && run < 0 {
					run = y
				}
				if (!both || y+step > g.to) && run >= 0 {
					end := y
					if run+coldSpan <= end {
						i := sort.Search(len(incidents), func(i int) bool { return incidents[i] > run })
						if i < len(incidents) && incidents[i] < end {
							isCold = true
						} else {
							isQuiet = true
						}
					}
					run = -1
				}
			}
		}
		if isCold {
			coldN++
		} else if isQuiet {
			quiet++
		}
	}
	return coldN, quiet
}

// mastersIn says whether either side's master had ships in the war: a
// battle of its own against the other side, or a fleet sent with either.
func mastersIn(st *record.State, w *War) bool {
	for i := range 2 {
		if shipsFor(st, w.Masters[i], w.Sides[i], w.Sides[1-i], w) {
			return true
		}
	}
	return false
}

// shipsFor says whether the patron's ships were in the war on its
// client's side: a battle against the enemy, or a fleet launched while it
// lasted to stand with the client or to strike the enemy.
func shipsFor(st *record.State, patron, client, enemy int, w *War) bool {
	for _, b := range st.Battles {
		if b.Year >= w.Began && b.Year <= w.End && (b.Attacker == patron && b.Defender == enemy || b.Attacker == enemy && b.Defender == patron) {
			return true
		}
	}
	for _, x := range st.Fleets {
		if x.Owner == patron && x.Launched >= w.Began && x.Launched <= w.End && (x.Target == enemy || x.Kind == "relief" && x.Target == client) {
			return true
		}
	}
	return false
}

// waves finds, for each people, the most peoples it took worlds from in
// war within one span; those at the wave's size or over are returned.
func waves(taken []*record.Event) []int {
	by := map[int][]*record.Event{}
	for _, e := range taken {
		if e.Object >= 0 {
			by[e.Subject] = append(by[e.Subject], e)
		}
	}
	var out []int
	for _, es := range by {
		sort.SliceStable(es, func(i, j int) bool { return es[i].Year < es[j].Year })
		best := 0
		for i := range es {
			from := map[int]bool{}
			for j := i; j < len(es) && es[j].Year-es[i].Year <= waveSpan; j++ {
				from[es[j].Object] = true
			}
			best = max(best, len(from))
		}
		if best >= waveFrom {
			out = append(out, best)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(out)))
	return out
}

// bonds reads every time a people came under a master: how, by what else
// the record holds that year (a war between the two ending, a meeting,
// the held people's birth or uplift, a rider taking it), and how long it
// lasted.
func bonds(changes map[int][]change, born map[int]record.Year, ridden map[int][]record.Year, met, warBound map[pair][]record.Year, ended func(int) record.Year) []Bond {
	var out []Bond
	for c, cs := range changes {
		for i, x := range cs {
			if x.master < 0 {
				continue
			}
			end := ended(c)
			if i+1 < len(cs) {
				end = cs[i+1].y
			}
			b := Bond{Vassal: x.vassal, Years: end - x.y, Origin: "other"}
			k := pairOf(c, x.master)
			switch {
			case slices.Contains(warBound[k], x.y):
				b.Origin = "war"
			case slices.Contains(ridden[c], x.y):
				b.Origin = "ridden"
			case born[c] == x.y:
				b.Origin = "born"
			case slices.Contains(met[k], x.y):
				b.Origin = "meeting"
			}
			out = append(out, b)
		}
	}
	return out
}
