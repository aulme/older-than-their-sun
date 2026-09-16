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
	flag.Parse()

	cfg := history.DefaultConfig()
	cfg.Stars = *stars
	w := history.Generate(*seed, cfg)
	legends.Write(os.Stdout, w)
}
