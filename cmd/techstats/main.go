// Command techstats runs a batch of worlds and reports how the tech tree is
// used: every node with what it does, how often it is reached, what the
// common builds are, and what the long-lived peoples knew.
//
//	go run ./cmd/techstats -seeds 10 -out reports/tech
//
// It writes civs.jsonl (one record per civilisation, for later questions),
// stats.txt (the one-line tuning stats per seed) and report.md.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"worldgen/internal/flow"
	"worldgen/internal/galaxy"
	"worldgen/internal/history"
	"worldgen/internal/legends"
	"worldgen/internal/mind"
	"worldgen/internal/tech"
)

// Rec is one civilisation, flattened.
type Rec struct {
	Seed      uint64             `json:"seed"`
	ID        int                `json:"id"`
	Name      string             `json:"name"`
	Species   string             `json:"species"` // the blood's name; the people's own is Name
	Sub       string             `json:"sub"`
	Mods      []string           `json:"mods"`
	Traits    []string           `json:"traits"`
	Posture   string             `json:"posture"`
	Honour    string             `json:"honour"`
	Tally     history.Tally      `json:"tally"`
	Met       int                `json:"met"`
	Nomad     bool               `json:"nomad"`
	Aloft     bool               `json:"aloft"`  // took to the sky at some point
	Stars     bool               `json:"stars"`  // reached the stars at some point
	Rested    bool               `json:"rested"` // came to rest after
	World     string             `json:"world"`
	Made      string             `json:"made,omitempty"`
	Born      float64            `json:"born_myr"` // Myr after the dawn of the age
	Lived     float64            `json:"lived_myr"`
	Fate      string             `json:"fate"`
	Cause     string             `json:"cause"`
	Into      string             `json:"into,omitempty"`
	Standing  bool               `json:"standing"` // still active at the present
	Era       int                `json:"era"`      // deepest era ever reached
	Peak      int                `json:"peak"`
	Ruled     int                `json:"ruled"`
	Uplifts   int                `json:"uplifts"`
	Known     []string           `json:"known"`   // at the end
	Ever      []string           `json:"ever"`    // ever known: learned or inherited
	Learned   map[string]float64 `json:"learned"` // node -> Myr after birth
	Miracles  map[string]string  `json:"miracles"`
	Word      string             `json:"word,omitempty"`
	Record    []string           `json:"record"`
	Scars     []string           `json:"scars"`
	Boons     []string           `json:"boons"`
	Mil       float64            `json:"mil"`
	Sur       float64            `json:"sur"`
	Soc       float64            `json:"soc"`
	Wis       float64            `json:"wis"`      // at the end
	PeakWis   float64            `json:"peak_wis"` // the most ever
	WisFrom   [5]float64         `json:"wis_from"` // species, tech, experience, boons, scars, as last derived
	Cycle     bool               `json:"knows_cycle"`
	DarkAges  int                `json:"dark_ages"`
	Renaiss   int                `json:"renaissances"`
	Held      int                `json:"held"`             // tales held at the end
	Myth      int                `json:"myth"`             // of them myth
	Monsters  int                `json:"monsters"`         // peoples remembered as monsters at the end
	Morality  string             `json:"morality"`         // what the people counts as wrong: amoral, individual, herd, fixed on X
	Excused   int                `json:"excused"`          // tales held whose fact is a crime and the people's judgment is not
	Condemned int                `json:"condemned"`        // tales held whose fact is no crime and the people's judgment is one
	Split     int                `json:"split,omitempty"`  // on the first people of a world: facts a crime to one people that knows them and a deed to another
	Shared    int                `json:"shared,omitempty"` // on the first people of a world: crimes known to two peoples or more
	LoreDials history.Dials      `json:"lore_dials"`
	Income    flow.Income        `json:"income"` // at the people's height of means
	Upkeep    flow.Income        `json:"upkeep"`
	Want      flow.Income        `json:"want"`
	Shed      map[string]int     `json:"shed"`              // ticks each node spent dark
	Had       []string           `json:"had"`               // rarities ever had, by key
	Harnessed []string           `json:"harnessed"`         // source kinds ever harnessed, by key
	Built     map[string]int     `json:"built"`             // structures raised, by key
	Granted   []string           `json:"granted"`           // nodes learned with their grant had
	FellDep   bool               `json:"fell_dependent"`    // depended on a partner at the moment of its fall
	Objects   []string           `json:"objects,omitempty"` // objects made: kind:form:fate, with "thinks" and cuttings given
	Ships     int                `json:"ships"`             // the most ships ever in being
	ShipsEnd  int                `json:"ships_end"`         // ships in being at the end, and of them laid up
	LaidUp    int                `json:"laid_up"`
	Origin    string             `json:"origin,omitempty"` // a branch, a cult, an uplift; "" for a cradle
	Sick      float64            `json:"sick_myr"`         // Myr after birth the first plague came; -1 for never
	Master    bool               `json:"master,omitempty"` // born under a master: made, or held
	ever      map[string]bool
	frontier  string
	signature string
}

func main() {
	seeds := flag.Int("seeds", 10, "how many worlds to run")
	from := flag.Uint64("from", 1, "first seed")
	at := flag.String("at", "sol", "where in the galaxy")
	out := flag.String("out", "reports/tech", "output directory")
	tuning := flag.String("tuning", "", "a JSON file of mind.Tuning; fields left out keep their defaults")
	var tunes []string
	flag.Func("tune", "one override of the mind's tuning, Group.Field=value; may repeat", func(s string) error { tunes = append(tunes, s); return nil })
	flag.Parse()

	tune, err := mind.Configure(*tuning, tunes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if _, err := galaxy.RegionByName(*at); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	type run struct {
		seed    uint64
		recs    []Rec
		wars    []WarRec
		battles []BattleRec
		sights  []SightRec
		meets   []MeetRec
		fleets  []FleetRec
		fields  FieldRec
		pairs   []PairRec
		ks      []ContractRec
		sells   []SellRec
		bloc    BlocRec
		plagues []PlagueRec
		stats   string
		ages    float64
	}
	runs := make([]run, *seeds)
	var wg sync.WaitGroup
	for i := range runs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			seed := *from + uint64(i)
			cfg := history.DefaultConfig()
			cfg.Region = *at
			cfg.Tuning = tune
			w := history.Generate(seed, cfg)
			var sb strings.Builder
			legends.Stats(&sb, w)
			sights, meets, fleets, fields := flattenSightings(w)
			ks, sells := flattenContracts(w)
			runs[i] = run{ks: ks, sells: sells, bloc: flattenBlocs(w), plagues: flattenPlagues(w), seed: seed, recs: flatten(w), wars: flattenWars(w), battles: flattenBattles(w), sights: sights, meets: meets, fleets: fleets, fields: fields, pairs: flattenPairs(w), stats: sb.String(), ages: float64(w.Present-w.Cfg.Dawn) / 1e6}
		}(i)
	}
	wg.Wait()

	var recs []Rec
	var wars []WarRec
	var battles []BattleRec
	var sights []SightRec
	var meets []MeetRec
	var fleets []FleetRec
	var fields []FieldRec
	var pairs []PairRec
	var ks []ContractRec
	var sells []SellRec
	var blocs []BlocRec
	var plagues []PlagueRec
	var stats []string
	ageSum := 0.0
	for _, r := range runs {
		recs = append(recs, r.recs...)
		wars = append(wars, r.wars...)
		battles = append(battles, r.battles...)
		sights = append(sights, r.sights...)
		meets = append(meets, r.meets...)
		fleets = append(fleets, r.fleets...)
		fields = append(fields, r.fields)
		pairs = append(pairs, r.pairs...)
		ks = append(ks, r.ks...)
		sells = append(sells, r.sells...)
		blocs = append(blocs, r.bloc)
		plagues = append(plagues, r.plagues...)
		stats = append(stats, r.stats)
		ageSum += r.ages
	}
	must(writeJSONL(filepath.Join(*out, "civs.jsonl"), recs))
	must(writeWars(filepath.Join(*out, "wars.jsonl"), wars))
	must(writeJSONL(filepath.Join(*out, "plagues.jsonl"), plagues))
	must(os.WriteFile(filepath.Join(*out, "stats.txt"), []byte(strings.Join(stats, "")), 0o644))
	f, err := os.Create(filepath.Join(*out, "report.md"))
	must(err)
	defer f.Close()
	report(f, recs, *seeds, *from, *at, ageSum/float64(*seeds))
	warReport(f, recs, wars, *seeds)
	exploreReport(f, recs)
	loreReport(f, recs)
	meansReport(f, recs)
	shipsReport(f, recs)
	battlesReport(f, recs, battles)
	sightingsReport(f, recs, sights, meets, fleets, fields)
	wisdomReport(f, recs, pairs)
	contractReport(f, sells, ks, wars)
	blocReport(f, blocs, recs)
	sickReport(f, plagues, recs, *seeds)
	fmt.Printf("%d civilisations over %d worlds; wrote %s\n", len(recs), *seeds, *out)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func flatten(w *history.World) []Rec {
	var out []Rec
	for _, c := range w.Civs {
		end := c.Fell
		if c.Active() {
			end = w.Present
		}
		r := Rec{
			Seed: w.Seed, ID: c.ID, Name: c.Name, Species: c.Species.Name, Sub: c.Species.Sub.String(), World: c.Species.World.Key, Made: c.Species.Made,
			Born: float64(c.Born-w.Cfg.Dawn) / 1e6, Lived: float64(end-c.Born) / 1e6,
			Fate: c.Fate.String(), Cause: c.Cause, Into: c.Into, Standing: c.Active(), Origin: c.Origin, Sick: -1, Master: c.Sire >= 0,
			Peak: c.Peak, Ruled: c.Ruled, Uplifts: c.Uplifts, Word: c.Word,
			Miracles: map[string]string{}, Learned: map[string]float64{},
			Record: append([]string(nil), c.Record...), Mil: c.Mil, Sur: c.Sur, Soc: c.Soc,
			Wis: c.Wis, PeakWis: c.PeakWis, WisFrom: history.WisdomParts(c),
			Cycle: c.KnowsCycle, DarkAges: c.DarkAges, Renaiss: c.Renaissances,
			ever: map[string]bool{},
		}
		if c.FirstPlague > 0 {
			r.Sick = float64(c.FirstPlague-c.Born) / 1e6
		}
		for _, d := range c.Species.Mods.Defs() {
			r.Mods = append(r.Mods, d.Key)
		}
		for _, t := range c.Species.Traits {
			r.Traits = append(r.Traits, t.Key)
			switch t.Group {
			case "stance":
				r.Posture = t.Key
			case "honour":
				r.Honour = t.Key
			}
		}
		r.Tally = c.Tally
		r.Met = len(c.Met)
		r.LoreDials = c.LoreDials
		r.Income, r.Upkeep, r.Want = c.HighIncome, c.HighUpkeep, c.HighWant
		r.Had, r.Harnessed, r.Granted = keys(c.Had), keys(c.Harnessed), keys(c.Granted)
		r.Built = map[string]int{}
		for k, n := range c.Built {
			r.Built[k] = n
		}
		r.Shed = map[string]int{}
		for k, n := range c.ShedTicks {
			r.Shed[k] = n
		}
		r.Held, r.Myth, r.Monsters = history.LoreCounts(w, c)
		r.Morality = c.Morality.Word()
		r.Excused, r.Condemned = history.Judged(w, c)
		if len(out) == 0 {
			r.Split, r.Shared = history.SplitFacts(w)
		}
		r.FellDep = c.FellDependent
		r.Ships = c.PeakShips
		r.ShipsEnd, _, r.LaidUp = history.ShipsOf(w, c)
		for _, src := range w.Sources {
			if src.Form == "" || src.Maker != c.ID {
				continue
			}
			o := src.Key + ":" + src.Form
			if src.Sentient {
				o += ":thinks"
			}
			if src.Given > 0 {
				o += fmt.Sprintf(":given %d", src.Given)
			}
			if src.Fate != "" {
				o += ":" + src.Fate
			}
			r.Objects = append(r.Objects, o)
		}
		r.Nomad = c.Has("nomadic")
		r.Stars = c.Starfaring > 0
		r.Rested = c.Rested
		for _, rec := range c.Record {
			if rec == "took to the sky" {
				r.Aloft = true
			}
		}
		for k := range c.Known {
			r.Known = append(r.Known, k)
			r.ever[k] = true
		}
		for k, y := range c.Learned {
			r.Learned[k] = float64(y-c.Born) / 1e6
			r.ever[k] = true
		}
		for k := range r.ever {
			r.Ever = append(r.Ever, k)
			if e := tech.Get(k).Era; e > r.Era {
				r.Era = e
			}
		}
		for k, v := range c.Miracles {
			r.Miracles[k] = v
		}
		r.Scars = keys(c.Scars)
		r.Boons = keys(c.Boons)
		sort.Strings(r.Known)
		sort.Strings(r.Ever)
		for _, sl := range []*[]string{&r.Record, &r.Known, &r.Ever, &r.Traits} {
			if *sl == nil {
				*sl = []string{}
			}
		}
		r.frontier = frontier(r.ever)
		r.signature = signature(r.ever)
		out = append(out, r)
	}
	return out
}

func keys(m map[string]bool) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func writeJSONL[T any](path string, recs []T) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	bw := bufio.NewWriter(f)
	enc := json.NewEncoder(bw)
	for _, r := range recs {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return bw.Flush()
}

// frontier is the shape of a build: the nodes at a people's deepest era
// that nothing else they knew depends on. Miracles count as era 4.
func frontier(ever map[string]bool) string {
	deepest := 0
	for k := range ever {
		deepest = max(deepest, tech.Get(k).Era)
	}
	if deepest < 3 {
		return ""
	}
	var out []string
	for k := range ever {
		n := tech.Get(k)
		if n.Era < deepest {
			continue
		}
		leaf := true
		for o := range ever {
			for _, p := range tech.Get(o).Prereqs {
				if p == k {
					leaf = false
				}
			}
		}
		if leaf {
			out = append(out, n.Name)
		}
	}
	sort.Strings(out)
	return strings.Join(out, " + ")
}

// signature is the two domains a people went deepest in, by count of era 2+ nodes.
func signature(ever map[string]bool) string {
	depth := map[string]int{}
	for k := range ever {
		if n := tech.Get(k); n.Era >= 2 {
			depth[n.Domain]++
		}
	}
	ds := append([]string(nil), tech.Domains...)
	sort.SliceStable(ds, func(i, j int) bool { return depth[ds[i]] > depth[ds[j]] })
	if depth[ds[0]] == 0 {
		return "(none)"
	}
	if depth[ds[1]] == 0 {
		return ds[0]
	}
	return ds[0] + " + " + ds[1]
}

// ---- the report ----

type counter struct {
	keys []string
	n    map[string]int
	sub  map[string][]float64 // lifetimes
}

func newCounter() *counter { return &counter{n: map[string]int{}, sub: map[string][]float64{}} }
func (c *counter) add(k string, lived float64) {
	if _, ok := c.n[k]; !ok {
		c.keys = append(c.keys, k)
	}
	c.n[k]++
	c.sub[k] = append(c.sub[k], lived)
}
func (c *counter) sorted() []string {
	ks := append([]string(nil), c.keys...)
	sort.SliceStable(ks, func(i, j int) bool {
		if c.n[ks[i]] != c.n[ks[j]] {
			return c.n[ks[i]] > c.n[ks[j]]
		}
		return ks[i] < ks[j]
	})
	return ks
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	if len(s)%2 == 1 {
		return s[len(s)/2]
	}
	return (s[len(s)/2-1] + s[len(s)/2]) / 2
}

func quantile(xs []float64, q float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	i := int(q * float64(len(s)-1))
	return s[i]
}

func pct(a, b int) string {
	if b == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", 100*float64(a)/float64(b))
}

func effects(n *tech.Node) string {
	var e []string
	add := func(name string, v float64) {
		if v != 0 {
			e = append(e, fmt.Sprintf("%s %+.1f", name, v))
		}
	}
	add("Mil", n.Mil)
	add("Sur", n.Sur)
	add("Soc", n.Soc)
	if n.Reach > 0 {
		e = append(e, fmt.Sprintf("reach %g ly", n.Reach))
	}
	if n.Speed > 0 {
		e = append(e, fmt.Sprintf("speed %g y/ly", n.Speed))
	}
	if n.Env > 0 {
		e = append(e, fmt.Sprintf("envelope +%d", n.Env))
	}
	if len(n.Focus) > 0 {
		var fs []string
		for _, d := range tech.Domains {
			if v, ok := n.Focus[d]; ok {
				fs = append(fs, fmt.Sprintf("%s ×%g", d, v))
			}
		}
		e = append(e, "tilt "+strings.Join(fs, ", "))
	}
	for _, k := range n.Structures {
		e = append(e, "unlocks "+tech.Structures[k].Name)
	}
	if n.Filter != "" {
		d, lv := history.FilterDiff(n.Filter)
		e = append(e, fmt.Sprintf("filter: %s (diff %g on %s)", history.FilterName(n.Filter), d, strings.Join(lv, "/")))
	}
	if n.Patience > 0 {
		e = append(e, fmt.Sprintf("after %g Myr of life", n.Patience/1000))
	}
	if n.Chance > 0 {
		e = append(e, fmt.Sprintf("%.0f%% chance per attempt", 100*n.Chance))
	}
	if n.Weight != 1 {
		e = append(e, fmt.Sprintf("weight %g", n.Weight))
	}
	return strings.Join(e, "; ")
}

func names(keys []string) string {
	var out []string
	for _, k := range keys {
		out = append(out, tech.Get(k).Name)
	}
	return strings.Join(out, ", ")
}

func report(out io.Writer, recs []Rec, seeds int, from uint64, at string, ageMean float64) {
	p := func(f string, a ...any) { fmt.Fprintf(out, f+"\n", a...) }
	N := len(recs)
	all := func(r Rec) bool { return true }
	count := func(pred func(Rec) bool) int {
		n := 0
		for _, r := range recs {
			if pred(r) {
				n++
			}
		}
		return n
	}
	lives := func(pred func(Rec) bool) []float64 {
		var xs []float64
		for _, r := range recs {
			if pred(r) {
				xs = append(xs, r.Lived)
			}
		}
		return xs
	}

	p("# The tech tree in use")
	p("")
	p("%d worlds (seeds %d to %d) at %s, %d civilisations, mean age length %.0f Myr. A node counts as reached if a people ever held it, by pursuit, by a find, or by inheritance from a parent people. Lifetimes are from birth to the end of activity, or to the present for those still standing.", seeds, from, from+uint64(seeds)-1, at, N, ageMean)
	p("")

	// ---- depth ----
	p("## How deep peoples get")
	p("")
	p("| Deepest era reached | Peoples | Share | Median life (Myr) |")
	p("|---|---|---|---|")
	for e := 0; e <= 4; e++ {
		pred := func(r Rec) bool { return r.Era == e }
		p("| %d %s | %d | %s | %.2f |", e, tech.EraNames[e], count(pred), pct(count(pred), N), median(lives(pred)))
	}
	p("")
	ns := []float64{}
	for _, r := range recs {
		ns = append(ns, float64(len(r.Ever)))
	}
	p("Nodes ever held per people: median %.0f, lower quartile %.0f, upper quartile %.0f, most %.0f of %d.", median(ns), quantile(ns, 0.25), quantile(ns, 0.75), quantile(ns, 1), len(tech.Nodes))
	stars := func(r Rec) bool {
		for k := range r.ever {
			if tech.Get(k).Reach >= 10 {
				return true
			}
		}
		return r.Miracles["directed_evolution"] != "" || r.Miracles["ftl"] != ""
	}
	p("Reached the stars (any node with reach 10 ly or more, or the Flesh or the Door): %d (%s), median life %.2f Myr against %.2f for those who did not.", count(stars), pct(count(stars), N), median(lives(stars)), median(lives(func(r Rec) bool { return !stars(r) })))
	p("")

	// ---- the tree ----
	p("## The tree")
	p("")
	p("Reached is the share of all peoples that ever held the node. Of eligible is the share among peoples that held every prerequisite, which is the real conversion rate. Learned at is the median time from a people's birth to the node, for those who learned it themselves. Life is the median lifetime of peoples that held it.")
	for e := 0; e <= 4; e++ {
		p("")
		title := tech.EraNames[e]
		if e == 4 {
			title = "exotic: the deep tree"
		}
		p("### Era %d, %s", e, title)
		p("")
		p("| Node | Domain | Needs | Price | Reached | Of eligible | Learned at | Life | Effects |")
		p("|---|---|---|---|---|---|---|---|---|")
		for _, n := range tech.Nodes {
			if n.Era != e || n.Miracle {
				continue
			}
			p("%s", row(recs, n))
		}
		p("")
		for _, n := range tech.Nodes {
			if n.Era != e || n.Miracle {
				continue
			}
			p("- **%s**: %s", n.Name, n.Desc)
		}
	}
	p("")
	p("### Miracles, the powers apart from the tree")
	p("")
	p("| Node | Domain | Needs | Price | Reached | Of eligible | Learned at | Life | Effects |")
	p("|---|---|---|---|---|---|---|---|---|")
	for _, n := range tech.Miracles {
		p("%s", row(recs, n))
	}
	p("")
	for _, n := range tech.Miracles {
		p("- **%s**: %s", n.Name, n.Desc)
	}
	p("")
	p("How miracles were held, by route (a people may hold one by more than one route):")
	p("")
	p("| Miracle | Born | Leap | Found | Wielded | Holders | Median life (Myr) |")
	p("|---|---|---|---|---|---|---|")
	for _, n := range tech.Miracles {
		by := map[string]int{}
		var xs []float64
		for _, r := range recs {
			if how, ok := r.Miracles[n.Key]; ok {
				by[how]++
				xs = append(xs, r.Lived)
			}
		}
		p("| %s | %d | %d | %d | %d | %d | %.2f |", n.Name, by["born"], by["leap"], by["found"], by["wielded"], len(xs), median(xs))
	}
	p("")

	// ---- filters at nodes ----
	p("## What happens at the nodes that test a people")
	p("")
	p("Outcomes recorded when a node's filter was faced. Foresaw means the Sight stepped around it.")
	p("")
	p("| Node | Filter | Faced | Overcame | Scarred | Fell | Foresaw |")
	p("|---|---|---|---|---|---|---|")
	for _, n := range tech.Nodes {
		if n.Filter == "" {
			continue
		}
		fn := history.FilterName(n.Filter)
		o := map[string]int{}
		for _, r := range recs {
			for _, rec := range r.Record {
				for _, verb := range []string{"overcame", "scarred by", "fell to", "foresaw"} {
					if strings.HasPrefix(rec, verb+" "+fn) {
						o[verb]++
					}
				}
			}
		}
		faced := o["overcame"] + o["scarred by"] + o["fell to"] + o["foresaw"]
		p("| %s | %s | %d | %s | %s | %s | %s |", n.Name, fn, faced, pct(o["overcame"], faced), pct(o["scarred by"], faced), pct(o["fell to"], faced), pct(o["foresaw"], faced))
	}
	p("")

	// ---- builds ----
	deep := func(r Rec) bool { return r.Era >= 3 }
	nDeep := count(deep)
	p("## Common builds")
	p("")
	p("Among the %d peoples (%s) that reached era 3, the interstellar tree. A build is read two ways: the two domains a people went deepest in (by count of era 2+ nodes, deepest first), and the frontier, the nodes at their deepest era that nothing else they knew depends on.", nDeep, pct(nDeep, N))
	p("")
	p("### By domain pair")
	p("")
	sig := newCounter()
	for _, r := range recs {
		if deep(r) {
			sig.add(r.signature, r.Lived)
		}
	}
	p("| Domains | Peoples | Share | Median life (Myr) | Hold a miracle |")
	p("|---|---|---|---|---|")
	for _, k := range sig.sorted() {
		m := 0
		for _, r := range recs {
			if deep(r) && r.signature == k && len(r.Miracles) > 0 {
				m++
			}
		}
		p("| %s | %d | %s | %.2f | %s |", k, sig.n[k], pct(sig.n[k], nDeep), median(sig.sub[k]), pct(m, sig.n[k]))
	}
	p("")
	p("### By frontier")
	p("")
	fr := newCounter()
	for _, r := range recs {
		if deep(r) {
			fr.add(r.frontier, r.Lived)
		}
	}
	p("The %d most common frontiers of %d distinct:", 20, len(fr.keys))
	p("")
	p("| Frontier | Peoples | Median life (Myr) |")
	p("|---|---|---|")
	for i, k := range fr.sorted() {
		if i >= 20 {
			break
		}
		p("| %s | %d | %.2f |", k, fr.n[k], median(fr.sub[k]))
	}
	p("")
	p("### Era 4 pairs")
	p("")
	p("Which deep nodes are held together, among peoples holding at least two era 4 nodes:")
	p("")
	pairs := newCounter()
	for _, r := range recs {
		var e4 []string
		for _, k := range r.Ever {
			if n := tech.Get(k); n.Era == 4 && !n.Miracle {
				e4 = append(e4, n.Name)
			}
		}
		for i := range e4 {
			for j := i + 1; j < len(e4); j++ {
				pairs.add(e4[i]+" + "+e4[j], r.Lived)
			}
		}
	}
	p("| Pair | Peoples |")
	p("|---|---|")
	for i, k := range pairs.sorted() {
		if i >= 15 {
			break
		}
		p("| %s | %d |", k, pairs.n[k])
	}
	p("")
	p("### The way to the stars")
	p("")
	p("The first node that gave a people reach of 10 ly or more:")
	p("")
	first := newCounter()
	for _, r := range recs {
		// in tree order, so ties fall the same way every run
		best, at := "", 1e9
		for _, n := range tech.Nodes {
			if y, ok := r.Learned[n.Key]; ok && n.Reach >= 10 && y < at {
				best, at = n.Name, y
			}
		}
		if best == "" {
			for _, n := range tech.Nodes {
				if r.ever[n.Key] && n.Reach >= 10 {
					best = n.Name + " (inherited)"
					break
				}
			}
		}
		if best == "" && (r.Miracles["directed_evolution"] != "" || r.Miracles["ftl"] != "") {
			best = "a miracle held from birth or by a find"
		}
		if best != "" {
			first.add(best, r.Lived)
		}
	}
	p("| Node | Peoples | Median life (Myr) |")
	p("|---|---|---|")
	for _, k := range first.sorted() {
		p("| %s | %d | %.2f |", k, first.n[k], median(first.sub[k]))
	}
	p("")

	// ---- success ----
	allLives := lives(all)
	top := quantile(allLives, 0.9)
	long := func(r Rec) bool { return r.Lived >= 10 }
	nLong := count(long)
	p("## Successful builds")
	p("")
	p("Success here is a long life, not survival to the present. The tenth-longest-lived of every hundred peoples lasted %.1f Myr or more. The cut used below is 10 Myr: %d peoples (%s). Their median birth was %.0f Myr into the age against %.0f for everyone, so part of a long life is simply being born early.", top, nLong, pct(nLong, N), median(func() []float64 {
		var xs []float64
		for _, r := range recs {
			if long(r) {
				xs = append(xs, r.Born)
			}
		}
		return xs
	}()), median(func() []float64 {
		var xs []float64
		for _, r := range recs {
			xs = append(xs, r.Born)
		}
		return xs
	}()))
	p("")
	p("### Life by what was held")
	p("")
	p("| Held | Peoples | Median life (Myr) | Lived 10+ Myr |")
	p("|---|---|---|---|")
	rowH := func(label string, pred func(Rec) bool) {
		n := count(pred)
		l := count(func(r Rec) bool { return pred(r) && long(r) })
		p("| %s | %d | %.2f | %s |", label, n, median(lives(pred)), pct(l, n))
	}
	rowH("everyone", all)
	rowH("no miracle", func(r Rec) bool { return len(r.Miracles) == 0 })
	rowH("any miracle", func(r Rec) bool { return len(r.Miracles) > 0 })
	for _, n := range tech.Miracles {
		rowH(n.Name, func(r Rec) bool { return r.Miracles[n.Key] != "" })
	}
	for _, how := range []string{"born", "leap", "found", "wielded"} {
		rowH("a miracle by "+how, func(r Rec) bool {
			for _, h := range r.Miracles {
				if h == how {
					return true
				}
			}
			return false
		})
	}
	rowH("knows the cycle", func(r Rec) bool { return r.Cycle })
	p("")
	p("### Nodes that mark the long-lived")
	p("")
	p("For every era 2+ node, its share among the long-lived against its share among all peoples that reached era 2. Lift above 1 means the survivors held it more often than the crowd.")
	p("")
	e2 := func(r Rec) bool { return r.Era >= 2 }
	nE2 := count(e2)
	type lift struct {
		name       string
		long, base float64
		lift       float64
	}
	var lifts []lift
	for _, n := range tech.Nodes {
		if n.Era < 2 || n.Miracle {
			continue
		}
		a := count(func(r Rec) bool { return long(r) && r.ever[n.Key] })
		b := count(func(r Rec) bool { return e2(r) && r.ever[n.Key] })
		if b < 5 {
			continue
		}
		la, lb := float64(a)/float64(max(nLong, 1)), float64(b)/float64(nE2)
		lifts = append(lifts, lift{n.Name, la, lb, la / lb})
	}
	sort.Slice(lifts, func(i, j int) bool { return lifts[i].lift > lifts[j].lift })
	p("| Node | Among long-lived | Among era 2+ | Lift |")
	p("|---|---|---|---|")
	for i, l := range lifts {
		if i < 15 || i >= len(lifts)-8 {
			p("| %s | %.0f%% | %.0f%% | %.2f |", l.name, 100*l.long, 100*l.base, l.lift)
		} else if i == 15 {
			p("| … | | | |")
		}
	}
	p("")
	p("### Domain pairs of the long-lived")
	p("")
	sigL := newCounter()
	for _, r := range recs {
		if long(r) {
			sigL.add(r.signature, r.Lived)
		}
	}
	p("| Domains | Peoples | Share of long-lived | Share of era 3+ |")
	p("|---|---|---|---|")
	for i, k := range sigL.sorted() {
		if i >= 12 {
			break
		}
		p("| %s | %d | %s | %s |", k, sigL.n[k], pct(sigL.n[k], nLong), pct(sig.n[k], nDeep))
	}
	p("")
	p("### Frontiers of the long-lived")
	p("")
	frL := newCounter()
	for _, r := range recs {
		if long(r) {
			frL.add(r.frontier, r.Lived)
		}
	}
	p("| Frontier | Peoples | Median life (Myr) |")
	p("|---|---|---|")
	for i, k := range frL.sorted() {
		if i >= 15 {
			break
		}
		name := k
		if name == "" {
			name = "(never past era 2)"
		}
		p("| %s | %d | %.2f |", name, frL.n[k], median(frL.sub[k]))
	}
	p("")
	p("### How the long-lived ended")
	p("")
	ends := newCounter()
	for _, r := range recs {
		if long(r) {
			f := r.Fate
			if r.Standing {
				f = "still standing"
			}
			ends.add(f, r.Lived)
		}
	}
	p("| End | Peoples |")
	p("|---|---|")
	for _, k := range ends.sorted() {
		p("| %s | %d |", k, ends.n[k])
	}
	p("")
	p("### The thirty longest lives")
	p("")
	sorted := append([]Rec(nil), recs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Lived > sorted[j].Lived })
	p("| Seed | People | Nature | Born (Myr) | Lived (Myr) | Era | Nodes | Peak | Miracles | Frontier | End |")
	p("|---|---|---|---|---|---|---|---|---|---|---|")
	for i, r := range sorted {
		if i >= 30 {
			break
		}
		var ms []string
		for k, how := range r.Miracles {
			ms = append(ms, tech.Get(k).Name+" ("+how+")")
		}
		sort.Strings(ms)
		end := r.Fate
		if r.Standing {
			end = "standing"
		}
		if r.Cause != "" {
			end += ": " + r.Cause
		}
		p("| %d | %s | %s | %.0f | %.1f | %d | %d | %d | %s | %s | %s |", r.Seed, r.Name, r.nature(), r.Born, r.Lived, r.Era, len(r.Ever), r.Peak, strings.Join(ms, ", "), r.frontier, end)
	}
	p("")
	p("## Natures")
	p("")
	p("Substrate and modifiers; a swarm is a biological people with the swarming trait, listed apart.")
	p("")
	kinds := newCounter()
	for _, r := range recs {
		kinds.add(r.kind(), r.Lived)
	}
	p("| Nature | Peoples | Median life (Myr) | Reached era 3 | Lived 10+ Myr |")
	p("|---|---|---|---|---|")
	for _, k := range kinds.sorted() {
		d := count(func(r Rec) bool { return r.kind() == k && deep(r) })
		l := count(func(r Rec) bool { return r.kind() == k && long(r) })
		p("| %s | %d | %.2f | %s | %s |", k, kinds.n[k], median(kinds.sub[k]), pct(d, kinds.n[k]), pct(l, kinds.n[k]))
	}
}

// nature is the substrate and the modifiers: "biological", "machine, hive".
func (r Rec) nature() string {
	if len(r.Mods) == 0 {
		return r.Sub
	}
	return r.Sub + ", " + strings.Join(r.Mods, ", ")
}

// kind is the nature with the swarm told apart, for the tables.
func (r Rec) kind() string {
	for _, t := range r.Traits {
		if t == "swarming" {
			return r.nature() + " (swarm)"
		}
	}
	return r.nature()
}

func row(recs []Rec, n *tech.Node) string {
	N := len(recs)
	reached, eligible := 0, 0
	var at, life []float64
	for _, r := range recs {
		ok := forMatches(n.For, r)
		for _, p := range n.Prereqs {
			if !r.ever[p] && !anySub(r, p) {
				ok = false
			}
		}
		if ok {
			eligible++
		}
		if r.ever[n.Key] {
			reached++
			life = append(life, r.Lived)
			if y, ok := r.Learned[n.Key]; ok {
				at = append(at, y)
			}
		}
	}
	price := fmt.Sprintf("%g", n.Price())
	learned := "-"
	if len(at) > 0 {
		learned = fmt.Sprintf("%.2f Myr", median(at))
	}
	return fmt.Sprintf("| %s | %s | %s | %s | %d (%s) | %s | %s | %.2f | %s |", n.Name, n.Domain, names(n.Prereqs), price, reached, pct(reached, N), pct(reached, eligible), learned, median(life), effects(n))
}

// forMatches says whether a node reserved for some peoples is on this one's tree.
func forMatches(f string, r Rec) bool {
	switch {
	case f == "":
		return true
	case strings.HasPrefix(f, "sub:"):
		return r.Sub == f[4:]
	case strings.HasPrefix(f, "mod:"):
		for _, m := range r.Mods {
			if m == f[4:] {
				return true
			}
		}
	case strings.HasPrefix(f, "world:"):
		return r.World == f[6:]
	case strings.HasPrefix(f, "trait:"):
		for _, t := range r.Traits {
			if t == f[6:] {
				return true
			}
		}
	}
	return false
}

// anySub says whether a stand-in for a prerequisite was ever held.
func anySub(r Rec, p string) bool {
	for _, s := range tech.Subs[p] {
		if r.ever[s] {
			return true
		}
	}
	return false
}
