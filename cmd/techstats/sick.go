package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"strconv"
	"worldgen/internal/plague"
)

// Sickness: plagues born by kind and cause, how far they went, what they
// took, and what the peoples did about them.

// PlagueRec is one plague.
type PlagueRec struct {
	Seed      uint64
	Name      string
	Kind      string
	Contagion float64
	Lethality float64
	Cause     string
	Born      float64 // Myr into the age
	Peak      int     // hosts at once
	Caught    int     // peoples that had it, in all
	Worlds    int
	Peoples   int // ended or brought low
	Cults     int
	Cures     int
	Refusals  int
	Woken     int
	Wildfire  bool
	Extinct   bool
	Made      bool // shaped as a weapon
	Maker     int  // the people that made it, or -1
	Rider     int  // the parasite people it became, or -1
	Poisoned  int  // peoples it was put in by stealth
}

func flattenPlagues(w *world) []PlagueRec {
	var out []PlagueRec
	for _, p := range w.State.Plagues {
		out = append(out, PlagueRec{
			Seed: w.seed(), Name: w.rd.Text("{plague:" + strconv.Itoa(p.ID) + "}"), Kind: p.Kind, Contagion: p.Contagion, Lethality: p.Lethality, Cause: p.Cause,
			Born: float64(p.Born-w.Dossier.Dawn) / 1e6, Peak: p.Peak, Caught: p.Caught, Worlds: p.Worlds, Peoples: p.Peoples,
			Cults: p.Cults, Cures: p.Cures, Refusals: p.Refusals, Woken: p.Woken, Wildfire: p.Wildfire, Extinct: p.Extinct,
			Made: p.Made, Maker: p.Maker, Rider: p.Rider, Poisoned: p.Poisonings,
		})
	}
	return out
}

// sickReport reads the plagues. The calibration targets are a third to a
// half of cradles meeting a plague in their first half million years,
// plague the second early killer after Overshoot, and one or two
// wildfires per age at 400 stars.
func sickReport(out io.Writer, plagues []PlagueRec, recs []Rec, seeds int) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Sickness")
	p("")
	p("A plague is an actor: born in a people at a rate its dirt raises and its medicine divides, carried on goods, occupation, landings and messages, fought each tick, paid for in worlds. Cradles are peoples born on their own world, not branches, cults or the made.")
	p("")
	byKind := map[string]int{}
	byCause := map[string]int{}
	var far []PlagueRec
	wild, extinct, worlds, peoples, cults, cures, refusals, woken, spread2 := 0, 0, 0, 0, 0, 0, 0, 0, 0
	var hosts []float64
	for _, pl := range plagues {
		byKind[pl.Kind]++
		byCause[pl.Cause]++
		hosts = append(hosts, float64(pl.Caught))
		if pl.Caught >= 2 {
			spread2++
		}
		if pl.Caught > 3 {
			far = append(far, pl)
		}
		if pl.Wildfire {
			wild++
		}
		if pl.Extinct {
			extinct++
		}
		worlds += pl.Worlds
		peoples += pl.Peoples
		cults += pl.Cults
		cures += pl.Cures
		refusals += pl.Refusals
		woken += pl.Woken
	}
	n := max(len(plagues), 1)
	fs := float64(max(seeds, 1))
	p("Plagues born %.1f per world: %s biological, %s memetic. By cause: %s.", float64(len(plagues))/fs, pct(byKind[plague.Biological.String()], n), pct(byKind[plague.Memetic.String()], n), causeLine(byCause, n))
	p("Hosts per plague: median %.0f, mean %.1f; %s reached a second host, %s went past three. Wildfires (ten hosts at once) %.1f per world. Extinct at the present %s.", median(hosts), mean(hosts), pct(spread2, n), pct(len(far), n), float64(wild)/fs, pct(extinct, n))
	if len(far) > 0 {
		var cs, ls []float64
		for _, pl := range far {
			cs = append(cs, pl.Contagion)
			ls = append(ls, pl.Lethality)
		}
		p("The ones that went past three hosts: contagion median %.2f, lethality median %.2f (all plagues are drawn uniform on both).", median(cs), median(ls))
	}
	p("Taken: %d worlds, %d peoples ended or brought low, %d cults formed; %d cures, %d peoples closed their ears or ports, %d reservoirs and walls woken.", worlds, peoples, cults, cures, refusals, woken)
	// cradles and the early killers
	cradles, met, metEarly := 0, 0, 0
	early := map[string]int{}
	earlyN := 0
	for _, r := range recs {
		if r.Made != "" || r.Origin != "" {
			continue
		}
		cradles++
		if r.Sick >= 0 {
			met++
			if r.Sick <= 0.5 {
				metEarly++
			}
		}
		if r.Standing || r.Lived > 0.5 {
			continue
		}
		earlyN++
		early[killer(r.Cause)]++
	}
	p("Cradles %d: %s met a plague at some point, %s in their first half million years.", cradles, pct(met, max(cradles, 1)), pct(metEarly, max(cradles, 1)))
	type kv struct {
		k string
		v int
	}
	var ks []kv
	for k, v := range early {
		ks = append(ks, kv{k, v})
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i].v > ks[j].v || (ks[i].v == ks[j].v && ks[i].k < ks[j].k) })
	var parts []string
	for i, x := range ks {
		if i >= 6 {
			break
		}
		parts = append(parts, fmt.Sprintf("%s %s", x.k, pct(x.v, max(earlyN, 1))))
	}
	p("Early killers (cradles ended within half a million years, %d): %s.", earlyN, strings.Join(parts, ", "))
	// the peoples' side
	sick, contained, cured, sickened := 0, 0, 0, 0
	for _, r := range recs {
		if r.Tally.Sickened > 0 {
			sick++
		}
		sickened += r.Tally.Sickened
		contained += r.Tally.Contained
		cured += r.Tally.Cured
	}
	p("Peoples that ever had one %s; infections per such people %.1f; ticks contained %d; peoples cured at least once %d.", pct(sick, max(len(recs), 1)), float64(sickened)/float64(max(sick, 1)), contained, cured)
	// parasites and the made
	born, woke, made, standing, ridden, risen := 0, 0, 0, 0, 0, 0
	var hostsPer []float64
	for _, r := range recs {
		if r.Sub != "parasite" {
			continue
		}
		switch {
		case strings.HasPrefix(r.Origin, "a rider born"):
			born++
		case r.Master:
			made++
		default:
			woke++
		}
		if r.Standing {
			standing++
		}
		ridden += r.Tally.Ridden
		risen += r.Tally.Risen
		hostsPer = append(hostsPer, float64(r.Tally.Ridden))
	}
	parasites := born + woke + made
	p("Parasites %.1f per world: %d born with their host, %d woke from a plague, %d made to think; %d standing at the present; peoples ridden %d (%.1f per parasite), risings %d.", float64(parasites)/fs, born, woke, made, standing, ridden, float64(ridden)/float64(max(parasites, 1)), risen)
	weapons, poisonings, attempts, detected, breakouts, leaks, makers, warring := 0, 0, 0, 0, 0, 0, 0, 0
	for _, pl := range plagues {
		if pl.Cause == "made" || pl.Cause == "breakout" {
			weapons++
			poisonings += pl.Poisoned
		}
	}
	for _, r := range recs {
		attempts += r.Tally.Attempts
		detected += r.Tally.Detected
		breakouts += r.Tally.Breakouts
		leaks += r.Tally.Leaks
		if r.Tally.Attempts > 0 {
			makers++
		}
		if r.Tally.Fought > 0 && r.Era >= 2 {
			warring++
		}
	}
	p("Engineered plagues %d (%.1f per world); %d peoples tried one on another, of the %d that fought a war at era 2 or above (%s); attempts %d, took %d, caught %d; got out at discovery %d, leaked while held %d.", weapons, float64(weapons)/fs, makers, warring, pct(makers, max(warring, 1)), attempts, poisonings, detected, breakouts-leaks, leaks)
}

// killer folds a cause of death into its family: the filter or the road
// that ended the people.
func killer(cause string) string {
	for _, k := range killers {
		if strings.Contains(cause, k.mark) {
			return k.name
		}
	}
	if cause == "" {
		return "none"
	}
	if i := strings.Index(cause, ","); i > 0 {
		cause = cause[:i]
	}
	if len(cause) > 40 {
		cause = cause[:40]
	}
	return cause
}

var killers = []struct{ mark, name string }{
	{"sickened and died", "plague"}, {"hollowed out by", "plague"}, {"listened to", "plague"}, {"nothing left to wear", "starved rider"},
	{"exhausted their world", "overshoot"},
	{"burned their world", "atomic"}, {"burned themselves out", "atomic"},
	{"collapsed under their own weight", "the weight of ages"}, {"faded away", "faded as a remnant"},
	{"fought over god", "the wars of faith"},
	{"consumed by their own machines", "thinking machines"},
	{"stopped dying", "the long silence"},
	{"broke their own star", "stellar engineering"},
	{"became strangers", "the distance"},
	{"went elsewhere", "transcendence"},
	{"went into the machines", "heaven"},
	{"lost their last world", "war"}, {"destroyed", "war"}, {"glassed", "war"},
}

func causeLine(m map[string]int, n int) string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	var parts []string
	for _, k := range ks {
		parts = append(parts, fmt.Sprintf("%s %s", k, pct(m[k], n)))
	}
	return strings.Join(parts, ", ")
}
