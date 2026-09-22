package writer

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"worldgen/internal/history"
	"worldgen/internal/record"
)

// The fold test: the chronicle is complete about the durable. A reducer
// folds chronicle.jsonl from the first record and must reproduce the
// durable part of state.json exactly; whatever is durable and changes
// without an event fails here, so no kind can go missing and no consumer
// has to guess. The table below is the line between durable and
// snapshot; FORMAT.md marks the same fields D and S.
//
// The durable fields, and the kinds that carry them:
//
//	star.held            world_held, world_lost
//	star.bio             bio
//	civ (existence)      civ_born: species, cradle, master
//	civ.species          civ_born, blood
//	civ.home             civ_born, seat
//	civ.stage            stage
//	civ.fate, civ.cause  fate
//	civ.systems          world_held, world_lost
//	civ.known            node_held, node_lost
//	civ.scars, civ.boons scar, boon
//	civ.miracles         miracle_held, miracle_lost
//	civ.line             line
//	civ.asleep           sleep
//	civ.master, .vassal  civ_born, master
//	civ.named            named
//	civ.infections       infected, cleared (the keys; the plague's hosts count follows)
//	civ.wars             war_opened, war_over
//	species.powers       civ_born, blood, power_held
//	remain (existence)   remain_left: kind, maker, star, state, cond, finder
//	remain.star          remain_moved
//	remain.state         remain_state
//	remain.cond          remain_cond
//	remain.finder        remain_finder
//	remain.listeners     listened
//	war (existence)      war_opened: sides, cause
//	war.over, war.result war_over
//	plague.hosts         infected, cleared, over the peoples still rising
//
// Everything else in the state is a snapshot: the levels, the flows and
// the dials derived every tick; the wear of the tales; the counters kept
// for the batch; a star's class and system, which are the substrate's
// and not the age's; the peoples' knowledge, which is the tellings' and
// the intel's business.

type foldStar struct {
	held int
	bio  string
}

type foldCiv struct {
	species    int
	cradle     int
	home       int
	stage      string
	fate       string
	cause      string
	systems    map[int]bool
	known      map[string]bool
	scars      map[string]bool
	boons      map[string]bool
	miracles   map[string]string
	line       []int
	asleep     bool
	master     int
	vassal     bool
	named      bool
	infections map[int]bool
	wars       map[int]bool
}

type foldRemain struct {
	kind      string
	maker     int
	star      int
	state     string
	cond      string
	finder    int
	listeners int
}

type foldWar struct {
	sides  [2]int
	cause  string
	over   bool
	result string
}

type fold struct {
	stars   []foldStar
	civs    map[int]*foldCiv
	powers  map[int][]string
	remains map[int]*foldRemain
	wars    map[int]*foldWar
}

func newFold(n int) *fold {
	f := &fold{stars: make([]foldStar, n), civs: map[int]*foldCiv{}, powers: map[int][]string{}, remains: map[int]*foldRemain{}, wars: map[int]*foldWar{}}
	for i := range f.stars {
		f.stars[i] = foldStar{held: -1, bio: "none"}
	}
	return f
}

// apply folds one record.
func (f *fold) apply(e *record.Event) error {
	civ := func() *foldCiv {
		c := f.civs[e.Subject]
		if c == nil {
			return nil
		}
		return c
	}
	remain := func() *foldRemain { return f.remains[e.Legacy] }
	switch e.Kind {
	case record.KCivBorn:
		if f.civs[e.Subject] != nil {
			return fmt.Errorf("people %d born twice", e.Subject)
		}
		c := &foldCiv{species: e.Int("species"), cradle: e.Star, home: e.Star, stage: "emergent", fate: "active", master: e.Int("master"),
			systems: map[int]bool{}, known: map[string]bool{}, scars: map[string]bool{}, boons: map[string]bool{}, miracles: map[string]string{}, infections: map[int]bool{}, wars: map[int]bool{}}
		f.civs[e.Subject] = c
		if _, ok := f.powers[c.species]; !ok {
			f.powers[c.species] = e.Strs("powers")
		}
	case record.KWorldHeld:
		f.stars[e.Star].held = e.Subject
		civ().systems[e.Star] = true
	case record.KWorldLost:
		if f.stars[e.Star].held == e.Subject {
			f.stars[e.Star].held = -1
		}
		delete(civ().systems, e.Star)
	case record.KBio:
		f.stars[e.Star].bio = e.Str("bio")
	case record.KSeat:
		civ().home = e.Star
	case record.KStage:
		civ().stage = e.Str("stage")
	case record.KFate:
		civ().fate, civ().cause = e.Str("fate"), e.Str("cause")
	case record.KNodeHeld:
		civ().known[e.Str("node")] = true
	case record.KNodeLost:
		delete(civ().known, e.Str("node"))
	case record.KScar:
		civ().scars[e.Str("scar")] = true
	case record.KBoon:
		civ().boons[e.Str("boon")] = true
	case record.KMiracleHeld:
		civ().miracles[e.Str("miracle")] = e.Str("route")
	case record.KMiracleLost:
		delete(civ().miracles, e.Str("miracle"))
	case record.KLine:
		civ().line = e.Ints("line")
	case record.KSleep:
		civ().asleep = e.Bool("asleep")
	case record.KMaster:
		civ().master, civ().vassal = e.Int("master"), e.Bool("vassal")
	case record.KNamed:
		civ().named = e.Bool("named")
	case record.KInfected:
		civ().infections[e.Plague] = true
	case record.KCleared:
		delete(civ().infections, e.Plague)
	case record.KBlood:
		c := civ()
		c.species = e.Int("species")
		if _, ok := f.powers[c.species]; !ok {
			f.powers[c.species] = e.Strs("powers")
		}
	case record.KPowerHeld:
		sp := e.Int("species")
		f.powers[sp] = append(f.powers[sp], e.Str("power"))
	case record.KRemainLeft:
		if f.remains[e.Legacy] != nil {
			return fmt.Errorf("remain %d left twice", e.Legacy)
		}
		f.remains[e.Legacy] = &foldRemain{kind: e.Str("kind"), maker: e.Int("maker"), star: e.Star, state: e.Str("state"), cond: e.Str("cond"), finder: e.Int("finder")}
	case record.KRemainState:
		remain().state = e.Str("state")
	case record.KRemainMoved:
		remain().star = e.Star
	case record.KRemainCond:
		remain().cond = e.Str("cond")
	case record.KRemainFinder:
		remain().finder = e.Int("finder")
	case record.KListened:
		remain().listeners++
	case record.KWarOpened:
		id := e.Int("war")
		if f.wars[id] != nil {
			return fmt.Errorf("war %d opened twice", id)
		}
		f.wars[id] = &foldWar{sides: [2]int{e.Subject, e.Object}, cause: e.Str("cause")}
		civ().wars[e.Object] = true
		f.civs[e.Object].wars[e.Subject] = true
	case record.KWarOver:
		wr := f.wars[e.Int("war")]
		wr.over, wr.result = true, e.Str("result")
		delete(civ().wars, e.Object)
		delete(f.civs[e.Object].wars, e.Subject)
	}
	return nil
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func idsOf(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func same[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestFold: the state's durable fields, folded from the chronicle, on
// four seeds; one of them cut short, since a run stopped mid-age must
// fold like any other.
func TestFold(t *testing.T) {
	for _, c := range []struct {
		seed  uint64
		stars int
		until history.Year
	}{{7, 120, 0}, {5, 200, 0}, {3, 200, 0}, {9, 200, 20_000_000}} {
		t.Run(fmt.Sprintf("seed%d", c.seed), func(t *testing.T) { foldOne(t, Run(generate(c.seed, c.stars, c.until), "sol")) })
	}
}

func foldOne(t *testing.T, r *record.Run) {
	f := newFold(len(r.State.Stars))
	for _, e := range r.Chronicle {
		if err := f.apply(e); err != nil {
			t.Fatalf("at event %d (%s, year %d): %v", e.ID, e.Kind, e.Year, err)
		}
	}
	var bad []string
	fail := func(format string, a ...any) {
		if len(bad) < 40 {
			bad = append(bad, fmt.Sprintf(format, a...))
		} else if len(bad) == 40 {
			bad = append(bad, "...")
		}
	}
	st := r.State
	for i, s := range st.Stars {
		if f.stars[i].held != s.Held {
			fail("star %d held: folded %d, state %d", i, f.stars[i].held, s.Held)
		}
		if f.stars[i].bio != s.Bio {
			fail("star %d bio: folded %s, state %s", i, f.stars[i].bio, s.Bio)
		}
	}
	if len(f.civs) != len(st.Civs) {
		fail("%d peoples folded, %d in the state", len(f.civs), len(st.Civs))
	}
	for _, c := range st.Civs {
		g := f.civs[c.ID]
		if g == nil {
			fail("people %d never born", c.ID)
			continue
		}
		check := func(field string, folded, state any) {
			if fmt.Sprint(folded) != fmt.Sprint(state) {
				fail("people %d %s: folded %v, state %v", c.ID, field, folded, state)
			}
		}
		check("species", g.species, c.Species)
		check("cradle", g.cradle, c.Cradle)
		check("home", g.home, c.Home)
		check("stage", g.stage, c.Stage)
		check("fate", g.fate, c.Fate)
		check("cause", g.cause, c.Cause)
		sys := append([]int{}, c.Systems...)
		sort.Ints(sys)
		check("systems", idsOf(g.systems), sys)
		check("known", keysOf(g.known), c.Known)
		check("scars", keysOf(g.scars), c.Scars)
		check("boons", keysOf(g.boons), c.Boons)
		check("miracles", g.miracles, c.Miracles)
		check("line", fmt.Sprint(g.line), fmt.Sprint(c.Line))
		check("asleep", g.asleep, c.Asleep)
		check("master", g.master, c.Master)
		check("vassal", g.vassal, c.Vassal)
		check("named", g.named, c.Named)
		var inf []int
		for id := range c.Infections {
			inf = append(inf, id)
		}
		sort.Ints(inf)
		check("infections", idsOf(g.infections), inf)
		check("wars", idsOf(g.wars), c.Wars)
	}
	for _, sp := range st.Species {
		if !same(f.powers[sp.ID], sp.Powers) {
			fail("species %d powers: folded %v, state %v", sp.ID, f.powers[sp.ID], sp.Powers)
		}
	}
	if len(f.remains) != len(st.Remains) {
		fail("%d remains folded, %d in the state", len(f.remains), len(st.Remains))
	}
	for _, l := range st.Remains {
		g := f.remains[l.ID]
		if g == nil {
			fail("remain %d never left", l.ID)
			continue
		}
		got := fmt.Sprint(g.kind, g.maker, g.star, g.state, g.cond, g.finder, g.listeners)
		want := fmt.Sprint(l.Kind, l.Maker, l.Star, l.State, l.Cond, l.Finder, l.Listeners)
		if got != want {
			fail("remain %d: folded [%s], state [%s]", l.ID, got, want)
		}
	}
	if len(f.wars) != len(st.Wars) {
		fail("%d wars folded, %d in the state", len(f.wars), len(st.Wars))
	}
	for _, wr := range st.Wars {
		g := f.wars[wr.ID]
		if g == nil {
			fail("war %d never opened", wr.ID)
			continue
		}
		if g.sides != wr.Sides || g.cause != wr.Cause || g.over != wr.Over || g.result != wr.Result {
			fail("war %d: folded %+v, state sides %v cause %s over %v result %s", wr.ID, *g, wr.Sides, wr.Cause, wr.Over, wr.Result)
		}
	}
	hosts := map[int]int{}
	for _, c := range st.Civs {
		if c.Stage == "emergent" || c.Stage == "interstellar" || c.Stage == "zenith" {
			for id := range f.civs[c.ID].infections {
				hosts[id]++
			}
		}
	}
	for _, p := range st.Plagues {
		if hosts[p.ID] != p.Hosts {
			fail("plague %d hosts: folded %d, state %d", p.ID, hosts[p.ID], p.Hosts)
		}
	}
	if len(bad) > 0 {
		t.Fatalf("the chronicle does not fold to the state:\n  %s", strings.Join(bad, "\n  "))
	}
}
