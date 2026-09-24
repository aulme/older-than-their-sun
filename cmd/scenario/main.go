// Command scenario runs a hand-built world from a spec and tells it: a
// few named peoples, their councils' reasons, their fleets, battles,
// terms and will, tick by tick. See internal/history/scenario.go for the
// spec, and internal/history/testdata/scenarios for the ones the tests run.
//
//	go run ./cmd/scenario internal/history/testdata/scenarios/strike-back.json
package main

import (
	"flag"
	"fmt"
	"os"

	"worldgen/internal/history"
)

func main() {
	ticks := flag.Int("ticks", 0, "ticks to run; the spec's own if 0")
	seed := flag.Uint64("seed", 0, "the seed; the spec's own if 0")
	all := flag.Bool("all", false, "tell every reason and debug line, not only the war's and the council's")
	var tunes []string
	flag.Func("tune", "one override of the mind's tuning, Group.Field=value; may repeat", func(s string) error { tunes = append(tunes, s); return nil })
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: scenario [flags] spec.json")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	s, err := history.LoadScenario(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *seed != 0 {
		s.Seed = *seed
	}
	s.Tune = append(s.Tune, tunes...)
	r, err := s.Build(os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	r.All = *all
	n := s.Ticks
	if *ticks > 0 {
		n = *ticks
	}
	r.Ticks(n)
}
