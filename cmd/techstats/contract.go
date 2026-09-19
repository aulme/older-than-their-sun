package main

import (
	"fmt"
	"io"
	"sort"

	"worldgen/internal/history"
	"worldgen/internal/mind"
)

// Contracts: what was asked and paid, how offers fared, how they ended,
// tribute against worlds, and who lived by the sword.

// ContractRec is one contract.
type ContractRec struct {
	Seed      uint64
	Ask, Pay  string
	State     string
	By        string // "buyer" or "seller": who proposed it
	Years     float64
	Tribute   bool
	BoughtOff bool
	Stars     bool // both parties had reached the stars
}

// SellRec is one people's side of the trade, for the per-people lines.
type SellRec struct {
	Seed             uint64
	Stars, Sellsword bool
	Hired, Sold      int
	Broke, Tributes  int
	Taught           int
}

func flattenContracts(w *history.World) ([]ContractRec, []SellRec) {
	var out []ContractRec
	for _, k := range w.Contracts {
		b, s := w.Civs[k.Buyer], w.Civs[k.Seller]
		r := ContractRec{Seed: w.Seed, Ask: k.Ask.Kind.String(), Pay: k.Pay.Kind.String(), State: k.State.String(), By: "buyer",
			Tribute: k.Tribute, BoughtOff: k.BoughtOff, Stars: b.Starfaring > 0 && s.Starfaring > 0}
		if k.By == k.Seller {
			r.By = "seller"
		}
		if k.State == history.Running {
			r.Years = float64(w.Present - k.Formed)
		} else if k.Ended > 0 && k.Formed > 0 {
			r.Years = float64(k.Ended - k.Formed)
		}
		out = append(out, r)
	}
	var sells []SellRec
	for _, c := range w.Civs {
		sells = append(sells, SellRec{Seed: w.Seed, Stars: c.Starfaring > 0, Sellsword: c.Sellsword, Hired: c.Tally.Hired, Sold: c.Tally.Sold, Broke: c.Tally.Broke, Tributes: c.Tally.Tributes, Taught: len(c.Taught)})
	}
	return out, sells
}

// contractReport reads the contracts. The calibration targets are
// contracts per starfaring people a small number and not zero, bought-off
// guards rare, and tribute a share of the yields.
func contractReport(out io.Writer, sells []SellRec, ks []ContractRec, wars []WarRec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Contracts")
	p("")
	p("A timed payment for a discrete thing: a flow, a rarity, access to one, a node taught, a fleet on guard, a fleet against a world, a world handed over, peace, a sighting, a broker's word. An offer and its answer travel at the lag of light; the contract forms when the answer arrives; a term undelivered two ticks running breaks it, a betrayal on the world.")
	p("")
	stars, hired, sold, broke, swords, taught := 0, 0, 0, 0, 0, 0
	any := 0
	for _, s := range sells {
		if !s.Stars {
			continue
		}
		stars++
		hired += s.Hired
		sold += s.Sold
		broke += s.Broke
		taught += s.Taught
		if s.Sellsword {
			swords++
		}
		if s.Hired+s.Sold > 0 {
			any++
		}
	}
	if stars == 0 {
		p("No people reached the stars.")
		return
	}
	p("Per starfaring people: %.2f contracts formed as the buyer, %.2f as the seller, %.2f broken by it; %s had at least one; %s learned a node by being taught; %s earned the name sellswords.",
		float64(hired)/float64(stars), float64(sold)/float64(stars), float64(broke)/float64(stars), pct(any, stars), pct(countSells(sells, func(s SellRec) bool { return s.Stars && s.Taught > 0 }), stars), pct(swords, stars))
	p("")
	p("| Asked for | Offered | By the seller | Formed | Done | Broken | Lapsed | Bought off | Mean years run |")
	p("|---|---|---|---|---|---|---|---|---|")
	byAsk := map[string][]ContractRec{}
	for _, k := range ks {
		if k.Tribute {
			continue
		}
		byAsk[k.Ask] = append(byAsk[k.Ask], k)
	}
	kinds := []string{}
	for i := mind.TermFlow; i <= mind.TermBroker; i++ {
		if len(byAsk[i.String()]) > 0 {
			kinds = append(kinds, i.String())
		}
	}
	for _, kind := range kinds {
		rows := byAsk[kind]
		offered, bySeller, formed, done, broken, lapsed, bought := len(rows), 0, 0, 0, 0, 0, 0
		years := 0.0
		for _, k := range rows {
			if k.By == "seller" {
				bySeller++
			}
			switch k.State {
			case "running", "done", "broken", "lapsed":
				formed++
				years += k.Years
			}
			switch k.State {
			case "done":
				done++
			case "broken":
				broken++
			case "lapsed":
				lapsed++
			}
			if k.BoughtOff {
				bought++
			}
		}
		mean := "-"
		if formed > 0 {
			mean = fmt.Sprintf("%.0f", years/float64(formed))
		}
		p("| %s | %d | %d | %s | %d | %d | %d | %d | %s |", kind, offered, bySeller, pct(formed, offered), done, broken, lapsed, bought, mean)
	}
	p("")
	pays := map[string]int{}
	for _, k := range ks {
		if !k.Tribute {
			pays[k.Pay]++
		}
	}
	var pk []string
	for k := range pays {
		pk = append(pk, k)
	}
	sort.Strings(pk)
	line := "Paid with: "
	for i, k := range pk {
		if i > 0 {
			line += ", "
		}
		line += fmt.Sprintf("%s %d", k, pays[k])
	}
	p("%s.", line)
	p("")
	yields, tributes := 0, 0
	for _, wr := range wars {
		switch wr.Result {
		case "tribute":
			tributes++
			yields++
		case "capitulation", "vassal", "enslaved", "extinction":
			yields++
		}
	}
	p("Wars that ended in a yield: %d, of which tribute instead of worlds %s. Tributes paid per starfaring people %.2f.", yields, pct(tributes, yields), float64(countTributes(sells))/float64(stars))
}

func countSells(sells []SellRec, f func(SellRec) bool) int {
	n := 0
	for _, s := range sells {
		if f(s) {
			n++
		}
	}
	return n
}

func countTributes(sells []SellRec) int {
	n := 0
	for _, s := range sells {
		if s.Stars {
			n += s.Tributes
		}
	}
	return n
}
