// Command worldgen generates a galactic history and prints its legends.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"worldgen/internal/galaxy"
	"worldgen/internal/history"
	"worldgen/internal/legends"
	"worldgen/internal/mind"
)

func main() {
	seed := flag.Uint64("seed", 1, "world seed")
	stars := flag.Int("stars", 400, "number of stars")
	full := flag.Bool("full", false, "print known tech per civilisation")
	debug := flag.Bool("debug", false, "log the state of the galaxy every million years")
	ai := flag.Bool("ai", false, "log the reason behind every decision a people makes")
	tuning := flag.String("tuning", "", "a JSON file of mind.Tuning; fields left out keep their defaults")
	var tunes []string
	flag.Func("tune", "one override of the mind's tuning, Group.Field=value; may repeat", func(s string) error { tunes = append(tunes, s); return nil })
	phases := flag.Bool("phases", false, "log each phase's time every million years")
	stats := flag.Bool("stats", false, "print one line of numbers instead of the legends")
	cpuprofile := flag.String("cpuprofile", "", "write a CPU profile of the run to this file")
	at := flag.String("at", "sol", "where in the galaxy: a named place, a feature such as \"Cygnus X-1\", or x,y,z in kpc (see -map)")
	mapOnly := flag.Bool("map", false, "print a chart of the galaxy, the laws from centre to rim, and the named places, then exit")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	rg, err := galaxy.RegionByName(*at)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *mapOnly {
		legends.Map(os.Stdout, rg)
		return
	}
	cfg := history.DefaultConfig()
	cfg.Region = *at
	cfg.Stars = *stars
	cfg.Debug = *debug
	cfg.TraceAI = *ai
	cfg.Profile = *phases
	if cfg.Tuning, err = mind.Configure(*tuning, tunes); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	w := history.Generate(*seed, cfg)
	if *stats {
		legends.Stats(os.Stdout, w)
		return
	}
	legends.Write(os.Stdout, w, *full)
}
