package history

import (
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"worldgen/internal/warshape"
)

// TestWarShape is the measurement behind the war gate (specs/plan.md
// steps 10 and 11; specs/proposals/war.md): over a batch of whole ages it
// reads the record, and only the record, for the seven patterns of war —
// short and long wars, old enemies, world wars, border disputes, cold
// wars, proxy wars and conquest waves — with the vassal's lot and the
// health of it beside them, per seed and over the batch.
//
//	WAR=1 go test ./internal/history -run TestWarShape -v -timeout 6h
//
// WAR_STARS (400), WAR_SEEDS (20) and WAR_FROM (1) set the batch.
func TestWarShape(t *testing.T) {
	if os.Getenv("WAR") == "" {
		t.Skip("a measurement, not a gate: set WAR=1 to run it")
	}
	stars, seeds, from := 400, 20, 1
	if s := os.Getenv("WAR_STARS"); s != "" {
		stars, _ = strconv.Atoi(s)
	}
	if s := os.Getenv("WAR_SEEDS"); s != "" {
		seeds, _ = strconv.Atoi(s)
	}
	if s := os.Getenv("WAR_FROM"); s != "" {
		from, _ = strconv.Atoi(s)
	}
	out := make([]*warshape.Shape, seeds)
	var wg sync.WaitGroup
	for i := range seeds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := DefaultConfig()
			cfg.Stars = stars
			start := time.Now()
			w := Generate(uint64(from+i), cfg)
			took := time.Since(start)
			out[i] = warshape.Read(w.Export(cfg.Region))
			out[i].Took = took
		}()
	}
	wg.Wait()
	for _, l := range warshape.Report(out) {
		t.Log(l)
	}
}
