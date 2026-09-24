package history

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// A watch is a set of detectors run inside Generate, once a tick through
// Config.Sample, that notice the shapes a batch goes wrong in — a pair of
// peoples at war again and again, a people passed from master to master,
// a war that will not end, a tick in which the war count jumps — and dump
// the stretch that led there for the peoples involved: their councils'
// last reasons (kept in a ring per people through Config.Reason), the
// events touching them, the wars between them. So one batch explains its
// own anomalies, and the next step is a scenario (scenario.go) built from
// the dump. A watch draws and writes nothing in the world: a watched run
// is the same history as an unwatched one (TestWatchSameHistory).

// Watch is the detectors' thresholds, and their state for one world.
type Watch struct {
	PairWars      int       // a pair past this many wars, then each doubling; 0 is off
	MasterChanges int       // a people's master changing more than this many times, then each doubling, not counting a master's heirs; 0 is off
	WarTicks      int       // a war open more than this many ticks; 0 is off
	WarJump       int       // more than this many wars declared in one tick; 0 is off
	Keep          int       // reasons kept per people
	Window        Year      // how far back a dump reads the events
	Lines         int       // the most event lines a dump prints
	All           bool      // keep every reason, not only the war's and the council's
	Out           io.Writer // where the dumps go
	Fired         []string  // what fired, one line each, in order
	// Stare is, over the long wars (twelve ticks or more) that saw a
	// battle, what each of their ticks was: fought, a fleet in flight,
	// or the two sides' last verdicts in a word (mind.WarVerdict.Key),
	// declarer's first. It is what the gate's "long war fought ticks"
	// row is made of.
	Stare map[string]int

	teller
	rings     map[int]*reasonRing
	wars      int                    // the wars read so far
	pairs     map[[2]int]int         // wars per pair
	pairNext  map[[2]int]int         // the count at which a pair fires next
	masters   map[int]int            // each people's master as last seen
	changes   map[int]int            // the changes seen
	chgNext   map[int]int            // the count at which a people fires next
	long      map[int]bool           // the wars already told as long
	ticks     map[int]map[string]int // each open war's ticks so far, by what they were
	firstOpen int                    // no war before this is open
	jumped    Year                   // when the jump last fired
}

// DefaultWatch is the thresholds a batch runs with: a loop pair at ten
// wars (the gate counts one at twenty), a people held by a sixth master,
// a war a hundred ticks long, fifty wars in a tick.
func DefaultWatch(out io.Writer) *Watch {
	return &Watch{PairWars: 10, MasterChanges: 6, WarTicks: 100, WarJump: 50, Keep: 30, Window: 30_000, Lines: 200, Out: out}
}

type reasonRing struct {
	at   []Year
	line []string
	next int
	full bool
}

func (g *reasonRing) add(y Year, s string) {
	g.at[g.next], g.line[g.next] = y, s
	g.next++
	if g.next == len(g.line) {
		g.next, g.full = 0, true
	}
}

// each walks the ring oldest first.
func (g *reasonRing) each(f func(Year, string)) {
	if g.full {
		for i := g.next; i < len(g.line); i++ {
			f(g.at[i], g.line[i])
		}
	}
	for i := 0; i < g.next; i++ {
		f(g.at[i], g.line[i])
	}
}

// Attach sets the watch on a config: the reasons hook, and the per-tick
// sample, run after whatever sample was already set.
func (wt *Watch) Attach(cfg *Config) {
	if wt.Keep == 0 {
		wt.Keep = 30
	}
	if wt.Window == 0 {
		wt.Window = 30_000
	}
	if wt.Lines == 0 {
		wt.Lines = 200
	}
	wt.names = map[int]string{}
	wt.rings = map[int]*reasonRing{}
	wt.pairs, wt.pairNext = map[[2]int]int{}, map[[2]int]int{}
	wt.masters, wt.changes, wt.chgNext = map[int]int{}, map[int]int{}, map[int]int{}
	wt.long = map[int]bool{}
	wt.ticks, wt.Stare = map[int]map[string]int{}, map[string]int{}
	cfg.Reason = func(w *World, c *Civ, what, why string) {
		if !wt.All && !warReason(what) {
			return
		}
		g := wt.rings[c.ID]
		if g == nil {
			g = &reasonRing{at: make([]Year, wt.Keep), line: make([]string, wt.Keep)}
			wt.rings[c.ID] = g
		}
		g.add(w.Now, what+" — "+why)
	}
	prior := cfg.Sample
	cfg.Sample = func(w *World) {
		if prior != nil {
			prior(w)
		}
		wt.sample(w)
	}
}

// sample runs the detectors on the tick just run.
func (wt *Watch) sample(w *World) {
	declared := map[int]int{}
	n := 0
	for ; wt.wars < len(w.Wars); wt.wars++ {
		wr := w.Wars[wt.wars]
		n++
		declared[wr.Sides[0]]++
		k := warKey(wr.Sides[0], wr.Sides[1])
		wt.pairs[k]++
		if wt.PairWars > 0 {
			next := wt.pairNext[k]
			if next == 0 {
				next = wt.PairWars
			}
			if wt.pairs[k] >= next {
				wt.pairNext[k] = 2 * next
				wt.fire(w, sprintf("pair civ%d–civ%d at %d wars", k[0], k[1], wt.pairs[k]), k[:])
			}
		}
	}
	if wt.WarJump > 0 && n > wt.WarJump && (wt.jumped == 0 || w.Now-wt.jumped >= 50_000) {
		wt.jumped = w.Now
		ids := make([]int, 0, len(declared))
		for id := range declared {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool {
			if declared[ids[i]] != declared[ids[j]] {
				return declared[ids[i]] > declared[ids[j]]
			}
			return ids[i] < ids[j]
		})
		wt.fire(w, sprintf("%d wars declared in one tick", n), ids[:min(3, len(ids))])
	}
	if wt.MasterChanges > 0 {
		for _, c := range w.Civs {
			if !c.Living() {
				continue
			}
			was, seen := wt.masters[c.ID]
			wt.masters[c.ID] = c.Master
			if !seen || was == c.Master {
				continue
			}
			if was >= 0 && c.Master >= 0 && w.kin(w.Civs[was], w.Civs[c.Master]) {
				continue // passed to an heir at a sundering, or back: the same realm's
			}
			wt.changes[c.ID]++
			next := wt.chgNext[c.ID]
			if next == 0 {
				next = wt.MasterChanges + 1
			}
			if wt.changes[c.ID] >= next {
				wt.chgNext[c.ID] = 2 * next
				ids := []int{c.ID}
				for _, m := range []int{was, c.Master} {
					if m >= 0 {
						ids = append(ids, m)
					}
				}
				wt.fire(w, sprintf("civ%d has changed master %d times", c.ID, wt.changes[c.ID]), ids)
			}
		}
	}
	wt.stare(w)
	if wt.WarTicks > 0 {
		for i := wt.firstOpen; i < len(w.Wars); i++ {
			wr := w.Wars[i]
			if wr.Over || wt.long[wr.ID] || int((w.Now-wr.Began)/w.Cfg.Step) <= wt.WarTicks {
				continue
			}
			wt.long[wr.ID] = true
			wt.fire(w, sprintf("war %d open %d ticks", wr.ID, (w.Now-wr.Began)/w.Cfg.Step), wr.Sides[:])
		}
	}
}

// stare tallies each open war's tick, and folds a war that has ended
// into Stare if it was long and fought.
func (wt *Watch) stare(w *World) {
	for id, m := range wt.ticks {
		wr := w.Wars[id]
		if !wr.Over {
			continue
		}
		if int((wr.Ended-wr.Began)/w.Cfg.Step) >= 12 && wr.Battles > 0 {
			for k, n := range m {
				wt.Stare[k] += n
			}
		}
		delete(wt.ticks, id)
	}
	for wt.firstOpen < len(w.Wars) && w.Wars[wt.firstOpen].Over {
		wt.firstOpen++
	}
	for i := wt.firstOpen; i < len(w.Wars); i++ {
		wr := w.Wars[i]
		if wr.Over {
			continue
		}
		a, b := w.Civs[wr.Sides[0]], w.Civs[wr.Sides[1]]
		k := "fought"
		switch {
		case wr.Battles > 0 && w.Now-wr.Fought < w.Cfg.Step:
		case w.fleetInFlight(a, b) || w.fleetInFlight(b, a):
			k = "flight"
		case wr.Gap != nil:
			k = "hunt"
		default:
			k = orWord(wr.Why[0]) + "/" + orWord(wr.Why[1])
		}
		if wt.ticks[wr.ID] == nil {
			wt.ticks[wr.ID] = map[string]int{}
		}
		wt.ticks[wr.ID][k]++
	}
}

func orWord(s string) string {
	if s == "" {
		return "unsat" // no council has sat on it yet
	}
	return s
}

// StareReport is the Stare tallies of several watches, most first, as
// shares of the ticks, the first n.
func StareReport(ws []*Watch, n int) string {
	all := map[string]int{}
	total := 0
	for _, wt := range ws {
		if wt == nil {
			continue
		}
		for k, v := range wt.Stare {
			all[k] += v
			total += v
		}
	}
	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if all[keys[i]] != all[keys[j]] {
			return all[keys[i]] > all[keys[j]]
		}
		return keys[i] < keys[j]
	})
	var parts []string
	for _, k := range keys[:min(n, len(keys))] {
		parts = append(parts, sprintf("%s %.0f%%", k, 100*float64(all[k])/float64(max(1, total))))
	}
	return sprintf("the ticks of long fought wars (%d): %s", total, strings.Join(parts, ", "))
}

// fire dumps the stretch for the peoples involved.
func (wt *Watch) fire(w *World, what string, ids []int) {
	head := sprintf("seed %d, year %d: %s", w.Seed, w.Now, what)
	wt.Fired = append(wt.Fired, head)
	if wt.Out == nil {
		return
	}
	var b strings.Builder
	p := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }
	p("=== watch: %s", head)
	set := map[int]bool{}
	for _, id := range ids {
		set[id] = true
		c := w.Civs[id]
		s := sprintf("  civ%d: %s, %d worlds, %d ships, mil %.1f reach %.0f, %s", id, stageWord(c), len(c.Systems), w.ships(c), c.Mil, c.Reach, c.posture())
		if c.Leader != nil {
			s += sprintf(" (leader %s %s)", c.Leader.Form, c.Leader.Stance)
		}
		if !c.Free() {
			s += sprintf(", held by civ%d (vassal %v)", c.Master, c.Vassal)
		}
		if c.Own >= 0 {
			s += ", a rider"
		}
		s += " — traits " + strings.Join(c.Species.TraitKeys(), " ")
		p("%s", s)
	}
	for _, a := range ids {
		for _, e := range ids {
			if a >= e {
				continue
			}
			ca := w.Civs[a]
			p("  civ%d toward civ%d: grudge %.2f/%.2f, fought %d, wary %.1f/%.1f, fathomed %v/%v, truce until %d/%d, claims %d/%d", a, e,
				ca.Grudge[e], w.Civs[e].Grudge[a], ca.Fought[e], ca.Wary[e], w.Civs[e].Wary[a], ca.Fathomed[e], w.Civs[e].Fathomed[a],
				ca.Truce[e], w.Civs[e].Truce[a], claimsOn(ca, w.Civs[e]), claimsOn(w.Civs[e], ca))
		}
	}
	var wars []*War
	for _, wr := range w.Wars {
		if set[wr.Sides[0]] && set[wr.Sides[1]] {
			wars = append(wars, wr)
		}
	}
	if len(wars) > 0 {
		p("  wars between them: %d; the last %d:", len(wars), min(8, len(wars)))
		for _, wr := range wars[max(0, len(wars)-8):] {
			end := "open"
			if wr.Over {
				end = sprintf("%s at %d (%d ticks)", wr.Result, wr.Ended, (wr.Ended-wr.Began)/w.Cfg.Step)
			}
			p("    war %d civ%d→civ%d at %d, %s for %s, %s; will %.2f/%.2f, %s/%s, %d battles, taken %d/%d", wr.ID, wr.Sides[0], wr.Sides[1], wr.Began, wr.Cause, wr.Aim, end,
				wr.Will[0], wr.Will[1], wr.Verdict[0], wr.Verdict[1], wr.Battles, wr.Taken[0]+wr.Glassed[0], wr.Taken[1]+wr.Glassed[1])
		}
	}
	for _, id := range ids {
		g := wt.rings[id]
		if g == nil {
			continue
		}
		p("  the last reasons of civ%d:", id)
		g.each(func(y Year, s string) { p("    %d %s", y, wt.untoken(s)) })
	}
	from := w.Now - wt.Window - w.Cfg.Step
	if len(wars) >= 3 {
		from = min(from, wars[len(wars)-3].Began) // the last three wars between them, whenever they were
	}
	var lines []string
	for i := len(w.Events) - 1; i >= 0 && len(lines) < wt.Lines; i-- {
		e := w.Events[i]
		if e.Year < from {
			break
		}
		if (!set[e.Subject] && !set[e.Object]) || quietKinds[e.Kind] || e.Kind == KReason {
			continue
		}
		lines = append(lines, sprintf("    %d %s", e.Year, wt.eventLine(e)))
	}
	p("  the events touching them since %d (%d lines):", from, len(lines))
	for i := len(lines) - 1; i >= 0; i-- {
		p("%s", lines[i])
	}
	fmt.Fprint(wt.Out, b.String())
}

func stageWord(c *Civ) string {
	switch c.Stage {
	case Remnant:
		return "a remnant"
	case Dead:
		return "gone"
	}
	return "active"
}

// claimsOn is how many of e's worlds c claims.
func claimsOn(c, e *Civ) int {
	n := 0
	for _, s := range e.Systems {
		if c.Claim[s] {
			n++
		}
	}
	return n
}
