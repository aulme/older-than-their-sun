// Command worldgen generates a galactic history and writes it as a run
// directory: the dossier, the state, the chronicle and the tellings,
// with the lookups, the codex and the format contract copied in. The
// legends, the readable view of a run, are rendered from the directory
// and nothing else.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"

	"worldgen/internal/galaxy"
	"worldgen/internal/history"
	"worldgen/internal/legends"
	"worldgen/internal/mind"
	"worldgen/internal/record"
	"worldgen/internal/writer"
)

func main() {
	seed := flag.Uint64("seed", 1, "world seed")
	stars := flag.Int("stars", 400, "number of stars")
	out := flag.String("out", "", "the run directory to write; the default is out/<seed>")
	until := flag.Int64("until", 0, "run the age to this year (years since the dawn) and write the directory as of then")
	show := flag.Bool("legends", false, "print the legends of the run after writing it")
	read := flag.String("read", "", "print the legends of a run directory written earlier, and generate nothing")
	full := flag.Bool("full", false, "with -legends or -read: print known tech per civilisation and the tellings of the dead")
	debug := flag.Bool("debug", false, "log the state of the galaxy every million years")
	ai := flag.Bool("ai", false, "log the reason behind every decision a people makes")
	tuning := flag.String("tuning", "", "a JSON file of mind.Tuning; fields left out keep their defaults")
	var tunes []string
	flag.Func("tune", "one override of the mind's tuning, Group.Field=value; may repeat", func(s string) error { tunes = append(tunes, s); return nil })
	phases := flag.Bool("phases", false, "log each phase's time every million years")
	stats := flag.Bool("stats", false, "print one line of numbers instead of the legends, and write nothing")
	cpuprofile := flag.String("cpuprofile", "", "write a CPU profile of the run to this file")
	wearEvery := flag.Int("wear", 4, "how many ticks apart a people's telling is put through the wearing; the rate is compounded over the gap, so a tale wears as often (see specs/plan.md step 6)")
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

	if *read != "" {
		r, err := record.Load(*read)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *stats {
			legends.Stats(os.Stdout, r)
			return
		}
		legends.Write(os.Stdout, r, *full)
		return
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
	cfg.WearEvery = *wearEvery
	cfg.Until = history.Year(*until)
	if cfg.Tuning, err = mind.Configure(*tuning, tunes); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	w := history.Generate(*seed, cfg)
	r := writer.Run(w, *at)
	if *stats {
		legends.Stats(os.Stdout, r)
		return
	}
	dir := *out
	if dir == "" {
		dir = filepath.Join("out", fmt.Sprint(*seed))
	}
	if err := record.Write(dir, r); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *show {
		// the view reads the directory it just wrote, and nothing else
		r, err := record.Load(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		legends.Write(os.Stdout, r, *full)
		return
	}
	fmt.Fprintf(os.Stderr, "wrote %s: %d events, %d tales, %d names\n", dir, len(r.Chronicle), len(r.Tellings), len(r.State.Names))
}
