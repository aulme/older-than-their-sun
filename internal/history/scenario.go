package history

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	"worldgen/internal/flow"
	"worldgen/internal/mind"
	"worldgen/internal/species"
	"worldgen/internal/tech"
)

// A scenario is a tiny world set up by hand: a few named peoples at
// stars of a small real field, with the ships, levels, relations and wars
// the spec gives them, run through the real tick for a number of ticks
// and told as a narrative of the named peoples only — what each council
// decided and why, the fleets, the battles, the terms, the will. It is
// the debugging tool for a behaviour seen in a batch: build the moment,
// run it, read it, change the rule, run it again. The same spec, kept in
// testdata/scenarios, is a regression test with assertions in Go
// (scenario_test.go); cmd/scenario runs a spec file and prints the tale.
//
// What a spec cannot pin: dials, which the traits set and recompute
// resets every tick, and leaders after the first, who rise as they would. Levels are the
// nodes known plus a boon at the home (a silent rarity that feeds the
// people and adds the military and reach asked for), so a people that
// learns a weapon grows stronger as it would, and a people that loses its
// home loses the boon to whoever took it.

// Scenario is the spec.
type Scenario struct {
	Name    string           `json:"name"`
	Seed    uint64           `json:"seed"`  // the field's and the streams' seed; 1 if unset
	Stars   int              `json:"stars"` // the field; 40 if unset
	Ticks   int              `json:"ticks"` // how long cmd/scenario runs it; 30 if unset
	Noisy   bool             `json:"noisy"` // keep the births, the sky's events and the plagues; a scenario is quiet by default
	Keep    []string         `json:"keep"`  // of those, the phases a quiet scenario keeps: "life", "cosmic", "plagues"
	Tune    []string         `json:"tune"`  // Group.Field=value, as -tune
	Peoples []ScenarioPeople `json:"peoples"`
	Bonds   []ScenarioBond   `json:"bonds"`
	Wars    []ScenarioWar    `json:"wars"`
	Watch   []string         `json:"watch"` // the peoples told of; all the named if unset
}

// ScenarioPeople is one named people.
type ScenarioPeople struct {
	Name   string   `json:"name"`
	Traits []string `json:"traits"` // trait keys, e.g. "conqueror", "pacifist", "faithless"
	Kind   string   `json:"kind"`   // the substrate: biological (default), machine, eldritch, parasite
	Mods   []string `json:"mods"`   // modifiers: planetary, hive, unconscious, replicator, antimemetic, evolver
	Powers []string `json:"powers"` // the eldritch's powers
	Star   int      `json:"star"`   // the home, if set (> 0); else placed near another
	Near   string   `json:"near"`   // placed near this people; the first people if unset
	Dist   float64  `json:"dist"`   // light years from it, the free star nearest this; 10 if unset
	Worlds int      `json:"worlds"` // worlds held, the home among them; 1 if unset
	Ships  *int     `json:"ships"`  // ships in the home's guard; 3 if unset
	Knows  []string `json:"knows"`  // nodes known, each with its prerequisites; relativistic if unset
	Mil    float64  `json:"mil"`    // the military level wanted at the start, if set
	Reach  float64  `json:"reach"`  // the reach wanted at the start, if set
	Spare  *float64 `json:"spare"`  // of each commodity a tick beyond what its nodes and ships take; 2 if unset
	Leader string   `json:"leader"` // a leader with this stance, reigning through the run
	Master string   `json:"master"` // held by this people
	Vassal bool     `json:"vassal"` // held as a vassal, else a slave
}

// ScenarioBond is what two peoples are to each other. Met and fathomed
// are both ways; the rest is A's toward B.
type ScenarioBond struct {
	A       string  `json:"a"`
	B       string  `json:"b"`
	Unmet   bool    `json:"unmet"`   // not met, and so neither fathomed
	Unknown bool    `json:"unknown"` // met, neither fathoming the other
	Grudge  float64 `json:"grudge"`  // A's grudge against B
	Claims  bool    `json:"claims"`  // A claims every world B holds
	Wary    float64 `json:"wary"`    // A's wariness of B
	Fought  int     `json:"fought"`  // wars fought between them before
	Truce   int     `json:"truce"`   // years of truce left between them
	Pact    string  `json:"pact"`    // "defence", "war" or "both": a pact between them
	Trade   bool    `json:"trade"`   // trading partners
}

// ScenarioWar is a war already running at the start.
type ScenarioWar struct {
	By    string     `json:"by"`
	On    string     `json:"on"`
	Cause string     `json:"cause"` // a war cause; border if unset
	Aim   string     `json:"aim"`   // overrides the aim the cause gives
	Will  [2]float64 `json:"will"`  // overrides each side's will, if set
}

// LoadScenario reads a spec from a JSON file.
func LoadScenario(path string) (*Scenario, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Scenario
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &s, nil
}

// Run is a scenario built and running.
type Run struct {
	S   *Scenario
	W   *World
	Out io.Writer // the narrative; nil for none
	All bool      // every reason and debug line, not only the war's and the council's

	teller
	ids     map[string]int
	watched map[int]bool
	Said    []Said      // every reason given in the run, for a test to read
	cursor  int         // the events told so far
	reasons []runReason // this tick's, as they came
	last    map[string]string
	tick    int
}

// Said is one reason a people gave: at which tick, who, about what, why.
type Said struct {
	Tick      int
	Civ       int
	What, Why string
}

type runReason struct {
	at        int // the length of the event log when it was given, to tell it in order
	civ       int
	what, why string
}

// Build makes the world the spec describes, with nothing yet run.
func (s *Scenario) Build(out io.Writer) (*Run, error) {
	if s.Seed == 0 {
		s.Seed = 1
	}
	if s.Stars == 0 {
		s.Stars = 40
	}
	if s.Ticks == 0 {
		s.Ticks = 30
	}
	cfg := DefaultConfig()
	cfg.Stars = s.Stars
	t, err := mind.Configure("", s.Tune)
	if err != nil {
		return nil, err
	}
	cfg.Tuning = t
	r := &Run{S: s, Out: out, teller: teller{names: map[int]string{}}, ids: map[string]int{}, watched: map[int]bool{}, last: map[string]string{}}
	cfg.Reason = func(w *World, c *Civ, what, why string) {
		r.reasons = append(r.reasons, runReason{len(w.Events), c.ID, what, why})
		r.Said = append(r.Said, Said{r.tick, c.ID, what, why})
	}
	w := newWorld(s.Seed, cfg)
	w.dt = float64(cfg.Step) / 1000
	w.Now = cfg.Dawn + cfg.Step // the dawn is year zero, and a year of zero is never
	if !s.Noisy {
		var ps []phase
		for _, p := range w.phases {
			switch p.Name {
			case "life", "cosmic", "plagues":
				if slices.Contains(s.Keep, p.Name) {
					ps = append(ps, p)
				}
			default:
				ps = append(ps, p)
			}
		}
		w.phases = ps
	}
	r.W = w
	if err := r.place(); err != nil {
		return nil, err
	}
	for i := range s.Peoples {
		if err := r.dress(&s.Peoples[i]); err != nil {
			return nil, err
		}
	}
	for _, b := range s.Bonds {
		if err := r.bond(b); err != nil {
			return nil, err
		}
	}
	for _, p := range s.Peoples {
		if p.Master == "" {
			continue
		}
		m, ok := r.ids[p.Master]
		if !ok {
			return nil, fmt.Errorf("%s: no people %q to be held by", p.Name, p.Master)
		}
		if p.Vassal {
			w.vassal(w.Civs[m], r.Civ(p.Name), "offer")
		} else {
			w.enslave(w.Civs[m], r.Civ(p.Name))
		}
	}
	for _, sw := range s.Wars {
		c, e := r.Civ(sw.By), r.Civ(sw.On)
		if c == nil || e == nil {
			return nil, fmt.Errorf("war %s on %s: no such people", sw.By, sw.On)
		}
		cause := sw.Cause
		if cause == "" {
			cause = "border"
		}
		wr := w.declare(c, e, because(cause))
		if wr == nil {
			return nil, fmt.Errorf("war %s on %s: not declared", sw.By, sw.On)
		}
		if sw.Aim != "" {
			wr.Aim = sw.Aim
		}
		for i, v := range sw.Will {
			if v != 0 {
				wr.Will[i] = v
			}
		}
	}
	watch := s.Watch
	if len(watch) == 0 {
		for _, p := range s.Peoples {
			watch = append(watch, p.Name)
		}
	}
	for _, n := range watch {
		id, ok := r.ids[n]
		if !ok {
			return nil, fmt.Errorf("watch: no people %q", n)
		}
		r.watched[id] = true
	}
	r.header()
	r.tell() // what the setting up said
	return r, nil
}

// place puts every people's home: at its star if given, else at the free
// star nearest the distance asked from the people it is placed near.
func (r *Run) place() error {
	w := r.W
	taken := map[int]bool{}
	free := func(s int) bool {
		return !taken[s] && s != w.G.Sol && w.Owner[s] < 0 && !w.G.Stars[s].Dead()
	}
	for i, p := range r.S.Peoples {
		if p.Name == "" {
			return fmt.Errorf("people %d has no name", i)
		}
		if _, dup := r.ids[p.Name]; dup {
			return fmt.Errorf("two peoples named %q", p.Name)
		}
		home := -1
		switch {
		case p.Star > 0:
			if p.Star >= len(w.G.Stars) || !free(p.Star) {
				return fmt.Errorf("%s: star %d is not free", p.Name, p.Star)
			}
			home = p.Star
		case i == 0:
			// the star with the most neighbours within a hop, so there is room
			best := 0
			for s := range w.G.Stars {
				if !free(s) {
					continue
				}
				if n := len(w.G.Near(s, fleetHop)); home < 0 || n > best {
					home, best = s, n
				}
			}
		default:
			near := r.S.Peoples[0].Name
			if p.Near != "" {
				near = p.Near
			}
			nid, ok := r.ids[near]
			if !ok {
				return fmt.Errorf("%s: no people %q placed before it to be near", p.Name, near)
			}
			d := p.Dist
			if d == 0 {
				d = 10
			}
			from := w.Civs[nid].Home
			bd := 0.0
			for s := range w.G.Stars {
				if !free(s) || s == from {
					continue
				}
				if x := math.Abs(w.G.Dist(from, s) - d); home < 0 || x < bd {
					home, bd = s, x
				}
			}
		}
		if home < 0 {
			return fmt.Errorf("%s: no free star", p.Name)
		}
		taken[home] = true
		sp := species.Fixed(p.Traits...)
		if p.Kind != "" || len(p.Mods) > 0 || len(p.Powers) > 0 {
			sub := p.Kind
			if sub == "" {
				sub = "biological"
			}
			sp = species.Rebuild(0, sub, p.Mods, species.Sound, p.Powers, "lush", p.Traits, species.Making{})
		}
		w.Bio[home] = BioComplex
		c := w.spawnCiv(home, sp, -1)
		r.ids[p.Name] = c.ID
		r.names[c.ID] = p.Name
	}
	// the colonies, once every home is placed: the free stars nearest home
	for _, p := range r.S.Peoples {
		c := r.Civ(p.Name)
		for _, s := range w.G.Near(c.Home, 3*fleetHop) {
			if len(c.Systems) >= max(1, p.Worlds) {
				break
			}
			if free(s) {
				taken[s] = true
				w.setOwner(s, c.ID)
				c.Systems = append(c.Systems, s)
			}
		}
		if len(c.Systems) < p.Worlds {
			return fmt.Errorf("%s: room for %d worlds of %d", p.Name, len(c.Systems), p.Worlds)
		}
		c.Peak = len(c.Systems)
	}
	return nil
}

// dress gives a placed people its nodes, its boon, its ships.
func (r *Run) dress(p *ScenarioPeople) error {
	w := r.W
	c := r.Civ(p.Name)
	knows := p.Knows
	if len(knows) == 0 {
		knows = []string{"relativistic"}
	}
	for _, k := range knows {
		if tech.Get(k) == nil {
			return fmt.Errorf("%s: no node %q", p.Name, k)
		}
		for _, n := range tech.Closure(k) {
			w.know(c, n)
		}
	}
	c.Starfaring = w.Now
	w.recompute(c)
	ships := 3
	if p.Ships != nil {
		ships = *p.Ships
	}
	if ships > 0 && c.Species.Profile().Can(species.Launches) {
		w.addGuard(c, c.Home, ships)
	}
	// the boon: fed beyond the nodes' and the ships' needs, and the levels asked
	var need flow.Income
	for _, u := range w.uses(c) {
		need.Add(u.Need)
	}
	need.Add(w.keepOf(c).Scale(float64(ships)))
	spare := 2.0
	if p.Spare != nil {
		spare = *p.Spare
	}
	for k := range need {
		need[k] += spare
	}
	boon := &Source{ID: len(w.Sources), Key: "scenario", Star: c.Home, Yield: need, Rarity: true, Holder: -1, Carried: -1, Legacy: -1, Maker: -1}
	if p.Mil != 0 {
		boon.Levels[0] = p.Mil - c.Mil
	}
	if p.Reach != 0 {
		boon.Reach = p.Reach - c.Reach
	}
	w.Sources = append(w.Sources, boon)
	w.sourcesAt[c.Home] = append(w.sourcesAt[c.Home], boon.ID)
	w.flows(c)
	w.recompute(c)
	if p.Leader != "" {
		l := w.raiseLeader(c, "scenario", w.leaderStream(c))
		l.Stance = p.Leader
		if !l.Deathless {
			l.Until = w.Now + Year(r.S.Ticks+1)*w.Cfg.Step // it reigns through the run
		}
		w.recompute(c)
	}
	return nil
}

// bond sets what two peoples are to each other.
func (r *Run) bond(b ScenarioBond) error {
	w := r.W
	c, e := r.Civ(b.A), r.Civ(b.B)
	if c == nil || e == nil {
		return fmt.Errorf("bond %s–%s: no such people", b.A, b.B)
	}
	if !b.Unmet {
		addMet(c, e.ID)
		addMet(e, c.ID)
		c.Reached[e.ID], e.Reached[c.ID] = true, true // met in the flesh: no meeting to come
		if !b.Unknown {
			c.Fathomed[e.ID], e.Fathomed[c.ID] = true, true
		}
		w.observe(c, e, e.Home, 0)
		w.observe(e, c, c.Home, 0)
	}
	if b.Grudge > 0 {
		c.Grudge[e.ID] = b.Grudge
	}
	if b.Claims {
		if c.Claim == nil {
			c.Claim = map[int]bool{}
		}
		for _, s := range e.Systems {
			c.Claim[s] = true
		}
	}
	if b.Wary > 0 {
		if c.Wary == nil {
			c.Wary = map[int]float64{}
		}
		c.Wary[e.ID] = b.Wary
	}
	if b.Fought > 0 {
		c.Fought[e.ID], e.Fought[c.ID] = b.Fought, b.Fought
	}
	if b.Truce > 0 {
		c.Truce[e.ID], e.Truce[c.ID] = w.Now+Year(b.Truce), w.Now+Year(b.Truce)
	}
	if b.Trade {
		startTrade(c, e)
	}
	switch b.Pact {
	case "":
	case "defence":
		w.formPact(c, e, Defensive, -1, -1)
	case "war":
		w.formPact(c, e, Aggressive, -1, -1)
	case "both":
		w.formPact(c, e, Both, -1, -1)
	default:
		return fmt.Errorf("bond %s–%s: no pact %q", b.A, b.B, b.Pact)
	}
	return nil
}

// Named gives a people made during the run — an heir, a rider woken —
// a name, and watches it.
func (r *Run) Named(name string, c *Civ) {
	r.ids[name], r.names[c.ID], r.watched[c.ID] = c.ID, name, true
}

// Civ is the named people, or nil.
func (r *Run) Civ(name string) *Civ {
	if id, ok := r.ids[name]; ok {
		return r.W.Civs[id]
	}
	return nil
}

// teller names peoples and tells events plainly, for a scenario's
// narrative and a watcher's dump.
type teller struct {
	names map[int]string // a people's name where it has one; civN otherwise
}

// Name is what a narrative calls a people: its name in the spec, else
// its token's number.
func (r *teller) Name(id int) string {
	if id < 0 {
		return "-"
	}
	if n, ok := r.names[id]; ok {
		return n
	}
	return "civ" + itoa(id)
}

// Tick runs one tick and tells it.
func (r *Run) Tick() {
	r.tick++
	r.W.runPhases()
	r.W.Now += r.W.Cfg.Step
	r.tell()
}

// Ticks runs n ticks.
func (r *Run) Ticks(n int) {
	for range n {
		r.Tick()
	}
}

// Saying is every reason a named people gave whose subject starts so.
func (r *Run) Saying(name, what string) []Said {
	id := r.ids[name]
	var out []Said
	for _, s := range r.Said {
		if s.Civ == id && strings.HasPrefix(s.What, what) {
			out = append(out, s)
		}
	}
	return out
}

// Wars is every war between two named peoples, in order.
func (r *Run) Wars(a, b string) []*War {
	ca, cb := r.Civ(a), r.Civ(b)
	var out []*War
	for _, wr := range r.W.Wars {
		if warKey(wr.Sides[0], wr.Sides[1]) == warKey(ca.ID, cb.ID) {
			out = append(out, wr)
		}
	}
	return out
}

// Narrative.

func (r *Run) printf(format string, a ...any) {
	if r.Out != nil {
		fmt.Fprintf(r.Out, format, a...)
	}
}

// header names the peoples and where they stand.
func (r *Run) header() {
	w := r.W
	r.printf("scenario %s: seed %d, %d stars\n", r.S.Name, r.S.Seed, len(w.G.Stars))
	for _, p := range r.S.Peoples {
		c := r.Civ(p.Name)
		var dists []string
		for _, q := range r.S.Peoples {
			if q.Name != p.Name {
				dists = append(dists, sprintf("%s %.1f ly", q.Name, w.G.Dist(c.Home, r.Civ(q.Name).Home)))
			}
		}
		r.printf("  %s = civ%d at s%d, %s (traits %s); %s\n", p.Name, c.ID, c.Home, c.posture(), strings.Join(c.Species.TraitKeys(), " "), strings.Join(dists, ", "))
	}
}

var tokenRe = regexp.MustCompile(`\{(civ|star):(\d+)[^}]*\}`)

// untoken puts names for the tokens in a line.
func (r *teller) untoken(s string) string {
	return tokenRe.ReplaceAllStringFunc(s, func(t string) string {
		m := tokenRe.FindStringSubmatch(t)
		id := atoiOr(m[2])
		if m[1] == "star" {
			return "s" + m[2]
		}
		return r.Name(id)
	})
}

func atoiOr(s string) int {
	n := 0
	for _, ch := range s {
		n = n*10 + int(ch-'0')
	}
	return n
}

// quietKinds are the events a narrative leaves out unless asked for all.
var quietKinds = map[Kind]bool{KDebug: true, KSystem: true, KPortrait: true, KNodeHeld: true}

// tell prints what happened since the last telling to the watched: the
// events and the reasons in the order they came, then what changed in
// the state of each watched people, each war and each fleet.
func (r *Run) tell() {
	w := r.W
	if r.Out == nil {
		r.cursor, r.reasons = len(w.Events), nil
		return
	}
	var lines []string
	ri := 0
	flush := func(upto int) {
		for ; ri < len(r.reasons) && r.reasons[ri].at <= upto; ri++ {
			rs := r.reasons[ri]
			if !r.watched[rs.civ] || (!r.All && !warReason(rs.what)) {
				continue
			}
			lines = append(lines, sprintf("  %s: %s — %s", r.Name(rs.civ), r.untoken(rs.what), r.untoken(rs.why)))
		}
	}
	for i := r.cursor; i < len(w.Events); i++ {
		flush(i)
		e := w.Events[i]
		if !r.watched[e.Subject] && !r.watched[e.Object] {
			continue
		}
		if !r.All && (quietKinds[e.Kind] || e.Kind == KReason) {
			continue
		}
		lines = append(lines, "  "+r.eventLine(e))
	}
	flush(len(w.Events))
	r.cursor, r.reasons = len(w.Events), nil
	for _, l := range r.state() {
		key := l[:strings.Index(l, ":")]
		if r.last[key] != l {
			r.last[key] = l
			lines = append(lines, "  = "+l)
		}
	}
	if len(lines) == 0 {
		return
	}
	if r.tick == 0 {
		r.printf("setting up\n")
	} else {
		r.printf("t%d (year %d)\n", r.tick, w.Now-w.Cfg.Step)
	}
	for _, l := range lines {
		r.printf("%s\n", l)
	}
}

// warReason says whether a reason is the war's or the council's: the
// ones a narrative of war tells by default.
func warReason(what string) bool {
	for _, p := range []string{"at war with", "offered", "council", "strike", "war on", "pact", "call", "yoke", "truce", "muster", "campaign", "intercept", "on the ", "sizing", "meeting of"} { // "on the" is the council's verdict on a people it might strike
		if strings.Contains(what, p) {
			return true
		}
	}
	return false
}

// eventLine is one event, told plainly.
func (r *teller) eventLine(e *Event) string {
	s := string(e.Kind) + " " + r.Name(e.Subject)
	if e.Object >= 0 {
		s += " → " + r.Name(e.Object)
	}
	if e.Star >= 0 {
		s += sprintf(" at s%d", e.Star)
	}
	if e.N != 0 {
		s += sprintf(" n=%d", e.N)
	}
	for _, k := range e.Keys() {
		v := e.P[k]
		if id, ok := v.(int); ok && civParam[k] {
			s += " " + k + "=" + r.Name(id)
			continue
		}
		s += sprintf(" %s=%v", k, r.untoken(fmt.Sprint(v)))
	}
	if e.Kind == KReason {
		return r.untoken(s)
	}
	return s
}

// civParam are the event parameters that name a people.
var civParam = map[string]bool{"by": true, "against": true, "via": true, "host": true, "master": true, "principal": true, "ally": true, "enemy": true}

// state is a line per watched people, per war among them and per fleet
// of theirs out, each keyed by what comes before its colon.
func (r *Run) state() []string {
	w := r.W
	var out []string
	ids := make([]int, 0, len(r.watched))
	for id := range r.watched {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		c := w.Civs[id]
		if !c.Living() {
			out = append(out, sprintf("%s: %s, %s", r.Name(id), stageWord(c), c.Cause))
			continue
		}
		s := sprintf("%s: %d worlds, %d ships (%d at home), mil %.1f reach %.0f, %s", r.Name(id), len(c.Systems), w.ships(c), shipsAt(w, c, c.Home), c.Mil, c.Reach, c.posture())
		if c.Leader != nil {
			s += sprintf(" (leader: %s %s)", c.Leader.Form, c.Leader.Stance)
		}
		if !c.Free() {
			k := "slave"
			if c.Vassal {
				k = "vassal"
			}
			s += " — " + k + " of " + r.Name(c.Master)
			if w.tributary(c) {
				how := "short when it must"
				if c.PaysFirst {
					how = "first"
				}
				s += sprintf(", tribute %.2f paid %s (short %d of %d ticks)", c.Rate, how, c.Tally.TributeShort, c.Tally.TributeTicks)
			}
		}
		if c.Rival >= 0 {
			s += ", rival " + r.Name(c.Rival)
		}
		out = append(out, s)
	}
	for _, wr := range w.Wars {
		if !r.watched[wr.Sides[0]] && !r.watched[wr.Sides[1]] {
			continue
		}
		a, b := r.Name(wr.Sides[0]), r.Name(wr.Sides[1])
		if wr.Over {
			out = append(out, sprintf("war %d %s→%s: over, %s after %d ticks, %d battles, taken %d/%d", wr.ID, a, b, wr.Result, (wr.Ended-wr.Began)/1000, wr.Battles, wr.Taken[0]+wr.Glassed[0], wr.Taken[1]+wr.Glassed[1]))
			continue
		}
		out = append(out, sprintf("war %d %s→%s: for %s, will %.2f/%.2f, %s/%s, %d battles, taken %d/%d", wr.ID, a, b, wr.Aim, wr.Will[0], wr.Will[1], wr.Verdict[0], wr.Verdict[1], wr.Battles, wr.Taken[0]+wr.Glassed[0], wr.Taken[1]+wr.Glassed[1]))
	}
	for _, id := range ids {
		c := w.Civs[id]
		for _, x := range w.fleetsOf(c) {
			switch x.Kind {
			case Guard, Scout, Survey, Roam:
				continue
			}
			s := sprintf("fleet %d of %s: %s of %d", x.ID, r.Name(id), x.Kind, x.Ships)
			if x.Target >= 0 {
				s += " against " + r.Name(x.Target)
			}
			switch {
			case x.LaidUp:
				s += sprintf(", laid up at s%d", x.Base)
			case x.Returning:
				s += sprintf(", going home to s%d", x.Star)
			case x.Base < 0:
				s += sprintf(", bound for s%d", x.Star)
			default:
				s += sprintf(", at s%d", x.Base)
			}
			out = append(out, s)
		}
		if c.Muster != nil {
			out = append(out, sprintf("muster of %s: against %s since year %d", r.Name(id), r.Name(c.Muster.Target), c.Muster.Since))
		}
	}
	return out
}

func shipsAt(w *World, c *Civ, star int) int {
	if g := w.guardAt(c, star); g != nil {
		return g.Ships
	}
	return 0
}
