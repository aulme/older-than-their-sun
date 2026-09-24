package history

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"worldgen/internal/mind"
	"worldgen/internal/record"
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
// WAR_STARS (400), WAR_SEEDS (20) and WAR_FROM (1) set the batch, or
// WAR_LIST=1,5,9 names its seeds;
// WAR_TUNE=Group.Field=v,... sets the mind's tuning for the batch;
// WAR_OUT=dir writes each seed's record to dir/seed-N, for cmd/warshape
// to read again without running it; WAR_SAVE=file keeps the batch's gate and WAR_BASE=file prints an
// earlier one's beside it; WAR_WATCH=dir runs the watchers (watch.go) on every seed and writes
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
	list := make([]uint64, 0, seeds)
	if s := os.Getenv("WAR_LIST"); s != "" {
		for _, x := range strings.Split(s, ",") {
			n, err := strconv.Atoi(x)
			if err != nil {
				t.Fatal(err)
			}
			list = append(list, uint64(n))
		}
		seeds = len(list)
	} else {
		for i := range seeds {
			list = append(list, uint64(from+i))
		}
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
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("seed %d panicked: %v\n%s", list[i], p, debug.Stack()) // the rest of the batch is still read
					out[i] = nil
				}
			}()
			cfg := DefaultConfig()
			cfg.Stars = stars
			cfg.Tuning = tuning
			if watchDir != "" {
				f, err := os.Create(filepath.Join(watchDir, sprintf("seed-%d.txt", list[i])))
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
			w := Generate(list[i], cfg)
			took := time.Since(start)
			r := w.Export(cfg.Region)
			if dir := os.Getenv("WAR_OUT"); dir != "" {
				if err := record.Write(filepath.Join(dir, sprintf("seed-%d", list[i])), r); err != nil {
					t.Error(err)
				}
			}
			out[i] = warshape.Read(r)
			out[i].Took = took
		}()
	}
	wg.Wait()
	read := out[:0:0]
	for _, x := range out {
		if x != nil {
			read = append(read, x)
		}
	}
	out = read
	for _, l := range warshape.Report(out) {
		t.Log(l)
	}
	gate := warshape.Gate(out)
	var base []warshape.Row
	if p := os.Getenv("WAR_BASE"); p != "" {
		if base, err = warshape.LoadGate(p); err != nil {
			t.Error(err)
		}
	}
	for _, l := range warshape.GateLines(gate, base) {
		t.Log(l)
	}
	if p := os.Getenv("WAR_SAVE"); p != "" {
		if err := warshape.SaveGate(p, gate); err != nil {
			t.Error(err)
		}
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
