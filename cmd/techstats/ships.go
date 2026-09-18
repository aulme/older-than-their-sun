package main

import (
	"fmt"
	"io"

	"worldgen/internal/tech"
)

// shipsReport reads the fleets: how many ships peoples held at their
// height, what they built, lost and let rot, what the building cost in
// ship-kyr of flow, and how much of their starfaring lives they spent at
// their want and with ships laid up. The calibration targets are a
// handful to a few dozen ships at a people's zenith, and laying up as a
// minority of the youth that grows in the waning.
func shipsReport(out io.Writer, recs []Rec) {
	p := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	p("")
	p("## Ships")
	p("")
	p("A ship costs four thousand years of its keep to build at a dock; a ship in being reserves its keep each tick; a fleet the flow cannot keep is laid up and rots at a tenth per thousand years. The want is what the policies ask for: the campaign the council could not man, the explorers, and a floor by fear.")
	p("")
	p("| Deepest era | Peoples starfaring | Ships at the height: median, quartiles, most | Built | Lost in battle | Rotted | Ship-kyr building per ship built | Ticks at the want | Ticks with ships laid up |")
	p("|---|---|---|---|---|---|---|---|---|")
	for e := 2; e <= 4; e++ {
		var peak []float64
		built, lost, rotted, star, atWant, laid := 0, 0, 0, 0, 0, 0
		building := 0.0
		for _, r := range recs {
			if r.Era != e || r.Tally.StarTicks == 0 {
				continue
			}
			peak = append(peak, float64(r.Ships))
			built += r.Tally.Built
			lost += r.Tally.ShipsLost
			rotted += r.Tally.Rotted
			building += r.Tally.Building
			star += r.Tally.StarTicks
			atWant += r.Tally.AtWant
			laid += r.Tally.LaidTick
		}
		if len(peak) == 0 {
			continue
		}
		per := 0.0
		if built > 0 {
			per = building / float64(built)
		}
		p("| %d %s | %d | %.0f, %.0f to %.0f, %.0f | %d | %d | %d | %.1f | %s | %s |", e, tech.EraNames[e], len(peak),
			median(peak), quantile(peak, 0.25), quantile(peak, 0.75), quantile(peak, 1), built, lost, rotted, per, pct(atWant, star), pct(laid, star))
	}
	p("")
	standing, withShips, laidUp := 0, 0, 0
	for _, r := range recs {
		if !r.Standing || r.Tally.StarTicks == 0 {
			continue
		}
		standing++
		if r.ShipsEnd > 0 {
			withShips++
		}
		if r.LaidUp > 0 {
			laidUp++
		}
	}
	p("Of %d starfaring peoples standing at the present, %d hold a ship (%s) and %d have ships laid up (%s).", standing, withShips, pct(withShips, standing), laidUp, pct(laidUp, standing))
}
