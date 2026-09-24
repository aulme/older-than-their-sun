package history

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"worldgen/internal/mind"
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
// WAR_STARS (400), WAR_SEEDS (20) and WAR_FROM (1) set the batch;
// WAR_TUNE=Group.Field=v,... sets the mind's tuning for the batch;
// WAR_WATCH=dir runs the watchers (watch.go) on every seed and writes
// what they catch to dir/seed-N.txt, with a line per catch in the log.
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
	var tunes []string
	if s := os.Getenv("WAR_TUNE"); s != "" {
		tunes = strings.Split(s, ",")
	}
	tuning, err := mind.Configure("", tunes)
	if err != nil {
		t.Fatal(err)
	}
	watchDir := os.Getenv("WAR_WATCH")
	if watchDir != "" {
		if err := os.MkdirAll(watchDir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	fired := make([][]string, seeds)
	watches := make([]*Watch, seeds)
	out := make([]*warshape.Shape, seeds)
	var wg sync.WaitGroup
	for i := range seeds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := DefaultConfig()
			cfg.Stars = stars
			cfg.Tuning = tuning
			if watchDir != "" {
				f, err := os.Create(filepath.Join(watchDir, sprintf("seed-%d.txt", from+i)))
				if err != nil {
					t.Error(err)
					return
				}
				defer f.Close()
				wt := DefaultWatch(f)
				wt.Attach(&cfg)
				defer func() { fired[i], watches[i] = wt.Fired, wt }()
			}
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
	for _, f := range fired {
		for _, l := range f {
			t.Log("watch: " + l)
		}
	}
	if watchDir != "" {
		t.Log(StareReport(watches, 20))
	}
}
