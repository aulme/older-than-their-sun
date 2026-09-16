// Command worldgen generates a galactic history and prints its legends.
package main

import (
	"flag"
	"os"

	"worldgen/internal/history"
	"worldgen/internal/legends"
)

func main() {
	seed := flag.Uint64("seed", 1, "world seed")
	stars := flag.Int("stars", 400, "number of stars")
	full := flag.Bool("full", false, "print known tech per civilisation")
	debug := flag.Bool("debug", false, "log the state of the galaxy every million years")
	stats := flag.Bool("stats", false, "print one line of numbers instead of the legends")
	flag.Parse()

	cfg := history.DefaultConfig()
	cfg.Stars = *stars
	cfg.Debug = *debug
	w := history.Generate(*seed, cfg)
	if *stats {
		legends.Stats(os.Stdout, w)
		return
	}
	legends.Write(os.Stdout, w, *full)
}
