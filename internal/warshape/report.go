package warshape

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Report is the batch read: a row a seed, then each pattern over the
// batch, the vassal's lot and the health, as markdown lines.
func Report(ss []*Shape) []string {
	var out []string
	add := func(f string, a ...any) { out = append(out, fmt.Sprintf(f, a...)) }

	add("| seed | peoples | wars | fleet wars | fought | short | long | long fought/carried | old enemies | loops | widest | world wars | border | cold | quiet | proxy | waves | first/pair | wars/kpt | index | Myr | took |")
	add("|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|")
	for _, s := range ss {
		f := fought(s.Wars)
		sh, lg, lf := lengths(f)
		old, loops := enemies(s)
		wide := System{}
		if len(s.Systems) > 0 {
			wide = s.Systems[0]
		}
		b := border(s.Wars)
		capped := ""
		if s.Capped {
			capped = " capped"
		}
		add("| %d | %d | %d | %d | %d | %s | %s | %s | %d | %d | %d/%d | %d | %s | %d | %d | %d/%d/%d | %d | %.3f | %.2f | %.2f | %.0f%s | %s |",
			s.Seed, s.Peoples, len(s.Wars), len(fleetWars(s.Wars)), len(f), pct(sh, len(f)), pct(lg, len(f)), frac(lf)+"/"+frac(longCarried(f)), old, loops, wide.Peak, wide.Fronts, worldWars(s.Systems),
			pct(b.border, b.large), s.Cold, s.Quiet, s.VassalWars, s.MasterIn, s.Proxy, len(s.Waves), ratio(s.First, s.Pairs), ratio(1000*len(s.Wars), s.PeopleTicks),
			s.Decline.Index, s.Ages, capped, s.Took.Round(time.Second))
	}
	add("")
	add("Fleet wars: wars between two peoples that send fleets; fought: wars with a battle. Short: fought wars of %d ticks or under; long: %d or over; long fought/carried: the median share of a long war's ticks with a battle, and with a battle or a campaign in flight. Old enemies: pairs with %d fought wars or more; loops: pairs with more than %d wars. Widest: the most peoples at war at once in one system of wars, and its fronts with a battle; world wars: systems of %d peoples or more at once on %d fought fronts or more. Border: of the wars between realms of %d worlds or more, the share over within %d ticks with a world or two taken. Cold: pairs that have fought, both large, at peace %d kyr or more with an incident (a short war or a battle with no war); quiet: as long with none. Proxy: wars between vassals of different masters / a master's ships in them / the masters not at war. Waves: peoples that took worlds from %d peoples or more within %d kyr. Wars/kpt: wars per thousand people-ticks.",
		shortTicks, longTicks, oldWars, loopWars, worldPeak, worldFront, large, shortTicks, coldSpan/1000, waveFrom, waveSpan/1000)

	var all []*War
	var sys []System
	var bonds []Bond
	var attacks []Attack
	var wv []int
	stray, pairs, first, pt, cold, quiet, vw, mi, px, tr := 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
	var pw, pf []int
	for _, s := range ss {
		all = append(all, s.Wars...)
		sys = append(sys, s.Systems...)
		bonds = append(bonds, s.Bonds...)
		attacks = append(attacks, s.Attacks...)
		wv = append(wv, s.Waves...)
		pw = append(pw, s.PairWars...)
		pf = append(pf, s.PairFought...)
		stray += s.Stray
		pairs += s.Pairs
		first += s.First
		pt += s.PeopleTicks
		cold += s.Cold
		quiet += s.Quiet
		vw += s.VassalWars
		mi += s.MasterIn
		px += s.Proxy
		tr += s.Tributes
	}
	sortSystems(sys)
	f := fought(all)
	seeds := func(ok func(*Shape) bool) int {
		n := 0
		for _, s := range ss {
			if ok(s) {
				n++
			}
		}
		return n
	}

	add("")
	add("**1 Short and long.** %d wars, %d of them between two peoples that send fleets; %d fought (%s of those, %s of all). Fought wars by length in ticks: %s. Short %s, long %s; the median long war has battles in %s of its ticks, and is carried (a battle, or a campaign in flight) in %s.",
		len(all), len(fleetWars(all)), len(f), pct(len(f), len(fleetWars(all))), pct(len(f), len(all)), quartiles(ints(f, func(w *War) int { return w.Ticks })), pct(countW(f, func(w *War) bool { return w.Ticks <= shortTicks }), len(f)),
		pct(countW(f, func(w *War) bool { return w.Ticks >= longTicks }), len(f)), frac(longFought(f)), frac(longCarried(f)))
	add("Every war by length: %s; unfought wars: %s.", quartiles(ints(all, func(w *War) int { return w.Ticks })), quartiles(ints(unfought(all), func(w *War) int { return w.Ticks })))

	old, loops := 0, 0
	most := 0
	for i := range pw {
		if pf[i] >= oldWars && pw[i] <= loopWars {
			old++
		}
		if pw[i] > loopWars {
			loops++
		}
		most = max(most, pw[i])
	}
	add("**2 Old enemies.** %d pairs have fought a war; %d have three fought wars or more (not loops), in %d of %d seeds; %d loop pairs; the most wars between one pair %d. Pairs by wars: %s.",
		len(pw), old, seeds(func(s *Shape) bool { o, _ := enemies(s); return o > 0 }), len(ss), loops, most, histogram(pw, []int{1, 2, 3, 5, 10, 21}))

	peaks := ints2(sys, func(x System) int { return x.Peak })
	add("**3 World wars.** %d systems of wars; peoples at war at once in a system: %s; %d world wars, in %d of %d seeds. The widest: %s.",
		len(sys), histogram(peaks, []int{2, 3, 4, 6, 8, 12}), worldWars(sys), seeds(func(s *Shape) bool { return worldWars(s.Systems) > 0 }), len(ss), widest(sys, 5))

	pacts, pmax, sep, absent, relief := 0, 0, 0, 0, 0
	for _, s := range ss {
		pacts += s.Pacts
		pmax = max(pmax, s.PactMax)
		sep += s.Separate
		absent += s.Absent
		relief += s.Relief
	}
	joined := countW(all, func(w *War) bool { return w.Principal >= 0 })
	jf := fought(filterW(all, func(w *War) bool { return w.Principal >= 0 }))
	jc := countW(all, func(w *War) bool { return w.Principal >= 0 && w.Campaigns[0] > 0 })
	add("Allies: %d pacts, the largest %d members at the end; %d wars joined by pact (%s of wars), %s of them fought, the ally sending a campaign in %s; they ended: %s. Separate peaces %d; allies called that did not come %d; relief fleets %d.",
		pacts, pmax, joined, pct(joined, len(all)), pct(len(jf), joined), pct(jc, joined), results(filterW(all, func(w *War) bool { return w.Principal >= 0 })), sep, absent, relief)
	add("Rivalries: wars by their place between the pair, and their aims: %s.", rivalries(all))

	b := border(all)
	add("**4 Border disputes.** %d wars between realms of %d worlds or more: %s. Those for a world (the aim, or before the record held it the cause \"a border\"): %d.", b.large, large, b.kinds(), b.cause)

	dark := 0
	for _, s := range ss {
		dark += s.Dark
	}
	built := 0
	for _, s := range ss {
		built += s.ColdBuilt
	}
	add("**5 Cold wars.** %d pairs with a cold war, in %d of %d seeds; %d of them with the build-up (each naming the other its rival in the stretch), in %d seeds; %d quiet pairs. Battles fought in no war between their two: %d; fleets meeting in the dark in no war: %d.",
		cold, seeds(func(s *Shape) bool { return s.Cold > 0 }), len(ss), built, seeds(func(s *Shape) bool { return s.ColdBuilt > 0 }), quiet, stray, dark)

	add("**6 Proxy wars.** %d wars between vassals of different masters; %d with a master's ships in them; %d proxy wars, the masters not at war.", vw, mi, px)

	add("**7 Conquest waves.** %d peoples took worlds from %d peoples or more within %d kyr, in %d of %d seeds; the widest took from %s.",
		len(wv), waveFrom, waveSpan/1000, seeds(func(s *Shape) bool { return len(s.Waves) > 0 }), len(ss), top(wv, 5))

	add("")
	var rates []Rate
	up, down, owed, short, often, owing := 0, 0, 0, 0, 0, 0
	for _, s := range ss {
		rates = append(rates, s.Rates...)
		up += s.RateUp
		down += s.RateDown
		owed += s.OwedTicks
		short += s.ShortTicks
		often += s.ShortOften
		owing += s.Owing
	}
	add("**Vassalage.** Bonds begun, by how, vassal and slave: %s. Median years held: vassals %s, slaves %s. Attacks on vassals: %d, by who attacked and what the patron did: %s. Tribute facts (a yielding people paying in a commodity): %d.",
		origins(bonds), medianYears(bonds, true), medianYears(bonds, false), len(attacks), attackKinds(attacks), tr)
	ab, cw := 0, 0
	for _, s := range ss {
		ab += s.Abandoned
		cw += s.ClientWars
	}
	add("Patrons called by a client struck: joined %d wars for it; abandoned it %d times (the betrayal).", cw, ab)
	add("Standing tributes: %d bonds, the rate %s; by how the bond came about: %s. Reviews raised a rate %d times and lowered one %d. Of %d vassals owing, %d paid short in more than a tenth of their ticks; ticks paid short %d of %d owed (%s).",
		len(rates), quartilesF(rateOf(rates, "")), rateHows(rates), up, down, owing, often, short, owed, pct(short, owed))

	add("")
	c0 := countW(f, func(w *War) bool { return w.Campaigns[1] > 0 })
	none := countW(all, func(w *War) bool { return w.Campaigns[0]+w.Campaigns[1] == 0 })
	add("**Health.** Campaigns a war: the declarer %.2f, the side declared on %.2f; the side declared on struck back in %s of fought wars; %s of wars had no campaign. Battles a war: mean %.2f, %s. Worlds taken a war: mean %.2f. First wars per distinct pair %.3f; wars per thousand people-ticks %.2f.",
		mean(all, func(w *War) float64 { return float64(w.Campaigns[0]) }), mean(all, func(w *War) float64 { return float64(w.Campaigns[1]) }), pct(c0, len(f)), pct(none, len(all)),
		mean(all, func(w *War) float64 { return float64(w.Battles) }), quartiles(ints(all, func(w *War) int { return w.Battles })), mean(all, func(w *War) float64 { return float64(w.Taken) }),
		ratio(first, pairs), ratio(1000*len(all), pt))
	add("How wars ended: %s.", results(all))
	terms := map[string]int{}
	nt := 0
	for _, s := range ss {
		for k, v := range s.Terms {
			terms[k] += v
			nt += v
		}
	}
	add("Terms by what was given: %s. Fought wars by aim, the share long and short: %s.", counts(terms, nt), byAim(f))
	var took time.Duration
	capped := 0
	for _, s := range ss {
		took += s.Took
		if s.Capped {
			capped++
		}
	}
	add("Ages: %s Myr; decline index at the present %s, at the waning %s; %d of %d capped; runs %s in all.",
		quartilesF(floats(ss, func(s *Shape) float64 { return s.Ages })), quartilesF(floats(ss, func(s *Shape) float64 { return s.Decline.Index })),
		quartilesF(floats(ss, func(s *Shape) float64 { return s.Decline.AtWaning })), capped, len(ss), took.Round(time.Second))
	return out
}

func fought(ws []*War) []*War {
	var out []*War
	for _, w := range ws {
		if w.Fought() {
			out = append(out, w)
		}
	}
	return out
}

// fleetWars is the wars between two peoples that send fleets: the rest
// are a living world's, a sleeper's or an unmaking's blow at a people
// that cannot answer it with a fleet, and no battle is fought in them.
func fleetWars(ws []*War) []*War {
	var out []*War
	for _, w := range ws {
		if w.Fleets {
			out = append(out, w)
		}
	}
	return out
}

func filterW(ws []*War, ok func(*War) bool) []*War {
	var out []*War
	for _, w := range ws {
		if ok(w) {
			out = append(out, w)
		}
	}
	return out
}

// rivalries is the wars by their place between the pair (first, second,
// third and later), each with the share fought and its aims: the
// rivalry's escalation, read.
func rivalries(ws []*War) string {
	var parts []string
	for _, band := range []struct {
		name   string
		lo, hi int
	}{{"first", 1, 1}, {"second", 2, 2}, {"third and later", 3, 1 << 30}} {
		b := filterW(ws, func(w *War) bool { return w.Nth >= band.lo && w.Nth <= band.hi && w.Principal < 0 })
		m := map[string]int{}
		for _, w := range b {
			m[w.Aim]++
		}
		parts = append(parts, fmt.Sprintf("%s %d, fought %s, for %s", band.name, len(b), pct(len(fought(b)), len(b)), counts(m, len(b))))
	}
	return strings.Join(parts, "; ")
}

func unfought(ws []*War) []*War {
	var out []*War
	for _, w := range ws {
		if !w.Fought() {
			out = append(out, w)
		}
	}
	return out
}

func countW(ws []*War, ok func(*War) bool) int {
	n := 0
	for _, w := range ws {
		if ok(w) {
			n++
		}
	}
	return n
}

// lengths is the short and long counts of fought wars, and the median
// share of a long war's ticks with a battle in them.
func lengths(f []*War) (int, int, float64) {
	return countW(f, func(w *War) bool { return w.Ticks <= shortTicks }), countW(f, func(w *War) bool { return w.Ticks >= longTicks }), longFought(f)
}

func longFought(f []*War) float64 {
	var xs []float64
	for _, w := range f {
		if w.Ticks >= longTicks {
			xs = append(xs, float64(w.BattleTicks)/float64(w.Ticks))
		}
	}
	return median(xs)
}

// longCarried is the median share of a long fought war's ticks with a
// battle or a campaign in flight: the war carried, a season a tick. The
// gate reads this (settled 2026-09-24: a fleet on its way is the war
// being fought, not the stare).
func longCarried(f []*War) float64 {
	var xs []float64
	for _, w := range f {
		if w.Ticks >= longTicks {
			xs = append(xs, float64(w.CarriedTicks)/float64(w.Ticks))
		}
	}
	return median(xs)
}

func enemies(s *Shape) (int, int) {
	old, loops := 0, 0
	for i := range s.PairWars {
		if s.PairWars[i] > loopWars {
			loops++
		} else if s.PairFought[i] >= oldWars {
			old++
		}
	}
	return old, loops
}

func worldWars(sys []System) int {
	n := 0
	for _, x := range sys {
		if x.Peak >= worldPeak && x.Fronts >= worldFront {
			n++
		}
	}
	return n
}

func widest(sys []System, n int) string {
	var parts []string
	for i, x := range sys {
		if i == n {
			break
		}
		parts = append(parts, fmt.Sprintf("%d peoples on %d fought fronts (%d wars)", x.Peak, x.Fronts, x.Wars))
	}
	return strings.Join(parts, "; ")
}

// borders is the wars between large realms by kind.
type borders struct {
	large, border, cause int
	kind                 map[string]int
}

func border(ws []*War) borders {
	b := borders{kind: map[string]int{}}
	for _, w := range ws {
		if w.Worlds[0] < large || w.Worlds[1] < large {
			continue
		}
		b.large++
		if w.Aim == "world" || (w.Aim == "" && w.Cause == "border") {
			b.cause++
		}
		k := "long"
		switch {
		case w.Ticks > shortTicks:
		case w.Taken == 0 && !w.Fought():
			k = "short, unfought"
		case w.Taken == 0:
			k = "short, nothing taken"
		case w.Taken <= 2:
			k = "border: short, a world or two taken"
			b.border++
		default:
			k = "short, more taken"
		}
		b.kind[k]++
	}
	return b
}

func (b borders) kinds() string {
	return counts(b.kind, b.large)
}

// byAim is the fought wars by aim, each with its share long and short.
func byAim(f []*War) string {
	var parts []string
	for _, aim := range []string{"world", "tribute", "redress", "submission", "ending", "defence", "hold", ""} {
		ws := filterW(f, func(w *War) bool { return w.Aim == aim })
		if len(ws) == 0 {
			continue
		}
		name := aim
		if name == "" {
			name = "(none)"
		}
		parts = append(parts, fmt.Sprintf("%s %d, long %s, short %s", name, len(ws), pct(countW(ws, func(w *War) bool { return w.Ticks >= longTicks }), len(ws)), pct(countW(ws, func(w *War) bool { return w.Ticks <= shortTicks }), len(ws))))
	}
	return strings.Join(parts, "; ")
}

func rateOf(rs []Rate, how string) []float64 {
	var out []float64
	for _, r := range rs {
		if how == "" || r.How == how {
			out = append(out, r.Rate)
		}
	}
	return out
}

// rateHows is the median rate by how the bond came about.
func rateHows(rs []Rate) string {
	var parts []string
	for _, how := range []string{"surrender", "offer", "sought"} {
		xs := rateOf(rs, how)
		if len(xs) == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %.2f (%d)", how, median(xs), len(xs)))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

func origins(bs []Bond) string {
	m := map[string]int{}
	for _, b := range bs {
		as := "slave"
		if b.Vassal {
			as = "vassal"
		}
		m[as+" by "+b.Origin]++
	}
	return counts(m, len(bs))
}

func medianYears(bs []Bond, vassal bool) string {
	var xs []float64
	for _, b := range bs {
		if b.Vassal == vassal {
			xs = append(xs, float64(b.Years)/1000)
		}
	}
	if len(xs) == 0 {
		return "-"
	}
	return fmt.Sprintf("%.0f kyr (%d)", median(xs), len(xs))
}

func attackKinds(as []Attack) string {
	m := map[string]int{}
	for _, a := range as {
		if a.By == "patron" {
			m["by the patron"]++
			continue
		}
		m[a.By+", the patron "+a.Patron]++
	}
	return counts(m, len(as))
}

func results(ws []*War) string {
	m := map[string]int{}
	for _, w := range ws {
		r := w.Result
		if r == "" {
			r = "(none)"
		}
		m[r]++
	}
	return counts(m, len(ws))
}

// counts prints a tally, the largest first.
func counts(m map[string]int, of int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if m[keys[i]] != m[keys[j]] {
			return m[keys[i]] > m[keys[j]]
		}
		return keys[i] < keys[j]
	})
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %d (%s)", k, m[k], pct(m[k], of)))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

// histogram counts xs into the bands that begin at each edge.
func histogram(xs []int, edges []int) string {
	var parts []string
	for i, e := range edges {
		n := 0
		for _, x := range xs {
			if x >= e && (i+1 == len(edges) || x < edges[i+1]) {
				n++
			}
		}
		label := fmt.Sprintf("%d+", e)
		if i+1 < len(edges) {
			if edges[i+1]-1 == e {
				label = fmt.Sprint(e)
			} else {
				label = fmt.Sprintf("%d-%d", e, edges[i+1]-1)
			}
		}
		parts = append(parts, fmt.Sprintf("%s: %d", label, n))
	}
	return strings.Join(parts, ", ")
}

func top(xs []int, n int) string {
	xs = append([]int(nil), xs...)
	sort.Sort(sort.Reverse(sort.IntSlice(xs)))
	var parts []string
	for i, x := range xs {
		if i == n {
			break
		}
		parts = append(parts, fmt.Sprint(x))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ") + " peoples"
}

func ints(ws []*War, f func(*War) int) []int {
	out := make([]int, len(ws))
	for i, w := range ws {
		out[i] = f(w)
	}
	return out
}

func ints2(sys []System, f func(System) int) []int {
	out := make([]int, len(sys))
	for i, x := range sys {
		out[i] = f(x)
	}
	return out
}

func floats(ss []*Shape, f func(*Shape) float64) []float64 {
	out := make([]float64, len(ss))
	for i, s := range ss {
		out[i] = f(s)
	}
	return out
}

func mean(ws []*War, f func(*War) float64) float64 {
	if len(ws) == 0 {
		return 0
	}
	t := 0.0
	for _, w := range ws {
		t += f(w)
	}
	return t / float64(len(ws))
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	return s[len(s)/2]
}

// quartiles prints the lower quartile, the median and the upper.
func quartiles(xs []int) string {
	if len(xs) == 0 {
		return "none"
	}
	s := append([]int(nil), xs...)
	sort.Ints(s)
	return fmt.Sprintf("%d / %d / %d (max %d)", s[len(s)/4], s[len(s)/2], s[3*len(s)/4], s[len(s)-1])
}

func quartilesF(xs []float64) string {
	if len(xs) == 0 {
		return "none"
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	return fmt.Sprintf("%.2f / %.2f / %.2f", s[len(s)/4], s[len(s)/2], s[3*len(s)/4])
}

func pct(n, of int) string {
	if of == 0 {
		return "-"
	}
	return fmt.Sprintf("%.0f%%", 100*float64(n)/float64(of))
}

func frac(x float64) string { return fmt.Sprintf("%.0f%%", 100*x) }

func ratio(n, of int) float64 {
	if of == 0 {
		return 0
	}
	return float64(n) / float64(of)
}
