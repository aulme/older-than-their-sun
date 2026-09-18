package main

import (
	"fmt"
	"io"
	"math"

	"worldgen/internal/history"
)

// BattleRec is one battle at a world, for the report.
type BattleRec struct {
	Seed     uint64
	Year     int
	Star     int
	Attacker int
	Ships    int     // the attacker's
	Held     int     // ships and guns in the sky
	Gap      float64 // the attacker's levels less the defender's
	Won      bool
	Outcome  string
	Fewer    bool // the side with fewer ships won
}

func flattenBattles(w *history.World) []BattleRec {
	var out []BattleRec
	for _, b := range w.Battles {
		fewer := false
		if b.Ships != b.Held {
			fewer = b.Won == (b.Ships < b.Held)
		}
		out = append(out, BattleRec{Seed: w.Seed, Year: int(b.Year), Star: b.Star, Attacker: b.Attacker, Ships: b.Ships, Held: b.Held, Gap: b.Gap, Won: b.Won, Outcome: b.Outcome, Fewer: fewer})
	}
	return out
}

// battlesReport reads the fighting: battles by outcome, how often the
// side with fewer ships won by the gap in levels, garrison moves and
// musters, and worlds struck with an empty sky. The calibration targets
// are the smaller fleet winning often at two levels' gap and rarely at
// none, and sieges of several ticks appearing.
func battlesReport(out io.Writer, recs []Rec, battles []BattleRec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Battles")
	p("")
	p("A battle is a campaign fleet at a world against its sky: guns, guard and relief, each at its owner's quality; one a tick at a world. The loser of the roll withdraws if it lost ships; a world with a gun standing is never taken; a world with nothing in its sky is taken without a battle.")
	p("")
	by := map[string]int{}
	for _, b := range battles {
		by[b.Outcome]++
	}
	p("%d battles: %d taken with nothing in the sky, %d taken after a fight, %d where the ships withdrew and the guns stood, %d where a gun stood and the fleets stayed behind it, %d held by the defender, %d where the attacker was broken.",
		len(battles), by["empty"], by["taken"], by["withdrew"], by["guns"], by["held"], by["broken"])
	p("")
	p("| Gap in levels (attacker less defender) | Battles fought | Won by the side with fewer ships |")
	p("|---|---|---|")
	buckets := []struct {
		name   string
		lo, hi float64
	}{
		{"within half a level", 0, 0.5}, {"half to two levels", 0.5, 2}, {"two to four", 2, 4}, {"four or more", 4, 1e9},
	}
	for _, bk := range buckets {
		n, fewer := 0, 0
		for _, b := range battles {
			if b.Outcome == "empty" || b.Ships == b.Held {
				continue
			}
			if g := math.Abs(b.Gap); g < bk.lo || g >= bk.hi {
				continue
			}
			n++
			if b.Fewer {
				fewer++
			}
		}
		p("| %s | %d | %s |", bk.name, n, pct(fewer, n))
	}
	p("")
	// a siege is the same attacker at the same world in battle after battle
	// within a few ticks of the last
	type at struct {
		seed           uint64
		star, attacker int
	}
	last := map[at]int{}
	run := map[at]int{}
	sieges, long := 0, 0
	for _, b := range battles {
		k := at{b.Seed, b.Star, b.Attacker}
		if y, ok := last[k]; ok && b.Year-y <= 3000 {
			run[k]++
			if run[k] == 1 {
				sieges++
			}
			if run[k] == 3 {
				long++
			}
		} else {
			run[k] = 0
		}
		last[k] = b.Year
	}
	p("Sieges (a second battle at a world by the same fleet within three thousand years) %d, of four battles or more %d.", sieges, long)
	p("")
	garrisons, musters, empty, fought, won := 0, 0, 0, 0, 0
	for _, r := range recs {
		garrisons += r.Tally.Garrisons
		musters += r.Tally.Musters
		empty += r.Tally.EmptySky
		fought += r.Tally.Battles
		won += r.Tally.Won
	}
	p("Garrison moves %d, musters %d; worlds struck with an empty sky %d of %d battles; the attacker won the roll in %s.", garrisons, musters, empty, fought, pct(won, fought))
}
